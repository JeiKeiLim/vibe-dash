# Story 3.5.1: Directory Manager with Collision Handling

Status: done

## Story

As a user,
I want each project to have its own subdirectory with collision handling,
So that projects with the same name from different locations are stored separately.

## Acceptance Criteria

1. **AC1: Basic directory creation** - Given project at `/home/user/api-service`, when adding the project, then directory `~/.vibe-dash/api-service/` is created.

2. **AC2: First collision (parent disambiguation)** - Given project `api-service` already tracked and new project at `/home/user/client-b/api-service`, when adding the new project, then directory `~/.vibe-dash/client-b-api-service/` is created.

3. **AC3: Second collision (grandparent)** - Given collision still exists after parent disambiguation, when adding another project, then grandparent directory is added (e.g., `work-client-b-api-service`).

4. **AC4: Symlink handling** - Given project path with symlinks, when calculating directory name, then canonical path (via `filepath.EvalSymlinks`) is used.

5. **AC5: Deterministic naming** - Given same project path added twice, when calculating directory name, then same directory name is returned (deterministic).

6. **AC6: Error handling** - Given directory creation fails (permission denied, disk full), when `EnsureProjectDir` is called, then descriptive error is returned with path and cause.

7. **AC7: Directory name normalization** - Given project path with special characters, when calculating directory name, then characters are normalized (spaces/special chars → hyphens, lowercase).

8. **AC8: Max recursion depth** - Given collision persists after 10 parent levels, when calculating directory name, then error is returned (unresolvable collision).

## Tasks / Subtasks

- [x] Task 1: Create `DirectoryManager` interface in core ports (AC: 1, 5, 7)
  - [x] Subtask 1.1: Create `internal/core/ports/directory.go`
  - [x] Subtask 1.2: Define `DirectoryManager` interface with methods:
    - `GetProjectDirName(ctx context.Context, projectPath string) (string, error)` - deterministic directory name calculation
    - `EnsureProjectDir(ctx context.Context, projectPath string) (string, error)` - creates directory, returns full path
  - [x] Subtask 1.3: Interface must use ONLY stdlib types and domain types (zero adapter imports)

- [x] Task 2: Implement collision resolution algorithm (AC: 1, 2, 3, 4, 5, 7, 8)
  - [x] Subtask 2.1: Create `internal/adapters/filesystem/directory.go`
  - [x] Subtask 2.2: Create constructor `NewDirectoryManager(basePath string)` accepting injectable base path for testability
  - [x] Subtask 2.3: Implement `GetProjectDirName()` with recursive collision resolution:
    1. Resolve canonical path via `CanonicalPath()` from `paths.go`
    2. Extract base directory name
    3. Normalize: lowercase, replace non-alphanumeric/hyphen with hyphen
    4. Check if directory exists in `~/.vibe-dash/`
    5. If no collision: return name
    6. If collision: check if same canonical path → return existing name (determinism)
    7. If different project: prepend parent directory (normalized)
    8. Recurse up to 10 levels, then return error
  - [x] Subtask 2.4: Implement `EnsureProjectDir()` that creates `<basePath>/<name>/` directory atomically

- [x] Task 3: Implement collision detection (AC: 2, 3, 5)
  - [x] Subtask 3.1: Query master config for existing project → directory mappings (requires config loader integration point)
  - [x] Subtask 3.2: Iterate through parent directories until unique name found
  - [x] Subtask 3.3: Handle edge case where project is at filesystem root (use "root" as final segment)

- [x] Task 4: Write comprehensive unit tests (AC: 1-8)
  - [x] Subtask 4.1: Test basic directory name derivation
  - [x] Subtask 4.2: Test first collision (parent disambiguation)
  - [x] Subtask 4.3: Test second collision (grandparent)
  - [x] Subtask 4.4: Test deeply nested paths (5+ levels)
  - [x] Subtask 4.5: Test special characters in path (spaces, unicode, colons)
  - [x] Subtask 4.6: Test determinism (same path = same result)
  - [x] Subtask 4.7: Test error handling (permission denied, invalid path)
  - [x] Subtask 4.8: Test symlink resolution
  - [x] Subtask 4.9: Test max recursion depth exceeded
  - [x] Subtask 4.10: Test case sensitivity (macOS/Windows: `Api-Service` and `api-service`)
  - [x] Subtask 4.11: Test trailing slash handling

