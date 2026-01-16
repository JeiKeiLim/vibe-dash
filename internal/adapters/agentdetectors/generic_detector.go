package agentdetectors

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

const genericDetectorName = "Generic"

// Default threshold matches existing WaitingDetector behavior
const defaultThreshold = 10 * time.Minute

// GenericDetector detects agent activity state using file modification times.
// This is a FALLBACK detector when tool-specific logs are unavailable.
// Implements ports.AgentActivityDetector interface.
type GenericDetector struct {
	threshold time.Duration    // Inactivity threshold (default 10 minutes)
	now       func() time.Time // For testing (default time.Now)
}

// GenericDetectorOption is a functional option for configuring GenericDetector.
type GenericDetectorOption func(*GenericDetector)

// WithThreshold sets a custom inactivity threshold.
func WithThreshold(d time.Duration) GenericDetectorOption {
	return func(g *GenericDetector) {
		if d > 0 {
			g.threshold = d
		}
	}
}

// WithNow sets a custom time function (for testing).
func WithNow(fn func() time.Time) GenericDetectorOption {
	return func(g *GenericDetector) {
		if fn != nil {
			g.now = fn
		}
	}
}

// NewGenericDetector creates a new detector with optional configuration.
func NewGenericDetector(opts ...GenericDetectorOption) *GenericDetector {
	g := &GenericDetector{
		threshold: defaultThreshold,
		now:       time.Now,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// Compile-time interface compliance check
var _ ports.AgentActivityDetector = (*GenericDetector)(nil)

// Name returns the detector identifier.
func (g *GenericDetector) Name() string {
	return genericDetectorName
}

// Detect determines the current agent activity state for a project.
func (g *GenericDetector) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
	// Respect context cancellation at entry
	select {
	case <-ctx.Done():
		return domain.NewAgentState(genericDetectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
	default:
	}

	// Find most recent file modification
	mostRecentModTime, err := g.findMostRecentModification(ctx, projectPath)
	if err != nil || mostRecentModTime.IsZero() {
		// Path doesn't exist, inaccessible, or no files - graceful unknown
		return domain.NewAgentState(genericDetectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
	}

	// Calculate duration since last activity
	duration := g.now().Sub(mostRecentModTime)

	// Handle future timestamps (clock skew)
	if duration < 0 {
		duration = 0
	}

	// Determine state based on threshold
	if duration >= g.threshold {
		return domain.NewAgentState(genericDetectorName, domain.AgentWaitingForUser, duration, domain.ConfidenceUncertain), nil
	}
	return domain.NewAgentState(genericDetectorName, domain.AgentWorking, duration, domain.ConfidenceUncertain), nil
}

// findMostRecentModification walks the directory tree and returns the most recent file modification time.
func (g *GenericDetector) findMostRecentModification(ctx context.Context, dir string) (time.Time, error) {
	// Check context at entry
	select {
	case <-ctx.Done():
		return time.Time{}, nil
	default:
	}

	// Check if path exists
	info, err := os.Stat(dir)
	if err != nil {
		slog.Debug("path stat failed", "path", dir, "error", err)
		return time.Time{}, err
	}
	if !info.IsDir() {
		// Single file - return its mtime
		return info.ModTime(), nil
	}

	var mostRecent time.Time
	filesChecked := 0

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		// Skip errors (permission denied, etc.) with debug logging
		if err != nil {
			slog.Debug("skipping path due to error", "path", path, "error", err)
			return nil // Continue walking
		}

		// Check context periodically (every 100 files balances responsiveness vs syscall overhead)
		filesChecked++
		if filesChecked%100 == 0 {
			select {
			case <-ctx.Done():
				slog.Debug("context cancelled during walk", "filesChecked", filesChecked)
				return filepath.SkipAll
			default:
			}
		}

		// Skip directories
		if d.IsDir() {
			// Skip hidden directories (except root)
			if path != dir && strings.HasPrefix(d.Name(), ".") {
				slog.Debug("skipping hidden directory", "path", path)
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(d.Name(), ".") {
			slog.Debug("skipping hidden file", "path", path)
			return nil
		}

		// Get file info for modification time
		info, err := d.Info()
		if err != nil {
			slog.Debug("skipping inaccessible file", "path", path, "error", err)
			return nil // Skip inaccessible files
		}

		if info.ModTime().After(mostRecent) {
			mostRecent = info.ModTime()
		}

		return nil
	})

	if err != nil {
		return time.Time{}, err
	}

	return mostRecent, nil
}
