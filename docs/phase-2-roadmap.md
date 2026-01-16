# Phase 2 Improvements Roadmap

> **Status:** In Progress
> **Created:** 2026-01-15
> **Last Updated:** 2026-01-15
> **Context:** MVP complete, planning next polish phase

## Overview

This document captures the improvement planning discussion for vdash Phase 2. It serves as a reference for future conversations to pick up where we left off.

## Evaluated Items

### Approved for Phase 2

| Item | Effort | Status | Notes |
|------|--------|--------|-------|
| Dynamic binary name | Low | Ready to implement | Use `os.Args[0]` instead of hardcoded "vdash" |
| Progress Metrics & Graphs | Medium | Design complete | Dedicated Stats View, ntcharts, loose coupling |
| Methodology switching detection | Medium | Design complete | Detect all, compare artifact timestamps |
| Sub-1-minute agent detection | Medium | Design complete | Claude Code log-based, adapter pattern |
| Plugin architecture for detectors | Medium | Future | Great to have, design when ready |
| API/HTTP adapter | Medium | Architecture ready | Hexagonal architecture supports this; implement when needed |
| Performance benchmarks | Medium | NFRs defined | Create benchmark story when desired |

### Needs Discussion

| Item | Question | Status |
|------|----------|--------|
| Progress Metrics & Graphs | Detailed TUI wireframes, key bindings | Deferred to Product Brief |
| Sub-1-minute agent detection | - | Design complete ✓ |
| Methodology switching detection | - | Design complete ✓ |

### Deferred (Not Phase 2)

| Item | Reasoning | Decision Date |
|------|-----------|---------------|
| Fuzzy search across projects | Low priority - user manages ~3 projects currently. Can wait until project count grows. | 2026-01-15 |
| `vdash recent` command | Dashboard already sufficient for morning routine. No additional value. | 2026-01-15 |
| Windows platform support | User doesn't have Windows PC. Defer until needed. | 2026-01-15 |

### Already Implemented (Discovered During Review)

| Item | Implementation | Reference |
|------|----------------|-----------|
| Per-project notes | 'n' key in TUI, `vdash note` CLI | Story 3.7 |
| Pipeline summary output | `make test/build/lint` show summary boxes | Story 9.5-5 |
| TUI behavioral testing | Teatest adopted, tests in `teatest_*_test.go` | Story 9.1 |

## Detailed Item Notes

### 1. Dynamic Binary Name

**Problem:** Binary name "vdash" is hardcoded in several places.

**Solution:** Use `os.Args[0]` or `RootCmd.Use` for dynamic name.

**Files to update:**
- Version output template
- Help text examples (debatable - may prefer canonical name)

**Priority:** Low effort, low impact. Good quick win.

**Reference:** `docs/IMPROVEMENTS.md`

---

### 2. Progress Metrics & Graphs

**User Interest:** High - wants to see progress over time visually.

**Status:** Design decisions in progress (2026-01-15)

#### Data Collection Strategy

**Philosophy:** Collect everything possible, decide visualization later.

**Architecture Decision: Loose Coupling**
- **Separate database:** `~/.vibe-dash/metrics.db` (isolated from `state.db`)
- **Separate packages:** `internal/adapters/metrics/` and `internal/adapters/persistence/metrics/`
- **Event-based:** Core emits events, metrics subscribes - removable without breaking core
- **Rationale:** Experimental feature; must be droppable with minimal impact

**Storage Decision: Raw Events**
- Store every event with timestamp (no pre-aggregation)
- Aggregate on-demand during read/display
- Retention policy can be added later if storage grows
- Estimated: ~500 events/day for 3 projects = ~180k/year = ~20MB (negligible)

#### Data Sources Available

**Tier 1 - Already in Database (queryable now):**
- `created_at` - project age/tenure
- `current_stage` - stage distribution
- `detected_method` - methodology breakdown
- `confidence` - detection quality
- `last_activity_at` - inactivity duration
- `state`, `hibernated_at` - hibernation metrics

