# Story 14.4: Implement Coexistence Warning for Tie-Breaker

Status: done

## Story

As a user,
I want to see a warning when the dashboard can't determine which methodology is active,
So that I understand why both methodologies are shown.

## User-Visible Changes

- **New:** When two methodologies have similar timestamps (within 1 hour), both are returned with a coexistence warning flag that downstream consumers (TUI) can use to display an appropriate warning message

## Acceptance Criteria

1. **AC1: Tie detection triggers warning flag** - Given project has Speckit artifacts from 30 minutes ago AND BMAD artifacts from 45 minutes ago, when methodology detection runs (difference < 1 hour), then both methodologies are returned with CoexistenceWarning=true

2. **AC2: Warning includes tie message** - When coexistence is detected, results include CoexistenceMessage="Multiple methodologies detected with similar activity"

3. **AC3: Clear winner has no warning** - Given methodologies with >1 hour timestamp difference, when selection runs, then CoexistenceWarning=false on the winner result

4. **AC4: Single methodology has no warning** - Given only one methodology detected, then CoexistenceWarning=false

5. **AC5: No methodologies has no warning** - Given no methodology detected (unknown result), then CoexistenceWarning=false

6. **AC6: Threshold boundary test** - Given methodologies with EXACTLY 1 hour difference, then CoexistenceWarning=true (threshold is inclusive)

## Tasks / Subtasks

- [x] Task 1: Add CoexistenceWarning fields to DetectionResult (AC: 1, 3, 4, 5)
  - [x] Subtask 1.1: Add `CoexistenceWarning bool` field to `DetectionResult` struct in `internal/core/domain/detection_result.go`
  - [x] Subtask 1.2: Add `CoexistenceMessage string` field for optional warning text
  - [x] Subtask 1.3: Add `WithCoexistenceWarning(msg string) DetectionResult` builder method (returns copy with both fields set)
  - [x] Subtask 1.4: Add `HasCoexistenceWarning() bool` accessor method

- [x] Task 2: Verify existing tie detection behavior (AC: 1, 6)
  - [x] Subtask 2.1: Confirm `SelectByTimestamp` in `selection.go` already returns `(nil, false)` for tie case - read and verify
  - [x] Subtask 2.2: Confirm existing `selection_test.go` tests cover exact 1-hour boundary - the test "exact 1 hour boundary - should be tie" validates this

- [x] Task 3: Update DetectionService.DetectWithCoexistenceSelection (AC: 1, 2, 3)
  - [x] Subtask 3.1: Add exported constant `CoexistenceWarningMessage = "Multiple methodologies detected with similar activity"` at package level
  - [x] Subtask 3.2: When `SelectByTimestamp` returns `(nil, false)` (tie case), create new result pointers with warning set
  - [x] Subtask 3.3: Use pattern: `modifiedResult := results[i].WithCoexistenceWarning(CoexistenceWarningMessage); results[i] = &modifiedResult`
  - [x] Subtask 3.4: When clear winner exists, return as-is (CoexistenceWarning defaults to false in Go)

- [x] Task 4: Add unit tests for domain layer (AC: 1, 2, 3, 4, 5)
  - [x] Subtask 4.1: Create `internal/core/domain/detection_result_test.go` (file does not exist yet)
  - [x] Subtask 4.2: Test `WithCoexistenceWarning` sets both `CoexistenceWarning=true` and `CoexistenceMessage`
  - [x] Subtask 4.3: Test `HasCoexistenceWarning()` returns correct bool
  - [x] Subtask 4.4: Test default struct has `CoexistenceWarning=false` (Go zero value)

- [x] Task 5: Add unit tests for service layer coexistence warning (AC: 1, 2, 3, 4, 5, 6)
  - [x] Subtask 5.1: Test tie case - both results have `CoexistenceWarning=true` and message set
  - [x] Subtask 5.2: Test clear winner - winner has `CoexistenceWarning=false`
  - [x] Subtask 5.3: Test single methodology - `CoexistenceWarning=false`
  - [x] Subtask 5.4: Test no methodology (unknown) - `CoexistenceWarning=false`
  - [x] Subtask 5.5: Test exact 1 hour boundary - `CoexistenceWarning=true` (critical boundary test for AC6)
  - [x] Subtask 5.6: Reuse existing `createTestResultWithTimestamp` helper from `detection_service_test.go`

