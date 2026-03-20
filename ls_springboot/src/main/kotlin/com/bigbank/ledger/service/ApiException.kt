package com.bigbank.ledger.service

import org.springframework.http.HttpStatus

class ApiException(
    val httpStatus: HttpStatus,
    val code: String,
    override val message: String,
) : RuntimeException(message)
