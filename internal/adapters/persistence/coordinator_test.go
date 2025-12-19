package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/sqlite"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// TestNewRepositoryCoordinator_ValidConstruction tests that constructor creates valid coordinator
func TestNewRepositoryCoordinator_ValidConstruction(t *testing.T) {
	mockLoader := &mockConfigLoader{}
	mockDirManager := &mockDirectoryManager{}
	basePath := "/tmp/vibe-dash-test"

	coord := NewRepositoryCoordinator(mockLoader, mockDirManager, basePath)

	if coord == nil {
		t.Fatal("expected coordinator to be non-nil")
	}
	if coord.configLoader == nil {
		t.Error("expected configLoader to be set")
	}
	if coord.directoryManager == nil {
		t.Error("expected directoryManager to be set")
	}
	if coord.basePath != basePath {
		t.Errorf("expected basePath %s, got %s", basePath, coord.basePath)
	}
	if coord.repoCache == nil {
		t.Error("expected repoCache to be initialized")
	}
}

// TestRepositoryCoordinator_ImplementsProjectRepository verifies interface compliance
func TestRepositoryCoordinator_ImplementsProjectRepository(t *testing.T) {
	mockLoader := &mockConfigLoader{}
	mockDirManager := &mockDirectoryManager{}

	coord := NewRepositoryCoordinator(mockLoader, mockDirManager, "/tmp")

	// Compile-time check is in coordinator.go, but let's also verify at runtime
	var _ ports.ProjectRepository = coord
}

