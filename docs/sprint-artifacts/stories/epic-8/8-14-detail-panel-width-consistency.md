# Story 8.14: Detail Panel Width Consistency

Status: done

**Priority: P3 Cosmetic**

## Story

As a **user toggling the detail panel on and off**,
I want **the content area width to remain consistent**,
So that **the layout doesn't shift unexpectedly and distract me from the dashboard content**.

## Background

When toggling the detail panel (pressing 'd'), the overall content area width shifts. The detail view content appears wider than the project list, causing a visual "jump" in the layout.

**Root Cause Identified:**

The width used for the detail panel doesn't consistently respect the `maxContentWidth` constraint:

1. Project list uses cached width `m.projectList.Width()` at lines 1476, 1518, 1535, which differs from effective width
2. `renderMainContent()` at line 1485 returns `m.projectList.View()` without calling `SetSize()` first
3. `renderDashboard()` creates `renderModel` with capped width, but child methods don't use it consistently

**Key Insight:** `renderDashboard()` already creates `renderModel` with `renderModel.width = effectiveWidth`. The fix is to ensure child render methods use `m.width` (which IS the capped width when called on renderModel) instead of accessing cached component widths.

**TL;DR Fix Summary:**
- Change 3 instances of `m.projectList.Width()` to `m.width`: Lines 1476, 1518, 1535
- Add 1 SetSize call before `.View()`: Line 1485

## Acceptance Criteria

### AC1: Consistent Width When Toggling Detail
- Given `max_content_width` is configured (e.g., 120)
- When user presses 'd' to toggle detail panel
- Then the content area width remains the same
- And no visual shift/jump occurs

### AC2: Horizontal Layout Uses Capped Width
- Given horizontal layout mode with detail panel visible
- When terminal width exceeds `max_content_width`
- Then both project list AND detail panel use capped width
- And they are visually aligned

### AC3: Vertical Layout Uses Capped Width
- Given vertical (side-by-side) layout mode with detail panel visible
- When terminal width exceeds `max_content_width`
- Then both components fit within capped width
- And no overflow occurs

### AC4: Project List Width Matches Detail
- Given detail panel is visible
- When comparing project list width to detail panel width
- Then both use the same effective width
- And there's no mismatch between components

### AC5: Story 8.12 Height Logic Preserved
- Given horizontal layout with insufficient height
- When height < HorizontalDetailThreshold (16)
- Then detail panel is hidden (existing behavior)
- And only project list is shown

### AC6: Test Coverage
- Given all changes are made
- When `make test && make lint` runs
- Then all tests pass and no lint errors

## Tasks / Subtasks

- [x] Task 1: Fix renderHorizontalSplit width usage (AC: 2, 4, 5)
  - [x] 1.1: Line 1526 uses `m.width` correctly - NO CHANGE needed
  - [x] 1.2: At line 1535, change `projectList.SetSize(m.projectList.Width(), listHeight)` to `projectList.SetSize(m.width, listHeight)` - use receiver's width, not cached
  - [x] 1.3: At line 1518 (height < threshold case), change from:
    ```go
    return m.projectList.View()
    ```
    To:
    ```go
    projectList := m.projectList
    projectList.SetSize(m.width, height)
    return projectList.View()
    ```
  - [x] 1.4: Verify height-priority logic at lines 1516-1517 is unchanged (only modifying line 1518)

- [x] Task 2: Fix renderMainContent project-list-only case (AC: 1, 4)
  - [x] 2.1: At lines 1484-1485, change from:
    ```go
    if !m.showDetailPanel {
        return m.projectList.View()
    }
    ```
    To:
    ```go
    if !m.showDetailPanel {
        projectList := m.projectList
        projectList.SetSize(m.width, height)  // Use receiver's width AND height
        return projectList.View()
    }
    ```
  - [x] 2.2: At line 1476 (hint case), change from:
    ```go
    projectList.SetSize(m.projectList.Width(), height-1)
    ```
    To:
    ```go
    projectList.SetSize(m.width, height-1)  // Use receiver's width, not cached
    ```

- [x] Task 3: Verify vertical layout (AC: 3)
  - [x] 3.1: Lines 1495-1496 calculate widths FROM `m.width`:
    ```go
    listWidth := int(float64(m.width) * 0.6)
    detailWidth := m.width - listWidth - 1
    ```
  - [x] 3.2: Lines 1500 and 1503 use those calculated widths in SetSize() - this is correct
  - [x] 3.3: No changes needed - vertical layout already uses `m.width` correctly

