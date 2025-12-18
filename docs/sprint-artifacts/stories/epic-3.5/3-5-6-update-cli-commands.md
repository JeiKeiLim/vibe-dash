# Story 3.5.6: Update CLI Commands

Status: done

## Story

As a user,
I want CLI commands to work with the new storage structure,
So that I can add, list, and remove projects as before.

## Acceptance Criteria

1. **AC1: `vibe add` creates project directory structure** - Given new storage structure, when running `vibe add /path/to/project`, then project directory, state.db, and config.yaml are created in `~/.vibe-dash/<project>/`.

2. **AC2: `vibe remove` deletes entire project directory** - Given project exists, when running `vibe remove <name>`, then entire project directory is deleted from `~/.vibe-dash/`.

3. **AC3: `vibe list` shows all projects from new structure** - Given multiple projects, when running `vibe list`, then all projects are shown from the new per-project storage structure.

4. **AC4: Collision handling automatic** - Given project with collision, when running `vibe add`, then collision is resolved automatically using `DirectoryManager`.

5. **AC5: Service layer uses RepositoryCoordinator** - Given new wiring in `main.go`, when service layer uses `ports.ProjectRepository`, then `RepositoryCoordinator` is used transparently (no service code changes).

6. **AC6: Existing CLI tests pass** - Given refactored CLI commands, when running `go test ./internal/adapters/cli/...`, then all existing tests pass.

7. **AC7: CLI test infrastructure updated** - Given new dependencies (ConfigLoader, DirectoryManager), when running CLI tests, then tests use appropriate mocks or real implementations.

## Tasks / Subtasks

- [x] Task 1: Create `GetDefaultBasePath` helper (AC: 5)
  - [x] Subtask 1.1: Create `internal/config/paths.go` with helper function:
    ```go
    package config

    import (
        "os"
        "path/filepath"
    )

    // GetDefaultBasePath returns the default vibe-dash storage directory.
    // Returns ~/.vibe-dash on success, empty string on home dir lookup failure.
    func GetDefaultBasePath() string {
        home, err := os.UserHomeDir()
        if err != nil {
            return ""
        }
        return filepath.Join(home, ".vibe-dash")
    }
    ```
  - [x] Subtask 1.2: Add test for `GetDefaultBasePath` in `internal/config/paths_test.go`

- [x] Task 2: Update `ports.DirectoryManager` interface (AC: 2)
  - [x] Subtask 2.1: Add `DeleteProjectDir` method to `internal/core/ports/directory.go`:
    ```go
    // DeleteProjectDir removes the project directory and all its contents.
    // The projectPath is the canonical project path (used to look up directory name).
    // Returns nil if directory doesn't exist (idempotent).
    // Returns error if deletion fails for reasons other than non-existence.
    DeleteProjectDir(ctx context.Context, projectPath string) error
    ```

- [x] Task 3: Implement `DeleteProjectDir` in FilesystemDirectoryManager (AC: 2)
  - [x] Subtask 3.1: Add method to `internal/adapters/filesystem/directory.go`:
    ```go
    // DeleteProjectDir removes the project directory for the given project path.
    // Returns nil if directory doesn't exist (idempotent).
    func (dm *FilesystemDirectoryManager) DeleteProjectDir(ctx context.Context, projectPath string) error {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // Look up directory name for this project path
        if dm.configLookup == nil {
            return nil // No lookup available, nothing to delete
        }

        canonicalPath, err := CanonicalPath(projectPath)
        if err != nil {
            return nil // Path doesn't exist, nothing to delete
        }

        dirName := dm.configLookup.GetDirForPath(canonicalPath)
        if dirName == "" {
            return nil // Not tracked, nothing to delete
        }

        dirPath := filepath.Join(dm.basePath, dirName)

        // Safety check - only delete within base path
        cleanDirPath := filepath.Clean(dirPath)
        cleanBasePath := filepath.Clean(dm.basePath)
        if !strings.HasPrefix(cleanDirPath, cleanBasePath) {
            return fmt.Errorf("%w: directory %s is outside base path", domain.ErrPathNotAccessible, dirPath)
        }

        // Check if directory exists
        if _, err := os.Stat(dirPath); os.IsNotExist(err) {
            return nil // Already deleted, idempotent
        }

        if err := os.RemoveAll(dirPath); err != nil {
            return fmt.Errorf("%w: failed to delete project directory %s: %v", domain.ErrPathNotAccessible, dirPath, err)
        }
        return nil
    }
    ```
  - [x] Subtask 3.2: Add `"strings"` to imports in `directory.go`
  - [x] Subtask 3.3: Add unit tests for `DeleteProjectDir` in `directory_test.go`

