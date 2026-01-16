package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/statsview"
)

// Story 16.3/16.4/16.5: renderStatsView renders Stats View with sparklines or breakdown.
// Layout: Header with "STATS" title and [ESC] hint, project list with sparklines.
// Story 16.5: Checks statsBreakdownProject to switch between list and breakdown views.
func (m Model) renderStatsView() string {
	// Story 16.5: If breakdown project is selected, show breakdown view
	if m.statsBreakdownProject != nil {
		return m.renderStatsBreakdownView()
	}
	return m.renderStatsProjectListView()
}

// renderStatsProjectListView renders the project list with sparklines (original renderStatsView).
func (m Model) renderStatsProjectListView() string {
	// Calculate effective width (Story 8.10: respect max content width)
	effectiveWidth := m.width
	if m.isWideWidth() {
		effectiveWidth = m.maxContentWidth
	}

	// Header: "STATS" title on left, "[ ] Range  [ESC] Back" hint on right (Story 16.6)
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Render("STATS")
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[ ] Range  [ESC] Back")

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

	// Story 16.6: Use dynamic date range label from statsDateRange
	headerLine := fmt.Sprintf("%-*s  %-*s  %s",
		nameWidth, "Project",
		sparklineWidth, m.statsDateRange.HeaderLabel(),
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

// renderStatsBreakdownView renders the detailed time-per-stage breakdown for a project.
// Story 16.5: Shows horizontal bars with duration and percentage per stage.
func (m Model) renderStatsBreakdownView() string {
	// Calculate effective width (Story 8.10: respect max content width)
	effectiveWidth := m.width
	if m.isWideWidth() {
		effectiveWidth = m.maxContentWidth
	}

	// Header: "STATS" title on left, "[ESC] Back to Project List" hint on right
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Render("STATS")
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[ESC] Back to Project List")

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

	// Project info section
	projectStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	projectName := m.statsBreakdownProject.Name
	if len(projectName) > effectiveWidth-15 {
		projectName = projectName[:effectiveWidth-18] + "..."
	}

	var contentLines []string
	contentLines = append(contentLines, "")
	contentLines = append(contentLines, projectStyle.Render("Project: ")+projectName)
	// Story 16.6: Use dynamic date range label from statsDateRange
	contentLines = append(contentLines, dimStyle.Render("Period: "+m.statsDateRange.BreakdownLabel()))
	contentLines = append(contentLines, "")
	contentLines = append(contentLines, strings.Repeat("─", effectiveWidth))
	contentLines = append(contentLines, "")
	contentLines = append(contentLines, projectStyle.Render("Stage Breakdown:"))
	contentLines = append(contentLines, "")

	// Render breakdown or "no data" message
	if len(m.statsBreakdownDurations) == 0 {
		contentLines = append(contentLines, dimStyle.Render("  No stage data available"))
	} else {
		// Render breakdown bars
		breakdownStr := statsview.RenderBreakdown(m.statsBreakdownDurations, effectiveWidth-4)
		for _, line := range strings.Split(breakdownStr, "\n") {
			contentLines = append(contentLines, "  "+line)
		}

		// Calculate and display total
		contentLines = append(contentLines, "")
		total := statsview.CalculateTotalDuration(m.statsBreakdownDurations)
		totalStr := formatTotalDuration(total)
		contentLines = append(contentLines, projectStyle.Render("Total: ")+totalStr)
	}

	contentLines = append(contentLines, "")
	contentLines = append(contentLines, strings.Repeat("─", effectiveWidth))

	// Fill remaining height with empty lines
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, "")
	}

	content := strings.Join(contentLines[:contentHeight], "\n")

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

// formatTotalDuration formats total duration for display in breakdown view.
func formatTotalDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}

	return "< 1m"
}
