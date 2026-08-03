// Package routers init services,
// register routers and create http.Server
package routers

import (
	"context"
	"fmt"
	"net/http"

	"gateway/internal/services"
	"gateway/internal/usersservice"

	"go.uber.org/zap"
)

type Server struct {
	srv  *http.Server
	log  *zap.Logger
	svcs []services.Service
}

func Init(log *zap.Logger) (*Server, error) {
	const op = "routers.Init"

	mux := http.NewServeMux()
	svcs, err := initServices(mux, log)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var handler http.Handler = mux

	attachMiddlewares(handler)

	return &Server{
		srv: &http.Server{
			Addr:    ":8080",
			Handler: mux,
		},
		log:  log,
		svcs: svcs,
	}, nil
}

// initServices creates all services
// and register they endpoints to the mux
func initServices(mux *http.ServeMux, log *zap.Logger) ([]services.Service, error) {
	const op = "routers.initServices"

	svcs := make([]services.Service, 0, 1)

	us, err := usersservice.NewUS(mux, log)
	if err != nil {
		return nil, fmt.Errorf("%s: init users-service: %w", op, err)
	}
	svcs = append(svcs, us)

	for _, svc := range svcs {
		svc.RegisterRoutes(mux)
	}

	return svcs, nil
}

// attachMiddlewares create middlewares and
// attach them to the handler
func attachMiddlewares(handler http.Handler) {
	const op = "routers.attachMiddlewares"
}

// Close gracefully shuts down http server
// and all registered services
func (s *Server) Close(ctx context.Context) error {
	const op = "routers.Close"

	for _, svc := range s.svcs {
		if err := svc.Close(ctx); err != nil {
			s.log.Error("Failed to shutdown service",
				zap.String("Service name", svc.GetName()),
				zap.String("op", op))
		}
	}

	return nil
}
