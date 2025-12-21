# Story 4.6.2: Spec Update - Stage Mapping Table

Status: done

## Story

As a developer maintaining vibe-dash,
I want the Stage Mapping Table specification updated to cover ALL state combinations identified in Story 4.6.1,
so that Story 4.6.3 implementation has a complete, unambiguous specification to implement against.

## Acceptance Criteria

1. **Complete Stage Mapping Table:** Given the 22 gaps identified in Story 4.6.1, when this story completes, then the Stage Mapping Table in Story 4.5.2 is updated to include:
   - All 23+ cases from the Investigation Results
   - Error cases (E1-E4)
   - Inconsistent state cases (rows 14-16)
   - LLM typo normalization rules (rows 19-20)
   - Data quality warning cases (rows 21-23)

2. **Decision Log Transferred:** Given the 7 decisions made in Story 4.6.1, when this story completes, then each decision is incorporated into the Stage Mapping Table with:
   - The chosen option clearly stated
   - Rationale documented inline or in dev notes
   - No ambiguity about expected behavior

3. **Priority Levels Defined:** Given the 22 gaps categorized in Story 4.6.1, when this story completes, then the spec includes:
   - P1 (Must Fix): G1, G7, G15 with full specifications
   - P2 (Should Fix): G2, G3, G8, G14, G17, G19, G22 with specifications
   - P3 (Nice to Have): remaining gaps with specifications

4. **Normalization Rules Specified:** Given the LLM typo handling requirement (G15), when this story completes, then the spec includes:
   - Complete list of status variations to normalize
   - Mapping table from variation → canonical value
   - Case handling rules

5. **Test Matrix Updated:** Given the expanded Stage Mapping Table, when this story completes, then the Test Matrix section includes test cases for:
   - All P1 gaps
   - All P2 gaps
   - Representative P3 gaps

6. **Story 4.6.3 Ready:** Given spec updates complete, when this story completes, then Story 4.6.3 can be implemented without ambiguity by:
   - Reading ONLY Story 4.5.2 (with updates) and Story 4.6.3
   - Not requiring re-reading of Story 4.6.1 investigation results

## Tasks / Subtasks

