# Story 3.3: Detail Panel Component

**Status:** done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/model.go`, `internal/adapters/tui/components/` |
| **Key Dependencies** | Lipgloss styles (Story 1.6), domain.Project, ProjectListModel (Story 3.1), KeyBindings (Story 3.2) |
| **Files to Create** | `components/detail_panel.go`, `components/detail_panel_test.go` |
| **Files to Modify** | `domain/project.go` (Task 0: add Confidence, DetectionReasoning), `persistence/sqlite/repository.go` (Task 0: map fields), `model.go` (add detail panel state + toggle), `keys.go` (add KeyDetail) |
| **Location** | `internal/adapters/tui/`, `internal/adapters/tui/components/` |
| **Interfaces Used** | `domain.Project`, `domain.Stage`, `domain.Confidence` |

### Quick Task Summary (6 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 0 | Add Confidence/DetectionReasoning to domain.Project | Prerequisite: extend domain entity |
| 1 | Create DetailPanel component | `detail_panel.go` with rendering logic |
| 2 | Add KeyDetail constant and 'd' handling | Update `keys.go` and `model.go` |
| 3 | Update Model for detail panel state | Add `showDetailPanel` field and toggle logic |
| 4 | Integrate into dashboard layout | Split-pane rendering in `renderDashboard()` |
| 5 | Add tests | Detail panel rendering tests, toggle behavior tests |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Framework versions | Per go.mod | Uses existing Bubble Tea/Lipgloss versions from go.mod |
| Panel position | Right side | Standard TUI pattern (list left, detail right) |
| Panel width | 40% of terminal width | Balanced view with project list |
| Toggle key | 'd' | Per epics.md AC specification |
| Height threshold | < 30 rows | Below this, panel closed by default |
| Border style | BorderStyle (square corners) | Per styles.go definition |

## Story

**As a** user,
**I want** to see detailed information about the selected project,
**So that** I understand detection reasoning and project context.

## Acceptance Criteria

```gherkin
AC1: Given a project is selected
     When detail panel is visible
     Then I see:
       - Panel title with project name
       - Path field showing full canonical path
       - Method field (e.g., "Speckit")
       - Stage field (e.g., "Plan")
       - Confidence field (Certain/Likely/Uncertain)
       - Detection reasoning text
       - Notes field (showing notes or "(none)")
       - Added date (formatted date)
       - Last Active time (relative time format)

AC2: Given dashboard is displayed
     When I press 'd'
     Then detail panel toggles visibility
     And layout adjusts smoothly (no flicker)

AC3: Given detail panel is closed
     When toggled open
     Then project list shrinks to accommodate panel
     And selected project details are displayed

AC4: Given terminal height < 30 rows
     When dashboard loads
     Then detail panel is closed by default
     And hint shows at bottom of project list: "Press [d] for details"

AC5: Given terminal height >= 35 rows
     When dashboard loads
     Then detail panel is open by default

AC6: Given terminal height 30-34 rows
     When dashboard loads
     Then detail panel is closed by default
     (User can still toggle with 'd')

AC7: Given project has no notes
     When detail panel is visible
     Then Notes field shows: "(none)"

AC8: Given detection confidence is Uncertain (ConfidenceUncertain)
     When detail panel is visible
     Then Confidence shows "Uncertain" with UncertainStyle
     And Detection reasoning explains the uncertainty
