# Story 3.8: Toggle Favorite

**Status:** Done

## Executive Summary

Add favorite toggling capability to projects via TUI ('f' key) and CLI (`vibe favorite`). Implementation reuses existing patterns: KeyFavorite constant exists in keys.go, IsFavorite field exists in domain.Project (line 21), delegate.go already has column constants. Main work is adding 'f' key handler (follow Story 3.7 pattern), showing ⭐ indicator in delegate.go render, and CLI command using existing DI pattern from add.go. No database migration needed - IsFavorite column already exists via domain mapping in SQLite.

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/model.go` (KeyFavorite handling), `internal/adapters/cli/favorite.go` (new file) |
| **Key Dependencies** | ProjectRepository (`core/ports/repository.go`), ProjectItemDelegate (`components/delegate.go`) |
| **Files to Modify** | `model.go` (add favorite toggle handler), `components/delegate.go` (add ⭐ indicator in row rendering), `components/detail_panel.go` (add IsFavorite display) |
| **Files to Create** | `cli/favorite.go`, `cli/favorite_test.go`, `model_favorite_test.go` |
| **Location** | `internal/adapters/tui/`, `internal/adapters/cli/` |
| **Interfaces Used** | `ports.ProjectRepository.Save()` |

### Quick Task Summary (4 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Implement TUI favorite toggle handler | Handle 'f' key, toggle IsFavorite field, save to repo, show feedback |
| 2 | Add favorite indicator to project row | Display ⭐ prefix for favorited projects in delegate.go |
| 3 | Create CLI favorite command | `vibe favorite <project> [--off]` for non-interactive favorite management |
| 4 | Add tests | TUI favorite toggle behavior + CLI favorite command tests |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Indicator position | Before project name | Visually prominent, matches UX spec "⭐ prefix" |
| Feedback style | Via status bar | Consistent with Story 3.7 note feedback |
| CLI toggle behavior | Toggle by default, --off to remove | Matches PRD FR54 "toggle", --off for explicit removal |
| Confirmation dialog | None needed | Per AC: "No confirmation needed (easily reversible)" |
| Favorite style | Magenta foreground | Per Architecture styles.go FavoriteStyle (color 5) |

## Story

**As a** user,
**I want** to mark projects as favorites,
**So that** they stay visible regardless of activity.

## Acceptance Criteria

```gherkin
AC1: Given a project is selected
     When I press 'f'
     Then favorite status toggles immediately
     And ⭐ indicator appears/disappears in project row
     And feedback shows: "⭐ Favorited" or "☆ Unfavorited"

AC2: Given project is favorited
     When displayed in project list
     Then it displays with ⭐ prefix
     And it never auto-hibernates (Epic 5 implementation)

AC3: Given I run `vibe favorite client-bravo` from CLI
     When the command executes
     Then favorite is toggled via CLI
     And exit code 0 on success

AC4: Given I run `vibe favorite client-bravo --off` from CLI
     When the command executes
     Then favorite is explicitly removed via CLI
     And exit code 0 on success

AC5: Given project is not favorited
     When I run `vibe favorite client-bravo --off`
     Then no error occurs (idempotent)
     And message shows current state: "☆ client-bravo is not favorited"

AC6: Given favorite toggle completes
     When the operation succeeds
     Then detail panel updates if visible
     And project list re-renders with updated indicator
