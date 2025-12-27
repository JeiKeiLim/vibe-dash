package styles

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// UseColor determines if colors should be used based on NO_COLOR and TERM env vars.
// Respects NO_COLOR environment variable per accessibility guidelines.
// Note: Color profile initialization (lipgloss.SetColorProfile) remains in tui/styles.go
// near Bubble Tea initialization.
var UseColor = os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"

// =============================================================================
// Color Reference (16-color ANSI palette for maximum terminal compatibility)
// =============================================================================
//
// | ANSI Color | Value | Usage                                    |
// |------------|-------|------------------------------------------|
// | 1          | Red   | WaitingStyle (bold)                      |
// | 2          | Green | RecentStyle                              |
// | 3          | Yellow| ActiveStyle, WarningStyle (bold)         |
// | 5          | Magenta| FavoriteStyle                           |
// | 6          | Cyan  | SelectedStyle (background)               |
// | 8          | Bright Black | UncertainStyle, BorderStyle       |
// | 39         | Cyan  | TitleStyle (foreground, bold)            |
// | 240        | Gray  | BoxStyle border                          |
// =============================================================================

// =============================================================================
// Base Styles (from Story 1.5)
// =============================================================================

// BoxStyle is used for bordered containers with rounded corners.
// Uses gray (240) for the border color.
var BoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(1, 2)

// TitleStyle is used for headings with bold cyan text.
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39")) // Cyan

// HintStyle is used for dimmed help text.
var HintStyle = lipgloss.NewStyle().
	Faint(true)

// =============================================================================
// Dashboard Component Styles (Story 1.6)
// Uses 16-color ANSI palette for maximum terminal compatibility.
// =============================================================================

// SelectedStyle is used for the currently selected row in lists.
// Uses cyan background for visibility on both dark and light themes.
var SelectedStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("6")) // Cyan

// WaitingStyle is used ONLY for the WAITING indicator (killer feature).
// Bold red to catch peripheral vision. Reserved exclusively for agent waiting state.
var WaitingStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("1")) // Red

// RecentStyle is used for today indicator (within 24 hours).
var RecentStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("2")) // Green

// ActiveStyle is used for this week indicator (within 7 days).
var ActiveStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("3")) // Yellow

// UncertainStyle is used for uncertain detection state.
var UncertainStyle = lipgloss.NewStyle().
	Faint(true).
	Foreground(lipgloss.Color("8")) // Bright black (gray)

// FavoriteStyle is used for favorite/starred project indicator.
var FavoriteStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("5")) // Magenta

// WarningStyle is used for warning indicators (e.g., missing path).
var WarningStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("3")) // Yellow

// DimStyle is used for hints, secondary info, and less important text.
// Note: Functionally similar to HintStyle but with different semantic purpose.
// HintStyle is for help text overlays, DimStyle is for general dimming.
var DimStyle = lipgloss.NewStyle().
	Faint(true)

// BorderStyle is used for panel boundaries with square corners.
// Uses ANSI color 8 (bright black/gray) for 16-color palette compatibility.
var BorderStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("8"))

// HorizontalBorderStyle removes top border for horizontal layout stacking (Story 8.12).
// Saves 1 vertical line when detail panel is below project list.
var HorizontalBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("8")).
	BorderTop(false).
	BorderLeft(true).
	BorderRight(true).
	BorderBottom(true)
