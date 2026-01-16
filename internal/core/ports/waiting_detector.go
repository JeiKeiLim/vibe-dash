package ports

import (
	"context"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// WaitingDetector determines if a project's AI agent is waiting for user input.
// Implementations should be stateless and safe for concurrent use.
type WaitingDetector interface {
	// IsWaiting returns true if the project's agent appears to be waiting.
	// Returns false for nil projects, hibernated projects, or newly added projects.
	IsWaiting(ctx context.Context, project *domain.Project) bool

	// WaitingDuration returns how long the project has been waiting.
	// Returns 0 if project is not waiting.
	WaitingDuration(ctx context.Context, project *domain.Project) time.Duration

	// AgentState returns the full agent detection state including confidence and tool.
	// Story 15.7: Enables detail panel to display confidence level.
	AgentState(ctx context.Context, project *domain.Project) domain.AgentState
}
