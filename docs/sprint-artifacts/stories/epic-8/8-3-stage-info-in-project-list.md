# Story 8.3: Stage Info in Project List Row

Status: done

## Story

As a **user glancing at the dashboard**,
I want **to see current stage (spec/epic/story) directly in the project list**,
So that **I don't need to open detail view to understand where I am**.

## Acceptance Criteria

1. **AC1: BMAD Project Stage Display**
   - Given a BMAD project with active epic/story
   - When viewing the project list
   - Then I see condensed stage info like "E8 S8.3 review" in the row

2. **AC2: Speckit Project Stage Display**
   - Given a Speckit project with detected stage
   - When viewing the project list
   - Then I see "Plan", "Specify", etc. in the row

3. **AC3: Stage Text Truncation**
   - Given stage text is longer than available space
   - When rendering the row
   - Then text is truncated with ellipsis, not wrapped

4. **AC4: Responsive Width Handling**
   - Given the stage info display
   - When terminal width changes
   - Then stage info adapts (full -> shortened -> hidden at narrow widths)

5. **AC5: Replace Existing Stage Column**
   - Given the current "Specify/Plan/Tasks/Implement" column
   - When stage info is displayed
   - Then it replaces the existing stage column (not added as new column)

6. **AC6: Unknown/No Stage**
   - Given a project with StageUnknown or no detection result
   - When viewing the project list
   - Then show the stage as empty or "-" (not "Unknown")

## Tasks / Subtasks

- [x] Task 1: Create stage info formatter function (AC: 1, 2, 5, 6)
  - [x] 1.1: Create `internal/shared/stageformat/stageformat.go` with `FormatStageInfo(project *domain.Project) string`
  - [x] 1.2: For BMAD (`DetectedMethod == "bmad"`): Parse `DetectionReasoning` to extract epic/story info
  - [x] 1.3: Format BMAD as "E{epic} S{story} {status}" e.g., "E8 S8.3 review"
  - [x] 1.4: For Speckit (`DetectedMethod == "speckit"`): Use `CurrentStage.String()` directly ("Plan", "Specify", etc.)
  - [x] 1.5: For Unknown (`DetectedMethod == "unknown"` or `CurrentStage == StageUnknown`): Return "-" (not empty, not "Unknown")
  - [x] 1.6: Create `internal/shared/stageformat/stageformat_test.go` with table-driven tests for all format variations

- [x] Task 2: Update delegate column layout (AC: 3, 4, 5)
  - [x] 2.1: Increase `colStage` from 10 to dynamic width (16-20 chars for "E8 S8.3 review")
  - [x] 2.2: Adjust `calculateNameWidth()` to account for new stage column size
  - [x] 2.3: Add width breakpoints: full (>100), short (80-100), hidden (<80)
  - [x] 2.4: Create `FormatStageInfoWithWidth(project, maxWidth)` for truncation

- [x] Task 3: Update delegate render (AC: 1, 2, 3, 5)
  - [x] 3.1: Replace `item.Project.CurrentStage.String()` call with new formatter
  - [x] 3.2: Use `styles.DimStyle` for stage info (secondary info, not primary focus)
  - [x] 3.3: Truncate with "..." if exceeds column width

- [x] Task 4: Handle responsive breakpoints (AC: 4)
  - [x] 4.1: At width >= 100: Show full stage info "E8 S8.3 review"
  - [x] 4.2: At width 80-99: Show shortened "E8 S8.3"
  - [x] 4.3: At width < 80: Hide stage column entirely
  - [x] 4.4: Update `calculateNameWidth()` to handle all breakpoints

- [x] Task 5: Write comprehensive tests (AC: all)
  - [x] 5.1: Unit tests for `FormatStageInfo()` with BMAD reasoning variations
  - [x] 5.2: Unit tests for `FormatStageInfoWithWidth()` truncation
  - [x] 5.3: Integration test for delegate render with stage info
  - [x] 5.4: Test responsive behavior at different widths

## Dev Notes

### Problem Analysis

**Current Display (delegate.go renders):**
```
  ⭐vibe-dash        ✨ Tasks      ⏸️ WAITING 1h   2h ago
```

The "Tasks" column shows generic stage (Specify/Plan/Tasks/Implement) which requires opening detail panel to see the actual epic/story being worked on. For BMAD projects, the real value is knowing "Epic 8, Story 8.3 in review" at a glance.

**Target Display:**
```
  ⭐vibe-dash        ✨ E8 S8.3 review  ⏸️ WAITING 1h   2h ago
```

