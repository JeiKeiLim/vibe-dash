# Story 14.2: Add Artifact Timestamp to Detection Results

Status: done

## Story

As a developer,
I want each detector to report the most recent artifact modification time,
So that the registry can compare timestamps across methodologies.

## User-Visible Changes

None - this is an internal infrastructure change that enables methodology timestamp comparison. User-facing changes will be visible in Stories 14.3-14.5 when the selection logic uses these timestamps.

## Acceptance Criteria

1. **AC1: DetectionResult has ArtifactTimestamp field** - Given the domain model, when a DetectionResult is created, then it has an `ArtifactTimestamp time.Time` field

2. **AC2: Speckit reports most recent mtime** - Given Speckit detector finds `specs/001-feature/plan.md` (modified 2h ago) and `specs/001-feature/spec.md` (modified 1d ago), when detection runs, then DetectionResult.ArtifactTimestamp is 2h ago (most recent)

3. **AC3: BMAD reports most recent mtime** - Given BMAD detector finds `sprint-status.yaml` (modified 30m ago), when detection runs, then DetectionResult.ArtifactTimestamp is 30m ago

4. **AC4: Zero time for no-timestamp cases** - Given a detector cannot determine artifact timestamps (e.g., only marker directory exists with no timestampable files), when detection runs, then DetectionResult.ArtifactTimestamp is zero time (`time.Time{}`)

5. **AC5: Constructor backward compatible** - The existing `NewDetectionResult()` constructor signature is unchanged; timestamp is set via `WithTimestamp()` fluent method

6. **AC6: Existing tests pass** - All existing DetectionResult tests and detector tests continue to pass without modification (backward compatible change)

## Tasks / Subtasks

- [x] Task 1: Update domain model (AC: 1, 5)
  - [x] Subtask 1.1: Add `import "time"` to detection_result.go (required for time.Time)
  - [x] Subtask 1.2: Add `ArtifactTimestamp time.Time` field to `DetectionResult` struct
  - [x] Subtask 1.3: Add `WithTimestamp(t time.Time) DetectionResult` fluent method (returns copy with timestamp set)
  - [x] Subtask 1.4: Add `HasTimestamp() bool` helper method (returns `!dr.ArtifactTimestamp.IsZero()`)
  - [x] Subtask 1.5: Do NOT modify `NewDetectionResult()` signature - backward compatibility via fluent method

- [x] Task 2: Update Speckit detector to track timestamps (AC: 2, 4)
  - [x] Subtask 2.1: Modify `findMostRecentDir()` to return `(string, string, time.Time)` - add dirMtime as third return value. Convert int64 to time.Time: `time.Unix(dirMods[0].modTime, 0)`
  - [x] Subtask 2.2: Modify `analyzeSpecDir()` signature to accept dirMtime and return combined timestamp
  - [x] Subtask 2.3: In `analyzeSpecDir()`, track file mtimes using `os.Stat()` for each artifact file (implement.md, tasks.md, plan.md, spec.md)
  - [x] Subtask 2.4: Use max(dirMtime, fileMtimes) for final timestamp - compare with `time.After()`
  - [x] Subtask 2.5: In `Detect()`, call `result.WithTimestamp(timestamp)` before returning

- [x] Task 3: Update BMAD detector to track timestamps (AC: 3, 4)
  - [x] Subtask 3.1: Modify `findSprintStatusPath()` in stage_parser.go to return `(string, time.Time)` - capture mtime when file found
  - [x] Subtask 3.2: Modify `detectStage()` signature to return timestamp as fourth value: `(Stage, Confidence, string, time.Time)`
  - [x] Subtask 3.3: Track config.yaml mtime in `findBMADConfigWithMtime()` as new function (created)
  - [x] Subtask 3.4: Use max of sprint-status mtime and config mtime for final timestamp
  - [x] Subtask 3.5: In `Detect()`, call `result.WithTimestamp(timestamp)` before returning

