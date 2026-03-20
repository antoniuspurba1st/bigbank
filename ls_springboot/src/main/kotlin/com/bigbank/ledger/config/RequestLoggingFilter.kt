package com.bigbank.ledger.config

import jakarta.servlet.FilterChain
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import org.slf4j.LoggerFactory
import org.springframework.core.Ordered
import org.springframework.core.annotation.Order
import org.springframework.stereotype.Component
import org.springframework.web.filter.OncePerRequestFilter
import java.util.concurrent.atomic.AtomicLong

@Component
@Order(Ordered.LOWEST_PRECEDENCE)
class RequestLoggingFilter : OncePerRequestFilter() {

    private val requestLogger = LoggerFactory.getLogger(javaClass)
    private val requestCount = AtomicLong(0)
    private val errorCount = AtomicLong(0)

    override fun doFilterInternal(
        request: HttpServletRequest,
        response: HttpServletResponse,
        filterChain: FilterChain,
    ) {
        val startedAt = System.nanoTime()

        try {
            filterChain.doFilter(request, response)
        } finally {
            val totalRequests = requestCount.incrementAndGet()
            val totalErrors = if (response.status >= 400) {
                errorCount.incrementAndGet()
            } else {
                errorCount.get()
            }
            val latencyMs = (System.nanoTime() - startedAt) / 1_000_000
            val errorRate = totalErrors.toDouble() / totalRequests.toDouble()
            val correlationId = request.getAttribute(CorrelationIdFilter.CORRELATION_ID_ATTRIBUTE) as? String ?: "unknown"

            requestLogger.info(
                "request_completed method={} path={} status={} latencyMs={} correlationId={} requestCount={} errorCount={} errorRate={}",
                request.method,
                request.requestURI,
                response.status,
                latencyMs,
                correlationId,
                totalRequests,
                totalErrors,
                String.format("%.4f", errorRate),
            )
        }
    }
}
