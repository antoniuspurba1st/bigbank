package model

type FraudCheckResponse struct {
	Status        string         `json:"status"`
	Message       string         `json:"message"`
	CorrelationID string         `json:"correlation_id"`
	Data          *FraudDecision `json:"data"`
}

type FraudDecision struct {
	Decision  string `json:"decision"`
	Approved  bool   `json:"approved"`
	Reason    string `json:"reason"`
	CheckedAt string `json:"checked_at"`
}

type LedgerTransferRequest struct {
	Reference   string  `json:"reference"`
	FromAccount string  `json:"from_account"`
	ToAccount   string  `json:"to_account"`
	Amount      float64 `json:"amount"`
}

type LedgerTransferEnvelope struct {
	Status        string                `json:"status"`
	Message       string                `json:"message"`
	CorrelationID string                `json:"correlation_id"`
	Data          *LedgerTransferResult `json:"data"`
}

type LedgerTransferResult struct {
	TransactionID string  `json:"transaction_id"`
	Reference     string  `json:"reference"`
	FromAccount   string  `json:"from_account"`
	ToAccount     string  `json:"to_account"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	Duplicate     bool    `json:"duplicate"`
	CreatedAt     string  `json:"created_at"`
}
