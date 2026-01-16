package detection

import (
	"context"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestNewAgentWaitingAdapter(t *testing.T) {
	svc := NewAgentDetectionService()
	adapter := NewAgentWaitingAdapter(svc)

	if adapter == nil {
		t.Fatal("NewAgentWaitingAdapter returned nil")
	}
	if adapter.service != svc {
		t.Error("service not set correctly")
	}
	if adapter.cache == nil {
		t.Error("cache should be initialized")
	}
	if adapter.now == nil {
		t.Error("now function should be initialized")
	}
}

func TestAgentWaitingAdapter_IsWaiting(t *testing.T) {
	tests := []struct {
		name        string
		project     *domain.Project
		claudeState domain.AgentState
		want        bool
	}{
		{
			name:    "nil project returns false",
			project: nil,
			want:    false,
		},
		{
			name: "hibernated project returns false",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateHibernated,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour, domain.ConfidenceCertain),
			want: false, // Hibernated projects never show as waiting
		},
		{
			name: "WaitingForUser status returns true",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateActive,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour, domain.ConfidenceCertain),
			want: true,
		},
		{
			name: "Working status returns false",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateActive,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWorking,
				5*time.Minute, domain.ConfidenceCertain),
			want: false,
		},
		{
			name: "Inactive status returns false",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateActive,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentInactive,
				24*time.Hour, domain.ConfidenceCertain),
			want: false,
		},
		{
			name: "Unknown status returns false",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateActive,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentUnknown,
				0, domain.ConfidenceUncertain),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claudeMock := &mockDetector{
				name:  "Claude Code",
				state: tt.claudeState,
			}
			genericMock := &mockDetector{
				name:  "Generic",
				state: domain.NewAgentState("Generic", domain.AgentUnknown, 0, domain.ConfidenceUncertain),
			}

			svc := NewAgentDetectionService(
				WithClaudeDetector(claudeMock),
				WithGenericDetector(genericMock),
			)
			adapter := NewAgentWaitingAdapter(svc)

			got := adapter.IsWaiting(context.Background(), tt.project)
			if got != tt.want {
				t.Errorf("IsWaiting() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentWaitingAdapter_WaitingDuration(t *testing.T) {
	tests := []struct {
		name         string
		project      *domain.Project
		claudeState  domain.AgentState
		wantDuration time.Duration
	}{
		{
			name:         "nil project returns 0",
			project:      nil,
			wantDuration: 0,
		},
		{
			name: "hibernated project returns 0",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateHibernated,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour, domain.ConfidenceCertain),
			wantDuration: 0,
		},
		{
			name: "WaitingForUser returns duration",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateActive,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour, domain.ConfidenceCertain),
			wantDuration: 2 * time.Hour,
		},
		{
			name: "Working returns 0",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateActive,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWorking,
				5*time.Minute, domain.ConfidenceCertain),
			wantDuration: 0,
		},
		{
			name: "Unknown returns 0",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateActive,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentUnknown,
				0, domain.ConfidenceUncertain),
			wantDuration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claudeMock := &mockDetector{
				name:  "Claude Code",
				state: tt.claudeState,
			}
			genericMock := &mockDetector{
				name:  "Generic",
				state: domain.NewAgentState("Generic", domain.AgentUnknown, 0, domain.ConfidenceUncertain),
			}

			svc := NewAgentDetectionService(
				WithClaudeDetector(claudeMock),
				WithGenericDetector(genericMock),
			)
			adapter := NewAgentWaitingAdapter(svc)

			got := adapter.WaitingDuration(context.Background(), tt.project)
			if got != tt.wantDuration {
				t.Errorf("WaitingDuration() = %v, want %v", got, tt.wantDuration)
			}
		})
	}
}

func TestAgentWaitingAdapter_Cache(t *testing.T) {
	t.Run("cache prevents repeated detection within TTL", func(t *testing.T) {
		claudeMock := &mockDetector{
			name: "Claude Code",
			state: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour, domain.ConfidenceCertain),
		}
		genericMock := &mockDetector{
			name:  "Generic",
			state: domain.NewAgentState("Generic", domain.AgentUnknown, 0, domain.ConfidenceUncertain),
		}

		svc := NewAgentDetectionService(
			WithClaudeDetector(claudeMock),
			WithGenericDetector(genericMock),
		)
		adapter := NewAgentWaitingAdapter(svc)

		project := &domain.Project{
			Path:  "/test/path",
			State: domain.StateActive,
		}

		// First call
		result1 := adapter.IsWaiting(context.Background(), project)
		if !result1 {
			t.Error("First call should return true")
		}

		// Change mock state - but cache should return old value
		claudeMock.state = domain.NewAgentState("Claude Code", domain.AgentWorking,
			time.Minute, domain.ConfidenceCertain)

		// Second call within TTL - should still return cached value (true)
		result2 := adapter.IsWaiting(context.Background(), project)
		if !result2 {
			t.Error("Second call within cache TTL should return cached value (true)")
		}
	})

	t.Run("cache expires after TTL", func(t *testing.T) {
		claudeMock := &mockDetector{
			name: "Claude Code",
			state: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour, domain.ConfidenceCertain),
		}
		genericMock := &mockDetector{
			name:  "Generic",
			state: domain.NewAgentState("Generic", domain.AgentUnknown, 0, domain.ConfidenceUncertain),
		}

		svc := NewAgentDetectionService(
			WithClaudeDetector(claudeMock),
			WithGenericDetector(genericMock),
		)
		adapter := NewAgentWaitingAdapter(svc)

		// Control time for testing
		currentTime := time.Now()
		adapter.now = func() time.Time { return currentTime }

		project := &domain.Project{
			Path:  "/test/path",
			State: domain.StateActive,
		}

		// First call - caches result
		result1 := adapter.IsWaiting(context.Background(), project)
		if !result1 {
			t.Error("First call should return true")
		}

		// Change mock state
		claudeMock.state = domain.NewAgentState("Claude Code", domain.AgentWorking,
			time.Minute, domain.ConfidenceCertain)

		// Advance time past TTL (5 seconds)
		currentTime = currentTime.Add(6 * time.Second)

		// Third call after TTL - should detect again
		result3 := adapter.IsWaiting(context.Background(), project)
		if result3 {
			t.Error("Call after cache TTL should return new value (false)")
		}
	})
}

