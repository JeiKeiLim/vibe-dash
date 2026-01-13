package logreaders

import (
	"context"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// mockLogReader implements ports.LogReader for testing
type mockLogReader struct {
	name    string
	canRead bool
}

var _ ports.LogReader = (*mockLogReader)(nil)

func (m *mockLogReader) Tool() string {
	return m.name
}

func (m *mockLogReader) CanRead(_ context.Context, _ string) bool {
	return m.canRead
}

func (m *mockLogReader) ListSessions(_ context.Context, _ string) ([]domain.LogSession, error) {
	return nil, nil
}

func (m *mockLogReader) ReadSession(_ context.Context, _ string) ([]domain.LogEntry, error) {
	return nil, nil
}

func (m *mockLogReader) TailSession(_ context.Context, _ string) (<-chan domain.LogEntry, error) {
	return nil, nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if len(registry.Readers()) != 0 {
		t.Errorf("new registry should have no readers, got %d", len(registry.Readers()))
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()
	reader1 := &mockLogReader{name: "reader1", canRead: true}
	reader2 := &mockLogReader{name: "reader2", canRead: false}

	registry.Register(reader1)
	registry.Register(reader2)

	readers := registry.Readers()
	if len(readers) != 2 {
		t.Errorf("expected 2 readers, got %d", len(readers))
	}
	if readers[0].Tool() != "reader1" {
		t.Errorf("first reader should be reader1, got %s", readers[0].Tool())
	}
	if readers[1].Tool() != "reader2" {
		t.Errorf("second reader should be reader2, got %s", readers[1].Tool())
	}
}

func TestRegistryGetReader(t *testing.T) {
	tests := []struct {
		name       string
		readers    []*mockLogReader
		wantReader string
		wantNil    bool
	}{
		{
			name:    "no readers registered",
			readers: nil,
			wantNil: true,
		},
		{
			name: "first matching reader returned",
			readers: []*mockLogReader{
				{name: "reader1", canRead: false},
				{name: "reader2", canRead: true},
				{name: "reader3", canRead: true},
			},
			wantReader: "reader2",
		},
		{
			name: "no matching reader",
			readers: []*mockLogReader{
				{name: "reader1", canRead: false},
				{name: "reader2", canRead: false},
			},
			wantNil: true,
		},
		{
			name: "single matching reader",
			readers: []*mockLogReader{
				{name: "reader1", canRead: true},
			},
			wantReader: "reader1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()
			for _, r := range tt.readers {
				registry.Register(r)
			}

			ctx := context.Background()
			reader := registry.GetReader(ctx, "/some/project")

			if tt.wantNil {
				if reader != nil {
					t.Errorf("expected nil reader, got %s", reader.Tool())
				}
				return
			}

			if reader == nil {
				t.Fatal("expected non-nil reader, got nil")
			}
			if reader.Tool() != tt.wantReader {
				t.Errorf("expected reader %s, got %s", tt.wantReader, reader.Tool())
			}
		})
	}
}

func TestRegistryGetReaderContextCancellation(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&mockLogReader{name: "reader1", canRead: true})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	reader := registry.GetReader(ctx, "/some/project")
	if reader != nil {
		t.Error("expected nil when context is cancelled")
	}
}
