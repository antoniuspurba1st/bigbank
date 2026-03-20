package com.bigbank.ledger.event

import org.slf4j.LoggerFactory
import org.springframework.scheduling.annotation.Async
import org.springframework.stereotype.Component
import org.springframework.transaction.event.TransactionPhase
import org.springframework.transaction.event.TransactionalEventListener

@Component
class TransactionCompletedEventHandler {

    private val logger = LoggerFactory.getLogger(javaClass)

    @Async
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    fun onTransactionCompleted(event: TransactionCompletedEvent) {
        logger.info(
            "transaction_completed reference={} transactionId={} fromAccount={} toAccount={} amount={} correlationId={}",
            event.reference,
            event.transactionId,
            event.fromAccount,
            event.toAccount,
            event.amount,
            event.correlationId,
        )
    }
}
