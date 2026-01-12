# Story 12.1: Claude Code Log Viewer

Status: complete

## Story

As a developer monitoring multiple AI-assisted projects,
I want to view Claude Code session logs directly from the vibe-dash TUI,
so that I can quickly check what Claude is doing without manually navigating to log files.

## User-Visible Changes

- **New:** Press `Enter` on a project to open full-screen Claude Code log view (if logs exist)
- **New:** Live-tailing log view with auto-scroll to latest entries
- **New:** `S` key in log view to switch between sessions
- **New:** `Shift+Enter` from project list to open session picker directly
- **New:** Brief feedback message "No Claude Code logs for this project" when logs don't exist

## Acceptance Criteria

1. **AC1:** When user presses `Enter` on a project in normal view mode, if Claude Code logs exist for that project, the TUI switches to a full-screen log view showing the latest session's log entries.

2. **AC2:** The log view displays raw JSON entries from the JSONL file, one entry per logical block, with timestamp and entry type visible.

3. **AC3:** The log view auto-scrolls to show the latest entry. When user manually scrolls up (using `k` or `up`), auto-scroll pauses. Pressing `G` resumes auto-scroll and jumps to latest.

4. **AC4:** Log entries update in near real-time (within 2 seconds) as Claude Code writes new entries to the session file.

5. **AC5:** Pressing `Esc` in log view returns to the project list (normal view mode).

6. **AC6:** Pressing `S` in log view shows a session picker listing all available sessions for the project, sorted by most recent first.

7. **AC7:** Session picker displays: session ID (truncated), start time, entry count, and summary (if available).

8. **AC8:** If no Claude Code logs directory exists for a project (`~/.claude/projects/{path-with-dashes}/`), pressing `Enter` shows a brief flash message "No Claude Code logs for this project" for 2 seconds.

9. **AC9:** `LogReader` port interface follows the same pattern as `MethodDetector` with `CanRead()` and adapter pattern for future extensibility.

10. **AC10:** Help overlay (`?`) is updated to show `Enter` key binding for "View logs" when in normal view.

## Tasks / Subtasks

- [x] Task 1: Create LogReader port interface (AC: 9)
  - [x] 1.1: Define `LogReader` interface in `internal/core/ports/log_reader.go`
  - [x] 1.2: Define `LogSession` and `LogEntry` domain types in `internal/core/domain/log_entry.go`
  - [x] 1.3: Define `LogReaderRegistry` interface in `internal/core/ports/log_reader.go`
  - [x] 1.4: Implement `LogReaderRegistry` in `internal/adapters/logreaders/registry.go`

- [x] Task 2: Implement ClaudeCodeReader adapter (AC: 1, 4, 6, 7)
  - [x] 2.1: Create `internal/adapters/logreaders/claude_code.go`
  - [x] 2.2: Implement `PathToClaudeDir()` function (project path to Claude directory mapping)
  - [x] 2.3: Implement `CanRead()` - check if Claude logs directory exists
  - [x] 2.4: Implement `ListSessions()` - list JSONL files sorted by mtime
  - [x] 2.5: Implement `ReadSession()` - parse JSONL with error handling (skip invalid lines)
  - [x] 2.6: Implement `TailSession()` - polling-based live tail (2s interval, context-cancellable)
  - [x] 2.7: Extract session metadata (summary, timestamps, entry count, skipped count)

- [x] Task 3: Create log view TUI mode (AC: 1, 2, 3, 5)
  - [x] 3.1: Add `viewModeLogs` to viewMode enum in `validation.go` (AFTER viewModeHibernated)
  - [x] 3.2: Add key constants to `keys.go`: `KeyLogSession = "S"`, `KeyLogJumpEnd = "G"`
  - [x] 3.3: Create `LogViewModel` component in `internal/adapters/tui/components/log_view.go`
  - [x] 3.4: Implement scrollable log entry display with raw JSON formatting
  - [x] 3.5: Implement auto-scroll behavior with pause on manual scroll
  - [x] 3.6: Handle `Enter` key in normal mode (check viewMode order: hibernated → logs → no-op)
  - [x] 3.7: Handle `Esc` to return to normal view (cancel tail context)
  - [x] 3.8: Handle `G` key to jump to latest and resume auto-scroll
  - [x] 3.9: Add `logTailTickMsg` (2s interval, separate from main tickMsg)
  - [x] 3.10: Add `logTailCtx`/`logTailCancel` fields to Model for goroutine cleanup

- [x] Task 4: Create session picker component (AC: 6, 7)
  - [x] 4.1: Create session picker overlay (follow `renderConfirmRemoveDialog` pattern)
  - [x] 4.2: Implement scrollable selection
  - [x] 4.3: Display: `{truncatedID} | {startTime} | {entryCount} entries`
  - [x] 4.4: Handle `S` key in log view to show picker
  - [x] 4.5: Handle session selection to switch log view to selected session

