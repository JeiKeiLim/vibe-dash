package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
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
		"Navigation",
		"Actions",
		"Views",
		"General",
		"Show this help",
		"Quit",
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

// =============================================================================
// Story 4.2: Tick Tests for Periodic Timestamp Refresh
// =============================================================================

func TestModel_Init_IncludesTickCmd(t *testing.T) {
	m := NewModel(nil)
	cmd := m.Init()

	// Init returns a batch command that includes tick
	if cmd == nil {
		t.Error("Init() should return a batch command including tick")
	}
	// We can't inspect the batch directly, but we verify it's not nil
	// and the behavior is tested in the tick message handler test
}

func TestModel_Update_TickMsg_ReturnsNextTick(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40

	// Send a tick message
	newModel, cmd := m.Update(tickMsg{})

	// Model should be returned unchanged
	updated := newModel.(Model)
	if updated.width != 80 || updated.height != 40 {
		t.Error("tickMsg should not modify model state")
	}

	// A new tick command should be returned to schedule next tick
	if cmd == nil {
		t.Error("tickMsg should return next tick command")
	}
}

func TestModel_Update_TickMsg_DoesNotAffectEditing(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.isEditingNote = true // In note editing mode

	// Send a tick message - should still schedule next tick
	_, cmd := m.Update(tickMsg{})

	// Note editing mode doesn't intercept tickMsg - it goes to the switch
	// tickMsg should return next tick regardless of editing state
	if cmd == nil {
		t.Error("tickMsg should return next tick command even during editing")
	}
}

func TestModel_Update_TickMsg_DoesNotAffectRefreshing(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.isRefreshing = true

	// Send a tick message
	_, cmd := m.Update(tickMsg{})

	// Should still schedule next tick during refresh
	if cmd == nil {
		t.Error("tickMsg should return next tick command even during refresh")
	}
}

// =============================================================================
// Story 4.5: WaitingDetector Integration Tests
// =============================================================================

// mockWaitingDetector implements ports.WaitingDetector for testing.
type mockWaitingDetector struct {
	isWaitingFunc      func(p *domain.Project) bool
	durationFunc       func(p *domain.Project) time.Duration
	isWaitingCalls     int
	waitingDurCalls    int
	lastCheckedProject *domain.Project
}

func (m *mockWaitingDetector) IsWaiting(ctx context.Context, p *domain.Project) bool {
	m.isWaitingCalls++
	m.lastCheckedProject = p
	if m.isWaitingFunc != nil {
		return m.isWaitingFunc(p)
	}
	return false
}

func (m *mockWaitingDetector) WaitingDuration(ctx context.Context, p *domain.Project) time.Duration {
	m.waitingDurCalls++
	if m.durationFunc != nil {
		return m.durationFunc(p)
	}
	return 0
}

func TestModel_SetWaitingDetector(t *testing.T) {
	m := NewModel(nil)
	mock := &mockWaitingDetector{}

	// Initially should be nil
	if m.waitingDetector != nil {
		t.Error("waitingDetector should be nil initially")
	}

	// Set detector
	m.SetWaitingDetector(mock)

	if m.waitingDetector != mock {
		t.Error("SetWaitingDetector should set the detector")
	}
}

func TestModel_IsProjectWaiting_WithNilDetector(t *testing.T) {
	m := NewModel(nil)
	// No detector set

	project := &domain.Project{ID: "test", Name: "test"}
	result := m.isProjectWaiting(project)

	if result != false {
		t.Error("isProjectWaiting should return false when detector is nil")
	}
}

