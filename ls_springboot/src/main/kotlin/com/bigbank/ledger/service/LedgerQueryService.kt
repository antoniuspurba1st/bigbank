package com.bigbank.ledger.service

import com.bigbank.ledger.api.TransactionListItem
import com.bigbank.ledger.repository.LedgerTransactionRepository
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional

@Service
class LedgerQueryService(
    private val ledgerTransactionRepository: LedgerTransactionRepository,
) {

    @Transactional(readOnly = true)
    fun listTransactions(): List<TransactionListItem> {
        return ledgerTransactionRepository.findAllWithAccountsOrderByCreatedAtDesc()
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
}
