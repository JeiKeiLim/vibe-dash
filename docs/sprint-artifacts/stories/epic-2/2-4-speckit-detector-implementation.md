# Story 2.4: Speckit Detector Implementation

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | New file `internal/adapters/detectors/speckit/detector.go` |
| **Key Dependencies** | `ports.MethodDetector`, `domain.DetectionResult`, `domain.Stage`, `domain.Confidence` |
| **Files to Create** | `detector.go`, `detector_test.go`, `stages.go` (optional), `registry.go` |
| **Location** | `internal/adapters/detectors/speckit/` + `internal/adapters/detectors/registry.go` |
| **Interfaces to Implement** | `ports.MethodDetector` |
| **Target Accuracy** | 95% (launch blocker) |

### Quick Task Summary (6 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Create speckit detector structure | Implement `ports.MethodDetector` interface |
| 2 | Implement CanDetect method | Quick check for Speckit markers |
| 3 | Implement Detect method | Full stage detection with confidence |
| 4 | Create detector registry | Central detector registration |
| 5 | Populate test fixtures | Real Speckit directory structures |
| 6 | Tests + 95% accuracy validation | Table-driven tests against fixtures |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Package location | `internal/adapters/detectors/speckit/` | Per architecture, detectors are adapters |
| Interface | `ports.MethodDetector` | Already defined in Story 1.3 |
| Detection markers | `specs/`, `.speckit/`, `.specify/` | Per PRD Speckit methodology |
| Stage detection | File presence heuristics | `spec.md`, `plan.md`, `tasks.md`, `implement.md` |
| Multiple specs | Most recently modified wins | FR10: Use most recent spec directory |
| Confidence levels | Certain, Likely, Uncertain | Domain types from Story 1.2 |
| Registry pattern | `internal/adapters/detectors/registry.go` | Architecture: single coordinator |

## Story

**As a** system,
**I want** to detect Speckit methodology and stage,
**So that** users see accurate workflow state.

## Acceptance Criteria

```gherkin
AC1: Given a project directory with specs/ OR .speckit/ OR .specify/
     When checking for Speckit methodology
     Then CanDetect() returns true
     And Name() returns "speckit"

AC2: Given specs/NNN-feature/ contains only spec.md
     When Detect() is called
     Then Stage = StageSpecify
     And Confidence = ConfidenceCertain
     And Reasoning = "spec.md exists, no plan.md"

AC3: Given specs/NNN-feature/ contains spec.md + plan.md
     When Detect() is called
     Then Stage = StagePlan
     And Confidence = ConfidenceCertain
     And Reasoning = "plan.md exists, no tasks.md"

AC4: Given specs/NNN-feature/ contains spec.md + plan.md + tasks.md
     When Detect() is called
     Then Stage = StageTasks
     And Confidence = ConfidenceCertain
     And Reasoning = "tasks.md exists"

AC5: Given specs/NNN-feature/ contains implement.md
     When Detect() is called
     Then Stage = StageImplement
     And Confidence = ConfidenceCertain
     And Reasoning = "implement.md exists"

AC6: Given multiple spec directories exist
     When Detect() is called
     Then most recently modified spec directory is used for stage
     And Reasoning mentions which directory was used

AC7: Given artifacts are ambiguous (e.g., partial.md only)
     When Detect() is called
     Then Confidence = ConfidenceUncertain
     And Reasoning explains the ambiguity

AC8: Given no Speckit markers found (no specs/, .speckit/, .specify/)
     When CanDetect() is called
     Then returns false

AC9: Given detection runs
     When ctx is cancelled
     Then detection stops promptly (within 100ms)
     And returns ctx.Err()

AC10: Given test fixtures run
      When accuracy is calculated
      Then accuracy >= 95% (launch blocker)
```

## Tasks / Subtasks

- [x] **Task 1: Create speckit detector structure** (AC: 1, 8)
  - [x] 1.1 Create `internal/adapters/detectors/speckit/` directory
  - [x] 1.2 Create `detector.go` with struct implementing `ports.MethodDetector`
  - [x] 1.3 Implement `Name() string` returning `"speckit"`
  - [x] 1.4 Define Speckit marker directories: `specs/`, `.speckit/`, `.specify/`
  - [x] 1.5 Use `New*` constructor pattern: `func NewSpeckitDetector() *SpeckitDetector`

- [x] **Task 2: Implement CanDetect method** (AC: 1, 8)
  - [x] 2.1 Check if any marker directory exists at path:
    - `{path}/specs/`
    - `{path}/.speckit/`
    - `{path}/.specify/`
  - [x] 2.2 Use `os.Stat()` or `os.ReadDir()` for existence check
  - [x] 2.3 Return `true` if any marker found, `false` otherwise
  - [x] 2.4 Respect context cancellation (`ctx.Done()`)

