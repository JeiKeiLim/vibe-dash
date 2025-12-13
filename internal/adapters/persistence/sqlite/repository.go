package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

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

// SQLiteRepository implements ports.ProjectRepository using SQLite
type SQLiteRepository struct {
	dbPath string
}

// NewSQLiteRepository creates a new SQLite repository with fail-fast schema initialization.
// If dbPath is empty, defaults to ~/.vibe-dash/projects.db
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(home, ".vibe-dash", "projects.db")
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	repo := &SQLiteRepository{dbPath: dbPath}

	// Fail-fast: initialize schema on construction
	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// initSchema creates tables and runs migrations
func (r *SQLiteRepository) initSchema() error {
	ctx := context.Background()
	db, err := r.openDB(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	return RunMigrations(ctx, db)
}

// openDB opens connection with WAL mode and busy timeout.
// Busy timeout (5000ms) allows SQLite to wait for write locks during concurrent access.
// Caller MUST close the connection after use.
// Returns ErrDatabaseCorrupted with recovery suggestion if database file is corrupted.
func (r *SQLiteRepository) openDB(ctx context.Context) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "sqlite3", r.dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		if wrappedErr := wrapDBError(err, r.dbPath); wrappedErr != err {
			return nil, wrappedErr
		}
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

// Save creates or updates a project in the repository (upsert by ID).
// Note: This method modifies project.UpdatedAt to the current time before saving.
// The caller's project object will reflect this change even if the save fails afterward.
func (r *SQLiteRepository) Save(ctx context.Context, project *domain.Project) error {
	if err := project.Validate(); err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	db, err := r.openDB(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	project.UpdatedAt = time.Now()

	_, err = db.ExecContext(ctx, insertOrReplaceProjectSQL,
		project.ID,
		project.Name,
		project.Path,
		nullString(project.DisplayName),
		nullString(project.DetectedMethod),
		project.CurrentStage.String(),
		"", // confidence - populated by detection service (Story 2.4)
		"", // detection_reasoning - populated by detection service (Story 2.4)
		boolToInt(project.IsFavorite),
		stateToString(project.State),
		nullString(project.Notes),
		project.LastActivityAt.Format(time.RFC3339Nano),
		project.CreatedAt.Format(time.RFC3339Nano),
		project.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	return nil
}

// FindByID retrieves a project by its unique identifier (path hash).
// Returns domain.ErrProjectNotFound if no project exists with the given ID.
func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	db, err := r.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var row projectRow
	if err := db.GetContext(ctx, &row, selectByIDSQL, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to find project by ID: %w", err)
	}

	return rowToProject(&row)
}

// FindByPath retrieves a project by its canonical absolute path.
// Returns domain.ErrProjectNotFound if no project exists at the given path.
func (r *SQLiteRepository) FindByPath(ctx context.Context, path string) (*domain.Project, error) {
	db, err := r.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var row projectRow
	if err := db.GetContext(ctx, &row, selectByPathSQL, path); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to find project by path: %w", err)
	}

	return rowToProject(&row)
}

// FindAll retrieves all projects regardless of their state.
// Returns an empty slice (not nil) if no projects exist.
func (r *SQLiteRepository) FindAll(ctx context.Context) ([]*domain.Project, error) {
	db, err := r.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var rows []projectRow
	if err := db.SelectContext(ctx, &rows, selectAllSQL); err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	// CRITICAL: Return empty slice, not nil
	projects := make([]*domain.Project, 0, len(rows))
	for _, row := range rows {
		project, err := rowToProject(&row)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// FindActive retrieves only projects with StateActive.
// Returns an empty slice (not nil) if no active projects exist.
func (r *SQLiteRepository) FindActive(ctx context.Context) ([]*domain.Project, error) {
	db, err := r.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var rows []projectRow
	if err := db.SelectContext(ctx, &rows, selectActiveSQL); err != nil {
		return nil, fmt.Errorf("failed to query active projects: %w", err)
	}

	// CRITICAL: Return empty slice, not nil
	projects := make([]*domain.Project, 0, len(rows))
	for _, row := range rows {
		project, err := rowToProject(&row)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// FindHibernated retrieves only projects with StateHibernated.
// Returns an empty slice (not nil) if no hibernated projects exist.
func (r *SQLiteRepository) FindHibernated(ctx context.Context) ([]*domain.Project, error) {
	db, err := r.openDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var rows []projectRow
	if err := db.SelectContext(ctx, &rows, selectHibernatedSQL); err != nil {
		return nil, fmt.Errorf("failed to query hibernated projects: %w", err)
	}

	// CRITICAL: Return empty slice, not nil
	projects := make([]*domain.Project, 0, len(rows))
	for _, row := range rows {
		project, err := rowToProject(&row)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// Delete removes a project by its unique identifier.
// Returns domain.ErrProjectNotFound if no project exists with the given ID.
func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	db, err := r.openDB(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.ExecContext(ctx, deleteByIDSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

// UpdateState changes a project's active/hibernated state.
// Returns domain.ErrProjectNotFound if no project exists with the given ID.
func (r *SQLiteRepository) UpdateState(ctx context.Context, id string, state domain.ProjectState) error {
	db, err := r.openDB(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.ExecContext(ctx, updateStateSQL,
		stateToString(state),
		time.Now().Format(time.RFC3339Nano),
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

// projectRow is the database row representation for scanning
type projectRow struct {
	ID                 string         `db:"id"`
	Name               string         `db:"name"`
	Path               string         `db:"path"`
	DisplayName        sql.NullString `db:"display_name"`
	DetectedMethod     sql.NullString `db:"detected_method"`
	CurrentStage       string         `db:"current_stage"`
	Confidence         sql.NullString `db:"confidence"`          // Read but not mapped to Project
	DetectionReasoning sql.NullString `db:"detection_reasoning"` // Read but not mapped to Project
	IsFavorite         int            `db:"is_favorite"`
	State              string         `db:"state"`
	Notes              sql.NullString `db:"notes"`
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

	return &domain.Project{
		ID:             row.ID,
		Name:           row.Name,
		Path:           row.Path,
		DisplayName:    row.DisplayName.String,
		DetectedMethod: row.DetectedMethod.String,
		CurrentStage:   stage,
		IsFavorite:     row.IsFavorite == 1,
		State:          state,
		Notes:          row.Notes.String,
		LastActivityAt: lastActivity,
		CreatedAt:      created,
		UpdatedAt:      updated,
	}, nil
	// Note: Confidence and DetectionReasoning are NOT mapped - they belong to DetectionResult
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
