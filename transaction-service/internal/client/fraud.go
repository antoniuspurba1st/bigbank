package client

import (
	"context"
	"net/http"
	"strings"
	"time"

	"transaction-service/internal/model"
)

type FraudClient struct {
	httpClient *jsonHTTPClient
}

func NewFraudClient(baseURL string, timeout time.Duration, retries int) *FraudClient {
	return &FraudClient{
		httpClient: newJSONHTTPClient(baseURL, timeout, retries),
	}
}

func (c *FraudClient) Check(
	ctx context.Context,
	correlationID string,
	request model.TransferRequest,
) (model.FraudDecision, *model.AppError) {
	payload := map[string]interface{}{
		"reference":    strings.TrimSpace(request.Reference),
		"from_account": strings.TrimSpace(request.FromAccount),
		"to_account":   strings.TrimSpace(request.ToAccount),
		"amount":       request.Amount,
	}

	response := model.FraudCheckResponse{}
	if err := c.httpClient.postJSON(ctx, "/fraud/check", correlationID, payload, &response); err != nil {
		err.Message = "Fraud service request failed"
		if err.StatusCode == http.StatusServiceUnavailable {
			err.Code = "FRAUD_SERVICE_UNAVAILABLE"
		}

		return model.FraudDecision{}, err
	}

	if response.Data == nil {
		return model.FraudDecision{}, &model.AppError{
			StatusCode: http.StatusBadGateway,
			Code:       "FRAUD_RESPONSE_INVALID",
			Message:    "Fraud service returned an empty decision",
		}
	}

	return *response.Data, nil
}
