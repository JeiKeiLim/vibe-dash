package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Story 16.3: renderStatsView renders the full-screen Stats View.
// Layout: Header with "STATS" title and [ESC] hint, centered placeholder content.
func (m Model) renderStatsView() string {
	// Calculate effective width (Story 8.10: respect max content width)
	effectiveWidth := m.width
	if m.isWideWidth() {
		effectiveWidth = m.maxContentWidth
	}

	// Header: "STATS" title on left, "[ESC] Back to Dashboard" hint on right
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Render("STATS")
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[ESC] Back to Dashboard")

	// Calculate spacing between title and hint
	titleWidth := lipgloss.Width(title)
	hintWidth := lipgloss.Width(hint)
	spacing := effectiveWidth - titleWidth - hintWidth
	if spacing < 1 {
		spacing = 1
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		title,
		strings.Repeat(" ", spacing),
		hint,
	)

	// Content height (account for header and status bar)
	contentHeight := m.height - statusBarHeight(m.height) - 2 // -2 for header padding
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Centered placeholder content (AC4: no metrics.db access in this story)
	content := lipgloss.NewStyle().
		Width(effectiveWidth).
		Height(contentHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("241")).
		Render("Project metrics will appear here")

	// Combine header and content
	statsContent := lipgloss.JoinVertical(lipgloss.Left, header, content)

	// Add status bar
	statusBar := m.statusBar.View()
	combined := lipgloss.JoinVertical(lipgloss.Left, statsContent, statusBar)

	// Center if wide width (AC5: responsive layout)
	if m.isWideWidth() {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, combined)
	}
	return combined
}
