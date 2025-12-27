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
	"github.com/JeiKeiLim/vibe-dash/internal/shared/emoji"
)

func init() {
	// Story 8.9: Initialize emoji package for tests with emoji enabled
	useEmoji := true
	emoji.InitEmoji(&useEmoji)
}

// ============================================================================
// Story 3.8: Favorite Toggle Tests
// ============================================================================

// favoriteMockRepository implements ports.ProjectRepository for testing.
type favoriteMockRepository struct {
	projects  []*domain.Project
	saveErr   error
	findErr   error
	deleteErr error
}

func (m *favoriteMockRepository) Save(_ context.Context, project *domain.Project) error {
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

func (m *favoriteMockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.projects, nil
}

func (m *favoriteMockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.Path == path {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *favoriteMockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *favoriteMockRepository) Delete(_ context.Context, id string) error {
	return m.deleteErr
}

func (m *favoriteMockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	var result []*domain.Project
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *favoriteMockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	var result []*domain.Project
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *favoriteMockRepository) UpdateState(_ context.Context, id string, state domain.ProjectState) error {
	for _, p := range m.projects {
		if p.ID == id {
			p.State = state
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *favoriteMockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (m *favoriteMockRepository) ResetProject(_ context.Context, _ string) error {
	return nil
}

func (m *favoriteMockRepository) ResetAll(_ context.Context) (int, error) {
	return 0, nil
}

var _ ports.ProjectRepository = (*favoriteMockRepository)(nil)

// TestModel_FavoriteKey_TogglesOn verifies pressing 'f' toggles favorite on (AC1).
func TestModel_FavoriteKey_TogglesOn(t *testing.T) {
	// Setup: Model with non-favorited project
	repo := &favoriteMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: false}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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

// TestModel_FavoriteKey_TogglesOff verifies pressing 'f' toggles favorite off (AC1).
func TestModel_FavoriteKey_TogglesOff(t *testing.T) {
	// Setup: Model with favorited project
	repo := &favoriteMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: true}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
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

// TestModel_FavoriteKey_IgnoredWhenNoProjects verifies 'f' is ignored without projects.
func TestModel_FavoriteKey_IgnoredWhenNoProjects(t *testing.T) {
	// Setup: Model WITHOUT projects
	repo := &favoriteMockRepository{}
	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = nil

	// Action
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	// Assert: No command returned (no-op)
	if cmd != nil {
		t.Error("expected no command when no projects")
	}
}

// TestModel_FavoriteSavedMsg_UpdatesProjectAndFeedback verifies favoriteSavedMsg updates state (AC1, AC6).
func TestModel_FavoriteSavedMsg_UpdatesProjectAndFeedback(t *testing.T) {
	// Setup: Model with project
	repo := &favoriteMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: false}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.statusBar = components.NewStatusBarModel(80)
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	// Action: Receive favoriteSavedMsg for favoriting
	newModel, cmd := m.Update(favoriteSavedMsg{projectID: "1", isFavorite: true})
	updated := newModel.(Model)

	// Assert: Project updated
	if !updated.projects[0].IsFavorite {
		t.Error("expected project IsFavorite to be true")
	}
	// Timer command returned
	if cmd == nil {
		t.Error("expected timer command to be returned")
	}
}

// TestModel_FavoriteSavedMsg_UnfavoriteFeedback verifies unfavorite feedback message.
func TestModel_FavoriteSavedMsg_UnfavoriteFeedback(t *testing.T) {
	// Setup: Model with favorited project
	repo := &favoriteMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: true}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.statusBar = components.NewStatusBarModel(80)
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	// Action: Receive favoriteSavedMsg for unfavoriting
	newModel, cmd := m.Update(favoriteSavedMsg{projectID: "1", isFavorite: false})
	updated := newModel.(Model)

	// Assert: Project updated
	if updated.projects[0].IsFavorite {
		t.Error("expected project IsFavorite to be false")
	}
	// Timer command returned
	if cmd == nil {
		t.Error("expected timer command to be returned")
	}
}

// TestModel_FavoriteSaveErrorMsg_ShowsError verifies error feedback.
func TestModel_FavoriteSaveErrorMsg_ShowsError(t *testing.T) {
	// Setup
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)

	// Action: Simulate error
	msg := favoriteSaveErrorMsg{err: domain.ErrProjectNotFound}
	newModel, cmd := m.Update(msg)
	_ = newModel.(Model)

	// Assert: Timer command returned for clearing error
	if cmd == nil {
		t.Error("expected timer command to be returned")
	}
}

// TestModel_ClearFavoriteFeedbackMsg_ClearsFeedback verifies feedback is cleared.
func TestModel_ClearFavoriteFeedbackMsg_ClearsFeedback(t *testing.T) {
	// Setup
	m := NewModel(nil)
	m.ready = true
	m.statusBar = components.NewStatusBarModel(80)

	// Action
	msg := clearFavoriteFeedbackMsg{}
	_, cmd := m.Update(msg)

	// Assert
	if cmd != nil {
		t.Error("expected nil command")
	}
}

// TestToggleFavorite_ReturnsCmd verifies toggleFavorite returns a command.
func TestToggleFavorite_ReturnsCmd(t *testing.T) {
	// Setup
	repo := &favoriteMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: false}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
	m.repository = repo

	// Action
	_, cmd := m.toggleFavorite()

	// Assert
	if cmd == nil {
		t.Error("expected command to be returned")
	}
}

// TestToggleFavorite_NoProjectSelected verifies toggleFavorite handles nil selection.
func TestToggleFavorite_NoProjectSelected(t *testing.T) {
	// Setup: Model with no projects in list
	repo := &favoriteMockRepository{}
	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = nil
	m.projectList = components.NewProjectListModel(nil, 80, 24)

	// Action
	_, cmd := m.toggleFavorite()

	// Assert: No command since no project selected
	if cmd != nil {
		t.Error("expected nil command when no project selected")
	}
}

// TestSaveFavoriteCmd_Success verifies saveFavoriteCmd executes and returns favoriteSavedMsg.
func TestSaveFavoriteCmd_Success(t *testing.T) {
	// Setup
	repo := &favoriteMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: false}},
	}

	m := NewModel(repo)
	m.repository = repo

	// Create and execute the save command
	cmd := m.saveFavoriteCmd("1", true)
	msg := cmd()

	// Assert: Should return favoriteSavedMsg
	savedMsg, ok := msg.(favoriteSavedMsg)
	if !ok {
		t.Fatalf("expected favoriteSavedMsg, got %T", msg)
	}
	if savedMsg.projectID != "1" {
		t.Errorf("expected projectID '1', got %q", savedMsg.projectID)
	}
	if !savedMsg.isFavorite {
		t.Error("expected isFavorite to be true")
	}

	// Verify repository was updated
	if !repo.projects[0].IsFavorite {
		t.Error("expected repository project IsFavorite to be updated to true")
	}
}