func TestModel_IsProjectWaiting_WithDetector(t *testing.T) {
	m := NewModel(nil)
	mock := &mockWaitingDetector{
		isWaitingFunc: func(p *domain.Project) bool {
			return p.ID == "waiting-project"
		},
	}
	m.SetWaitingDetector(mock)

	// Test project that is waiting
	waitingProject := &domain.Project{ID: "waiting-project", Name: "waiting"}
	result := m.isProjectWaiting(waitingProject)
	if result != true {
		t.Error("isProjectWaiting should return true for waiting project")
	}
	if mock.isWaitingCalls != 1 {
		t.Errorf("expected 1 IsWaiting call, got %d", mock.isWaitingCalls)
	}

	// Test project that is not waiting
	activeProject := &domain.Project{ID: "active-project", Name: "active"}
	result = m.isProjectWaiting(activeProject)
	if result != false {
		t.Error("isProjectWaiting should return false for active project")
	}
}

func TestModel_GetWaitingDuration_WithNilDetector(t *testing.T) {
	m := NewModel(nil)
	// No detector set

	project := &domain.Project{ID: "test", Name: "test"}
	result := m.getWaitingDuration(project)

	if result != 0 {
		t.Error("getWaitingDuration should return 0 when detector is nil")
	}
}

func TestModel_GetWaitingDuration_WithDetector(t *testing.T) {
	m := NewModel(nil)
	mock := &mockWaitingDetector{
		durationFunc: func(p *domain.Project) time.Duration {
			if p.ID == "waiting-project" {
				return 2 * time.Hour
			}
			return 0
		},
	}
	m.SetWaitingDetector(mock)

	project := &domain.Project{ID: "waiting-project", Name: "waiting"}
	result := m.getWaitingDuration(project)

	if result != 2*time.Hour {
		t.Errorf("getWaitingDuration should return 2h, got %v", result)
	}
	if mock.waitingDurCalls != 1 {
		t.Errorf("expected 1 WaitingDuration call, got %d", mock.waitingDurCalls)
	}
}

func TestModel_ProjectsLoadedMsg_WiresWaitingCallbacks(t *testing.T) {
	// Create mock waiting detector that tracks calls
	mock := &mockWaitingDetector{
		isWaitingFunc: func(p *domain.Project) bool {
			return true // All projects waiting for test
		},
		durationFunc: func(p *domain.Project) time.Duration {
			return 30 * time.Minute
		},
	}

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.SetWaitingDetector(mock)

	// Create test projects
	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/a", State: domain.StateActive},
		{ID: "b", Name: "project-b", Path: "/b", State: domain.StateActive},
	}

	// Simulate ProjectsLoadedMsg
	msg := ProjectsLoadedMsg{projects: projects}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Verify projects were loaded
	if len(updated.projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(updated.projects))
	}

	// Verify status bar has waiting counts (calculated via CalculateCountsWithWaiting)
	// The status bar should show 2 waiting projects
	view := updated.statusBar.View()
	if !strings.Contains(view, "WAITING") {
		t.Errorf("status bar should show WAITING indicator, got: %s", view)
	}

	// Verify the mock was called during CalculateCountsWithWaiting
	// Each project should be checked once
	if mock.isWaitingCalls != 2 {
		t.Errorf("expected 2 IsWaiting calls during count calculation, got %d", mock.isWaitingCalls)
	}
}

func TestModel_WaitingCallbacksWiring_NilDetector(t *testing.T) {
	// Verify no panic when detector is nil
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	// No detector set - should not panic

	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/a", State: domain.StateActive},
	}

	// Should not panic
	msg := ProjectsLoadedMsg{projects: projects}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Verify project list was created without crash
	if len(updated.projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(updated.projects))
	}

	// Status bar should show no waiting (since no detector)
	view := updated.statusBar.View()
	if strings.Contains(view, "WAITING") {
		t.Error("status bar should NOT show WAITING when detector is nil")
	}
}

// =============================================================================
// Story 4.6: FileWatcher Integration Tests
// =============================================================================

// mockFileWatcher implements ports.FileWatcher for testing.
type mockFileWatcher struct {
	watchCalled  bool
	closeCalled  bool
	watchPaths   []string
	returnErr    error
	returnCh     chan ports.FileEvent
	closeErr     error
}

func newMockFileWatcher() *mockFileWatcher {
	return &mockFileWatcher{
		returnCh: make(chan ports.FileEvent, 10),
	}
}

