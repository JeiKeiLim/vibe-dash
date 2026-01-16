# Story 15.5: Implement Generic File-Activity Fallback Detector

Status: done

## Story

As a user,
I want agent detection to work for non-Claude-Code projects,
So that I still get waiting detection (even if less accurate) for any project.

## User-Visible Changes

None - this is internal infrastructure completing the generic fallback detector. User-visible changes will come in Story 15.6 when agent detection is integrated into the TUI dashboard.

## Acceptance Criteria

1. **AC1:** Given a project without Claude Code logs, generic detector can detect agent state from file activity
2. **AC2:** Given no file activity for 10+ minutes (configurable), detection returns `WaitingForUser` with `ConfidenceUncertain`
3. **AC3:** Given recent file activity (< threshold), detection returns `Working` with `ConfidenceUncertain`
4. **AC4:** GenericDetector implements `AgentActivityDetector` interface from Story 15.1
5. **AC5:** Detection threshold defaults to 10 minutes but can be configured via functional options
6. **AC6:** Detection respects context cancellation, returning within 100ms when cancelled
7. **AC7:** Given project path doesn't exist or is inaccessible, detection returns `AgentUnknown` with `ConfidenceUncertain` (no error)
8. **AC8:** Duration is calculated from most recent file modification time in project directory

## Tasks / Subtasks

- [x] Task 1: Create GenericDetector struct (AC: 4, 5)
  - [x] 1.1: Create `internal/adapters/agentdetectors/generic_detector.go`
  - [x] 1.2: Define struct with `threshold time.Duration` field (default 10 minutes)
  - [x] 1.3: Implement `NewGenericDetector(opts ...GenericDetectorOption)` constructor with functional options
  - [x] 1.4: Add `WithThreshold(d time.Duration)` option for configurable threshold
  - [x] 1.5: Add `now func() time.Time` field for testing (default `time.Now`)
  - [x] 1.6: Add `WithNow(fn func() time.Time)` option for testing

- [x] Task 2: Implement Name() method (AC: 4)
  - [x] 2.1: Return "Generic" constant string

- [x] Task 3: Implement Detect() method (AC: 1, 2, 3, 6, 7, 8)
  - [x] 3.1: Check `ctx.Done()` before starting (respect cancellation)
  - [x] 3.2: Call `findMostRecentModification(ctx, projectPath)` to find latest file modification
  - [x] 3.3: If error or zero time returned → return `AgentUnknown` with `ConfidenceUncertain` (no error)
  - [x] 3.4: Calculate duration: `now() - mostRecentModTime`
  - [x] 3.5: If duration >= threshold → return `AgentWaitingForUser` with `ConfidenceUncertain`
  - [x] 3.6: Otherwise → return `AgentWorking` with `ConfidenceUncertain`

