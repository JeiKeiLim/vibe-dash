package statsview

import (
	"strings"
	"testing"
	"time"
)

func TestCalculateFromFullTransitions_NoTransitions(t *testing.T) {
	now := time.Now()
	result := CalculateFromFullTransitions(nil, now)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}

	result = CalculateFromFullTransitions([]FullTransition{}, now)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestCalculateFromFullTransitions_SingleTransition(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	transitions := []FullTransition{
		{FromStage: "", ToStage: "Plan", TransitionedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)},
	}

	result := CalculateFromFullTransitions(transitions, now)

	if len(result) != 1 {
		t.Fatalf("Expected 1 duration, got %d", len(result))
	}
	if result[0].Stage != "Plan" {
		t.Errorf("Expected stage 'Plan', got '%s'", result[0].Stage)
	}
	if result[0].Duration != 2*time.Hour {
		t.Errorf("Expected 2 hours, got %v", result[0].Duration)
	}
	if !result[0].IsCurrent {
		t.Errorf("Expected stage to be current")
	}
}

func TestCalculateFromFullTransitions_MultipleTransitions(t *testing.T) {
	now := time.Date(2025, 1, 15, 19, 0, 0, 0, time.UTC)
	transitions := []FullTransition{
		{FromStage: "", ToStage: "Plan", TransitionedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)},
		{FromStage: "Plan", ToStage: "Tasks", TransitionedAt: time.Date(2025, 1, 15, 13, 0, 0, 0, time.UTC)},
		{FromStage: "Tasks", ToStage: "Implement", TransitionedAt: time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)},
	}

	result := CalculateFromFullTransitions(transitions, now)

	if len(result) != 3 {
		t.Fatalf("Expected 3 durations, got %d", len(result))
	}

	// Results are sorted by duration (descending)
	// Implement: 5h (current), Plan: 3h, Tasks: 1h
	stageMap := make(map[string]StageDuration)
	for _, d := range result {
		stageMap[d.Stage] = d
	}

	if stageMap["Plan"].Duration != 3*time.Hour {
		t.Errorf("Plan: expected 3h, got %v", stageMap["Plan"].Duration)
	}
	if stageMap["Tasks"].Duration != 1*time.Hour {
		t.Errorf("Tasks: expected 1h, got %v", stageMap["Tasks"].Duration)
	}
	if stageMap["Implement"].Duration != 5*time.Hour {
		t.Errorf("Implement: expected 5h, got %v", stageMap["Implement"].Duration)
	}
	if !stageMap["Implement"].IsCurrent {
		t.Errorf("Implement should be marked as current")
	}
	if stageMap["Plan"].IsCurrent || stageMap["Tasks"].IsCurrent {
		t.Errorf("Plan and Tasks should not be marked as current")
	}
}

func TestCalculateFromFullTransitions_CurrentStageCalculation(t *testing.T) {
	// Test that current stage duration is calculated from last transition to now
	baseTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	now := time.Date(2025, 1, 15, 15, 30, 0, 0, time.UTC)

	transitions := []FullTransition{
		{FromStage: "", ToStage: "Plan", TransitionedAt: baseTime},
	}

	result := CalculateFromFullTransitions(transitions, now)

	if len(result) != 1 {
		t.Fatalf("Expected 1 duration, got %d", len(result))
	}

	// 15:30 - 10:00 = 5.5 hours
	expected := 5*time.Hour + 30*time.Minute
	if result[0].Duration != expected {
		t.Errorf("Expected %v, got %v", expected, result[0].Duration)
	}
}

