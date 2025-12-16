# Story 3.4: Status Bar Component

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/model.go`, `internal/adapters/tui/components/` |
| **Key Dependencies** | Lipgloss styles (Story 1.6), domain.Project, domain.ProjectState (Story 1.2), ProjectListModel (Story 3.1), DetailPanelModel (Story 3.3) |
| **Files to Create** | `components/status_bar.go`, `components/status_bar_test.go` |
| **Files to Modify** | `model.go` (add status bar state + rendering + height adjustment) |
| **Location** | `internal/adapters/tui/`, `internal/adapters/tui/components/` |
| **Interfaces Used** | `domain.Project`, `domain.ProjectState` |

### Quick Task Summary (5 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Create StatusBar component | `status_bar.go` with counts and shortcuts rendering |
| 2 | Add counting logic | Exported `CalculateCounts()` function |
| 3 | Integrate into Model | Add statusBar field, initialization, height-adjusted rendering |
| 4 | Handle responsive width | Abbreviate shortcuts when terminal narrow |
| 5 | Add tests | StatusBar rendering tests, count tests, responsive tests, integration tests |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Position | Fixed at bottom | Per UX spec - always visible regardless of scroll |
| Layout | Two-line format | Line 1: counts, Line 2: shortcuts (per epics.md AC) |
| WAITING style | WaitingStyle (bold red) | Per styles.go - reserved exclusively for WAITING |
| Width threshold | < 80 columns | Below this, abbreviate shortcuts |
| Counts source | Projects slice in Model | Filter projects by state to calculate counts |
| CalculateCounts | Exported function | Called from model.go via `components.CalculateCounts()` |

## Story

**As a** user,
**I want** a persistent status bar,
**So that** I always see summary counts and available shortcuts.

## Acceptance Criteria

```gherkin
AC1: Given dashboard is displayed
     When status bar renders
     Then it shows at bottom:
       - Line 1: "│ N active │ M hibernated │ ⏸️ K WAITING │" (counts)
       - Line 2: "│ [j/k] nav [d] details [f] fav [?] help [q] quit │" (shortcuts)

AC2: Given status bar is rendered
     When layout updates
     Then status bar is always visible (fixed position)
     And does not scroll with project list

AC3: Given projects loaded
     When counts are displayed
     Then counts update in real-time when project state changes

AC4: Given WAITING count > 0
     When status bar renders
     Then "⏸️ N WAITING" displays in bold red (WaitingStyle)
     And "WAITING" section is prominent

AC5: Given WAITING count = 0
     When status bar renders
     Then WAITING section is hidden OR shows "0 waiting" dimmed

AC6: [FUTURE Story 5.4] Given user is in hibernated view
     When status bar renders
     Then shortcuts show "[h] back to active" instead of "[h] hibernated"
     NOTE: Implement placeholder only - full logic in Story 5.4

AC7: Given terminal width < 80
     When status bar renders
     Then shortcuts are abbreviated:
       - "[j/k] nav" → "[j/k]"
       - "[d] details" → "[d]"
       - "[f] fav" → "[f]"
       - etc.
