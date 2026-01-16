# Validation Report

**Document:** docs/sprint-artifacts/stories/epic-16/16-4-implement-activity-sparklines.md
**Checklist:** .bmad/bmm/workflows/4-implementation/create-story/checklist.md
**Date:** 2026-01-16

## Summary
- Overall: 28/35 passed (80%)
- Critical Issues: 3
- Enhancements: 4
- LLM Optimizations: 2

## Section Results

### 1. Story Structure
Pass Rate: 5/5 (100%)

✓ **User-Visible Changes section present**
Evidence: Lines 11-15 - "## User-Visible Changes" section exists with 3 bulleted items (New indicators)

✓ **Story section with proper format**
Evidence: Lines 6-9 - Follows "As a user, I want..., So that..." format

✓ **Acceptance Criteria with BDD format**
Evidence: Lines 18-31 - 5 acceptance criteria with Given/When/Then format

✓ **Tasks/Subtasks breakdown**
Evidence: Lines 33-167 - 9 detailed tasks with subtasks and code snippets

✓ **Dev Notes section**
Evidence: Lines 169-341 - Comprehensive architecture, file mods, implementation details

### 2. Technical Accuracy
Pass Rate: 8/10 (80%)

✓ **Repository method signatures match existing patterns**
Evidence: Task 1 code snippets follow `(r *MetricsRepository)` receiver pattern from Story 16.1 (repository.go:79-103)

✓ **Graceful degradation pattern consistent**
Evidence: Lines 264-288 show return `nil` on errors, `slog.Warn()` logging - matches Story 16.1 patterns

✓ **Width calculation pattern correct**
Evidence: Lines 302-307 correctly reference `m.isWideWidth()` and `m.maxContentWidth` pattern from project-context.md

⚠ **PARTIAL: StageTransition struct location mismatch**
Evidence: Task 1 says add `StageTransition` to `helpers.go` (line 47-57), but a `stageTransitionRow` already exists in helpers.go:12-18. The story should clarify this is a NEW public type distinct from the internal row struct.
Impact: Developer may be confused about whether to modify existing or add new struct.

⚠ **PARTIAL: Interface location creates import cycle risk**
Evidence: Task 4 (lines 95-111) puts `MetricsReader` interface in `internal/adapters/tui/statsview/interfaces.go`, but also references `StageTransition` type. If `StageTransition` is in `metrics/helpers.go`, the TUI package would import from persistence layer, potentially violating hexagonal architecture.
Impact: Potential architecture boundary violation.

✗ **FAIL: Missing selectByProjectSQL query with time filter**
Evidence: Task 1 references `selectByProjectSQL` (line 46), but existing queries.go:15-17 shows this query returns ALL transitions for a project with no time filter. Story 16.4 needs transitions filtered to last 30 days for sparklines. No new query constant is specified.
Impact: Performance issue and incorrect data retrieval.

### 3. File Modification Accuracy
Pass Rate: 6/7 (86%)

✓ **repository.go changes identified**
Evidence: Lines 182-184 correctly identify adding GetTransitionsByProject, GetTransitionsByTimeRange

✓ **New files correctly specified**
Evidence: Lines 186-188 - sparkline.go, activity.go, interfaces.go as NEW files

✓ **model.go changes identified**
Evidence: Line 190 - Add metricsReader field and setter

✗ **FAIL: Missing queries.go modification**
Evidence: Lines 182-191 file modification table does not include `queries.go`, but a new query for time-filtered project transitions is needed (see Technical Accuracy fail above).
Impact: Incomplete file list will cause missed implementation.

✓ **statsview.go location correct**
Evidence: Line 189 references `internal/adapters/tui/statsview.go` which exists (verified at /internal/adapters/tui/statsview.go)

✓ **main.go wiring identified**
Evidence: Line 191 - Wire MetricsRepository to Model

✓ **Correct test file locations**
Evidence: Lines 147-162 - sparkline_test.go, activity_test.go, repository_test.go in correct packages

### 4. Previous Story Pattern Adherence
Pass Rate: 4/4 (100%)

✓ **Story 16.1 repository patterns referenced**
Evidence: Lines 299-303 explicitly reference graceful degradation, slog.Warn, sync.Once patterns from Story 16.1

✓ **Story 16.3 TUI patterns referenced**
Evidence: Lines 305-309 reference width calculation, content height, exit handling from Story 16.3

✓ **project-context.md patterns followed**
Evidence: Dev Notes reference hexagonal architecture, naming conventions, testing rules

✓ **Testing strategy follows established patterns**
Evidence: Lines 310-320 - Unit test focus, manual verification steps match project-context.md patterns

### 5. LLM-Dev-Agent Optimization
Pass Rate: 5/9 (56%)

✓ **Clear code snippets provided**
Evidence: Tasks 1-4 contain complete Go code snippets with proper formatting

✓ **Normalization algorithm complete**
Evidence: Lines 207-233 provide full implementation of normalize() function

✓ **Time bucketing strategy documented**
Evidence: Lines 236-257 detail bucket calculation with code example

⚠ **PARTIAL: Verbosity in Dev Notes**
Evidence: Lines 169-341 contain extensive reference material that could be condensed. Some sections duplicate content from earlier tasks.
Impact: Higher token consumption, potential for dev agent to skip important details.