func TestAgentWaitingAdapter_ClearCache(t *testing.T) {
	claudeMock := &mockDetector{
		name: "Claude Code",
		state: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
			2*time.Hour, domain.ConfidenceCertain),
	}
	genericMock := &mockDetector{
		name:  "Generic",
		state: domain.NewAgentState("Generic", domain.AgentUnknown, 0, domain.ConfidenceUncertain),
	}

	svc := NewAgentDetectionService(
		WithClaudeDetector(claudeMock),
		WithGenericDetector(genericMock),
	)
	adapter := NewAgentWaitingAdapter(svc)

	project := &domain.Project{
		Path:  "/test/path",
		State: domain.StateActive,
	}

	// Populate cache
	_ = adapter.IsWaiting(context.Background(), project)

	// Verify cache has entry
	adapter.mu.RLock()
	cacheLen := len(adapter.cache)
	adapter.mu.RUnlock()
	if cacheLen == 0 {
		t.Error("Cache should have entry after IsWaiting call")
	}

	// Clear cache
	adapter.ClearCache()

	// Verify cache is empty
	adapter.mu.RLock()
	cacheLenAfter := len(adapter.cache)
	adapter.mu.RUnlock()
	if cacheLenAfter != 0 {
		t.Errorf("Cache should be empty after ClearCache, got %d entries", cacheLenAfter)
	}
}

func TestAgentWaitingAdapter_AgentState(t *testing.T) {
	tests := []struct {
		name           string
		project        *domain.Project
		claudeState    domain.AgentState
		wantStatus     domain.AgentStatus
		wantTool       string
		wantConfidence domain.Confidence
	}{
		{
			name:           "nil project returns empty state",
			project:        nil,
			wantStatus:     domain.AgentUnknown,
			wantTool:       "",
			wantConfidence: domain.ConfidenceUncertain,
		},
		{
			name: "hibernated project returns empty state",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateHibernated,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour, domain.ConfidenceCertain),
			wantStatus:     domain.AgentUnknown,
			wantTool:       "",
			wantConfidence: domain.ConfidenceUncertain,
		},
		{
			name: "active project returns full state",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateActive,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour, domain.ConfidenceCertain),
			wantStatus:     domain.AgentWaitingForUser,
			wantTool:       "Claude Code",
			wantConfidence: domain.ConfidenceCertain,
		},
		{
			name: "working state returns correct tool and confidence",
			project: &domain.Project{
				Path:  "/test/path",
				State: domain.StateActive,
			},
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWorking,
				5*time.Minute, domain.ConfidenceCertain),
			wantStatus:     domain.AgentWorking,
			wantTool:       "Claude Code",
			wantConfidence: domain.ConfidenceCertain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claudeMock := &mockDetector{
				name:  "Claude Code",
				state: tt.claudeState,
			}
			genericMock := &mockDetector{
				name:  "Generic",
				state: domain.NewAgentState("Generic", domain.AgentUnknown, 0, domain.ConfidenceUncertain),
			}

			svc := NewAgentDetectionService(
				WithClaudeDetector(claudeMock),
				WithGenericDetector(genericMock),
			)
			adapter := NewAgentWaitingAdapter(svc)

			got := adapter.AgentState(context.Background(), tt.project)
			if got.Status != tt.wantStatus {
				t.Errorf("AgentState().Status = %v, want %v", got.Status, tt.wantStatus)
			}
			if got.Tool != tt.wantTool {
				t.Errorf("AgentState().Tool = %v, want %v", got.Tool, tt.wantTool)
			}
			if got.Confidence != tt.wantConfidence {
				t.Errorf("AgentState().Confidence = %v, want %v", got.Confidence, tt.wantConfidence)
			}
		})
	}
}

func TestAgentWaitingAdapter_InterfaceCompliance(t *testing.T) {
	svc := NewAgentDetectionService()
	adapter := NewAgentWaitingAdapter(svc)

	// Test that adapter implements WaitingDetector interface
	// This is checked at compile time by var _ check in source,
	// but we verify runtime behavior here

	project := &domain.Project{
		Path:  "/nonexistent/path",
		State: domain.StateActive,
	}

	// IsWaiting should not panic
	_ = adapter.IsWaiting(context.Background(), project)

	// WaitingDuration should not panic
	_ = adapter.WaitingDuration(context.Background(), project)

	// AgentState should not panic (Story 15.7)
	_ = adapter.AgentState(context.Background(), project)
}
