# Story 1.2: Domain Entities

**Status:** Done

## Story

**As a** developer,
**I want** core domain entities defined with zero external dependencies,
**So that** the domain layer is pure and testable.

## Acceptance Criteria

```gherkin
AC1: Given I need to model the core domain
     When I create domain entities in internal/core/domain/
     Then Project entity exists with all required fields

AC2: Given I need stage representation
     When I create the Stage enum
     Then it contains: StageUnknown, StageSpecify, StagePlan, StageTasks, StageImplement
     And Stage.String() returns human-readable names

AC3: Given I need detection result representation
     When I create the DetectionResult value object
     Then it contains: Method, Stage, Confidence, Reasoning fields

AC4: Given I need confidence levels
     When I create the Confidence enum
     Then it contains: ConfidenceCertain, ConfidenceLikely, ConfidenceUncertain
     And Confidence.String() returns human-readable names

AC5: Given I need project state representation
     When I create the ProjectState enum
     Then it contains: StateActive, StateHibernated

AC6: Given I need domain error types
     When I create errors.go
     Then it defines: ErrProjectNotFound, ErrProjectAlreadyExists, ErrDetectionFailed, ErrConfigInvalid, ErrPathNotAccessible, ErrInvalidStage, ErrInvalidConfidence

AC7: Given I need pure domain layer
     When I inspect all files in internal/core/domain/
     Then ZERO external imports exist (only stdlib: errors, time, fmt, strings, crypto/sha256, encoding/hex)
```

## Tasks / Subtasks

- [x] **Task 1: Create Project entity** (AC: 1)
  - [x] 1.1 Create `internal/core/domain/project.go`
  - [x] 1.2 Define Project struct with all fields:
    - ID string (generated from path hash - see ID Generation Strategy below)
    - Name string
    - Path string (canonical)
    - DisplayName string (optional nickname)
    - DetectedMethod string
    - CurrentStage Stage
    - IsFavorite bool
    - State ProjectState
    - Notes string
    - LastActivityAt time.Time
    - CreatedAt time.Time
    - UpdatedAt time.Time
  - [x] 1.3 Create NewProject constructor function with validation (see Validation Rules below)
  - [x] 1.4 Add helper methods: IsHibernated(), IsActive(), HasDisplayName()
  - [x] 1.5 Add Validate() method for re-validation after modification
  - [x] 1.6 Write unit tests in `project_test.go`

- [x] **Task 2: Create Stage enum** (AC: 2)
  - [x] 2.1 Create `internal/core/domain/stage.go`
  - [x] 2.2 Define Stage type as int with const iota:
    - StageUnknown = 0
    - StageSpecify = 1
    - StagePlan = 2
    - StageTasks = 3
    - StageImplement = 4
  - [x] 2.3 Implement Stage.String() method returning human-readable names (default returns "Unknown")
  - [x] 2.4 Implement ParseStage(string) (Stage, error):
    - Case-insensitive matching ("plan", "Plan", "PLAN" all work)
    - Return StageUnknown, ErrInvalidStage for unrecognized input
  - [x] 2.5 Write unit tests in `stage_test.go` (include invalid input cases)

- [x] **Task 3: Create Confidence enum** (AC: 4)
  - [x] 3.1 Create `internal/core/domain/confidence.go`
  - [x] 3.2 Define Confidence type as int with const iota:
    - ConfidenceUncertain = 0
    - ConfidenceLikely = 1
    - ConfidenceCertain = 2
  - [x] 3.3 Implement Confidence.String() method (default returns "Unknown")
  - [x] 3.4 Implement ParseConfidence(string) (Confidence, error):
    - Case-insensitive matching
    - Return ConfidenceUncertain, ErrInvalidConfidence for unrecognized input
  - [x] 3.5 Write unit tests in `confidence_test.go` (include invalid input cases)

- [x] **Task 4: Create ProjectState enum** (AC: 5)
  - [x] 4.1 Create `internal/core/domain/state.go`
  - [x] 4.2 Define ProjectState type as int with const iota:
    - StateActive = 0
    - StateHibernated = 1
  - [x] 4.3 Implement ProjectState.String() method (default returns "Unknown")
  - [x] 4.4 Write unit tests in `state_test.go`

