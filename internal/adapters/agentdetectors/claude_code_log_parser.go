package agentdetectors

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultTailEntries = 50
	smallFileThreshold = 64 * 1024 // 64KB - read entire file below this
	chunkSize          = 32 * 1024 // 32KB chunks for backward reading
)

// ClaudeLogEntry represents a parsed log entry from Claude Code JSONL logs.
type ClaudeLogEntry struct {
	Type       string    // "user", "assistant", "system", "summary"
	StopReason string    // "end_turn", "tool_use", etc. (only for assistant)
	Timestamp  time.Time // Entry timestamp (supports RFC3339 and RFC3339Nano)
	RawJSON    []byte    // Original JSON for debugging/troubleshooting
}

// IsAssistant returns true if this is an assistant message entry.
func (e ClaudeLogEntry) IsAssistant() bool {
	return e.Type == "assistant"
}

// IsEndTurn returns true if the assistant completed and is waiting for user.
func (e ClaudeLogEntry) IsEndTurn() bool {
	return e.StopReason == "end_turn"
}

// IsToolUse returns true if the assistant is actively using tools (working).
func (e ClaudeLogEntry) IsToolUse() bool {
	return e.StopReason == "tool_use"
}

// ClaudeCodeLogParser parses Claude Code JSONL log files with tail optimization.
type ClaudeCodeLogParser struct {
	tailEntries int // Number of entries to read from end (default 50)
}

// LogParserOption is a functional option for configuring ClaudeCodeLogParser.
type LogParserOption func(*ClaudeCodeLogParser)

// WithTailEntries sets the number of entries to read from the end of the file.
func WithTailEntries(n int) LogParserOption {
	return func(p *ClaudeCodeLogParser) {
		if n > 0 {
			p.tailEntries = n
		}
	}
}

// NewClaudeCodeLogParser creates a new parser with optional configuration.
func NewClaudeCodeLogParser(opts ...LogParserOption) *ClaudeCodeLogParser {
	p := &ClaudeCodeLogParser{
		tailEntries: defaultTailEntries,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// FindMostRecentSession finds the most recently modified JSONL session file.
// Returns empty string if no sessions found or directory doesn't exist.
func (p *ClaudeCodeLogParser) FindMostRecentSession(ctx context.Context, claudeDir string) (string, error) {
	select {
	case <-ctx.Done():
		return "", nil
	default:
	}

	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var mostRecent string
	var mostRecentTime time.Time

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return "", nil
		default:
		}

		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		// Match PathMatcher/LogReader behavior: *.jsonl but not agent-*.jsonl
		if !strings.HasSuffix(name, ".jsonl") || strings.HasPrefix(name, "agent-") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(mostRecentTime) {
			mostRecentTime = info.ModTime()
			mostRecent = filepath.Join(claudeDir, name)
		}
	}

	return mostRecent, nil
}

// ParseLastAssistantEntry finds the most recent assistant entry in a session.
// Returns nil if no assistant entry found.
func (p *ClaudeCodeLogParser) ParseLastAssistantEntry(ctx context.Context, sessionPath string) (*ClaudeLogEntry, error) {
	entries, err := p.readTail(ctx, sessionPath, p.tailEntries)
	if err != nil {
		return nil, err
	}

	// Search backwards for most recent assistant entry
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].IsAssistant() {
			return &entries[i], nil
		}
	}

	return nil, nil
}

// readTail reads the last n lines from a file efficiently.
// Strategy:
// 1. Seek to end of file
// 2. Read backwards in chunks until n newlines found
// 3. Parse forward from that position
func (p *ClaudeCodeLogParser) readTail(ctx context.Context, filePath string, n int) ([]ClaudeLogEntry, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Empty file check
	if stat.Size() == 0 {
		return []ClaudeLogEntry{}, nil
	}

	// For small files, just read everything and limit at the end
	if stat.Size() < smallFileThreshold {
		entries, err := p.readAll(ctx, file)
		if err != nil {
			return nil, err
		}
		// Limit to last n entries
		if len(entries) > n {
			entries = entries[len(entries)-n:]
		}
		return entries, nil
	}

	// Read backwards to find n entries
	return p.readBackwards(ctx, file, stat.Size(), n)
}

