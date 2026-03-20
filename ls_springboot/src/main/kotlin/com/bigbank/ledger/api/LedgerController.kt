package com.bigbank.ledger.api

import com.bigbank.ledger.config.CorrelationIdFilter
import com.bigbank.ledger.service.LedgerCommandService
import com.bigbank.ledger.service.LedgerQueryService
import jakarta.servlet.http.HttpServletRequest
import org.springframework.http.HttpStatus
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.ResponseStatus
import org.springframework.web.bind.annotation.RestController

@RestController
@RequestMapping("/ledger")
class LedgerController(
    private val ledgerCommandService: LedgerCommandService,
    private val ledgerQueryService: LedgerQueryService,
) {

    @PostMapping("/transfer")
    @ResponseStatus(HttpStatus.OK)
    fun transfer(
        @RequestBody request: LedgerTransferRequest,
        httpRequest: HttpServletRequest,
    ): ApiResponse<LedgerTransferResponse> {
        val correlationId = httpRequest.getAttribute(CorrelationIdFilter.CORRELATION_ID_ATTRIBUTE) as String
        val result = ledgerCommandService.transfer(request, correlationId)

        return ApiResponse(
            status = "success",
            message = if (result.duplicate) {
                "Duplicate request returned existing transaction"
            } else {
                "Transfer posted successfully"
            },
            correlationId = correlationId,
            data = result,
        )
    }

    @GetMapping("/transactions")
    fun transactions(
        httpRequest: HttpServletRequest,
        @RequestParam(name = "limit", required = false) limit: Int?,
    ): ApiResponse<List<TransactionListItem>> {
        val correlationId = httpRequest.getAttribute(CorrelationIdFilter.CORRELATION_ID_ATTRIBUTE) as String

        return ApiResponse(
            status = "success",
            message = "Transactions fetched successfully",
            correlationId = correlationId,
            data = ledgerQueryService.listTransactions(limit ?: LedgerQueryService.DEFAULT_LIMIT),
        )
    }
}
