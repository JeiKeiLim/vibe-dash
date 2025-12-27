# Story 8.12: Horizontal Layout Height Handling

Status: done

## Story

As a **user viewing the dashboard in horizontal layout mode**,
I want **the project list to be prioritized and properly anchored when terminal height is insufficient**,
So that **I can always see my project list without it being cropped or shifting unexpectedly**.

## Background

From manual testing of Story 8.6 (Horizontal Split Layout Option), three layout issues were discovered:

1. **Height priority**: When terminal height is insufficient to show both project list and detail panel, the current fixed 60/40 split causes project list to be cropped. Project list should be prioritized.

2. **Anchor point**: Project list view appears "attached" to detail panel - when navigating between projects with different detail heights, the project list shifts/crops unexpectedly. Each component should have independent anchor points.

3. **Margin**: There's unnecessary visual margin between project list and detail panel. Detail panel border takes 2 lines which may be excessive for stacked layout.

## Acceptance Criteria

### AC1: Project List Priority
- Given terminal height < threshold (16 lines)
- When detail panel is toggled ON
- Then project list uses ALL available height, detail panel is hidden automatically
- And status bar shows hint: "[d] Detail hidden - insufficient height"

### AC2: Minimum Height Thresholds
- Given horizontal mode with detail visible
- When terminal height >= 16 lines
- Then both components visible with proportional heights
- When terminal height >= 30 lines
- Then use full 60/40 split
- Document thresholds as constants in views.go

### AC3: Independent Anchor Points
- Given project list rendered above detail panel
- When user navigates up/down through projects
- Then project list remains at fixed top position
- And detail panel content scrolls independently (if needed)
- And navigation never causes project list to shift vertically

### AC4: Reduced Margin/Padding
- Given horizontal layout with detail visible
- When rendering detail panel
- Then use borderless or minimal border (top border only removed) to save vertical space
- And maximize content area for both components

### AC5: Graceful Degradation
- Given terminal height < MinHeight (20 lines)
- When in horizontal mode
- Then existing "too small" view shown (model.go:1360-1361)
- And layout changes don't break this behavior

### AC6: Config Interaction
- Given config `detail_layout: horizontal`
- When applying height priority rules
- Then vertical layout remains unchanged (side-by-side, no height priority)

### AC7: Test Coverage
- Given all changes are made
- When `make test && make lint` runs
- Then all tests pass and no lint errors

## Tasks / Subtasks

- [x] Task 1: Add height threshold constants (AC: 2, 5)
  - [x] 1.1: In `internal/adapters/tui/views.go`, add constants AFTER `HeightThresholdTall` (~line 26):
    ```go
    // Story 8.12: Horizontal layout height thresholds
    const (
        // MinListHeightHorizontal is minimum lines for project list in horizontal mode
        MinListHeightHorizontal = 10
        // MinDetailHeightHorizontal is minimum lines for detail panel in horizontal mode
        MinDetailHeightHorizontal = 6
        // HorizontalDetailThreshold is the height at which both list and detail fit
        HorizontalDetailThreshold = MinListHeightHorizontal + MinDetailHeightHorizontal // 16
    )
    ```
  - [x] 1.2: Export constants for use in model.go and tests

- [x] Task 2: Implement height-priority logic in renderHorizontalSplit (AC: 1, 2, 3)
  - [x] 2.1: In `internal/adapters/tui/model.go:renderHorizontalSplit()` (~line 1484-1503), replace fixed 60/40 split:
    ```go
    func (m Model) renderHorizontalSplit(height int) string {
        // Story 8.12: Height-priority algorithm
        // Priority: project list always visible, detail panel collapsible

        listHeight := height
        detailHeight := 0
        showDetail := true

        if height < HorizontalDetailThreshold {
            // Insufficient height - hide detail, give all to list
            showDetail = false
            listHeight = height
        } else if height < 30 {
            // Tight fit - minimum detail, rest to list
            detailHeight = MinDetailHeightHorizontal
            listHeight = height - detailHeight
        } else {
            // Comfortable - use 60/40 split
            listHeight = int(float64(height) * 0.6)
            detailHeight = height - listHeight
        }

        // Create copies with updated sizes - full width for horizontal layout
        projectList := m.projectList
        projectList.SetSize(m.width, listHeight)

        // Render project list
        listView := projectList.View()

        if !showDetail {
            return listView
        }

        detailPanel := m.detailPanel
        detailPanel.SetSize(m.width, detailHeight)

        // Render stacked vertically
        detailView := detailPanel.View()

        return lipgloss.JoinVertical(lipgloss.Left, listView, detailView)
    }
    ```
  - [x] 2.2: Add height hint to status bar when detail auto-hidden (REQUIRED per AC1):
    - When `showDetail == false` due to height threshold, update status bar message
    - Status bar hint: "[d] Detail hidden - insufficient height"

