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

// Story 8.5: Tests for favorites sorting and selection preservation

func createTestProjectWithFavorite(name string, displayName string, isFavorite bool) *domain.Project {
	p := createTestProject(name, displayName)
	p.IsFavorite = isFavorite
	return p
}

func TestNewProjectListModel_FavoritesSortFirst(t *testing.T) {
	projects := []*domain.Project{
		createTestProjectWithFavorite("alpha", "", false),
		createTestProjectWithFavorite("beta", "", true),
		createTestProjectWithFavorite("gamma", "", false),
	}
	model := NewProjectListModel(projects, 80, 24)

	// Favorite "beta" should be first
	selected := model.SelectedProject()
	if selected == nil {
		t.Fatal("SelectedProject() should not be nil")
	}
	if !selected.IsFavorite {
		t.Error("First project should be a favorite")
	}
	if selected.Name != "beta" {
		t.Errorf("First project should be 'beta' (favorite), got %q", selected.Name)
	}
}

func TestNewProjectListModel_FavoritesSortedAlphabetically(t *testing.T) {
	projects := []*domain.Project{
		createTestProjectWithFavorite("zebra", "", true),
		createTestProjectWithFavorite("alpha", "", true),
		createTestProjectWithFavorite("middle", "", true),
	}
	model := NewProjectListModel(projects, 80, 24)

	// All favorites should be sorted alphabetically
	projectList := model.Projects()
	expected := []string{"alpha", "middle", "zebra"}
	for i, name := range expected {
		if projectList[i].Name != name {
			t.Errorf("Projects()[%d].Name = %q, want %q", i, projectList[i].Name, name)
		}
	}
}

func TestNewProjectListModel_NonFavoritesSortedAfterFavorites(t *testing.T) {
	projects := []*domain.Project{
		createTestProjectWithFavorite("zebra-nonfav", "", false),
		createTestProjectWithFavorite("alpha-fav", "", true),
		createTestProjectWithFavorite("beta-nonfav", "", false),
		createTestProjectWithFavorite("gamma-fav", "", true),
	}
	model := NewProjectListModel(projects, 80, 24)

	// Expected order: alpha-fav, gamma-fav, beta-nonfav, zebra-nonfav
	projectList := model.Projects()
	expected := []string{"alpha-fav", "gamma-fav", "beta-nonfav", "zebra-nonfav"}
	for i, name := range expected {
		if projectList[i].Name != name {
			t.Errorf("Projects()[%d].Name = %q, want %q", i, projectList[i].Name, name)
		}
	}
}

