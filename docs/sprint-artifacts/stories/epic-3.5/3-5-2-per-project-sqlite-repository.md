# Story 3.5.2: Per-Project SQLite Repository

Status: done

## Story

As a developer,
I want each project to have its own SQLite database,
So that project data is isolated and can be backed up independently.

## Acceptance Criteria

1. **AC1: Database location** - Given project directory exists at `~/.vibe-dash/<project>/`, when repository is created for project, then `state.db` is created at `~/.vibe-dash/<project>/state.db`.

2. **AC2: Data persistence** - Given project repository, when saving project state, then data is written to project-specific `state.db` (not centralized database).

3. **AC3: WAL mode enabled** - Given project `state.db`, when opening connection, then WAL mode is enabled via DSN: `<dbPath>?_journal_mode=WAL&_busy_timeout=5000`.

4. **AC4: Corruption handling** - Given corrupted `state.db`, when corruption is detected, then recovery suggestion includes project-specific path and instructs user to re-add project.

5. **AC5: Schema compatibility** - Given new project database, when schema is initialized, then same schema version (v2 from `schema.go:4`) and structure as existing repository is used.

6. **AC6: Lazy connection** - Given multiple project databases, when accessing a project's data, then connection is opened on-demand and closed after operation (no persistent pool).

7. **AC7: Project isolation** - Given two projects tracked, when modifying one project's state, then the other project's database is not affected.

8. **AC8: Integration with DirectoryManager** - Given `DirectoryManager.EnsureProjectDir()` returns project directory path, when creating repository, then database is created in that directory and `.project-path` marker file existence is verified.

9. **AC9: Thread-safety** - Given concurrent access from TUI refresh and file watcher goroutines, when operations execute simultaneously, then no data races occur (each operation opens/closes its own connection via WAL mode).

## Tasks / Subtasks

- [x] Task 1: Review existing interface compatibility (AC: 1, 2, 5)
  - [x] Subtask 1.1: Confirm `ports.ProjectRepository` interface (lines 22-59) works for per-project context
  - [x] Subtask 1.2: Existing interface works as-is since it operates on single project at a time
  - [x] Subtask 1.3: No interface modification needed

- [x] Task 2: Create `NewProjectRepository` constructor (AC: 1, 3, 6, 8, 9)
  - [x] Subtask 2.1: Create `internal/adapters/persistence/sqlite/project_repository.go`
  - [x] Subtask 2.2: Define connection constant: `const walConnectionParams = "?_journal_mode=WAL&_busy_timeout=5000"`
  - [x] Subtask 2.3: Implement constructor with validation:
    - Validate `projectDir` exists (`os.Stat`)
    - Verify `.project-path` marker file exists (created by DirectoryManager)
    - Construct `state.db` path as `filepath.Join(projectDir, "state.db")`
    - Initialize schema (fail-fast)
    - Return `domain.ErrPathNotAccessible` on validation failures
  - [x] Subtask 2.4: Struct fields: `dbPath string`, `projectDir string` (no mutable state = thread-safe)

- [x] Task 3: Copy helper functions from repository.go (AC: 2, 5, 6)
  - [x] Subtask 3.1: Copy private helpers (cannot import, must duplicate):
    - `wrapDBError()` (lines 22-38) - modify to use project-specific path
    - `projectRow` struct (lines 306-322)
    - `rowToProject()` (lines 324-362)
    - `nullString()` (lines 364-370)
    - `boolToInt()` (lines 372-378)
    - `stateToString()` (lines 380-390)
  - [x] Subtask 3.2: Update `wrapDBError()` recovery message: `"delete %s and re-add project via 'vibe add <path>'"`

- [x] Task 4: Implement `openDB(ctx)` method (AC: 3, 6, 9)
  - [x] Subtask 4.1: Use lazy connection pattern with `walConnectionParams` constant
  - [x] Subtask 4.2: Pattern: `sqlx.ConnectContext(ctx, "sqlite3", r.dbPath+walConnectionParams)`
  - [x] Subtask 4.3: Wrap errors with `wrapDBError()` for corruption detection
  - [x] Subtask 4.4: Thread-safety: Each call creates new connection (no shared state)

- [x] Task 5: Implement interface methods (AC: 2, 5, 6, 7)
  - [x] Subtask 5.1: All methods follow pattern: `db, err := r.openDB(ctx); defer db.Close()`
  - [x] Subtask 5.2: Implement using existing SQL from `queries.go` (already exported)
  - [x] Subtask 5.3: `FindAll/FindActive/FindHibernated` return 0-1 results (single project per DB)

- [x] Task 6: Reuse schema and migrations (AC: 5)
  - [x] Subtask 6.1: Use `RunMigrations()` from `migrations.go:35-59` (already exported)
  - [x] Subtask 6.2: Schema version 2 (verify via `SchemaVersion` constant from `schema.go:4`)
  - [x] Subtask 6.3: Indexes on `path` and `state` kept for compatibility (minor overhead for 0-1 rows)

