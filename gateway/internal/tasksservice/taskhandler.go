// Package taskservice taskhandler.go implements
// endpoints of tasks-service
package taskservice

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"gateway/internal/db"
	"gateway/internal/utils"

	"go.uber.org/zap"
)

func (ts *taskservice) newTask(w http.ResponseWriter, r *http.Request) {
	const op = "tasksservice.newTask"

	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		TeamID      int64  `json:"team_id"`
		AssigneeID  *int64 `json:"assignee_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.TeamID <= 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newTask := &db.Task{
		Title:       req.Title,
		Description: req.Description,
		TeamID:      req.TeamID,
		AssigneeID:  req.AssigneeID,
		CreatedBy:   userID,
	}

	id, err := ts.db.CreateTask(r.Context(), newTask)
	if err != nil {
		if errors.Is(err, db.ErrNotTeamMember) {
			http.Error(w, "forbidden: not a team member", http.StatusForbidden)
			return
		}
		if errors.Is(err, db.ErrInvalidAssignee) {
			http.Error(w, "invalid assignee: not a team member", http.StatusBadRequest)
			return
		}
		ts.log.Error("failed to create task", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	createdTask, err := ts.db.GetTaskByID(r.Context(), id)
	if err != nil {
		ts.log.Error("failed to fetch created task", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createdTask)
}

func (ts *taskservice) getTasks(w http.ResponseWriter, r *http.Request) {
	const op = "tasksservice.getTasks"

	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	teamIDStr := q.Get("team_id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil || teamID <= 0 {
		http.Error(w, "invalid team_id parameter", http.StatusBadRequest)
		return
	}

	isMember, err := ts.db.IsTeamMember(r.Context(), teamID, userID)
	if err != nil {
		ts.log.Error("failed to check team membership", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "forbidden: not a team member", http.StatusForbidden)
		return
	}

	var assigneeID *int64
	if aStr := q.Get("assignee_id"); aStr != "" {
		if parsedA, err := strconv.ParseInt(aStr, 10, 64); err == nil {
			assigneeID = &parsedA
		}
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := db.TaskFilter{
		TeamID:     teamID,
		Status:     q.Get("status"),
		AssigneeID: assigneeID,
		Limit:      limit,
		Offset:     offset,
	}

	tasks, err := ts.db.GetTasks(r.Context(), filter)
	if err != nil {
		ts.log.Error("failed to get tasks", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []db.Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tasks)
}

func (ts *taskservice) updTask(w http.ResponseWriter, r *http.Request) {
	const op = "tasksservice.updTask"

	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskIDStr := r.PathValue("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil || taskID <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	existingTask, err := ts.db.GetTaskByID(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		ts.log.Error("failed to get task for update check", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	isMember, err := ts.db.IsTeamMember(r.Context(), existingTask.TeamID, userID)
	if err != nil {
		ts.log.Error("failed to check team membership", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "forbidden: not a team member", http.StatusForbidden)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
		AssigneeID  *int64 `json:"assignee_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updateData := &db.Task{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssigneeID:  req.AssigneeID,
	}

	if err := ts.db.UpdateTask(r.Context(), updateData, userID); err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, db.ErrInvalidAssignee) {
			http.Error(w, "invalid assignee: not a team member", http.StatusBadRequest)
			return
		}
		ts.log.Error("failed to update task", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	updatedTask, err := ts.db.GetTaskByID(r.Context(), taskID)
	if err != nil {
		ts.log.Error("failed to fetch updated task", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updatedTask)
}

func (ts *taskservice) getTaskHistory(w http.ResponseWriter, r *http.Request) {
	const op = "tasksservice.getTaskHistory"

	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskIDStr := r.PathValue("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil || taskID <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	task, err := ts.db.GetTaskByID(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		ts.log.Error("failed to fetch task for history permissions check", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	isMember, err := ts.db.IsTeamMember(r.Context(), task.TeamID, userID)
	if err != nil {
		ts.log.Error("failed to check team membership", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "forbidden: not a team member", http.StatusForbidden)
		return
	}

	history, err := ts.db.GetTaskHistory(r.Context(), taskID)
	if err != nil {
		ts.log.Error("failed to get task history", zap.String("op", op), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if history == nil {
		history = []db.TaskHistory{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(history)
}
