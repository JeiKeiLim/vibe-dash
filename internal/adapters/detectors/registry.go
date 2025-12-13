// Package detectors provides workflow methodology detection implementations.
// It contains the Registry that coordinates all registered detectors and
// individual detector packages (e.g., speckit, bmad).
//
// Thread Safety: Registry is NOT safe for concurrent modification.
// All Register() calls must complete before any DetectAll() calls.
// Typical usage: register all detectors during application initialization,
// then use DetectAll() concurrently from multiple goroutines.
package detectors

import (
	"context"
	"fmt"
	"strings"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Registry manages registered method detectors and coordinates detection.
// It is the only component that knows about all detector implementations.
// Services should call Registry.DetectAll(), never individual detectors.
type Registry struct {
	detectors []ports.MethodDetector
}

// Compile-time interface compliance check
var _ ports.DetectorRegistry = (*Registry)(nil)

// NewRegistry creates a new detector registry with no registered detectors.
func NewRegistry() *Registry {
	return &Registry{
		detectors: make([]ports.MethodDetector, 0),
	}
}

// Register adds a detector to the registry.
// Detectors are tried in the order they are registered.
func (r *Registry) Register(detector ports.MethodDetector) {
	r.detectors = append(r.detectors, detector)
}

// Detectors returns the list of registered detectors.
func (r *Registry) Detectors() []ports.MethodDetector {
	return r.detectors
}

// DetectAll tries each registered detector until one succeeds.
// It returns the first successful detection result.
// If no detector matches, it returns a result with Method="unknown".
// Any detector errors are collected and included in the final reasoning.
func (r *Registry) DetectAll(ctx context.Context, path string) (*domain.DetectionResult, error) {
	var detectorErrors []string

	for _, detector := range r.detectors {
		// Check context before each detector
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if detector.CanDetect(ctx, path) {
			result, err := detector.Detect(ctx, path)
			if err == nil && result != nil {
				return result, nil
			}
			// Collect error for final reasoning
			if err != nil {
				detectorErrors = append(detectorErrors, fmt.Sprintf("%s: %v", detector.Name(), err))
			}
		}
	}

	// No detector matched - build reasoning with any errors encountered
	reasoning := "no methodology markers found"
	if len(detectorErrors) > 0 {
		reasoning = fmt.Sprintf("detection failed (%s)", strings.Join(detectorErrors, "; "))
	}

	result := domain.NewDetectionResult(
		"unknown",
		domain.StageUnknown,
		domain.ConfidenceUncertain,
		reasoning,
	)
	return &result, nil
}