- [x] Task 4: Update `main.go` wiring (AC: 5)
  - [x] Subtask 4.1: Add required imports to `cmd/vibe/main.go`:
    ```go
    import (
        "context"
        "log/slog"
        "os"
        "os/signal"
        "syscall"

        "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
        "github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors"
        "github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors/speckit"
        "github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
        "github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence"
        "github.com/JeiKeiLim/vibe-dash/internal/config"
        "github.com/JeiKeiLim/vibe-dash/internal/core/services"
    )
    ```
  - [x] Subtask 4.2: Create configPathAdapter (unexported type in main.go):
    ```go
    // configPathAdapter implements ports.ProjectPathLookup for DirectoryManager.
    type configPathAdapter struct {
        loader *config.ViperLoader
    }

    func (a *configPathAdapter) GetDirForPath(path string) string {
        cfg, err := a.loader.Load(context.Background())
        if err != nil {
            return ""
        }
        dirName, _ := cfg.GetDirectoryName(path)
        return dirName
    }
    ```
  - [x] Subtask 4.3: Get base path with nil safety:
    ```go
    basePath := config.GetDefaultBasePath()
    if basePath == "" {
        return fmt.Errorf("failed to determine base path: cannot access home directory")
    }
    ```
  - [x] Subtask 4.4: Create DirectoryManager with nil check:
    ```go
    configAdapter := &configPathAdapter{loader: loader}
    dirMgr := filesystem.NewDirectoryManager(basePath, configAdapter)
    if dirMgr == nil {
        return fmt.Errorf("failed to initialize directory manager: cannot determine base path")
    }
    ```
  - [x] Subtask 4.5: Create RepositoryCoordinator:
    ```go
    coordinator := persistence.NewRepositoryCoordinator(loader, dirMgr, basePath)
    ```
  - [x] Subtask 4.6: Wire all dependencies to CLI:
    ```go
    cli.SetRepository(coordinator)
    cli.SetDirectoryManager(dirMgr)
    cli.SetBasePath(basePath)
    ```
  - [x] Subtask 4.7: Remove old `sqlite.NewSQLiteRepository` code and import

- [x] Task 5: Update `vibe add` command (AC: 1, 4)
  - [x] Subtask 5.1: Verify `add.go` works with `RepositoryCoordinator.Save()` - should work as-is because coordinator handles directory creation internally
  - [x] Subtask 5.2: **DO NOT remove collision handling** from `add.go` - it handles DISPLAY NAME collisions (user-facing), which is DIFFERENT from DirectoryManager's FILESYSTEM collisions. Both are needed:
    - DirectoryManager: resolves `~/.vibe-dash/api-service/` vs `~/.vibe-dash/client-b-api-service/` (filesystem)
    - add.go: resolves dashboard display names to avoid user confusion (user interface)
  - [x] Subtask 5.3: Test that `vibe add /path` creates `~/.vibe-dash/<project>/` with:
    - `.project-path` marker file
    - `state.db` SQLite database
  - [x] Subtask 5.4: Update success message to reflect new structure if needed

