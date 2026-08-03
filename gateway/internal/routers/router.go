// Package routers init services,
// register routers and create http.Server
package routers

import (
	"context"
	"fmt"
	"net/http"

	"gateway/internal/middlewares"
	"gateway/internal/rdb"
	"gateway/internal/services"
	taskservice "gateway/internal/tasksservice"
	"gateway/internal/teamservice"
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

	rc, err := rdb.NewRC(log)
	if err != nil {
		return nil, fmt.Errorf("%s: init rdb: %w", op, err)
	}

	handler := attachMiddlewares(mux, rc, log)

	log.Debug("Successfully created server",
		zap.String("op", op))

	return &Server{
		srv: &http.Server{
			Addr:    ":8080",
			Handler: handler,
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

	teamsS, err := teamservice.NewTS(mux, log)
	if err != nil {
		return nil, fmt.Errorf("%s: init team-service: %w", op, err)
	}

	tasksS, err := taskservice.NewTS(mux, log)
	if err != nil {
		return nil, fmt.Errorf("%s: init tasks-service: %w", op, err)
	}

	svcs = append(svcs, us, teamsS, tasksS)

	for _, svc := range svcs {
		svc.RegisterRoutes(mux)
	}

	return svcs, nil
}

// attachMiddlewares create middlewares and
// attach them to the handler
func attachMiddlewares(mux http.Handler, rc *rdb.RedisClient, log *zap.Logger) http.Handler {
	const op = "routers.attachMiddlewares"

	rt := middlewares.NewRateLimiter(rc, log)

	return middlewares.Chain(
		mux,
		middlewares.Recovery(log),
		middlewares.Logging(log),
		rt,
	)
}

// Start starts the http server
func (s *Server) Start() error {
	const op = "router.Start"
	s.log.Info("Starting server...", zap.String("op", op))

	if err := s.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("%s: listen and serve: %w", op, err)
	}
	return nil
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
