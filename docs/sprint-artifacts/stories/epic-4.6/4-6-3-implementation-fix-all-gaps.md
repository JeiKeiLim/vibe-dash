# Story 4.6.3: Implementation - Fix All Gaps

Status: done

## Story

As a developer using vibe-dash,
I want all 22 stage detection gaps fixed as specified in Story 4.5.2,
so that stage detection accurately reflects my project state in all scenarios.

## Acceptance Criteria

1. **P1 Gaps Fixed:** Given the P1 gaps (G1, G7, G15), when implementation completes, then:
   - G1: Epic in-progress with all stories done shows "Epic N stories complete, update epic status"
   - G7: Epic done with active stories shows warning with actual story status
   - G15: LLM typo variations (spaces, underscores, synonyms) are normalized correctly

2. **P2 Gaps Fixed:** Given the P2 gaps (G2, G3, G8, G14, G17, G19, G22), when implementation completes, then:
   - G2/G3: `drafted` and `ready-for-dev` stories display appropriately
   - G8: Epic backlog with active stories shows warning
   - G14: Orphan stories (no matching epic) are warned
   - G17: Status synonyms are normalized
   - G19: Story order is deterministic (first by sorted key)
   - G22: Empty status values are warned

3. **Status Normalization:** Given the normalizeStatus() function, when implemented, then all test cases in the Normalization Test Cases table pass

4. **Test Coverage:** Given the Test Matrix, when tests are written, then:
   - All P1 gap test cases pass
   - All P2 gap test cases pass
   - All Normalization test cases pass
   - Existing tests continue to pass

5. **No Regressions:** Given the existing test suite, when all fixes are applied, then `go test ./internal/adapters/detectors/bmad/...` passes with 100% existing test cases

## Tasks / Subtasks

### P1 Priority (Must Fix)