- [x] Task 5: Handle no-logs case gracefully (AC: 8)
  - [x] 5.1: Add `flashMessage string` and `flashMessageTime time.Time` to Model
  - [x] 5.2: Add `flashMsg` and `clearFlashMsg` message types
  - [x] 5.3: Show "No Claude Code logs for this project" on Enter when CanRead() is false
  - [x] 5.4: Use `tea.Tick(3*time.Second, ...)` → `clearFlashMsg{}`

- [x] Task 6: Update help overlay (AC: 10)
  - [x] 6.1: Add `Enter` → "View logs" to help overlay (views.go)

- [x] Task 7: Wire dependencies and integration
  - [x] 7.1: Add `SetLogReaderRegistry(registry ports.LogReaderRegistry)` to Model
  - [x] 7.2: Create `LogReaderRegistry` in main.go (after detectors.NewRegistry())
  - [x] 7.3: Register ClaudeCodeReader: `logRegistry.Register(logreaders.NewClaudeCodeReader())`
  - [x] 7.4: Add `cli.SetLogReaderRegistry(logRegistry)` in main.go
  - [x] 7.5: Update tui.Run to accept LogReaderRegistry parameter

## Dev Notes

### Architecture Pattern: LogReader and LogReaderRegistry

```go
// internal/core/ports/log_reader.go
type LogReader interface {
    // CanRead: true if logs directory exists for this project
    CanRead(ctx context.Context, projectPath string) bool

    // Tool: agentic tool name (e.g., "Claude Code", "Cursor")
    Tool() string

    // ListSessions: available sessions sorted by recency (newest first)
    ListSessions(ctx context.Context, projectPath string) ([]LogSession, error)

    // TailSession: streams new entries. Caller MUST cancel ctx when done.
    // Implementation MUST stop polling and close channel when ctx.Done().
    TailSession(ctx context.Context, sessionPath string) (<-chan LogEntry, error)
}

type LogReaderRegistry interface {
    // Register: add a log reader (called at startup)
    Register(reader LogReader)

    // GetReader: returns first reader where CanRead()==true, or nil
    GetReader(ctx context.Context, projectPath string) LogReader
}
```

### Claude Code Log Location

```
~/.claude/projects/{escaped-path}/

escaped-path = strings.ReplaceAll(projectPath, "/", "-")

Example: /Users/limjk/GitHub/JeiKeiLim/vibe-dash
      → ~/.claude/projects/-Users-limjk-GitHub-JeiKeiLim-vibe-dash/

Files: {uuid}.jsonl (main sessions), agent-{hash}.jsonl (sub-agents)
Entry Types: summary, user, assistant, system, file-history-snapshot
Key Fields: timestamp, type, message.content, sessionId
```

### Key Binding Constants

Add to `internal/adapters/tui/keys.go`:
```go
// Log View (Story 12.1)
KeyLogSession = "S"     // Session picker in log view
KeyLogJumpEnd = "G"     // Jump to end, resume auto-scroll
```

### View Mode Integration (CRITICAL)

Add `viewModeLogs` to enum in `internal/adapters/tui/validation.go` **AFTER** `viewModeHibernated`:
```go
const (
    viewModeNormal viewMode = iota
    viewModeValidation
    viewModeHibernated
    viewModeLogs  // Story 12.1: Must be LAST to preserve enum values
)
```

In `handleKeyMsg`, the `"enter"` case must check viewMode order:
```go
case "enter":
    // Story 11.4: Wake hibernated project (existing)
    if m.viewMode == viewModeHibernated && len(m.hibernatedProjects) > 0 {
        // ... existing code ...
        return m, nil
    }
    // Story 12.1: Open log view in normal mode
    if m.viewMode == viewModeNormal {
        return m.handleEnterForLogs()
    }
    return m, nil
```

### JSONL Error Handling

```go
// ReadSession implementation must handle:
// - Skip invalid JSON lines (log at Debug, continue)
// - Max line length: 1MB per entry (prevent OOM)
// - Return partial results on error (don't fail entire session)
// - Track skippedCount in LogSession for diagnostics

type LogSession struct {
    ID           string
    Path         string
    StartTime    time.Time
    EntryCount   int
    SkippedCount int  // Lines that failed to parse
    Summary      string
}
```

### Flash Message Implementation

Add to `internal/adapters/tui/model.go`:
```go
// Model struct additions:
flashMessage     string
flashMessageTime time.Time

// Message types:
type flashMsg struct{ text string }
type clearFlashMsg struct{}
```

In `View()`:
```go
// After ready check, before other overlays:
if m.flashMessage != "" && time.Since(m.flashMessageTime) < 2*time.Second {
    return renderFlashMessage(m.flashMessage, m.width, m.height)
}
```

In `views.go`, add `renderFlashMessage()` following `renderConfirmRemoveDialog` pattern (centered, minimal box).

### Live Tail Tick Integration

Use **separate** tick from main `tickMsg` (which is 5s for waiting detection):

```go
// Message type:
type logTailTickMsg time.Time

// Model additions:
logTailCtx    context.Context
logTailCancel context.CancelFunc

// Start tick only in viewModeLogs (2s interval for AC4):
func (m Model) logTailTickCmd() tea.Cmd {
    if m.viewMode != viewModeLogs {
        return nil
    }
    return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        return logTailTickMsg(t)
    })
}

// On Esc in log view - cancel tail context:
if m.logTailCancel != nil {
    m.logTailCancel()
}
m.viewMode = viewModeNormal
```

