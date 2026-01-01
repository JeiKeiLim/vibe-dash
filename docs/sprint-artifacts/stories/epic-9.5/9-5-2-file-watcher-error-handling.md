# Story 9.5.2: File Watcher Error Handling

Status: done

**Priority: High**

## Story

As a **user of vibe-dash**,
I want **the dashboard to gracefully handle watcher restart race conditions**,
So that **I don't see false "File watching unavailable" warnings during normal operation**.

## Background

**Origin:** Story 9.5-1 Investigation (2026-01-01)

Story 9.5-1 investigated the root cause of the intermittent "File watching unavailable" warning that appears briefly during watcher restarts. The investigation identified a race condition between context cancellation and channel close in the `waitForNextFileEventCmd` select statement.

### Root Cause Summary (From Story 9.5-1)

When `startFileWatcherForProjects()` is called to restart the watcher:

1. `m.watchCancel()` cancels the old context (line 1599-1600)
2. `m.fileWatcher.Watch()` closes the old event channel (inside Watch())
3. The old `waitForNextFileEventCmd` has a `select{}` monitoring BOTH channels
4. Go's select is non-deterministic when multiple cases are ready
5. If the channel close case wins the race, `fileWatcherErrorMsg` is sent
6. This triggers the "File watching unavailable" warning (false positive)

### Recommended Fix (Proposal A from Story 9.5-1)

Add a grace period check: record timestamp of last watcher restart, ignore `fileWatcherErrorMsg` received within 500ms of restart. This works across goroutine boundaries without requiring synchronization.

## Acceptance Criteria

### AC1: Grace Period Field Added
- Given the Model struct
- When a new field `lastWatcherRestart time.Time` is added
- Then it tracks the timestamp of the most recent Watch() call

### AC2: Timestamp Recorded on Restart
- Given `startFileWatcherForProjects()` is called
- When Watch() is about to be called
- Then `m.lastWatcherRestart = time.Now()` is set BEFORE the Watch() call

### AC3: Grace Period Check in Error Handler
- Given `fileWatcherErrorMsg` is received
- When `time.Since(m.lastWatcherRestart) < 500*time.Millisecond`
- Then the error is logged at debug level and ignored (return m, nil)
- And `fileWatcherAvailable` is NOT set to false
- And no warning is displayed in status bar

### AC4: Genuine Errors Still Handled
- Given `fileWatcherErrorMsg` is received
- When `time.Since(m.lastWatcherRestart) >= 500*time.Millisecond`
- Then existing error handling proceeds unchanged
- And `fileWatcherAvailable = false` is set
- And warning is displayed in status bar

### AC5: No Visual Blinking
- Given the dashboard is running with file watching enabled
- When user presses 'r' to refresh or periodic refresh triggers
- Then no "File watching unavailable" warning should flash
- And status bar remains stable

### AC6: Tests Updated
- Given the fix is implemented
- When tests are run
- Then model_test.go includes test cases for:
  - Error within grace period (ignored)
  - Error after grace period (handled)
  - Grace period timing boundary (499ms vs 501ms)

## Tasks / Subtasks

- [x] Task 1: Add lastWatcherRestart field to Model (AC: 1)
  - [x] 1.1: Add `lastWatcherRestart time.Time` field to Model struct after `stageRefreshInterval` (line ~104)
  - [x] 1.2: Document field purpose in concise comment

- [x] Task 2: Record timestamp in startFileWatcherForProjects (AC: 2)
  - [x] 2.1: Add `m.lastWatcherRestart = time.Now()` after context creation, before Watch() call (after line 1604)
  - [x] 2.2: Ensure timestamp is set BEFORE Watch() returns

- [x] Task 3: Implement grace period check in fileWatcherErrorMsg handler (AC: 3, 4)
  - [x] 3.1: Add grace period check at the start of fileWatcherErrorMsg case (line 945)
  - [x] 3.2: If within 500ms, log at Debug level and return (m, nil) - DO NOT call statusBar.SetWatcherWarning()
  - [x] 3.3: If beyond 500ms, proceed with existing Warn-level error handling

- [x] Task 4: Write unit tests (AC: 6)
  - [x] 4.1: Extend existing `TestUpdate_FileWatcherErrorMsg` (line ~1347) or add parallel test
  - [x] 4.2: Test case: error at 100ms → ignored, fileWatcherAvailable stays true
  - [x] 4.3: Test case: error at 600ms → handled, fileWatcherAvailable becomes false
  - [x] 4.4: Test case: boundary at 499ms → ignored
  - [x] 4.5: Test case: boundary at 501ms → handled
  - [x] 4.6: Test case: zero-value lastWatcherRestart (app just started) → handled as genuine error
  - [x] 4.7: Verify status bar NOT updated during grace period (check view contains no "unavailable")

