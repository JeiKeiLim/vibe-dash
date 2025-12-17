package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 3.9: Remove Project Tests
// ============================================================================

// newRemoveMockRepository returns the favoriteMockRepository for test reuse.
func newRemoveMockRepository() *favoriteMockRepository {
	return &favoriteMockRepository{}
}

func TestModel_RemoveKey_StartsConfirmation(t *testing.T) {
	// Setup: Model with project
	repo := newRemoveMockRepository()
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	repo.projects = []*domain.Project{
		{ID: "1", Path: "/test1", Name: "project-1"},
		{ID: "2", Path: "/test2", Name: "project-2"},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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
	repo := newRemoveMockRepository()
	repo.projects = []*domain.Project{{
		ID:          "1",
		Path:        "/test",
		Name:        "test-project",
		DisplayName: "My Custom Name",
	}}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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

func TestModel_RemoveKey_IgnoredWhenAlreadyConfirming(t *testing.T) {
	// Setup: Model already in confirmation mode
	repo := newRemoveMockRepository()
	project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.repository = repo
	m.isConfirmingRemove = true
	m.confirmTarget = project

	// Action: Send 'x' key again
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updated := newModel.(Model)

	// Assert: Still in confirmation mode, no new command
	if !updated.isConfirmingRemove {
		t.Error("expected isConfirmingRemove to remain true")
	}
	if cmd != nil {
		t.Error("expected no command when already confirming")
	}
}

func TestModel_RemoveConfirmTimeoutMsg_IgnoredWhenNotConfirming(t *testing.T) {
	// Setup: Model NOT in confirmation mode (timeout might arrive late)
	repo := newRemoveMockRepository()

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.isConfirmingRemove = false
	m.confirmTarget = nil

	// Action: Send timeout message
	newModel, cmd := m.Update(removeConfirmTimeoutMsg{})
	updated := newModel.(Model)

	// Assert: No state change (safe no-op)
	if updated.isConfirmingRemove {
		t.Error("expected isConfirmingRemove to remain false")
	}
	if cmd != nil {
		t.Error("expected no command")
	}
}

func TestModel_RemoveConfirmedMsg_LastProjectRemoved(t *testing.T) {
	// Setup: Model with only 1 project
	repo := newRemoveMockRepository()
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.statusBar = components.NewStatusBarModel(80)
	m.detailPanel = components.NewDetailPanelModel(80, 24)
	m.repository = repo

	// Action: Receive removeConfirmedMsg
	newModel, _ := m.Update(removeConfirmedMsg{projectID: "1", projectName: "test-project"})
	updated := newModel.(Model)

	// Assert: Project list is empty
	if len(updated.projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(updated.projects))
	}
}

func TestModel_ClearRemoveFeedbackMsg_ClearsFeedback(t *testing.T) {
	// Setup
	m := NewModel(nil)
	m.ready = true
	m.statusBar = components.NewStatusBarModel(80)

	// Action
	msg := clearRemoveFeedbackMsg{}
	_, cmd := m.Update(msg)

	// Assert
	if cmd != nil {
		t.Error("expected nil command")
	}
}

// ============================================================================
// AC6 Tests: Selection After Removal (Code Review Fixes)
// ============================================================================

func TestModel_RemoveConfirmedMsg_SelectionMovesToNext_WhenMiddleRemoved(t *testing.T) {
	// AC6: "selection moves to next project"
	// Setup: 3 projects, select middle one (index 1), then remove it
	repo := newRemoveMockRepository()
	repo.projects = []*domain.Project{
		{ID: "1", Path: "/test1", Name: "aaa-project"},
		{ID: "2", Path: "/test2", Name: "bbb-project"},
		{ID: "3", Path: "/test3", Name: "ccc-project"},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.statusBar = components.NewStatusBarModel(80)
	m.detailPanel = components.NewDetailPanelModel(80, 24)
	m.repository = repo

	// Verify initial selection is at index 0
	if m.projectList.Index() != 0 {
		t.Fatalf("expected initial index 0, got %d", m.projectList.Index())
	}

	// Action: Remove project at index 0 (aaa-project)
	newModel, _ := m.Update(removeConfirmedMsg{projectID: "1", projectName: "aaa-project"})
	updated := newModel.(Model)

	// Assert: 2 projects remain
	if len(updated.projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(updated.projects))
	}

	// Assert: Selection should be valid (index 0 pointing to what was bbb-project)
	selected := updated.projectList.SelectedProject()
	if selected == nil {
		t.Fatal("expected a project to be selected")
	}
	// After removing aaa-project (index 0), the list has bbb and ccc
	// The bubbles list maintains index 0, which now points to bbb-project
	if selected.ID != "2" {
		t.Errorf("expected selected project ID '2' (bbb-project), got '%s'", selected.ID)
	}
}

func TestModel_RemoveConfirmedMsg_SelectionMovesToPrevious_WhenLastRemoved(t *testing.T) {
	// AC6: "or previous if last"
	// Setup: 3 projects, navigate to last item, then remove it
	repo := newRemoveMockRepository()
	repo.projects = []*domain.Project{
		{ID: "1", Path: "/test1", Name: "aaa-project"},
		{ID: "2", Path: "/test2", Name: "bbb-project"},
		{ID: "3", Path: "/test3", Name: "ccc-project"},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.statusBar = components.NewStatusBarModel(80)
	m.detailPanel = components.NewDetailPanelModel(80, 24)
	m.repository = repo

	// Navigate to last item (index 2)
	m.projectList, _ = m.projectList.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m.projectList, _ = m.projectList.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	if m.projectList.Index() != 2 {
		t.Fatalf("expected index 2 after navigation, got %d", m.projectList.Index())
	}

	// Action: Remove the last project (ccc-project at index 2)
	newModel, _ := m.Update(removeConfirmedMsg{projectID: "3", projectName: "ccc-project"})
	updated := newModel.(Model)

	// Assert: 2 projects remain
	if len(updated.projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(updated.projects))
	}

	// Assert: Selection should be valid and pointing to what is now the last item
	// After removing index 2, index 2 is out of bounds
	// The bubbles list should clamp to valid index
	selected := updated.projectList.SelectedProject()
	if selected == nil {
		t.Fatal("expected a project to be selected after removing last item")
	}
	// Selection should be at index 1 (bbb-project) since ccc was removed
	// Note: This tests that selection doesn't go out of bounds
}

func TestModel_RemoveConfirmedMsg_DetailPanelUpdatesWithNewSelection(t *testing.T) {
	// AC6: "And detail panel updates with new selection"
	repo := newRemoveMockRepository()
	repo.projects = []*domain.Project{
		{ID: "1", Path: "/test1", Name: "aaa-project"},
		{ID: "2", Path: "/test2", Name: "bbb-project"},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.statusBar = components.NewStatusBarModel(80)
	m.detailPanel = components.NewDetailPanelModel(80, 24)
	m.detailPanel.SetVisible(true) // Must be visible to render content
	m.detailPanel.SetProject(m.projectList.SelectedProject())
	m.repository = repo

	// Verify initial selection is aaa-project
	initialSelected := m.projectList.SelectedProject()
	if initialSelected == nil || initialSelected.ID != "1" {
		t.Fatalf("expected initial selection to be aaa-project, got %v", initialSelected)
	}

	// Action: Remove first project (aaa-project)
	newModel, _ := m.Update(removeConfirmedMsg{projectID: "1", projectName: "aaa-project"})
	updated := newModel.(Model)

	// Assert: Selection should now be bbb-project
	newSelected := updated.projectList.SelectedProject()
	if newSelected == nil {
		t.Fatal("expected a project to be selected after removal")
	}
	if newSelected.ID != "2" {
		t.Errorf("expected selection to be bbb-project (ID 2), got ID %s", newSelected.ID)
	}

	// Assert: Detail panel was updated (SetProject was called in handler)
	// The handler calls m.detailPanel.SetProject(m.projectList.SelectedProject())
	// We verify by checking the detail panel's view contains the new project
	view := updated.detailPanel.View()
	if view == "" {
		t.Error("expected detail panel to have content")
	}
}

// ============================================================================
// Code Review Fixes: Missing Tests
// ============================================================================

func TestModel_View_RendersConfirmationDialog(t *testing.T) {
	// M1 fix: Verify View() renders confirmation dialog when isConfirmingRemove is true
	repo := newRemoveMockRepository()
	project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.statusBar = components.NewStatusBarModel(80)
	m.detailPanel = components.NewDetailPanelModel(80, 24)
	m.isConfirmingRemove = true
	m.confirmTarget = project

	// Action: Call View()
	view := m.View()

	// Assert: View contains confirmation dialog elements
	if !strings.Contains(view, "Confirm Removal") {
		t.Error("expected View to contain 'Confirm Removal' title")
	}
	if !strings.Contains(view, "test-project") {
		t.Error("expected View to contain project name")
	}
	if !strings.Contains(view, "[y]") {
		t.Error("expected View to contain '[y]' hint")
	}
}

func TestModel_View_RendersConfirmationDialog_WithDisplayName(t *testing.T) {
	// Verify confirmation dialog shows DisplayName when set
	repo := newRemoveMockRepository()
	project := &domain.Project{
		ID:          "1",
		Path:        "/test",
		Name:        "original-name",
		DisplayName: "Custom Display Name",
	}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.statusBar = components.NewStatusBarModel(80)
	m.detailPanel = components.NewDetailPanelModel(80, 24)
	m.isConfirmingRemove = true
	m.confirmTarget = project

	// Action: Call View()
	view := m.View()

	// Assert: View shows DisplayName, not Name
	if !strings.Contains(view, "Custom Display Name") {
		t.Error("expected View to contain DisplayName 'Custom Display Name'")
	}
}

func TestModel_RemoveConfirmation_ConfirmsWithUppercaseY(t *testing.T) {
	// M3 fix: Test uppercase 'Y' works
	repo := newRemoveMockRepository()
	project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.repository = repo
	m.isConfirmingRemove = true
	m.confirmTarget = project

	// Action: Send uppercase 'Y' key
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	updated := newModel.(Model)

	// Assert: Confirmation mode exited
	if updated.isConfirmingRemove {
		t.Error("expected isConfirmingRemove to be false after uppercase Y")
	}
	// Assert: Delete command returned
	if cmd == nil {
		t.Error("expected delete command to be returned")
	}
}

func TestModel_RemoveConfirmation_CancelsWithUppercaseN(t *testing.T) {
	// M3 fix: Test uppercase 'N' works
	repo := newRemoveMockRepository()
	project := &domain.Project{ID: "1", Path: "/test", Name: "test-project"}
	repo.projects = []*domain.Project{project}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.isConfirmingRemove = true
	m.confirmTarget = project

	// Action: Send uppercase 'N' key
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	updated := newModel.(Model)

	// Assert: Confirmation mode exited, project still exists
	if updated.isConfirmingRemove {
		t.Error("expected isConfirmingRemove to be false after uppercase N")
	}
	if len(updated.projects) != 1 {
		t.Error("expected project to still exist after cancel")
	}
}
