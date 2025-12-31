# Story 9.1: TUI Testing Tools Research

Status: done

## Story

As a **developer working on vibe-dash TUI features**,
I want **a comprehensive evaluation of TUI testing tools and approaches**,
So that **we can choose the right testing infrastructure to prevent visual/behavioral bugs like those in Story 8.12**.

## Background

From Epic 8 Retrospective (2025-12-31):

> "Story 8-12 Required ~20 Iterations" - Horizontal layout height handling was particularly challenging. Visual/behavioral issues hard to debug without direct observation. User doesn't know implementation details, had to guide via observations. No automated way to catch anchor point stability issues.

The current testing approach uses traditional unit tests (model_test.go, views_test.go) which verify logic but cannot catch:
- Layout shifts when navigating between projects
- Anchor point stability across different content heights
- Visual regression when terminal is resized
- Resource leaks (like fsnotify FD leak in Story 8.13) during extended runtime

This research story evaluates three approaches: **teatest** (official Charmbracelet library), **VHS** (tape-based recording), and **custom solutions**.

**Scope:** This research focuses on Charmbracelet ecosystem tools (teatest, VHS) plus custom solutions built on existing vibe-dash infrastructure. Third-party terminal emulators (e.g., `hinshun/vt10x`) are out of scope for MVP - we prioritize tools with native Bubble Tea integration.

## Acceptance Criteria

### AC1: Teatest Evaluation
- Given the Charmbracelet teatest library
- When evaluated against vibe-dash requirements
- Then document:
  - API coverage (terminal size simulation, key input, output capture)
  - Golden file workflow for regression testing
  - CI compatibility (color profile normalization)
  - Limitations and experimental status
  - Example code adapted to vibe-dash patterns

### AC2: VHS Evaluation
- Given the Charmbracelet VHS tool
- When evaluated against vibe-dash requirements
- Then document:
  - Tape file syntax for automated scenarios
  - ASCII/text output for golden file comparison
  - CI integration via vhs-action
  - Dependencies (ttyd, ffmpeg) and installation complexity
  - Limitations for non-visual testing (e.g., model state assertions)

### AC3: Custom Solution Evaluation
- Given the option to build custom testing infrastructure
- When compared against teatest and VHS
- Then document:
  - Required components (terminal simulator, output capturer, diff engine)
  - Implementation effort estimate
  - Maintenance burden
  - Advantages for vibe-dash-specific needs
  - **Hybrid approach viability**: Evaluate teatest for interactive behavioral tests + lightweight custom solution for resource monitoring (long-running sessions, FD leak detection)

