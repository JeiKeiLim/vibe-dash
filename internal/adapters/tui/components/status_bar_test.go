package components

import (
	"strings"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// Task 1: StatusBar component tests

func TestStatusBar_View_BasicCounts(t *testing.T) {
	// AC1: verify counts render with pipes
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)

	view := sb.View()

	// Line 1 should have counts with pipe separators
	if !strings.Contains(view, "5 active") {
		t.Errorf("expected '5 active' in view, got: %s", view)
	}
	if !strings.Contains(view, "2 hibernated") {
		t.Errorf("expected '2 hibernated' in view, got: %s", view)
	}
	// Should have pipe separators
	if !strings.Contains(view, "│") {
		t.Errorf("expected pipe separator '│' in view, got: %s", view)
	}
}

func TestStatusBar_View_WaitingHighlighted(t *testing.T) {
	// AC4: verify WAITING uses waitingStyle when count > 0
	sb := NewStatusBarModel(100)
	sb.SetCounts(3, 1, 2)

	view := sb.View()

	// Should contain WAITING indicator
	if !strings.Contains(view, "WAITING") {
		t.Errorf("expected 'WAITING' in view when count > 0, got: %s", view)
	}
	if !strings.Contains(view, "⏸️") {
		t.Errorf("expected pause emoji in WAITING section, got: %s", view)
	}
	if !strings.Contains(view, "2") {
		t.Errorf("expected waiting count '2' in view, got: %s", view)
	}
}

func TestStatusBar_View_WaitingZero(t *testing.T) {
	// Epic 4 Hotfix H1: verify "0 waiting" shown (dim style) so users know feature exists
	sb := NewStatusBarModel(100)
	sb.SetCounts(3, 1, 0)

	view := sb.View()

	// "0 waiting" should be shown (not hidden) so users know the feature exists
	if !strings.Contains(view, "0 waiting") {
		t.Errorf("expected '0 waiting' in view when count is 0 (Epic 4 Hotfix H1), got: %s", view)
	}
	// Should NOT show "WAITING" in caps (that's for count > 0)
	if strings.Contains(view, "WAITING") {
		t.Errorf("expected 'WAITING' to NOT appear when count is 0 (use '0 waiting' instead), got: %s", view)
	}
}

func TestStatusBar_View_ShortcutsFull(t *testing.T) {
	// AC7: verify full shortcuts at width >= 80
	sb := NewStatusBarModel(100) // width >= 80

	view := sb.View()

	// Should have full shortcuts with descriptions
	if !strings.Contains(view, "[j/k] nav") {
		t.Errorf("expected '[j/k] nav' in full shortcuts, got: %s", view)
	}
	if !strings.Contains(view, "[d] details") {
		t.Errorf("expected '[d] details' in full shortcuts, got: %s", view)
	}
	if !strings.Contains(view, "[r] refresh") {
		t.Errorf("expected '[r] refresh' in full shortcuts, got: %s", view)
	}
	if !strings.Contains(view, "[q] quit") {
		t.Errorf("expected '[q] quit' in full shortcuts, got: %s", view)
	}
}

func TestStatusBar_View_ShortcutsAbbreviated(t *testing.T) {
	// AC7: verify abbreviated shortcuts at width < 80
	sb := NewStatusBarModel(60) // width < 80

	view := sb.View()

	// Should have abbreviated shortcuts (keys only)
	if !strings.Contains(view, "[j/k]") {
		t.Errorf("expected '[j/k]' in abbreviated shortcuts, got: %s", view)
	}
	if !strings.Contains(view, "[d]") {
		t.Errorf("expected '[d]' in abbreviated shortcuts, got: %s", view)
	}
	if !strings.Contains(view, "[f]") {
		t.Errorf("expected '[f]' in abbreviated shortcuts, got: %s", view)
	}
	if !strings.Contains(view, "[r]") {
		t.Errorf("expected '[r]' in abbreviated shortcuts, got: %s", view)
	}
	// Should NOT have full descriptions
	if strings.Contains(view, "[d] details") {
		t.Errorf("expected abbreviated shortcuts without 'details', got: %s", view)
	}
}