func TestProjectListModel_SelectByIndex(t *testing.T) {
	projects := []*domain.Project{
		createTestProject("a-project", ""),
		createTestProject("b-project", ""),
		createTestProject("c-project", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	// Initial selection is 0
	if model.Index() != 0 {
		t.Errorf("Initial Index() = %d, want 0", model.Index())
	}

	// Select index 2
	model.SelectByIndex(2)
	if model.Index() != 2 {
		t.Errorf("After SelectByIndex(2), Index() = %d, want 2", model.Index())
	}

	selected := model.SelectedProject()
	if selected == nil || selected.Name != "c-project" {
		t.Errorf("SelectedProject().Name = %q, want 'c-project'", selected.Name)
	}
}

func TestProjectListModel_SelectByIndex_OutOfBounds(t *testing.T) {
	projects := []*domain.Project{
		createTestProject("a-project", ""),
		createTestProject("b-project", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	// Initial selection is 0
	model.SelectByIndex(1)
	if model.Index() != 1 {
		t.Errorf("After SelectByIndex(1), Index() = %d, want 1", model.Index())
	}

	// Select out of bounds (negative) - should not change selection
	model.SelectByIndex(-1)
	if model.Index() != 1 {
		t.Errorf("After SelectByIndex(-1), Index() = %d, want 1 (unchanged)", model.Index())
	}

	// Select out of bounds (too high) - should not change selection
	model.SelectByIndex(10)
	if model.Index() != 1 {
		t.Errorf("After SelectByIndex(10), Index() = %d, want 1 (unchanged)", model.Index())
	}
}

func TestProjectListModel_Projects(t *testing.T) {
	projects := []*domain.Project{
		createTestProject("zebra", ""),
		createTestProject("alpha", ""),
	}
	model := NewProjectListModel(projects, 80, 24)

	// Projects() should return sorted list
	projectList := model.Projects()
	if len(projectList) != 2 {
		t.Errorf("len(Projects()) = %d, want 2", len(projectList))
	}
	// Should be sorted alphabetically
	if projectList[0].Name != "alpha" {
		t.Errorf("Projects()[0].Name = %q, want 'alpha'", projectList[0].Name)
	}
	if projectList[1].Name != "zebra" {
		t.Errorf("Projects()[1].Name = %q, want 'zebra'", projectList[1].Name)
	}
}

func TestProjectListModel_SetProjects_ResortsFavorites(t *testing.T) {
	// Start with non-favorite projects
	projects := []*domain.Project{
		createTestProjectWithFavorite("alpha", "", false),
		createTestProjectWithFavorite("beta", "", false),
	}
	model := NewProjectListModel(projects, 80, 24)

	// Now update one to be favorite
	projects[1].IsFavorite = true
	model.SetProjects(projects)

	// Beta (now favorite) should be first
	selected := model.SelectedProject()
	if selected == nil || selected.Name != "beta" {
		if selected != nil {
			t.Errorf("After making 'beta' favorite, first project should be 'beta', got %q", selected.Name)
		} else {
			t.Error("SelectedProject() returned nil")
		}
	}
}

// TestProjectListModel_Story85_SelectionPreservedAfterFavoriteToggle tests AC5:
// Selection should be preserved (by project ID) after toggling favorite status
// and re-sorting the list. This simulates the full favoriteSavedMsg flow in model.go.
func TestProjectListModel_Story85_SelectionPreservedAfterFavoriteToggle(t *testing.T) {
	// Create list with mixed favorites/non-favorites
	// Order after sort: gamma-fav, alpha (non-fav), beta (non-fav), delta (non-fav)
	projects := []*domain.Project{
		createTestProjectWithFavorite("alpha", "", false),
		createTestProjectWithFavorite("beta", "", false),
		createTestProjectWithFavorite("gamma-fav", "", true),
		createTestProjectWithFavorite("delta", "", false),
	}
	model := NewProjectListModel(projects, 80, 24)

	// Initial order: gamma-fav, alpha, beta, delta
	// Select "beta" at index 2
	model.SelectByIndex(2)
	selected := model.SelectedProject()
	if selected == nil || selected.Name != "beta" {
		t.Fatalf("Initial selection should be 'beta', got %v", selected)
	}
	selectedID := selected.ID

	// Simulate favoriteSavedMsg: toggle beta to favorite
	// Find project in slice and update
	for _, p := range projects {
		if p.ID == selectedID {
			p.IsFavorite = true
			break
		}
	}

	// Re-sort via SetProjects (this is what model.go does)
	model.SetProjects(projects)

	// After re-sort, order is: beta (now fav), gamma-fav, alpha, delta
	// Beta is now at index 0

	// Restore selection by ID (simulate model.go pattern)
	for i, p := range model.Projects() {
		if p.ID == selectedID {
			model.SelectByIndex(i)
			break
		}
	}

	// Verify: Same project (beta) is still selected, now at different position
	finalSelected := model.SelectedProject()
	if finalSelected == nil {
		t.Fatal("Selection should not be nil after restore")
	}
	if finalSelected.ID != selectedID {
		t.Errorf("Selection should be preserved by ID: got %q, want ID %q", finalSelected.Name, selectedID)
	}
	if finalSelected.Name != "beta" {
		t.Errorf("Selected project should still be 'beta', got %q", finalSelected.Name)
	}
	// Verify position changed (was at 2, now at 0)
	if model.Index() != 0 {
		t.Errorf("Beta should now be at index 0 (first favorite), got index %d", model.Index())
	}
}
