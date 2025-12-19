package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/sqlite"
	"github.com/JeiKeiLim/vibe-dash/internal/config"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Compile-time interface check
var _ ports.ProjectRepository = (*RepositoryCoordinator)(nil)

// RepositoryCoordinator aggregates multiple per-project repositories.
// Implements ports.ProjectRepository for seamless service layer integration.
// Thread safety is achieved through sync.RWMutex protecting the repo cache.
type RepositoryCoordinator struct {
	configLoader       ports.ConfigLoader
	directoryManager   ports.DirectoryManager
	basePath           string
	repoCache          map[string]*sqlite.ProjectRepository
	projectIDToDirName map[string]string // project ID -> dirName for O(1) lookup in UpdateLastActivity
	mu                 sync.RWMutex
}

// NewRepositoryCoordinator creates a new coordinator with lazy-loaded repositories.
// The coordinator will enumerate projects via configLoader and create directories via directoryManager.
func NewRepositoryCoordinator(
	configLoader ports.ConfigLoader,
	directoryManager ports.DirectoryManager,
	basePath string,
) *RepositoryCoordinator {
	return &RepositoryCoordinator{
		configLoader:       configLoader,
		directoryManager:   directoryManager,
		basePath:           basePath,
		repoCache:          make(map[string]*sqlite.ProjectRepository),
		projectIDToDirName: make(map[string]string),
	}
}

// getProjectRepo returns a cached or newly created repository for the given directory.
// Uses double-checked locking pattern for thread safety.
func (c *RepositoryCoordinator) getProjectRepo(ctx context.Context, dirName string) (*sqlite.ProjectRepository, error) {
	// Context cancellation check first
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Read lock first for cache check
	c.mu.RLock()
	if repo, ok := c.repoCache[dirName]; ok {
		c.mu.RUnlock()
		return repo, nil
	}
	c.mu.RUnlock()

	// Write lock for creation
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if repo, ok := c.repoCache[dirName]; ok {
		return repo, nil
	}

	projectDir := filepath.Join(c.basePath, dirName)
	repo, err := sqlite.NewProjectRepository(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open project %s: %w", dirName, err)
	}

	c.repoCache[dirName] = repo
	return repo, nil
}

// invalidateCache removes a directory from the cache.
// Used after Delete operations to prevent stale cache entries.
func (c *RepositoryCoordinator) invalidateCache(dirName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.repoCache, dirName)
}

// Close clears the repository cache for clean shutdown.
// Note: sqlite.ProjectRepository uses lazy connections (open-per-operation),
// so no explicit close needed per repo.
func (c *RepositoryCoordinator) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.repoCache = make(map[string]*sqlite.ProjectRepository)
	c.projectIDToDirName = make(map[string]string)
	return nil
}

// getAllRepos returns all project repositories and their directory names.
// Config load failure is fatal (returns error), individual repo failures are logged and skipped.
func (c *RepositoryCoordinator) getAllRepos(ctx context.Context) ([]*sqlite.ProjectRepository, []string, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}

	cfg, err := c.configLoader.Load(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	repos := make([]*sqlite.ProjectRepository, 0, len(cfg.Projects))
	dirNames := make([]string, 0, len(cfg.Projects))
	for dirName := range cfg.Projects {
		repo, err := c.getProjectRepo(ctx, dirName)
		if err != nil {
			slog.Warn("skipping corrupted project", "directory", dirName, "error", err)
			continue // Graceful degradation for individual project
		}
		repos = append(repos, repo)
		dirNames = append(dirNames, dirName)
	}
	return repos, dirNames, nil
}

