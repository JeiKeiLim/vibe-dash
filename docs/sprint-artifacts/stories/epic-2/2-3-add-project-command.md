# Story 2.3: Add Project Command

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | New file `internal/adapters/cli/add.go` |
| **Key Dependencies** | `filesystem.CanonicalPath`, `domain.NewProject`, `sqlite.SQLiteRepository` |
| **Files to Create** | add.go, add_test.go |
| **Location** | internal/adapters/cli/ |
| **Domain Errors** | `domain.ErrPathNotAccessible`, `domain.ErrProjectAlreadyExists` |
| **Exit Codes** | 0 (success), 1 (general error - already tracked or path not found) |

### Quick Task Summary (5 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Create add command structure | Cobra command with path argument |
| 2 | Implement path resolution | Use `filesystem.CanonicalPath` from Story 2.2 |
| 3 | Implement collision detection | Check `repo.FindByPath()` before save |
| 4 | Implement project creation and save | Use `domain.NewProject` + `repo.Save()` |
| 5 | Tests + validation | 10+ test cases including edge cases |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Package location | `internal/adapters/cli/` | Per architecture, CLI adapter layer |
| Path resolution | Call `filesystem.CanonicalPath()` | Reuse Story 2.2 utilities, handle symlinks |
| Default path | Current directory (`.`) | FR1: `vibe add .` should work |
| Collision check | `repo.FindByPath()` | Canonical path ensures symlinks detected |
| Custom name | `--name` flag | FR5: Optional display name |
| Output format | Success: `✓ Added: name`, Error: `✗ Message` | UX feedback patterns |
| Detection | **NOT** in this story | Speckit detector (Story 2.4) needed first |

## Story

**As a** user,
**I want** to add projects using `vibe add`,
**So that** they appear in my dashboard.

## Acceptance Criteria

```gherkin
AC1: Given I am in a project directory
     When I run `vibe add .`
     Then project is added with:
       - Name derived from directory name
       - Canonical path stored
       - State = Active
     And confirmation message shows:
       "✓ Added: project-name"
       "  Path: /canonical/path"

AC2: Given an absolute path /path/to/project
     When I run `vibe add /path/to/project`
     Then project at specified path is added

AC3: Given project at /path already tracked
     When I run `vibe add /path` (or symlink to same location)
     Then error message shows:
       "✗ Project already tracked: project-name"
     And exit code is 1 (via domain.ErrProjectAlreadyExists → ExitGeneralError)

AC4: Given path /invalid/path doesn't exist
     When I run `vibe add /invalid/path`
     Then error message shows:
       "✗ Path not found: /invalid/path"
     And exit code is 1 (via domain.ErrPathNotAccessible → ExitGeneralError)

AC5: Given I want a custom display name
     When I run `vibe add . --name "Custom Name"`
     Then project is added with display_name = "Custom Name"
     And confirmation shows:
       "✓ Added: Custom Name"
       "  Path: /canonical/path"

AC6: Given I run `vibe add` with no arguments
     When executed
     Then current directory (`.`) is used as default path

AC7: Given path "~" (tilde)
     When I run `vibe add ~`
     Then home directory is added correctly
```

## Tasks / Subtasks

- [x] **Task 1: Create add command structure** (AC: 1, 2, 6)
  - [x] 1.1 Create `internal/adapters/cli/add.go`
  - [x] 1.2 Define `addCmd` with cobra.Command:
    - Use: `add [path]`
    - Short: `Add a project to tracking`
    - Args: `cobra.MaximumNArgs(1)` (default to `.` if none)
  - [x] 1.3 Add `--name` flag (string, optional) for display name (AC5)
  - [x] 1.4 Register addCmd as subcommand of RootCmd in init()
  - [x] 1.5 Inject repository dependency via package-level variable or closure

- [x] **Task 2: Implement path resolution** (AC: 1, 2, 4, 6, 7)
  - [x] 2.1 Get path argument (default to `.` if not provided)
  - [x] 2.2 Call `filesystem.CanonicalPath(path)` for symlink resolution
  - [x] 2.3 Handle `domain.ErrPathNotAccessible`:
    - Return wrapped error (Cobra/main.go handles exit code via `MapErrorToExitCode`)

