# Story 16.3: Create Stats View TUI Component

Status: done

## Story

As a user,
I want a dedicated full-screen Stats View,
So that I can focus on project metrics without dashboard clutter.

## User-Visible Changes

- **New:** Press `'s'` key from Dashboard to open full-screen Stats View
- **New:** Stats View displays header "STATS" with placeholder content area
- **New:** Press `Esc` or `'q'` in Stats View to return to Dashboard (preserving project selection)
- **New:** Status bar shows `[s] stats` hint in normal dashboard mode

## Acceptance Criteria

1. **Given** I press `'s'` from Dashboard
   **When** Stats View opens
   **Then** it shows full-screen metrics view with "STATS" header

2. **And** given I press `Esc` or `'q'` in Stats View
   **When** handling key
   **Then** it returns to Dashboard with original project selection preserved

3. **And** Stats View has header "STATS" with `[ESC] Back to Dashboard` hint

4. **And** Stats View renders placeholder content without accessing metrics.db
   - This story creates the view shell and navigation
   - Data fetching and display added in Story 16.4+

5. **And** Stats View respects terminal dimensions (responsive layout)

## Tasks / Subtasks

- [x] Task 1: Add viewModeStats to viewMode enum (AC: #1)
  - [x] In `internal/adapters/tui/validation.go` (lines 14-22), add to const block:
    ```go
    viewModeStats   // Story 16.3: Stats View
    ```

- [x] Task 2: Add Stats View state to Model (AC: #1, #2)
  - [x] In `internal/adapters/tui/model.go` Model struct (lines 38-174), add fields:
    ```go
    statsViewScroll       int  // Scroll position in stats view
    statsActiveProjectIdx int  // Saved dashboard selection for restoration
    ```
  - [x] Add `enterStatsView()` method:
    ```go
    func (m *Model) enterStatsView() {
        m.statsActiveProjectIdx = m.projectList.Index()
        m.viewMode = viewModeStats
        m.statsViewScroll = 0
    }
    ```
  - [x] Add `exitStatsView()` method with bounds check:
    ```go
    func (m *Model) exitStatsView() {
        m.viewMode = viewModeNormal
        if m.statsActiveProjectIdx >= 0 && m.statsActiveProjectIdx < len(m.projects) {
            m.projectList.SelectByIndex(m.statsActiveProjectIdx)
        }
    }
    ```

- [x] Task 3: Add Stats key binding (AC: #1)
  - [x] In `internal/adapters/tui/keys.go`, add constant (lines 5-35):
    ```go
    KeyStats = "s"
    ```
  - [x] Add to KeyBindings struct (lines 39-69):
    ```go
    Stats string
    ```
  - [x] In `DefaultKeyBindings()` (lines 72-104):
    ```go
    Stats: KeyStats,
    ```

- [x] Task 4: Implement Stats View key handler (AC: #2)
  - [x] In `internal/adapters/tui/model.go`, add method:
    ```go
    func (m Model) handleStatsViewKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
        switch msg.String() {
        case KeyEscape, KeyQuit:
            m.exitStatsView()
            return m, nil
        case KeyDown, KeyDownArrow:
            // Future story: scroll down
            return m, nil
        case KeyUp, KeyUpArrow:
            // Future story: scroll up
            return m, nil
        }
        return m, nil
    }
    ```

- [x] Task 5: Route Stats key handler in Update (AC: #1, #2)
  - [x] In `Update()` method (line ~857), add after text view check:
    ```go
    if m.viewMode == viewModeStats {
        return m.handleStatsViewKeyMsg(msg)
    }
    ```

- [x] Task 6: Add Stats View entry point (AC: #1)
  - [x] In `handleKeyMsg()` (line ~1758), add case:
    ```go
    case KeyStats:
        if m.showHelp || m.isEditingNote || m.isConfirmingRemove {
            return m, nil
        }
        m.enterStatsView()
        return m, nil
    ```

- [x] Task 7: Implement Stats View render (AC: #3, #4, #5)
  - [x] Create `internal/adapters/tui/statsview.go` with render function:
    ```go
    func (m Model) renderStatsView() string {
        // Calculate effective width
        effectiveWidth := m.width
        if m.isWideWidth() {
            effectiveWidth = m.maxContentWidth
        }

        // Header
        title := lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("99")).
            Render("STATS")
        hint := lipgloss.NewStyle().
            Foreground(lipgloss.Color("241")).
            Render("[ESC] Back to Dashboard")
        header := lipgloss.JoinHorizontal(lipgloss.Top,
            title,
            strings.Repeat(" ", effectiveWidth-lipgloss.Width(title)-lipgloss.Width(hint)),
            hint,
        )

        // Content height (account for header and status bar)
        contentHeight := m.height - statusBarHeight(m.height) - 2
        content := lipgloss.NewStyle().
            Width(effectiveWidth).
            Height(contentHeight).
            Align(lipgloss.Center, lipgloss.Center).
            Foreground(lipgloss.Color("241")).
            Render("Project metrics will appear here")

        return lipgloss.JoinVertical(lipgloss.Left, header, content)
    }
    ```

- [x] Task 8: Wire Stats View into View() method (AC: #3)
  - [x] In `View()` method (line ~2120), add after text view check:
    ```go
    if m.viewMode == viewModeStats {
        return m.renderStatsView() + "\n" + m.statusBar.View()
    }
    ```

- [x] Task 9: Update status bar hints (AC: #1)
  - [x] In `internal/adapters/tui/components/status_bar.go` (line ~249), modify:
    ```go
    // From:
    shortcutsFull = "│ [j/k] nav [d] details [f] fav [r] refresh [?] help [q] quit │"
    // To:
    shortcutsFull = "│ [j/k] nav [d] details [s] stats [f] fav [r] refresh [?] help [q] quit │"

    // Also update abbreviated version:
    shortcutsAbbrev = "│ [j/k] [d] [s] [f] [r] [?] [q] │"
    ```

- [x] Task 10: Write unit tests
  - [x] Create `internal/adapters/tui/statsview_test.go`:
    ```go
    func TestEnterStatsView_CapturesSelection(t *testing.T) { ... }
    func TestExitStatsView_RestoresSelection(t *testing.T) { ... }
    func TestExitStatsView_BoundsCheck(t *testing.T) { ... }
    func TestHandleStatsViewKeyMsg_EscExits(t *testing.T) { ... }
    func TestHandleStatsViewKeyMsg_QExits(t *testing.T) { ... }
    ```

## Dev Notes

### Architecture Alignment

Implements **FR-P2-14** and **FR-P2-15** from Phase 2 PRD.

**Isolation Principle:** Stats View is structurally independent - can be removed without affecting core dashboard. Data access deferred to Story 16.4+.

### File Modifications

| File | Changes |
|------|---------|
| `internal/adapters/tui/validation.go:14-22` | Add `viewModeStats` to enum |
| `internal/adapters/tui/model.go:38-174` | Add state fields to Model struct |
| `internal/adapters/tui/model.go:~857` | Add routing in Update() |
| `internal/adapters/tui/model.go:~1758` | Add entry case in handleKeyMsg() |
| `internal/adapters/tui/model.go:~2120` | Add render branch in View() |
| `internal/adapters/tui/keys.go:5-104` | Add `KeyStats`, KeyBindings.Stats |
| `internal/adapters/tui/statsview.go` | **NEW:** Render function |
| `internal/adapters/tui/statsview_test.go` | **NEW:** Unit tests |
| `internal/adapters/tui/components/status_bar.go:~249` | Update shortcut strings |

### Key Pattern References

**Width calculation** (NOT `m.effectiveWidth()` - that doesn't exist):
```go
effectiveWidth := m.width
if m.isWideWidth() {
    effectiveWidth = m.maxContentWidth
}
```

**Content height** (status bar is 2-3 lines):
```go
contentHeight := m.height - statusBarHeight(m.height) - headerHeight
```

**Exit bounds check** (projects may be removed while in stats view):
```go
if m.statsActiveProjectIdx >= 0 && m.statsActiveProjectIdx < len(m.projects) {
    m.projectList.SelectByIndex(m.statsActiveProjectIdx)
}
```

### Future Integration (Story 16.4+)

Stats View will connect to `MetricsRepository` (from Story 16.1) via:
- `GetTransitionsByProject(projectID string) []StageTransition`
- `GetTransitionsByTimeRange(from, to time.Time) []StageTransition`

### NFR Compliance

| Requirement | Target | This Story |
|-------------|--------|------------|
| NFR-P2-5 | Stats View render < 500ms | Placeholder = trivial |
| NFR-P2-7 | Metrics failure doesn't crash | No data access yet |

## Dev Agent Record

### Context Reference
- Story 16.3 spec followed exactly
- Pattern references from existing text view implementation (Story 12.1)
- Width calculation pattern from project-context.md

### Agent Model Used
Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References
N/A - No debug issues encountered

### Completion Notes List
1. All 10 tasks completed with red-green-refactor approach
2. Created 16 unit tests covering all acceptance criteria (corrected from initial 14)
3. All 1374 existing tests pass (no regressions)
4. Lint and fmt pass
5. Stats View follows existing TUI patterns (text view, hibernated view)
6. Key binding 's' blocks during overlays (help, note edit, remove confirm)
7. Selection restoration includes bounds check for safety

### Code Review Fixes Applied
- [M1] Fixed condensed status bar missing `[s]` stats shortcut hint (status_bar.go:190)
- [M3] Added test coverage for scroll stub keys j/k/arrows in Stats View (statsview_test.go)

### File List

- internal/adapters/tui/validation.go (modified) - Added viewModeStats enum
- internal/adapters/tui/model.go (modified) - Added state fields, enter/exit methods, key handler, routing
- internal/adapters/tui/keys.go (modified) - Added KeyStats constant and binding
- internal/adapters/tui/statsview.go (new) - Stats View render function
- internal/adapters/tui/statsview_test.go (new) - 16 unit tests (includes code review additions)
- internal/adapters/tui/export_test.go (modified) - Added test helper exports
- internal/adapters/tui/components/status_bar.go (modified) - Updated shortcut hints (full + condensed mode)

## User Testing Guide

**Time needed:** 2 minutes

### Step 1: Build and Run
```bash
make build
./bin/vdash
```

### Step 2: Verify Status Bar Hint
| Check | Expected |
|-------|----------|
| Status bar shows `[s] stats` | Between `[d]` and `[f]` hints |

### Step 3: Open Stats View
- Press `s` key

| Check | Expected |
|-------|----------|
| Full-screen view opens | Yes |
| Header shows "STATS" | Bold, purple text |
| Header shows "[ESC] Back to Dashboard" | Right-aligned hint |
| Content shows placeholder | "Project metrics will appear here" |

### Step 4: Return to Dashboard
- Press `Esc` or `q` key

| Check | Expected |
|-------|----------|
| Returns to Dashboard | Yes |
| Original project selection preserved | Same project highlighted |

### Step 5: Verify Blocking
- Press `?` to show help, then `s`

| Check | Expected |
|-------|----------|
| Stats View does NOT open | Help overlay blocks `s` key |

### Decision Guide
| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Any check fails | Do NOT approve, document issue |