- [x] Task 6: Run `make fmt && make lint && make test` to verify all passes

## Dev Notes

### Domain Model Changes

**File:** `internal/core/domain/detection_result.go`

Add fields after existing `ArtifactTimestamp`:

```go
type DetectionResult struct {
	Method              string     // "speckit", "bmad", "unknown"
	Stage               Stage      // Detected stage
	Confidence          Confidence // How certain the detection is
	Reasoning           string     // Human-readable explanation (FR11, FR26)
	ArtifactTimestamp   time.Time  // Most recent artifact modification time (zero if unknown)
	CoexistenceWarning  bool       // True when multiple methodologies have similar timestamps
	CoexistenceMessage  string     // Warning message for TUI display
}
```

Add builder and accessor methods:

```go
// WithCoexistenceWarning returns a copy with coexistence warning set.
// Used by DetectionService when tie-breaker can't determine clear winner.
func (dr DetectionResult) WithCoexistenceWarning(msg string) DetectionResult {
	dr.CoexistenceWarning = true
	dr.CoexistenceMessage = msg
	return dr
}

// HasCoexistenceWarning returns true if coexistence warning is set.
// Used by TUI (Story 14.5) to determine if warning should be displayed.
func (dr DetectionResult) HasCoexistenceWarning() bool {
	return dr.CoexistenceWarning
}
```

### Detection Service Changes

**File:** `internal/core/services/detection_service.go`

Add constant at package level (after imports):

```go
// CoexistenceWarningMessage is the standard message shown when methodologies have similar timestamps.
// Exported for use by TUI layer in Story 14.5.
const CoexistenceWarningMessage = "Multiple methodologies detected with similar activity"
```

Update `DetectWithCoexistenceSelection` - add after the current "Tie - return all results" comment:

```go
// Tie case: Set coexistence warning on all results
warningResults := make([]*domain.DetectionResult, len(results))
for i, r := range results {
	modified := r.WithCoexistenceWarning(CoexistenceWarningMessage)
	warningResults[i] = &modified
}

// Tie - return all results with warning for caller to handle coexistence display
return nil, warningResults, nil
```

### AC4/AC5 Implicit Handling

**Why no explicit code for AC4/AC5:**
- AC4 (single methodology): `SelectByTimestamp` returns `(result, true)` for single result, so warning code path is skipped. Go zero value means `CoexistenceWarning=false`.
- AC5 (no methodology): Returns unknown result created fresh with zero values, so `CoexistenceWarning=false`.
- This is correct behavior - no code changes needed for these ACs.

### Architecture Compliance

| Rule | Compliance |
|------|------------|
| Hexagonal boundary | Changes in `internal/core/domain/` and `internal/core/services/` only |
| No adapter imports in core | Using only stdlib (`time`) |
| Domain types | Adding fields to existing `DetectionResult` |
| Table-driven tests | Standard Go testing pattern |

### Test Pattern - Domain Layer

**File:** `internal/core/domain/detection_result_test.go` (CREATE)

```go
package domain

import "testing"

func TestDetectionResult_CoexistenceWarning(t *testing.T) {
	tests := []struct {
		name         string
		applyWarning bool
		wantWarning  bool
		wantMessage  string
	}{
		{
			name:         "default has no warning",
			applyWarning: false,
			wantWarning:  false,
			wantMessage:  "",
		},
		{
			name:         "with warning sets flag and message",
			applyWarning: true,
			wantWarning:  true,
			wantMessage:  "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDetectionResult("speckit", StagePlan, ConfidenceCertain, "test")
			if tt.applyWarning {
				r = r.WithCoexistenceWarning("test message")
			}

			if r.HasCoexistenceWarning() != tt.wantWarning {
				t.Errorf("HasCoexistenceWarning() = %v, want %v", r.HasCoexistenceWarning(), tt.wantWarning)
			}
			if r.CoexistenceMessage != tt.wantMessage {
				t.Errorf("CoexistenceMessage = %q, want %q", r.CoexistenceMessage, tt.wantMessage)
			}
		})
	}
}

func TestDetectionResult_CoexistenceWarning_PreservesOtherFields(t *testing.T) {
	original := NewDetectionResult("bmad", StageTasks, ConfidenceLikely, "found .bmad")
	modified := original.WithCoexistenceWarning("test")

	// Verify original unchanged
	if original.CoexistenceWarning {
		t.Error("original should not be modified")
	}

	// Verify other fields preserved
	if modified.Method != "bmad" {
		t.Errorf("Method = %q, want %q", modified.Method, "bmad")
	}
	if modified.Stage != StageTasks {
		t.Errorf("Stage = %v, want %v", modified.Stage, StageTasks)
	}
}
```