```

## Tasks / Subtasks

- [x] **Task 0: Add Confidence and DetectionReasoning to domain.Project** (PREREQUISITE)
  - [x] 0.1 Edit `internal/core/domain/project.go`, add fields to Project struct:
    ```go
    Confidence         Confidence // Detection confidence level (FR12)
    DetectionReasoning string     // Human-readable detection explanation (FR11, FR26)
    ```
  - [x] 0.2 Edit `internal/adapters/persistence/sqlite/repository.go`:
    - In `rowToProject()` function (~line 344-358), add mapping:
      ```go
      // After line 351 (CurrentStage):
      confidence, _ := domain.ParseConfidence(row.Confidence.String)

      // In return statement, add:
      Confidence:         confidence,
      DetectionReasoning: row.DetectionReasoning.String,
      ```
  - [x] 0.3 Run `make test` to verify no regressions
  - [x] 0.4 Run `make lint` to verify no errors
  - Note: Database schema already has these columns; this task maps them to domain entity

- [x] **Task 1: Create DetailPanel component** (AC: 1, 7, 8)
  - [x] 1.1 Create `internal/adapters/tui/components/detail_panel.go` with `package components`
  - [x] 1.2 Define `DetailPanelModel` struct:
    ```go
    type DetailPanelModel struct {
        project *domain.Project
        width   int
        height  int
        visible bool
    }
    ```
  - [x] 1.3 Implement `NewDetailPanelModel(width, height int) DetailPanelModel`
  - [x] 1.4 Implement `SetProject(project *domain.Project)` to update displayed project
  - [x] 1.5 Implement `SetSize(width, height int)` for responsive layout
  - [x] 1.6 Implement `SetVisible(visible bool)` to show/hide panel
  - [x] 1.7 Implement `IsVisible() bool` getter
  - [x] 1.8 Implement `View() string` rendering method:
    - Use BorderStyle for panel border
    - Title: "DETAILS: {project_name}" using titleStyle
    - Format each field with label-value alignment
    - Use shared/timeformat.FormatRelativeTime for Last Active
    - Handle nil project gracefully (return empty or placeholder)
  - [x] 1.9 Handle edge cases:
    - Notes empty → show "(none)"
    - Confidence = ConfidenceUncertain → apply UncertainStyle
    - DetectionReasoning empty → show "No detection reasoning available"

- [x] **Task 2: Add KeyDetail constant and 'd' handling** (AC: 2)
  - [x] 2.1 Add to `keys.go`:
    ```go
    KeyDetail = "d"
    ```
  - [x] 2.2 Add `Detail string` field to `KeyBindings` struct
  - [x] 2.3 Add `Detail: KeyDetail` to `DefaultKeyBindings()` return
  - [x] 2.4 In `model.go` `handleKeyMsg()`, add case for KeyDetail:
    ```go
    case KeyDetail:
        m.showDetailPanel = !m.showDetailPanel
        return m, nil
    ```

- [x] **Task 3: Update Model for detail panel state** (AC: 2, 3, 4, 5, 6)
  - [x] 3.1 Add fields to `Model` struct:
    ```go
    showDetailPanel bool
    detailPanel     components.DetailPanelModel
    ```
  - [x] 3.2 In `NewModel()`, initialize `showDetailPanel = false` (default)
  - [x] 3.3 Add `shouldShowDetailPanelByDefault(height int) bool` helper:
    ```go
    func shouldShowDetailPanelByDefault(height int) bool {
        return height >= 35
    }
    ```
  - [x] 3.4 In resizeTickMsg handler, set initial visibility on first ready:
    ```go
    // In resizeTickMsg handler, track if this is first ready
    wasReady := m.ready
    m.ready = true
    // ... existing dimension updates ...

    // Set initial detail panel visibility only on first ready
    if !wasReady {
        m.showDetailPanel = shouldShowDetailPanelByDefault(m.height)
    }
    ```
  - [x] 3.5 In ProjectsLoadedMsg handler, sync detail panel with selected project:
    ```go
    if len(m.projects) > 0 {
        m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
        m.detailPanel.SetProject(m.projectList.SelectedProject())
    }
    ```
  - [x] 3.6 When project list selection changes (in Update), update detail panel:
    ```go
    m.detailPanel.SetProject(m.projectList.SelectedProject())
    ```

- [x] **Task 4: Integrate into dashboard layout** (AC: 1, 2, 3, 4)
  - [x] 4.1 Update `renderDashboard()` in `model.go`:
    ```go
    func (m Model) renderDashboard() string {
        if !m.showDetailPanel {
            // Full-width project list
            return m.projectList.View()
        }

        // Split layout: list (60%) | detail (40%)
        listWidth := int(float64(m.width) * 0.6)
        detailWidth := m.width - listWidth - 1 // -1 for separator

        // Update component sizes
        m.projectList.SetSize(listWidth, m.height)
        m.detailPanel.SetSize(detailWidth, m.height)

        // Render side by side
        listView := m.projectList.View()
        detailView := m.detailPanel.View()

        return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
    }
    ```
  - [x] 4.2 Import `lipgloss` in model.go if not already imported
  - [x] 4.3 Handle height < 30 case - show hint at bottom of project list view:
    ```go
    // In renderDashboard(), when height < 30 and panel closed:
    if m.height < 30 && !m.showDetailPanel {
        hint := DimStyle.Render("Press [d] for details")
        return m.projectList.View() + "\n" + hint
    }
    ```
  - [x] 4.4 Ensure detail panel updates when selection changes via key navigation

- [x] **Task 5: Add tests** (AC: all)
  - [x] 5.1 Create `internal/adapters/tui/components/detail_panel_test.go`
  - [x] 5.2 Add `TestDetailPanel_View_BasicFields` - verify all fields render
  - [x] 5.3 Add `TestDetailPanel_View_EmptyNotes` - verify "(none)" shown
  - [x] 5.4 Add `TestDetailPanel_View_UncertainConfidence` - verify style applied
  - [x] 5.5 Add `TestDetailPanel_View_NilProject` - verify graceful handling
  - [x] 5.6 Add `TestDetailPanel_SetSize` - verify dimensions update
  - [x] 5.7 Add `TestDetailPanel_Visibility` - verify show/hide behavior
  - [x] 5.8 Add to `model_test.go`:
    - `TestModel_DetailPanelToggle` - verify 'd' key toggles panel
    - `TestModel_DetailPanelDefaultState_Height29` - verify closed (< 30)
    - `TestModel_DetailPanelDefaultState_Height30` - verify closed (30-34 range)
    - `TestModel_DetailPanelDefaultState_Height34` - verify closed (30-34 range)
    - `TestModel_DetailPanelDefaultState_Height35` - verify open (>= 35)
    - `TestModel_DetailPanelDefaultState_Height50` - verify open (>= 35)
    - `TestModel_DetailPanelUpdateOnSelection` - verify panel updates when selection changes
    - `TestModel_DetailPanelHint_ShortTerminal` - verify hint shows when height < 30 and panel closed
  - [x] 5.9 Add KeyDetail assertion to `TestDefaultKeyBindings` in `model_test.go`
  - [x] 5.10 Run `make test` and verify all pass
  - [x] 5.11 Run `make lint` and verify no errors
  - [x] 5.12 Run `make build` and verify successful build

## Dev Notes

### Panel Layout Design

```
┌─ DETAILS: client-bravo ─────────────────────┐
│ Path:       /home/user/projects/client-bravo │
│ Method:     Speckit                          │
│ Stage:      Plan                             │
│ Confidence: Certain                          │
│ Detection:  plan.md exists, no tasks.md      │
│ Notes:      Waiting on client API specs      │
│ Added:      2025-12-01                        │
│ Last Active: 2h ago                          │
└──────────────────────────────────────────────┘
```

### Field Rendering Pattern

Use consistent label width for alignment:

```go
const labelWidth = 12

