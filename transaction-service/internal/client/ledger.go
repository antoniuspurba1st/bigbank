package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"transaction-service/internal/model"
)

type LedgerClient struct {
	httpClient *jsonHTTPClient
}

func NewLedgerClient(baseURL string, timeout time.Duration, maxAttempts int) *LedgerClient {
	return &LedgerClient{
		httpClient: newJSONHTTPClient(baseURL, timeout, maxAttempts),
	}
}

func (c *LedgerClient) Transfer(
	ctx context.Context,
	correlationID string,
	request model.TransferRequest,
) (model.LedgerTransferResult, *model.AppError) {
	payload := model.LedgerTransferRequest{
		Reference:   strings.TrimSpace(request.Reference),
		FromAccount: strings.TrimSpace(request.FromAccount),
		ToAccount:   strings.TrimSpace(request.ToAccount),
		Amount:      request.Amount,
	}

	response := model.LedgerTransferEnvelope{}
	if err := c.httpClient.postJSON(ctx, "/ledger/transfer", correlationID, payload, &response); err != nil {
		err.Message = "Ledger service request failed"
		if err.StatusCode == http.StatusServiceUnavailable {
			err.Code = "LEDGER_SERVICE_UNAVAILABLE"
		}

		return model.LedgerTransferResult{}, err
	}

	if response.Data == nil {
		return model.LedgerTransferResult{}, &model.AppError{
			StatusCode: http.StatusBadGateway,
			Code:       "LEDGER_RESPONSE_INVALID",
			Message:    "Ledger service returned an empty transaction",
		}
	}

	return *response.Data, nil
}

func (c *LedgerClient) ListTransactions(
	ctx context.Context,
	correlationID string,
	page int,
	limit int,
) (model.TransactionHistoryPage, *model.AppError) {
	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("limit", strconv.Itoa(limit))

	response := model.LedgerTransactionsEnvelope{}
	if err := c.httpClient.getJSON(ctx, "/ledger/transactions", correlationID, query, &response); err != nil {
		err.Message = "Ledger service request failed"
		if err.StatusCode == http.StatusServiceUnavailable {
			err.Code = "LEDGER_SERVICE_UNAVAILABLE"
		}

		return model.TransactionHistoryPage{}, err
	}

	if response.Data == nil {
		return model.TransactionHistoryPage{}, &model.AppError{
			StatusCode: http.StatusBadGateway,
			Code:       "LEDGER_RESPONSE_INVALID",
			Message:    "Ledger service returned an empty transaction page",
		}
	}

	return *response.Data, nil
}