- [x] Task 5: Create integration test with real file system (AC: 1, 2, 6)
  - [x] Subtask 5.1: Create temporary directories for collision testing
  - [x] Subtask 5.2: Verify directory actually gets created
  - [x] Subtask 5.3: Test permission error handling with read-only directory

## Dev Notes

### Architecture Alignment

This story implements the `DirectoryManager` component specified in Architecture:

> **Project Directory Collision Handling** - Resolution Algorithm (Recursive):
> 1. Use project directory name: `~/.vibe-dash/api-service/`
> 2. On collision: prepend parent directory: `~/.vibe-dash/client-b-api-service/`
> 3. Still collision: prepend grandparent: `~/.vibe-dash/work-client-b-api-service/`
> 4. Continue up directory tree until unique (max 10 levels)
>
> [Source: docs/architecture.md:336-358]

### Hexagonal Architecture Boundaries

**CRITICAL:** Follow hexagonal architecture boundaries:

```
internal/core/ports/directory.go     → Interface definition (ZERO external imports)
internal/adapters/filesystem/directory.go → Implementation (can import stdlib + core)
```

The interface in `ports/` must NOT import anything from adapters. Use only:
- Standard library types (`context.Context`, `string`, `error`)
- Domain types from `internal/core/domain/`

[Source: docs/project-context.md:25-35]

### Interface Definition

```go
// DirectoryManager handles project directory naming and creation.
// Implements collision resolution per PRD specification.
type DirectoryManager interface {
    // GetProjectDirName returns deterministic directory name for project.
    // Uses collision resolution if name already exists for different project.
    // Returns error if path invalid or collision unresolvable after 10 levels.
    GetProjectDirName(ctx context.Context, projectPath string) (string, error)

    // EnsureProjectDir creates project directory if not exists.
    // Returns full path to created/existing directory.
    // Returns domain.ErrPathNotAccessible on permission/filesystem errors.
    EnsureProjectDir(ctx context.Context, projectPath string) (string, error)
}
```

Note: `ctx context.Context` first parameter required per project-context.md:45.

### Constructor Pattern

```go
// NewDirectoryManager creates DirectoryManager with configurable base path.
// basePath defaults to ~/.vibe-dash if empty string provided.
func NewDirectoryManager(basePath string, configLookup ProjectPathLookup) *FilesystemDirectoryManager
```

Injectable `basePath` enables test isolation. `ProjectPathLookup` interface provides existing mappings for determinism.

### Directory Name Normalization Rules

Apply before collision check:

1. Convert to lowercase
2. Replace spaces with hyphens: `my project` → `my-project`
3. Replace special characters (non-alphanumeric except hyphen) with hyphens
4. Collapse multiple consecutive hyphens: `my--project` → `my-project`
5. Trim leading/trailing hyphens

Example transformations:
- `/path/to/My Project` → `my-project`
- `/path/to/api:service` → `api-service`
- `/path/to/日本語プロジェクト` → Unicode preserved or transliterated per stdlib

### Determinism via Master Config Lookup

To achieve AC5 (determinism), the implementation needs to query existing project mappings:

```go
// ProjectPathLookup provides existing project → directory mappings.
// Used to ensure same project path always returns same directory name.
type ProjectPathLookup interface {
    // GetDirForPath returns existing directory name for canonical path.
    // Returns empty string if path not previously registered.
    GetDirForPath(canonicalPath string) string
}
```

This will be implemented by master config in Story 3.5.4. For Story 3.5.1, use a simple in-memory map for testing.

### Existing Code to Leverage

```go
// Use these from internal/adapters/filesystem/paths.go
CanonicalPath(path string) (string, error)  // Resolves symlinks + absolute path
ExpandHome(path string) (string, error)     // Expands ~ to home dir
```