func TestStatusBar_View_ShortcutsBoundary(t *testing.T) {
	// L3: Test exact boundary at width=80 (AC7: < 80 abbreviated, >= 80 full)
	sb := NewStatusBarModel(80) // Exactly at threshold

	view := sb.View()

	// At width=80, should show FULL shortcuts (not abbreviated)
	if !strings.Contains(view, "[d] details") {
		t.Errorf("expected full shortcuts at width 80, got: %s", view)
	}
}

func TestStatusBar_View_ZeroCounts(t *testing.T) {
	// L2: Verify "0 active" and "0 hibernated" display correctly
	sb := NewStatusBarModel(100)
	sb.SetCounts(0, 0, 0)

	view := sb.View()

	if !strings.Contains(view, "0 active") {
		t.Errorf("expected '0 active' in view, got: %s", view)
	}
	if !strings.Contains(view, "0 hibernated") {
		t.Errorf("expected '0 hibernated' in view, got: %s", view)
	}
	// WAITING should be hidden when 0
	if strings.Contains(view, "WAITING") {
		t.Errorf("expected WAITING to be hidden when count is 0, got: %s", view)
	}
}

func TestStatusBar_SetCounts(t *testing.T) {
	sb := NewStatusBarModel(100)

	sb.SetCounts(10, 5, 3)

	view := sb.View()
	if !strings.Contains(view, "10 active") {
		t.Errorf("expected updated count '10 active', got: %s", view)
	}
	if !strings.Contains(view, "5 hibernated") {
		t.Errorf("expected updated count '5 hibernated', got: %s", view)
	}
	if !strings.Contains(view, "3") && !strings.Contains(view, "WAITING") {
		t.Errorf("expected updated waiting count, got: %s", view)
	}
}

func TestStatusBar_SetWidth(t *testing.T) {
	sb := NewStatusBarModel(100)

	// Initially full shortcuts
	view1 := sb.View()
	if !strings.Contains(view1, "[d] details") {
		t.Errorf("expected full shortcuts at width 100, got: %s", view1)
	}

	// After reducing width, should abbreviate
	sb.SetWidth(60)
	view2 := sb.View()
	if strings.Contains(view2, "[d] details") {
		t.Errorf("expected abbreviated shortcuts at width 60, got: %s", view2)
	}
}

func TestStatusBar_SetInHibernatedView(t *testing.T) {
	// Placeholder method for Story 5.4 - verify it sets the field
	sb := NewStatusBarModel(100)

	// Initially should be false
	if sb.inHibernatedView {
		t.Error("inHibernatedView should initially be false")
	}

	// Set to true
	sb.SetInHibernatedView(true)
	if !sb.inHibernatedView {
		t.Error("SetInHibernatedView(true) should set inHibernatedView to true")
	}

	// Set back to false
	sb.SetInHibernatedView(false)
	if sb.inHibernatedView {
		t.Error("SetInHibernatedView(false) should set inHibernatedView to false")
	}
}

// Task 2: CalculateCounts tests

func TestCalculateCounts(t *testing.T) {
	projects := []*domain.Project{
		{State: domain.StateActive},
		{State: domain.StateActive},
		{State: domain.StateHibernated},
	}

	active, hibernated, waiting := CalculateCounts(projects)

	if active != 2 {
		t.Errorf("expected active=2, got %d", active)
	}
	if hibernated != 1 {
		t.Errorf("expected hibernated=1, got %d", hibernated)
	}
	if waiting != 0 {
		t.Errorf("expected waiting=0 (placeholder), got %d", waiting)
	}
}

