package com.bigbank.ledger.service

import com.bigbank.ledger.api.TransactionPageResponse
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
    fun listTransactions(page: Int = DEFAULT_PAGE, limit: Int = DEFAULT_LIMIT): TransactionPageResponse {
        val sanitizedPage = page.coerceAtLeast(0)
        val sanitizedLimit = limit.coerceIn(1, MAX_LIMIT)
        val transactionPage = ledgerTransactionRepository.findAllByOrderByCreatedAtDesc(
            PageRequest.of(sanitizedPage, sanitizedLimit),
        )

        return TransactionPageResponse(
            items = transactionPage.content
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
            },
            page = sanitizedPage,
            limit = sanitizedLimit,
            totalItems = transactionPage.totalElements,
            totalPages = transactionPage.totalPages,
            hasNext = transactionPage.hasNext(),
            hasPrevious = transactionPage.hasPrevious(),
        )
    }

    companion object {
        const val DEFAULT_PAGE = 0
        const val DEFAULT_LIMIT = 50
        const val MAX_LIMIT = 200
    }
}
