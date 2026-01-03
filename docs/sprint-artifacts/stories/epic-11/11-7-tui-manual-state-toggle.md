# Story 11.7: TUI Manual State Toggle

Status: done

## Story

- **As a** user viewing the dashboard
- **I want to** hibernate or activate projects directly from the TUI using a keybinding
- **So that** I can manage project visibility without leaving the dashboard or using CLI commands

## User-Visible Changes

- **New:** Press `H` to hibernate the selected project (from active list)
- **New:** Press `H` to activate the selected project (from hibernated view)
- **New:** Feedback message "Hibernated: project-name" or "Activated: project-name" on success
- **New:** Error feedback "Cannot hibernate favorite project" when attempting to hibernate a favorite
- **Changed:** Help overlay (`?`) includes `H` keybinding with description "Hibernate/Activate project"

## Context & Background

### Previous Stories
- **Story 11.4** (Hibernated Projects View): Created `viewModeHibernated` with `hibernatedList`, `h` key toggle, Enter to wake
- **Story 11.5** (Manual State Control): Created CLI `vibe hibernate` and `vibe activate` commands using `StateService.Hibernate()` and `StateService.Activate()`
- **Story 11.6** (Hibernation Threshold): Added `hibernation_days` config with Viper binding

### Current State
- `StateActivator` interface at `internal/core/ports/state.go` has both `Activate()` and `Hibernate()` methods
  - **Note:** Interface was extended in Story 11.5 to include `Hibernate()` for CLI access (originally only `Activate()` for TUI auto-activation)
- `stateService ports.StateActivator` already wired in Model struct in `model.go`
- `KeyHibernated = "h"` already defined in `keys.go` (lowercase for view toggle)
- `KeyBindings` struct and `DefaultKeyBindings()` function in `keys.go` need new `StateToggle` field
- Help overlay in `views.go` `renderHelpOverlay()` function already shows `h` for "View hibernated projects"
- Favorite toggle pattern (`f` key) exists in `handleKeyMsg()` function in `model.go`

### Functional Requirements
- **FR57**: Manually hibernate or activate projects via CLI/TUI

## Acceptance Criteria

### AC1: Keybinding for State Toggle (Active → Hibernated)
- **Given** a project is selected in the active projects list
- **When** user presses `H` key (uppercase)
- **Then** the project is hibernated via `stateService.Hibernate()`
- **And** project is removed from active list

### AC2: Activate from Hibernated View
- **Given** a project is selected in the hibernated projects view (after pressing `h`)
- **When** user presses `H` key (uppercase)
- **Then** the project is activated via `stateService.Activate()`
- **And** project is removed from hibernated list

### AC3: Favorite Protection
- **Given** a favorite project is selected in active list
- **When** user presses `H` to hibernate
- **Then** show error feedback "Cannot hibernate favorite project"
- **And** project remains in active list
- **And** feedback clears after 3 seconds (consistent with other feedback)

### AC4: Visual Feedback on Success
- **Given** user successfully hibernates/activates a project
- **When** the operation completes
- **Then** show feedback message "Hibernated: project-name" or "Activated: project-name"
- **And** feedback clears after 3 seconds (status bar pattern)

### AC5: List Refresh After State Change
- **Given** a project state changes via `H` key
- **When** the operation completes
- **Then** appropriate list is refreshed from database (loadProjectsCmd or loadHibernatedProjectsCmd)
- **And** project counts in status bar are updated via `CalculateCountsWithWaiting`

### AC6: Help Overlay Updated
- **Given** user opens help overlay (`?`)
- **When** viewing keybindings
- **Then** `H` keybinding is listed with description "Hibernate/Activate project"
- **And** appears in "Actions" section after existing actions

### AC7: No-op When No Projects
- **Given** the project list is empty (no active or no hibernated projects)
- **When** user presses `H`
- **Then** nothing happens (no error, no feedback)

## Tasks / Subtasks