### AC4: Comparative Analysis
- Given all three approaches evaluated
- When compared against vibe-dash testing requirements
- Then produce a comparison table covering:
  - Terminal size simulation capability
  - Key input simulation capability
  - Output snapshot/golden file support
  - Model state assertions
  - Long-running session support
  - CI pipeline integration
  - Implementation complexity
  - Maintenance burden
  - **Test debuggability**: When a test fails, how easy is it to understand *why*? (Critical for avoiding Story 8.12's 20-iteration debugging problem)
  - **Output determinism**: Does same input produce byte-identical output across runs? (Critical for CI reliability)

### AC5: Recommendation Document
- Given the comparative analysis complete
- When presenting findings
- Then provide:
  - Recommended approach with justification
  - Implementation roadmap for Stories 9.2-9.6
  - Risk assessment and mitigation strategies
  - Integration points with existing test infrastructure
  - **Determinism strategy**: Document known sources of non-determinism (timestamps, ANSI color codes, terminal state) and mitigation approach for CI reliability
  - **Golden file update workflow**: How to update golden files when intentional changes are made (avoid "update all" anti-pattern)

### AC6: Proof of Concept
- Given the recommended approach
- When validating feasibility
- Then create a minimal proof-of-concept that:
  - Simulates a 80x24 terminal
  - Sends navigation keys (j, k)
  - Captures output for comparison
  - Demonstrates it works with existing vibe-dash Model
  - **Success metric**: PoC test MUST be able to detect an intentional regression (e.g., if navigation key is broken, test fails; if anchor point shifts, test fails)

### AC7: Documentation
- Given research and PoC complete
- When documenting findings
- Then create `docs/testing/tui-testing-research.md` containing:
  - Executive summary
  - Detailed evaluation of each approach
  - Comparative analysis table
  - Recommendation with rationale
  - PoC code and results
  - Next steps for Epic 9

## Tasks / Subtasks

- [ ] Task 1: Teatest Deep Dive (AC: 1, 6)
  - [ ] 1.1: Install and configure teatest in test environment
    ```bash
    go get github.com/charmbracelet/x/exp/teatest@latest
    ```
  - [ ] 1.2: Study teatest API documentation and examples
  - [ ] 1.2a: **Verify canonical import path** - Check pkg.go.dev and Charm GitHub for current package location (Charm moves packages; `x/exp/teatest` may have graduated)
  - [ ] 1.3: Create PoC test using vibe-dash Model:
    ```go
    func TestModel_Navigation_Teatest(t *testing.T) {
        m := NewModel(mockRepo)
        tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

        // Wait for initial render
        teatest.WaitFor(t, tm.Output(),
            func(bts []byte) bool {
                return bytes.Contains(bts, []byte("Projects"))
            },
        )

        // Send navigation key
        tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

        // Capture output for golden file comparison
        out, _ := io.ReadAll(tm.FinalOutput(t))
        teatest.RequireEqualOutput(t, out)
    }
    ```
  - [ ] 1.4: Document API coverage, limitations, and gotchas
  - [ ] 1.5: Evaluate CI compatibility (need `lipgloss.SetColorProfile(termenv.Ascii)`)

- [ ] Task 2: VHS Exploration (AC: 2)
  - [ ] 2.1: Install VHS and dependencies (ttyd, ffmpeg)
  - [ ] 2.2: Create sample tape file for vibe-dash:
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
  - [ ] 2.3: Test ASCII output for golden file comparison
  - [ ] 2.4: Evaluate vhs-action for GitHub Actions integration
  - [ ] 2.5: Document dependency complexity and CI requirements

- [ ] Task 3: Custom Solution Analysis (AC: 3)
  - [ ] 3.1: Review existing test patterns in model_test.go
  - [ ] 3.2: Identify gaps between current tests and behavioral testing needs
  - [ ] 3.3: Design custom terminal simulator architecture:
    - Virtual terminal buffer (width x height grid)
    - ANSI escape sequence parser
    - Snapshot capture and diff engine
    - Integration with existing mock infrastructure
  - [ ] 3.4: Estimate implementation effort for custom solution
  - [ ] 3.5: Compare maintenance burden vs. using external tools

- [ ] Task 4: Comparative Analysis (AC: 4)
  - [ ] 4.1: Create evaluation criteria matrix
  - [ ] 4.2: Score each approach against criteria
  - [ ] 4.3: Document trade-offs and decision factors
  - [ ] 4.4: Consider vibe-dash-specific requirements:
    - Horizontal/vertical layout switching
    - Anchor point stability across navigation
    - Terminal resize behavior
    - Long-running session stability (for 8.13-type bugs)

- [ ] Task 5: Recommendation and Roadmap (AC: 5)
  - [ ] 5.1: Synthesize findings into recommendation
  - [ ] 5.2: Create implementation roadmap mapping to Stories 9.2-9.6:
    - Story 9.2: Terminal size simulation framework
    - Story 9.3: Anchor point stability tests
    - Story 9.4: Layout consistency tests
    - Story 9.5: Long-running session tests
    - Story 9.6: CI pipeline integration
  - [ ] 5.3: Identify risks and mitigation strategies
  - [ ] 5.4: Define success metrics for Epic 9

- [ ] Task 6: Documentation (AC: 7)
  - [ ] 6.1: Create `docs/testing/` directory if not exists
  - [ ] 6.2: Write `tui-testing-research.md` with all findings
  - [ ] 6.3: Include code examples and PoC results
  - [ ] 6.4: Add references to source materials

- [ ] Task 7: Validation (AC: all)
  - [ ] 7.1: Review document with project context
  - [ ] 7.2: Verify PoC runs successfully
  - [ ] 7.3: Ensure recommendations align with architecture.md patterns
  - [ ] 7.4: Run `make lint` on any new test code

## Dev Notes

### Research Sources

**Teatest:**
- Official package: `github.com/charmbracelet/x/exp/teatest`
- Blog post: https://carlosbecker.com/posts/teatest/
- Charm blog: https://charm.land/blog/teatest/
- pkg.go.dev: https://pkg.go.dev/github.com/charmbracelet/x/exp/teatest
- Status: Experimental (no backwards compatibility promise)
- **SM Verified (2025-12-31):** Package path confirmed active at `x/exp/teatest`. Still experimental, imported by 3 packages. API matches documentation below.

**VHS:**
- GitHub: https://github.com/charmbracelet/vhs
- Documentation: https://github.com/charmbracelet/vhs/blob/main/README.md
- GitHub Action: https://github.com/charmbracelet/vhs-action
- Dependencies: ttyd, ffmpeg

**Existing vibe-dash Tests:**
- `internal/adapters/tui/model_test.go` - Unit tests for Model
- `internal/adapters/tui/model_responsive_test.go` - Responsive layout tests
- `internal/adapters/tui/views_test.go` - View rendering tests
- Pattern: Standard Go table-driven tests, no behavioral testing

### Teatest API Summary

```go
// Core functions
teatest.NewTestModel(t, model, ...options) *TestModel
teatest.WithInitialTermSize(width, height) Option

// Output handling
tm.FinalOutput(t) io.Reader           // Wait for program to finish
tm.Output() io.Reader                  // Continuous output stream
teatest.WaitFor(t, output, predicate)  // Wait for condition
teatest.RequireEqualOutput(t, out)     // Golden file comparison

// Input simulation
tm.Send(tea.Msg)                       // Send any Bubble Tea message

// Model access
tm.FinalModel(t) tea.Model             // Get final model state

// CI compatibility
lipgloss.SetColorProfile(termenv.Ascii) // Force consistent colors

// CI environment variables (set in test init or CI config)
// FORCE_COLOR=0  // Ensures no ANSI codes leak into golden files
// NO_COLOR=1     // Alternative standard for disabling colors
```

**CI Determinism Note:** Golden file tests may fail in CI if color output differs. Set `FORCE_COLOR=0` or `NO_COLOR=1` environment variable in GitHub Actions workflow, OR call `lipgloss.SetColorProfile(termenv.Ascii)` in test init function.

### VHS Tape Syntax Summary

```tape
# Output formats
Output output.gif              # GIF recording (primary use case)
Output output.mp4              # Video
Output golden.ascii            # ASCII text capture (see limitation below)

# Terminal dimensions
Set Width 80
Set Height 24
Set FontSize 14

# Commands
Type "command"                 # Type text
Enter                          # Press enter
Sleep 1s                       # Wait
Wait /pattern/                 # Wait for output pattern

# Control keys
Ctrl+C                         # Send control sequence
Up/Down/Left/Right             # Arrow keys
```

**VHS ASCII Limitation:** VHS is primarily designed for visual outputs (GIF, MP4 for demos/docs). The `.ascii` output captures terminal text but may include timing artifacts, ANSI escape sequences, or platform-specific variations. For byte-exact golden file comparison, teatest is more reliable. VHS is better suited for demo generation and visual documentation than automated regression testing.

### Existing Test Patterns

From `model_test.go`:

```go
// Current approach - manual state verification
func TestModel_Update_WindowSize(t *testing.T) {
    m := NewModel(nil)
    msg := tea.WindowSizeMsg{Width: 80, Height: 24}
    newModel, cmd := m.Update(msg)
    updated := newModel.(Model)

    if updated.width != 80 || updated.height != 24 {
        t.Errorf("Expected 80x24, got %dx%d", updated.width, updated.height)
    }
}
```

**Gap:** Tests verify model state but not visual output. Cannot catch:
- Anchor point shifting
- Layout proportions being wrong
- Content overflow/clipping

### Key Research Actions

1. **Layout Stability Test (teatest):**
   - Capture golden file output BEFORE navigation
   - Send `j` key, capture output AFTER navigation
   - If bytes differ unexpectedly → layout shifted (regression detected)
   - **Test approach:** Compare specific screen regions, not full output

2. **VHS Behavioral Testing:**
   - Create tape that exercises navigation sequence
   - Capture `.ascii` output
   - Compare with stored baseline
   - **Limitation:** ASCII output less reliable than teatest for byte-exact comparison

3. **Hybrid Approach Decision:**
   - **Teatest:** Interactive tests with state assertions + golden file comparison
   - **VHS:** Demo generation, visual documentation, README recordings
   - **Custom:** Long-running session monitoring (FD counts, goroutine counts)

4. **Long-Running Session Tests:**
   - Story 8.13's FD leak only appeared after extended runtime
   - Neither teatest nor VHS designed for this
   - **Approach:** Custom test using `runtime.NumGoroutine()` and `/proc/self/fd` monitoring
   - Run for N minutes, assert resource counts stable

### Architecture Compliance

**Files to Create:**
- `docs/testing/tui-testing-research.md` - Research findings
- `internal/adapters/tui/teatest_poc_test.go` - PoC (if teatest chosen)

**No Core Changes** - This is a research story. Implementation in Stories 9.2-9.6.

**Testing Standards from architecture.md:**
- Tests co-located with source (`_test.go` suffix)
- Table-driven tests preferred
- Build tags for slow tests (`//go:build integration`)

### Previous Story Learnings

**From Story 8.12 (Horizontal Layout Height Handling):**
- Visual bugs required ~20 iterations to fix
- User observation was the only way to catch anchor point issues
- Unit tests passed but behavior was wrong
- Need: Automated visual regression testing

**From Story 8.13 (fsnotify File Handle Leak):**
- Bug only appeared after extended runtime
- Triggered by periodic refresh (Story 8.11)
- macOS-specific kqueue behavior
- Need: Long-running session tests

**From Story 8.4 (Layout Width Bugs):**
- Race conditions between WindowSizeMsg and ProjectsLoadedMsg
- Component dimensions sensitive to initialization order
- Need: Terminal resize behavior tests

### Evaluation Criteria

| Criterion | Weight | Description |
|-----------|--------|-------------|
| Terminal size simulation | High | Core requirement - must simulate different sizes |
| Key input simulation | High | Essential for navigation testing |
| Output snapshot | High | Golden file comparison for regression |
| Model state assertions | Medium | Verify internal state, not just output |
| Long-running support | Medium | Catch resource leaks over time |
| CI integration | High | Must work in GitHub Actions |
| Implementation effort | Medium | Time to adopt and integrate |
| Maintenance burden | Medium | Ongoing cost to maintain |
| **Test debuggability** | High | When test fails, how easy to understand why? (TEA addition) |
| **Output determinism** | High | Same input → byte-identical output across runs (TEA addition) |

### Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Teatest is experimental | Medium | Monitor Charm releases, have fallback plan |
| VHS requires ffmpeg | Low | Document installation, use Docker in CI |
| Custom solution scope creep | High | Define clear boundaries, time-box research |
| CI flakiness | Medium | Use ASCII mode, fix color profiles |
| Test maintenance overhead | Medium | Start with critical paths only |
| **Golden file brittleness** | High | Document update workflow, avoid "update all" pattern (TEA addition) |
| **Non-deterministic output** | High | Normalize timestamps, force ASCII color profile (TEA addition) |

### TEA (Test Architect) Pre-Validation Review

**Reviewed:** 2025-12-31 by Murat (TEA agent)

**Files Reviewed:**
- `internal/adapters/tui/model_test.go` (~800 lines)
- `internal/adapters/tui/views_test.go` (~440 lines)
- `internal/adapters/tui/model_responsive_test.go` (~630 lines)

**Gap Analysis Confirmation:**

The story's gap analysis is **accurate**. Current tests use:

| Pattern | What It Catches | What It Misses |
|---------|-----------------|----------------|
| `strings.Contains(view, "expected")` | Content presence | Layout position, alignment |
| Manual `m.width = 80` | State initialization | Actual terminal behavior |
| `m.Update(tea.KeyMsg{...})` | Single key handling | Key sequence navigation |
| Table-driven boundary tests | Edge case logic | Visual rendering at edges |

**Key Observation:** Tests are *logical verification*, not *behavioral verification*. Example from `model_responsive_test.go:416-461` tests verify "non-empty view" and "contains newline" but cannot verify the split is visually correct or stable.

**Amendments Added:**
1. Scope boundary (Charmbracelet ecosystem + custom, no third-party VT emulators)
2. Hybrid approach evaluation option in AC3
3. Test debuggability + output determinism criteria in AC4
4. Determinism strategy + golden file workflow in AC5
5. PoC success metric in AC6
6. Task 1.2a for verifying teatest package location

**Risk Assessment Addition:** Golden file brittleness - any styling change breaks all golden files. Recommend documenting update strategy in AC5.

### References

- [Source: docs/sprint-artifacts/retrospectives/epic-8-retro-2025-12-31.md] - Epic 8 retrospective with TUI testing discussion
- [Source: docs/architecture.md#Test Organization] - Testing standards
- [Source: docs/project-context.md#Testing Rules] - Testing rules
- [Source: internal/adapters/tui/model_test.go] - Existing test patterns

## User Testing Guide

**Time needed:** N/A - Research story, no user-facing changes

This story produces documentation and a PoC, not user-visible functionality.

### Validation Checklist

| Check | Expected |
|-------|----------|
| `docs/testing/tui-testing-research.md` exists | Yes |
| Document contains teatest evaluation | Yes |
| Document contains VHS evaluation | Yes |
| Document contains custom solution evaluation | Yes |
| Document contains comparative analysis table | Yes |
| Document contains recommendation | Yes |
| PoC test file compiles | Yes |
| PoC test runs successfully | Yes |
| Recommendation aligns with project architecture | Yes |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-9/9-1-tui-testing-tools-research.md`
- Project context: `docs/project-context.md`
- Epic source: `docs/sprint-artifacts/retrospectives/epic-8-retro-2025-12-31.md`

### Agent Model Used

Claude Opus 4.5

### Debug Log References

N/A

### Completion Notes List

**Completed 2025-12-31 by TEA (Murat)**

1. **Teatest PoC:** 5/5 tests passing, demonstrates all key capabilities
2. **Recommendation:** Hybrid approach (teatest + custom resource monitor)
3. **VHS verdict:** Documentation only, not suitable for automated testing
4. **Key finding:** vibe-dash uses Bubble Tea v1, so must use teatest (not teatest/v2)
5. **CI strategy:** Use NO_COLOR=1, FORCE_COLOR=0 for deterministic golden files
6. **Roadmap:** Stories 9.2-9.6 defined for implementation

**Deliverables:**
- `docs/testing/tui-testing-research.md` - Comprehensive research document
- `internal/adapters/tui/teatest_poc_test.go` - Working PoC (5 tests)

### File List

| File | Purpose |
|------|---------|
| `docs/testing/tui-testing-research.md` | Research findings document |
| `internal/adapters/tui/teatest_poc_test.go` | Proof of concept test (if teatest chosen) |

## Change Log

- 2025-12-31: TEA (Test Architect) pre-validation review
  - Reviewed existing test files (model_test.go, views_test.go, model_responsive_test.go)
  - Confirmed gap analysis accuracy - current tests are logical, not behavioral
  - Added scope boundary (Charmbracelet ecosystem focus)
  - Added hybrid approach evaluation to AC3
  - Added test debuggability + output determinism criteria to AC4
  - Added determinism strategy + golden file workflow to AC5
  - Added PoC success metric to AC6
  - Added Task 1.2a for verifying teatest package location
  - Added golden file brittleness + non-deterministic output risks
  - Story ready for SM validation

- 2025-12-31: Story created by SM agent from Epic 8 retrospective action items
  - Epic 9 (TUI Behavioral Testing) defined as pre-post-MVP gate
  - Story 9-1 focuses on tool research before implementation stories
  - Comprehensive developer context from web research on teatest and VHS

- 2025-12-31: SM Validation (Bob)
  - Verified teatest package path still active at `x/exp/teatest` (pkg.go.dev confirmed)
  - Added CI determinism note (FORCE_COLOR=0, NO_COLOR=1 environment variables)
  - Clarified VHS ASCII output limitation (better for demos than byte-exact testing)
  - Converted "Key Questions" to actionable "Key Research Actions"
  - Story validated and approved for development

- 2025-12-31: Development Complete (TEA - Murat)
  - Installed teatest: `github.com/charmbracelet/x/exp/teatest@latest`
  - Created PoC with 5 tests: BasicModelInitialization, Navigation, DetectsIntentionalRegression, FinalModelState, OutputDeterminism
  - All 5 tests PASS in 1.5s execution time
  - Researched VHS: Recommended for demos only, not automated testing
  - Analyzed custom solution: Needed for long-running resource monitoring
  - Created comparative analysis with weighted scoring matrix
  - Wrote comprehensive research document: `docs/testing/tui-testing-research.md`
  - Defined implementation roadmap for Stories 9.2-9.6
  - All linting and tests pass
  - Status changed to `review`
