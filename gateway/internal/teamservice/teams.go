// Package teamservice implement 'Service' interface
// for any teams operations
package teamservice

import (
	"context"
	"net/http"

	"gateway/internal/services"

	"go.uber.org/zap"
)

type teamservice struct {
	name string
	log  *zap.Logger
}

func NewTS(mux *http.ServeMux, log *zap.Logger) (services.Service, error) {
	const op = "team_service.NewTS"

	return &teamservice{
		name: "team_service",
		log:  log,
	}, nil
}

func (us *teamservice) GetName() string {
	return us.name
}

func (us *teamservice) Close(ctx context.Context) error {
	const op = "team_service.Close"
	return nil
}

func (us *teamservice) RegisterRoutes(mux *http.ServeMux) {
	const op = "team_service.RegisterRoutes"

	mux.HandleFunc("POST /api/v1/teams", nil)
	mux.HandleFunc("GET /api/v1/teams", nil)
	mux.HandleFunc("POST /api/v1/teams/{id}/invite", nil)
}
