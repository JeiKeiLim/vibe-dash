# Story 3.5.3: Per-Project Config Files

Status: done

## Story

As a user,
I want project-specific settings in separate config files,
So that I can configure each project independently.

## Acceptance Criteria

1. **AC1: Config file creation** - Given project is added via `vibe add`, when project directory is created at `~/.vibe-dash/<project>/`, then `config.yaml` is created at `~/.vibe-dash/<project>/config.yaml` with default structure.

2. **AC2: Config file reading** - Given project config exists at `~/.vibe-dash/<project>/config.yaml`, when reading project settings, then values from project config are parsed and used.

3. **AC3: Project config precedence** - Given project config and master config both have `custom_hibernation_days`, when resolving effective setting, then project config value takes precedence over master config.

4. **AC4: Master config fallback** - Given only master config has a setting (project config omits it), when resolving effective setting, then master config value is used as fallback.

5. **AC5: Default fallback** - Given neither project config nor master config has a setting, when resolving effective setting, then built-in default from `ports.NewConfig()` is used.

6. **AC6: Detection metadata storage** - Given project has been scanned, when saving project state, then `detected_method` and `last_scanned` are written to project config.

7. **AC7: Invalid config handling** - Given project config has syntax errors or invalid values, when loading config, then warning is logged and defaults are used (graceful degradation per Architecture).

8. **AC8: Context cancellation** - Given config load/save operation in progress, when context is cancelled, then operation returns `ctx.Err()` promptly.

9. **AC9: Integration with DirectoryManager** - Given `DirectoryManager.EnsureProjectDir()` creates project directory, when `ProjectConfigLoader` is created for that directory, then config file is created if not exists and `.project-path` marker file presence is verified.

## Tasks / Subtasks

- [x] Task 1: Create `ProjectConfigLoader` interface in core ports (AC: 1, 2, 6, 8)
  - [x] Subtask 1.1: Create `internal/core/ports/project_config.go`
  - [x] Subtask 1.2: Define `ProjectConfigLoader` interface with methods:
    - `Load(ctx context.Context) (*ProjectConfigData, error)` - load project-specific config
    - `Save(ctx context.Context, data *ProjectConfigData) error` - persist project config
  - [x] Subtask 1.3: Define `ProjectConfigData` struct with fields:
    - `DetectedMethod string` - methodology detected (speckit, bmad, etc.)
    - `LastScanned time.Time` - last detection timestamp
    - `CustomHibernationDays *int` - optional override (nil = use global)
    - `AgentWaitingThresholdMinutes *int` - optional override (nil = use global)
    - `Notes string` - project notes/memo
  - [x] Subtask 1.4: Interface must use ONLY stdlib types and domain types (zero adapter imports)

- [x] Task 2: Implement `ViperProjectConfigLoader` (AC: 1, 2, 6, 7, 8, 9)
  - [x] Subtask 2.1: Create `internal/config/project_config_loader.go`
  - [x] Subtask 2.2: Create constructor `NewProjectConfigLoader(projectDir string) (*ViperProjectConfigLoader, error)`
  - [x] Subtask 2.3: Constructor validates:
    - `projectDir` exists via `os.Stat`
    - `.project-path` marker file exists (created by DirectoryManager)
    - Returns `domain.ErrPathNotAccessible` on validation failures
  - [x] Subtask 2.4: Implement `Load(ctx)`:
    - Check context cancellation first
    - If config.yaml doesn't exist, create with defaults
    - Use Viper to parse YAML
    - Map to `ProjectConfigData` struct
    - Handle parse errors gracefully (log warning, return defaults)
  - [x] Subtask 2.5: Implement `Save(ctx, data)`:
    - Check context cancellation first
    - Set Viper values from `ProjectConfigData`
    - Write config.yaml with comments

- [x] Task 3: Implement config cascade resolution (AC: 3, 4, 5)
  - [x] Subtask 3.1: Create `ConfigResolver` in `internal/config/resolver.go`
  - [x] Subtask 3.2: Define `ResolverConfig` struct holding master config + project config
  - [x] Subtask 3.3: Implement `GetEffectiveHibernationDays(projectID string) int`:
    - Check project config `CustomHibernationDays` first
    - Fall back to master config `HibernationDays`
    - Fall back to default (14)
  - [x] Subtask 3.4: Implement `GetEffectiveWaitingThreshold(projectID string) int`:
    - Check project config `AgentWaitingThresholdMinutes` first
    - Fall back to master config `AgentWaitingThresholdMinutes`
    - Fall back to default (10)

