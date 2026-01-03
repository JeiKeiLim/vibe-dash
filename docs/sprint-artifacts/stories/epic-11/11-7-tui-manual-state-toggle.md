# Story 11.7: TUI Manual State Toggle

## Story
**As a** user viewing the dashboard,
**I want to** hibernate or activate projects directly from the TUI using a keybinding,
**So that** I can manage project visibility without leaving the dashboard or using CLI commands.

## Acceptance Criteria

### AC1: Keybinding for State Toggle
- **Given** a project is selected in the active projects list
- **When** user presses `H` key
- **Then** the project is hibernated and removed from active list

### AC2: Activate from Hibernated View
- **Given** a project is selected in the hibernated projects view (after pressing `h`)
- **When** user presses `H` key
- **Then** the project is activated and removed from hibernated list

### AC3: Favorite Protection
- **Given** a favorite project is selected
- **When** user presses `H` to hibernate
- **Then** show error feedback "Cannot hibernate favorite project"
- **And** project remains in active list

### AC4: Visual Feedback on Success
- **Given** user successfully hibernates/activates a project
- **When** the operation completes
- **Then** show feedback message (e.g., "Hibernated: project-name" or "Activated: project-name")
- **And** feedback clears after 2 seconds

### AC5: List Refresh After State Change
- **Given** a project state changes via `H` key
- **When** the operation completes
- **Then** both active and hibernated project lists are refreshed from database
- **And** project counts in status bar are updated

### AC6: Help Overlay Updated
- **Given** user opens help overlay (`?`)
- **When** viewing keybindings
- **Then** `H` keybinding is listed with description "Hibernate/Activate project"

### AC7: No-op When No Projects
- **Given** the project list is empty (no active or no hibernated projects)
- **When** user presses `H`
- **Then** nothing happens (no error, no feedback)

## Technical Notes

### Implementation Approach
Follow the existing favorite toggle pattern (`f` key):
1. Add `H` key handler in `handleKeyMsg()`
2. Create `toggleStateCmd()` that calls `stateService.Hibernate/Activate`
3. Define `stateToggledMsg` for success/error results
4. Handle message to update lists and show feedback
5. Update help overlay keybindings

### Key Files
- `internal/adapters/tui/model.go` - keybinding and command
- `internal/adapters/tui/model_test.go` - unit tests
- `internal/adapters/tui/views.go` - help overlay update

### Dependencies
- Story 11.5 (StateService with Hibernate/Activate methods) - DONE
- Story 11.4 (Hibernated projects view) - DONE

## Estimated Effort
Small (2-4 hours)

## Priority
Medium - Enhances UX by allowing state management without CLI

## Dev Agent Record
- **Created**: 2026-01-02
- **Status**: ready-for-dev
