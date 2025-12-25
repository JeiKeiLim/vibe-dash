# Story 7.5: Verbose and Debug Logging

Status: review

## Story

As a **user**,
I want **logging options for troubleshooting**,
So that **I can diagnose issues when the tool behaves unexpectedly**.

## Acceptance Criteria

1. **AC1: Default Logging - Errors Only**
   - Given default logging (no flags)
   - When vibe runs normally
   - Then only error-level logs go to stderr
   - And TUI is clean (no log noise in terminal)
   - And stdout contains only user-facing output

2. **AC2: Verbose Mode (--verbose / -v)**
   - Given I run `vibe --verbose` or `vibe -v`
   - When operations occur
   - Then info-level logs are visible on stderr
   - Example output: `time=... level=INFO msg="config loaded" path=/path/to/config`
   - And file/line info is NOT included (cleaner output)
   - And logs go to stderr (not stdout)

3. **AC3: Debug Mode (--debug)**
   - Given I run `vibe --debug`
   - When operations occur
   - Then debug-level logs are visible on stderr
   - Example output: `time=... level=DEBUG source=detector.go:45 msg="checking specs directory" path=/project/specs`
   - And file/line info IS included (source=file.go:line)
   - And verbose-level (INFO) logs are also shown
   - And logs go to stderr

4. **AC4: Debug Precedence Over Verbose**
   - Given I run `vibe --verbose --debug`
   - When logging is initialized
   - Then debug mode takes precedence
   - And debug-level logging is active (not just info)

5. **AC5: JSON Output Not Polluted**
   - Given I run `vibe list --json` (default logging)
   - Then stdout contains clean JSON only
   - And stderr contains only errors (if any)
   - Given I run `vibe list --json --debug 2>/dev/null`
   - Then stdout contains clean JSON only
   - And debug logs can be redirected separately with `2>debug.log`

6. **AC6: Quiet Mode Compatibility**
   - Given I run `vibe add . --quiet --debug`
   - When operation completes
   - Then success messages are suppressed (per --quiet)
   - But debug logs still go to stderr
   - And errors still go to stderr

7. **AC7: Consistent Log Format**
   - Given any logging mode is active
   - When logs are emitted
   - Then format follows slog TextHandler pattern:
     - Default: `time=... level=ERROR msg="error message" key=value`
     - Verbose: `time=... level=INFO msg="info message" key=value`
     - Debug: `time=... level=DEBUG source=file.go:123 msg="debug message" key=value`

## Out of Scope (Deferred)

- **Log file output**: Persisting logs to a file (e.g., `~/.vibe-dash/logs/`) - Post-MVP
- **Log rotation**: Automatic cleanup of old logs - Post-MVP
- **JSON log format**: Structured JSON logs for machine parsing - Post-MVP
- **TUI log panel**: Displaying logs within TUI interface - Post-MVP

## Epic 7 Context

Story 7.5 is part of Epic 7 (Error Handling & Polish) which focuses on graceful error handling, helpful feedback, and final polish. Previous stories established patterns:
- **Story 7.1**: `slog.Error()` for watcher failures, structured key-value logging
- **Story 7.2**: `slog.Warn()` for config issues, `slog.Info()` for successful operations
- **Story 7.3**: `slog.Error()` with recovery context, `slog.Info()` for recovery success
- **Story 7.4**: No logging changes (UI-focused story)

This story ensures the existing logging infrastructure is properly exposed to users via CLI flags.

## Critical Implementation Notes

### Hexagonal Architecture Boundary Warning

**CRITICAL**: `detection_service.go` is in `internal/core/services/` - the core layer. Per project-context.md "Log once at handling site only", core services should have MINIMAL logging. Detection results are already logged at the adapter level (TUI/CLI).

**DO NOT add extensive logging to core services**. Instead:
- Add debug logs ONLY in adapter-layer detectors (`speckit/detector.go`, `bmad/detector.go`)
- Core service logging should be limited to error conditions

### Existing Test Infrastructure

`flags_test.go` already contains logging tests at lines 179, 191, 203:
- `TestInitLogging_DefaultLevel` pattern exists
- Verify existing coverage before adding new tests

### Log Key Naming Convention

Per `project-context.md`, use `snake_case` for slog keys:
```go
// Correct
slog.Debug("checking path", "check_path", path, "project_name", name)

// Incorrect
slog.Debug("checking path", "checkPath", path, "projectName", name)
```

## Tasks / Subtasks

