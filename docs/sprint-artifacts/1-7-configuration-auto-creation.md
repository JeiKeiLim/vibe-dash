# Story 1.7: Configuration Auto-Creation

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | Called from CLI root command on startup |
| **Key Dependencies** | github.com/spf13/viper (NEW - needs go get) |
| **Files to Create** | config/config.go, config/loader.go, config/defaults.go, config_test.go |
| **Location** | internal/config/ |
| **Config Path** | ~/.vibe-dash/config.yaml |

### Configuration Default Values

| Setting | Default | YAML Key |
|---------|---------|----------|
| `HibernationDays` | 14 | `settings.hibernation_days` |
| `RefreshIntervalSeconds` | 10 | `settings.refresh_interval_seconds` |
| `RefreshDebounceMs` | 200 | `settings.refresh_debounce_ms` |
| `AgentWaitingThresholdMinutes` | 10 | `settings.agent_waiting_threshold_minutes` |
| `Projects` | {} (empty map) | `projects` |

## Story

**As a** user,
**I want** configuration to auto-create on first run,
**So that** I don't need manual setup.

## Acceptance Criteria

```gherkin
AC1: Given I run vibe for the first time
     When ~/.vibe-dash/ directory doesn't exist
     Then it is created automatically
     And ~/.vibe-dash/config.yaml is created with defaults:
       """yaml
       settings:
         hibernation_days: 14
         refresh_interval_seconds: 10
         refresh_debounce_ms: 200
         agent_waiting_threshold_minutes: 10
       projects: {}
       """

AC2: Given I run vibe again
     When ~/.vibe-dash/config.yaml already exists
     Then existing config is preserved
     And no duplicate creation occurs

AC3: Given config.yaml has syntax errors
     When vibe starts
     Then error is reported with line number if possible
     And application continues with defaults
     And exit code is 0 (degraded operation, not failure)

AC4: Given config.yaml has invalid values
     When vibe starts
     Then specific invalid value is reported
     And application continues with default for that value
     And exit code is 0

AC5: Given ~/.vibe-dash/ directory exists but is not writable
     When vibe tries to create config.yaml
     Then appropriate error message is shown
     And application continues with defaults
```

## Tasks / Subtasks

- [x] **Task 1: Add Viper dependency** (AC: all)
  - [x] 1.1 Run `go get github.com/spf13/viper && go mod tidy` and verify go.mod

- [x] **Task 2: Create config/defaults.go** (AC: 1)
  - [x] 2.1 Create internal/config/defaults.go
  - [x] 2.2 Define DefaultConfigDir constant (resolves to ~/.vibe-dash)
  - [x] 2.3 Define DefaultConfigFileName constant (config.yaml)
  - [x] 2.4 Create GetDefaultConfigPath() function that handles cross-platform home directory
  - [x] 2.5 Create GetConfigDir() function

- [x] **Task 3: Create config/loader.go implementing ConfigLoader interface** (AC: 1, 2, 3, 4, 5)
  - [x] 3.1 Create internal/config/loader.go
  - [x] 3.2 Define ViperLoader struct implementing ports.ConfigLoader
  - [x] 3.3 Implement NewViperLoader constructor
  - [x] 3.4 Implement Load(ctx) method:
    - [x] 3.4.1 Check if config dir exists, create if not
    - [x] 3.4.2 Check if config file exists
    - [x] 3.4.3 If file doesn't exist, create with defaults
    - [x] 3.4.4 If file exists, read and parse with Viper
    - [x] 3.4.5 Handle syntax errors gracefully (log warning, return defaults)
    - [x] 3.4.6 Validate loaded values, use defaults for invalid ones
  - [x] 3.5 Implement Save(ctx, config) method
  - [x] 3.6 Add helper function for ensureConfigDir()
  - [x] 3.7 Add helper function for writeDefaultConfig()

- [x] **Task 4: Create config/config.go for Viper-specific config mapping** (AC: 1)
  - [x] 4.1 Create YAML struct tags for Config to match expected file format
  - [x] 4.2 Create function to map Viper config to ports.Config struct
  - [x] 4.3 Create function to map ports.Config to YAML-writable format

