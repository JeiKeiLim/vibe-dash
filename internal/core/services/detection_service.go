// Package services provides core business logic for the vibe-dash application.
// Services orchestrate domain logic and coordinate between ports.
//
// Hexagonal Architecture Boundary: Services live in the core layer and must NOT
// import from internal/adapters/. They depend only on domain types and port interfaces.
package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// DetectionService orchestrates methodology detection across registered detectors.
// It provides the core business logic for workflow detection.
//
// Thread Safety: Safe for concurrent use. All methods are stateless and
// delegate to the underlying registry which handles its own thread safety.
type DetectionService struct {
	registry ports.DetectorRegistry
}

// Compile-time interface compliance check
var _ ports.Detector = (*DetectionService)(nil)

// NewDetectionService creates a new detection service with the given registry.
// Panics if registry is nil - this is a programming error that should be caught early.
func NewDetectionService(registry ports.DetectorRegistry) *DetectionService {
	if registry == nil {
		panic("DetectionService requires non-nil registry")
	}
	return &DetectionService{
		registry: registry,
	}
}

// Detect performs methodology detection on the given path.
// Returns the first successful detection result, or a result with
// Method="unknown" if no detector matches.
//
// Return type is *DetectionResult (pointer) for single detection.
// See DetectMultiple for []*DetectionResult (slice of pointers) when
// detecting multiple methodologies.
func (s *DetectionService) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
	// Validate path is not empty
	if path == "" {
		return nil, fmt.Errorf("%w: empty path", domain.ErrPathNotAccessible)
	}

	// Check context at entry
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result, err := s.registry.DetectAll(ctx, path)
	if err != nil {
		// Wrap with domain error for consistent error handling
		return nil, fmt.Errorf("%w: %v", domain.ErrDetectionFailed, err)
	}

	return result, nil
}

// DetectMultiple checks ALL registered detectors and returns all matching results.
// This supports FR14: detecting multiple methodologies in the same project.
//
// Return type is []*DetectionResult (slice of pointers) because multiple
// methodologies may be detected. See Detect for single-methodology detection.
//
// Error Handling: DetectMultiple continues checking all detectors even if some fail.
// Individual detector errors are logged but don't stop iteration. Returns empty
// slice if no detectors match (not an error - just no methodologies found).
func (s *DetectionService) DetectMultiple(ctx context.Context, path string) ([]*domain.DetectionResult, error) {
	// Validate path is not empty
	if path == "" {
		return nil, fmt.Errorf("%w: empty path", domain.ErrPathNotAccessible)
	}

	// Check context at entry
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var results []*domain.DetectionResult

	for _, detector := range s.registry.Detectors() {
		// Check context before each detector
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if detector.CanDetect(ctx, path) {
			result, err := detector.Detect(ctx, path)
			if err != nil {
				// AC5: Log detector errors but continue to next detector (resilient design)
				slog.Debug("detector error during multi-detection",
					"detector", detector.Name(),
					"path", path,
					"error", err,
				)
				continue
			}
			if result != nil {
				results = append(results, result)
			}
		}
	}

	return results, nil
}
