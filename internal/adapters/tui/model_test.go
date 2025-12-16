package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestNewModel(t *testing.T) {
	m := NewModel(nil)

	if m.ready {
		t.Error("NewModel().ready should be false")
	}
	if m.showHelp {
		t.Error("NewModel().showHelp should be false")
	}
	if m.width != 0 {
		t.Errorf("NewModel().width should be 0, got %d", m.width)
	}
	if m.height != 0 {
		t.Errorf("NewModel().height should be 0, got %d", m.height)
	}
	if m.viewMode != viewModeNormal {
		t.Error("NewModel().viewMode should be viewModeNormal")
	}
}

func TestModel_Init(t *testing.T) {
	m := NewModel(nil)
	cmd := m.Init()

	// Init now returns a validation command (even with nil repo)
	if cmd == nil {
		t.Error("Init() should return a validation command")
	}
}

func TestModel_Update_QuitKey(t *testing.T) {
	m := NewModel(nil)
	m.ready = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	// Check if cmd is tea.Quit
	if cmd == nil {
		t.Error("'q' key should return tea.Quit command")
	}
}

func TestModel_Update_CtrlC(t *testing.T) {
	m := NewModel(nil)
	m.ready = true

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Ctrl+C should return tea.Quit command")
	}
}

func TestModel_Update_HelpToggle(t *testing.T) {
	m := NewModel(nil)
	m.ready = true

	// Press '?' to show help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if !updated.showHelp {
		t.Error("'?' key should toggle showHelp to true")
	}

	// Press '?' again to hide help
	newModel2, _ := updated.Update(msg)
	updated2 := newModel2.(Model)

	if updated2.showHelp {
		t.Error("'?' key should toggle showHelp to false")
	}
}

func TestModel_Update_HelpCloseOnAnyKey(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.showHelp = true // Help is showing

	// Press any key (not '?') to close help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.showHelp {
		t.Error("Any key should close help overlay")
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	m := NewModel(nil)

	// Send WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Should have pending resize (debounce)
	if !updated.hasPendingResize {
		t.Error("WindowSizeMsg should set hasPendingResize to true")
	}
	if updated.pendingWidth != 80 || updated.pendingHeight != 24 {
		t.Errorf("WindowSizeMsg should set pending dimensions to 80x24, got %dx%d", updated.pendingWidth, updated.pendingHeight)
	}
	if cmd == nil {
		t.Error("WindowSizeMsg should return tick command for debounce")
	}

	// Simulate debounce tick
	newModel2, _ := updated.Update(resizeTickMsg{})
	updated2 := newModel2.(Model)

	if updated2.width != 80 || updated2.height != 24 {
		t.Errorf("After tick, dimensions should be 80x24, got %dx%d", updated2.width, updated2.height)
	}
	if !updated2.ready {
		t.Error("ready should be true after WindowSizeMsg + tick")
	}
	if updated2.hasPendingResize {
		t.Error("hasPendingResize should be false after tick")
	}
}

func TestModel_Update_ResizeTickWithNoPending(t *testing.T) {
	// Edge case: resizeTickMsg arrives when no resize is pending
	// This can happen with rapid resize events causing multiple ticks
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 24
	m.hasPendingResize = false // No pending resize

	// Send tick without pending resize
	newModel, cmd := m.Update(resizeTickMsg{})
	updated := newModel.(Model)

	// Should not change dimensions
	if updated.width != 80 || updated.height != 24 {
		t.Errorf("Dimensions should remain 80x24, got %dx%d", updated.width, updated.height)
	}
	// Should return nil command
	if cmd != nil {
		t.Error("resizeTickMsg with no pending resize should return nil command")
	}
}

func TestModel_View_NotReady(t *testing.T) {
	m := NewModel(nil)
	m.ready = false

	view := m.View()

	if view != "Initializing..." {
		t.Errorf("View when not ready should be 'Initializing...', got %q", view)
	}
}

func TestModel_View_TooSmall(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 50  // Less than MinWidth (60)
	m.height = 15 // Less than MinHeight (20)

	view := m.View()

	if !strings.Contains(view, "Terminal too small") {
		t.Error("View should show 'Terminal too small' message")
	}
	if !strings.Contains(view, "60x20") {
		t.Error("View should show minimum required dimensions")
	}
}

func TestModel_View_EmptyView(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 24

	view := m.View()

	// Check for expected content
	expectedStrings := []string{
		"VIBE DASHBOARD",
		"Welcome to Vibe Dashboard",
		"vibe add",
		"[?] for help",
		"[q] to quit",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(view, s) {
			t.Errorf("EmptyView missing: %q", s)
		}
	}
}