- [x] **Task 5: Integrate config loading into CLI startup** (AC: all)
  - [x] 5.1 Modify cmd/vibe/main.go to call config loader before CLI execution
  - [x] 5.2 Pass config to TUI (or store globally for initial MVP)
  - [x] 5.3 Handle config load errors gracefully (log, continue with defaults)

- [x] **Task 6: Write Tests** (AC: all)
  - [x] 6.1 Create internal/config/loader_test.go
  - [x] 6.2 Test: Config directory creation on first run
  - [x] 6.3 Test: Config file creation with correct defaults
  - [x] 6.4 Test: Existing config file is not overwritten
  - [x] 6.5 Test: Syntax error handling (invalid YAML)
  - [x] 6.6 Test: Invalid value handling (negative hibernation_days)
  - [x] 6.7 Test: Cross-platform home directory resolution
  - [x] 6.8 Use temporary directories for test isolation

- [x] **Task 7: Integration and Validation** (AC: all)
  - [x] 7.1 Run `make build` and verify compilation
  - [x] 7.2 Run `make lint` and fix any issues
  - [x] 7.3 Run `make test` and verify all tests pass
  - [x] 7.4 Manual test: Delete ~/.vibe-dash/, run vibe, verify config created
  - [x] 7.5 Manual test: Edit config with syntax error, run vibe, verify warning and defaults used
  - [x] 7.6 Manual test: Set invalid value (hibernation_days: -1), verify warning

## Implementation Order (Recommended)

Execute tasks in this order to minimize rework:

1. **Task 1: Add Viper** - Get dependency in place
2. **Task 2: defaults.go** - Path resolution and constants
3. **Task 3: loader.go** - Core config loading logic
4. **Task 4: config.go** - YAML mapping helpers
5. **Task 6: Tests** - Test the loader before integration
6. **Task 5: Integration** - Wire into CLI
7. **Task 7: Validation** - Final checks

## Dev Notes

### CRITICAL Requirements (Must Not Miss)

| Requirement | Why | Reference |
|-------------|-----|-----------|
| **Auto-create directory** | FR44: Auto-create default config on first project add | PRD FR44 |
| **Graceful degradation** | Exit code 0 on config errors, continue with defaults | AC3 |
| **Centralized storage** | ~/.vibe-dash/ NOT per-project | Architecture lines 164-166 |
| **Viper for parsing** | Architecture decision for config cascade | Architecture line 309-330 |
| **Implement ConfigLoader interface** | Port already defined in ports/config.go | Story 1.3 |

### Existing ConfigLoader Interface (from ports/config.go)

**CRITICAL:** The `ConfigLoader` interface is already defined in `internal/core/ports/config.go` (lines 131-142). Your `ViperLoader` MUST implement this interface. Do NOT recreate it - import and implement it.

The interface requires:
- `Load(ctx context.Context) (*Config, error)` - Read config, return defaults on error
- `Save(ctx context.Context, config *Config) error` - Persist config to YAML

### Config Struct (Already Defined)

The `ports.Config` struct is defined in `internal/core/ports/config.go` with these fields:

```go
type Config struct {
    HibernationDays              int                       // Default: 14
    RefreshIntervalSeconds       int                       // Default: 10
    RefreshDebounceMs            int                       // Default: 200
    AgentWaitingThresholdMinutes int                       // Default: 10
    Projects                     map[string]ProjectConfig  // Per-project settings
}
```

Use `ports.NewConfig()` to create a Config with defaults - DO NOT construct Config{} directly.

### Viper Integration Pattern

