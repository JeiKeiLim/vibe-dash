# Story 3.5.4: Master Config as Path Index

Status: done

## Story

As a user,
I want the master config to serve as a project path index with global settings,
So that project mappings and defaults are centrally managed.

## Acceptance Criteria

1. **AC1: Storage version field** - Given new master config format, when reading config, then `storage_version: 2` is expected and validated.

2. **AC2: Project path mapping** - Given projects are tracked, when viewing master config, then each project has entry with `path` and `directory_name` mapping (directory_name is the subdirectory in `~/.vibe-dash/`).

3. **AC3: Project entry on add** - Given project is added via `vibe add`, when master config is updated, then project entry includes `path`, `directory_name`, `favorite` status, and optionally `display_name`.

4. **AC4: Global settings section** - Given global settings in master config `settings` section, when no project-specific override exists in per-project config, then global setting is used.

5. **AC5: Backward compatibility** - Given old format config without `storage_version`, when loading config, then migration is attempted gracefully with warning logged.

6. **AC6: Project removal cleanup** - Given project is removed via `vibe remove`, when master config is updated, then project entry is removed from `projects` map.

7. **AC7: Directory name lookup** - Given canonical project path, when looking up directory name, then `GetDirectoryName(path)` returns the mapped directory name from config.

8. **AC8: Path lookup from directory** - Given directory name (e.g., `api-service`), when looking up original path, then `GetProjectPath(directoryName)` returns the canonical path.

9. **AC9: Context cancellation** - Given config load/save operation in progress, when context is cancelled, then operation returns `ctx.Err()` promptly.

10. **AC10: ProjectPathLookup interface** - Given `Config` struct, when passed to `DirectoryManager` constructor, then it satisfies `ports.ProjectPathLookup` interface via `GetDirForPath(path string) string` method.

## Tasks / Subtasks

- [x] Task 1: Update master config schema to v2 format (AC: 1, 2, 4) **ALREADY DONE**
  - [x] Subtask 1.1: `StorageVersion int` field exists at `ports/config.go:14-15`
  - [x] Subtask 1.2: `NewConfig()` sets `StorageVersion: 2` at `ports/config.go:72`
  - [x] Subtask 1.3: `DirectoryName string` field exists at `ports/config.go:45-48`
  - [x] Subtask 1.4: `Validate()` validates storage_version at `ports/config.go:111-113`

- [x] Task 2: Update `ViperLoader` to handle v2 format (AC: 1, 3, 5, 6, 9)
  - [x] Subtask 2.1: Update `mapViperToConfig()` to read `storage_version` from YAML root
  - [x] Subtask 2.2: Update `mapViperToConfig()` - v2 uses directory_name as map key (key IS directory_name)
  - [x] Subtask 2.3: Update `Save()` to write `storage_version: 2` at YAML root level
  - [x] Subtask 2.4: Update `Save()` to write `directory_name` field for each project entry
  - [x] Subtask 2.5: Add `migrateV1ToV2()` for backward compatibility when `storage_version` missing/1
  - [x] Subtask 2.6: Update `fixInvalidValues()` to handle invalid storage_version (set to 2)

- [x] Task 3: Implement path index lookup methods AND `ProjectPathLookup` interface (AC: 7, 8, 10)
  - [x] Subtask 3.1: Add `GetDirForPath(canonicalPath string) string` method (implements `ports.ProjectPathLookup` for DirectoryManager)
  - [x] Subtask 3.2: Add `GetDirectoryName(path string) (string, bool)` method (wrapper with bool return)
  - [x] Subtask 3.3: Add `GetProjectPath(directoryName string) (string, bool)` method
  - [x] Subtask 3.4: Add `SetProjectEntry(directoryName, path, displayName string, favorite bool)` method
  - [x] Subtask 3.5: Add `RemoveProject(directoryName string) bool` method
  - [x] Subtask 3.6: Add compile-time interface check: `var _ ports.ProjectPathLookup = (*Config)(nil)`

- [x] Task 4: Update default config file template (AC: 1, 4)
  - [x] Subtask 4.1: Update `writeDefaultConfig()` in `loader.go` to include `storage_version: 2` at top
  - [x] Subtask 4.2: Update template comments to document v2 format (directory_name as key)

