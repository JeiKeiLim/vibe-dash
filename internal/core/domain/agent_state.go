package domain

import (
	"fmt"
	"time"
)

// AgentState represents the complete detected state of an AI agent for a project.
type AgentState struct {
	Tool       string        // "Claude Code", "Generic", "Unknown"
	Status     AgentStatus   // Working, WaitingForUser, Inactive, Unknown
	Duration   time.Duration // How long in current state
	Confidence Confidence    // High (log-based), Low (heuristic)
}

// NewAgentState creates a new AgentState with the given values.
func NewAgentState(tool string, status AgentStatus, duration time.Duration, confidence Confidence) AgentState {
	return AgentState{
		Tool:       tool,
		Status:     status,
		Duration:   duration,
		Confidence: confidence,
	}
}

// IsWaiting returns true if the agent is waiting for user input.
func (s AgentState) IsWaiting() bool {
	return s.Status == AgentWaitingForUser
}

// IsWorking returns true if the agent is actively working.
func (s AgentState) IsWorking() bool {
	return s.Status == AgentWorking
}

// IsInactive returns true if the agent has no recent activity.
func (s AgentState) IsInactive() bool {
	return s.Status == AgentInactive
}

// IsUnknown returns true if the agent state cannot be determined.
func (s AgentState) IsUnknown() bool {
	return s.Status == AgentUnknown
}

// Summary returns concise string for logging: "Claude Code/Waiting (Certain)"
func (s AgentState) Summary() string {
	return fmt.Sprintf("%s/%s (%s)", s.Tool, s.Status, s.Confidence)
}
