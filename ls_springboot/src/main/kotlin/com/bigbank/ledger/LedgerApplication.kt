package com.bigbank.ledger

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.scheduling.annotation.EnableAsync

@EnableAsync
@SpringBootApplication
class LedgerApplication

fun main(args: Array<String>) {
	runApplication<LedgerApplication>(*args)
}
