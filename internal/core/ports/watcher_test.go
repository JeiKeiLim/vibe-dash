package ports_test

import (
	"context"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

func TestFileOperation_String(t *testing.T) {
	tests := []struct {
		name     string
		op       ports.FileOperation
		expected string
	}{
		{"Create", ports.FileOpCreate, "create"},
		{"Modify", ports.FileOpModify, "modify"},
		{"Delete", ports.FileOpDelete, "delete"},
		{"Unknown negative", ports.FileOperation(-1), "unknown"},
		{"Unknown large", ports.FileOperation(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.op.String()
			if result != tt.expected {
				t.Errorf("FileOperation(%d).String() = %q, want %q", tt.op, result, tt.expected)
			}
		})
	}
}

func TestFileOperation_Valid(t *testing.T) {
	tests := []struct {
		name     string
		op       ports.FileOperation
		expected bool
	}{
		{"Create is valid", ports.FileOpCreate, true},
		{"Modify is valid", ports.FileOpModify, true},
		{"Delete is valid", ports.FileOpDelete, true},
		{"Negative is invalid", ports.FileOperation(-1), false},
		{"Large is invalid", ports.FileOperation(100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.op.Valid()
			if result != tt.expected {
				t.Errorf("FileOperation(%d).Valid() = %v, want %v", tt.op, result, tt.expected)
			}
		})
	}
}

func TestFileEvent_Struct(t *testing.T) {
	now := time.Now()
	event := ports.FileEvent{
		Path:      "/test/path/file.go",
		Operation: ports.FileOpModify,
		Timestamp: now,
	}

	if event.Path != "/test/path/file.go" {
		t.Errorf("FileEvent.Path = %q, want %q", event.Path, "/test/path/file.go")
	}
	if event.Operation != ports.FileOpModify {
		t.Errorf("FileEvent.Operation = %v, want %v", event.Operation, ports.FileOpModify)
	}
	if !event.Timestamp.Equal(now) {
		t.Errorf("FileEvent.Timestamp = %v, want %v", event.Timestamp, now)
	}
}

// mockWatcher verifies interface compliance at compile time
type mockWatcher struct {
	events chan ports.FileEvent
	closed bool
}

func newMockWatcher() *mockWatcher {
	return &mockWatcher{
		events: make(chan ports.FileEvent, 10),
	}
}

func (m *mockWatcher) Watch(ctx context.Context, paths []string) (<-chan ports.FileEvent, error) {
	return m.events, nil
}

func (m *mockWatcher) Close() error {
	if !m.closed {
		close(m.events)
		m.closed = true
	}
	return nil
}

// Compile-time interface compliance check
var _ ports.FileWatcher = (*mockWatcher)(nil)

func TestFileWatcher_InterfaceCompliance(t *testing.T) {
	var watcher ports.FileWatcher = newMockWatcher()

	t.Run("Watch returns event channel", func(t *testing.T) {
		ctx := context.Background()
		paths := []string{"/test/path1", "/test/path2"}

		events, err := watcher.Watch(ctx, paths)
		if err != nil {
			t.Fatalf("Watch() error = %v, want nil", err)
		}
		if events == nil {
			t.Fatal("Watch() returned nil channel")
		}
	})

	t.Run("Close returns no error", func(t *testing.T) {
		err := watcher.Close()
		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
	})

	t.Run("Watch accepts cancelled context without panic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		freshWatcher := newMockWatcher()
		_, err := freshWatcher.Watch(ctx, []string{"/path"})
		// NOTE: This test only verifies the method signature accepts context.
		// Actual cancellation behavior (closing channel, stopping watch)
		// is tested at the adapter implementation level, not interface level.
		// Interface tests verify the contract shape; adapter tests verify behavior.
		_ = err
	})
}
