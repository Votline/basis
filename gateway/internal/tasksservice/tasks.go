// Package taskservice implement 'Service' interface
// for any tasks operations
package taskservice

import (
	"context"
	"net/http"

	"gateway/internal/services"

	"go.uber.org/zap"
)

type taskservice struct {
	name string
	log  *zap.Logger
}

func NewTS(mux *http.ServeMux, log *zap.Logger) (services.Service, error) {
	const op = "task_service.NewTS"

	return &taskservice{
		name: "task_service",
		log:  log,
	}, nil
}

func (us *taskservice) GetName() string {
	return us.name
}

func (us *taskservice) Close(ctx context.Context) error {
	const op = "task_service.Close"
	return nil
}

func (us *taskservice) RegisterRoutes(mux *http.ServeMux) {
	const op = "task_service.RegisterRoutes"

	mux.HandleFunc("POST /api/v1/tasks", nil)
	mux.HandleFunc("GET /api/v1/tasks", nil)
	mux.HandleFunc("PUT /api/v1/tasks/{id}", nil)
	mux.HandleFunc("GET /api/v1/tasks/{id}/history", nil)
}
