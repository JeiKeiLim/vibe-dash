# Story 15.4: Implement Agent State Detection Logic

Status: done

## Story

As a developer,
I want to determine agent state from parsed log entries,
So that the dashboard can show Working/WaitingForUser/Inactive status.

## User-Visible Changes

None - this is internal infrastructure completing the Claude Code detector. User-visible changes will come in Story 15.6 when agent detection is integrated into the TUI dashboard.

## Acceptance Criteria

1. **AC1:** Given last assistant message has `stop_reason: "end_turn"`, detection returns `WaitingForUser` with `ConfidenceCertain`
2. **AC2:** Given last assistant message has `stop_reason: "tool_use"`, detection returns `Working` with `ConfidenceCertain`
3. **AC3:** Given no assistant messages in last N entries, detection returns `Inactive` with `ConfidenceCertain`
4. **AC4:** Given last message timestamp is 2 hours ago, Duration field is set to 2 hours (time since last activity). If timestamp parsing failed, Duration is 0.
5. **AC5:** ClaudeCodeDetector implements `AgentActivityDetector` interface from Story 15.1
6. **AC6:** ClaudeCodeDetector uses `ClaudeCodePathMatcher` (Story 15.2) and `ClaudeCodeLogParser` (Story 15.3)
7. **AC7:** Detection respects context cancellation, returning within 100ms when cancelled
8. **AC8:** Given Claude Code logs don't exist for project, detection returns `AgentUnknown` with `ConfidenceUncertain` (no error)

## Tasks / Subtasks

- [x] Task 1: Create ClaudeCodeDetector struct (AC: 5, 6)
  - [x] 1.1: Create `internal/adapters/agentdetectors/claude_code_detector.go`
  - [x] 1.2: Define struct with `pathMatcher *ClaudeCodePathMatcher` and `logParser *ClaudeCodeLogParser` fields
  - [x] 1.3: Implement `NewClaudeCodeDetector(opts ...DetectorOption)` constructor with functional options
  - [x] 1.4: Add `WithPathMatcher(*ClaudeCodePathMatcher)` option for dependency injection (testing)
  - [x] 1.5: Add `WithLogParser(*ClaudeCodeLogParser)` option for dependency injection (testing)
  - [x] 1.6: Default constructor creates new PathMatcher and LogParser if not provided

- [x] Task 2: Implement Name() method (AC: 5)
  - [x] 2.1: Return "Claude Code" constant string

- [x] Task 3: Implement Detect() method orchestration (AC: 5, 6, 7, 8)
  - [x] 3.1: Check `ctx.Done()` before starting (respect cancellation)
  - [x] 3.2: Call `pathMatcher.Match(ctx, projectPath)` to find Claude logs directory
  - [x] 3.3: If empty string returned AND err is nil → return `AgentUnknown` with `ConfidenceUncertain`
  - [x] 3.4: If err is not nil → return the error to caller (unexpected filesystem error)
  - [x] 3.5: Call `logParser.FindMostRecentSession(ctx, claudeDir)` to find latest session
  - [x] 3.6: If empty string returned → return `AgentInactive` with `ConfidenceCertain` (Claude logs exist but no active session)
  - [x] 3.7: Call `logParser.ParseLastAssistantEntry(ctx, sessionPath)` to get last assistant entry
  - [x] 3.8: If nil returned → return `AgentInactive` with `ConfidenceCertain` (session has no assistant entries)
  - [x] 3.9: Delegate to `determineState(*ClaudeLogEntry)` for state interpretation

- [x] Task 4: Implement determineState helper method (AC: 1, 2, 3, 4)
  - [x] 4.1: Create private `determineState(entry *ClaudeLogEntry) domain.AgentState` method
  - [x] 4.2: Calculate duration: `time.Since(entry.Timestamp)`. If Timestamp is zero, use duration 0.
  - [x] 4.3: If `entry.IsEndTurn()` → return `AgentWaitingForUser` with `ConfidenceCertain`
  - [x] 4.4: If `entry.IsToolUse()` → return `AgentWorking` with `ConfidenceCertain`
  - [x] 4.5: Otherwise → return `AgentUnknown` with `ConfidenceUncertain` (unrecognized stop_reason)

- [x] Task 5: Write comprehensive unit tests (AC: 1-8)
  - [x] 5.1: Create `internal/adapters/agentdetectors/claude_code_detector_test.go`
  - [x] 5.2: Test `Name()` returns "Claude Code"
  - [x] 5.3: Test `Detect()` with end_turn stop_reason returns WaitingForUser
  - [x] 5.4: Test `Detect()` with tool_use stop_reason returns Working
  - [x] 5.5: Test `Detect()` with no assistant entries returns Inactive
  - [x] 5.6: Test `Detect()` with no Claude logs returns Unknown (no error)
  - [x] 5.7: Test `Detect()` with cancelled context returns promptly (< 100ms)
  - [x] 5.8: Test Duration calculation is approximately correct (within 1 second tolerance)
  - [x] 5.9: Use real PathMatcher/LogParser with temp directories for integration-style tests
  - [x] 5.10: Test unrecognized stop_reason (e.g., "max_tokens") returns Unknown

