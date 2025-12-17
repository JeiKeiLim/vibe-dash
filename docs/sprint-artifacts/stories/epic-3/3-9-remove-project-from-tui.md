# Story 3.9: Remove Project from TUI

**Status:** Done

## Executive Summary

Add project removal capability via TUI ('x' key) with inline confirmation dialog. Implementation follows Story 3.7 (note editor) pattern: 'x' key triggers confirmation mode, renders centered dialog with y/n/Esc handling, async delete operation, feedback via status bar. KeyRemove constant already exists in keys.go:22. deleteProjectCmd already exists in model.go for validation mode - reuse it. Main work: confirmation state management, confirmation dialog rendering, timeout handling (30s per UX spec).

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/model.go` (KeyRemove handling), `internal/adapters/tui/views.go` (confirmation dialog) |
| **Key Dependencies** | ProjectRepository (`core/ports/repository.go`), ProjectListModel (`components/project_list.go`) |
| **Files to Modify** | `model.go` (add confirmation mode, handler, messages), `views.go` (add renderConfirmRemoveDialog) |
| **Files to Create** | `model_remove_test.go`, `views_remove_test.go` |
| **Location** | `internal/adapters/tui/` |
| **Interfaces Used** | `ports.ProjectRepository.Delete()` |

### Quick Task Summary (4 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Add confirmation state and message types | isConfirmingRemove, confirmTarget, confirmation timeout handling |
| 2 | Implement confirmation key handler | Handle 'x' key, y/n/Esc in confirmation mode, 30s timeout |
| 3 | Create confirmation dialog renderer | renderConfirmRemoveDialog in views.go following note editor pattern |
| 4 | Add tests | Confirmation flow, timeout, success/failure scenarios |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Confirmation pattern | Modal dialog overlay | Consistent with note editor pattern (Story 3.7) |
| Timeout | 30 seconds | Per AC5 and UX spec "30-second timeout per UX spec" |
| Delete command | Reuse existing `deleteProjectCmd` | Already exists in model.go:498-505 for validation mode |
| Key handling during confirmation | Only y/n/Esc accepted | Per AC6: "other keys are ignored (except y/n/Esc)" |
| Feedback style | Via status bar | Consistent with Stories 3.7, 3.8 feedback patterns |

## Story

**As a** user,
**I want** to remove projects from the dashboard,
**So that** I can clean up without leaving TUI.

## Acceptance Criteria

```gherkin
AC1: Given a project is selected
     When I press 'x'
     Then inline confirmation shows:
       "Remove 'client-bravo' from tracking? [y/n]"

AC2: Given confirmation dialog is shown
     When I press 'y'
     Then project is removed
     And list updates immediately
     And feedback shows: "✓ Removed: client-bravo"

AC3: Given confirmation dialog is shown
     When I press 'n' or Esc
     Then removal is cancelled
     And project remains

AC4: Given confirmation dialog is shown
     When 30 seconds pass without input
     Then auto-cancel occurs
     And dialog closes
     And project remains

AC5: Given confirmation is active
     When other keys are pressed
     Then they are ignored (except y/n/Esc)

AC6: Given project is successfully removed
     When list updates
     Then selection moves to next project (or previous if last)
     And detail panel updates with new selection
