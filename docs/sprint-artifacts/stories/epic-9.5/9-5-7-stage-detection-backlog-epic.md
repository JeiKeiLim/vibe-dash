# Story 9.5.7: Stage Detection for Backlog Epics

Status: done

**Priority: Medium**

## Story

As a **developer viewing a BMAD project dashboard**,
I want **the stage display to show the next epic and story when all completed epics are done but the next epic is still in backlog**,
So that **I can see where to focus next instead of seeing "Unknown" or an unhelpful stage**.

## User-Visible Changes

- **Changed:** When all active epics are done and the next epic is in backlog, dashboard now shows the next epic/story info (e.g., "E9.5 S9.5-1 backlog") instead of "Unknown" or just "Done"

## Background

**Origin:** Epic 9 Retrospective (2026-01-01) → New Issue "Stage Detection for Backlog Epics"

### Problem Statement

Currently, when:
1. All completed epics (e.g., Epic 9) have status `done`
2. The next epic (e.g., Epic 9.5) has status `backlog` or `in-progress`
3. No stories in the next epic have started yet

The dashboard shows "Unknown" stage or misleading information because:
- The `determineStageFromStatus()` function only looks for `firstInProgressEpic`
- If the first non-done epic is in `backlog` status (not `in-progress`), it's counted but not selected for story analysis
- The function falls through to "Unknown" because no in-progress epic was found

### Actual Behavior

```
Epic 9: done
Epic 9.5: in-progress (or backlog)
  9-5-7-stage-detection-backlog-epic: backlog

Result: Falls through, shows "Unknown" or "All epics complete"
```

### Expected Behavior

When an epic is `in-progress` or `backlog` (not `done`, not `deferred`):
- Should analyze its stories
- Should show the first actionable story (by priority: review > in-progress > ready-for-dev > drafted > backlog)
- Dashboard should display: "E9.5 S9.5-7 backlog" (or whichever is the first story)

### Root Cause

In `internal/adapters/detectors/bmad/stage_parser.go:252-264`:

```go
for _, epicKey := range epicOrder {
    epic := epics[epicKey]
    switch epic.status {
    case "backlog":
        backlogCount++
    case "in-progress", "contexted":
        if firstInProgressEpic == nil {
            firstInProgressEpic = epic
        }
    case "done":
        doneCount++
    }
}
```

The issue: `backlog` status only increments counter, doesn't capture the epic for story analysis like `in-progress` does.

Then at line 314:
```go
if len(epics) > 0 && backlogCount == len(epics) {
    return domain.StageSpecify, domain.ConfidenceCertain, appendWarnings("No epics in progress - planning phase")
}
```

This catches "all backlog" but not "some done, some backlog, none in-progress".

## Acceptance Criteria

### AC1: Backlog Epic Story Analysis
- Given an epic has status `backlog`
- And there are stories in that epic
- When determining stage from sprint status
- Then analyze the epic's stories using priority selection (same as in-progress)
- And return appropriate stage based on story status

### AC2: Mixed Done/Backlog Epic Handling
- Given some epics are `done` and some are `backlog`
- And no epics are `in-progress`
- When determining stage from sprint status
- Then select the first backlog epic (by natural order) for story analysis
- And return stage info based on that epic's stories

### AC3: First Backlog Epic Selection
- Given multiple epics with `backlog` status
- When selecting which epic to analyze
- Then select the first one by natural ordering (epic-9 < epic-9-5 < epic-10)
- And analyze its stories

### AC4: Backlog Epic Without Stories
- Given an epic with `backlog` status
- And the epic has no stories yet
- When determining stage from sprint status
- Then return `StagePlan` with reasoning "Epic X.Y in backlog, needs story planning"

### AC5: Priority Over All-Done Check
- Given some epics are `done` and one is `backlog`
- When the all-done check runs (line 290)
- Then it should NOT trigger because not all epics are done
- And backlog epic should be processed instead

### AC6: Existing G-Tests Pass
- Given all existing stage detection tests
- When running the test suite
- Then all tests continue to pass (no regressions)

### AC7: New Tests for Backlog Epic Scenarios
- Given new test cases for backlog epic handling
- When running the test suite
- Then tests verify:
  - Mixed done/backlog detection works
  - Backlog epic story analysis works
  - Backlog epic without stories returns correct stage

### AC8: Stage Format Parsing
- Given DetectionReasoning contains backlog epic info
- When `stageformat.FormatStageInfo()` parses it
- Then it should display correctly (e.g., "E9.5 S9.5-7 backlog")