- [x] Task 6: Verify integration and compliance
  - [x] 6.1: Ensure ClaudeCodeDetector satisfies `ports.AgentActivityDetector` interface (compile-time check)
  - [x] 6.2: Run `make lint && make test` - all must pass
  - [x] 6.3: Update `doc.go` to document ClaudeCodeDetector as the main implementation

## Dev Notes

### Detection Flow Overview

```
1. Check context cancellation → return early if cancelled
2. PathMatcher.Match(projectPath) → find Claude logs directory
   - Empty + no error → AgentUnknown (no Claude logs)
   - Error → propagate error
3. LogParser.FindMostRecentSession() → find latest session file
   - Empty → AgentInactive (logs exist, no sessions)
4. LogParser.ParseLastAssistantEntry() → get last assistant message
   - nil → AgentInactive (session has no assistant entries)
5. determineState(entry) → interpret stop_reason
   - "end_turn" → AgentWaitingForUser
   - "tool_use" → AgentWorking
   - other → AgentUnknown
```

### Struct Definition

```go
// ClaudeCodeDetector detects agent activity state by parsing Claude Code JSONL logs.
// Implements ports.AgentActivityDetector interface.
type ClaudeCodeDetector struct {
    pathMatcher *ClaudeCodePathMatcher
    logParser   *ClaudeCodeLogParser
}

// DetectorOption is a functional option for configuring ClaudeCodeDetector.
type DetectorOption func(*ClaudeCodeDetector)

// WithPathMatcher sets a custom path matcher (for testing).
func WithPathMatcher(pm *ClaudeCodePathMatcher) DetectorOption {
    return func(d *ClaudeCodeDetector) {
        d.pathMatcher = pm
    }
}

// WithLogParser sets a custom log parser (for testing).
func WithLogParser(lp *ClaudeCodeLogParser) DetectorOption {
    return func(d *ClaudeCodeDetector) {
        d.logParser = lp
    }
}

// NewClaudeCodeDetector creates a new detector with optional configuration.
func NewClaudeCodeDetector(opts ...DetectorOption) *ClaudeCodeDetector {
    d := &ClaudeCodeDetector{}
    for _, opt := range opts {
        opt(d)
    }
    if d.pathMatcher == nil {
        d.pathMatcher = NewClaudeCodePathMatcher()
    }
    if d.logParser == nil {
        d.logParser = NewClaudeCodeLogParser()
    }
    return d
}

// Compile-time interface compliance check
var _ ports.AgentActivityDetector = (*ClaudeCodeDetector)(nil)
```

### Detect Method Implementation

```go
func (d *ClaudeCodeDetector) Name() string {
    return "Claude Code"
}

func (d *ClaudeCodeDetector) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
    // Respect context cancellation at entry
    select {
    case <-ctx.Done():
        return domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
    default:
    }

    // Step 1: Find Claude logs directory
    claudeDir, err := d.pathMatcher.Match(ctx, projectPath)
    if err != nil {
        // Unexpected error (permissions, etc.) - propagate
        return domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain), err
    }
    if claudeDir == "" {
        // Claude Code not installed or no logs for this project - graceful
        return domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
    }

    // Step 2: Find most recent session
    sessionPath, err := d.logParser.FindMostRecentSession(ctx, claudeDir)
    if err != nil {
        return domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain), err
    }
    if sessionPath == "" {
        // Logs directory exists but no session files - inactive
        return domain.NewAgentState("Claude Code", domain.AgentInactive, 0, domain.ConfidenceCertain), nil
    }

    // Step 3: Parse last assistant entry
    entry, err := d.logParser.ParseLastAssistantEntry(ctx, sessionPath)
    if err != nil {
        return domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain), err
    }
    if entry == nil {
        // Session exists but no assistant entries - inactive
        return domain.NewAgentState("Claude Code", domain.AgentInactive, 0, domain.ConfidenceCertain), nil
    }

    // Step 4: Determine state from entry
    return d.determineState(entry), nil
}
```

### determineState Helper

```go
func (d *ClaudeCodeDetector) determineState(entry *ClaudeLogEntry) domain.AgentState {
    duration := time.Since(entry.Timestamp)

    // Handle zero timestamp (parsing failed) - still determine state from stop_reason
    if entry.Timestamp.IsZero() {
        duration = 0
    }

    switch {
    case entry.IsEndTurn():
        return domain.NewAgentState("Claude Code", domain.AgentWaitingForUser, duration, domain.ConfidenceCertain)
    case entry.IsToolUse():
        return domain.NewAgentState("Claude Code", domain.AgentWorking, duration, domain.ConfidenceCertain)
    default:
        return domain.NewAgentState("Claude Code", domain.AgentUnknown, duration, domain.ConfidenceUncertain)
    }
}
```

