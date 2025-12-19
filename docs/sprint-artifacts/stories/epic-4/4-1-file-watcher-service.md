# Story 4.1: File Watcher Service

Status: Done

## Story

As a **system**,
I want **to monitor file changes in tracked projects**,
so that **I can detect activity and inactivity for the "Agent Waiting" detection feature**.

## Acceptance Criteria

1. **AC1: FileWatcher Implementation**
   - Given projects are tracked
   - When FileWatcher starts
   - Then fsnotify watches are created for each project's root directory
   - And key subdirectories (specs/, .bmad/, src/) are watched where they exist

2. **AC2: FileEvent Emission**
   - Given FileWatcher is running
   - When a file is created, modified, or deleted in a watched directory
   - Then a FileEvent is emitted with:
     - Path (canonical absolute path)
     - Operation (create/modify/delete)
     - Timestamp

3. **AC3: Event Debouncing**
   - Given multiple file events fire rapidly (e.g., editor save sequence)
   - When events occur within the debounce window (200ms default)
   - Then events are aggregated/debounced
   - And a single event is emitted after the debounce period of silence

4. **AC4: Watch Failure Handling**
   - Given FileWatcher encounters an error on a specific path
   - When watch fails on that path
   - Then error is logged with `slog.Error`
   - And other watches continue operating
   - And the failed path is not silently dropped

5. **AC5: Dynamic Watch Registration**
   - Given a new project is added while watcher is running
   - When `vibe add .` is executed
   - Then watch is registered for the new project automatically

6. **AC6: Dynamic Watch Unregistration**
   - Given a project is removed while watcher is running
   - When `vibe remove <project>` is executed
   - Then watch is unregistered for the removed project
   - And resources are cleaned up

7. **AC7: Graceful Shutdown**
   - Given FileWatcher is running
   - When application exits (Ctrl+C or SIGTERM)
   - Then all watches are cleaned up gracefully
   - And the event channel is closed
   - And no goroutine leaks occur

8. **AC8: Context Cancellation**
   - Given FileWatcher is watching paths
   - When the context is cancelled
   - Then Watch stops emitting events
   - And cleanup happens within 5 seconds

## Tasks / Subtasks

- [x] Task 1: Implement FsnotifyWatcher adapter (AC: 1, 2, 7, 8)
  - [x] 1.1: Create `internal/adapters/filesystem/watcher.go`
  - [x] 1.2: Implement `FsnotifyWatcher` struct implementing `ports.FileWatcher`
  - [x] 1.3: Implement `Watch(ctx, paths)` method with fsnotify setup
  - [x] 1.4: Implement `Close()` method for graceful cleanup
  - [x] 1.5: Handle context cancellation properly

- [x] Task 2: Implement debounced event handling (AC: 3)
  - [x] 2.1: Debounce logic integrated into FsnotifyWatcher (not separate wrapper)
  - [x] 2.2: Implement timer-based debounce logic (200ms default)
  - [x] 2.3: Aggregate rapid events into single emission per path
  - [x] 2.4: Make debounce window configurable via constructor

- [x] Task 3: Implement error handling and logging (AC: 4)
  - [x] 3.1: Log watch failures with `slog.Error` including path
  - [x] 3.2: Continue operating other watches on partial failure
  - [x] 3.3: Return meaningful errors from Watch() for caller handling

- [x] Task 4: Implement dynamic watch management (AC: 5, 6)
  - [x] 4.1: Add `AddPath(path string) error` method
  - [x] 4.2: Add `RemovePath(path string) error` method
  - [x] 4.3: Thread-safe path management with mutex

- [x] Task 5: Write comprehensive tests (AC: all)
  - [x] 5.1: Unit tests for FsnotifyWatcher (19 test functions, 21+ test cases with subtests)
  - [x] 5.2: Debounce tests included in unit tests
  - [x] 5.3: Integration tests with real filesystem changes
  - [x] 5.4: Test context cancellation behavior
  - [x] 5.5: Test graceful shutdown

## Dev Notes

### Architecture Compliance

**Hexagonal Architecture Boundaries:**
- Implementation: `internal/adapters/filesystem/watcher.go`
- Interface: `ports.FileWatcher` (defined in `internal/core/ports/watcher.go`)
- Core services consume via dependency injection
- **CRITICAL:** Adapter imports core (ports), core NEVER imports adapter

**Existing Interface (DO NOT MODIFY):**
```go
// internal/core/ports/watcher.go - ALREADY EXISTS
type FileWatcher interface {
    Watch(ctx context.Context, paths []string) (<-chan FileEvent, error)
    Close() error
}

type FileEvent struct {
    Path      string        // Canonical absolute path
    Operation FileOperation // FileOpCreate, FileOpModify, FileOpDelete
    Timestamp time.Time
}

type FileOperation int
const (
    FileOpCreate FileOperation = iota
    FileOpModify
    FileOpDelete
)
```

The `AddPath` and `RemovePath` methods for dynamic registration are implementation details - NOT part of the interface. Core services will call `Watch()` again with updated path list, or implementations can expose these as optional enhancement.

