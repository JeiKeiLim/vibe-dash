# Story 2.5: Detection Service

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | New file `internal/core/services/detection_service.go` |
| **Key Dependencies** | `ports.MethodDetector`, `domain.DetectionResult`, `adapters/detectors.Registry` |
| **Files to Create** | `detection_service.go`, `detection_service_test.go` |
| **Location** | `internal/core/services/` |
| **Interfaces to Implement** | None (creates a service that uses ports.MethodDetector) |
| **Integration Point** | Orchestrates `adapters/detectors.Registry.DetectAll()` |

### Quick Task Summary (5 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Create DetectionService structure | Service with registry dependency injection |
| 2 | Implement Detect method | Delegates to registry with context support |
| 3 | Implement DetectAll method for multiple paths | Batch detection with error aggregation |
| 4 | Integrate with add project command | Wire service into CLI layer |
| 5 | Tests + integration validation | Table-driven tests with mocks |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Package location | `internal/core/services/` | Per architecture, services are in core |
| Registry dependency | Injected via constructor | Hexagonal architecture - services don't know about adapters |
| Error handling | Return domain errors | Use `domain.ErrDetectionFailed` wrapping |
| Multiple detection | Collect all results | FR14: Detect multiple methodologies |
| Context propagation | Required first parameter | All service methods support cancellation |

## Story

**As a** system,
**I want** a detection service that orchestrates detectors,
**So that** multiple methodologies can be supported.

## Acceptance Criteria

```gherkin
AC1: Given DetectionService with registered detectors
     When Detect(ctx, path) is called
     Then the registry's DetectAll() is invoked
     And DetectionResult is returned

AC2: Given multiple detectors registered
     When Detect(path) is called
     Then detectors are tried via registry
     And first matching detector's result is returned

AC3: Given request to detect multiple methodologies (post-MVP support)
     When DetectMultiple(ctx, path) is called
     Then ALL detectors are checked (not just first match)
     And all matching results are returned
     And FR14 is satisfied (multiple methodologies)

AC4: Given no detector matches
     When Detect(path) is called
     Then DetectionResult with:
       - Method = "unknown"
       - Stage = StageUnknown
       - Confidence = ConfidenceUncertain
       - Reasoning = "No methodology markers found"

AC5: Given detector throws error
     When Detect(path) is called
     Then error is logged
     And detection continues to next detector
     And partial results returned if any succeed

AC6: Given context is cancelled
     When Detect(ctx, path) is called
     Then detection stops promptly (within 100ms)
     And returns ctx.Err()

AC7: Given DetectionService is created
     When constructor is called
     Then service is properly initialized with injected registry
     And follows New* constructor pattern
```

## Tasks / Subtasks

- [x] **Task 1: Create DetectionService structure** (AC: 1, 7)
  - [x] 1.1 Create `internal/core/services/detection_service.go`
  - [x] 1.2 Define `DetectionService` struct with registry interface field
  - [x] 1.3 Create registry interface in `internal/core/ports/detection.go` for dependency injection:
    - `DetectorRegistry` interface with `DetectAll(ctx, path) (*DetectionResult, error)`
  - [x] 1.4 Implement `NewDetectionService(registry DetectorRegistry) *DetectionService`
  - [x] 1.5 Document service purpose and thread-safety guarantees

- [x] **Task 2: Implement Detect method** (AC: 1, 2, 4, 5, 6)
  - [x] 2.1 Implement `Detect(ctx context.Context, path string) (*domain.DetectionResult, error)`
  - [x] 2.2 Check context cancellation at entry
  - [x] 2.3 Delegate to registry.DetectAll()
  - [x] 2.4 Handle registry errors gracefully (wrap with domain error)
  - [x] 2.5 Ensure result is never nil (return unknown if registry fails)

- [x] **Task 3: Implement DetectMultiple method** (AC: 3)
  - [x] 3.1 Implement `DetectMultiple(ctx context.Context, path string) ([]*domain.DetectionResult, error)`
  - [x] 3.2 Iterate all registered detectors (not first-match)
  - [x] 3.3 Collect all successful detection results
  - [x] 3.4 Return slice of results (may be empty)
  - [x] 3.5 Document as preparation for FR14 (multiple methodologies)

