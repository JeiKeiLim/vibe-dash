package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/statsview"
)

// Story 16.3/16.4: renderStatsView renders the full-screen Stats View with activity sparklines.
// Layout: Header with "STATS" title and [ESC] hint, project list with sparklines.
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

	// Story 16.4: Determine bucket count based on terminal width (AC4)
	buckets := m.getSparklineBuckets(effectiveWidth)

	// Build project list content with sparklines
	content := m.renderStatsProjectList(effectiveWidth, contentHeight, buckets)

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

// getSparklineBuckets returns the number of sparkline buckets based on terminal width (AC4).
// Narrow terminals: 7 buckets (minimum)
// Wide terminals: up to 14 buckets
func (m Model) getSparklineBuckets(width int) int {
	if width < 60 {
		return 7
	}
	if width > 100 {
		return 14
	}
	// Linear scale between 60-100: 7 to 14 buckets
	return 7 + (width-60)/6
}

// renderStatsProjectList renders the list of projects with sparklines.
func (m Model) renderStatsProjectList(width, height, buckets int) string {
	if len(m.projects) == 0 {
		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("241")).
			Render("No projects to display")
	}

	var lines []string

	// Column header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Bold(true)

	// Calculate column widths
	sparklineWidth := buckets + 2 // sparkline chars + padding
	stageWidth := 15              // e.g., "Implementati..."
	nameWidth := width - sparklineWidth - stageWidth - 4

	if nameWidth < 10 {
		nameWidth = 10
	}

	headerLine := fmt.Sprintf("%-*s  %-*s  %s",
		nameWidth, "Project",
		sparklineWidth, "Activity (30d)",
		"Stage")
	lines = append(lines, headerStyle.Render(headerLine))

	// Separator
	lines = append(lines, strings.Repeat("─", width))

	// Project rows (scrollable)
	rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	sparklineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))  // Cyan for sparklines
	flatSparkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // Dim for no activity

	// Calculate visible rows (account for header and separator)
	visibleRows := height - 2
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Apply scroll offset
	startIdx := m.statsViewScroll
	endIdx := startIdx + visibleRows
	if endIdx > len(m.projects) {
		endIdx = len(m.projects)
	}
	if startIdx > len(m.projects)-visibleRows && len(m.projects) > visibleRows {
		startIdx = len(m.projects) - visibleRows
		if startIdx < 0 {
			startIdx = 0
		}
	}

	for i := startIdx; i < endIdx; i++ {
		p := m.projects[i]

		// Get activity counts for sparkline (uses project Path as ID - same as metricsRecorder)
		counts := m.getProjectActivity(p.Path, buckets)
		var sparkline string
		var sparkStyle lipgloss.Style

		if len(counts) == 0 {
			// No activity - flat sparkline (AC3)
			sparkline = strings.Repeat("▁", buckets)
			sparkStyle = flatSparkStyle
		} else {
			sparkline = statsview.RenderSparkline(counts)
			sparkStyle = sparklineStyle
		}

		// Truncate project name if needed
		name := p.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-1] + "…"
		}

		// Format stage (truncate if needed)
		stage := p.CurrentStage.String()
		if len(stage) > stageWidth {
			stage = stage[:stageWidth-1] + "…"
		}
		if stage == "" || stage == "Unknown" {
			stage = "Unknown"
		}

		// Build row
		row := fmt.Sprintf("%-*s  %s  %s",
			nameWidth, rowStyle.Render(name),
			sparkStyle.Render(fmt.Sprintf("%-*s", buckets, sparkline)),
			dimStyle.Render(fmt.Sprintf("%-*s", stageWidth, stage)),
		)
		lines = append(lines, row)
	}

	// Fill remaining height with empty lines
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines[:height], "\n")
}
