package com.bigbank.ledger

import com.bigbank.ledger.api.LedgerTransferRequest
import com.bigbank.ledger.config.CorrelationIdFilter
import com.bigbank.ledger.domain.Account
import com.bigbank.ledger.domain.JournalEntry
import com.bigbank.ledger.domain.LedgerTransaction
import com.bigbank.ledger.event.TransactionCompletedEvent
import com.bigbank.ledger.event.TransactionCompletedSideEffect
import com.bigbank.ledger.repository.AccountRepository
import com.bigbank.ledger.repository.JournalEntryRepository
import com.bigbank.ledger.repository.LedgerTransactionRepository
import com.bigbank.ledger.service.ApiException
import com.bigbank.ledger.service.LedgerCommandService
import com.bigbank.ledger.service.LedgerQueryService
import com.bigbank.ledger.service.LedgerTransferPersistenceHook
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.TestConfiguration
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Import
import org.springframework.context.annotation.Primary
import org.springframework.dao.DataIntegrityViolationException
import org.springframework.jdbc.core.JdbcTemplate
import org.springframework.test.context.ActiveProfiles
import org.springframework.test.context.event.ApplicationEvents
import org.springframework.test.context.event.RecordApplicationEvents
import java.math.BigDecimal
import java.sql.Timestamp
import java.time.Instant
import java.util.UUID
import java.util.concurrent.Callable
import java.util.concurrent.CopyOnWriteArrayList
import java.util.concurrent.CountDownLatch
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

