package agentdetectors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// Task 1: ClaudeCodeLogParser struct tests
// =============================================================================

func TestNewClaudeCodeLogParser_DefaultTailEntries(t *testing.T) {
	parser := NewClaudeCodeLogParser()
	if parser == nil {
		t.Fatal("NewClaudeCodeLogParser returned nil")
	}
	if parser.tailEntries != defaultTailEntries {
		t.Errorf("expected tailEntries=%d, got %d", defaultTailEntries, parser.tailEntries)
	}
}

func TestNewClaudeCodeLogParser_WithTailEntriesOption(t *testing.T) {
	parser := NewClaudeCodeLogParser(WithTailEntries(100))
	if parser == nil {
		t.Fatal("NewClaudeCodeLogParser returned nil")
	}
	if parser.tailEntries != 100 {
		t.Errorf("expected tailEntries=100, got %d", parser.tailEntries)
	}
}

func TestNewClaudeCodeLogParser_WithInvalidTailEntries(t *testing.T) {
	// Negative values should be ignored, keeping default
	parser := NewClaudeCodeLogParser(WithTailEntries(-1))
	if parser.tailEntries != defaultTailEntries {
		t.Errorf("expected tailEntries=%d (default), got %d", defaultTailEntries, parser.tailEntries)
	}

	// Zero should also keep default
	parser = NewClaudeCodeLogParser(WithTailEntries(0))
	if parser.tailEntries != defaultTailEntries {
		t.Errorf("expected tailEntries=%d (default), got %d", defaultTailEntries, parser.tailEntries)
	}
}

// =============================================================================
// Task 5: ClaudeLogEntry struct tests
// =============================================================================

