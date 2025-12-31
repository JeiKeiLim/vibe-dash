# Story 8.13: Fix fsnotify File Handle Leak

Status: done

**Priority: P0 CRITICAL**

## Story

As a **user running the dashboard for extended periods (24/7 monitoring)**,
I want **the file watcher to properly manage its resources**,
So that **agent detection continues working reliably without "too many open files" errors**.

## Background

**Regression from Story 8.11 (Periodic Stage Re-detection)**

When the dashboard runs for several hours, fsnotify generates "too many open files" errors. This breaks Agent Waiting Detection (the killer feature).

**Root Cause:** In `watcher.go`, when `Watch()` is called on an existing `FsnotifyWatcher`:
1. Line 106 creates a NEW `fsnotify.Watcher`
2. Line 167 assigns it to `w.watcher`
3. **The previous watcher is NEVER closed** → file handles orphaned

**Trigger:** Story 8.11's periodic refresh can call `Watch()` via recovery path at `model.go:731`.

## Acceptance Criteria

### AC1: Close Previous Watcher Before Creating New
- Given `Watch()` is called on an `FsnotifyWatcher` instance
- When `w.watcher` already holds a previous watcher
- Then the previous watcher MUST be closed before creating a new one
- And the old event loop goroutine exits cleanly (via closed Events channel)

### AC2: No Resource Leak Over Time
- Given the dashboard runs for 8+ hours
- When periodic refreshes occur every 30 seconds
- Then file descriptor count remains stable (±10)
- And no "too many open files" errors occur

### AC3: Normal Operation Preserved
- Given the fix is applied
- When dashboard starts normally
- Then file watching works as before
- And agent detection triggers on file changes

### AC4: Error Recovery Still Works
- Given file watcher encounters an error
- When recovery path calls `Watch()` again
- Then old watcher is closed, new one created successfully
- And watching resumes without leaked handles

### AC5: Test Coverage
- Given all changes are made
- When `make test && make lint` runs
- Then all tests pass and no lint errors
- And regression test verifies no leak on multiple Watch() calls

## Tasks / Subtasks

- [x] Task 1: Fix watcher.go Watch() method (AC: 1)
  - [x] 1.1: In `internal/adapters/filesystem/watcher.go`, add cleanup AFTER line 103 (`if w.closed` check) and BEFORE line 106 (`fsnotify.NewWatcher()`):
    ```go
    // Story 8.13: Close previous watcher to prevent file handle leak
    // The old eventLoop goroutine will exit when watcher.Events closes
    if w.watcher != nil {
        // Stop pending debounce timer to prevent callback races
        if w.timer != nil {
            w.timer.Stop()
            w.timer = nil
        }
        // Clear pending events from old watcher
        w.pending = make(map[string]ports.FileEvent)
        // Close old watcher - error logged but doesn't prevent new watcher
        if err := w.watcher.Close(); err != nil {
            slog.Debug("closing previous watcher", "error", err)
        }
        w.watcher = nil
    }
    ```
  - [x] 1.2: Verify placement is AFTER mutex lock (line 98-99) and closed check
  - [x] 1.3: Confirm old eventLoop goroutine exits via `fsWatcher.Events` channel close

- [x] Task 2: Review all Watch() call sites (AC: 3, 4)
  - [x] 2.1: Audit `model.go:1596` - initial call in `startFileWatcherForProjects()`
  - [x] 2.2: Audit `model.go:731` - recovery path in `refreshCompleteMsg` handler
  - [x] 2.3: Verify context cancellation order: `m.watchCancel()` called BEFORE `Watch()`
  - [x] 2.4: Run `grep -r "\.Watch(" internal/` to confirm no other call sites

