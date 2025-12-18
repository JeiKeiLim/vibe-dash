# Story 3.5.5: Repository Coordinator

Status: done

## Story

As a developer,
I want a unified interface over multiple per-project repositories,
So that the service layer doesn't need to change.

## Acceptance Criteria

1. **AC1: ProjectRepository interface compliance** - Given `RepositoryCoordinator` struct, when used by service layer, then it implements `ports.ProjectRepository` interface with zero service layer code changes.

2. **AC2: FindAll aggregation** - Given multiple projects exist in separate per-project databases, when calling `FindAll(ctx)`, then all projects from all DBs are aggregated and returned as a single slice.

3. **AC3: Save routing** - Given project to save, when calling `Save(ctx, project)`, then data is written to the correct project's state.db file (determined by project path → directory mapping).

4. **AC4: Delete routing** - Given project ID to delete, when calling `Delete(ctx, id)`, then project is removed from the correct project's state.db.

5. **AC5: Lazy loading** - Given 20 projects tracked, when dashboard loads, then connections are opened lazily on-demand (not all databases opened at startup).

6. **AC6: FindByID/FindByPath search** - Given project ID or path, when calling `FindByID(ctx, id)` or `FindByPath(ctx, path)`, then the coordinator searches across all project databases and returns the matching project or `domain.ErrProjectNotFound`.

7. **AC7: FindActive/FindHibernated aggregation** - Given projects with different states across databases, when calling `FindActive(ctx)` or `FindHibernated(ctx)`, then matching projects from all DBs are aggregated.

8. **AC8: UpdateState routing** - Given project ID and new state, when calling `UpdateState(ctx, id, state)`, then the correct project's database is updated.

9. **AC9: Context cancellation** - Given any repository operation in progress, when context is cancelled, then operation returns `ctx.Err()` promptly.

10. **AC10: Graceful degradation** - Given corrupted project database (one of many), when coordinator enumerates projects, then corrupted DB is logged with warning and skipped, other projects returned successfully.

11. **AC11: Empty slice return** - Given no projects match query (Find*, FindActive, FindHibernated), when called, then empty slice `[]` is returned (not nil).

12. **AC12: Close lifecycle** - Given coordinator with cached repositories, when `Close(ctx)` is called, then all cached connections are closed and cache is cleared.

13. **AC13: New project creation via Save** - Given project not in master config, when calling `Save(ctx, project)`, then directory is created via DirectoryManager, config is updated, and data is saved.

## Tasks / Subtasks

- [x] Task 1: Create `RepositoryCoordinator` struct (AC: 1, 5, 12, 13)
  - [x] Subtask 1.1: Create `internal/adapters/persistence/coordinator.go`
  - [x] Subtask 1.2: Define struct with fields:
    - `configLoader ports.ConfigLoader` - loads master config for project enumeration
    - `directoryManager ports.DirectoryManager` - creates directories for new projects (REQUIRED for AC13)
    - `basePath string` - base path for project directories (`~/.vibe-dash`)
    - `repoCache map[string]*sqlite.ProjectRepository` - lazy-loaded repos
    - `mu sync.RWMutex` - thread-safe cache access
  - [x] Subtask 1.3: Create constructor:
    ```go
    func NewRepositoryCoordinator(
        configLoader ports.ConfigLoader,
        directoryManager ports.DirectoryManager,
        basePath string,
    ) *RepositoryCoordinator
    ```
  - [x] Subtask 1.4: Add compile-time interface check at package level:
    ```go
    var _ ports.ProjectRepository = (*RepositoryCoordinator)(nil)
    ```

- [x] Task 2: Implement helper methods for lazy loading and cache management (AC: 5, 9, 10, 12)
  - [x] Subtask 2.1: Implement `getProjectRepo(ctx context.Context, directoryName string) (*sqlite.ProjectRepository, error)`:
    - Context cancellation check first
    - Read lock → check cache → return if exists
    - Write lock → double-check → create → cache → return
    - Wrap errors with directory context
  - [x] Subtask 2.2: Implement `invalidateCache(directoryName string)`:
    - Write lock → delete from cache map
    - Used after Delete operations to prevent stale cache
  - [x] Subtask 2.3: Implement `Close(ctx context.Context) error`:
    - Write lock entire cache
    - Clear cache map (set to empty)
    - Note: sqlite.ProjectRepository uses lazy connections (open-per-operation), so no explicit close needed per repo

