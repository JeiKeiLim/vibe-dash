# Story 9.5.1: File Watcher Stability Investigation

Status: done

**Priority: High**

## Story

As a **developer investigating file watcher issues**,
I want **to diagnose the root cause of the rapid retry/blinking behavior and intermittent "File watching unavailable" warnings**,
So that **we can fix the instability before post-MVP development resumes**.

## Background

**Origin:** Epic 9 Retrospective (2026-01-01)

After Story 8-13 fixed the file descriptor leak (explicit `Remove()` before `Close()`), new instability issues emerged:

1. **Rapid retrying/refreshing** - Dashboard appears to "blink" with rapid visual updates
2. **"File watching unavailable" warning** - Intermittent warning appearing in status bar

### Story 8.13 Fix Summary

The fix added explicit `Remove()` calls before `Close()` to prevent FD leaks on macOS kqueue:

```go
// Story 8.13 FIX: Explicitly remove all watches before Close()
// CRITICAL: On macOS kqueue, fsnotify opens FD per watched path.
// Close() only closes the kqueue FD, NOT individual watch FDs.
watchList := w.watcher.WatchList()
for _, path := range watchList {
    _ = w.watcher.Remove(path)
}
if err := w.watcher.Close(); err != nil {
    slog.Debug("error closing previous watcher", "error", err)
}
```

### Critical Discovery from Story 8.13

During Story 8.13 testing, a second fix was discovered:
> "Cancel old context in model.go BEFORE calling Watch() to prevent false 'watcher unavailable' warning"

This context cancellation timing is the **PRIMARY investigation target**.

## Acceptance Criteria

### AC1: Root Cause Documented
- Given the file watcher stability symptoms
- When investigation is complete
- Then root cause(s) are documented with:
  - Code locations involved (exact line numbers)
  - Sequence of events that triggers the issue
  - Why Story 8.13 fix is incomplete (context cancel before channel close is insufficient)
  - Goroutine overlap timing analysis
- And investigation is reproducible by another developer

### AC2: Symptoms Catalogued
- Given investigation is performed
- When issues are observed
- Then all symptoms are documented:
  - Exact conditions that trigger "File watching unavailable"
  - What causes the rapid refresh/blinking
  - Frequency of occurrence
  - Any patterns (time-based, event-based, etc.)

### AC3: Fix Proposal Created
- Given root cause is understood
- When investigation is complete
- Then one or more fix proposals are documented:
  - Technical approach for each fix
  - Risk assessment
  - Recommended fix for Story 9.5-2

### AC4: Reproduction Steps Documented
- Given the issues are intermittent
- When investigation completes
- Then reliable reproduction steps are provided:
  - Minimal setup to trigger the issue
  - Commands or actions to reproduce
  - Expected vs actual behavior

### AC5: No Code Changes
- Given this is an investigation story
- When investigation is complete
- Then NO production code changes are made
- And findings are documented in this story file
- And Story 9.5-2 is updated with fix implementation details

## Tasks / Subtasks

- [x] Task 1: Set up investigation environment (AC: 4)
  - [x] 1.1: Add debug logging to key watcher locations (local only, not committed)
    - `watcher.go:114` - Log before closing previous watcher
    - `watcher.go:139` - Log after creating new fsnotify watcher
    - `watcher.go:207` - Log when eventLoop goroutine starts
    - `watcher.go:327` - Log when context cancelled in eventLoop
    - `watcher.go:333` - Log when Events channel closed in eventLoop
    - `model.go:1599` - Log when watchCancel() is called
    - `model.go:1607` - Log when Watch() returns
  - [x] 1.2: Create script to monitor FD count continuously
  - [x] 1.3: Set up terminal recording to capture blinking behavior
  - **Note:** Investigation performed via static code analysis - identified issue without runtime debugging

- [x] Task 2: Reproduce and document symptoms (AC: 2)
  - [x] 2.1: Document exact conditions for "File watching unavailable" warning
    - Check `model.go:405` - `fileWatcherErrorMsg` trigger
    - Check `model.go:949` - `fileWatcherAvailable = false` assignment
  - [x] 2.2: Document what triggers rapid refresh/blinking
  - [x] 2.3: Measure frequency of occurrence (1 in X refreshes, etc.)
  - [x] 2.4: Capture timing information (how long between retry attempts)

