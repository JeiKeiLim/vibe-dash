# Story 16.2: Implement Stage Transition Event Recording

Status: Done

## Story

As a developer,
I want the system to record when projects change stages,
So that metrics can be calculated from historical data.

## User-Visible Changes

None - this is an internal infrastructure change. Stage transitions are recorded to the metrics database for future stats view visualization (Story 16.3+). Users will not see any immediate change in dashboard behavior.

## Acceptance Criteria

1. **Given** project stage changes from "Plan" to "Tasks"
   **When** detection service runs (via TUI refresh)
   **Then** stage_transitions record is created with from_stage="Plan", to_stage="Tasks"

2. **And** given metrics recording fails (e.g., disk full, permission denied)
   **When** stage transition happens
   **Then** core functionality continues (graceful degradation - no crash, warning logged)

3. **And** events include ISO 8601 timestamp (time.RFC3339Nano format)

4. **And** given first detection for a project (no previous stage)
   **When** stage transition happens
   **Then** from_stage is empty string ("") and to_stage is the detected stage

5. **And** given project stage is unchanged (e.g., still "Plan")
   **When** detection runs
   **Then** NO transition event is recorded (avoid duplicate records)

6. **And** debounce rapid transitions to prevent excessive event recording
   **When** same project transitions multiple times within 10 seconds
   **Then** only the final transition is recorded (configurable threshold)

## Tasks / Subtasks

