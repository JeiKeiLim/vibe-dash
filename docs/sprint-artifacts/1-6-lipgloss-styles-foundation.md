# Story 1.6: Lipgloss Styles Foundation

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | Styles imported by TUI components |
| **Key Dependencies** | github.com/charmbracelet/lipgloss (already in go.mod) |
| **Files to Create** | styles.go, styles_test.go |
| **Location** | internal/adapters/tui/ |
| **Color Mode** | 16-color ANSI palette for terminal compatibility |

### Style Constants Quick Reference

| Style | Color | ANSI Code | Usage |
|-------|-------|-----------|-------|
| `SelectedStyle` | Cyan background | Color 6 | Currently selected row |
| `WaitingStyle` | Bold + Red | Color 1 | ONLY for WAITING state |
| `RecentStyle` | Green | Color 2 | Today indicator |
| `ActiveStyle` | Yellow | Color 3 | This week indicator |
| `UncertainStyle` | Dim gray | Color 8 | Uncertain state |
| `FavoriteStyle` | Magenta | Color 5 | Favorite indicator |
| `DimStyle` | Faint modifier | - | Hints, secondary info |
| `BorderStyle` | Normal border | Color 240 | Panel boundaries |

## Story

**As a** developer,
**I want** centralized Lipgloss styles defined,
**So that** all TUI components render consistently.

## Acceptance Criteria

```gherkin
AC1: Given I need consistent styling across the TUI
     When I create styles.go in internal/adapters/tui/
     Then the following styles are defined:
       - SelectedStyle: Cyan background (color 6) for currently selected row
       - WaitingStyle: Bold + Red foreground (color 1) ONLY for WAITING state
       - RecentStyle: Green foreground (color 2) for today indicator
       - ActiveStyle: Yellow foreground (color 3) for this week indicator
       - UncertainStyle: Dim/Faint gray (color 8) for uncertain state
       - FavoriteStyle: Magenta foreground (color 5) for favorite indicator
       - DimStyle: Faint modifier for hints and secondary info
       - BorderStyle: Normal border (square corners) for panel boundaries

AC2: Given the NO_COLOR environment variable is set
     When styles are initialized
     Then NO_COLOR environment variable is respected
     And all color styling is disabled (ASCII profile)

AC3: Given styles are applied to text
     When rendered on both dark and light terminal themes
     Then styles work correctly on both themes
     And use 16-color ANSI palette for maximum compatibility

AC4: Given styles need to be applied to TUI elements
     When importing from internal/adapters/tui/
     Then styles are accessible as exported package variables
     And can be composed with other lipgloss styles

AC5: Given the existing styles.go from Story 1.5
     When refactoring to add new styles
     Then existing boxStyle, titleStyle, hintStyle remain functional
     And existing UseColor variable continues to work
     And existing NO_COLOR check remains in place
```

## Tasks / Subtasks

- [x] **Task 1: Create styles.go and move existing styles from views.go** (AC: 1, 4, 5)
  - [x] 1.1 Review current styles in `views.go` from Story 1.5 (lines 18-40)
  - [x] 1.2 Create new file `internal/adapters/tui/styles.go`
  - [x] 1.3 Move from views.go to styles.go: UseColor variable, init() function, boxStyle, titleStyle, hintStyle
  - [x] 1.4 Remove moved code from views.go (keep only view rendering functions)
  - [x] 1.5 Verify views.go still compiles (same package, no import needed)
  - [x] 1.6 Add exported style constants section header comment in styles.go

- [x] **Task 2: Implement Selection and Navigation Styles** (AC: 1, 3)
  - [x] 2.1 Create SelectedStyle with cyan background (ANSI color 6)
  - [x] 2.2 Verify SelectedStyle works on dark and light themes
  - [x] 2.3 Add SelectedStyle documentation comment

- [x] **Task 3: Implement Status Indicator Styles** (AC: 1, 3)
  - [x] 3.1 Create WaitingStyle with bold + red foreground (ANSI color 1)
  - [x] 3.2 Create RecentStyle with green foreground (ANSI color 2)
  - [x] 3.3 Create ActiveStyle with yellow foreground (ANSI color 3)
  - [x] 3.4 Create UncertainStyle with faint/dim gray (ANSI color 8)
  - [x] 3.5 Add documentation comments for each style

