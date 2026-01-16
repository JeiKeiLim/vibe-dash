package cli

import (
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/metrics"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Version info - set from main.go
var (
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
)

// SetVersion sets the version info for the CLI and configures the version template.
func SetVersion(version, commit, date string) {
	appVersion = version
	appCommit = commit
	appDate = date
	setupVersion() // Configure cobra version template
}

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

// stateService handles state activation for auto-activation on file events (Story 11.3).
var stateService ports.StateActivator

// logReaderRegistry handles log reading for Claude Code logs (Story 12.1).
var logReaderRegistry ports.LogReaderRegistry

// metricsRecorder handles stage transition recording for progress metrics (Story 16.2).
// Optional dependency - nil means metrics are disabled.
var metricsRecorder *metrics.MetricsRecorder

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

// SetStateService sets the state service for auto-activation (Story 11.3).
func SetStateService(svc ports.StateActivator) {
	stateService = svc
}

// SetLogReaderRegistry sets the log reader registry for log viewing (Story 12.1).
func SetLogReaderRegistry(registry ports.LogReaderRegistry) {
	logReaderRegistry = registry
}

// SetMetricsRecorder sets the metrics recorder for stage transition tracking (Story 16.2).
func SetMetricsRecorder(recorder *metrics.MetricsRecorder) {
	metricsRecorder = recorder
}

// GetMetricsRecorder returns the metrics recorder (may be nil).
func GetMetricsRecorder() *metrics.MetricsRecorder {
	return metricsRecorder
}
