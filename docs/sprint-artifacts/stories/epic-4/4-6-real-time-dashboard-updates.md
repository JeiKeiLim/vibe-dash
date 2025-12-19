# Story 4.6: Real-Time Dashboard Updates

Status: done

## Story

As a **user**,
I want **the dashboard to update automatically**,
so that **I see current state without manual refresh (FR27)**.

## Acceptance Criteria

1. **AC1: Dashboard Auto-Update on File Changes**
   - Given dashboard is open and file watcher is running
   - When file changes in tracked project
   - Then dashboard updates within 5-10 seconds:
     - LastActivity timestamp updates
     - WAITING state clears if was waiting
     - Recency indicator may change (e.g., none -> "this week")

2. **AC2: Smooth Re-Render Without Jarring Visual Changes**
   - Given dashboard is displaying projects
   - When update occurs
   - Then no jarring visual changes
   - And selection position preserved
   - And smooth re-render

3. **AC3: File Watcher Failure Warning**
   - Given dashboard is open
   - When file watcher fails
   - Then status bar shows warning: "File watching unavailable"
   - And manual refresh ([r]) still works

4. **AC4: Background Update Processing**
   - Given TUI is in background (unfocused)
   - When file events occur
   - Then updates still occur
   - And visible immediately when focused

5. **AC5: Activity Tracker Integration**
   - Given FileWatcher emits events
   - When event is received by TUI
   - Then project's LastActivityAt is updated:
     - Project identified by path prefix matching
     - Timestamp from event used
     - Repository updated via existing UpdateLastActivity()

6. **AC6: TUI Refresh After Activity Update**
   - Given activity is recorded
   - When next render cycle occurs
   - Then project's LastActivityAt reflects new activity
   - And relative time display updates (e.g., "just now")
   - And waiting state clears via WaitingDetector (if was waiting)

7. **AC7: Debounced Event Processing**
   - Given multiple file events fire rapidly (editor save burst)
   - When events are processed
   - Then debouncing occurs (200ms default from FileWatcher)
   - And single aggregated update triggers TUI refresh
   - And no UI flickering

8. **AC8: Channel-Based Communication Pattern**
   - Given FileWatcher watches project paths
   - When TUI initializes with FileWatcher
   - Then TUI subscribes to FileWatcher's event channel
   - And events are processed via Bubble Tea Cmd pattern
   - And goroutine properly handles context cancellation

9. **AC9: Project Path Mapping**
   - Given file event with path
   - When processing event
   - Then event path is mapped to correct project
   - Using path prefix matching against project.Path
   - And unmatched events are ignored (logged at debug level)

10. **AC10: Graceful Shutdown**
    - Given TUI is running with FileWatcher
    - When user quits (q/Ctrl+C)
    - Then FileWatcher.Close() is called
    - And context is cancelled
    - And no goroutine leaks

## File Structure

```
internal/core/ports/
    watcher.go                    # Existing: FileWatcher interface (Watch, Close only)

internal/core/services/
    activity_tracker.go           # Existing: ActivityTracker.ProcessEvents()

internal/adapters/filesystem/
    watcher.go                    # Existing: FsnotifyWatcher implementation

internal/adapters/tui/
    model.go                      # UPDATE: Add FileWatcher integration
    model_test.go                 # UPDATE: Add file watcher tests

internal/adapters/cli/
    add.go                        # UPDATE: Add fileWatcher variable and setter
    root.go                       # UPDATE: Pass FileWatcher to TUI

cmd/vibe/main.go                  # UPDATE: Create FsnotifyWatcher, inject to CLI
```

## Tasks / Subtasks

- [x] Task 1: Create FsnotifyWatcher in main.go (AC: 10) **[CRITICAL - Do First]**
  - [x] 1.1: Import `filesystem` package in main.go
  - [x] 1.2: Create `fileWatcher := filesystem.NewFsnotifyWatcher(time.Duration(cfg.RefreshDebounceMs) * time.Millisecond)`
  - [x] 1.3: Add `defer fileWatcher.Close()` for cleanup
  - [x] 1.4: Call `cli.SetFileWatcher(fileWatcher)` after creation
  - [x] 1.5: Log debug info about file watcher initialization

