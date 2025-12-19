package services

import (
	"context"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// mockThresholdResolver implements ports.ThresholdResolver for testing.
type mockThresholdResolver struct {
	defaultThreshold  int
	projectThresholds map[string]int
}

func newMockResolver(threshold int) *mockThresholdResolver {
	return &mockThresholdResolver{
		defaultThreshold:  threshold,
		projectThresholds: make(map[string]int),
	}
}

func (m *mockThresholdResolver) Resolve(projectID string) int {
	if threshold, ok := m.projectThresholds[projectID]; ok {
		return threshold
	}
	return m.defaultThreshold
}

func (m *mockThresholdResolver) setProjectThreshold(projectID string, threshold int) {
	m.projectThresholds[projectID] = threshold
}

func TestIsWaiting_NilProject(t *testing.T) {
	resolver := newMockResolver(10)
	detector := NewWaitingDetector(resolver)

	// Must not panic, should return false
	result := detector.IsWaiting(context.Background(), nil)
	if result {
		t.Errorf("IsWaiting(nil) = %v, want false", result)
	}
}

func TestIsWaiting_ThresholdDisabled(t *testing.T) {
	resolver := newMockResolver(0) // Disabled

	detector := NewWaitingDetector(resolver)

	now := time.Now()
	project := &domain.Project{
		ID:             "test-project",
		State:          domain.StateActive,
		CreatedAt:      now.Add(-2 * time.Hour),
		LastActivityAt: now.Add(-1 * time.Hour), // 1 hour ago - would be waiting if enabled
	}

	// Even with long inactivity, disabled threshold means not waiting
	if detector.IsWaiting(context.Background(), project) {
		t.Errorf("IsWaiting() with threshold=0 should return false")
	}
}

func TestIsWaiting_NewlyAddedProject(t *testing.T) {
	resolver := newMockResolver(10)

	detector := NewWaitingDetector(resolver)

	// Newly added project: LastActivityAt == CreatedAt
	now := time.Now()
	project := &domain.Project{
		ID:             "new-project",
		State:          domain.StateActive,
		CreatedAt:      now,
		LastActivityAt: now, // Same as CreatedAt - no activity observed yet
	}

	// Mock time to be 15 minutes later
	detector.now = func() time.Time {
		return now.Add(15 * time.Minute)
	}

	if detector.IsWaiting(context.Background(), project) {
		t.Errorf("IsWaiting() for newly added project should return false")
	}
}

func TestIsWaiting_HibernatedProject(t *testing.T) {
	resolver := newMockResolver(10)

	detector := NewWaitingDetector(resolver)

	now := time.Now()
	project := &domain.Project{
		ID:             "hibernated-project",
		State:          domain.StateHibernated, // Hibernated
		CreatedAt:      now.Add(-2 * time.Hour),
		LastActivityAt: now.Add(-1 * time.Hour), // 1 hour inactive
	}

	if detector.IsWaiting(context.Background(), project) {
		t.Errorf("IsWaiting() for hibernated project should return false")
	}
}

func TestIsWaiting_BoundaryPrecision(t *testing.T) {
	resolver := newMockResolver(10)

	detector := NewWaitingDetector(resolver)

	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		lastActivityAgo time.Duration
		expected        bool
	}{
		{"under threshold 9m59s", 9*time.Minute + 59*time.Second, false},
		{"at threshold 10m", 10 * time.Minute, true},
		{"over threshold 10m01s", 10*time.Minute + 1*time.Second, true},
		{"well over threshold 1h", 1 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &domain.Project{
				ID:             "test-project",
				State:          domain.StateActive,
				CreatedAt:      baseTime.Add(-2 * time.Hour), // Created 2 hours before activity
				LastActivityAt: baseTime,
			}

			// Mock time to be lastActivityAgo after baseTime
			detector.now = func() time.Time {
				return baseTime.Add(tt.lastActivityAgo)
			}

			result := detector.IsWaiting(context.Background(), project)
			if result != tt.expected {
				t.Errorf("IsWaiting() at %v = %v, want %v", tt.lastActivityAgo, result, tt.expected)
			}
		})
	}
}

func TestIsWaiting_PerProjectThreshold(t *testing.T) {
	resolver := newMockResolver(10)                   // Global threshold
	resolver.setProjectThreshold("custom-project", 5) // Per-project override

	detector := NewWaitingDetector(resolver)
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Project with custom threshold
	customProject := &domain.Project{
		ID:             "custom-project",
		State:          domain.StateActive,
		CreatedAt:      baseTime.Add(-1 * time.Hour),
		LastActivityAt: baseTime,
	}

	// Project using global threshold
	globalProject := &domain.Project{
		ID:             "global-project",
		State:          domain.StateActive,
		CreatedAt:      baseTime.Add(-1 * time.Hour),
		LastActivityAt: baseTime,
	}

	// Mock time to be 6 minutes after last activity
	detector.now = func() time.Time {
		return baseTime.Add(6 * time.Minute)
	}

	// Custom project (5 min threshold) should be waiting at 6 minutes
	if !detector.IsWaiting(context.Background(), customProject) {
		t.Errorf("IsWaiting() for custom project (5min threshold) at 6min should return true")
	}

	// Global project (10 min threshold) should NOT be waiting at 6 minutes
	if detector.IsWaiting(context.Background(), globalProject) {
		t.Errorf("IsWaiting() for global project (10min threshold) at 6min should return false")
	}
}

