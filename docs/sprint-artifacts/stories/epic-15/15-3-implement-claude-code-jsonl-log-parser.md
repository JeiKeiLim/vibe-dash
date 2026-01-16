# Story 15.3: Implement Claude Code JSONL Log Parser

Status: Done

## Story

As a developer,
I want to parse Claude Code's JSONL log files efficiently,
So that I can extract the last assistant message and its stop_reason for agent state detection.

## User-Visible Changes

None - this is internal infrastructure for the Claude Code agent detector. User-visible changes will come in Story 15.6 when agent detection is integrated into the TUI dashboard.

## Acceptance Criteria

1. **AC1:** Given a JSONL log file with 1000 entries, parser reads only the last N entries (tail-optimized, not full file scan)
2. **AC2:** Given log entry `{"type": "assistant", "message": {...}, "stop_reason": "end_turn", "timestamp": "..."}`, parser extracts type, stop_reason, and timestamp correctly
3. **AC3:** Given malformed JSONL line, parser skips the line gracefully and continues without error
4. **AC4:** Given an empty file, parser returns empty result without error
5. **AC5:** Given file permission errors, parser returns appropriate error
6. **AC6:** Parser respects context cancellation, stopping work within 100ms when cancelled
7. **AC7:** Parser provides `FindMostRecentSession(ctx, claudeDir)` to locate the most recently modified JSONL file
8. **AC8:** Parser provides `ParseLastAssistantEntry(ctx, sessionPath)` returning the last assistant-type log entry

## Tasks / Subtasks

- [x] Task 1: Create ClaudeCodeLogParser struct (AC: 1, 3, 4, 5, 6)
  - [x] 1.1: Create `internal/adapters/agentdetectors/claude_code_log_parser.go`
  - [x] 1.2: Define struct with configurable `tailEntries` count (default 50)
  - [x] 1.3: Implement `NewClaudeCodeLogParser(opts ...Option)` constructor with functional options

- [x] Task 2: Implement FindMostRecentSession method (AC: 7)
  - [x] 2.1: Read directory entries from claudeDir
  - [x] 2.2: Filter for `*.jsonl` files, excluding `agent-*.jsonl` (matching PathMatcher/LogReader behavior)
  - [x] 2.3: Sort by modification time descending
  - [x] 2.4: Return path to most recent file, or empty string if none found
  - [x] 2.5: Respect context cancellation during directory scan

- [x] Task 3: Implement tail-optimized JSONL reading (AC: 1, 3, 4)
  - [x] 3.1: Implement `readTail(ctx, filePath, n int)` to read last N lines
  - [x] 3.2: Use seek-from-end strategy: read backwards from EOF to find N complete lines
  - [x] 3.3: Handle edge case: file smaller than N lines (read entire file)
  - [x] 3.4: Parse each line as JSON, skip malformed lines gracefully
  - [x] 3.5: Return slice of parsed entries in chronological order (oldest first)

- [x] Task 4: Implement ParseLastAssistantEntry method (AC: 2, 8)
  - [x] 4.1: Call readTail to get last N entries
  - [x] 4.2: Iterate in reverse to find most recent assistant entry
  - [x] 4.3: Extract and return: type, stop_reason, timestamp, message content
  - [x] 4.4: Return nil if no assistant entry found in tail

- [x] Task 5: Define ClaudeLogEntry struct for parsed data (AC: 2)
  - [x] 5.1: Create struct with fields: Type, StopReason, Timestamp, RawJSON (optional for debugging)
  - [x] 5.2: Add helper methods: `IsAssistant() bool`, `IsEndTurn() bool`, `IsToolUse() bool`
  - [x] 5.3: StopReason detection: check both `stop_reason` and `message.stop_reason` (Claude format variations)

