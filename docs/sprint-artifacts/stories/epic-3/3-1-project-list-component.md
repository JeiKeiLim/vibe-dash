# Story 3.1: Project List Component

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/model.go`, `internal/adapters/tui/components/` |
| **Key Dependencies** | Bubbles list (`github.com/charmbracelet/bubbles`), Lipgloss styles (Story 1.6), domain.Project, ProjectRepository |
| **Files to Create** | `components/project_list.go`, `components/delegate.go`, `timeline.go` |
| **Files to Modify** | `model.go`, `views.go`, `app.go` |
| **Location** | `internal/adapters/tui/`, `internal/adapters/tui/components/` |
| **Interfaces Used** | `ports.ProjectRepository.FindAll()`, `list.ItemDelegate` |

### Quick Task Summary (7 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 0 | Verify Bubbles dependency | Ensure `github.com/charmbracelet/bubbles` is in go.mod |
| 1 | Create timeline/recency utilities | `timeline.go` with `RecencyIndicator()` (reuse `FormatRelativeTime` from cli/list.go) |
| 2 | Create ProjectItem type | Wrapper implementing list.Item interface |
| 3 | Create ProjectItemDelegate | Custom delegate for rendering project rows |
| 4 | Create ProjectList component | Bubbles list wrapper with project loading |
| 5 | Integrate list into Model | Replace EmptyView with project list when projects exist |
| 6 | Add tests and verify | Unit tests for all components, integration testing |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| List library | Bubbles `list` | Standard Charm component, scrollable, customizable delegate |
| Row format | `> name    indicator  stage  waiting  time` | Per epic spec, column-based layout |
| Selection marker | `>` prefix | Standard TUI pattern, non-intrusive |
| Recency threshold (today) | < 24 hours | ✨ indicates activity today |
| Recency threshold (week) | < 7 days | ⚡ indicates activity this week |
| Sorting | Alphabetical by effectiveName | Consistent with `vibe list` command (Story 2.7) |

## Story

**As a** user,
**I want** to see my projects in a scrollable list,
**So that** I can browse all tracked projects.

## Acceptance Criteria

```gherkin
AC1: Given projects are tracked
     When dashboard loads
     Then project list displays with columns:
       - Selection indicator (> when selected)
       - Project name (or display_name if set)
       - Recency indicator (✨ today, ⚡ this week)
       - Stage name
       - Status (⏸️ WAITING if applicable)
       - Last activity time

AC2: Given projects are tracked
     When dashboard loads
     Then projects are sorted alphabetically by name

AC3: Given projects are tracked
     When dashboard loads
     Then first project is selected by default

AC4: Given project list exceeds visible rows
     When user navigates (j/k - Story 3.2)
     Then list scrolls to keep selection visible
     And scroll behavior is smooth

AC5: Given project list is scrollable
     When list is rendered
     Then scroll position indicator shows (e.g., "1/5")

AC6: Given no projects exist
     When dashboard loads
     Then EmptyView is shown instead of project list
     (This is existing behavior, must be preserved)
```

**Example row format:**
```
> client-bravo     ⚡ Plan      ⏸️ WAITING   2h ago
  client-alpha     ✨ Tasks                  5m ago
  project-x           Specify                3d ago
