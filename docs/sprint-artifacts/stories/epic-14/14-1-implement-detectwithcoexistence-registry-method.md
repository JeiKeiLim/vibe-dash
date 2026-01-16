# Story 14.1: Implement DetectWithCoexistence Registry Method

Status: done

## Story

As a developer,
I want the detector registry to run all detectors and collect all matches,
So that methodology coexistence can be properly evaluated.

## User-Visible Changes

None - this is an internal infrastructure change that enables methodology coexistence detection. User-facing changes will be visible in Stories 14.3-14.5.

## Acceptance Criteria

1. **AC1: Multiple matches collected** - Given a project with both Speckit and BMAD artifacts, when `DetectWithCoexistence(ctx, path)` is called, then it returns `[]*DetectionResult` containing both matches

2. **AC2: Single match returns single result** - Given a project with only Speckit artifacts, when `DetectWithCoexistence(ctx, path)` is called, then it returns `[]*DetectionResult` with single Speckit match

3. **AC3: No match returns empty slice** - Given a project with no methodology artifacts, when `DetectWithCoexistence(ctx, path)` is called, then it returns empty slice (no matches, no error)

4. **AC4: Context cancellation honored** - Given a cancelled context, when `DetectWithCoexistence(ctx, path)` is called, then it returns `context.Canceled` error promptly

5. **AC5: Detector errors handled gracefully** - Given a detector that returns an error, when `DetectWithCoexistence(ctx, path)` is called, then it continues to the next detector and logs the error (resilient design)

6. **AC6: Does NOT replace DetectAll** - The existing `DetectAll()` method continues to work unchanged (first-match-wins behavior preserved for backward compatibility)

## Tasks / Subtasks

- [x] Task 1: Add `DetectWithCoexistence` method to Registry (AC: 1, 2, 3, 6)
  - [x] Subtask 1.1: Implement method signature `func (r *Registry) DetectWithCoexistence(ctx context.Context, path string) ([]*domain.DetectionResult, error)`
  - [x] Subtask 1.2: Iterate ALL detectors (not first-match-wins)
  - [x] Subtask 1.3: Collect all successful detection results into slice
  - [x] Subtask 1.4: Return empty slice when no detectors match (not an error)

- [x] Task 2: Add context cancellation support (AC: 4)
  - [x] Subtask 2.1: Check context before iterating detectors
  - [x] Subtask 2.2: Check context between each detector invocation
  - [x] Subtask 2.3: Return `ctx.Err()` when cancelled

