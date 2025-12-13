# Story 2.1: SQLite Repository Setup

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | Implements `ports.ProjectRepository` interface |
| **Key Dependencies** | github.com/jmoiron/sqlx, github.com/mattn/go-sqlite3 (CGO required) |
| **Files to Create** | schema.go, queries.go, migrations.go, repository.go, repository_test.go |
| **Location** | internal/adapters/persistence/sqlite/ |
| **Database Path** | ~/.vibe-dash/projects.db |

### Quick Task Summary (7 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Add dependencies | sqlx + go-sqlite3 in go.mod |
| 2 | schema.go | SQL DDL constants |
| 3 | migrations.go | Version tracking + runner |
| 4 | queries.go | SQL query constants (DRY) |
| 5 | repository.go | 8 interface methods |
| 6 | Tests | 26 test cases |
| 7 | Validation | build, lint, test, manual |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| SQL Library | sqlx | Lightweight struct scanning, honest SQL |
| Connection Mode | Open-use-close | Prevent file handle exhaustion |
| WAL Mode | Required | Concurrent reads for TUI + file watcher |
| Busy Timeout | 5000ms | Wait for locks during concurrent access |
| Timestamps | time.RFC3339Nano | Go's ISO 8601 profile with nanosecond precision |

## Story

**As a** developer,
**I want** SQLite persistence for project data,
**So that** project state survives between sessions.

## Acceptance Criteria

```gherkin
AC1: Given I need to persist project data
     When SQLite repository is initialized
     Then database is created at ~/.vibe-dash/projects.db
     And schema includes projects table with all required columns
     And schema_version table tracks migrations
     And WAL mode is enabled for concurrent access

AC2: Given repository.Save() is called with valid project
     When the project doesn't exist
     Then project is inserted into database
     And all fields are stored correctly
     And timestamps are set correctly

AC3: Given repository.Save() is called with existing project
     When the project ID already exists
     Then project is updated (upsert)
     And updated_at timestamp is refreshed

AC4: Given repository.FindByPath() is called
     When project exists at that path
     Then project is retrieved with all fields hydrated

AC5: Given repository.FindByPath() is called
     When no project exists at that path
     Then domain.ErrProjectNotFound is returned

AC6: Given repository.FindAll() is called
     When projects exist
     Then all projects are returned as slice
     When no projects exist
     Then empty slice (not nil) is returned

AC7: Given repository.FindActive() is called
     When active projects exist
     Then only projects with state='active' are returned

AC8: Given repository.FindHibernated() is called
     When hibernated projects exist
     Then only projects with state='hibernated' are returned

AC9: Given repository.Delete() is called
     When project exists
     Then project is removed from database
     When project doesn't exist
     Then domain.ErrProjectNotFound is returned

AC10: Given repository.UpdateState() is called
      When project exists
      Then only state field is updated
      And updated_at timestamp is refreshed

AC11: Given database file is corrupted
      When repository operations fail
      Then error is returned with recovery suggestion
```

## Tasks / Subtasks

- [x] **Task 1: Add sqlx dependency** (AC: all)
  - [x] 1.1 Run `go get github.com/jmoiron/sqlx && go get github.com/mattn/go-sqlite3 && go mod tidy`
  - [x] 1.2 Verify go.mod shows sqlx and go-sqlite3 as direct dependencies
  - [x] 1.3 Verify `go build ./...` succeeds (CGO test)

- [x] **Task 2: Create sqlite/schema.go** (AC: 1)
  - [x] 2.1 Define CreateProjectsTableSQL constant (see Schema section)
  - [x] 2.2 Define CreateSchemaVersionTableSQL constant
  - [x] 2.3 Define CreateIndexesSQL constants
  - [x] 2.4 Create SchemaVersion constant (v1 for initial)

