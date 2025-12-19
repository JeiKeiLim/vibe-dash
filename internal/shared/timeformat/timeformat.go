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

// FormatWaitingDuration returns a duration string for the WAITING indicator.
// If detailed is false, returns compact format: "15m", "2h", "1d" (for project list)
// If detailed is true, returns precise format: "2h 15m", "1d 5h" (for detail panel)
// Used in TUI status column: "⏸️ WAITING 2h"
// Negative durations are clamped to "0m".
//
// Note on d=0 behavior: Returns "0m" in both modes rather than "0h 0m" for detailed.
// This is intentional - zero duration means "just started waiting" which is best
// expressed as "0m" (minimal time unit) rather than the verbose "0h 0m".
func FormatWaitingDuration(d time.Duration, detailed bool) string {
	// Clamp negative durations to zero (defensive)
	if d < 0 {
		d = 0
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours >= 24 {
		days := hours / 24
		if detailed {
			remainingHours := hours % 24
			return fmt.Sprintf("%dd %dh", days, remainingHours)
		}
		return fmt.Sprintf("%dd", days)
	}
	if hours >= 1 {
		if detailed {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", minutes)
}