### Detection Logic Summary

| Condition | Agent State | Confidence |
|-----------|-------------|------------|
| `stop_reason: "end_turn"` | WaitingForUser | Certain |
| `stop_reason: "tool_use"` | Working | Certain |
| No assistant entries | Inactive | Certain |
| No Claude logs | Unknown | Uncertain |
| Unrecognized stop_reason (e.g., "max_tokens") | Unknown | Uncertain |

### Edge Cases

| Case | Input | Behavior |
|------|-------|----------|
| Claude Code not installed | `~/.claude/projects` doesn't exist | Return AgentUnknown, ConfidenceUncertain, nil error |
| No logs for this project | Project path doesn't match any Claude project | Return AgentUnknown, ConfidenceUncertain, nil error |
| Empty session directory | Logs dir exists, no *.jsonl files | Return AgentInactive, ConfidenceCertain |
| Session with only user messages | Session file has no assistant entries | Return AgentInactive, ConfidenceCertain |
| Zero timestamp | Timestamp parsing failed | Duration = 0, state determined by stop_reason |
| Unknown stop_reason | e.g., "max_tokens" or future values | Return AgentUnknown, ConfidenceUncertain |
| Context cancelled | Any point in detection | Return immediately with AgentUnknown, nil error |
| PathMatcher error | Unexpected filesystem error | Return error to caller |
| LogParser error | Permission denied, etc. | Return error to caller |

### Testing Strategy

```go
func TestName(t *testing.T) {
    d := NewClaudeCodeDetector()
    if got := d.Name(); got != "Claude Code" {
        t.Errorf("Name() = %q, want %q", got, "Claude Code")
    }
}

func TestDetect_EndTurn_WaitingForUser(t *testing.T) {
    // Create temp directory structure mimicking ~/.claude/projects/-path-/
    tmpDir := t.TempDir()
    projectPath := "/test/project"
    claudeDir := filepath.Join(tmpDir, ".claude", "projects", "-test-project")
    os.MkdirAll(claudeDir, 0755)

    // Write JSONL with end_turn entry
    entry := `{"type":"assistant","stop_reason":"end_turn","timestamp":"2026-01-16T12:00:00Z"}`
    os.WriteFile(filepath.Join(claudeDir, "session.jsonl"), []byte(entry), 0644)

    // Create detector with custom path matcher that returns our temp dir
    // (Since PathMatcher uses real home dir, override for testing)
    d := NewClaudeCodeDetector()
    state, err := d.Detect(context.Background(), projectPath)

    // Note: This test requires HOME env manipulation or custom PathMatcher
    // See integration test pattern below
}

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

func TestDetect_Duration(t *testing.T) {
    // Setup: Create entry with known timestamp
    // Verify: time.Since(timestamp) is approximately correct
    // Use tolerance of 1 second for test timing variance
}

// Integration-style test with real filesystem
func TestDetect_Integration(t *testing.T) {
    // 1. Create temp dir structure
    // 2. Set HOME env to temp dir (restore after)
    // 3. Create Claude-like project structure
    // 4. Run detector
    // 5. Verify results
}

// Interface compliance test (compile-time)
var _ ports.AgentActivityDetector = (*ClaudeCodeDetector)(nil)
```

### Hexagonal Architecture

```
internal/adapters/agentdetectors/
├── doc.go                             # Package documentation (update for 15.4)
├── claude_code_path_matcher.go        # Path matching (Story 15.2)
├── claude_code_path_matcher_test.go   # Tests (Story 15.2)
├── claude_code_log_parser.go          # JSONL parsing (Story 15.3)
├── claude_code_log_parser_test.go     # Tests (Story 15.3)
├── claude_code_detector.go            # NEW: Detector implementation (this story)
└── claude_code_detector_test.go       # NEW: Tests (this story)
```

**Architecture Position:**
- ClaudeCodeDetector is an ADAPTER implementing `ports.AgentActivityDetector` interface
- Composes PathMatcher (15.2) and LogParser (15.3) as dependencies
- The interface is defined in `internal/core/ports/` (hexagonal boundary)

### Previous Story Learnings

