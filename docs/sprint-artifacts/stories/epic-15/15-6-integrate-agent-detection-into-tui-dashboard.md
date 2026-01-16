# Story 15.6: Integrate Agent Detection into TUI Dashboard

Status: done

## Story

As a user,
I want to see agent status for each project in the dashboard,
So that I know which agents need my attention immediately.

## User-Visible Changes

- **New:** Dashboard project rows show "⏸️ WAITING 2h" with bold red styling when an agent is waiting for user input (Claude Code detection or file-activity fallback)
- **New:** Dashboard status bar shows "⏸️ N WAITING" in red when any project has an agent waiting
- **Changed:** WAITING detection now uses Claude Code log parsing (< 1 second, high confidence) instead of purely file-activity heuristics (10-minute threshold, low confidence)
- **Changed:** Working agents show "🔄 Working" or no special indicator (normal state)

## Acceptance Criteria

1. **AC1:** Given project has Claude Code with state `WaitingForUser` for 2 hours, dashboard shows "WAITING 2h" in status column with bold red styling
2. **AC2:** Given project has agent state `Working` (actively using tools), dashboard shows no WAITING indicator (normal state)
3. **AC3:** Given project has agent state `Inactive` (no recent sessions), dashboard shows no WAITING indicator
4. **AC4:** Given project has no Claude Code logs, fallback to GenericDetector (file-activity based, 10-minute threshold)
5. **AC5:** Given status bar counts, when any project is WaitingForUser, status bar shows "N WAITING" in red (existing behavior continues to work)
6. **AC6:** Detection completes within 1 second per project (NFR-P2-1)
7. **AC7:** Agent detection respects context cancellation, returning within 100ms when cancelled
8. **AC8:** Agent detection runs during dashboard refresh and periodic stage re-detection

## Tasks / Subtasks

- [x] Task 1: Create AgentDetectionService in adapters layer (AC: 4, 6, 7)
  - [x] 1.1: Create `internal/adapters/detection/agent_detection_service.go` (NOTE: adapters, not services - avoids hexagonal violation)
  - [x] 1.2: Define struct with `claudeDetector ports.AgentActivityDetector` and `genericDetector ports.AgentActivityDetector` fields (use interface for DI)
  - [x] 1.3: Implement `NewAgentDetectionService(opts ...Option)` constructor with functional options
  - [x] 1.4: Add `WithClaudeDetector(ports.AgentActivityDetector)` option for dependency injection
  - [x] 1.5: Add `WithGenericDetector(ports.AgentActivityDetector)` option for dependency injection
  - [x] 1.6: Default constructor creates both detectors (ClaudeCodeDetector, GenericDetector) if not provided

- [x] Task 2: Implement Detect() method with fallback logic (AC: 4, 6, 7)
  - [x] 2.1: Create `Detect(ctx context.Context, projectPath string) (domain.AgentState, error)` method
  - [x] 2.2: Set 1-second timeout context for individual detector calls
  - [x] 2.3: Call `claudeDetector.Detect()` first
  - [x] 2.4: If Claude returns AgentUnknown (no logs), fall back to `genericDetector.Detect()`
  - [x] 2.5: Log detection source ("Claude Code" vs "Generic") at debug level
  - [x] 2.6: Return error only for unexpected failures (context cancel, permission errors)

- [x] Task 3: Implement AgentActivityDetector port interface (AC: 6, 7)
  - [x] 3.1: Add interface methods `Detect()` and `Name()` per ports.AgentActivityDetector
  - [x] 3.2: Name() returns "Agent Detection Service"
  - [x] 3.3: Add compile-time interface compliance check

- [x] Task 4: Implement WaitingDetector port adapter (AC: 5, backward compatibility)
  - [x] 4.1: Create `internal/adapters/detection/agent_waiting_adapter.go`
  - [x] 4.2: Wrap AgentDetectionService to satisfy existing `ports.WaitingDetector` interface
  - [x] 4.3: `IsWaiting()` returns true when AgentState.Status == AgentWaitingForUser
  - [x] 4.4: `WaitingDuration()` returns AgentState.Duration
  - [x] 4.5: Cache detection results per project for 5 seconds to avoid repeated filesystem scans

