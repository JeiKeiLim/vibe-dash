# Story 4.2: Activity Timestamp Tracking

Status: done

## Story

As a **system**,
I want **to track last activity time per project based on file events**,
so that **I can calculate inactivity duration for the "Agent Waiting" detection feature (FR34-38)**.

## Acceptance Criteria

1. **AC1: Activity Timestamp Update from FileEvents**
   - Given FileWatcher is emitting events from Story 4.1
   - When a FileEvent occurs for a project directory
   - Then `project.LastActivityAt` is updated to the event timestamp
   - And the database is updated with the new timestamp

2. **AC2: Path-to-Project Mapping**
   - Given a FileEvent with a file path
   - When processing the event
   - Then the system correctly identifies which project the path belongs to
   - And only updates the matching project's `LastActivityAt`

3. **AC3: Relative Time Display**
   - Given TUI is displaying projects
   - When showing "last active" timestamp
   - Then display shows human-readable relative time:
     - `"just now"` for < 1 minute
     - `"5m ago"` for < 1 hour
     - `"2h ago"` for < 24 hours
     - `"3d ago"` for < 7 days
     - `"2w ago"` for >= 7 days

4. **AC4: Periodic Display Refresh**
   - Given TUI is open
   - When time passes
   - Then timestamps update every minute via Bubble Tea tick
   - And display refreshes without jarring visual changes

5. **AC5: Manual Refresh Updates Timestamps**
   - Given user presses 'r' for refresh
   - When refresh is triggered
   - Then all timestamps recalculate immediately
   - And display updates to reflect current relative times

6. **AC6: Context Cancellation Support**
   - Given ActivityTracker is processing events
   - When context is cancelled
   - Then event processing stops gracefully
   - And any pending database updates complete or abort cleanly

7. **AC7: Graceful Degradation**
   - Given ActivityTracker cannot identify the project for an event
   - When the event path doesn't match any tracked project
   - Then the event is logged at debug level and skipped
   - And processing continues for other events

## Tasks / Subtasks

- [x] Task 1: Create ActivityTracker Service (AC: 1, 2, 6, 7)
  - [x] 1.1: Create `internal/core/services/activity_tracker.go`
  - [x] 1.2: Implement `ActivityTracker` struct with repository dependency
  - [x] 1.3: Implement `ProcessEvents(ctx, <-chan FileEvent)` method
  - [x] 1.4: Implement path-to-project matching logic with symlink handling
  - [x] 1.5: Implement `SetProjects(projects []*domain.Project)` method to populate cache
  - [x] 1.6: Update `LastActivityAt` in repository on event

- [x] Task 2: Add UpdateLastActivity to Repository Interface (AC: 1)
  - [x] 2.1: Add `UpdateLastActivity(ctx, id string, timestamp time.Time) error` to `ports.ProjectRepository`
  - [x] 2.2: Implement in `sqlite.ProjectRepository` at line ~200 (after UpdateState)
  - [x] 2.3: Implement in `persistence.RepositoryCoordinator` with project lookup

- [x] Task 3: Create RelativeTime Formatting Utility (AC: 3)
  - [x] 3.1: Already exists in `internal/shared/timeformat/timeformat.go`
  - [x] 3.2: `FormatRelativeTime(t time.Time) string` already implemented
  - [x] 3.3: Handle all time ranges from spec (just now, Xm ago, Xh ago, Xd ago, Xw ago)
  - [x] 3.4: Unit tests already exist in `internal/shared/timeformat/timeformat_test.go`

- [x] Task 4: Integrate with TUI Display (AC: 3, 4, 5)
  - [x] 4.1: `components/delegate.go` already uses `FormatRelativeTime` (line 175)
  - [x] 4.2: Add Bubble Tea tick command for periodic updates (60s interval)
  - [x] 4.3: Ensure refresh command ('r') recalculates all timestamps via View()

- [x] Task 5: Write Comprehensive Tests (AC: all)
  - [x] 5.1: Unit tests for ActivityTracker in `activity_tracker_test.go`
  - [x] 5.2: Unit tests for path-to-project matching
  - [x] 5.3: Unit tests for RelativeTime formatting with boundary values
  - [x] 5.4: Integration tests with mock FileWatcher
  - [x] 5.5: Test context cancellation behavior

