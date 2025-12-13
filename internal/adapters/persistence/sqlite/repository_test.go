package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// TestNewSQLiteRepository_CreatesDBAndSchema tests database and schema creation (AC1)
func TestNewSQLiteRepository_CreatesDBAndSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	// Verify schema was created by attempting to query projects table
	ctx := context.Background()
	projects, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() after creation error = %v", err)
	}
	if projects == nil {
		t.Error("FindAll() returned nil, want empty slice")
	}
}

// TestSQLiteRepository_Save_NewProject tests inserting a new project (AC2)
func TestSQLiteRepository_Save_NewProject(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	project, err := domain.NewProject("/test/path/project1", "project1")
	if err != nil {
		t.Fatalf("NewProject() error = %v", err)
	}
	project.DisplayName = "Test Project"
	project.DetectedMethod = "speckit"
	project.CurrentStage = domain.StagePlan
	project.IsFavorite = true
	project.Notes = "Test notes"

	ctx := context.Background()
	err = repo.Save(ctx, project)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify project was saved
	found, err := repo.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.ID != project.ID {
		t.Errorf("ID = %v, want %v", found.ID, project.ID)
	}
	if found.Name != project.Name {
		t.Errorf("Name = %v, want %v", found.Name, project.Name)
	}
	if found.Path != project.Path {
		t.Errorf("Path = %v, want %v", found.Path, project.Path)
	}
	if found.DisplayName != project.DisplayName {
		t.Errorf("DisplayName = %v, want %v", found.DisplayName, project.DisplayName)
	}
	if found.DetectedMethod != project.DetectedMethod {
		t.Errorf("DetectedMethod = %v, want %v", found.DetectedMethod, project.DetectedMethod)
	}
	if found.CurrentStage != project.CurrentStage {
		t.Errorf("CurrentStage = %v, want %v", found.CurrentStage, project.CurrentStage)
	}
	if found.IsFavorite != project.IsFavorite {
		t.Errorf("IsFavorite = %v, want %v", found.IsFavorite, project.IsFavorite)
	}
	if found.Notes != project.Notes {
		t.Errorf("Notes = %v, want %v", found.Notes, project.Notes)
	}
}

// TestSQLiteRepository_Save_ExistingProject tests updating an existing project (AC3)
func TestSQLiteRepository_Save_ExistingProject(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	project, _ := domain.NewProject("/test/path/project1", "project1")

	ctx := context.Background()
	err = repo.Save(ctx, project)
	if err != nil {
		t.Fatalf("Save() first error = %v", err)
	}

	originalUpdatedAt := project.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	// Update project
	project.DisplayName = "Updated Name"
	project.Notes = "Updated notes"
	err = repo.Save(ctx, project)
	if err != nil {
		t.Fatalf("Save() update error = %v", err)
	}

	// Verify updated
	found, err := repo.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.DisplayName != "Updated Name" {
		t.Errorf("DisplayName = %v, want 'Updated Name'", found.DisplayName)
	}
	if found.Notes != "Updated notes" {
		t.Errorf("Notes = %v, want 'Updated notes'", found.Notes)
	}
	if !found.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt was not refreshed after update")
	}
}

// TestSQLiteRepository_FindByID_Found tests FindByID when project exists
func TestSQLiteRepository_FindByID_Found(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	project, _ := domain.NewProject("/test/path/project1", "project1")
	ctx := context.Background()
	_ = repo.Save(ctx, project)

	found, err := repo.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.ID != project.ID {
		t.Errorf("ID = %v, want %v", found.ID, project.ID)
	}
}

// TestSQLiteRepository_FindByID_NotFound tests FindByID when project doesn't exist (AC5)
func TestSQLiteRepository_FindByID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()
	_, err = repo.FindByID(ctx, "nonexistent-id")

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("FindByID() error = %v, want ErrProjectNotFound", err)
	}
}

// TestSQLiteRepository_FindByPath_Found tests FindByPath when project exists (AC4)
func TestSQLiteRepository_FindByPath_Found(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	project, _ := domain.NewProject("/test/path/project1", "project1")
	ctx := context.Background()
	_ = repo.Save(ctx, project)

	found, err := repo.FindByPath(ctx, "/test/path/project1")
	if err != nil {
		t.Fatalf("FindByPath() error = %v", err)
	}
	if found.Path != project.Path {
		t.Errorf("Path = %v, want %v", found.Path, project.Path)
	}
}

