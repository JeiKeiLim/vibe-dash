package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/statsview"
)

// walConnectionParams defines SQLite connection parameters for WAL mode and busy timeout.
// WAL mode allows concurrent reads while writing.
// Busy timeout (5000ms) allows SQLite to wait for write locks during concurrent access.
const walConnectionParams = "?_journal_mode=WAL&_busy_timeout=5000"

// MetricsRepository handles metrics persistence with graceful degradation.
// All public methods return nil/empty on errors - metrics failures never crash the dashboard.
type MetricsRepository struct {
	dbPath     string
	schemaOnce sync.Once
}

// NewMetricsRepository creates a repository for the metrics database.
// Does NOT initialize schema on construction - lazy init on first use.
func NewMetricsRepository(dbPath string) *MetricsRepository {
	return &MetricsRepository{dbPath: dbPath}
}

// openDB opens connection with WAL mode. Caller MUST close after use.
func (r *MetricsRepository) openDB(ctx context.Context) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "sqlite3", r.dbPath+walConnectionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to open metrics database: %w", err)
	}
	return db, nil
}

// ensureSchema creates tables on first access (idempotent, thread-safe)
func (r *MetricsRepository) ensureSchema(ctx context.Context, db *sqlx.DB) error {
	var schemaErr error
	r.schemaOnce.Do(func() {
		// Create schema_version table
		if _, err := db.ExecContext(ctx, CreateSchemaVersionTableSQL); err != nil {
			schemaErr = fmt.Errorf("failed to create schema_version table: %w", err)
			return
		}

		// Create stage_transitions table
		if _, err := db.ExecContext(ctx, CreateStageTransitionsTableSQL); err != nil {
			schemaErr = fmt.Errorf("failed to create stage_transitions table: %w", err)
			return
		}

		// Create indexes
		if _, err := db.ExecContext(ctx, CreateIndexProjectSQL); err != nil {
			schemaErr = fmt.Errorf("failed to create project index: %w", err)
			return
		}
		if _, err := db.ExecContext(ctx, CreateIndexTimeSQL); err != nil {
			schemaErr = fmt.Errorf("failed to create time index: %w", err)
			return
		}

		// Record schema version
		timestamp := time.Now().UTC().Format(time.RFC3339Nano)
		_, err := db.ExecContext(ctx, "INSERT OR IGNORE INTO schema_version (version, applied_at) VALUES (?, ?)",
			SchemaVersion, timestamp)
		if err != nil {
			schemaErr = fmt.Errorf("failed to record schema version: %w", err)
			return
		}
	})
	return schemaErr
}

// RecordTransition records a stage change event. Returns nil on ANY error (graceful degradation).
func (r *MetricsRepository) RecordTransition(ctx context.Context, projectID, fromStage, toStage string) error {
	db, err := r.openDB(ctx)
	if err != nil {
		slog.Warn("metrics database unavailable, skipping transition recording",
			"error", err, "project_id", projectID)
		return nil // Graceful degradation
	}
	defer db.Close()

	if err := r.ensureSchema(ctx, db); err != nil {
		slog.Warn("metrics schema creation failed", "error", err)
		return nil
	}

	id := generateUUID()
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	_, err = db.ExecContext(ctx, insertTransitionSQL, id, projectID, fromStage, toStage, timestamp)
	if err != nil {
		slog.Warn("failed to record stage transition", "error", err, "project_id", projectID)
		return nil
	}
	return nil
}

// GetTransitionsByProject retrieves transitions for a project since the given time.
// Returns empty slice on any error (graceful degradation).
func (r *MetricsRepository) GetTransitionsByProject(ctx context.Context, projectID string, since time.Time) []StageTransition {
	db, err := r.openDB(ctx)
	if err != nil {
		slog.Warn("metrics database unavailable", "error", err)
		return nil
	}
	defer db.Close()

	if err := r.ensureSchema(ctx, db); err != nil {
		slog.Warn("metrics schema error", "error", err)
		return nil
	}

	sinceStr := since.UTC().Format(time.RFC3339Nano)
	var rows []stageTransitionRow
	if err := db.SelectContext(ctx, &rows, selectByProjectWithTimeSQL, projectID, sinceStr); err != nil {
		slog.Warn("failed to query transitions", "error", err, "project_id", projectID)
		return nil
	}

	result := make([]StageTransition, len(rows))
	for i, row := range rows {
		result[i] = rowToTransition(&row)
	}
	return result
}

// GetTransitionsByTimeRange retrieves transitions within a time range.
// Returns empty slice on any error (graceful degradation).
func (r *MetricsRepository) GetTransitionsByTimeRange(ctx context.Context, from, to time.Time) []StageTransition {
	db, err := r.openDB(ctx)
	if err != nil {
		slog.Warn("metrics database unavailable", "error", err)
		return nil
	}
	defer db.Close()

	if err := r.ensureSchema(ctx, db); err != nil {
		slog.Warn("metrics schema error", "error", err)
		return nil
	}

	fromStr := from.UTC().Format(time.RFC3339Nano)
	toStr := to.UTC().Format(time.RFC3339Nano)
	var rows []stageTransitionRow
	if err := db.SelectContext(ctx, &rows, selectByTimeRangeSQL, fromStr, toStr); err != nil {
		slog.Warn("failed to query transitions by time range", "error", err)
		return nil
	}

	result := make([]StageTransition, len(rows))
	for i, row := range rows {
		result[i] = rowToTransition(&row)
	}
	return result
}

// GetTransitionTimestamps returns timestamps for TUI sparkline rendering.
// Implements statsview.MetricsReader interface.
// Avoids exposing internal types to TUI layer.
func (r *MetricsRepository) GetTransitionTimestamps(ctx context.Context, projectID string, since time.Time) []statsview.Transition {
	transitions := r.GetTransitionsByProject(ctx, projectID, since)
	result := make([]statsview.Transition, len(transitions))
	for i, t := range transitions {
		result[i] = statsview.Transition{TransitionedAt: t.TransitionedAt}
	}
	return result
}