- [x] Task 6: Update `vibe remove` command (AC: 2)
  - [x] Subtask 6.1: Verify current `remove.go` - uses `repository.Delete(ctx, id)`
  - [x] Subtask 6.2: Note: `RepositoryCoordinator.Delete()` removes from DB and config, but does NOT delete the physical directory
  - [x] Subtask 6.3: Add directory deletion using DirectoryManager after `repository.Delete()`:
    ```go
    // Delete from repository (removes from DB and config)
    if err := repository.Delete(ctx, project.ID); err != nil {
        return fmt.Errorf("failed to remove project: %w", err)
    }

    // Delete project directory (non-fatal if fails)
    if directoryManager != nil {
        if err := directoryManager.DeleteProjectDir(ctx, project.Path); err != nil {
            slog.Warn("failed to delete project directory", "path", project.Path, "error", err)
            // Continue - project removed from tracking, directory left behind
        }
    }
    ```
  - [x] Subtask 6.4: Add required imports to `remove.go`:
    ```go
    import (
        "log/slog"
        // ... existing imports ...
    )
    ```

- [x] Task 7: Verify `vibe list` command (AC: 3)
  - [x] Subtask 7.1: Verify `list.go` works with `RepositoryCoordinator.FindAll()` - should work as-is
  - [x] Subtask 7.2: Test that list aggregates from all per-project databases
  - [x] Subtask 7.3: No code changes expected - just verification

- [x] Task 8: Update CLI package dependencies (AC: 5, 7)
  - [x] Subtask 8.1: Add new package-level variables to `internal/adapters/cli/deps.go` (new file):
    ```go
    package cli

    import "github.com/JeiKeiLim/vibe-dash/internal/core/ports"

    // directoryManager handles project directory operations (delete).
    var directoryManager ports.DirectoryManager

    // basePath is the base directory for project storage (~/.vibe-dash).
    var basePath string

    // SetDirectoryManager sets the directory manager for CLI commands.
    func SetDirectoryManager(dm ports.DirectoryManager) {
        directoryManager = dm
    }

    // SetBasePath sets the base path for project storage.
    func SetBasePath(path string) {
        basePath = path
    }
    ```
  - [x] Subtask 8.2: Note: `repository` and `SetRepository` already exist in `add.go` - reuse those
  - [x] Subtask 8.3: Update `main.go` to call new setters (done in Task 4.6)

- [x] Task 9: Update CLI tests (AC: 6, 7)
  - [x] Subtask 9.1: Review all existing CLI tests for repository mock usage
  - [x] Subtask 9.2: Create mock for `ports.DirectoryManager` in `internal/adapters/cli/mocks_test.go`:
    ```go
    type mockDirectoryManager struct {
        deleteErr error
        deleteCalls []string // Track deleted paths
    }

    func (m *mockDirectoryManager) GetProjectDirName(ctx context.Context, projectPath string) (string, error) {
        return filepath.Base(projectPath), nil
    }

    func (m *mockDirectoryManager) EnsureProjectDir(ctx context.Context, projectPath string) (string, error) {
        return filepath.Join("/tmp/test", filepath.Base(projectPath)), nil
    }

    func (m *mockDirectoryManager) DeleteProjectDir(ctx context.Context, projectPath string) error {
        m.deleteCalls = append(m.deleteCalls, projectPath)
        return m.deleteErr
    }
    ```
  - [x] Subtask 9.3: Update remove tests to verify `DeleteProjectDir` is called
  - [x] Subtask 9.4: Run `go test ./internal/adapters/cli/...` and fix any failures
  - [x] Subtask 9.5: Verify all existing tests still pass with new wiring

- [x] Task 10: Add new CLI tests for storage structure (AC: 1, 2, 4)
  - [x] Subtask 10.1: Test `vibe add` creates correct directory structure (via RepositoryCoordinator)
  - [x] Subtask 10.2: Test `vibe remove` calls `DeleteProjectDir` after repo delete
  - [x] Subtask 10.3: Test remove handles `DeleteProjectDir` error gracefully (non-fatal)
  - [x] Subtask 10.4: Test collision handling works (display names via add.go, filesystem via DirectoryManager)

- [x] Task 11: Run full test suite
  - [x] Subtask 11.1: Run `make test` - all unit tests
  - [x] Subtask 11.2: Run `make lint` - no linting errors (warnings in pre-existing test code only)
  - [x] Subtask 11.3: Manual testing:
    1. Fresh start: `rm -rf ~/.vibe-dash/`
    2. Add: `vibe add .` - creates structure
    3. List: `vibe list` - shows project
    4. Remove: `vibe remove <name>` - cleans up completely
    5. Verify: `ls ~/.vibe-dash/` - project directory gone

