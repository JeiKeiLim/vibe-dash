package cli_test

import (
	"context"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/testhelpers"
)

// MockRepository wraps testhelpers.MockRepository with aliases for CLI test compatibility.
// This allows tests to use field access syntax (mock.Projects) which is mapped to exported fields.
type MockRepository = testhelpers.MockRepository

// NewMockRepository creates a new MockRepository using testhelpers.
func NewMockRepository() *MockRepository {
	return testhelpers.NewMockRepository()
}

// MockDetector wraps testhelpers.MockDetector for CLI test compatibility.
type MockDetector = testhelpers.MockDetector

// NewMockDetector creates a new MockDetector using testhelpers.
func NewMockDetector() *MockDetector {
	return testhelpers.NewMockDetector()
}

// MockDirectoryManager wraps testhelpers.MockDirectoryManager for CLI test compatibility.
type MockDirectoryManager = testhelpers.MockDirectoryManager

// NewMockDirectoryManager creates a new MockDirectoryManager using testhelpers.
func NewMockDirectoryManager() *MockDirectoryManager {
	return testhelpers.NewMockDirectoryManager()
}

// SetMockDirectoryManager is a helper to set the mock directory manager for tests.
func SetMockDirectoryManager(dm *MockDirectoryManager) {
	cli.SetDirectoryManager(dm)
}

// ClearMockDirectoryManager resets the directory manager to nil.
func ClearMockDirectoryManager() {
	cli.SetDirectoryManager(nil)
}

// MockWaitingDetector provides a mock for ports.WaitingDetector for CLI tests.
// This is specific to CLI tests and not in shared testhelpers.
type MockWaitingDetector struct {
	isWaiting       bool
	waitingDuration time.Duration
}

func (m *MockWaitingDetector) IsWaiting(_ context.Context, _ *domain.Project) bool {
	return m.isWaiting
}

func (m *MockWaitingDetector) WaitingDuration(_ context.Context, _ *domain.Project) time.Duration {
	return m.waitingDuration
}