- [x] Task 1: Fix G1 - All Stories Done Detection (AC: #1)
  - [x] 1.1 Add check after story iteration loop in determineStageFromStatus()
  - [x] 1.2 Check if all stories in epic have normalized status `done`
  - [x] 1.3 Return StageImplement "Epic N stories complete, update epic status"
  - [x] 1.4 Write test: epic in-progress, all stories done
  - [x] 1.5 Write test: multiple epics, one with all stories done

- [x] Task 2: Fix G7 - Epic Done with Active Stories (AC: #1)
  - [x] 2.1 After all-epics-done check, scan stories in done epics
  - [x] 2.2 If story in-progress, return StageImplement with warning
  - [x] 2.3 If story in review, return StageTasks with warning
  - [x] 2.4 Write test: epic done, story in-progress
  - [x] 2.5 Write test: epic done, story in review

- [x] Task 3: Fix G15/G17 - Status Normalization (AC: #1, #3) **REQUIRED BEFORE Tasks 4-7**
  - [x] 3.1 Add normalizeStatus() function (see Dev Notes for complete code)
  - [x] 3.2 Replace `strings.ToLower(value)` with `normalizeStatus(value)` in epic parsing
  - [x] 3.3 Replace `strings.ToLower(value)` with `normalizeStatus(value)` in story parsing
  - [x] 3.4 Write unit tests for normalizeStatus() (16 cases)

### P2 Priority (Should Fix) - Requires Task 3 Complete

- [x] Task 4: Fix G2/G3 - Drafted and Ready-for-Dev (AC: #2)
  - [x] 4.1 Add `drafted` case in story status switch
  - [x] 4.2 Add `ready-for-dev` case in story status switch
  - [x] 4.3 Implement story priority selection (see Dev Notes for algorithm)
  - [x] 4.4 Write tests for drafted-only and ready-for-dev-only scenarios

- [x] Task 5: Fix G8 - Epic Backlog with Active Stories (AC: #2)
  - [x] 5.1 After backlog epic detection, scan for active stories
  - [x] 5.2 Return StageSpecify with warning about inconsistent state
  - [x] 5.3 Write tests for backlog epic with in-progress/done stories

- [x] Task 6: Fix G19 - Deterministic Story Order (AC: #2)
  - [x] 6.1 Sort stories within each epic before iteration
  - [x] 6.2 Use first match by sorted order
  - [x] 6.3 Write test with multiple stories verifying first-by-sorted-key wins

- [x] Task 7: Fix G14/G22 - Data Quality Warnings (AC: #2)
  - [x] 7.1 Track orphan stories (story prefix doesn't match any epic)
  - [x] 7.2 Check for empty status values before normalization
  - [x] 7.3 Include warnings in reasoning string (don't fail detection)
  - [x] 7.4 Write tests for orphan stories and empty status

### Final Verification

- [x] Task 8: Full Test Suite Verification (AC: #4, #5)
  - [x] 8.1 Run full test suite
  - [x] 8.2 Run linter
  - [x] 8.3 Verify 100% existing tests still pass (see anti-regression list)
  - [x] 8.4 Count new test cases added (target: 20+) - **29 new tests added**
  - [x] 8.5 Update sprint-status.yaml: `4-6-3-implementation-fix-all-gaps: done`

## Dev Notes

### Files to Modify

| File | Action |
|------|--------|
| `internal/adapters/detectors/bmad/stage_parser.go` | MODIFY - Add normalizeStatus(), fix gap logic |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | MODIFY - Add 20+ new test cases |
| `docs/sprint-artifacts/sprint-status.yaml` | UPDATE - Mark story done when complete |

### normalizeStatus() - Complete Implementation (COPY THIS)

```go
// normalizeStatus converts common LLM variations to canonical status values.
// Apply BEFORE switch statement comparison.
func normalizeStatus(status string) string {
    // 1. Lowercase and trim
    s := strings.ToLower(strings.TrimSpace(status))

    // 2. Normalize separators: spaces and underscores → hyphens
    s = strings.ReplaceAll(s, " ", "-")
    s = strings.ReplaceAll(s, "_", "-")

    // 3. Map synonyms (G17)
    synonyms := map[string]string{
        "complete":    "done",
        "completed":   "done",
        "finished":    "done",
        "wip":         "in-progress",
        "inprogress":  "in-progress",
        "reviewing":   "review",
        "in-review":   "review",
        "code-review": "review",
    }

    if canonical, ok := synonyms[s]; ok {
        return canonical
    }
    return s
}
```

### Story Status Priority Algorithm (G19)

When multiple stories exist, select by this priority order:
```go
// Priority: review > in-progress > ready-for-dev > drafted > backlog > done
var storyPriority = map[string]int{
    "review":        1,
    "in-progress":   2,
    "ready-for-dev": 3,
    "drafted":       4,
    "backlog":       5,
    "done":          6,
}

// After sorting stories by key, iterate and track highest priority story
var selectedStory string
var selectedPriority = 999

for _, story := range sortedStories {
    normalized := normalizeStatus(story.status)
    if p, ok := storyPriority[normalized]; ok && p < selectedPriority {
        selectedStory = story.key
        selectedPriority = p
    }
}
```

### Code Pattern Search Guide

Instead of line numbers, search for these patterns:

| Gap | Search For | Action |
|-----|-----------|--------|
| G15 (epic) | `status: strings.ToLower(value)` in epic loop | Replace with `status: normalizeStatus(value)` |
| G15 (story) | `status: strings.ToLower(value)` in story loop | Replace with `status: normalizeStatus(value)` |
| G1 | `// Epic in-progress but no stories started` | Insert all-done check BEFORE this comment |
| G2/G3 | `case "in-progress":` in story switch | Add `case "drafted":` and `case "ready-for-dev":` |
| G7/G8 | `if len(epics) > 0 && doneCount == len(epics)` | After this block, add inconsistent state checks |
| G19 | `for _, story := range firstInProgressEpic.stories` | Sort stories before this loop |

### Anti-Regression Test Functions

These existing tests MUST continue to pass:
- `TestParseSprintStatus`
- `TestParseSprintStatus_FileNotFound`
- `TestParseSprintStatus_ContextCancelled`
- `TestParseSprintStatus_ContextTimeout`
- `TestDetermineStageFromStatus` (all cases)
- `TestDetectStageFromArtifacts`
- `TestExtractStoryPrefix`
- `TestFormatStoryKey`
- `TestFormatEpicKey`

### G1 Implementation (All Stories Done Check)

Insert this after the story iteration loop (before "Epic in-progress but no stories started"):

```go
// G1: Check if ALL stories in this epic are done
allDone := true
hasStories := len(firstInProgressEpic.stories) > 0
for _, story := range firstInProgressEpic.stories {
    if normalizeStatus(story.status) != "done" {
        allDone = false
        break
    }
}
if hasStories && allDone {
    return domain.StageImplement, domain.ConfidenceCertain,
        formatEpicKey(firstInProgressEpic.key) + " stories complete, update epic status"
}
```

### G7/G8 Implementation (Inconsistent State Checks)

After the all-epics-done check:

```go
// G7: Check for done epics with active stories
for _, epicKey := range epicOrder {
    epic := epics[epicKey]
    if normalizeStatus(epic.status) == "done" {
        for _, story := range epic.stories {
            normalized := normalizeStatus(story.status)
            if normalized == "review" {
                return domain.StageTasks, domain.ConfidenceLikely,
                    "Epic done but Story " + formatStoryKey(story.key) + " in review"
            }
            if normalized == "in-progress" {
                return domain.StageImplement, domain.ConfidenceLikely,
                    "Epic done but Story " + formatStoryKey(story.key) + " in-progress"
            }
        }
    }
}

// G8: Check for backlog epics with active stories
for _, epicKey := range epicOrder {
    epic := epics[epicKey]
    if normalizeStatus(epic.status) == "backlog" {
        for _, story := range epic.stories {
            normalized := normalizeStatus(story.status)
            if normalized == "in-progress" || normalized == "done" || normalized == "review" {
                return domain.StageSpecify, domain.ConfidenceLikely,
                    "Epic backlog but Story " + formatStoryKey(story.key) + " active"
            }
        }
    }
}
```

### Normalization Test Cases

| Input Status | Expected Output |
|--------------|-----------------|
| `"in progress"` | `"in-progress"` |
| `"in_progress"` | `"in-progress"` |
| `"IN-PROGRESS"` | `"in-progress"` |
| `"wip"` | `"in-progress"` |
| `"complete"` | `"done"` |
| `"completed"` | `"done"` |
| `"finished"` | `"done"` |
| `"reviewing"` | `"review"` |
| `"in-review"` | `"review"` |
| `"code-review"` | `"review"` |

### Out of Scope (P3 Gaps)

The following gaps from Story 4.5.2 are P3 (nice to have) and not included in this story:
- G4, G5, G6: Multi-story display enhancements (partially addressed by G19)
- G9, G10: Unknown status value messages
- G11, G12, G13: Whitespace and case edge cases (addressed by G15 normalization)
- G16, G18: Deep nesting support
- G20, G21: Contexted patterns and uppercase keys

### Verification Commands

```bash
# Full test suite (includes all anti-regression tests)
go test ./internal/adapters/detectors/bmad/... -v

# Linter
golangci-lint run ./internal/adapters/detectors/bmad/...

# Specific gap tests (after implementation)
go test ./internal/adapters/detectors/bmad/... -run "TestNormalizeStatus" -v
go test ./internal/adapters/detectors/bmad/... -run "AllStoriesDone" -v
go test ./internal/adapters/detectors/bmad/... -run "Inconsistent" -v
```

### References

| Document | Path |
|----------|------|
| **PRIMARY SPEC** | `docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md` |
| Project Context | `docs/project-context.md` |
| Implementation Target | `internal/adapters/detectors/bmad/stage_parser.go` |
| Test File | `internal/adapters/detectors/bmad/stage_parser_test.go` |

### New Gaps Discovered During Implementation

Document any new gaps found here. Add G-TBD with description in completion notes.

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Task 3 (G15/G17)**: Implemented first as foundation. Added `normalizeStatus()` function with synonym mapping (wip→in-progress, complete/completed/finished→done, reviewing/in-review/code-review→review). Handles spaces, underscores, and case variations. 16 test cases added.

2. **Task 1 (G1)**: Added all-stories-done detection after the story iteration loop. Returns "Epic N stories complete, update epic status" when epic is in-progress but all stories are done. 2 test cases added.

3. **Task 2 (G7)**: Moved G7 check BEFORE all-epics-done shortcut to catch inconsistent states. Returns appropriate warnings when epic is marked done but has active stories. 2 test cases added.

4. **Task 4 (G2/G3)**: Refactored story selection to use priority-based algorithm. Added `drafted` and `ready-for-dev` handling with appropriate messaging. Priority order: review > in-progress > ready-for-dev > drafted > backlog > done. 4 test cases added.

5. **Task 5 (G8)**: Added check for backlog epics with active stories. Returns StageSpecify with warning. 2 test cases added.

6. **Task 6 (G19)**: Implemented deterministic story ordering by sorting stories before priority selection. Ensures first-by-sorted-key wins when priorities are equal. 1 test case added.

7. **Task 7 (G14/G22)**: Added warnings tracking for orphan stories and empty status values. Warnings appended to reasoning string but don't fail detection. Warnings sorted for deterministic output. 2 test cases added.

8. **No new gaps discovered** during implementation.

### Code Review Fixes (2025-12-21)

9. **CR-1**: Fixed G14/G22 warning propagation - `appendWarnings()` was only called in fallback path. Now called on ALL return paths to ensure orphan story and empty status warnings appear in reasoning.

10. **CR-2**: Updated G14/G22 test cases to verify warnings are included in reasoning string (e.g., `"Story 1.1 being implemented [Warning: orphan story 2.1]"`).

11. **CR-3**: Replaced magic number `999` with named constant `unsetPriority` for clarity.

12. **CR-4**: Removed duplicate YAML header metadata from sprint-status.yaml.

### File List

| File | Action | Description |
|------|--------|-------------|
| `internal/adapters/detectors/bmad/stage_parser.go` | MODIFIED | Added normalizeStatus(), G1/G7/G8 checks, G2/G3 priority algorithm, G19 sorting, G14/G22 warnings, CR: fixed appendWarnings propagation |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | MODIFIED | Added 29 new test cases covering all gaps, CR: fixed G14/G22 test expectations |
| `docs/sprint-artifacts/sprint-status.yaml` | MODIFIED | Story marked done, CR: removed duplicate headers |
| `docs/sprint-artifacts/stories/epic-4.6/4-6-3-implementation-fix-all-gaps.md` | MODIFIED | Tasks marked complete, Dev Agent Record filled, Code Review notes added |