## Tasks / Subtasks

- [x] Task 1: Capture first backlog epic for story analysis (AC: 1, 2, 3)
  - [x] 1.1: Modify epic status switch at line 252-264 to capture `firstBacklogEpic`:
    ```go
    var firstBacklogEpic *epicInfo
    // ...
    case "backlog":
        backlogCount++
        if firstBacklogEpic == nil {
            firstBacklogEpic = epic
        }
    ```
  - [x] 1.2: Ensure natural ordering is already applied (epicOrder is sorted)

- [x] Task 2: Add backlog epic story analysis fallback (AC: 1, 2, 4, 5)
  - [x] 2.1: After line 297 (all-done check), add check for firstBacklogEpic:
    ```go
    // Check for backlog epic when no in-progress epic exists
    if firstInProgressEpic == nil && firstBacklogEpic != nil {
        // Analyze backlog epic's stories using same priority logic
        // (similar to in-progress epic story analysis at line 319+)
    }
    ```
  - [x] 2.2: If backlog epic has stories, apply same priority selection
  - [x] 2.3: If backlog epic has no stories, return:
    - `StagePlan, ConfidenceCertain, "Epic X.Y in backlog, needs story planning"`

- [x] Task 3: Refactor story analysis to helper function (AC: 1, 2)
  - [x] 3.1: Extract lines 320-410 (in-progress story analysis) to helper:
    ```go
    func analyzeEpicStories(epic *epicInfo, appendWarnings func(string) string) (domain.Stage, domain.Confidence, string, bool)
    ```
  - [x] 3.2: Return `bool` to indicate if analysis produced a result
  - [x] 3.3: Use helper for both in-progress and backlog epic analysis

- [x] Task 4: Write tests for backlog epic scenarios (AC: 6, 7)
  - [x] 4.1: Add test `TestDetermineStageFromStatus_G25_BacklogEpicWithStories`:
    - Scenario: All done except one backlog epic with backlog stories
    - Expected: Returns story info from backlog epic
  - [x] 4.2: Add test `TestDetermineStageFromStatus_G26_BacklogEpicNoStories`:
    - Scenario: Backlog epic with no stories
    - Expected: Returns "Epic X.Y in backlog, needs story planning"
  - [x] 4.3: Add test `TestDetermineStageFromStatus_G27_MixedDoneBacklog`:
    - Scenario: Some done, some backlog, none in-progress
    - Expected: Selects first backlog epic for analysis
  - [x] 4.4: Run all existing tests to verify no regressions

- [x] Task 5: Verify stage format parsing (AC: 8)
  - [x] 5.1: Run existing stageformat tests to ensure "backlog" status is already handled
  - [x] 5.2: Add test for epic-level backlog reasoning (added 3 tests)

## Dev Notes

### Architecture Pattern

This change follows the existing G24 (deferred epic) pattern from Story 7-8:
- Add a new epic info capture variable (`firstBacklogEpic` like `firstInProgressEpic`)
- Insert check at appropriate priority in the stage determination flow
- Maintain existing test patterns (G1-G24)

### File Locations

| File | Change |
|------|--------|
| `internal/adapters/detectors/bmad/stage_parser.go` | Main logic: capture backlog epic, add fallback analysis |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | Add G25, G26, G27 test cases |
| `internal/shared/stageformat/stageformat_test.go` | Verify backlog status parsing (if needed) |

### Implementation Priority

1. **Option A (Minimal):** Just capture `firstBacklogEpic` and add one fallback check
   - Pros: Smaller change, less risk
   - Cons: Some code duplication with in-progress analysis

2. **Option B (Refactored):** Extract story analysis to helper, use for both cases
   - Pros: DRY, cleaner code
   - Cons: Larger change, more testing needed

**Implemented:** Option B - The `analyzeEpicStories()` helper is used by both in-progress and backlog epic paths. Code review M1 fix refactored the in-progress path to use the shared helper, eliminating ~70 lines of duplicated code.

### Detection Reasoning Format

The reasoning string format should match existing patterns:
- Story-level: `"Story X.Y.Z status"` - parsed by `parseStoryReasoning()`
- Epic-level: `"Epic X.Y status"` - parsed by `parseEpicReasoning()`

For backlog epic without stories:
```go
return domain.StagePlan, domain.ConfidenceCertain,
    appendWarnings(formatEpicKey(firstBacklogEpic.key) + " in backlog, needs story planning")
```

