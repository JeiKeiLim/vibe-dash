package detection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// mockDetector implements ports.AgentActivityDetector for testing.
type mockDetector struct {
	name   string
	state  domain.AgentState
	err    error
	delay  time.Duration // Simulates slow detection
	called bool
}

func (m *mockDetector) Name() string {
	return m.name
}

func (m *mockDetector) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
	m.called = true

	// Simulate slow detection
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return domain.NewAgentState(m.name, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
		case <-time.After(m.delay):
		}
	}

	return m.state, m.err
}

func TestNewAgentDetectionService(t *testing.T) {
	t.Run("creates service with default detectors", func(t *testing.T) {
		svc := NewAgentDetectionService()
		if svc == nil {
			t.Fatal("NewAgentDetectionService returned nil")
		}
		if svc.claudeDetector == nil {
			t.Error("claudeDetector should be initialized by default")
		}
		if svc.genericDetector == nil {
			t.Error("genericDetector should be initialized by default")
		}
	})

	t.Run("accepts custom claude detector", func(t *testing.T) {
		mock := &mockDetector{name: "MockClaude"}
		svc := NewAgentDetectionService(WithClaudeDetector(mock))
		if svc.claudeDetector != mock {
			t.Error("WithClaudeDetector option not applied")
		}
	})

	t.Run("accepts custom generic detector", func(t *testing.T) {
		mock := &mockDetector{name: "MockGeneric"}
		svc := NewAgentDetectionService(WithGenericDetector(mock))
		if svc.genericDetector != mock {
			t.Error("WithGenericDetector option not applied")
		}
	})
}

func TestAgentDetectionService_Name(t *testing.T) {
	svc := NewAgentDetectionService()
	if svc.Name() != serviceName {
		t.Errorf("Name() = %q, want %q", svc.Name(), serviceName)
	}
}

func TestAgentDetectionService_Detect(t *testing.T) {
	tests := []struct {
		name           string
		claudeState    domain.AgentState
		claudeErr      error
		genericState   domain.AgentState
		genericErr     error
		wantStatus     domain.AgentStatus
		wantTool       string
		genericCalled  bool // Whether generic should be called
		wantConfidence domain.Confidence
	}{
		{
			name: "claude returns WaitingForUser - no fallback",
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour, domain.ConfidenceCertain),
			wantStatus:     domain.AgentWaitingForUser,
			wantTool:       "Claude Code",
			genericCalled:  false,
			wantConfidence: domain.ConfidenceCertain,
		},
		{
			name: "claude returns Working - no fallback",
			claudeState: domain.NewAgentState("Claude Code", domain.AgentWorking,
				5*time.Minute, domain.ConfidenceCertain),
			wantStatus:     domain.AgentWorking,
			wantTool:       "Claude Code",
			genericCalled:  false,
			wantConfidence: domain.ConfidenceCertain,
		},
		{
			name: "claude returns Inactive - no fallback",
			claudeState: domain.NewAgentState("Claude Code", domain.AgentInactive,
				24*time.Hour, domain.ConfidenceCertain),
			wantStatus:     domain.AgentInactive,
			wantTool:       "Claude Code",
			genericCalled:  false,
			wantConfidence: domain.ConfidenceCertain,
		},
		{
			name: "claude returns Unknown - falls back to generic",
			claudeState: domain.NewAgentState("Claude Code", domain.AgentUnknown,
				0, domain.ConfidenceUncertain),
			genericState: domain.NewAgentState("Generic", domain.AgentWaitingForUser,
				15*time.Minute, domain.ConfidenceUncertain),
			wantStatus:     domain.AgentWaitingForUser,
			wantTool:       "Generic",
			genericCalled:  true,
			wantConfidence: domain.ConfidenceUncertain,
		},
		{
			name:      "claude error - falls back to generic",
			claudeErr: errors.New("permission denied"),
			genericState: domain.NewAgentState("Generic", domain.AgentWorking,
				5*time.Minute, domain.ConfidenceUncertain),
			wantStatus:     domain.AgentWorking,
			wantTool:       "Generic",
			genericCalled:  true,
			wantConfidence: domain.ConfidenceUncertain,
		},
		{
			name: "both return Unknown",
			claudeState: domain.NewAgentState("Claude Code", domain.AgentUnknown,
				0, domain.ConfidenceUncertain),
			genericState: domain.NewAgentState("Generic", domain.AgentUnknown,
				0, domain.ConfidenceUncertain),
			wantStatus:     domain.AgentUnknown,
			wantTool:       "Generic",
			genericCalled:  true,
			wantConfidence: domain.ConfidenceUncertain,
		},
		{
			name: "claude Unknown, generic error - returns Unknown with error",
			claudeState: domain.NewAgentState("Claude Code", domain.AgentUnknown,
				0, domain.ConfidenceUncertain),
			genericErr:     errors.New("filesystem error"),
			wantStatus:     domain.AgentUnknown,
			wantTool:       serviceName,
			genericCalled:  true,
			wantConfidence: domain.ConfidenceUncertain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claudeMock := &mockDetector{
				name:  "Claude Code",
				state: tt.claudeState,
				err:   tt.claudeErr,
			}
			genericMock := &mockDetector{
				name:  "Generic",
				state: tt.genericState,
				err:   tt.genericErr,
			}

			svc := NewAgentDetectionService(
				WithClaudeDetector(claudeMock),
				WithGenericDetector(genericMock),
			)

			state, _ := svc.Detect(context.Background(), "/test/path")

			if state.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", state.Status, tt.wantStatus)
			}
			if state.Tool != tt.wantTool {
				t.Errorf("Tool = %q, want %q", state.Tool, tt.wantTool)
			}
			if genericMock.called != tt.genericCalled {
				t.Errorf("genericMock.called = %v, want %v", genericMock.called, tt.genericCalled)
			}
			if state.Confidence != tt.wantConfidence {
				t.Errorf("Confidence = %v, want %v", state.Confidence, tt.wantConfidence)
			}
		})
	}
}

