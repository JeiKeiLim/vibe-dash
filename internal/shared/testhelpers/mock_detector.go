package testhelpers

import (
	"context"
	"sync"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// MockDetector provides a configurable mock for ports.Detector.
// Supports result/error injection for testing detection scenarios.
type MockDetector struct {
	mu sync.RWMutex

	// Result configuration
	detectResult                   *domain.DetectionResult
	detectErr                      error
	multipleResult                 []*domain.DetectionResult
	multipleErr                    error
	coexistenceSelectionWinner     *domain.DetectionResult
	coexistenceSelectionAll        []*domain.DetectionResult
	coexistenceSelectionErr        error
	coexistenceSelectionResultsSet bool

	// Call tracking
	detectCalls []string
}

// NewMockDetector creates a new MockDetector.
func NewMockDetector() *MockDetector {
	return &MockDetector{}
}

// SetResult sets the result to return from Detect calls.
func (m *MockDetector) SetResult(result *domain.DetectionResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.detectResult = result
}

// SetError sets the error to return from Detect calls.
func (m *MockDetector) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.detectErr = err
}

// SetMultipleResult sets the result to return from DetectMultiple calls.
func (m *MockDetector) SetMultipleResult(results []*domain.DetectionResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.multipleResult = results
}

// SetMultipleError sets the error to return from DetectMultiple calls.
func (m *MockDetector) SetMultipleError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.multipleErr = err
}

// DetectCalls returns the list of paths passed to Detect.
func (m *MockDetector) DetectCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.detectCalls))
	copy(result, m.detectCalls)
	return result
}

// Detect implements ports.Detector.
func (m *MockDetector) Detect(_ context.Context, path string) (*domain.DetectionResult, error) {
	m.mu.Lock()
	m.detectCalls = append(m.detectCalls, path)
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.detectResult, m.detectErr
}

// DetectMultiple implements ports.Detector.
func (m *MockDetector) DetectMultiple(_ context.Context, _ string) ([]*domain.DetectionResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.multipleResult, m.multipleErr
}

// SetCoexistenceSelectionResult sets the results to return from DetectWithCoexistenceSelection.
func (m *MockDetector) SetCoexistenceSelectionResult(winner *domain.DetectionResult, all []*domain.DetectionResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.coexistenceSelectionWinner = winner
	m.coexistenceSelectionAll = all
	m.coexistenceSelectionResultsSet = true
}

// SetCoexistenceSelectionError sets the error to return from DetectWithCoexistenceSelection.
func (m *MockDetector) SetCoexistenceSelectionError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.coexistenceSelectionErr = err
}

// DetectWithCoexistenceSelection implements ports.Detector.
func (m *MockDetector) DetectWithCoexistenceSelection(_ context.Context, _ string) (*domain.DetectionResult, []*domain.DetectionResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.coexistenceSelectionErr != nil {
		return nil, nil, m.coexistenceSelectionErr
	}
	if m.coexistenceSelectionResultsSet {
		return m.coexistenceSelectionWinner, m.coexistenceSelectionAll, nil
	}
	// Default: delegate to single detection result
	return m.detectResult, nil, m.detectErr
}