func TestModel_View_EmptyViewWithStatusBar(t *testing.T) {
	// AC2: Status bar should always be visible, even in empty view
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 24
	m.statusBar = components.NewStatusBarModel(m.width)

	view := m.View()

	// Status bar should be present in empty view (AC2)
	if !strings.Contains(view, "active") {
		t.Error("Empty view should include status bar with 'active' count")
	}
	if !strings.Contains(view, "hibernated") {
		t.Error("Empty view should include status bar with 'hibernated' count")
	}
	// Should have shortcuts line
	if !strings.Contains(view, "[j/k]") {
		t.Error("Empty view should include status bar shortcuts")
	}
}

func TestModel_View_HelpOverlay(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 24
	m.showHelp = true

	view := m.View()

	expectedStrings := []string{
		"KEYBOARD SHORTCUTS",
		"Toggle this help",
		"Quit",
		"Force quit",
		"Press any key to close",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(view, s) {
			t.Errorf("HelpOverlay missing: %q", s)
		}
	}
}

func TestDefaultKeyBindings(t *testing.T) {
	kb := DefaultKeyBindings()

	if kb.Quit != "q" {
		t.Errorf("Expected Quit to be 'q', got %q", kb.Quit)
	}
	if kb.ForceQuit != "ctrl+c" {
		t.Errorf("Expected ForceQuit to be 'ctrl+c', got %q", kb.ForceQuit)
	}
	if kb.Help != "?" {
		t.Errorf("Expected Help to be '?', got %q", kb.Help)
	}
	if kb.Escape != "esc" {
		t.Errorf("Expected Escape to be 'esc', got %q", kb.Escape)
	}
	if kb.Detail != "d" {
		t.Errorf("Expected Detail to be 'd', got %q", kb.Detail)
	}
}

func TestUseColorLogic(t *testing.T) {
	// Test the UseColor logic (NO_COLOR="" && TERM!="dumb" => true)
	// Note: We can't modify the actual UseColor variable since it's set at init time,
	// but we can verify the logic by testing the same condition
	testCases := []struct {
		noColor  string
		term     string
		expected bool
	}{
		{"", "xterm", true},          // Default: colors enabled
		{"", "xterm-256color", true}, // 256-color terminal
		{"1", "xterm", false},        // NO_COLOR set
		{"true", "xterm", false},     // NO_COLOR set (any value)
		{"", "dumb", false},          // Dumb terminal
		{"1", "dumb", false},         // Both NO_COLOR and dumb
	}

	for _, tc := range testCases {
		// Simulate the UseColor logic
		result := tc.noColor == "" && tc.term != "dumb"
		if result != tc.expected {
			t.Errorf("UseColor logic for NO_COLOR=%q, TERM=%q: expected %v, got %v",
				tc.noColor, tc.term, tc.expected, result)
		}
	}
}

func TestMinDimensions(t *testing.T) {
	if MinWidth != 60 {
		t.Errorf("MinWidth should be 60, got %d", MinWidth)
	}
	if MinHeight != 20 {
		t.Errorf("MinHeight should be 20, got %d", MinHeight)
	}
}

func TestModel_View_ValidationMode(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 24
	m.viewMode = viewModeValidation
	m.invalidProjects = []InvalidProject{
		{
			Project: &domain.Project{
				Name: "test-invalid-project",
				Path: "/nonexistent/path/to/project",
			},
			Error: domain.ErrPathNotAccessible,
		},
	}
	m.currentInvalidIdx = 0

	view := m.View()

	// Check for expected validation dialog content
	expectedStrings := []string{
		"Warning",
		"test-invalid-project",
		"/nonexistent/path/to/project",
		"[D] Delete",
		"[M] Move",
		"[K] Keep",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(view, s) {
			t.Errorf("Validation dialog missing: %q", s)
		}
	}
}

func TestModel_View_ValidationModeWithError(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 24
	m.viewMode = viewModeValidation
	m.invalidProjects = []InvalidProject{
		{
			Project: &domain.Project{
				Name: "test-project",
				Path: "/some/path",
			},
		},
	}
	m.currentInvalidIdx = 0
	m.validationError = "operation failed"

	view := m.View()

	// Check that error is displayed
	if !strings.Contains(view, "Error:") {
		t.Error("Validation dialog should display error message")
	}
	if !strings.Contains(view, "operation failed") {
		t.Error("Validation dialog should contain error text")
	}
}

