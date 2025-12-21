# Story 4.5.2: BMAD v6 Stage Detection Logic

Status: done

## Story

As a developer using BMAD Method,
I want vibe-dash to show my current workflow stage,
so that I know where I am in the BMAD process.

## Acceptance Criteria

1. Given a BMAD v6 project with `sprint-status.yaml`, when stage detection runs, then parses `development_status` section and determines current phase from epic statuses
2. Given stage detection with epic/story analysis, then maps to standardized stages:
   - All epics backlog → "Specify"
   - Epic in-progress with stories → "Implement"
   - Stories in review → "Tasks" (indicates review phase)
   - All epics done → "Plan" (complete, ready for next cycle)
3. Given detection with reasoning, then provides explanation like: "Epic 4 in-progress, Story 4.3 being implemented"
4. Given no sprint-status.yaml found, then falls back to artifact detection (PRD/Architecture/Epics files)
5. Given context is cancelled during detection, the detector returns ctx.Err() promptly
6. Given malformed or empty sprint-status.yaml, then returns StageUnknown with ConfidenceUncertain

## Tasks / Subtasks

- [x] Task 1: Add sprint-status.yaml parsing (AC: #1, #2, #3, #5)
  - [x] Create `internal/adapters/detectors/bmad/stage_parser.go`
  - [x] Implement `parseSprintStatus(ctx, path) (*SprintStatus, error)`
  - [x] Parse YAML `development_status` section using SprintStatus struct
  - [x] Handle context cancellation with select statements between I/O

- [x] Task 2: Implement stage determination logic (AC: #1, #2, #3)
  - [x] Create `determineStageFromStatus(status *SprintStatus) (domain.Stage, domain.Confidence, string)`
  - [x] Implement epic status analysis:
    - Count epics by status (backlog, in-progress, done)
    - Find first in-progress epic
    - Find first in-progress story within that epic
  - [x] Map to stages per Stage Mapping Table
  - [x] Build reasoning string with current epic/story context

- [x] Task 3: Implement fallback artifact detection (AC: #4)
  - [x] Create `detectStageFromArtifacts(ctx, path) (domain.Stage, domain.Confidence, string, error)`
  - [x] Check context.Done() between each file scan
  - [x] Check for files in order: Epics → Architecture → PRD
  - [x] Return ConfidenceLikely for all artifact-based detection

- [x] Task 4: Integrate into existing Detect method (AC: #1, #4, #5, #6)
  - [x] Modify `detector.go` Detect method to call stage detection after version extraction
  - [x] Try sprint-status.yaml first, fallback to artifacts on error
  - [x] Update reasoning to include both version AND stage context
  - [x] Handle all error cases gracefully (never panic)

- [x] Task 5: Write comprehensive unit tests (AC: #1-6)
  - [x] Test stage detection from sprint-status.yaml
  - [x] Test all epic status combinations (see Test Matrix)
  - [x] Test story status parsing
  - [x] Test fallback to artifact detection
  - [x] Test malformed YAML handling (empty, syntax error, missing key)
  - [x] Test context cancellation (immediate and timeout)
  - [x] Test empty/missing files

## Dev Notes

### Required Import

Add to `stage_parser.go`:
```go
import (
    "context"
    "os"
    "path/filepath"
    "regexp"
    "strings"

    "gopkg.in/yaml.v3"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)
```

### SprintStatus Struct Definition

```go
// SprintStatus represents the parsed sprint-status.yaml file.
type SprintStatus struct {
    DevelopmentStatus map[string]string `yaml:"development_status"`
}

// parseSprintStatus reads and parses the sprint-status.yaml file.
func parseSprintStatus(ctx context.Context, path string) (*SprintStatus, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    var status SprintStatus
    if err := yaml.Unmarshal(data, &status); err != nil {
        return nil, err
    }

    return &status, nil
}
```

### Context Cancellation Pattern (MANDATORY)

**Copy this pattern from Story 4.5.1 exactly:**

```go
// Check context BEFORE every I/O operation
select {
case <-ctx.Done():
    return nil, ctx.Err()
default:
}

// Perform I/O operation
data, err := os.ReadFile(path)
if err != nil {
    return nil, err
}

// Check context AFTER every I/O operation
select {
case <-ctx.Done():
    return nil, ctx.Err()
default:
}
```

### Sprint Status YAML Structure

Location: `docs/sprint-artifacts/sprint-status.yaml` (configurable via `sprint_artifacts` in config.yaml)

```yaml
development_status:
  # Epic 1: Foundation & First Launch
  epic-1: done
  1-1-project-scaffolding: done
  1-2-domain-entities: done

  # Epic 4.5: BMAD Method v6 State Detection
  epic-4-5: in-progress
  4-5-1-bmad-v6-detector-implementation: done
  4-5-2-bmad-v6-stage-detection-logic: ready-for-dev
  4-5-3-bmad-test-fixtures: backlog
```

### Key Pattern Recognition

**Epic key regex:** `^epic-\d+(-\d+)?$` (matches `epic-1`, `epic-4-5`)
**Story key regex:** `^\d+-\d+-` (matches `1-1-project-scaffolding`, `4-5-2-bmad-v6-...`)
**Skip pattern:** `*-retrospective` (not relevant for stage detection)

**Epic status values:** `backlog`, `in-progress`, `contexted`, `done`
**Story status values:** `backlog`, `drafted`, `ready-for-dev`, `in-progress`, `review`, `done`

### Stage Mapping Table (Complete Return Value Contract)

| # | Condition | domain.Stage | domain.Confidence | Reasoning Template |
|---|-----------|--------------|-------------------|-------------------|
| **Error Cases** |
| E1 | sprint-status.yaml missing | (fallback) | - | Falls back to artifact detection |
| E2 | sprint-status.yaml malformed | StageUnknown | ConfidenceUncertain | "sprint-status.yaml parse error" |
| E3 | sprint-status.yaml empty | StageUnknown | ConfidenceUncertain | "sprint-status.yaml is empty" |
| E4 | development_status key missing | StageUnknown | ConfidenceUncertain | "No development_status section" |
| **Epic-Level Cases** |
| 1 | All epics backlog | StageSpecify | ConfidenceCertain | "No epics in progress - planning phase" |
| 2 | All epics done | StageImplement | ConfidenceCertain | "All epics complete - project done" |
| 3 | Mixed: some done, none in-progress | StageSpecify | ConfidenceCertain | "No active epic - planning next" |
| **Story-Level Cases (Epic In-Progress)** |
| 4 | No stories | StagePlan | ConfidenceCertain | "Epic N started, preparing stories" |
| 5 | All stories backlog | StagePlan | ConfidenceCertain | "Epic N started, preparing stories" |
| 6 | Has drafted stories only | StagePlan | ConfidenceCertain | "Story X.Y drafted, awaiting approval" |
| 7 | Has ready-for-dev stories only | StagePlan | ConfidenceCertain | "Story X.Y ready for development" |
| 8 | Has in-progress story | StageImplement | ConfidenceCertain | "Story X.Y being implemented" |
| 9 | Has review story | StageTasks | ConfidenceCertain | "Story X.Y in code review" |
| 10 | Has in-progress AND review | StageTasks | ConfidenceCertain | "Story X.Y in code review" |
| **11** | **All stories done** | **StageImplement** | **ConfidenceCertain** | **"Epic N stories complete, update epic status"** |
| **Multi-Story Cases** |
| 12 | Multiple in-progress | StageImplement | ConfidenceCertain | "Story X.Y being implemented (+N more)" |
| 13 | Multiple review | StageTasks | ConfidenceCertain | "Story X.Y in code review (+N more)" |
| **Inconsistent State Cases** |
| 14 | Epic done, stories in-progress | StageImplement | ConfidenceLikely | "Epic done but Story X.Y in-progress" |
| 15 | Epic done, stories in review | StageTasks | ConfidenceLikely | "Epic done but Story X.Y in review" |
| 16 | Epic backlog, stories active | StageSpecify | ConfidenceLikely | "Epic backlog but Story X.Y active" |
| **Unknown Status Cases** |
| 17 | Unknown epic status | StageUnknown | ConfidenceUncertain | "Unknown epic status 'VALUE'" |
| 18 | Unknown story status | (continue) | - | "Unknown story status 'VALUE'" |
| **LLM Typo Normalization** |
| 19 | Status with spaces/underscores | (normalize) | - | Normalize "in progress" → "in-progress" |
| 20 | Synonym mapping | (normalize) | - | Map "complete"→"done", "wip"→"in-progress" |
| **Data Quality Warning Cases** |
| 21 | Orphan story (no epic match) | (warn) | ConfidenceLikely | "Story X.Y has no matching epic" |
| 22 | Deep story prefix (4-5-6-xxx) | (warn) | ConfidenceLikely | "Story prefix depth exceeds epic depth" |
| 23 | Empty status value `""` | (warn) | ConfidenceLikely | "Empty status for key X" |
| **Fallback Artifact Detection** |
| F1 | Has epic*.md | StageImplement | ConfidenceLikely | "Epics defined but no sprint status" |
| F2 | Has architecture*.md | StagePlan | ConfidenceLikely | "Architecture designed, no epics yet" |
| F3 | Has prd*.md | StageSpecify | ConfidenceLikely | "PRD created, architecture pending" |
| F4 | No artifacts | StageUnknown | ConfidenceUncertain | "No BMAD artifacts detected" |

### Status Normalization (LLM Typo Handling)

Sprint-status.yaml is LLM-generated during workflow execution. LLMs may produce variations that must be normalized before comparison:

```go
// normalizeStatus converts common variations to canonical status values.
// Apply BEFORE switch statement comparison.
func normalizeStatus(status string) string {
    // 1. Lowercase everything
    s := strings.ToLower(strings.TrimSpace(status))

    // 2. Normalize separators: spaces and underscores → hyphens
    s = strings.ReplaceAll(s, " ", "-")
    s = strings.ReplaceAll(s, "_", "-")

    // 3. Map synonyms
    synonyms := map[string]string{
        "complete":    "done",
        "completed":   "done",
        "finished":    "done",
        "wip":         "in-progress",
        "inprogress":  "in-progress",
        "reviewing":   "review",
        "in-review":   "review",
        "code-review": "review",
    }

    if canonical, ok := synonyms[s]; ok {
        return canonical
    }
    return s
}
```

**Case Normalization Rule:** Convert all status values to lowercase before comparison.

**Normalization Examples:**
| Input | Normalized Output |
|-------|-------------------|
| `"in progress"` | `"in-progress"` |
| `"In_Progress"` | `"in-progress"` |
| `"IN-PROGRESS"` | `"in-progress"` |
| `"complete"` | `"done"` |
| `"completed"` | `"done"` |
| `"finished"` | `"done"` |
| `"wip"` | `"in-progress"` |
| `"reviewing"` | `"review"` |
| `"in-review"` | `"review"` |
| `"code-review"` | `"review"` |

### Story Status Priority Order

When multiple stories have different statuses, use this priority (highest first):
1. `review` - someone waiting for feedback
2. `in-progress` - active development
3. `ready-for-dev` - queued for development
4. `drafted` - story being prepared
5. `backlog` - not started
6. `done` - already completed

### Fallback Detection Implementation

```go
func detectStageFromArtifacts(ctx context.Context, projectPath string) (domain.Stage, domain.Confidence, string, error) {
    docsPath := filepath.Join(projectPath, "docs")

    // Check context before scanning
    select {
    case <-ctx.Done():
        return domain.StageUnknown, domain.ConfidenceUncertain, "", ctx.Err()
    default:
    }

    // Check for epics file (highest priority - furthest along)
    epicPatterns := []string{"*epic*.md", "*Epic*.md"}
    for _, pattern := range epicPatterns {
        matches, _ := filepath.Glob(filepath.Join(docsPath, pattern))
        if len(matches) > 0 {
            return domain.StageImplement, domain.ConfidenceLikely, "Epics defined but no sprint status", nil
        }
    }

    select {
    case <-ctx.Done():
        return domain.StageUnknown, domain.ConfidenceUncertain, "", ctx.Err()
    default:
    }

    // Check for architecture file
    archPatterns := []string{"*architecture*.md", "*Architecture*.md"}
    for _, pattern := range archPatterns {
        matches, _ := filepath.Glob(filepath.Join(docsPath, pattern))
        if len(matches) > 0 {
            return domain.StagePlan, domain.ConfidenceLikely, "Architecture designed, no epics yet", nil
        }
    }

    select {
    case <-ctx.Done():
        return domain.StageUnknown, domain.ConfidenceUncertain, "", ctx.Err()
    default:
    }

    // Check for PRD file
    prdPatterns := []string{"*prd*.md", "*PRD*.md"}
    for _, pattern := range prdPatterns {
        matches, _ := filepath.Glob(filepath.Join(docsPath, pattern))
        if len(matches) > 0 {
            return domain.StageSpecify, domain.ConfidenceLikely, "PRD created, architecture pending", nil
        }
    }

    return domain.StageUnknown, domain.ConfidenceUncertain, "No BMAD artifacts detected", nil
}
```

### Sprint Status Location Discovery

Search order:
1. `{project}/docs/sprint-artifacts/sprint-status.yaml` (default BMAD location)
2. `{project}/docs/sprint-status.yaml` (alternative location)

### Integration into detector.go

**Modify the existing Detect() method after line 131:**

```go
// After building the initial result with version info...
// Now detect stage
stage, confidence, stageReasoning := d.detectStage(ctx, path)

// Combine version reasoning with stage reasoning
var fullReasoning string
if version != "" {
    fullReasoning = "BMAD v" + version + ", " + stageReasoning
} else {
    fullReasoning = "BMAD detected, " + stageReasoning
}

// Use the more confident confidence level
finalConfidence := domain.ConfidenceCertain
if confidence == domain.ConfidenceUncertain {
    finalConfidence = domain.ConfidenceLikely
}

result := domain.NewDetectionResult(
    d.Name(),
    stage,
    finalConfidence,
    fullReasoning,
)
return &result, nil
```

### DO / DON'T Quick Reference

| DO | DON'T |
|----|-------|
| Use existing domain.Stage values (StageSpecify, StagePlan, etc.) | Add new stages to domain/stage.go |
| Parse YAML with `gopkg.in/yaml.v3` and struct unmarshaling | Use regex to parse YAML content |
| Use case-insensitive string comparison for status values | Assume exact case matching |
| Return StageUnknown + ConfidenceUncertain on parse errors | Crash or panic on malformed YAML |
| Check `ctx.Done()` before AND after every I/O operation | Block indefinitely on file reads |
| Build detailed reasoning: "Epic 4 in-progress (Story 4.3 implementing)" | Return vague reasoning like "in progress" |
| Use `domain.NewDetectionResult()` constructor | Construct DetectionResult struct manually |
| Match Story 4.5.1 code style exactly | Invent new patterns |

### Test Matrix (Required Coverage)

**Sprint Status Parsing Tests (Original):**
| Test Case | Input | Expected Stage | Expected Confidence |
|-----------|-------|----------------|---------------------|
| All epics backlog | `epic-1: backlog, epic-2: backlog` | StageSpecify | ConfidenceCertain |
| One epic in-progress, stories backlog | `epic-1: in-progress, 1-1-x: backlog` | StagePlan | ConfidenceCertain |
| Story in-progress | `epic-1: in-progress, 1-1-x: in-progress` | StageImplement | ConfidenceCertain |
| Story in review | `epic-1: in-progress, 1-1-x: review` | StageTasks | ConfidenceCertain |
| All epics done | `epic-1: done, epic-2: done` | StageImplement | ConfidenceCertain |
| Mixed: some done, one in-progress | `epic-1: done, epic-2: in-progress` | (analyze epic-2) | ConfidenceCertain |

**P1 Gap Test Cases (Story 4.6.3):**
| Gap | Test Case | Input | Expected Stage | Expected Reasoning |
|-----|-----------|-------|----------------|--------------------|
| G1 | All stories done in in-progress epic | `epic-1: in-progress, 1-1-x: done, 1-2-x: done` | StageImplement | "Epic 1 stories complete, update epic status" |
| G1 | All stories done with multiple epics | `epic-1: done, epic-2: in-progress, 2-1-x: done` | StageImplement | "Epic 2 stories complete, update epic status" |
| G7 | Epic done, story in-progress | `epic-1: done, 1-1-x: in-progress` | StageImplement | "Epic done but Story 1.1 in-progress" |
| G7 | Epic done, story in review | `epic-1: done, 1-1-x: review` | StageTasks | "Epic done but Story 1.1 in review" |
| G15 | LLM typo "in progress" (space) | `epic-1: in progress, 1-1-x: in-progress` | StageImplement | "Story 1.1 being implemented" |
| G15 | LLM typo "complete" | `epic-1: complete` | StageImplement | "All epics complete" |
| G15 | LLM typo "wip" | `epic-1: in-progress, 1-1-x: wip` | StageImplement | "Story 1.1 being implemented" |
| G15 | LLM typo "code-review" | `epic-1: in-progress, 1-1-x: code-review` | StageTasks | "Story 1.1 in code review" |

**P2 Gap Test Cases (Story 4.6.3):**
| Gap | Test Case | Input | Expected Stage | Expected Reasoning |
|-----|-----------|-------|----------------|--------------------|
| G2 | Story drafted only | `epic-1: in-progress, 1-1-x: drafted` | StagePlan | "Story 1.1 drafted, awaiting approval" |
| G3 | Story ready-for-dev only | `epic-1: in-progress, 1-1-x: ready-for-dev` | StagePlan | "Story 1.1 ready for development" |
| G8 | Epic backlog, story in-progress | `epic-1: backlog, 1-1-x: in-progress` | StageSpecify | "Epic backlog but Story 1.1 active" |
| G8 | Epic backlog, story done | `epic-1: backlog, 1-1-x: done` | StageSpecify | "Epic backlog but Story 1.1 done" |
| G14 | Orphan story (no matching epic) | `epic-1: in-progress, 2-1-x: in-progress` | StageImplement | (warn about orphan story 2.1) |
| G17 | Synonym "completed" | `epic-1: completed` | StageImplement | "All epics complete" |
| G19 | Multiple stories, order test | `epic-1: in-progress, 1-2-x: in-progress, 1-1-x: in-progress` | StageImplement | "Story 1.1 being implemented" (first by sorted key) |
| G22 | Empty status value | `epic-1: in-progress, 1-1-x: ""` | StagePlan | (warn about empty status) |

**Inconsistent State Test Cases:**
| Gap | Test Case | Input | Expected Stage | Expected Reasoning |
|-----|-----------|-------|----------------|--------------------|
| G7 | Epic done with multiple active stories | `epic-1: done, 1-1-x: in-progress, 1-2-x: review` | StageTasks | "Epic done but Story 1.2 in review" |
| G8 | Epic backlog with multiple stories | `epic-1: backlog, 1-1-x: done, 1-2-x: in-progress` | StageSpecify | "Epic backlog but Story 1.2 active" |

**Normalization Test Cases:**
| Gap | Input Status | Normalized Value | Test Purpose |
|-----|--------------|------------------|--------------|
| G15 | `"in progress"` | `"in-progress"` | Space to hyphen |
| G15 | `"in_progress"` | `"in-progress"` | Underscore to hyphen |
| G15 | `"IN-PROGRESS"` | `"in-progress"` | Case normalization |
| G15 | `"wip"` | `"in-progress"` | Abbreviation |
| G15 | `"complete"` | `"done"` | Synonym |
| G15 | `"completed"` | `"done"` | Synonym |
| G15 | `"finished"` | `"done"` | Synonym |
| G15 | `"reviewing"` | `"review"` | Synonym |
| G15 | `"in-review"` | `"review"` | Synonym |
| G15 | `"code-review"` | `"review"` | Synonym |

**Error Handling Tests:**
| Test Case | Expected Stage | Expected Confidence |
|-----------|----------------|---------------------|
| Empty sprint-status.yaml | StageUnknown | ConfidenceUncertain |
| Invalid YAML syntax | StageUnknown | ConfidenceUncertain |
| Missing development_status key | StageUnknown | ConfidenceUncertain |
| File read error | (fallback to artifacts) | - |

**Context Cancellation Tests:**
| Test Case | Expected Result |
|-----------|-----------------|
| Context cancelled before read | `ctx.Err()` returned immediately |
| Context cancelled during parse | `ctx.Err()` returned promptly |
| Context with 1ns timeout | `context.DeadlineExceeded` |

### File Structure (Final)

```
internal/adapters/detectors/bmad/
├── detector.go           # MODIFY - add detectStage() call in Detect()
├── detector_test.go      # MODIFY - add stage integration tests
├── stage_parser.go       # NEW - SprintStatus struct, parseSprintStatus(), determineStageFromStatus()
└── stage_parser_test.go  # NEW - comprehensive parser tests
```

### Critical Constraints

1. **Do NOT modify `internal/core/domain/stage.go`** - Use existing Stage enum values only
2. **Context cancellation is mandatory** - Check ctx.Done() before AND after every I/O operation
3. **Graceful degradation** - Always return a result, never panic
4. **Follow Story 4.5.1 patterns exactly** - Same code style, same error handling
5. **Use `domain.NewDetectionResult()` constructor** - Never construct DetectionResult struct manually
6. **ConfidenceLikely for all fallback detection** - Only ConfidenceCertain when sprint-status.yaml parsed successfully

### Implementation Priority (Gap Classification from Story 4.6.1)

| Priority | Gap IDs | Description |
|----------|---------|-------------|
| **P1 (Must Fix)** | G1, G7, G15 | Actively misleading users OR high LLM probability |
| **P2 (Should Fix)** | G2, G3, G8, G14, G17, G19, G22 | User sees less helpful info OR data quality issues |
| **P3 (Nice to Have)** | G4, G5, G6, G9, G10, G11, G12, G13, G16, G18, G20, G21 | Edge cases, UX polish |

**Gap Details:**
| Gap ID | Scenario | Current Behavior | Expected Behavior |
|--------|----------|------------------|-------------------|
| G1 | Epic in-progress, all stories done | "preparing stories" | "Epic N complete, update epic status" |
| G2 | Story status `drafted` | Falls through | "Story X.Y drafted, awaiting approval" |
| G3 | Story status `ready-for-dev` | Falls through | "Story X.Y ready for development" |
| G4 | Multiple stories in-progress | Shows first found | Show first by sorted key |
| G5 | Multiple stories in review | Shows first found | Show first by sorted key |
| G6 | Story `done` status | Falls through | Count toward epic completion check |
| G7 | Epic done but stories active | "All epics complete" | Warning about inconsistent state |
| G8 | Epic backlog but stories active | Stories ignored | Warning about inconsistent state |
| G9 | Unknown epic status | Generic message | Include status value in message |
| G10 | Unknown story status | Falls through | Include status value in message |
| G11 | Whitespace in status | Not matched | Normalize whitespace |
| G12 | Multi-epic order sensitivity | Lexicographic sort | Semantic order |
| G13 | Uppercase epic key | Not matched | Case-insensitive match |
| G14 | Orphan story (no matching epic) | Silently ignored | Warn about orphan |
| G15 | LLM typos | Falls through | Normalize common variations |
| G16 | Sub-epic depth >2 levels | Not matched | Support or warn |
| G17 | Status synonyms | Falls through | Map to canonical |
| G18 | Deep story prefix (4-5-6-xxx) | May not match | Handle or warn |
| G19 | Story order within epic | Last-wins | First by sorted key |
| G20 | `contexted` epic patterns | Not explicitly tested | Verify same as in-progress |
| G21 | Uppercase story keys | May not match | Case-normalize keys |
| G22 | Empty status value | Falls through | Warn about empty status |

### Decision Summary (from Story 4.6.1)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| D1: `drafted` display | Show "Story X.Y drafted, awaiting approval" | User should know which story is being refined |
| D2: `ready-for-dev` display | Plan "Story X.Y ready for development" | Not yet implementing, but work is ready |
| D3: Multi-status priority | Review > in-progress | Review is higher-urgency action |
| D4: Retrospective handling | Ignore | Retros don't block development |
| D5: Inconsistent states | Warn in reasoning string | Help catch data entry errors |
| D6: Empty epic | "preparing stories" | Correct - genuinely preparing |
| D7: All stories done | "Epic N stories complete, update epic status" | Guide user to update sprint-status.yaml |

### References

- [Source: internal/adapters/detectors/bmad/detector.go] - Story 4.5.1 implementation to extend
- [Source: internal/adapters/detectors/bmad/detector_test.go] - Test patterns to follow
- [Source: internal/core/domain/stage.go] - Stage enum (DO NOT MODIFY)
- [Source: docs/sprint-artifacts/sprint-status.yaml] - Real-world example to parse
- [Source: docs/project-context.md] - Testing and architecture rules

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- All tests passing: `go test ./internal/adapters/detectors/bmad/... -v`
- Lint passing for bmad package: `golangci-lint run ./internal/adapters/detectors/bmad/...`

### Completion Notes List

- Implemented `stage_parser.go` with SprintStatus struct, parseSprintStatus(), determineStageFromStatus(), detectStageFromArtifacts(), findSprintStatusPath(), and detectStage() methods
- Modified `detector.go` Detect() method to integrate stage detection after version extraction
- Updated `detector_test.go` to expect new combined reasoning format
- Created comprehensive `stage_parser_test.go` with 30+ test cases covering:
  - Sprint status parsing (valid, empty, malformed YAML)
  - All epic status combinations (backlog, in-progress, done, mixed)
  - Story status handling (in-progress, review)
  - Fallback artifact detection (epics, architecture, PRD)
  - Context cancellation (immediate and timeout)
  - Helper function unit tests
  - Full integration tests

### File List

- `internal/adapters/detectors/bmad/stage_parser.go` (NEW - 351 lines)
- `internal/adapters/detectors/bmad/stage_parser_test.go` (NEW - 839 lines)
- `internal/adapters/detectors/bmad/detector.go` (MODIFIED - updated Detect() method)
- `internal/adapters/detectors/bmad/detector_test.go` (MODIFIED - updated expected values)
