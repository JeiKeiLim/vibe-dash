# Story 7.11: M3 - Shared Styles Package

Status: done

## Story

As a **developer maintaining vibe-dash TUI components**,
I want **Lipgloss styles centralized in a shared package**,
So that **I don't duplicate style definitions across TUI components and can maintain visual consistency in one place**.

## Problem Statement

Technical debt item M3 has been carried forward 6 times since Epic 3. The codebase has accumulated duplicated Lipgloss style definitions across TUI files that should be centralized.

**Current State (Verified):**

| File | Duplicated Styles | Lines |
|------|------------------|-------|
| `tui/styles.go` | Primary definitions: `boxStyle`, `titleStyle`, `hintStyle`, `SelectedStyle`, `WaitingStyle`, `RecentStyle`, `ActiveStyle`, `UncertainStyle`, `FavoriteStyle`, `WarningStyle`, `DimStyle`, `BorderStyle` | ~80 lines |
| `tui/components/status_bar.go` | `statusBarWaitingStyle`, `statusBarWaitingZeroStyle`, `statusBarWarningStyle` (redefined from styles.go) | ~12 lines |
| `tui/components/detail_panel.go` | `detailBorderStyle`, `detailTitleStyle`, `uncertainStyle`, `detailWaitingStyle` (redefined from styles.go) | ~18 lines |
| `tui/components/delegate.go` | `selectedStyle`, `waitingStyle`, `recentStyle`, `activeStyle`, `dimStyle`, `favoriteStyle` (redefined from styles.go) | ~20 lines |

**Root Cause:** Go's import cycle prevention (`components` cannot import `tui` package). Each component file duplicates styles from `tui/styles.go` with comments like "mirrored from tui/styles.go to avoid import cycle".

**Impact:**
- Style changes require updates in 4 files
- Risk of visual inconsistency if one file not updated
- ~50 lines of maintenance burden
- Comments warn "Keep in sync with styles.go" but no enforcement

## Acceptance Criteria

1. **AC1: Shared Styles Package Created**
   - Given a new package at `internal/shared/styles/`
   - When imported by TUI files
   - Then provides all Lipgloss style definitions
   - And follows the existing `internal/shared/` pattern (like `testhelpers/`, `timeformat/`)

2. **AC2: Core Styles Consolidated**
   - Given the shared package
   - When TUI components need styles
   - Then `styles.SelectedStyle`, `styles.WaitingStyle`, etc. available
   - And all current styles from `tui/styles.go` exported

3. **AC3: Import Cycle Resolved**
   - Given `tui/components/` needs styles
   - When importing `internal/shared/styles`
   - Then no import cycle (shared package has no TUI dependencies)
   - And components can use same style instances as `tui/styles.go`

4. **AC4: styles.go Updated to Re-export**
   - Given existing `tui/styles.go` file
   - When refactored
   - Then re-exports from shared package (backward compatible)
   - And helper functions (`ApplySelected`, `ApplyIndicator`) remain in place
   - And `init()` function for color profile setup remains in `tui/styles.go`

5. **AC5: Components Updated**
   - Given `status_bar.go`, `detail_panel.go`, `delegate.go`
   - When updated to use shared package
   - Then local style redefinitions removed
   - And imports `github.com/JeiKeiLim/vibe-dash/internal/shared/styles`

6. **AC6: UseColor Variable Consolidated**
   - Given `UseColor` bool in `tui/styles.go`
   - When moved to shared package
   - Then `styles.UseColor` available
   - And color profile initialization remains in `tui/styles.go` (near Bubble Tea)

7. **AC7: All Tests Pass**
   - Given refactoring complete
   - When running `go test ./...`
   - Then all existing tests pass
   - And no test regressions

8. **AC8: Package Documented**
   - Given shared styles package
   - When developers read package doc
   - Then clear documentation of each style's purpose
   - And each public variable has godoc

## Tasks / Subtasks

- [x] Task 1: Create shared styles package structure (AC: 1, 8)
  - [x] 1.1: Create directory `internal/shared/styles/`
  - [x] 1.2: Create `doc.go` with package documentation
  - [x] 1.3: Verify package imports correctly