[Source: internal/adapters/filesystem/paths.go:40-52]

### Error Handling Pattern

Follow established domain error pattern:

```go
import "github.com/JeiKeiLim/vibe-dash/internal/core/domain"

// Error format: %w: context message: underlying cause
return fmt.Errorf("%w: failed to create directory %s: %v", domain.ErrPathNotAccessible, path, err)

// New error for collision limit
var ErrCollisionUnresolvable = errors.New("directory name collision unresolvable")
```

[Source: internal/core/domain/errors.go]

### Edge Cases

| Case | Handling |
|------|----------|
| Filesystem root `/` | Use "root" as final segment |
| Single-level `/project` | No parent available; return "project" or error |
| Trailing slash `/path/` | Strip before processing |
| 10+ level collision | Return `ErrCollisionUnresolvable` |
| Case sensitivity | Normalize to lowercase; treats `Api-Service` and `api-service` as same |
| Concurrent creation | Use `os.MkdirAll` (atomic); check-then-create race handled |

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Interface | PascalCase | `DirectoryManager` |
| Method | PascalCase | `GetProjectDirName`, `EnsureProjectDir` |
| File | snake_case | `directory.go` |
| Test file | snake_case + _test | `directory_test.go` |
| Domain error | Err prefix | `ErrCollisionUnresolvable` |

### Project Structure Notes

New files to create:
- `internal/core/ports/directory.go` - Interface definition
- `internal/core/domain/errors.go` - Add `ErrCollisionUnresolvable` (if not exists)
- `internal/adapters/filesystem/directory.go` - Implementation
- `internal/adapters/filesystem/directory_test.go` - Unit tests

### Dependencies on This Story

This story is a foundational component for:
- **Story 3.5.2**: Per-Project SQLite Repository (needs `EnsureProjectDir` to create project directories)
- **Story 3.5.3**: Per-Project Config Files (needs directory path)
- **Story 3.5.5**: Repository Coordinator (needs to locate project directories)
- **Story 3.5.6**: Update CLI Commands (uses `DirectoryManager` for `vibe add`)

### Previous Story Learnings

From Story 3.5.0 (Cleanup Existing Storage):
- Existing tests may reference old storage structure
- TODO comments with `TODO(story:3.5.7)` mark tests needing updates
- Clean start approach - no migration logic needed

[Source: docs/sprint-artifacts/stories/epic-3.5/3-5-0-cleanup-existing-storage.md:159-166]

### Testing Strategy

Table-driven tests required per Architecture:

```go
func TestGetProjectDirName(t *testing.T) {
    tests := []struct {
        name        string
        projectPath string
        existing    map[string]string // canonical path → directory name
        expected    string
        expectError error
    }{
        {
            name:        "basic directory name",
            projectPath: "/home/user/api-service",
            existing:    map[string]string{},
            expected:    "api-service",
        },
        {
            name:        "first collision",
            projectPath: "/home/user/client-b/api-service",
            existing:    map[string]string{"/home/user/api-service": "api-service"},
            expected:    "client-b-api-service",
        },
        {
            name:        "special characters normalized",
            projectPath: "/home/user/My Project",
            existing:    map[string]string{},
            expected:    "my-project",
        },
        {
            name:        "deterministic - same path returns same name",
            projectPath: "/home/user/api-service",
            existing:    map[string]string{"/home/user/api-service": "api-service"},
            expected:    "api-service",
        },
        {
            name:        "max depth exceeded",
            projectPath: "/a/b/c/d/e/f/g/h/i/j/k/project", // 11+ levels of collision
            existing:    map[string]string{...}, // collisions at every level
            expectError: domain.ErrCollisionUnresolvable,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lookup := &mockLookup{paths: tt.existing}
            dm := NewDirectoryManager("", lookup)
            got, err := dm.GetProjectDirName(context.Background(), tt.projectPath)
            // assertions...
        })
    }
}
```

### Manual Testing

After implementation, verify with:

1. **Add first project:**
   ```bash
   vibe add /path/to/api-service
   ls ~/.vibe-dash/  # Should show: api-service/
   ```

2. **Add collision:**
   ```bash
   vibe add /other/path/api-service
   ls ~/.vibe-dash/  # Should show both directories with disambiguation
   ```

3. **Verify determinism:**
   ```bash
   # Add same project twice should return same directory name
   vibe add /path/to/api-service
   # Should not create duplicate
   ```

## Dev Agent Record

### Context Reference

- Epic 3.5: Storage Structure Alignment
- Story Dependencies: Foundation for Stories 3.5.2, 3.5.3, 3.5.5, 3.5.6
- PRD Reference: Lines 647-659 (Collision Handling)
- Architecture Reference: Lines 336-358 (Directory Manager)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

1. **Interface Design**: Created `DirectoryManager` interface in `ports/directory.go` with `GetProjectDirName` and `EnsureProjectDir` methods. Also added `ProjectPathLookup` interface for determinism support.

2. **Domain Error**: Added `ErrCollisionUnresolvable` to `domain/errors.go` for max depth exceeded scenarios.

3. **Implementation**: `FilesystemDirectoryManager` in `adapters/filesystem/directory.go` implements the full collision resolution algorithm:
   - Uses `CanonicalPath()` for symlink resolution
   - Normalizes names: lowercase, special chars → hyphens, collapse multiple hyphens
   - Iterates through parent directories on collision (max 10 levels)
   - Uses `.project-path` marker file for determinism check

4. **Determinism**: Two mechanisms ensure determinism:
   - `ProjectPathLookup` interface for external config integration
   - `.project-path` marker file written to each project directory

5. **Testing**: 35+ test cases covering all acceptance criteria including edge cases (symlinks, permissions, max depth, case sensitivity, trailing slashes).

### File List

**New Files:**
- `internal/core/ports/directory.go` - DirectoryManager and ProjectPathLookup interfaces
- `internal/adapters/filesystem/directory.go` - FilesystemDirectoryManager implementation
- `internal/adapters/filesystem/directory_test.go` - Unit tests (30+ test functions)
- `internal/adapters/filesystem/directory_integration_test.go` - Integration tests (5 test functions)

**Modified Files:**
- `internal/core/domain/errors.go` - Added ErrCollisionUnresolvable

### Code Review Fixes Applied (2025-12-18)

**HIGH Issues Fixed:**
1. **HIGH-1**: Context parameter now checked for cancellation in both `GetProjectDirName` and `EnsureProjectDir`
2. **HIGH-2**: `NewDirectoryManager` now returns nil if `UserHomeDir()` fails (instead of creating `//.vibe-dash`)
3. **HIGH-3**: Added test `TestNewDirectoryManager_NilConfigLookup` for nil configLookup handling
4. **HIGH-4**: Added Unicode path handling tests (`TestGetProjectDirName_UnicodePath`, `TestGetProjectDirName_PureUnicodeFallback`)

**MEDIUM Issues Fixed:**
1. **MEDIUM-2**: Added test `TestGetProjectDirName_RelativePath` for relative path handling
2. **MEDIUM-4**: Added test `TestGetProjectDirName_WithDigits` for digit handling in paths
3. Empty normalized names now skipped during collision resolution (Unicode fallback to parent)

**LOW Issues:** Not fixed (cosmetic only)
- LOW-1: Regex variable names unchanged (no functional impact)
- LOW-2: Test comment inconsistency unchanged (no functional impact)

**Tests Added:**
- `TestNewDirectoryManager_NilConfigLookup` - nil lookup parameter
- `TestGetProjectDirName_ContextCancellation` - context cancellation handling
- `TestEnsureProjectDir_ContextCancellation` - context cancellation handling
- `TestGetProjectDirName_UnicodePath` - Unicode character handling
- `TestGetProjectDirName_PureUnicodeFallback` - pure Unicode → parent fallback
- `TestGetProjectDirName_WithDigits` - digit preservation in names
- `TestGetProjectDirName_RelativePath` - relative path conversion

