package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestNewProjectRepository(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string // returns projectDir
		wantErr     bool
		errContains string
	}{
		{
			name: "creates state.db in correct location",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			wantErr: false,
		},
		{
			name: "fails if directory does not exist",
			setup: func(t *testing.T) string {
				t.Helper()
				return "/nonexistent/directory/path"
			},
			wantErr:     true,
			errContains: "project directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := tt.setup(t)

			repo, err := NewProjectRepository(projectDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if repo == nil {
				t.Error("expected non-nil repository")
				return
			}

			// Verify state.db was created
			dbPath := filepath.Join(projectDir, "state.db")
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				t.Errorf("state.db not created at %s", dbPath)
			}
		})
	}
}

func TestProjectRepository_WALMode(t *testing.T) {
	dir := t.TempDir()

	repo, err := NewProjectRepository(dir)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	// Open connection and check WAL mode
	db, err := repo.openDB(context.Background())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	var journalMode string
	if err := db.Get(&journalMode, "PRAGMA journal_mode;"); err != nil {
		t.Fatalf("failed to get journal_mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("expected journal_mode 'wal', got %q", journalMode)
	}
}

func TestProjectRepository_SchemaVersion(t *testing.T) {
	dir := t.TempDir()

	repo, err := NewProjectRepository(dir)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	db, err := repo.openDB(context.Background())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	var version int
	if err := db.Get(&version, "SELECT MAX(version) FROM schema_version;"); err != nil {
		t.Fatalf("failed to get schema version: %v", err)
	}

	if version != SchemaVersion {
		t.Errorf("expected schema version %d, got %d", SchemaVersion, version)
	}
}

func TestProjectRepository_CorruptionErrorMessage(t *testing.T) {
	projectDir := "/test/project/dir"
	dbPath := filepath.Join(projectDir, "state.db")

	// Test the wrapDBErrorForProject function
	err := wrapDBErrorForProject(
		domain.ErrProjectNotFound, // Not a corruption error
		dbPath,
	)

	// Should return the original error for non-corruption
	if err != domain.ErrProjectNotFound {
		t.Errorf("expected original error for non-corruption, got %v", err)
	}

	// Test with corruption-like error message
	corruptionErr := wrapDBErrorForProject(
		&mockError{"database disk image is malformed"},
		dbPath,
	)

	if !strings.Contains(corruptionErr.Error(), dbPath) {
		t.Errorf("corruption error should contain project path %q", dbPath)
	}
	if !strings.Contains(corruptionErr.Error(), "re-add project") {
		t.Errorf("corruption error should contain recovery suggestion")
	}
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

// setupProjectRepo is a helper to create a project repository for testing
func setupProjectRepo(t *testing.T) (*ProjectRepository, string) {
	t.Helper()
	dir := t.TempDir()

	repo, err := NewProjectRepository(dir)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	return repo, dir
}

// createTestProject creates a valid project for testing
func createTestProject(id, name, path string) *domain.Project {
	now := time.Now()
	return &domain.Project{
		ID:             id,
		Name:           name,
		Path:           path,
		CurrentStage:   domain.StageUnknown,
		State:          domain.StateActive,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func TestProjectRepository_Save(t *testing.T) {
	repo, _ := setupProjectRepo(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		project *domain.Project
		wantErr bool
	}{
		{
			name:    "saves valid project",
			project: createTestProject("test-id-1", "my-project", "/path/to/project"),
			wantErr: false,
		},
		{
			name: "saves project with all fields",
			project: &domain.Project{
				ID:                 "test-id-2",
				Name:               "full-project",
				Path:               "/path/to/full-project",
				DisplayName:        "Full Project",
				DetectedMethod:     "speckit",
				CurrentStage:       domain.StagePlan,
				Confidence:         domain.ConfidenceCertain,
				DetectionReasoning: "Found speckit.yaml",
				IsFavorite:         true,
				State:              domain.StateActive,
				Notes:              "Test notes",
				PathMissing:        false,
				LastActivityAt:     time.Now(),
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(ctx, tt.project)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify project was saved by retrieving it
			found, err := repo.FindByID(ctx, tt.project.ID)
			if err != nil {
				t.Errorf("failed to find saved project: %v", err)
				return
			}
			if found.ID != tt.project.ID {
				t.Errorf("expected ID %q, got %q", tt.project.ID, found.ID)
			}
			if found.Name != tt.project.Name {
				t.Errorf("expected Name %q, got %q", tt.project.Name, found.Name)
			}
		})
	}
}

func TestProjectRepository_FindByID(t *testing.T) {
	repo, _ := setupProjectRepo(t)
	ctx := context.Background()

	// Save a project first
	project := createTestProject("find-by-id-test", "test-project", "/test/path")
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "finds existing project",
			id:      "find-by-id-test",
			wantErr: nil,
		},
		{
			name:    "returns ErrProjectNotFound for missing project",
			id:      "nonexistent-id",
			wantErr: domain.ErrProjectNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByID(ctx, tt.id)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if found.ID != tt.id {
				t.Errorf("expected ID %q, got %q", tt.id, found.ID)
			}
		})
	}
}

func TestProjectRepository_FindByPath(t *testing.T) {
	repo, _ := setupProjectRepo(t)
	ctx := context.Background()

	// Save a project first
	project := createTestProject("find-by-path-test", "test-project", "/unique/test/path")
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{
			name:    "finds existing project by path",
			path:    "/unique/test/path",
			wantErr: nil,
		},
		{
			name:    "returns ErrProjectNotFound for missing path",
			path:    "/nonexistent/path",
			wantErr: domain.ErrProjectNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByPath(ctx, tt.path)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if found.Path != tt.path {
				t.Errorf("expected path %q, got %q", tt.path, found.Path)
			}
		})
	}
}