- [x] **Task 4: Integrate with add project command** (AC: N/A - integration)
  - [x] 4.1 Update `cmd/vibe/main.go` to create DetectionService
  - [x] 4.2 Wire DetectionService into CLI commands that need detection
  - [x] 4.3 Update `vibe add` command to use DetectionService
  - [x] 4.4 Ensure detection results populate project entity

- [x] **Task 5: Write tests** (AC: all)
  - [x] 5.1 Create `internal/core/services/detection_service_test.go`
  - [x] 5.2 Test: Constructor creates valid service
  - [x] 5.3 Test: Constructor panics with nil registry
  - [x] 5.4 Test: Detect() delegates to registry
  - [x] 5.5 Test: Detect() returns unknown when no detector matches
  - [x] 5.6 Test: Detect() handles context cancellation
  - [x] 5.7 Test: Detect() cancellation timing < 100ms (AC6)
  - [x] 5.8 Test: Detect() returns error for empty path
  - [x] 5.9 Test: DetectMultiple() collects all matching results
  - [x] 5.10 Test: DetectMultiple() returns empty slice when no matches
  - [x] 5.11 Test: DetectMultiple() returns empty slice when all detectors fail
  - [x] 5.12 Test: DetectMultiple() returns error for empty path
  - [x] 5.13 Create mock registry for testing
  - [x] 5.14 Run `make build`, `make lint`, `make test`

## Dev Notes

### CRITICAL: Hexagonal Architecture - Service in Core, Registry in Adapters

The `DetectionService` lives in `internal/core/services/` and MUST NOT import from `internal/adapters/`. To achieve this:

1. **Define an interface** in `internal/core/ports/detection.go` that the service depends on
2. **Adapters implement** this interface (Registry already does the work)
3. **Injection at startup** in `cmd/vibe/main.go` wires the concrete Registry to the service

```go
// internal/core/ports/detection.go
package ports

import (
    "context"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// DetectorRegistry coordinates detection across multiple MethodDetectors.
// This interface is implemented by adapters/detectors.Registry.
type DetectorRegistry interface {
    // DetectAll tries each registered detector until one succeeds.
    // Returns a result with Method="unknown" if no detector matches.
    DetectAll(ctx context.Context, path string) (*domain.DetectionResult, error)

    // Detectors returns all registered detectors for multi-methodology detection.
    Detectors() []MethodDetector
}
```

**CRITICAL: Interface-Implementation Relationship**

The existing `adapters/detectors.Registry` struct **already implements** this interface - no code changes needed there. The `Registry` struct has:
- `DetectAll(ctx, path)` method (lines 45-85 in registry.go)
- `Detectors()` method (lines 40-42 in registry.go)

This port interface exists to allow `DetectionService` (in core) to depend on an **abstraction** rather than the concrete `Registry` (in adapters). This is the hexagonal architecture pattern - adapters implement ports, core depends only on ports.

### Service Implementation Pattern

