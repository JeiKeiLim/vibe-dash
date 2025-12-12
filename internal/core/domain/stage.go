package domain

import "strings"

// Stage represents the workflow stage of a project
type Stage int

const (
	StageUnknown Stage = iota // Zero value = Unknown (safe default)
	StageSpecify
	StagePlan
	StageTasks
	StageImplement
)

// String returns human-readable name. Default returns "Unknown" for safety.
func (s Stage) String() string {
	switch s {
	case StageSpecify:
		return "Specify"
	case StagePlan:
		return "Plan"
	case StageTasks:
		return "Tasks"
	case StageImplement:
		return "Implement"
	default:
		return "Unknown"
	}
}

// ParseStage converts string to Stage. Case-insensitive.
func ParseStage(s string) (Stage, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "specify":
		return StageSpecify, nil
	case "plan":
		return StagePlan, nil
	case "tasks":
		return StageTasks, nil
	case "implement":
		return StageImplement, nil
	case "unknown", "":
		return StageUnknown, nil
	default:
		return StageUnknown, ErrInvalidStage
	}
}