- [x] Task 2: Move core styles to shared package (AC: 2, 6)
  - [x] 2.1: Create `internal/shared/styles/styles.go`
  - [x] 2.2: Move `UseColor` variable to shared package
  - [x] 2.3: Move all exported styles: `SelectedStyle`, `WaitingStyle`, `RecentStyle`, `ActiveStyle`, `UncertainStyle`, `FavoriteStyle`, `WarningStyle`, `DimStyle`, `BorderStyle`
  - [x] 2.4: Move unexported base styles as exported: `BoxStyle`, `TitleStyle`, `HintStyle`
  - [x] 2.5: Add godoc for each style explaining its purpose

- [x] Task 3: Update tui/styles.go to re-export (AC: 4)
  - [x] 3.1: Import shared styles package
  - [x] 3.2: Create type aliases or variable assignments to maintain backward compatibility
  - [x] 3.3: Keep `ApplySelected()` and `ApplyIndicator()` helper functions (they use shared styles)
  - [x] 3.4: Keep `init()` function for `lipgloss.SetColorProfile()`
  - [x] 3.5: Run tests, verify `tui` package works

- [x] Task 4: Update components/delegate.go (AC: 5, 7)
  - [x] 4.1: Import `internal/shared/styles`
  - [x] 4.2: Remove local style definitions: `selectedStyle`, `waitingStyle`, `recentStyle`, `activeStyle`, `dimStyle`, `favoriteStyle`
  - [x] 4.3: Update references to use `styles.SelectedStyle`, etc.
  - [x] 4.4: Run tests, verify delegate works

- [x] Task 5: Update components/status_bar.go (AC: 5, 7)
  - [x] 5.1: Import `internal/shared/styles`
  - [x] 5.2: Remove local style definitions: `statusBarWaitingStyle`, `statusBarWaitingZeroStyle`, `statusBarWarningStyle`
  - [x] 5.3: Update references to use `styles.WaitingStyle`, `styles.DimStyle`, `styles.WarningStyle`
  - [x] 5.4: Note: `statusBarWaitingZeroStyle` is just `DimStyle` - use that
  - [x] 5.5: Run tests, verify status bar works

- [x] Task 6: Update components/detail_panel.go (AC: 5, 7)
  - [x] 6.1: Import `internal/shared/styles`
  - [x] 6.2: Remove local style definitions: `detailBorderStyle`, `detailTitleStyle`, `uncertainStyle`, `detailWaitingStyle`
  - [x] 6.3: Update references to use `styles.BorderStyle`, `styles.TitleStyle`, `styles.UncertainStyle`, `styles.WaitingStyle`
  - [x] 6.4: Note: `dimStyle` already imported from delegate.go via package - verify usage
  - [x] 6.5: Run tests, verify detail panel works

- [x] Task 7: Final verification (AC: 7, 8)
  - [x] 7.1: Run `go test ./...` - all tests pass
  - [x] 7.2: Run `golangci-lint run` - no lint errors
  - [x] 7.3: Run `./bin/vibe` - TUI displays correctly with all styling
  - [x] 7.4: Verify no duplicate style definitions remain

## Dev Notes

### Package Location Decision

**Chosen:** `internal/shared/styles/`

**Rationale:**
- Follows existing pattern: `internal/shared/testhelpers/`, `internal/shared/timeformat/`, `internal/shared/project/`
- `internal/` keeps it private to the module
- `shared/` signals cross-package utility
- `styles/` clearly indicates Lipgloss styling purpose

### Style Export Strategy

The key insight is that `tui/components/` cannot import `tui/` (import cycle), but both can import `internal/shared/styles/`.

**Before:**
```
tui/styles.go (defines styles)
    ↓ (can't import - cycle)
tui/components/*.go (redefinies same styles)
```

**After:**
```
internal/shared/styles/styles.go (defines styles)
    ↑                    ↑
tui/styles.go        tui/components/*.go
(re-exports +        (imports directly)
 helper funcs)
```

### Style Mapping Reference

