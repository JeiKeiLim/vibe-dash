# Story 9.5.5: Pipeline Summary Output

Status: done

**Priority: Low**

## Story

As a **developer using the vibe-dash CLI or CI pipeline**,
I want **a summary of test results and build status printed after pipeline completion**,
So that **I can quickly see the outcome without scrolling through verbose output**.

## Background

**Origin:** Epic 8 Retrospective → Action Item P1 (carried forward 4 times)

The process improvement P1 "Pipeline summary output after completion" has been carried forward through 4 retrospectives:
- Epic 8 → P1 created
- Epic 9 → P1 not addressed, carried forward
- Epic 9.5 → Formalized as Story 9.5-5

### Current State

- `make test` runs `go test ./...` - no summary, just raw output
- `make test-all` runs with `-tags=integration` - same issue
- `make build` shows `CGO_ENABLED=1 go build ...` then output binary path
- CI workflow (`.github/workflows/ci.yml`) runs lint → unit tests → integration tests → build, but each job outputs raw logs

### Desired State

After each major Makefile target completes, print a concise summary:

```
════════════════════════════════════════════════════════════════
 PIPELINE SUMMARY
════════════════════════════════════════════════════════════════
 ✓ Lint        PASS  (2.3s)
 ✓ Unit Tests  PASS  (4.7s, 156 tests)
 ✓ Build       PASS  (1.2s, bin/vibe)
════════════════════════════════════════════════════════════════
```

For CI, use GitHub Actions job summary feature (`$GITHUB_STEP_SUMMARY`).

## Acceptance Criteria

### AC1: Makefile Summary for `make test`
- Given `make test` is run
- When all tests pass
- Then print summary: `✓ Unit Tests  PASS  (Xs, N tests)`
- And when any test fails
- Then print summary: `✗ Unit Tests  FAIL  (N passed, M failed)`

### AC2: Makefile Summary for `make test-all`
- Given `make test-all` is run
- When integration tests pass
- Then print summary: `✓ Integration Tests  PASS  (Xs, N tests)`

### AC3: Makefile Summary for `make build`
- Given `make build` is run
- When build succeeds
- Then print summary: `✓ Build  PASS  (Xs, bin/vibe vX.Y.Z)`

### AC4: Makefile Summary for `make lint`
- Given `make lint` is run
- When linting passes
- Then print summary: `✓ Lint  PASS  (Xs)`

### AC5: CI Job Summary (GitHub Actions)
- Given CI workflow completes
- When all jobs pass
- Then GitHub job summary shows:
  - Overall status (PASS/FAIL)
  - Each job's status and duration
  - Link to any failed step

### AC6: Error Preservation
- Given any step fails
- When summary is printed
- Then original error output is NOT suppressed
- And error details remain visible above summary

### AC7: Non-Breaking Change
- Given existing scripts depend on Makefile output
- When summary is added
- Then exit codes remain unchanged (0 for success, non-0 for failure)
- And summary appears AFTER raw output

## Tasks / Subtasks

- [x] Task 1: Create summary helper script (AC: 1, 2, 3, 4, 6, 7)
  - [x] 1.1: Create `scripts/summary.sh` with format functions
  - [x] 1.2: Add `print_test_summary()` - parses `go test` output for counts
  - [x] 1.3: Add `print_build_summary()` - shows binary path and version
  - [x] 1.4: Add `print_lint_summary()` - shows pass/fail status
  - [x] 1.5: Ensure ANSI colors only when terminal (check `[ -t 1 ]`)

- [x] Task 2: Update Makefile targets (AC: 1, 2, 3, 4, 7)
  - [x] 2.1: Update `test` target to capture output and call summary
  - [x] 2.2: Update `test-all` target similarly
  - [x] 2.3: Update `build` target to print summary after build
  - [x] 2.4: Update `lint` target to print summary
  - [x] 2.5: Ensure exit codes are preserved through piping