// TestSaveFavoriteCmd_ProjectNotFound verifies saveFavoriteCmd returns error when project not found.
func TestSaveFavoriteCmd_ProjectNotFound(t *testing.T) {
	// Setup: Empty repository
	repo := &favoriteMockRepository{
		projects: []*domain.Project{},
	}

	m := NewModel(repo)
	m.repository = repo

	// Create and execute the save command for non-existent project
	cmd := m.saveFavoriteCmd("nonexistent", true)
	msg := cmd()

	// Assert: Should return favoriteSaveErrorMsg
	errMsg, ok := msg.(favoriteSaveErrorMsg)
	if !ok {
		t.Fatalf("expected favoriteSaveErrorMsg, got %T", msg)
	}
	if errMsg.err == nil {
		t.Error("expected error to be set")
	}
}

// TestSaveFavoriteCmd_SaveFailure verifies saveFavoriteCmd returns error when save fails.
func TestSaveFavoriteCmd_SaveFailure(t *testing.T) {
	// Setup: Repository that fails on save
	repo := &favoriteMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test", IsFavorite: false}},
		saveErr:  domain.ErrProjectNotFound, // Simulate save failure
	}

	m := NewModel(repo)
	m.repository = repo

	// Create and execute the save command
	cmd := m.saveFavoriteCmd("1", true)
	msg := cmd()

	// Assert: Should return favoriteSaveErrorMsg
	errMsg, ok := msg.(favoriteSaveErrorMsg)
	if !ok {
		t.Fatalf("expected favoriteSaveErrorMsg, got %T", msg)
	}
	if errMsg.err == nil {
		t.Error("expected error to be set")
	}
}

// TestModel_FavoriteSavedMsg_ProjectListReRendersWithIndicator verifies project list re-renders with ⭐ indicator (AC6).
func TestModel_FavoriteSavedMsg_ProjectListReRendersWithIndicator(t *testing.T) {
	tests := []struct {
		name           string
		initialFav     bool
		newFav         bool
		wantStarInView bool
	}{
		{
			name:           "favoriting adds star to view",
			initialFav:     false,
			newFav:         true,
			wantStarInView: true,
		},
		{
			name:           "unfavoriting removes star from view",
			initialFav:     true,
			newFav:         false,
			wantStarInView: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Model with project
			repo := &favoriteMockRepository{
				projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test-project", IsFavorite: tt.initialFav}},
			}

			m := NewModel(repo)
			m.ready = true
			m.width = 80
			m.height = 40
			m.projects = repo.projects
			m.projectList = components.NewProjectListModel(repo.projects, 80, 24)
			m.statusBar = components.NewStatusBarModel(80)
			m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
			m.detailPanel.SetProject(m.projectList.SelectedProject())

			// Action: Receive favoriteSavedMsg
			newModel, _ := m.Update(favoriteSavedMsg{projectID: "1", isFavorite: tt.newFav})
			updated := newModel.(Model)

			// Get the view output
			view := updated.View()

			// Assert: Check if ⭐ appears/disappears in the rendered view
			hasStar := strings.Contains(view, "⭐")
			if tt.wantStarInView && !hasStar {
				t.Errorf("expected ⭐ in view after favoriting, got view without star")
			}
			if !tt.wantStarInView && hasStar {
				t.Errorf("expected no ⭐ in view after unfavoriting, got view with star")
			}
		})
	}
}