- [x] Task 3: Create horizontal-specific detail panel style (AC: 4)
  - [x] 3.1: In `internal/shared/styles/styles.go`, add AFTER `BorderStyle` (after line 99):
    ```go
    // HorizontalBorderStyle removes top border for horizontal layout stacking (Story 8.12).
    // Saves 1 vertical line when detail panel is below project list.
    var HorizontalBorderStyle = lipgloss.NewStyle().
        Border(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("8")).
        BorderTop(false).
        BorderLeft(true).
        BorderRight(true).
        BorderBottom(true)
    ```
    **Note:** Use same `lipgloss.NormalBorder()` and color "8" as existing BorderStyle for consistency.
  - [x] 3.2: Modify `DetailPanelModel` to accept layout mode parameter:
    - Add `isHorizontal bool` field to DetailPanelModel struct at line 27 (after `durationGetter`)
    - Add `SetHorizontalMode(bool)` method
    - In `renderProject()` at line 172, use `HorizontalBorderStyle` when `isHorizontal == true`
  - [x] 3.3: Wire layout mode from Model to DetailPanelModel in `renderHorizontalSplit()`:
    ```go
    detailPanel := m.detailPanel
    detailPanel.SetHorizontalMode(true) // Use horizontal border style
    detailPanel.SetSize(m.width, detailHeight)
    ```

- [x] Task 4: Fix anchor point independence (AC: 3)
  - [x] 4.1: In `renderHorizontalSplit()`, use `lipgloss.Height()` to enforce fixed heights:
    ```go
    // Enforce fixed heights to prevent content from pushing layout
    listContainer := lipgloss.NewStyle().Height(listHeight)
    detailContainer := lipgloss.NewStyle().Height(detailHeight)

    listView := listContainer.Render(projectList.View())
    detailView := detailContainer.Render(detailPanel.View())
    return lipgloss.JoinVertical(lipgloss.Left, listView, detailView)
    ```
  - [x] 4.2: Ensure project list doesn't inherit scroll position from detail panel
  - [x] 4.3: Test navigation stability across projects with different detail lengths

- [x] Task 5: Add tests (AC: 7)
  - [x] 5.1: In `internal/adapters/tui/model_test.go`, add test cases:
    - `TestRenderHorizontalSplit_HeightPriority_BelowThreshold` - detail hidden
    - `TestRenderHorizontalSplit_HeightPriority_AtThreshold` - both visible, minimal split
    - `TestRenderHorizontalSplit_HeightPriority_Comfortable` - 60/40 split
    - `TestRenderHorizontalSplit_AnchorStability` - navigation doesn't shift list
  - [x] 5.2: In `internal/shared/styles/styles_test.go`, add test for HorizontalBorderStyle:
    - Verify `BorderTop(false)` is set
    - Verify uses NormalBorder and color "8"
  - [x] 5.3: In `internal/adapters/tui/components/detail_panel_test.go`, add test for horizontal mode:
    - `TestDetailPanel_HorizontalMode_UsesBorderlessTop` - verifies SetHorizontalMode(true) affects rendering
  - [x] 5.4: Run `make test` - all tests pass
  - [x] 5.5: Run `make lint` - no warnings

## Dev Notes

### Current Implementation (renderHorizontalSplit at model.go:1484-1503)

