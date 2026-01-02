package services

import (
	"context"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// mockStateRepo is a minimal mock for testing StateService
type mockStateRepo struct {
	projects map[string]*domain.Project
	saveErr  error
}

func newMockStateRepo() *mockStateRepo {
	return &mockStateRepo{
		projects: make(map[string]*domain.Project),
	}
}

func (m *mockStateRepo) Save(ctx context.Context, project *domain.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.projects[project.ID] = project
	return nil
}

func (m *mockStateRepo) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return nil, domain.ErrProjectNotFound
}

// Unused interface methods - minimal implementation
func (m *mockStateRepo) FindByPath(context.Context, string) (*domain.Project, error) {
	return nil, nil
}
func (m *mockStateRepo) FindAll(context.Context) ([]*domain.Project, error)        { return nil, nil }
func (m *mockStateRepo) FindActive(context.Context) ([]*domain.Project, error)     { return nil, nil }
func (m *mockStateRepo) FindHibernated(context.Context) ([]*domain.Project, error) { return nil, nil }
func (m *mockStateRepo) Delete(context.Context, string) error                      { return nil }
func (m *mockStateRepo) UpdateState(context.Context, string, domain.ProjectState) error {
	return nil
}
func (m *mockStateRepo) UpdateLastActivity(context.Context, string, time.Time) error {
	return nil
}
func (m *mockStateRepo) ResetProject(context.Context, string) error { return nil }
func (m *mockStateRepo) ResetAll(context.Context) (int, error)      { return 0, nil }

func TestNewStateService(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestStateService_Hibernate_ValidTransition(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	// Create active project
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	repo.projects[project.ID] = project

	ctx := context.Background()
	err := svc.Hibernate(ctx, project.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify state changed
	updated := repo.projects[project.ID]
	if updated.State != domain.StateHibernated {
		t.Errorf("expected StateHibernated, got %v", updated.State)
	}

	// Verify HibernatedAt is set
	if updated.HibernatedAt == nil {
		t.Error("expected HibernatedAt to be set")
	}
}

func TestStateService_Activate_ValidTransition(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	// Create hibernated project
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateHibernated
	hibernatedTime := time.Now().Add(-24 * time.Hour)
	project.HibernatedAt = &hibernatedTime
	repo.projects[project.ID] = project

	ctx := context.Background()
	err := svc.Activate(ctx, project.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify state changed
	updated := repo.projects[project.ID]
	if updated.State != domain.StateActive {
		t.Errorf("expected StateActive, got %v", updated.State)
	}

	// Verify HibernatedAt is cleared
	if updated.HibernatedAt != nil {
		t.Error("expected HibernatedAt to be nil after activation")
	}
}

func TestStateService_Hibernate_AlreadyHibernated_ReturnsError(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	// Create already hibernated project
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateHibernated
	hibernatedTime := time.Now()
	project.HibernatedAt = &hibernatedTime
	repo.projects[project.ID] = project

	ctx := context.Background()
	err := svc.Hibernate(ctx, project.ID)

	if err == nil {
		t.Fatal("expected error for already hibernated project")
	}
	if err != domain.ErrInvalidStateTransition {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestStateService_Activate_AlreadyActive_ReturnsError(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	// Create already active project
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	repo.projects[project.ID] = project

	ctx := context.Background()
	err := svc.Activate(ctx, project.ID)

	if err == nil {
		t.Fatal("expected error for already active project")
	}
	if err != domain.ErrInvalidStateTransition {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestStateService_Hibernate_Favorite_ReturnsError(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	// Create favorite project
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	project.IsFavorite = true
	repo.projects[project.ID] = project

	ctx := context.Background()
	err := svc.Hibernate(ctx, project.ID)

	if err == nil {
		t.Fatal("expected error for favorite project")
	}
	if err != domain.ErrFavoriteCannotHibernate {
		t.Errorf("expected ErrFavoriteCannotHibernate, got %v", err)
	}
}

func TestStateService_Hibernate_ProjectNotFound_ReturnsError(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	ctx := context.Background()
	err := svc.Hibernate(ctx, "nonexistent-id")

	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
	if err != domain.ErrProjectNotFound {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestStateService_Activate_ProjectNotFound_ReturnsError(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	ctx := context.Background()
	err := svc.Activate(ctx, "nonexistent-id")

	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
	if err != domain.ErrProjectNotFound {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestStateService_Hibernate_UpdatesUpdatedAt(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	// Create active project with old UpdatedAt
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	oldUpdatedAt := time.Now().Add(-24 * time.Hour)
	project.UpdatedAt = oldUpdatedAt
	repo.projects[project.ID] = project

	ctx := context.Background()
	beforeHibernate := time.Now()
	err := svc.Hibernate(ctx, project.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify UpdatedAt was updated
	updated := repo.projects[project.ID]
	if !updated.UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
	if updated.UpdatedAt.Before(beforeHibernate) {
		t.Error("expected UpdatedAt to be after operation start")
	}
}

func TestStateService_Activate_UpdatesUpdatedAt(t *testing.T) {
	repo := newMockStateRepo()
	svc := NewStateService(repo)

	// Create hibernated project with old UpdatedAt
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateHibernated
	hibernatedTime := time.Now().Add(-48 * time.Hour)
	project.HibernatedAt = &hibernatedTime
	oldUpdatedAt := time.Now().Add(-24 * time.Hour)
	project.UpdatedAt = oldUpdatedAt
	repo.projects[project.ID] = project

	ctx := context.Background()
	beforeActivate := time.Now()
	err := svc.Activate(ctx, project.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify UpdatedAt was updated
	updated := repo.projects[project.ID]
	if !updated.UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
	if updated.UpdatedAt.Before(beforeActivate) {
		t.Error("expected UpdatedAt to be after operation start")
	}
}

// errSaveTest is a sentinel error for testing Save() failures
var errSaveTest = domain.ErrConfigInvalid // reuse existing error for test

func TestStateService_Hibernate_SaveError_ReturnsError(t *testing.T) {
	repo := newMockStateRepo()
	repo.saveErr = errSaveTest // Configure mock to fail on Save()
	svc := NewStateService(repo)

	// Create active project
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	repo.projects[project.ID] = project

	ctx := context.Background()
	err := svc.Hibernate(ctx, project.ID)

	if err == nil {
		t.Fatal("expected error from Save() failure")
	}
	if err != errSaveTest {
		t.Errorf("expected errSaveTest, got %v", err)
	}
}

func TestStateService_Activate_SaveError_ReturnsError(t *testing.T) {
	repo := newMockStateRepo()
	repo.saveErr = errSaveTest // Configure mock to fail on Save()
	svc := NewStateService(repo)

	// Create hibernated project
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateHibernated
	hibernatedTime := time.Now().Add(-24 * time.Hour)
	project.HibernatedAt = &hibernatedTime
	repo.projects[project.ID] = project

	ctx := context.Background()
	err := svc.Activate(ctx, project.ID)

	if err == nil {
		t.Fatal("expected error from Save() failure")
	}
	if err != errSaveTest {
		t.Errorf("expected errSaveTest, got %v", err)
	}
}
