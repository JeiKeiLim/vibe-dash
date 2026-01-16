package domain

import (
	"fmt"
	"time"
)

// DetectionResult represents the result of detecting a project's workflow methodology and stage
type DetectionResult struct {
	Method            string     // "speckit", "bmad", "unknown"
	Stage             Stage      // Detected stage
	Confidence        Confidence // How certain the detection is
	Reasoning         string     // Human-readable explanation (FR11, FR26)
	ArtifactTimestamp time.Time  // Most recent artifact modification time (zero if unknown)
}

// NewDetectionResult creates a new DetectionResult with the given values
func NewDetectionResult(method string, stage Stage, confidence Confidence, reasoning string) DetectionResult {
	return DetectionResult{
		Method:     method,
		Stage:      stage,
		Confidence: confidence,
		Reasoning:  reasoning,
	}
}

// IsUncertain returns true if the confidence level is uncertain
func (dr DetectionResult) IsUncertain() bool {
	return dr.Confidence == ConfidenceUncertain
}

// IsCertain returns true if the confidence level is certain
func (dr DetectionResult) IsCertain() bool {
	return dr.Confidence == ConfidenceCertain
}

// IsLikely returns true if the confidence level is likely
func (dr DetectionResult) IsLikely() bool {
	return dr.Confidence == ConfidenceLikely
}

// Summary returns concise string for logging: "speckit/Plan (Certain)"
func (dr DetectionResult) Summary() string {
	return fmt.Sprintf("%s/%s (%s)", dr.Method, dr.Stage, dr.Confidence)
}

// WithTimestamp returns a copy of the DetectionResult with the ArtifactTimestamp set
func (dr DetectionResult) WithTimestamp(t time.Time) DetectionResult {
	dr.ArtifactTimestamp = t
	return dr
}

// HasTimestamp returns true if the DetectionResult has a non-zero ArtifactTimestamp
func (dr DetectionResult) HasTimestamp() bool {
	return !dr.ArtifactTimestamp.IsZero()
}