```go
func (m Model) renderHorizontalSplit(height int) string {
    // Calculate heights: 60% list, 40% detail
    listHeight := int(float64(height) * 0.6)
    detailHeight := height - listHeight

    // Create copies with updated sizes - full width for horizontal layout
    projectList := m.projectList
    projectList.SetSize(m.width, listHeight)

    detailPanel := m.detailPanel
    detailPanel.SetSize(m.width, detailHeight)

    // Render stacked vertically
    listView := projectList.View()
    detailView := detailPanel.View()

    return lipgloss.JoinVertical(lipgloss.Left, listView, detailView)
}
```

**Problems with current implementation:**
1. Fixed 60/40 split doesn't adapt to small heights
2. No minimum height guarantees
3. No anchor point enforcement via lipgloss.Height()
4. Uses BorderStyle with full border (wastes 1 line on top)

### Height Priority Algorithm

```
Height < 16: Hide detail, project list gets 100%
Height 16-29: Minimal detail (6 lines), list gets rest
Height >= 30: Standard 60/40 split
```

**Rationale:**
- 10 lines is minimum useful project list (header + ~8 projects visible)
- 6 lines is minimum useful detail (title + 4 fields)
- 30 lines is comfortable for full split

**Note:** View() at model.go:1360-1361 guards MinHeight=20. The `height` parameter passed to `renderHorizontalSplit()` is `contentHeight` which equals `m.height - statusBarHeight()`, so we never receive height < 18 in practice. The HorizontalDetailThreshold=16 is therefore safely below this minimum.

### Anchor Point Fix Pattern

The issue occurs because `lipgloss.JoinVertical()` doesn't enforce fixed heights - it allows content to push siblings. Fix:

```go
// BEFORE: Content can push siblings
listView := projectList.View()
detailView := detailPanel.View()
return lipgloss.JoinVertical(lipgloss.Left, listView, detailView)

// AFTER: Fixed heights prevent content pushing
listContainer := lipgloss.NewStyle().Height(listHeight)
detailContainer := lipgloss.NewStyle().Height(detailHeight)
listView := listContainer.Render(projectList.View())
detailView := detailContainer.Render(detailPanel.View())
return lipgloss.JoinVertical(lipgloss.Left, listView, detailView)
```

### Border Optimization

Current `BorderStyle` uses full rounded border (4 sides). For horizontal stacking, removing top border saves 1 line:

```go
// Current - 4 borders = wastes 2 vertical lines (top + bottom)
BorderStyle.BorderTop(true).BorderBottom(true)

// Optimized - 3 borders = saves 1 vertical line
HorizontalBorderStyle.BorderTop(false).BorderBottom(true)
```

### Architecture Compliance

**Files to Modify:**
- `internal/adapters/tui/views.go` - Add height threshold constants
- `internal/adapters/tui/model.go` - Refactor renderHorizontalSplit()
- `internal/shared/styles/styles.go` - Add HorizontalBorderStyle
- `internal/adapters/tui/components/detail_panel.go` - Add horizontal mode support
- `internal/adapters/tui/model_test.go` - Add height priority tests

**No New Files** - extends existing patterns.

### Previous Story Learnings

**From Story 8.6 (Horizontal Split Layout):**
- Use raw `m.width` in renderHorizontalSplit (not effectiveWidth) - centering handled by outer View()
- Resize handling works automatically via resizeTickMsg
- Default changed to horizontal based on user feedback

**From Story 8.10 (Column Rebalancing):**
- lipgloss.Height() and lipgloss.Width() are powerful for enforcing dimensions
- Percentage-based calculations need min/max bounds

**From Story 8.11 (Stage Re-detection):**
- Config changes require updates to BOTH ports/config.go AND config/loader.go
- Zero vs nil semantics matter

**From Story 8.4 (Layout Width Bugs):**
- Race conditions occur when ProjectsLoadedMsg arrives before WindowSizeMsg
- Component dimensions must be set after m.ready=true

### Key Code Locations

