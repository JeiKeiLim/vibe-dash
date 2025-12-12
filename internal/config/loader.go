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

	// Validate and fix invalid values
	if err := cfg.Validate(); err != nil {
		slog.Warn("config validation failed, using defaults for invalid values", "error", err)
		cfg = l.fixInvalidValues(cfg)
	}

	return cfg, nil
}

// Save persists the given configuration to YAML file.
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

	// Set Viper values from config
	l.v.Set("settings.hibernation_days", config.HibernationDays)
	l.v.Set("settings.refresh_interval_seconds", config.RefreshIntervalSeconds)
	l.v.Set("settings.refresh_debounce_ms", config.RefreshDebounceMs)
	l.v.Set("settings.agent_waiting_threshold_minutes", config.AgentWaitingThresholdMinutes)

	// Convert projects to map format
	projects := make(map[string]interface{})
	for id, pc := range config.Projects {
		projectData := map[string]interface{}{
			"path": pc.Path,
		}
		if pc.DisplayName != "" {
			projectData["display_name"] = pc.DisplayName
		}
		if pc.IsFavorite {
			projectData["favorite"] = pc.IsFavorite
		}
		if pc.HibernationDays != nil {
			projectData["hibernation_days"] = *pc.HibernationDays
		}
		if pc.AgentWaitingThresholdMinutes != nil {
			projectData["agent_waiting_threshold_minutes"] = *pc.AgentWaitingThresholdMinutes
		}
		projects[id] = projectData
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
func (l *ViperLoader) writeDefaultConfig() error {
	cfg := ports.NewConfig()

	content := fmt.Sprintf(`# Vibe Dashboard Configuration
# Auto-generated on first run

settings:
  hibernation_days: %d
  refresh_interval_seconds: %d
  refresh_debounce_ms: %d
  agent_waiting_threshold_minutes: %d

projects: {}
`, cfg.HibernationDays, cfg.RefreshIntervalSeconds, cfg.RefreshDebounceMs, cfg.AgentWaitingThresholdMinutes)

	return os.WriteFile(l.configPath, []byte(content), 0644)
}

// mapViperToConfig converts Viper config values to ports.Config struct.
func (l *ViperLoader) mapViperToConfig() *ports.Config {
	cfg := ports.NewConfig()

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
	projectsMap := l.v.GetStringMap("projects")
	for id, v := range projectsMap {
		projectData, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		pc := ports.ProjectConfig{}
		if path, ok := projectData["path"].(string); ok {
			pc.Path = path
		}
		if name, ok := projectData["display_name"].(string); ok {
			pc.DisplayName = name
		}
		if fav, ok := projectData["favorite"].(bool); ok {
			pc.IsFavorite = fav
		}
		// Handle optional overrides (pointer fields)
		if hd, ok := projectData["hibernation_days"].(int); ok {
			pc.HibernationDays = &hd
		}
		if awt, ok := projectData["agent_waiting_threshold_minutes"].(int); ok {
			pc.AgentWaitingThresholdMinutes = &awt
		}

		cfg.Projects[id] = pc
	}

	return cfg
}

// fixInvalidValues corrects invalid config values by replacing with defaults.
// Logs a warning for each corrected value.
func (l *ViperLoader) fixInvalidValues(cfg *ports.Config) *ports.Config {
	defaults := ports.NewConfig()

	if cfg.HibernationDays < 0 {
		slog.Warn("invalid hibernation_days, using default",
			"invalid_value", cfg.HibernationDays,
			"default_value", defaults.HibernationDays)
		cfg.HibernationDays = defaults.HibernationDays
	}

	if cfg.RefreshIntervalSeconds <= 0 {
		slog.Warn("invalid refresh_interval_seconds, using default",
			"invalid_value", cfg.RefreshIntervalSeconds,
			"default_value", defaults.RefreshIntervalSeconds)
		cfg.RefreshIntervalSeconds = defaults.RefreshIntervalSeconds
	}

	if cfg.RefreshDebounceMs <= 0 {
		slog.Warn("invalid refresh_debounce_ms, using default",
			"invalid_value", cfg.RefreshDebounceMs,
			"default_value", defaults.RefreshDebounceMs)
		cfg.RefreshDebounceMs = defaults.RefreshDebounceMs
	}

	if cfg.AgentWaitingThresholdMinutes < 0 {
		slog.Warn("invalid agent_waiting_threshold_minutes, using default",
			"invalid_value", cfg.AgentWaitingThresholdMinutes,
			"default_value", defaults.AgentWaitingThresholdMinutes)
		cfg.AgentWaitingThresholdMinutes = defaults.AgentWaitingThresholdMinutes
	}

	// Fix per-project overrides
	for id, pc := range cfg.Projects {
		if pc.HibernationDays != nil && *pc.HibernationDays < 0 {
			slog.Warn("invalid project hibernation_days, removing override",
				"project", id, "invalid_value", *pc.HibernationDays)
			pc.HibernationDays = nil
			cfg.Projects[id] = pc
		}
		if pc.AgentWaitingThresholdMinutes != nil && *pc.AgentWaitingThresholdMinutes < 0 {
			slog.Warn("invalid project agent_waiting_threshold_minutes, removing override",
				"project", id, "invalid_value", *pc.AgentWaitingThresholdMinutes)
			pc.AgentWaitingThresholdMinutes = nil
			cfg.Projects[id] = pc
		}
	}

	return cfg
}
