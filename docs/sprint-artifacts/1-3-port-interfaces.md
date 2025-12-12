# Story 1.3: Port Interfaces

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Interfaces** | MethodDetector, ProjectRepository, FileWatcher, ConfigLoader |
| **Structs** | Config, ProjectConfig, FileEvent |
| **Enums** | FileOperation (Create/Modify/Delete) |
| **Files** | detector.go, repository.go, watcher.go, config.go (+ tests) |
| **Imports** | stdlib only (context, time, fmt) + internal domain |

## Story

**As a** developer,
**I want** port interfaces defined for all external dependencies,
**So that** adapters can be injected and the core remains testable.

## Acceptance Criteria

```gherkin
AC1: Given I need to define boundaries between core and adapters
     When I create interfaces in internal/core/ports/
     Then MethodDetector interface exists with methods:
       - Name() string
       - CanDetect(ctx context.Context, path string) bool
       - Detect(ctx context.Context, path string) (*domain.DetectionResult, error)

AC2: Given I need project persistence abstraction
     When I create ProjectRepository interface
     Then it exists with methods:
       - Save(ctx context.Context, project *domain.Project) error
       - FindByID(ctx context.Context, id string) (*domain.Project, error)
       - FindByPath(ctx context.Context, path string) (*domain.Project, error)
       - FindAll(ctx context.Context) ([]*domain.Project, error)
       - FindActive(ctx context.Context) ([]*domain.Project, error)
       - FindHibernated(ctx context.Context) ([]*domain.Project, error)
       - Delete(ctx context.Context, id string) error
       - UpdateState(ctx context.Context, id string, state domain.ProjectState) error

AC3: Given I need file watching abstraction
     When I create FileWatcher interface
     Then it exists with methods:
       - Watch(ctx context.Context, paths []string) (<-chan FileEvent, error)
       - Close() error
     And FileEvent struct exists with Path, Operation, Timestamp fields

AC4: Given I need configuration loading abstraction
     When I create ConfigLoader interface
     Then it exists with methods:
       - Load(ctx context.Context) (*Config, error)
       - Save(ctx context.Context, config *Config) error
     And Config struct exists with all configuration fields

AC5: Given I need pure port layer
     When I inspect all interfaces in internal/core/ports/
     Then all methods accept context.Context as first parameter (where applicable)
     And only domain types are used (no adapter-specific types)
     And only stdlib imports are allowed (context, time) plus internal domain package
```

## Tasks / Subtasks

- [x] **Task 1: Create MethodDetector interface** (AC: 1)
  - [x] 1.1 Create `internal/core/ports/detector.go`
  - [x] 1.2 Define MethodDetector interface with methods:
    - Name() string - returns detector name (e.g., "speckit")
    - CanDetect(ctx context.Context, path string) bool - checks if detector can handle this path
    - Detect(ctx context.Context, path string) (*domain.DetectionResult, error) - performs detection
  - [x] 1.3 Add documentation comments explaining interface purpose
  - [x] 1.4 Write interface compliance test in `detector_test.go`

- [x] **Task 2: Create ProjectRepository interface** (AC: 2)
  - [x] 2.1 Create `internal/core/ports/repository.go`
  - [x] 2.2 Define ProjectRepository interface with all CRUD methods
  - [x] 2.3 Ensure all methods accept context.Context as first parameter
  - [x] 2.4 Use domain.Project and domain.ProjectState types
  - [x] 2.5 Write interface compliance test in `repository_test.go`

- [x] **Task 3: Create FileWatcher interface and FileEvent** (AC: 3)
  - [x] 3.1 Create `internal/core/ports/watcher.go`
  - [x] 3.2 Define FileOperation type (create, modify, delete) with String() and Valid() methods
  - [x] 3.3 Define FileEvent struct:
    - Path string (canonical path)
    - Operation FileOperation
    - Timestamp time.Time
  - [x] 3.4 Define FileWatcher interface:
    - Watch(ctx context.Context, paths []string) (<-chan FileEvent, error)
    - Close() error
  - [x] 3.5 Write interface compliance test in `watcher_test.go` (include FileOperation.Valid() tests)

- [x] **Task 4: Create ConfigLoader interface and Config struct** (AC: 4)
  - [x] 4.1 Create `internal/core/ports/config.go`
  - [x] 4.2 Define Config struct with fields from architecture:
    - HibernationDays int (default: 14)
    - RefreshIntervalSeconds int (default: 10)
    - RefreshDebounceMs int (default: 200)
    - AgentWaitingThresholdMinutes int (default: 10)
    - Projects map[string]ProjectConfig
  - [x] 4.3 Define ProjectConfig struct for per-project overrides:
    - Path string
    - DisplayName string
    - IsFavorite bool
    - HibernationDays *int (nil = use global)
    - AgentWaitingThresholdMinutes *int (nil = use global)
  - [x] 4.4 Define ConfigLoader interface:
    - Load(ctx context.Context) (*Config, error)
    - Save(ctx context.Context, config *Config) error
  - [x] 4.5 Add NewConfig() constructor with defaults
  - [x] 4.6 Add GetProjectConfig(projectID string) method on Config
  - [x] 4.7 Add Validate() method with bounds checking for all int fields
  - [x] 4.8 Write tests in `config_test.go` (include Validate() edge cases)

