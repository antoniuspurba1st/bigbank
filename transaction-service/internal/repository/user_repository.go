package repository

import (
	"database/sql"
	"errors"
	"strings"

	"transaction-service/internal/model"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserEmailExists   = errors.New("email already registered")
	ErrUserPhoneExists   = errors.New("phone number already registered")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback if not committed

	// Insert user
	userQuery := `
		INSERT INTO users (id, email, password_hash, phone, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.Exec(
		userQuery,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Phone,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		errMsg := err.Error()
		// Check for PostgreSQL unique constraint violations
		if strings.Contains(errMsg, "duplicate key value violates unique constraint") {
			if strings.Contains(errMsg, "users_email_key") || strings.Contains(errMsg, "email") {
				return ErrUserEmailExists
			}
			if strings.Contains(errMsg, "users_phone_key") || strings.Contains(errMsg, "phone") {
				return ErrUserPhoneExists
			}
			// Fallback for any other unique constraint
			return ErrUserAlreadyExists
		}
		return err
	}

	// Create account for the user
	accountID := uuid.New().String()
	accountNumber := "ACC-" + strings.ToUpper(uuid.New().String())
	ownerName := user.Email
	accountQuery := `
		INSERT INTO accounts (id, user_id, account_number, owner_name, balance, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.Exec(
		accountQuery,
		accountID,
		user.ID,
		accountNumber,
		ownerName,
		0.00,
		user.CreatedAt,
	)

	if err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, phone, created_at, updated_at 
		FROM users 
		WHERE email = $1
	`

	row := r.db.QueryRow(query, email)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByID(id string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, phone, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindAccountByUserID(userID string) (string, float64, error) {
	query := `
		SELECT account_number, balance
		FROM accounts
		WHERE user_id = $1
		LIMIT 1
	`

	var accountNumber string
	var balance float64
	row := r.db.QueryRow(query, userID)
	if err := row.Scan(&accountNumber, &balance); err != nil {
		if err == sql.ErrNoRows {
			return "", 0, ErrUserNotFound
		}
		return "", 0, err
	}

	return accountNumber, balance, nil
}

func (r *UserRepository) FindAccountByEmail(email string) (string, float64, error) {
	query := `
		SELECT a.account_number, a.balance
		FROM accounts a
		INNER JOIN users u ON a.user_id = u.id
		WHERE u.email = $1
		LIMIT 1
	`

	var accountNumber string
	var balance float64
	row := r.db.QueryRow(query, email)
	if err := row.Scan(&accountNumber, &balance); err != nil {
		if err == sql.ErrNoRows {
			return "", 0, ErrUserNotFound
		}
		return "", 0, err
	}

	return accountNumber, balance, nil
}

func (r *UserRepository) Update(user *model.User) error {
	query := `
		UPDATE users 
		SET email = $1, password_hash = $2, phone = $3, updated_at = $4 
		WHERE id = $5
	`

	_, err := r.db.Exec(
		query,
		user.Email,
		user.PasswordHash,
		user.Phone,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return ErrUserAlreadyExists
		}
		return err
	}

	return nil
}
