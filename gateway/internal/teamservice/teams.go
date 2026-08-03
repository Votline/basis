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

	log.Debug("Successfully created team service",
		zap.String("op", op))

	return &teamservice{
		name: "team_service",
		log:  log,
	}, nil
}

func (ts *teamservice) GetName() string {
	return ts.name
}

func (ts *teamservice) Close(ctx context.Context) error {
	const op = "team_service.Close"
	return nil
}

func (ts *teamservice) RegisterRoutes(mux *http.ServeMux) {
	const op = "team_service.RegisterRoutes"

	mux.HandleFunc("POST /api/v1/teams", ts.newTeam)
	mux.HandleFunc("GET /api/v1/teams", ts.getTeams)
	mux.HandleFunc("POST /api/v1/teams/{id}/invite", ts.inviteByID)
}