- [x] Task 3: Add CI job summary (AC: 5)
  - [x] 3.1: Add `>> $GITHUB_STEP_SUMMARY` to each CI job
  - [x] 3.2: Create summary Markdown table in final build job
  - [x] 3.3: Include timing from `${{ steps.*.outputs.duration }}`

- [x] Task 4: Verify non-breaking (AC: 6, 7)
  - [x] 4.1: Run `make test && echo $?` - should still be 0
  - [x] 4.2: Run `make test` with failing test - should still be non-0
  - [x] 4.3: Verify CI pipeline still works

## Dev Notes

### Previous Story Learnings

**From Story 9.5-4 (Pre-existing Test Failures Cleanup):**
- Skip-by-default pattern (`GOLDEN_TESTS=1`) keeps CI clean
- Tests require `GOLDEN_TESTS=1` / `FRAMEWORK_TESTS=1` / `STRESS_TESTS=1` env vars

**From Story 9.6 (CI Pipeline Integration):**
- CI uses `macos-latest` for integration tests (FD counting)
- Environment variables set at workflow level: `NO_COLOR=1`, `FORCE_COLOR=0`, `TERM=dumb`
- Jobs run in order: lint → unit-tests → integration-tests → build

### Implementation Details

**CRITICAL: Makefile Shell Requirement**

Add at top of Makefile (required for `pipefail`):
```makefile
SHELL := /bin/bash
```

**scripts/summary.sh - Complete Implementation (All 3 Functions):**

```bash
#!/bin/bash
# Pipeline summary helper for vibe-dash
# Usage: source scripts/summary.sh; print_test_summary <exit_code> <duration> <output_file>

BOLD='\033[1m'
GREEN='\033[32m'
RED='\033[31m'
RESET='\033[0m'

# Disable colors if not terminal
if [ ! -t 1 ]; then
    BOLD='' GREEN='' RED='' RESET=''
fi

print_separator() {
    echo "════════════════════════════════════════════════════════════════"
}

# Parse test counts from go test -v output
# CRITICAL: Requires -v flag to get individual test results (--- PASS: lines)
# Without -v, only package-level results (ok/FAIL) are available
print_test_summary() {
    local exit_code=$1
    local duration=$2
    local output_file=$3

    # Count package-level results (works with both -v and non-v)
    local passed_pkgs=$(grep -c "^ok\s" "$output_file" 2>/dev/null || echo 0)
    local failed_pkgs=$(grep -c "^FAIL\s" "$output_file" 2>/dev/null || echo 0)

    # Count individual tests (only works with -v flag)
    local passed_tests=$(grep -c "^--- PASS:" "$output_file" 2>/dev/null || echo 0)
    local failed_tests=$(grep -c "^--- FAIL:" "$output_file" 2>/dev/null || echo 0)

    # Use individual test counts if available, otherwise show package counts
    local count_label=""
    if [ "$passed_tests" -gt 0 ] || [ "$failed_tests" -gt 0 ]; then
        count_label="${passed_tests} tests"
    else
        count_label="${passed_pkgs} packages"
    fi

    print_separator
    echo " PIPELINE SUMMARY"
    print_separator
    if [ "$exit_code" -eq 0 ]; then
        echo -e " ${GREEN}✓${RESET} Tests  ${GREEN}PASS${RESET}  (${duration}s, ${count_label})"
    else
        if [ "$failed_tests" -gt 0 ]; then
            echo -e " ${RED}✗${RESET} Tests  ${RED}FAIL${RESET}  (${passed_tests} passed, ${failed_tests} failed)"
        else
            echo -e " ${RED}✗${RESET} Tests  ${RED}FAIL${RESET}  (${passed_pkgs} ok, ${failed_pkgs} failed packages)"
        fi
    fi
    print_separator

    # Cleanup temp file
    rm -f "$output_file" 2>/dev/null
}

print_lint_summary() {
    local exit_code=$1
    local duration=$2

    print_separator
    echo " PIPELINE SUMMARY"
    print_separator
    if [ "$exit_code" -eq 0 ]; then
        echo -e " ${GREEN}✓${RESET} Lint   ${GREEN}PASS${RESET}  (${duration}s)"
    else
        echo -e " ${RED}✗${RESET} Lint   ${RED}FAIL${RESET}  (see errors above)"
    fi
    print_separator
}

print_build_summary() {
    local exit_code=$1
    local duration=$2
    local binary_path=$3
    local version=$4

    print_separator
    echo " PIPELINE SUMMARY"
    print_separator
    if [ "$exit_code" -eq 0 ]; then
        echo -e " ${GREEN}✓${RESET} Build  ${GREEN}PASS${RESET}  (${duration}s, ${binary_path} ${version})"
    else
        echo -e " ${RED}✗${RESET} Build  ${RED}FAIL${RESET}  (see errors above)"
    fi
    print_separator
}
```

