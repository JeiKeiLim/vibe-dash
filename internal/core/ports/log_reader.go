// Package ports defines interfaces for external adapters.
// All interfaces in this package represent contracts between the core domain
// and external dependencies (databases, file systems, detectors, etc.).
//
// Hexagonal Architecture Boundary: This package has ZERO external dependencies.
// Only stdlib (context, time) and internal domain package imports are allowed.
package ports

import (
	"context"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// LogReader provides log reading capabilities for an agentic tool.
// Implementations scan project directories for logs from specific tools
// (e.g., Claude Code, Cursor, Aider) and provide session access.
//
// All methods accepting context.Context must respect cancellation:
// - Check ctx.Done() before long-running operations
// - Return ctx.Err() wrapped with context when cancelled
// - Stop work promptly (within 100ms) when cancellation is signaled
type LogReader interface {
	// Tool returns the agentic tool name (e.g., "Claude Code", "Cursor").
	// Used for display purposes and to identify which reader produced results.
	Tool() string

	// CanRead performs a quick check to determine if this reader
	// can access logs for the given project path. This typically checks
	// if the expected log directory exists.
	//
	// Returns true if log reading should be attempted.
	CanRead(ctx context.Context, projectPath string) bool

	// ListSessions returns available log sessions for a project.
	// Sessions are sorted by recency (newest first).
	//
	// Returns an empty slice if no sessions exist.
	// Returns an error if the log directory exists but cannot be read.
	ListSessions(ctx context.Context, projectPath string) ([]domain.LogSession, error)

	// ReadSession reads all entries from a session file.
	// Invalid JSON lines are skipped (logged at debug level) rather than failing.
	//
	// Returns partial results if some entries fail to parse.
	// Returns an error if the session file cannot be opened.
	ReadSession(ctx context.Context, sessionPath string) ([]domain.LogEntry, error)

	// TailSession streams new log entries as they are written.
	// The implementation polls at a reasonable interval (e.g., 2 seconds).
	//
	// Caller MUST cancel ctx when done reading to stop the polling goroutine.
	// The returned channel is closed when ctx is cancelled or on error.
	//
	// Returns an error if the session file cannot be opened initially.
	TailSession(ctx context.Context, sessionPath string) (<-chan domain.LogEntry, error)
}

// LogReaderRegistry coordinates log reading across multiple LogReaders.
// This follows the same pattern as DetectorRegistry but for log readers.
//
// Thread Safety: Implementations must be safe for concurrent GetReader() calls
// after all Register() calls have completed during initialization.
type LogReaderRegistry interface {
	// Register adds a log reader to the registry.
	// Readers are checked in the order they are registered.
	Register(reader LogReader)

	// GetReader returns the first reader where CanRead() returns true,
	// or nil if no reader can handle this project.
	GetReader(ctx context.Context, projectPath string) LogReader

	// Readers returns all registered log readers.
	Readers() []LogReader
}
