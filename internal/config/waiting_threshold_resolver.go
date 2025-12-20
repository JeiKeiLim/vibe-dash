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
	globalConfig    *ports.Config
	vibeHome        string     // ~/.vibe-dash path for project config resolution
	cliOverrideFunc func() int // Lazy evaluation - called at Resolve time after flags are parsed
}

// NewWaitingThresholdResolver creates a resolver with cascade support.
// cliOverrideFunc is called lazily at Resolve time to get the CLI override value.
// This allows flags to be parsed by Cobra before the value is read.
// Returns: -1 means "use config", 0 means "disabled", >0 means threshold minutes.
func NewWaitingThresholdResolver(
	globalConfig *ports.Config,
	vibeHome string,
	cliOverrideFunc func() int,
) *WaitingThresholdResolver {
	return &WaitingThresholdResolver{
		globalConfig:    globalConfig,
		vibeHome:        vibeHome,
		cliOverrideFunc: cliOverrideFunc,
	}
}

// Resolve returns the effective waiting threshold for a project.
// Cascade: CLI flag > per-project config file > global config > default (10)
func (r *WaitingThresholdResolver) Resolve(projectID string) int {
	// 1. CLI flag takes highest priority (if set)
	// Called lazily to ensure Cobra has parsed flags
	if r.cliOverrideFunc != nil {
		cliOverride := r.cliOverrideFunc()
		if cliOverride >= 0 {
			return cliOverride
		}
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