- [x] Task 1: Verify existing flag infrastructure (AC: 2, 3, 4)
  - [x] 1.1: Confirm `--verbose` and `--debug` flags exist in `flags.go:19-20`
  - [x] 1.2: Confirm `initLogging()` function configures slog correctly (lines 34-53)
  - [x] 1.3: Verify debug takes precedence (switch statement at lines 38-48)
  - [x] 1.4: Document any gaps between current implementation and ACs

- [x] Task 2: Add strategic Info-level logs in ADAPTER layer only (AC: 2)
  - [x] 2.1: Add `slog.Info()` in `add.go` after successful project save
  - [x] 2.2: Verify `loader.go` already has config loading logs (lines 53,68,82)
  - [x] 2.3: Verify `coordinator.go` already has recovery success logs (lines 93,129,588)
  - [x] 2.4: Skip detection_service.go (core layer - violates architecture)

- [x] Task 3: Add strategic Debug-level logs in detectors (AC: 3)
  - [x] 3.1: Add `slog.Debug()` in `speckit/detector.go` for marker and file checks
  - [x] 3.2: Add `slog.Debug()` in `bmad/detector.go` for marker and config checks
  - [x] 3.3: Verify `activity_tracker.go` already has debug logs (line 65)
  - [x] 3.4: SKIPPED - `waiting_detector.go` is in core layer (architecture violation)

- [x] Task 4: Verify stderr routing (AC: 5, 6)
  - [x] 4.1: Confirm `slog.NewTextHandler(os.Stderr, opts)` in initLogging()
  - [x] 4.2: Test `vibe list --json` produces clean stdout
  - [x] 4.3: Test `vibe list --json --debug 2>/dev/null` produces clean JSON

- [x] Task 5: Verify and extend tests (AC: all)
  - [x] 5.1: Check existing tests in `flags_test.go:179,191,203` for coverage
  - [x] 5.2: Existing test `TestDebugPrecedenceOverVerbose` already covers this
  - [x] 5.3: Add test `TestLoggingGoesToStderr` for stderr routing
  - [x] 5.4: Add test `TestQuietAndDebugCombination` for quiet + debug

- [x] Task 6: Manual verification (AC: all)
  - [x] 6.1: Test all flag combinations per Manual Testing Guide
  - [x] 6.2: Verify log output format matches AC7
  - [x] 6.3: Help text already clear - no changes needed

## Dev Notes

### Existing Infrastructure (VERIFY FIRST)

The logging infrastructure is ALREADY COMPLETE in `flags.go`:

| Location | Code | Status |
|----------|------|--------|
| `flags.go:19` | `--verbose` flag with `-v` short | DONE |
| `flags.go:20` | `--debug` flag | DONE |
| `flags.go:34-53` | `initLogging()` configures slog | DONE |
| `flags.go:39-41` | Debug level with AddSource=true | DONE |
| `flags.go:42-44` | Verbose level without source | DONE |
| `flags.go:45-47` | Default error level | DONE |
| `flags.go:51` | TextHandler to stderr | DONE |

### Existing slog Usage Summary

**Already Logging (DO NOT DUPLICATE):**
- `watcher.go`: Error/Warn for watch failures
- `coordinator.go`: Error/Warn/Info for DB operations and recovery
- `loader.go`: Warn for config issues
- `activity_tracker.go:65`: Debug for path matching
- `model.go`: Debug for refresh, watcher events

**Needs Debug Logs Added:**
- `speckit/detector.go`: No current logging
- `bmad/detector.go`: No current logging
- `waiting_detector.go`: Only has Warn for nil project

### Implementation Examples

**Add to `speckit/detector.go` (CanDetect and Detect methods):**
```go
// In CanDetect:
slog.Debug("checking speckit markers", "path", path)
for _, marker := range markerDirs {
    markerPath := filepath.Join(path, marker)
    slog.Debug("checking marker directory", "marker", marker, "full_path", markerPath)
    // ... existing check
}

// In Detect:
slog.Debug("analyzing spec directory", "specs_dir", specsDir)
slog.Debug("found spec subdirectories", "count", len(specDirs))
```

**Add to `bmad/detector.go` (CanDetect and Detect methods):**
```go
// In CanDetect:
slog.Debug("checking bmad markers", "path", path)

// In Detect:
slog.Debug("reading bmad config", "config_path", cfgPath)
slog.Debug("extracted version", "version", version)
```

