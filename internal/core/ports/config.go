package ports

import (
	"context"
	"fmt"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// Config represents application configuration settings.
// Storage format v2 uses directory names as project keys.
// Global settings apply to all projects unless overridden per-project.
type Config struct {
	// StorageVersion is the config format version (must be 2)
	StorageVersion int

	// HibernationDays is the number of days of inactivity before auto-hibernation (FR28)
	// Default: 14. Set to 0 to disable auto-hibernation.
	HibernationDays int

	// RefreshIntervalSeconds is the TUI refresh interval in seconds
	// Default: 10. Must be > 0.
	RefreshIntervalSeconds int

	// RefreshDebounceMs is the debounce delay for file events in milliseconds
	// Default: 200. Must be > 0.
	RefreshDebounceMs int

	// AgentWaitingThresholdMinutes is minutes of inactivity before showing ⏸️ WAITING (FR34-38)
	// Default: 10. Set to 0 to disable waiting detection.
	AgentWaitingThresholdMinutes int

	// Projects contains per-project configuration overrides
	// Key is the directory name (v2 format uses directory_name as map key)
	Projects map[string]ProjectConfig
}

// ProjectConfig represents per-project configuration in master config (FR47).
// Per-project setting overrides (hibernation, waiting threshold) are now in
// per-project config files (~/.vibe-dash/<project>/config.yaml) - see Story 3.5.3.
type ProjectConfig struct {
	// Path is the canonical absolute path to the project directory
	Path string

	// DirectoryName is the subdirectory name under ~/.vibe-dash/
	// Usually derived from path but may be disambiguated (e.g., "client-b-api-service")
	// This field mirrors the map key for lookup convenience.
	DirectoryName string

	// DisplayName is a user-set custom name for the project (FR5)
	// Empty string means use the derived name
	DisplayName string

	// IsFavorite marks the project as always visible regardless of activity (FR30)
	IsFavorite bool

	// HibernationDays overrides the global setting for this project
	// nil means use global setting
	// DEPRECATED: Use per-project config file instead (Story 3.5.3)
	HibernationDays *int

	// AgentWaitingThresholdMinutes overrides the global setting for this project
	// nil means use global setting
	// DEPRECATED: Use per-project config file instead (Story 3.5.3)
	AgentWaitingThresholdMinutes *int
}

// NewConfig creates a Config with default values.
// Call this instead of creating Config{} directly to ensure defaults are set.
func NewConfig() *Config {
	return &Config{
		StorageVersion:               2,
		HibernationDays:              14,
		RefreshIntervalSeconds:       10,
		RefreshDebounceMs:            200,
		AgentWaitingThresholdMinutes: 10,
		Projects:                     make(map[string]ProjectConfig),
	}
}

// GetProjectConfig returns the ProjectConfig for the given project ID.
// Returns the config and true if found, or zero value and false if not found.
func (c *Config) GetProjectConfig(projectID string) (ProjectConfig, bool) {
	pc, ok := c.Projects[projectID]
	return pc, ok
}

// GetEffectiveHibernationDays returns the hibernation days for a project.
// Returns the project-specific override if set, otherwise the global value.
func (c *Config) GetEffectiveHibernationDays(projectID string) int {
	if pc, ok := c.Projects[projectID]; ok && pc.HibernationDays != nil {
		return *pc.HibernationDays
	}
	return c.HibernationDays
}

// GetEffectiveWaitingThreshold returns the agent waiting threshold for a project.
// Returns the project-specific override if set, otherwise the global value.
func (c *Config) GetEffectiveWaitingThreshold(projectID string) int {
	if pc, ok := c.Projects[projectID]; ok && pc.AgentWaitingThresholdMinutes != nil {
		return *pc.AgentWaitingThresholdMinutes
	}
	return c.AgentWaitingThresholdMinutes
}

// Compile-time interface compliance check
var _ ProjectPathLookup = (*Config)(nil)

// GetDirForPath implements ports.ProjectPathLookup interface.
// Used by DirectoryManager to ensure same path always returns same directory name.
// Returns empty string if path not registered.
func (c *Config) GetDirForPath(canonicalPath string) string {
	if c.Projects == nil {
		return ""
	}
	for dirName, pc := range c.Projects {
		if pc.Path == canonicalPath {
			return dirName
		}
	}
	return ""
}

// GetDirectoryName returns the directory name for a project path.
// Wrapper around GetDirForPath with bool return for convenience.
func (c *Config) GetDirectoryName(path string) (string, bool) {
	dirName := c.GetDirForPath(path)
	return dirName, dirName != ""
}

// GetProjectPath returns the canonical path for a directory name.
func (c *Config) GetProjectPath(directoryName string) (string, bool) {
	if c.Projects == nil {
		return "", false
	}
	if pc, ok := c.Projects[directoryName]; ok {
		return pc.Path, true
	}
	return "", false
}

// SetProjectEntry adds or updates a project entry in the config.
func (c *Config) SetProjectEntry(directoryName, path, displayName string, favorite bool) {
	if c.Projects == nil {
		c.Projects = make(map[string]ProjectConfig)
	}
	c.Projects[directoryName] = ProjectConfig{
		Path:          path,
		DirectoryName: directoryName,
		DisplayName:   displayName,
		IsFavorite:    favorite,
	}
}

// RemoveProject removes a project entry from the config.
func (c *Config) RemoveProject(directoryName string) bool {
	if c.Projects == nil {
		return false
	}
	if _, ok := c.Projects[directoryName]; ok {
		delete(c.Projects, directoryName)
		return true
	}
	return false
}

// Validate checks that Config values are within acceptable ranges.
// Call after loading config or modifying values programmatically.
// Returns domain.ErrConfigInvalid wrapped with specific details on validation failure.
func (c *Config) Validate() error {
	// Validate storage version (must be 2 for v2 format)
	if c.StorageVersion != 2 {
		return fmt.Errorf("%w: storage_version must be 2, got %d", domain.ErrConfigInvalid, c.StorageVersion)
	}

	if c.HibernationDays < 0 {
		return fmt.Errorf("%w: hibernation_days must be >= 0, got %d", domain.ErrConfigInvalid, c.HibernationDays)
	}
	if c.RefreshIntervalSeconds <= 0 {
		return fmt.Errorf("%w: refresh_interval_seconds must be > 0, got %d", domain.ErrConfigInvalid, c.RefreshIntervalSeconds)
	}
	if c.RefreshDebounceMs <= 0 {
		return fmt.Errorf("%w: refresh_debounce_ms must be > 0, got %d", domain.ErrConfigInvalid, c.RefreshDebounceMs)
	}
	if c.AgentWaitingThresholdMinutes < 0 {
		return fmt.Errorf("%w: agent_waiting_threshold_minutes must be >= 0, got %d", domain.ErrConfigInvalid, c.AgentWaitingThresholdMinutes)
	}

	// Validate per-project overrides
	for projectID, pc := range c.Projects {
		if pc.HibernationDays != nil && *pc.HibernationDays < 0 {
			return fmt.Errorf("%w: project %s hibernation_days must be >= 0, got %d",
				domain.ErrConfigInvalid, projectID, *pc.HibernationDays)
		}
		if pc.AgentWaitingThresholdMinutes != nil && *pc.AgentWaitingThresholdMinutes < 0 {
			return fmt.Errorf("%w: project %s agent_waiting_threshold_minutes must be >= 0, got %d",
				domain.ErrConfigInvalid, projectID, *pc.AgentWaitingThresholdMinutes)
		}
	}

	return nil
}

// ConfigLoader defines the interface for loading and saving application configuration.
// Implementations handle the underlying storage mechanism (YAML files, environment, etc.).
//
// All methods accept context.Context to support:
// - Request cancellation (e.g., user hits 'q' in TUI during slow I/O)
// - Timeout handling for unresponsive storage backends
type ConfigLoader interface {
	// Load reads configuration from the storage backend.
	// Returns a Config with values from storage merged with defaults.
	// Returns error if the storage backend is inaccessible or config is malformed.
	// Implementations should check ctx.Done() before blocking I/O operations.
	Load(ctx context.Context) (*Config, error)

	// Save persists the given configuration to the storage backend.
	// Returns error if the storage backend is not writable.
	// Implementations should check ctx.Done() before blocking I/O operations.
	Save(ctx context.Context, config *Config) error
}
