# Sprint Artifacts

This directory contains all sprint-related documents for the Vibe Dashboard project, organized by document type and epic.

## Directory Structure

```
sprint-artifacts/
├── README.md                    # This file - guidelines and conventions
├── sprint-status.yaml           # Central status tracking for all epics/stories
├── stories/
│   ├── epic-1/                  # Foundation & First Launch
│   │   └── {story-id}.md
│   ├── epic-2/                  # Project Management & Detection
│   │   └── {story-id}.md
│   └── epic-N/                  # Future epics
│       └── {story-id}.md
├── validations/
│   ├── epic-1/
│   │   └── validation-report-{story-id}.md
│   ├── epic-2/
│   │   └── validation-report-{story-id}.md
│   └── epic-N/
│       └── validation-report-{story-id}.md
└── retrospectives/
    ├── epic-1-retrospective.md
    ├── epic-2-retrospective.md
    └── epic-N-retrospective.md
```

## Document Types

### Stories (`stories/epic-N/`)
- **Purpose:** Complete story specifications with acceptance criteria, tasks, dev notes
- **Naming:** `{epic}-{story}-{slug}.md` (e.g., `2-6-project-name-collision-handling.md`)
- **Created by:** SM Agent (`*create-story`)
- **Status tracking:** Updated through the story lifecycle in `sprint-status.yaml`

### Validation Reports (`validations/epic-N/`)
- **Purpose:** Story draft quality validation before development
- **Naming:** `validation-report-{story-id}.md` or `validation-report-{story-id}-{date}.md`
- **Created by:** SM Agent (`*validate-create-story`)
- **Contains:** Checklist results, issues found, recommendations

### Retrospectives (`retrospectives/`)
- **Purpose:** Epic completion review - what went well, improvements, action items
- **Naming:** `epic-{N}-retrospective.md`
- **Created by:** SM Agent (`*epic-retrospective`)
- **Contains:** Summary, learnings, patterns to carry forward

---

## Workflow Rules

### Rule 1: Create Story → Validate Story

```
*create-story  →  *validate-create-story
```

**Always validate a story draft before marking it ready-for-dev.**

- SM creates story draft with `*create-story`
- SM validates the draft with `*validate-create-story` (preferably in fresh context)
- Validation report saved to `validations/epic-N/`
- Story updated based on validation findings
- Only then mark story as `ready-for-dev`

### Rule 2: Dev Story → Code Review

```
*dev-story  →  *code-review
```

**Always code review after development is complete.**

- Dev implements story with `*dev-story`
- Dev marks story as `review` status
- Reviewer runs adversarial `*code-review`
- Issues categorized by severity (H/M/L)
- Dev fixes issues and re-runs review if needed
- Only then mark story as `done`

### Rule 3: After Code Review → Update Status → Commit

```
*code-review (pass)  →  update sprint-status.yaml  →  git commit
```

**Always update status and commit after completing a story.**

1. Update story status to `done` in `sprint-status.yaml`
2. Commit all changes with descriptive message:
   ```
   feat: Story X.Y - {title}

   - Summary of implementation
   - Key changes made
   - Tests added/modified
   ```
3. Push if appropriate

---

## Story Lifecycle

```
backlog → drafted → ready-for-dev → in-progress → review → done
           ↑                                        ↓
           └────── validation-report ───────────────┘
```

| Status | Description | Updated By |
|--------|-------------|------------|
| `backlog` | Story exists only in epics.md | - |
| `drafted` | Story file created by `*create-story` | SM |
| `ready-for-dev` | Validated and approved for development | SM |
| `in-progress` | Developer actively implementing | Dev |
| `review` | Implementation complete, awaiting review | Dev |
| `done` | Code review passed, committed | Dev |

---

## Epic Lifecycle

```
backlog → in-progress → done
              ↓
    (all stories done)
              ↓
      retrospective
```

| Status | Description |
|--------|-------------|
| `backlog` | Epic defined but not started |
| `in-progress` | At least one story being worked on |
| `done` | All stories complete, retrospective done |

---

## Naming Conventions

### Story Files
```
{epic#}-{story#}-{slug}.md

Examples:
- 1-1-project-scaffolding.md
- 2-6-project-name-collision-handling.md
- 3-1-project-list-component.md
```

### Validation Reports
```
validation-report-{epic#}-{story#}[-{date}].md

Examples:
- validation-report-1-1-project-scaffolding.md
- validation-report-2-3-2025-12-13.md
```

### Retrospectives
```
epic-{N}-retrospective.md

Examples:
- epic-1-retrospective.md
- epic-2-retrospective.md
```

---

## Adding New Epics

When starting a new epic:

1. Create directory: `stories/epic-N/`
2. Create directory: `validations/epic-N/`
3. Update `sprint-status.yaml` with epic and story entries
4. Mark epic as `in-progress`

---

## Quick Reference: SM Commands

| Command | Purpose | Output Location |
|---------|---------|-----------------|
| `*create-story` | Create story draft | `stories/epic-N/` |
| `*validate-create-story` | Validate story draft | `validations/epic-N/` |
| `*sprint-planning` | Generate/update sprint-status.yaml | `sprint-status.yaml` |
| `*epic-retrospective` | Facilitate epic retrospective | `retrospectives/` |

---

## Quick Reference: Dev Commands

| Command | Purpose | Updates |
|---------|---------|---------|
| `*dev-story` | Implement a story | Story file, source code |
| `*code-review` | Review implementation | Story file (fixes applied) |

---

## Maintenance

- Keep `sprint-status.yaml` as single source of truth for status
- Archive completed epic folders if needed (not recommended, keep for reference)
- Retrospective action items should be tracked in subsequent epic planning
