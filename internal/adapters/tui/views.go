package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/emoji"
)

// Minimum terminal dimensions
const (
	MinWidth  = 60
	MinHeight = 20
)

// NarrowWarning returns the narrow terminal warning text (Story 3.10 AC2, Story 8.9: emoji fallback).
func NarrowWarning() string {
	return emoji.Warning() + " Narrow terminal - some info hidden"
}

// HeightThresholdTall is the terminal height at which detail panel opens by default (Story 3.10 AC6/AC7).
const HeightThresholdTall = 35

// Story 8.12: Horizontal layout height thresholds
const (
	// MinListHeightHorizontal is minimum lines for project list in horizontal mode
	MinListHeightHorizontal = 10
	// MinDetailHeightHorizontal is minimum lines for detail panel in horizontal mode
	MinDetailHeightHorizontal = 6
	// HorizontalDetailThreshold is the height at which both list and detail fit
	HorizontalDetailThreshold = MinListHeightHorizontal + MinDetailHeightHorizontal // 16
	// HorizontalComfortableThreshold is height at which full 60/40 split is used (code review L1)
	HorizontalComfortableThreshold = 30
)

// renderEmptyView renders the welcome screen when no projects are present.
func renderEmptyView(width, height int) string {
	// Build the content
	title := titleStyle.Render("VIBE DASHBOARD")

	content := strings.Join([]string{
		"",
		"Welcome to Vibe Dashboard! 🎯",
		"",
		"Add your first project:",
		"$ vibe add /path/to/project",
		"",
		"Or from a project directory:",
		"$ cd my-project && vibe add .",
		"",
		hintStyle.Render("Press [?] for help, [q] to quit"),
		"",
	}, "\n")

	// Create the box with title
	box := boxStyle.
		Width(40).
		Render(content)

	// Add title to the border
	lines := strings.Split(box, "\n")
	if len(lines) > 0 {
		// Replace part of the top border with title
		topBorder := lines[0]
		titleWithDash := fmt.Sprintf("\u2500 %s ", title)

		// Find where to insert the title (after the corner)
		if len(topBorder) > 3 {
			lines[0] = string(topBorder[0]) + titleWithDash + topBorder[len(titleWithDash)+1:]
		}
		box = strings.Join(lines, "\n")
	}

	// Center the box in the terminal
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// renderHelpOverlay renders the help screen with keyboard shortcuts.
// Story 8.7: cfg parameter displays current config values (nil-safe - uses defaults).
func renderHelpOverlay(width, height int, cfg *ports.Config) string {
	title := titleStyle.Render("KEYBOARD SHORTCUTS")

	// Story 8.7: Use defaults if config is nil (AC5: no crash on missing config)
	if cfg == nil {
		cfg = ports.NewConfig()
	}

	// Unicode arrows: \u2193 renders as ↓, \u2191 renders as ↑
	content := strings.Join([]string{
		"",
		"Navigation",
		">        Selection indicator (focused)",
		"j/\u2193     Move down",
		"k/\u2191     Move up",
		"",
		"Actions",
		"d        Toggle detail panel",
		"f        Toggle favorite",
		"n        Edit notes",
		"x        Remove project",
		"H        Hibernate/Activate",
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
		formatMaxWidth(cfg.MaxContentWidth),
		"",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		hintStyle.Render("Config: ~/.vibe-dash/config.yaml"),
		hintStyle.Render("Per-project: ~/.vibe-dash/<project>/config.yaml"),
		"",
		hintStyle.Render("Press any key to close"),
		"",
	}, "\n")

	box := boxStyle.
		Width(52). // Longest line (44) + padding (4) + border (2) + buffer (2)
		Render(content)

	// Add title to the border
	lines := strings.Split(box, "\n")
	if len(lines) > 0 {
		topBorder := lines[0]
		titleWithDash := fmt.Sprintf("\u2500 %s ", title)

		if len(topBorder) > 3 {
			lines[0] = string(topBorder[0]) + titleWithDash + topBorder[len(titleWithDash)+1:]
		}
		box = strings.Join(lines, "\n")
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// renderTooSmallView renders a message when terminal is too small.
func renderTooSmallView(width, height int) string {
	msg := fmt.Sprintf("Terminal too small. Minimum %dx%d required.\nCurrent: %dx%d",
		MinWidth, MinHeight, width, height)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, msg)
}

// renderNoteEditor renders the inline note editor dialog (Story 3.7).
// Follows renderHelpOverlay pattern for dialog styling and centering.
func renderNoteEditor(projectName string, input textinput.Model, width, height int) string {
	// Dialog dimensions - ensure minimum width of 30, cap at 60
	dialogWidth := width - 4
	if dialogWidth < 30 {
		dialogWidth = 30
	}
	if dialogWidth > 60 {
		dialogWidth = 60
	}

	// Title
	title := titleStyle.Render(fmt.Sprintf("Edit note for \"%s\"", projectName))

	// Input line with > prefix
	inputLine := "> " + input.View()

	// Instructions
	instructions := hintStyle.Render("[Enter] save  [Esc] cancel")

	// Content
	content := strings.Join([]string{
		"",
		inputLine,
		"",
		instructions,
		"",
	}, "\n")

	// Dialog box style (same as help overlay)
	box := boxStyle.
		Width(dialogWidth).
		Render(content)

	// Add title to the border (same pattern as renderHelpOverlay)
	lines := strings.Split(box, "\n")
	if len(lines) > 0 {
		topBorder := lines[0]
		titleWithDash := fmt.Sprintf("\u2500 %s ", title)

		if len(topBorder) > 3 {
			lines[0] = string(topBorder[0]) + titleWithDash + topBorder[len(titleWithDash)+1:]
		}
		box = strings.Join(lines, "\n")
	}

	// Center in terminal
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// renderConfirmRemoveDialog renders the inline remove confirmation dialog (Story 3.9).
// Follows renderNoteEditor pattern for dialog styling and centering.
func renderConfirmRemoveDialog(projectName string, width, height int) string {
	// Dialog dimensions - cap width at 60
	dialogWidth := width - 4
	if dialogWidth < 30 {
		dialogWidth = 30
	}
	if dialogWidth > 60 {
		dialogWidth = 60
	}

	// Title
	title := titleStyle.Render("Confirm Removal")

	// Content
	content := strings.Join([]string{
		"",
		fmt.Sprintf("Remove '%s' from tracking?", projectName),
		"",
		hintStyle.Render("[y] Yes  [n] No  [Esc] Cancel"),
		"",
	}, "\n")

	// Dialog box style (same as help overlay)
	box := boxStyle.
		Width(dialogWidth).
		Render(content)

	// Add title to the border (same pattern as renderHelpOverlay)
	lines := strings.Split(box, "\n")
	if len(lines) > 0 {
		topBorder := lines[0]
		titleWithDash := fmt.Sprintf("\u2500 %s ", title)

		if len(topBorder) > 3 {
			lines[0] = string(topBorder[0]) + titleWithDash + topBorder[len(titleWithDash)+1:]
		}
		box = strings.Join(lines, "\n")
	}

	// Center in terminal
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// formatMaxWidth formats the max_content_width setting for display.
// Story 8.10: Shows "unlimited" for 0, otherwise shows the value.
func formatMaxWidth(maxWidth int) string {
	if maxWidth == 0 {
		return "Max Width:   unlimited"
	}
	return fmt.Sprintf("Max Width:   %d", maxWidth)
}

// renderNarrowWarning renders the narrow terminal warning bar (Story 3.10 AC2).
func renderNarrowWarning(width int) string {
	// Use yellow (ANSI 3) for warning text, consistent with WarningStyle
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")). // Yellow (ANSI 3)
		Bold(true)

	warning := warningStyle.Render(NarrowWarning())
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, warning)
}

// renderHibernatedEmptyView renders the empty state for hibernated projects view (Story 11.4 AC6).
func renderHibernatedEmptyView(width, height int) string {
	content := strings.Join([]string{
		"",
		"No hibernated projects.",
		"",
		hintStyle.Render("Projects auto-hibernate after inactivity"),
		hintStyle.Render("Press [h] to return to active view"),
		"",
	}, "\n")

	box := boxStyle.
		Width(40).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// renderHibernatedView renders the hibernated projects list (Story 11.4).
// Reuses the same ProjectListModel component as active view but with hibernated data.
func renderHibernatedView(hibernatedList *components.ProjectListModel, showDetail bool, detailPanel *components.DetailPanelModel, width, height, maxContentWidth int) string {
	// Calculate effective width (cap at maxContentWidth - Story 8.10)
	effectiveWidth := width
	if maxContentWidth > 0 && width > maxContentWidth {
		effectiveWidth = maxContentWidth
	}

	// Update list dimensions
	hibernatedList.SetSize(effectiveWidth, height)

	// Render project list
	listView := hibernatedList.View()

	// If detail panel is visible, render it below list (horizontal layout only for hibernated view)
	if showDetail {
		detailPanel.SetProject(hibernatedList.SelectedProject())
		detailPanel.SetVisible(true)
		detailPanel.SetHorizontalMode(false)
		detailPanel.SetSize(effectiveWidth, 0)
		detailView := detailPanel.View()

		// Count detail lines
		detailLines := strings.Count(detailView, "\n") + 1
		listHeight := height - detailLines - 1
		if listHeight > 0 {
			hibernatedList.SetSize(effectiveWidth, listHeight)
			listView = hibernatedList.View()
			listView += "\n" + detailView
		}
	}

	// Center if width is capped
	if maxContentWidth > 0 && width > maxContentWidth {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top, listView)
	}
	return listView
}
