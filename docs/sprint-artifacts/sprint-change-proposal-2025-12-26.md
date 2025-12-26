# Sprint Change Proposal: Epic 8 Story Additions

**Date:** 2025-12-26
**Triggered By:** User feedback during Story 8.4 code review
**Scope:** Minor - Direct story additions to existing epic
**Approved:** Pending

---

## 1. Issue Summary

During Story 8.4 (Fix Layout Width Bugs) code review, two UX gaps were identified:

| Issue | Problem | Impact |
|-------|---------|--------|
| **MaxContentWidth cap** | 120-column cap wastes space on wide monitors; column proportions unbalanced (name too wide, stage too narrow) | Users feel they're wasting screen space |
| **Stage detection not auto-refreshing** | File watcher updates waiting status but stage info only updates on manual [r] | Users expect stage info to stay current automatically |

Both issues align with Epic 8's goal: "Improve dashboard glanceability for observation-only use case."

---

## 2. Impact Analysis

### Epic Impact
- **Epic 8 (UX Polish):** Add 2 new stories (8-10, 8-11)
- **Other Epics:** No impact
- **Epic order/priority:** No change

### Artifact Conflicts
- **PRD:** No conflicts - features are UX improvements within existing scope
- **Architecture:** No conflicts - changes are TUI adapter layer only
- **UI/UX:** Enhances existing patterns, no new paradigms

### Technical Impact
- Files affected: `model.go`, `views.go`, `project_item_delegate.go`, `config/`
- No database changes
- No API changes
- Backward compatible (defaults preserve current behavior)

---

## 3. Recommended Approach

**Path:** Direct Adjustment (Option 1)

| Factor | Assessment |
|--------|------------|
| Effort | Low-Medium |
| Risk | Low |
| Timeline Impact | None - fits within Epic 8 |
| Rationale | Stories fit naturally into existing epic theme |

---

## 4. Detailed Change Proposals

### Story 8-10: Full-Width Layout & Column Rebalancing

**Priority:** High

**Problem:**
1. MaxContentWidth=120 wastes space on wide monitors
2. Column proportions unbalanced: name too wide, stage/status too narrow

**Solution:**
- Add `max_content_width` config option (0=disabled, N=cap, default 120)
- Rebalance column proportions: name ~25%, stage ~40%, status ~20%
- Add per-column max widths for ultra-wide terminals

**Acceptance Criteria Summary:**
- AC1-2: Rebalanced columns at normal widths
- AC3-5: Configurable max content width
- AC6: Column max widths prevent absurd stretching

---

### Story 8-11: Periodic Stage Re-Detection

**Priority:** High

**Problem:** Stage detection only runs on manual [r] refresh, while waiting status auto-updates

**Solution:**
- Add periodic stage detection timer (default 30s, configurable)
- Reuses existing detection service
- Timer resets on manual refresh

**Design Decision:** Periodic timer chosen over smart file-watching because:
- Method-agnostic (works for BMAD, Speckit, future methods)
- No hardcoded file patterns per detection method
- Simple to implement and configure

**Acceptance Criteria Summary:**
- AC1-3: Configurable interval (default 30s, 0=disabled)
- AC4-5: Updates UI, resets on manual refresh
- AC6: Batch detection for all projects

---

## 5. Implementation Handoff

**Scope Classification:** Minor

**Route To:** Development team (direct implementation)

**Deliverables:**
- [x] Story 8-10 drafted and approved
- [x] Story 8-11 drafted and approved
- [ ] Stories added to sprint-status.yaml
- [ ] Story files created in epic-8 folder
- [ ] Epic file updated with new stories

**Implementation Order:**
1. Story 8-10 first (layout foundation)
2. Story 8-11 second (independent feature)

**Success Criteria:**
- Both stories pass code review
- Manual testing confirms features work as specified
- Backward compatibility maintained (defaults preserve current behavior)

---

## Sign-off

| Role | Name | Status |
|------|------|--------|
| Project Lead | Jongkuk Lim | Pending |
| Dev Agent | Amelia | Drafted |

---

*Generated: 2025-12-26*
*Workflow: correct-course*
