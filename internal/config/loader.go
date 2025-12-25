package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

const currentStorageVersion = 2

// ViperLoader implements ports.ConfigLoader using Viper for YAML parsing.
type ViperLoader struct {
	configPath string
	v          *viper.Viper
}

// NewViperLoader creates a ConfigLoader that reads from the specified config path.
// If configPath is empty, uses the default ~/.vibe-dash/config.yaml
func NewViperLoader(configPath string) *ViperLoader {
	if configPath == "" {
		configPath = GetDefaultConfigPath()
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	return &ViperLoader{
		configPath: configPath,
		v:          v,
	}
}

// Load reads configuration from YAML file.
// Creates config directory and file with defaults if they don't exist.
// Returns defaults on any error (graceful degradation per AC3-AC5).
func (l *ViperLoader) Load(ctx context.Context) (*ports.Config, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Ensure config directory exists
	if err := l.ensureConfigDir(); err != nil {
		slog.Warn("could not create config directory, using defaults", "error", err)
		return ports.NewConfig(), nil
	}

	// Check if config file exists
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		// Create default config file
		if err := l.writeDefaultConfig(); err != nil {
			slog.Warn("could not write default config, using defaults", "error", err)
			return ports.NewConfig(), nil
		}
	}

	// Read config file
	if err := l.v.ReadInConfig(); err != nil {
		slog.Warn("config syntax error, using defaults", "error", err, "path", l.configPath)
		return ports.NewConfig(), nil
	}

	// Map Viper values to Config struct
	cfg := l.mapViperToConfig()

	// Check if migration is needed (v1 → v2)
	if cfg.StorageVersion != currentStorageVersion {
		cfg = l.migrateV1ToV2(cfg)
	}

	// Validate and fix invalid values
	if err := cfg.Validate(); err != nil {
		slog.Warn("config validation failed, using defaults for invalid values", "error", err)
		cfg = l.fixInvalidValues(cfg)
	}

	return cfg, nil
}

// Save persists the given configuration to YAML file.
// Writes v2 format with storage_version at root and directory_name as project key.
// DOES NOT write deprecated fields (HibernationDays, AgentWaitingThresholdMinutes per-project).
func (l *ViperLoader) Save(ctx context.Context, config *ports.Config) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Ensure config directory exists
	if err := l.ensureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// CRITICAL: Write storage_version at root level (Subtask 2.3)
	l.v.Set("storage_version", config.StorageVersion)

	// Global settings
	l.v.Set("settings.hibernation_days", config.HibernationDays)
	l.v.Set("settings.refresh_interval_seconds", config.RefreshIntervalSeconds)
	l.v.Set("settings.refresh_debounce_ms", config.RefreshDebounceMs)
	l.v.Set("settings.agent_waiting_threshold_minutes", config.AgentWaitingThresholdMinutes)

	// Projects - directory_name as key, do NOT write deprecated fields (Subtask 2.4)
	projects := make(map[string]interface{})
	for dirName, pc := range config.Projects {
		projectData := map[string]interface{}{
			"path":           pc.Path,
			"directory_name": pc.DirectoryName, // Explicit for clarity
		}
		if pc.DisplayName != "" {
			projectData["display_name"] = pc.DisplayName
		}
		if pc.IsFavorite {
			projectData["favorite"] = pc.IsFavorite
		}
		// NOTE: HibernationDays and AgentWaitingThresholdMinutes are NOT written
		// These are deprecated - use per-project config files instead (Story 3.5.3)
		projects[dirName] = projectData
	}
	l.v.Set("projects", projects)

	// Write config file
	return l.v.WriteConfig()
}

// ensureConfigDir creates the config directory if it doesn't exist.
func (l *ViperLoader) ensureConfigDir() error {
	configDir := filepath.Dir(l.configPath)
	return os.MkdirAll(configDir, 0755)
}

