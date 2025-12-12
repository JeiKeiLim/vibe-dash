package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Minimum terminal dimensions
const (
	MinWidth  = 60
	MinHeight = 20
)

// UseColor determines if colors should be used based on NO_COLOR and TERM env vars.
var UseColor = os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"

// Styles for the TUI
var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")) // Cyan

	hintStyle = lipgloss.NewStyle().
			Faint(true)
)

func init() {
	if !UseColor {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

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

	content := strings.Join([]string{
		"",
		"General",
		"?        Toggle this help",
		"q        Quit",
		"Ctrl+C   Force quit",
		"",
		hintStyle.Render("Press any key to close"),
		"",
	}, "\n")

	box := boxStyle.
		Width(40).
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
