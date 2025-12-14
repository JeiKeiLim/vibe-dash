# Story 2.7: List Projects Command

**Status:** dev-complete

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | Create `internal/adapters/cli/list.go` |
| **Key Dependencies** | `ports.ProjectRepository`, `domain.Project`, existing CLI infrastructure |
| **Files to Create** | `list.go`, `list_test.go` |
| **Files to Modify** | None |
| **Location** | `internal/adapters/cli/` |
| **Interfaces Used** | `ports.ProjectRepository.FindAll()` |

### Quick Task Summary (5 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Implement basic list command | Plain text output with columns: name, stage, last active |
| 2 | Add --json flag | JSON output with api_version, snake_case keys, ISO 8601 timestamps |
| 3 | Handle empty project list | Helpful message for plain text, empty array for JSON |
| 4 | Add sorting (alphabetical) | Sort both plain and JSON by effective name (DisplayName if set, else Name) |
| 5 | Tests + integration validation | Table-driven tests for all scenarios |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Data source | `FindAll()` | List all projects regardless of state (per AC) |
| Default sort | Alphabetical by name | Matches AC output format |
| Time format (plain) | Relative ("5m ago") | Human-friendly, matches Story 3.1 pattern |
| Time format (JSON) | ISO 8601 UTC | Architecture convention |
| API version | `"v1"` | Start with v1 for forward compatibility |
| Column alignment | Fixed-width padding | Readable table output |

## Story

**As a** user,
**I want** to list tracked projects from CLI,
**So that** I can see what's being tracked.

## Acceptance Criteria

```gherkin
AC1: Given projects are tracked
     When I run `vibe list`
     Then I see plain text output with columns:
       - Project name (or display_name if set)
       - Stage name
       - Relative time since last activity

AC2: Given projects are tracked
     When I run `vibe list --json`
     Then I see JSON output with:
       - api_version: "v1"
       - projects array with full project details
       - All keys in snake_case
       - Timestamps in ISO 8601 UTC format

AC3: Given no projects exist
     When I run `vibe list`
     Then plain text shows: "No projects tracked. Run 'vibe add .' to add one."
     And exit code is 0

AC4: Given no projects exist
     When I run `vibe list --json`
     Then JSON shows: {"api_version": "v1", "projects": []}
     And exit code is 0

AC5: Given projects exist with mixed states
     When I run `vibe list`
     Then ALL projects are shown (active and hibernated)
     And projects are sorted alphabetically by display name (or name)
```

## Tasks / Subtasks

- [x] **Task 1: Implement basic list command** (AC: 1)
  - [x] 1.1 Create `internal/adapters/cli/list.go` with `newListCmd()` function
  - [x] 1.2 Add package-level `var listJSON bool` for flag state
  - [x] 1.3 Add `ResetListFlags()` function for test isolation (matches Story 2.6 pattern)
  - [x] 1.4 Add `--json` flag: `cmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")`
  - [x] 1.5 Implement `runList()` that calls `repository.FindAll(ctx)`
  - [x] 1.6 Format plain text output with aligned columns using `cmd.OutOrStdout()`
  - [x] 1.7 Register command in `init()` with `RootCmd.AddCommand(newListCmd())`
  - [x] 1.8 Add `RegisterListCommand(parent *cobra.Command)` for testing (matches add.go pattern)

- [x] **Task 2: Implement JSON output** (AC: 2)
  - [x] 2.1 Create JSON response struct with `api_version` and `projects` fields
  - [x] 2.2 Create project JSON struct with all snake_case fields per Architecture spec
  - [x] 2.3 Format timestamps as ISO 8601 UTC (RFC3339 format)
  - [x] 2.4 Use `json.NewEncoder().SetIndent()` for readable output (more efficient than MarshalIndent)
  - [x] 2.5 Output to `cmd.OutOrStdout()` for testability

- [x] **Task 3: Handle empty project list** (AC: 3, 4)
  - [x] 3.1 Plain text: Show helpful message with add command hint
  - [x] 3.2 JSON: Return `{"api_version": "v1", "projects": []}`
  - [x] 3.3 Ensure exit code is 0 (not an error condition)

