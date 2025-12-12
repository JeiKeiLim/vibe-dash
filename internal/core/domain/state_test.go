package domain

import (
	"errors"
	"testing"
)

func TestProjectState_String(t *testing.T) {
	tests := []struct {
		name  string
		state ProjectState
		want  string
	}{
		{"active", StateActive, "Active"},
		{"hibernated", StateHibernated, "Hibernated"},
		{"invalid negative", ProjectState(-1), "Unknown"},
		{"invalid large", ProjectState(100), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("ProjectState.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProjectState_Constants(t *testing.T) {
	// Verify iota values are as expected
	if StateActive != 0 {
		t.Errorf("StateActive = %d, want 0", StateActive)
	}
	if StateHibernated != 1 {
		t.Errorf("StateHibernated = %d, want 1", StateHibernated)
	}
}

func TestProjectState_ZeroValue(t *testing.T) {
	// Zero value should be StateActive (safe default for new projects)
	var state ProjectState
	if state != StateActive {
		t.Errorf("Zero value of ProjectState = %v, want StateActive", state)
	}
}

func TestParseProjectState(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ProjectState
		wantErr error
	}{
		{"valid lowercase active", "active", StateActive, nil},
		{"valid lowercase hibernated", "hibernated", StateHibernated, nil},
		{"valid uppercase", "ACTIVE", StateActive, nil},
		{"valid mixed case", "Hibernated", StateHibernated, nil},
		{"with leading spaces", "  active", StateActive, nil},
		{"with trailing spaces", "hibernated  ", StateHibernated, nil},
		{"with both spaces", "  active  ", StateActive, nil},
		{"empty string", "", StateActive, nil},
		{"invalid", "invalid", StateActive, ErrInvalidProjectState},
		{"gibberish", "xyz123", StateActive, ErrInvalidProjectState},
		{"partial match", "act", StateActive, ErrInvalidProjectState},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProjectState(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseProjectState() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseProjectState() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ParseProjectState() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("ParseProjectState() = %v, want %v", got, tt.want)
			}
		})
	}
}
