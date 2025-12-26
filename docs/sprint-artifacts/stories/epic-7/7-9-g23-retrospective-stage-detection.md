# Story 7.9: G23 - Retrospective Stage Detection

Status: done

## Story

As a **user viewing my BMAD project status**,
I want **the dashboard to show meaningful context when all epics are done but a retrospective is in-progress**,
So that **I see "Retrospective for Epic N in progress" instead of "Unable to determine stage"**.

## Problem Statement

Gap G23 was identified during Epic 4.6 retrospective (2025-12-22): When all epics are done and all stories are done, but a retrospective is in-progress, the stage detection returns "Unable to determine stage".

**Current Behavior:**
```yaml
development_status:
  epic-6: done
  epic-7: done
  epic-7-retrospective: in-progress  # Currently ignored
  7-8-feature: done
```
Result: "Unable to determine stage" (fallback case at line 355)

**Expected Behavior:**
- Detect retrospective-in-progress entries
- Show "Retrospective for Epic N in progress" when all epics done + retro active
- Fall back to "All epics complete" when no retrospective active

**Root Cause Location:**
- `internal/adapters/detectors/bmad/stage_parser.go:117-122` - Retrospectives are explicitly skipped
- `internal/adapters/detectors/bmad/stage_parser.go:237-240` - All-done check returns early before checking retros
- `internal/adapters/detectors/bmad/stage_parser.go:354-356` - Fallback returns "Unable to determine stage"

**Design Decision (from Epic 4.6.1 D4):**
"Ignore retrospectives - they don't block development." This is correct for ACTIVE development, but when everything is done, we should provide meaningful context.

## Acceptance Criteria

1. **AC1: Retrospective Status Detected**
   - Given development_status contains entries matching `*-retrospective`
   - When parsing sprint-status.yaml
   - Then retrospective entries are tracked (not just skipped)
   - And their epic number is extracted as string (e.g., "epic-7-retrospective" → "7", "epic-4-5-retrospective" → "4-5")

2. **AC2: All Epics Done + Retro In-Progress**
   - Given all epics have status "done"
   - And a retrospective has status "in-progress" or synonym
   - When determining stage
   - Then returns StageImplement (project is in maintenance mode)
   - And Confidence is Certain
   - And Reasoning is "Retrospective for Epic N in progress"

3. **AC3: All Epics Done + No Active Retro**
   - Given all epics have status "done"
   - And no retrospective is in-progress (all are "completed" or "optional")
   - When determining stage
   - Then returns existing behavior: StageImplement with "All epics complete - project done"

4. **AC4: Mixed Epic States - Retros Ignored**
   - Given some epics are in-progress or backlog
   - And any retrospectives exist
   - When determining stage
   - Then retrospectives are still ignored (current behavior preserved)
   - And active epic/story logic takes precedence

5. **AC5: Multiple In-Progress Retros**
   - Given multiple retrospectives are in-progress
   - When determining stage
   - Then uses the FIRST retrospective (by string sort of epicNum, e.g., "4-5" < "6" < "7")
   - And reasoning shows that epic's retrospective

6. **AC6: Retrospective Status Normalization**
   - Given retrospective status variations
   - When normalizing via existing `normalizeStatus()` then checking `isActiveStatus()`
   - Then active statuses after normalization: "in-progress", "wip", "started"
   - Note: "WIP" → normalizeStatus → "in-progress" (via synonym map) → isActiveStatus → true

7. **AC7: Existing G-Tests Continue Passing**
   - Given all existing tests (G1-G22, G24, Edge cases)
   - When running test suite
   - Then all tests pass (no regressions)

8. **AC8: New Tests for Retrospective Scenarios**
   - Given new G23 test cases
   - When running test suite
   - Then tests verify all retrospective detection scenarios

