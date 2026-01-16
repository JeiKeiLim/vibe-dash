# vdash Phase 2 - Epic Breakdown

**Author:** Jongkuk Lim
**Date:** 2026-01-16
**Baseline:** [Phase 1 Epics](./epics.md) - Phase 1 implementation complete

---

## Overview

This document provides the epic and story breakdown for **vdash Phase 2**, transforming the dashboard from a passive state tracker into an **active collaboration partner** with Claude Code.

**Phase 2 Focus:** "Where am I?" (solved in MVP) → "**What needs me now?**"

**Key Deliverables:**
- Sub-1-minute Agent Detection for Claude Code
- Methodology Switching Detection Fix
- Dynamic Binary Name Polish
- Progress Metrics & Stats View (Experimental)

---

## Functional Requirements Inventory (Phase 2)

### Agent Detection (FR-P2-1 to FR-P2-6)

| ID | Description | Priority |
|----|-------------|----------|
| FR-P2-1 | System can detect Claude Code tool usage via JSONL log parsing | Must Ship |
| FR-P2-2 | System can identify agent state as Working, WaitingForUser, or Inactive | Must Ship |
| FR-P2-3 | System can display elapsed time since agent entered current state | Must Ship |
| FR-P2-4 | System can show confidence level (High for log-based, Low for heuristic) | Must Ship |
| FR-P2-5 | System falls back to file-activity detection for non-Claude-Code projects | Must Ship |
| FR-P2-6 | System can detect Claude Code session by matching project path to `~/.claude/projects/` | Must Ship |

### Methodology Detection (FR-P2-7 to FR-P2-11)

| ID | Description | Priority |
|----|-------------|----------|
| FR-P2-7 | System can detect multiple methodologies in same project simultaneously | Must Ship |
| FR-P2-8 | System can compare artifact timestamps across detected methodologies | Must Ship |
| FR-P2-9 | System can select active methodology based on most recent artifact modification | Must Ship |
| FR-P2-10 | System can display coexistence warning when similar timestamps (within 1 hour) | Must Ship |
| FR-P2-11 | System displays both methodologies in TUI when tie-breaker applies | Must Ship |

### Progress Metrics (FR-P2-12 to FR-P2-18)

| ID | Description | Priority |
|----|-------------|----------|
| FR-P2-12 | System can record stage transition events with timestamps | Should Ship |
| FR-P2-13 | System can store metrics in separate database (`metrics.db`) | Should Ship |
| FR-P2-14 | Users can view Stats View via `'s'` key from Dashboard | Should Ship |
| FR-P2-15 | Users can exit Stats View via `Esc` or `'q'` key | Should Ship |
| FR-P2-16 | System can display activity sparklines per project | Should Ship |
| FR-P2-17 | System can display time-per-stage breakdown | Should Ship |
| FR-P2-18 | Users can select date range for metrics display | Should Ship |

### Polish (FR-P2-19 to FR-P2-20)

| ID | Description | Priority |
|----|-------------|----------|
| FR-P2-19 | System uses actual binary name (`os.Args[0]`) in error messages | Must Ship |
| FR-P2-20 | System uses actual binary name in version output | Must Ship |

---

## FR Coverage Map

| Epic | FRs Covered |
|------|-------------|
| Epic 1: Dynamic Binary Name Polish | FR-P2-19, FR-P2-20 |
| Epic 2: Methodology Switching Detection | FR-P2-7, FR-P2-8, FR-P2-9, FR-P2-10, FR-P2-11 |
| Epic 3: Sub-1-Minute Agent Detection | FR-P2-1, FR-P2-2, FR-P2-3, FR-P2-4, FR-P2-5, FR-P2-6 |
| Epic 4: Progress Metrics & Stats View | FR-P2-12, FR-P2-13, FR-P2-14, FR-P2-15, FR-P2-16, FR-P2-17, FR-P2-18 |

**Total: 20 FRs across 4 Epics**

---

## Epic 1: Dynamic Binary Name Polish

**User Value:** Users see consistent, accurate binary name in all CLI output regardless of how the binary is installed or renamed.

**FR Coverage:** FR-P2-19, FR-P2-20

