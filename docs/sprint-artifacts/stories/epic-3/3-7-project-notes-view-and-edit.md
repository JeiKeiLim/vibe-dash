# Story 3.7: Project Notes (View & Edit)

**Status:** done

## Executive Summary

Add note editing capability to projects via TUI ('n' key) and CLI (`vibe note`). Implementation reuses existing patterns: KeyNotes constant exists, Notes field exists in domain.Project, detail panel already displays notes. Main work is adding textinput dialog (follow `renderHelpOverlay` pattern), handling Enter/Esc keys, and CLI command using existing DI pattern from add.go. No database migration needed - Notes column already exists via domain mapping.

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/model.go` (KeyNotes handling), `internal/adapters/cli/note.go` (new file) |
| **Key Dependencies** | ProjectRepository (`core/ports/repository.go`), DetailPanelModel (`components/detail_panel.go`), Bubbles textinput component |
| **Files to Modify** | `model.go` (add note editing state/handlers), `components/detail_panel.go` (notes display already exists), `views.go` (add note editor dialog render) |
| **Files to Create** | `cli/note.go`, `cli/note_test.go`, `model_notes_test.go`, `components/note_editor.go` (optional - can be inline) |
| **Location** | `internal/adapters/tui/`, `internal/adapters/cli/` |
| **Interfaces Used** | `ports.ProjectRepository.Save()` |

### Quick Task Summary (5 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Add note editing state to TUI model | isEditingNote bool, noteInput textinput.Model, original note tracking |
| 2 | Implement TUI note editor handler | Handle 'n' key, show inline editor, handle Enter/Esc |
| 3 | Add note editor dialog rendering | Inline text input with instructions, centered modal |
| 4 | Create CLI note command | `vibe note <project> "note text"` for non-interactive note editing |
| 5 | Add tests | TUI note editing behavior + CLI note command tests |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Editor component | Bubbles textinput | Standard library, consistent with UX spec |
| Dialog style | Inline modal overlay | Matches confirmation dialog pattern (Story 3.9 future) |
| Note persistence | Direct repository.Save() | Simple - update project.Notes field and save |
| CLI command | `vibe note <project> "text"` | Matches FR55 - set notes via CLI |
| Empty note handling | Clear note from DB | Per AC: "When note is empty and I save, note is cleared" |

## Story

**As a** user,
**I want** to add notes to projects,
**So that** I can capture context that detection can't know.

## Acceptance Criteria

```gherkin
AC1: Given a project is selected
     When I press 'n'
     Then inline note editor opens:
       """
       ┌─ Edit note for "client-bravo" ────────────────┐
       │ > Waiting on client API specs█                │
       │ [Enter] save  [Esc] cancel                    │
       └───────────────────────────────────────────────┘
       """

AC2: Given note editor is open
     When I type text and press Enter
     Then note is saved to database
     And detail panel updates to show note
     And feedback shows: "✓ Note saved"

AC3: Given note editor is open
     When I press Esc
     Then edit is cancelled
     And original note preserved

AC4: Given note editor is open
     When note is empty and I save
     Then note is cleared from database
     And detail panel shows "(none)" for notes

AC5: Given I run `vibe note client-bravo "New note"` from CLI
     When the command executes
     Then note is set via CLI
     And exit code 0 on success

AC6: Given I run `vibe note client-bravo` from CLI (no note argument)
     When the command executes
     Then current note is displayed
     Or if no note exists, shows "(no note set)"

AC7: Given note editor is open
     When I navigate with j/k keys
     Then navigation is blocked (input captured)
     And text input receives the keys
