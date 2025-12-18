package ports

import (
	"context"
	"fmt"
	"time"
)

// ProjectConfigData holds project-specific configuration.
// Stored in ~/.vibe-dash/<project>/config.yaml
type ProjectConfigData struct {
	// DetectedMethod is the methodology detected (e.g., "speckit", "bmad")
	DetectedMethod string

	// LastScanned is when detection was last run (ISO 8601 UTC)
	LastScanned time.Time

	// CustomHibernationDays overrides global setting. nil = use global.
	CustomHibernationDays *int

	// AgentWaitingThresholdMinutes overrides global setting. nil = use global.
	AgentWaitingThresholdMinutes *int

	// Notes is user-defined project notes/memo
	Notes string
}

// NewProjectConfigData creates a ProjectConfigData with zero values.
// Used for graceful degradation when config cannot be loaded.
func NewProjectConfigData() *ProjectConfigData {
	return &ProjectConfigData{}
}

// Validate checks that ProjectConfigData values are within acceptable ranges.
// Returns nil if valid, or an error describing the first invalid value found.
// Unlike Config.Validate(), this allows nil pointer fields (meaning "use global").
func (d *ProjectConfigData) Validate() error {
	if d.CustomHibernationDays != nil && *d.CustomHibernationDays < 0 {
		return fmt.Errorf("custom_hibernation_days must be >= 0, got %d", *d.CustomHibernationDays)
	}
	if d.AgentWaitingThresholdMinutes != nil && *d.AgentWaitingThresholdMinutes < 0 {
		return fmt.Errorf("agent_waiting_threshold_minutes must be >= 0, got %d", *d.AgentWaitingThresholdMinutes)
	}
	return nil
}

// ProjectConfigLoader handles per-project configuration files.
// Each project has its own config.yaml in its vibe-dash directory.
type ProjectConfigLoader interface {
	// Load reads project-specific configuration from YAML file.
	// Creates default config file if it doesn't exist.
	// Returns defaults on syntax/parse errors (graceful degradation).
	Load(ctx context.Context) (*ProjectConfigData, error)

	// Save persists project configuration to YAML file.
	// Creates file if it doesn't exist.
	Save(ctx context.Context, data *ProjectConfigData) error
}
