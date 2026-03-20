package model

type TransactionHistoryItem struct {
	TransactionID string  `json:"transaction_id"`
	Reference     string  `json:"reference"`
	FromAccount   string  `json:"from_account"`
	ToAccount     string  `json:"to_account"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
}

type TransactionHistoryPage struct {
	Items       []TransactionHistoryItem `json:"items"`
	Page        int                      `json:"page"`
	Limit       int                      `json:"limit"`
	TotalItems  int64                    `json:"total_items"`
	TotalPages  int                      `json:"total_pages"`
	HasNext     bool                     `json:"has_next"`
	HasPrevious bool                     `json:"has_previous"`
}

type LedgerTransactionsEnvelope struct {
	Status        string                  `json:"status"`
	Message       string                  `json:"message"`
	CorrelationID string                  `json:"correlation_id"`
	Data          *TransactionHistoryPage `json:"data"`
}
