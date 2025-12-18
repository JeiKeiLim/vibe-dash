//go:build integration

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// mockPathLookupForIntegration implements ports.ProjectPathLookup for testing
type mockPathLookupForIntegration struct {
	paths map[string]string // canonical path -> directory name
}

// GetDirForPath returns existing directory name for canonical path.
// Returns empty string if path not previously registered.
func (m *mockPathLookupForIntegration) GetDirForPath(canonicalPath string) string {
	if m.paths == nil {
		return ""
	}
	return m.paths[canonicalPath]
}

func (m *mockPathLookupForIntegration) RegisterPath(canonical, dirName string) {
	if m.paths == nil {
		m.paths = make(map[string]string)
	}
	m.paths[canonical] = dirName
}

// Subtask 8.1-8.2: Test full lifecycle with real directories
func TestIntegration_ProjectRepository_FullLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create a "real" project directory
	projectPath := filepath.Join(tempDir, "projects", "my-app")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	// Use DirectoryManager to create project directory with marker
	lookup := &mockPathLookupForIntegration{}
	dm := filesystem.NewDirectoryManager(basePath, lookup)
	ctx := context.Background()

	projectDir, err := dm.EnsureProjectDir(ctx, projectPath)
	if err != nil {
		t.Fatalf("failed to ensure project dir: %v", err)
	}

	// Register the path for lookup
	lookup.RegisterPath(projectPath, filepath.Base(projectDir))

	// Create repository for this project
	repo, err := NewProjectRepository(projectDir)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	// Verify state.db was created
	dbPath := filepath.Join(projectDir, "state.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("state.db not created at %s", dbPath)
	}

	// Create and save a project
	now := time.Now()
	project := &domain.Project{
		ID:             "test-project-id",
		Name:           "my-app",
		Path:           projectPath,
		CurrentStage:   domain.StageUnknown,
		State:          domain.StateActive,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Find the project
	found, err := repo.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("failed to find project: %v", err)
	}

	if found.Name != project.Name {
		t.Errorf("name mismatch: got %q, want %q", found.Name, project.Name)
	}

	// Update state
	if err := repo.UpdateState(ctx, project.ID, domain.StateHibernated); err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	// Verify state updated
	found, _ = repo.FindByID(ctx, project.ID)
	if found.State != domain.StateHibernated {
		t.Errorf("state not updated: got %v, want %v", found.State, domain.StateHibernated)
	}

	// Delete the project
	if err := repo.Delete(ctx, project.ID); err != nil {
		t.Fatalf("failed to delete project: %v", err)
	}

	// Verify deleted
	_, err = repo.FindByID(ctx, project.ID)
	if err != domain.ErrProjectNotFound {
		t.Errorf("expected ErrProjectNotFound after delete, got %v", err)
	}
}