- [x] Task 3: Implement helper method `getAllRepos` for enumeration (AC: 5, 9, 10)
  - [x] Subtask 3.1: Implement `getAllRepos(ctx context.Context) ([]*sqlite.ProjectRepository, []string, error)`:
    - Returns: (repos slice, directory names slice in same order, error)
    - Directory names needed for cache invalidation on Delete
  - [x] Subtask 3.2: Context cancellation check first
  - [x] Subtask 3.3: Load config via `configLoader.Load(ctx)` - return error if config load fails (not just log)
  - [x] Subtask 3.4: Iterate `config.Projects` map:
    - For each project, call `getProjectRepo(ctx, dirName)`
    - On individual repo error: `slog.Warn("skipping corrupted project", "directory", dirName, "error", err)` and continue
    - Track successful repos AND their directory names
  - [x] Subtask 3.5: Return `make([]*sqlite.ProjectRepository, 0)` if all fail (graceful degradation)

- [x] Task 4: Implement `FindAll` aggregation (AC: 2, 9, 11)
  - [x] Subtask 4.1: Context cancellation check at start
  - [x] Subtask 4.2: Call `getAllRepos(ctx)` - propagate error if config load fails
  - [x] Subtask 4.3: For each repo, call `repo.FindAll(ctx)` and aggregate results
  - [x] Subtask 4.4: Return `make([]*domain.Project, 0)` if no projects (CRITICAL: empty slice, not nil)

- [x] Task 5: Implement `FindActive` and `FindHibernated` aggregation (AC: 7, 9, 11)
  - [x] Subtask 5.1: Context cancellation check at start
  - [x] Subtask 5.2: Call `getAllRepos(ctx)` - propagate error if config load fails
  - [x] Subtask 5.3: For each repo, call `repo.FindActive(ctx)` / `repo.FindHibernated(ctx)` and aggregate
  - [x] Subtask 5.4: Return `make([]*domain.Project, 0)` if no matching (CRITICAL: empty slice, not nil)

- [x] Task 6: Implement `Save` routing (AC: 3, 9, 13)
  - [x] Subtask 6.1: Context cancellation check at start
  - [x] Subtask 6.2: Load config via `configLoader.Load(ctx)`
  - [x] Subtask 6.3: Try `cfg.GetDirectoryName(project.Path)` to find existing directory
  - [x] Subtask 6.4: If NOT found (new project - AC13):
    - Call `directoryManager.EnsureProjectDir(ctx, project.Path)` to create directory
    - Extract directory name from returned full path: `filepath.Base(fullPath)`
    - Update config: `cfg.SetProjectEntry(dirName, project.Path, project.DisplayName, project.IsFavorite)`
    - Save config: `configLoader.Save(ctx, cfg)`
  - [x] Subtask 6.5: Get or create repo via `getProjectRepo(ctx, directoryName)`
  - [x] Subtask 6.6: Call `repo.Save(ctx, project)`

- [x] Task 7: Implement `FindByID` and `FindByPath` (AC: 6, 9)
  - [x] Subtask 7.1: Context cancellation check at start
  - [x] Subtask 7.2: For `FindByPath`:
    - First try fast path: `cfg.GetDirectoryName(path)` → get single repo → `repo.FindByPath`
    - If directory not in config: fall back to iterating all repos (handles edge case of unregistered path)
  - [x] Subtask 7.3: For `FindByID`: iterate all repos (via `getAllRepos`), call `repo.FindByID` until found
  - [x] Subtask 7.4: Return `domain.ErrProjectNotFound` if not found in any repo

