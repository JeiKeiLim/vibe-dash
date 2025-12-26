# Story 8.7: Config Display in TUI

Status: done

## Story

As a **user**,
I want **to see current configuration values in the TUI**,
So that **I know what thresholds and settings are active**.

## Acceptance Criteria

1. **AC1: Config Section in Help Overlay**
   - Given the help overlay (?)
   - When I view it
   - Then I see a "Settings" section with current config values:
     - Agent waiting threshold (minutes)
     - Refresh interval (seconds)
     - Debounce (ms)
     - Detail layout mode
     - Hibernation days

2. **AC2: Config Values Display Format**
   - Given config display in help overlay
   - When values are shown
   - Then each value has a clear label and unit
   - Example: "Waiting threshold: 10 min"

3. **AC3: Per-Project Override Indication** *(OPTIONAL - deferred if time-constrained)*
   - Given a project is selected
   - When I view detail panel
   - Then I see project-specific config if different from global
   - Format: "(project: Xmin)" after global value

4. **AC4: Config Source Path**
   - Given the help overlay
   - When viewing config section
   - Then the config file path is shown (already exists, keep as-is)

5. **AC5: No Crash on Missing Config**
   - Given config values cannot be read
   - When help overlay renders
   - Then default values are shown with no crash

## Tasks / Subtasks

- [x] Task 1: Add Config Accessor to CLI Package (AC: 1, 5)
  - [x] 1.1: Add `appConfig *ports.Config` variable to `cli/deps.go`
  - [x] 1.2: Add `SetConfig(cfg *ports.Config)` function to store config pointer
  - [x] 1.3: Add `GetConfig() *ports.Config` function with nil-safety (returns `ports.NewConfig()` if nil)
  - [x] 1.4: In `main.go`, call `cli.SetConfig(cfg)` after config is loaded (line ~97)

- [x] Task 2: Update Help Overlay to Show Config (AC: 1, 2, 4)
  - [x] 2.1: In `renderHelpOverlay()`, accept config param and add nil-safe default handling
  - [x] 2.2: Add "Settings" section after existing keyboard shortcuts separator
  - [x] 2.3: Display settings using format: `Waiting:     10 min`
  - [x] 2.4: Display: `Refresh:     10 sec`, `Debounce:    200 ms`, `Layout:      horizontal`, `Hibernation: 14 days`
  - [x] 2.5: Keep existing config path hints at bottom unchanged
  - [x] 2.6: Verify box width (52 chars) is sufficient for settings section

- [ ] Task 3: Show Project-Specific Override in Detail Panel (AC: 3) *(OPTIONAL - DEFERRED)*
  - [ ] 3.1: In `renderProject()`, call `cli.GetConfig()` to check for per-project overrides
  - [ ] 3.2: If project has override and differs from global, show "(project: Xm)" after Waiting field
  - [ ] 3.3: Use existing `formatField()` helper for consistent formatting

- [x] Task 4: Add Tests (AC: all)
  - [x] 4.1: Add `TestRenderHelpOverlay_ShowsConfigValues` in `internal/adapters/tui/views_test.go`
  - [x] 4.2: Add `TestRenderHelpOverlay_NilConfig_UsesDefaults` - verify no crash, uses defaults
  - [x] 4.3: Add `TestGetConfig_ReturnsDefaults_WhenNil` in `internal/adapters/cli/deps_test.go`
  - [x] 4.4: Add `TestRenderHelpOverlay_ConfigPathsStillPresent` to verify AC4
  - [ ] 4.5: (Optional) Add `TestDetailPanel_ShowsOverride` if AC3 implemented - DEFERRED

## Dev Notes

### Problem Statement

Users don't know what configuration values are active when running the dashboard. They may have set a custom waiting threshold via CLI flag or config file, but the TUI doesn't show what's actually in effect.

### Implementation Strategy

**Pattern: CLI Package Accessor (follows Story 8.6)**

Use `cli.SetConfig()` / `cli.GetConfig()` pattern - same as `detailLayout` and `configWarning`. This avoids signature changes and keeps views.go simple.