- [x] **Task 3: Implement Detect method** (AC: 2, 3, 4, 5, 6, 7, 9)
  - [x] 3.1 Find the specs directory (first match: `specs/`, `.speckit/`, `.specify/`)
  - [x] 3.2 List all spec subdirectories (e.g., `001-feature/`, `002-auth/`)
  - [x] 3.3 **If multiple subdirectories:** Find most recently modified
    - Use `os.Stat()` to get ModTime
    - Select directory with latest modification
    - Include in reasoning which directory was used
  - [x] 3.4 Analyze artifacts in selected directory:
    - `implement.md` present → StageImplement + ConfidenceCertain
    - `tasks.md` present → StageTasks + ConfidenceCertain
    - `plan.md` present → StagePlan + ConfidenceCertain
    - `spec.md` present only → StageSpecify + ConfidenceCertain
    - Partial files → StageUnknown + ConfidenceUncertain
  - [x] 3.5 Build reasoning string explaining detection logic
  - [x] 3.6 Return `*domain.DetectionResult` with all fields populated
  - [x] 3.7 Handle errors gracefully (return error, not panic)
  - [x] 3.8 Respect context cancellation throughout

- [x] **Task 4: Create detector registry** (AC: N/A - architecture requirement)
  - [x] 4.1 Create `internal/adapters/detectors/registry.go`
  - [x] 4.2 Define `Registry` struct holding slice of `ports.MethodDetector`
  - [x] 4.3 Implement `Register(detector ports.MethodDetector)`
  - [x] 4.4 Implement `DetectAll(ctx context.Context, path string) (*domain.DetectionResult, error)`
    - Iterate detectors, call `CanDetect()` then `Detect()` on first match
    - Return first successful detection
    - If no detector matches, return result with Method="unknown"
  - [x] 4.5 Use `New*` constructor: `func NewRegistry() *Registry`

- [x] **Task 5: Populate test fixtures** (AC: 10)
  - [x] 5.1 Update `test/fixtures/speckit-stage-specify/`:
    - Create `specs/001-feature/spec.md` with minimal content
  - [x] 5.2 Update `test/fixtures/speckit-stage-plan/`:
    - Create `specs/001-feature/spec.md`
    - Create `specs/001-feature/plan.md`
  - [x] 5.3 Update `test/fixtures/speckit-stage-tasks/`:
    - Create `specs/001-feature/spec.md`
    - Create `specs/001-feature/plan.md`
    - Create `specs/001-feature/tasks.md`
  - [x] 5.4 Update `test/fixtures/speckit-stage-implement/`:
    - Create `specs/001-feature/spec.md`
    - Create `specs/001-feature/plan.md`
    - Create `specs/001-feature/tasks.md`
    - Create `specs/001-feature/implement.md`
  - [x] 5.5 Update `test/fixtures/speckit-uncertain/`:
    - Create `specs/001-feature/partial.md` (no standard files)
  - [x] 5.6 Verify `test/fixtures/no-method-detected/` has no specs/ directory
  - [x] 5.7 Verify `test/fixtures/empty-project/` is empty

- [x] **Task 6: Write tests and validate accuracy** (AC: all)
  - [x] 6.1 Create `internal/adapters/detectors/speckit/detector_test.go`
  - [x] 6.2 Test: `Name()` returns "speckit"
  - [x] 6.3 Test: `CanDetect()` returns true for specs/ directory
  - [x] 6.4 Test: `CanDetect()` returns true for .speckit/ directory
  - [x] 6.5 Test: `CanDetect()` returns true for .specify/ directory
  - [x] 6.6 Test: `CanDetect()` returns false for no markers
  - [x] 6.7 Test: `Detect()` returns StageSpecify for spec.md only
  - [x] 6.8 Test: `Detect()` returns StagePlan for spec.md + plan.md
  - [x] 6.9 Test: `Detect()` returns StageTasks for spec.md + plan.md + tasks.md
  - [x] 6.10 Test: `Detect()` returns StageImplement for implement.md present
  - [x] 6.11 Test: `Detect()` returns ConfidenceUncertain for ambiguous case
  - [x] 6.12 Test: `Detect()` uses most recent directory when multiple exist
  - [x] 6.13 Test: Context cancellation stops detection (returns context.Canceled)
  - [x] 6.14 **Test: Context cancellation timing (AC9 - within 100ms)**
  - [x] 6.15 **Test: Failure paths (non-existent path, empty specs dir)**
  - [x] 6.16 Create `internal/adapters/detectors/registry_test.go`
  - [x] 6.17 Test: Registry iterates detectors correctly
  - [x] 6.18 Test: Registry returns unknown when no detector matches
  - [x] 6.19 **Create accuracy test**: Run against ALL fixtures, calculate percentage
  - [x] 6.20 Run `make build`, `make lint`, `make test`
  - [x] 6.21 Verify accuracy >= 95%