**Makefile Target Updates (All 4 Targets):**

```makefile
# SHELL must be bash for pipefail support (add at top of Makefile)
SHELL := /bin/bash

test:
	@set -o pipefail; \
	start=$$(date +%s); \
	go test -v ./... 2>&1 | tee /tmp/vibe-test-output.txt; \
	exit_code=$$?; \
	end=$$(date +%s); \
	. scripts/summary.sh && print_test_summary $$exit_code $$((end-start)) /tmp/vibe-test-output.txt; \
	exit $$exit_code

test-all:
	@set -o pipefail; \
	start=$$(date +%s); \
	go test -v -tags=integration -timeout=10m ./... 2>&1 | tee /tmp/vibe-test-output.txt; \
	exit_code=$$?; \
	end=$$(date +%s); \
	. scripts/summary.sh && print_test_summary $$exit_code $$((end-start)) /tmp/vibe-test-output.txt; \
	exit $$exit_code

lint:
	@start=$$(date +%s); \
	$(shell go env GOPATH)/bin/golangci-lint run; \
	exit_code=$$?; \
	end=$$(date +%s); \
	. scripts/summary.sh && print_lint_summary $$exit_code $$((end-start)); \
	exit $$exit_code

build:
	@start=$$(date +%s); \
	CGO_ENABLED=1 go build $(LDFLAGS) -o bin/vibe ./cmd/vibe; \
	exit_code=$$?; \
	end=$$(date +%s); \
	. scripts/summary.sh && print_build_summary $$exit_code $$((end-start)) bin/vibe $(VERSION); \
	exit $$exit_code
```

**Note:** `-v` flag added to `go test` commands to enable individual test counting.

**CI Job Summary:** Add job summary step to build job in `.github/workflows/ci.yml` (see Story 9.6 for CI structure).

### Key Code Locations

| File | Lines | Function | Change Required |
|------|-------|----------|-----------------|
| `scripts/summary.sh` | N/A | (new file) | Create with `print_test_summary`, `print_lint_summary`, `print_build_summary` |
| `Makefile` | 1 | (top) | Add `SHELL := /bin/bash` for pipefail support |
| `Makefile` | 15-26 | `test`, `test-all`, `build`, `lint` | Add summary output after each target |
| `.github/workflows/ci.yml` | 80-93 | `build` job | Add `$GITHUB_STEP_SUMMARY` step |

### Boundaries & Anti-Patterns

| Boundary | Details |
|----------|---------|
| **In scope** | Makefile, CI workflow, new `scripts/summary.sh` |
| **Out of scope** | TUI changes, core logic, new features, go test output format |

| Don't | Do Instead |
|-------|------------|
| Suppress original output | Show summary AFTER raw output |
| Change exit codes | Preserve original exit code exactly |
| Use `/bin/sh` features | Specify `SHELL := /bin/bash` (pipefail requires bash) |
| Assume color support | Check `[ -t 1 ]` before ANSI codes |
| Run `go test` without `-v` | Use `-v` flag to get individual test counts |
| Leave temp files | Cleanup `/tmp/vibe-test-output.txt` after use |

### Testing Strategy

1. **Manual verification:**
   - Run `make test` with passing tests → verify summary shows PASS
   - Introduce failing test → verify summary shows FAIL
   - Run `make build` → verify summary shows version

