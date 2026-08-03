// Package usersservice implement 'Service' interface
// for any users operations
package usersservice

import (
	"context"
	"net/http"

	"gateway/internal/services"

	"go.uber.org/zap"
)

type UsersService struct {
	name string
	log  *zap.Logger
}

func NewUS(mux *http.ServeMux, log *zap.Logger) (services.Service, error) {
	const op = "users_service.NewUS"

	return &UsersService{
		name: "users_service",
		log:  log,
	}, nil
}

func (us *UsersService) GetName() string {
	return us.name
}

func (us *UsersService) Close(ctx context.Context) error {
	const op = "users_service.Close"
	return nil
}

func (us *UsersService) RegisterRoutes(mux *http.ServeMux) {
	const op = "users_service.RegisterRoutes"

	mux.HandleFunc("POST /api/v1/register", nil)
	mux.HandleFunc("POST /api/v1/login", nil)
}