```

## Tasks / Subtasks

- [x] **Task 1: Implement TUI favorite toggle handler** (AC: 1, 6)

  **PATTERN REFERENCE:** Follow `handleKeyMsg` in `model.go:517-576` for key handling pattern, and `saveNoteCmd` pattern for async save operation.

  - [x] 1.1 Add favorite message types to `model.go`:
    ```go
    // favoriteSavedMsg signals favorite was toggled successfully (Story 3.8).
    type favoriteSavedMsg struct {
        projectID   string
        isFavorite  bool
    }

    // favoriteSaveErrorMsg signals favorite save failed (Story 3.8).
    type favoriteSaveErrorMsg struct {
        err error
    }
    ```

  - [x] 1.2 Add KeyFavorite case to handleKeyMsg in `model.go` (after KeyNotes case ~line 564):
    ```go
    case KeyFavorite:
        if len(m.projects) == 0 {
            return m, nil // No project to favorite
        }
        return m.toggleFavorite()
    ```

  - [x] 1.3 Implement toggleFavorite method in `model.go`:
    ```go
    // toggleFavorite toggles the favorite status of the selected project (Story 3.8).
    func (m Model) toggleFavorite() (tea.Model, tea.Cmd) {
        selected := m.projectList.SelectedProject()
        if selected == nil {
            return m, nil
        }

        newFavorite := !selected.IsFavorite
        return m, m.saveFavoriteCmd(selected.ID, newFavorite)
    }
    ```

  - [x] 1.4 Implement saveFavoriteCmd in `model.go`:
    ```go
    // saveFavoriteCmd creates a command that saves the favorite status to repository (Story 3.8).
    func (m Model) saveFavoriteCmd(projectID string, isFavorite bool) tea.Cmd {
        return func() tea.Msg {
            ctx := context.Background()

            // Find project
            project, err := m.repository.FindByID(ctx, projectID)
            if err != nil {
                return favoriteSaveErrorMsg{err: err}
            }

            // Update favorite status
            project.IsFavorite = isFavorite
            project.UpdatedAt = time.Now()

            // Save
            if err := m.repository.Save(ctx, project); err != nil {
                return favoriteSaveErrorMsg{err: err}
            }

            return favoriteSavedMsg{projectID: projectID, isFavorite: isFavorite}
        }
    }
    ```

  - [x] 1.5 Handle favorite save messages in Update function (after noteSaveErrorMsg case ~line 401):
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
        // Set feedback message
        var feedback string
        if msg.isFavorite {
            feedback = "⭐ Favorited"
        } else {
            feedback = "☆ Unfavorited"
        }
        m.statusBar.SetRefreshComplete(feedback)
        // Clear after 3 seconds
        return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
            return clearFavoriteFeedbackMsg{}
        })

    case favoriteSaveErrorMsg:
        m.statusBar.SetRefreshComplete("✗ Failed to toggle favorite")
        return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
            return clearFavoriteFeedbackMsg{}
        })

    case clearFavoriteFeedbackMsg:
        m.statusBar.SetRefreshComplete("")
        return m, nil
    ```

  - [x] 1.6 Add clearFavoriteFeedbackMsg type:
    ```go
    // clearFavoriteFeedbackMsg signals to clear favorite feedback message (Story 3.8).
    type clearFavoriteFeedbackMsg struct{}
    ```

- [x] **Task 2: Add favorite indicator to project row** (AC: 1, 2, 6)

  **PATTERN REFERENCE:** Modify `renderRow` in `delegate.go:110-167` to include ⭐ indicator.

  - [x] 2.1 Add favorite style constant in `delegate.go` (add after dimStyle ~line 43):
    ```go
    // favoriteStyle mirrors tui.FavoriteStyle (ANSI color 5 magenta) - keep in sync with styles.go
    favoriteStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("5")) // Magenta
    ```

  - [x] 2.2 Add favorite column constant in `delegate.go` (add after colSelection ~line 18):
    ```go
    colFavorite  = 2  // "⭐" or "  "
    ```

  - [x] 2.3 Update calculateNameWidth to account for favorite column:
    ```go
    func (d ProjectItemDelegate) calculateNameWidth() int {
        // Calculate available space for name
        // width - selection - favorite - indicator - stage - waiting - time - spacing
        nameWidth := d.width - colSelection - colFavorite - colIndicator - colStage - colWaiting - colTime - 5

        if nameWidth < colNameMin {
            nameWidth = colNameMin
        }
        if nameWidth < 1 {
            nameWidth = 1
        }

        return nameWidth
    }
    ```

  - [x] 2.4 Update renderRow to include favorite indicator in `delegate.go`:

    **INSERTION POINT:** After selection indicator (line 119), BEFORE project name (line 122).
    The new column order is: `[Selection] → [Favorite] → [Name] → [Indicator] → ...`

    ```go
    // Selection indicator (existing code ~line 114-119)
    if isSelected {
        sb.WriteString("> ")
    } else {
        sb.WriteString("  ")
    }

    // Favorite indicator (Story 3.8) - INSERT THIS BLOCK
    if item.Project.IsFavorite {
        sb.WriteString(favoriteStyle.Render("⭐"))
    } else {
        sb.WriteString("  ")
    }

    // Project name (existing code continues ~line 122)
    ```

  - [x] 2.5 Add IsFavorite display to detail_panel.go (NOT currently implemented):

    **VERIFIED:** `detail_panel.go:140-145` only displays Notes, NOT IsFavorite. Must add this.

    In `detail_panel.go` `renderProject()` method, add after Notes field (line 145):
    ```go
    // Favorite status (Story 3.8)
    favorite := "No"
    if p.IsFavorite {
        favorite = "⭐ Yes"
    }
    lines = append(lines, formatField("Favorite", favorite))
    ```

