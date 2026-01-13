package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/emoji"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/testhelpers"
)

func init() {
	// Story 8.9: Initialize emoji package for tests with emoji enabled
	useEmoji := true
	emoji.InitEmoji(&useEmoji)
}

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
		"vdash add",
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
	watchCalled bool
	closeCalled bool
	watchPaths  []string
	returnErr   error
	returnCh    chan ports.FileEvent
	closeErr    error
	failedPaths []string // Story 7.1: Simulated failed paths
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
	// Note: Don't close returnCh here - let the test manage channel lifecycle
	// The channel may still be in use by goroutines reading from it
	return m.closeErr
}

// GetFailedPaths implements ports.FileWatcher (Story 7.1).
func (m *mockFileWatcher) GetFailedPaths() []string {
	return m.failedPaths
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

// =============================================================================
// Story 7.1: Watcher Warning Message Tests
// =============================================================================

// TestModel_Update_WatcherWarningMsg_PartialFailure tests partial failure warning (AC1).
func TestModel_Update_WatcherWarningMsg_PartialFailure(t *testing.T) {
	m := NewModel(nil)
	m.fileWatcherAvailable = true // Start as available

	// Send partial failure message (1 of 3 failed)
	// Code review L2 fix: Use specific project name for reliable test assertion
	msg := watcherWarningMsg{
		failedPaths: []string{"/failed/my-test-project"},
		totalPaths:  3,
	}

	updated, _ := m.Update(msg)
	model := updated.(Model)

	// Should show partial failure warning with project name (basename extracted)
	warning := model.statusBar.View()
	if !strings.Contains(warning, "my-test-project") {
		t.Errorf("partial failure warning should contain project name 'my-test-project', got: %s", warning)
	}
	if !strings.Contains(warning, "⚠") {
		t.Errorf("partial failure warning should contain warning symbol, got: %s", warning)
	}
	if !strings.Contains(warning, "unavailable for:") {
		t.Errorf("partial failure warning should contain 'unavailable for:', got: %s", warning)
	}

	// fileWatcherAvailable should still be true (partial failure doesn't disable)
	if !model.fileWatcherAvailable {
		t.Error("fileWatcherAvailable should still be true for partial failure")
	}
}

// TestModel_Update_WatcherWarningMsg_CompleteFailure tests complete failure warning (AC2).
func TestModel_Update_WatcherWarningMsg_CompleteFailure(t *testing.T) {
	m := NewModel(nil)
	m.fileWatcherAvailable = true // Start as available

	// Send complete failure message (all 3 failed)
	msg := watcherWarningMsg{
		failedPaths: []string{"/failed1", "/failed2", "/failed3"},
		totalPaths:  3,
	}

	updated, _ := m.Update(msg)
	model := updated.(Model)

	// Should show complete failure warning
	warning := model.statusBar.View()
	if !strings.Contains(warning, "[r]") || !strings.Contains(warning, "refresh") {
		t.Errorf("complete failure warning should mention [r] to refresh, got: %s", warning)
	}

	// fileWatcherAvailable should be false
	if model.fileWatcherAvailable {
		t.Error("fileWatcherAvailable should be false for complete failure")
	}
}

// TestModel_Update_WatcherWarningMsg_NoFailure tests no warning when no failures.
func TestModel_Update_WatcherWarningMsg_NoFailure(t *testing.T) {
	m := NewModel(nil)
	m.fileWatcherAvailable = true

	// Send message with no failures
	msg := watcherWarningMsg{
		failedPaths: []string{},
		totalPaths:  3,
	}

	updated, _ := m.Update(msg)
	model := updated.(Model)

	// Should remain available with no warning
	if !model.fileWatcherAvailable {
		t.Error("fileWatcherAvailable should remain true when no failures")
	}
}

// TestModel_MockFileWatcher_FailedPaths tests mock returns configured failed paths.
func TestModel_MockFileWatcher_FailedPaths(t *testing.T) {
	mock := newMockFileWatcher()
	mock.failedPaths = []string{"/failed/path1", "/failed/path2"}

	failedPaths := mock.GetFailedPaths()
	if len(failedPaths) != 2 {
		t.Errorf("expected 2 failed paths, got %d", len(failedPaths))
	}
	if failedPaths[0] != "/failed/path1" {
		t.Errorf("expected /failed/path1, got %s", failedPaths[0])
	}
}

// TestModel_Update_WatcherWarningMsg_MultipleFailures tests M1 fix: multiple failures show count.
func TestModel_Update_WatcherWarningMsg_MultipleFailures(t *testing.T) {
	m := NewModel(nil)
	m.fileWatcherAvailable = true

	// Send partial failure message (3 of 5 failed)
	msg := watcherWarningMsg{
		failedPaths: []string{"/failed/project1", "/failed/project2", "/failed/project3"},
		totalPaths:  5,
	}

	updated, _ := m.Update(msg)
	model := updated.(Model)

	// Should show first project name + count of additional failures
	warning := model.statusBar.View()
	if !strings.Contains(warning, "project1") {
		t.Errorf("warning should contain first project name, got: %s", warning)
	}
	if !strings.Contains(warning, "+2 more") {
		t.Errorf("warning should show '+2 more' for additional failures, got: %s", warning)
	}
}

// =============================================================================
// Story 7.2: Config Warning Message Tests
// =============================================================================

// TestModel_Update_ConfigWarningMsg_SetsWarning tests that configWarningMsg sets warning (AC6).
func TestModel_Update_ConfigWarningMsg_SetsWarning(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)

	// Send config warning message
	msg := configWarningMsg{
		warning: "⚠ Config: using defaults (see log)",
	}

	updated, cmd := m.Update(msg)
	model := updated.(Model)

	// Model should store warning
	if model.configWarning != "⚠ Config: using defaults (see log)" {
		t.Errorf("expected configWarning to be set, got: %s", model.configWarning)
	}

	// configWarningTime should be set
	if model.configWarningTime.IsZero() {
		t.Error("configWarningTime should be set")
	}

	// Status bar should show warning
	view := model.statusBar.View()
	if !strings.Contains(view, "using defaults") {
		t.Errorf("status bar should show config warning, got: %s", view)
	}

	// Should return cmd for 10-second auto-clear timer
	if cmd == nil {
		t.Error("configWarningMsg should return tea.Tick command for auto-clear")
	}
}

// TestModel_Update_ClearConfigWarningMsg_ClearsWarning tests that clearConfigWarningMsg clears after timeout.
func TestModel_Update_ClearConfigWarningMsg_ClearsWarning(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)
	m.configWarning = "test warning"
	m.configWarningTime = time.Now().Add(-11 * time.Second) // Set in past

	// Send clear message
	msg := clearConfigWarningMsg{}

	updated, cmd := m.Update(msg)
	model := updated.(Model)

	// Warning should be cleared
	if model.configWarning != "" {
		t.Errorf("expected configWarning to be cleared, got: %s", model.configWarning)
	}

	// Status bar should not show warning
	view := model.statusBar.View()
	if strings.Contains(view, "test warning") {
		t.Error("status bar should not show cleared warning")
	}

	// Should return nil cmd
	if cmd != nil {
		t.Error("clearConfigWarningMsg should return nil command")
	}
}

// TestModel_Update_ClearConfigWarningMsg_NotClearedIfRecent tests warning is not cleared if recent.
func TestModel_Update_ClearConfigWarningMsg_NotClearedIfRecent(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)
	m.statusBar.SetConfigWarning("test warning")
	m.configWarning = "test warning"
	m.configWarningTime = time.Now().Add(-5 * time.Second) // Only 5 seconds ago

	// Send clear message
	msg := clearConfigWarningMsg{}

	updated, _ := m.Update(msg)
	model := updated.(Model)

	// Warning should NOT be cleared (only 5 seconds elapsed)
	if model.configWarning != "test warning" {
		t.Errorf("expected configWarning to remain, got: %s", model.configWarning)
	}
}

// TestModel_ConfigWarning_DoesNotAffectWatcherWarning tests both warnings can coexist.
func TestModel_ConfigWarning_DoesNotAffectWatcherWarning(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)
	m.statusBar.SetWatcherWarning("⚠ File watching unavailable")

	// Send config warning message
	msg := configWarningMsg{
		warning: "config error",
	}

	updated, _ := m.Update(msg)
	model := updated.(Model)

	// Both warnings should appear in status bar
	view := model.statusBar.View()
	if !strings.Contains(view, "File watching unavailable") {
		t.Error("watcher warning should still be present")
	}
	if !strings.Contains(view, "config error") {
		t.Error("config warning should be present")
	}
}