- [x] Task 1: Add KeyStateToggle constant and update KeyBindings struct (AC: #1, #2, #7)
  - [x] 1.1: Add `KeyStateToggle = "H"` constant to `keys.go` (uppercase H)
  - [x] 1.2: Add `StateToggle string` field to `KeyBindings` struct in `keys.go`
  - [x] 1.3: Add `StateToggle: KeyStateToggle` to `DefaultKeyBindings()` function in `keys.go`
  - [x] 1.4: Add `case KeyStateToggle:` in `handleKeyMsg()` switch statement in `model.go`
  - [x] 1.5: Check `viewMode == viewModeHibernated` to determine action (hibernate vs activate)
  - [x] 1.6: Get selected project from `m.hibernatedList.SelectedProject()` or `m.projectList.SelectedProject()`
  - [x] 1.7: Return `nil` if no selected project (AC7)

- [x] Task 2: Create stateToggleCmd and stateToggledMsg (AC: #1, #2, #3)
  - [x] 2.1: Create `stateToggledMsg` struct with `projectID`, `projectName`, `action` ("hibernated"/"activated"), `err`
  - [x] 2.2: Create `stateToggleCmd(projectID, projectName string, hibernate bool)` tea.Cmd
  - [x] 2.3: If `hibernate == true`: call `m.stateService.Hibernate(ctx, projectID)`
  - [x] 2.4: If `hibernate == false`: call `m.stateService.Activate(ctx, projectID)`
  - [x] 2.5: Return `stateToggledMsg` with appropriate action string

- [x] Task 3: Handle stateToggledMsg in Update() (AC: #3, #4, #5)
  - [x] 3.1: Add `case stateToggledMsg:` in Update() switch
  - [x] 3.2: If `msg.err == domain.ErrFavoriteCannotHibernate`: show "Cannot hibernate favorite project" feedback (AC3)
  - [x] 3.3: If `msg.err == domain.ErrInvalidStateTransition`: handle idempotent case (reload list silently)
  - [x] 3.4: If `msg.err != nil` (other error): log warning, show generic error feedback
  - [x] 3.5: On success: show "Hibernated: name" or "Activated: name" via `m.statusBar.SetRefreshComplete()`
  - [x] 3.6: If hibernated from active view: return `m.loadProjectsCmd()` (AC5)
  - [x] 3.7: If activated from hibernated view: return `m.loadHibernatedProjectsCmd()` (AC5)
  - [x] 3.8: Start 3-second timer for `clearStateToggleFeedbackMsg`

- [x] Task 4: Add clearStateToggleFeedbackMsg handler (AC: #4)
  - [x] 4.1: ~~Create `clearStateToggleFeedbackMsg` struct~~ (N/A - using alternative 4.3)
  - [x] 4.2: ~~Add handler in Update()~~ (N/A - using alternative 4.3)
  - [x] 4.3: ALTERNATIVE: Reuse existing `clearRemoveFeedbackMsg` pattern instead of new msg type

- [x] Task 5: Update help overlay (AC: #6)
  - [x] 5.1: In `views.go:renderHelpOverlay()`, add `"H        Hibernate/Activate"` after `"x        Remove project"` line
  - [x] 5.2: Verify help overlay width (52) still accommodates new line

- [x] Task 6: Write comprehensive tests (AC: #1-7)
  - [x] 6.1: Add `TestModel_StateToggle_HibernateFromActiveView_AC1` - H key hibernates selected project
  - [x] 6.2: Add `TestModel_StateToggle_ActivateFromHibernatedView_AC2` - H key activates from hibernated view
  - [x] 6.3: Add `TestModel_StateToggle_FavoriteProtection_AC3` - Error feedback for favorite project
  - [x] 6.4: Add `TestModel_StateToggle_SuccessFeedback_AC4` - Verify status bar feedback message
  - [x] 6.5: Add `TestModel_StateToggle_ListRefresh_AC5` - Verify loadProjectsCmd returned
  - [x] 6.6: Add `TestModel_StateToggle_NoOpWhenEmpty_AC7` - No crash when list is empty
  - [x] 6.7: Add `TestModel_StateToggle_IdempotentBehavior` - ErrInvalidStateTransition handled gracefully

## Technical Implementation Guide

### Overview
Add uppercase `H` key handler that toggles project state between active and hibernated. In active view, H hibernates; in hibernated view, H activates. Follows existing favorite toggle pattern.

### Architecture Compliance

```
internal/adapters/tui/keys.go           ←  ADD KeyStateToggle = "H"
         ↓ used by
internal/adapters/tui/model.go          ←  ADD handler, cmd, msg
         ↓ calls
internal/core/ports/state.go            ←  EXISTING: StateActivator interface
         ↓ implements
internal/core/services/state_service.go ←  EXISTING: Hibernate(), Activate()
```

### File Changes

#### 1. `internal/adapters/tui/keys.go` (MODIFY)

**Add new constant after existing keys (around line 24-28):**

```go
const (
    // ... existing keys ...
    KeyStateToggle = "H" // Uppercase H for state toggle (Story 11.7)
)
```

**Add field to KeyBindings struct (after Hibernated field around line 54):**

```go
type KeyBindings struct {
    // ... existing fields ...
    // Views
    Hibernated string
    StateToggle string // Story 11.7: Manual state toggle
}
```

**Add to DefaultKeyBindings() (after Hibernated line around line 81):**

```go
func DefaultKeyBindings() KeyBindings {
    return KeyBindings{
        // ... existing bindings ...
        // Views
        Hibernated:  KeyHibernated,
        StateToggle: KeyStateToggle, // Story 11.7
    }
}
```

#### 2. `internal/adapters/tui/model.go` (MODIFY)

**Required imports (ensure these are present):**

```go
import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "time"
    // ... existing imports ...
)
```

**Add message type (after `projectActivatedMsg` struct):**

```go
// stateToggledMsg signals a project state was toggled via H key (Story 11.7).
type stateToggledMsg struct {
    projectID   string
    projectName string
    action      string // "hibernated" or "activated"
    err         error
}
```

**Add command (after `activateProjectCmd` function):**

```go
// stateToggleCmd toggles project state (Story 11.7).
// If hibernate=true, calls Hibernate(); otherwise calls Activate().
func (m Model) stateToggleCmd(projectID, projectName string, hibernate bool) tea.Cmd {
    if m.stateService == nil {
        return func() tea.Msg {
            return stateToggledMsg{
                projectID:   projectID,
                projectName: projectName,
                err:         errors.New("state service not available"),
            }
        }
    }
    return func() tea.Msg {
        ctx := context.Background()
        var err error
        var action string
        if hibernate {
            err = m.stateService.Hibernate(ctx, projectID)
            action = "hibernated"
        } else {
            err = m.stateService.Activate(ctx, projectID)
            action = "activated"
        }
        return stateToggledMsg{projectID: projectID, projectName: projectName, action: action, err: err}
    }
}
```

**Add key handler in `handleKeyMsg()` (after `KeyRemove` case, before the view mode forwarding block):**

```go
case KeyStateToggle:
    // Story 11.7: Toggle project state with H key
    if m.viewMode == viewModeHibernated {
        // In hibernated view: activate selected project (AC2)
        if len(m.hibernatedProjects) == 0 {
            return m, nil // AC7: No-op when empty
        }
        selected := m.hibernatedList.SelectedProject()
        if selected == nil {
            return m, nil
        }
        return m, m.stateToggleCmd(selected.ID, project.EffectiveName(selected), false)
    }
    // In active view: hibernate selected project (AC1)
    if len(m.projects) == 0 {
        return m, nil // AC7: No-op when empty
    }
    selected := m.projectList.SelectedProject()
    if selected == nil {
        return m, nil
    }
    return m, m.stateToggleCmd(selected.ID, project.EffectiveName(selected), true)
```

**Add message handler in `Update()` (after `projectActivatedMsg` case):**

```go
case stateToggledMsg:
    // Story 11.7: Handle state toggle result
    if msg.err != nil {
        if errors.Is(msg.err, domain.ErrFavoriteCannotHibernate) {
            // AC3: Favorite protection - show error feedback
            m.statusBar.SetRefreshComplete("Cannot hibernate favorite project")
            return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
                return clearRemoveFeedbackMsg{} // Reuse existing clear msg
            })
        }
        if errors.Is(msg.err, domain.ErrInvalidStateTransition) {
            // Idempotent case - reload appropriate list silently
            if msg.action == "hibernated" {
                return m, m.loadProjectsCmd()
            }
            return m, m.loadHibernatedProjectsCmd()
        }
        // General error
        slog.Warn("failed to toggle project state", "error", msg.err)
        m.statusBar.SetRefreshComplete("✗ State change failed")
        return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
            return clearRemoveFeedbackMsg{}
        })
    }
    // AC4: Success feedback (avoid deprecated strings.Title)
    var feedback string
    if msg.action == "hibernated" {
        feedback = fmt.Sprintf("✓ Hibernated: %s", msg.projectName)
    } else {
        feedback = fmt.Sprintf("✓ Activated: %s", msg.projectName)
    }
    m.statusBar.SetRefreshComplete(feedback)
    // AC5: Reload appropriate list
    var reloadCmd tea.Cmd
    if msg.action == "hibernated" {
        reloadCmd = m.loadProjectsCmd()
    } else {
        reloadCmd = m.loadHibernatedProjectsCmd()
    }
    return m, tea.Batch(
        reloadCmd,
        tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
            return clearRemoveFeedbackMsg{}
        }),
    )
```

#### 3. `internal/adapters/tui/views.go` (MODIFY)

**Update help overlay (in `renderHelpOverlay()` function, Actions section):**

```go
"Actions",
"d        Toggle detail panel",
"f        Toggle favorite",
"n        Edit notes",
"x        Remove project",
"H        Hibernate/Activate",  // Story 11.7: Add this line
"a        Add project",
"r        Refresh/rescan",
```

### Testing Strategy

#### Unit Tests (in `internal/adapters/tui/model_test.go`)

| Test Case | Coverage |
|-----------|----------|
| `TestModel_StateToggle_HibernateFromActiveView_AC1` | H key in active view calls Hibernate() |
| `TestModel_StateToggle_ActivateFromHibernatedView_AC2` | H key in hibernated view calls Activate() |
| `TestModel_StateToggle_FavoriteProtection_AC3` | ErrFavoriteCannotHibernate shows error feedback |
| `TestModel_StateToggle_SuccessFeedback_AC4` | Success message in status bar |
| `TestModel_StateToggle_ListRefresh_AC5` | Returns loadProjectsCmd/loadHibernatedProjectsCmd |
| `TestModel_StateToggle_NoOpWhenEmpty_AC7` | Returns nil when no projects |
| `TestModel_StateToggle_IdempotentBehavior` | ErrInvalidStateTransition handled gracefully |

### Edge Cases

1. **No stateService**: Return error message in stateToggledMsg
2. **Favorite project**: Return ErrFavoriteCannotHibernate from StateService
3. **Already in target state**: StateService returns ErrInvalidStateTransition, reload list silently
4. **Concurrent state change**: Race with auto-activation (Story 11.3) - ErrInvalidStateTransition is expected
5. **Empty list**: Return nil, no-op (AC7)

## Dev Notes

| Decision | Rationale |
|----------|-----------|
| Uppercase `H` vs lowercase `h` | `h` already used for view toggle (Story 11.4), `H` is action |
| Reuse `clearRemoveFeedbackMsg` | Avoid proliferation of similar message types |
| 3-second feedback timeout | Consistent with remove/favorite patterns |
| Reload full list on success | Ensures counts and sorting are correct |
| Explicit feedback strings | Avoid deprecated `strings.Title`; matches CLI format from Story 11.5 |
| Update KeyBindings struct | Maintains consistency between constants and configurable bindings |

### Reuse vs Create Quick Reference

| Item | Action | Source |
|------|--------|--------|
| `StateActivator.Hibernate()` | REUSE | `internal/core/ports/state.go:12` |
| `StateActivator.Activate()` | REUSE | `internal/core/ports/state.go:14` |
| `m.stateService` | REUSE | `internal/adapters/tui/model.go:118` |
| `project.EffectiveName()` | REUSE | `internal/shared/project/project.go` |
| `domain.ErrFavoriteCannotHibernate` | REUSE | `internal/core/domain/errors.go` |
| `domain.ErrInvalidStateTransition` | REUSE | `internal/core/domain/errors.go` |
| `clearRemoveFeedbackMsg` | REUSE | `internal/adapters/tui/model.go:211` |
| `KeyHibernated = "h"` | REUSE | `internal/adapters/tui/keys.go:27` (for view toggle) |
| `KeyStateToggle = "H"` | CREATE | NEW constant |
| `stateToggledMsg` | CREATE | NEW message type |
| `stateToggleCmd` | CREATE | NEW command |

## Dependencies

- **Story 11.4**: Hibernated projects view with `hibernatedList` - DONE
- **Story 11.5**: StateService with Hibernate/Activate methods accessible in TUI - DONE

## File List

**MODIFY:**
- `internal/adapters/tui/keys.go` - Add `KeyStateToggle = "H"`, update `KeyBindings` struct and `DefaultKeyBindings()`
- `internal/adapters/tui/model.go` - Add stateToggledMsg, stateToggleCmd, handlers
- `internal/adapters/tui/model_test.go` - Add Story 11.7 tests
- `internal/adapters/tui/views.go` - Add H to help overlay
- `docs/sprint-artifacts/sprint-status.yaml` - Story status update

**DO NOT MODIFY:**
- `internal/core/ports/state.go` - Interface already has Hibernate() and Activate()
- `internal/core/services/state_service.go` - Implementation already exists
- `internal/core/domain/errors.go` - ErrFavoriteCannotHibernate already exists

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Setup Test Data

```bash
make build

# Add test projects
./bin/vibe add /tmp/test-state-toggle
mkdir -p /tmp/test-state-toggle

# Verify project is active
./bin/vibe list
```

### Step 2: Test Hibernate from Active View (AC1)

| Check | Expected | Status |
|-------|----------|--------|
| Run `./bin/vibe`, select project | Project visible in list | |
| Press `H` (uppercase) | Feedback "✓ Hibernated: project-name" | |
| Project count decreases | Status bar shows updated count | |
| Press `h` (lowercase) | Hibernated view shows the project | |

### Step 3: Test Activate from Hibernated View (AC2)

| Check | Expected | Status |
|-------|----------|--------|
| In hibernated view, select project | Project visible | |
| Press `H` (uppercase) | Feedback "✓ Activated: project-name" | |
| View switches to active | Project back in active list | |

### Step 4: Test Favorite Protection (AC3)

```bash
# Make project a favorite first
./bin/vibe favorite test-state-toggle
./bin/vibe
# Select the favorited project
# Press H
```

| Check | Expected | Status |
|-------|----------|--------|
| Press H on favorite | "Cannot hibernate favorite project" feedback | |
| Project remains in list | Not hibernated | |

### Step 5: Test Help Overlay (AC6)

| Check | Expected | Status |
|-------|----------|--------|
| Press `?` | Help overlay shows | |
| Find "H" keybinding | "Hibernate/Activate" description visible | |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Any check fails | Do NOT approve, document issue |

## Verification Checklist

Before marking complete, verify:

- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/adapters/tui/...` passes
- [ ] `golangci-lint run` passes
- [ ] User Testing Guide Step 2: Hibernate works
- [ ] User Testing Guide Step 3: Activate works
- [ ] User Testing Guide Step 4: Favorite protection works
- [ ] User Testing Guide Step 5: Help overlay updated

## Story Wrap Up (Agent Populates After Completion)

### Completion Checklist
- [x] All ACs verified
- [x] Tests pass
- [x] Code review findings addressed
- [x] Documentation updated if needed

### Dev Agent Record

**Created:** 2026-01-02
**SM Validation:** 2026-01-03
**Implemented:** 2026-01-03

**Validation Report:** `validation-report-11-7-20260103.md`

**Improvements Applied (3 Critical, 4 Enhancements):**
1. [Critical] Fixed deprecated `strings.Title` - replaced with explicit feedback strings
2. [Critical] Fixed incorrect file path `name.go` → `project.go` in Reuse table
3. [Critical] Added `KeyBindings` struct and `DefaultKeyBindings()` updates to implementation guide
4. [Enhancement] Added required imports list to code examples
5. [Enhancement] Referenced Story 11.5 interface extension explicitly in Current State
6. [Enhancement] Changed line number references to function-based anchors for robustness
7. [Enhancement] Updated Dev Notes table with new decisions

**Implementation Summary:**
- All 6 tasks completed per story specification
- Task 4 used ALTERNATIVE approach: Reused `clearRemoveFeedbackMsg` instead of creating new message type
- 9 unit tests added covering all 7 ACs plus error handling and help overlay
- All tests pass, linter clean, build successful

**Code Review Fixes Applied:**
1. [H1] Fixed Task 4.1/4.2 marking - correctly show N/A for alternative path
2. [M1] Added `TestHelpOverlay_ContainsStateToggleKeybinding_AC6` test for help overlay
3. [M2] Added 5-second context timeout to `stateToggleCmd()` to prevent TUI freeze
4. [L2] Updated completion checklist to checked state

**Files Modified:**
- `internal/adapters/tui/keys.go` - KeyStateToggle constant, KeyBindings struct, DefaultKeyBindings()
- `internal/adapters/tui/model.go` - stateToggledMsg, stateToggleCmd() with timeout, KeyStateToggle handler, stateToggledMsg handler
- `internal/adapters/tui/views.go` - Help overlay updated with "H Hibernate/Activate"
- `internal/adapters/tui/model_test.go` - 9 tests for Story 11.7 (8 + AC6 help overlay)
- `docs/sprint-artifacts/sprint-status.yaml` - Status update