- [x] Task 5: Gracefully deprecate per-project overrides from master config (AC: 4)
  - [x] Subtask 5.1: `HibernationDays *int` already marked DEPRECATED at `ports/config.go:57-60` - keep for migration
  - [x] Subtask 5.2: `AgentWaitingThresholdMinutes *int` already marked DEPRECATED at `ports/config.go:62-65` - keep for migration
  - [x] Subtask 5.3: Update `Save()` to NOT write these deprecated fields
  - [x] Subtask 5.4: Keep `mapViperToConfig()` reading these fields for v1 backward compatibility
  - [x] Subtask 5.5: Add `slog.Warn` when deprecated fields are read from v1 config

- [x] Task 6: Write unit tests (AC: 1-9)
  - [x] Subtask 6.1: Test `NewConfig()` sets storage_version to 2 - exists at `config_test.go:409-415`
  - [x] Subtask 6.2: Test `Validate()` validates storage_version - exists at `config_test.go:417-458`
  - [x] Subtask 6.3: Test `Save()` writes storage_version and directory_name
  - [x] Subtask 6.4: Test `GetDirForPath()` returns correct directory name (ProjectPathLookup interface)
  - [x] Subtask 6.5: Test `GetDirectoryName()` finds correct mapping
  - [x] Subtask 6.6: Test `GetProjectPath()` reverse lookup works
  - [x] Subtask 6.7: Test `SetProjectEntry()` adds new project with directory_name
  - [x] Subtask 6.8: Test `RemoveProject()` removes project entry
  - [x] Subtask 6.9: Test v1 to v2 migration with graceful fallback (missing storage_version)
  - [x] Subtask 6.10: Test v1 migration collision warning (duplicate base names)
  - [x] Subtask 6.11: Test context cancellation for Load and Save
  - [x] Subtask 6.12: Test empty Path field handling (skip entry, log warning)

- [x] Task 7: Write integration tests (AC: 2, 3, 6)
  - [x] Subtask 7.1: Test full lifecycle: add project → save → load → verify directory mapping
  - [x] Subtask 7.2: Test remove project → save → load → verify entry removed
  - [x] Subtask 7.3: Test `Config` as `ProjectPathLookup` with `DirectoryManager` (verify determinism)

## Dev Notes

### Master Config v2 Format (Target)

```yaml
# ~/.vibe-dash/config.yaml
storage_version: 2

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects:
  api-service:                      # directory_name as key
    path: "/home/user/api-service"  # canonical path
    favorite: false
  client-b-api-service:             # disambiguated directory name
    path: "/home/user/client-b/api-service"
    display_name: "Client B API"
    favorite: true
```

### Key Changes from Current Format

| Aspect | Current (v1) | New (v2) |
|--------|--------------|----------|
| Storage version | Not present | `storage_version: 2` at top |
| Project key | Arbitrary ID (path hash) | `directory_name` (matches subdirectory) |
| Per-project hibernation | In master config | Moved to `~/.vibe-dash/<project>/config.yaml` |
| Per-project waiting threshold | In master config | Moved to `~/.vibe-dash/<project>/config.yaml` |
| Directory name | Not stored | Explicit `directory_name` or key itself |

### Architecture Alignment

**PRD Specification (lines 607-621):**
```yaml
# Master configuration index (single source of truth)
storage_version: 2

settings:
  hibernation_days: 14              # Global default
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects:
  api-service:                      # Directory name as key
    path: "/home/user/work/api-service"
    favorite: false
```

**Configuration Priority (unchanged from PRD):**
```
1. CLI flags           --hibernation-days=7
2. Project config      ~/.vibe-dash/<project>/config.yaml  (Story 3.5.3)
3. Master config       ~/.vibe-dash/config.yaml            (THIS STORY)
4. Built-in defaults   hibernation_days: 14
```

**Architecture (lines 309-329):**
> Configuration Precedence (Explicit):
> 1. CLI flags (highest priority)
> 2. Project config
> 3. Master config
> 4. Built-in defaults (lowest priority)

### Files to Modify

| File | Changes |
|------|---------|
| `internal/core/ports/config.go` | Add `StorageVersion`, `DirectoryName` fields; add lookup methods |
| `internal/core/ports/config_test.go` | Add tests for new fields and methods |
| `internal/config/loader.go` | Update `mapViperToConfig()`, `Save()`, `writeDefaultConfig()` |
| `internal/config/loader_test.go` | Add tests for v2 format, migration |

### Updated `ports.Config` Struct