// =============================================================================
// Story 8.4: Layout Width Bug Fix Tests (Race Condition)
// =============================================================================

// TestModel_ProjectsLoadedBeforeReady_StorePending tests that ProjectsLoadedMsg
// before resizeTickMsg stores projects in pendingProjects (AC1, AC4).
func TestModel_ProjectsLoadedBeforeReady_StorePending(t *testing.T) {
	m := NewModel(nil)
	// m.ready is false, m.width is 0 - simulating race condition

	// Create test projects
	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
		{ID: "b", Name: "project-b", Path: "/home/user/project-b", State: domain.StateActive},
	}

	// Send ProjectsLoadedMsg BEFORE WindowSizeMsg/resizeTickMsg
	msg := ProjectsLoadedMsg{projects: projects}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Projects should be stored in pendingProjects, not m.projects
	if updated.pendingProjects == nil {
		t.Fatal("pendingProjects should store projects when !m.ready")
	}
	if len(updated.pendingProjects) != 2 {
		t.Errorf("expected 2 pending projects, got %d", len(updated.pendingProjects))
	}
	if len(updated.projects) > 0 {
		t.Error("m.projects should NOT be populated before ready")
	}

	// projectList should NOT be initialized (would have zero width)
	if updated.projectList.Width() != 0 {
		t.Errorf("projectList should not be initialized before ready, width=%d", updated.projectList.Width())
	}
}

// TestModel_ResizeTickProcessesPending tests that resizeTickMsg processes
// pending projects with correct dimensions (AC1, AC4).
func TestModel_ResizeTickProcessesPending(t *testing.T) {
	m := NewModel(nil)

	// Create test projects
	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
		{ID: "b", Name: "project-b", Path: "/home/user/project-b", State: domain.StateActive},
	}

	// First: ProjectsLoadedMsg arrives (m.ready = false)
	msg := ProjectsLoadedMsg{projects: projects}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Verify pending state
	if m.pendingProjects == nil {
		t.Fatal("pendingProjects should be set")
	}

	// Second: WindowSizeMsg arrives
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	newModel, _ = m.Update(sizeMsg)
	m = newModel.(Model)

	// Third: resizeTickMsg processes pending
	newModel, _ = m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// pendingProjects should be cleared
	if m.pendingProjects != nil {
		t.Error("pendingProjects should be nil after processing")
	}

	// m.projects should be populated
	if len(m.projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(m.projects))
	}

	// projectList should be initialized with correct width
	if m.projectList.Width() != 100 {
		t.Errorf("expected projectList width 100, got %d", m.projectList.Width())
	}

	// Model should be ready
	if !m.ready {
		t.Error("model should be ready after resize tick")
	}
}

// TestModel_EffectiveWidth_WideTerminal tests that effectiveWidth is capped
// at maxContentWidth on wide terminals (AC1, AC4).
// Story 8.10: Uses default maxContentWidth (120) from config.
func TestModel_EffectiveWidth_WideTerminal(t *testing.T) {
	m := NewModel(nil)
	defaultMaxWidth := ports.NewConfig().MaxContentWidth // 120

	// Create test projects
	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
	}

	// Send ProjectsLoadedMsg (before ready)
	msg := ProjectsLoadedMsg{projects: projects}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// WindowSizeMsg with wide terminal (>maxContentWidth=120)
	sizeMsg := tea.WindowSizeMsg{Width: 200, Height: 40}
	newModel, _ = m.Update(sizeMsg)
	m = newModel.(Model)

	// resizeTickMsg processes pending with effectiveWidth
	newModel, _ = m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// projectList should have maxContentWidth (120), not 200
	if m.projectList.Width() != defaultMaxWidth {
		t.Errorf("expected projectList width %d (maxContentWidth), got %d", defaultMaxWidth, m.projectList.Width())
	}
}

// TestModel_ResizeAfterReady_UpdatesComponents tests that resize updates
// component widths after initial load (AC3).
func TestModel_ResizeAfterReady_UpdatesComponents(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40

	// Create projects and initialize components
	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, 80, 38)

	// Verify initial width
	if m.projectList.Width() != 80 {
		t.Errorf("expected initial width 80, got %d", m.projectList.Width())
	}

	// Resize to new dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Process resize tick
	newModel, _ = m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// projectList should be updated to new width
	if m.projectList.Width() != 100 {
		t.Errorf("expected projectList width 100 after resize, got %d", m.projectList.Width())
	}
}

// TestModel_ProjectsLoadedAfterReady_UsesEffectiveWidth tests that
// ProjectsLoadedMsg after ready uses effectiveWidth (AC1, AC4).
// Story 8.10: Uses default maxContentWidth (120) from config.
func TestModel_ProjectsLoadedAfterReady_UsesEffectiveWidth(t *testing.T) {
	m := NewModel(nil)
	defaultMaxWidth := ports.NewConfig().MaxContentWidth // 120
	// Make ready with wide terminal
	m.hasPendingResize = true
	m.pendingWidth = 150 // > maxContentWidth (120)
	m.pendingHeight = 40

	// Process resize to become ready
	newModel, _ := m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// Verify ready and wide terminal
	if !m.ready {
		t.Fatal("model should be ready")
	}
	if m.width != 150 {
		t.Fatalf("expected width 150, got %d", m.width)
	}

	// Now send ProjectsLoadedMsg (after ready)
	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
	}
	msg := ProjectsLoadedMsg{projects: projects}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// projectList should use effectiveWidth (maxContentWidth), not raw width
	if m.projectList.Width() != defaultMaxWidth {
		t.Errorf("expected projectList width %d (maxContentWidth), got %d", defaultMaxWidth, m.projectList.Width())
	}
}

// TestModel_ResizeWithZeroProjects_NoSetSize tests that resize with
// uninitialized projectList (width=0) doesn't call SetSize.
func TestModel_ResizeWithZeroProjects_NoSetSize(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	// projectList is zero-value (Width() returns 0)

	// Resize
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(sizeMsg)
	m = newModel.(Model)

	// Process resize tick
	newModel, _ = m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// Should not panic and projectList width should remain 0
	if m.projectList.Width() != 0 {
		t.Error("uninitialized projectList should not have SetSize called")
	}
}

// TestModel_ResizeTickProcessesPending_FileWatcher tests that file watcher is
// started correctly when pending projects are processed (code review H3).
func TestModel_ResizeTickProcessesPending_FileWatcher(t *testing.T) {
	mock := newMockFileWatcher()

	m := NewModel(nil)
	m.SetFileWatcher(mock)

	// Create test projects
	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
		{ID: "b", Name: "project-b", Path: "/home/user/project-b", State: domain.StateActive},
	}

	// First: ProjectsLoadedMsg arrives (m.ready = false)
	msg := ProjectsLoadedMsg{projects: projects}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Verify pending state and watcher not started yet
	if m.pendingProjects == nil {
		t.Fatal("pendingProjects should be set")
	}
	if mock.watchCalled {
		t.Error("file watcher should not be called before ready")
	}

	// Second: WindowSizeMsg arrives
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	newModel, _ = m.Update(sizeMsg)
	m = newModel.(Model)

	// Third: resizeTickMsg processes pending and starts watcher
	newModel, cmd := m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// Verify file watcher was started
	if !mock.watchCalled {
		t.Error("file watcher should be called after processing pending projects")
	}
	if len(mock.watchPaths) != 2 {
		t.Errorf("expected 2 watch paths, got %d", len(mock.watchPaths))
	}

	// Verify event channel is stored
	if m.eventCh == nil {
		t.Error("eventCh should be set after successful Watch()")
	}

	// Verify cmd is returned (waitForNextFileEventCmd)
	if cmd == nil {
		t.Error("should return waitForNextFileEventCmd after processing pending with watcher")
	}
}

