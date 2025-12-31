package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/emoji"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/project"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/styles"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)

const labelWidth = 12

// DetailPanelModel displays detailed information about a selected project.
type DetailPanelModel struct {
	project        *domain.Project
	width          int
	height         int
	visible        bool
	waitingChecker WaitingChecker        // nil = no waiting display (Story 4.5)
	durationGetter WaitingDurationGetter // nil = no duration display (Story 4.5)
	isHorizontal   bool                  // Story 8.12: Use horizontal border style when true
}

// NewDetailPanelModel creates a new DetailPanelModel with the given dimensions.
// Backward-compatible constructor (waiting callbacks = nil).
func NewDetailPanelModel(width, height int) DetailPanelModel {
	return DetailPanelModel{
		width:   width,
		height:  height,
		visible: false,
	}
}

// SetWaitingCallbacks sets the waiting detection callbacks.
// Story 4.5: Enables waiting status display in detail panel.
func (m *DetailPanelModel) SetWaitingCallbacks(checker WaitingChecker, getter WaitingDurationGetter) {
	m.waitingChecker = checker
	m.durationGetter = getter
}

// SetProject updates the displayed project.
func (m *DetailPanelModel) SetProject(p *domain.Project) {
	m.project = p
}

// Project returns the currently displayed project.
// Story 4.6: Used to check if file events affect the displayed project.
func (m DetailPanelModel) Project() *domain.Project {
	return m.project
}

// SetSize updates the panel dimensions for responsive layout.
func (m *DetailPanelModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetVisible shows or hides the panel.
func (m *DetailPanelModel) SetVisible(visible bool) {
	m.visible = visible
}

// SetHorizontalMode sets horizontal layout mode for border styling (Story 8.12).
// When true, uses HorizontalBorderStyle (no top border) to save vertical space.
func (m *DetailPanelModel) SetHorizontalMode(horizontal bool) {
	m.isHorizontal = horizontal
}

// Width returns the current width of the detail panel.
// Story 8.12: Used for save/restore during horizontal layout render.
func (m DetailPanelModel) Width() int {
	return m.width
}

// Height returns the current height of the detail panel.
// Story 8.12: Used for save/restore during horizontal layout render.
func (m DetailPanelModel) Height() int {
	return m.height
}

// IsVisible returns whether the panel is visible.
func (m DetailPanelModel) IsVisible() bool {
	return m.visible
}

// View renders the detail panel to a string.
func (m DetailPanelModel) View() string {
	if !m.visible {
		return ""
	}

	if m.project == nil {
		return m.renderEmpty()
	}

	return m.renderProject()
}

// renderEmpty renders a placeholder when no project is selected.
func (m DetailPanelModel) renderEmpty() string {
	// Use shared BorderStyle
	panelBorder := styles.BorderStyle.
		Width(m.width - 2).
		Height(m.height - 2)

	content := styles.DimStyle.Render("No project selected")

	return panelBorder.Render(content)
}

// renderProject renders the project details.
func (m DetailPanelModel) renderProject() string {
	p := m.project

	// Build content lines
	var lines []string

	// Title with project name using shared TitleStyle
	effectiveName := project.EffectiveName(p)
	title := styles.TitleStyle.Render("DETAILS: " + effectiveName)
	lines = append(lines, title)
	lines = append(lines, "") // Empty line after title

	// Path
	lines = append(lines, formatField("Path", p.Path))

	// Method
	method := p.DetectedMethod
	if method == "" {
		method = "Unknown"
	}
	lines = append(lines, formatField("Method", method))

	// Stage
	stage := p.CurrentStage.String()
	if stage == "" {
		stage = "Unknown"
	}
	lines = append(lines, formatField("Stage", stage))

	// Detection reasoning
	reasoning := p.DetectionReasoning
	if reasoning == "" {
		reasoning = "No detection reasoning available"
	}
	lines = append(lines, formatField("Detection", reasoning))

	// Notes
	notes := p.Notes
	if notes == "" {
		notes = "(none)"
	}
	lines = append(lines, formatField("Notes", notes))

	// Favorite status (Story 3.8, Story 8.9: emoji fallback)
	favorite := "No"
	if p.IsFavorite {
		favorite = emoji.Star() + " Yes"
	}
	lines = append(lines, formatField("Favorite", favorite))

	// Added date
	addedDate := p.CreatedAt.Format("2006-01-02")
	lines = append(lines, formatField("Added", addedDate))

	// Last Active (relative time)
	lastActive := timeformat.FormatRelativeTime(p.LastActivityAt)
	lines = append(lines, formatField("Last Active", lastActive))

	// Waiting status (Story 4.5, Story 8.9: emoji fallback) - only shown when project is waiting
	if m.waitingChecker != nil && m.waitingChecker(p) {
		duration := time.Duration(0)
		if m.durationGetter != nil {
			duration = m.durationGetter(p)
		}
		waitingText := fmt.Sprintf("%s %s", emoji.Waiting(), timeformat.FormatWaitingDuration(duration, true))
		styledWaiting := styles.WaitingStyle.Render(waitingText)
		lines = append(lines, formatField("Waiting", styledWaiting))
	}

	// Join lines
	content := strings.Join(lines, "\n")

	// Story 8.12: Use HorizontalBorderStyle when in horizontal mode (saves 1 vertical line)
	borderStyle := styles.BorderStyle
	if m.isHorizontal {
		borderStyle = styles.HorizontalBorderStyle
	}

	// Apply border style with dimensions
	panelBorder := borderStyle.
		Width(m.width-2).
		Height(m.height-2).
		Padding(0, 1)

	return panelBorder.Render(content)
}

// formatField formats a label-value pair with consistent alignment.
func formatField(label, value string) string {
	paddedLabel := lipgloss.NewStyle().
		Width(labelWidth).
		Render(label + ":")
	return paddedLabel + " " + value
}
