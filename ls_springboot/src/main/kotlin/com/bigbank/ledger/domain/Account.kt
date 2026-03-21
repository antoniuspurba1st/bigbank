package com.bigbank.ledger.domain

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.GeneratedValue
import jakarta.persistence.GenerationType
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.Table
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

@Entity
@Table(
    name = "accounts",
    indexes = [
        Index(name = "idx_accounts_account_number", columnList = "account_number"),
    ],
)
class Account(
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    val id: UUID? = null,

    @Column(name = "account_number", nullable = false, unique = true, length = 64)
    val accountNumber: String,

    @Column(name = "owner_name", nullable = false, length = 128)
    val ownerName: String,

    @Column(name = "balance", nullable = false, precision = 19, scale = 2)
    var balance: BigDecimal = BigDecimal.ZERO,

    @Column(name = "created_at", nullable = false, updatable = false)
    val createdAt: Instant = Instant.now(),
)
