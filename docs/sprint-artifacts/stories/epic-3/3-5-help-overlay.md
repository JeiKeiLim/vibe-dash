# Story 3.5: Help Overlay

**Status:** dev-complete

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/views.go`, `internal/adapters/tui/model.go` |
| **Key Dependencies** | Lipgloss styles (Story 1.6), KeyBindings (keys.go), boxStyle/titleStyle/hintStyle (views.go) |
| **Files to Modify** | `views.go` (enhance renderHelpOverlay), `keys.go` (add missing key constants only) |
| **Files to Create** | `views_test.go` (help overlay rendering tests) |
| **Location** | `internal/adapters/tui/` |
| **Interfaces Used** | None (pure view function) |

### Quick Task Summary (3 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Add missing key constants | Add only NEW constants to `keys.go` (KeyDetail already exists) |
| 2 | Enhance help overlay content | Full keyboard shortcuts organized by category |
| 3 | Add tests | Help overlay rendering tests in `views_test.go` |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Overlay positioning | Centered modal | Per UX spec - overlays dashboard content |
| Categories | Navigation, Actions, Views, General | Organized by function for discoverability |
| Close behavior | Any key closes | Per AC2 - already implemented in model.go `handleKeyMsg` |
| Box width | 46 characters | Longest line (31) + padding (4) + border (2) + buffer (9) |
| Help key | '?' (already implemented) | Per FR65 and Story 1.5 |
| Test package | `package tui` | Required to test unexported `renderHelpOverlay` function |

## Story

**As a** user,
**I want** to see all keyboard shortcuts,
**So that** I can learn available actions.

## Acceptance Criteria

```gherkin
AC1: Given dashboard is displayed
     When I press '?'
     Then help overlay appears with:
       - Title: "KEYBOARD SHORTCUTS"
       - Navigation section: j/k/arrow keys
       - Actions section: d (details), f (fav), n (notes), x (remove), a (add), r (refresh)
       - Views section: h (hibernated projects)
       - General section: ?, q, Esc
       - Hint: "Press any key to close"

AC2: Given help overlay is visible
     When I press any key
     Then help overlay closes
     And previous view is restored

AC3: Given help overlay is visible
     When I press '?' key specifically
     Then help overlay closes (toggle behavior)