```

## Tasks / Subtasks

- [x] **Task 0: Verify Bubbles dependency** (Pre-requisite)
  - [x] 0.1 Check if `github.com/charmbracelet/bubbles` exists in `go.mod`
  - [x] 0.2 If not present, run: `go get github.com/charmbracelet/bubbles`
  - [x] 0.3 Verify import works: `import "github.com/charmbracelet/bubbles/list"`

- [x] **Task 1: Create timeline/recency utilities** (AC: 1)
  - [x] 1.1 Create `internal/adapters/tui/timeline.go` with `package tui`
  - [x] 1.2 Implement `RecencyIndicator(lastActivity time.Time) string`:
    - Handle zero time: `if t.IsZero() { return "" }`
    - Returns `✨` if < 24 hours ago
    - Returns `⚡` if < 7 days ago but >= 24 hours
    - Returns empty string otherwise
  - [x] 1.3 **REUSE** `FormatRelativeTime()` from `internal/adapters/cli/list.go:122-138`
    - Option A: Import and call directly (preferred if package visibility allows)
    - Option B: Extract to shared package `internal/shared/timeformat/` ✓ IMPLEMENTED
    - Option C: Copy with attribution comment if extraction is out of scope
  - [x] 1.4 Write unit tests in `timeline_test.go` including zero timestamp edge case

- [x] **Task 2: Create ProjectItem type** (AC: 1, 2, 3)
  - [x] 2.1 Create `internal/adapters/tui/components/models.go` with `package components`
  - [x] 2.2 Define `ProjectItem` struct wrapping `*domain.Project`
  - [x] 2.3 Implement `list.Item` interface:
    - `FilterValue() string` - returns effectiveName for filtering
    - `Title() string` - returns effectiveName
    - `Description() string` - returns stage string
  - [x] 2.4 **REUSE** `effectiveName()` pattern from `internal/adapters/cli/list.go:88-93`
    - Extracted to shared package `internal/shared/project/` ✓ IMPLEMENTED

- [x] **Task 3: Create ProjectItemDelegate** (AC: 1, 5)
  - [x] 3.1 Create `internal/adapters/tui/components/delegate.go` with `package components`
  - [x] 3.2 Define `ProjectItemDelegate` struct implementing `list.ItemDelegate`
  - [x] 3.3 Implement `Render(w io.Writer, m list.Model, index int, item list.Item)`:
    - Selection indicator: `>` when selected, ` ` otherwise
    - Project name with SelectedStyle when selected
    - Recency indicator (✨/⚡) with RecentStyle/ActiveStyle
    - Stage name
    - WAITING indicator using WaitingStyle if applicable (placeholder for Story 4.x)
    - Relative time using FormatRelativeTime()
  - [x] 3.4 Implement `Height() int` - return 1 (single-line rows)
  - [x] 3.5 Implement `Spacing() int` - return 0 (no spacing)
  - [x] 3.6 Implement `Update(msg tea.Msg, m *list.Model) tea.Cmd` - return nil (no custom updates)
  - [x] 3.7 Use column-width calculations for alignment (see Dev Notes)

- [x] **Task 4: Create ProjectList component** (AC: 1, 2, 3, 4, 5, 6)
  - [x] 4.1 Create `internal/adapters/tui/components/project_list.go` with `package components`
  - [x] 4.2 Define `ProjectListModel` struct with:
    - `list list.Model` - the Bubbles list
    - `projects []*domain.Project` - backing data
    - `width, height int` - dimensions for layout
  - [x] 4.3 Implement `NewProjectListModel(projects []*domain.Project, width, height int) ProjectListModel`:
    - Sort projects alphabetically by effectiveName
    - Convert to ProjectItem slice
    - Initialize Bubbles list with custom delegate
    - Configure list: hide help, hide title, set dimensions
    - Select first item
  - [x] 4.4 Implement `SetProjects(projects []*domain.Project)` for refreshing
  - [x] 4.5 Implement `SetSize(width, height int)` for responsive layout
  - [x] 4.6 Implement `View() string` - render the list
  - [x] 4.7 Implement `Update(msg tea.Msg) (ProjectListModel, tea.Cmd)` - delegate to list
  - [x] 4.8 Implement `SelectedProject() *domain.Project` - return currently selected
  - [x] 4.9 Implement `HasProjects() bool` - check if list is non-empty

- [x] **Task 5: Integrate list into Model** (AC: 1-6)
  - [x] 5.1 Add `projectList components.ProjectListModel` field to Model struct
  - [x] 5.2 Add `projects []*domain.Project` field for backing data
  - [x] 5.3 Create `loadProjectsCmd()` command that calls `repo.FindAll(ctx)`
  - [x] 5.4 Create `ProjectsLoadedMsg` type for async loading result
  - [x] 5.5 Handle `ProjectsLoadedMsg` in Update():
    - Store projects
    - Initialize ProjectListModel
    - Switch to normal view mode (not validation)
  - [x] 5.6 Update Init() to chain: validatePathsCmd → loadProjectsCmd
  - [x] 5.7 Update View() to:
    - Return EmptyView if no projects AND not in validation mode
    - Return ProjectList.View() if projects exist
    - Return validation dialog if in validation mode
  - [x] 5.8 Update WindowSizeMsg handler to call projectList.SetSize()
  - [x] 5.9 Forward key messages to projectList.Update() when in normal mode

- [x] **Task 6: Add tests and verify** (AC: all)
  - [x] 6.1 Create `internal/adapters/tui/components/delegate_test.go`
  - [x] 6.2 Create `internal/adapters/tui/components/project_list_test.go`
  - [x] 6.3 Test RecencyIndicator with various timestamps including:
    - Zero timestamp (`time.Time{}`) should return empty string
    - Future timestamp should handle gracefully
    - Boundary conditions (exactly 24h, exactly 7d)
  - [x] 6.4 Test FormatRelativeTime edge cases including zero timestamp
  - [x] 6.5 Test ProjectItemDelegate renders correctly
  - [x] 6.6 Test ProjectListModel sorting
  - [x] 6.7 Test Model integration: empty vs populated
  - [x] 6.8 Run `make test` and verify all pass
  - [x] 6.9 Run `make lint` and verify no errors
  - [x] 6.10 Run `make build` and verify successful build
  - [x] 6.11 Manual test: run `vibe`, verify list appears with test projects

## Dev Notes

### Project Row Rendering Layout

Use fixed column widths for alignment. Calculate dynamically based on terminal width:

```go
// Minimum layout (60 chars per UX spec):
// [>][name............][ind][stage......][waiting....][time....]
// [ 1][      20       ][ 2 ][    10     ][    12     ][   8    ] = 53 chars + padding

