package service

import (
	"context"
	"log"
	"net/http"
	"strings"

	"transaction-service/internal/model"
)

type fraudChecker interface {
	Check(ctx context.Context, correlationID string, request model.TransferRequest) (model.FraudDecision, *model.AppError)
}

type ledgerTransferer interface {
	Transfer(ctx context.Context, correlationID string, request model.TransferRequest) (model.LedgerTransferResult, *model.AppError)
}

type TransferService struct {
	fraudClient  fraudChecker
	ledgerClient ledgerTransferer
}

func NewTransferService(fraudClient fraudChecker, ledgerClient ledgerTransferer) *TransferService {
	return &TransferService{
		fraudClient:  fraudClient,
		ledgerClient: ledgerClient,
	}
}

func (s *TransferService) Execute(
	ctx context.Context,
	correlationID string,
	request model.TransferRequest,
) (model.APIResponse, *model.AppError) {
	normalized, validationErr := normalizeAndValidate(request)
	if validationErr != nil {
		return model.APIResponse{}, validationErr
	}

	log.Printf(
		"correlation_id=%s event=transfer_received reference=%s from_account=%s to_account=%s amount=%.2f",
		correlationID,
		normalized.Reference,
		normalized.FromAccount,
		normalized.ToAccount,
		normalized.Amount,
	)

	fraudDecision, fraudErr := s.fraudClient.Check(ctx, correlationID, normalized)
	if fraudErr != nil {
		log.Printf(
			"correlation_id=%s event=fraud_failed code=%s error=%s",
			correlationID,
			fraudErr.Code,
			fraudErr.Error(),
		)

		return model.APIResponse{}, fraudErr
	}

	log.Printf(
		"correlation_id=%s event=fraud_checked reference=%s decision=%s reason=%s",
		correlationID,
		normalized.Reference,
		fraudDecision.Decision,
		fraudDecision.Reason,
	)

	if !fraudDecision.Approved {
		return model.APIResponse{
			Status:        "rejected",
			Message:       "Transfer rejected by fraud rules",
			CorrelationID: correlationID,
			Data: model.TransferResult{
				Reference:     normalized.Reference,
				Amount:        normalized.Amount,
				FraudDecision: fraudDecision.Decision,
				Duplicate:     false,
			},
		}, nil
	}

	ledgerResult, ledgerErr := s.ledgerClient.Transfer(ctx, correlationID, normalized)
	if ledgerErr != nil {
		log.Printf(
			"correlation_id=%s event=ledger_failed code=%s error=%s",
			correlationID,
			ledgerErr.Code,
			ledgerErr.Error(),
		)

		return model.APIResponse{}, ledgerErr
	}

	log.Printf(
		"correlation_id=%s event=transfer_completed reference=%s transaction_id=%s duplicate=%t",
		correlationID,
		ledgerResult.Reference,
		ledgerResult.TransactionID,
		ledgerResult.Duplicate,
	)

	return model.APIResponse{
		Status:        "success",
		Message:       "Transfer processed successfully",
		CorrelationID: correlationID,
		Data: model.TransferResult{
			Reference:     ledgerResult.Reference,
			Amount:        ledgerResult.Amount,
			FraudDecision: fraudDecision.Decision,
			LedgerStatus:  ledgerResult.Status,
			TransactionID: ledgerResult.TransactionID,
			Duplicate:     ledgerResult.Duplicate,
			CreatedAt:     ledgerResult.CreatedAt,
		},
	}, nil
}

func normalizeAndValidate(request model.TransferRequest) (model.TransferRequest, *model.AppError) {
	normalized := model.TransferRequest{
		Reference:   strings.TrimSpace(request.Reference),
		FromAccount: strings.TrimSpace(request.FromAccount),
		ToAccount:   strings.TrimSpace(request.ToAccount),
		Amount:      request.Amount,
	}

	switch {
	case normalized.Reference == "":
		return model.TransferRequest{}, validationError("INVALID_REFERENCE", "Reference is required")
	case len(normalized.Reference) > 128:
		return model.TransferRequest{}, validationError("INVALID_REFERENCE", "Reference is too long")
	case normalized.FromAccount == "" || normalized.ToAccount == "":
		return model.TransferRequest{}, validationError("INVALID_ACCOUNT", "Both accounts are required")
	case normalized.FromAccount == normalized.ToAccount:
		return model.TransferRequest{}, validationError("SAME_ACCOUNT_TRANSFER", "Source and destination accounts must differ")
	case normalized.Amount <= 0:
		return model.TransferRequest{}, validationError("INVALID_AMOUNT", "Amount must be greater than zero")
	default:
		return normalized, nil
	}
}

func validationError(code, message string) *model.AppError {
	return &model.AppError{
		StatusCode: http.StatusBadRequest,
		Code:       code,
		Message:    message,
	}
}