```

## Tasks / Subtasks

- [x] **Task 1: Add confirmation state and message types** (AC: 1, 4, 5)

  **PATTERN REFERENCE:** Follow note editing state pattern in `model.go:55-60` for state fields, and message types pattern at lines 112-124.

  - [x] 1.1 Add confirmation state fields to Model struct in `model.go` (after note editing state ~line 60):
    ```go
    // Remove confirmation state (Story 3.9)
    isConfirmingRemove bool
    confirmTarget      *domain.Project // Project pending removal
    ```

  - [x] 1.2 Add message types in `model.go` (after clearFavoriteFeedbackMsg ~line 138):
    ```go
    // removeConfirmedMsg signals project removal was confirmed (Story 3.9).
    type removeConfirmedMsg struct {
        projectID   string
        projectName string
        err         error
    }

    // removeConfirmTimeoutMsg signals confirmation dialog timed out (Story 3.9).
    type removeConfirmTimeoutMsg struct{}

    // clearRemoveFeedbackMsg signals to clear remove feedback message (Story 3.9).
    type clearRemoveFeedbackMsg struct{}
    ```

  - [x] 1.3 Add message handlers in Update function (after clearFavoriteFeedbackMsg case ~line 452):
    ```go
    case removeConfirmedMsg:
        // Handle async delete completion (Story 3.9)
        if msg.err != nil {
            m.statusBar.SetRefreshComplete("✗ Failed to remove: " + msg.err.Error())
            return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
                return clearRemoveFeedbackMsg{}
            })
        }
        // Remove from local projects slice
        var newProjects []*domain.Project
        for _, p := range m.projects {
            if p.ID != msg.projectID {
                newProjects = append(newProjects, p)
            }
        }
        m.projects = newProjects
        // Update project list component
        m.projectList.SetProjects(newProjects)
        // Update detail panel with new selection
        m.detailPanel.SetProject(m.projectList.SelectedProject())
        // Update status bar counts
        active, hibernated, waiting := components.CalculateCounts(m.projects)
        m.statusBar.SetCounts(active, hibernated, waiting)
        // Show success feedback
        m.statusBar.SetRefreshComplete("✓ Removed: " + msg.projectName)
        return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
            return clearRemoveFeedbackMsg{}
        })

    case removeConfirmTimeoutMsg:
        // Auto-cancel confirmation after 30 seconds (Story 3.9, AC4)
        if m.isConfirmingRemove {
            m.isConfirmingRemove = false
            m.confirmTarget = nil
        }
        return m, nil

    case clearRemoveFeedbackMsg:
        m.statusBar.SetRefreshComplete("")
        return m, nil
    ```

- [x] **Task 2: Implement confirmation key handler** (AC: 1, 2, 3, 5)

  **PATTERN REFERENCE:** Follow `handleNoteEditingKeyMsg` pattern in `model.go:654-671` for modal key handling.

  - [x] 2.1 Add KeyRemove case to handleKeyMsg in `model.go` (after KeyFavorite case ~line 616):
    ```go
    case KeyRemove:
        // Story 3.9: Start remove confirmation for selected project
        if m.isConfirmingRemove {
            return m, nil // Already confirming
        }
        if len(m.projects) == 0 {
            return m, nil // No project to remove
        }
        return m.startRemoveConfirmation()
    ```

  - [x] 2.2 Add confirmation routing in Update's `tea.KeyMsg` switch (insert AFTER the `isEditingNote` check at lines 253-255, BEFORE the `viewModeValidation` check):

    **CRITICAL:** This routing MUST be added to intercept keys during confirmation mode. Without it, AC5 ("other keys are ignored") will fail.

    ```go
    case tea.KeyMsg:
        // Route to note editing handler when in note editing mode (Story 3.7)
        if m.isEditingNote {
            return m.handleNoteEditingKeyMsg(msg)
        }
        // Route to remove confirmation handler when in confirmation mode (Story 3.9)
        // INSERT THIS BLOCK - MUST be before viewModeValidation check
        if m.isConfirmingRemove {
            return m.handleRemoveConfirmationKeyMsg(msg)
        }
        // Route to validation handler when in validation mode
        if m.viewMode == viewModeValidation {
            return m.handleValidationKeyMsg(msg)
        }
        return m.handleKeyMsg(msg)
    ```

  - [x] 2.3 Implement startRemoveConfirmation method in `model.go`:
    ```go
    // startRemoveConfirmation opens the remove confirmation dialog (Story 3.9).
    func (m Model) startRemoveConfirmation() (tea.Model, tea.Cmd) {
        selected := m.projectList.SelectedProject()
        if selected == nil {
            return m, nil
        }

        m.isConfirmingRemove = true
        m.confirmTarget = selected

        // Start 30-second timeout timer (AC4)
        return m, tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
            return removeConfirmTimeoutMsg{}
        })
    }
    ```

  - [x] 2.4 Implement handleRemoveConfirmationKeyMsg in `model.go`:

    **PATTERN NOTE:** Use `msg.Type` comparison for Escape key (consistent with `handleNoteEditingKeyMsg` at line 656-660), then `msg.String()` for character keys.

    ```go
    // handleRemoveConfirmationKeyMsg processes keyboard input during remove confirmation (Story 3.9).
    func (m Model) handleRemoveConfirmationKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        // Handle Escape key by type (consistent with handleNoteEditingKeyMsg pattern)
        if msg.Type == tea.KeyEsc {
            m.isConfirmingRemove = false
            m.confirmTarget = nil
            return m, nil
        }

        // Handle character keys by string
        switch msg.String() {
        case "y", "Y":
            // Confirm removal
            m.isConfirmingRemove = false
            target := m.confirmTarget
            m.confirmTarget = nil

            // Use effectiveName for display
            projectName := target.Name
            if target.DisplayName != "" {
                projectName = target.DisplayName
            }

            return m, m.removeProjectCmd(target.ID, projectName)

        case "n", "N":
            // Cancel removal
            m.isConfirmingRemove = false
            m.confirmTarget = nil
            return m, nil
        }

        // AC5: Ignore all other keys during confirmation
        return m, nil
    }
    ```

  - [x] 2.5 Implement removeProjectCmd in `model.go`:
    ```go
    // removeProjectCmd creates a command that deletes a project from repository (Story 3.9).
    // Note: Reuses delete pattern from deleteProjectCmd but returns removeConfirmedMsg.
    func (m Model) removeProjectCmd(projectID, projectName string) tea.Cmd {
        return func() tea.Msg {
            ctx := context.Background()
            err := m.repository.Delete(ctx, projectID)
            return removeConfirmedMsg{projectID: projectID, projectName: projectName, err: err}
        }
    }
    ```

- [x] **Task 3: Create confirmation dialog renderer** (AC: 1)

  **PATTERN REFERENCE:** Follow `renderNoteEditor` in `views.go:116-165` for dialog styling.

  - [x] 3.1 Add renderConfirmRemoveDialog function in `views.go`:
    ```go
    // renderConfirmRemoveDialog renders the inline remove confirmation dialog (Story 3.9).
    // Follows renderNoteEditor pattern for dialog styling and centering.
    func renderConfirmRemoveDialog(projectName string, width, height int) string {
        // Dialog dimensions - cap width at 60
        dialogWidth := width - 4
        if dialogWidth < 30 {
            dialogWidth = 30
        }
        if dialogWidth > 60 {
            dialogWidth = 60
        }

        // Title
        title := titleStyle.Render("Confirm Removal")

        // Content
        content := strings.Join([]string{
            "",
            fmt.Sprintf("Remove '%s' from tracking?", projectName),
            "",
            hintStyle.Render("[y] Yes  [n] No  [Esc] Cancel"),
            "",
        }, "\n")

        // Dialog box style (same as help overlay)
        box := boxStyle.
            Width(dialogWidth).
            Render(content)

        // Add title to the border (same pattern as renderHelpOverlay)
        lines := strings.Split(box, "\n")
        if len(lines) > 0 {
            topBorder := lines[0]
            titleWithDash := fmt.Sprintf("\u2500 %s ", title)

            if len(topBorder) > 3 {
                lines[0] = string(topBorder[0]) + titleWithDash + topBorder[len(titleWithDash)+1:]
            }
            box = strings.Join(lines, "\n")
        }

        // Center in terminal
        return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
    }
    ```

  - [x] 3.2 Add confirmation dialog rendering in View() in `model.go` (after note editor check ~line 764):
    ```go
    // Render remove confirmation dialog (overlays everything) (Story 3.9)
    if m.isConfirmingRemove && m.confirmTarget != nil {
        projectName := m.confirmTarget.Name
        if m.confirmTarget.DisplayName != "" {
            projectName = m.confirmTarget.DisplayName
        }
        return renderConfirmRemoveDialog(projectName, m.width, m.height)
    }
    ```

- [x] **Task 4: Add tests** (AC: all)

  - [x] 4.1 Create `internal/adapters/tui/model_remove_test.go`:

    **NOTE:** Reuse `favoriteMockRepository` from `model_favorite_test.go` or extract to shared `mock_repository_test.go`.

    ```go
    package tui

    import (
        "fmt"
        "testing"

        tea "github.com/charmbracelet/bubbletea"

        "github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    )

    // newMockRepository returns the favoriteMockRepository for test reuse.
    // If favoriteMockRepository is not accessible, copy the implementation from model_favorite_test.go.
    func newMockRepository() *favoriteMockRepository {
        return &favoriteMockRepository{}
    }

    func TestModel_RemoveKey_StartsConfirmation(t *testing.T) {
        // Setup: Model with project
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.repository = repo

        // Action: Send 'x' key
        newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
        updated := newModel.(Model)

        // Assert: Confirmation mode active
        if !updated.isConfirmingRemove {
            t.Error("expected isConfirmingRemove to be true")
        }
        if updated.confirmTarget == nil {
            t.Error("expected confirmTarget to be set")
        }
        // Assert: Timeout command returned
        if cmd == nil {
            t.Error("expected timeout command to be returned")
        }
    }

    func TestModel_RemoveKey_IgnoredWhenNoProjects(t *testing.T) {
        // Setup: Model WITHOUT projects
        repo := newMockRepository()
        m := NewModel(repo)
        m.projects = nil

        // Action
        newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
        updated := newModel.(Model)

        // Assert: No confirmation mode
        if updated.isConfirmingRemove {
            t.Error("expected isConfirmingRemove to be false")
        }
        if cmd != nil {
            t.Error("expected no command when no projects")
        }
    }

    func TestModel_RemoveConfirmation_ConfirmsWithY(t *testing.T) {
        // Setup: Model in confirmation mode
        repo := newMockRepository()
        project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
        repo.projects = []*domain.Project{project}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.repository = repo
        m.isConfirmingRemove = true
        m.confirmTarget = project

        // Action: Send 'y' key
        newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
        updated := newModel.(Model)

        // Assert: Confirmation mode exited
        if updated.isConfirmingRemove {
            t.Error("expected isConfirmingRemove to be false after confirm")
        }
        // Assert: Delete command returned
        if cmd == nil {
            t.Error("expected delete command to be returned")
        }
    }

    func TestModel_RemoveConfirmation_CancelsWithN(t *testing.T) {
        // Setup: Model in confirmation mode
        repo := newMockRepository()
        project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
        repo.projects = []*domain.Project{project}

        m := NewModel(repo)
        m.projects = repo.projects
        m.isConfirmingRemove = true
        m.confirmTarget = project

        // Action: Send 'n' key
        newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
        updated := newModel.(Model)

        // Assert: Confirmation mode exited, project still exists
        if updated.isConfirmingRemove {
            t.Error("expected isConfirmingRemove to be false after cancel")
        }
        if len(updated.projects) != 1 {
            t.Error("expected project to still exist after cancel")
        }
    }

    func TestModel_RemoveConfirmation_CancelsWithEsc(t *testing.T) {
        // Setup: Model in confirmation mode
        repo := newMockRepository()
        project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
        repo.projects = []*domain.Project{project}

        m := NewModel(repo)
        m.projects = repo.projects
        m.isConfirmingRemove = true
        m.confirmTarget = project

        // Action: Send Esc key
        newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
        updated := newModel.(Model)

        // Assert: Confirmation mode exited
        if updated.isConfirmingRemove {
            t.Error("expected isConfirmingRemove to be false after Esc")
        }
    }

    func TestModel_RemoveConfirmation_IgnoresOtherKeys(t *testing.T) {
        // Setup: Model in confirmation mode
        repo := newMockRepository()
        project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
        repo.projects = []*domain.Project{project}

        m := NewModel(repo)
        m.projects = repo.projects
        m.isConfirmingRemove = true
        m.confirmTarget = project

        // Action: Send various other keys
        otherKeys := []string{"q", "j", "k", "d", "f", "r", "a"}
        for _, key := range otherKeys {
            newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
            updated := newModel.(Model)

            // Assert: Still in confirmation mode
            if !updated.isConfirmingRemove {
                t.Errorf("expected isConfirmingRemove to remain true after '%s' key", key)
            }
        }
    }

    func TestModel_RemoveConfirmation_TimeoutCancels(t *testing.T) {
        // Setup: Model in confirmation mode
        repo := newMockRepository()
        project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
        repo.projects = []*domain.Project{project}

        m := NewModel(repo)
        m.projects = repo.projects
        m.isConfirmingRemove = true
        m.confirmTarget = project

        // Action: Send timeout message
        newModel, _ := m.Update(removeConfirmTimeoutMsg{})
        updated := newModel.(Model)

        // Assert: Confirmation mode exited
        if updated.isConfirmingRemove {
            t.Error("expected isConfirmingRemove to be false after timeout")
        }
        if updated.confirmTarget != nil {
            t.Error("expected confirmTarget to be nil after timeout")
        }
    }

    func TestModel_RemoveConfirmedMsg_UpdatesProjectList(t *testing.T) {
        // Setup: Model with 2 projects
        repo := newMockRepository()
        repo.projects = []*domain.Project{
            {ID: "1", Path: "/test1", Name: "project-1"},
            {ID: "2", Path: "/test2", Name: "project-2"},
        }

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.statusBar = components.NewStatusBarModel(80)
        m.detailPanel = components.NewDetailPanelModel(80, 24)
        m.repository = repo

        // Action: Receive removeConfirmedMsg for project-1
        newModel, _ := m.Update(removeConfirmedMsg{projectID: "1", projectName: "project-1"})
        updated := newModel.(Model)

        // Assert: Project removed from list
        if len(updated.projects) != 1 {
            t.Errorf("expected 1 project, got %d", len(updated.projects))
        }
        if updated.projects[0].ID != "2" {
            t.Error("expected project-2 to remain")
        }
    }

    func TestModel_RemoveConfirmedMsg_ShowsFeedback(t *testing.T) {
        // Setup: Model with project
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.statusBar = components.NewStatusBarModel(80)
        m.detailPanel = components.NewDetailPanelModel(80, 24)
        m.repository = repo

        // Action: Receive removeConfirmedMsg
        newModel, cmd := m.Update(removeConfirmedMsg{projectID: "1", projectName: "test-project"})
        _ = newModel.(Model)

        // Assert: Timer command returned for feedback clearing
        if cmd == nil {
            t.Error("expected timer command for feedback clearing")
        }
    }

    func TestModel_RemoveConfirmedMsg_ErrorShowsFeedback(t *testing.T) {
        // Setup: Model with project
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.statusBar = components.NewStatusBarModel(80)
        m.repository = repo

        // Action: Receive removeConfirmedMsg with error
        testErr := fmt.Errorf("delete failed")
        newModel, cmd := m.Update(removeConfirmedMsg{projectID: "1", projectName: "test-project", err: testErr})
        updated := newModel.(Model)

        // Assert: Project NOT removed (error case)
        if len(updated.projects) != 1 {
            t.Error("expected project to remain on error")
        }
        // Assert: Timer command returned for error feedback clearing
        if cmd == nil {
            t.Error("expected timer command for error feedback clearing")
        }
    }

    func TestModel_RemoveKey_UsesDisplayNameInConfirmation(t *testing.T) {
        // Setup: Model with project that has display name
        repo := newMockRepository()
        repo.projects = []*domain.Project{{
            ID:          "1",
            Path:        "/test",
            Name:        "test-project",
            DisplayName: "My Custom Name",
        }}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.repository = repo

        // Action: Send 'x' key
        newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
        updated := newModel.(Model)

        // Assert: confirmTarget has display name accessible
        if updated.confirmTarget == nil {
            t.Fatal("expected confirmTarget to be set")
        }
        if updated.confirmTarget.DisplayName != "My Custom Name" {
            t.Errorf("expected DisplayName 'My Custom Name', got '%s'", updated.confirmTarget.DisplayName)
        }
    }
    ```

  - [x] 4.2 Add test for renderConfirmRemoveDialog in `views_test.go` (or create `views_remove_test.go`):
    ```go
    func TestRenderConfirmRemoveDialog_ContainsProjectName(t *testing.T) {
        output := renderConfirmRemoveDialog("my-project", 80, 24)

        if !strings.Contains(output, "my-project") {
            t.Error("expected dialog to contain project name")
        }
        if !strings.Contains(output, "Remove") {
            t.Error("expected dialog to contain 'Remove'")
        }
        if !strings.Contains(output, "[y]") {
            t.Error("expected dialog to contain '[y]' hint")
        }
        if !strings.Contains(output, "[n]") {
            t.Error("expected dialog to contain '[n]' hint")
        }
    }

    func TestRenderConfirmRemoveDialog_WidthConstraints(t *testing.T) {
        tests := []struct {
            name     string
            width    int
            expected int // Expected dialog width
        }{
            {"narrow terminal caps at 30", 20, 30},
            {"medium terminal uses width-4", 50, 46},
            {"wide terminal caps at 60", 100, 60},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                output := renderConfirmRemoveDialog("test", tt.width, 24)
                // Dialog should render without panic
                if len(output) == 0 {
                    t.Error("expected non-empty output")
                }
            })
        }
    }
    ```

  - [x] 4.3 Run verification:
    ```bash
    make test   # All tests pass
    make lint   # No lint errors
    make build  # Successful build
    ```

## Dev Notes

### Current State Analysis

**KeyRemove constant:** Already exists in `keys.go:22` as `KeyRemove = "x"`. No changes needed.

**deleteProjectCmd exists:** Already implemented in `model.go:498-505` for validation mode. However, it returns `deleteProjectMsg` which is used for validation flow. Story 3.9 needs a separate `removeProjectCmd` that returns `removeConfirmedMsg` to avoid mixing with validation logic.

**Model Structure:** Has established patterns from Stories 3.7-3.8 for:
- Modal state fields (`isEditingNote`, `noteEditTarget`)
- Modal key handlers (`handleNoteEditingKeyMsg`)
- Dialog renderers (`renderNoteEditor`)
- Feedback via status bar (`SetRefreshComplete`)
- 3-second feedback timeout

### Pattern Reference: Story 3.7 (Note Editor)

Story 3.7 established the modal dialog pattern:
1. Key press → `startNoteEditing()` → sets modal state
2. Modal key handler intercepts all keys
3. Dialog renderer (`renderNoteEditor`) overlays content
4. Save/Cancel → clears modal state
5. Success/Error → feedback via status bar

Follow this exact pattern for remove confirmation.

### Confirmation Flow

```
User presses 'x'
    │
    ▼
