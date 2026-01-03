# Story 11.6: Hibernation Threshold Configuration

Status: done

## Story

- **As a** user
- **I want** to configure hibernation threshold
- **So that** I can tune auto-hibernation to my project lifecycle

## User-Visible Changes

- **New:** `vibe config set <project> hibernation-days <N>` command sets per-project hibernation threshold
- **New:** Per-project override in `~/.vibe-dash/<project>/config.yaml` via `custom_hibernation_days` field
- **Changed:** `vibe config set --help` now shows `hibernation-days` as a supported key

## Context & Background

### Previous Stories
- **Story 11.1** (Project State Model): Created `StateService` with state transitions
- **Story 11.2** (Auto-Hibernation): Implemented `HibernationService` using `Config.GetEffectiveHibernationDays(projectID)` which checks DEPRECATED master config field
- **Story 11.5** (Manual State Control): CLI `vibe hibernate` and `vibe activate` commands

### The Gap
`HibernationService.CheckAndHibernate()` currently calls:
```go
thresholdDays := h.config.GetEffectiveHibernationDays(project.ID)
```

This method (config.go:109-114) checks:
1. `Config.Projects[projectID].HibernationDays` (DEPRECATED - master config)
2. Falls back to `Config.HibernationDays` (global)

It does NOT check per-project config files (`~/.vibe-dash/<project>/config.yaml`).

This story:
1. Adds `hibernation-days` to `vibe config set` CLI (copy `waiting-threshold` pattern)
2. Updates `HibernationService` to read per-project config files for threshold

### Functional Requirements
- **FR32**: Configure hibernation threshold (global or per-project)
- **FR47**: Per-project override via config file

## Acceptance Criteria

### AC1: Global Threshold (Already Working)
Covered by Story 11.2 via `Config.HibernationDays`. No changes needed.

### AC2: CLI Set Per-Project Hibernation Threshold
- **When** I run `vibe config set my-project hibernation-days 30`
- **Then** `~/.vibe-dash/my-project/config.yaml` contains `custom_hibernation_days: 30`
- **And** output shows: "Set hibernation-days=30 for project my-project"

### AC3: Per-Project Override Takes Precedence
- **Given** global `hibernation_days: 14` and project has `custom_hibernation_days: 30`
- **When** auto-hibernation check runs
- **Then** project uses 30-day threshold (not 14)

### AC4: Disable Per-Project Auto-Hibernation
- **Given** project has `custom_hibernation_days: 0`
- **When** auto-hibernation check runs
- **Then** project never auto-hibernates

### AC5: Invalid Value Rejection
- **When** I run `vibe config set project hibernation-days -5`
- **Then** error shows: "hibernation-days must be >= 0, got -5"
- **And** exit code is 3 (ExitConfigInvalid)

### AC6: Non-Numeric Value Rejection
- **When** I run `vibe config set project hibernation-days abc`
- **Then** error shows: "invalid value for hibernation-days: abc"
- **And** exit code is 3

### AC7: Project Not Found
- **When** I run `vibe config set nonexistent hibernation-days 7`
- **Then** error shows: "project directory not found: nonexistent"
- **And** exit code is non-zero

### AC8: Help Text Updated
- **When** I run `vibe config set --help`
- **Then** "hibernation-days" appears in supported keys list

## Tasks / Subtasks

