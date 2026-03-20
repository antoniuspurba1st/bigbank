package com.bigbank.ledger.event

interface TransactionCompletedSideEffect {
    fun handle(event: TransactionCompletedEvent)
}
