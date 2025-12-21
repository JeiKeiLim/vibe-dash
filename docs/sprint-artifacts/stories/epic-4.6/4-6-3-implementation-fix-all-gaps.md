# Story 4.6.3: Implementation - Fix All Gaps

Status: backlog

## Story

As a developer using vibe-dash,
I want all 22 stage detection gaps fixed as specified in Story 4.5.2,
so that stage detection accurately reflects my project state in all scenarios.

## Acceptance Criteria

1. **P1 Gaps Fixed:** Given the P1 gaps (G1, G7, G15), when implementation completes, then:
   - G1: Epic in-progress with all stories done shows "Epic N stories complete, update epic status"
   - G7: Epic done with active stories shows warning with actual story status
   - G15: LLM typo variations (spaces, underscores, synonyms) are normalized correctly

2. **P2 Gaps Fixed:** Given the P2 gaps (G2, G3, G8, G14, G17, G19, G22), when implementation completes, then:
   - G2/G3: `drafted` and `ready-for-dev` stories display appropriately
   - G8: Epic backlog with active stories shows warning
   - G14: Orphan stories (no matching epic) are warned
   - G17: Status synonyms are normalized
   - G19: Story order is deterministic (first by sorted key)
   - G22: Empty status values are warned

3. **Status Normalization:** Given the normalizeStatus() function from Story 4.5.2 spec, when implemented, then all test cases in the Normalization Test Cases table pass

4. **Test Coverage:** Given the Test Matrix in Story 4.5.2, when tests are written, then:
   - All P1 gap test cases pass
   - All P2 gap test cases pass
   - All Normalization test cases pass
   - Existing tests continue to pass

5. **No Regressions:** Given the existing test suite, when all fixes are applied, then `go test ./internal/adapters/detectors/bmad/...` passes with 100% existing test cases

## Tasks / Subtasks

### P1 Priority (Must Fix)

