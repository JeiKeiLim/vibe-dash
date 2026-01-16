# Story 14.3: Implement Most-Recent-Artifact-Wins Selection

Status: done

## Story

As a user,
I want the dashboard to show the methodology I'm currently using (based on recent activity),
So that switching methodologies is seamlessly detected.

## User-Visible Changes

- **Changed:** When a project has both Speckit and BMAD artifacts, the dashboard now shows the methodology with the most recent artifact modification (previously showed first-detected methodology due to first-match-wins bug)

## Acceptance Criteria

1. **AC1: Most recent methodology wins** - Given project has Speckit artifacts from 1 week ago AND BMAD artifacts from 1 hour ago, when methodology detection runs, then BMAD is selected as active methodology

2. **AC2: Works both directions** - Given project has BMAD artifacts from 2 days ago AND Speckit artifacts from 5 minutes ago, when methodology detection runs, then Speckit is selected as active methodology

3. **AC3: Threshold for clear winner** - Given project has methodologies with timestamps > 1 hour apart, when selection logic runs, then it returns single winner (not both)

4. **AC4: Single methodology unchanged** - Given project has only Speckit artifacts (no BMAD), when detection runs, then Speckit is selected (existing behavior preserved)

5. **AC5: No methodology unchanged** - Given project has no methodology artifacts, when detection runs, then result is Method="unknown" (existing behavior preserved)

6. **AC6: Zero timestamp fallback** - Given one methodology has timestamp and another has zero timestamp, when selection runs, then methodology with valid timestamp wins

7. **AC7: Both zero timestamps fallback** - Given both methodologies have zero timestamps, when selection runs, then first-registered detector wins (deterministic fallback)

8. **AC8: Integration with detection service** - The new selection logic is available via `DetectionService.DetectWithCoexistenceSelection()` method

## Tasks / Subtasks

- [x] Task 1: Add `SelectByTimestamp` helper function in domain package (AC: 1, 2, 3, 6, 7)
  - [x] Subtask 1.1: Create `internal/core/domain/selection.go` with `SelectByTimestamp` function
  - [x] Subtask 1.2: Add constant `CoexistenceThreshold = 1 * time.Hour`
  - [x] Subtask 1.3: Implement clear winner logic: return `(result, true)` when difference > 1 hour
  - [x] Subtask 1.4: Implement tie logic: return `(nil, false)` when difference <= 1 hour
  - [x] Subtask 1.5: Handle edge case: empty slice → return `(nil, false)`
  - [x] Subtask 1.6: Handle edge case: single result → return `(result, true)`
  - [x] Subtask 1.7: Handle zero timestamps: valid timestamp beats zero timestamp
  - [x] Subtask 1.8: Handle both zero: first in slice wins (deterministic fallback)

- [x] Task 2: Add unit tests for SelectByTimestamp in `internal/core/domain/selection_test.go`
  - [x] Subtask 2.1: Test clear winner - BMAD 1 hour ago, Speckit 1 week ago → BMAD wins
  - [x] Subtask 2.2: Test clear winner - Speckit 5 min ago, BMAD 2 days ago → Speckit wins
  - [x] Subtask 2.3: Test tie - both within 30 minutes → returns `(nil, false)`
  - [x] Subtask 2.4: Test exact 1 hour boundary → tie (threshold is INCLUSIVE: `<=`)
  - [x] Subtask 2.5: Test single result → returns `(result, true)`
  - [x] Subtask 2.6: Test empty slice → returns `(nil, false)`
  - [x] Subtask 2.7: Test zero timestamp vs valid → valid timestamp wins
  - [x] Subtask 2.8: Test both zero timestamps → first in slice wins
  - [x] Subtask 2.9: Test three+ results → still finds most recent correctly

- [x] Task 3: Add `DetectWithCoexistenceSelection` to DetectionService (AC: 8)
  - [x] Subtask 3.1: Add method to `internal/core/services/detection_service.go`
  - [x] Subtask 3.2: Signature: `func (s *DetectionService) DetectWithCoexistenceSelection(ctx context.Context, path string) (*domain.DetectionResult, []*domain.DetectionResult, error)`
  - [x] Subtask 3.3: Call `s.registry.DetectWithCoexistence(ctx, path)` first
  - [x] Subtask 3.4: Apply `domain.SelectByTimestamp(results)` to determine winner
  - [x] Subtask 3.5: Return `(winner, allResults, nil)` when clear winner
  - [x] Subtask 3.6: Return `(nil, allResults, nil)` when tie (caller handles)
  - [x] Subtask 3.7: Return unknown result when no methodologies detected (pattern: `domain.NewDetectionResult("unknown", domain.StageUnknown, domain.ConfidenceUncertain, "no methodology markers found")`)