From Story 15.1:
- Use `domain.ConfidenceCertain` for high confidence (not ConfidenceHigh - it doesn't exist)
- Use `domain.ConfidenceUncertain` for uncertain cases
- `domain.NewAgentState()` constructor takes (tool, status, duration, confidence)

From Story 15.2:
- PathMatcher.Match() returns `("", nil)` for graceful "not found" cases
- PathMatcher.Match() returns `("", err)` for unexpected filesystem errors
- Context cancellation returns empty result, not error
- Cache improves performance for repeated calls

From Story 15.3:
- LogParser.FindMostRecentSession() returns `("", nil)` if no sessions
- LogParser.ParseLastAssistantEntry() returns `(nil, nil)` if no assistant entries
- ClaudeLogEntry has IsEndTurn() and IsToolUse() helper methods
- Timestamp may be zero if parsing failed (handle gracefully)

### Integration with Future Stories

Story 15.5 (Generic Detector) will implement the same `AgentActivityDetector` interface with file-activity fallback:
```go
type GenericDetector struct {
    threshold time.Duration // Default 10 minutes
}

func (d *GenericDetector) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
    // Check file modification time
    // If no activity for > threshold → WaitingForUser (ConfidenceUncertain)
    // Otherwise → Working (ConfidenceUncertain)
}
```

Story 15.6 (TUI Integration) will use both detectors:
```go
// In dashboard:
state, _ := claudeDetector.Detect(ctx, project.Path)
if state.IsUnknown() {
    state, _ = genericDetector.Detect(ctx, project.Path)
}
// Display state in project row
```

### doc.go Update

Add to existing doc.go:
```go
// ClaudeCodeDetector is the main implementation of AgentActivityDetector
// for detecting Claude Code agent state. It composes ClaudeCodePathMatcher
// and ClaudeCodeLogParser to provide high-confidence detection by parsing
// Claude Code's JSONL log files.
```

### References

- [Source: docs/epics-phase2.md#Story-3.4] - Story acceptance criteria
- [Source: internal/core/ports/agent_activity_detector.go] - Interface definition (Story 15.1)
- [Source: internal/adapters/agentdetectors/claude_code_path_matcher.go] - PathMatcher (Story 15.2)
- [Source: internal/adapters/agentdetectors/claude_code_log_parser.go] - LogParser (Story 15.3)
- [Source: internal/core/domain/agent_state.go] - AgentState struct
- [Source: internal/core/domain/agent_status.go] - AgentStatus enum
- [Source: internal/core/domain/confidence.go] - Confidence enum (ConfidenceCertain, ConfidenceUncertain)

## Dev Agent Record

### Context Reference

- Phase 2 Epic 15: Sub-1-Minute Agent Detection (THE killer feature)
- FR Coverage: FR-P2-1 (JSONL parsing), FR-P2-2 (Working/Waiting/Inactive states), FR-P2-3 (elapsed time), FR-P2-4 (confidence level)
- Prerequisite: Story 15.1 (AgentActivityDetector interface) - DONE
- Prerequisite: Story 15.2 (ClaudeCodePathMatcher) - DONE
- Prerequisite: Story 15.3 (ClaudeCodeLogParser) - DONE

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None

### Completion Notes List

- Implemented ClaudeCodeDetector struct with functional options pattern (WithPathMatcher, WithLogParser)
- Detector composes ClaudeCodePathMatcher and ClaudeCodeLogParser from Stories 15.2 and 15.3
- Detection flow: PathMatcher.Match() → LogParser.FindMostRecentSession() → LogParser.ParseLastAssistantEntry() → determineState()
- All 8 acceptance criteria covered with 18 unit tests (14 test functions + 5 table-driven subtests)
- Tests use HOME env manipulation with t.TempDir() for real filesystem integration testing
- Context cancellation returns within 100ms (verified by TestDetect_ContextCancelled_ReturnsPromptly)
- All 1247 project tests pass, lint clean

### Code Review Fixes Applied (2026-01-16)

Adversarial code review identified 8 issues (3 HIGH, 3 MEDIUM, 2 LOW). All HIGH and MEDIUM issues fixed:

**HIGH issues fixed:**
- H1: Added context cancellation check between PathMatcher.Match() and LogParser.FindMostRecentSession()
- H2: Added context cancellation check between FindMostRecentSession() and ParseLastAssistantEntry()
- H3: Removed redundant interface compliance assertion from test file (kept production code assertion)

**MEDIUM issues fixed:**
- M2: Added negative duration clamping for future timestamps (clock skew handling)
- M3: Extracted `"Claude Code"` to `const detectorName` to avoid repeated string literals

**Tests added:**
- TestDetect_FutureTimestamp_DurationZero: Verifies future timestamps clamp duration to 0
- Added table-driven test case for future timestamp in TestDetermineState_TableDriven

All 1248 project tests pass, lint clean after fixes.

### File List

- `internal/adapters/agentdetectors/claude_code_detector.go` (new, updated with review fixes)
- `internal/adapters/agentdetectors/claude_code_detector_test.go` (new, updated with review fixes)
- `internal/adapters/agentdetectors/doc.go` (updated)

