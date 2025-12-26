# Story 8.12: Horizontal Layout Height Handling

Status: backlog

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
- [ ] When terminal height < threshold, project list uses available height
- [ ] Detail panel collapses/hides when insufficient height
- [ ] User can still toggle detail panel with 'd' if they prefer to see details

### AC2: Minimum Height Thresholds
- [ ] Define minimum heights for horizontal mode (e.g., list needs 10+ lines to be useful)
- [ ] Below minimum, either: (a) auto-hide detail, (b) switch to overlay mode, or (c) show "[d] for details" hint
- [ ] Document thresholds in code comments

### AC3: Independent Anchor Points
- [ ] Project list anchored at top - stable position regardless of detail content
- [ ] Detail panel anchored below project list - content can vary without affecting list
- [ ] Navigation between projects doesn't cause project list to shift

### AC4: Reduced Margin/Padding
- [ ] Minimize visual gap between project list and detail panel
- [ ] Consider borderless or minimal border variant for horizontal layout
- [ ] Maintain visual separation without wasting vertical space

### AC5: Scrollable Regions (Nice to Have)
- [ ] Project list scrolls independently if content exceeds allocated height
- [ ] Detail panel scrolls independently for long content
- [ ] Scroll position preserved when navigating

## Technical Notes

### Current Implementation
```go
// renderHorizontalSplit in model.go
listHeight := int(float64(height) * 0.6)  // Fixed 60%
detailHeight := height - listHeight        // Fixed 40%
```

### Proposed Changes

1. **Height calculation with minimum thresholds**:
```go
const (
    MinListHeight = 10   // Minimum lines for project list
    MinDetailHeight = 6  // Minimum lines for detail panel
)

func (m Model) renderHorizontalSplit(height int) string {
    // Priority: list first
    listHeight := height
    detailHeight := 0

    if height >= MinListHeight + MinDetailHeight {
        // Both fit - use 60/40 split
        listHeight = int(float64(height) * 0.6)
        detailHeight = height - listHeight
    } else if height >= MinListHeight {
        // Only list fits - hide detail
        detailHeight = 0
    }
    // ... render logic
}
```

2. **Component rendering with fixed height containers**:
```go
// Use lipgloss.Height() to enforce fixed height per component
listContainer := lipgloss.NewStyle().Height(listHeight)
detailContainer := lipgloss.NewStyle().Height(detailHeight)
```

3. **Borderless detail for horizontal**:
```go
// Create variant of BorderStyle without top border for horizontal layout
horizontalDetailStyle := styles.BorderStyle.BorderTop(false)
```

### Files to Modify
- `internal/adapters/tui/model.go` - renderHorizontalSplit() refactor
- `internal/adapters/tui/components/detail_panel.go` - Add horizontal layout variant
- `internal/shared/styles/styles.go` - Add horizontal-specific styles (optional)
- `internal/adapters/tui/model_test.go` - Tests for new thresholds

## Dependencies

- Story 8.6 (Horizontal Split Layout Option) - DONE

## Test Cases

1. Height = 50 lines: Both components visible, 60/40 split
2. Height = 30 lines: Both visible, but detail may be smaller
3. Height = 15 lines: Project list gets priority, detail hidden or minimal
4. Height = 10 lines: Project list only (minimum viable)
5. Navigate between projects with different detail lengths: List stable
6. Resize terminal rapidly: Layout adapts without visual glitches

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Unit tests for height threshold logic
- [ ] Manual testing confirms:
  - [ ] Project list always visible and stable
  - [ ] No cropping when navigating
  - [ ] Smooth degradation at small heights
- [ ] Code review passed

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

(To be filled by Dev Agent)

### Debug Log References

(To be filled by Dev Agent)

### Completion Notes List

(To be filled by Dev Agent)

### File List

(To be filled by Dev Agent)
