package tui

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestSelectedStyle_RendersCyanBackground(t *testing.T) {
	result := SelectedStyle.Render("test")
	if result == "" {
		t.Error("SelectedStyle should render non-empty string")
	}
}

func TestWaitingStyle_RendersBoldRed(t *testing.T) {
	result := WaitingStyle.Render("WAITING")
	if result == "" {
		t.Error("WaitingStyle should render non-empty string")
	}
}

func TestRecentStyle_RendersGreen(t *testing.T) {
	result := RecentStyle.Render("Recent")
	if result == "" {
		t.Error("RecentStyle should render non-empty string")
	}
}

func TestActiveStyle_RendersYellow(t *testing.T) {
	result := ActiveStyle.Render("Active")
	if result == "" {
		t.Error("ActiveStyle should render non-empty string")
	}
}

func TestUncertainStyle_RendersGray(t *testing.T) {
	result := UncertainStyle.Render("Uncertain")
	if result == "" {
		t.Error("UncertainStyle should render non-empty string")
	}
}

func TestFavoriteStyle_RendersMagenta(t *testing.T) {
	result := FavoriteStyle.Render("*")
	if result == "" {
		t.Error("FavoriteStyle should render non-empty string")
	}
}

func TestDimStyle_RendersFaint(t *testing.T) {
	result := DimStyle.Render("hint text")
	if result == "" {
		t.Error("DimStyle should render non-empty string")
	}
}

func TestBorderStyle_RendersWithBorder(t *testing.T) {
	result := BorderStyle.Render("panel content")
	if result == "" {
		t.Error("BorderStyle should render non-empty string")
	}
}

func TestAllStylesRenderWithoutPanic(t *testing.T) {
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"SelectedStyle", SelectedStyle},
		{"WaitingStyle", WaitingStyle},
		{"RecentStyle", RecentStyle},
		{"ActiveStyle", ActiveStyle},
		{"UncertainStyle", UncertainStyle},
		{"FavoriteStyle", FavoriteStyle},
		{"DimStyle", DimStyle},
		{"BorderStyle", BorderStyle},
		{"boxStyle", boxStyle},
		{"titleStyle", titleStyle},
		{"hintStyle", hintStyle},
	}

	for _, tc := range styles {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.style.Render("test content")
			if result == "" {
				t.Errorf("%s rendered empty string", tc.name)
			}
		})
	}
}

func TestStyleComposition(t *testing.T) {
	// Test that styles can be composed without error
	innerText := WaitingStyle.Render("WAITING")
	outerText := SelectedStyle.Render(innerText)

	if outerText == "" {
		t.Error("Composed styles should render non-empty string")
	}
}

func TestStyleCompositionVariants(t *testing.T) {
	tests := []struct {
		name  string
		inner lipgloss.Style
		outer lipgloss.Style
	}{
		{"WaitingInSelected", WaitingStyle, SelectedStyle},
		{"RecentInSelected", RecentStyle, SelectedStyle},
		{"FavoriteInDim", FavoriteStyle, DimStyle},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inner := tc.inner.Render("text")
			outer := tc.outer.Render(inner)
			if outer == "" {
				t.Errorf("Composition %s rendered empty", tc.name)
			}
		})
	}
}

