package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// ============================================================================
// Story 3.7: Note Editing Tests
// ============================================================================

// notesMockRepository implements ports.ProjectRepository for testing.
type notesMockRepository struct {
	projects  []*domain.Project
	saveErr   error
	findErr   error
	deleteErr error
}

func (m *notesMockRepository) Save(_ context.Context, project *domain.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	// Update existing project or add new
	for i, p := range m.projects {
		if p.ID == project.ID {
			m.projects[i] = project
			return nil
		}
	}
	m.projects = append(m.projects, project)
	return nil
}

func (m *notesMockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.projects, nil
}

func (m *notesMockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.Path == path {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *notesMockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *notesMockRepository) Delete(_ context.Context, id string) error {
	return m.deleteErr
}

func (m *notesMockRepository) UpdatePath(_ context.Context, id, newPath string) error {
	return nil
}

func (m *notesMockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	var result []*domain.Project
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *notesMockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	var result []*domain.Project
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *notesMockRepository) UpdateState(_ context.Context, id string, state domain.ProjectState) error {
	for _, p := range m.projects {
		if p.ID == id {
			p.State = state
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *notesMockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}

var _ ports.ProjectRepository = (*notesMockRepository)(nil)

// TestModel_NotesKey_OpensEditor verifies pressing 'n' opens note editor (AC1).
func TestModel_NotesKey_OpensEditor(t *testing.T) {
	// Setup: Model with projects
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "existing"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	if updated.noteEditTarget == nil {
		t.Error("expected noteEditTarget to be set")
	}
}

// TestModel_NotesKey_IgnoredWhenNoProjects verifies 'n' is ignored without projects.
func TestModel_NotesKey_IgnoredWhenNoProjects(t *testing.T) {
	// Setup: Model WITHOUT projects
	repo := &notesMockRepository{}
	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = nil

	// Action
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated := newModel.(Model)

	// Assert
	if updated.isEditingNote {
		t.Error("expected isEditingNote to be false when no projects")
	}
}

// TestModel_NotesKey_IgnoredWhenAlreadyEditing verifies duplicate 'n' is ignored.
func TestModel_NotesKey_IgnoredWhenAlreadyEditing(t *testing.T) {
	// Setup: Model already in note editing mode
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.isEditingNote = true // Already editing

	// Action
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated := newModel.(Model)

	// Assert
	if !updated.isEditingNote {
		t.Error("expected isEditingNote to remain true")
	}
	if cmd != nil {
		t.Error("expected nil command when already editing")
	}
}

// TestModel_NotesEditor_EscCancels verifies Esc cancels editing (AC3).
func TestModel_NotesEditor_EscCancels(t *testing.T) {
	// Setup: Model in note editing mode
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "original"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.isEditingNote = true
	m.originalNote = "original"
	m.noteInput.SetValue("modified but not saved")

	// Action: Press Esc
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := newModel.(Model)

	// Assert: Editing cancelled
	if updated.isEditingNote {
		t.Error("expected isEditingNote to be false after Esc")
	}
}

// TestModel_NotesEditor_EnterSaves verifies Enter saves note (AC2).
func TestModel_NotesEditor_EnterSaves(t *testing.T) {
	// Setup: Model in note editing mode
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "original"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
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

// TestModel_NotesEditor_NavigationBlocked verifies navigation is blocked during editing (AC7).
func TestModel_NotesEditor_NavigationBlocked(t *testing.T) {
	// Setup: Model in note editing mode with multiple projects
	repo := &notesMockRepository{
		projects: []*domain.Project{
			{ID: "1", Path: "/test1", Name: "test1"},
			{ID: "2", Path: "/test2", Name: "test2"},
		},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.isEditingNote = true
	m.noteInput.SetValue("typing")

	initialSelection := m.projectList.Index()

	// Action: Press 'j' (navigation key)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := newModel.(Model)

	// Assert: Selection should NOT change (input captures key)
	if updated.projectList.Index() != initialSelection {
		t.Error("expected navigation to be blocked during note editing")
	}
	// isEditingNote should still be true
	if !updated.isEditingNote {
		t.Error("expected isEditingNote to remain true")
	}
}

// TestModel_NoteSavedMsg_UpdatesState verifies noteSavedMsg updates project state (AC2).
func TestModel_NoteSavedMsg_UpdatesState(t *testing.T) {
	// Setup: Model with project
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "old"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	// Action: Simulate noteSavedMsg
	msg := noteSavedMsg{projectID: "1", newNote: "new note content"}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Assert: Project notes updated
	if updated.projects[0].Notes != "new note content" {
		t.Errorf("expected project notes to be 'new note content', got %q", updated.projects[0].Notes)
	}
	// Feedback message set
	if updated.noteFeedback != "✓ Note saved" {
		t.Errorf("expected feedback '✓ Note saved', got %q", updated.noteFeedback)
	}
	// Timer command returned
	if cmd == nil {
		t.Error("expected timer command to be returned")
	}
}

// TestModel_NoteSaveErrorMsg_ShowsError verifies error feedback.
func TestModel_NoteSaveErrorMsg_ShowsError(t *testing.T) {
	// Setup
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40

	// Action: Simulate error
	msg := noteSaveErrorMsg{err: domain.ErrProjectNotFound}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Assert: Error feedback shown
	if updated.noteFeedback != "✗ Failed to save note" {
		t.Errorf("expected error feedback, got %q", updated.noteFeedback)
	}
	if cmd == nil {
		t.Error("expected timer command to be returned")
	}
}

// TestModel_ClearNoteFeedbackMsg_ClearsFeedback verifies feedback is cleared.
func TestModel_ClearNoteFeedbackMsg_ClearsFeedback(t *testing.T) {
	// Setup
	m := NewModel(nil)
	m.ready = true
	m.noteFeedback = "✓ Note saved"

	// Action
	msg := clearNoteFeedbackMsg{}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Assert
	if updated.noteFeedback != "" {
		t.Errorf("expected empty feedback, got %q", updated.noteFeedback)
	}
	if cmd != nil {
		t.Error("expected nil command")
	}
}

// TestModel_NoteEditor_RendersDialog verifies dialog rendering (AC1).
func TestModel_NoteEditor_RendersDialog(t *testing.T) {
	// Setup
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test-project", Notes: "existing note"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)

	// Open note editor
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated := newModel.(Model)

	// Render view
	view := updated.View()

	// Assert dialog content
	expectedStrings := []string{
		"Edit note for",
		"test-project",
		"[Enter] save",
		"[Esc] cancel",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(view, s) {
			t.Errorf("Note editor dialog missing: %q", s)
		}
	}
}

// TestModel_NoteEditor_EmptyNoteSavesEmpty verifies empty note clears (AC4).
func TestModel_NoteEditor_EmptyNoteSavesEmpty(t *testing.T) {
	// Setup
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "existing"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.repository = repo
	m.isEditingNote = true
	m.noteEditTarget = repo.projects[0]
	m.noteInput.SetValue("") // Empty note

	// Action: Save
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newModel.(Model)

	// Assert
	if updated.isEditingNote {
		t.Error("expected isEditingNote to be false")
	}
	if cmd == nil {
		t.Error("expected save command")
	}
}

// TestStartNoteEditing_InitializesTextInput verifies text input is properly initialized.
func TestStartNoteEditing_InitializesTextInput(t *testing.T) {
	// Setup
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "pre-existing note"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)

	// Action
	newModel, _ := m.startNoteEditing()
	updated := newModel.(Model)

	// Assert
	if updated.noteInput.Value() != "pre-existing note" {
		t.Errorf("expected noteInput to have pre-existing note, got %q", updated.noteInput.Value())
	}
	if updated.noteInput.CharLimit != 500 {
		t.Errorf("expected CharLimit of 500, got %d", updated.noteInput.CharLimit)
	}
}

// TestSaveNoteCmd_Success verifies saveNoteCmd executes and returns noteSavedMsg.
func TestSaveNoteCmd_Success(t *testing.T) {
	// Setup
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "old"}},
	}

	m := NewModel(repo)
	m.repository = repo

	// Create and execute the save command
	cmd := m.saveNoteCmd("1", "new note content")
	msg := cmd()

	// Assert: Should return noteSavedMsg
	savedMsg, ok := msg.(noteSavedMsg)
	if !ok {
		t.Fatalf("expected noteSavedMsg, got %T", msg)
	}
	if savedMsg.projectID != "1" {
		t.Errorf("expected projectID '1', got %q", savedMsg.projectID)
	}
	if savedMsg.newNote != "new note content" {
		t.Errorf("expected newNote 'new note content', got %q", savedMsg.newNote)
	}

	// Verify repository was updated
	if repo.projects[0].Notes != "new note content" {
		t.Errorf("expected repository project notes to be updated, got %q", repo.projects[0].Notes)
	}
}

// TestSaveNoteCmd_ProjectNotFound verifies saveNoteCmd returns error when project not found.
func TestSaveNoteCmd_ProjectNotFound(t *testing.T) {
	// Setup: Empty repository
	repo := &notesMockRepository{
		projects: []*domain.Project{},
	}

	m := NewModel(repo)
	m.repository = repo

	// Create and execute the save command for non-existent project
	cmd := m.saveNoteCmd("nonexistent", "note")
	msg := cmd()

	// Assert: Should return noteSaveErrorMsg
	errMsg, ok := msg.(noteSaveErrorMsg)
	if !ok {
		t.Fatalf("expected noteSaveErrorMsg, got %T", msg)
	}
	if errMsg.err == nil {
		t.Error("expected error to be set")
	}
}

// TestSaveNoteCmd_SaveFailure verifies saveNoteCmd returns error when save fails.
func TestSaveNoteCmd_SaveFailure(t *testing.T) {
	// Setup: Repository that fails on save
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "old"}},
		saveErr:  domain.ErrProjectNotFound, // Simulate save failure
	}

	m := NewModel(repo)
	m.repository = repo

	// Create and execute the save command
	cmd := m.saveNoteCmd("1", "new note")
	msg := cmd()

	// Assert: Should return noteSaveErrorMsg
	errMsg, ok := msg.(noteSaveErrorMsg)
	if !ok {
		t.Fatalf("expected noteSaveErrorMsg, got %T", msg)
	}
	if errMsg.err == nil {
		t.Error("expected error to be set")
	}
}

// TestSaveNoteCmd_EmptyNote verifies empty notes are saved correctly (AC4).
func TestSaveNoteCmd_EmptyNote(t *testing.T) {
	// Setup
	repo := &notesMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", Notes: "existing note"}},
	}

	m := NewModel(repo)
	m.repository = repo

	// Create and execute the save command with empty note
	cmd := m.saveNoteCmd("1", "")
	msg := cmd()

	// Assert: Should return noteSavedMsg with empty note
	savedMsg, ok := msg.(noteSavedMsg)
	if !ok {
		t.Fatalf("expected noteSavedMsg, got %T", msg)
	}
	if savedMsg.newNote != "" {
		t.Errorf("expected empty newNote, got %q", savedMsg.newNote)
	}

	// Verify repository was updated with empty note
	if repo.projects[0].Notes != "" {
		t.Errorf("expected repository project notes to be empty, got %q", repo.projects[0].Notes)
	}
}