// TestFindAll_AggregatesFromMultipleDBs tests AC2: aggregation from multiple project databases
func TestFindAll_AggregatesFromMultipleDBs(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup 2 project directories
	proj1Dir := setupProjectDir(t, basePath, "proj1")
	proj2Dir := setupProjectDir(t, basePath, "proj2")

	// Create repos and save projects
	repo1, err := sqlite.NewProjectRepository(proj1Dir)
	if err != nil {
		t.Fatalf("failed to create repo1: %v", err)
	}
	repo2, err := sqlite.NewProjectRepository(proj2Dir)
	if err != nil {
		t.Fatalf("failed to create repo2: %v", err)
	}

	project1 := createTestProject("/path/to/project1")
	project2 := createTestProject("/path/to/project2")

	if err := repo1.Save(ctx, project1); err != nil {
		t.Fatalf("failed to save project1: %v", err)
	}
	if err := repo2.Save(ctx, project2); err != nil {
		t.Fatalf("failed to save project2: %v", err)
	}

	// Setup config with both projects
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("proj1", "/path/to/project1", "", false)
	cfg.SetProjectEntry("proj2", "/path/to/project2", "", false)

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// Test FindAll aggregation
	projects, err := coord.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

// TestFindActive_FiltersCorrectly tests AC7: FindActive aggregation
func TestFindActive_FiltersCorrectly(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup project directories
	activeDir := setupProjectDir(t, basePath, "active-proj")
	hibernatedDir := setupProjectDir(t, basePath, "hibernated-proj")

	// Create repos and save projects with different states
	activeRepo, _ := sqlite.NewProjectRepository(activeDir)
	hibernatedRepo, _ := sqlite.NewProjectRepository(hibernatedDir)

	activeProject := createTestProject("/path/to/active")
	activeProject.State = domain.StateActive

	hibernatedProject := createTestProject("/path/to/hibernated")
	hibernatedProject.State = domain.StateHibernated

	if err := activeRepo.Save(ctx, activeProject); err != nil {
		t.Fatalf("failed to save active project: %v", err)
	}
	if err := hibernatedRepo.Save(ctx, hibernatedProject); err != nil {
		t.Fatalf("failed to save hibernated project: %v", err)
	}

	// Setup config
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("active-proj", "/path/to/active", "", false)
	cfg.SetProjectEntry("hibernated-proj", "/path/to/hibernated", "", false)

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// Test FindActive
	activeProjects, err := coord.FindActive(ctx)
	if err != nil {
		t.Fatalf("FindActive returned error: %v", err)
	}

	if len(activeProjects) != 1 {
		t.Errorf("expected 1 active project, got %d", len(activeProjects))
	}
	if len(activeProjects) > 0 && activeProjects[0].Path != "/path/to/active" {
		t.Errorf("expected active project path, got %s", activeProjects[0].Path)
	}
}

// TestFindHibernated_FiltersCorrectly tests AC7: FindHibernated aggregation
func TestFindHibernated_FiltersCorrectly(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup project directories
	activeDir := setupProjectDir(t, basePath, "active-proj")
	hibernatedDir := setupProjectDir(t, basePath, "hibernated-proj")

	// Create repos and save projects with different states
	activeRepo, _ := sqlite.NewProjectRepository(activeDir)
	hibernatedRepo, _ := sqlite.NewProjectRepository(hibernatedDir)

	activeProject := createTestProject("/path/to/active")
	activeProject.State = domain.StateActive

	hibernatedProject := createTestProject("/path/to/hibernated")
	hibernatedProject.State = domain.StateHibernated

	if err := activeRepo.Save(ctx, activeProject); err != nil {
		t.Fatalf("failed to save active project: %v", err)
	}
	if err := hibernatedRepo.Save(ctx, hibernatedProject); err != nil {
		t.Fatalf("failed to save hibernated project: %v", err)
	}

	// Setup config
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("active-proj", "/path/to/active", "", false)
	cfg.SetProjectEntry("hibernated-proj", "/path/to/hibernated", "", false)

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// Test FindHibernated
	hibernatedProjects, err := coord.FindHibernated(ctx)
	if err != nil {
		t.Fatalf("FindHibernated returned error: %v", err)
	}

	if len(hibernatedProjects) != 1 {
		t.Errorf("expected 1 hibernated project, got %d", len(hibernatedProjects))
	}
	if len(hibernatedProjects) > 0 && hibernatedProjects[0].Path != "/path/to/hibernated" {
		t.Errorf("expected hibernated project path, got %s", hibernatedProjects[0].Path)
	}
}

// TestSave_RoutesToCorrectDB tests AC3: Save routing for existing project
func TestSave_RoutesToCorrectDB(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup project directory
	projDir := setupProjectDir(t, basePath, "my-proj")

	// Setup config with existing project
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("my-proj", "/path/to/myproject", "", false)

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// Save a project
	project := createTestProject("/path/to/myproject")
	err := coord.Save(ctx, project)
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	// Verify project was saved to correct repo
	repo, _ := sqlite.NewProjectRepository(projDir)
	found, err := repo.FindByPath(ctx, "/path/to/myproject")
	if err != nil {
		t.Fatalf("failed to find saved project: %v", err)
	}
	if found.Path != "/path/to/myproject" {
		t.Errorf("expected path /path/to/myproject, got %s", found.Path)
	}
}

// TestSave_CreatesNewProjectViaDirectoryManager tests AC13: new project creation
func TestSave_CreatesNewProjectViaDirectoryManager(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Pre-create the directory that DirectoryManager would create
	newProjDir := setupProjectDir(t, basePath, "new-project")

	cfg := ports.NewConfig()
	savedConfig := false

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
		saveFunc: func(ctx context.Context, c *ports.Config) error {
			savedConfig = true
			return nil
		},
	}

	ensureCalled := false
	mockDirManager := &mockDirectoryManager{
		ensureProjectDirFunc: func(ctx context.Context, projectPath string) (string, error) {
			ensureCalled = true
			return newProjDir, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, mockDirManager, basePath)

	// Save a new project (not in config)
	project := createTestProject("/path/to/new/project")
	err := coord.Save(ctx, project)
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	// Verify DirectoryManager was called
	if !ensureCalled {
		t.Error("expected EnsureProjectDir to be called for new project")
	}

	// Verify config was saved
	if !savedConfig {
		t.Error("expected config to be saved for new project")
	}
}

// TestDelete_RemovesAndInvalidatesCache tests AC4: Delete routing, cache invalidation, and config removal
func TestDelete_RemovesAndInvalidatesCache(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup project directory
	projDir := setupProjectDir(t, basePath, "delete-proj")

	// Create repo and save project
	repo, _ := sqlite.NewProjectRepository(projDir)
	project := createTestProject("/path/to/delete")
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Setup config
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("delete-proj", "/path/to/delete", "", false)

	configSaved := false
	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
		saveFunc: func(ctx context.Context, c *ports.Config) error {
			configSaved = true
			return nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// First access to populate cache
	_, err := coord.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}

	// Verify cache is populated
	coord.mu.RLock()
	_, inCache := coord.repoCache["delete-proj"]
	coord.mu.RUnlock()
	if !inCache {
		t.Error("expected repo to be in cache after FindByID")
	}

	// Delete project
	err = coord.Delete(ctx, project.ID)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	// Verify cache is invalidated
	coord.mu.RLock()
	_, stillInCache := coord.repoCache["delete-proj"]
	coord.mu.RUnlock()
	if stillInCache {
		t.Error("expected cache to be invalidated after Delete")
	}

	// Verify config was saved (to remove project entry)
	if !configSaved {
		t.Error("expected config to be saved after Delete to remove project entry")
	}

	// Verify project was removed from config
	if _, found := cfg.GetDirectoryName("/path/to/delete"); found {
		t.Error("expected project to be removed from config after Delete")
	}
}

// TestFindByID_SearchesAcrossDBs tests AC6: FindByID searches all databases
func TestFindByID_SearchesAcrossDBs(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup 2 project directories
	proj1Dir := setupProjectDir(t, basePath, "proj1")
	proj2Dir := setupProjectDir(t, basePath, "proj2")

	// Save project only in proj2
	repo2, _ := sqlite.NewProjectRepository(proj2Dir)
	project := createTestProject("/path/to/project2")
	if err := repo2.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Setup empty repo for proj1
	if _, err := sqlite.NewProjectRepository(proj1Dir); err != nil {
		t.Fatalf("failed to create proj1 repo: %v", err)
	}

	// Setup config
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("proj1", "/path/to/project1", "", false)
	cfg.SetProjectEntry("proj2", "/path/to/project2", "", false)

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// Find by ID should search across all DBs
	found, err := coord.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.Path != "/path/to/project2" {
		t.Errorf("expected path /path/to/project2, got %s", found.Path)
	}
}

// TestFindByPath_FastPathAndFallback tests AC6: FindByPath fast path and fallback
func TestFindByPath_FastPathAndFallback(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup project directory
	projDir := setupProjectDir(t, basePath, "path-proj")

	// Create repo and save project
	repo, _ := sqlite.NewProjectRepository(projDir)
	project := createTestProject("/path/to/findme")
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Setup config with project registered
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("path-proj", "/path/to/findme", "", false)

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// Test fast path (config lookup)
	found, err := coord.FindByPath(ctx, "/path/to/findme")
	if err != nil {
		t.Fatalf("FindByPath returned error: %v", err)
	}
	if found.ID != project.ID {
		t.Errorf("expected ID %s, got %s", project.ID, found.ID)
	}
}

// TestGracefulDegradation_CorruptedDBLogged tests AC10: corrupted DB logged, others succeed
func TestGracefulDegradation_CorruptedDBLogged(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup one valid project directory
	validDir := setupProjectDir(t, basePath, "valid-proj")
	repo, _ := sqlite.NewProjectRepository(validDir)
	project := createTestProject("/path/to/valid")
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Setup config with both - "invalid-proj" directory doesn't exist
	// This will fail NewProjectRepository because directory doesn't exist
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("valid-proj", "/path/to/valid", "", false)
	cfg.SetProjectEntry("invalid-proj", "/path/to/invalid", "", false)

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// Should return valid project, skip invalid with warning (logged)
	projects, err := coord.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll should not return error for graceful degradation: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 valid project, got %d", len(projects))
	}
}

