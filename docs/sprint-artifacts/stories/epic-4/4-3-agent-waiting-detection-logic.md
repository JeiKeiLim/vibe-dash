# Story 4.3: Agent Waiting Detection Logic

Status: done

## Story

As a **system**,
I want **to detect when an AI agent is waiting for user input based on file inactivity**,
so that **users are alerted to blocked agents with the ⏸️ WAITING indicator, preventing lost workflow momentum (FR34-38)**.

## Acceptance Criteria

1. **AC1: Waiting State Detection Based on Inactivity Threshold**
   - Given `agent_waiting_threshold_minutes` is 10 (default from config)
   - When a project has no file activity for 10+ minutes
   - Then the project is marked as "waiting"
   - And `IsWaiting()` returns `true`

2. **AC2: Waiting State Clears on Activity**
   - Given a project is in "waiting" state
   - When file activity resumes (detected by ActivityTracker from Story 4.2)
   - Then waiting state clears automatically
   - And `IsWaiting()` returns `false`

3. **AC3: Waiting Duration Calculation**
   - Given a project is in waiting state
   - When calculating waiting duration
   - Then `WaitingDuration()` returns time since `LastActivityAt`:
     - Format: "15m" for minutes
     - Format: "2h" for hours
     - Format: "1d" for days

4. **AC4: Hibernated Projects Excluded from Waiting Detection**
   - Given a project is in `StateHibernated`
   - When checking waiting state
   - Then `IsWaiting()` returns `false`
   - And no false WAITING indicators are shown

5. **AC5: New Projects Not Initially Waiting**
   - Given a project is newly added or has no prior activity
   - When checking waiting state immediately after add
   - Then `IsWaiting()` returns `false`
   - And waiting only triggers after observed activity followed by silence

6. **AC6: Threshold Boundary Precision**
   - Given threshold is 10 minutes
   - When project is inactive for 9m59s
   - Then `IsWaiting()` returns `false`
   - When project is inactive for 10m01s
   - Then `IsWaiting()` returns `true`

7. **AC7: Per-Project Threshold Override**
   - Given a project has `agent_waiting_threshold_minutes: 5` configured
   - When project is inactive for 5+ minutes
   - Then that project shows waiting state
   - And other projects still use global 10-minute threshold

8. **AC8: Disabled Waiting Detection**
   - Given `agent_waiting_threshold_minutes: 0` (disabled)
   - When checking any project's waiting state
   - Then `IsWaiting()` always returns `false`
   - And no waiting indicators are shown

## Tasks / Subtasks

- [x] Task 1: Create WaitingDetector Service (AC: 1, 2, 5, 6, 8)
  - [x] 1.1: Create `internal/core/services/waiting_detector.go`
  - [x] 1.2: Implement `WaitingDetector` struct with `*ports.Config` dependency
  - [x] 1.3: Implement `IsWaiting(ctx context.Context, project *domain.Project) bool` method
  - [x] 1.4: Implement `WaitingDuration(ctx context.Context, project *domain.Project) time.Duration` method
  - [x] 1.5: Handle nil project (return false, log warning)
  - [x] 1.6: Handle threshold=0 (disabled)
  - [x] 1.7: Handle newly added projects (LastActivityAt == CreatedAt)

- [x] Task 2: Integrate Per-Project Threshold (AC: 7)
  - [x] 2.1: Use `config.GetEffectiveWaitingThreshold(projectID)` for per-project override
  - [x] 2.2: Fallback to global threshold if no per-project override

- [x] Task 3: Filter Hibernated Projects (AC: 4)
  - [x] 3.1: Check `project.State == StateHibernated` in IsWaiting() - return false
  - [x] 3.2: Hibernated projects never show as waiting

- [x] Task 4: Add Waiting Duration Formatting Utility (AC: 3)
  - [x] 4.1: Add `FormatWaitingDuration(d time.Duration) string` to existing `internal/shared/timeformat/timeformat.go`
  - [x] 4.2: Format: "15m", "2h", "1d" (compact for TUI)
  - [x] 4.3: Add tests to existing `timeformat_test.go`

