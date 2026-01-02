package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// mockHibernationRepo extends mockStateRepo with FindActive support
type mockHibernationRepo struct {
	projects map[string]*domain.Project
	saveErr  error
}

func newMockHibernationRepo() *mockHibernationRepo {
	return &mockHibernationRepo{
		projects: make(map[string]*domain.Project),
	}
}

func (m *mockHibernationRepo) Save(ctx context.Context, project *domain.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.projects[project.ID] = project
	return nil
}

func (m *mockHibernationRepo) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return nil, domain.ErrProjectNotFound
}

func (m *mockHibernationRepo) FindActive(ctx context.Context) ([]*domain.Project, error) {
	var active []*domain.Project
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			active = append(active, p)
		}
	}
	return active, nil
}

// Unused interface methods - return domain.ErrProjectNotFound for consistency
func (m *mockHibernationRepo) FindByPath(context.Context, string) (*domain.Project, error) {
	return nil, domain.ErrProjectNotFound
}
func (m *mockHibernationRepo) FindAll(context.Context) ([]*domain.Project, error) { return nil, nil }
func (m *mockHibernationRepo) FindHibernated(context.Context) ([]*domain.Project, error) {
	return nil, nil
}
func (m *mockHibernationRepo) Delete(context.Context, string) error { return nil }
func (m *mockHibernationRepo) UpdateState(context.Context, string, domain.ProjectState) error {
	return nil
}
func (m *mockHibernationRepo) UpdateLastActivity(context.Context, string, time.Time) error {
	return nil
}
func (m *mockHibernationRepo) ResetProject(context.Context, string) error { return nil }
func (m *mockHibernationRepo) ResetAll(context.Context) (int, error)      { return 0, nil }

func TestNewHibernationService(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()

	svc := NewHibernationService(repo, stateSvc, config)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

// TestHibernationService_InactiveProjectHibernates (AC1)
// Project with 14+ days inactivity gets hibernated
func TestHibernationService_InactiveProjectHibernates(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig() // Default 14 days

	// Create project with 15 days inactivity
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	project.LastActivityAt = time.Now().Add(-15 * 24 * time.Hour)
	repo.projects[project.ID] = project

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 hibernated, got %d", count)
	}

	// Verify project is now hibernated
	if repo.projects[project.ID].State != domain.StateHibernated {
		t.Error("expected project to be hibernated")
	}
}

// TestHibernationService_ActiveProjectStaysActive (AC1)
// Project with <14 days inactivity stays active
func TestHibernationService_ActiveProjectStaysActive(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()

	// Create project with only 10 days inactivity
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	project.LastActivityAt = time.Now().Add(-10 * 24 * time.Hour)
	repo.projects[project.ID] = project

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 hibernated, got %d", count)
	}

	// Verify project is still active
	if repo.projects[project.ID].State != domain.StateActive {
		t.Error("expected project to stay active")
	}
}

// TestHibernationService_FavoriteNeverHibernates (AC2, FR30)
// Favorite project never hibernates regardless of inactivity
func TestHibernationService_FavoriteNeverHibernates(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()

	// Create favorite project with 100 days inactivity
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	project.IsFavorite = true
	project.LastActivityAt = time.Now().Add(-100 * 24 * time.Hour)
	repo.projects[project.ID] = project

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 hibernated (favorite protected), got %d", count)
	}

	// Verify project is still active
	if repo.projects[project.ID].State != domain.StateActive {
		t.Error("expected favorite project to stay active")
	}
}

// TestHibernationService_HibernatedProjectSkipped (AC7)
// Already hibernated project is skipped (via FindActive)
func TestHibernationService_HibernatedProjectSkipped(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()

	// Create already hibernated project
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateHibernated
	hibernatedAt := time.Now().Add(-5 * 24 * time.Hour)
	project.HibernatedAt = &hibernatedAt
	project.LastActivityAt = time.Now().Add(-100 * 24 * time.Hour)
	repo.projects[project.ID] = project

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 hibernated (already hibernated), got %d", count)
	}

	// Verify project is still hibernated
	if repo.projects[project.ID].State != domain.StateHibernated {
		t.Error("expected project to stay hibernated")
	}
}

// TestHibernationService_DisabledWithZeroDays (AC5)
// hibernation_days = 0 disables auto-hibernation
func TestHibernationService_DisabledWithZeroDays(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()
	config.HibernationDays = 0 // Disable auto-hibernation

	// Create project with 100 days inactivity
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	project.LastActivityAt = time.Now().Add(-100 * 24 * time.Hour)
	repo.projects[project.ID] = project

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 hibernated (disabled), got %d", count)
	}

	// Verify project is still active
	if repo.projects[project.ID].State != domain.StateActive {
		t.Error("expected project to stay active when auto-hibernation disabled")
	}
}