```go
// Config represents application configuration settings.
// Storage format v2 uses directory names as project keys.
type Config struct {
    // StorageVersion is the config format version (must be 2)
    StorageVersion int

    // HibernationDays is the number of days of inactivity before auto-hibernation (FR28)
    // Default: 14. Set to 0 to disable auto-hibernation.
    HibernationDays int

    // RefreshIntervalSeconds is the TUI refresh interval in seconds
    // Default: 10. Must be > 0.
    RefreshIntervalSeconds int

    // RefreshDebounceMs is the debounce delay for file events in milliseconds
    // Default: 200. Must be > 0.
    RefreshDebounceMs int

    // AgentWaitingThresholdMinutes is minutes of inactivity before showing WAITING (FR34-38)
    // Default: 10. Set to 0 to disable waiting detection.
    AgentWaitingThresholdMinutes int

    // Projects maps directory_name to project configuration
    // Key is the directory name under ~/.vibe-dash/
    Projects map[string]ProjectConfig
}
```

### Updated `ports.ProjectConfig` Struct

**NOTE:** `HibernationDays` and `AgentWaitingThresholdMinutes` are kept for backward compatibility during v1→v2 migration but marked DEPRECATED. The `Save()` method will NOT write these fields.

```go
// ProjectConfig represents per-project configuration in master config (FR47).
// Per-project setting overrides (hibernation, waiting threshold) are now in
// per-project config files (~/.vibe-dash/<project>/config.yaml) - see Story 3.5.3.
type ProjectConfig struct {
    Path          string  // Canonical absolute path
    DirectoryName string  // Subdirectory name under ~/.vibe-dash/ (mirrors map key)
    DisplayName   string  // User-set custom name (FR5), empty = use derived name
    IsFavorite    bool    // Always visible regardless of activity (FR30)

    // DEPRECATED: Use per-project config file instead (Story 3.5.3)
    // Kept for v1→v2 migration backward compatibility - NOT written on Save()
    HibernationDays              *int
    AgentWaitingThresholdMinutes *int
}
```

### Path Index Methods + ProjectPathLookup Interface

**CRITICAL:** `Config` must implement `ports.ProjectPathLookup` interface so `DirectoryManager` can use it for deterministic directory naming.

```go
// Compile-time interface compliance check
var _ ports.ProjectPathLookup = (*Config)(nil)

// GetDirForPath implements ports.ProjectPathLookup interface.
// Used by DirectoryManager to ensure same path always returns same directory name.
// Returns empty string if path not registered.
func (c *Config) GetDirForPath(canonicalPath string) string {
    for dirName, pc := range c.Projects {
        if pc.Path == canonicalPath {
            return dirName
        }
    }
    return ""
}

// GetDirectoryName returns the directory name for a project path.
// Wrapper around GetDirForPath with bool return for convenience.
func (c *Config) GetDirectoryName(path string) (string, bool) {
    dirName := c.GetDirForPath(path)
    return dirName, dirName != ""
}

// GetProjectPath returns the canonical path for a directory name.
func (c *Config) GetProjectPath(directoryName string) (string, bool) {
    if pc, ok := c.Projects[directoryName]; ok {
        return pc.Path, true
    }
    return "", false
}

// SetProjectEntry adds or updates a project entry in the config.
func (c *Config) SetProjectEntry(directoryName, path, displayName string, favorite bool) {
    if c.Projects == nil {
        c.Projects = make(map[string]ProjectConfig)
    }
    c.Projects[directoryName] = ProjectConfig{
        Path:          path,
        DirectoryName: directoryName,
        DisplayName:   displayName,
        IsFavorite:    favorite,
    }
}

// RemoveProject removes a project entry from the config.
func (c *Config) RemoveProject(directoryName string) bool {
    if _, ok := c.Projects[directoryName]; ok {
        delete(c.Projects, directoryName)
        return true
    }
    return false
}
```

### V1 to V2 Migration

