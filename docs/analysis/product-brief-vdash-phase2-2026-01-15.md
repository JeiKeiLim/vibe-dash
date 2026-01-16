---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments:
  - 'docs/phase-2-roadmap.md'
  - 'docs/analysis/product-brief-bmad-test-2025-12-05.md'
workflowType: 'product-brief'
lastStep: 5
project_name: 'vdash-phase2'
user_name: 'Jongkuk Lim'
date: '2026-01-15'
---

# Product Brief: vdash Phase 2

**Date:** 2026-01-15
**Author:** Jongkuk Lim

---

## Executive Summary

**vdash Phase 2** transforms the dashboard from a passive state tracker into an **active collaboration partner** with Claude Code. The flagship feature—sub-1-minute agent detection—eliminates the 10-minute blind spot, letting developers know instantly when their AI agent is waiting for input versus actively working.

**Core Enhancement:** Real-time Claude Code integration via structured log parsing (`~/.claude/projects/**/*.jsonl`), detecting `stop_reason: "end_turn"` as the signal that Claude is waiting for human input.

**Supporting Features:**
- **Progress Metrics & Graphs:** Dedicated Stats View visualizing project velocity, time-per-stage, and work patterns (experimental)
- **Methodology Switching Fix:** Detect all methodologies, compare artifact timestamps—projects switching from Speckit to BMAD now detected correctly
- **Dynamic Binary Name:** Polish fix using `os.Args[0]`

**Architecture Principle:** Adapter pattern ensures Claude Code gets premium detection today while the door stays open for Cursor, Copilot, and Windsurf when their logs become accessible.

**Target User:** Jeff, the power user already living in vdash, who wants zero friction knowing when to context-switch to a waiting Claude session.

---

## Core Vision

### Problem Statement

vdash MVP solved "where am I in my workflow?" but left a critical gap: **when is my AI agent waiting for me?** The current 10-minute file-activity heuristic is too slow. Developers either:
- Check Claude sessions compulsively (context-switching overhead)
- Miss waiting agents for extended periods (lost momentum)

The dashboard shows project state but not **agent state** with enough granularity to be actionable.

### Problem Impact

**Lost Time:**
- 10-minute detection lag means up to 10 minutes of idle agent time per session
- Multiply across 3-5 active Claude sessions = significant cumulative delay

**Cognitive Tax:**
- Uncertainty about agent state forces manual checking
- "Is Claude done yet?" becomes a recurring distraction

**Momentum Loss:**
- Agent finishes, waits silently, developer doesn't notice
- Flow state interrupted when developer finally checks and context-switches late

### Why Existing Solutions Fall Short

**Current vdash (MVP):**
- 10-minute threshold based on file activity only
- No insight into actual Claude session state
- Can't distinguish "Claude working on tools" from "Claude waiting for input"

**Claude Dashboard:**
- Shows sessions exist, not whether agent is working or waiting
- No integration with project methodology state

**Manual Checking:**
- Requires switching to Claude window
- Defeats purpose of centralized dashboard

### Proposed Solution

**Real-Time Agent Detection for Claude Code:**

Parse Claude Code's structured JSONL logs to detect agent state instantly:
```
assistant + stop_reason: "end_turn"  → WAITING FOR USER
assistant + stop_reason: "tool_use"  → WORKING (tool in progress)
```

**Adapter Architecture:**
```go
type AgentActivityDetector interface {
    Detect(ctx context.Context, projectPath string) (AgentState, error)
}
```
- `claude_code.go` — Log-based detection (immediate, high confidence)
- `generic.go` — File activity fallback (10-minute threshold)

**Progress Metrics (Experimental):**
- Separate `metrics.db` for loose coupling (droppable if unproven)
- Dedicated Stats View (`'s'` key) with ntcharts visualizations
- Stage transition tracking for time-per-stage insights

**Methodology Switching Fix:**
- Detect ALL methodologies (not first-match-wins)
- Compare artifact timestamps to determine active methodology
- Tie-breaker: show both with warning on fresh clone

### Key Differentiators

**1. Only Dashboard with Claude Code Log Integration**
No other tool parses Claude Code's JSONL for real-time agent state. This is premium detection unavailable elsewhere.

**2. Instant vs 10-Minute Detection**
Moving from file-activity heuristics to structured log parsing is a 600x improvement in detection latency (10 minutes → <1 second).

**3. Architecture-First Extensibility**
Adapter pattern means adding Cursor/Copilot/Windsurf detection is a new file, not a rewrite. Phase 2 invests in the pattern, not just Claude Code.

