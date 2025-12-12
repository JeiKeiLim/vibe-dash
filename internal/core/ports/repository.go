package ports

import (
	"context"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ProjectRepository defines persistence operations for projects.
// This interface abstracts the underlying storage mechanism (SQLite, in-memory, etc.)
// allowing the core domain to remain independent of storage technology choices.
//
// All methods accept context.Context as first parameter to support:
// - Request cancellation (e.g., user hits 'q' in TUI)
// - Timeout handling for slow storage operations
// - Tracing and observability propagation
//
// Error handling:
// - Return domain.ErrProjectNotFound when project doesn't exist
// - Return domain.ErrProjectAlreadyExists for duplicate path conflicts
// - Wrap storage-specific errors with context for debugging
type ProjectRepository interface {
	// Save creates or updates a project in the repository.
	// If a project with the same ID exists, it is updated.
	// Returns error if the operation fails.
	Save(ctx context.Context, project *domain.Project) error

	// FindByID retrieves a project by its unique identifier (path hash).
	// Returns domain.ErrProjectNotFound if no project exists with the given ID.
	FindByID(ctx context.Context, id string) (*domain.Project, error)

	// FindByPath retrieves a project by its canonical absolute path.
	// Returns domain.ErrProjectNotFound if no project exists at the given path.
	FindByPath(ctx context.Context, path string) (*domain.Project, error)

	// FindAll retrieves all projects regardless of their state.
	// Returns an empty slice (not nil) if no projects exist.
	FindAll(ctx context.Context) ([]*domain.Project, error)

	// FindActive retrieves only projects with StateActive.
	// Used for the main dashboard view showing current projects.
	// Returns an empty slice (not nil) if no active projects exist.
	FindActive(ctx context.Context) ([]*domain.Project, error)

	// FindHibernated retrieves only projects with StateHibernated.
	// Used for the hibernated projects view (FR33).
	// Returns an empty slice (not nil) if no hibernated projects exist.
	FindHibernated(ctx context.Context) ([]*domain.Project, error)

	// Delete removes a project by its unique identifier.
	// Returns domain.ErrProjectNotFound if no project exists with the given ID.
	// This is a hard delete - the project is permanently removed.
	Delete(ctx context.Context, id string) error

	// UpdateState changes a project's active/hibernated state.
	// Used for manual hibernation (FR31) and auto-hibernation (FR28).
	// Returns domain.ErrProjectNotFound if no project exists with the given ID.
	UpdateState(ctx context.Context, id string, state domain.ProjectState) error
}
