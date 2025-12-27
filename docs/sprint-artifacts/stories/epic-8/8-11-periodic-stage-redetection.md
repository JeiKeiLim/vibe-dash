# Story 8.11: Periodic Stage Re-Detection

Status: review

## Story

As a **user monitoring project progress**,
I want **stage detection to refresh automatically**,
So that **I see updated epic/story status without pressing [r] manually**.

## Problem Statement

Current auto-refresh behavior gap:
- **Waiting status:** Updates via file watcher events (event-driven, no polling) - works great
- **Stage detection:** Only runs on manual [r] refresh - users miss stage changes

Stage info (Epic X, Story Y) requires parsing workflow files, so it runs separately from the 5-second tick that updates waiting counts. Users monitoring 24/7 must manually press [r] to see stage changes.

## Design Decision

**Option B: Configurable periodic timer (default 30s)**

Rationale:
- Balances freshness vs I/O efficiency
- User can adjust based on preference
- 30s default is unnoticeable delay in practice
- Method-agnostic (works for BMAD, Speckit, future methods)
- No hardcoded file patterns per detection method
- Consistent with existing tick pattern (Story 8.2 uses 5s for waiting, this uses 30s for stage)

## Acceptance Criteria

```gherkin
AC1: Given default config (no setting)
     When 30 seconds pass since last detection
     Then stage detection runs automatically for all projects

AC2: Given config `stage_refresh_interval: 60s`
     When 60 seconds pass
     Then stage detection runs automatically

AC3: Given config `stage_refresh_interval: 0`
     When any time passes
     Then stage detection only runs on manual [r] refresh

AC4: Given periodic detection completes
     When new stage info differs from current
     Then project list and detail panel update immediately

AC5: Given user presses [r] manually
     When refresh completes
     Then periodic timer resets (avoids redundant detection)

AC6: Given stage detection is running
     When multiple projects exist
     Then all projects are re-detected in single batch

AC7: Given all changes are made
     When `make test && make lint` runs
     Then all tests pass and no lint errors
```

## Tasks / Subtasks

- [x] Task 1: Add `stage_refresh_interval` config option (AC: 1, 2, 3)
  - [x] 1.1: In `internal/core/ports/config.go`, add field AFTER `MaxContentWidth` (around line 44):
    ```go
    // StageRefreshIntervalSeconds is the interval for periodic stage re-detection (Story 8.11)
    // Default: 30. Set to 0 to disable periodic stage detection.
    StageRefreshIntervalSeconds int
    ```
  - [x] 1.2: In `internal/core/ports/config.go:NewConfig()`, set default value `30`:
    ```go
    StageRefreshIntervalSeconds: 30, // Story 8.11: default 30s for stage re-detection
    ```
  - [x] 1.3: In `internal/core/ports/config.go:Validate()`, add validation AFTER MaxContentWidth check:
    ```go
    // Validate stage_refresh_interval (Story 8.11)
    if c.StageRefreshIntervalSeconds < 0 {
        return fmt.Errorf("%w: stage_refresh_interval must be >= 0, got %d", domain.ErrConfigInvalid, c.StageRefreshIntervalSeconds)
    }
    ```
  - [x] 1.4: In `internal/core/ports/config_test.go`, add tests:
    - Test `StageRefreshIntervalSeconds = 0` is VALID (disabled mode)
    - Test `StageRefreshIntervalSeconds = 30` is VALID (default)
    - Test `StageRefreshIntervalSeconds = -1` triggers validation error
  - [x] 1.5: In `internal/config/loader.go:Save()`, add AFTER max_content_width:
    ```go
    l.v.Set("settings.stage_refresh_interval", config.StageRefreshIntervalSeconds) // Story 8.11
    ```
  - [x] 1.6: In `internal/config/loader.go:mapViperToConfig()`, add AFTER max_content_width binding:
    ```go
    // Story 8.11: Read stage_refresh_interval setting
    if l.v.IsSet("settings.stage_refresh_interval") {
        cfg.StageRefreshIntervalSeconds = l.v.GetInt("settings.stage_refresh_interval")
    }
    ```
  - [x] 1.7: In `internal/config/loader.go:fixInvalidValues()`, add AFTER max_content_width fix:
    ```go
    // Fix invalid stage_refresh_interval (Story 8.11)
    if cfg.StageRefreshIntervalSeconds < 0 {
        slog.Warn("invalid stage_refresh_interval, using default",
            "path", l.configPath,
            "invalid_value", cfg.StageRefreshIntervalSeconds,
            "default_value", defaults.StageRefreshIntervalSeconds)
        cfg.StageRefreshIntervalSeconds = defaults.StageRefreshIntervalSeconds
    }
    ```
  - [x] 1.8: In `internal/config/loader.go:writeDefaultConfig()`, add commented example:
    ```yaml
    # stage_refresh_interval: 30  # seconds, 0 = disabled (default: 30)
    ```
  - [x] 1.9: In `internal/config/loader_test.go`, add test for loading/saving StageRefreshIntervalSeconds

