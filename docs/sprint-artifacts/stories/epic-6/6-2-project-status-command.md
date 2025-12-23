# Story 6.2: Project Status Command

Status: Done

## Story

As a **scripter**,
I want **to query specific project status non-interactively**,
So that **I can check state in CI/CD pipelines and automation scripts**.

## Acceptance Criteria

1. **AC1: Single project status - plain text**
   - Given project "client-alpha" exists
   - When I run `vibe status client-alpha`
   - Then output shows:
     ```
     client-alpha
       Path:        /home/user/projects/client-alpha
       Method:      Speckit
       Stage:       Plan
       Confidence:  Certain
       State:       Active
       Favorite:    No
       Notes:       Waiting on API specs
       Last Active: 2h ago
     ```

2. **AC2: Single project status - JSON**
   - Given project "client-alpha" exists
   - When I run `vibe status client-alpha --json`
   - Then output is valid JSON with single project object:
     ```json
     {
       "api_version": "v1",
       "project": {
         "name": "client-alpha",
         "display_name": null,
         "path": "/home/user/client-alpha",
         "method": "speckit",
         "stage": "plan",
         "confidence": "certain",
         "state": "active",
         "is_favorite": false,
         "is_waiting": true,
         "waiting_duration_minutes": 45,
         "notes": "Waiting on API specs",
         "detection_reasoning": "plan.md exists",
         "last_activity_at": "2025-12-11T10:30:00Z"
       }
     }
     ```

3. **AC3: Project not found**
   - Given project "nonexistent" does not exist
   - When I run `vibe status nonexistent`
   - Then error shows: "✗ Project not found: nonexistent"
   - And exit code is 2 (ExitProjectNotFound)

4. **AC4: Lookup by display name**
   - Given project has display_name "My Cool App"
   - When I run `vibe status "My Cool App"`
   - Then project is found and status displayed

5. **AC5: Lookup by path**
   - Given project exists at /home/user/myproject
   - When I run `vibe status /home/user/myproject`
   - Then project is found by path and status displayed

6. **AC6: All projects status**
   - Given multiple projects exist
   - When I run `vibe status --all`
   - Then all projects are listed (same format as `vibe list`)
   - And exit code is 0

7. **AC7: API version flag**
   - Given I run `vibe status client-alpha --json --api-version=v1`
   - Then explicit v1 schema is used
   - When I run `vibe status client-alpha --json --api-version=v99`
   - Then error: "unsupported API version: v99"

8. **AC8: Empty project name**
   - Given I run `vibe status` (no project name, no --all)
   - Then error shows: "requires a project name or --all flag"
   - And exit code is 1

## Tasks / Subtasks

- [x] Task 1: Create status command (AC: 1, 2, 3, 8)
  - [x] 1.1: Create `internal/adapters/cli/status.go` with newStatusCmd()
  - [x] 1.2: Add flags: `--json`, `--api-version`, `--all` using package-level vars
  - [x] 1.3: Implement runStatus() with repository nil check and project lookup
  - [x] 1.4: Implement formatStatusPlainText() for single project indented output
  - [x] 1.5: Register command in init() → RootCmd.AddCommand(newStatusCmd())

- [x] Task 2: Project lookup logic (AC: 3, 4, 5)
  - [x] 2.1: Extend findProjectByName() pattern from remove.go to findProjectByIdentifier()
  - [x] 2.2: Lookup order: Name match → DisplayName match → Path match (canonicalize input)
  - [x] 2.3: Use domain.ErrProjectNotFound with SilenceErrors pattern for exit code mapping