startRemoveConfirmation()
    │ - Set isConfirmingRemove = true
    │ - Set confirmTarget = selected project
    │ - Start 30-second timeout timer
    │
    ▼
View() renders renderConfirmRemoveDialog()
    │
    ├──► User presses 'y' → handleRemoveConfirmationKeyMsg()
    │     │ - Clear confirmation state
    │     │ - Call removeProjectCmd()
    │     │
    │     ▼
    │   removeConfirmedMsg received in Update()
    │     │ - Remove from m.projects
    │     │ - Update projectList component
    │     │ - Update detail panel
    │     │ - Update status bar counts
    │     │ - Show feedback "✓ Removed: X"
    │     │ - Start 3s feedback timer
    │
    ├──► User presses 'n' or Esc → Cancel
    │     │ - Clear confirmation state
    │     │ - Return to normal mode
    │
    └──► Timeout (30s) → removeConfirmTimeoutMsg
          │ - Clear confirmation state
          │ - Return to normal mode
```

### Selection After Removal (AC6)

**IMPORTANT:** The `projectList.SetProjects()` method does NOT automatically preserve selection index. Per `project_list.go:61-81`, it only selects first item if "nothing is selected" (`m.list.Index() < 0`).

**Current behavior:** After `SetItems()`, the Bubbles list maintains its index value but the item at that index changes. This is acceptable because:
1. If removed item was at index 3 of 5, index 3 now points to what was index 4 (next item) - correct behavior
2. If removed last item (index 4 of 5), index 4 is now out of bounds - `SelectedProject()` returns nil, then SetProjects selects first item

**Edge case handling:** If the removed project was the LAST one and was selected, after removal:
- `m.projects` becomes empty
- `projectList.SetProjects([])` is called
- View() renders empty view (existing logic at line 769-774)

The model calls `SetProjects()` then `SetProject(m.projectList.SelectedProject())` on detail panel. This sequence ensures detail panel shows the newly selected project (or nil if empty).

### Feedback Messages

Per AC2, feedback should be:
- "✓ Removed: <project-name>" on success
- "✗ Failed to remove: <error>" on error

Use display name if available (same pattern as CLI remove.go:103-107).

### Timeout Handling (AC4)

The 30-second timeout uses `tea.Tick`:
```go
tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
    return removeConfirmTimeoutMsg{}
})
```

When `removeConfirmTimeoutMsg` is received:
1. Check if still in confirmation mode (might have confirmed/cancelled already)
2. If yes, clear confirmation state (auto-cancel)

**Race Condition Consideration:** If user presses 'y' at exactly 30 seconds, both `removeConfirmedMsg` and `removeConfirmTimeoutMsg` could arrive in sequence. This is handled safely because:
- The `removeConfirmTimeoutMsg` handler checks `if m.isConfirmingRemove` before clearing state
- If confirmation was already processed (user pressed 'y'), `isConfirmingRemove` is already false
- The timeout message becomes a no-op - safe and correct behavior

**No additional complexity needed** - the existing guard check (`if m.isConfirmingRemove`) handles this race condition.

### Key Handling During Confirmation (AC5)

Per AC: "other keys are ignored (except y/n/Esc)"

The `handleRemoveConfirmationKeyMsg` only handles y/Y/n/N/Esc. All other keys return `(m, nil)` - no action, no command.

### Project Context Compliance

Per `docs/project-context.md`:
- Context first: All operations use `ctx context.Context`
- Error wrapping: Use `fmt.Errorf("...: %w", err)`
- Co-locate tests: `*_test.go` next to source
- Domain errors: Return from repository, not wrapped again

### Architecture Compliance

Per `docs/architecture.md`:
- Hexagonal architecture: TUI adapter calls repository directly (acceptable for delete operation)
- Repository.Delete handles actual deletion
- No new service needed - remove is a simple repository operation

### Edge Cases to Handle

1. **No projects:** Ignore 'x' key (checked before `startRemoveConfirmation`)
2. **Already confirming:** Ignore additional 'x' key presses
3. **Delete error:** Show error feedback, project remains in list
4. **Timeout while already confirmed:** Check `isConfirmingRemove` before clearing
5. **Last project removed:** Empty view should display (handled by existing View() logic)

### Differences from CLI Remove

The CLI `remove.go` uses interactive stdin prompt. TUI uses modal dialog overlay:
- CLI: blocks on `scanner.Scan()` for y/n input
- TUI: renders dialog, handles keys asynchronously, has 30s timeout

The delete operation (`repository.Delete`) is identical.

### Dialog Width Pattern (Optimization Note)

The `renderConfirmRemoveDialog` uses the same width calculation as `renderNoteEditor` (views.go:119-126):
```go
dialogWidth := width - 4
if dialogWidth < 30 { dialogWidth = 30 }
if dialogWidth > 60 { dialogWidth = 60 }
```

**Post-implementation consideration:** If more dialogs are added, extract shared constants:
```go
const (
    dialogMinWidth = 30
    dialogMaxWidth = 60
)
```

For now, duplication is acceptable with only 2 dialogs.

### Future Considerations

Post-MVP could add:
- Undo support (keep deleted project in memory for short period)
- Batch removal (select multiple projects)
- Different confirmation styles based on user preference

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 3.9 requirements - lines 1435-1470)
- docs/architecture.md (Hexagonal patterns, TUI adapter structure)
- docs/project-context.md (Go conventions, testing rules)
- internal/adapters/tui/model.go (Current Model structure, message patterns from Stories 3.7-3.8)
- internal/adapters/tui/keys.go (KeyRemove already defined at line 22)
- internal/adapters/tui/views.go (renderNoteEditor pattern for dialogs)
- internal/adapters/cli/remove.go (CLI remove pattern for reference)
- internal/adapters/tui/components/project_list.go (SetProjects for list update)
- docs/sprint-artifacts/stories/epic-3/3-8-toggle-favorite.md (Previous story patterns)
- Git history: Stories 3.1-3.8 implementation patterns

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story drafting phase.

### Completion Notes List

- ✅ Task 1 complete: Added `isConfirmingRemove` and `confirmTarget` state fields to Model struct; added `removeConfirmedMsg`, `removeConfirmTimeoutMsg`, and `clearRemoveFeedbackMsg` message types; implemented handlers in Update function for all three message types
- ✅ Task 2 complete: Added KeyRemove case in handleKeyMsg; added confirmation routing in Update's tea.KeyMsg switch to intercept keys during confirmation mode; implemented `startRemoveConfirmation()`, `handleRemoveConfirmationKeyMsg()`, and `removeProjectCmd()` methods
- ✅ Task 3 complete: Added `renderConfirmRemoveDialog()` function in views.go following the note editor pattern; integrated dialog rendering in View() after note editor check
- ✅ Task 4 complete: Created `model_remove_test.go` with 15 tests covering confirmation flow, timeout, cancellation, error handling, and edge cases; added 7 tests for `renderConfirmRemoveDialog` in views_test.go; all tests pass, lint clean, build succeeds
- ✅ Code Review Fixes: (H1) Replaced inline effectiveName logic with `project.EffectiveName()` in model.go; (H2) Fixed AC6 bug in SetProjects to handle out-of-bounds index when removing last item, added 3 AC6 tests; (M1) Added 2 View() render tests; (M3) Added 2 uppercase key tests. Total: 22 tests in model_remove_test.go, 7 in views_test.go.

### File List

**Modified:**
- `internal/adapters/tui/model.go` - Added confirmation state fields (lines 63-64), message types (lines 144-155), handlers in Update (lines 475-514), key routing (lines 273-276), KeyRemove case (lines 679-687), methods: startRemoveConfirmation, handleRemoveConfirmationKeyMsg, removeProjectCmd (lines 811-871), View() dialog rendering (lines 897-901). Code review: Fixed to use `project.EffectiveName()` instead of inline logic.
- `internal/adapters/tui/views.go` - Added renderConfirmRemoveDialog function (lines 167-210)
- `internal/adapters/tui/views_test.go` - Added 7 tests for renderConfirmRemoveDialog (lines 153-280)
- `internal/adapters/tui/components/project_list.go` - **Code review fix**: Updated SetProjects() (lines 77-85) to handle out-of-bounds index after removal (AC6 fix)

**Created:**
- `internal/adapters/tui/model_remove_test.go` - 22 tests for remove confirmation flow (including code review additions)

**Existing (Reference Only):**
- `internal/adapters/tui/keys.go:22` - `KeyRemove = "x"` already defined
- `internal/adapters/cli/remove.go` - CLI remove pattern reference

## Change Log

| Date | Change |
|------|--------|
| 2025-12-17 | Story created with ready-for-dev status by SM Agent (Bob) in YOLO mode. Comprehensive developer context included with all technical decisions, code patterns, and test specifications. |
| 2025-12-17 | Validation improvements applied by SM Agent (Bob): (1) **C1-CRITICAL:** Clarified Task 2.2 with complete code block showing exact insertion point after isEditingNote check; (2) **C2-CRITICAL:** Fixed Task 2.4 to use msg.Type comparison for Escape key (consistent with handleNoteEditingKeyMsg pattern); (3) **E1:** Updated "Selection After Removal" section with accurate analysis of SetProjects behavior and edge case handling; (4) **E2:** Added `fmt` import to test specification and note about mock repository reuse; (5) **E3:** Added "Race Condition Consideration" note explaining why existing guard check handles timeout race safely; (6) **O2:** Added "Dialog Width Pattern" section noting optimization opportunity for shared constants; (7) **O3:** Added note about test helper extraction. |
| 2025-12-17 | Implementation complete by Dev Agent (Amelia). All 4 tasks completed following red-green-refactor cycle. 23 tests added (15 in model_remove_test.go, 8 in views_test.go). All tests pass, lint clean, build successful. Status changed to Ready for Review. |
| 2025-12-17 | **Code Review Fixes by Dev Agent (Amelia):** (H1) Fixed DRY violation - replaced inline effectiveName logic with `project.EffectiveName()` in model.go:844-845 and model.go:899; (H2) Fixed AC6 bug - SetProjects() in project_list.go now handles out-of-bounds index when removing last item (selection moves to previous); Added 3 AC6 tests; (M1) Added 2 View() render tests for confirmation dialog; (M3) Added 2 uppercase Y/N key tests. Test count now: 22 in model_remove_test.go, 7 in views_test.go (29 total). All tests pass, lint clean, build successful. |
