// Package routers init services,
// register routers and create http.Server
package routers

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	srv *http.Server
	log *zap.Logger
}

func Init(log *zap.Logger) (*Server, error) {
	const op = "routers.Init"

	mux := http.NewServeMux()
	if err := initServices(mux); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var handler http.Handler = mux

	attachMiddlewares(handler)

	return &Server{
		srv: &http.Server{
			Addr:    ":8080",
			Handler: mux,
		},
		log: log,
	}, nil
}

// initServices creates all services
// and register they endpoints to the mux
func initServices(mux *http.ServeMux) error {
	const op = "routers.initServices"
	return nil
}

// attachMiddlewares create middlewares and
// attach them to the handler
func attachMiddlewares(handler http.Handler) {
	const op = "routers.attachMiddlewares"
	return
}

// Close gracefully shuts down http server
// and all registered services
func (s *Server) Close(ctx context.Context) error {
	const op = "routers.Close"
	return nil
}