```go
// internal/core/services/detection_service.go
package services

import (
    "context"
    "fmt"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// DetectionService orchestrates methodology detection across registered detectors.
// It provides the core business logic for workflow detection.
//
// Thread Safety: Safe for concurrent use. All methods are stateless and
// delegate to the underlying registry which handles its own thread safety.
type DetectionService struct {
    registry ports.DetectorRegistry
}

// NewDetectionService creates a new detection service with the given registry.
// Panics if registry is nil - this is a programming error that should be caught early.
func NewDetectionService(registry ports.DetectorRegistry) *DetectionService {
    if registry == nil {
        panic("DetectionService requires non-nil registry")
    }
    return &DetectionService{
        registry: registry,
    }
}

// Detect performs methodology detection on the given path.
// Returns the first successful detection result, or a result with
// Method="unknown" if no detector matches.
//
// Return type is *DetectionResult (pointer) for single detection.
// See DetectMultiple for []*DetectionResult (slice of pointers) when
// detecting multiple methodologies.
func (s *DetectionService) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
    // Validate path is not empty
    if path == "" {
        return nil, fmt.Errorf("%w: empty path", domain.ErrPathNotAccessible)
    }

    // Check context at entry
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    result, err := s.registry.DetectAll(ctx, path)
    if err != nil {
        // Wrap with domain error for consistent error handling
        return nil, fmt.Errorf("%w: %v", domain.ErrDetectionFailed, err)
    }

    return result, nil
}

// DetectMultiple checks ALL registered detectors and returns all matching results.
// This supports FR14: detecting multiple methodologies in the same project.
//
// Return type is []*DetectionResult (slice of pointers) because multiple
// methodologies may be detected. See Detect for single-methodology detection.
//
// Error Handling: DetectMultiple continues checking all detectors even if some fail.
// Individual detector errors are logged but don't stop iteration. Returns empty
// slice if no detectors match (not an error - just no methodologies found).
func (s *DetectionService) DetectMultiple(ctx context.Context, path string) ([]*domain.DetectionResult, error) {
    // Validate path is not empty
    if path == "" {
        return nil, fmt.Errorf("%w: empty path", domain.ErrPathNotAccessible)
    }

    // Check context at entry
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    var results []*domain.DetectionResult

    for _, detector := range s.registry.Detectors() {
        // Check context before each detector
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }

        if detector.CanDetect(ctx, path) {
            result, err := detector.Detect(ctx, path)
            if err == nil && result != nil {
                results = append(results, result)
            }
            // Continue to check other detectors even if one fails (resilient design)
            // Errors are silently skipped - detector errors don't propagate
        }
    }

    return results, nil
}
```

### Integration with Add Command

The `vibe add` command currently exists but doesn't perform detection. Wire the detection service:

```go
// cmd/vibe/main.go (simplified wiring)

import (
    "github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors"
    "github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors/speckit"
    "github.com/JeiKeiLim/vibe-dash/internal/core/services"
)

func main() {
    // 1. Create registry and register detectors
    registry := detectors.NewRegistry()
    registry.Register(speckit.NewSpeckitDetector())

    // 2. Create detection service with registry
    detectionSvc := services.NewDetectionService(registry)

    // 3. Pass to CLI commands that need it
    // ...
}
```

### Test Pattern (Mock Registry)

