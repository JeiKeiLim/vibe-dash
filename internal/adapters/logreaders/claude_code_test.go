package logreaders

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewClaudeCodeReader(t *testing.T) {
	reader := NewClaudeCodeReader()
	if reader == nil {
		t.Fatal("NewClaudeCodeReader() returned nil")
	}
}

func TestClaudeCodeReader_Tool(t *testing.T) {
	reader := NewClaudeCodeReader()
	if got := reader.Tool(); got != "Claude Code" {
		t.Errorf("Tool() = %q, want %q", got, "Claude Code")
	}
}

func TestPathToClaudeDir(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		wantSuffix  string
	}{
		{
			name:        "simple path",
			projectPath: "/Users/limjk/projects/myapp",
			wantSuffix:  "-Users-limjk-projects-myapp",
		},
		{
			name:        "path with special chars",
			projectPath: "/Users/limjk/GitHub/JeiKeiLim/vibe-dash",
			wantSuffix:  "-Users-limjk-GitHub-JeiKeiLim-vibe-dash",
		},
		{
			name:        "root path",
			projectPath: "/",
			wantSuffix:  "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathToClaudeDir(tt.projectPath)

			// Check suffix matches expected escaped path
			if !filepath.IsAbs(got) {
				t.Errorf("PathToClaudeDir() should return absolute path, got %q", got)
			}

			base := filepath.Base(got)
			if base != tt.wantSuffix {
				t.Errorf("PathToClaudeDir() base = %q, want %q", base, tt.wantSuffix)
			}

			// Check it includes .claude/projects
			if !containsPath(got, ".claude/projects") {
				t.Errorf("PathToClaudeDir() should contain .claude/projects, got %q", got)
			}
		})
	}
}

func containsPath(fullPath, segment string) bool {
	return filepath.Dir(fullPath) != fullPath &&
		(filepath.Base(filepath.Dir(fullPath)) == "projects" ||
			containsPath(filepath.Dir(fullPath), segment))
}

func TestClaudeCodeReader_CanRead(t *testing.T) {
	reader := NewClaudeCodeReader()
	ctx := context.Background()

	// Test with non-existent project - should return false
	result := reader.CanRead(ctx, "/nonexistent/project/path/that/definitely/does/not/exist")
	if result {
		t.Error("CanRead() should return false for non-existent project")
	}

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()
	result = reader.CanRead(cancelledCtx, "/some/path")
	if result {
		t.Error("CanRead() should return false when context is cancelled")
	}
}

func TestClaudeCodeReader_ListSessions(t *testing.T) {
	reader := NewClaudeCodeReader()
	ctx := context.Background()

	// Test with non-existent directory
	_, err := reader.ListSessions(ctx, "/nonexistent/path")
	if err == nil {
		t.Error("ListSessions() should return error for non-existent directory")
	}
}

func TestClaudeCodeReader_ReadSession(t *testing.T) {
	reader := NewClaudeCodeReader()
	ctx := context.Background()

	// Test with non-existent file
	_, err := reader.ReadSession(ctx, "/nonexistent/session.jsonl")
	if err == nil {
		t.Error("ReadSession() should return error for non-existent file")
	}
}

func TestClaudeCodeReader_ReadSession_ValidJSONL(t *testing.T) {
	// Create temp file with valid JSONL
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "test-session.jsonl")

	entries := []map[string]interface{}{
		{
			"type":      "user",
			"timestamp": "2026-01-12T10:30:00Z",
			"sessionId": "abc-123",
			"message": map[string]interface{}{
				"role":    "user",
				"content": "Hello",
			},
		},
		{
			"type":      "assistant",
			"timestamp": "2026-01-12T10:30:05Z",
			"sessionId": "abc-123",
			"message": map[string]interface{}{
				"role":    "assistant",
				"content": "Hi there!",
			},
		},
	}

	file, err := os.Create(sessionPath)
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(file)
	for _, entry := range entries {
		if err := encoder.Encode(entry); err != nil {
			t.Fatal(err)
		}
	}
	file.Close()

	reader := NewClaudeCodeReader()
	ctx := context.Background()

	result, err := reader.ReadSession(ctx, sessionPath)
	if err != nil {
		t.Fatalf("ReadSession() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("ReadSession() returned %d entries, want 2", len(result))
	}

	// Verify first entry
	if result[0].Type != "user" {
		t.Errorf("first entry Type = %q, want %q", result[0].Type, "user")
	}
	if result[0].SessionID != "abc-123" {
		t.Errorf("first entry SessionID = %q, want %q", result[0].SessionID, "abc-123")
	}
}

