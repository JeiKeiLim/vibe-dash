# Story 7.7: Graceful Shutdown

Status: done

## Story

As a **user**,
I want **the application to shut down cleanly on Ctrl+C or q**,
So that **database writes complete and no data is lost**.

## Acceptance Criteria

1. **AC1: Shutdown Message Displayed (CLI Context)**
   - Given the TUI is running
   - When user presses Ctrl+C or 'q'
   - Then application exits gracefully within 5 seconds
   - And debug log shows "shutdown signal received"

2. **AC2: Signal Handling Respects Timeout**
   - Given the application receives SIGINT or SIGTERM
   - When shutdown is initiated
   - Then cleanup completes within 5-second timeout
   - And application forcefully exits if timeout exceeded
   - And log message indicates forced exit if applicable

3. **AC3: Database Connections Closed Cleanly**
   - Given active database operations may be in progress
   - When shutdown is initiated
   - Then in-flight operations complete (within timeout)
   - And all SQLite connections are properly closed
   - And no WAL/SHM files are left in inconsistent state

4. **AC4: File Watcher Cleaned Up**
   - Given file watcher is actively monitoring paths
   - When shutdown is initiated
   - Then file watcher is closed before exit
   - And pending debounced events are flushed
   - And debug log confirms watcher cleanup

5. **AC5: Context Propagation Works End-to-End**
   - Given operations use context.Context for cancellation
   - When shutdown signal is received
   - Then context cancellation propagates to all running operations
   - And operations check context.Done() and exit promptly

6. **AC6: Exit Code is Zero on Clean Shutdown**
   - Given shutdown completes within timeout
   - When user initiated shutdown (q, Ctrl+C, SIGTERM)
   - Then exit code is 0
   - When forced exit after timeout
   - Then exit code is 1

7. **AC7: CLI Commands Respect Cancellation**
   - Given a CLI command (add, remove, list, reset) is running
   - When Ctrl+C is pressed
   - Then command exits promptly
   - And partial writes are avoided where possible

8. **AC8: Rapid Double Signal Handling**
   - Given a shutdown is already in progress
   - When a second SIGINT/SIGTERM is received
   - Then application exits immediately with code 1
   - And log shows "force exit on repeated signal"

## Out of Scope (Deferred)

- **Persistent undo log**: Transaction journaling for recovery - Post-MVP
- **Connection pooling shutdown**: No pooling in current SQLite implementation
- **Graceful HTTP server shutdown**: No HTTP server in MVP

## Epic 7 Context

Story 7.7 is the final story in Epic 7 (Error Handling & Polish) which focuses on graceful error handling, helpful feedback, and final polish. Previous stories established patterns:
- **Story 7.1**: File watcher error recovery with graceful degradation
- **Story 7.2**: Configuration error handling with graceful degradation
- **Story 7.3**: Database recovery with auto-recovery and CLI reset command
- **Story 7.4**: Progress indicators for long operations
- **Story 7.5**: Verbose and debug logging (slog infrastructure)
- **Story 7.6**: Feedback messages polish with consistent prefixes

This story ensures the application handles shutdown as gracefully as it handles errors.

## Critical Implementation Notes

### Key Implementation Gaps to Address

**Gap 1: No Shutdown Timeout in main.go**
Current signal handler (main.go:44-48) calls `cancel()` but has no timeout enforcement. If cleanup hangs, the app hangs forever.

**Gap 2: coordinator.Close() Never Called**
The coordinator's Close() method exists (coordinator.go:144-155) but is never invoked. Note: Close() checks ctx.Done() first - MUST use fresh context for cleanup, not cancelled ctx.

**Gap 3: No "Done" Signal Mechanism**
No way for run() to signal completion back to main goroutine for clean exit.

### Implementation Pattern

Replace current main.go signal handling with timeout-aware pattern:

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    done := make(chan struct{})
    forceExit := make(chan struct{})

    // Setup signal handling
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigCh
        slog.Info("shutdown signal received")
        cancel()

        // Start timeout countdown
        select {
        case <-time.After(5 * time.Second):
            slog.Warn("shutdown timeout exceeded, forcing exit")
            os.Exit(1)
        case <-done:
            // Clean exit - run() completed
        case <-sigCh:
            // Second signal - force exit immediately
            slog.Warn("force exit on repeated signal")
            os.Exit(1)
        }
    }()

    // Run application
    if err := run(ctx); err != nil {
        if !cli.IsSilentError(err) {
            slog.Error("application error", "error", err)
        }
        close(done)
        os.Exit(cli.MapErrorToExitCode(err))
    }
    close(done)
}
```

### Coordinator Cleanup Pattern

**CRITICAL**: coordinator.Close() checks ctx.Done() and returns early. MUST use fresh context:

```go
func run(ctx context.Context) error {
    // ... existing setup ...
    coordinator := persistence.NewRepositoryCoordinator(loader, dirMgr, basePath)

    // Cleanup with FRESH context (not cancelled ctx)
    defer func() {
        cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cleanupCancel()
        if err := coordinator.Close(cleanupCtx); err != nil {
            slog.Error("coordinator cleanup failed", "error", err)
        }
        slog.Debug("coordinator closed")
    }()

    // ... rest of run() ...
}
```

### Files to Modify

| File | Change |
|------|--------|
| `cmd/vibe/main.go` | Add done channel, timeout logic, rapid signal handling, coordinator cleanup |

### TUI Already Handles Cleanup

The TUI (model.go:988-1005) already cleans up file watcher on KeyQuit/KeyForceQuit:
```go
case KeyQuit, KeyForceQuit:
    if m.watchCancel != nil {
        m.watchCancel()
    }
    if m.fileWatcher != nil {
        m.fileWatcher.Close()
    }
    return m, tea.Quit
