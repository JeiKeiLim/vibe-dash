package components

import (
	"fmt"
	"strings"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/styles"
)

// Shortcut string constants (AC7: responsive width)
const (
	// Full shortcuts (width >= 80) - with pipe separators
	shortcutsFull = "│ [j/k] nav [d] details [f] fav [r] refresh [?] help [q] quit │"

	// Abbreviated shortcuts (width < 80)
	shortcutsAbbrev = "│ [j/k] [d] [f] [r] [?] [q] │"

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
	isCondensed      bool // True when height < 20 (Story 3.10 AC5)

	// Refresh state (Story 3.6)
	isRefreshing    bool
	refreshProgress int
	refreshTotal    int
	lastRefreshMsg  string // "Refreshed N projects" or error message

	// Watcher warning (Story 4.6)
	watcherWarning string // Empty means no warning, "⚠️ File watching unavailable" on error

	// Config warning (Story 7.2)
	configWarning string // Config error message (separate from watcher warning)

	// Loading state (Story 7.4)
	isLoading bool // True when loading projects
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

// SetCondensed sets the condensed mode for short terminals (Story 3.10 AC5).
func (s *StatusBarModel) SetCondensed(condensed bool) {
	s.isCondensed = condensed
}

// SetLoading sets the loading state (Story 7.4).
func (s *StatusBarModel) SetLoading(isLoading bool) {
	s.isLoading = isLoading
}

// SetRefreshing updates the refresh state (Story 3.6).
func (s *StatusBarModel) SetRefreshing(isRefreshing bool, progress, total int) {
	s.isRefreshing = isRefreshing
	s.refreshProgress = progress
	s.refreshTotal = total
}

// SetRefreshComplete sets the completion message (Story 3.6).
func (s *StatusBarModel) SetRefreshComplete(msg string) {
	s.lastRefreshMsg = msg
}

// SetWatcherWarning sets the file watcher warning message (Story 4.6).
// Pass empty string to clear the warning.
func (s *StatusBarModel) SetWatcherWarning(warning string) {
	s.watcherWarning = warning
}

// SetConfigWarning sets the config warning message (Story 7.2).
// Pass empty string to clear the warning.
func (s *StatusBarModel) SetConfigWarning(warning string) {
	s.configWarning = warning
}

// View renders the status bar to a string.
// Returns two lines: counts line and shortcuts line (AC1).
// Returns single line when condensed mode is active (Story 3.10 AC5).
func (s StatusBarModel) View() string {
	if s.isCondensed {
		// Single line: abbreviated counts + shortcuts (AC5)
		return s.renderCondensed()
	}
	countsLine := s.renderCounts()
	shortcutsLine := s.renderShortcuts()
	return countsLine + "\n" + shortcutsLine
}

// renderCondensed renders a single-line status bar for short terminals (Story 3.10 AC5).
// Must preserve all features from renderCounts() to avoid regression (Story 3.6).
func (s StatusBarModel) renderCondensed() string {
	// Story 7.4 AC1: Show loading indicator first
	if s.isLoading {
		return "│ Loading... │ [q] │"
	}

	// Show refresh spinner when refreshing (Story 3.6)
	if s.isRefreshing {
		return fmt.Sprintf("│ Refreshing %d/%d │ [j/k][?][q] │", s.refreshProgress, s.refreshTotal)
	}

	// Abbreviated counts (Story 3.10)
	counts := fmt.Sprintf("%dA %dH", s.activeCount, s.hibernatedCount)

	// Epic 4 Hotfix H1: Always show waiting count in condensed mode too
	if s.waitingCount > 0 {
		counts += " " + styles.WaitingStyle.Render(fmt.Sprintf("%dW", s.waitingCount))
	} else {
		counts += " " + styles.DimStyle.Render("0W")
	}

	// Include refresh message if present (Story 3.6)
	if s.lastRefreshMsg != "" {
		counts += " " + s.lastRefreshMsg
	}

	// Show abbreviated watcher warning if present (Story 4.6 AC3, Story 7.1: yellow styling)
	if s.watcherWarning != "" {
		counts += " " + styles.WarningStyle.Render("⚠️")
	}

	// Story 7.2: Show abbreviated config warning if present (AC6)
	if s.configWarning != "" {
		counts += " " + styles.WarningStyle.Render("⚠ cfg")
	}

	return "│ " + counts + " │ [j/k][?][q] │"
}

// renderCounts renders the counts line with pipe separators (AC1, AC4, AC5).
func (s StatusBarModel) renderCounts() string {
	// Story 7.4 AC1: Show loading indicator first
	if s.isLoading {
		return "│ Loading projects... │"
	}

	// Show refresh spinner when refreshing (Story 3.6 AC1)
	if s.isRefreshing {
		spinnerText := fmt.Sprintf("Refreshing... (%d/%d)", s.refreshProgress, s.refreshTotal)
		return "│ " + spinnerText + " │"
	}

	parts := []string{
		fmt.Sprintf("%d active", s.activeCount),
		fmt.Sprintf("%d hibernated", s.hibernatedCount),
	}

	// Epic 4 Hotfix H1: Always show waiting count so users know feature exists
	// AC4: If waitingCount > 0, show with waitingStyle (bold red)
	// H1: If waitingCount == 0, show with dim style
	if s.waitingCount > 0 {
		waitingText := styles.WaitingStyle.Render(fmt.Sprintf("⏸️ %d WAITING", s.waitingCount))
		parts = append(parts, waitingText)
	} else {
		waitingText := styles.DimStyle.Render("0 waiting")
		parts = append(parts, waitingText)
	}

	// Show refresh result for 3 seconds after completion (Story 3.6)
	if s.lastRefreshMsg != "" {
		parts = append(parts, s.lastRefreshMsg)
	}

	// Show watcher warning if present (Story 4.6 AC3, Story 7.1: yellow styling)
	if s.watcherWarning != "" {
		parts = append(parts, styles.WarningStyle.Render(s.watcherWarning))
	}

	// Story 7.2: Show config warning if present (AC6)
	if s.configWarning != "" {
		parts = append(parts, styles.WarningStyle.Render(s.configWarning))
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
// Backward-compatible: waiting is always 0.
// Use CalculateCountsWithWaiting() for actual waiting detection (Story 4.5).
// NOTE: Called from model.go via components.CalculateCounts()
func CalculateCounts(projects []*domain.Project) (active, hibernated, waiting int) {
	return CalculateCountsWithWaiting(projects, nil)
}

// CalculateCountsWithWaiting returns active, hibernated, and waiting counts.
// Story 4.5: Accepts a WaitingChecker callback to determine if active projects are waiting.
// Only active projects can be waiting (hibernated projects never show as waiting).
// If checker is nil, waiting count is always 0 (backward compatible).
func CalculateCountsWithWaiting(projects []*domain.Project, checker WaitingChecker) (active, hibernated, waiting int) {
	for _, p := range projects {
		switch p.State {
		case domain.StateActive:
			active++
			if checker != nil && checker(p) {
				waiting++
			}
		case domain.StateHibernated:
			hibernated++
		}
	}
	return
}
