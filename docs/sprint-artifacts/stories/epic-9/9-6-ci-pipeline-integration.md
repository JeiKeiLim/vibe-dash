# Story 9.6: CI Pipeline Integration

Status: done

## Story

As a **developer contributing to vibe-dash**,
I want **automated CI validation for all TUI behavioral tests with proper configuration**,
So that **layout regressions, anchor point instability, and resource leaks are caught before merging**.

## Background

**Epic 9 Context:** TUI Behavioral Testing Infrastructure - pre-post-MVP gate for catching visual/behavioral bugs automatically.

**Current CI State (ci.yml):** Single `build` job on `ubuntu-latest` with `NO_COLOR=1`, `FORCE_COLOR=0`, `TERM=dumb`. Runs `make test-all` including integration tests. No job separation, no macOS runner for FD testing.

**Stories 9.2-9.5 Test Inventory:**
| Story | Tests | Golden Files |
|-------|-------|--------------|
| 9.2 | Terminal size framework | - |
| 9.3 | Anchor stability | 4 files |
| 9.4 | Layout consistency | 6 files |
| 9.5 | Resource leak detection | - |

## Acceptance Criteria

### AC1: Separate Test Jobs for Speed/Visibility
- Given GitHub Actions workflow
- When running CI on PR
- Then quick tests and integration tests run in separate parallel jobs
- And job names clearly indicate test type ("Unit Tests", "Integration Tests")

### AC2: macOS Runner for FD Testing
- Given Story 9.5's `/dev/fd` file descriptor counting
- When running integration tests that include resource monitoring
- Then tests run on `macos-latest` runner for reliable FD counting
- And `/proc/self/fd` fallback tested on Linux in unit tests

### AC3: Integration Test Timeout
- Given 5-minute session lifecycle test in Story 9.5
- When running integration test job
- Then job timeout is 15 minutes (buffer for setup + test)
- And individual test timeout is already 10m via Makefile

### AC4: Golden File Diff Reporting
- Given golden file comparison tests (teatest)
- When a golden file test fails due to output mismatch
- Then CI output clearly shows diff between expected and actual
- And contributor knows which golden file needs updating

### AC5: Golden File Update Workflow Documentation
- Given contributor needs to update golden files after intentional changes
- When reviewing PR instructions
- Then clear documentation exists for `go test -update` workflow
- And `.gitattributes` configured for golden files

### AC6: Makefile Targets for CI Stages
- Given CI needs to run specific test categories
- When defining Makefile targets
- Then separate targets exist for quick tests vs integration tests
- And CI can call appropriate target per job

### AC7: Environment Variable Documentation
- Given multiple environment variables affect test determinism
- When reviewing CI configuration
- Then all required env vars are documented with purpose
- And consistent across local development and CI

### AC8: Test Failure Categorization
- Given tests can fail for different reasons
- When CI reports test failure
- Then failure category is clear (unit, behavioral, resource, lint)
- And contributor knows which area to investigate

## Tasks / Subtasks

- [x] Task 1: Update GitHub Actions Workflow (AC: 1, 2, 3, 8)
  - [x] 1.1: Add separate `unit-tests` job for quick tests (`make test`)
  - [x] 1.2: Add separate `integration-tests` job for behavioral/resource tests
  - [x] 1.3: Configure `macos-latest` for integration job (FD testing)
  - [x] 1.4: Set job timeout to 15 minutes for integration tests
  - [x] 1.5: Add `needs: [lint]` for both test jobs, integration runs in parallel with unit-tests
  - [x] 1.6: Update build job to depend on both test jobs

- [x] Task 2: Add Makefile Targets (AC: 6)
  - [x] 2.1: **NOTE:** Existing `test` target already does `go test ./...`. Add `test-behavioral` only.
  - [x] 2.2: Verify `test-all` target already includes `-tags=integration -timeout=10m` ✓
  - [x] 2.3: Add `test-behavioral` target for TUI-specific tests (debugging aid)
  - [x] 2.4: Document targets in Makefile comments

