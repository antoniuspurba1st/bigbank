package com.bigbank.ledger.domain

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.EnumType
import jakarta.persistence.Enumerated
import jakarta.persistence.FetchType
import jakarta.persistence.GeneratedValue
import jakarta.persistence.GenerationType
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.JoinColumn
import jakarta.persistence.ManyToOne
import jakarta.persistence.Table
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

@Entity
@Table(
    name = "ledger_transactions",
    indexes = [
        Index(name = "idx_ledger_transactions_created_at", columnList = "created_at"),
        Index(name = "idx_ledger_transactions_from_account", columnList = "from_account_id"),
        Index(name = "idx_ledger_transactions_to_account", columnList = "to_account_id"),
    ],
)
class LedgerTransaction(
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    val id: UUID? = null,

    @Column(name = "reference", nullable = false, unique = true, length = 128)
    val reference: String,

    @ManyToOne(fetch = FetchType.LAZY, optional = false)
    @JoinColumn(name = "from_account_id", nullable = false)
    val fromAccount: Account,

    @ManyToOne(fetch = FetchType.LAZY, optional = false)
    @JoinColumn(name = "to_account_id", nullable = false)
    val toAccount: Account,

    @Column(name = "amount", nullable = false, precision = 19, scale = 2)
    val amount: BigDecimal,

    @Enumerated(EnumType.STRING)
    @Column(name = "status", nullable = false, length = 32)
    val status: TransactionStatus = TransactionStatus.COMPLETED,

    @Column(name = "correlation_id", nullable = false, length = 128)
    val correlationId: String,

    @Column(name = "created_at", nullable = false, updatable = false)
    val createdAt: Instant = Instant.now(),
)