- [x] Task 1: Update Stage Mapping Table in Story 4.5.2 (AC: #1, #2, #6)
  - **Target File:** `docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md`
  - [x] 1.1 Replace existing Stage Mapping Table with complete 23+ row table from Story 4.6.1 (copy verbatim from "Updated Stage Mapping Table" section)
  - [x] 1.2 Add Error Cases section (E1-E4)
  - [x] 1.3 Add Inconsistent State Cases section (rows 14-16)
  - [x] 1.4 Add LLM Normalization Cases section (rows 19-20)
  - [x] 1.5 Add Data Quality Warning Cases section (rows 21-23)
  - **Verify:** Story 4.5.2 contains complete table with 23+ rows covering all cases

- [x] Task 2: Add Status Normalization Specification (AC: #4)
  - [x] 2.1 Create new "Status Normalization" section in Story 4.5.2 (after Stage Mapping Table)
  - [x] 2.2 Copy normalizeStatus() function example from Dev Notes below (lines 153-180)
  - [x] 2.3 Reference complete mapping table in Status Normalization Mapping section below (do not duplicate)
  - [x] 2.4 Document case normalization rule (all lowercase before comparison)
  - **Verify:** Story 4.5.2 has Status Normalization section with Go code example and mapping table

- [x] Task 3: Add Gap Priority Classification (AC: #3)
  - [x] 3.1 Add "Implementation Priority" section to Story 4.5.2 Dev Notes
  - [x] 3.2 Copy P1/P2/P3 gap classification from Gap Priority Reference below
  - [x] 3.3 Preserve gap IDs (G1-G22) exactly as documented in Story 4.6.1 - do not renumber
  - **Verify:** Story 4.5.2 Dev Notes has Implementation Priority section with P1/P2/P3 classification

- [x] Task 4: Update Test Matrix (AC: #5)
  - [x] 4.1 Locate or create "Test Matrix" section in Story 4.5.2 (create if not exists)
  - [x] 4.2 Add test cases for G1 (all stories done in in-progress epic)
  - [x] 4.3 Add test cases for G7 (epic done, stories active)
  - [x] 4.4 Add test cases for G15 (LLM typo variations)
  - [x] 4.5 Add test cases for G2/G3 (drafted/ready-for-dev)
  - [x] 4.6 Add test cases for inconsistent states (G8)
  - [x] 4.7 Add test cases for normalization
  - **Verify:** Test Matrix section has at least 10 new test case rows covering P1/P2 gaps

- [x] Task 5: Create Story 4.6.3 Skeleton (AC: #6)
  - **Target File:** `docs/sprint-artifacts/stories/epic-4.6/4-6-3-implementation-fix-all-gaps.md`
  - [x] 5.1 Create story file with standard structure:
    - Status: backlog
    - Story statement (As a developer... I want... so that...)
    - Acceptance Criteria (reference updated Stage Mapping Table in Story 4.5.2)
    - Tasks organized by P1/P2/P3 priority from this story
    - Dev Notes referencing Story 4.5.2 as PRIMARY spec
  - [x] 5.2 List files to modify: `stage_parser.go`, `stage_parser_test.go`
  - [x] 5.3 Add explicit note: "Do NOT re-read Story 4.6.1 - all specifications are in Story 4.5.2"
  - **Verify:** Story 4.6.3 file exists with complete structure and references Story 4.5.2 only

- [x] Task 6: Final Verification (AC: #6)
  - [x] 6.1 Read Story 4.5.2 + Story 4.6.3 only and verify implementation is unambiguous
  - [x] 6.2 Update sprint-status.yaml: `4-6-2-spec-update-stage-mapping: done`
  - **Verify:** Can implement Story 4.6.3 by reading ONLY Story 4.5.2 without Story 4.6.1

## Dev Notes

### This is a SPECIFICATION ONLY story - no code changes

**Primary deliverable:** Updates to Story 4.5.2 documentation only.

**Output artifacts:**
1. Updated `docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md`
2. Created `docs/sprint-artifacts/stories/epic-4.6/4-6-3-implementation-fix-all-gaps.md` (skeleton)

### Source Material (DO NOT RE-INVESTIGATE)

All findings are complete in Story 4.6.1 (`docs/sprint-artifacts/stories/epic-4.6/4-6-1-investigation-gap-analysis.md`):
- **State Matrix:** Section "Investigation Results" → "Task 1.3: Base Cross-Product Matrix"
- **Gap Analysis:** Section "Gap Analysis Results" → "Task 4: Complete Gap List"
- **Decision Log:** Section "Decision Log" (7 decisions: D1-D7)
- **Updated Stage Mapping Table:** Section "Updated Stage Mapping Table"

**CRITICAL:** Do NOT re-investigate. Copy content from Story 4.6.1 section headers listed above.

### Complete Stage Mapping Table (Copy from Story 4.6.1)

| # | Condition | Stage | Confidence | Reasoning Template |
|---|-----------|-------|------------|-------------------|
| **Error Cases** |
| E1 | sprint-status.yaml missing | (fallback) | - | Falls back to artifact detection |
| E2 | sprint-status.yaml malformed | Unknown | Uncertain | "sprint-status.yaml parse error" |
| E3 | sprint-status.yaml empty | Unknown | Uncertain | "sprint-status.yaml is empty" |
| E4 | development_status key missing | Unknown | Uncertain | "No development_status section" |
| **Epic-Level Cases** |
| 1 | All epics backlog | Specify | Certain | "No epics in progress - planning phase" |
| 2 | All epics done | Implement | Certain | "All epics complete - project done" |
| 3 | Mixed: some done, none in-progress | Specify | Certain | "No active epic - planning next" |
| **Story-Level Cases (Epic In-Progress)** |
| 4 | No stories | Plan | Certain | "Epic N started, preparing stories" |
| 5 | All stories backlog | Plan | Certain | "Epic N started, preparing stories" |
| 6 | Has drafted stories only | Plan | Certain | "Story X.Y drafted, awaiting approval" |
| 7 | Has ready-for-dev stories only | Plan | Certain | "Story X.Y ready for development" |
| 8 | Has in-progress story | Implement | Certain | "Story X.Y being implemented" |
| 9 | Has review story | Tasks | Certain | "Story X.Y in code review" |
| 10 | Has in-progress AND review | Tasks | Certain | "Story X.Y in code review" |
| **11** | **All stories done** | **Implement** | **Certain** | **"Epic N stories complete, update epic status"** |
| **Multi-Story Cases** |
| 12 | Multiple in-progress | Implement | Certain | "Story X.Y being implemented (+N more)" |
| 13 | Multiple review | Tasks | Certain | "Story X.Y in code review (+N more)" |
| **Inconsistent State Cases (NEW)** |
| 14 | Epic done, stories in-progress | Implement | Likely | "Epic done but Story X.Y in-progress" |
| 15 | Epic done, stories in review | Tasks | Likely | "Epic done but Story X.Y in review" |
| 16 | Epic backlog, stories active | Specify | Likely | "Epic backlog but Story X.Y active" |
| **Unknown Status Cases** |
| 17 | Unknown epic status | Unknown | Uncertain | "Unknown epic status 'VALUE'" |
| 18 | Unknown story status | (continue) | - | "Unknown story status 'VALUE'" |
| **LLM Typo Normalization (NEW - G15)** |
| 19 | Status with spaces/underscores | (normalize) | - | Normalize "in progress" → "in-progress" |
| 20 | Synonym mapping | (normalize) | - | Map "complete"→"done", "wip"→"in-progress" |
| **Data Quality Cases (NEW - G14, G18-G22)** |
| 21 | Orphan story (no epic match) | (warn) | Likely | "Story X.Y has no matching epic" |
| 22 | Deep story prefix (4-5-6-xxx) | (warn) | Likely | "Story prefix depth exceeds epic depth" |
| 23 | Empty status value `""` | (warn) | Likely | "Empty status for key X" |
| **Fallback Artifact Detection** |
| F1 | Has epic*.md | Implement | Likely | "Epics defined but no sprint status" |
| F2 | Has architecture*.md | Plan | Likely | "Architecture designed, no epics yet" |
| F3 | Has prd*.md | Specify | Likely | "PRD created, architecture pending" |
| F4 | No artifacts | Unknown | Uncertain | "No BMAD artifacts detected" |

### Status Normalization Mapping (Copy from Story 4.6.1)

```go
// normalizeStatus converts common variations to canonical status values.
// Apply BEFORE switch statement comparison.
func normalizeStatus(status string) string {
    // 1. Lowercase everything
    s := strings.ToLower(strings.TrimSpace(status))

    // 2. Normalize separators: spaces and underscores → hyphens
    s = strings.ReplaceAll(s, " ", "-")
    s = strings.ReplaceAll(s, "_", "-")

    // 3. Map synonyms
    synonyms := map[string]string{
        "complete":    "done",
        "completed":   "done",
        "finished":    "done",
        "wip":         "in-progress",
        "inprogress":  "in-progress",
        "reviewing":   "review",
        "in-review":   "review",
        "code-review": "review",
    }

    if canonical, ok := synonyms[s]; ok {
        return canonical
    }
    return s
}
```

### Story Status Priority Order

When multiple stories have different statuses, use this priority (highest first):
1. `review` - someone waiting for feedback
2. `in-progress` - active development
3. `ready-for-dev` - queued for development
4. `drafted` - story being prepared
5. `backlog` - not started
6. `done` - already completed

### Gap Priority Reference (from Story 4.6.1)

| Priority | Gap IDs | Description |
|----------|---------|-------------|
| **P1 (Must Fix)** | G1, G7, G15 | Actively misleading users OR high LLM probability |
| **P2 (Should Fix)** | G2, G3, G8, G14, G17, G19, G22 | User sees less helpful info OR data quality issues |
| **P3 (Nice to Have)** | G4, G5, G6, G9, G10, G11, G12, G13, G16, G18, G20, G21 | Edge cases, UX polish |

### Decision Summary (from Story 4.6.1)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| D1: `drafted` display | Show "Story X.Y drafted, awaiting approval" | User should know which story is being refined |
| D2: `ready-for-dev` display | Plan "Story X.Y ready for development" | Not yet implementing, but work is ready |
| D3: Multi-status priority | Review > in-progress | Review is higher-urgency action |
| D4: Retrospective handling | Ignore | Retros don't block development |
| D5: Inconsistent states | Warn in reasoning string | Help catch data entry errors |
| D6: Empty epic | "preparing stories" | Correct - genuinely preparing |
| D7: All stories done | "Epic N stories complete, update epic status" | Guide user to update sprint-status.yaml |

### Files to Modify

| File | Action | Notes |
|------|--------|-------|
| `docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md` | UPDATE | Add complete Stage Mapping Table, normalization spec |
| `docs/sprint-artifacts/stories/epic-4.6/4-6-3-implementation-fix-all-gaps.md` | CREATE | Implementation story skeleton |
| `docs/sprint-artifacts/sprint-status.yaml` | UPDATE | Mark 4-6-2 status → done when complete |

### Story 4.5.2 Section Order (for updates)

Add/update sections in this order within Story 4.5.2:
1. **Stage Mapping Table** (replace existing) - 23+ rows
2. **Status Normalization** (new section) - Go code example + mapping table
3. **Story Status Priority Order** (new section) - 6-item priority list
4. **Test Matrix** (new or update section) - P1/P2 test cases
5. **Implementation Priority** (add to Dev Notes) - P1/P2/P3 gap classification
6. **Decision Summary** (add to Dev Notes) - 7 decisions from Story 4.6.1

### What If Guidance

| Situation | Action |
|-----------|--------|
| Story 4.5.2 too long | Split into sections with clear headers; keep Stage Mapping Table as primary |
| Ambiguity discovered | Check Story 4.6.1 Decision Log → if not resolved, document and propose |
| New gap discovered | Add G-TBD to Story 4.6.1 retroactively; include in Story 4.6.3 scope |

### References

| Document | Role | Path |
|----------|------|------|
| **Story 4.6.1** | PRIMARY - All specs come from here | `docs/sprint-artifacts/stories/epic-4.6/4-6-1-investigation-gap-analysis.md` |
| Story 4.5.2 | Target file to update | `docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md` |
| Epic 4.6 Proposal | Epic scope context | `docs/sprint-artifacts/stories/epic-4.5/epic-4.6-proposal.md` |
| Project Context | Conventions | `docs/project-context.md` |

## Dev Agent Record

### Context Reference

- Story 4.6.1 (Investigation): `docs/sprint-artifacts/stories/epic-4.6/4-6-1-investigation-gap-analysis.md`
- Story 4.5.2 (Target): `docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- This is a specification-only story - no code changes or test runs required
- All content transferred from Story 4.6.1 Investigation Results

### Completion Notes List

1. **Task 1 Complete:** Updated Stage Mapping Table in Story 4.5.2 with complete 23+ row table including:
   - Error Cases (E1-E4)
   - Epic-Level Cases (1-3)
   - Story-Level Cases (4-11)
   - Multi-Story Cases (12-13)
   - Inconsistent State Cases (14-16)
   - Unknown Status Cases (17-18)
   - LLM Typo Normalization (19-20)
   - Data Quality Warning Cases (21-23)
   - Fallback Artifact Detection (F1-F4)

2. **Task 2 Complete:** Added Status Normalization section with:
   - Complete normalizeStatus() Go function
   - Case normalization rule
   - 10-row normalization examples table

3. **Task 3 Complete:** Added Implementation Priority section with:
   - P1/P2/P3 gap classification
   - 22 gap details table
   - Decision Summary (7 decisions D1-D7)

4. **Task 4 Complete:** Expanded Test Matrix with 28 new test cases:
   - 8 P1 gap test cases (G1, G7, G15)
   - 8 P2 gap test cases (G2, G3, G8, G14, G17, G19, G22)
   - 2 inconsistent state test cases
   - 10 normalization test cases

5. **Task 5 Complete:** Created Story 4.6.3 skeleton with:
   - Standard story structure
   - 8 tasks organized by P1/P2/P3 priority
   - Explicit note: "DO NOT RE-READ STORY 4.6.1"
   - References Story 4.5.2 as PRIMARY SPEC

6. **Task 6 Complete:** Final verification confirms Story 4.6.3 can be implemented by reading ONLY Story 4.5.2

### File List

- `docs/sprint-artifacts/stories/epic-4.5/4-5-2-bmad-v6-stage-detection-logic.md` (UPDATED - Stage Mapping Table, Status Normalization, Implementation Priority, Decision Summary, Test Matrix)
- `docs/sprint-artifacts/stories/epic-4.6/4-6-3-implementation-fix-all-gaps.md` (CREATED - Story skeleton with P1/P2/P3 tasks)
- `docs/sprint-artifacts/sprint-status.yaml` (UPDATED - 4-6-2-spec-update-stage-mapping: done)
