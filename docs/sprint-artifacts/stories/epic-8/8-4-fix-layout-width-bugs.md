# Story 8.4: Fix Layout Width Bugs

Status: done

## Story

As a **user**,
I want **consistent full-width layout**,
So that **the dashboard uses all available terminal space**.

## Acceptance Criteria

1. **AC1: Full Width on Initial Launch**
   - Given vibe is launched
   - When the initial render completes
   - Then the layout uses full terminal width (no margins)

2. **AC2: Full Width Detail Panel**
   - Given detail panel is opened
   - When the panel renders
   - Then it uses full available width (no margins)

3. **AC3: Full Width After Resize**
   - Given any terminal resize
   - When re-render occurs
   - Then full width is maintained

4. **AC4: Consistent Width Across All States**
   - Given any view state (empty, loading, projects, detail open/closed)
   - When rendered
   - Then all views use full width consistently

## Tasks / Subtasks

- [x] Task 1: Fix ProjectsLoadedMsg race condition (AC: 1, 4) **CRITICAL**
  - [x] 1.1: Add `pendingProjects []*domain.Project` field to Model struct (model.go:40 area)
  - [x] 1.2: In `ProjectsLoadedMsg` handler (model.go:523-593): Check if `!m.ready`, store in `m.pendingProjects`, return early
  - [x] 1.3: In `resizeTickMsg` handler (model.go:437-472): After setting `m.ready = true`, check for `m.pendingProjects != nil`
  - [x] 1.4: If pending projects exist, create components using `effectiveWidth` (not raw `m.width`)
  - [x] 1.5: Clear `m.pendingProjects = nil` after processing

- [x] Task 2: Fix effectiveWidth in ProjectsLoadedMsg (AC: 1, 4) **CRITICAL**
  - [x] 2.1: When creating projectList/detailPanel, use effectiveWidth calculation: `min(m.width, MaxContentWidth)`
  - [x] 2.2: Replace `m.width` with `effectiveWidth` in `NewProjectListModel()` call (model.go:537)
  - [x] 2.3: Replace `m.width` with `effectiveWidth` in `NewDetailPanelModel()` call (model.go:543)

- [x] Task 3: Add re-size trigger after ready (AC: 1, 3)
  - [x] 3.1: In `resizeTickMsg`, ALWAYS call `SetSize()` on components if they exist (remove `len(m.projects) > 0` guard or expand it)
  - [x] 3.2: Check for zero-valued projectList: `if m.projectList.Width() > 0 { m.projectList.SetSize(...) }`

- [x] Task 4: Fix detail panel border width (AC: 2)
  - [x] 4.1: Review `detail_panel.go` View() method for `BorderStyle.Width(m.width-2)` pattern
  - [x] 4.2: Ensure detail panel width is passed correctly from `renderMainContent()` split calculation

- [x] Task 5: Add tests (AC: all)
  - [x] 5.1: Test: `ProjectsLoadedMsg` before `resizeTickMsg` stores in pendingProjects
  - [x] 5.2: Test: `resizeTickMsg` after pending projects creates components with correct width
  - [x] 5.3: Test: Components get correct effectiveWidth on wide terminals (>MaxContentWidth)
  - [x] 5.4: Test: Resize after projects loaded updates all component widths

## Dev Notes

### Root Cause (CONFIRMED)

**The bug is a race condition between `WindowSizeMsg` and `ProjectsLoadedMsg`:**

In `model.go:537`, when projects are loaded:
```go
m.projectList = components.NewProjectListModel(m.projects, m.width, contentHeight)
```

If `m.width` is still 0 (WindowSizeMsg not yet processed), the delegate gets width=0, causing layout issues. Even after `resizeTickMsg` fires, the component was already created with wrong width.

**Secondary issue:** The `resizeTickMsg` handler (lines 467-470) only updates sizes `if len(m.projects) > 0`, but if projects loaded BEFORE resize, this branch helps nothing.

### Fix Strategy: Delayed Component Creation

Modify the flow so components are only created AFTER `m.ready = true`:

**Step 1: Add pending state (model.go struct around line 40):**
```go
// Layout width fix - Story 8.4
pendingProjects []*domain.Project // Projects waiting for ready state
```