### Session Picker Pattern

Follow `renderConfirmRemoveDialog` pattern (views.go:212):
- Modal overlay, centered in terminal
- Use `list.Model` from bubbles for scrollable selection
- Max 10 sessions visible, scrollable if more
- Display format: `{truncatedID} | {startTime} | {entryCount} entries`

### Display Format (MVP - Raw JSON)

```
09:54:10  USER
────────────────────────────────────────
{"role":"user","content":"..."}

09:54:24  ASSISTANT
────────────────────────────────────────
{"model":"claude-opus-4-5-20251101","content":[{"type":"tool_use"...}]}
```

### Project Structure

| Location | Purpose |
|----------|---------|
| `internal/core/ports/log_reader.go` | LogReader, LogReaderRegistry interfaces |
| `internal/core/domain/log_entry.go` | LogSession, LogEntry types |
| `internal/adapters/logreaders/registry.go` | Registry implementation |
| `internal/adapters/logreaders/claude_code.go` | Claude Code adapter |
| `internal/adapters/tui/model.go` | viewModeTextView handling, flash messages, live tailing |
| `internal/adapters/tui/keys.go` | KeyLogSession, KeyLogJumpEnd, KeyShiftEnter constants |
| `internal/adapters/tui/views.go` | Help overlay with log view key bindings |

### Testing Approach

- Unit: Path mapping, JSONL parsing (with malformed input)
- Unit: LogReaderRegistry.GetReader iteration
- Integration: Sample JSONL fixtures in `test/fixtures/claude-logs/`
- TUI: View mode transitions, flash message timing

### References

- [Source: docs/project-context.md#Hexagonal-Architecture-Boundaries] - Core cannot import adapters
- [Source: docs/architecture.md#Plugin-Architecture] - Follow MethodDetector pattern
- [Source: internal/adapters/tui/validation.go:17-21] - viewMode enum pattern
- [Source: internal/adapters/detectors/registry.go:23-46] - Registry pattern
- [Source: internal/core/ports/detector.go:43-73] - MethodDetector interface
- [Source: internal/adapters/tui/views.go:212-255] - renderConfirmRemoveDialog pattern
- [Source: internal/adapters/tui/model.go:454-458] - tickCmd pattern

### Future Extensibility

The LogReader adapter pattern enables future support for Cursor, Aider, Windsurf, and other tools. Each implements the same interface with tool-specific path mapping and log parsing.

**Post-MVP opportunity:** Log-based agent waiting detection could use `AskUserQuestion` tool calls for certain waiting state vs current file-inactivity heuristics.

## Dev Agent Record

### Context Reference

Party mode requirements elicitation session (2026-01-12):
- User: Jongkuk Lim
- Agents consulted: Mary (Analyst), Winston (Architect), Sally (UX Designer), Murat (TEA)

### Agent Model Used

claude-opus-4-5-20251101

### Debug Log References

N/A - New story

### Completion Notes List

- Requirements gathered via party mode multi-agent discussion
- Architecture decision: Use adapter pattern matching MethodDetector
- UX decision: Full-screen log view (Option A), not panel-based
- Technical decision: Polling-based live tail (boring technology that works)
- Scope decision: Raw JSON display for MVP (pretty formatting deferred to user's CLI tool)

### File List

Files created:
- `internal/core/ports/log_reader.go` - LogReader, LogReaderRegistry interfaces
- `internal/core/domain/log_entry.go` - LogSession, LogEntry domain types
- `internal/core/domain/log_entry_test.go` - Unit tests for log entry types
- `internal/adapters/logreaders/registry.go` - LogReaderRegistry implementation
- `internal/adapters/logreaders/registry_test.go` - Registry unit tests
- `internal/adapters/logreaders/claude_code.go` - Claude Code adapter
- `internal/adapters/logreaders/claude_code_test.go` - Claude Code adapter tests
- `test/fixtures/claude-logs/valid-session.jsonl` - valid JSONL entries
- `test/fixtures/claude-logs/malformed-session.jsonl` - partial/corrupt entries

Files modified:
- `internal/adapters/tui/validation.go` - add viewModeTextView enum value
- `internal/adapters/tui/model.go` - log view handling, flash messages, live tailing
- `internal/adapters/tui/keys.go` - add KeyLogSession, KeyLogJumpEnd, KeyShiftEnter
- `internal/adapters/tui/views.go` - help overlay with log view key bindings
- `internal/adapters/cli/deps.go` - add SetLogReaderRegistry function
- `internal/adapters/cli/root.go` - pass logReaderRegistry to tui.Run
- `internal/adapters/tui/app.go` - accept logReaderRegistry parameter
- `cmd/vibe/main.go` - wire LogReaderRegistry

Implementation Notes:
- Log view embedded in model.go rather than separate component file (simpler integration)
- Uses viewModeTextView instead of viewModeLogs (more descriptive name)
- Live tailing implemented with 2s polling and auto-scroll
- jq fallback to raw display when not available