func formatField(label, value string) string {
    paddedLabel := lipgloss.NewStyle().Width(labelWidth).Render(label + ":")
    return paddedLabel + " " + value
}
```

### Confidence Styling

```go
func renderConfidence(conf domain.Confidence) string {
    switch conf {
    case domain.ConfidenceCertain:
        return "Certain"
    case domain.ConfidenceLikely:
        return "Likely"
    case domain.ConfidenceUncertain:
        return UncertainStyle.Render("Uncertain")
    default:
        return "Unknown"
    }
}
```

### Split Layout with Lipgloss

```go
import "github.com/charmbracelet/lipgloss"

// Join views horizontally with proper alignment
combined := lipgloss.JoinHorizontal(
    lipgloss.Top,
    listView,
    detailView,
)
```

### Height Threshold Logic

```
Height < 30:    Panel CLOSED by default, show hint
Height 30-34:   Panel CLOSED by default (user can toggle)
Height >= 35:   Panel OPEN by default
```

### Reuse from Previous Stories

**From Story 3.1:**
- `components.ProjectListModel` - The list model to display alongside
- `shared/timeformat.FormatRelativeTime` - For "Last Active" field
- `shared/project.EffectiveName` - For panel title

**From Story 1.6 (styles.go):**
- `BorderStyle` - For panel border
- `titleStyle` - For "DETAILS:" header
- `UncertainStyle` - For uncertain confidence
- `DimStyle` - For hints

**From Story 3.2 (keys.go):**
- `KeyBindings` struct pattern - Follow same pattern for KeyDetail

### Testing Pattern

Follow existing test patterns from `model_test.go`:

```go
func TestModel_DetailPanelToggle(t *testing.T) {
    m := createModelWithProjects(t)
    m.showDetailPanel = false

    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
    newModel, _ := m.Update(msg)
    updated := newModel.(Model)

    if !updated.showDetailPanel {
        t.Error("Detail panel should be visible after pressing 'd'")
    }

    // Toggle again
    newModel, _ = updated.Update(msg)
    updated = newModel.(Model)

    if updated.showDetailPanel {
        t.Error("Detail panel should be hidden after pressing 'd' again")
    }
}
```

### Project Context Rules (CRITICAL)

From `project-context.md`:
- Context first: Methods accept `context.Context` as first parameter (not needed for pure UI components)
- Co-locate tests: `detail_panel_test.go` next to `detail_panel.go`
- Naming: PascalCase for exported (DetailPanelModel), camelCase for unexported
- Test pattern: Table-driven tests with `tests []struct{...}`

### Architecture Boundaries

- DetailPanel is a pure UI component - no repository calls
- Data flows: Model → DetailPanel via SetProject()
- DetailPanel renders what it's given, Model handles state

### Domain Types Reference

**NOTE:** Task 0 adds `Confidence` and `DetectionReasoning` fields to `domain.Project`.

From `internal/core/domain/` (after Task 0):
```go
// Project fields relevant to detail panel (internal/core/domain/project.go):
type Project struct {
    Name               string
    Path               string
    DisplayName        string
    DetectedMethod     string
    CurrentStage       Stage
    Confidence         Confidence  // Added by Task 0
    DetectionReasoning string      // Added by Task 0
    Notes              string
    CreatedAt          time.Time
    LastActivityAt     time.Time
}