- [x] **Task 3: Create CLI favorite command** (AC: 3, 4, 5)

  **CRITICAL:** Use dependency injection pattern like `add.go:22-32`, NOT direct initialization.
  The `repository` package variable is already defined in `add.go`.

  - [x] 3.1 Create `internal/adapters/cli/favorite.go`:
    ```go
    package cli

    import (
        "fmt"
        "time"

        "github.com/spf13/cobra"

        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    )

    // favoriteOff holds the --off flag value
    var favoriteOff bool

    // ResetFavoriteFlags resets favorite command flags for testing.
    func ResetFavoriteFlags() {
        favoriteOff = false
    }

    // newFavoriteCmd creates the favorite command.
    func newFavoriteCmd() *cobra.Command {
        cmd := &cobra.Command{
            Use:   "favorite <project-name>",
            Short: "Toggle or remove favorite status for a project",
            Long: `Toggle or remove favorite status for a tracked project.

By default, toggles the favorite status (on→off or off→on).
Use --off to explicitly remove favorite status.

Favorited projects:
  - Display with ⭐ prefix in dashboard
  - Never auto-hibernate (always visible)

Examples:
  vibe favorite my-project       # Toggle favorite status
  vibe favorite my-project --off # Remove favorite status`,
            Args: cobra.ExactArgs(1),
            RunE: runFavorite,
        }

        cmd.Flags().BoolVar(&favoriteOff, "off", false, "Remove favorite status (instead of toggle)")

        return cmd
    }

    // RegisterFavoriteCommand registers the favorite command with the given parent.
    // Used for testing to create fresh command trees.
    func RegisterFavoriteCommand(parent *cobra.Command) {
        parent.AddCommand(newFavoriteCmd())
    }

    func init() {
        RootCmd.AddCommand(newFavoriteCmd())
    }

    func runFavorite(cmd *cobra.Command, args []string) error {
        ctx := cmd.Context()

        // Use package-level repository (injected via SetRepository in main.go)
        if repository == nil {
            return fmt.Errorf("repository not initialized")
        }

        projectName := args[0]

        // Find project by name or display name
        projects, err := repository.FindAll(ctx)
        if err != nil {
            return fmt.Errorf("failed to load projects: %w", err)
        }

        var targetProject *domain.Project
        for _, p := range projects {
            if p.Name == projectName || p.DisplayName == projectName {
                targetProject = p
                break
            }
        }

        if targetProject == nil {
            return fmt.Errorf("%w: %s", domain.ErrProjectNotFound, projectName)
        }

        // Determine new favorite status
        var newFavorite bool
        if favoriteOff {
            // Explicit --off: remove favorite
            if !targetProject.IsFavorite {
                // Already not favorited - idempotent success (AC5)
                fmt.Fprintf(cmd.OutOrStdout(), "☆ %s is not favorited\n", projectName)
                return nil
            }
            newFavorite = false
        } else {
            // Toggle mode
            newFavorite = !targetProject.IsFavorite
        }

        // Update and save
        targetProject.IsFavorite = newFavorite
        targetProject.UpdatedAt = time.Now()

        if err := repository.Save(ctx, targetProject); err != nil {
            return fmt.Errorf("failed to save favorite status: %w", err)
        }

        // Success output
        if newFavorite {
            fmt.Fprintf(cmd.OutOrStdout(), "⭐ Favorited: %s\n", projectName)
        } else {
            fmt.Fprintf(cmd.OutOrStdout(), "☆ Unfavorited: %s\n", projectName)
        }

        return nil
    }
    ```

  - [x] 3.2 Verify command registration:
    - `init()` registers with `RootCmd` (production)
    - `RegisterFavoriteCommand()` for test isolation
    - Uses `cmd.OutOrStdout()` for testable output

- [x] **Task 4: Add tests** (AC: all)

  - [x] 4.1 Create `internal/adapters/tui/model_favorite_test.go`:
    ```go
    package tui

    import (
        "testing"

        tea "github.com/charmbracelet/bubbletea"

        "github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    )

    func TestModel_FavoriteKey_TogglesOn(t *testing.T) {
        // Setup: Model with non-favorited project
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: false}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.repository = repo

        // Action: Send 'f' key
        newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
        _ = newModel.(Model)

        // Assert: Command returned for async save
        if cmd == nil {
            t.Error("expected command to be returned for async save")
        }
    }

    func TestModel_FavoriteKey_TogglesOff(t *testing.T) {
        // Setup: Model with favorited project
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: true}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.repository = repo

        // Action: Send 'f' key
        newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
        _ = newModel.(Model)

        // Assert: Command returned for async save
        if cmd == nil {
            t.Error("expected command to be returned for async save")
        }
    }

    func TestModel_FavoriteKey_IgnoredWhenNoProjects(t *testing.T) {
        // Setup: Model WITHOUT projects
        repo := newMockRepository()
        m := NewModel(repo)
        m.projects = nil

        // Action
        _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

        // Assert: No command returned (no-op)
        if cmd != nil {
            t.Error("expected no command when no projects")
        }
    }

    func TestModel_FavoriteSavedMsg_UpdatesProjectAndFeedback(t *testing.T) {
        // Setup: Model with project
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: false}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.statusBar = components.NewStatusBarModel(80)

        // Action: Receive favoriteSavedMsg
        newModel, _ := m.Update(favoriteSavedMsg{projectID: "1", isFavorite: true})
        updated := newModel.(Model)

        // Assert: Project updated
        if !updated.projects[0].IsFavorite {
            t.Error("expected project IsFavorite to be true")
        }
    }
    ```

  - [x] 4.2 Create `internal/adapters/cli/favorite_test.go`:
    ```go
    package cli_test

    import (
        "bytes"
        "fmt"
        "strings"
        "testing"

        "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    )

    // mockRepo implements ports.ProjectRepository for testing.
    // Add these fields to existing mockRepo if it exists, or create new:
    type mockRepo struct {
        projects   []*domain.Project
        findAllErr error // Return this error from FindAll
        saveErr    error // Return this error from Save
    }

    func TestFavoriteCmd_ToggleOn(t *testing.T) {
        // Setup
        projects := []*domain.Project{
            {ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
        }
        mockRepo := &mockRepo{projects: projects}
        root := cli.NewRootCmd()
        cli.RegisterFavoriteCommand(root)
        cli.SetRepository(mockRepo)
        cli.ResetFavoriteFlags()

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"favorite", "test-project"})

        // Execute
        err := root.Execute()

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "⭐ Favorited") {
            t.Errorf("expected '⭐ Favorited', got: %s", out.String())
        }
        if !projects[0].IsFavorite {
            t.Error("expected project IsFavorite to be true")
        }
    }

    func TestFavoriteCmd_ToggleOff(t *testing.T) {
        // Setup
        projects := []*domain.Project{
            {ID: "1", Path: "/test", Name: "test-project", IsFavorite: true},
        }
        mockRepo := &mockRepo{projects: projects}
        root := cli.NewRootCmd()
        cli.RegisterFavoriteCommand(root)
        cli.SetRepository(mockRepo)
        cli.ResetFavoriteFlags()

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"favorite", "test-project"})

        // Execute
        err := root.Execute()

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "☆ Unfavorited") {
            t.Errorf("expected '☆ Unfavorited', got: %s", out.String())
        }
        if projects[0].IsFavorite {
            t.Error("expected project IsFavorite to be false")
        }
    }

    func TestFavoriteCmd_ExplicitOff(t *testing.T) {
        // Setup
        projects := []*domain.Project{
            {ID: "1", Path: "/test", Name: "test-project", IsFavorite: true},
        }
        mockRepo := &mockRepo{projects: projects}
        root := cli.NewRootCmd()
        cli.RegisterFavoriteCommand(root)
        cli.SetRepository(mockRepo)
        cli.ResetFavoriteFlags()

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"favorite", "test-project", "--off"})

        // Execute
        err := root.Execute()

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "☆ Unfavorited") {
            t.Errorf("expected '☆ Unfavorited', got: %s", out.String())
        }
        if projects[0].IsFavorite {
            t.Error("expected project IsFavorite to be false")
        }
    }

    func TestFavoriteCmd_OffIdempotent(t *testing.T) {
        // Setup: Project already not favorited
        projects := []*domain.Project{
            {ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
        }
        mockRepo := &mockRepo{projects: projects}
        root := cli.NewRootCmd()
        cli.RegisterFavoriteCommand(root)
        cli.SetRepository(mockRepo)
        cli.ResetFavoriteFlags()

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"favorite", "test-project", "--off"})

        // Execute
        err := root.Execute()

        // Assert: No error, idempotent message
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "is not favorited") {
            t.Errorf("expected idempotent message, got: %s", out.String())
        }
    }

    func TestFavoriteCmd_ProjectNotFound(t *testing.T) {
        // Setup
        root := cli.NewRootCmd()
        cli.RegisterFavoriteCommand(root)
        cli.SetRepository(&mockRepo{projects: nil})
        cli.ResetFavoriteFlags()

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"favorite", "nonexistent"})

        // Execute
        err := root.Execute()

        // Assert
        if err == nil {
            t.Fatal("expected error for non-existent project")
        }
        if !strings.Contains(err.Error(), "not found") {
            t.Errorf("expected 'not found' error, got: %v", err)
        }
    }

    func TestFavoriteCmd_FindAllError(t *testing.T) {
        // Setup: Mock repo that returns error on FindAll
        mockRepo := &mockRepo{findAllErr: fmt.Errorf("database error")}
        root := cli.NewRootCmd()
        cli.RegisterFavoriteCommand(root)
        cli.SetRepository(mockRepo)
        cli.ResetFavoriteFlags()

        root.SetArgs([]string{"favorite", "test-project"})

        // Execute
        err := root.Execute()

        // Assert
        if err == nil {
            t.Fatal("expected error")
        }
        if !strings.Contains(err.Error(), "failed to load projects") {
            t.Errorf("expected load error, got: %v", err)
        }
    }

    func TestFavoriteCmd_SaveError(t *testing.T) {
        // Setup: Mock repo that returns error on Save
        projects := []*domain.Project{
            {ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
        }
        mockRepo := &mockRepo{projects: projects, saveErr: fmt.Errorf("save failed")}
        root := cli.NewRootCmd()
        cli.RegisterFavoriteCommand(root)
        cli.SetRepository(mockRepo)
        cli.ResetFavoriteFlags()

        root.SetArgs([]string{"favorite", "test-project"})

        // Execute
        err := root.Execute()

        // Assert
        if err == nil {
            t.Fatal("expected error")
        }
        if !strings.Contains(err.Error(), "failed to save") {
            t.Errorf("expected save error, got: %v", err)
        }
    }
    ```

  - [x] 4.3 Add delegate favorite indicator test in `delegate_test.go`:
    ```go
    func TestProjectItemDelegate_RendersFavoriteIndicator(t *testing.T) {
        tests := []struct {
            name       string
            isFavorite bool
            wantStar   bool
        }{
            {"favorited shows star", true, true},
            {"not favorited no star", false, false},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                project := &domain.Project{
                    ID:         "1",
                    Name:       "test",
                    Path:       "/test",
                    IsFavorite: tt.isFavorite,
                }
                item := ProjectItem{Project: project}
                delegate := NewProjectItemDelegate(80)

                var buf bytes.Buffer
                // Create mock list model for rendering
                items := []list.Item{item}
                l := list.New(items, delegate, 80, 10)

                delegate.Render(&buf, l, 0, item)
                output := buf.String()

                hasStar := strings.Contains(output, "⭐")
                if tt.wantStar && !hasStar {
                    t.Errorf("expected ⭐ indicator, got: %s", output)
                }
                if !tt.wantStar && hasStar {
                    t.Errorf("did not expect ⭐ indicator, got: %s", output)
                }
            })
        }
    }
    ```

  - [x] 4.4 Run verification:
    ```bash
    make test   # All tests pass
    make lint   # No lint errors
    make build  # Successful build
    ```

## Dev Notes

### Current State Analysis

**KeyFavorite constant:** Already exists in `keys.go:20` as `KeyFavorite = "f"`. No changes needed.

**IsFavorite field:** Already exists in `domain/project.go:21`. Boolean type, defaults to false.

**SQLite Schema:** IsFavorite column already exists via `is_favorite INTEGER DEFAULT 0` in projects table. No migration needed - toggling the boolean and calling `repository.Save()` persists the change.

**Model Structure:** Has patterns from Story 3.7 (notes) for async save operations and feedback messages.

**Delegate Structure:** Column constants defined in `delegate.go:17-24`. Need to add `colFavorite` and update `calculateNameWidth`.

### Pattern Reference: Story 3.7

Story 3.7 (Notes) established the async save pattern:
1. Key press → method call → return `tea.Cmd`
2. Cmd executes `repository.FindByID`, updates field, calls `repository.Save`
3. Returns success/error message
4. Update handler updates local state, shows feedback, clears after 3s

Follow this exact pattern for favorite toggle.

### Feedback Messages

Per AC1, feedback should be:
- "⭐ Favorited" when toggling ON
- "☆ Unfavorited" when toggling OFF

Use `statusBar.SetRefreshComplete()` for display (same pattern as Story 3.7).

### Column Layout Impact

Adding `colFavorite = 2` will reduce available name width by 2 characters. This is acceptable given the value of visual favorite indication.

Updated column order in renderRow:
```
[Selection][Favorite][Name][Indicator][Stage][Waiting][Time]
   2          2      var      3       10       14      8
