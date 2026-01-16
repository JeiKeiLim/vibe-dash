# Story 15.7: Display Confidence Level in Detail Panel

Status: review

## Story

As a user,
I want to see how confident the agent detection is,
So that I know whether to trust log-based vs heuristic detection.

## User-Visible Changes

- **New:** Detail panel agent status shows confidence level and detection source (e.g., "Agent: ⏸️ WAITING 2h (High confidence - Claude Code logs)")
- **Changed:** When using fallback detection, shows "(Low confidence - file activity)" to inform user the detection is heuristic-based
- **Changed:** Confidence level styling differentiates High confidence (normal) from Low confidence (dim text via `styles.DimStyle`)

## Acceptance Criteria

1. **AC1:** Given Claude Code log-based detection, when detail panel shows agent status, displays "Agent: ⏸️ WAITING 2h (High confidence - Claude Code logs)"
2. **AC2:** Given generic file-based detection, when detail panel shows agent status, displays "Agent: ⏸️ WAITING 15m (Low confidence - file activity)"
3. **AC3:** Detail panel shows confidence level only when a project is selected AND agent is waiting
4. **AC4:** High confidence text displays in normal styling
5. **AC5:** Low confidence text displays with `styles.DimStyle` (faint text)
6. **AC6:** Working/Inactive/Unknown agent states do NOT display confidence in detail panel (only waiting status triggers display)
7. **AC7:** Detection source ("Claude Code logs" vs "file activity") is accurately displayed based on which detector provided the result
8. **AC8:** Existing tests continue to pass with new functionality

## Tasks / Subtasks

- [x] Task 1: Extend WaitingDetector interface to expose AgentState (AC: 1, 2, 7)
  - [x] 1.1: Add `AgentState(ctx context.Context, project *domain.Project) domain.AgentState` method to `ports.WaitingDetector` interface in `internal/core/ports/waiting_detector.go`
  - [x] 1.2: Implement `AgentState()` method in `AgentWaitingAdapter` - extract `project.Path` and call `detectWithCache(ctx, project.Path)`
  - [x] 1.3: Update compile-time interface check: `var _ ports.WaitingDetector = (*AgentWaitingAdapter)(nil)`
  - [x] 1.4: Add unit test for `AgentState()` method in `internal/adapters/detection/agent_waiting_adapter_test.go`

- [x] Task 2: Create callback type for AgentState retrieval (AC: 1, 2, 7)
  - [x] 2.1: Add `AgentStateGetter func(p *domain.Project) domain.AgentState` type to `internal/adapters/tui/components/delegate.go`
  - [x] 2.2: Document that callback captures model context via closure (consistent with existing WaitingChecker pattern)

- [x] Task 3: Wire AgentStateGetter callback through TUI model (AC: 1, 2, 7)
  - [x] 3.1: Add `agentStateGetter AgentStateGetter` field to `DetailPanelModel` struct in `internal/adapters/tui/components/detail_panel.go`
  - [x] 3.2: Add `SetAgentStateCallback(getter AgentStateGetter)` method to `DetailPanelModel`
  - [x] 3.3: Create callback wrapper in `internal/adapters/tui/model.go` using closure pattern: `func(p *domain.Project) domain.AgentState { return m.waitingDetector.AgentState(m.ctx, p) }`
  - [x] 3.4: Wire callback in `initComponents()` after model construction

- [x] Task 4: Update detail panel rendering to show confidence (AC: 1, 2, 3, 4, 5, 6, 7)
  - [x] 4.1: Add private helper functions to `internal/adapters/tui/components/detail_panel.go`: `confidenceToText(c domain.Confidence) string` and `toolToSourceText(tool string) string`
  - [x] 4.2: Modify `renderProject()` to check `m.agentStateGetter` - if nil, fall back to existing `m.waitingChecker` behavior without confidence display
  - [x] 4.3: When `agentStateGetter` returns waiting state, format as: "⏸️ Xh Ym (Confidence - Source)"
  - [x] 4.4: Apply `styles.DimStyle` for `ConfidenceUncertain` confidence (verified to exist at `internal/shared/styles/styles.go:92`)
  - [x] 4.5: Only display agent status row when `state.IsWaiting()` returns true

