package ports

import "context"

// StateActivator handles project state transitions (Active <-> Hibernated).
// Story 11.3: Added Activate() for TUI auto-activation on file events.
// Story 11.5: Added Hibernate() for CLI manual hibernation command.
type StateActivator interface {
	// Hibernate transitions a project from Active to Hibernated state.
	// Returns ErrInvalidStateTransition if project is already hibernated.
	// Returns ErrFavoriteCannotHibernate if project is marked as favorite.
	// Returns ErrProjectNotFound if project doesn't exist.
	Hibernate(ctx context.Context, projectID string) error

	// Activate transitions a project from Hibernated to Active state.
	// Returns ErrInvalidStateTransition if project is already active.
	// Returns ErrProjectNotFound if project doesn't exist.
	Activate(ctx context.Context, projectID string) error
}