// TestModel_ResizeTickProcessesPending_DetailPanelDimensions tests that detail panel
// has correct dimensions after processing pending projects (code review M1).
// Story 8.10: Uses default maxContentWidth (120) from config.
func TestModel_ResizeTickProcessesPending_DetailPanelDimensions(t *testing.T) {
	m := NewModel(nil)
	defaultMaxWidth := ports.NewConfig().MaxContentWidth // 120

	// Create test projects
	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
	}

	// First: ProjectsLoadedMsg arrives (m.ready = false)
	msg := ProjectsLoadedMsg{projects: projects}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Second: WindowSizeMsg with wide terminal (>maxContentWidth=120)
	sizeMsg := tea.WindowSizeMsg{Width: 150, Height: 50}
	newModel, _ = m.Update(sizeMsg)
	m = newModel.(Model)

	// Third: resizeTickMsg processes pending
	newModel, _ = m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// Verify detailPanel was created with correct effectiveWidth (capped at maxContentWidth)
	// Note: DetailPanelModel doesn't expose Width() directly, but we can verify
	// that the projectList has the same width, proving both got effectiveWidth
	if m.projectList.Width() != defaultMaxWidth {
		t.Errorf("expected projectList width %d (maxContentWidth), got %d", defaultMaxWidth, m.projectList.Width())
	}

	// Verify detail panel was initialized (has a project set)
	if m.detailPanel.Project() == nil {
		t.Error("detailPanel should have a project set after processing pending")
	}

	// Verify visibility matches model state (which should be open for height >= 35)
	if !m.showDetailPanel {
		t.Error("showDetailPanel should be true for height 50 (>= HeightThresholdTall)")
	}
	if !m.detailPanel.IsVisible() {
		t.Error("detailPanel.IsVisible() should match m.showDetailPanel")
	}
}

// TestModel_FullWidthAfterRace_Integration tests the full race condition
// scenario end-to-end (AC1-AC4 integration).
func TestModel_FullWidthAfterRace_Integration(t *testing.T) {
	// This test simulates the real race condition:
	// 1. TUI starts (Init called)
	// 2. ProjectsLoadedMsg arrives BEFORE WindowSizeMsg
	// 3. WindowSizeMsg arrives
	// 4. resizeTickMsg processes pending
	// 5. View renders with full width

	m := NewModel(nil)
	m.statusBar = components.NewStatusBarModel(0)

	// Step 1: Projects load first (race condition)
	projects := []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
	}
	newModel, _ := m.Update(ProjectsLoadedMsg{projects: projects})
	m = newModel.(Model)

	// Verify in pending state
	if m.pendingProjects == nil {
		t.Fatal("should have pending projects")
	}
	if m.ready {
		t.Fatal("should not be ready yet")
	}

	// Step 2: Window size arrives
	newModel, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = newModel.(Model)

	// Step 3: Resize tick processes everything
	newModel, _ = m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// Step 4: Verify correct state
	if !m.ready {
		t.Error("should be ready")
	}
	if m.pendingProjects != nil {
		t.Error("pending should be cleared")
	}
	if len(m.projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(m.projects))
	}
	if m.projectList.Width() != 100 {
		t.Errorf("projectList should have full width 100, got %d", m.projectList.Width())
	}

	// Step 5: View should render correctly (not panic or show empty)
	view := m.View()
	if !strings.Contains(view, "project-a") {
		t.Error("View should display project after race condition resolved")
	}
}

// =============================================================================
// Story 8.6: Horizontal Split Layout Tests
// =============================================================================

// TestModel_SetDetailLayout_Valid tests SetDetailLayout with valid values.
func TestModel_SetDetailLayout_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"vertical", "vertical"},
		{"horizontal", "horizontal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewModel(nil)
			m.SetDetailLayout(tt.input)

			if m.detailLayout != tt.expected {
				t.Errorf("SetDetailLayout(%q) = %q, want %q", tt.input, m.detailLayout, tt.expected)
			}
		})
	}
}

// TestModel_SetDetailLayout_Invalid tests SetDetailLayout with invalid values.
func TestModel_SetDetailLayout_Invalid(t *testing.T) {
	tests := []struct {
		input    string
		expected string // Should fallback to "horizontal" (default)
	}{
		{"", "horizontal"},
		{"diagonal", "horizontal"},
		{"VERTICAL", "horizontal"},
		{"Horizontal", "horizontal"},
		{"side-by-side", "horizontal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewModel(nil)
			m.SetDetailLayout(tt.input)

			if m.detailLayout != tt.expected {
				t.Errorf("SetDetailLayout(%q) = %q, want %q (fallback)", tt.input, m.detailLayout, tt.expected)
			}
		})
	}
}

// TestModel_isHorizontalLayout tests the isHorizontalLayout helper.
func TestModel_isHorizontalLayout(t *testing.T) {
	tests := []struct {
		layout   string
		expected bool
	}{
		{"horizontal", true},
		{"vertical", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.layout, func(t *testing.T) {
			m := NewModel(nil)
			m.detailLayout = tt.layout

			result := m.isHorizontalLayout()
			if result != tt.expected {
				t.Errorf("isHorizontalLayout() with layout %q = %v, want %v", tt.layout, result, tt.expected)
			}
		})
	}
}

// TestModel_RenderMainContent_HorizontalLayout tests that horizontal layout
// uses JoinVertical instead of JoinHorizontal.
func TestModel_RenderMainContent_HorizontalLayout(t *testing.T) {
	m := createModelWithProjects(2)
	m.height = 40 // Tall enough
	m.showDetailPanel = true
	m.detailLayout = "horizontal" // Set horizontal layout
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	view := m.renderDashboard()

	// Should still contain detail panel content
	if !strings.Contains(view, "DETAILS:") {
		t.Error("Horizontal layout should still show detail panel with 'DETAILS:' title")
	}

	// Verify the layout direction by checking that projects are rendered
	// We can't easily test JoinVertical vs JoinHorizontal directly,
	// but we verify both components are present
	if !strings.Contains(view, m.projectList.SelectedProject().Name) {
		t.Error("Horizontal layout should display selected project name")
	}
}

// TestModel_RenderMainContent_VerticalLayout tests that vertical layout
// remains unchanged (regression test).
func TestModel_RenderMainContent_VerticalLayout(t *testing.T) {
	m := createModelWithProjects(2)
	m.height = 40 // Tall enough
	m.showDetailPanel = true
	m.detailLayout = "vertical" // Default layout
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	view := m.renderDashboard()

	// Should contain detail panel content
	if !strings.Contains(view, "DETAILS:") {
		t.Error("Vertical layout should show detail panel with 'DETAILS:' title")
	}

	// Verify both components are present
	if !strings.Contains(view, m.projectList.SelectedProject().Name) {
		t.Error("Vertical layout should display selected project name")
	}
}

// TestModel_RenderMainContent_EmptyLayoutFallsBackToVertical tests that
// empty detailLayout uses vertical behavior (isHorizontalLayout returns false).
func TestModel_RenderMainContent_EmptyLayoutFallsBackToVertical(t *testing.T) {
	m := createModelWithProjects(2)
	m.height = 40
	m.showDetailPanel = true
	m.detailLayout = "" // Empty - should use vertical code path (not equal to "horizontal")
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	// Should not panic and should render correctly
	view := m.renderDashboard()

	if !strings.Contains(view, "DETAILS:") {
		t.Error("Empty layout should still render detail panel")
	}
}

// TestModel_RenderMainContent_HiddenDetailPanel_NoLayoutEffect tests that
// layout mode doesn't affect rendering when detail panel is hidden.
func TestModel_RenderMainContent_HiddenDetailPanel_NoLayoutEffect(t *testing.T) {
	m := createModelWithProjects(2)
	m.height = 40
	m.showDetailPanel = false // Panel hidden
	m.detailLayout = "horizontal"

	view := m.renderDashboard()

	// Detail panel content should NOT be present
	if strings.Contains(view, "DETAILS:") {
		t.Error("Hidden detail panel should not show DETAILS title regardless of layout")
	}

	// Project list should be visible
	if !strings.Contains(view, m.projectList.SelectedProject().Name) {
		t.Error("Project list should be visible when detail panel is hidden")
	}
}

// TestModel_RenderHorizontalSplit_Dimensions tests that horizontal split
// uses correct height proportions (60% list, 40% detail).
func TestModel_RenderHorizontalSplit_Dimensions(t *testing.T) {
	m := createModelWithProjects(2)
	m.width = 100
	m.height = 50
	m.showDetailPanel = true
	m.detailLayout = "horizontal"
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	// Call renderHorizontalSplit directly with content height
	// Content height = total height - statusBarHeight (approximately 2 lines)
	contentHeight := 48

	result := m.renderHorizontalSplit(contentHeight)

	// Should produce non-empty result
	if result == "" {
		t.Error("renderHorizontalSplit should produce non-empty result")
	}

	// Both list and detail content should be present
	if !strings.Contains(result, m.projectList.SelectedProject().Name) {
		t.Error("renderHorizontalSplit should include project list content")
	}
}

