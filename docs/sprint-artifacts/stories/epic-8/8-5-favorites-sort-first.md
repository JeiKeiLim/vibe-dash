# Story 8.5: Favorites Sort First

Status: done

## Story

As a **user with favorite projects**,
I want **favorites to appear at the top of the list**,
So that **my most important projects are immediately visible**.

## Acceptance Criteria

1. **AC1: Favorites Appear Before Non-Favorites**
   - Given projects with some marked as favorites
   - When viewing the project list
   - Then favorites appear before non-favorites

2. **AC2: Favorites Sorted Alphabetically Among Themselves**
   - Given multiple favorites
   - When viewing the list
   - Then favorites are sorted alphabetically among themselves (case-insensitive)

3. **AC3: Non-Favorites Sorted Alphabetically After Favorites**
   - Given non-favorites
   - When viewing the list
   - Then non-favorites are sorted alphabetically after all favorites (case-insensitive)

4. **AC4: Toggling Favorite Updates Sort Position**
   - Given a non-favorite project
   - When I toggle it to favorite (press 'f')
   - Then it moves to the favorites section at top (sorted alphabetically within favorites)

5. **AC5: Selection Preserved After Toggle**
   - Given a project is selected
   - When I toggle its favorite status
   - Then the same project remains selected (at its new position)

## Tasks / Subtasks

- [x] Task 1: Update SortByName to sort favorites first (AC: 1, 2, 3)
  - [x] 1.1: Modify `internal/shared/project/project.go` `SortByName()` function
  - [x] 1.2: Primary sort: favorites (`IsFavorite=true`) before non-favorites
  - [x] 1.3: Secondary sort: alphabetical by EffectiveName (case-insensitive)
  - [x] 1.4: Add unit tests for sorting behavior (include nil/empty slice edge cases)

- [x] Task 2: Add list re-sort after favorite toggle (AC: 4) **CRITICAL**
  - [x] 2.1: In `favoriteSavedMsg` handler (model.go:720), add `m.projectList.SetProjects(m.projects)` after local state update
  - [x] 2.2: `SetProjects()` internally calls `SortByName()` - verify this path works

- [x] Task 3: Preserve selection after favorite toggle (AC: 5) **CRITICAL**
  - [x] 3.1: In `favoriteSavedMsg` handler, after `SetProjects()`, find project index by ID
  - [x] 3.2: Call `m.projectList.SelectByIndex(newIndex)` to restore selection
  - [x] 3.3: Pattern: Loop through `m.projectList.Projects()` to find matching ID, get index
  - [x] 3.4: Add test verifying selection preserved after toggle

- [x] Task 4: Add comprehensive tests (AC: all)
  - [x] 4.1: Test: Favorites sort before non-favorites
  - [x] 4.2: Test: Alphabetical sort within favorites group
  - [x] 4.3: Test: Alphabetical sort within non-favorites group
  - [x] 4.4: Test: Toggle favorite moves project to correct position
  - [x] 4.5: Test: Selection preserved after toggle (critical path)
  - [x] 4.6: Test: Edge cases (nil slice, empty slice, all favorites, no favorites)

## Dev Notes

### Critical Implementation Gap (MUST READ)

**The `favoriteSavedMsg` handler does NOT re-sort or preserve selection.** Current code at model.go:720-741:

```go
case favoriteSavedMsg:
    // Update local project state (Story 3.8)
    for _, p := range m.projects {
        if p.ID == msg.projectID {
            p.IsFavorite = msg.isFavorite
            break
        }
    }
    // Update detail panel
    m.detailPanel.SetProject(m.projectList.SelectedProject())
    // ...feedback...
```

**MISSING:** No `SetProjects()` call, no selection preservation. Must add:

```go
case favoriteSavedMsg:
    // Capture selected project ID BEFORE re-sort
    selectedID := ""
    if selected := m.projectList.SelectedProject(); selected != nil {
        selectedID = selected.ID
    }

    // Update local project state
    for _, p := range m.projects {
        if p.ID == msg.projectID {
            p.IsFavorite = msg.isFavorite
            break
        }
    }

    // Re-sort list (triggers SortByName via SetProjects)
    m.projectList.SetProjects(m.projects)

    // Restore selection by ID (project may have moved position)
    if selectedID != "" {
        for i, p := range m.projects {
            if p.ID == selectedID {
                m.projectList.list.Select(i)
                break
            }
        }
    }

    // Update detail panel with (possibly moved) selection
    m.detailPanel.SetProject(m.projectList.SelectedProject())
    // ...rest unchanged...
```

**Note:** `m.projectList.list` is private - need to add `SelectByIndex(idx int)` method to ProjectListModel.

### SortByName Fix

Current (project.go:21): Only sorts by name. Fix:

```go
func SortByName(projects []*domain.Project) {
    sort.Slice(projects, func(i, j int) bool {
        if projects[i].IsFavorite != projects[j].IsFavorite {
            return projects[i].IsFavorite // true before false
        }
        return strings.ToLower(EffectiveName(projects[i])) < strings.ToLower(EffectiveName(projects[j]))
    })
}
```

### Key Code Locations

