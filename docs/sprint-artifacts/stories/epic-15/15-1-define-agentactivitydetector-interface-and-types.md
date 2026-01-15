# Story 15.1: Define AgentActivityDetector Interface and Types

Status: done

## Story

As a developer,
I want a well-defined interface for agent activity detection,
So that multiple detector implementations can be plugged in.

## User-Visible Changes

None - this is an internal infrastructure change establishing the foundation for sub-1-minute agent detection. User-visible changes will come in Story 15.6 when this is integrated into the TUI dashboard.

## Acceptance Criteria

1. **AC1:** Interface is defined in `internal/core/ports/agent_activity_detector.go` with proper package-level documentation following `log_reader.go` pattern
2. **AC2:** `AgentState` struct contains: Tool (string), Status (AgentStatus), Duration (time.Duration), Confidence (Confidence)
3. **AC3:** `AgentStatus` type uses `int` with `iota` pattern (matching `Confidence` and `Stage` patterns), with values: AgentUnknown (zero value), AgentWorking, AgentWaitingForUser, AgentInactive
4. **AC4:** Interface has `Detect(ctx context.Context, projectPath string) (domain.AgentState, error)` method
5. **AC5:** Interface has `Name() string` method returning detector identifier ("Claude Code", "Generic")
6. **AC6:** Domain types follow existing patterns from `domain/confidence.go` and `domain/stage.go`
7. **AC7:** `ErrInvalidAgentStatus` sentinel error is added to `domain/errors.go`

## Tasks / Subtasks

- [x] Task 1: Create AgentStatus type in domain layer
  - [x] 1.1: Add `ErrInvalidAgentStatus` to `internal/core/domain/errors.go`
  - [x] 1.2: Create `internal/core/domain/agent_status.go`
  - [x] 1.3: Define `type AgentStatus int` with iota constants: `AgentUnknown` (MUST be first for zero-value safety), `AgentWorking`, `AgentWaitingForUser`, `AgentInactive`
  - [x] 1.4: Add `String()` method returning "Unknown", "Working", "Waiting", "Inactive"
  - [x] 1.5: Add `ParseAgentStatus(s string) (AgentStatus, error)` function (case-insensitive, returns `ErrInvalidAgentStatus` on invalid input)
  - [x] 1.6: Write unit tests in `internal/core/domain/agent_status_test.go` (see Testing Patterns below)

- [x] Task 2: Create AgentState struct in domain layer
  - [x] 2.1: Create `internal/core/domain/agent_state.go`
  - [x] 2.2: Define AgentState struct with fields: Tool (string), Status (AgentStatus), Duration (time.Duration), Confidence (Confidence)
  - [x] 2.3: Add `NewAgentState(tool string, status AgentStatus, duration time.Duration, confidence Confidence) AgentState` constructor
  - [x] 2.4: Add helper methods: `IsWaiting() bool`, `IsWorking() bool`, `IsInactive() bool`, `IsUnknown() bool`
  - [x] 2.5: Add `Summary() string` method returning `"{Tool}/{Status} ({Confidence})"` format
  - [x] 2.6: Write unit tests in `internal/core/domain/agent_state_test.go`

- [x] Task 3: Create AgentActivityDetector interface in ports layer
  - [x] 3.1: Create `internal/core/ports/agent_activity_detector.go` with package-level doc comment
  - [x] 3.2: Define `AgentActivityDetector` interface with `Detect()` and `Name()` methods
  - [x] 3.3: Add comprehensive doc comments for each method (see Interface Specification below)
  - [x] 3.4: Create `internal/core/ports/agent_activity_detector_test.go` with interface documentation tests

- [x] Task 4: Verify hexagonal architecture compliance
  - [x] 4.1: Run `go build ./internal/core/...` to verify no external dependencies
  - [x] 4.2: Run `go test ./internal/core/...` to ensure all tests pass
  - [x] 4.3: Run `make lint` to verify code style compliance

## Dev Notes

### Critical Pattern Requirements

**MUST follow these patterns exactly:**

| Pattern | Existing Reference | New Implementation |
|---------|-------------------|-------------------|
| Type definition | `type Confidence int` (confidence.go:6) | `type AgentStatus int` |
| Zero-value safety | `ConfidenceUncertain` is first/iota=0 | `AgentUnknown` MUST be first |
| String() method | confidence.go:15-26 | Same switch pattern |
| Parse function | `ParseConfidence` (confidence.go:29-42) | `ParseAgentStatus` |
| Error handling | `ErrInvalidConfidence` (errors.go:13) | `ErrInvalidAgentStatus` |

### Interface Specification