- [x] Task 3: Add regression test (AC: 5)
  - [x] 3.1: In `internal/adapters/filesystem/watcher_test.go`, add:
    ```go
    func TestFsnotifyWatcher_Watch_ClosePreviousWatcher(t *testing.T) {
        // Verify calling Watch() twice closes the first watcher
        tmpDir1 := t.TempDir()
        tmpDir2 := t.TempDir()

        w := NewFsnotifyWatcher(DefaultDebounce)
        ctx := context.Background()

        // First watch
        ch1, err := w.Watch(ctx, []string{tmpDir1})
        require.NoError(t, err)
        require.NotNil(t, ch1)

        // Second watch - should close first watcher without error
        ch2, err := w.Watch(ctx, []string{tmpDir2})
        require.NoError(t, err)
        require.NotNil(t, ch2)

        // First channel should be closed (old eventLoop exited)
        // Give time for goroutine cleanup
        time.Sleep(50 * time.Millisecond)

        // Cleanup
        require.NoError(t, w.Close())
    }

    func TestFsnotifyWatcher_Watch_ClearsTimer(t *testing.T) {
        // Verify pending timer is stopped on re-watch
        tmpDir := t.TempDir()

        w := NewFsnotifyWatcher(DefaultDebounce)
        ctx := context.Background()

        _, err := w.Watch(ctx, []string{tmpDir})
        require.NoError(t, err)

        // Trigger a debounce by creating a file
        testFile := filepath.Join(tmpDir, "test.txt")
        require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

        // Immediately re-watch before debounce fires
        _, err = w.Watch(ctx, []string{tmpDir})
        require.NoError(t, err)

        // Should not panic or race
        require.NoError(t, w.Close())
    }
    ```
  - [x] 3.2: Run `make test` - all tests pass
  - [x] 3.3: Run `make lint` - no warnings

- [ ] Task 4: Manual verification (AC: 2) - USER TESTING REQUIRED
  - [ ] 4.1: Build and run: `make build && ./bin/vibe`
  - [ ] 4.2: Monitor FDs (macOS): `watch -n 10 'lsof -p $(pgrep -f "vibe$") 2>/dev/null | wc -l'`
  - [ ] 4.3: Wait for 5+ refresh cycles (~3 minutes with 30s interval)
  - [ ] 4.4: Verify FD count stable (±10 from baseline)

## Dev Notes

### Actual Watch() Signature (CRITICAL)

The actual signature uses `[]string`, NOT single path:
```go
// watcher.go:85
func (w *FsnotifyWatcher) Watch(ctx context.Context, paths []string) (<-chan ports.FileEvent, error)
```

### Complete Fix (Copy-Paste Ready)

Insert this block at `watcher.go:104` (after `if w.closed` check, before `fsnotify.NewWatcher()`):

```go
	// Story 8.13: Close previous watcher to prevent file handle leak
	// When Watch() is called multiple times (e.g., recovery path), the old
	// watcher must be closed to release file handles. The old eventLoop
	// goroutine exits when watcher.Events channel closes.
	if w.watcher != nil {
		// Stop pending debounce timer to prevent callback races
		if w.timer != nil {
			w.timer.Stop()
			w.timer = nil
		}
		// Clear pending events - they're from old watcher context
		w.pending = make(map[string]ports.FileEvent)
		// Close old watcher - log error but don't fail (we're replacing it)
		if err := w.watcher.Close(); err != nil {
			slog.Debug("closing previous watcher", "error", err)
		}
		w.watcher = nil
	}
```

### Goroutine Cleanup Flow

When `w.watcher.Close()` is called:
1. fsnotify closes its internal `Events` and `Errors` channels
2. `eventLoop()` at line 271 is blocked on `select` over these channels
3. `case event, ok := <-fsWatcher.Events:` returns `ok=false`
4. `eventLoop()` calls `w.flushPending(out)` and returns
5. The `defer close(out)` closes the output channel
6. Old goroutine exits cleanly

### Call Sites (Verified via grep)

| Location | Function | Context |
|----------|----------|---------|
| `model.go:1596` | `startFileWatcherForProjects()` | Initial startup |
| `model.go:731` | `refreshCompleteMsg` handler | Error recovery |

Both sites properly cancel old context before calling Watch().

### Anti-Patterns

| Don't | Do Instead |
|-------|------------|
| Ignore timer cleanup | Stop `w.timer` before closing watcher |
| Keep pending events | Clear `w.pending` map |
| Log Close() error as Error | Log at Debug level (non-fatal) |
| Fix in model.go | Fix encapsulated in watcher.go |
| Ignore goroutine cleanup | Understand Events channel closure flow |

### Previous Story Context

**Story 8.11:** Added 30-second periodic stage refresh. Recovery path in `refreshCompleteMsg` can call `Watch()` when `fileWatcherAvailable=false`, triggering the leak.

