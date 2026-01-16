---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
inputDocuments:
  - 'docs/analysis/product-brief-vdash-phase2-2026-01-15.md'
  - 'docs/phase-2-roadmap.md'
  - 'docs/prd.md'
workflowType: 'prd'
lastStep: 10
project_name: 'vdash-phase2'
user_name: 'Jongkuk Lim'
date: '2026-01-15'
---

# Product Requirements Document - vdash Phase 2

**Author:** Jongkuk Lim
**Date:** 2026-01-15
**Baseline:** [Phase 1 PRD](./prd.md) - All Phase 1 requirements remain in effect unless explicitly modified here.

---

## Executive Summary

**vdash Phase 2** transforms the dashboard from a passive state tracker into an **active collaboration partner** with Claude Code. The flagship feature—sub-1-minute agent detection—eliminates the 10-minute blind spot, letting developers know instantly when their AI agent is waiting for input versus actively working.

**Phase 2 Focus:** "Where am I?" (solved in MVP) → "**What needs me now?**"

### Key Deliverables

| Feature | Type | Priority |
|---------|------|----------|
| Sub-1-minute Agent Detection | Core | Must Ship |
| Methodology Switching Fix | Bug Fix | Must Ship |
| Dynamic Binary Name | Polish | Must Ship |
| Progress Metrics & Stats View | Experimental | Should Ship |

### Target User

**Jeff (Evolved)** - The power user already living in vdash who manages 3-6 Claude Code sessions simultaneously. He's no longer learning the tool; he's pushing its limits and frustrated by the 10-minute lag in knowing when Claude needs him.

### Success Criteria (Phase 2)

| Metric | MVP Baseline | Phase 2 Target |
|--------|--------------|----------------|
| Agent state detection latency | 10 minutes | < 1 second |
| Detection accuracy (Claude Code) | N/A | 95%+ |
| Methodology coexistence handling | First-match-wins (buggy) | Most-recent-artifact-wins |
| GitHub stars growth | Baseline | +100 from Phase 2 |

---

## Phase 2 Scope

### Tier 1: Must Ship

#### 1. Sub-1-Minute Agent Detection for Claude Code

**Problem:** Current 10-minute file-activity heuristic is too slow. Developers either check Claude sessions compulsively or miss waiting agents for extended periods.

**Solution:** Parse Claude Code's structured JSONL logs to detect agent state instantly.

**Detection Logic:**
```
assistant + stop_reason: "end_turn"  → WAITING FOR USER
assistant + stop_reason: "tool_use"  → WORKING (tool in progress)
```

**Architecture:**
- `AgentActivityDetector` interface in `internal/core/ports/`
- `claude_code.go` adapter - Log-based detection (immediate, high confidence)
- `generic.go` adapter - File activity fallback (10-minute threshold for non-Claude projects)

**Log Location:** `~/.claude/projects/{project-hash}/*.jsonl`

#### 2. Methodology Switching Detection Fix

**Problem:** Projects that switch methodologies (e.g., Speckit → BMAD) are detected incorrectly due to first-match-wins registry strategy.

**Solution:**
1. Run ALL detectors (not first-match-wins)
2. Compare artifact timestamps
3. Return methodology with most recent activity
4. Tie-breaker: Show both with warning (handles fresh clone)

**Timestamp Sources:**
- Speckit: mtime of most recent spec folder
- BMAD: mtime of `sprint-status.yaml`, `config.yaml`, or `implementation-artifacts/`

#### 3. Dynamic Binary Name

**Problem:** Binary name "vdash" is hardcoded in error messages and version output.

**Solution:** Use `os.Args[0]` or `RootCmd.Use` for dynamic name resolution.

**Scope:**
- Version output template
- Error messages
- Help examples remain canonical "vdash" (documentation consistency)

### Tier 2: Should Ship (Experimental)

#### 4. Progress Metrics & Stats View

**Purpose:** Visualize project velocity, time-per-stage, and work patterns.

**Architecture (Loose Coupling):**
- Separate database: `~/.vibe-dash/metrics.db`
- Separate packages: `internal/adapters/metrics/`, `internal/adapters/persistence/metrics/`
- Event-based: Core emits events, metrics subscribes
- **Removable:** If unproven, droppable without breaking core

**TUI Design:**
- Dedicated full-screen Stats View (not panel toggle)
- Entry: `'s'` key from Dashboard
- Exit: `Esc` or `'q'` back to Dashboard

**Minimum Viable Stats View:**
- Activity sparklines per project
- Time-per-stage breakdown
- Basic date range selector

**Required Schema Addition:**
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

### Out of Scope (Phase 2)

| Item | Reason |
|------|--------|
| Agent detection for Cursor/Copilot/Windsurf | No structured logs available |
| Plugin architecture for detectors | Future - design when community interest grows |
| API/HTTP adapter | Architecture ready, implement when needed |
| Windows support | User doesn't have Windows PC |
| Fuzzy search | Low priority with ~3 projects |
| `vdash recent` command | Dashboard sufficient for morning routine |

---

## Functional Requirements (Phase 2 Delta)

### Agent Detection

- **FR-P2-1:** System can detect Claude Code tool usage via JSONL log parsing
- **FR-P2-2:** System can identify agent state as Working, WaitingForUser, or Inactive
- **FR-P2-3:** System can display elapsed time since agent entered current state
- **FR-P2-4:** System can show confidence level (High for log-based, Low for heuristic)
- **FR-P2-5:** System falls back to file-activity detection for non-Claude-Code projects
- **FR-P2-6:** System can detect Claude Code session by matching project path to `~/.claude/projects/` structure