- [x] Task 3: JSON output for single project (AC: 2, 7)
  - [x] 3.1: Create StatusResponse struct (api_version + single ProjectSummary)
  - [x] 3.2: Reuse ProjectSummary from list.go (import, don't duplicate)
  - [x] 3.3: Apply waitingDetector and nullable field patterns from Story 6.1
  - [x] 3.4: Validate API version at start of runStatus (only v1 supported)

- [x] Task 4: All projects mode (AC: 6)
  - [x] 4.1: When --all flag set, call formatPlainText(cmd, projects) from list.go
  - [x] 4.2: When --all && --json, call formatJSON(ctx, cmd, projects) from list.go
  - [x] 4.3: Return ListResponse format (projects array), not StatusResponse

- [x] Task 5: Unit tests
  - [x] 5.1: Test single project plain text output format matches indented spec
  - [x] 5.2: Test single project JSON output with all fields (verify `project` not `projects`)
  - [x] 5.3: Test project not found returns exit code 2 via MapErrorToExitCode
  - [x] 5.4: Test lookup by display_name finds project
  - [x] 5.5: Test lookup by path finds project (test with canonical path)
  - [x] 5.6: Test --all flag produces same output as `vibe list`
  - [x] 5.7: Test missing project name shows usage error with exit code 1
  - [x] 5.8: Test invalid API version rejected with error message

- [x] Task 6: Integration verification
  - [x] 6.1: Manual end-to-end verification with real binary (see User Testing Guide)

## Dev Notes

### CRITICAL: Reuse Patterns from Story 6.1 and Existing CLI Commands

Story 6.1 established JSON output patterns. `remove.go` has project lookup. **DO NOT reinvent**:

| Pattern | Location | How to Reuse |
|---------|----------|--------------|
| ProjectSummary struct | `list.go:123-137` | Import directly, same package |
| Nullable field pattern | `list.go:147-163` | Copy pattern for single project |
| waitingDetector access | `add.go:30` | Package variable, already accessible |
| API version validation | `list.go:72-74` | Same pattern: `if apiVersion != "v1"` |
| Exit code mapping | `exitcodes.go:29-44` | Use MapErrorToExitCode() |
| findProjectByName | `remove.go:69-81` | Extend for path lookup |
| SilenceErrors pattern | `remove.go:97-100` | Use for "not found" custom message |
| ResetFlags pattern | `list.go:24-28` | Same pattern for test isolation |

### StatusResponse vs ListResponse

```go
// StatusResponse - single project output (vibe status <name>)
type StatusResponse struct {
    APIVersion string         `json:"api_version"`
    Project    ProjectSummary `json:"project"` // Single object, NOT array
}

// ListResponse - already exists in list.go for --all mode
// type ListResponse struct {
//     APIVersion string           `json:"api_version"`
//     Projects   []ProjectSummary `json:"projects"` // Array
// }
```

**Key Difference:** `status <name> --json` returns `project` (singular), `status --all --json` returns `projects` (array via ListResponse).

### findProjectByIdentifier Implementation

Extend the `findProjectByName` pattern from `remove.go:69-81`:

```go
// findProjectByIdentifier finds a project by name, display_name, or path.
// Extends findProjectByName pattern from remove.go.
// Lookup order: Name → DisplayName → Path (canonicalized)
func findProjectByIdentifier(ctx context.Context, identifier string) (*domain.Project, error) {
    if repository == nil {
        return nil, fmt.Errorf("repository not initialized")
    }

    projects, err := repository.FindAll(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to find project: %w", err)
    }

    // 1. Try exact name match (highest priority)
    for _, p := range projects {
        if p.Name == identifier {
            return p, nil
        }
    }

    // 2. Try display name match
    for _, p := range projects {
        if p.DisplayName == identifier {
            return p, nil
        }
    }

    // 3. Try path match (canonicalize input, ignore errors for non-paths)
    canonicalInput, err := filesystem.CanonicalPath(identifier)
    if err == nil { // Only try path match if input resolves to a valid path
        for _, p := range projects {
            if p.Path == canonicalInput {
                return p, nil
            }
        }
    }

    return nil, fmt.Errorf("%w: %s", domain.ErrProjectNotFound, identifier)
}
```

### runStatus Implementation Structure

```go
func runStatus(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()

    // 1. Repository nil check (same as list.go:67-69)
    if repository == nil {
        return fmt.Errorf("repository not initialized")
    }

    // 2. API version validation (same as list.go:72-74)
    if statusAPIVersion != "v1" {
        return fmt.Errorf("unsupported API version: %s", statusAPIVersion)
    }

    // 3. Handle --all mode (delegate to list functions)
    if statusAll {
        projects, err := repository.FindAll(ctx)
        if err != nil {
            return fmt.Errorf("failed to list projects: %w", err)
        }
        project.SortByName(projects)

        if statusJSON {
            return formatJSON(ctx, cmd, projects) // From list.go
        }
        formatPlainText(cmd, projects) // From list.go
        return nil
    }

    // 4. Require project identifier if not --all (AC8)
    if len(args) == 0 {
        return fmt.Errorf("requires a project name or --all flag")
    }

    // 5. Find project by identifier
    identifier := args[0]
    project, err := findProjectByIdentifier(ctx, identifier)
    if err != nil {
        if errors.Is(err, domain.ErrProjectNotFound) {
            // SilenceErrors pattern from remove.go:97-100
            fmt.Fprintf(cmd.OutOrStdout(), "✗ Project not found: %s\n", identifier)
            cmd.SilenceErrors = true
        }
        return err
    }

    // 6. Output single project
    if statusJSON {
        return formatStatusJSON(ctx, cmd, project)
    }
    formatStatusPlainText(cmd, project)
    return nil
}
```

### Plain Text Format (Single Project)

Different from list's table format - use indented key-value pairs:

```go
func formatStatusPlainText(cmd *cobra.Command, p *domain.Project) {
    // First line: effective name (DisplayName if set, else Name)
    name := p.Name
    if p.DisplayName != "" {
        name = p.DisplayName
    }
    fmt.Fprintf(cmd.OutOrStdout(), "%s\n", name)

    // Indented details
    fmt.Fprintf(cmd.OutOrStdout(), "  Path:        %s\n", p.Path)
    fmt.Fprintf(cmd.OutOrStdout(), "  Method:      %s\n", strings.Title(p.DetectedMethod))
    fmt.Fprintf(cmd.OutOrStdout(), "  Stage:       %s\n", p.CurrentStage.String())
    fmt.Fprintf(cmd.OutOrStdout(), "  Confidence:  %s\n", p.Confidence.String())
    fmt.Fprintf(cmd.OutOrStdout(), "  State:       %s\n", p.State.String())

    favorite := "No"
    if p.IsFavorite {
        favorite = "Yes"
    }
    fmt.Fprintf(cmd.OutOrStdout(), "  Favorite:    %s\n", favorite)

    if p.Notes != "" {
        fmt.Fprintf(cmd.OutOrStdout(), "  Notes:       %s\n", p.Notes)
    }

    fmt.Fprintf(cmd.OutOrStdout(), "  Last Active: %s\n", timeformat.FormatRelativeTime(p.LastActivityAt))
}
```

### formatStatusJSON Implementation

```go
func formatStatusJSON(ctx context.Context, cmd *cobra.Command, p *domain.Project) error {
    // Build ProjectSummary using same nullable patterns as list.go:147-189
    var displayName *string
    if p.DisplayName != "" {
        displayName = &p.DisplayName
    }

    var notes *string
    if p.Notes != "" {
        notes = &p.Notes
    }

    var detectionReasoning *string
    if p.DetectionReasoning != "" {
        detectionReasoning = &p.DetectionReasoning
    }

    isWaiting := false
    var waitingMinutes *int
    if waitingDetector != nil {
        isWaiting = waitingDetector.IsWaiting(ctx, p)
        if isWaiting {
            mins := int(waitingDetector.WaitingDuration(ctx, p).Minutes())
            waitingMinutes = &mins
        }
    }

    response := StatusResponse{
        APIVersion: statusAPIVersion,
        Project: ProjectSummary{
            Name:                   p.Name,
            DisplayName:            displayName,
            Path:                   p.Path,
            Method:                 p.DetectedMethod,
            Stage:                  strings.ToLower(p.CurrentStage.String()),
            Confidence:             strings.ToLower(p.Confidence.String()),
            State:                  strings.ToLower(p.State.String()),
            IsFavorite:             p.IsFavorite,
            IsWaiting:              isWaiting,
            WaitingDurationMinutes: waitingMinutes,
            Notes:                  notes,
            DetectionReasoning:     detectionReasoning,
            LastActivityAt:         p.LastActivityAt.UTC().Format(time.RFC3339),
        },
    }

    encoder := json.NewEncoder(cmd.OutOrStdout())
    encoder.SetIndent("", "  ")
    return encoder.Encode(response)
}
```

### Flag Setup and Reset Pattern

```go
// Package-level flags (same pattern as list.go:18-21)
var statusJSON bool
var statusAPIVersion string
var statusAll bool

// ResetStatusFlags resets status command flags for testing.
// Same pattern as ResetListFlags() in list.go:24-28.
func ResetStatusFlags() {
    statusJSON = false
    statusAPIVersion = "v1"
    statusAll = false
}

func newStatusCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "status [project-name]",
        Short: "Show status of a tracked project",
        Long: `Show detailed status of a specific project or all projects.

Use a project name, display name, or path to identify the project.
Use --all to show all projects (same as 'vibe list').

Examples:
  vibe status client-alpha          # By name
  vibe status "My Cool App"         # By display name
  vibe status /home/user/project    # By path
  vibe status client-alpha --json   # JSON output
  vibe status --all                 # All projects`,
        Args: cobra.MaximumNArgs(1),
        RunE: runStatus,
    }

    cmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
    cmd.Flags().StringVar(&statusAPIVersion, "api-version", "v1", "API version for JSON output")
    cmd.Flags().BoolVar(&statusAll, "all", false, "Show all projects")

    return cmd
}

