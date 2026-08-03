// Package services provides service interfac and
// execution methods
package services

import (
	"context"
	"net/http"
)

type Service interface {
	// GetName returns service name
	GetName() string

	// RegisterRoutes register endpoints for service
	RegisterRoutes(mux *http.ServeMux)

	// Close gracefully shutdowns service
	Close(ctx context.Context) error
}
