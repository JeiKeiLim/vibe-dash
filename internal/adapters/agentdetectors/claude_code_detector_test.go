package agentdetectors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// TestNewClaudeCodeDetector_DefaultsCreated verifies constructor creates dependencies.
func TestNewClaudeCodeDetector_DefaultsCreated(t *testing.T) {
	d := NewClaudeCodeDetector()

	if d == nil {
		t.Fatal("NewClaudeCodeDetector() returned nil")
	}
	if d.pathMatcher == nil {
		t.Error("pathMatcher should be created by default")
	}
	if d.logParser == nil {
		t.Error("logParser should be created by default")
	}
}

// TestNewClaudeCodeDetector_WithOptions verifies functional options work.
func TestNewClaudeCodeDetector_WithOptions(t *testing.T) {
	customPathMatcher := NewClaudeCodePathMatcher()
	customLogParser := NewClaudeCodeLogParser()

	d := NewClaudeCodeDetector(
		WithPathMatcher(customPathMatcher),
		WithLogParser(customLogParser),
	)

	if d.pathMatcher != customPathMatcher {
		t.Error("WithPathMatcher option not applied")
	}
	if d.logParser != customLogParser {
		t.Error("WithLogParser option not applied")
	}
}

// TestClaudeCodeDetector_Name verifies Name returns "Claude Code".
func TestClaudeCodeDetector_Name(t *testing.T) {
	d := NewClaudeCodeDetector()

	if got := d.Name(); got != "Claude Code" {
		t.Errorf("Name() = %q, want %q", got, "Claude Code")
	}
}

// setupClaudeTestDir creates a temp directory structure for integration tests.
// Returns: tmpHomeDir, projectPath, claudeProjectDir
func setupClaudeTestDir(t *testing.T) (string, string, string) {
	t.Helper()

	// Create temp home dir
	tmpHomeDir := t.TempDir()

	// Project path
	projectPath := "/test/my-project"

	// Convert path to Claude format: /test/my-project → -test-my-project
	escapedPath := strings.ReplaceAll(projectPath, "/", "-")
	claudeProjectDir := filepath.Join(tmpHomeDir, ".claude", "projects", escapedPath)
	if err := os.MkdirAll(claudeProjectDir, 0755); err != nil {
		t.Fatalf("failed to create claude project dir: %v", err)
	}

	return tmpHomeDir, projectPath, claudeProjectDir
}

