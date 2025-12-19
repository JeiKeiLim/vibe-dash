// Package services contains core business logic services.
package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// WaitingDetector determines if a project's AI agent is waiting for user input.
// Stateless service - safe for concurrent use.
type WaitingDetector struct {
	resolver ports.ThresholdResolver
	now      func() time.Time // Injected for testing
}

// NewWaitingDetector creates a new WaitingDetector with threshold resolver.
func NewWaitingDetector(resolver ports.ThresholdResolver) *WaitingDetector {
	return &WaitingDetector{
		resolver: resolver,
		now:      time.Now,
	}
}

// IsWaiting returns true if the project's AI agent appears to be waiting for input.
// ctx is included for API consistency with other services (may be used for cancellation in future).
func (d *WaitingDetector) IsWaiting(ctx context.Context, project *domain.Project) bool {
	_ = ctx // Reserved for future use (e.g., cancellation support)
	// Nil project check
	if project == nil {
		slog.Warn("IsWaiting called with nil project")
		return false
	}

	// Hibernated projects are never "waiting"
	if project.State == domain.StateHibernated {
		return false
	}

	// Get effective threshold via resolver (CLI > per-project file > global > default)
	thresholdMinutes := d.resolver.Resolve(project.ID)

	// Threshold of 0 means detection is disabled
	if thresholdMinutes == 0 {
		return false
	}

	// Newly added project: if LastActivityAt equals CreatedAt, no activity observed yet
	if project.LastActivityAt.Equal(project.CreatedAt) {
		return false
	}

	threshold := time.Duration(thresholdMinutes) * time.Minute
	inactiveDuration := d.now().Sub(project.LastActivityAt)

	return inactiveDuration >= threshold
}

// WaitingDuration returns how long the project has been in waiting state.
// Returns 0 if project is not waiting.
func (d *WaitingDetector) WaitingDuration(ctx context.Context, project *domain.Project) time.Duration {
	if !d.IsWaiting(ctx, project) {
		return 0
	}
	return d.now().Sub(project.LastActivityAt)
}
