package com.bigbank.ledger.service

import com.bigbank.ledger.api.LedgerTransferRequest
import com.bigbank.ledger.api.LedgerTransferResponse
import com.bigbank.ledger.domain.LedgerTransaction
import com.bigbank.ledger.repository.AccountRepository
import com.bigbank.ledger.repository.LedgerTransactionRepository
import org.springframework.dao.DataAccessException
import org.springframework.http.HttpStatus
import org.springframework.orm.jpa.JpaSystemException
import org.springframework.stereotype.Service
import java.math.BigDecimal
import java.math.RoundingMode

@Service
class LedgerCommandService(
    private val accountRepository: AccountRepository,
    private val ledgerTransactionRepository: LedgerTransactionRepository,
    private val ledgerPostingExecutor: LedgerPostingExecutor,
) {
    private val referencePattern = Regex("^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$")
    private val accountPattern = Regex("^[A-Z0-9-]{3,32}$")

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

        return try {
            ledgerPostingExecutor.postTransfer(command, correlationId)
        } catch (ex: DataAccessException) {
            resolvePersistenceConflict(command, ex)
        } catch (ex: JpaSystemException) {
            resolvePersistenceConflict(command, ex)
        }
    }

    private fun normalize(request: LedgerTransferRequest): LedgerTransferCommand {
        val reference = request.reference?.trim().orEmpty()
        val fromAccount = request.fromAccount?.trim().orEmpty()
        val toAccount = request.toAccount?.trim().orEmpty()
        val amount = request.amount?.setScale(2, RoundingMode.HALF_UP)
            ?: throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_AMOUNT", "Amount is required")

        return LedgerTransferCommand(
            reference = reference,
            fromAccount = fromAccount,
            toAccount = toAccount,
            amount = amount,
        )
    }

    private fun validate(command: LedgerTransferCommand) {
        if (command.reference.isBlank()) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_REFERENCE", "Reference is required")
        }
        if (command.reference.length > 128) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_REFERENCE", "Reference is too long")
        }
        if (!referencePattern.matches(command.reference)) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_REFERENCE", "Reference contains unsupported characters")
        }
        if (command.fromAccount.isBlank() || command.toAccount.isBlank()) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_ACCOUNT", "Both accounts are required")
        }
        if (!accountPattern.matches(command.fromAccount) || !accountPattern.matches(command.toAccount)) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_ACCOUNT", "Account format is invalid")
        }
        if (command.fromAccount == command.toAccount) {
            throw ApiException(HttpStatus.BAD_REQUEST, "SAME_ACCOUNT_TRANSFER", "Source and destination accounts must differ")
        }
        if (command.amount <= BigDecimal.ZERO) {
            throw ApiException(HttpStatus.BAD_REQUEST, "INVALID_AMOUNT", "Amount must be greater than zero")
        }
    }

    private fun toIdempotentResponse(
        existing: LedgerTransaction,
        command: LedgerTransferCommand,
    ): LedgerTransferResponse {
        if (existing.fromAccount.accountNumber != command.fromAccount ||
            existing.toAccount.accountNumber != command.toAccount ||
            existing.amount.compareTo(command.amount) != 0
        ) {
            throw ApiException(HttpStatus.CONFLICT, "REFERENCE_CONFLICT", "Reference already used for a different transfer")
        }

        return existing.toResponse(duplicate = true)
    }

    private fun resolvePersistenceConflict(
        command: LedgerTransferCommand,
        ex: RuntimeException,
    ): LedgerTransferResponse {
        val existing = ledgerTransactionRepository.findByReference(command.reference)
            ?: throw ex

        return toIdempotentResponse(existing, command)
    }
}