- [x] **Task 3: Implement collision detection** (AC: 3)
  - [x] 3.1 Call `repo.FindByPath(ctx, canonicalPath)`
  - [x] 3.2 If project exists (err == nil):
    - Return `fmt.Errorf("%w: %s", domain.ErrProjectAlreadyExists, existing.Name)`
    - (Cobra/main.go handles exit code via `MapErrorToExitCode`)
  - [x] 3.3 If err == `domain.ErrProjectNotFound`: proceed (expected)
  - [x] 3.4 If other error: return wrapped error

- [x] **Task 4: Create and save project** (AC: 1, 2, 5)
  - [x] 4.1 Call `domain.NewProject(canonicalPath, "")` to create project
  - [x] 4.2 If `--name` flag provided, set `project.DisplayName = name`
  - [x] 4.3 Call `repo.Save(ctx, project)`
  - [x] 4.4 Print success message:
    - `✓ Added: {displayName or name}`
    - `  Path: {canonicalPath}`
  - [x] 4.5 Return exit code 0

- [x] **Task 5: Write Tests and Validation** (AC: all)
  - [x] 5.1 Create `internal/adapters/cli/add_test.go`
  - [x] 5.2 Test: add with `.` adds current directory
  - [x] 5.3 Test: add with absolute path
  - [x] 5.4 Test: add with `--name` flag sets DisplayName
  - [x] 5.5 Test: add non-existent path returns error + exit 1 (corrected from exit 2)
  - [x] 5.6 Test: add already-tracked path returns error + exit 1
  - [x] 5.7 Test: add symlink to already-tracked location returns error (collision)
  - [x] 5.8 Test: add `~` expands to home directory
  - [x] 5.9 Test: add with no args defaults to `.`
  - [x] 5.10 Test: verify project saved with correct fields
  - [x] 5.11 Run `make build`, `make lint`, `make test`

## Dev Notes

### CRITICAL: Use Repository Interface, Not Direct SQLite

```go
// Good - use interface for testability
type ProjectAdder struct {
    repo ports.ProjectRepository
}

// Bad - direct SQLite coupling
repo := sqlite.NewSQLiteRepository("")
```

### CRITICAL: Detection NOT Included in This Story

This story adds projects WITHOUT methodology detection. Detection is Story 2.4.
After add, projects have:
- DetectedMethod = "" (empty)
- CurrentStage = StageUnknown
- Confidence = "" (empty)

Detection will be added in Story 2.5 (Detection Service) which orchestrates Story 2.4's Speckit detector.

### Implementation Pattern

```go
package cli

import (
    "errors"
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Repository is injected at startup (cmd/vibe/main.go)
var Repository ports.ProjectRepository

var addCmd = &cobra.Command{
    Use:   "add [path]",
    Short: "Add a project to tracking",
    Long: `Add a project to the vibe-dash dashboard.

If no path is provided, the current directory is used.
The path is resolved to its canonical form (following symlinks).

Examples:
  vibe add .                  # Add current directory
  vibe add /path/to/project   # Add specific path
  vibe add . --name "My App"  # Add with custom display name`,
    Args: cobra.MaximumNArgs(1),
    RunE: runAdd,
}

var addName string

func init() {
    RootCmd.AddCommand(addCmd)
    addCmd.Flags().StringVar(&addName, "name", "", "Custom display name for the project")
}