### Implementation Patterns

**Constructor Pattern (follow existing code):**
```go
// Follow NewDirectoryManager pattern from directory.go:34
func NewFsnotifyWatcher(debounce time.Duration) *FsnotifyWatcher {
    if debounce == 0 {
        debounce = 200 * time.Millisecond // Default from architecture
    }
    return &FsnotifyWatcher{
        debounce: debounce,
    }
}
```

**Context Propagation (MANDATORY - follow existing patterns):**
```go
// Pattern from directory.go:53-59
func (w *FsnotifyWatcher) Watch(ctx context.Context, paths []string) (<-chan FileEvent, error) {
    // MUST check context at start
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    // ... implementation
}
```

**Debounce Pattern (from Architecture doc):**
```go
type DebouncedWatcher struct {
    watcher  *fsnotify.Watcher
    debounce time.Duration
    timer    *time.Timer
    mu       sync.Mutex
    pending  map[string]FileEvent // Path -> most recent event
}

func (d *DebouncedWatcher) handleEvent(event fsnotify.Event) {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Track event by path
    d.pending[event.Name] = translateEvent(event)

    if d.timer != nil {
        d.timer.Stop()
    }
    d.timer = time.AfterFunc(d.debounce, func() {
        d.mu.Lock()
        defer d.mu.Unlock()
        for _, evt := range d.pending {
            // Emit to output channel
        }
        d.pending = make(map[string]FileEvent)
    })
}
```

**Error Handling (log once at handling site):**
```go
// Follow slog pattern from existing codebase
if err := watcher.Add(path); err != nil {
    slog.Error("failed to watch path", "path", path, "error", err)
    // Continue with other paths - graceful degradation
}
```

**Domain Error Usage:**
```go
// Use domain.ErrPathNotAccessible for inaccessible paths
// See internal/core/domain/errors.go:11
if _, err := os.Stat(path); err != nil {
    return nil, fmt.Errorf("%w: %s: %v", domain.ErrPathNotAccessible, path, err)
}
```

### Technical Requirements

**Library:** github.com/fsnotify/fsnotify
- Already in go.mod as indirect dependency (via viper)
- Add as direct dependency: `go get github.com/fsnotify/fsnotify`

**Configuration Values:**

| Setting | Default | Config Key | Source |
|---------|---------|------------|--------|
| Debounce window | 200ms | `refresh_debounce_ms` | Architecture doc |
| Buffer size | 100 | (internal) | Prevent slow consumer blocking |

**IMPORTANT: Debounce Conflict Resolution**
- `project-context.md` states 5-10 second debounce
- `architecture.md` specifies 200ms debounce for file events
- **Use 200ms** - the 5-10s applies to UI refresh, not event emission

**Key Subdirectories to Watch:**
- `specs/` - Speckit methodology artifacts
- `.bmad/` - BMAD methodology artifacts
- `src/` - Source code changes

**fsnotify Limitations:**
- Does NOT support recursive watching natively
- MVP: Watch root + key subdirectories explicitly
- Use `watcher.Add()` for each directory individually
- Post-MVP: Consider recursive polling or platform-specific solutions

### Code Reuse

**Existing filesystem package patterns (MUST FOLLOW):**
- Package: `github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem`
- Use `CanonicalPath()` from `paths.go:37` for path normalization
- Follow error wrapping pattern from `paths.go:27`
- Use `slog` for logging (consistent with rest of codebase)

**Existing Test Patterns:**
- Table-driven tests (see `paths_test.go`)
- Integration tests use `//go:build integration` tag
- Co-locate test files with source

### File Structure

```
internal/adapters/filesystem/
    watcher.go                      # FsnotifyWatcher + DebouncedWatcher
    watcher_test.go                 # Unit tests
    watcher_integration_test.go     # Integration tests (build tag)
```

**Alignment with Existing Files:**
- `directory.go` - Follow constructor and method patterns
- `paths.go` - Use CanonicalPath() for path normalization
- `paths_test.go` - Follow table-driven test structure

### Previous Story Intelligence

**Key Learnings from Epic 3.5 to Apply:**
1. **Single source of truth** - Same code path for all consumers of file events
2. **Context cancellation at start** - Every method checks context first
3. **Graceful degradation** - Log warnings, don't crash on partial failures
4. **Empty slice return (not nil)** - For JSON compatibility in future
5. **Integration tests separate** - Use build tag `//go:build integration`

**Pattern from RepositoryCoordinator (Epic 3.5):**
- Clean interface implementation
- Thread-safe with mutex where needed
- Proper resource cleanup in Close()
- Idempotent Close() - safe to call multiple times

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Create watcher in core/ | Keep in adapters/filesystem/ |
| Import fsnotify in core | Core only uses ports.FileWatcher |
| Log errors at every layer | Log once at handling site |
| Return nil channel on error | Return nil, error |
| Skip context check | Always check ctx.Done() first |
| Block on slow consumer | Use buffered channel (100) |
| Forget Close() cleanup | Defer watcher.Close() and stop timers |

