# Epic 3 Retrospective: Dashboard Visualization

**Date:** 2025-12-18
**Facilitator:** Bob (Scrum Master)
**Epic Status:** Complete (10/10 stories)

---

## Epic Summary

| Metric | Value |
|--------|-------|
| **Epic** | Epic 3: Dashboard Visualization |
| **Stories Completed** | 10 of 10 (100%) |
| **Key Deliverables** | Full TUI dashboard with project list, detail panel, status bar, keyboard shortcuts, help overlay, manual refresh, project notes, favorites, remove project, responsive layout |

### Stories Delivered

| Story | Title | Status |
|-------|-------|--------|
| 3.1 | Project List Component | done |
| 3.2 | Keyboard Navigation | done |
| 3.3 | Detail Panel Component | done |
| 3.4 | Status Bar Component | done |
| 3.5 | Help Overlay | done |
| 3.6 | Manual Refresh | done |
| 3.7 | Project Notes (View & Edit) | done |
| 3.8 | Toggle Favorite | done |
| 3.9 | Remove Project from TUI | done |
| 3.10 | Responsive Layout | done |

---

## Significant Discovery: Storage Structure Misalignment

### Context

During the retrospective, **Jongkuk Lim (Project Lead)** identified that the current implementation's storage structure does not match the PRD specification. This was not captured in any existing epic or story.

### PRD Specification (lines 597-604)

The PRD specifies a **per-project subdirectory structure**:

```
~/.vibe-dash/
  ├── config.yaml                 # Master config (single source of truth)
  ├── api-service/
  │   ├── config.yaml             # Project-specific settings
  │   └── state.db                # Per-project SQLite database
  └── client-b-api-service/
      ├── config.yaml
      └── state.db
```

**Key PRD References:**
- FR39: Store project path mappings in centralized master config (`~/.vibe-dash/config.yaml`)
- FR40: Store project-specific settings in project config files (`~/.vibe-dash/<project>/config.yaml`)
- Line 166: "Projects tracked centrally with per-project SQLite databases"
- Line 624: Project config at `~/.vibe-dash/<project>/config.yaml`
- Line 841: Per-project SQLite database at `~/.vibe-dash/<project>/state.db`

### Current Implementation

The actual implementation uses a **single centralized database**:

```
~/.vibe-dash/
  ├── config.yaml                 # Master config ✅ Correct
  └── projects.db                 # Single centralized DB ❌ Deviation
```

**Implementation References:**
- `internal/adapters/persistence/sqlite/repository.go:53`: `dbPath = filepath.Join(home, ".vibe-dash", "projects.db")`
- `cmd/vibe/main.go:54`: `sqlite.NewSQLiteRepository("")` uses default centralized path

### What's Missing

1. **Per-project subdirectories** (`~/.vibe-dash/<project-name>/`)
2. **Per-project SQLite databases** (`state.db` per project)
3. **Per-project config files** (`config.yaml` per project)
4. **Project-name collision handling** for directory names (parent directory disambiguation)
5. **Configuration cascade** (project config overrides master config)

### Impact Assessment

| Area | Impact |
|------|--------|
| **Data Isolation** | All projects share one DB - no isolation |
| **Project-specific Settings** | Cannot set per-project hibernation thresholds, display names via config |
| **Scalability** | Single DB may become bottleneck with many projects |
| **Backup/Restore** | Cannot backup/restore individual project state |
| **Epic 4 Dependency** | Agent Waiting Detection will write more state data - building on wrong structure makes later migration harder |

### Decision

**Create Epic 3.5: Storage Structure Alignment**

- Must complete BEFORE Epic 4 (Agent Waiting Detection)
- Requires thorough planning with Architect
- Needs data migration strategy for existing users
- Small, safe stories with rollback capability

### Planning Requirements

1. **Architect Review** - Winston to review PRD storage spec in detail
2. **Gap Analysis** - Full documentation of current vs expected state
3. **Migration Strategy** - How to safely move existing `projects.db` data to per-project structure
4. **Story Breakdown** - Small, independently deployable stories
5. **Rollback Plan** - Recovery path if migration fails
6. **Test Coverage** - Comprehensive tests for migration path

---

## Action Item: Manual Testing Steps in Stories

### Context

**Jongkuk Lim (Project Lead)** observed that during Epic 3, the workflow was:
- Create story → Validate story → Dev story → Code review
- No manual verification of actual functionality was performed

The trust in automated tests was high, but there was no easy way to manually verify features work as expected in the actual application.

### Problem

Stories lack clear "how to test this manually" instructions, making it difficult for:
- Project Lead to spot-check functionality
- QA to perform exploratory testing
- New team members to understand expected behavior

### Solution

Add a **"Manual Testing Steps"** section to the story template with:
- Clear, numbered steps
- Expected outcomes for each step
- Prerequisites (test data, environment setup)

### Example Format

```markdown
## Manual Testing

**Prerequisites:** At least 2 projects added via `vibe add`

1. Run `vibe` to launch dashboard
2. Press 'n' to open note editor
3. **Expected:** Dialog appears with title "Edit note for <project-name>"
4. Type "Test note content" and press Enter
5. **Expected:** Dialog closes, status bar shows "✓ Note saved"
6. Press 'd' to view detail panel
7. **Expected:** Notes section shows "Test note content"
```