```go
// internal/core/services/detection_service_test.go
package services_test

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
    "github.com/JeiKeiLim/vibe-dash/internal/core/services"
)

// mockRegistry implements ports.DetectorRegistry for testing
type mockRegistry struct {
    detectAllResult *domain.DetectionResult
    detectAllError  error
    detectors       []ports.MethodDetector
}

func (m *mockRegistry) DetectAll(ctx context.Context, path string) (*domain.DetectionResult, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    return m.detectAllResult, m.detectAllError
}

func (m *mockRegistry) Detectors() []ports.MethodDetector {
    return m.detectors
}

func TestDetectionService_Detect(t *testing.T) {
    tests := []struct {
        name           string
        registryResult *domain.DetectionResult
        registryError  error
        expectError    bool
        expectMethod   string
    }{
        {
            name: "successful detection",
            registryResult: func() *domain.DetectionResult {
                r := domain.NewDetectionResult("speckit", domain.StagePlan, domain.ConfidenceCertain, "plan.md found")
                return &r
            }(),
            expectMethod: "speckit",
        },
        {
            name: "no detector matches returns unknown",
            registryResult: func() *domain.DetectionResult {
                r := domain.NewDetectionResult("unknown", domain.StageUnknown, domain.ConfidenceUncertain, "no markers")
                return &r
            }(),
            expectMethod: "unknown",
        },
        {
            name:          "registry error wraps as domain error",
            registryError: errors.New("internal error"),
            expectError:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockRegistry{
                detectAllResult: tt.registryResult,
                detectAllError:  tt.registryError,
            }
            svc := services.NewDetectionService(mock)

            result, err := svc.Detect(context.Background(), "/some/path")

            if tt.expectError {
                if err == nil {
                    t.Error("expected error, got nil")
                }
                if !errors.Is(err, domain.ErrDetectionFailed) {
                    t.Errorf("expected ErrDetectionFailed, got %v", err)
                }
                return
            }

            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if result.Method != tt.expectMethod {
                t.Errorf("Method = %q, want %q", result.Method, tt.expectMethod)
            }
        })
    }
}

func TestDetectionService_ContextCancellation(t *testing.T) {
    mock := &mockRegistry{
        detectAllResult: func() *domain.DetectionResult {
            r := domain.NewDetectionResult("speckit", domain.StagePlan, domain.ConfidenceCertain, "")
            return &r
        }(),
    }
    svc := services.NewDetectionService(mock)

    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately

    _, err := svc.Detect(ctx, "/some/path")
    if err != context.Canceled {
        t.Errorf("expected context.Canceled, got %v", err)
    }
}

// TestDetectionService_CancellationTiming verifies AC6: cancellation responds within 100ms
func TestDetectionService_CancellationTiming(t *testing.T) {
    mock := &mockRegistry{
        detectAllResult: func() *domain.DetectionResult {
            r := domain.NewDetectionResult("speckit", domain.StagePlan, domain.ConfidenceCertain, "")
            return &r
        }(),
    }
    svc := services.NewDetectionService(mock)

    ctx, cancel := context.WithCancel(context.Background())

    // Start detection in goroutine
    done := make(chan error, 1)
    go func() {
        _, err := svc.Detect(ctx, "/some/path")
        done <- err
    }()

    // Cancel after brief delay
    time.Sleep(10 * time.Millisecond)
    cancelStart := time.Now()
    cancel()

    // Wait for completion
    select {
    case <-done:
        elapsed := time.Since(cancelStart)
        if elapsed > 100*time.Millisecond {
            t.Errorf("Cancellation took %v, expected < 100ms (AC6)", elapsed)
        }
    case <-time.After(200 * time.Millisecond):
        t.Error("Detection did not respond to cancellation within 200ms")
    }
}

func TestDetectionService_NilRegistry(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Error("expected panic for nil registry")
        }
    }()
    services.NewDetectionService(nil)
}

func TestDetectionService_EmptyPath(t *testing.T) {
    mock := &mockRegistry{}
    svc := services.NewDetectionService(mock)

    _, err := svc.Detect(context.Background(), "")
    if err == nil {
        t.Error("expected error for empty path")
    }
    if !errors.Is(err, domain.ErrPathNotAccessible) {
        t.Errorf("expected ErrPathNotAccessible, got %v", err)
    }
}

func TestDetectionService_DetectMultiple_AllFail(t *testing.T) {
    // mockDetector that always fails
    failingDetector := &mockDetector{
        canDetect: true,
        detectErr: errors.New("always fails"),
    }
    mock := &mockRegistry{
        detectors: []ports.MethodDetector{failingDetector},
    }
    svc := services.NewDetectionService(mock)

    results, err := svc.DetectMultiple(context.Background(), "/some/path")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(results) != 0 {
        t.Errorf("expected empty slice, got %d results", len(results))
    }
}

// mockDetector for DetectMultiple tests
type mockDetector struct {
    canDetect    bool
    detectResult *domain.DetectionResult
    detectErr    error
}

func (m *mockDetector) Name() string { return "mock" }
func (m *mockDetector) CanDetect(ctx context.Context, path string) bool { return m.canDetect }
func (m *mockDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
    return m.detectResult, m.detectErr
}
```

### Previous Story Learnings (Story 2.4)

From the Story 2.4 code review:

1. **Thread-safety documentation** - Document thread safety guarantees in service docstrings
2. **Error aggregation** - When multiple operations can fail, collect and report all errors
3. **Context cancellation timing** - Test that cancellation responds within 100ms (AC6)
4. **Failure paths** - Test error conditions explicitly (registry errors, empty results)
5. **Clean constructor pattern** - Use `New*` functions consistently
6. **Domain error wrapping** - Use `fmt.Errorf("%w: %v", domain.Error, err)` for context

### Code Review Learnings (from Story 2.3, 2.4)

**CRITICAL: Apply these patterns to avoid code review fix cycles:**

1. **Interface in ports, implementation in adapters** - DetectorRegistry interface in ports, Registry struct in adapters
2. **Wrap errors with domain types** - Return `ErrDetectionFailed` not raw errors
3. **Document thread safety** - Add explicit documentation about concurrent access
4. **Test cancellation timing** - Include test verifying <100ms cancellation response