- [x] Task 8: Implement `Delete` routing (AC: 4, 9)
  - [x] Subtask 8.1: Context cancellation check at start
  - [x] Subtask 8.2: Find project first via `FindByID(ctx, id)` to get path
  - [x] Subtask 8.3: Load config, get directory name: `cfg.GetDirectoryName(project.Path)`
  - [x] Subtask 8.4: Get repo via `getProjectRepo(ctx, directoryName)`
  - [x] Subtask 8.5: Call `repo.Delete(ctx, id)`
  - [x] Subtask 8.6: Call `invalidateCache(directoryName)` to prevent stale cache

- [x] Task 9: Implement `UpdateState` routing (AC: 8, 9)
  - [x] Subtask 9.1: Context cancellation check at start
  - [x] Subtask 9.2: Find project first via `FindByID(ctx, id)` to get path
  - [x] Subtask 9.3: Load config, get directory name, get repo
  - [x] Subtask 9.4: Call `repo.UpdateState(ctx, id, state)`

- [x] Task 10: Write unit tests (AC: all) - table-driven per project standards
  - [x] Subtask 10.1: Test `NewRepositoryCoordinator` creates valid coordinator
  - [x] Subtask 10.2: Test `FindAll` aggregates from multiple project DBs
  - [x] Subtask 10.3: Test `FindActive`/`FindHibernated` filter correctly
  - [x] Subtask 10.4: Test `Save` routes to correct project DB (existing project)
  - [x] Subtask 10.5: Test `Save` creates new project via DirectoryManager (AC13)
  - [x] Subtask 10.6: Test `Delete` removes from correct DB and invalidates cache
  - [x] Subtask 10.7: Test `FindByID` searches across all DBs
  - [x] Subtask 10.8: Test `FindByPath` fast path (config lookup) and fallback
  - [x] Subtask 10.9: Test graceful degradation (corrupted DB logged, others succeed)
  - [x] Subtask 10.10: Test context cancellation returns promptly
  - [x] Subtask 10.11: Test empty results return empty slice (not nil)
  - [x] Subtask 10.12: Test `Close` clears cache

- [x] Task 11: Write integration tests (AC: 1, 2, 3, 4, 12, 13)
  - [x] Subtask 11.1: Test full lifecycle: add 3 projects → FindAll returns 3
  - [x] Subtask 11.2: Test Save → FindByID → Delete cycle
  - [x] Subtask 11.3: Test service layer uses coordinator without code changes
  - [x] Subtask 11.4: Verify lazy loading (DB opened only when accessed)
  - [x] Subtask 11.5: Test new project creation flow (AC13)
  - [x] Subtask 11.6: Test Close lifecycle

## Dev Notes

### Architecture Overview

The `RepositoryCoordinator` is the **aggregator** that implements `ports.ProjectRepository` while internally delegating to multiple per-project `sqlite.ProjectRepository` instances. It also depends on `ports.DirectoryManager` for creating new project directories.

```
                    ┌────────────────────────────┐
                    │     Service Layer          │
                    │  (DetectionService, etc.)  │
                    └─────────────┬──────────────┘
                                  │ uses ports.ProjectRepository
                    ┌─────────────▼──────────────┐
                    │   RepositoryCoordinator    │
                    │  implements ProjectRepository│
                    │  + DirectoryManager (new)   │
                    │  + ConfigLoader (enum)      │
                    └─────────────┬──────────────┘
                                  │ delegates to
           ┌──────────────────────┼──────────────────────┐
           ▼                      ▼                      ▼
    ┌─────────────┐        ┌─────────────┐        ┌─────────────┐
    │ ProjectRepo │        │ ProjectRepo │        │ ProjectRepo │
    │ (api-service)│       │ (client-b)  │        │ (my-project)│
    └─────────────┘        └─────────────┘        └─────────────┘
         │                      │                      │
    ~/.vibe-dash/         ~/.vibe-dash/          ~/.vibe-dash/
    api-service/          client-b-api/          my-project/
    state.db              state.db               state.db
```