## Dev Notes

### Architecture Compliance

**Hexagonal Architecture Boundaries:**
- ActivityTracker: `internal/core/services/activity_tracker.go` (service layer)
- RelativeTime utility: `internal/adapters/tui/utils/time.go` (TUI adapter - display-only)
- **CRITICAL:** ActivityTracker is in core/services but uses port interfaces only
- Repository interface extension maintains clean architecture

**Existing Interface to Extend:**
```go
// internal/core/ports/repository.go - ADD this method after UpdateState (line ~58)
type ProjectRepository interface {
    // ... existing methods ...

    // UpdateLastActivity updates only the LastActivityAt timestamp.
    // Returns domain.ErrProjectNotFound if no project exists with the given ID.
    // This is an optimized update for high-frequency activity tracking.
    UpdateLastActivity(ctx context.Context, id string, timestamp time.Time) error
}
```

### ActivityTracker Implementation

**Complete Struct Definition:**
```go
// internal/core/services/activity_tracker.go
package services

import (
    "context"
    "log/slog"
    "strings"
    "sync"
    "time"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// ActivityTracker consumes FileEvents and updates project LastActivityAt
type ActivityTracker struct {
    repo     ports.ProjectRepository
    projects map[string]*domain.Project // Path prefix -> project (cached)
    mu       sync.RWMutex
}

// NewActivityTracker creates a new tracker with the given repository.
// Call SetProjects() to populate the path cache before ProcessEvents().
func NewActivityTracker(repo ports.ProjectRepository) *ActivityTracker {
    return &ActivityTracker{
        repo:     repo,
        projects: make(map[string]*domain.Project),
    }
}

// SetProjects populates the path cache for path-to-project matching.
// Call this on startup and when projects are added/removed.
func (t *ActivityTracker) SetProjects(projects []*domain.Project) {
    t.mu.Lock()
    defer t.mu.Unlock()

    t.projects = make(map[string]*domain.Project, len(projects))
    for _, p := range projects {
        // Normalize path by removing trailing slash
        path := strings.TrimSuffix(p.Path, "/")
        t.projects[path] = p
    }
}

// ProcessEvents consumes events from FileWatcher and updates LastActivityAt.
// Blocks until context is cancelled or channel is closed.
func (t *ActivityTracker) ProcessEvents(ctx context.Context, events <-chan ports.FileEvent) {
    for {
        select {
        case <-ctx.Done():
            return
        case event, ok := <-events:
            if !ok {
                return
            }
            t.handleEvent(ctx, event)
        }
    }
}

func (t *ActivityTracker) handleEvent(ctx context.Context, event ports.FileEvent) {
    project := t.findProjectForPath(event.Path)
    if project == nil {
        slog.Debug("event path not matched to any project", "path", event.Path)
        return
    }

    if err := t.repo.UpdateLastActivity(ctx, project.ID, event.Timestamp); err != nil {
        slog.Warn("failed to update last activity", "project_id", project.ID, "error", err)
    }
}

// findProjectForPath matches event path to project using path prefix
func (t *ActivityTracker) findProjectForPath(eventPath string) *domain.Project {
    t.mu.RLock()
    defer t.mu.RUnlock()

    // Normalize event path
    eventPath = strings.TrimSuffix(eventPath, "/")

    for projectPath, project := range t.projects {
        // Check if event path starts with project path followed by "/" or is exact match
        if eventPath == projectPath || strings.HasPrefix(eventPath, projectPath+"/") {
            return project
        }
    }
    return nil
}
```

**Path Matching Details:**
- Event paths from Story 4.1 are already canonical (symlinks resolved)
- Match using `strings.HasPrefix(eventPath, projectPath+"/")` to ensure proper directory boundary
- Direct equality check for exact path match (file in project root)

### SQLite Implementation

