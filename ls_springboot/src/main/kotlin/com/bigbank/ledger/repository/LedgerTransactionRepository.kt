package com.bigbank.ledger.repository

import com.bigbank.ledger.domain.LedgerTransaction
import org.springframework.data.jpa.repository.EntityGraph
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.data.domain.Pageable
import java.util.UUID

interface LedgerTransactionRepository : JpaRepository<LedgerTransaction, UUID> {
    @EntityGraph(attributePaths = ["fromAccount", "toAccount"])
    fun findByReference(reference: String): LedgerTransaction?

    @EntityGraph(attributePaths = ["fromAccount", "toAccount"])
    fun findAllByOrderByCreatedAtDesc(pageable: Pageable): List<LedgerTransaction>
}