### Methodology Detection

- **FR-P2-7:** System can detect multiple methodologies in same project simultaneously
- **FR-P2-8:** System can compare artifact timestamps across detected methodologies
- **FR-P2-9:** System can select active methodology based on most recent artifact modification
- **FR-P2-10:** System can display coexistence warning when methodologies have similar timestamps (within 1 hour)
- **FR-P2-11:** System displays both methodologies in TUI when tie-breaker applies

### Progress Metrics (Experimental)

- **FR-P2-12:** System can record stage transition events with timestamps
- **FR-P2-13:** System can store metrics in separate database (`metrics.db`)
- **FR-P2-14:** Users can view Stats View via `'s'` key from Dashboard
- **FR-P2-15:** Users can exit Stats View via `Esc` or `'q'` key
- **FR-P2-16:** System can display activity sparklines per project
- **FR-P2-17:** System can display time-per-stage breakdown
- **FR-P2-18:** Users can select date range for metrics display

### Polish

- **FR-P2-19:** System uses actual binary name (`os.Args[0]`) in error messages
- **FR-P2-20:** System uses actual binary name in version output

---

## Non-Functional Requirements (Phase 2 Delta)

### Performance

- **NFR-P2-1:** Agent state detection completes in < 1 second for Claude Code projects
- **NFR-P2-2:** Claude Code log parsing reads only last N entries (tail-optimized, not full file)
- **NFR-P2-3:** Stats View renders in < 500ms for up to 1 year of metrics data
- **NFR-P2-4:** Metrics database growth < 20MB per year (estimated ~500 events/day for 3 projects)

### Reliability

- **NFR-P2-5:** Agent detection gracefully falls back to generic detector if Claude Code logs unavailable
- **NFR-P2-6:** Methodology detection returns valid result even when artifacts have identical timestamps
- **NFR-P2-7:** Metrics feature failure does not affect core dashboard functionality
- **NFR-P2-8:** Claude Code log format changes detected and handled gracefully (version detection)

### Extensibility

- **NFR-P2-9:** `AgentActivityDetector` interface supports adding new tool detectors without modifying existing code
- **NFR-P2-10:** Metrics system uses event-based architecture allowing removal without core changes

---

## Technical Architecture

### Agent Detection Adapter Pattern

```
internal/
├── core/ports/
│   └── agent_activity_detector.go    # Interface definition
├── adapters/agentdetectors/
│   ├── claude_code.go                # Log-based (immediate, high confidence)
│   └── generic.go                    # File activity (10-min threshold)
```

**Interface:**
```go
type AgentActivityDetector interface {
    Detect(ctx context.Context, projectPath string) (AgentState, error)
}

type AgentState struct {
    Tool        string        // "Claude Code", "Unknown"
    Status      AgentStatus   // Working, WaitingForUser, Inactive
    Duration    time.Duration // How long in current state
    Confidence  Confidence    // High, Low
}
```

### Methodology Detection Enhancement

```go
// registry.go - New method
func (r *Registry) DetectWithCoexistence(ctx context.Context, path string) ([]*DetectionResult, error)
```

**Selection Logic:**
1. Run ALL detectors
2. If single match → return it
3. If multiple → compare artifact timestamps
4. If clear winner (>1 hour difference) → return single result
5. If tie → return all with coexistence warning

### Metrics Architecture (Loose Coupling)

```
internal/
├── adapters/metrics/                  # Event collection
├── adapters/persistence/metrics/      # metrics.db access
└── adapters/tui/statsview/            # Dedicated TUI view
```

**Database:** `~/.vibe-dash/metrics.db` (separate from `state.db`)

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Claude Code log format changes | Version detection, graceful fallback to generic detector |
| Large session log files (performance) | Tail-read optimization, only parse last N entries |
| Metrics DB growth | Retention policy if needed (estimated ~20MB/year - negligible) |
| Stats View complexity creep | Ship minimal, iterate based on usage; cleanly removable |
| Methodology timestamp comparison edge cases | Tie-breaker shows both; user activity resolves naturally |

---

## Go/No-Go Criteria

### Ship Phase 2 if:
- [ ] Agent detection latency < 1 second for Claude Code projects
- [ ] Detection accuracy 95%+ on test sessions
- [ ] Methodology switching works for claude-code-log-viewer-cli test case
- [ ] No regressions in MVP functionality
- [ ] Stats View renders (even if basic)

### Delay if:
- [ ] Agent detection accuracy < 90%
- [ ] Core dashboard performance degraded
- [ ] Methodology switching breaks existing single-methodology projects

---

## Implementation Order

1. **Dynamic Binary Name** — Quick win, ship first
2. **Methodology Switching** — Bug fix, unblocks testing
3. **Sub-1-Minute Agent Detection** — Core feature, most effort
4. **Progress Metrics** — Experimental, ship last

---

## References

- [Phase 1 PRD](./prd.md) - Baseline requirements (all remain in effect)
- [Phase 2 Product Brief](./analysis/product-brief-vdash-phase2-2026-01-15.md) - Detailed problem analysis
- [Phase 2 Roadmap](./phase-2-roadmap.md) - Technical design decisions
