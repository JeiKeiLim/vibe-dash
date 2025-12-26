package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/JeiKeiLim/vibe-dash/internal/shared/styles"
)

// UseColor re-exports the color setting from shared styles.
// Determines if colors should be used based on NO_COLOR and TERM env vars.
var UseColor = styles.UseColor

// ============================================================================
// Base Styles (re-exported from shared package)
// ============================================================================

var (
	// boxStyle is used for bordered containers with rounded corners.
	boxStyle = styles.BoxStyle

	// titleStyle is used for headings.
	titleStyle = styles.TitleStyle

	// hintStyle is used for dimmed help text.
	hintStyle = styles.HintStyle
)

// ============================================================================
// Dashboard Component Styles (re-exported from shared package)
// Uses 16-color ANSI palette for maximum terminal compatibility.
// ============================================================================

// SelectedStyle is used for the currently selected row in lists.
// Uses cyan background for visibility on both dark and light themes.
var SelectedStyle = styles.SelectedStyle

// WaitingStyle is used ONLY for the WAITING indicator (killer feature).
// Bold red to catch peripheral vision. Reserved exclusively for agent waiting state.
var WaitingStyle = styles.WaitingStyle

// RecentStyle is used for today indicator (within 24 hours).
var RecentStyle = styles.RecentStyle

// ActiveStyle is used for this week indicator (within 7 days).
var ActiveStyle = styles.ActiveStyle

// UncertainStyle is used for uncertain detection state.
var UncertainStyle = styles.UncertainStyle

// FavoriteStyle is used for favorite/starred project indicator.
var FavoriteStyle = styles.FavoriteStyle

// WarningStyle is used for warning indicators (e.g., missing path).
var WarningStyle = styles.WarningStyle

// DimStyle is used for hints, secondary info, and less important text.
// Note: Functionally similar to hintStyle but with different semantic purpose.
// hintStyle is for help text overlays, DimStyle is for general dimming.
var DimStyle = styles.DimStyle

// BorderStyle is used for panel boundaries with square corners.
// Uses ANSI color 8 (bright black/gray) for 16-color palette compatibility.
var BorderStyle = styles.BorderStyle

// ============================================================================
// Style Helper Functions
// ============================================================================

// ApplySelected wraps text with selection highlighting.
func ApplySelected(text string) string {
	return SelectedStyle.Render(text)
}

// ApplyIndicator applies the appropriate style to text based on indicator type.
// Supported types: "waiting", "recent", "active", "uncertain", "favorite", "dim", "warning".
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
	case "dim":
		return DimStyle.Render(text)
	case "warning":
		return WarningStyle.Render(text)
	default:
		return text
	}
}

func init() {
	if !UseColor {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}
