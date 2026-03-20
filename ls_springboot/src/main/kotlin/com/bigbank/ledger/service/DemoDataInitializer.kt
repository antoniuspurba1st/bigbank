package com.bigbank.ledger.service

import com.bigbank.ledger.domain.Account
import com.bigbank.ledger.repository.AccountRepository
import org.slf4j.LoggerFactory
import org.springframework.boot.CommandLineRunner
import org.springframework.context.annotation.Profile
import org.springframework.stereotype.Component

@Component
@Profile("!test")
class DemoDataInitializer(
    private val accountRepository: AccountRepository,
) : CommandLineRunner {

    private val logger = LoggerFactory.getLogger(javaClass)

    override fun run(vararg args: String) {
        val demoAccounts = listOf(
            Account(accountNumber = "ACC-001", ownerName = "Alice"),
            Account(accountNumber = "ACC-002", ownerName = "Bob"),
            Account(accountNumber = "ACC-003", ownerName = "Charlie"),
        )

        val missingAccounts = demoAccounts.filter { account ->
            accountRepository.findByAccountNumber(account.accountNumber) == null
        }

        if (missingAccounts.isEmpty()) {
            return
        }

        accountRepository.saveAll(missingAccounts)
        logger.info("seeded {} demo ledger accounts", missingAccounts.size)
    }
}