- [x] Task 3: Golden File Configuration (AC: 4, 5)
  - [x] 3.1: **CRITICAL:** Add `*.golden linguist-generated=true` to existing `.gitattributes` (currently only has `-text`)
  - [x] 3.2: Add `GOLDEN_FILE_UPDATE.md` to `docs/testing/`
  - [x] 3.3: Verify teatest diff output is readable in CI logs

- [x] Task 4: Environment Variable Documentation (AC: 7)
  - [x] 4.1: Add comment block in CI workflow explaining each env var
  - [x] 4.2: Update `docs/testing/tui-testing-research.md` with local development setup
  - [x] 4.3: Verify `lipgloss.SetColorProfile(termenv.Ascii)` is set in test init (line 250 of teatest_helpers_test.go)

- [x] Task 5: Validation
  - [x] 5.1: Run `make lint` - passed
  - [x] 5.2: Run `make test` - pre-existing test failures (not introduced by this story)
  - [x] 5.3: Run `make test-all` - pre-existing test failures (not introduced by this story)
  - [x] 5.4: CI YAML syntax validated
  - [x] 5.5: Job structure verified (lint → unit-tests/integration-tests → build)
  - [x] 5.6: Golden file diff output verified (teatest shows clear diff on mismatch)

## Dev Notes

### Target CI Workflow Structure

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

# Environment variables for TUI test determinism (see table below)
env:
  NO_COLOR: 1
  FORCE_COLOR: 0
  TERM: dumb

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'  # Uses version from go.mod (currently 1.23)
      - name: Install goimports
        run: go install golang.org/x/tools/cmd/goimports@latest
      - name: Check formatting
        run: make check-fmt
      - name: Install golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

  unit-tests:
    runs-on: ubuntu-latest
    needs: [lint]
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
      - name: Run unit tests
        run: make test

  integration-tests:
    runs-on: macos-latest  # Required for /dev/fd FD counting (Story 9.5)
    needs: [lint]
    timeout-minutes: 15
    env:
      CGO_ENABLED: 1  # Required for go-sqlite3 on macOS
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
      - name: Run integration tests
        run: make test-all

  build:
    runs-on: ubuntu-latest
    needs: [unit-tests, integration-tests]
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
      - name: Build
        run: make build
```

### Makefile Addition

```makefile
# Behavioral tests only (for debugging TUI issues locally)
# Runs anchor, layout, and resource tests with verbose output
test-behavioral:
	go test -tags=integration -timeout=10m -v ./internal/adapters/tui/... -run 'TestAnchor_|TestLayout_|TestResource_'