```

No TUI changes needed - cleanup already works via context cancellation.

### Testing Strategy

1. **Unit tests**: Test timeout behavior with mock long-running operations
2. **Integration tests**: Use `os/exec` to spawn subprocess, send signal, verify exit
3. **Manual verification**:
   - `vibe` then 'q' - verify clean exit, code 0
   - `vibe` then Ctrl+C - verify clean exit, code 0
   - `kill -SIGTERM $(pgrep vibe)` - verify clean exit
   - `vibe --debug` then 'q' - verify cleanup logs appear

### Edge Cases

1. **Rapid double Ctrl+C**: Second signal exits immediately (code 1)
2. **Already shutting down**: Ignore duplicate single signals, force on second
3. **No file watcher**: Nil check already handled in TUI

## Tasks / Subtasks

- [x] Task 1: Add shutdown timeout to main.go (AC: 2, 6, 8)
  - [x] 1.1: Add `done` channel after context creation
  - [x] 1.2: Modify signal goroutine with timeout select (5 seconds)
  - [x] 1.3: Add second signal handler for rapid exit
  - [x] 1.4: Close `done` channel when run() completes

- [x] Task 2: Add coordinator cleanup with fresh context (AC: 3)
  - [x] 2.1: Add defer with context.WithTimeout(context.Background(), 5s)
  - [x] 2.2: Call coordinator.Close(cleanupCtx)
  - [x] 2.3: Add slog.Debug for successful close

- [x] Task 3: Verify context propagation (AC: 5, 7)
  - [x] 3.1: Grep for operations NOT checking ctx.Done() in coordinator - VERIFIED: 15+ methods check ctx.Done()
  - [x] 3.2: Verify CLI commands pass context properly (confirmed - add.go:110 uses cmd.Context())
  - [x] 3.3: Document that existing code already respects context

- [x] Task 4: Add debug logging for cleanup (AC: 4)
  - [x] 4.1: Add slog.Debug in main.go when coordinator closes
  - [x] 4.2: Added slog.Debug to watcher.Close() for "file watcher closed"

- [x] Task 5: Add unit tests (AC: all)
  - [x] 5.1: Created shutdown_test.go with timeout value tests
  - [x] 5.2: Created shutdown_integration_test.go with SIGINT/SIGTERM tests
  - [x] 5.3: Integration tests verify exit code 0 on clean shutdown

- [x] Task 6: Manual verification (AC: all)
  - [x] 6.1: Build with `make build` - SUCCESS
  - [x] 6.2: Test version/list commands exit cleanly - VERIFIED
  - [x] 6.3: Debug logging verified: `vibe --debug list` shows "file watcher closed" and "coordinator closed"
  - [x] 6.4: Integration tests pass for SIGINT/SIGTERM handling

## Dev Notes

### Shutdown Flow Diagram

```
User presses q/Ctrl+C OR SIGINT/SIGTERM received
                |
                v
        main.go: cancel()
        Start 5s timeout
                |
   +------------+------------+
   |            |            |
   v            v            v
TUI exits   CLI exits   Second signal
(tea.Quit)  (ctx.Done)  (force exit)
   |            |            |
   v            v            v
fileWatcher  operations   os.Exit(1)
.Close()     return         |
   |            |            |
   +-----+------+            |
         |                   |
         v                   |
    run() returns            |
    defer: coordinator.Close(freshCtx)
         |                   |
         v                   |
    close(done)              |
         |                   |
         v                   |
    Signal handler exits     |
    os.Exit(0)               |
                             |
    <-- 5s timeout -->   os.Exit(1)