## Dev Notes

### Architecture Overview

This story wires the new storage components (from Stories 3.5.1-3.5.5) into the CLI layer. The key change is replacing the single-DB `SQLiteRepository` with the multi-DB `RepositoryCoordinator`.

```
                    ┌─────────────────────────────────────┐
                    │              cmd/vibe/main.go        │
                    │  - Creates ConfigLoader              │
                    │  - Creates DirectoryManager          │
                    │  - Creates RepositoryCoordinator     │
                    │  - Wires to CLI package              │
                    └──────────────┬──────────────────────┘
                                   │ SetRepository(), SetConfigLoader(), etc.
                    ┌──────────────▼──────────────────────┐
                    │         internal/adapters/cli/       │
                    │  - add.go uses repository.Save()     │
                    │  - remove.go uses repository.Delete()│
                    │    + directoryManager.DeleteDir()    │
                    │  - list.go uses repository.FindAll() │
                    └──────────────┬──────────────────────┘
                                   │ ports.ProjectRepository
                    ┌──────────────▼──────────────────────┐
                    │      RepositoryCoordinator           │
                    │  - Aggregates per-project DBs        │
                    │  - Creates new projects via Save()   │
                    └──────────────┬──────────────────────┘
                                   │
           ┌───────────────────────┼───────────────────────┐
           ▼                       ▼                       ▼
    ~/.vibe-dash/           ~/.vibe-dash/           ~/.vibe-dash/
    api-service/            client-b/               my-project/
    state.db                state.db                state.db
```

### Key Design Decisions

**1. RepositoryCoordinator handles project creation:**
- When `Save()` is called for a new project, coordinator:
  - Uses `DirectoryManager.EnsureProjectDir()` to create directory
  - Updates master config via `ConfigLoader.Save()`
  - Saves project data to per-project DB
- This means `add.go` doesn't need to change much - just call `repo.Save()` as before

**2. Remove command uses DirectoryManager.DeleteProjectDir:**
- `RepositoryCoordinator.Delete()` removes project from DB and config
- But it does NOT delete the physical directory (by design - separation of concerns)
- `remove.go` calls `DirectoryManager.DeleteProjectDir()` after repo delete
- Deletion failure is NON-FATAL - log warning and continue (project removed from tracking)

**3. Two types of collision handling (BOTH needed):**
- **DirectoryManager (filesystem):** Resolves `~/.vibe-dash/api-service/` vs `~/.vibe-dash/client-b-api-service/` when two projects have same basename
- **add.go (display names):** Resolves user-facing display names in dashboard to avoid confusion
- These are DIFFERENT concerns - do NOT remove either one

**4. DirectoryManager needs config lookup:**
- `DirectoryManager` uses `ports.ProjectPathLookup` to find existing mappings
- Create `configPathAdapter` in `main.go` (unexported type) that wraps `ViperLoader`:
  ```go
  type configPathAdapter struct {
      loader *config.ViperLoader
  }

  func (a *configPathAdapter) GetDirForPath(path string) string {
      cfg, _ := a.loader.Load(context.Background())
      dirName, _ := cfg.GetDirectoryName(path)
      return dirName
  }
  ```

**5. Test strategy:**
- Most existing tests should pass unchanged (they mock `ports.ProjectRepository`)
- Add mock for `ports.DirectoryManager` to test remove command
- Use `t.TempDir()` for filesystem tests

### Current main.go Analysis

Current `main.go` (lines 18-67):
```go
func run(ctx context.Context) error {
    loader := config.NewViperLoader("")
    cfg, _ := loader.Load(ctx)

    // OLD: Single database
    repo, err := sqlite.NewSQLiteRepository("")

    // OLD: Set repository directly
    cli.SetRepository(repo)

    // ... detection service setup ...
}
```

