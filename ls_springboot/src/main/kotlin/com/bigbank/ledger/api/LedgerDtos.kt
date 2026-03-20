package com.bigbank.ledger.api

import tools.jackson.databind.PropertyNamingStrategies
import tools.jackson.databind.annotation.JsonNaming
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class LedgerTransferRequest(
    val reference: String?,
    val fromAccount: String?,
    val toAccount: String?,
    val amount: BigDecimal?,
)

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class LedgerTransferResponse(
    val transactionId: UUID,
    val reference: String,
    val fromAccount: String,
    val toAccount: String,
    val amount: BigDecimal,
    val status: String,
    val duplicate: Boolean,
    val createdAt: Instant,
)

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class TransactionListItem(
    val transactionId: UUID,
    val reference: String,
    val fromAccount: String,
    val toAccount: String,
    val amount: BigDecimal,
    val status: String,
    val createdAt: Instant,
)