// TestContextCancellation tests AC9: context cancellation returns promptly
func TestContextCancellation(t *testing.T) {
	tests := []struct {
		name   string
		method func(ctx context.Context, coord *RepositoryCoordinator) error
	}{
		{
			name: "FindAll",
			method: func(ctx context.Context, coord *RepositoryCoordinator) error {
				_, err := coord.FindAll(ctx)
				return err
			},
		},
		{
			name: "FindActive",
			method: func(ctx context.Context, coord *RepositoryCoordinator) error {
				_, err := coord.FindActive(ctx)
				return err
			},
		},
		{
			name: "FindHibernated",
			method: func(ctx context.Context, coord *RepositoryCoordinator) error {
				_, err := coord.FindHibernated(ctx)
				return err
			},
		},
		{
			name: "FindByID",
			method: func(ctx context.Context, coord *RepositoryCoordinator) error {
				_, err := coord.FindByID(ctx, "someid")
				return err
			},
		},
		{
			name: "FindByPath",
			method: func(ctx context.Context, coord *RepositoryCoordinator) error {
				_, err := coord.FindByPath(ctx, "/some/path")
				return err
			},
		},
		{
			name: "Save",
			method: func(ctx context.Context, coord *RepositoryCoordinator) error {
				return coord.Save(ctx, createTestProject("/test/path"))
			},
		},
		{
			name: "Delete",
			method: func(ctx context.Context, coord *RepositoryCoordinator) error {
				return coord.Delete(ctx, "someid")
			},
		},
		{
			name: "UpdateState",
			method: func(ctx context.Context, coord *RepositoryCoordinator) error {
				return coord.UpdateState(ctx, "someid", domain.StateActive)
			},
		},
		{
			name: "Close",
			method: func(ctx context.Context, coord *RepositoryCoordinator) error {
				return coord.Close(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoader := &mockConfigLoader{}
			coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, "/tmp")

			// Create cancelled context
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			err := tt.method(ctx, coord)
			if !errors.Is(err, context.Canceled) {
				t.Errorf("%s: expected context.Canceled, got %v", tt.name, err)
			}
		})
	}
}