// TestHibernationService_PerProjectOverride (AC5)
// Per-project override is respected
func TestHibernationService_PerProjectOverride(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()
	config.HibernationDays = 14 // Global: 14 days

	// Create project with 10 days inactivity
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	project.LastActivityAt = time.Now().Add(-10 * 24 * time.Hour)
	repo.projects[project.ID] = project

	// Set per-project override: 7 days (project should be hibernated)
	sevenDays := 7
	config.Projects = map[string]ports.ProjectConfig{
		project.ID: {HibernationDays: &sevenDays},
	}

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 hibernated (per-project override), got %d", count)
	}

	// Verify project is now hibernated
	if repo.projects[project.ID].State != domain.StateHibernated {
		t.Error("expected project to be hibernated based on per-project override")
	}
}

// TestHibernationService_BoundaryCondition (AC1)
// Exactly 14 days inactivity should NOT hibernate (need > 14 days)
func TestHibernationService_BoundaryCondition(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig() // Default 14 days

	// Create project with exactly 14 days inactivity (minus 1 minute to ensure we're under)
	// The boundary is STRICTLY greater than 14 days, so 14 days exactly should NOT hibernate
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	// Use 14 days minus 1 minute to ensure we're solidly under the threshold
	project.LastActivityAt = time.Now().Add(-14*24*time.Hour + time.Minute)
	repo.projects[project.ID] = project

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 hibernated (boundary: under 14 days), got %d", count)
	}

	// Verify project is still active (need > 14 days, not >= 14 days)
	if repo.projects[project.ID].State != domain.StateActive {
		t.Error("expected project to stay active at boundary")
	}
}

// TestHibernationService_JustOverBoundary
// 14 days + 1 second should hibernate
func TestHibernationService_JustOverBoundary(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig() // Default 14 days

	// Create project with 14 days + 1 second inactivity
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	project.LastActivityAt = time.Now().Add(-14*24*time.Hour - time.Second)
	repo.projects[project.ID] = project

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 hibernated (just over 14 days), got %d", count)
	}

	// Verify project is now hibernated
	if repo.projects[project.ID].State != domain.StateHibernated {
		t.Error("expected project to be hibernated when just over boundary")
	}
}

// TestHibernationService_PartialFailure (AC8)
// Continues processing after single project fails
func TestHibernationService_PartialFailure(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()

	// Create two projects with 15 days inactivity
	project1, _ := domain.NewProject("/path/to/project1", "test-project-1")
	project1.State = domain.StateActive
	project1.LastActivityAt = time.Now().Add(-15 * 24 * time.Hour)
	project1.IsFavorite = true // This will cause Hibernate to fail with ErrFavoriteCannotHibernate
	repo.projects[project1.ID] = project1

	project2, _ := domain.NewProject("/path/to/project2", "test-project-2")
	project2.State = domain.StateActive
	project2.LastActivityAt = time.Now().Add(-15 * 24 * time.Hour)
	repo.projects[project2.ID] = project2

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error (partial failure should be handled), got %v", err)
	}
	// project1 is favorite so skipped before Hibernate() is called
	// project2 should be hibernated
	if count != 1 {
		t.Errorf("expected 1 hibernated, got %d", count)
	}

	// Verify project2 is now hibernated
	if repo.projects[project2.ID].State != domain.StateHibernated {
		t.Error("expected project2 to be hibernated")
	}
}

// TestHibernationService_PartialFailure_SaveError
// Tests that partial failure handling works when Save() fails
func TestHibernationService_PartialFailure_SaveError(t *testing.T) {
	repo := newMockHibernationRepo()

	// Create project with 15 days inactivity
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	project.LastActivityAt = time.Now().Add(-15 * 24 * time.Hour)
	repo.projects[project.ID] = project

	// Make Save() fail
	repo.saveErr = errors.New("save failed")

	stateSvc := NewStateService(repo)
	config := ports.NewConfig()
	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	// Should not return error (partial failure tolerance)
	if err != nil {
		t.Fatalf("expected no error (partial failure should be handled), got %v", err)
	}
	// Count should be 0 because hibernation failed
	if count != 0 {
		t.Errorf("expected 0 hibernated (save failed), got %d", count)
	}
}

