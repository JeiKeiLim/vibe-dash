# Story 9.2: Terminal Size Simulation Framework

Status: done

## Story

As a **developer working on vibe-dash TUI tests**,
I want **a reusable teatest-based framework for simulating different terminal sizes**,
So that **I can write consistent, maintainable tests that verify layout behavior across various terminal dimensions**.

## Background

From Story 9.1 Research (2025-12-31):

**Teatest was selected** as the primary behavioral testing framework for vibe-dash TUI. The PoC in `internal/adapters/tui/teatest_poc_test.go` demonstrates:
- Terminal size simulation via `teatest.WithInitialTermSize(80, 24)`
- Key input simulation via `tm.Send(tea.KeyMsg{...})`
- Output capture via `tm.FinalOutput(t)`
- Model state access via `tm.FinalModel(t)`
- Deterministic output when color profile forced to ASCII

**This story establishes the foundation** for all subsequent Epic 9 stories by creating:
1. Reusable test helper functions
2. Golden file infrastructure
3. CI-compatible configuration

## Acceptance Criteria

### AC1: Test Helper Module
- Given the need for consistent teatest setup across tests
- When creating a new behavioral test
- Then developers can use `NewTeatestModel(t, opts...)` with sensible defaults:
  - Default terminal size: 80x24
  - Color profile: ASCII (deterministic)
  - Common mock repository pre-configured
  - Wait timeout: 3 seconds

### AC2: Terminal Size Presets
- Given common terminal size scenarios
- When writing tests
- Then developers can use predefined size constants:
  - `TermSizeStandard` (80x24) - traditional terminal
  - `TermSizeNarrow` (40x24) - mobile/narrow view trigger
  - `TermSizeWide` (160x24) - ultra-wide trigger
  - `TermSizeTall` (80x40) - extended height
  - `TermSizeMinimum` (20x10) - minimum viable
  - `TermSizeUltraWide` (200x30) - beyond max_content_width

### AC3: Golden File Directory Structure
- Given the need for organized golden files
- When golden files are created
- Then they are stored in `internal/adapters/tui/testdata/golden/` with:
  - Subdirectories by test category (e.g., `navigation/`, `layout/`, `resize/`)
  - File naming: `{TestName}.golden`
  - `.gitattributes` entry for `*.golden -text` (preserve line endings)

### AC4: CI Environment Configuration
- Given CI runs on different environments
- When behavioral tests run in GitHub Actions
- Then output is deterministic:
  - `NO_COLOR=1` environment variable set
  - `TERM=dumb` for consistent terminal behavior
  - `lipgloss.SetColorProfile(termenv.Ascii)` in test init

### AC5: Terminal Resize Simulation
- Given tests need to verify resize behavior
- When calling `ResizeTerminal(tm, width, height)`
- Then a `tea.WindowSizeMsg` is sent and processed
- And subsequent output reflects the new dimensions

### AC6: Helper Function Documentation
- Given new developers joining the project
- When looking at the test helper file
- Then clear godoc comments explain:
  - Purpose of each helper function
  - Example usage patterns
  - Relationship to Story 9.1 PoC

### AC7: Existing PoC Integration
- Given the PoC in `teatest_poc_test.go` exists
- When this story is complete
- Then PoC tests are refactored to use the new helpers
- And all 5 PoC tests continue to pass

## Tasks / Subtasks

- [x] Task 1: Create Test Helper Module (AC: 1, 6)
  - [x] 1.1: Create `internal/adapters/tui/teatest_helpers_test.go`
  - [x] 1.2: Implement `NewTeatestModel(t, opts...)` with functional options pattern

    **CRITICAL IMPORTS** (from teatest_poc_test.go):
    ```go
    import (
        "github.com/charmbracelet/x/exp/teatest"  // Experimental - pin version
        "github.com/charmbracelet/lipgloss"
        "github.com/muesli/termenv"
        tea "github.com/charmbracelet/bubbletea"
    )
    ```

    **Pattern:** Follow `TestTeatest_BasicModelInitialization` from `teatest_poc_test.go:44-81`:
    - Force ASCII: `lipgloss.SetColorProfile(termenv.Ascii)`
    - Reuse existing `teatestMockRepository` (lines 291-343) - do NOT create new mock
    - Default size 80x24, timeout 3-5 seconds

  - [x] 1.3: Implement functional option types:
    ```go
    type teatestConfig struct {
        width, height int
        repo          ports.ProjectRepository
        projects      []*domain.Project
    }

    type TeatestOption func(*teatestConfig)

    func WithTermSize(w, h int) TeatestOption
    func WithRepository(repo ports.ProjectRepository) TeatestOption
    func WithProjects(projects []*domain.Project) TeatestOption
    ```
  - [x] 1.4: Add comprehensive godoc comments