**Add to `waiting_detector.go` (IsWaiting method):**
```go
// After threshold calculation:
slog.Debug("checking waiting state",
    "project", project.Name,
    "elapsed_minutes", inactiveDuration.Minutes(),
    "threshold_minutes", thresholdMinutes,
    "is_waiting", inactiveDuration >= threshold)
```

**Add to `add.go` (after successful save):**
```go
// After repository.Save() succeeds:
slog.Info("project added",
    "name", project.Name,
    "path", canonicalPath,
    "method", project.DetectedMethod)
```

### Files to Modify

| File | Action | Why |
|------|--------|-----|
| `speckit/detector.go` | ADD Debug logs | No current logging for file checks |
| `bmad/detector.go` | ADD Debug logs | No current logging for detection |
| `waiting_detector.go` | ADD Debug log | Only has Warn, needs Debug for thresholds |
| `add.go` | ADD Info log | Log successful project additions |
| `flags_test.go` | VERIFY/EXTEND | Check existing tests, add missing |

**DO NOT MODIFY** (already has appropriate logging):
- `detection_service.go` - Core layer, minimal logging by design
- `loader.go` - Already has config loading warnings
- `coordinator.go` - Already has comprehensive logging
- `watcher.go` - Already has error/warn logging

## Manual Testing Guide

**Time needed:** 5 minutes

### Quick Verification Commands

```bash
# Build first
make build

# AC1: Default - should show clean output
./bin/vibe list

# AC2: Verbose - should show level=INFO messages on stderr
./bin/vibe list --verbose 2>&1 | grep "level=INFO"

# AC3: Debug - should show level=DEBUG with source= on stderr
./bin/vibe list --debug 2>&1 | grep "level=DEBUG.*source="

# AC4: Precedence - debug source info should appear
./bin/vibe list -v --debug 2>&1 | grep "source="

# AC5: JSON not polluted
./bin/vibe list --json | python3 -m json.tool

# AC6: Quiet + debug
./bin/vibe add . --quiet --debug 2>&1
# Should see debug logs but NO "Added:" success message
```

### Expected Output Formats

| Mode | Sample Output |
|------|---------------|
| Default | (no log output unless errors) |
| Verbose | `time=... level=INFO msg="config loaded" path=...` |
| Debug | `time=... level=DEBUG source=detector.go:45 msg="checking markers"...` |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All quick verification commands work | Mark `review` |
| level=INFO appears with --verbose | Pass AC2 |
| level=DEBUG + source= appears with --debug | Pass AC3 |
| JSON parses cleanly with --json | Pass AC5 |
| Any check fails | Document issue, do NOT mark review |

## Dependencies

- Story 7.1-7.4 completed (established logging patterns)
- `flags.go` infrastructure exists and is verified working

## References

- [Source: internal/adapters/cli/flags.go:34-53 - initLogging()]
- [Source: docs/architecture.md#Logging & Observability]
- [Source: docs/project-context.md#Go Patterns - Log once]

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

Task 3.4 deviation: `waiting_detector.go` is in `internal/core/services/` (core layer). Per project-context.md "Log once at handling site only", core services should have MINIMAL logging. The story's own Critical Implementation Notes explicitly warn: "DO NOT add extensive logging to core services." This applies equally to waiting_detector.go, so Task 3.4 was intentionally skipped to preserve architecture boundaries.

### Completion Notes List

- All 7 ACs verified working via manual testing
- Existing flag infrastructure was complete (`flags.go:19-53`)
- Added Info-level log in `add.go:213` for project additions (includes method + stage)
- Added Debug-level logs in `speckit/detector.go` and `bmad/detector.go`
- Added 2 new tests: `TestQuietAndDebugCombination` and `TestLoggingGoesToStderr`
- All tests pass (100%), linting passes
- JSON output remains clean when using debug mode (logs go to stderr)
- Debug mode correctly shows source file/line info (e.g., `source=detector.go:45`)

### File List

- `internal/adapters/cli/add.go` - Added slog.Info after successful save (method + stage)
- `internal/adapters/detectors/speckit/detector.go` - Added slog.Debug for marker checks
- `internal/adapters/detectors/bmad/detector.go` - Added slog.Debug for marker and config checks (+ marker found log)
- `internal/adapters/cli/flags_test.go` - Added TestQuietAndDebugCombination, TestLoggingGoesToStderr

## Change Log

- 2025-12-25: Code review fixes - added stage to add.go log, marker found log to bmad detector
- 2025-12-25: Story 7.5 implemented - verbose and debug logging (Amelia/Claude Opus 4.5)