// TestHibernationService_MultipleProjects
// Tests processing multiple projects with different conditions
func TestHibernationService_MultipleProjects(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig() // Default 14 days

	// Create 5 projects with various conditions
	// Project 1: Inactive (20 days) - should hibernate
	p1, _ := domain.NewProject("/path/to/p1", "p1")
	p1.State = domain.StateActive
	p1.LastActivityAt = time.Now().Add(-20 * 24 * time.Hour)
	repo.projects[p1.ID] = p1

	// Project 2: Active (5 days) - should stay active
	p2, _ := domain.NewProject("/path/to/p2", "p2")
	p2.State = domain.StateActive
	p2.LastActivityAt = time.Now().Add(-5 * 24 * time.Hour)
	repo.projects[p2.ID] = p2

	// Project 3: Favorite (30 days inactive) - should stay active
	p3, _ := domain.NewProject("/path/to/p3", "p3")
	p3.State = domain.StateActive
	p3.IsFavorite = true
	p3.LastActivityAt = time.Now().Add(-30 * 24 * time.Hour)
	repo.projects[p3.ID] = p3

	// Project 4: Already hibernated - should be skipped
	p4, _ := domain.NewProject("/path/to/p4", "p4")
	p4.State = domain.StateHibernated
	hibernatedAt := time.Now().Add(-10 * 24 * time.Hour)
	p4.HibernatedAt = &hibernatedAt
	p4.LastActivityAt = time.Now().Add(-50 * 24 * time.Hour)
	repo.projects[p4.ID] = p4

	// Project 5: Inactive (16 days) - should hibernate
	p5, _ := domain.NewProject("/path/to/p5", "p5")
	p5.State = domain.StateActive
	p5.LastActivityAt = time.Now().Add(-16 * 24 * time.Hour)
	repo.projects[p5.ID] = p5

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 hibernated (p1 and p5), got %d", count)
	}

	// Verify states
	if repo.projects[p1.ID].State != domain.StateHibernated {
		t.Error("expected p1 to be hibernated")
	}
	if repo.projects[p2.ID].State != domain.StateActive {
		t.Error("expected p2 to stay active")
	}
	if repo.projects[p3.ID].State != domain.StateActive {
		t.Error("expected p3 (favorite) to stay active")
	}
	if repo.projects[p4.ID].State != domain.StateHibernated {
		t.Error("expected p4 to stay hibernated")
	}
	if repo.projects[p5.ID].State != domain.StateHibernated {
		t.Error("expected p5 to be hibernated")
	}
}

// TestHibernationService_EmptyRepository
// Tests handling of empty repository
func TestHibernationService_EmptyRepository(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 hibernated (empty repo), got %d", count)
	}
}

// TestHibernationService_FindActiveError
// Tests that FindActive error propagates correctly
func TestHibernationService_FindActiveError(t *testing.T) {
	repo := &mockHibernationRepoWithFindActiveErr{
		findActiveErr: errors.New("database connection failed"),
	}
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	count, err := svc.CheckAndHibernate(ctx)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "database connection failed" {
		t.Errorf("expected 'database connection failed' error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 hibernated on error, got %d", count)
	}
}

// mockHibernationRepoWithFindActiveErr is a mock that returns error on FindActive
type mockHibernationRepoWithFindActiveErr struct {
	mockHibernationRepo
	findActiveErr error
}

func (m *mockHibernationRepoWithFindActiveErr) FindActive(ctx context.Context) ([]*domain.Project, error) {
	return nil, m.findActiveErr
}

// TestHibernationService_HibernatedAtTimestamp (AC6)
// Verifies HibernatedAt is set correctly when auto-hibernating
func TestHibernationService_HibernatedAtTimestamp(t *testing.T) {
	repo := newMockHibernationRepo()
	stateSvc := NewStateService(repo)
	config := ports.NewConfig()

	// Create project with 15 days inactivity
	project, _ := domain.NewProject("/path/to/project", "test-project")
	project.State = domain.StateActive
	project.LastActivityAt = time.Now().Add(-15 * 24 * time.Hour)
	repo.projects[project.ID] = project

	svc := NewHibernationService(repo, stateSvc, config)
	ctx := context.Background()

	beforeHibernate := time.Now()
	_, err := svc.CheckAndHibernate(ctx)
	afterHibernate := time.Now()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify HibernatedAt is set to approximately now
	hibernatedAt := repo.projects[project.ID].HibernatedAt
	if hibernatedAt == nil {
		t.Fatal("expected HibernatedAt to be set")
	}
	if hibernatedAt.Before(beforeHibernate) || hibernatedAt.After(afterHibernate) {
		t.Errorf("expected HibernatedAt to be between %v and %v, got %v",
			beforeHibernate, afterHibernate, *hibernatedAt)
	}

	// Verify UpdatedAt is also set
	updatedAt := repo.projects[project.ID].UpdatedAt
	if updatedAt.Before(beforeHibernate) || updatedAt.After(afterHibernate) {
		t.Errorf("expected UpdatedAt to be between %v and %v, got %v",
			beforeHibernate, afterHibernate, updatedAt)
	}
}