**Technical Context:**
- Binary name currently hardcoded as "vdash" in version output and error messages
- Use `os.Args[0]` or `cobra.Command.Use` for dynamic resolution
- Help examples remain canonical "vdash" for documentation consistency

**Dependencies:** None (first epic, quick win)

---

### Story 1.1: Create Binary Name Resolution Utility

As a developer,
I want a utility function that resolves the actual binary name at runtime,
So that all CLI output consistently reflects how the user invoked the command.

**Acceptance Criteria:**

**Given** the binary is invoked as `./vdash`
**When** the utility is called
**Then** it returns "vdash"

**And** given the binary is invoked as `./my-custom-name`
**When** the utility is called
**Then** it returns "my-custom-name"

**And** given the binary is invoked via symlink `~/bin/v`
**When** the utility is called
**Then** it returns "v" (the symlink name)

**Technical Notes:**
- Create `internal/adapters/cli/binaryname.go`
- Use `filepath.Base(os.Args[0])` for resolution
- Handle edge case: empty Args[0] → fallback to "vdash"
- Export function: `func BinaryName() string`

**Prerequisites:** None

---

### Story 1.2: Update Version Command to Use Dynamic Name

As a user,
I want the version output to show the actual binary name I used,
So that copy-pasting version info is accurate for bug reports.

**Acceptance Criteria:**

**Given** I run `./vdash version`
**When** the version is displayed
**Then** output shows "vdash version X.Y.Z" (not hardcoded)

**And** given I run `./my-tool version` (renamed binary)
**When** the version is displayed
**Then** output shows "my-tool version X.Y.Z"

**Technical Notes:**
- Location: `internal/adapters/cli/version.go`
- Replace hardcoded "vdash" with `BinaryName()` call
- Version template: `%s version %s` where first %s is dynamic

**Prerequisites:** Story 1.1

---

### Story 1.3: Update Error Messages to Use Dynamic Name

As a user,
I want error messages to use the actual binary name I invoked,
So that suggested commands in errors are copy-pasteable.

**Acceptance Criteria:**

**Given** I run `./v add /nonexistent` (binary renamed to "v")
**When** path validation fails
**Then** error message says "v: path not found" (not "vdash:")

**And** given error messages suggest usage examples
**When** displayed to user
**Then** examples use actual binary name (e.g., "try: v add .")

**Technical Notes:**
- Search for hardcoded "vdash" in error messages across `internal/adapters/cli/`
- Replace with `BinaryName()` call or format string
- Exclude: help text examples (keep canonical for docs)

**Prerequisites:** Story 1.1

---

## Epic 2: Methodology Switching Detection

**User Value:** Users who switch methodologies mid-project (e.g., Speckit → BMAD) see the correct current methodology detected, not stale detection from the first-match-wins strategy.

**FR Coverage:** FR-P2-7, FR-P2-8, FR-P2-9, FR-P2-10, FR-P2-11

**Technical Context:**
- Current registry uses first-match-wins (runs detectors until one matches)
- New approach: run ALL detectors, compare artifact timestamps
- Most recent artifact modification wins
- Tie-breaker (< 1 hour difference): show both with warning

**Dependencies:** Epic 1 (recommended to ship together)

---

### Story 2.1: Implement DetectWithCoexistence Registry Method

As a developer,
I want the detector registry to run all detectors and collect all matches,
So that methodology coexistence can be properly evaluated.

**Acceptance Criteria:**

**Given** a project with both Speckit and BMAD artifacts
**When** `DetectWithCoexistence(ctx, path)` is called
**Then** it returns `[]*DetectionResult` containing both matches

**And** given a project with only Speckit artifacts
**When** `DetectWithCoexistence(ctx, path)` is called
**Then** it returns `[]*DetectionResult` with single Speckit match

**And** given a project with no methodology artifacts
**When** `DetectWithCoexistence(ctx, path)` is called
**Then** it returns empty slice (no matches)

**Technical Notes:**
- Location: `internal/adapters/detectors/registry.go`
- New method signature: `func (r *Registry) DetectWithCoexistence(ctx context.Context, path string) ([]*DetectionResult, error)`
- Does NOT replace `DetectAll()` - this is a new method
- Each DetectionResult needs `ArtifactTimestamp time.Time` field

