package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// HibernationService handles auto-hibernation of inactive projects.
// Projects are hibernated when they have no activity for a configurable number of days.
// Favorites are never auto-hibernated (FR30).
type HibernationService struct {
	repo         ports.ProjectRepository
	stateService *StateService
	config       *ports.Config
}

// Compile-time interface compliance check
var _ ports.HibernationService = (*HibernationService)(nil)

// NewHibernationService creates a new HibernationService.
func NewHibernationService(
	repo ports.ProjectRepository,
	stateService *StateService,
	config *ports.Config,
) *HibernationService {
	return &HibernationService{
		repo:         repo,
		stateService: stateService,
		config:       config,
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

		// Get effective threshold (respects per-project override)
		thresholdDays := h.config.GetEffectiveHibernationDays(project.ID)

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