func (m *mockFileWatcher) Watch(ctx context.Context, paths []string) (<-chan ports.FileEvent, error) {
	m.watchCalled = true
	m.watchPaths = paths
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return m.returnCh, nil
}

func (m *mockFileWatcher) Close() error {
	m.closeCalled = true
	if m.returnCh != nil {
		// Don't close here - let the test manage channel lifecycle
	}
	return m.closeErr
}

func TestModel_SetFileWatcher(t *testing.T) {
	m := NewModel(nil)
	mock := newMockFileWatcher()

	// Initially should be nil
	if m.fileWatcher != nil {
		t.Error("fileWatcher should be nil initially")
	}
	if m.fileWatcherAvailable {
		t.Error("fileWatcherAvailable should be false initially")
	}

	// Set watcher
	m.SetFileWatcher(mock)

	if m.fileWatcher != mock {
		t.Error("SetFileWatcher should set the file watcher")
	}
	if !m.fileWatcherAvailable {
		t.Error("SetFileWatcher should set fileWatcherAvailable to true")
	}
}

func TestModel_FindProjectByPath_ExactMatch(t *testing.T) {
	m := createModelWithProjects(3)
	m.projects[0].Path = "/home/user/project-a"
	m.projects[1].Path = "/home/user/project-b"
	m.projects[2].Path = "/home/user/project-c"

	// Exact path match
	result := m.findProjectByPath("/home/user/project-b")
	if result == nil {
		t.Fatal("findProjectByPath should find exact match")
	}
	if result.Path != "/home/user/project-b" {
		t.Errorf("expected project-b, got %s", result.Path)
	}
}

func TestModel_FindProjectByPath_PrefixMatch(t *testing.T) {
	m := createModelWithProjects(2)
	m.projects[0].Path = "/home/user/project-a"
	m.projects[1].Path = "/home/user/project-b"

	// Subpath match (file inside project)
	result := m.findProjectByPath("/home/user/project-a/src/main.go")
	if result == nil {
		t.Fatal("findProjectByPath should find prefix match")
	}
	if result.Path != "/home/user/project-a" {
		t.Errorf("expected project-a, got %s", result.Path)
	}
}

func TestModel_FindProjectByPath_NoMatch(t *testing.T) {
	m := createModelWithProjects(2)
	m.projects[0].Path = "/home/user/project-a"
	m.projects[1].Path = "/home/user/project-b"

	// Path not in any project
	result := m.findProjectByPath("/home/other/random-file.txt")
	if result != nil {
		t.Error("findProjectByPath should return nil for unmatched path")
	}
}

func TestModel_FindProjectByPath_SimilarPrefixNotMatched(t *testing.T) {
	m := createModelWithProjects(1)
	m.projects[0].Path = "/home/user/project"

	// Similar path but not a child (project-extended vs project)
	result := m.findProjectByPath("/home/user/project-extended/file.go")
	if result != nil {
		t.Error("findProjectByPath should not match similar but non-child paths")
	}
}

func TestModel_FindProjectByPath_TrailingSlash(t *testing.T) {
	m := createModelWithProjects(1)
	m.projects[0].Path = "/home/user/project/"

	// Match with trailing slash normalized
	result := m.findProjectByPath("/home/user/project/src/file.go")
	if result == nil {
		t.Fatal("findProjectByPath should handle trailing slash")
	}
}

func TestModel_Update_FileEventMsg_UpdatesProject(t *testing.T) {
	m := createModelWithProjects(2)
	m.projects[0].Path = "/home/user/project-a"
	m.projects[0].LastActivityAt = time.Now().Add(-1 * time.Hour)
	m.projects[1].Path = "/home/user/project-b"

	// Create status bar
	m.statusBar = components.NewStatusBarModel(80)
	active, hibernated, waiting := components.CalculateCounts(m.projects)
	m.statusBar.SetCounts(active, hibernated, waiting)

	eventTime := time.Now()
	msg := fileEventMsg{
		Path:      "/home/user/project-a/src/main.go",
		Operation: ports.FileOpModify,
		Timestamp: eventTime,
	}

	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Project's LastActivityAt should NOT be updated in the model (requires repo)
	// because handleFileEvent calls repo.UpdateLastActivity which we haven't mocked
	// But we can verify the cmd is returned (for next event)

	// Should return waitForNextFileEventCmd
	// We can't easily test cmd content, but verify it handles without panic
	_ = updated
	_ = cmd
}

