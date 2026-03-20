package com.bigbank.ledger.service

import com.bigbank.ledger.api.ApiErrorResponse
import com.bigbank.ledger.config.CorrelationIdFilter
import jakarta.servlet.http.HttpServletRequest
import org.springframework.dao.DataIntegrityViolationException
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.http.converter.HttpMessageNotReadableException
import org.springframework.web.bind.annotation.ExceptionHandler
import org.springframework.web.bind.annotation.RestControllerAdvice

@RestControllerAdvice
class GlobalExceptionHandler {

    @ExceptionHandler(ApiException::class)
    fun handleApiException(
        ex: ApiException,
        request: HttpServletRequest,
    ): ResponseEntity<ApiErrorResponse> {
        return ResponseEntity.status(ex.httpStatus).body(
            ApiErrorResponse(
                code = ex.code,
                message = ex.message,
                correlationId = correlationId(request),
            ),
        )
    }

    @ExceptionHandler(HttpMessageNotReadableException::class)
    fun handleMalformedRequest(request: HttpServletRequest): ResponseEntity<ApiErrorResponse> {
        return ResponseEntity.badRequest().body(
            ApiErrorResponse(
                code = "MALFORMED_REQUEST",
                message = "Request body is malformed or missing required fields",
                correlationId = correlationId(request),
            ),
        )
    }

    @ExceptionHandler(DataIntegrityViolationException::class)
    fun handleIntegrityViolation(request: HttpServletRequest): ResponseEntity<ApiErrorResponse> {
        return ResponseEntity.status(HttpStatus.CONFLICT).body(
            ApiErrorResponse(
                code = "DATA_INTEGRITY_VIOLATION",
                message = "The request violated a database constraint",
                correlationId = correlationId(request),
            ),
        )
    }

    @ExceptionHandler(Exception::class)
    fun handleUnexpectedException(request: HttpServletRequest): ResponseEntity<ApiErrorResponse> {
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(
            ApiErrorResponse(
                code = "INTERNAL_ERROR",
                message = "Unexpected server error",
                correlationId = correlationId(request),
            ),
        )
    }

    private fun correlationId(request: HttpServletRequest): String {
        return request.getAttribute(CorrelationIdFilter.CORRELATION_ID_ATTRIBUTE) as? String ?: "unknown"
    }
}
