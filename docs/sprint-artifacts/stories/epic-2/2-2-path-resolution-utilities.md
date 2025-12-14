# Story 2.2: Path Resolution Utilities

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | New file `internal/adapters/filesystem/paths.go` |
| **Key Dependencies** | stdlib only: `path/filepath`, `os`, `strings` |
| **Files to Create** | paths.go, paths_test.go |
| **Location** | internal/adapters/filesystem/ |
| **Domain Error** | `domain.ErrPathNotAccessible` (line 11 in domain/errors.go) |

### Quick Task Summary (4 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Implement ResolvePath | Absolute path resolution with existence check |
| 2 | Implement CanonicalPath | Symlink resolution via filepath.EvalSymlinks() |
| 3 | Implement ExpandHome | `~` and `~/` expansion only (NOT `~user`) |
| 4 | Tests + validation | 15 test cases including edge cases |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Package location | `internal/adapters/filesystem/` | Per architecture, OS abstraction layer |
| Path functions | Standalone functions (no struct) | No state needed, simpler testing |
| Error handling | Wrap with `domain.ErrPathNotAccessible` | Consistent domain errors |
| Symlink resolution | `filepath.EvalSymlinks()` | Stdlib canonical path |
| Home expansion | `os.UserHomeDir()` | Cross-platform support |
| `~user` syntax | **NOT SUPPORTED** | MVP simplicity; treat `~foo` as `~/foo` |

## Story

**As a** developer,
**I want** canonical path resolution,
**So that** symlinks and relative paths are handled correctly.

## Acceptance Criteria

```gherkin
AC1: Given absolute path "/home/user/project"
     When resolving path
     Then same absolute path is returned (after existence check)

AC2: Given relative path "."
     When resolving from /home/user/project
     Then "/home/user/project" is returned

AC3: Given "~/project"
     When resolving (user home is /home/user)
     Then "/home/user/project" is returned

AC4: Given symlink "/home/user/link" -> "/home/user/actual"
     When resolving symlink path via CanonicalPath
     Then canonical path "/home/user/actual" is returned

AC5: Given non-existent path "/invalid/path"
     When resolving
     Then domain.ErrPathNotAccessible is returned

AC6: Given empty string ""
     When resolving
     Then domain.ErrPathNotAccessible is returned

AC7: Given two symlinks pointing to same physical location
     When resolving both via CanonicalPath
     Then both return identical canonical paths (collision detection)
```

## Tasks / Subtasks

- [x] **Task 1: Implement ResolvePath function** (AC: 1, 2, 5, 6)
  - [x] 1.1 Create `internal/adapters/filesystem/paths.go`
  - [x] 1.2 Add godoc comment explaining function behavior
  - [x] 1.3 Handle empty string input -> return ErrPathNotAccessible
  - [x] 1.4 Call ExpandHome() first for ~ handling
  - [x] 1.5 Convert to absolute path using filepath.Abs()
  - [x] 1.6 Verify path exists using os.Stat()
  - [x] 1.7 Return domain.ErrPathNotAccessible if path doesn't exist

- [x] **Task 2: Implement CanonicalPath function** (AC: 4, 7)
  - [x] 2.1 Add godoc comment explaining symlink resolution
  - [x] 2.2 Call ResolvePath() first (validates existence)
  - [x] 2.3 Use filepath.EvalSymlinks() for symlink resolution
  - [x] 2.4 Return domain.ErrPathNotAccessible on error

- [x] **Task 3: Implement ExpandHome helper** (AC: 3)
  - [x] 3.1 Add godoc comment noting `~user` is NOT supported
  - [x] 3.2 Return original path unchanged if no ~ prefix
  - [x] 3.3 Handle "~" alone -> return home directory
  - [x] 3.4 Handle "~/" prefix -> replace with home + rest
  - [x] 3.5 Handle "~foo" (no slash) -> treat as "~/foo" (documented limitation)
  - [x] 3.6 Wrap os.UserHomeDir() error with domain.ErrPathNotAccessible

