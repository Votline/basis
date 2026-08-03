// Package middlewares contains middlewares implementations
package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Middleware interface {
	Process(next http.Handler) http.Handler
}

type MiddlewareFunc func(next http.Handler) http.Handler

func (f MiddlewareFunc) Process(next http.Handler) http.Handler {
	return f(next)
}

// Chain wrapping final Header in chain of middlewares
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i].Process(h)
	}
	return h
}

// Logging middleware logs every request
type loggingMiddleware struct {
	log *zap.Logger
}

func Logging(log *zap.Logger) Middleware {
	return &loggingMiddleware{log: log}
}

func (m *loggingMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		m.log.Info("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

// Recovery middleware recovers from panic
type recoveryMiddleware struct {
	log *zap.Logger
}

func Recovery(log *zap.Logger) Middleware {
	return &recoveryMiddleware{log: log}
}

func (m *recoveryMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.log.Error("panic recovered in http handler",
					zap.Any("error", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
