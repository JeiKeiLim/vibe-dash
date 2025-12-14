package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/project"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)

// Styles duplicated from tui/styles.go to avoid import cycle
// (components package cannot import tui package).
// Keep in sync with styles.go definitions.
// Note: dimStyle is already declared in delegate.go
var (
	// detailBorderStyle matches tui.BorderStyle
	detailBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("8"))

	// detailTitleStyle matches tui.TitleStyle
	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39")) // Cyan

	// uncertainStyle matches tui.UncertainStyle
	uncertainStyle = lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color("8")) // Bright black (gray)
)

const labelWidth = 12

// DetailPanelModel displays detailed information about a selected project.
type DetailPanelModel struct {
	project *domain.Project
	width   int
	height  int
	visible bool
}

// NewDetailPanelModel creates a new DetailPanelModel with the given dimensions.
func NewDetailPanelModel(width, height int) DetailPanelModel {
	return DetailPanelModel{
		width:   width,
		height:  height,
		visible: false,
	}
}

// SetProject updates the displayed project.
func (m *DetailPanelModel) SetProject(p *domain.Project) {
	m.project = p
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
	// Use local detailBorderStyle (matches tui.BorderStyle)
	panelBorder := detailBorderStyle.
		Width(m.width - 2).
		Height(m.height - 2)

	content := dimStyle.Render("No project selected")

	return panelBorder.Render(content)
}

// renderProject renders the project details.
func (m DetailPanelModel) renderProject() string {
	p := m.project

	// Build content lines
	var lines []string

	// Title with project name using local detailTitleStyle (matches tui.TitleStyle)
	effectiveName := project.EffectiveName(p)
	title := detailTitleStyle.Render("DETAILS: " + effectiveName)
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

	// Confidence (with style for uncertain)
	lines = append(lines, formatField("Confidence", renderConfidence(p.Confidence)))

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

	// Added date
	addedDate := p.CreatedAt.Format("2006-01-02")
	lines = append(lines, formatField("Added", addedDate))

	// Last Active (relative time)
	lastActive := timeformat.FormatRelativeTime(p.LastActivityAt)
	lines = append(lines, formatField("Last Active", lastActive))

	// Join lines
	content := strings.Join(lines, "\n")

	// Apply local detailBorderStyle with dimensions (matches tui.BorderStyle)
	panelBorder := detailBorderStyle.
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

// renderConfidence returns the confidence level with appropriate styling.
func renderConfidence(conf domain.Confidence) string {
	switch conf {
	case domain.ConfidenceCertain:
		return "Certain"
	case domain.ConfidenceLikely:
		return "Likely"
	case domain.ConfidenceUncertain:
		// Apply local uncertainStyle (matches tui.UncertainStyle)
		return uncertainStyle.Render("Uncertain")
	default:
		return fmt.Sprintf("Unknown (%d)", conf)
	}
}