- [x] **Task 5: Validate port layer purity** (AC: 5)
  - [x] 5.1 Verify all interfaces use context.Context as first parameter where applicable
  - [x] 5.2 Verify only domain package imports (no adapter imports)
  - [x] 5.3 Verify only stdlib imports: context, time
  - [x] 5.4 Run `make lint` to ensure code quality
  - [x] 5.5 Run `make test` to verify all tests pass

- [x] **Task 6: Remove placeholder file** (cleanup)
  - [x] 6.1 Remove `internal/core/ports/.keep` after creating real files

## Dev Notes

### Architecture Compliance - CRITICAL

**Hexagonal Architecture Boundary Rule:**

```
internal/core/ports/ → external    FORBIDDEN
```

**ALL files in `internal/core/ports/` MUST have ZERO external dependencies.**

Only allowed imports:
- `context` (stdlib)
- `time` (stdlib)
- `fmt` (stdlib - for validation errors)
- `github.com/JeiKeiLim/vibe-dash/internal/core/domain` (internal domain package)

**DO NOT import:**
- Any `github.com/*` external packages
- Any adapter packages (`internal/adapters/*`)
- Any external libraries

### Context Propagation Pattern - MANDATORY

From Architecture doc: "All service methods accept `context.Context` as first parameter"

```go
// CORRECT - supports cancellation when user hits 'q' in TUI
func (d MethodDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error)

// WRONG - no cancellation support
func (d MethodDetector) Detect(path string) (*domain.DetectionResult, error)
```

**Context Cancellation Contract:**

All methods accepting `context.Context` MUST respect cancellation:
- When `ctx.Done()` fires, stop work promptly (within 100ms)
- Return `ctx.Err()` wrapped with context: `fmt.Errorf("operation cancelled: %w", ctx.Err())`
- Do NOT leave partial state - either complete or rollback cleanly
- Check `ctx.Err()` before long-running operations

### Interface Implementation Pattern

Interfaces define behavior that adapters implement:

```go
// Port interface (in internal/core/ports/)
type MethodDetector interface {
    Name() string
    CanDetect(ctx context.Context, path string) bool
    Detect(ctx context.Context, path string) (*domain.DetectionResult, error)
}

// Adapter implementation (in internal/adapters/detectors/speckit/)
type SpeckitDetector struct{}

func (d *SpeckitDetector) Name() string { return "speckit" }
func (d *SpeckitDetector) CanDetect(ctx context.Context, path string) bool { /* ... */ }
func (d *SpeckitDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) { /* ... */ }
```

### FileEvent Value Object

```go
// FileOperation represents the type of file system event
type FileOperation int

const (
    FileOpCreate FileOperation = iota
    FileOpModify
    FileOpDelete
)

func (op FileOperation) String() string {
    switch op {
    case FileOpCreate:
        return "create"
    case FileOpModify:
        return "modify"
    case FileOpDelete:
        return "delete"
    default:
        return "unknown"
    }
}

// Valid returns true if the FileOperation is a known value
func (op FileOperation) Valid() bool {
    return op >= FileOpCreate && op <= FileOpDelete
}

// FileEvent represents a file system change event
type FileEvent struct {
    Path      string        // Canonical path of changed file
    Operation FileOperation // Type of operation
    Timestamp time.Time     // When the event occurred
}
```

### Config Struct with Defaults