- [x] Task 4: Add tests to existing detection_result_test.go (file exists - ADD tests, don't create new file)
  - [x] Subtask 4.1: Test `WithTimestamp()` returns new DetectionResult with timestamp set (table-driven)
  - [x] Subtask 4.2: Test `HasTimestamp()` returns false for zero time, true for non-zero (table-driven)
  - [x] Subtask 4.3: Test zero value DetectionResult has zero ArtifactTimestamp
  - [x] Subtask 4.4: Test original result unchanged after WithTimestamp (immutability check)

- [x] Task 5: Write unit tests for Speckit timestamp tracking
  - [x] Subtask 5.1: Test timestamp reflects most recent spec directory mtime (table-driven with createFileWithMtime helper)
  - [x] Subtask 5.2: Test timestamp reflects most recent artifact file mtime
  - [x] Subtask 5.3: Test zero time when only marker directory exists with no files

- [x] Task 6: Write unit tests for BMAD timestamp tracking
  - [x] Subtask 6.1: Test timestamp reflects sprint-status.yaml mtime
  - [x] Subtask 6.2: Test timestamp reflects config.yaml mtime when no sprint-status
  - [x] Subtask 6.3: Test HasTimestamp returns true for detected BMAD project

- [x] Task 7: Verify backward compatibility (AC: 6)
  - [x] Subtask 7.1: Run `make test` to ensure all existing tests pass
  - [x] Subtask 7.2: Verify existing `NewDetectionResult()` calls compile without changes

- [x] Task 8: Run `make fmt && make lint && make test` to verify

## Dev Notes

### Domain Model Changes

**File:** `internal/core/domain/detection_result.go`

**Required changes:**

1. Add `import "time"` at the top (currently only imports "fmt")
2. Add field to struct:
```go
ArtifactTimestamp time.Time // Most recent artifact modification time (zero if unknown)
```
3. Add methods (do NOT modify NewDetectionResult):
```go
func (dr DetectionResult) WithTimestamp(t time.Time) DetectionResult {
    dr.ArtifactTimestamp = t
    return dr
}

func (dr DetectionResult) HasTimestamp() bool {
    return !dr.ArtifactTimestamp.IsZero()
}
```

### Speckit Detector Changes

**File:** `internal/adapters/detectors/speckit/detector.go`

**Current state analysis:**
- `findMostRecentDir()` stores mtime as `int64` in `dirMods[0].modTime`
- Returns `(string, string)` - needs third return value

**Required changes:**

1. **Modify `findMostRecentDir` signature and return:**
```go
// Change from: func (d *SpeckitDetector) findMostRecentDir(baseDir string, dirs []os.DirEntry) (string, string)
// Change to:   func (d *SpeckitDetector) findMostRecentDir(baseDir string, dirs []os.DirEntry) (string, string, time.Time)

// At the end, convert int64 to time.Time:
dirMtime := time.Unix(dirMods[0].modTime, 0)
return dirMods[0].name, reasoning, dirMtime
```

2. **Modify `analyzeSpecDir` signature to accept and track timestamps:**
```go
// Change from: func (d *SpeckitDetector) analyzeSpecDir(dirPath string, extraReasoning string) (*domain.DetectionResult, error)
// Change to:   func (d *SpeckitDetector) analyzeSpecDir(dirPath string, extraReasoning string, dirMtime time.Time) (*domain.DetectionResult, error)

// Track artifact file mtimes at start of function:
maxMtime := dirMtime
for _, file := range []string{"implement.md", "tasks.md", "plan.md", "spec.md"} {
    if info, err := os.Stat(filepath.Join(dirPath, file)); err == nil {
        if info.ModTime().After(maxMtime) {
            maxMtime = info.ModTime()
        }
    }
}

// Change result creation at the end:
result := domain.NewDetectionResult(d.Name(), stage, confidence, reasoning).WithTimestamp(maxMtime)
return &result, nil
```

3. **Update call site in `Detect()`:**
```go
// Change from: targetDir, reasoning := d.findMostRecentDir(specsDir, specDirs)
// Change to:   targetDir, reasoning, dirMtime := d.findMostRecentDir(specsDir, specDirs)

// Change from: return d.analyzeSpecDir(filepath.Join(specsDir, targetDir), reasoning)
// Change to:   return d.analyzeSpecDir(filepath.Join(specsDir, targetDir), reasoning, dirMtime)
```

### BMAD Detector Changes

**Files:** `internal/adapters/detectors/bmad/detector.go` AND `internal/adapters/detectors/bmad/stage_parser.go`

**Current state analysis:**
- `findSprintStatusPath()` is in stage_parser.go, returns only `string`
- `detectStage()` is in detector.go, returns `(Stage, Confidence, string)`
- Config mtime is not currently tracked

**Required changes in stage_parser.go:**

1. **Modify `findSprintStatusPath` signature:**
```go
// Change from: func findSprintStatusPath(projectPath string, cfg *BMADConfig) string
// Change to:   func findSprintStatusPath(projectPath string, cfg *BMADConfig) (string, time.Time)

// When file found, capture mtime:
if info, err := os.Stat(statusPath); err == nil {
    return statusPath, info.ModTime()
}

// Return zero time when not found:
return "", time.Time{}
```

**Required changes in detector.go:**

2. **Modify `detectStage` to return timestamp:**
```go
// Change from: func (d *BMADDetector) detectStage(ctx context.Context, path string, bmadDir string) (domain.Stage, domain.Confidence, string)
// Change to:   func (d *BMADDetector) detectStage(ctx context.Context, path string, bmadDir string) (domain.Stage, domain.Confidence, string, time.Time)

// Update call to findSprintStatusPath:
statusPath, statusMtime := findSprintStatusPath(path, cfg)

// Track config mtime as fallback (in findBMADConfig or Detect):
var configMtime time.Time
for _, cfgRelPath := range configPaths {
    cfgPath := filepath.Join(bmadDir, cfgRelPath)
    if info, err := os.Stat(cfgPath); err == nil {
        if info.ModTime().After(configMtime) {
            configMtime = info.ModTime()
        }
    }
}

// Use max of timestamps:
artifactMtime := statusMtime
if configMtime.After(artifactMtime) {
    artifactMtime = configMtime
}
return stage, confidence, reasoning, artifactMtime
```

3. **Update `Detect()` to use timestamp:**
```go
// Change from: stage, stageConfidence, stageReasoning := d.detectStage(ctx, path, bmadDir)
// Change to:   stage, stageConfidence, stageReasoning, artifactMtime := d.detectStage(ctx, path, bmadDir)

// At result creation:
result := domain.NewDetectionResult(d.Name(), stage, finalConfidence, fullReasoning).WithTimestamp(artifactMtime)
return &result, nil
```

### Timestamp Sources

| Methodology | Primary Timestamp Source | Fallback Sources |
|-------------|--------------------------|------------------|
| Speckit | mtime of most recent spec folder files | spec directory itself |
| BMAD | `sprint-status.yaml` mtime | `config.yaml` mtime |

### Testing Strategy

**Test helper for controlled file mtimes:**
```go
func createFileWithMtime(t *testing.T, path string, content string, mtime time.Time) {
    t.Helper()
    require.NoError(t, os.WriteFile(path, []byte(content), 0644))
    require.NoError(t, os.Chtimes(path, mtime, mtime))
}
```

**Table-driven test pattern (per project convention):**
```go
func TestSpeckitDetector_Timestamp(t *testing.T) {
    tests := []struct {
        name         string
        setupFiles   func(t *testing.T, dir string)
        wantTimestamp time.Time
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { /* ... */ })
    }
}
```

### Files to Modify

| File | Changes |
|------|---------|
| `internal/core/domain/detection_result.go` | Add field + 2 methods |
| `internal/core/domain/detection_result_test.go` | ADD tests (file exists) |
| `internal/adapters/detectors/speckit/detector.go` | Modify 3 functions |
| `internal/adapters/detectors/speckit/detector_test.go` | Add timestamp tests |
| `internal/adapters/detectors/bmad/detector.go` | Modify detectStage + Detect |
| `internal/adapters/detectors/bmad/stage_parser.go` | Modify findSprintStatusPath |
| `internal/adapters/detectors/bmad/detector_test.go` | Add timestamp tests |

**Architecture compliance:** Domain model change uses only stdlib (`time` package) - no external dependencies in core layer.

### Story Dependency

- **Prerequisite:** Story 14.1 (DetectWithCoexistence) - COMPLETED ✓
- **Enables:** Story 14.3 (Most-Recent-Artifact-Wins Selection)
- **Enables:** Story 14.4 (Coexistence Warning)

### Critical Rules

1. **Domain layer:** Only stdlib imports allowed in `internal/core/` - `time` is stdlib ✓
2. **Backward compatibility:** Do NOT change `NewDetectionResult()` signature
3. **Error handling:** Use zero time (`time.Time{}`) for timestamp errors - never fail detection
4. **Testing convention:** Table-driven tests, co-located with source files

### References

| Document | Section |
|----------|---------|
| PRD | `docs/prd-phase2.md` - FR-P2-8 |
| Epic | `docs/epics-phase2.md` - Story 2.2 |
| Previous Story | Story 14.1 (completed) - patterns to follow |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- Added `ArtifactTimestamp time.Time` field to `DetectionResult` struct
- Added `WithTimestamp(t time.Time) DetectionResult` fluent method for setting timestamps
- Added `HasTimestamp() bool` helper method
- Updated Speckit detector to track max(dirMtime, fileMtimes) from spec directories and artifact files
- Updated BMAD detector to track max(sprintStatusMtime, configMtime)
- Created `findBMADConfigWithMtime()` function to replace `findBMADConfig()` for mtime tracking
- Added comprehensive timestamp tests for domain model, Speckit detector, and BMAD detector
- All existing tests pass (1318 total tests) - backward compatibility verified
- `make fmt && make lint && make test` all pass

**Code Review Fixes Applied:**
- M1: Added TestDetectionResult_WithTimestamp_Chaining test for fluent method chaining behavior
- M2: Consolidated Speckit artifact file checking into single-pass loop, extracted `speckitArtifacts` constant
- M3: Enhanced findBMADConfigWithMtime godoc to document merge priority and AC3 timestamp tracking

### File List

| File | Changes |
|------|---------|
| `internal/core/domain/detection_result.go` | Added ArtifactTimestamp field + WithTimestamp() + HasTimestamp() methods |
| `internal/core/domain/detection_result_test.go` | Added timestamp tests: TestDetectionResult_WithTimestamp, TestDetectionResult_WithTimestamp_Immutability, TestDetectionResult_HasTimestamp, TestDetectionResult_HasTimestamp_ZeroValue, TestDetectionResult_WithTimestamp_Chaining |
| `internal/adapters/detectors/speckit/detector.go` | Modified findMostRecentDir(), analyzeSpecDir() with consolidated artifact file checking, added speckitArtifacts constant |
| `internal/adapters/detectors/speckit/detector_test.go` | Added TestSpeckitDetector_Timestamp, TestSpeckitDetector_Timestamp_HasTimestamp |
| `internal/adapters/detectors/bmad/detector.go` | Modified Detect(), added timestamp tracking for all code paths |
| `internal/adapters/detectors/bmad/stage_parser.go` | Modified findSprintStatusPath(), detectStage(), added findBMADConfigWithMtime() with enhanced documentation |
| `internal/adapters/detectors/bmad/stage_parser_test.go` | Updated findSprintStatusPath tests for (string, time.Time) return value |
| `internal/adapters/detectors/bmad/detector_test.go` | Added TestBMADDetector_Timestamp, TestBMADDetector_Timestamp_HasTimestamp, TestBMADDetector_Timestamp_MaxOfConfigAndStatus |
