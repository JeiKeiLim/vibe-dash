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