- [x] Task 5: Manual verification (AC: 5)
  - [x] 5.1: Run `./bin/vibe` with multiple projects
  - [x] 5.2: Press 'r' repeatedly to trigger restarts
  - [x] 5.3: Verify no warning flashes
  - [x] 5.4: Wait 30 seconds for periodic refresh, verify no warning

## Dev Notes

### Key Code Locations (Verified)

| File | Lines | Function | Change Required |
|------|-------|----------|-----------------|
| `internal/adapters/tui/model.go` | 25-104 | Model struct | Add `lastWatcherRestart time.Time` after line 103 |
| `internal/adapters/tui/model.go` | 1604 | startFileWatcherForProjects | Add timestamp after context creation |
| `internal/adapters/tui/model.go` | 945-950 | fileWatcherErrorMsg handler | Add grace period check before existing logic |
| `internal/adapters/tui/model_test.go` | ~1347 | TestUpdate_FileWatcherErrorMsg | Extend with grace period tests |

**Note:** `time` package already imported at line 10 of model.go.

### Implementation Details

**Field Addition (after stageRefreshInterval, line ~103):**
```go
	// Story 8.11: Periodic stage re-detection interval (0 = disabled)
	stageRefreshInterval int

	// Story 9.5-2: Grace period for restart race condition
	lastWatcherRestart time.Time
}
```

**Timestamp Recording (startFileWatcherForProjects, after line 1604):**
```go
	m.watchCtx, m.watchCancel = context.WithCancel(context.Background())

	m.lastWatcherRestart = time.Now() // Grace period for restart race

	eventCh, err := m.fileWatcher.Watch(m.watchCtx, paths)
```

**Grace Period Check (fileWatcherErrorMsg handler, line 945):**
```go
	case fileWatcherErrorMsg:
		// Story 9.5-2: Ignore transient errors within 500ms of watcher restart
		// Zero-value check: !m.lastWatcherRestart.IsZero() ensures app-startup errors are handled
		if !m.lastWatcherRestart.IsZero() && time.Since(m.lastWatcherRestart) < 500*time.Millisecond {
			slog.Debug("ignoring transient watcher error", "error", msg.err)
			return m, nil
		}

		// Story 4.6: Handle genuine file watcher error
		slog.Warn("file watcher error", "error", msg.err)
		m.fileWatcherAvailable = false
		m.statusBar.SetWatcherWarning(emoji.Warning() + " File watching unavailable")
		return m, nil
```

**Critical:** The zero-value check (`!m.lastWatcherRestart.IsZero()`) ensures errors that occur BEFORE any Watch() call (e.g., during app startup) are treated as genuine failures, not ignored.

### Why 500ms Grace Period?

Per Story 9.5-1: Conservative margin for slow schedulers under high CPU load. Genuine watcher failures persist beyond 500ms.

### Testing Strategy

Extend existing `TestUpdate_FileWatcherErrorMsg` with table-driven subtests:

```go
func TestFileWatcherErrorMsg_GracePeriod(t *testing.T) {
	tests := []struct {
		name                   string
		lastRestart            time.Duration // negative = in the past
		wantAvailable          bool
		wantStatusBarWarning   bool
	}{
		{"within grace period (100ms)", -100 * time.Millisecond, true, false},
		{"after grace period (600ms)", -600 * time.Millisecond, false, true},
		{"boundary ignored (499ms)", -499 * time.Millisecond, true, false},
		{"boundary handled (501ms)", -501 * time.Millisecond, false, true},
		{"zero value (app startup)", 0, false, true}, // Special: don't set lastWatcherRestart
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			m.fileWatcherAvailable = true
			if tt.lastRestart != 0 {
				m.lastWatcherRestart = time.Now().Add(tt.lastRestart)
			}
			// else: leave as zero value

			msg := fileWatcherErrorMsg{err: fmt.Errorf("channel closed")}
			result, _ := m.Update(msg)
			updated := result.(Model)

			if updated.fileWatcherAvailable != tt.wantAvailable {
				t.Errorf("fileWatcherAvailable = %v, want %v", updated.fileWatcherAvailable, tt.wantAvailable)
			}
			hasWarning := strings.Contains(updated.statusBar.View(), "unavailable")
			if hasWarning != tt.wantStatusBarWarning {
				t.Errorf("status bar warning = %v, want %v", hasWarning, tt.wantStatusBarWarning)
			}
		})
	}
}
```

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Use a flag that gets reset before message arrives | Timestamp + duration check |
| Add synchronization (WaitGroup, channels) | Simple time-based check |
| Forget zero-value check for `lastWatcherRestart` | Always check `!IsZero()` before grace period logic |

