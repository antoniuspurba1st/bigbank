package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"transaction-service/internal/model"
	"transaction-service/internal/repository"
	"transaction-service/internal/service"
)

func init() {
	sessionManager.sessions = make(map[string]*Session)
}

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

type idempotencyStub struct {
	startErr error
}

func (s *idempotencyStub) Start(key string) error {
	return s.startErr
}

func (s *idempotencyStub) Complete(key string) error {
	return nil
}

func (s *idempotencyStub) Fail(key string) error {
	return nil
}

func TestHandleTransferRejectsMalformedJSON(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	request := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(`{"reference":`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-Email", "user@example.com")
	request.Header.Set("Idempotency-Key", "test-key-malformed")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}

	apiErr := model.APIError{}
	if err := json.Unmarshal(response.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("expected valid api error json, got %v", err)
	}

	if apiErr.Error == "" {
		t.Fatalf("expected error message, got empty string")
	}
}

func TestHandleTransferRejectsUnknownField(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/transfer",
		strings.NewReader(`{"reference":"ref-1","from_account":"ACC-1","to_account":"ACC-2","amount":10,"unexpected":true}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-Email", "user@example.com")
	request.Header.Set("Idempotency-Key", "test-key-unknown-field")
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
		&idempotencyStub{},
		nil,
		&AuthHandler{},
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
		&idempotencyStub{},
		nil,
		&AuthHandler{},
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
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	request := httptest.NewRequest(http.MethodGet, "/transactions?page=-1", nil)
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestHandleTransferDuplicateIdempotencyKey(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{startErr: repository.ErrIdempotencyAlreadyExists},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"reference":"ref-1","from_account":"ACC-001","to_account":"ACC-002","amount":10}`
	request := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-Email", "user@example.com")
	request.Header.Set("Idempotency-Key", "duplicate-key")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", response.Code)
	}

	apiErr := model.APIError{}
	if err := json.Unmarshal(response.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("expected valid api error json, got %v", err)
	}

	if apiErr.Error != "Duplicate request" {
		t.Fatalf("expected error message 'Duplicate request', got %q", apiErr.Error)
	}
}

// PHASE 6: TRANSFER ENDPOINT TESTS

func TestHandleTransferSuccessful(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{
				decision: model.FraudDecision{Approved: true, Decision: "APPROVED", Reason: ""},
			},
			&ledgerTransfererStub{
				result: model.LedgerTransferResult{
					TransactionID: "tx-1",
					Reference:     "ref-1",
					Status:        "COMPLETED",
					Duplicate:     false,
				},
			},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"reference":"ref-1","from_account":"ACC-001","to_account":"ACC-002","amount":100.50}`
	request := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-Email", "user@example.com")
	request.Header.Set("Idempotency-Key", "unique-key-1")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	apiResp := model.APIResponse{}
	if err := json.Unmarshal(response.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("expected valid api response json, got %v", err)
	}

	if apiResp.Status != "success" {
		t.Fatalf("expected status 'success', got %q", apiResp.Status)
	}
}

func TestHandleTransferFraudRejection(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{
				decision: model.FraudDecision{Approved: false, Decision: "REJECTED", Reason: "High value transfer"},
			},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"reference":"ref-1","from_account":"ACC-001","to_account":"ACC-002","amount":50000.00}`
	request := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-Email", "user@example.com")
	request.Header.Set("Idempotency-Key", "unique-key-2")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 (rejection is not error), got %d", response.Code)
	}

	apiResp := model.APIResponse{}
	if err := json.Unmarshal(response.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("expected valid api response json, got %v", err)
	}

	if apiResp.Status != "rejected" {
		t.Fatalf("expected status 'rejected', got %q", apiResp.Status)
	}
}

func TestHandleTransferInsufficientBalance(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{
				decision: model.FraudDecision{Approved: true},
			},
			&ledgerTransfererStub{
				err: &model.AppError{
					StatusCode: http.StatusBadRequest,
					Code:       "INSUFFICIENT_FUNDS",
					Message:    "Insufficient funds",
				},
			},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"reference":"ref-1","from_account":"ACC-001","to_account":"ACC-002","amount":999999.99}`
	request := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-Email", "user@example.com")
	request.Header.Set("Idempotency-Key", "unique-key-3")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}

	apiErr := model.APIError{}
	if err := json.Unmarshal(response.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("expected valid api error json, got %v", err)
	}

	if apiErr.Error == "" {
		t.Fatalf("expected error message, got empty string")
	}
}

func TestHandleTransferMissingAuthHeaders(t *testing.T) {
	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"reference":"ref-1","from_account":"ACC-001","to_account":"ACC-002","amount":100}`
	request := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	// Intentionally omit X-Session-ID and X-User-Email
	request.Header.Set("Idempotency-Key", "unique-key-4")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestHandleTransferInvalidAmount(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"reference":"ref-1","from_account":"ACC-001","to_account":"ACC-002","amount":-50.00}`
	request := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-Email", "user@example.com")
	request.Header.Set("Idempotency-Key", "unique-key-5")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestHandleTransferLedgerServiceFailure(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{
				decision: model.FraudDecision{Approved: true},
			},
			&ledgerTransfererStub{
				err: &model.AppError{
					StatusCode: http.StatusServiceUnavailable,
					Code:       "LEDGER_SERVICE_UNAVAILABLE",
					Message:    "Ledger service request failed",
				},
			},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"reference":"ref-1","from_account":"ACC-001","to_account":"ACC-002","amount":100}`
	request := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-Email", "user@example.com")
	request.Header.Set("Idempotency-Key", "unique-key-6")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
}