```go
// Package ports defines interfaces for external adapters.
// All interfaces in this package represent contracts between the core domain
// and external dependencies (databases, file systems, detectors, etc.).
//
// Hexagonal Architecture Boundary: This package has ZERO external dependencies.
// Only stdlib (context, time) and internal domain package imports are allowed.
package ports

import (
    "context"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// AgentActivityDetector detects the activity state of an AI agent for a project.
// Implementations parse tool-specific logs or use heuristics to determine
// whether an agent is Working, WaitingForUser, or Inactive.
//
// All methods accepting context.Context must respect cancellation:
// - Check ctx.Done() before long-running operations
// - Return ctx.Err() wrapped with context when cancelled
// - Stop work promptly (within 100ms) when cancellation is signaled
type AgentActivityDetector interface {
    // Detect determines the current agent activity state for a project.
    // Returns AgentState with Status, Duration, Tool, and Confidence.
    //
    // Returns error if detection cannot be performed (e.g., path not accessible).
    // Returns AgentState with Status=AgentUnknown if detection completes but
    // agent state cannot be determined.
    Detect(ctx context.Context, projectPath string) (domain.AgentState, error)

    // Name returns the detector identifier (e.g., "Claude Code", "Generic").
    // Used for logging and to populate AgentState.Tool when this detector matches.
    Name() string
}
```

### Domain Types Specification

```go
// internal/core/domain/agent_status.go
package domain

import "strings"

// AgentStatus represents the detected state of an AI agent.
// Uses int with iota to match Confidence and Stage patterns.
type AgentStatus int

const (
    AgentUnknown        AgentStatus = iota // Zero value - cannot determine state
    AgentWorking                           // Agent actively processing/using tools
    AgentWaitingForUser                    // Agent waiting for user input (THE target state)
    AgentInactive                          // No recent agent activity
)

// String returns human-readable name. Default returns "Unknown" for safety.
func (s AgentStatus) String() string {
    switch s {
    case AgentWorking:
        return "Working"
    case AgentWaitingForUser:
        return "Waiting"
    case AgentInactive:
        return "Inactive"
    default:
        return "Unknown"
    }
}

// ParseAgentStatus converts string to AgentStatus. Case-insensitive.
func ParseAgentStatus(s string) (AgentStatus, error) {
    switch strings.ToLower(strings.TrimSpace(s)) {
    case "working":
        return AgentWorking, nil
    case "waiting", "waitingforuser":
        return AgentWaitingForUser, nil
    case "inactive":
        return AgentInactive, nil
    case "unknown", "":
        return AgentUnknown, nil
    default:
        return AgentUnknown, ErrInvalidAgentStatus
    }
}
```

```go
// internal/core/domain/agent_state.go
package domain

import (
    "fmt"
    "time"
)

// AgentState represents the complete detected state of an AI agent for a project.
type AgentState struct {
    Tool       string        // "Claude Code", "Generic", "Unknown"
    Status     AgentStatus   // Working, WaitingForUser, Inactive, Unknown
    Duration   time.Duration // How long in current state
    Confidence Confidence    // High (log-based), Low (heuristic)
}

// NewAgentState creates a new AgentState with the given values.
func NewAgentState(tool string, status AgentStatus, duration time.Duration, confidence Confidence) AgentState {
    return AgentState{
        Tool:       tool,
        Status:     status,
        Duration:   duration,
        Confidence: confidence,
    }
}

// IsWaiting returns true if the agent is waiting for user input.
func (s AgentState) IsWaiting() bool {
    return s.Status == AgentWaitingForUser
}

// IsWorking returns true if the agent is actively working.
func (s AgentState) IsWorking() bool {
    return s.Status == AgentWorking
}

// IsInactive returns true if the agent has no recent activity.
func (s AgentState) IsInactive() bool {
    return s.Status == AgentInactive
}

// IsUnknown returns true if the agent state cannot be determined.
func (s AgentState) IsUnknown() bool {
    return s.Status == AgentUnknown
}

// Summary returns concise string for logging: "Claude Code/Waiting (Certain)"
func (s AgentState) Summary() string {
    return fmt.Sprintf("%s/%s (%s)", s.Tool, s.Status, s.Confidence)
}
```

### Testing Patterns

Follow `confidence_test.go` exactly. Required test cases:

**agent_status_test.go:**
```go
func TestAgentStatus_String(t *testing.T) {
    tests := []struct {
        name   string
        status AgentStatus
        want   string
    }{
        {"unknown", AgentUnknown, "Unknown"},
        {"working", AgentWorking, "Working"},
        {"waiting", AgentWaitingForUser, "Waiting"},
        {"inactive", AgentInactive, "Inactive"},
        {"invalid negative", AgentStatus(-1), "Unknown"},
        {"invalid large", AgentStatus(100), "Unknown"},
    }
    // ... table-driven test implementation
}

func TestParseAgentStatus(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    AgentStatus
        wantErr error
    }{
        {"valid lowercase working", "working", AgentWorking, nil},
        {"valid waiting", "waiting", AgentWaitingForUser, nil},
        {"valid waitingforuser", "waitingforuser", AgentWaitingForUser, nil},
        {"valid uppercase", "WORKING", AgentWorking, nil},
        {"with spaces", "  working  ", AgentWorking, nil},
        {"empty string", "", AgentUnknown, nil},
        {"invalid", "invalid", AgentUnknown, ErrInvalidAgentStatus},
    }
    // ... table-driven test implementation
}

func TestAgentStatus_ZeroValue(t *testing.T) {
    var status AgentStatus // zero value
    if status != AgentUnknown {
        t.Errorf("zero value = %v, want AgentUnknown", status)
    }
}
```