- [x] Task 4: Add tests (AC: 6)
  - [x] 4.1: In `model_responsive_test.go`, add test for width consistency:
    - `TestRenderHorizontalSplit_UsesReceiverWidth` - verify detailPanel and projectList use m.width
  - [x] 4.2: Add test for project-list-only case:
    - `TestRenderMainContent_ProjectListOnlyUsesReceiverWidth`
  - [x] 4.3: Add test for height < threshold case in renderHorizontalSplit:
    - `TestRenderHorizontalSplit_BelowThreshold_UsesReceiverWidth` - verify projectList uses m.width even when detail hidden
  - [x] 4.4: Run `make test` - all tests pass
  - [x] 4.5: Run `make lint` - no warnings

## Dev Notes

### Current Implementation Analysis (Verified Line Numbers)

**renderDashboard() at model.go:1423-1466:**
```go
func (m Model) renderDashboard() string {
    // Lines 1426-1429: Calculate effective width
    effectiveWidth := m.width
    if m.isWideWidth() {
        effectiveWidth = m.maxContentWidth
    }

    // Lines 1439-1443: Create copy with capped width
    renderModel := m
    renderModel.width = effectiveWidth

    mainContent := renderModel.renderMainContent(contentHeight)
    // ...
}
```

**The renderModel Pattern:** When `renderModel.renderMainContent()` is called, `m.width` inside that method IS the effectiveWidth. No parameter passing needed - just use `m.width` consistently.

**renderHorizontalSplit() at model.go:1517-1543 (FIXED):**
```go
func (m Model) renderHorizontalSplit(height int) string {
    // Lines 1518-1522: Height priority check with SetSize fix
    if height < HorizontalDetailThreshold {
        projectList := m.projectList
        projectList.SetSize(m.width, height)  // FIXED - uses receiver's width
        return projectList.View()
    }

    // Line 1530: Uses m.width correctly
    detailPanel.SetSize(m.width, 0)

    // Line 1539: FIXED - Uses m.width instead of cached
    projectList.SetSize(m.width, listHeight)
}
```

**renderMainContent() at model.go:1469-1512 (FIXED):**
```go
func (m Model) renderMainContent(height int) string {
    // Lines 1473-1480: Hint case - FIXED
    if m.height >= MinHeight && m.height < HeightThresholdTall && !m.showDetailPanel {
        projectList := m.projectList
        projectList.SetSize(m.width, height-1)  // FIXED - uses receiver's width
    }

    // Lines 1484-1488: Project list only case - FIXED
    if !m.showDetailPanel {
        projectList := m.projectList
        projectList.SetSize(m.width, height)  // FIXED - uses receiver's width
        return projectList.View()
    }
    // ...
}
```

### Fix Strategy (Simpler Than Original Proposal)

**No new parameters needed.** The `renderModel` pattern already sets `m.width = effectiveWidth`. Just ensure all render methods use `m.width` instead of cached component widths.

```go
// WRONG: Using cached component width
projectList.SetSize(m.projectList.Width(), height)

// CORRECT: Using receiver's width (already capped in renderModel)
projectList.SetSize(m.width, height)
```

### Architecture Compliance

- **Single file change:** `model.go` only
- **No interface changes:** No method signature changes
- **No behavior change:** Just fixes visual consistency
- **Preserves Story 8.12:** Height-priority logic at lines 1516-1519 unchanged
- **Uses existing pattern:** Leverages renderModel width capping

### Key Code Locations (Verified)

| File | Lines | Purpose | Change Needed |
|------|-------|---------|---------------|
| `model.go:1426-1429` | effectiveWidth calculation | None - source of capped width |
| `model.go:1439-1443` | renderModel creation | None - already caps width |
| `model.go:1476` | Hint case SetSize | Change `m.projectList.Width()` to `m.width` |
| `model.go:1484-1485` | Project list only case | Add SetSize with `m.width, height` |
| `model.go:1516-1517` | Height priority check | None - preserve exactly |
| `model.go:1518` | Below threshold return | Add SetSize with `m.width, height` before View() |
| `model.go:1526` | detailPanel.SetSize | Already uses `m.width` - no change |
| `model.go:1535` | projectList.SetSize | Change `m.projectList.Width()` to `m.width` |