**4. Experimental Features Done Right**
Progress Metrics uses separate database and packages—if it doesn't prove value, it's cleanly removable without touching core functionality.

---

## Target Users

### Primary User: Jeff (Evolved)

Phase 2 serves the same Jeff from MVP—but now he's a **power user** who has internalized vdash into his workflow. He's no longer learning the tool; he's pushing its limits.

**Jeff's Phase 2 Reality:**
- vdash is always open in a terminal pane
- He manages 3-6 Claude Code sessions across projects simultaneously
- He's frustrated by the 10-minute lag in knowing when Claude needs him
- He's developed a habit of compulsively checking Claude windows

**The Phase 2 Pain:**
Jeff's morning used to be solved by MVP ("where was I?"). Now his **throughout-the-day** problem is: "Which of my Claude sessions needs me right now?" He's context-switching to check Claude status instead of trusting the dashboard.

**Phase 2 Success for Jeff:**
- Dashboard shows real-time Claude state: Working vs Waiting
- Zero compulsive Claude window checks
- Immediate awareness when any session needs input
- Optional: Insights into his work patterns via Stats View

### Secondary: The Data-Curious Developer

A subset of Jeff who wants to understand their own productivity patterns:
- "How long do I typically spend in each stage?"
- "What days/times am I most productive?"
- "How many stage transitions this week?"

Progress Metrics serves this user—but they're secondary. If this user doesn't materialize, the feature is cleanly removable.

### Non-Users (Explicitly Out of Scope)

- **Cursor/Copilot/Windsurf users:** No structured logs available. They stay on MVP behavior.
- **Teams:** Phase 2 remains single-developer focused.
- **Non-vibe-coding users:** Must use BMAD or Speckit methodology.

---

## Success Metrics

### Core Success: Agent Detection

| Metric | MVP Baseline | Phase 2 Target |
|--------|--------------|----------------|
| Agent state detection latency | 10 minutes | < 1 second |
| Detection accuracy (Claude Code) | N/A | 95%+ |
| Detection confidence indicator | N/A | High/Low shown in UI |

**Jeff Test (Phase 2 Edition):**
- Jeff sees "⏸️ WAITING" indicator within 5 seconds of Claude finishing
- Jeff stops checking Claude windows manually
- Jeff trusts the dashboard enough to focus on other work while Claude runs

### Feature-Specific Metrics

**Sub-1-Minute Agent Detection:**
- [ ] Parses Claude Code JSONL correctly
- [ ] Detects `end_turn` vs `tool_use` accurately
- [ ] Falls back gracefully for non-Claude-Code projects
- [ ] Shows confidence level (High for log-based, Low for heuristic)

**Methodology Switching:**
- [ ] Projects with both BMAD + Speckit detected correctly
- [ ] Most recent artifact wins
- [ ] Fresh clone shows both with warning (tie-breaker)
- [ ] Zero regression on single-methodology projects

