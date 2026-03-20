package com.bigbank.ledger

import com.bigbank.ledger.api.LedgerTransferRequest
import com.bigbank.ledger.domain.Account
import com.bigbank.ledger.repository.AccountRepository
import com.bigbank.ledger.repository.JournalEntryRepository
import com.bigbank.ledger.repository.LedgerTransactionRepository
import com.bigbank.ledger.service.ApiException
import com.bigbank.ledger.service.LedgerCommandService
import com.bigbank.ledger.service.LedgerQueryService
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.test.context.ActiveProfiles
import java.math.BigDecimal
import java.util.UUID

@SpringBootTest
@ActiveProfiles("test")
class LedgerApplicationTests(
    @Autowired private val healthController: HealthController,
    @Autowired private val ledgerCommandService: LedgerCommandService,
    @Autowired private val ledgerQueryService: LedgerQueryService,
    @Autowired private val accountRepository: AccountRepository,
    @Autowired private val journalEntryRepository: JournalEntryRepository,
    @Autowired private val ledgerTransactionRepository: LedgerTransactionRepository,
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
    }

    @Test
    fun `health endpoint is up`() {
        val response = healthController.health()

        kotlin.test.assertEquals("UP", response["status"])
        kotlin.test.assertEquals("ledger", response["service"])
    }

    @Test
    fun `valid transfer creates transaction and journal entries`() {
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

        kotlin.test.assertEquals(reference, response.reference)
        kotlin.test.assertEquals(false, response.duplicate)
        kotlin.test.assertEquals("COMPLETED", response.status)
        kotlin.test.assertEquals(1, transactions.size)
        kotlin.test.assertEquals(2, journalEntries.size)
        kotlin.test.assertEquals(
            0,
            journalEntries[0].amount.compareTo(journalEntries[1].amount),
        )
    }

    @Test
    fun `duplicate transfer returns existing transaction safely`() {
        val reference = UUID.randomUUID().toString()
        val request = LedgerTransferRequest(
            reference = reference,
            fromAccount = "ACC-100",
            toAccount = "ACC-200",
            amount = BigDecimal("700.00"),
        )

        ledgerCommandService.transfer(request, "corr-duplicate-1")
        val duplicate = ledgerCommandService.transfer(request, "corr-duplicate-2")

        kotlin.test.assertEquals(true, duplicate.duplicate)
        kotlin.test.assertEquals(1, ledgerTransactionRepository.count())
        kotlin.test.assertEquals(2, journalEntryRepository.count())
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

        kotlin.test.assertEquals("INVALID_AMOUNT", exception.code)
    }

    @Test
    fun `transactions endpoint returns persisted transfers`() {
        val reference = UUID.randomUUID().toString()

        ledgerCommandService.transfer(
            LedgerTransferRequest(
                reference = reference,
                fromAccount = "ACC-100",
                toAccount = "ACC-200",
                amount = BigDecimal("250.00"),
            ),
            "corr-list-transactions",
        )

        val transactions = ledgerQueryService.listTransactions()

        kotlin.test.assertEquals(1, transactions.size)
        kotlin.test.assertEquals(reference, transactions[0].reference)
    }
}
