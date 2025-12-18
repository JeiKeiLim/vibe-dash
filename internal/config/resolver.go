package config

import (
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Default values matching ports.NewConfig()
const (
	defaultHibernationDays              = 14
	defaultAgentWaitingThresholdMinutes = 10
)

// ConfigResolver resolves effective settings using the cascade:
// project config → master config → defaults
//
// Configuration Priority (per PRD lines 632-636):
//  1. CLI flags (not handled here - handled by CLI layer)
//  2. Project config (~/.vibe-dash/<project>/config.yaml)
//  3. Master config (~/.vibe-dash/config.yaml)
//  4. Built-in defaults (hibernation_days: 14, agent_waiting: 10)
type ConfigResolver struct {
	masterConfig  *ports.Config
	projectConfig *ports.ProjectConfigData
}

// NewConfigResolver creates a resolver with the given master and project configs.
// Either or both configs can be nil - the resolver will use appropriate fallbacks.
func NewConfigResolver(masterConfig *ports.Config, projectConfig *ports.ProjectConfigData) *ConfigResolver {
	return &ConfigResolver{
		masterConfig:  masterConfig,
		projectConfig: projectConfig,
	}
}

// GetEffectiveHibernationDays returns the hibernation days using the cascade:
// 1. Project config CustomHibernationDays (if set)
// 2. Master config HibernationDays (if master config exists)
// 3. Default value (14)
func (r *ConfigResolver) GetEffectiveHibernationDays() int {
	// 1. Project config override
	if r.projectConfig != nil && r.projectConfig.CustomHibernationDays != nil {
		return *r.projectConfig.CustomHibernationDays
	}
	// 2. Master config
	if r.masterConfig != nil {
		return r.masterConfig.HibernationDays
	}
	// 3. Default
	return defaultHibernationDays
}

// GetEffectiveWaitingThreshold returns the agent waiting threshold using the cascade:
// 1. Project config AgentWaitingThresholdMinutes (if set)
// 2. Master config AgentWaitingThresholdMinutes (if master config exists)
// 3. Default value (10)
func (r *ConfigResolver) GetEffectiveWaitingThreshold() int {
	// 1. Project config override
	if r.projectConfig != nil && r.projectConfig.AgentWaitingThresholdMinutes != nil {
		return *r.projectConfig.AgentWaitingThresholdMinutes
	}
	// 2. Master config
	if r.masterConfig != nil {
		return r.masterConfig.AgentWaitingThresholdMinutes
	}
	// 3. Default
	return defaultAgentWaitingThresholdMinutes
}