### Scope Boundaries

- All changes: `internal/adapters/tui/model.go` and `model_test.go`
- No changes to `watcher.go`, ports, or core packages

### References

| Document | Section | Relevance |
|----------|---------|-----------|
| `docs/sprint-artifacts/stories/epic-9.5/9-5-1-file-watcher-stability-investigation.md` | Root Cause + Fix Proposals | Investigation findings, Proposal A selected |
| `docs/sprint-artifacts/stories/epic-8/8-13-fsnotify-file-handle-leak-fix.md` | Complete story | Previous fix that introduced context cancel timing |
| `docs/project-context.md` | Story Completion | User verification required before done |
| `docs/architecture.md` | File Watcher Patterns | Design intent |

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Basic Check
1. Build the application: `make build`
2. Start with multiple projects: `./bin/vibe`
3. Press 'r' 5-10 times rapidly
4. **Expected:** Status bar remains stable, no warning flashes
5. **Red flag:** Any brief "File watching unavailable" warning

### Step 2: Periodic Refresh Check
1. Leave dashboard running for 60 seconds (2 refresh cycles)
2. Observe status bar during periodic refreshes
3. **Expected:** No warning flashes
4. **Red flag:** Warning appearing at 30-second intervals

### Step 3: Genuine Failure Test (Optional)
1. Remove read permissions from a project directory: `chmod 000 /path/to/project`
2. Press 'r' to refresh
3. **Expected:** Warning appears and STAYS (genuine failure)
4. Restore permissions: `chmod 755 /path/to/project`

### Decision Guide
| Situation | Action |
|-----------|--------|
| No warning flashes during rapid refresh | Pass Step 1 |
| No warning flashes during periodic refresh | Pass Step 2 |
| Warning flashes briefly then disappears | FAIL - grace period not working |
| All checks pass | Mark `done` |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-9.5/9-5-2-file-watcher-error-handling.md`
- Investigation: `docs/sprint-artifacts/stories/epic-9.5/9-5-1-file-watcher-stability-investigation.md`
- Project context: `docs/project-context.md`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- No debug logging required for this implementation

### Completion Notes List

1. **Field Added**: `lastWatcherRestart time.Time` field added to Model struct at line 105-106
2. **Timestamp Recording**: `m.lastWatcherRestart = time.Now()` set in `startFileWatcherForProjects()` before Watch() call at line 1609-1610
3. **Grace Period Check**: Implemented at lines 948-954 in fileWatcherErrorMsg handler:
   - Zero-value check (`!IsZero()`) ensures app-startup errors are treated as genuine failures
   - 500ms threshold with `time.Since()` comparison
   - Debug-level logging for ignored transient errors
4. **Tests Added**: 3 new test functions with 7 total test cases:
   - `TestFileWatcherErrorMsg_GracePeriod` (5 subtests: 100ms, 600ms, 499ms, 501ms, zero-value)
   - `TestFileWatcherErrorMsg_GracePeriod_StatusBarNotUpdated` (AC3 detail verification)
   - `TestStartFileWatcher_SetsLastWatcherRestart` (AC2 verification)
5. **Pre-existing Issue**: Golden file test `TestAnchor_Golden_ResizeWideToNarrow` fails (tracked in Story 9.5-4)

### File List

| File | Change Type | Lines Modified |
|------|-------------|----------------|
| `internal/adapters/tui/model.go` | Modified | +9 lines (field, timestamp, grace period check) |
| `internal/adapters/tui/model_test.go` | Modified | +140 lines (3 test functions, fmt import) |

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-01 | SM (Bob) | Initial story creation via *create-story workflow |
| 2026-01-01 | SM (Bob) | Validation improvements: C1-C3 (critical fixes), E1-E4 (enhancements), L1-L3 (LLM optimizations) |
| 2026-01-01 | Dev Agent (Amelia) | Implementation complete: AC1-AC4, AC6 verified. Status: review (awaiting manual verification AC5) |
| 2026-01-01 | Code Review | L1 fix applied: Debug log now includes elapsed_ms for debugging. Manual verification AC5 completed. Status: done |