func TestCalculateFromFullTransitions_RevisitedStages(t *testing.T) {
	// Test that revisiting a stage accumulates time
	now := time.Date(2025, 1, 15, 20, 0, 0, 0, time.UTC)
	transitions := []FullTransition{
		{FromStage: "", ToStage: "Plan", TransitionedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)},
		{FromStage: "Plan", ToStage: "Tasks", TransitionedAt: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)},
		{FromStage: "Tasks", ToStage: "Plan", TransitionedAt: time.Date(2025, 1, 15, 13, 0, 0, 0, time.UTC)}, // Back to Plan
		{FromStage: "Plan", ToStage: "Implement", TransitionedAt: time.Date(2025, 1, 15, 15, 0, 0, 0, time.UTC)},
	}

	result := CalculateFromFullTransitions(transitions, now)

	stageMap := make(map[string]StageDuration)
	for _, d := range result {
		stageMap[d.Stage] = d
	}

	// Plan: 2h (first visit) + 2h (second visit) = 4h
	// Tasks: 1h
	// Implement: 5h (current)
	if stageMap["Plan"].Duration != 4*time.Hour {
		t.Errorf("Plan: expected 4h (accumulated), got %v", stageMap["Plan"].Duration)
	}
	if stageMap["Tasks"].Duration != 1*time.Hour {
		t.Errorf("Tasks: expected 1h, got %v", stageMap["Tasks"].Duration)
	}
	if stageMap["Implement"].Duration != 5*time.Hour {
		t.Errorf("Implement: expected 5h, got %v", stageMap["Implement"].Duration)
	}
}

func TestRenderBreakdown_EmptyDurations(t *testing.T) {
	result := RenderBreakdown(nil, 80)
	expected := "No stage data available"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	result = RenderBreakdown([]StageDuration{}, 80)
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRenderBreakdown_WithBars(t *testing.T) {
	durations := []StageDuration{
		{Stage: "Implement", Duration: 5 * time.Hour, IsCurrent: true},
		{Stage: "Plan", Duration: 3 * time.Hour, IsCurrent: false},
		{Stage: "Tasks", Duration: 1 * time.Hour, IsCurrent: false},
	}

	result := RenderBreakdown(durations, 80)

	// Check that output contains stage names
	if !strings.Contains(result, "Implement") {
		t.Error("Expected 'Implement' in output")
	}
	if !strings.Contains(result, "Plan") {
		t.Error("Expected 'Plan' in output")
	}
	if !strings.Contains(result, "Tasks") {
		t.Error("Expected 'Tasks' in output")
	}

	// Check current stage indicator
	if !strings.Contains(result, "→") {
		t.Error("Expected '→' current stage indicator in output")
	}

	// Check bar characters
	if !strings.Contains(result, "█") {
		t.Error("Expected filled bar character in output")
	}

	// Check percentages
	if !strings.Contains(result, "%") {
		t.Error("Expected percentage values in output")
	}
}

func TestFormatDuration_Various(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "< 1m"},
		{5 * time.Minute, "5m"},
		{45 * time.Minute, "45m"},
		{1 * time.Hour, "1h"},
		{2*time.Hour + 30*time.Minute, "2h 30m"},
		{24 * time.Hour, "1d"},
		{48*time.Hour + 5*time.Hour, "2d 5h"},
		{72 * time.Hour, "3d"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = '%s', want '%s'", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestCalculateTotalDuration(t *testing.T) {
	durations := []StageDuration{
		{Stage: "Plan", Duration: 3 * time.Hour},
		{Stage: "Tasks", Duration: 1 * time.Hour},
		{Stage: "Implement", Duration: 5 * time.Hour},
	}

	total := CalculateTotalDuration(durations)
	expected := 9 * time.Hour
	if total != expected {
		t.Errorf("Expected %v, got %v", expected, total)
	}
}

func TestCalculateTotalDuration_Empty(t *testing.T) {
	total := CalculateTotalDuration(nil)
	if total != 0 {
		t.Errorf("Expected 0, got %v", total)
	}

	total = CalculateTotalDuration([]StageDuration{})
	if total != 0 {
		t.Errorf("Expected 0, got %v", total)
	}
}

func TestRenderBreakdown_NarrowWidth(t *testing.T) {
	durations := []StageDuration{
		{Stage: "Implement", Duration: 5 * time.Hour, IsCurrent: true},
	}

	// Even with narrow width, should still produce output
	result := RenderBreakdown(durations, 40)
	if result == "" || result == "No stage data available" {
		t.Error("Expected valid output for narrow width")
	}
}
