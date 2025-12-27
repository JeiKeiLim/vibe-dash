# Story 8.10: Full-Width Layout & Column Rebalancing

Status: done

## Story

As a **user viewing the project list**,
I want **column widths optimized for actual content**,
So that **stage info and waiting status are readable without wasted space on short project names**.

## Problem Statement

1. `MaxContentWidth = 120` cap wastes space on wide monitors
2. Current column proportions are unbalanced:
   - Project name: Too wide (names are typically short)
   - Stage info: Too narrow (truncates important info)
   - Waiting status: Too narrow ("WAITING 2h 30m" barely fits)

**Origin:** User feedback during Story 8.4 code review, approved via correct-course workflow 2025-12-26.

## Acceptance Criteria

### Column Rebalancing (All Widths)

```gherkin
AC1: Given normal terminal width (80-120 cols)
     When viewing project list
     Then column proportions prioritize stage and status readability:
     - Project name: ~25% (enough for typical names)
     - Method: fixed ~6 chars
     - Stage info: ~40% (epic/story info needs space)
     - Status/Waiting: ~20% (fit "WAITING 2h 30m")
     - Activity: remaining

AC2: Given a long project name
     When name exceeds allocated width
     Then name truncates with ellipsis (not stage info)
```

### Full-Width Option (Wide Monitors)

```gherkin
AC3: Given config `max_content_width: 0` (disabled)
     When terminal is 200+ columns
     Then content uses full width with proportional columns

AC4: Given config `max_content_width: N` (custom cap)
     When terminal exceeds N columns
     Then content is capped and centered

AC5: Given no config set (default)
     When running
     Then backward-compatible 120-column cap
```

### Column Max Widths (Prevent Absurd Stretching)

```gherkin
AC6: Given ultra-wide terminal (>180 cols)
     When proportional widths calculated
     Then columns have sensible max widths:
     - Project name: max 40 chars
     - Stage info: max 80 chars
     - Status: max 25 chars
```

### Tests Pass

```gherkin
AC7: Given all changes are made
     When `make test && make lint` runs
     Then all tests pass and no lint errors
```

## Tasks / Subtasks

- [x] Task 1: Add `max_content_width` config option (AC: 3, 4, 5)
  - [x] 1.1: In `internal/core/ports/config.go`, add `MaxContentWidth int` field to `Config` struct with comment
  - [x] 1.2: In `internal/core/ports/config.go`, set default value `120` in `NewConfig()`
  - [x] 1.3: In `internal/core/ports/config.go`, add validation (>= 0) to `Validate()` method
  - [x] 1.4: In `internal/config/loader.go`, add Viper binding for `settings.max_content_width`
  - [x] 1.5: In `internal/config/loader.go`, add `fixInvalidValues` case for MaxContentWidth
  - [x] 1.6: In `internal/config/loader.go`, update `writeDefaultConfig()` with commented `max_content_width`
  - [x] 1.7: In `internal/config/loader.go`, add `l.v.Set("settings.max_content_width", config.MaxContentWidth)` to Save()

- [x] Task 2: Update TUI to use config-based MaxContentWidth (AC: 3, 4, 5)
  - [x] 2.1: In `internal/adapters/tui/views.go:26`, REMOVE `const MaxContentWidth = 120` - will come from config
  - [x] 2.2: In `internal/adapters/tui/model.go`, add `maxContentWidth int` field to Model struct
  - [x] 2.3: In `internal/adapters/tui/model.go`, wire config.MaxContentWidth to model in NewModel() or init
  - [x] 2.4: In `internal/adapters/tui/model.go:316`, convert `isWideWidth(width int)` function to `isWideWidth()` method:
    - Change signature from `func isWideWidth(width int) bool` to `func (m Model) isWideWidth() bool`
    - Update logic to use `m.maxContentWidth` (0 = always return false for unlimited width)
  - [x] 2.5: Update ALL 4 isWideWidth call sites to use method pattern:
    - `model.go:492`: `if isWideWidth(m.width)` → `if m.isWideWidth()`
    - `model.go:610`: `if isWideWidth(m.width)` → `if m.isWideWidth()`
    - `model.go:1354`: `if isWideWidth(m.width)` → `if m.isWideWidth()`
    - `model.go:1388`: `if isWideWidth(m.width)` → `if m.isWideWidth()`
  - [x] 2.6: Update test files that reference MaxContentWidth constant:
    - [x] 2.6.1: `internal/adapters/tui/model_test.go` - Lines 1963-2169 have 15+ references to MaxContentWidth
    - [x] 2.6.2: `internal/adapters/tui/model_responsive_test.go:372` - Update TestIsWideWidth for method signature
    - [x] 2.6.3: Add new tests for max_content_width=0 behavior (full-width mode)

