package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"gateway/internal/rdb"
	"gateway/internal/utils"

	"go.uber.org/zap"
)

type RateLimiter struct {
	// limit is the requests limit for one user
	limit int64

	// window is the time when user cannot send new requests
	window time.Duration

	client *rdb.RedisClient
	log    *zap.Logger
}

func NewRateLimiter(client *rdb.RedisClient, log *zap.Logger) Middleware {
	limit := int64(utils.GetEnvInt("REDIS_RATELIMIT_LIMIT", 100))
	window := time.Duration(utils.GetEnvInt("REDIS_RATELIMIT_WINDOW", 1)) * time.Minute

	return &RateLimiter{
		limit:  limit,
		window: window,
		client: client,
		log:    log,
	}
}

func (rl *RateLimiter) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientKey := r.RemoteAddr
		if userID, ok := ctx.Value("user_id").(int64); ok && userID > 0 {
			clientKey = fmt.Sprintf("user:%d", userID)
		}

		redisKey := fmt.Sprintf("rate_limit:%s", clientKey)

		count, err := rl.client.Incr(ctx, redisKey)
		if err != nil {
			rl.log.Error("rate limiter redis error", zap.Error(err))
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			if err := rl.client.Expire(ctx, redisKey, rl.window); err != nil {
				rl.log.Error("rate limiter expire redis error", zap.Error(err))
			}
		}

		if count > rl.limit {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too many requests"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