## Dev Notes

### CRITICAL: Implement MethodDetector Interface Exactly

The interface is already defined in `internal/core/ports/detector.go`:

```go
type MethodDetector interface {
    Name() string
    CanDetect(ctx context.Context, path string) bool
    Detect(ctx context.Context, path string) (*domain.DetectionResult, error)
}
```

### CRITICAL: Use Existing Domain Types

All required types exist in `internal/core/domain/`:

```go
// Stage (from stage.go)
domain.StageUnknown
domain.StageSpecify
domain.StagePlan
domain.StageTasks
domain.StageImplement

// Confidence (from confidence.go)
domain.ConfidenceUncertain
domain.ConfidenceLikely
domain.ConfidenceCertain

// DetectionResult (from detection_result.go)
domain.NewDetectionResult(method, stage, confidence, reasoning)
```

### Implementation Pattern for Detector

```go
package speckit

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "sort"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// SpeckitDetector implements ports.MethodDetector for Speckit methodology
type SpeckitDetector struct {
    // marker directories to check for Speckit presence
    markerDirs []string
}

// NewSpeckitDetector creates a new Speckit detector
func NewSpeckitDetector() *SpeckitDetector {
    return &SpeckitDetector{
        markerDirs: []string{"specs", ".speckit", ".specify"},
    }
}

// Name returns the detector identifier
func (d *SpeckitDetector) Name() string {
    return "speckit"
}

// CanDetect checks if any Speckit marker directory exists
func (d *SpeckitDetector) CanDetect(ctx context.Context, path string) bool {
    select {
    case <-ctx.Done():
        return false
    default:
    }

    for _, marker := range d.markerDirs {
        markerPath := filepath.Join(path, marker)
        if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
            return true
        }
    }
    return false
}

// Detect performs full Speckit methodology detection
func (d *SpeckitDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
    // Check context first
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Find the specs directory
    specsDir := ""
    for _, marker := range d.markerDirs {
        markerPath := filepath.Join(path, marker)
        if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
            specsDir = markerPath
            break
        }
    }

    if specsDir == "" {
        // No marker found - shouldn't happen if CanDetect was called first
        return nil, fmt.Errorf("no speckit markers found at %s", path)
    }

    // Find spec subdirectories
    entries, err := os.ReadDir(specsDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read specs directory: %w", err)
    }

    // Filter to directories only
    var specDirs []os.DirEntry
    for _, entry := range entries {
        if entry.IsDir() {
            specDirs = append(specDirs, entry)
        }
    }

    if len(specDirs) == 0 {
        // Empty specs directory
        result := domain.NewDetectionResult(
            d.Name(),
            domain.StageUnknown,
            domain.ConfidenceUncertain,
            "specs directory exists but contains no spec subdirectories",
        )
        return &result, nil
    }

    // Find most recently modified spec directory
    targetDir, reasoning := d.findMostRecentDir(specsDir, specDirs)

    // Check context again before file analysis
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Analyze artifacts in target directory
    return d.analyzeSpecDir(filepath.Join(specsDir, targetDir), reasoning)
}

// findMostRecentDir finds the most recently modified directory
func (d *SpeckitDetector) findMostRecentDir(baseDir string, dirs []os.DirEntry) (string, string) {
    if len(dirs) == 1 {
        return dirs[0].Name(), ""
    }

    type dirMod struct {
        name    string
        modTime int64
    }

    var dirMods []dirMod
    for _, dir := range dirs {
        info, err := dir.Info()
        if err != nil {
            continue
        }
        dirMods = append(dirMods, dirMod{name: dir.Name(), modTime: info.ModTime().Unix()})
    }

    // Sort by modification time descending
    sort.Slice(dirMods, func(i, j int) bool {
        return dirMods[i].modTime > dirMods[j].modTime
    })

    if len(dirMods) == 0 {
        return dirs[0].Name(), ""
    }

    reasoning := fmt.Sprintf("using most recently modified: %s", dirMods[0].name)
    return dirMods[0].name, reasoning
}

// analyzeSpecDir determines the stage based on artifact files
func (d *SpeckitDetector) analyzeSpecDir(dirPath string, extraReasoning string) (*domain.DetectionResult, error) {
    // Check for artifact files (order matters: check highest stage first)
    hasImplement := d.fileExists(filepath.Join(dirPath, "implement.md"))
    hasTasks := d.fileExists(filepath.Join(dirPath, "tasks.md"))
    hasPlan := d.fileExists(filepath.Join(dirPath, "plan.md"))
    hasSpec := d.fileExists(filepath.Join(dirPath, "spec.md"))

    var stage domain.Stage
    var confidence domain.Confidence
    var reasoning string

    switch {
    case hasImplement:
        stage = domain.StageImplement
        confidence = domain.ConfidenceCertain
        reasoning = "implement.md exists"
    case hasTasks:
        stage = domain.StageTasks
        confidence = domain.ConfidenceCertain
        reasoning = "tasks.md exists"
    case hasPlan:
        stage = domain.StagePlan
        confidence = domain.ConfidenceCertain
        reasoning = "plan.md exists, no tasks.md"
    case hasSpec:
        stage = domain.StageSpecify
        confidence = domain.ConfidenceCertain
        reasoning = "spec.md exists, no plan.md"
    default:
        stage = domain.StageUnknown
        confidence = domain.ConfidenceUncertain
        reasoning = "no standard Speckit artifacts found"
    }

    // Append extra reasoning if present
    if extraReasoning != "" {
        reasoning = reasoning + " (" + extraReasoning + ")"
    }

    result := domain.NewDetectionResult(d.Name(), stage, confidence, reasoning)
    return &result, nil
}

// fileExists checks if a file exists
func (d *SpeckitDetector) fileExists(path string) bool {
    info, err := os.Stat(path)
    return err == nil && !info.IsDir()
}
```

