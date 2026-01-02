# Validation Report

**Document:** docs/sprint-artifacts/stories/epic-11/11-1-project-state-model.md
**Checklist:** .bmad/bmm/workflows/4-implementation/create-story/checklist.md
**Date:** 2026-01-02
**Validator:** SM (Bob) - Story Context Quality Competition

## Summary

- Overall: 21/28 items passed (75%)
- Critical Issues: 3
- Enhancement Opportunities: 4
- LLM Optimization Suggestions: 5

## Section Results

### Step 2.1: Epics and Stories Analysis
Pass Rate: 3/4 (75%)

✓ PASS - Epic objectives and business value
Evidence: Lines 4-9 describe story goal clearly.

✓ PASS - Cross-story dependencies
Evidence: Line 46 references Story 11.4 dependency.

⚠ PARTIAL - Acceptance criteria completeness
Evidence: AC1 states "ALREADY IMPLEMENTED" but dev agent may not verify this. Need explicit instruction.
Impact: Dev agent may skip verification of existing code behavior.

✓ PASS - Technical requirements extracted
Evidence: Dev Notes section comprehensively analyzes existing implementation.

### Step 2.2: Architecture Deep-Dive
Pass Rate: 4/6 (67%)

✓ PASS - Technical stack with versions
Evidence: Consistent with architecture.md hexagonal pattern.

✓ PASS - Code structure and organization patterns
Evidence: Lines 119-129 show architecture compliance diagram.

✗ FAIL - Database schema changes require migration version increment
Evidence: Story says "00X_add_hibernated_at.sql" but current SchemaVersion is 2. Must specify version 3.
Impact: Wrong migration version will cause conflicts or be skipped.

✓ PASS - Testing standards mentioned
Evidence: AC7 requires 95%+ coverage.

⚠ PARTIAL - Missing specific file locations for helpers.go/queries.go updates
Evidence: File List mentions repository.go but not the query constant files.
Impact: Dev agent may not update all required files.

✗ FAIL - Missing UpdateState SQL query update for HibernatedAt
Evidence: queries.go line 32 shows `updateStateSQL` updates only state and updated_at, not hibernated_at.
Impact: StateService will call UpdateState but HibernatedAt won't be persisted.

### Step 2.3: Previous Story Intelligence
Pass Rate: 1/1 (100%)

✓ PASS - First story in epic
Evidence: No previous story to analyze.

### Step 2.4: Git History Analysis
Pass Rate: 1/1 (100%)

✓ PASS - Recent patterns referenced
Evidence: Story references existing files with line numbers.

### Step 2.5: Latest Technical Research
Pass Rate: 1/1 (100%)

✓ PASS - No external library updates needed
Evidence: Uses existing Go stdlib and domain patterns.

### Step 3: Disaster Prevention Gap Analysis
Pass Rate: 5/10 (50%)

#### 3.1 Reinvention Prevention
✓ PASS - Code reuse identified
Evidence: Lines 88-103 show existing implementation analysis table.

⚠ PARTIAL - Missing explicit instruction to extend existing UpdateState
Evidence: Story says create StateService but doesn't specify UpdateState enhancement.
Impact: Dev may create parallel implementation instead of enhancing existing.

#### 3.2 Technical Specification Gaps

✗ FAIL - Missing projectRow struct update for HibernatedAt
Evidence: helpers.go projectRow (lines 16-32) needs HibernatedAt field added.
Impact: Database will store HibernatedAt but Go won't read it.

✓ PASS - API contract clear
Evidence: Service signatures defined in Tasks 2.2-2.4.

⚠ PARTIAL - Missing rowToProject() update for HibernatedAt
Evidence: helpers.go rowToProject (lines 34-72) needs HibernatedAt conversion.
Impact: Projects loaded from DB will have nil HibernatedAt.

#### 3.3 File Structure Gaps

✓ PASS - StateService location correct
Evidence: `internal/core/services/state_service.go` matches architecture.

✓ PASS - Migration file location correct
Evidence: `internal/adapters/persistence/sqlite/migrations/` correct per pattern.

#### 3.4 Regression Prevention

✓ PASS - Existing tests preserved
Evidence: Task 5.5 mentions integration tests for persistence.

#### 3.5 Implementation Gaps

✓ PASS - Error handling specified
Evidence: AC5 specifies domain error for invalid transitions.