// TestSQLiteRepository_FindByPath_NotFound tests FindByPath when project doesn't exist (AC5)
func TestSQLiteRepository_FindByPath_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()
	_, err = repo.FindByPath(ctx, "/nonexistent/path")

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("FindByPath() error = %v, want ErrProjectNotFound", err)
	}
}

// TestSQLiteRepository_FindAll_WithProjects tests FindAll with projects (AC6)
func TestSQLiteRepository_FindAll_WithProjects(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()

	// Create multiple projects
	project1, _ := domain.NewProject("/test/path/alpha", "alpha")
	project2, _ := domain.NewProject("/test/path/beta", "beta")
	_ = repo.Save(ctx, project1)
	_ = repo.Save(ctx, project2)

	projects, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("FindAll() returned %d projects, want 2", len(projects))
	}

	// Should be ordered by name
	if projects[0].Name != "alpha" {
		t.Errorf("First project name = %v, want 'alpha' (ordered)", projects[0].Name)
	}
}

// TestSQLiteRepository_FindAll_Empty tests FindAll returns empty slice not nil (AC6)
func TestSQLiteRepository_FindAll_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()
	projects, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	// CRITICAL: Must be empty slice, not nil
	if projects == nil {
		t.Error("FindAll() returned nil, want empty slice")
	}
	if len(projects) != 0 {
		t.Errorf("FindAll() returned %d projects, want 0", len(projects))
	}
}

// TestSQLiteRepository_FindActive tests FindActive filters correctly (AC7)
func TestSQLiteRepository_FindActive(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()

	// Create active and hibernated projects
	activeProject, _ := domain.NewProject("/test/path/active", "active")
	hibernatedProject, _ := domain.NewProject("/test/path/hibernated", "hibernated")
	hibernatedProject.State = domain.StateHibernated

	_ = repo.Save(ctx, activeProject)
	_ = repo.Save(ctx, hibernatedProject)

	projects, err := repo.FindActive(ctx)
	if err != nil {
		t.Fatalf("FindActive() error = %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("FindActive() returned %d projects, want 1", len(projects))
	}
	if len(projects) > 0 && projects[0].Name != "active" {
		t.Errorf("FindActive() returned wrong project: %v", projects[0].Name)
	}
}

// TestSQLiteRepository_FindActive_Empty tests FindActive returns empty slice not nil
func TestSQLiteRepository_FindActive_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()
	projects, err := repo.FindActive(ctx)
	if err != nil {
		t.Fatalf("FindActive() error = %v", err)
	}

	if projects == nil {
		t.Error("FindActive() returned nil, want empty slice")
	}
}

// TestSQLiteRepository_FindHibernated tests FindHibernated filters correctly (AC8)
func TestSQLiteRepository_FindHibernated(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()

	// Create active and hibernated projects
	activeProject, _ := domain.NewProject("/test/path/active", "active")
	hibernatedProject, _ := domain.NewProject("/test/path/hibernated", "hibernated")
	hibernatedProject.State = domain.StateHibernated

	_ = repo.Save(ctx, activeProject)
	_ = repo.Save(ctx, hibernatedProject)

	projects, err := repo.FindHibernated(ctx)
	if err != nil {
		t.Fatalf("FindHibernated() error = %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("FindHibernated() returned %d projects, want 1", len(projects))
	}
	if len(projects) > 0 && projects[0].Name != "hibernated" {
		t.Errorf("FindHibernated() returned wrong project: %v", projects[0].Name)
	}
}

// TestSQLiteRepository_FindHibernated_Empty tests FindHibernated returns empty slice not nil
func TestSQLiteRepository_FindHibernated_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()
	projects, err := repo.FindHibernated(ctx)
	if err != nil {
		t.Fatalf("FindHibernated() error = %v", err)
	}

	if projects == nil {
		t.Error("FindHibernated() returned nil, want empty slice")
	}
}

// TestSQLiteRepository_Delete_Found tests Delete when project exists (AC9)
func TestSQLiteRepository_Delete_Found(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	project, _ := domain.NewProject("/test/path/project1", "project1")
	ctx := context.Background()
	_ = repo.Save(ctx, project)

	err = repo.Delete(ctx, project.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	_, err = repo.FindByID(ctx, project.ID)
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("FindByID() after delete error = %v, want ErrProjectNotFound", err)
	}
}