// PHASE 6: TOPUP ENDPOINT TESTS

func TestHandleTopupSuccessful(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	userRepoStub := &userRepositoryStub{}
	userRepoStub.findAccountByUserIDResult.accountNumber = "ACC-001"
	userRepoStub.findAccountByUserIDResult.balance = 5000.00
	userRepoStub.findAccountByUserIDResult.err = nil

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{
				result: model.LedgerTransferResult{
					TransactionID: "topup-1",
					Reference:     "TOPUP-auto-1",
					Status:        "COMPLETED",
					Duplicate:     false,
				},
			},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{
			userRepo: userRepoStub,
		},
	)

	requestBody := `{"amount":1000.00}`
	request := httptest.NewRequest(http.MethodPost, "/topup", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-ID", "test-user-id")
	request.Header.Set("Idempotency-Key", "topup-key-1")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	apiResp := model.APIResponse{}
	if err := json.Unmarshal(response.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("expected valid api response json, got %v", err)
	}

	if apiResp.Status != "success" {
		t.Fatalf("expected status 'success', got %q", apiResp.Status)
	}
}

func TestHandleTopupInvalidAmount(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"amount":0}`
	request := httptest.NewRequest(http.MethodPost, "/topup", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-ID", "test-user-id")
	request.Header.Set("Idempotency-Key", "topup-key-2")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestHandleTopupAmountExceedsLimit(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"amount":99999999.99}`
	request := httptest.NewRequest(http.MethodPost, "/topup", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-ID", "test-user-id")
	request.Header.Set("Idempotency-Key", "topup-key-3")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestHandleTopupAccountNotFound(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	userRepoStub := &userRepositoryStub{}
	userRepoStub.findAccountByUserIDResult.err = repository.ErrUserNotFound

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{
			userRepo: userRepoStub,
		},
	)

	requestBody := `{"amount":100.00}`
	request := httptest.NewRequest(http.MethodPost, "/topup", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-ID", "unknown-user")
	request.Header.Set("Idempotency-Key", "topup-key-4")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestHandleTopupMissingAuthHeaders(t *testing.T) {
	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{},
	)

	requestBody := `{"amount":100.00}`
	request := httptest.NewRequest(http.MethodPost, "/topup", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	// Intentionally omit X-Session-ID, X-User-ID, and X-User-Email
	request.Header.Set("Idempotency-Key", "topup-key-5")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestHandleTopupDuplicateRequest(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	userRepoStub := &userRepositoryStub{}
	userRepoStub.findAccountByUserIDResult.accountNumber = "ACC-001"

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{startErr: repository.ErrIdempotencyAlreadyExists},
		nil,
		&AuthHandler{
			userRepo: userRepoStub,
		},
	)

	requestBody := `{"amount":100.00}`
	request := httptest.NewRequest(http.MethodPost, "/topup", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-ID", "test-user-id")
	request.Header.Set("Idempotency-Key", "topup-duplicate-key")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", response.Code)
	}
}

func TestHandleTopupServiceTimeout(t *testing.T) {
	sessionID := sessionManager.CreateSession("test-user-id", "user@example.com")

	userRepoStub := &userRepositoryStub{}
	userRepoStub.findAccountByUserIDResult.accountNumber = "ACC-001"

	handler := NewHTTPHandler(
		service.NewTransferService(
			&fraudCheckerStub{},
			&ledgerTransfererStub{
				err: &model.AppError{
					StatusCode: http.StatusGatewayTimeout,
					Code:       "SERVICE_TIMEOUT",
					Message:    "Request to ledger service timed out",
				},
			},
		),
		service.NewTransactionQueryService(&ledgerReaderStub{}),
		&idempotencyStub{},
		nil,
		&AuthHandler{
			userRepo: userRepoStub,
		},
	)

	requestBody := `{"amount":100.00}`
	request := httptest.NewRequest(http.MethodPost, "/topup", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Session-ID", sessionID)
	request.Header.Set("X-User-ID", "test-user-id")
	request.Header.Set("Idempotency-Key", "topup-key-6")
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", response.Code)
	}
}
