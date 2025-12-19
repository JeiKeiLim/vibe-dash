package ports_test

import (
	"context"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// mockRepository verifies interface compliance at compile time
type mockRepository struct {
	projects map[string]*domain.Project
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		projects: make(map[string]*domain.Project),
	}
}

func (m *mockRepository) Save(ctx context.Context, project *domain.Project) error {
	m.projects[project.ID] = project
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return nil, domain.ErrProjectNotFound
}

func (m *mockRepository) FindByPath(ctx context.Context, path string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.Path == path {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *mockRepository) FindAll(ctx context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0, len(m.projects))
	for _, p := range m.projects {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockRepository) FindActive(ctx context.Context) ([]*domain.Project, error) {
	var result []*domain.Project
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockRepository) FindHibernated(ctx context.Context) ([]*domain.Project, error) {
	var result []*domain.Project
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.projects[id]; !ok {
		return domain.ErrProjectNotFound
	}
	delete(m.projects, id)
	return nil
}

func (m *mockRepository) UpdateState(ctx context.Context, id string, state domain.ProjectState) error {
	p, ok := m.projects[id]
	if !ok {
		return domain.ErrProjectNotFound
	}
	p.State = state
	return nil
}

func (m *mockRepository) UpdateLastActivity(ctx context.Context, id string, timestamp time.Time) error {
	p, ok := m.projects[id]
	if !ok {
		return domain.ErrProjectNotFound
	}
	p.LastActivityAt = timestamp
	return nil
}

// Compile-time interface compliance check
var _ ports.ProjectRepository = (*mockRepository)(nil)

func TestProjectRepository_InterfaceCompliance(t *testing.T) {
	var repo ports.ProjectRepository = newMockRepository()

	// Create test project
	project, err := domain.NewProject("/test/path", "test-project")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	t.Run("Save accepts context and project", func(t *testing.T) {
		ctx := context.Background()
		err := repo.Save(ctx, project)
		if err != nil {
			t.Errorf("Save() error = %v, want nil", err)
		}
	})

	t.Run("FindByID returns saved project", func(t *testing.T) {
		ctx := context.Background()
		found, err := repo.FindByID(ctx, project.ID)
		if err != nil {
			t.Fatalf("FindByID() error = %v, want nil", err)
		}
		if found.ID != project.ID {
			t.Errorf("FindByID().ID = %q, want %q", found.ID, project.ID)
		}
	})

	t.Run("FindByPath returns saved project", func(t *testing.T) {
		ctx := context.Background()
		found, err := repo.FindByPath(ctx, project.Path)
		if err != nil {
			t.Fatalf("FindByPath() error = %v, want nil", err)
		}
		if found.Path != project.Path {
			t.Errorf("FindByPath().Path = %q, want %q", found.Path, project.Path)
		}
	})

	t.Run("FindAll returns all projects", func(t *testing.T) {
		ctx := context.Background()
		all, err := repo.FindAll(ctx)
		if err != nil {
			t.Fatalf("FindAll() error = %v, want nil", err)
		}
		if len(all) != 1 {
			t.Errorf("FindAll() returned %d projects, want 1", len(all))
		}
	})

	t.Run("FindActive returns only active projects", func(t *testing.T) {
		ctx := context.Background()
		active, err := repo.FindActive(ctx)
		if err != nil {
			t.Fatalf("FindActive() error = %v, want nil", err)
		}
		if len(active) != 1 {
			t.Errorf("FindActive() returned %d projects, want 1", len(active))
		}
	})

	t.Run("FindHibernated returns empty for all active", func(t *testing.T) {
		ctx := context.Background()
		hibernated, err := repo.FindHibernated(ctx)
		if err != nil {
			t.Fatalf("FindHibernated() error = %v, want nil", err)
		}
		if len(hibernated) != 0 {
			t.Errorf("FindHibernated() returned %d projects, want 0", len(hibernated))
		}
	})

	t.Run("UpdateState changes project state to Hibernated", func(t *testing.T) {
		ctx := context.Background()
		err := repo.UpdateState(ctx, project.ID, domain.StateHibernated)
		if err != nil {
			t.Fatalf("UpdateState() error = %v, want nil", err)
		}

		// Verify state changed
		found, _ := repo.FindByID(ctx, project.ID)
		if found.State != domain.StateHibernated {
			t.Errorf("UpdateState() state = %v, want %v", found.State, domain.StateHibernated)
		}
	})

	t.Run("UpdateState changes project state back to Active", func(t *testing.T) {
		ctx := context.Background()
		// Project is already hibernated from previous test
		err := repo.UpdateState(ctx, project.ID, domain.StateActive)
		if err != nil {
			t.Fatalf("UpdateState() error = %v, want nil", err)
		}

		// Verify state changed back to Active
		found, _ := repo.FindByID(ctx, project.ID)
		if found.State != domain.StateActive {
			t.Errorf("UpdateState() state = %v, want %v", found.State, domain.StateActive)
		}

		// Verify project now appears in FindActive results
		active, _ := repo.FindActive(ctx)
		foundInActive := false
		for _, p := range active {
			if p.ID == project.ID {
				foundInActive = true
				break
			}
		}
		if !foundInActive {
			t.Error("Project should appear in FindActive() after UpdateState to StateActive")
		}
	})

	t.Run("Delete removes project", func(t *testing.T) {
		ctx := context.Background()
		err := repo.Delete(ctx, project.ID)
		if err != nil {
			t.Fatalf("Delete() error = %v, want nil", err)
		}

		// Verify deleted
		_, err = repo.FindByID(ctx, project.ID)
		if err != domain.ErrProjectNotFound {
			t.Errorf("FindByID() after Delete() error = %v, want ErrProjectNotFound", err)
		}
	})

	t.Run("FindByID returns ErrProjectNotFound for missing", func(t *testing.T) {
		ctx := context.Background()
		_, err := repo.FindByID(ctx, "nonexistent-id")
		if err != domain.ErrProjectNotFound {
			t.Errorf("FindByID() error = %v, want ErrProjectNotFound", err)
		}
	})
}