```

## Tasks / Subtasks

- [x] **Task 1: Add note editing state to TUI model** (AC: 1, 7)
  - [x] 1.1 Add note-editing related fields to Model struct in `model.go`:
    ```go
    // Note editing state (Story 3.7)
    isEditingNote  bool
    noteInput      textinput.Model  // From charmbracelet/bubbles
    originalNote   string           // For cancel restoration
    noteEditTarget *domain.Project  // Project being edited
    ```
  - [x] 1.2 Add note message types:
    ```go
    // noteSavedMsg signals note was saved successfully (Story 3.7).
    type noteSavedMsg struct {
        projectID string
        newNote   string
    }

    // noteSaveErrorMsg signals note save failed (Story 3.7).
    type noteSaveErrorMsg struct {
        err error
    }

    // clearNoteFeedbackMsg signals to clear note feedback message (Story 3.7).
    type clearNoteFeedbackMsg struct{}
    ```
  - [x] 1.3 Add feedback state for note operations:
    ```go
    // Add to Model struct
    noteFeedback string  // "✓ Note saved" or error message
    ```
  - [x] 1.4 Import bubbles textinput in model.go:
    ```go
    import "github.com/charmbracelet/bubbles/textinput"
    ```

- [x] **Task 2: Implement TUI note editor handler** (AC: 1, 2, 3, 4, 7)
  - [x] 2.1 Add KeyNotes case to handleKeyMsg in `model.go`:
    ```go
    case KeyNotes:
        if m.isEditingNote {
            return m, nil // Ignore if already editing
        }
        if len(m.projects) == 0 {
            return m, nil // No project to edit
        }
        return m.startNoteEditing()
    ```
  - [x] 2.2 Implement startNoteEditing method:
    ```go
    // startNoteEditing opens the note editor for the selected project.
    func (m Model) startNoteEditing() (tea.Model, tea.Cmd) {
        selected := m.projectList.SelectedProject()
        if selected == nil {
            return m, nil
        }

        m.isEditingNote = true
        m.noteEditTarget = selected
        m.originalNote = selected.Notes

        // Initialize text input with current note
        ti := textinput.New()
        ti.Placeholder = "Enter note..."
        ti.Focus()
        ti.CharLimit = 500  // Reasonable limit for notes
        ti.Width = m.width - 10  // Leave padding for dialog border
        ti.SetValue(selected.Notes)
        m.noteInput = ti

        return m, textinput.Blink
    }
    ```
  - [x] 2.3 Add note editing mode to Update function (route input to textinput):
    ```go
    // In Update, at the start of tea.KeyMsg handling:
    case tea.KeyMsg:
        // Handle note editing mode first (captures all input)
        if m.isEditingNote {
            return m.handleNoteEditingKeyMsg(msg)
        }
        // ... existing key handling
    ```
  - [x] 2.4 Implement handleNoteEditingKeyMsg:
    ```go
    // handleNoteEditingKeyMsg processes keyboard input during note editing (Story 3.7).
    func (m Model) handleNoteEditingKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.Type {
        case tea.KeyEnter:
            // Save note
            return m.saveNote()
        case tea.KeyEsc:
            // Cancel editing, restore original
            m.isEditingNote = false
            m.noteInput = textinput.Model{} // Clear
            return m, nil
        }

        // Forward to text input
        var cmd tea.Cmd
        m.noteInput, cmd = m.noteInput.Update(msg)
        return m, cmd
    }
    ```
  - [x] 2.5 Implement saveNote method:
    ```go
    // saveNote saves the note to the repository (Story 3.7).
    func (m Model) saveNote() (tea.Model, tea.Cmd) {
        newNote := strings.TrimSpace(m.noteInput.Value())
        m.isEditingNote = false

        return m, m.saveNoteCmd(m.noteEditTarget.ID, newNote)
    }

    // saveNoteCmd creates a command that saves the note to repository.
    func (m Model) saveNoteCmd(projectID, note string) tea.Cmd {
        return func() tea.Msg {
            ctx := context.Background()

            // Find project
            project, err := m.repository.FindByID(ctx, projectID)
            if err != nil {
                return noteSaveErrorMsg{err: err}
            }

            // Update note
            project.Notes = note
            project.UpdatedAt = time.Now()

            // Save
            if err := m.repository.Save(ctx, project); err != nil {
                return noteSaveErrorMsg{err: err}
            }

            return noteSavedMsg{projectID: projectID, newNote: note}
        }
    }
    ```
  - [x] 2.6 Handle note save messages in Update:
    ```go
    case noteSavedMsg:
        // Update local project state
        for _, p := range m.projects {
            if p.ID == msg.projectID {
                p.Notes = msg.newNote
                break
            }
        }
        // Update detail panel
        m.detailPanel.SetProject(m.projectList.SelectedProject())
        // Set feedback message
        m.noteFeedback = "✓ Note saved"
        // Clear after 3 seconds
        return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
            return clearNoteFeedbackMsg{}
        })

    case noteSaveErrorMsg:
        m.noteFeedback = "✗ Failed to save note"
        return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
            return clearNoteFeedbackMsg{}
        })

    case clearNoteFeedbackMsg:
        m.noteFeedback = ""
        return m, nil
    ```

- [x] **Task 3: Add note editor dialog rendering** (AC: 1)

  **PATTERN REFERENCE:** Follow `renderHelpOverlay()` in `views.go:59-106` for dialog styling and centering. Use same `boxStyle`, `titleStyle`, and `lipgloss.Place()` centering approach.

  - [x] 3.1 Add renderNoteEditor function in `views.go`:
    ```go
    // renderNoteEditor renders the inline note editor dialog (Story 3.7).
    func renderNoteEditor(projectName string, input textinput.Model, width, height int) string {
        // Dialog dimensions
        dialogWidth := min(width-4, 60)

        // Title
        title := fmt.Sprintf("Edit note for \"%s\"", projectName)

        // Input line with > prefix
        inputLine := "> " + input.View()

        // Instructions
        instructions := DimStyle.Render("[Enter] save  [Esc] cancel")

        // Content
        content := lipgloss.JoinVertical(lipgloss.Left,
            inputLine,
            "",
            instructions,
        )

        // Dialog box style
        dialogStyle := BorderStyle.
            Width(dialogWidth).
            Padding(0, 1)

        dialogBox := dialogStyle.Render(content)

        // Add title header
        titleStyle := TitleStyle.Bold(true)
        header := titleStyle.Render(title)

        dialog := lipgloss.JoinVertical(lipgloss.Left, header, dialogBox)

        // Center in terminal
        return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
    }
    ```
  - [x] 3.2 Update View() in model.go to render note editor:
    ```go
    // In View(), after help overlay check:
    // Render note editor dialog (overlays everything)
    if m.isEditingNote && m.noteEditTarget != nil {
        projectName := project.EffectiveName(m.noteEditTarget)
        return renderNoteEditor(projectName, m.noteInput, m.width, m.height)
    }
    ```
  - [x] 3.3 Add feedback message to status bar or main view:
    ```go
    // In renderDashboard or status bar, show feedback if set
    // Option A: Add to status bar
    if m.noteFeedback != "" {
        // Show in status bar
        m.statusBar.SetFeedback(m.noteFeedback)
    }
    // OR Option B: Show inline in dashboard (simpler)
    ```
    **Note:** Simplest approach is to show feedback inline temporarily. Can use existing status bar SetRefreshComplete pattern.

- [x] **Task 4: Create CLI note command** (AC: 5, 6)

  **CRITICAL:** Use dependency injection pattern like `add.go`, NOT direct initialization.
  The `repository` package variable is already defined in `add.go`.

  - [x] 4.1 Create `internal/adapters/cli/note.go`:
    ```go
    package cli

    import (
        "fmt"

        "github.com/spf13/cobra"
    )

    // newNoteCmd creates the note command.
    func newNoteCmd() *cobra.Command {
        return &cobra.Command{
            Use:   "note <project-name> [note-text]",
            Short: "View or set notes for a project",
            Long: `View or set notes for a tracked project.

    If note-text is provided, sets the project note.
    If no note-text is provided, displays the current note.

    Examples:
      vibe note my-project "Waiting on API specs"   # Set note
      vibe note my-project ""                        # Clear note
      vibe note my-project                           # View current note`,
            Args: cobra.RangeArgs(1, 2),
            RunE: runNote,
        }
    }

    // RegisterNoteCommand registers the note command with the given parent.
    // Used for testing to create fresh command trees.
    func RegisterNoteCommand(parent *cobra.Command) {
        parent.AddCommand(newNoteCmd())
    }

    func init() {
        RootCmd.AddCommand(newNoteCmd())
    }

    func runNote(cmd *cobra.Command, args []string) error {
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

        // View mode (no note argument)
        if len(args) == 1 {
            if targetProject.Notes == "" {
                fmt.Fprintln(cmd.OutOrStdout(), "(no note set)")
            } else {
                fmt.Fprintln(cmd.OutOrStdout(), targetProject.Notes)
            }
            return nil
        }

        // Set mode (note argument provided)
        newNote := args[1]
        targetProject.Notes = newNote
        targetProject.UpdatedAt = time.Now()

        if err := repository.Save(ctx, targetProject); err != nil {
            return fmt.Errorf("failed to save note: %w", err)
        }

        if newNote == "" {
            fmt.Fprintln(cmd.OutOrStdout(), "✓ Note cleared")
        } else {
            fmt.Fprintln(cmd.OutOrStdout(), "✓ Note saved")
        }

        return nil
    }
    ```

  - [x] 4.2 Add required imports to note.go:
    ```go
    import (
        "fmt"
        "time"

        "github.com/spf13/cobra"

        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    )
    ```

  - [x] 4.3 Verify command registration:
    - `init()` registers with `RootCmd` (production)
    - `RegisterNoteCommand()` for test isolation
    - Uses `cmd.OutOrStdout()` for testable output

- [x] **Task 5: Add tests** (AC: all)

  - [x] 5.1 Create `internal/adapters/tui/model_notes_test.go`:
    ```go
    package tui

    import (
        "testing"

        tea "github.com/charmbracelet/bubbletea"

        "github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    )

    func TestModel_NotesKey_OpensEditor(t *testing.T) {
        // Setup: Model with projects
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "existing"}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)

        // Action: Send 'n' key
        newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
        updated := newModel.(Model)

        // Assert
        if !updated.isEditingNote {
            t.Error("expected isEditingNote to be true after pressing 'n'")
        }
        if updated.originalNote != "existing" {
            t.Errorf("expected originalNote to be 'existing', got %q", updated.originalNote)
        }
        if cmd == nil {
            t.Error("expected command (textinput.Blink) to be returned")
        }
    }

    func TestModel_NotesKey_IgnoredWhenNoProjects(t *testing.T) {
        // Setup: Model WITHOUT projects
        repo := newMockRepository()
        m := NewModel(repo)
        m.projects = nil

        // Action
        newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
        updated := newModel.(Model)

        // Assert
        if updated.isEditingNote {
            t.Error("expected isEditingNote to be false when no projects")
        }
    }

    func TestModel_NotesEditor_EscCancels(t *testing.T) {
        // Setup: Model in note editing mode
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "original"}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.isEditingNote = true
        m.originalNote = "original"
        m.noteInput.SetValue("modified but not saved")

        // Action: Press Esc
        newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
        updated := newModel.(Model)

        // Assert: Editing cancelled, original preserved
        if updated.isEditingNote {
            t.Error("expected isEditingNote to be false after Esc")
        }
        // Note should NOT be modified (cancel restores original)
    }

    func TestModel_NotesEditor_EnterSaves(t *testing.T) {
        // Setup: Model in note editing mode
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "original"}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.isEditingNote = true
        m.noteEditTarget = repo.projects[0]
        m.noteInput.SetValue("new note")
        m.repository = repo

        // Action: Press Enter
        newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
        updated := newModel.(Model)

        // Assert: Editing stopped, command returned
        if updated.isEditingNote {
            t.Error("expected isEditingNote to be false after Enter")
        }
        if cmd == nil {
            t.Error("expected save command to be returned")
        }
    }

    func TestModel_NotesEditor_NavigationBlocked(t *testing.T) {
        // Setup: Model in note editing mode
        repo := newMockRepository()
        repo.projects = []*domain.Project{{ID: "1"}, {ID: "2"}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
        m.isEditingNote = true
        m.noteInput.SetValue("typing")

        initialSelection := m.projectList.SelectedIndex()

        // Action: Press 'j' (navigation key)
        newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
        updated := newModel.(Model)

        // Assert: Selection should NOT change (input captures key)
        if updated.projectList.SelectedIndex() != initialSelection {
            t.Error("expected navigation to be blocked during note editing")
        }
    }
    ```

  - [x] 5.2 Create `internal/adapters/cli/note_test.go`:
    ```go
    package cli_test

    import (
        "bytes"
        "strings"
        "testing"
        "time"

        "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    )

    func TestNoteCmd_ViewNote(t *testing.T) {
        // Setup
        projects := []*domain.Project{
            {ID: "1", Path: "/test", Name: "test-project", Notes: "existing note"},
        }
        root := cli.NewRootCmd()
        cli.RegisterNoteCommand(root)
        cli.SetRepository(&mockRepo{projects: projects})

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"note", "test-project"})

        // Execute
        err := root.Execute()

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "existing note") {
            t.Errorf("expected note content, got: %s", out.String())
        }
    }

    func TestNoteCmd_ViewNoNote(t *testing.T) {
        // Setup: Project with no note
        projects := []*domain.Project{
            {ID: "1", Path: "/test", Name: "test-project", Notes: ""},
        }
        root := cli.NewRootCmd()
        cli.RegisterNoteCommand(root)
        cli.SetRepository(&mockRepo{projects: projects})

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"note", "test-project"})

        // Execute
        err := root.Execute()

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "(no note set)") {
            t.Errorf("expected '(no note set)', got: %s", out.String())
        }
    }

    func TestNoteCmd_SetNote(t *testing.T) {
        // Setup
        projects := []*domain.Project{
            {ID: "1", Path: "/test", Name: "test-project", Notes: ""},
        }
        mockRepo := &mockRepo{projects: projects}
        root := cli.NewRootCmd()
        cli.RegisterNoteCommand(root)
        cli.SetRepository(mockRepo)

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"note", "test-project", "new note content"})

        // Execute
        err := root.Execute()

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "✓ Note saved") {
            t.Errorf("expected success message, got: %s", out.String())
        }
        // Verify note was updated in mock
        if projects[0].Notes != "new note content" {
            t.Errorf("expected note to be updated, got: %s", projects[0].Notes)
        }
    }

    func TestNoteCmd_ClearNote(t *testing.T) {
        // Setup
        projects := []*domain.Project{
            {ID: "1", Path: "/test", Name: "test-project", Notes: "existing"},
        }
        mockRepo := &mockRepo{projects: projects}
        root := cli.NewRootCmd()
        cli.RegisterNoteCommand(root)
        cli.SetRepository(mockRepo)

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"note", "test-project", ""})

        // Execute
        err := root.Execute()

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "✓ Note cleared") {
            t.Errorf("expected clear message, got: %s", out.String())
        }
        if projects[0].Notes != "" {
            t.Errorf("expected note to be cleared, got: %s", projects[0].Notes)
        }
    }

    func TestNoteCmd_ProjectNotFound(t *testing.T) {
        // Setup
        root := cli.NewRootCmd()
        cli.RegisterNoteCommand(root)
        cli.SetRepository(&mockRepo{projects: nil})

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"note", "nonexistent"})

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
    ```

  - [x] 5.3 Run verification:
    ```bash
    make test   # All tests pass
    make lint   # No lint errors
    make build  # Successful build
    ```

## Dev Notes

### Current State Analysis

**KeyNotes constant:** Already exists in `keys.go` as `KeyNotes = "n"`. No changes needed.

**Detail Panel Notes Display:** Already implemented in `detail_panel.go:141-145`. Shows notes or "(none)" if empty.

**Model Structure:** Has patterns from Story 3.6 (refresh) for async operations and feedback messages.

**Project.Notes field:** Exists in `domain/project.go:23`. String type, no special validation.

**SQLite Schema:** Notes column already exists in `projects` table via domain.Project struct mapping in SQLite repository. **No migration needed** - the field is persisted via existing `repository.Save()`.

### Bubbles TextInput Component

Import from `github.com/charmbracelet/bubbles/textinput`. Key methods:
- `textinput.New()` - Create new input
- `ti.Focus()` - Enable cursor blinking
- `ti.SetValue(string)` - Set initial value
- `ti.Value()` - Get current value
- `ti.Update(msg)` - Handle key messages
- `ti.View()` - Render input

The textinput handles all text editing (cursor, backspace, typing) automatically.

### Input Capture During Note Editing (AC7)

When `isEditingNote == true`, ALL key messages must be routed to `handleNoteEditingKeyMsg` BEFORE the normal `handleKeyMsg`. This ensures:
- Navigation keys (j/k) go to text input, not project list
- Only Enter and Esc have special handling

### Note Editor Dialog Pattern

**CRITICAL:** Follow `renderHelpOverlay()` in `views.go:59-106` exactly:
1. Use `boxStyle` for border (line 89)
2. Use `titleStyle` for header (line 60)
3. Use `lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, ...)` for centering (line 105)
4. Return from View() to overlay everything - dialog replaces normal view

### Feedback Message Pattern

Reuse pattern from Story 3.6:
1. Set feedback string after operation
2. Start 3-second timer with `tea.Tick`
3. Clear feedback when timer fires

For simplicity, can use `statusBar.SetRefreshComplete()` since it already supports arbitrary messages.

### CLI Project Lookup

The CLI `note` command needs to find a project by name. Since `ProjectRepository` doesn't have a `FindByName` method:
1. Call `FindAll(ctx)`
2. Iterate and match by `Name` or `DisplayName`
3. Return `ErrProjectNotFound` if not found

This is a simple linear search - acceptable for typical project counts (<100).

### Project Context Compliance

Per `docs/project-context.md`:
- Context first: All service methods use `ctx context.Context`
- Error wrapping: Use `fmt.Errorf("...: %w", err)`
- Log at handling site: Log errors where handled
- Co-locate tests: `*_test.go` next to source
- Use domain errors: Return `ErrProjectNotFound`, not raw errors

### Edge Cases to Handle

1. **Empty note on Enter:** Clear the note (AC4)
2. **Very long note:** CharLimit in textinput (500 chars recommended)
3. **Project list empty:** Ignore 'n' key
4. **Already editing:** Ignore 'n' key (no nested dialogs)
5. **Project deleted while editing:** Save will fail, show error feedback

### Architecture Compliance

Per `docs/architecture.md`:
- Hexagonal architecture: TUI adapter calls repository directly (acceptable for simple CRUD)
- No new service needed: Note editing is simple field update
- Repository.Save handles both create and update

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 3.7 requirements - lines 1353-1397)
- docs/architecture.md (Hexagonal patterns, TUI adapter structure)
- docs/project-context.md (Go conventions, testing rules)
- internal/adapters/tui/model.go (Current Model structure, Update pattern, refresh message pattern)
- internal/adapters/tui/keys.go (KeyNotes already defined)
- internal/adapters/tui/views.go (Dialog rendering patterns)
- internal/adapters/tui/components/detail_panel.go (Notes display already exists)
- internal/core/domain/project.go (Notes field definition)
- internal/core/ports/repository.go (Repository interface)
- internal/adapters/cli/add.go (CLI DI pattern, repository package variable)
- docs/sprint-artifacts/stories/epic-3/3-6-manual-refresh.md (Previous story patterns, message handling)
- Git history: Stories 3.1-3.6 implementation patterns

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story drafting phase.

### Completion Notes List

(To be filled during implementation)

### File List

**Modified:**
- `internal/adapters/tui/model.go` - Add note editing state, messages, handlers
- `internal/adapters/tui/views.go` - Add note editor dialog rendering function (follow renderHelpOverlay pattern)
- `go.mod` - **NO CHANGE NEEDED** - `charmbracelet/bubbles v0.21.0` already in go.mod

**Created:**
- `internal/adapters/cli/note.go` - CLI note command (uses existing DI pattern)
- `internal/adapters/cli/note_test.go` - CLI note tests
- `internal/adapters/tui/model_notes_test.go` - TUI note editing behavior tests

**Existing (Reference Only):**
- `internal/adapters/cli/add.go:22-32` - Package-level `repository` variable + SetRepository()
- `internal/adapters/tui/keys.go:21` - `KeyNotes = "n"` already defined
- `internal/adapters/tui/components/detail_panel.go:141-145` - Notes display already implemented

## Change Log

| Date | Change |
|------|--------|
| 2025-12-17 | Story created with ready-for-dev status by SM Agent (Bob) in YOLO mode. Comprehensive developer context included with all technical decisions, code patterns, and test specifications. |
| 2025-12-17 | Validation improvements applied: Added executive summary, explicit dialog pattern reference to Task 3.1 (renderHelpOverlay:59-106), SQLite schema confirmation in Dev Notes, enhanced pattern references with line numbers. |
| 2025-12-17 | Code review fixes applied: (H1/H2) noteFeedback now displayed via status bar SetRefreshComplete(), (M1) dialog width bounds checking added (min 30), (M2) added saveNoteCmd async execution tests, (M3) added CLI error case tests for FindAll/Save failures and special characters. |