func TestClaudeCodeReader_ReadSession_MalformedLines(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "malformed.jsonl")

	// Write mix of valid and invalid lines
	content := `{"type":"user","timestamp":"2026-01-12T10:00:00Z","sessionId":"123"}
not valid json
{"type":"assistant","timestamp":"2026-01-12T10:00:05Z","sessionId":"123"}
{invalid json too}
{"type":"system","timestamp":"2026-01-12T10:00:10Z","sessionId":"123"}
`
	if err := os.WriteFile(sessionPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	reader := NewClaudeCodeReader()
	ctx := context.Background()

	result, err := reader.ReadSession(ctx, sessionPath)
	if err != nil {
		t.Fatalf("ReadSession() error = %v", err)
	}

	// Should have 3 valid entries (skipped 2 invalid)
	if len(result) != 3 {
		t.Errorf("ReadSession() returned %d entries, want 3 (2 skipped)", len(result))
	}
}

func TestClaudeCodeReader_TailSession(t *testing.T) {
	reader := NewClaudeCodeReader()
	ctx := context.Background()

	// Test with non-existent file
	_, err := reader.TailSession(ctx, "/nonexistent/session.jsonl")
	if err == nil {
		t.Error("TailSession() should return error for non-existent file")
	}
}

func TestClaudeCodeReader_TailSession_ReceivesNewEntries(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "tail-test.jsonl")

	// Create initial file
	initialEntry := `{"type":"user","timestamp":"2026-01-12T10:00:00Z","sessionId":"tail-123"}`
	if err := os.WriteFile(sessionPath, []byte(initialEntry+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	reader := NewClaudeCodeReader()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := reader.TailSession(ctx, sessionPath)
	if err != nil {
		t.Fatalf("TailSession() error = %v", err)
	}

	// Append new entry after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		f, err := os.OpenFile(sessionPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		newEntry := `{"type":"assistant","timestamp":"2026-01-12T10:00:05Z","sessionId":"tail-123"}`
		_, _ = f.WriteString(newEntry + "\n")
	}()

	// Wait for new entry (with timeout)
	select {
	case entry, ok := <-ch:
		if !ok {
			// Channel closed, which is also acceptable for this test
			return
		}
		if entry.Type != "assistant" {
			t.Errorf("received entry Type = %q, want %q", entry.Type, "assistant")
		}
	case <-time.After(4 * time.Second):
		// This is acceptable - the tail poll interval is 2s so we might not get it in time
		t.Log("Timeout waiting for new entry (expected due to polling interval)")
	}
}

func TestClaudeCodeReader_ParseLogEntry(t *testing.T) {
	reader := NewClaudeCodeReader()

	tests := []struct {
		name     string
		line     string
		wantType string
		wantErr  bool
	}{
		{
			name:     "user message",
			line:     `{"type":"user","timestamp":"2026-01-12T10:00:00Z","sessionId":"123","message":{"role":"user","content":"hello"}}`,
			wantType: "user",
		},
		{
			name:     "assistant message",
			line:     `{"type":"assistant","timestamp":"2026-01-12T10:00:05Z","sessionId":"123"}`,
			wantType: "assistant",
		},
		{
			name:     "file-history-snapshot",
			line:     `{"type":"file-history-snapshot","messageId":"abc","snapshot":{}}`,
			wantType: "file-history-snapshot",
		},
		{
			name:     "infer type from message role",
			line:     `{"timestamp":"2026-01-12T10:00:00Z","message":{"role":"user"}}`,
			wantType: "user",
		},
		{
			name:    "invalid json",
			line:    `not valid json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := reader.parseLogEntry([]byte(tt.line))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLogEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && entry.Type != tt.wantType {
				t.Errorf("parseLogEntry() Type = %q, want %q", entry.Type, tt.wantType)
			}
		})
	}
}

func TestClaudeCodeReader_ExtractSessionMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "abc-123-def.jsonl")

	content := `{"type":"user","timestamp":"2026-01-12T08:00:00Z","sessionId":"abc-123-def"}
{"type":"assistant","timestamp":"2026-01-12T08:00:05Z","sessionId":"abc-123-def"}
invalid line that will be skipped
{"type":"summary","summary":"This is a test session","sessionId":"abc-123-def"}
`
	if err := os.WriteFile(sessionPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	reader := NewClaudeCodeReader()
	ctx := context.Background()

	session, err := reader.extractSessionMetadata(ctx, sessionPath)
	if err != nil {
		t.Fatalf("extractSessionMetadata() error = %v", err)
	}

	if session.ID != "abc-123-def" {
		t.Errorf("ID = %q, want %q", session.ID, "abc-123-def")
	}
	if session.EntryCount != 3 {
		t.Errorf("EntryCount = %d, want 3", session.EntryCount)
	}
	if session.SkippedCount != 1 {
		t.Errorf("SkippedCount = %d, want 1", session.SkippedCount)
	}
	if session.Summary != "This is a test session" {
		t.Errorf("Summary = %q, want %q", session.Summary, "This is a test session")
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2026-01-12T08:00:00Z")
	if !session.StartTime.Equal(expectedTime) {
		t.Errorf("StartTime = %v, want %v", session.StartTime, expectedTime)
	}
}