// readAll reads entire file and parses all entries.
func (p *ClaudeCodeLogParser) readAll(ctx context.Context, file *os.File) ([]ClaudeLogEntry, error) {
	var entries []ClaudeLogEntry
	scanner := bufio.NewScanner(file)
	// Use larger buffer for potentially long lines
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		entry, err := p.parseLine(line)
		if err != nil {
			// Skip malformed lines (AC3)
			continue
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("error scanning file: %w", err)
	}

	return entries, nil
}

// readBackwards reads file from end to find last n entries.
func (p *ClaudeCodeLogParser) readBackwards(ctx context.Context, file *os.File, fileSize int64, n int) ([]ClaudeLogEntry, error) {
	// Start from end, work backwards
	offset := fileSize
	// Collect lines in reverse order (newest first), then reverse at the end
	// This avoids O(n²) prepend allocations
	var reversedLines [][]byte
	var partial []byte

	for offset > 0 && len(reversedLines) < n {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
		}

		// Calculate chunk to read
		readSize := int64(chunkSize)
		if readSize > offset {
			readSize = offset
		}
		offset -= readSize

		// Read chunk
		buf := make([]byte, readSize)
		if _, err := file.ReadAt(buf, offset); err != nil {
			return nil, fmt.Errorf("failed to read at offset %d: %w", offset, err)
		}

		// Prepend partial line from previous iteration
		if len(partial) > 0 {
			buf = append(buf, partial...)
			partial = nil
		}

		// Split into lines (reverse order)
		chunkLines := bytes.Split(buf, []byte("\n"))

		// First element might be partial (no leading newline)
		if offset > 0 {
			partial = chunkLines[0]
			chunkLines = chunkLines[1:]
		}

		// Append lines in reverse order (newest first for now)
		for i := len(chunkLines) - 1; i >= 0; i-- {
			if len(chunkLines[i]) > 0 {
				reversedLines = append(reversedLines, chunkLines[i])
			}
			if len(reversedLines) >= n {
				break
			}
		}
	}

	// Handle any remaining partial line from the beginning of file
	if len(partial) > 0 && len(reversedLines) < n {
		reversedLines = append(reversedLines, partial)
	}

	// Take only last n lines (from reversed perspective, first n)
	if len(reversedLines) > n {
		reversedLines = reversedLines[:n]
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(reversedLines)-1; i < j; i, j = i+1, j-1 {
		reversedLines[i], reversedLines[j] = reversedLines[j], reversedLines[i]
	}

	// Parse lines into entries
	var entries []ClaudeLogEntry
	for _, line := range reversedLines {
		entry, err := p.parseLine(line)
		if err != nil {
			// Skip malformed lines (AC3)
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// parseLine parses a single JSONL line into a ClaudeLogEntry.
func (p *ClaudeCodeLogParser) parseLine(line []byte) (ClaudeLogEntry, error) {
	if len(line) == 0 {
		return ClaudeLogEntry{}, fmt.Errorf("empty line")
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(line, &raw); err != nil {
		// Debug log only - skip malformed lines gracefully (AC3)
		slog.Debug("skipping malformed log line", "error", err)
		return ClaudeLogEntry{}, err
	}

	entry := ClaudeLogEntry{
		RawJSON: line,
	}

	// Extract type
	if t, ok := raw["type"].(string); ok {
		entry.Type = t
	} else if msg, ok := raw["message"].(map[string]interface{}); ok {
		// Fallback: infer from message.role
		if role, ok := msg["role"].(string); ok {
			entry.Type = role
		}
	}

	// Extract stop_reason (check both locations)
	if sr, ok := raw["stop_reason"].(string); ok {
		entry.StopReason = sr
	} else if msg, ok := raw["message"].(map[string]interface{}); ok {
		if sr, ok := msg["stop_reason"].(string); ok {
			entry.StopReason = sr
		}
	}

	// Extract timestamp (support both RFC3339 and RFC3339Nano for Claude format variations)
	if ts, ok := raw["timestamp"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			entry.Timestamp = parsed
		} else if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			entry.Timestamp = parsed
		}
		// If neither format works, Timestamp remains zero value (handled gracefully)
	}

	return entry, nil
}