### Step 4: LLM Optimization Analysis
Pass Rate: 3/5 (60%)

⚠ PARTIAL - Verbose Dev Notes table
Evidence: 11-line table could be reduced to critical items only.
Impact: Token waste on implemented items.

⚠ PARTIAL - Redundant AC1 marking "ALREADY IMPLEMENTED"
Evidence: AC1 adds confusion - either skip or verify behavior.
Impact: Dev agent unclear if verification needed.

✓ PASS - Clear task breakdown
Evidence: Tasks 1-5 are actionable with subtasks.

⚠ PARTIAL - Missing explicit wire-up instruction for StateService
Evidence: No instruction to register StateService in main.go or wire dependencies.
Impact: Service created but never callable.

✓ PASS - File List section present
Evidence: Lines 175-182 list expected files.

### User-Visible Changes Section
Pass Rate: 1/1 (100%)

✓ PASS - Section present and complete
Evidence: Lines 12-13 state "None - this is an internal domain model enhancement."

## Failed Items

### 1. Migration Version Not Specified (CRITICAL)
**Location:** Dev Notes, line 133
**Issue:** Uses "00X" placeholder instead of specific version 3
**Recommendation:** Change to `003_add_hibernated_at.sql` and add to migrations.go var

### 2. UpdateState SQL Missing HibernatedAt (CRITICAL)
**Location:** queries.go:32
**Issue:** `updateStateSQL` only updates state and updated_at
**Recommendation:** Create new `updateStateWithHibernatedAtSQL` or modify UpdateState signature

### 3. projectRow/rowToProject Missing HibernatedAt (CRITICAL)
**Location:** helpers.go:16-72
**Issue:** No HibernatedAt field in DB mapping
**Recommendation:** Add `HibernatedAt sql.NullString` to projectRow, parse in rowToProject

## Partial Items

### 1. File List Incomplete
**Missing Files:**
- `internal/adapters/persistence/sqlite/helpers.go` (projectRow update)
- `internal/adapters/persistence/sqlite/queries.go` (new SQL constants)
- `internal/adapters/persistence/sqlite/schema.go` (SchemaVersion bump)

### 2. No Wire-up Instructions
**Missing:** How to inject StateService into TUI/CLI
**Recommendation:** Add note about future Story 11.5 wire-up (not this story's scope)

### 3. AC1 Clarity
**Issue:** Marked "ALREADY IMPLEMENTED" is confusing
**Recommendation:** Change to verification task - confirm existing enum values match spec

### 4. Existing UpdateState Enhancement Not Specified
**Issue:** Story creates new StateService but UpdateState repo method exists
**Recommendation:** Clarify StateService calls repo.UpdateState with additional HibernatedAt logic

## Recommendations

### 1. Must Fix (Critical Failures)

1. **Specify migration version 3:**
   ```sql
   -- migrations.go - Add as version 3
   ALTER TABLE projects ADD COLUMN hibernated_at TEXT;
   ```

2. **Add to File List:**
   - `internal/adapters/persistence/sqlite/helpers.go` (add HibernatedAt to projectRow, rowToProject)
   - `internal/adapters/persistence/sqlite/queries.go` (add updateStateWithHibernatedAtSQL if needed)
   - `internal/adapters/persistence/sqlite/schema.go` (update SchemaVersion = 3)

3. **Clarify UpdateState enhancement:**
   Add subtask 4.3: "Update existing UpdateState() to also set hibernated_at column or create new UpdateStateWithHibernation()"

### 2. Should Improve (Important Gaps)

1. **Clarify AC1 verification:** Remove "ALREADY IMPLEMENTED" or change to "Verify existing implementation"

2. **Add wire-up scope note:** "StateService wire-up to TUI/CLI deferred to Story 11.5"

3. **Streamline Dev Notes table:** Remove rows for implemented items, keep only what needs creation

### 3. Consider (Minor Improvements)

1. **Add error wrapping pattern example:** Show how ErrInvalidStateTransition should be wrapped with context

2. **Mention test mock pattern:** Note that StateService tests should use mock repository per existing test patterns

### 4. LLM Optimization Improvements

1. **Reduce token count in Dev Notes:** Replace verbose table with focused list
2. **Add explicit "do not modify" markers:** For unchanged files to prevent over-editing
3. **Consolidate File List:** Group by action (CREATE vs MODIFY)
4. **Add validation checklist:** Quick-check items for dev agent to verify before completing
