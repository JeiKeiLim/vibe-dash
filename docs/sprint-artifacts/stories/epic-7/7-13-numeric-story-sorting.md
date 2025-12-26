# Story 7.13: Numeric Story Sorting

Status: done

## Story

As a **user viewing my BMAD project status**,
I want **stories to sort in natural numeric order (7-1, 7-2, ..., 7-10)**,
So that **the "next story" detection correctly identifies 7-2 after 7-1, not 7-10**.

## Problem Statement

Current implementation uses lexicographic (string) sorting for story/epic keys. This causes:
- `"7-10"` to sort before `"7-2"` because `"1" < "2"` in ASCII
- Dashboard shows wrong "next story" when story numbers exceed single digits
- Same issue affects epic ordering (`epic-10` before `epic-2`)
- Retrospective ordering also affected (`epic-10-retrospective` before `epic-2-retrospective`)

**Root cause locations (verified from source):**

| Location | Line | Current Code | Issue |
|----------|------|--------------|-------|
| Story sorting | `stage_parser.go:322-323` | `sortedStories[i].key < sortedStories[j].key` | Lexicographic |
| Epic sorting | `stage_parser.go:185` | `sort.Strings(epicOrder)` | Lexicographic |
| Retrospective sorting | `stage_parser.go:264-265` | `retrospectives[i].epicNum < retrospectives[j].epicNum` | Lexicographic |

## Acceptance Criteria

1. **AC1: Stories Sort Numerically Within Epic**
   - Given stories 7-1, 7-2, 7-3, 7-9, 7-10, 7-11
   - When sorted for stage detection
   - Then order is: 7-1, 7-2, 7-3, 7-9, 7-10, 7-11
   - And NOT: 7-1, 7-10, 7-11, 7-2, 7-3, 7-9

2. **AC2: Epics Sort Numerically**
   - Given epics epic-1, epic-2, epic-10, epic-3
   - When sorted for processing
   - Then order is: epic-1, epic-2, epic-3, epic-10
   - And NOT: epic-1, epic-10, epic-2, epic-3

3. **AC3: Sub-Epics Sort Correctly**
   - Given epics epic-4, epic-4-5, epic-4-6, epic-5
   - When sorted for processing
   - Then order is: epic-4, epic-4-5, epic-4-6, epic-5

4. **AC4: Next Story Detection Correct**
   - Given story 7-1 is done
   - And stories 7-2 through 7-10 are in backlog
   - When determining next story
   - Then "next" is 7-2, NOT 7-10

5. **AC5: Existing Tests Pass**
   - Given all existing G-tests (gap tests) in stage_parser_test.go
   - When tests are run
   - Then all pass (no regressions)

6. **AC6: New Tests for Numeric Ordering**
   - Given new test cases for double-digit story numbers
   - When running test suite
   - Then tests verify correct numeric ordering for stories, epics, and retrospectives

7. **AC7: Retrospectives Sort Numerically**
   - Given retrospectives epic-6-retrospective, epic-10-retrospective, epic-7-retrospective
   - When all epics are done and retrospectives are checked
   - Then order is: epic-6-retrospective, epic-7-retrospective, epic-10-retrospective

## Tasks / Subtasks

- [x] Task 1: Implement naturalCompare helper function (AC: 1, 2, 3, 7)
  - [x] 1.1: Create `naturalCompare(a, b string) bool` in `stage_parser.go`
  - [x] 1.2: Parse numeric segments as integers, compare lexicographically otherwise
  - [x] 1.3: Handle edge cases: pure numeric, mixed alphanumeric, empty strings, sub-epic format (4-5)
  - [x] 1.4: Unit test the helper function directly

- [x] Task 2: Replace story sorting (AC: 1, 4)
  - [x] 2.1: Update `stage_parser.go:322-323` to use `naturalCompare`
  - [x] 2.2: Verify existing G19 test still passes