- [x] Task 4: Update `ports.Detector` interface (AC: 8)
  - [x] Subtask 4.1: Add `DetectWithCoexistenceSelection(ctx context.Context, path string) (*domain.DetectionResult, []*domain.DetectionResult, error)` to `Detector` interface in `internal/core/ports/detector.go`
  - [x] Subtask 4.2: Update `mockDetector` in `detection_service_test.go` to implement new method (delegate to registry.DetectWithCoexistence + domain.SelectByTimestamp)

- [x] Task 5: Add unit tests for DetectionService.DetectWithCoexistenceSelection
  - [x] Subtask 5.1: Test clear winner scenario (mock returns 2 results with >1hr difference)
  - [x] Subtask 5.2: Test tie scenario (mock returns 2 results with <1hr difference)
  - [x] Subtask 5.3: Test single result scenario
  - [x] Subtask 5.4: Test no results scenario (returns unknown)
  - [x] Subtask 5.5: Test context cancellation
  - [x] Subtask 5.6: Test empty path error

- [x] Task 6: Run `make fmt && make lint && make test` to verify all passes

## Dev Notes

### Selection Logic

**File:** `internal/core/domain/selection.go` (NEW FILE - in domain to respect hexagonal boundaries)

```go
package domain

import "time"

// CoexistenceThreshold defines minimum timestamp difference for clear winner.
// If timestamps are within this threshold, it's considered a tie.
// Threshold is INCLUSIVE: exactly 1 hour difference = tie.
const CoexistenceThreshold = 1 * time.Hour

// SelectByTimestamp chooses the methodology with most recent artifact timestamp.
//
// Return semantics:
//   - (winner, true)  → clear winner exists (difference > CoexistenceThreshold)
//   - (nil, false)    → tie (difference <= threshold), caller handles coexistence
//   - (result, true)  → single result provided
//   - (nil, false)    → empty slice
//   - (first, true)   → all zero timestamps, deterministic fallback to first
func SelectByTimestamp(results []*DetectionResult) (*DetectionResult, bool) {
    if len(results) == 0 {
        return nil, false
    }
    if len(results) == 1 {
        return results[0], true
    }

    // Find most recent and second most recent
    var mostRecent, secondRecent *DetectionResult
    var mostRecentTime, secondRecentTime time.Time

    for _, r := range results {
        ts := r.ArtifactTimestamp
        if ts.After(mostRecentTime) {
            secondRecent = mostRecent
            secondRecentTime = mostRecentTime
            mostRecent = r
            mostRecentTime = ts
        } else if ts.After(secondRecentTime) {
            secondRecent = r
            secondRecentTime = ts
        }
    }

    // Both zero timestamps: first in slice wins (deterministic fallback)
    if mostRecentTime.IsZero() && secondRecentTime.IsZero() {
        return results[0], true
    }

    // Check if clear winner (difference > threshold, NOT >=)
    diff := mostRecentTime.Sub(secondRecentTime)
    if diff > CoexistenceThreshold {
        return mostRecent, true
    }

    // Tie case - no clear winner (difference <= 1 hour)
    return nil, false
}
```

### Detection Service Changes

**File:** `internal/core/services/detection_service.go`

Add new method that combines coexistence detection with selection:

```go
// DetectWithCoexistenceSelection runs all detectors and selects based on timestamps.
//
// Return semantics:
//   - (winner, allResults, nil)   → clear winner (>1 hour timestamp difference)
//   - (nil, allResults, nil)      → tie (<=1 hour), caller handles coexistence UI
//   - (unknownResult, nil, nil)   → no methodologies detected
//   - (nil, nil, err)             → error (empty path, context cancelled, detection failed)
func (s *DetectionService) DetectWithCoexistenceSelection(ctx context.Context, path string) (*domain.DetectionResult, []*domain.DetectionResult, error) {
    if path == "" {
        return nil, nil, fmt.Errorf("%w: empty path", domain.ErrPathNotAccessible)
    }

    select {
    case <-ctx.Done():
        return nil, nil, ctx.Err()
    default:
    }

    results, err := s.registry.DetectWithCoexistence(ctx, path)
    if err != nil {
        return nil, nil, fmt.Errorf("%w: %v", domain.ErrDetectionFailed, err)
    }

    if len(results) == 0 {
        // No methodology detected - return unknown result (consistent with Detect behavior)
        unknown := domain.NewDetectionResult(
            "unknown",
            domain.StageUnknown,
            domain.ConfidenceUncertain,
            "no methodology markers found",
        )
        return &unknown, nil, nil
    }

    // Use selector from domain package (maintains hexagonal boundary)
    winner, hasWinner := domain.SelectByTimestamp(results)
    if hasWinner {
        return winner, results, nil
    }

    // Tie - return all results for caller to handle coexistence display
    return nil, results, nil
}
```

