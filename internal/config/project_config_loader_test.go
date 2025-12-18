package config_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JeiKeiLim/vibe-dash/internal/config"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Helper function to create a valid project directory with marker file
func setupProjectDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	markerPath := filepath.Join(dir, ".project-path")
	err := os.WriteFile(markerPath, []byte("/path/to/project"), 0644)
	require.NoError(t, err)
	return dir
}

func TestNewProjectConfigLoader(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError error
	}{
		{
			name: "valid directory with marker",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				err := os.WriteFile(filepath.Join(dir, ".project-path"), []byte("/path"), 0644)
				require.NoError(t, err)
				return dir
			},
			expectError: nil,
		},
		{
			name: "directory without marker",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			expectError: domain.ErrPathNotAccessible,
		},
		{
			name: "non-existent directory",
			setup: func(t *testing.T) string {
				return "/non/existent/path"
			},
			expectError: domain.ErrPathNotAccessible,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			_, err := config.NewProjectConfigLoader(dir)
			if tt.expectError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectConfigLoader_Load_CreatesDefaultConfig(t *testing.T) {
	dir := setupProjectDir(t)
	configPath := filepath.Join(dir, "config.yaml")

	// Ensure config doesn't exist initially
	_, err := os.Stat(configPath)
	require.True(t, os.IsNotExist(err), "config.yaml should not exist initially")

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := loader.Load(ctx)

	require.NoError(t, err)
	assert.NotNil(t, data)

	// Verify config file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config.yaml should be created after Load")

	// Verify default values
	assert.Equal(t, "", data.DetectedMethod)
	assert.True(t, data.LastScanned.IsZero())
	assert.Nil(t, data.CustomHibernationDays)
	assert.Nil(t, data.AgentWaitingThresholdMinutes)
	assert.Equal(t, "", data.Notes)
}

func TestProjectConfigLoader_Load_ParsesExistingConfig(t *testing.T) {
	dir := setupProjectDir(t)
	configPath := filepath.Join(dir, "config.yaml")

	// Create config with specific values
	configContent := `detected_method: "speckit"
last_scanned: "2025-12-18T10:30:00Z"
custom_hibernation_days: 7
agent_waiting_threshold_minutes: 5
notes: "Test project notes"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := loader.Load(ctx)

	require.NoError(t, err)
	assert.Equal(t, "speckit", data.DetectedMethod)

	expectedTime, _ := time.Parse(time.RFC3339, "2025-12-18T10:30:00Z")
	assert.Equal(t, expectedTime, data.LastScanned)

	require.NotNil(t, data.CustomHibernationDays)
	assert.Equal(t, 7, *data.CustomHibernationDays)

	require.NotNil(t, data.AgentWaitingThresholdMinutes)
	assert.Equal(t, 5, *data.AgentWaitingThresholdMinutes)

	assert.Equal(t, "Test project notes", data.Notes)
}

func TestProjectConfigLoader_Load_HandlesSyntaxErrors(t *testing.T) {
	dir := setupProjectDir(t)
	configPath := filepath.Join(dir, "config.yaml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: {"), 0644)
	require.NoError(t, err)

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := loader.Load(ctx)

	// Should return defaults on syntax error (graceful degradation)
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "", data.DetectedMethod)
}

func TestProjectConfigLoader_Load_HandlesInvalidValues(t *testing.T) {
	dir := setupProjectDir(t)
	configPath := filepath.Join(dir, "config.yaml")

	// Write config with invalid value types
	configContent := `detected_method: 123
custom_hibernation_days: "not a number"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := loader.Load(ctx)

	// Should return defaults on invalid values (graceful degradation)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestProjectConfigLoader_Load_ContextCancellation(t *testing.T) {
	dir := setupProjectDir(t)

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = loader.Load(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestProjectConfigLoader_Save(t *testing.T) {
	dir := setupProjectDir(t)

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	hibernationDays := 7
	waitingThreshold := 5
	now := time.Now().UTC().Truncate(time.Second)

	data := &ports.ProjectConfigData{
		DetectedMethod:               "bmad",
		LastScanned:                  now,
		CustomHibernationDays:        &hibernationDays,
		AgentWaitingThresholdMinutes: &waitingThreshold,
		Notes:                        "Project memo",
	}

	ctx := context.Background()
	err = loader.Save(ctx, data)
	require.NoError(t, err)

	// Verify by loading back
	loadedData, err := loader.Load(ctx)
	require.NoError(t, err)

	assert.Equal(t, "bmad", loadedData.DetectedMethod)
	assert.Equal(t, now, loadedData.LastScanned)
	require.NotNil(t, loadedData.CustomHibernationDays)
	assert.Equal(t, 7, *loadedData.CustomHibernationDays)
	require.NotNil(t, loadedData.AgentWaitingThresholdMinutes)
	assert.Equal(t, 5, *loadedData.AgentWaitingThresholdMinutes)
	assert.Equal(t, "Project memo", loadedData.Notes)
}

func TestProjectConfigLoader_Save_ContextCancellation(t *testing.T) {
	dir := setupProjectDir(t)

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	data := ports.NewProjectConfigData()
	err = loader.Save(ctx, data)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestProjectConfigLoader_Save_WithNilOptionalFields(t *testing.T) {
	dir := setupProjectDir(t)

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	// Save data with nil optional fields
	data := &ports.ProjectConfigData{
		DetectedMethod:               "speckit",
		CustomHibernationDays:        nil,
		AgentWaitingThresholdMinutes: nil,
	}

	ctx := context.Background()
	err = loader.Save(ctx, data)
	require.NoError(t, err)

	// Verify by loading back
	loadedData, err := loader.Load(ctx)
	require.NoError(t, err)

	assert.Equal(t, "speckit", loadedData.DetectedMethod)
	assert.Nil(t, loadedData.CustomHibernationDays)
	assert.Nil(t, loadedData.AgentWaitingThresholdMinutes)
}

// Compile-time check that ViperProjectConfigLoader implements ProjectConfigLoader
func TestViperProjectConfigLoader_ImplementsInterface(t *testing.T) {
	dir := setupProjectDir(t)
	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	// This will fail at compile time if the interface is not implemented
	var _ ports.ProjectConfigLoader = loader
}

func TestProjectConfigLoader_Load_EmptyOptionalFields(t *testing.T) {
	dir := setupProjectDir(t)
	configPath := filepath.Join(dir, "config.yaml")

	// Config with empty/missing optional fields
	configContent := `detected_method: "speckit"
notes: ""
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := loader.Load(ctx)

	require.NoError(t, err)
	assert.Equal(t, "speckit", data.DetectedMethod)
	assert.Nil(t, data.CustomHibernationDays, "Missing field should result in nil pointer")
	assert.Nil(t, data.AgentWaitingThresholdMinutes, "Missing field should result in nil pointer")
}

func TestProjectConfigLoader_ErrorWrapping(t *testing.T) {
	// Verify error wrapping uses domain errors
	_, err := config.NewProjectConfigLoader("/non/existent/path")
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrPathNotAccessible), "Should wrap with domain.ErrPathNotAccessible")
}

func TestProjectConfigLoader_Load_HandlesNegativeValues(t *testing.T) {
	// AC7: Invalid values (negative numbers) should be handled gracefully
	dir := setupProjectDir(t)
	configPath := filepath.Join(dir, "config.yaml")

	// Write config with negative override values
	configContent := `detected_method: "speckit"
custom_hibernation_days: -5
agent_waiting_threshold_minutes: -10
notes: "Test project"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := loader.Load(ctx)

	// Should succeed with graceful degradation
	require.NoError(t, err)
	assert.NotNil(t, data)

	// Invalid negative values should be removed (nil = use global)
	assert.Nil(t, data.CustomHibernationDays,
		"Negative custom_hibernation_days should be removed (nil = use global)")
	assert.Nil(t, data.AgentWaitingThresholdMinutes,
		"Negative agent_waiting_threshold_minutes should be removed (nil = use global)")

	// Valid values should be preserved
	assert.Equal(t, "speckit", data.DetectedMethod)
	assert.Equal(t, "Test project", data.Notes)
}

func TestProjectConfigLoader_ISO8601Timestamps(t *testing.T) {
	// AC6: ISO 8601 UTC timestamps per Architecture
	dir := setupProjectDir(t)
	configPath := filepath.Join(dir, "config.yaml")

	loader, err := config.NewProjectConfigLoader(dir)
	require.NoError(t, err)

	// Save a specific timestamp
	testTime := time.Date(2025, 12, 18, 10, 30, 0, 0, time.UTC)
	data := &ports.ProjectConfigData{
		DetectedMethod: "speckit",
		LastScanned:    testTime,
	}

	ctx := context.Background()
	err = loader.Save(ctx, data)
	require.NoError(t, err)

	// Read raw file content to verify ISO 8601 format
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	// Should contain RFC3339/ISO 8601 format timestamp
	assert.Contains(t, string(content), "2025-12-18T10:30:00Z",
		"Timestamp should be in ISO 8601 UTC format")

	// Verify it loads back correctly
	loadedData, err := loader.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, testTime, loadedData.LastScanned)
}