- [x] Task 3: Add resilient error handling (AC: 5)
  - [x] Subtask 3.1: Log detector errors with slog.Debug (consistent with existing pattern)
  - [x] Subtask 3.2: Continue iteration on detector error (don't short-circuit)
  - [x] Subtask 3.3: Skip nil results from detectors

- [x] Task 4: Update DetectorRegistry interface in ports (AC: 6)
  - [x] Subtask 4.1: Add `DetectWithCoexistence` to `ports.DetectorRegistry` interface
  - [x] Subtask 4.2: Ensure existing `DetectAll` signature unchanged

- [x] Task 5: Write comprehensive unit tests
  - [x] Subtask 5.1: Test multiple detectors returning matches
  - [x] Subtask 5.2: Test single detector match
  - [x] Subtask 5.3: Test no matches returns empty slice
  - [x] Subtask 5.4: Test context cancellation
  - [x] Subtask 5.5: Test error handling (detector errors don't stop iteration)
  - [x] Subtask 5.6: Test nil result handling
  - [x] Subtask 5.7: Test empty registry returns empty slice (mirrors TestRegistry_DetectAll_EmptyRegistry pattern)

- [x] Task 6: Run `make fmt && make lint && make test` to verify

## Dev Notes

### Registry Pattern Reference

The Registry (`internal/adapters/detectors/registry.go`) is the **only** component that knows about all detector implementations. Services interact with detectors through the registry, never directly.

**Current `DetectAll` behavior (preserve this):**
- First-match-wins: returns immediately when first detector succeeds
- Used by `DetectionService.Detect()` for single-methodology projects
- Cannot be changed (backward compatibility)

**New `DetectWithCoexistence` behavior:**
- Run ALL detectors regardless of matches
- Collect all successful results
- Return slice (empty if no matches)
- Used by future Story 14.3 for methodology comparison

### Implementation Pattern

**Copy the `DetectMultiple` pattern from detection_service.go:78-119, adapting for Registry:**

1. Check context at entry with select/case on `ctx.Done()`
2. Iterate `r.detectors` (not `s.registry.Detectors()`)
3. Check context before each detector
4. Call `CanDetect` then `Detect`
5. On error: `slog.Debug("detector error during coexistence detection", ...)` and `continue`
6. Collect non-nil results into slice
7. Return results slice (empty slice on no matches, NOT an error)

**Key difference from DetectAll:** Do NOT return early on first match. Collect ALL matches.

### Interface Update

Add to `internal/core/ports/detector.go` (DetectorRegistry interface):

```go
// DetectWithCoexistence runs ALL registered detectors and returns all matches.
// Unlike DetectAll which returns first match, this method collects all results
// for methodology coexistence evaluation.
// Returns empty slice if no detectors match (not an error).
DetectWithCoexistence(ctx context.Context, path string) ([]*domain.DetectionResult, error)
```

### Testing Pattern

Use existing mock pattern from `registry_test.go`:

```go
func TestRegistry_DetectWithCoexistence_MultipleMethods(t *testing.T) {
    r := detectors.NewRegistry()
    ctx := context.Background()

    result1 := domain.NewDetectionResult("speckit", domain.StageSpecify, domain.ConfidenceCertain, "spec.md exists")
    result2 := domain.NewDetectionResult("bmad", domain.StageTasks, domain.ConfidenceCertain, "sprint-status.yaml exists")

    detector1 := &mockDetector{name: "speckit", canDetect: true, result: &result1}
    detector2 := &mockDetector{name: "bmad", canDetect: true, result: &result2}

    r.Register(detector1)
    r.Register(detector2)

    results, err := r.DetectWithCoexistence(ctx, "/some/path")
    if err != nil {
        t.Fatalf("DetectWithCoexistence() error = %v", err)
    }

    if len(results) != 2 {
        t.Errorf("DetectWithCoexistence() returned %d results, want 2", len(results))
    }
    // Both detectors should have been called
    if detector1.detectCalls != 1 {
        t.Errorf("First detector called %d times, want 1", detector1.detectCalls)
    }
    if detector2.detectCalls != 1 {
        t.Errorf("Second detector called %d times, want 1", detector2.detectCalls)
    }
}
```

### Project Structure Notes

**Files to modify:**
1. `internal/core/ports/detector.go` - Add method to `DetectorRegistry` interface
2. `internal/adapters/detectors/registry.go` - Implement `DetectWithCoexistence`
3. `internal/adapters/detectors/registry_test.go` - Add tests

**Interface Compilation Verification:**
After adding `DetectWithCoexistence` to the `DetectorRegistry` interface, verify the compile-time check still passes:
```go
var _ ports.DetectorRegistry = (*Registry)(nil)  // registry.go:28
```
This ensures Registry implements all interface methods.

**Alignment with hexagonal architecture:**
- Registry is in adapters layer (implements port interface)
- Interface definition stays in ports layer
- No domain layer changes needed for this story

**Story Dependency:**
This story enables Story 14.2 (Add Artifact Timestamp to Detection Results) which will extend `DetectionResult` with `ArtifactTimestamp time.Time` for methodology comparison.

### Critical Rules from project-context.md

1. **Registry Pattern:** Services call `registry.DetectAll()` or `registry.DetectWithCoexistence()`, never individual detectors directly
2. **Context first:** All methods accept `context.Context` as first parameter
3. **Error wrapping:** Use `fmt.Errorf("...: %w", err)` pattern
4. **Log once:** Log at handling site only, not during propagation

### References

- [PRD: docs/prd-phase2.md#Methodology Detection] - FR-P2-7: Detect multiple methodologies simultaneously
- [Epic: docs/epics-phase2.md#Story 2.1] - Story definition (Epic 14 in sprint = Epic 2 in epics-phase2.md)
- [Code: internal/adapters/detectors/registry.go] - Current Registry implementation
- [Code: internal/core/services/detection_service.go:78-119] - DetectMultiple pattern to follow
- [Arch: docs/architecture.md + docs/project-context.md] - Registry pattern: services call registry, never detectors directly

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

1. Implemented `DetectWithCoexistence` method in `internal/adapters/detectors/registry.go`:
   - Follows the pattern from `detection_service.go:78-119`
   - Runs ALL registered detectors (not first-match-wins like DetectAll)
   - Returns `[]*domain.DetectionResult` slice with all matches
   - Returns empty slice (not nil) when no detectors match
   - Context cancellation checked at entry and before each detector
   - Detector errors logged with `slog.Debug` and iteration continues
   - Nil results skipped

2. Updated `ports.DetectorRegistry` interface with new method signature

3. Fixed `mockRegistry` in `detection_service_test.go` to implement the new interface method

4. Added 7 comprehensive unit tests covering:
   - Multiple detectors returning matches (AC1)
   - Single detector match (AC2)
   - No matches returns empty slice (AC3)
   - Context cancellation (AC4)
   - Error handling - detector errors don't stop iteration (AC5)
   - Nil result handling
   - Empty registry returns empty slice

5. All tests pass: `make fmt && make lint && make test` ✓

6. Code review completed:
   - Fixed: Story File List to include sprint-status.yaml
   - Fixed: Added order preservation verification in `TestRegistry_DetectWithCoexistence_MultipleMethods`
   - Note: Future refactoring opportunity - `DetectionService.DetectMultiple` could delegate to `registry.DetectWithCoexistence()` for DRY (out of scope for this story)

### File List

- `internal/core/ports/detector.go` - Added `DetectWithCoexistence` to interface
- `internal/adapters/detectors/registry.go` - Implemented `DetectWithCoexistence` method
- `internal/adapters/detectors/registry_test.go` - Added 7 unit tests
- `internal/core/services/detection_service_test.go` - Fixed mockRegistry to implement new interface
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status