- [x] Task 6: Handle error cases gracefully (AC: 3, 5, 6)
  - [x] 6.1: File permission errors return wrapped error with context
  - [x] 6.2: Malformed JSON lines logged at debug level using slog, not returned as error
  - [x] 6.3: Context cancellation returns promptly (check before and during I/O operations within 100ms)
  - [x] 6.4: Empty file returns empty slice, nil error

- [x] Task 7: Write comprehensive unit tests (AC: 1-8)
  - [x] 7.1: Create `internal/adapters/agentdetectors/claude_code_log_parser_test.go`
  - [x] 7.2: Test tail reading with file larger than N entries
  - [x] 7.3: Test tail reading with file smaller than N entries
  - [x] 7.4: Test malformed line handling (skip and continue)
  - [x] 7.5: Test empty file handling
  - [x] 7.6: Test FindMostRecentSession with multiple files
  - [x] 7.7: Test FindMostRecentSession with no files
  - [x] 7.8: Test ParseLastAssistantEntry finds correct entry
  - [x] 7.9: Test context cancellation during read
  - [x] 7.10: Test stop_reason extraction from both locations
  - [x] 7.11: Test timestamp parsing with both RFC3339 and RFC3339Nano formats
  - [x] 7.12: Add benchmark test for tail-read performance on 10,000 entry file

- [x] Task 8: Verify integration
  - [x] 8.1: Ensure consistent with `logreaders/claude_code.go` patterns
  - [x] 8.2: Run `make lint && make test` - all must pass

## Dev Notes

### Claude Code JSONL Format

Based on `logreaders/claude_code.go` and PRD analysis, Claude Code writes logs as JSONL where each line is a JSON object:

```jsonl
{"type": "user", "message": {"role": "user", "content": "..."}, "timestamp": "2026-01-16T10:00:00Z"}
{"type": "assistant", "message": {"role": "assistant", "content": [...]}, "stop_reason": "tool_use", "timestamp": "2026-01-16T10:00:05Z"}
{"type": "assistant", "message": {"role": "assistant", "content": [...]}, "stop_reason": "end_turn", "timestamp": "2026-01-16T10:00:10Z"}
```

**Critical Detection Logic (from PRD):**
- `stop_reason: "end_turn"` → Agent is **WAITING FOR USER** (High confidence)
- `stop_reason: "tool_use"` → Agent is **WORKING** (High confidence)

### Struct Definitions

```go
// ClaudeLogEntry represents a parsed log entry from Claude Code JSONL logs.
// Memory-optimized: RawJSON is stored only when debugging is needed.
type ClaudeLogEntry struct {
    Type       string    // "user", "assistant", "system", "summary"
    StopReason string    // "end_turn", "tool_use", etc. (only for assistant)
    Timestamp  time.Time // Entry timestamp (supports RFC3339 and RFC3339Nano)
    RawJSON    []byte    // Original JSON for debugging (optional, can be nil)
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
```

```go
// ClaudeCodeLogParser parses Claude Code JSONL log files with tail optimization.
type ClaudeCodeLogParser struct {
    tailEntries int // Number of entries to read from end (default 50)
}

// Option is a functional option for configuring ClaudeCodeLogParser.
type Option func(*ClaudeCodeLogParser)

// WithTailEntries sets the number of entries to read from the end of the file.
func WithTailEntries(n int) Option {
    return func(p *ClaudeCodeLogParser) {
        if n > 0 {
            p.tailEntries = n
        }
    }
}

const defaultTailEntries = 50

// NewClaudeCodeLogParser creates a new parser with optional configuration.
func NewClaudeCodeLogParser(opts ...Option) *ClaudeCodeLogParser {
    p := &ClaudeCodeLogParser{
        tailEntries: defaultTailEntries,
    }
    for _, opt := range opts {
        opt(p)
    }
    return p
}
```

### Tail-Read Strategy

Reading from the end of the file is critical for performance (files can have 10,000+ entries).

**Algorithm Summary:**
1. For files < 64KB: Read entire file (faster than seek overhead)
2. For larger files: Read 32KB chunks from EOF backwards
3. Accumulate complete lines until N lines found
4. Parse each line as JSON, skip malformed lines
5. Return lines in chronological order (oldest first)

