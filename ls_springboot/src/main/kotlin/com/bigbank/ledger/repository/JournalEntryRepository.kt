package com.bigbank.ledger.repository

import com.bigbank.ledger.domain.JournalEntry
import org.springframework.data.jpa.repository.JpaRepository
import java.util.UUID

interface JournalEntryRepository : JpaRepository<JournalEntry, UUID> {
    fun findAllByTransactionId(transactionId: UUID): List<JournalEntry>
}