@SpringBootTest
@ActiveProfiles("test")
@RecordApplicationEvents
@Import(LedgerApplicationTests.TestConfig::class)
class LedgerApplicationTests(
    @Autowired private val healthController: HealthController,
    @Autowired private val ledgerCommandService: LedgerCommandService,
    @Autowired private val ledgerQueryService: LedgerQueryService,
    @Autowired private val accountRepository: AccountRepository,
    @Autowired private val journalEntryRepository: JournalEntryRepository,
    @Autowired private val ledgerTransactionRepository: LedgerTransactionRepository,
    @Autowired private val jdbcTemplate: JdbcTemplate,
    @Autowired private val recordingSideEffect: RecordingTransactionCompletedSideEffect,
    @Autowired private val controllableHook: ControllableLedgerTransferPersistenceHook,
) {

    @BeforeEach
    fun setUp() {
        journalEntryRepository.deleteAll()
        ledgerTransactionRepository.deleteAll()
        accountRepository.deleteAll()

        accountRepository.saveAll(
            listOf(
                Account(accountNumber = "ACC-100", ownerName = "Dina"),
                Account(accountNumber = "ACC-200", ownerName = "Eka"),
            ),
        )

        recordingSideEffect.reset()
        controllableHook.reset()
    }

    @AfterEach
    fun tearDown() {
        recordingSideEffect.reset()
        controllableHook.reset()
    }

    @Test
    fun `health endpoint is up`() {
        val response = healthController.health(
            org.springframework.mock.web.MockHttpServletRequest().apply {
                setAttribute(CorrelationIdFilter.CORRELATION_ID_ATTRIBUTE, "corr-health")
            },
        )

        assertEquals("success", response.status)
        assertEquals("corr-health", response.correlationId)
        assertEquals("UP", response.data?.get("status"))
        assertEquals("ledger", response.data?.get("service"))
    }

    @Test
    fun `valid transfer creates transaction journals and emits event`(applicationEvents: ApplicationEvents) {
        val reference = UUID.randomUUID().toString()

        val response = ledgerCommandService.transfer(
            LedgerTransferRequest(
                reference = reference,
                fromAccount = "ACC-100",
                toAccount = "ACC-200",
                amount = BigDecimal("1500.00"),
            ),
            correlationId = "corr-valid-transfer",
        )

        val transactions = ledgerTransactionRepository.findAll()
        val journalEntries = journalEntryRepository.findAll()

        assertEquals(reference, response.reference)
        assertFalse(response.duplicate)
        assertEquals("COMPLETED", response.status)
        assertEquals(1, transactions.size)
        assertEquals(2, journalEntries.size)
        assertEquals(0, journalEntries[0].amount.compareTo(journalEntries[1].amount))
        assertTrue(recordingSideEffect.await())
        assertEquals(1, applicationEvents.stream(TransactionCompletedEvent::class.java).count())
        assertEquals(reference, recordingSideEffect.capturedEvents().single().reference)
    }

    @Test
    fun `duplicate transfer returns existing transaction safely`(applicationEvents: ApplicationEvents) {
        val reference = UUID.randomUUID().toString()
        val request = LedgerTransferRequest(
            reference = reference,
            fromAccount = "ACC-100",
            toAccount = "ACC-200",
            amount = BigDecimal("700.00"),
        )

        ledgerCommandService.transfer(request, "corr-duplicate-1")
        val duplicate = ledgerCommandService.transfer(request, "corr-duplicate-2")

        assertTrue(duplicate.duplicate)
        assertEquals(1, ledgerTransactionRepository.count())
        assertEquals(2, journalEntryRepository.count())
        assertEquals(1, applicationEvents.stream(TransactionCompletedEvent::class.java).count())
    }

    @Test
    fun `invalid amount is rejected`() {
        val exception = assertThrows<ApiException> {
            ledgerCommandService.transfer(
                LedgerTransferRequest(
                    reference = "ref-invalid-amount",
                    fromAccount = "ACC-100",
                    toAccount = "ACC-200",
                    amount = BigDecimal.ZERO,
                ),
                "corr-invalid-amount",
            )
        }

        assertEquals("INVALID_AMOUNT", exception.code)
    }

    @Test
    fun `transactions endpoint returns persisted transfers with limit`() {
        repeat(3) { index ->
            ledgerCommandService.transfer(
                LedgerTransferRequest(
                    reference = "ref-list-$index",
                    fromAccount = "ACC-100",
                    toAccount = "ACC-200",
                    amount = BigDecimal("250.00"),
                ),
                "corr-list-transactions-$index",
            )
        }

        val transactions = ledgerQueryService.listTransactions(limit = 1)

        assertEquals(1, transactions.size)
        assertEquals("ref-list-2", transactions[0].reference)
    }

    @Test
    fun `concurrent duplicate reference remains idempotent`() {
        val request = LedgerTransferRequest(
            reference = "ref-concurrent-duplicate",
            fromAccount = "ACC-100",
            toAccount = "ACC-200",
            amount = BigDecimal("88.00"),
        )
        val executor = Executors.newFixedThreadPool(6)
        val startGate = CountDownLatch(1)

        try {
            val futures = (1..6).map { index ->
                executor.submit(Callable {
                    startGate.await(5, TimeUnit.SECONDS)
                    ledgerCommandService.transfer(request, "corr-concurrent-duplicate-$index")
                })
            }

            startGate.countDown()
            val responses = futures.map { it.get(10, TimeUnit.SECONDS) }

            assertEquals(1, responses.count { !it.duplicate })
            assertEquals(5, responses.count { it.duplicate })
            assertEquals(1, ledgerTransactionRepository.count())
            assertEquals(2, journalEntryRepository.count())
            assertTrue(recordingSideEffect.await())
            assertEquals(1, recordingSideEffect.capturedEvents().size)
        } finally {
            executor.shutdownNow()
        }
    }

    @Test
    fun `concurrent distinct transfers stay balanced`() {
        val executor = Executors.newFixedThreadPool(5)
        val startGate = CountDownLatch(1)

        try {
            val futures = (1..5).map { index ->
                executor.submit(Callable {
                    startGate.await(5, TimeUnit.SECONDS)
                    ledgerCommandService.transfer(
                        LedgerTransferRequest(
                            reference = "ref-concurrent-$index",
                            fromAccount = "ACC-100",
                            toAccount = "ACC-200",
                            amount = BigDecimal("42.00"),
                        ),
                        "corr-concurrent-$index",
                    )
                })
            }

            recordingSideEffect.reset(expectedEvents = 5)
            startGate.countDown()
            futures.forEach { it.get(10, TimeUnit.SECONDS) }

            val journalEntries = journalEntryRepository.findAll()
            val debitTotal = journalEntries
                .filter { it.entryType.name == "DEBIT" }
                .fold(BigDecimal.ZERO) { total, entry -> total + entry.amount }
            val creditTotal = journalEntries
                .filter { it.entryType.name == "CREDIT" }
                .fold(BigDecimal.ZERO) { total, entry -> total + entry.amount }

            assertEquals(5, ledgerTransactionRepository.count())
            assertEquals(10, journalEntryRepository.count())
            assertEquals(0, debitTotal.compareTo(creditTotal))
            assertTrue(recordingSideEffect.await())
        } finally {
            executor.shutdownNow()
        }
    }

    @Test
    fun `forced failure rolls back partial state`() {
        controllableHook.failNextPersist()

        assertThrows<IllegalStateException> {
            ledgerCommandService.transfer(
                LedgerTransferRequest(
                    reference = "ref-rollback",
                    fromAccount = "ACC-100",
                    toAccount = "ACC-200",
                    amount = BigDecimal("55.00"),
                ),
                "corr-rollback",
            )
        }

        assertEquals(0, ledgerTransactionRepository.count())
        assertEquals(0, journalEntryRepository.count())
    }

    @Test
    fun `database unique constraint blocks duplicate reference`() {
        val fromAccount = accountRepository.findByAccountNumber("ACC-100")!!
        val toAccount = accountRepository.findByAccountNumber("ACC-200")!!

        ledgerTransactionRepository.saveAndFlush(
            LedgerTransaction(
                reference = "ref-db-unique",
                fromAccount = fromAccount,
                toAccount = toAccount,
                amount = BigDecimal("10.00"),
                correlationId = "corr-db-1",
            ),
        )

        assertThrows<DataIntegrityViolationException> {
            ledgerTransactionRepository.saveAndFlush(
                LedgerTransaction(
                    reference = "ref-db-unique",
                    fromAccount = fromAccount,
                    toAccount = toAccount,
                    amount = BigDecimal("11.00"),
                    correlationId = "corr-db-2",
                ),
            )
        }

        assertEquals(1, ledgerTransactionRepository.count())
    }

    @Test
    fun `database foreign keys block invalid journal entry`() {
        assertThrows<DataIntegrityViolationException> {
            jdbcTemplate.update(
                """
                insert into journal_entries (id, amount, created_at, entry_type, account_id, transaction_id)
                values (?, ?, ?, ?, ?, ?)
                """.trimIndent(),
                UUID.randomUUID(),
                BigDecimal("9.99"),
                Timestamp.from(Instant.now()),
                "DEBIT",
                UUID.randomUUID(),
                UUID.randomUUID(),
            )
        }
    }

    @TestConfiguration
    class TestConfig {
        @Bean
        @Primary
        fun recordingTransactionCompletedSideEffect() = RecordingTransactionCompletedSideEffect()

        @Bean
        @Primary
        fun controllableLedgerTransferPersistenceHook() = ControllableLedgerTransferPersistenceHook()
    }
}

class RecordingTransactionCompletedSideEffect : TransactionCompletedSideEffect {
    private val events = CopyOnWriteArrayList<TransactionCompletedEvent>()
    @Volatile
    private var latch = CountDownLatch(1)

    override fun handle(event: TransactionCompletedEvent) {
        events.add(event)
        latch.countDown()
    }

    fun reset(expectedEvents: Int = 1) {
        events.clear()
        latch = CountDownLatch(expectedEvents)
    }

    fun await(timeoutSeconds: Long = 5): Boolean = latch.await(timeoutSeconds, TimeUnit.SECONDS)

    fun capturedEvents(): List<TransactionCompletedEvent> = events.toList()
}

class ControllableLedgerTransferPersistenceHook : LedgerTransferPersistenceHook {
    private val failNext = AtomicBoolean(false)

    override fun beforeJournalPersist(transaction: LedgerTransaction, entries: List<JournalEntry>) {
        if (failNext.compareAndSet(true, false)) {
            throw IllegalStateException("forced journal persistence failure")
        }
    }

    fun failNextPersist() {
        failNext.set(true)
    }

    fun reset() {
        failNext.set(false)
    }
}
