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

func TestFormatWaitingDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		// Minutes
		{"0 duration", 0, "0m"},
		{"1 minute", 1 * time.Minute, "1m"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"15 minutes", 15 * time.Minute, "15m"},
		{"59 minutes", 59 * time.Minute, "59m"},

		// Hours (boundary at 60 minutes)
		{"60 minutes = 1h", 60 * time.Minute, "1h"},
		{"1 hour", 1 * time.Hour, "1h"},
		{"2 hours", 2 * time.Hour, "2h"},
		{"12 hours", 12 * time.Hour, "12h"},
		{"23 hours", 23 * time.Hour, "23h"},

		// Days (boundary at 24 hours)
		{"24 hours = 1d", 24 * time.Hour, "1d"},
		{"25 hours", 25 * time.Hour, "1d"},
		{"48 hours = 2d", 48 * time.Hour, "2d"},
		{"7 days", 7 * 24 * time.Hour, "7d"},
		{"30 days", 30 * 24 * time.Hour, "30d"},

		// Edge cases
		{"59m59s still shows 59m", 59*time.Minute + 59*time.Second, "59m"},
		{"1h30m shows 1h", 1*time.Hour + 30*time.Minute, "1h"},
		{"47h truncates to 1d", 47 * time.Hour, "1d"},
		{"negative duration shows 0m", -5 * time.Minute, "0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatWaitingDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("FormatWaitingDuration(%v) = %q, want %q", tt.duration, got, tt.expected)
			}
		})
	}
}