func TestCalculateCounts_MixedStates(t *testing.T) {
	projects := []*domain.Project{
		{State: domain.StateActive},
		{State: domain.StateHibernated},
		{State: domain.StateActive},
		{State: domain.StateHibernated},
		{State: domain.StateActive},
	}

	active, hibernated, waiting := CalculateCounts(projects)

	if active != 3 {
		t.Errorf("expected active=3, got %d", active)
	}
	if hibernated != 2 {
		t.Errorf("expected hibernated=2, got %d", hibernated)
	}
	if waiting != 0 {
		t.Errorf("expected waiting=0, got %d", waiting)
	}
}

func TestCalculateCounts_Empty(t *testing.T) {
	var projects []*domain.Project

	active, hibernated, waiting := CalculateCounts(projects)

	if active != 0 {
		t.Errorf("expected active=0, got %d", active)
	}
	if hibernated != 0 {
		t.Errorf("expected hibernated=0, got %d", hibernated)
	}
	if waiting != 0 {
		t.Errorf("expected waiting=0, got %d", waiting)
	}
}

func TestCalculateCounts_AllActive(t *testing.T) {
	projects := []*domain.Project{
		{State: domain.StateActive},
		{State: domain.StateActive},
		{State: domain.StateActive},
	}

	active, hibernated, waiting := CalculateCounts(projects)

	if active != 3 {
		t.Errorf("expected active=3, got %d", active)
	}
	if hibernated != 0 {
		t.Errorf("expected hibernated=0, got %d", hibernated)
	}
	if waiting != 0 {
		t.Errorf("expected waiting=0, got %d", waiting)
	}
}

func TestCalculateCounts_AllHibernated(t *testing.T) {
	projects := []*domain.Project{
		{State: domain.StateHibernated},
		{State: domain.StateHibernated},
	}

	active, hibernated, waiting := CalculateCounts(projects)

	if active != 0 {
		t.Errorf("expected active=0, got %d", active)
	}
	if hibernated != 2 {
		t.Errorf("expected hibernated=2, got %d", hibernated)
	}
	if waiting != 0 {
		t.Errorf("expected waiting=0, got %d", waiting)
	}
}

// Tests for two-line layout
func TestStatusBar_TwoLineLayout(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)

	view := sb.View()
	lines := strings.Split(view, "\n")

	if len(lines) != 2 {
		t.Errorf("expected 2 lines in status bar, got %d: %s", len(lines), view)
	}
}

// ============================================================================
// Story 3.6: Refresh State Tests
// ============================================================================

func TestStatusBar_RefreshingState(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)
	sb.SetRefreshing(true, 2, 5)

	view := sb.View()

	if !strings.Contains(view, "Refreshing...") {
		t.Errorf("expected 'Refreshing...' in output, got: %s", view)
	}
	if !strings.Contains(view, "2/5") {
		t.Errorf("expected progress '2/5' in output, got: %s", view)
	}
}

func TestStatusBar_RefreshingStateZeroProgress(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetRefreshing(true, 0, 3)

	view := sb.View()

	if !strings.Contains(view, "Refreshing...") {
		t.Errorf("expected 'Refreshing...' in output, got: %s", view)
	}
	if !strings.Contains(view, "0/3") {
		t.Errorf("expected progress '0/3' in output, got: %s", view)
	}
}

func TestStatusBar_RefreshingHidesNormalCounts(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)
	sb.SetRefreshing(true, 1, 3)

	view := sb.View()

	// Should NOT show normal counts while refreshing
	if strings.Contains(view, "5 active") {
		t.Errorf("expected normal counts to be hidden while refreshing, got: %s", view)
	}
	if strings.Contains(view, "2 hibernated") {
		t.Errorf("expected normal counts to be hidden while refreshing, got: %s", view)
	}
}

func TestStatusBar_RefreshComplete(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)
	sb.SetRefreshComplete("Refreshed 3 projects")

	view := sb.View()

	if !strings.Contains(view, "Refreshed 3 projects") {
		t.Errorf("expected completion message 'Refreshed 3 projects' in output, got: %s", view)
	}
	// Should also show normal counts
	if !strings.Contains(view, "5 active") {
		t.Errorf("expected normal counts with completion message, got: %s", view)
	}
}