### Implementation Pattern for Registry

```go
package detectors

import (
    "context"
    "fmt"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Registry manages registered method detectors
type Registry struct {
    detectors []ports.MethodDetector
}

// NewRegistry creates a new detector registry
func NewRegistry() *Registry {
    return &Registry{
        detectors: make([]ports.MethodDetector, 0),
    }
}

// Register adds a detector to the registry
func (r *Registry) Register(detector ports.MethodDetector) {
    r.detectors = append(r.detectors, detector)
}

// DetectAll tries each registered detector until one succeeds
func (r *Registry) DetectAll(ctx context.Context, path string) (*domain.DetectionResult, error) {
    for _, detector := range r.detectors {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }

        if detector.CanDetect(ctx, path) {
            result, err := detector.Detect(ctx, path)
            if err == nil && result != nil {
                return result, nil
            }
            // Log error but continue to next detector
        }
    }

    // No detector matched
    result := domain.NewDetectionResult(
        "unknown",
        domain.StageUnknown,
        domain.ConfidenceUncertain,
        "no methodology markers found",
    )
    return &result, nil
}

// Detectors returns the list of registered detectors
func (r *Registry) Detectors() []ports.MethodDetector {
    return r.detectors
}
```

### Test Pattern (Table-Driven with Fixtures)

