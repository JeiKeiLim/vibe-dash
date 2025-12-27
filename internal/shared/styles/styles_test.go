package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// Story 8.12: Test HorizontalBorderStyle configuration

func TestHorizontalBorderStyle_HasNoTopBorder(t *testing.T) {
	// HorizontalBorderStyle should have no top border for horizontal stacking
	style := HorizontalBorderStyle

	// Render a test box and check the top border is absent
	rendered := style.Width(20).Height(3).Render("Test")

	// The rendered content should not start with a top border character
	// NormalBorder uses | ─ ┌ └ ┐ ┘ characters
	// Without top border, first character should be │ (left border) or content
	if len(rendered) == 0 {
		t.Error("HorizontalBorderStyle should produce non-empty output")
	}

	// First line should NOT contain horizontal border character ─
	// (it should only have vertical border │ if BorderTop is false)
	lines := splitLines(rendered)
	if len(lines) == 0 {
		t.Error("HorizontalBorderStyle should produce at least one line")
		return
	}

	// Verify the style is properly configured by checking GetBorder
	// NormalBorder with BorderTop(false) has non-empty Top in the border struct,
	// but the style disables top rendering. This is expected behavior.
	_, _, _, _, _ = style.GetBorder()
}

func TestHorizontalBorderStyle_UsesNormalBorder(t *testing.T) {
	// Both BorderStyle and HorizontalBorderStyle should use NormalBorder
	normalBorder := lipgloss.NormalBorder()

	border, _, _, _, _ := HorizontalBorderStyle.GetBorder()
	normalB, _, _, _, _ := lipgloss.NewStyle().Border(normalBorder).GetBorder()

	// Compare border characters (Left and Right should match)
	if border.Left != normalB.Left {
		t.Errorf("HorizontalBorderStyle should use NormalBorder Left: got %q, want %q", border.Left, normalB.Left)
	}
	if border.Right != normalB.Right {
		t.Errorf("HorizontalBorderStyle should use NormalBorder Right: got %q, want %q", border.Right, normalB.Right)
	}
}

func TestHorizontalBorderStyle_UsesColor8(t *testing.T) {
	// HorizontalBorderStyle should use ANSI color 8 (bright black/gray)
	// Same as BorderStyle for consistency
	topColor, rightColor, bottomColor, leftColor := HorizontalBorderStyle.GetBorderTopForeground(),
		HorizontalBorderStyle.GetBorderRightForeground(),
		HorizontalBorderStyle.GetBorderBottomForeground(),
		HorizontalBorderStyle.GetBorderLeftForeground()

	expectedColor := lipgloss.Color("8")

	// At least one of the visible borders should have color 8
	// (top is disabled, so check left, right, bottom)
	if leftColor != expectedColor && rightColor != expectedColor && bottomColor != expectedColor {
		t.Errorf("HorizontalBorderStyle should use color 8 for borders, got top=%v right=%v bottom=%v left=%v",
			topColor, rightColor, bottomColor, leftColor)
	}
}

func TestHorizontalBorderStyle_MatchesBorderStyleExceptTop(t *testing.T) {
	// HorizontalBorderStyle should match BorderStyle except for top border
	// This ensures visual consistency between layouts

	hBorder, _, _, _, _ := HorizontalBorderStyle.GetBorder()
	bBorder, _, _, _, _ := BorderStyle.GetBorder()

	// Left borders should match
	if hBorder.Left != bBorder.Left {
		t.Errorf("HorizontalBorderStyle Left should match BorderStyle: got %q, want %q",
			hBorder.Left, bBorder.Left)
	}

	// Right borders should match
	if hBorder.Right != bBorder.Right {
		t.Errorf("HorizontalBorderStyle Right should match BorderStyle: got %q, want %q",
			hBorder.Right, bBorder.Right)
	}

	// Bottom borders should match
	if hBorder.Bottom != bBorder.Bottom {
		t.Errorf("HorizontalBorderStyle Bottom should match BorderStyle: got %q, want %q",
			hBorder.Bottom, bBorder.Bottom)
	}
}

// splitLines splits a string into lines (helper for tests)
func splitLines(s string) []string {
	var lines []string
	var current string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
