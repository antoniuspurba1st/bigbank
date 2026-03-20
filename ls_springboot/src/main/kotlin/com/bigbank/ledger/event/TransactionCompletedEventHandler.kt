package com.bigbank.ledger.event

import org.springframework.scheduling.annotation.Async
import org.springframework.stereotype.Component
import org.springframework.transaction.event.TransactionPhase
import org.springframework.transaction.event.TransactionalEventListener

@Component
class TransactionCompletedEventHandler(
    private val transactionCompletedSideEffect: TransactionCompletedSideEffect,
) {

    @Async
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    fun onTransactionCompleted(event: TransactionCompletedEvent) {
        transactionCompletedSideEffect.handle(event)
    }
}