- [x] Task 5: Verify TUI wiring works with new adapter (AC: 1, 2, 3, 8)
  - [x] 5.1: Verify `ports.WaitingDetector` interface compatibility (no code changes needed in TUI)
  - [x] 5.2: Verify `tui.Run()` accepts the new `AgentWaitingAdapter` (implements same interface)
  - [x] 5.3: Verify callbacks (`isProjectWaiting`, `getWaitingDuration`) work correctly
  - [x] 5.4: Verify hibernated project filtering still works
  - [x] 5.5: NOTE: No changes to model.go or app.go required - interface unchanged

- [x] Task 6: Update main.go and cli wiring (AC: 8) - REPLACES existing WaitingDetector
  - [x] 6.1: In `cmd/vdash/main.go`: Remove `services.NewWaitingDetector()` creation
  - [x] 6.2: In `cmd/vdash/main.go`: Create `detection.NewAgentDetectionService()` and `detection.NewAgentWaitingAdapter()`
  - [x] 6.3: In `cmd/vdash/main.go`: Pass adapter to `cli.SetWaitingDetector()` (existing function)
  - [x] 6.4: In `internal/adapters/tui/app.go`: No changes needed (uses existing waitingDetector parameter)
  - [x] 6.5: Handle nil case gracefully (detection disabled)

- [x] Task 7: Trigger agent detection during refresh (AC: 8)
  - [x] 7.1: Call agent detection during manual refresh (r key)
  - [x] 7.2: Call agent detection during periodic stage re-detection (Story 8.11)
  - [x] 7.3: Call agent detection when file watcher events trigger updates

- [x] Task 8: Write comprehensive unit tests (AC: 1-8)
  - [x] 8.1: Create `internal/adapters/detection/agent_detection_service_test.go`
  - [x] 8.2: Test Claude detection → WaitingForUser → returns correctly
  - [x] 8.3: Test Claude detection → Unknown → falls back to Generic
  - [x] 8.4: Test Generic detection → WaitingForUser → returns correctly
  - [x] 8.5: Test timeout after 1 second
  - [x] 8.6: Test context cancellation returns within 100ms
  - [x] 8.7: Create `internal/adapters/detection/agent_waiting_adapter_test.go`
  - [x] 8.8: Test IsWaiting() returns true for WaitingForUser status
  - [x] 8.9: Test WaitingDuration() returns correct duration
  - [x] 8.10: Test cache TTL (5 seconds)
  - [x] 8.11: Test context cancellation check between Claude and Generic detection calls

- [x] Task 9: Verify integration and compliance
  - [x] 9.1: Run `make lint && make test` - all must pass
  - [ ] 9.2: Verify existing WAITING display still works (backward compatibility) - USER TESTING
  - [ ] 9.3: Manual test: Run vdash with a Claude Code project, verify detection works - USER TESTING
  - [ ] 9.4: Manual test: Run vdash with a non-Claude-Code project, verify fallback works - USER TESTING

## Dev Notes

### Architecture Overview

```
internal/adapters/detection/
├── agent_detection_service.go    # NEW: Orchestrates Claude + Generic detectors
├── agent_detection_service_test.go
├── agent_waiting_adapter.go      # NEW: Adapts AgentDetectionService to WaitingDetector interface
├── agent_waiting_adapter_test.go

cmd/vdash/
├── main.go                       # UPDATE: Create AgentDetectionService, replace old WaitingDetector
```

**Why adapters, not services?** The AgentDetectionService directly composes adapters (ClaudeCodeDetector, GenericDetector). Per hexagonal architecture, core/services MUST NOT import adapters. Moving to adapters/detection resolves this.

### Detection Flow

```
1. TUI refresh/tick triggers detection for each project
2. AgentDetectionService.Detect(ctx, projectPath)
   ├── ClaudeCodeDetector.Detect()
   │   ├── PathMatcher.Match() → find Claude logs directory
   │   ├── LogParser.FindMostRecentSession() → find latest session
   │   ├── LogParser.ParseLastAssistantEntry() → get last assistant message
   │   └── determineState() → WaitingForUser / Working / Inactive / Unknown
   │
   └── If Unknown → GenericDetector.Detect()
       ├── findMostRecentModification() → scan project files
       └── Compare against 10-minute threshold → WaitingForUser / Working / Unknown
3. AgentWaitingAdapter wraps result for TUI components
4. TUI renders WAITING indicator based on IsWaiting() result
```