**agent_state_test.go:**
```go
func TestAgentState_Helpers(t *testing.T) {
    tests := []struct {
        name       string
        status     AgentStatus
        isWaiting  bool
        isWorking  bool
        isInactive bool
        isUnknown  bool
    }{
        {"working", AgentWorking, false, true, false, false},
        {"waiting", AgentWaitingForUser, true, false, false, false},
        {"inactive", AgentInactive, false, false, true, false},
        {"unknown", AgentUnknown, false, false, false, true},
    }
    // ... table-driven test implementation
}

func TestAgentState_Summary(t *testing.T) {
    state := NewAgentState("Claude Code", AgentWaitingForUser, 2*time.Hour, ConfidenceCertain)
    want := "Claude Code/Waiting (Certain)"
    if got := state.Summary(); got != want {
        t.Errorf("Summary() = %q, want %q", got, want)
    }
}
```

### Difference from Existing WaitingDetector

**Do NOT modify `internal/core/ports/waiting_detector.go` in this story.**

The existing `WaitingDetector` uses a simple boolean approach based on 10-minute file activity threshold. This new interface provides:

| Aspect | WaitingDetector (existing) | AgentActivityDetector (new) |
|--------|---------------------------|----------------------------|
| Detection | Binary: waiting/not waiting | Rich: Working/Waiting/Inactive/Unknown |
| Confidence | Implicit (always low) | Explicit: High (log-based), Low (heuristic) |
| Duration | Separate method | Included in AgentState |
| Tool awareness | None | Explicit: which tool is detected |

The new interface will **replace** `WaitingDetector` usage in the TUI once fully implemented (Story 15.6).

### Project Structure Impact

```
internal/core/domain/
├── agent_status.go      # NEW: AgentStatus int type with iota
├── agent_status_test.go # NEW: Tests for AgentStatus
├── agent_state.go       # NEW: AgentState struct
├── agent_state_test.go  # NEW: Tests for AgentState
├── errors.go            # MODIFIED: Add ErrInvalidAgentStatus
└── ... (existing files unchanged)

internal/core/ports/
├── agent_activity_detector.go      # NEW: Interface definition
├── agent_activity_detector_test.go # NEW: Interface documentation tests
├── waiting_detector.go             # UNCHANGED (do not modify)
└── ... (existing files unchanged)
```

### Verification Commands

After implementation, run:
```bash
go build ./internal/core/... && go test ./internal/core/... && make lint
```

All commands must pass with no errors.

### References

- [Source: docs/prd-phase2.md#Technical-Architecture] - Interface definition specification
- [Source: docs/epics-phase2.md#Epic-3-Story-3.1] - Acceptance criteria and technical notes
- [Source: docs/project-context.md#Phase-2-Additions] - Agent Detection Interface requirements
- [Source: docs/architecture.md#Implementation-Patterns] - Go code conventions and naming
- [Source: internal/core/domain/confidence.go] - Pattern for int type with iota
- [Source: internal/core/domain/stage.go] - Pattern for int type with iota
- [Source: internal/core/ports/detector.go] - Pattern for interface documentation
- [Source: internal/core/ports/log_reader.go] - Pattern for package-level docs

## Dev Agent Record

### Context Reference

- Phase 2 Epic 15: Sub-1-Minute Agent Detection (THE killer feature)
- FR Coverage: FR-P2-1, FR-P2-2, FR-P2-4

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None required - clean implementation with all tests passing.

### Completion Notes List

- Implemented `AgentStatus` type following `Confidence` pattern exactly
- Created `AgentState` struct with all helper methods and `Summary()` format
- Interface defined with comprehensive doc comments following `LogReader` pattern
- All tests follow table-driven pattern from `confidence_test.go`
- Verified hexagonal architecture: `go build ./internal/core/...` passes with zero external dependencies
- Full test suite: 1194 tests pass with no regressions
- Lint passes after `make fmt`

### File List

Files created:
- `internal/core/domain/agent_status.go`
- `internal/core/domain/agent_status_test.go`
- `internal/core/domain/agent_state.go`
- `internal/core/domain/agent_state_test.go`
- `internal/core/ports/agent_activity_detector.go`
- `internal/core/ports/agent_activity_detector_test.go`

Files modified:
- `internal/core/domain/errors.go` - Added `ErrInvalidAgentStatus`
- `docs/project-context.md` - Added Phase 2 Additions section documenting AgentActivityDetector interface and related patterns

### Change Log

- 2026-01-16: Story 15.1 implementation complete - AgentActivityDetector interface and domain types
- 2026-01-16: Code review complete - Added boundary test case for invalid AgentStatus, improved context cancellation test documentation, updated File List to include project-context.md
