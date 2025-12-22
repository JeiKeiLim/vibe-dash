# Epic 8: UX Polish

**Status:** Backlog
**Priority:** High - Contains critical core feature gaps
**Created:** 2025-12-22
**Origin:** Systematic UX elicitation session with Project Lead

---

## Epic Overview

### Problem Statement

During a comprehensive UX elicitation session (Options C, B, A methodology), multiple usability gaps were identified. The dashboard is an **observation-only tool** - users don't take actions, they observe and understand project status. Current UX doesn't optimize for glanceability.

**Key Insight:** "I don't take action on this dashboard. I observe and get what I need to understand on what to do next."

### Critical Issues Discovered

1. **Agent waiting detection only works on root directory** - not recursive (killer feature broken)
2. **Refresh doesn't run often enough** - 24/7 use case not supported
3. **Current stage requires 3 steps to find** - open detail → find detection → read

### Elicitation Methodology Used

| Phase | Method | Issues Found |
|-------|--------|--------------|
| C | Mine 7 retrospectives | 5 UX items |
| B | Dogfooding session | 7 issues (incl. P1 critical) |
| A | Feature walkthrough | 6 issues |

---

## Stories

### Story 8.1: Recursive File Watching

**Status:** Backlog
**Priority:** Critical

**As a** user tracking multiple projects,
**I want** file changes in subdirectories to update agent waiting status,
**So that** the killer feature works regardless of where I'm editing files.

**Problem:** Currently, creating files in subdirectories doesn't trigger agent waiting status updates. Only root directory changes are detected.

**Acceptance Criteria:**

```gherkin
AC1: Given a project at ~/projects/my-app
     When I create/modify a file at ~/projects/my-app/src/main.go
     Then the project's last_activity_at timestamp updates

AC2: Given a project with subdirectory activity
     When the waiting threshold passes since last subdirectory change
     Then the project shows "waiting" status

AC3: Given recursive watching is enabled
     When monitoring performance is measured
     Then CPU usage remains under 5% for projects with 1000+ files
```

**Tasks:**
- [ ] Investigate current fsnotify watcher configuration
- [ ] Enable recursive directory watching
- [ ] Test performance with large project directories
- [ ] Add integration tests for subdirectory detection

**Manual Testing:**
1. Run `./bin/vibe`
2. Create file in project subdirectory: `touch ~/project/src/test.txt`
3. Verify last_activity_at updates in database
4. Wait for threshold, verify waiting status appears

---

### Story 8.2: Auto-Refresh Reliability

**Status:** Backlog
**Priority:** Critical

**As a** user running vibe 24/7,
**I want** the dashboard to refresh automatically and reliably,
**So that** I always see current status without manual refresh.

**Problem:** Refresh doesn't run often enough. Users report needing to press 'r' manually.

**Acceptance Criteria:**

```gherkin
AC1: Given vibe is running for 1+ hours
     When no user interaction occurs
     Then status updates continue appearing automatically

AC2: Given auto-refresh is running
     When a project's waiting state changes
     Then the UI updates within 5 seconds

AC3: Given the tick interval
     When waiting counts are recalculated
     Then tickMsg handler triggers status bar update
```

**Tasks:**
- [ ] Review tickMsg handler for status recalculation
- [ ] Verify tick interval is appropriate for 24/7 use
- [ ] Add waiting count recalculation to tick cycle
- [ ] Test long-running session reliability

**Manual Testing:**
1. Run `./bin/vibe --waiting-threshold=1`
2. Touch a file in a tracked project
3. Wait 2+ minutes without interaction
4. Verify waiting indicator appears without pressing 'r'

---

### Story 8.3: Stage Info in Project List Row

**Status:** Backlog
**Priority:** Critical

**As a** user glancing at the dashboard,
**I want** to see current stage (spec/epic/story) directly in the project list,
**So that** I don't need to open detail view to understand where I am.

**Problem:** Current list shows only "Tasks", "Unknown", "Waiting 10m" which is not informative. Users need to: open detail → find detection section → read stage. Three steps for primary information.

**Current Display:**
```
  vibe-dash        BMAD     Tasks                    ● working
```

**Target Display:**
```
  vibe-dash        BMAD   Epic 4.6, Story 4.6.4 in review   ● working
```

**Acceptance Criteria:**