This would generate: `"Epic 9.5 in backlog, needs story planning"`

The `stageformat.go` should already handle this via `parseEpicReasoning()` with a new abbreviation for "backlog" status.

### Previous Story Learnings

**From Story 9.5-6 (User-Visible Changes):**
- Documentation-only changes are low-risk
- Template guidance helps future story creation

**From Story 7-8 (G24 Deferred Epic):**
- Pattern for adding new epic status handling
- Use `isDeferred()` helper pattern for new checks
- Test naming: G24, G25, etc.

### References

| Document | Section |
|----------|---------|
| `docs/sprint-artifacts/retrospectives/epic-9-retro-2026-01-01.md` | "Stage Detection for Backlog Epics" issue |
| `docs/sprint-artifacts/stories/epic-7/7-8-g24-deferred-epic-stage-detection.md` | Pattern for epic status handling |
| `internal/adapters/detectors/bmad/stage_parser.go` | Lines 252-264 (epic switch), 290-298 (all-done check), 318-410 (story analysis) |
| `internal/shared/stageformat/stageformat.go` | Lines 96-131 (story parsing), 133-161 (epic parsing) |

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Simulate Backlog Epic Scenario

Temporarily modify `docs/sprint-artifacts/sprint-status.yaml`:
```yaml
epic-9-5: backlog  # Change from in-progress to backlog
```

### Step 2: Run Dashboard

```bash
./bin/vibe
```

### Step 3: Check Stage Display

| Scenario | Expected | Red Flag |
|----------|----------|----------|
| Epic 9.5 backlog, stories in backlog | Stage shows "E9.5 S9.5-7 backlog" or similar | Shows "Unknown" or "-" |
| Epic 9.5 backlog, no stories | Stage shows "E9.5 backlog" or "Planning" | Shows "Unknown" |

### Step 4: Restore sprint-status.yaml

Revert the `epic-9-5` status to `in-progress`.

### Decision Guide

| Situation | Action |
|-----------|--------|
| Backlog epic shows story info | Mark `done` |
| Backlog epic shows "Unknown" | FAIL - logic not working |
| Backlog epic shows empty stage | FAIL - stageformat parsing issue |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-9.5/9-5-7-stage-detection-backlog-epic.md`
- Previous story: `docs/sprint-artifacts/stories/epic-9.5/9-5-6-user-visible-changes-section.md`
- Project context: `docs/project-context.md`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Clean implementation without issues.

### Completion Notes List

- Story drafted via *create-story workflow (YOLO mode)
- Based on Epic 9 retrospective issue analysis
- Pattern follows G24 (deferred epic) implementation from Story 7-8
- Implementation completed 2026-01-02:
  - Moved `epicInfo` type to package level for reuse by `analyzeEpicStories` helper
  - Added `firstBacklogEpic` capture in epic status switch (line 251-261)
  - Added backlog epic fallback check after all-done check (line 323-335)
  - Created `analyzeEpicStories()` helper function (lines 600-693)
  - Added G25, G26, G27 tests with 7 test scenarios total
  - Updated `abbreviateEpicStatus()` to handle "backlog" status
  - Added 3 stageformat tests for epic backlog parsing
- Code review fixes applied 2026-01-02:
  - M1: Refactored in-progress path to use `analyzeEpicStories()` helper (~70 lines removed)
  - L3: Fixed nested warning pattern in `analyzeEpicStories()` - now properly adds to warnings slice
  - L2: Updated Dev Notes to reflect Option B implementation
- All tests pass, lint clean

### File List

| File | Change |
|------|--------|
| `internal/adapters/detectors/bmad/stage_parser.go` | Modified - add `epicInfo` type at package level, `firstBacklogEpic` capture, backlog epic fallback, `analyzeEpicStories()` helper |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | Modified - add G25, G26, G27 test cases (7 scenarios) |
| `internal/shared/stageformat/stageformat.go` | Modified - add "backlog" case to `abbreviateEpicStatus()` |
| `internal/shared/stageformat/stageformat_test.go` | Modified - add 3 tests for epic backlog parsing |
| `docs/sprint-artifacts/sprint-status.yaml` | Modified - story status update |

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-02 | SM (Bob) | Initial story creation via *create-story workflow (YOLO mode) |
| 2026-01-02 | Dev (Amelia) | Implementation complete - all ACs verified |