```go
// migrateV1ToV2 attempts to migrate v1 config format to v2.
// V1 used arbitrary project IDs; v2 uses directory names.
// Deprecated fields are warned but preserved for read compatibility.
func (l *ViperLoader) migrateV1ToV2(cfg *ports.Config) *ports.Config {
    slog.Warn("migrating config from v1 to v2 format")
    cfg.StorageVersion = 2

    // Migrate projects - use base directory name as key
    newProjects := make(map[string]ports.ProjectConfig)
    for oldKey, pc := range cfg.Projects {
        if pc.Path == "" {
            slog.Warn("migration: skipping project with empty path", "key", oldKey)
            continue
        }

        // Warn about deprecated fields (still readable but won't be written)
        if pc.HibernationDays != nil {
            slog.Warn("migration: hibernation_days in master config is deprecated, use per-project config",
                "path", pc.Path, "value", *pc.HibernationDays)
        }
        if pc.AgentWaitingThresholdMinutes != nil {
            slog.Warn("migration: agent_waiting_threshold_minutes in master config is deprecated, use per-project config",
                "path", pc.Path, "value", *pc.AgentWaitingThresholdMinutes)
        }

        // Use base name of path as directory name
        dirName := filepath.Base(pc.Path)

        // Skip if collision (warn user)
        if _, exists := newProjects[dirName]; exists {
            slog.Warn("migration collision - project skipped, please re-add",
                "path", pc.Path, "directory_name", dirName)
            continue
        }

        pc.DirectoryName = dirName
        newProjects[dirName] = pc
    }
    cfg.Projects = newProjects

    return cfg
}
```

### Updated Save() Method (Excerpt)

**CRITICAL:** `Save()` must write `storage_version` and NOT write deprecated fields.

```go
func (l *ViperLoader) Save(ctx context.Context, config *ports.Config) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    if err := l.ensureConfigDir(); err != nil {
        return fmt.Errorf("failed to create config directory: %w", err)
    }

    // CRITICAL: Write storage_version at root level
    l.v.Set("storage_version", config.StorageVersion)

    // Global settings
    l.v.Set("settings.hibernation_days", config.HibernationDays)
    l.v.Set("settings.refresh_interval_seconds", config.RefreshIntervalSeconds)
    l.v.Set("settings.refresh_debounce_ms", config.RefreshDebounceMs)
    l.v.Set("settings.agent_waiting_threshold_minutes", config.AgentWaitingThresholdMinutes)

    // Projects - directory_name as key, do NOT write deprecated fields
    projects := make(map[string]interface{})
    for dirName, pc := range config.Projects {
        projectData := map[string]interface{}{
            "path":           pc.Path,
            "directory_name": pc.DirectoryName, // Explicit for clarity
        }
        if pc.DisplayName != "" {
            projectData["display_name"] = pc.DisplayName
        }
        if pc.IsFavorite {
            projectData["favorite"] = pc.IsFavorite
        }
        // NOTE: HibernationDays and AgentWaitingThresholdMinutes are NOT written
        // These are deprecated - use per-project config files instead (Story 3.5.3)
        projects[dirName] = projectData
    }
    l.v.Set("projects", projects)

    return l.v.WriteConfig()
}
```

### Updated writeDefaultConfig() (Excerpt)

```go
func (l *ViperLoader) writeDefaultConfig() error {
    cfg := ports.NewConfig()

    content := fmt.Sprintf(`# Vibe Dashboard Master Configuration
# Auto-generated - storage_version: 2 format
# Per-project settings are in ~/.vibe-dash/<project>/config.yaml

storage_version: 2

settings:
  hibernation_days: %d
  refresh_interval_seconds: %d
  refresh_debounce_ms: %d
  agent_waiting_threshold_minutes: %d

# Projects map: directory_name → project info
# Keys are subdirectory names under ~/.vibe-dash/
projects: {}
`, cfg.HibernationDays, cfg.RefreshIntervalSeconds, cfg.RefreshDebounceMs, cfg.AgentWaitingThresholdMinutes)

    return os.WriteFile(l.configPath, []byte(content), 0644)
}
```

### Error Handling

| Error | When | Recovery |
|-------|------|----------|
| Invalid storage_version | storage_version != 2 | Attempt migration if v1, fail if unknown |
| Missing project path | Project entry without path | Skip entry, log warning |
| Duplicate directory name | Migration collision | Skip duplicate, warn user to re-add |
| `ctx.Err()` | Context cancelled | Return immediately |

### Dependencies

**Depends on (COMPLETED):**
- Story 3.5.1: DirectoryManager - Provides directory name generation algorithm
- Story 3.5.3: Per-Project Config Files - Per-project overrides now live there

**Required by:**
- Story 3.5.5: Repository Coordinator - Uses master config to enumerate projects
- Story 3.5.6: Update CLI Commands - `vibe add` and `vibe remove` update master config

### Previous Story Learnings (3.5.3)

From Story 3.5.3 (Per-Project Config Files) - commit `46110f5`:
- Graceful degradation: log warnings, return defaults on errors (not errors)
- Context cancellation check at start of all Load/Save methods
- Table-driven tests are required per Architecture
- Use `slog.Warn` for recoverable issues
- Validation method should return domain errors wrapped with details