**Key Insight:** BMAD projects have rich stage info in `DetectionReasoning` field. Speckit projects only have basic `CurrentStage`. Solution must handle both gracefully.

### Stage Info Source

The `DetectionReasoning` field contains rich stage information. Examples from `stage_parser.go`:

| Reasoning Pattern | Extracted Info |
|-------------------|----------------|
| `"Story 8.3.1 in code review"` | E8 S8.3.1 review |
| `"Story 4.5.2 being implemented"` | E4 S4.5.2 impl |
| `"Story 1.2 ready for development"` | E1 S1.2 ready |
| `"Epic 4.5 started, preparing stories"` | E4.5 prep |
| `"All epics complete - project done"` | Done |
| `"Retrospective for Epic 7 in progress"` | E7 retro |
| `"No epics in progress - planning phase"` | Planning |

### Parsing Strategy

The reasoning follows consistent patterns from `stage_parser.go:360-410`. Use switch on prefix patterns:

```go
// parseBMADReasoning extracts display info from DetectionReasoning.
// MUST handle all patterns from stage_parser.go.
func parseBMADReasoning(reasoning string) string {
    switch {
    case strings.HasPrefix(reasoning, "Story "):
        // "Story 8.3 in code review" -> "E8 S8.3 review"
        // "Story 4.5.2 being implemented" -> "E4 S4.5.2 impl"
        // "Story 1.2 ready for development" -> "E1 S1.2 ready"
        // "Story 1.2 drafted" -> "E1 S1.2 draft"
        // "Story 1.2 in backlog" -> "E1 S1.2 backlog"
    case strings.HasPrefix(reasoning, "Epic "):
        // "Epic 4.5 started, preparing stories" -> "E4.5 prep"
        // "Epic 4.5 stories complete" -> "E4.5 done"
    case strings.Contains(reasoning, "Retrospective for Epic"):
        // "Retrospective for Epic 7 in progress" -> "E7 retro"
    case strings.Contains(reasoning, "All epics complete"):
        return "Done"
    case strings.Contains(reasoning, "planning phase"):
        return "Planning"
    default:
        return ""  // Unknown patterns fall back to CurrentStage.String()
    }
}
```

**Status abbreviations (MUST match consistently):**
| Full Status | Abbreviation |
|-------------|--------------|
| in code review | review |
| being implemented | impl |
| ready for development | ready |
| drafted | draft |
| in backlog | backlog |
| started, preparing | prep |
| stories complete | done |
| Retrospective | retro |

### Column Layout Changes

**Current columns (delegate.go:18-26):**
```go
colSelection = 2   // "> " or "  "
colFavorite  = 2   // styled "⭐" or "  " (Story 3.8)
colNameMin   = 15  // Minimum name width
colIndicator = 3   // "✨ " or "⚡ " or "   "
colStage     = 10  // "Implement" is longest  <-- CHANGE THIS
colWaiting   = 14  // "⏸️ WAITING Xh" or empty
colTime      = 8   // "2w ago" max
```

**New layout - REPLACE colStage constant:**
```go
colStage     = 16  // "E8 S8.3 review" needs 14 chars, 16 for padding
// Note: "E10 S10.10 ready" = 16 chars (max realistic case)
```

**IMPORTANT:** Do NOT add a new column. REPLACE the stage column rendering logic only.

### Width Breakpoints

| Width | Stage Display | Example |
|-------|--------------|---------|
| >= 100 | Full | "E8 S8.3 review" |
| 80-99 | Short | "E8 S8.3" |
| < 80 | Hidden | (column removed) |

### Architecture Compliance

- **New file:** `internal/shared/stageformat/stageformat.go` (follows shared package pattern from Story 7.10, 7.11)
- **New file:** `internal/shared/stageformat/stageformat_test.go` (co-located tests)
- **New file:** `internal/shared/stageformat/doc.go` (package documentation)
- **Modify:** `internal/adapters/tui/components/delegate.go` (column layout and render)
- No core changes - this is presentation layer only

### Required Imports in stageformat Package

```go
import (
    "regexp"
    "strings"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)
```

**DO NOT import from adapters or TUI packages - shared packages must be dependency-free.**

### Previous Story Learnings

**From Story 8.1 & 8.2:**
- Code review will catch issues - be thorough with tests
- Integration tests important for UI behavior
- Use existing patterns from Story 7.10 (shared packages)
- 5-second tick interval means stage info will update quickly

