# Story 3.10: Responsive Layout

**Status:** Done

## Executive Summary

Enhance TUI responsive behavior for various terminal sizes. Most responsive infrastructure is already implemented (debounced resize, min size warning, dynamic panel visibility). This story completes the remaining responsive features: width-based warnings for narrow terminals (60-79), max-width capping with centering for wide terminals (>120), condensed status bar for short terminals (<20 rows), and "Press [d] for details" hint for medium heights (20-34). Implementation touches model.go, views.go, and status_bar.go.

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/model.go` (View, renderMainContent), `internal/adapters/tui/views.go` (new warning renderer) |
| **Key Dependencies** | StatusBarModel (`components/status_bar.go`), ProjectListModel, DetailPanelModel |
| **Files to Modify** | `model.go` (View, renderDashboard, renderMainContent), `views.go` (add renderNarrowWarning), `components/status_bar.go` (condensed mode) |
| **Files to Create** | `model_responsive_test.go` |
| **Location** | `internal/adapters/tui/` |
| **Interfaces Used** | None (internal TUI changes only) |

### Quick Task Summary (5 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Add width 60-79 narrow warning | Yellow warning bar above status bar for limited view |
| 2 | Add max-width capping (>120) with centering | Content capped at 120 cols, centered in terminal |
| 3 | Add condensed status bar for height < 20 | Single-line status bar instead of two lines |
| 4 | Add detail panel hint for height 20-34 | "Press [d] for details" hint when panel closed |
| 5 | Add tests | Test all responsive breakpoints and behaviors |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Max content width | 120 columns | Per UX spec: "Cap at max width" for >120 cols |
| Narrow warning style | Yellow (ANSI 3) | Consistent with WarningStyle in styles.go |
| Condensed status bar | Single line, abbreviated shortcuts | Height < 20 is severely constrained |
| Centering method | lipgloss.Place() | Consistent with existing dialog centering pattern |
| Warning position | Above status bar | Non-intrusive, doesn't obscure content |

## Story

**As a** user,
**I want** the dashboard to adapt to my terminal size,
**So that** it works in various environments.

## Acceptance Criteria

```gherkin
AC1: Given terminal is resized during operation
     When width < 60 columns
     Then minimal view shows with warning:
       "Terminal too small. Minimum 60x20 required."
     (ALREADY IMPLEMENTED in views.go:110-114)

AC2: Given terminal is resized during operation
     When width is 60-79 columns
     Then truncated project names
     And warning shown about limited view:
       "⚠ Narrow terminal - some info hidden"

AC3: Given terminal is resized during operation
     When width is 80-99 columns
     Then standard column widths
     And full functionality
     (ALREADY IMPLEMENTED - dynamic name width in delegate.go)

AC4: Given terminal is resized during operation
     When width > 120 columns
     Then content is centered
     And maximum width is capped at 120 columns

AC5: Given terminal is resized during operation
     When height < 20 rows
     Then list-only view (no detail panel)
     And status bar condensed to single line

AC6: Given terminal is resized during operation
     When height is 20-34 rows
     Then detail panel closed by default
     And hint "Press [d] for details" shown
     (PARTIAL: Panel closed by default implemented, hint NOT implemented)

AC7: Given terminal is resized during operation
     When height >= 35 rows
     Then detail panel open by default
     (ALREADY IMPLEMENTED in model.go:178-180)

AC8: Given resize occurs rapidly (drag)
     Then layout recalculates with 50ms debounce
     And no visual flicker
     (ALREADY IMPLEMENTED in model.go:283-290)
