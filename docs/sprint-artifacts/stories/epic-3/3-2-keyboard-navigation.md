# Story 3.2: Keyboard Navigation

**Status:** done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/model.go`, `internal/adapters/tui/keys.go` |
| **Key Dependencies** | Bubbles list (already configured in Story 3.1), existing ProjectListModel |
| **Files to Create** | None |
| **Files to Modify** | `keys.go` (add KeyEscape only), `model.go` (add Esc handler), `model_test.go` (add tests) |
| **Location** | `internal/adapters/tui/` |
| **Interfaces Used** | `list.KeyMap`, `tea.KeyMsg` |

### Quick Task Summary (4 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Add KeyEscape constant | Add `KeyEscape = "esc"` to keys.go |
| 2 | Verify Bubbles navigation | Confirm j/k already works, document boundary behavior |
| 3 | Handle Esc key in Model | Add Esc case to handleKeyMsg() |
| 4 | Add tests to model_test.go | Navigation and Esc behavior tests |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Navigation library | Bubbles list KeyMap | Already configured in Story 3.1, handles j/k natively |
| Boundary behavior | Stop at edges (no wrap) | Default Bubbles behavior, acceptable for MVP |
| Selection persistence | Preserve | Selection stays on view switches (Bubbles default) |
| Response time | <50ms | Bubbles is optimized, manual verify only |
| g/G navigation | Post-MVP | Do not implement in this story |

## Story

**As a** user,
**I want** to navigate with keyboard shortcuts,
**So that** I can efficiently browse projects.

## Acceptance Criteria

```gherkin
AC1: Given project list is displayed
     When I press 'j' or '↓'
     Then selection moves down one item
     And if at last item, stays at last (Bubbles default - no wrap)

AC2: Given project list is displayed
     When I press 'k' or '↑'
     Then selection moves up one item
     And if at first item, stays at first (Bubbles default - no wrap)

AC3: (post-MVP) Given project list is displayed
     When I press 'g'
     Then selection moves to first item

AC4: (post-MVP) Given project list is displayed
     When I press 'G'
     Then selection moves to last item

AC5: Given project list is displayed
     When I press 'q'
     Then TUI exits cleanly

AC6: Given an active prompt is displayed
     When I press 'Esc'
     Then prompt is cancelled
     And underlying view is restored

AC7: Given no active prompt is displayed
     When I press 'Esc'
     Then nothing happens (no exit, no action)

AC8: Given user navigates through the list
     Then navigation feels instant (<50ms response)

AC9: Given user navigates and switches views
     Then selection state persists during view switches
