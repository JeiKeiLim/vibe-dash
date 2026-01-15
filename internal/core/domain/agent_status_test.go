package domain

import (
	"errors"
	"testing"
)

func TestAgentStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status AgentStatus
		want   string
	}{
		{"unknown", AgentUnknown, "Unknown"},
		{"working", AgentWorking, "Working"},
		{"waiting", AgentWaitingForUser, "Waiting"},
		{"inactive", AgentInactive, "Inactive"},
		{"invalid negative", AgentStatus(-1), "Unknown"},
		{"invalid boundary", AgentStatus(4), "Unknown"},
		{"invalid large", AgentStatus(100), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("AgentStatus.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseAgentStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    AgentStatus
		wantErr error
	}{
		{"valid lowercase working", "working", AgentWorking, nil},
		{"valid waiting", "waiting", AgentWaitingForUser, nil},
		{"valid waitingforuser", "waitingforuser", AgentWaitingForUser, nil},
		{"valid inactive", "inactive", AgentInactive, nil},
		{"valid unknown", "unknown", AgentUnknown, nil},
		{"valid uppercase", "WORKING", AgentWorking, nil},
		{"valid mixed case", "Working", AgentWorking, nil},
		{"with leading spaces", "  working", AgentWorking, nil},
		{"with trailing spaces", "working  ", AgentWorking, nil},
		{"with both spaces", "  working  ", AgentWorking, nil},
		{"empty string", "", AgentUnknown, nil},
		{"invalid", "invalid", AgentUnknown, ErrInvalidAgentStatus},
		{"gibberish", "xyz123", AgentUnknown, ErrInvalidAgentStatus},
		{"partial match", "work", AgentUnknown, ErrInvalidAgentStatus},
		{"typo workin", "workin", AgentUnknown, ErrInvalidAgentStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAgentStatus(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseAgentStatus() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseAgentStatus() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ParseAgentStatus() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("ParseAgentStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentStatus_ZeroValue(t *testing.T) {
	var status AgentStatus // zero value
	if status != AgentUnknown {
		t.Errorf("zero value = %v, want AgentUnknown", status)
	}
}

func TestAgentStatus_Constants(t *testing.T) {
	// Verify iota values are as expected
	if AgentUnknown != 0 {
		t.Errorf("AgentUnknown = %d, want 0", AgentUnknown)
	}
	if AgentWorking != 1 {
		t.Errorf("AgentWorking = %d, want 1", AgentWorking)
	}
	if AgentWaitingForUser != 2 {
		t.Errorf("AgentWaitingForUser = %d, want 2", AgentWaitingForUser)
	}
	if AgentInactive != 3 {
		t.Errorf("AgentInactive = %d, want 3", AgentInactive)
	}
}