| Current Location | Style Name | Shared Package Name |
|-----------------|------------|---------------------|
| `tui/styles.go` | `boxStyle` (unexported) | `styles.BoxStyle` |
| `tui/styles.go` | `titleStyle` (unexported) | `styles.TitleStyle` |
| `tui/styles.go` | `hintStyle` (unexported) | `styles.HintStyle` |
| `tui/styles.go` | `SelectedStyle` | `styles.SelectedStyle` |
| `tui/styles.go` | `WaitingStyle` | `styles.WaitingStyle` |
| `tui/styles.go` | `RecentStyle` | `styles.RecentStyle` |
| `tui/styles.go` | `ActiveStyle` | `styles.ActiveStyle` |
| `tui/styles.go` | `UncertainStyle` | `styles.UncertainStyle` |
| `tui/styles.go` | `FavoriteStyle` | `styles.FavoriteStyle` |
| `tui/styles.go` | `WarningStyle` | `styles.WarningStyle` |
| `tui/styles.go` | `DimStyle` | `styles.DimStyle` |
| `tui/styles.go` | `BorderStyle` | `styles.BorderStyle` |
| `components/delegate.go` | `selectedStyle` | → `styles.SelectedStyle` |
| `components/delegate.go` | `waitingStyle` | → `styles.WaitingStyle` |
| `components/delegate.go` | `recentStyle` | → `styles.RecentStyle` |
| `components/delegate.go` | `activeStyle` | → `styles.ActiveStyle` |
| `components/delegate.go` | `dimStyle` | → `styles.DimStyle` |
| `components/delegate.go` | `favoriteStyle` | → `styles.FavoriteStyle` |
| `components/status_bar.go` | `statusBarWaitingStyle` | → `styles.WaitingStyle` |
| `components/status_bar.go` | `statusBarWaitingZeroStyle` | → `styles.DimStyle` |
| `components/status_bar.go` | `statusBarWarningStyle` | → `styles.WarningStyle` |
| `components/detail_panel.go` | `detailBorderStyle` | → `styles.BorderStyle` |
| `components/detail_panel.go` | `detailTitleStyle` | → `styles.TitleStyle` |
| `components/detail_panel.go` | `uncertainStyle` | → `styles.UncertainStyle` |
| `components/detail_panel.go` | `detailWaitingStyle` | → `styles.WaitingStyle` |

### Verified Style Inventory (from current source files)

**Source: `delegate.go:29-49` (6 styles to remove):**
```go
selectedStyle  // → styles.SelectedStyle
waitingStyle   // → styles.WaitingStyle
recentStyle    // → styles.RecentStyle
activeStyle    // → styles.ActiveStyle
dimStyle       // → styles.DimStyle
favoriteStyle  // → styles.FavoriteStyle
```

**Source: `status_bar.go:15-28` (3 styles to remove):**
```go
statusBarWaitingStyle     // → styles.WaitingStyle
statusBarWaitingZeroStyle // → styles.DimStyle
statusBarWarningStyle     // → styles.WarningStyle
```

**Source: `detail_panel.go:19-39` (4 styles to remove):**
```go
detailBorderStyle   // → styles.BorderStyle
detailTitleStyle    // → styles.TitleStyle
uncertainStyle      // → styles.UncertainStyle
detailWaitingStyle  // → styles.WaitingStyle
```

### Color Constants (Preserved Exactly)

| ANSI Color | Value | Usage |
|------------|-------|-------|
| 1 | Red | WaitingStyle (bold) |
| 2 | Green | RecentStyle |
| 3 | Yellow | ActiveStyle, WarningStyle (bold) |
| 5 | Magenta | FavoriteStyle |
| 6 | Cyan | SelectedStyle (background) |
| 8 | Bright Black | UncertainStyle, BorderStyle |
| 39 | Cyan | TitleStyle (foreground, bold) |
| 240 | Gray | BoxStyle border |

### Implementation Approach

**Generate shared package:** Copy style definitions from `tui/styles.go` to `internal/shared/styles/styles.go`:
- Export all styles with PascalCase names (e.g., `boxStyle` → `BoxStyle`)
- Move `UseColor` variable to shared package
- Add godoc comments for each exported style
- Reference: Use color constants from mapping above

**Refactor tui/styles.go:** Replace definitions with re-exports:
- Import `github.com/JeiKeiLim/vibe-dash/internal/shared/styles`
- Assign each style: `var SelectedStyle = styles.SelectedStyle`
- Keep `ApplySelected()`, `ApplyIndicator()` helper functions (reference `styles.XxxStyle`)
- Keep `init()` function for `lipgloss.SetColorProfile()`

