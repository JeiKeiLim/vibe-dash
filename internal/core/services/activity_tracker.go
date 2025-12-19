package services

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// ActivityTracker consumes FileEvents and updates project LastActivityAt.
// It maps file event paths to projects and updates their activity timestamps.
//
// Thread Safety: Safe for concurrent use. Uses RWMutex for project cache.
type ActivityTracker struct {
	repo     ports.ProjectRepository
	projects map[string]*domain.Project // Path prefix -> project (cached)
	mu       sync.RWMutex
}

// NewActivityTracker creates a new tracker with the given repository.
// Call SetProjects() to populate the path cache before ProcessEvents().
func NewActivityTracker(repo ports.ProjectRepository) *ActivityTracker {
	return &ActivityTracker{
		repo:     repo,
		projects: make(map[string]*domain.Project),
	}
}

// SetProjects populates the path cache for path-to-project matching.
// Call this on startup and when projects are added/removed.
func (t *ActivityTracker) SetProjects(projects []*domain.Project) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.projects = make(map[string]*domain.Project, len(projects))
	for _, p := range projects {
		// Normalize path by removing trailing slash
		path := strings.TrimSuffix(p.Path, "/")
		t.projects[path] = p
	}
}

// ProcessEvents consumes events from FileWatcher and updates LastActivityAt.
// Blocks until context is cancelled or channel is closed.
func (t *ActivityTracker) ProcessEvents(ctx context.Context, events <-chan ports.FileEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			t.handleEvent(ctx, event)
		}
	}
}

func (t *ActivityTracker) handleEvent(ctx context.Context, event ports.FileEvent) {
	project := t.findProjectForPath(event.Path)
	if project == nil {
		slog.Debug("event path not matched to any project", "path", event.Path)
		return
	}

	if err := t.repo.UpdateLastActivity(ctx, project.ID, event.Timestamp); err != nil {
		slog.Warn("failed to update last activity", "project_id", project.ID, "error", err)
	}
}

// findProjectForPath matches event path to project using path prefix.
// Returns nil if no matching project is found.
func (t *ActivityTracker) findProjectForPath(eventPath string) *domain.Project {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Normalize event path
	eventPath = strings.TrimSuffix(eventPath, "/")

	for projectPath, project := range t.projects {
		// Check if event path starts with project path followed by "/" or is exact match
		if eventPath == projectPath || strings.HasPrefix(eventPath, projectPath+"/") {
			return project
		}
	}
	return nil
}
