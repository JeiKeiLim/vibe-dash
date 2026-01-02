# Epic 9.5 Retrospective - Stability & BMAD Update

**Date:** 2026-01-02
**Facilitator:** Bob (Scrum Master)
**Participants:** Alice (Product Owner), Charlie (Senior Dev), Dana (QA Engineer), Elena (Junior Dev), Jongkuk Lim (Project Lead)

---

## Epic Summary

| Metric | Value |
|--------|-------|
| **Epic** | Epic 9.5: Stability & BMAD Update |
| **Goal** | Fix file watcher instability, update BMAD detector, resolve carried-forward items |
| **Stories Completed** | 7/7 (100%) |
| **Origin** | Epic 9 retrospective - issues discovered after Story 8-13 fix |
| **Key Insight** | "Cleanup epics that address accumulated issues are valuable" |

### Stories Delivered

| Story | Title | Key Achievement |
|-------|-------|-----------------|
| 9.5-1 | File Watcher Stability Investigation | Root cause: race condition in select{} between context cancel and channel close |
| 9.5-2 | File Watcher Error Handling | 500ms grace period fix for transient restart errors |
| 9.5-3 | BMAD Directory Structure Update | Support `.bmad`, `_bmad`, `_bmad-output` directories |
| 9.5-4 | Pre-existing Test Failures Cleanup | Skip-by-default pattern for flaky golden tests |
| 9.5-5 | Pipeline Summary Output | `scripts/summary.sh` with colored test/build summaries |
| 9.5-6 | User-Visible Changes Section | Story template + checklist updated |
| 9.5-7 | Stage Detection for Backlog Epics | `analyzeEpicStories()` helper, backlog epic support |

---

## Previous Retrospective Follow-Up (Epic 9)

| # | Action Item | Status | Notes |
|---|-------------|--------|-------|
| C1 | Create Epic 9.5 in sprint-status.yaml | ✅ Done | Epic completed with 7 stories |
| C2 | Investigate file watcher stability | ✅ Done | Stories 9.5-1, 9.5-2 |
| C3 | Research BMAD v6 directory changes | ✅ Done | Story 9.5-3 |
| P1 | Pipeline summary output | ✅ **Done** | Story 9.5-5 - Finally resolved after 4 carries |
| P2 | User-Visible Changes section | ✅ **Done** | Story 9.5-6 - Finally resolved after 4 carries |
| P3 | Structured observation handoff | ✅ **Dropped** | Replaced by Epic 9 golden tests + manual verification |

**Result:** 5/6 action items completed, 1 dropped (no longer needed).

---

## What Went Well

### 1. Investigation-First Approach
- Story 9.5-1 properly traced the race condition before attempting fixes
- Led to clean, targeted solution in Story 9.5-2 (grace period check)
- No rework or pivots needed

### 2. Tight Feedback Loop
- User-reported issues (file watcher blinking, BMAD detection) verified directly by user
- Reduced back-and-forth iterations
- User knows exactly what "fixed" should look like

### 3. Finally Resolved Carried-Forward Items
- P1 (Pipeline summary) and P2 (User-Visible Changes) were carried forward 4 times
- Formalizing them as stories (9.5-5, 9.5-6) forced resolution
- Both now complete

### 4. DRY Refactoring in Code Review
- Story 9.5-7 code review identified duplicate code
- Refactored to `analyzeEpicStories()` helper, removed ~70 lines
- Improves maintainability

### 5. Bug Fix During Retrospective
- Discovered integration tests broken since Story 3.5.9 (Dec 19)
- Fixed immediately during retro session
- CI now green with `-tags=integration`

---

## What Didn't Go Well

### 1. Pre-existing Integration Test Failures
- Story 3.5.9 removed `.project-path` marker files
- Integration tests weren't fully updated
- Broken for 2+ weeks without detection
- **Root cause:** Integration tests not running in CI by default

### 2. Golden Test Flakiness
- Skip-by-default pattern works but feels like tech debt
- Tests require `GOLDEN_TESTS=1` to run
- May miss regressions if not run regularly

---

## Issues Fixed During Retrospective

### Integration Test Failures (Commit c1a9bd6)

**Problem:** 8 integration tests failing due to obsolete `.project-path` marker assertions

**Files Fixed:**
- `directory_integration_test.go` - Remove marker assertions, update determinism/symlink tests
- `project_config_integration_test.go` - Remove marker setup, delete obsolete test
- `project_repository_integration_test.go` - Remove marker check, delete obsolete test
- `coordinator_integration_test.go` - Remove marker creation in mock/helper

**Resolution:** All integration tests now pass

---

## Action Items

### Critical Path

| # | Action Item | Owner | Status |
|---|-------------|-------|--------|
| C1 | Verify CI runs integration tests | Dev | To verify |
| C2 | Mark Epic 9.5 complete | SM | ✅ Done (this retro) |

### Process Improvements - Final Status

| # | Action Item | Status | Resolution |
|---|-------------|--------|------------|
| P1 | Pipeline summary output | ✅ Done | Story 9.5-5 |
| P2 | User-Visible Changes section | ✅ Done | Story 9.5-6 |
| P3 | Structured observation handoff | ✅ Dropped | Replaced by Epic 9 golden tests + manual verification |

---

## Lessons Learned

1. **Integration tests need CI coverage** - Tests were broken for 2+ weeks without detection because they only run with `-tags=integration`

2. **Cleanup epics are valuable** - Epic 9.5 resolved 4x carried-forward items plus 3 newly discovered issues

3. **Investigation-first pays off** - Story 9.5-1 investigation enabled clean, targeted fix in 9.5-2

4. **User as validator works well** - Tight feedback loop between user-reported issues and user verification prevented wasted iterations

5. **Process improvements get done when formalized** - P1/P2 were carried forward 4 times as action items, but resolved immediately when formalized as stories

---

## Epic Prioritization Decision

**Decision:** Defer Epic 10, proceed to Epic 11

| Epic | Decision | Rationale |
|------|----------|-----------|
| 10 (Scalable File Watching) | **Deferred** | Current architecture is stable after 8.13, 9.5-1, 9.5-2 fixes |
| 11 (Project Hibernation) | **Next** | User-facing feature, MVP complete, ready for post-MVP features |

---

## Retrospective Metrics

| Metric | Value |
|--------|-------|
| Stories Completed | 7/7 (100%) |
| Stories Added Mid-Epic | 0 |
| Previous Action Items Resolved | 5/6 (P3 dropped) |
| New Action Items | 2 (C1 to verify, C2 done) |
| Bug Fixed During Retro | 1 (integration tests - commit c1a9bd6) |
| Commits During Retro | 1 |

---

## Epic Sequence Update

| Epic | Name | Status |
|------|------|--------|
| 1-8 | MVP Epics | Done |
| 9 | TUI Behavioral Testing Infrastructure | Done |
| 9.5 | Stability & BMAD Update | **Done** |
| 10 | Scalable File Watching | Backlog (deferred) |
| **11** | **Project Hibernation** | **Next** |

---

*Retrospective facilitated by Bob (Scrum Master)*
*Document generated: 2026-01-02*
