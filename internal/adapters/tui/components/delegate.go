package components

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/stageformat"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/styles"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)

// Column widths for project row layout
const (
	colSelection = 2  // "> " or "  "
	colFavorite  = 2  // styled "⭐" or "  " (Story 3.8)
	colNameMin   = 15 // Minimum name width
	colIndicator = 3  // "✨ " or "⚡ " or "   "
	colStage     = 16 // "E8 S8.3 review" needs 14, allow 16 for padding (Story 8.3)
	colWaiting   = 14 // "⏸️ WAITING Xh" or empty
	colTime      = 8  // "2w ago" max

	// Width breakpoints for responsive stage display (Story 8.3)
	widthBreakpointFull  = 100 // >= 100: Full stage info "E8 S8.3 review"
	widthBreakpointShort = 80  // 80-99: Shortened "E8 S8.3"
	colStageShort        = 10  // Column width for shortened stage
	colStageFull         = 16  // Column width for full stage
)

// WaitingChecker checks if a project is waiting.
// Used by components to check waiting state without importing ports.
type WaitingChecker func(p *domain.Project) bool

// WaitingDurationGetter gets waiting duration for a project.
// Used by components to get waiting duration without importing ports.
type WaitingDurationGetter func(p *domain.Project) time.Duration

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
func (d ProjectItemDelegate) stageColumnWidth() int {
	if d.width >= widthBreakpointFull {
		return colStageFull // Full stage info: "E8 S8.3 review"
	}
	if d.width >= widthBreakpointShort {
		return colStageShort // Shortened: "E8 S8.3"
	}
	return 0 // Hidden at narrow widths
}

// showStageColumn returns true if stage column should be shown.
// Story 8.3: Hide at width < 80.
func (d ProjectItemDelegate) showStageColumn() bool {
	return d.width >= widthBreakpointShort
}

// calculateNameWidth calculates the dynamic name column width based on terminal width.
func (d ProjectItemDelegate) calculateNameWidth() int {
	// Calculate available space for name
	// width - selection - favorite - indicator - stage - waiting - time - spacing (Story 3.8: added favorite)
	// Spacing breakdown: 5 = 1 (after name) + 1 (after indicator) + 1 (after stage) + 1 (after waiting) + 1 (before time)
	// Story 8.3: Stage column size is responsive
	stageWidth := d.stageColumnWidth()
	nameWidth := d.width - colSelection - colFavorite - colIndicator - stageWidth - colWaiting - colTime - 5

	// Story 8.3: If stage is hidden, reclaim that space
	if !d.showStageColumn() {
		nameWidth += 1 // Reclaim the space after stage column
	}

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
		sb.WriteString(styles.FavoriteStyle.Render("⭐"))
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

	// Recency indicator with styling
	indicator := timeformat.RecencyIndicator(item.Project.LastActivityAt)
	switch indicator {
	case "✨":
		sb.WriteString(styles.RecentStyle.Render("✨"))
	case "⚡":
		sb.WriteString(styles.ActiveStyle.Render("⚡"))
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

	// WAITING indicator (Story 4.5)
	waiting := d.waitingIndicator(item.Project)
	if waiting != "" {
		waitingStr := fmt.Sprintf("%-*s", colWaiting, waiting)
		sb.WriteString(styles.WaitingStyle.Render(waitingStr))
	} else {
		sb.WriteString(fmt.Sprintf("%-*s", colWaiting, ""))
	}
	sb.WriteString(" ")

	// Last activity time
	lastActive := timeformat.FormatRelativeTime(item.Project.LastActivityAt)
	timeStr := fmt.Sprintf("%*s", colTime, lastActive)
	sb.WriteString(styles.DimStyle.Render(timeStr))

	return sb.String()
}

// waitingIndicator returns the waiting indicator string for a project.
// Story 4.5: Uses callbacks to determine waiting state and duration.
// Format: "⏸️ WAITING Xh" where X is the compact duration.
func (d ProjectItemDelegate) waitingIndicator(p *domain.Project) string {
	if d.waitingChecker == nil || !d.waitingChecker(p) {
		return ""
	}
	duration := time.Duration(0)
	if d.durationGetter != nil {
		duration = d.durationGetter(p)
	}
	return fmt.Sprintf("⏸️ WAITING %s", timeformat.FormatWaitingDuration(duration, false))
}
