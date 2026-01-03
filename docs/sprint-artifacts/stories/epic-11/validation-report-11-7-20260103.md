# Validation Report

**Document:** docs/sprint-artifacts/stories/epic-11/11-7-tui-manual-state-toggle.md
**Checklist:** .bmad/bmm/workflows/4-implementation/create-story/checklist.md
**Date:** 2026-01-03

## Summary
- Overall: 28/33 passed (85%)
- Critical Issues: 3
- Enhancements: 4
- LLM Optimizations: 2

## Section Results

### 2.1 Epics and Stories Analysis
Pass Rate: 5/5 (100%)

✓ PASS - Epic 11 context clearly captured (Stories 11.1-11.6 referenced)
Evidence: Lines 21-27 reference all prior stories with specific details

✓ PASS - Story requirements documented with user story format
Evidence: Lines 7-9 "As a user...I want to...So that..."

✓ PASS - Acceptance criteria complete with 7 ACs
Evidence: Lines 38-79 with Given/When/Then format

✓ PASS - Technical requirements captured
Evidence: Lines 126-139 Architecture Compliance diagram

✓ PASS - Cross-story dependencies listed
Evidence: Lines 336-337 Dependencies section

### 2.2 Architecture Deep-Dive
Pass Rate: 6/7 (86%)

✓ PASS - File structure documented
Evidence: Lines 141-145 lists files to modify

✓ PASS - Code patterns follow existing conventions
Evidence: Lines 127-139 "follows favorite toggle pattern"

✓ PASS - Testing standards documented
Evidence: Lines 288-298 Testing Strategy table

✗ FAIL - **strings.Title is deprecated in Go 1.18+**
Impact: The code example at line 254 uses `strings.Title(msg.action)` which is deprecated. Should use `cases.Title(language.English).String(msg.action)` from `golang.org/x/text/cases` or simple string formatting.
Recommendation: Replace with direct string formatting like `"✓ Hibernated"` / `"✓ Activated"` to avoid deprecated function AND match the CLI output format from Story 11.5.

✓ PASS - Domain errors documented for reuse
Evidence: Lines 326-328 reference ErrFavoriteCannotHibernate, ErrInvalidStateTransition

✓ PASS - View layer update documented
Evidence: Lines 271-283 help overlay changes

✓ PASS - Database/State updates documented
Evidence: Lines 256-268 list refresh and reload commands

### 2.3 Previous Story Intelligence
Pass Rate: 4/5 (80%)

✓ PASS - Story 11.4-11.6 learnings referenced
Evidence: Lines 21-27 comprehensive previous story section

✓ PASS - Reuse patterns documented
Evidence: Lines 320-331 "Reuse vs Create Quick Reference"

⚠ PARTIAL - Dev notes from Story 11.5 not fully leveraged
Gap: Story 11.5 extended StateActivator interface to include Hibernate() (line 1171-1175 of 11.5 story). Story 11.7 already references this but could note the interface was recently extended.
Impact: Minor - documentation could be clearer

✓ PASS - Code patterns from previous stories reused
Evidence: Line 127 "Follows existing favorite toggle pattern"

✓ PASS - Files created/modified patterns followed
Evidence: Lines 341-351 File List section

### 2.4 Git History Analysis
Pass Rate: N/A (Not applicable for story validation - done at implementation)

➖ N/A - Git history analysis is implementation-time concern

### 2.5 Technical Research
Pass Rate: 3/4 (75%)

✓ PASS - Bubble Tea patterns correctly used
Evidence: Lines 158-196 tea.Cmd and tea.Msg patterns

✓ PASS - Project naming utilities correctly referenced
Evidence: Line 325 references project.EffectiveName()

✗ FAIL - **Incorrect path in Reuse table**
Impact: Line 325 states `internal/shared/project/name.go` but file is actually `internal/shared/project/project.go`. Dev agent will waste time looking for wrong file.
Recommendation: Fix path to `internal/shared/project/project.go`

✓ PASS - Status bar pattern correctly used
Evidence: Lines 234-255 SetRefreshComplete pattern

### 3.1 Reinvention Prevention Gaps
Pass Rate: 3/3 (100%)

✓ PASS - Existing clearRemoveFeedbackMsg reused
Evidence: Lines 236-237, 249-251, 265-266 reuse existing pattern

✓ PASS - Existing StateActivator interface reused
Evidence: Lines 322-324 reference ports/state.go

✓ PASS - Existing loadProjectsCmd/loadHibernatedProjectsCmd reused
Evidence: Lines 256-268

### 3.2 Technical Specification DISASTERS
Pass Rate: 4/5 (80%)

✓ PASS - Correct view mode check
Evidence: Line 204 checks viewModeHibernated

✓ PASS - Nil checks included
Evidence: Lines 174-178 stateService nil check

✓ PASS - Error handling for all cases
Evidence: Lines 231-251 handles all error types

