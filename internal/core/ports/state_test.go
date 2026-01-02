package ports_test

import (
	"context"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// mockStateActivator verifies the StateActivator interface contract.
type mockStateActivator struct {
	activateCalls []string
}

func (m *mockStateActivator) Activate(ctx context.Context, projectID string) error {
	m.activateCalls = append(m.activateCalls, projectID)
	return nil
}

// Compile-time interface compliance check
var _ ports.StateActivator = (*mockStateActivator)(nil)

func TestStateActivator_Interface(t *testing.T) {
	// Verify the interface can be implemented and used
	var activator ports.StateActivator = &mockStateActivator{}

	err := activator.Activate(context.Background(), "test-project-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mock := activator.(*mockStateActivator)
	if len(mock.activateCalls) != 1 || mock.activateCalls[0] != "test-project-id" {
		t.Error("Activate should track calls with project ID")
	}
}