**From Story 7.13 (numeric sorting):**
- `formatStoryKey()` in `stage_parser.go:559-570` already extracts numeric parts
- Reuse pattern: `strings.Split(key, "-")` + `isNumeric()` check
- This function is NOT exported - you CANNOT import it. Reimplement the logic locally.

**From Story 7.11 (shared styles):**
- Created `internal/shared/styles/` package pattern
- Include `doc.go` with package documentation
- Export all public symbols with clear naming

**From Story 4.5 (waiting indicator):**
- Added `WaitingChecker` and `WaitingDurationGetter` callback pattern to delegate
- Stage formatter follows same pattern - function-based, not interface

### Key Code Locations

| Component | File | Line | Notes |
|-----------|------|------|-------|
| Column constants | delegate.go | 18-26 | Change colStage from 10 to 16 |
| calculateNameWidth | delegate.go | 106-120 | Update spacing calculation for new colStage |
| renderRow stage section | delegate.go | 164-168 | Replace `CurrentStage.String()` with stageformat call |
| Stage enum String() | stage.go | 17-30 | Fallback for non-BMAD (Speckit uses this) |
| BMAD reasoning patterns | stage_parser.go | 367-409 | Reference for all DetectionReasoning formats |
| formatStoryKey (unexported) | stage_parser.go | 559-570 | Pattern only - reimplement locally |
| DimStyle | styles.go | 92-93 | Use for stage info (secondary info) |
| Project.DetectedMethod | project.go | 17 | Check "bmad" vs "speckit" vs "unknown" |
| Project.DetectionReasoning | project.go | 20 | Source for BMAD parsing |

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Parse reasoning with fragile regex | Use switch on prefix patterns with strings.HasPrefix/Contains |
| Add new column (more clutter) | Replace existing stage column render only |
| Show "Unknown" for no stage | Return "-" (dash) for unknown/empty |
| Import from detectors package | Reimplement parsing in shared package (avoids circular deps) |
| Use complex regex for story numbers | Use simple string split on "." and " " |
| Forget to handle empty DetectionReasoning | Fall back to CurrentStage.String() |
| Skip tests for edge cases | Test: empty reasoning, malformed reasoning, very long epic/story numbers |
| Import adapters from shared | Shared packages ONLY import from core/domain |

### Essential Test Cases for stageformat_test.go

```go
var formatStageInfoTests = []struct {
    name     string
    project  domain.Project
    expected string
}{
    // BMAD reasoning patterns
    {"bmad story review", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "Story 8.3 in code review"}, "E8 S8.3 review"},
    {"bmad story impl", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "Story 4.5.2 being implemented"}, "E4 S4.5.2 impl"},
    {"bmad story ready", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "Story 1.2 ready for development"}, "E1 S1.2 ready"},
    {"bmad story drafted", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "Story 1.2 drafted, needs review"}, "E1 S1.2 draft"},
    {"bmad story backlog", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "Story 1.2 in backlog, needs drafting"}, "E1 S1.2 backlog"},
    {"bmad epic prep", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "Epic 4.5 started, preparing stories"}, "E4.5 prep"},
    {"bmad epic done", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "Epic 4.5 stories complete, update epic status"}, "E4.5 done"},
    {"bmad retro", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "Retrospective for Epic 7 in progress"}, "E7 retro"},
    {"bmad all done", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "All epics complete - project done"}, "Done"},
    {"bmad planning", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "No epics in progress - planning phase"}, "Planning"},
    {"bmad empty reasoning", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "", CurrentStage: domain.StagePlan}, "Plan"},
    {"bmad unknown pattern", domain.Project{DetectedMethod: "bmad", DetectionReasoning: "Something unexpected", CurrentStage: domain.StageImplement}, "Implement"},

    // Speckit - uses CurrentStage.String() directly
    {"speckit specify", domain.Project{DetectedMethod: "speckit", CurrentStage: domain.StageSpecify}, "Specify"},
    {"speckit plan", domain.Project{DetectedMethod: "speckit", CurrentStage: domain.StagePlan}, "Plan"},
    {"speckit tasks", domain.Project{DetectedMethod: "speckit", CurrentStage: domain.StageTasks}, "Tasks"},
    {"speckit implement", domain.Project{DetectedMethod: "speckit", CurrentStage: domain.StageImplement}, "Implement"},

    // Unknown method - return "-"
    {"unknown method", domain.Project{DetectedMethod: "unknown"}, "-"},
    {"empty method", domain.Project{DetectedMethod: ""}, "-"},
    {"unknown stage", domain.Project{DetectedMethod: "speckit", CurrentStage: domain.StageUnknown}, "-"},
}
```