⚠ **PARTIAL: Missing concrete wiring code for main.go**
Evidence: Task 7 (lines 140-144) says "In cmd/vdash/main.go or dependency wiring" but provides no code snippet. Compare to Tasks 1-6 which all have concrete code.
Impact: Developer must figure out wiring location and code pattern independently.

✗ **FAIL: BucketActivityCounts uses wrong type**
Evidence: Task 3 (lines 78-94) signature uses `[]StageTransition` but Task 4 defines `MetricsReader` returning `[]StageTransition`. If `StageTransition` is in metrics package, the statsview package would need to import it, creating coupling. Should use a local type or interface.
Impact: Package dependency issue.

✗ **FAIL: Task 5 fetching strategy incomplete**
Evidence: Lines 113-130 show `getProjectActivity()` calls `GetTransitionsByProject()` which returns ALL transitions, not just last 30 days. Need to filter by time range after fetch OR use GetTransitionsByTimeRange.
Impact: Incorrect implementation guidance.

✗ **FAIL: Missing context parameter in Task 3**
Evidence: Task 3 `BucketActivityCounts` has no context.Context parameter, but all repository methods require context. The caller must manage context but this isn't shown.
Impact: Minor - context handling unclear.

### 6. Disaster Prevention
Pass Rate: 5/5 (100%)

✓ **Reinvention prevention addressed**
Evidence: Story explicitly references existing selectByProjectSQL, stageTransitionRow from Story 16.1

✓ **Security requirements met**
Evidence: No SQL injection risk - using parameterized queries from existing patterns

✓ **Performance requirements documented**
Evidence: Lines 293-297 NFR compliance table shows O(n) render requirement

✓ **Graceful degradation emphasized**
Evidence: AC5 (line 31), Task 5 (line 123-124), multiple references throughout

✓ **Test coverage specified**
Evidence: Tasks 8-9 (lines 146-167) specify comprehensive unit tests with edge cases

## Failed Items

### F1: Missing time-filtered query for project transitions (CRITICAL)
**Location:** Task 1 / queries.go reference
**Issue:** Story references `selectByProjectSQL` but this returns all transitions. For 30-day sparklines, need filtered query or post-filter.
**Recommendation:** Add new query constant or modify Task 1 to include time-range parameter in GetTransitionsByProject:
```go
func (r *MetricsRepository) GetTransitionsByProject(ctx context.Context, projectID string, since time.Time) []StageTransition
```

### F2: StageTransition type location creates coupling (CRITICAL)
**Location:** Tasks 1, 3, 4
**Issue:** `StageTransition` defined in metrics package, but used in statsview package. This creates adapter-to-adapter import.
**Recommendation:** Define a local `Transition` type in statsview package that holds only needed fields (ProjectID, TransitionedAt), or pass []time.Time to BucketActivityCounts.

### F3: Missing queries.go in file modification table (MEDIUM)
**Location:** Dev Notes > File Modifications
**Issue:** Table omits queries.go despite need for new/modified query.
**Recommendation:** Add queries.go to file modifications table.

### F4: Task 7 lacks concrete wiring code (MEDIUM)
**Location:** Task 7
**Issue:** No code snippet for main.go wiring unlike all other tasks.
**Recommendation:** Add concrete code showing MetricsRepository creation and Model.SetMetricsReader() call.

## Partial Items

### P1: BucketActivityCounts time handling
**Location:** Task 3 (lines 78-94)
**Issue:** Function receives transitions but must know reference time (now). Should receive explicit "now" parameter or use time.Now() internally.
**Missing:** Add `now time.Time` parameter or document internal time.Now() usage.

### P2: Interface placement architecture concern
**Location:** Task 4 (lines 95-111)
**Issue:** MetricsReader interface in statsview references StageTransition type from metrics package.
**Missing:** Either move interface to core/ports or use a local type to avoid cross-adapter imports.

## Recommendations

### 1. Must Fix: Time-filtered query (F1)
Add to Task 1:
```go
const selectByProjectWithTimeSQL = `
SELECT ` + transitionColumns + ` FROM stage_transitions
WHERE project_id = ? AND transitioned_at >= ?
ORDER BY transitioned_at ASC`
```
Update GetTransitionsByProject signature to accept `since time.Time`.

### 2. Must Fix: Type coupling (F2)
Change Task 3 to use simpler type:
```go
// BucketActivityCounts calculates activity per time bucket.
// timestamps: transition times (oldest first)
// buckets: number of time buckets
// timeRange: total time period to cover
// now: reference time for calculations
func BucketActivityCounts(timestamps []time.Time, buckets int, timeRange time.Duration, now time.Time) []int
```

### 3. Should Improve: Add wiring code to Task 7
```go
// In run() after dm initialization:
metricsDBPath := filepath.Join(dm.BaseDir(), "metrics.db")
metricsRepo := metrics.NewMetricsRepository(metricsDBPath)
// In TUI model creation:
tuiModel.SetMetricsReader(metricsRepo)
```

### 4. Consider: Reduce Dev Notes verbosity
- Move detailed algorithm explanations to code comments instead of duplicating in story
- Remove redundant pattern references (e.g., Story 16.1/16.3 patterns already in project-context.md)