**Required Changes:**
```go
// configPathAdapter implements ports.ProjectPathLookup for DirectoryManager.
// Defined in main.go as unexported type.
type configPathAdapter struct {
    loader *config.ViperLoader
}

func (a *configPathAdapter) GetDirForPath(path string) string {
    cfg, err := a.loader.Load(context.Background())
    if err != nil {
        return ""
    }
    dirName, _ := cfg.GetDirectoryName(path)
    return dirName
}

func run(ctx context.Context) error {
    loader := config.NewViperLoader("")
    cfg, _ := loader.Load(ctx) // Intentionally ignore error - graceful degradation

    // Store config for later use (unchanged from before)
    slog.Debug("config loaded", ...)

    // NEW: Get base path with safety check
    basePath := config.GetDefaultBasePath()
    if basePath == "" {
        return fmt.Errorf("failed to determine base path: cannot access home directory")
    }

    // NEW: Create config adapter for DirectoryManager
    configAdapter := &configPathAdapter{loader: loader}

    // NEW: Create DirectoryManager with nil check
    dirMgr := filesystem.NewDirectoryManager(basePath, configAdapter)
    if dirMgr == nil {
        return fmt.Errorf("failed to initialize directory manager: cannot determine base path")
    }

    // NEW: Create RepositoryCoordinator
    coordinator := persistence.NewRepositoryCoordinator(loader, dirMgr, basePath)

    // Set repository (coordinator implements ports.ProjectRepository)
    cli.SetRepository(coordinator)

    // NEW: Set DirectoryManager for remove command
    cli.SetDirectoryManager(dirMgr)

    // NEW: Set basePath for remove command
    cli.SetBasePath(basePath)

    // Initialize detection service with registry (unchanged)
    registry := detectors.NewRegistry()
    registry.Register(speckit.NewSpeckitDetector())
    detectionSvc := services.NewDetectionService(registry)
    cli.SetDetectionService(detectionSvc)

    return cli.Execute(ctx)
}
```

### Files to Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/config/paths.go` | CREATE | Add `GetDefaultBasePath()` helper |
| `internal/core/ports/directory.go` | MODIFY | Add `DeleteProjectDir` to interface |
| `internal/adapters/filesystem/directory.go` | MODIFY | Implement `DeleteProjectDir` method |
| `cmd/vibe/main.go` | MODIFY | Wire RepositoryCoordinator, DirectoryManager, add configPathAdapter |
| `internal/adapters/cli/deps.go` | CREATE | New package-level variables for directoryManager and basePath |
| `internal/adapters/cli/remove.go` | MODIFY | Add directory deletion after DB delete |
| `internal/adapters/cli/add.go` | VERIFY | Should work as-is with coordinator (keep display name collision handling) |
| `internal/adapters/cli/list.go` | VERIFY | Should work as-is with coordinator |

### Files to Create

| File | Purpose |
|------|---------|
| `internal/config/paths.go` | `GetDefaultBasePath()` helper function |
| `internal/config/paths_test.go` | Tests for paths.go |
| `internal/adapters/cli/deps.go` | Shared CLI dependencies (SetDirectoryManager, SetBasePath) |
| `internal/adapters/cli/mocks_test.go` | Mock for DirectoryManager in tests |
| `internal/adapters/filesystem/directory_test.go` | Tests for DeleteProjectDir (add to existing) |

### Implementation Patterns from Previous Stories

**Context Cancellation (REQUIRED at start of every public method):**
```go
select {
case <-ctx.Done():
    return ctx.Err()
default:
}
```

**Error Wrapping:**
```go
return fmt.Errorf("failed to delete project directory: %w", err)
```

**Directory Deletion Safety:**
```go
// Only delete if within base path (prevent accidental deletion of user files)
if strings.HasPrefix(dirPath, basePath) {
    if err := os.RemoveAll(dirPath); err != nil {
        slog.Warn("failed to delete project directory", "path", dirPath, "error", err)
        // Non-fatal - project removed from tracking, directory left behind
    }
}
```

### Existing Code Context

