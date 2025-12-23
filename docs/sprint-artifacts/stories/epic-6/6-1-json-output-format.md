# Story 6.1: JSON Output Format

Status: done

## Story

As a **scripter**,
I want **JSON output with API versioning**,
So that **my scripts are stable across updates**.

## Acceptance Criteria

1. **AC1: JSON flag on list command**
   - Given I run `vibe list --json`
   - Then output is valid JSON with `api_version` and `projects` array
   - And all field names use snake_case per Architecture spec

2. **AC2: API version field**
   - Given JSON output is requested
   - Then response includes `"api_version": "v1"` at root level
   - And this version remains stable for backwards compatibility

3. **AC3: Explicit version flag**
   - Given I run `vibe list --json --api-version=v1`
   - Then explicit v1 schema is used
   - And future v2 requests return v2 format when available

4. **AC4: Complete project fields in JSON**
   - Given JSON output for a project
   - Then includes all required fields:
     - `name` (string): Directory-derived name
     - `display_name` (string|null): User-set nickname, null if not set
     - `path` (string): Canonical absolute path
     - `method` (string): "speckit", "bmad", or "unknown"
     - `stage` (string): lowercase stage name
     - `confidence` (string): "certain", "likely", or "uncertain"
     - `state` (string): "active" or "hibernated"
     - `is_favorite` (bool): Favorite status
     - `is_waiting` (bool): Agent waiting detection status
     - `waiting_duration_minutes` (int|null): Minutes waiting, null if not waiting
     - `last_activity_at` (string): ISO 8601 UTC timestamp

5. **AC5: Empty projects array**
   - Given no projects are tracked
   - Then JSON output shows: `{"api_version": "v1", "projects": []}`
   - And exit code is 0 (not an error condition)

6. **AC6: Notes field inclusion**
   - Given a project has notes set
   - Then JSON includes `notes` field with the note content
   - And if no notes, `notes` is null or omitted

7. **AC7: Detection reasoning in JSON**
   - Given JSON output with verbose mode or specific flag
   - Then includes `detection_reasoning` field with human-readable explanation
   - And this helps debugging detection issues via scripts

## Tasks / Subtasks

- [x] Task 1: Add new fields to ProjectSummary struct in `list.go`
  - [x] 1.1: Add `IsWaiting bool` and `WaitingDurationMinutes *int`
  - [x] 1.2: Add `Notes *string` and `DetectionReasoning *string`
  - [x] 1.3: Fix `Confidence` to use `p.Confidence.String()` (currently hardcoded!)

- [x] Task 2: Wire WaitingDetector access in CLI
  - [x] 2.1: Note: `waitingDetector` already exists in `add.go` (same package, accessible)
  - [x] 2.2: Verified main.go calls `cli.SetWaitingDetector()` at line 121
  - [x] 2.3: In `formatJSON()`, call `waitingDetector.IsWaiting(ctx, p)` and `WaitingDuration()`

- [x] Task 3: Add `--api-version` flag
  - [x] 3.1: Add `apiVersion` string flag, default "v1"
  - [x] 3.2: Validate in runList: if not "v1", return error "unsupported API version: %s"

- [x] Task 4: Populate new fields in formatJSON()
  - [x] 4.1: Use `strings.ToLower(p.Confidence.String())` for confidence
  - [x] 4.2: Convert `p.Notes` to `*string` (nil if empty)
  - [x] 4.3: Convert `p.DetectionReasoning` to `*string` (nil if empty)
  - [x] 4.4: Convert waiting duration to minutes int

- [x] Task 5: Unit tests
  - [x] 5.1: Test all fields present in JSON output
  - [x] 5.2: Test nullable fields are null when empty
  - [x] 5.3: Test `--api-version=v1` accepted, `v99` rejected

- [x] Task 6: Integration test
  - [x] 6.1: End-to-end JSON parsing with `json.Unmarshal`

## Dev Notes

### CRITICAL: Domain Fields Already Exist

These fields are **already stored** in `domain/project.go` - DO NOT recreate:

| Domain Field | Line | JSON Output |
|--------------|------|-------------|
| `Confidence Confidence` | 19 | `strings.ToLower(p.Confidence.String())` |
| `DetectionReasoning string` | 20 | Use directly, nullable if empty |
| `Notes string` | 23 | Use directly, nullable if empty |

### CRITICAL: WaitingDetector Already Wired

WaitingDetector is **already injected** via `cli.SetWaitingDetector()` in `main.go:121`.

The `waitingDetector` package variable is defined in `add.go:30` along with its setter `SetWaitingDetector()`.
Since `list.go` is in the same `cli` package, it can access this variable directly.

In `formatJSON()`:
```go
isWaiting := false
var waitingMinutes *int
if waitingDetector != nil {
    isWaiting = waitingDetector.IsWaiting(ctx, p)
    if isWaiting {
        mins := int(waitingDetector.WaitingDuration(ctx, p).Minutes())
        waitingMinutes = &mins
    }
}
```

### Nullable Field Pattern (MUST FOLLOW)

For optional fields, use pointer types to output `null`:
```go
type ProjectSummary struct {
    DisplayName           *string `json:"display_name"`
    WaitingDurationMinutes *int   `json:"waiting_duration_minutes"`
    Notes                 *string `json:"notes"`
    DetectionReasoning    *string `json:"detection_reasoning"`
}
```