// RegisterStatusCommand registers the status command with the given parent command.
// Used for testing to create fresh command trees.
func RegisterStatusCommand(parent *cobra.Command) {
    parent.AddCommand(newStatusCmd())
}

func init() {
    RootCmd.AddCommand(newStatusCmd())
}
```

### Required Imports

```go
import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/spf13/cobra"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/shared/project"
    "github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)
```

### File Locations

| File | Purpose |
|------|---------|
| `internal/adapters/cli/status.go` | Main implementation |
| `internal/adapters/cli/status_test.go` | Unit tests |
| `internal/adapters/cli/list.go` | Reuse ProjectSummary, ListResponse, formatPlainText, formatJSON |
| `internal/adapters/cli/remove.go` | Reference findProjectByName pattern |
| `internal/adapters/cli/exitcodes.go` | MapErrorToExitCode |

### Exit Code Mapping (Already Implemented)

From `exitcodes.go:10-16`:
- `ExitSuccess = 0`
- `ExitGeneralError = 1`
- `ExitProjectNotFound = 2` ← AC3 requires this
- `ExitConfigInvalid = 3`
- `ExitDetectionFailed = 4`

### Anti-Patterns to AVOID

| DON'T | DO |
|-------|-----|
| Create new ProjectSummary struct | Reuse existing from `list.go:123-137` |
| Hardcode confidence values | Use `strings.ToLower(p.Confidence.String())` |
| Create separate JSON building for status | Reuse nullable field patterns from `list.go` |
| Forget repository == nil check | Add check at start of runStatus |
| Use `fmt.Sprintf` for JSON | Use `json.NewEncoder` with SetIndent |
| Duplicate findProjectByName logic | Extend existing pattern from `remove.go` |
| Show double error for not found | Use SilenceErrors pattern from `remove.go:97-100` |
| Forget ResetStatusFlags for tests | Follow pattern from `list.go:24-28` |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-6/6-2-project-status-command.md`
- Previous story: `docs/sprint-artifacts/stories/epic-6/6-1-json-output-format.md`
- Project context: `docs/project-context.md`
- Architecture (exit codes): `docs/architecture.md` (Error-to-Exit-Code Mapping)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - No debugging issues encountered