```go
package config

import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "path/filepath"

    "github.com/spf13/viper"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// ViperLoader implements ports.ConfigLoader using Viper for YAML parsing.
type ViperLoader struct {
    configPath string
    v          *viper.Viper
}

// NewViperLoader creates a ConfigLoader that reads from the specified config path.
// If configPath is empty, uses the default ~/.vibe-dash/config.yaml
func NewViperLoader(configPath string) *ViperLoader {
    if configPath == "" {
        configPath = GetDefaultConfigPath()
    }

    v := viper.New()
    v.SetConfigFile(configPath)
    v.SetConfigType("yaml")

    return &ViperLoader{
        configPath: configPath,
        v:          v,
    }
}

func (l *ViperLoader) Load(ctx context.Context) (*ports.Config, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Ensure config directory exists
    if err := l.ensureConfigDir(); err != nil {
        slog.Warn("could not create config directory, using defaults", "error", err)
        return ports.NewConfig(), nil
    }

    // Check if config file exists
    if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
        // Create default config file
        if err := l.writeDefaultConfig(); err != nil {
            slog.Warn("could not write default config, using defaults", "error", err)
            return ports.NewConfig(), nil
        }
    }

    // Read config file
    if err := l.v.ReadInConfig(); err != nil {
        slog.Warn("config syntax error, using defaults", "error", err, "path", l.configPath)
        return ports.NewConfig(), nil
    }

    // Map Viper values to Config struct
    cfg := l.mapViperToConfig()

    // Validate and fix invalid values
    if err := cfg.Validate(); err != nil {
        slog.Warn("config validation failed, using defaults for invalid values", "error", err)
        cfg = l.fixInvalidValues(cfg)
    }

    return cfg, nil
}
```

### Cross-Platform Home Directory Resolution

```go
// GetDefaultConfigPath returns the default config file path.
// Uses os.UserHomeDir() for cross-platform compatibility.
func GetDefaultConfigPath() string {
    home, err := os.UserHomeDir()
    if err != nil {
        // Fallback to current directory if home can't be determined
        slog.Warn("could not determine home directory", "error", err)
        return filepath.Join(".", ".vibe-dash", "config.yaml")
    }
    return filepath.Join(home, ".vibe-dash", "config.yaml")
}

// GetConfigDir returns the config directory path.
func GetConfigDir() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return filepath.Join(".", ".vibe-dash")
    }
    return filepath.Join(home, ".vibe-dash")
}
```

### writeDefaultConfig Implementation (CRITICAL)

```go
// writeDefaultConfig creates the default config file with comments.
func (l *ViperLoader) writeDefaultConfig() error {
    cfg := ports.NewConfig()

    // Set Viper values for YAML generation
    l.v.Set("settings.hibernation_days", cfg.HibernationDays)
    l.v.Set("settings.refresh_interval_seconds", cfg.RefreshIntervalSeconds)
    l.v.Set("settings.refresh_debounce_ms", cfg.RefreshDebounceMs)
    l.v.Set("settings.agent_waiting_threshold_minutes", cfg.AgentWaitingThresholdMinutes)
    l.v.Set("projects", map[string]interface{}{})

    return l.v.WriteConfig()
}
```

**Note:** Viper's WriteConfig() produces valid YAML but without comments. If comments are desired, write manually:

```go
func (l *ViperLoader) writeDefaultConfigWithComments() error {
    cfg := ports.NewConfig()

    content := fmt.Sprintf(`# Vibe Dashboard Configuration
# Auto-generated on first run

settings:
  hibernation_days: %d
  refresh_interval_seconds: %d
  refresh_debounce_ms: %d
  agent_waiting_threshold_minutes: %d

projects: {}
`, cfg.HibernationDays, cfg.RefreshIntervalSeconds, cfg.RefreshDebounceMs, cfg.AgentWaitingThresholdMinutes)

    return os.WriteFile(l.configPath, []byte(content), 0644)
}
```

### mapViperToConfig Implementation (CRITICAL)