**Buffer Size Rationale:**
- 64KB threshold: Most entries are 200-500 bytes, so 64KB = ~130-320 entries. Full read is efficient.
- 32KB chunks: Balances memory usage with syscall overhead for backward reading.

Strategy:

```go
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

    // For small files, just read everything
    if stat.Size() < 64*1024 { // 64KB threshold
        return p.readAll(ctx, file)
    }

    // Read backwards to find n entries
    return p.readBackwards(ctx, file, stat.Size(), n)
}
```

### Backward Reading Algorithm

```go
const chunkSize = 32 * 1024 // 32KB chunks

func (p *ClaudeCodeLogParser) readBackwards(ctx context.Context, file *os.File, fileSize int64, n int) ([]ClaudeLogEntry, error) {
    // Start from end, work backwards
    offset := fileSize
    var lines [][]byte
    var partial []byte

    for offset > 0 && len(lines) < n {
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

        // Prepend lines (in reverse order since we're reading backwards)
        for i := len(chunkLines) - 1; i >= 0; i-- {
            if len(chunkLines[i]) > 0 {
                lines = append([][]byte{chunkLines[i]}, lines...)
            }
            if len(lines) >= n {
                break
            }
        }
    }

    // Take only last n lines
    if len(lines) > n {
        lines = lines[len(lines)-n:]
    }

    // Parse lines into entries
    var entries []ClaudeLogEntry
    for _, line := range lines {
        entry, err := p.parseLine(line)
        if err != nil {
            // Skip malformed lines (AC3)
            continue
        }
        entries = append(entries, entry)
    }

    return entries, nil
}
```

### Entry Parsing

```go
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
```

### Public API

```go
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
```

### Edge Cases

| Case | Input | Behavior |
|------|-------|----------|
| Empty file | 0 bytes | Return empty slice, nil error |
| Small file | < 50 entries | Read all entries, return all |
| Large file | 10,000 entries | Read last 50 (default), return parsed |
| Malformed line | Invalid JSON | Skip line, log at debug level, continue |
| Permission denied | Can't read file | Return wrapped error with context |
| No assistant entries | User messages only | Return nil from ParseLastAssistantEntry |
| No JSONL files | Empty directory | Return empty string from FindMostRecentSession |
| Context cancelled | Any | Return within 100ms (empty result, no error) |
| agent-*.jsonl files | Sub-sessions | Skip (per PathMatcher/LogReader behavior) |
| RFC3339 timestamp | 2026-01-16T10:00:00Z | Parse successfully |
| RFC3339Nano timestamp | 2026-01-16T10:00:00.123Z | Parse successfully (fallback) |
| Invalid timestamp | Unparseable | Timestamp remains zero value |

### Testing Strategy

```go
func TestReadTail_LargeFile(t *testing.T) {
    // Create temp file with 100 entries
    // Parser configured for tail=20
    // Verify only last 20 returned
}

func TestReadTail_SmallFile(t *testing.T) {
    // Create temp file with 5 entries
    // Parser configured for tail=50
    // Verify all 5 returned
}

func TestReadTail_MalformedLine(t *testing.T) {
    // Create temp file with valid JSON + one malformed line + more valid
    // Verify malformed skipped, valid entries returned
}

func TestReadTail_EmptyFile(t *testing.T) {
    // Create empty temp file
    // Verify empty slice returned, no error
}

func TestFindMostRecentSession_MultipleFiles(t *testing.T) {
    // Create temp dir with 3 .jsonl files, different mtimes
    // Also create agent-test.jsonl (should be skipped)
    // Verify most recent non-agent file returned
}

func TestFindMostRecentSession_NoFiles(t *testing.T) {
    // Create empty temp dir
    // Verify empty string returned
}

func TestParseLastAssistantEntry_FindsCorrect(t *testing.T) {
    // File with: user, assistant (tool_use), user, assistant (end_turn)
    // Verify returns last assistant with end_turn
}

func TestParseLastAssistantEntry_NoAssistant(t *testing.T) {
    // File with only user messages
    // Verify returns nil
}

func TestContextCancellation(t *testing.T) {
    // Cancel context before calling
    // Verify returns quickly with empty result
}

func TestStopReason_BothLocations(t *testing.T) {
    // Test stop_reason at top level
    // Test stop_reason inside message object
    // Verify both work
}

func TestTimestamp_BothFormats(t *testing.T) {
    // Test RFC3339 format: "2026-01-16T10:00:00Z"
    // Test RFC3339Nano format: "2026-01-16T10:00:00.123456789Z"
    // Verify both parse correctly
}

func BenchmarkReadTail_LargeFile(b *testing.B) {
    // Create temp file with 10,000 entries
    // Benchmark tail reading with default 50 entries
    // Target: < 10ms per operation for NFR-P2-1 compliance
}
```