- [x] Task 5: Write Comprehensive Tests (AC: all)
  - [x] 5.1: Unit tests for WaitingDetector in `waiting_detector_test.go`
  - [x] 5.2: Boundary tests: 9m59s vs 10m01s
  - [x] 5.3: Nil project test (must not panic)
  - [x] 5.4: Tests for hibernated projects
  - [x] 5.5: Tests for newly added projects
  - [x] 5.6: Tests for per-project threshold override
  - [x] 5.7: Tests for threshold=0 (disabled)

## Dev Notes

### Architecture Compliance

**Hexagonal Architecture Boundaries:**
- WaitingDetector: `internal/core/services/waiting_detector.go` (service layer)
- Duration formatting: `internal/shared/timeformat/` (shared utility - already exists)
- **CRITICAL:** WaitingDetector is in core/services and uses only domain entities + ports.Config

**No Repository Interface Needed:**
- Waiting detection is a pure calculation based on `project.LastActivityAt` and config threshold
- No database operations required in the detector itself
- ActivityTracker (Story 4.2) handles database updates for `LastActivityAt`

### WaitingDetector Implementation

**Complete Struct Definition:**
```go
// internal/core/services/waiting_detector.go
package services

import (
    "context"
    "log/slog"
    "time"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// WaitingDetector determines if a project's AI agent is waiting for user input.
// Stateless service - safe for concurrent use.
type WaitingDetector struct {
    config *ports.Config
    now    func() time.Time // Injected for testing
}

// NewWaitingDetector creates a new WaitingDetector with config dependency.
func NewWaitingDetector(config *ports.Config) *WaitingDetector {
    return &WaitingDetector{
        config: config,
        now:    time.Now,
    }
}

// IsWaiting returns true if the project's AI agent appears to be waiting for input.
func (d *WaitingDetector) IsWaiting(ctx context.Context, project *domain.Project) bool {
    // Nil project check
    if project == nil {
        slog.Warn("IsWaiting called with nil project")
        return false
    }

    // Hibernated projects are never "waiting"
    if project.State == domain.StateHibernated {
        return false
    }

    // Get effective threshold (per-project override or global)
    thresholdMinutes := d.config.GetEffectiveWaitingThreshold(project.ID)

    // Threshold of 0 means detection is disabled
    if thresholdMinutes == 0 {
        return false
    }

    // Newly added project: if LastActivityAt equals CreatedAt, no activity observed yet
    if project.LastActivityAt.Equal(project.CreatedAt) {
        return false
    }

    threshold := time.Duration(thresholdMinutes) * time.Minute
    inactiveDuration := d.now().Sub(project.LastActivityAt)

    return inactiveDuration >= threshold
}

// WaitingDuration returns how long the project has been in waiting state.
// Returns 0 if project is not waiting.
func (d *WaitingDetector) WaitingDuration(ctx context.Context, project *domain.Project) time.Duration {
    if !d.IsWaiting(ctx, project) {
        return 0
    }
    return d.now().Sub(project.LastActivityAt)
}
```

**Testing with Clock Mock:**
```go
// For testing, create detector then inject mock time function
detector := NewWaitingDetector(config)
detector.now = func() time.Time {
    return fixedTime // control time in tests
}
```

### Waiting Duration Formatting

**Extend existing `internal/shared/timeformat/timeformat.go`** (add to existing file, not new file):
```go
// FormatWaitingDuration returns a compact duration string for the WAITING indicator.
// Format: "15m" (minutes), "2h" (hours), "1d" (days)
// Used in TUI status column: "⏸️ WAITING 2h"
func FormatWaitingDuration(d time.Duration) string {
    switch {
    case d < time.Hour:
        return fmt.Sprintf("%dm", int(d.Minutes()))
    case d < 24*time.Hour:
        return fmt.Sprintf("%dh", int(d.Hours()))
    default:
        return fmt.Sprintf("%dd", int(d.Hours()/24))
    }
}
```
**Note:** The `fmt` import already exists in this file.

### Edge Case Handling

**Observed Activity Heuristic:** Check `LastActivityAt.Equal(CreatedAt)`. Both are set to `time.Now()` when project is added. After file activity, `LastActivityAt` diverges via ActivityTracker.

**Vacation Scenario:** Long inactivity (7d+) still shows waiting - intentional design per epics.

### Application Integration

**Wire-up in `cmd/vibe/main.go`:**
```go
// After loading config...
waitingDetector := services.NewWaitingDetector(config)

// Pass to TUI model
tuiModel := tui.NewModel(
    projectService,
    waitingDetector, // Add new dependency
    // ... other deps
)
```