```

## Tasks / Subtasks

- [x] **Task 1: Create StatusBar component** (AC: 1, 4, 5)
  - [x] 1.1 Create `internal/adapters/tui/components/status_bar.go` with:
    ```go
    package components

    import (
        "fmt"
        "strings"

        "github.com/charmbracelet/lipgloss"
        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    )
    ```
  - [x] 1.2 Define package-level styles (follow Story 3.3 detail_panel.go pattern):
    ```go
    // Keep in sync with tui/styles.go
    var (
        waitingStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("1")) // Red

        // dimStyle already declared in delegate.go, reuse or rename
        statusBarDimStyle = lipgloss.NewStyle().Faint(true)
    )
    ```
  - [x] 1.3 Define `StatusBarModel` struct:
    ```go
    type StatusBarModel struct {
        activeCount      int
        hibernatedCount  int
        waitingCount     int
        width            int
        inHibernatedView bool // Placeholder for Story 5.4
    }
    ```
  - [x] 1.4 Implement `NewStatusBarModel(width int) StatusBarModel`
  - [x] 1.5 Implement `SetCounts(active, hibernated, waiting int)` method
  - [x] 1.6 Implement `SetWidth(width int)` method
  - [x] 1.7 Implement `SetInHibernatedView(inView bool)` method (placeholder - does not affect rendering until Story 5.4)
  - [x] 1.8 Implement `View() string` rendering method:
    ```go
    func (s StatusBarModel) View() string {
        countsLine := s.renderCounts()
        shortcutsLine := s.renderShortcuts()
        return countsLine + "\n" + shortcutsLine
    }
    ```
  - [x] 1.9 Implement `renderCounts() string` helper:
    - Format with pipe separators: `"│ N active │ M hibernated │"`
    - If waitingCount > 0: append ` ⏸️ K WAITING │` with waitingStyle
    - If waitingCount == 0: hide WAITING section OR show dimmed
  - [x] 1.10 Implement `renderShortcuts() string` helper:
    - Return with pipe separators: `"│ [j/k] nav ... │"`
    - Check width threshold (80)
    - Return full shortcuts if width >= 80, abbreviated if < 80

- [x] **Task 2: Add counting logic** (AC: 3)
  - [x] 2.1 Create **exported** `CalculateCounts()` function in `status_bar.go`:
    ```go
    // CalculateCounts returns active, hibernated, and waiting counts from projects.
    // NOTE: Called from model.go via components.CalculateCounts()
    func CalculateCounts(projects []*domain.Project) (active, hibernated, waiting int) {
        for _, p := range projects {
            switch p.State {
            case domain.StateActive:
                active++
            case domain.StateHibernated:
                hibernated++
            }
            // TODO(Story-4.3): Add waiting detection when p.IsWaiting field exists
            // For now, waiting count is always 0
        }
        return
    }
    ```
  - [x] 2.2 Verify: `domain.ProjectState` exists in `internal/core/domain/state.go` (confirmed: StateActive, StateHibernated)

- [x] **Task 3: Integrate into Model** (AC: 1, 2, 3)
  - [x] 3.1 Add field to `Model` struct in `model.go`:
    ```go
    // Status bar state (Story 3.4)
    statusBar components.StatusBarModel
    ```
  - [x] 3.2 In `NewModel()`, initialize statusBar with zero width (will be set on first resize):
    ```go
    func NewModel(repo ports.ProjectRepository) Model {
        return Model{
            ready:           false,
            showHelp:        false,
            showDetailPanel: false,
            viewMode:        viewModeNormal,
            repository:      repo,
            statusBar:       components.NewStatusBarModel(0), // Width set in resizeTickMsg
        }
    }
    ```
  - [x] 3.3 In `resizeTickMsg` handler, update statusBar width AND pass adjusted height to components:
    ```go
    case resizeTickMsg:
        if m.hasPendingResize {
            // ... existing ready/dimension logic ...

            // Update status bar width
            m.statusBar.SetWidth(m.width)

            // Update component dimensions with adjusted height (subtract 2 for status bar)
            if len(m.projects) > 0 {
                contentHeight := m.height - 2  // Reserve 2 lines for status bar
                m.projectList.SetSize(m.width, contentHeight)
                m.detailPanel.SetSize(m.width, contentHeight)
            }
        }
        return m, nil
    ```
  - [x] 3.4 In `ProjectsLoadedMsg` handler, update status bar counts:
    ```go
    if len(m.projects) > 0 {
        m.projectList = components.NewProjectListModel(m.projects, m.width, m.height-2)
        m.detailPanel = components.NewDetailPanelModel(m.width, m.height-2)
        m.detailPanel.SetProject(m.projectList.SelectedProject())
        m.detailPanel.SetVisible(m.showDetailPanel)

        // Update status bar counts
        active, hibernated, waiting := components.CalculateCounts(m.projects)
        m.statusBar.SetCounts(active, hibernated, waiting)
    }
    ```
  - [x] 3.5 Refactor `renderDashboard()` to include status bar:
    ```go
    func (m Model) renderDashboard() string {
        contentHeight := m.height - 2  // Reserve 2 lines for status bar
        mainContent := m.renderMainContent(contentHeight)
        statusBar := m.statusBar.View()
        return lipgloss.JoinVertical(lipgloss.Left, mainContent, statusBar)
    }
    ```
  - [x] 3.6 Extract `renderMainContent(height int) string` helper from existing renderDashboard logic:
    ```go
    func (m Model) renderMainContent(height int) string {
        // Handle height < 30 case - show hint when panel closed (AC4 from Story 3.3)
        if height < 28 && !m.showDetailPanel {  // 30-2 for status bar
            hint := DimStyle.Render("Press [d] for details")
            return m.projectList.View() + "\n" + hint
        }

        if !m.showDetailPanel {
            return m.projectList.View()
        }

        // Split layout: list (60%) | detail (40%)
        listWidth := int(float64(m.width) * 0.6)
        detailWidth := m.width - listWidth - 1

        projectList := m.projectList
        projectList.SetSize(listWidth, height)

        detailPanel := m.detailPanel
        detailPanel.SetSize(detailWidth, height)

        return lipgloss.JoinHorizontal(lipgloss.Top, projectList.View(), detailPanel.View())
    }
    ```

- [x] **Task 4: Handle responsive width** (AC: 7)
  - [x] 4.1 Define shortcut string constants in `status_bar.go`:
    ```go
    const (
        // Full shortcuts (width >= 80) - with pipe separators
        shortcutsFull = "│ [j/k] nav [d] details [f] fav [?] help [q] quit │"

        // Abbreviated shortcuts (width < 80)
        shortcutsAbbrev = "│ [j/k] [d] [f] [?] [q] │"

        // Future Story 5.4: Hibernated view shortcuts
        shortcutsHibernatedFull   = "│ [j/k] nav [h] back to active [?] help [q] quit │"
        shortcutsHibernatedAbbrev = "│ [j/k] [h] [?] [q] │"
    )
    ```
  - [x] 4.2 Implement `renderShortcuts()`:
    ```go
    func (s StatusBarModel) renderShortcuts() string {
        // TODO(Story-5.4): Use hibernated shortcuts when s.inHibernatedView is true
        if s.width >= 80 {
            return shortcutsFull
        }
        return shortcutsAbbrev
    }
    ```

- [x] **Task 5: Add tests** (AC: all)
  - [x] 5.1 Create `internal/adapters/tui/components/status_bar_test.go`
  - [x] 5.2 Add `TestStatusBar_View_BasicCounts` - verify counts render with pipes
  - [x] 5.3 Add `TestStatusBar_View_WaitingHighlighted` - verify WAITING uses waitingStyle
  - [x] 5.4 Add `TestStatusBar_View_WaitingZero` - verify WAITING hidden/dimmed when 0
  - [x] 5.5 Add `TestStatusBar_View_ShortcutsFull` - verify full shortcuts at width >= 80
  - [x] 5.6 Add `TestStatusBar_View_ShortcutsAbbreviated` - verify abbreviated at width < 80
  - [x] 5.7 Add `TestStatusBar_SetCounts` - verify counts update correctly
  - [x] 5.8 Add `TestStatusBar_SetWidth` - verify width updates correctly
  - [x] 5.9 Add `TestCalculateCounts` - verify counting logic with various project states
  - [x] 5.10 Add `TestCalculateCounts_MixedStates` - active + hibernated mixed
  - [x] 5.11 Add to `model_test.go`:
    - `TestModel_StatusBarIntegration` - verify status bar renders in dashboard
    - `TestModel_StatusBarCountsUpdate` - verify counts update after project load
    - `TestModel_StatusBarHeightReservation` - verify 2 lines reserved for status bar
    - `TestModel_StatusBarWidthUpdate` - verify width updates on resize
  - [x] 5.12 Run `make test` and verify all pass
  - [x] 5.13 Run `make lint` and verify no errors
  - [x] 5.14 Run `make build` and verify successful build

## Dev Notes

### Required Imports for status_bar.go

```go
package components

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)
```

### Style Pattern (Follow Story 3.3 detail_panel.go)

Components cannot import `tui` package (circular dependency). Define package-level styles:

```go
// Keep in sync with tui/styles.go
var (
    waitingStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("1")) // Red - matches WaitingStyle

    // Note: dimStyle may already be declared in delegate.go
    // If so, reuse it or use a different name
    statusBarDimStyle = lipgloss.NewStyle().Faint(true)
)
```

### Status Bar Layout

```
┌──────────────────────────────────────────────────────────────┐
│ Project List / Detail Panel content                          │
│ ...                                                          │
├──────────────────────────────────────────────────────────────┤
│ 5 active │ 2 hibernated │ ⏸️ 1 WAITING                       │  <- Line 1 (counts)
│ [j/k] nav [d] details [f] fav [?] help [q] quit              │  <- Line 2 (shortcuts)
└──────────────────────────────────────────────────────────────┘
```

### Height Adjustment (CRITICAL)

Status bar takes 2 lines. ALL components must receive adjusted height:

```go
contentHeight := m.height - 2  // Reserve 2 lines for status bar

