# Epic Acceptance Checklist

**Purpose:** Prevent "all stories done but feature doesn't work" scenario.
**When:** Before marking any epic as complete.
**Source:** Epic 4 Retrospective (2025-12-20) - Action Item H2

---

## Pre-Completion Checklist

Before marking an epic complete, verify ALL of the following:

### 1. Technical Completion

- [ ] All stories in the epic are marked "done"
- [ ] All tests pass (`go test ./...`)
- [ ] Build succeeds (`go build -o bin/vibe ./cmd/vibe`)
- [ ] No critical bugs in backlog related to this epic

### 2. End-to-End User Verification

**CRITICAL:** This is where we've failed twice (Epic 3.5, Epic 4).

- [ ] **Fresh perspective test:** Can a user who didn't write the code discover and use the feature?
- [ ] **Zero-state visibility:** Is the feature visible even when in "empty" or "zero" state?
  - Example: "0 waiting" should be shown, not hidden
  - Example: Empty list should show helpful guidance
- [ ] **Happy path works:** The main use case works end-to-end
- [ ] **Feature is observable:** User can verify the feature is working (not silent)

### 3. Documentation Check

- [ ] README updated if new commands or features added
- [ ] Help text (`--help`) is accurate
- [ ] Any new configuration options are documented

---

## Epic-Specific Verification

For each epic, define specific scenarios to verify:

### Epic 4: Agent Waiting Detection

- [ ] Open TUI, modify a file in tracked project, verify `last_activity_at` updates
- [ ] Wait threshold time, verify WAITING indicator appears
- [ ] Modify file, verify WAITING indicator clears
- [ ] Status bar shows waiting count (even if 0)

### Epic 5: Hibernation (Template)

- [ ] Project auto-hibernates after threshold days
- [ ] Activity in hibernated project auto-activates it
- [ ] Hibernation state is visible in UI
- [ ] `vibe list` shows hibernation status

---

## Failure Examples (Learn From Past)

### Epic 3.5 Failure
- **Symptom:** `vibe add` worked, but TUI didn't show the project
- **Root Cause:** TUI wasn't wired to use the new repository
- **Lesson:** End-to-end testing required, not just unit tests

### Epic 4 Failure
- **Symptom:** All stories done, but user couldn't see WAITING feature
- **Root Causes:**
  1. Status bar hid WAITING when count=0
  2. FileWatcher only works while TUI is running
  3. No visual indication feature exists
- **Lesson:** Feature must be visible even in "zero state"

---

## Process Integration

This checklist should be used:

1. **Before sprint review:** SM verifies checklist with dev team
2. **During retrospective:** Reference for what went well/wrong
3. **In epic template:** Add as standard section

---

*Created from Epic 4 Retrospective Action Item H2*
*Date: 2025-12-20*