```

### CLI Flag Behavior

The `--off` flag provides explicit favorite removal:
- Without `--off`: Toggle behavior (FR54: "toggle")
- With `--off`: Explicit removal (FR31: "remove favorite status")

This matches the ergonomics of other CLI tools where `--off` is the explicit negation.

### Project Context Compliance

Per `docs/project-context.md`:
- Context first: All operations use `ctx context.Context`
- Error wrapping: Use `fmt.Errorf("...: %w", err)`
- Co-locate tests: `*_test.go` next to source
- Use domain errors: Return `ErrProjectNotFound`, not raw errors

### Edge Cases to Handle

1. **No projects:** Ignore 'f' key (checked before `toggleFavorite`)
2. **Project deleted while toggling:** Save will fail, show error feedback
3. **Repository error:** Show "✗ Failed to toggle favorite" feedback
4. **CLI idempotent --off:** Show "is not favorited" without error (AC5)

### Architecture Compliance

Per `docs/architecture.md`:
- Hexagonal architecture: TUI adapter calls repository directly (acceptable for simple field update)
- No new service needed: Favorite toggle is simple field update
- Repository.Save handles update

### Future Considerations (Epic 5)

Story 5.2 (Auto-Hibernation) states: "When project is favorited, then it never auto-hibernates."
This story does NOT implement that behavior - it just toggles the IsFavorite flag. Epic 5 will check `IsFavorite` when deciding whether to auto-hibernate.

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 3.8 requirements - lines 1398-1432)
- docs/architecture.md (Hexagonal patterns, TUI adapter structure)
- docs/project-context.md (Go conventions, testing rules)
- internal/adapters/tui/model.go (Current Model structure, message patterns from Story 3.7)
- internal/adapters/tui/keys.go (KeyFavorite already defined at line 20)
- internal/adapters/tui/components/delegate.go (Row rendering pattern, column constants)
- internal/core/domain/project.go (IsFavorite field at line 21)
- internal/core/ports/repository.go (Repository interface)
- internal/adapters/cli/add.go (CLI DI pattern, repository package variable)
- docs/sprint-artifacts/stories/epic-3/3-7-project-notes-view-and-edit.md (Previous story patterns)
- Git history: Stories 3.1-3.7 implementation patterns

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story drafting phase.

### Completion Notes List

- ✅ Task 1: Implemented TUI favorite toggle handler - Added message types (favoriteSavedMsg, favoriteSaveErrorMsg, clearFavoriteFeedbackMsg), KeyFavorite case in handleKeyMsg, toggleFavorite and saveFavoriteCmd methods. Follows Story 3.7 async save pattern.
- ✅ Task 2: Added favorite indicator to project row - Added colFavorite column constant, favoriteStyle (magenta), updated calculateNameWidth, added ⭐ indicator in renderRow, added IsFavorite display in detail_panel.go.
- ✅ Task 3: Created CLI favorite command - Implemented `vibe favorite <project> [--off]` with toggle behavior, idempotent --off flag, proper error handling. Uses existing DI pattern.
- ✅ Task 4: Added comprehensive tests - 12 TUI tests, 11 CLI tests, 2 delegate tests. All pass with `make test && make lint && make build`.

### File List

**Modified:**
- `internal/adapters/tui/model.go` - Add favorite toggle handler, message types, Update cases
- `internal/adapters/tui/components/delegate.go` - Add favorite indicator column, favoriteStyle, update renderRow
- `internal/adapters/tui/components/detail_panel.go` - Add IsFavorite display field after Notes
- `internal/adapters/tui/components/delegate_test.go` - Add favorite indicator tests

**Created:**
- `internal/adapters/cli/favorite.go` - CLI favorite command (uses existing DI pattern)
- `internal/adapters/cli/favorite_test.go` - CLI favorite tests
- `internal/adapters/tui/model_favorite_test.go` - TUI favorite toggle behavior tests

**Existing (Reference Only):**
- `internal/adapters/cli/add.go:22-32` - Package-level `repository` variable + SetRepository()
- `internal/adapters/tui/keys.go:20` - `KeyFavorite = "f"` already defined
- `internal/core/domain/project.go:21` - `IsFavorite bool` already defined

## Change Log

| Date | Change |
|------|--------|
| 2025-12-17 | Story created with ready-for-dev status by SM Agent (Bob) in YOLO mode. Comprehensive developer context included with all technical decisions, code patterns, and test specifications. |
| 2025-12-17 | Validation improvements applied: (1) Task 2.5 updated with concrete detail_panel.go IsFavorite implementation - verified NOT present in current code; (2) Added `fmt` import to favorite_test.go; (3) Added style sync comment to favoriteStyle definition; (4) Added mockRepo struct definition with findAllErr/saveErr fields for error testing; (5) Clarified column insertion order in Task 2.4; (6) Updated Files to Modify to include detail_panel.go. |
| 2025-12-17 | Implementation complete by Dev Agent (Amelia). All 4 tasks completed: TUI favorite toggle handler (model.go), favorite indicator in delegate.go and detail_panel.go, CLI favorite command, comprehensive tests (25 total tests). All acceptance criteria satisfied. Tests pass, lint clean, build successful. |
| 2025-12-17 | Code review by Dev Agent (Amelia): 0 Critical, 4 Medium, 3 Low issues found. Fixes applied: (1) Removed unused UpdatePath method from model_favorite_test.go; (2) Added spacing breakdown comment in delegate.go:103; (3) Added TestModel_FavoriteSavedMsg_ProjectListReRendersWithIndicator test for AC6 coverage; (4) Updated column comment for accuracy. All tests pass, lint clean, build successful. Status updated to Done. |
