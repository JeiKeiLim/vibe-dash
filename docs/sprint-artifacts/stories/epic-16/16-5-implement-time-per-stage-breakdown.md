# Story 16.5: Implement Time-Per-Stage Breakdown

Status: done

## Story

As a user,
I want to see how much time I spend in each stage,
So that I can identify bottlenecks in my workflow.

## User-Visible Changes

- **New:** Stats View displays time-per-stage breakdown for the selected project
- **New:** Shows duration spent in each stage (e.g., "Plan: 3h | Tasks: 1h | Implement: 5h")
- **New:** Current (in-progress) stage shows time since last transition
- **New:** Press `Enter` on a project to see detailed breakdown, `Esc` to return to list

## Acceptance Criteria

1. **Given** project has transitions: Plan(3h) → Tasks(1h) → Implement(5h)
   **When** Stats View shows breakdown
   **Then** displays: "Plan: 3h | Tasks: 1h | Implement: 5h"

2. **And** breakdown shows percentage or bar chart representation
   - Visual representation: horizontal bars or percentage values
   - Bar width proportional to time spent

3. **And** handles in-progress stage (current stage shows time since last transition)
   - Current stage shows elapsed time since last recorded transition
   - Marked with indicator (e.g., "→" prefix or highlight) to show it's active

4. **And** Stats View continues to work when metrics.db is unavailable (graceful degradation)
   - Shows "No stage data available" message when no metrics exist

5. **And** supports keyboard navigation
   - `Enter` on project: show detailed breakdown view
   - `Esc`: return to project list
   - `j`/`k` or `↑`/`↓`: navigate between projects

## Tasks / Subtasks

