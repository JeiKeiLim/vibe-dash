package components

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/emoji"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/stageformat"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/styles"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)

// Column widths for project row layout
const (
	colSelection = 2 // "> " or "  "
	colFavorite  = 2 // styled "⭐" or "  " (Story 3.8)
	colIndicator = 3 // "✨ " or "⚡ " or "   "
	colTime      = 8 // "2w ago" max

	// Fixed spacing between columns
	colSpacing = 5 // 1 (after name) + 1 (after indicator) + 1 (after stage) + 1 (after waiting) + 1 (before time)

	// Column width percentages (of available width after fixed columns) - Story 8.10
	colNamePct   = 25 // ~25% for project name
	colStagePct  = 40 // ~40% for stage info (needs space for "E8 S8.3 review")
	colStatusPct = 20 // ~20% for waiting status (fit "WAITING 2h 30m")
	// Remaining ~15% for activity time (fixed at 8 chars)

	// Column minimum widths
	colNameMin    = 10 // Minimum name width
	colStageMin   = 10 // Minimum stage width
	colWaitingMin = 19 // Minimum waiting width (fits "[W] WAITING 23h 59m")

	// Column maximum widths - prevent absurd stretching on ultra-wide (Story 8.10 AC6)
	colNameMax    = 40 // Project name max
	colStageMax   = 80 // Stage info max (E8 S8.10 + full status = ~40, double for safety)
	colWaitingMax = 25 // Waiting status max ("[W] WAITING 23h 59m" = 19 chars)

	// Width breakpoints for responsive stage display (Story 8.3)
	widthBreakpointFull  = 100 // >= 100: Full stage info "E8 S8.3 review"
	widthBreakpointShort = 80  // 80-99: Shortened "E8 S8.3"
)

// WaitingChecker checks if a project is waiting.
// Used by components to check waiting state without importing ports.
type WaitingChecker func(p *domain.Project) bool

// WaitingDurationGetter gets waiting duration for a project.
// Used by components to get waiting duration without importing ports.
type WaitingDurationGetter func(p *domain.Project) time.Duration

// AgentStateGetter returns the full agent detection state for a project.
// Story 15.7: Used by detail panel to display confidence level and detection source.
// Note: Does not take context parameter - caller captures context via closure.
type AgentStateGetter func(p *domain.Project) domain.AgentState

// ProjectItemDelegate is a custom delegate for rendering project rows.
type ProjectItemDelegate struct {
	width          int
	waitingChecker WaitingChecker        // nil = no waiting display (Story 4.5)
	durationGetter WaitingDurationGetter // nil = no duration display (Story 4.5)
}

// NewProjectItemDelegate creates a new ProjectItemDelegate with the given width.
// Backward-compatible constructor (waiting callbacks = nil).
func NewProjectItemDelegate(width int) ProjectItemDelegate {
	return ProjectItemDelegate{width: width}
}

// NewProjectItemDelegateWithWaiting creates a delegate with waiting detection callbacks.
// Story 4.5: Enables WAITING indicator display in project rows.
func NewProjectItemDelegateWithWaiting(width int, checker WaitingChecker, getter WaitingDurationGetter) ProjectItemDelegate {
	return ProjectItemDelegate{
		width:          width,
		waitingChecker: checker,
		durationGetter: getter,
	}
}

