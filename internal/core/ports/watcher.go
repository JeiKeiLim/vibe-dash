package ports

import (
	"context"
	"time"
)

// FileOperation represents the type of file system event.
// Used to distinguish between file creation, modification, and deletion.
type FileOperation int

const (
	// FileOpCreate indicates a new file was created
	FileOpCreate FileOperation = iota
	// FileOpModify indicates an existing file was modified
	FileOpModify
	// FileOpDelete indicates a file was deleted
	FileOpDelete
)

// String returns the human-readable name of the file operation.
// Returns "unknown" for invalid operation values.
func (op FileOperation) String() string {
	switch op {
	case FileOpCreate:
		return "create"
	case FileOpModify:
		return "modify"
	case FileOpDelete:
		return "delete"
	default:
		return "unknown"
	}
}

// Valid returns true if the FileOperation is a known, valid value.
// Use this to validate operation values before processing.
func (op FileOperation) Valid() bool {
	return op >= FileOpCreate && op <= FileOpDelete
}

// FileEvent represents a file system change event.
// Events are emitted by FileWatcher when files in watched directories change.
type FileEvent struct {
	// Path is the canonical absolute path of the changed file
	Path string

	// Operation indicates what type of change occurred
	Operation FileOperation

	// Timestamp is when the event occurred (or was detected)
	Timestamp time.Time
}

// FileWatcher defines the interface for monitoring file system changes.
// Implementations watch specified directories and emit events when files change.
// This is a key interface for the "Agent Waiting" detection feature (FR34-38).
//
// Context cancellation:
// - The Watch method should respect context cancellation
// - When context is cancelled, stop emitting events and clean up resources
// - Close() should be called to release all resources
type FileWatcher interface {
	// Watch starts monitoring the specified paths for file system changes.
	// Returns a channel that emits FileEvent for each detected change.
	// The channel is closed when the watcher is closed or an error occurs.
	//
	// Paths should be canonical absolute paths to directories.
	// Subdirectories may or may not be watched depending on implementation.
	//
	// The returned channel should be buffered to prevent blocking on slow consumers.
	// Events may be coalesced/debounced according to implementation settings.
	Watch(ctx context.Context, paths []string) (<-chan FileEvent, error)

	// Close stops watching and releases all resources.
	// The event channel returned by Watch will be closed.
	// Close is idempotent - calling it multiple times is safe.
	Close() error
}