```go
package speckit_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors/speckit"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestSpeckitDetector_Name(t *testing.T) {
    d := speckit.NewSpeckitDetector()
    if got := d.Name(); got != "speckit" {
        t.Errorf("Name() = %q, want %q", got, "speckit")
    }
}

func TestSpeckitDetector_CanDetect(t *testing.T) {
    tests := []struct {
        name     string
        fixture  string
        expected bool
    }{
        {"specs directory present", "speckit-stage-specify", true},
        {"no markers present", "no-method-detected", false},
        {"empty project", "empty-project", false},
    }

    d := speckit.NewSpeckitDetector()
    ctx := context.Background()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fixturePath := filepath.Join("..", "..", "..", "..", "test", "fixtures", tt.fixture)
            got := d.CanDetect(ctx, fixturePath)
            if got != tt.expected {
                t.Errorf("CanDetect(%s) = %v, want %v", tt.fixture, got, tt.expected)
            }
        })
    }
}

func TestSpeckitDetector_Detect(t *testing.T) {
    tests := []struct {
        name           string
        fixture        string
        expectedStage  domain.Stage
        expectedConf   domain.Confidence
    }{
        {"specify stage", "speckit-stage-specify", domain.StageSpecify, domain.ConfidenceCertain},
        {"plan stage", "speckit-stage-plan", domain.StagePlan, domain.ConfidenceCertain},
        {"tasks stage", "speckit-stage-tasks", domain.StageTasks, domain.ConfidenceCertain},
        {"implement stage", "speckit-stage-implement", domain.StageImplement, domain.ConfidenceCertain},
        {"uncertain case", "speckit-uncertain", domain.StageUnknown, domain.ConfidenceUncertain},
    }

    d := speckit.NewSpeckitDetector()
    ctx := context.Background()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fixturePath := filepath.Join("..", "..", "..", "..", "test", "fixtures", tt.fixture)
            result, err := d.Detect(ctx, fixturePath)

            if err != nil {
                t.Fatalf("Detect() error = %v", err)
            }
            if result.Stage != tt.expectedStage {
                t.Errorf("Detect().Stage = %v, want %v", result.Stage, tt.expectedStage)
            }
            if result.Confidence != tt.expectedConf {
                t.Errorf("Detect().Confidence = %v, want %v", result.Confidence, tt.expectedConf)
            }
            if result.Method != "speckit" {
                t.Errorf("Detect().Method = %q, want %q", result.Method, "speckit")
            }
            if result.Reasoning == "" {
                t.Error("Detect().Reasoning should not be empty")
            }
        })
    }
}

func TestSpeckitDetector_ContextCancellation(t *testing.T) {
    d := speckit.NewSpeckitDetector()

    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately

    fixturePath := filepath.Join("..", "..", "..", "..", "test", "fixtures", "speckit-stage-specify")
    _, err := d.Detect(ctx, fixturePath)

    if err != context.Canceled {
        t.Errorf("Detect() with cancelled context should return context.Canceled, got %v", err)
    }
}

// TestSpeckitDetector_ContextCancellationTiming verifies AC9: cancellation responds within 100ms
func TestSpeckitDetector_ContextCancellationTiming(t *testing.T) {
    d := speckit.NewSpeckitDetector()
    fixturePath := filepath.Join("..", "..", "..", "..", "test", "fixtures", "speckit-stage-specify")

    ctx, cancel := context.WithCancel(context.Background())

    // Start detection in goroutine
    done := make(chan error, 1)
    go func() {
        _, err := d.Detect(ctx, fixturePath)
        done <- err
    }()

    // Cancel after brief delay to ensure detection has started
    time.Sleep(10 * time.Millisecond)
    cancelStart := time.Now()
    cancel()

    // Wait for completion with timeout
    select {
    case <-done:
        elapsed := time.Since(cancelStart)
        // AC9: Should respond within 100ms of cancellation
        if elapsed > 100*time.Millisecond {
            t.Errorf("Cancellation took %v, expected < 100ms (AC9 requirement)", elapsed)
        }
    case <-time.After(200 * time.Millisecond):
        t.Error("Detection did not respond to cancellation within 200ms timeout")
    }
}

// TestSpeckitDetector_FailurePaths tests error handling for edge cases
func TestSpeckitDetector_FailurePaths(t *testing.T) {
    d := speckit.NewSpeckitDetector()
    ctx := context.Background()

    t.Run("non-existent path returns error", func(t *testing.T) {
        _, err := d.Detect(ctx, "/non/existent/path/that/does/not/exist")
        if err == nil {
            t.Error("expected error for non-existent path")
        }
    })

    t.Run("empty specs directory returns uncertain", func(t *testing.T) {
        // Create temp dir with empty specs/
        tmpDir := t.TempDir()
        specsDir := filepath.Join(tmpDir, "specs")
        if err := os.MkdirAll(specsDir, 0755); err != nil {
            t.Fatal(err)
        }

        result, err := d.Detect(ctx, tmpDir)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if result.Stage != domain.StageUnknown {
            t.Errorf("expected StageUnknown, got %v", result.Stage)
        }
        if result.Confidence != domain.ConfidenceUncertain {
            t.Errorf("expected ConfidenceUncertain, got %v", result.Confidence)
        }
    })
}

// TestDetectionAccuracy runs against all fixtures and calculates accuracy
// This is the launch blocker test - must be >= 95%
func TestDetectionAccuracy(t *testing.T) {
    testCases := []struct {
        fixture       string
        expectedStage domain.Stage
        shouldDetect  bool // false for non-speckit fixtures
    }{
        {"speckit-stage-specify", domain.StageSpecify, true},
        {"speckit-stage-plan", domain.StagePlan, true},
        {"speckit-stage-tasks", domain.StageTasks, true},
        {"speckit-stage-implement", domain.StageImplement, true},
        {"speckit-uncertain", domain.StageUnknown, true}, // Uncertain is correct for ambiguous
        {"no-method-detected", domain.StageUnknown, false},
        {"empty-project", domain.StageUnknown, false},
    }

    d := speckit.NewSpeckitDetector()
    ctx := context.Background()

    correct := 0
    total := len(testCases)

    for _, tc := range testCases {
        fixturePath := filepath.Join("..", "..", "..", "..", "test", "fixtures", tc.fixture)

        canDetect := d.CanDetect(ctx, fixturePath)

        if tc.shouldDetect {
            if !canDetect {
                t.Logf("FAIL: %s - CanDetect returned false, expected true", tc.fixture)
                continue
            }

            result, err := d.Detect(ctx, fixturePath)
            if err != nil {
                t.Logf("FAIL: %s - Detect error: %v", tc.fixture, err)
                continue
            }

            if result.Stage == tc.expectedStage {
                correct++
                t.Logf("PASS: %s - Stage: %v", tc.fixture, result.Stage)
            } else {
                t.Logf("FAIL: %s - Got %v, expected %v", tc.fixture, result.Stage, tc.expectedStage)
            }
        } else {
            // Should NOT detect as Speckit
            if !canDetect {
                correct++
                t.Logf("PASS: %s - Correctly not detected as Speckit", tc.fixture)
            } else {
                t.Logf("FAIL: %s - Should not be detected as Speckit", tc.fixture)
            }
        }
    }

    accuracy := float64(correct) / float64(total) * 100
    t.Logf("\n=== DETECTION ACCURACY: %.1f%% (%d/%d) ===", accuracy, correct, total)

    if accuracy < 95.0 {
        t.Errorf("Detection accuracy %.1f%% is below 95%% launch blocker threshold", accuracy)
    }
}
```

