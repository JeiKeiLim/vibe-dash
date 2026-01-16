package statsview

import (
	"context"
	"time"
)

// Transition is a simplified type for sparkline calculations.
// Contains only the timestamp needed for activity bucketing.
// Avoids importing metrics package types into TUI layer.
type Transition struct {
	TransitionedAt time.Time
}

// MetricsReader is the interface for reading metrics data.
// TUI depends on interface, not concrete repository.
// Returns []Transition to avoid cross-adapter coupling.
type MetricsReader interface {
	// GetTransitionTimestamps retrieves transition times for a project since given time.
	// Returns empty slice on any error (graceful degradation).
	GetTransitionTimestamps(ctx context.Context, projectID string, since time.Time) []Transition
}

// FullTransitionReader provides full transition data for breakdown calculation.
// Extends MetricsReader with richer data needed for time-per-stage calculations.
type FullTransitionReader interface {
	MetricsReader
	// GetFullTransitions retrieves full transition data for a project since given time.
	// Returns empty slice on any error (graceful degradation).
	GetFullTransitions(ctx context.Context, projectID string, since time.Time) []FullTransition
}