const (
    colSelection = 2   // "> " or "  "
    colNameMin   = 15
    colIndicator = 3   // "✨ " or "⚡ " or "   "
    colStage     = 10  // "Implement" is longest
    colWaiting   = 14  // "⏸️ WAITING Xh" or empty
    colTime      = 8   // "2w ago" max
)

// Dynamic name width with safety bounds:
nameWidth := width - colSelection - colIndicator - colStage - colWaiting - colTime - 4 // 4 for spacing
if nameWidth < colNameMin {
    nameWidth = colNameMin
}
if nameWidth < 1 {
    nameWidth = 1 // Absolute minimum to prevent negative widths
}
```

### effectiveName Pattern (REUSE from Story 2.7)

**Existing implementation:** `internal/adapters/cli/list.go:88-93`

Reuse this function rather than duplicating. Options:
1. Export as `EffectiveName()` in cli package and import
2. Extract to `internal/shared/project.go` for cross-package use
3. If scope-limited, copy with comment: `// Copied from cli/list.go - consider extraction`

### Bubbles List Configuration

```go
import "github.com/charmbracelet/bubbles/list"

delegate := NewProjectItemDelegate()
items := make([]list.Item, len(projects))
for i, p := range projects {
    items[i] = ProjectItem{Project: p}
}

l := list.New(items, delegate, width, height)
l.SetShowHelp(false)           // We have our own help (Story 3.5)
l.SetShowTitle(false)          // Title in our own header
l.SetShowStatusBar(true)       // Shows "1/5" pagination
l.SetFilteringEnabled(false)   // For now, enable in post-MVP
l.KeyMap = list.DefaultKeyMap()
l.KeyMap.ForceQuit.Unbind()    // We handle quit ourselves
```

### Recency Indicator Logic

```go
func RecencyIndicator(lastActivity time.Time) string {
    // Handle zero/unset timestamps gracefully
    if lastActivity.IsZero() {
        return ""
    }
    since := time.Since(lastActivity)
    switch {
    case since < 24*time.Hour:
        return "✨"  // Today
    case since < 7*24*time.Hour:
        return "⚡"  // This week
    default:
        return ""    // Older
    }
}
```

### FormatRelativeTime (REUSE from Story 2.7)

**Existing implementation:** `internal/adapters/cli/list.go:122-138`

This function already exists and is tested. Reuse rather than duplicate:
1. Export as `FormatRelativeTime()` in cli package
2. Or extract to shared package
3. If reusing directly isn't possible, add zero-time handling:

```go
// Only if creating new implementation
func FormatRelativeTime(t time.Time) string {
    if t.IsZero() {
        return "never"
    }
    // ... rest of implementation from cli/list.go:122-138
}
```

### WAITING Indicator (Placeholder for Story 4.x)

For now, the Project entity doesn't have a "waiting" field. Story 4.3 will add agent waiting detection. For this story, implement the rendering but skip the WAITING column if no projects are waiting:

```go
// Placeholder: waiting detection comes in Story 4.3
func isWaiting(p *domain.Project) bool {
    // TODO: Implement in Story 4.3
    // For now, always return false
    return false
}

func waitingIndicator(p *domain.Project) string {
    if !isWaiting(p) {
        return ""
    }
    // Duration will come from Story 4.3 implementation
    return "⏸️ WAITING"
}
```

### Model Integration Pattern

Follow the existing pattern from model.go for async commands:

