// Package rdb handler.go implement methods
// for all redis operations
package rdb

import (
	"context"
	"fmt"
	"time"
)

// Incr increments the value of key and return new value
func (rc *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	const op = "rdb.handler.Incr"
	res, err := rc.rdb.Incr(ctx, key).Result()
	if err != nil {
		return -1, fmt.Errorf("%s: :%w", op, err)
	}
	return res, nil
}

// Expire set the expiration time
func (rc *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	const op = "rdb.handler.Expire"
	if err := rc.rdb.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("%s: :%w", op, err)
	}
	return nil
}