### Integration with Existing WaitingDetector (MIGRATION)

The existing `services.WaitingDetector` (Story 4.3/4.4) uses threshold-based detection from `Project.LastActivityAt`. This story **REPLACES** it with the new `AgentWaitingAdapter`.

**Why replace instead of wrap?**
- New detection is more accurate (log-based vs threshold)
- Threshold resolver becomes obsolete for agent detection
- GenericDetector internally uses 10-minute threshold as fallback
- Simplifies architecture (one detection path, not two)

**Changes required in `cmd/vdash/main.go`:**

```go
// BEFORE (lines ~154-172):
thresholdResolver := config.NewWaitingThresholdResolver(cfg, basePath, cli.GetWaitingThreshold)
waitingDetector := services.NewWaitingDetector(thresholdResolver)
cli.SetWaitingDetector(waitingDetector)

// AFTER:
agentService := detection.NewAgentDetectionService()
waitingDetector := detection.NewAgentWaitingAdapter(agentService)
cli.SetWaitingDetector(waitingDetector) // Same function, new implementation
```

**Impact:**
- `WaitingThresholdResolver` no longer needed (can be removed or kept for backward compat)
- CLI flag `--waiting-threshold` no longer affects detection (GenericDetector uses 10-min default)
- If per-project threshold customization is desired, add `WithThreshold` option to GenericDetector

### AgentDetectionService Implementation

```go
package detection

import (
    "context"
    "log/slog"
    "time"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/agentdetectors"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

const serviceName = "Agent Detection Service"

// detectionTimeout is the maximum time for a single detection.
// Per NFR-P2-1: Agent detection latency < 1 second.
const detectionTimeout = 1 * time.Second

// AgentDetectionService orchestrates multiple agent detectors with fallback.
// It tries Claude Code detection first (high confidence), then falls back
// to generic file-activity detection (low confidence).
//
// Located in adapters layer (not services) because it directly composes
// adapter-layer detectors. Per hexagonal architecture, core/services
// MUST NOT import adapters.
type AgentDetectionService struct {
    claudeDetector  ports.AgentActivityDetector // Use interface for DI/testing
    genericDetector ports.AgentActivityDetector
}

// ServiceOption configures AgentDetectionService.
type ServiceOption func(*AgentDetectionService)

// WithClaudeDetector sets a custom Claude detector (for testing).
func WithClaudeDetector(d ports.AgentActivityDetector) ServiceOption {
    return func(s *AgentDetectionService) {
        s.claudeDetector = d
    }
}

// WithGenericDetector sets a custom generic detector (for testing).
func WithGenericDetector(d ports.AgentActivityDetector) ServiceOption {
    return func(s *AgentDetectionService) {
        s.genericDetector = d
    }
}

// NewAgentDetectionService creates a new service with optional configuration.
func NewAgentDetectionService(opts ...ServiceOption) *AgentDetectionService {
    s := &AgentDetectionService{}
    for _, opt := range opts {
        opt(s)
    }
    if s.claudeDetector == nil {
        s.claudeDetector = agentdetectors.NewClaudeCodeDetector()
    }
    if s.genericDetector == nil {
        s.genericDetector = agentdetectors.NewGenericDetector()
    }
    return s
}

// Compile-time interface compliance check.
var _ ports.AgentActivityDetector = (*AgentDetectionService)(nil)

// Name returns the service identifier.
func (s *AgentDetectionService) Name() string {
    return serviceName
}

// Detect determines the agent activity state for a project.
// Uses Claude Code detection with fallback to generic file-activity detection.
func (s *AgentDetectionService) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
    // Apply timeout per NFR-P2-1 (< 1 second)
    ctx, cancel := context.WithTimeout(ctx, detectionTimeout)
    defer cancel()

    // Respect context cancellation at entry
    select {
    case <-ctx.Done():
        return domain.NewAgentState(serviceName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
    default:
    }

    // Step 1: Try Claude Code detection (high confidence)
    claudeState, err := s.claudeDetector.Detect(ctx, projectPath)
    if err != nil {
        slog.Debug("Claude Code detection error, falling back to generic",
            "path", projectPath, "error", err)
        // Fall through to generic detection
    } else if !claudeState.IsUnknown() {
        // Claude detection succeeded with known state
        slog.Debug("Agent detection via Claude Code",
            "path", projectPath, "status", claudeState.Status.String())
        return claudeState, nil
    }

    // Check context between detector calls (Story 15.4 learning)
    select {
    case <-ctx.Done():
        return domain.NewAgentState(serviceName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
    default:
    }

    // Step 2: Fall back to generic file-activity detection (low confidence)
    genericState, err := s.genericDetector.Detect(ctx, projectPath)
    if err != nil {
        slog.Debug("Generic detection error",
            "path", projectPath, "error", err)
        return domain.NewAgentState(serviceName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), err
    }

    slog.Debug("Agent detection via generic file activity",
        "path", projectPath, "status", genericState.Status.String())
    return genericState, nil
}
```

