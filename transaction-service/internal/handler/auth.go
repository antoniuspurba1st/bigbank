package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"transaction-service/internal/model"
	"transaction-service/internal/repository"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
}

func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
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

	if req.Email == "" || len(req.Password) < 6 {
		writeError(w, correlationIDFromRequest(r), &model.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_INPUT",
			Message:    "Email required and password must be at least 6 characters",
		})
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

	// In a complete implementation we would issue a real JWT here.
	// Since UI is using localStorage session simulation for frontend gating per instructions
	// we just return success and user payload. Focus is making the product demoable quickly.

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
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