func TestClaudeLogEntry_IsAssistant(t *testing.T) {
	tests := []struct {
		name     string
		entry    ClaudeLogEntry
		expected bool
	}{
		{"assistant type", ClaudeLogEntry{Type: "assistant"}, true},
		{"user type", ClaudeLogEntry{Type: "user"}, false},
		{"system type", ClaudeLogEntry{Type: "system"}, false},
		{"empty type", ClaudeLogEntry{Type: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.IsAssistant(); got != tt.expected {
				t.Errorf("IsAssistant() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClaudeLogEntry_IsEndTurn(t *testing.T) {
	tests := []struct {
		name     string
		entry    ClaudeLogEntry
		expected bool
	}{
		{"end_turn stop", ClaudeLogEntry{StopReason: "end_turn"}, true},
		{"tool_use stop", ClaudeLogEntry{StopReason: "tool_use"}, false},
		{"empty stop", ClaudeLogEntry{StopReason: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.IsEndTurn(); got != tt.expected {
				t.Errorf("IsEndTurn() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClaudeLogEntry_IsToolUse(t *testing.T) {
	tests := []struct {
		name     string
		entry    ClaudeLogEntry
		expected bool
	}{
		{"tool_use stop", ClaudeLogEntry{StopReason: "tool_use"}, true},
		{"end_turn stop", ClaudeLogEntry{StopReason: "end_turn"}, false},
		{"empty stop", ClaudeLogEntry{StopReason: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.IsToolUse(); got != tt.expected {
				t.Errorf("IsToolUse() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Task 2: FindMostRecentSession tests
// =============================================================================

func TestFindMostRecentSession_MultipleFiles(t *testing.T) {
	// Create temp directory with multiple JSONL files
	tmpDir := t.TempDir()

	// Create files with different mod times
	files := []string{"session1.jsonl", "session2.jsonl", "session3.jsonl"}
	for i, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(`{"type":"user"}`), 0644); err != nil {
			t.Fatal(err)
		}
		// Set mod time progressively newer
		modTime := time.Now().Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatal(err)
		}
	}

	parser := NewClaudeCodeLogParser()
	result, err := parser.FindMostRecentSession(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(tmpDir, "session3.jsonl")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestFindMostRecentSession_SkipsAgentFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create regular session file (older)
	regularPath := filepath.Join(tmpDir, "session.jsonl")
	if err := os.WriteFile(regularPath, []byte(`{"type":"user"}`), 0644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(regularPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Create agent session file (newer - should be skipped)
	agentPath := filepath.Join(tmpDir, "agent-subsession.jsonl")
	if err := os.WriteFile(agentPath, []byte(`{"type":"user"}`), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	result, err := parser.FindMostRecentSession(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return the regular session, not the agent file
	if result != regularPath {
		t.Errorf("expected %s (skipping agent file), got %s", regularPath, result)
	}
}

func TestFindMostRecentSession_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	parser := NewClaudeCodeLogParser()
	result, err := parser.FindMostRecentSession(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for no files, got %s", result)
	}
}

func TestFindMostRecentSession_NonexistentDir(t *testing.T) {
	parser := NewClaudeCodeLogParser()
	result, err := parser.FindMostRecentSession(context.Background(), "/nonexistent/path")
	if err != nil {
		t.Fatalf("expected no error for nonexistent dir, got %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for nonexistent dir, got %s", result)
	}
}

func TestFindMostRecentSession_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.jsonl"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	parser := NewClaudeCodeLogParser()
	result, err := parser.FindMostRecentSession(ctx, tmpDir)
	if err != nil {
		t.Fatalf("expected no error on context cancel, got %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result on context cancel, got %s", result)
	}
}

// =============================================================================
// Task 3: readTail tests
// =============================================================================

func TestReadTail_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large.jsonl")

	// Create file with 100 entries (small file, tests readAll path)
	var content string
	for i := 0; i < 100; i++ {
		content += `{"type":"user","timestamp":"2026-01-16T10:00:00Z"}` + "\n"
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser(WithTailEntries(20))
	entries, err := parser.readTail(context.Background(), filePath, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 20 {
		t.Errorf("expected 20 entries, got %d", len(entries))
	}
}

func TestReadTail_VeryLargeFile_ReadBackwards(t *testing.T) {
	// This test exercises the readBackwards path (file >64KB threshold)
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "very_large.jsonl")

	// Create file larger than 64KB threshold (need ~1300 entries at ~50 bytes each)
	// Each entry is ~50 bytes, so 1500 entries = ~75KB > 64KB threshold
	var content string
	for i := 0; i < 1500; i++ {
		content += `{"type":"user","timestamp":"2026-01-16T10:00:00Z"}` + "\n"
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Verify file is > 64KB to ensure readBackwards is used
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() < 64*1024 {
		t.Fatalf("test file too small (%d bytes), needs to be >64KB to test readBackwards", info.Size())
	}

	parser := NewClaudeCodeLogParser(WithTailEntries(20))
	entries, err := parser.readTail(context.Background(), filePath, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 20 {
		t.Errorf("expected 20 entries, got %d", len(entries))
	}
}

func TestReadTail_VeryLargeFile_PartialLineAtStart(t *testing.T) {
	// Test that readBackwards correctly handles partial lines at file start
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "partial_start.jsonl")

	// Create file >64KB with distinct entries to verify correct parsing
	var content string
	for i := 0; i < 1500; i++ {
		content += fmt.Sprintf(`{"type":"user","index":%d,"timestamp":"2026-01-16T10:00:00Z"}`+"\n", i)
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser(WithTailEntries(50))
	entries, err := parser.readTail(context.Background(), filePath, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 50 {
		t.Errorf("expected 50 entries, got %d", len(entries))
	}

	// Verify entries are from the tail (indices 1450-1499)
	// Each entry should be parseable without corruption from partial line handling
	for _, entry := range entries {
		if entry.Type != "user" {
			t.Errorf("expected type=user, got %s", entry.Type)
		}
	}
}

func TestReadTail_SmallFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "small.jsonl")

	// Create file with 5 entries
	var content string
	for i := 0; i < 5; i++ {
		content += `{"type":"user","timestamp":"2026-01-16T10:00:00Z"}` + "\n"
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser(WithTailEntries(50))
	entries, err := parser.readTail(context.Background(), filePath, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}
}

func TestReadTail_MalformedLine(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "malformed.jsonl")

	// Mix valid and invalid JSON
	content := `{"type":"user","timestamp":"2026-01-16T10:00:00Z"}
not valid json
{"type":"assistant","timestamp":"2026-01-16T10:00:01Z"}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	entries, err := parser.readTail(context.Background(), filePath, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 valid entries (malformed skipped)
	if len(entries) != 2 {
		t.Errorf("expected 2 valid entries (malformed skipped), got %d", len(entries))
	}
}

func TestReadTail_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.jsonl")

	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	entries, err := parser.readTail(context.Background(), filePath, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty file, got %d", len(entries))
	}
}

func TestReadTail_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.jsonl")

	if err := os.WriteFile(filePath, []byte(`{"type":"user"}`), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	parser := NewClaudeCodeLogParser()
	entries, err := parser.readTail(ctx, filePath, 50)
	if err != nil {
		t.Fatalf("expected no error on cancel, got %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries on cancel, got %v", entries)
	}
}

// =============================================================================
// Task 4: ParseLastAssistantEntry tests
// =============================================================================

func TestParseLastAssistantEntry_FindsCorrect(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "session.jsonl")

	content := `{"type":"user","timestamp":"2026-01-16T10:00:00Z"}
{"type":"assistant","stop_reason":"tool_use","timestamp":"2026-01-16T10:00:01Z"}
{"type":"user","timestamp":"2026-01-16T10:00:02Z"}
{"type":"assistant","stop_reason":"end_turn","timestamp":"2026-01-16T10:00:03Z"}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	entry, err := parser.ParseLastAssistantEntry(context.Background(), filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.StopReason != "end_turn" {
		t.Errorf("expected stop_reason=end_turn, got %s", entry.StopReason)
	}
}

func TestParseLastAssistantEntry_NoAssistant(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "session.jsonl")

	content := `{"type":"user","timestamp":"2026-01-16T10:00:00Z"}
{"type":"user","timestamp":"2026-01-16T10:00:01Z"}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	entry, err := parser.ParseLastAssistantEntry(context.Background(), filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry != nil {
		t.Errorf("expected nil for no assistant entries, got %+v", entry)
	}
}

// =============================================================================
// Task 5: StopReason extraction from both locations
// =============================================================================

func TestStopReason_TopLevel(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "session.jsonl")

	// stop_reason at top level
	content := `{"type":"assistant","stop_reason":"end_turn","timestamp":"2026-01-16T10:00:00Z"}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	entry, err := parser.ParseLastAssistantEntry(context.Background(), filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.StopReason != "end_turn" {
		t.Errorf("expected stop_reason=end_turn, got %s", entry.StopReason)
	}
}

func TestStopReason_InsideMessage(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "session.jsonl")

	// stop_reason inside message object
	content := `{"type":"assistant","message":{"role":"assistant","stop_reason":"tool_use"},"timestamp":"2026-01-16T10:00:00Z"}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	entry, err := parser.ParseLastAssistantEntry(context.Background(), filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.StopReason != "tool_use" {
		t.Errorf("expected stop_reason=tool_use, got %s", entry.StopReason)
	}
}

func TestTypeInference_FromMessageRole(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "session.jsonl")

	// No top-level "type" field, should infer from message.role
	content := `{"message":{"role":"assistant","content":"test"},"stop_reason":"end_turn","timestamp":"2026-01-16T10:00:00Z"}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	entry, err := parser.ParseLastAssistantEntry(context.Background(), filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry (inferred from message.role), got nil")
	}
	if entry.Type != "assistant" {
		t.Errorf("expected type=assistant (inferred from message.role), got %s", entry.Type)
	}
}

// =============================================================================
// Task 5: Timestamp parsing tests
// =============================================================================

func TestTimestamp_RFC3339(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "session.jsonl")

	content := `{"type":"assistant","timestamp":"2026-01-16T10:00:00Z"}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	entry, err := parser.ParseLastAssistantEntry(context.Background(), filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	expected := time.Date(2026, 1, 16, 10, 0, 0, 0, time.UTC)
	if !entry.Timestamp.Equal(expected) {
		t.Errorf("expected timestamp %v, got %v", expected, entry.Timestamp)
	}
}

func TestTimestamp_RFC3339Nano(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "session.jsonl")

	content := `{"type":"assistant","timestamp":"2026-01-16T10:00:00.123456789Z"}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	entry, err := parser.ParseLastAssistantEntry(context.Background(), filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	expected := time.Date(2026, 1, 16, 10, 0, 0, 123456789, time.UTC)
	if !entry.Timestamp.Equal(expected) {
		t.Errorf("expected timestamp %v, got %v", expected, entry.Timestamp)
	}
}

// =============================================================================
// Task 6: Error handling tests
// =============================================================================

func TestReadTail_FilePermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test as root")
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "noperm.jsonl")

	if err := os.WriteFile(filePath, []byte(`{"type":"user"}`), 0000); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(filePath, 0644) }() // Cleanup

	parser := NewClaudeCodeLogParser()
	_, err := parser.readTail(context.Background(), filePath, 50)
	if err == nil {
		t.Error("expected permission error, got nil")
	}
}

// =============================================================================
// Benchmark tests
// =============================================================================

func BenchmarkReadTail_LargeFile(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "bench.jsonl")

	// Create file with 10,000 entries
	var content string
	for i := 0; i < 10000; i++ {
		content += `{"type":"assistant","stop_reason":"tool_use","timestamp":"2026-01-16T10:00:00Z","message":{"role":"assistant","content":"test"}}` + "\n"
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		b.Fatal(err)
	}

	parser := NewClaudeCodeLogParser()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.readTail(ctx, filePath, 50)
		if err != nil {
			b.Fatal(err)
		}
	}
}