**Add to internal/adapters/persistence/sqlite/queries.go:**
```go
// UpdateLastActivity - add after updateStateSQL constant
const updateLastActivitySQL = `
    UPDATE projects
    SET last_activity_at = ?, updated_at = ?
    WHERE id = ?
`
```

**Add to internal/adapters/persistence/sqlite/project_repository.go (after UpdateState method ~line 200):**
```go
// UpdateLastActivity updates only the LastActivityAt timestamp.
// Returns domain.ErrProjectNotFound if no project exists with the given ID.
func (r *ProjectRepository) UpdateLastActivity(ctx context.Context, id string, timestamp time.Time) error {
    db, err := r.openDB(ctx)
    if err != nil {
        return err
    }
    defer db.Close()

    now := time.Now().UTC().Format(time.RFC3339)
    ts := timestamp.UTC().Format(time.RFC3339)

    result, err := db.ExecContext(ctx, updateLastActivitySQL, ts, now, id)
    if err != nil {
        return fmt.Errorf("failed to update last activity: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rowsAffected == 0 {
        return domain.ErrProjectNotFound
    }

    return nil
}
```

**Add to internal/adapters/persistence/coordinator.go (after UpdateState method):**
```go
// UpdateLastActivity updates the LastActivityAt for a project.
// Delegates to the appropriate per-project repository.
func (c *RepositoryCoordinator) UpdateLastActivity(ctx context.Context, id string, timestamp time.Time) error {
    c.mu.RLock()
    repo, exists := c.projectRepos[id]
    c.mu.RUnlock()

    if !exists {
        return domain.ErrProjectNotFound
    }

    return repo.UpdateLastActivity(ctx, id, timestamp)
}
```

### Relative Time Formatting

**internal/adapters/tui/utils/time.go:**
```go
package utils

import (
    "fmt"
    "time"
)

// FormatRelativeTime returns a human-readable relative time string.
// Implements AC3 time format requirements.
func FormatRelativeTime(t time.Time) string {
    d := time.Since(t)

    switch {
    case d < time.Minute:
        return "just now"
    case d < time.Hour:
        return fmt.Sprintf("%dm ago", int(d.Minutes()))
    case d < 24*time.Hour:
        return fmt.Sprintf("%dh ago", int(d.Hours()))
    case d < 7*24*time.Hour:
        return fmt.Sprintf("%dd ago", int(d.Hours()/24))
    default:
        return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
    }
}
```

**Unit Test Examples (utils/time_test.go):**
```go
func TestFormatRelativeTime(t *testing.T) {
    tests := []struct {
        name     string
        duration time.Duration
        expected string
    }{
        {"just now - 0s", 0, "just now"},
        {"just now - 30s", 30 * time.Second, "just now"},
        {"just now - 59s", 59 * time.Second, "just now"},
        {"1m ago", time.Minute, "1m ago"},
        {"5m ago", 5 * time.Minute, "5m ago"},
        {"59m ago", 59 * time.Minute, "59m ago"},
        {"1h ago", time.Hour, "1h ago"},
        {"23h ago", 23 * time.Hour, "23h ago"},
        {"1d ago", 24 * time.Hour, "1d ago"},
        {"6d ago", 6 * 24 * time.Hour, "6d ago"},
        {"1w ago", 7 * 24 * time.Hour, "1w ago"},
        {"2w ago", 14 * 24 * time.Hour, "2w ago"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FormatRelativeTime(time.Now().Add(-tt.duration))
            if got != tt.expected {
                t.Errorf("FormatRelativeTime() = %q, want %q", got, tt.expected)
            }
        })
    }
}
```

### TUI Integration

**Tick Command Pattern - add to internal/adapters/tui/model.go:**
```go
// tickMsg is sent every 60 seconds to trigger timestamp recalculation
type tickMsg time.Time

// tickCmd returns a command that ticks every 60 seconds
func tickCmd() tea.Cmd {
    return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

**Update Model.Update() to handle tick:**
```go
case tickMsg:
    // Timestamps automatically recalculate in View() via FormatRelativeTime()
    // Just return next tick command to continue periodic updates
    return m, tickCmd()
