package com.bigbank.ledger.service

import com.bigbank.ledger.api.LedgerTransferResponse
import com.bigbank.ledger.domain.LedgerTransaction

internal fun LedgerTransaction.toResponse(duplicate: Boolean): LedgerTransferResponse {
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
