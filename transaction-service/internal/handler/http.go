package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"transaction-service/internal/model"
	"transaction-service/internal/service"
)

type HTTPHandler struct {
	transferService *service.TransferService
}

func NewHTTPHandler(transferService *service.TransferService) *HTTPHandler {
	return &HTTPHandler{transferService: transferService}
}

func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/transfer", h.handleTransfer)

	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFromRequest(r)
	w.Header().Set("X-Correlation-Id", correlationID)

	writeJSON(w, http.StatusOK, model.APIResponse{
		Status:        "success",
		Message:       "Transaction service is healthy",
		CorrelationID: correlationID,
		Data: map[string]string{
			"service": "transaction",
			"status":  "UP",
		},
	})
}

func (h *HTTPHandler) handleTransfer(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFromRequest(r)
	w.Header().Set("X-Correlation-Id", correlationID)

	if r.Method != http.MethodPost {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Only POST /transfer is supported",
		})
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	request := model.TransferRequest{}
	if err := decoder.Decode(&request); err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "MALFORMED_REQUEST",
			Message:    "Request body is malformed",
			Err:        err,
		})
		return
	}

	response, appErr := h.transferService.Execute(r.Context(), correlationID, request)
	if appErr != nil {
		writeError(w, correlationID, appErr)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func writeError(w http.ResponseWriter, correlationID string, appErr *model.AppError) {
	log.Printf(
		"correlation_id=%s event=request_failed code=%s status=%d error=%s",
		correlationID,
		appErr.Code,
		appErr.StatusCode,
		appErr.Error(),
	)

	writeJSON(w, appErr.StatusCode, model.APIError{
		Status:        "error",
		Code:          appErr.Code,
		Message:       appErr.Message,
		CorrelationID: correlationID,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("event=response_encode_failed error=%s", err)
	}
}

func correlationIDFromRequest(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("X-Correlation-Id"))
	if header != "" {
		return header
	}

	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "corr-fallback"
	}

	return hex.EncodeToString(buffer)
}
