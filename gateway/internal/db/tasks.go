// Package db tasks.go contains methods and structs for tasks-service
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	TeamID      int64     `json:"team_id"`
	AssigneeID  *int64    `json:"assignee_id"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskHistory struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	ChangedBy int64     `json:"changed_by"`
	FieldName string    `json:"field_changed"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskFilter struct {
	TeamID     int64
	Status     string
	AssigneeID *int64
	Limit      int
	Offset     int
}

var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrNotTeamMember   = errors.New("user is not a member of the team")
	ErrInvalidAssignee = errors.New("assignee must be a member of the team")
)

// IsTeamMember checks if user belongs to the specified team
func (m *DB) IsTeamMember(ctx context.Context, teamID, userID int64) (bool, error) {
	const op = "db.tasks.IsTeamMember"

	query := `SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = ? AND user_id = ?)`
	var exists bool
	err := m.DB.QueryRowContext(ctx, query, teamID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return exists, nil
}

// CreateTask inserts a new task into MySQL
func (m *DB) CreateTask(ctx context.Context, task *Task) (int64, error) {
	const op = "db.tasks.CreateTask"

	// Verify author belongs to the team
	isMember, err := m.IsTeamMember(ctx, task.TeamID, task.CreatedBy)
	if err != nil {
		return 0, fmt.Errorf("%s: check author membership: %w", op, err)
	}
	if !isMember {
		return 0, ErrNotTeamMember
	}

	// verify assignee belongs to team
	if task.AssigneeID != nil {
		isAssigneeMember, err := m.IsTeamMember(ctx, task.TeamID, *task.AssigneeID)
		if err != nil {
			return 0, fmt.Errorf("%s: check assignee membership: %w", op, err)
		}
		if !isAssigneeMember {
			return 0, ErrInvalidAssignee
		}
	}

	if task.Status == "" {
		task.Status = "todo"
	}

	query := `
		INSERT INTO tasks (title, description, status, team_id, assignee_id, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	res, err := m.DB.ExecContext(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.TeamID,
		task.AssigneeID,
		task.CreatedBy,
	)
	if err != nil {
		return 0, fmt.Errorf("%s: exec insert: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: last insert id: %w", op, err)
	}

	return id, nil
}

// GetTaskByID retrieves task by its ID
func (m *DB) GetTaskByID(ctx context.Context, id int64) (*Task, error) {
	const op = "db.tasks.GetTaskByID"

	query := `
		SELECT id, title, description, status, team_id, assignee_id, created_by, created_at, updated_at
		FROM tasks WHERE id = ?
	`

	task := &Task{}
	var assignee sql.NullInt64

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.TeamID,
		&assignee,
		&task.CreatedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("%s: scan: %w", op, err)
	}

	if assignee.Valid {
		val := assignee.Int64
		task.AssigneeID = &val
	}

	return task, nil
}

// GetTasks fetches filtered list of tasks with pagination
func (m *DB) GetTasks(ctx context.Context, filter TaskFilter) ([]Task, error) {
	const op = "db.tasks.GetTasks"

	var conditions []string
	var args []any

	conditions = append(conditions, "team_id = ?")
	args = append(args, filter.TeamID)

	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	if filter.AssigneeID != nil {
		conditions = append(conditions, "assignee_id = ?")
		args = append(args, *filter.AssigneeID)
	}

	whereClause := strings.Join(conditions, " AND ")

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, status, team_id, assignee_id, created_by, created_at, updated_at
		FROM tasks
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		var t Task
		var assignee sql.NullInt64

		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.TeamID,
			&assignee,
			&t.CreatedBy,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan row: %w", op, err)
		}

		if assignee.Valid {
			val := assignee.Int64
			t.AssigneeID = &val
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows err: %w", op, err)
	}

	return tasks, nil
}

// UpdateTask updates task details and logs history in a single transaction
func (m *DB) UpdateTask(ctx context.Context, updatedTask *Task, changedBy int64) error {
	const op = "db.tasks.UpdateTask"

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	// lock existing task for update
	var oldTask Task
	var assignee sql.NullInt64
	querySelect := `
		SELECT id, title, description, status, team_id, assignee_id, created_by
		FROM tasks WHERE id = ? FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, querySelect, updatedTask.ID).Scan(
		&oldTask.ID,
		&oldTask.Title,
		&oldTask.Description,
		&oldTask.Status,
		&oldTask.TeamID,
		&assignee,
		&oldTask.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("%s: lock task: %w", op, err)
	}
	if assignee.Valid {
		val := assignee.Int64
		oldTask.AssigneeID = &val
	}

	// validate
	if updatedTask.AssigneeID != nil {
		isAssigneeMember, err := m.IsTeamMember(ctx, oldTask.TeamID, *updatedTask.AssigneeID)
		if err != nil {
			return fmt.Errorf("%s: check new assignee: %w", op, err)
		}
		if !isAssigneeMember {
			return ErrInvalidAssignee
		}
	}

	// track changes
	type change struct {
		field string
		oldV  string
		newV  string
	}
	var changes []change

	if updatedTask.Title != "" && updatedTask.Title != oldTask.Title {
		changes = append(changes, change{"title", oldTask.Title, updatedTask.Title})
		oldTask.Title = updatedTask.Title
	}
	if updatedTask.Description != oldTask.Description {
		changes = append(changes, change{"description", oldTask.Description, updatedTask.Description})
		oldTask.Description = updatedTask.Description
	}
	if updatedTask.Status != "" && updatedTask.Status != oldTask.Status {
		changes = append(changes, change{"status", oldTask.Status, updatedTask.Status})
		oldTask.Status = updatedTask.Status
	}

	oldAssigneeStr := ""
	if oldTask.AssigneeID != nil {
		oldAssigneeStr = fmt.Sprintf("%d", *oldTask.AssigneeID)
	}
	newAssigneeStr := ""
	if updatedTask.AssigneeID != nil {
		newAssigneeStr = fmt.Sprintf("%d", *updatedTask.AssigneeID)
	}
	if oldAssigneeStr != newAssigneeStr {
		changes = append(changes, change{"assignee_id", oldAssigneeStr, newAssigneeStr})
		oldTask.AssigneeID = updatedTask.AssigneeID
	}

	if len(changes) == 0 {
		return nil
	}

	// execute UPDATE
	queryUpdate := `
		UPDATE tasks
		SET title = ?, description = ?, status = ?, assignee_id = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, queryUpdate,
		oldTask.Title,
		oldTask.Description,
		oldTask.Status,
		oldTask.AssigneeID,
		oldTask.ID,
	)
	if err != nil {
		return fmt.Errorf("%s: exec update: %w", op, err)
	}

	// insert history logs
	queryHistory := `
		INSERT INTO task_history (task_id, changed_by, field_changed, old_value, new_value)
		VALUES (?, ?, ?, ?, ?)
	`
	histStmt, err := tx.PrepareContext(ctx, queryHistory)
	if err != nil {
		return fmt.Errorf("%s: prep history stmt: %w", op, err)
	}
	defer histStmt.Close()

	for _, c := range changes {
		_, err := histStmt.ExecContext(ctx, oldTask.ID, changedBy, c.field, c.oldV, c.newV)
		if err != nil {
			return fmt.Errorf("%s: insert history record: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit tx: %w", op, err)
	}

	updatedTask.TeamID = oldTask.TeamID
	return nil
}

// GetTaskHistory returns audit logs for a given task
func (m *DB) GetTaskHistory(ctx context.Context, taskID int64) ([]TaskHistory, error) {
	const op = "db.tasks.GetTaskHistory"

	// Меняем created_at на changed_at в SELECT и ORDER BY
	query := `
		SELECT id, task_id, changed_by, field_changed, old_value, new_value, changed_at
		FROM task_history
		WHERE task_id = ?
		ORDER BY changed_at DESC
	`

	rows, err := m.DB.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	history := make([]TaskHistory, 0)
	for rows.Next() {
		var h TaskHistory
		if err := rows.Scan(
			&h.ID,
			&h.TaskID,
			&h.ChangedBy,
			&h.FieldName,
			&h.OldValue,
			&h.NewValue,
			&h.CreatedAt, // сканим таймштамп из базы прямо в твой h.CreatedAt
		); err != nil {
			return nil, fmt.Errorf("%s: scan row: %w", op, err)
		}
		history = append(history, h)
	}

	return history, nil
}