// TestDetect_EndTurn_WaitingForUser tests AC1: end_turn → WaitingForUser.
func TestDetect_EndTurn_WaitingForUser(t *testing.T) {
	tmpHomeDir, projectPath, claudeProjectDir := setupClaudeTestDir(t)

	// Write JSONL with end_turn entry
	entry := `{"type":"assistant","stop_reason":"end_turn","timestamp":"2026-01-16T12:00:00Z"}`
	if err := os.WriteFile(filepath.Join(claudeProjectDir, "session.jsonl"), []byte(entry), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache() // Clear cache since HOME changed

	state, err := d.Detect(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if state.Status != domain.AgentWaitingForUser {
		t.Errorf("Status = %v, want WaitingForUser", state.Status)
	}
	if state.Confidence != domain.ConfidenceCertain {
		t.Errorf("Confidence = %v, want Certain", state.Confidence)
	}
	if state.Tool != "Claude Code" {
		t.Errorf("Tool = %q, want %q", state.Tool, "Claude Code")
	}
}

// TestDetect_ToolUse_Working tests AC2: tool_use → Working.
func TestDetect_ToolUse_Working(t *testing.T) {
	tmpHomeDir, projectPath, claudeProjectDir := setupClaudeTestDir(t)

	// Write JSONL with tool_use entry
	entry := `{"type":"assistant","stop_reason":"tool_use","timestamp":"2026-01-16T12:00:00Z"}`
	if err := os.WriteFile(filepath.Join(claudeProjectDir, "session.jsonl"), []byte(entry), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache()

	state, err := d.Detect(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if state.Status != domain.AgentWorking {
		t.Errorf("Status = %v, want Working", state.Status)
	}
	if state.Confidence != domain.ConfidenceCertain {
		t.Errorf("Confidence = %v, want Certain", state.Confidence)
	}
}

// TestDetect_NoAssistantEntries_Inactive tests AC3: no assistant entries → Inactive.
func TestDetect_NoAssistantEntries_Inactive(t *testing.T) {
	tmpHomeDir, projectPath, claudeProjectDir := setupClaudeTestDir(t)

	// Write JSONL with only user entries
	entries := `{"type":"user","message":"hello"}
{"type":"user","message":"world"}`
	if err := os.WriteFile(filepath.Join(claudeProjectDir, "session.jsonl"), []byte(entries), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache()

	state, err := d.Detect(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if state.Status != domain.AgentInactive {
		t.Errorf("Status = %v, want Inactive", state.Status)
	}
	if state.Confidence != domain.ConfidenceCertain {
		t.Errorf("Confidence = %v, want Certain", state.Confidence)
	}
}

// TestDetect_Duration tests AC4: Duration reflects time since last activity.
func TestDetect_Duration(t *testing.T) {
	tmpHomeDir, projectPath, claudeProjectDir := setupClaudeTestDir(t)

	// Timestamp 2 hours ago
	twoHoursAgo := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	entry := `{"type":"assistant","stop_reason":"end_turn","timestamp":"` + twoHoursAgo + `"}`
	if err := os.WriteFile(filepath.Join(claudeProjectDir, "session.jsonl"), []byte(entry), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache()

	state, err := d.Detect(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// Duration should be approximately 2 hours (within 1 second tolerance)
	expected := 2 * time.Hour
	tolerance := 5 * time.Second
	if state.Duration < expected-tolerance || state.Duration > expected+tolerance {
		t.Errorf("Duration = %v, want approximately %v", state.Duration, expected)
	}
}

// TestDetect_ZeroTimestamp_DurationZero tests AC4: zero timestamp → Duration = 0.
func TestDetect_ZeroTimestamp_DurationZero(t *testing.T) {
	tmpHomeDir, projectPath, claudeProjectDir := setupClaudeTestDir(t)

	// Entry without timestamp
	entry := `{"type":"assistant","stop_reason":"end_turn"}`
	if err := os.WriteFile(filepath.Join(claudeProjectDir, "session.jsonl"), []byte(entry), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache()

	state, err := d.Detect(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if state.Duration != 0 {
		t.Errorf("Duration = %v, want 0 (timestamp parsing failed)", state.Duration)
	}
	// Should still detect state correctly from stop_reason
	if state.Status != domain.AgentWaitingForUser {
		t.Errorf("Status = %v, want WaitingForUser", state.Status)
	}
}

// TestDetect_ContextCancelled_ReturnsPromptly tests AC7: context cancellation.
func TestDetect_ContextCancelled_ReturnsPromptly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	d := NewClaudeCodeDetector()

	start := time.Now()
	state, err := d.Detect(ctx, "/some/path")
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Errorf("Detect took %v, want < 100ms", elapsed)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v, want AgentUnknown", state.Status)
	}
}

// TestDetect_NoClaudeLogs_Unknown tests AC8: no Claude logs → AgentUnknown.
func TestDetect_NoClaudeLogs_Unknown(t *testing.T) {
	// Use temp dir as HOME where no .claude exists
	tmpHomeDir := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache()

	state, err := d.Detect(context.Background(), "/nonexistent/project")
	if err != nil {
		t.Errorf("Detect() should not return error for missing logs, got: %v", err)
	}

	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v, want AgentUnknown", state.Status)
	}
	if state.Confidence != domain.ConfidenceUncertain {
		t.Errorf("Confidence = %v, want Uncertain", state.Confidence)
	}
}

// TestDetect_EmptySessionDirectory_Inactive tests empty session dir → Inactive.
func TestDetect_EmptySessionDirectory_Inactive(t *testing.T) {
	tmpHomeDir, projectPath, claudeProjectDir := setupClaudeTestDir(t)

	// Create a .jsonl file to pass PathMatcher, but it's empty
	if err := os.WriteFile(filepath.Join(claudeProjectDir, "session.jsonl"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache()

	state, err := d.Detect(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// Empty file → no assistant entries → Inactive
	if state.Status != domain.AgentInactive {
		t.Errorf("Status = %v, want Inactive", state.Status)
	}
}

// TestDetect_UnrecognizedStopReason_Unknown tests unknown stop_reason → Unknown.
func TestDetect_UnrecognizedStopReason_Unknown(t *testing.T) {
	tmpHomeDir, projectPath, claudeProjectDir := setupClaudeTestDir(t)

	// Entry with unrecognized stop_reason
	entry := `{"type":"assistant","stop_reason":"max_tokens","timestamp":"2026-01-16T12:00:00Z"}`
	if err := os.WriteFile(filepath.Join(claudeProjectDir, "session.jsonl"), []byte(entry), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache()

	state, err := d.Detect(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v, want Unknown", state.Status)
	}
	if state.Confidence != domain.ConfidenceUncertain {
		t.Errorf("Confidence = %v, want Uncertain", state.Confidence)
	}
}

// TestDetect_MultipleEntriesReturnsLast tests that last assistant entry is used.
func TestDetect_MultipleEntriesReturnsLast(t *testing.T) {
	tmpHomeDir, projectPath, claudeProjectDir := setupClaudeTestDir(t)

	// Multiple entries - last is tool_use
	entries := `{"type":"assistant","stop_reason":"end_turn","timestamp":"2026-01-16T11:00:00Z"}
{"type":"user","message":"continue"}
{"type":"assistant","stop_reason":"tool_use","timestamp":"2026-01-16T12:00:00Z"}`
	if err := os.WriteFile(filepath.Join(claudeProjectDir, "session.jsonl"), []byte(entries), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache()

	state, err := d.Detect(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// Should reflect last assistant entry (tool_use → Working)
	if state.Status != domain.AgentWorking {
		t.Errorf("Status = %v, want Working (from last entry)", state.Status)
	}
}

// TestDetect_FutureTimestamp_DurationZero tests that future timestamps (clock skew) result in Duration = 0.
func TestDetect_FutureTimestamp_DurationZero(t *testing.T) {
	tmpHomeDir, projectPath, claudeProjectDir := setupClaudeTestDir(t)

	// Timestamp 1 hour in the future (simulates clock skew)
	futureTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	entry := `{"type":"assistant","stop_reason":"end_turn","timestamp":"` + futureTime + `"}`
	if err := os.WriteFile(filepath.Join(claudeProjectDir, "session.jsonl"), []byte(entry), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHomeDir)
	defer os.Setenv("HOME", origHome)

	d := NewClaudeCodeDetector()
	d.pathMatcher.ClearCache()

	state, err := d.Detect(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// Duration should be 0 for future timestamps (negative duration clamped to 0)
	if state.Duration != 0 {
		t.Errorf("Duration = %v, want 0 (future timestamp should clamp to 0)", state.Duration)
	}
	// Should still detect state correctly from stop_reason
	if state.Status != domain.AgentWaitingForUser {
		t.Errorf("Status = %v, want WaitingForUser", state.Status)
	}
}

// TestDetermineState_TableDriven uses table-driven tests for determineState.
func TestDetermineState_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		entry          *ClaudeLogEntry
		wantStatus     domain.AgentStatus
		wantConfidence domain.Confidence
	}{
		{
			name:           "end_turn returns WaitingForUser",
			entry:          &ClaudeLogEntry{Type: "assistant", StopReason: "end_turn", Timestamp: time.Now()},
			wantStatus:     domain.AgentWaitingForUser,
			wantConfidence: domain.ConfidenceCertain,
		},
		{
			name:           "tool_use returns Working",
			entry:          &ClaudeLogEntry{Type: "assistant", StopReason: "tool_use", Timestamp: time.Now()},
			wantStatus:     domain.AgentWorking,
			wantConfidence: domain.ConfidenceCertain,
		},
		{
			name:           "max_tokens returns Unknown",
			entry:          &ClaudeLogEntry{Type: "assistant", StopReason: "max_tokens", Timestamp: time.Now()},
			wantStatus:     domain.AgentUnknown,
			wantConfidence: domain.ConfidenceUncertain,
		},
		{
			name:           "empty stop_reason returns Unknown",
			entry:          &ClaudeLogEntry{Type: "assistant", StopReason: "", Timestamp: time.Now()},
			wantStatus:     domain.AgentUnknown,
			wantConfidence: domain.ConfidenceUncertain,
		},
		{
			name:           "zero timestamp still detects state",
			entry:          &ClaudeLogEntry{Type: "assistant", StopReason: "end_turn"},
			wantStatus:     domain.AgentWaitingForUser,
			wantConfidence: domain.ConfidenceCertain,
		},
		{
			name:           "future timestamp (clock skew) returns duration 0",
			entry:          &ClaudeLogEntry{Type: "assistant", StopReason: "end_turn", Timestamp: time.Now().Add(1 * time.Hour)},
			wantStatus:     domain.AgentWaitingForUser,
			wantConfidence: domain.ConfidenceCertain,
		},
	}

	d := NewClaudeCodeDetector()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := d.determineState(tt.entry)
			if state.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", state.Status, tt.wantStatus)
			}
			if state.Confidence != tt.wantConfidence {
				t.Errorf("Confidence = %v, want %v", state.Confidence, tt.wantConfidence)
			}
		})
	}
}
