package config

// TODO(story:3.5.7): Update test fixture to use per-project storage structure at ~/.vibe-dash/<project>/state.db
// Current tests assume single master config.yaml structure. After Epic 3.5 completes,
// tests should verify master index + per-project config.yaml files work correctly.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// TestViperLoader_Load_CreatesDirectoryAndFile tests AC1: First run creates directory and file
func TestViperLoader_Load_CreatesDirectoryAndFile(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".vibe-dash", "config.yaml")

	loader := NewViperLoader(configPath)

	ctx := context.Background()
	cfg, err := loader.Load(ctx)

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(configPath)); os.IsNotExist(err) {
		t.Error("config directory was not created")
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Verify defaults
	if cfg.HibernationDays != 14 {
		t.Errorf("HibernationDays = %d, want 14", cfg.HibernationDays)
	}
	if cfg.RefreshIntervalSeconds != 10 {
		t.Errorf("RefreshIntervalSeconds = %d, want 10", cfg.RefreshIntervalSeconds)
	}
	if cfg.RefreshDebounceMs != 200 {
		t.Errorf("RefreshDebounceMs = %d, want 200", cfg.RefreshDebounceMs)
	}
	if cfg.AgentWaitingThresholdMinutes != 10 {
		t.Errorf("AgentWaitingThresholdMinutes = %d, want 10", cfg.AgentWaitingThresholdMinutes)
	}
}

// TestViperLoader_Load_FileContentCorrect verifies the created file has correct content
func TestViperLoader_Load_FileContentCorrect(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".vibe-dash", "config.yaml")

	loader := NewViperLoader(configPath)
	_, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Read created file
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Verify expected content
	contentStr := string(content)
	expectedPhrases := []string{
		"hibernation_days: 14",
		"refresh_interval_seconds: 10",
		"refresh_debounce_ms: 200",
		"agent_waiting_threshold_minutes: 10",
		"projects: {}",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(contentStr, phrase) {
			t.Errorf("config file missing expected content: %s", phrase)
		}
	}
}

// TestViperLoader_Load_PreservesExistingConfig tests AC2: Existing config is preserved
func TestViperLoader_Load_PreservesExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write custom config with user's values
	customYAML := `settings:
  hibernation_days: 30
  refresh_interval_seconds: 5
  refresh_debounce_ms: 100
  agent_waiting_threshold_minutes: 15
`
	if err := os.WriteFile(configPath, []byte(customYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should preserve user's custom values, not overwrite with defaults
	if cfg.HibernationDays != 30 {
		t.Errorf("HibernationDays = %d, want 30 (user value)", cfg.HibernationDays)
	}
	if cfg.RefreshIntervalSeconds != 5 {
		t.Errorf("RefreshIntervalSeconds = %d, want 5 (user value)", cfg.RefreshIntervalSeconds)
	}
	if cfg.RefreshDebounceMs != 100 {
		t.Errorf("RefreshDebounceMs = %d, want 100 (user value)", cfg.RefreshDebounceMs)
	}
	if cfg.AgentWaitingThresholdMinutes != 15 {
		t.Errorf("AgentWaitingThresholdMinutes = %d, want 15 (user value)", cfg.AgentWaitingThresholdMinutes)
	}
}

// TestViperLoader_Load_SyntaxError tests AC3: Syntax errors result in defaults
// AC3 states: "error is reported with line number if possible"
// Viper/YAML parser includes line info in error messages when available.
// The warning is logged via slog.Warn with the error details.
func TestViperLoader_Load_SyntaxError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write invalid YAML - this produces a parse error with line info
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)

	ctx := context.Background()
	cfg, err := loader.Load(ctx)

	// Should NOT return error - graceful degradation
	if err != nil {
		t.Errorf("Load() should not return error on syntax error, got %v", err)
	}

	// Should return defaults
	if cfg.HibernationDays != 14 {
		t.Errorf("HibernationDays = %d, want 14 (default)", cfg.HibernationDays)
	}
	if cfg.RefreshIntervalSeconds != 10 {
		t.Errorf("RefreshIntervalSeconds = %d, want 10 (default)", cfg.RefreshIntervalSeconds)
	}
}