**Key Pattern from 3.5.3:**
```go
// Context cancellation pattern (REQUIRED at start of Load/Save):
select {
case <-ctx.Done():
    return nil, ctx.Err()
default:
}

// Graceful degradation pattern:
if err := l.v.ReadInConfig(); err != nil {
    slog.Warn("config syntax error, using defaults",
        "error", err, "path", l.configPath)
    return ports.NewConfig(), nil
}
```

### Integration with DirectoryManager

The `DirectoryManager` from Story 3.5.1 generates directory names with collision resolution. The master config must store these directory names exactly as DirectoryManager generates them:

```go
// Example usage (in future Story 3.5.6):
dirName, err := directoryManager.EnsureProjectDir(ctx, canonicalPath)
if err != nil {
    return err
}

// Store in master config with directory name as key
cfg.SetProjectEntry(dirName, canonicalPath, "", false)
if err := configLoader.Save(ctx, cfg); err != nil {
    return err
}
```

### Testing Strategy

**Unit Tests (table-driven):**

```go
// Test ProjectPathLookup interface compliance
func TestConfig_GetDirForPath(t *testing.T) {
    tests := []struct {
        name     string
        projects map[string]ports.ProjectConfig
        path     string
        wantDir  string
    }{
        {
            name: "existing project returns directory name",
            projects: map[string]ports.ProjectConfig{
                "api-service": {Path: "/home/user/api-service", DirectoryName: "api-service"},
            },
            path:    "/home/user/api-service",
            wantDir: "api-service",
        },
        {
            name:     "non-existent path returns empty string",
            projects: map[string]ports.ProjectConfig{},
            path:     "/non/existent",
            wantDir:  "",
        },
        {
            name: "disambiguated directory name",
            projects: map[string]ports.ProjectConfig{
                "client-b-api-service": {Path: "/home/user/client-b/api-service", DirectoryName: "client-b-api-service"},
            },
            path:    "/home/user/client-b/api-service",
            wantDir: "client-b-api-service",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := &ports.Config{StorageVersion: 2, Projects: tt.projects}
            got := cfg.GetDirForPath(tt.path)
            require.Equal(t, tt.wantDir, got)
        })
    }
}

func TestConfig_GetDirectoryName(t *testing.T) {
    tests := []struct {
        name      string
        projects  map[string]ports.ProjectConfig
        path      string
        wantDir   string
        wantFound bool
    }{
        {
            name: "existing project",
            projects: map[string]ports.ProjectConfig{
                "api-service": {Path: "/home/user/api-service", DirectoryName: "api-service"},
            },
            path:      "/home/user/api-service",
            wantDir:   "api-service",
            wantFound: true,
        },
        {
            name:      "non-existent project",
            projects:  map[string]ports.ProjectConfig{},
            path:      "/non/existent",
            wantDir:   "",
            wantFound: false,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := &ports.Config{StorageVersion: 2, Projects: tt.projects}
            gotDir, gotFound := cfg.GetDirectoryName(tt.path)
            require.Equal(t, tt.wantDir, gotDir)
            require.Equal(t, tt.wantFound, gotFound)
        })
    }
}

// Test interface compliance at compile time
var _ ports.ProjectPathLookup = (*ports.Config)(nil)
```

**Loader Tests (v2 format + migration):**

```go
func TestViperLoader_Save_WritesStorageVersion(t *testing.T) {
    dir := t.TempDir()
    configPath := filepath.Join(dir, "config.yaml")
    loader := NewViperLoader(configPath)

    cfg := ports.NewConfig()
    cfg.SetProjectEntry("api-service", "/path/to/api", "", false)

    err := loader.Save(context.Background(), cfg)
    require.NoError(t, err)

    // Read raw YAML and verify storage_version present
    data, _ := os.ReadFile(configPath)
    require.Contains(t, string(data), "storage_version: 2")
    // Verify deprecated fields NOT written
    require.NotContains(t, string(data), "hibernation_days:")
}

func TestViperLoader_Load_MigratesV1(t *testing.T) {
    dir := t.TempDir()
    configPath := filepath.Join(dir, "config.yaml")

    // Write v1 format (no storage_version)
    v1Config := `settings:
  hibernation_days: 14
projects:
  old-key:
    path: "/home/user/my-project"
    favorite: true
    hibernation_days: 7
