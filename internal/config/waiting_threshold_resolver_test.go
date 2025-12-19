package config

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

func intPtr(v int) *int { return &v }

func TestWaitingThresholdResolver_CLIWinsOverAll(t *testing.T) {
	// CLI flag = 5, project config = 10, global config = 20
	// Expected: 5 (CLI wins)
	tmpDir := t.TempDir()

	// Create project directory and config
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write per-project config with threshold 10
	configContent := `agent_waiting_threshold_minutes: 10
`
	if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	globalConfig := &ports.Config{
		AgentWaitingThresholdMinutes: 20,
	}

	resolver := NewWaitingThresholdResolver(globalConfig, tmpDir, 5)
	result := resolver.Resolve("test-project")

	if result != 5 {
		t.Errorf("Resolve() = %d, want 5 (CLI should win)", result)
	}
}

func TestWaitingThresholdResolver_CLIDisabled(t *testing.T) {
	// CLI flag = 0 (disabled), project config = 10
	// Expected: 0 (CLI wins, detection disabled)
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `agent_waiting_threshold_minutes: 10
`
	if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	globalConfig := &ports.Config{
		AgentWaitingThresholdMinutes: 20,
	}

	resolver := NewWaitingThresholdResolver(globalConfig, tmpDir, 0)
	result := resolver.Resolve("test-project")

	if result != 0 {
		t.Errorf("Resolve() = %d, want 0 (CLI disabled should win)", result)
	}
}

func TestWaitingThresholdResolver_ProjectWins(t *testing.T) {
	// CLI flag = -1 (not set), project config = 15, global config = 20
	// Expected: 15 (project config wins)
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `agent_waiting_threshold_minutes: 15
`
	if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	globalConfig := &ports.Config{
		AgentWaitingThresholdMinutes: 20,
	}

	resolver := NewWaitingThresholdResolver(globalConfig, tmpDir, -1)
	result := resolver.Resolve("test-project")

	if result != 15 {
		t.Errorf("Resolve() = %d, want 15 (project config should win)", result)
	}
}

func TestWaitingThresholdResolver_GlobalWins(t *testing.T) {
	// CLI flag = -1 (not set), no project config, global config = 20
	// Expected: 20 (global config wins)
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// No per-project config file, or config without threshold override
	configContent := `notes: "test notes"
`
	if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	globalConfig := &ports.Config{
		AgentWaitingThresholdMinutes: 20,
	}

	resolver := NewWaitingThresholdResolver(globalConfig, tmpDir, -1)
	result := resolver.Resolve("test-project")

	if result != 20 {
		t.Errorf("Resolve() = %d, want 20 (global config should win)", result)
	}
}

func TestWaitingThresholdResolver_DefaultFallback(t *testing.T) {
	// CLI flag = -1 (not set), no project config, global config = 0
	// Expected: 10 (default)
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	globalConfig := &ports.Config{
		AgentWaitingThresholdMinutes: 0, // No global setting
	}

	resolver := NewWaitingThresholdResolver(globalConfig, tmpDir, -1)
	result := resolver.Resolve("test-project")

	if result != 10 {
		t.Errorf("Resolve() = %d, want 10 (default)", result)
	}
}

func TestWaitingThresholdResolver_NilGlobalConfig(t *testing.T) {
	// CLI flag = -1, nil global config
	// Expected: 10 (default)
	tmpDir := t.TempDir()

	resolver := NewWaitingThresholdResolver(nil, tmpDir, -1)
	result := resolver.Resolve("nonexistent-project")

	if result != 10 {
		t.Errorf("Resolve() = %d, want 10 (default with nil global config)", result)
	}
}

func TestWaitingThresholdResolver_ProjectConfigDisabled(t *testing.T) {
	// CLI flag = -1, project config = 0 (disabled), global = 20
	// Expected: 0 (project disabled wins over global)
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `agent_waiting_threshold_minutes: 0
`
	if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	globalConfig := &ports.Config{
		AgentWaitingThresholdMinutes: 20,
	}

	resolver := NewWaitingThresholdResolver(globalConfig, tmpDir, -1)
	result := resolver.Resolve("test-project")

	if result != 0 {
		t.Errorf("Resolve() = %d, want 0 (project disabled should win)", result)
	}
}

func TestWaitingThresholdResolver_NonexistentProjectDir(t *testing.T) {
	// CLI flag = -1, project dir doesn't exist, global = 20
	// Expected: 20 (global wins when project dir missing)
	tmpDir := t.TempDir()

	globalConfig := &ports.Config{
		AgentWaitingThresholdMinutes: 20,
	}

	resolver := NewWaitingThresholdResolver(globalConfig, tmpDir, -1)
	result := resolver.Resolve("nonexistent-project")

	if result != 20 {
		t.Errorf("Resolve() = %d, want 20 (global config when project missing)", result)
	}
}

func TestWaitingThresholdResolver_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		cliOverride   int
		projectConfig *int // nil = no per-project override
		globalConfig  int
		expected      int
	}{
		{"CLI wins over all", 5, intPtr(10), 20, 5},
		{"CLI disabled (0)", 0, intPtr(10), 20, 0},
		{"CLI not set, project wins", -1, intPtr(15), 20, 15},
		{"CLI not set, project disabled", -1, intPtr(0), 20, 0},
		{"CLI not set, no project, global wins", -1, nil, 20, 20},
		{"All defaults", -1, nil, 0, 10}, // default fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			projectDir := filepath.Join(tmpDir, "test-project")
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				t.Fatal(err)
			}

			// Write per-project config if threshold is specified
			if tt.projectConfig != nil {
				configContent := `agent_waiting_threshold_minutes: ` + strconv.Itoa(*tt.projectConfig) + "\n"
				if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte(configContent), 0644); err != nil {
					t.Fatal(err)
				}
			}

			globalConfig := &ports.Config{
				AgentWaitingThresholdMinutes: tt.globalConfig,
			}

			resolver := NewWaitingThresholdResolver(globalConfig, tmpDir, tt.cliOverride)
			result := resolver.Resolve("test-project")

			if result != tt.expected {
				t.Errorf("Resolve() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestWaitingThresholdResolver_ProjectConfigNotPresent_UsesGlobal(t *testing.T) {
	// CLI flag = -1 (not set), project config file exists but WITHOUT threshold field, global = 25
	// Expected: 25 (global wins when project config exists but doesn't have threshold)
	// This tests the distinction between nil (not present) vs *int = 0 (explicitly disabled)
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Config file exists but threshold field is NOT present (different from threshold: 0)
	configContent := `notes: "some project notes"
active: true
`
	if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	globalConfig := &ports.Config{
		AgentWaitingThresholdMinutes: 25,
	}

	resolver := NewWaitingThresholdResolver(globalConfig, tmpDir, -1)
	result := resolver.Resolve("test-project")

	if result != 25 {
		t.Errorf("Resolve() = %d, want 25 (global should win when project config exists but has no threshold)", result)
	}
}