```go
// Config represents application configuration
type Config struct {
    // Global settings (FR39, FR46)
    HibernationDays               int                     // Days before auto-hibernate (default: 14)
    RefreshIntervalSeconds        int                     // TUI refresh interval (default: 10)
    RefreshDebounceMs             int                     // Debounce for file events (default: 200)
    AgentWaitingThresholdMinutes  int                     // Minutes before ⏸️ WAITING (default: 10)
    Projects                      map[string]ProjectConfig // Per-project settings
}

// ProjectConfig represents per-project configuration overrides (FR47)
type ProjectConfig struct {
    Path                          string // Project path
    DisplayName                   string // Custom display name (FR5)
    IsFavorite                    bool   // Favorite status (FR30)
    HibernationDays               *int   // Override global (nil = use global)
    AgentWaitingThresholdMinutes  *int   // Override global (nil = use global)
}

// NewConfig creates a Config with default values
func NewConfig() *Config {
    return &Config{
        HibernationDays:              14,
        RefreshIntervalSeconds:       10,
        RefreshDebounceMs:            200,
        AgentWaitingThresholdMinutes: 10,
        Projects:                     make(map[string]ProjectConfig),
    }
}

// GetEffectiveHibernationDays returns project-specific or global value
func (c *Config) GetEffectiveHibernationDays(projectID string) int {
    if pc, ok := c.Projects[projectID]; ok && pc.HibernationDays != nil {
        return *pc.HibernationDays
    }
    return c.HibernationDays
}

// GetEffectiveWaitingThreshold returns project-specific or global value
func (c *Config) GetEffectiveWaitingThreshold(projectID string) int {
    if pc, ok := c.Projects[projectID]; ok && pc.AgentWaitingThresholdMinutes != nil {
        return *pc.AgentWaitingThresholdMinutes
    }
    return c.AgentWaitingThresholdMinutes
}

// Validate checks Config values are within acceptable ranges
// Call after loading config or modifying values
func (c *Config) Validate() error {
    if c.HibernationDays < 0 {
        return fmt.Errorf("hibernation_days must be >= 0, got %d", c.HibernationDays)
    }
    if c.RefreshIntervalSeconds <= 0 {
        return fmt.Errorf("refresh_interval_seconds must be > 0, got %d", c.RefreshIntervalSeconds)
    }
    if c.RefreshDebounceMs <= 0 {
        return fmt.Errorf("refresh_debounce_ms must be > 0, got %d", c.RefreshDebounceMs)
    }
    if c.AgentWaitingThresholdMinutes < 0 {
        return fmt.Errorf("agent_waiting_threshold_minutes must be >= 0, got %d", c.AgentWaitingThresholdMinutes)
    }
    return nil
}
```

**Important:** Always call `Validate()` after loading config from disk or modifying values programmatically.

### ProjectRepository Interface Pattern

```go
// ProjectRepository defines persistence operations for projects (FR39-45)
type ProjectRepository interface {
    // Save creates or updates a project
    Save(ctx context.Context, project *domain.Project) error

    // FindByID retrieves a project by its unique identifier
    FindByID(ctx context.Context, id string) (*domain.Project, error)

    // FindByPath retrieves a project by its canonical path
    FindByPath(ctx context.Context, path string) (*domain.Project, error)

    // FindAll retrieves all projects regardless of state
    FindAll(ctx context.Context) ([]*domain.Project, error)

    // FindActive retrieves only active (non-hibernated) projects
    FindActive(ctx context.Context) ([]*domain.Project, error)

    // FindHibernated retrieves only hibernated projects
    FindHibernated(ctx context.Context) ([]*domain.Project, error)

    // Delete removes a project by ID
    Delete(ctx context.Context, id string) error

    // UpdateState changes the project's active/hibernated state
    UpdateState(ctx context.Context, id string, state domain.ProjectState) error

    // OPTIONAL (consider adding post-MVP for FR24 optimization):
    // CountByState returns count of projects by state (more efficient than FindAll + filter)
    // CountByState(ctx context.Context, state domain.ProjectState) (int, error)
}
```

### Error Handling Pattern

Repository methods should return domain errors:

```go
// When project not found in repository
return nil, domain.ErrProjectNotFound

// When path is inaccessible during detection
return nil, fmt.Errorf("%w: cannot access %s", domain.ErrPathNotAccessible, path)

// When detection fails
return nil, fmt.Errorf("%w: no markers found in %s", domain.ErrDetectionFailed, path)
```

### Testing Port Interfaces

Since interfaces define contracts, tests verify:
1. Interface is syntactically correct
2. Documentation is present
3. Domain types are used correctly

```go
// detector_test.go
package ports_test

import (
    "context"
    "testing"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// mockDetector verifies interface compliance
type mockDetector struct{}

func (m *mockDetector) Name() string { return "mock" }
func (m *mockDetector) CanDetect(ctx context.Context, path string) bool { return true }
func (m *mockDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
    result := domain.NewDetectionResult("mock", domain.StageUnknown, domain.ConfidenceUncertain, "mock detection")
    return &result, nil
}

// Compile-time interface compliance check
var _ ports.MethodDetector = (*mockDetector)(nil)

func TestMethodDetector_InterfaceCompliance(t *testing.T) {
    var d ports.MethodDetector = &mockDetector{}

    if d.Name() != "mock" {
        t.Errorf("Name() = %q, want %q", d.Name(), "mock")
    }
}
```

### Previous Story Context (Story 1.2)

**Domain entities available:**
- `domain.Project` - with ID, Path, Name, State, etc.
- `domain.ProjectState` - StateActive, StateHibernated
- `domain.Stage` - StageUnknown, StageSpecify, StagePlan, StageTasks, StageImplement
- `domain.Confidence` - ConfidenceUncertain, ConfidenceLikely, ConfidenceCertain
- `domain.DetectionResult` - with Method, Stage, Confidence, Reasoning