**TUI Integration - `internal/adapters/tui/model.go`:**
```go
type Model struct {
    // ... existing fields
    waitingDetector *services.WaitingDetector
}

// In View() or project rendering, call:
if m.waitingDetector.IsWaiting(ctx, project) {
    duration := m.waitingDetector.WaitingDuration(ctx, project)
    formatted := timeformat.FormatWaitingDuration(duration)
    // Render "⏸️ WAITING 2h" indicator
}
```

**Note:** The TUI already has a 60-second tick (Story 4.2). Waiting state recalculates automatically on each View() call since it's based on current time minus LastActivityAt.

### File Structure

```
internal/core/services/
    waiting_detector.go           # WaitingDetector service
    waiting_detector_test.go      # Comprehensive unit tests

internal/shared/timeformat/
    timeformat.go                 # Add FormatWaitingDuration (extend existing)
    timeformat_test.go            # Add FormatWaitingDuration tests
```

### Previous Story Intelligence

**Key Learnings from Story 4.1 (File Watcher):**
1. Thread-safe operations - but WaitingDetector is stateless, no mutex needed
2. Context check pattern - not needed here as this is pure calculation
3. Testable time - use `now func() time.Time` for clock control in tests

**Key Learnings from Story 4.2 (Activity Tracking):**
1. `LastActivityAt` is updated via `repo.UpdateLastActivity()` on file events
2. Timestamps are in UTC (RFC3339 format in DB)
3. `FormatRelativeTime` already exists in `internal/shared/timeformat/`
4. 60-second tick refreshes timestamps in TUI (View() calls formatting functions)

**Event Paths are Already Canonical:**
Story 4.1 ensures all file event paths are canonicalized before emitting to ActivityTracker.

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Store "isWaiting" in database | Calculate dynamically from LastActivityAt |
| Modify project entity in detector | Return bool/duration, don't mutate |
| Hard-code threshold value | Use config.GetEffectiveWaitingThreshold() |
| Use time.Now() directly | Inject `now` function for testability |
| Check waiting for nil project | Return false, log warning |
| Ignore hibernated state | Always check State first |

### Testing Strategy

**Unit Tests (waiting_detector_test.go):**
```go
func TestIsWaiting(t *testing.T) {
    tests := []struct {
        name               string
        lastActivityAgo    time.Duration
        createdAgo         time.Duration
        state              domain.ProjectState
        thresholdMinutes   int
        expected           bool
    }{
        // Boundary tests
        {"under threshold 9m59s", 9*time.Minute + 59*time.Second, 1*time.Hour, domain.StateActive, 10, false},
        {"at threshold 10m", 10 * time.Minute, 1*time.Hour, domain.StateActive, 10, true},
        {"over threshold 10m01s", 10*time.Minute + 1*time.Second, 1*time.Hour, domain.StateActive, 10, true},

        // Edge cases
        {"nil project", 0, 0, domain.StateActive, 10, false}, // Must not panic
        {"newly added project", 5*time.Minute, 5*time.Minute, domain.StateActive, 10, false},
        {"hibernated project", 1*time.Hour, 2*time.Hour, domain.StateHibernated, 10, false},
        {"threshold disabled", 1*time.Hour, 2*time.Hour, domain.StateActive, 0, false},

        // Per-project threshold
        {"custom 5min threshold", 6*time.Minute, 1*time.Hour, domain.StateActive, 5, true},
        {"under custom threshold", 4*time.Minute, 1*time.Hour, domain.StateActive, 5, false},
    }
    // ...
}

func TestIsWaiting_NilProject(t *testing.T) {
    config := ports.NewConfig()
    detector := NewWaitingDetector(config)

    // Must not panic, should return false
    result := detector.IsWaiting(context.Background(), nil)
    if result != false {
        t.Errorf("IsWaiting(nil) = %v, want false", result)
    }
}
```

**Duration Formatting Tests:**
```go
func TestFormatWaitingDuration(t *testing.T) {
    tests := []struct {
        duration time.Duration
        expected string
    }{
        {0, "0m"},
        {5 * time.Minute, "5m"},
        {59 * time.Minute, "59m"},
        {60 * time.Minute, "1h"},
        {23 * time.Hour, "23h"},
        {24 * time.Hour, "1d"},
        {48 * time.Hour, "2d"},
    }
    // ...
}
```