9. **AC9: Deferred Epic Retrospectives Ignored**
   - Given an epic is deferred (per G24 logic)
   - And a retrospective exists for that deferred epic
   - When determining stage
   - Then the retrospective is ignored (deferred epics don't have active retrospectives)

## Tasks / Subtasks

- [x] Task 1: Add retrospective tracking during first pass (AC: 1)
  - [x] 1.1: Add `retroKeyRegex` at line ~24 (after epicKeyRegex):
    ```go
    var retroKeyRegex = regexp.MustCompile(`^epic-(\d+(?:-\d+)?)-retrospective$`)
    ```
  - [x] 1.2: Add `isActiveStatus()` helper at line ~65 (after isDeferred):
    ```go
    // isActiveStatus checks if a normalized status indicates an active/in-progress state.
    // Used for retrospective detection when all epics are done.
    func isActiveStatus(normalizedStatus string) bool {
        active := map[string]bool{
            "in-progress": true,  // Most common - normalizeStatus maps WIP/wip/in_progress here
            "started":     true,  // Direct usage
        }
        return active[normalizedStatus]
    }
    ```
  - [x] 1.3: Add `retroInfo` struct at line ~103 (near epicInfo):
    ```go
    type retroInfo struct {
        key     string // e.g., "epic-7-retrospective"
        epicNum string // e.g., "7" or "4-5" for sub-epics
        status  string // normalized status
    }
    ```
  - [x] 1.4: Modify first pass at line ~117 to collect retrospectives instead of just skipping. Add `var retrospectives []retroInfo` before the loop and track deferred epic retrospectives to skip them.

- [x] Task 2: Add retrospective check after all-done detection (AC: 2, 3, 5, 9)
  - [x] 2.1: After epic collection, sort retrospectives by epicNum (string sort, before all-done check):
    ```go
    sort.Slice(retrospectives, func(i, j int) bool {
        return retrospectives[i].epicNum < retrospectives[j].epicNum
    })
    ```
  - [x] 2.2: Modify the all-done block at line ~237-240 to check for active retrospectives BEFORE returning "All epics complete":
    ```go
    // All epics done - G23: Check for in-progress retrospective
    if len(epics) > 0 && doneCount == len(epics) {
        for _, retro := range retrospectives {
            if isActiveStatus(retro.status) {
                return domain.StageImplement, domain.ConfidenceCertain,
                    appendWarnings("Retrospective for Epic " + retro.epicNum + " in progress")
            }
        }
        return domain.StageImplement, domain.ConfidenceCertain,
            appendWarnings("All epics complete - project done")
    }
    ```

- [x] Task 3: Add G23 test cases (AC: 6, 7, 8, 9)
  - [x] 3.1: Add TestIsActiveStatus unit tests (after TestIsDeferred):
    - "in-progress" → true
    - "started" → true
    - "done" → false
    - "completed" → false (Note: normalizeStatus maps to "done")
    - "optional" → false
    - "backlog" → false
  - [x] 3.2: Add TestIsActiveStatusWithNormalization (verify combined flow like TestIsDeferredWithNormalization):
    - "WIP" → normalizeStatus → "in-progress" → isActiveStatus → true
    - "In Progress" → normalizeStatus → "in-progress" → isActiveStatus → true
    - "in_progress" → normalizeStatus → "in-progress" → isActiveStatus → true
    - "DONE" → normalizeStatus → "done" → isActiveStatus → false
  - [x] 3.3: Add G23 test cases to TestDetermineStageFromStatus (after G24 tests):
    - "G23: all epics done with retro in-progress" → StageImplement, Certain, "Retrospective for Epic 7 in progress"
    - "G23: all epics done with retro completed" → StageImplement, Certain, "All epics complete - project done"
    - "G23: all epics done with retro optional" → StageImplement, Certain, "All epics complete - project done"
    - "G23: epic in-progress - retro ignored" → current story logic takes precedence
    - "G23: multiple retros - uses first by epic number" → Epic 6 reported (6 < 7 string sort)
    - "G23: retro status normalization - WIP" → "Retrospective for Epic 7 in progress"
    - "G23: sub-epic retrospective format" (epic-4-5-retrospective) → "Retrospective for Epic 4-5 in progress"
    - "G23: deferred epic retrospective ignored" → deferred epic retro not reported
  - [x] 3.4: Run full test suite: `go test ./internal/adapters/detectors/bmad/... -v`
  - [x] 3.5: Verify no regressions in existing tests

- [x] Task 4: Dogfooding verification (AC: 2)
  - [x] 4.1: Build: `make build`
  - [x] 4.2: Modify sprint-status.yaml to simulate all-done + retro in-progress:
    ```yaml
    epic-7: done
    epic-7-retrospective: in-progress
    ```
  - [x] 4.3: Run: `./bin/vibe status vibe-dash`
  - [x] 4.4: Verify: Shows "Retrospective for Epic 7 in progress"
  - [x] 4.5: Revert sprint-status.yaml changes

## Dev Notes

### Implementation Pattern

Follow the existing pattern from Story 7-8 (G24 deferred detection):
1. Add new regex for matching keys (`retroKeyRegex`)
2. Add helper function (`isActiveStatus`) - similar to `isDeferred`
3. Track entries during first pass instead of just skipping
4. Use tracked data at appropriate decision point (inside all-done block)

### Why `isActiveStatus` Only Checks "in-progress" and "started"

The existing `normalizeStatus()` function already maps status synonyms:
- "wip" → "in-progress" (via synonyms map at line 46)
- "WIP" → "in-progress" (lowercase first, then synonyms)
- "in_progress" → "in-progress" (separator normalization)

So `isActiveStatus()` only needs to check the canonical normalized forms. The status "wip" raw would need to NOT be normalized first to match, but we always call `normalizeStatus()` before `isActiveStatus()`.

### Key Insertion Points

| Location | Line | Change |
|----------|------|--------|
| `stage_parser.go` | ~24 | Add `retroKeyRegex` after `epicKeyRegex` |
| `stage_parser.go` | ~65 | Add `isActiveStatus()` after `isDeferred()` |
| `stage_parser.go` | ~103 | Add `retroInfo` struct near `epicInfo` |
| `stage_parser.go` | ~117-122 | Modify to collect retrospectives (track deferred epic retros too) |
| `stage_parser.go` | ~237-240 | Add retrospective check inside all-done block |
| `stage_parser_test.go` | After TestIsDeferred | Add TestIsActiveStatus, TestIsActiveStatusWithNormalization |
| `stage_parser_test.go` | After G24 tests | Add G23 test cases |

### Why This Approach

1. **Minimal Change**: Only modify the "all epics done" code path
2. **No Regression Risk**: Retrospectives are still ignored during active development (AC4)
3. **Deterministic**: String sort by epicNum ensures consistent behavior ("4-5" < "6" < "7")
4. **Pattern Consistency**: Follows same patterns as G24 (deferred handling)
5. **Deferred Integration**: Respects G24 by skipping retrospectives for deferred epics

### Testing Commands

```bash
# Run specific test file
go test ./internal/adapters/detectors/bmad/... -v

# Run only G23 tests (after adding)
go test ./internal/adapters/detectors/bmad/... -run "G23" -v

# Run isActiveStatus tests
go test ./internal/adapters/detectors/bmad/... -run "TestIsActiveStatus" -v

# Full verification
make test
make lint
```

### Project Structure Notes

- Follows hexagonal architecture: detector in `internal/adapters/detectors/bmad/`
- Core domain types in `internal/core/domain/` (Stage, Confidence constants)
- Tests co-located with source (Go convention)

### References

- [Source: docs/sprint-artifacts/retrospectives/epic-4.6-retro-2025-12-22.md] - G23 gap identification
- [Source: Story 7-8] - Implementation pattern for G24 deferred handling
- [Source: internal/adapters/detectors/bmad/stage_parser.go:117-122] - Current retrospective skip logic
- [Source: internal/adapters/detectors/bmad/stage_parser.go:237-240] - All-done check logic
- [Source: internal/adapters/detectors/bmad/stage_parser.go:41-56] - normalizeStatus() synonym handling

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Build and Prepare
```bash
make build
```

### Step 2: Simulate All-Done + Retro In-Progress
Temporarily modify `docs/sprint-artifacts/sprint-status.yaml`:
```yaml
# Change epic-7 to done (save current status first)
epic-7: done
# Add or change retrospective to in-progress
epic-7-retrospective: in-progress
```

### Step 3: Verify Output
```bash
./bin/vibe status vibe-dash
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Stage shown | "Implement" | "Unknown" |
| Reasoning | "Retrospective for Epic 7 in progress" | "Unable to determine stage" |
| No errors | Clean output | Panic or error |

### Step 4: Revert Changes
Restore sprint-status.yaml to original state.

### Decision Guide
| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Shows "Unable to determine stage" | Do NOT approve, retrospective detection failed |
| Wrong epic number in reasoning | Do NOT approve, document issue |

## Dependencies

- Story 7-8 completed (G24 deferred handling pattern exists)
- Story 4.6.3 completed (normalizeStatus() function exists)
- No blocking dependencies

## Dev Agent Record

### Context Reference

N/A - Story fully specified with implementation guidance

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

1. **Implementation completed** - All 4 tasks done following the story specification exactly
2. **Tests passing** - 8 new G23 tests added, all 42 tests in `stage_parser_test.go` pass
3. **Dogfooding verified** - `./bin/vibe status vibe-dash` correctly showed "Retrospective for Epic 7 in progress" when simulated
4. **Key implementation detail** - Retrospectives are collected in a second loop (after epics) to ensure deferred epics are already known before filtering retrospectives

### Code Review Record

**Review Date:** 2025-12-26
**Reviewer:** Amelia (Dev Agent)
**Result:** PASS with fixes applied

**Issues Found:**
- M1: `isActiveStatus` map recreated on every call → FIXED: Moved to package-level `activeStatuses` var
- M2: Doc comment style inconsistency → SKIPPED (acceptable)
- M3: Missing "started" status test case → FIXED: Added "G23: retro status started" test
- M4: Story line numbers outdated → SKIPPED (code is correct)
- L1/L2: Style/doc observations → SKIPPED

**Fixes Applied:**
1. `stage_parser.go:70-80` - Moved active map to package-level `activeStatuses` var for efficiency
2. `stage_parser_test.go:831-843` - Added "G23: retro status started" test case

**All Tests Passing:** 43 tests in stage_parser_test.go (42 original + 1 new)

### File List

| File | Change |
|------|--------|
| `internal/adapters/detectors/bmad/stage_parser.go` | Add `retroKeyRegex` (line 26-28), `isActiveStatus()` helper (line 70-78), `retroInfo` struct (line 127-132); collect retrospectives in second loop (line 163-180); sort and check retrospectives in all-done block (line 261-291) |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | Add `TestIsActiveStatus` (line 128-154), `TestIsActiveStatusWithNormalization` (line 156-191), 8 G23 test cases in `TestDetermineStageFromStatus` (line 731-844) |