### AgentWaitingAdapter Implementation

```go
package detection

import (
    "context"
    "sync"
    "time"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// cacheTTL is the time-to-live for cached detection results.
// Prevents repeated filesystem scans during rapid TUI updates.
const cacheTTL = 5 * time.Second

// cacheEntry holds a cached detection result.
type cacheEntry struct {
    state     domain.AgentState
    timestamp time.Time
}

// AgentWaitingAdapter adapts AgentDetectionService to the existing WaitingDetector interface.
// This enables backward compatibility with TUI components that expect WaitingDetector.
type AgentWaitingAdapter struct {
    service *AgentDetectionService // Same package, no import needed
    cache   map[string]cacheEntry  // projectPath → cached state
    mu      sync.RWMutex
    now     func() time.Time // For testing
}

// NewAgentWaitingAdapter creates a new adapter wrapping the detection service.
func NewAgentWaitingAdapter(service *AgentDetectionService) *AgentWaitingAdapter {
    return &AgentWaitingAdapter{
        service: service,
        cache:   make(map[string]cacheEntry),
        now:     time.Now,
    }
}

// Compile-time interface compliance check.
var _ ports.WaitingDetector = (*AgentWaitingAdapter)(nil)

// IsWaiting returns true if the project's agent is waiting for user input.
func (a *AgentWaitingAdapter) IsWaiting(ctx context.Context, project *domain.Project) bool {
    if project == nil || project.State == domain.StateHibernated {
        return false
    }

    state := a.detectWithCache(ctx, project.Path)
    return state.IsWaiting()
}

// WaitingDuration returns how long the project has been waiting.
func (a *AgentWaitingAdapter) WaitingDuration(ctx context.Context, project *domain.Project) time.Duration {
    if project == nil || project.State == domain.StateHibernated {
        return 0
    }

    state := a.detectWithCache(ctx, project.Path)
    if !state.IsWaiting() {
        return 0
    }
    return state.Duration
}

// detectWithCache returns cached result if fresh, otherwise performs detection.
func (a *AgentWaitingAdapter) detectWithCache(ctx context.Context, projectPath string) domain.AgentState {
    // Check cache first
    a.mu.RLock()
    if entry, ok := a.cache[projectPath]; ok {
        if a.now().Sub(entry.timestamp) < cacheTTL {
            a.mu.RUnlock()
            return entry.state
        }
    }
    a.mu.RUnlock()

    // Cache miss or stale - perform detection
    state, _ := a.service.Detect(ctx, projectPath)

    // Update cache
    a.mu.Lock()
    a.cache[projectPath] = cacheEntry{
        state:     state,
        timestamp: a.now(),
    }
    a.mu.Unlock()

    return state
}

// ClearCache clears all cached entries (for testing).
func (a *AgentWaitingAdapter) ClearCache() {
    a.mu.Lock()
    a.cache = make(map[string]cacheEntry)
    a.mu.Unlock()
}
```

