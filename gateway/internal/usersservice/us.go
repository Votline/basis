// Package usersservice us.go implement 'Service' interface
// for any users operations
package usersservice

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"gateway/internal/db"
	"gateway/internal/services"

	"go.uber.org/zap"
)

type UsersService struct {
	name      string
	jwtSecret []byte
	log       *zap.Logger
	db        *db.DB
}

func NewUS(mux *http.ServeMux, log *zap.Logger) (services.Service, error) {
	const op = "users_service.NewUS"

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		return nil, fmt.Errorf("%s: get jwt secret: nil jwt secret:", op)
	}

	db, err := db.NewDB(log)
	if err != nil {
		return nil, fmt.Errorf("%s: create db: %w", op, err)
	}

	log.Debug("Successfully created users service",
		zap.String("op", op))

	return &UsersService{
		name:      "users_service",
		jwtSecret: jwtSecret,
		log:       log,
		db:        db,
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

	mux.HandleFunc("POST /api/v1/register", us.register)
	mux.HandleFunc("POST /api/v1/login", us.login)
}