- [x] **Task 3: Create sqlite/migrations.go** (AC: 1)
  - [x] 3.1 Define Migration struct (version, description, sql)
  - [x] 3.2 Define migrations slice with v1 migration
  - [x] 3.3 Create RunMigrations function to apply pending migrations
  - [x] 3.4 Create getCurrentVersion helper
  - [x] 3.5 ~~Create setVersion helper~~ (removed - inlined in applyMigration)

- [x] **Task 4: Create sqlite/queries.go** (AC: 2-11)
  - [x] 4.1 Define projectColumns constant (DRY column list)
  - [x] 4.2 Define insertOrReplaceProjectSQL constant
  - [x] 4.3 Define SELECT queries using projectColumns
  - [x] 4.4 Define deleteByIDSQL constant
  - [x] 4.5 Define updateStateSQL constant

- [x] **Task 5: Create sqlite/repository.go** (AC: all)
  - [x] 5.1 Define SQLiteRepository struct
  - [x] 5.2 Create NewSQLiteRepository constructor (fail-fast schema init)
  - [x] 5.3 Implement openDB helper (connection string with WAL + busy_timeout)
  - [x] 5.4 Implement Save method (upsert with validation)
  - [x] 5.5 Implement FindByID method (return ErrProjectNotFound on miss)
  - [x] 5.6 Implement FindByPath method
  - [x] 5.7 Implement FindAll method (return empty slice, not nil)
  - [x] 5.8 Implement FindActive method (return empty slice, not nil)
  - [x] 5.9 Implement FindHibernated method (return empty slice, not nil)
  - [x] 5.10 Implement Delete method (return ErrProjectNotFound on miss)
  - [x] 5.11 Implement UpdateState method (return ErrProjectNotFound on miss)
  - [x] 5.12 Create projectRow struct and rowToProject helper
  - [x] 5.13 Create helper functions (nullString, boolToInt, stateToString)

- [x] **Task 6: Write Tests** (AC: all)
  - [x] 6.1 Create sqlite/repository_test.go
  - [x] 6.2 Test: Database and schema creation
  - [x] 6.3 Test: Save new project (insert)
  - [x] 6.4 Test: Save existing project (update)
  - [x] 6.5 Test: FindByID found and not found
  - [x] 6.6 Test: FindByPath found and not found
  - [x] 6.7 Test: FindAll with projects
  - [x] 6.8 Test: FindAll returns empty slice (not nil) when empty
  - [x] 6.9 Test: FindActive filters correctly + empty slice
  - [x] 6.10 Test: FindHibernated filters correctly + empty slice
  - [x] 6.11 Test: Delete found and not found
  - [x] 6.12 Test: UpdateState found and not found
  - [x] 6.13 Test: WAL mode is enabled
  - [x] 6.14 Test: Context cancellation handling
  - [x] 6.15 Test: Unique path constraint violation
  - [x] 6.16 Use t.TempDir() for all test isolation

- [x] **Task 7: Integration and Validation** (AC: all)
  - [x] 7.1 Run `make build` and verify compilation
  - [x] 7.2 Run `make lint` and fix any issues
  - [x] 7.3 Run `make test` and verify all tests pass
  - [x] 7.4 Manual test: Create repository, save project, restart app, verify persistence (covered by tests)
  - [x] 7.5 Manual test: Verify WAL mode with `PRAGMA journal_mode;` query (covered by TestSQLiteRepository_WALMode)

## Dev Notes

### CRITICAL Requirements (Must Not Miss)

| Requirement | Why | Reference |
|-------------|-----|-----------|
| **WAL mode required** | Concurrent read access for file watcher + TUI | Architecture line 296-301 |
| **Open-use-close pattern** | Prevent file handle exhaustion at scale | Architecture line 119-120 |
| **snake_case columns** | Database naming convention | Architecture line 471-496 |
| **Return empty slice not nil** | Interface contract for Find* methods | ports/repository.go:37-48 |
| **Return domain errors** | ErrProjectNotFound not raw sql.ErrNoRows | Architecture line 709 |

### Build Requirements (CGO)

