# Story 8.8: Clarify or Remove Confidence Display

Status: done

## Story

As a **user**,
I want **detection confidence to be meaningful or removed**,
So that **I'm not confused by unexplained labels**.

## Problem Statement

The detail panel displays "Certain", "Likely", "Uncertain" confidence labels that confuse users:
- **What it refers to** - Users don't know if it's about method detection or stage detection
- **What the values mean** - No explanation of the difference between "Certain" and "Likely"
- **What action to take** - No guidance on what to do when confidence is "Uncertain"

**Current Display (`detail_panel.go:128`):**
```
Confidence:  Uncertain
```

**User Reaction:** "What does this mean? Should I be worried?"

## Design Decision: REMOVE Confidence Display (TUI Only)

| Removed From | Kept In |
|--------------|---------|
| Detail panel (TUI) | `domain.Confidence` type |
| | SQLite schema (`confidence` column) |
| | JSON output (`vibe list --json`) |
| | CLI text output (`vibe status <project>`) |

**Rationale:** Confidence is an internal implementation detail. The "Detection" field already explains detection logic to users.

## Acceptance Criteria

1. **AC1: Confidence Removed from Detail Panel**
   - Given a project is selected
   - When I press 'd' to open detail panel
   - Then there is no "Confidence:" field displayed

2. **AC2: Detection Reasoning Preserved**
   - Given a project is selected
   - When I view the detail panel
   - Then "Detection:" field still shows the reasoning text

3. **AC3: JSON Output Unchanged**
   - Given `vibe list --json` is run
   - When I inspect the output
   - Then `"confidence"` field still appears for scripting compatibility

4. **AC4: CLI Status Unchanged**
   - Given `vibe status <project>` is run
   - When I view the output (text or JSON)
   - Then confidence information still appears for scripting compatibility

5. **AC5: Domain Type Preserved**
   - Given the detection system runs
   - When detectors return results
   - Then `domain.Confidence` is still used internally (no changes to detection logic)

6. **AC6: Database Schema Preserved**
   - Given projects are stored
   - When inspecting the database schema
   - Then `confidence` column still exists for detection result caching

## Tasks / Subtasks

- [x] Task 1: Remove Confidence Display from Detail Panel (AC: 1, 2)
  - [x] 1.1: In `internal/adapters/tui/components/detail_panel.go`, remove line 128
  - [x] 1.2: Removed `renderConfidence()` function (linter flagged unused)
  - [x] 1.3: Verified "Detection:" field still renders

- [x] Task 2: Update Tests (AC: 1)
  - [x] 2.1: Removed `{"confidence field", "Certain"}` from `TestDetailPanel_View_BasicFields`
  - [x] 2.2: Deleted entire test `TestDetailPanel_View_UncertainConfidence`
  - [x] 2.3: Deleted entire test `TestDetailPanel_View_LikelyConfidence`

- [x] Task 3: Verify Scripting Unchanged (AC: 3, 4)
  - [x] 3.1: `./bin/vibe list --json` → `"certain"` ✓
  - [x] 3.2: `./bin/vibe status vibe-dash --json` → `"certain"` ✓
  - [x] 3.3: `./bin/vibe status vibe-dash` → `Confidence:  Certain` ✓

- [x] Task 4: Run Tests and Lint
  - [x] 4.1: `make test` - all tests pass ✓
  - [x] 4.2: `make lint` - no warnings ✓

## Dev Notes

### Implementation Strategy

**Single-line removal in `detail_panel.go:128`:**

```go
// internal/adapters/tui/components/detail_panel.go

// BEFORE (line 128):
lines = append(lines, formatField("Confidence", renderConfidence(p.Confidence)))

// AFTER: Line deleted. Adjacent lines:
// Line 127: lines = append(lines, formatField("Stage", stage))
// Line 128: [DELETED]
// Line 129 (now 128): reasoning := p.DetectionReasoning
```

### Key Code Locations

| File | Action |
|------|--------|
| `internal/adapters/tui/components/detail_panel.go` | Removed confidence display line and `renderConfidence()` function |
| `internal/adapters/tui/components/detail_panel_test.go` | Removed confidence assertion + 2 confidence-only tests |
| `internal/adapters/cli/list.go` | NO CHANGE (JSON confidence field preserved) |
| `internal/adapters/cli/status.go` | NO CHANGE (CLI confidence display preserved) |