// writeDefaultConfig creates the default config file with comments.
// Writes v2 format with storage_version at top.
func (l *ViperLoader) writeDefaultConfig() error {
	cfg := ports.NewConfig()

	content := fmt.Sprintf(`# Vibe Dashboard Master Configuration
# Auto-generated - storage_version: 2 format
# Per-project settings are in ~/.vibe-dash/<project>/config.yaml

storage_version: 2

settings:
  hibernation_days: %d
  refresh_interval_seconds: %d
  refresh_debounce_ms: %d
  agent_waiting_threshold_minutes: %d

# Projects map: directory_name → project info
# Keys are subdirectory names under ~/.vibe-dash/
projects: {}
`, cfg.HibernationDays, cfg.RefreshIntervalSeconds, cfg.RefreshDebounceMs, cfg.AgentWaitingThresholdMinutes)

	return os.WriteFile(l.configPath, []byte(content), 0644)
}

// mapViperToConfig converts Viper config values to ports.Config struct.
func (l *ViperLoader) mapViperToConfig() *ports.Config {
	cfg := ports.NewConfig()

	// Read storage_version from root level (Subtask 2.1)
	// If not set, set to 0 to trigger migration (v1 format didn't have storage_version)
	if l.v.IsSet("storage_version") {
		cfg.StorageVersion = l.v.GetInt("storage_version")
	} else {
		cfg.StorageVersion = 0 // Trigger migration
	}

	// Only override defaults if values are explicitly set in config file
	if l.v.IsSet("settings.hibernation_days") {
		cfg.HibernationDays = l.v.GetInt("settings.hibernation_days")
	}
	if l.v.IsSet("settings.refresh_interval_seconds") {
		cfg.RefreshIntervalSeconds = l.v.GetInt("settings.refresh_interval_seconds")
	}
	if l.v.IsSet("settings.refresh_debounce_ms") {
		cfg.RefreshDebounceMs = l.v.GetInt("settings.refresh_debounce_ms")
	}
	if l.v.IsSet("settings.agent_waiting_threshold_minutes") {
		cfg.AgentWaitingThresholdMinutes = l.v.GetInt("settings.agent_waiting_threshold_minutes")
	}

	// Map projects if present
	// In v2 format, the map key IS the directory_name (Subtask 2.2)
	projectsMap := l.v.GetStringMap("projects")
	for dirName, v := range projectsMap {
		projectData, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		pc := ports.ProjectConfig{}
		// Key is directory_name in v2 format
		pc.DirectoryName = dirName

		if path, ok := projectData["path"].(string); ok {
			pc.Path = path
		}
		if name, ok := projectData["display_name"].(string); ok {
			pc.DisplayName = name
		}
		if fav, ok := projectData["favorite"].(bool); ok {
			pc.IsFavorite = fav
		}
		// Handle optional overrides (pointer fields) - DEPRECATED but kept for backward compatibility
		// Log warning when deprecated fields are read (Subtask 5.5)
		if hd, ok := projectData["hibernation_days"].(int); ok {
			slog.Warn("deprecated: hibernation_days in master config project entry, use per-project config",
				"directory_name", dirName, "value", hd)
			pc.HibernationDays = &hd
		}
		if awt, ok := projectData["agent_waiting_threshold_minutes"].(int); ok {
			slog.Warn("deprecated: agent_waiting_threshold_minutes in master config project entry, use per-project config",
				"directory_name", dirName, "value", awt)
			pc.AgentWaitingThresholdMinutes = &awt
		}

		cfg.Projects[dirName] = pc
	}

	return cfg
}