### Delegate Modification Pattern

In `delegate.go:renderRow()`, replace lines 164-168:

```go
// BEFORE (current):
stage := item.Project.CurrentStage.String()
stageStr := fmt.Sprintf("%-*s", colStage, stage)
sb.WriteString(stageStr)

// AFTER (new):
stage := stageformat.FormatStageInfo(&item.Project)
stageStr := fmt.Sprintf("%-*s", colStage, stage)
sb.WriteString(styles.DimStyle.Render(stageStr))
```

Add import: `"github.com/JeiKeiLim/vibe-dash/internal/shared/stageformat"`

### Complete stageformat.go Skeleton

```go
// Package stageformat formats stage information for display in TUI.
package stageformat

import (
    "strings"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// FormatStageInfo returns condensed stage info for project list display.
// For BMAD: Parses DetectionReasoning -> "E8 S8.3 review"
// For Speckit: Returns CurrentStage.String() -> "Plan", "Specify", etc.
// For Unknown: Returns "-"
func FormatStageInfo(p *domain.Project) string {
    // Handle unknown/empty method first
    if p.DetectedMethod == "" || p.DetectedMethod == "unknown" {
        return "-"
    }

    // Handle unknown stage
    if p.CurrentStage == domain.StageUnknown {
        return "-"
    }

    // BMAD: Parse rich reasoning
    if p.DetectedMethod == "bmad" {
        if result := parseBMADReasoning(p.DetectionReasoning); result != "" {
            return result
        }
        // Fallback to CurrentStage for unknown patterns
        return p.CurrentStage.String()
    }

    // Speckit and others: Use CurrentStage directly
    return p.CurrentStage.String()
}

// FormatStageInfoWithWidth returns stage info truncated to maxWidth.
// Adds "..." if truncated.
func FormatStageInfoWithWidth(p *domain.Project, maxWidth int) string {
    info := FormatStageInfo(p)
    if len(info) <= maxWidth {
        return info
    }
    if maxWidth <= 3 {
        return info[:maxWidth]
    }
    return info[:maxWidth-3] + "..."
}

// parseBMADReasoning extracts display info from DetectionReasoning.
func parseBMADReasoning(reasoning string) string {
    if reasoning == "" {
        return ""
    }

    switch {
    case strings.HasPrefix(reasoning, "Story "):
        return parseStoryReasoning(reasoning)
    case strings.HasPrefix(reasoning, "Epic "):
        return parseEpicReasoning(reasoning)
    case strings.Contains(reasoning, "Retrospective for Epic"):
        return parseRetroReasoning(reasoning)
    case strings.Contains(reasoning, "All epics complete"):
        return "Done"
    case strings.Contains(reasoning, "planning phase"):
        return "Planning"
    default:
        return ""
    }
}

// parseStoryReasoning handles "Story X.Y.Z status" patterns.
// "Story 8.3 in code review" -> "E8 S8.3 review"
func parseStoryReasoning(reasoning string) string {
    // Extract story number after "Story "
    // Format: "Story 8.3 <status description>"
    // TODO: Implement parsing logic
    return ""
}

// parseEpicReasoning handles "Epic X.Y status" patterns.
func parseEpicReasoning(reasoning string) string {
    // TODO: Implement
    return ""
}

// parseRetroReasoning handles retrospective patterns.
func parseRetroReasoning(reasoning string) string {
    // TODO: Implement
    return ""
}
```

### References

| Document | Section | Relevance |
|----------|---------|-----------|
| docs/project-context.md | Hexagonal Architecture | Shared package pattern |
| docs/architecture.md | Implementation Patterns | Naming conventions |
| internal/adapters/detectors/bmad/stage_parser.go | 367-409 | DetectionReasoning format patterns |
| internal/shared/styles/styles.go | DimStyle (92-93) | Style for secondary info |
| internal/adapters/tui/components/delegate.go | 164-168 | Stage render location |

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Basic Stage Display Check

```bash
make build && ./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| BMAD project row | Shows "E8 S8.3 ..." format | Shows "Tasks" or "Implement" |
| Speckit project row | Shows "Plan" / "Specify" etc. | Shows gibberish |
| Unknown project row | Shows empty or "-" | Shows "Unknown" |

### Step 2: Current vibe-dash Project

```bash
# Since vibe-dash itself is being tracked, verify self-detection
./bin/vibe
# Find vibe-dash row - should show current epic/story
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| vibe-dash row | Shows active epic/story info | Shows generic stage |
| Stage info readable | Can understand at a glance | Too much info cluttered |