**CRITICAL: Do NOT pass config as parameter to renderHelpOverlay(). Call `cli.GetConfig()` inside the function.**

### Step 1: Add Config Accessor (deps.go)

```go
// internal/adapters/cli/deps.go - add after detailLayout variable (~line 14)

// appConfig stores the loaded configuration for TUI access (Story 8.7).
var appConfig *ports.Config

// SetConfig stores the config for TUI components to access.
func SetConfig(cfg *ports.Config) {
    appConfig = cfg
}

// GetConfig returns the stored config, or defaults if not set.
// CRITICAL: Always returns non-nil - never crashes on missing config.
func GetConfig() *ports.Config {
    if appConfig == nil {
        return ports.NewConfig()
    }
    return appConfig
}
```

### Step 2: Wire in main.go

```go
// cmd/vibe/main.go - add after cli.SetDetailLayout(cfg.DetailLayout) (~line 97)
cli.SetConfig(cfg)
```

### Step 3: Update renderHelpOverlay (views.go)

**IMPORTANT: Do NOT change function signature. Call cli.GetConfig() inside.**

```go
// internal/adapters/tui/views.go - inside renderHelpOverlay()

func renderHelpOverlay(width, height int) string {
    // Story 8.7: Get config for settings display
    cfg := cli.GetConfig()

    // ... existing title and navigation section ...

    content := strings.Join([]string{
        "",
        "Navigation",
        "j/\u2193     Move down",
        "k/\u2191     Move up",
        "",
        "Actions",
        "d        Toggle detail panel",
        "f        Toggle favorite",
        "n        Edit notes",
        "x        Remove project",
        "a        Add project",
        "r        Refresh/rescan",
        "",
        "Views",
        "h        View hibernated projects",
        "",
        "General",
        "?        Show this help",
        "q        Quit",
        "Esc      Cancel/close",
        "",
        "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
        "Settings",
        fmt.Sprintf("Waiting:     %d min", cfg.AgentWaitingThresholdMinutes),
        fmt.Sprintf("Refresh:     %d sec", cfg.RefreshIntervalSeconds),
        fmt.Sprintf("Debounce:    %d ms", cfg.RefreshDebounceMs),
        fmt.Sprintf("Layout:      %s", cfg.DetailLayout),
        fmt.Sprintf("Hibernation: %d days", cfg.HibernationDays),
        "",
        "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
        hintStyle.Render("Config: ~/.vibe-dash/config.yaml"),
        hintStyle.Render("Per-project: ~/.vibe-dash/<project>/config.yaml"),
        "",
        hintStyle.Render("Press any key to close"),
        "",
    }, "\n")
    // ... rest unchanged ...
}
```

### Step 4: Detail Panel Override (OPTIONAL - AC3)

If time permits, add to `detail_panel.go` inside `renderProject()` after Waiting field:

```go
// Story 8.7 (Optional): Show per-project override if different from global
cfg := cli.GetConfig()
if cfg != nil && p.DirectoryName != "" {
    projectCfg, ok := cfg.GetProjectConfig(p.DirectoryName)
    if ok && projectCfg.AgentWaitingThresholdMinutes != nil {
        globalThreshold := cfg.AgentWaitingThresholdMinutes
        if *projectCfg.AgentWaitingThresholdMinutes != globalThreshold {
            overrideText := fmt.Sprintf("(project: %dm)", *projectCfg.AgentWaitingThresholdMinutes)
            lines = append(lines, formatField("Config", overrideText))
        }
    }
}
```

**Note:** Per-project overrides are deprecated (Story 3.5.3), so this AC is low priority.

### Key Code Locations

| File | Purpose |
|------|---------|
| `internal/adapters/cli/deps.go` | Add `appConfig`, `SetConfig()`, `GetConfig()` |
| `cmd/vibe/main.go` | Add `cli.SetConfig(cfg)` after line 97 |
| `internal/adapters/tui/views.go` | Update `renderHelpOverlay()` - add Settings section |
| `internal/adapters/tui/components/detail_panel.go` | (Optional) Add override display |