- [x] Task 3: Rebalance column proportions in delegate.go (AC: 1, 2)
  - [x] 3.1: In `internal/adapters/tui/components/delegate.go:19-34`, redefine column width percentages:
    ```go
    // Column width percentages (of available width after fixed columns)
    // Fixed columns: selection(2) + favorite(2) + indicator(3) + time(8) + spacing(5) = 20
    const (
        colNamePct   = 25  // ~25% for project name
        colStagePct  = 40  // ~40% for stage info
        colStatusPct = 20  // ~20% for waiting status
        // Remaining for activity time
    )
    ```
  - [x] 3.2: Update `calculateNameWidth()` at line 132 to use percentage-based calculation:
    ```go
    // Calculate available width after fixed columns
    fixedCols := colSelection + colFavorite + colIndicator + colTime + 5 // 5 for spacing
    availableWidth := d.width - fixedCols
    nameWidth := int(float64(availableWidth) * 0.25)  // 25% of available
    ```
  - [x] 3.3: Update `stageColumnWidth()` to use percentage-based width (not fixed 16)
  - [x] 3.4: Ensure name truncation takes priority over stage truncation (AC2)

- [x] Task 4: Add per-column max widths (AC: 6)
  - [x] 4.1: In `internal/adapters/tui/components/delegate.go`, add max width constants:
    ```go
    const (
        colNameMax   = 40  // Project name max
        colStageMax  = 80  // Stage info max
        colStatusMax = 25  // Waiting status max
    )
    ```
  - [x] 4.2: Apply caps in width calculation functions using `min(calculated, max)`
  - [x] 4.3: Test with 200+ column terminal width

- [x] Task 5: Run tests and lint (AC: 7)
  - [x] 5.1: `make test` - all tests pass
  - [x] 5.2: `make lint` - no warnings
  - [x] 5.3: Update existing tests that assert specific column widths
  - [x] 5.4: Add test for max_content_width=0 behavior (full-width)

## Dev Notes

### Key Code Locations

| File | Action | Details |
|------|--------|---------|
| `internal/core/ports/config.go:13-44` | ADD field | Add `MaxContentWidth int` to Config struct |
| `internal/core/ports/config.go:76-88` | UPDATE | Add default in NewConfig() |
| `internal/core/ports/config.go:179-216` | UPDATE | Add validation in Validate() |
| `internal/config/loader.go:109-117` | UPDATE | Add Viper binding in Save() |
| `internal/config/loader.go:176-249` | UPDATE | Add binding in mapViperToConfig() |
| `internal/config/loader.go:253-325` | UPDATE | Add fix in fixInvalidValues() |
| `internal/config/loader.go:150-173` | UPDATE | Add comment in writeDefaultConfig() |
| `internal/adapters/tui/views.go:25-26` | DELETE | Remove MaxContentWidth constant |
| `internal/adapters/tui/model.go:314-318` | UPDATE | Update isWideWidth() to use config |
| `internal/adapters/tui/model.go:490-534` | UPDATE | Use config-based max width |
| `internal/adapters/tui/components/delegate.go:19-34` | UPDATE | New column proportions |
| `internal/adapters/tui/components/delegate.go:131-153` | UPDATE | Percentage-based calculateNameWidth() |

### Architecture Compliance

**Hexagonal Architecture Requirements:**
- Config change in `internal/core/ports/config.go` (domain layer - zero external deps)
- Viper binding in `internal/config/loader.go` (adapter layer)
- TUI uses config via injected Config struct (not global)