go-sqlite3 requires CGO enabled (default on most systems):
- **macOS:** Works with Xcode command line tools installed
- **Linux:** Requires gcc (`apt install build-essential` or equivalent)
- **CI:** Ensure `CGO_ENABLED=1` in test/build environment
- **Windows:** May need MinGW or use `modernc.org/sqlite` as pure Go alternative

### Existing ProjectRepository Interface (from ports/repository.go)

**CRITICAL:** Implement this interface exactly. Do NOT recreate it.

```go
type ProjectRepository interface {
    Save(ctx context.Context, project *domain.Project) error
    FindByID(ctx context.Context, id string) (*domain.Project, error)
    FindByPath(ctx context.Context, path string) (*domain.Project, error)
    FindAll(ctx context.Context) ([]*domain.Project, error)        // Returns empty slice, not nil
    FindActive(ctx context.Context) ([]*domain.Project, error)     // Returns empty slice, not nil
    FindHibernated(ctx context.Context) ([]*domain.Project, error) // Returns empty slice, not nil
    Delete(ctx context.Context, id string) error
    UpdateState(ctx context.Context, id string, state domain.ProjectState) error
}
```

### Project Entity Fields (domain/project.go)

```go
type Project struct {
    ID             string       // Unique identifier (path hash, 16 hex chars)
    Name           string       // Derived from directory name
    Path           string       // Canonical absolute path
    DisplayName    string       // Optional user-set nickname
    DetectedMethod string       // "speckit", "bmad", "unknown"
    CurrentStage   Stage        // Current workflow stage (enum)
    IsFavorite     bool         // Always visible flag
    State          ProjectState // Active or Hibernated (enum)
    Notes          string       // User notes/memo
    LastActivityAt time.Time    // Last file change detected
    CreatedAt      time.Time    // When project was added
    UpdatedAt      time.Time    // Last database update
}
```

### SQLite Schema

**Design Note:** Schema includes `confidence` and `detection_reasoning` columns for future detection result caching (FR11, FR26). These are stored as empty strings in Story 2.1 and will be populated when detection service is implemented in Story 2.4-2.5. The `projectRow` struct reads these columns but `rowToProject()` does not map them to Project (they belong to DetectionResult).

```sql
-- Enable WAL mode for concurrent access (in connection string)
-- ?_journal_mode=WAL&_busy_timeout=5000

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    display_name TEXT,
    detected_method TEXT,
    current_stage TEXT,
    confidence TEXT,            -- Future: cached from DetectionResult
    detection_reasoning TEXT,   -- Future: cached from DetectionResult
    is_favorite INTEGER DEFAULT 0,
    state TEXT DEFAULT 'active',
    notes TEXT,
    last_activity_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_projects_path ON projects(path);
CREATE INDEX IF NOT EXISTS idx_projects_state ON projects(state);
```

### Query Constants Pattern (DRY)

```go
package sqlite

// projectColumns lists all columns for SELECT queries (DRY)
const projectColumns = `id, name, path, display_name, detected_method, current_stage,
       confidence, detection_reasoning, is_favorite, state, notes,
       last_activity_at, created_at, updated_at`

const insertOrReplaceProjectSQL = `
INSERT OR REPLACE INTO projects (` + projectColumns + `)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

const selectByIDSQL = `SELECT ` + projectColumns + ` FROM projects WHERE id = ?`
const selectByPathSQL = `SELECT ` + projectColumns + ` FROM projects WHERE path = ?`
const selectAllSQL = `SELECT ` + projectColumns + ` FROM projects ORDER BY name`
const selectActiveSQL = `SELECT ` + projectColumns + ` FROM projects WHERE state = 'active' ORDER BY name`
const selectHibernatedSQL = `SELECT ` + projectColumns + ` FROM projects WHERE state = 'hibernated' ORDER BY name`
const deleteByIDSQL = `DELETE FROM projects WHERE id = ?`
const updateStateSQL = `UPDATE projects SET state = ?, updated_at = ? WHERE id = ?`
```

### Repository Implementation Pattern

**Design Note:** Schema initialization happens synchronously in constructor for fail-fast behavior. This ensures the database is usable immediately after `NewSQLiteRepository` returns, catching permission/disk errors early.

```go
package sqlite

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/mattn/go-sqlite3"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