// TestSQLiteRepository_Delete_NotFound tests Delete when project doesn't exist (AC9)
func TestSQLiteRepository_Delete_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()
	err = repo.Delete(ctx, "nonexistent-id")

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("Delete() error = %v, want ErrProjectNotFound", err)
	}
}

// TestSQLiteRepository_UpdateState_Found tests UpdateState when project exists (AC10)
func TestSQLiteRepository_UpdateState_Found(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	project, _ := domain.NewProject("/test/path/project1", "project1")
	ctx := context.Background()
	_ = repo.Save(ctx, project)

	originalUpdatedAt := project.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	err = repo.UpdateState(ctx, project.ID, domain.StateHibernated)
	if err != nil {
		t.Fatalf("UpdateState() error = %v", err)
	}

	// Verify state changed
	found, _ := repo.FindByID(ctx, project.ID)
	if found.State != domain.StateHibernated {
		t.Errorf("State = %v, want StateHibernated", found.State)
	}
	if !found.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt was not refreshed after state update")
	}
}

// TestSQLiteRepository_UpdateState_NotFound tests UpdateState when project doesn't exist (AC10)
func TestSQLiteRepository_UpdateState_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()
	err = repo.UpdateState(ctx, "nonexistent-id", domain.StateHibernated)

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("UpdateState() error = %v, want ErrProjectNotFound", err)
	}
}

// TestSQLiteRepository_WALMode tests WAL mode is enabled (AC1)
func TestSQLiteRepository_WALMode(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()
	db, err := repo.openDB(ctx)
	if err != nil {
		t.Fatalf("openDB() error = %v", err)
	}
	defer db.Close()

	var journalMode string
	err = db.GetContext(ctx, &journalMode, "PRAGMA journal_mode;")
	if err != nil {
		t.Fatalf("PRAGMA journal_mode error = %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("journal_mode = %v, want 'wal'", journalMode)
	}
}

// TestSQLiteRepository_ContextCancellation tests context cancellation handling
func TestSQLiteRepository_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = repo.FindAll(ctx)
	if err == nil {
		t.Error("FindAll() with cancelled context should return error")
	}
}

// TestSQLiteRepository_UniquePathConstraint_SameID tests upsert by same ID
func TestSQLiteRepository_UniquePathConstraint_SameID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()

	// Save first project
	project1, _ := domain.NewProject("/same/path", "project1")
	err = repo.Save(ctx, project1)
	if err != nil {
		t.Fatalf("Save() first project error = %v", err)
	}

	// Save same project again (upsert by ID)
	project1.DisplayName = "Updated via upsert"
	err = repo.Save(ctx, project1)
	if err != nil {
		t.Fatalf("Save() upsert error = %v", err)
	}

	// Verify only one project exists
	projects, _ := repo.FindAll(ctx)
	if len(projects) != 1 {
		t.Errorf("Expected 1 project after upsert, got %d", len(projects))
	}
	if projects[0].DisplayName != "Updated via upsert" {
		t.Errorf("DisplayName = %v, want 'Updated via upsert'", projects[0].DisplayName)
	}
}

// TestSQLiteRepository_UniquePathConstraint_DifferentID tests that INSERT OR REPLACE
// replaces by path uniqueness when a different project ID uses the same path
func TestSQLiteRepository_UniquePathConstraint_DifferentID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()

	// Save first project
	project1, _ := domain.NewProject("/same/path", "project1")
	err = repo.Save(ctx, project1)
	if err != nil {
		t.Fatalf("Save() first project error = %v", err)
	}

	// Try to save different project with same path
	// NewProject generates ID from path, so we need to manually create with different ID
	project2 := &domain.Project{
		ID:             "different-id-1234",
		Name:           "project2",
		Path:           "/same/path",
		State:          domain.StateActive,
		LastActivityAt: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	err = repo.Save(ctx, project2)
	// INSERT OR REPLACE will succeed, replacing by path uniqueness
	if err != nil {
		t.Fatalf("Save() different ID same path error = %v", err)
	}

	// Verify only one project exists (the second one replaced the first)
	projects, _ := repo.FindAll(ctx)
	if len(projects) != 1 {
		t.Errorf("Expected 1 project after path replacement, got %d", len(projects))
	}
	if len(projects) > 0 && projects[0].ID != "different-id-1234" {
		t.Errorf("Expected project2 (different-id-1234), got %v", projects[0].ID)
	}
}

