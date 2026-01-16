// Package speckit implements detection for the Speckit workflow methodology.
// It scans project directories for Speckit markers (specs/, .speckit/, .specify/)
// and determines the current workflow stage based on artifact files.
package speckit

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

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

	slog.Debug("checking speckit markers", "path", path)
	for _, marker := range markerDirs {
		markerPath := filepath.Join(path, marker)
		slog.Debug("checking marker directory", "marker", marker, "full_path", markerPath)
		if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
			slog.Debug("speckit marker found", "marker", marker)
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

	slog.Debug("analyzing specs directory", "specs_dir", specsDir)

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

	slog.Debug("found spec subdirectories", "count", len(specDirs))

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
	targetDir, reasoning, dirMtime := d.findMostRecentDir(specsDir, specDirs)

	// Check context again before file analysis
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Analyze artifacts in target directory
	return d.analyzeSpecDir(filepath.Join(specsDir, targetDir), reasoning, dirMtime)
}

// findMostRecentDir finds the most recently modified directory.
// If modification times cannot be determined, falls back to first directory with explanation.
// Epic 4 Hotfix H4: When mtimes are equal (e.g., after git clone), uses lexicographic
// sort descending so higher-numbered directories (005-*) are preferred over lower (001-*).
// Returns the directory name, reasoning string, and directory modification time.
func (d *SpeckitDetector) findMostRecentDir(baseDir string, dirs []os.DirEntry) (string, string, time.Time) {
	if len(dirs) == 1 {
		info, err := dirs[0].Info()
		if err == nil {
			return dirs[0].Name(), fmt.Sprintf("spec: %s", dirs[0].Name()), info.ModTime()
		}
		return dirs[0].Name(), fmt.Sprintf("spec: %s", dirs[0].Name()), time.Time{}
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
		return dirs[0].Name(), reasoning, time.Time{}
	}

	reasoning := fmt.Sprintf("spec: %s, most recently modified", dirMods[0].name)
	dirMtime := time.Unix(dirMods[0].modTime, 0)
	return dirMods[0].name, reasoning, dirMtime
}

// speckitArtifacts are the standard Speckit artifact files checked for stage detection.
// Order matters: highest stage first (implement > tasks > plan > spec).
var speckitArtifacts = []string{"implement.md", "tasks.md", "plan.md", "spec.md"}

// analyzeSpecDir determines the stage based on artifact files.
// dirMtime is the directory modification time from findMostRecentDir.
// Returns DetectionResult with ArtifactTimestamp set to max(dirMtime, file mtimes).
func (d *SpeckitDetector) analyzeSpecDir(dirPath string, extraReasoning string, dirMtime time.Time) (*domain.DetectionResult, error) {
	// Track the maximum modification time and file existence in a single pass
	maxMtime := dirMtime
	fileExists := make(map[string]bool)

	for _, file := range speckitArtifacts {
		filePath := filepath.Join(dirPath, file)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			fileExists[file] = true
			if info.ModTime().After(maxMtime) {
				maxMtime = info.ModTime()
			}
		}
	}

	hasImplement := fileExists["implement.md"]
	hasTasks := fileExists["tasks.md"]
	hasPlan := fileExists["plan.md"]
	hasSpec := fileExists["spec.md"]

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

	result := domain.NewDetectionResult(d.Name(), stage, confidence, reasoning).WithTimestamp(maxMtime)
	return &result, nil
}
