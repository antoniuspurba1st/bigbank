package com.bigbank.ledger.repository

import com.bigbank.ledger.domain.Account
import org.springframework.data.jpa.repository.JpaRepository
import java.util.UUID

interface AccountRepository : JpaRepository<Account, UUID> {
    fun findByAccountNumber(accountNumber: String): Account?
}
