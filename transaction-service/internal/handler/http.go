package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"transaction-service/internal/model"
	"transaction-service/internal/repository"
	"transaction-service/internal/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type idempotencyRepository interface {
	Start(key string) error
	Complete(key string) error
	Fail(key string) error
}

type HTTPHandler struct {
	transferService      *service.TransferService
	transactionQueryServ *service.TransactionQueryService
	authHandler          *AuthHandler
	idempotencyRepo      idempotencyRepository
	observability        *observabilityMiddleware
	DB                   *sql.DB
	ledgerURL            string
	fraudURL             string
}

func NewHTTPHandler(
	transferService *service.TransferService,
	transactionQueryService *service.TransactionQueryService,
	idempotencyRepo idempotencyRepository,
	db *sql.DB,
	authHandler ...*AuthHandler,
) *HTTPHandler {
	var auth *AuthHandler
	if len(authHandler) > 0 {
		auth = authHandler[0]
	}

	return &HTTPHandler{
		transferService:      transferService,
		transactionQueryServ: transactionQueryService,
		authHandler:          auth,
		idempotencyRepo:      idempotencyRepo,
		observability:        newObservabilityMiddleware(),
		DB:                   db,
		ledgerURL:            "http://localhost:8080",
		fraudURL:             "http://localhost:8082",
	}
}

// NewHTTPHandlerWithURLs creates HTTPHandler with custom service URLs
func NewHTTPHandlerWithURLs(
	transferService *service.TransferService,
	transactionQueryService *service.TransactionQueryService,
	idempotencyRepo idempotencyRepository,
	db *sql.DB,
	ledgerURL, fraudURL string,
	authHandler ...*AuthHandler,
) *HTTPHandler {
	var auth *AuthHandler
	if len(authHandler) > 0 {
		auth = authHandler[0]
	}

	return &HTTPHandler{
		transferService:      transferService,
		transactionQueryServ: transactionQueryService,
		authHandler:          auth,
		idempotencyRepo:      idempotencyRepo,
		observability:        newObservabilityMiddleware(),
		DB:                   db,
		ledgerURL:            ledgerURL,
		fraudURL:             fraudURL,
	}
}

func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/ready", h.handleReady)
	mux.Handle("/metrics", promhttp.Handler())

	// Transaction routes (protected)
	mux.HandleFunc("/transfer", authMiddleware(h.handleTransfer))
	mux.HandleFunc("/topup", authMiddleware(h.handleTopup))
	mux.HandleFunc("/transactions", h.handleTransactions)

	// Auth routes
	if h.authHandler != nil {
		mux.HandleFunc("/auth/register", h.authHandler.HandleRegister)
		mux.HandleFunc("/auth/login", h.authHandler.HandleLogin)
		mux.HandleFunc("/auth/me", h.authHandler.HandleGetMe)
		mux.HandleFunc("/auth/phone", h.authHandler.HandleUpdatePhone)
		mux.HandleFunc("/auth/password", h.authHandler.HandleUpdatePassword)
		mux.HandleFunc("/auth/email", h.authHandler.HandleUpdateEmail)
	}

	return h.observability.wrap(securityHeadersMiddleware(corsMiddleware(mux)))
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFromRequest(r)
	w.Header().Set("X-Correlation-Id", correlationID)
	w.Header().Set("Content-Type", "application/json")

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "UP",
	})
}