### TUI Wiring (NO CHANGES REQUIRED)

The TUI already uses `WaitingDetector` interface via callbacks. The new `AgentWaitingAdapter` implements `ports.WaitingDetector`, so it's a drop-in replacement:

```go
// In internal/adapters/tui/model.go - NO CHANGES NEEDED
// Existing code already works:
// - m.waitingDetector is type ports.WaitingDetector (interface)
// - SetWaitingDetector() accepts any ports.WaitingDetector
// - Callbacks use m.isProjectWaiting() and m.getWaitingDuration() wrappers

// The only change is in main.go (Task 6) which creates the new detector
```

### main.go Wiring (THE KEY CHANGE)

```go
// In cmd/vdash/main.go - replace existing WaitingDetector creation:

import (
    "github.com/JeiKeiLim/vibe-dash/internal/adapters/detection"
    // ... existing imports
)

func main() {
    // ... existing setup code ...

    // REMOVE these lines (old threshold-based detection):
    // thresholdResolver := config.NewWaitingThresholdResolver(cfg, basePath, cli.GetWaitingThreshold)
    // waitingDetector := services.NewWaitingDetector(thresholdResolver)

    // ADD these lines (new agent-based detection):
    agentService := detection.NewAgentDetectionService()
    waitingDetector := detection.NewAgentWaitingAdapter(agentService)

    // This line stays the same:
    cli.SetWaitingDetector(waitingDetector)

    // ... rest of main() unchanged ...
}
```

### Display Integration

The existing TUI components already support WAITING display via WaitingChecker/WaitingDurationGetter callbacks:

1. **delegate.go:waitingIndicator()** - Already renders "WAITING Xh Ym" in project rows
2. **detail_panel.go** - Already shows "Waiting: [emoji] Xh Ym" when waiting
3. **status_bar.go** - Already shows "N WAITING" count

This story only needs to:
1. Replace the old WaitingDetector with the new AgentWaitingAdapter
2. The display logic remains unchanged (backward compatible)

### Detection Confidence Handling

| Detector | Confidence | Display Impact |
|----------|------------|----------------|
| Claude Code | Certain | Normal WAITING styling |
| Generic | Uncertain | Same WAITING styling (Story 15.7 adds confidence to detail panel) |

Story 15.7 (next story) will add confidence level display in detail panel. This story focuses on wiring and basic functionality.

### Testing Strategy

1. **Unit tests** for AgentDetectionService:
   - Claude detection success → returns Claude result
   - Claude returns Unknown → falls back to Generic
   - Both detectors fail → returns Unknown
   - Timeout after 1 second → returns Unknown
   - Context cancellation → returns within 100ms

2. **Unit tests** for AgentWaitingAdapter:
   - IsWaiting returns true for WaitingForUser
   - IsWaiting returns false for Working/Inactive/Unknown
   - WaitingDuration returns correct value
   - Cache prevents repeated detections within 5 seconds
   - Cache expires after 5 seconds

3. **Integration tests** (manual):
   - Run vdash on a project with Claude Code logs → verify detection works
   - Run vdash on a project without Claude Code → verify fallback works
   - Toggle Claude session waiting/working → verify dashboard updates

### Hexagonal Architecture

```
internal/adapters/detection/        # NEW PACKAGE
├── agent_detection_service.go      # Orchestrates detectors (adapters layer)
├── agent_detection_service_test.go
├── agent_waiting_adapter.go        # Implements ports.WaitingDetector
├── agent_waiting_adapter_test.go

internal/adapters/agentdetectors/   # EXISTING (Stories 15.2-15.5)
├── claude_code_detector.go         # Implements ports.AgentActivityDetector
├── generic_detector.go             # Implements ports.AgentActivityDetector
└── ... (other files)

cmd/vdash/
├── main.go                         # UPDATE: Create detection service

internal/adapters/tui/              # NO CHANGES (uses interface)
├── model.go                        # Uses ports.WaitingDetector (already)
└── app.go                          # Receives WaitingDetector via Run()
```

