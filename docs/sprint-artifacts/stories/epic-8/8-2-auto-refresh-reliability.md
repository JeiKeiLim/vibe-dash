# Story 8.2: Auto-Refresh Reliability

Status: done

## Story

As a **user running vibe 24/7**,
I want **the dashboard to refresh automatically and reliably**,
So that **I always see current status without manual refresh**.

## Acceptance Criteria

1. **AC1: Long-Running Session Reliability**
   - Given vibe is running for 1+ hours
   - When no user interaction occurs
   - Then status updates continue appearing automatically

2. **AC2: Timely Waiting State Updates**
   - Given auto-refresh is running
   - When a project's waiting state changes
   - Then the UI updates within 5 seconds

3. **AC3: Status Bar Count Accuracy**
   - Given the tick interval
   - When waiting counts are recalculated
   - Then tickMsg handler triggers status bar update

4. **AC4: No Excessive CPU Usage**
   - Given the increased tick frequency
   - When running 24/7
   - Then CPU usage remains under 2% (idle state)

5. **AC5: Graceful Long-Running Behavior**
   - Given vibe runs for 24+ hours
   - When tick commands accumulate
   - Then no memory leak or goroutine leak occurs

## Tasks / Subtasks

- [x] Task 1: Reduce tick interval for responsive waiting detection (AC: 2, 3)
  - [x] 1.1: Change `tickCmd()` in model.go:321-326 from `time.Minute` to `5 * time.Second`
  - [x] 1.2: Update comment to: `"tickCmd returns a command that ticks every 5 seconds for responsive waiting detection (Story 8.2)"`
  - [x] 1.3: Verify tickMsg handler at model.go:782-793 correctly recalculates counts (no changes expected)

- [x] Task 2: Verify performance requirements (AC: 4, 5)
  - [x] 2.1: Add benchmark test: `BenchmarkTickHandler` - verify <100µs per tick
  - [x] 2.2: Run with `-benchmem` to verify zero allocations in tick handler
  - [x] 2.3: Verify guard clause exists: `if len(m.projects) > 0` (already present at line 788)
  - [x] 2.4: Add test for zero projects edge case

- [x] Task 3: Add long-running session test (AC: 1, 5)
  - [x] 3.1: Create `model_tick_test.go` with simulated 1-hour run
  - [x] 3.2: Send 720 tick messages (60 min × 12 ticks/min at 5s interval)
  - [x] 3.3: Record goroutine count before/after with `runtime.NumGoroutine()` - expect same count (±1)
  - [x] 3.4: Verify model state unchanged after simulation (no corruption)
  - [x] 3.5: Verify no heap growth with `runtime.ReadMemStats()` - delta <1MB

- [x] Task 4: Verify integration with file watcher (AC: 1, 2)
  - [x] 4.1: Test: file event resets waiting timestamp, next tick shows "working"
  - [x] 4.2: Test: no file event for threshold duration, tick correctly shows "waiting"
  - [x] 4.3: Add integration test combining watcher + tick timing

## Dev Notes

### Problem Summary

Users running vibe 24/7 report needing to press 'r' manually because status feels stale. The current 60-second tick interval means waiting state changes can take up to 11 minutes to appear (10 min threshold + 60s tick delay). Reducing to 5-second tick gives maximum 5-second delay.

### Implementation Change

**Single change required in `model.go:321-326`:**

```diff
-// tickCmd returns a command that ticks every 60 seconds for timestamp refresh (Story 4.2, AC4).
+// tickCmd returns a command that ticks every 5 seconds for responsive waiting detection (Story 8.2).
+// Story 4.2 AC4 originally used 1-minute interval, but 24/7 usage requires faster updates.
 func tickCmd() tea.Cmd {
-    return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
+    return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
         return tickMsg(t)
     })
 }
```

**The tickMsg handler (model.go:782-793) needs NO changes** - it's already optimized:
- Guard clause for empty projects: `if len(m.projects) > 0` ✓
- O(n) calculation via `CalculateCountsWithWaiting()` ✓
- Re-schedules next tick ✓

### Technical Requirements

| Setting | Old | New | Verification |
|---------|-----|-----|--------------|
| Tick interval | 60s | 5s | Check tickCmd() code |
| Max waiting detection delay | 60s | 5s | User testing step 2 |
| CPU overhead (idle) | <2% | <2% | `top -l 1 \| grep vibe` |
| Memory per hour | Stable | Stable | `runtime.ReadMemStats()` |
| Goroutine count | N | N | `runtime.NumGoroutine()` |

### Performance Verification Commands

```bash
# Benchmark tick handler (should show 0 allocs, <100µs)
go test -bench=BenchmarkTickHandler -benchmem -count=5 ./internal/adapters/tui/

# CPU usage during operation (should be <2%)
top -l 1 | grep vibe

# Goroutine count (verify no growth over time)
# Add to test: fmt.Println(runtime.NumGoroutine())
```

### Code Locations