**Summary of Changes (4 locations):**
1. Line 1476: `m.projectList.Width()` → `m.width`
2. Line 1485: Add `projectList.SetSize(m.width, height)` before View()
3. Line 1518: Add `projectList.SetSize(m.width, height)` before View()
4. Line 1535: `m.projectList.Width()` → `m.width`

### Previous Story Learnings

**From Story 8.10 (Column Rebalancing):**
- Added `maxContentWidth` config and `effectiveWidth` calculation
- Established the `renderModel` pattern for width capping in `renderDashboard()`
- This story should LEVERAGE this pattern, not duplicate it

**From Story 8.12 (Horizontal Layout Height):**
- Added height-priority algorithm at lines 1516-1519
- If height < HorizontalDetailThreshold, show only list
- **CRITICAL:** This logic MUST be preserved - do not modify height handling

**From Story 8.4 (Race Condition Fix):**
- Added `m.projectList.Width()` calls which cache initial width
- This caching is the ROOT CAUSE of the width inconsistency bug

### Anti-Patterns to Avoid

| Don't | Do Instead | Why |
|-------|------------|-----|
| Use `m.projectList.Width()` for sizing | Use `m.width` consistently | Cached width may differ from effective width |
| Add new parameters to render methods | Use existing `m.width` from receiver | renderModel already caps width |
| Modify effectiveWidth calculation | Just use `m.width` in child methods | Calculation is correct, usage is not |
| Touch lines 1516-1517 (the IF condition) | Preserve height-priority exactly | Story 8.12 height logic must remain |
| Modify `renderDashboard()` | Fix child methods only | The parent is correct |
| Return `m.projectList.View()` directly | Always call `SetSize()` first | View() uses cached dimensions |

### Testing Scenarios

**Unit Tests:**
1. `renderHorizontalSplit` normal case: Create model with width=200, maxContentWidth=120. Verify components sized to 120, not 200.
2. `renderHorizontalSplit` below threshold: Create model with height < 16. Verify projectList still uses m.width (not cached).
3. `renderMainContent` with `!showDetailPanel`: Verify projectList gets SetSize with model width before View().
4. Height priority preserved: height < 16 → only list shown (existing behavior must work)

**Manual Tests:**
```bash
make build && ./bin/vibe
# Set terminal to wide width (> 120 cols)
# Press 'd' to toggle detail - NO SHIFT should occur
# Try both horizontal and vertical layouts ('L' key)
# Verify no visual jump when toggling
```

### References

- Investigation report from Explore agent
- Story 8.10 - Introduced maxContentWidth and effectiveWidth
- Story 8.12 - Modified renderHorizontalSplit() height handling (must preserve)
- Story 8.4 - Introduced cached width pattern (root cause of bug)
- `internal/adapters/tui/model.go` - All changes in this file

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Toggle Detail Panel

```bash
make build && ./bin/vibe
# Ensure terminal is wider than 120 columns
# Press 'd' multiple times to toggle detail panel
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Content width | Stays constant | Shifts/jumps |
| Project list | Same width with/without detail | Width changes |
| Visual alignment | Clean, no jitter | Layout shifts |

### Step 2: Horizontal Layout

```bash
./bin/vibe
# Horizontal layout is default
# Press 'd' to show detail
# Compare widths visually - list and detail should align
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| List width | Same as detail width | Different widths |
| Alignment | Left edges align | Misaligned |

### Step 3: Verify Height Priority Still Works

```bash
# Resize terminal to short height (< 20 rows)
./bin/vibe
# Press 'd' - detail should be hidden due to insufficient height
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Short terminal | Only list visible | Detail shown despite no space |
| Status bar hint | "[d] Detail hidden..." message | No hint |

### Step 4: Vertical Layout

```bash
./bin/vibe
# Press 'L' for vertical (side-by-side)
# Press 'd' to toggle detail
# Compare widths - list and detail should fit within max_content_width
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Total width | Respects max_content_width | Overflow |
| Proportions | 60/40 split within bounds | Wrong ratios |
| Toggle | No shift when toggling 'd' | Width jumps |

### Step 5: Both Layouts at Wide Width