**Step 2: Update ProjectsLoadedMsg handler (model.go:523):**
```go
case ProjectsLoadedMsg:
    m.isLoading = false
    m.statusBar.SetLoading(false)

    if msg.err != nil {
        slog.Error("Failed to load projects", "error", msg.err)
        m.projects = nil
        return m, nil
    }

    // Story 8.4: Defer component creation until ready
    if !m.ready {
        m.pendingProjects = msg.projects
        return m, nil
    }

    // Continue with normal processing (existing code)
    m.projects = msg.projects
    // ... rest of handler
```

**Step 3: Update resizeTickMsg handler (model.go:437, after setting ready=true):**
```go
case resizeTickMsg:
    if m.hasPendingResize {
        wasReady := m.ready
        m.width = m.pendingWidth
        // ... existing code ...
        m.ready = true

        // Story 8.4: Process pending projects now that we have dimensions
        if m.pendingProjects != nil {
            m.projects = m.pendingProjects
            m.pendingProjects = nil

            // Calculate effectiveWidth (same pattern as line 455-458)
            effectiveWidth := m.width
            if isWideWidth(m.width) {
                effectiveWidth = MaxContentWidth
            }
            contentHeight := m.height - statusBarHeight(m.height)

            // Create components with correct dimensions
            m.projectList = components.NewProjectListModel(m.projects, effectiveWidth, contentHeight)
            m.projectList.SetDelegateWaitingCallbacks(m.isProjectWaiting, m.getWaitingDuration)

            m.detailPanel = components.NewDetailPanelModel(effectiveWidth, contentHeight)
            m.detailPanel.SetProject(m.projectList.SelectedProject())
            m.detailPanel.SetVisible(m.showDetailPanel)
            m.detailPanel.SetWaitingCallbacks(m.isProjectWaiting, m.getWaitingDuration)

            // Update status bar
            active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
            m.statusBar.SetCounts(active, hibernated, waiting)

            // Start file watcher (copy from ProjectsLoadedMsg)
            // ... file watcher code ...
        }

        // Update existing component sizes (always, not just if projects > 0)
        if m.projectList != (components.ProjectListModel{}) {
            m.projectList.SetSize(effectiveWidth, contentHeight)
            m.detailPanel.SetSize(effectiveWidth, contentHeight)
        }
    }
```

### Key Code Locations

| Fix Location | File:Line | Change |
|--------------|-----------|--------|
| Model struct | model.go:40 | Add `pendingProjects` field |
| ProjectsLoadedMsg | model.go:523 | Check `!m.ready`, defer to pending |
| resizeTickMsg | model.go:437 | Process pending after ready |
| Component size update | model.go:467-470 | Remove/expand `len(m.projects) > 0` guard |

### Architecture Compliance

- **Modify:** `internal/adapters/tui/model.go` (initialization flow, pending state)
- **No new files** - fixes existing code only
- **No core changes** - TUI adapter layer only

### Previous Story Learnings

**From Story 8.1, 8.2, 8.3:**
- Code review catches edge cases - be thorough with tests
- Bubble Tea's tick/msg pattern requires careful state management
- Race conditions between async messages are common bug sources

**From Story 3.10 (responsive layout):**
- `effectiveWidth = min(m.width, MaxContentWidth)` pattern
- `statusBarHeight(m.height)` returns 1 or 2 based on height

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Create components before `m.ready = true` | Store in pending, process after ready |
| Use raw `m.width` for wide terminals | Use `effectiveWidth` with MaxContentWidth cap |
| Guard SetSize with `len(m.projects) > 0` | Check for zero-value component instead |
| Skip nil/zero-value checks | Guard all component access |

### Testing Strategy

```go
func TestModel_PendingProjectsProcessing(t *testing.T) {
    // Verify projects stored as pending when !ready
    model := NewModel(...)
    model.Update(ProjectsLoadedMsg{projects: []*domain.Project{...}})
    if model.pendingProjects == nil {
        t.Error("should store projects as pending before ready")
    }
    if len(model.projects) > 0 {
        t.Error("should not process projects before ready")
    }
}

func TestModel_PendingProcessedOnResize(t *testing.T) {
    // Verify pending projects processed after resizeTickMsg
    model := NewModel(...)
    model.Update(ProjectsLoadedMsg{projects: []*domain.Project{...}})
    model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
    // Wait for debounce
    model.Update(resizeTickMsg{})

    if model.pendingProjects != nil {
        t.Error("pending should be cleared after processing")
    }
    if len(model.projects) == 0 {
        t.Error("projects should be populated")
    }
    if model.projectList.width != 120 {
        t.Errorf("projectList should have width 120, got %d", model.projectList.width)
    }
}
```