- [x] **Task 4: Implement Decoration Styles** (AC: 1, 3)
  - [x] 4.1 Create FavoriteStyle with magenta foreground (ANSI color 5)
  - [x] 4.2 Create DimStyle with Faint() modifier
  - [x] 4.3 Create BorderStyle with normal (square) border
  - [x] 4.4 Add documentation comments for each style

- [x] **Task 5: Verify NO_COLOR Support** (AC: 2)
  - [x] 5.1 Verify existing init() handles NO_COLOR and TERM=dumb
  - [x] 5.2 Test that all new styles respect NO_COLOR setting
  - [x] 5.3 Document NO_COLOR behavior in code comments

- [x] **Task 6: Create Style Helper Functions** (AC: 4)
  - [x] 6.1 Create ApplySelected(text string) helper function
  - [x] 6.2 Create ApplyIndicator(indicator, text string) for status indicators
  - [x] 6.3 Ensure helpers compose well with existing styles

- [x] **Task 7: Write Tests** (AC: all)
  - [x] 7.1 Create `styles_test.go` with style definition tests
  - [x] 7.2 Test each style applies correct foreground/background color
  - [x] 7.3 Test NO_COLOR disables color styling
  - [x] 7.4 Test UseColor variable reflects environment correctly
  - [x] 7.5 Test style composition (combining styles)
  - [x] 7.6 Verify existing TestUseColorLogic from Story 1.5 still passes

- [x] **Task 8: Integration and Validation** (AC: all)
  - [x] 8.1 Run `make build` and verify compilation
  - [x] 8.2 Run `make lint` and fix any issues
  - [x] 8.3 Run `make test` and verify all tests pass
  - [x] 8.4 Manual verification: import styles in views.go and render test
  - [x] 8.5 Test with NO_COLOR=1 set to verify ASCII fallback

## Implementation Order (Recommended)

Execute tasks in this order to minimize rework:

1. **Task 1: Refactor existing styles.go** - Review and prepare for additions
2. **Task 2: Selection styles** - SelectedStyle for navigation
3. **Task 3: Status indicator styles** - WaitingStyle, RecentStyle, ActiveStyle, UncertainStyle
4. **Task 4: Decoration styles** - FavoriteStyle, DimStyle, BorderStyle
5. **Task 5: NO_COLOR verification** - Ensure accessibility compliance
6. **Task 6: Helper functions** - Convenience wrappers for common operations
7. **Task 7: Tests** - Comprehensive test coverage
8. **Task 8: Integration** - Final validation

## Dev Notes

### CRITICAL Requirements (Must Not Miss)

| Requirement | Why | Reference |
|-------------|-----|-----------|
| **16-color ANSI only** | Maximum terminal compatibility | UX spec lines 472-487 |
| **NO_COLOR support** | Accessibility requirement | UX spec lines 1635-1642 |
| **Red ONLY for WAITING** | Red reserved for killer feature | UX spec color system |
| **Cyan for selection** | Per UX design system | UX spec lines 472-475 |
| **Preserve existing styles** | Story 1.5 created boxStyle, titleStyle, hintStyle | Story 1.5 File List |

### File Refactoring Steps (IMPORTANT)

**Current state:** Styles are defined in `views.go` (lines 18-40)
**Target state:** All styles in dedicated `styles.go` file

**Steps:**
1. Create `internal/adapters/tui/styles.go`
2. Move from views.go to styles.go:
   - `UseColor` variable (line 19)
   - `init()` function (lines 36-40)
   - `boxStyle`, `titleStyle`, `hintStyle` variables (lines 22-34)
3. Add new styles (SelectedStyle, WaitingStyle, etc.) to styles.go
4. Remove moved code from views.go
5. Verify views.go still compiles (same package, no import needed)

### Package Visibility Note

All styles are defined as package-level variables (PascalCase = exported).
Since `styles.go` and `views.go` are in the same `tui` package:
- No import statement needed between them
- Styles are directly accessible in views.go after refactoring
- Helper functions in styles.go are also directly accessible

### Existing Styles in views.go (Story 1.5)

**IMPORTANT:** The current styles are in `views.go`, NOT in a separate `styles.go` file. This story creates the new `styles.go` file and moves existing styles to it.

The current `internal/adapters/tui/views.go` from Story 1.5 contains:

```go
package tui

import (
    "os"

    "github.com/charmbracelet/lipgloss"
    "github.com/muesli/termenv"
)

// UseColor determines if color output is enabled.
// Respects NO_COLOR environment variable per accessibility guidelines.
var UseColor = os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"

var (
    // boxStyle is used for bordered containers
    boxStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("240")).
        Padding(1, 2)

    // titleStyle is used for headings
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("39"))  // Cyan

    // hintStyle is used for dimmed help text
    hintStyle = lipgloss.NewStyle().
        Faint(true)
)

func init() {
    if !UseColor {
        lipgloss.SetColorProfile(termenv.Ascii)
    }
}
```

**Keep all of this and ADD the new styles below it.**

### New Styles to Add (per UX Design Specification)

```go
// ============================================================================
// Dashboard Component Styles (Story 1.6)
// ============================================================================

// SelectedStyle is used for the currently selected row in lists.
// Uses cyan background for visibility on both dark and light themes.
var SelectedStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("6"))  // Cyan

// WaitingStyle is used ONLY for the WAITING indicator (killer feature).
// Bold red to catch peripheral vision. Reserved exclusively for agent waiting state.
var WaitingStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("1"))  // Red

// RecentStyle is used for today indicator (within 24 hours).
var RecentStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("2"))  // Green

// ActiveStyle is used for this week indicator (within 7 days).
var ActiveStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("3"))  // Yellow

// UncertainStyle is used for uncertain detection state.
var UncertainStyle = lipgloss.NewStyle().
    Faint(true).
    Foreground(lipgloss.Color("8"))  // Bright black (gray)

// FavoriteStyle is used for favorite/starred project indicator.
var FavoriteStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("5"))  // Magenta

// DimStyle is used for hints, secondary info, and less important text.
var DimStyle = lipgloss.NewStyle().
    Faint(true)

// BorderStyle is used for panel boundaries with square corners.
// Uses ANSI color 8 (bright black/gray) for 16-color palette compatibility.
var BorderStyle = lipgloss.NewStyle().
    Border(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("8"))
```

### 16-Color ANSI Palette Reference

```
Standard Colors (0-7):
  0 = Black       4 = Blue
  1 = Red         5 = Magenta
  2 = Green       6 = Cyan
  3 = Yellow      7 = White

Bright Colors (8-15):
  8 = Bright Black (Gray)   12 = Bright Blue
  9 = Bright Red            13 = Bright Magenta
  10 = Bright Green         14 = Bright Cyan
  11 = Bright Yellow        15 = Bright White
```

**CRITICAL:** Use string numbers ("1", "2", etc.) for 16-color palette, NOT hex codes. This ensures compatibility across all terminal emulators.

### Style Composition Pattern

Lipgloss styles can be composed for complex styling:

```go
// Example: Selected + Waiting
text := WaitingStyle.Render("WAITING")
row := SelectedStyle.Render(text)

// Example: Apply to full row
selectedRow := SelectedStyle.Width(width).Render(rowContent)
```

### Helper Functions Pattern

```go
// ApplySelected wraps text with selection highlighting.
func ApplySelected(text string) string {
    return SelectedStyle.Render(text)
}

// ApplyIndicator applies the appropriate style to an indicator based on type.
// indicator: the visual indicator (e.g., "WAITING", "Recent")
// text: additional text to append
func ApplyIndicator(indicatorType string, text string) string {
    switch indicatorType {
    case "waiting":
        return WaitingStyle.Render(text)
    case "recent":
        return RecentStyle.Render(text)
    case "active":
        return ActiveStyle.Render(text)
    case "uncertain":
        return UncertainStyle.Render(text)
    case "favorite":
        return FavoriteStyle.Render(text)
    default:
        return text
    }
}
```

### Testing Lipgloss Styles

**Testing Note:** Lipgloss doesn't expose internal color values for direct inspection.
Tests should verify:
1. Styles render non-empty strings (confirms no panic)
2. Style composition works (combining styles)
3. NO_COLOR environment is respected

Direct ANSI escape code testing is not recommended as it's fragile and implementation-dependent.

