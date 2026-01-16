# Validation Report

**Document:** docs/sprint-artifacts/stories/epic-16/16-5-implement-time-per-stage-breakdown.md
**Checklist:** .bmad/bmm/workflows/4-implementation/create-story/checklist.md
**Date:** 2026-01-16

## Summary
- Overall: 18/23 passed (78%)
- Critical Issues: 3
- Enhancement Opportunities: 4
- LLM Optimization: 2

## Section Results

### Section 1: Story Context Quality
Pass Rate: 4/5 (80%)

✓ PASS: Story file has correct metadata (Status: ready-for-dev)
Evidence: Line 3 shows `Status: ready-for-dev`

✓ PASS: User story follows proper format (As a/I want/So that)
Evidence: Lines 7-9 follow correct format

✓ PASS: Acceptance Criteria present with Given/When/Then structure
Evidence: Lines 20-38 contain 5 acceptance criteria in proper BDD format

✓ PASS: User-Visible Changes section present and non-empty
Evidence: Lines 12-17 list 4 new user-visible changes

⚠ PARTIAL: Dev Notes section exists but missing key learnings from Story 16.4
Evidence: Lines 230-415 have comprehensive notes, but missing explicit reference to `statsActiveProjectIdx` pattern and selection restoration logic from 16.4

### Section 2: Technical Specification Completeness
Pass Rate: 5/7 (71%)

✓ PASS: Tasks are numbered and have subtasks with checkboxes
Evidence: Lines 42-229 contain 9 tasks with detailed subtasks

✓ PASS: File paths are specified for each change
Evidence: Task 1 (line 43): `internal/adapters/tui/statsview/breakdown.go`, etc.

✓ PASS: Code examples provided with Go signatures
Evidence: Lines 44-72, 88-97, 108-118, etc. include interface and function signatures

✗ FAIL: Missing `statsViewSelected` field reference - field does not exist in model.go
Impact: Story refers to `m.statsViewSelected` (line 187) but codebase uses `statsActiveProjectIdx` for dashboard selection preservation. Developer will get compile error or implement incorrectly.
Evidence: Grep search shows no `statsViewSelected` in model.go; Story 16.4 uses `statsActiveProjectIdx`

⚠ PARTIAL: Missing exact location for key handling - Task 6 mentions creating `statsview_update.go` but should integrate into existing model.go Update method
Impact: Creating a separate file for Update method fragment violates existing pattern - all TUI state handling is in model.go
Evidence: Line 170 suggests creating `statsview_update.go` but codebase pattern puts key handling in model.go Update method

✓ PASS: Interface extension properly specified (FullTransitionReader)
Evidence: Lines 88-97 define interface extension pattern

✓ PASS: Import direction documented correctly
Evidence: Lines 242-251 specify clean architecture import flow

### Section 3: Previous Story Intelligence
Pass Rate: 4/5 (80%)

✓ PASS: References to Story 16.4 patterns included
Evidence: Lines 362-366 reference metricsReader pattern, graceful degradation, getProjectActivity pattern

✓ PASS: References to Story 16.3 patterns included
Evidence: Lines 367-370 reference view switching, key bindings, scroll state management

✗ FAIL: Missing critical pattern from Story 16.4 - `enterStatsView()` method
Impact: Story doesn't mention that entering stats view already has a method `enterStatsView()` that sets `statsActiveProjectIdx`. The breakdown navigation needs to work WITH this, not alongside.
Evidence: Story 16.4 completed file list shows model.go was modified to add enterStatsView/exitStatsView methods. Story 16.5 doesn't reference these.

✓ PASS: File modification list included
Evidence: Lines 255-261 table shows all files to modify/create

✓ PASS: Architecture alignment documented
Evidence: Lines 234-251 document FR-P2-17 implementation and isolation principle

### Section 4: Anti-Pattern Prevention
Pass Rate: 3/4 (75%)

✓ PASS: Code reuse opportunities identified (extends existing metricsReader)
Evidence: Lines 237-239 specify extending existing interface

✓ PASS: Graceful degradation pattern specified
Evidence: AC #4 (lines 32-33), Lines 357-358 specify nil-safe checks

⚠ PARTIAL: Missing warning about potential race condition
Impact: Story doesn't warn that fetching full transitions while user navigates could cause UI lag
Evidence: No mention of async loading or caching strategy for breakdown data