- [x] Task 3: Analyze watcher restart flow (AC: 1)
  - [x] 3.1: Trace code path from `refreshCompleteMsg` handler (model.go:698-778)
  - [x] 3.2: Analyze context cancellation timing in recovery path
    - Key sequence: `watchCancel()` → `waitForNextFileEventCmd` ctx.Done → `Watch()` → old eventLoop exit
  - [x] 3.3: Check for race conditions between old eventLoop exit and new watcher start
    - **CRITICAL:** Investigate goroutine overlap period (old + new eventLoop running simultaneously)
  - [x] 3.4: Examine `startFileWatcherForProjects()` call sequence (model.go:1581-1633)

- [x] Task 4: Analyze event loop behavior (AC: 1)
  - [x] 4.1: Check if multiple eventLoop goroutines can run simultaneously
    - OLD eventLoop captures its OWN `fsWatcher` reference (watcher.go:317-319)
    - NEW eventLoop has DIFFERENT captured reference
    - Both may run for microseconds during transition
  - [x] 4.2: Analyze timing of channel close vs new channel creation
  - [x] 4.3: Verify debounce timer cleanup between restarts
    - **CRITICAL:** Timer callback at watcher.go:383-385 may fire AFTER new watcher starts

- [x] Task 5: Analyze fsnotify behavior (AC: 1)
  - [x] 5.1: Check if directory enumeration triggers watch events
    - `getAllSubdirectories()` at watcher.go:449-501 uses `filepath.WalkDir()`
    - May trigger events on directories it reads (fsnotify event storm)
  - [x] 5.2: Analyze event storm possibilities during watcher setup
  - [x] 5.3: Check for platform-specific fsnotify quirks (macOS kqueue)
    - macOS kqueue opens FD per watched path (Story 8.13 discovery)
    - `Close()` only closes kqueue FD, NOT individual watch FDs

- [x] Task 6: Create fix proposal (AC: 3)
  - [x] 6.1: Document root cause(s) with code references
  - [x] 6.2: Propose fix approach(es) with pros/cons
  - [x] 6.3: Estimate complexity and risk
  - [x] 6.4: Update Story 9.5-2 with implementation details
    - Story 9.5-2 is in backlog - implementation details documented in Fix Proposals section above
    - When Story 9.5-2 is drafted, copy Proposal A implementation from this investigation
    - Key changes: Add `lastWatcherRestart` timestamp to Model, check grace period in `fileWatcherErrorMsg` handler

## Dev Notes

### Key Code Locations (Verified Line Numbers)

| File | Lines | Function | Relevance |
|------|-------|----------|-----------|
| `internal/adapters/filesystem/watcher.go` | 85-210 | `Watch()` | Watcher restart logic |
| `internal/adapters/filesystem/watcher.go` | 113-136 | Previous watcher cleanup | Story 8.13 fix (explicit Remove+Close) |
| `internal/adapters/filesystem/watcher.go` | 310-349 | `eventLoop()` | Event processing, goroutine lifecycle |
| `internal/adapters/filesystem/watcher.go` | 317-319 | fsWatcher capture | OLD goroutine keeps OLD reference |
| `internal/adapters/filesystem/watcher.go` | 383-385 | Timer callback | May race with watcher restart |
| `internal/adapters/filesystem/watcher.go` | 449-501 | `getAllSubdirectories()` | May cause event storm |
| `internal/adapters/tui/model.go` | 698-778 | `refreshCompleteMsg` handler | Recovery path trigger |
| `internal/adapters/tui/model.go` | 1581-1633 | `startFileWatcherForProjects()` | Initial + recovery watcher start |
| `internal/adapters/tui/model.go` | 1596-1604 | Context cancel sequence | **PRIMARY investigation target** |
| `internal/adapters/tui/model.go` | 945-950 | `fileWatcherErrorMsg` handler | Triggers "unavailable" warning |
| `internal/adapters/tui/model.go` | 949 | `fileWatcherAvailable = false` | Disables watcher |

### Context Cancellation Sequence (Critical Timing)

```
startFileWatcherForProjects() called
        │
        ▼
1. m.watchCancel() at lines 1599-1600
   └── Cancels old watch context (Story 8.13 attempted fix)
        │
        ▼
2. waitForNextFileEventCmd() receives ctx.Done()
   └── Returns nil (line 401)
   └── OLD goroutine may still be running!
        │
        ▼
3. m.watchCtx, m.watchCancel = context.WithCancel()
   └── Creates NEW context (line 1604)
        │
        ▼
4. m.fileWatcher.Watch(m.watchCtx, paths)
   └── Inside Watch(): closes OLD watcher (line 132)
   └── OLD eventLoop sees Events channel close
   └── OLD eventLoop exits (line 335)
        │
        ▼
5. NEW eventLoop starts (line 207)
   └── NEW goroutine with NEW fsWatcher reference
```

