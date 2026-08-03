// Package rdb rdb.go provides connection to redis
package rdb

import (
	"context"
	"fmt"
	"os"
	"time"

	"gateway/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient struct {
	rdb *redis.Client
	log *zap.Logger
}

func NewRC(log *zap.Logger) (*RedisClient, error) {
	const op = "rdb.rdb.New"

	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	pingTimeout := time.Duration(utils.GetEnvInt("REDIS_PING_TIMEOUT", 5)) * time.Second

	addr := fmt.Sprintf("%s:%s", host, port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("%s: ping redis at %s failed: %w", op, addr, err)
	}

	log.Info("connected to redis successfully", zap.String("addr", addr))

	return &RedisClient{
		rdb: rdb,
		log: log,
	}, nil
}

func (rc *RedisClient) Close() error {
	const op = "rdb.rdb.Close"

	if err := rc.rdb.Close(); err != nil {
		return fmt.Errorf("%s: redis close: %w", op, err)
	}
	return nil
}