- [x] Task 5: Write unit tests for new functionality (AC: 8)
  - [x] 5.1: Add tests to `internal/adapters/detection/agent_waiting_adapter_test.go` for `AgentState()` method
  - [x] 5.2: Add tests to `internal/adapters/tui/components/detail_panel_test.go` for confidence text formatting
  - [x] 5.3: Test confidence text: ConfidenceCertain → "High confidence", ConfidenceUncertain → "Low confidence"
  - [x] 5.4: Test source text: "Claude Code" → "Claude Code logs", "Generic" → "file activity"
  - [x] 5.5: Test dim styling applied for ConfidenceUncertain
  - [x] 5.6: Test agent status row only renders when IsWaiting() is true
  - [x] 5.7: Test nil agentStateGetter fallback to existing waitingChecker behavior

- [x] Task 6: Verify integration and compliance
  - [x] 6.1: Run `make lint && make test` - all must pass
  - [x] 6.2: Manual test: Claude Code project → verify "High confidence - Claude Code logs" shows
  - [x] 6.3: Manual test: Non-Claude project → verify "Low confidence - file activity" shows
  - [x] 6.4: Manual test: Non-waiting project → verify no agent status row displays

## Dev Notes

### Architecture Overview

```
internal/core/ports/
├── waiting_detector.go            # UPDATE: Add AgentState() method to interface

internal/adapters/detection/
├── agent_waiting_adapter.go       # UPDATE: Implement AgentState() method
├── agent_waiting_adapter_test.go  # UPDATE: Add tests for AgentState()

internal/adapters/tui/components/
├── delegate.go                    # UPDATE: Add AgentStateGetter type
├── detail_panel.go                # UPDATE: Render confidence level with helpers
├── detail_panel_test.go           # UPDATE: Add confidence formatting tests

internal/adapters/tui/
├── model.go                       # UPDATE: Wire AgentStateGetter callback
```

### Required Imports

When implementing, ensure these imports are present in `detail_panel.go`:

```go
import (
    "fmt"
    "strings"
    "time"

    "github.com/charmbracelet/lipgloss"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/shared/emoji"
    "github.com/JeiKeiLim/vibe-dash/internal/shared/project"
    "github.com/JeiKeiLim/vibe-dash/internal/shared/styles"
    "github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)
```

### WaitingDetector Interface Extension

Add new method to existing interface:

```go
// internal/core/ports/waiting_detector.go
type WaitingDetector interface {
    IsWaiting(ctx context.Context, project *domain.Project) bool
    WaitingDuration(ctx context.Context, project *domain.Project) time.Duration
    // NEW - Story 15.7: Returns full AgentState for confidence/tool display
    AgentState(ctx context.Context, project *domain.Project) domain.AgentState
}
```

### AgentWaitingAdapter Implementation

```go
// internal/adapters/detection/agent_waiting_adapter.go

// AgentState returns the full agent detection state including confidence and tool.
// Story 15.7: Enables detail panel to display confidence level.
func (a *AgentWaitingAdapter) AgentState(ctx context.Context, project *domain.Project) domain.AgentState {
    if project == nil || project.State == domain.StateHibernated {
        return domain.NewAgentState("", domain.AgentUnknown, 0, domain.ConfidenceUncertain)
    }
    return a.detectWithCache(ctx, project.Path) // Uses existing cache
}
```

### Callback Type Definition

```go
// internal/adapters/tui/components/delegate.go

// AgentStateGetter returns the full agent detection state for a project.
// Story 15.7: Used by detail panel to display confidence level and detection source.
// Note: Does not take context parameter - caller captures context via closure.
type AgentStateGetter func(p *domain.Project) domain.AgentState
```

### Detail Panel Implementation