- [x] Task 4: Implement findMostRecentModification helper (AC: 8)
  - [x] 4.1: Create private `findMostRecentModification(ctx context.Context, dir string) (time.Time, error)` method
  - [x] 4.2: Walk directory tree using `filepath.WalkDir`
  - [x] 4.3: Track most recent modification time (compare each file's ModTime)
  - [x] 4.4: Skip directories and hidden files (starting with `.`) - use `slog.Debug` for visibility
  - [x] 4.5: Check `ctx.Done()` periodically during walk (every 100 files to balance responsiveness vs overhead)
  - [x] 4.6: Return zero time if directory doesn't exist or has no files
  - [x] 4.7: Handle permission errors gracefully (skip inaccessible files with `slog.Debug`, continue scanning)

- [x] Task 5: Write comprehensive unit tests (AC: 1-8)
  - [x] 5.1: Create `internal/adapters/agentdetectors/generic_detector_test.go`
  - [x] 5.2: Test `Name()` returns "Generic"
  - [x] 5.3: Test `Detect()` with recent activity returns Working
  - [x] 5.4: Test `Detect()` with old activity returns WaitingForUser
  - [x] 5.5: Test `Detect()` with non-existent path returns Unknown (no error)
  - [x] 5.6: Test `Detect()` with cancelled context returns promptly (< 100ms)
  - [x] 5.7: Test custom threshold via `WithThreshold` option
  - [x] 5.8: Test Duration calculation is approximately correct (within 1 second tolerance)
  - [x] 5.9: Test hidden files are skipped
  - [x] 5.10: Test Confidence is always Uncertain

- [x] Task 6: Verify integration and compliance
  - [x] 6.1: Ensure GenericDetector satisfies `ports.AgentActivityDetector` interface (compile-time check)
  - [x] 6.2: Run `make lint && make test` - all must pass
  - [x] 6.3: Update `doc.go` to add: `- GenericDetector (Story 15.5): File activity fallback for any project with low confidence`

## Dev Notes

### Performance Requirement

Per NFR-P2-1 and NFR-P1, detection should complete within 1 second for typical projects. The filesystem walk is the primary cost:
- For projects with < 1,000 files: Sub-100ms (no concern)
- For projects with 1,000-10,000 files: 100-500ms (acceptable)
- For projects with > 10,000 files: May exceed 1 second

**Mitigation:** Context cancellation check every 100 files enables early termination. The TUI can set a detection timeout to prevent UI blocking.

### Detection Flow Overview

```
1. Check context cancellation → return early if cancelled
2. findMostRecentModification(projectPath) → get latest file mtime
   - Zero time → AgentUnknown (path doesn't exist or no files)
   - Error → AgentUnknown
3. Calculate duration: now - mostRecentModTime
4. If duration >= threshold → AgentWaitingForUser
5. Otherwise → AgentWorking
6. ALL states return ConfidenceUncertain (heuristic-based)
```

### Required Imports

```go
import (
    "context"
    "io/fs"
    "log/slog"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)
```

### Struct Definition

```go
// GenericDetector detects agent activity state using file modification times.
// This is a FALLBACK detector when tool-specific logs are unavailable.
// Implements ports.AgentActivityDetector interface.
type GenericDetector struct {
    threshold time.Duration   // Inactivity threshold (default 10 minutes)
    now       func() time.Time // For testing (default time.Now)
}

// GenericDetectorOption is a functional option for configuring GenericDetector.
type GenericDetectorOption func(*GenericDetector)

// Default threshold matches existing WaitingDetector behavior
const defaultThreshold = 10 * time.Minute

// WithThreshold sets a custom inactivity threshold.
func WithThreshold(d time.Duration) GenericDetectorOption {
    return func(g *GenericDetector) {
        if d > 0 {
            g.threshold = d
        }
    }
}

// WithNow sets a custom time function (for testing).
func WithNow(fn func() time.Time) GenericDetectorOption {
    return func(g *GenericDetector) {
        if fn != nil {
            g.now = fn
        }
    }
}

// NewGenericDetector creates a new detector with optional configuration.
func NewGenericDetector(opts ...GenericDetectorOption) *GenericDetector {
    g := &GenericDetector{
        threshold: defaultThreshold,
        now:       time.Now,
    }
    for _, opt := range opts {
        opt(g)
    }
    return g
}

// Compile-time interface compliance check
var _ ports.AgentActivityDetector = (*GenericDetector)(nil)
```

### Detect Method Implementation

```go
const detectorName = "Generic"

func (g *GenericDetector) Name() string {
    return detectorName
}

func (g *GenericDetector) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
    // Respect context cancellation at entry
    select {
    case <-ctx.Done():
        return domain.NewAgentState(detectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
    default:
    }

    // Find most recent file modification
    mostRecentModTime, err := g.findMostRecentModification(ctx, projectPath)
    if err != nil || mostRecentModTime.IsZero() {
        // Path doesn't exist, inaccessible, or no files - graceful unknown
        return domain.NewAgentState(detectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
    }

    // Calculate duration since last activity
    duration := g.now().Sub(mostRecentModTime)

    // Handle future timestamps (clock skew)
    if duration < 0 {
        duration = 0
    }

    // Determine state based on threshold
    if duration >= g.threshold {
        return domain.NewAgentState(detectorName, domain.AgentWaitingForUser, duration, domain.ConfidenceUncertain), nil
    }
    return domain.NewAgentState(detectorName, domain.AgentWorking, duration, domain.ConfidenceUncertain), nil
}
```

### findMostRecentModification Implementation

```go
func (g *GenericDetector) findMostRecentModification(ctx context.Context, dir string) (time.Time, error) {
    // Check context at entry
    select {
    case <-ctx.Done():
        return time.Time{}, nil
    default:
    }

    // Check if path exists
    info, err := os.Stat(dir)
    if err != nil {
        slog.Debug("path stat failed", "path", dir, "error", err)
        return time.Time{}, err
    }
    if !info.IsDir() {
        // Single file - return its mtime
        return info.ModTime(), nil
    }

    var mostRecent time.Time
    filesChecked := 0

    err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        // Skip errors (permission denied, etc.) with debug logging
        if err != nil {
            slog.Debug("skipping path due to error", "path", path, "error", err)
            return nil // Continue walking
        }

        // Check context periodically (every 100 files balances responsiveness vs syscall overhead)
        filesChecked++
        if filesChecked%100 == 0 {
            select {
            case <-ctx.Done():
                slog.Debug("context cancelled during walk", "filesChecked", filesChecked)
                return filepath.SkipAll
            default:
            }
        }

        // Skip directories
        if d.IsDir() {
            // Skip hidden directories (except root)
            // Note: On Windows, hidden attribute detection would need OS abstraction (post-MVP)
            if path != dir && strings.HasPrefix(d.Name(), ".") {
                slog.Debug("skipping hidden directory", "path", path)
                return filepath.SkipDir
            }
            return nil
        }

        // Skip hidden files
        if strings.HasPrefix(d.Name(), ".") {
            slog.Debug("skipping hidden file", "path", path)
            return nil
        }

        // Get file info for modification time
        info, err := d.Info()
        if err != nil {
            slog.Debug("skipping inaccessible file", "path", path, "error", err)
            return nil // Skip inaccessible files
        }

        if info.ModTime().After(mostRecent) {
            mostRecent = info.ModTime()
        }

        return nil
    })

    if err != nil {
        return time.Time{}, err
    }

    return mostRecent, nil
}
```

### Detection Logic Summary

| Condition | Agent State | Confidence |
|-----------|-------------|------------|
| No file activity for >= threshold | WaitingForUser | Uncertain |
| Recent file activity (< threshold) | Working | Uncertain |
| Path doesn't exist | Unknown | Uncertain |
| No accessible files | Unknown | Uncertain |
| Context cancelled | Unknown | Uncertain |

### Edge Cases

| Case | Input | Behavior |
|------|-------|----------|
| Project path doesn't exist | `/nonexistent/path` | Return AgentUnknown, ConfidenceUncertain, nil error |
| Empty directory | Valid dir, no files | Return AgentUnknown (mostRecent.IsZero()) |
| Single file path | `/path/to/file.go` | Return state based on file's mtime |
| Hidden files only | `.git/`, `.env` | Return AgentUnknown (skipped) |
| Permission denied | Can't read some files | Skip those files, continue with accessible ones |
| Context cancelled | Any point | Return AgentUnknown within 100ms |
| Future timestamp | File mtime in future | Clamp duration to 0, return Working |
| Threshold = 0 | Invalid config | Use default 10 minutes |
| Negative threshold | Invalid config | Use default 10 minutes |

### Testing Strategy

```go
func TestName(t *testing.T) {
    g := NewGenericDetector()
    if got := g.Name(); got != "Generic" {
        t.Errorf("Name() = %q, want %q", got, "Generic")
    }
}

func TestDetect_RecentActivity_Working(t *testing.T) {
    tmpDir := t.TempDir()
    // Create file with recent mtime (now)
    os.WriteFile(filepath.Join(tmpDir, "recent.go"), []byte(""), 0644)

    g := NewGenericDetector(WithThreshold(10 * time.Minute))
    state, err := g.Detect(context.Background(), tmpDir)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if state.Status != domain.AgentWorking {
        t.Errorf("Status = %v, want AgentWorking", state.Status)
    }
    if state.Confidence != domain.ConfidenceUncertain {
        t.Errorf("Confidence = %v, want ConfidenceUncertain", state.Confidence)
    }
}

func TestDetect_OldActivity_WaitingForUser(t *testing.T) {
    tmpDir := t.TempDir()
    filePath := filepath.Join(tmpDir, "old.go")
    os.WriteFile(filePath, []byte(""), 0644)

    // Set file mtime to 15 minutes ago
    oldTime := time.Now().Add(-15 * time.Minute)
    os.Chtimes(filePath, oldTime, oldTime)

    g := NewGenericDetector(WithThreshold(10 * time.Minute))
    state, err := g.Detect(context.Background(), tmpDir)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if state.Status != domain.AgentWaitingForUser {
        t.Errorf("Status = %v, want AgentWaitingForUser", state.Status)
    }
}

func TestDetect_NonexistentPath_Unknown(t *testing.T) {
    g := NewGenericDetector()
    state, err := g.Detect(context.Background(), "/nonexistent/path/12345")

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if state.Status != domain.AgentUnknown {
        t.Errorf("Status = %v, want AgentUnknown", state.Status)
    }
}

func TestDetect_ContextCancelled_ReturnsPromptly(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately

    g := NewGenericDetector()

    start := time.Now()
    state, err := g.Detect(ctx, "/some/path")
    elapsed := time.Since(start)

    if elapsed > 100*time.Millisecond {
        t.Errorf("Detect took %v, want < 100ms", elapsed)
    }
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    if state.Status != domain.AgentUnknown {
        t.Errorf("Status = %v, want AgentUnknown", state.Status)
    }
}

func TestDetect_CustomThreshold(t *testing.T) {
    tmpDir := t.TempDir()
    filePath := filepath.Join(tmpDir, "test.go")
    os.WriteFile(filePath, []byte(""), 0644)

    // Set file mtime to 3 minutes ago
    oldTime := time.Now().Add(-3 * time.Minute)
    os.Chtimes(filePath, oldTime, oldTime)

    // Default threshold (10 min) → Working
    g1 := NewGenericDetector()
    state1, _ := g1.Detect(context.Background(), tmpDir)
    if state1.Status != domain.AgentWorking {
        t.Errorf("Default threshold: Status = %v, want AgentWorking", state1.Status)
    }

    // Custom threshold (2 min) → WaitingForUser
    g2 := NewGenericDetector(WithThreshold(2 * time.Minute))
    state2, _ := g2.Detect(context.Background(), tmpDir)
    if state2.Status != domain.AgentWaitingForUser {
        t.Errorf("Custom threshold: Status = %v, want AgentWaitingForUser", state2.Status)
    }
}

func TestDetect_HiddenFilesSkipped(t *testing.T) {
    tmpDir := t.TempDir()

    // Create only hidden files
    os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte(""), 0644)
    os.Mkdir(filepath.Join(tmpDir, ".git"), 0755)
    os.WriteFile(filepath.Join(tmpDir, ".git", "HEAD"), []byte(""), 0644)

    g := NewGenericDetector()
    state, err := g.Detect(context.Background(), tmpDir)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // Should be Unknown since all files are hidden
    if state.Status != domain.AgentUnknown {
        t.Errorf("Status = %v, want AgentUnknown (hidden files skipped)", state.Status)
    }
}

func TestDetect_WithNow(t *testing.T) {
    tmpDir := t.TempDir()
    os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte(""), 0644)

    // Mock time to be 15 minutes in the future
    mockNow := func() time.Time {
        return time.Now().Add(15 * time.Minute)
    }

    g := NewGenericDetector(WithNow(mockNow), WithThreshold(10*time.Minute))
    state, _ := g.Detect(context.Background(), tmpDir)

    if state.Status != domain.AgentWaitingForUser {
        t.Errorf("Status = %v, want AgentWaitingForUser", state.Status)
    }
}

// Interface compliance test (compile-time)
var _ ports.AgentActivityDetector = (*GenericDetector)(nil)
```

### Hexagonal Architecture

```
internal/adapters/agentdetectors/
├── doc.go                             # Package documentation (update for 15.5)
├── claude_code_path_matcher.go        # Path matching (Story 15.2)
├── claude_code_path_matcher_test.go   # Tests (Story 15.2)
├── claude_code_log_parser.go          # JSONL parsing (Story 15.3)
├── claude_code_log_parser_test.go     # Tests (Story 15.3)
├── claude_code_detector.go            # Claude Code detector (Story 15.4)
├── claude_code_detector_test.go       # Tests (Story 15.4)
├── generic_detector.go                # NEW: Generic fallback detector (this story)
└── generic_detector_test.go           # NEW: Tests (this story)
```

**Architecture Position:**
- GenericDetector is an ADAPTER implementing `ports.AgentActivityDetector` interface
- It's a FALLBACK when tool-specific detectors (like ClaudeCodeDetector) don't match
- Confidence is always `Uncertain` because it's heuristic-based (no log parsing)

### Previous Story Learnings

From Story 15.1:
- Use `domain.ConfidenceUncertain` for low confidence (heuristic-based)
- Use `domain.NewAgentState()` constructor takes (tool, status, duration, confidence)
- Zero-value safety: `AgentUnknown` is the default status

From Story 15.4:
- Use `const detectorName = "Generic"` to avoid repeated string literals
- Clamp negative durations to 0 for future timestamps (clock skew handling)
- Context cancellation check should happen at entry AND periodically during long operations

From existing WaitingDetector (services/waiting_detector.go):
- Default threshold is 10 minutes (configurable via resolver)
- Threshold of 0 means detection is disabled
- Time injection via `now func() time.Time` field for testing

### Difference from Existing WaitingDetector

| Aspect | WaitingDetector (existing) | GenericDetector (new) |
|--------|---------------------------|----------------------|
| Input | `*domain.Project` (with LastActivityAt field) | `string` projectPath |
| Detection | Uses pre-tracked LastActivityAt | Scans filesystem for most recent mtime |
| Threshold | Resolver cascade (CLI > per-project > global) | Simple option (default 10 min) |
| Output | `bool` (IsWaiting) | `AgentState` (rich status + confidence) |
| Interface | `ports.WaitingDetector` | `ports.AgentActivityDetector` |

**Important:** The GenericDetector scans the filesystem directly. This is slower but works for ANY project, unlike WaitingDetector which requires a tracked `Project` with `LastActivityAt`.

### Story 15.6 TUI Integration Requirements

Story 15.6 (TUI Integration) will use GenericDetector as fallback. Key requirements for compatibility:

1. **Fallback Pattern:** GenericDetector is used when ClaudeCodeDetector returns `AgentUnknown` or errors
2. **Duration Format:** Duration must be in same `time.Duration` format as ClaudeCodeDetector for consistent TUI display
3. **Tool Name Display:** Tool name "Generic" will display in detail panel as detection source
4. **Confidence Indication:** `ConfidenceUncertain` will be styled differently (dim or with caveat) vs ClaudeCodeDetector's `ConfidenceCertain`

```go
// In dashboard (Story 15.6):
state, err := claudeDetector.Detect(ctx, project.Path)
if state.IsUnknown() || err != nil {
    // Claude Code detection failed or not available, use fallback
    state, _ = genericDetector.Detect(ctx, project.Path)
}
// Display state in project row with appropriate styling based on Confidence
```

**Status Bar Integration:**
- WaitingForUser from GenericDetector counts toward "⏸️ N WAITING" in status bar
- Confidence level may influence visual priority (Certain > Uncertain)

### doc.go Update

Update existing doc.go to add GenericDetector documentation:

```go
// Package agentdetectors provides implementations of the AgentActivityDetector
// interface for detecting AI coding agent activity in projects.
//
// This package is part of the adapters layer in the hexagonal architecture.
// It contains infrastructure code that interacts with the filesystem to detect
// various AI agent states (working, waiting, idle).
//
// Implementations:
//   - ClaudeCodePathMatcher (Story 15.2): Matches project paths to Claude Code log directories
//   - ClaudeCodeLogParser (Story 15.3): Parses Claude Code JSONL logs with tail optimization
//   - ClaudeCodeDetector (Story 15.4): Main implementation for Claude Code with high confidence
//   - GenericDetector (Story 15.5): File activity fallback for any project with low confidence
package agentdetectors
```

### Platform Considerations

**Current (MVP - Linux/macOS):**
- Hidden file detection uses dot prefix (`strings.HasPrefix(d.Name(), ".")`)
- This works correctly on Unix-like systems

**Post-MVP (Windows):**
- Windows uses file attributes, not dot prefix, for hidden files
- Architecture.md (lines 93-99) specifies OS abstraction layer from day 1
- When Windows support is added, GenericDetector should use `internal/adapters/filesystem/` abstractions
- Consider adding `IsHidden(path string) bool` to filesystem package

### Filesystem Package Consideration

The `internal/adapters/filesystem/` package provides platform abstractions. Before implementing, verify:
- `paths.go` - Does it have reusable file scanning utilities?
- `platform.go` / `platform_unix.go` - Any hidden file detection helpers?

If no reusable patterns exist, document this as rationale for standalone implementation.

### References

- [Source: docs/epics-phase2.md#Story-3.5] - Story acceptance criteria
- [Source: docs/project-context.md#Phase-2-Additions] - Agent detection architecture
- [Source: docs/architecture.md#lines-93-99] - OS abstraction layer requirement
- [Source: internal/core/ports/agent_activity_detector.go] - Interface definition (Story 15.1)
- [Source: internal/adapters/agentdetectors/claude_code_detector.go] - ClaudeCodeDetector pattern (Story 15.4)
- [Source: internal/core/services/waiting_detector.go] - Existing threshold-based detection pattern
- [Source: internal/core/domain/agent_state.go] - AgentState struct
- [Source: internal/core/domain/confidence.go] - Confidence enum (ConfidenceUncertain)
- [Source: internal/adapters/filesystem/] - Platform abstraction layer (check for reusable patterns)

## Dev Agent Record

### Context Reference

- Phase 2 Epic 15: Sub-1-Minute Agent Detection (THE killer feature)
- FR Coverage: FR-P2-5 (System falls back to file-activity detection for non-Claude-Code projects)
- Prerequisite: Story 15.1 (AgentActivityDetector interface) - DONE

### Agent Model Used

Claude Opus 4.5

### Debug Log References

None

### Completion Notes List

- Implemented GenericDetector with functional options pattern (WithThreshold, WithNow)
- Default threshold is 10 minutes, matching existing WaitingDetector behavior
- Invalid threshold values (0, negative) are ignored, keeping default
- Detect() walks filesystem using filepath.WalkDir to find most recent file modification
- Hidden files and directories (starting with `.`) are skipped
- Context cancellation is checked at entry and every 100 files during walk
- All states return ConfidenceUncertain since this is heuristic-based (no log parsing)
- Comprehensive tests cover all acceptance criteria (20 tests)
- Lint and full test suite (1267 tests) pass

### File List

- internal/adapters/agentdetectors/generic_detector.go (NEW)
- internal/adapters/agentdetectors/generic_detector_test.go (NEW)
- internal/adapters/agentdetectors/doc.go (MODIFIED)

## Change Log

- 2026-01-16: Story implementation complete - GenericDetector implemented with 20 comprehensive tests
- 2026-01-16: Code review complete - Added 2 tests (future timestamp, context cancellation during walk), fixed test naming conflicts, enhanced doc.go description. Total 22 tests, all 1269 project tests pass.