**Tier 2 - Observable but not stored (needs collection):**
- File events (create/modify/delete) - currently aggregated to LastActivityAt only
- Activity patterns (time of day, day of week)
- Stage transitions - only current stage stored, no history

**Tier 3 - New tables needed:**
- `stage_transitions` - track stage changes with timestamps
- `file_events` - raw file activity events

#### TUI Design Decision

**Dedicated Stats View** (like Log Viewer pattern)
- Full screen for visualization (not a panel)
- Own key bindings for navigation
- Entry: `'s'` key from Dashboard
- Exit: `Esc` or `'q'` back to Dashboard

**Rationale:**
- Stats panel would conflict with detail panel toggle
- Complex interaction (date range, metric type) needs dedicated space
- Charts/graphs need full width

**Proposed structure:**
```
Dashboard → 's' → Stats View → 'q' → Dashboard
                      │
                      ├── Tab/Section: Activity Timeline
                      ├── Tab/Section: Stage Progress
                      └── Tab/Section: Project Comparison
```

#### Key Insight: Missing Stage History

Current gap: Only `current_stage` stored. No record of when stages changed.

**Required new table:**
```sql
CREATE TABLE stage_transitions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    from_stage TEXT NOT NULL,
    to_stage TEXT NOT NULL,
    transitioned_at TEXT NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id)
);
```

This enables: time-per-stage, progression velocity, regression detection.

#### Visualization Direction (Approved)

**Charting Library:** ntcharts (native Bubble Tea/LipGloss integration)
- Time Series, Bar Chart, Sparkline, Heat Map, Line Chart
- Braille characters for high-resolution plots
- Mouse support via BubbleZone

**Visual Inspiration:** Gonzo TUI
- 2x2 grid layouts
- ASCII heatmaps (░▒▓█)
- Vim navigation + mouse

**Rough Concept:**
- Tabbed sections (Activity, Stages, Patterns)
- Date range selector
- Sparklines for compact activity per project
- Heatmap for work patterns (hour × day)
- Stage progression timeline

#### Still to Detail (Product Brief Phase)

- Exact layout dimensions and responsiveness
- Specific charts for each metric section
- Key bindings and interaction flow
- Date range picker UI design
- Small terminal handling
- Color scheme decisions
- Empty state design

**Discussion Log:**
- 2026-01-15: Decided on loose coupling (separate DB), raw event storage, dedicated TUI view
- 2026-01-15: Approved ntcharts + Gonzo-inspired design direction; detailed wireframes deferred to Product Brief

---

### 3. Methodology Switching Detection

**Problem:** Projects that switch methodologies (e.g., Speckit → BMAD) are detected as the first-registered methodology due to first-match-wins strategy.

**Example:** claude-code-log-viewer-cli has both `specs/` and `_bmad/` folders. Currently detected as Speckit despite BMAD being actively used.

**Status:** Design complete (2026-01-15)

#### Root Cause

**Registration order in `cmd/vdash/main.go`:**
```go
registry.Register(speckit.NewSpeckitDetector())  // First - always wins
registry.Register(bmad.NewBMADDetector())        // Never reached if Speckit matches
```

When both methodologies exist, Speckit's `CanDetect()` returns true and BMAD is never evaluated.

#### Design Decision: Most Recent Artifact

**Approach:**
1. Run ALL detectors (not first-match-wins)
2. If only one matches → return it
3. If multiple match → compare artifact timestamps
4. Return the one with more recent activity

**Timestamp Sources:**
- **Speckit:** mtime of most recent spec folder
- **BMAD:** mtime of `sprint-status.yaml`, `config.yaml`, or `implementation-artifacts/`

**Tie-breaker (fresh clone scenario):**
- When timestamps are equal, show BOTH with a warning
- User's first activity will update one methodology's artifacts
- Next detection will have clear winner

#### Implementation

**New method in registry.go:**
```go
func (r *Registry) DetectWithCoexistence(ctx context.Context, path string) ([]*domain.DetectionResult, error) {
    var results []*domain.DetectionResult

    for _, detector := range r.detectors {
        if detector.CanDetect(ctx, path) {
            if result, err := detector.Detect(ctx, path); err == nil {
                results = append(results, result)
            }
        }
    }

    if len(results) <= 1 {
        return results, nil  // Single or no match
    }

    // Multiple matches - compare timestamps
    return selectByMostRecent(results), nil
}
```

