package testhelpers

import (
	"context"
	"sync"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// MockRepository provides a configurable mock for ports.ProjectRepository.
// Supports error injection and call tracking for comprehensive test coverage.
type MockRepository struct {
	mu sync.RWMutex

	// Projects stores the mock project data
	Projects map[string]*domain.Project

	// Error injection
	saveErr    error
	deleteErr  error
	findErr    error
	findAllErr error
	resetErr   error

	// Call tracking
	saveCalls   []string
	deleteCalls []string
	resetCalls  []string
}

// NewMockRepository creates a new MockRepository with empty project map.
func NewMockRepository() *MockRepository {
	return &MockRepository{
		Projects: make(map[string]*domain.Project),
	}
}

// WithProjects sets initial projects and returns the mock for chaining.
func (m *MockRepository) WithProjects(projects []*domain.Project) *MockRepository {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range projects {
		m.Projects[p.Path] = p
	}
	return m
}

// SetSaveError sets the error to return from Save calls.
func (m *MockRepository) SetSaveError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveErr = err
}

// SetDeleteError sets the error to return from Delete calls.
func (m *MockRepository) SetDeleteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteErr = err
}

// SetFindError sets the error to return from FindByID and FindByPath calls.
func (m *MockRepository) SetFindError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.findErr = err
}

// SetFindAllError sets the error to return from FindAll calls.
func (m *MockRepository) SetFindAllError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.findAllErr = err
}

// SetResetError sets the error to return from ResetProject and ResetAll calls.
func (m *MockRepository) SetResetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resetErr = err
}

// SaveCalls returns the list of paths passed to Save.
func (m *MockRepository) SaveCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.saveCalls))
	copy(result, m.saveCalls)
	return result
}

// DeleteCalls returns the list of IDs passed to Delete.
func (m *MockRepository) DeleteCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.deleteCalls))
	copy(result, m.deleteCalls)
	return result
}

// ResetCalls returns the list of project IDs passed to ResetProject.
func (m *MockRepository) ResetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.resetCalls))
	copy(result, m.resetCalls)
	return result
}

// Save implements ports.ProjectRepository.
func (m *MockRepository) Save(_ context.Context, project *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveCalls = append(m.saveCalls, project.Path)
	if m.saveErr != nil {
		return m.saveErr
	}
	m.Projects[project.Path] = project
	return nil
}

// FindByID implements ports.ProjectRepository.
func (m *MockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, p := range m.Projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

// FindByPath implements ports.ProjectRepository.
func (m *MockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.findErr != nil {
		return nil, m.findErr
	}
	if p, ok := m.Projects[path]; ok {
		return p, nil
	}
	return nil, domain.ErrProjectNotFound
}

// FindAll implements ports.ProjectRepository.
func (m *MockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	result := make([]*domain.Project, 0, len(m.Projects))
	for _, p := range m.Projects {
		result = append(result, p)
	}
	return result, nil
}

// FindActive implements ports.ProjectRepository.
func (m *MockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	result := make([]*domain.Project, 0)
	for _, p := range m.Projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

// FindHibernated implements ports.ProjectRepository.
func (m *MockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	result := make([]*domain.Project, 0)
	for _, p := range m.Projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

// Delete implements ports.ProjectRepository.
func (m *MockRepository) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalls = append(m.deleteCalls, id)
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for path, p := range m.Projects {
		if p.ID == id {
			delete(m.Projects, path)
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

// UpdateState implements ports.ProjectRepository.
func (m *MockRepository) UpdateState(_ context.Context, id string, state domain.ProjectState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.Projects {
		if p.ID == id {
			p.State = state
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

// UpdateLastActivity implements ports.ProjectRepository.
func (m *MockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}

// ResetProject implements ports.ProjectRepository.
func (m *MockRepository) ResetProject(_ context.Context, projectID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resetCalls = append(m.resetCalls, projectID)
	if m.resetErr != nil {
		return m.resetErr
	}
	return nil
}

// ResetAll implements ports.ProjectRepository.
func (m *MockRepository) ResetAll(_ context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.resetErr != nil {
		return 0, m.resetErr
	}
	return len(m.Projects), nil
}
