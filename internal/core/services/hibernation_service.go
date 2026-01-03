package services

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/config"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// HibernationService handles auto-hibernation of inactive projects.
// Projects are hibernated when they have no activity for a configurable number of days.
// Favorites are never auto-hibernated (FR30).
type HibernationService struct {
	repo         ports.ProjectRepository
	stateService *StateService
	config       *ports.Config
	vibeHome     string // Base path for per-project configs (~/.vibe-dash)
}

// Compile-time interface compliance check
var _ ports.HibernationService = (*HibernationService)(nil)

// NewHibernationService creates a new HibernationService.
func NewHibernationService(
	repo ports.ProjectRepository,
	stateService *StateService,
	cfg *ports.Config,
	vibeHome string,
) *HibernationService {
	return &HibernationService{
		repo:         repo,
		stateService: stateService,
		config:       cfg,
		vibeHome:     vibeHome,
	}
}

// CheckAndHibernate processes all active projects and hibernates inactive ones.
// Returns count of successfully hibernated projects.
// Continues processing if individual projects fail (partial failure tolerance).
func (h *HibernationService) CheckAndHibernate(ctx context.Context) (int, error) {
	// CRITICAL: Use FindActive(), NOT FindAll()
	// FindActive returns only projects with State == domain.StateActive
	// This automatically excludes already-hibernated projects (AC7)
	projects, err := h.repo.FindActive(ctx)
	if err != nil {
		return 0, err
	}

	hibernatedCount := 0
	for _, project := range projects {
		// Skip favorites (FR30, AC2)
		if project.IsFavorite {
			continue
		}

		// Get effective threshold (respects per-project config file override)
		thresholdDays := h.getEffectiveHibernationDays(ctx, project)

		// Check if auto-hibernation is disabled
		if thresholdDays == 0 {
			continue
		}

		// Check inactivity
		if !h.isInactive(project.LastActivityAt, thresholdDays) {
			continue
		}

		// Hibernate via StateService (reuse Story 11.1)
		if err := h.stateService.Hibernate(ctx, project.ID); err != nil {
			// Log but continue processing other projects (partial failure tolerance)
			slog.Warn("failed to hibernate project", "project_id", project.ID, "error", err)
			continue
		}

		hibernatedCount++
	}

	return hibernatedCount, nil
}

// isInactive checks if a project is inactive based on the threshold.
// IMPORTANT: Use > not >= for boundary condition.
// Project with exactly 14 days inactivity should NOT be hibernated yet.
func (h *HibernationService) isInactive(lastActivityAt time.Time, thresholdDays int) bool {
	threshold := time.Duration(thresholdDays) * 24 * time.Hour
	return time.Since(lastActivityAt) > threshold
}

// getEffectiveHibernationDays returns threshold for project.
// Priority: per-project config file > global config
func (h *HibernationService) getEffectiveHibernationDays(ctx context.Context, project *domain.Project) int {
	// Resolve directory name from config (canonical path -> dir name)
	dirName := h.config.GetDirForPath(project.Path)
	if dirName == "" {
		// Project not in config - use global
		return h.config.HibernationDays
	}

	projectDir := filepath.Join(h.vibeHome, dirName)
	loader, err := config.NewProjectConfigLoader(projectDir)
	if err != nil {
		return h.config.HibernationDays
	}

	data, err := loader.Load(ctx)
	if err != nil {
		slog.Debug("failed to load per-project config, using global",
			"project", project.Name, "error", err)
		return h.config.HibernationDays
	}

	if data.CustomHibernationDays != nil {
		return *data.CustomHibernationDays
	}

	return h.config.HibernationDays
}