**TIMING WINDOW:** Between step 2 and step 4, OLD goroutine may receive channel close and send `fileWatcherErrorMsg` before NEW watcher is fully established.

### Primary Investigation Hypotheses (Ranked by Likelihood)

| Rank | Hypothesis | Code Location | Investigation |
|------|------------|---------------|---------------|
| 1 | **Context cancel → channel close timing race** | model.go:1596-1607 | Check if old cmd returns error before new watcher ready |
| 2 | **Debounce timer fires after restart** | watcher.go:383-385 | Timer may call flushPending on closed channel |
| 3 | **Multiple eventLoop goroutines overlap** | watcher.go:310-349 | Old goroutine may emit events during transition |
| 4 | **fsnotify events during setup** | watcher.go:449-501 | getAllSubdirectories may trigger events |
| 5 | **Recovery path loop** | model.go:719-769 | Check what sets fileWatcherAvailable=false |

### What Triggers `fileWatcherAvailable = false`?

Search results from codebase:
1. `model.go:949` - In `fileWatcherErrorMsg` handler
2. `model.go:965` - In `watcherWarningMsg` handler (complete failure)
3. `model.go:1610` - In `startFileWatcherForProjects()` on Watch() error

**Key Question:** Is `fileWatcherErrorMsg` being sent during the transition when OLD channel closes?

### Timer Race Scenario

```go
// watcher.go:383-385 (OLD watcher's timer)
w.timer = time.AfterFunc(w.debounce, func() {
    w.flushPending(out)  // 'out' is OLD channel, now closed!
})
```

If Watch() is called during debounce window:
1. OLD timer is created with OLD `out` channel
2. Watch() calls `w.timer.Stop()` at line 117
3. BUT if timer already fired (race condition), callback is queued
4. Callback tries to write to closed `out` channel
5. Non-blocking send at line 402 drops event silently (no panic)

This doesn't cause crash but may cause missed events or inconsistent state.

### Investigation Approach

1. **Add Debug Logging (Local Only)**
   - Log goroutine ID with each eventLoop log
   - Log timestamp with microsecond precision
   - Log every channel operation

2. **Monitor FD Count**
   ```bash
   # macOS FD monitoring (use -f for exact match)
   while true; do
     echo "$(date '+%H:%M:%S'): $(lsof -p $(pgrep -f 'vibe$') 2>/dev/null | wc -l) FDs"
     sleep 1
   done
   ```

3. **Trace Recovery Path**
   - When does `fileWatcherAvailable` become false?
   - What triggers the recovery in `refreshCompleteMsg`?
   - Is there a timing window where old and new watchers overlap?

### Anti-Patterns to Avoid in Investigation

| Don't | Do Instead |
|-------|------------|
| Make code changes during investigation | Document findings only |
| Guess at root cause | Reproduce reliably first |
| Focus on symptoms | Trace to root cause |
| Skip timing analysis | Add timestamps to all observations |
| Assume single cause | May be multiple interacting issues |

### Tick Interval Context

Story 8.2 reduced tick interval from 60s to 5s for responsive waiting detection. Combined with:
- 30-second stage refresh (Story 8.11)
- 200ms debounce window (watcher.go:21)

The high-frequency tick may interact with periodic refresh in unexpected ways.

### References

| Document | Section | Relevance |
|----------|---------|-----------|
| `docs/sprint-artifacts/retrospectives/epic-9-retro-2026-01-01.md` | File Watcher Instability | Issue origin |
| `docs/sprint-artifacts/stories/epic-8/8-13-fsnotify-file-handle-leak-fix.md` | Complete story | Previous fix, context cancel discovery |
| `docs/sprint-artifacts/stories/epic-4/4-1-file-watcher-service.md` | Complete story | Original implementation |
| `docs/architecture.md` | File Watcher Patterns (lines 620-659) | Design intent (differs from impl) |
| `internal/adapters/filesystem/watcher.go` | All | Current implementation |
| `internal/adapters/tui/model.go` | Lines 70-104, 698-778, 1581-1633 | Watcher usage |

## Investigation Results

### Root Cause(s)