```go
// mapViperToConfig converts Viper config values to ports.Config struct.
func (l *ViperLoader) mapViperToConfig() *ports.Config {
    cfg := ports.NewConfig()

    // Only override defaults if values are explicitly set in config file
    if l.v.IsSet("settings.hibernation_days") {
        cfg.HibernationDays = l.v.GetInt("settings.hibernation_days")
    }
    if l.v.IsSet("settings.refresh_interval_seconds") {
        cfg.RefreshIntervalSeconds = l.v.GetInt("settings.refresh_interval_seconds")
    }
    if l.v.IsSet("settings.refresh_debounce_ms") {
        cfg.RefreshDebounceMs = l.v.GetInt("settings.refresh_debounce_ms")
    }
    if l.v.IsSet("settings.agent_waiting_threshold_minutes") {
        cfg.AgentWaitingThresholdMinutes = l.v.GetInt("settings.agent_waiting_threshold_minutes")
    }

    // Map projects if present
    projectsMap := l.v.GetStringMap("projects")
    for id, v := range projectsMap {
        projectData, ok := v.(map[string]interface{})
        if !ok {
            continue
        }

        pc := ports.ProjectConfig{}
        if path, ok := projectData["path"].(string); ok {
            pc.Path = path
        }
        if name, ok := projectData["display_name"].(string); ok {
            pc.DisplayName = name
        }
        if fav, ok := projectData["favorite"].(bool); ok {
            pc.IsFavorite = fav
        }
        // Handle optional overrides (pointer fields)
        if hd, ok := projectData["hibernation_days"].(int); ok {
            pc.HibernationDays = &hd
        }
        if awt, ok := projectData["agent_waiting_threshold_minutes"].(int); ok {
            pc.AgentWaitingThresholdMinutes = &awt
        }

        cfg.Projects[id] = pc
    }

    return cfg
}
```

### fixInvalidValues Implementation (CRITICAL)

```go
// fixInvalidValues corrects invalid config values by replacing with defaults.
// Logs a warning for each corrected value.
func (l *ViperLoader) fixInvalidValues(cfg *ports.Config) *ports.Config {
    defaults := ports.NewConfig()

    if cfg.HibernationDays < 0 {
        slog.Warn("invalid hibernation_days, using default",
            "invalid_value", cfg.HibernationDays,
            "default_value", defaults.HibernationDays)
        cfg.HibernationDays = defaults.HibernationDays
    }

    if cfg.RefreshIntervalSeconds <= 0 {
        slog.Warn("invalid refresh_interval_seconds, using default",
            "invalid_value", cfg.RefreshIntervalSeconds,
            "default_value", defaults.RefreshIntervalSeconds)
        cfg.RefreshIntervalSeconds = defaults.RefreshIntervalSeconds
    }

    if cfg.RefreshDebounceMs <= 0 {
        slog.Warn("invalid refresh_debounce_ms, using default",
            "invalid_value", cfg.RefreshDebounceMs,
            "default_value", defaults.RefreshDebounceMs)
        cfg.RefreshDebounceMs = defaults.RefreshDebounceMs
    }

    if cfg.AgentWaitingThresholdMinutes < 0 {
        slog.Warn("invalid agent_waiting_threshold_minutes, using default",
            "invalid_value", cfg.AgentWaitingThresholdMinutes,
            "default_value", defaults.AgentWaitingThresholdMinutes)
        cfg.AgentWaitingThresholdMinutes = defaults.AgentWaitingThresholdMinutes
    }

    // Fix per-project overrides
    for id, pc := range cfg.Projects {
        if pc.HibernationDays != nil && *pc.HibernationDays < 0 {
            slog.Warn("invalid project hibernation_days, removing override",
                "project", id, "invalid_value", *pc.HibernationDays)
            pc.HibernationDays = nil
            cfg.Projects[id] = pc
        }
        if pc.AgentWaitingThresholdMinutes != nil && *pc.AgentWaitingThresholdMinutes < 0 {
            slog.Warn("invalid project agent_waiting_threshold_minutes, removing override",
                "project", id, "invalid_value", *pc.AgentWaitingThresholdMinutes)
            pc.AgentWaitingThresholdMinutes = nil
            cfg.Projects[id] = pc
        }
    }

    return cfg
}
```

### YAML File Format

The config file should match this format:

```yaml
# Vibe Dashboard Configuration
# Auto-generated on first run

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects: {}
```

### YAML Struct Tags Pattern

Create a separate struct for YAML serialization that maps to the ports.Config:

```go
// yamlConfig represents the YAML file structure.
// Used for serialization/deserialization with Viper.
type yamlConfig struct {
    Settings yamlSettings           `yaml:"settings"`
    Projects map[string]yamlProject `yaml:"projects"`
}

type yamlSettings struct {
    HibernationDays              int `yaml:"hibernation_days"`
    RefreshIntervalSeconds       int `yaml:"refresh_interval_seconds"`
    RefreshDebounceMs            int `yaml:"refresh_debounce_ms"`
    AgentWaitingThresholdMinutes int `yaml:"agent_waiting_threshold_minutes"`
}

type yamlProject struct {
    Path                         string `yaml:"path"`
    DisplayName                  string `yaml:"display_name,omitempty"`
    IsFavorite                   bool   `yaml:"favorite,omitempty"`
    HibernationDays              *int   `yaml:"hibernation_days,omitempty"`
    AgentWaitingThresholdMinutes *int   `yaml:"agent_waiting_threshold_minutes,omitempty"`
}
```