**Prerequisites:** None

---

### Story 2.2: Add Artifact Timestamp to Detection Results

As a developer,
I want each detector to report the most recent artifact modification time,
So that the registry can compare timestamps across methodologies.

**Acceptance Criteria:**

**Given** Speckit detector finds `specs/001-feature/plan.md` (modified 2h ago) and `specs/001-feature/spec.md` (modified 1d ago)
**When** detection runs
**Then** DetectionResult.ArtifactTimestamp is 2h ago (most recent)

**And** given BMAD detector finds `sprint-status.yaml` (modified 30m ago)
**When** detection runs
**Then** DetectionResult.ArtifactTimestamp is 30m ago

**Technical Notes:**
- Update `internal/core/domain/detection_result.go` to add `ArtifactTimestamp time.Time`
- Update Speckit detector to track most recent mtime across all spec folders
- Update BMAD detector to check: `sprint-status.yaml`, `config.yaml`, `implementation-artifacts/`
- Use `os.Stat()` for mtime retrieval

**Prerequisites:** Story 2.1

---

### Story 2.3: Implement Most-Recent-Artifact-Wins Selection

As a user,
I want the dashboard to show the methodology I'm currently using (based on recent activity),
So that switching methodologies is seamlessly detected.

**Acceptance Criteria:**

**Given** project has Speckit artifacts from 1 week ago AND BMAD artifacts from 1 hour ago
**When** methodology detection runs
**Then** BMAD is selected as active methodology

**And** given project has BMAD artifacts from 2 days ago AND Speckit artifacts from 5 minutes ago
**When** methodology detection runs
**Then** Speckit is selected as active methodology

**Technical Notes:**
- Add selection logic in `DetectWithCoexistence()` or new helper
- Compare ArtifactTimestamp across all DetectionResults
- Return single winner when difference > 1 hour
- Update dashboard to use `DetectWithCoexistence()` instead of `DetectAll()`

**Prerequisites:** Story 2.2

---

### Story 2.4: Implement Coexistence Warning for Tie-Breaker

As a user,
I want to see a warning when the dashboard can't determine which methodology is active,
So that I understand why both are shown.

**Acceptance Criteria:**

**Given** project has Speckit artifacts from 30 minutes ago AND BMAD artifacts from 45 minutes ago
**When** methodology detection runs (difference < 1 hour)
**Then** both methodologies are returned with coexistence warning

**And** the warning text is "Multiple methodologies detected with similar activity"

**Technical Notes:**
- Threshold constant: `const CoexistenceThreshold = 1 * time.Hour`
- Add `CoexistenceWarning bool` field to DetectionResult or return type
- TUI will check this flag to display warning

**Prerequisites:** Story 2.3

---

### Story 2.5: Update TUI to Display Methodology Coexistence

As a user,
I want to see both methodologies in the dashboard when coexistence is detected,
So that I can understand my project's mixed state.

**Acceptance Criteria:**

**Given** DetectWithCoexistence returns two methodologies with CoexistenceWarning=true
**When** dashboard renders
**Then** stage column shows "Speckit/BMAD" or "⚠️ Mixed"

**And** detail panel shows warning message explaining both methodologies are active

**And** status column shows the combined stage from primary methodology

**Technical Notes:**
- Update `internal/adapters/tui/dashboard.go` to handle multiple results
- Display format options: "Speckit/BMAD", "⚠️ Speckit+BMAD", or custom
- Detail panel: "Warning: Both Speckit (Plan stage) and BMAD (Epic stage) detected"
- Primary methodology for stage display: use most recent even in tie-breaker

**Prerequisites:** Story 2.4

---

## Epic 3: Sub-1-Minute Agent Detection

**User Value:** Users know INSTANTLY (< 1 second) when their Claude Code agent is waiting for input, eliminating the 10-minute blind spot. This is THE killer feature of Phase 2.

**FR Coverage:** FR-P2-1, FR-P2-2, FR-P2-3, FR-P2-4, FR-P2-5, FR-P2-6