**Pattern to Follow (from Story 8.6, 8.9):**
1. Add field to `ports.Config` struct with comment
2. Set default in `NewConfig()`
3. Add validation in `Validate()`
4. Bind in `loader.go:mapViperToConfig()`
5. Add fix in `loader.go:fixInvalidValues()`
6. Set in `loader.go:Save()`
7. Add commented example in `loader.go:writeDefaultConfig()`

### Implemented Column Constants (delegate.go:19-47)

```go
// Column widths for project row layout
const (
    colSelection = 2  // "> " or "  "
    colFavorite  = 2  // styled star or "  " (Story 3.8)
    colIndicator = 3  // "✨ " or "⚡ " or "   "
    colTime      = 8  // "2w ago" max
    colSpacing   = 5  // spacing between columns

    // Column width percentages (of available width after fixed columns) - Story 8.10
    colNamePct   = 25 // ~25% for project name
    colStagePct  = 40 // ~40% for stage info
    colStatusPct = 20 // ~20% for waiting status

    // Column minimum widths
    colNameMin    = 10 // Minimum name width
    colStageMin   = 10 // Minimum stage width
    colWaitingMin = 19 // Minimum waiting width (fits "[W] WAITING 23h 59m")

    // Column maximum widths - prevent absurd stretching on ultra-wide (Story 8.10 AC6)
    colNameMax    = 40 // Project name max
    colStageMax   = 80 // Stage info max
    colWaitingMax = 25 // Waiting status max
)
```

### Proposed Column Layout (100 char terminal as baseline)

| Column | Current | Proposed | Notes |
|--------|---------|----------|-------|
| Selection | 2 | 2 | Fixed "> " |
| Favorite | 2 | 2 | Fixed star |
| Name | ~35 (dynamic) | ~20-25 (25%) | Cap at 40 |
| Indicator | 3 | 3 | Fixed recency |
| Stage | 16 | ~35-40 (40%) | Cap at 80 |
| Waiting | 14 | ~18-20 (20%) | Cap at 25 |
| Time | 8 | 8 | Fixed |

### isWideWidth() Logic Update

**CRITICAL: This changes from a function to a method - all call sites must be updated!**

**Current (model.go:316):**
```go
func isWideWidth(width int) bool {
    return width > MaxContentWidth  // MaxContentWidth = 120 constant
}

// Called as: isWideWidth(m.width) in 4 places
```

**New (model.go - method using config):**
```go
func (m Model) isWideWidth() bool {
    // max_content_width: 0 means unlimited (always use full width)
    if m.maxContentWidth == 0 {
        return false
    }
    return m.width > m.maxContentWidth
}

// Called as: m.isWideWidth() - NO PARAMETER
```

**Call sites to update (4 locations):**
- `model.go:492`: ProjectsLoadedMsg handler
- `model.go:610`: resizeTickMsg handler
- `model.go:1354`: renderMainContent effectiveWidth calculation
- `model.go:1388`: View() centering logic

### Previous Story Learnings

**From Story 8.9:**
- Config additions need both `ports/config.go` AND `config/loader.go` updates
- Follow the config pattern: NewConfig default -> Validate -> mapViperToConfig -> Save -> writeDefaultConfig
- Default config file includes commented options for discoverability

**From Story 8.6:**
- Config-driven TUI behavior works well
- Model struct can hold config values for quick access
- Validation should allow sensible values and fix invalid ones gracefully

**From Story 8.4:**
- effectiveWidth logic in model.go is the key control point
- Multiple places reference MaxContentWidth (5+ locations in model.go)
- Search thoroughly for all usages when removing the constant

### Emoji Fallback Consideration

**IMPORTANT:** The waiting column must accommodate both emoji and fallback character widths.

| Indicator | Emoji | Fallback | Width Difference |
|-----------|-------|----------|------------------|
| Waiting | `⏸️` (2 display chars) | `[W]` (3 chars) | +1 char |
| Star | `⭐` (1 display char) | `*` (1 char) | 0 |

