package config

import (
	"context"
	"path/filepath"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

const defaultWaitingThreshold = 10

// Compile-time interface compliance check
var _ ports.ThresholdResolver = (*WaitingThresholdResolver)(nil)

// WaitingThresholdResolver implements ports.ThresholdResolver with cascade logic.
type WaitingThresholdResolver struct {
	globalConfig *ports.Config
	vibeHome     string // ~/.vibe-dash path for project config resolution
	cliOverride  int    // -1 means not set (use config)
}

// NewWaitingThresholdResolver creates a resolver with cascade support.
// cliOverride: -1 means "use config", 0 means "disabled", >0 means threshold minutes.
func NewWaitingThresholdResolver(
	globalConfig *ports.Config,
	vibeHome string,
	cliOverride int,
) *WaitingThresholdResolver {
	return &WaitingThresholdResolver{
		globalConfig: globalConfig,
		vibeHome:     vibeHome,
		cliOverride:  cliOverride,
	}
}

// Resolve returns the effective waiting threshold for a project.
// Cascade: CLI flag > per-project config file > global config > default (10)
func (r *WaitingThresholdResolver) Resolve(projectID string) int {
	// 1. CLI flag takes highest priority (if set)
	if r.cliOverride >= 0 {
		return r.cliOverride
	}

	// 2. Per-project config file (~/.vibe-dash/<project>/config.yaml)
	projectDir := filepath.Join(r.vibeHome, projectID)
	loader, err := NewProjectConfigLoader(projectDir)
	if err == nil {
		data, loadErr := loader.Load(context.Background())
		if loadErr == nil && data.AgentWaitingThresholdMinutes != nil {
			return *data.AgentWaitingThresholdMinutes
		}
	}

	// 3. Global config
	if r.globalConfig != nil && r.globalConfig.AgentWaitingThresholdMinutes > 0 {
		return r.globalConfig.AgentWaitingThresholdMinutes
	}

	// 4. Default
	return defaultWaitingThreshold
}
