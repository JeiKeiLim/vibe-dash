package ports

import "context"

// HibernationService handles auto-hibernation of inactive projects.
// Projects are hibernated when they have no activity for a configurable number of days.
// Favorites are never auto-hibernated (FR30).
type HibernationService interface {
	// CheckAndHibernate processes all active projects and hibernates inactive ones.
	// Returns count of successfully hibernated projects.
	// Continues processing if individual projects fail (partial failure tolerance).
	CheckAndHibernate(ctx context.Context) (int, error)
}
