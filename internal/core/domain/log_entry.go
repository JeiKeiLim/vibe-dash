// Package domain defines core business entities.
package domain

import (
	"encoding/json"
	"time"
)

// LogEntry represents a single entry from an agentic tool's session log.
// The raw JSON content is preserved for display, while key fields are extracted
// for filtering and display purposes.
type LogEntry struct {
	// Timestamp when this entry was created
	Timestamp time.Time

	// Type of entry (e.g., "user", "assistant", "system", "summary")
	Type string

	// RawJSON contains the original JSON entry for display
	RawJSON json.RawMessage

	// SessionID identifies the session this entry belongs to
	SessionID string
}

// LogSession represents metadata about a log session.
type LogSession struct {
	// ID is the unique identifier for this session (typically UUID)
	ID string

	// Path is the full filesystem path to the session log file
	Path string

	// StartTime is when the session began (from first entry or file mtime)
	StartTime time.Time

	// EntryCount is the number of successfully parsed entries
	EntryCount int

	// SkippedCount is the number of lines that failed to parse
	SkippedCount int

	// Summary is an optional session summary extracted from log content
	Summary string
}

// NewLogEntry creates a new LogEntry with the given values.
func NewLogEntry(timestamp time.Time, entryType string, rawJSON json.RawMessage, sessionID string) LogEntry {
	return LogEntry{
		Timestamp: timestamp,
		Type:      entryType,
		RawJSON:   rawJSON,
		SessionID: sessionID,
	}
}

// NewLogSession creates a new LogSession with the given values.
func NewLogSession(id, path string, startTime time.Time, entryCount, skippedCount int, summary string) LogSession {
	return LogSession{
		ID:           id,
		Path:         path,
		StartTime:    startTime,
		EntryCount:   entryCount,
		SkippedCount: skippedCount,
		Summary:      summary,
	}
}
