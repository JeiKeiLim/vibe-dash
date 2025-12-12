package domain

import (
	"errors"
	"testing"
)

func TestStage_String(t *testing.T) {
	tests := []struct {
		name  string
		stage Stage
		want  string
	}{
		{"unknown", StageUnknown, "Unknown"},
		{"specify", StageSpecify, "Specify"},
		{"plan", StagePlan, "Plan"},
		{"tasks", StageTasks, "Tasks"},
		{"implement", StageImplement, "Implement"},
		{"invalid negative", Stage(-1), "Unknown"},
		{"invalid large", Stage(100), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stage.String(); got != tt.want {
				t.Errorf("Stage.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseStage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Stage
		wantErr error
	}{
		{"valid lowercase specify", "specify", StageSpecify, nil},
		{"valid lowercase plan", "plan", StagePlan, nil},
		{"valid lowercase tasks", "tasks", StageTasks, nil},
		{"valid lowercase implement", "implement", StageImplement, nil},
		{"valid uppercase", "PLAN", StagePlan, nil},
		{"valid mixed case", "Plan", StagePlan, nil},
		{"valid MixeD CaSe", "TaSKs", StageTasks, nil},
		{"with leading spaces", "  plan", StagePlan, nil},
		{"with trailing spaces", "plan  ", StagePlan, nil},
		{"with both spaces", "  plan  ", StagePlan, nil},
		{"empty string", "", StageUnknown, nil},
		{"unknown string", "unknown", StageUnknown, nil},
		{"invalid", "invalid", StageUnknown, ErrInvalidStage},
		{"gibberish", "xyz123", StageUnknown, ErrInvalidStage},
		{"partial match", "pla", StageUnknown, ErrInvalidStage},
		{"with numbers", "plan2", StageUnknown, ErrInvalidStage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStage(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseStage() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseStage() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ParseStage() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("ParseStage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStage_Constants(t *testing.T) {
	// Verify iota values are as expected
	if StageUnknown != 0 {
		t.Errorf("StageUnknown = %d, want 0", StageUnknown)
	}
	if StageSpecify != 1 {
		t.Errorf("StageSpecify = %d, want 1", StageSpecify)
	}
	if StagePlan != 2 {
		t.Errorf("StagePlan = %d, want 2", StagePlan)
	}
	if StageTasks != 3 {
		t.Errorf("StageTasks = %d, want 3", StageTasks)
	}
	if StageImplement != 4 {
		t.Errorf("StageImplement = %d, want 4", StageImplement)
	}
}
