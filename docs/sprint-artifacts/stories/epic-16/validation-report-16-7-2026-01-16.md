# Validation Report

**Document:** docs/sprint-artifacts/stories/epic-16/16-7-wire-stats-view-into-dashboard.md
**Checklist:** .bmad/bmm/workflows/4-implementation/create-story/checklist.md
**Date:** 2026-01-16

## Summary
- Overall: 14/17 passed (82%)
- Critical Issues: 1
- Enhancement Opportunities: 3
- Optimizations: 2

## Section Results

### Step 1: Load and Understand the Target
Pass Rate: 4/4 (100%)

[PASS] Story file loaded and parsed correctly
Evidence: Story file at `docs/sprint-artifacts/stories/epic-16/16-7-wire-stats-view-into-dashboard.md` contains complete structure

[PASS] Metadata extracted correctly (epic_num=16, story_num=7, story_key=16-7)
Evidence: Title "Story 16.7: Wire Stats View into Dashboard" and Status: ready-for-dev

[PASS] Workflow variables resolved correctly
Evidence: story_dir, output_folder, epics_file correctly referenced

[PASS] Current status understood
Evidence: Story shows 95% complete - only Task 5 (help overlay) needs implementation

### Step 2: Exhaustive Source Document Analysis
Pass Rate: 5/5 (100%)

[PASS] Epics file analyzed
Evidence: `docs/epics-phase2.md` Story 4.7 (lines 756-779) defines acceptance criteria:
- AC#1: Press 's' opens Stats View
- AC#2: Status bar shows `[s] stats` hint
- AC#3: Stats View preserves Dashboard selection on return

[PASS] Architecture understood
Evidence: Hexagonal architecture from `docs/architecture.md` - TUI adapters in `internal/adapters/tui/`

[PASS] Previous story context loaded
Evidence: Story 16.6 (date range) and 16.3 (Stats View base) patterns analyzed:
- enterStatsView()/exitStatsView() at model.go:3395-3410
- viewModeStats at validation.go:22
- KeyStats = "s" at keys.go:37

[PASS] Git history analyzed
Evidence: Recent commits show incremental implementation pattern:
- 0f248b6: Story 16.6 date range
- 5b96685: Story 16.5 time breakdown
- e431013: Story 16.4 sparklines
- d7434cc: Story 16.3 Stats View TUI

[PASS] Technical research complete
Evidence: Go patterns verified - Bubble Tea view mode switching

### Step 3: Disaster Prevention Gap Analysis
Pass Rate: 3/6 (50%)

[PASS] Reinvention prevention
Evidence: Story correctly identifies that Tasks 1-4, 6 are ALREADY DONE in previous stories. Clear DO NOT duplicate warnings.

[PASS] Technical specification correct
Evidence: File path `internal/adapters/tui/views.go` line 115 correctly identified

[FAIL] **CRITICAL: Help overlay line number is WRONG**
Evidence: Story says line 115, but actual views.go shows:
- Line 113: "Views" section header
- Line 114: `h        View hibernated projects`
- Line 115: Empty string `""`
The 's' entry should be inserted at line 115 BEFORE the empty string (between h and blank line).
Impact: Developer may insert at wrong location or be confused

[PARTIAL] File structure requirements
Evidence: Story mentions views.go but doesn't show the exact line context for verification

[PASS] Regression prevention
Evidence: Story includes Task 7 (make test) and Task 8 (manual verification)

[PARTIAL] User-Visible Changes section
Evidence: Section exists and has content, but could be more precise about what's NEW vs VERIFIED

### Step 4: LLM-Dev-Agent Optimization Analysis
Pass Rate: 2/2 (100%)

[PASS] Token efficiency
Evidence: Story is well-structured with clear sections, not overly verbose

[PASS] Actionable instructions
Evidence: Tasks have specific file:line references and code snippets

## Failed Items

### ✗ Help overlay line number is WRONG (CRITICAL)
**Issue:** Story Task 5 states "add line after line 114" and "Location: Views section (between line 114 'h View hibernated projects' and line 115 empty string)". However, the correct insertion point should specify inserting the new line TO BECOME line 115, pushing the empty string to line 116.
**Impact:** Developer confusion, potential wrong placement
**Recommendation:** Update to explicit insertion code:
```go
// In views.go, Views section (lines 113-115), add 's' entry:
"Views",
"h        View hibernated projects",
"s        View stats and metrics",  // <-- ADD THIS LINE
"",
```

## Partial Items

### ⚠ File structure requirements could be more explicit
**Issue:** Story provides line number but not the surrounding context for verification
**What's missing:** Show before/after code for views.go modification
**Recommendation:** Add explicit before/after code block

### ⚠ User-Visible Changes section formatting
**Issue:** Uses "Verified" prefix for existing functionality - could be clearer
**What's missing:** Distinction between what this story adds vs confirms
**Recommendation:** Simplify to focus on NEW change only

## Recommendations

### 1. Must Fix: (Critical)
- Fix Task 5 instructions with exact code insertion context
- Clarify line 115 is WHERE to insert (not after)

### 2. Should Improve: (High Value)
- Add before/after code block for views.go modification
- Add exact grep command to verify current state
- Reduce redundant "ALREADY DONE" repetition in Dev Notes

### 3. Consider: (Nice to Have)
- Consolidate "Existing Code References" into more compact table
- Remove duplicate line number references that are also in Tasks
- Add test command to verify help overlay content

## LLM Optimization Improvements

1. **Reduce verbosity in Tasks section** - Tasks 1-4, 6 are marked ALREADY DONE but still show full details; these can be summarized
2. **Consolidate code references** - Same line numbers appear in multiple sections; single source of truth preferred
3. **Clarify the ONLY action** - Story is 95% done; make the single remaining task ultra-clear at the top
