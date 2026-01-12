// Package logreaders provides log reading implementations for agentic tools.
// It contains the Registry that coordinates all registered log readers and
// individual reader packages (e.g., claude_code).
//
// Thread Safety: Registry is NOT safe for concurrent modification.
// All Register() calls must complete before any GetReader() calls.
// Typical usage: register all readers during application initialization,
// then use GetReader() concurrently from multiple goroutines.
package logreaders

import (
	"context"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Registry manages registered log readers and coordinates access.
// It is the only component that knows about all log reader implementations.
// Services should call Registry.GetReader(), never individual readers.
type Registry struct {
	readers []ports.LogReader
}

// Compile-time interface compliance check
var _ ports.LogReaderRegistry = (*Registry)(nil)

// NewRegistry creates a new log reader registry with no registered readers.
func NewRegistry() *Registry {
	return &Registry{
		readers: make([]ports.LogReader, 0),
	}
}

// Register adds a log reader to the registry.
// Readers are checked in the order they are registered.
func (r *Registry) Register(reader ports.LogReader) {
	r.readers = append(r.readers, reader)
}

// Readers returns the list of registered log readers.
func (r *Registry) Readers() []ports.LogReader {
	return r.readers
}

// GetReader returns the first reader where CanRead() returns true,
// or nil if no reader can handle this project.
func (r *Registry) GetReader(ctx context.Context, projectPath string) ports.LogReader {
	for _, reader := range r.readers {
		// Check context before each reader
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if reader.CanRead(ctx, projectPath) {
			return reader
		}
	}
	return nil
}
