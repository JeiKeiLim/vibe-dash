// Package bmad implements detection for the BMAD v6 workflow methodology.
// It scans project directories for BMAD markers (.bmad/ folder with bmm/config.yaml)
// and extracts version information from the config header.
package bmad

import (
	"context"
	"os"
	"path/filepath"
	"regexp"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Compile-time interface compliance check
var _ ports.MethodDetector = (*BMADDetector)(nil)

// markerDirs are the directories that indicate a BMAD v6 project.
// BMAD v6 uses .bmad/ folder (v4 used .bmad-core/ which is not supported here).
var markerDirs = []string{".bmad"}

// configPath is the relative path to the BMAD config file within the .bmad folder.
const configPath = "bmm/config.yaml"

// versionRegex extracts version from the config header comment.
// Example: "# Version: 6.0.0-alpha.13" -> "6.0.0-alpha.13"
var versionRegex = regexp.MustCompile(`# Version:\s*(\S+)`)

// BMADDetector implements ports.MethodDetector for BMAD v6 methodology.
type BMADDetector struct{}

// NewBMADDetector creates a new BMAD v6 detector.
func NewBMADDetector() *BMADDetector {
	return &BMADDetector{}
}

// Name returns the detector identifier.
func (d *BMADDetector) Name() string {
	return "bmad"
}

// CanDetect checks if .bmad/ folder exists at the given path.
// This is a FAST O(1) check - only verifies the folder exists.
// Does NOT check for config.yaml - that's Detect's responsibility.
func (d *BMADDetector) CanDetect(ctx context.Context, path string) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	for _, marker := range markerDirs {
		markerPath := filepath.Join(path, marker)
		if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

// Detect performs full BMAD v6 methodology detection on the given path.
// It checks for .bmad/bmm/config.yaml and extracts version from the header.
// Returns ConfidenceCertain if config.yaml exists with version.
// Returns ConfidenceLikely if .bmad/ exists but config.yaml is missing.
// Always returns StageUnknown (stage detection is Story 4.5-2).
func (d *BMADDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
	// Check context first
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Find the .bmad directory
	bmadDir := ""
	for _, marker := range markerDirs {
		markerPath := filepath.Join(path, marker)
		if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
			bmadDir = markerPath
			break
		}
	}

	if bmadDir == "" {
		return nil, nil // Not a BMAD project - return nil (not an error)
	}

	// Check context before continuing
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check for config.yaml
	cfgPath := filepath.Join(bmadDir, configPath)
	version, err := extractVersion(cfgPath)

	if err != nil {
		// config.yaml doesn't exist or can't be read - lower confidence
		result := domain.NewDetectionResult(
			d.Name(),
			domain.StageUnknown, // Always Unknown in this story
			domain.ConfidenceLikely,
			".bmad folder exists but config.yaml not found",
		)
		return &result, nil
	}

	// Check context again after file read
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Build reasoning with version if available
	var reasoning string
	if version != "" {
		reasoning = "BMAD v" + version + " detected (.bmad/bmm/config.yaml found)"
	} else {
		reasoning = "BMAD detected (.bmad/bmm/config.yaml found, version not in header)"
	}

	result := domain.NewDetectionResult(
		d.Name(),
		domain.StageUnknown, // Always Unknown in this story
		domain.ConfidenceCertain,
		reasoning,
	)
	return &result, nil
}

// extractVersion reads the config file and extracts version from header comment.
// Returns empty string if version not found (not an error).
// Returns error if file cannot be read.
func extractVersion(configPath string) (string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	match := versionRegex.FindSubmatch(data)
	if match == nil {
		return "", nil // No version found, not an error
	}
	return string(match[1]), nil
}
