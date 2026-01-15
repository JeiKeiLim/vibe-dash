package domain

import (
	"testing"
	"time"
)

func TestNewAgentState(t *testing.T) {
	state := NewAgentState("Claude Code", AgentWaitingForUser, 2*time.Hour, ConfidenceCertain)

	if state.Tool != "Claude Code" {
		t.Errorf("Tool = %q, want %q", state.Tool, "Claude Code")
	}
	if state.Status != AgentWaitingForUser {
		t.Errorf("Status = %v, want %v", state.Status, AgentWaitingForUser)
	}
	if state.Duration != 2*time.Hour {
		t.Errorf("Duration = %v, want %v", state.Duration, 2*time.Hour)
	}
	if state.Confidence != ConfidenceCertain {
		t.Errorf("Confidence = %v, want %v", state.Confidence, ConfidenceCertain)
	}
}

func TestAgentState_Helpers(t *testing.T) {
	tests := []struct {
		name       string
		status     AgentStatus
		isWaiting  bool
		isWorking  bool
		isInactive bool
		isUnknown  bool
	}{
		{"working", AgentWorking, false, true, false, false},
		{"waiting", AgentWaitingForUser, true, false, false, false},
		{"inactive", AgentInactive, false, false, true, false},
		{"unknown", AgentUnknown, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewAgentState("Test", tt.status, time.Second, ConfidenceUncertain)

			if got := state.IsWaiting(); got != tt.isWaiting {
				t.Errorf("IsWaiting() = %v, want %v", got, tt.isWaiting)
			}
			if got := state.IsWorking(); got != tt.isWorking {
				t.Errorf("IsWorking() = %v, want %v", got, tt.isWorking)
			}
			if got := state.IsInactive(); got != tt.isInactive {
				t.Errorf("IsInactive() = %v, want %v", got, tt.isInactive)
			}
			if got := state.IsUnknown(); got != tt.isUnknown {
				t.Errorf("IsUnknown() = %v, want %v", got, tt.isUnknown)
			}
		})
	}
}

func TestAgentState_Summary(t *testing.T) {
	tests := []struct {
		name       string
		tool       string
		status     AgentStatus
		confidence Confidence
		want       string
	}{
		{
			"full state",
			"Claude Code",
			AgentWaitingForUser,
			ConfidenceCertain,
			"Claude Code/Waiting (Certain)",
		},
		{
			"generic tool working",
			"Generic",
			AgentWorking,
			ConfidenceLikely,
			"Generic/Working (Likely)",
		},
		{
			"unknown state",
			"Unknown",
			AgentUnknown,
			ConfidenceUncertain,
			"Unknown/Unknown (Uncertain)",
		},
		{
			"inactive state",
			"Test Tool",
			AgentInactive,
			ConfidenceCertain,
			"Test Tool/Inactive (Certain)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewAgentState(tt.tool, tt.status, time.Hour, tt.confidence)
			if got := state.Summary(); got != tt.want {
				t.Errorf("Summary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAgentState_ZeroValue(t *testing.T) {
	var state AgentState

	if state.Tool != "" {
		t.Errorf("zero value Tool = %q, want empty string", state.Tool)
	}
	if state.Status != AgentUnknown {
		t.Errorf("zero value Status = %v, want AgentUnknown", state.Status)
	}
	if state.Duration != 0 {
		t.Errorf("zero value Duration = %v, want 0", state.Duration)
	}
	if state.Confidence != ConfidenceUncertain {
		t.Errorf("zero value Confidence = %v, want ConfidenceUncertain", state.Confidence)
	}

	// Zero-value state should report as unknown
	if !state.IsUnknown() {
		t.Error("zero value IsUnknown() = false, want true")
	}
}