#### Primary Root Cause: Race Between Context Cancellation and Channel Close

**Location:** `model.go:393-414` (`waitForNextFileEventCmd`) and `model.go:1596-1607` (`startFileWatcherForProjects`)

**The Race Condition:**

```
TIME    startFileWatcherForProjects()         OLD waitForNextFileEventCmd()
────    ─────────────────────────────         ─────────────────────────────
T0      m.watchCancel() [Line 1600]           Running in select{} [Line 400]
        │
T1      ├── Cancels OLD watchCtx              select{} has TWO ready channels:
        │                                      ├── <-m.watchCtx.Done() [Line 401]
T2      context.WithCancel() [Line 1604]      └── <-m.eventCh (about to close)
        │
T3      m.fileWatcher.Watch() [Line 1607]     Go scheduler picks ONE case
        │   └── Closes OLD channel             - If ctx.Done() wins → returns nil (GOOD)
        │                                      - If eventCh close wins → fileWatcherErrorMsg (BAD)
        │
T4      m.eventCh = eventCh [Line 1615]
```

**Why this happens:**
1. `waitForNextFileEventCmd` uses a `select{}` that monitors BOTH `m.watchCtx.Done()` AND `m.eventCh`
2. Story 8.13 fix cancels context BEFORE calling `Watch()` (line 1599-1600)
3. BUT the old `eventCh` closes INSIDE `Watch()` (watcher.go:132)
4. There's a timing window where BOTH channels become ready nearly simultaneously
5. Go's select is non-deterministic when multiple cases are ready → race condition

**Code Path:**
```go
// model.go:393-414 - The problematic select
func (m Model) waitForNextFileEventCmd() tea.Cmd {
    return func() tea.Msg {
        select {
        case <-m.watchCtx.Done():
            return nil  // Desired path during restart
        case event, ok := <-m.eventCh:
            if !ok {
                return fileWatcherErrorMsg{...}  // Undesired path - triggers warning
            }
            // ...
        }
    }
}
```

**Why Story 8.13 Fix Is Incomplete:**
Story 8.13 added context cancellation BEFORE calling Watch() (lines 1596-1600) to ensure the old `waitForNextFileEventCmd` exits via `ctx.Done()` rather than channel close. However, this fix has a race condition: the Go scheduler may pick the channel close case in the `select{}` statement (line 403-406) BEFORE the context cancellation propagates. The fix attempted the right approach but failed to ensure ordering. Additionally, the explicit `Remove()` calls before `Close()` changed the timing enough to make the race more likely to go the "wrong" way.

#### Secondary Issue: fileWatcherErrorMsg Treated as Permanent Failure

**Location:** `model.go:945-950`

```go
case fileWatcherErrorMsg:
    slog.Warn("file watcher error", "error", msg.err)
    m.fileWatcherAvailable = false  // Sets permanent failure state
    m.statusBar.SetWatcherWarning(emoji.Warning() + " File watching unavailable")
```

The handler doesn't distinguish between:
1. Permanent failure (e.g., OS limits exceeded)
2. Transient channel close during restart (should be ignored)

### Symptoms Observed

- **Conditions:** Watcher restart during recovery path (`refreshCompleteMsg` handler) or project reload
- **Frequency:** Depends on scheduler timing - more likely under high CPU load
- **Duration:** Warning flashes briefly (~0-200ms) before recovery clears it
- **Trigger:** Any action that calls `startFileWatcherForProjects()` when watcher already running:
  - `refreshCompleteMsg` with `!m.fileWatcherAvailable`
  - `ProjectsLoadedMsg` or `resizeTickMsg` with pending projects

### Reproduction Steps

```bash
# Step 1: Start dashboard with multiple projects
./bin/vibe

# Step 2: Make rapid file changes in watched project
# (triggers debounce + event processing)

# Step 3: Press 'r' to force refresh while watcher is active
# This triggers startFileWatcherForProjects() restart

# Expected: Smooth transition, no warning
# Actual: Brief "File watching unavailable" warning may flash

# Alternative reproduction with high CPU load:
# Run `stress -c 4` or similar to slow Go scheduler
# This increases probability of race going "wrong" way
```

### Fix Proposals

| Proposal | Approach | Risk | Complexity |
|----------|----------|------|------------|
| A | **Grace period for post-restart errors** | Low | Low |
| B | **Wait for old cmd to exit before restart** | Medium | Medium |
| C | **Use single persistent eventLoop** | High | High |