```go
// Message types
type ProjectsLoadedMsg struct {
    projects []*domain.Project
    err      error
}

// Command function
func loadProjectsCmd(repo ports.ProjectRepository) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        projects, err := repo.FindAll(ctx)
        return ProjectsLoadedMsg{projects: projects, err: err}
    }
}

// In Update() - with proper error handling:
case ProjectsLoadedMsg:
    if msg.err != nil {
        // Log error per project conventions (log at handling site only)
        slog.Error("Failed to load projects", "error", msg.err)
        // Set error state for UI feedback (consider adding loadError field to Model)
        m.projects = nil
        // Could show error in status bar or emit error message
        return m, nil
    }
    m.projects = msg.projects
    if len(m.projects) > 0 {
        m.projectList = components.NewProjectListModel(m.projects, m.width, m.height)
    }
    return m, nil
```

### Sorting Implementation (REUSE from Story 2.7)

**Existing implementation:** `internal/adapters/cli/list.go:96-102`

The `sortProjects()` function already exists. Reuse the same approach or call directly if exported.

### View Switching Logic

```go
func (m Model) View() string {
    // Priority order for rendering:

    // 1. Too small terminal
    if m.width < MinWidth || m.height < MinHeight {
        return m.renderTooSmallView()
    }

    // 2. Help overlay (overlays everything)
    if m.showHelp {
        return m.renderHelpOverlay()
    }

    // 3. Validation mode (path validation dialog)
    if m.viewMode == viewModeValidation {
        return m.renderValidationDialog()
    }

    // 4. Empty view (no projects)
    if len(m.projects) == 0 {
        return m.renderEmptyView()
    }

    // 5. Normal view (project list)
    return m.renderDashboard()
}

func (m Model) renderDashboard() string {
    // For now, just the project list
    // Story 3.3 adds detail panel
    // Story 3.4 adds status bar
    return m.projectList.View()
}
```

### File Structure After Implementation

```
internal/adapters/tui/
├── app.go                    # Entry point (modify)
├── model.go                  # Main model (modify)
├── views.go                  # View rendering (modify)
├── styles.go                 # Lipgloss styles (no change)
├── keys.go                   # Key bindings (no change)
├── validation.go             # Path validation (no change)
├── timeline.go               # NEW: recency and time formatting
├── timeline_test.go          # NEW: timeline tests
├── components/
│   ├── .keep                 # Remove this file
│   ├── models.go             # NEW: ProjectItem type
│   ├── delegate.go           # NEW: ProjectItemDelegate
│   ├── delegate_test.go      # NEW: delegate tests
│   ├── project_list.go       # NEW: ProjectListModel
│   └── project_list_test.go  # NEW: list tests
```

### Testing Patterns

Follow existing test patterns from model_test.go:

```go
func TestRecencyIndicator_Today(t *testing.T) {
    now := time.Now()
    tests := []struct {
        name     string
        activity time.Time
        expected string
    }{
        {"just now", now, "✨"},
        {"1 hour ago", now.Add(-1 * time.Hour), "✨"},
        {"23 hours ago", now.Add(-23 * time.Hour), "✨"},
        {"25 hours ago", now.Add(-25 * time.Hour), "⚡"},
        {"6 days ago", now.Add(-6 * 24 * time.Hour), "⚡"},
        {"8 days ago", now.Add(-8 * 24 * time.Hour), ""},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := RecencyIndicator(tt.activity)
            if got != tt.expected {
                t.Errorf("RecencyIndicator() = %q, want %q", got, tt.expected)
            }
        })
    }
}
```

### Project Context Rules (CRITICAL)

From `project-context.md`:
- Context first: All service methods accept `context.Context` as first parameter
- Co-locate tests: `timeline_test.go` next to `timeline.go`
- Naming: PascalCase for exported, camelCase for unexported
- Test pattern: Table-driven tests with `tests []struct{...}`

### Architecture Boundaries

- `components/` directory contains UI components only
- Components don't directly call repository - Model handles data loading
- Components receive data via constructor or setter methods
- Styles are imported from `../styles.go`

### Previous Story Patterns

From Story 2.10 (golden path test fixtures):
- Comprehensive test case tables
- Clear documentation of expected behaviors
- Edge case coverage (empty list, single project, many projects)

### References