```

**Note on AC1/AC2:** Bubbles list does NOT wrap at boundaries by default. This is acceptable MVP behavior (standard list navigation).

## Tasks / Subtasks

- [x] **Task 1: Add KeyEscape constant** (AC: 6, 7)
  - [x] 1.1 Add to `keys.go` after existing KeyHelp:
    ```go
    KeyEscape = "esc"
    ```
  - [x] 1.2 Add `Escape string` field to `KeyBindings` struct
  - [x] 1.3 Add `Escape: KeyEscape` to `DefaultKeyBindings()` return
  - Note: j/k/up/down constants NOT needed - Bubbles handles these internally via DefaultKeyMap()

- [x] **Task 2: Verify Bubbles navigation works** (AC: 1, 2)
  - [x] 2.1 Confirm `list.DefaultKeyMap()` already binds j/k and arrows (Story 3.1: `project_list.go`)
  - [x] 2.2 Test boundary behavior manually: at first/last item, verify it stops (no wrap)
  - [x] 2.3 Document actual behavior in Completion Notes
  - Note: This is verification only - NO code changes needed for navigation

- [x] **Task 3: Handle Esc key in Model** (AC: 6, 7)
  - [x] 3.1 In `model.go` `handleKeyMsg()`, add Esc case after existing key handlers:
    ```go
    case KeyEscape:
        // No-op in normal mode - future stories (3.7, 3.9) will add prompt cancellation
        return m, nil
    ```
  - [x] 3.2 Note: Esc while help is showing already works via "any key closes help" logic
  - Note: KeyEscape is in same package (tui) - no import needed

- [x] **Task 4: Add navigation tests to existing model_test.go** (AC: 1, 2, 5, 6, 7, 8, 9)
  - [x] 4.1 Open existing `internal/adapters/tui/model_test.go` (file already exists with ~350 lines)
  - [x] 4.2 Add `TestModel_Navigation_JMovesDown` - verify j key moves selection
  - [x] 4.3 Add `TestModel_Navigation_KMovesUp` - verify k key moves selection
  - [x] 4.4 Add `TestModel_Navigation_BoundaryBehavior` - test at first/last item
  - [x] 4.5 Add `TestModel_Escape_NormalMode` - test Esc returns nil cmd
  - [x] 4.6 Add `TestModel_Escape_WhileHelpShowing` - test Esc closes help
  - [x] 4.7 Add `TestModel_SelectionPersistence` - verify selection after WindowSizeMsg
  - [x] 4.8 Run `make test` and verify all pass
  - [x] 4.9 Run `make lint` and verify no errors
  - [x] 4.10 Run `make build` and verify successful build
  - Note: AC5 (q quits) already tested in `TestModel_Update_QuitKey`. AC8 (<50ms) is manual verification only.

## Dev Notes

### Navigation Already Works (Story 3.1)

Story 3.1 configured `list.DefaultKeyMap()` which binds:
- `CursorUp` to 'k' and 'up' arrow
- `CursorDown` to 'j' and 'down' arrow

This story primarily:
1. Adds KeyEscape constant and Esc handling
2. Verifies and documents boundary behavior
3. Adds tests for navigation

### Boundary Behavior (No Wrap)

Bubbles list stops at boundaries by default:
- At index 0, pressing 'k': stays at 0 (no wrap to last)
- At last item, pressing 'j': stays at last (no wrap to first)

This is acceptable MVP behavior. Document actual behavior in Completion Notes.

### Testing Pattern

Use table-driven tests following existing `model_test.go` patterns:

```go
func TestModel_Navigation_JMovesDown(t *testing.T) {
    m := NewModel(mockRepo)
    m.ready = true
    m.projects = []*domain.Project{{Name: "a"}, {Name: "b"}, {Name: "c"}}
    m.projectList = components.NewProjectListModel(m.projects, 80, 24)

    // Initial selection should be 0
    if m.projectList.Index() != 0 {
        t.Error("Initial selection should be 0")
    }

    // Press j
    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
    newModel, _ := m.Update(msg)
    updated := newModel.(Model)

    if updated.projectList.Index() != 1 {
        t.Errorf("After 'j', selection should be 1, got %d", updated.projectList.Index())
    }
}
```

### Project Context Rules

- Co-locate tests: add to existing `model_test.go`
- Table-driven tests with `tests []struct{...}` pattern
- PascalCase for exported, camelCase for unexported

### Post-MVP (Do NOT Implement)

g/G navigation (AC3, AC4) is post-MVP - do not implement.

## Dev Agent Record

### Context Reference

Source docs: epics.md (lines 1124-1165), architecture.md, project-context.md, Story 3.1

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story drafting phase.

### Completion Notes List

**Implementation completed 2025-12-14 by Dev Agent (Amelia):**

1. **Task 1 (KeyEscape constant):**
   - Added `KeyEscape = "esc"` constant to `keys.go:9`
   - Added `Escape string` field to `KeyBindings` struct at `keys.go:18`
   - Added `Escape: KeyEscape` to `DefaultKeyBindings()` return at `keys.go:27`

2. **Task 2 (Bubbles navigation verification):**
   - Confirmed `list.DefaultKeyMap()` configured at `project_list.go:44`
   - Verified from `bubbles@v0.21.0/list/keys.go:37-44`: `CursorUp` binds `up`/`k`, `CursorDown` binds `down`/`j`
   - **Boundary behavior:** Bubbles list stops at edges by default (no wrap) - acceptable MVP behavior
   - Note: Bubbles also binds `g`/`G` via `GoToStart`/`GoToEnd` (post-MVP per story)

3. **Task 3 (Esc handling):**
   - Added `KeyEscape` case to `handleKeyMsg()` at `model.go:335-337`
   - Returns `nil` cmd (no-op in normal mode) - future stories will add prompt cancellation
   - Esc while help showing works via existing "any key closes help" logic

4. **Task 4 (Navigation tests):**
   - Added 8 new tests to `model_test.go`:
     - `TestModel_Navigation_JMovesDown` (lines 375-391)
     - `TestModel_Navigation_KMovesUp` (lines 393-413)
     - `TestModel_Navigation_ArrowDownMovesDown` (lines 415-431)
     - `TestModel_Navigation_ArrowUpMovesUp` (lines 433-453)
     - `TestModel_Navigation_BoundaryBehavior` (lines 455-490)
     - `TestModel_Escape_NormalMode` (lines 492-511)
     - `TestModel_Escape_WhileHelpShowing` (lines 513-526)
     - `TestModel_SelectionPersistence` (lines 528-555)
   - Added helper `createModelWithProjects()` at lines 355-373
   - Added `kb.Escape` assertion to `TestDefaultKeyBindings` at lines 249-251
   - All tests pass, lint clean, build successful

### File List

**Modified:**
- `internal/adapters/tui/keys.go` - Added KeyEscape constant, Escape field, DefaultKeyBindings update
- `internal/adapters/tui/model.go` - Added Esc handling in handleKeyMsg()
- `internal/adapters/tui/model_test.go` - Added 8 navigation tests + helper function + Escape assertion

## Change Log

| Date | Change |
|------|--------|
| 2025-12-14 | Story created with ready-for-dev status by SM Agent (Bob) in YOLO mode |
| 2025-12-14 | **Validation Review Applied:** C1: Clarified model_test.go creation. C2: Fixed Esc key handling analysis. E1-E4: Added import notes, wrapping verification, test cases, key additions context. O1-O3: Simplified Esc handling, removed post-MVP redundancy, condensed code examples. L1-L3: Optimized token usage in references, diagrams, and patterns. Reviewed by SM Agent (Bob). |
| 2025-12-14 | **Second Validation Review Applied:** C1: Removed unnecessary navigation key constants (j/k/up/down) - only KeyEscape needed. E1: Updated ACs to reflect Bubbles non-wrapping behavior. E2: Added import guidance for KeyEscape. E3: Fixed Task 4.1 - model_test.go already exists. O1: Simplified Task 2 to verification only. O2: Removed redundant AC testing. L1-L2: Condensed Dev Notes, removed redundant line references. Reviewed by SM Agent (Bob). |
| 2025-12-14 | **Implementation Complete:** All 4 tasks completed. Added KeyEscape constant, verified Bubbles navigation, added Esc handler, wrote 6 navigation tests. All tests pass, lint clean, build successful. Status: review. Implemented by Dev Agent (Amelia). |
| 2025-12-14 | **Code Review Fixes Applied:** M1: Added 2 arrow key tests (`TestModel_Navigation_ArrowDownMovesDown`, `TestModel_Navigation_ArrowUpMovesUp`) to cover AC1/AC2 arrow key requirements. M2: Added `kb.Escape` assertion to `TestDefaultKeyBindings`. L1: Corrected all test line numbers in Completion Notes. Total tests now 8. Reviewed and fixed by Dev Agent (Amelia). |