### Completion Notes List

- Implemented `vibe status` command with full functionality per AC1-AC8
- Reused existing patterns from list.go (ProjectSummary, formatJSON, formatPlainText)
- Extended findProjectByName pattern to findProjectByIdentifier (name → display_name → path lookup order)
- Single project JSON returns `project` object (not `projects` array)
- --all mode delegates to list.go functions for identical output
- Added SilenceErrors + SilenceUsage for clean "project not found" error output
- All 15 unit tests pass covering all acceptance criteria
- Integration tests verified with real binary

### File List

**New Files:**
- `internal/adapters/cli/status.go` - Main status command implementation
- `internal/adapters/cli/status_test.go` - Unit tests for status command

**Modified Files:**
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Build and Basic Check

```bash
cd ~/GitHub/JeiKeiLim/vibe-dash
make build

# Test plain text status
./bin/vibe status vibe-dash
```

**Expected:**
- Project details displayed with indented format
- Shows Path, Method, Stage, State, Last Active
- First line is effective name

### Step 2: JSON Output

```bash
# Test JSON status for single project
./bin/vibe status vibe-dash --json | jq '.'

# Verify it's a single project object, not array
./bin/vibe status vibe-dash --json | jq '.project.name'

# Verify NO projects array exists
./bin/vibe status vibe-dash --json | jq '.projects' # Should return null
```

**Expected:**
- `api_version: "v1"`
- `project` is an object (not array)
- All fields from AC2 present
- `.projects` returns null (not used for single project)

### Step 3: Project Not Found

```bash
./bin/vibe status nonexistent-project-xyz
echo "Exit code: $?"
```

**Expected:**
- Error message: "✗ Project not found: nonexistent-project-xyz"
- Exit code: 2 (NOT 1)
- No duplicate error message from Cobra

### Step 4: All Projects

```bash
./bin/vibe status --all
./bin/vibe status --all --json

# Compare with vibe list
diff <(./bin/vibe list) <(./bin/vibe status --all)
diff <(./bin/vibe list --json) <(./bin/vibe status --all --json)
```

**Expected:**
- Same output as `vibe list` (both modes)
- JSON shows `projects` array (not single `project`)

### Step 5: Lookup by Path

```bash
# Test lookup by canonical path
./bin/vibe status ~/GitHub/JeiKeiLim/vibe-dash
./bin/vibe status /Users/limjk/GitHub/JeiKeiLim/vibe-dash
```

**Expected:**
- Project found and status displayed (same as by name)

### Step 6: Missing Arguments

```bash
./bin/vibe status
echo "Exit code: $?"
```

**Expected:**
- Error: "requires a project name or --all flag"
- Exit code: 1

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark story `done` |
| JSON has `projects` instead of `project` | Do NOT approve, structure incorrect |
| Wrong exit code for not found | Do NOT approve, should be 2 |
| Missing fields in JSON | Do NOT approve, list missing fields |
| Double error message on not found | Do NOT approve, check SilenceErrors |
| --all differs from list output | Do NOT approve, should be identical |