### Architecture Decision: Selector Location

The `DetectionService` is in `internal/core/services` which **MUST NOT** import from `internal/adapters/`.

**Decision:** `SelectByTimestamp` goes in `internal/core/domain/selection.go` because:
- It operates purely on domain types (`[]*DetectionResult`)
- It uses only stdlib (`time` package)
- This maintains hexagonal architecture boundaries

### Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/core/domain/selection.go` | **CREATE** | SelectByTimestamp + CoexistenceThreshold |
| `internal/core/domain/selection_test.go` | **CREATE** | Unit tests for selection logic |
| `internal/core/services/detection_service.go` | **MODIFY** | Add DetectWithCoexistenceSelection method |
| `internal/core/services/detection_service_test.go` | **MODIFY** | Add tests for new service method |
| `internal/core/ports/detector.go` | **MODIFY** | Add new method to Detector interface |

**Architecture compliance:** All changes in `internal/core/` use only stdlib (`time` package) - hexagonal boundary maintained.

### Timestamp Comparison Pattern

```go
// Correct: Using time.After for comparison
if t1.After(t2) { ... }

// Correct: Duration subtraction
diff := t1.Sub(t2)
if diff > threshold { ... }

// Avoid: Direct comparison (can fail with timezone issues)
// if t1 > t2 { ... }  // Won't compile anyway
```

### Edge Cases and Return Values

| Scenario | Winner | hasClear | Reasoning |
|----------|--------|----------|-----------|
| Empty slice | `nil` | `false` | Nothing to select |
| Single result | `result` | `true` | Single = winner by default |
| >1 hour difference | most recent | `true` | Clear winner |
| Exactly 1 hour | `nil` | `false` | Threshold inclusive |
| <1 hour difference | `nil` | `false` | Tie case |
| Both zero timestamps | first in slice | `true` | Deterministic fallback |
| One zero, one valid | valid | `true` | Valid timestamp wins |
| Three+ results | most recent | varies | Same logic applies |

### Testing Pattern

**File:** `internal/core/domain/selection_test.go`