### References

| Document | Relevance |
|----------|-----------|
| docs/project-context.md | Story Completion - User verification required |
| internal/adapters/tui/model.go:437-472 | resizeTickMsg handler |
| internal/adapters/tui/model.go:523-593 | ProjectsLoadedMsg handler |
| internal/adapters/tui/components/project_list.go:21 | NewProjectListModel |

## User Testing Guide

**Time needed:** 3-5 minutes

### Step 1: Initial Launch Check

```bash
make build && ./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Content alignment | Full width from column 0 | Visible margins on left/right |
| Project list rows | Extend to terminal edge | Truncated or indented |
| Status bar | Full width at bottom | Gaps on sides |

### Step 2: Refresh Compare

```bash
# While vibe is running:
# Press 'r' to refresh
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Width before/after | No change (already full) | Width changes after refresh |
| Content alignment | Consistent full width | Layout jumps on refresh |

### Step 3: Detail Panel Check

```bash
# Press 'd' to toggle detail panel
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Panel alignment | Full width on its portion | Margins inside panel |
| Split ratio | 60%/40% visible | Uneven or gapped |
| Panel content | Text to border edges | Text indented from border |

### Step 4: Resize Check

```bash
# Manually resize terminal window
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Layout adaptation | Immediate full-width | Temporary margins |
| All components | Update together | Some parts lag behind |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Initial launch has margins | Check pendingProjects implementation |
| Detail panel has margins | Check renderMainContent split calculation |
| Resize causes margins | Check resizeTickMsg SetSize calls |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- **Task 1**: Added `pendingProjects []*domain.Project` field to Model struct. Modified `ProjectsLoadedMsg` handler to check `!m.ready` and store projects in pending state instead of immediately processing. Modified `resizeTickMsg` handler to process pending projects after setting `m.ready = true`, creating components with correct effectiveWidth.

- **Task 2**: Updated `ProjectsLoadedMsg` handler to use `effectiveWidth` (min of m.width and MaxContentWidth) when creating projectList and detailPanel components, preventing wide terminals from getting oversized components.

- **Task 3**: Changed the component size update guard from `len(m.projects) > 0` to `m.projectList.Width() > 0` to properly detect initialized vs zero-value components. Added `Width()` method to ProjectListModel.

- **Task 4**: Reviewed detail_panel.go - existing implementation correctly uses `Width(m.width-2)` for border. The fix in Tasks 1-3 ensures the width passed to components is correct from the start.

- **Task 5**: Added 7 comprehensive tests covering:
  - `TestModel_ProjectsLoadedBeforeReady_StorePending` - Verifies race condition handling
  - `TestModel_ResizeTickProcessesPending` - Verifies pending projects processed correctly
  - `TestModel_EffectiveWidth_WideTerminal` - Verifies MaxContentWidth cap
  - `TestModel_ResizeAfterReady_UpdatesComponents` - Verifies resize updates existing components
  - `TestModel_ProjectsLoadedAfterReady_UsesEffectiveWidth` - Verifies effectiveWidth used after ready
  - `TestModel_ResizeWithZeroProjects_NoSetSize` - Verifies no SetSize on uninitialized components
  - `TestModel_FullWidthAfterRace_Integration` - End-to-end race condition test

### File List

- `internal/adapters/tui/model.go` - Added pendingProjects field, modified ProjectsLoadedMsg and resizeTickMsg handlers
- `internal/adapters/tui/components/project_list.go` - Added Width() getter method
- `internal/adapters/tui/model_test.go` - Added 7 tests for Story 8.4

### Change Log

- 2025-12-26: Story 8.4 implementation complete - Fixed layout width race condition bug
- 2025-12-26: Code review fixes applied:
  - H1: Extracted duplicated file watcher code to `startFileWatcherForProjects()` helper
  - H2: Enhanced comments to document race condition WHY, not just WHAT
  - H3: Added test `TestModel_ResizeTickProcessesPending_FileWatcher`
  - M1: Added test `TestModel_ResizeTickProcessesPending_DetailPanelDimensions`
  - M3: Enhanced comments in model.go to explain race condition context

