# Story 7.8: G24 - Deferred Epic Stage Detection

Status: done

## Story

As a **user viewing my BMAD project status**,
I want **epics marked as "deferred" to be skipped in stage detection**,
So that **the dashboard shows correct stage information for active epics instead of "Unknown status"**.

## Problem Statement

Epic 5 (Hibernation) was deferred to post-MVP with status "deferred-post-mvp" in sprint-status.yaml. The current stage detection logic in `determineStageFromStatus()` doesn't handle this status:

1. Epics with "deferred" status are counted in the total but fall through the switch statement
2. This causes incorrect stage detection results
3. Dashboard shows "Unknown status" or incorrect reasoning

**Root Cause Location:**
- `internal/adapters/detectors/bmad/stage_parser.go:181-191` - Epic status switch statement
- `internal/adapters/detectors/bmad/stage_parser.go:28-57` - `normalizeStatus()` function

**Current behavior:**
```yaml
development_status:
  epic-5: deferred-post-mvp  # Falls through switch - not counted as backlog, in-progress, or done
  epic-6: done
  epic-7: in-progress
  7-8-feature: backlog
```
Result: Stage detection may fail or return incorrect stage

**Expected behavior:**
- Deferred epics should be completely skipped (not counted in totals)
- Stage detection should focus only on active epics (backlog, in-progress, done)
- Stories belonging to deferred epics should NOT trigger orphan warnings

## Acceptance Criteria

1. **AC1: Deferred Epic Status Detected**
   - Given epic status contains "deferred" (any variation)
   - When `isDeferred()` helper is called after normalization
   - Then returns true for any deferred status
   - Variations: deferred, deferred-post-mvp, deferred-to-v2, post-mvp

2. **AC2: Deferred Epics Skipped in Counting**
   - Given sprint-status.yaml has epics with "deferred" status
   - When counting epics for stage determination
   - Then deferred epics are completely excluded from epics map
   - And their stories are never processed (no orphan warnings)

3. **AC3: Stage Detection Works with Mix of Deferred and Active Epics**
   - Given some epics are deferred and some are active
   - When determining stage from sprint-status
   - Then only active epics (backlog, in-progress, done) are considered
   - And deferred epics don't affect the result

4. **AC4: All Epics Deferred Scenario**
   - Given ALL epics in sprint-status are deferred
   - When determining stage
   - Then returns StageUnknown with ConfidenceUncertain
   - And reasoning is "All epics deferred - no active development"

5. **AC5: Existing G-Tests Continue Passing**
   - Given all existing tests (G1, G2, G3, G7, G8, G14, G15, G17, G19, G22, Edge15)
   - When running test suite
   - Then all tests pass (no regressions)

6. **AC6: New Tests for Deferred Scenarios**
   - Given new G24 test cases for deferred epic handling
   - When running test suite
   - Then tests verify all deferred variations work correctly

## Tasks / Subtasks

- [x] Task 1: Add isDeferred() helper function (AC: 1, 2)
  - [x] 1.1: Create helper at line ~58 (after normalizeStatus):
    ```go
    func isDeferred(normalizedStatus string) bool {
        return strings.Contains(normalizedStatus, "deferred") ||
               strings.HasPrefix(normalizedStatus, "post-mvp")
    }
    ```
  - [x] 1.2: Add unit tests in TestIsDeferred (after TestNormalizeStatus):
    - "deferred" → true
    - "deferred-post-mvp" → true
    - "deferred-to-v2" → true
    - "post-mvp" → true
    - "in-progress" → false
    - "done" → false
    - "backlog" → false

- [x] Task 2: Update determineStageFromStatus() epic collection (AC: 2, 3, 4)
  - [x] 2.1: Modify first pass at line 110-123 to skip deferred epics:
    ```go
    // First pass: collect epics (skip deferred)
    for key, value := range status.DevelopmentStatus {
        if strings.HasSuffix(key, "-retrospective") {
            continue
        }

        if epicKeyRegex.MatchString(key) {
            normalized := normalizeStatus(value)

            // G24: Skip deferred epics entirely
            if isDeferred(normalized) {
                continue
            }

            epics[key] = &epicInfo{
                key:    key,
                status: normalized,
            }
            epicOrder = append(epicOrder, key)
        }
    }
    ```
  - [x] 2.2: Add check for "all epics deferred" after line 126 (after sort):
    ```go
    // G24: Check if all epics are deferred (no active epics found)
    if len(epics) == 0 {
        return domain.StageUnknown, domain.ConfidenceUncertain,
            "All epics deferred - no active development"
    }
    ```
  - [x] 2.3: Track deferred epics to skip their stories (prevents orphan warnings)