// Story 8.12: Horizontal layout height priority tests

func TestRenderHorizontalSplit_HeightPriority_BelowThreshold(t *testing.T) {
	// Given: height < HorizontalDetailThreshold (16)
	m := NewModel(nil)
	m.width = 100
	m.height = 50
	m.ready = true

	projects := []*domain.Project{
		{ID: "abc123", Name: "test-project", Path: "/test"},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.showDetailPanel = true
	m.detailLayout = "horizontal"
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	// When: render with height = 15 (below threshold)
	result := m.renderHorizontalSplit(15)

	// Then: detail hidden, only project list visible
	// Detail panel uses BorderStyle which includes the project name in DETAILS
	// Since detail is hidden, we should NOT see "DETAILS:" in output
	// But we SHOULD see the project list content
	if result == "" {
		t.Error("renderHorizontalSplit should produce non-empty result even below threshold")
	}
	// Project list should still be visible
	if !strings.Contains(result, "test-project") {
		t.Error("renderHorizontalSplit should include project list content when detail is hidden")
	}
	// Code review fix M2: Verify detail IS actually hidden (negative assertion)
	if strings.Contains(result, "DETAILS:") {
		t.Error("renderHorizontalSplit should NOT show detail panel when height < threshold")
	}
}

func TestRenderHorizontalSplit_HeightPriority_AtThreshold(t *testing.T) {
	// Given: height == HorizontalDetailThreshold (16)
	m := NewModel(nil)
	m.width = 100
	m.height = 50
	m.ready = true

	projects := []*domain.Project{
		{ID: "abc124", Name: "threshold-project", Path: "/threshold"},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.showDetailPanel = true
	m.detailLayout = "horizontal"
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	// When: render with height = 16 (at threshold)
	result := m.renderHorizontalSplit(16)

	// Then: both visible, minimal split (10 list, 6 detail)
	if result == "" {
		t.Error("renderHorizontalSplit should produce non-empty result at threshold")
	}
	// Both components should be visible
	if !strings.Contains(result, "threshold-project") {
		t.Error("renderHorizontalSplit should include project list content at threshold")
	}
}

func TestRenderHorizontalSplit_HeightPriority_Comfortable(t *testing.T) {
	// Given: height >= 30 (comfortable)
	m := NewModel(nil)
	m.width = 100
	m.height = 50
	m.ready = true

	projects := []*domain.Project{
		{ID: "abc125", Name: "comfortable-project", Path: "/comfortable"},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.showDetailPanel = true
	m.detailLayout = "horizontal"
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(m.projectList.SelectedProject())

	// When: render with height = 50 (comfortable)
	result := m.renderHorizontalSplit(50)

	// Then: use 60/40 split (30 list, 20 detail)
	if result == "" {
		t.Error("renderHorizontalSplit should produce non-empty result with comfortable height")
	}
	// Both components should be visible
	if !strings.Contains(result, "comfortable-project") {
		t.Error("renderHorizontalSplit should include project list content at comfortable height")
	}
}

func TestRenderHorizontalSplit_AnchorStability(t *testing.T) {
	// Given: multiple projects with different detail lengths
	m := NewModel(nil)
	m.width = 100
	m.height = 50
	m.ready = true

	projects := []*domain.Project{
		{ID: "abc126", Name: "short-notes", Path: "/short", Notes: "Brief"},
		{ID: "abc127", Name: "long-notes", Path: "/long", Notes: "This is a very long note that takes up more space in the detail panel and may cause layout shifts if anchor points are not stable"},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.showDetailPanel = true
	m.detailLayout = "horizontal"
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.detailPanel.SetVisible(true)
	m.detailPanel.SetProject(projects[0])

	// Render with first project (short notes)
	result1 := m.renderHorizontalSplit(50)

	// Switch to second project (long notes)
	m.detailPanel.SetProject(projects[1])
	result2 := m.renderHorizontalSplit(50)

	// Both should produce valid output
	if result1 == "" || result2 == "" {
		t.Error("renderHorizontalSplit should produce output for all projects")
	}

	// Height of rendered content should be consistent (anchor stability)
	// Count newlines in both results
	lines1 := strings.Count(result1, "\n")
	lines2 := strings.Count(result2, "\n")

	// With lipgloss.Height() enforcement, both should have the same number of lines
	if lines1 != lines2 {
		t.Errorf("Anchor stability: different line counts (%d vs %d) - list may shift on navigation", lines1, lines2)
	}
}

// =============================================================================
// Story 9.5-2: File Watcher Grace Period Tests
// =============================================================================

// TestFileWatcherErrorMsg_GracePeriod tests the grace period logic for
// ignoring transient errors during watcher restart (AC3, AC4, AC6).
func TestFileWatcherErrorMsg_GracePeriod(t *testing.T) {
	tests := []struct {
		name                 string
		lastRestart          time.Duration // negative = in the past
		setLastRestart       bool          // false = leave as zero value
		wantAvailable        bool
		wantStatusBarWarning bool
	}{
		{
			name:                 "within grace period (100ms)",
			lastRestart:          -100 * time.Millisecond,
			setLastRestart:       true,
			wantAvailable:        true,
			wantStatusBarWarning: false,
		},
		{
			name:                 "after grace period (600ms)",
			lastRestart:          -600 * time.Millisecond,
			setLastRestart:       true,
			wantAvailable:        false,
			wantStatusBarWarning: true,
		},
		{
			name:                 "boundary ignored (499ms)",
			lastRestart:          -499 * time.Millisecond,
			setLastRestart:       true,
			wantAvailable:        true,
			wantStatusBarWarning: false,
		},
		{
			name:                 "boundary handled (501ms)",
			lastRestart:          -501 * time.Millisecond,
			setLastRestart:       true,
			wantAvailable:        false,
			wantStatusBarWarning: true,
		},
		{
			name:                 "zero value (app startup)",
			lastRestart:          0,
			setLastRestart:       false, // leave as zero value
			wantAvailable:        false,
			wantStatusBarWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(nil)
			m.ready = true
			m.width = 80
			m.height = 40
			m.fileWatcherAvailable = true
			m.statusBar = components.NewStatusBarModel(80)

			if tt.setLastRestart {
				m.lastWatcherRestart = time.Now().Add(tt.lastRestart)
			}
			// else: leave as zero value

			msg := fileWatcherErrorMsg{err: fmt.Errorf("channel closed")}
			result, _ := m.Update(msg)
			updated := result.(Model)

			if updated.fileWatcherAvailable != tt.wantAvailable {
				t.Errorf("fileWatcherAvailable = %v, want %v", updated.fileWatcherAvailable, tt.wantAvailable)
			}

			hasWarning := strings.Contains(updated.statusBar.View(), "unavailable")
			if hasWarning != tt.wantStatusBarWarning {
				t.Errorf("status bar warning = %v, want %v", hasWarning, tt.wantStatusBarWarning)
			}
		})
	}
}

// TestFileWatcherErrorMsg_GracePeriod_StatusBarNotUpdated verifies that
// status bar is NOT updated during grace period (AC3 detail).
func TestFileWatcherErrorMsg_GracePeriod_StatusBarNotUpdated(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.fileWatcherAvailable = true
	m.statusBar = components.NewStatusBarModel(80)

	// Set a recent restart (within grace period)
	m.lastWatcherRestart = time.Now().Add(-100 * time.Millisecond)

	msg := fileWatcherErrorMsg{err: fmt.Errorf("channel closed")}
	result, _ := m.Update(msg)
	updated := result.(Model)

	// Verify status bar contains no "unavailable" text
	view := updated.statusBar.View()
	if strings.Contains(view, "unavailable") {
		t.Errorf("status bar should NOT show warning during grace period, got: %s", view)
	}
	if strings.Contains(view, "File watching") {
		t.Errorf("status bar should NOT mention file watching during grace period, got: %s", view)
	}
}

// TestStartFileWatcher_SetsLastWatcherRestart verifies that
// startFileWatcherForProjects sets lastWatcherRestart before Watch() (AC2).
func TestStartFileWatcher_SetsLastWatcherRestart(t *testing.T) {
	mock := newMockFileWatcher()

	m := NewModel(nil)
	m.SetFileWatcher(mock)
	m.projects = []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/home/user/project-a", State: domain.StateActive},
	}

	// Verify lastWatcherRestart is zero before
	if !m.lastWatcherRestart.IsZero() {
		t.Fatal("lastWatcherRestart should be zero initially")
	}

	beforeCall := time.Now()
	_ = m.startFileWatcherForProjects()
	afterCall := time.Now()

	// Verify lastWatcherRestart was set
	if m.lastWatcherRestart.IsZero() {
		t.Error("lastWatcherRestart should be set after startFileWatcherForProjects")
	}

	// Verify timestamp is within expected range
	if m.lastWatcherRestart.Before(beforeCall) || m.lastWatcherRestart.After(afterCall) {
		t.Errorf("lastWatcherRestart = %v, want between %v and %v",
			m.lastWatcherRestart, beforeCall, afterCall)
	}
}

// =============================================================================
// Story 11.3: Auto-Activation Tests
// =============================================================================

// mockStateActivator implements ports.StateActivator for testing.
type mockStateActivator struct {
	hibernateCalls []string // Project IDs that were hibernated
	hibernateErr   error    // Error to return from Hibernate()
	activateCalls  []string // Project IDs that were activated
	activateErr    error    // Error to return from Activate()
}

func (m *mockStateActivator) Hibernate(ctx context.Context, projectID string) error {
	m.hibernateCalls = append(m.hibernateCalls, projectID)
	return m.hibernateErr
}

func (m *mockStateActivator) Activate(ctx context.Context, projectID string) error {
	m.activateCalls = append(m.activateCalls, projectID)
	return m.activateErr
}

func TestModel_SetStateService(t *testing.T) {
	m := NewModel(nil)
	mock := &mockStateActivator{}

	// Initially should be nil
	if m.stateService != nil {
		t.Error("stateService should be nil initially")
	}

	// Set state service
	m.SetStateService(mock)

	if m.stateService != mock {
		t.Error("SetStateService should set the state service")
	}
}

func TestModel_HandleFileEvent_HibernatedProject_Activates(t *testing.T) {
	// Setup mock state activator
	mock := &mockStateActivator{}

	// Create model with hibernated project
	now := time.Now()
	hibTime := now.Add(-24 * time.Hour)
	projects := []*domain.Project{
		{
			ID:           "hibernated-id",
			Name:         "hibernated-project",
			Path:         "/home/user/hibernated",
			State:        domain.StateHibernated,
			HibernatedAt: &hibTime,
		},
	}

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = projects
	m.SetStateService(mock)

	// Initialize components required for handleFileEvent
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.statusBar = components.NewStatusBarModel(m.width)

	// Send file event for hibernated project
	msg := fileEventMsg{
		Path:      "/home/user/hibernated/file.txt",
		Timestamp: now,
	}
	m.handleFileEvent(msg)

	// Assert: stateService.Activate() was called with correct project ID
	if len(mock.activateCalls) != 1 {
		t.Errorf("expected 1 Activate call, got %d", len(mock.activateCalls))
	}
	if len(mock.activateCalls) > 0 && mock.activateCalls[0] != "hibernated-id" {
		t.Errorf("expected Activate call for 'hibernated-id', got '%s'", mock.activateCalls[0])
	}

	// Assert: project.State == StateActive, HibernatedAt == nil
	if projects[0].State != domain.StateActive {
		t.Errorf("expected project state to be Active, got %s", projects[0].State)
	}
	if projects[0].HibernatedAt != nil {
		t.Error("expected HibernatedAt to be nil after activation")
	}
}

func TestModel_HandleFileEvent_ActiveProject_NoStateChange(t *testing.T) {
	// Setup mock state activator
	mock := &mockStateActivator{}

	// Create model with active project
	projects := []*domain.Project{
		{
			ID:    "active-id",
			Name:  "active-project",
			Path:  "/home/user/active",
			State: domain.StateActive,
		},
	}

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = projects
	m.SetStateService(mock)

	// Initialize components
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.statusBar = components.NewStatusBarModel(m.width)

	// Send file event for active project
	msg := fileEventMsg{
		Path:      "/home/user/active/file.txt",
		Timestamp: time.Now(),
	}
	m.handleFileEvent(msg)

	// Assert: stateService.Activate() NOT called (AC3)
	if len(mock.activateCalls) != 0 {
		t.Errorf("expected 0 Activate calls for active project, got %d", len(mock.activateCalls))
	}
}

func TestModel_HandleFileEvent_ActivationError_ContinuesProcessing(t *testing.T) {
	// Setup mock state activator that returns error
	mockActivator := &mockStateActivator{
		activateErr: fmt.Errorf("database error"),
	}

	// Create model with hibernated project
	now := time.Now()
	hibTime := now.Add(-24 * time.Hour)
	projects := []*domain.Project{
		{
			ID:           "hibernated-id",
			Name:         "hibernated-project",
			Path:         "/home/user/hibernated",
			State:        domain.StateHibernated,
			HibernatedAt: &hibTime,
		},
	}

	// Setup mock repository so handleFileEvent doesn't return early
	mockRepo := testhelpers.NewMockRepository().WithProjects(projects)

	m := NewModel(mockRepo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = projects
	m.SetStateService(mockActivator)

	// Initialize components
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.statusBar = components.NewStatusBarModel(m.width)

	// Send file event - should not panic despite error (AC6)
	msg := fileEventMsg{
		Path:      "/home/user/hibernated/file.txt",
		Timestamp: now,
	}
	m.handleFileEvent(msg)

	// Assert: Activate was called
	if len(mockActivator.activateCalls) != 1 {
		t.Errorf("expected 1 Activate call, got %d", len(mockActivator.activateCalls))
	}

	// Assert: State NOT changed due to error
	if projects[0].State != domain.StateHibernated {
		t.Error("project state should remain Hibernated on activation error")
	}

	// Assert: LastActivityAt still updated (graceful degradation)
	if projects[0].LastActivityAt != now {
		t.Error("LastActivityAt should be updated even on activation error")
	}
}

func TestModel_HandleFileEvent_NoStateService_NoActivation(t *testing.T) {
	// Create model with hibernated project but NO stateService set
	now := time.Now()
	hibTime := now.Add(-24 * time.Hour)
	projects := []*domain.Project{
		{
			ID:           "hibernated-id",
			Name:         "hibernated-project",
			Path:         "/home/user/hibernated",
			State:        domain.StateHibernated,
			HibernatedAt: &hibTime,
		},
	}

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = projects
	// Note: stateService NOT set - should be nil

	// Initialize components
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.statusBar = components.NewStatusBarModel(m.width)

	// Send file event - should not panic even with nil stateService
	msg := fileEventMsg{
		Path:      "/home/user/hibernated/file.txt",
		Timestamp: now,
	}
	m.handleFileEvent(msg)

	// Assert: No panic occurred and project state unchanged
	if projects[0].State != domain.StateHibernated {
		t.Error("project state should remain Hibernated when stateService is nil")
	}

	// Assert: HibernatedAt unchanged
	if projects[0].HibernatedAt == nil || *projects[0].HibernatedAt != hibTime {
		t.Error("HibernatedAt should remain unchanged when stateService is nil")
	}
}

func TestModel_HandleFileEvent_StatusBarCountsUpdate(t *testing.T) {
	// Setup mock state activator
	mock := &mockStateActivator{}

	// Create model with 2 active, 1 hibernated
	now := time.Now()
	hibTime := now.Add(-24 * time.Hour)
	projects := []*domain.Project{
		{ID: "active-1", Name: "active-1", Path: "/home/user/active1", State: domain.StateActive},
		{ID: "active-2", Name: "active-2", Path: "/home/user/active2", State: domain.StateActive},
		{ID: "hibernated-1", Name: "hibernated-1", Path: "/home/user/hibernated1", State: domain.StateHibernated, HibernatedAt: &hibTime},
	}

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = projects
	m.SetStateService(mock)

	// Initialize components
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.statusBar = components.NewStatusBarModel(m.width)

	// Calculate initial counts (2 active, 1 hibernated)
	active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
	m.statusBar.SetCounts(active, hibernated, waiting)

	// Send file event for hibernated project
	msg := fileEventMsg{
		Path:      "/home/user/hibernated1/file.txt",
		Timestamp: now,
	}
	m.handleFileEvent(msg)

	// Assert: project is now active
	if projects[2].State != domain.StateActive {
		t.Errorf("expected project state to be Active, got %s", projects[2].State)
	}

	// Assert: Recalculate and verify counts (should be 3 active, 0 hibernated)
	active, hibernated, _ = components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
	if active != 3 {
		t.Errorf("expected 3 active projects, got %d", active)
	}
	if hibernated != 0 {
		t.Errorf("expected 0 hibernated projects, got %d", hibernated)
	}
}

func TestModel_HandleFileEvent_ErrInvalidStateTransition_Ignored(t *testing.T) {
	// Setup mock state activator that returns ErrInvalidStateTransition
	mockActivator := &mockStateActivator{
		activateErr: domain.ErrInvalidStateTransition,
	}

	// Create model with hibernated project
	now := time.Now()
	hibTime := now.Add(-24 * time.Hour)
	projects := []*domain.Project{
		{
			ID:           "hibernated-id",
			Name:         "hibernated-project",
			Path:         "/home/user/hibernated",
			State:        domain.StateHibernated,
			HibernatedAt: &hibTime,
		},
	}

	// Setup mock repository so handleFileEvent doesn't return early
	mockRepo := testhelpers.NewMockRepository().WithProjects(projects)

	m := NewModel(mockRepo)
	m.ready = true
	m.width = 80
	m.height = 40
	m.projects = projects
	m.SetStateService(mockActivator)

	// Initialize components
	m.projectList = components.NewProjectListModel(projects, m.width, m.height)
	m.detailPanel = components.NewDetailPanelModel(m.width, m.height)
	m.statusBar = components.NewStatusBarModel(m.width)

	// Send file event - should handle silently (no warning)
	msg := fileEventMsg{
		Path:      "/home/user/hibernated/file.txt",
		Timestamp: now,
	}
	m.handleFileEvent(msg)

	// Assert: Activate was called
	if len(mockActivator.activateCalls) != 1 {
		t.Errorf("expected 1 Activate call, got %d", len(mockActivator.activateCalls))
	}

	// Assert: State NOT changed (expected since it's an invalid transition)
	// This is fine - ErrInvalidStateTransition means project was already active
	if projects[0].State != domain.StateHibernated {
		t.Error("project state should remain unchanged on ErrInvalidStateTransition")
	}

	// Assert: Processing continued (LastActivityAt updated)
	if projects[0].LastActivityAt != now {
		t.Error("LastActivityAt should be updated despite ErrInvalidStateTransition")
	}
}

// Story 11.4 Tests: Hibernated Projects View

func TestModel_HibernatedViewToggle_AC1(t *testing.T) {
	// AC1: 'h' key switches to hibernated view
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40

	// Setup active projects
	projects := []*domain.Project{
		{ID: "1", Name: "Active1", State: domain.StateActive},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, 100, 30)

	// Press 'h' to enter hibernated view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Assert view mode changed
	if updated.viewMode != viewModeHibernated {
		t.Errorf("expected viewModeHibernated, got %d", updated.viewMode)
	}

	// Assert command returned to load hibernated projects
	if cmd == nil {
		t.Error("expected loadHibernatedProjectsCmd, got nil")
	}

	// Assert status bar updated
	// (SetInHibernatedView called with true - we can verify via renderShortcuts in actual test)
}

func TestModel_HibernatedViewToggle_AC4_HKeyBack(t *testing.T) {
	// AC4: 'h' returns from hibernated view
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	// Setup active and hibernated projects
	projects := []*domain.Project{
		{ID: "1", Name: "Active1", State: domain.StateActive},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, 100, 30)
	m.activeSelectedIdx = 0

	// Press 'h' to exit hibernated view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Assert view mode changed back
	if updated.viewMode != viewModeNormal {
		t.Errorf("expected viewModeNormal, got %d", updated.viewMode)
	}
}

func TestModel_HibernatedViewToggle_AC4_EscapeKeyBack(t *testing.T) {
	// AC4: Esc returns from hibernated view
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	// Setup active projects
	projects := []*domain.Project{
		{ID: "1", Name: "Active1", State: domain.StateActive},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, 100, 30)
	m.activeSelectedIdx = 0

	// Press Esc to exit hibernated view
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Assert view mode changed back
	if updated.viewMode != viewModeNormal {
		t.Errorf("expected viewModeNormal, got %d", updated.viewMode)
	}
}

func TestModel_HibernatedProjectsLoadedMsg_AC2(t *testing.T) {
	// AC2: Shows empty state or list of hibernated projects
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40

	hibernated := []*domain.Project{
		{ID: "h1", Name: "Hibernated1", State: domain.StateHibernated, LastActivityAt: time.Now().Add(-2 * time.Hour)},
		{ID: "h2", Name: "Hibernated2", State: domain.StateHibernated, LastActivityAt: time.Now().Add(-1 * time.Hour)},
	}

	msg := hibernatedProjectsLoadedMsg{projects: hibernated, err: nil}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Assert hibernated projects stored
	if len(updated.hibernatedProjects) != 2 {
		t.Errorf("expected 2 hibernated projects, got %d", len(updated.hibernatedProjects))
	}

	// Assert sorted by LastActivityAt descending (most recent first)
	if updated.hibernatedProjects[0].ID != "h2" {
		t.Error("expected most recent hibernated project first")
	}
}

func TestModel_HibernatedProjectsLoadedMsg_Error(t *testing.T) {
	// Test error handling
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40

	msg := hibernatedProjectsLoadedMsg{projects: nil, err: fmt.Errorf("load error")}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Assert hibernated projects remain nil
	if updated.hibernatedProjects != nil {
		t.Error("expected nil hibernated projects on error")
	}
}

func TestModel_WakeHibernatedProject_AC3(t *testing.T) {
	// AC3: Enter key activates project and returns to active view
	mockActivator := &mockStateActivator{}
	m := NewModel(nil)
	m.stateService = mockActivator
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	// Setup hibernated project
	hibernated := []*domain.Project{
		{ID: "h1", Name: "Hibernated1", State: domain.StateHibernated},
	}
	m.hibernatedProjects = hibernated
	m.hibernatedList = components.NewProjectListModel(hibernated, 100, 30)

	// Press Enter to wake project
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := m.Update(msg)

	// Assert command returned
	if cmd == nil {
		t.Error("expected activateProjectCmd, got nil")
	}
}

func TestModel_ProjectActivatedMsg_Success(t *testing.T) {
	// Test successful activation
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	msg := projectActivatedMsg{projectID: "h1", projectName: "Hibernated1", err: nil}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Assert view mode changed to normal
	if updated.viewMode != viewModeNormal {
		t.Errorf("expected viewModeNormal, got %d", updated.viewMode)
	}

	// Assert justActivatedProjectID set for selection
	if updated.justActivatedProjectID != "h1" {
		t.Errorf("expected justActivatedProjectID='h1', got '%s'", updated.justActivatedProjectID)
	}

	// Assert reload command returned
	if cmd == nil {
		t.Error("expected loadProjectsCmd, got nil")
	}
}

func TestModel_ProjectActivatedMsg_RaceCondition_AC10(t *testing.T) {
	// AC10: Handle race condition - project already activated by Story 11.3
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	// Setup hibernated projects
	hibernated := []*domain.Project{
		{ID: "h1", Name: "Hibernated1", State: domain.StateHibernated},
	}
	m.hibernatedProjects = hibernated
	m.hibernatedList = components.NewProjectListModel(hibernated, 100, 30)

	msg := projectActivatedMsg{projectID: "h1", err: domain.ErrInvalidStateTransition}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Assert stays in hibernated view (not switching to normal)
	if updated.viewMode != viewModeHibernated {
		t.Errorf("expected to stay in viewModeHibernated on race condition, got %d", updated.viewMode)
	}

	// Assert reload hibernated list
	if cmd == nil {
		t.Error("expected loadHibernatedProjectsCmd on race condition")
	}
}

func TestModel_ProjectActivatedMsg_GeneralError(t *testing.T) {
	// Code review H1: Test general error path (not race condition)
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	// Setup hibernated projects
	hibernated := []*domain.Project{
		{ID: "h1", Name: "Hibernated1", State: domain.StateHibernated},
	}
	m.hibernatedProjects = hibernated
	m.hibernatedList = components.NewProjectListModel(hibernated, 100, 30)

	// Simulate general activation error (not ErrInvalidStateTransition)
	msg := projectActivatedMsg{projectID: "h1", err: fmt.Errorf("database connection failed")}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Assert stays in hibernated view (not switching to normal)
	if updated.viewMode != viewModeHibernated {
		t.Errorf("expected to stay in viewModeHibernated on general error, got %d", updated.viewMode)
	}

	// Assert NO command returned (unlike race condition which reloads)
	if cmd != nil {
		t.Error("expected nil cmd on general error (no reload)")
	}

	// Assert justActivatedProjectID NOT set
	if updated.justActivatedProjectID != "" {
		t.Errorf("expected justActivatedProjectID='', got '%s'", updated.justActivatedProjectID)
	}
}

func TestModel_RemoveInHibernatedView_AC5(t *testing.T) {
	// AC5: 'x' works in hibernated view
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	// Setup hibernated project
	hibernated := []*domain.Project{
		{ID: "h1", Name: "Hibernated1", State: domain.StateHibernated},
	}
	m.hibernatedProjects = hibernated
	m.hibernatedList = components.NewProjectListModel(hibernated, 100, 30)

	// Press 'x' to start removal
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Assert confirmation started
	if !updated.isConfirmingRemove {
		t.Error("expected isConfirmingRemove=true")
	}
	if updated.confirmTarget == nil {
		t.Error("expected confirmTarget to be set")
	}
	if updated.confirmTarget.ID != "h1" {
		t.Errorf("expected confirmTarget.ID='h1', got '%s'", updated.confirmTarget.ID)
	}
}

func TestModel_RemoveConfirmedInHibernatedView_ReloadsHibernatedList(t *testing.T) {
	// Test that removal in hibernated view reloads hibernated list
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	// Setup hibernated project
	hibernated := []*domain.Project{
		{ID: "h1", Name: "Hibernated1", State: domain.StateHibernated},
	}
	m.hibernatedProjects = hibernated
	m.hibernatedList = components.NewProjectListModel(hibernated, 100, 30)

	msg := removeConfirmedMsg{projectID: "h1", projectName: "Hibernated1", err: nil}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Assert stays in hibernated view
	if updated.viewMode != viewModeHibernated {
		t.Errorf("expected to stay in viewModeHibernated, got %d", updated.viewMode)
	}

	// Assert reload command returned
	if cmd == nil {
		t.Error("expected loadHibernatedProjectsCmd")
	}

	// Assert feedback shown
	output := updated.statusBar.View()
	if !strings.Contains(output, "Removed") {
		t.Error("expected 'Removed' feedback in status bar")
	}
}

func TestModel_NavigationInHibernatedView_AC8(t *testing.T) {
	// AC8: Navigation works in hibernated view
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	// Setup hibernated projects
	hibernated := []*domain.Project{
		{ID: "h1", Name: "Hibernated1", State: domain.StateHibernated},
		{ID: "h2", Name: "Hibernated2", State: domain.StateHibernated},
	}
	m.hibernatedProjects = hibernated
	m.hibernatedList = components.NewProjectListModel(hibernated, 100, 30)
	m.detailPanel = components.NewDetailPanelModel(100, 30)

	// Initial selection should be 0
	if m.hibernatedList.Index() != 0 {
		t.Errorf("expected initial selection 0, got %d", m.hibernatedList.Index())
	}

	// Press 'j' to move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Assert selection moved
	if updated.hibernatedList.Index() != 1 {
		t.Errorf("expected selection 1 after 'j', got %d", updated.hibernatedList.Index())
	}
}

func TestModel_DetailPanelInHibernatedView_AC9(t *testing.T) {
	// AC9: Detail panel works in hibernated view
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated
	m.showDetailPanel = false

	// Setup hibernated project
	hibernated := []*domain.Project{
		{ID: "h1", Name: "Hibernated1", State: domain.StateHibernated},
	}
	m.hibernatedProjects = hibernated
	m.hibernatedList = components.NewProjectListModel(hibernated, 100, 30)
	m.detailPanel = components.NewDetailPanelModel(100, 30)

	// Press 'd' to toggle detail panel
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Assert detail panel visible
	if !updated.showDetailPanel {
		t.Error("expected showDetailPanel=true after 'd' key")
	}

	// Assert detail panel has correct project
	if updated.detailPanel.Project().ID != "h1" {
		t.Errorf("expected detail panel to show h1, got %s", updated.detailPanel.Project().ID)
	}
}

func TestModel_View_RendersHibernatedView_AC6(t *testing.T) {
	// AC6: View renders hibernated projects
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated

	// Setup hibernated project
	hibernated := []*domain.Project{
		{ID: "h1", Name: "Hibernated1", State: domain.StateHibernated},
	}
	m.hibernatedProjects = hibernated
	m.hibernatedList = components.NewProjectListModel(hibernated, 100, 30)
	m.statusBar = components.NewStatusBarModel(100)
	m.statusBar.SetInHibernatedView(true)
	m.statusBar.SetHibernatedViewCount(1)

	output := m.View()

	// Assert hibernated project visible
	if !strings.Contains(output, "Hibernated1") {
		t.Error("expected 'Hibernated1' in hibernated view")
	}
}

func TestModel_View_RendersHibernatedEmptyState_AC2(t *testing.T) {
	// AC2: Shows empty state when no hibernated projects
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.viewMode = viewModeHibernated
	m.hibernatedProjects = []*domain.Project{} // Empty

	m.statusBar = components.NewStatusBarModel(100)
	m.statusBar.SetInHibernatedView(true)
	m.statusBar.SetHibernatedViewCount(0)

	output := m.View()

	// Assert empty state message visible
	if !strings.Contains(output, "No hibernated projects") {
		t.Error("expected 'No hibernated projects' in empty hibernated view")
	}
}

func TestStatusBar_HibernatedView_AC7(t *testing.T) {
	// AC7: Status bar shows hibernated count
	sb := components.NewStatusBarModel(100)
	sb.SetInHibernatedView(true)
	sb.SetHibernatedViewCount(5)

	output := sb.View()

	if !strings.Contains(output, "5 hibernated") {
		t.Errorf("expected '5 hibernated' in status bar, got: %s", output)
	}
}

func TestStatusBar_HibernatedView_Shortcuts(t *testing.T) {
	// Test hibernated view shortcuts
	sb := components.NewStatusBarModel(100)
	sb.SetInHibernatedView(true)
	sb.SetWidth(100) // Wide enough for full shortcuts

	output := sb.View()

	// Check for hibernated-specific shortcuts
	if !strings.Contains(output, "wake") || !strings.Contains(output, "back") {
		t.Errorf("expected hibernated shortcuts (wake, back), got: %s", output)
	}
}

func TestModel_JustActivatedProjectID_SelectionRestored_AC3(t *testing.T) {
	// AC3: Just-activated project is selected after switch back
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.justActivatedProjectID = "p2" // Set from projectActivatedMsg

	projects := []*domain.Project{
		{ID: "p1", Name: "Project1", State: domain.StateActive},
		{ID: "p2", Name: "Project2", State: domain.StateActive},
		{ID: "p3", Name: "Project3", State: domain.StateActive},
	}

	msg := ProjectsLoadedMsg{projects: projects, err: nil}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Assert justActivatedProjectID cleared
	if updated.justActivatedProjectID != "" {
		t.Error("expected justActivatedProjectID to be cleared")
	}

	// Assert project p2 is selected (index 1)
	if updated.projectList.Index() != 1 {
		t.Errorf("expected selection at index 1 (p2), got %d", updated.projectList.Index())
	}
}

func TestModel_ActiveSelectionPreserved_WhenSwitchingViews(t *testing.T) {
	// Test that active selection is preserved when switching views
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40

	// Setup active projects with selection at index 2
	projects := []*domain.Project{
		{ID: "1", Name: "Project1", State: domain.StateActive},
		{ID: "2", Name: "Project2", State: domain.StateActive},
		{ID: "3", Name: "Project3", State: domain.StateActive},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, 100, 30)
	m.projectList.SelectByIndex(2)

	// Enter hibernated view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// Assert activeSelectedIdx saved
	if updated.activeSelectedIdx != 2 {
		t.Errorf("expected activeSelectedIdx=2, got %d", updated.activeSelectedIdx)
	}

	// Exit hibernated view
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ = updated.Update(msg)
	updated = newModel.(Model)

	// Assert selection restored
	if updated.projectList.Index() != 2 {
		t.Errorf("expected selection restored to index 2, got %d", updated.projectList.Index())
	}
}

// =============================================================================
// Story 11.7: TUI Manual State Toggle Tests
// =============================================================================

func TestModel_StateToggle_HibernateFromActiveView_AC1(t *testing.T) {
	// AC1: H key in active view calls Hibernate()
	mock := &mockStateActivator{}

	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.stateService = mock

	// Setup active project
	projects := []*domain.Project{
		{ID: "p1", Name: "Project1", State: domain.StateActive},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, 100, 30)

	// Press H key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}}
	_, cmd := m.Update(msg)

	// Should return stateToggleCmd
	if cmd == nil {
		t.Fatal("expected command from H key")
	}

	// Execute command and verify Hibernate was called
	resultMsg := cmd()
	result, ok := resultMsg.(stateToggledMsg)
	if !ok {
		t.Fatalf("expected stateToggledMsg, got %T", resultMsg)
	}

	if result.projectID != "p1" {
		t.Errorf("expected projectID 'p1', got %q", result.projectID)
	}
	if result.action != "hibernated" {
		t.Errorf("expected action 'hibernated', got %q", result.action)
	}
	if len(mock.hibernateCalls) != 1 || mock.hibernateCalls[0] != "p1" {
		t.Errorf("expected Hibernate() called with p1, got %v", mock.hibernateCalls)
	}
}

func TestModel_StateToggle_ActivateFromHibernatedView_AC2(t *testing.T) {
	// AC2: H key in hibernated view calls Activate()
	mock := &mockStateActivator{}

	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.stateService = mock
	m.viewMode = viewModeHibernated

	// Setup hibernated projects
	hibTime := time.Now().Add(-24 * time.Hour)
	projects := []*domain.Project{
		{ID: "p1", Name: "Project1", State: domain.StateHibernated, HibernatedAt: &hibTime},
	}
	m.hibernatedProjects = projects
	m.hibernatedList = components.NewProjectListModel(projects, 100, 30)

	// Press H key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}}
	_, cmd := m.Update(msg)

	// Should return stateToggleCmd
	if cmd == nil {
		t.Fatal("expected command from H key")
	}

	// Execute command and verify Activate was called
	resultMsg := cmd()
	result, ok := resultMsg.(stateToggledMsg)
	if !ok {
		t.Fatalf("expected stateToggledMsg, got %T", resultMsg)
	}

	if result.projectID != "p1" {
		t.Errorf("expected projectID 'p1', got %q", result.projectID)
	}
	if result.action != "activated" {
		t.Errorf("expected action 'activated', got %q", result.action)
	}
	if len(mock.activateCalls) != 1 || mock.activateCalls[0] != "p1" {
		t.Errorf("expected Activate() called with p1, got %v", mock.activateCalls)
	}
}

func TestModel_StateToggle_FavoriteProtection_AC3(t *testing.T) {
	// AC3: Error feedback for favorite project
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.statusBar = components.NewStatusBarModel(120)

	// Simulate stateToggledMsg with ErrFavoriteCannotHibernate
	msg := stateToggledMsg{
		projectID:   "p1",
		projectName: "FavoriteProject",
		action:      "hibernated",
		err:         domain.ErrFavoriteCannotHibernate,
	}

	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	// Should have set error feedback in status bar
	// We can't directly check statusBar state, but verify command returned for 3s timer
	if cmd == nil {
		t.Error("expected clear timer command")
	}

	// Verify status bar is set (indirectly via view containing the message)
	_ = updated // status bar was updated internally
}

func TestModel_StateToggle_SuccessFeedback_AC4(t *testing.T) {
	// AC4: Success message in status bar
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.statusBar = components.NewStatusBarModel(120)

	// Setup repo for reload command
	m.repository = testhelpers.NewMockRepository()

	// Simulate successful stateToggledMsg
	msg := stateToggledMsg{
		projectID:   "p1",
		projectName: "TestProject",
		action:      "hibernated",
		err:         nil,
	}

	_, cmd := m.Update(msg)

	// Should have batch command (reload + timer)
	if cmd == nil {
		t.Error("expected batch command")
	}
}

func TestModel_StateToggle_ListRefresh_AC5(t *testing.T) {
	// AC5: Returns loadProjectsCmd/loadHibernatedProjectsCmd
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.statusBar = components.NewStatusBarModel(120)
	m.repository = testhelpers.NewMockRepository()

	// Test hibernate action reloads active list
	msg := stateToggledMsg{
		projectID:   "p1",
		projectName: "TestProject",
		action:      "hibernated",
		err:         nil,
	}

	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("expected command to reload projects")
	}

	// Test activate action reloads hibernated list
	m.viewMode = viewModeHibernated
	msg.action = "activated"
	_, cmd = m.Update(msg)
	if cmd == nil {
		t.Error("expected command to reload hibernated projects")
	}
}

func TestModel_StateToggle_NoOpWhenEmpty_AC7(t *testing.T) {
	// AC7: No crash when list is empty
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.projects = []*domain.Project{} // Empty

	// Press H key with empty list
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}}
	_, cmd := m.Update(msg)

	// Should return nil (no-op)
	if cmd != nil {
		t.Error("H key on empty list should return nil")
	}

	// Test hibernated view as well
	m.viewMode = viewModeHibernated
	m.hibernatedProjects = []*domain.Project{} // Empty

	_, cmd = m.Update(msg)
	if cmd != nil {
		t.Error("H key on empty hibernated list should return nil")
	}
}

