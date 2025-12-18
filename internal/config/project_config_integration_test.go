//go:build integration

package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/config"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// TestIntegration_ProjectConfigLoader_FullLifecycle tests the full lifecycle:
// create loader → save → load → verify
func TestIntegration_ProjectConfigLoader_FullLifecycle(t *testing.T) {
	// Setup: Create temp directory with .project-path marker
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "test-project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create marker file (simulating DirectoryManager)
	markerPath := filepath.Join(projectDir, ".project-path")
	err = os.WriteFile(markerPath, []byte("/path/to/original/project"), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// Create loader
	loader, err := config.NewProjectConfigLoader(projectDir)
	require.NoError(t, err)

	// Initial load should create default config
	data, err := loader.Load(ctx)
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "", data.DetectedMethod)

	// Save with detection data
	hibernationDays := 7
	waitingThreshold := 5
	scanTime := time.Now().UTC().Truncate(time.Second)

	saveData := &ports.ProjectConfigData{
		DetectedMethod:               "speckit",
		LastScanned:                  scanTime,
		CustomHibernationDays:        &hibernationDays,
		AgentWaitingThresholdMinutes: &waitingThreshold,
		Notes:                        "Integration test project",
	}

	err = loader.Save(ctx, saveData)
	require.NoError(t, err)

	// Create new loader instance to simulate fresh load
	loader2, err := config.NewProjectConfigLoader(projectDir)
	require.NoError(t, err)

	// Load and verify
	loadedData, err := loader2.Load(ctx)
	require.NoError(t, err)

	assert.Equal(t, "speckit", loadedData.DetectedMethod)
	assert.Equal(t, scanTime, loadedData.LastScanned)
	require.NotNil(t, loadedData.CustomHibernationDays)
	assert.Equal(t, 7, *loadedData.CustomHibernationDays)
	require.NotNil(t, loadedData.AgentWaitingThresholdMinutes)
	assert.Equal(t, 5, *loadedData.AgentWaitingThresholdMinutes)
	assert.Equal(t, "Integration test project", loadedData.Notes)
}

// TestIntegration_ProjectConfigLoader_WithDirectoryManager tests integration
// with the real DirectoryManager.EnsureProjectDir()
func TestIntegration_ProjectConfigLoader_WithDirectoryManager(t *testing.T) {
	// Create temp base directory for vibe-dash storage
	tempBase := t.TempDir()

	// Create a fake project directory
	fakeProjectPath := filepath.Join(tempBase, "source-projects", "my-api-service")
	err := os.MkdirAll(fakeProjectPath, 0755)
	require.NoError(t, err)

	// Create DirectoryManager with temp base
	dm := filesystem.NewDirectoryManager(tempBase, nil)
	require.NotNil(t, dm)

	ctx := context.Background()

	// EnsureProjectDir creates the vibe-dash project directory with marker file
	projectDir, err := dm.EnsureProjectDir(ctx, fakeProjectPath)
	require.NoError(t, err)
	require.NotEmpty(t, projectDir)

	// Verify .project-path marker exists
	markerPath := filepath.Join(projectDir, ".project-path")
	_, err = os.Stat(markerPath)
	require.NoError(t, err, ".project-path marker should exist after EnsureProjectDir")

	// Now create ProjectConfigLoader for this directory
	loader, err := config.NewProjectConfigLoader(projectDir)
	require.NoError(t, err)

	// Load config (should create default)
	data, err := loader.Load(ctx)
	require.NoError(t, err)
	assert.NotNil(t, data)

	// Verify config.yaml was created
	configPath := filepath.Join(projectDir, "config.yaml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config.yaml should exist after Load")

	// Save detection data
	data.DetectedMethod = "speckit"
	data.LastScanned = time.Now().UTC().Truncate(time.Second)
	hibernation := 21
	data.CustomHibernationDays = &hibernation

	err = loader.Save(ctx, data)
	require.NoError(t, err)

	// Verify data persisted by creating new loader
	loader2, err := config.NewProjectConfigLoader(projectDir)
	require.NoError(t, err)

	loadedData, err := loader2.Load(ctx)
	require.NoError(t, err)

	assert.Equal(t, "speckit", loadedData.DetectedMethod)
	require.NotNil(t, loadedData.CustomHibernationDays)
	assert.Equal(t, 21, *loadedData.CustomHibernationDays)
}

// TestIntegration_ConfigResolver_WithRealConfigs tests ConfigResolver with
// real master and project configs
func TestIntegration_ConfigResolver_WithRealConfigs(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()

	// Create master config with custom values
	masterConfigPath := filepath.Join(tempDir, "config.yaml")
	masterContent := `settings:
  hibernation_days: 21
  refresh_interval_seconds: 15
  refresh_debounce_ms: 300
  agent_waiting_threshold_minutes: 12

projects: {}
`
	err := os.WriteFile(masterConfigPath, []byte(masterContent), 0644)
	require.NoError(t, err)

	// Load master config
	masterLoader := config.NewViperLoader(masterConfigPath)
	masterConfig, err := masterLoader.Load(ctx)
	require.NoError(t, err)

	// Verify master config loaded
	assert.Equal(t, 21, masterConfig.HibernationDays)
	assert.Equal(t, 12, masterConfig.AgentWaitingThresholdMinutes)

	// Create project directory with marker
	projectDir := filepath.Join(tempDir, "test-project")
	err = os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(projectDir, ".project-path"), []byte("/test/path"), 0644)
	require.NoError(t, err)

	// Create project config with partial overrides
	projectConfigPath := filepath.Join(projectDir, "config.yaml")
	projectContent := `detected_method: "bmad"
last_scanned: "2025-12-18T10:00:00Z"
custom_hibernation_days: 7
# agent_waiting_threshold_minutes not set - should use master config
notes: "Test project"
`
	err = os.WriteFile(projectConfigPath, []byte(projectContent), 0644)
	require.NoError(t, err)

	// Load project config
	projectLoader, err := config.NewProjectConfigLoader(projectDir)
	require.NoError(t, err)
	projectConfig, err := projectLoader.Load(ctx)
	require.NoError(t, err)

	// Verify project config loaded
	assert.Equal(t, "bmad", projectConfig.DetectedMethod)
	require.NotNil(t, projectConfig.CustomHibernationDays)
	assert.Equal(t, 7, *projectConfig.CustomHibernationDays)
	assert.Nil(t, projectConfig.AgentWaitingThresholdMinutes)

	// Test ConfigResolver cascade
	resolver := config.NewConfigResolver(masterConfig, projectConfig)

	// Hibernation should use project override (7)
	assert.Equal(t, 7, resolver.GetEffectiveHibernationDays(),
		"Project config should override master for hibernation_days")

	// Waiting threshold should fall back to master (12) since project doesn't override
	assert.Equal(t, 12, resolver.GetEffectiveWaitingThreshold(),
		"Master config should be used for waiting_threshold since project doesn't override")
}

// TestIntegration_ProjectConfigLoader_RejectionWithoutMarker verifies
// that ProjectConfigLoader refuses to work with directories not created by DirectoryManager
func TestIntegration_ProjectConfigLoader_RejectionWithoutMarker(t *testing.T) {
	tempDir := t.TempDir()

	// Try to create loader for directory without marker
	_, err := config.NewProjectConfigLoader(tempDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".project-path")
}