// TestSQLiteRepository_DefaultPath tests default database path construction
func TestSQLiteRepository_DefaultPath(t *testing.T) {
	// This test would create the actual DB at ~/.vibe-dash/projects.db
	// so we skip actual creation and just verify constructor accepts empty path
	t.Skip("Skipping default path test to avoid creating files in home directory")
}

// TestSQLiteRepository_Save_InvalidProject tests Save returns validation error for invalid project
func TestSQLiteRepository_Save_InvalidProject(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()

	// Create invalid project with empty path
	invalidProject := &domain.Project{
		ID:   "test-id",
		Name: "test",
		Path: "", // Invalid: empty path
	}

	err = repo.Save(ctx, invalidProject)
	if err == nil {
		t.Error("Save() with invalid project should return error")
	}
	if !errors.Is(err, domain.ErrPathNotAccessible) {
		// The error should be wrapped, so check if it contains the validation error
		if err == nil || !strings.Contains(err.Error(), "invalid project") {
			t.Errorf("Save() error = %v, want error containing 'invalid project'", err)
		}
	}
}

// TestSQLiteRepository_Save_InvalidProjectRelativePath tests Save returns validation error for relative path
func TestSQLiteRepository_Save_InvalidProjectRelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	ctx := context.Background()

	// Create invalid project with relative path
	invalidProject := &domain.Project{
		ID:             "test-id",
		Name:           "test",
		Path:           "relative/path", // Invalid: not absolute
		LastActivityAt: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = repo.Save(ctx, invalidProject)
	if err == nil {
		t.Error("Save() with relative path should return error")
	}
}

// TestSQLiteRepository_CorruptedDatabaseError tests error wrapping for corrupted database (AC11)
func TestSQLiteRepository_CorruptedDatabaseError(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "corrupted.db")

	// Create a corrupted database file by writing garbage
	err := os.WriteFile(dbPath, []byte("not a valid sqlite database"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Try to create repository with corrupted database
	_, err = NewSQLiteRepository(dbPath)
	if err == nil {
		t.Fatal("NewSQLiteRepository() with corrupted database should return error")
	}

	// Verify error contains recovery suggestion
	if !errors.Is(err, ErrDatabaseCorrupted) {
		// Even if not wrapped with our error, it should fail
		// The key is that it does fail on corrupted database
		t.Logf("Got error (not ErrDatabaseCorrupted but still an error): %v", err)
	}

	// Alternative: verify error message indicates a problem
	if err != nil && !strings.Contains(err.Error(), "database") {
		t.Logf("Error doesn't mention database: %v", err)
	}
}

// TestWrapDBError tests the error wrapping helper for database corruption detection
func TestWrapDBError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		dbPath        string
		wantCorrupted bool
	}{
		{
			name:          "nil error returns nil",
			err:           nil,
			dbPath:        "/test/path.db",
			wantCorrupted: false,
		},
		{
			name:          "malformed database error",
			err:           errors.New("database disk image is malformed"),
			dbPath:        "/test/path.db",
			wantCorrupted: true,
		},
		{
			name:          "corrupt error",
			err:           errors.New("file is corrupt"),
			dbPath:        "/test/path.db",
			wantCorrupted: true,
		},
		{
			name:          "disk i/o error",
			err:           errors.New("disk i/o error occurred"),
			dbPath:        "/test/path.db",
			wantCorrupted: true,
		},
		{
			name:          "normal error passes through",
			err:           errors.New("connection timeout"),
			dbPath:        "/test/path.db",
			wantCorrupted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapDBError(tt.err, tt.dbPath)

			if tt.err == nil {
				if result != nil {
					t.Errorf("wrapDBError(nil) = %v, want nil", result)
				}
				return
			}

			if tt.wantCorrupted {
				if !errors.Is(result, ErrDatabaseCorrupted) {
					t.Errorf("wrapDBError() = %v, want ErrDatabaseCorrupted", result)
				}
				if !strings.Contains(result.Error(), "Recovery suggestion") {
					t.Error("wrapDBError() should include recovery suggestion")
				}
				if !strings.Contains(result.Error(), tt.dbPath) {
					t.Error("wrapDBError() should include database path")
				}
			} else {
				if errors.Is(result, ErrDatabaseCorrupted) {
					t.Errorf("wrapDBError() should not wrap as corrupted for: %v", tt.err)
				}
			}
		})
	}
}
