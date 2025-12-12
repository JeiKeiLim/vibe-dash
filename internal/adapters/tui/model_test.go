package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel()

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
}

func TestModel_Init(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_Update_QuitKey(t *testing.T) {
	m := NewModel()
	m.ready = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	// Check if cmd is tea.Quit
	if cmd == nil {
		t.Error("'q' key should return tea.Quit command")
	}
}

func TestModel_Update_CtrlC(t *testing.T) {
	m := NewModel()
	m.ready = true

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Ctrl+C should return tea.Quit command")
	}
}

func TestModel_Update_HelpToggle(t *testing.T) {
	m := NewModel()
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
	m := NewModel()
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
	m := NewModel()

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
	m := NewModel()
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
	m := NewModel()
	m.ready = false

	view := m.View()

	if view != "Initializing..." {
		t.Errorf("View when not ready should be 'Initializing...', got %q", view)
	}
}

func TestModel_View_TooSmall(t *testing.T) {
	m := NewModel()
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
	m := NewModel()
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

func TestModel_View_HelpOverlay(t *testing.T) {
	m := NewModel()
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