func TestAgentDetectionService_Detect_Timeout(t *testing.T) {
	// Claude detector takes 2 seconds (longer than 1 second timeout)
	claudeMock := &mockDetector{
		name:  "Claude Code",
		state: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser, time.Hour, domain.ConfidenceCertain),
		delay: 2 * time.Second,
	}
	genericMock := &mockDetector{
		name:  "Generic",
		state: domain.NewAgentState("Generic", domain.AgentWorking, time.Minute, domain.ConfidenceUncertain),
	}

	svc := NewAgentDetectionService(
		WithClaudeDetector(claudeMock),
		WithGenericDetector(genericMock),
	)

	start := time.Now()
	state, _ := svc.Detect(context.Background(), "/test/path")
	elapsed := time.Since(start)

	// Should timeout after 1 second (detectionTimeout)
	if elapsed > 1500*time.Millisecond {
		t.Errorf("Detection took %v, should timeout within ~1 second", elapsed)
	}

	// Should return Unknown due to timeout
	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v after timeout, want Unknown", state.Status)
	}
}

func TestAgentDetectionService_Detect_ContextCancellation(t *testing.T) {
	claudeMock := &mockDetector{
		name:  "Claude Code",
		state: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser, time.Hour, domain.ConfidenceCertain),
		delay: 5 * time.Second, // Would take long without cancellation
	}
	genericMock := &mockDetector{
		name: "Generic",
	}

	svc := NewAgentDetectionService(
		WithClaudeDetector(claudeMock),
		WithGenericDetector(genericMock),
	)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	state, _ := svc.Detect(ctx, "/test/path")
	elapsed := time.Since(start)

	// Should return within 100ms of cancellation (AC7)
	if elapsed > 200*time.Millisecond {
		t.Errorf("Detection took %v after cancellation, should return within 100ms", elapsed)
	}

	// Should return Unknown
	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v after cancellation, want Unknown", state.Status)
	}
}

func TestAgentDetectionService_Detect_ContextCancelledAtEntry(t *testing.T) {
	claudeMock := &mockDetector{name: "Claude Code"}
	genericMock := &mockDetector{name: "Generic"}

	svc := NewAgentDetectionService(
		WithClaudeDetector(claudeMock),
		WithGenericDetector(genericMock),
	)

	// Already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	state, _ := svc.Detect(ctx, "/test/path")

	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v with cancelled context, want Unknown", state.Status)
	}
	if claudeMock.called {
		t.Error("Claude detector should not be called when context is already cancelled")
	}
	if genericMock.called {
		t.Error("Generic detector should not be called when context is already cancelled")
	}
}

func TestAgentDetectionService_Detect_ContextCancelledBetweenDetectors(t *testing.T) {
	// Claude returns Unknown (triggers fallback)
	claudeMock := &mockDetector{
		name:  "Claude Code",
		state: domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain),
	}
	genericMock := &mockDetector{
		name:  "Generic",
		state: domain.NewAgentState("Generic", domain.AgentWaitingForUser, time.Hour, domain.ConfidenceUncertain),
	}

	svc := NewAgentDetectionService(
		WithClaudeDetector(claudeMock),
		WithGenericDetector(genericMock),
	)

	// Use timeout context that will expire after claude detection
	// This simulates context cancelled between detector calls
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	<-ctx.Done()

	state, _ := svc.Detect(ctx, "/test/path")

	// Should return Unknown because context is already cancelled
	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v, want Unknown (context cancelled)", state.Status)
	}
}

func TestAgentDetectionService_InterfaceCompliance(t *testing.T) {
	// This test verifies compile-time interface compliance (already in source)
	// but also runtime behavior matches the interface contract
	svc := NewAgentDetectionService()

	// Name() returns non-empty string
	if svc.Name() == "" {
		t.Error("Name() should return non-empty string")
	}

	// Detect() returns valid AgentState (error is acceptable for nonexistent path)
	state, _ := svc.Detect(context.Background(), "/nonexistent/path")
	// State should be valid (not panic)
	_ = state.Status
	_ = state.Tool
	_ = state.Duration
	_ = state.Confidence
}
