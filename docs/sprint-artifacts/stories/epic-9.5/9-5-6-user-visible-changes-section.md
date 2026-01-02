# Story 9.5.6: User-Visible Changes Section

Status: done

**Priority: Low**

## Story

As a **developer reviewing completed stories**,
I want **a "User-Visible Changes" section in each story file that summarizes what the user will notice after implementation**,
So that **I can quickly communicate changes to stakeholders and prepare release notes without reading technical details**.

## Background

**Origin:** Epic 8 Retrospective → Action Item P2 (carried forward 4 times)

The process improvement P2 "User-Visible Changes section in stories" has been carried forward through 4 retrospectives:
- Epic 7 → P2 created (first mention)
- Epic 8 → P2 not addressed, carried forward (2nd time)
- Epic 9 → P2 not addressed, carried forward (4th time)
- Epic 9.5 → Formalized as Story 9.5-6

### Problem Statement

Currently, story files contain technical implementation details (tasks, dev notes, code locations) but no quick summary of what the user will actually see or experience after the story is completed. This causes:

1. **Release note friction** - Stakeholders must parse technical details to understand user impact
2. **Review overhead** - Reviewers must infer user-facing changes from code changes
3. **Knowledge gaps** - Future developers don't know what changed from the user's perspective
4. **Testing confusion** - QA doesn't have a clear list of what to verify from user POV

### Desired State

Every story file includes a `## User-Visible Changes` section near the top that lists:
- What the user will see differently (UI changes, new commands, output format changes)
- What the user will experience differently (performance, behavior changes)
- What the user WON'T notice (internal refactors, code cleanup) - explicit "none" if applicable

### Examples

**Good Example (Story 8-3: Stage Info in Project List):**
```markdown
## User-Visible Changes

- **New:** Project list now shows Epic/Story info in the "Stage" column (e.g., "Epic 4 / Story 4-3")
- **Changed:** Stage column is wider (40% of width instead of 20%)
- **Removed:** Confidence percentage no longer appears in the list view
```

**Good Example (Story 7-10: Shared Test Helpers - Internal):**
```markdown
## User-Visible Changes

None - this is an internal code refactoring. Users will not notice any difference.
```

## Acceptance Criteria

### AC1: Update Story Template
- Given the story template at `.bmad/bmm/workflows/4-implementation/create-story/template.md`
- When a developer views the template
- Then there is a `## User-Visible Changes` section after the `## Story` section
- And the template includes placeholder text explaining what to include

### AC2: Section Position
- Given a story file with `## User-Visible Changes` section
- When a reviewer opens the file
- Then the section appears BEFORE `## Acceptance Criteria`
- And the section appears AFTER `## Story`

### AC3: Internal Change Handling
- Given a story that has no user-visible changes (internal refactor)
- When the developer fills in the section
- Then they write "None - [brief reason]" (not left blank)
- And the template guidance explains this pattern

### AC4: Checklist Update
- Given the story validation checklist at `.bmad/bmm/workflows/4-implementation/create-story/checklist.md`
- When a validator runs the checklist
- Then there is a validation item for "User-Visible Changes section is present and not empty"

### AC5: Backward Compatibility
- Given existing story files without the section
- When the new template is applied to future stories
- Then existing stories are NOT modified (no mass updates)
- And only new stories created after this change include the section

### AC6: Documentation Update
- Given this change is a process improvement
- When the story is complete
- Then `docs/project-context.md` is updated with guidance on the section
- And the Story Completion section mentions the User-Visible Changes requirement

## Tasks / Subtasks

- [x] Task 1: Update story template (AC: 1, 2, 3)
  - [x] 1.1: Edit `.bmad/bmm/workflows/4-implementation/create-story/template.md`
  - [x] 1.2: Add `## User-Visible Changes` section after `## Story` section
  - [x] 1.3: Add placeholder text with examples (UI, behavior, "None" pattern)

- [x] Task 2: Update validation checklist (AC: 4)
  - [x] 2.1: Edit `.bmad/bmm/workflows/4-implementation/create-story/checklist.md`
  - [x] 2.2: Add validation item for User-Visible Changes section

- [x] Task 3: Update project context (AC: 6)
  - [x] 3.1: Edit `docs/project-context.md`
  - [x] 3.2: Add User-Visible Changes requirement to Story Completion section

- [x] Task 4: Create this story's User-Visible Changes section (dogfooding) (AC: 1)
  - [x] 4.1: Add the section to this story file as the first example

## User-Visible Changes

None - this is a process improvement to BMAD Method story templates. Existing vibe-dash functionality is unchanged. Users of vibe-dash will not notice any difference.

**For BMAD Method users:**
- **New:** Story template now includes `## User-Visible Changes` section
- **New:** Validation checklist now checks for this section

## Dev Notes

### Previous Story Learnings

**From Story 9.5-5 (Pipeline Summary Output):**
- Process improvements don't require code changes to vibe-dash itself
- Template/documentation changes are low-risk and quick to implement

**From Epic 8 Retrospective:**
- P2 was identified because developers had to read technical details to understand user impact
- The observation handoff issue (P3) is related but separate - focuses on testing guidance

### Implementation Details

**Template Change (`.bmad/bmm/workflows/4-implementation/create-story/template.md`):**