- [x] Task 2: Implement Terminal Size Presets (AC: 2)
  - [x] 2.1: Create terminal size constants:
    ```go
    const (
        TermWidthStandard = 80
        TermHeightStandard = 24

        TermWidthNarrow = 40
        TermWidthWide = 160
        TermWidthUltraWide = 200

        TermHeightTall = 40
        TermHeightMinimum = 10
        TermWidthMinimum = 20
    )

    var (
        TermSizeStandard  = [2]int{80, 24}
        TermSizeNarrow    = [2]int{40, 24}
        TermSizeWide      = [2]int{160, 24}
        TermSizeTall      = [2]int{80, 40}
        TermSizeMinimum   = [2]int{20, 10}
        TermSizeUltraWide = [2]int{200, 30}
    )
    ```
  - [x] 2.2: Add `WithTermSizePreset(preset [2]int) TeatestOption`
  - [x] 2.3: Document each preset's purpose in comments

- [x] Task 3: Create Golden File Infrastructure (AC: 3)
  - [x] 3.1: Create directory structure:
    ```
    internal/adapters/tui/testdata/
    └── golden/
        ├── navigation/
        ├── layout/
        └── resize/
    ```
  - [x] 3.2: Update project `.gitattributes`:
    ```
    *.golden -text
    ```
  - [x] 3.3: Create helper for golden file paths:
    ```go
    func GoldenFilePath(category, testName string) string {
        return filepath.Join("testdata", "golden", category, testName+".golden")
    }
    ```

- [x] Task 4: Configure CI Environment (AC: 4)
  - [x] 4.1: Color profile set per-test in NewTeatestModel (NOT global init - see note below)
    **Note:** Global init() with os.Setenv affects other tests in package. Instead:
    - `lipgloss.SetColorProfile(termenv.Ascii)` called in NewTeatestModel
    - CI workflow sets env vars externally
  - [x] 4.2: CI workflow exists - added env vars:
      ```yaml
      env:
        NO_COLOR: 1
        FORCE_COLOR: 0
        TERM: dumb
      ```
  - [x] 4.3: Verify tests pass in both local and CI-like environment

- [x] Task 5: Implement Terminal Resize Simulation (AC: 5)
  - [x] 5.1: Create resize helper:
    ```go
    func ResizeTerminal(tm *teatest.TestModel, width, height int) {
        tm.Send(tea.WindowSizeMsg{Width: width, Height: height})
        // Allow time for resize to process
        time.Sleep(50 * time.Millisecond)
    }
    ```
  - [x] 5.2: Add test for resize behavior:
    ```go
    func TestTeatest_TerminalResize(t *testing.T) {
        tm := NewTeatestModel(t, WithTermSizePreset(TermSizeStandard))

        // Resize to narrow
        ResizeTerminal(tm, TermWidthNarrow, TermHeightStandard)

        // Verify model received new dimensions
        // ... capture output and verify layout changed
    }
    ```

- [x] Task 6: Refactor Existing PoC Tests (AC: 7)
  - [x] 6.1: Review existing `teatest_poc_test.go`
  - [x] 6.2: Refactor `TestTeatest_BasicModelInitialization` to use helpers
  - [x] 6.3: Refactor `TestTeatest_Navigation` to use helpers (partial - uses size constants)
  - [x] 6.4: Refactor `TestTeatest_FinalModelState` to use helpers
  - [x] 6.5: Refactor `TestTeatest_OutputDeterminism` to use helpers
  - [x] 6.6: Keep `TestTeatest_DetectsIntentionalRegression` as-is (direct View() comparison)
  - [x] 6.7: Verify all 5 PoC tests still pass

