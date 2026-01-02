package ports_test

import (
	"context"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// mockStateActivator verifies the StateActivator interface contract.
type mockStateActivator struct {
	hibernateCalls []string
	activateCalls  []string
}

func (m *mockStateActivator) Hibernate(ctx context.Context, projectID string) error {
	m.hibernateCalls = append(m.hibernateCalls, projectID)
	return nil
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

	// Test Activate
	err := activator.Activate(context.Background(), "test-project-id")
	if err != nil {
		t.Errorf("unexpected Activate error: %v", err)
	}

	mock := activator.(*mockStateActivator)
	if len(mock.activateCalls) != 1 || mock.activateCalls[0] != "test-project-id" {
		t.Error("Activate should track calls with project ID")
	}

	// Test Hibernate (M1: Added in code review - both methods should be tested)
	err = activator.Hibernate(context.Background(), "hibernate-project-id")
	if err != nil {
		t.Errorf("unexpected Hibernate error: %v", err)
	}
	if len(mock.hibernateCalls) != 1 || mock.hibernateCalls[0] != "hibernate-project-id" {
		t.Error("Hibernate should track calls with project ID")
	}
}
