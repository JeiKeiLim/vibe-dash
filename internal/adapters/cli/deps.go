package cli

import "github.com/JeiKeiLim/vibe-dash/internal/core/ports"

// directoryManager handles project directory operations (delete).
var directoryManager ports.DirectoryManager

// configWarning stores any config loading warning (Story 7.2).
// Set during main.go initialization when config has errors.
var configWarning string

// detailLayout stores the detail panel layout mode (Story 8.6).
// Valid values: "horizontal" (default, stacked), "vertical" (side-by-side).
var detailLayout = "horizontal"

// appConfig stores the loaded configuration for TUI access (Story 8.7).
var appConfig *ports.Config

// hibernationService handles auto-hibernation of inactive projects (Story 11.2).
var hibernationService ports.HibernationService

// SetDirectoryManager sets the directory manager for CLI commands.
func SetDirectoryManager(dm ports.DirectoryManager) {
	directoryManager = dm
}

// SetConfigWarning sets the config warning for JSON output (Story 7.2, AC7).
// Pass empty string to clear the warning.
func SetConfigWarning(warning string) {
	configWarning = warning
}

// GetConfigWarning returns the current config warning (Story 7.2).
func GetConfigWarning() string {
	return configWarning
}

// SetDetailLayout sets the detail panel layout mode (Story 8.6).
// Valid values: "horizontal" (default, stacked), "vertical" (side-by-side).
func SetDetailLayout(layout string) {
	if layout == "horizontal" || layout == "vertical" {
		detailLayout = layout
	} else {
		detailLayout = "horizontal" // Fallback to default
	}
}

// GetDetailLayout returns the current detail panel layout mode (Story 8.6).
func GetDetailLayout() string {
	return detailLayout
}

// SetConfig stores the config for TUI components to access (Story 8.7).
func SetConfig(cfg *ports.Config) {
	appConfig = cfg
}

// GetConfig returns the stored config, or defaults if not set (Story 8.7).
// CRITICAL: Always returns non-nil - never crashes on missing config.
func GetConfig() *ports.Config {
	if appConfig == nil {
		return ports.NewConfig()
	}
	return appConfig
}

// SetHibernationService sets the hibernation service for auto-hibernation (Story 11.2).
func SetHibernationService(svc ports.HibernationService) {
	hibernationService = svc
}