- [x] **Task 4: Write Tests and Validation** (AC: all)
  - [x] 4.1 Create `internal/adapters/filesystem/paths_test.go`
  - [x] 4.2 Test: ResolvePath with empty string returns ErrPathNotAccessible
  - [x] 4.3 Test: ResolvePath with absolute path (existing)
  - [x] 4.4 Test: ResolvePath with "." returns current directory
  - [x] 4.5 Test: ResolvePath with relative path
  - [x] 4.6 Test: ResolvePath with non-existent path returns ErrPathNotAccessible
  - [x] 4.7 Test: CanonicalPath resolves symlinks (t.Skip if not supported)
  - [x] 4.8 Test: CanonicalPath with regular path (no symlink)
  - [x] 4.9 Test: CanonicalPath with non-existent path returns error
  - [x] 4.10 Test: ExpandHome with "~" only
  - [x] 4.11 Test: ExpandHome with "~/" prefix
  - [x] 4.12 Test: ExpandHome with "~foo" (no slash) - document as ~/foo
  - [x] 4.13 Test: ExpandHome with no ~ returns original
  - [x] 4.14 Test: ExpandHome with ~ in middle (not prefix) returns unchanged
  - [x] 4.15 Test: Two symlinks to same location produce same canonical path
  - [x] 4.16 Run `make build`, `make lint`, `make test`

## Dev Notes

### CRITICAL: Error Handling Pattern

All path errors MUST return `domain.ErrPathNotAccessible`:
```go
// Good - wraps with domain error
return "", fmt.Errorf("%w: %s", domain.ErrPathNotAccessible, path)

// Bad - returns raw error
return "", err
```

### CRITICAL: `~user` Syntax NOT Supported

Unix `~user` syntax (e.g., `~bob/project` -> `/home/bob/project`) requires user lookup via `os/user` package. For MVP simplicity:
- `~foo` is treated as `~/foo` (current user's home + "foo")
- Document this in godoc
- Full `~user` support can be added post-MVP if needed

### Platform Notes

| Platform | Behavior |
|----------|----------|
| Linux/macOS | Full support including symlinks |
| Windows | `~` expansion works; symlinks may require admin privileges |

Symlink tests use `t.Skip()` if platform doesn't support them.

### Implementation Pattern

```go
package filesystem

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ResolvePath converts any path to absolute, verifying existence.
// Returns domain.ErrPathNotAccessible if path is empty, doesn't exist, or can't be accessed.
func ResolvePath(path string) (string, error) {
    if path == "" {
        return "", fmt.Errorf("%w: empty path", domain.ErrPathNotAccessible)
    }

    expanded, err := ExpandHome(path)
    if err != nil {
        return "", err // Already wrapped with domain error
    }

    absPath, err := filepath.Abs(expanded)
    if err != nil {
        return "", fmt.Errorf("%w: %s", domain.ErrPathNotAccessible, err)
    }

    if _, err := os.Stat(absPath); err != nil {
        return "", fmt.Errorf("%w: %s", domain.ErrPathNotAccessible, path)
    }

    return absPath, nil
}

// CanonicalPath resolves symlinks to get the "true" physical path.
// Used for collision detection (same physical location via different paths).
// Returns domain.ErrPathNotAccessible if path doesn't exist or can't be resolved.
func CanonicalPath(path string) (string, error) {
    resolved, err := ResolvePath(path)
    if err != nil {
        return "", err
    }

    canonical, err := filepath.EvalSymlinks(resolved)
    if err != nil {
        return "", fmt.Errorf("%w: symlink resolution failed for %s", domain.ErrPathNotAccessible, path)
    }

    return canonical, nil
}

// ExpandHome expands ~ prefix to user's home directory.
// NOTE: ~user syntax (e.g., ~bob) is NOT supported - treated as ~/user.
// Returns original path unchanged if no ~ prefix.
// Returns domain.ErrPathNotAccessible if home directory cannot be determined.
func ExpandHome(path string) (string, error) {
    if !strings.HasPrefix(path, "~") {
        return path, nil
    }

    home, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("%w: cannot determine home directory", domain.ErrPathNotAccessible)
    }

    if path == "~" {
        return home, nil
    }

    if strings.HasPrefix(path, "~/") {
        return filepath.Join(home, path[2:]), nil
    }

    // ~foo (no slash) -> treated as ~/foo (documented limitation)
    return filepath.Join(home, path[1:]), nil
}
```

### Test Pattern (Table-Driven)

```go
func TestExpandHome(t *testing.T) {
    home, err := os.UserHomeDir()
    if err != nil {
        t.Fatalf("os.UserHomeDir() error = %v", err)
    }

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"tilde only", "~", home, false},
        {"tilde slash", "~/project", filepath.Join(home, "project"), false},
        {"tilde no slash", "~foo", filepath.Join(home, "foo"), false}, // Documented: ~foo = ~/foo
        {"no tilde", "/absolute/path", "/absolute/path", false},
        {"tilde in middle", "/path/~/here", "/path/~/here", false},
        {"empty string", "", "", false}, // ExpandHome doesn't validate empty
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ExpandHome(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExpandHome() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, got, tt.want)
            }
        })
    }
}

func TestResolvePath_EmptyString(t *testing.T) {
    _, err := ResolvePath("")
    if !errors.Is(err, domain.ErrPathNotAccessible) {
        t.Errorf("ResolvePath(\"\") error = %v, want ErrPathNotAccessible", err)
    }
}

func TestCanonicalPath_Symlink(t *testing.T) {
    tmpDir := t.TempDir()
    actualDir := filepath.Join(tmpDir, "actual")
    if err := os.Mkdir(actualDir, 0755); err != nil {
        t.Fatalf("Failed to create dir: %v", err)
    }

    linkPath := filepath.Join(tmpDir, "link")
    if err := os.Symlink(actualDir, linkPath); err != nil {
        t.Skipf("Symlinks not supported: %v", err)
    }

    canonical, err := CanonicalPath(linkPath)
    if err != nil {
        t.Fatalf("CanonicalPath() error = %v", err)
    }
    if canonical != actualDir {
        t.Errorf("CanonicalPath() = %v, want %v", canonical, actualDir)
    }
}
```

### Previous Story Learnings (Story 2.1)

1. **Use t.TempDir()** - Automatic cleanup for test isolation
2. **t.Skip for platform features** - Skip symlink tests on unsupported platforms
3. **Table-driven tests** - Use `tests := []struct{}` for similar test cases
4. **Wrap errors with context** - `fmt.Errorf("failed to X: %w", err)`
5. **Return domain errors** - Use `domain.ErrPathNotAccessible` not raw errors

### Future Usage (Story 2.6 - Collision Detection)

```go
// In project_service.go (future story)
canonical, err := filesystem.CanonicalPath(path)
if err != nil {
    return err
}
existing, err := s.repo.FindByPath(ctx, canonical)
if err == nil {
    return fmt.Errorf("%w: already tracked as %s", domain.ErrProjectAlreadyExists, existing.Name)
}
```

### Files to Create

| File | Purpose |
|------|---------|
| `internal/adapters/filesystem/paths.go` | 3 exported functions with godoc |
| `internal/adapters/filesystem/paths_test.go` | 16 test cases |

**Note:** Do NOT create `platform.go` in this story. It belongs to Story 4.1 (FileWatcher).

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 2.2 requirements)
- docs/architecture.md (filesystem adapter structure, lines 805-808)
- docs/project-context.md (Go patterns, error handling)
- internal/core/domain/errors.go (ErrPathNotAccessible - line 11)
- docs/sprint-artifacts/2-1-sqlite-repository-setup.md (Previous story learnings)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Validation Applied

