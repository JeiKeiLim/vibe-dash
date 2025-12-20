# Story 4.5.1: BMAD v6 Detector Implementation

Status: done

## Story

As a developer using BMAD Method,
I want vibe-dash to detect my BMAD v6 project structure,
so that my project appears with the correct methodology identified.

## Acceptance Criteria

1. Given a project directory with `.bmad/` folder AND `.bmad/bmm/config.yaml` exists, when MethodDetector.Detect() is called, then returns MethodInfo with Method="bmad", version extracted from config.yaml, and Confidence=High
2. Given a project directory without `.bmad/` folder, when MethodDetector.Detect() is called, then returns nil (not a BMAD project)
3. Given a project with `.bmad-core/` (v4 structure), when MethodDetector.Detect() is called, then returns nil (v4 not supported in this story)
4. Given `.bmad/` exists but `bmm/config.yaml` is missing, then detection should still succeed with a warning or lower confidence
5. Given context is cancelled during detection, the detector returns ctx.Err() promptly

## Tasks / Subtasks

- [x] Task 1: Create detector package structure (AC: #1, #2, #3)
  - [x] Create `internal/adapters/detectors/bmad/detector.go`
  - [x] Create `internal/adapters/detectors/bmad/detector_test.go`
  - [x] Define markerDirs constant: `[]string{".bmad"}`

- [x] Task 2: Implement MethodDetector interface (AC: #1, #2, #3, #5)
  - [x] Implement `Name() string` returning "bmad"
  - [x] Implement `CanDetect(ctx, path) bool` - ONLY check if `.bmad/` folder exists (fast O(1) check)
  - [x] Implement `Detect(ctx, path) (*domain.DetectionResult, error)` - check config.yaml, extract version
  - [x] Handle context cancellation with select statements

- [x] Task 3: Extract version from config.yaml header (AC: #1, #4)
  - [x] Read `.bmad/bmm/config.yaml` as raw text (NOT yaml.Unmarshal - version is in comment)
  - [x] Extract version from file header comment using regex `# Version: (\S+)`
  - [x] Include version in detection reasoning: "BMAD v6.0.0-alpha.13 detected"
  - [x] Handle missing config.yaml gracefully (still detect, lower confidence)

- [x] Task 4: Register detector in registry (AC: #1)
  - [x] Import `bmad` package in `cmd/vibe/main.go`
  - [x] Add `registry.Register(bmad.NewBMADDetector())` after line 98 (after speckit registration)
  - [x] Note: Registration order determines priority - BMAD registered after Speckit

- [x] Task 5: Write unit tests (AC: #1, #2, #3, #4, #5)
  - [x] Test CanDetect with valid `.bmad/bmm/config.yaml`
  - [x] Test CanDetect without `.bmad/` folder
  - [x] Test CanDetect with `.bmad-core/` (should return false)
  - [x] Test Detect with full v6 structure
  - [x] Test Detect with missing config.yaml
  - [x] Test context cancellation

## Dev Notes

### Architecture Compliance

- **Hexagonal architecture**: Detector goes in `internal/adapters/detectors/bmad/`
- **MethodDetector interface**: Defined in `internal/core/ports/detector.go`
- **Registry pattern**: All detectors register with `internal/adapters/detectors/registry.go`
- **Zero core imports from adapters**: Detector only imports from `internal/core/domain` and `internal/core/ports`

### Out of Scope (Story 4.5-2)

**DO NOT implement in this story:**
- Stage detection from `sprint-status.yaml`
- Parsing epic/story status values
- BMAD phase-to-stage mapping
- Any logic involving `docs/sprint-artifacts/sprint-status.yaml`

**Always return `domain.StageUnknown` for the Stage field.**

### MethodDetector Interface (from internal/core/ports/detector.go)

```go
type MethodDetector interface {
    // Name returns unique identifier: "bmad"
    Name() string

    // CanDetect performs QUICK check - just verify .bmad/ folder exists
    // Should be O(1) - single os.Stat call on directory
    // Do NOT check for config.yaml here (that's Detect's job)
    CanDetect(ctx context.Context, path string) bool

    // Detect performs FULL detection with config parsing
    // Called only if CanDetect returned true
    // Returns nil, error if detection cannot be performed
    // Returns result with ConfidenceLikely if config.yaml missing
    Detect(ctx context.Context, path string) (*domain.DetectionResult, error)
}
```

### DO / DON'T Quick Reference

| DO | DON'T |
|----|-------|
| Return `domain.StageUnknown` always | Implement stage detection (Story 4.5-2) |
| Use `os.ReadFile` for version extraction | Use `yaml.Unmarshal` (version is in comment) |
| Check `.bmad/` folder in CanDetect | Check config.yaml in CanDetect |
| Use `domain.NewDetectionResult()` constructor | Construct struct manually |
| Follow SpeckitDetector patterns exactly | Invent new patterns |
| Use `regexp.MustCompile` for version regex | Parse version from YAML fields |

### Existing Patterns to Follow

Reference implementation: `internal/adapters/detectors/speckit/detector.go`

```go
// Pattern from SpeckitDetector
type BMADDetector struct{}

func NewBMADDetector() *BMADDetector {
    return &BMADDetector{}
}

func (d *BMADDetector) Name() string {
    return "bmad"
}

// CanDetect: FAST check - only verify .bmad/ folder exists
func (d *BMADDetector) CanDetect(ctx context.Context, path string) bool {
    select {
    case <-ctx.Done():
        return false
    default:
    }
    // ONLY check if .bmad/ folder exists (O(1) operation)
    // Do NOT check for config.yaml here - that's Detect's responsibility
    bmadPath := filepath.Join(path, ".bmad")
    info, err := os.Stat(bmadPath)
    return err == nil && info.IsDir()
}
```

### BMAD v6 Structure (from vibe-dash .bmad/)

```
.bmad/
├── _cfg/                    # IDE/agent configurations
├── core/                    # Core BMAD infrastructure
├── docs/                    # BMAD documentation
└── bmm/                     # BMAD Method Module
    ├── config.yaml          # Version: 6.0.0-alpha.13
    ├── agents/              # Agent definitions
    └── workflows/           # Phase-organized workflows
        ├── 1-analysis/      # Research, product-brief
        ├── 2-plan-workflows/# PRD, UX design
        ├── 3-solutioning/   # Architecture, epics-and-stories
        └── 4-implementation/# dev-story, code-review, sprint-planning
```

### BMAD v6 Detection Markers

**Primary marker (high confidence):**
- Path: `.bmad/bmm/config.yaml`
- Contains: Version header comment `# Version: 6.x.x`
- Contains: `project_name`, `bmad_folder`, etc.

**Secondary marker (medium confidence):**
- Path: `.bmad/` folder exists
- But `bmm/config.yaml` is missing (incomplete install)

**NOT supported (v4):**
- Path: `.bmad-core/` folder (deferred for future)
- Path: `tools/core-config.yaml` or `tools/bmad-config.yaml`

### Version Extraction (IMPORTANT: Not YAML Parsing)

The version is in the file's **COMMENT HEADER**, not in YAML content:

```yaml
# BMM Module Configuration
# Generated by BMAD installer
# Version: 6.0.0-alpha.13  <- Extract THIS via regex
# Date: 2025-12-04T00:10:41.176Z

project_name: bmad-test    <- NOT from here (this is YAML content)
```

**Implementation pattern:**
```go
func extractVersion(configPath string) (string, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return "", err
    }

    re := regexp.MustCompile(`# Version:\s*(\S+)`)
    match := re.FindSubmatch(data)
    if match == nil {
        return "", nil // No version found, not an error
    }
    return string(match[1]), nil
}
```

**DO NOT use `yaml.Unmarshal`** - the version field is not part of the YAML structure.

### Return Values (MUST Use Constructor)

**Always use `domain.NewDetectionResult()` constructor** (defined in `internal/core/domain/detection_result.go`):

For successful detection (config.yaml found):
```go
result := domain.NewDetectionResult(
    "bmad",                    // Method name
    domain.StageUnknown,       // Stage - ALWAYS Unknown in this story
    domain.ConfidenceCertain,  // High confidence when config.yaml exists
    "BMAD v6.0.0-alpha.13 detected (.bmad/bmm/config.yaml found)",
)
return &result, nil
```

For missing config.yaml but `.bmad/` exists:
```go
result := domain.NewDetectionResult(
    "bmad",
    domain.StageUnknown,       // Stage - ALWAYS Unknown in this story
    domain.ConfidenceLikely,   // Lower confidence
    ".bmad folder exists but config.yaml not found",
)
return &result, nil
```

### Registry Registration (cmd/vibe/main.go)

Add after line 96 (after speckit registration):

```go
import (
    // ... existing imports ...
    "github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors/bmad"
)

// In run() function, around line 95-97:
registry := detectors.NewRegistry()
registry.Register(speckit.NewSpeckitDetector())
registry.Register(bmad.NewBMADDetector())  // ADD THIS LINE
```

### Reference Implementation Note

**Source:** `github.com/ibadmore/bmad-progress-dashboard` targets BMAD **v4** (`.bmad-core/`).

Our detector targets **v6 only** (`.bmad/` folder). See `docs/sprint-change-proposal-2025-12-20.md` for full analysis.

### Testing Standards

- Co-locate tests: `detector_test.go` next to `detector.go`
- Table-driven tests using `tests []struct{...}` pattern
- Use temp directories with marker folders (not mocks)
- Test context cancellation explicitly

### Critical Constraints

1. **No stage detection** - Always return `domain.StageUnknown` (Story 4.5-2 handles stages)
2. **v6 only** - Do NOT support `.bmad-core/` (v4)
3. **No YAML parsing for version** - Use regex on file content (version is in comment header)
4. **Follow Speckit detector patterns exactly** - Same structure, same context handling
5. **Use `domain.NewDetectionResult()` constructor** - Do not construct struct manually

### References

- [Source: internal/adapters/detectors/speckit/detector.go] - Reference implementation
- [Source: internal/core/ports/detector.go] - MethodDetector interface
- [Source: internal/adapters/detectors/registry.go] - Registry pattern
- [Source: docs/architecture.md#Plugin-Architecture] - Plugin design
- [Source: docs/epics.md#Epic-4.5] - Epic requirements
- [Source: .bmad/bmm/config.yaml] - Real-world v6 config example
- [Source: docs/sprint-change-proposal-2025-12-20.md] - BMAD detection research and reference implementation analysis

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

claude-opus-4-5-20250101

### Debug Log References

### Completion Notes List

1. **Implementation Complete**: BMAD v6 detector implemented following SpeckitDetector patterns exactly
2. **All Tests Pass**: 11 unit tests covering all acceptance criteria (AC #1-5)
3. **Integration Verified**: `vibe add .` correctly detects BMAD v6 project with `Method: bmad (Unknown)`
4. **Stage Detection Deferred**: Returns `domain.StageUnknown` as specified (Story 4.5-2 scope)
5. **Version Extraction**: Uses regex on file header comment, not YAML parsing
6. **Linting Clean**: BMAD detector code passes golangci-lint with no warnings
7. **Code Review Fix**: Added compile-time interface compliance check `var _ ports.MethodDetector = (*BMADDetector)(nil)`

### File List

- `internal/adapters/detectors/bmad/detector.go` - BMAD v6 detector implementation
- `internal/adapters/detectors/bmad/detector_test.go` - Unit tests (11 tests)
- `cmd/vibe/main.go` - Added BMAD detector registration (lines 15, 98)

