# Epic 3.5 Retrospective: Storage Structure Alignment

**Date:** 2025-12-19
**Facilitator:** Bob (Scrum Master)
**Epic Status:** Complete (10/10 stories)

---

## Epic Summary

| Metric | Value |
|--------|-------|
| **Epic** | Epic 3.5: Storage Structure Alignment |
| **Stories Completed** | 10 of 10 (100%) |
| **Origin** | Discovered during Epic 3 Retrospective |
| **Key Deliverables** | Per-project storage structure, DirectoryManager with collision handling, RepositoryCoordinator, CLI/TUI wiring, removal of .project-path redundancy |

### Stories Delivered

| Story | Title | Status |
|-------|-------|--------|
| 3.5.0 | Cleanup Existing Storage | done |
| 3.5.1 | Directory Manager with Collision Handling | done |
| 3.5.2 | Per-Project SQLite Repository | done |
| 3.5.3 | Per-Project Config Files | done |
| 3.5.4 | Master Config as Path Index | done |
| 3.5.5 | Repository Coordinator | done |
| 3.5.6 | Update CLI Commands | done |
| 3.5.7 | Integration Testing | done |
| 3.5.8 | Fix TUI Repository Wiring | done |
| 3.5.9 | Remove .project-path Redundancy | done |

**Note:** Stories 3.5.8 and 3.5.9 were added after the original plan (3.5.0-3.5.7) due to critical issues discovered post-implementation.

---

## Significant Discovery: The "Bigger Picture" Gap

### Context

After completing Story 3.5.7 (Integration Testing), the epic appeared complete. All tests passed. However, **Jongkuk Lim (Project Lead)** discovered that the TUI dashboard was still using the OLD centralized `projects.db` instead of the new per-project storage.

### What Happened

```
┌─────────────────────────────────────────────────────────────────┐
│  WHAT WE DID                     │  WHAT WE MISSED              │
├──────────────────────────────────┼──────────────────────────────┤
│  ✅ Built DirectoryManager       │  ❌ Verify app uses it       │
│  ✅ Built RepositoryCoordinator  │  ❌ Verify TUI uses it       │
│  ✅ Updated CLI commands         │  ❌ End-to-end verification  │
│  ✅ Passed all unit tests        │  ❌ "Does it actually work?" │
│  ✅ Passed integration tests     │  ❌ Epic-level acceptance    │
└──────────────────────────────────┴──────────────────────────────┘
```

### Root Cause: Inconsistent Code Paths

The bug was masked because `vibe list` and `vibe` (dashboard) used DIFFERENT code paths:

| Command | Code Path | Repository Used | Result |
|---------|-----------|-----------------|--------|
| `vibe list` | `list.go` → injected `repository` | RepositoryCoordinator | ✅ Worked |
| `vibe` (dashboard) | `root.go` → `sqlite.NewSQLiteRepository("")` | OLD centralized DB | ❌ Empty |

**Key Insight from Jongkuk Lim:** "Using the exact same approach with list and dashboard would have prevented this problem earlier."

If both commands had used the same code path, either BOTH would work or BOTH would fail - making the bug immediately visible.

### Impact

- Stories 3.5.8 and 3.5.9 were created to fix the issue
- Scope expanded from 8 stories to 10 stories
- Demonstrated that story-level testing doesn't guarantee epic-level success

---

## Previous Retrospective Follow-Through (Epic 3)

| # | Action Item | Status | Evidence |
|---|-------------|--------|----------|
| 1 | Add "Manual Testing Steps" to story template | ✅ Applied | All Epic 3.5 stories include manual testing sections |
| 2 | **Create Epic 3.5: Storage Structure Alignment** | ✅ COMPLETED | This epic - fully delivered |
| 3 | Consider shared styles package for tui/components | ⏳ Deferred | Not addressed in 3.5 |
| 4 | Create shared test helpers package | ⏳ Deferred | Not addressed in 3.5 |
| 5 | Add architecture diagrams to complex stories | ⏳ Partial | ASCII diagrams in 3.5.5 story |

---

## What Went Well

### 1. Hexagonal Architecture Held Strong
Despite completely rewriting the storage layer, the service layer required ZERO changes. The `RepositoryCoordinator` implements `ports.ProjectRepository` transparently.

### 2. Code Review Effectiveness
Every story went through adversarial code review. Real bugs were caught:
- Story 3.5.5: Delete didn't remove project from config (HIGH severity)
- Story 3.5.8: Identified .project-path redundancy → created 3.5.9

### 3. Discovery-Driven Scope
Stories 3.5.8 and 3.5.9 emerged from real problems found during testing and code review. The process surfaced issues rather than hiding them.

### 4. Clean Implementation Pattern
Consistent patterns across all stories:
- Context cancellation at start of every method
- Graceful degradation with `slog.Warn`
- Empty slice return (not nil) for JSON compatibility
- Table-driven tests

### 5. Manual Testing Steps Adoption
Every story in Epic 3.5 included manual testing instructions, which helped catch the TUI issue.

---

## What Could Be Improved