- [x] Task 7: Add Framework Demonstration Tests (AC: 1, 2, 5)
  - [x] 7.1: Create `internal/adapters/tui/teatest_framework_test.go`
  - [x] 7.2: Add test using each terminal size preset (TestFramework_TerminalSizePresets)
  - [x] 7.3: Add test demonstrating resize simulation (TestFramework_ResizeSimulation, TestFramework_MultipleResizes)
  - [x] 7.4: Add test demonstrating project injection (TestFramework_ProjectInjection)

- [x] Task 8: Validation
  - [x] 8.1: Run `make lint` - must pass ✅
  - [x] 8.2: Run `make test` - all tests pass ✅
  - [x] 8.3: Run tests with `NO_COLOR=1` - output identical ✅
  - [x] 8.4: Verify golden file directory exists (even if empty for now) ✅

## Dev Notes

### Key Learnings from Story 9.1

**From `teatest_poc_test.go`:**
1. **Mock Repository Required:** Model.Init() panics without repository - always inject one
2. **Wait for Ready State:** Use `teatest.WaitFor()` to wait for model initialization
3. **Small Delays After Keys:** Add `time.Sleep(50*ms)` after key sends for processing
4. **Color Profile Critical:** Must force ASCII for deterministic golden files
5. **FinalOutput vs Output:** Use `FinalOutput(t)` after `WaitFinished(t)` for complete capture

**From TUI Testing Research Document:**
- Teatest is experimental (`x/exp/teatest`) - pin version, monitor Charm releases
- Golden file update: `go test -update` to regenerate
- Intermediate output capture is complex - prefer FinalOutput pattern
- Cannot detect timing-sensitive issues (race conditions)

### Architectural Compliance

**Location:** `internal/adapters/tui/` (test files co-located with source)

**Pattern:** Test helper files use `_test.go` suffix per Go convention:
- `teatest_helpers_test.go` - Helper functions
- `teatest_framework_test.go` - Framework demonstration tests
- `teatest_poc_test.go` - Original PoC (refactored)

**No Core Changes:** This story only adds test infrastructure, no production code changes.

### Existing Mock Repository Pattern

From `teatest_poc_test.go:292-343`, use the existing `teatestMockRepository` as base:
```go
type teatestMockRepository struct {
    projects []*domain.Project
}

// Implements all ports.ProjectRepository methods
func (r *teatestMockRepository) FindAll(ctx context.Context) ([]*domain.Project, error) {
    return r.projects, nil
}
// ... other methods
```

### Terminal Size Thresholds

**Constants from `model.go` and `views.go`:**
| Constant | Value | Usage |
|----------|-------|-------|
| `MinWidth` | 60 | Minimum viable terminal width |
| `MinHeight` | 20 | Minimum viable terminal height |
| `HeightThresholdTall` | 35 | Auto-open detail panel |
| `HorizontalDetailThreshold` | 25 | Minimum height for horizontal detail |

**Threshold Behaviors:**
- **Narrow:** `width < 80` triggers vertical-only layout
- **Wide:** `width > maxContentWidth` (default 120) enables centered content
- **isNarrowWidth():** Returns true for 60-79 width range
- **isWideWidth():** Model method using config maxContentWidth

**Test Coverage Matrix:**
| Scenario | Size | Behavior to Test |
|----------|------|------------------|
| Minimum viable | 20x10 | Graceful degradation |
| Narrow | 40x24 | Vertical-only, narrow warning |
| At threshold | 79x24, 80x24 | Threshold boundary |
| Standard | 80x24 | Normal layout |
| Wide | 160x24 | Content capping |
| Ultra-wide | 200x30 | Beyond maxContentWidth |

### Golden File Workflow

1. **Create test:** Write test with `teatest.RequireEqualOutput(t, out)`
2. **First run:** Test fails (no golden file exists)
3. **Generate:** Run `go test -update` to create `.golden` file
4. **Verify:** Manually inspect golden file content
5. **Commit:** Include `.golden` file in commit
6. **Future runs:** Test compares against stored golden file

