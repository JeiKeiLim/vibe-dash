package ports_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

func TestNewProjectConfigData(t *testing.T) {
	data := ports.NewProjectConfigData()

	assert.NotNil(t, data)
	assert.Equal(t, "", data.DetectedMethod)
	assert.True(t, data.LastScanned.IsZero())
	assert.Nil(t, data.CustomHibernationDays)
	assert.Nil(t, data.AgentWaitingThresholdMinutes)
	assert.Equal(t, "", data.Notes)
}

func TestProjectConfigDataFields(t *testing.T) {
	// Test that all expected fields are accessible and settable
	now := time.Now().UTC()
	hibernationDays := 7
	waitingThreshold := 5

	data := &ports.ProjectConfigData{
		DetectedMethod:               "speckit",
		LastScanned:                  now,
		CustomHibernationDays:        &hibernationDays,
		AgentWaitingThresholdMinutes: &waitingThreshold,
		Notes:                        "Test project notes",
	}

	assert.Equal(t, "speckit", data.DetectedMethod)
	assert.Equal(t, now, data.LastScanned)
	assert.NotNil(t, data.CustomHibernationDays)
	assert.Equal(t, 7, *data.CustomHibernationDays)
	assert.NotNil(t, data.AgentWaitingThresholdMinutes)
	assert.Equal(t, 5, *data.AgentWaitingThresholdMinutes)
	assert.Equal(t, "Test project notes", data.Notes)
}

func TestProjectConfigDataOptionalFields(t *testing.T) {
	// Test that nil pointer fields work correctly for optional overrides
	data := ports.NewProjectConfigData()
	data.DetectedMethod = "bmad"

	assert.Equal(t, "bmad", data.DetectedMethod)
	assert.Nil(t, data.CustomHibernationDays, "CustomHibernationDays should be nil when not set")
	assert.Nil(t, data.AgentWaitingThresholdMinutes, "AgentWaitingThresholdMinutes should be nil when not set")
}

// Compile-time interface verification using a mock
type mockProjectConfigLoader struct{}

func (m *mockProjectConfigLoader) Load(_ context.Context) (*ports.ProjectConfigData, error) {
	return ports.NewProjectConfigData(), nil
}

func (m *mockProjectConfigLoader) Save(_ context.Context, _ *ports.ProjectConfigData) error {
	return nil
}

// Compile-time check that mockProjectConfigLoader implements ProjectConfigLoader
var _ ports.ProjectConfigLoader = (*mockProjectConfigLoader)(nil)

func TestProjectConfigData_Validate(t *testing.T) {
	tests := []struct {
		name        string
		data        *ports.ProjectConfigData
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid with nil pointers",
			data:        ports.NewProjectConfigData(),
			expectError: false,
		},
		{
			name: "valid with positive hibernation days",
			data: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				val := 7
				d.CustomHibernationDays = &val
				return d
			}(),
			expectError: false,
		},
		{
			name: "valid with zero hibernation days (disables feature)",
			data: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				val := 0
				d.CustomHibernationDays = &val
				return d
			}(),
			expectError: false,
		},
		{
			name: "invalid negative hibernation days",
			data: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				val := -5
				d.CustomHibernationDays = &val
				return d
			}(),
			expectError: true,
			errorMsg:    "custom_hibernation_days must be >= 0",
		},
		{
			name: "invalid negative waiting threshold",
			data: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				val := -10
				d.AgentWaitingThresholdMinutes = &val
				return d
			}(),
			expectError: true,
			errorMsg:    "agent_waiting_threshold_minutes must be >= 0",
		},
		{
			name: "valid with all fields set",
			data: func() *ports.ProjectConfigData {
				d := ports.NewProjectConfigData()
				d.DetectedMethod = "speckit"
				d.LastScanned = time.Now().UTC()
				hd := 21
				d.CustomHibernationDays = &hd
				awt := 5
				d.AgentWaitingThresholdMinutes = &awt
				d.Notes = "Test notes"
				return d
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
