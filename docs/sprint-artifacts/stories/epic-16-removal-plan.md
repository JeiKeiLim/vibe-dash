# Epic 16 Removal Plan - Stats View / Progress Metrics

**Created:** 2026-01-16
**Decision:** Epic 16 retrospective - feature validated as not useful
**Goal:** Clean removal of all Stats View code
**Status:** ✅ COMPLETE

---

## Phase 1: Delete Entire Packages

### 1.1 Delete `internal/adapters/persistence/metrics/`
- [x] Delete directory (9 files)

### 1.2 Delete `internal/adapters/tui/statsview/`
- [x] Delete directory (9 files)

### 1.3 Delete Standalone Files
- [x] Delete `internal/adapters/tui/statsview.go`
- [x] Delete `internal/adapters/tui/statsview_test.go`

**Phase 1 Verification:** ✅ PASS

---

## Phase 2: Remove Wiring (main.go, CLI, app.go)

### 2.1 `cmd/vdash/main.go`
- [x] Remove import: `"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/metrics"`
- [x] Remove metrics initialization, wiring, defer flush
- [x] Remove unused `path/filepath` import

### 2.2 `internal/adapters/cli/deps.go`
- [x] Remove import: `"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/metrics"`
- [x] Remove metricsRecorder, metricsReader vars
- [x] Remove getter/setter functions

### 2.3 `internal/adapters/cli/root.go`
- [x] Update comment (remove metrics references)
- [x] Remove last 2 params from `tui.Run()` call

### 2.4 `internal/adapters/tui/app.go`
- [x] Update comment
- [x] Remove last 2 params from `Run()` signature
- [x] Remove metrics wiring in Run

**Phase 2 Verification:** ✅ PASS

---

## Phase 3: Clean TUI model.go

### 3.1 Remove Interfaces
- [x] Remove metricsRecorderInterface, metricsReaderInterface

### 3.2 Remove State Fields
- [x] Remove all stats-related fields in Model struct

### 3.3 Remove Methods
- [x] Remove SetMetricsRecorder, SetMetricsReader
- [x] Remove getProjectActivity, getStageBreakdown
- [x] Remove enterStatsView, exitStatsView
- [x] Remove metricsRecorder.OnDetection call in refresh

### 3.4 Remove Key Handler
- [x] Remove case KeyStats in handleKeyMsg
- [x] Remove handleStatsViewKeyMsg function

### 3.5 Remove View Rendering
- [x] Remove stats view check in View()
- [x] Remove viewModeStats check in Update()

### 3.6 Remove statsview Import
- [x] Remove import: `"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/statsview"`

**Phase 3 Verification:** ✅ PASS

---

## Phase 4: Update UI Elements

### 4.1 `internal/adapters/tui/validation.go`
- [x] Remove `viewModeStats` from enum

### 4.2 `internal/adapters/tui/keys.go`
- [x] Remove KeyStats constant
- [x] Remove Stats field in KeyBindings struct
- [x] Remove Stats in DefaultKeyBindings

### 4.3 `internal/adapters/tui/components/status_bar.go`
- [x] Remove `[s] stats` from shortcutsFull
- [x] Remove `[s]` from shortcutsAbbrev
- [x] Remove `[s]` from condensed mode

### 4.4 `internal/adapters/tui/views.go`
- [x] Remove `"s        View stats and metrics",`

### 4.5 `internal/adapters/tui/export_test.go`
- [x] Remove stats test helpers

### 4.6 `internal/adapters/tui/views_test.go`
- [x] Remove stats shortcut check from TestRenderHelpOverlay_ContainsViewShortcuts

**Phase 4 Verification:** ✅ PASS

---

## Phase 5: Final Verification

### 5.1 Build & Lint
- [x] `make fmt` passes
- [x] `make lint` passes
- [x] `make build` passes

### 5.2 Tests
- [x] `make test` passes (1340 tests)

### 5.3 Manual Test
- [ ] Run `./bin/vdash`
- [ ] Press `s` - should do nothing
- [ ] Press `?` - help should NOT show 's' shortcut
- [ ] Status bar should NOT show `[s] stats`

### 5.4 Cleanup
- [ ] User deletes `~/.vibe-dash/metrics.db`

---

## Summary

| Phase | Status | Notes |
|-------|--------|-------|
| 1 | ✅ Complete | Deleted 2 packages + 2 standalone files |
| 2 | ✅ Complete | Modified 4 files (main, cli, app) |
| 3 | ✅ Complete | Modified model.go (many scattered changes) |
| 4 | ✅ Complete | Modified 6 files (validation, keys, status_bar, views, export_test, views_test) |
| 5 | ✅ Complete | All automated verification passed |

**Isolation Principle validated** - clean removal with no core impact.