func runAdd(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()

    // Get path (default to ".")
    path := "."
    if len(args) > 0 {
        path = args[0]
    }

    // Resolve to canonical path (handles ~, symlinks, existence check)
    canonicalPath, err := filesystem.CanonicalPath(path)
    if err != nil {
        // ErrPathNotAccessible will be mapped to exit code 1 by MapErrorToExitCode
        return fmt.Errorf("path not found: %s: %w", path, err)
    }

    // Check if already tracked (collision detection)
    existing, err := Repository.FindByPath(ctx, canonicalPath)
    if err == nil {
        // Project exists - return domain error for proper exit code mapping
        displayName := existing.Name
        if existing.DisplayName != "" {
            displayName = existing.DisplayName
        }
        return fmt.Errorf("%w: %s", domain.ErrProjectAlreadyExists, displayName)
    }
    if !errors.Is(err, domain.ErrProjectNotFound) {
        // Unexpected error
        return fmt.Errorf("failed to check existing project: %w", err)
    }
    // domain.ErrProjectNotFound is expected - continue

    // Create new project
    project, err := domain.NewProject(canonicalPath, "")
    if err != nil {
        return fmt.Errorf("failed to create project: %w", err)
    }

    // Set custom display name if provided
    if addName != "" {
        project.DisplayName = addName
    }

    // Save to repository
    if err := Repository.Save(ctx, project); err != nil {
        return fmt.Errorf("failed to save project: %w", err)
    }

    // Success output
    displayName := project.Name
    if project.DisplayName != "" {
        displayName = project.DisplayName
    }
    fmt.Printf("✓ Added: %s\n", displayName)
    fmt.Printf("  Path: %s\n", canonicalPath)

    return nil
}
```

**CRITICAL: Error Handling Flow**

Commands return errors, NOT call `os.Exit()`. The error-to-exit-code mapping happens in `main.go`:

```go
// cmd/vibe/main.go pattern
func main() {
    if err := cli.Execute(ctx); err != nil {
        fmt.Fprintf(os.Stderr, "✗ %v\n", err)
        os.Exit(cli.MapErrorToExitCode(err))
    }
}
```

### Exit Codes (per exitcodes.go)

| Scenario | Domain Error | Exit Code | Constant |
|----------|--------------|-----------|----------|
| Success | nil | 0 | ExitSuccess |
| Already tracked | ErrProjectAlreadyExists | 1 | ExitGeneralError |
| Path not found | ErrPathNotAccessible | 1 | ExitGeneralError |
| Project not found | ErrProjectNotFound | 2 | ExitProjectNotFound |
| Config invalid | ErrConfigInvalid | 3 | ExitConfigInvalid |

Reference: `internal/adapters/cli/exitcodes.go` - uses `MapErrorToExitCode(err)` function

### Test Pattern (Table-Driven with Mock Repository)

```go
package cli_test