```bash
./bin/vibe  # In wide terminal (> 150 cols)
# Toggle between layouts with 'L'
# Toggle detail with 'd' in each layout
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Horizontal | Content centered, respects cap | Stretches full width |
| Vertical | Content centered, respects cap | Stretches full width |

### Decision Guide

| Situation | Action |
|-----------|--------|
| No shift when toggling | Mark `done` |
| Shift still occurs | Check m.width usage in render methods |
| Tests fail | Review SetSize calls |
| Height logic broken | Revert lines 1516-1517 condition, check line 1518 fix |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-8/8-14-detail-panel-width-consistency.md`
- Project context: `docs/project-context.md`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- Test initially showed 360 char lines - traced to UTF-8 box-drawing characters (3 bytes each)
- Border line: 120 display columns × 3 bytes = 360 bytes
- Not a bug - just byte count vs display width difference

### Completion Notes List

1. **Task 1.2**: Changed `m.projectList.Width()` to `m.width` at line 1535 in `renderHorizontalSplit`
2. **Task 1.3**: Added `SetSize(m.width, height)` before `View()` at line 1518 (below-threshold case)
3. **Task 2.1**: Added `SetSize(m.width, height)` before `View()` at line 1485 (detail panel hidden case)
4. **Task 2.2**: Changed `m.projectList.Width()` to `m.width` at line 1476 (hint case)
5. **Task 3**: Verified vertical layout already uses `m.width` correctly - no changes needed
6. **Task 4**: Added 3 tests to `model_responsive_test.go`:
   - `TestRenderHorizontalSplit_UsesReceiverWidth`
   - `TestRenderMainContent_ProjectListOnlyUsesReceiverWidth`
   - `TestRenderHorizontalSplit_BelowThreshold_UsesReceiverWidth`

### File List

| File | Change Type |
|------|-------------|
| `internal/adapters/tui/model.go` | Modified - 4 width fixes |
| `internal/adapters/tui/model_responsive_test.go` | Modified - 3 new tests |
| `docs/sprint-artifacts/stories/epic-8/8-14-detail-panel-width-consistency.md` | Modified - status + tasks |

## Change Log

- 2025-12-31: Story created by SM agent
  - Root cause: effectiveWidth not passed consistently to render methods
  - Fix: Add effectiveWidth parameter to renderMainContent, renderHorizontalSplit, renderVerticalSplit
  - Affects model.go only - single file fix
- 2025-12-31: Story validated and improved by SM agent (first pass)
  - Corrected line numbers to match current model.go
  - Simplified fix strategy: use existing renderModel pattern instead of adding parameters
  - Identified actual bugs: cached `m.projectList.Width()` usage at lines 1535 and 1475
  - Added AC5 to preserve Story 8.12 height-priority logic
  - Added anti-pattern: don't touch lines 1516-1519
  - Updated Dev Notes with accurate code snippets from current implementation
- 2025-12-31: Fresh re-validation by SM agent (second pass) - 3 critical issues found
  - **C1 FIXED:** Added missing bug at line 1518 - `m.projectList.View()` uses cached dimensions
  - **C2 FIXED:** Line 1485 needs both width AND height in SetSize
  - **C3 FIXED:** Clarified Task 3 that lines 1495-1496 calculate widths, actual SetSize at 1500/1503
  - **E1 ADDED:** New test case for line 1518 below-threshold scenario
  - **E2 FIXED:** Made vertical layout testing mandatory (Step 4), added Step 5 for wide width testing
  - **O1 ADDED:** TL;DR Fix Summary in Background section
  - **O2 ADDED:** Summary of Changes table in Key Code Locations
  - Updated Task 1 with new subtask 1.3 for line 1518 fix
  - Updated Anti-Patterns with "Never return m.projectList.View() directly"
  - Updated Decision Guide to reference line 1518 fix specifically
- 2025-12-31: Code review by Dev Agent (Amelia) - 3 MEDIUM, 2 LOW issues found and fixed
  - **M1 FIXED:** Test assertions too shallow - added width verification to all 3 tests
  - **M2 FIXED:** Missing edge case test - added `TestRenderMainContent_HintCase_UsesCappedWidth`
  - **L1 FIXED:** Improved test naming for clarity (e.g., `_UsesCappedWidth` suffix)
  - **L2 FIXED:** Updated Dev Notes line numbers to match current implementation
  - All tests pass, lint clean
  - Story status: `done`