| File | Line | Purpose |
|------|------|---------|
| `internal/shared/project/project.go` | 20-27 | `SortByName()` - update sort logic |
| `internal/adapters/tui/components/project_list.go` | 21-26 | `NewProjectListModel()` calls SortByName |
| `internal/adapters/tui/components/project_list.go` | 62-86 | `SetProjects()` calls SortByName, handles selection |
| `internal/adapters/tui/model.go` | 720-741 | `favoriteSavedMsg` handler - **MUST MODIFY** |
| `internal/adapters/tui/model.go` | 1150-1159 | `toggleFavorite()` - triggers async save |

### Required New Method

Add to `project_list.go`:

```go
// SelectByIndex selects the item at the given index.
// Used for selection preservation after list re-sort.
func (m *ProjectListModel) SelectByIndex(idx int) {
    if idx >= 0 && idx < len(m.projects) {
        m.list.Select(idx)
    }
}
```

### Architecture Compliance

- **Modify:** `internal/shared/project/project.go` - update sort logic
- **Modify:** `internal/adapters/tui/model.go` - favoriteSavedMsg handler (add re-sort + selection)
- **Add:** `internal/adapters/tui/components/project_list.go` - SelectByIndex method
- **Tests:** `internal/shared/project/project_test.go` (create if needed)

### Previous Story Learnings

**From Story 8.4:**
- Race conditions between async messages require careful state management
- Pattern: Capture state → update → restore (use for selection preservation)

**From Story 3.9 (Remove project - project_list.go:77-85):**
- SetProjects already handles selection bounds - follow same pattern for re-selection by ID

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Sort only by name | Sort by favorite first, then name |
| Update favorite in-place without re-sort | Call `SetProjects()` after state update |
| Assume list position preserved after toggle | Capture ID → SetProjects → find by ID → Select |
| Ignore case in sort | Use `strings.ToLower()` for comparison |
| Access private `m.list.Select()` from model.go | Add `SelectByIndex()` public method |

### References

| Document | Relevance |
|----------|-----------|
| docs/project-context.md | Story Completion - User verification required |
| internal/shared/project/project.go:20-27 | Current SortByName implementation |
| internal/adapters/tui/model.go:720-741 | favoriteSavedMsg handler to modify |
| internal/adapters/tui/components/project_list.go:62-86 | SetProjects with selection handling |

## User Testing Guide

**Time needed:** 3-5 minutes

### Step 1: Setup Test Data

```bash
# Build and run
make build && ./bin/vibe

# If you don't have multiple projects, add test projects:
# (or use existing projects)
```

### Step 2: Verify Favorites Sort First

| Check | Expected | Red Flag |
|-------|----------|----------|
| Favorite projects | Appear at TOP of list | Mixed with non-favorites |
| Non-favorite projects | Appear AFTER all favorites | Some above favorites |
| Alphabetical order | Within each group | Random order within groups |

### Step 3: Toggle and Verify

```bash
# Select a non-favorite project
# Press 'f' to toggle favorite
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Project moves | To top section (favorites) | Stays in place |
| Alphabetical position | Correct within favorites | At top/bottom of favorites |
| Selection | Same project selected | Selection jumps to different project |

### Step 4: Unfavorite and Verify

```bash
# With favorite selected, press 'f' to unfavorite
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Project moves | To bottom section (non-favorites) | Stays at top |
| Alphabetical position | Correct within non-favorites | Wrong position |
| Selection | Same project selected | Selection lost |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Favorites not at top | Check SortByName implementation |
| Wrong alphabetical order | Check case-insensitive comparison |
| Selection lost after toggle | Check selection preservation in toggleFavorite |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- **Task 1**: Updated `SortByName()` in `project.go` to use two-phase sorting: favorites first (primary), then alphabetical (secondary, case-insensitive). Added 7 new unit tests covering all edge cases.
- **Task 2**: Modified `favoriteSavedMsg` handler in `model.go` to call `SetProjects()` after updating favorite status, triggering automatic re-sort via `SortByName()`.
- **Task 3**: Added `SelectByIndex()` and `Projects()` methods to `ProjectListModel`. Implemented selection preservation pattern: capture ID → re-sort → find by ID → restore selection.
- **Task 4**: Added comprehensive test suite in `project_list_test.go` covering favorites sorting, `SelectByIndex` bounds checking, and re-sort behavior.

### Code Review Fixes (2025-12-26)

- **H1**: Added `TestProjectListModel_Story85_SelectionPreservedAfterFavoriteToggle` - comprehensive test for AC5 selection preservation after favorite toggle
- **M4**: Added edge case handling in `model.go:738-752` - if project not found after re-sort, selects first item as fallback
- **L1**: Removed redundant comment from `Projects()` method in `project_list.go`

### File List

- `internal/shared/project/project.go` (modified) - SortByName with favorites-first sorting
- `internal/shared/project/project_test.go` (modified) - 7 new tests for favorites sorting
- `internal/adapters/tui/components/project_list.go` (modified) - Added SelectByIndex() and Projects() methods
- `internal/adapters/tui/components/project_list_test.go` (modified) - 9 new tests for Story 8.5 (8 initial + 1 code review)
- `internal/adapters/tui/model.go` (modified) - favoriteSavedMsg handler with re-sort and selection preservation + edge case handling

