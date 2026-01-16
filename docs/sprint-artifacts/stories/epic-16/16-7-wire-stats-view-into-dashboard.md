# Story 16.7: Wire Stats View into Dashboard

Status: done

## Story

As a user,
I want to access Stats View from the Dashboard,
So that switching between views is seamless.

## User-Visible Changes

- **New:** Help overlay (?) shows `s        View stats and metrics` in Views section

## Acceptance Criteria

1. **Given** I am in Dashboard view
   **When** I press `'s'`
   **Then** Stats View opens

2. **And** status bar shows `[s] stats` hint

3. **And** Stats View preserves Dashboard selection on return

4. **And** Help overlay shows 's' shortcut in Views section

## Implementation Summary

**This story is 95% complete.** Previous stories already implemented:
- Key binding 's' (Story 16.3)
- Key handler calling `enterStatsView()` (Story 16.3)
- Status bar `[s] stats` hint (Story 16.3)
- Dashboard state preservation (Story 16.3)
- Metrics wiring (Stories 16.2, 16.4)

**The ONLY remaining work:** Add 's' shortcut to help overlay in `views.go`.

## Tasks / Subtasks

- [x] **Task 1: Add 's' to Help overlay (AC: #4)**

  **File:** `internal/adapters/tui/views.go`

  **Current code (lines 113-115):**
  ```go
  "Views",
  "h        View hibernated projects",
  "",
  ```

  **Modified code:**
  ```go
  "Views",
  "h        View hibernated projects",
  "s        View stats and metrics",
  "",
  ```

  **Verification command:**
  ```bash
  grep -n "View hibernated" internal/adapters/tui/views.go
  # Should show line 114
  ```

- [x] **Task 2: Verify tests pass**
  ```bash
  make test && make lint
  ```

- [x] **Task 3: Manual verification**
  - Build: `make build`
  - Run: `./bin/vdash`
  - Press `?` - verify 's' appears in Views section
  - Press `s` - verify Stats View opens
  - Press `Esc` - verify dashboard returns with selection preserved

## Dev Notes

### FR Coverage

| Requirement | Description | Status |
|-------------|-------------|--------|
| FR-P2-14 | View Stats View via 's' key | Done (16.3) |
| FR-P2-15 | Exit Stats View via Esc/'q' | Done (16.3) |
| AC #4 | Help overlay shows 's' | Done (16.7) |

### Already Implemented (Do Not Duplicate)

| Component | Location | Story |
|-----------|----------|-------|
| KeyStats = "s" | keys.go:37 | 16.3 |
| Key handler | model.go:2006-2012 | 16.3 |
| Status bar hint | status_bar.go:16,20,191 | 16.3 |
| State preservation | model.go:3395-3410 | 16.3 |
| Metrics wiring | main.go:203-218, app.go:63-70 | 16.2, 16.4 |

### NFR Compliance

| Requirement | Target | Status |
|-------------|--------|--------|
| NFR-P2-5 | Stats View render < 500ms | Validated in 16.3 |
| NFR-P2-7 | Metrics failure doesn't crash | Validated in 16.1 |

### References

- [docs/epics-phase2.md#Story 4.7](../../../epics-phase2.md)
- [Story 16.3: Stats View TUI](./16-3-create-stats-view-tui-component.md)

## Dev Agent Record

### Context Reference

- views.go help overlay: lines 83-165
- Views section: lines 113-115 (insert after line 114)
- Pattern: Follow existing shortcut format with 8-space padding

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Implementation Log

- Added `"s        View stats and metrics",` to help overlay in views.go line 115
- All tests pass (make test)
- Linter passes (make lint)
- Build successful (make build)
- Verified grep shows both 'h' and 's' entries in Views section

### File List

| File | Status | Changes |
|------|--------|---------|
| `internal/adapters/tui/views.go` | MODIFIED | Inserted `"s        View stats and metrics",` at line 115 |
| `internal/adapters/tui/views_test.go` | MODIFIED | Added test for 's' shortcut in Views section (code review fix) |

### Code Review Record

**Review Date:** 2026-01-16
**Reviewer:** Claude Opus 4.5 (claude-opus-4-5-20251101)

**Issues Found:**
- 0 High, 0 Medium, 2 Low

**Fixes Applied:**
- L1: Added test for 's' shortcut in `TestRenderHelpOverlay_ContainsViewShortcuts`
- L2: Noted - line number references in story are historical documentation

**Final Validation:**
- All 1455 tests pass
- Lint clean
- All ACs verified implemented

## User Testing Guide

**Time needed:** 2 minutes

### Quick Verification

1. **Build:** `make build`
2. **Run:** `./bin/vdash`
3. **Test help overlay:** Press `?`, look for `s` in Views section
4. **Test navigation:** Press `s` to open Stats, `Esc` to return

### Expected Results

| Check | Expected |
|-------|----------|
| Help overlay Views section | Shows `h` and `s` entries |
| Press 's' from Dashboard | Stats View opens |
| Press 'Esc' from Stats | Dashboard returns, selection preserved |
| Status bar | Shows `[s] stats` hint |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Help overlay missing 's' | Check views.go line 115 |
| Other features broken | Check Story 16.3 implementation |
