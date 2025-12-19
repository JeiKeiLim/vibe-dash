package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Project config YAML schema documentation:
//
// ~/.vibe-dash/<project>/config.yaml
//
//	detected_method: "speckit"           # Methodology detected (speckit, bmad, etc.)
//	last_scanned: "2025-12-18T10:30:00Z" # ISO 8601 UTC timestamp
//	custom_hibernation_days: 7           # Optional override (omit to use global)
//	agent_waiting_threshold_minutes: 5   # Optional override (omit to use global)
//	notes: "Main API service"            # Project notes/memo

// ViperProjectConfigLoader implements ports.ProjectConfigLoader using Viper for YAML parsing.
// Each project has its own config.yaml in its vibe-dash directory (~/.vibe-dash/<project>/).
type ViperProjectConfigLoader struct {
	projectDir string
	configPath string
	v          *viper.Viper
}

// NewProjectConfigLoader creates a loader for project-specific config.
// projectDir must exist as a directory.
func NewProjectConfigLoader(projectDir string) (*ViperProjectConfigLoader, error) {
	// Validate projectDir exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: project directory does not exist: %s",
			domain.ErrPathNotAccessible, projectDir)
	}

	configPath := filepath.Join(projectDir, "config.yaml")
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	return &ViperProjectConfigLoader{
		projectDir: projectDir,
		configPath: configPath,
		v:          v,
	}, nil
}

// ProjectDir returns the project directory this loader operates on.
// Useful for debugging and logging.
func (l *ViperProjectConfigLoader) ProjectDir() string {
	return l.projectDir
}

// ConfigPath returns the path to the config file this loader reads/writes.
// Useful for debugging and logging.
func (l *ViperProjectConfigLoader) ConfigPath() string {
	return l.configPath
}

// Load reads project-specific configuration from YAML file.
// Creates default config file if it doesn't exist.
// Returns defaults on syntax/parse errors (graceful degradation).
func (l *ViperProjectConfigLoader) Load(ctx context.Context) (*ports.ProjectConfigData, error) {
	// Check context cancellation first
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check if config file exists
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		// Create default config file
		if err := l.writeDefaultConfig(); err != nil {
			slog.Warn("could not write default project config, using defaults",
				"error", err, "path", l.configPath)
			return ports.NewProjectConfigData(), nil
		}
	}

	// Read config file
	if err := l.v.ReadInConfig(); err != nil {
		slog.Warn("project config syntax error, using defaults",
			"error", err, "path", l.configPath)
		return ports.NewProjectConfigData(), nil
	}

	// Map Viper values to ProjectConfigData struct
	data := l.mapViperToProjectConfigData()

	// Validate and fix invalid values (AC7: graceful degradation)
	if err := data.Validate(); err != nil {
		slog.Warn("project config validation failed, fixing invalid values",
			"error", err, "path", l.configPath)
		data = l.fixInvalidValues(data)
	}

	return data, nil
}

// Save persists project configuration to YAML file.
// Creates file if it doesn't exist.
func (l *ViperProjectConfigLoader) Save(ctx context.Context, data *ports.ProjectConfigData) error {
	// Check context cancellation first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Set Viper values from ProjectConfigData
	l.v.Set("detected_method", data.DetectedMethod)

	// Format LastScanned as ISO 8601 UTC
	if !data.LastScanned.IsZero() {
		l.v.Set("last_scanned", data.LastScanned.Format(time.RFC3339))
	} else {
		l.v.Set("last_scanned", "")
	}

	// Handle optional override fields
	if data.CustomHibernationDays != nil {
		l.v.Set("custom_hibernation_days", *data.CustomHibernationDays)
	} else {
		// Remove the key if nil (don't persist null values)
		l.v.Set("custom_hibernation_days", nil)
	}

	if data.AgentWaitingThresholdMinutes != nil {
		l.v.Set("agent_waiting_threshold_minutes", *data.AgentWaitingThresholdMinutes)
	} else {
		l.v.Set("agent_waiting_threshold_minutes", nil)
	}

	l.v.Set("notes", data.Notes)

	// Write config file
	return l.v.WriteConfigAs(l.configPath)
}

// writeDefaultConfig creates the default config file with comments.
func (l *ViperProjectConfigLoader) writeDefaultConfig() error {
	content := `# Project-specific vibe-dash configuration
# Auto-generated - modify as needed

# Methodology detected by vibe-dash
detected_method: ""
last_scanned: ""

# Optional: Override global hibernation threshold
# custom_hibernation_days: 7

# Optional: Override global agent waiting threshold
# agent_waiting_threshold_minutes: 5

# Project notes/memo
notes: ""
`
	return os.WriteFile(l.configPath, []byte(content), 0644)
}

// fixInvalidValues corrects invalid project config values by removing invalid overrides.
// Logs a warning for each corrected value.
func (l *ViperProjectConfigLoader) fixInvalidValues(data *ports.ProjectConfigData) *ports.ProjectConfigData {
	if data.CustomHibernationDays != nil && *data.CustomHibernationDays < 0 {
		slog.Warn("invalid custom_hibernation_days, removing override",
			"invalid_value", *data.CustomHibernationDays)
		data.CustomHibernationDays = nil
	}

	if data.AgentWaitingThresholdMinutes != nil && *data.AgentWaitingThresholdMinutes < 0 {
		slog.Warn("invalid agent_waiting_threshold_minutes, removing override",
			"invalid_value", *data.AgentWaitingThresholdMinutes)
		data.AgentWaitingThresholdMinutes = nil
	}

	return data
}

// mapViperToProjectConfigData converts Viper config values to ProjectConfigData struct.
func (l *ViperProjectConfigLoader) mapViperToProjectConfigData() *ports.ProjectConfigData {
	data := ports.NewProjectConfigData()

	// DetectedMethod
	if l.v.IsSet("detected_method") {
		data.DetectedMethod = l.v.GetString("detected_method")
	}

	// LastScanned - parse ISO 8601 UTC timestamp
	if l.v.IsSet("last_scanned") {
		lastScannedStr := l.v.GetString("last_scanned")
		if lastScannedStr != "" {
			t, err := time.Parse(time.RFC3339, lastScannedStr)
			if err != nil {
				slog.Warn("invalid last_scanned timestamp, using zero value",
					"error", err, "value", lastScannedStr)
			} else {
				data.LastScanned = t
			}
		}
	}

	// CustomHibernationDays - optional override
	if l.v.IsSet("custom_hibernation_days") && l.v.Get("custom_hibernation_days") != nil {
		val := l.v.GetInt("custom_hibernation_days")
		data.CustomHibernationDays = &val
	}

	// AgentWaitingThresholdMinutes - optional override
	if l.v.IsSet("agent_waiting_threshold_minutes") && l.v.Get("agent_waiting_threshold_minutes") != nil {
		val := l.v.GetInt("agent_waiting_threshold_minutes")
		data.AgentWaitingThresholdMinutes = &val
	}

	// Notes
	if l.v.IsSet("notes") {
		data.Notes = l.v.GetString("notes")
	}

	return data
}