```

### Timeout Value Justification

5 seconds chosen per Architecture and Epics specification:
- SQLite WAL checkpoint: typically <100ms
- File watcher close: ~10ms
- Debounce flush: max 200ms
- User expectation: "instant" is <1s, "acceptable" is <5s
- Total worst case: <500ms, leaving 4.5s margin

### Exit Code Convention

| Scenario | Exit Code |
|----------|-----------|
| Clean shutdown (q, Ctrl+C, SIGTERM) | 0 |
| Forced exit (timeout or double signal) | 1 |
| CLI error (project not found, etc.) | Per cli/exit_codes.go |

## Dependencies

- Story 7.5 completed (slog infrastructure for debug logging)
- Story 4.6 completed (file watcher with Close() method)
- Story 7.3 completed (coordinator Close() method exists)

## References

- [Source: cmd/vibe/main.go:37-58 - Current signal handling]
- [Source: internal/adapters/persistence/coordinator.go:141-155 - Close() method]
- [Source: internal/adapters/filesystem/watcher.go:144-170 - Close() method]
- [Source: internal/adapters/tui/model.go:988-1005 - KeyQuit handling]
- [Source: docs/architecture.md:660-726 - Graceful Shutdown Pattern]
- [Source: docs/epics.md:2823-2845 - Story 7.7 specification]

## Dev Agent Record

**Implemented:** 2025-12-26 by Amelia (Dev Agent)

### Implementation Summary

1. **main.go refactored** for timeout-aware graceful shutdown:
   - Added `shutdownTimeout` constant (5 seconds per architecture spec)
   - Added `done` channel for signaling clean run() completion
   - Signal goroutine now uses select with timeout, done, and second-signal cases
   - Coordinator cleanup with fresh context (not cancelled ctx)

2. **watcher.go enhanced** with debug logging on Close()

3. **Context propagation verified**:
   - coordinator.go: All 15+ methods check ctx.Done()
   - CLI commands use cmd.Context()

### Tests Created

- `cmd/vibe/shutdown_test.go` - Unit tests for timeout constant
- `cmd/vibe/shutdown_integration_test.go` - Integration tests for signal handling

### Verification

- All unit tests pass: `go test ./...`
- All integration tests pass: `go test -tags=integration ./cmd/vibe/...`
- Lint passes: `make lint`
- Debug logging verified: `vibe --debug list` shows cleanup messages

## File List

| File | Change |
|------|--------|
| `cmd/vibe/main.go` | Added shutdownTimeout, done channel, timeout select, coordinator cleanup |
| `internal/adapters/filesystem/watcher.go` | Added slog.Debug on Close() |
| `cmd/vibe/shutdown_test.go` | NEW: Unit tests for shutdown timeout |
| `cmd/vibe/shutdown_integration_test.go` | NEW: Integration tests for signal handling |
| `docs/sprint-artifacts/stories/epic-7/7-7-graceful-shutdown.md` | Updated status to review |

## User Testing Guide

**Time needed:** 2 minutes

### Step 1: Verify Debug Logging

```bash
./bin/vibe --debug list 2>&1 | grep -E "(watcher|coordinator) closed"
```

**Expected output:**
- "file watcher closed" appears
- "coordinator closed" appears

### Step 2: Verify Clean Exit

```bash
./bin/vibe --version
echo "Exit code: $?"
```

**Expected:** Exit code 0

### Decision Guide

| Situation | Action |
|-----------|--------|
| Debug logs show cleanup messages | Mark `done` |
| Exit code is 0 for normal commands | Mark `done` |
| Missing cleanup logs | Do NOT approve, check main.go defer |
| Non-zero exit code | Do NOT approve, check for errors |

## Change Log

- 2025-12-26: Code review fixes applied (Amelia/Dev Agent)
  - Fixed H1: Race condition - close(done) now happens before os.Exit
  - Fixed H2: Integration tests now assert on exit codes instead of just logging
  - Fixed H3: Added TestIntegration_RapidDoubleSignal for AC8 coverage
  - Fixed H4: Changed shutdown log from Debug to Info level for user visibility
  - Fixed L1: Corrected import grouping for time package
  - Fixed M3: Added behavioral unit tests (TestDoneChannel_Behavior, TestDoneChannel_MultipleReaders)
  - All tests pass, lint passes
- 2025-12-26: Story 7.7 implemented (Amelia/Dev Agent)
  - Added timeout-aware shutdown to main.go
  - Added coordinator cleanup with fresh context
  - Added debug logging to file watcher Close()
  - Created unit and integration tests
  - All tests pass, lint passes
- 2025-12-26: Story validation and improvements applied (Bob/Claude Opus 4.5)
  - Fixed timeout value: 3s -> 5s per Architecture/Epics specification
  - Added AC8 for rapid double signal handling
  - Fixed coordinator cleanup pattern: must use fresh context, not cancelled ctx
  - Removed non-rendering TUI shutdown message from scope
  - Added complete main.go implementation pattern
  - Consolidated implementation sections for clarity
  - Added shutdown flow diagram
  - Added reference to architecture graceful shutdown section
- 2025-12-26: Story 7.7 created - Graceful Shutdown (Bob/Claude Opus 4.5)