func TestStatusBar_RefreshComplete_Cleared(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)
	sb.SetRefreshComplete("Refreshed 3 projects")
	sb.SetRefreshComplete("") // Clear

	view := sb.View()

	if strings.Contains(view, "Refreshed") {
		t.Errorf("expected completion message to be cleared, got: %s", view)
	}
	// Should still show normal counts
	if !strings.Contains(view, "5 active") {
		t.Errorf("expected normal counts after clearing, got: %s", view)
	}
}

func TestStatusBar_RefreshComplete_Error(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)
	sb.SetRefreshComplete("Refresh failed")

	view := sb.View()

	if !strings.Contains(view, "Refresh failed") {
		t.Errorf("expected error message 'Refresh failed' in output, got: %s", view)
	}
}

func TestStatusBar_SetRefreshing_EndRefresh(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)
	sb.SetRefreshing(true, 2, 3)
	sb.SetRefreshing(false, 0, 0) // End refresh

	view := sb.View()

	// Should NOT show refreshing anymore
	if strings.Contains(view, "Refreshing...") {
		t.Errorf("expected refreshing to end, got: %s", view)
	}
	// Should show normal counts again
	if !strings.Contains(view, "5 active") {
		t.Errorf("expected normal counts after refresh ends, got: %s", view)
	}
}

// ============================================================================
// Story 3.10: Condensed Mode Tests
// ============================================================================

func TestStatusBarModel_CondensedMode(t *testing.T) {
	sb := NewStatusBarModel(80)
	sb.SetCounts(5, 3, 2)
	sb.SetCondensed(true)

	view := sb.View()

	// Should be single line
	if strings.Count(view, "\n") > 0 {
		t.Error("condensed view should be single line")
	}

	// Should contain abbreviated counts
	if !strings.Contains(view, "5A") {
		t.Error("expected abbreviated active count '5A'")
	}
	if !strings.Contains(view, "3H") {
		t.Error("expected abbreviated hibernated count '3H'")
	}
	if !strings.Contains(view, "2W") {
		t.Error("expected abbreviated waiting count '2W'")
	}
}

func TestStatusBarModel_CondensedMode_NoWaiting(t *testing.T) {
	sb := NewStatusBarModel(80)
	sb.SetCounts(5, 3, 0) // No waiting projects
	sb.SetCondensed(true)

	view := sb.View()

	// Epic 4 Hotfix H1: Should show "0W" so users know feature exists
	if !strings.Contains(view, "0W") {
		t.Error("condensed view should show '0W' when count is 0 (Epic 4 Hotfix H1)")
	}
}

func TestStatusBarModel_NormalMode(t *testing.T) {
	sb := NewStatusBarModel(80)
	sb.SetCounts(5, 3, 2)
	sb.SetCondensed(false)

	view := sb.View()

	// Should be two lines
	if strings.Count(view, "\n") != 1 {
		t.Error("normal view should have exactly one newline (two lines)")
	}
}

// C1 FIX VERIFICATION: Condensed mode must preserve refresh features
func TestStatusBarModel_CondensedMode_ShowsRefreshSpinner(t *testing.T) {
	sb := NewStatusBarModel(80)
	sb.SetCounts(5, 3, 0)
	sb.SetCondensed(true)
	sb.SetRefreshing(true, 2, 5)

	view := sb.View()

	// Should show refresh progress even in condensed mode
	if !strings.Contains(view, "Refreshing") {
		t.Error("condensed view should show refresh spinner when refreshing")
	}
	if !strings.Contains(view, "2/5") {
		t.Error("condensed view should show refresh progress")
	}
	// Should include navigation shortcuts (code review fix)
	if !strings.Contains(view, "[j/k]") {
		t.Error("condensed view should include navigation shortcuts [j/k]")
	}
}

