//go:build integration

package persistence

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/sqlite"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Integration tests for RepositoryCoordinator
// Run with: go test -tags=integration ./internal/adapters/persistence/...

// TestIntegration_FullLifecycle tests AC1, AC2, AC3, AC4, AC12, AC13
func TestIntegration_FullLifecycle(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup 3 project directories
	proj1Dir := setupIntegrationProjectDir(t, basePath, "project-1")
	proj2Dir := setupIntegrationProjectDir(t, basePath, "project-2")
	proj3Dir := setupIntegrationProjectDir(t, basePath, "project-3")

	// Pre-populate databases
	repo1, _ := sqlite.NewProjectRepository(proj1Dir)
	repo2, _ := sqlite.NewProjectRepository(proj2Dir)
	repo3, _ := sqlite.NewProjectRepository(proj3Dir)

	project1 := createIntegrationTestProject("/integration/project1")
	project2 := createIntegrationTestProject("/integration/project2")
	project3 := createIntegrationTestProject("/integration/project3")

	repo1.Save(ctx, project1)
	repo2.Save(ctx, project2)
	repo3.Save(ctx, project3)

	// Setup config
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("project-1", "/integration/project1", "Project One", false)
	cfg.SetProjectEntry("project-2", "/integration/project2", "Project Two", true)
	cfg.SetProjectEntry("project-3", "/integration/project3", "Project Three", false)

	mockLoader := &integrationMockConfigLoader{
		config: cfg,
	}

	coord := NewRepositoryCoordinator(mockLoader, &integrationMockDirectoryManager{basePath: basePath}, basePath)

	// Test FindAll returns 3 projects
	projects, err := coord.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}
	if len(projects) != 3 {
		t.Errorf("expected 3 projects, got %d", len(projects))
	}

	// Test Close lifecycle
	err = coord.Close(ctx)
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	// Verify cache is cleared
	coord.mu.RLock()
	cacheLen := len(coord.repoCache)
	coord.mu.RUnlock()
	if cacheLen != 0 {
		t.Errorf("expected cache to be empty after Close, got %d entries", cacheLen)
	}
}

// TestIntegration_SaveFindDeleteCycle tests AC3, AC4, AC6
func TestIntegration_SaveFindDeleteCycle(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup project directory
	projDir := setupIntegrationProjectDir(t, basePath, "cycle-proj")

	// Setup config
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("cycle-proj", "/integration/cycle", "", false)

	mockLoader := &integrationMockConfigLoader{
		config: cfg,
	}

	coord := NewRepositoryCoordinator(mockLoader, &integrationMockDirectoryManager{basePath: basePath}, basePath)

	// Save
	project := createIntegrationTestProject("/integration/cycle")
	err := coord.Save(ctx, project)
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	// Find by ID
	found, err := coord.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.Path != project.Path {
		t.Errorf("expected path %s, got %s", project.Path, found.Path)
	}

	// Delete
	err = coord.Delete(ctx, project.ID)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	// Verify deleted
	_, err = coord.FindByID(ctx, project.ID)
	if err != domain.ErrProjectNotFound {
		t.Errorf("expected ErrProjectNotFound after delete, got %v", err)
	}

	// Verify directly in repo
	repo, _ := sqlite.NewProjectRepository(projDir)
	_, err = repo.FindByID(ctx, project.ID)
	if err != domain.ErrProjectNotFound {
		t.Errorf("expected project deleted from repo, got %v", err)
	}
}

// TestIntegration_ServiceLayerCompatibility tests AC1: zero service layer code changes
func TestIntegration_ServiceLayerCompatibility(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup project directory
	setupIntegrationProjectDir(t, basePath, "service-proj")

	// Setup config
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("service-proj", "/integration/service", "", false)

	mockLoader := &integrationMockConfigLoader{
		config: cfg,
	}

	coord := NewRepositoryCoordinator(mockLoader, &integrationMockDirectoryManager{basePath: basePath}, basePath)

	// Use coordinator through ports.ProjectRepository interface (service layer pattern)
	var repo ports.ProjectRepository = coord

	// All interface methods should work through the interface type
	project := createIntegrationTestProject("/integration/service")

	// Save
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("interface Save failed: %v", err)
	}

	// FindByID
	if _, err := repo.FindByID(ctx, project.ID); err != nil {
		t.Fatalf("interface FindByID failed: %v", err)
	}

	// FindByPath
	if _, err := repo.FindByPath(ctx, project.Path); err != nil {
		t.Fatalf("interface FindByPath failed: %v", err)
	}

	// FindAll
	if _, err := repo.FindAll(ctx); err != nil {
		t.Fatalf("interface FindAll failed: %v", err)
	}

	// FindActive
	if _, err := repo.FindActive(ctx); err != nil {
		t.Fatalf("interface FindActive failed: %v", err)
	}

	// FindHibernated
	if _, err := repo.FindHibernated(ctx); err != nil {
		t.Fatalf("interface FindHibernated failed: %v", err)
	}

	// UpdateState
	if err := repo.UpdateState(ctx, project.ID, domain.StateHibernated); err != nil {
		t.Fatalf("interface UpdateState failed: %v", err)
	}

	// Delete
	if err := repo.Delete(ctx, project.ID); err != nil {
		t.Fatalf("interface Delete failed: %v", err)
	}
}