- [x] Task 2: Create `stageRefreshTickMsg` and timer (AC: 1, 2, 3, 6)
  - [x] 2.1: In `internal/adapters/tui/model.go`, add `stageRefreshTickMsg` type definition AFTER `tickMsg` (around line 191):
    ```go
    // stageRefreshTickMsg triggers periodic stage re-detection (Story 8.11).
    type stageRefreshTickMsg time.Time
    ```
  - [x] 2.2: In `internal/adapters/tui/model.go:Model` struct, add field AFTER `maxContentWidth` (around line 101):
    ```go
    // Story 8.11: Periodic stage re-detection interval (0 = disabled)
    stageRefreshInterval int
    ```
  - [x] 2.3: In `internal/adapters/tui/model.go:SetConfig()` method (around line 280-286), add wiring:
    ```go
    m.stageRefreshInterval = cfg.StageRefreshIntervalSeconds // Story 8.11
    ```
  - [x] 2.4: In `internal/adapters/tui/model.go`, create method AFTER `tickCmd()` (around line 375):
    ```go
    // stageRefreshTickCmd returns command for periodic stage detection (Story 8.11).
    func (m Model) stageRefreshTickCmd() tea.Cmd {
        if m.stageRefreshInterval == 0 {
            return nil
        }
        return tea.Tick(time.Duration(m.stageRefreshInterval)*time.Second, func(t time.Time) tea.Msg {
            return stageRefreshTickMsg(t)
        })
    }
    ```
  - [x] 2.5: In `internal/adapters/tui/model.go:ProjectsLoadedMsg` handler (around line 617-646), start timer:
    ```go
    // Story 8.11: Start periodic stage refresh timer
    if m.stageRefreshInterval > 0 {
        return m, tea.Batch(m.stageRefreshTickCmd(), /* existing commands if any */)
    }
    ```
    **NOTE:** Timer MUST start in ProjectsLoadedMsg (not Init) because detection requires projects to be loaded first.

- [x] Task 3: Handle `stageRefreshTickMsg` to trigger detection (AC: 4, 6)
  - [x] 3.1: In `internal/adapters/tui/model.go:Update()` switch statement, add case for `stageRefreshTickMsg`:
    ```go
    case stageRefreshTickMsg:
        // Story 8.11: Periodic stage re-detection
        // Skip if disabled, already refreshing, or no projects
        if m.stageRefreshInterval == 0 || m.isRefreshing || len(m.projects) == 0 {
            return m, m.stageRefreshTickCmd()
        }
        // MANDATORY: Reuse existing startRefresh() - provides user feedback
        return m.startRefresh()
    ```
    **CRITICAL:** Reuse `startRefresh()` (which calls `refreshProjectsCmd`). Do NOT create a separate method. The `isRefreshing` flag provides user feedback which is acceptable.
  - [x] 3.2: Verify `refreshCompleteMsg` handler (around line 648-666) already:
    - Updates project list and detail panel via reload
    - Clears isRefreshing flag
    - Shows completion message in status bar
  - [x] 3.3: In the `stageRefreshTickMsg` handler, ensure timer reschedule:
    The `startRefresh()` returns `(tea.Model, tea.Cmd)` - the returned command is `refreshProjectsCmd()`. The timer reschedule happens in `refreshCompleteMsg` handler (see Task 4).