**CLI package-level variables (from `add.go:22-38`):**
```go
var repository ports.ProjectRepository
var detectionService ports.Detector

func SetRepository(repo ports.ProjectRepository)
func SetDetectionService(svc ports.Detector)
```

**New variables in `cli/deps.go` (to create):**
```go
var directoryManager ports.DirectoryManager
var basePath string

func SetDirectoryManager(dm ports.DirectoryManager)
func SetBasePath(path string)
```

**remove.go current implementation (lines 82-132):**
- Uses `findProjectByName()` to locate project
- Calls `repository.Delete(ctx, project.ID)`
- Does NOT delete physical directory

**Required addition in runRemove() after repository.Delete():**
```go
// Delete from repository (removes from DB and config)
if err := repository.Delete(ctx, project.ID); err != nil {
    return fmt.Errorf("failed to remove project: %w", err)
}

// NEW: Delete project directory using DirectoryManager
// This is NON-FATAL - project is already removed from tracking
if directoryManager != nil {
    if err := directoryManager.DeleteProjectDir(ctx, project.Path); err != nil {
        slog.Warn("failed to delete project directory", "path", project.Path, "error", err)
        // Continue - project removed from tracking, directory left behind
    }
}

fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed: %s\n", displayName)
return nil
```

**Required imports for remove.go:**
```go
import (
    // ... existing imports ...
    "log/slog"
)
```

### Thread Safety Considerations

- CLI commands run sequentially (single user interaction)
- `RepositoryCoordinator` handles thread safety internally
- No additional locking needed in CLI layer

### Error Handling

| Scenario | Error Type | Handling |
|----------|------------|----------|
| Coordinator creation fails | Fatal | Return error, exit |
| Add project fails | Domain error | Map to exit code |
| Remove project fails | Domain error | Map to exit code |
| Directory deletion fails | Warning | Log warning, continue (non-fatal) |
| List projects fails | Domain error | Map to exit code |

### Testing Strategy

**Unit Tests:**
- Mock `ports.ProjectRepository` (RepositoryCoordinator implements this)
- Mock `ports.DirectoryManager` for remove tests
- Test each command in isolation

**Integration Tests:**
- Use `t.TempDir()` for real filesystem
- Test full lifecycle: add → list → remove
- Verify directory structure created/deleted

**Manual Testing Checklist:**
1. Fresh start: `rm -rf ~/.vibe-dash/`
2. Add project: `vibe add .`
   - Verify `~/.vibe-dash/<project>/` created
   - Verify `state.db` exists
   - Verify `.project-path` marker exists
3. Add collision: `vibe add /other/path/with/same/basename`
   - Verify unique directory created with parent prefix
4. List: `vibe list`
   - Verify both projects shown
5. Remove: `vibe remove <name>`
   - Verify project removed from list
   - Verify directory deleted from filesystem
6. Final check: `ls ~/.vibe-dash/`
   - Only remaining project(s) should exist

### Dependencies

**Depends on (COMPLETED):**
- Story 3.5.1: `DirectoryManager` - Creates directories with `.project-path` marker
- Story 3.5.2: `sqlite.ProjectRepository` - Per-project database
- Story 3.5.3: Per-Project Config Files - Project config cascade
- Story 3.5.4: Master Config as Path Index - `config.GetDirectoryName()`
- Story 3.5.5: `RepositoryCoordinator` - Aggregates multiple DBs

**Required by:**
- Story 3.5.7: Integration Testing - Full lifecycle tests

### Previous Story Learnings (3.5.5)

From Story 3.5.5 code review - patterns to follow:

1. **Delete must update config** - Coordinator already does this
2. **Context cancellation** at start of every method
3. **Graceful degradation** - log warnings, continue on non-fatal errors
4. **Safety checks** for file operations (prevent accidental deletion)

### Project Structure Notes

After this story, the wiring is complete:
```
main.go
    │
    ├── ConfigLoader (ViperLoader)
    │       └── Reads/writes ~/.vibe-dash/config.yaml
    │
    ├── DirectoryManager (FilesystemDirectoryManager)
    │       └── Creates/manages ~/.vibe-dash/<project>/
    │
    └── RepositoryCoordinator
            ├── Uses ConfigLoader to enumerate projects
            ├── Uses DirectoryManager to create new project dirs
            └── Delegates to per-project ProjectRepository
```

