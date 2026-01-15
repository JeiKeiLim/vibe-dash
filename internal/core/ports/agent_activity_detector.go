// Package ports defines interfaces for external adapters.
// All interfaces in this package represent contracts between the core domain
// and external dependencies (databases, file systems, detectors, etc.).
//
// Hexagonal Architecture Boundary: This package has ZERO external dependencies.
// Only stdlib (context, time) and internal domain package imports are allowed.
package ports

import (
	"context"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// AgentActivityDetector detects the activity state of an AI agent for a project.
// Implementations parse tool-specific logs or use heuristics to determine
// whether an agent is Working, WaitingForUser, or Inactive.
//
// All methods accepting context.Context must respect cancellation:
// - Check ctx.Done() before long-running operations
// - Return ctx.Err() wrapped with context when cancelled
// - Stop work promptly (within 100ms) when cancellation is signaled
type AgentActivityDetector interface {
	// Detect determines the current agent activity state for a project.
	// Returns AgentState with Status, Duration, Tool, and Confidence.
	//
	// Returns error if detection cannot be performed (e.g., path not accessible).
	// Returns AgentState with Status=AgentUnknown if detection completes but
	// agent state cannot be determined.
	Detect(ctx context.Context, projectPath string) (domain.AgentState, error)

	// Name returns the detector identifier (e.g., "Claude Code", "Generic").
	// Used for logging and to populate AgentState.Tool when this detector matches.
	Name() string
}