- [x] Task 4: Reset timer on refresh completion (AC: 5)
  - [x] 4.1: In `refreshCompleteMsg` handler (around line 648-720), add timer restart:
    ```go
    // Story 8.11 AC5: Reset stage refresh timer after ANY refresh completes
    // This prevents redundant detection right after manual [r] or periodic refresh
    cmds = append(cmds, m.stageRefreshTickCmd())
    ```
    **WHERE:** Add this BEFORE the final `return m, tea.Batch(cmds...)` in the handler.
  - [x] 4.2: No special tracking needed for manual vs periodic refresh - both go through `startRefresh()` and `refreshCompleteMsg`, so both reset the timer.
  - [x] 4.3: Verify the flow:
    - User presses [r] → `startRefresh()` → sets `isRefreshing=true`
    - `refreshProjectsCmd()` runs detection
    - `refreshCompleteMsg` received → clears `isRefreshing`, restarts timer
    - If periodic tick fires while `isRefreshing=true`, it's skipped (Task 3.1 guard)

- [x] Task 5: Add tests (AC: 7)
  - [x] 5.1: Added tests to `internal/adapters/tui/model_tick_test.go` (consolidated with existing tick tests):
    - Test stageRefreshTickCmd returns nil when interval is 0
    - Test stageRefreshTickCmd returns command when interval > 0
    - Test stageRefreshTickMsg triggers detection
    - Test timer resets after manual refresh
    - Test detection updates project stage info
  - [x] 5.2: In `internal/core/ports/config_test.go`, add tests for StageRefreshIntervalSeconds validation
  - [x] 5.3: In `internal/config/loader_test.go`, add tests for config loading/saving
  - [x] 5.4: Run `make test` - all tests pass
  - [x] 5.5: Run `make lint` - no warnings

## Dev Notes

### Two Timers Coexist (IMPORTANT)

**This story adds a SECOND timer alongside existing tickCmd:**
- `tickCmd()` (5s) - Recalculates waiting status counts (Story 8.2)
- `stageRefreshTickCmd()` (30s default) - Runs full stage re-detection (this story)

Both timers run independently. The 5s tick handles UI responsiveness for waiting counts. The 30s tick handles heavier I/O-bound stage detection.

### Architecture Compliance

**Hexagonal Architecture Requirements:**
- Config change in `internal/core/ports/config.go` (domain layer - zero external deps)
- Viper binding in `internal/config/loader.go` (adapter layer)
- TUI uses config via injected Config struct (not global)

**Config Pattern (same as Story 8.10):**
See Story 8.10 Dev Notes "Pattern to Follow" for the exact 7-step config addition pattern.

### Existing Tick Pattern Reference