### Forward Dependencies

This service will be consumed by:
- **Story 2.6: Project Name Collision Handling** - Uses detection to show methodology in collision UI
- **Story 2.7: List Projects Command** - Displays detection results in list output
- **Story 2.9: Path Validation at Launch** - Re-runs detection when path updated

### Edge Case Handling

| Scenario | Expected Behavior |
|----------|-------------------|
| Empty path string | Return error (invalid path) |
| Non-existent path | Delegate to registry, registry handles error |
| Permission denied on path | Registry handles, returns error in reasoning |
| Registry returns nil result | Service ensures non-nil result (unknown) |
| Context cancelled mid-detection | Return immediately with ctx.Err() |
| All detectors fail | Return unknown result with aggregated errors in reasoning |

### Performance Requirements

- Detection should complete in <100ms for typical projects (NFR-P1)
- Context cancellation must respond within 100ms (AC6)
- Service adds minimal overhead (<1ms) on top of registry

### File Paths (Relative from Project Root)

| File | Purpose |
|------|---------|
| `internal/core/ports/detection.go` | DetectorRegistry interface definition |
| `internal/core/services/detection_service.go` | DetectionService implementation |
| `internal/core/services/detection_service_test.go` | Service tests with mock registry |

### Architecture Compliance Checklist

- [ ] Service in `internal/core/services/` (correct core layer)
- [ ] Interface in `internal/core/ports/detection.go` for registry
- [ ] NO imports from `internal/adapters/` in service file
- [ ] Uses domain types (`domain.DetectionResult`, `domain.ErrDetectionFailed`)
- [ ] Uses `New*` constructor pattern
- [ ] Context propagation (uses `ctx context.Context`)
- [ ] Thread-safety documented
- [ ] Domain error wrapping for external failures

### Project Structure Notes

**Alignment with unified project structure:**
- Service in `internal/core/services/` per Architecture section
- Interface defined in `internal/core/ports/` for adapter injection
- Follows hexagonal architecture - core doesn't import adapters

**Files to Create:**
```
internal/core/
├── ports/
│   └── detection.go           # DetectorRegistry interface (new)
└── services/
    ├── detection_service.go   # DetectionService (new)
    └── detection_service_test.go  # Service tests (new)
```

### References