**Key Dependencies:**
- `ports.ConfigLoader` - enumerates all registered projects via `config.Projects` map
- `ports.DirectoryManager` - creates directories for new projects (AC13)
- `sqlite.ProjectRepository` - per-project database access

### Key Design Decisions

**1. Lazy Loading (AC5):**
- Repositories are created on first access, not at coordinator construction
- Uses cache (`repoCache map[string]*sqlite.ProjectRepository`) to avoid recreating
- Prevents file handle exhaustion with many projects
- Cache cleared via `Close()` method for lifecycle management

**2. Directory Discovery:**
- Master config (`ports.ConfigLoader`) provides project → directory mappings
- Each project's directory contains `.project-path` marker (from Story 3.5.1)
- Coordinator uses `config.Projects` to enumerate all tracked projects

**3. Graceful Degradation (AC10):**
- If one project's DB is corrupted, log warning and continue
- Other projects should still be accessible
- Pattern from Story 3.5.3: `slog.Warn` + continue, don't fail entire operation
- Config load failure IS fatal (returns error, not empty result)

**4. New Project Creation (AC13):**
- When `Save()` is called for a project not in config
- Use `DirectoryManager.EnsureProjectDir()` to create directory with `.project-path` marker
- Update config via `cfg.SetProjectEntry()` then `configLoader.Save()`
- Only then proceed to save project data

**5. Cache Invalidation:**
- After `Delete()`, call `invalidateCache(dirName)` to prevent stale entries
- `Close()` clears entire cache for clean shutdown

### Implementation Patterns

**Context Cancellation (REQUIRED at start of every public method):**
```go
select {
case <-ctx.Done():
    return nil, ctx.Err()
default:
}
```

**Struct Definition:**
```go
// RepositoryCoordinator aggregates multiple per-project repositories.
// Implements ports.ProjectRepository for seamless service layer integration.
type RepositoryCoordinator struct {
    configLoader     ports.ConfigLoader
    directoryManager ports.DirectoryManager
    basePath         string
    repoCache        map[string]*sqlite.ProjectRepository
    mu               sync.RWMutex
}

// Compile-time interface check
var _ ports.ProjectRepository = (*RepositoryCoordinator)(nil)

func NewRepositoryCoordinator(
    configLoader ports.ConfigLoader,
    directoryManager ports.DirectoryManager,
    basePath string,
) *RepositoryCoordinator {
    return &RepositoryCoordinator{
        configLoader:     configLoader,
        directoryManager: directoryManager,
        basePath:         basePath,
        repoCache:        make(map[string]*sqlite.ProjectRepository),
    }
}
```

**Lazy Loading + Cache Invalidation:**
```go
func (c *RepositoryCoordinator) getProjectRepo(ctx context.Context, dirName string) (*sqlite.ProjectRepository, error) {
    // Context check
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Read lock first
    c.mu.RLock()
    if repo, ok := c.repoCache[dirName]; ok {
        c.mu.RUnlock()
        return repo, nil
    }
    c.mu.RUnlock()

    // Write lock for creation
    c.mu.Lock()
    defer c.mu.Unlock()

    // Double-check after acquiring write lock
    if repo, ok := c.repoCache[dirName]; ok {
        return repo, nil
    }

    projectDir := filepath.Join(c.basePath, dirName)
    repo, err := sqlite.NewProjectRepository(projectDir)
    if err != nil {
        return nil, fmt.Errorf("failed to open project %s: %w", dirName, err)
    }

    c.repoCache[dirName] = repo
    return repo, nil
}

func (c *RepositoryCoordinator) invalidateCache(dirName string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.repoCache, dirName)
}

func (c *RepositoryCoordinator) Close(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    c.mu.Lock()
    defer c.mu.Unlock()
    c.repoCache = make(map[string]*sqlite.ProjectRepository)
    return nil
}
```