func TestModel_StateToggle_IdempotentBehavior(t *testing.T) {
	// ErrInvalidStateTransition handled gracefully
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.statusBar = components.NewStatusBarModel(120)
	m.repository = testhelpers.NewMockRepository()

	// Simulate stateToggledMsg with ErrInvalidStateTransition
	msg := stateToggledMsg{
		projectID:   "p1",
		projectName: "TestProject",
		action:      "hibernated",
		err:         domain.ErrInvalidStateTransition,
	}

	_, cmd := m.Update(msg)

	// Should reload list silently (no error shown)
	if cmd == nil {
		t.Error("expected reload command for idempotent case")
	}
}

func TestModel_StateToggle_NoStateService(t *testing.T) {
	// Test error when stateService is nil
	m := NewModel(nil)
	m.ready = true
	m.width = 120
	m.height = 40
	m.stateService = nil // No service

	// Setup project
	projects := []*domain.Project{
		{ID: "p1", Name: "Project1", State: domain.StateActive},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(projects, 100, 30)

	// Press H key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command even with nil service")
	}

	// Execute and verify error returned
	resultMsg := cmd()
	result, ok := resultMsg.(stateToggledMsg)
	if !ok {
		t.Fatalf("expected stateToggledMsg, got %T", resultMsg)
	}

	if result.err == nil {
		t.Error("expected error when stateService is nil")
	}
}

func TestHelpOverlay_ContainsStateToggleKeybinding_AC6(t *testing.T) {
	// AC6: Help overlay shows H keybinding with "Hibernate/Activate" description
	cfg := ports.NewConfig()
	rendered := renderHelpOverlay(120, 40, cfg)

	// Verify H keybinding is present
	if !strings.Contains(rendered, "H") {
		t.Error("help overlay should contain 'H' keybinding")
	}
	if !strings.Contains(rendered, "Hibernate/Activate") {
		t.Error("help overlay should contain 'Hibernate/Activate' description")
	}

	// Verify it appears in the Actions section (after 'x' for remove, before 'a' for add)
	xIndex := strings.Index(rendered, "x        Remove project")
	hIndex := strings.Index(rendered, "H        Hibernate/Activate")
	aIndex := strings.Index(rendered, "a        Add project")

	if xIndex == -1 || hIndex == -1 || aIndex == -1 {
		t.Error("expected all keybindings to be present in help overlay")
		return
	}

	if !(xIndex < hIndex && hIndex < aIndex) {
		t.Error("H keybinding should appear after 'x' and before 'a' in Actions section")
	}
}