**Architecture Notes:**
- AgentDetectionService is in **adapters layer** (not services) because it composes adapters
- Per hexagonal architecture: core/services MUST NOT import adapters
- AgentDetectionService uses `ports.AgentActivityDetector` interface for DI
- AgentWaitingAdapter implements `ports.WaitingDetector` for TUI compatibility
- TUI code needs NO changes - it already uses the interface

### Previous Story Learnings

From Story 15.3 (ClaudeCodeLogParser):
- Tail-optimized reading (last 4KB) for performance on large log files
- `ParseLastAssistantEntry()` returns `(nil, nil)` when no assistant entries found
- Context cancellation checked every 100 entries during parsing
- JSONL parsing skips malformed lines gracefully (continues to next line)
- Session files sorted by modification time to find most recent

From Story 15.4 (ClaudeCodeDetector):
- Use `const detectorName` to avoid repeated string literals
- Clamp negative durations to 0 for future timestamps (clock skew)
- Context cancellation check between steps for responsiveness
- `AgentInactive` returned when logs exist but no assistant messages
- Error propagation: PathMatcher errors propagate, parse errors → Unknown

From Story 15.5 (GenericDetector):
- Default threshold is 10 minutes (configurable via `WithThreshold`)
- All states return `ConfidenceUncertain` (heuristic-based)
- Skip hidden files and directories (starting with `.`)
- Context cancellation checked every 100 files during filesystem walk
- Permission errors logged at debug level, scanning continues

From Story 4.5 (existing WAITING implementation):
- TUI components use WaitingChecker/WaitingDurationGetter callbacks
- `SetDelegateWaitingCallbacks()` wires detection into project list
- Hibernated projects never show as waiting (`state == StateHibernated` → false)

### Edge Cases

| Case | Behavior |
|------|----------|
| Claude Code not installed | Falls back to Generic detector |
| No Claude logs for project | Falls back to Generic detector |
| Detection takes > 1 second | Timeout, returns Unknown |
| Context cancelled during detection | Returns Unknown within 100ms |
| Hibernated project | Never shows WAITING (existing behavior) |
| New project (just added) | Detection runs on next tick |
| Nil project | Returns false for IsWaiting (existing behavior) |

### Performance Considerations

1. **Detection Timeout:** 1 second per project max (NFR-P2-1)
2. **Cache TTL:** 5 seconds prevents repeated filesystem scans
3. **Parallel Detection:** Not implemented in this story (future optimization)
4. **Lazy Detection:** Only detects when TUI renders, not on every tick

### References