- [x] Task 3: Add G24 test cases (AC: 5, 6)
  - [x] 3.1: Add to TestDetermineStageFromStatus (after line 513):
    ```go
    // G24: Deferred epic handling
    {
        name: "G24: single deferred epic with active epics",
        status: &SprintStatus{
            DevelopmentStatus: map[string]string{
                "epic-1":      "deferred-post-mvp",
                "1-1-feature": "deferred-post-mvp",
                "epic-2":      "in-progress",
                "2-1-feature": "in-progress",
            },
        },
        wantStage:      domain.StageImplement,
        wantConfidence: domain.ConfidenceCertain,
        wantReasoning:  "Story 2.1 being implemented",
    },
    {
        name: "G24: all epics deferred",
        status: &SprintStatus{
            DevelopmentStatus: map[string]string{
                "epic-1":      "deferred-post-mvp",
                "epic-2":      "deferred",
            },
        },
        wantStage:      domain.StageUnknown,
        wantConfidence: domain.ConfidenceUncertain,
        wantReasoning:  "All epics deferred - no active development",
    },
    {
        name: "G24: deferred-to-v2 variation",
        status: &SprintStatus{
            DevelopmentStatus: map[string]string{
                "epic-1":      "deferred-to-v2",
                "epic-2":      "in-progress",
                "2-1-feature": "ready-for-dev",
            },
        },
        wantStage:      domain.StagePlan,
        wantConfidence: domain.ConfidenceCertain,
        wantReasoning:  "Story 2.1 ready for development",
    },
    {
        name: "G24: post-mvp variation",
        status: &SprintStatus{
            DevelopmentStatus: map[string]string{
                "epic-1":      "post-mvp",
                "epic-2":      "backlog",
            },
        },
        wantStage:      domain.StageSpecify,
        wantConfidence: domain.ConfidenceCertain,
        wantReasoning:  "No epics in progress - planning phase",
    },
    {
        name: "G24: deferred epic stories not flagged as orphans",
        status: &SprintStatus{
            DevelopmentStatus: map[string]string{
                "epic-1":      "deferred-post-mvp",
                "1-1-feature": "backlog",  // Story for deferred epic
                "1-2-feature": "drafted",  // Another story for deferred epic
                "epic-2":      "in-progress",
                "2-1-feature": "in-progress",
            },
        },
        wantStage:      domain.StageImplement,
        wantConfidence: domain.ConfidenceCertain,
        wantReasoning:  "Story 2.1 being implemented",  // No orphan warnings
    },
    ```
  - [x] 3.2: Run full test suite: `go test ./internal/adapters/detectors/bmad/... -v`
  - [x] 3.3: Verify no regressions in existing G-tests

- [x] Task 4: Dogfooding verification (AC: 3)
  - [x] 4.1: Build: `make build`
  - [x] 4.2: Run: `./bin/vibe list`
  - [x] 4.3: Verify: Epic 7 shows "Plan" stage correctly, no "Unknown status" from Epic 5

## Dev Notes

### Implementation Pattern

Use `strings.Contains` check after normalization (more resilient than expanding synonyms):

```go
// Add after normalizeStatus() at line ~58
func isDeferred(normalizedStatus string) bool {
    return strings.Contains(normalizedStatus, "deferred") ||
           strings.HasPrefix(normalizedStatus, "post-mvp")
}
```

This handles all variations:
- `deferred` → contains "deferred" ✓
- `deferred-post-mvp` → contains "deferred" ✓
- `deferred-to-v2` → contains "deferred" ✓
- `post-mvp` → has prefix "post-mvp" ✓

