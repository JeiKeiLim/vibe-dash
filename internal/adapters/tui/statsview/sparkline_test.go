package statsview

import (
	"testing"
)

func TestRenderSparkline_EmptyInput(t *testing.T) {
	result := RenderSparkline(nil)
	if result != "" {
		t.Errorf("expected empty string for nil input, got %q", result)
	}

	result = RenderSparkline([]int{})
	if result != "" {
		t.Errorf("expected empty string for empty slice, got %q", result)
	}
}

func TestRenderSparkline_AllZeros(t *testing.T) {
	tests := []struct {
		name   string
		counts []int
		want   string
	}{
		{"single zero", []int{0}, "▁"},
		{"three zeros", []int{0, 0, 0}, "▁▁▁"},
		{"seven zeros", []int{0, 0, 0, 0, 0, 0, 0}, "▁▁▁▁▁▁▁"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderSparkline(tt.counts)
			if result != tt.want {
				t.Errorf("RenderSparkline(%v) = %q, want %q", tt.counts, result, tt.want)
			}
		})
	}
}

func TestRenderSparkline_SingleValue(t *testing.T) {
	// Single non-zero value should map to middle char
	result := RenderSparkline([]int{5})
	// Single value with max=5 normalizes to level 0 (the value itself becomes the max)
	// Actually, single value should show highest level since it's the max
	if result != "█" {
		t.Errorf("expected '█' for single non-zero value, got %q", result)
	}

	// Single value of 1
	result = RenderSparkline([]int{1})
	if result != "█" {
		t.Errorf("expected '█' for single value of 1, got %q", result)
	}
}

func TestRenderSparkline_IncreasingValues(t *testing.T) {
	// Values 0,1,2,3,4,5,6,7 should map to all 8 characters
	counts := []int{0, 1, 2, 3, 4, 5, 6, 7}
	result := RenderSparkline(counts)

	// Each position should have increasing height
	expected := "▁▂▃▄▅▆▇█"
	if result != expected {
		t.Errorf("RenderSparkline(%v) = %q, want %q", counts, result, expected)
	}
}

func TestRenderSparkline_MaxValues(t *testing.T) {
	// All max values should render as highest bar
	counts := []int{10, 10, 10}
	result := RenderSparkline(counts)
	if result != "███" {
		t.Errorf("expected '███' for all max values, got %q", result)
	}
}

func TestRenderSparkline_MixedValues(t *testing.T) {
	// Test with realistic sparkline pattern
	counts := []int{2, 5, 1, 8, 3, 7}
	result := RenderSparkline(counts)

	// Verify length matches input
	if len([]rune(result)) != len(counts) {
		t.Errorf("expected %d chars, got %d", len(counts), len([]rune(result)))
	}

	// Verify max value maps to highest char
	runes := []rune(result)
	if runes[3] != '█' { // index 3 has value 8 (max)
		t.Errorf("expected max value (8) to map to '█', got %c", runes[3])
	}
}

func TestRenderSparkline_LargeValues(t *testing.T) {
	// Large values should still normalize correctly
	counts := []int{100, 200, 300, 400, 500, 600, 700, 800}
	result := RenderSparkline(counts)

	expected := "▁▂▃▄▅▆▇█"
	if result != expected {
		t.Errorf("large values: got %q, want %q", result, expected)
	}
}

func TestRenderSparkline_WithZeroMax(t *testing.T) {
	// When some values are zero but there's a max
	counts := []int{0, 0, 5, 0}
	result := RenderSparkline(counts)
	runes := []rune(result)

	// Zero values should be lowest char
	if runes[0] != '▁' {
		t.Errorf("expected zero to map to '▁', got %c", runes[0])
	}

	// Max value should be highest char
	if runes[2] != '█' {
		t.Errorf("expected max to map to '█', got %c", runes[2])
	}
}