- [x] Task 2: Add FileWatcher to CLI Package (AC: 8, 10)
  - [x] 2.1: Add `fileWatcher ports.FileWatcher` variable to add.go (follows existing pattern)
  - [x] 2.2: Add `SetFileWatcher(watcher ports.FileWatcher)` function
  - [x] 2.3: Update `tui.Run()` signature to accept FileWatcher
  - [x] 2.4: Pass fileWatcher to `tui.Run()` in root.go

- [x] Task 3: Add FileWatcher Integration to TUI Model (AC: 1, 4, 8, 10)
  - [x] 3.1: Add `fileWatcher ports.FileWatcher` field to Model struct
  - [x] 3.2: Add `eventCh <-chan ports.FileEvent` field to store event channel
  - [x] 3.3: Add `watchCtx context.Context` and `watchCancel context.CancelFunc` fields
  - [x] 3.4: Add `fileWatcherAvailable bool` field (default true, set false on error)
  - [x] 3.5: Add `SetFileWatcher(watcher ports.FileWatcher)` method
  - [x] 3.6: Create `fileEventMsg` type for Bubble Tea message passing
  - [x] 3.7: Create `fileWatcherErrorMsg` type for error handling

- [x] Task 4: Implement File Watcher Startup (AC: 1, 8)
  - [x] 4.1: In `ProjectsLoadedMsg` handler, collect all project paths
  - [x] 4.2: Create watch context: `m.watchCtx, m.watchCancel = context.WithCancel(context.Background())`
  - [x] 4.3: Call `eventCh, err := m.fileWatcher.Watch(m.watchCtx, paths)`
  - [x] 4.4: Store channel: `m.eventCh = eventCh`
  - [x] 4.5: Handle error: set `m.fileWatcherAvailable = false`, log warning
  - [x] 4.6: Return `waitForNextFileEventCmd()` to start listening

- [x] Task 5: Implement Event Processing (AC: 5, 6, 9)
  - [x] 5.1: Create `waitForNextFileEventCmd()` that reads from `m.eventCh`
  - [x] 5.2: Handle `fileEventMsg` in Update() method
  - [x] 5.3: Create `handleFileEvent(msg fileEventMsg)` helper
  - [x] 5.4: Implement `findProjectByPath(eventPath string) *domain.Project`
  - [x] 5.5: Update project's LastActivityAt in repository via `repo.UpdateLastActivity()`
  - [x] 5.6: Update local m.projects slice to match
  - [x] 5.7: Recalculate status bar counts (waiting may have cleared)
  - [x] 5.8: Return `waitForNextFileEventCmd()` to continue listening

- [x] Task 6: Handle File Watcher Errors (AC: 3)
  - [x] 6.1: Handle `fileWatcherErrorMsg` in Update()
  - [x] 6.2: Set `m.fileWatcherAvailable = false`
  - [x] 6.3: Update status bar to show warning when unavailable
  - [x] 6.4: Ensure manual refresh ([r]) still works regardless

- [x] Task 7: Graceful Shutdown (AC: 10)
  - [x] 7.1: In quit handler (KeyQuit, KeyForceQuit), cancel watch context
  - [x] 7.2: Call `m.fileWatcher.Close()` before tea.Quit
  - [x] 7.3: Ensure waitForNextFileEventCmd() exits on context cancel

- [x] Task 8: Comprehensive Testing (AC: all)
  - [x] 8.1: Create MockFileWatcher for testing
  - [x] 8.2: Unit tests for fileEventMsg handling
  - [x] 8.3: Unit tests for path-to-project mapping
  - [x] 8.4: Unit tests for activity update flow
  - [x] 8.5: Test graceful shutdown with active watcher
  - [x] 8.6: Test fileWatcherAvailable flag on error

