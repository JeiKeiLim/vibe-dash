package domain

import "strings"

// AgentStatus represents the detected state of an AI agent.
// Uses int with iota to match Confidence and Stage patterns.
type AgentStatus int

const (
	AgentUnknown        AgentStatus = iota // Zero value - cannot determine state
	AgentWorking                           // Agent actively processing/using tools
	AgentWaitingForUser                    // Agent waiting for user input (THE target state)
	AgentInactive                          // No recent agent activity
)

// String returns human-readable name. Default returns "Unknown" for safety.
func (s AgentStatus) String() string {
	switch s {
	case AgentWorking:
		return "Working"
	case AgentWaitingForUser:
		return "Waiting"
	case AgentInactive:
		return "Inactive"
	default:
		return "Unknown"
	}
}

// ParseAgentStatus converts string to AgentStatus. Case-insensitive.
func ParseAgentStatus(s string) (AgentStatus, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "working":
		return AgentWorking, nil
	case "waiting", "waitingforuser":
		return AgentWaitingForUser, nil
	case "inactive":
		return AgentInactive, nil
	case "unknown", "":
		return AgentUnknown, nil
	default:
		return AgentUnknown, ErrInvalidAgentStatus
	}
}