- [x] **Task 5: Create DetectionResult value object** (AC: 3)
  - [x] 5.1 Create `internal/core/domain/detection_result.go`
  - [x] 5.2 Define DetectionResult struct:
    - Method string
    - Stage Stage
    - Confidence Confidence
    - Reasoning string
  - [x] 5.3 Create NewDetectionResult constructor
  - [x] 5.4 Add helper methods:
    - IsUncertain() bool
    - IsCertain() bool
    - Summary() string (returns "method/stage (confidence)" for logging)
  - [x] 5.5 Write unit tests in `detection_result_test.go`

- [x] **Task 6: Create domain errors** (AC: 6)
  - [x] 6.1 Create `internal/core/domain/errors.go`
  - [x] 6.2 Define sentinel errors using errors.New():
    - ErrProjectNotFound
    - ErrProjectAlreadyExists
    - ErrDetectionFailed
    - ErrConfigInvalid
    - ErrPathNotAccessible
    - ErrInvalidStage (for ParseStage invalid input)
    - ErrInvalidConfidence (for ParseConfidence invalid input)
  - [x] 6.3 Write unit tests in `errors_test.go`

- [x] **Task 7: Validate zero external dependencies** (AC: 7)
  - [x] 7.1 Run `go list -m` on domain package to verify no external imports
  - [x] 7.2 Verify only stdlib imports: errors, time, fmt, strings, crypto/sha256, encoding/hex
  - [x] 7.3 Run `make lint` to ensure code quality
  - [x] 7.4 Run `make test` to verify all tests pass

- [x] **Task 8: Remove placeholder files** (cleanup)
  - [x] 8.1 Remove `internal/core/domain/.keep` after creating real files

## Dev Notes

### Architecture Compliance - CRITICAL

**Hexagonal Architecture Boundary Rule:**

```
internal/core/domain/ → external    FORBIDDEN
```

**ALL files in `internal/core/domain/` MUST have ZERO external dependencies.**

Only allowed imports:
- `errors` (stdlib)
- `time` (stdlib)
- `fmt` (stdlib)
- `strings` (stdlib)
- `crypto/sha256` (stdlib - for ID generation)
- `encoding/hex` (stdlib - for ID generation)

**DO NOT import:**
- Any `github.com/*` packages
- Any adapter packages
- Any external libraries

### ID Generation Strategy

Use path-based hash for deterministic, collision-resistant IDs:

```go
import (
    "crypto/sha256"
    "encoding/hex"
)

// GenerateID creates a deterministic ID from canonical path
func GenerateID(canonicalPath string) string {
    hash := sha256.Sum256([]byte(canonicalPath))
    return hex.EncodeToString(hash[:])[:16] // First 16 chars of hex digest
}
```

**Rationale:**
- Same project always gets same ID (deterministic)
- Collision-resistant for practical project counts
- No external UUID dependency needed
- 16 hex chars = 64 bits = sufficient uniqueness

### NewProject Validation Rules

The `NewProject` constructor MUST validate:

```go
func NewProject(path, name string) (*Project, error) {
    // Path: REQUIRED, must be non-empty
    if path == "" {
        return nil, ErrPathNotAccessible
    }

    // Path: must be absolute (starts with /)
    if !strings.HasPrefix(path, "/") {
        return nil, fmt.Errorf("%w: path must be absolute", ErrPathNotAccessible)
    }

    // Name: derive from path if empty
    if name == "" {
        name = filepath.Base(path) // Note: use strings split, not filepath (external)
    }

    now := time.Now()
    return &Project{
        ID:             GenerateID(path),
        Name:           name,
        Path:           path,
        State:          StateActive,
        CreatedAt:      now,
        UpdatedAt:      now,
        LastActivityAt: now,
    }, nil
}
```

**Validation Errors:**
- Empty path → `ErrPathNotAccessible`
- Relative path → `ErrPathNotAccessible` with context

### Enum Implementation Pattern

All enums follow the same pattern. Apply consistently to Stage, Confidence, and ProjectState:

