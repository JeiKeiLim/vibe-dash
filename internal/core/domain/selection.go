package domain

import "time"

// CoexistenceThreshold defines minimum timestamp difference for clear winner.
// If timestamps are within this threshold, it's considered a tie.
// Threshold is INCLUSIVE: exactly 1 hour difference = tie.
const CoexistenceThreshold = 1 * time.Hour

// SelectByTimestamp chooses the methodology with most recent artifact timestamp.
//
// Return semantics:
//   - (winner, true)  → clear winner exists (difference > CoexistenceThreshold)
//   - (nil, false)    → tie (difference <= threshold), caller handles coexistence
//   - (result, true)  → single result provided
//   - (nil, false)    → empty slice
//   - (first, true)   → all zero timestamps, deterministic fallback to first
func SelectByTimestamp(results []*DetectionResult) (*DetectionResult, bool) {
	if len(results) == 0 {
		return nil, false
	}
	if len(results) == 1 {
		return results[0], true
	}

	// Find most recent and second most recent
	var mostRecent *DetectionResult
	var mostRecentTime, secondRecentTime time.Time

	for _, r := range results {
		ts := r.ArtifactTimestamp
		if ts.After(mostRecentTime) {
			secondRecentTime = mostRecentTime
			mostRecent = r
			mostRecentTime = ts
		} else if ts.After(secondRecentTime) {
			secondRecentTime = ts
		}
	}

	// Both zero timestamps: first in slice wins (deterministic fallback)
	if mostRecentTime.IsZero() && secondRecentTime.IsZero() {
		return results[0], true
	}

	// Check if clear winner (difference > threshold, NOT >=)
	diff := mostRecentTime.Sub(secondRecentTime)
	if diff > CoexistenceThreshold {
		return mostRecent, true
	}

	// Tie case - no clear winner (difference <= 1 hour)
	return nil, false
}