**getAllRepos - Returns Error on Config Failure:**
```go
func (c *RepositoryCoordinator) getAllRepos(ctx context.Context) ([]*sqlite.ProjectRepository, []string, error) {
    select {
    case <-ctx.Done():
        return nil, nil, ctx.Err()
    default:
    }

    cfg, err := c.configLoader.Load(ctx)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to load config: %w", err) // Fatal - propagate
    }

    repos := make([]*sqlite.ProjectRepository, 0, len(cfg.Projects))
    dirNames := make([]string, 0, len(cfg.Projects))
    for dirName := range cfg.Projects {
        repo, err := c.getProjectRepo(ctx, dirName)
        if err != nil {
            slog.Warn("skipping corrupted project", "directory", dirName, "error", err)
            continue // Graceful degradation for individual project
        }
        repos = append(repos, repo)
        dirNames = append(dirNames, dirName)
    }
    return repos, dirNames, nil
}
```

**Save with New Project Creation (AC13):**
```go
func (c *RepositoryCoordinator) Save(ctx context.Context, project *domain.Project) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    cfg, err := c.configLoader.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    dirName, found := cfg.GetDirectoryName(project.Path)
    if !found {
        // NEW project - create directory and update config
        fullPath, err := c.directoryManager.EnsureProjectDir(ctx, project.Path)
        if err != nil {
            return fmt.Errorf("failed to create project directory: %w", err)
        }
        dirName = filepath.Base(fullPath)
        cfg.SetProjectEntry(dirName, project.Path, project.DisplayName, project.IsFavorite)
        if err := c.configLoader.Save(ctx, cfg); err != nil {
            return fmt.Errorf("failed to save config: %w", err)
        }
    }

    repo, err := c.getProjectRepo(ctx, dirName)
    if err != nil {
        return err
    }
    return repo.Save(ctx, project)
}
```

### Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/adapters/persistence/coordinator.go` | CREATE | Main coordinator implementation with compile-time interface check |
| `internal/adapters/persistence/coordinator_test.go` | CREATE | Unit tests (table-driven) |
| `internal/adapters/persistence/coordinator_integration_test.go` | CREATE | Integration tests with `//go:build integration` tag |

**Required Imports for coordinator.go:**
```go
import (
    "context"
    "fmt"
    "log/slog"
    "path/filepath"
    "sync"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/sqlite"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)
```

### Existing Code Context

**ports.ProjectRepository interface (from `internal/core/ports/repository.go:22-59`):**
```go
type ProjectRepository interface {
    Save(ctx context.Context, project *domain.Project) error
    FindByID(ctx context.Context, id string) (*domain.Project, error)
    FindByPath(ctx context.Context, path string) (*domain.Project, error)
    FindAll(ctx context.Context) ([]*domain.Project, error)
    FindActive(ctx context.Context) ([]*domain.Project, error)
    FindHibernated(ctx context.Context) ([]*domain.Project, error)
    Delete(ctx context.Context, id string) error
    UpdateState(ctx context.Context, id string, state domain.ProjectState) error
}
```

**ports.ConfigLoader interface (from `internal/core/ports/config.go:204-221`):**
```go
type ConfigLoader interface {
    Load(ctx context.Context) (*Config, error)
    Save(ctx context.Context, config *Config) error
}
```

**ports.DirectoryManager interface (from `internal/core/ports/directory.go:13-44`):**
```go
type DirectoryManager interface {
    GetProjectDirName(ctx context.Context, projectPath string) (string, error)
    EnsureProjectDir(ctx context.Context, projectPath string) (string, error)
}
```

**ports.Config lookup methods (from `internal/core/ports/config.go:106-165`):**
```go
// GetDirectoryName returns directory name for project path
cfg.GetDirectoryName(path string) (string, bool)

// GetProjectPath returns canonical path for directory name
cfg.GetProjectPath(directoryName string) (string, bool)

// SetProjectEntry adds/updates a project entry
cfg.SetProjectEntry(directoryName, path, displayName string, favorite bool)

// Projects map[string]ProjectConfig - directory_name as key
cfg.Projects
```