- [x] Task 1: Add hibernation-days to config set command (AC: #2, #5, #6, #7, #8)
  - [x] 1.1: Add "hibernation-days" case to `runConfigSet()` switch in `config_cmd.go`
  - [x] 1.2: Create `setProjectHibernationDays()` helper (copy `setProjectWaitingThreshold` pattern exactly)
  - [x] 1.3: Update `configSetCmd.Long` to include "hibernation-days" in supported keys
  - [x] 1.4: Validation: `>= 0` (0 disables auto-hibernation for this project)

- [x] Task 2: Update HibernationService to read per-project config (AC: #3, #4)
  - [x] 2.1: Add `vibeHome string` field and constructor param to `HibernationService`
  - [x] 2.2: Create `getEffectiveHibernationDays()` method that loads per-project config
  - [x] 2.3: Replace `h.config.GetEffectiveHibernationDays(project.ID)` call with new method
  - [x] 2.4: Resolve project directory from config (use `Config.GetDirForPath()` pattern)
  - [x] 2.5: Handle config load errors gracefully (log warning, use global)

- [x] Task 3: Update main.go wiring
  - [x] 3.1: Pass `basePath` (vibeHome) to `NewHibernationService()`

- [x] Task 4: Write tests (AC: #2-8)
  - [x] 4.1: Add `TestConfigSet_HibernationDays_*` tests to config_cmd_test.go (copy waiting-threshold test patterns)
  - [x] 4.2: Add `TestHibernationService_PerProjectConfigOverride` to hibernation_service_test.go
  - [x] 4.3: Add `TestHibernationService_PerProjectZeroDisables`
  - [x] 4.4: Add `TestHibernationService_PerProjectConfigLoadError_FallbackToGlobal`

## Technical Implementation Guide

### File Changes

#### 1. `internal/adapters/cli/config_cmd.go` (MODIFY)

Add "hibernation-days" to supported keys and create helper:

```go
// Update configSetCmd.Long to include hibernation-days
Long: `Set a configuration value for a specific project.

Supported keys:
  hibernation-days     Days of inactivity before auto-hibernation (0 to disable)
  waiting-threshold    Agent waiting threshold in minutes (0 to disable)
...`,

// Add case to runConfigSet switch
case "hibernation-days":
    var intVal int
    intVal, err = strconv.Atoi(value)
    if err != nil {
        err = fmt.Errorf("%w: invalid value for hibernation-days: %s", domain.ErrConfigInvalid, value)
    } else if intVal < 0 {
        err = fmt.Errorf("%w: hibernation-days must be >= 0, got %d", domain.ErrConfigInvalid, intVal)
    } else {
        return setProjectHibernationDays(cmd.Context(), cmd, projectID, intVal)
    }

// Add helper (copy setProjectWaitingThreshold pattern EXACTLY)
func setProjectHibernationDays(ctx context.Context, cmd *cobra.Command, projectID string, days int) error {
    projectDir := filepath.Join(vibeHome, projectID)

    if _, err := os.Stat(projectDir); os.IsNotExist(err) {
        return fmt.Errorf("project directory not found: %s", projectID)
    }

    loader, err := config.NewProjectConfigLoader(projectDir)
    if err != nil {
        return fmt.Errorf("failed to access project config: %w", err)
    }

    data, err := loader.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load project config: %w", err)
    }

    data.CustomHibernationDays = &days

    if err := loader.Save(ctx, data); err != nil {
        return fmt.Errorf("failed to save project config: %w", err)
    }

    fmt.Fprintf(cmd.OutOrStdout(), "Set hibernation-days=%d for project %s\n", days, projectID)
    return nil
}
```

#### 2. `internal/core/services/hibernation_service.go` (MODIFY)

Add vibeHome field and per-project config loading:

```go
type HibernationService struct {
    repo         ports.ProjectRepository
    stateService *StateService
    config       *ports.Config
    vibeHome     string // NEW: Base path for per-project configs
}

func NewHibernationService(
    repo ports.ProjectRepository,
    stateService *StateService,
    cfg *ports.Config,
    vibeHome string, // NEW PARAMETER
) *HibernationService {
    return &HibernationService{
        repo:         repo,
        stateService: stateService,
        config:       cfg,
        vibeHome:     vibeHome,
    }
}

// getEffectiveHibernationDays returns threshold for project.
// Priority: per-project config file > global config
func (h *HibernationService) getEffectiveHibernationDays(ctx context.Context, project *domain.Project) int {
    // Resolve directory name from config (canonical path -> dir name)
    dirName := h.config.GetDirForPath(project.Path)
    if dirName == "" {
        // Project not in config - use global
        return h.config.HibernationDays
    }

    projectDir := filepath.Join(h.vibeHome, dirName)
    loader, err := config.NewProjectConfigLoader(projectDir)
    if err != nil {
        return h.config.HibernationDays
    }

    data, err := loader.Load(ctx)
    if err != nil {
        slog.Debug("failed to load per-project config, using global",
            "project", project.Name, "error", err)
        return h.config.HibernationDays
    }

    if data.CustomHibernationDays != nil {
        return *data.CustomHibernationDays
    }

    return h.config.HibernationDays
}

// Update CheckAndHibernate to use new method
func (h *HibernationService) CheckAndHibernate(ctx context.Context) (int, error) {
    // ... existing code ...
    for _, project := range projects {
        // ... existing skip favorite logic ...

        // CHANGE: Use new method instead of h.config.GetEffectiveHibernationDays(project.ID)
        thresholdDays := h.getEffectiveHibernationDays(ctx, project)
        // ... rest unchanged ...
    }
}
```

**CRITICAL:** Import `config` package and `path/filepath`:
```go
import (
    "path/filepath"
    "github.com/JeiKeiLim/vibe-dash/internal/config"
)
```

#### 3. `cmd/vibe/main.go` (MODIFY)

Update HibernationService creation:

```go
// Line ~178: Add basePath to NewHibernationService call
hibernationSvc := services.NewHibernationService(coordinator, stateService, cfg, basePath)
```

### Testing Strategy

| Test Case | File | Pattern |
|-----------|------|---------|
| Set hibernation-days success | config_cmd_test.go | Copy `TestConfigSet_WaitingThreshold_Success` |
| Set hibernation-days negative | config_cmd_test.go | Copy `TestConfigSet_WaitingThreshold_InvalidValue` |
| Set hibernation-days non-numeric | config_cmd_test.go | Copy `TestConfigSet_WaitingThreshold_NonNumeric` |
| Set hibernation-days zero (valid) | config_cmd_test.go | New - verify 0 is accepted |
| Per-project config override | hibernation_service_test.go | Need temp dir with config.yaml |
| Per-project zero disables | hibernation_service_test.go | Verify 0 skips project |
| Config load error fallback | hibernation_service_test.go | Verify graceful fallback |

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Use `Config.GetDirForPath()` for directory resolution | Project.Name is path basename, not config directory key |
| Pass vibeHome to HibernationService | Avoids circular import; keeps core clean |
| Import `config` package in services | Pragmatic - NewProjectConfigLoader is adapter code but needed here |
| Graceful fallback on config errors | Don't break hibernation if one project's config is corrupted |

### Architecture Note
This story introduces a config adapter import into core/services. This is acceptable because:
1. `NewProjectConfigLoader` is a factory function, not interface implementation
2. Alternative (injecting loader factory) adds complexity without benefit
3. The import is isolated to this one service method

## Dev Notes

### Reuse vs Create

| Item | Action | Source |
|------|--------|--------|
| `setProjectWaitingThreshold()` pattern | COPY PATTERN | `config_cmd.go:79-110` |
| `NewProjectConfigLoader()` | REUSE | `internal/config/project_config_loader.go:37` |
| `ProjectConfigData.CustomHibernationDays` | REUSE | `internal/core/ports/project_config.go:19` |
| `Config.GetDirForPath()` | REUSE | `internal/core/ports/config.go:131` |
| CLI test patterns | COPY PATTERN | `config_cmd_test.go` (waiting-threshold tests) |
| `mockHibernationRepo` | REUSE | `hibernation_service_test.go:14` |

### Edge Cases

1. **Project directory doesn't exist**: Error "project directory not found: X"
2. **Config file doesn't exist**: Created automatically by loader
3. **Config file syntax error**: Graceful degradation to global
4. **Project not in master config**: Use global (GetDirForPath returns "")

## File List

**MODIFY:**
- `internal/adapters/cli/config_cmd.go` - Add hibernation-days case + helper
- `internal/adapters/cli/config_cmd_test.go` - Add hibernation-days tests
- `internal/core/services/hibernation_service.go` - Add vibeHome, getEffectiveHibernationDays
- `internal/core/services/hibernation_service_test.go` - Add per-project config tests
- `cmd/vibe/main.go` - Pass basePath to NewHibernationService

**DO NOT MODIFY:**
- `internal/core/ports/project_config.go` - CustomHibernationDays field exists
- `internal/config/project_config_loader.go` - Already parses custom_hibernation_days

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Setup

```bash
make build
mkdir -p /tmp/test-hibernation-config
./bin/vibe add /tmp/test-hibernation-config
./bin/vibe list  # Verify project added
```

### Step 2: Test CLI Config Set

| Check | Expected |
|-------|----------|
| `./bin/vibe config set test-hibernation-config hibernation-days 7` | "Set hibernation-days=7 for project test-hibernation-config" |
| `cat ~/.vibe-dash/test-hibernation-config/config.yaml` | Contains `custom_hibernation_days: 7` |

### Step 3: Test Invalid Values

| Check | Expected |
|-------|----------|
| `./bin/vibe config set test-hibernation-config hibernation-days -5` | Error: "hibernation-days must be >= 0" |
| `./bin/vibe config set test-hibernation-config hibernation-days abc` | Error: "invalid value for hibernation-days" |

### Step 4: Test Zero (Disable)

| Check | Expected |
|-------|----------|
| `./bin/vibe config set test-hibernation-config hibernation-days 0` | "Set hibernation-days=0" (valid) |

### Step 5: Verify Help

| Check | Expected |
|-------|----------|
| `./bin/vibe config set --help` | Shows "hibernation-days" in supported keys |

### Step 6: Cleanup

```bash
./bin/vibe remove test-hibernation-config -y
rm -rf /tmp/test-hibernation-config
```

## Verification Checklist

- [x] `go build ./...` succeeds
- [x] `go test ./internal/adapters/cli/...` passes
- [x] `go test ./internal/core/services/...` passes
- [x] `golangci-lint run` passes
- [x] User Testing Guide Step 2: CLI config set works (tested via automated tests)
- [x] User Testing Guide Step 3: Invalid values rejected (exit code 3 verified)
- [x] User Testing Guide Step 5: Help text updated (hibernation-days in list)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. Added `hibernation-days` case to `runConfigSet()` switch matching existing `waiting-threshold` pattern
2. Created `setProjectHibernationDays()` helper copying `setProjectWaitingThreshold` pattern exactly
3. Updated `configSetCmd.Long` to include `hibernation-days` as first supported key
4. Added `vibeHome` field to `HibernationService` struct and constructor
5. Created `getEffectiveHibernationDays()` method that:
   - Resolves directory name via `Config.GetDirForPath()`
   - Loads per-project config via `NewProjectConfigLoader()`
   - Falls back to global on any error (graceful degradation)
6. Updated existing `TestHibernationService_PerProjectOverride` to use per-project config file instead of deprecated `Config.Projects` map
7. Added 3 new service tests for zero-disables, config-load-error-fallback, and no-config-file scenarios
8. Added 6 CLI tests covering success, zero, negative, non-numeric, project-not-found, and exit-code scenarios

### File List

**MODIFIED:**
- `internal/adapters/cli/config_cmd.go` - Added hibernation-days case and setProjectHibernationDays() helper
- `internal/adapters/cli/config_cmd_test.go` - Added 6 TestConfigSet_HibernationDays_* tests
- `internal/core/services/hibernation_service.go` - Added vibeHome field, getEffectiveHibernationDays() method
- `internal/core/services/hibernation_service_test.go` - Updated PerProjectOverride test, added 3 new tests
- `cmd/vibe/main.go` - Pass basePath to NewHibernationService()
- `internal/core/domain/project.go` - Formatting only (struct field alignment)

### Code Review (2026-01-03)

**Reviewer:** Claude Opus 4.5 (code-review workflow)

**Findings Fixed:**
- M1: Added `internal/core/domain/project.go` to File List (documentation gap)
- M2: Made error message consistent - `setProjectHibernationDays()` now includes `(expected at %s)` like `setProjectWaitingThreshold()`
- L1: Added `TestConfigSet_Help_ShowsHibernationDays` test for AC8 verification
- L2: Updated `createTestConfigCommand()` helper to include Long description for help tests

**All 8 ACs verified. All tests pass. Lint clean.**