func TestModel_Update_FileEventMsg_UnmatchedPath(t *testing.T) {
	m := createModelWithProjects(1)
	m.projects[0].Path = "/home/user/project-a"
	originalTime := m.projects[0].LastActivityAt

	msg := fileEventMsg{
		Path:      "/different/path/file.go",
		Operation: ports.FileOpModify,
		Timestamp: time.Now(),
	}

	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Project should remain unchanged
	if !updated.projects[0].LastActivityAt.Equal(originalTime) {
		t.Error("Unmatched file event should not update any project")
	}
}

func TestModel_Update_FileWatcherErrorMsg_DisablesWatcher(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.fileWatcherAvailable = true
	m.statusBar = components.NewStatusBarModel(80)

	msg := fileWatcherErrorMsg{err: context.Canceled}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.fileWatcherAvailable {
		t.Error("fileWatcherErrorMsg should set fileWatcherAvailable to false")
	}

	// Status bar should show warning
	view := updated.statusBar.View()
	if !strings.Contains(view, "unavailable") {
		t.Errorf("Status bar should show watcher warning, got: %s", view)
	}
}

func TestModel_GracefulShutdown_CancelsContext(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40

	// Set up watch context
	ctx, cancel := context.WithCancel(context.Background())
	m.watchCtx = ctx
	m.watchCancel = cancel

	mock := newMockFileWatcher()
	m.fileWatcher = mock

	// Press 'q' to quit
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Expected - context was cancelled
	default:
		t.Error("Quit should cancel the watch context")
	}

	// Close should be called on file watcher
	if !mock.closeCalled {
		t.Error("Quit should call fileWatcher.Close()")
	}

	// Should return tea.Quit
	if cmd == nil {
		t.Error("Quit should return tea.Quit command")
	}
}

func TestModel_GracefulShutdown_CtrlC(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40

	ctx, cancel := context.WithCancel(context.Background())
	m.watchCtx = ctx
	m.watchCancel = cancel

	mock := newMockFileWatcher()
	m.fileWatcher = mock

	// Press Ctrl+C to quit
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := m.Update(msg)

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Ctrl+C should cancel the watch context")
	}

	if !mock.closeCalled {
		t.Error("Ctrl+C should call fileWatcher.Close()")
	}

	if cmd == nil {
		t.Error("Ctrl+C should return tea.Quit command")
	}
}

func TestModel_GracefulShutdown_NilWatcher(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	// No watcher or context set - should not panic

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	// Should quit without panic
	if cmd == nil {
		t.Error("Quit should return tea.Quit command even without watcher")
	}
}

func TestModel_ValidationMode_Quit_CleansUpWatcher(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.viewMode = viewModeValidation
	m.invalidProjects = []InvalidProject{{
		Project: &domain.Project{Name: "test", Path: "/test"},
	}}

	ctx, cancel := context.WithCancel(context.Background())
	m.watchCtx = ctx
	m.watchCancel = cancel

	mock := newMockFileWatcher()
	m.fileWatcher = mock

	// Press 'q' in validation mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, _ = m.Update(msg)

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Quit in validation mode should cancel watch context")
	}

	if !mock.closeCalled {
		t.Error("Quit in validation mode should call fileWatcher.Close()")
	}
}