- [x] Task 4: Define project config YAML schema (AC: 1, 6)
  - [x] Subtask 4.1: Document schema in project_config_loader.go:
    ```yaml
    # ~/.vibe-dash/<project>/config.yaml
    detected_method: "speckit"
    last_scanned: "2025-12-18T10:30:00Z"
    custom_hibernation_days: 7          # Optional override
    agent_waiting_threshold_minutes: 5  # Optional override
    notes: "Main API service"           # Project notes
    ```
  - [x] Subtask 4.2: Use ISO 8601 UTC timestamps per Architecture

- [x] Task 5: Write unit tests (AC: 1-9)
  - [x] Subtask 5.1: Test `NewProjectConfigLoader` validates directory existence
  - [x] Subtask 5.2: Test `NewProjectConfigLoader` validates `.project-path` marker
  - [x] Subtask 5.3: Test `Load` creates default config if missing
  - [x] Subtask 5.4: Test `Load` parses existing config correctly
  - [x] Subtask 5.5: Test `Load` handles syntax errors gracefully
  - [x] Subtask 5.6: Test `Load` handles invalid values gracefully
  - [x] Subtask 5.7: Test `Save` writes config correctly
  - [x] Subtask 5.8: Test context cancellation for `Load` and `Save`
  - [x] Subtask 5.9: Test `ConfigResolver` cascade: project → master → default

- [x] Task 6: Write integration tests (AC: 1, 2, 9)
  - [x] Subtask 6.1: Create temp directories with `.project-path` marker
  - [x] Subtask 6.2: Test full lifecycle: create loader → save → load → verify
  - [x] Subtask 6.3: Test integration with real `DirectoryManager.EnsureProjectDir()`

## Dev Notes

### Files to Create (Canonical Reference)

| File | Purpose | Package |
|------|---------|---------|
| `internal/core/ports/project_config.go` | Interface + `ProjectConfigData` struct | `ports` |
| `internal/config/project_config_loader.go` | `ViperProjectConfigLoader` implementation | `config` |
| `internal/config/project_config_loader_test.go` | Unit tests (table-driven) | `config` |
| `internal/config/resolver.go` | Config cascade resolution | `config` |
| `internal/config/resolver_test.go` | Resolver tests | `config` |

**Note:** `internal/config/` is used for Viper-based loaders per existing codebase pattern (see `loader.go`). This is treated as adapter code despite not being in `internal/adapters/`.

### Architecture Alignment

**PRD Specification (lines 624-630):**
```yaml
# ~/.vibe/<project>/config.yaml  # PRD uses ~/.vibe/ but we use ~/.vibe-dash/
# Project-specific settings only (no path duplication)
detected_method: "speckit"
last_scanned: "2025-12-08T07:03:00Z"
custom_hibernation_days: null  # Override global setting if needed
```

**⚠️ Path Note:** PRD uses `~/.vibe/` but Architecture and implementation use `~/.vibe-dash/`. This story uses `~/.vibe-dash/` per Architecture specification (authoritative).

**Configuration Priority (PRD lines 632-636):**
```
1. CLI flags           --hibernation-days=7
2. Project config      ~/.vibe-dash/<project>/config.yaml
3. Master config       ~/.vibe-dash/config.yaml
4. Built-in defaults   hibernation_days: 14
```

**Architecture Specification (lines 309-329):**
> Config Files:
> - Master: `~/.vibe-dash/config.yaml`
> - Per-project: `~/.vibe-dash/<project>/config.yaml`
>
> Configuration Precedence (Explicit):
> 1. CLI flags (highest priority)
> 2. Project config
> 3. Master config
> 4. Built-in defaults (lowest priority)

### Hexagonal Architecture

```
internal/core/ports/project_config.go     → Interface definition (ZERO external imports)
internal/config/project_config_loader.go  → Viper implementation
internal/config/resolver.go               → Cascade resolution logic
```

**CRITICAL: Core never imports adapters.** The `ProjectConfigLoader` interface in `ports/` must NOT import Viper or any adapter code.

### Integration with Existing Config System

