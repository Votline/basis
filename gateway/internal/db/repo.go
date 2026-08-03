// Package db repo.go contains connection to mysql db
// add creating table
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type MySQL struct {
	DB  *sql.DB
	log *zap.Logger
}

func NewDB(dsn string, log *zap.Logger) (*MySQL, error) {
	const op = "db.repo.NewDB"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: open db: %w", op, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%s: ping db: %w", op, err)
	}

	return &MySQL{
		DB:  db,
		log: log,
	}, nil
}

func (m *MySQL) Close() error {
	const op = "db.repo.Close"

	if err := m.DB.Close(); err != nil {
		return fmt.Errorf("%s: close db: %w", op, err)
	}

	return nil
}