- [x] Task 3: Replace epic sorting (AC: 2, 3)
  - [x] 3.1: Update `stage_parser.go:185` - replace `sort.Strings(epicOrder)` with custom sort using naturalCompare
  - [x] 3.2: Verify sub-epic ordering (epic-4, epic-4-5, epic-4-6, epic-5)

- [x] Task 4: Replace retrospective sorting (AC: 7)
  - [x] 4.1: Update `stage_parser.go:264-265` to use `naturalCompare` on `epicNum`
  - [x] 4.2: Add test for retrospective ordering with double-digit epic numbers

- [x] Task 5: Add comprehensive test cases (AC: 5, 6)
  - [x] 5.1: Add test: stories 7-1 through 7-12, verify order
  - [x] 5.2: Add test: epics 1-12, verify order
  - [x] 5.3: Add test: mixed scenario (sub-epics + double-digit stories)
  - [x] 5.4: Add test: retrospectives with double-digit epics (G23 scenario)
  - [x] 5.5: Run full test suite, ensure no regressions

## Dev Notes

### Implementation Approach

Use a chunk-based natural sort algorithm that handles the sub-epic format:

```go
// naturalCompare returns true if a should sort before b using natural ordering.
// Handles: "7-2" vs "7-10", "epic-4" vs "epic-10", "epic-4-5" vs "epic-5"
func naturalCompare(a, b string) bool {
    // Split into chunks of digits and non-digits
    // Compare chunk by chunk:
    //   - Numeric chunks: compare as integers (7 < 10)
    //   - Non-numeric chunks: compare as strings ("epic-" == "epic-")
}
```

**Key insight:** The algorithm must handle:
- Simple stories: `7-2` vs `7-10` → chunks `[7, -, 2]` vs `[7, -, 10]` → 2 < 10
- Sub-epic stories: `4-5-2-foo` vs `4-5-10-bar` → chunks `[4, -, 5, -, 2, -, foo]` vs `[4, -, 5, -, 10, -, bar]` → 2 < 10
- Epics: `epic-4-5` vs `epic-10` → chunks `[epic, -, 4, -, 5]` vs `[epic, -, 10]` → 4 < 10

### Files to Modify

| File | Change |
|------|--------|
| `internal/adapters/detectors/bmad/stage_parser.go:185` | Replace `sort.Strings` with custom sort |
| `internal/adapters/detectors/bmad/stage_parser.go:264-265` | Replace string compare with naturalCompare |
| `internal/adapters/detectors/bmad/stage_parser.go:322-323` | Replace string compare with naturalCompare |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | Add numeric ordering tests |

### Testing Strategy

```bash
# Run stage_parser tests
go test ./internal/adapters/detectors/bmad/... -run TestDetermineStageFromStatus -v

# Run naturalCompare unit tests
go test ./internal/adapters/detectors/bmad/... -run TestNaturalCompare -v

# Full suite - verify no regressions
go test ./internal/adapters/detectors/bmad/... -v

# Lint check
golangci-lint run ./internal/adapters/detectors/bmad/...
```

### Test Cases to Add

```go
// TestNaturalCompare verifies the natural sort helper
func TestNaturalCompare(t *testing.T) {
    tests := []struct {
        a, b string
        want bool // a < b
    }{
        {"7-2", "7-10", true},     // 2 < 10
        {"7-10", "7-2", false},    // 10 > 2
        {"epic-2", "epic-10", true},
        {"epic-4-5", "epic-5", true},  // 4 < 5
        {"epic-10", "epic-4-5", false}, // 10 > 4
        {"1-1", "1-2", true},
        {"1-9", "1-10", true},
    }
    // ... test implementation
}

// TestDetermineStageFromStatus_DoubleDigitStories verifies natural ordering
func TestDetermineStageFromStatus_DoubleDigitStories(t *testing.T) {
    status := &SprintStatus{
        DevelopmentStatus: map[string]string{
            "epic-7":       "in-progress",
            "7-1-feature":  "done",
            "7-2-feature":  "backlog",
            "7-10-feature": "backlog",
            "7-11-feature": "backlog",
        },
    }
    stage, _, reasoning := determineStageFromStatus(status)

    // Should pick 7-2, NOT 7-10
    if !strings.Contains(reasoning, "7.2") {
        t.Errorf("expected 7.2 as next story, got: %s", reasoning)
    }
}
```

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Parse story keys assuming fixed format | Use generic chunk-based natural compare |
| Import external natural sort library | Implement inline (~30 lines) |
| Change existing test assertions | Only add new tests for numeric ordering |
| Modify G19 test expected values | Verify G19 still passes with new sorting |