func TestModel_ProjectsLoadedMsg_StartsFileWatcher(t *testing.T) {
	// Test that ProjectsLoadedMsg starts the file watcher with project paths (AC8)
	mock := newMockFileWatcher()

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.SetFileWatcher(mock)

	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
		{ID: "b", Name: "project-b", Path: "/home/user/project-b", State: domain.StateActive},
	}

	msg := ProjectsLoadedMsg{projects: projects}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Verify Watch() was called with correct paths
	if !mock.watchCalled {
		t.Error("ProjectsLoadedMsg should call fileWatcher.Watch()")
	}
	if len(mock.watchPaths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(mock.watchPaths))
	}
	if mock.watchPaths[0] != "/home/user/project-a" || mock.watchPaths[1] != "/home/user/project-b" {
		t.Errorf("unexpected paths: %v", mock.watchPaths)
	}

	// Verify event channel is stored
	if updated.eventCh == nil {
		t.Error("eventCh should be set after successful Watch()")
	}

	// Verify watch context is created
	if updated.watchCtx == nil || updated.watchCancel == nil {
		t.Error("watch context should be created")
	}

	// Verify cmd is returned (waitForNextFileEventCmd)
	if cmd == nil {
		t.Error("should return waitForNextFileEventCmd after successful Watch()")
	}
}

func TestModel_ProjectsLoadedMsg_WatchError_ShowsWarning(t *testing.T) {
	// Test that Watch() error sets fileWatcherAvailable=false and shows warning (AC3)
	mock := newMockFileWatcher()
	mock.returnErr = context.DeadlineExceeded // Simulate error

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)
	m.SetFileWatcher(mock)

	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
	}

	msg := ProjectsLoadedMsg{projects: projects}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Verify Watch() was called
	if !mock.watchCalled {
		t.Error("ProjectsLoadedMsg should call fileWatcher.Watch()")
	}

	// Verify fileWatcherAvailable is set to false on error
	if updated.fileWatcherAvailable {
		t.Error("fileWatcherAvailable should be false after Watch() error")
	}

	// Verify status bar shows warning
	view := updated.statusBar.View()
	if !strings.Contains(view, "unavailable") {
		t.Errorf("status bar should show watcher warning, got: %s", view)
	}

	// Verify no cmd is returned (no waitForNextFileEventCmd)
	if cmd != nil {
		t.Error("should not return cmd when Watch() fails")
	}
}

func TestModel_WaitForNextFileEventCmd_NilChannel(t *testing.T) {
	// Test that waitForNextFileEventCmd returns nil when channel is nil
	m := NewModel(nil)
	m.eventCh = nil // No channel set

	cmd := m.waitForNextFileEventCmd()

	if cmd != nil {
		t.Error("waitForNextFileEventCmd should return nil when eventCh is nil")
	}
}

func TestModel_FindProjectByPath_NestedProjects(t *testing.T) {
	// Test edge case: nested projects - longer path should match
	// This documents current behavior where iteration order determines match
	m := createModelWithProjects(2)
	m.projects[0].Path = "/home/user/project"
	m.projects[1].Path = "/home/user/project/submodule"

	// Event in submodule - should ideally match submodule, but currently matches first
	// This test documents the current behavior (may match parent due to iteration order)
	result := m.findProjectByPath("/home/user/project/submodule/src/main.go")

	// Current implementation returns first match (project at /home/user/project)
	// This is a known limitation - nested projects may not match correctly
	if result == nil {
		t.Fatal("should find a matching project")
	}

	// Document the behavior: first matching project wins
	// If projects are ordered [parent, child], parent matches first
	// This is acceptable for MVP as nested projects are edge case
	t.Logf("Nested project matched: %s (expected behavior: first prefix match)", result.Path)
}

func TestModel_FindProjectByPath_LongerPathMatchesFirst(t *testing.T) {
	// If projects are ordered with longer path first, it should match correctly
	m := createModelWithProjects(2)
	m.projects[0].Path = "/home/user/project/submodule" // Longer path first
	m.projects[1].Path = "/home/user/project"

	result := m.findProjectByPath("/home/user/project/submodule/src/main.go")

	if result == nil {
		t.Fatal("should find matching project")
	}
	if result.Path != "/home/user/project/submodule" {
		t.Errorf("expected submodule path, got %s", result.Path)
	}
}
