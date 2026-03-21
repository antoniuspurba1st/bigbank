package com.bigbank.ledger.service

import com.bigbank.ledger.api.ApiErrorResponse
import com.bigbank.ledger.config.CorrelationIdFilter
import jakarta.servlet.http.HttpServletRequest
import org.slf4j.LoggerFactory
import org.springframework.dao.DataIntegrityViolationException
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.http.converter.HttpMessageNotReadableException
import org.springframework.web.bind.annotation.ExceptionHandler
import org.springframework.web.bind.annotation.RestControllerAdvice

@RestControllerAdvice
class GlobalExceptionHandler {
    private val logger = LoggerFactory.getLogger(javaClass)

    @ExceptionHandler(ApiException::class)
    fun handleApiException(
        ex: ApiException,
        request: HttpServletRequest,
    ): ResponseEntity<ApiErrorResponse> {
        logger.info(
            "api_exception code={} status={} message={}",
            ex.code,
            ex.httpStatus.value(),
            ex.message,
        )
        return ResponseEntity.status(ex.httpStatus).body(
            ApiErrorResponse(
                error = ex.message ?: "An error occurred",
            ),
        )
    }

    @ExceptionHandler(HttpMessageNotReadableException::class)
    fun handleMalformedRequest(request: HttpServletRequest): ResponseEntity<ApiErrorResponse> {
        return ResponseEntity.badRequest().body(
            ApiErrorResponse(
                error = "Request body is malformed or missing required fields",
            ),
        )
    }

    @ExceptionHandler(DataIntegrityViolationException::class)
    fun handleIntegrityViolation(request: HttpServletRequest): ResponseEntity<ApiErrorResponse> {
        return ResponseEntity.status(HttpStatus.CONFLICT).body(
            ApiErrorResponse(
                error = "The request violated a database constraint",
            ),
        )
    }

    @ExceptionHandler(Exception::class)
    fun handleUnexpectedException(request: HttpServletRequest): ResponseEntity<ApiErrorResponse> {
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(
            ApiErrorResponse(
                error = "An unexpected server error occurred",
            ),
        )
    }

    private fun correlationId(request: HttpServletRequest): String {
        return request.getAttribute(CorrelationIdFilter.CORRELATION_ID_ATTRIBUTE) as? String ?: "unknown"
    }
}