// In resizeTickMsg:
m.projectList.SetSize(m.width, contentHeight)
m.detailPanel.SetSize(m.width, contentHeight)

// In ProjectsLoadedMsg:
m.projectList = components.NewProjectListModel(m.projects, m.width, contentHeight)
m.detailPanel = components.NewDetailPanelModel(m.width, contentHeight)
```

### Counts Rendering with Pipes

```go
func (s StatusBarModel) renderCounts() string {
    parts := []string{
        fmt.Sprintf("%d active", s.activeCount),
        fmt.Sprintf("%d hibernated", s.hibernatedCount),
    }

    if s.waitingCount > 0 {
        waitingText := waitingStyle.Render(fmt.Sprintf("⏸️ %d WAITING", s.waitingCount))
        parts = append(parts, waitingText)
    }

    return "│ " + strings.Join(parts, " │ ") + " │"
}
```

### WAITING Logic Placeholder

```go
// In CalculateCounts():
// TODO(Story-4.3): Add waiting detection when p.IsWaiting field exists
// For now, waiting count is always 0
```

When Story 4.3 is implemented:
```go
if p.IsWaiting {
    waiting++
}
```

### Responsive Width Thresholds

| Width | Behavior |
|-------|----------|
| >= 80 | Full shortcuts with descriptions |
| < 80 | Abbreviated shortcuts (keys only) |
| < 60 | Minimal view warning (handled in renderTooSmallView) |

### Component Pattern Consistency

StatusBar follows same pattern as ProjectListModel and DetailPanelModel:
- `NewStatusBarModel(width int)` constructor
- `SetSize()` / `SetWidth()` for responsive layout
- `View() string` for rendering
- No `Init()` method needed (consistent with other components)
- No `Update()` method needed (state managed by Model)

### Reuse from Previous Stories

**From Story 1.6 (styles.go):**
- `WaitingStyle` reference (reproduce in component as `waitingStyle`)
- `DimStyle` reference (reproduce as `statusBarDimStyle`)

**From Story 3.1 (project_list.go):**
- `ProjectListModel.SetSize()` - now receives adjusted height

**From Story 3.3 (detail_panel.go):**
- Package-level style variable pattern (avoid import cycle)
- `renderDashboard()` pattern - extend with status bar

### Domain Types (Verified)

From `internal/core/domain/state.go`:
```go
type ProjectState int

