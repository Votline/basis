// Package db teams.go contains methods for team management
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Team struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	Role      string    `json:"role,omitempty"`
}

type TeamMember struct {
	TeamID   int64     `json:"team_id"`
	UserID   int64     `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

var (
	ErrTeamNotFound     = errors.New("team not found")
	ErrAlreadyInTeam    = errors.New("user is already a member of this team")
	ErrPermissionDenied = errors.New("permission denied")
)

// CreateTeam creates new teams and makes creator - owner
func (m *DB) CreateTeam(ctx context.Context, name string, creatorID int64) (*Team, error) {
	const op = "db.teams.CreateTeam"

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	// create team
	queryTeam := `INSERT INTO teams (name, created_by) VALUES (?, ?)`
	res, err := tx.ExecContext(ctx, queryTeam, name, creatorID)
	if err != nil {
		return nil, fmt.Errorf("%s: insert team: %w", op, err)
	}

	teamID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("%s: last insert id: %w", op, err)
	}

	// creator as owner
	queryMember := `INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, 'owner')`
	_, err = tx.ExecContext(ctx, queryMember, teamID, creatorID)
	if err != nil {
		return nil, fmt.Errorf("%s: insert team member owner: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return &Team{
		ID:        teamID,
		Name:      name,
		CreatedBy: creatorID,
		Role:      "owner",
	}, nil
}

// GetUserTeams returns list of all, where contains userID
func (m *DB) GetUserTeams(ctx context.Context, userID int64) ([]Team, error) {
	const op = "db.teams.GetUserTeams"

	query := `
		SELECT t.id, t.name, t.created_by, t.created_at, tm.role 
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = ?
		ORDER BY t.created_at DESC`

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	var teams []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt, &t.Role); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		teams = append(teams, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows err: %w", op, err)
	}

	return teams, nil
}

// GetMemberRole returns the role of the user in the team
func (m *DB) GetMemberRole(ctx context.Context, teamID, userID int64) (string, error) {
	const op = "db.teams.GetMemberRole"

	query := `SELECT role FROM team_members WHERE team_id = ? AND user_id = ?`

	var role string
	err := m.DB.QueryRowContext(ctx, query, teamID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrPermissionDenied
		}
		return "", fmt.Errorf("%s: scan: %w", op, err)
	}

	return role, nil
}

// AddTeamMember add user to command with role
func (m *DB) AddTeamMember(ctx context.Context, teamID, userID int64, role string) error {
	const op = "db.teams.AddTeamMember"

	query := `INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)`

	_, err := m.DB.ExecContext(ctx, query, teamID, userID, role)
	if err != nil {
		if isDuplicateEntryError(err) {
			return ErrAlreadyInTeam
		}
		return fmt.Errorf("%s: exec: %w", op, err)
	}

	return nil
}