// TestViperLoader_Load_SyntaxErrorWithLineNumber tests AC3 with multi-line YAML
// This ensures errors on specific lines are handled gracefully
func TestViperLoader_Load_SyntaxErrorWithLineNumber(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write multi-line YAML with error on line 4
	invalidYAML := `settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  invalid_line: [unclosed bracket
  agent_waiting_threshold_minutes: 10
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())

	// Should NOT return error - graceful degradation (AC3: exit code 0)
	if err != nil {
		t.Errorf("Load() should not return error on syntax error, got %v", err)
	}

	// Should return defaults since parsing failed
	if cfg.HibernationDays != 14 {
		t.Errorf("HibernationDays = %d, want 14 (default)", cfg.HibernationDays)
	}
}

// TestViperLoader_Load_InvalidValues tests AC4: Invalid values use defaults
func TestViperLoader_Load_InvalidValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write config with invalid values
	invalidYAML := `settings:
  hibernation_days: -5
  refresh_interval_seconds: 0
  refresh_debounce_ms: -100
  agent_waiting_threshold_minutes: -1
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())

	// Should NOT return error - graceful degradation
	if err != nil {
		t.Errorf("Load() should not return error on invalid values, got %v", err)
	}

	// Should use defaults for all invalid fields
	if cfg.HibernationDays != 14 {
		t.Errorf("HibernationDays = %d, want 14 (default)", cfg.HibernationDays)
	}
	if cfg.RefreshIntervalSeconds != 10 {
		t.Errorf("RefreshIntervalSeconds = %d, want 10 (default)", cfg.RefreshIntervalSeconds)
	}
	if cfg.RefreshDebounceMs != 200 {
		t.Errorf("RefreshDebounceMs = %d, want 200 (default)", cfg.RefreshDebounceMs)
	}
	if cfg.AgentWaitingThresholdMinutes != 10 {
		t.Errorf("AgentWaitingThresholdMinutes = %d, want 10 (default)", cfg.AgentWaitingThresholdMinutes)
	}
}

// TestViperLoader_Load_PartialInvalidValues tests that valid values are preserved
func TestViperLoader_Load_PartialInvalidValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write config with some invalid values
	mixedYAML := `settings:
  hibernation_days: 21
  refresh_interval_seconds: -5
  refresh_debounce_ms: 150
  agent_waiting_threshold_minutes: -2
`
	if err := os.WriteFile(configPath, []byte(mixedYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())

	if err != nil {
		t.Errorf("Load() should not return error, got %v", err)
	}

	// Valid values should be preserved
	if cfg.HibernationDays != 21 {
		t.Errorf("HibernationDays = %d, want 21 (valid user value)", cfg.HibernationDays)
	}
	if cfg.RefreshDebounceMs != 150 {
		t.Errorf("RefreshDebounceMs = %d, want 150 (valid user value)", cfg.RefreshDebounceMs)
	}

	// Invalid values should use defaults
	if cfg.RefreshIntervalSeconds != 10 {
		t.Errorf("RefreshIntervalSeconds = %d, want 10 (default)", cfg.RefreshIntervalSeconds)
	}
	if cfg.AgentWaitingThresholdMinutes != 10 {
		t.Errorf("AgentWaitingThresholdMinutes = %d, want 10 (default)", cfg.AgentWaitingThresholdMinutes)
	}
}

// TestViperLoader_Load_ContextCancellation tests context cancellation
func TestViperLoader_Load_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	loader := NewViperLoader(configPath)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := loader.Load(ctx)

	if err == nil {
		t.Error("Load() should return error on cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("Load() error = %v, want context.Canceled", err)
	}
}

// TestViperLoader_Save tests Save() method persists config correctly
func TestViperLoader_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".vibe-dash")
	configPath := filepath.Join(configDir, "config.yaml")

	// Create directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	loader := NewViperLoader(configPath)

	// Create and save custom config
	cfg := ports.NewConfig()
	cfg.HibernationDays = 21
	cfg.RefreshIntervalSeconds = 15

	err := loader.Save(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Create new loader and verify saved values persist
	loader2 := NewViperLoader(configPath)
	cfg2, err := loader2.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if cfg2.HibernationDays != 21 {
		t.Errorf("HibernationDays = %d, want 21", cfg2.HibernationDays)
	}
	if cfg2.RefreshIntervalSeconds != 15 {
		t.Errorf("RefreshIntervalSeconds = %d, want 15", cfg2.RefreshIntervalSeconds)
	}
}

// TestViperLoader_Save_ContextCancellation tests Save with cancelled context
func TestViperLoader_Save_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	loader := NewViperLoader(configPath)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := loader.Save(ctx, ports.NewConfig())

	if err == nil {
		t.Error("Save() should return error on cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("Save() error = %v, want context.Canceled", err)
	}
}