### Test Fixture Structure (Must Create)

```
test/fixtures/
├── speckit-stage-specify/
│   └── specs/
│       └── 001-feature/
│           └── spec.md
├── speckit-stage-plan/
│   └── specs/
│       └── 001-feature/
│           ├── spec.md
│           └── plan.md
├── speckit-stage-tasks/
│   └── specs/
│       └── 001-feature/
│           ├── spec.md
│           ├── plan.md
│           └── tasks.md
├── speckit-stage-implement/
│   └── specs/
│       └── 001-feature/
│           ├── spec.md
│           ├── plan.md
│           ├── tasks.md
│           └── implement.md
├── speckit-uncertain/
│   └── specs/
│       └── 001-feature/
│           └── partial.md
├── no-method-detected/
│   └── README.md
└── empty-project/
```

### Previous Story Learnings (Story 2.3)

1. **Table-driven tests** - Use `tests := []struct{}` pattern for comprehensive coverage
2. **Context cancellation** - Always check `ctx.Done()` before and during long operations
3. **Error wrapping** - Use `fmt.Errorf("...: %w", err)` for context
4. **Domain types** - Return domain errors and use domain types consistently
5. **Test isolation** - Use `t.TempDir()` when creating test directories
6. **Constructor pattern** - Always use `New*` functions for struct creation

### Code Review Learnings (from Story 2.3)

**CRITICAL: Apply these patterns to avoid code review fix cycles:**

1. **Clean error messages** - Avoid double-wrapping errors:
   - Good: `fmt.Errorf("failed to read specs: %w", err)`
   - Bad: `fmt.Errorf("detection failed: %w: %w", path, err)`
2. **Test failure paths explicitly** - Include tests for:
   - Permission denied on specs directory
   - Corrupted/unreadable spec directories
   - Empty specs directory (no subdirectories)
3. **No stateful variables** - This detector has no flags, but if any are added, reset between tests

### Forward Dependencies

This detector will be consumed by:
- **Story 2.5: Detection Service** - Orchestrates this detector via `registry.DetectAll()`
- **Story 2.3: Add Project Command** - After Story 2.5, add command can call DetectionService to populate project detection fields

### Edge Case Handling

| Scenario | Expected Behavior |
|----------|-------------------|
| Empty specs directory (no subdirs) | Return StageUnknown + ConfidenceUncertain with reasoning |
| specs/ contains files not directories | Skip non-directories, analyze only subdirectories |
| Empty spec files (0 bytes) | Treat as if file exists (presence-based detection) |
| Unreadable directory (permission denied) | Return error, do not panic |
| Symlink loop in specs/ | os.ReadDir handles gracefully, return error if fails |
| Very deep directory nesting | Only analyze first-level subdirectories of specs/ |

### Performance Requirements

- Detection should complete in <100ms for typical projects (NFR-P1)
- Limit directory recursion to first-level subdirectories only (no deep scanning)
- Use `os.ReadDir()` not deprecated `ioutil.ReadDir()` (Go 1.16+ optimization)
- Context cancellation must respond within 100ms (AC9)