### Step 3: Responsive Width Test

```bash
# Resize terminal to different widths
./bin/vibe
```

| Width | Expected | Red Flag |
|-------|----------|----------|
| Wide (>=100) | Full stage info visible | Truncated unnecessarily |
| Medium (80-99) | Shortened stage info | Full info overflows |
| Narrow (<80) | Stage column hidden | Layout broken |

### Step 4: Detail Panel Comparison

```bash
./bin/vibe
# Select a BMAD project, press 'd' to open detail
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| List stage matches detail | Stage info consistent | Different values shown |
| Detail has more info | List is condensed version | List shows more than detail |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Stage parsing wrong | Check reasoning patterns in stage_parser.go |
| Width handling broken | Check breakpoint logic in delegate.go |
| Performance issues | Profile FormatStageInfo calls |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

1. Created `internal/shared/stageformat/` package with:
   - `doc.go`: Package documentation
   - `stageformat.go`: Core formatting functions
   - `stageformat_test.go`: Comprehensive table-driven tests

2. `FormatStageInfo()` handles all BMAD reasoning patterns from stage_parser.go:
   - Story patterns: "E8 S8.3 review", "E4 S4.5.2 impl", "E1 S1.2 ready/draft/backlog"
   - Epic patterns: "E4.5 prep", "E4.5 done"
   - Special cases: "E7 retro", "Done", "Planning"

3. `FormatStageInfoWithWidth()` handles truncation with "..." for narrow widths

4. Updated `delegate.go` with responsive breakpoints:
   - Width >= 100: Full stage info (16 chars)
   - Width 80-99: Shortened stage (10 chars)
   - Width < 80: Stage column hidden

5. Applied `styles.DimStyle` to stage info per AC requirements

6. All tests pass: `go test ./...` passes
7. Lint clean: `make lint` passes
8. Build succeeds: `make build` passes

### Code Review Applied (2025-12-26)

**Issues Fixed:**
- **M1:** Added "done" status handling to `abbreviateStatus()` for completed stories
- **M2:** Added nil pointer guard to `FormatStageInfo()` and `FormatStageInfoWithWidth()`
- **M4:** Added test case for width=4 truncation edge case
- **M5:** Added test case for "unknown status" pattern from stage_parser
- **L1:** Enhanced doc.go with usage examples
- **L2:** Added explicit empty string check in `extractEpicFromStory()`

**Tests Added:**
- `TestFormatStageInfo_NilProject` - nil pointer safety
- `TestFormatStageInfoWithWidth_NilProject` - nil pointer safety
- `TestFormatStageInfoWithWidth/width_4_with_ellipsis` - edge case
- `TestFormatStageInfo/bmad_story_done` - done status handling
- `TestFormatStageInfo/bmad_unknown_status_pattern` - unknown status pattern
- Added "done", "done and completed", "completed successfully" to TestAbbreviateStatus

### Bug Fix Applied (2025-12-26)

**Critical Bug:** Stage info not showing in TUI project list

**Root Cause:** The BMAD detector (detector.go:130) prefixes reasoning with "BMAD vX.X.X, "
but `parseBMADReasoning()` expected patterns starting directly with "Story " or "Epic ".

Example:
- Detector output: `"BMAD v6.0.0-alpha.13, Story 8.4 in backlog, needs drafting"`
- Expected input: `"Story 8.4 in backlog, needs drafting"`

**Fix:** Added `stripBMADVersionPrefix()` function to remove the version prefix before parsing.

**Tests Added:**
- `TestStripBMADVersionPrefix` - 6 test cases for prefix stripping
- `TestFormatStageInfo/bmad_with_version_prefix_*` - 4 test cases with real detector output

### File List

| File | Action | Description |
|------|--------|-------------|
| `internal/shared/stageformat/doc.go` | Created | Package documentation |
| `internal/shared/stageformat/stageformat.go` | Created | Stage info formatting with BMAD reasoning parsing |
| `internal/shared/stageformat/stageformat_test.go` | Created | Table-driven tests for all format variations |
| `internal/adapters/tui/components/delegate.go` | Modified | Added responsive stage column, uses stageformat |
| `internal/adapters/tui/components/delegate_test.go` | Modified | Added Story 8.3 tests for stage display |