// TestViperLoader_Load_WithProjects tests loading config with project entries
func TestViperLoader_Load_WithProjects(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write config with projects
	configWithProjects := `settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects:
  project-1:
    path: /home/user/projects/myapp
    display_name: My Application
    favorite: true
  project-2:
    path: /home/user/projects/other
`
	if err := os.WriteFile(configPath, []byte(configWithProjects), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify projects loaded correctly
	if len(cfg.Projects) != 2 {
		t.Errorf("Projects count = %d, want 2", len(cfg.Projects))
	}

	p1, ok := cfg.Projects["project-1"]
	if !ok {
		t.Fatal("project-1 not found in config")
	}
	if p1.Path != "/home/user/projects/myapp" {
		t.Errorf("project-1 Path = %s, want /home/user/projects/myapp", p1.Path)
	}
	if p1.DisplayName != "My Application" {
		t.Errorf("project-1 DisplayName = %s, want 'My Application'", p1.DisplayName)
	}
	if !p1.IsFavorite {
		t.Error("project-1 IsFavorite = false, want true")
	}

	p2, ok := cfg.Projects["project-2"]
	if !ok {
		t.Fatal("project-2 not found in config")
	}
	if p2.Path != "/home/user/projects/other" {
		t.Errorf("project-2 Path = %s, want /home/user/projects/other", p2.Path)
	}
}

// TestViperLoader_Save_WithProjects tests saving config with projects
func TestViperLoader_Save_WithProjects(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".vibe-dash")
	configPath := filepath.Join(configDir, "config.yaml")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	loader := NewViperLoader(configPath)

	cfg := ports.NewConfig()
	cfg.Projects["test-project"] = ports.ProjectConfig{
		Path:        "/test/path",
		DisplayName: "Test Project",
		IsFavorite:  true,
	}

	if err := loader.Save(context.Background(), cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reload and verify
	loader2 := NewViperLoader(configPath)
	cfg2, err := loader2.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	p, ok := cfg2.Projects["test-project"]
	if !ok {
		t.Fatal("test-project not found after reload")
	}
	if p.Path != "/test/path" {
		t.Errorf("Path = %s, want /test/path", p.Path)
	}
	if p.DisplayName != "Test Project" {
		t.Errorf("DisplayName = %s, want 'Test Project'", p.DisplayName)
	}
}

// TestViperLoader_Load_UnwritableDirectory tests AC5: unwritable directory handling
func TestViperLoader_Load_UnwritableDirectory(t *testing.T) {
	// Skip on Windows where permission model is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("skipping permission test on Windows")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".vibe-dash")

	// Create directory first
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Make directory read-only (no write permission)
	if err := os.Chmod(configDir, 0555); err != nil {
		t.Fatalf("failed to chmod config dir: %v", err)
	}

	// Cleanup: restore permissions so t.TempDir() can clean up
	t.Cleanup(func() {
		_ = os.Chmod(configDir, 0755) // Best effort, ignore error in cleanup
	})

	configPath := filepath.Join(configDir, "config.yaml")
	loader := NewViperLoader(configPath)

	ctx := context.Background()
	cfg, err := loader.Load(ctx)

	// Should NOT return error - graceful degradation (AC5)
	if err != nil {
		t.Errorf("Load() should not return error on unwritable directory, got %v", err)
	}

	// Should return defaults
	if cfg.HibernationDays != 14 {
		t.Errorf("HibernationDays = %d, want 14 (default)", cfg.HibernationDays)
	}

	// Config file should NOT have been created (directory is read-only)
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("config file should not exist in read-only directory")
	}
}

// TestViperLoader_Load_EmptyFile tests loading empty config file
func TestViperLoader_Load_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write empty file
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())

	// Should return defaults for empty file
	if err != nil {
		t.Errorf("Load() should not return error on empty file, got %v", err)
	}

	// Verify defaults are used
	if cfg.HibernationDays != 14 {
		t.Errorf("HibernationDays = %d, want 14 (default)", cfg.HibernationDays)
	}
}

// TestGetDefaultConfigPath tests the path helper function
func TestGetDefaultConfigPath(t *testing.T) {
	path := GetDefaultConfigPath()

	// Should end with expected path components
	if !strings.Contains(path, ".vibe-dash") {
		t.Errorf("GetDefaultConfigPath() = %s, should contain '.vibe-dash'", path)
	}
	if !strings.Contains(path, "config.yaml") {
		t.Errorf("GetDefaultConfigPath() = %s, should contain 'config.yaml'", path)
	}
}

// TestGetConfigDir tests the config directory helper function
func TestGetConfigDir(t *testing.T) {
	dir := GetConfigDir()

	// Should end with expected directory name
	if !strings.Contains(dir, ".vibe-dash") {
		t.Errorf("GetConfigDir() = %s, should contain '.vibe-dash'", dir)
	}
}

// TestNewViperLoader_DefaultPath tests that empty path uses default
func TestNewViperLoader_DefaultPath(t *testing.T) {
	loader := NewViperLoader("")

	// Should use default path
	expected := GetDefaultConfigPath()
	if loader.configPath != expected {
		t.Errorf("configPath = %s, want %s", loader.configPath, expected)
	}
}

// TestNewViperLoader_CustomPath tests custom path is used
func TestNewViperLoader_CustomPath(t *testing.T) {
	customPath := "/custom/path/config.yaml"
	loader := NewViperLoader(customPath)

	if loader.configPath != customPath {
		t.Errorf("configPath = %s, want %s", loader.configPath, customPath)
	}
}