## Dev Notes

### Architecture Compliance (CRITICAL)

**Hexagonal Architecture Boundaries:**
```
cmd/vibe/main.go                  -> Creates FsnotifyWatcher, passes to CLI
internal/adapters/cli/add.go      -> Stores fileWatcher variable (existing pattern)
internal/adapters/cli/root.go     -> Passes FileWatcher to tui.Run()
internal/adapters/tui/model.go    -> Uses ports.FileWatcher interface
internal/adapters/filesystem/     -> FsnotifyWatcher implementation
internal/core/ports/watcher.go    -> FileWatcher interface definition
```

**Import Rules:**
- TUI model imports `internal/core/ports` for FileWatcher interface
- TUI model CANNOT import `internal/adapters/filesystem` directly
- main.go imports filesystem to create concrete FsnotifyWatcher

### ACTUAL FileWatcher Interface (ports/watcher.go)

**CRITICAL: Use the real interface - it does NOT have AddPath/RemovePath:**

```go
// FileWatcher defines the interface for monitoring file system changes.
type FileWatcher interface {
    // Watch starts monitoring the specified paths for file system changes.
    // Returns a channel that emits FileEvent for each detected change.
    Watch(ctx context.Context, paths []string) (<-chan FileEvent, error)

    // Close stops watching and releases all resources.
    Close() error
}

// FileEvent represents a file system change event.
type FileEvent struct {
    Path      string
    Operation FileOperation
    Timestamp time.Time
}
```

### ACTUAL ActivityTracker Usage Pattern

**CRITICAL: ActivityTracker.ProcessEvents() consumes the channel directly:**

```go
// From services/activity_tracker.go - ProcessEvents blocks on channel
func (t *ActivityTracker) ProcessEvents(ctx context.Context, events <-chan ports.FileEvent) {
    for {
        select {
        case <-ctx.Done():
            return
        case event, ok := <-events:
            if !ok { return }
            t.handleEvent(ctx, event)
        }
    }
}
```

**For Story 4.6, we have two options:**

**Option A (Simpler): TUI handles events directly, updates repo itself**
- TUI subscribes to FileWatcher channel
- TUI maps path to project, calls `repo.UpdateLastActivity(ctx, projectID, timestamp)`
- TUI updates local state and re-renders

**Option B: Use ActivityTracker in separate goroutine**
- ActivityTracker.ProcessEvents() runs in goroutine consuming events
- TUI polls repository periodically for changes
- More complex, less responsive

**Recommendation: Use Option A** - Direct TUI handling is simpler and more responsive.

### main.go Changes Required

```go
// In run() function, after waitingDetector creation:

// Story 4.6: Create FileWatcher for real-time dashboard updates
debounce := time.Duration(cfg.RefreshDebounceMs) * time.Millisecond
if debounce == 0 {
    debounce = filesystem.DefaultDebounce // 200ms
}
fileWatcher := filesystem.NewFsnotifyWatcher(debounce)
defer fileWatcher.Close()

slog.Debug("file watcher initialized", "debounce_ms", cfg.RefreshDebounceMs)

// Pass to CLI
cli.SetFileWatcher(fileWatcher)
```

### add.go Changes Required

```go
// Add to package variables (after waitingDetector):
var fileWatcher ports.FileWatcher

// Add setter function:
func SetFileWatcher(watcher ports.FileWatcher) {
    fileWatcher = watcher
}
```

### root.go Changes Required

```go
// Update tui.Run call:
if err := tui.Run(cmd.Context(), repository, detectionService, waitingDetector, fileWatcher); err != nil {
```

### app.go Signature Update

```go
// Update Run function signature:
func Run(ctx context.Context, repo ports.ProjectRepository, detector ports.Detector,
         waitingDetector ports.WaitingDetector, fileWatcher ports.FileWatcher) error {
    m := NewModel(repo)
    if detector != nil {
        m.SetDetectionService(detector)
    }
    if waitingDetector != nil {
        m.SetWaitingDetector(waitingDetector)
    }
    // Story 4.6: Wire file watcher
    if fileWatcher != nil {
        m.SetFileWatcher(fileWatcher)
    }
    // ... rest unchanged
}
```

