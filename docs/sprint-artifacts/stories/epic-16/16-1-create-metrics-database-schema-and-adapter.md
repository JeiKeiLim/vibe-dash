# Story 16.1: Create Metrics Database Schema and Adapter

Status: done

## Story

As a developer,
I want a separate metrics database with clean schema,
So that metrics can be removed without affecting core functionality.

## User-Visible Changes

None - this is an internal infrastructure change for the experimental metrics feature. The metrics database enables future stats view functionality but has no immediate user-facing impact.

## Acceptance Criteria

1. **Given** first metrics event recorded
   **When** adapter initializes
   **Then** creates `~/.vibe-dash/metrics.db` (separate from per-project state.db files)

2. **And** schema includes `stage_transitions` table with columns:
   - `id TEXT PRIMARY KEY` - UUID for event
   - `project_id TEXT NOT NULL` - References project path hash
   - `from_stage TEXT NOT NULL` - Previous stage (empty string for first detection)
   - `to_stage TEXT NOT NULL` - New stage
   - `transitioned_at TEXT NOT NULL` - ISO 8601 timestamp (time.RFC3339Nano format)

3. **And** metrics database failure does not crash the dashboard
   - Graceful degradation: log warning, continue without metrics
   - Core TUI functionality remains unaffected

4. **And** database uses SQLite WAL mode for concurrent read access

5. **And** adapter follows existing repository patterns (lazy connections, open-use-close)

## Tasks / Subtasks