- [x] **Task 4: Add sorting** (AC: 5)
  - [x] 4.1 Create `sortProjects(projects []*domain.Project)` function
  - [x] 4.2 Sort alphabetically by effective name (case-insensitive)
  - [x] 4.3 Apply sorting BEFORE both plain text and JSON output

- [x] **Task 5: Write tests** (AC: all)
  - [x] 5.1 Test: Multiple projects listed in plain text with correct columns
  - [x] 5.2 Test: Projects sorted alphabetically by effective name
  - [x] 5.3 Test: JSON output has correct structure and api_version
  - [x] 5.4 Test: Empty list shows helpful message (plain text)
  - [x] 5.5 Test: Empty list returns empty array (JSON)
  - [x] 5.6 Test: DisplayName shown when set, otherwise Name
  - [x] 5.7 Test: Both active and hibernated projects included
  - [x] 5.8 Test: Repository error returns exit code 1
  - [x] 5.9 Run `make build`, `make lint`, `make test`

## Dev Notes

### List Command Structure

```go
package cli

import (
    "encoding/json"
    "fmt"
    "sort"
    "strings"
    "time"

    "github.com/spf13/cobra"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// listJSON holds the --json flag value
var listJSON bool

// ResetListFlags resets list command flags for testing.
// Call this before each test to ensure clean state.
func ResetListFlags() {
    listJSON = false
}

// newListCmd creates the list command.
func newListCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List all tracked projects",
        Long: `List all projects tracked by vibe-dash.

Shows project name, workflow stage, and time since last activity.
Use --json for machine-readable output.

Examples:
  vibe list           # Plain text output
  vibe list --json    # JSON output for scripting`,
        Args: cobra.NoArgs,
        RunE: runList,
    }

    cmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")

    return cmd
}

// RegisterListCommand registers the list command with the given parent command.
// Used for testing to create fresh command trees.
func RegisterListCommand(parent *cobra.Command) {
    parent.AddCommand(newListCmd())
}

func init() {
    RootCmd.AddCommand(newListCmd())
}
```

### Plain Text Output Format

Follow the AC example format. Column widths adjust based on content but have minimum widths:

```
PROJECT          STAGE      LAST ACTIVE
client-alpha     Plan       5m ago
client-bravo     Tasks      2h ago
client-charlie   Implement  3d ago
```

**Column specifications:**
- PROJECT: Left-aligned, max 40 chars (truncate with "..." if longer)
- STAGE: Left-aligned, 10 chars
- LAST ACTIVE: Right-aligned, 12 chars

```go
func formatPlainText(cmd *cobra.Command, projects []*domain.Project) {
    // Header
    fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10s %12s\n", "PROJECT", "STAGE", "LAST ACTIVE")

    for _, p := range projects {
        name := effectiveName(p)
        if len(name) > 40 {
            name = name[:37] + "..."
        }

        stage := p.CurrentStage.String()
        lastActive := formatRelativeTime(p.LastActivityAt)

        fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10s %12s\n", name, stage, lastActive)
    }
}

// effectiveName returns DisplayName if set, otherwise Name.
// Used for display and sorting per AC5.
func effectiveName(p *domain.Project) string {
    if p.DisplayName != "" {
        return p.DisplayName
    }
    return p.Name
}
```

### Relative Time Formatting

```go
func formatRelativeTime(t time.Time) string {
    d := time.Since(t)

    switch {
    case d < time.Minute:
        return "just now"
    case d < time.Hour:
        return fmt.Sprintf("%dm ago", int(d.Minutes()))
    case d < 24*time.Hour:
        return fmt.Sprintf("%dh ago", int(d.Hours()))
    case d < 7*24*time.Hour:
        return fmt.Sprintf("%dd ago", int(d.Hours()/24))
    default:
        return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
    }
}
```

### JSON Output Structure

Per Architecture spec (JSON/YAML Format Conventions):

```go
// ListResponse represents the JSON output structure
type ListResponse struct {
    APIVersion string           `json:"api_version"`
    Projects   []ProjectSummary `json:"projects"`
}

