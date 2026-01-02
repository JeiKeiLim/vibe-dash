package ports

import "context"

// StateActivator handles project state activation.
// Extracted interface for testability in TUI layer (Story 11.3).
type StateActivator interface {
	// Activate transitions a project from Hibernated to Active state.
	// Returns ErrInvalidStateTransition if project is already active.
	// Returns ErrProjectNotFound if project doesn't exist.
	Activate(ctx context.Context, projectID string) error
}