// Subtask 8.3: Test two projects have isolated databases
func TestIntegration_ProjectRepository_Isolation(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create two project directories
	project1Path := filepath.Join(tempDir, "projects", "project-a")
	project2Path := filepath.Join(tempDir, "projects", "project-b")

	if err := os.MkdirAll(project1Path, 0755); err != nil {
		t.Fatalf("failed to create project1 dir: %v", err)
	}
	if err := os.MkdirAll(project2Path, 0755); err != nil {
		t.Fatalf("failed to create project2 dir: %v", err)
	}

	lookup := &mockPathLookupForIntegration{}
	dm := filesystem.NewDirectoryManager(basePath, lookup)
	ctx := context.Background()

	// Setup both project directories
	projectDir1, err := dm.EnsureProjectDir(ctx, project1Path)
	if err != nil {
		t.Fatalf("failed to ensure project1 dir: %v", err)
	}
	lookup.RegisterPath(project1Path, filepath.Base(projectDir1))

	projectDir2, err := dm.EnsureProjectDir(ctx, project2Path)
	if err != nil {
		t.Fatalf("failed to ensure project2 dir: %v", err)
	}
	lookup.RegisterPath(project2Path, filepath.Base(projectDir2))

	// Create repositories
	repo1, err := NewProjectRepository(projectDir1)
	if err != nil {
		t.Fatalf("failed to create repo1: %v", err)
	}

	repo2, err := NewProjectRepository(projectDir2)
	if err != nil {
		t.Fatalf("failed to create repo2: %v", err)
	}

	// Verify separate database files
	db1Path := filepath.Join(projectDir1, "state.db")
	db2Path := filepath.Join(projectDir2, "state.db")

	if db1Path == db2Path {
		t.Fatalf("database paths should be different")
	}

	// Create projects
	now := time.Now()
	project1 := &domain.Project{
		ID:             "project-1-id",
		Name:           "project-a",
		Path:           project1Path,
		Notes:          "This is project 1",
		CurrentStage:   domain.StageUnknown,
		State:          domain.StateActive,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	project2 := &domain.Project{
		ID:             "project-2-id",
		Name:           "project-b",
		Path:           project2Path,
		Notes:          "This is project 2",
		CurrentStage:   domain.StagePlan,
		State:          domain.StateHibernated,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Save to respective repos
	if err := repo1.Save(ctx, project1); err != nil {
		t.Fatalf("failed to save project1: %v", err)
	}
	if err := repo2.Save(ctx, project2); err != nil {
		t.Fatalf("failed to save project2: %v", err)
	}

	// Verify isolation: repo1 only has project1
	all1, err := repo1.FindAll(ctx)
	if err != nil {
		t.Fatalf("repo1 FindAll failed: %v", err)
	}
	if len(all1) != 1 {
		t.Errorf("repo1 should have 1 project, got %d", len(all1))
	}
	if len(all1) > 0 && all1[0].ID != project1.ID {
		t.Errorf("repo1 has wrong project: got %q, want %q", all1[0].ID, project1.ID)
	}

	// Verify isolation: repo2 only has project2
	all2, err := repo2.FindAll(ctx)
	if err != nil {
		t.Fatalf("repo2 FindAll failed: %v", err)
	}
	if len(all2) != 1 {
		t.Errorf("repo2 should have 1 project, got %d", len(all2))
	}
	if len(all2) > 0 && all2[0].ID != project2.ID {
		t.Errorf("repo2 has wrong project: got %q, want %q", all2[0].ID, project2.ID)
	}

	// Modify project1 - should not affect project2
	if err := repo1.UpdateState(ctx, project1.ID, domain.StateHibernated); err != nil {
		t.Fatalf("failed to update project1 state: %v", err)
	}

	// Verify project2 is unchanged
	found2, err := repo2.FindByID(ctx, project2.ID)
	if err != nil {
		t.Fatalf("failed to find project2: %v", err)
	}
	if found2.Notes != project2.Notes {
		t.Errorf("project2 notes changed unexpectedly")
	}

	// Delete project1 - should not affect project2
	if err := repo1.Delete(ctx, project1.ID); err != nil {
		t.Fatalf("failed to delete project1: %v", err)
	}

	// Verify project2 still exists
	found2, err = repo2.FindByID(ctx, project2.ID)
	if err != nil {
		t.Fatalf("project2 should still exist: %v", err)
	}
	if found2.Name != project2.Name {
		t.Errorf("project2 changed unexpectedly")
	}
}

// Subtask 8.4: Test integration with real DirectoryManager.EnsureProjectDir()
func TestIntegration_ProjectRepository_WithDirectoryManager(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	projectPath := filepath.Join(tempDir, "projects", "integration-test")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	lookup := &mockPathLookupForIntegration{}
	dm := filesystem.NewDirectoryManager(basePath, lookup)
	ctx := context.Background()

	// Use DirectoryManager to set up project directory
	projectDir, err := dm.EnsureProjectDir(ctx, projectPath)
	if err != nil {
		t.Fatalf("EnsureProjectDir failed: %v", err)
	}

	// Verify .project-path marker exists
	markerPath := filepath.Join(projectDir, ".project-path")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Fatalf(".project-path marker not found")
	}

	// Create repository - should succeed because marker exists
	repo, err := NewProjectRepository(projectDir)
	if err != nil {
		t.Fatalf("NewProjectRepository failed: %v", err)
	}

	// Verify repository is functional
	now := time.Now()
	project := &domain.Project{
		ID:             "dm-integration-test",
		Name:           "integration-test",
		Path:           projectPath,
		CurrentStage:   domain.StageUnknown,
		State:          domain.StateActive,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	found, err := repo.FindByPath(ctx, projectPath)
	if err != nil {
		t.Fatalf("failed to find by path: %v", err)
	}

	if found.ID != project.ID {
		t.Errorf("ID mismatch: got %q, want %q", found.ID, project.ID)
	}
}

// Test that repository fails without DirectoryManager setup
func TestIntegration_ProjectRepository_FailsWithoutMarker(t *testing.T) {
	tempDir := t.TempDir()

	// Create directory but NOT the .project-path marker
	projectDir := filepath.Join(tempDir, "project-without-marker")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	// This should fail because .project-path marker is missing
	_, err := NewProjectRepository(projectDir)
	if err == nil {
		t.Fatal("expected error without .project-path marker")
	}

	if !strings.Contains(err.Error(), "missing .project-path") {
		t.Errorf("error should mention missing marker: %v", err)
	}
}
