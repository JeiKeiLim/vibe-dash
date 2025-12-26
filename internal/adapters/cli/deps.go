package cli

import "github.com/JeiKeiLim/vibe-dash/internal/core/ports"

// directoryManager handles project directory operations (delete).
var directoryManager ports.DirectoryManager

// configWarning stores any config loading warning (Story 7.2).
// Set during main.go initialization when config has errors.
var configWarning string

// detailLayout stores the detail panel layout mode (Story 8.6).
// Valid values: "vertical" (default, side-by-side), "horizontal" (stacked).
var detailLayout = "vertical"

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
// Valid values: "vertical" (default, side-by-side), "horizontal" (stacked).
func SetDetailLayout(layout string) {
	if layout == "horizontal" || layout == "vertical" {
		detailLayout = layout
	} else {
		detailLayout = "vertical" // Fallback to default
	}
}

// GetDetailLayout returns the current detail panel layout mode (Story 8.6).
func GetDetailLayout() string {
	return detailLayout
}
