package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewLogEntry(t *testing.T) {
	timestamp := time.Date(2026, 1, 12, 10, 30, 0, 0, time.UTC)
	rawJSON := json.RawMessage(`{"role":"user","content":"hello"}`)
	sessionID := "abc-123"

	entry := NewLogEntry(timestamp, "user", rawJSON, sessionID)

	if !entry.Timestamp.Equal(timestamp) {
		t.Errorf("Timestamp = %v, want %v", entry.Timestamp, timestamp)
	}
	if entry.Type != "user" {
		t.Errorf("Type = %q, want %q", entry.Type, "user")
	}
	if string(entry.RawJSON) != string(rawJSON) {
		t.Errorf("RawJSON = %s, want %s", entry.RawJSON, rawJSON)
	}
	if entry.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", entry.SessionID, sessionID)
	}
}

func TestNewLogSession(t *testing.T) {
	startTime := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)

	session := NewLogSession(
		"uuid-123",
		"/path/to/session.jsonl",
		startTime,
		100,
		5,
		"User asked about Go patterns",
	)

	if session.ID != "uuid-123" {
		t.Errorf("ID = %q, want %q", session.ID, "uuid-123")
	}
	if session.Path != "/path/to/session.jsonl" {
		t.Errorf("Path = %q, want %q", session.Path, "/path/to/session.jsonl")
	}
	if !session.StartTime.Equal(startTime) {
		t.Errorf("StartTime = %v, want %v", session.StartTime, startTime)
	}
	if session.EntryCount != 100 {
		t.Errorf("EntryCount = %d, want %d", session.EntryCount, 100)
	}
	if session.SkippedCount != 5 {
		t.Errorf("SkippedCount = %d, want %d", session.SkippedCount, 5)
	}
	if session.Summary != "User asked about Go patterns" {
		t.Errorf("Summary = %q, want %q", session.Summary, "User asked about Go patterns")
	}
}

func TestLogEntryFields(t *testing.T) {
	tests := []struct {
		name      string
		entryType string
		rawJSON   string
	}{
		{
			name:      "user entry",
			entryType: "user",
			rawJSON:   `{"role":"user"}`,
		},
		{
			name:      "assistant entry",
			entryType: "assistant",
			rawJSON:   `{"role":"assistant","content":"Hello!"}`,
		},
		{
			name:      "system entry",
			entryType: "system",
			rawJSON:   `{"type":"system"}`,
		},
		{
			name:      "summary entry",
			entryType: "summary",
			rawJSON:   `{"type":"summary","text":"Session summary"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewLogEntry(
				time.Now(),
				tt.entryType,
				json.RawMessage(tt.rawJSON),
				"session-1",
			)

			if entry.Type != tt.entryType {
				t.Errorf("Type = %q, want %q", entry.Type, tt.entryType)
			}

			// Verify RawJSON is valid JSON
			var parsed interface{}
			if err := json.Unmarshal(entry.RawJSON, &parsed); err != nil {
				t.Errorf("RawJSON is not valid JSON: %v", err)
			}
		})
	}
}

func TestLogSessionZeroValues(t *testing.T) {
	// Test with zero values to ensure no panics
	session := NewLogSession("", "", time.Time{}, 0, 0, "")

	if session.ID != "" {
		t.Errorf("ID should be empty, got %q", session.ID)
	}
	if session.EntryCount != 0 {
		t.Errorf("EntryCount should be 0, got %d", session.EntryCount)
	}
}