// ProjectSummary represents a single project in JSON output
type ProjectSummary struct {
    Name           string  `json:"name"`
    DisplayName    *string `json:"display_name"`     // null if not set
    Path           string  `json:"path"`
    Method         string  `json:"method"`
    Stage          string  `json:"stage"`            // lowercase per Architecture spec
    Confidence     string  `json:"confidence"`       // Default "uncertain" until DetectionResult stored
    State          string  `json:"state"`            // lowercase: "active" or "hibernated"
    IsFavorite     bool    `json:"is_favorite"`
    LastActivityAt string  `json:"last_activity_at"` // ISO 8601 UTC (RFC3339)
}

func formatJSON(cmd *cobra.Command, projects []*domain.Project) error {
    response := ListResponse{
        APIVersion: "v1",
        Projects:   make([]ProjectSummary, 0, len(projects)),
    }

    for _, p := range projects {
        var displayName *string
        if p.DisplayName != "" {
            displayName = &p.DisplayName
        }

        response.Projects = append(response.Projects, ProjectSummary{
            Name:           p.Name,
            DisplayName:    displayName,
            Path:           p.Path,
            Method:         p.DetectedMethod,
            Stage:          strings.ToLower(p.CurrentStage.String()),
            Confidence:     "uncertain", // Default until DetectionResult is stored
            State:          strings.ToLower(p.State.String()),
            IsFavorite:     p.IsFavorite,
            LastActivityAt: p.LastActivityAt.UTC().Format(time.RFC3339),
        })
    }

    encoder := json.NewEncoder(cmd.OutOrStdout())
    encoder.SetIndent("", "  ")
    return encoder.Encode(response)
}
```

### Sorting Logic

Sort by effective name (DisplayName if set, else Name), case-insensitive:

```go
func sortProjects(projects []*domain.Project) {
    sort.Slice(projects, func(i, j int) bool {
        nameI := effectiveName(projects[i])
        nameJ := effectiveName(projects[j])
        return strings.ToLower(nameI) < strings.ToLower(nameJ)
    })
}
```

### Empty List Handling

```go
func runList(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()

    if repository == nil {
        return fmt.Errorf("repository not initialized")
    }

    projects, err := repository.FindAll(ctx)
    if err != nil {
        return fmt.Errorf("failed to list projects: %w", err)
    }

    // Sort alphabetically
    sortProjects(projects)

    if listJSON {
        return formatJSON(cmd, projects)
    }

    // Plain text output
    if len(projects) == 0 {
        fmt.Fprintf(cmd.OutOrStdout(), "No projects tracked. Run 'vibe add .' to add one.\n")
        return nil
    }

    formatPlainText(cmd, projects)
    return nil
}
```

### Test Patterns

Follow existing test patterns from `add_test.go`. Use external test package `cli_test` to match project conventions.

**Critical test setup:**
```go
package cli_test  // External test package - matches add_test.go pattern

