package ports

import "context"

// ProjectPathLookup provides existing project → directory mappings.
// Used to ensure same project path always returns same directory name (determinism).
type ProjectPathLookup interface {
	// GetDirForPath returns existing directory name for canonical path.
	// Returns empty string if path not previously registered.
	GetDirForPath(canonicalPath string) string
}

// DirectoryManager handles project directory naming and creation.
// Implements collision resolution per PRD specification:
//   - Use project directory name: ~/.vibe-dash/api-service/
//   - On collision: prepend parent directory: ~/.vibe-dash/client-b-api-service/
//   - Still collision: prepend grandparent: ~/.vibe-dash/work-client-b-api-service/
//   - Continue up directory tree until unique (max 10 levels)
//
// All methods accept context.Context as first parameter per project standards.
//
// Error handling:
//   - Return domain.ErrPathNotAccessible for invalid paths or permission errors
//   - Return domain.ErrCollisionUnresolvable if collision persists after 10 levels
type DirectoryManager interface {
	// GetProjectDirName returns deterministic directory name for project.
	// Uses collision resolution if name already exists for different project.
	// The returned name is normalized (lowercase, special chars → hyphens).
	//
	// Returns error if:
	//   - Path is invalid or doesn't exist (domain.ErrPathNotAccessible)
	//   - Collision unresolvable after 10 levels (domain.ErrCollisionUnresolvable)
	GetProjectDirName(ctx context.Context, projectPath string) (string, error)

	// EnsureProjectDir creates project directory if not exists.
	// Calls GetProjectDirName internally to determine directory name.
	// Returns full path to created/existing directory.
	//
	// Returns error if:
	//   - Path is invalid (domain.ErrPathNotAccessible)
	//   - Directory creation fails - permission denied, disk full (domain.ErrPathNotAccessible)
	//   - Collision unresolvable (domain.ErrCollisionUnresolvable)
	EnsureProjectDir(ctx context.Context, projectPath string) (string, error)

	// DeleteProjectDir removes the project directory and all its contents.
	// The projectPath is the canonical project path (used to look up directory name).
	// Returns nil if directory doesn't exist (idempotent).
	// Returns error if deletion fails for reasons other than non-existence.
	DeleteProjectDir(ctx context.Context, projectPath string) error
}
