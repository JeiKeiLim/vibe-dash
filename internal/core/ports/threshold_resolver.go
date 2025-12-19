package ports

// ThresholdResolver determines the effective waiting threshold for a project.
// Implementations handle the cascade: CLI > per-project config > global config > default.
type ThresholdResolver interface {
	// Resolve returns the effective waiting threshold in minutes for the given project.
	// Returns 0 if detection is disabled, positive value for threshold.
	// The cascade priority is: CLI flag > per-project config FILE > global config > default (10).
	Resolve(projectID string) int
}
