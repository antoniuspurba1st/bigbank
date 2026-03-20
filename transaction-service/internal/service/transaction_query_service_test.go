package service

import (
	"context"
	"net/http"
	"testing"

	"transaction-service/internal/model"
)

type ledgerReaderStub struct {
	page  model.TransactionHistoryPage
	err   *model.AppError
	calls int
}

func (stub *ledgerReaderStub) ListTransactions(_ context.Context, _ string, _ int, _ int) (model.TransactionHistoryPage, *model.AppError) {
	stub.calls++
	return stub.page, stub.err
}

func TestListTransactionsReturnsPagedEnvelope(t *testing.T) {
	ledger := &ledgerReaderStub{
		page: model.TransactionHistoryPage{
			Items: []model.TransactionHistoryItem{
				{
					TransactionID: "tx-1",
					Reference:     "ref-1",
					FromAccount:   "ACC-001",
					ToAccount:     "ACC-002",
					Amount:        12.34,
					Status:        "COMPLETED",
					CreatedAt:     "2026-03-20T10:00:00Z",
				},
			},
			Page:        0,
			Limit:       10,
			TotalItems:  1,
			TotalPages:  1,
			HasNext:     false,
			HasPrevious: false,
		},
	}

	service := NewTransactionQueryService(ledger)
	response, err := service.ListTransactions(context.Background(), "corr-query", 0, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Status != "success" {
		t.Fatalf("expected success status, got %s", response.Status)
	}

	if ledger.calls != 1 {
		t.Fatalf("expected ledger reader to be called once, got %d", ledger.calls)
	}
}

func TestListTransactionsPropagatesLedgerFailure(t *testing.T) {
	service := NewTransactionQueryService(&ledgerReaderStub{
		err: &model.AppError{
			StatusCode: http.StatusBadGateway,
			Code:       "LEDGER_RESPONSE_INVALID",
			Message:    "bad response",
		},
	})

	_, err := service.ListTransactions(context.Background(), "corr-query-err", 0, 10)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Code != "LEDGER_RESPONSE_INVALID" {
		t.Fatalf("expected LEDGER_RESPONSE_INVALID, got %s", err.Code)
	}
}