Conversion pattern:
```go
var notes *string
if p.Notes != "" {
    notes = &p.Notes
}
```

### Architecture Compliance

**JSON/YAML Format Conventions:**
- Keys: snake_case
- Timestamps: ISO 8601 UTC (RFC3339)
- Nullable fields: Use `*type` pointers

**Exit Codes:** 0=success, 1=error, 2=not found

### Anti-Patterns to AVOID

| DON'T | DO |
|-------|-----|
| Hardcode `"uncertain"` for confidence | Use `strings.ToLower(p.Confidence.String())` |
| Add Notes/Confidence/DetectionReasoning to domain | Already exist in `domain/project.go` |
| Create new dependency injection pattern | Follow existing `add.go` + `Set*()` pattern |
| Use regular `int` for waiting_duration_minutes | Use `*int` pointer for nullable |

### File Locations

| File | Purpose |
|------|---------|
| `internal/adapters/cli/list.go` | Main implementation - add fields to ProjectSummary |
| `internal/adapters/cli/add.go` | Contains `waitingDetector` package variable + setter |
| `internal/adapters/cli/list_test.go` | Unit tests |

### Example JSON Output

```json
{
  "api_version": "v1",
  "projects": [{
    "name": "vibe-dash",
    "display_name": null,
    "path": "/Users/limjk/GitHub/JeiKeiLim/vibe-dash",
    "method": "bmad",
    "stage": "implement",
    "confidence": "certain",
    "state": "active",
    "is_favorite": true,
    "is_waiting": true,
    "waiting_duration_minutes": 45,
    "notes": "Working on Epic 6",
    "detection_reasoning": "Epic 4 in-progress",
    "last_activity_at": "2025-12-23T10:30:00Z"
  }]
}
```

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-6/6-1-json-output-format.md`
- Project context: `docs/project-context.md`

### Agent Model Used

Claude Opus 4.5

### Debug Log References

N/A - All tests passing, no debugging required.

### Completion Notes List

1. Added new fields to `ProjectSummary` struct: `IsWaiting`, `WaitingDurationMinutes`, `Notes`, `DetectionReasoning`
2. Fixed hardcoded `"uncertain"` confidence - now uses `strings.ToLower(p.Confidence.String())`
3. Added `--api-version` flag with validation (only `v1` accepted)
4. Integrated `waitingDetector` (already wired in `add.go`) into `formatJSON()`
5. All nullable fields correctly use pointer types (`*string`, `*int`) for proper JSON null output
6. Added 5 new unit tests covering all new functionality
7. End-to-end test verified with actual binary using `jq`

### File List

| File | Change |
|------|--------|
| `internal/adapters/cli/list.go` | Added new fields to `ProjectSummary`, updated `formatJSON()`, added `--api-version` flag |
| `internal/adapters/cli/list_test.go` | Added 5 new tests: `TestList_JSON_AllFieldsPresent`, `TestList_JSON_NullableFieldsWhenEmpty`, `TestList_JSON_APIVersionValidation`, `TestList_JSON_ConfidenceLevels`, `TestList_JSON_WaitingDetectorNotSet` |

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Build and Basic Check

```bash
cd ~/GitHub/JeiKeiLim/vibe-dash
make build

# Test JSON output
./bin/vibe list --json
```

**Expected:**
- Valid JSON with `api_version: "v1"`
- All projects listed with required fields
- `is_waiting` and `waiting_duration_minutes` fields present

### Step 2: Field Verification

```bash
# Parse with jq to verify structure
./bin/vibe list --json | jq '.projects[0] | keys'
```

**Expected fields:**
- `name`, `display_name`, `path`, `method`, `stage`, `confidence`
- `state`, `is_favorite`, `is_waiting`, `waiting_duration_minutes`
- `notes`, `last_activity_at`

### Step 3: API Version Flag

```bash
# Test explicit version
./bin/vibe list --json --api-version=v1

# Test invalid version (should error)
./bin/vibe list --json --api-version=v99
```

**Expected:**
- v1: Same output as without flag
- v99: Error message about unsupported version

### Step 4: Empty Projects

```bash
# Temporarily move config to test empty state (optional)
# Or verify empty array format in test
./bin/vibe list --json | jq '.projects | length'
```

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark story `done` |
| JSON structure incorrect | Do NOT approve, document issue |
| Missing fields | Do NOT approve, list missing fields |
| jq parsing fails | Do NOT approve, check JSON validity |

## Code Review Record

### Review Date
2025-12-23

### Reviewer
Amelia (Dev Agent) - Claude Opus 4.5

### Issues Found & Fixed

| ID | Severity | Description | Fix Applied |
|----|----------|-------------|-------------|
| M1 | Medium | `detection_reasoning` used `omitempty` but other nullable fields output explicit `null` - inconsistent | Removed `omitempty` from `DetectionReasoning` JSON tag |
| M2 | Medium | Story File List missing `docs/sprint-artifacts/sprint-status.yaml` | Workflow artifact, not code - noted only |
| M3 | Medium | Dev Notes incorrectly stated to add to `deps.go` but implementation uses `add.go` | Updated Dev Notes to reflect actual implementation |

### Verification

- All unit tests pass: `go test ./internal/adapters/cli/... -v` ✓
- Linting passes: `make lint` ✓
- JSON output verified: `./bin/vibe list --json` produces valid JSON with all required fields
- API version validation works: `v99` correctly rejected with error message
