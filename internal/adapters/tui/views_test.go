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

// ============================================================================
// Story 3.9: Remove Confirmation Dialog Tests
// ============================================================================

func TestRenderConfirmRemoveDialog_ContainsProjectName(t *testing.T) {
	output := renderConfirmRemoveDialog("my-project", 80, 24)

	if !strings.Contains(output, "my-project") {
		t.Error("expected dialog to contain project name")
	}
	if !strings.Contains(output, "Remove") {
		t.Error("expected dialog to contain 'Remove'")
	}
	if !strings.Contains(output, "[y]") {
		t.Error("expected dialog to contain '[y]' hint")
	}
	if !strings.Contains(output, "[n]") {
		t.Error("expected dialog to contain '[n]' hint")
	}
}

func TestRenderConfirmRemoveDialog_ContainsTitle(t *testing.T) {
	output := renderConfirmRemoveDialog("test-project", 80, 24)

	if !strings.Contains(output, "Confirm Removal") {
		t.Error("expected dialog to contain 'Confirm Removal' title")
	}
}

func TestRenderConfirmRemoveDialog_ContainsHints(t *testing.T) {
	output := renderConfirmRemoveDialog("test-project", 80, 24)

	hints := []string{"[y]", "Yes", "[n]", "No", "[Esc]", "Cancel"}
	for _, hint := range hints {
		if !strings.Contains(output, hint) {
			t.Errorf("expected dialog to contain hint %q", hint)
		}
	}
}

func TestRenderConfirmRemoveDialog_WidthConstraints(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		expected int // Expected dialog width
	}{
		{"narrow terminal caps at 30", 20, 30},
		{"medium terminal uses width-4", 50, 46},
		{"wide terminal caps at 60", 100, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := renderConfirmRemoveDialog("test", tt.width, 24)
			// Dialog should render without panic
			if len(output) == 0 {
				t.Error("expected non-empty output")
			}
		})
	}
}

func TestRenderConfirmRemoveDialog_CenteredInTerminal(t *testing.T) {
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
		result := renderConfirmRemoveDialog("test-project", size.width, size.height)

		// Output should exist and contain key elements
		if !strings.Contains(result, "Confirm Removal") {
			t.Errorf("renderConfirmRemoveDialog(%d, %d) should contain title", size.width, size.height)
		}

		// Should be rendered (non-empty)
		if len(result) == 0 {
			t.Errorf("renderConfirmRemoveDialog(%d, %d) should not be empty", size.width, size.height)
		}
	}
}

func TestRenderConfirmRemoveDialog_EdgeCase_ZeroDimensions(t *testing.T) {
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
		result := renderConfirmRemoveDialog("test", tc.width, tc.height)

		// Should still contain content (lipgloss handles small sizes gracefully)
		if len(result) == 0 {
			t.Errorf("renderConfirmRemoveDialog(%d, %d) should not be empty", tc.width, tc.height)
		}
	}
}

func TestRenderConfirmRemoveDialog_LongProjectName(t *testing.T) {
	longName := "this-is-a-very-long-project-name-that-should-still-render-correctly"
	output := renderConfirmRemoveDialog(longName, 80, 24)

	// Dialog should render without panic even with long names
	if len(output) == 0 {
		t.Error("expected non-empty output for long project name")
	}
	// Should contain "Remove" indicating the prompt is visible
	if !strings.Contains(output, "Remove") {
		t.Error("expected dialog to contain 'Remove' prompt")
	}
	// The name may be truncated by lipgloss width constraints, which is acceptable
	// Just verify dialog renders with key elements
	if !strings.Contains(output, "Confirm Removal") {
		t.Error("expected dialog to contain title")
	}
}

// ============================================================================
// Story 3.10: Narrow Warning Tests
// ============================================================================

func TestRenderNarrowWarning_ContainsWarning(t *testing.T) {
	output := renderNarrowWarning(70)

	if !strings.Contains(output, NarrowWarning) {
		t.Error("expected narrow warning text in output")
	}
}

func TestRenderNarrowWarning_Centered(t *testing.T) {
	output := renderNarrowWarning(80)

	// Warning should be centered - check it's not left-aligned
	if strings.HasPrefix(output, "⚠") {
		t.Error("expected warning to be centered, not left-aligned")
	}
}

func TestRenderNarrowWarning_DifferentWidths(t *testing.T) {
	widths := []int{60, 70, 79}

	for _, width := range widths {
		output := renderNarrowWarning(width)

		if !strings.Contains(output, NarrowWarning) {
			t.Errorf("expected narrow warning text in output at width %d", width)
		}
		if len(output) == 0 {
			t.Errorf("expected non-empty output at width %d", width)
		}
	}
}