When calculating `colWaiting` width, ensure it fits the longest case: `"[W] WAITING 23h 59m"` = 19 chars.
Current colWaiting=14 may need increase to 20 for full format with fallback.

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Keep MaxContentWidth as constant | Move to config for user control |
| Hardcode column widths | Use percentage-based calculations |
| Forget to update tests | Search for MaxContentWidth in test files |
| Make columns stretch infinitely | Apply max-width caps for ultra-wide |
| Update only some MaxContentWidth refs | Search and update ALL references |
| Import config directly in views.go | Pass config through model struct |
| Apply percentages to total width | Subtract fixed columns first, then apply percentages to remaining |
| Leave isWideWidth as function | Convert to method and update ALL 4 call sites |

### Testing Strategy

**Unit Tests:**
1. Test `isWideWidth()` with max_content_width=0 (should return false always)
2. Test `isWideWidth()` with max_content_width=120 (default behavior)
3. Test `isWideWidth()` with max_content_width=200 (custom cap)
4. Test column width calculations at 80, 100, 120, 160, 200 column widths
5. Test column max widths are applied at ultra-wide (300+ columns)

**Manual Testing Widths:**
```bash
# Test at various widths (resize terminal or use stty)
./bin/vibe  # Default terminal

# Test max_content_width=0 (unlimited)
# Add to ~/.vibe-dash/config.yaml:
# settings:
#   max_content_width: 0
./bin/vibe  # Should use full terminal width

# Test column proportions
# Verify stage info gets more space than project name
```

### Project Structure Notes

- Changes are entirely in TUI adapter layer (`internal/adapters/tui/`) and config (`internal/config/`, `internal/core/ports/`)
- No database changes
- No CLI command changes
- Backward compatible (default 120 matches current behavior)

### References

- [Source: internal/adapters/tui/views.go:26 - MaxContentWidth constant to remove]
- [Source: internal/adapters/tui/model.go:316 - isWideWidth() function to convert to method]
- [Source: internal/adapters/tui/model.go:492,610,1354,1388 - isWideWidth() call sites to update]
- [Source: internal/adapters/tui/components/delegate.go:19-34 - column constants]
- [Source: internal/adapters/tui/components/delegate.go:132 - calculateNameWidth()]
- [Source: internal/adapters/tui/model_test.go:1963-2169 - Tests referencing MaxContentWidth]
- [Source: internal/adapters/tui/model_responsive_test.go:372 - TestIsWideWidth to update]
- [Source: internal/core/ports/config.go - Config struct pattern]
- [Source: internal/config/loader.go - Viper binding pattern]
- [Source: docs/sprint-artifacts/sprint-change-proposal-2025-12-26.md - Story origin]

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Default Behavior (Backward Compatibility)

```bash
make build && ./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Terminal 120+ cols | Content capped and centered | Content stretches edge to edge |
| Column proportions | Stage info has more space than before | Stage truncated, name has wasted space |

### Step 2: Unlimited Width

```bash
# Add to ~/.vibe-dash/config.yaml:
# settings:
#   max_content_width: 0
./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Terminal 200+ cols | Content uses full width | Still capped at 120 |
| Columns | Proportional, not absurdly stretched | One column takes all space |

### Step 3: Custom Cap

