package service

import (
	"context"
	"log"

	"transaction-service/internal/model"
)

type ledgerReader interface {
	ListTransactions(ctx context.Context, correlationID string, page int, limit int) (model.TransactionHistoryPage, *model.AppError)
}

type TransactionQueryService struct {
	ledgerClient ledgerReader
}

func NewTransactionQueryService(ledgerClient ledgerReader) *TransactionQueryService {
	return &TransactionQueryService{ledgerClient: ledgerClient}
}

func (s *TransactionQueryService) ListTransactions(
	ctx context.Context,
	correlationID string,
	page int,
	limit int,
) (model.APIResponse, *model.AppError) {
	sanitizedPage := max(page, 0)
	sanitizedLimit := limit
	if sanitizedLimit <= 0 {
		sanitizedLimit = defaultTransactionLimit
	}
	if sanitizedLimit > maxTransactionLimit {
		sanitizedLimit = maxTransactionLimit
	}

	log.Printf(
		"correlation_id=%s event=transactions_requested page=%d limit=%d",
		correlationID,
		sanitizedPage,
		sanitizedLimit,
	)

	pageResult, appErr := s.ledgerClient.ListTransactions(ctx, correlationID, sanitizedPage, sanitizedLimit)
	if appErr != nil {
		log.Printf(
			"correlation_id=%s event=transactions_failed code=%s error=%s",
			correlationID,
			appErr.Code,
			appErr.Error(),
		)
		return model.APIResponse{}, appErr
	}

	return model.APIResponse{
		Status:        "success",
		Message:       "Transactions fetched successfully",
		CorrelationID: correlationID,
		Data:          pageResult,
	}, nil
}

const (
	defaultTransactionLimit = 10
	maxTransactionLimit     = 100
)
