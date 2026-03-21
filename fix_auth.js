const fs = require('fs');
const path = 'C:\\Users\\ronal\\Desktop\\com.bigbank\\transaction-service\\internal\\handler\\auth.go';
const content = fs.readFileSync(path, 'utf8');

// The full corrected auth.go content
const newContent = `package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"transaction-service/internal/model"
	"transaction-service/internal/repository"
)

var (
	passwordMinLength    = 8
	passwordUpperRegex   = regexp.MustCompile(` + '`' + `[A-Z]` + '`' + `)
	passwordNumberRegex  = regexp.MustCompile(` + '`' + `[0-9]` + '`' + `)
	passwordSpecialRegex = regexp.MustCompile(` + '`' + `[!@#$%^&*()_+\\-=\\[\\]{};\':"\\\\|,.<>\\/?]` + '`' + `)
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
	Email    string ` + '`' + `json:"email"` + '`' + `
	Password string ` + '`' + `json:"password"` + '`' + `
	Phone    string ` + '`' + `json:"phone"` + '`' + `
}

type LoginRequest struct {
	Email    string ` + '`' + `json:"email"` + '`' + `
	Password string ` + '`' + `json:"password"` + '`' + `
}

type UpdatePhoneRequest struct {
	Phone string ` + '`' + `json:"phone"` + '`' + `
}

type UpdatePasswordRequest struct {
	CurrentPassword string ` + '`' + `json:"current_password"` + '`' + `
	NewPassword     string ` + '`' + `json:"new_password"` + '`' + `
}

type UpdateEmailRequest struct {
	NewEmail string ` + '`' + `json:"new_email"` + '`' + `
}
`;
console.log('preview length:', newContent.length);