**Relationship with `ports.Config` and `ports.ProjectConfig`:**

The existing codebase has `ports.Config` with a `Projects map[string]ProjectConfig` (see `ports/config.go:29-54`). This story introduces a **separate** per-project YAML file system:

| Component | Location | Purpose |
|-----------|----------|---------|
| `ports.Config` | Master config | Global settings + project path index |
| `ports.ProjectConfig` | In master config | Path, display name, favorite status |
| **`ProjectConfigData`** (NEW) | Per-project YAML | Detection metadata, overrides for that project |

**How They Interact:**
1. `ports.Config` stores **which projects exist** and their paths
2. `ProjectConfigData` stores **per-project settings** in `~/.vibe-dash/<project>/config.yaml`
3. `ConfigResolver` combines both to resolve effective settings

**Existing `Config.GetEffectiveHibernationDays()` at `ports/config.go:77-82`:**
- Currently reads from `ProjectConfig.HibernationDays` in master config
- After this story: `ConfigResolver.GetEffectiveHibernationDays()` will read from per-project YAML first
- The existing method can remain for backward compatibility; `ConfigResolver` is the new primary API

### Interface Definition

```go
// ProjectConfigData holds project-specific configuration.
// Stored in ~/.vibe-dash/<project>/config.yaml
type ProjectConfigData struct {
    // DetectedMethod is the methodology detected (e.g., "speckit", "bmad")
    DetectedMethod string

    // LastScanned is when detection was last run (ISO 8601 UTC)
    LastScanned time.Time

    // CustomHibernationDays overrides global setting. nil = use global.
    CustomHibernationDays *int

    // AgentWaitingThresholdMinutes overrides global setting. nil = use global.
    AgentWaitingThresholdMinutes *int

    // Notes is user-defined project notes/memo
    Notes string
}

// ProjectConfigLoader handles per-project configuration files.
// Each project has its own config.yaml in its vibe-dash directory.
type ProjectConfigLoader interface {
    // Load reads project-specific configuration from YAML file.
    // Creates default config file if it doesn't exist.
    // Returns defaults on syntax/parse errors (graceful degradation).
    Load(ctx context.Context) (*ProjectConfigData, error)

    // Save persists project configuration to YAML file.
    // Creates file if it doesn't exist.
    Save(ctx context.Context, data *ProjectConfigData) error
}

// NewProjectConfigData creates a ProjectConfigData with zero values.
// Used for graceful degradation when config cannot be loaded.
func NewProjectConfigData() *ProjectConfigData {
    return &ProjectConfigData{}
}
```

### Constructor Pattern

```go
// NewProjectConfigLoader creates a loader for project-specific config.
// projectDir must exist and contain .project-path marker (created by DirectoryManager).
func NewProjectConfigLoader(projectDir string) (*ViperProjectConfigLoader, error) {
    // Validate projectDir exists
    if _, err := os.Stat(projectDir); os.IsNotExist(err) {
        return nil, fmt.Errorf("%w: project directory does not exist: %s",
            domain.ErrPathNotAccessible, projectDir)
    }

    // Verify DirectoryManager created this directory
    markerPath := filepath.Join(projectDir, ".project-path")
    if _, err := os.Stat(markerPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("%w: directory not created by DirectoryManager (missing .project-path): %s",
            domain.ErrPathNotAccessible, projectDir)
    }

    configPath := filepath.Join(projectDir, "config.yaml")
    v := viper.New()
    v.SetConfigFile(configPath)
    v.SetConfigType("yaml")

    return &ViperProjectConfigLoader{
        projectDir: projectDir,
        configPath: configPath,
        v:          v,
    }, nil
}
```

### Default Config File Content

```go
func (l *ViperProjectConfigLoader) writeDefaultConfig() error {
    content := `# Project-specific vibe-dash configuration
# Auto-generated - modify as needed

# Methodology detected by vibe-dash
detected_method: ""
last_scanned: ""

# Optional: Override global hibernation threshold
# custom_hibernation_days: 7

# Optional: Override global agent waiting threshold
# agent_waiting_threshold_minutes: 5

# Project notes/memo
notes: ""
`
    return os.WriteFile(l.configPath, []byte(content), 0644)
}
```

### Config Cascade Resolution

