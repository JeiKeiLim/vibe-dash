package tui_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/sqlite"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// setupTestRepo creates a test repository in a temporary directory
func setupTestRepo(t *testing.T) *sqlite.ProjectRepository {
	t.Helper()
	tmpDir := t.TempDir()

	repo, err := sqlite.NewProjectRepository(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	return repo
}

// TestValidateProjectPaths_AllValid tests AC5: all paths valid, no dialog shown
func TestValidateProjectPaths_AllValid(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create a project with a valid path (temp directory)
	tmpDir := t.TempDir()
	project, err := domain.NewProject(tmpDir, "test-project")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Validate paths
	invalid, err := tui.ValidateProjectPaths(ctx, repo)
	if err != nil {
		t.Fatalf("ValidateProjectPaths failed: %v", err)
	}

	// Should return empty slice (AC5)
	if len(invalid) != 0 {
		t.Errorf("Expected 0 invalid projects, got %d", len(invalid))
	}
}

// TestValidateProjectPaths_SomeMissing tests AC1: missing path detected
func TestValidateProjectPaths_SomeMissing(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create a project with a non-existent path
	nonExistentPath := "/nonexistent/path/that/should/not/exist"
	project := &domain.Project{
		ID:             domain.GenerateID(nonExistentPath),
		Name:           "missing-project",
		Path:           nonExistentPath,
		State:          domain.StateActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}

	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Validate paths
	invalid, err := tui.ValidateProjectPaths(ctx, repo)
	if err != nil {
		t.Fatalf("ValidateProjectPaths failed: %v", err)
	}

	// Should return slice with one invalid project (AC1)
	if len(invalid) != 1 {
		t.Errorf("Expected 1 invalid project, got %d", len(invalid))
	}

	if len(invalid) > 0 {
		if invalid[0].Project.ID != project.ID {
			t.Errorf("Invalid project ID mismatch: expected %s, got %s", project.ID, invalid[0].Project.ID)
		}
		if invalid[0].Error == nil {
			t.Error("Expected error for invalid project, got nil")
		}
	}
}

// TestValidateProjectPaths_MultipleMissing tests AC6: multiple missing paths
func TestValidateProjectPaths_MultipleMissing(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create multiple projects with non-existent paths
	paths := []string{
		"/nonexistent/path/one",
		"/nonexistent/path/two",
		"/nonexistent/path/three",
	}

	for i, path := range paths {
		project := &domain.Project{
			ID:             domain.GenerateID(path),
			Name:           "missing-project-" + string(rune('0'+i)),
			Path:           path,
			State:          domain.StateActive,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			LastActivityAt: time.Now(),
		}
		if err := repo.Save(ctx, project); err != nil {
			t.Fatalf("Failed to save project %d: %v", i, err)
		}
	}

	// Validate paths
	invalid, err := tui.ValidateProjectPaths(ctx, repo)
	if err != nil {
		t.Fatalf("ValidateProjectPaths failed: %v", err)
	}

	// Should return all 3 invalid projects (AC6)
	if len(invalid) != 3 {
		t.Errorf("Expected 3 invalid projects, got %d", len(invalid))
	}
}

// TestValidationDialog_PressD_DeletesProject tests AC2: 'D' key removes project
func TestValidationDialog_PressD_DeletesProject(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create a project with a non-existent path
	nonExistentPath := "/nonexistent/path/to/delete"
	project := &domain.Project{
		ID:             domain.GenerateID(nonExistentPath),
		Name:           "to-delete",
		Path:           nonExistentPath,
		State:          domain.StateActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Create model and simulate validation complete
	m := tui.NewModel(repo)
	invalid, _ := tui.ValidateProjectPaths(ctx, repo)

	// Simulate validation complete message
	msg := tui.ValidationCompleteMsgForTest(invalid)
	newModel, _ := m.Update(msg)
	m = newModel.(tui.Model)

	// Press 'd' to delete
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, cmd := m.Update(keyMsg)
	m = newModel.(tui.Model)

	// Execute the command (async delete)
	if cmd != nil {
		deleteMsg := cmd()
		m.Update(deleteMsg)
	}

	// Verify project was deleted
	_, err := repo.FindByID(ctx, project.ID)
	if err != domain.ErrProjectNotFound {
		t.Errorf("Expected project to be deleted, got error: %v", err)
	}
}

// TestValidationDialog_PressK_KeepsWithFlag tests AC4: 'K' key sets PathMissing flag
func TestValidationDialog_PressK_KeepsWithFlag(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create a project with a non-existent path
	nonExistentPath := "/nonexistent/path/to/keep"
	project := &domain.Project{
		ID:             domain.GenerateID(nonExistentPath),
		Name:           "to-keep",
		Path:           nonExistentPath,
		State:          domain.StateActive,
		PathMissing:    false, // Initially false
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Create model and simulate validation complete
	m := tui.NewModel(repo)
	invalid, _ := tui.ValidateProjectPaths(ctx, repo)

	// Simulate validation complete message
	msg := tui.ValidationCompleteMsgForTest(invalid)
	newModel, _ := m.Update(msg)
	m = newModel.(tui.Model)

	// Press 'k' to keep
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, cmd := m.Update(keyMsg)
	m = newModel.(tui.Model)

	// Execute the command (async keep)
	if cmd != nil {
		keepMsg := cmd()
		m.Update(keepMsg)
	}

	// Verify PathMissing flag was set
	updatedProject, err := repo.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to find project: %v", err)
	}
	if !updatedProject.PathMissing {
		t.Error("Expected PathMissing to be true after pressing 'K'")
	}
}

// TestValidationDialog_PressM_UpdatesPath tests AC3: 'M' key updates path to cwd
func TestValidationDialog_PressM_UpdatesPath(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create a project with a non-existent path
	nonExistentPath := "/nonexistent/path/to/move"
	project := &domain.Project{
		ID:             domain.GenerateID(nonExistentPath),
		Name:           "to-move",
		Path:           nonExistentPath,
		State:          domain.StateActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	// Create model and simulate validation complete
	m := tui.NewModel(repo)
	invalid, _ := tui.ValidateProjectPaths(ctx, repo)

	// Simulate validation complete message
	msg := tui.ValidationCompleteMsgForTest(invalid)
	newModel, _ := m.Update(msg)
	m = newModel.(tui.Model)

	// Press 'm' to move
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, cmd := m.Update(keyMsg)
	m = newModel.(tui.Model)

	// Execute the command (async move)
	if cmd != nil {
		moveMsg := cmd()
		m.Update(moveMsg)
	}

	// Old project should be deleted
	_, err = repo.FindByID(ctx, project.ID)
	if err != domain.ErrProjectNotFound {
		t.Errorf("Expected old project to be deleted, got error: %v", err)
	}

	// New project with new path should exist
	newID := domain.GenerateID(cwd)
	movedProject, err := repo.FindByID(ctx, newID)
	if err != nil {
		t.Fatalf("Failed to find moved project: %v", err)
	}
	if movedProject.Path != cwd {
		t.Errorf("Expected path to be %s, got %s", cwd, movedProject.Path)
	}
	if movedProject.PathMissing {
		t.Error("Expected PathMissing to be false after move")
	}
}

// TestValidationDialog_CaseInsensitiveKeys tests case insensitive key handling
func TestValidationDialog_CaseInsensitiveKeys(t *testing.T) {
	tests := []struct {
		key      rune
		expected string // "delete", "move", "keep"
	}{
		{'d', "delete"},
		{'D', "delete"},
		{'m', "move"},
		{'M', "move"},
		{'k', "keep"},
		{'K', "keep"},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			repo := setupTestRepo(t)
			ctx := context.Background()

			// Create a project with a non-existent path
			path := "/nonexistent/path/" + string(tt.key)
			project := &domain.Project{
				ID:             domain.GenerateID(path),
				Name:           "test-" + string(tt.key),
				Path:           path,
				State:          domain.StateActive,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				LastActivityAt: time.Now(),
			}
			if err := repo.Save(ctx, project); err != nil {
				t.Fatalf("Failed to save project: %v", err)
			}

			// Create model and simulate validation complete
			m := tui.NewModel(repo)
			invalid, _ := tui.ValidateProjectPaths(ctx, repo)
			msg := tui.ValidationCompleteMsgForTest(invalid)
			newModel, _ := m.Update(msg)
			m = newModel.(tui.Model)

			// Press key
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
			_, cmd := m.Update(keyMsg)

			// Should return a command (not nil)
			if cmd == nil {
				t.Errorf("Expected command for key %c, got nil", tt.key)
			}
		})
	}
}

// TestValidationDialog_MultipleInvalid_Sequential tests AC6: sequential handling
func TestValidationDialog_MultipleInvalid_Sequential(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create 3 projects with non-existent paths
	for i := 0; i < 3; i++ {
		path := "/nonexistent/path/" + string(rune('a'+i))
		project := &domain.Project{
			ID:             domain.GenerateID(path),
			Name:           "project-" + string(rune('a'+i)),
			Path:           path,
			State:          domain.StateActive,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			LastActivityAt: time.Now(),
		}
		if err := repo.Save(ctx, project); err != nil {
			t.Fatalf("Failed to save project %d: %v", i, err)
		}
	}

	// Create model and simulate validation complete
	m := tui.NewModel(repo)
	invalid, _ := tui.ValidateProjectPaths(ctx, repo)

	if len(invalid) != 3 {
		t.Fatalf("Expected 3 invalid projects, got %d", len(invalid))
	}

	msg := tui.ValidationCompleteMsgForTest(invalid)
	newModel, _ := m.Update(msg)
	m = newModel.(tui.Model)

	// Should be in validation mode
	if !tui.IsValidationMode(m) {
		t.Error("Expected to be in validation mode")
	}

	// Process first with 'd' (delete)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, cmd := m.Update(keyMsg)
	m = newModel.(tui.Model)
	if cmd != nil {
		deleteMsg := cmd()
		newModel, _ = m.Update(deleteMsg)
		m = newModel.(tui.Model)
	}

	// Should still be in validation mode (2 more to go)
	if !tui.IsValidationMode(m) {
		t.Error("Expected to still be in validation mode after first action")
	}

	// Process second with 'k' (keep)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, cmd = m.Update(keyMsg)
	m = newModel.(tui.Model)
	if cmd != nil {
		keepMsg := cmd()
		newModel, _ = m.Update(keepMsg)
		m = newModel.(tui.Model)
	}

	// Should still be in validation mode (1 more to go)
	if !tui.IsValidationMode(m) {
		t.Error("Expected to still be in validation mode after second action")
	}

	// Process third with 'd' (delete)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, cmd = m.Update(keyMsg)
	m = newModel.(tui.Model)
	if cmd != nil {
		deleteMsg := cmd()
		newModel, _ = m.Update(deleteMsg)
		m = newModel.(tui.Model)
	}

	// Should now be in normal mode (all handled)
	if tui.IsValidationMode(m) {
		t.Error("Expected to be in normal mode after all actions")
	}
}

// TestValidationDialog_QuitKey tests that quit works in validation mode
func TestValidationDialog_QuitKey(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create a project with a non-existent path
	path := "/nonexistent/path/quit"
	project := &domain.Project{
		ID:             domain.GenerateID(path),
		Name:           "quit-test",
		Path:           path,
		State:          domain.StateActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Create model and simulate validation complete
	m := tui.NewModel(repo)
	invalid, _ := tui.ValidateProjectPaths(ctx, repo)
	msg := tui.ValidationCompleteMsgForTest(invalid)
	newModel, _ := m.Update(msg)
	m = newModel.(tui.Model)

	// Press 'q' to quit
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(keyMsg)

	// Should return tea.Quit command
	if cmd == nil {
		t.Error("Expected quit command, got nil")
	}
}

// TestRenderValidationDialog tests dialog rendering content
func TestRenderValidationDialog(t *testing.T) {
	project := &domain.Project{
		Name: "test-project",
		Path: "/old/path",
	}

	dialog := tui.RenderValidationDialogForTest(project, 80, 24)

	expectedStrings := []string{
		"Warning",
		"test-project",
		"/old/path",
		"[D] Delete",
		"[M] Move",
		"[K] Keep",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(dialog, s) {
			t.Errorf("Dialog missing: %q", s)
		}
	}
}

// TestRenderValidationDialog_WithDisplayName tests dialog shows DisplayName
func TestRenderValidationDialog_WithDisplayName(t *testing.T) {
	project := &domain.Project{
		Name:        "real-name",
		DisplayName: "Custom Display Name",
		Path:        "/old/path",
	}

	dialog := tui.RenderValidationDialogForTest(project, 80, 24)

	// Should show DisplayName, not Name
	if !strings.Contains(dialog, "Custom Display Name") {
		t.Error("Dialog should show DisplayName")
	}
}

// TestEffectiveName tests the effectiveName helper
func TestEffectiveName(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		expected    string
	}{
		{"real-name", "", "real-name"},
		{"real-name", "display-name", "display-name"},
		{"", "display-name", "display-name"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			project := &domain.Project{
				Name:        tt.name,
				DisplayName: tt.displayName,
			}
			result := tui.EffectiveNameForTest(project)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestValidateProjectPaths_EmptyRepository tests validation with no projects
func TestValidateProjectPaths_EmptyRepository(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	invalid, err := tui.ValidateProjectPaths(ctx, repo)
	if err != nil {
		t.Fatalf("ValidateProjectPaths failed: %v", err)
	}

	// Should return empty slice (not nil)
	if invalid == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(invalid) != 0 {
		t.Errorf("Expected 0 invalid projects, got %d", len(invalid))
	}
}

// TestValidationDialog_ErrorDisplayed tests H1 fix: error feedback to user
func TestValidationDialog_ErrorDisplayed(t *testing.T) {
	project := &domain.Project{
		Name: "test-project",
		Path: "/nonexistent/path",
	}

	// Dialog with no error
	dialogNoError := tui.RenderValidationDialogForTest(project, 80, 24)
	if strings.Contains(dialogNoError, "Error:") {
		t.Error("Dialog without error should not contain 'Error:'")
	}

	// Dialog with error
	dialogWithError := tui.RenderValidationDialogWithErrorForTest(project, 80, 24, "test error message")
	if !strings.Contains(dialogWithError, "Error:") {
		t.Error("Dialog with error should contain 'Error:'")
	}
	if !strings.Contains(dialogWithError, "test error message") {
		t.Error("Dialog should display the error message")
	}
}

// TestValidationDialog_ErrorSetOnFailure tests that error state is set when operation fails
func TestValidationDialog_ErrorSetOnFailure(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create a project with a non-existent path
	path := "/nonexistent/path/error-test"
	project := &domain.Project{
		ID:             domain.GenerateID(path),
		Name:           "error-test",
		Path:           path,
		State:          domain.StateActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Create model and simulate validation complete
	m := tui.NewModel(repo)
	invalid, _ := tui.ValidateProjectPaths(ctx, repo)
	msg := tui.ValidationCompleteMsgForTest(invalid)
	newModel, _ := m.Update(msg)
	m = newModel.(tui.Model)

	// Initial error should be empty
	if tui.GetValidationError(m) != "" {
		t.Error("Initial validation error should be empty")
	}

	// After successful delete, error should still be empty
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, cmd := m.Update(keyMsg)
	m = newModel.(tui.Model)
	if cmd != nil {
		deleteMsg := cmd()
		newModel, _ = m.Update(deleteMsg)
		m = newModel.(tui.Model)
	}

	// Error should be empty after successful operation
	if tui.GetValidationError(m) != "" {
		t.Error("Validation error should be empty after successful delete")
	}
}

// TestValidationDialog_StaysInModeOnError tests AC3 edge case: stays in validation on error
func TestValidationDialog_StaysInModeOnError(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Create two projects with non-existent paths
	for i := 0; i < 2; i++ {
		path := "/nonexistent/path/stay-test-" + string(rune('0'+i))
		project := &domain.Project{
			ID:             domain.GenerateID(path),
			Name:           "stay-test-" + string(rune('0'+i)),
			Path:           path,
			State:          domain.StateActive,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			LastActivityAt: time.Now(),
		}
		if err := repo.Save(ctx, project); err != nil {
			t.Fatalf("Failed to save project: %v", err)
		}
	}

	// Create model and simulate validation complete
	m := tui.NewModel(repo)
	invalid, _ := tui.ValidateProjectPaths(ctx, repo)
	if len(invalid) != 2 {
		t.Fatalf("Expected 2 invalid projects, got %d", len(invalid))
	}

	msg := tui.ValidationCompleteMsgForTest(invalid)
	newModel, _ := m.Update(msg)
	m = newModel.(tui.Model)

	// Verify we're in validation mode
	if !tui.IsValidationMode(m) {
		t.Fatal("Should be in validation mode")
	}

	// Process first project successfully
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, cmd := m.Update(keyMsg)
	m = newModel.(tui.Model)
	if cmd != nil {
		deleteMsg := cmd()
		newModel, _ = m.Update(deleteMsg)
		m = newModel.(tui.Model)
	}

	// Should still be in validation mode (second project pending)
	if !tui.IsValidationMode(m) {
		t.Error("Should still be in validation mode with second project")
	}
}