| # | Issue | Root Cause | Solution |
|---|-------|------------|----------|
| 1 | Epic goal not verified after stories complete | No epic-level acceptance test | Add Epic Acceptance Test |
| 2 | Bug masked by inconsistent code paths | `list` and `dashboard` used different repository access | Single source of truth pattern |
| 3 | Integration tests tested components, not wiring | Tests verified coordinator works, not that app uses it | Automated smoke test |
| 4 | Scope creep (+2 stories) | Issues discovered post-implementation | Earlier end-to-end verification |

---

## Key Insights

### 1. Bigger Picture Ownership
Individual stories can pass while the epic goal fails. Someone must verify the end-to-end goal, not just story acceptance criteria.

**Quote from Jongkuk Lim:** "It's like we are trying our best at our job but nobody was watching the bigger picture."

### 2. Consistency Prevents Hidden Bugs
If two features need the same data, they MUST use the same code path. The inconsistency between `vibe list` and `vibe` dashboard masked the bug - one worked, one didn't, giving false confidence.

### 3. Code Review Catches Bugs, Not Missing Wiring
Code reviews caught implementation bugs but couldn't catch that working code wasn't plugged in. Different verification needed.

### 4. Automated Detection is Possible
A simple smoke test that runs the actual binary and verifies add/list/dashboard consistency would have caught this automatically.

### 5. Scope Creep Through Discovery is Healthy
Stories 3.5.8 and 3.5.9 emerged from real problems. This is the process working correctly - surfacing issues rather than shipping broken code.

---

## Action Items

| # | Action Item | Owner | Priority | Target |
|---|-------------|-------|----------|--------|
| 1 | **Epic Acceptance Test** - Add explicit end-to-end verification to epic template. Must run after final story, before epic marked "done" | SM (Bob) | High | Before Epic 4 |
| 2 | **Single Source of Truth** - All commands must use injected repository. Add to code review checklist: "Does this command use the injected repository?" | Charlie (Senior Dev) | High | Immediate |
| 3 | **Automated Smoke Test** - Build & run binary to verify add/list/dashboard consistency. Test should fail if any command uses different data source | Dev Team (Amelia) | High | Epic 4 Sprint 1 |
| 4 | **Shared test helpers package** (carried from Epic 3) | Dev Team | Medium | Epic 4 |
| 5 | **Shared styles package for TUI** (carried from Epic 3) | Dev Team | Low | Future |

### Epic Acceptance Test Format

```markdown
## Epic Acceptance Test

**Epic Goal:** [One sentence describing the PRIMARY goal]

**Test Steps:**
1. [Setup step]
2. [Action step]
3. [Verification step with expected outcome]
...

**When to Run:** After final story marked "done", BEFORE epic marked "done"
**Who Runs:** Project Lead or SM
```

**Example for Epic 3.5:**
```markdown
## Epic Acceptance Test

**Epic Goal:** Switch from centralized to per-project storage

**Test Steps:**
1. rm -rf ~/.vibe-dash/
2. ./bin/vibe add .
3. ./bin/vibe list → Should show project
4. ./bin/vibe → Dashboard should show project
5. ls ~/.vibe-dash/ → Should show per-project directory, NO projects.db
```

---

## Team Recognition

| Agent | Contribution |
|-------|--------------|
| **Amelia (Dev)** | Implemented all 10 stories with comprehensive tests, quick turnaround on 3.5.8 and 3.5.9 fixes |
| **Charlie (Architect)** | Hexagonal architecture design held strong through major refactoring |
| **Dana (QA)** | Thorough validation, integration test coverage |
| **Elena (Junior Dev)** | Good questions that improved story clarity |
| **Jongkuk Lim (Project Lead)** | Critical observations: caught the TUI wiring issue, identified the "bigger picture" gap, proposed consistency-based prevention |

---

## Next Steps

1. **Immediate:** Update code review checklist with "uses injected repository?" check
2. **Before Epic 4:** Add Epic Acceptance Test template to epic files
3. **Epic 4 Sprint 1:** Implement automated smoke test
4. **Begin Epic 4:** Agent Waiting Detection - now with stable storage foundation

---

## Final Storage Structure

```
~/.vibe-dash/
├── config.yaml                    # Master config with project index (storage_version: 2)
└── <project-name>/                # Per-project directory
    ├── config.yaml                # Project-specific settings
    └── state.db                   # Per-project SQLite database
```

**Key Changes from Pre-Epic 3.5:**
- ❌ No more centralized `projects.db`
- ❌ No more `.project-path` marker files
- ✅ Per-project isolation
- ✅ Config cascade (project → master → defaults)
- ✅ Collision handling via DirectoryManager

---

## Retrospective Sign-off

**Facilitator:** Bob (Scrum Master)
**Date:** 2025-12-19
**Status:** Complete
**Next Epic:** Epic 4 - Agent Waiting Detection

---

**Key Outcomes:**
1. Critical process improvement identified: Epic Acceptance Test
2. Architectural principle reinforced: Same data = same code path
3. Automation opportunity identified: Smoke test for end-to-end verification
4. Epic 3.5 fully delivered with storage structure aligned to PRD specification
5. 5 action items captured (3 new, 2 carried forward)
