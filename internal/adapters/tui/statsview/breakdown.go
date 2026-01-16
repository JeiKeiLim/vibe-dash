package statsview

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// StageDuration represents time spent in a single stage.
type StageDuration struct {
	Stage     string        // Stage name (e.g., "Plan", "Tasks", "Implement")
	Duration  time.Duration // Time spent in this stage
	IsCurrent bool          // True if this is the current (in-progress) stage
}

// FullTransition includes full stage transition data for breakdown calculation.
type FullTransition struct {
	FromStage      string
	ToStage        string
	TransitionedAt time.Time
}

// CalculateFromFullTransitions computes durations from full transition data.
// transitions: stage transition records (oldest first - ASC order)
// now: reference time for current stage calculation (enables testing)
// Returns: slice of StageDuration (one per unique stage visited)
func CalculateFromFullTransitions(transitions []FullTransition, now time.Time) []StageDuration {
	if len(transitions) == 0 {
		return nil
	}

	// Map to accumulate time per stage
	stageTimes := make(map[string]time.Duration)
	var lastStage string
	var lastTime time.Time

	for i, t := range transitions {
		if i > 0 {
			// Add duration to previous stage
			duration := t.TransitionedAt.Sub(lastTime)
			stageTimes[lastStage] += duration
		}
		lastStage = t.ToStage
		lastTime = t.TransitionedAt
	}

	// Add current stage duration (time since last transition)
	if lastStage != "" {
		stageTimes[lastStage] += now.Sub(lastTime)
	}

	// Convert to slice, mark last stage as current
	var result []StageDuration
	for stage, duration := range stageTimes {
		result = append(result, StageDuration{
			Stage:     stage,
			Duration:  duration,
			IsCurrent: stage == lastStage,
		})
	}

	// Sort by total duration (descending) - shows largest time sinks first
	sort.Slice(result, func(i, j int) bool {
		return result[i].Duration > result[j].Duration
	})

	return result
}

// RenderBreakdown renders time-per-stage as text with horizontal bars.
// maxWidth: available terminal width for rendering
// Returns formatted string with stage names, durations, and bars.
func RenderBreakdown(durations []StageDuration, maxWidth int) string {
	if len(durations) == 0 {
		return "No stage data available"
	}

	// Calculate total duration for percentages
	var totalDuration time.Duration
	for _, d := range durations {
		totalDuration += d.Duration
	}

	// Find the longest stage name for column alignment
	maxStageWidth := 0
	for _, d := range durations {
		// Account for "→ " prefix on current stage
		stageLen := len(d.Stage)
		if d.IsCurrent {
			stageLen += 2 // "→ " prefix
		}
		if stageLen > maxStageWidth {
			maxStageWidth = stageLen
		}
	}

	// Find the longest duration string for alignment
	maxDurWidth := 0
	for _, d := range durations {
		durStr := formatDuration(d.Duration)
		if len(durStr) > maxDurWidth {
			maxDurWidth = len(durStr)
		}
	}

	// Bar rendering constants
	const minBarWidth = 10
	const percentWidth = 6 // "(XX%)" + space
	const padding = 4      // Spaces between columns

	// Calculate available bar width
	barWidth := maxWidth - maxStageWidth - maxDurWidth - percentWidth - padding
	if barWidth < minBarWidth {
		barWidth = minBarWidth
	}
	if barWidth > 20 {
		barWidth = 20 // Cap bar width for consistency
	}

	// Find max duration for bar scaling
	var maxDuration time.Duration
	for _, d := range durations {
		if d.Duration > maxDuration {
			maxDuration = d.Duration
		}
	}

	var lines []string
	for _, d := range durations {
		// Stage name with current indicator
		stageName := d.Stage
		if d.IsCurrent {
			stageName = "→ " + stageName
		}

		// Duration string
		durStr := formatDuration(d.Duration)

		// Calculate bar fill
		var filled int
		if maxDuration > 0 && totalDuration > 0 {
			filled = int(float64(barWidth) * float64(d.Duration) / float64(maxDuration))
		}
		if filled < 0 {
			filled = 0
		}
		if filled > barWidth {
			filled = barWidth
		}
		empty := barWidth - filled

		bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

		// Calculate percentage
		var percent int
		if totalDuration > 0 {
			percent = int(float64(d.Duration) / float64(totalDuration) * 100)
		}

		// Build line
		line := fmt.Sprintf("%-*s  %-*s  %s  (%2d%%)",
			maxStageWidth, stageName,
			maxDurWidth, durStr,
			bar,
			percent,
		)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatDuration formats time.Duration for display.
// Examples: "3h", "45m", "2h 30m", "< 1m", "2d 5h"
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "< 1m"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dm", minutes)
}

// CalculateTotalDuration returns the sum of all stage durations.
func CalculateTotalDuration(durations []StageDuration) time.Duration {
	var total time.Duration
	for _, d := range durations {
		total += d.Duration
	}
	return total
}