// TestIntegration_LazyLoading tests AC5: connections opened lazily on-demand
// AC4: Given 20 projects tracked, verify lazy loading behavior by checking cache state
func TestIntegration_LazyLoading(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// AC4 requires 20 projects tracked
	const projectCount = 20

	// Setup 20 project directories (AC4 requirement)
	for i := 1; i <= projectCount; i++ {
		dirName := "lazy-proj-" + strconv.Itoa(i)
		projDir := setupIntegrationProjectDir(t, basePath, dirName)

		// Pre-populate each database with a project
		repo, err := sqlite.NewProjectRepository(projDir)
		if err != nil {
			t.Fatalf("failed to create repo for %s: %v", dirName, err)
		}
		project := createIntegrationTestProject("/integration/lazy/" + dirName)
		if err := repo.Save(ctx, project); err != nil {
			t.Fatalf("failed to save project %s: %v", dirName, err)
		}
	}

	// Setup config with 20 projects
	cfg := ports.NewConfig()
	for i := 1; i <= projectCount; i++ {
		dirName := "lazy-proj-" + strconv.Itoa(i)
		cfg.SetProjectEntry(dirName, "/integration/lazy/"+dirName, "", false)
	}

	mockLoader := &integrationMockConfigLoader{
		config: cfg,
	}

	coord := NewRepositoryCoordinator(mockLoader, &integrationMockDirectoryManager{basePath: basePath}, basePath)

	// Initially cache should be empty (lazy loading)
	coord.mu.RLock()
	initialCacheLen := len(coord.repoCache)
	coord.mu.RUnlock()
	if initialCacheLen != 0 {
		t.Errorf("expected empty cache initially, got %d entries", initialCacheLen)
	}

	// Access one project via FindByID
	targetProject := createIntegrationTestProject("/integration/lazy/lazy-proj-1")
	_, err := coord.FindByID(ctx, targetProject.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}

	// Cache should have entry for the accessed project
	coord.mu.RLock()
	_, hasTarget := coord.repoCache["lazy-proj-1"]
	cacheSizeAfterFind := len(coord.repoCache)
	coord.mu.RUnlock()
	if !hasTarget {
		t.Error("expected cache to have lazy-proj-1 after access")
	}

	// Not all 20 should be loaded after accessing just one (FindByID searches sequentially until found)
	// Some will be loaded during iteration, but we verify lazy behavior exists
	t.Logf("Cache size after FindByID: %d / %d projects", cacheSizeAfterFind, projectCount)

	// Close and verify cache is cleared
	err = coord.Close(ctx)
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	coord.mu.RLock()
	cacheSizeAfterClose := len(coord.repoCache)
	coord.mu.RUnlock()
	if cacheSizeAfterClose != 0 {
		t.Errorf("expected cache to be empty after Close, got %d entries", cacheSizeAfterClose)
	}

	// FindAll should work after Close (lazy reload)
	projects, err := coord.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll after Close returned error: %v", err)
	}
	if len(projects) != projectCount {
		t.Errorf("expected %d projects, got %d", projectCount, len(projects))
	}
}