- [Source: docs/epics.md#story-3.1] Story requirements (lines 1080-1119)
- [Source: docs/architecture.md#tui-framework] Bubble Tea patterns
- [Source: docs/project-context.md] Go patterns and testing rules
- [Source: internal/adapters/tui/model.go] Existing TUI model structure
- [Source: internal/adapters/tui/styles.go] Lipgloss style definitions
- [Source: internal/adapters/cli/list.go] effectiveName pattern, sorting
- [Source: docs/ux.md] (if exists) ProjectItemDelegate specification

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 3.1 requirements - lines 1080-1119)
- docs/architecture.md (TUI framework, Bubble Tea patterns)
- docs/project-context.md (Go conventions, testing rules)
- docs/prd.md (Dashboard visualization requirements FR15-17)
- internal/adapters/tui/ (Current TUI implementation)
- internal/adapters/cli/list.go (Project listing patterns)
- internal/core/domain/project.go (Project entity fields)
- docs/sprint-artifacts/stories/epic-2/2-10-golden-path-test-fixtures.md (Story format example)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - No debug issues encountered.

### Completion Notes List

**Implementation Summary:**

1. **Shared Packages Created**: Extracted `FormatRelativeTime` and `RecencyIndicator` to `internal/shared/timeformat/` package. Extracted `EffectiveName` and `SortByName` to `internal/shared/project/` package for cross-adapter reuse.

2. **Components Implemented**:
   - `ProjectItem` wraps domain.Project and implements list.Item interface
   - `ProjectItemDelegate` renders project rows with selection indicator, name, recency, stage, waiting placeholder, and time
   - `ProjectListModel` wraps Bubbles list with sorting, selection, and responsive sizing

3. **Model Integration**: Added async project loading with `loadProjectsCmd()` and `ProjectsLoadedMsg`. Updated View() to render project list when projects exist, preserving EmptyView for empty state.

4. **Import Cycle Resolution**: Resolved import cycle between tui and components by using shared/timeformat directly in delegate and duplicating style definitions locally in components package.

5. **All ACs Satisfied**:
   - AC1: Project list displays all columns (selection, name, recency, stage, waiting, time)
   - AC2: Projects sorted alphabetically by effective name
   - AC3: First project selected by default
   - AC4: Bubbles list handles scrolling natively
   - AC5: Status bar shows pagination ("1/5")
   - AC6: EmptyView preserved for empty project list

### File List

**Created:**
- `internal/shared/timeformat/timeformat.go` - Shared time formatting utilities
- `internal/shared/timeformat/timeformat_test.go` - Time formatting tests
- `internal/shared/project/project.go` - Shared project utilities (EffectiveName, SortByName)
- `internal/shared/project/project_test.go` - Project utilities tests (code review fix H1)
- `internal/adapters/tui/timeline.go` - TUI wrapper for timeformat functions
- `internal/adapters/tui/timeline_test.go` - Timeline tests
- `internal/adapters/tui/components/models.go` - ProjectItem type
- `internal/adapters/tui/components/models_test.go` - ProjectItem interface tests (code review fix M1)
- `internal/adapters/tui/components/delegate.go` - ProjectItemDelegate
- `internal/adapters/tui/components/delegate_test.go` - Delegate tests (enhanced with recency indicator tests - code review fix H2)
- `internal/adapters/tui/components/project_list.go` - ProjectListModel (SetSize fixed - code review fix M2)
- `internal/adapters/tui/components/project_list_test.go` - List tests

**Modified:**
- `go.mod` - Added github.com/charmbracelet/bubbles dependency
- `internal/adapters/cli/list.go` - Updated to use shared timeformat and project packages
- `internal/adapters/tui/model.go` - Added projectList, projects, loadProjectsCmd, ProjectsLoadedMsg, renderDashboard

**Deleted:**
- `internal/adapters/tui/components/.keep` - No longer needed

## Change Log

| Date | Change |
|------|--------|
| 2025-12-14 | Story created with ready-for-dev status by SM Agent (Bob) |
| 2025-12-14 | **Validation Review Applied:** Added Task 0 (Bubbles dependency), code reuse guidance for effectiveName/FormatRelativeTime/sortProjects, package declarations, improved error handling, width validation bounds, zero timestamp handling. Reviewed by SM Agent (Bob). |
| 2025-12-14 | **Implementation Complete:** All 7 tasks completed. Created shared packages (timeformat, project), TUI components (ProjectItem, ProjectItemDelegate, ProjectListModel), integrated into Model with async loading. All tests pass, lint clean, build successful. Ready for code review. Dev Agent (Amelia). |
| 2025-12-14 | **Code Review Fixes Applied:** H1: Added project_test.go for shared/project package. H2: Added recency indicator tests to delegate_test.go. M1: Added models_test.go for ProjectItem interface. M2: Fixed SetSize to propagate delegate width via list.SetDelegate(). All tests pass, lint clean, build successful. Dev Agent (Amelia). |
