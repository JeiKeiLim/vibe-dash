package domain

import "strings"

// Confidence represents the certainty level of a detection result
type Confidence int

const (
	ConfidenceUncertain Confidence = iota // Zero value = Uncertain (safe default)
	ConfidenceLikely
	ConfidenceCertain
)

// String returns human-readable name. Default returns "Unknown" for safety.
func (c Confidence) String() string {
	switch c {
	case ConfidenceUncertain:
		return "Uncertain"
	case ConfidenceLikely:
		return "Likely"
	case ConfidenceCertain:
		return "Certain"
	default:
		return "Unknown"
	}
}

// ParseConfidence converts string to Confidence. Case-insensitive.
func ParseConfidence(s string) (Confidence, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "uncertain":
		return ConfidenceUncertain, nil
	case "likely":
		return ConfidenceLikely, nil
	case "certain":
		return ConfidenceCertain, nil
	case "":
		return ConfidenceUncertain, nil
	default:
		return ConfidenceUncertain, ErrInvalidConfidence
	}
}