### Model Changes Required

```go
type Model struct {
    // ... existing fields ...

    // Story 4.6: File watcher for real-time updates
    fileWatcher          ports.FileWatcher
    eventCh              <-chan ports.FileEvent
    watchCtx             context.Context
    watchCancel          context.CancelFunc
    fileWatcherAvailable bool // false if watcher failed to start
}

// SetFileWatcher sets the file watcher for real-time updates (Story 4.6).
func (m *Model) SetFileWatcher(watcher ports.FileWatcher) {
    m.fileWatcher = watcher
    m.fileWatcherAvailable = true // Assume available until proven otherwise
}
```

### Bubble Tea Event Pattern (Definitive)

```go
// Message types
type fileEventMsg struct {
    Path      string
    Operation ports.FileOperation
    Timestamp time.Time
}

type fileWatcherErrorMsg struct {
    err error
}

// In ProjectsLoadedMsg handler, after loading projects:
if m.fileWatcher != nil && len(m.projects) > 0 {
    // Collect project paths
    paths := make([]string, len(m.projects))
    for i, p := range m.projects {
        paths[i] = p.Path
    }

    // Create watch context for cancellation
    m.watchCtx, m.watchCancel = context.WithCancel(context.Background())

    // Start watching
    eventCh, err := m.fileWatcher.Watch(m.watchCtx, paths)
    if err != nil {
        slog.Warn("failed to start file watcher", "error", err)
        m.fileWatcherAvailable = false
    } else {
        m.eventCh = eventCh
        return m, m.waitForNextFileEventCmd()
    }
}

// waitForNextFileEventCmd waits for the next event from the stored channel
func (m Model) waitForNextFileEventCmd() tea.Cmd {
    if m.eventCh == nil {
        return nil
    }
    return func() tea.Msg {
        select {
        case <-m.watchCtx.Done():
            return nil // Context cancelled, stop listening
        case event, ok := <-m.eventCh:
            if !ok {
                return fileWatcherErrorMsg{err: fmt.Errorf("watcher channel closed")}
            }
            return fileEventMsg{
                Path:      event.Path,
                Operation: event.Operation,
                Timestamp: event.Timestamp,
            }
        }
    }
}

// In Update() switch:
case fileEventMsg:
    m.handleFileEvent(msg)
    // Re-subscribe to wait for next event
    return m, m.waitForNextFileEventCmd()

case fileWatcherErrorMsg:
    slog.Warn("file watcher error", "error", msg.err)
    m.fileWatcherAvailable = false
    return m, nil
```

### Event Handler Implementation

```go
func (m *Model) handleFileEvent(msg fileEventMsg) {
    // Find project by path prefix
    project := m.findProjectByPath(msg.Path)
    if project == nil {
        slog.Debug("event path not matched to project", "path", msg.Path)
        return
    }

    // Update repository
    ctx := context.Background()
    if err := m.repository.UpdateLastActivity(ctx, project.ID, msg.Timestamp); err != nil {
        slog.Warn("failed to update activity", "project_id", project.ID, "error", err)
        return
    }

    // Update local state
    project.LastActivityAt = msg.Timestamp

    // Update detail panel if this is selected project
    if m.detailPanel.Project() != nil && m.detailPanel.Project().ID == project.ID {
        m.detailPanel.SetProject(project)
    }

    // Recalculate status bar (waiting may have cleared)
    active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
    m.statusBar.SetCounts(active, hibernated, waiting)
}

func (m Model) findProjectByPath(eventPath string) *domain.Project {
    eventPath = strings.TrimSuffix(eventPath, "/")
    for _, p := range m.projects {
        projectPath := strings.TrimSuffix(p.Path, "/")
        if eventPath == projectPath || strings.HasPrefix(eventPath, projectPath+"/") {
            return p
        }
    }
    return nil
}
```