```gherkin
AC1: Given a BMAD project with active epic/story
     When viewing the project list
     Then I see "Epic X.Y, Story X.Y.Z status" in the row

AC2: Given a Speckit project with active spec
     When viewing the project list
     Then I see "Spec NNN, Stage name" in the row

AC3: Given stage text is longer than available space
     When rendering the row
     Then text is truncated with ellipsis, not wrapped

AC4: Given the stage info display
     When terminal width changes
     Then stage info adapts (full → shortened → hidden)
```

**Tasks:**
- [ ] Design shortened stage format for each method type
- [ ] Update delegate render to include stage column
- [ ] Handle width constraints and truncation
- [ ] Add responsive breakpoints for stage display

**Manual Testing:**
1. Run `./bin/vibe`
2. Verify BMAD projects show epic/story info in list
3. Verify Speckit projects show spec/stage in list
4. Resize terminal, verify graceful degradation

---

### Story 8.4: Fix Layout Width Bugs

**Status:** Backlog
**Priority:** High

**As a** user,
**I want** consistent full-width layout,
**So that** the dashboard uses all available terminal space.

**Problem:** On launch, there's margin on left/right. After refresh, it goes full width. Detail panel also has margin issues.

**Acceptance Criteria:**

```gherkin
AC1: Given vibe is launched
     When the initial render completes
     Then the layout uses full terminal width (no margins)

AC2: Given detail panel is opened
     When the panel renders
     Then it uses full available width (no margins)

AC3: Given any terminal resize
     When re-render occurs
     Then full width is maintained
```

**Tasks:**
- [ ] Debug initial render width calculation
- [ ] Fix detail panel width constraints
- [ ] Ensure WindowSizeMsg handling is consistent
- [ ] Add visual regression tests if possible

**Manual Testing:**
1. Run `./bin/vibe`
2. Check for margins on initial launch
3. Press 'r' to refresh, compare width
4. Press 'd' to open detail, check for margins

---

### Story 8.5: Favorites Sort First

**Status:** Backlog
**Priority:** High

**As a** user with favorite projects,
**I want** favorites to appear at the top of the list,
**So that** my most important projects are immediately visible.

**Problem:** Favorites are not sorted to the top. Sort order unclear.

**Acceptance Criteria:**

```gherkin
AC1: Given projects with some marked as favorites
     When viewing the project list
     Then favorites appear before non-favorites

AC2: Given multiple favorites
     When viewing the list
     Then favorites are sorted alphabetically among themselves

AC3: Given non-favorites
     When viewing the list
     Then non-favorites are sorted alphabetically after all favorites
```

**Tasks:**
- [ ] Review current sort implementation
- [ ] Add favorites-first sort priority
- [ ] Maintain alphabetical within each group
- [ ] Update tests for sort order

**Manual Testing:**
1. Add 3+ projects, favorite 1-2
2. Run `./bin/vibe`
3. Verify favorites appear at top
4. Verify alphabetical order within groups

---

### Story 8.6: Horizontal Split Layout Option

**Status:** Backlog
**Priority:** Medium

**As a** user who prefers horizontal layouts,
**I want** to configure detail panel as top/bottom split instead of left/right,
**So that** I can see both project list and full-width detail simultaneously.

**Problem:** Current left/right split reduces project list width. Users want option for top/bottom.

**Acceptance Criteria:**

```gherkin
AC1: Given config `detail_layout: horizontal`
     When I press 'd' to toggle detail
     Then detail panel appears below project list

AC2: Given config `detail_layout: vertical` (default)
     When I press 'd' to toggle detail
     Then detail panel appears to the right (current behavior)

AC3: Given horizontal layout
     When detail is visible
     Then both project list and detail use full terminal width
```

**Tasks:**
- [ ] Add `detail_layout` config option
- [ ] Implement horizontal split renderer
- [ ] Update layout calculations for horizontal mode
- [ ] Test with various terminal sizes

**Manual Testing:**
1. Set `detail_layout: horizontal` in config
2. Run `./bin/vibe`
3. Press 'd', verify detail appears below
4. Verify full-width rendering for both panels

---

### Story 8.7: Config Display in TUI

**Status:** Backlog
**Priority:** Medium

**As a** user,
**I want** to see current configuration values in the TUI,
**So that** I know what thresholds and settings are active.