const (
    StateActive ProjectState = iota
    StateHibernated
)
```

From `internal/core/domain/project.go`:
```go
type Project struct {
    // ...
    State ProjectState // Active or Hibernated
    // ...
}
```

### Integration with Future Stories

**Story 4.3 (Agent Waiting Detection):**
- Will add `IsWaiting` field to domain.Project
- Update `CalculateCounts()` to count waiting projects

**Story 5.4 (Hibernated Projects View):**
- `SetInHibernatedView(true)` will switch to hibernated shortcut set
- Shows "[h] back to active" instead of standard shortcuts

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 3.4 requirements - lines 1223-1260)
- docs/architecture.md (TUI framework, Bubble Tea patterns)
- docs/project-context.md (Go conventions, testing rules)
- internal/adapters/tui/model.go (Current TUI after Stories 3.1-3.3)
- internal/adapters/tui/styles.go (Lipgloss styles)
- internal/adapters/tui/components/detail_panel.go (Component and style patterns)
- internal/core/domain/state.go (ProjectState enum)
- internal/core/domain/project.go (Project struct with State field)
- docs/sprint-artifacts/stories/epic-3/3-3-detail-panel-component.md (Previous story patterns)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story drafting phase.

### Completion Notes List

- **Task 1:** Created `StatusBarModel` struct with `activeCount`, `hibernatedCount`, `waitingCount`, `width`, and `inHibernatedView` fields. Implemented `NewStatusBarModel()`, `SetCounts()`, `SetWidth()`, `SetInHibernatedView()`, `View()`, `renderCounts()`, and `renderShortcuts()` methods. Used `statusBarWaitingStyle` for bold red WAITING display (AC4). Hidden WAITING section when count is 0 (AC5).
- **Task 2:** Implemented exported `CalculateCounts()` function that counts active and hibernated projects by iterating over `domain.Project.State`. Waiting count is always 0 (placeholder for Story 4.3).
- **Task 3:** Integrated StatusBar into Model by adding `statusBar` field, initializing in `NewModel()`, updating width in `resizeTickMsg`, updating counts in `ProjectsLoadedMsg`, and refactoring `renderDashboard()` with `renderMainContent()` helper. All components now receive `height-2` to reserve 2 lines for status bar (AC2).
- **Task 4:** Implemented responsive shortcuts with `widthThreshold=80`. Full shortcuts with descriptions at width >= 80, abbreviated shortcuts (keys only) at width < 80 (AC7). Placeholder constants for Story 5.4 hibernated view shortcuts (prefixed with underscore to avoid lint errors).
- **Task 5:** Created 13 unit tests in `status_bar_test.go` and 4 integration tests in `model_test.go`. All tests pass. Lint and build pass.

### File List

**Created:**
- `internal/adapters/tui/components/status_bar.go` - StatusBarModel component with CalculateCounts()
- `internal/adapters/tui/components/status_bar_test.go` - StatusBar tests (16 tests)
- `docs/sprint-artifacts/validations/epic-3/validation-report-3-4-2025-12-16.md` - SM validation report

**Modified:**
- `internal/adapters/tui/model.go`:
  - Added `statusBar components.StatusBarModel` field
  - Initialized in `NewModel()`
  - Updated counts in `ProjectsLoadedMsg`
  - Updated width in `resizeTickMsg`
  - Passed adjusted height (m.height-2) to all components
  - Refactored `renderDashboard()` to include status bar
  - Extracted `renderMainContent()` helper
  - **[Code Review Fix]** Added status bar to empty view (AC2 compliance)
- `internal/adapters/tui/model_test.go` - Added 5 status bar integration tests
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status

## Change Log

| Date | Change |
|------|--------|
| 2025-12-14 | Story created with ready-for-dev status by SM Agent (Bob) in YOLO mode. |
| 2025-12-16 | **Validation Review Applied:** (C1) Fixed CalculateCounts to be exported for model.go access. (C2) Added explicit import section for status_bar.go. (C3) Added StatusBarModel initialization in NewModel(). (E1) Added height adjustment instructions for resizeTickMsg and ProjectsLoadedMsg. (E2) Documented no Init() needed (consistent with other components). (E3) Added TestCalculateCounts_MixedStates test. (E4) Added pipe separators to shortcut constants. (O1) Used TODO(Story-X.X) format for placeholders. (O2) Added explicit Story 3.3 pattern reference. (O3) Marked AC6 as [FUTURE Story 5.4]. (L1-L3) Consolidated Dev Notes and removed redundant sections. Reviewed by SM Agent (Bob). |
| 2025-12-16 | **Implementation Complete:** All 5 tasks completed. Created status_bar.go with StatusBarModel, CalculateCounts(), responsive shortcuts. Integrated into model.go with height adjustment. Added 17 tests (13 unit + 4 integration). All tests pass, lint clean, build successful. Story ready for review. |
| 2025-12-16 | **Code Review Fixes Applied:** (H1) Added status bar to empty view for AC2 compliance. (M1) Updated File List with validation report and sprint-status.yaml. (M2) Added TestStatusBar_SetInHibernatedView test. (L2) Added TestStatusBar_View_ZeroCounts test. (L3) Added TestStatusBar_View_ShortcutsBoundary test for width=80. Total tests now 21 (16 unit + 5 integration). All tests pass. Reviewed by Dev Agent (Amelia). |