Section inserted at lines 11-22 (after `so that {{benefit}}.` and before `## Acceptance Criteria`):
```markdown
## User-Visible Changes

<!-- List what the user will notice after this story is complete -->
<!-- Use bullet points with New/Changed/Removed prefixes -->
<!-- If no user-facing changes, write: None - [brief reason why] -->

- **New:** [Feature/behavior users didn't have before]
- **Changed:** [Existing feature that works differently now]
- **Removed:** [Feature/behavior no longer available]

<!-- Or for internal changes: -->
None - this is an internal [refactoring/cleanup/infrastructure] change.
```

**Checklist Addition (`.bmad/bmm/workflows/4-implementation/create-story/checklist.md`):**

The checklist is a quality competition prompt with structured sections. Add to the "Disaster Prevention Gap Analysis" section (Step 3) under a new category:

```markdown
#### 3.6 User-Visible Changes Verification

- **Section presence:** Is `## User-Visible Changes` section present after `## Story`?
- **Content completeness:** Does it have at least one New/Changed/Removed item OR explicit "None - [reason]"?
- **Empty section:** Is it left blank without explanation? (FAIL if so)
```

**Project Context Addition (`docs/project-context.md`):**

Add after line 217 (after "What User Checks" table) before "Post-MVP References":
```markdown
### User-Visible Changes Requirement

Every story MUST include a `## User-Visible Changes` section that describes:
- What users will see/experience differently (New/Changed/Removed)
- Or explicitly "None - [reason]" for internal changes

This section:
- Enables quick release note generation
- Helps reviewers focus on user impact
- Documents historical changes for future developers

**Reference example:** See this story file itself (Story 9.5-6) lines 112-118 for a correctly formatted section.

### Boundaries & Anti-Patterns

| Boundary | Details |
|----------|---------|
| **In scope** | BMAD template, checklist, project-context.md |
| **Out of scope** | Existing story files (no mass updates), vibe-dash code |

| Don't | Do Instead |
|-------|------------|
| Update all existing story files | Only affect new stories |
| Make section optional | Require explicit "None" for internal changes |
| Put section at end of file | Place it early (after Story, before AC) |

### References

| Document | Relevance |
|----------|-----------|
| `docs/sprint-artifacts/retrospectives/epic-8-retro-2025-12-31.md` | P2 origin |
| This story's own `## User-Visible Changes` section (line 112) | Dogfooding reference |

## User Testing Guide

**Time needed:** 2 minutes

### Step 1: Verify Template Updated
```bash
cat .bmad/bmm/workflows/4-implementation/create-story/template.md
```
- **Expected:** See `## User-Visible Changes` section after `## Story`
- **Red flag:** Section missing or in wrong position

### Step 2: Verify Checklist Updated
```bash
grep -i "user-visible" .bmad/bmm/workflows/4-implementation/create-story/checklist.md
```
- **Expected:** Match showing validation item for User-Visible Changes
- **Red flag:** No match found

### Step 3: Verify Project Context Updated
```bash
grep -i "user-visible" docs/project-context.md
```
- **Expected:** Match showing requirement documentation
- **Red flag:** No match found

### Decision Guide

| Situation | Action |
|-----------|--------|
| All 3 files updated with User-Visible Changes content | Mark `done` |
| Any file missing the section | FAIL - update missing file |
| Section exists but no helpful placeholder text | FAIL - add guidance text |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-9.5/9-5-6-user-visible-changes-section.md`
- Previous story: `docs/sprint-artifacts/stories/epic-9.5/9-5-5-pipeline-summary-output.md`
- Project context: `docs/project-context.md`
- Template reference: `.bmad/bmm/workflows/4-implementation/create-story/template.md`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - No debug logs needed for this documentation-only story.

### Completion Notes List

- All 4 tasks completed - documentation-only story, no code changes
- Template updated with `## User-Visible Changes` section after `## Story` with placeholder guidance
- Checklist updated with section 3.6 for User-Visible Changes verification
- Project context updated with requirement documentation
- Story already had User-Visible Changes section (lines 112-118) - dogfooding complete

### File List

| File | Change |
|------|--------|
| `.bmad/bmm/workflows/4-implementation/create-story/template.md` | Modified - add User-Visible Changes section |
| `.bmad/bmm/workflows/4-implementation/create-story/checklist.md` | Modified - add validation item |
| `docs/project-context.md` | Modified - add requirement documentation |
| `docs/sprint-artifacts/sprint-status.yaml` | Modified - story status update |

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-02 | SM (Bob) | Initial story creation via *create-story workflow (YOLO mode) |
| 2026-01-02 | SM (Bob) | Validation improvements: (C1) Fixed template line number reference; (E1) Clarified section positioning; (E2) Updated checklist addition to match actual file structure; (E3) Added self-reference as example; (L1-L3) Removed redundant Testing Strategy, consolidated References |
| 2026-01-02 | Dev (Amelia) | Implementation complete: All 4 tasks done, template/checklist/project-context updated |
| 2026-01-02 | Dev (Amelia) | Code review: 4M/3L issues found and fixed - M1 template placeholders clarified, M2 checklist numbering fixed, M3 line reference removed, M4 dev notes updated to reflect implementation |
