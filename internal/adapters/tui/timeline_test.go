package tui

import (
	"testing"
	"time"
)

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
		{"23 hours ago", now.Add(-23 * time.Hour), "✨"},
		{"25 hours ago", now.Add(-25 * time.Hour), "⚡"},
		{"6 days ago", now.Add(-6 * 24 * time.Hour), "⚡"},
		{"8 days ago", now.Add(-8 * 24 * time.Hour), ""},
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

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"zero time", time.Time{}, "never"},
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5m ago"},
		{"2 hours ago", now.Add(-2 * time.Hour), "2h ago"},
		{"3 days ago", now.Add(-3 * 24 * time.Hour), "3d ago"},
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
