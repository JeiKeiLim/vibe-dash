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