### Error Handling Strategy

**Graceful degradation** is the key principle:

1. **Directory creation fails** → Log warning, return defaults, continue
2. **Config file doesn't exist** → Create with defaults, continue
3. **Config file has syntax error** → Log warning with details, return defaults, continue
4. **Config has invalid values** → Log warning, use default for that field, continue

**NEVER** exit with non-zero code for config issues. The application should always start.

### Testing Pattern with Temp Directories

```go
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
}

func TestViperLoader_Load_SyntaxError(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.yaml")

    // Write invalid YAML
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
}

// Test AC2: Existing config is preserved (idempotency)
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

// Test AC4: Invalid values trigger warning and use defaults
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

// Test Save() method persists config correctly
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
```

### Concurrency Note (IMPORTANT)

Config loading and saving is **NOT thread-safe** in this implementation. This is acceptable for MVP because:
- vibe is a single-process CLI application
- Config is loaded once at startup
- Save operations are user-initiated (rare)

If multi-process access is needed post-MVP, consider:
- File locking via `syscall.Flock()` (Unix) or `LockFileEx()` (Windows)
- Or accept last-write-wins semantics with user documentation

### Config Usage Roadmap

After this story, config is loaded but not actively consumed. Future stories will wire config values:

| Story | Config Field Used | Purpose |
|-------|-------------------|---------|
| 4.1 | RefreshDebounceMs | File watcher debounce delay |
| 4.4 | AgentWaitingThresholdMinutes | WAITING indicator threshold |
| 5.6 | HibernationDays | Auto-hibernation threshold |

For now, config loading validates the system works end-to-end. The loaded config should be passed through to services when they're implemented.

### Integration with CLI (main.go)

```go
// cmd/vibe/main.go - minimal integration
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
    "github.com/JeiKeiLim/vibe-dash/internal/config"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown signals
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigCh
        cancel()
    }()

    // Load config (always succeeds, may log warnings)
    loader := config.NewViperLoader("")
    cfg, _ := loader.Load(ctx) // Intentionally ignore error - graceful degradation

    // Store config for later use (MVP: package-level variable)
    // Future stories will wire this to services via dependency injection
    _ = cfg // Consumed by: Story 4.1, 4.4, 5.6

    if err := cli.Execute(ctx); err != nil {
        os.Exit(1)
    }
}
```

### Files to Create

| File | Purpose |
|------|---------|
| `internal/config/defaults.go` | Path constants and helper functions |
| `internal/config/loader.go` | ViperLoader implementation of ConfigLoader |
| `internal/config/config.go` | YAML struct definitions for serialization |
| `internal/config/loader_test.go` | Tests for config loading |

### Dependencies to Add

```bash
go get github.com/spf13/viper
go mod tidy
```

Viper version will be automatically resolved to latest stable.

### DO NOT (Anti-Patterns)

| DO NOT | DO INSTEAD |
|--------|------------|
| Exit with non-zero code on config error | Log warning, continue with defaults |
| Create per-project config directories | Use centralized ~/.vibe-dash/ |
| Construct Config{} directly | Use ports.NewConfig() |
| Hardcode home directory path | Use os.UserHomeDir() |
| Panic on config errors | Handle gracefully with logging |
| Test with real ~/.vibe-dash/ | Use t.TempDir() for isolation |

### Project Structure Notes

**Alignment with Architecture:**

- Config loading in `internal/config/` (per Architecture Section: Project Structure)
- Implements `ports.ConfigLoader` interface from Story 1.3
- Tests co-located as `loader_test.go`
- Uses Viper per Architecture "Configuration System" decision

**File Location Verification:**

The following directories/files already exist:
- `internal/config/` directory (with .keep file)
- `internal/core/ports/config.go` (ConfigLoader interface and Config struct)