```go
// internal/adapters/tui/components/detail_panel.go

type DetailPanelModel struct {
    // ... existing fields ...
    agentStateGetter AgentStateGetter // Story 15.7: Full agent state for confidence display
}

// SetAgentStateCallback sets the agent state retrieval callback.
// Story 15.7: Enables confidence level display in detail panel.
func (m *DetailPanelModel) SetAgentStateCallback(getter AgentStateGetter) {
    m.agentStateGetter = getter
}

// confidenceToText converts Confidence enum to display text.
// Note: ConfidenceLikely included for future extensibility (current detectors only return Certain/Uncertain)
func confidenceToText(c domain.Confidence) string {
    switch c {
    case domain.ConfidenceCertain:
        return "High confidence"
    case domain.ConfidenceLikely:
        return "Medium confidence" // Future extensibility
    default:
        return "Low confidence"
    }
}

// toolToSourceText converts detector tool name to user-friendly source text.
func toolToSourceText(tool string) string {
    switch tool {
    case "Claude Code":
        return "Claude Code logs"
    case "Generic":
        return "file activity"
    default:
        return tool
    }
}

// formatAgentStatusWithConfidence formats agent status with confidence info.
// Only called when state.IsWaiting() is true.
func formatAgentStatusWithConfidence(state domain.AgentState) string {
    durationText := timeformat.FormatWaitingDuration(state.Duration, true)
    confidenceText := confidenceToText(state.Confidence)
    sourceText := toolToSourceText(state.Tool)

    // Format: "⏸️ Xh Ym (High confidence - Claude Code logs)"
    statusPart := fmt.Sprintf("%s %s", emoji.Waiting(), durationText)
    confPart := fmt.Sprintf("(%s - %s)", confidenceText, sourceText)

    // Apply dim styling for uncertain confidence
    if state.Confidence == domain.ConfidenceUncertain {
        confPart = styles.DimStyle.Render(confPart)
    }

    styledStatus := styles.WaitingStyle.Render(statusPart)
    return fmt.Sprintf("%s %s", styledStatus, confPart)
}
```

### Updated renderProject() in detail_panel.go

```go
// In renderProject() - replace existing waiting status section:

// Waiting status with confidence (Story 15.7)
if m.agentStateGetter != nil {
    state := m.agentStateGetter(p)
    if state.IsWaiting() {
        styledWaiting := formatAgentStatusWithConfidence(state)
        lines = append(lines, formatField("Waiting", styledWaiting))
    }
} else if m.waitingChecker != nil && m.waitingChecker(p) {
    // Fallback: existing behavior without confidence (backward compatibility)
    duration := time.Duration(0)
    if m.durationGetter != nil {
        duration = m.durationGetter(p)
    }
    waitingText := fmt.Sprintf("%s %s", emoji.Waiting(), timeformat.FormatWaitingDuration(duration, true))
    styledWaiting := styles.WaitingStyle.Render(waitingText)
    lines = append(lines, formatField("Waiting", styledWaiting))
}
```

### Wiring in model.go

```go
// internal/adapters/tui/model.go

// In initComponents() or after detailPanel construction:
if m.waitingDetector != nil {
    // Capture context in closure (consistent with existing WaitingChecker pattern)
    agentStateGetter := func(p *domain.Project) domain.AgentState {
        return m.waitingDetector.AgentState(m.ctx, p)
    }
    m.detailPanel.SetAgentStateCallback(agentStateGetter)
}
```

### Confidence Display Format

| Detector | Confidence Returned | Display Text |
|----------|---------------------|--------------|
| ClaudeCodeDetector | ConfidenceCertain | "High confidence - Claude Code logs" |
| GenericDetector | ConfidenceUncertain | "Low confidence - file activity" (dim styled) |

### Styling Reference

- `styles.WaitingStyle`: Bold red text for "⏸️ WAITING Xh Ym" (lines 63-65 of styles.go)
- `styles.DimStyle`: Faint text for low confidence info (lines 92-93 of styles.go)

### Edge Cases

| Case | Behavior |
|------|----------|
| Project not waiting | No agent status row displayed |
| nil project | Empty AgentState returned |
| Hibernated project | Empty AgentState returned |
| agentStateGetter is nil | Fall back to existing waitingChecker behavior (no confidence shown) |
| Unknown tool name | Display tool name directly in source text |

### Backward Compatibility

The new `AgentState()` method is **additive** - existing code using `IsWaiting()` and `WaitingDuration()` continues to work unchanged. Detail panel gracefully falls back when `agentStateGetter` is nil.

### Previous Story Learnings

**From Story 15.6 (AgentDetectionService):**
- `AgentWaitingAdapter.detectWithCache()` already caches full `AgentState`
- Cache TTL is 5 seconds to prevent repeated filesystem scans
- Hibernated projects return false for IsWaiting (same for AgentState)

**From Story 15.4 (ClaudeCodeDetector):**
- Returns `ConfidenceCertain` for all detected states (log-based = high confidence)
- Tool name is "Claude Code"

**From Story 15.5 (GenericDetector):**
- Returns `ConfidenceUncertain` for all states (heuristic-based = low confidence)
- Tool name is "Generic"

