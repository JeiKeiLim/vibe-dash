package services

import (
	"context"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Compile-time interface compliance check
var _ ports.StateActivator = (*StateService)(nil)

// StateService handles project state transitions (Active <-> Hibernated).
// It enforces business rules:
// - Only Active projects can be hibernated
// - Only Hibernated projects can be activated
// - Favorite projects cannot be hibernated (FR30)
type StateService struct {
	repo ports.ProjectRepository
}

// NewStateService creates a new StateService with the given repository.
func NewStateService(repo ports.ProjectRepository) *StateService {
	return &StateService{
		repo: repo,
	}
}

// Hibernate transitions a project from Active to Hibernated state.
// Returns ErrInvalidStateTransition if project is already hibernated.
// Returns ErrFavoriteCannotHibernate if project is marked as favorite (FR30).
// Returns ErrProjectNotFound if project doesn't exist.
func (s *StateService) Hibernate(ctx context.Context, projectID string) error {
	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Check if favorite (FR30: favorites never auto-hibernate)
	if project.IsFavorite {
		return domain.ErrFavoriteCannotHibernate
	}

	// Validate transition: must be Active
	if project.State == domain.StateHibernated {
		return domain.ErrInvalidStateTransition
	}

	// Update state and timestamps
	now := time.Now()
	project.State = domain.StateHibernated
	project.HibernatedAt = &now
	project.UpdatedAt = now

	return s.repo.Save(ctx, project)
}

// Activate transitions a project from Hibernated to Active state.
// Returns ErrInvalidStateTransition if project is already active.
// Returns ErrProjectNotFound if project doesn't exist.
func (s *StateService) Activate(ctx context.Context, projectID string) error {
	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Validate transition: must be Hibernated
	if project.State == domain.StateActive {
		return domain.ErrInvalidStateTransition
	}

	// Update state and timestamps (use single now for consistency)
	now := time.Now()
	project.State = domain.StateActive
	project.HibernatedAt = nil
	project.UpdatedAt = now

	return s.repo.Save(ctx, project)
}