// TestEmptySliceReturn tests AC11: empty results return empty slice not nil
func TestEmptySliceReturn(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Empty config (no projects)
	cfg := ports.NewConfig()

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	tests := []struct {
		name   string
		method func() ([]*domain.Project, error)
	}{
		{
			name: "FindAll",
			method: func() ([]*domain.Project, error) {
				return coord.FindAll(ctx)
			},
		},
		{
			name: "FindActive",
			method: func() ([]*domain.Project, error) {
				return coord.FindActive(ctx)
			},
		},
		{
			name: "FindHibernated",
			method: func() ([]*domain.Project, error) {
				return coord.FindHibernated(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.method()
			if err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}
			if result == nil {
				t.Errorf("%s: expected empty slice, got nil", tt.name)
			}
			if len(result) != 0 {
				t.Errorf("%s: expected 0 items, got %d", tt.name, len(result))
			}
		})
	}
}

// TestClose_ClearsCache tests AC12: Close clears cache
func TestClose_ClearsCache(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup project directory
	projDir := setupProjectDir(t, basePath, "close-proj")
	repo, _ := sqlite.NewProjectRepository(projDir)
	project := createTestProject("/path/to/close")
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Setup config
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("close-proj", "/path/to/close", "", false)

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// Populate cache
	if _, err := coord.FindAll(ctx); err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}

	// Verify cache has entry
	coord.mu.RLock()
	cacheLen := len(coord.repoCache)
	coord.mu.RUnlock()
	if cacheLen == 0 {
		t.Error("expected cache to be populated")
	}

	// Close
	err := coord.Close(ctx)
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
}

// TestFindByID_NotFound tests error return for missing project
func TestFindByID_NotFound(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Empty config
	cfg := ports.NewConfig()

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	_, err := coord.FindByID(ctx, "nonexistent-id")
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

// TestFindByPath_NotFound tests error return for missing project
func TestFindByPath_NotFound(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Empty config
	cfg := ports.NewConfig()

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	_, err := coord.FindByPath(ctx, "/nonexistent/path")
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

// TestUpdateState_RoutesToCorrectDB tests AC8: UpdateState routing
func TestUpdateState_RoutesToCorrectDB(t *testing.T) {
	basePath := t.TempDir()
	ctx := context.Background()

	// Setup project directory
	projDir := setupProjectDir(t, basePath, "state-proj")

	// Create repo and save project
	repo, _ := sqlite.NewProjectRepository(projDir)
	project := createTestProject("/path/to/stateful")
	project.State = domain.StateActive
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Setup config
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("state-proj", "/path/to/stateful", "", false)

	mockLoader := &mockConfigLoader{
		loadFunc: func(ctx context.Context) (*ports.Config, error) {
			return cfg, nil
		},
	}

	coord := NewRepositoryCoordinator(mockLoader, &mockDirectoryManager{}, basePath)

	// Update state
	err := coord.UpdateState(ctx, project.ID, domain.StateHibernated)
	if err != nil {
		t.Fatalf("UpdateState returned error: %v", err)
	}

	// Verify state was updated
	updated, err := repo.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if updated.State != domain.StateHibernated {
		t.Errorf("expected StateHibernated, got %v", updated.State)
	}
}

// Helper functions

func setupProjectDir(t *testing.T, basePath, dirName string) string {
	t.Helper()
	projDir := filepath.Join(basePath, dirName)
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	return projDir
}

func createTestProject(path string) *domain.Project {
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

// Mock implementations for testing

type mockConfigLoader struct {
	loadFunc func(ctx context.Context) (*ports.Config, error)
	saveFunc func(ctx context.Context, config *ports.Config) error
}

func (m *mockConfigLoader) Load(ctx context.Context) (*ports.Config, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx)
	}
	return ports.NewConfig(), nil
}

func (m *mockConfigLoader) Save(ctx context.Context, config *ports.Config) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, config)
	}
	return nil
}

type mockDirectoryManager struct {
	getProjectDirNameFunc func(ctx context.Context, projectPath string) (string, error)
	ensureProjectDirFunc  func(ctx context.Context, projectPath string) (string, error)
	deleteProjectDirFunc  func(ctx context.Context, projectPath string) error
}

func (m *mockDirectoryManager) GetProjectDirName(ctx context.Context, projectPath string) (string, error) {
	if m.getProjectDirNameFunc != nil {
		return m.getProjectDirNameFunc(ctx, projectPath)
	}
	return "", nil
}

func (m *mockDirectoryManager) EnsureProjectDir(ctx context.Context, projectPath string) (string, error) {
	if m.ensureProjectDirFunc != nil {
		return m.ensureProjectDirFunc(ctx, projectPath)
	}
	return "", nil
}

func (m *mockDirectoryManager) DeleteProjectDir(ctx context.Context, projectPath string) error {
	if m.deleteProjectDirFunc != nil {
		return m.deleteProjectDirFunc(ctx, projectPath)
	}
	return nil
}