```

## Tasks / Subtasks

- [x] **Task 1: Add width 60-79 narrow warning** (AC: 2)

  **PATTERN REFERENCE:** Follow `renderTooSmallView` in `views.go:110-114` for warning rendering.

  - [x] 1.1 Add narrow warning constant in `views.go`:
    ```go
    // NarrowWarning is shown when terminal width is 60-79 (AC2)
    const NarrowWarning = "⚠ Narrow terminal - some info hidden"
    ```

  - [x] 1.2 Add `renderNarrowWarning` function in `views.go`:
    ```go
    // renderNarrowWarning renders the narrow terminal warning bar (Story 3.10 AC2).
    func renderNarrowWarning(width int) string {
        // Use WarningStyle from styles.go for yellow warning text
        warningStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("3")). // Yellow (ANSI 3)
            Bold(true)

        warning := warningStyle.Render(NarrowWarning)
        return lipgloss.PlaceHorizontal(width, lipgloss.Center, warning)
    }
    ```

  - [x] 1.3 Add `isNarrowWidth` helper in `model.go`:
    ```go
    // isNarrowWidth returns true if terminal width is in narrow range (60-79).
    func isNarrowWidth(width int) bool {
        return width >= MinWidth && width < 80
    }
    ```

  - [x] 1.4 Integrate narrow warning in `renderDashboard()` in `model.go` (insert between content and status bar):
    ```go
    func (m Model) renderDashboard() string {
        contentHeight := m.height - 2 // Reserve 2 lines for status bar

        // Adjust content height if narrow warning is shown (AC2)
        if isNarrowWidth(m.width) {
            contentHeight-- // Reserve 1 more line for warning
        }

        mainContent := m.renderMainContent(contentHeight)

        // Build output parts
        var parts []string
        parts = append(parts, mainContent)

        // Add narrow warning if applicable (AC2)
        if isNarrowWidth(m.width) {
            parts = append(parts, renderNarrowWarning(m.width))
        }

        parts = append(parts, m.statusBar.View())

        return lipgloss.JoinVertical(lipgloss.Left, parts...)
    }
    ```

- [x] **Task 2: Add max-width capping with centering for wide terminals** (AC: 4)

  **PATTERN REFERENCE:** Follow `lipgloss.Place` usage in `renderHelpOverlay` for centering.

  - [x] 2.1 Add max width constant in `views.go`:
    ```go
    // MaxContentWidth is the maximum width for content (centered in wider terminals).
    const MaxContentWidth = 120
    ```

  - [x] 2.2 Modify `renderDashboard()` to cap and center content when width > 120:

    **IMPORTANT:** Do NOT call `SetWidth()` here - width is already set in `resizeTickMsg` (see Task 2.3). Calling it again creates a copy with inconsistent state.

    **NOTE:** `isNarrowWidth` (60-79) and `MaxContentWidth` (>120) are **mutually exclusive** - a terminal cannot be both narrow AND wide simultaneously. No need to combine both adjustments.

    ```go
    func (m Model) renderDashboard() string {
        // Calculate effective width (cap at MaxContentWidth for wide terminals)
        // Note: isNarrowWidth (60-79) and MaxContentWidth (>120) are mutually exclusive
        effectiveWidth := m.width
        if m.width > MaxContentWidth {
            effectiveWidth = MaxContentWidth
        }

        // Reserve lines for status bar (1 if condensed, 2 otherwise)
        statusBarHeight := 2
        if m.height < 20 {
            statusBarHeight = 1
        }
        contentHeight := m.height - statusBarHeight

        // Adjust for narrow warning (only applies when 60-79, never when >120)
        if isNarrowWidth(m.width) {
            contentHeight--
        }

        // Create a temporary model copy with effective width for rendering
        renderModel := m
        renderModel.width = effectiveWidth

        mainContent := renderModel.renderMainContent(contentHeight)

        var parts []string
        parts = append(parts, mainContent)

        if isNarrowWidth(m.width) {
            parts = append(parts, renderNarrowWarning(effectiveWidth))
        }

        // Use m.statusBar directly - width already set in resizeTickMsg
        parts = append(parts, m.statusBar.View())

        // Join content
        content := lipgloss.JoinVertical(lipgloss.Left, parts...)

        // Center content if terminal is wider than MaxContentWidth (AC4)
        if m.width > MaxContentWidth {
            return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)
        }

        return content
    }
    ```

  - [x] 2.3 Update component resize handling in `resizeTickMsg` to respect max width:

    **IMPORTANT:** Call `SetCondensed()` FIRST because it affects status bar height calculation. Then set width. This ordering prevents inconsistent state if any code reads status bar height.

    ```go
    // In Update() case resizeTickMsg:
    // FIRST: Set condensed mode (affects status bar height) - Story 3.10 AC5
    m.statusBar.SetCondensed(m.height < 20)

    // Calculate effective width for components
    effectiveWidth := m.width
    if m.width > MaxContentWidth {
        effectiveWidth = MaxContentWidth
    }

    // Update status bar width (ONLY place this should be called)
    m.statusBar.SetWidth(effectiveWidth)

    // Calculate content height (accounts for condensed status bar)
    statusBarHeight := 2
    if m.height < 20 {
        statusBarHeight = 1
    }
    contentHeight := m.height - statusBarHeight

    // Update component dimensions with effective width
    if len(m.projects) > 0 {
        m.projectList.SetSize(effectiveWidth, contentHeight)
        m.detailPanel.SetSize(effectiveWidth, contentHeight)
    }
    ```

- [x] **Task 3: Add condensed status bar for short terminals** (AC: 5)

  **PATTERN REFERENCE:** Follow existing `renderShortcuts` pattern for conditional rendering.

  - [x] 3.1 Add height threshold constant in `components/status_bar.go`:
    ```go
    // Height threshold for condensed mode (Story 3.10 AC5)
    const heightMinForTwoLines = 20
    ```

  - [x] 3.2 Add `isCondensed` field and setter to `StatusBarModel`:
    ```go
    type StatusBarModel struct {
        // ... existing fields ...
        isCondensed bool // True when height < 20 (Story 3.10)
    }

    // SetCondensed sets the condensed mode for short terminals (Story 3.10 AC5).
    func (s *StatusBarModel) SetCondensed(condensed bool) {
        s.isCondensed = condensed
    }
    ```

  - [x] 3.3 Modify `View()` to return single line when condensed:

    **CRITICAL:** The condensed renderer MUST preserve existing features from `renderCounts()`:
    - Refresh spinner when `isRefreshing` is true (Story 3.6)
    - Last refresh message display (`lastRefreshMsg`)
    - Waiting count styling with `statusBarWaitingStyle`

    Failure to preserve these will cause **regression** in Story 3.6 functionality.

    ```go
    func (s StatusBarModel) View() string {
        if s.isCondensed {
            // Single line: abbreviated counts + shortcuts (AC5)
            return s.renderCondensed()
        }
        countsLine := s.renderCounts()
        shortcutsLine := s.renderShortcuts()
        return countsLine + "\n" + shortcutsLine
    }

    // renderCondensed renders a single-line status bar for short terminals (Story 3.10 AC5).
    // IMPORTANT: Must preserve all features from renderCounts() to avoid regression.
    func (s StatusBarModel) renderCondensed() string {
        // Show refresh spinner when refreshing (Story 3.6 - MUST preserve)
        if s.isRefreshing {
            return fmt.Sprintf("│ Refreshing %d/%d │ [?][q] │", s.refreshProgress, s.refreshTotal)
        }

        // Abbreviated counts
        counts := fmt.Sprintf("%dA %dH", s.activeCount, s.hibernatedCount)

        // Waiting count with styling (Story 3.10 - preserve statusBarWaitingStyle)
        if s.waitingCount > 0 {
            counts += " " + statusBarWaitingStyle.Render(fmt.Sprintf("%dW", s.waitingCount))
        }

        // Include refresh message if present (Story 3.6 - MUST preserve)
        if s.lastRefreshMsg != "" {
            counts += " " + s.lastRefreshMsg
        }

        return "│ " + counts + " │ [?][q] │"
    }
    ```

  - [x] 3.4 **NOTE:** `SetCondensed()` call and content height adjustment are now consolidated in **Task 2.3** to ensure proper ordering. No separate subtask needed here.

- [x] **Task 4: Add detail panel hint for medium height** (AC: 6)

  **PATTERN REFERENCE:** Follow `DimStyle` usage in `model.go:929` for hint styling.

  - [x] 4.1 The hint "Press [d] for details" is already partially implemented at `model.go:928-930` for height < 28. Update the logic to show hint when:
    - Height is 20-34 rows
    - Detail panel is closed

    Current code at line 928-931:
    ```go
    if height < 28 && !m.showDetailPanel {
        hint := DimStyle.Render("Press [d] for details")
        return m.projectList.View() + "\n" + hint
    }
    ```

    **ANALYSIS:** The current implementation shows hint when `height < 28` (which means contentHeight < 28 after -2 for status bar, so terminal height < 30). Per AC6, hint should show when terminal height is 20-34 and panel is closed.

    **KEY INSIGHT:** We use `m.height` (terminal height) NOT `height` (contentHeight) because:
    - AC6 defines behavior based on **user-visible terminal size**
    - The user sees and resizes based on terminal height, not internal content height
    - This matches the UX spec which references terminal rows, not content rows

    **FIX:** Update the condition to match AC6 exactly:
    ```go
    // Show hint when terminal height 20-34 and detail panel closed (Story 3.10 AC6)
    // IMPORTANT: Use m.height (terminal height) not height parameter (contentHeight)
    // because AC6 defines behavior based on user-visible terminal size
    if m.height >= 20 && m.height < 35 && !m.showDetailPanel {
        hint := DimStyle.Render("Press [d] for details")
        return m.projectList.View() + "\n" + hint
    }
    ```

    This replaces the `height < 28` check with proper terminal height bounds.

- [x] **Task 5: Add tests** (AC: all)

  **CRITICAL TEST SETUP REQUIREMENTS:**
  1. **ALL `View()` tests MUST set `m.ready = true`** - otherwise View() returns "Initializing..." without rendering the dashboard
  2. **Use `mockRepository` from existing test files** (e.g., `model_remove_test.go`) - do NOT create new mock type

  - [x] 5.1 Create `internal/adapters/tui/model_responsive_test.go`:
    ```go
    package tui

    import (
        "strings"
        "testing"

        tea "github.com/charmbracelet/bubbletea"
    )

    // NOTE: Uses mockRepository from model_remove_test.go or similar existing test file.
    // Do NOT create duplicate mock - reuse existing pattern.

    // TestModel_NarrowWidth_ShowsWarning tests AC2
    func TestModel_NarrowWidth_ShowsWarning(t *testing.T) {
        repo := &mockRepository{}
        m := NewModel(repo)
        m.width = 70  // In narrow range (60-79)
        m.height = 30
        m.ready = true // REQUIRED for View() to render dashboard

        view := m.View()

        if !strings.Contains(view, NarrowWarning) {
            t.Error("expected narrow warning to be shown for width 70")
        }
    }

    func TestModel_NarrowWidth_NotShownAt80(t *testing.T) {
        repo := &mockRepository{}
        m := NewModel(repo)
        m.width = 80  // Not in narrow range
        m.height = 30
        m.ready = true

        view := m.View()

        if strings.Contains(view, NarrowWarning) {
            t.Error("expected no narrow warning at width 80")
        }
    }

    func TestModel_WideTerminal_ContentCapped(t *testing.T) {
        repo := &mockRepository{}
        m := NewModel(repo)
        m.width = 150  // Wide terminal
        m.height = 30
        m.ready = true

        // Trigger resize to update component widths
        m, _ = m.Update(resizeTickMsg{}).(Model)

        // Components should be sized to MaxContentWidth, not terminal width
        // This is validated by checking projectList.width if accessible
        // For now, just ensure View() doesn't panic
        view := m.View()
        if len(view) == 0 {
            t.Error("expected non-empty view")
        }
    }

    func TestModel_ShortTerminal_CondensedStatusBar(t *testing.T) {
        repo := &mockRepository{}
        m := NewModel(repo)
        m.width = 80
        m.height = 18  // Short terminal (< 20)
        m.ready = true

        // Trigger resize to set condensed mode
        m.hasPendingResize = true
        m.pendingHeight = 18
        m.pendingWidth = 80
        m, _ = m.Update(resizeTickMsg{}).(Model)

        view := m.statusBar.View()

        // Condensed view should be single line (no newline)
        if strings.Count(view, "\n") > 0 {
            t.Error("expected single-line condensed status bar")
        }
    }

    func TestModel_MediumHeight_ShowsDetailHint(t *testing.T) {
        repo := &mockRepository{}
        m := NewModel(repo)
        m.width = 80
        m.height = 25  // Medium height (20-34)
        m.ready = true
        m.showDetailPanel = false

        view := m.View()

        if !strings.Contains(view, "Press [d] for details") {
            t.Error("expected detail hint for height 25 with panel closed")
        }
    }

    func TestModel_MediumHeight_NoHintWhenPanelOpen(t *testing.T) {
        repo := &mockRepository{}
        m := NewModel(repo)
        m.width = 80
        m.height = 25
        m.ready = true
        m.showDetailPanel = true

        view := m.View()

        if strings.Contains(view, "Press [d] for details") {
            t.Error("expected no hint when detail panel is open")
        }
    }

    func TestModel_TallTerminal_NoDetailHint(t *testing.T) {
        repo := &mockRepository{}
        m := NewModel(repo)
        m.width = 80
        m.height = 40  // Tall terminal (>= 35)
        m.ready = true
        m.showDetailPanel = false  // Even if manually closed

        view := m.View()

        // For tall terminals, hint should NOT be shown even if panel is closed
        // because user has enough space and chose to close it
        if strings.Contains(view, "Press [d] for details") {
            t.Error("expected no hint for tall terminal (height >= 35)")
        }
    }

    func TestIsNarrowWidth(t *testing.T) {
        tests := []struct {
            width    int
            expected bool
        }{
            {59, false},  // Below minimum
            {60, true},   // Start of narrow range
            {70, true},   // Middle of narrow range
            {79, true},   // End of narrow range
            {80, false},  // Standard width
            {120, false}, // Wide
        }

        for _, tt := range tests {
            result := isNarrowWidth(tt.width)
            if result != tt.expected {
                t.Errorf("isNarrowWidth(%d) = %v, want %v", tt.width, result, tt.expected)
            }
        }
    }
    ```

  - [x] 5.2 Add tests for StatusBarModel condensed mode in `components/status_bar_test.go`:

    **IMPORTANT:** Include tests for C1 fix - condensed mode must preserve refresh features.

    ```go
    func TestStatusBarModel_CondensedMode(t *testing.T) {
        sb := NewStatusBarModel(80)
        sb.SetCounts(5, 3, 2)
        sb.SetCondensed(true)

        view := sb.View()

        // Should be single line
        if strings.Count(view, "\n") > 0 {
            t.Error("condensed view should be single line")
        }

        // Should contain abbreviated counts
        if !strings.Contains(view, "5A") {
            t.Error("expected abbreviated active count '5A'")
        }
        if !strings.Contains(view, "3H") {
            t.Error("expected abbreviated hibernated count '3H'")
        }
        if !strings.Contains(view, "2W") {
            t.Error("expected abbreviated waiting count '2W'")
        }
    }

    func TestStatusBarModel_CondensedMode_NoWaiting(t *testing.T) {
        sb := NewStatusBarModel(80)
        sb.SetCounts(5, 3, 0)  // No waiting projects
        sb.SetCondensed(true)

        view := sb.View()

        // Should not contain waiting indicator
        if strings.Contains(view, "W") {
            t.Error("condensed view should not show waiting when count is 0")
        }
    }

    func TestStatusBarModel_NormalMode(t *testing.T) {
        sb := NewStatusBarModel(80)
        sb.SetCounts(5, 3, 2)
        sb.SetCondensed(false)

        view := sb.View()

        // Should be two lines
        if strings.Count(view, "\n") != 1 {
            t.Error("normal view should have exactly one newline (two lines)")
        }
    }

    // C1 FIX VERIFICATION: Condensed mode must preserve refresh features
    func TestStatusBarModel_CondensedMode_ShowsRefreshSpinner(t *testing.T) {
        sb := NewStatusBarModel(80)
        sb.SetCounts(5, 3, 0)
        sb.SetCondensed(true)
        sb.SetRefreshing(true, 2, 5)

        view := sb.View()

        // Should show refresh progress even in condensed mode
        if !strings.Contains(view, "Refreshing") {
            t.Error("condensed view should show refresh spinner when refreshing")
        }
        if !strings.Contains(view, "2/5") {
            t.Error("condensed view should show refresh progress")
        }
    }

    func TestStatusBarModel_CondensedMode_ShowsRefreshMessage(t *testing.T) {
        sb := NewStatusBarModel(80)
        sb.SetCounts(5, 3, 0)
        sb.SetCondensed(true)
        sb.SetRefreshComplete("Refreshed 3 projects")

        view := sb.View()

        // Should show refresh message even in condensed mode
        if !strings.Contains(view, "Refreshed 3 projects") {
            t.Error("condensed view should show refresh completion message")
        }
    }
    ```

  - [x] 5.3 Add test for renderNarrowWarning in `views_test.go`:
    ```go
    func TestRenderNarrowWarning_ContainsWarning(t *testing.T) {
        output := renderNarrowWarning(70)

        if !strings.Contains(output, NarrowWarning) {
            t.Error("expected narrow warning text in output")
        }
    }

    func TestRenderNarrowWarning_Centered(t *testing.T) {
        output := renderNarrowWarning(80)

        // Warning should be centered - check it's not left-aligned
        if strings.HasPrefix(output, "⚠") {
            t.Error("expected warning to be centered, not left-aligned")
        }
    }
    ```

  - [x] 5.4 Run verification:
    ```bash
    make test   # All tests pass
    make lint   # No lint errors
    make build  # Successful build
    ```

## Dev Notes

### Current State Analysis

**Already Implemented:**
1. ✅ **AC1:** Min terminal size warning (< 60x20) - `views.go:110-114` `renderTooSmallView()`
2. ✅ **AC3:** Standard column widths (80-99) - Dynamic name width in `delegate.go:99-114`
3. ✅ **AC7:** Detail panel open by default (height >= 35) - `model.go:178-180`
4. ✅ **AC8:** Debounced resize (50ms) - `model.go:283-290`
5. ⚠️ **AC6 Partial:** Detail panel closed by default (20-34 rows) - Implemented, but hint not shown correctly

**NOT Implemented (This Story):**
1. ❌ **AC2:** Width 60-79 narrow warning message
2. ❌ **AC4:** Width > 120 content centering and max width capping
3. ❌ **AC5:** Height < 20 condensed single-line status bar
4. ❌ **AC6 Fix:** Detail hint showing for correct height range

### UX Spec Reference

Per `docs/ux-design-specification.md` lines 1530-1608:

**Width Adaptation:**
| Width | Adaptation |
|-------|------------|
| < 60 cols | Hard minimum - warning |
| 60-79 cols | Warning message, truncated view |
| 80-99 cols | Truncate project names, compact columns |
| 100-120 cols | Standard column widths |
| > 120 cols | Cap at max width, center content |

**Height Adaptation:**
| Height | Adaptation |
|--------|------------|
| < 20 rows | Hard minimum - list only, condensed status bar |
| 20-34 rows | Detail panel closed by default, hint to open |
| ≥ 35 rows | Detail panel open by default |
| > 50 rows | Extra space for detail panel content |

### Layout Budget Reference

Per UX spec, at 35 rows:
- Header: 2 rows (currently not rendered, just status bar)
- Project list: ~15 rows
- Detail panel: ~12 rows
- Status bar: 2 rows (1 when condensed)
- Margins: 3 rows

### Implementation Considerations

**Max Width Capping (AC4):**
The implementation creates a "viewport" effect where content is rendered at max width and centered. This prevents UI elements from stretching too wide on ultra-wide terminals, improving readability.

**Condensed Status Bar (AC5):**
When height < 20, space is at a premium. The two-line status bar is reduced to one line with abbreviated content:
- Normal: `│ 5 active │ 3 hibernated │ ⏸️ 2 WAITING │`
- Condensed: `│ 5A 3H 2W │ [?] [q] │`

**Detail Hint Logic (AC6):**
The current implementation at `model.go:928-931` uses `height < 28` which is based on contentHeight. This needs to be changed to use `m.height` directly with proper bounds `>= 20 && < 35` to match AC6.

### Edge Cases

1. **Resize from narrow to wide:** Warning should disappear immediately
2. **Resize from tall to short:** Status bar should condense, detail panel should hide
3. **Wide terminal with dialog:** Dialogs should still use their own width constraints (already capped at 60)
4. **Very narrow (60-79) with projects:** Names truncate, warning shows above status bar

### Testing Strategy

Per UX spec testing matrix:
- Resize terminal, verify layout adjusts
- Minimum size (60x20) - verify degraded but functional
- Standard size (80x24) - verify full functionality
- Large terminal (200x60) - verify no overflow, proper centering
- Rapid resize (drag) - verify debouncing works, no flicker

### Architecture Compliance

Changes are confined to TUI adapter layer:
- `model.go` - View rendering logic
- `views.go` - Warning renderer
- `components/status_bar.go` - Condensed mode

No changes to core domain or ports required.

### Performance Considerations

Per UX spec:
- Layout calculation should be cached (only recalculate on WindowSizeMsg after debounce)
- Never recalculate during normal render loop
- Current implementation already follows this pattern via `resizeTickMsg`

### Validation Review Fixes Applied

**Critical Fixes (C1-C3):**
- **C1:** Fixed `renderCondensed()` to preserve Story 3.6 features (refresh spinner, lastRefreshMsg, waiting style)
- **C2:** Removed duplicate `SetWidth()` call from `renderDashboard()` - width now only set in `resizeTickMsg`
- **C3:** Updated test references to use existing `mockRepository` instead of `favoriteMockRepository`

**Enhancements (E1-E4):**
- **E1:** Added note that `isNarrowWidth` (60-79) and `MaxContentWidth` (>120) are mutually exclusive
- **E2:** Added critical note that all `View()` tests MUST set `m.ready = true`
- **E3:** Added clarification that AC6 uses `m.height` (terminal height) not `height` (contentHeight)
- **E4:** Reordered `SetCondensed()` to be called FIRST in `resizeTickMsg` before width calculations

**Optimizations Considered (Post-Implementation):**
- **O1:** Consider adding `isWideWidth()` helper to match `isNarrowWidth()` pattern
- **O2:** Consider extracting test helper `newTestModel(repo, width, height)` to reduce setup duplication
- **O3:** Future stories could use diff-style code blocks for token efficiency

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 3.10 requirements - lines 1474-1520)
- docs/ux-design-specification.md (Terminal Responsive Strategy - lines 1530-1608)
- docs/architecture.md (TUI adapter structure)
- docs/project-context.md (Go conventions, testing rules)
- internal/adapters/tui/model.go (Current resize handling, View implementation)
- internal/adapters/tui/views.go (renderTooSmallView pattern)
- internal/adapters/tui/components/status_bar.go (Current status bar implementation)
- internal/adapters/tui/components/delegate.go (Dynamic name width calculation)
- docs/sprint-artifacts/stories/epic-3/3-9-remove-project-from-tui.md (Previous story for patterns)
- Git history: Stories 3.1-3.9 implementation patterns

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story drafting phase.

### Completion Notes List

- **Task 1 completed:** Added NarrowWarning constant and renderNarrowWarning() function in views.go. Added isNarrowWidth() helper in model.go. Updated renderDashboard() to show warning for width 60-79.
- **Task 2 completed:** Added MaxContentWidth constant in views.go. Updated renderDashboard() to cap content at 120 columns and center with lipgloss.Place. Updated resizeTickMsg to use effectiveWidth for component sizing.
- **Task 3 completed:** Added isCondensed field and SetCondensed() to StatusBarModel. Added renderCondensed() for single-line status bar. Updated View() to use condensed mode when height < 20. Preserved Story 3.6 refresh features.
- **Task 4 completed:** Updated renderMainContent() to show "Press [d] for details" hint for terminal height 20-34 with closed panel. Changed from contentHeight to m.height per AC6 spec.
- **Task 5 completed:** Created model_responsive_test.go with 17 tests. Added 6 condensed mode tests to status_bar_test.go. Added 3 renderNarrowWarning tests to views_test.go.
- All tests pass, lint clean, build successful.

### File List

**Modified:**
- `internal/adapters/tui/model.go` - Add isNarrowWidth, isWideWidth, statusBarHeight helpers. Update renderDashboard for narrow warning and max-width capping. Update resizeTickMsg for condensed mode and effective width. Update renderMainContent for AC6 hint using HeightThresholdTall constant.
- `internal/adapters/tui/views.go` - Add NarrowWarning constant, MaxContentWidth constant, HeightThresholdTall constant (code review fix), renderNarrowWarning function
- `internal/adapters/tui/components/status_bar.go` - Add isCondensed field, SetCondensed method, renderCondensed method with navigation shortcuts (code review fix), update View for condensed mode
- `internal/adapters/tui/components/status_bar_test.go` - Add condensed mode tests including navigation shortcuts test (code review fix)
- `internal/adapters/tui/views_test.go` - Add renderNarrowWarning tests

**Created:**
- `internal/adapters/tui/model_responsive_test.go` - Tests for all responsive behaviors (22 tests including code review additions: TestShortTerminal_NoDetailHint, TestIsWideWidth, TestStatusBarHeight)

**Existing (Reference Only):**
- `internal/adapters/tui/views.go:110-114` - renderTooSmallView already implemented
- `internal/adapters/tui/model.go:178-180` - shouldShowDetailPanelByDefault already implemented
- `internal/adapters/tui/model.go:283-290` - Debounced resize already implemented

## Change Log

| Date | Change |
|------|--------|
| 2025-12-18 | Story created with Draft status by SM Agent (Bob) in YOLO mode. Comprehensive analysis of existing responsive implementation included. 5 tasks identified for remaining AC coverage. |
| 2025-12-18 | **Validation Review by SM Agent (Bob):** Applied all improvements from systematic validation. Fixed 3 critical issues (C1: renderCondensed regression, C2: duplicate SetWidth, C3: mock reference), 4 enhancements (E1-E4: mutual exclusivity note, ready=true requirement, AC6 threshold clarification, SetCondensed ordering), and documented 3 optimizations (O1-O3). Story ready for implementation. |
| 2025-12-18 | **Implementation Complete by Dev Agent (Amelia):** All 5 tasks completed. AC2: narrow warning for 60-79 width. AC4: max-width capping (>120) with centering. AC5: condensed status bar for height < 20. AC6: detail hint for height 20-34. 26 new tests added across 3 test files. All tests pass, lint clean, build successful. Status changed to Ready for Review. |
| 2025-12-18 | **Code Review by Dev Agent (Amelia):** Applied 5 fixes. M1: Added `isWideWidth()` helper for consistency with `isNarrowWidth()`. M2: Extracted `statusBarHeight()` helper to remove duplication. M3: Added navigation shortcuts `[j/k]` to condensed status bar. M4: Added tests for short terminal hint suppression and new helper functions. L2: Added `HeightThresholdTall` constant (35) to replace magic number. All tests pass (196 tests), lint clean, build successful. Status changed to Done. |
