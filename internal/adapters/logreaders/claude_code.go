// Package logreaders provides log reading implementations for agentic tools.
package logreaders

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

const (
	// claudeProjectsDir is the base directory for Claude Code project logs
	claudeProjectsDir = ".claude/projects"

	// tailPollInterval is how often we check for new log entries (AC4: within 2 seconds)
	tailPollInterval = 2 * time.Second

	// maxLineLength limits individual line size to prevent OOM (1MB)
	maxLineLength = 1024 * 1024
)

// ClaudeCodeReader implements LogReader for Claude Code session logs.
// It reads JSONL files from ~/.claude/projects/{escaped-path}/.
type ClaudeCodeReader struct{}

// Compile-time interface compliance check
var _ ports.LogReader = (*ClaudeCodeReader)(nil)

// NewClaudeCodeReader creates a new Claude Code log reader.
func NewClaudeCodeReader() *ClaudeCodeReader {
	return &ClaudeCodeReader{}
}

// Tool returns the agentic tool name.
func (r *ClaudeCodeReader) Tool() string {
	return "Claude Code"
}

// CanRead checks if Claude Code logs exist for this project.
func (r *ClaudeCodeReader) CanRead(ctx context.Context, projectPath string) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	dir := r.pathToClaudeDir(projectPath)
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ListSessions returns available log sessions sorted by recency (newest first).
func (r *ClaudeCodeReader) ListSessions(ctx context.Context, projectPath string) ([]domain.LogSession, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	dir := r.pathToClaudeDir(projectPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read Claude logs directory: %w", err)
	}

	var sessions []domain.LogSession
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		// Skip agent sub-sessions (agent-*.jsonl)
		if strings.HasPrefix(entry.Name(), "agent-") {
			continue
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		sessionPath := filepath.Join(dir, entry.Name())
		session, err := r.extractSessionMetadata(ctx, sessionPath)
		if err != nil {
			// Skip sessions we can't read, but continue with others
			continue
		}
		sessions = append(sessions, session)
	}

	// Sort by start time, newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	return sessions, nil
}

// ReadSession reads all entries from a session file.
func (r *ClaudeCodeReader) ReadSession(ctx context.Context, sessionPath string) ([]domain.LogEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	file, err := os.Open(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer file.Close()

	var entries []domain.LogEntry
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maxLineLength)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return entries, ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		entry, err := r.parseLogEntry(line)
		if err != nil {
			// Skip invalid lines per spec (log at debug, continue)
			continue
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		// Return partial results on error
		return entries, fmt.Errorf("error reading session: %w", err)
	}

	return entries, nil
}

// TailSession streams new log entries as they are written.
// Caller MUST cancel ctx when done to stop the polling goroutine.
func (r *ClaudeCodeReader) TailSession(ctx context.Context, sessionPath string) (<-chan domain.LogEntry, error) {
	// Verify file exists initially
	info, err := os.Stat(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access session file: %w", err)
	}

	ch := make(chan domain.LogEntry, 100)
	lastOffset := info.Size()

	go func() {
		defer close(ch)
		ticker := time.NewTicker(tailPollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				newEntries, newOffset, err := r.readNewEntries(sessionPath, lastOffset)
				if err != nil {
					// On error, continue trying
					continue
				}
				lastOffset = newOffset

				for _, entry := range newEntries {
					select {
					case <-ctx.Done():
						return
					case ch <- entry:
					}
				}
			}
		}
	}()

	return ch, nil
}

// pathToClaudeDir converts a project path to the Claude logs directory.
// Example: /Users/limjk/GitHub/JeiKeiLim/vibe-dash
//
//	→ ~/.claude/projects/-Users-limjk-GitHub-JeiKeiLim-vibe-dash/
func (r *ClaudeCodeReader) pathToClaudeDir(projectPath string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}

	// Replace / with -
	escapedPath := strings.ReplaceAll(projectPath, "/", "-")
	return filepath.Join(homeDir, claudeProjectsDir, escapedPath)
}

// extractSessionMetadata reads a session file and extracts metadata.
func (r *ClaudeCodeReader) extractSessionMetadata(ctx context.Context, sessionPath string) (domain.LogSession, error) {
	file, err := os.Open(sessionPath)
	if err != nil {
		return domain.LogSession{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return domain.LogSession{}, err
	}

	// Extract session ID from filename (UUID.jsonl)
	sessionID := strings.TrimSuffix(filepath.Base(sessionPath), ".jsonl")

	var entryCount, skippedCount int
	var startTime time.Time
	var summary string

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maxLineLength)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			break
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(line, &raw); err != nil {
			skippedCount++
			continue
		}
		entryCount++

		// Extract timestamp from first entry with timestamp
		if startTime.IsZero() {
			if ts, ok := raw["timestamp"].(string); ok {
				if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
					startTime = parsed
				}
			}
		}

		// Look for summary type
		if entryType, ok := raw["type"].(string); ok && entryType == "summary" {
			if s, ok := raw["summary"].(string); ok {
				summary = s
			}
		}
	}

	// Use file mtime if no timestamp found
	if startTime.IsZero() {
		startTime = info.ModTime()
	}

	return domain.NewLogSession(
		sessionID,
		sessionPath,
		startTime,
		entryCount,
		skippedCount,
		summary,
	), nil
}

// parseLogEntry parses a single JSONL line into a LogEntry.
func (r *ClaudeCodeReader) parseLogEntry(line []byte) (domain.LogEntry, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(line, &raw); err != nil {
		return domain.LogEntry{}, err
	}

	// Extract type
	entryType, _ := raw["type"].(string)
	if entryType == "" {
		// Try to infer from message role
		if msg, ok := raw["message"].(map[string]interface{}); ok {
			if role, ok := msg["role"].(string); ok {
				entryType = role
			}
		}
	}

	// Extract timestamp
	var timestamp time.Time
	if ts, ok := raw["timestamp"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			timestamp = parsed
		}
	}

	// Extract session ID
	sessionID, _ := raw["sessionId"].(string)

	// Keep raw JSON for display
	rawJSON := json.RawMessage(line)

	return domain.NewLogEntry(timestamp, entryType, rawJSON, sessionID), nil
}

// readNewEntries reads entries added since the given offset.
func (r *ClaudeCodeReader) readNewEntries(sessionPath string, offset int64) ([]domain.LogEntry, int64, error) {
	file, err := os.Open(sessionPath)
	if err != nil {
		return nil, offset, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, offset, err
	}

	// No new data
	if info.Size() <= offset {
		return nil, offset, nil
	}

	// Seek to last known position
	if _, err := file.Seek(offset, 0); err != nil {
		return nil, offset, err
	}

	var entries []domain.LogEntry
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maxLineLength)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		entry, err := r.parseLogEntry(line)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, info.Size(), scanner.Err()
}

// PathToClaudeDir is exported for testing purposes.
func PathToClaudeDir(projectPath string) string {
	reader := &ClaudeCodeReader{}
	return reader.pathToClaudeDir(projectPath)
}