- [x] Task 7: Write unit tests (AC: 1-9)
  - [x] Subtask 7.1: Test `NewProjectRepository` creates `state.db` in correct location
  - [x] Subtask 7.2: Test `NewProjectRepository` fails if `.project-path` marker missing
  - [x] Subtask 7.3: Test `Save`/`FindByID`/`FindByPath`/`Delete`/`UpdateState` operations
  - [x] Subtask 7.4: Test WAL mode: `PRAGMA journal_mode;` returns `wal`
  - [x] Subtask 7.5: Test corruption error messages include project path
  - [x] Subtask 7.6: Test schema version: `SELECT version FROM schema_version` = 2
  - [x] Subtask 7.7: Test concurrent access (multiple goroutines)

- [x] Task 8: Write integration tests (AC: 1, 2, 7, 8)
  - [x] Subtask 8.1: Create temp directories with `.project-path` marker
  - [x] Subtask 8.2: Test full lifecycle: create repo → save → find → delete
  - [x] Subtask 8.3: Test two projects have isolated databases (modify one, verify other unchanged)
  - [x] Subtask 8.4: Test integration with real `DirectoryManager.EnsureProjectDir()`

## Dev Notes

### Architecture Alignment

**PRD Specification (lines 597-604):**
```
~/.vibe-dash/
  ├── config.yaml                 # Master index
  ├── api-service/
  │   ├── config.yaml             # Project-specific settings
  │   └── state.db                # Per-project SQLite database
```

**Architecture Specification (lines 296-302):**
> SQLite Concurrency: WAL Mode
> - Enable Write-Ahead Logging for concurrent read access
> - Single writer, multiple readers pattern
> - Connection opened on-demand, closed when operation completes

### Hexagonal Architecture

```
internal/core/ports/repository.go          → Existing interface (reuse as-is)
internal/adapters/persistence/sqlite/project_repository.go → New implementation
```

**CRITICAL: Core never imports adapters.** New `ProjectRepository` implements `ports.ProjectRepository`.

### Existing Code Reference

| Component | Location | Action |
|-----------|----------|--------|
| Interface | `ports/repository.go:22-59` | Implement as-is |
| Schema | `schema.go:4` (SchemaVersion=2) | Reuse |
| Queries | `queries.go` | Reuse all SQL constants |
| Migrations | `migrations.go:35-59` | Reuse `RunMigrations()` |
| Helpers | `repository.go:305-390` | Copy (private functions) |
| Error | `repository.go:19-38` | Copy and modify for project path |

### Connection Pattern

```go
const walConnectionParams = "?_journal_mode=WAL&_busy_timeout=5000"

func (r *ProjectRepository) openDB(ctx context.Context) (*sqlx.DB, error) {
    return sqlx.ConnectContext(ctx, "sqlite3", r.dbPath+walConnectionParams)
}

// All methods follow this pattern:
func (r *ProjectRepository) Method(ctx context.Context, ...) error {
    db, err := r.openDB(ctx)
    if err != nil {
        return err
    }
    defer db.Close()  // CRITICAL: Always close

    // ... operation ...
}
```

### Thread Safety

Per project-context.md and Architecture, thread safety is achieved through:
1. **Lazy connections**: Each operation opens its own connection
2. **WAL mode**: Allows concurrent reads while writing
3. **Busy timeout**: 5000ms wait for locks
4. **No shared state**: `ProjectRepository` struct has no mutable fields

### Constructor Implementation

```go
func NewProjectRepository(projectDir string) (*ProjectRepository, error) {
    // Validate projectDir exists
    if _, err := os.Stat(projectDir); os.IsNotExist(err) {
        return nil, fmt.Errorf("%w: project directory does not exist: %s",
            domain.ErrPathNotAccessible, projectDir)
    }

    // Verify DirectoryManager created this directory (marker file check)
    markerPath := filepath.Join(projectDir, ".project-path")
    if _, err := os.Stat(markerPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("%w: directory not created by DirectoryManager (missing .project-path): %s",
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
```

### Difference from Existing Repository

| Aspect | `SQLiteRepository` | `ProjectRepository` |
|--------|-------------------|---------------------|
| Database | `~/.vibe-dash/projects.db` | `~/.vibe-dash/<project>/state.db` |
| Projects per DB | Many | One (0-1 rows) |
| Constructor | `NewSQLiteRepository(dbPath)` | `NewProjectRepository(projectDir)` |
| Validation | Creates directory if missing | Requires existing dir + `.project-path` marker |
| Used by | Legacy (will deprecate) | `RepositoryCoordinator` (3.5.5) |

### Error Handling

```go
// Domain errors to use:
domain.ErrProjectNotFound     // Project not in DB
domain.ErrPathNotAccessible   // Directory/file access issues
ErrDatabaseCorrupted          // SQLite corruption (from repository.go:20)

// Error wrapping pattern:
return fmt.Errorf("failed to save project: %w", err)

// Corruption recovery message:
return fmt.Errorf("%w: %v. Recovery: delete %s and re-add project via 'vibe add <path>'",
    ErrDatabaseCorrupted, err, r.dbPath)
```

