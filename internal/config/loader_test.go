package config

// TODO(story:3.5.7): Update test fixture to use per-project storage structure at ~/.vibe-dash/<project>/state.db
// Current tests assume single master config.yaml structure. After Epic 3.5 completes,
// tests should verify master index + per-project config.yaml files work correctly.

import (
	"context"
	"fmt"
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
		"storage_version: 2", // Must be present for v2 format (Task 4)
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

	// Verify storage_version is near the top (before settings)
	storageVersionIdx := strings.Index(contentStr, "storage_version: 2")
	settingsIdx := strings.Index(contentStr, "settings:")
	if storageVersionIdx == -1 {
		t.Error("storage_version not found in config file")
	} else if settingsIdx != -1 && storageVersionIdx > settingsIdx {
		t.Error("storage_version should appear before settings section")
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

// TestViperLoader_Load_WithProjects tests loading config with project entries (v2 format)
func TestViperLoader_Load_WithProjects(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write v2 config with projects (directory_name as key)
	configWithProjects := `storage_version: 2

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects:
  myapp:
    path: /home/user/projects/myapp
    display_name: My Application
    favorite: true
  other:
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

	p1, ok := cfg.Projects["myapp"]
	if !ok {
		t.Fatal("myapp not found in config")
	}
	if p1.Path != "/home/user/projects/myapp" {
		t.Errorf("myapp Path = %s, want /home/user/projects/myapp", p1.Path)
	}
	if p1.DirectoryName != "myapp" {
		t.Errorf("myapp DirectoryName = %s, want myapp", p1.DirectoryName)
	}
	if p1.DisplayName != "My Application" {
		t.Errorf("myapp DisplayName = %s, want 'My Application'", p1.DisplayName)
	}
	if !p1.IsFavorite {
		t.Error("myapp IsFavorite = false, want true")
	}

	p2, ok := cfg.Projects["other"]
	if !ok {
		t.Fatal("other not found in config")
	}
	if p2.Path != "/home/user/projects/other" {
		t.Errorf("other Path = %s, want /home/user/projects/other", p2.Path)
	}
	if p2.DirectoryName != "other" {
		t.Errorf("other DirectoryName = %s, want other", p2.DirectoryName)
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

// Story 3.5.4 Tests: Master Config as Path Index

// TestViperLoader_Load_ReadsStorageVersion tests reading storage_version from YAML (Subtask 2.1)
func TestViperLoader_Load_ReadsStorageVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write v2 config with storage_version
	v2Config := `storage_version: 2

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects:
  api-service:
    path: "/home/user/api-service"
    directory_name: api-service
    favorite: false
`
	if err := os.WriteFile(configPath, []byte(v2Config), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.StorageVersion != 2 {
		t.Errorf("StorageVersion = %d, want 2", cfg.StorageVersion)
	}
}

// TestViperLoader_Load_V2Format_DirectoryNameAsKey tests v2 format with directory_name as map key (Subtask 2.2)
func TestViperLoader_Load_V2Format_DirectoryNameAsKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write v2 config where key is directory_name
	v2Config := `storage_version: 2

settings:
  hibernation_days: 14

projects:
  api-service:
    path: "/home/user/api-service"
    favorite: false
  client-b-api-service:
    path: "/home/user/client-b/api-service"
    display_name: "Client B API"
    favorite: true
`
	if err := os.WriteFile(configPath, []byte(v2Config), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check first project
	p1, ok := cfg.Projects["api-service"]
	if !ok {
		t.Fatal("api-service not found in projects")
	}
	if p1.Path != "/home/user/api-service" {
		t.Errorf("api-service Path = %s, want /home/user/api-service", p1.Path)
	}
	if p1.DirectoryName != "api-service" {
		t.Errorf("api-service DirectoryName = %s, want api-service", p1.DirectoryName)
	}

	// Check second project
	p2, ok := cfg.Projects["client-b-api-service"]
	if !ok {
		t.Fatal("client-b-api-service not found in projects")
	}
	if p2.Path != "/home/user/client-b/api-service" {
		t.Errorf("client-b-api-service Path = %s, want /home/user/client-b/api-service", p2.Path)
	}
	if p2.DirectoryName != "client-b-api-service" {
		t.Errorf("client-b-api-service DirectoryName = %s, want client-b-api-service", p2.DirectoryName)
	}
	if p2.DisplayName != "Client B API" {
		t.Errorf("client-b-api-service DisplayName = %s, want 'Client B API'", p2.DisplayName)
	}
}

// TestViperLoader_Save_WritesStorageVersion tests Save() writes storage_version at root (Subtask 2.3)
func TestViperLoader_Save_WritesStorageVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	loader := NewViperLoader(configPath)

	cfg := ports.NewConfig()
	cfg.Projects["api-service"] = ports.ProjectConfig{
		Path:          "/path/to/api",
		DirectoryName: "api-service",
	}

	err := loader.Save(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read raw YAML and verify storage_version present
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "storage_version: 2") {
		t.Errorf("config file should contain 'storage_version: 2', got:\n%s", content)
	}
}

// TestViperLoader_Save_WritesDirectoryName tests Save() writes directory_name for each project (Subtask 2.4)
func TestViperLoader_Save_WritesDirectoryName(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	loader := NewViperLoader(configPath)

	cfg := ports.NewConfig()
	cfg.Projects["api-service"] = ports.ProjectConfig{
		Path:          "/path/to/api-service",
		DirectoryName: "api-service",
		IsFavorite:    true,
	}
	cfg.Projects["client-b-api-service"] = ports.ProjectConfig{
		Path:          "/path/to/client-b/api-service",
		DirectoryName: "client-b-api-service",
		DisplayName:   "Client B API",
	}

	err := loader.Save(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reload and verify directory names are preserved
	loader2 := NewViperLoader(configPath)
	cfg2, err := loader2.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	p1, ok := cfg2.Projects["api-service"]
	if !ok {
		t.Fatal("api-service not found after reload")
	}
	if p1.DirectoryName != "api-service" {
		t.Errorf("api-service DirectoryName = %s, want api-service", p1.DirectoryName)
	}
	if p1.Path != "/path/to/api-service" {
		t.Errorf("api-service Path = %s, want /path/to/api-service", p1.Path)
	}

	p2, ok := cfg2.Projects["client-b-api-service"]
	if !ok {
		t.Fatal("client-b-api-service not found after reload")
	}
	if p2.DirectoryName != "client-b-api-service" {
		t.Errorf("client-b-api-service DirectoryName = %s, want client-b-api-service", p2.DirectoryName)
	}
}

// TestViperLoader_Load_MigratesV1ToV2 tests migration from v1 format (Subtask 2.5)
func TestViperLoader_Load_MigratesV1ToV2(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write v1 format (no storage_version, arbitrary project key)
	v1Config := `settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects:
  old-arbitrary-key:
    path: "/home/user/my-project"
    favorite: true
    hibernation_days: 7
`
	if err := os.WriteFile(configPath, []byte(v1Config), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should be migrated to v2
	if cfg.StorageVersion != 2 {
		t.Errorf("StorageVersion = %d, want 2 (migrated)", cfg.StorageVersion)
	}

	// Project should be migrated with base name as key
	if _, ok := cfg.Projects["my-project"]; !ok {
		// List actual keys
		var keys []string
		for k := range cfg.Projects {
			keys = append(keys, k)
		}
		t.Fatalf("my-project not found in migrated projects, got keys: %v", keys)
	}

	p := cfg.Projects["my-project"]
	if p.Path != "/home/user/my-project" {
		t.Errorf("Path = %s, want /home/user/my-project", p.Path)
	}
	if p.DirectoryName != "my-project" {
		t.Errorf("DirectoryName = %s, want my-project", p.DirectoryName)
	}
	if !p.IsFavorite {
		t.Error("IsFavorite = false, want true")
	}
}

// TestViperLoader_Load_MigratesV1_CollisionWarning tests migration with collision (Subtask 2.5)
func TestViperLoader_Load_MigratesV1_CollisionWarning(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write v1 format with two projects that would collide (same base name)
	v1Config := `settings:
  hibernation_days: 14

projects:
  project-a:
    path: "/home/user/work/api-service"
    favorite: false
  project-b:
    path: "/home/user/client/api-service"
    favorite: true
`
	if err := os.WriteFile(configPath, []byte(v1Config), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// StorageVersion should be 2
	if cfg.StorageVersion != 2 {
		t.Errorf("StorageVersion = %d, want 2", cfg.StorageVersion)
	}

	// Only one project should remain (collision - second one skipped)
	if len(cfg.Projects) != 1 {
		t.Errorf("Projects count = %d, want 1 (one should be skipped due to collision)", len(cfg.Projects))
	}

	// api-service should be the key
	if _, ok := cfg.Projects["api-service"]; !ok {
		t.Error("api-service not found in migrated projects")
	}
}

// TestViperLoader_Load_MigratesV1_EmptyPathSkipped tests migration skips entries with empty path (Subtask 2.5)
func TestViperLoader_Load_MigratesV1_EmptyPathSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write v1 format with empty path
	v1Config := `settings:
  hibernation_days: 14

projects:
  project-with-no-path:
    path: ""
    favorite: false
  valid-project:
    path: "/home/user/valid-project"
    favorite: true
`
	if err := os.WriteFile(configPath, []byte(v1Config), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Only valid project should remain
	if len(cfg.Projects) != 1 {
		t.Errorf("Projects count = %d, want 1 (empty path should be skipped)", len(cfg.Projects))
	}

	if _, ok := cfg.Projects["valid-project"]; !ok {
		t.Error("valid-project not found in migrated projects")
	}
}

// Story 3.5.4 Integration Tests

// TestViperLoader_Integration_FullLifecycle tests add project → save → load → verify (Subtask 7.1)
func TestViperLoader_Integration_FullLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Step 1: Create new config and add projects
	loader := NewViperLoader(configPath)
	cfg := ports.NewConfig()

	// Add first project
	cfg.SetProjectEntry("api-service", "/home/user/api-service", "", false)

	// Add second project with display name
	cfg.SetProjectEntry("client-b-api-service", "/home/user/client-b/api-service", "Client B API", true)

	// Step 2: Save
	err := loader.Save(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Step 3: Create fresh loader and load
	loader2 := NewViperLoader(configPath)
	cfg2, err := loader2.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Step 4: Verify storage_version
	if cfg2.StorageVersion != 2 {
		t.Errorf("StorageVersion = %d, want 2", cfg2.StorageVersion)
	}

	// Step 5: Verify projects count
	if len(cfg2.Projects) != 2 {
		t.Errorf("Projects count = %d, want 2", len(cfg2.Projects))
	}

	// Step 6: Verify first project via directory mapping
	dirName1, found := cfg2.GetDirectoryName("/home/user/api-service")
	if !found {
		t.Fatal("api-service path lookup failed")
	}
	if dirName1 != "api-service" {
		t.Errorf("GetDirectoryName() = %s, want api-service", dirName1)
	}

	// Step 7: Verify reverse lookup
	path1, found := cfg2.GetProjectPath("api-service")
	if !found {
		t.Fatal("api-service directory lookup failed")
	}
	if path1 != "/home/user/api-service" {
		t.Errorf("GetProjectPath() = %s, want /home/user/api-service", path1)
	}

	// Step 8: Verify second project with display name
	p2 := cfg2.Projects["client-b-api-service"]
	if p2.DisplayName != "Client B API" {
		t.Errorf("DisplayName = %s, want 'Client B API'", p2.DisplayName)
	}
	if !p2.IsFavorite {
		t.Error("IsFavorite = false, want true")
	}
}

// TestViperLoader_Integration_RemoveProject tests remove project → save → load (Subtask 7.2)
func TestViperLoader_Integration_RemoveProject(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Step 1: Create config with two projects
	loader := NewViperLoader(configPath)
	cfg := ports.NewConfig()
	cfg.SetProjectEntry("api-service", "/home/user/api-service", "", false)
	cfg.SetProjectEntry("web-service", "/home/user/web-service", "Web App", true)

	// Save initial state
	err := loader.Save(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Step 2: Remove one project
	removed := cfg.RemoveProject("api-service")
	if !removed {
		t.Fatal("RemoveProject() returned false")
	}

	// Step 3: Save after removal
	err = loader.Save(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Save() after removal error = %v", err)
	}

	// Step 4: Fresh load and verify
	loader2 := NewViperLoader(configPath)
	cfg2, err := loader2.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Step 5: Verify project count
	if len(cfg2.Projects) != 1 {
		t.Errorf("Projects count = %d, want 1", len(cfg2.Projects))
	}

	// Step 6: Verify removed project is gone
	if _, found := cfg2.GetProjectPath("api-service"); found {
		t.Error("api-service should have been removed")
	}

	// Step 7: Verify remaining project
	if _, found := cfg2.GetProjectPath("web-service"); !found {
		t.Error("web-service should still exist")
	}
}

// TestConfig_ProjectPathLookup_DeterminismWithDirectoryManager tests Config implements ProjectPathLookup (Subtask 7.3)
func TestConfig_ProjectPathLookup_DeterminismWithDirectoryManager(t *testing.T) {
	// This test verifies that Config satisfies the ProjectPathLookup interface
	// and that the same path always returns the same directory name (determinism)

	cfg := ports.NewConfig()

	// Add project entries
	cfg.SetProjectEntry("api-service", "/home/user/api-service", "", false)
	cfg.SetProjectEntry("client-b-api-service", "/home/user/client-b/api-service", "", false)

	// Test that Config satisfies ProjectPathLookup interface
	var lookup ports.ProjectPathLookup = cfg

	// Test 1: Same path should always return same directory name (determinism)
	for i := 0; i < 100; i++ {
		dirName := lookup.GetDirForPath("/home/user/api-service")
		if dirName != "api-service" {
			t.Errorf("iteration %d: GetDirForPath() = %s, want api-service", i, dirName)
		}
	}

	// Test 2: Different paths should return different directory names
	dir1 := lookup.GetDirForPath("/home/user/api-service")
	dir2 := lookup.GetDirForPath("/home/user/client-b/api-service")
	if dir1 == dir2 {
		t.Errorf("Different paths returned same directory: %s", dir1)
	}

	// Test 3: Non-registered path returns empty string
	dirName := lookup.GetDirForPath("/non/existent/path")
	if dirName != "" {
		t.Errorf("Non-existent path should return empty string, got %s", dirName)
	}
}

// TestViperLoader_Save_DoesNotWriteDeprecatedFields tests Save() omits deprecated fields (Subtask 5.3)
func TestViperLoader_Save_DoesNotWriteDeprecatedFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	loader := NewViperLoader(configPath)

	// Create config with deprecated fields set
	hibernationDays := 7
	waitingThreshold := 5
	cfg := ports.NewConfig()
	cfg.Projects["api-service"] = ports.ProjectConfig{
		Path:                         "/path/to/api",
		DirectoryName:                "api-service",
		HibernationDays:              &hibernationDays,  // DEPRECATED
		AgentWaitingThresholdMinutes: &waitingThreshold, // DEPRECATED
	}

	err := loader.Save(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reload and verify deprecated fields are NOT present in project config
	loader2 := NewViperLoader(configPath)
	cfg2, err := loader2.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	pc, ok := cfg2.Projects["api-service"]
	if !ok {
		t.Fatal("api-service not found after reload")
	}

	// Deprecated fields should be nil after reload (not written, so not read back)
	if pc.HibernationDays != nil {
		t.Errorf("HibernationDays should be nil after reload, got %v", *pc.HibernationDays)
	}
	if pc.AgentWaitingThresholdMinutes != nil {
		t.Errorf("AgentWaitingThresholdMinutes should be nil after reload, got %v", *pc.AgentWaitingThresholdMinutes)
	}
}

// TestViperLoader_Load_FixesInvalidStorageVersion tests fixInvalidValues handles invalid storage_version (Subtask 2.6)
func TestViperLoader_Load_FixesInvalidStorageVersion(t *testing.T) {
	tests := []struct {
		name           string
		storageVersion int
	}{
		{"storage_version_3", 3},
		{"storage_version_99", 99},
		{"storage_version_negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			// Write config with invalid storage_version
			configContent := `storage_version: ` + fmt.Sprintf("%d", tt.storageVersion) + `

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects: {}
`
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			loader := NewViperLoader(configPath)
			cfg, err := loader.Load(context.Background())
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Should be fixed to 2
			if cfg.StorageVersion != 2 {
				t.Errorf("StorageVersion = %d, want 2 (fixed)", cfg.StorageVersion)
			}
		})
	}
}

// TestViperLoader_Load_V2Format_KeyTakesPrecedenceOverDirectoryName tests key wins when key and directory_name mismatch
func TestViperLoader_Load_V2Format_KeyTakesPrecedenceOverDirectoryName(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write v2 config where key and directory_name field mismatch
	v2Config := `storage_version: 2

settings:
  hibernation_days: 14

projects:
  correct-key:
    path: "/home/user/my-project"
    directory_name: "wrong-value"
    favorite: false
`
	if err := os.WriteFile(configPath, []byte(v2Config), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Map key is "correct-key", so that's what we look up
	p, ok := cfg.Projects["correct-key"]
	if !ok {
		t.Fatal("correct-key not found in projects")
	}

	// DirectoryName should be set from map key, not from the field
	if p.DirectoryName != "correct-key" {
		t.Errorf("DirectoryName = %s, want 'correct-key' (key takes precedence)", p.DirectoryName)
	}
}

// Story 8.10: Test negative max_content_width validation
func TestViperLoader_Load_InvalidMaxContentWidth_Negative(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write config with negative max_content_width
	invalidConfig := `storage_version: 2

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10
  detail_layout: horizontal
  max_content_width: -50

projects: {}
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())

	// Should NOT return error - graceful degradation
	if err != nil {
		t.Errorf("Load() should not return error on invalid max_content_width, got %v", err)
	}

	// Should use default value (120) for invalid max_content_width
	if cfg.MaxContentWidth != 120 {
		t.Errorf("MaxContentWidth = %d, want 120 (default after fix)", cfg.MaxContentWidth)
	}
}

// Story 8.10: Test zero max_content_width is valid (unlimited mode)
func TestViperLoader_Load_MaxContentWidth_Zero_IsValid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write config with max_content_width: 0 (unlimited)
	validConfig := `storage_version: 2

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10
  detail_layout: horizontal
  max_content_width: 0

projects: {}
`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Zero should be preserved (unlimited mode)
	if cfg.MaxContentWidth != 0 {
		t.Errorf("MaxContentWidth = %d, want 0 (unlimited)", cfg.MaxContentWidth)
	}
}