### Graceful Shutdown

```go
// In handleKeyMsg, update quit handlers:
case KeyQuit, KeyForceQuit:
    // Story 4.6: Clean up file watcher
    if m.watchCancel != nil {
        m.watchCancel()
    }
    if m.fileWatcher != nil {
        m.fileWatcher.Close()
    }
    return m, tea.Quit
```

### Status Bar Warning

```go
// In status bar rendering, check fileWatcherAvailable
// Add to StatusBarModel or handle in Model's View()
if !m.fileWatcherAvailable {
    // Show warning in feedback area: "⚠️ File watching unavailable"
}
```

### Previous Story Learnings (from Story 4.5)

Key patterns established that apply to 4.6:

1. **Callback Pattern for Components**: Components can't import interfaces, use callbacks
2. **Context Usage**: Use `context.Background()` in Bubble Tea since Update()/View() don't provide ctx
3. **Backward Compatibility**: Add new functions/methods rather than breaking existing ones
4. **Interface in Ports**: Define interfaces in `internal/core/ports` for dependency injection
5. **CLI Variable Pattern**: Store injected dependencies in add.go package variables

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Import `filesystem.FsnotifyWatcher` in TUI | Use `ports.FileWatcher` interface |
| Use `AddPath()`/`RemovePath()` methods | These don't exist - use `Watch()` with all paths |
| Block in Update() waiting for events | Use Cmd pattern for async operations |
| Create ActivityTracker goroutine | Handle events directly in TUI for simplicity |
| Skip context cancellation on shutdown | Always cancel watch context on quit |
| Forget to call FileWatcher.Close() | Add to shutdown sequence |

### Edge Cases to Test

| Scenario | Expected Behavior |
|----------|-------------------|
| No FileWatcher injected | TUI works normally, no file watching |
| FileWatcher fails to start | Warning shown, manual refresh works |
| Event for unknown path | Logged at debug level, ignored |
| Rapid file events | Debounced by FileWatcher (200ms) |
| TUI quit during watch | Context cancelled, channel closes cleanly |
| All projects removed | Stop watching (empty paths list) |
| Activity clears waiting state | WAITING indicator disappears on next render |
| Channel closes unexpectedly | fileWatcherAvailable = false, warning shown |

### References

| Document | Section | Relevance |
|----------|---------|-----------|
| docs/prd.md | FR27 | Auto-detect file system changes |
| docs/epics.md | Story 4.6 | Epic 4 story definition |
| docs/architecture.md | File Watcher Patterns | Debounce 200ms |
| docs/project-context.md | Hexagonal Architecture | Core never imports adapters |
| internal/core/ports/watcher.go | FileWatcher interface | **Actual interface (Watch, Close only)** |
| internal/adapters/filesystem/watcher.go | FsnotifyWatcher | Implementation reference |
| internal/core/services/activity_tracker.go | ActivityTracker | ProcessEvents pattern |
| internal/adapters/tui/model.go | Model struct | TUI state management |
| internal/adapters/cli/add.go | Package variables | Injection pattern |

### Manual Testing Steps

After implementation, verify:

1. **Basic File Watching:**
   ```bash
   ./bin/vibe add /path/to/project
   ./bin/vibe
   # In another terminal: touch /path/to/project/test.txt
   # Dashboard should update "Last Active" within 5-10 seconds
   ```

2. **Waiting State Clears:**
   ```bash
   ./bin/vibe --waiting-threshold=1
   # Wait 1+ minute until project shows WAITING
   # Touch a file in the project
   # WAITING indicator should disappear
   ```

3. **Selection Preserved:**
   ```bash
   ./bin/vibe
   # Select a project (j/k)
   # Touch a file in different project
   # Selection should remain on originally selected project
   ```

4. **File Watcher Failure:**
   ```bash
   # Test by passing invalid config or mocking Watch() to return error
   # Should see warning in status bar
   # Manual refresh (r) should still work
   ```