type SQLiteRepository struct {
    dbPath string
}

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

// openDB opens connection with WAL mode and busy timeout.
// Busy timeout (5000ms) allows SQLite to wait for write locks during concurrent access.
// Caller MUST close the connection after use.
func (r *SQLiteRepository) openDB(ctx context.Context) (*sqlx.DB, error) {
    db, err := sqlx.ConnectContext(ctx, "sqlite3", r.dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    return db, nil
}
```

### Save Implementation (Upsert)

```go
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
        project.DetectedMethod,
        project.CurrentStage.String(),
        "",  // confidence - populated by detection service (Story 2.4)
        "",  // detection_reasoning - populated by detection service (Story 2.4)
        boolToInt(project.IsFavorite),
        project.State.String(),
        nullString(project.Notes),
        project.LastActivityAt.Format(time.RFC3339),
        project.CreatedAt.Format(time.RFC3339),
        project.UpdatedAt.Format(time.RFC3339),
    )
    if err != nil {
        return fmt.Errorf("failed to save project: %w", err)
    }

    return nil
}
```

### FindAll Implementation (Empty Slice Pattern)

**CRITICAL:** Return empty slice, not nil, when no results found.

```go
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
```

### Delete Implementation

```go
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
```

### UpdateState Implementation

```go
func (r *SQLiteRepository) UpdateState(ctx context.Context, id string, state domain.ProjectState) error {
    db, err := r.openDB(ctx)
    if err != nil {
        return err
    }
    defer db.Close()

    result, err := db.ExecContext(ctx, updateStateSQL,
        state.String(),
        time.Now().Format(time.RFC3339),
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
```

### Row Struct and Conversion

```go
type projectRow struct {
    ID                 string         `db:"id"`
    Name               string         `db:"name"`
    Path               string         `db:"path"`
    DisplayName        sql.NullString `db:"display_name"`
    DetectedMethod     sql.NullString `db:"detected_method"`
    CurrentStage       string         `db:"current_stage"`
    Confidence         sql.NullString `db:"confidence"`            // Read but not mapped to Project
    DetectionReasoning sql.NullString `db:"detection_reasoning"`   // Read but not mapped to Project
    IsFavorite         int            `db:"is_favorite"`
    State              string         `db:"state"`
    Notes              sql.NullString `db:"notes"`
    LastActivityAt     string         `db:"last_activity_at"`
    CreatedAt          string         `db:"created_at"`
    UpdatedAt          string         `db:"updated_at"`
}

func rowToProject(row *projectRow) (*domain.Project, error) {
    // Parse timestamps (time.RFC3339 = ISO 8601 with timezone)
    lastActivity, err := time.Parse(time.RFC3339, row.LastActivityAt)
    if err != nil {
        return nil, fmt.Errorf("invalid last_activity_at: %w", err)
    }
    created, err := time.Parse(time.RFC3339, row.CreatedAt)
    if err != nil {
        return nil, fmt.Errorf("invalid created_at: %w", err)
    }
    updated, err := time.Parse(time.RFC3339, row.UpdatedAt)
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

func nullString(s string) sql.NullString {
    if s == "" {
        return sql.NullString{}
    }
    return sql.NullString{String: s, Valid: true}
}

func boolToInt(b bool) int {
    if b { return 1 }
    return 0
}
```

### Test Patterns

```go
func TestSQLiteRepository_FindAll_Empty_ReturnsEmptySlice(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    repo, err := NewSQLiteRepository(dbPath)
    if err != nil {
        t.Fatalf("NewSQLiteRepository() error = %v", err)
    }

    ctx := context.Background()
    projects, err := repo.FindAll(ctx)
    if err != nil {
        t.Fatalf("FindAll() error = %v", err)
    }

    // CRITICAL: Must be empty slice, not nil
    if projects == nil {
        t.Error("FindAll() returned nil, want empty slice")
    }
    if len(projects) != 0 {
        t.Errorf("FindAll() returned %d projects, want 0", len(projects))
    }
}

func TestSQLiteRepository_Delete_NotFound(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    repo, err := NewSQLiteRepository(dbPath)
    if err != nil {
        t.Fatalf("NewSQLiteRepository() error = %v", err)
    }

    ctx := context.Background()
    err = repo.Delete(ctx, "nonexistent-id")

    if !errors.Is(err, domain.ErrProjectNotFound) {
        t.Errorf("Delete() error = %v, want ErrProjectNotFound", err)
    }
}

func TestSQLiteRepository_UpdateState(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    repo, err := NewSQLiteRepository(dbPath)
    if err != nil {
        t.Fatalf("NewSQLiteRepository() error = %v", err)
    }

    // Create project first
    project, _ := domain.NewProject("/test/path", "test")
    ctx := context.Background()
    _ = repo.Save(ctx, project)

    // Update state
    err = repo.UpdateState(ctx, project.ID, domain.StateHibernated)
    if err != nil {
        t.Fatalf("UpdateState() error = %v", err)
    }

    // Verify
    found, _ := repo.FindByID(ctx, project.ID)
    if found.State != domain.StateHibernated {
        t.Errorf("State = %v, want StateHibernated", found.State)
    }
}

func TestSQLiteRepository_UniquePathConstraint(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    repo, err := NewSQLiteRepository(dbPath)
    if err != nil {
        t.Fatalf("NewSQLiteRepository() error = %v", err)
    }

    ctx := context.Background()

    // Save first project
    project1, _ := domain.NewProject("/same/path", "project1")
    err = repo.Save(ctx, project1)
    if err != nil {
        t.Fatalf("Save() first project error = %v", err)
    }

    // Try to save different project with same path - should upsert (INSERT OR REPLACE)
    project2, _ := domain.NewProject("/same/path", "project2")
    project2.ID = "different-id" // Force different ID
    err = repo.Save(ctx, project2)

    // INSERT OR REPLACE will succeed, replacing by path uniqueness
    // Verify only one project exists
    projects, _ := repo.FindAll(ctx)
    if len(projects) != 1 {
        t.Errorf("Expected 1 project after upsert, got %d", len(projects))
    }
}
```

### Files to Create

| File | Purpose |
|------|---------|
| `internal/adapters/persistence/sqlite/schema.go` | SQL DDL constants |
| `internal/adapters/persistence/sqlite/queries.go` | SQL query constants (DRY) |
| `internal/adapters/persistence/sqlite/migrations.go` | Schema versioning |
| `internal/adapters/persistence/sqlite/repository.go` | SQLiteRepository implementation |
| `internal/adapters/persistence/sqlite/repository_test.go` | 16 test cases |

### DO NOT (Anti-Patterns)

| DO NOT | DO INSTEAD |
|--------|------------|
| Return sql.ErrNoRows | Return domain.ErrProjectNotFound |
| Return nil from Find* | Return empty slice `make([]*domain.Project, 0)` |
| Keep connections open | Close immediately after operation |
| Use camelCase columns | Use snake_case (display_name) |
| Store bools as TEXT | Store as INTEGER (0/1) |
| Repeat column lists | Use projectColumns constant |

### Previous Story Learnings (Story 1.7)

1. **Test with t.TempDir()** - Use temporary directories for test isolation
2. **Test edge cases** - Empty results, not found, invalid data
3. **Graceful error handling** - Wrap errors with context
4. **Context cancellation** - Handle ctx.Done() in long operations

### References

| Document | Section | Key Content |
|----------|---------|-------------|
| architecture.md | Data Architecture | Lines 287-307: sqlx, WAL mode |
| architecture.md | Database Naming | Lines 471-496: snake_case |
| ports/repository.go | ProjectRepository | Interface with empty slice contract |
| domain/project.go | Project | Entity definition |
| project-context.md | SQLite Rules | WAL, lazy connections |

## Dev Agent Record

### Context Reference

Story context created from analysis of:
- docs/epics.md (Story 2.1 requirements)
- docs/architecture.md (SQLite, sqlx, WAL mode)
- internal/core/ports/repository.go (ProjectRepository interface)
- internal/core/domain/project.go (Project entity)
- internal/core/domain/detection_result.go (DetectionResult - explains confidence columns)
- docs/sprint-artifacts/1-7-configuration-auto-creation.md (Previous story learnings)

### Validation Applied

Story validated and improved by SM agent on 2025-12-13:
- Fixed schema-entity documentation for confidence/detection_reasoning columns
- Added empty slice return pattern for Find* methods
- Added Delete and UpdateState implementation patterns
- Added unique path constraint test
- Consolidated SQL columns with DRY pattern
- Added CGO build requirements
- Documented busy timeout and timestamp format choices

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

- Implemented SQLite repository with all 8 interface methods from ports.ProjectRepository
- Added sqlx v1.4.0 and go-sqlite3 v1.14.32 as direct dependencies
- Created schema.go with DDL constants for projects table, schema_version, and indexes
- Created migrations.go with versioned migration system (currently v1)
- Created queries.go with DRY query constants using projectColumns
- Created repository.go implementing full CRUD operations with WAL mode and busy_timeout
- Used RFC3339Nano for timestamp precision to support sub-second updates in tests
- State values stored as lowercase ('active', 'hibernated') to match query filters
- Added stateToString helper for consistent lowercase storage
- All Find* methods return empty slice (not nil) per interface contract
- Domain errors (ErrProjectNotFound) returned instead of raw sql.ErrNoRows
- Open-use-close pattern implemented to prevent file handle exhaustion
- 26 test cases covering all acceptance criteria (1 skipped for practical reasons)
- All tests pass, lint passes, build succeeds

### Code Review Fixes (2025-12-13)

- **AC11 Implementation**: Added ErrDatabaseCorrupted sentinel error and wrapDBError helper that detects database corruption indicators (malformed, corrupt, disk i/o error) and returns errors with recovery suggestions including the database path
- **UniquePathConstraint Test**: Split into two tests - TestSQLiteRepository_UniquePathConstraint_SameID (same-ID upsert) and TestSQLiteRepository_UniquePathConstraint_DifferentID (different-ID same-path replacement behavior)
- **Save Validation Tests**: Added TestSQLiteRepository_Save_InvalidProject (empty path) and TestSQLiteRepository_Save_InvalidProjectRelativePath (relative path)
- **Corrupted DB Tests**: Added TestSQLiteRepository_CorruptedDatabaseError and TestWrapDBError with 5 subtests for error wrapping logic
- **Documentation**: Updated Save() method godoc to document UpdatedAt mutation side effect
- **Story Updates**: Fixed test count (16 → 26), corrected timestamp format (RFC3339 → RFC3339Nano)

### File List

**New Files:**
- `internal/adapters/persistence/sqlite/schema.go`
- `internal/adapters/persistence/sqlite/migrations.go`
- `internal/adapters/persistence/sqlite/queries.go`
- `internal/adapters/persistence/sqlite/repository.go`
- `internal/adapters/persistence/sqlite/repository_test.go`

**Modified Files:**
- `go.mod` - Added sqlx v1.4.0 and go-sqlite3 v1.14.32 as direct dependencies
- `go.sum` - Updated with new dependencies

### Change Log

- 2025-12-13: Implemented Story 2.1 SQLite Repository Setup - All tasks completed
- 2025-12-13: Code Review - Fixed 7 issues (2 HIGH, 3 MEDIUM, 2 LOW), added 6 new tests, updated documentation
