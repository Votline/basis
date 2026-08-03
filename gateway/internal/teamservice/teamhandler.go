// Package teamservice teamhandler.go implements
// endpoints of team-service
package teamservice

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"gateway/internal/db"
	"gateway/internal/utils"

	"go.uber.org/zap"
)

func (ts *teamservice) newTeam(w http.ResponseWriter, r *http.Request) {
	const op = "teamservice.newTeam"

	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	team, err := ts.db.CreateTeam(r.Context(), req.Name, userID)
	if err != nil {
		ts.log.Error("failed to create team", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(team)
}

func (ts *teamservice) getTeams(w http.ResponseWriter, r *http.Request) {
	const op = "teamservice.getTeams"

	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	teams, err := ts.db.GetUserTeams(r.Context(), userID)
	if err != nil {
		ts.log.Error("failed to get user teams", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if teams == nil {
		teams = []db.Team{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(teams)
}

func (ts *teamservice) inviteByID(w http.ResponseWriter, r *http.Request) {
	const op = "teamservice.inviteByID"

	inviterID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || inviterID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	teamIDStr := r.PathValue("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil || teamID <= 0 {
		http.Error(w, "invalid team id", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID int64  `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID <= 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	role, err := ts.db.GetMemberRole(r.Context(), teamID, inviterID)
	if err != nil {
		if errors.Is(err, db.ErrPermissionDenied) {
			http.Error(w, "forbidden: not a team member", http.StatusForbidden)
			return
		}
		ts.log.Error("failed to check member role", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if role != "owner" && role != "admin" {
		http.Error(w, "forbidden: only owner or admin can invite", http.StatusForbidden)
		return
	}

	if err := ts.db.AddTeamMember(r.Context(), teamID, req.UserID, req.Role); err != nil {
		if errors.Is(err, db.ErrAlreadyInTeam) {
			http.Error(w, "user is already in team", http.StatusConflict)
			return
		}
		ts.log.Error("failed to add team member", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "invited",
	})
}