Story validated by SM agent on 2025-12-13:
- Added empty string test case (AC6)
- Documented `~user` syntax as unsupported (critical gap fix)
- Removed platform.go from scope (not needed for this story)
- Added Windows platform behavior notes
- Wrapped ExpandHome error with domain.ErrPathNotAccessible for consistency
- Consolidated task list from 5 to 4 tasks
- Increased test count from 14 to 16 (added empty string and ~foo tests)
- Removed redundant reference tables and anti-pattern sections
- Optimized implementation examples for LLM token efficiency

### Implementation Plan

Followed red-green-refactor cycle:
1. Created test file with 15 test cases (RED phase - tests failed due to missing implementation)
2. Implemented ResolvePath, CanonicalPath, and ExpandHome functions (GREEN phase - tests pass)
3. Fixed platform-specific test assertions for macOS `/var` -> `/private/var` symlink (REFACTOR phase)

### Completion Notes

All acceptance criteria satisfied:
- AC1: Absolute paths returned unchanged after existence check
- AC2: Relative paths resolved to absolute via filepath.Abs()
- AC3: `~` and `~/` expansion works via os.UserHomeDir()
- AC4: Symlinks resolved via filepath.EvalSymlinks()
- AC5: Non-existent paths return domain.ErrPathNotAccessible
- AC6: Empty string returns domain.ErrPathNotAccessible
- AC7: Two symlinks to same location produce identical canonical paths

Test Coverage: 15 test cases covering all ACs plus edge cases (empty string, tilde in middle, ~foo without slash).

### File List

**Files Created:**
- `internal/adapters/filesystem/paths.go` - Path resolution functions (ResolvePath, CanonicalPath, ExpandHome)
- `internal/adapters/filesystem/paths_test.go` - 15 test cases

**Files Modified:**
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status
- `docs/architecture.md` - Added paths_test.go to file structure (code review fix)

### Change Log

- 2025-12-13: Story 2.2 drafted by SM agent
- 2025-12-13: Story validated and improved - ready for development
- 2025-12-13: Story 2.2 implemented by Dev agent - all 4 tasks complete, 15 tests pass
- 2025-12-13: Code review completed - 4 medium, 2 low issues found and fixed
