package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"transaction-service/internal/model"
	"transaction-service/internal/service"
)

type fraudCheckerStub struct {
	decision model.FraudDecision
	err      *model.AppError
}

func (stub *fraudCheckerStub) Check(_ context.Context, _ string, _ model.TransferRequest) (model.FraudDecision, *model.AppError) {
	return stub.decision, stub.err
}

type ledgerTransfererStub struct {
	result model.LedgerTransferResult
	err    *model.AppError
}

func (stub *ledgerTransfererStub) Transfer(_ context.Context, _ string, _ model.TransferRequest) (model.LedgerTransferResult, *model.AppError) {
	return stub.result, stub.err
}

type ledgerReaderStub struct {
	page model.TransactionHistoryPage
	err  *model.AppError
}

func (stub *ledgerReaderStub) ListTransactions(_ context.Context, _ string, _ int, _ int) (model.TransactionHistoryPage, *model.AppError) {
	return stub.page, stub.err
}

func TestHandleTransferRejectsMalformedJSON(t *testing.T) {
	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
	)

	request := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(`{"reference":`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}

	apiErr := model.APIError{}
	if err := json.Unmarshal(response.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("expected valid api error json, got %v", err)
	}

	if apiErr.Code != "MALFORMED_REQUEST" {
		t.Fatalf("expected MALFORMED_REQUEST, got %s", apiErr.Code)
	}
}

func TestHandleTransferRejectsUnknownField(t *testing.T) {
	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/transfer",
		strings.NewReader(`{"reference":"ref-1","from_account":"ACC-1","to_account":"ACC-2","amount":10,"unexpected":true}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestHandleHealthSetsCorrelationID(t *testing.T) {
	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
	)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	if response.Header().Get("X-Correlation-Id") == "" {
		t.Fatal("expected X-Correlation-Id header")
	}
}

func TestHandleTransactionsReturnsPagedHistory(t *testing.T) {
	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{
			page: model.TransactionHistoryPage{
				Items: []model.TransactionHistoryItem{
					{
						TransactionID: "tx-1",
						Reference:     "ref-1",
						FromAccount:   "ACC-001",
						ToAccount:     "ACC-002",
						Amount:        42.5,
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
		}),
	)

	request := httptest.NewRequest(http.MethodGet, "/transactions?page=0&limit=10", nil)
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	apiResponse := model.APIResponse{}
	if err := json.Unmarshal(response.Body.Bytes(), &apiResponse); err != nil {
		t.Fatalf("expected valid api response json, got %v", err)
	}
}

func TestHandleTransactionsRejectsNegativePage(t *testing.T) {
	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
	)

	request := httptest.NewRequest(http.MethodGet, "/transactions?page=-1", nil)
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}