**tickCmd() pattern (model.go:368-374):**
```go
func tickCmd() tea.Cmd {
    return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

Follow this exact pattern for `stageRefreshTickCmd()` but:
1. Use method receiver `(m Model)` to access config value
2. Return `nil` when `m.stageRefreshInterval == 0` (disabled mode)

### Detection Reuse Strategy

**MANDATORY: Reuse startRefresh()/refreshProjectsCmd()**

Do NOT create a separate `stageDetectionCmd()`. The existing `refreshProjectsCmd()` at model.go:410-453:
- Calls `m.detectionService.Detect()` for each project
- Updates project fields (DetectedMethod, CurrentStage, Confidence, DetectionReasoning, UpdatedAt)
- Saves to repository
- Sets `isRefreshing = true` (shows progress in status bar - this is ACCEPTABLE)

The `isRefreshing` flag provides user feedback that something is happening, which is good UX.

### Timer Reset Logic

**Flow for both manual [r] and periodic refresh:**
1. `startRefresh()` called → sets `isRefreshing = true`
2. `refreshProjectsCmd()` runs detection for all projects
3. `refreshCompleteMsg` received → clears `isRefreshing`, restarts timer

**Timer restart location:** In `refreshCompleteMsg` handler, add `m.stageRefreshTickCmd()` to the batch of returned commands.

**Guard against concurrent execution:** The `stageRefreshTickMsg` handler checks `m.isRefreshing` and skips if true.

### Previous Story Learnings

**From Story 8.10:** Config additions require updates to BOTH `ports/config.go` AND `config/loader.go`. Follow the exact pattern.

**From Story 8.9:** Zero vs nil semantics matter. Here, `0 = disabled` (integer semantics, not pointer/nil).

**From Story 8.2:** Stage detection is I/O-heavy (file parsing). 30s default balances freshness vs efficiency.

**From Story 8.4:** Race conditions occur when ProjectsLoadedMsg arrives before WindowSizeMsg. Timer MUST start in ProjectsLoadedMsg handler (not Init) to ensure projects exist.

### Config YAML Example

```yaml
settings:
  # ... existing settings ...
  stage_refresh_interval: 30  # seconds, 0 = disabled (default: 30)
```

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Hardcode 30 seconds | Use `m.stageRefreshInterval` from config |
| Create new `stageDetectionCmd()` | Reuse existing `startRefresh()` |
| Forget to reschedule tick | Return `stageRefreshTickCmd()` in `refreshCompleteMsg` |
| Fire detection while refreshing | Guard with `m.isRefreshing` check |
| Start timer in Init() | Start in ProjectsLoadedMsg (projects must exist) |
| Use `tea.Every` | Use `tea.Tick` (consistent with tickCmd pattern) |
| Forget to reset on manual refresh | Timer resets via shared `refreshCompleteMsg` path |
| Create method without receiver | `stageRefreshTickCmd()` needs `(m Model)` receiver for config access |

### Line Number Note

Line numbers in this story are approximate as of 2025-12-27. Use function names and code comments as authoritative locators. Key functions:
- `tickCmd()` - reference pattern for tick commands
- `startRefresh()` - reuse for stage detection
- `refreshCompleteMsg` case - add timer restart here
- `SetConfig()` - wire `stageRefreshInterval` here

### Testing Strategy

**Unit Tests:**
1. Test config validation: StageRefreshIntervalSeconds >= 0
2. Test stageRefreshTickCmd() returns nil when interval is 0
3. Test stageRefreshTickCmd() returns command when interval > 0
4. Test stageRefreshTickMsg handler triggers detection
5. Test timer resets after refreshCompleteMsg

**Integration Tests (Manual):**
```bash
# Test default (30s)
./bin/vibe
# Wait 35 seconds, verify stage info updates

# Test disabled (0)
# Set stage_refresh_interval: 0 in config
./bin/vibe
# Wait 60 seconds, verify stage info does NOT update
# Press 'r', verify stage info updates

# Test custom (10s for faster testing)
# Set stage_refresh_interval: 10 in config
./bin/vibe
# Wait 15 seconds, verify stage info updates
```

### Project Structure Notes

- Changes are entirely in TUI adapter layer (`internal/adapters/tui/`) and config (`internal/config/`, `internal/core/ports/`)
- No database schema changes
- No CLI command changes
- Backward compatible (default 30s adds new behavior without breaking existing)

### References

**Key Files:**
- `internal/adapters/tui/model.go` - Main TUI model with tick patterns, SetConfig(), ProjectsLoadedMsg handler, refreshCompleteMsg handler
- `internal/core/ports/config.go` - Config struct, NewConfig(), Validate()
- `internal/config/loader.go` - Viper bindings: Save(), mapViperToConfig(), fixInvalidValues(), writeDefaultConfig()

**Key Functions (use as locators):**
- `tickCmd()` - Pattern to follow for timer commands
- `startRefresh()` - Reuse for stage detection
- `refreshProjectsCmd()` - Actual detection logic
- `SetConfig()` - Wire stageRefreshInterval here

**Related Stories:**
- Story 8.10 - Config pattern to follow exactly
- Story 8.2 - 5-second tickCmd() implementation
- Story 8.4 - Race condition fix (pendingProjects pattern)
- docs/project-context.md - Project patterns and anti-patterns

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Default Behavior (30s interval)

```bash
make build && ./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Stage info in project list | Updates automatically after ~30s | Only updates on [r] press |
| Manual [r] press | Triggers immediate refresh | No refresh or error |

