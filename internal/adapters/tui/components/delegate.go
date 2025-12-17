package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)

// Column widths for project row layout
const (
	colSelection = 2  // "> " or "  "
	colFavorite  = 2  // styled "⭐" or "  " (Story 3.8)
	colNameMin   = 15 // Minimum name width
	colIndicator = 3  // "✨ " or "⚡ " or "   "
	colStage     = 10 // "Implement" is longest
	colWaiting   = 14 // "⏸️ WAITING Xh" or empty
	colTime      = 8  // "2w ago" max
)

// Styles for project row rendering (mirrored from tui/styles.go to avoid import cycle)
var (
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("6")) // Cyan

	waitingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("1")) // Red

	recentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")) // Green

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")) // Yellow

	dimStyle = lipgloss.NewStyle().
			Faint(true)

	// favoriteStyle mirrors tui.FavoriteStyle (ANSI color 5 magenta) - keep in sync with styles.go (Story 3.8)
	favoriteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")) // Magenta
)

// ProjectItemDelegate is a custom delegate for rendering project rows.
type ProjectItemDelegate struct {
	width int
}

// NewProjectItemDelegate creates a new ProjectItemDelegate with the given width.
func NewProjectItemDelegate(width int) ProjectItemDelegate {
	return ProjectItemDelegate{width: width}
}

// SetWidth updates the delegate's width for responsive layout.
func (d *ProjectItemDelegate) SetWidth(width int) {
	d.width = width
}

// Height returns the height of each item (single-line rows).
func (d ProjectItemDelegate) Height() int {
	return 1
}

// Spacing returns the spacing between items.
func (d ProjectItemDelegate) Spacing() int {
	return 0
}

// Update handles messages for the delegate.
func (d ProjectItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render renders a single project row.
func (d ProjectItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(ProjectItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Calculate dynamic name width
	nameWidth := d.calculateNameWidth()

	// Build the row
	row := d.renderRow(item, isSelected, nameWidth)

	// Write to output
	fmt.Fprint(w, row)
}

// calculateNameWidth calculates the dynamic name column width based on terminal width.
func (d ProjectItemDelegate) calculateNameWidth() int {
	// Calculate available space for name
	// width - selection - favorite - indicator - stage - waiting - time - spacing (Story 3.8: added favorite)
	// Spacing breakdown: 5 = 1 (after name) + 1 (after indicator) + 1 (after stage) + 1 (after waiting) + 1 (before time)
	nameWidth := d.width - colSelection - colFavorite - colIndicator - colStage - colWaiting - colTime - 5

	if nameWidth < colNameMin {
		nameWidth = colNameMin
	}
	if nameWidth < 1 {
		nameWidth = 1 // Absolute minimum to prevent negative widths
	}

	return nameWidth
}

// renderRow renders a single project row with all columns.
func (d ProjectItemDelegate) renderRow(item ProjectItem, isSelected bool, nameWidth int) string {
	var sb strings.Builder

	// Selection indicator
	if isSelected {
		sb.WriteString("> ")
	} else {
		sb.WriteString("  ")
	}

	// Favorite indicator (Story 3.8)
	if item.Project.IsFavorite {
		sb.WriteString(favoriteStyle.Render("⭐"))
	} else {
		sb.WriteString("  ")
	}

	// Project name (truncate if needed)
	name := item.EffectiveName()
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}
	nameStr := fmt.Sprintf("%-*s", nameWidth, name)
	if isSelected {
		nameStr = selectedStyle.Render(nameStr)
	}
	sb.WriteString(nameStr)
	sb.WriteString(" ")

	// Recency indicator with styling
	indicator := timeformat.RecencyIndicator(item.Project.LastActivityAt)
	switch indicator {
	case "✨":
		sb.WriteString(recentStyle.Render("✨"))
	case "⚡":
		sb.WriteString(activeStyle.Render("⚡"))
	default:
		sb.WriteString("  ")
	}
	sb.WriteString(" ")

	// Stage name
	stage := item.Project.CurrentStage.String()
	stageStr := fmt.Sprintf("%-*s", colStage, stage)
	sb.WriteString(stageStr)
	sb.WriteString(" ")

	// WAITING indicator (placeholder for Story 4.x)
	waiting := waitingIndicator(item.Project)
	if waiting != "" {
		waitingStr := fmt.Sprintf("%-*s", colWaiting, waiting)
		sb.WriteString(waitingStyle.Render(waitingStr))
	} else {
		sb.WriteString(fmt.Sprintf("%-*s", colWaiting, ""))
	}
	sb.WriteString(" ")

	// Last activity time
	lastActive := timeformat.FormatRelativeTime(item.Project.LastActivityAt)
	timeStr := fmt.Sprintf("%*s", colTime, lastActive)
	sb.WriteString(dimStyle.Render(timeStr))

	return sb.String()
}

// isWaiting checks if a project is in waiting state.
// Placeholder: waiting detection comes in Story 4.3
func isWaiting(_ *domain.Project) bool {
	// TODO: Implement in Story 4.3
	// For now, always return false
	return false
}

// waitingIndicator returns the waiting indicator string for a project.
func waitingIndicator(p *domain.Project) string {
	if !isWaiting(p) {
		return ""
	}
	// Duration will come from Story 4.3 implementation
	return "⏸️ WAITING"
}