### Goroutine Safety

**Event Processing Loop:**
```go
func (w *FsnotifyWatcher) eventLoop(ctx context.Context, out chan<- FileEvent) {
    defer close(out)
    defer w.watcher.Close()

    for {
        select {
        case <-ctx.Done():
            return
        case event, ok := <-w.watcher.Events:
            if !ok {
                return
            }
            w.handleEvent(event, out)
        case err, ok := <-w.watcher.Errors:
            if !ok {
                return
            }
            slog.Error("fsnotify error", "error", err)
        }
    }
}
```

**Shutdown Sequence (from Architecture):**
1. Receive signal (SIGINT/SIGTERM) or context cancellation
2. Stop timers in DebouncedWatcher
3. Close fsnotify watcher
4. Close output channel
5. All within 5 second timeout

### References

| Document | Section | Relevance |
|----------|---------|-----------|
| docs/architecture.md | File Watcher Patterns | Debounce implementation |
| docs/architecture.md | Graceful Shutdown Pattern | Cleanup sequence |
| docs/architecture.md | Project Directory Structure | File locations |
| docs/prd.md | FR27 | Auto-detect file system changes |
| docs/prd.md | FR34-38 | Agent waiting detection (depends on this) |
| docs/epics.md | Story 4.1-4.6 | Epic 4 context and dependencies |
| docs/project-context.md | File Watching | OS abstraction |
| internal/core/ports/watcher.go | FileWatcher interface | Interface to implement |
| internal/adapters/filesystem/directory.go | NewDirectoryManager | Constructor pattern |
| internal/adapters/filesystem/paths.go | CanonicalPath | Path normalization |

### Manual Testing Steps

After implementation, verify:

1. **Basic Watch:**
   ```bash
   # Create test project and watch
   mkdir -p /tmp/test-watch/src
   # In test code, watch /tmp/test-watch
   touch /tmp/test-watch/test.txt  # Should emit create event
   echo "hello" >> /tmp/test-watch/test.txt  # Should emit modify event
   rm /tmp/test-watch/test.txt  # Should emit delete event
   ```

2. **Debounce:**
   ```bash
   # Rapid writes should emit single event
   for i in {1..10}; do echo $i >> /tmp/test-watch/test.txt; done
   # Should see ONE event after debounce period, not 10
   ```

3. **Subdirectory Watch:**
   ```bash
   touch /tmp/test-watch/src/main.go  # Should emit event
   ```

4. **Graceful Shutdown:**
   ```bash
   # Start watcher, then Ctrl+C
   # No goroutine leaks, clean exit
   # Verify with: go test -race
   ```

5. **Error Handling:**
   ```bash
   # Try to watch non-existent path
   # Should log error and continue with other paths
   ```

### Downstream Dependencies

**Story 4.2 depends on this story for:**
- FileEvent emission to update `project.LastActivityAt`
- Watcher running continuously for activity tracking

**Story 4.3 depends on this story for:**
- Inactivity detection based on absence of file events
- WAITING state calculation from last activity

**Story 4.6 depends on this story for:**
- Real-time dashboard updates via Bubble Tea Msg

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- Race condition test passed: `go test -race ./internal/adapters/filesystem/...`
- All 19 unit test functions pass (21+ test cases with subtests)
- Integration tests created with `//go:build integration` tag

### Completion Notes List

- Implemented `FsnotifyWatcher` struct implementing `ports.FileWatcher` interface
- Debounce logic integrated directly into watcher (200ms default, configurable)
- Events aggregated by path during debounce window - only last event per path emitted
- Context cancellation properly handled at Watch() start and in event loop
- Close() is idempotent - safe to call multiple times
- AddPath() and RemovePath() methods for dynamic watch management
- Thread-safe with mutex protection for shared state
- Graceful degradation on partial path failures (logs error, continues with valid paths)
- Event channel buffered (100) to prevent blocking on slow consumers
- Uses CanonicalPath() for path normalization (symlink resolution)
- All domain errors wrapped with domain.ErrPathNotAccessible

### File List

- `internal/adapters/filesystem/watcher.go` (NEW) - FsnotifyWatcher implementation
- `internal/adapters/filesystem/watcher_test.go` (NEW) - Unit tests (20 test cases)
- `internal/adapters/filesystem/watcher_integration_test.go` (NEW) - Integration tests

### Change Log

| Date | Author | Change |
|------|--------|--------|
| 2025-12-19 | SM (Bob) | Initial story creation via *create-story workflow |
| 2025-12-19 | SM (Bob) | Story validation and improvements applied |
| 2025-12-19 | Dev Agent (Claude Opus 4.5) | Implemented FsnotifyWatcher with debouncing, error handling, dynamic watch management, and comprehensive tests |
| 2025-12-19 | Dev Agent (Claude Opus 4.5) | Code Review: Fixed race condition in flushPending, added closed check, improved documentation and comments |