```go
type Stage int

const (
    StageUnknown Stage = iota  // Zero value = Unknown (safe default)
    StageSpecify
    StagePlan
    StageTasks
    StageImplement
)

// String returns human-readable name. Default returns "Unknown" for safety.
func (s Stage) String() string {
    switch s {
    case StageSpecify:
        return "Specify"
    case StagePlan:
        return "Plan"
    case StageTasks:
        return "Tasks"
    case StageImplement:
        return "Implement"
    default:
        return "Unknown"
    }
}

// ParseStage converts string to Stage. Case-insensitive.
func ParseStage(s string) (Stage, error) {
    switch strings.ToLower(strings.TrimSpace(s)) {
    case "specify":
        return StageSpecify, nil
    case "plan":
        return StagePlan, nil
    case "tasks":
        return StageTasks, nil
    case "implement":
        return StageImplement, nil
    case "unknown", "":
        return StageUnknown, nil
    default:
        return StageUnknown, ErrInvalidStage
    }
}
```

### Sentinel Error Pattern

```go
var (
    ErrProjectNotFound      = errors.New("project not found")
    ErrProjectAlreadyExists = errors.New("project already exists")
    ErrDetectionFailed      = errors.New("detection failed")
    ErrConfigInvalid        = errors.New("configuration invalid")
    ErrPathNotAccessible    = errors.New("path not accessible")
    ErrInvalidStage         = errors.New("invalid stage")
    ErrInvalidConfidence    = errors.New("invalid confidence level")
)
```

### Error-to-Exit-Code Mapping

| Domain Error | Exit Code |
|--------------|-----------|
| `ErrProjectNotFound` | 2 |
| `ErrConfigInvalid` | 3 |
| `ErrDetectionFailed` | 4 |
| (any unhandled) | 1 |

These domain errors will be mapped to exit codes in the CLI adapter layer (Story 1.4+).

### Project Entity Fields

```go
type Project struct {
    ID             string       // Unique identifier (path hash, 16 hex chars)
    Name           string       // Derived from directory name
    Path           string       // Canonical absolute path
    DisplayName    string       // Optional user-set nickname (FR5)
    DetectedMethod string       // "speckit", "bmad", "unknown"
    CurrentStage   Stage        // Current workflow stage
    IsFavorite     bool         // Always visible regardless of activity (FR30)
    State          ProjectState // Active or Hibernated (FR28-33)
    Notes          string       // User notes/memo (FR21)
    LastActivityAt time.Time    // Last file change detected (FR34-38)
    CreatedAt      time.Time    // When project was added
    UpdatedAt      time.Time    // Last database update
}

// Helper methods
func (p *Project) IsHibernated() bool { return p.State == StateHibernated }
func (p *Project) IsActive() bool     { return p.State == StateActive }
func (p *Project) HasDisplayName() bool { return p.DisplayName != "" }

// Validate checks Project invariants. Use after modification.
func (p *Project) Validate() error {
    if p.Path == "" {
        return ErrPathNotAccessible
    }
    if p.ID == "" {
        return fmt.Errorf("project ID cannot be empty")
    }
    return nil
}
```

### DetectionResult Value Object

```go
type DetectionResult struct {
    Method     string     // "speckit", "bmad", "unknown"
    Stage      Stage      // Detected stage
    Confidence Confidence // How certain the detection is
    Reasoning  string     // Human-readable explanation (FR11, FR26)
}

// Helper methods
func (dr DetectionResult) IsUncertain() bool { return dr.Confidence == ConfidenceUncertain }
func (dr DetectionResult) IsCertain() bool   { return dr.Confidence == ConfidenceCertain }

// Summary returns concise string for logging: "speckit/Plan (Certain)"
func (dr DetectionResult) Summary() string {
    return fmt.Sprintf("%s/%s (%s)", dr.Method, dr.Stage, dr.Confidence)
}
```

The `Reasoning` field is critical for user trust - it explains WHY the stage was detected (e.g., "plan.md exists, no tasks.md found").

### Confidence Levels

| Level | Description | UI Indicator |
|-------|-------------|--------------|
| ConfidenceCertain | Clear artifact markers | Normal display |
| ConfidenceLikely | Strong indicators | Normal display |
| ConfidenceUncertain | Ambiguous markers | 🤷 indicator |

### Testing Requirements

- **Location:** Same package, `_test.go` suffix
- **Pattern:** Table-driven tests with edge cases
- **Coverage:** All public methods, constructors, and Parse functions

**Test Cases to Include:**

```go
// Stage tests
func TestParseStage(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Stage
        wantErr error
    }{
        {"valid lowercase", "plan", StagePlan, nil},
        {"valid uppercase", "PLAN", StagePlan, nil},
        {"valid mixed", "Plan", StagePlan, nil},
        {"with spaces", "  plan  ", StagePlan, nil},
        {"empty string", "", StageUnknown, nil},
        {"invalid", "invalid", StageUnknown, ErrInvalidStage},
    }
    // ... test implementation
}
```

