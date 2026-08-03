// Package db repo.go contains connection to mysql db
// add creating table
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gateway/internal/utils"

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

	moc := utils.GetEnvInt("DB_MAX_OPEN_CONNS", 25)
	mic := utils.GetEnvInt("DB_MAX_IDLE_CONNS", 25)
	cml := time.Duration(utils.GetEnvInt("DB_CONN_MAX_LIFETIME", 5)) * time.Minute
	mit := time.Duration(utils.GetEnvInt("DB_CONN_MAX_IDLE_TIME", 5)) * time.Minute
	pingTimeout := time.Duration(utils.GetEnvInt("DB_PING_TIMEOUT", 5)) * time.Second

	db.SetMaxOpenConns(moc)
	db.SetMaxIdleConns(mic)
	db.SetConnMaxLifetime(cml)
	db.SetConnMaxIdleTime(mit)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
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