| File | Function/Area | Purpose |
|------|---------------|---------|
| `model.go:1484-1503` | `renderHorizontalSplit()` | Main implementation target |
| `model.go:1445-1458` | `renderMainContent()` | Calls renderHorizontalSplit |
| `views.go:15-18` | `MinWidth`, `MinHeight` | Existing dimension constants |
| `views.go:26` | `HeightThresholdTall` | Pattern for new constants |
| `styles/styles.go:20-32` | `BorderStyle` | Pattern for HorizontalBorderStyle |
| `detail_panel.go:98-178` | `renderProject()` | Uses BorderStyle for rendering |

### Testing Scenarios

**Unit Tests:**
1. Height = 50: Both visible, 60/40 split (30 list, 20 detail)
2. Height = 25: Both visible, minimal split (19 list, 6 detail)
3. Height = 15: Detail hidden, list gets 15 (below threshold)
4. Height = 16: Both visible, minimal split (10 list, 6 detail) - threshold
5. Navigation: Select different projects, verify list doesn't shift

**Manual Tests:**
```bash
# Test different terminal heights
# Use terminal app to resize window to specific heights
./bin/vibe
# Resize to 50 lines - verify 60/40 split
# Resize to 25 lines - verify minimal split
# Resize to 15 lines - verify detail hidden
# Navigate up/down - verify list stays stable
```

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Hardcode height thresholds in renderHorizontalSplit | Use exported constants from views.go |
| Skip lipgloss.Height() enforcement | Always wrap components in fixed-height containers |
| Modify vertical layout behavior | Only change horizontal layout |
| Forget to update detail panel with layout mode | Wire SetHorizontalMode() call |
| Use absolute heights instead of calculated | Calculate based on available height |
| Remove status bar from height calculation | Status bar height already subtracted before renderHorizontalSplit |
| Modify refreshCompleteMsg handler logic | Story 8.11 timer logic must remain intact |
| Use RoundedBorder for HorizontalBorderStyle | Use NormalBorder to match existing BorderStyle |

### References

**Key Files:**
- `internal/adapters/tui/model.go` - Main TUI model with renderHorizontalSplit
- `internal/adapters/tui/views.go` - Height constants and view helpers
- `internal/shared/styles/styles.go` - Shared lipgloss styles
- `internal/adapters/tui/components/detail_panel.go` - Detail panel component

**Related Stories:**
- Story 8.6 - Created horizontal layout mode (this story fixes issues)
- Story 8.10 - Column rebalancing pattern
- Story 3.10 - Responsive layout thresholds
- docs/project-context.md - Project patterns and anti-patterns

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Height Priority (Detail Auto-Hide)

```bash
make build && ./bin/vibe
# Resize terminal to ~15 lines height
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Detail panel | Hidden automatically | Still visible (cropped) |
| Project list | Full height, all visible | Cropped or missing rows |

### Step 2: Threshold Behavior

```bash
# Resize terminal to exactly 16 lines (use terminal settings if possible)
./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Both visible | List ~10 lines, detail ~6 lines | One component missing |
| Proportions | Detail minimized, list maximized | Even split |

### Step 3: Anchor Point Stability

```bash
./bin/vibe
# Press 'd' to ensure detail is visible
# Navigate up/down through projects with j/k
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Navigation | Project list stays at top | List shifts/jumps on navigation |
| Selection | Cursor moves within list | Whole list moves |
| Detail content | Updates independently | Affects list position |

### Step 4: Border Optimization

```bash
./bin/vibe
# Press 'd' to show detail panel
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Detail border | No top border in horizontal mode | Full border (extra line) |
| Vertical mode | Full border (unchanged) | Border missing |

### Step 5: Resize Transitions

```bash
./bin/vibe
# Rapidly resize terminal up and down
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Transitions | Smooth, no flicker | Flicker or layout break |
| Recovery | Returns to correct layout | Stuck in wrong state |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Detail doesn't auto-hide | Check height threshold logic |
| List shifts on navigation | Check lipgloss.Height() enforcement |
| Wrong proportions | Check height calculation algorithm |
| Tests fail | Check constant exports and mock setup |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-8/8-12-horizontal-layout-height-handling.md`
- Project context: `docs/project-context.md`

### Agent Model Used

Claude Opus 4.5

### Debug Log References

N/A - No debug issues encountered

### Completion Notes List