### Architecture Compliance

- **Modify:** `internal/adapters/cli/deps.go` - add config accessor (follows existing pattern)
- **Modify:** `cmd/vibe/main.go` - call `cli.SetConfig(cfg)`
- **Modify:** `internal/adapters/tui/views.go` - add Settings section to help overlay
- **No signature changes** - `renderHelpOverlay()` stays `(width, height int)`
- **No new files** - extends existing patterns
- **views.go imports cli package** - allowed (both are adapters layer)

### Previous Story Learnings

**From Story 8.6 (Horizontal Split Layout):**
- Config wiring uses `cli.Set*()` / `cli.Get*()` pattern in deps.go
- main.go calls `cli.SetDetailLayout(cfg.DetailLayout)` after config load
- **Follow EXACT same pattern: `cli.SetConfig(cfg)` in main.go, `cli.GetConfig()` in views.go**

**From Story 7.2 (Config Error Handling):**
- Config warnings use `cli.SetConfigWarning()` / `cli.GetConfigWarning()` accessor pattern
- Graceful degradation: never crash on config issues
- Default values from `ports.NewConfig()` if config unavailable

**From Story 3.5 (Help Overlay):**
- `renderHelpOverlay()` already includes config file path hints (lines 96-97)
- Box width is 52 characters - **verified sufficient** for settings section (longest line ~24 chars)
- Use `hintStyle` for non-primary information

### Config Values to Display

| Config Field | Display Label | Format | Max Width |
|-------------|---------------|--------|-----------|
| `AgentWaitingThresholdMinutes` | `Waiting:` | `%d min` | 20 chars |
| `RefreshIntervalSeconds` | `Refresh:` | `%d sec` | 19 chars |
| `RefreshDebounceMs` | `Debounce:` | `%d ms` | 19 chars |
| `DetailLayout` | `Layout:` | `%s` | 22 chars |
| `HibernationDays` | `Hibernation:` | `%d days` | 22 chars |

**Box width verification:** Longest line is 22 chars. With 12-char label alignment: ~34 chars total. Box width 52 is sufficient.

### Per-Project Config Note (AC3 - OPTIONAL)

Per-project overrides are **deprecated** (Story 3.5.3). Only implement AC3 if time permits. The implementation should:
1. Check `cfg.GetProjectConfig(p.DirectoryName)` for override
2. Only show if override differs from global value
3. Use format: `(project: Xm)` after Waiting field

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Change `renderHelpOverlay()` signature | Call `cli.GetConfig()` inside function |
| Pass config through Model | Use `cli.GetConfig()` directly in views |
| Import `ports` in views.go | Import `cli` package (both are adapters) |
| Crash if config is nil | `cli.GetConfig()` always returns non-nil |
| Add config field to Model struct | Not needed - views.go calls cli.GetConfig() directly |

### Testing Strategy

**CLI Package Tests (`internal/adapters/cli/deps_test.go`):**
```go
func TestGetConfig_ReturnsDefaults_WhenNil(t *testing.T) {
    // Reset state
    appConfig = nil

    cfg := GetConfig()
    assert.NotNil(t, cfg)
    assert.Equal(t, 10, cfg.AgentWaitingThresholdMinutes)
}

func TestSetConfig_StoresConfig(t *testing.T) {
    cfg := &ports.Config{AgentWaitingThresholdMinutes: 5}
    SetConfig(cfg)

    got := GetConfig()
    assert.Equal(t, 5, got.AgentWaitingThresholdMinutes)
}
```

**Views Tests (`internal/adapters/tui/views_test.go`):**
```go
func TestRenderHelpOverlay_ShowsConfigValues(t *testing.T) {
    // Set config with known values
    cli.SetConfig(&ports.Config{
        AgentWaitingThresholdMinutes: 15,
        RefreshIntervalSeconds: 30,
    })

    result := renderHelpOverlay(80, 40)

    assert.Contains(t, result, "Settings")
    assert.Contains(t, result, "15 min")
    assert.Contains(t, result, "30 sec")
}
```

### References

