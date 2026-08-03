// Package taskservice tasks.go implement 'Service' interface
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

	log.Debug("Successfully created task service",
		zap.String("op", op))

	return &taskservice{
		name: "task_service",
		log:  log,
	}, nil
}

func (ts *taskservice) GetName() string {
	return ts.name
}

func (ts *taskservice) Close(ctx context.Context) error {
	const op = "task_service.Close"
	return nil
}

func (ts *taskservice) RegisterRoutes(mux *http.ServeMux) {
	const op = "task_service.RegisterRoutes"

	mux.HandleFunc("POST /api/v1/tasks", ts.newTask)
	mux.HandleFunc("GET /api/v1/tasks", ts.getTasks)
	mux.HandleFunc("PUT /api/v1/tasks/{id}", ts.updTask)
	mux.HandleFunc("GET /api/v1/tasks/{id}/history", ts.getTaskHistory)
}