func TestProjectRepository_FindAll(t *testing.T) {
	repo, _ := setupProjectRepo(t)
	ctx := context.Background()

	// Initially empty
	projects, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if projects == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}

	// Add a project
	project := createTestProject("find-all-test", "test-project", "/test/path")
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Should find 1 project (per-project DB has 0-1 rows)
	projects, err = repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
}

func TestProjectRepository_FindActive(t *testing.T) {
	repo, _ := setupProjectRepo(t)
	ctx := context.Background()

	// Add active project
	project := createTestProject("find-active-test", "active-project", "/test/active")
	project.State = domain.StateActive
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Should find active project
	projects, err := repo.FindActive(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 active project, got %d", len(projects))
	}

	// Hibernate the project
	if err := repo.UpdateState(ctx, project.ID, domain.StateHibernated); err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	// Should not find any active projects
	projects, err = repo.FindActive(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 active projects, got %d", len(projects))
	}
}

func TestProjectRepository_FindHibernated(t *testing.T) {
	repo, _ := setupProjectRepo(t)
	ctx := context.Background()

	// Initially no hibernated projects
	projects, err := repo.FindHibernated(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 hibernated projects, got %d", len(projects))
	}

	// Add and hibernate a project
	project := createTestProject("find-hibernated-test", "hibernated-project", "/test/hibernated")
	project.State = domain.StateHibernated
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Should find hibernated project
	projects, err = repo.FindHibernated(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 hibernated project, got %d", len(projects))
	}
}

func TestProjectRepository_Delete(t *testing.T) {
	repo, _ := setupProjectRepo(t)
	ctx := context.Background()

	// Save a project first
	project := createTestProject("delete-test", "delete-project", "/test/delete")
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Verify it exists
	if _, err := repo.FindByID(ctx, project.ID); err != nil {
		t.Fatalf("project should exist: %v", err)
	}

	// Delete it
	if err := repo.Delete(ctx, project.ID); err != nil {
		t.Fatalf("failed to delete project: %v", err)
	}

	// Verify it's gone
	_, err := repo.FindByID(ctx, project.ID)
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound after delete, got %v", err)
	}

	// Deleting again should return ErrProjectNotFound
	err = repo.Delete(ctx, project.ID)
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound for second delete, got %v", err)
	}
}

func TestProjectRepository_UpdateState(t *testing.T) {
	repo, _ := setupProjectRepo(t)
	ctx := context.Background()

	// Save an active project
	project := createTestProject("update-state-test", "state-project", "/test/state")
	project.State = domain.StateActive
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Hibernate it
	if err := repo.UpdateState(ctx, project.ID, domain.StateHibernated); err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	// Verify state changed
	found, err := repo.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("failed to find project: %v", err)
	}
	if found.State != domain.StateHibernated {
		t.Errorf("expected state Hibernated, got %v", found.State)
	}

	// Activate it again
	if err := repo.UpdateState(ctx, project.ID, domain.StateActive); err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	found, _ = repo.FindByID(ctx, project.ID)
	if found.State != domain.StateActive {
		t.Errorf("expected state Active, got %v", found.State)
	}

	// Update non-existent project
	err = repo.UpdateState(ctx, "nonexistent-id", domain.StateActive)
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectRepository_ConcurrentAccess(t *testing.T) {
	repo, _ := setupProjectRepo(t)
	ctx := context.Background()

	// Save initial project
	project := createTestProject("concurrent-test", "concurrent-project", "/test/concurrent")
	if err := repo.Save(ctx, project); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	// Run concurrent operations
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*3)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)

		// Concurrent reads
		go func() {
			defer wg.Done()
			if _, err := repo.FindByID(ctx, project.ID); err != nil {
				errors <- err
			}
		}()

		// Concurrent FindAll
		go func() {
			defer wg.Done()
			if _, err := repo.FindAll(ctx); err != nil {
				errors <- err
			}
		}()

		// Concurrent state updates (alternating)
		go func(i int) {
			defer wg.Done()
			state := domain.StateActive
			if i%2 == 0 {
				state = domain.StateHibernated
			}
			if err := repo.UpdateState(ctx, project.ID, state); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("concurrent operation failed: %v", err)
	}
}
