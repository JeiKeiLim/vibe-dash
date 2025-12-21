# Epic 4.6 Proposal: BMAD Stage Detection Validation & Hardening

**Created:** 2025-12-21
**Source:** Epic 4.5 Retrospective
**Priority:** High
**Status:** Proposed

---

## Rationale

During the Epic 4.5 retrospective, we discovered that the BMAD stage detection logic has gaps in handling certain state combinations. Rather than fixing bugs reactively, this epic focuses on **comprehensive validation** to identify ALL potential gaps, update specifications, and implement fixes with full test coverage.

**Why an Epic (not a Hotfix):**
1. Scope is unknown until investigation completes
2. Requires spec updates (Stage Mapping Table)
3. Multiple code changes across files
4. Comprehensive test coverage for all state combinations
5. Stage detection is a core feature - deserves proper attention

---

## Epic Goal

**Ensure BMAD stage detection is accurate and complete for ALL possible epic/story state combinations.**

**User Value:** "When I look at my BMAD project in vibe-dash, the stage displayed accurately reflects where I am in the workflow - no false positives, no missing cases."

---

## Phase 1: Investigation (Story 4.6.1)

### Objective
Identify ALL gaps in current stage detection logic by testing every possible state combination.

### Deliverables
- Complete state combination matrix
- List of all gaps found
- Updated Stage Mapping Table specification
- Decision document for ambiguous cases (drafted, ready-for-dev, etc.)

### Known Gaps (Starting Point)

| ID | Scenario | Current Behavior | Impact |
|----|----------|------------------|--------|
| G1 | Epic in-progress, all stories done | "preparing stories" | HIGH - Misleading |
| G2 | Story status 'drafted' | Falls through | MEDIUM - No visibility |
| G3 | Story status 'ready-for-dev' | Falls through | MEDIUM - No visibility |
| G4 | Multiple stories in-progress | Shows first only | LOW - Incomplete info |

### State Matrix to Validate

**Epic Statuses:** `backlog`, `contexted`, `in-progress`, `done`
**Story Statuses:** `backlog`, `drafted`, `ready-for-dev`, `in-progress`, `review`, `done`

**Minimum Test Cases:**

| # | Epic Status | Story States | Expected Stage | Expected Reasoning |
|---|-------------|--------------|----------------|-------------------|
| 1 | backlog | all backlog | Specify | "No epics in progress" |
| 2 | in-progress | all backlog | Plan | "Epic N started, preparing stories" |
| 3 | in-progress | some drafted | ? | TBD |
| 4 | in-progress | some ready-for-dev | ? | TBD |
| 5 | in-progress | one in-progress | Implement | "Story N.M being implemented" |
| 6 | in-progress | multiple in-progress | ? | TBD |
| 7 | in-progress | one review | Tasks | "Story N.M in code review" |
| 8 | in-progress | one review + one in-progress | ? | TBD |
| 9 | in-progress | all done | ? | **GAP** |
| 10 | in-progress | mixed done + backlog | ? | TBD |
| 11 | done | all done | Implement | "All epics complete" |
| 12 | contexted | all backlog | ? | TBD |

### Questions to Resolve

1. Should 'drafted' stories affect stage display?
2. Should 'ready-for-dev' stories affect stage display?
3. When epic is in-progress but all stories done, what stage?
4. Should we show counts (e.g., "2 of 5 stories in-progress")?
5. How to handle retrospective status?
6. Priority when multiple states exist (review > in-progress > drafted)?

---

## Phase 2: Specification Update (Story 4.6.2)

### Objective
Update the Stage Mapping Table to cover ALL cases identified in Phase 1.

### Deliverables
- Updated Stage Mapping Table in story 4.5.2
- Decision log for all ambiguous cases
- Approval from Product Owner

---

## Phase 3: Implementation (Story 4.6.3)

### Objective
Implement fixes for all gaps identified, matching updated specification.

### Deliverables
- Updated `stage_parser.go` with all cases
- Unit tests for every row in the state matrix
- Integration tests using vibe-dash dogfooding

### Files to Modify
- `internal/adapters/detectors/bmad/stage_parser.go`
- `internal/adapters/detectors/bmad/stage_parser_test.go`
- `docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md`

---

## Phase 4: Verification (Story 4.6.4)

### Objective
Verify stage detection works correctly in real-world usage.

### Deliverables
- Dogfooding test with vibe-dash at multiple stages
- Manual verification walkthrough
- **Binary rebuild and verification** (lesson from Epic 4.5)

---

## Process Improvement

### Binary Rebuild Issue

**Root Cause from Epic 4.5:** Binary built 4+ hours before BMAD detector was committed.

**Proposed Fix:** Add to dev-story or code-review workflow:
```
[ ] Run `make build` after implementation
[ ] Verify feature works with rebuilt binary
```

---

## Dependencies

- None on Epic 5 (Hibernation)
- Can run in parallel with Epic 5

---

## Success Criteria

1. All state combinations in matrix return correct stage and reasoning
2. vibe-dash dogfooding shows accurate stage at any sprint phase
3. Zero false positives or misleading messages
4. 100% test coverage for stage detection logic
5. Binary rebuild verification included in workflow

---

## Estimated Scope

| Story | Effort |
|-------|--------|
| 4.6.1 Investigation | Small |
| 4.6.2 Spec Update | Small |
| 4.6.3 Implementation | Medium |
| 4.6.4 Verification | Small |

**Total:** Medium epic (4 focused stories)

---

## References

- [Source: internal/adapters/detectors/bmad/stage_parser.go]
- [Source: docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md]
- [Source: docs/sprint-artifacts/retrospectives/epic-4.5-retro-2025-12-21.md]