2. **Exit code verification:**
   ```bash
   make test && echo "Exit: $?"   # Should be 0
   # Introduce failure
   make test || echo "Exit: $?"   # Should be non-0
   ```

3. **CI verification:**
   - Push branch with CI workflow changes
   - Check Actions tab for job summary rendering

### References

| Document | Section | Relevance |
|----------|---------|-----------|
| `docs/sprint-artifacts/retrospectives/epic-9-retro-2026-01-01.md` | P1 action item | Origin of this story |
| `docs/sprint-artifacts/stories/epic-9/9-6-ci-pipeline-integration.md` | Complete story | CI structure reference |
| `Makefile` | All | Current implementation |
| `.github/workflows/ci.yml` | All | Current CI structure |

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Basic Check
```bash
make test
```
- **Expected:** After test output, see `════════` summary block
- **Red flag:** No summary, or summary before test output

### Step 2: Verify Exit Codes
```bash
make test && echo "SUCCESS" || echo "FAILURE"
# Expected: "SUCCESS" (assuming all tests pass)
```

### Step 3: Check Build Summary
```bash
make build
```
- **Expected:** See `✓ Build  PASS  (Xs, bin/vibe vX.Y.Z)` summary
- **Red flag:** Missing version or path

### Decision Guide

| Situation | Action |
|-----------|--------|
| Summary appears after output, exit codes correct | Mark `done` |
| Summary suppresses original output | FAIL - fix piping |
| Exit code wrong after summary | FAIL - check pipefail |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-9.5/9-5-5-pipeline-summary-output.md`
- Previous story: `docs/sprint-artifacts/stories/epic-9.5/9-5-4-pre-existing-test-failures-cleanup.md`
- Project context: `docs/project-context.md`
- CI reference: `docs/sprint-artifacts/stories/epic-9/9-6-ci-pipeline-integration.md`

### Agent Model Used

Claude Opus 4.5

### Debug Log References

N/A - No debug logs needed for this story.

### Completion Notes List

1. Created `scripts/summary.sh` with three functions: `print_test_summary()`, `print_lint_summary()`, `print_build_summary()`
2. Terminal color detection via `[ -t 1 ]` check - colors disabled for non-TTY (CI)
3. Added `SHELL := /bin/bash` at top of Makefile for pipefail support
4. Updated `test`, `test-all`, `build`, `lint` targets with timing and summary output
5. Exit codes preserved using `exit_code` variable capture and explicit `exit $$exit_code`
6. CI workflow updated with job summaries for all 4 jobs using `$GITHUB_STEP_SUMMARY`
7. Pipeline summary table added to build job showing all job statuses
8. Verified: `make test` shows "1045 tests", `make build` shows version, `make lint` passes

### File List

| File | Change |
|------|--------|
| `scripts/summary.sh` | New file - summary helper functions |
| `Makefile` | Modified - add summary output to targets, PID-based temp files |
| `.github/workflows/ci.yml` | Modified - add job summary output |
| `docs/sprint-artifacts/sprint-status.yaml` | Modified - story status update |

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-01 | SM (Bob) | Initial story creation via *create-story workflow (YOLO mode) |
| 2026-01-01 | SM (Bob) | Validation improvements: (C1) Fixed test count parsing - now uses -v flag with fallback to package counts; (C2) Added all 3 summary functions; (E1) Added SHELL := /bin/bash requirement; (E2) Added temp file cleanup; (L1-L3) Consolidated sections, removed duplicate info |
| 2026-01-02 | Dev (Amelia) | Implementation complete - all 4 tasks done, all ACs verified |
| 2026-01-02 | Dev (Amelia) | Code review: (M1) N/A - execute bit tracked by git automatically; (M2) Note only - dot-sourcing fine with SHELL=bash; (M3) Fixed - temp file uses PID for uniqueness; (L1) Added sprint-status.yaml to File List; (L3) Removed unused BOLD variable |