// createModelWithProjects creates a test model with multiple projects for navigation tests.
func createModelWithProjects(count int) Model {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 24

	// Create test projects
	projects := make([]*domain.Project, count)
	for i := 0; i < count; i++ {
		projects[i] = &domain.Project{
			ID:   string(rune('a' + i)),
			Name: string(rune('a' + i)),
			Path: "/path/" + string(rune('a'+i)),
		}
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	return m
}

func TestModel_Navigation_JMovesDown(t *testing.T) {
	m := createModelWithProjects(3)

	// Initial selection should be 0
	if m.projectList.Index() != 0 {
		t.Errorf("Initial selection should be 0, got %d", m.projectList.Index())
	}

	// Press j to move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.projectList.Index() != 1 {
		t.Errorf("After 'j', selection should be 1, got %d", updated.projectList.Index())
	}
}

func TestModel_Navigation_KMovesUp(t *testing.T) {
	m := createModelWithProjects(3)

	// Move to second item first
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.projectList.Index() != 1 {
		t.Errorf("After 'j', selection should be 1, got %d", updated.projectList.Index())
	}

	// Press k to move up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = updated.Update(msg)
	updated = newModel.(Model)

	if updated.projectList.Index() != 0 {
		t.Errorf("After 'k', selection should be 0, got %d", updated.projectList.Index())
	}
}

func TestModel_Navigation_ArrowDownMovesDown(t *testing.T) {
	m := createModelWithProjects(3)

	// Initial selection should be 0
	if m.projectList.Index() != 0 {
		t.Errorf("Initial selection should be 0, got %d", m.projectList.Index())
	}

	// Press down arrow to move down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.projectList.Index() != 1 {
		t.Errorf("After down arrow, selection should be 1, got %d", updated.projectList.Index())
	}
}

func TestModel_Navigation_ArrowUpMovesUp(t *testing.T) {
	m := createModelWithProjects(3)

	// Move to second item first using down arrow
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.projectList.Index() != 1 {
		t.Errorf("After down arrow, selection should be 1, got %d", updated.projectList.Index())
	}

	// Press up arrow to move up
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = updated.Update(msg)
	updated = newModel.(Model)

	if updated.projectList.Index() != 0 {
		t.Errorf("After up arrow, selection should be 0, got %d", updated.projectList.Index())
	}
}