### Previous Story Learnings (Story 1.6)

From the completed Story 1.6:

1. **Code review fixes applied** - Watch for similar issues (test coverage, edge cases)
2. **Test assertions** - Include content verification in tests
3. **Documentation comments** - Add clear doc comments for exported functions
4. **Edge case tests** - Test empty inputs, invalid values, boundary conditions

### References

| Document | Section | Key Content |
|----------|---------|-------------|
| architecture.md | Configuration System | Lines 309-330: Viper integration, cascade |
| architecture.md | Project Structure | Lines 782-790: internal/config/ |
| prd.md | Configuration Schema | Lines 596-638: Config structure |
| prd.md | FR39-47 | Configuration management requirements |
| epics.md | Story 1.7 | Lines 473-513: Full acceptance criteria |
| ports/config.go | ConfigLoader | Interface and Config struct |
| project-context.md | Technology Stack | Viper Latest |

### Previous Story Files Available

| Story | Status | Key Learnings |
|-------|--------|---------------|
| 1.1 | Done | Project scaffolding complete |
| 1.2 | Done | Domain entities in internal/core/domain/ |
| 1.3 | Done | Port interfaces in internal/core/ports/ (ConfigLoader defined) |
| 1.4 | Done | CLI framework, exit codes, flags, context |
| 1.5 | Done | TUI shell, EmptyView, help overlay |
| 1.6 | Done | Lipgloss styles, code review patterns |

## Dev Agent Record

### Context Reference

Story context created from comprehensive analysis of:
- docs/epics.md (Story 1.7 requirements)
- docs/architecture.md (Configuration System, Viper integration)
- docs/prd.md (FR39-47, Configuration Schema)
- docs/project-context.md (Technology stack)
- internal/core/ports/config.go (ConfigLoader interface already defined)
- docs/sprint-artifacts/1-6-lipgloss-styles-foundation.md (Previous story learnings)
- Git history (recent commits 0634e77, a086e6a)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None - story created in YOLO mode per SM activation step 4.

### Completion Notes List

**Implementation Summary:**
- Implemented ViperLoader struct implementing ports.ConfigLoader interface
- Config auto-creates ~/.vibe-dash/config.yaml with commented defaults on first run
- Graceful degradation: syntax errors and invalid values log warnings, app continues with defaults
- All ACs validated via 22 unit tests covering all edge cases
- Manual tests confirmed: AC1 (auto-creation), AC3 (syntax errors), AC4 (invalid values)

**Technical Decisions:**
- Used Viper's map-based approach for YAML parsing instead of separate struct tags (simpler, less code)
- Config.go repurposed as package documentation showing YAML format
- Config loaded at startup in main.go, logged at debug level for MVP
- Future stories will wire config values to services via dependency injection

**Test Coverage:**
- 22 tests covering all acceptance criteria
- Tests use t.TempDir() for isolation
- Tests verify file creation, content, preservation, syntax errors, invalid values, projects, context cancellation

### File List

**New Files:**
- internal/config/defaults.go - Path constants and helper functions
- internal/config/loader.go - ViperLoader implementing ConfigLoader interface
- internal/config/config.go - Package documentation with YAML format
- internal/config/defaults_test.go - Tests for defaults functions
- internal/config/loader_test.go - Tests for ViperLoader (24 tests including AC5 unwritable dir)
- docs/sprint-artifacts/validation-report-1-7-configuration-auto-creation.md - Manual test validation

**Modified Files:**
- cmd/vibe/main.go - Added config loading on startup
- go.mod - Added github.com/spf13/viper as direct dependency
- go.sum - Updated with Viper and transitive dependencies

### Code Review Fixes Applied

| Issue | Severity | Fix Applied |
|-------|----------|-------------|
| Viper marked as indirect in go.mod | HIGH | Ran `go mod tidy` to move to direct deps |
| AC5 unwritable directory not tested | HIGH | Added `TestViperLoader_Load_UnwritableDirectory` |
| Custom contains() helper | MEDIUM | Replaced with `strings.Contains` |
| AC3 line number test missing | MEDIUM | Added `TestViperLoader_Load_SyntaxErrorWithLineNumber` |
| File List incomplete | MEDIUM | Added validation report to File List |
