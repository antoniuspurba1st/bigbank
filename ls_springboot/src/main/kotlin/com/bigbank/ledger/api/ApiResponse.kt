package com.bigbank.ledger.api

import tools.jackson.databind.PropertyNamingStrategies
import tools.jackson.databind.annotation.JsonNaming

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class ApiResponse<T>(
    val status: String,
    val message: String,
    val correlationId: String,
    val data: T? = null,
)