- [Source: docs/epics.md#Story 2.5: Detection Service] - Full acceptance criteria
- [Source: docs/architecture.md#Registry Coordination Role] - Registry pattern, "Services call registry, not detectors directly"
- [Source: docs/architecture.md#Architectural Boundaries] - Hexagonal architecture rules
- [Source: docs/architecture.md#Context Propagation] - Context usage in services
- [Source: docs/project-context.md#Hexagonal Architecture Boundaries] - Core imports nothing from adapters
- [Source: docs/project-context.md#Go Patterns] - Context first, New* constructors, error wrapping
- [Source: internal/core/ports/detector.go] - MethodDetector interface
- [Source: internal/adapters/detectors/registry.go] - Existing Registry implementation
- [Source: docs/sprint-artifacts/2-4-speckit-detector-implementation.md] - Previous story patterns

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 2.5 requirements, lines 751-791)
- docs/architecture.md (Registry pattern, hexagonal boundaries, service structure)
- docs/project-context.md (Go patterns, hexagonal rules)
- internal/core/ports/detector.go (MethodDetector interface)
- internal/adapters/detectors/registry.go (Existing Registry - to be wrapped by interface)
- docs/sprint-artifacts/2-4-speckit-detector-implementation.md (Previous story learnings)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Validation Notes

Story drafted by SM agent with comprehensive context:
- All 7 acceptance criteria mapped from epics.md
- Hexagonal architecture strictly followed (interface in ports, service in core)
- Previous story learnings incorporated
- Implementation examples provided with full context cancellation handling
- Test patterns include mock registry
- Integration path documented (cmd/vibe/main.go wiring)

Ultimate context engine analysis completed - comprehensive developer guide created

### Story Validation (2025-12-13)

**Validated by:** SM Agent (Bob) using `validate-workflow.xml`
**Checklist:** `create-story/checklist.md`
**Result:** 26/28 items passed (93%) → Improvements applied → All issues resolved

**Improvements Applied:**
1. ✅ **Interface-Implementation Relationship** - Added clarification that existing Registry already implements DetectorRegistry
2. ✅ **DetectMultiple Error Handling** - Clarified resilient design: continues on failure, returns empty slice
3. ✅ **Empty Path Validation** - Added to both Detect() and DetectMultiple() methods
4. ✅ **Nil Registry Protection** - Constructor now panics on nil registry (fail-fast)
5. ✅ **Cancellation Timing Test** - Added test verifying AC6 (<100ms response)
6. ✅ **DetectMultiple All Fail Test** - Added test for empty results when all detectors fail
7. ✅ **Return Type Documentation** - Documented *DetectionResult vs []*DetectionResult
8. ✅ **New Test Subtasks** - Added 5.3, 5.7, 5.8, 5.11, 5.12 to Task 5

**Story Status:** Ready for implementation with comprehensive developer guidance

### Completion Notes

**Implementation completed (2025-12-13) by Dev Agent (Amelia):**

1. **DetectorRegistry interface** added to `internal/core/ports/detector.go` - provides abstraction for Registry without breaking hexagonal architecture
2. **DetectionService** created in `internal/core/services/detection_service.go`:
   - `NewDetectionService(registry)` constructor with nil-check panic
   - `Detect(ctx, path)` for single-methodology detection with context support
   - `DetectMultiple(ctx, path)` for FR14 multi-methodology detection
   - Thread-safety documented, domain error wrapping implemented
3. **Integration with CLI** - `vibe add` now performs detection and displays methodology/stage
4. **Comprehensive tests** - 14 test functions covering all ACs, including cancellation timing test

**All tests pass, lint clean, build succeeds.**

### File List

| File | Operation |
|------|-----------|
| `internal/core/ports/detector.go` | Modified (added DetectorRegistry + Detector interfaces) |
| `internal/core/services/detection_service.go` | Created |
| `internal/core/services/detection_service_test.go` | Created |
| `cmd/vibe/main.go` | Modified (wire detection service) |
| `internal/adapters/cli/add.go` | Modified (use detection service via ports.Detector interface) |
| `internal/adapters/cli/add_test.go` | Modified (added detection integration tests) |
| `internal/adapters/detectors/registry.go` | Modified (added interface compliance check) |
| `docs/sprint-artifacts/sprint-status.yaml` | Modified (status tracking) |

## Change Log

| Date | Change |
|------|--------|
| 2025-12-13 | Story created with ready-for-dev status |
| 2025-12-13 | **Validation improvements applied:** (1) CRITICAL: Added Interface-Implementation Relationship clarification explaining existing Registry already implements DetectorRegistry. (2) CRITICAL: Clarified DetectMultiple error handling - continues on failure, returns empty slice. (3) Added empty path validation to Detect() and DetectMultiple(). (4) Added nil registry protection to NewDetectionService(). (5) Added cancellation timing test verifying AC6 (<100ms). (6) Added test for DetectMultiple when all detectors fail. (7) Documented return type differences (*DetectionResult vs []*DetectionResult). (8) Added new test subtasks 5.3, 5.7, 5.8, 5.11, 5.12. |
| 2025-12-13 | **Implementation completed:** All 5 tasks completed. DetectorRegistry interface in ports, DetectionService with Detect/DetectMultiple methods, CLI integration with vibe add command, 14 comprehensive tests. All tests pass, lint clean. Status: Ready for Review |
| 2025-12-13 | **Code Review Fixes Applied:** (M1) Added slog.Debug logging for detector errors in DetectMultiple per AC5. (M3) Added 3 CLI integration tests for detection: TestAdd_WithDetectionService, TestAdd_DetectionFailureIsNonFatal, TestAdd_WithoutDetectionService. (M4) Created ports.Detector interface, refactored CLI to use interface instead of concrete type for testability. (L2) Added compile-time interface compliance checks for Registry and DetectionService. Updated File List with all modified files. |