// FindAll retrieves all projects from all project databases.
// Returns empty slice (not nil) if no projects exist.
// Also populates the projectID -> dirName cache for UpdateLastActivity optimization.
func (c *RepositoryCoordinator) FindAll(ctx context.Context) ([]*domain.Project, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	repos, dirNames, err := c.getAllRepos(ctx)
	if err != nil {
		return nil, err
	}

	// CRITICAL: Return empty slice, not nil
	projects := make([]*domain.Project, 0)
	for i, repo := range repos {
		repoProjects, err := repo.FindAll(ctx)
		if err != nil {
			slog.Warn("error reading from project repo", "error", err)
			continue
		}
		// Populate projectID -> dirName cache for each project
		for _, p := range repoProjects {
			c.mu.Lock()
			c.projectIDToDirName[p.ID] = dirNames[i]
			c.mu.Unlock()
		}
		projects = append(projects, repoProjects...)
	}
	return projects, nil
}

// FindActive retrieves all active projects from all databases.
// Returns empty slice (not nil) if no active projects exist.
func (c *RepositoryCoordinator) FindActive(ctx context.Context) ([]*domain.Project, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	repos, _, err := c.getAllRepos(ctx)
	if err != nil {
		return nil, err
	}

	// CRITICAL: Return empty slice, not nil
	projects := make([]*domain.Project, 0)
	for _, repo := range repos {
		repoProjects, err := repo.FindActive(ctx)
		if err != nil {
			slog.Warn("error reading active from project repo", "error", err)
			continue
		}
		projects = append(projects, repoProjects...)
	}
	return projects, nil
}

// FindHibernated retrieves all hibernated projects from all databases.
// Returns empty slice (not nil) if no hibernated projects exist.
func (c *RepositoryCoordinator) FindHibernated(ctx context.Context) ([]*domain.Project, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	repos, _, err := c.getAllRepos(ctx)
	if err != nil {
		return nil, err
	}

	// CRITICAL: Return empty slice, not nil
	projects := make([]*domain.Project, 0)
	for _, repo := range repos {
		repoProjects, err := repo.FindHibernated(ctx)
		if err != nil {
			slog.Warn("error reading hibernated from project repo", "error", err)
			continue
		}
		projects = append(projects, repoProjects...)
	}
	return projects, nil
}

// Save creates or updates a project in the appropriate project database.
// For new projects not in config, creates directory via DirectoryManager and updates config.
func (c *RepositoryCoordinator) Save(ctx context.Context, project *domain.Project) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cfg, err := c.configLoader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dirName, found := cfg.GetDirectoryName(project.Path)
	if !found {
		// NEW project - create directory and update config (AC13)
		fullPath, err := c.directoryManager.EnsureProjectDir(ctx, project.Path)
		if err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}
		dirName = filepath.Base(fullPath)

		// Create per-project config.yaml (Story 3.5.9: removes .project-path redundancy)
		projectConfigLoader, err := config.NewProjectConfigLoader(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create project config loader: %w", err)
		}
		if _, err := projectConfigLoader.Load(ctx); err != nil {
			slog.Warn("failed to create project config.yaml", "path", fullPath, "error", err)
			// Non-fatal - continue without per-project config
		}

		cfg.SetProjectEntry(dirName, project.Path, project.DisplayName, project.IsFavorite)
		if err := c.configLoader.Save(ctx, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	repo, err := c.getProjectRepo(ctx, dirName)
	if err != nil {
		return err
	}
	return repo.Save(ctx, project)
}

// FindByID searches for a project by ID across all databases.
// Returns domain.ErrProjectNotFound if not found in any database.
func (c *RepositoryCoordinator) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	repos, _, err := c.getAllRepos(ctx)
	if err != nil {
		return nil, err
	}

	for _, repo := range repos {
		project, err := repo.FindByID(ctx, id)
		if err == nil {
			return project, nil
		}
		if err != domain.ErrProjectNotFound {
			slog.Warn("error searching for project by ID", "id", id, "error", err)
		}
	}
	return nil, domain.ErrProjectNotFound
}