**Selection logic:**
```go
func selectByMostRecent(results []*domain.DetectionResult) []*domain.DetectionResult {
    // Get artifact mtime for each result
    // If clear winner (>1 hour difference) → return single result
    // If tie (within 1 hour) → return all with reasoning noting coexistence
}
```

#### TUI Display

**Single methodology:** Normal display (no change)

**Multiple methodologies (tie):**
```
Method: speckit / bmad (coexisting)
Stage: [from most recent or show both]
⚠ Multiple methodologies detected - using most recent activity
```

#### Discussion Log

- 2026-01-15: Identified issue with claude-code-log-viewer-cli (Speckit detected, BMAD active)
- 2026-01-15: Root cause: first-match-wins registry strategy
- 2026-01-15: Decided on Option 1: detect all, compare artifact timestamps
- 2026-01-15: Tie-breaker: show both with warning (handles fresh clone case)

---

### 4. Sub-1-minute Agent Detection (Claude Code)

**Current State:** 10-minute threshold for detecting "agent waiting" state.

**User Interest:** High - would refine core UX experience.

**Status:** Design complete (2026-01-15)

#### Research Findings: AI Tool Log Accessibility

| Tool | Log Location | Format | Accessibility |
|------|-------------|--------|---------------|
| **Claude Code** | `~/.claude/projects/**/*.jsonl` | Structured JSONL | Excellent |
| **Cursor** | `~/Library/Application Support/Cursor/logs/` | VS Code extension logs | Medium |
| **Copilot** | VS Code Output panel only | Unstructured | Poor |
| **Windsurf** | Built-in command only | Unknown | Poor |

**Conclusion:** Only Claude Code has reliable, structured, file-based logs suitable for real-time parsing.

#### Claude Code Log Structure Analysis

**Entry Types:**
```
assistant + stop_reason: "end_turn"  → AI finished, WAITING FOR USER
assistant + stop_reason: "tool_use"  → AI is working (tool in progress)
user + tool_result content           → Tool output (AI still working)
user + actual content                → User gave input
```

**The Key Signal: `stop_reason: "end_turn"`**

When the last `assistant` entry has `stop_reason: "end_turn"`, the AI has finished and is waiting for user input. The `timestamp` field indicates when this state began.

#### Design Decision: Adapter Pattern

**Strategy:**
- Claude Code: Use log parsing for immediate/accurate detection
- Other tools: Fall back to current behavior (file activity threshold)

**Proposed Interface:**
```go
// ports/agent_activity_detector.go
type AgentActivityDetector interface {
    Detect(ctx context.Context, projectPath string) (AgentState, error)
}

type AgentState struct {
    Tool        string        // "Claude Code", "Unknown"
    Status      AgentStatus   // Working, WaitingForUser, Inactive
    Duration    time.Duration // How long in current state
    Confidence  Confidence    // High (log-based), Low (heuristic)
}

type AgentStatus int
const (
    AgentStatusUnknown AgentStatus = iota
    AgentStatusWorking
    AgentStatusWaitingForUser
    AgentStatusInactive
)
```

**Package Structure:**
```
internal/
├── core/ports/
│   └── agent_activity_detector.go  # Interface
├── adapters/agentdetectors/
│   ├── claude_code.go              # Log-based (immediate)
│   └── generic.go                  # File activity (10-min threshold)
```

**Detection Logic (Claude Code):**
```
1. Find most recent session file
2. Read last N entries
3. Find last "assistant" entry
4. If stop_reason == "end_turn":
   - Status = WaitingForUser
   - Duration = now - entry.timestamp
   - Confidence = High
5. Else (stop_reason == "tool_use" OR last is tool_result):
   - Status = Working
   - Confidence = High
```

#### Benefits

- **Immediate detection:** No 10-minute wait for Claude Code projects
- **Accurate state:** Know if AI is working vs waiting
- **Extensible:** Adapter pattern allows adding other tools later
- **Backward compatible:** Generic detector maintains current behavior

