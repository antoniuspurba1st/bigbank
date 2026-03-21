package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"transaction-service/internal/model"
	"transaction-service/internal/repository"
)

// userRepositoryStub implements repository.UserRepository for testing
type userRepositoryStub struct {
	findByIDResult            *model.User
	findByIDError             error
	findAccountByUserIDResult struct {
		accountNumber string
		balance       float64
		err           error
	}
	findByEmailResult *model.User
	findByEmailError  error
}

func (stub *userRepositoryStub) Create(user *model.User) error {
	return nil
}

func (stub *userRepositoryStub) FindByEmail(email string) (*model.User, error) {
	return stub.findByEmailResult, stub.findByEmailError
}

func (stub *userRepositoryStub) FindByID(id string) (*model.User, error) {
	return stub.findByIDResult, stub.findByIDError
}

func (stub *userRepositoryStub) FindAccountByUserID(userID string) (string, float64, error) {
	return stub.findAccountByUserIDResult.accountNumber, stub.findAccountByUserIDResult.balance, stub.findAccountByUserIDResult.err
}

func (stub *userRepositoryStub) FindAccountByEmail(email string) (string, float64, error) {
	return "", 0, repository.ErrUserNotFound
}

func (stub *userRepositoryStub) Update(user *model.User) error {
	return nil
}

func TestHandleGetMe_ValidUser(t *testing.T) {
	// Setup mock repository
	stub := &userRepositoryStub{
		findByIDResult: &model.User{
			ID:        "user-123",
			Email:     "user@test.com",
			Phone:     "+1234567890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		findByIDError: nil,
		findAccountByUserIDResult: struct {
			accountNumber string
			balance       float64
			err           error
		}{
			accountNumber: "ACC-001",
			balance:       0.0,
			err:           nil,
		},
	}

	handler := NewAuthHandler(stub)

	// Create request with user ID header
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("X-User-ID", "user-123")
	req.Header.Set("X-Correlation-Id", "test-correlation-id")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleGetMe(w, req)

	// Assert status code
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Assert response structure
	if response["status"] != "success" {
		t.Errorf("expected status 'success', got %v", response["status"])
	}

	user, ok := response["user"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'user' field in response")
	}

	if user["id"] != "user-123" {
		t.Errorf("expected user id 'user-123', got %v", user["id"])
	}

	if user["email"] != "user@test.com" {
		t.Errorf("expected user email 'user@test.com', got %v", user["email"])
	}

	if user["phone"] != "+1234567890" {
		t.Errorf("expected user phone '+1234567890', got %v", user["phone"])
	}

	account, ok := response["account"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'account' field in response")
	}

	if account["account_number"] != "ACC-001" {
		t.Errorf("expected account number 'ACC-001', got %v", account["account_number"])
	}

	if account["balance"] != 0.0 {
		t.Errorf("expected balance 0.0, got %v", account["balance"])
	}
}

func TestHandleGetMe_UnknownUser(t *testing.T) {
	// Setup mock repository
	stub := &userRepositoryStub{
		findByIDResult: nil,
		findByIDError:  repository.ErrUserNotFound,
	}

	handler := NewAuthHandler(stub)

	// Create request with user ID header
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("X-User-ID", "unknown-user")
	req.Header.Set("X-Correlation-Id", "test-correlation-id")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleGetMe(w, req)

	// Assert status code
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Assert error response
	if response["error"] == nil {
		t.Errorf("expected error field, got none")
	}

	if response["error"] != "User not found" {
		t.Errorf("expected error 'User not found', got %v", response["error"])
	}
}

func TestHandleGetMe_NoSession(t *testing.T) {
	// Setup mock repository (not used in this test)
	stub := &userRepositoryStub{}

	handler := NewAuthHandler(stub)

	// Create request without authentication headers
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("X-Correlation-Id", "test-correlation-id")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleGetMe(w, req)

	// Assert status code
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Assert error response
	if response["error"] == nil {
		t.Errorf("expected error field, got none")
	}

	if response["error"] != "Authentication required" {
		t.Errorf("expected error 'Authentication required', got %v", response["error"])
	}
}