```go
// ConfigResolver resolves effective settings using the cascade:
// project config → master config → defaults
type ConfigResolver struct {
    masterConfig  *ports.Config
    projectConfig *ports.ProjectConfigData
}

func (r *ConfigResolver) GetEffectiveHibernationDays() int {
    // 1. Project config override
    if r.projectConfig != nil && r.projectConfig.CustomHibernationDays != nil {
        return *r.projectConfig.CustomHibernationDays
    }
    // 2. Master config
    if r.masterConfig != nil {
        return r.masterConfig.HibernationDays
    }
    // 3. Default
    return 14
}
```

### Error Handling

| Error | When | Recovery |
|-------|------|----------|
| `domain.ErrPathNotAccessible` | Dir missing, marker missing, permission denied | Return error, user must re-add project |
| `domain.ErrConfigInvalid` | Invalid values in config | Log warning, use defaults (graceful degradation) |
| `ctx.Err()` | Context cancelled | Return immediately |

```go
// Domain errors to use:
domain.ErrPathNotAccessible   // Directory/file access issues
domain.ErrConfigInvalid       // Invalid config values (per ports/config.go)

// Context cancellation pattern (REQUIRED at start of Load/Save):
select {
case <-ctx.Done():
    return nil, ctx.Err()
default:
}

// Graceful degradation pattern (per Architecture):
if err := l.v.ReadInConfig(); err != nil {
    slog.Warn("project config syntax error, using defaults",
        "error", err, "path", l.configPath)
    return NewProjectConfigData(), nil  // Return defaults, not error
}
```

### Existing Code to Leverage

| File | Pattern to Reuse |
|------|------------------|
| `config/loader.go:41-80` | `Load()` pattern with graceful degradation |
| `config/loader.go:83-126` | `Save()` pattern with Viper |
| `config/loader.go:135-151` | `writeDefaultConfig()` with comments |
| `ports/config.go:56-66` | `NewConfig()` defaults pattern |
| `ports/config.go:96-123` | `Validate()` pattern |

### Difference from Master ConfigLoader

| Aspect | `ViperLoader` (Master) | `ViperProjectConfigLoader` (Project) |
|--------|------------------------|--------------------------------------|
| Config path | `~/.vibe-dash/config.yaml` | `~/.vibe-dash/<project>/config.yaml` |
| Scope | Global settings + project path index | Project-specific settings only |
| Fields | `HibernationDays`, `Projects` map | `DetectedMethod`, `CustomHibernationDays`, etc. |
| Validation | Validates directory marker | Same - validates `.project-path` marker |
| Used by | Application startup | Per-project operations |

### Dependencies

**Depends on (COMPLETED):**
- Story 3.5.1: DirectoryManager - Provides `.project-path` marker file

**Required by:**
- Story 3.5.4: Master Config as Path Index - Master config references project directories
- Story 3.5.5: Repository Coordinator - Uses project config for per-project settings
- Story 3.5.6: Update CLI Commands - `vibe add` creates project config

### Previous Story Learnings (3.5.2)

From Story 3.5.2 (Per-Project SQLite Repository) - commit `66a847f`:
- Constructor validates `.project-path` marker file presence
- Use `context.Context` for cancellation support in all methods
- Return `domain.ErrPathNotAccessible` on validation failures
- Table-driven tests are required per Architecture
- Graceful degradation: log warnings, return defaults on errors
- Helper functions duplicated with suffix to avoid name collisions (e.g., `wrapDBErrorForProject`)

**Key Pattern from 3.5.2 Constructor:**
```go
// Validate projectDir exists
if _, err := os.Stat(projectDir); os.IsNotExist(err) {
    return nil, fmt.Errorf("%w: project directory does not exist: %s",
        domain.ErrPathNotAccessible, projectDir)
}

// Verify DirectoryManager created this directory
markerPath := filepath.Join(projectDir, ".project-path")
if _, err := os.Stat(markerPath); os.IsNotExist(err) {
    return nil, fmt.Errorf("%w: directory not created by DirectoryManager (missing .project-path): %s",
        domain.ErrPathNotAccessible, projectDir)
}
```

### Testing Strategy

**Unit Tests (table-driven):**

