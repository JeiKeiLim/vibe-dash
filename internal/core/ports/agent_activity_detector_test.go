package ports

import (
	"context"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// mockAgentActivityDetector is a minimal implementation for interface documentation tests.
type mockAgentActivityDetector struct {
	name  string
	state domain.AgentState
	err   error
}

func (m *mockAgentActivityDetector) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
	return m.state, m.err
}

func (m *mockAgentActivityDetector) Name() string {
	return m.name
}

// TestAgentActivityDetector_InterfaceContract validates that the interface
// contract can be fulfilled by a simple implementation.
func TestAgentActivityDetector_InterfaceContract(t *testing.T) {
	// Verify interface can be implemented
	var detector AgentActivityDetector = &mockAgentActivityDetector{
		name: "Test Detector",
		state: domain.NewAgentState(
			"Test Detector",
			domain.AgentWaitingForUser,
			5*time.Minute,
			domain.ConfidenceCertain,
		),
	}

	// Test Name() method
	if got := detector.Name(); got != "Test Detector" {
		t.Errorf("Name() = %q, want %q", got, "Test Detector")
	}

	// Test Detect() method
	state, err := detector.Detect(context.Background(), "/test/path")
	if err != nil {
		t.Errorf("Detect() error = %v, want nil", err)
	}
	if state.Tool != "Test Detector" {
		t.Errorf("Detect().Tool = %q, want %q", state.Tool, "Test Detector")
	}
	if state.Status != domain.AgentWaitingForUser {
		t.Errorf("Detect().Status = %v, want %v", state.Status, domain.AgentWaitingForUser)
	}
}

// TestAgentActivityDetector_ContextCancellation documents that implementations
// must respect context cancellation. This test demonstrates the EXPECTED usage
// pattern - real implementations MUST check ctx.Err() and return context.Canceled
// or context.DeadlineExceeded when the context is cancelled.
//
// Note: The mock does NOT validate cancellation (it's a simple mock).
// Integration tests for real implementations should verify proper cancellation behavior.
func TestAgentActivityDetector_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	detector := &mockAgentActivityDetector{
		name:  "Test",
		state: domain.AgentState{Status: domain.AgentUnknown},
	}

	// This demonstrates the expected usage pattern.
	// Real implementations MUST check ctx.Done() and return ctx.Err() when cancelled.
	state, _ := detector.Detect(ctx, "/test")

	// Mock returns normally even with cancelled context.
	// Real implementations would return (AgentState{}, ctx.Err()).
	if state.Status != domain.AgentUnknown {
		t.Errorf("Detect().Status = %v, expected AgentUnknown from mock", state.Status)
	}
}

// TestAgentActivityDetector_DetectReturnsAgentState documents the expected
// AgentState fields when detection succeeds.
func TestAgentActivityDetector_DetectReturnsAgentState(t *testing.T) {
	detector := &mockAgentActivityDetector{
		name: "Claude Code",
		state: domain.NewAgentState(
			"Claude Code",
			domain.AgentWorking,
			10*time.Minute,
			domain.ConfidenceCertain,
		),
	}

	state, err := detector.Detect(context.Background(), "/project")
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// Verify all AgentState fields are populated
	if state.Tool == "" {
		t.Error("Tool should not be empty")
	}
	if state.Duration == 0 {
		t.Error("Duration should be set when agent is active")
	}
	if state.Confidence == domain.ConfidenceUncertain && state.Status != domain.AgentUnknown {
		t.Error("Non-unknown status should have non-uncertain confidence")
	}
}

// TestAgentActivityDetector_ErrorHandling documents error scenarios.
func TestAgentActivityDetector_ErrorHandling(t *testing.T) {
	detector := &mockAgentActivityDetector{
		name:  "Failing Detector",
		state: domain.AgentState{Status: domain.AgentUnknown},
		err:   domain.ErrPathNotAccessible,
	}

	_, err := detector.Detect(context.Background(), "/nonexistent")
	if err == nil {
		t.Error("Detect() should return error for inaccessible path")
	}
}
