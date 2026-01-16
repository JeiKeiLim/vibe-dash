package domain

import (
	"testing"
	"time"
)

// createResultWithTimestamp is a test helper to create DetectionResult with timestamp
func createResultWithTimestamp(method string, timestamp time.Time) *DetectionResult {
	result := NewDetectionResult(method, StagePlan, ConfidenceCertain, "test").WithTimestamp(timestamp)
	return &result
}

func TestSelectByTimestamp(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		results        []*DetectionResult
		wantWinnerNil  bool
		wantWinnerName string // Method name of expected winner
		wantHasClear   bool
	}{
		{
			name:          "empty slice",
			results:       []*DetectionResult{},
			wantWinnerNil: true,
			wantHasClear:  false,
		},
		{
			name:           "single result",
			results:        []*DetectionResult{createResultWithTimestamp("speckit", now)},
			wantWinnerNil:  false,
			wantWinnerName: "speckit",
			wantHasClear:   true,
		},
		{
			name: "clear winner - BMAD more recent (1 week vs 1 hour ago)",
			results: []*DetectionResult{
				createResultWithTimestamp("speckit", now.Add(-7*24*time.Hour)), // 1 week ago
				createResultWithTimestamp("bmad", now.Add(-1*time.Hour)),       // 1 hour ago
			},
			wantWinnerNil:  false,
			wantWinnerName: "bmad",
			wantHasClear:   true,
		},
		{
			name: "clear winner - Speckit more recent (5 min vs 2 days ago)",
			results: []*DetectionResult{
				createResultWithTimestamp("speckit", now.Add(-5*time.Minute)), // 5 min ago
				createResultWithTimestamp("bmad", now.Add(-2*24*time.Hour)),   // 2 days ago
			},
			wantWinnerNil:  false,
			wantWinnerName: "speckit",
			wantHasClear:   true,
		},
		{
			name: "tie - within threshold (30 min difference)",
			results: []*DetectionResult{
				createResultWithTimestamp("speckit", now.Add(-30*time.Minute)),
				createResultWithTimestamp("bmad", now.Add(-45*time.Minute)),
			},
			wantWinnerNil: true,
			wantHasClear:  false,
		},
		{
			name: "exact 1 hour boundary - should be tie (threshold inclusive)",
			results: []*DetectionResult{
				createResultWithTimestamp("speckit", now),
				createResultWithTimestamp("bmad", now.Add(-1*time.Hour)), // exactly 1 hour
			},
			wantWinnerNil: true,
			wantHasClear:  false,
		},
		{
			name: "just over threshold - clear winner (1h1m difference)",
			results: []*DetectionResult{
				createResultWithTimestamp("speckit", now),
				createResultWithTimestamp("bmad", now.Add(-61*time.Minute)),
			},
			wantWinnerNil:  false,
			wantWinnerName: "speckit",
			wantHasClear:   true,
		},
		{
			name: "zero timestamp vs valid - valid wins",
			results: []*DetectionResult{
				createResultWithTimestamp("speckit", time.Time{}), // zero
				createResultWithTimestamp("bmad", now),            // valid
			},
			wantWinnerNil:  false,
			wantWinnerName: "bmad",
			wantHasClear:   true,
		},
		{
			name: "both zero timestamps - first wins",
			results: []*DetectionResult{
				createResultWithTimestamp("speckit", time.Time{}),
				createResultWithTimestamp("bmad", time.Time{}),
			},
			wantWinnerNil:  false,
			wantWinnerName: "speckit", // first in slice
			wantHasClear:   true,
		},
		{
			name: "three results - most recent wins",
			results: []*DetectionResult{
				createResultWithTimestamp("speckit", now.Add(-7*24*time.Hour)), // 1 week ago
				createResultWithTimestamp("bmad", now.Add(-2*24*time.Hour)),    // 2 days ago
				createResultWithTimestamp("custom", now.Add(-1*time.Hour)),     // 1 hour ago
			},
			wantWinnerNil:  false,
			wantWinnerName: "custom",
			wantHasClear:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winner, hasClear := SelectByTimestamp(tt.results)

			if tt.wantWinnerNil && winner != nil {
				t.Errorf("SelectByTimestamp() winner = %v, want nil", winner)
			}
			if !tt.wantWinnerNil && winner == nil {
				t.Error("SelectByTimestamp() winner = nil, want non-nil")
			}
			if !tt.wantWinnerNil && winner != nil && winner.Method != tt.wantWinnerName {
				t.Errorf("SelectByTimestamp() winner.Method = %q, want %q", winner.Method, tt.wantWinnerName)
			}
			if hasClear != tt.wantHasClear {
				t.Errorf("SelectByTimestamp() hasClear = %v, want %v", hasClear, tt.wantHasClear)
			}
		})
	}
}

func TestCoexistenceThreshold_Value(t *testing.T) {
	// Verify the threshold constant is 1 hour as specified in AC3
	if CoexistenceThreshold != time.Hour {
		t.Errorf("CoexistenceThreshold = %v, want %v", CoexistenceThreshold, time.Hour)
	}
}
