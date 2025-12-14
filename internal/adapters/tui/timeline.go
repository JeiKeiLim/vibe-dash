package tui

import (
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)

// RecencyIndicator returns an indicator emoji based on recency.
// Returns "✨" for activity within 24 hours (today).
// Returns "⚡" for activity within 7 days (this week).
// Returns empty string for older activity or zero time.
// This is a convenience wrapper around timeformat.RecencyIndicator.
func RecencyIndicator(lastActivity time.Time) string {
	return timeformat.RecencyIndicator(lastActivity)
}

// FormatRelativeTime formats a time as relative (e.g., "5m ago", "2h ago").
// Returns "never" for zero time.
// This is a convenience wrapper around timeformat.FormatRelativeTime.
func FormatRelativeTime(t time.Time) string {
	return timeformat.FormatRelativeTime(t)
}