### References

- [Source: docs/sprint-artifacts/stories/epic-3.5/epic-3.5-storage-structure.md#Story 3.5.6]
- [Source: docs/architecture.md#lines 798-807] - CLI adapter structure
- [Source: docs/project-context.md#Go Patterns] - Context first, error wrapping
- [Source: internal/adapters/cli/add.go] - Current add command implementation
- [Source: internal/adapters/cli/remove.go] - Current remove command implementation
- [Source: internal/adapters/persistence/coordinator.go] - RepositoryCoordinator
- [Source: internal/adapters/filesystem/directory.go] - DirectoryManager
- [Source: internal/config/loader.go] - ViperLoader
- [Source: cmd/vibe/main.go] - Current wiring

## Dev Agent Record

### Context Reference

- Epic 3.5: Storage Structure Alignment
- Story Dependencies: Depends on 3.5.1-3.5.5; required by 3.5.7
- PRD Reference: Lines 597-605 (Storage structure)
- Architecture Reference: Lines 798-807 (CLI adapter)
- Project Context: CLI commands, Go patterns

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - All tests pass, no debugging required.

### Completion Notes List

1. Created `GetDefaultBasePath()` helper in `internal/config/paths.go` to provide consistent base path lookup
2. Added `DeleteProjectDir` method to `ports.DirectoryManager` interface for project directory cleanup
3. Implemented `DeleteProjectDir` in `FilesystemDirectoryManager` with safety checks (only deletes within base path)
4. Updated `main.go` to wire:
   - `RepositoryCoordinator` (replaces old `sqlite.NewSQLiteRepository`)
   - `DirectoryManager` for directory operations
   - `configPathAdapter` to bridge ViperLoader to ProjectPathLookup interface
5. Created `internal/adapters/cli/deps.go` with `SetDirectoryManager()` setter
6. Updated `remove.go` to call `DirectoryManager.DeleteProjectDir()` after `repository.Delete()` - non-fatal if fails
7. Verified `add.go` and `list.go` work unchanged with `RepositoryCoordinator`
8. Created `MockDirectoryManager` in `mocks_test.go` for CLI testing
9. Added tests: `TestRemove_CallsDeleteProjectDir`, `TestRemove_DirectoryDeletionErrorIsNonFatal`, `TestRemove_NoDirectoryManagerIsGraceful`
10. Fixed `mockDirectoryManager` in `coordinator_test.go` to implement new `DeleteProjectDir` method
11. **[Code Review Fix]** Removed unused `basePath` variable and `SetBasePath()` from `deps.go` - not used by any CLI command

### File List

| File | Action | Description |
|------|--------|-------------|
| `internal/config/paths.go` | CREATED | `GetDefaultBasePath()` helper function |
| `internal/config/paths_test.go` | CREATED | Tests for `GetDefaultBasePath` |
| `internal/core/ports/directory.go` | MODIFIED | Added `DeleteProjectDir` to interface |
| `internal/adapters/filesystem/directory.go` | MODIFIED | Implemented `DeleteProjectDir` method |
| `internal/adapters/filesystem/directory_test.go` | MODIFIED | Added tests for `DeleteProjectDir` |
| `cmd/vibe/main.go` | MODIFIED | New wiring with RepositoryCoordinator, DirectoryManager, configPathAdapter |
| `internal/adapters/cli/deps.go` | CREATED | SetDirectoryManager function (SetBasePath removed after code review) |
| `internal/adapters/cli/remove.go` | MODIFIED | Added DirectoryManager.DeleteProjectDir call after repo delete |
| `internal/adapters/cli/mocks_test.go` | CREATED | MockDirectoryManager for testing |
| `internal/adapters/cli/remove_test.go` | MODIFIED | Added tests for directory deletion behavior |
| `internal/adapters/persistence/coordinator_test.go` | MODIFIED | Added DeleteProjectDir to mockDirectoryManager |