// FindByPath searches for a project by path.
// First tries fast path via config lookup, falls back to iterating all repos.
// Returns domain.ErrProjectNotFound if not found.
func (c *RepositoryCoordinator) FindByPath(ctx context.Context, path string) (*domain.Project, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cfg, err := c.configLoader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Fast path: try config lookup first
	if dirName, found := cfg.GetDirectoryName(path); found {
		repo, err := c.getProjectRepo(ctx, dirName)
		if err != nil {
			// Fall back to iteration if single repo fails
			slog.Warn("fast path failed, falling back to iteration", "directory", dirName, "error", err)
		} else {
			project, err := repo.FindByPath(ctx, path)
			if err == nil {
				return project, nil
			}
			if err != domain.ErrProjectNotFound {
				slog.Warn("error in fast path FindByPath", "path", path, "error", err)
			}
		}
	}

	// Fallback: iterate all repos (handles edge case of unregistered path)
	repos, _, err := c.getAllRepos(ctx)
	if err != nil {
		return nil, err
	}

	for _, repo := range repos {
		project, err := repo.FindByPath(ctx, path)
		if err == nil {
			return project, nil
		}
		if err != domain.ErrProjectNotFound {
			slog.Warn("error searching for project by path", "path", path, "error", err)
		}
	}
	return nil, domain.ErrProjectNotFound
}

// Delete removes a project by ID from the appropriate database and config.
// Returns domain.ErrProjectNotFound if not found.
func (c *RepositoryCoordinator) Delete(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Find project first to get path
	project, err := c.FindByID(ctx, id)
	if err != nil {
		return err
	}

	cfg, err := c.configLoader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dirName, found := cfg.GetDirectoryName(project.Path)
	if !found {
		return fmt.Errorf("project path not in config: %s", project.Path)
	}

	repo, err := c.getProjectRepo(ctx, dirName)
	if err != nil {
		return err
	}

	if err := repo.Delete(ctx, id); err != nil {
		return err
	}

	// Remove project from config to prevent orphaned entries
	cfg.RemoveProject(dirName)
	if err := c.configLoader.Save(ctx, cfg); err != nil {
		slog.Warn("failed to remove project from config after delete", "directory", dirName, "error", err)
		// Don't fail the delete - project is already removed from DB
	}

	// Invalidate cache after successful delete
	c.invalidateCache(dirName)
	return nil
}

// UpdateState changes a project's state in the appropriate database.
// Returns domain.ErrProjectNotFound if not found.
func (c *RepositoryCoordinator) UpdateState(ctx context.Context, id string, state domain.ProjectState) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Find project first to get path
	project, err := c.FindByID(ctx, id)
	if err != nil {
		return err
	}

	cfg, err := c.configLoader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dirName, found := cfg.GetDirectoryName(project.Path)
	if !found {
		return fmt.Errorf("project path not in config: %s", project.Path)
	}

	repo, err := c.getProjectRepo(ctx, dirName)
	if err != nil {
		return err
	}

	return repo.UpdateState(ctx, id, state)
}

// UpdateLastActivity updates the LastActivityAt for a project.
// Delegates to the appropriate per-project repository.
// Uses cached projectID -> dirName mapping for O(1) lookup (high-frequency operation).
// Returns domain.ErrProjectNotFound if not found.
func (c *RepositoryCoordinator) UpdateLastActivity(ctx context.Context, id string, timestamp time.Time) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Try cached lookup first (O(1) - optimized for high-frequency calls)
	c.mu.RLock()
	dirName, cached := c.projectIDToDirName[id]
	c.mu.RUnlock()

	if cached {
		repo, err := c.getProjectRepo(ctx, dirName)
		if err == nil {
			return repo.UpdateLastActivity(ctx, id, timestamp)
		}
		// If repo fails, fall through to full lookup
	}

	// Cache miss or repo error - fall back to full lookup
	project, err := c.FindByID(ctx, id)
	if err != nil {
		return err
	}

	cfg, err := c.configLoader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dirName, found := cfg.GetDirectoryName(project.Path)
	if !found {
		return fmt.Errorf("project path not in config: %s", project.Path)
	}

	// Cache the mapping for future calls
	c.mu.Lock()
	c.projectIDToDirName[id] = dirName
	c.mu.Unlock()

	repo, err := c.getProjectRepo(ctx, dirName)
	if err != nil {
		return err
	}

	return repo.UpdateLastActivity(ctx, id, timestamp)
}