- [x] Task 1: Create time calculation module (AC: #1, #3)
  - [x]Create `internal/adapters/tui/statsview/breakdown.go`:
    ```go
    package statsview

    import "time"

    // StageDuration represents time spent in a single stage.
    type StageDuration struct {
        Stage    string        // Stage name (e.g., "Plan", "Tasks", "Implement")
        Duration time.Duration // Time spent in this stage
        IsCurrent bool         // True if this is the current (in-progress) stage
    }

    // CalculateStageDurations computes time spent in each stage from transitions.
    // transitions: stage transition records (oldest first - ASC order)
    // now: reference time for current stage calculation (enables testing)
    // Returns: slice of StageDuration (one per unique stage visited)
    func CalculateStageDurations(transitions []Transition, now time.Time) []StageDuration

    // FullTransition includes full stage transition data for breakdown calculation.
    type FullTransition struct {
        FromStage      string
        ToStage        string
        TransitionedAt time.Time
    }

    // CalculateFromFullTransitions computes durations from full transition data.
    // Needed because we need to_stage to know what stage we transitioned INTO.
    func CalculateFromFullTransitions(transitions []FullTransition, now time.Time) []StageDuration
    ```
  - [x]Implementation:
    - Iterate through transitions chronologically
    - Calculate duration between consecutive transitions
    - Last stage (to_stage of final transition) is current - calculate time to `now`
    - Handle edge case: no transitions → return empty slice
    - Handle edge case: single transition → current stage from transition start to now
  - [x]Duration calculation algorithm:
    ```go
    // For transitions: t1(→A), t2(→B), t3(→C)
    // A duration = t2.time - t1.time
    // B duration = t3.time - t2.time
    // C duration = now - t3.time (current stage)
    ```

- [x] Task 2: Add repository method for full transitions (AC: #1)
  - [x]In `internal/adapters/tui/statsview/interfaces.go`, add new interface method:
    ```go
    // FullTransitionReader provides full transition data for breakdown calculation.
    // Extends MetricsReader with richer data needed for time calculations.
    type FullTransitionReader interface {
        MetricsReader
        // GetFullTransitions retrieves full transition data for a project since given time.
        // Returns empty slice on any error (graceful degradation).
        GetFullTransitions(ctx context.Context, projectID string, since time.Time) []FullTransition
    }
    ```
  - [x]In `internal/adapters/persistence/metrics/repository.go`, add:
    ```go
    // GetFullTransitions retrieves full transition data for breakdown display.
    // Implements statsview.FullTransitionReader interface.
    func (r *MetricsRepository) GetFullTransitions(ctx context.Context, projectID string, since time.Time) []statsview.FullTransition
    ```
  - [x]Convert from `StageTransition` to `statsview.FullTransition`

- [x] Task 3: Create breakdown rendering module (AC: #1, #2)
  - [x]In `internal/adapters/tui/statsview/breakdown.go`, add rendering:
    ```go
    // RenderBreakdown renders time-per-stage as text with horizontal bars.
    // maxWidth: available terminal width for rendering
    // Returns formatted string with stage names, durations, and bars.
    func RenderBreakdown(durations []StageDuration, maxWidth int) string

    // formatDuration formats time.Duration for display.
    // Examples: "3h", "45m", "2h 30m", "< 1m"
    func formatDuration(d time.Duration) string
    ```
  - [x]Bar rendering:
    - Calculate max bar width (e.g., maxWidth - stageNameWidth - durationWidth - padding)
    - Normalize durations to bar widths (longest = full width)
    - Use box drawing characters: `█` for filled, `░` for empty
  - [x]Output format example:
    ```
    Plan:      3h 15m  ████████░░░░░░░░░░░  (32%)
    Tasks:     1h 20m  ███░░░░░░░░░░░░░░░░  (13%)
    → Implement: 5h 30m  █████████████████░░  (55%)
    ```

- [x] Task 4: Add breakdown view state to Model (AC: #5)
  - [x]In `internal/adapters/tui/model.go`, add state fields:
    ```go
    // Stats View breakdown state (Story 16.5)
    statsBreakdownProject   *domain.Project    // Currently selected project for breakdown (nil = list view)
    statsBreakdownDurations []statsview.StageDuration // Cached durations for display
    ```
  - [x]Add method to fetch breakdown data:
    ```go
    // getStageBreakdown fetches stage durations for a project (Story 16.5).
    // Returns nil if metricsReader is not a FullTransitionReader or unavailable.
    func (m Model) getStageBreakdown(projectID string) []statsview.StageDuration
    ```

- [x] Task 5: Update Stats View to support breakdown view (AC: #1, #2, #3, #5)
  - [x]In `internal/adapters/tui/statsview.go`:
    - Add `renderStatsBreakdownView(project *domain.Project, durations []statsview.StageDuration) string` method
    - Modify `renderStatsView()` to check `m.statsBreakdownProject`:
      ```go
      func (m Model) renderStatsView() string {
          if m.statsBreakdownProject != nil {
              return m.renderStatsBreakdownView()
          }
          return m.renderStatsProjectList(...) // existing list rendering
      }
      ```
    - Key handling is in model.go Update method (see Task 6), NOT here
  - [x]Breakdown view layout:
    ```
    STATS                          [ESC] Back to Project List

    Project: my-project
    Period: Last 30 days

    ─────────────────────────────────────────────────

    Stage Breakdown:

    Plan:        3h 15m  ████████░░░░░░░░░░░  (32%)
    Tasks:       1h 20m  ███░░░░░░░░░░░░░░░░  (13%)
    → Implement: 5h 30m  █████████████████░░  (55%)

    Total: 10h 5m

    ─────────────────────────────────────────────────
    ```

- [x] Task 6: Update key bindings for breakdown navigation (AC: #5)
  - [x]In `internal/adapters/tui/model.go`, add breakdown key handling in the existing Update method:
    - Locate the `case viewModeStats:` block in the Update method
    - Add breakdown-aware key handling:
    ```go
    // In model.go Update method, within case viewModeStats:
    case "esc", "q":
        if m.statsBreakdownProject != nil {
            // Exit breakdown view, return to project list
            m.statsBreakdownProject = nil
            m.statsBreakdownDurations = nil
            return m, nil
        }
        // No breakdown active - exit Stats View entirely (existing behavior)
        m.exitStatsView()
        return m, nil
    case "enter":
        if m.statsBreakdownProject == nil && len(m.projects) > 0 {
            // Enter breakdown view for selected project
            // Use statsViewScroll to find currently visible/selected project
            selectedIdx := m.statsViewScroll // Or implement proper selection tracking
            if selectedIdx >= 0 && selectedIdx < len(m.projects) {
                p := m.projects[selectedIdx]
                m.statsBreakdownProject = p
                m.statsBreakdownDurations = m.getStageBreakdown(p.Path)
            }
        }
        return m, nil
    ```
  - [x]DO NOT create `statsview_update.go` - all key handling stays in model.go Update method

- [x] Task 7: Update Model interface to use FullTransitionReader (AC: #1)
  - [x]In `internal/adapters/tui/model.go`, update interface:
    ```go
    // Change metricsReaderInterface to support both sparklines and breakdown
    type metricsReaderInterface interface {
        GetTransitionTimestamps(ctx context.Context, projectID string, since time.Time) []statsview.Transition
        GetFullTransitions(ctx context.Context, projectID string, since time.Time) []statsview.FullTransition
    }
    ```
  - [x]MetricsRepository already implements both methods, no changes needed there

- [x] Task 8: Write comprehensive unit tests
  - [x]`internal/adapters/tui/statsview/breakdown_test.go`:
    ```go
    func TestCalculateFromFullTransitions_NoTransitions(t *testing.T) { ... }
    func TestCalculateFromFullTransitions_SingleTransition(t *testing.T) { ... }
    func TestCalculateFromFullTransitions_MultipleTransitions(t *testing.T) { ... }
    func TestCalculateFromFullTransitions_CurrentStageCalculation(t *testing.T) { ... }
    func TestRenderBreakdown_EmptyDurations(t *testing.T) { ... }
    func TestRenderBreakdown_WithBars(t *testing.T) { ... }
    func TestFormatDuration_Various(t *testing.T) { ... }
    ```
  - [x]`internal/adapters/persistence/metrics/repository_test.go`:
    - Add tests for `GetFullTransitions` method

- [x] Task 9: Update statsview_test.go for breakdown integration
  - [x]Test Stats View key handling (Enter → breakdown, Esc → back)
  - [x]Test breakdown rendering with mock data
  - [x]Test graceful degradation when metricsReader is nil

## Dev Notes

### Architecture Alignment

Implements **FR-P2-17** from Phase 2 PRD: "System can display time-per-stage breakdown"

**Isolation Principle:** Breakdown feature builds on existing Stats View infrastructure:
- Reuses existing `metricsReader` pattern from Story 16.4
- Extends interface (backward compatible) for full transition data
- Separate breakdown rendering in `statsview/breakdown.go`

**Import Direction (Same as Story 16.4):**
```
internal/adapters/persistence/metrics
    └── imports → internal/adapters/tui/statsview (for FullTransition type)

internal/adapters/tui
    └── imports → internal/adapters/tui/statsview (for interface and types)

internal/adapters/tui/statsview
    └── NO imports from persistence layer (clean separation)
```

### File Modifications

| File | Changes |
|------|---------|
| `internal/adapters/tui/statsview/interfaces.go` | Add `FullTransition` type, `FullTransitionReader` interface |
| `internal/adapters/tui/statsview/breakdown.go` | **NEW:** Time calculation (`CalculateFromFullTransitions`), rendering (`RenderBreakdown`) |
| `internal/adapters/persistence/metrics/repository.go` | Add `GetFullTransitions` method |
| `internal/adapters/tui/statsview.go` | Add `renderStatsBreakdownView()`, modify `renderStatsView()` for view switching |
| `internal/adapters/tui/model.go` | Add `statsBreakdownProject`, `statsBreakdownDurations` fields, `getStageBreakdown()` method, update `metricsReaderInterface`, breakdown key handling in Update method |

**DO NOT CREATE:**
- `statsview_update.go` - key handling goes in model.go Update method

### Breakdown View Navigation Flow

```
Dashboard (viewMode=viewModeNormal)
        │
        │ [s] key → enterStatsView() saves statsActiveProjectIdx
        ▼
Stats View (viewMode=viewModeStats, statsBreakdownProject=nil)
        │
        │ [Enter] on project → sets statsBreakdownProject
        ▼
Stats View Breakdown (viewMode=viewModeStats, statsBreakdownProject!=nil)
        │
        │ [Esc] or [q] → clears statsBreakdownProject
        ▼
Stats View (viewMode=viewModeStats, statsBreakdownProject=nil)
        │
        │ [Esc] or [q] → exitStatsView() restores statsActiveProjectIdx
        ▼
Dashboard (viewMode=viewModeNormal)
```

**Key Insight:** Breakdown view is a SUB-STATE of `viewModeStats`, not a separate viewMode.
Check `m.statsBreakdownProject != nil` to determine if in breakdown detail view.

### Duration Calculation Details

**Algorithm for calculating stage durations:**

```go
func CalculateFromFullTransitions(transitions []FullTransition, now time.Time) []StageDuration {
    if len(transitions) == 0 {
        return nil
    }

    // Map to accumulate time per stage
    stageTimes := make(map[string]time.Duration)
    var lastStage string
    var lastTime time.Time

    for i, t := range transitions {
        if i > 0 {
            // Add duration to previous stage
            duration := t.TransitionedAt.Sub(lastTime)
            stageTimes[lastStage] += duration
        }
        lastStage = t.ToStage
        lastTime = t.TransitionedAt
    }

    // Add current stage duration (time since last transition)
    if lastStage != "" {
        stageTimes[lastStage] += now.Sub(lastTime)
    }

    // Convert to slice, mark last stage as current
    var result []StageDuration
    for stage, duration := range stageTimes {
        result = append(result, StageDuration{
            Stage:     stage,
            Duration:  duration,
            IsCurrent: stage == lastStage,
        })
    }

    // Sort by total duration (descending) or by occurrence order
    // Consider sorting by first occurrence for logical flow
    sort.Slice(result, func(i, j int) bool {
        return result[i].Duration > result[j].Duration
    })

    return result
}
```

**Edge Cases:**
1. **No transitions:** Return empty slice, display "No stage data available"
2. **Single transition:** One stage with duration = now - transition time (current)
3. **Revisited stages:** Durations accumulate (e.g., Plan→Tasks→Plan→Implement means Plan has combined time)
4. **Very long durations:** Format as days (e.g., "2d 5h") when duration exceeds 24 hours
5. **Zero duration stages:** May occur with rapid transitions - show "< 1m"

**Stage Ordering Options:**
- **Duration descending** (current algorithm): Shows largest time sinks first
- **First occurrence order**: More logical workflow view (Plan → Tasks → Implement)
- Implementation uses duration descending; consider making configurable in future

### Bar Rendering Details

**Bar character selection:**
- `█` (U+2588 FULL BLOCK) for filled portion
- `░` (U+2591 LIGHT SHADE) for unfilled portion

**Width calculation:**
```go
const maxBarWidth = 20  // Fixed bar width for consistency
const paddingWidth = 4  // Space between columns

// stageWidth: longest stage name + 2
// durationWidth: longest duration string + 2
// barWidth: maxBarWidth (fixed)
// percentWidth: 6 chars for "(XX%)"
```

### NFR Compliance

| Requirement | Target | This Story |
|-------------|--------|------------|
| NFR-P2-5 | Stats View render < 500ms | Breakdown calculation O(n) where n = transitions |
| NFR-P2-7 | Metrics failure doesn't crash | nil-safe checks, graceful degradation |

### Critical Implementation Notes

**DO NOT:**
- Create `statsview_update.go` - all key handling stays in model.go
- Use a separate `viewMode` for breakdown - it's a sub-state of `viewModeStats`
- Block UI while fetching transitions - cache in `statsBreakdownDurations`

**MUST:**
- Check `m.statsBreakdownProject != nil` to determine breakdown vs list view
- Clear `statsBreakdownDurations` when exiting breakdown view
- Reuse existing `enterStatsView()` / `exitStatsView()` methods from Story 16.4
- Ensure `Esc` in breakdown returns to list (not dashboard)

### Previous Story Patterns

**From Story 16.4 (Sparklines):**
- `metricsReader` interface pattern for dependency injection
- Graceful degradation: return nil on errors
- `getProjectActivity()` pattern for fetching metrics data
- Stats View layout calculations (header, content height, width adaptation)

**From Story 16.3 (Stats View):**
- View switching pattern (Dashboard ↔ Stats View)
- Key binding integration in Update method
- Scroll state management (`statsViewScroll`)
- `viewModeStats` constant for view mode switching

**From Story 16.4 (View Switching Methods):**
- `enterStatsView()` method saves dashboard selection in `statsActiveProjectIdx`
- `exitStatsView()` method restores dashboard selection from `statsActiveProjectIdx`
- Breakdown view is a SUB-STATE of `viewModeStats` (check `statsBreakdownProject != nil`)

### Testing Strategy

**Unit Tests Focus:**
1. Duration calculation correctness with various transition patterns
2. Bar rendering with edge cases (empty, single value, max values)
3. Duration formatting (hours, minutes, combined)
4. Repository method returning correct data structure
5. Breakdown view state transitions (enter/exit breakdown)

**Mock Pattern (from Story 16.4 statsview_test.go):**
```go
// mockMetricsReader implements metricsReaderInterface for testing
type mockMetricsReader struct {
    transitions     []statsview.Transition
    fullTransitions []statsview.FullTransition
}

func (m *mockMetricsReader) GetTransitionTimestamps(ctx context.Context, projectID string, since time.Time) []statsview.Transition {
    return m.transitions
}

func (m *mockMetricsReader) GetFullTransitions(ctx context.Context, projectID string, since time.Time) []statsview.FullTransition {
    return m.fullTransitions
}
```

**Manual Verification:**
- Add transitions via normal vdash usage
- Press `s` to enter Stats View
- Press `Enter` on project to see breakdown
- Verify durations match expected values
- Press `Esc` to return to list (not dashboard)
- Press `Esc` again to return to dashboard
- Verify dashboard selection is preserved
- Test with project that has no metrics data

### Project Structure Notes

**New Files:**
- `internal/adapters/tui/statsview/breakdown.go` - Core breakdown logic
- `internal/adapters/tui/statsview/breakdown_test.go` - Tests for breakdown

**Modified Files:**
- `internal/adapters/tui/statsview/interfaces.go` - Add `FullTransition` type, `FullTransitionReader` interface
- `internal/adapters/tui/statsview.go` - Add breakdown view rendering
- `internal/adapters/tui/model.go` - Add state fields, interface update, key handling, `getStageBreakdown()` method
- `internal/adapters/persistence/metrics/repository.go` - Add `GetFullTransitions` method

### References

- [Source: docs/epics-phase2.md#Story 4.5]
- [Source: docs/project-context.md#Phase 2 Additions - Progress Metrics]
- [Source: docs/architecture.md#Hexagonal Architecture Boundaries]
- [Source: docs/sprint-artifacts/stories/epic-16/16-4-implement-activity-sparklines.md]

## Dev Agent Record

### Context Reference

- Story 16.4 for sparkline patterns, metrics reader wiring, `enterStatsView()`/`exitStatsView()` methods
- Story 16.3 for Stats View structure, `viewModeStats`, key handling in model.go Update
- Phase 2 PRD FR-P2-17 requirement
- project-context.md for testing and architecture patterns
- Existing `metricsReaderInterface` at model.go:40-43
- Existing `statsViewScroll` at model.go:183

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - No issues encountered during implementation.

### Completion Notes List

- All 9 tasks completed successfully
- FullTransition type added to statsview/interfaces.go for cross-layer data transfer
- Duration calculation algorithm handles edge cases: no transitions, single transition, revisited stages
- Bar rendering uses Unicode block characters (█, ░) for visual representation
- Breakdown view is a sub-state of viewModeStats (checked via statsBreakdownProject != nil)
- Keyboard navigation: Enter → breakdown, Esc → back to list, Esc again → dashboard
- All tests pass (unit tests for breakdown calculation, repository, and integration)
- make fmt && make lint passes

### File List

**New Files:**
- `internal/adapters/tui/statsview/breakdown.go` - Time calculation and rendering
- `internal/adapters/tui/statsview/breakdown_test.go` - Unit tests for breakdown

**Modified Files:**
- `internal/adapters/tui/statsview/interfaces.go` - Added FullTransition type, FullTransitionReader interface
- `internal/adapters/persistence/metrics/repository.go` - Added GetFullTransitions method
- `internal/adapters/persistence/metrics/repository_test.go` - Tests for GetFullTransitions
- `internal/adapters/tui/model.go` - Added state fields, interface update, getStageBreakdown method, breakdown key handling
- `internal/adapters/tui/statsview.go` - Added renderStatsBreakdownView, modified renderStatsView for switching
- `internal/adapters/tui/statsview_test.go` - Added breakdown integration tests

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Build and Run
```bash
make build
./bin/vdash
```

### Step 2: Enter Stats View
1. Press `s` to open Stats View
2. **Expected:** See project list with sparklines and "STATS" header
3. **Expected:** "[ESC] Back to Dashboard" hint in top-right

### Step 3: Enter Breakdown View
1. Press `Enter` on a project
2. **Expected:** See "Stage Breakdown:" section with horizontal bars
3. **Expected:** Current stage marked with "→" prefix
4. **Expected:** Duration and percentage shown for each stage
5. **Expected:** "[ESC] Back to Project List" hint changes in header

### Step 4: Navigation Test
1. Press `Esc` from breakdown view
2. **Expected:** Returns to project list (NOT dashboard)
3. Press `Esc` again
4. **Expected:** Returns to dashboard
5. **Expected:** Original project selection preserved

### Step 5: Graceful Degradation
1. For a project with no stage transitions recorded
2. Press `Enter` in Stats View
3. **Expected:** "No stage data available" message (no crash)

### Decision Guide
| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Breakdown shows wrong durations | Do NOT approve, check CalculateFromFullTransitions |
| Esc key behavior incorrect | Do NOT approve, check handleStatsViewKeyMsg |
| Crash on empty metrics | Do NOT approve, check nil handling |