### Previous Story Learnings

**From Story 8.7 (Config Display in TUI):**
- Detail panel modifications are straightforward
- Keep test coverage for remaining functionality
- Changes are isolated to `detail_panel.go`

**From Story 3.3 (Detail Panel Component):**
- `formatField()` helper used for all label-value pairs
- All tests use `strings.Contains()` for assertions

### Architecture Compliance

- **Modify:** `internal/adapters/tui/components/detail_panel.go` - remove 1 line
- **Modify:** `internal/adapters/tui/components/detail_panel_test.go` - remove 1 assertion line + 2 entire tests
- **NO changes to:**
  - `internal/core/domain/confidence.go` (domain type preserved)
  - `internal/adapters/persistence/sqlite/` (schema preserved)
  - `internal/adapters/cli/list.go` (JSON API preserved)
  - `internal/adapters/cli/status.go` (CLI API preserved)

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Remove confidence from JSON output | Keep for API stability (`list.go`, `status.go`) |
| Remove confidence from `vibe status` text | Keep for CLI compatibility |
| "Update" tests that exist only for confidence | DELETE the entire test functions |
| Add explanatory text to TUI | Just remove the line - simpler is better |

### Testing Strategy

**Detail Panel Tests (`internal/adapters/tui/components/detail_panel_test.go`):**

```go
// TestDetailPanel_View_BasicFields - REMOVE this line from the tests slice:
{"confidence field", "Certain"},  // DELETE

// DELETE entire test functions:
// - TestDetailPanel_View_UncertainConfidence (lines 84-107)
// - TestDetailPanel_View_LikelyConfidence (lines 202-224)
```

**Verification Commands:**
```bash
# 1. Build
make build

# 2. Run TUI and open detail panel
./bin/vibe
# Press 'd' - verify NO "Confidence:" line

# 3. Verify JSON still has confidence
./bin/vibe list --json | jq '.projects[0].confidence'
# Expected: "certain", "likely", or "uncertain"

# 4. Verify CLI status still has confidence
./bin/vibe status vibe-dash
# Expected: "Confidence:  Certain" line present

# 5. Run tests
make test

# 6. Run lint
make lint
```

## User Testing Guide

**Time needed:** 2 minutes

### Step 1: Detail Panel Display

```bash
make build && ./bin/vibe
# Select any project, press 'd' for detail panel
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| "Confidence:" line | NOT present | Still visible |
| "Detection:" line | Present with reasoning | Missing |
| Other fields | All present (Path, Method, Stage, etc.) | Any missing |

### Step 2: Scripting API Unchanged

```bash
./bin/vibe list --json | jq '.projects[0].confidence'
./bin/vibe status vibe-dash
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| JSON confidence | `"certain"`, `"likely"`, or `"uncertain"` | Missing or null |
| CLI `Confidence:` line | Present in text output | Missing |

### Decision Guide

| Situation | Action |
|-----------|--------|
| Detail panel shows no Confidence + JSON/CLI have confidence | Mark `done` |
| Detail panel still shows Confidence | Check line 128 removal |
| JSON/CLI missing confidence | **BUG** - wrong files modified |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- Removed Confidence display from TUI detail panel as designed
- Story notes said "linter allows unused private functions" but golangci-lint flagged `renderConfidence()` as unused - removed function to pass lint
- All scripting APIs (JSON/CLI) verified to preserve confidence field
- All tests pass, lint passes

**Code Review Fixes Applied:**
- Removed unused `Confidence` field from 6 test struct literals (code cleanup)
- Added negative test assertion ensuring "Confidence:" never appears in TUI (regression protection)
- Updated Dev Notes table to reflect actual implementation (removed stale line numbers)

### File List

- `internal/adapters/tui/components/detail_panel.go` - Removed confidence field display (line 128) and `renderConfidence()` function
- `internal/adapters/tui/components/detail_panel_test.go` - Removed confidence field assertion and two confidence-only tests
