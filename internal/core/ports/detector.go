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

// Detector provides methodology detection for a project path.
// This interface is implemented by services.DetectionService and allows
// CLI and other consumers to depend on an abstraction for testing.
type Detector interface {
	// Detect performs methodology detection on the given path.
	// Returns the first successful detection result, or a result with
	// Method="unknown" if no detector matches.
	Detect(ctx context.Context, path string) (*domain.DetectionResult, error)

	// DetectMultiple checks ALL registered detectors and returns all matching results.
	// This supports FR14: detecting multiple methodologies in the same project.
	DetectMultiple(ctx context.Context, path string) ([]*domain.DetectionResult, error)
}

// DetectorRegistry coordinates detection across multiple MethodDetectors.
// This interface is implemented by adapters/detectors.Registry.
//
// Thread Safety: Implementations must be safe for concurrent DetectAll() calls
// after all Register() calls have completed during initialization.
type DetectorRegistry interface {
	// DetectAll tries each registered detector until one succeeds.
	// Returns a result with Method="unknown" if no detector matches.
	DetectAll(ctx context.Context, path string) (*domain.DetectionResult, error)

	// Detectors returns all registered detectors for multi-methodology detection.
	Detectors() []MethodDetector
}

// MethodDetector defines the interface for workflow methodology detection.
// Implementations scan project directories for markers that indicate
// which development methodology is being used (e.g., speckit, bmad).
//
// All methods accepting context.Context must respect cancellation:
// - Check ctx.Done() before long-running operations
// - Return ctx.Err() wrapped with context when cancelled
// - Stop work promptly (within 100ms) when cancellation is signaled
type MethodDetector interface {
	// Name returns the unique identifier for this detector (e.g., "speckit", "bmad").
	// Used for logging and to populate DetectionResult.Method.
	Name() string

	// CanDetect performs a quick check to determine if this detector
	// can potentially handle the given path. This is a lightweight check
	// that may look for specific file patterns or directory structures
	// without performing full detection.
	//
	// Returns true if the detector should attempt full detection.
	// A return value of true does not guarantee detection will succeed.
	CanDetect(ctx context.Context, path string) bool

	// Detect performs full workflow methodology detection on the given path.
	// Returns a DetectionResult containing the detected method, stage,
	// confidence level, and human-readable reasoning.
	//
	// Returns an error if detection cannot be performed (e.g., path not accessible).
	// Returns a result with ConfidenceUncertain if detection completes but
	// methodology cannot be determined.
	Detect(ctx context.Context, path string) (*domain.DetectionResult, error)
}
