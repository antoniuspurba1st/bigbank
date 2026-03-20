package com.bigbank.ledger.service

import com.bigbank.ledger.api.LedgerTransferRequest
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
import org.springframework.dao.DataIntegrityViolationException
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.math.BigDecimal
import java.math.RoundingMode

@Service
class LedgerCommandService(
    private val accountRepository: AccountRepository,
    private val ledgerTransactionRepository: LedgerTransactionRepository,
    private val journalEntryRepository: JournalEntryRepository,
    private val eventPublisher: ApplicationEventPublisher,
) {

    private val logger = LoggerFactory.getLogger(javaClass)

    @Transactional
    fun transfer(request: LedgerTransferRequest, correlationId: String): LedgerTransferResponse {
        val command = normalize(request)
        validate(command)

        ledgerTransactionRepository.findByReference(command.reference)?.let { existing ->
            return toIdempotentResponse(existing, command)
        }

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

        val savedTransaction = try {
            ledgerTransactionRepository.saveAndFlush(transaction)
        } catch (_: DataIntegrityViolationException) {
            val existing = ledgerTransactionRepository.findByReference(command.reference)
                ?: throw ApiException(HttpStatus.CONFLICT, "REFERENCE_CONFLICT", "Transfer reference already exists")
            return toIdempotentResponse(existing, command)
        }

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

    private fun normalize(request: LedgerTransferRequest): TransferCommand {
        val reference = request.reference?.trim().orEmpty()
        val fromAccount = request.fromAccount?.trim().orEmpty()
        val toAccount = request.toAccount?.trim().orEmpty()
        val amount = request.amount?.setScale(2, RoundingMode.HALF_UP)
            ?: throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_AMOUNT", "Amount is required")

        return TransferCommand(
            reference = reference,
            fromAccount = fromAccount,
            toAccount = toAccount,
            amount = amount,
        )
    }

    private fun validate(command: TransferCommand) {
        if (command.reference.isBlank()) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_REFERENCE", "Reference is required")
        }
        if (command.reference.length > 128) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_REFERENCE", "Reference is too long")
        }
        if (command.fromAccount.isBlank() || command.toAccount.isBlank()) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_ACCOUNT", "Both accounts are required")
        }
        if (command.fromAccount == command.toAccount) {
            throw ApiException(HttpStatus.BAD_REQUEST, "SAME_ACCOUNT_TRANSFER", "Source and destination accounts must differ")
        }
        if (command.amount <= BigDecimal.ZERO) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_AMOUNT", "Amount must be greater than zero")
        }
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

    private fun toIdempotentResponse(
        existing: LedgerTransaction,
        command: TransferCommand,
    ): LedgerTransferResponse {
        if (existing.fromAccount.accountNumber != command.fromAccount ||
            existing.toAccount.accountNumber != command.toAccount ||
            existing.amount.compareTo(command.amount) != 0
        ) {
            throw ApiException(HttpStatus.CONFLICT, "REFERENCE_CONFLICT", "Reference already used for a different transfer")
        }

        return existing.toResponse(duplicate = true)
    }

    private fun LedgerTransaction.toResponse(duplicate: Boolean): LedgerTransferResponse {
        return LedgerTransferResponse(
            transactionId = id ?: error("transaction id must be assigned"),
            reference = reference,
            fromAccount = fromAccount.accountNumber,
            toAccount = toAccount.accountNumber,
            amount = amount,
            status = status.name,
            duplicate = duplicate,
            createdAt = createdAt,
        )
    }

    private data class TransferCommand(
        val reference: String,
        val fromAccount: String,
        val toAccount: String,
        val amount: BigDecimal,
    )
}