```

## Tasks / Subtasks

- [x] **Task 1: Add missing key constants to keys.go** (AC: 1)
  - [x] 1.1 Add all missing key constants (KeyDetail already exists - DO NOT re-add):
    ```go
    // Key binding constants for the TUI.
    // These define the keyboard shortcuts used throughout the application.
    const (
        // General (existing)
        KeyQuit      = "q"
        KeyForceQuit = "ctrl+c"
        KeyHelp      = "?"
        KeyEscape    = "esc"
        KeyDetail    = "d"  // Already exists - DO NOT duplicate

        // Navigation (NEW)
        KeyDown      = "j"
        KeyDownArrow = "down"
        KeyUp        = "k"
        KeyUpArrow   = "up"

        // Actions (NEW)
        KeyFavorite  = "f"
        KeyNotes     = "n"
        KeyRemove    = "x"
        KeyAdd       = "a"
        KeyRefresh   = "r"

        // Views (NEW)
        KeyHibernated = "h"
    )
    ```
  - [x] 1.2 Add new fields to KeyBindings struct (preserve existing fields):
    ```go
    type KeyBindings struct {
        // General (existing)
        Quit      string
        ForceQuit string
        Help      string
        Escape    string
        Detail    string  // Already exists

        // Navigation (NEW)
        Down      string
        DownArrow string
        Up        string
        UpArrow   string

        // Actions (NEW)
        Favorite  string
        Notes     string
        Remove    string
        Add       string
        Refresh   string

        // Views (NEW)
        Hibernated string
    }
    ```
  - [x] 1.3 Update DefaultKeyBindings() to include new fields

- [x] **Task 2: Enhance renderHelpOverlay in views.go** (AC: 1)
  - [x] 2.1 Update help overlay content with full shortcuts organized by category:
    ```go
    func renderHelpOverlay(width, height int) string {
        title := titleStyle.Render("KEYBOARD SHORTCUTS")

        // Unicode arrows: \u2193 renders as ↓, \u2191 renders as ↑
        content := strings.Join([]string{
            "",
            "Navigation",
            "j/\u2193     Move down",
            "k/\u2191     Move up",
            "",
            "Actions",
            "d        Toggle detail panel",
            "f        Toggle favorite",
            "n        Edit notes",
            "x        Remove project",
            "a        Add project",
            "r        Refresh/rescan",
            "",
            "Views",
            "h        View hibernated projects",
            "",
            "General",
            "?        Show this help",
            "q        Quit",
            "Esc      Cancel/close",
            "",
            hintStyle.Render("Press any key to close"),
            "",
        }, "\n")

        box := boxStyle.
            Width(46). // Longest line (31) + padding (4) + border (2) + buffer (9)
            Render(content)

        // Add title to the border (existing pattern from renderEmptyView)
        lines := strings.Split(box, "\n")
        if len(lines) > 0 {
            topBorder := lines[0]
            titleWithDash := fmt.Sprintf("\u2500 %s ", title)

            if len(topBorder) > 3 {
                lines[0] = string(topBorder[0]) + titleWithDash + topBorder[len(titleWithDash)+1:]
            }
            box = strings.Join(lines, "\n")
        }

        return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
    }
    ```
  - [x] 2.2 No new imports needed - views.go already imports `strings`, `fmt`, and `lipgloss`

- [x] **Task 3: Add tests** (AC: all)
  - [x] 3.1 Create `views_test.go` with `package tui` (NOT `package tui_test` - needed to test unexported functions):
    ```go
    package tui

    import (
        "strings"
        "testing"
    )
    ```
  - [x] 3.2 Add help overlay content tests:
    - [x] `TestRenderHelpOverlay_ContainsAllSections` - verify Navigation, Actions, Views, General sections present
    - [x] `TestRenderHelpOverlay_ContainsNavigationShortcuts` - verify j/k/↓/↑ listed
    - [x] `TestRenderHelpOverlay_ContainsActionShortcuts` - verify d/f/n/x/a/r listed
    - [x] `TestRenderHelpOverlay_ContainsViewShortcuts` - verify h listed
    - [x] `TestRenderHelpOverlay_ContainsGeneralShortcuts` - verify ?/q/Esc listed
    - [x] `TestRenderHelpOverlay_ContainsCloseHint` - verify "Press any key to close" present
  - [x] 3.3 Verify existing help close behavior tests in `model_test.go`:
    - [x] Confirm `TestModel_HelpToggle` exists and covers toggle behavior
    - [x] Add `TestModel_HelpCloses_OnAnyKey` if not present - verify help closes on non-'?' keys
  - [x] 3.4 Run verification commands:
    ```bash
    make test   # All tests pass
    make lint   # No lint errors
    make build  # Successful build
    ```

**NOTE:** Help close behavior (AC2, AC3) is already implemented in `model.go` `handleKeyMsg` function. Verify by running `make run` and pressing '?' to toggle help, then any key to close.

## Dev Notes

### Current State Analysis

**Help Overlay (views.go:59-90):** Currently shows only General section with ?, q, Ctrl+C. Needs expansion per AC1.

**Help Close (model.go `handleKeyMsg`):** Already implemented - any key closes help, '?' toggles. Satisfies AC2 and AC3.

**Key Constants (keys.go):** KeyDetail already exists. Add only missing constants for navigation/actions/views.

### Shortcut Categories (Per AC1)

| Category | Shortcuts | Functional Status |
|----------|-----------|-------------------|
| Navigation | j/k, ↓/↑ | Implemented in Story 3.2 |
| Actions | d | Implemented in Story 3.3 |
| Actions | f, n, x, a, r | Future stories (non-functional until implemented) |
| Views | h | Future Story 5.4 (non-functional until implemented) |
| General | ?, q, Esc | Implemented |

**NOTE:** Shortcuts for future stories (f, n, x, a, r, h) are documented for discoverability but will not function until their respective stories are implemented.

### Reuse from Previous Stories

**From views.go (Story 1.5):**
- `boxStyle`, `titleStyle`, `hintStyle` - existing styles
- Title injection pattern in `renderEmptyView` - copy for help overlay

**From model.go (Story 1.5):**
- `handleKeyMsg` help close logic - already complete

### Testing Pattern

Use `package tui` (not `package tui_test`) to access unexported `renderHelpOverlay`:

```go
func TestRenderHelpOverlay_ContainsAllSections(t *testing.T) {
    result := renderHelpOverlay(100, 40)

    sections := []string{"Navigation", "Actions", "Views", "General"}
    for _, section := range sections {
        if !strings.Contains(result, section) {
            t.Errorf("Expected help overlay to contain section %q", section)
        }
    }
}
```

### Project Context Compliance

Per `docs/project-context.md`:
- Co-locate tests: `views_test.go` next to `views.go`
- Table-driven tests where appropriate
- Run `make fmt`, `make lint`, `make test` before commit

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 3.5 requirements - lines 1264-1315)
- docs/architecture.md (TUI framework patterns)
- docs/project-context.md (Go conventions, testing rules)
- internal/adapters/tui/views.go (Current renderHelpOverlay implementation - lines 58-90)
- internal/adapters/tui/model.go (Help toggle and close behavior - handleKeyMsg function)
- internal/adapters/tui/keys.go (Current key constants - KeyDetail already exists)
- internal/adapters/tui/styles.go (Lipgloss styles)
- docs/sprint-artifacts/stories/epic-3/3-4-status-bar-component.md (Previous story patterns)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story drafting phase.

### Completion Notes List

- All tasks completed successfully
- Added 11 new key constants (navigation, actions, views)
- Enhanced `renderHelpOverlay` with full keyboard shortcuts organized by category
- Created 8 new tests in `views_test.go` covering all AC requirements
- Updated existing `TestModel_View_HelpOverlay` test to verify new content
- Verified existing help close behavior tests in `model_test.go` cover AC2/AC3
- All tests pass, no lint errors, build successful

### File List

**Modified:**
- `internal/adapters/tui/keys.go` - Added 11 new key constants (KeyDown, KeyDownArrow, KeyUp, KeyUpArrow, KeyFavorite, KeyNotes, KeyRemove, KeyAdd, KeyRefresh, KeyHibernated) and corresponding KeyBindings struct fields
- `internal/adapters/tui/views.go` - Enhanced `renderHelpOverlay` function with full keyboard shortcuts organized by Navigation, Actions, Views, General categories
- `internal/adapters/tui/model_test.go` - Updated `TestModel_View_HelpOverlay` test to verify new content sections
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status tracking

**Created:**
- `internal/adapters/tui/views_test.go` - 9 tests: `TestRenderHelpOverlay_ContainsAllSections`, `TestRenderHelpOverlay_ContainsNavigationShortcuts`, `TestRenderHelpOverlay_ContainsActionShortcuts`, `TestRenderHelpOverlay_ContainsViewShortcuts`, `TestRenderHelpOverlay_ContainsGeneralShortcuts`, `TestRenderHelpOverlay_ContainsCloseHint`, `TestRenderHelpOverlay_ContainsTitle`, `TestRenderHelpOverlay_CenteredInTerminal`, `TestRenderHelpOverlay_EdgeCase_ZeroDimensions`
- `docs/sprint-artifacts/validations/epic-3/validation-report-3-5-2025-12-16.md` - Story validation report

## Change Log

| Date | Change |
|------|--------|
| 2025-12-16 | Story created with ready-for-dev status by SM Agent (Bob) in YOLO mode. |
| 2025-12-16 | **Validation Review Applied:** (C1) Fixed Task 1 to clarify KeyDetail already exists - avoid duplicate. (C2) Restructured KeyBindings task to preserve existing fields. (C3) Added box width calculation breakdown. (C4) Specified `package tui` requirement for views_test.go. (E1) Removed hardcoded line numbers, referenced function names. (E2) Added explicit note that no new imports needed. (E3) Added test package declaration guidance. (E4) Added verification step for help close behavior. (E5) Updated testing pattern reference. (O1) Consolidated key constants into single task. (O2) Added Unicode render clarification. (O3) Removed redundant verification task - replaced with NOTE. (L1) Removed duplicate code from Dev Notes. (L2) Eliminated verbose Task 3 that did nothing. (L3) Clarified future shortcuts are non-functional placeholders. Reduced from 4 tasks to 3 tasks. Reviewed by SM Agent (Bob). |
| 2025-12-16 | **Implementation Complete:** All 3 tasks completed. Added 11 key constants, enhanced help overlay with 4 categories (Navigation, Actions, Views, General), created 8 new tests in views_test.go. All tests pass, lint clean, build successful. Implemented by Dev Agent (Amelia). |
| 2025-12-16 | **Code Review Fixes Applied:** (L1) Fixed `TestRenderHelpOverlay_ContainsViewShortcuts` to use specific pattern "h        View hibernated" instead of just "h". (L2) Fixed box width comment math: 32+4+2+8=46. (L3) Added `TestRenderHelpOverlay_EdgeCase_ZeroDimensions` test. (M2) Updated File List to document sprint-status.yaml and validation-report changes. Reviewed by Dev Agent (Amelia). |