| Component | File | Line | Change |
|-----------|------|------|--------|
| tickCmd | model.go | 321-326 | Change `time.Minute` → `5*time.Second` |
| tickMsg handler | model.go | 782-793 | No changes |
| CalculateCountsWithWaiting | status_bar.go | 236-249 | No changes |

### Architecture Compliance

- Change is within `internal/adapters/tui/model.go` (adapter layer) ✓
- No changes to core/ports - tick is presentation-layer concern ✓
- Public API unchanged ✓

### Anti-Patterns to Avoid

| Don't | Do Instead | Why |
|-------|------------|-----|
| Sub-second tick (e.g., 500ms) | 5 seconds | 12 ticks/min is sufficient, lower wastes CPU |
| Allocations in tick handler | Reuse existing model state | Hot path must be allocation-free |
| Skip guard clause | Keep `if len(m.projects) > 0` | Avoid work when no projects |
| Create goroutines per tick | Use tea.Tick's chain model | Prevents goroutine leaks |
| Log on every tick | Log only on state changes | Reduces noise and overhead |
| Make tick configurable | Hardcode 5s | No config proliferation |

### Previous Story Learnings

**From Story 8.1:**
- Code review catches real issues - expect feedback on test coverage and edge cases
- Integration tests important for timing-sensitive features

**From Epic 4 Hotfix H5:**
- Waiting count recalc was added to tick handler (model.go:786-790) for exactly this reason
- Pattern already proven for 24/7 reliability

### References

| Document | Relevance |
|----------|-----------|
| docs/architecture.md (lines 620-658) | Debounce patterns |
| docs/sprint-artifacts/stories/epic-4/4-2-activity-timestamp-tracking.md | Original tick design |
| docs/sprint-artifacts/stories/epic-8/epic-8-ux-polish.md (lines 77-115) | Original requirements |

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Basic Functionality

```bash
make build && ./bin/vibe --waiting-threshold=1
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Dashboard opens | Shows project list | Crash or hang |
| Initial state | Active projects show "working" | Shows "waiting" immediately |

### Step 2: Wait for Threshold + Observe Transition

Wait approximately 70 seconds (1 min threshold + up to 5s tick delay + tolerance).

| Check | Expected | Red Flag |
|-------|----------|----------|
| After ~70 seconds | Status bar shows "1 WAITING" | No change after 90+ seconds |
| Project row | Shows "⏸️ Waiting 1m" | Still shows "● working" |

### Step 3: File Event Clears Waiting

```bash
# In another terminal
touch ~/your-project/.bmad/test.md
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Within 5 seconds | Status bar shows "0 waiting" (dim) | Still shows "1 WAITING" after 10s |
| Project row | Shows "● working" | Still shows "⏸️ Waiting" |

### Step 4: Long-Running CPU Check (Optional)

```bash
# Leave running 10+ minutes, check periodically
top -l 1 | grep vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| CPU usage | <2% idle | Sustained 5%+ |
| Memory | Stable | Growing continuously |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Waiting detection slow (>10s delay) | Verify tick interval is 5s in code |
| High CPU (>5%) | Profile tick handler: `go test -bench=BenchmarkTickHandler -benchmem` |
| Memory growing | Check for goroutine leaks: `runtime.NumGoroutine()` |

## Downstream Dependencies

**No downstream dependencies** - this is a performance/reliability improvement.

**Benefits other stories:**
- Story 8.3 (Stage info in list) benefits from responsive updates
- All future TUI stories benefit from reliable auto-refresh

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

### Completion Notes List

1. Changed tick interval from 60 seconds to 5 seconds in `model.go:321-327`
2. Added comprehensive test file `model_tick_test.go` with:
   - Unit tests for tick command and handler
   - Benchmark tests (BenchmarkTickHandler: ~10µs/op, well under 100µs target)
   - Long-running session simulation (720 ticks = 1 hour @ 5s intervals)
   - Goroutine leak detection (verified ±1 tolerance)
   - Heap growth verification (<1MB growth)
   - Model state integrity verification
   - File watcher + tick integration tests
3. All existing tests continue to pass
4. Linter passes with no issues
5. Benchmark results: ~10µs/op with 5 allocations (~27KB) per tick from tea.Tick command creation. This is unavoidable with Bubble Tea's Elm architecture but has no practical impact - GC reclaims memory efficiently as demonstrated by TestLongRunningSession_NoGoroutineLeak showing negative heap growth over 720 ticks

### Code Review Fixes Applied

1. **H1 (HIGH)**: Fixed stale `tickMsg` type comment at model.go:176-177 - now references Story 8.2 and 5-second interval
2. **M1 (MEDIUM)**: Updated Completion Note #5 to accurately reflect allocation reality (~27KB/tick with GC efficiency)
3. **M3 (MEDIUM)**: Enhanced test comment in TestTickCmd_Returns5SecondInterval with explicit note about Bubble Tea API limitation
4. **L1 (LOW)**: Added documentation to benchmark functions explaining allocation source

### File List

- `internal/adapters/tui/model.go` - Changed tick interval from 60s to 5s, fixed tickMsg comment
- `internal/adapters/tui/model_tick_test.go` - New test file with benchmarks and integration tests, enhanced comments

