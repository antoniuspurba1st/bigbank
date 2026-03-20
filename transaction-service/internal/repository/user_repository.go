package repository

import (
	"database/sql"
	"errors"

	"transaction-service/internal/model"

	_ "github.com/lib/pq"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, phone, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Phone,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// Basic check for postgres unique violation on email
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return ErrUserAlreadyExists
		}
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