import (
    "context"
    "os"
    "testing"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// MockRepository implements ports.ProjectRepository for testing
type MockRepository struct {
    projects map[string]*domain.Project
    saveErr  error
}

func NewMockRepository() *MockRepository {
    return &MockRepository{projects: make(map[string]*domain.Project)}
}

func (m *MockRepository) Save(ctx context.Context, project *domain.Project) error {
    if m.saveErr != nil {
        return m.saveErr
    }
    m.projects[project.Path] = project
    return nil
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
    for _, p := range m.projects {
        if p.ID == id {
            return p, nil
        }
    }
    return nil, domain.ErrProjectNotFound
}

func (m *MockRepository) FindByPath(ctx context.Context, path string) (*domain.Project, error) {
    if p, ok := m.projects[path]; ok {
        return p, nil
    }
    return nil, domain.ErrProjectNotFound
}

func (m *MockRepository) FindAll(ctx context.Context) ([]*domain.Project, error) {
    result := make([]*domain.Project, 0, len(m.projects))
    for _, p := range m.projects {
        result = append(result, p)
    }
    return result, nil
}

func (m *MockRepository) FindActive(ctx context.Context) ([]*domain.Project, error) {
    result := make([]*domain.Project, 0)
    for _, p := range m.projects {
        if p.State == domain.StateActive {
            result = append(result, p)
        }
    }
    return result, nil
}

func (m *MockRepository) FindHibernated(ctx context.Context) ([]*domain.Project, error) {
    result := make([]*domain.Project, 0)
    for _, p := range m.projects {
        if p.State == domain.StateHibernated {
            result = append(result, p)
        }
    }
    return result, nil
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
    for path, p := range m.projects {
        if p.ID == id {
            delete(m.projects, path)
            return nil
        }
    }
    return domain.ErrProjectNotFound
}

func (m *MockRepository) UpdateState(ctx context.Context, id string, state domain.ProjectState) error {
    for _, p := range m.projects {
        if p.ID == id {
            p.State = state
            return nil
        }
    }
    return domain.ErrProjectNotFound
}

func TestAdd_CurrentDirectory(t *testing.T) {
    // Setup mock
    mock := NewMockRepository()
    cli.Repository = mock

    // Create temp directory
    tmpDir := t.TempDir()

    // Change to temp directory
    oldWd, _ := os.Getwd()
    os.Chdir(tmpDir)
    defer os.Chdir(oldWd)

    // Execute add command via Cobra
    cmd := cli.RootCmd
    cmd.SetArgs([]string{"add", "."})
    err := cmd.Execute()

    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    // Verify project was saved
    if len(mock.projects) != 1 {
        t.Errorf("expected 1 project, got %d", len(mock.projects))
    }
}

func TestAdd_AlreadyTracked(t *testing.T) {
    tmpDir := t.TempDir()

    // Setup mock with existing project at tmpDir path
    mock := NewMockRepository()
    mock.projects[tmpDir] = &domain.Project{
        ID:   "abc123",
        Name: "existing-project",
        Path: tmpDir,
    }
    cli.Repository = mock

    // Execute add with same path
    cmd := cli.RootCmd
    cmd.SetArgs([]string{"add", tmpDir})
    err := cmd.Execute()

    // Should return ErrProjectAlreadyExists (mapped to exit code 1)
    if err == nil {
        t.Fatal("expected error for already tracked project")
    }
    if cli.MapErrorToExitCode(err) != cli.ExitGeneralError {
        t.Errorf("expected exit code %d, got %d", cli.ExitGeneralError, cli.MapErrorToExitCode(err))
    }
}
```

### Previous Story Learnings (Story 2.2)

1. **Use `filesystem.CanonicalPath()`** - Already handles ~, symlinks, existence check
2. **Use `domain.NewProject()`** - Validates path, derives name, generates ID
3. **Return domain errors** - Use `errors.Is()` for checking error types
4. **Table-driven tests** - Use `tests := []struct{}` for similar test cases
5. **t.TempDir()** - Automatic cleanup for test isolation

### Dependency Injection Strategy

The repository must be injectable for testing. Two options:

**Option A: Package variable (simpler, current pattern)**
```go
// In add.go
var Repository ports.ProjectRepository

// In main.go
func main() {
    repo, _ := sqlite.NewSQLiteRepository("")
    cli.Repository = repo
    cli.Execute(ctx)
}
```

**Option B: Closure injection (more explicit)**
```go
// In add.go
func NewAddCmd(repo ports.ProjectRepository) *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // Use repo directly
        },
    }
}
```

Use **Option A** for MVP simplicity (matches current root.go pattern).

### Files to Create

| File | Purpose |
|------|---------|
| `internal/adapters/cli/add.go` | Add command implementation |
| `internal/adapters/cli/add_test.go` | Tests with mock repository |

### Files to Modify

| File | Change |
|------|--------|
| `cmd/vibe/main.go` | Inject repository before `cli.Execute()` |

### Architecture Compliance Checklist

- [x] Command in `internal/adapters/cli/` (correct adapter layer)
- [x] Uses `ports.ProjectRepository` interface (not direct SQLite)
- [x] Uses `filesystem.CanonicalPath()` (reuses Story 2.2)
- [x] Uses `domain.NewProject()` (creates via constructor)
- [x] Returns domain errors (ErrPathNotAccessible, ErrProjectNotFound)
- [x] Follows exit code mapping (exitcodes.go)
- [x] Context propagation (uses `cmd.Context()`)
- [x] Error wrapping with context (`fmt.Errorf("...: %w", err)`)

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 2.3 requirements, lines 639-686)
- docs/architecture.md (CLI adapter structure, exit codes, hexagonal pattern)
- docs/project-context.md (Go patterns, error handling)
- internal/adapters/cli/root.go (current CLI structure)
- internal/adapters/cli/exitcodes.go (exit code mapping)
- internal/adapters/filesystem/paths.go (CanonicalPath function from Story 2.2)
- internal/adapters/persistence/sqlite/repository.go (repository implementation)
- internal/core/domain/project.go (NewProject constructor)
- internal/core/ports/repository.go (ProjectRepository interface)
- docs/sprint-artifacts/2-2-path-resolution-utilities.md (Previous story learnings)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Validation Notes

Story drafted by SM agent with comprehensive context:
- All 7 acceptance criteria mapped from epics.md
- Detection intentionally excluded (requires Stories 2.4, 2.5)
- Exit codes aligned with architecture document
- Test strategy includes mock repository pattern
- Dependency injection strategy documented
- Implementation example provided with full error handling

**Story validated by SM agent on 2025-12-13:**

Critical fixes applied:
1. **Exit code mapping corrected** - Story now aligns with `exitcodes.go`:
   - ErrProjectAlreadyExists → ExitGeneralError (1)
   - ErrPathNotAccessible → ExitGeneralError (1)
   - Removed incorrect exit code 2 for path not found
2. **Removed os.Exit() from implementation** - Commands now return errors per architecture pattern
   - Added main.go error handling flow documentation
3. **Mock repository completed** - All 8 interface methods now implemented
4. **Domain error usage** - Collision detection now returns `domain.ErrProjectAlreadyExists` for proper exit code mapping
5. **Test examples improved** - Tests now use actual Cobra execution and verify exit codes via `MapErrorToExitCode()`

### Implementation Notes (Dev Agent)

**Implemented by:** Claude Opus 4.5 (claude-opus-4-5-20251101)
**Date:** 2025-12-13

#### Implementation Summary

1. **Created `internal/adapters/cli/add.go`**: Full implementation of the add command with:
   - Cobra command structure with `add [path]` usage
   - `--name` flag for custom display names
   - Path resolution using `filesystem.CanonicalPath()`
   - Collision detection via `repo.FindByPath()`
   - Project creation using `domain.NewProject()`
   - Success/error output per UX patterns

2. **Created `internal/adapters/cli/add_test.go`**: 9 comprehensive tests:
   - `TestAdd_CurrentDirectory` - add with "."
   - `TestAdd_AbsolutePath` - add with absolute path
   - `TestAdd_WithNameFlag` - custom display name
   - `TestAdd_NonExistentPath` - error handling + exit code 1
   - `TestAdd_AlreadyTracked` - collision detection + exit code 1
   - `TestAdd_SymlinkCollision` - symlink to tracked path
   - `TestAdd_HomeDirectory` - tilde expansion
   - `TestAdd_NoArgs_DefaultsToCurrentDirectory` - default path
   - `TestAdd_VerifyProjectFields` - field validation

3. **Modified `cmd/vibe/main.go`**: Injected repository before `cli.Execute()` to wire production SQLite repository.

4. **Design decisions**:
   - Used Option A (package variable) for dependency injection for MVP simplicity
   - Added `SetRepository()` function for testability
   - Added `NewRootCmd()` and `RegisterAddCommand()` for isolated tests
   - Output written to `cmd.OutOrStdout()` for test capture

#### Completion Notes

- All 7 acceptance criteria satisfied
- All tests pass (10 tests covering all scenarios including Save() failure)
- `make build`, `make lint`, `make test` all pass
- Manual testing verified with actual binary:
  - `vibe add .` works
  - `vibe add /path` works
  - `vibe add . --name "Custom"` works
  - `vibe add ~` works
  - Already tracked returns exit code 1
  - Non-existent path returns exit code 1

#### Code Review Fixes (2025-12-13)

**Issues Fixed:**

1. **Flag state pollution** (MEDIUM): Added `ResetAddFlags()` function and call it in test helper to prevent `addName` flag value leaking between tests.

2. **Verbose error message** (MEDIUM): Removed redundant error wrapping in path not found case. Error now shows `path not accessible: path does not exist: /path` instead of `path not found: /path: path not accessible: path does not exist: /path`.

3. **Missing Save() failure test** (LOW): Added `TestAdd_SaveFailure` to verify error handling when repository.Save() fails.

**Issues Noted (Not Fixed):**

- MockRepository duplication: Could extract to shared test utilities, but acceptable for MVP
- Double command registration pattern: Works correctly, just verbose - not a bug

## File List

### Files Created

| File | Purpose |
|------|---------|
| `internal/adapters/cli/add.go` | Add command implementation |
| `internal/adapters/cli/add_test.go` | Tests with mock repository (10 tests) |

### Files Modified

| File | Change |
|------|--------|
| `cmd/vibe/main.go` | Added SQLite repository injection |

## Change Log

| Date | Change |
|------|--------|
| 2025-12-13 | Story implementation completed - add command with full test coverage |
| 2025-12-13 | Code review fixes applied: flag reset for tests, cleaner error messages, Save() failure test |