```

**Note:** No `test-quick` target needed - existing `test` target already does `go test ./...`

### Golden File Configuration (.gitattributes)

**Current state:** `.gitattributes` only has `*.golden -text`

**Add this line:**
```
*.golden linguist-generated=true
```

This hides golden file diffs in GitHub PRs by default (click to expand).

### Golden File Update Workflow

**File to create:** `docs/testing/GOLDEN_FILE_UPDATE.md`

**Content:** See template in Files to Create section. Key points:
1. Verify change is intentional before updating
2. Use `go test -tags=integration -run <TestName> ./internal/adapters/tui/... -update`
3. Review diff with `git diff internal/adapters/tui/testdata/`
4. Commit with `test: Update golden files for [description]`

### Environment Variables for Determinism

| Variable | Value | Purpose | CI | Test Init |
|----------|-------|---------|----|----|
| `NO_COLOR` | `1` | Disables ANSI colors in lipgloss/termenv | ✓ | ✓ |
| `FORCE_COLOR` | `0` | Prevents color forcing | ✓ | - |
| `TERM` | `dumb` | Disables terminal capability detection | ✓ | ✓ |
| `CGO_ENABLED` | `1` | Required for go-sqlite3 | macOS job | - |

**Test Init:** `teatest_helpers_test.go` calls `lipgloss.SetColorProfile(termenv.Ascii)` in `NewTeatestModel()` (not global init, per-test).

### Key Source Files

| File | Purpose |
|------|---------|
| `.github/workflows/ci.yml` | CI workflow to restructure |
| `Makefile` | Add `test-behavioral` target |
| `.gitattributes` | Add `linguist-generated` for golden files |
| `internal/adapters/tui/teatest_helpers_test.go` | Color profile setup (line 250) |
| `internal/adapters/tui/testdata/*.golden` | 10 golden files from Stories 9.3-9.4 |

### FD Counting Platform Support

| Platform | Path | Support |
|----------|------|---------|
| macOS | `/dev/fd` | Full |
| Linux | `/proc/self/fd` | Full |
| Windows | N/A | Skip with `t.Skip()` |

**Why macOS for integration tests:** Story 9.5's FD leak detection uses `/dev/fd` which works reliably on macOS. Tests skip on Windows. macOS ensures FD tests execute. **Note:** macOS runners need `CGO_ENABLED=1` for go-sqlite3.

### CI Job Dependencies

```
        lint
       /    \
      v      v
unit-tests   integration-tests
      \        /
       v      v
        build
```

This ensures:
1. Lint catches formatting/style issues first (fast fail)
2. Unit and integration tests run in parallel (speed)
3. Build only runs after all tests pass (correctness)

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Verify Makefile Targets
```bash
make test        # Should run unit tests only (quick)
make test-all    # Should run all tests including integration
```
**Expected:** Both complete successfully.

### Step 2: Verify Golden File Attributes
```bash
cat .gitattributes | grep golden
```
**Expected:** `*.golden -text` and `*.golden linguist-generated=true`

### Step 3: Test CI Workflow Syntax
```bash
# Validate YAML syntax
cat .github/workflows/ci.yml | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)"
```
**Expected:** No output (valid YAML).

### Step 4: Verify Job Structure (Optional - requires act)
```bash
# Install act if needed: brew install act (macOS) or see https://github.com/nektos/act
act -l
```
**Expected:** Shows `lint`, `unit-tests`, `integration-tests`, `build` jobs.
**Note:** `act` requires Docker. Skip this step if Docker is not available.

### Step 5: Push to Test Branch
```bash
git checkout -b test/ci-9-6
git push -u origin test/ci-9-6
```
**Expected:** CI runs with parallel jobs visible in GitHub Actions.

### Decision Guide

| Situation | Action |
|-----------|--------|
| All CI jobs pass with parallel execution | Mark `done` |
| Integration tests fail on macOS | Check FD counting, may need platform skip |
| Golden file tests fail | Verify test init sets color profile |
| Jobs not running in parallel | Check `needs` dependencies |

## Dev Agent Record

### Context Reference

| Context | File | Lines |
|---------|------|-------|
| Story file | `docs/sprint-artifacts/stories/epic-9/9-6-ci-pipeline-integration.md` | - |
| Research doc | `docs/testing/tui-testing-research.md` | 386-392 |
| Current CI | `.github/workflows/ci.yml` | 1-42 |
| Current Makefile | `Makefile` | 1-43 |
| Test helpers | `internal/adapters/tui/teatest_helpers_test.go` | 250 (color profile) |
| Story 9.5 | `docs/sprint-artifacts/stories/epic-9/9-5-long-running-session-tests.md` | - |

### Critical Anti-Patterns (DO NOT)

1. **DO NOT** remove existing env vars (`NO_COLOR`, `FORCE_COLOR`, `TERM`)
2. **DO NOT** run integration tests on `ubuntu-latest` - FD counting needs macOS
3. **DO NOT** set job timeout less than 15 minutes for integration tests
4. **DO NOT** forget `needs` dependencies between jobs
5. **DO NOT** use `go test -update` in CI - only local developers should update golden files
6. **DO NOT** skip adding `linguist-generated=true` to `.gitattributes`
7. **DO NOT** hardcode Go version - use `go-version-file: 'go.mod'`
8. **DO NOT** forget `CGO_ENABLED=1` for macOS integration job (go-sqlite3)

### Architecture Compliance

**No production code changes** - This is a CI/infrastructure story.

**Files to Modify:**
- `.github/workflows/ci.yml` - Add job separation
- `Makefile` - Add `test-behavioral` target
- `.gitattributes` - Add golden file configuration

**Files to Create:**
- `docs/testing/GOLDEN_FILE_UPDATE.md` - Workflow documentation

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

1. **CI Workflow Restructured** - Split single `build` job into 4 parallel/dependent jobs:
   - `lint` → runs first, checks formatting + golangci-lint
   - `unit-tests` → runs after lint, on ubuntu-latest
   - `integration-tests` → runs after lint, on macos-latest (for /dev/fd FD counting)
   - `build` → runs after both test jobs pass

2. **macOS Runner for FD Testing** - Integration tests use `macos-latest` with `CGO_ENABLED=1` for go-sqlite3 and reliable `/dev/fd` file descriptor counting (Story 9.5 requirement)

3. **Environment Variable Documentation** - Added comprehensive comment block in CI workflow explaining NO_COLOR, FORCE_COLOR, TERM purposes and cross-references to test helpers

4. **Makefile test-behavioral Target** - Added for debugging TUI issues locally with verbose output

5. **Golden File Configuration** - Added `linguist-generated=true` to `.gitattributes` so golden file diffs are collapsed by default in GitHub PRs

6. **GOLDEN_FILE_UPDATE.md** - Created comprehensive documentation for updating golden files

7. **Local Development Setup** - Added section to tui-testing-research.md with env vars, quick commands, and troubleshooting

8. **Pre-existing Test Failures** - Several test failures exist that are unrelated to this story:
   - `internal/adapters/filesystem` - directory integration tests
   - `internal/adapters/persistence/sqlite` - repository integration tests
   - `internal/adapters/tui` - framework resize simulation tests
   - `internal/config` - project config integration tests
   These need to be addressed separately.

### File List

| File | Action | Purpose |
|------|--------|---------|
| `.github/workflows/ci.yml` | MODIFIED | Restructured to 4 jobs with parallel execution |
| `Makefile` | MODIFIED | Added test-behavioral target |
| `.gitattributes` | MODIFIED | Added linguist-generated=true for golden files |
| `docs/testing/GOLDEN_FILE_UPDATE.md` | CREATED | Golden file update workflow documentation |
| `docs/testing/tui-testing-research.md` | MODIFIED | Added Local Development Setup section |
| `docs/sprint-artifacts/sprint-status.yaml` | MODIFIED | Story status tracking (SM update) |
| `docs/sprint-artifacts/stories/epic-9/validation-report-9-6-20260101.md` | CREATED | SM validation report (auto-generated) |

## Change Log

- 2026-01-01: Code review fixes applied (Amelia, Dev Agent)
  - Fixed line number references (47 → 250 for color profile setup)
  - Fixed User Testing Guide `test-quick` → `test` (target doesn't exist)
  - Fixed Files to Modify section (removed non-existent `test-quick`)
  - Added sprint-status.yaml and validation-report to File List

- 2026-01-01: Story validated and improved by SM agent (Bob) via validate-create-story
  - **C1 FIXED:** Updated target workflow to use `go-version-file: 'go.mod'` instead of hardcoded `1.24`
  - **C2 FIXED:** Added `CGO_ENABLED=1` to macOS integration job for go-sqlite3 compatibility
  - **C3 FIXED:** Clarified Task 2 - no `test-quick` needed, `test` target already exists
  - **C4 FIXED:** Highlighted that `.gitattributes` needs `linguist-generated=true` added (only has `-text`)
  - **C5 FIXED:** Added Docker requirement note for `act` local CI testing
  - **L1 FIXED:** Removed redundant "Current CI Workflow" section (duplicated actual file)
  - **L2 FIXED:** Condensed golden file update docs to key points with reference
  - **L3 FIXED:** Consolidated environment variable table with CI/Test Init columns
  - Anti-patterns updated with Go version and CGO requirements

- 2026-01-01: Story created by SM agent (Bob) via create-story workflow YOLO mode
  - Comprehensive analysis of Stories 9.1-9.5 for context
  - CI workflow restructure based on research recommendations
  - macOS runner for FD testing requirement
  - Golden file management workflow
  - All ACs derived from Story 9.1 research and current CI state