```bash
# Add to ~/.vibe-dash/config.yaml:
# settings:
#   max_content_width: 160
./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Terminal 180 cols | Content capped at 160, centered | Content uses full 180 |
| Terminal 140 cols | Content uses full 140 (no cap applied) | Content capped at 120 |

### Step 4: Column Balance

Visually verify at 100-column terminal:
- Project names: ~20-25 chars before truncation
- Stage info: ~35-40 chars before truncation
- Waiting: Fits "WAITING 2h 30m" without truncation

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Still using constant | Check views.go deletion and model.go wiring |
| max_content_width=0 not working | Check isWideWidth() logic update |
| Columns not rebalanced | Check delegate.go percentage calculations |
| Tests fail | Check for hardcoded MaxContentWidth in tests |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- **Task 1:** Added `MaxContentWidth int` field to Config struct with default 120, validation (>= 0), Viper bindings, and commented example in default config template
- **Task 2:** Removed hardcoded `MaxContentWidth = 120` constant from views.go, added `maxContentWidth int` field to Model struct, converted `isWideWidth()` to method using config value (0 = unlimited), updated all 4 call sites, updated test files (model_test.go, model_responsive_test.go) to use dynamic values from config
- **Task 3:** Implemented percentage-based column calculations with `availableWidth()` helper, `calculateNameWidth()`, `stageColumnWidth()`, and new `waitingColumnWidth()` method. Changed from fixed widths to: name ~25%, stage ~40%, waiting ~20% of available space after fixed columns
- **Task 4:** Added per-column max widths (colNameMax=40, colStageMax=80, colWaitingMax=25) to prevent absurd stretching on ultra-wide terminals. Applied caps in all width calculation functions
- **Task 5:** All tests pass (`make test`), no lint warnings (`make lint`). Updated `TestProjectItemDelegate_StageColumnWidth_Responsive` to test new percentage-based behavior with min/max range validation. Added `TestModel_FullWidthMode_MaxContentWidthZero` for unlimited width mode testing

### File List

- `internal/core/ports/config.go` - Added MaxContentWidth field, default in NewConfig(), validation in Validate()
- `internal/config/loader.go` - Added Viper binding, fixInvalidValues case, updated writeDefaultConfig() template
- `internal/adapters/tui/views.go` - Removed MaxContentWidth constant, added formatMaxWidth() helper for help overlay
- `internal/adapters/tui/views_test.go` - Added TestFormatMaxWidth_Unlimited, TestFormatMaxWidth_WithValue, TestRenderHelpOverlay_ShowsMaxWidth tests
- `internal/adapters/tui/model.go` - Added maxContentWidth field, updated NewModel(), SetConfig(), converted isWideWidth() to method, updated 4 call sites
- `internal/adapters/tui/model_test.go` - Updated 3 tests to use `ports.NewConfig().MaxContentWidth` instead of constant
- `internal/adapters/tui/model_responsive_test.go` - Updated TestIsWideWidth to test method with different maxContentWidth values, added TestModel_FullWidthMode_MaxContentWidthZero
- `internal/adapters/tui/components/delegate.go` - Added percentage constants, min/max width constants, availableWidth() helper, updated calculateNameWidth(), stageColumnWidth(), added waitingColumnWidth(), increased colWaitingMin to 19
- `internal/adapters/tui/components/delegate_test.go` - Updated TestProjectItemDelegate_StageColumnWidth_Responsive, added TestProjectItemDelegate_WaitingColumnWidth_Responsive, TestProjectItemDelegate_WaitingColumnWidth_MaxCap

## Change Log

- 2025-12-26: Story created via correct-course workflow (user feedback during 8.4 review)
- 2025-12-27: Enriched with comprehensive developer context by SM agent
- 2025-12-27: SM validation applied improvements:
  - Added explicit isWideWidth() call site updates (4 locations) to Task 2.5
  - Added specific test file references to Task 2.6 (model_test.go:1963-2169, model_responsive_test.go:372)
  - Added percentage calculation clarification in Task 3.2
  - Added Emoji Fallback Consideration section for waiting column width
  - Updated References with correct line numbers and test file locations
  - Added 2 new anti-patterns to avoid table
- 2025-12-27: Code review completed, fixes applied:
  - **M1 (FIXED):** Added `max_content_width` display to help overlay with formatMaxWidth() helper (shows "unlimited" for 0)
  - **M3 (FIXED):** Increased `colWaitingMin` from 14 to 19 to fit "[W] WAITING 23h 59m" per story Dev Notes requirement
  - **L3 (FIXED):** Added TestProjectItemDelegate_WaitingColumnWidth_Responsive and TestProjectItemDelegate_WaitingColumnWidth_MaxCap tests
  - **L1 (SKIPPED):** Comment cosmetic - not worth changing
  - **L2 (SKIPPED):** Config loading tests already comprehensive
  - All tests pass, lint clean