// Confidence levels (internal/core/domain/confidence.go):
const (
    ConfidenceUncertain Confidence = iota  // Zero value = Uncertain
    ConfidenceLikely
    ConfidenceCertain
)
```

**Important:** The zero value for Confidence is `ConfidenceUncertain`, not `ConfidenceUnknown`.

### Files Modified in Stories 3.1 and 3.2

Key patterns from recent commits:
- `model.go`: ProjectsLoadedMsg handling, projectList integration, renderDashboard()
- `keys.go`: KeyEscape added following KeyBindings struct pattern
- Components follow NewXxxModel() constructor pattern with SetSize(), View(), Update()

### References

- [Source: docs/epics.md#story-3.3] Story requirements (lines 1168-1219)
- [Source: docs/architecture.md#tui-framework] Bubble Tea patterns
- [Source: docs/project-context.md] Go patterns and testing rules
- [Source: internal/adapters/tui/model.go] Current model structure (Story 3.1/3.2)
- [Source: internal/adapters/tui/styles.go] Lipgloss style definitions
- [Source: internal/adapters/tui/components/project_list.go] Component patterns

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 3.3 requirements - lines 1168-1219)
- docs/architecture.md (TUI framework, Bubble Tea patterns)
- docs/project-context.md (Go conventions, testing rules)
- internal/adapters/tui/model.go (Current TUI implementation after Stories 3.1, 3.2)
- internal/adapters/tui/styles.go (Lipgloss styles)
- internal/adapters/tui/keys.go (Key binding patterns)
- internal/adapters/tui/components/project_list.go (Component patterns)
- docs/sprint-artifacts/stories/epic-3/3-1-project-list-component.md (Previous story learnings)
- docs/sprint-artifacts/stories/epic-3/3-2-keyboard-navigation.md (Previous story learnings)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story drafting phase.

### Completion Notes List

- **Task 0:** Added `Confidence` and `DetectionReasoning` fields to `domain.Project` struct. Updated `rowToProject()` in repository.go to map these fields from database. Updated `Save()` to persist these fields.
- **Task 1:** Created `DetailPanelModel` component with `View()`, `SetProject()`, `SetSize()`, `SetVisible()`, `IsVisible()`. Handles nil project gracefully, shows "(none)" for empty notes, applies `UncertainStyle` for uncertain confidence, shows placeholder for empty detection reasoning.
- **Task 2:** Added `KeyDetail = "d"` constant and `Detail` field to `KeyBindings` struct. Added toggle handling in `handleKeyMsg()`.
- **Task 3:** Added `showDetailPanel` and `detailPanel` fields to Model. Implemented `shouldShowDetailPanelByDefault()` helper (>= 35 rows = open). Updated `resizeTickMsg` to set initial visibility on first ready. Updated `ProjectsLoadedMsg` to sync detail panel with selected project. Detail panel updates on selection change via key navigation.
- **Task 4:** Updated `renderDashboard()` with split layout (60%/40%). Added hint "Press [d] for details" when height < 30 and panel closed.
- **Task 5:** Added comprehensive tests: 9 detail panel component tests, 11 model tests for toggle/default state/hint behavior. All tests pass, lint clean, build succeeds.
- **Code Review Fixes:** Extracted shared styles to package-level vars (detailBorderStyle, detailTitleStyle, uncertainStyle) with "Keep in sync with styles.go" documentation to avoid import cycle. Added TestModel_DetailPanelSplitLayout test. Fixed height consistency between renderEmpty() and renderProject().

### File List

**Created:**
- `internal/adapters/tui/components/detail_panel.go` - DetailPanelModel component
- `internal/adapters/tui/components/detail_panel_test.go` - 9 tests for detail panel

**Modified (Task 0 - Domain Changes):**
- `internal/core/domain/project.go` - Added Confidence and DetectionReasoning fields to Project struct
- `internal/adapters/persistence/sqlite/repository.go` - Mapped new fields in rowToProject() and Save()

**Modified (Tasks 1-5 - TUI Changes):**
- `internal/adapters/tui/keys.go` - Added KeyDetail constant and Detail field to KeyBindings
- `internal/adapters/tui/model.go` - Added detail panel state, toggle logic, split layout rendering, height-based default visibility
- `internal/adapters/tui/model_test.go` - Added 12 detail panel tests including toggle, default state, hint behavior, and split layout verification

## Change Log

| Date | Change |
|------|--------|
| 2025-12-14 | Story created with ready-for-dev status by SM Agent (Bob) in YOLO mode. Comprehensive context analysis from epics, architecture, project-context, and previous stories 3.1/3.2. |
| 2025-12-14 | **Validation Review Applied:** CRITICAL: Added Task 0 prerequisite to add Confidence/DetectionReasoning fields to domain.Project (fields don't exist in current domain entity but are required by AC1/AC8). Fixed Task 3.4 initialization logic (wasReady pattern). Clarified AC4 hint location (bottom of project list, not status area). Added boundary condition tests (height 29/30/34/35). Updated Domain Types Reference to reflect accurate domain structure after Task 0. Added version note. Updated File List to include domain files. Reviewed by SM Agent (Bob). |
| 2025-12-14 | **Implementation Complete:** All 6 tasks completed by Dev Agent (Amelia). Created DetailPanelModel component with full field rendering, edge case handling, and visibility control. Added KeyDetail 'd' key binding with toggle logic. Implemented height-based default visibility (>= 35 rows = open). Split layout (60%/40%) with hint for short terminals. 20 tests added covering all ACs. All tests pass, lint clean, build succeeds. Status changed to in-review. |
| 2025-12-14 | **Code Review Applied:** Fixed 6 issues. H1: Extracted shared styles to package-level vars with "Keep in sync" documentation (detailBorderStyle, detailTitleStyle, uncertainStyle) - avoids import cycle since components cannot import tui. M1-M2: Same fix as H1. M3: Added TestModel_DetailPanelSplitLayout test verifying dual-pane render. M4: Added consistent Height to both renderEmpty() and renderProject(). L1-L2: Documented as acceptable (magic numbers match styles.go, test date is static fixture). All tests pass, lint clean, build succeeds. Reviewed by Dev Agent (Amelia). |
