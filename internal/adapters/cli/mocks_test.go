package cli_test

import (
	"context"
	"path/filepath"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
)

// MockDirectoryManager implements ports.DirectoryManager for testing.
type MockDirectoryManager struct {
	deleteErr   error
	deleteCalls []string // Track deleted paths
}

func NewMockDirectoryManager() *MockDirectoryManager {
	return &MockDirectoryManager{
		deleteCalls: make([]string, 0),
	}
}

func (m *MockDirectoryManager) GetProjectDirName(_ context.Context, projectPath string) (string, error) {
	return filepath.Base(projectPath), nil
}

func (m *MockDirectoryManager) EnsureProjectDir(_ context.Context, projectPath string) (string, error) {
	return filepath.Join("/tmp/test", filepath.Base(projectPath)), nil
}

func (m *MockDirectoryManager) DeleteProjectDir(_ context.Context, projectPath string) error {
	m.deleteCalls = append(m.deleteCalls, projectPath)
	return m.deleteErr
}

// SetMockDirectoryManager is a helper to set the mock directory manager for tests.
func SetMockDirectoryManager(dm *MockDirectoryManager) {
	cli.SetDirectoryManager(dm)
}

// ClearMockDirectoryManager resets the directory manager to nil.
func ClearMockDirectoryManager() {
	cli.SetDirectoryManager(nil)
}