### Owner

SM (Bob) - Update story template for Epic 3.5 and beyond

---

## Previous Retrospective Action Items (from Epic 2)

| # | Action Item | Status | Evidence |
|---|-------------|--------|----------|
| 1 | Wire detection service to TUI for real-time updates | ⏳ Deferred | Planned for Epic 4 |
| 2 | Consider smaller story breakdown for complex features | ✅ Applied | Epic 3 stories were well-scoped |
| 3 | Run detector edge case analysis earlier in story prep | ✅ Applied | Validation reviews caught issues early |

---

## What Went Well

### 1. Shared Package Extraction Pattern
The extraction of `internal/shared/timeformat/` and `internal/shared/project/` in Story 3.1 prevented code duplication across CLI and TUI adapters. This pattern should continue for cross-adapter utilities.

### 2. Code Review Loop Effectiveness
Every story went through adversarial review with specific H/M/L severity issues. Real bugs were caught - like `SetSize` not propagating delegate width in Story 3.1, DRY violations, and missing test coverage.

### 3. Bubbles Component Pattern Maturity
By mid-epic, the team had clear understanding of the Bubble Tea pattern: message types → commands → Update handler → View renderer. Story structure became predictable and repeatable.

### 4. Responsive Layout Thoroughness
Story 3.10 handled edge cases beyond basic requirements: narrow terminals (60-79), wide terminals (>120), short terminals (<20), medium terminals (20-34). Real-world terminal variations are now covered.

### 5. Team Trust & Autonomy
Project Lead could approve work confidently without micromanaging, indicating mature team process, clear deliverables, and reliable story completion.

---

## What Could Be Improved

| # | Issue | Potential Solution | Priority |
|---|-------|-------------------|----------|
| 1 | Manual testing steps missing in stories | Add section to story template | High ✅ *Actioned* |
| 2 | Storage structure deviation from PRD | Create Epic 3.5 | Critical ✅ *Actioned* |
| 3 | Import cycle workarounds (tui/components) | Consider shared styles package | Low |
| 4 | Repetitive test mock setup across files | Create shared test helpers package | Medium |
| 5 | Bubbles delegate pattern learning curve | Add architecture diagrams to complex stories | Low |
| 6 | Cross-referencing multiple docs for context | Add "Quick Context" summary to stories | Low |

---

## Key Insights

### 1. Component Pattern Maturity
The Bubbles/Bubble Tea patterns are now well-established. Message types → Commands → Update → View flow is repeatable and understood by the team.

### 2. Shared Package Strategy Works
Extracting `shared/timeformat` and `shared/project` prevented duplication. This pattern should continue for cross-adapter utilities.

### 3. Code Review Catches Real Bugs
The adversarial review process found actual issues (H1/H2 severity) in multiple stories. Worth the time investment.

### 4. Responsive Design Requires Explicit Testing
Story 3.10 taught us that terminal size edge cases (narrow, wide, short, tall) need explicit test coverage - they don't "just work."

### 5. Spec Compliance Needs Verification
The storage structure gap shows we need periodic "spec alignment checks" - not just feature completion, but architectural compliance.

---

## Action Items for Next Epic

| # | Action Item | Owner | Priority |
|---|-------------|-------|----------|
| 1 | Add "Manual Testing Steps" section to story template | SM (Bob) | High |
| 2 | **Create Epic 3.5: Storage Structure Alignment** | Architect (Winston) | **CRITICAL** |
| 3 | Consider shared styles package for tui/components | Dev Team | Low |
| 4 | Create shared test helpers package | Dev Team | Medium |
| 5 | Add architecture diagrams to complex stories | SM (Bob) | Low |

---

## Team Recognition

| Agent | Contribution |
|-------|--------------|
| **Amelia (Dev)** | Implemented all 10 stories with comprehensive tests, consistent quality throughout |
| **Alice (PM)** | Clear acceptance criteria, caught UX edge cases in requirements |
| **Charlie (Architect)** | Hexagonal patterns held strong, component architecture guidance |
| **Dana (QA)** | Thorough validation reviews, test coverage focus |
| **Elena (Junior Dev)** | Quick learner, good questions that improved documentation |
| **Jongkuk Lim (Project Lead)** | Trust in team, critical observation on storage structure and manual testing gaps - both significant improvements |

---

## Next Steps

1. **Immediate:** Schedule Architect session for Epic 3.5 planning
2. **Before Epic 4:** Complete Epic 3.5 implementation
3. **Ongoing:** Apply manual testing template to all new stories

---

## Retrospective Sign-off

**Facilitator:** Bob (Scrum Master)
**Date:** 2025-12-18
**Status:** Complete
**Next Epic:** Epic 3.5 - Storage Structure Alignment (must complete before Epic 4)

---

**Key Outcomes:**
1. Two significant process improvements identified (manual testing, spec compliance checks)
2. Critical architectural gap discovered (storage structure) - Epic 3.5 created
3. Team patterns validated (code review, shared packages, component architecture)
4. 5 action items captured for continuous improvement
