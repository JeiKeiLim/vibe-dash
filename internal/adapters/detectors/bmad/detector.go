// Package bmad implements detection for the BMAD v6 workflow methodology.
// It scans project directories for BMAD markers (.bmad/ folder with bmm/config.yaml)
// and extracts version information from the config header.
package bmad

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Compile-time interface compliance check
var _ ports.MethodDetector = (*BMADDetector)(nil)

// markerDirs are the directories that indicate a BMAD v6 project.
// Priority order (first match wins):
//   - .bmad: Original v6 hidden folder
//   - _bmad: New v6 visible folder (LLM-friendly, Alpha.22+)
//   - _bmad-output: Output artifacts folder (indicates BMAD project)
var markerDirs = []string{".bmad", "_bmad", "_bmad-output"}

// configPaths are the relative paths to check for BMAD config within marker folders.
// Both core/config.yaml and bmm/config.yaml are valid (user's choice during install).
// Try core first as it typically contains the version header.
var configPaths = []string{"core/config.yaml", "bmm/config.yaml"}

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

	slog.Debug("checking bmad markers", "path", path)
	for _, marker := range markerDirs {
		markerPath := filepath.Join(path, marker)
		if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
			slog.Debug("bmad marker found", "marker", marker)
			return true
		}
	}
	return false
}

// Detect performs full BMAD v6 methodology detection on the given path.
// It checks for marker directories (.bmad, _bmad, _bmad-output) in priority order,
// then looks for config.yaml in core/ or bmm/ subdirectories to extract version.
// Stage detection uses sprint-status.yaml or artifact analysis.
// Returns ConfidenceCertain if config.yaml exists with version and stage detected.
// Returns ConfidenceLikely if marker exists but config.yaml is missing or stage uncertain.
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

	// Special case: _bmad-output has no config.yaml
	if strings.HasSuffix(bmadDir, "_bmad-output") {
		result := domain.NewDetectionResult(
			d.Name(),
			domain.StageUnknown,
			domain.ConfidenceLikely,
			"BMAD detected (_bmad-output), config not expected",
		)
		return &result, nil
	}

	// Try each config path in order
	var version string
	var cfgFound bool
	for _, cfgRelPath := range configPaths {
		cfgPath := filepath.Join(bmadDir, cfgRelPath)
		slog.Debug("checking bmad config", "config_path", cfgPath)
		v, err := extractVersion(cfgPath)
		if err == nil {
			version = v
			cfgFound = true
			slog.Debug("bmad config found", "config_path", cfgPath, "version", version)
			break // Use first valid config found
		}
	}

	if !cfgFound {
		// No config found - return lower confidence
		result := domain.NewDetectionResult(
			d.Name(),
			domain.StageUnknown,
			domain.ConfidenceLikely,
			filepath.Base(bmadDir)+" folder exists but config.yaml not found",
		)
		return &result, nil
	}

	// Check context again after file read
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Detect stage from sprint-status.yaml or artifacts
	// Pass bmadDir so we can read config for artifact paths
	stage, stageConfidence, stageReasoning := d.detectStage(ctx, path, bmadDir)

	// Build combined reasoning
	var fullReasoning string
	if version != "" {
		fullReasoning = "BMAD v" + version + ", " + stageReasoning
	} else {
		fullReasoning = "BMAD detected, " + stageReasoning
	}

	// Use the more confident confidence level
	// If stage detection is uncertain, downgrade to Likely
	finalConfidence := domain.ConfidenceCertain
	if stageConfidence == domain.ConfidenceUncertain {
		finalConfidence = domain.ConfidenceLikely
	}

	result := domain.NewDetectionResult(
		d.Name(),
		stage,
		finalConfidence,
		fullReasoning,
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

// BMADConfig holds parsed BMAD configuration values relevant to stage detection.
type BMADConfig struct {
	OutputFolder            string `yaml:"output_folder"`
	SprintArtifacts         string `yaml:"sprint_artifacts"`
	ImplementationArtifacts string `yaml:"implementation_artifacts"`
}

// parseConfig reads and parses a BMAD config file to extract artifact paths.
// Returns nil if the file cannot be read or parsed (not an error for detection).
func parseConfig(configPath string) *BMADConfig {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var cfg BMADConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	return &cfg
}

// resolveConfigPath replaces {project-root} placeholder with actual project path.
func resolveConfigPath(configValue, projectPath string) string {
	if configValue == "" {
		return ""
	}
	return strings.ReplaceAll(configValue, "{project-root}", projectPath)
}

// findBMADConfig searches for and parses BMAD config from the marker directory.
// Tries both core/config.yaml and bmm/config.yaml, merging values from both.
func findBMADConfig(bmadDir string) *BMADConfig {
	var merged BMADConfig

	for _, cfgRelPath := range configPaths {
		cfgPath := filepath.Join(bmadDir, cfgRelPath)
		if cfg := parseConfig(cfgPath); cfg != nil {
			// Merge: later values override earlier if non-empty
			if cfg.OutputFolder != "" {
				merged.OutputFolder = cfg.OutputFolder
			}
			if cfg.SprintArtifacts != "" {
				merged.SprintArtifacts = cfg.SprintArtifacts
			}
			if cfg.ImplementationArtifacts != "" {
				merged.ImplementationArtifacts = cfg.ImplementationArtifacts
			}
		}
	}

	// Return nil if nothing was found
	if merged.OutputFolder == "" && merged.SprintArtifacts == "" && merged.ImplementationArtifacts == "" {
		return nil
	}

	return &merged
}