### File Paths (Relative from Project Root)

| File | Purpose |
|------|---------|
| `internal/adapters/detectors/speckit/detector.go` | Speckit detector implementation |
| `internal/adapters/detectors/speckit/detector_test.go` | Detector tests with accuracy validation |
| `internal/adapters/detectors/registry.go` | Central detector registry |
| `internal/adapters/detectors/registry_test.go` | Registry tests |
| `test/fixtures/speckit-stage-*/specs/001-feature/*.md` | Speckit test fixtures |

### Architecture Compliance Checklist

- [x] Detector in `internal/adapters/detectors/speckit/` (correct adapter layer)
- [x] Implements `ports.MethodDetector` interface exactly
- [x] Uses domain types (`domain.Stage`, `domain.Confidence`, `domain.DetectionResult`)
- [x] Uses `New*` constructor pattern
- [x] Context propagation (uses `ctx context.Context`)
- [x] Returns `*domain.DetectionResult`, not value
- [x] Registry in `internal/adapters/detectors/registry.go` (coordinator role)
- [x] 95% accuracy validated (launch blocker) - 100% achieved

### Project Structure Notes

**Alignment with unified project structure:**
- `internal/adapters/detectors/` directory follows Architecture section "Project Structure"
- `registry.go` is the ONLY component that knows about all detector implementations
- Detector package structure: `speckit/detector.go`, `speckit/detector_test.go`

**Files to Create:**
```
internal/adapters/detectors/
├── registry.go           # Detector registry (new)
├── registry_test.go      # Registry tests (new)
└── speckit/
    ├── detector.go       # Speckit detector (new)
    └── detector_test.go  # Detector tests (new)
```

### References

