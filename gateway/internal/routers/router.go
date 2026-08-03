// Package routers init services,
// register routers and create http.Server
package routers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"gateway/internal/db"
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
	protectedMux := http.NewServeMux()

	svcs, err := initServices(mux, protectedMux, log)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("%s: get jwt secret: no jwt secret found", op)
	}
	auth := middlewares.NewAuth(jwtSecret, log)

	mux.Handle("/", auth.Process(protectedMux))

	rc, err := rdb.NewRC(log)
	if err != nil {
		return nil, fmt.Errorf("%s: init rdb: %w", op, err)
	}

	handler := attachMiddlewares(mux, rc, log)

	log.Debug("Successfully created server", zap.String("op", op))

	return &Server{
		srv: &http.Server{
			Addr:    ":8080",
			Handler: handler,
		},
		log:  log,
		svcs: svcs,
	}, nil
}

// initServices creates all services and registers their endpoints to respective muxes
func initServices(mux, pmux *http.ServeMux, log *zap.Logger) ([]services.Service, error) {
	const op = "routers.initServices"

	svcs := make([]services.Service, 0, 3)

	db, err := db.NewDB(log)
	if err != nil {
		return nil, fmt.Errorf("%s: create db: %w", op, err)
	}

	us, err := usersservice.NewUS(mux, db, log)
	if err != nil {
		return nil, fmt.Errorf("%s: init users-service: %w", op, err)
	}
	us.RegisterRoutes(mux)

	teamsS, err := teamservice.NewTS(pmux, db, log)
	if err != nil {
		return nil, fmt.Errorf("%s: init team-service: %w", op, err)
	}
	teamsS.RegisterRoutes(pmux)

	tasksS, err := taskservice.NewTS(pmux, log)
	if err != nil {
		return nil, fmt.Errorf("%s: init tasks-service: %w", op, err)
	}
	tasksS.RegisterRoutes(pmux)

	svcs = append(svcs, us, teamsS, tasksS)

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