**Update component files:** Replace local style vars with shared imports:
- Add import: `"github.com/JeiKeiLim/vibe-dash/internal/shared/styles"`
- Delete lines defining local styles (see inventory above)
- Replace references: `selectedStyle` → `styles.SelectedStyle`

**Test files:** No changes needed - test files don't use these styles directly

### Files to Modify

| File | Action | Lines Changed |
|------|--------|---------------|
| `internal/shared/styles/doc.go` | NEW | ~10 |
| `internal/shared/styles/styles.go` | NEW | ~65 |
| `internal/adapters/tui/styles.go` | MODIFY - re-export from shared | ~60 → ~45 |
| `internal/adapters/tui/components/delegate.go` | MODIFY - use shared styles | Remove ~20 lines |
| `internal/adapters/tui/components/status_bar.go` | MODIFY - use shared styles | Remove ~12 lines |
| `internal/adapters/tui/components/detail_panel.go` | MODIFY - use shared styles | Remove ~18 lines |

**Estimated Total:** ~50 lines of duplication removed, ~75 lines in shared package

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Move `init()` function to shared | Keep in `tui/styles.go` near Bubble Tea initialization |
| Change any color values | Keep exact same ANSI colors (see table above) |
| Forget to update all 3 component files | Check all: delegate.go, status_bar.go, detail_panel.go |

### Testing Commands

```bash
# Run all tests
go test ./...

# Run specific TUI tests
go test ./internal/adapters/tui/... -v

# Build and run for visual verification
make build && ./bin/vibe

# Verify no duplicate style definitions remain (CRITICAL)
grep -r "lipgloss.NewStyle" internal/adapters/tui/ --include="*.go" | grep -v "_test.go"
# Expected: Only internal/adapters/tui/styles.go (should NOT list component files)

# Verify shared package is imported correctly
grep -r "shared/styles" internal/adapters/tui/components/ --include="*.go"
# Expected: All 3 component files should show the import
```

### Completion Verification Checklist

Before marking story complete, verify:

- [x] `internal/shared/styles/styles.go` exists with all 12 styles
- [x] `internal/shared/styles/doc.go` exists with package documentation
- [x] `grep -r "lipgloss.NewStyle" internal/adapters/tui/components/` returns only `detail_panel.go:formatField()` dynamic width style (acceptable - not a duplicated style definition)
- [x] All 3 component files import `internal/shared/styles`
- [x] `tui/styles.go` imports shared package and re-exports
- [x] `tui/styles.go` still has `ApplySelected()` and `ApplyIndicator()` functions
- [x] `tui/styles.go` still has `init()` function for color profile
- [x] `go test ./...` passes
- [x] `golangci-lint run` passes
- [x] `./bin/vibe` displays all colors correctly (see visual verification table)

### Project Structure After Completion

```
internal/
├── shared/
│   ├── testhelpers/     # Story 7.10 (complete)
│   ├── timeformat/      # Existing
│   ├── project/         # Existing
│   └── styles/          # NEW - Story 7.11
│       ├── doc.go       # Package documentation
│       └── styles.go    # All Lipgloss style definitions
└── adapters/
    └── tui/
        ├── styles.go           # Re-exports from shared, keeps helpers
        └── components/
            ├── delegate.go     # Uses styles.SelectedStyle etc.
            ├── status_bar.go   # Uses styles.WaitingStyle etc.
            └── detail_panel.go # Uses styles.BorderStyle etc.
```

### Visual Verification Checklist

After refactoring, visually verify in `./bin/vibe`:

| Element | Expected Style | Check |
|---------|---------------|-------|
| Selected project row | Cyan background | [x] |
| WAITING indicator | Bold red | [x] |
| Recent indicator (✨) | Green | [x] |
| Active indicator (⚡) | Yellow | [x] |
| Favorite indicator (⭐) | Magenta | [x] |
| Warning (⚠️) | Yellow | [x] |
| Dim/faint text | Faint/gray | [x] |
| Panel borders | Gray | [x] |
| Detail panel title | Bold cyan | [x] |

### References

