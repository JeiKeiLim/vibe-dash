package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// ============================================================================
// Story 3.6: Manual Refresh Tests
// ============================================================================

// refreshMockDetector implements ports.Detector for testing.
type refreshMockDetector struct {
	detectFunc func(ctx context.Context, path string) (*domain.DetectionResult, error)
}

func (m *refreshMockDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
	if m.detectFunc != nil {
		return m.detectFunc(ctx, path)
	}
	return &domain.DetectionResult{Method: "test", Stage: domain.StagePlan}, nil
}

func (m *refreshMockDetector) DetectMultiple(ctx context.Context, path string) ([]*domain.DetectionResult, error) {
	return nil, nil
}

// refreshMockRepository implements ports.ProjectRepository for testing.
type refreshMockRepository struct {
	projects  []*domain.Project
	saveErr   error
	findErr   error
	deleteErr error
}

func (m *refreshMockRepository) Save(ctx context.Context, project *domain.Project) error {
	return m.saveErr
}

func (m *refreshMockRepository) FindAll(ctx context.Context) ([]*domain.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.projects, nil
}

func (m *refreshMockRepository) FindByPath(ctx context.Context, path string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.Path == path {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *refreshMockRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *refreshMockRepository) Delete(ctx context.Context, id string) error {
	return m.deleteErr
}

func (m *refreshMockRepository) UpdatePath(ctx context.Context, id, newPath string) error {
	return nil
}

func (m *refreshMockRepository) FindActive(ctx context.Context) ([]*domain.Project, error) {
	var result []*domain.Project
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *refreshMockRepository) FindHibernated(ctx context.Context) ([]*domain.Project, error) {
	var result []*domain.Project
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *refreshMockRepository) UpdateState(ctx context.Context, id string, state domain.ProjectState) error {
	for _, p := range m.projects {
		if p.ID == id {
			p.State = state
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *refreshMockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (m *refreshMockRepository) ResetProject(_ context.Context, _ string) error {
	return nil
}

func (m *refreshMockRepository) ResetAll(_ context.Context) (int, error) {
	return 0, nil
}

var _ ports.ProjectRepository = (*refreshMockRepository)(nil)
var _ ports.Detector = (*refreshMockDetector)(nil)

// TestModel_RefreshKey_StartsRefresh verifies pressing 'r' initiates refresh (AC1).
func TestModel_RefreshKey_StartsRefresh(t *testing.T) {
	repo := &refreshMockRepository{
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
	m.projectList = components.NewProjectListModel(m.projects, m.width, m.height)
	m.statusBar = components.NewStatusBarModel(m.width)
	m.SetDetectionService(&refreshMockDetector{
		detectFunc: func(ctx context.Context, path string) (*domain.DetectionResult, error) {
			return &domain.DetectionResult{Method: "bmad", Stage: domain.StagePlan}, nil
		},
	})

	// Press 'r' to start refresh
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if !updated.isRefreshing {
		t.Error("expected isRefreshing to be true after pressing 'r'")
	}
	if cmd == nil {
		t.Error("expected command to be returned for async refresh")
	}
	if updated.refreshTotal != 2 {
		t.Errorf("expected refreshTotal to be 2, got %d", updated.refreshTotal)
	}
}

// TestModel_RefreshKey_DisabledWithoutDetectionService verifies refresh is disabled without detector (AC1).
func TestModel_RefreshKey_DisabledWithoutDetectionService(t *testing.T) {
	repo := &refreshMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(m.projects, m.width, m.height)
	// detectionService is nil

	// Press 'r' without detection service
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.isRefreshing {
		t.Error("expected isRefreshing to be false when no detection service")
	}
	if updated.refreshError == "" {
		t.Error("expected refreshError to be set when no detection service")
	}
	if updated.refreshError != "Detection service not available" {
		t.Errorf("expected error message 'Detection service not available', got: %s", updated.refreshError)
	}
}

// TestModel_RefreshKey_IgnoredWhenRefreshing verifies duplicate refresh is prevented (AC1).
func TestModel_RefreshKey_IgnoredWhenRefreshing(t *testing.T) {
	repo := &refreshMockRepository{
		projects: []*domain.Project{{ID: "1", Path: "/test", Name: "test"}},
	}

	m := NewModel(repo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = repo.projects
	m.isRefreshing = true // Already refreshing
	m.SetDetectionService(&refreshMockDetector{})

	// Press 'r' while already refreshing
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if cmd != nil {
		t.Error("expected nil command when already refreshing")
	}
	if !updated.isRefreshing {
		t.Error("expected isRefreshing to remain true")
	}
}

// TestModel_NavigationDuringRefresh verifies navigation works during refresh (AC5).
func TestModel_NavigationDuringRefresh(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", Path: "/test1", Name: "test1"},
		{ID: "2", Path: "/test2", Name: "test2"},
		{ID: "3", Path: "/test3", Name: "test3"},
	}

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.isRefreshing = true // Currently refreshing

	// Initial selection should be 0
	if m.projectList.Index() != 0 {
		t.Errorf("initial selection should be 0, got %d", m.projectList.Index())
	}

	// Press 'j' to navigate down during refresh
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Navigation should work (AC5: non-blocking refresh)
	if updated.projectList.Index() != 1 {
		t.Errorf("expected navigation to work during refresh, selection should be 1, got %d", updated.projectList.Index())
	}
	if !updated.isRefreshing {
		t.Error("isRefreshing should remain true during navigation")
	}
}

// TestModel_RefreshCompleteMsg_Success verifies successful refresh completion (AC1).
func TestModel_RefreshCompleteMsg_Success(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.isRefreshing = true
	m.statusBar = components.NewStatusBarModel(m.width)

	// Simulate refresh completion
	msg := refreshCompleteMsg{refreshedCount: 3, failedCount: 0, err: nil}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if updated.isRefreshing {
		t.Error("expected isRefreshing to be false after completion")
	}
	if updated.refreshError != "" {
		t.Errorf("expected no error, got: %s", updated.refreshError)
	}
	if cmd == nil {
		t.Error("expected batch command for reload and timer")
	}
}

// TestModel_RefreshCompleteMsg_PartialFailure verifies partial success handling (AC3).
func TestModel_RefreshCompleteMsg_PartialFailure(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.isRefreshing = true
	m.statusBar = components.NewStatusBarModel(m.width)

	// Simulate partial success (some failed)
	msg := refreshCompleteMsg{refreshedCount: 2, failedCount: 1, err: nil}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if updated.isRefreshing {
		t.Error("expected isRefreshing to be false after completion")
	}
	// Partial success still shows success (AC3: partial success is reported)
	if updated.refreshError != "" {
		t.Errorf("expected no error on partial success, got: %s", updated.refreshError)
	}
	if cmd == nil {
		t.Error("expected batch command for reload and timer")
	}
}

// TestModel_RefreshCompleteMsg_AllFailed verifies all-failed case (AC3).
func TestModel_RefreshCompleteMsg_AllFailed(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.isRefreshing = true
	m.statusBar = components.NewStatusBarModel(m.width)

	// Simulate all projects failed
	msg := refreshCompleteMsg{
		refreshedCount: 0,
		failedCount:    3,
		err:            errors.New("all projects failed to refresh"),
	}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if updated.isRefreshing {
		t.Error("expected isRefreshing to be false after failure")
	}
	if updated.refreshError != "all projects failed to refresh" {
		t.Errorf("expected error message, got: %s", updated.refreshError)
	}
	// No reload command on total failure
	if cmd != nil {
		t.Error("expected nil command on total failure")
	}
}

// TestModel_ClearRefreshMsgMsg verifies message clearing after timer (AC1).
func TestModel_ClearRefreshMsgMsg(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(m.width)
	m.statusBar.SetRefreshComplete("Refreshed 3 projects")

	// Verify message is set before clearing
	viewBefore := m.statusBar.View()
	if !strings.Contains(viewBefore, "Refreshed 3 projects") {
		t.Errorf("expected message before clearing, got: %s", viewBefore)
	}

	// Simulate timer expiry
	msg := clearRefreshMsgMsg{}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Message should be cleared
	viewAfter := updated.statusBar.View()
	if strings.Contains(viewAfter, "Refreshed") {
		t.Errorf("expected message to be cleared after timer, got: %s", viewAfter)
	}
	if cmd != nil {
		t.Error("expected nil command")
	}
}

// TestModel_SetDetectionService verifies setter works correctly.
func TestModel_SetDetectionService(t *testing.T) {
	m := NewModel(nil)

	if m.detectionService != nil {
		t.Error("expected detectionService to be nil initially")
	}

	detector := &refreshMockDetector{}
	m.SetDetectionService(detector)

	if m.detectionService == nil {
		t.Error("expected detectionService to be set")
	}
	if m.detectionService != detector {
		t.Error("expected detectionService to be the provided detector")
	}
}

// ============================================================================
// Story 7.3: Database Recovery TUI Tests
// ============================================================================

// TestModel_ProjectCorruptionMsg_SingleProject tests AC7: single project corruption warning
func TestModel_ProjectCorruptionMsg_SingleProject(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(m.width)

	// Send projectCorruptionMsg with single project
	msg := projectCorruptionMsg{
		projects: []string{"my-project"},
	}

	updatedModel, _ := m.Update(msg)
	updated := updatedModel.(Model)

	// Verify corruptedProjects is set
	if len(updated.corruptedProjects) != 1 {
		t.Errorf("expected 1 corrupted project, got %d", len(updated.corruptedProjects))
	}
	if updated.corruptedProjects[0] != "my-project" {
		t.Errorf("expected 'my-project', got %s", updated.corruptedProjects[0])
	}

	// Verify status bar warning contains expected text (AC7)
	view := updated.statusBar.View()
	if !strings.Contains(view, "corrupted") {
		t.Errorf("expected status bar to show corruption warning, got: %s", view)
	}
	if !strings.Contains(view, "my-project") {
		t.Errorf("expected status bar to mention project name, got: %s", view)
	}
	if !strings.Contains(view, "vibe reset") {
		t.Errorf("expected status bar to suggest 'vibe reset', got: %s", view)
	}
}

// TestModel_ProjectCorruptionMsg_MultipleProjects tests AC7: multiple projects corruption warning
func TestModel_ProjectCorruptionMsg_MultipleProjects(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(m.width)

	// Send projectCorruptionMsg with multiple projects
	msg := projectCorruptionMsg{
		projects: []string{"project1", "project2", "project3"},
	}

	updatedModel, _ := m.Update(msg)
	updated := updatedModel.(Model)

	// Verify corruptedProjects count
	if len(updated.corruptedProjects) != 3 {
		t.Errorf("expected 3 corrupted projects, got %d", len(updated.corruptedProjects))
	}

	// Verify status bar shows count and --all flag (AC7)
	view := updated.statusBar.View()
	if !strings.Contains(view, "3") {
		t.Errorf("expected status bar to show count '3', got: %s", view)
	}
	if !strings.Contains(view, "corrupted") {
		t.Errorf("expected status bar to show 'corrupted', got: %s", view)
	}
	if !strings.Contains(view, "vibe reset --all") {
		t.Errorf("expected status bar to suggest 'vibe reset --all', got: %s", view)
	}
}

// TestModel_ProjectCorruptionMsg_EmptyList tests no warning for empty list
func TestModel_ProjectCorruptionMsg_EmptyList(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(m.width)

	// Send projectCorruptionMsg with no projects
	msg := projectCorruptionMsg{
		projects: []string{},
	}

	updatedModel, _ := m.Update(msg)
	updated := updatedModel.(Model)

	// Verify corruptedProjects is empty
	if len(updated.corruptedProjects) != 0 {
		t.Errorf("expected 0 corrupted projects, got %d", len(updated.corruptedProjects))
	}
}