**Technical Context:**
- Claude Code writes JSONL logs to `~/.claude/projects/{project-hash}/*.jsonl`
- Detection logic: `assistant + stop_reason: "end_turn"` → WAITING, `assistant + stop_reason: "tool_use"` → WORKING
- New interface: `AgentActivityDetector` in `internal/core/ports/`
- Two adapters: `claude_code.go` (log-based, high confidence), `generic.go` (file-based, low confidence)
- Tail-optimized reading (last N entries) for performance

**Dependencies:** Epic 2 (can be parallelized with Epic 2 stories 3-5)

---

### Story 3.1: Define AgentActivityDetector Interface and Types

As a developer,
I want a well-defined interface for agent activity detection,
So that multiple detector implementations can be plugged in.

**Acceptance Criteria:**

**Given** the interface is defined
**When** a new detector is implemented
**Then** it conforms to `AgentActivityDetector` interface

**And** AgentState struct contains: Tool, Status, Duration, Confidence

**And** AgentStatus enum has: Working, WaitingForUser, Inactive, Unknown

**Technical Notes:**
- Location: `internal/core/ports/agent_activity_detector.go`
- Interface:
  ```go
  type AgentActivityDetector interface {
      Detect(ctx context.Context, projectPath string) (AgentState, error)
      Name() string  // "Claude Code", "Generic"
  }
  ```
- Types:
  ```go
  type AgentState struct {
      Tool       string        // "Claude Code", "Unknown"
      Status     AgentStatus   // Working, WaitingForUser, Inactive
      Duration   time.Duration // How long in current state
      Confidence Confidence    // High, Low
  }

  type AgentStatus string
  const (
      AgentWorking       AgentStatus = "working"
      AgentWaitingForUser AgentStatus = "waiting"
      AgentInactive      AgentStatus = "inactive"
      AgentUnknown       AgentStatus = "unknown"
  )
  ```

**Prerequisites:** None

---

### Story 3.2: Implement Claude Code Project Path Matcher

As a developer,
I want to match a project path to its Claude Code log directory,
So that I can find the relevant JSONL logs for detection.

**Acceptance Criteria:**

**Given** project path `/Users/jongkuk/projects/vibe-dash`
**When** matcher runs
**Then** it finds corresponding directory in `~/.claude/projects/` by comparing paths

**And** given Claude Code hashes project paths for directory names
**When** matcher runs
**Then** it reads each project's metadata to find path match

**And** given no matching Claude Code project exists
**When** matcher runs
**Then** it returns empty result (no error, just no match)