- [Source: docs/sprint-artifacts/retrospectives/epic-6-retro-2025-12-25.md] - M3 action item, 6th carry-forward
- [Source: internal/adapters/tui/styles.go:1-122] - Primary style definitions
- [Source: internal/adapters/tui/components/delegate.go:28-49] - Duplicated styles with "mirrored" comment
- [Source: internal/adapters/tui/components/status_bar.go:15-28] - Duplicated styles with "duplicated" comment
- [Source: internal/adapters/tui/components/detail_panel.go:19-39] - Duplicated styles with "duplicated" comment
- [Source: docs/project-context.md:25-37] - Hexagonal architecture boundaries
- [Source: internal/shared/testhelpers/] - Reference for shared package pattern (Story 7.10)

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Verify Tests Pass

```bash
make test
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| All tests pass | `ok` for each package | Any `FAIL` |
| No compile errors | Clean build | Import errors |

### Step 2: Verify Shared Package Exists

```bash
ls -la internal/shared/styles/
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Files exist | 2 `.go` files | Empty directory |
| doc.go present | Package documentation | Missing doc |

### Step 3: Verify Duplicates Removed

```bash
# Should only show styles.go, NOT component files
grep -r "lipgloss.NewStyle" internal/adapters/tui/ --include="*.go" | grep -v "_test.go" | grep -v "styles.go"
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| No duplicates in components | No output or empty | delegate.go, status_bar.go, detail_panel.go listed |

### Step 4: Visual Verification

```bash
make build && ./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Selected row highlighting | Cyan background | No highlighting |
| WAITING text (if any) | Bold red | Different color |
| Recency icons (✨⚡) | Green/Yellow | No color or wrong color |
| Favorite icon (⭐) | Magenta | No color |
| Panel borders | Visible gray lines | Missing borders |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All tests pass, no duplicates, visuals correct | Mark `done` |
| Tests fail | Do NOT approve, investigate failures |
| Duplicates remain in components | Do NOT approve, document which files |
| Visual styling broken | Do NOT approve, compare before/after |

## Dependencies

- Story 7.10 (M2 - Shared Test Helpers) - Complete, provides pattern reference
- No blocking dependencies for this story

## Dev Agent Record

### Context Reference

N/A - Story fully specified with implementation guidance

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- Created `internal/shared/styles/` package with `doc.go` and `styles.go`
- Moved all 12 style definitions (BoxStyle, TitleStyle, HintStyle, SelectedStyle, WaitingStyle, RecentStyle, ActiveStyle, UncertainStyle, FavoriteStyle, WarningStyle, DimStyle, BorderStyle) to shared package
- Moved `UseColor` variable to shared package
- Added comprehensive godoc for each exported style with color reference table
- Updated `tui/styles.go` to re-export from shared package while maintaining backward compatibility
- Kept `ApplySelected()`, `ApplyIndicator()` helper functions and `init()` color profile setup in `tui/styles.go`
- Updated all 3 component files to import and use shared styles:
  - `delegate.go`: Removed 6 style definitions, updated 6 references
  - `status_bar.go`: Removed 3 style definitions, updated 6 references
  - `detail_panel.go`: Removed 4 style definitions, updated 5 references
- All tests pass (`go test ./...`)
- Linter passes (`golangci-lint run`)
- Build succeeds (`make build`)

### File List

| File | Action |
|------|--------|
| `internal/shared/styles/doc.go` | NEW |
| `internal/shared/styles/styles.go` | NEW |
| `internal/adapters/tui/styles.go` | MODIFIED - re-exports from shared |
| `internal/adapters/tui/components/delegate.go` | MODIFIED - uses shared styles |
| `internal/adapters/tui/components/status_bar.go` | MODIFIED - uses shared styles |
| `internal/adapters/tui/components/detail_panel.go` | MODIFIED - uses shared styles |

### Code Review Record

**Reviewed by:** Claude Opus 4.5 (code-review workflow)
**Date:** 2025-12-26

**Issues Found:** 0 High, 1 Medium, 4 Low

**M1 (Accepted as-is):** `detail_panel.go:184` uses `lipgloss.NewStyle()` for dynamic label width formatting. This is a builder pattern for dynamic width, not a duplicated static style definition. The story goal was eliminating static style duplication, and this use case is acceptable.

**L1-L4:** Documentation verification checklists updated to reflect completion status.

**Verdict:** Story APPROVED. All ACs met.