// SetWaitingCallbacks sets the waiting detection callbacks.
// Story 4.5: Used when model wires callbacks after construction.
func (d *ProjectItemDelegate) SetWaitingCallbacks(checker WaitingChecker, getter WaitingDurationGetter) {
	d.waitingChecker = checker
	d.durationGetter = getter
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

// stageColumnWidth returns the stage column width based on terminal width.
// Story 8.3: Responsive breakpoints for stage display.
// Story 8.10: Uses percentage-based calculation with max cap.
func (d ProjectItemDelegate) stageColumnWidth() int {
	if d.width < widthBreakpointShort {
		return 0 // Hidden at narrow widths
	}

	// Calculate available width after fixed columns
	availableWidth := d.availableWidth()

	// Calculate stage width from percentage
	stageWidth := int(float64(availableWidth) * float64(colStagePct) / 100.0)

	// Apply min/max constraints
	if stageWidth < colStageMin {
		stageWidth = colStageMin
	}
	if stageWidth > colStageMax {
		stageWidth = colStageMax
	}

	return stageWidth
}

// showStageColumn returns true if stage column should be shown.
// Story 8.3: Hide at width < 80.
func (d ProjectItemDelegate) showStageColumn() bool {
	return d.width >= widthBreakpointShort
}

// availableWidth returns the width available for dynamic columns (name, stage, waiting).
// Story 8.10: Subtracts fixed columns (selection, favorite, indicator, time, spacing).
func (d ProjectItemDelegate) availableWidth() int {
	// Fixed columns: selection(2) + favorite(2) + indicator(3) + time(8) + spacing(5) = 20
	fixedWidth := colSelection + colFavorite + colIndicator + colTime + colSpacing
	available := d.width - fixedWidth
	if available < 0 {
		return 0
	}
	return available
}

// calculateNameWidth calculates the dynamic name column width based on terminal width.
// Story 8.10: Uses percentage-based calculation with min/max constraints.
// Name truncates first to preserve stage info readability (AC2).
func (d ProjectItemDelegate) calculateNameWidth() int {
	// Calculate available width after fixed columns
	availableWidth := d.availableWidth()

	// Calculate name width from percentage
	nameWidth := int(float64(availableWidth) * float64(colNamePct) / 100.0)

	// Apply min/max constraints (Story 8.10 AC6)
	if nameWidth < colNameMin {
		nameWidth = colNameMin
	}
	if nameWidth > colNameMax {
		nameWidth = colNameMax
	}

	return nameWidth
}

// waitingColumnWidth calculates the waiting column width.
// Story 8.10: Uses percentage-based calculation with min/max constraints.
func (d ProjectItemDelegate) waitingColumnWidth() int {
	// Calculate available width after fixed columns
	availableWidth := d.availableWidth()

	// Calculate waiting width from percentage
	waitingWidth := int(float64(availableWidth) * float64(colStatusPct) / 100.0)

	// Apply min/max constraints (Story 8.10 AC6)
	if waitingWidth < colWaitingMin {
		waitingWidth = colWaitingMin
	}
	if waitingWidth > colWaitingMax {
		waitingWidth = colWaitingMax
	}

	return waitingWidth
}

// renderRow renders a single project row with all columns.
// Story 8.14: Row is padded to full width to ensure consistent alignment with bordered detail panel.
func (d ProjectItemDelegate) renderRow(item ProjectItem, isSelected bool, nameWidth int) string {
	var sb strings.Builder

	// Selection indicator
	if isSelected {
		sb.WriteString("> ")
	} else {
		sb.WriteString("  ")
	}

	// Favorite indicator (Story 3.8, Story 8.9: emoji fallback)
	if item.Project.IsFavorite {
		sb.WriteString(styles.FavoriteStyle.Render(emoji.Star()))
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
		nameStr = styles.SelectedStyle.Render(nameStr)
	}
	sb.WriteString(nameStr)
	sb.WriteString(" ")

	// Recency indicator with styling (Story 8.9: emoji fallback)
	indicator := timeformat.RecencyIndicator(item.Project.LastActivityAt)
	switch indicator {
	case "✨":
		sb.WriteString(styles.RecentStyle.Render(emoji.Today()))
	case "⚡":
		sb.WriteString(styles.ActiveStyle.Render(emoji.ThisWeek()))
	default:
		sb.WriteString("  ")
	}
	sb.WriteString(" ")

	// Stage info (Story 8.3: Rich BMAD stage display with responsive breakpoints)
	if d.showStageColumn() {
		stageWidth := d.stageColumnWidth()
		stage := stageformat.FormatStageInfoWithWidth(item.Project, stageWidth)
		stageStr := fmt.Sprintf("%-*s", stageWidth, stage)
		sb.WriteString(styles.DimStyle.Render(stageStr))
		sb.WriteString(" ")
	}

	// WAITING indicator (Story 4.5, Story 8.10: dynamic width)
	waitingWidth := d.waitingColumnWidth()
	waiting := d.waitingIndicator(item.Project)
	if waiting != "" {
		waitingStr := fmt.Sprintf("%-*s", waitingWidth, waiting)
		sb.WriteString(styles.WaitingStyle.Render(waitingStr))
	} else {
		sb.WriteString(fmt.Sprintf("%-*s", waitingWidth, ""))
	}
	sb.WriteString(" ")

	// Last activity time
	lastActive := timeformat.FormatRelativeTime(item.Project.LastActivityAt)
	timeStr := fmt.Sprintf("%*s", colTime, lastActive)
	sb.WriteString(styles.DimStyle.Render(timeStr))

	// Story 8.14: Pad row to full width to ensure consistent alignment
	// with bordered detail panel. Without this, rows are shorter than the
	// detail panel border, causing visual shift when toggling detail panel.
	row := sb.String()
	return lipgloss.NewStyle().Width(d.width).Render(row)
}

// waitingIndicator returns the waiting indicator string for a project.
// Story 4.5: Uses callbacks to determine waiting state and duration.
// Story 8.9: Uses emoji fallback for waiting indicator.
// Format: "⏸️ WAITING Xh" where X is the compact duration.
func (d ProjectItemDelegate) waitingIndicator(p *domain.Project) string {
	if d.waitingChecker == nil || !d.waitingChecker(p) {
		return ""
	}
	duration := time.Duration(0)
	if d.durationGetter != nil {
		duration = d.durationGetter(p)
	}
	return fmt.Sprintf("%s WAITING %s", emoji.Waiting(), timeformat.FormatWaitingDuration(duration, false))
}
