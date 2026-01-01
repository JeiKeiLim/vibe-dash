# TUI Testing Tools Research

**Story:** 9.1 - TUI Testing Tools Research
**Author:** Murat (TEA - Test Architect Agent)
**Date:** 2025-12-31
**Status:** Complete

---

## Executive Summary

This document evaluates three approaches for behavioral TUI testing in vibe-dash: **teatest** (Charmbracelet's official testing library), **VHS** (tape-based recording), and **custom solutions**.

**Recommendation:** Adopt a **hybrid approach**:
1. **Teatest** for interactive behavioral tests with golden file comparison
2. **Custom lightweight solution** for long-running resource monitoring (goroutine leaks, FD leaks)
3. **VHS** for demo generation and documentation only (not automated testing)

This recommendation is based on:
- Teatest's native Bubble Tea integration and proven PoC success (5/5 tests passed)
- VHS's limitations for byte-exact testing
- The specific needs identified from Story 8.12 (~20 iterations) and Story 8.13 (FD leak)

---

## Problem Statement

From Epic 8 Retrospective (2025-12-31):

> "Story 8-12 Required ~20 Iterations" - Horizontal layout height handling was particularly challenging. Visual/behavioral issues hard to debug without direct observation. User doesn't know implementation details, had to guide via observations. No automated way to catch anchor point stability issues.

**Current testing gaps identified:**
- Tests verify **logic** but not **visual output**
- No **anchor point stability** testing
- No **layout regression** detection
- No **long-running session** testing (Story 8.13 FD leak wasn't caught)
- ~400 tests across 17 files focus on state transitions, not rendered output

---

## Approach 1: Teatest (Charmbracelet Official)

### Overview

- **Package:** `github.com/charmbracelet/x/exp/teatest`
- **Status:** Experimental (no backwards compatibility promise)
- **Version tested:** v0.0.0-20251215102626-e0db08df7383
- **Compatible with:** Bubble Tea v1.x (vibe-dash uses v1.3.10)

### API Coverage

| Feature | Support | Notes |
|---------|---------|-------|
| Terminal size simulation | `WithInitialTermSize(80, 24)` | Full support |
| Key input simulation | `tm.Send(tea.KeyMsg{...})` | Full support |
| Output capture | `tm.FinalOutput(t)` | Captures all rendered output |
| Golden file comparison | `RequireEqualOutput(t, out)` | Built-in with `-update` flag |
| Model state assertions | `tm.FinalModel(t)` | Access internal state |
| Intermediate output | `tm.Output()` | Streaming, complex to use |
| CI compatibility | `lipgloss.SetColorProfile(termenv.Ascii)` | Required for determinism |

### PoC Results

**All 5 proof-of-concept tests passed:**

```
=== RUN   TestTeatest_BasicModelInitialization
    Teatest basic initialization: SUCCESS (output length: 77 bytes)
--- PASS: TestTeatest_BasicModelInitialization (0.00s)

=== RUN   TestTeatest_Navigation
    Teatest navigation simulation: SUCCESS (final output: 242 bytes)
--- PASS: TestTeatest_Navigation (0.25s)

=== RUN   TestTeatest_DetectsIntentionalRegression
    Teatest regression detection: SUCCESS (views differ as expected)
--- PASS: TestTeatest_DetectsIntentionalRegression (0.00s)

=== RUN   TestTeatest_FinalModelState
    Final model state - ready: true, width: 80, height: 24, showHelp: false
    Teatest final model state access: SUCCESS
--- PASS: TestTeatest_FinalModelState (0.40s)

=== RUN   TestTeatest_OutputDeterminism
    Teatest output determinism: SUCCESS (identical outputs)
--- PASS: TestTeatest_OutputDeterminism (0.20s)

PASS
ok  	github.com/JeiKeiLim/vibe-dash/internal/adapters/tui	1.660s
```

### Key Findings

**Strengths:**
1. Native Bubble Tea integration - works directly with `tea.Model`
2. Golden file comparison built-in with `RequireEqualOutput`
3. Can access final model state for state assertions
4. Deterministic output when color profile is forced to ASCII
5. Fast execution (~1.7s for 5 tests)
6. No external dependencies

**Limitations:**
1. Experimental status - API may change
2. Not designed for long-running session tests
3. Intermediate output capture is complex (streaming buffer)
4. Cannot detect timing-sensitive issues (race conditions)

### CI Determinism Strategy

To ensure golden files produce identical output across environments:

```go
func init() {
    // Force ASCII color profile for deterministic output
    lipgloss.SetColorProfile(termenv.Ascii)

    // Set environment variables as backup
    os.Setenv("NO_COLOR", "1")
    os.Setenv("TERM", "dumb")
}
```

**GitHub Actions workflow addition:**
```yaml
env:
  NO_COLOR: 1
  FORCE_COLOR: 0
```

### Golden File Update Workflow

1. **Initial creation:** Run `go test -update` to generate golden files
2. **Intentional changes:** Run `go test -update` after verifying output is correct
3. **Avoid "update all":** Only update specific test's golden file, not entire suite
4. **Review diffs:** Use `git diff` on `.golden` files before committing
5. **Git attributes:** Add `*.golden -text` to `.gitattributes` to preserve line endings

---

## Approach 2: VHS (Tape-Based Recording)

### Overview

- **Repository:** https://github.com/charmbracelet/vhs
- **Stars:** 18.1k (well-maintained)
- **Dependencies:** ttyd, ffmpeg
- **Output formats:** GIF, MP4, WebM, PNG, ASCII/TXT

### Sample Tape File

```tape
Output demo.ascii
Set Width 80
Set Height 24
Set FontSize 14

Type "./bin/vibe"
Enter
Sleep 1s

Type "j"
Sleep 500ms
Type "k"
Sleep 500ms

Type "q"
```

### Key Findings

**Strengths:**
1. Excellent for demo generation (GIF for README, etc.)
2. Human-readable tape syntax
3. GitHub Action available (`charmbracelet/vhs-action`)
4. Good for visual documentation
5. Can output ASCII for basic comparison

**Limitations:**
1. **Not designed for byte-exact testing** - ASCII output includes timing artifacts
2. **Cannot access model state** - purely visual
3. **External dependencies** - requires ttyd and ffmpeg
4. **Slower execution** - spawns real processes
5. **Less precise** - timing-based, not event-based
6. **CI complexity** - need to install dependencies

### Verdict

**VHS is recommended for demo generation only, not automated testing.**

For automated regression testing, teatest provides more reliable, faster, and more precise results without external dependencies.

---

## Approach 3: Custom Solution

### Current Test Infrastructure Analysis

**Existing coverage (~400 tests across 17 files):**

| Pattern | What It Catches | What It Misses |
|---------|-----------------|----------------|
| `strings.Contains(view, "expected")` | Content presence | Layout position |
| Manual `m.width = 80` | State initialization | Actual terminal behavior |
| `m.Update(tea.KeyMsg{...})` | Single key handling | Key sequence navigation |
| Table-driven boundary tests | Edge case logic | Visual rendering at edges |

**Gap Analysis:**

| Need | Current Status | Solution |
|------|----------------|----------|
| Layout regression | Not tested | Teatest golden files |
| Anchor point stability | Not tested | Teatest golden files |
| Terminal resize behavior | Partial (logic only) | Teatest with size changes |
| Long-running sessions | Not tested | Custom resource monitor |
| Goroutine leaks | Not tested | goleak + custom |
| FD leaks | Not tested | Custom `/dev/fd` monitor |

### Custom Solution Components Needed

For gaps that teatest doesn't cover:

**1. Long-Running Session Monitor**
```go
type ResourceMonitor struct {
    initialGoroutines int
    initialFDs        int
    checkInterval     time.Duration
}

func (m *ResourceMonitor) Start() {
    m.initialGoroutines = runtime.NumGoroutine()
    m.initialFDs = countOpenFDs() // macOS: read /dev/fd
}

func (m *ResourceMonitor) AssertNoLeaks(t *testing.T) {
    currentGoroutines := runtime.NumGoroutine()
    currentFDs := countOpenFDs()

    if currentGoroutines > m.initialGoroutines + threshold {
        t.Errorf("Goroutine leak: %d -> %d",
            m.initialGoroutines, currentGoroutines)
    }
    // Similar for FDs
}
```

**2. macOS File Descriptor Counting**
```go
func countOpenFDs() int {
    entries, _ := os.ReadDir("/dev/fd")
    return len(entries)
}
```

**3. Integration with goleak**
```go
import "go.uber.org/goleak"

func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```

### Implementation Effort Estimate

| Component | Effort | Maintenance |
|-----------|--------|-------------|
| Teatest integration | Low (PoC done) | Low |
| Golden file infrastructure | Low | Medium (update workflow) |
| Resource monitor | Medium | Low |
| goleak integration | Low | Low |
| CI pipeline changes | Low | Low |

---

## Comparative Analysis

### Evaluation Matrix

| Criterion | Weight | Teatest | VHS | Custom |
|-----------|--------|---------|-----|--------|
| Terminal size simulation | High | 5 | 4 | 3 |
| Key input simulation | High | 5 | 3 | 4 |
| Output snapshot | High | 5 | 3 | 3 |
| Model state assertions | Medium | 5 | 0 | 4 |
| Long-running support | Medium | 2 | 1 | 5 |
| CI integration | High | 5 | 3 | 4 |
| Implementation effort | Medium | 5 | 4 | 2 |
| Maintenance burden | Medium | 4 | 3 | 3 |
| Test debuggability | High | 4 | 2 | 4 |
| Output determinism | High | 5 | 2 | 4 |
| **Weighted Total** | | **46** | **25** | **36** |

**Scoring:** 5=Excellent, 4=Good, 3=Adequate, 2=Poor, 1=Inadequate, 0=Not Supported

### Trade-offs Summary

| Approach | Best For | Not Good For |
|----------|----------|--------------|
| **Teatest** | Behavioral tests, golden files, state assertions | Long-running sessions, timing issues |
| **VHS** | Demo generation, visual documentation | Automated testing, CI |
| **Custom** | Resource monitoring, long-running tests | Complex behavioral tests |

---

## Recommendation

### Primary: Teatest for Behavioral Testing

Adopt teatest as the primary behavioral testing framework for vibe-dash TUI.

**Implementation Roadmap:**

| Story | Description | Effort |
|-------|-------------|--------|
| 9.2 | Terminal size simulation framework | Small |
| 9.3 | Anchor point stability tests (golden files) | Medium |
| 9.4 | Layout consistency tests | Medium |
| 9.5 | Long-running session tests (custom) | Medium |
| 9.6 | CI pipeline integration | Small |

### Secondary: Custom Resource Monitor

For Story 8.13-type bugs (FD leaks during extended runtime):

1. Integrate `go.uber.org/goleak` for goroutine leak detection
2. Add custom FD counter for macOS using `/dev/fd`
3. Create long-running test with 5+ minute runtime
4. Run as separate `//go:build integration` test

### Tertiary: VHS for Documentation

Use VHS only for:
- README demo GIFs
- Feature documentation
- User-facing visual guides

Do NOT use for automated testing.

---

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Teatest API changes | Medium | Medium | Pin version, monitor Charm releases |
| Golden file brittleness | High | Medium | Document update workflow, avoid "update all" |
| Non-deterministic output | High | Low | Force ASCII color profile, set NO_COLOR |
| CI test flakiness | Medium | Low | Add retries, fix timing-sensitive tests |
| Resource monitor overhead | Low | Low | Run as separate integration test |

---

## Implementation Steps

### Story 9.2: Terminal Size Simulation Framework

1. Add teatest dependency (already done)
2. Create `internal/adapters/tui/teatest_helpers_test.go` with common setup
3. Add golden file directory: `internal/adapters/tui/testdata/golden/`
4. Update `.gitattributes` for golden files
5. Add GitHub Actions workflow changes

### Story 9.3: Anchor Point Stability Tests

1. Create tests that navigate through project list
2. Capture golden files before/after navigation
3. Verify selected item position doesn't shift unexpectedly
4. Test with different terminal heights

### Story 9.4: Layout Consistency Tests

1. Test horizontal vs vertical layout switching
2. Test narrow/wide terminal transitions
3. Verify detail panel proportions
4. Test edge cases (minimum dimensions, very wide terminals)

### Story 9.5: Long-Running Session Tests

1. Integrate goleak for goroutine monitoring
2. Add FD counter for macOS
3. Create 5-minute session test
4. Run file watcher, auto-refresh, navigation in loop
5. Assert no resource growth

### Story 9.6: CI Pipeline Integration

1. Add `NO_COLOR=1` and `FORCE_COLOR=0` to GitHub Actions
2. Configure golden file update workflow
3. Add integration test job for long-running tests
4. Set up golden file diff reporting

---

## PoC Code Location

The proof-of-concept tests are in:

```
internal/adapters/tui/teatest_poc_test.go
```

This file contains 5 working tests demonstrating:
1. Basic model initialization with teatest
2. Navigation key simulation
3. Intentional regression detection
4. Final model state access
5. Output determinism verification

---

## References

### Teatest
- [Package Documentation](https://pkg.go.dev/github.com/charmbracelet/x/exp/teatest)
- [Carlos Becker's Blog Post](https://carlosbecker.com/posts/teatest/)
- [Charm Blog](https://charm.land/blog/teatest/)

### VHS
- [GitHub Repository](https://github.com/charmbracelet/vhs)
- [GitHub Action](https://github.com/charmbracelet/vhs-action)

### Resource Monitoring
- [uber-go/goleak](https://github.com/uber-go/goleak)
- [Counting FDs on macOS](https://zameermanji.com/blog/2021/8/1/counting-open-file-descriptors-on-macos/)
- [Goroutine Leak Detection](https://medium.com/@siddharthnarayan/detecting-goroutine-thread-leaks-in-go-tools-and-techniques-4ca4b154c9d3)

### vibe-dash Context
- Epic 8 Retrospective: `docs/sprint-artifacts/retrospectives/epic-8-retro-2025-12-31.md`
- Architecture: `docs/architecture.md`
- Project Context: `docs/project-context.md`

---

## Appendix: Test Patterns

### Golden File Test Pattern

```go
func TestLayoutStability_Vertical(t *testing.T) {
    // Force deterministic output
    lipgloss.SetColorProfile(termenv.Ascii)

    repo := createMockRepoWithProjects(5)
    m := NewModel(repo)

    tm := teatest.NewTestModel(t, m,
        teatest.WithInitialTermSize(80, 40))

    // Navigate to third item
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

    // Quit
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
    tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

    // Compare against golden file
    out, _ := io.ReadAll(tm.FinalOutput(t))
    teatest.RequireEqualOutput(t, out)
}
```

### Resource Monitor Pattern

```go
func TestLongRunningSession_NoLeaks(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping long-running test in short mode")
    }

    defer goleak.VerifyNone(t)

    initialFDs := countOpenFDs()

    // Run for 5 minutes with activity
    for i := 0; i < 300; i++ {
        // Simulate activity
        time.Sleep(time.Second)

        // Check periodically
        if i % 60 == 0 {
            currentFDs := countOpenFDs()
            if currentFDs > initialFDs + 10 {
                t.Errorf("FD growth detected: %d -> %d",
                    initialFDs, currentFDs)
            }
        }
    }
}
```

---

## Local Development Setup

To ensure your local environment produces deterministic output matching CI:

### Required Environment Variables

Set these in your shell profile (`~/.bashrc`, `~/.zshrc`) or before running tests:

```bash
export NO_COLOR=1
export FORCE_COLOR=0
export TERM=dumb
```

### Quick Test Commands

```bash
# Run unit tests only (fast, no integration)
make test

# Run all tests including integration/behavioral
make test-all

# Run TUI behavioral tests only (debugging)
make test-behavioral

# Update golden files after intentional changes
go test -tags=integration -run 'TestLayout_' ./internal/adapters/tui/... -update
```

### Troubleshooting Non-Deterministic Output

If golden file tests fail locally but you haven't changed any code:

1. **Check environment variables**: Ensure NO_COLOR, FORCE_COLOR, TERM are set correctly
2. **Verify color profile**: Tests call `lipgloss.SetColorProfile(termenv.Ascii)` in setup
3. **Check terminal emulator**: Some terminals override TERM settings
4. **Compare byte-by-byte**: Use `hexdump -C file.golden | head` to spot invisible differences

### CI vs Local Consistency

The CI workflow (`.github/workflows/ci.yml`) sets the same environment variables globally:

```yaml
env:
  NO_COLOR: 1
  FORCE_COLOR: 0
  TERM: dumb
```

Your local tests should match CI output when these variables are set correctly.

---

## Change Log

- 2026-01-01: Story 9.6 CI integration - added Local Development Setup section
  - Environment variables documentation
  - Quick test commands reference
  - Troubleshooting guide for non-deterministic output

- 2025-12-31: Initial research document created
  - Teatest PoC completed with 5/5 passing tests
  - VHS evaluated for demo generation
  - Custom solution designed for resource monitoring
  - Comparative analysis and recommendation finalized
