package com.bigbank.ledger.service

import com.bigbank.ledger.domain.JournalEntry
import com.bigbank.ledger.domain.LedgerTransaction
import org.springframework.stereotype.Component

interface LedgerTransferPersistenceHook {
    fun beforeJournalPersist(transaction: LedgerTransaction, entries: List<JournalEntry>)
}

@Component
class NoopLedgerTransferPersistenceHook : LedgerTransferPersistenceHook {
    override fun beforeJournalPersist(transaction: LedgerTransaction, entries: List<JournalEntry>) = Unit
}