// TestIntegration_NewProjectCreation tests AC13: new project creation via Save
func TestIntegration_NewProjectCreation(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	cfg := ports.NewConfig()
	configSaved := false
	newDirCreated := ""

	mockLoader := &integrationMockConfigLoader{
		config: cfg,
		saveFunc: func(ctx context.Context, c *ports.Config) error {
			configSaved = true
			return nil
		},
	}

	mockDirManager := &integrationMockDirectoryManager{
		basePath: basePath,
		ensureFunc: func(ctx context.Context, projectPath string) (string, error) {
			// Create directory like real DirectoryManager would
			dirName := filepath.Base(projectPath)
			fullPath := filepath.Join(basePath, dirName)
			os.MkdirAll(fullPath, 0755)
			// Create marker file
			markerPath := filepath.Join(fullPath, ".project-path")
			os.WriteFile(markerPath, []byte(projectPath), 0644)
			newDirCreated = fullPath
			return fullPath, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, mockDirManager, basePath)

	// Save a new project (not in config)
	newProject := createIntegrationTestProject("/integration/brand-new")
	err := coord.Save(ctx, newProject)
	if err != nil {
		t.Fatalf("Save for new project returned error: %v", err)
	}

	// Verify directory was created
	if newDirCreated == "" {
		t.Error("expected new directory to be created")
	}

	// Verify config was saved
	if !configSaved {
		t.Error("expected config to be saved")
	}

	// Verify project can be found
	found, err := coord.FindByID(ctx, newProject.ID)
	if err != nil {
		t.Fatalf("FindByID for new project returned error: %v", err)
	}
	if found.Path != newProject.Path {
		t.Errorf("expected path %s, got %s", newProject.Path, found.Path)
	}
}

// TestIntegration_CloseLifecycle tests AC12: Close clears cache properly
func TestIntegration_CloseLifecycle(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup multiple projects
	for i := 1; i <= 3; i++ {
		dirName := "close-proj-" + strconv.Itoa(i)
		projDir := setupIntegrationProjectDir(t, basePath, dirName)
		repo, _ := sqlite.NewProjectRepository(projDir)
		project := createIntegrationTestProject("/integration/close/" + dirName)
		repo.Save(ctx, project)
	}

	cfg := ports.NewConfig()
	for i := 1; i <= 3; i++ {
		dirName := "close-proj-" + strconv.Itoa(i)
		cfg.SetProjectEntry(dirName, "/integration/close/"+dirName, "", false)
	}

	mockLoader := &integrationMockConfigLoader{
		config: cfg,
	}

	coord := NewRepositoryCoordinator(mockLoader, &integrationMockDirectoryManager{basePath: basePath}, basePath)

	// Access all projects to populate cache
	_, err := coord.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}

	// Verify cache is populated
	coord.mu.RLock()
	cacheLenBefore := len(coord.repoCache)
	coord.mu.RUnlock()
	if cacheLenBefore == 0 {
		t.Error("expected cache to be populated before Close")
	}

	// Close
	err = coord.Close(ctx)
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	// Verify cache is cleared
	coord.mu.RLock()
	cacheLenAfter := len(coord.repoCache)
	coord.mu.RUnlock()
	if cacheLenAfter != 0 {
		t.Errorf("expected cache to be empty after Close, got %d entries", cacheLenAfter)
	}

	// Coordinator should still work after Close (lazy loading will recreate)
	_, err = coord.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll after Close returned error: %v", err)
	}
}

// Helper functions for integration tests

func setupIntegrationProjectDir(t *testing.T, basePath, dirName string) string {
	t.Helper()
	projDir := filepath.Join(basePath, dirName)
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	markerPath := filepath.Join(projDir, ".project-path")
	if err := os.WriteFile(markerPath, []byte("/test/path"), 0644); err != nil {
		t.Fatalf("failed to create marker file: %v", err)
	}
	return projDir
}

func createIntegrationTestProject(path string) *domain.Project {
	now := time.Now()
	return &domain.Project{
		ID:             domain.GenerateID(path),
		Name:           filepath.Base(path),
		Path:           path,
		State:          domain.StateActive,
		CreatedAt:      now,
		UpdatedAt:      now,
		LastActivityAt: now,
	}
}

// Mock implementations for integration tests

type integrationMockConfigLoader struct {
	config   *ports.Config
	saveFunc func(ctx context.Context, config *ports.Config) error
}

func (m *integrationMockConfigLoader) Load(ctx context.Context) (*ports.Config, error) {
	return m.config, nil
}

func (m *integrationMockConfigLoader) Save(ctx context.Context, config *ports.Config) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, config)
	}
	return nil
}

type integrationMockDirectoryManager struct {
	basePath   string
	ensureFunc func(ctx context.Context, projectPath string) (string, error)
	deleteFunc func(ctx context.Context, projectPath string) error
}

func (m *integrationMockDirectoryManager) GetProjectDirName(ctx context.Context, projectPath string) (string, error) {
	return filepath.Base(projectPath), nil
}

func (m *integrationMockDirectoryManager) EnsureProjectDir(ctx context.Context, projectPath string) (string, error) {
	if m.ensureFunc != nil {
		return m.ensureFunc(ctx, projectPath)
	}
	dirName := filepath.Base(projectPath)
	return filepath.Join(m.basePath, dirName), nil
}

func (m *integrationMockDirectoryManager) DeleteProjectDir(ctx context.Context, projectPath string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, projectPath)
	}
	return nil // Default: no-op
}