**Technical Notes:**
- Location: `internal/adapters/agentdetectors/claude_code.go`
- Claude Code stores project path in metadata within each project folder
- Read `~/.claude/projects/*/` directories and match by stored path
- Cache matches to avoid repeated filesystem scans
- Handle case: Claude Code not installed (directory doesn't exist)

**Prerequisites:** Story 3.1

---

### Story 3.3: Implement Claude Code JSONL Log Parser

As a developer,
I want to parse Claude Code's JSONL log files efficiently,
So that I can extract the last assistant message and its stop_reason.

**Acceptance Criteria:**

**Given** a JSONL log file with 1000 entries
**When** parser runs
**Then** it reads only the last N entries (tail-optimized, not full file)

**And** given log entry `{"type": "assistant", "stop_reason": "end_turn", "timestamp": "..."}`
**When** parsed
**Then** it extracts type, stop_reason, and timestamp correctly

**And** given malformed JSONL line
**When** parsing
**Then** it skips the line gracefully and continues

**Technical Notes:**
- Location: `internal/adapters/agentdetectors/claude_code.go`
- Tail-read optimization: seek to end of file, read backwards until N entries found
- N = 50 entries should be sufficient (configurable)
- Use `encoding/json` for parsing
- Handle: empty files, permission errors, corrupt entries

**Prerequisites:** Story 3.2

---

### Story 3.4: Implement Agent State Detection Logic

As a developer,
I want to determine agent state from parsed log entries,
So that the dashboard can show Working/WaitingForUser/Inactive.

**Acceptance Criteria:**

**Given** last assistant message has `stop_reason: "end_turn"`
**When** detection runs
**Then** state is `WaitingForUser` with High confidence

**And** given last assistant message has `stop_reason: "tool_use"`
**When** detection runs
**Then** state is `Working` with High confidence

**And** given no assistant messages in last N entries
**When** detection runs
**Then** state is `Inactive` with High confidence

**And** given last message timestamp is 2 hours ago
**When** detection runs
**Then** Duration is 2 hours

**Technical Notes:**
- Detection logic in `internal/adapters/agentdetectors/claude_code.go`
- Look for most recent `type: "assistant"` entry
- Extract `stop_reason` field
- Calculate Duration from timestamp to now
- Return AgentState with all fields populated

**Prerequisites:** Story 3.3

---

### Story 3.5: Implement Generic File-Activity Fallback Detector

As a user,
I want agent detection to work for non-Claude-Code projects,
So that I still get waiting detection (even if less accurate).

**Acceptance Criteria:**

**Given** a project without Claude Code logs
**When** detection runs
**Then** generic detector is used as fallback

**And** given no file activity for 10+ minutes
**When** generic detector runs
**Then** state is `WaitingForUser` with Low confidence

**And** given recent file activity (< 10 minutes)
**When** generic detector runs
**Then** state is `Working` with Low confidence

**Technical Notes:**
- Location: `internal/adapters/agentdetectors/generic.go`
- Use existing file activity tracking (last modified time)
- Threshold: 10 minutes (from Phase 1 behavior)
- Confidence: always Low (heuristic-based)
- Implements same `AgentActivityDetector` interface

**Prerequisites:** Story 3.1

---

### Story 3.6: Integrate Agent Detection into TUI Dashboard

As a user,
I want to see agent status for each project in the dashboard,
So that I know which agents need my attention.

**Acceptance Criteria:**

**Given** project has Claude Code with state `WaitingForUser` for 2 hours
**When** dashboard renders
**Then** shows "⏸️ WAITING 2h" in status column with bold red styling

**And** given project has agent state `Working`
**When** dashboard renders
**Then** shows "🔄 Working" or no special indicator (normal state)

**And** given status bar counts
**When** any project is WaitingForUser
**Then** status bar shows "⏸️ N WAITING" in red

**Technical Notes:**
- Update `internal/adapters/tui/dashboard.go` to call agent detector
- Add agent status to project row rendering
- Integrate with existing ⏸️ WAITING styling (bold red)
- Update status bar WAITING count

**Prerequisites:** Story 3.4, Story 3.5

---

### Story 3.7: Display Confidence Level in Detail Panel

As a user,
I want to see how confident the agent detection is,
So that I know whether to trust log-based vs heuristic detection.

**Acceptance Criteria:**

**Given** Claude Code log-based detection
**When** detail panel shows agent status
**Then** displays "Agent: ⏸️ WAITING 2h (High confidence - Claude Code logs)"

**And** given generic file-based detection
**When** detail panel shows agent status
**Then** displays "Agent: ⏸️ WAITING 15m (Low confidence - file activity)"

**Technical Notes:**
- Update `internal/adapters/tui/components/detail/panel.go`
- Add agent status section to detail panel
- Show: status, duration, confidence level, detection method
- Confidence affects display: High = normal, Low = dim or with caveat

**Prerequisites:** Story 3.6

---

## Epic 4: Progress Metrics & Stats View (Experimental)

**User Value:** Users can visualize project velocity, time-per-stage patterns, and work history. This is an experimental feature - designed to be cleanly removable if unproven.

**FR Coverage:** FR-P2-12, FR-P2-13, FR-P2-14, FR-P2-15, FR-P2-16, FR-P2-17, FR-P2-18

**Technical Context:**
- **Isolation Principle:** Separate database (`metrics.db`), separate packages, event-based architecture
- Location: `~/.vibe-dash/metrics.db` (NOT in `state.db`)
- New packages: `internal/adapters/metrics/`, `internal/adapters/persistence/metrics/`, `internal/adapters/tui/statsview/`
- TUI entry: `'s'` key from Dashboard → Stats View → `Esc`/`'q'` back
- NFRs: < 500ms render for 1 year data, < 20MB/year growth

**Dependencies:** Epics 1-3 (ship last as experimental)

---

### Story 4.1: Create Metrics Database Schema and Adapter

As a developer,
I want a separate metrics database with clean schema,
So that metrics can be removed without affecting core functionality.

**Acceptance Criteria:**

**Given** first metrics event recorded
**When** adapter initializes
**Then** creates `~/.vibe-dash/metrics.db` (separate from state.db)

**And** schema includes `stage_transitions` table with: id, project_id, from_stage, to_stage, transitioned_at

**And** metrics database failure does not crash the dashboard

**Technical Notes:**
- Location: `internal/adapters/persistence/metrics/repository.go`
- Schema:
  ```sql
  CREATE TABLE stage_transitions (
      id TEXT PRIMARY KEY,
      project_id TEXT NOT NULL,
      from_stage TEXT NOT NULL,
      to_stage TEXT NOT NULL,
      transitioned_at TEXT NOT NULL
  );
  ```
- Use SQLite WAL mode
- Graceful degradation: if metrics.db fails, log warning, continue without metrics

**Prerequisites:** None

---

### Story 4.2: Implement Stage Transition Event Recording

As a developer,
I want the system to record when projects change stages,
So that metrics can be calculated from historical data.

**Acceptance Criteria:**

**Given** project stage changes from "Plan" to "Tasks"
**When** detection service runs
**Then** stage_transitions record is created with from_stage="Plan", to_stage="Tasks"

**And** given metrics recording fails
**When** stage transition happens
**Then** core functionality continues (graceful degradation)

**And** events include ISO 8601 timestamp

**Technical Notes:**
- Location: `internal/adapters/metrics/recorder.go`
- Subscribe to stage changes from DetectionService (event pattern)
- Generate UUID for event id
- Handle: first detection (from_stage = ""), rapid transitions (debounce?)

**Prerequisites:** Story 4.1

---

### Story 4.3: Create Stats View TUI Component

As a user,
I want a dedicated full-screen Stats View,
So that I can focus on metrics without dashboard clutter.

**Acceptance Criteria:**

**Given** I press `'s'` from Dashboard
**When** Stats View opens
**Then** it shows full-screen metrics view

**And** given I press `Esc` or `'q'` in Stats View
**When** handling key
**Then** it returns to Dashboard

**And** Stats View has header "STATS" and shows project list with metrics

**Technical Notes:**
- Location: `internal/adapters/tui/statsview/view.go`
- Implement as separate Bubble Tea model (not panel toggle)
- Main model switches between Dashboard and StatsView
- Pass project list and metrics data to StatsView

**Prerequisites:** Story 4.1, Story 4.2

---

### Story 4.4: Implement Activity Sparklines

As a user,
I want to see activity sparklines per project,
So that I can quickly visualize work patterns.

**Acceptance Criteria:**

**Given** project has stage transitions over past 30 days
**When** Stats View renders
**Then** shows sparkline character graph (e.g., `▁▂▃▅▂▁▇▅▂`)

**And** sparkline represents activity density per day/week

**And** projects with no recent activity show flat sparkline

**Technical Notes:**
- Location: `internal/adapters/tui/statsview/sparkline.go`
- Use Unicode block characters: `▁▂▃▄▅▆▇█`
- Calculate activity buckets (e.g., 7 days or 4 weeks)
- Handle: empty data, single event, many events

**Prerequisites:** Story 4.3

---

### Story 4.5: Implement Time-Per-Stage Breakdown

As a user,
I want to see how much time I spend in each stage,
So that I can identify bottlenecks in my workflow.

**Acceptance Criteria:**

**Given** project has transitions: Plan(3h) → Tasks(1h) → Implement(5h)
**When** Stats View shows breakdown
**Then** displays: "Plan: 3h | Tasks: 1h | Implement: 5h"

**And** breakdown shows percentage or bar chart representation

**And** handles in-progress stage (current stage shows time since last transition)

**Technical Notes:**
- Location: `internal/adapters/tui/statsview/breakdown.go`
- Calculate duration between consecutive transitions
- Current stage: now - last_transition_time
- Display options: simple text, bar chart, or percentage

**Prerequisites:** Story 4.3

---

### Story 4.6: Add Date Range Selector

As a user,
I want to select a date range for metrics,
So that I can focus on specific time periods.

**Acceptance Criteria:**

**Given** Stats View is open
**When** I navigate date range
**Then** metrics update to show only selected period

**And** default range is "Last 30 days"

**And** presets available: 7 days, 30 days, 90 days, 1 year, All time

**Technical Notes:**
- Location: `internal/adapters/tui/statsview/daterange.go`
- Keyboard navigation: `[` / `]` to cycle presets, or number keys
- Filter stage_transitions query by date range
- Display current range in header

**Prerequisites:** Story 4.4, Story 4.5

---

### Story 4.7: Wire Stats View into Dashboard

As a user,
I want to access Stats View from the Dashboard,
So that switching between views is seamless.

**Acceptance Criteria:**

**Given** I am in Dashboard view
**When** I press `'s'`
**Then** Stats View opens

**And** status bar shows `[s] stats` hint

**And** Stats View preserves Dashboard selection on return

**Technical Notes:**
- Update `internal/adapters/tui/model.go` to handle view switching
- Add `'s'` to key bindings in `internal/adapters/tui/keys.go`
- Update status bar to show stats shortcut
- Store Dashboard state when switching to Stats View

**Prerequisites:** Story 4.6

---

## FR Coverage Matrix

| FR ID | Description | Epic | Story |
|-------|-------------|------|-------|
| FR-P2-1 | Detect Claude Code via JSONL log parsing | Epic 3 | 3.3, 3.4 |
| FR-P2-2 | Identify agent state (Working/Waiting/Inactive) | Epic 3 | 3.4 |
| FR-P2-3 | Display elapsed time since state change | Epic 3 | 3.4, 3.6 |
| FR-P2-4 | Show confidence level (High/Low) | Epic 3 | 3.7 |
| FR-P2-5 | Fallback to file-activity detection | Epic 3 | 3.5 |
| FR-P2-6 | Match project path to Claude Code logs | Epic 3 | 3.2 |
| FR-P2-7 | Detect multiple methodologies | Epic 2 | 2.1 |
| FR-P2-8 | Compare artifact timestamps | Epic 2 | 2.2 |
| FR-P2-9 | Select based on most recent artifact | Epic 2 | 2.3 |
| FR-P2-10 | Display coexistence warning | Epic 2 | 2.4 |
| FR-P2-11 | Display both methodologies on tie | Epic 2 | 2.5 |
| FR-P2-12 | Record stage transition events | Epic 4 | 4.2 |
| FR-P2-13 | Store metrics in separate database | Epic 4 | 4.1 |
| FR-P2-14 | View Stats View via 's' key | Epic 4 | 4.7 |
| FR-P2-15 | Exit Stats View via Esc/'q' | Epic 4 | 4.3 |
| FR-P2-16 | Display activity sparklines | Epic 4 | 4.4 |
| FR-P2-17 | Display time-per-stage breakdown | Epic 4 | 4.5 |
| FR-P2-18 | Select date range for metrics | Epic 4 | 4.6 |
| FR-P2-19 | Use actual binary name in errors | Epic 1 | 1.3 |
| FR-P2-20 | Use actual binary name in version | Epic 1 | 1.2 |

---

## Summary

| Epic | Stories | FRs | Priority |
|------|---------|-----|----------|
| Epic 1: Dynamic Binary Name | 3 | 2 | Must Ship (Quick Win) |
| Epic 2: Methodology Switching | 5 | 5 | Must Ship (Bug Fix) |
| Epic 3: Agent Detection | 7 | 6 | Must Ship (Core Feature) |
| Epic 4: Progress Metrics | 7 | 7 | Should Ship (Experimental) |
| **Total** | **22** | **20** | |

**Implementation Order:**
1. Epic 1 - Ship first (quick win, foundation)
2. Epic 2 - Bug fix, unblocks testing
3. Epic 3 - Core feature, most effort
4. Epic 4 - Experimental, ship last

**Go/No-Go Criteria (from PRD):**
- [ ] Agent detection latency < 1 second for Claude Code projects
- [ ] Detection accuracy 95%+ on test sessions
- [ ] Methodology switching works for test cases
- [ ] No regressions in MVP functionality
- [ ] Stats View renders (even if basic)

---

_For implementation: Use the `create-story` workflow to generate individual story implementation plans from this epic breakdown._