- [Source: docs/epics-phase2.md#Story-3.6] - Story acceptance criteria
- [Source: docs/project-context.md#Phase-2-Additions] - Agent detection architecture
- [Source: docs/prd-phase2.md#Agent-Detection] - NFR-P2-1 (< 1 second latency)
- [Source: internal/core/ports/agent_activity_detector.go] - Interface definition (Story 15.1)
- [Source: internal/adapters/agentdetectors/claude_code_detector.go] - ClaudeCodeDetector (Story 15.4)
- [Source: internal/adapters/agentdetectors/generic_detector.go] - GenericDetector (Story 15.5)
- [Source: internal/core/ports/waiting_detector.go] - WaitingDetector interface (existing)
- [Source: internal/adapters/tui/components/delegate.go] - Project row WAITING display
- [Source: internal/adapters/tui/components/detail_panel.go] - Detail panel WAITING display
- [Source: internal/adapters/tui/components/status_bar.go] - Status bar WAITING count
- [Source: docs/sprint-artifacts/stories/epic-15/15-3-implement-claude-code-jsonl-log-parser.md] - Previous story learnings (tail optimization, context cancellation)
- [Source: docs/sprint-artifacts/stories/epic-15/15-4-implement-agent-state-detection-logic.md] - Previous story learnings (context checks, duration clamping)
- [Source: docs/sprint-artifacts/stories/epic-15/15-5-implement-generic-file-activity-fallback-detector.md] - Previous story learnings (filesystem walk, hidden file skip)
- [Source: cmd/vdash/main.go] - Existing WaitingDetector wiring (lines ~154-172)

## Dev Agent Record

### Context Reference

- Phase 2 Epic 15: Sub-1-Minute Agent Detection (THE killer feature)
- FR Coverage: FR-P2-1 (JSONL parsing), FR-P2-2 (agent states), FR-P2-3 (elapsed time), FR-P2-5 (fallback)
- Prerequisite: Story 15.1 (AgentActivityDetector interface) - DONE
- Prerequisite: Story 15.2 (ClaudeCodePathMatcher) - DONE
- Prerequisite: Story 15.3 (ClaudeCodeLogParser) - DONE
- Prerequisite: Story 15.4 (ClaudeCodeDetector) - DONE
- Prerequisite: Story 15.5 (GenericDetector) - DONE

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None

### Completion Notes List

1. **Architecture**: Created `internal/adapters/detection/` package for agent detection orchestration. Per hexagonal architecture, this is in adapters layer (not services) because it directly composes adapter-layer detectors.

2. **AgentDetectionService**: Orchestrates Claude Code and Generic detectors with fallback logic. Tries Claude Code first (high confidence), falls back to Generic (file-activity based) when Claude returns Unknown or errors.

3. **AgentWaitingAdapter**: Wraps AgentDetectionService to implement `ports.WaitingDetector` interface. This enables drop-in replacement without TUI changes. Includes 5-second cache TTL to prevent repeated filesystem scans during rapid TUI updates.

4. **main.go Changes**: Replaced old threshold-based WaitingDetector (`services.NewWaitingDetector(thresholdResolver)`) with new agent-based detection (`detection.NewAgentDetectionService()` + `detection.NewAgentWaitingAdapter()`). The old WaitingThresholdResolver is no longer used.

5. **TUI Wiring**: No changes needed - TUI uses `ports.WaitingDetector` interface via callbacks (`isProjectWaiting`, `getWaitingDuration`). The existing wiring in model.go already triggers detection during tick, manual refresh ('r' key), and file watcher events.

6. **Test Coverage**: 39 new tests covering detection service fallback logic, timeout (1 second), context cancellation (< 100ms), cache behavior (5-second TTL), and interface compliance. All 1283 tests pass.

7. **Impact of Removal**: The `WaitingThresholdResolver` and `services.NewWaitingDetector()` are no longer used. The `--waiting-threshold` CLI flag no longer affects detection (GenericDetector uses 10-minute default). Per-project threshold customization could be added via `WithThreshold` option to GenericDetector if needed.

### File List

- `internal/adapters/detection/agent_detection_service.go` (NEW - orchestrates detectors)
- `internal/adapters/detection/agent_detection_service_test.go` (NEW - 13 tests)
- `internal/adapters/detection/agent_waiting_adapter.go` (NEW - implements ports.WaitingDetector)
- `internal/adapters/detection/agent_waiting_adapter_test.go` (NEW - 26 tests)
- `cmd/vdash/main.go` (MODIFIED - replace old WaitingDetector with new AgentWaitingAdapter)

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Build and Run

```bash
make build
./bin/vdash
```

### Step 2: Verify Dashboard Displays

| Scenario | Expected Behavior | How to Test |
|----------|-------------------|-------------|
| Project with Claude Code logs | Shows "WAITING Xh" or normal state based on log parsing | Add a project where you've used Claude Code |
| Project without Claude Code | Falls back to file-activity detection (10-min threshold) | Add a project without `.claude/` logs |
| Hibernated project | Never shows WAITING | Hibernate a project, verify no WAITING indicator |
| Status bar count | Shows "N WAITING" in red when any project is waiting | Check status bar at bottom |

### Step 3: Verify Detection Updates

1. Press `r` key - status should refresh
2. Wait for periodic tick (~1 second) - counts should update if state changed
3. Modify a file in tracked project - file watcher should trigger update

### Decision Guide

| Situation | Action |
|-----------|--------|
| WAITING indicator shows for projects with Claude Code in waiting state | PASS |
| WAITING indicator shows for projects after 10+ minutes of inactivity (fallback) | PASS |
| Hibernated projects never show WAITING | PASS |
| Status bar "N WAITING" count is accurate | PASS |
| Detection completes within 1 second (no noticeable lag) | PASS |
| Any check fails | Do NOT approve, document issue |