```go
func TestNewProjectConfigLoader(t *testing.T) {
    tests := []struct {
        name        string
        setup       func(t *testing.T) string // returns projectDir
        expectError error
    }{
        {
            name: "valid directory with marker",
            setup: func(t *testing.T) string {
                dir := t.TempDir()
                os.WriteFile(filepath.Join(dir, ".project-path"), []byte("/path"), 0644)
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
            _, err := NewProjectConfigLoader(dir)
            if tt.expectError != nil {
                require.ErrorIs(t, err, tt.expectError)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Manual Testing

After implementation:

1. **Verify config creation on add:**
   ```bash
   mkdir -p ~/.vibe-dash/test-project
   echo "/path/to/test-project" > ~/.vibe-dash/test-project/.project-path
   # ProjectConfigLoader should create config.yaml on first Load()
   ```

2. **Verify config reading:**
   ```bash
   cat ~/.vibe-dash/test-project/config.yaml
   # Should show default structure
   ```

3. **Verify cascade resolution:**
   ```bash
   # Set project override
   echo "custom_hibernation_days: 7" >> ~/.vibe-dash/test-project/config.yaml
   # Master config has 14, project should use 7
   ```

4. **Verify graceful degradation:**
   ```bash
   echo "invalid: yaml: content" > ~/.vibe-dash/test-project/config.yaml
   # Should log warning and use defaults
   ```

### Project Structure Notes

Alignment with PRD storage structure:
```
~/.vibe-dash/
  ├── config.yaml                 # Master index (Story 3.5.4)
  ├── api-service/
  │   ├── .project-path           # Marker file (Story 3.5.1)
  │   ├── config.yaml             # THIS STORY - project-specific settings
  │   └── state.db                # Per-project SQLite (Story 3.5.2)
```

## Dev Agent Record

### Context Reference

- Epic 3.5: Storage Structure Alignment
- Story Dependencies: Depends on 3.5.1, required by 3.5.4, 3.5.5, 3.5.6
- PRD Reference: Lines 624-636 (Per-project config)
- Architecture Reference: Lines 309-329 (Config cascade)
- Project Context: Lines 58-67 (Configuration Cascade)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None - implementation completed without issues.

### Completion Notes List

- **AC1-AC2:** `ProjectConfigLoader` interface and `ViperProjectConfigLoader` implementation handle config file creation and reading
- **AC3-AC5:** `ConfigResolver` implements cascade: project config → master config → defaults (14 for hibernation, 10 for waiting)
- **AC6:** `DetectedMethod` and `LastScanned` fields in `ProjectConfigData` with ISO 8601 UTC timestamps
- **AC7:** Graceful degradation via `slog.Warn` on syntax errors, returning defaults instead of errors
- **AC8:** Context cancellation check at start of `Load()` and `Save()` methods
- **AC9:** Constructor validates `.project-path` marker file presence before allowing config operations
- All tests pass: 47 unit tests + 4 integration tests
- Follows existing patterns from `config/loader.go` and `sqlite/project_repository.go`
- Zero adapter imports in core ports (verified via `go list`)

### File List

| File | Status | Purpose |
|------|--------|---------|
| `internal/core/ports/project_config.go` | Created | `ProjectConfigLoader` interface + `ProjectConfigData` struct + `Validate()` |
| `internal/core/ports/project_config_test.go` | Created | Interface usage tests + validation tests |
| `internal/config/project_config_loader.go` | Created | `ViperProjectConfigLoader` implementation with validation |
| `internal/config/project_config_loader_test.go` | Created | Unit tests (table-driven) including negative value handling |
| `internal/config/resolver.go` | Created | `ConfigResolver` cascade resolution |
| `internal/config/resolver_test.go` | Created | Cascade resolution tests |
| `internal/config/project_config_integration_test.go` | Created | Integration tests with DirectoryManager |
| `go.mod` | Modified | Added testify as direct dependency |

### Change Log

- 2025-12-18: Story 3.5.3 implemented - Per-project config files with cascade resolution
- 2025-12-18: [Code Review] Added `ProjectConfigData.Validate()` method for AC7 compliance
- 2025-12-18: [Code Review] Added `fixInvalidValues()` for graceful degradation on negative values
- 2025-12-18: [Code Review] Added `ProjectDir()` and `ConfigPath()` getter methods
- 2025-12-18: [Code Review] Added test for negative value handling (`TestProjectConfigLoader_Load_HandlesNegativeValues`)
- 2025-12-18: [Code Review] Added validation tests (`TestProjectConfigData_Validate`)