func TestIsWaiting_ActiveProjectWithInactivity(t *testing.T) {
	resolver := newMockResolver(10)

	detector := NewWaitingDetector(resolver)

	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	project := &domain.Project{
		ID:             "active-project",
		State:          domain.StateActive,
		CreatedAt:      baseTime.Add(-2 * time.Hour), // Created 2 hours before activity
		LastActivityAt: baseTime,                     // Activity was observed
	}

	// Mock time to be 15 minutes after last activity
	detector.now = func() time.Time {
		return baseTime.Add(15 * time.Minute)
	}

	if !detector.IsWaiting(context.Background(), project) {
		t.Errorf("IsWaiting() for active project with 15min inactivity should return true")
	}
}

func TestWaitingDuration_NotWaiting(t *testing.T) {
	resolver := newMockResolver(10)

	detector := NewWaitingDetector(resolver)

	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	project := &domain.Project{
		ID:             "active-project",
		State:          domain.StateActive,
		CreatedAt:      baseTime.Add(-1 * time.Hour),
		LastActivityAt: baseTime,
	}

	// Mock time to be 5 minutes after last activity (under threshold)
	detector.now = func() time.Time {
		return baseTime.Add(5 * time.Minute)
	}

	duration := detector.WaitingDuration(context.Background(), project)
	if duration != 0 {
		t.Errorf("WaitingDuration() when not waiting should return 0, got %v", duration)
	}
}

func TestWaitingDuration_WhenWaiting(t *testing.T) {
	resolver := newMockResolver(10)

	detector := NewWaitingDetector(resolver)

	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	project := &domain.Project{
		ID:             "waiting-project",
		State:          domain.StateActive,
		CreatedAt:      baseTime.Add(-1 * time.Hour),
		LastActivityAt: baseTime,
	}

	// Mock time to be 25 minutes after last activity (over threshold)
	detector.now = func() time.Time {
		return baseTime.Add(25 * time.Minute)
	}

	duration := detector.WaitingDuration(context.Background(), project)
	expected := 25 * time.Minute
	if duration != expected {
		t.Errorf("WaitingDuration() when waiting = %v, want %v", duration, expected)
	}
}

func TestWaitingDuration_NilProject(t *testing.T) {
	resolver := newMockResolver(10)
	detector := NewWaitingDetector(resolver)

	// Must not panic, should return 0
	duration := detector.WaitingDuration(context.Background(), nil)
	if duration != 0 {
		t.Errorf("WaitingDuration(nil) = %v, want 0", duration)
	}
}

func TestIsWaiting_TableDriven(t *testing.T) {
	tests := []struct {
		name             string
		lastActivityAgo  time.Duration
		createdAgo       time.Duration
		state            domain.ProjectState
		thresholdMinutes int
		expected         bool
	}{
		// Boundary tests
		{"under threshold 9m59s", 9*time.Minute + 59*time.Second, 1 * time.Hour, domain.StateActive, 10, false},
		{"at threshold 10m", 10 * time.Minute, 1 * time.Hour, domain.StateActive, 10, true},
		{"over threshold 10m01s", 10*time.Minute + 1*time.Second, 1 * time.Hour, domain.StateActive, 10, true},

		// Edge cases
		{"newly added project", 5 * time.Minute, 5 * time.Minute, domain.StateActive, 10, false}, // CreatedAt == LastActivityAt
		{"hibernated project", 1 * time.Hour, 2 * time.Hour, domain.StateHibernated, 10, false},
		{"threshold disabled", 1 * time.Hour, 2 * time.Hour, domain.StateActive, 0, false},

		// Various thresholds
		{"custom 5min threshold active", 6 * time.Minute, 1 * time.Hour, domain.StateActive, 5, true},
		{"under custom 5min threshold", 4 * time.Minute, 1 * time.Hour, domain.StateActive, 5, false},
		{"1min threshold at boundary", 1 * time.Minute, 1 * time.Hour, domain.StateActive, 1, true},
		{"1min threshold under", 59 * time.Second, 1 * time.Hour, domain.StateActive, 1, false},

		// Long inactivity (vacation scenario)
		{"7 days inactivity", 7 * 24 * time.Hour, 14 * 24 * time.Hour, domain.StateActive, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := newMockResolver(tt.thresholdMinutes)

			detector := NewWaitingDetector(resolver)

			baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

			// Setup project with appropriate timestamps:
			// - CreatedAt is fixed at baseTime minus createdAgo
			// - LastActivityAt is baseTime (so "now" minus lastActivityAgo = correct inactivity)
			// - Special case: newly added projects have LastActivityAt == CreatedAt
			createdAt := baseTime.Add(-tt.createdAgo)
			lastActivityAt := baseTime // Will result in correct inactivity when now = baseTime + lastActivityAgo

			// For "newly added" test case: LastActivityAt == CreatedAt (no activity observed yet)
			if tt.createdAgo == tt.lastActivityAgo {
				lastActivityAt = createdAt
			}

			project := &domain.Project{
				ID:             "test-project",
				State:          tt.state,
				CreatedAt:      createdAt,
				LastActivityAt: lastActivityAt,
			}

			// Mock time
			detector.now = func() time.Time {
				return baseTime.Add(tt.lastActivityAgo)
			}

			result := detector.IsWaiting(context.Background(), project)
			if result != tt.expected {
				t.Errorf("IsWaiting() = %v, want %v", result, tt.expected)
			}
		})
	}
}