- [Source: docs/epics.md#Story 2.4: Speckit Detector Implementation] - Full acceptance criteria
- [Source: docs/architecture.md#Plugin Architecture] - MethodDetector interface pattern
- [Source: docs/architecture.md#Registry Coordination Role] - Registry pattern
- [Source: docs/architecture.md#95% Detection Accuracy Measurement] - Launch blocker definition
- [Source: docs/project-context.md#Hexagonal Architecture Boundaries] - Adapter placement
- [Source: docs/project-context.md#Testing Rules] - 95% accuracy requirement
- [Source: internal/core/ports/detector.go] - MethodDetector interface definition
- [Source: internal/core/domain/stage.go] - Stage enum
- [Source: internal/core/domain/confidence.go] - Confidence enum
- [Source: internal/core/domain/detection_result.go] - DetectionResult struct

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 2.4 requirements, lines 689-747)
- docs/architecture.md (Plugin architecture, registry pattern, accuracy measurement)
- docs/project-context.md (Go patterns, testing rules, accuracy requirement)
- internal/core/ports/detector.go (MethodDetector interface)
- internal/core/domain/stage.go (Stage enum)
- internal/core/domain/confidence.go (Confidence enum)
- internal/core/domain/detection_result.go (DetectionResult struct)
- docs/sprint-artifacts/2-3-add-project-command.md (Previous story learnings)
- test/fixtures/ (Existing fixture structure with .keep files)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Validation Notes

Story drafted by SM agent with comprehensive context:
- All 10 acceptance criteria mapped from epics.md
- Detection logic aligned with PRD Speckit methodology
- 95% accuracy requirement emphasized as launch blocker
- Implementation examples provided with full context cancellation handling
- Test patterns include accuracy calculation
- Registry pattern follows Architecture document exactly

Ultimate context engine analysis completed - comprehensive developer guide created

### Story Validation (2025-12-13)

**Validated by:** SM Agent (Bob) using `validate-workflow.xml`
**Checklist:** `create-story/checklist.md`
**Result:** 28/32 items passed (87.5%) → Improvements applied → All critical issues resolved

**Improvements Applied:**
1. ✅ **Code Review Learnings** - Added section with patterns from Story 2.3 code review fixes
2. ✅ **Context Cancellation Timing Test** - Added test verifying AC9 100ms requirement
3. ✅ **Failure Paths Tests** - Added tests for non-existent path and empty specs directory
4. ✅ **Forward Dependencies** - Added section documenting Story 2.5 consumption
5. ✅ **Edge Case Handling** - Added table with 6 edge case scenarios and expected behavior
6. ✅ **Performance Requirements** - Added section with <100ms detection requirement
7. ✅ **Task 6 Updated** - Added subtasks 6.14 and 6.15 for new tests

**Story Status:** Ready for implementation with comprehensive developer guidance

### Implementation Notes (2025-12-13)

**Implemented by:** Dev Agent (Amelia) - Claude Opus 4.5

**Implementation Summary:**
- Created `internal/adapters/detectors/speckit/detector.go` implementing `ports.MethodDetector`
- Created `internal/adapters/detectors/registry.go` for detector coordination
- Populated all test fixtures with proper Speckit directory structures
- Comprehensive test coverage with 100% detection accuracy achieved

**Key Implementation Decisions:**
1. Used package-level `markerDirs` slice for Speckit markers
2. Detection order: `specs/` → `.speckit/` → `.specify/` (first match wins)
3. Stage priority: implement.md → tasks.md → plan.md → spec.md
4. Context cancellation checked at multiple points for prompt response (<100ms)
5. Registry continues to next detector on error (resilient design)

**Tests Created:**
- `detector_test.go`: 14 test functions covering all ACs
- `registry_test.go`: 10 test functions for registry behavior
- `TestDetectionAccuracy`: Launch blocker test (100% accuracy achieved)

### Completion Notes

All 6 tasks completed successfully:
1. ✅ Speckit detector structure with `New*` constructor pattern
2. ✅ CanDetect method with context cancellation support
3. ✅ Detect method with multi-directory support and stage detection
4. ✅ Registry with first-match-wins detection coordination
5. ✅ Test fixtures populated with realistic Speckit structures
6. ✅ Comprehensive tests with 100% detection accuracy (exceeds 95% threshold)

## File List

| File | Operation |
|------|-----------|
| `internal/adapters/detectors/speckit/detector.go` | Created, Modified (code review fixes) |
| `internal/adapters/detectors/speckit/detector_test.go` | Created, Modified (code review fixes) |
| `internal/adapters/detectors/registry.go` | Created, Modified (code review fixes) |
| `internal/adapters/detectors/registry_test.go` | Created, Modified (code review fixes) |
| `test/fixtures/speckit-stage-specify/specs/001-feature/spec.md` | Created |
| `test/fixtures/speckit-stage-plan/specs/001-feature/spec.md` | Created |
| `test/fixtures/speckit-stage-plan/specs/001-feature/plan.md` | Created |
| `test/fixtures/speckit-stage-tasks/specs/001-feature/spec.md` | Created |
| `test/fixtures/speckit-stage-tasks/specs/001-feature/plan.md` | Created |
| `test/fixtures/speckit-stage-tasks/specs/001-feature/tasks.md` | Created |
| `test/fixtures/speckit-stage-implement/specs/001-feature/spec.md` | Created |
| `test/fixtures/speckit-stage-implement/specs/001-feature/plan.md` | Created |
| `test/fixtures/speckit-stage-implement/specs/001-feature/tasks.md` | Created |
| `test/fixtures/speckit-stage-implement/specs/001-feature/implement.md` | Created |
| `test/fixtures/speckit-uncertain/specs/001-feature/partial.md` | Created |
| `test/fixtures/speckit-dotspeckit-marker/.speckit/001-feature/spec.md` | Created (code review) |
| `test/fixtures/speckit-dotspecify-marker/.specify/001-feature/spec.md` | Created (code review) |
| `test/fixtures/speckit-dotspecify-marker/.specify/001-feature/plan.md` | Created (code review) |
| `test/fixtures/no-method-detected/README.md` | Modified |
| `docs/sprint-artifacts/sprint-status.yaml` | Modified |
| `docs/sprint-artifacts/2-4-speckit-detector-implementation.md` | Modified |

## Change Log

| Date | Change |
|------|--------|
| 2025-12-13 | Story created with ready-for-dev status |
| 2025-12-13 | **Validation improvements applied:** Added Code Review Learnings from Story 2.3, Forward Dependencies section, Edge Case Handling table, Performance Requirements, Context Cancellation Timing test (AC9), Failure Paths tests. Updated Task 6 with new test subtasks (6.14, 6.15). |
| 2025-12-13 | **Implementation complete:** All 6 tasks completed. Speckit detector and registry created with comprehensive tests. 100% detection accuracy achieved. All tests pass, lint clean, build successful. Status: Ready for Review. |
| 2025-12-13 | **Code review fixes applied:** (H1) Added thread-safety documentation to registry.go. (H2) DetectAll now collects and reports detector errors in reasoning. (H3) findMostRecentDir now provides explanation when falling back to first directory. (M1) MultipleDirectories test now verifies reasoning mentions directory name and selection criteria. (M3) Added .speckit and .specify marker fixtures to accuracy test (9 fixtures, 100% accuracy). (M5) Documented constructor pattern choice for markerDirs. (L2) Marked Architecture Compliance Checklist as complete. Status: Done. |