- [x] Task 1: Create metrics repository package structure (AC: #1)
  - [x] Create `internal/adapters/persistence/metrics/` directory
  - [x] Create `repository.go` with `MetricsRepository` struct containing `dbPath string` field
  - [x] Create `schema.go` with schema SQL constants
  - [x] Create `queries.go` with SQL query constants
  - [x] Create `helpers.go` with row struct and conversion helpers

- [x] Task 2: Implement metrics database schema (AC: #2, #4)
  - [x] Define `CreateSchemaVersionTableSQL` (reuse pattern from sqlite/schema.go)
  - [x] Define `CreateStageTransitionsTableSQL`
  - [x] Define `CreateIndexProjectSQL` for project_id lookups
  - [x] Define `CreateIndexTimeSQL` for time-range queries
  - [x] Schema version starts at v1

- [x] Task 3: Implement MetricsRepository core methods (AC: #1, #5)
  - [x] `NewMetricsRepository(dbPath string) *MetricsRepository` - constructor (no error return, lazy init)
  - [x] `ensureSchema(ctx context.Context, db *sqlx.DB) error` - creates tables if not exist
  - [x] `openDB(ctx context.Context) (*sqlx.DB, error)` - WAL mode connection
  - [x] Database path: `~/.vibe-dash/metrics.db` (get base path from `DirectoryManager.BaseDir()`)

- [x] Task 4: Implement stage transition recording method (AC: #2)
  - [x] `RecordTransition(ctx context.Context, projectID, fromStage, toStage string) error`
  - [x] Generate UUID using `crypto/rand` (see pattern below)
  - [x] Use `time.RFC3339Nano` timestamp format
  - [x] Return `nil` on any error (graceful degradation)

- [x] Task 5: Implement graceful degradation (AC: #3)
  - [x] All public methods catch errors internally
  - [x] Log warnings with `slog.Warn()` including error context
  - [x] Return `nil` or empty results - never propagate errors to callers
  - [x] Test degradation with simulated DB failures

- [x] Task 6: Write comprehensive unit tests
  - [x] `repository_test.go` - Unit tests with temp directories
  - [x] Test schema creation on first access
  - [x] Test transition recording with valid data
  - [x] Test graceful degradation on DB errors (permission denied, etc.)
  - [x] `repository_integration_test.go` - Integration tests with `//go:build integration` tag

## Dev Notes

### Architecture Alignment

This story implements the **Isolation Principle** from Phase 2:
- Separate database (`metrics.db`) - NOT in per-project `state.db`
- Separate package (`internal/adapters/persistence/metrics/`)
- Graceful degradation - core dashboard unaffected by metrics failures
- Event-based architecture ready for Story 16.2 (recorder integration)

### File Structure

```
internal/adapters/persistence/metrics/
├── repository.go          # MetricsRepository implementation
├── schema.go              # Schema SQL constants
├── queries.go             # Query SQL constants
├── helpers.go             # Row struct, UUID generator, helpers
├── repository_test.go     # Unit tests
└── repository_integration_test.go  # Integration tests (build tag)
```

### Required Imports

```go
import (
    "context"
    "crypto/rand"
    "fmt"
    "log/slog"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/mattn/go-sqlite3"
)
```

### Schema SQL (schema.go)

```go
package metrics

const SchemaVersion = 1

const CreateSchemaVersionTableSQL = `
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);`

const CreateStageTransitionsTableSQL = `
CREATE TABLE IF NOT EXISTS stage_transitions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    from_stage TEXT NOT NULL,
    to_stage TEXT NOT NULL,
    transitioned_at TEXT NOT NULL
);`

const CreateIndexProjectSQL = `CREATE INDEX IF NOT EXISTS idx_stage_transitions_project ON stage_transitions(project_id);`
const CreateIndexTimeSQL = `CREATE INDEX IF NOT EXISTS idx_stage_transitions_time ON stage_transitions(transitioned_at);`
```

### Helpers (helpers.go)

```go
package metrics

import (
    "crypto/rand"
    "fmt"
)

// stageTransitionRow is the database row representation for scanning
type stageTransitionRow struct {
    ID             string `db:"id"`
    ProjectID      string `db:"project_id"`
    FromStage      string `db:"from_stage"`
    ToStage        string `db:"to_stage"`
    TransitionedAt string `db:"transitioned_at"`
}

// generateUUID creates a UUID v4 using crypto/rand (no external dependencies)
func generateUUID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    // Set version (4) and variant bits
    b[6] = (b[6] & 0x0f) | 0x40
    b[8] = (b[8] & 0x3f) | 0x80
    return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
```

### Repository Pattern (repository.go)

```go
package metrics

const walConnectionParams = "?_journal_mode=WAL&_busy_timeout=5000"

// MetricsRepository handles metrics persistence with graceful degradation.
// All public methods return nil/empty on errors - metrics failures never crash the dashboard.
type MetricsRepository struct {
    dbPath        string
    schemaCreated bool
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

// ensureSchema creates tables on first access (idempotent)
func (r *MetricsRepository) ensureSchema(ctx context.Context, db *sqlx.DB) error {
    if r.schemaCreated {
        return nil
    }
    // Execute schema creation SQL...
    r.schemaCreated = true
    return nil
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
```

### Queries (queries.go)

```go
package metrics

const transitionColumns = `id, project_id, from_stage, to_stage, transitioned_at`

const insertTransitionSQL = `
INSERT INTO stage_transitions (` + transitionColumns + `)
VALUES (?, ?, ?, ?, ?)`

const selectByProjectSQL = `
SELECT ` + transitionColumns + ` FROM stage_transitions
WHERE project_id = ? ORDER BY transitioned_at DESC`

const selectByTimeRangeSQL = `
SELECT ` + transitionColumns + ` FROM stage_transitions
WHERE transitioned_at >= ? AND transitioned_at <= ?
ORDER BY transitioned_at DESC`
```

### Database Path Resolution

Get the base directory from `DirectoryManager`:

```go
import "github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"

// In main.go or wherever wiring happens:
dm := filesystem.NewDirectoryManager()
metricsDBPath := filepath.Join(dm.BaseDir(), "metrics.db")
metricsRepo := metrics.NewMetricsRepository(metricsDBPath)
```

### Testing Patterns

**Unit Test (repository_test.go):**
```go
func TestRecordTransition_Success(t *testing.T) {
    tmpDir := t.TempDir()
    repo := NewMetricsRepository(filepath.Join(tmpDir, "metrics.db"))

    err := repo.RecordTransition(context.Background(), "proj-123", "plan", "tasks")

    if err != nil {
        t.Errorf("expected nil error, got %v", err)
    }
}

func TestRecordTransition_GracefulDegradation(t *testing.T) {
    // Use invalid path to simulate failure
    repo := NewMetricsRepository("/nonexistent/dir/metrics.db")

    err := repo.RecordTransition(context.Background(), "proj-123", "plan", "tasks")

    // Should return nil (graceful degradation), not error
    if err != nil {
        t.Errorf("expected nil (graceful degradation), got %v", err)
    }
}
```

**Integration Test (repository_integration_test.go):**
```go
//go:build integration

package metrics_test

func TestRecordTransition_Integration(t *testing.T) {
    // Test with real SQLite database
}
```

### NFR Compliance

- **NFR-P2-4:** Database growth < 20MB/year
  - ~500 events/day × 3 projects × 365 days = 547,500 events/year
  - Each row ~100 bytes = ~55MB worst case, but realistic usage ~18,250 events = ~1.8MB/year

- **NFR-P2-7:** Metrics failure does not affect core dashboard
  - All methods return `nil` on error
  - Logging only via `slog.Warn()`
  - No error propagation to callers

### Story 16.2 Preparation

This story prepares for Story 16.2 (Stage Transition Event Recording) by providing:
- `RecordTransition()` method for the event recorder to call
- `selectByProjectSQL` and `selectByTimeRangeSQL` for future stats queries

Add these query methods in Story 16.2 or later:
```go
func (r *MetricsRepository) GetTransitionsByProject(ctx context.Context, projectID string) ([]StageTransition, error)
func (r *MetricsRepository) GetTransitionsByTimeRange(ctx context.Context, from, to time.Time) ([]StageTransition, error)
```

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- All tasks completed per story requirements
- All 11 unit tests pass covering schema creation, transition recording, graceful degradation, WAL mode, UUID generation, and indexes
- Integration tests added with `//go:build integration` tag for concurrent writes and volume testing
- Query constants for Story 16.2 marked with `//nolint:unused` to pass linting
- Full test suite (1351 tests) passes with no regressions
- Linting passes clean

### Code Review Fixes Applied

- **M1 Fixed:** Changed `schemaCreated bool` to `sync.Once` for thread-safe schema initialization (repository.go)
- **M3 Fixed:** Added `//nolint:unused` comment to `stageTransitionRow` struct (helpers.go)
- **H2 Fixed:** Updated `project-context.md` to correctly document metrics.db schema (no FK to projects)
- Tests updated to work with new sync.Once pattern
- All 1351 tests pass after fixes

### File List

- `internal/adapters/persistence/metrics/repository.go` - MetricsRepository implementation
- `internal/adapters/persistence/metrics/schema.go` - Schema SQL constants (SchemaVersion=1)
- `internal/adapters/persistence/metrics/queries.go` - Query SQL constants
- `internal/adapters/persistence/metrics/helpers.go` - Row struct and UUID generator
- `internal/adapters/persistence/metrics/repository_test.go` - Unit tests (11 tests)
- `internal/adapters/persistence/metrics/repository_integration_test.go` - Integration tests
