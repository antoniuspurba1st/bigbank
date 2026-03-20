package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"transaction-service/internal/model"
	"transaction-service/internal/service"
)

type HTTPHandler struct {
	transferService      *service.TransferService
	transactionQueryServ *service.TransactionQueryService
	authHandler          *AuthHandler
	observability        *observabilityMiddleware
}

func NewHTTPHandler(
	transferService *service.TransferService,
	transactionQueryService *service.TransactionQueryService,
	authHandler *AuthHandler,
) *HTTPHandler {
	return &HTTPHandler{
		transferService:      transferService,
		transactionQueryServ: transactionQueryService,
		authHandler:          authHandler,
		observability:        newObservabilityMiddleware(),
	}
}

func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.handleHealth)
	
	// Transaction routes
	mux.HandleFunc("/transfer", h.handleTransfer)
	mux.HandleFunc("/transactions", h.handleTransactions)

	// Auth routes
	if h.authHandler != nil {
		mux.HandleFunc("/auth/register", h.authHandler.HandleRegister)
		mux.HandleFunc("/auth/login", h.authHandler.HandleLogin)
		mux.HandleFunc("/auth/phone", h.authHandler.HandleUpdatePhone)
		// /auth/email and /auth/password would go here following the same pattern
	}

	return h.observability.wrap(corsMiddleware(mux))
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

func (h *HTTPHandler) handleTransactions(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFromRequest(r)
	w.Header().Set("X-Correlation-Id", correlationID)

	if r.Method != http.MethodGet {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Only GET /transactions is supported",
		})
		return
	}

	page, err := parseIntQuery(r, "page", 0)
	if err != nil {
		writeError(w, correlationID, err)
		return
	}

	limit, err := parseIntQuery(r, "limit", serviceDefaultTransactionLimit())
	if err != nil {
		writeError(w, correlationID, err)
		return
	}

	response, appErr := h.transactionQueryServ.ListTransactions(r.Context(), correlationID, page, limit)
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

func parseIntQuery(r *http.Request, key string, fallback int) (int, *model.AppError) {
	rawValue := strings.TrimSpace(r.URL.Query().Get(key))
	if rawValue == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_QUERY",
			Message:    key + " must be a valid integer",
			Err:        err,
		}
	}

	if parsed < 0 {
		return 0, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_QUERY",
			Message:    key + " must not be negative",
		}
	}

	return parsed, nil
}

func serviceDefaultTransactionLimit() int {
	return 10
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Correlation-Id, X-User-Email")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