// fixInvalidValues corrects invalid config values by replacing with defaults.
// Logs a warning for each corrected value with path context (AC1, AC2).
func (l *ViperLoader) fixInvalidValues(cfg *ports.Config) *ports.Config {
	defaults := ports.NewConfig()

	if cfg.HibernationDays < 0 {
		slog.Warn("invalid hibernation_days, using default",
			"path", l.configPath,
			"invalid_value", cfg.HibernationDays,
			"default_value", defaults.HibernationDays)
		cfg.HibernationDays = defaults.HibernationDays
	}

	if cfg.RefreshIntervalSeconds <= 0 {
		slog.Warn("invalid refresh_interval_seconds, using default",
			"path", l.configPath,
			"invalid_value", cfg.RefreshIntervalSeconds,
			"default_value", defaults.RefreshIntervalSeconds)
		cfg.RefreshIntervalSeconds = defaults.RefreshIntervalSeconds
	}

	if cfg.RefreshDebounceMs <= 0 {
		slog.Warn("invalid refresh_debounce_ms, using default",
			"path", l.configPath,
			"invalid_value", cfg.RefreshDebounceMs,
			"default_value", defaults.RefreshDebounceMs)
		cfg.RefreshDebounceMs = defaults.RefreshDebounceMs
	}

	if cfg.AgentWaitingThresholdMinutes < 0 {
		slog.Warn("invalid agent_waiting_threshold_minutes, using default",
			"path", l.configPath,
			"invalid_value", cfg.AgentWaitingThresholdMinutes,
			"default_value", defaults.AgentWaitingThresholdMinutes)
		cfg.AgentWaitingThresholdMinutes = defaults.AgentWaitingThresholdMinutes
	}

	// Fix per-project overrides
	for id, pc := range cfg.Projects {
		if pc.HibernationDays != nil && *pc.HibernationDays < 0 {
			slog.Warn("invalid project hibernation_days, removing override",
				"path", l.configPath,
				"project", id, "invalid_value", *pc.HibernationDays)
			pc.HibernationDays = nil
			cfg.Projects[id] = pc
		}
		if pc.AgentWaitingThresholdMinutes != nil && *pc.AgentWaitingThresholdMinutes < 0 {
			slog.Warn("invalid project agent_waiting_threshold_minutes, removing override",
				"path", l.configPath,
				"project", id, "invalid_value", *pc.AgentWaitingThresholdMinutes)
			pc.AgentWaitingThresholdMinutes = nil
			cfg.Projects[id] = pc
		}
	}

	// Fix invalid storage_version (Subtask 2.6)
	if cfg.StorageVersion != currentStorageVersion {
		slog.Warn("invalid storage_version, using default",
			"path", l.configPath,
			"invalid_value", cfg.StorageVersion,
			"default_value", currentStorageVersion)
		cfg.StorageVersion = currentStorageVersion
	}

	return cfg
}

// migrateV1ToV2 attempts to migrate v1 config format to v2.
// V1 used arbitrary project IDs; v2 uses directory names.
// Deprecated fields are warned but preserved for read compatibility.
func (l *ViperLoader) migrateV1ToV2(cfg *ports.Config) *ports.Config {
	slog.Warn("migrating config from v1 to v2 format",
		"old_version", cfg.StorageVersion,
		"new_version", currentStorageVersion)

	cfg.StorageVersion = currentStorageVersion

	// Migrate projects - use base directory name as key
	newProjects := make(map[string]ports.ProjectConfig)
	for oldKey, pc := range cfg.Projects {
		if pc.Path == "" {
			slog.Warn("migration: skipping project with empty path", "key", oldKey)
			continue
		}

		// Warn about deprecated fields (still readable but won't be written)
		if pc.HibernationDays != nil {
			slog.Warn("migration: hibernation_days in master config is deprecated, use per-project config",
				"path", pc.Path, "value", *pc.HibernationDays)
		}
		if pc.AgentWaitingThresholdMinutes != nil {
			slog.Warn("migration: agent_waiting_threshold_minutes in master config is deprecated, use per-project config",
				"path", pc.Path, "value", *pc.AgentWaitingThresholdMinutes)
		}

		// Use base name of path as directory name
		dirName := filepath.Base(pc.Path)

		// Skip if collision (warn user)
		if _, exists := newProjects[dirName]; exists {
			slog.Warn("migration collision - project skipped, please re-add",
				"path", pc.Path, "directory_name", dirName)
			continue
		}

		pc.DirectoryName = dirName
		newProjects[dirName] = pc
	}
	cfg.Projects = newProjects

	return cfg
}
