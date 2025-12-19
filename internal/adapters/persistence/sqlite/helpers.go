package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ErrDatabaseCorrupted indicates the database file is corrupted and includes recovery suggestion
var ErrDatabaseCorrupted = errors.New("database file is corrupted")

// wrapDBError checks for database corruption indicators and wraps errors with recovery suggestions
func wrapDBError(err error, dbPath string) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())
	// Check for common SQLite corruption indicators
	if strings.Contains(errStr, "malformed") ||
		strings.Contains(errStr, "corrupt") ||
		strings.Contains(errStr, "disk i/o error") ||
		strings.Contains(errStr, "database disk image is malformed") {
		return fmt.Errorf("%w: %v. Recovery suggestion: delete %s and restart the application to recreate the database",
			ErrDatabaseCorrupted, err, dbPath)
	}
	return err
}

// projectRow is the database row representation for scanning
type projectRow struct {
	ID                 string         `db:"id"`
	Name               string         `db:"name"`
	Path               string         `db:"path"`
	DisplayName        sql.NullString `db:"display_name"`
	DetectedMethod     sql.NullString `db:"detected_method"`
	CurrentStage       string         `db:"current_stage"`
	Confidence         sql.NullString `db:"confidence"`
	DetectionReasoning sql.NullString `db:"detection_reasoning"`
	IsFavorite         int            `db:"is_favorite"`
	State              string         `db:"state"`
	Notes              sql.NullString `db:"notes"`
	PathMissing        int            `db:"path_missing"`
	LastActivityAt     string         `db:"last_activity_at"`
	CreatedAt          string         `db:"created_at"`
	UpdatedAt          string         `db:"updated_at"`
}

// rowToProject converts a database row to a domain Project
func rowToProject(row *projectRow) (*domain.Project, error) {
	// Parse timestamps (time.RFC3339Nano for nanosecond precision)
	lastActivity, err := time.Parse(time.RFC3339Nano, row.LastActivityAt)
	if err != nil {
		return nil, fmt.Errorf("invalid last_activity_at: %w", err)
	}
	created, err := time.Parse(time.RFC3339Nano, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at: %w", err)
	}
	updated, err := time.Parse(time.RFC3339Nano, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at: %w", err)
	}

	// Parse enums (use zero value on error)
	stage, _ := domain.ParseStage(row.CurrentStage)
	state, _ := domain.ParseProjectState(row.State)
	confidence, _ := domain.ParseConfidence(row.Confidence.String)

	return &domain.Project{
		ID:                 row.ID,
		Name:               row.Name,
		Path:               row.Path,
		DisplayName:        row.DisplayName.String,
		DetectedMethod:     row.DetectedMethod.String,
		CurrentStage:       stage,
		Confidence:         confidence,
		DetectionReasoning: row.DetectionReasoning.String,
		IsFavorite:         row.IsFavorite == 1,
		State:              state,
		Notes:              row.Notes.String,
		PathMissing:        row.PathMissing == 1,
		LastActivityAt:     lastActivity,
		CreatedAt:          created,
		UpdatedAt:          updated,
	}, nil
}

// nullString converts a string to sql.NullString, treating empty strings as NULL
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// boolToInt converts a boolean to int (0/1) for SQLite storage
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// stateToString converts ProjectState to lowercase string for database storage
func stateToString(state domain.ProjectState) string {
	switch state {
	case domain.StateActive:
		return "active"
	case domain.StateHibernated:
		return "hibernated"
	default:
		return "active"
	}
}