**Progress Metrics (Experimental):**
- [ ] Stats View renders without crashing
- [ ] At least one useful visualization ships
- [ ] Feature is removable without breaking core
- [ ] Jeff uses it at least once (low bar—it's experimental)

**Dynamic Binary Name:**
- [ ] Error messages use actual binary name
- [ ] `os.Args[0]` or `RootCmd.Use` implemented

### Business Metrics

| Metric | Target |
|--------|--------|
| GitHub stars growth | +100 from Phase 2 features |
| User retention (existing) | >90% continue using after upgrade |
| New feature adoption | >50% of users try Stats View |
| Bug reports (methodology) | Zero "wrong methodology detected" reports |

### Go/No-Go Criteria

**Ship Phase 2 if:**
- Agent detection latency < 1 second for Claude Code projects
- Methodology switching works for test case (claude-code-log-viewer-cli)
- No regressions in MVP functionality
- Stats View renders (even if basic)

**Delay if:**
- Agent detection accuracy < 90%
- Core dashboard performance degraded
- Methodology switching breaks existing projects

---

## MVP Scope (Phase 2)

### Tier 1: Must Ship (Core)

**Sub-1-Minute Agent Detection for Claude Code**

| Component | Description |
|-----------|-------------|
| `AgentActivityDetector` interface | Port in `internal/core/ports/` |
| `claude_code.go` adapter | Log-based detection in `internal/adapters/agentdetectors/` |
| `generic.go` adapter | File activity fallback (existing behavior) |
| Dashboard integration | Show Working/Waiting/Inactive state per project |
| Confidence indicator | High (log-based) vs Low (heuristic) |

**Implementation Details:**
```go
type AgentState struct {
    Tool        string        // "Claude Code", "Unknown"
    Status      AgentStatus   // Working, WaitingForUser, Inactive
    Duration    time.Duration // How long in current state
    Confidence  Confidence    // High, Low
}
```

**Log Parsing Logic:**
1. Find most recent session: `~/.claude/projects/{project-hash}/*.jsonl`
2. Read last N entries (tail-optimized)
3. Find last `assistant` entry
4. `stop_reason: "end_turn"` → Waiting
5. `stop_reason: "tool_use"` → Working

### Tier 1: Must Ship (Fix)

**Methodology Switching Detection**

| Component | Description |
|-----------|-------------|
| `DetectWithCoexistence()` | New method in registry.go |
| Timestamp comparison | Most recent artifact wins |
| Tie-breaker logic | Show both if within 1 hour |
| TUI display | "Method: speckit / bmad (coexisting)" when tied |

**Selection Logic:**
```go
func (r *Registry) DetectWithCoexistence(ctx, path) ([]*DetectionResult, error) {
    // Run ALL detectors, not first-match-wins
    // Compare artifact timestamps
    // Return single winner or both if tie
}
```

### Tier 1: Must Ship (Polish)

**Dynamic Binary Name**
- Use `os.Args[0]` or `RootCmd.Use` in error messages
- Update version output template
- Keep help examples as canonical "vdash" (documentation consistency)

### Tier 2: Should Ship (Experimental)

**Progress Metrics & Stats View**

| Component | Description |
|-----------|-------------|
| `metrics.db` | Separate database at `~/.vibe-dash/metrics.db` |
| `stage_transitions` table | Track stage changes with timestamps |
| Stats View TUI | Full-screen view via `'s'` key |
| ntcharts integration | Terminal-native visualizations |
| Event collection | Store raw events, aggregate on read |

**Architecture (Loose Coupling):**
```
internal/
├── adapters/metrics/           # Event collection
├── adapters/persistence/metrics/  # metrics.db access
└── adapters/tui/statsview/     # Dedicated TUI view
```

**Minimum Viable Stats View:**
- Activity sparklines per project
- Time-per-stage breakdown (if transitions tracked)
- Date range selector (basic)

**Deferred to Later:**
- Heat maps for work patterns
- Project comparison views
- Advanced date range UI

### Out of Scope (Phase 2)

| Item | Reason |
|------|--------|
| Agent detection for Cursor/Copilot/Windsurf | No structured logs available |
| Plugin architecture for detectors | Future—design when community interest |
| API/HTTP adapter | Architecture ready, implement when needed |
| Performance benchmarks | Create when desired, NFRs defined |
| Windows support | User doesn't have Windows PC |
| Fuzzy search | Low priority, ~3 projects currently |
| `vdash recent` command | Dashboard sufficient for morning routine |

### Implementation Order

1. **Dynamic Binary Name** — Quick win, ship first
2. **Methodology Switching** — Bug fix, unblocks testing
3. **Sub-1-Minute Agent Detection** — Core feature, most effort
4. **Progress Metrics** — Experimental, ship last

### Technical Risks

| Risk | Mitigation |
|------|------------|
| Claude Code log format changes | Version detection, graceful fallback |
| Log file performance (large sessions) | Tail-read optimization, only last N entries |
| Metrics DB growth | Retention policy if needed (estimated ~20MB/year) |
| Stats View complexity | Ship minimal, iterate based on usage |

---

## Phase 2 Timeline Estimate

| Milestone | Scope |
|-----------|-------|
| **Week 1** | Dynamic binary name + Methodology switching fix |
| **Week 2-3** | Agent detection adapter pattern + Claude Code implementation |
| **Week 4** | Stats View foundation + basic visualizations |
| **Week 5** | Polish, testing, documentation |

**Note:** No hard deadlines—ship when ready, quality over speed.

---

## Summary

Phase 2 evolves vdash from "where am I?" to "**what needs me now?**"

The sub-1-minute agent detection for Claude Code is the headline feature—a 600x improvement in awareness latency. Progress Metrics ships as an experiment. Methodology switching gets fixed properly.

Jeff's new reality: Dashboard open, Claude sessions running, instant notification when any agent needs input. Zero compulsive window-checking. Maximum flow state.

**Next Step:** Create PRD from this Product Brief.
