package components

import (
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func createTestProject(name string, displayName string) *domain.Project {
	return &domain.Project{
		ID:             domain.GenerateID("/test/" + name),
		Name:           name,
		DisplayName:    displayName,
		Path:           "/test/" + name,
		CurrentStage:   domain.StageImplement,
		LastActivityAt: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func TestNewProjectListModel_EmptyList(t *testing.T) {
	projects := []*domain.Project{}
	model := NewProjectListModel(projects, 80, 24)

	if model.HasProjects() {
		t.Error("HasProjects() should return false for empty list")
	}

	if model.Len() != 0 {
		t.Errorf("Len() = %d, want 0", model.Len())
	}

	if model.SelectedProject() != nil {
		t.Error("SelectedProject() should return nil for empty list")
	}
}

func TestNewProjectListModel_SingleProject(t *testing.T) {
	projects := []*domain.Project{
		createTestProject("test-project", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	if !model.HasProjects() {
		t.Error("HasProjects() should return true")
	}

	if model.Len() != 1 {
		t.Errorf("Len() = %d, want 1", model.Len())
	}

	selected := model.SelectedProject()
	if selected == nil {
		t.Error("SelectedProject() should return first project")
	}
	if selected != nil && selected.Name != "test-project" {
		t.Errorf("SelectedProject().Name = %q, want %q", selected.Name, "test-project")
	}
}

func TestNewProjectListModel_Sorting(t *testing.T) {
	// Create projects in non-alphabetical order
	projects := []*domain.Project{
		createTestProject("zebra", ""),
		createTestProject("alpha", ""),
		createTestProject("middle", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	// First selected should be "alpha" after sorting
	selected := model.SelectedProject()
	if selected == nil {
		t.Fatal("SelectedProject() should not be nil")
	}
	if selected.Name != "alpha" {
		t.Errorf("First project after sorting should be 'alpha', got %q", selected.Name)
	}
}

func TestNewProjectListModel_SortingWithDisplayName(t *testing.T) {
	// Create projects where display name affects sort order
	projects := []*domain.Project{
		createTestProject("zulu", "alpha-display"),  // Display name starts with 'a'
		createTestProject("alpha", ""),              // Name starts with 'a'
		createTestProject("bravo", "zebra-display"), // Display name starts with 'z'
	}
	model := NewProjectListModel(projects, 80, 24)

	// "alpha" should come first (name starts with 'a', no display name)
	// "zulu" should come second (display name "alpha-display" starts with 'a')
	// "bravo" should come last (display name "zebra-display" starts with 'z')
	selected := model.SelectedProject()
	if selected == nil {
		t.Fatal("SelectedProject() should not be nil")
	}
	if selected.Name != "alpha" {
		t.Errorf("First project should be 'alpha', got %q", selected.Name)
	}
}

func TestNewProjectListModel_SelectsFirstItem(t *testing.T) {
	projects := []*domain.Project{
		createTestProject("project-a", ""),
		createTestProject("project-b", ""),
		createTestProject("project-c", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	if model.Index() != 0 {
		t.Errorf("Index() = %d, want 0", model.Index())
	}
}

func TestProjectListModel_SetProjects(t *testing.T) {
	// Start with empty list
	model := NewProjectListModel([]*domain.Project{}, 80, 24)

	if model.HasProjects() {
		t.Error("Initial list should be empty")
	}

	// Add projects
	projects := []*domain.Project{
		createTestProject("new-project", ""),
	}
	model.SetProjects(projects)

	if !model.HasProjects() {
		t.Error("HasProjects() should return true after SetProjects")
	}

	if model.Len() != 1 {
		t.Errorf("Len() = %d, want 1", model.Len())
	}
}

func TestProjectListModel_SetSize(t *testing.T) {
	projects := []*domain.Project{
		createTestProject("project", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	// Change size
	model.SetSize(120, 40)

	// View should still work (basic sanity check)
	view := model.View()
	if view == "" {
		t.Error("View() should return non-empty string")
	}
}

func TestProjectListModel_View(t *testing.T) {
	projects := []*domain.Project{
		createTestProject("test-project", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	view := model.View()
	if view == "" {
		t.Error("View() should return non-empty string for non-empty list")
	}
}

func TestProjectListModel_Update(t *testing.T) {
	projects := []*domain.Project{
		createTestProject("project", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	// Update should not panic with nil message
	newModel, cmd := model.Update(nil)
	_ = newModel
	_ = cmd
	// Just verify it doesn't panic
}

func TestProjectListModel_Index(t *testing.T) {
	projects := []*domain.Project{
		createTestProject("a-project", ""),
		createTestProject("b-project", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	if model.Index() != 0 {
		t.Errorf("Index() = %d, want 0", model.Index())
	}
}

func TestProjectListModel_Len(t *testing.T) {
	tests := []struct {
		name     string
		projects []*domain.Project
		want     int
	}{
		{
			name:     "empty list",
			projects: []*domain.Project{},
			want:     0,
		},
		{
			name: "single project",
			projects: []*domain.Project{
				createTestProject("project", ""),
			},
			want: 1,
		},
		{
			name: "multiple projects",
			projects: []*domain.Project{
				createTestProject("a", ""),
				createTestProject("b", ""),
				createTestProject("c", ""),
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewProjectListModel(tt.projects, 80, 24)
			if got := model.Len(); got != tt.want {
				t.Errorf("Len() = %d, want %d", got, tt.want)
			}
		})
	}
}
