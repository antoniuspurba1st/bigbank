package com.bigbank.ledger.service

import java.math.BigDecimal

data class LedgerTransferCommand(
    val reference: String,
    val fromAccount: String,
    val toAccount: String,
    val amount: BigDecimal,
)