**Update workflow:**
```bash
# When intentional changes made
go test -run TestSpecificTest -update
git diff testdata/golden/  # Review changes
git add testdata/golden/TestSpecificTest.golden
```

### CI Configuration Notes

The GitHub Actions workflow may not exist yet. Check for:
- `.github/workflows/ci.yml`
- `.github/workflows/test.yml`

If exists, add environment variables. If not, document for future CI setup.

### Project Structure Notes

See **File List** section in Dev Agent Record below for complete list of files to create/modify.

### Testing Standards

From `docs/project-context.md`:
- **Co-locate tests:** `_test.go` next to source
- **Table-driven:** Use `tests []struct{...}` pattern
- **Build tags for integration:** `//go:build integration`
- **Mocks in test file:** Extract to `*_mock.go` only if reused

This story creates shared test helpers, so extraction to separate file is appropriate.

### Previous Story Learnings

**From Story 8.12 (Horizontal Layout):**
- Visual bugs required ~20 iterations to fix
- Layout anchor point stability was critical
- Tests should verify position doesn't shift unexpectedly

**From Story 8.4 (Layout Width Bugs):**
- Race conditions between WindowSizeMsg and ProjectsLoadedMsg
- Component dimensions sensitive to initialization order
- ResizeTerminal helper must account for message ordering

### References

