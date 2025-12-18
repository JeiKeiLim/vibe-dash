package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/JeiKeiLim/vibe-dash/internal/config"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

func TestNewConfigResolver(t *testing.T) {
	masterConfig := ports.NewConfig()
	projectConfig := ports.NewProjectConfigData()

	resolver := config.NewConfigResolver(masterConfig, projectConfig)
	assert.NotNil(t, resolver)
}

func TestConfigResolver_GetEffectiveHibernationDays(t *testing.T) {
	tests := []struct {
		name           string
		masterConfig   *ports.Config
		projectConfig  *ports.ProjectConfigData
		expectedResult int
	}{
		{
			name: "project config override takes precedence",
			masterConfig: func() *ports.Config {
				c := ports.NewConfig()
				c.HibernationDays = 21 // Master config value
				return c
			}(),
			projectConfig: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				val := 7 // Project override
				d.CustomHibernationDays = &val
				return d
			}(),
			expectedResult: 7,
		},
		{
			name: "master config fallback when project config is nil",
			masterConfig: func() *ports.Config {
				c := ports.NewConfig()
				c.HibernationDays = 21
				return c
			}(),
			projectConfig:  nil,
			expectedResult: 21,
		},
		{
			name: "master config fallback when project override is nil",
			masterConfig: func() *ports.Config {
				c := ports.NewConfig()
				c.HibernationDays = 21
				return c
			}(),
			projectConfig: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				d.CustomHibernationDays = nil // Not overridden
				return d
			}(),
			expectedResult: 21,
		},
		{
			name:           "default fallback when both configs are nil",
			masterConfig:   nil,
			projectConfig:  nil,
			expectedResult: 14, // Built-in default
		},
		{
			name:         "default fallback when master config is nil",
			masterConfig: nil,
			projectConfig: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				d.CustomHibernationDays = nil
				return d
			}(),
			expectedResult: 14, // Built-in default
		},
		{
			name: "project override of zero is valid (disables hibernation)",
			masterConfig: func() *ports.Config {
				c := ports.NewConfig()
				c.HibernationDays = 14
				return c
			}(),
			projectConfig: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				val := 0 // Zero means disabled
				d.CustomHibernationDays = &val
				return d
			}(),
			expectedResult: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := config.NewConfigResolver(tt.masterConfig, tt.projectConfig)
			result := resolver.GetEffectiveHibernationDays()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestConfigResolver_GetEffectiveWaitingThreshold(t *testing.T) {
	tests := []struct {
		name           string
		masterConfig   *ports.Config
		projectConfig  *ports.ProjectConfigData
		expectedResult int
	}{
		{
			name: "project config override takes precedence",
			masterConfig: func() *ports.Config {
				c := ports.NewConfig()
				c.AgentWaitingThresholdMinutes = 15 // Master config value
				return c
			}(),
			projectConfig: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				val := 5 // Project override
				d.AgentWaitingThresholdMinutes = &val
				return d
			}(),
			expectedResult: 5,
		},
		{
			name: "master config fallback when project config is nil",
			masterConfig: func() *ports.Config {
				c := ports.NewConfig()
				c.AgentWaitingThresholdMinutes = 15
				return c
			}(),
			projectConfig:  nil,
			expectedResult: 15,
		},
		{
			name: "master config fallback when project override is nil",
			masterConfig: func() *ports.Config {
				c := ports.NewConfig()
				c.AgentWaitingThresholdMinutes = 15
				return c
			}(),
			projectConfig: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				d.AgentWaitingThresholdMinutes = nil // Not overridden
				return d
			}(),
			expectedResult: 15,
		},
		{
			name:           "default fallback when both configs are nil",
			masterConfig:   nil,
			projectConfig:  nil,
			expectedResult: 10, // Built-in default
		},
		{
			name:         "default fallback when master config is nil",
			masterConfig: nil,
			projectConfig: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				d.AgentWaitingThresholdMinutes = nil
				return d
			}(),
			expectedResult: 10, // Built-in default
		},
		{
			name: "project override of zero is valid (disables waiting detection)",
			masterConfig: func() *ports.Config {
				c := ports.NewConfig()
				c.AgentWaitingThresholdMinutes = 10
				return c
			}(),
			projectConfig: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				val := 0 // Zero means disabled
				d.AgentWaitingThresholdMinutes = &val
				return d
			}(),
			expectedResult: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := config.NewConfigResolver(tt.masterConfig, tt.projectConfig)
			result := resolver.GetEffectiveWaitingThreshold()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestConfigResolver_CascadeOrder(t *testing.T) {
	// AC3: Project config precedence
	// AC4: Master config fallback
	// AC5: Default fallback

	t.Run("full cascade: project -> master -> default", func(t *testing.T) {
		// Step 1: With all three layers
		hibernation := 7
		waiting := 5
		projectConfig := &ports.ProjectConfigData{
			CustomHibernationDays:        &hibernation,
			AgentWaitingThresholdMinutes: &waiting,
		}
		masterConfig := ports.NewConfig()
		masterConfig.HibernationDays = 21
		masterConfig.AgentWaitingThresholdMinutes = 15

		resolver := config.NewConfigResolver(masterConfig, projectConfig)

		// Project values should win
		assert.Equal(t, 7, resolver.GetEffectiveHibernationDays())
		assert.Equal(t, 5, resolver.GetEffectiveWaitingThreshold())
	})

	t.Run("partial cascade: no project override -> master", func(t *testing.T) {
		// Project config exists but doesn't override
		projectConfig := ports.NewProjectConfigData()
		masterConfig := ports.NewConfig()
		masterConfig.HibernationDays = 21
		masterConfig.AgentWaitingThresholdMinutes = 15

		resolver := config.NewConfigResolver(masterConfig, projectConfig)

		// Master values should be used
		assert.Equal(t, 21, resolver.GetEffectiveHibernationDays())
		assert.Equal(t, 15, resolver.GetEffectiveWaitingThreshold())
	})

	t.Run("no project config -> master", func(t *testing.T) {
		masterConfig := ports.NewConfig()
		masterConfig.HibernationDays = 21
		masterConfig.AgentWaitingThresholdMinutes = 15

		resolver := config.NewConfigResolver(masterConfig, nil)

		// Master values should be used
		assert.Equal(t, 21, resolver.GetEffectiveHibernationDays())
		assert.Equal(t, 15, resolver.GetEffectiveWaitingThreshold())
	})

	t.Run("no configs at all -> built-in defaults", func(t *testing.T) {
		resolver := config.NewConfigResolver(nil, nil)

		// Built-in defaults from ports.NewConfig()
		assert.Equal(t, 14, resolver.GetEffectiveHibernationDays())
		assert.Equal(t, 10, resolver.GetEffectiveWaitingThreshold())
	})
}