### Performance

- Target: Operations complete in <10ms to support NFR-P1 (<100ms render)
- Single-project DBs with 0-1 rows are extremely fast
- Lazy open/close adds ~1-2ms overhead per operation (acceptable)
- RepositoryCoordinator aggregates across DBs (Story 3.5.5 handles latency)

### Dependencies

**Depends on (COMPLETED):**
- Story 3.5.1: DirectoryManager - Provides `EnsureProjectDir()` and `.project-path` marker

**Required by:**
- Story 3.5.5: Repository Coordinator - Aggregates multiple `ProjectRepository` instances
- Story 3.5.6: Update CLI Commands - Uses `ProjectRepository` via Coordinator

### Previous Story Learnings (3.5.1)

From Story 3.5.1 (Directory Manager):
- `FilesystemDirectoryManager.EnsureProjectDir()` returns full path like `/home/user/.vibe-dash/api-service`
- `.project-path` marker file exists in each project directory (contains canonical project path)
- Use `context.Context` for cancellation support in all methods
- Table-driven tests are required per Architecture

### Files to Create

| File | Purpose |
|------|---------|
| `project_repository.go` | Per-project repository implementation |
| `project_repository_test.go` | Unit tests (table-driven) |

### Manual Testing

After implementation:

1. **Verify database creation:**
   ```bash
   mkdir -p ~/.vibe-dash/test-project
   echo "/path/to/test-project" > ~/.vibe-dash/test-project/.project-path
   # Repository should create state.db on construction
   ```

2. **Verify SQLite configuration:**
   ```bash
   sqlite3 ~/.vibe-dash/test-project/state.db "PRAGMA journal_mode;"
   # Output: wal

   sqlite3 ~/.vibe-dash/test-project/state.db "SELECT version FROM schema_version;"
   # Output: 2
   ```

3. **Verify data isolation:**
   ```bash
   ls ~/.vibe-dash/project-a/state.db
   ls ~/.vibe-dash/project-b/state.db
   # Modifying one doesn't affect other
   ```

## Dev Agent Record

### Context Reference

- Epic 3.5: Storage Structure Alignment
- Story Dependencies: Depends on 3.5.1 (Directory Manager), required by 3.5.5 (Repository Coordinator)
- PRD Reference: Lines 597-604 (Per-project SQLite structure)
- Architecture Reference: Lines 296-302 (WAL mode, lazy connections)
- Project Context: Lines 99-104 (SQLite Rules)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None

### Completion Notes List

1. **Interface Compatibility Verified**: The existing `ports.ProjectRepository` interface works as-is for per-project context since all methods operate on a single project at a time. No modifications needed.

2. **Constructor Implementation**: `NewProjectRepository(projectDir string)` validates:
   - Directory existence via `os.Stat`
   - `.project-path` marker file presence (created by DirectoryManager)
   - Constructs `state.db` path and initializes schema with fail-fast behavior

3. **Helper Functions Duplicated**: Private helpers from `repository.go` were duplicated with `ForProject` suffix to avoid name collisions:
   - `wrapDBErrorForProject()` - updated recovery message for project-specific path
   - `projectRowForProject` struct
   - `rowToProjectForProject()`, `nullStringForProject()`, `boolToIntForProject()`, `stateToStringForProject()`

4. **Lazy Connection Pattern**: Each operation opens/closes its own connection using WAL mode (`?_journal_mode=WAL&_busy_timeout=5000`). No persistent connection pool = thread-safe by design.

5. **Schema Reuse**: Uses existing `RunMigrations()` and schema (version 2). Same table structure as `SQLiteRepository` but per-project isolation.

6. **Test Coverage**:
   - Unit tests: constructor validation, WAL mode, schema version, CRUD operations, concurrent access (10 goroutines × 3 operations each)
   - Integration tests: full lifecycle, project isolation, DirectoryManager integration

7. **All Acceptance Criteria Met**:
   - AC1: Database at `~/.vibe-dash/<project>/state.db` ✓
   - AC2: Project-specific data persistence ✓
   - AC3: WAL mode enabled via DSN ✓
   - AC4: Corruption handling with project-specific recovery message ✓
   - AC5: Schema version 2 compatibility ✓
   - AC6: Lazy connection pattern ✓
   - AC7: Project isolation verified with integration tests ✓
   - AC8: Integration with DirectoryManager + `.project-path` marker ✓
   - AC9: Thread-safety via lazy connections + WAL mode ✓

### File List

| File | Action | Description |
|------|--------|-------------|
| `internal/adapters/persistence/sqlite/project_repository.go` | Created | Per-project repository implementation |
| `internal/adapters/persistence/sqlite/project_repository_test.go` | Created | Unit tests (table-driven) |
| `internal/adapters/persistence/sqlite/project_repository_integration_test.go` | Created | Integration tests with `//go:build integration` tag |