5. **Graceful Shutdown:**
   ```bash
   ./bin/vibe
   # Press q
   # Should exit cleanly, no goroutine leak messages
   ```

6. **Multiple Rapid Events:**
   ```bash
   ./bin/vibe
   # Run: for i in {1..10}; do touch /path/to/project/test$i.txt; done
   # Should see single update, no flickering
   ```

### Downstream Dependencies

**Epic 5 (Hibernation) depends on this story for:**
- Auto-activation when activity detected in hibernated project
- File watching to trigger state transitions

**Epic 7 (Error Handling) builds on:**
- FileWatcher error recovery patterns
- Warning display in status bar

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Implementation completed successfully

### Completion Notes List

1. **Task 1 (main.go)**: Created FileWatcher in run() with debounce from config, added defer Close(), injected via cli.SetFileWatcher()
2. **Task 2 (CLI)**: Added fileWatcher variable and SetFileWatcher() to add.go, updated tui.Run() signature in app.go, passed watcher in root.go
3. **Task 3 (Model struct)**: Added fileWatcher, eventCh, watchCtx, watchCancel, fileWatcherAvailable fields; created fileEventMsg and fileWatcherErrorMsg types
4. **Task 4 (Watcher startup)**: In ProjectsLoadedMsg handler, collect paths, create context, call Watch(), store channel, return waitForNextFileEventCmd()
5. **Task 5 (Event processing)**: Implemented waitForNextFileEventCmd(), handleFileEvent(), findProjectByPath() with prefix matching
6. **Task 6 (Error handling)**: fileWatcherErrorMsg handler sets flag false, calls statusBar.SetWatcherWarning()
7. **Task 7 (Shutdown)**: Both KeyQuit and KeyForceQuit handlers cancel context and call Close(); also added to validation mode quit handler
8. **Task 8 (Tests)**: Added mockFileWatcher, tests for SetFileWatcher, findProjectByPath variations, fileEventMsg, fileWatcherErrorMsg, graceful shutdown

**Note**: Implementation was interrupted by macOS crash. Code review verified all ACs implemented correctly. Story file updated post-review.

### File List

| File | Change Type | Description |
|------|-------------|-------------|
| cmd/vibe/main.go | Modified | Create FsnotifyWatcher, defer Close(), call cli.SetFileWatcher() |
| internal/adapters/cli/add.go | Modified | Add fileWatcher variable and SetFileWatcher() function |
| internal/adapters/cli/root.go | Modified | Pass fileWatcher to tui.Run() |
| internal/adapters/tui/app.go | Modified | Update Run() signature to accept FileWatcher parameter |
| internal/adapters/tui/model.go | Modified | Add FileWatcher integration: fields, message types, handlers, event processing |
| internal/adapters/tui/model_test.go | Modified | Add comprehensive FileWatcher tests (mockFileWatcher, event handling, shutdown) |
| internal/adapters/tui/components/detail_panel.go | Modified | Add Project() getter for file event handling |
| internal/adapters/tui/components/status_bar.go | Modified | Add SetWatcherWarning() and watcher warning display |
| internal/adapters/tui/components/status_bar_test.go | Modified | Add watcher warning tests |

### Change Log

| Date | Author | Change |
|------|--------|--------|
| 2025-12-19 | SM (Bob) | Initial story creation via *create-story workflow (YOLO mode) |
| 2025-12-19 | SM (Bob) | Validation: Fixed 6 critical issues - corrected FileWatcher interface (removed AddPath/RemovePath), fixed ActivityTracker API (ProcessEvents not RecordActivity), added concrete main.go/CLI wiring code, fixed tui.Run signature update pattern, added event channel storage pattern, streamlined Bubble Tea event handling to single definitive approach |
| 2025-12-20 | Dev Agent (Amelia) | Implementation of all tasks (interrupted by macOS crash) |
| 2025-12-20 | Code Review | Verified implementation against all 10 ACs, marked tasks complete, filled Dev Agent Record |
