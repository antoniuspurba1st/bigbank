package com.bigbank.ledger.event

import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

data class TransactionCompletedEvent(
    val transactionId: UUID,
    val reference: String,
    val fromAccount: String,
    val toAccount: String,
    val amount: BigDecimal,
    val correlationId: String,
    val createdAt: Instant,
)
