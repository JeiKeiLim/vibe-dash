package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Minimum terminal dimensions
const (
	MinWidth  = 60
	MinHeight = 20
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
func renderHelpOverlay(width, height int) string {
	title := titleStyle.Render("KEYBOARD SHORTCUTS")

	// Unicode arrows: \u2193 renders as ↓, \u2191 renders as ↑
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
		hintStyle.Render("Press any key to close"),
		"",
	}, "\n")

	box := boxStyle.
		Width(46). // Longest line (32) + padding (4) + border (2) + buffer (8)
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
