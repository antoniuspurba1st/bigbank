package com.bigbank.ledger.service

import com.bigbank.ledger.api.LedgerTransferResponse
import com.bigbank.ledger.domain.EntryType
import com.bigbank.ledger.domain.JournalEntry
import com.bigbank.ledger.domain.LedgerTransaction
import com.bigbank.ledger.event.TransactionCompletedEvent
import com.bigbank.ledger.repository.AccountRepository
import com.bigbank.ledger.repository.JournalEntryRepository
import com.bigbank.ledger.repository.LedgerTransactionRepository
import org.slf4j.LoggerFactory
import org.springframework.context.ApplicationEventPublisher
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Component
import org.springframework.transaction.annotation.Propagation
import org.springframework.transaction.annotation.Transactional
import java.math.BigDecimal

@Component
class LedgerPostingExecutor(
    private val accountRepository: AccountRepository,
    private val ledgerTransactionRepository: LedgerTransactionRepository,
    private val journalEntryRepository: JournalEntryRepository,
    private val eventPublisher: ApplicationEventPublisher,
    private val ledgerTransferPersistenceHook: LedgerTransferPersistenceHook,
) {

    private val logger = LoggerFactory.getLogger(javaClass)

    @Transactional(propagation = Propagation.REQUIRES_NEW)
    fun postTransfer(command: LedgerTransferCommand, correlationId: String): LedgerTransferResponse {
        val fromAccount = accountRepository.findByAccountNumber(command.fromAccount)
            ?: throw ApiException(HttpStatus.NOT_FOUND, "ACCOUNT_NOT_FOUND", "Source account does not exist")
        val toAccount = accountRepository.findByAccountNumber(command.toAccount)
            ?: throw ApiException(HttpStatus.NOT_FOUND, "ACCOUNT_NOT_FOUND", "Destination account does not exist")

        if (fromAccount.id == toAccount.id) {
            throw ApiException(HttpStatus.BAD_REQUEST, "SAME_ACCOUNT_TRANSFER", "Source and destination accounts must differ")
        }

        val transaction = LedgerTransaction(
            reference = command.reference,
            fromAccount = fromAccount,
            toAccount = toAccount,
            amount = command.amount,
            correlationId = correlationId,
        )

        val savedTransaction = ledgerTransactionRepository.saveAndFlush(transaction)

        val debitEntry = JournalEntry(
            transaction = savedTransaction,
            account = fromAccount,
            entryType = EntryType.DEBIT,
            amount = command.amount,
        )
        val creditEntry = JournalEntry(
            transaction = savedTransaction,
            account = toAccount,
            entryType = EntryType.CREDIT,
            amount = command.amount,
        )

        enforceBalancedEntries(listOf(debitEntry, creditEntry))
        ledgerTransferPersistenceHook.beforeJournalPersist(savedTransaction, listOf(debitEntry, creditEntry))
        journalEntryRepository.saveAll(listOf(debitEntry, creditEntry))

        eventPublisher.publishEvent(
            TransactionCompletedEvent(
                transactionId = savedTransaction.id ?: error("transaction id must be assigned"),
                reference = savedTransaction.reference,
                fromAccount = fromAccount.accountNumber,
                toAccount = toAccount.accountNumber,
                amount = savedTransaction.amount,
                correlationId = correlationId,
                createdAt = savedTransaction.createdAt,
            ),
        )

        logger.info(
            "transfer_posted reference={} amount={} fromAccount={} toAccount={} correlationId={}",
            savedTransaction.reference,
            savedTransaction.amount,
            fromAccount.accountNumber,
            toAccount.accountNumber,
            correlationId,
        )

        return savedTransaction.toResponse(duplicate = false)
    }

    private fun enforceBalancedEntries(entries: List<JournalEntry>) {
        val debitTotal = entries
            .filter { it.entryType == EntryType.DEBIT }
            .fold(BigDecimal.ZERO) { total, entry -> total + entry.amount }
        val creditTotal = entries
            .filter { it.entryType == EntryType.CREDIT }
            .fold(BigDecimal.ZERO) { total, entry -> total + entry.amount }

        if (debitTotal.compareTo(creditTotal) != 0) {
            throw ApiException(HttpStatus.INTERNAL_SERVER_ERROR, "UNBALANCED_JOURNAL", "Debit and credit entries must balance")
        }
    }
}