**Domain errors available:**
- `domain.ErrProjectNotFound`
- `domain.ErrProjectAlreadyExists`
- `domain.ErrDetectionFailed`
- `domain.ErrConfigInvalid`
- `domain.ErrPathNotAccessible`

**Key patterns from Story 1.2:**
- Zero external dependencies in core layer
- Enum pattern with String() and Parse*() methods
- Table-driven tests with edge cases
- Sentinel error pattern

### Files to Create

| File | Purpose |
|------|---------|
| `internal/core/ports/detector.go` | MethodDetector interface |
| `internal/core/ports/detector_test.go` | Interface compliance tests |
| `internal/core/ports/repository.go` | ProjectRepository interface |
| `internal/core/ports/repository_test.go` | Interface compliance tests |
| `internal/core/ports/watcher.go` | FileWatcher interface + FileEvent |
| `internal/core/ports/watcher_test.go` | Interface compliance tests |
| `internal/core/ports/config.go` | ConfigLoader interface + Config struct |
| `internal/core/ports/config_test.go` | Config tests |

### DO NOT (Anti-Patterns)

| DO NOT | DO INSTEAD |
|--------|------------|
| Import any external packages | Use only stdlib (context, time) + domain |
| Import adapter packages | Adapters implement ports, not vice versa |
| Add implementation details | Interfaces define contracts only |
| Skip context.Context parameter | Always first parameter for async-safe methods |
| Use adapter-specific types | Use domain types only |
| Create complex logic in ports | Keep interfaces minimal, logic in services |
| Return concrete types | Return interface-compatible types |

### Project Structure Notes

- **Location:** `internal/core/ports/`
- **Package name:** `ports`
- **Naming:** Interface files named after the concept (detector.go, repository.go)

### References

- [Architecture: Project Structure](docs/architecture.md#complete-project-directory-structure)
- [Architecture: Architectural Boundaries](docs/architecture.md#architectural-boundaries)
- [Architecture: Context Propagation](docs/architecture.md#go-code-conventions)
- [PRD: FR9-14](docs/prd.md) - Workflow Detection
- [PRD: FR27](docs/prd.md) - File System Changes Detection
- [PRD: FR39-47](docs/prd.md) - Configuration Management

## Dev Agent Record

### Context Reference

Story context created from comprehensive analysis of:
- docs/epics.md (Story 1.3 requirements)
- docs/architecture.md (hexagonal boundaries, ports section, Go conventions)
- docs/prd.md (functional requirements)
- docs/project-context.md (critical rules)
- docs/sprint-artifacts/1-2-domain-entities.md (previous story learnings)
- internal/core/domain/*.go (existing domain implementations)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None - clean implementation with no issues.

### Completion Notes List

- ✅ Task 1: Created MethodDetector interface with Name(), CanDetect(), Detect() methods
- ✅ Task 2: Created ProjectRepository interface with all CRUD methods (Save, FindByID, FindByPath, FindAll, FindActive, FindHibernated, Delete, UpdateState)
- ✅ Task 3: Created FileWatcher interface and FileEvent struct with FileOperation enum (Create/Modify/Delete)
- ✅ Task 4: Created ConfigLoader interface, Config struct with defaults, ProjectConfig for per-project overrides, Validate() method
- ✅ Task 5: Verified port layer purity - only stdlib (context, time, fmt) and domain imports
- ✅ Task 6: Removed .keep placeholder file
- All tests pass (100% coverage on interface compliance)
- All linting passes (golangci-lint)
- Zero external dependencies in ports package

### File List

**New Files:**
- internal/core/ports/detector.go
- internal/core/ports/detector_test.go
- internal/core/ports/repository.go
- internal/core/ports/repository_test.go
- internal/core/ports/watcher.go
- internal/core/ports/watcher_test.go
- internal/core/ports/config.go
- internal/core/ports/config_test.go

**Deleted Files:**
- internal/core/ports/.keep

### Change Log

- 2025-12-12: Story 1.3 implementation complete - all port interfaces defined with comprehensive tests
- 2025-12-12: Code Review fixes applied:
  - H1: Added context.Context to ConfigLoader.Load() and ConfigLoader.Save() methods (architecture compliance)
  - M1: Added per-project validation to Config.Validate() for HibernationDays and AgentWaitingThresholdMinutes overrides
  - M2: Wrapped all validation errors with domain.ErrConfigInvalid for consistent error handling
  - M3: Added StateActive transition test to repository_test.go (hibernate → active workflow)
  - M4: Clarified context cancellation test comments - interface tests verify signature, adapter tests verify behavior
  - Added config.go import of domain package for error wrapping
  - Added 6 new test cases for ProjectConfig validation including nil Projects map handling