**How to verify stage change:**
1. In separate terminal, modify a workflow file (e.g., change story status in sprint-status.yaml)
2. Wait 35 seconds (30s interval + 5s buffer)
3. Dashboard should show updated stage info

### Step 2: Disabled (0)

```bash
# Add to ~/.vibe-dash/config.yaml:
# settings:
#   stage_refresh_interval: 0
./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Wait 60 seconds | Stage info does NOT update | Stage info updates |
| Press [r] | Stage info updates immediately | No update |

### Step 3: Custom Interval (10s for testing)

```bash
# Add to ~/.vibe-dash/config.yaml:
# settings:
#   stage_refresh_interval: 10
./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Wait 15 seconds | Stage info updates | Only updates on [r] |

### Step 4: Timer Reset on Manual Refresh

```bash
# Use 10s interval for easier testing
./bin/vibe
# Wait 5 seconds, then press [r]
# Timer should reset - next auto-refresh in 10s from [r] press, not 5s
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Manual [r] mid-cycle | Resets timer, next auto-refresh in full interval | Immediate re-detection after [r] |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Auto-refresh not working | Check stageRefreshTickMsg handler and timer start |
| Timer not resetting | Check refreshCompleteMsg handler |
| Config not loading | Check loader.go bindings |
| Tests fail | Check config validation and mock detector setup |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5

### Debug Log References

### Completion Notes List

- All 5 tasks completed
- All 7 acceptance criteria verified via implementation
- Tests added to model_tick_test.go and config_test.go
- `make test && make lint` pass

### File List

- internal/core/ports/config.go - Added StageRefreshIntervalSeconds field, default, validation
- internal/core/ports/config_test.go - Added validation tests for StageRefreshIntervalSeconds
- internal/config/loader.go - Added Viper bindings (Save, mapViperToConfig, fixInvalidValues, writeDefaultConfig)
- internal/config/loader_test.go - Added loading/saving tests for StageRefreshIntervalSeconds (code review fix H2)
- internal/adapters/tui/model.go - Added stageRefreshTickMsg type, stageRefreshInterval field, stageRefreshTickCmd(), handler, timer reset in refreshCompleteMsg
- internal/adapters/tui/model_tick_test.go - Added Story 8.11 tests

## Change Log

- 2025-12-26: Story created via correct-course workflow (user feedback during 8.4 review)
- 2025-12-27: Enriched with comprehensive developer context by SM agent (YOLO mode)
- 2025-12-27: SM validation applied improvements:
  - **C1 FIXED:** Added explicit code snippets for config field placement with comments
  - **C2 FIXED:** Added explicit SetConfig() wiring code in Task 2.3
  - **C3 FIXED:** Clarified timer starts in ProjectsLoadedMsg handler (not Init) with explanation
  - **E1 FIXED:** Added explicit test cases for 0 (valid/disabled), 30 (valid/default), -1 (invalid)
  - **E2 FIXED:** Added YAML key format example in Task 1.8
  - **E3 FIXED:** Changed from "Option A recommended" to "MANDATORY: Reuse startRefresh()"
  - **E4 FIXED:** Added "Two Timers Coexist" section explaining tickCmd() vs stageRefreshTickCmd()
  - **O1 FIXED:** Condensed config pattern reference to point to Story 8.10
  - **O2 FIXED:** Added "Line Number Note" section about approximate line numbers
  - Updated anti-patterns table with 1 new entry (method receiver requirement)
  - Streamlined Dev Notes by removing duplicate information
  - Updated References section with function-based locators