func TestStatusBarModel_CondensedMode_ShowsRefreshMessage(t *testing.T) {
	sb := NewStatusBarModel(80)
	sb.SetCounts(5, 3, 0)
	sb.SetCondensed(true)
	sb.SetRefreshComplete("Refreshed 3 projects")

	view := sb.View()

	// Should show refresh message even in condensed mode
	if !strings.Contains(view, "Refreshed 3 projects") {
		t.Error("condensed view should show refresh completion message")
	}
}

func TestStatusBarModel_SetCondensed(t *testing.T) {
	sb := NewStatusBarModel(80)

	// Initially not condensed
	if sb.isCondensed {
		t.Error("expected isCondensed to be false initially")
	}

	// Set to condensed
	sb.SetCondensed(true)
	if !sb.isCondensed {
		t.Error("expected isCondensed to be true after SetCondensed(true)")
	}

	// Set back to normal
	sb.SetCondensed(false)
	if sb.isCondensed {
		t.Error("expected isCondensed to be false after SetCondensed(false)")
	}
}

// Code review fix: Condensed mode must include navigation shortcuts
func TestStatusBarModel_CondensedMode_IncludesNavigationShortcuts(t *testing.T) {
	sb := NewStatusBarModel(80)
	sb.SetCounts(5, 3, 0)
	sb.SetCondensed(true)

	view := sb.View()

	// Should include navigation shortcuts even in condensed mode
	if !strings.Contains(view, "[j/k]") {
		t.Error("condensed view should include navigation shortcuts [j/k]")
	}
	if !strings.Contains(view, "[?]") {
		t.Error("condensed view should include help shortcut [?]")
	}
	if !strings.Contains(view, "[q]") {
		t.Error("condensed view should include quit shortcut [q]")
	}
}

// ============================================================================
// Story 4.5: CalculateCountsWithWaiting Tests
// ============================================================================

func TestCalculateCountsWithWaiting_SomeWaiting(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", State: domain.StateActive}, // waiting
		{ID: "2", State: domain.StateActive}, // not waiting
		{ID: "3", State: domain.StateActive}, // waiting
		{ID: "4", State: domain.StateHibernated},
	}

	// Checker that marks projects 1 and 3 as waiting
	checker := func(p *domain.Project) bool {
		return p.ID == "1" || p.ID == "3"
	}

	active, hibernated, waiting := CalculateCountsWithWaiting(projects, checker)

	if active != 3 {
		t.Errorf("expected active=3, got %d", active)
	}
	if hibernated != 1 {
		t.Errorf("expected hibernated=1, got %d", hibernated)
	}
	if waiting != 2 {
		t.Errorf("expected waiting=2 (projects 1 and 3), got %d", waiting)
	}
}

func TestCalculateCountsWithWaiting_NilChecker(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", State: domain.StateActive},
		{ID: "2", State: domain.StateActive},
	}

	// nil checker should behave like CalculateCounts (waiting=0)
	active, hibernated, waiting := CalculateCountsWithWaiting(projects, nil)

	if active != 2 {
		t.Errorf("expected active=2, got %d", active)
	}
	if hibernated != 0 {
		t.Errorf("expected hibernated=0, got %d", hibernated)
	}
	if waiting != 0 {
		t.Errorf("expected waiting=0 with nil checker, got %d", waiting)
	}
}

func TestCalculateCountsWithWaiting_AllWaiting(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", State: domain.StateActive},
		{ID: "2", State: domain.StateActive},
		{ID: "3", State: domain.StateActive},
	}

	// All active projects are waiting
	checker := func(p *domain.Project) bool { return true }

	active, hibernated, waiting := CalculateCountsWithWaiting(projects, checker)

	if active != 3 {
		t.Errorf("expected active=3, got %d", active)
	}
	if hibernated != 0 {
		t.Errorf("expected hibernated=0, got %d", hibernated)
	}
	if waiting != 3 {
		t.Errorf("expected waiting=3 (all active), got %d", waiting)
	}
}