import (
    "bytes"
    "encoding/json"
    "strings"
    "testing"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// executeListCommand helper - matches executeAddCommand pattern
func executeListCommand(args []string) (string, error) {
    cli.ResetListFlags()
    cmd := cli.NewRootCmd()
    cli.RegisterListCommand(cmd)

    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetErr(&buf)

    fullArgs := append([]string{"list"}, args...)
    cmd.SetArgs(fullArgs)

    err := cmd.Execute()
    return buf.String(), err
}
```

**Table-driven test pattern (per project-context.md):**
```go
func TestList_PlainText(t *testing.T) {
    tests := []struct {
        name           string
        projects       []*domain.Project
        wantContains   []string
        wantNotContain []string
    }{
        {
            name: "multiple projects sorted alphabetically",
            projects: func() []*domain.Project {
                p1, _ := domain.NewProject("/path/to/bravo", "")
                p1.CurrentStage = domain.StageTasks
                p2, _ := domain.NewProject("/path/to/alpha", "")
                p2.CurrentStage = domain.StagePlan
                return []*domain.Project{p1, p2}
            }(),
            wantContains: []string{"PROJECT", "alpha", "bravo", "Plan", "Tasks"},
        },
        {
            name:         "empty shows helpful message",
            projects:     []*domain.Project{},
            wantContains: []string{"No projects tracked", "vibe add"},
        },
        {
            name: "DisplayName shown when set",
            projects: func() []*domain.Project {
                p1, _ := domain.NewProject("/path/to/dir", "")
                p1.DisplayName = "Custom Name"
                return []*domain.Project{p1}
            }(),
            wantContains:   []string{"Custom Name"},
            wantNotContain: []string{"dir"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := NewMockRepository()
            for _, p := range tt.projects {
                mock.projects[p.Path] = p
            }
            cli.SetRepository(mock)

            output, err := executeListCommand([]string{})
            if err != nil {
                t.Fatalf("expected no error, got: %v", err)
            }

            for _, want := range tt.wantContains {
                if !strings.Contains(output, want) {
                    t.Errorf("expected output to contain %q, got: %s", want, output)
                }
            }
            for _, notWant := range tt.wantNotContain {
                if strings.Contains(output, notWant) {
                    t.Errorf("expected output to NOT contain %q, got: %s", notWant, output)
                }
            }
        })
    }
}

func TestList_JSON_Structure(t *testing.T) {
    mock := NewMockRepository()
    p1, _ := domain.NewProject("/path/to/test", "")
    p1.CurrentStage = domain.StagePlan
    p1.State = domain.StateActive
    mock.projects[p1.Path] = p1
    cli.SetRepository(mock)

    output, err := executeListCommand([]string{"--json"})
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    var response struct {
        APIVersion string `json:"api_version"`
        Projects   []struct {
            Name  string `json:"name"`
            Stage string `json:"stage"`
            State string `json:"state"`
        } `json:"projects"`
    }
    if err := json.Unmarshal([]byte(output), &response); err != nil {
        t.Fatalf("invalid JSON: %v", err)
    }

    if response.APIVersion != "v1" {
        t.Errorf("expected api_version v1, got %s", response.APIVersion)
    }
    if len(response.Projects) != 1 {
        t.Errorf("expected 1 project, got %d", len(response.Projects))
    }
    // Verify lowercase stage per Architecture spec
    if response.Projects[0].Stage != "plan" {
        t.Errorf("expected lowercase stage 'plan', got %s", response.Projects[0].Stage)
    }
}

func TestList_BothActiveAndHibernated(t *testing.T) {
    mock := NewMockRepository()
    p1, _ := domain.NewProject("/path/to/active", "")
    p1.State = domain.StateActive
    mock.projects[p1.Path] = p1

    p2, _ := domain.NewProject("/path/to/hibernated", "")
    p2.State = domain.StateHibernated
    mock.projects[p2.Path] = p2

    cli.SetRepository(mock)

    output, err := executeListCommand([]string{})
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    // AC5: ALL projects shown regardless of state
    if !strings.Contains(output, "active") {
        t.Error("expected active project in output")
    }
    if !strings.Contains(output, "hibernated") {
        t.Error("expected hibernated project in output")
    }
}
```

### Architecture Compliance Checklist

- [x] CLI command in `internal/adapters/cli/`
- [x] Uses repository interface from `ports.ProjectRepository`
- [x] Uses domain types (`domain.Project`, `domain.Stage`, `domain.ProjectState`)
- [x] Context propagation (uses `cmd.Context()`)
- [x] Uses `cmd.OutOrStdout()` for testable output
- [x] Follows JSON convention: snake_case keys, ISO 8601 timestamps, lowercase enum values
- [x] Exit code 0 for success (including empty list)
- [x] Exit code 1 for repository failure (general error)
- [x] External test package (`package cli_test`) matches existing pattern
- [x] `ResetListFlags()` for test isolation
- [x] `RegisterListCommand()` for testable command registration

### Previous Story Patterns (Story 2.6)

Apply these patterns from previous stories:
1. **Package-level flags** - Use `var listJSON bool` pattern
2. **ResetFlags function** - Add `ResetListFlags()` for test isolation
3. **RegisterCommand pattern** - Add `RegisterListCommand(parent *cobra.Command)` for testing
4. **cmd.OutOrStdout()** - All output through this for testability
5. **NewMockRepository** - Reuse existing mock from add_test.go
6. **Table-driven tests** - Use `tests []struct{...}` pattern
7. **External test package** - Use `package cli_test` to match existing pattern
8. **executeCommand helper** - Create `executeListCommand()` matching `executeAddCommand()`

### File Paths

| File | Purpose |
|------|---------|
| `internal/adapters/cli/list.go` | List command implementation |
| `internal/adapters/cli/list_test.go` | List command tests |

### References

- [Source: docs/epics.md#story-2.7] Story requirements (lines 836-889)
- [Source: docs/architecture.md#json-yaml-format-conventions] JSON output conventions
- [Source: docs/project-context.md] Go patterns, error handling
- [Source: internal/adapters/cli/add.go] CLI command patterns, repository injection
- [Source: internal/adapters/cli/add_test.go] Test patterns, MockRepository
- [Source: docs/sprint-artifacts/2-6-project-name-collision-handling.md] Previous story patterns

## Dev Agent Record

### Code Review Applied

**Review Date:** 2025-12-13
**Reviewer:** Dev Agent (Amelia) - Adversarial Code Review

**Issues Found & Fixed:**
- M1: Added `TestList_JSON_SortedAlphabetically` - verifies JSON output maintains alphabetical sort order
- M2: Added `TestList_JSON_SortedByEffectiveName` - verifies JSON output sorts by DisplayName when set

**Issues Assessed & Kept:**
- M3: "uncertain" hardcoded string is correct for JSON output (lowercase per Architecture spec)
- L1: Story 3.1 reference in docs is forward reference (acceptable)
- L2: MockRepository reuse is fine (same test package `cli_test`)

**Test Count:** 14 → 16 (2 new JSON sorting tests added)

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 2.7 requirements)
- docs/architecture.md (JSON conventions, CLI patterns)
- docs/project-context.md (Go patterns, hexagonal rules)
- internal/adapters/cli/add.go (CLI command implementation pattern)
- internal/adapters/cli/add_test.go (Test patterns, MockRepository)
- internal/core/domain/project.go (Project entity structure)
- internal/core/domain/stage.go (Stage enum and String())
- internal/core/domain/state.go (ProjectState enum and String())
- internal/core/ports/repository.go (ProjectRepository.FindAll interface)
- docs/sprint-artifacts/2-6-project-name-collision-handling.md (Previous story learnings)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - No issues encountered during development.

### Completion Notes List

1. All 16 unit tests pass covering all acceptance criteria (14 original + 2 from code review)
2. Implementation follows Story 2.6 patterns (ResetListFlags, RegisterListCommand, external test package)
3. JSON output conforms to Architecture spec (snake_case, ISO 8601, lowercase enums)
4. Plain text output with aligned columns and relative time formatting
5. Sorting by effective name (DisplayName if set, else Name) case-insensitive
6. Code review added JSON sorting verification tests (M1, M2 fixes)

### File List

| File | Action | Description |
|------|--------|-------------|
| `internal/adapters/cli/list.go` | Created | List command implementation with plain/JSON output |
| `internal/adapters/cli/list_test.go` | Created | 14 unit tests covering all ACs |

## Change Log

| Date | Change |
|------|--------|
| 2025-12-13 | Story created with ready-for-dev status by SM Agent (Bob) |
| 2025-12-13 | **Validation improvements applied:** (1) Added Task 4 for sorting as separate task per AC5. (2) Expanded Task 1 to include ResetListFlags(), RegisterListCommand() matching Story 2.6 patterns. (3) Added Task 5.8 for repository error handling test. (4) Fixed List Command Structure to include `strings` import for ToLower. (5) Added ResetListFlags() and RegisterListCommand() functions to code sample. (6) Updated JSON struct comments with actionable guidance (confidence default, lowercase requirements). (7) Rewrote Test Patterns section with external test package `cli_test`, executeListCommand helper, and table-driven pattern. (8) Expanded Architecture Compliance Checklist with exit codes, test patterns. (9) Expanded Previous Story Patterns with 8 items. (10) Column width for LAST ACTIVE changed from 10 to 12 for better alignment. |
| 2025-12-13 | **Dev complete:** Implemented by Dev Agent (Amelia). Created list.go and list_test.go with 14 passing tests. All make build/lint/test pass. |
| 2025-12-13 | **Code review applied:** Added 2 JSON sorting tests (TestList_JSON_SortedAlphabetically, TestList_JSON_SortedByEffectiveName). Total tests: 16. All issues assessed and resolved. |