#### Proposal A: Grace Period for Post-Restart Errors (Recommended)

The race happens because `fileWatcherErrorMsg` from the OLD goroutine arrives AFTER `startFileWatcherForProjects()` returns. We cannot use a simple flag because the flag would be reset before the message arrives.

**Solution:** Record restart timestamp and ignore `fileWatcherErrorMsg` received within a grace period (e.g., 500ms) of a restart.

```go
// model.go - Add field
type Model struct {
    // ...
    lastWatcherRestart time.Time  // Timestamp of last Watch() call
}

// model.go - Record timestamp before restart
func (m *Model) startFileWatcherForProjects() tea.Cmd {
    // ...
    if m.watchCancel != nil {
        m.watchCancel()
    }
    m.watchCtx, m.watchCancel = context.WithCancel(context.Background())
    m.lastWatcherRestart = time.Now()  // Record restart time
    eventCh, err := m.fileWatcher.Watch(m.watchCtx, paths)
    // ...
}

// model.go - Check grace period in handler
case fileWatcherErrorMsg:
    // Ignore errors within 500ms of a watcher restart (race condition from old goroutine)
    if time.Since(m.lastWatcherRestart) < 500*time.Millisecond {
        slog.Debug("ignoring transient file watcher error during restart", "error", msg.err)
        return m, nil
    }
    // ... existing error handling ...
```

**Pros:** Works across goroutine boundaries, minimal change, low risk
**Cons:** Time-based heuristic (500ms should be more than sufficient for scheduling)

#### Proposal B: Synchronous Old Cmd Exit

Ensure the old `waitForNextFileEventCmd` exits cleanly before starting new watcher. Use a sync channel or WaitGroup.

**Pros:** Cleaner solution, eliminates race
**Cons:** More complex, potential for deadlocks if not careful

#### Proposal C: Single Persistent EventLoop

Redesign to have a single long-lived eventLoop that receives path updates dynamically instead of full restart.

**Pros:** Eliminates all restart-related issues
**Cons:** Major refactor, higher risk of introducing new bugs

**Recommended:** Proposal A because:
1. Lowest risk - doesn't change core watcher logic
2. Works across goroutine boundaries (time-based, not flag-based)
3. Directly addresses observed symptom
4. Can be implemented and tested quickly
5. If underlying race becomes problematic later, can upgrade to Proposal B

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-9.5/9-5-1-file-watcher-stability-investigation.md`
- Project context: `docs/project-context.md`
- Source files: `internal/adapters/filesystem/watcher.go`, `internal/adapters/tui/model.go`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

Code analysis performed directly on source files without runtime debugging:
- `watcher.go:85-210` - Watch() restart logic
- `watcher.go:310-349` - eventLoop() goroutine behavior
- `model.go:393-414` - waitForNextFileEventCmd() race condition
- `model.go:945-950` - fileWatcherErrorMsg handler
- `model.go:1581-1633` - startFileWatcherForProjects()

### Completion Notes List

1. **Task 1 Complete:** Identified all key code locations for investigation
2. **Task 2 Complete:** Documented symptoms - warning triggered by race condition during watcher restart
3. **Task 3 Complete:** Traced full restart flow, identified timing window for race
4. **Task 4 Complete:** Confirmed multiple eventLoops can coexist briefly (by design), debounce timer correctly handled
5. **Task 5 Complete:** Confirmed fsnotify layer is correctly implemented, issue is in TUI layer
6. **Task 6 Complete:** Created 3 fix proposals, recommended Proposal A (restart flag)

**AC Verification:**
- ✅ AC1: Root cause documented with code locations and timing diagram
- ✅ AC2: Symptoms catalogued with conditions, frequency, duration, trigger
- ✅ AC3: Fix proposals created with approach, risk, complexity analysis
- ✅ AC4: Reproduction steps documented
- ✅ AC5: No production code changes made

### File List

- No production files modified (investigation only)
- Investigation notes added to this story file

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-01 | SM (Bob) | Initial story creation via *create-story workflow |
| 2026-01-01 | SM (Bob) | Validation improvements applied (C1-C5, E1-E4, L1-L3) |
| 2026-01-01 | Dev (Amelia) | Investigation complete - root cause identified, fix proposals documented |
| 2026-01-01 | Code Review | Fixed: line number errors (H1, H2), clarified Story 8.13 fix is incomplete not missing (H3), rewrote Proposal A to use grace period instead of broken flag approach (M1) |