```

**Initialize tick in Model.Init():**
```go
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        // ... existing commands ...
        tickCmd(), // Start periodic timestamp refresh
    )
}
```

**Update delegate.go to use FormatRelativeTime:**
```go
// In project row rendering, replace hardcoded time with:
import "github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/utils"

// ... in Render method:
lastActive := utils.FormatRelativeTime(project.LastActivityAt)
```

### File Structure

```
internal/core/services/
    activity_tracker.go           # ActivityTracker service
    activity_tracker_test.go      # Unit tests

internal/adapters/tui/utils/
    time.go                       # FormatRelativeTime function
    time_test.go                  # Time formatting tests

internal/core/ports/
    repository.go                 # Add UpdateLastActivity method

internal/adapters/persistence/sqlite/
    project_repository.go         # Add UpdateLastActivity implementation
    queries.go                    # Add updateLastActivitySQL constant

internal/adapters/persistence/
    coordinator.go                # Add UpdateLastActivity delegation

internal/adapters/tui/
    model.go                      # Add tickMsg and tickCmd
    components/delegate.go        # Use FormatRelativeTime
```

### Previous Story Intelligence

**Key Learnings from Story 4.1 to Apply:**
1. **Context check at start** - All methods check ctx.Done() first
2. **Graceful degradation** - Log warnings at debug level for unmatched paths
3. **Thread-safe operations** - Use RWMutex for project cache
4. **Channel handling** - Check `ok` on channel receive for closed detection
5. **Debounced events** - Events from 4.1 are already aggregated, reducing DB writes

**Event Paths are Already Canonical:**
Story 4.1's FsnotifyWatcher uses `CanonicalPath()` before emitting events, so ActivityTracker does NOT need to re-canonicalize paths.

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Put ActivityTracker in adapters/ | Keep in core/services/ |
| Query DB on every event | Cache project paths in memory |
| Log every event at info level | Use debug level for event processing |
| Skip context check | Always check ctx.Done() first |
| Re-canonicalize event paths | Trust paths from Story 4.1 FileWatcher |
| Full project Save on activity | Use optimized UpdateLastActivity |

### Testing Strategy

**Unit Tests:**
- `TestActivityTracker_ProcessEvents` - Basic event processing
- `TestActivityTracker_PathToProjectMatching` - Various path scenarios including nested
- `TestActivityTracker_UnknownPath` - Graceful handling of unmatched paths
- `TestActivityTracker_ContextCancellation` - Clean shutdown
- `TestFormatRelativeTime` - All time ranges with boundary tests

**Integration Tests:**
- `TestActivityTracker_Integration` - With mock FileWatcher and real repo
- `TestTUITimestampRefresh` - Verify display updates (if TUI testing is set up)

**Edge Cases to Test:**
- Event for deleted project (path no longer in cache)
- Multiple events for same project in quick succession
- Event path is deeply nested subdirectory of project path
- Empty project cache scenario
- Concurrent updates to same project

### Downstream Dependencies

**Story 4.3 depends on this story for:**
- `LastActivityAt` being kept up-to-date
- Foundation for calculating "waiting" duration
- `time.Since(LastActivityAt)` determines waiting state

**Story 4.5 depends on this story for:**
- Accurate "last active: Xm ago" display
- Visual update timing (1-minute refresh)

**Story 4.6 depends on this story for:**
- Real-time dashboard showing fresh timestamps
- Recency indicators (fresh, today, this week) calculation

### References

| Document | Section | Line/Location |
|----------|---------|---------------|
| docs/architecture.md | Go Code Conventions | Constructor pattern, context first |
| docs/architecture.md | Database Naming | snake_case columns |
| docs/epics.md | Story 4.2 | Lines 1605-1639 |
| docs/prd.md | FR34-38 | Agent waiting detection |
| docs/project-context.md | SQLite Rules | Lazy connections, WAL mode |
| internal/core/ports/repository.go | ProjectRepository | Line 22-59 |
| internal/core/ports/watcher.go | FileEvent | Line 10-16 |
| internal/adapters/filesystem/watcher.go | FsnotifyWatcher | Lines 39-49 (constructor) |
| internal/adapters/persistence/coordinator.go | RepositoryCoordinator | Delegation pattern |

### Manual Testing Steps

After implementation, verify:

1. **Activity Update:**
   ```bash
   # Add a test project
   ./bin/vibe add /tmp/test-project

   # Touch a file
   touch /tmp/test-project/test.txt

   # Verify LastActivityAt updated in TUI or via JSON output
   ./bin/vibe list --json | jq '.projects[0].last_activity_at'
   ```

2. **Relative Time Display:**
   ```bash
   # Run TUI and observe "last active" column
   ./bin/vibe

   # Verify formats:
   # - Recent activity shows "just now"
   # - After 1 minute shows "1m ago"
   ```

3. **Tick Update:**
   ```bash
   # Leave TUI open for 2+ minutes
   # Observe timestamp values updating every minute
   ```

4. **Manual Refresh:**
   ```bash
   # In TUI, press 'r'
   # Verify all timestamps recalculate immediately
   ```

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

### Completion Notes List

- **AC1**: Implemented `UpdateLastActivity` method in ports, sqlite, and coordinator
- **AC2**: Implemented path-to-project matching with path prefix matching and normalization
- **AC3**: Leveraged existing `FormatRelativeTime` from `internal/shared/timeformat/timeformat.go`
- **AC4**: Added 60-second tick command in `model.go` for periodic timestamp refresh
- **AC5**: Timestamps automatically recalculate on refresh via `View()` calling `FormatRelativeTime`
- **AC6**: Context cancellation supported in `ProcessEvents` with select on `ctx.Done()`
- **AC7**: Unknown paths logged at debug level and gracefully skipped
- **Note**: `FormatRelativeTime` already existed in shared/timeformat - reused instead of duplicating

### File List

**New Files:**
- `internal/core/services/activity_tracker.go` - ActivityTracker service implementation
- `internal/core/services/activity_tracker_test.go` - Comprehensive unit tests (13 tests)

**Modified Files:**
- `internal/core/ports/repository.go` - Added `UpdateLastActivity` method to interface
- `internal/adapters/persistence/sqlite/queries.go` - Added `updateLastActivitySQL` constant
- `internal/adapters/persistence/sqlite/project_repository.go` - Implemented `UpdateLastActivity`
- `internal/adapters/persistence/sqlite/project_repository_test.go` - Added 4 UpdateLastActivity tests
- `internal/adapters/persistence/coordinator.go` - Implemented `UpdateLastActivity` delegation
- `internal/adapters/tui/model.go` - Added `tickMsg`, `tickCmd()`, Init batch with tick, Update handler
- `internal/adapters/tui/model_test.go` - Added 4 tick-related tests
- `internal/adapters/cli/add_test.go` - Updated mock with `UpdateLastActivity`
- `internal/adapters/cli/favorite_test.go` - Updated mock with `UpdateLastActivity`
- `internal/adapters/cli/note_test.go` - Updated mock with `UpdateLastActivity`
- `internal/adapters/cli/list_test.go` - Updated mock with `UpdateLastActivity`
- `internal/core/ports/repository_test.go` - Updated mock with `UpdateLastActivity`
- `internal/adapters/tui/model_favorite_test.go` - Updated mock with `UpdateLastActivity`
- `internal/adapters/tui/model_notes_test.go` - Updated mock with `UpdateLastActivity`
- `internal/adapters/tui/model_refresh_test.go` - Updated mock with `UpdateLastActivity`

### Change Log

| Date | Author | Change |
|------|--------|--------|
| 2025-12-19 | SM (Bob) | Initial story creation via *create-story workflow |
| 2025-12-19 | SM (Bob) | Validation improvements: Added SetProjects method, coordinator implementation, path matching details, TUI tick integration, test boundary cases |
| 2025-12-19 | Dev (Amelia) | Implementation complete: ActivityTracker service, UpdateLastActivity in repository, tick command for TUI, comprehensive tests |
| 2025-12-19 | Dev (Amelia) | Code review fixes: H2 - Added O(1) projectID cache in coordinator, M2 - Added repository error test, updated story status to done |
