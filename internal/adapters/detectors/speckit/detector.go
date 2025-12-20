// Package speckit implements detection for the Speckit workflow methodology.
// It scans project directories for Speckit markers (specs/, .speckit/, .specify/)
// and determines the current workflow stage based on artifact files.
package speckit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// markerDirs are the directories that indicate a Speckit project.
// Using package-level constant rather than instance field because:
// 1. Markers are fixed by Speckit methodology spec, not configurable
// 2. Simplifies detector to zero-allocation struct (no fields)
// 3. Tests use temp directories with marker folders, not mock markers
var markerDirs = []string{"specs", ".speckit", ".specify"}

// SpeckitDetector implements ports.MethodDetector for Speckit methodology.
type SpeckitDetector struct{}

// NewSpeckitDetector creates a new Speckit detector.
func NewSpeckitDetector() *SpeckitDetector {
	return &SpeckitDetector{}
}

// Name returns the detector identifier.
func (d *SpeckitDetector) Name() string {
	return "speckit"
}

// CanDetect checks if any Speckit marker directory exists at the given path.
func (d *SpeckitDetector) CanDetect(ctx context.Context, path string) bool {
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

// Detect performs full Speckit methodology detection on the given path.
func (d *SpeckitDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
	// Check context first
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Find the specs directory
	specsDir := ""
	for _, marker := range markerDirs {
		markerPath := filepath.Join(path, marker)
		if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
			specsDir = markerPath
			break
		}
	}

	if specsDir == "" {
		return nil, fmt.Errorf("no speckit markers found at %s", path)
	}

	// Find spec subdirectories
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	// Filter to directories only
	var specDirs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			specDirs = append(specDirs, entry)
		}
	}

	if len(specDirs) == 0 {
		// Empty specs directory
		result := domain.NewDetectionResult(
			d.Name(),
			domain.StageUnknown,
			domain.ConfidenceUncertain,
			"specs directory exists but contains no spec subdirectories",
		)
		return &result, nil
	}

	// Check context before continuing
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Find most recently modified spec directory
	targetDir, reasoning := d.findMostRecentDir(specsDir, specDirs)

	// Check context again before file analysis
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Analyze artifacts in target directory
	return d.analyzeSpecDir(filepath.Join(specsDir, targetDir), reasoning)
}

// findMostRecentDir finds the most recently modified directory.
// If modification times cannot be determined, falls back to first directory with explanation.
// Epic 4 Hotfix H4: When mtimes are equal (e.g., after git clone), uses lexicographic
// sort descending so higher-numbered directories (005-*) are preferred over lower (001-*).
func (d *SpeckitDetector) findMostRecentDir(baseDir string, dirs []os.DirEntry) (string, string) {
	if len(dirs) == 1 {
		return dirs[0].Name(), ""
	}

	type dirMod struct {
		name    string
		modTime int64
	}

	var dirMods []dirMod
	for _, dir := range dirs {
		info, err := dir.Info()
		if err != nil {
			continue
		}
		dirMods = append(dirMods, dirMod{name: dir.Name(), modTime: info.ModTime().Unix()})
	}

	// Sort by modification time descending, then by name descending as tiebreaker
	// Epic 4 Hotfix H4: Tiebreaker ensures consistent behavior when mtimes are equal
	// (common after git clone). Higher-numbered specs (005-*) are preferred.
	sort.Slice(dirMods, func(i, j int) bool {
		if dirMods[i].modTime != dirMods[j].modTime {
			return dirMods[i].modTime > dirMods[j].modTime
		}
		// Tiebreaker: sort by name descending (005 > 001)
		return dirMods[i].name > dirMods[j].name
	})

	if len(dirMods) == 0 {
		// Could not determine modification times, fall back to last directory alphabetically
		// (highest-numbered if using numeric prefixes)
		sort.Slice(dirs, func(i, j int) bool {
			return dirs[i].Name() > dirs[j].Name()
		})
		reasoning := fmt.Sprintf("unable to determine modification times, using highest-numbered: %s", dirs[0].Name())
		return dirs[0].Name(), reasoning
	}

	reasoning := fmt.Sprintf("using most recently modified: %s", dirMods[0].name)
	return dirMods[0].name, reasoning
}

// analyzeSpecDir determines the stage based on artifact files.
func (d *SpeckitDetector) analyzeSpecDir(dirPath string, extraReasoning string) (*domain.DetectionResult, error) {
	// Check for artifact files (order matters: check highest stage first)
	hasImplement := d.fileExists(filepath.Join(dirPath, "implement.md"))
	hasTasks := d.fileExists(filepath.Join(dirPath, "tasks.md"))
	hasPlan := d.fileExists(filepath.Join(dirPath, "plan.md"))
	hasSpec := d.fileExists(filepath.Join(dirPath, "spec.md"))

	var stage domain.Stage
	var confidence domain.Confidence
	var reasoning string

	switch {
	case hasImplement:
		stage = domain.StageImplement
		confidence = domain.ConfidenceCertain
		reasoning = "implement.md exists"
	case hasTasks:
		stage = domain.StageTasks
		confidence = domain.ConfidenceCertain
		reasoning = "tasks.md exists"
	case hasPlan:
		stage = domain.StagePlan
		confidence = domain.ConfidenceCertain
		reasoning = "plan.md exists, no tasks.md"
	case hasSpec:
		stage = domain.StageSpecify
		confidence = domain.ConfidenceCertain
		reasoning = "spec.md exists, no plan.md"
	default:
		stage = domain.StageUnknown
		confidence = domain.ConfidenceUncertain
		reasoning = "no standard Speckit artifacts found"
	}

	// Append extra reasoning if present
	if extraReasoning != "" {
		reasoning = reasoning + " (" + extraReasoning + ")"
	}

	result := domain.NewDetectionResult(d.Name(), stage, confidence, reasoning)
	return &result, nil
}

// fileExists checks if a file exists and is not a directory.
func (d *SpeckitDetector) fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