### Hexagonal Architecture

```
internal/adapters/agentdetectors/
├── doc.go                             # Package documentation (from 15.2)
├── claude_code_path_matcher.go        # Path matching (from 15.2)
├── claude_code_path_matcher_test.go   # Tests (from 15.2)
├── claude_code_log_parser.go          # NEW: JSONL parsing (this story)
└── claude_code_log_parser_test.go     # NEW: Tests (this story)

This is an ADAPTER (not core/ports), so:
- Can use os, filepath, json, bytes, strings, log/slog packages
- Can access filesystem directly
- Does NOT implement AgentActivityDetector interface (that's Story 15.4)
- Used as HELPER by ClaudeCodeDetector which is created in Story 15.4
```

**Story 15.3 vs 15.4 Responsibility:**
- **Story 15.3 (this story):** Creates ClaudeCodeLogParser helper for JSONL parsing
- **Story 15.4:** Creates ClaudeCodeDetector that implements AgentActivityDetector interface, USING the parser from 15.3

### Context Cancellation Pattern (AC6 Compliance)

All methods MUST respect context cancellation within 100ms:

```go
// Pattern for context-aware operations:
// 1. Check ctx.Done() BEFORE starting I/O operations
select {
case <-ctx.Done():
    return nil, nil // Return early with empty result, not error
default:
}

// 2. Check ctx.Done() INSIDE loops between operations
for offset > 0 && len(lines) < n {
    select {
    case <-ctx.Done():
        return nil, nil
    default:
    }
    // ... perform I/O ...
}

// 3. Return empty result (nil/empty slice), not ctx.Err()
// This follows the PathMatcher pattern from Story 15.2
```

### Integration with Future Stories

Story 15.4 (Agent State Detection Logic) will use this parser created in Story 15.3:

```go
// In story 15.4:
type ClaudeCodeDetector struct {
    pathMatcher *ClaudeCodePathMatcher  // From Story 15.2
    logParser   *ClaudeCodeLogParser    // From Story 15.3 (this story)
}

func (d *ClaudeCodeDetector) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
    // 1. Find Claude logs directory (Story 15.2)
    claudeDir, err := d.pathMatcher.Match(ctx, projectPath)
    if err != nil || claudeDir == "" {
        return domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
    }

    // 2. Find most recent session (Story 15.3)
    sessionPath, err := d.logParser.FindMostRecentSession(ctx, claudeDir)
    if err != nil || sessionPath == "" {
        return domain.NewAgentState("Claude Code", domain.AgentInactive, 0, domain.ConfidenceCertain), nil
    }

    // 3. Parse last assistant entry (Story 15.3)
    entry, err := d.logParser.ParseLastAssistantEntry(ctx, sessionPath)
    if err != nil || entry == nil {
        return domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
    }

    // 4. Determine state from stop_reason (Story 15.4)
    if entry.IsEndTurn() {
        return domain.NewAgentState("Claude Code", domain.AgentWaitingForUser, time.Since(entry.Timestamp), domain.ConfidenceCertain), nil
    }
    if entry.IsToolUse() {
        return domain.NewAgentState("Claude Code", domain.AgentWorking, time.Since(entry.Timestamp), domain.ConfidenceCertain), nil
    }
    return domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
}
```

