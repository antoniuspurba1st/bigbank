package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"transaction-service/internal/model"
)

type jsonHTTPClient struct {
	baseURL     string
	client      *http.Client
	maxAttempts int
}

func newJSONHTTPClient(baseURL string, timeout time.Duration, maxAttempts int) *jsonHTTPClient {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	return &jsonHTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: timeout,
		},
		maxAttempts: maxAttempts,
	}
}

func (c *jsonHTTPClient) postJSON(
	ctx context.Context,
	path string,
	correlationID string,
	payload interface{},
	output interface{},
) *model.AppError {
	body, err := json.Marshal(payload)
	if err != nil {
		return &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "REQUEST_ENCODING_FAILED",
			Message:    "Failed to encode downstream request",
			Err:        err,
		}
	}

	url := c.baseURL + path

	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		req, requestErr := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			url,
			bytes.NewReader(body),
		)
		if requestErr != nil {
			return &model.AppError{
				StatusCode: http.StatusInternalServerError,
				Code:       "REQUEST_BUILD_FAILED",
				Message:    "Failed to build downstream request",
				Err:        requestErr,
			}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Correlation-Id", correlationID)

		resp, callErr := c.client.Do(req)
		if callErr != nil {
			if attempt < c.maxAttempts {
				log.Printf("retry attempt=%d/%d due=network error: %v", attempt, c.maxAttempts, callErr)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			var netErr net.Error
			if errors.As(callErr, &netErr) && netErr.Timeout() {
				return &model.AppError{
					StatusCode: http.StatusGatewayTimeout,
					Code:       "SERVICE_TIMEOUT",
					Message:    "Service timeout",
					Err:        callErr,
				}
			}

			return &model.AppError{
				StatusCode: http.StatusServiceUnavailable,
				Code:       "DOWNSTREAM_UNAVAILABLE",
				Message:    "Downstream service is unavailable",
				Err:        callErr,
			}
		}

		readErr := decodeResponse(resp, output)
		if readErr == nil {
			return nil
		}

		var appErr *model.AppError
		if errors.As(readErr, &appErr) {
			if appErr.StatusCode >= http.StatusInternalServerError && attempt < c.maxAttempts {
				log.Printf("retry attempt=%d/%d due=server error %d", attempt, c.maxAttempts, appErr.StatusCode)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			return appErr
		}

		return &model.AppError{
			StatusCode: http.StatusBadGateway,
			Code:       "DOWNSTREAM_RESPONSE_INVALID",
			Message:    "Failed to decode downstream response",
			Err:        readErr,
		}
	}

	return &model.AppError{
		StatusCode: http.StatusBadGateway,
		Code:       "DOWNSTREAM_RESPONSE_INVALID",
		Message:    "Downstream response failed after retries",
	}
}

func (c *jsonHTTPClient) getJSON(
	ctx context.Context,
	path string,
	correlationID string,
	query url.Values,
	output interface{},
) *model.AppError {
	urlValue := c.baseURL + path
	if len(query) > 0 {
		urlValue += "?" + query.Encode()
	}

	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		req, requestErr := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			urlValue,
			nil,
		)
		if requestErr != nil {
			return &model.AppError{
				StatusCode: http.StatusInternalServerError,
				Code:       "REQUEST_BUILD_FAILED",
				Message:    "Failed to build downstream request",
				Err:        requestErr,
			}
		}

		req.Header.Set("X-Correlation-Id", correlationID)

		resp, callErr := c.client.Do(req)
		if callErr != nil {
			if attempt < c.maxAttempts {
				log.Printf("retry attempt=%d/%d due=network error: %v", attempt, c.maxAttempts, callErr)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			var netErr net.Error
			if errors.As(callErr, &netErr) && netErr.Timeout() {
				return &model.AppError{
					StatusCode: http.StatusGatewayTimeout,
					Code:       "SERVICE_TIMEOUT",
					Message:    "Service timeout",
					Err:        callErr,
				}
			}

			return &model.AppError{
				StatusCode: http.StatusServiceUnavailable,
				Code:       "DOWNSTREAM_UNAVAILABLE",
				Message:    "Downstream service is unavailable",
				Err:        callErr,
			}
		}

		readErr := decodeResponse(resp, output)
		if readErr == nil {
			return nil
		}

		var appErr *model.AppError
		if errors.As(readErr, &appErr) {
			if appErr.StatusCode >= http.StatusInternalServerError && attempt < c.maxAttempts {
				log.Printf("retry attempt=%d/%d due=server error %d", attempt, c.maxAttempts, appErr.StatusCode)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			return appErr
		}

		return &model.AppError{
			StatusCode: http.StatusBadGateway,
			Code:       "DOWNSTREAM_RESPONSE_INVALID",
			Message:    "Failed to decode downstream response",
			Err:        readErr,
		}
	}

	return &model.AppError{
		StatusCode: http.StatusBadGateway,
		Code:       "DOWNSTREAM_RESPONSE_INVALID",
		Message:    "Downstream response failed after retries",
	}
}

func decodeResponse(resp *http.Response, output interface{}) error {
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &model.AppError{
			StatusCode: http.StatusBadGateway,
			Code:       "DOWNSTREAM_READ_FAILED",
			Message:    "Failed to read downstream response",
			Err:        err,
		}
	}

	if resp.StatusCode >= http.StatusBadRequest {
		apiErr := model.APIError{}
		if json.Unmarshal(responseBody, &apiErr) == nil && apiErr.Error != "" {
			return &model.AppError{
				StatusCode: resp.StatusCode,
				Code:       "DOWNSTREAM_REQUEST_FAILED",
				Message:    apiErr.Error,
			}
		}

		return &model.AppError{
			StatusCode: resp.StatusCode,
			Code:       "DOWNSTREAM_REQUEST_FAILED",
			Message:    string(bytes.TrimSpace(responseBody)),
		}
	}

	if err := json.Unmarshal(responseBody, output); err != nil {
		return err
	}

	return nil
}

func backoff(attempt int) time.Duration {
	return time.Duration(attempt+1) * 150 * time.Millisecond
}
