package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"transaction-service/internal/model"
	"transaction-service/internal/repository"
)

var (
	hasUppercase = regexp.MustCompile(`[A-Z]`)
	hasLowercase = regexp.MustCompile(`[a-z]`)
	hasNumber    = regexp.MustCompile(`[0-9]`)
)

type Session struct {
	UserID     string
	Email      string
	CreatedAt  time.Time
	LastActive time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

const (
	idleTimeout        = 15 * time.Minute
	maxSessionDuration = 24 * time.Hour
)

func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	if !hasUppercase.MatchString(password) {
		return false
	}
	if !hasLowercase.MatchString(password) {
		return false
	}
	if !hasNumber.MatchString(password) {
		return false
	}
	return true
}

func (sm *SessionManager) CreateSession(userID, email string) string {
	sessionID := uuid.New().String()
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[sessionID] = &Session{
		UserID:     userID,
		Email:      email,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	return sessionID
}

func (sm *SessionManager) ValidateSession(sessionID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, false
	}

	now := time.Now()
	if now.Sub(session.LastActive) > idleTimeout {
		return nil, false
	}
	if now.Sub(session.CreatedAt) > maxSessionDuration {
		return nil, false
	}

	return session, true
}

func (sm *SessionManager) RefreshSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, exists := sm.sessions[sessionID]; exists {
		session.LastActive = time.Now()
	}
}

func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

var sessionManager = &SessionManager{
	sessions: make(map[string]*Session),
}

type UserRepository interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	FindByID(id string) (*model.User, error)
	FindAccountByUserID(userID string) (string, float64, error)
	FindAccountByEmail(email string) (string, float64, error)
	Update(user *model.User) error
}

type AuthHandler struct {
	userRepo       UserRepository
	sessionManager *SessionManager
}

func NewAuthHandler(userRepo UserRepository) *AuthHandler {
	return &AuthHandler{
		userRepo:       userRepo,
		sessionManager: sessionManager,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdatePhoneRequest struct {
	Phone string `json:"phone"`
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Method not allowed",
		})
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "MALFORMED_REQUEST",
			Message:    "Invalid request body",
		})
		return
	}

	if req.Email == "" {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_INPUT",
			Message:    "Email is required",
		})
		return
	}

	if !validatePassword(req.Password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password does not meet requirements"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Error processing password",
		})
		return
	}

	now := time.Now()
	user := &model.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Phone:        req.Phone,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.userRepo.Create(user); err != nil {
		if err == repository.ErrUserAlreadyExists {
			writeError(w, correlationIDFromRequest(r), &model.AppError{
				StatusCode: http.StatusConflict,
				Code:       "USER_EXISTS",
				Message:    "User with this email already exists",
			})
			return
		}
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to create user",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"phone": user.Phone,
		},
	})
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Method not allowed",
		})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "MALFORMED_REQUEST",
			Message:    "Invalid request body",
		})
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			writeError(w, correlationIDFromRequest(r), &model.AppError{
				StatusCode: http.StatusUnauthorized,
				Code:       "UNAUTHORIZED",
				Message:    "Invalid credentials",
			})
			return
		}
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Database error",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "Invalid credentials",
		})
		return
	}

	sessionID := sessionManager.CreateSession(user.ID, user.Email)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "success",
		"session_id": sessionID,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"phone": user.Phone,
		},
	})
}

func (h *AuthHandler) HandleUpdatePhone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Method not allowed",
		})
		return
	}

	// Basic prototype: Using query param or minimal header auth for identity
	email := r.Header.Get("X-User-Email")
	if email == "" {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "Unauthorized",
		})
		return
	}

	var req UpdatePhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "MALFORMED_REQUEST",
			Message:    "Invalid request body",
		})
		return
	}

	user, err := h.userRepo.FindByEmail(email)
	if err != nil {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusNotFound,
			Code:       "USER_NOT_FOUND",
			Message:    "User not found",
		})
		return
	}

	user.Phone = req.Phone
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(user); err != nil {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to update profile",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{

		"status": "success",

		"message": "Profile updated successfully",
	})

}

func (h *AuthHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFromRequest(r)

	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	email := strings.TrimSpace(r.Header.Get("X-User-Email"))

	if userID == "" && email == "" {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "Authentication required",
		})
		return
	}

	var user *model.User
	var err error
	if userID != "" {
		user, err = h.userRepo.FindByID(userID)
	} else {
		user, err = h.userRepo.FindByEmail(email)
	}

	if err != nil {
		if err == repository.ErrUserNotFound {
			writeError(w, correlationID, &model.AppError{
				StatusCode: http.StatusNotFound,
				Code:       "USER_NOT_FOUND",
				Message:    "User not found",
			})
			return
		}

		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to fetch user profile",
			Err:        err,
		})
		return
	}

	accountNumber, balance, err := h.userRepo.FindAccountByUserID(user.ID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			writeError(w, correlationID, &model.AppError{
				StatusCode: http.StatusNotFound,
				Code:       "ACCOUNT_NOT_FOUND",
				Message:    "Account not found",
			})
			return
		}

		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to fetch account for user",
			Err:        err,
		})
		return
	}

	log.Printf("correlation_id=%s event=get_me user_id=%s email=%s account=%s balance=%.2f", correlationID, user.ID, user.Email, accountNumber, balance)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"phone": user.Phone,
		},
		"account": map[string]interface{}{
			"account_number": accountNumber,
			"balance":        balance,
		},
	})
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFromRequest(r)

	if r.Method != http.MethodPut {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Method not allowed",
		})
		return
	}

	sessionID := r.Header.Get("X-Session-ID")
	session, valid := sessionManager.ValidateSession(sessionID)
	if !valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Session expired", "redirect": "/login"})
		return
	}
	sessionManager.RefreshSession(sessionID)

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "MALFORMED_REQUEST",
			Message:    "Invalid request body",
		})
		return
	}

	if !validatePassword(req.NewPassword) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password does not meet requirements"})
		return
	}

	user, err := h.userRepo.FindByID(session.UserID)
	if err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusNotFound,
			Code:       "USER_NOT_FOUND",
			Message:    "User not found",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "Current password is incorrect",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Error processing password",
		})
		return
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(user); err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to update password",
		})
		return
	}

	sessionManager.DeleteSession(sessionID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Password updated successfully",
	})
}

type UpdateEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) HandleUpdateEmail(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFromRequest(r)

	if r.Method != http.MethodPut {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Method not allowed",
		})
		return
	}

	sessionID := r.Header.Get("X-Session-ID")
	session, valid := sessionManager.ValidateSession(sessionID)
	if !valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Session expired", "redirect": "/login"})
		return
	}
	sessionManager.RefreshSession(sessionID)

	var req UpdateEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "MALFORMED_REQUEST",
			Message:    "Invalid request body",
		})
		return
	}

	if req.Email == "" {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_INPUT",
			Message:    "Email is required",
		})
		return
	}

	user, err := h.userRepo.FindByID(session.UserID)
	if err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusNotFound,
			Code:       "USER_NOT_FOUND",
			Message:    "User not found",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "Password is incorrect",
		})
		return
	}

	existingUser, _ := h.userRepo.FindByEmail(req.Email)
	if existingUser != nil && existingUser.ID != user.ID {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusConflict,
			Code:       "EMAIL_EXISTS",
			Message:    "Email already in use",
		})
		return
	}

	user.Email = req.Email
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(user); err != nil {
		writeError(w, correlationID, &model.AppError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INTERNAL_ERROR",
			Message:    "Failed to update email",
		})
		return
	}

	sessionManager.DeleteSession(sessionID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Email updated successfully",
	})
}