- [x] Task 1: Create MetricsRecorder component (AC: #1, #2, #3)
  - [x] Create `internal/adapters/persistence/metrics/recorder.go` (same package as repository)
  - [x] Define `MetricsRecorder` struct with fields:
    - `repo *MetricsRepository`
    - `previousStages map[string]domain.Stage`
    - `pendingTransitions map[string]*pendingTransition`
    - `debounceWindow time.Duration` (default 10 seconds)
    - `mu sync.Mutex`
  - [x] Constructor: `NewMetricsRecorder(repo *MetricsRepository) *MetricsRecorder`
  - [x] All errors handled gracefully (log + return nil)
  - [x] **Project ID:** Use `project.Path` (canonical path) - consistent across sessions

- [x] Task 2: Implement stage change detection logic (AC: #4, #5)
  - [x] Method: `OnDetection(ctx context.Context, projectID string, newStage domain.Stage)` - main entry point
  - [x] Compare with previousStages map - if different, call scheduleTransition
  - [x] Handle first detection: `fromStage=""` when no previous stage recorded
  - [x] Update previousStages map after scheduling

- [x] Task 3: Implement debouncing for rapid transitions (AC: #6)
  - [x] Define `pendingTransition` struct: `fromStage`, `toStage domain.Stage`, `timer *time.Timer`
  - [x] When transition detected, store in pendingTransitions with timer
  - [x] Timer callback uses `context.Background()` (not passed ctx which may be cancelled)
  - [x] When timer fires, commit to database via `commitTransition()`
  - [x] If new transition arrives during window, stop existing timer, update toStage (keep original fromStage)
  - [x] Add `Flush(ctx context.Context)` method to force commit pending (for shutdown)

- [x] Task 4: Wire MetricsRecorder into TUI via CLI setter pattern (AC: #1)
  - [x] Add `cli.SetMetricsRecorder(recorder *metrics.MetricsRecorder)` function in `internal/adapters/cli/deps.go`
  - [x] Add `metricsRecorder *metrics.MetricsRecorder` field to TUI Model struct
  - [x] Add `SetMetricsRecorder(recorder *metrics.MetricsRecorder)` method on Model
  - [x] In `refreshProjectsCmd()`: After detecting stage change, call `m.metricsRecorder.OnDetection(ctx, project.Path, newStage)`
  - [x] Check `m.metricsRecorder != nil` before calling (optional dependency)

- [x] Task 5: Wire MetricsRecorder in main.go (AC: #1)
  - [x] Create MetricsRepository: `metricsDBPath := filepath.Join(basePath, "metrics.db")`
  - [x] Create MetricsRecorder: `metricsRecorder := metrics.NewMetricsRecorder(metricsRepo)`
  - [x] Pass to CLI: `cli.SetMetricsRecorder(metricsRecorder)`
  - [x] Add shutdown hook: `metricsRecorder.Flush(context.Background())` in defer before coordinator.Close

- [x] Task 6: Write comprehensive unit tests
  - [x] `recorder_test.go` in `internal/adapters/persistence/metrics/`
  - [x] `TestOnDetection_FirstDetection` - from_stage should be empty
  - [x] `TestOnDetection_StageChange` - transition scheduled
  - [x] `TestOnDetection_NoChange` - no transition scheduled
  - [x] `TestOnDetection_Debouncing` - rapid changes result in single record after window
  - [x] `TestOnDetection_NilRepository` - graceful no-op
  - [x] `TestFlush_CommitsPending` - all pending transitions committed
  - [x] `TestFlush_CancelsTimers` - no goroutine leaks after flush

- [x] Task 7: Write integration tests
  - [x] `recorder_integration_test.go` with `//go:build integration` tag
  - [x] Test end-to-end: OnDetection → Flush → query database → verify records
  - [x] Test concurrent transitions from multiple projects

## Dev Notes

### Architecture Alignment

This story extends the metrics infrastructure from Story 16.1. Both components live in the same package:
- **MetricsRepository** (Story 16.1): Low-level database access in `internal/adapters/persistence/metrics/repository.go`
- **MetricsRecorder** (Story 16.2): High-level event handling in `internal/adapters/persistence/metrics/recorder.go`

The recorder sits between the TUI and repository, providing:
- Stage change detection by comparing previous vs current
- Debouncing to prevent database spam during rapid stage changes
- Graceful degradation consistent with Story 16.1 patterns

### TUI Integration Pattern

**Follow existing setter pattern** (NOT constructor modification):

```go
// internal/adapters/cli/deps.go - add near other Set* functions
var metricsRecorder *metrics.MetricsRecorder

func SetMetricsRecorder(recorder *metrics.MetricsRecorder) {
    metricsRecorder = recorder
}

func GetMetricsRecorder() *metrics.MetricsRecorder {
    return metricsRecorder
}
```

```go
// internal/adapters/tui/model.go - add method
func (m *Model) SetMetricsRecorder(recorder *metrics.MetricsRecorder) {
    m.metricsRecorder = recorder
}
```

### main.go Wiring Location

Add after HibernationService setup (around line 185):

```go
// Story 16.2: Create MetricsRecorder for stage transition tracking
metricsDBPath := filepath.Join(basePath, "metrics.db")
metricsRepo := metrics.NewMetricsRepository(metricsDBPath)
metricsRecorder := metrics.NewMetricsRecorder(metricsRepo)
cli.SetMetricsRecorder(metricsRecorder)

// Add to shutdown sequence (before coordinator.Close defer)
defer func() {
    metricsRecorder.Flush(context.Background())
}()

slog.Debug("metrics recorder initialized", "db_path", metricsDBPath)
```

### MetricsRecorder Implementation

```go
package metrics

import (
    "context"
    "log/slog"
    "sync"
    "time"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

const DefaultDebounceWindow = 10 * time.Second

type pendingTransition struct {
    fromStage domain.Stage
    toStage   domain.Stage
    timer     *time.Timer
}

type MetricsRecorder struct {
    repo               *MetricsRepository
    previousStages     map[string]domain.Stage
    pendingTransitions map[string]*pendingTransition
    debounceWindow     time.Duration
    mu                 sync.Mutex
}

func NewMetricsRecorder(repo *MetricsRepository) *MetricsRecorder {
    return &MetricsRecorder{
        repo:               repo,
        previousStages:     make(map[string]domain.Stage),
        pendingTransitions: make(map[string]*pendingTransition),
        debounceWindow:     DefaultDebounceWindow,
    }
}

func (r *MetricsRecorder) OnDetection(ctx context.Context, projectID string, newStage domain.Stage) {
    if r == nil || r.repo == nil {
        return
    }

    r.mu.Lock()
    defer r.mu.Unlock()

    prevStage, exists := r.previousStages[projectID]
    if exists && prevStage == newStage {
        return
    }

    fromStage := domain.Stage("")
    if exists {
        fromStage = prevStage
    }

    r.scheduleTransition(projectID, fromStage, newStage)
    r.previousStages[projectID] = newStage
}

func (r *MetricsRecorder) scheduleTransition(projectID string, fromStage, toStage domain.Stage) {
    if pending, exists := r.pendingTransitions[projectID]; exists {
        pending.timer.Stop()
        pending.toStage = toStage
    } else {
        r.pendingTransitions[projectID] = &pendingTransition{
            fromStage: fromStage,
            toStage:   toStage,
        }
    }

    pending := r.pendingTransitions[projectID]
    pending.timer = time.AfterFunc(r.debounceWindow, func() {
        r.commitTransition(projectID)
    })
}

func (r *MetricsRecorder) commitTransition(projectID string) {
    r.mu.Lock()
    pending, exists := r.pendingTransitions[projectID]
    if !exists {
        r.mu.Unlock()
        return
    }
    fromStage := string(pending.fromStage)
    toStage := string(pending.toStage)
    delete(r.pendingTransitions, projectID)
    r.mu.Unlock()

    _ = r.repo.RecordTransition(context.Background(), projectID, fromStage, toStage)
}

func (r *MetricsRecorder) Flush(ctx context.Context) {
    if r == nil {
        return
    }

    r.mu.Lock()
    projectIDs := make([]string, 0, len(r.pendingTransitions))
    for id, pending := range r.pendingTransitions {
        pending.timer.Stop()
        projectIDs = append(projectIDs, id)
    }
    r.mu.Unlock()

    for _, id := range projectIDs {
        r.commitTransition(id)
    }
}
```

### TUI refresh Integration Point

Location: `internal/adapters/tui/model.go` in `refreshProjectsCmd()` around line 810-815.

**Critical pattern:** The current code compares `project.CurrentStage` (before) with `primary.Stage` (after). You need to capture the stage change BEFORE updating the project:

```go
// In refreshProjectsCmd, after getting detection results:
oldStage := currentProject.CurrentStage // Capture BEFORE update

// ... existing code that updates currentProject.CurrentStage = primary.Stage ...

// After updating project, record transition if metrics enabled
if m.metricsRecorder != nil && oldStage != primary.Stage {
    m.metricsRecorder.OnDetection(ctx, currentProject.Path, primary.Stage)
}
```

### Graceful Degradation Patterns

Following Story 16.1 established patterns:
1. `MetricsRecorder` accepts nil repository → all operations become no-ops
2. Check `r == nil` at start of public methods (nil receiver safety)
3. All public methods return void → never crash callers
4. Errors logged with `slog.Warn()` including context
5. Core TUI functionality unaffected by metrics failures

### NFR Compliance

- **NFR-P2-7:** Metrics failure does not affect core dashboard
  - Recorder handles nil repository
  - All errors logged but not propagated
- **Database Growth:** Debouncing reduces event volume
  - Without debouncing: Events every refresh (10s interval)
  - With debouncing (10s window): At most 1 event per 10s per project

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - No debug issues encountered

### Completion Notes List

- Implemented MetricsRecorder component with stage change detection and debouncing
- Used `transitionRecorder` interface internally for testability (accepts any type implementing `RecordTransition`)
- TUI integration uses `metricsRecorderInterface` interface to decouple from concrete type
- OnDetection called in refreshProjectsCmd() BEFORE updating currentProject.CurrentStage
- Added Flush() call in main.go shutdown sequence to ensure pending transitions are committed
- All 9 unit tests pass (TestOnDetection_*, TestFlush_*, TestOnDetection_MultipleProjects)
- All 4 integration tests pass (EndToEnd, ConcurrentProjects, RFC3339Nano_Timestamp, GracefulDegradation)
- Full test suite passes (1360 tests)
- Lint and format checks pass

**Code Review Fixes (2026-01-16):**
- Fixed CRITICAL bug: metricsRecorder was not wired to TUI model (metrics were never recorded!)
- Added metricsRecorder parameter to `tui.Run()` function signature
- Updated `cli/root.go` to pass metricsRecorder to TUI
- Fixed misleading comment in model.go refreshProjectsCmd

### File List

- internal/adapters/persistence/metrics/recorder.go (new)
- internal/adapters/persistence/metrics/recorder_test.go (new)
- internal/adapters/persistence/metrics/recorder_integration_test.go (new)
- internal/adapters/cli/deps.go (modified)
- internal/adapters/cli/root.go (modified - code review fix)
- internal/adapters/tui/app.go (modified - code review fix)
- internal/adapters/tui/model.go (modified)
- cmd/vdash/main.go (modified)
