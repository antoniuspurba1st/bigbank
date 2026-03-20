package com.bigbank.ledger.repository

import com.bigbank.ledger.domain.LedgerTransaction
import org.springframework.data.jpa.repository.EntityGraph
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.data.jpa.repository.Query
import java.util.UUID

interface LedgerTransactionRepository : JpaRepository<LedgerTransaction, UUID> {
    @EntityGraph(attributePaths = ["fromAccount", "toAccount"])
    fun findByReference(reference: String): LedgerTransaction?

    @Query(
        """
        select transaction
        from LedgerTransaction transaction
        join fetch transaction.fromAccount
        join fetch transaction.toAccount
        order by transaction.createdAt desc
        """,
    )
    fun findAllWithAccountsOrderByCreatedAtDesc(): List<LedgerTransaction>
}
