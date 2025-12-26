package testhelpers

import (
	"context"
	"path/filepath"
	"sync"
)

// MockDirectoryManager provides a configurable mock for ports.DirectoryManager.
// Supports error injection and call tracking for testing directory operations.
type MockDirectoryManager struct {
	mu sync.RWMutex

	// Error injection
	deleteErr error

	// Call tracking
	deleteCalls []string
}

// NewMockDirectoryManager creates a new MockDirectoryManager.
func NewMockDirectoryManager() *MockDirectoryManager {
	return &MockDirectoryManager{
		deleteCalls: make([]string, 0),
	}
}

// DeleteCalls returns the list of project paths passed to DeleteProjectDir.
func (m *MockDirectoryManager) DeleteCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.deleteCalls))
	copy(result, m.deleteCalls)
	return result
}

// SetDeleteError sets the error to return from DeleteProjectDir calls.
func (m *MockDirectoryManager) SetDeleteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteErr = err
}

// GetProjectDirName implements ports.DirectoryManager.
func (m *MockDirectoryManager) GetProjectDirName(_ context.Context, projectPath string) (string, error) {
	return filepath.Base(projectPath), nil
}

// EnsureProjectDir implements ports.DirectoryManager.
func (m *MockDirectoryManager) EnsureProjectDir(_ context.Context, projectPath string) (string, error) {
	return filepath.Join("/tmp/test", filepath.Base(projectPath)), nil
}

// DeleteProjectDir implements ports.DirectoryManager.
func (m *MockDirectoryManager) DeleteProjectDir(_ context.Context, projectPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalls = append(m.deleteCalls, projectPath)
	return m.deleteErr
}