func TestCalculateCountsWithWaiting_NoWaiting(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", State: domain.StateActive},
		{ID: "2", State: domain.StateActive},
	}

	// No projects are waiting
	checker := func(p *domain.Project) bool { return false }

	active, _, waiting := CalculateCountsWithWaiting(projects, checker)

	if active != 2 {
		t.Errorf("expected active=2, got %d", active)
	}
	if waiting != 0 {
		t.Errorf("expected waiting=0, got %d", waiting)
	}
}

func TestCalculateCountsWithWaiting_HibernatedNeverWaiting(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", State: domain.StateHibernated},
		{ID: "2", State: domain.StateHibernated},
	}

	// Checker always returns true, but hibernated projects shouldn't count
	checker := func(p *domain.Project) bool { return true }

	active, hibernated, waiting := CalculateCountsWithWaiting(projects, checker)

	if active != 0 {
		t.Errorf("expected active=0, got %d", active)
	}
	if hibernated != 2 {
		t.Errorf("expected hibernated=2, got %d", hibernated)
	}
	if waiting != 0 {
		t.Errorf("expected waiting=0 (hibernated never wait), got %d", waiting)
	}
}

func TestCalculateCountsWithWaiting_Empty(t *testing.T) {
	var projects []*domain.Project
	checker := func(p *domain.Project) bool { return true }

	active, hibernated, waiting := CalculateCountsWithWaiting(projects, checker)

	if active != 0 || hibernated != 0 || waiting != 0 {
		t.Errorf("expected all zeros for empty projects, got active=%d hibernated=%d waiting=%d", active, hibernated, waiting)
	}
}

func TestCalculateCounts_BackwardCompatible(t *testing.T) {
	// CalculateCounts should continue to work without waiting detection
	projects := []*domain.Project{
		{State: domain.StateActive},
		{State: domain.StateActive},
		{State: domain.StateHibernated},
	}

	active, hibernated, waiting := CalculateCounts(projects)

	if active != 2 {
		t.Errorf("expected active=2, got %d", active)
	}
	if hibernated != 1 {
		t.Errorf("expected hibernated=1, got %d", hibernated)
	}
	if waiting != 0 {
		t.Errorf("CalculateCounts should always return waiting=0, got %d", waiting)
	}
}

// ============================================================================
// Story 4.6: Watcher Warning Tests
// ============================================================================

func TestStatusBarModel_SetWatcherWarning(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)

	// Initially no warning
	view := sb.View()
	if strings.Contains(view, "unavailable") {
		t.Error("expected no warning initially")
	}

	// Set warning
	sb.SetWatcherWarning("⚠️ File watching unavailable")
	view = sb.View()
	if !strings.Contains(view, "File watching unavailable") {
		t.Errorf("expected warning to appear, got: %s", view)
	}
}

func TestStatusBarModel_WatcherWarning_CondensedMode(t *testing.T) {
	sb := NewStatusBarModel(80)
	sb.SetCounts(5, 3, 0)
	sb.SetCondensed(true)
	sb.SetWatcherWarning("⚠️ File watching unavailable")

	view := sb.View()

	// Should show abbreviated warning emoji in condensed mode
	if !strings.Contains(view, "⚠️") {
		t.Error("condensed view should show warning emoji")
	}
}

func TestStatusBarModel_WatcherWarning_ClearWarning(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)
	sb.SetWatcherWarning("⚠️ File watching unavailable")

	// Clear warning
	sb.SetWatcherWarning("")
	view := sb.View()
	if strings.Contains(view, "unavailable") {
		t.Error("expected warning to be cleared")
	}
}

func TestStatusBarModel_WatcherWarning_WithRefreshMessage(t *testing.T) {
	sb := NewStatusBarModel(100)
	sb.SetCounts(5, 2, 0)
	sb.SetRefreshComplete("Refreshed 5 projects")
	sb.SetWatcherWarning("⚠️ File watching unavailable")

	view := sb.View()

	// Both messages should appear
	if !strings.Contains(view, "Refreshed 5 projects") {
		t.Error("expected refresh message to appear")
	}
	if !strings.Contains(view, "File watching unavailable") {
		t.Error("expected watcher warning to appear")
	}
}
