package domain

import "testing"

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
}
