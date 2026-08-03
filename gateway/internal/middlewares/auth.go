// Package middlewares auth.go checks the authorization header
// and sets the user data in context
package middlewares

import (
	"context"
	"net/http"
	"strings"

	"gateway/internal/utils"

	"go.uber.org/zap"
)

type Auth struct {
	jwtSecret []byte
	log       *zap.Logger
}

func NewAuth(jwtSecret string, log *zap.Logger) Middleware {
	return &Auth{
		jwtSecret: []byte(jwtSecret),
		log:       log,
	}
}

func (a *Auth) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized: missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "unauthorized: invalid header format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		claims, err := utils.ParseJWT(tokenStr, a.jwtSecret)
		if err != nil {
			a.log.Warn("invalid jwt token", zap.Error(err))
			http.Error(w, "unauthorized: invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