### Key Insertion Points

| Location | Line | Change |
|----------|------|--------|
| `stage_parser.go` | ~58 | Add `isDeferred()` helper after `normalizeStatus()` |
| `stage_parser.go` | 116 | Add `if isDeferred(normalized) { continue }` |
| `stage_parser.go` | 127 | Add empty epics check before counting |
| `stage_parser_test.go` | ~57 | Add `TestIsDeferred` after `TestNormalizeStatus` |
| `stage_parser_test.go` | ~513 | Add G24 test cases after Edge15 tests |

### Why Stories Don't Need Orphan Warning Suppression

The story association logic (line 132-163) only looks for epics IN the `epics` map:
```go
if epic, ok := epics[epicKey]; ok {
    // Associate story with epic
}
```

Since deferred epics are skipped in the first pass, they never enter the map. Their stories simply don't find a matching epic and are silently ignored (no warning added). This is the desired behavior - we don't want G14 orphan warnings for intentionally deferred work.

### Files to Modify

| File | Change |
|------|--------|
| `internal/adapters/detectors/bmad/stage_parser.go` | Add `isDeferred()`, update epic collection |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | Add `TestIsDeferred`, add G24 test cases |

### Testing Commands

```bash
# Run specific test file
go test ./internal/adapters/detectors/bmad/... -v

# Run only G24 tests (after adding)
go test ./internal/adapters/detectors/bmad/... -run "G24" -v

# Run isDeferred tests
go test ./internal/adapters/detectors/bmad/... -run "TestIsDeferred" -v

# Full verification
make test
make lint
```

## User Testing Guide

**Time needed:** 2 minutes

### Step 1: Build and Run
```bash
make build
./bin/vibe list
```

### Step 2: Verify Output
| Check | Expected | Red Flag |
|-------|----------|----------|
| Epic 7 stage | Shows "in-progress" or current story status | Shows "Unknown status" |
| Reasoning | References active story (e.g., "Story 7.x...") | References Epic 5 or "deferred" |
| No errors | Clean output | Any panic or error |

### Decision Guide
| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Shows "Unknown status" | Do NOT approve, deferred handling failed |
| Wrong story referenced | Do NOT approve, document issue |

## Dependencies

- Story 4.6.3 completed (normalizeStatus() function exists) ✓
- No blocking dependencies

## Dev Agent Record

### Context Reference

N/A - Story was fully specified with implementation guidance

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

1. Added `isDeferred()` helper function at `stage_parser.go:59-64` - checks for "deferred" substring or "post-mvp" prefix
2. Modified epic collection loop to track deferred epics in a separate map and skip them from active processing
3. Added early return for "all epics deferred" scenario
4. Updated story association logic to skip stories for deferred epics (prevents orphan warnings)
5. Added 5 G24 test cases covering all acceptance criteria
6. All existing G-tests pass (G1, G2, G3, G7, G8, G14, G15, G17, G19, G22, Edge15) - no regressions
7. Dogfooding verification: `./bin/vibe status vibe-dash` shows "Plan" stage with "Certain" confidence, no "Unknown status" errors

### Code Review Record

**Reviewer:** Dev Agent (Adversarial Code Review)
**Date:** 2025-12-26
**Model:** Claude Opus 4.5

**Issues Found:** 0 High, 3 Medium, 2 Low

**Fixes Applied:**
- M1: Added edge case test for story with in-progress status belonging to deferred epic
- M2: Added `TestIsDeferredWithNormalization` to verify uppercase/space/underscore variations work
- M3: Updated test comments to be more explicit about what's being verified
- L1: Enhanced comment at `stage_parser.go:127` to explain WHY we track deferred epics

**Final Test Results:** All 6 G24 tests pass, all existing G-tests pass, linter clean

### File List

| File | Change |
|------|--------|
| `internal/adapters/detectors/bmad/stage_parser.go` | Added `isDeferred()` helper, updated epic collection to skip deferred, added deferred epic tracking for story association, improved comment clarity |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | Added `TestIsDeferred` test suite, added `TestIsDeferredWithNormalization`, added 6 G24 test cases to `TestDetermineStageFromStatus` |