- [Source: docs/testing/tui-testing-research.md] - Comprehensive teatest evaluation
- [Source: internal/adapters/tui/teatest_poc_test.go] - Working PoC with 5 tests
- [Source: docs/architecture.md#Test Organization] - Testing standards
- [Source: docs/project-context.md#Testing Rules] - Testing rules
- [Source: internal/adapters/tui/model.go] - Model implementation for terminal handling

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Verify Tests Pass

```bash
cd /Users/limjk/GitHub/JeiKeiLim/vibe-dash
make test
```

**Expected:** All tests pass, including:
- All 5 original PoC tests in `teatest_poc_test.go`
- New framework tests in `teatest_framework_test.go`

### Step 2: Verify Determinism

```bash
# Run twice, compare output
NO_COLOR=1 go test ./internal/adapters/tui/... -v -run Teatest 2>&1 | head -50
NO_COLOR=1 go test ./internal/adapters/tui/... -v -run Teatest 2>&1 | head -50
```

**Expected:** Output is identical between runs (no timestamps or variable content in test output).

### Step 3: Verify Golden File Directory

```bash
ls -la internal/adapters/tui/testdata/golden/
```

**Expected:** Directory structure exists:
```
testdata/golden/
├── navigation/
├── layout/
└── resize/
```

### Step 4: Verify .gitattributes

```bash
grep "golden" .gitattributes
```

**Expected:** Contains `*.golden -text`

### Decision Guide

| Situation | Action |
|-----------|--------|
| All tests pass, directories exist | Mark `done` |
| Tests fail | Do NOT approve, document issue |
| Missing directories | Do NOT approve, document issue |
| .gitattributes missing entry | Do NOT approve, document issue |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-9/9-2-terminal-size-simulation-framework.md`
- Research doc: `docs/testing/tui-testing-research.md`
- PoC source: `internal/adapters/tui/teatest_poc_test.go`
- Project context: `docs/project-context.md`

### Critical Anti-Patterns (DO NOT)

1. **DO NOT** create new mock repository - reuse `teatestMockRepository` from `teatest_poc_test.go:291-343`
2. **DO NOT** use wrong teatest import - must be `github.com/charmbracelet/x/exp/teatest`
3. **DO NOT** forget ASCII color profile - `lipgloss.SetColorProfile(termenv.Ascii)` is REQUIRED
4. **DO NOT** skip the `time.Sleep(50*time.Millisecond)` after key sends - see PoC line 120-124
5. **DO NOT** use `tea.KeyDown`/`tea.KeyUp` types - use `tea.KeyRunes` with 'j'/'k' chars (see PoC)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

1. **Task 1 (Helper Module):** Created `teatest_helpers_test.go` with functional options pattern. Helper `NewTeatestModel(t, opts...)` provides:
   - Default 80x24 terminal
   - ASCII color profile for determinism
   - Mock repository from teatest_poc_test.go reused
   - Comprehensive godoc comments with usage examples

2. **Task 2 (Size Presets):** All 6 presets implemented as constants and [2]int variables. Each documented with purpose.

3. **Task 3 (Golden Files):** Directory structure created with .gitkeep placeholders. `.gitattributes` added with `*.golden -text`.

4. **Task 4 (CI Config):**
   - **Design decision:** Avoided global `init()` with `os.Setenv` as it affects other tests in package (broke `TestUseColorReflectsEnvironment`).
   - Instead: `lipgloss.SetColorProfile(termenv.Ascii)` called per-test in `NewTeatestModel`.
   - CI workflow updated with NO_COLOR, FORCE_COLOR, TERM env vars.

5. **Task 5 (Resize):** `ResizeTerminal(tm, w, h)` helper sends WindowSizeMsg with 50ms delay. Tests verify resize works.

6. **Task 6 (PoC Refactor):** 3 of 5 PoC tests refactored to use `NewTeatestModel`. Navigation test partially refactored (uses constants but needs manual component setup). All 5 tests pass.

7. **Task 7 (Framework Tests):** 9 demonstration tests created:
   - TestFramework_TerminalSizePresets (6 subtests for each preset)
   - TestFramework_CustomTerminalSize
   - TestFramework_ResizeSimulation
   - TestFramework_MultipleResizes
   - TestFramework_ProjectInjection
   - TestFramework_GoldenFilePath (3 subtests)
   - TestFramework_FullWorkflow
   - TestFramework_DeterministicOutput

8. **Task 8 (Validation):** All checks pass: lint, tests, directories exist, gitattributes configured.

### File List

| File | Action | Purpose |
|------|--------|---------|
| `internal/adapters/tui/teatest_helpers_test.go` | CREATED | Reusable test helper functions |
| `internal/adapters/tui/teatest_framework_test.go` | CREATED | Framework demonstration tests |
| `internal/adapters/tui/testdata/golden/navigation/.gitkeep` | CREATED | Golden directory placeholder |
| `internal/adapters/tui/testdata/golden/layout/.gitkeep` | CREATED | Golden directory placeholder |
| `internal/adapters/tui/testdata/golden/resize/.gitkeep` | CREATED | Golden directory placeholder |
| `.gitattributes` | CREATED | Add `*.golden -text` |
| `internal/adapters/tui/teatest_poc_test.go` | MODIFIED | Refactored to use NewTeatestModel helper |
| `.github/workflows/ci.yml` | MODIFIED | Added NO_COLOR, FORCE_COLOR, TERM env vars |

## Change Log

- 2025-12-31: Code review passed by Dev agent (Amelia, Claude Opus 4.5)
  - Removed redundant max/min helper functions (Go 1.21+ built-ins)
  - Fixed misleading comments about unused imports
  - Added documentation for UltraWide height (30) rationale
  - Clarified 100ms vs 50ms sleep timing difference
  - Story status: done

- 2025-12-31: Implementation complete by Dev agent (Claude Opus 4.5)
  - All 8 tasks completed
  - 9 framework tests + 5 refactored PoC tests all pass
  - Key learning: avoid global init() with os.Setenv - affects other tests
  - Story status: review

- 2025-12-31: Story validated by SM agent (Bob) via validate-create-story workflow
  - Added CRITICAL IMPORTS section with explicit package paths (C1, C2 fix)
  - Added reference to existing teatestMockRepository instead of new mock (C3, E1 fix)
  - Clarified CI workflow check for non-existent `.github/workflows/` (E4 fix)
  - Added Terminal Size Constants table with actual values from model.go (E3 fix)
  - Consolidated file lists to avoid duplication (O2 fix)
  - Added Critical Anti-Patterns section in Dev Agent Record
  - All improvements applied per user request

- 2025-12-31: Story created by SM agent (Bob)
  - Comprehensive story context from Story 9.1 research
  - All acceptance criteria derived from research recommendations
  - Tasks based on implementation roadmap from `tui-testing-research.md`
  - Dev notes include learnings from PoC and Epic 8 stories
  - Ready for development without additional elicitation
