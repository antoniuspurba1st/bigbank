package com.bigbank.ledger.service

import com.bigbank.ledger.api.TransactionListItem
import com.bigbank.ledger.repository.LedgerTransactionRepository
import org.springframework.data.domain.PageRequest
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional

@Service
class LedgerQueryService(
    private val ledgerTransactionRepository: LedgerTransactionRepository,
) {

    @Transactional(readOnly = true)
    fun listTransactions(limit: Int = DEFAULT_LIMIT): List<TransactionListItem> {
        val sanitizedLimit = limit.coerceIn(1, MAX_LIMIT)

        return ledgerTransactionRepository.findAllByOrderByCreatedAtDesc(PageRequest.of(0, sanitizedLimit))
            .map { transaction ->
                TransactionListItem(
                    transactionId = transaction.id ?: error("transaction id must be assigned"),
                    reference = transaction.reference,
                    fromAccount = transaction.fromAccount.accountNumber,
                    toAccount = transaction.toAccount.accountNumber,
                    amount = transaction.amount,
                    status = transaction.status.name,
                    createdAt = transaction.createdAt,
                )
            }
    }

    companion object {
        const val DEFAULT_LIMIT = 50
        const val MAX_LIMIT = 200
    }
}