```go
package domain

import (
    "testing"
    "time"
)

// createResultWithTimestamp is a test helper to create DetectionResult with timestamp
func createResultWithTimestamp(method string, timestamp time.Time) *DetectionResult {
    result := NewDetectionResult(method, StagePlan, ConfidenceCertain, "test").WithTimestamp(timestamp)
    return &result
}

func TestSelectByTimestamp(t *testing.T) {
    now := time.Now()

    tests := []struct {
        name           string
        results        []*DetectionResult
        wantWinnerNil  bool
        wantWinnerName string // Method name of expected winner
        wantHasClear   bool
    }{
        {
            name:          "empty slice",
            results:       []*DetectionResult{},
            wantWinnerNil: true,
            wantHasClear:  false,
        },
        {
            name:           "single result",
            results:        []*DetectionResult{createResultWithTimestamp("speckit", now)},
            wantWinnerNil:  false,
            wantWinnerName: "speckit",
            wantHasClear:   true,
        },
        {
            name: "clear winner - BMAD more recent (2h difference)",
            results: []*DetectionResult{
                createResultWithTimestamp("speckit", now.Add(-7*24*time.Hour)), // 1 week ago
                createResultWithTimestamp("bmad", now.Add(-1*time.Hour)),        // 1 hour ago
            },
            wantWinnerNil:  false,
            wantWinnerName: "bmad",
            wantHasClear:   true,
        },
        {
            name: "tie - within threshold (30 min difference)",
            results: []*DetectionResult{
                createResultWithTimestamp("speckit", now.Add(-30*time.Minute)),
                createResultWithTimestamp("bmad", now.Add(-45*time.Minute)),
            },
            wantWinnerNil: true,
            wantHasClear:  false,
        },
        {
            name: "exact 1 hour boundary - should be tie (threshold inclusive)",
            results: []*DetectionResult{
                createResultWithTimestamp("speckit", now),
                createResultWithTimestamp("bmad", now.Add(-1*time.Hour)), // exactly 1 hour
            },
            wantWinnerNil: true,
            wantHasClear:  false,
        },
        {
            name: "just over threshold - clear winner (1h1m difference)",
            results: []*DetectionResult{
                createResultWithTimestamp("speckit", now),
                createResultWithTimestamp("bmad", now.Add(-61*time.Minute)),
            },
            wantWinnerNil:  false,
            wantWinnerName: "speckit",
            wantHasClear:   true,
        },
        {
            name: "zero timestamp vs valid - valid wins",
            results: []*DetectionResult{
                createResultWithTimestamp("speckit", time.Time{}),  // zero
                createResultWithTimestamp("bmad", now),              // valid
            },
            wantWinnerNil:  false,
            wantWinnerName: "bmad",
            wantHasClear:   true,
        },
        {
            name: "both zero timestamps - first wins",
            results: []*DetectionResult{
                createResultWithTimestamp("speckit", time.Time{}),
                createResultWithTimestamp("bmad", time.Time{}),
            },
            wantWinnerNil:  false,
            wantWinnerName: "speckit", // first in slice
            wantHasClear:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            winner, hasClear := SelectByTimestamp(tt.results)

            if tt.wantWinnerNil && winner != nil {
                t.Errorf("SelectByTimestamp() winner = %v, want nil", winner)
            }
            if !tt.wantWinnerNil && winner == nil {
                t.Error("SelectByTimestamp() winner = nil, want non-nil")
            }
            if !tt.wantWinnerNil && winner != nil && winner.Method != tt.wantWinnerName {
                t.Errorf("SelectByTimestamp() winner.Method = %q, want %q", winner.Method, tt.wantWinnerName)
            }
            if hasClear != tt.wantHasClear {
                t.Errorf("SelectByTimestamp() hasClear = %v, want %v", hasClear, tt.wantHasClear)
            }
        })
    }
}
```

### Story Dependencies

| Dependency | Status | Relationship |
|------------|--------|--------------|
| Story 14.1 (DetectWithCoexistence) | COMPLETED ✓ | Prerequisite - provides `[]*DetectionResult` |
| Story 14.2 (ArtifactTimestamp) | COMPLETED ✓ | Prerequisite - provides timestamp field |
| Story 14.4 (Coexistence Warning) | Pending | This story enables 14.4 (tie-breaker UI) |
| Story 14.5 (TUI Coexistence Display) | Pending | This story enables 14.5 |

### Critical Rules

1. **Hexagonal boundaries:** `internal/core/` must NOT import from `internal/adapters/` — selector MUST be in domain
2. **Context first:** All service methods accept `context.Context` as first parameter
3. **Error wrapping:** Use `fmt.Errorf("%w: ...", domain.ErrSomething, err)` pattern
4. **Table-driven tests:** Standard Go testing pattern with `tests []struct{...}`
5. **Threshold semantics:** `<=` means tie (exactly 1 hour = tie), `>` means clear winner

### Anti-Patterns to Avoid

| Don't | Do Instead | Why |
|-------|------------|-----|
| Put selector in `internal/adapters/detectors/` | Put in `internal/core/domain/` | Hexagonal boundary violation |
| Use `>=` for threshold comparison | Use `>` only | Exactly 1 hour should be a tie per AC3 |
| Call `detectors.SelectByTimestamp()` from service | Call `domain.SelectByTimestamp()` | Core layer cannot import adapters |
| Return error when no methodologies | Return unknown result | Consistent with existing `Detect()` behavior |
| Guess timestamp for zero values | First-in-slice wins as fallback | Deterministic behavior for testing |

### References

| Document | Section |
|----------|---------|
| PRD | `docs/prd-phase2.md` - FR-P2-9: Select based on most recent artifact |
| Epic | `docs/epics-phase2.md` - Story 2.3 (mapped to 14.3) |
| Previous Story | Story 14.2 - ArtifactTimestamp field and patterns |
| Architecture | `docs/architecture.md` - Hexagonal boundaries |
| Project Context | `docs/project-context.md` - Critical implementation rules |

### Mock Update for detection_service_test.go

The `mockRegistry` in `detection_service_test.go` already has `DetectWithCoexistence`. Add field to control its return:

```go
type mockRegistry struct {
    detectAllResult            *domain.DetectionResult
    detectAllError             error
    detectWithCoexistenceResults []*domain.DetectionResult // ADD THIS
    detectors                  []ports.MethodDetector
}

func (m *mockRegistry) DetectWithCoexistence(ctx context.Context, path string) ([]*domain.DetectionResult, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    // Return configured results for testing
    if m.detectWithCoexistenceResults != nil {
        return m.detectWithCoexistenceResults, nil
    }
    // Fallback: delegate to detectors (existing behavior)
    var results []*domain.DetectionResult
    for _, d := range m.detectors {
        if d.CanDetect(ctx, path) {
            result, err := d.Detect(ctx, path)
            if err == nil && result != nil {
                results = append(results, result)
            }
        }
    }
    return results, nil
}
```

### Service Test Pattern

```go
func TestDetectionService_DetectWithCoexistenceSelection(t *testing.T) {
    now := time.Now()

    tests := []struct {
        name           string
        path           string
        mockResults    []*domain.DetectionResult
        wantWinnerNil  bool
        wantWinnerName string
        wantAllCount   int
        wantErr        bool
    }{
        {
            name:         "empty path returns error",
            path:         "",
            mockResults:  nil,
            wantErr:      true,
        },
        {
            name:           "clear winner",
            path:           "/test",
            mockResults: []*domain.DetectionResult{
                createResultWithTimestamp("speckit", now.Add(-7*24*time.Hour)),
                createResultWithTimestamp("bmad", now),
            },
            wantWinnerNil:  false,
            wantWinnerName: "bmad",
            wantAllCount:   2,
        },
        {
            name:        "tie returns nil winner",
            path:        "/test",
            mockResults: []*domain.DetectionResult{
                createResultWithTimestamp("speckit", now),
                createResultWithTimestamp("bmad", now.Add(-30*time.Minute)),
            },
            wantWinnerNil: true,
            wantAllCount:  2,
        },
        {
            name:           "no results returns unknown",
            path:           "/test",
            mockResults:    []*domain.DetectionResult{},
            wantWinnerNil:  false,
            wantWinnerName: "unknown",
            wantAllCount:   0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            registry := &mockRegistry{
                detectWithCoexistenceResults: tt.mockResults,
            }
            svc := NewDetectionService(registry)

            winner, all, err := svc.DetectWithCoexistenceSelection(context.Background(), tt.path)

            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if tt.wantErr {
                return
            }

            // Verify winner
            if tt.wantWinnerNil && winner != nil {
                t.Errorf("winner = %v, want nil", winner)
            }
            if !tt.wantWinnerNil && winner == nil {
                t.Error("winner = nil, want non-nil")
            }
            if !tt.wantWinnerNil && winner != nil && winner.Method != tt.wantWinnerName {
                t.Errorf("winner.Method = %q, want %q", winner.Method, tt.wantWinnerName)
            }

            // Verify all results count
            if len(all) != tt.wantAllCount {
                t.Errorf("len(all) = %d, want %d", len(all), tt.wantAllCount)
            }
        })
    }
}
```

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

### Completion Notes List

- Implemented `SelectByTimestamp` function in domain package respecting hexagonal boundaries
- Added `CoexistenceThreshold = 1 * time.Hour` constant
- Selection logic correctly handles: empty slice, single result, clear winner (>1hr diff), tie (<=1hr diff), zero timestamps
- Added `DetectWithCoexistenceSelection` method to `DetectionService`
- Updated `ports.Detector` interface with new method
- Updated all mock implementations (shared testhelpers, CLI refresh, TUI refresh) to implement new interface
- All 1328 tests pass, lint clean

### Code Review Fixes Applied

- **M1 Fixed**: Added `TestDetectionService_DetectWithCoexistenceSelection_RegistryError` test to verify error propagation from registry is properly wrapped with `ErrDetectionFailed`

### File List

| File | Action |
|------|--------|
| `internal/core/domain/selection.go` | **CREATED** |
| `internal/core/domain/selection_test.go` | **CREATED** |
| `internal/core/services/detection_service.go` | MODIFIED |
| `internal/core/services/detection_service_test.go` | MODIFIED |
| `internal/core/ports/detector.go` | MODIFIED |
| `internal/shared/testhelpers/mock_detector.go` | MODIFIED |
| `internal/adapters/cli/refresh_test.go` | MODIFIED |
| `internal/adapters/tui/model_refresh_test.go` | MODIFIED |
