package model

type TransferRequest struct {
	Reference   string  `json:"reference"`
	FromAccount string  `json:"from_account"`
	ToAccount   string  `json:"to_account"`
	Amount      float64 `json:"amount"`
}

type TransferResult struct {
	Reference     string  `json:"reference"`
	Amount        float64 `json:"amount"`
	FraudDecision string  `json:"fraud_decision"`
	LedgerStatus  string  `json:"ledger_status,omitempty"`
	TransactionID string  `json:"transaction_id,omitempty"`
	Duplicate     bool    `json:"duplicate"`
	CreatedAt     string  `json:"created_at,omitempty"`
}