#### Discussion Log

- 2026-01-15: Researched log accessibility for Cursor, Copilot, Windsurf - only Claude Code viable
- 2026-01-15: Analyzed Claude Code JSONL structure - `stop_reason: "end_turn"` is the key signal
- 2026-01-15: Decided on adapter pattern - Claude Code gets premium detection, others use file activity

---

### 5. Plugin Architecture for Detectors

**Vision:** Public API for community-contributed method detectors.

**Current State:** Internal detector interface exists (`MethodDetector` port).

**Next Steps:**
1. Document the detector interface
2. Create contribution guide
3. Enable external detector implementations

**Priority:** Future - design when community interest grows.

---

### 6. API/HTTP Adapter (Web Dashboard Readiness)

**Architecture Status:** Ready for implementation.

**How it works:**
- Hexagonal architecture with ports & adapters
- Services return domain types (not TUI-specific)
- CLI already has `--json` output with `--api-v1` versioning
- Adding HTTP adapter = create `internal/adapters/api/http/`
- Zero core changes needed

**External developers can already:**
- Read SQLite database directly
- Parse CLI JSON output

**Priority:** Implement when someone needs it.

---

### 7. Performance Benchmarks

**NFRs Defined (from PRD):**
- `NFR-P1`: Dashboard renders in <100ms for 20 projects
- `NFR-P2`: Dashboard startup in <1 second
- `NFR-P3`: CLI commands respond in <500ms
- `NFR-P4`: Project initialization <2 seconds
- `NFR-P5`: TUI auto-refreshes every 5-10 seconds
- `NFR-P6`: File changes detected within 5-10 seconds

**Current State:** No `_bench_test.go` files exist yet.

**Next Steps:** Create Go benchmark tests validating NFRs when desired.

---

## Discussion Sessions

### Session 1: Progress Metrics & Graphs
- **Date:** 2026-01-15
- **Status:** In Progress
- **Decisions Made:**
  - Loose coupling architecture (separate metrics.db, separate packages)
  - Raw event storage (no pre-aggregation)
  - Dedicated TUI view (not panel toggle)
  - Need stage_transitions table for time-per-stage metrics
- **Still Open:** Visualization format, date range UI, priority of metrics

### Session 2: Sub-1-minute Agent Detection
- **Date:** 2026-01-15
- **Status:** Complete
- **Decisions Made:**
  - Adapter pattern for extensibility
  - Claude Code: Log-based detection using `stop_reason: "end_turn"`
  - Other tools: Fall back to file activity threshold (existing behavior)
  - Interface: `AgentActivityDetector` with `AgentState` result
- **Key Finding:** Only Claude Code has structured, accessible logs

### Session 3: Methodology Switching Detection
- **Date:** 2026-01-15
- **Status:** Complete
- **Decisions Made:**
  - Detect ALL methodologies (not first-match-wins)
  - Compare artifact timestamps to determine active methodology
  - Tie-breaker (fresh clone): Show both with warning
  - User's first activity resolves the tie naturally
- **Key Finding:** First-match-wins registry strategy causes incorrect detection when projects switch methodologies

---

## Phase 2 Scope Summary

**Confirmed for Phase 2:**
- Dynamic binary name (quick win)
- Progress Metrics & Graphs (dedicated Stats View, ntcharts)
- Methodology switching detection (detect all, compare timestamps)
- Sub-1-minute agent detection for Claude Code (adapter pattern)

**Future (not Phase 2):**
- Plugin architecture for detectors
- API/HTTP adapter
- Performance benchmarks
- Windows support
- Fuzzy search
- `vdash recent` command
- Agent detection for other tools (Cursor, Copilot, Windsurf)

---

## Next Steps

1. ~~Complete discussion: Progress Metrics & Graphs~~ ✓
2. ~~Complete discussion: Sub-1-minute Agent Detection~~ ✓
3. ~~Complete discussion: Methodology Switching Detection~~ ✓
4. Create Product Brief for Phase 2