func TestModel_Navigation_BoundaryBehavior(t *testing.T) {
	m := createModelWithProjects(3)

	// Test at first item: pressing 'k' should stay at first (no wrap)
	if m.projectList.Index() != 0 {
		t.Errorf("Initial selection should be 0, got %d", m.projectList.Index())
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.projectList.Index() != 0 {
		t.Errorf("At first item, 'k' should stay at 0, got %d", updated.projectList.Index())
	}

	// Move to last item
	for i := 0; i < 2; i++ {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		newModel, _ = updated.Update(msg)
		updated = newModel.(Model)
	}

	if updated.projectList.Index() != 2 {
		t.Errorf("After moving to last, selection should be 2, got %d", updated.projectList.Index())
	}

	// Test at last item: pressing 'j' should stay at last (no wrap)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ = updated.Update(msg)
	updated = newModel.(Model)

	if updated.projectList.Index() != 2 {
		t.Errorf("At last item, 'j' should stay at 2, got %d", updated.projectList.Index())
	}
}

func TestModel_Escape_NormalMode(t *testing.T) {
	m := createModelWithProjects(3)

	// Press Esc in normal mode - should be no-op (return nil cmd)
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Model state should be unchanged
	if updated.showHelp {
		t.Error("Esc should not toggle help")
	}
	if updated.projectList.Index() != m.projectList.Index() {
		t.Error("Esc should not change selection")
	}
	// Cmd should be nil (no-op)
	if cmd != nil {
		t.Error("Esc in normal mode should return nil command")
	}
}

func TestModel_Escape_WhileHelpShowing(t *testing.T) {
	m := createModelWithProjects(3)
	m.showHelp = true // Help is showing

	// Press Esc to close help
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Help should be closed (any key closes help, including Esc)
	if updated.showHelp {
		t.Error("Esc should close help overlay")
	}
}

func TestModel_SelectionPersistence(t *testing.T) {
	m := createModelWithProjects(5)

	// Move to third item
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	for i := 0; i < 2; i++ {
		newModel, _ := m.Update(msg)
		m = newModel.(Model)
	}

	if m.projectList.Index() != 2 {
		t.Errorf("Selection should be at index 2, got %d", m.projectList.Index())
	}

	// Simulate window resize (which might affect state)
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Process resize tick
	newModel, _ = m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// Selection should persist
	if m.projectList.Index() != 2 {
		t.Errorf("Selection should persist after resize at index 2, got %d", m.projectList.Index())
	}
}

// ============================================================================
// Detail Panel Tests (Story 3.3)
// ============================================================================

func TestModel_DetailPanelToggle(t *testing.T) {
	m := createModelWithProjects(3)
	m.showDetailPanel = false

	// Press 'd' to show detail panel
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if !updated.showDetailPanel {
		t.Error("Detail panel should be visible after pressing 'd'")
	}
	if !updated.detailPanel.IsVisible() {
		t.Error("Detail panel component should be visible after pressing 'd'")
	}

	// Toggle again
	newModel, _ = updated.Update(msg)
	updated = newModel.(Model)

	if updated.showDetailPanel {
		t.Error("Detail panel should be hidden after pressing 'd' again")
	}
	if updated.detailPanel.IsVisible() {
		t.Error("Detail panel component should be hidden after pressing 'd' again")
	}
}

func TestModel_DetailPanelDefaultState_Height29(t *testing.T) {
	m := NewModel(nil)

	// Simulate window size with height 29 (< 30)
	m.hasPendingResize = true
	m.pendingWidth = 80
	m.pendingHeight = 29

	newModel, _ := m.Update(resizeTickMsg{})
	updated := newModel.(Model)

	// Per AC4: height < 30 => panel closed by default
	if updated.showDetailPanel {
		t.Error("Detail panel should be closed by default when height < 30")
	}
}

func TestModel_DetailPanelDefaultState_Height30(t *testing.T) {
	m := NewModel(nil)

	// Simulate window size with height 30 (in 30-34 range)
	m.hasPendingResize = true
	m.pendingWidth = 80
	m.pendingHeight = 30

	newModel, _ := m.Update(resizeTickMsg{})
	updated := newModel.(Model)

	// Per AC6: height 30-34 => panel closed by default
	if updated.showDetailPanel {
		t.Error("Detail panel should be closed by default when height is 30")
	}
}

func TestModel_DetailPanelDefaultState_Height34(t *testing.T) {
	m := NewModel(nil)

	// Simulate window size with height 34 (in 30-34 range)
	m.hasPendingResize = true
	m.pendingWidth = 80
	m.pendingHeight = 34

	newModel, _ := m.Update(resizeTickMsg{})
	updated := newModel.(Model)

	// Per AC6: height 30-34 => panel closed by default
	if updated.showDetailPanel {
		t.Error("Detail panel should be closed by default when height is 34")
	}
}

func TestModel_DetailPanelDefaultState_Height35(t *testing.T) {
	m := NewModel(nil)

	// Simulate window size with height 35 (>= 35)
	m.hasPendingResize = true
	m.pendingWidth = 80
	m.pendingHeight = 35

	newModel, _ := m.Update(resizeTickMsg{})
	updated := newModel.(Model)

	// Per AC5: height >= 35 => panel open by default
	if !updated.showDetailPanel {
		t.Error("Detail panel should be open by default when height >= 35")
	}
}

func TestModel_DetailPanelDefaultState_Height50(t *testing.T) {
	m := NewModel(nil)

	// Simulate window size with height 50 (>= 35)
	m.hasPendingResize = true
	m.pendingWidth = 80
	m.pendingHeight = 50

	newModel, _ := m.Update(resizeTickMsg{})
	updated := newModel.(Model)

	// Per AC5: height >= 35 => panel open by default
	if !updated.showDetailPanel {
		t.Error("Detail panel should be open by default when height >= 35")
	}
}

func TestModel_DetailPanelUpdateOnSelection(t *testing.T) {
	m := createModelWithProjects(3)
	// Initialize detail panel
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	// Get initial selected project
	initial := m.projectList.SelectedProject()
	if initial == nil {
		t.Fatal("Initial selected project should not be nil")
	}

	// Press 'j' to move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Selection should have changed
	if updated.projectList.Index() != 1 {
		t.Errorf("Selection should move to index 1, got %d", updated.projectList.Index())
	}
}

func TestModel_DetailPanelHint_ShortTerminal(t *testing.T) {
	m := createModelWithProjects(3)
	m.height = 25 // Less than 30
	m.showDetailPanel = false

	view := m.renderDashboard()

	// Should contain hint for short terminals
	if !strings.Contains(view, "Press [d] for details") {
		t.Error("Short terminal should show 'Press [d] for details' hint")
	}
}

func TestModel_DetailPanelHint_NotShownWhenOpen(t *testing.T) {
	m := createModelWithProjects(3)
	m.height = 25 // Less than 30
	m.showDetailPanel = true
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	view := m.renderDashboard()

	// Should NOT contain hint when panel is open
	if strings.Contains(view, "Press [d] for details") {
		t.Error("Hint should not be shown when detail panel is open")
	}
}

func TestShouldShowDetailPanelByDefault(t *testing.T) {
	tests := []struct {
		height   int
		expected bool
	}{
		{25, false}, // < 35
		{29, false}, // < 35
		{30, false}, // < 35
		{34, false}, // < 35
		{35, true},  // >= 35
		{40, true},  // >= 35
		{50, true},  // >= 35
		{100, true}, // >= 35
	}

	for _, tt := range tests {
		result := shouldShowDetailPanelByDefault(tt.height)
		if result != tt.expected {
			t.Errorf("shouldShowDetailPanelByDefault(%d) = %v, want %v", tt.height, result, tt.expected)
		}
	}
}

func TestModel_DetailPanelSplitLayout(t *testing.T) {
	m := createModelWithProjects(3)
	m.height = 40 // Tall enough to show panel
	m.showDetailPanel = true
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	view := m.renderDashboard()

	// Split layout should include detail panel with DETAILS: title
	if !strings.Contains(view, "DETAILS:") {
		t.Error("Split layout should include detail panel with 'DETAILS:' title")
	}
}

// ============================================================================
// Status Bar Tests (Story 3.4)
// ============================================================================

func TestModel_StatusBarIntegration(t *testing.T) {
	m := createModelWithProjects(3)
	m.statusBar = components.NewStatusBarModel(m.width)
	active, hibernated, waiting := components.CalculateCounts(m.projects)
	m.statusBar.SetCounts(active, hibernated, waiting)

	view := m.renderDashboard()

	// Status bar should be part of dashboard
	if !strings.Contains(view, "active") {
		t.Error("Dashboard should include status bar with 'active' count")
	}
	if !strings.Contains(view, "hibernated") {
		t.Error("Dashboard should include status bar with 'hibernated' count")
	}
}

func TestModel_StatusBarCountsUpdate(t *testing.T) {
	// Create projects with mixed states
	projects := []*domain.Project{
		{ID: "a", Name: "a", Path: "/a", State: domain.StateActive},
		{ID: "b", Name: "b", Path: "/b", State: domain.StateActive},
		{ID: "c", Name: "c", Path: "/c", State: domain.StateHibernated},
	}

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(m.width)

	// Simulate ProjectsLoadedMsg
	msg := ProjectsLoadedMsg{projects: projects}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Check counts were updated
	view := updated.statusBar.View()
	if !strings.Contains(view, "2 active") {
		t.Errorf("Status bar should show '2 active', got: %s", view)
	}
	if !strings.Contains(view, "1 hibernated") {
		t.Errorf("Status bar should show '1 hibernated', got: %s", view)
	}
}

func TestModel_StatusBarHeightReservation(t *testing.T) {
	m := createModelWithProjects(3)
	m.height = 40
	m.statusBar = components.NewStatusBarModel(m.width)
	active, hibernated, waiting := components.CalculateCounts(m.projects)
	m.statusBar.SetCounts(active, hibernated, waiting)

	view := m.renderDashboard()
	lines := strings.Split(view, "\n")

	// Dashboard should have content area + 2 lines for status bar
	// The last 2 lines should be the status bar
	if len(lines) < 2 {
		t.Errorf("Dashboard should have at least 2 lines for status bar, got %d lines", len(lines))
	}

	// Last line should be shortcuts line (contains navigation keys)
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "[j/k]") && !strings.Contains(lastLine, "[q]") {
		t.Errorf("Last line should be shortcuts, got: %s", lastLine)
	}
}

func TestModel_StatusBarWidthUpdate(t *testing.T) {
	m := NewModel(nil)
	m.statusBar = components.NewStatusBarModel(100) // Initial wide

	// Simulate resize to narrow
	m.hasPendingResize = true
	m.pendingWidth = 60
	m.pendingHeight = 40

	newModel, _ := m.Update(resizeTickMsg{})
	updated := newModel.(Model)

	// Status bar should show abbreviated shortcuts at width < 80
	view := updated.statusBar.View()
	if strings.Contains(view, "[d] details") {
		t.Error("Status bar should use abbreviated shortcuts at width 60")
	}
}
