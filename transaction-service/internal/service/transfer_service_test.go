package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"transaction-service/internal/model"
)

type fraudCheckerStub struct {
	decision model.FraudDecision
	err      *model.AppError
	calls    int
}

func (stub *fraudCheckerStub) Check(_ context.Context, _ string, _ model.TransferRequest) (model.FraudDecision, *model.AppError) {
	stub.calls++
	return stub.decision, stub.err
}

type ledgerTransfererStub struct {
	result model.LedgerTransferResult
	err    *model.AppError
	calls  int
}

func (stub *ledgerTransfererStub) Transfer(_ context.Context, _ string, _ model.TransferRequest) (model.LedgerTransferResult, *model.AppError) {
	stub.calls++
	return stub.result, stub.err
}

func TestExecuteApprovedTransfer(t *testing.T) {
	fraud := &fraudCheckerStub{
		decision: model.FraudDecision{
			Decision: "approved",
			Approved: true,
			Reason:   "ok",
		},
	}
	ledger := &ledgerTransfererStub{
		result: model.LedgerTransferResult{
			TransactionID: "tx-1",
			Reference:     "ref-1",
			Amount:        100,
			Status:        "COMPLETED",
			Duplicate:     false,
		},
	}

	service := NewTransferService(fraud, ledger)
	response, err := service.Execute(context.Background(), "corr-1", model.TransferRequest{
		Reference:   "ref-1",
		FromAccount: "ACC-1",
		ToAccount:   "ACC-2",
		Amount:      100,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Status != "success" {
		t.Fatalf("expected success status, got %s", response.Status)
	}

	if fraud.calls != 1 {
		t.Fatalf("expected fraud client to be called once, got %d", fraud.calls)
	}

	if ledger.calls != 1 {
		t.Fatalf("expected ledger client to be called once, got %d", ledger.calls)
	}
}

func TestExecuteRejectedTransferSkipsLedger(t *testing.T) {
	fraud := &fraudCheckerStub{
		decision: model.FraudDecision{
			Decision: "rejected",
			Approved: false,
			Reason:   "too large",
		},
	}
	ledger := &ledgerTransfererStub{}

	service := NewTransferService(fraud, ledger)
	response, err := service.Execute(context.Background(), "corr-2", model.TransferRequest{
		Reference:   "ref-2",
		FromAccount: "ACC-1",
		ToAccount:   "ACC-2",
		Amount:      2000000,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Status != "rejected" {
		t.Fatalf("expected rejected status, got %s", response.Status)
	}

	if ledger.calls != 0 {
		t.Fatalf("expected ledger client not to be called, got %d", ledger.calls)
	}
}

func TestTopupCallsLedgerDirectly(t *testing.T) {
	ledger := &ledgerTransfererStub{
		result: model.LedgerTransferResult{
			TransactionID: "tx-topup-1",
			Reference:     "TOPUP-ref",
			Amount:        200,
			Status:        "COMPLETED",
			Duplicate:     false,
		},
	}

	service := NewTransferService(&fraudCheckerStub{}, ledger)
	response, err := service.Topup(context.Background(), "corr-5", "ACC-1", 200)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Status != "success" {
		t.Fatalf("expected success status, got %s", response.Status)
	}

	if ledger.calls != 1 {
		t.Fatalf("expected ledger client to be called once, got %d", ledger.calls)
	}
}

func TestTopupAcceptsLongAccountNumber(t *testing.T) {
	ledger := &ledgerTransfererStub{
		result: model.LedgerTransferResult{
			TransactionID: "tx-topup-2",
			Reference:     "TOPUP-ref-2",
			Amount:        150,
			Status:        "COMPLETED",
			Duplicate:     false,
		},
	}

	service := NewTransferService(&fraudCheckerStub{}, ledger)
	longAccount := "ACC-" + "01234567-89ab-cdef-0123-456789abcdef"
	response, err := service.Topup(context.Background(), "corr-6", longAccount, 150)

	if err != nil {
		t.Fatalf("expected no error for long account, got %v", err)
	}

	if response.Status != "success" {
		t.Fatalf("expected success status, got %s", response.Status)
	}

	if ledger.calls != 1 {
		t.Fatalf("expected ledger client to be called once, got %d", ledger.calls)
	}
}

func TestTopupRejectsInvalidAmount(t *testing.T) {
	service := NewTransferService(&fraudCheckerStub{}, &ledgerTransfererStub{})

	_, err := service.Topup(context.Background(), "corr-6", "ACC-1", 0)

	if err == nil {
		t.Fatal("expected validation error")
	}

	if err.Code != "INVALID_AMOUNT" {
		t.Fatalf("expected INVALID_AMOUNT, got %s", err.Code)
	}
}

func TestExecuteRejectsInvalidAmount(t *testing.T) {
	service := NewTransferService(&fraudCheckerStub{}, &ledgerTransfererStub{})

	_, err := service.Execute(context.Background(), "corr-3", model.TransferRequest{
		Reference:   "ref-3",
		FromAccount: "ACC-1",
		ToAccount:   "ACC-2",
		Amount:      0,
	})

	if err == nil {
		t.Fatal("expected validation error")
	}

	if err.Code != "INVALID_AMOUNT" {
		t.Fatalf("expected INVALID_AMOUNT, got %s", err.Code)
	}
}

func TestExecutePropagatesLedgerFailure(t *testing.T) {
	fraud := &fraudCheckerStub{
		decision: model.FraudDecision{
			Decision: "approved",
			Approved: true,
			Reason:   "ok",
		},
	}
	ledger := &ledgerTransfererStub{
		err: &model.AppError{
			StatusCode: http.StatusBadGateway,
			Code:       "LEDGER_FAILED",
			Message:    "ledger exploded",
			Err:        errors.New("boom"),
		},
	}

	service := NewTransferService(fraud, ledger)
	_, err := service.Execute(context.Background(), "corr-4", model.TransferRequest{
		Reference:   "ref-4",
		FromAccount: "ACC-1",
		ToAccount:   "ACC-2",
		Amount:      100,
	})

	if err == nil {
		t.Fatal("expected ledger error")
	}

	if err.Code != "LEDGER_FAILED" {
		t.Fatalf("expected LEDGER_FAILED, got %s", err.Code)
	}
}