### Test Pattern - Service Layer

**File:** `internal/core/services/detection_service_test.go` (MODIFY - add to existing file)

Add after existing `TestDetectionService_DetectWithCoexistenceSelection_*` tests:

```go
func TestDetectionService_DetectWithCoexistenceSelection_CoexistenceWarning(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		mockResults      []*domain.DetectionResult
		wantWarningOnAll bool
		wantMessage      string
	}{
		{
			name: "tie case - warning set on all results",
			mockResults: []*domain.DetectionResult{
				createTestResultWithTimestamp("speckit", now.Add(-30*time.Minute)),
				createTestResultWithTimestamp("bmad", now.Add(-45*time.Minute)),
			},
			wantWarningOnAll: true,
			wantMessage:      services.CoexistenceWarningMessage,
		},
		{
			name: "exact 1 hour boundary - warning set (AC6 inclusive)",
			mockResults: []*domain.DetectionResult{
				createTestResultWithTimestamp("speckit", now),
				createTestResultWithTimestamp("bmad", now.Add(-1*time.Hour)),
			},
			wantWarningOnAll: true,
			wantMessage:      services.CoexistenceWarningMessage,
		},
		{
			name: "clear winner (>1hr diff) - no warning",
			mockResults: []*domain.DetectionResult{
				createTestResultWithTimestamp("speckit", now.Add(-7*24*time.Hour)),
				createTestResultWithTimestamp("bmad", now),
			},
			wantWarningOnAll: false,
		},
		{
			name: "single result - no warning",
			mockResults: []*domain.DetectionResult{
				createTestResultWithTimestamp("speckit", now),
			},
			wantWarningOnAll: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &mockRegistry{
				detectWithCoexistenceResults:    tt.mockResults,
				detectWithCoexistenceResultsSet: true,
			}
			svc := services.NewDetectionService(registry)

			winner, all, err := svc.DetectWithCoexistenceSelection(context.Background(), "/test")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantWarningOnAll {
				// Tie case: winner should be nil, all results should have warning
				if winner != nil {
					t.Errorf("expected nil winner for tie case, got %v", winner)
				}
				for i, r := range all {
					if !r.HasCoexistenceWarning() {
						t.Errorf("result[%d] should have CoexistenceWarning=true", i)
					}
					if r.CoexistenceMessage != tt.wantMessage {
						t.Errorf("result[%d].CoexistenceMessage = %q, want %q", i, r.CoexistenceMessage, tt.wantMessage)
					}
				}
			} else {
				// Clear winner or single: check no warning on winner
				if winner != nil && winner.HasCoexistenceWarning() {
					t.Error("winner should not have coexistence warning")
				}
			}
		})
	}
}

func TestDetectionService_DetectWithCoexistenceSelection_UnknownHasNoWarning(t *testing.T) {
	// AC5: No methodologies returns unknown with CoexistenceWarning=false
	mock := &mockRegistry{
		detectWithCoexistenceResults:    []*domain.DetectionResult{},
		detectWithCoexistenceResultsSet: true,
	}
	svc := services.NewDetectionService(mock)

	winner, _, err := svc.DetectWithCoexistenceSelection(context.Background(), "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if winner == nil {
		t.Fatal("expected unknown result, got nil")
	}
	if winner.HasCoexistenceWarning() {
		t.Error("unknown result should not have coexistence warning (AC5)")
	}
}
```

### Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/core/domain/detection_result.go` | MODIFY | Add CoexistenceWarning, CoexistenceMessage fields + methods |
| `internal/core/domain/detection_result_test.go` | CREATE | Unit tests for new fields and methods |
| `internal/core/services/detection_service.go` | MODIFY | Set warning on tie case, add exported constant |
| `internal/core/services/detection_service_test.go` | MODIFY | Add coexistence warning tests including AC6 boundary |

### Story Dependencies

| Dependency | Status | Relationship |
|------------|--------|--------------|
| Story 14.1 (DetectWithCoexistence) | COMPLETED | Prerequisite - provides multi-detection |
| Story 14.2 (ArtifactTimestamp) | COMPLETED | Prerequisite - provides timestamp field |
| Story 14.3 (Most-Recent-Wins) | COMPLETED | Prerequisite - provides SelectByTimestamp |
| Story 14.5 (TUI Coexistence Display) | Pending | This story enables 14.5 - TUI will use `HasCoexistenceWarning()` and `CoexistenceMessage` |

### Critical Implementation Notes

1. **Tie detection already works:** `SelectByTimestamp` returns `(nil, false)` for timestamps within 1 hour. No changes needed to selection.go.

2. **Create new pointers for warning results:** Since `WithCoexistenceWarning` returns a value (not modifying in place), create new pointers:
   ```go
   modified := r.WithCoexistenceWarning(msg)
   warningResults[i] = &modified
   ```

3. **AC4/AC5 handled by Go zero values:** Single methodology and unknown cases skip the tie code path, so `CoexistenceWarning=false` by default.

4. **Exported constant for TUI:** `CoexistenceWarningMessage` is exported so Story 14.5 (TUI) can reference it for consistency.

### Anti-Patterns to Avoid

| Don't | Do Instead | Why |
|-------|------------|-----|
| Modify `SelectByTimestamp` to add warning | Handle in service layer | Domain selector only determines winner, service interprets |
| Use `*results[i] = results[i].With...` | Create new slice with new pointers | Clearer mutation semantics |
| Create new struct for coexistence | Add fields to `DetectionResult` | Keep API simple, avoid breaking changes |
| Return different type for tie case | Use existing return + flag | TUI can check `HasCoexistenceWarning()` uniformly |
| Skip AC6 boundary test in service | Add explicit 1-hour boundary test | Critical edge case for tie detection |

### References

| Document | Section |
|----------|---------|
| PRD | `docs/prd-phase2.md` - FR-P2-10: Display coexistence warning |
| Epic | `docs/epics-phase2.md` - Story 2.4 (mapped to 14.4) |
| Previous Story | Story 14.3 - SelectByTimestamp and tie detection |
| Architecture | `docs/architecture.md` - Hexagonal boundaries |
| Project Context | `docs/project-context.md` - Critical implementation rules |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

### Completion Notes List

- Added `CoexistenceWarning bool` and `CoexistenceMessage string` fields to `DetectionResult` struct
- Added `WithCoexistenceWarning(msg string) DetectionResult` builder method (returns copy)
- Added `HasCoexistenceWarning() bool` accessor method
- Added `CoexistenceWarningMessage` constant to services package
- Updated `DetectWithCoexistenceSelection` to set coexistence warning on tie case results
- All 6 ACs satisfied and tested:
  - AC1: Tie detection triggers warning flag (tested)
  - AC2: Warning includes standard message (tested)
  - AC3: Clear winner has no warning (tested)
  - AC4: Single methodology has no warning (tested)
  - AC5: No methodologies (unknown) has no warning (tested)
  - AC6: Exact 1-hour boundary triggers warning (tested)
- All 1332 tests pass, lint passes

### File List

- `internal/core/domain/detection_result.go` (MODIFIED) - Added CoexistenceWarning fields and methods
- `internal/core/domain/detection_result_test.go` (MODIFIED) - Added coexistence warning tests; code review: enhanced field preservation test, added zero value checks
- `internal/core/services/detection_service.go` (MODIFIED) - Added constant and tie case warning logic
- `internal/core/services/detection_service_test.go` (MODIFIED) - Added coexistence warning tests
