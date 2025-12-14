package timeformat

import (
	"testing"
	"time"
)

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"zero time", time.Time{}, "never"},
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute), "1m ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5m ago"},
		{"59 minutes ago", now.Add(-59 * time.Minute), "59m ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1h ago"},
		{"2 hours ago", now.Add(-2 * time.Hour), "2h ago"},
		{"23 hours ago", now.Add(-23 * time.Hour), "23h ago"},
		{"1 day ago", now.Add(-25 * time.Hour), "1d ago"},
		{"3 days ago", now.Add(-3 * 24 * time.Hour), "3d ago"},
		{"6 days ago", now.Add(-6 * 24 * time.Hour), "6d ago"},
		{"1 week ago", now.Add(-8 * 24 * time.Hour), "1w ago"},
		{"2 weeks ago", now.Add(-15 * 24 * time.Hour), "2w ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelativeTime(tt.time)
			if got != tt.expected {
				t.Errorf("FormatRelativeTime() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRecencyIndicator(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		activity time.Time
		expected string
	}{
		{"zero time", time.Time{}, ""},
		{"just now", now, "✨"},
		{"1 hour ago", now.Add(-1 * time.Hour), "✨"},
		{"12 hours ago", now.Add(-12 * time.Hour), "✨"},
		{"23 hours ago", now.Add(-23 * time.Hour), "✨"},
		{"exactly 24 hours ago", now.Add(-24 * time.Hour), "⚡"},
		{"25 hours ago", now.Add(-25 * time.Hour), "⚡"},
		{"2 days ago", now.Add(-2 * 24 * time.Hour), "⚡"},
		{"6 days ago", now.Add(-6 * 24 * time.Hour), "⚡"},
		{"exactly 7 days ago", now.Add(-7 * 24 * time.Hour), ""},
		{"8 days ago", now.Add(-8 * 24 * time.Hour), ""},
		{"30 days ago", now.Add(-30 * 24 * time.Hour), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RecencyIndicator(tt.activity)
			if got != tt.expected {
				t.Errorf("RecencyIndicator() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRecencyIndicator_FutureTime(t *testing.T) {
	// Future time should return "✨" since time.Since returns negative duration
	// which is less than 24 hours
	future := time.Now().Add(1 * time.Hour)
	got := RecencyIndicator(future)
	if got != "✨" {
		t.Errorf("RecencyIndicator(future) = %q, want %q", got, "✨")
	}
}
