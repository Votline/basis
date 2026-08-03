// Package db users.go contains methods only for users-service
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

var ErrUserAlreadyExists = errors.New("user already exists")

// CreateUser insert new user to mysql db
func (m *DB) CreateUser(ctx context.Context, email, passwordHash string) (int64, error) {
	const op = "db.users.CreateUser"

	query := `INSERT INTO users (email, password_hash) VALUES (?, ?)`

	res, err := m.DB.ExecContext(ctx, query, email, passwordHash)
	if err != nil {
		if isDuplicateEntryError(err) {
			return 0, ErrUserAlreadyExists
		}
		return 0, fmt.Errorf("%s: exec: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: last insert id: %w", op, err)
	}

	return id, nil
}

// GetUserByEmail select user from mysql by email
func (m *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const op = "db.GetUserByEmail"

	query := `SELECT id, email, password_hash, created_at FROM users WHERE email = ?`

	user := &User{}
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, sql.ErrNoRows)
		}
		return nil, fmt.Errorf("%s: scan: %w", op, err)
	}

	return user, nil
}

func isDuplicateEntryError(err error) bool {
	return err != nil && (contains(err.Error(), "1062") || contains(err.Error(), "Duplicate entry"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && containsSubstr(s, substr)))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