1. **Task 1**: Added height threshold constants (`MinListHeightHorizontal=10`, `MinDetailHeightHorizontal=6`, `HorizontalDetailThreshold=16`) to `views.go`
2. **Task 2**: Implemented height-priority algorithm in `renderHorizontalSplit()`:
   - Height < 16: Hide detail, list gets 100%
   - Height 16-29: Minimal detail (6 lines), list gets rest
   - Height >= 30: Standard 60/40 split
   - Added status bar hint via `SetHeightHint()` when detail auto-hidden
3. **Task 3**: Added `HorizontalBorderStyle` to `styles.go` (no top border) and `SetHorizontalMode()` to `DetailPanelModel`
4. **Task 4**: Fixed anchor point independence using `lipgloss.Height()` to enforce fixed container heights
5. **Task 5**: Added comprehensive tests for all new functionality:
   - 4 model tests for height priority and anchor stability
   - 4 styles tests for HorizontalBorderStyle
   - 2 detail panel tests for horizontal mode

### File List

| File | Changes |
|------|---------|
| `internal/adapters/tui/views.go` | Added 4 height threshold constants (including HorizontalComfortableThreshold) |
| `internal/adapters/tui/model.go` | Refactored `renderHorizontalSplit()` with height-priority logic and anchor enforcement |
| `internal/adapters/tui/components/status_bar.go` | Added `heightHint` field and `SetHeightHint()` method |
| `internal/adapters/tui/components/detail_panel.go` | Added `isHorizontal` field and `SetHorizontalMode()` method, updated `renderProject()` |
| `internal/shared/styles/styles.go` | Added `HorizontalBorderStyle` |
| `internal/adapters/tui/model_test.go` | Added 4 Story 8.12 tests + code review fix M2 |
| `internal/shared/styles/styles_test.go` | Created new file with 4 tests for HorizontalBorderStyle |
| `internal/adapters/tui/components/detail_panel_test.go` | Added 2 horizontal mode tests |
| `internal/adapters/tui/components/status_bar_test.go` | Added 4 height hint tests (code review fix M1) |

## Change Log

- 2025-12-26: Story created from manual testing feedback during Story 8.6 validation
- 2025-12-27: Enriched with comprehensive developer context by SM agent (YOLO mode)
  - Added detailed height priority algorithm
  - Added anchor point fix pattern with lipgloss.Height()
  - Added HorizontalBorderStyle optimization
  - Added specific code snippets for all tasks
  - Added key code locations table
  - Added previous story learnings section
  - Added comprehensive testing scenarios
  - Updated status from backlog to ready-for-dev
- 2025-12-27: SM validation improvements applied:
  - **C1 FIXED:** Added MinHeight guard clarification in Height Priority Algorithm section
  - **C2 FIXED:** Added anti-pattern for preserving Story 8.11 timer logic in refreshCompleteMsg
  - **C3 FIXED:** Corrected styles.go line reference from ~31 to after line 99
  - **E1 FIXED:** Task 3.1 now uses NormalBorder and color "8" to match existing BorderStyle
  - **E2 FIXED:** Task 3.2 now specifies exact line number (line 27) for isHorizontal field
  - **O1 FIXED:** Task 2.2 promoted from optional to REQUIRED per AC1 status bar hint
  - **O2 FIXED:** Added Task 5.3 for detail panel horizontal mode test
  - Added 2 new anti-patterns (refreshCompleteMsg, NormalBorder)
- 2025-12-27: **Dev Agent Implementation Complete**
  - All 5 tasks completed
  - All tests passing, lint clean
  - Status changed to `review`
- 2025-12-27: **Code Review Complete** (Dev Agent as Reviewer)
  - **M1 FIXED:** Added 4 tests for SetHeightHint in status_bar_test.go
  - **M2 FIXED:** Added negative assertion in TestRenderHorizontalSplit_HeightPriority_BelowThreshold
  - **M3 FIXED:** Added height hint logic to pendingProjects processing block for race condition edge case
  - **L1 FIXED:** Added HorizontalComfortableThreshold=30 constant, replaced magic number
  - All fixes applied, tests passing, lint clean
  - Status changed to `done`
