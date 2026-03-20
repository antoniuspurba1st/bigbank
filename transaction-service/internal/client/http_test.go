package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"transaction-service/internal/model"
)

func TestPostJSONRetriesServerErrorUntilSuccess(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentAttempt := attempts.Add(1)
		if currentAttempt == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","code":"TEMPORARY_FAILURE","message":"retry me","correlation_id":"corr-1"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","message":"ok","correlation_id":"corr-1","data":{"decision":"approved","approved":true,"reason":"ok","checked_at":"now"}}`))
	}))
	defer server.Close()

	httpClient := newJSONHTTPClient(server.URL, time.Second, 1)
	response := model.FraudCheckResponse{}

	err := httpClient.postJSON(
		context.Background(),
		"/fraud/check",
		"corr-1",
		map[string]any{"reference": "ref-1"},
		&response,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if attempts.Load() != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts.Load())
	}

	if response.Data == nil || !response.Data.Approved {
		t.Fatal("expected approved downstream response")
	}
}

func TestPostJSONTimeoutReturnsUnavailable(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		time.Sleep(120 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","message":"ok","correlation_id":"corr-2","data":{"decision":"approved","approved":true,"reason":"ok","checked_at":"now"}}`))
	}))
	defer server.Close()

	httpClient := newJSONHTTPClient(server.URL, 20*time.Millisecond, 2)
	response := model.FraudCheckResponse{}

	err := httpClient.postJSON(
		context.Background(),
		"/fraud/check",
		"corr-2",
		map[string]any{"reference": "ref-2"},
		&response,
	)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if err.Code != "DOWNSTREAM_UNAVAILABLE" {
		t.Fatalf("expected DOWNSTREAM_UNAVAILABLE, got %s", err.Code)
	}
}

func TestPostJSONServerErrorStopsAfterConfiguredRetries(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"status":"error","code":"TEMPORARY_FAILURE","message":"retry me","correlation_id":"corr-5"}`))
	}))
	defer server.Close()

	httpClient := newJSONHTTPClient(server.URL, time.Second, 2)
	response := model.FraudCheckResponse{}

	err := httpClient.postJSON(
		context.Background(),
		"/fraud/check",
		"corr-5",
		map[string]any{"reference": "ref-5"},
		&response,
	)
	if err == nil {
		t.Fatal("expected server error after retries")
	}

	if attempts.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestFraudClientMapsUnavailableError(t *testing.T) {
	fraudClient := NewFraudClient("http://127.0.0.1:1", 50*time.Millisecond, 1)

	_, err := fraudClient.Check(context.Background(), "corr-3", model.TransferRequest{
		Reference:   "ref-3",
		FromAccount: "ACC-1",
		ToAccount:   "ACC-2",
		Amount:      12.34,
	})
	if err == nil {
		t.Fatal("expected unavailable error")
	}

	if err.Code != "FRAUD_SERVICE_UNAVAILABLE" {
		t.Fatalf("expected FRAUD_SERVICE_UNAVAILABLE, got %s", err.Code)
	}
}

func TestLedgerClientRejectsEmptyResponseData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","message":"ok","correlation_id":"corr-4","data":null}`))
	}))
	defer server.Close()

	ledgerClient := NewLedgerClient(server.URL, time.Second, 0)

	_, err := ledgerClient.Transfer(context.Background(), "corr-4", model.TransferRequest{
		Reference:   "ref-4",
		FromAccount: "ACC-1",
		ToAccount:   "ACC-2",
		Amount:      99.99,
	})
	if err == nil {
		t.Fatal("expected invalid response error")
	}

	if err.Code != "LEDGER_RESPONSE_INVALID" {
		t.Fatalf("expected LEDGER_RESPONSE_INVALID, got %s", err.Code)
	}
}