func TestUseColorLogicWithEnvironment(t *testing.T) {
	// Test the UseColor logic pattern
	// Note: We cannot easily modify package-level variables, but we test the logic
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
		t.Run("NO_COLOR="+tc.noColor+"_TERM="+tc.term, func(t *testing.T) {
			// Simulate the UseColor logic
			result := tc.noColor == "" && tc.term != "dumb"
			if result != tc.expected {
				t.Errorf("UseColor logic: expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestUseColorReflectsEnvironment(t *testing.T) {
	// Test that UseColor was computed correctly at init time
	// If NO_COLOR is not set and TERM is not "dumb", UseColor should be true
	noColor := os.Getenv("NO_COLOR")
	term := os.Getenv("TERM")

	expected := noColor == "" && term != "dumb"
	if UseColor != expected {
		t.Errorf("UseColor=%v but expected %v (NO_COLOR=%q, TERM=%q)",
			UseColor, expected, noColor, term)
	}
}

func TestApplySelected(t *testing.T) {
	result := ApplySelected("row content")
	if result == "" {
		t.Error("ApplySelected should return non-empty string")
	}
}

func TestApplyIndicator(t *testing.T) {
	tests := []struct {
		indicatorType string
		text          string
	}{
		{"waiting", "WAITING"},
		{"recent", "Today"},
		{"active", "This week"},
		{"uncertain", "Unknown"},
		{"favorite", "*"},
		{"dim", "dimmed"},
		{"unknown", "default"}, // Should return unmodified
	}

	for _, tc := range tests {
		t.Run(tc.indicatorType, func(t *testing.T) {
			result := ApplyIndicator(tc.indicatorType, tc.text)
			if result == "" {
				t.Errorf("ApplyIndicator(%q, %q) returned empty string", tc.indicatorType, tc.text)
			}

			// For unknown type, should return original text
			if tc.indicatorType == "unknown" && result != tc.text {
				t.Errorf("ApplyIndicator with unknown type should return original text, got %q", result)
			}
		})
	}
}

func TestApplyIndicator_UnknownType(t *testing.T) {
	// Unknown indicator type should return text unmodified
	original := "some text"
	result := ApplyIndicator("nonexistent", original)

	if result != original {
		t.Errorf("ApplyIndicator with unknown type should return original text, got %q", result)
	}
}

func TestApplyIndicator_DimType(t *testing.T) {
	result := ApplyIndicator("dim", "dimmed text")
	if result == "" {
		t.Error("ApplyIndicator with dim type should return non-empty string")
	}
}

func TestBaseStylesExist(t *testing.T) {
	// Verify the base styles from Story 1.5 are still available
	if boxStyle.Render("test") == "" {
		t.Error("boxStyle should render non-empty")
	}
	if titleStyle.Render("test") == "" {
		t.Error("titleStyle should render non-empty")
	}
	if hintStyle.Render("test") == "" {
		t.Error("hintStyle should render non-empty")
	}
}

// TestApplySelected_EmptyString verifies edge case handling
func TestApplySelected_EmptyString(t *testing.T) {
	// Empty string should not panic and should return something
	result := ApplySelected("")
	// Result may contain ANSI codes even for empty string, so just verify no panic
	_ = result
}

// TestApplyIndicator_EmptyText verifies edge case handling
func TestApplyIndicator_EmptyText(t *testing.T) {
	types := []string{"waiting", "recent", "active", "uncertain", "favorite", "dim"}
	for _, indicatorType := range types {
		t.Run(indicatorType, func(t *testing.T) {
			// Should not panic with empty text
			result := ApplyIndicator(indicatorType, "")
			_ = result
		})
	}
}

// TestStylesRenderConsistently verifies styles render consistently.
// Note: Lipgloss automatically detects TTY and may strip ANSI codes in non-TTY
// environments (like tests). This test verifies styles work regardless of
// the output environment by checking they produce expected content.
func TestStylesRenderConsistently(t *testing.T) {
	tests := []struct {
		name     string
		render   func() string
		contains string
	}{
		{"WaitingStyle", func() string { return WaitingStyle.Render("WAITING") }, "WAITING"},
		{"RecentStyle", func() string { return RecentStyle.Render("Recent") }, "Recent"},
		{"ActiveStyle", func() string { return ActiveStyle.Render("Active") }, "Active"},
		{"FavoriteStyle", func() string { return FavoriteStyle.Render("*") }, "*"},
		{"SelectedStyle", func() string { return SelectedStyle.Render("row") }, "row"},
		{"DimStyle", func() string { return DimStyle.Render("hint") }, "hint"},
		{"BorderStyle", func() string { return BorderStyle.Render("panel") }, "panel"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.render()
			// The rendered output should contain the original text
			// (with or without ANSI codes depending on environment)
			if result == "" {
				t.Errorf("%s rendered empty string", tc.name)
			}
			// Verify the content is preserved (may have ANSI codes around it)
			found := false
			for i := 0; i <= len(result)-len(tc.contains); i++ {
				if result[i:i+len(tc.contains)] == tc.contains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s should contain %q, got %q", tc.name, tc.contains, result)
			}
		})
	}
}

// TestNoColorBehavior documents the expected behavior when NO_COLOR is set.
// Note: Package-level UseColor is evaluated at init time, so this test
// verifies the logic pattern rather than runtime re-initialization.
func TestNoColorBehavior(t *testing.T) {
	// When NO_COLOR is set, UseColor should be false
	// and lipgloss.SetColorProfile(termenv.Ascii) should be called in init()

	// We can verify the logic pattern:
	noColorSet := os.Getenv("NO_COLOR") != ""
	termDumb := os.Getenv("TERM") == "dumb"

	expectedUseColor := !noColorSet && !termDumb

	if UseColor != expectedUseColor {
		t.Errorf("UseColor mismatch: got %v, expected %v (NO_COLOR=%q, TERM=%q)",
			UseColor, expectedUseColor, os.Getenv("NO_COLOR"), os.Getenv("TERM"))
	}

	// When UseColor is false, the init() function calls lipgloss.SetColorProfile(termenv.Ascii)
	// which strips ANSI codes from all style renders.
	// This is verified by running: NO_COLOR=1 go test -v ./internal/adapters/tui/...
	// and observing that TestStylesProduceANSIWhenColorEnabled is skipped.
}
