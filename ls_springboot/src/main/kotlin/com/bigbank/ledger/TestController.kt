package com.bigbank.ledger

import com.bigbank.ledger.api.ApiResponse
import com.bigbank.ledger.config.CorrelationIdFilter
import jakarta.servlet.http.HttpServletRequest
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RestController

@RestController
class HealthController {

    @GetMapping("/health")
    fun health(request: HttpServletRequest): ApiResponse<Map<String, String>> {
        val correlationId = request.getAttribute(CorrelationIdFilter.CORRELATION_ID_ATTRIBUTE) as String

        return ApiResponse(
            status = "success",
            message = "Ledger service is healthy",
            correlationId = correlationId,
            data = mapOf("status" to "UP", "service" to "ledger"),
        )
    }
}