`
    os.WriteFile(configPath, []byte(v1Config), 0644)

    loader := NewViperLoader(configPath)
    cfg, err := loader.Load(context.Background())
    require.NoError(t, err)
    require.Equal(t, 2, cfg.StorageVersion)
    // Project should be migrated with base name as key
    _, ok := cfg.GetProjectPath("my-project")
    require.True(t, ok)
}
```

### Manual Testing

After implementation:

1. **Verify v2 config creation:**
   ```bash
   rm ~/.vibe-dash/config.yaml
   # Run vibe (after Story 3.5.6)
   cat ~/.vibe-dash/config.yaml
   # Should show storage_version: 2
   ```

2. **Verify project mapping:**
   ```bash
   vibe add /path/to/api-service
   cat ~/.vibe-dash/config.yaml
   # projects:
   #   api-service:
   #     path: "/path/to/api-service"
   #     favorite: false
   ```

3. **Verify lookup methods:**
   ```bash
   # In Go test:
   cfg.GetDirectoryName("/path/to/api-service") // → "api-service", true
   cfg.GetProjectPath("api-service") // → "/path/to/api-service", true
   ```

### Project Structure Notes

Alignment with PRD storage structure after this story:
```
~/.vibe-dash/
  ├── config.yaml                 # THIS STORY - Master index (storage_version: 2)
  ├── api-service/
  │   ├── .project-path           # Marker file (Story 3.5.1)
  │   ├── config.yaml             # Per-project settings (Story 3.5.3)
  │   └── state.db                # Per-project SQLite (Story 3.5.2)
  └── client-b-api-service/       # Disambiguated name
      ├── .project-path
      ├── config.yaml
      └── state.db
```

### References

- [Source: docs/sprint-artifacts/stories/epic-3.5/epic-3.5-storage-structure.md#Story 3.5.4]
- [Source: docs/prd.md#lines 607-621] - Master config format
- [Source: docs/architecture.md#lines 309-329] - Configuration cascade
- [Source: docs/project-context.md#Configuration Cascade] - Priority order
- [Source: internal/core/ports/config.go] - Current Config struct
- [Source: internal/config/loader.go] - Current ViperLoader implementation

## Dev Agent Record

### Context Reference

- Epic 3.5: Storage Structure Alignment
- Story Dependencies: Depends on 3.5.1, 3.5.3; required by 3.5.5, 3.5.6
- PRD Reference: Lines 607-621 (Master config format)
- Architecture Reference: Lines 309-329 (Config cascade)
- Project Context: Lines 58-67 (Configuration Cascade)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - All tests passed without debugging issues.

### Completion Notes List

1. **v2 Format Implementation**: Implemented `storage_version: 2` at YAML root with directory_name as project map key
2. **Migration**: `migrateV1ToV2()` automatically migrates old configs, logging warnings for deprecated fields and collisions
3. **ProjectPathLookup Interface**: `Config` now implements `ports.ProjectPathLookup` for `DirectoryManager` determinism
4. **Path Index Methods**: Added `GetDirForPath()`, `GetDirectoryName()`, `GetProjectPath()`, `SetProjectEntry()`, `RemoveProject()`
5. **Deprecated Fields**: `HibernationDays` and `AgentWaitingThresholdMinutes` per-project are preserved for read but NOT written on Save
6. **Test Coverage**: Comprehensive unit and integration tests covering all ACs

### Code Review Fixes Applied (2025-12-18)

1. **Fixed stale comment** (`config.go:34`): Updated comment from "path hash" to "directory_name" for v2 format
2. **Added `fixInvalidValues` test**: New test `TestViperLoader_Load_FixesInvalidStorageVersion` verifies storage_version correction
3. **Added deprecated field warning**: `mapViperToConfig()` now logs `slog.Warn` when deprecated per-project fields are read (Subtask 5.5)
4. **Added key precedence test**: New test `TestViperLoader_Load_V2Format_KeyTakesPrecedenceOverDirectoryName` for edge case
5. **Fixed test cases**: Added `StorageVersion: 2` to validation test cases that were failing for wrong reason

### File List

- `internal/core/ports/config.go` - Added path lookup methods, interface compliance check, fixed stale comment
- `internal/core/ports/config_test.go` - Added tests for lookup methods, interface compliance, fixed validation tests
- `internal/config/loader.go` - Updated for v2 format, migration, default template, deprecated field warnings
- `internal/config/loader_test.go` - Added v2 format, migration, integration tests, fixInvalidValues test, key precedence test
