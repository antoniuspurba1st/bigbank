package com.bigbank.ledger.api

import tools.jackson.databind.PropertyNamingStrategies
import tools.jackson.databind.annotation.JsonNaming

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class ApiErrorResponse(
    val status: String = "error",
    val code: String,
    val message: String,
    val correlationId: String,
)