### References

- [Source: docs/prd-phase2.md] - FR-P2-1 requirement (detect via JSONL log parsing)
- [Source: docs/epics-phase2.md#Story-3.3] - Story specification
- [Source: docs/project-context.md#Phase-2-Additions] - Detection logic specification
- [Source: internal/adapters/logreaders/claude_code.go] - Existing JSONL parsing patterns (lines 127-155)
- [Source: internal/adapters/logreaders/claude_code.go:79-86] - agent-*.jsonl skip pattern
- [Source: internal/adapters/logreaders/claude_code.go:289-321] - Entry parsing pattern
- [Source: internal/adapters/agentdetectors/claude_code_path_matcher.go] - PathMatcher from Story 15.2
- [Source: docs/sprint-artifacts/stories/epic-15/15-1-define-agentactivitydetector-interface-and-types.md] - Interface patterns
- [Source: docs/sprint-artifacts/stories/epic-15/15-2-implement-claude-code-project-path-matcher.md] - PathMatcher patterns

## Dev Agent Record

### Context Reference

- Phase 2 Epic 15: Sub-1-Minute Agent Detection (THE killer feature)
- FR Coverage: FR-P2-1 (detect Claude Code via JSONL log parsing)
- Prerequisite: Story 15.1 (AgentActivityDetector interface) - DONE
- Prerequisite: Story 15.2 (ClaudeCodePathMatcher) - DONE

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None

### Completion Notes List

- Implemented ClaudeCodeLogParser struct with functional options pattern (WithTailEntries)
- Created ClaudeLogEntry struct with Type, StopReason, Timestamp, RawJSON fields
- Added helper methods: IsAssistant(), IsEndTurn(), IsToolUse() for agent state detection
- Implemented tail-optimized JSONL reading:
  - Small files (<64KB): read all, limit to last N
  - Large files: read backwards from EOF in 32KB chunks
- FindMostRecentSession: filters *.jsonl (excludes agent-*.jsonl), returns most recently modified
- ParseLastAssistantEntry: returns last assistant entry from tail entries
- Context cancellation checks before and during I/O operations (AC6 compliant)
- StopReason extraction checks both top-level and message.stop_reason (Claude format variations)
- Timestamp parsing supports both RFC3339 and RFC3339Nano formats
- Benchmark: ~118μs for 10,000 entry file (well under 10ms NFR target)
- All 1233 tests pass, lint clean (after code review fixes)

### File List

- internal/adapters/agentdetectors/claude_code_log_parser.go (NEW)
- internal/adapters/agentdetectors/claude_code_log_parser_test.go (NEW)
- internal/adapters/agentdetectors/doc.go (MODIFIED - updated story references)
- docs/sprint-artifacts/stories/epic-15/15-3-implement-claude-code-jsonl-log-parser.md (MODIFIED)
- docs/sprint-artifacts/sprint-status.yaml (MODIFIED)

### Change Log

- 2026-01-16: Story implementation complete - ClaudeCodeLogParser with tail-optimized JSONL reading
- 2026-01-16: Code review fixes applied:
  - H1/H4: Added tests for readBackwards path (>64KB files) and partial line handling
  - H2: Updated doc.go with correct story references (15.2, 15.3, 15.4, 15.5)
  - H3: Renamed Option → LogParserOption to avoid name collision
  - M1: Fixed O(n²) memory allocation in readBackwards (collect reversed, reverse once at end)
  - M2: Updated ClaudeLogEntry comment to accurately describe RawJSON usage
  - M3: Added test for type inference from message.role fallback path