**Note on ConfidenceLikely:** The `ConfidenceLikely` case in `confidenceToText()` is included for future extensibility. Currently, no detector returns this value - ClaudeCodeDetector returns `ConfidenceCertain` and GenericDetector returns `ConfidenceUncertain`.

### References

- [Source: docs/epics-phase2.md#Story-3.7] - Story acceptance criteria
- [Source: docs/project-context.md#Phase-2-Additions] - Agent detection architecture
- [Source: docs/prd-phase2.md#Agent-Detection] - FR-P2-4 (confidence level)
- [Source: internal/core/domain/agent_state.go] - AgentState struct (Story 15.1)
- [Source: internal/core/domain/confidence.go] - Confidence enum
- [Source: internal/shared/styles/styles.go:92-93] - DimStyle definition
- [Source: internal/adapters/detection/agent_waiting_adapter.go] - Current adapter implementation (Story 15.6)
- [Source: internal/adapters/tui/components/detail_panel.go] - Detail panel rendering
- [Source: internal/adapters/tui/components/delegate.go] - Callback type definitions
- [Source: docs/sprint-artifacts/stories/epic-15/15-6-integrate-agent-detection-into-tui-dashboard.md] - Previous story learnings

## Dev Agent Record

### Context Reference

- Phase 2 Epic 15: Sub-1-Minute Agent Detection (THE killer feature)
- FR Coverage: FR-P2-4 (Show confidence level High/Low)
- Prerequisite: Story 15.1 (AgentActivityDetector interface) - DONE
- Prerequisite: Story 15.4 (ClaudeCodeDetector - provides ConfidenceCertain) - DONE
- Prerequisite: Story 15.5 (GenericDetector - provides ConfidenceUncertain) - DONE
- Prerequisite: Story 15.6 (TUI Dashboard Integration) - DONE

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None required - implementation followed story spec closely.

### Completion Notes List

1. Extended `ports.WaitingDetector` interface with `AgentState()` method for full detection state
2. Implemented `AgentState()` in `AgentWaitingAdapter` using existing `detectWithCache()` (no duplicate detection)
3. Added `AgentStateGetter` callback type in delegate.go following existing `WaitingChecker` pattern
4. Wired callback through TUI model with `getAgentState()` helper method
5. Updated `renderProject()` with:
   - `confidenceToText()`: Maps Confidence enum to "High/Medium/Low confidence"
   - `toolToSourceText()`: Maps tool names to "Claude Code logs" / "file activity"
   - `formatAgentStatusWithConfidence()`: Formats "⏸️ Xh Ym (confidence - source)"
   - DimStyle applied for ConfidenceUncertain
6. Also added `AgentState()` to legacy `services.WaitingDetector` for interface compliance
7. Updated mock implementations in test files for interface compliance
8. All 1292 tests pass, lint clean

### Code Review Fixes (2026-01-16)

Code review performed by Amelia (Dev Agent). Applied fixes:

1. **Enhanced mock implementations** - Added configurable `agentStateFunc` to `mockWaitingDetector` in model_test.go for better test control
2. **Added configurable AgentState** - MockWaitingDetector in cli/mocks_test.go now returns configurable agentState instead of empty
3. **Added missing tests for getAgentState() helper** - TestModel_GetAgentState_WithNilDetector and TestModel_GetAgentState_WithDetector
4. **Added missing AC6 tests** - TestDetailPanel_View_AgentStateGetter_Inactive and TestDetailPanel_View_AgentStateGetter_Unknown to fully cover AC6 (Working/Inactive/Unknown states don't show confidence)

All 1296 tests pass, lint clean (4 new tests added)

### File List

- `internal/core/ports/waiting_detector.go` - Added AgentState() method to interface
- `internal/adapters/detection/agent_waiting_adapter.go` - Implemented AgentState()
- `internal/adapters/detection/agent_waiting_adapter_test.go` - Added AgentState tests
- `internal/adapters/tui/components/delegate.go` - Added AgentStateGetter type
- `internal/adapters/tui/components/detail_panel.go` - Added confidence display logic
- `internal/adapters/tui/components/detail_panel_test.go` - Added confidence tests
- `internal/adapters/tui/model.go` - Added getAgentState() helper and wiring
- `internal/adapters/tui/model_test.go` - Updated mock for interface compliance
- `internal/adapters/cli/mocks_test.go` - Updated mock for interface compliance
- `internal/core/services/waiting_detector.go` - Added AgentState() for legacy service