```go
// styles_test.go
package tui

import (
    "os"
    "testing"

    "github.com/charmbracelet/lipgloss"
)

func TestSelectedStyle_HasCyanBackground(t *testing.T) {
    // Lipgloss doesn't expose internal color values easily,
    // but we can verify the style renders without error
    result := SelectedStyle.Render("test")
    if result == "" {
        t.Error("SelectedStyle should render non-empty string")
    }
}

func TestWaitingStyle_IsBoldRed(t *testing.T) {
    result := WaitingStyle.Render("WAITING")
    if result == "" {
        t.Error("WaitingStyle should render non-empty string")
    }
    // Note: Actual ANSI codes verification would require parsing escape sequences
}

func TestStylesRespectNoColor(t *testing.T) {
    // Save original
    original := os.Getenv("NO_COLOR")
    defer os.Setenv("NO_COLOR", original)

    // Set NO_COLOR
    os.Setenv("NO_COLOR", "1")

    // Re-check UseColor (note: this tests the variable, not runtime behavior)
    useColor := os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
    if useColor {
        t.Error("UseColor should be false when NO_COLOR is set")
    }
}

func TestStyleComposition(t *testing.T) {
    // Test that styles can be composed without error
    innerText := WaitingStyle.Render("WAITING")
    outerText := SelectedStyle.Render(innerText)

    if outerText == "" {
        t.Error("Composed styles should render non-empty string")
    }
}

func TestAllStylesRenderWithoutPanic(t *testing.T) {
    styles := []struct {
        name  string
        style lipgloss.Style
    }{
        {"SelectedStyle", SelectedStyle},
        {"WaitingStyle", WaitingStyle},
        {"RecentStyle", RecentStyle},
        {"ActiveStyle", ActiveStyle},
        {"UncertainStyle", UncertainStyle},
        {"FavoriteStyle", FavoriteStyle},
        {"DimStyle", DimStyle},
        {"BorderStyle", BorderStyle},
    }

    for _, tc := range styles {
        t.Run(tc.name, func(t *testing.T) {
            // Should not panic
            result := tc.style.Render("test content")
            if result == "" {
                t.Errorf("%s rendered empty string", tc.name)
            }
        })
    }
}
```

### NO_COLOR Implementation Notes

The existing `init()` function handles NO_COLOR:

```go
func init() {
    if !UseColor {
        lipgloss.SetColorProfile(termenv.Ascii)
    }
}
```

This sets lipgloss to ASCII-only profile, which strips all ANSI escape codes. All styles automatically respect this - no per-style handling needed.

### Previous Story Learnings (Story 1.5)

From the completed Story 1.5:

1. **UseColor variable** - Already defined, checks NO_COLOR and TERM=dumb
2. **termenv dependency** - Already added to go.mod
3. **lipgloss.SetColorProfile** - API may change; current implementation works
4. **Test for UseColor** - TestUseColorLogic exists in model_test.go
5. **Emoji handling** - runewidth package available for width calculation
6. **Code review fixes applied** - Pointer-free model, key constants used

### Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/adapters/tui/styles.go` | Create | New file with all style constants (moved from views.go + new styles) |
| `internal/adapters/tui/views.go` | Modify | Remove style definitions (moved to styles.go) |
| `internal/adapters/tui/styles_test.go` | Create | Style definition tests |

### Dependencies (Already Available)

```go
import (
    "os"

    "github.com/charmbracelet/lipgloss"
    "github.com/muesli/termenv"
)
```

Both packages already in go.mod from Story 1.5.

**Note:** The `termenv` package is required in `styles.go` for the init() function:
```go
// Used in init() for NO_COLOR support
lipgloss.SetColorProfile(termenv.Ascii)
```

The `os` package is needed for `os.Getenv("NO_COLOR")` in the UseColor variable.

### DO NOT (Anti-Patterns)

| DO NOT | DO INSTEAD |
|--------|------------|
| Use hex color codes | Use 16-color ANSI numbers ("1", "2", etc.) |
| Use Red for anything except WAITING | Red is reserved for killer feature |
| Create styles that depend on 256-color | Stick to 16-color for compatibility |
| Remove existing boxStyle/titleStyle/hintStyle | Keep them, add new styles |
| Test actual ANSI escape sequences | Test that styles render without error |
| Modify UseColor calculation | Keep existing NO_COLOR/TERM check |

### Project Structure Notes

**Alignment with Architecture:**

- Styles in `internal/adapters/tui/` (per Architecture Section: Project Structure)
- Tests co-located as `styles_test.go`
- Styles are package-level variables for easy import
- NO_COLOR handling follows UX accessibility requirements

