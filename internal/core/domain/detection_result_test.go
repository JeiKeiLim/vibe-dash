package domain

import (
	"testing"
	"time"
)

func TestNewDetectionResult(t *testing.T) {
	dr := NewDetectionResult("speckit", StagePlan, ConfidenceCertain, "plan.md exists")

	if dr.Method != "speckit" {
		t.Errorf("Method = %q, want %q", dr.Method, "speckit")
	}
	if dr.Stage != StagePlan {
		t.Errorf("Stage = %v, want %v", dr.Stage, StagePlan)
	}
	if dr.Confidence != ConfidenceCertain {
		t.Errorf("Confidence = %v, want %v", dr.Confidence, ConfidenceCertain)
	}
	if dr.Reasoning != "plan.md exists" {
		t.Errorf("Reasoning = %q, want %q", dr.Reasoning, "plan.md exists")
	}
}

func TestDetectionResult_IsUncertain(t *testing.T) {
	tests := []struct {
		name       string
		confidence Confidence
		want       bool
	}{
		{"uncertain", ConfidenceUncertain, true},
		{"likely", ConfidenceLikely, false},
		{"certain", ConfidenceCertain, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := DetectionResult{Confidence: tt.confidence}
			if got := dr.IsUncertain(); got != tt.want {
				t.Errorf("DetectionResult.IsUncertain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectionResult_IsCertain(t *testing.T) {
	tests := []struct {
		name       string
		confidence Confidence
		want       bool
	}{
		{"uncertain", ConfidenceUncertain, false},
		{"likely", ConfidenceLikely, false},
		{"certain", ConfidenceCertain, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := DetectionResult{Confidence: tt.confidence}
			if got := dr.IsCertain(); got != tt.want {
				t.Errorf("DetectionResult.IsCertain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectionResult_IsLikely(t *testing.T) {
	tests := []struct {
		name       string
		confidence Confidence
		want       bool
	}{
		{"uncertain", ConfidenceUncertain, false},
		{"likely", ConfidenceLikely, true},
		{"certain", ConfidenceCertain, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := DetectionResult{Confidence: tt.confidence}
			if got := dr.IsLikely(); got != tt.want {
				t.Errorf("DetectionResult.IsLikely() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectionResult_Summary(t *testing.T) {
	tests := []struct {
		name string
		dr   DetectionResult
		want string
	}{
		{
			name: "speckit plan certain",
			dr:   NewDetectionResult("speckit", StagePlan, ConfidenceCertain, "plan.md exists"),
			want: "speckit/Plan (Certain)",
		},
		{
			name: "bmad implement likely",
			dr:   NewDetectionResult("bmad", StageImplement, ConfidenceLikely, "tasks in progress"),
			want: "bmad/Implement (Likely)",
		},
		{
			name: "unknown unknown uncertain",
			dr:   NewDetectionResult("unknown", StageUnknown, ConfidenceUncertain, "no markers found"),
			want: "unknown/Unknown (Uncertain)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dr.Summary(); got != tt.want {
				t.Errorf("DetectionResult.Summary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectionResult_ZeroValue(t *testing.T) {
	var dr DetectionResult

	// Zero value should have sensible defaults
	if dr.Method != "" {
		t.Errorf("Zero value Method = %q, want empty string", dr.Method)
	}
	if dr.Stage != StageUnknown {
		t.Errorf("Zero value Stage = %v, want StageUnknown", dr.Stage)
	}
	if dr.Confidence != ConfidenceUncertain {
		t.Errorf("Zero value Confidence = %v, want ConfidenceUncertain", dr.Confidence)
	}
	if dr.Reasoning != "" {
		t.Errorf("Zero value Reasoning = %q, want empty string", dr.Reasoning)
	}
	// AC4: Zero value should have zero ArtifactTimestamp
	if !dr.ArtifactTimestamp.IsZero() {
		t.Errorf("Zero value ArtifactTimestamp = %v, want zero time", dr.ArtifactTimestamp)
	}
	// Story 14.4: Zero value should have no coexistence warning
	if dr.CoexistenceWarning {
		t.Errorf("Zero value CoexistenceWarning = %v, want false", dr.CoexistenceWarning)
	}
	if dr.CoexistenceMessage != "" {
		t.Errorf("Zero value CoexistenceMessage = %q, want empty string", dr.CoexistenceMessage)
	}
}

func TestDetectionResult_WithTimestamp(t *testing.T) {
	baseTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name          string
		timestamp     time.Time
		wantTimestamp time.Time
	}{
		{
			name:          "set non-zero timestamp",
			timestamp:     baseTime,
			wantTimestamp: baseTime,
		},
		{
			name:          "set zero timestamp",
			timestamp:     time.Time{},
			wantTimestamp: time.Time{},
		},
		{
			name:          "set different timestamp",
			timestamp:     baseTime.Add(2 * time.Hour),
			wantTimestamp: baseTime.Add(2 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDetectionResult("speckit", StagePlan, ConfidenceCertain, "test")
			result := dr.WithTimestamp(tt.timestamp)

			if !result.ArtifactTimestamp.Equal(tt.wantTimestamp) {
				t.Errorf("WithTimestamp() ArtifactTimestamp = %v, want %v", result.ArtifactTimestamp, tt.wantTimestamp)
			}
			// Verify other fields preserved
			if result.Method != dr.Method {
				t.Errorf("WithTimestamp() Method = %q, want %q", result.Method, dr.Method)
			}
			if result.Stage != dr.Stage {
				t.Errorf("WithTimestamp() Stage = %v, want %v", result.Stage, dr.Stage)
			}
			if result.Confidence != dr.Confidence {
				t.Errorf("WithTimestamp() Confidence = %v, want %v", result.Confidence, dr.Confidence)
			}
		})
	}
}

func TestDetectionResult_WithTimestamp_Immutability(t *testing.T) {
	// Test that original result is unchanged after WithTimestamp (AC5 - fluent method returns copy)
	original := NewDetectionResult("bmad", StageImplement, ConfidenceLikely, "original")
	timestamp := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	modified := original.WithTimestamp(timestamp)

	// Original should still have zero timestamp
	if !original.ArtifactTimestamp.IsZero() {
		t.Errorf("Original ArtifactTimestamp changed to %v, want zero time", original.ArtifactTimestamp)
	}

	// Modified should have the new timestamp
	if !modified.ArtifactTimestamp.Equal(timestamp) {
		t.Errorf("Modified ArtifactTimestamp = %v, want %v", modified.ArtifactTimestamp, timestamp)
	}
}

func TestDetectionResult_HasTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp time.Time
		want      bool
	}{
		{
			name:      "zero time returns false",
			timestamp: time.Time{},
			want:      false,
		},
		{
			name:      "non-zero time returns true",
			timestamp: time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
			want:      true,
		},
		{
			name:      "unix epoch returns true",
			timestamp: time.Unix(0, 0),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDetectionResult("speckit", StagePlan, ConfidenceCertain, "test").WithTimestamp(tt.timestamp)
			if got := dr.HasTimestamp(); got != tt.want {
				t.Errorf("HasTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectionResult_HasTimestamp_ZeroValue(t *testing.T) {
	// Test that zero value DetectionResult returns false for HasTimestamp
	var dr DetectionResult
	if dr.HasTimestamp() {
		t.Error("Zero value DetectionResult.HasTimestamp() = true, want false")
	}
}

func TestDetectionResult_WithTimestamp_Chaining(t *testing.T) {
	// Test that chaining WithTimestamp calls works correctly - last call wins
	t1 := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 6, 16, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2024, 6, 17, 14, 0, 0, 0, time.UTC)

	dr := NewDetectionResult("speckit", StagePlan, ConfidenceCertain, "test")

	// Chain multiple WithTimestamp calls
	result := dr.WithTimestamp(t1).WithTimestamp(t2).WithTimestamp(t3)

	// Final result should have t3
	if !result.ArtifactTimestamp.Equal(t3) {
		t.Errorf("Chained WithTimestamp() ArtifactTimestamp = %v, want %v (last timestamp)", result.ArtifactTimestamp, t3)
	}

	// Original should be unchanged
	if !dr.ArtifactTimestamp.IsZero() {
		t.Errorf("Original ArtifactTimestamp = %v, want zero time", dr.ArtifactTimestamp)
	}
}

func TestDetectionResult_CoexistenceWarning(t *testing.T) {
	tests := []struct {
		name         string
		applyWarning bool
		wantWarning  bool
		wantMessage  string
	}{
		{
			name:         "default has no warning",
			applyWarning: false,
			wantWarning:  false,
			wantMessage:  "",
		},
		{
			name:         "with warning sets flag and message",
			applyWarning: true,
			wantWarning:  true,
			wantMessage:  "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDetectionResult("speckit", StagePlan, ConfidenceCertain, "test")
			if tt.applyWarning {
				r = r.WithCoexistenceWarning("test message")
			}

			if r.HasCoexistenceWarning() != tt.wantWarning {
				t.Errorf("HasCoexistenceWarning() = %v, want %v", r.HasCoexistenceWarning(), tt.wantWarning)
			}
			if r.CoexistenceMessage != tt.wantMessage {
				t.Errorf("CoexistenceMessage = %q, want %q", r.CoexistenceMessage, tt.wantMessage)
			}
		})
	}
}

func TestDetectionResult_CoexistenceWarning_PreservesOtherFields(t *testing.T) {
	timestamp := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	original := NewDetectionResult("bmad", StageTasks, ConfidenceLikely, "found .bmad").WithTimestamp(timestamp)
	modified := original.WithCoexistenceWarning("test")

	// Verify original unchanged
	if original.CoexistenceWarning {
		t.Error("original should not be modified")
	}

	// Verify ALL fields preserved
	if modified.Method != "bmad" {
		t.Errorf("Method = %q, want %q", modified.Method, "bmad")
	}
	if modified.Stage != StageTasks {
		t.Errorf("Stage = %v, want %v", modified.Stage, StageTasks)
	}
	if modified.Confidence != ConfidenceLikely {
		t.Errorf("Confidence = %v, want %v", modified.Confidence, ConfidenceLikely)
	}
	if modified.Reasoning != "found .bmad" {
		t.Errorf("Reasoning = %q, want %q", modified.Reasoning, "found .bmad")
	}
	if !modified.ArtifactTimestamp.Equal(timestamp) {
		t.Errorf("ArtifactTimestamp = %v, want %v", modified.ArtifactTimestamp, timestamp)
	}
}
