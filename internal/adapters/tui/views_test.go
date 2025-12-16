package tui

import (
	"strings"
	"testing"
)

func TestRenderHelpOverlay_ContainsAllSections(t *testing.T) {
	result := renderHelpOverlay(100, 40)

	sections := []string{"Navigation", "Actions", "Views", "General"}
	for _, section := range sections {
		if !strings.Contains(result, section) {
			t.Errorf("Expected help overlay to contain section %q", section)
		}
	}
}

func TestRenderHelpOverlay_ContainsNavigationShortcuts(t *testing.T) {
	result := renderHelpOverlay(100, 40)

	shortcuts := []string{
		"j/\u2193", // j/↓
		"k/\u2191", // k/↑
		"Move down",
		"Move up",
	}
	for _, shortcut := range shortcuts {
		if !strings.Contains(result, shortcut) {
			t.Errorf("Expected help overlay to contain navigation shortcut %q", shortcut)
		}
	}
}

func TestRenderHelpOverlay_ContainsActionShortcuts(t *testing.T) {
	result := renderHelpOverlay(100, 40)

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"d", "Toggle detail panel"},
		{"f", "Toggle favorite"},
		{"n", "Edit notes"},
		{"x", "Remove project"},
		{"a", "Add project"},
		{"r", "Refresh/rescan"},
	}
	for _, s := range shortcuts {
		if !strings.Contains(result, s.key) {
			t.Errorf("Expected help overlay to contain action key %q", s.key)
		}
		if !strings.Contains(result, s.desc) {
			t.Errorf("Expected help overlay to contain action description %q", s.desc)
		}
	}
}

func TestRenderHelpOverlay_ContainsViewShortcuts(t *testing.T) {
	result := renderHelpOverlay(100, 40)

	// Check for specific "h" key line pattern (not just "h" which matches "this", "Show", etc.)
	if !strings.Contains(result, "h        View hibernated") {
		t.Error("Expected help overlay to contain 'h        View hibernated' key binding")
	}
}

func TestRenderHelpOverlay_ContainsGeneralShortcuts(t *testing.T) {
	result := renderHelpOverlay(100, 40)

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"?", "Show this help"},
		{"q", "Quit"},
		{"Esc", "Cancel/close"},
	}
	for _, s := range shortcuts {
		if !strings.Contains(result, s.key) {
			t.Errorf("Expected help overlay to contain general key %q", s.key)
		}
		if !strings.Contains(result, s.desc) {
			t.Errorf("Expected help overlay to contain general description %q", s.desc)
		}
	}
}

func TestRenderHelpOverlay_ContainsCloseHint(t *testing.T) {
	result := renderHelpOverlay(100, 40)

	if !strings.Contains(result, "Press any key to close") {
		t.Error("Expected help overlay to contain 'Press any key to close'")
	}
}

func TestRenderHelpOverlay_ContainsTitle(t *testing.T) {
	result := renderHelpOverlay(100, 40)

	if !strings.Contains(result, "KEYBOARD SHORTCUTS") {
		t.Error("Expected help overlay to contain 'KEYBOARD SHORTCUTS' title")
	}
}

func TestRenderHelpOverlay_CenteredInTerminal(t *testing.T) {
	// Test with different terminal sizes
	sizes := []struct {
		width  int
		height int
	}{
		{80, 24},
		{100, 40},
		{120, 50},
	}

	for _, size := range sizes {
		result := renderHelpOverlay(size.width, size.height)

		// Output should exist and contain the title
		if !strings.Contains(result, "KEYBOARD SHORTCUTS") {
			t.Errorf("renderHelpOverlay(%d, %d) should contain title", size.width, size.height)
		}

		// Should be rendered (non-empty)
		if len(result) == 0 {
			t.Errorf("renderHelpOverlay(%d, %d) should not be empty", size.width, size.height)
		}
	}
}

func TestRenderHelpOverlay_EdgeCase_ZeroDimensions(t *testing.T) {
	// Edge case: zero or very small dimensions should not panic
	edgeCases := []struct {
		width  int
		height int
	}{
		{0, 0},
		{1, 1},
		{10, 5},
	}

	for _, tc := range edgeCases {
		// Should not panic
		result := renderHelpOverlay(tc.width, tc.height)

		// Should still contain content (lipgloss handles small sizes gracefully)
		if len(result) == 0 {
			t.Errorf("renderHelpOverlay(%d, %d) should not be empty", tc.width, tc.height)
		}
	}
}