**Problem:** Users don't know what waiting threshold or other config values are active.

**Acceptance Criteria:**

```gherkin
AC1: Given the help overlay (?)
     When I view it
     Then I see current config values (threshold, etc.)

AC2: Given a project is selected
     When I view detail panel
     Then I see project-specific config if different from global

AC3: Given config display
     When values shown
     Then both global and project-level overrides are indicated
```

**Tasks:**
- [ ] Design config display location (help overlay? status bar? detail?)
- [ ] Extract current effective config values
- [ ] Render config section with clear labels
- [ ] Show project vs global distinction

**Manual Testing:**
1. Run `./bin/vibe`
2. Press '?' for help
3. Verify config values are displayed
4. Check project-specific overrides in detail view

---

### Story 8.8: Clarify or Remove Confidence Display

**Status:** Backlog
**Priority:** Medium

**As a** user,
**I want** detection confidence to be meaningful or removed,
**So that** I'm not confused by unexplained labels.

**Problem:** "Certain", "Likely" confidence labels are confusing. Users don't know:
- What it refers to (method? stage?)
- What the values mean
- What action to take based on confidence

**Options:**
1. **Clarify:** Add tooltip/help explaining confidence
2. **Improve:** Make confidence actionable ("Low confidence - check manually")
3. **Remove:** If not actionable, remove clutter

**Acceptance Criteria:**

```gherkin
AC1: Given confidence is kept
     When displayed
     Then meaning is clear (tooltip, label, or help text)

AC2: Given confidence is removed
     When viewing detection info
     Then only actionable information remains

AC3: Given the decision
     When implemented
     Then help overlay documents the choice
```

**Tasks:**
- [ ] Decide: clarify, improve, or remove
- [ ] Implement chosen approach
- [ ] Update help text accordingly
- [ ] Test user comprehension

---

### Story 8.9: Visual Polish Bundle

**Status:** Backlog
**Priority:** Low

**As a** user,
**I want** small visual improvements,
**So that** the dashboard feels polished.

**Scope:**
- P12: Focus indicator '>' feels collapsible - consider alternative
- P13: Project count redundant with status bar
- P14: Status bar decoration improvements
- P15: Time display in status bar (optional)
- P16: Custom status bar patterns (optional, post-MVP candidate)
- P17: Terminal emoji compatibility

**Acceptance Criteria:**

```gherkin
AC1: Given focus indicator
     When reviewed
     Then a clearer indicator is used or documented as intentional

AC2: Given project count in list header
     When evaluated
     Then redundancy with status bar is resolved

AC3: Given emoji usage
     When running in limited terminals
     Then fallback characters are used
```

**Tasks:**
- [ ] Review and decide on each P12-P17 item
- [ ] Implement quick wins
- [ ] Defer complex items to post-MVP
- [ ] Document decisions

---

## Post-MVP Items (Captured but Deferred)

| # | Item | Rationale |
|---|------|-----------|
| PM1 | Favorites own section | Power-user feature |
| PM2 | Epic 5 Hibernation | Deferred per 2025-12-22 decision |
| PM3 | Search projects | <10 projects manageable |
| PM4 | Progress graphs/comparison | Requires data collection first |
| PM5 | Copy/paste project info | Nice-to-have |
| PM6 | OpenTelemetry integration | Future architecture |
| PM7 | Config editing in TUI | Complex, view-only for now |

---

## Epic Acceptance Test

**Epic Goal:** Improve dashboard glanceability for observation-only use case

**Test Steps:**
1. Run `./bin/vibe`
2. Verify full-width layout on initial launch (no margins)
3. Verify favorites sorted to top
4. Verify stage info visible in project list rows
5. Touch file in project subdirectory, verify activity detected
6. Wait for threshold, verify waiting status appears without manual refresh
7. Press 'd', verify detail panel full-width
8. Press '?', verify config values shown

**When to Run:** After final story marked "done", BEFORE epic marked "done"
**Who Runs:** Project Lead

---

## Sign-off

| Role | Name | Status |
|------|------|--------|
| Product Owner | - | Pending |
| Scrum Master | Bob | Drafted |
| Project Lead | Jongkuk Lim | Pending |

---

*Created: 2025-12-22*
*Origin: Systematic UX elicitation (C→B→A methodology)*
