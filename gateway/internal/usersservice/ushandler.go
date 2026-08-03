// Package usersservice ushandler.go implements
// endpoints of users-service
package usersservice

import (
	"encoding/json"
	"errors"
	"net/http"

	"gateway/internal/db"
	"gateway/internal/utils"

	"go.uber.org/zap"
)

func (us *UsersService) register(w http.ResponseWriter, r *http.Request) {
	const op = "usersservice.register"

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		us.log.Error("failed to decode register req", zap.String("op", op), zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email, password and name are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, "password must be at least 6 characters long", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		us.log.Error("failed to hash password", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	userID, err := us.db.CreateUser(r.Context(), req.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, db.ErrUserAlreadyExists) {
			http.Error(w, "user with this email already exists", http.StatusConflict)
			return
		}
		us.log.Error("failed to create user", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":    userID,
		"email": req.Email,
	})
}

func (us *UsersService) login(w http.ResponseWriter, r *http.Request) {
	const op = "usersservice.login"
}