func (h *HTTPHandler) handleReady(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFromRequest(r)
	w.Header().Set("X-Correlation-Id", correlationID)
	w.Header().Set("Content-Type", "application/json")

	// Check database connectivity
	if h.DB == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "NOT_READY",
			"reason": "database not initialized",
		})
		return
	}

	if err := h.DB.Ping(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "NOT_READY",
			"reason": "database connection failed",
		})
		return
	}

	// Check ledger service availability
	ledgerHealthURL := h.ledgerURL + "/health"
	resp, err := http.Get(ledgerHealthURL)
	if err != nil || resp.StatusCode >= 400 {
		if err != nil {
			log.Printf("ledger_health_check_failed error=%v", err)
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "NOT_READY",
			"reason": "ledger service unavailable",
		})
		return
	}
	resp.Body.Close()

	// Check fraud service availability
	fraudHealthURL := h.fraudURL + "/health"
	resp, err = http.Get(fraudHealthURL)
	if err != nil || resp.StatusCode >= 400 {
		if err != nil {
			log.Printf("fraud_health_check_failed error=%v", err)
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "NOT_READY",
			"reason": "fraud service unavailable",
		})
		return
	}
	resp.Body.Close()

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "READY",
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

	email := r.Header.Get("X-User-Email")
	if email == "" {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "Authentication required",
		})
		return
	}

	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "MISSING_IDEMPOTENCY_KEY",
			Message:    "Idempotency-Key header is required",
		})
		return
	}

	if err := h.idempotencyRepo.Start(idempotencyKey); err != nil {
		if err == repository.ErrIdempotencyAlreadyExists {
			writeError(w, correlationID, &model.AppError{
				StatusCode: http.StatusConflict,
				Code:       "DUPLICATE_REQUEST",
				Message:    "Duplicate request",
			})
			return
		}

		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to reserve idempotency key",
			Err:        err,
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
		_ = h.idempotencyRepo.Fail(idempotencyKey)
		writeError(w, correlationID, appErr)
		return
	}

	if err := h.idempotencyRepo.Complete(idempotencyKey); err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to complete idempotency key",
			Err:        err,
		})
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

type TopupRequest struct {
	Amount float64 `json:"amount"`
}

func (h *HTTPHandler) handleTopup(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFromRequest(r)
	w.Header().Set("X-Correlation-Id", correlationID)

	if r.Method != http.MethodPost {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Only POST /topup is supported",
		})
		return
	}

	userID := r.Header.Get("X-User-ID")
	email := r.Header.Get("X-User-Email")
	if userID == "" && email == "" {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "Authentication required",
		})
		return
	}

	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "MISSING_IDEMPOTENCY_KEY",
			Message:    "Idempotency-Key header is required",
		})
		return
	}

	if err := h.idempotencyRepo.Start(idempotencyKey); err != nil {
		if err == repository.ErrIdempotencyAlreadyExists {
			writeError(w, correlationID, &model.AppError{
				StatusCode: http.StatusConflict,
				Code:       "DUPLICATE_REQUEST",
				Message:    "Duplicate request",
			})
			return
		}

		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to reserve idempotency key",
			Err:        err,
		})
		return
	}

	var req TopupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "MALFORMED_REQUEST",
			Message:    "Request body is malformed",
			Err:        err,
		})
		return
	}

	if req.Amount <= 0 {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_AMOUNT",
			Message:    "Amount must be greater than zero",
		})
		return
	}

	if req.Amount > 10000000 {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_AMOUNT",
			Message:    "Amount exceeds top-up limit",
		})
		return
	}

	var accountNumber string
	var err error
	if userID != "" {
		accountNumber, _, err = h.authHandler.userRepo.FindAccountByUserID(userID)
	} else {
		accountNumber, _, err = h.authHandler.userRepo.FindAccountByEmail(email)
	}

	if err != nil {
		if err == repository.ErrUserNotFound {
			_ = h.idempotencyRepo.Fail(idempotencyKey)
			writeError(w, correlationID, &model.AppError{
				StatusCode: http.StatusNotFound,
				Code:       "ACCOUNT_NOT_FOUND",
				Message:    "Account not found",
			})
			return
		}

		_ = h.idempotencyRepo.Fail(idempotencyKey)
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to resolve user account",
			Err:        err,
		})
		return
	}

	response, appErr := h.transferService.Topup(r.Context(), correlationID, accountNumber, req.Amount)
	if appErr != nil {
		_ = h.idempotencyRepo.Fail(idempotencyKey)
		writeError(w, correlationID, appErr)
		return
	}

	if err := h.idempotencyRepo.Complete(idempotencyKey); err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to complete idempotency key",
			Err:        err,
		})
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
		Error: appErr.Message,
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

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			sessionID = r.Header.Get("X-User-Session-ID")
		}

		_, valid := sessionManager.ValidateSession(sessionID)
		if !valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Session expired", "redirect": "/login"})
			return
		}
		sessionManager.RefreshSession(sessionID)
		next(w, r)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Correlation-Id, X-User-Email, X-User-ID, X-Session-ID, X-User-Session-ID, Idempotency-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