✗ FAIL - **Missing import statement in code example**
Impact: Code at line 254 uses `strings.Title` but the Technical Implementation Guide doesn't show required import. Additionally, `fmt` is used but not shown in imports for the message handler.
Recommendation: Add complete import list: `"context"`, `"errors"`, `"fmt"`, `"log/slog"`, `"strings"`, `"time"`

✓ PASS - Correct project selection method
Evidence: Lines 209, 219 use hibernatedList.SelectedProject() and projectList.SelectedProject()

### 3.3 File Structure DISASTERS
Pass Rate: 3/3 (100%)

✓ PASS - All files in correct locations
Evidence: Lines 341-351 show correct adapter/port paths

✓ PASS - Test file location correct
Evidence: Line 344 model_test.go in same package

✓ PASS - No new packages created
Evidence: Modification only, no new packages

### 3.4 Regression DISASTERS
Pass Rate: 4/4 (100%)

✓ PASS - Existing key bindings preserved
Evidence: Lines 147-152 add new KeyStateToggle without modifying existing

✓ PASS - View mode compatibility maintained
Evidence: Lines 204-224 check view mode before action

✓ PASS - Test coverage for all ACs
Evidence: Lines 116-122 list 7 test cases for 7 ACs

✓ PASS - Feedback timing consistent (3 seconds)
Evidence: Lines 235, 249, 265 all use 3*time.Second

### 3.5 Implementation DISASTERS
Pass Rate: 3/5 (60%)

✓ PASS - Task breakdown is actionable
Evidence: Lines 82-123 provide detailed subtasks

⚠ PARTIAL - **Line numbers in code examples are approximate**
Gap: Line numbers like "around line 240", "around line 528", "around line 1494", "around line 1250" are vague. After code changes in previous stories, these may be incorrect.
Recommendation: Remove line number hints or use grep patterns instead

✗ FAIL - **Missing KeyBindings struct update**
Impact: The story adds KeyStateToggle constant but does NOT update the KeyBindings struct (keys.go:33-55) and DefaultKeyBindings() function (keys.go:58-83). This creates inconsistency between constants and the configurable bindings struct.
Recommendation: Add `StateToggle string` to KeyBindings struct and add `StateToggle: KeyStateToggle` to DefaultKeyBindings()

✓ PASS - Edge cases documented
Evidence: Lines 300-306

## Failed Items

### 1. strings.Title Deprecation (Critical)
**Issue:** Line 254 uses deprecated `strings.Title()` function
**Recommendation:** Replace with:
```go
feedback := fmt.Sprintf("✓ %s: %s",
    map[bool]string{true: "Hibernated", false: "Activated"}[msg.action == "hibernated"],
    msg.projectName)
```
Or simply:
```go
var feedback string
if msg.action == "hibernated" {
    feedback = fmt.Sprintf("✓ Hibernated: %s", msg.projectName)
} else {
    feedback = fmt.Sprintf("✓ Activated: %s", msg.projectName)
}
```

### 2. Incorrect File Path (Critical)
**Issue:** Line 325 references `internal/shared/project/name.go`
**Recommendation:** Change to `internal/shared/project/project.go`

### 3. Missing KeyBindings Update (Critical)
**Issue:** KeyStateToggle constant added but KeyBindings struct not updated
**Recommendation:** Add to Technical Implementation Guide:
```go
// In KeyBindings struct (keys.go:33-55):
StateToggle string

// In DefaultKeyBindings() (keys.go:58-83):
StateToggle: KeyStateToggle,
```

## Partial Items

### 1. Story 11.5 Interface Extension Note
**Missing:** Explicit mention that StateActivator was recently extended with Hibernate()
**Improvement:** Add to Context: "Note: StateActivator interface was extended in Story 11.5 to include Hibernate() method for CLI access."

### 2. Approximate Line Numbers
**Missing:** Precise code insertion points
**Improvement:** Use function names as anchors instead: "Add after `projectActivatedMsg` message type", "Add case after `KeyRemove` handler"

## Recommendations

### Must Fix (Critical - 3 items)
1. Replace `strings.Title(msg.action)` with explicit string formatting
2. Fix file path from `name.go` to `project.go` in Reuse table
3. Add KeyBindings struct update to implementation guide

### Should Improve (Enhancement - 4 items)
1. Add imports list to code examples
2. Reference Story 11.5 interface extension explicitly
3. Use function-based anchors instead of line numbers
4. Add edge case test for concurrent state changes (race with 11.3 auto-activation)

### Consider (Nice to Have - 2 items)
1. Add behavioral test using teatest (Epic 9 infrastructure)
2. Add golden file for help overlay with new H keybinding

### LLM Optimization (2 items)
1. The code examples in Technical Implementation Guide are verbose - could use inline comments instead of separate explanation paragraphs
2. The Reuse vs Create table duplicates information already in Tasks section

---

## Interactive Improvement: Automatic Application

Per user request, applying ALL suggested improvements to the story file.
