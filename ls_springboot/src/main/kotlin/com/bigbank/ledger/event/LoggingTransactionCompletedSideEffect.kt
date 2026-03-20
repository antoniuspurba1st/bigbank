package com.bigbank.ledger.event

import org.slf4j.LoggerFactory
import org.springframework.stereotype.Component

@Component
class LoggingTransactionCompletedSideEffect : TransactionCompletedSideEffect {

    private val logger = LoggerFactory.getLogger(javaClass)

    override fun handle(event: TransactionCompletedEvent) {
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