// Story 8.11: Test stage_refresh_interval loading and saving
func TestViperLoader_Load_StageRefreshInterval(t *testing.T) {
	tests := []struct {
		name     string
		yamlVal  string
		expected int
	}{
		{
			name:     "default 30",
			yamlVal:  "30",
			expected: 30,
		},
		{
			name:     "disabled 0",
			yamlVal:  "0",
			expected: 0,
		},
		{
			name:     "custom 60",
			yamlVal:  "60",
			expected: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			validConfig := fmt.Sprintf(`storage_version: 2

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10
  detail_layout: horizontal
  stage_refresh_interval: %s

projects: {}
`, tt.yamlVal)

			if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			loader := NewViperLoader(configPath)
			cfg, err := loader.Load(context.Background())

			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg.StageRefreshIntervalSeconds != tt.expected {
				t.Errorf("StageRefreshIntervalSeconds = %d, want %d", cfg.StageRefreshIntervalSeconds, tt.expected)
			}
		})
	}
}

// Story 8.11: Test negative stage_refresh_interval is fixed to default
func TestViperLoader_Load_InvalidStageRefreshInterval_Negative(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidConfig := `storage_version: 2

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10
  detail_layout: horizontal
  stage_refresh_interval: -10

projects: {}
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	loader := NewViperLoader(configPath)
	cfg, err := loader.Load(context.Background())

	// Should NOT return error - graceful degradation
	if err != nil {
		t.Errorf("Load() should not return error on invalid stage_refresh_interval, got %v", err)
	}

	// Should use default value (30) for invalid stage_refresh_interval
	if cfg.StageRefreshIntervalSeconds != 30 {
		t.Errorf("StageRefreshIntervalSeconds = %d, want 30 (default after fix)", cfg.StageRefreshIntervalSeconds)
	}
}

// Story 8.11: Test stage_refresh_interval round-trip (save and reload)
func TestViperLoader_Save_StageRefreshInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval int
	}{
		{"default 30", 30},
		{"disabled 0", 0},
		{"custom 60", 60},
		{"custom 10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configDir := filepath.Join(tmpDir, ".vibe-dash")
			configPath := filepath.Join(configDir, "config.yaml")

			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatalf("failed to create config dir: %v", err)
			}

			loader := NewViperLoader(configPath)

			cfg := ports.NewConfig()
			cfg.StageRefreshIntervalSeconds = tt.interval

			err := loader.Save(context.Background(), cfg)
			if err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			// Reload and verify
			loader2 := NewViperLoader(configPath)
			cfg2, err := loader2.Load(context.Background())
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg2.StageRefreshIntervalSeconds != tt.interval {
				t.Errorf("StageRefreshIntervalSeconds = %d, want %d", cfg2.StageRefreshIntervalSeconds, tt.interval)
			}
		})
	}
}