**sqlite.NewProjectRepository (from `internal/adapters/persistence/sqlite/project_repository.go:38-64`):**
```go
// Creates per-project repository, validates .project-path marker exists
// Returns domain.ErrPathNotAccessible if directory missing or no marker
func NewProjectRepository(projectDir string) (*ProjectRepository, error)
```

### Thread Safety Considerations

- `repoCache` must be protected with `sync.RWMutex`
- Read operations use RLock for concurrent access
- Cache updates use Lock for exclusive access
- Each `sqlite.ProjectRepository` is already thread-safe (lazy connections - open-per-operation)
- Double-check pattern required in `getProjectRepo` (check after acquiring write lock)
- `invalidateCache` and `Close` require write lock

### Error Handling

| Scenario | Error Type | Handling |
|----------|------------|----------|
| Context cancelled | `ctx.Err()` | Return immediately (check at start of every method) |
| Project not in config (FindByID/FindByPath) | `domain.ErrProjectNotFound` | Return error |
| Project not in config (Save) | N/A - create new | Use DirectoryManager.EnsureProjectDir, update config |
| Corrupted individual project DB | `slog.Warn` | Skip this project, continue with others (graceful degradation) |
| Config load failure | Wrapped error | Return error (FATAL - not graceful) |
| All projects corrupted | Return empty slice | Not an error (graceful degradation) |
| DirectoryManager.EnsureProjectDir fails | Wrapped error | Return error (can't create new project) |

### Testing Strategy

**Unit Tests (mock ConfigLoader and DirectoryManager):**
- Mock `ports.ConfigLoader` to return controlled config
- Mock `ports.DirectoryManager` for new project creation tests
- Create temp directories with `.project-path` markers
- Test each method individually with known state
- Table-driven tests per project standards (`tests []struct{...}` pattern)

**Integration Tests (`//go:build integration`):**
- Real filesystem with `t.TempDir()`
- Real SQLite databases
- Full lifecycle tests
- New project creation flow (AC13)
- Close lifecycle (AC12)

### Dependencies

**Depends on (COMPLETED):**
- Story 3.5.1: `DirectoryManager` - Creates directories with `.project-path` marker, collision resolution
- Story 3.5.2: `sqlite.ProjectRepository` - Per-project database, validates marker exists
- Story 3.5.3: Per-Project Config Files - Project config cascade (per-project settings)
- Story 3.5.4: Master Config as Path Index - `config.GetDirectoryName()`, `config.SetProjectEntry()`, `config.Projects`

**Required by:**
- Story 3.5.6: Update CLI Commands - Wire `RepositoryCoordinator` in `main.go`
- Story 3.5.7: Integration Testing - Full lifecycle tests using coordinator

### Previous Story Learnings (3.5.4)

From Story 3.5.4 code review - MUST apply these patterns:

1. **Context cancellation** at start of every public method:
   ```go
   select {
   case <-ctx.Done():
       return nil, ctx.Err()
   default:
   }
   ```

2. **Graceful degradation**: log warnings with structured context, continue on individual failures:
   ```go
   slog.Warn("skipping corrupted project", "directory", dirName, "error", err)
   continue
   ```

3. **Compile-time interface check** at package level:
   ```go
   var _ ports.ProjectRepository = (*RepositoryCoordinator)(nil)
   ```

4. **Table-driven tests** for all edge cases - use `tests []struct{...}` pattern

5. **Empty slice return**: `make([]*domain.Project, 0)` not `nil` - CRITICAL for JSON serialization

6. **Error wrapping**: `fmt.Errorf("failed to X: %w", err)` - wrap with context

### Project Structure Notes

After this story, the full storage structure is:
```
~/.vibe-dash/
  ├── config.yaml                 # Master index (Story 3.5.4)
  ├── api-service/
  │   ├── .project-path           # Marker file (Story 3.5.1)
  │   ├── config.yaml             # Per-project settings (Story 3.5.3)
  │   └── state.db                # Per-project SQLite (Story 3.5.2)
  └── client-b-api-service/
      ├── .project-path
      ├── config.yaml
      └── state.db
```

`RepositoryCoordinator` reads `config.yaml`, enumerates projects, lazily opens each `state.db`. New projects created via `Save()` trigger `DirectoryManager.EnsureProjectDir()` to create the directory structure before saving.

### References

- [Source: docs/sprint-artifacts/stories/epic-3.5/epic-3.5-storage-structure.md#Story 3.5.5]
- [Source: docs/prd.md#lines 597-605] - Storage structure
- [Source: docs/architecture.md#lines 912-927] - Registry pattern for aggregation
- [Source: docs/project-context.md#SQLite Rules] - Lazy connections, open-per-operation
- [Source: docs/project-context.md#Go Patterns] - Context first, error wrapping
- [Source: internal/core/ports/repository.go:22-59] - ProjectRepository interface
- [Source: internal/core/ports/config.go:106-165] - Config path lookup methods
- [Source: internal/core/ports/config.go:204-221] - ConfigLoader interface
- [Source: internal/core/ports/directory.go:13-44] - DirectoryManager interface
- [Source: internal/adapters/persistence/sqlite/project_repository.go:38-64] - Per-project repository
- [Source: internal/adapters/filesystem/directory.go] - DirectoryManager implementation

## Dev Agent Record

### Context Reference

- Epic 3.5: Storage Structure Alignment
- Story Dependencies: Depends on 3.5.1, 3.5.2, 3.5.3, 3.5.4; required by 3.5.6, 3.5.7
- PRD Reference: Lines 597-605 (Storage structure)
- Architecture Reference: Lines 912-927 (Registry coordination pattern)
- Project Context: SQLite Rules (lazy connections, WAL mode)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

### Completion Notes List

1. All 11 tasks and 46 subtasks completed successfully
2. All unit tests pass (17 tests covering all ACs)
3. All integration tests pass (6 tests with `//go:build integration` tag)
4. Implementation follows all project patterns from previous stories:
   - Context cancellation checks at start of every public method
   - Graceful degradation with `slog.Warn` for individual project failures
   - Empty slice return `make([]*domain.Project, 0)` not nil
   - Compile-time interface check `var _ ports.ProjectRepository = (*RepositoryCoordinator)(nil)`
   - Error wrapping with `fmt.Errorf("failed to X: %w", err)`
5. Thread safety implemented with `sync.RWMutex` and double-check pattern
6. Lazy loading working correctly - repos only created on first access
7. New project creation (AC13) integrates with DirectoryManager and updates master config

### Code Review (2025-12-18)

**Reviewer:** Amelia (Dev Agent) - Adversarial Code Review

**Issues Found and Fixed:**

1. **[HIGH] Delete didn't remove project from config** (`coordinator.go:326-367`)
   - Problem: Delete removed from DB but left orphaned entry in config
   - Fix: Added `cfg.RemoveProject(dirName)` and `configLoader.Save()` after DB delete
   - Test updated: `TestDelete_RemovesAndInvalidatesCache` now verifies config removal

2. **[MEDIUM] Integration tests used unsafe string(rune) conversion** (`coordinator_integration_test.go`)
   - Problem: `string(rune('0'+i))` breaks for i > 9
   - Fix: Changed to `strconv.Itoa(i)` for safe integer-to-string conversion

**Issues Noted (Not Fixed - Optimization, Not Bug):**

3. **[MEDIUM] FindByPath loads config twice in some paths** - Performance optimization for future
4. **[MEDIUM] Delete/UpdateState load config after FindByID** - Accepted tradeoff for code clarity
5. **[MEDIUM] Map iteration non-deterministic** - Not a bug, just non-deterministic order

**All tests passing after fixes: 17 unit tests, 6 integration tests**

### File List

| File | Action | Lines |
|------|--------|-------|
| `internal/adapters/persistence/coordinator.go` | CREATE | 401 |
| `internal/adapters/persistence/coordinator_test.go` | CREATE | 816 |
| `internal/adapters/persistence/coordinator_integration_test.go` | CREATE | 441 |
