package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

var (
	ErrIdempotencyAlreadyExists = errors.New("idempotency key already exists")
)

type IdempotencyRepository struct {
	db *sql.DB
}

func NewIdempotencyRepository(db *sql.DB) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

func (r *IdempotencyRepository) Start(key string) error {
	if key == "" {
		return errors.New("idempotency key is required")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO idempotency_keys (idempotency_key, status, updated_at) VALUES ($1, 'in_progress', $2)`,
		key,
		time.Now().UTC(),
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			return ErrIdempotencyAlreadyExists
		}
		return err
	}

	return tx.Commit()
}

func (r *IdempotencyRepository) Complete(key string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE idempotency_keys SET status='completed', updated_at=$2 WHERE idempotency_key=$1`,
		key,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *IdempotencyRepository) Fail(key string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE idempotency_keys SET status='failed', updated_at=$2 WHERE idempotency_key=$1`,
		key,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