### Previous Story Context

**Story 1-1 (Done):**
- Directory `internal/core/domain/` already exists with `.keep` file
- Go module initialized: `github.com/JeiKeiLim/vibe-dash`
- Go version: 1.24.3
- Build system: `make build`, `make test`, `make lint` working

### DO NOT (Anti-Patterns)

| DO NOT | DO INSTEAD |
|--------|------------|
| Import any external packages | Use only stdlib (errors, time, fmt, strings, crypto/sha256) |
| Add database tags to structs | Keep domain pure, add tags in persistence adapter |
| Add JSON tags to structs | Keep domain pure, add serialization in adapters |
| Create complex validation in domain | Keep basic validation, complex logic in services |
| Use `Id` for identifier | Use `ID` (Go acronym convention) |
| Return raw strings for errors | Return domain error types (ErrInvalidStage) |
| Skip edge case tests | Test empty strings, invalid input, boundary conditions |

### Project Structure Notes

- **Alignment:** Files go in `internal/core/domain/` exactly as specified
- **Naming:** Lowercase filenames with underscores for multi-word (`detection_result.go`)
- **Package name:** `domain`

## Dev Agent Record

### Context Reference

Story context created from comprehensive analysis of:
- docs/epics.md (Story 1.2 requirements)
- docs/architecture.md (hexagonal boundaries, Go conventions)
- docs/prd.md (functional requirements FR9-14, FR28-33, FR34-38)
- docs/project-context.md (critical rules)
- docs/sprint-artifacts/1-1-project-scaffolding.md (previous story learnings)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- All tests pass with `make test`
- Lint passes with `make lint`
- Zero external dependencies verified with `go list -f '{{join .Imports "\n"}}' ./internal/core/domain/`

### Completion Notes List

- Implemented all domain entities following hexagonal architecture rules
- Project entity with path-based ID generation (SHA-256 hash, 16 hex chars)
- Stage enum with String() and ParseStage() methods (case-insensitive)
- Confidence enum with String() and ParseConfidence() methods
- ProjectState enum with String() and ParseProjectState() methods
- DetectionResult value object with Summary(), IsLikely(), IsUncertain(), IsCertain() helpers
- Domain sentinel errors for error handling consistency
- All entities use only stdlib imports (errors, time, fmt, strings, crypto/sha256, encoding/hex)
- Comprehensive table-driven tests for all entities and edge cases

### Code Review Fixes (2025-12-12)

**Reviewed by:** Amelia (Dev Agent - Code Review Mode)

**Issues Fixed:**
- **H1/H2:** Fixed trailing slash path name derivation bug in `NewProject()` - paths ending with `/` now correctly derive project name (e.g., `/home/user/project/` → "project", `/` → "root")
- **M1:** Added absolute path validation to `Project.Validate()` for consistency with `NewProject()`
- **M2:** Added `ParseProjectState()` function to state.go for API consistency with Stage and Confidence enums
- **M3:** Added `IsLikely()` helper method to DetectionResult for complete confidence level API
- Added `ErrInvalidProjectState` sentinel error
- Added comprehensive tests for all new functionality and edge cases

**Verification:**
- All tests pass: `make test`
- Lint passes: `make lint`
- Coverage: 100%
- Zero external imports verified

### File List

**Files (created + modified during review):**
- internal/core/domain/project.go (created, modified: trailing slash fix, Validate() fix)
- internal/core/domain/project_test.go (created, modified: added edge case tests)
- internal/core/domain/stage.go (created)
- internal/core/domain/stage_test.go (created)
- internal/core/domain/confidence.go (created)
- internal/core/domain/confidence_test.go (created)
- internal/core/domain/state.go (created, modified: added ParseProjectState())
- internal/core/domain/state_test.go (created, modified: added ParseProjectState tests)
- internal/core/domain/detection_result.go (created, modified: added IsLikely())
- internal/core/domain/detection_result_test.go (created, modified: added IsLikely tests)
- internal/core/domain/errors.go (created)
- internal/core/domain/errors_test.go (created)

**Deleted files:**
- internal/core/domain/.keep