### Downstream Dependencies

**Story 4.4 (Waiting Threshold Configuration) depends on this story for:**
- Core detection logic to consume threshold values
- Per-project override support via `GetEffectiveWaitingThreshold()`

**Story 4.5 (Waiting Indicator Display) depends on this story for:**
- `IsWaiting()` to determine when to show ⏸️ WAITING
- `WaitingDuration()` to show elapsed time like "2h"
- `FormatWaitingDuration()` for compact display

**Story 4.6 (Real-Time Dashboard Updates) depends on this story for:**
- Waiting state recalculation on each TUI tick
- Dynamic waiting state changes as time passes

### Key Code References

| File | Relevant Content |
|------|------------------|
| `internal/core/ports/config.go:97-104` | `GetEffectiveWaitingThreshold()` method |
| `internal/core/domain/project.go` | `Project.LastActivityAt`, `Project.State`, `Project.CreatedAt` |
| `internal/core/domain/state.go:13` | `StateHibernated` constant |
| `internal/shared/timeformat/timeformat.go` | Existing `FormatRelativeTime()` - extend with `FormatWaitingDuration()` |
| `internal/core/services/activity_tracker.go` | Story 4.2 - updates `LastActivityAt` |

### Manual Testing Steps

After implementation, verify:

1. **Basic Waiting Detection:**
   ```bash
   # Add a test project
   ./bin/vibe add /tmp/test-project

   # Touch a file (creates activity)
   touch /tmp/test-project/test.txt

   # Wait 10+ minutes, or mock time in test
   # Verify IsWaiting() returns true
   ```

2. **Waiting Clears on Activity:**
   ```bash
   # While project shows as waiting
   touch /tmp/test-project/another.txt
   # Verify IsWaiting() now returns false
   ```

3. **Hibernated Projects Excluded:**
   ```bash
   # Hibernate a project that would otherwise be waiting
   ./bin/vibe hibernate test-project
   # Verify IsWaiting() returns false despite inactivity
   ```

4. **Threshold Boundary:**
   ```bash
   # Unit tests verify 9m59s vs 10m01s boundary
   go test -v ./internal/core/services/... -run TestIsWaiting
   ```

5. **Duration Format:**
   ```bash
   # Unit tests verify format outputs
   go test -v ./internal/shared/timeformat/... -run TestFormatWaitingDuration
   ```

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

1. WaitingDetector service implemented with stateless design - calculates waiting state dynamically from `LastActivityAt` and current time
2. Per-project threshold override integrated via `config.GetEffectiveWaitingThreshold(project.ID)`
3. Hibernated projects correctly excluded from waiting detection
4. Newly added projects (LastActivityAt == CreatedAt) not falsely flagged as waiting
5. Boundary precision verified: 9m59s returns false, 10m00s returns true
6. FormatWaitingDuration added to timeformat package with compact format: "15m", "2h", "1d"
7. Comprehensive test coverage with table-driven tests for all edge cases
8. **Code Review Fixes:** Added negative duration clamping (defensive programming), improved test clarity with comments, added context parameter documentation

### File List

**New Files:**
- `internal/core/services/waiting_detector.go` - WaitingDetector service
- `internal/core/services/waiting_detector_test.go` - Comprehensive unit tests

**Modified Files:**
- `internal/shared/timeformat/timeformat.go` - Add FormatWaitingDuration function
- `internal/shared/timeformat/timeformat_test.go` - Add FormatWaitingDuration tests

### Change Log

| Date | Author | Change |
|------|--------|--------|
| 2025-12-19 | SM (Bob) | Initial story creation via *create-story workflow |
| 2025-12-19 | SM (Bob) | Validation: Added context.Context to method signatures, config as constructor dependency, nil project handling, TUI integration section, clarified task numbering |
| 2025-12-19 | Dev (Amelia) | Implementation complete: WaitingDetector service, FormatWaitingDuration utility, comprehensive tests - all ACs satisfied |
| 2025-12-19 | Dev (Amelia) | Code review fixes: Added negative duration clamping to FormatWaitingDuration, improved test comments/clarity, added context param documentation |