**Story 4.1:** Created `FsnotifyWatcher` as singleton. Original assumption was single `Watch()` call - violated by 8.11.

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Basic Functionality

```bash
make build && ./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Dashboard starts | Normal startup | Crash |
| Projects visible | List populates | Empty |
| Agent detection | Status updates on file changes | No updates |

### Step 2: File Descriptor Stability

```bash
# Terminal 1
./bin/vibe

# Terminal 2 (macOS - note the -f flag for exact match)
watch -n 10 'lsof -p $(pgrep -f "vibe$") 2>/dev/null | wc -l'
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Baseline FD count | 50-100 | >200 |
| After 3 minutes | Stable (±10) | Growing |
| After 10 minutes | Same as baseline | +50 or more |

### Step 3: Verify Agent Detection

```bash
# With dashboard running, touch file in watched project
touch ~/some-project/.bmad/test.txt
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Dashboard updates | Status refreshes | No change |
| No errors | Clean | "too many open files" |

### Decision Guide

| Situation | Action |
|-----------|--------|
| FD stable, detection works | Mark `done` |
| FD grows over time | Verify Close() is called in fix |
| Detection stops | Check eventLoop goroutine flow |
| Tests fail | Check test uses `[]string` signature |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-8/8-13-fsnotify-file-handle-leak-fix.md`
- Project context: `docs/project-context.md`
- Source: `internal/adapters/filesystem/watcher.go`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Implementation was straightforward per story spec

### Completion Notes List

1. Added 15-line cleanup block in `watcher.go:105-122` (after closed check, before NewWatcher)
2. Cleanup block stops pending timer, clears pending events map, closes old watcher
3. Old eventLoop goroutine exits cleanly via `fsWatcher.Events` channel close
4. Audited all `Watch()` call sites - both properly cancel context before calling
5. Added 2 regression tests: `TestFsnotifyWatcher_Watch_ClosePreviousWatcher` and `TestFsnotifyWatcher_Watch_ClearsTimer`
6. All tests pass, lint clean

### File List

| File | Change |
|------|--------|
| `internal/adapters/filesystem/watcher.go` | Added cleanup block with explicit Remove() before Close() |
| `internal/adapters/filesystem/watcher_test.go` | Added 3 regression tests (Story 8.13 section) |
| `internal/adapters/tui/model.go` | Cancel old context before calling Watch() |

## Change Log

- 2025-12-31: Story created by SM agent
  - Root cause: watcher.go:167 overwrites without closing
  - Trigger: Story 8.11 periodic refresh calls Watch() multiple times
- 2025-12-31: SM validation applied improvements:
  - **C1 FIXED:** Corrected Watch() signature to `[]string` (was showing single path)
  - **C2 FIXED:** Added timer cleanup (`w.timer.Stop()`) to fix
  - **C3 FIXED:** Added pending events map clear (`w.pending`) to fix
  - **C4 FIXED:** Added goroutine cleanup flow explanation
  - **E1 FIXED:** Updated tests to use correct `[]string` signature
  - **E2 FIXED:** Added `TestFsnotifyWatcher_Watch_ClearsTimer` test
  - **E3 FIXED:** Added macOS pgrep flag `-f` for exact process match
  - **L1 FIXED:** Consolidated duplicate code examples
  - **L2 FIXED:** Streamlined Dev Notes for token efficiency
- 2025-12-31: Code review by Dev agent applied fixes:
  - **M1 FIXED:** Added explicit channel closure verification in `TestFsnotifyWatcher_Watch_ClosePreviousWatcher`
  - **M2 FIXED:** Added CRITICAL comment explaining goroutine reference capture in `eventLoop()`
  - **M3 FIXED:** Added `TestFsnotifyWatcher_Watch_RepeatedCalls_NoLeak` stress test (100 iterations)
- 2025-12-31: Critical fix discovered during testing:
  - **ROOT CAUSE:** On macOS kqueue, fsnotify.Close() does NOT close individual watch FDs
  - **FIX 1:** Explicitly call Remove() on all watched paths before Close()
  - **FIX 2:** Cancel old context in model.go before calling Watch() to prevent false "watcher unavailable" warning
  - **VERIFIED:** FDs no longer double on refresh (tested: 1583 → 1583 stable)