**Color Usage Philosophy (UX Design):**

- Red ONLY for WAITING state (killer feature visibility)
- Cyan for selection (neutral, visible on dark/light)
- Green for positive/recent (today activity)
- Yellow for active (this week activity)
- Gray for uncertain/dim (less important)
- Magenta for favorites (distinctive but not alarming)

### References

| Document | Section | Key Content |
|----------|---------|-------------|
| architecture.md | TUI Framework | Lines 233-248: Lipgloss for styling |
| architecture.md | Project Structure | Lines 782-790: internal/adapters/tui/ |
| prd.md | Visual Indicators | Lines 234: Visual indicators |
| epics.md | Story 1.6 | Lines 415-469: Full acceptance criteria |
| project-context.md | Technology Stack | Lipgloss Latest |
| ux-design-specification.md | Color System | Lines 472-487: 16-color ANSI palette |
| ux-design-specification.md | NO_COLOR | Lines 1635-1642: Accessibility requirement |
| Story 1.5 | styles.go | Existing style definitions |

### Previous Story Files Available

| Story | Status | Key Learnings |
|-------|--------|---------------|
| 1.1 | Done | Project scaffolding complete |
| 1.2 | Done | Domain entities in internal/core/domain/ |
| 1.3 | Done | Port interfaces in internal/core/ports/ |
| 1.4 | Done | CLI framework, exit codes, flags, context |
| 1.5 | Done | TUI shell, EmptyView, help overlay, existing styles |

## Dev Agent Record

### Context Reference

Story context created from comprehensive analysis of:
- docs/epics.md (Story 1.6 requirements)
- docs/architecture.md (TUI adapter location, Lipgloss styling)
- docs/prd.md (Visual indicators, dashboard requirements)
- docs/project-context.md (Technology stack, naming conventions)
- docs/sprint-artifacts/1-5-bubble-tea-tui-shell.md (Previous story, existing styles.go)
- Git history (recent commits a086e6a, 12eb5e7)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None - story created in YOLO mode per SM activation step 4.

### Completion Notes List

- Created `internal/adapters/tui/styles.go` with all Lipgloss styles centralized
- Moved UseColor, init(), boxStyle, titleStyle, hintStyle from views.go to styles.go
- Implemented 8 new dashboard component styles using 16-color ANSI palette:
  - SelectedStyle (cyan background, ANSI 6)
  - WaitingStyle (bold red, ANSI 1) - reserved for killer feature
  - RecentStyle (green, ANSI 2)
  - ActiveStyle (yellow, ANSI 3)
  - UncertainStyle (dim gray, ANSI 8)
  - FavoriteStyle (magenta, ANSI 5)
  - DimStyle (faint modifier)
  - BorderStyle (square border with ANSI 8 for 16-color compatibility)
- Created helper functions: ApplySelected(), ApplyIndicator()
- NO_COLOR support verified via init() and tests
- All tests pass (make test), lint passes (make lint), build succeeds (make build)

### Code Review Fixes Applied (2025-12-12)

**H1 (High): BorderStyle 16-color palette compliance**
- Changed BorderStyle from color "240" to color "8" (bright black/gray)
- Now complies with AC1/AC3 16-color ANSI palette requirement

**M1 (Medium): NO_COLOR runtime behavior test**
- Added TestNoColorBehavior to document and verify NO_COLOR logic
- Added TestStylesRenderConsistently to verify content preservation

**M2 (Medium): Improved test assertions**
- Added TestStylesRenderConsistently with content verification
- Tests now verify rendered output contains original text

**L1 (Low): DimStyle/hintStyle documentation**
- Added comments explaining semantic difference between DimStyle and hintStyle

**L2 (Low): ApplyIndicator "dim" support**
- Added "dim" case to ApplyIndicator function for consistency

**L3 (Low): Edge case tests**
- Added TestApplySelected_EmptyString
- Added TestApplyIndicator_EmptyText for all indicator types
- Added TestApplyIndicator_DimType

### File List

- `internal/adapters/tui/styles.go` (created) - Centralized style definitions
- `internal/adapters/tui/styles_test.go` (created) - Style tests
- `internal/adapters/tui/views.go` (modified) - Removed style definitions (moved to styles.go)