| Document | Relevance |
|----------|-----------|
| `docs/project-context.md` | Story Completion - User verification required |
| `internal/adapters/tui/views.go` | `renderHelpOverlay()` - add Settings section |
| `internal/adapters/cli/deps.go` | Pattern: `detailLayout` accessor |
| `internal/core/ports/config.go` | Config struct with all fields |
| `cmd/vibe/main.go:97` | Add `cli.SetConfig(cfg)` after `cli.SetDetailLayout()`

## User Testing Guide

**Time needed:** 3-5 minutes

### Step 1: Default Config Display

```bash
# Ensure default config (or no config file)
make build && ./bin/vibe
# Press '?' for help overlay
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Settings section | Visible after keyboard shortcuts | Missing |
| Waiting threshold | `10 min` (default) | Different value |
| Refresh interval | `10 sec` | Different value |
| Layout mode | `horizontal` | Missing |
| Config paths | Still shown at bottom | Missing |

### Step 2: Custom Config Values

```bash
# Edit ~/.vibe-dash/config.yaml:
# settings:
#   agent_waiting_threshold_minutes: 5
#   refresh_interval_seconds: 30
#   hibernation_days: 7

./bin/vibe
# Press '?' for help
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Waiting threshold | `5 min` | Still `10 min` |
| Refresh interval | `30 sec` | Still `10 sec` |
| Hibernation | `7 days` | Still `14 days` |

### Step 3: Detail Panel (Optional AC3)

```bash
# Select a project with per-project override
# Press 'd' to open detail
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Config field | Shows "(project: Xm)" if override exists | Crashes or wrong value |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Settings section missing | Check renderHelpOverlay changes |
| Values wrong | Check cli.SetConfig() called in main.go |
| Crash on '?' | Check nil config handling |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

**Implementation Notes:**

1. **Import cycle prevention**: Story spec suggested using `cli.GetConfig()` directly in views.go, but this causes an import cycle (cli→tui→cli). Solution: Pass config via Model struct and tui.Run() parameter instead.

2. **Architecture decision**: Added config as a parameter to `tui.Run()` and stored in Model struct, following the same pattern as `detailLayout` (Story 8.6). This maintains clean dependency flow: main.go → cli → tui.

3. **Nil-safety**: `renderHelpOverlay()` accepts `*ports.Config` parameter with nil-safe handling - calls `ports.NewConfig()` if nil to show defaults.

4. **AC3 deferred**: Per-project override display (AC3) is marked OPTIONAL in the story and per-project overrides are deprecated (Story 3.5.3), so this AC is deferred.

5. **Test coverage**: Added 6 new tests covering config accessor functions and help overlay config display.

**Code Review Fixes (2025-12-27):**

6. **M2 Fix - Model.SetConfig nil-safety**: Added nil check to `Model.SetConfig()` to store defaults if nil config passed. Added explanatory comments about import cycle prevention pattern.

7. **M3 Fix - Test cleanup pattern**: Replaced manual `appConfig = nil` cleanup with `t.Cleanup()` in all config accessor tests for proper test isolation.

8. **L2 Fix - Documentation**: Added comments in `app.go` and `model.go` explaining why config is passed as parameter instead of using `cli.GetConfig()` directly (import cycle avoidance).

### File List

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/adapters/cli/deps.go` | Modified | Added `appConfig`, `SetConfig()`, `GetConfig()` |
| `cmd/vibe/main.go` | Modified | Added `cli.SetConfig(cfg)` after config load |
| `internal/adapters/tui/app.go` | Modified | Added `config` parameter to `Run()` |
| `internal/adapters/tui/model.go` | Modified | Added `config` field and `SetConfig()` method |
| `internal/adapters/tui/views.go` | Modified | Added Settings section to `renderHelpOverlay()` |
| `internal/adapters/cli/root.go` | Modified | Pass `appConfig` to `tui.Run()` |
| `internal/adapters/cli/deps_test.go` | Created | Tests for `SetConfig()`/`GetConfig()` |
| `internal/adapters/tui/views_test.go` | Modified | Added Story 8.7 config display tests |