### References

- [Source: internal/adapters/detectors/bmad/stage_parser.go:322-323] - Current story sort
- [Source: internal/adapters/detectors/bmad/stage_parser.go:185] - Current epic sort
- [Source: internal/adapters/detectors/bmad/stage_parser.go:264-265] - Current retrospective sort
- [Source: internal/adapters/detectors/bmad/stage_parser_test.go:580-594] - G19 deterministic order test
- [Source: docs/project-context.md] - Testing rules, naming conventions

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Verify Tests Pass

```bash
go test ./internal/adapters/detectors/bmad/... -v
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| All tests pass | `ok` for package | Any `FAIL` |
| New numeric tests exist | Tests with "DoubleDigit" in name | No numeric ordering tests |

### Step 2: Spot-Check with Mock Sprint Status

Create a temporary test sprint-status.yaml:

```bash
cat > /tmp/test-sprint-status.yaml << 'EOF'
development_status:
  epic-7: in-progress
  7-1-feature: done
  7-2-feature: backlog
  7-10-feature: backlog
  7-11-feature: backlog
EOF
```

The reasoning should say "Story 7.2 in backlog" NOT "Story 7.10".

### Step 3: Verify No Regressions

```bash
# Run all BMAD detector tests
go test ./internal/adapters/detectors/bmad/... -v | grep -E "(PASS|FAIL)"

# Should see all PASS, no FAIL
```

### Decision Guide

| Situation | Action |
|-----------|--------|
| All tests pass, 7.2 detected as next | Mark `done` |
| Tests fail | Do NOT approve, investigate |
| 7.10 detected before 7.2 | Do NOT approve, sorting not fixed |
| G19 test fails | Do NOT approve, regression introduced |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

### Completion Notes List

1. **naturalCompare implementation**: Created chunk-based natural sort algorithm that splits strings into numeric and non-numeric segments, comparing numeric segments as integers and non-numeric segments lexicographically.
2. **Helper functions added**: `splitIntoChunks()` and `parseUint()` support the naturalCompare function.
3. **All three sort locations updated**:
   - Story sorting at line 322-323
   - Epic sorting at line 185 (replaced `sort.Strings` with `sort.Slice` using naturalCompare)
   - Retrospective sorting at line 267-269
4. **Test coverage added**:
   - `TestNaturalCompare` - 29 test cases covering all edge cases
   - `TestDetermineStageFromStatus_DoubleDigitStories` - AC1/AC4 verification
   - `TestDetermineStageFromStatus_DoubleDigitEpics` - AC2/AC3 verification
   - `TestDetermineStageFromStatus_DoubleDigitRetrospectives` - AC7 verification
   - `TestDetermineStageFromStatus_MixedDoubleDigitScenario` - AC5 combined scenario
5. **All existing tests pass** - No regressions, G19 deterministic order test still passes
6. **Linter passes** - golangci-lint clean after goimports

### Code Review Fixes Applied
7. **M1 fix**: Added antisymmetry property verification in TestNaturalCompare and new TestNaturalCompare_Antisymmetry test
8. **M3 fix**: Added leading zeros handling test cases ("7-01" vs "7-2", "7-1" vs "7-01")
9. **L1 fix**: Improved splitIntoChunks comment to clarify chunk boundary behavior with dashes

### File List

| File | Change |
|------|--------|
| `internal/adapters/detectors/bmad/stage_parser.go` | Added `naturalCompare`, `splitIntoChunks`, `parseUint` functions; updated 3 sort locations |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | Added `TestNaturalCompare` and 4 integration tests for double-digit scenarios |