✓ PASS: Testing strategy defined with specific test cases
Evidence: Lines 211-228 specify comprehensive unit test list

### Section 5: LLM Optimization Analysis
Pass Rate: 2/2 (100%)

✓ PASS: Task breakdown is actionable with clear steps
Evidence: All 9 tasks have specific subtasks with code examples

✓ PASS: Duration calculation algorithm is explicit
Evidence: Lines 283-328 provide complete algorithm with code

## Failed Items

### 1. ✗ FAIL: Incorrect field name `statsViewSelected`
**Location:** Task 6, line 187
**Issue:** Story uses `m.statsViewSelected` but this field doesn't exist. The actual field is `statsActiveProjectIdx` (for preserving dashboard selection when entering Stats View).
**Recommendation:** Replace `statsViewSelected` with correct field. The breakdown view needs NEW state fields:
- `statsBreakdownProject *domain.Project` - for tracking breakdown view state
- `statsBreakdownDurations []StageDuration` - for cached data

### 2. ✗ FAIL: Missing `enterStatsView()` / `exitStatsView()` awareness
**Location:** Task 5 and Task 6
**Issue:** Story 16.4 implemented `enterStatsView()` and `exitStatsView()` methods in model.go. Story 16.5's breakdown navigation MUST integrate with these, but they aren't mentioned.
**Recommendation:** Add explicit guidance: "Modify existing `enterStatsView()` and `exitStatsView()` methods or add parallel `enterBreakdownView()` / `exitBreakdownView()` methods."

### 3. ✗ FAIL: Task 6 suggests wrong file location
**Location:** Task 6, line 170
**Issue:** Suggests creating `statsview_update.go` but TUI pattern puts all Update method handling in model.go
**Recommendation:** Change to: "In `internal/adapters/tui/model.go`, add case handling for stats breakdown view in the existing Update method"

## Partial Items

### 1. ⚠ Missing statsActiveProjectIdx awareness
**Gap:** Dev notes mention Story 16.3's scroll state but don't mention that `statsActiveProjectIdx` preserves dashboard selection - this is CRITICAL for restoration after exiting breakdown
**Fix:** Add explicit note about using `statsActiveProjectIdx` for dashboard restoration

### 2. ⚠ Missing async loading consideration
**Gap:** Breakdown data fetching could block UI for projects with many transitions
**Fix:** Consider adding async command pattern like `getStageBreakdownCmd()` that returns a message

### 3. ⚠ Missing viewMode integration
**Gap:** Story doesn't specify that breakdown view might need its own viewMode OR use the existing viewModeStats
**Fix:** Clarify: breakdown is a SUB-STATE of viewModeStats, not a separate viewMode. Use `statsBreakdownProject != nil` to distinguish list vs breakdown view.

## Recommendations

### Must Fix (Critical)

1. **Correct statsViewSelected → statsBreakdownProject pattern**
   - Replace all references to `statsViewSelected` with proper breakdown state fields
   - Add `statsBreakdownProject *domain.Project` and `statsBreakdownDurations []StageDuration` to model state

2. **Add enterStatsView/exitStatsView integration guidance**
   - Reference that Story 16.4 added these methods
   - Specify how breakdown navigation interacts with existing view switching

3. **Fix file location for key handling**
   - Remove suggestion to create statsview_update.go
   - Direct integration into model.go Update method

### Should Improve (Enhancement)

1. **Add explicit viewMode clarification**
   - Breakdown view is a sub-state of viewModeStats
   - Check `m.statsBreakdownProject != nil` to determine list vs breakdown

2. **Add stage ordering guidance**
   - Current algorithm sorts by duration descending
   - Consider sorting by FIRST OCCURRENCE for logical workflow order

3. **Add caching strategy note**
   - Suggest caching `statsBreakdownDurations` to avoid re-fetch on every render
   - Clear cache when exiting breakdown view

4. **Reference existing test patterns**
   - Story 16.4 tests in statsview_test.go provide patterns for mocking metricsReader

### LLM Optimization

1. **Reduce verbosity in Task 6 code example**
   - The handleStatsViewKey example is verbose; simplify to essential logic

2. **Add explicit "DO NOT" guidance**
   - DO NOT create statsview_update.go - use model.go
   - DO NOT use statsViewSelected - use statsBreakdownProject
