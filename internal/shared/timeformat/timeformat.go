// Package timeformat provides shared time formatting utilities.
package timeformat

import (
	"fmt"
	"time"
)

// FormatRelativeTime formats a time as relative (e.g., "5m ago", "2h ago").
// Returns "never" for zero time.
func FormatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	}
}

// RecencyIndicator returns an indicator emoji based on recency.
// Returns "✨" for activity within 24 hours (today).
// Returns "⚡" for activity within 7 days (this week).
// Returns empty string for older activity or zero time.
func RecencyIndicator(lastActivity time.Time) string {
	if lastActivity.IsZero() {
		return ""
	}

	since := time.Since(lastActivity)

	switch {
	case since < 24*time.Hour:
		return "✨" // Today
	case since < 7*24*time.Hour:
		return "⚡" // This week
	default:
		return "" // Older
	}
}

// FormatWaitingDuration returns a compact duration string for the WAITING indicator.
// Format: "15m" (minutes), "2h" (hours), "1d" (days)
// Used in TUI status column: "⏸️ WAITING 2h"
// Negative durations are clamped to "0m".
func FormatWaitingDuration(d time.Duration) string {
	// Clamp negative durations to zero (defensive)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