- [ ] Task 1: Fix G1 - All Stories Done Detection (AC: #1)
  - [ ] 1.1 Add `allStoriesDone` flag in determineStageFromStatus()
  - [ ] 1.2 Check if all stories in epic have status `done`
  - [ ] 1.3 Return StageImplement "Epic N stories complete, update epic status"
  - [ ] 1.4 Write test case: epic in-progress, all stories done
  - [ ] 1.5 Write test case: multiple epics, one with all stories done
  - **Verify:** `go test ./internal/adapters/detectors/bmad/... -run "G1|AllStoriesDone"`

- [ ] Task 2: Fix G7 - Epic Done with Active Stories (AC: #1)
  - [ ] 2.1 After all-epics-done check, scan for inconsistent story states
  - [ ] 2.2 If epic done but story in-progress, return StageImplement with warning
  - [ ] 2.3 If epic done but story in review, return StageTasks with warning
  - [ ] 2.4 Write test case: epic done, story in-progress
  - [ ] 2.5 Write test case: epic done, story in review
  - **Verify:** `go test ./internal/adapters/detectors/bmad/... -run "G7|EpicDone.*Active"`

- [ ] Task 3: Fix G15 - LLM Typo Normalization (AC: #1, #3)
  - [ ] 3.1 Implement normalizeStatus() function per Story 4.5.2 spec
  - [ ] 3.2 Apply normalizeStatus() to epic status before switch
  - [ ] 3.3 Apply normalizeStatus() to story status before switch
  - [ ] 3.4 Write unit tests for normalizeStatus() (10 cases from Normalization Test Cases)
  - [ ] 3.5 Write integration tests with typo variations
  - **Verify:** `go test ./internal/adapters/detectors/bmad/... -run "Normalize|Typo"`

### P2 Priority (Should Fix)

**Note:** Task 3 (G15 normalization) must complete before Tasks 4-7, as normalization is required for those test cases to pass correctly.

- [ ] Task 4: Fix G2/G3 - Drafted and Ready-for-Dev (AC: #2)
  - [ ] 4.1 Add case for `drafted` in story status switch
  - [ ] 4.2 Add case for `ready-for-dev` in story status switch
  - [ ] 4.3 Implement Story Status Priority Order (review > ip > rfd > drafted > backlog > done)
  - [ ] 4.4 Write test cases for drafted-only and ready-for-dev-only scenarios
  - **Verify:** `go test ./internal/adapters/detectors/bmad/... -run "Drafted|ReadyForDev"`

- [ ] Task 5: Fix G8 - Epic Backlog with Active Stories (AC: #2)
  - [ ] 5.1 After backlog epic detection, scan for active stories
  - [ ] 5.2 Return StageSpecify with warning about inconsistent state
  - [ ] 5.3 Write test cases for backlog epic with in-progress/done stories
  - **Verify:** `go test ./internal/adapters/detectors/bmad/... -run "G8|EpicBacklog.*Active"`

- [ ] Task 6: Fix G19 - Deterministic Story Order (AC: #2)
  - [ ] 6.1 Sort story keys before iteration
  - [ ] 6.2 Use first match by sorted order (not map iteration order)
  - [ ] 6.3 Write test case with multiple stories verifying order
  - **Verify:** `go test ./internal/adapters/detectors/bmad/... -run "StoryOrder"`

- [ ] Task 7: Fix G14/G22 - Data Quality Warnings (AC: #2)
  - [ ] 7.1 Track orphan stories (story prefix doesn't match any epic)
  - [ ] 7.2 Track empty status values
  - [ ] 7.3 Include warnings in reasoning string (don't fail detection)
  - [ ] 7.4 Write test cases for orphan stories and empty status
  - **Verify:** `go test ./internal/adapters/detectors/bmad/... -run "Orphan|EmptyStatus"`

### Final Verification

- [ ] Task 8: Full Test Suite Verification (AC: #4, #5)
  - [ ] 8.1 Run full test suite: `go test ./internal/adapters/detectors/bmad/... -v`
  - [ ] 8.2 Run linter: `golangci-lint run ./internal/adapters/detectors/bmad/...`
  - [ ] 8.3 Verify 100% existing tests still pass
  - [ ] 8.4 Count new test cases added (target: 20+ from Test Matrix)
  - [ ] 8.5 Update sprint-status.yaml: `4-6-3-implementation-fix-all-gaps: done`
  - **Verify:** All tests pass, no lint errors

## Dev Notes

### IMPORTANT: DO NOT RE-READ STORY 4.6.1

All specifications are in **Story 4.5.2** (`docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md`).

Story 4.6.1 is the investigation story - its findings have been transferred to Story 4.5.2.

### Files to Modify

| File | Action | Notes |
|------|--------|-------|
| `internal/adapters/detectors/bmad/stage_parser.go` | MODIFY | Add normalizeStatus(), fix gap logic |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | MODIFY | Add 20+ new test cases |
| `docs/sprint-artifacts/sprint-status.yaml` | UPDATE | Mark story done when complete |

### Key Sections in Story 4.5.2

1. **Stage Mapping Table** - Complete expected behavior for all 23+ cases
2. **Status Normalization** - normalizeStatus() code to copy
3. **Story Status Priority Order** - Priority for multi-status scenarios
4. **Implementation Priority** - Gap classification and details
5. **Decision Summary** - Design decisions with rationale
6. **Test Matrix** - All test cases to implement

### Implementation Order Rationale

P1 before P2 because:
- G1 (all stories done) is most visible to vibe-dash users
- G7 (epic done, stories active) is actively misleading
- G15 (LLM typos) has highest real-world frequency

### What If Guidance

| Situation | Action |
|-----------|--------|
| New gap discovered during implementation | Add G-TBD to Story 4.6.1, document in completion notes |
| Test case not in matrix | Add to matrix retroactively, implement test |
| Spec ambiguity found | Check Story 4.5.2 Decision Summary, propose if not covered |

### References

| Document | Role | Path |
|----------|------|------|
| **Story 4.5.2** | PRIMARY SPEC | `docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md` |
| Project Context | Conventions | `docs/project-context.md` |
| stage_parser.go | Implementation target | `internal/adapters/detectors/bmad/stage_parser.go` |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

<!-- Will be filled by dev agent -->

### Debug Log References

<!-- Will be filled by dev agent -->

### Completion Notes List

<!-- Will be filled by dev agent -->

### File List

<!-- Will be filled by dev agent -->
