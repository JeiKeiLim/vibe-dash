package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// walConnectionParams defines SQLite connection parameters for WAL mode and busy timeout.
// WAL mode allows concurrent reads while writing.
// Busy timeout (5000ms) allows SQLite to wait for write locks during concurrent access.
const walConnectionParams = "?_journal_mode=WAL&_busy_timeout=5000"

// ProjectRepository implements ports.ProjectRepository for per-project SQLite databases.
// Each project has its own state.db file located at ~/.vibe-dash/<project>/state.db.
// Thread safety is achieved through lazy connections (each operation opens/closes its own connection).
type ProjectRepository struct {
	dbPath     string // Full path to state.db
	projectDir string // Project directory path
}

// NewProjectRepository creates a new per-project repository with fail-fast schema initialization.
// The projectDir must exist as a directory.
//
// On success, the state.db is created and schema is initialized.
// Returns domain.ErrPathNotAccessible on validation failures.
func NewProjectRepository(projectDir string) (*ProjectRepository, error) {
	// Validate projectDir exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: project directory does not exist: %s",
			domain.ErrPathNotAccessible, projectDir)
	}

	dbPath := filepath.Join(projectDir, "state.db")
	repo := &ProjectRepository{
		dbPath:     dbPath,
		projectDir: projectDir,
	}

	// Fail-fast: initialize schema on construction
	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// initSchema creates tables and runs migrations
func (r *ProjectRepository) initSchema() error {
	ctx := context.Background()
	db, err := r.openDB(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	return RunMigrations(ctx, db)
}

// openDB opens connection with WAL mode and busy timeout.
// Caller MUST close the connection after use.
// Returns ErrDatabaseCorrupted with recovery suggestion if database file is corrupted.
func (r *ProjectRepository) openDB(ctx context.Context) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "sqlite3", r.dbPath+walConnectionParams)
	if err != nil {
		if wrappedErr := wrapDBErrorForProject(err, r.dbPath); wrappedErr != err {
			return nil, wrappedErr
		}
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

// wrapDBErrorForProject checks for database corruption indicators and wraps errors with project-specific recovery suggestions.
func wrapDBErrorForProject(err error, dbPath string) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())
	// Check for common SQLite corruption indicators
	if strings.Contains(errStr, "malformed") ||
		strings.Contains(errStr, "corrupt") ||
		strings.Contains(errStr, "disk i/o error") ||
		strings.Contains(errStr, "database disk image is malformed") {
		return fmt.Errorf("%w: %v. Recovery: delete %s and re-add project via 'vibe add <path>'",
			ErrDatabaseCorrupted, err, dbPath)
	}
	return err
}

// Save creates or updates a project in the repository (upsert by ID).
func (r *ProjectRepository) Save(ctx context.Context, project *domain.Project) error {
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
		nullString(project.Confidence.String()),
		nullString(project.DetectionReasoning),
		boolToInt(project.IsFavorite),
		stateToString(project.State),
		nullString(project.Notes),
		boolToInt(project.PathMissing),
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
func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
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
func (r *ProjectRepository) FindByPath(ctx context.Context, path string) (*domain.Project, error) {
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
// Note: For per-project databases, this returns 0-1 results.
func (r *ProjectRepository) FindAll(ctx context.Context) ([]*domain.Project, error) {
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
// Note: For per-project databases, this returns 0-1 results.
func (r *ProjectRepository) FindActive(ctx context.Context) ([]*domain.Project, error) {
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
// Note: For per-project databases, this returns 0-1 results.
func (r *ProjectRepository) FindHibernated(ctx context.Context) ([]*domain.Project, error) {
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
func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
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
func (r *ProjectRepository) UpdateState(ctx context.Context, id string, state domain.ProjectState) error {
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

// UpdateLastActivity updates only the LastActivityAt timestamp.
// Returns domain.ErrProjectNotFound if no project exists with the given ID.
func (r *ProjectRepository) UpdateLastActivity(ctx context.Context, id string, timestamp time.Time) error {
	db, err := r.openDB(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.ExecContext(ctx, updateLastActivitySQL,
		timestamp.UTC().Format(time.RFC3339Nano),
		time.Now().UTC().Format(time.RFC3339Nano),
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update last activity: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

// Note: This file reuses helper functions and types from repository.go in the same package:
// - projectRow struct for DB row scanning
// - rowToProject() for converting DB rows to domain.Project
// - nullString(), boolToInt(), stateToString() for value conversion
// This avoids code duplication while maintaining separate repository implementations.
