package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// Styles duplicated from tui/styles.go to avoid import cycle
// (components package cannot import tui package).
// Keep in sync with styles.go definitions.
var (
	// statusBarWaitingStyle matches tui.WaitingStyle (bold red for WAITING)
	statusBarWaitingStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("1")) // Red
)

// Shortcut string constants (AC7: responsive width)
const (
	// Full shortcuts (width >= 80) - with pipe separators
	shortcutsFull = "│ [j/k] nav [d] details [f] fav [?] help [q] quit │"

	// Abbreviated shortcuts (width < 80)
	shortcutsAbbrev = "│ [j/k] [d] [f] [?] [q] │"

	// Width threshold for abbreviation (AC7)
	widthThreshold = 80
)

// Future Story 5.4: Hibernated view shortcuts (placeholder constants)
// nolint:unused // Placeholder for Story 5.4 - Hibernated Projects View
const (
	_shortcutsHibernatedFull   = "│ [j/k] nav [h] back to active [?] help [q] quit │"
	_shortcutsHibernatedAbbrev = "│ [j/k] [h] [?] [q] │"
)

// StatusBarModel displays summary counts and keyboard shortcuts.
type StatusBarModel struct {
	activeCount      int
	hibernatedCount  int
	waitingCount     int
	width            int
	inHibernatedView bool // Placeholder for Story 5.4
}

// NewStatusBarModel creates a new StatusBarModel with the given width.
func NewStatusBarModel(width int) StatusBarModel {
	return StatusBarModel{
		width: width,
	}
}

// SetCounts updates the project counts.
func (s *StatusBarModel) SetCounts(active, hibernated, waiting int) {
	s.activeCount = active
	s.hibernatedCount = hibernated
	s.waitingCount = waiting
}

// SetWidth updates the status bar width for responsive layout.
func (s *StatusBarModel) SetWidth(width int) {
	s.width = width
}

// SetInHibernatedView sets whether user is in hibernated view (placeholder for Story 5.4).
// Does not affect rendering until Story 5.4 is implemented.
func (s *StatusBarModel) SetInHibernatedView(inView bool) {
	s.inHibernatedView = inView
}

// View renders the status bar to a string.
// Returns two lines: counts line and shortcuts line (AC1).
func (s StatusBarModel) View() string {
	countsLine := s.renderCounts()
	shortcutsLine := s.renderShortcuts()
	return countsLine + "\n" + shortcutsLine
}

// renderCounts renders the counts line with pipe separators (AC1, AC4, AC5).
func (s StatusBarModel) renderCounts() string {
	parts := []string{
		fmt.Sprintf("%d active", s.activeCount),
		fmt.Sprintf("%d hibernated", s.hibernatedCount),
	}

	// AC4: If waitingCount > 0, show with waitingStyle (bold red)
	// AC5: If waitingCount == 0, hide WAITING section
	if s.waitingCount > 0 {
		waitingText := statusBarWaitingStyle.Render(fmt.Sprintf("⏸️ %d WAITING", s.waitingCount))
		parts = append(parts, waitingText)
	}

	return "│ " + strings.Join(parts, " │ ") + " │"
}

// renderShortcuts renders the shortcuts line (AC7: responsive width).
func (s StatusBarModel) renderShortcuts() string {
	// TODO(Story-5.4): Use hibernated shortcuts when s.inHibernatedView is true
	if s.width >= widthThreshold {
		return shortcutsFull
	}
	return shortcutsAbbrev
}

// CalculateCounts returns active, hibernated, and waiting counts from projects.
// NOTE: Called from model.go via components.CalculateCounts()
func CalculateCounts(projects []*domain.Project) (active, hibernated, waiting int) {
	for _, p := range projects {
		switch p.State {
		case domain.StateActive:
			active++
		case domain.StateHibernated:
			hibernated++
		}
		// TODO(Story-4.3): Add waiting detection when p.IsWaiting field exists
		// For now, waiting count is always 0
	}
	return
}
