# Story 2.8: Remove Project Command

**Status:** done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | Create `internal/adapters/cli/remove.go` |
| **Key Dependencies** | `ports.ProjectRepository`, `domain.Project`, `domain.ErrProjectNotFound` |
| **Files to Create** | `remove.go`, `remove_test.go` |
| **Files to Modify** | None |
| **Location** | `internal/adapters/cli/` |
| **Interfaces Used** | `ports.ProjectRepository.FindAll()`, `ports.ProjectRepository.Delete()` |

### Quick Task Summary (5 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Implement basic remove command | Find project by name/display_name, call Delete() |
| 2 | Add confirmation prompt | Interactive y/n prompt using cmd.InOrStdin() |
| 3 | Implement --force flag | Skip confirmation, remove immediately |
| 4 | Handle edge cases | Project not found, empty input, interrupt |
| 5 | Tests + integration validation | Table-driven tests for all ACs |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Project lookup | `FindAll()` + filter by name/display_name | Consistent with Story 2.6 pattern, MVP simplicity |
| Deletion key | Use project ID from found project | Repository.Delete() takes ID, not name |
| Confirmation input | `cmd.InOrStdin()` + bufio.Scanner | Enables testing with mock stdin (Story 2.6 pattern) |
| Force flag | Skip confirmation prompt | Standard CLI pattern for automation |
| Exit codes | 0 success, 2 not found, 1 general error | Per Architecture spec |

## Story

**As a** user,
**I want** to remove projects from tracking,
**So that** I can clean up my dashboard.

## Acceptance Criteria

```gherkin
AC1: Given project "client-alpha" is tracked
     When I run `vibe remove client-alpha`
     Then confirmation prompt shows:
       "Remove 'client-alpha' from tracking? [y/n]"

AC2: Given confirmation prompt is displayed
     When I confirm with 'y' or 'Y' or 'yes'
     Then project is removed from database
     And message shows: "✓ Removed: client-alpha"
     And exit code is 0

AC3: Given confirmation prompt is displayed
     When I cancel with 'n' or 'N' or 'no'
     Then project remains tracked
     And message shows: "Cancelled"
     And exit code is 0

AC4: Given project "client-alpha" is tracked
     When I run `vibe remove client-alpha --force`
     Then no confirmation prompt
     And project is removed immediately
     And message shows: "✓ Removed: client-alpha"
     And exit code is 0

AC5: Given project has display_name "My Alpha Project"
     When I run `vibe remove "My Alpha Project"`
     Then project is found by display_name
     And removal proceeds normally

AC6: Given project doesn't exist
     When I run `vibe remove nonexistent`
     Then error shows: "✗ Project not found: nonexistent"
     And exit code is 2
```

## Tasks / Subtasks

- [x] **Task 1: Implement basic remove command** (AC: 1, 5, 6)
  - [x] 1.1 Create `internal/adapters/cli/remove.go` with `newRemoveCmd()` function
  - [x] 1.2 Add package-level `var removeForce bool` for flag state
  - [x] 1.3 Add `ResetRemoveFlags()` function for test isolation (matches Story 2.6 pattern)
  - [x] 1.4 Add `--force` flag: `cmd.Flags().BoolVar(&removeForce, "force", false, "Remove without confirmation")`
  - [x] 1.5 Create `findProjectByName(ctx, repo, name string) (*domain.Project, error)` using FindAll() + filter
  - [x] 1.6 Implement `runRemove()` that finds project and calls Delete()
  - [x] 1.7 Register command in `init()` with `RootCmd.AddCommand(newRemoveCmd())`
  - [x] 1.8 Add `RegisterRemoveCommand(parent *cobra.Command)` for testing (matches add.go pattern)

- [x] **Task 2: Implement confirmation prompt** (AC: 1, 2, 3)
  - [x] 2.1 Add `promptRemovalConfirmation(cmd *cobra.Command, projectName string) (bool, error)`
  - [x] 2.2 Use `bufio.NewScanner(cmd.InOrStdin())` for testable input
  - [x] 2.3 Display: "Remove 'project-name' from tracking? [y/n]: "
  - [x] 2.4 Accept y/Y/yes/YES/Yes for confirmation
  - [x] 2.5 Accept n/N/no/NO/No for cancellation
  - [x] 2.6 Handle Ctrl+C/Ctrl+D gracefully (check `scanner.Err()` and `io.EOF`)

- [x] **Task 3: Implement --force flag** (AC: 4)
  - [x] 3.1 When --force is set, skip prompt and delete immediately
  - [x] 3.2 Output success message same as confirmed removal

- [x] **Task 4: Handle edge cases** (AC: 6)
  - [x] 4.1 Project not found: Return `domain.ErrProjectNotFound` wrapped with name
  - [x] 4.2 Empty project name argument: Show usage help
  - [x] 4.3 Invalid confirmation input: Re-prompt (don't error)
  - [x] 4.4 EOF/interrupt: Return "operation cancelled" gracefully

- [x] **Task 5: Write tests** (AC: all)
  - [x] 5.1 Test: Project found and removed after 'y' confirmation
  - [x] 5.2 Test: Project NOT removed after 'n' cancellation
  - [x] 5.3 Test: --force removes without prompt
  - [x] 5.4 Test: Project not found returns exit code 2
  - [x] 5.5 Test: Project found by DisplayName (AC5)
  - [x] 5.6 Test: Case-insensitive confirmation (Y, yes, YES)
  - [x] 5.7 Test: Invalid input re-prompts
  - [x] 5.8 Test: Missing argument shows usage
  - [x] 5.9 Test: EOF/Ctrl+D returns gracefully with "Cancelled" message
  - [x] 5.10 Run `make build`, `make lint`, `make test`

## Dev Notes

### Project Lookup Strategy

**IMPORTANT:** Use `FindAll()` with in-memory filtering to find project by name OR display_name. This matches the `checkNameCollision()` pattern from `add.go` (lines 211-223) and avoids interface changes.

**NOTE:** The `findProjectByName()` function is similar to `checkNameCollision()` in `add.go`. Consider whether to extract a shared helper or keep them separate (they serve different purposes: collision detection vs. retrieval).

```go
// findProjectByName finds a project by name or display_name.
// Searches both fields to support AC5 (remove by display name).
func findProjectByName(ctx context.Context, repo ports.ProjectRepository, name string) (*domain.Project, error) {
    projects, err := repo.FindAll(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to find project: %w", err)
    }

    for _, p := range projects {
        if p.Name == name || p.DisplayName == name {
            return p, nil
        }
    }
    return nil, fmt.Errorf("%w: %s", domain.ErrProjectNotFound, name)
}
```

### Remove Command Structure

```go
package cli

import (
    "bufio"
    "context"
    "errors"
    "fmt"
    "io"
    "strings"

    "github.com/spf13/cobra"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// removeForce holds the --force flag value
var removeForce bool

// ResetRemoveFlags resets remove command flags for testing.
// Call this before each test to ensure clean state.
func ResetRemoveFlags() {
    removeForce = false
}

// newRemoveCmd creates the remove command.
func newRemoveCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "remove <project-name>",
        Short: "Remove a project from tracking",
        Long: `Remove a project from the vibe-dash dashboard.

The project is identified by its name or display name.
By default, confirmation is required before removal.
Use --force to skip confirmation.

Examples:
  vibe remove client-alpha          # Remove with confirmation
  vibe remove client-alpha --force  # Remove immediately
  vibe remove "My Project"          # Remove by display name`,
        Args: cobra.ExactArgs(1),
        RunE: runRemove,
    }

    cmd.Flags().BoolVar(&removeForce, "force", false, "Remove without confirmation")

    return cmd
}

// RegisterRemoveCommand registers the remove command with the given parent command.
// Used for testing to create fresh command trees.
func RegisterRemoveCommand(parent *cobra.Command) {
    parent.AddCommand(newRemoveCmd())
}

func init() {
    RootCmd.AddCommand(newRemoveCmd())
}
```

### Remove Command Implementation

**NOTE:** For output display name, follow the `effectiveName()` pattern from `list.go:88-93` for consistency.

```go
// runRemove implements the remove command logic.
func runRemove(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()

    if repository == nil {
        return fmt.Errorf("repository not initialized")
    }

    projectName := args[0]

    // Find project by name or display_name
    project, err := findProjectByName(ctx, repository, projectName)
    if err != nil {
        return err // ErrProjectNotFound is already wrapped with name
    }

    // Display name for output - same pattern as list.go effectiveName()
    displayName := project.Name
    if project.DisplayName != "" {
        displayName = project.DisplayName
    }

    // Confirmation unless --force
    if !removeForce {
        confirmed, err := promptRemovalConfirmation(cmd, displayName)
        if err != nil {
            if errors.Is(err, io.EOF) {
                fmt.Fprintf(cmd.OutOrStdout(), "Cancelled\n")
                return nil
            }
            return err
        }
        if !confirmed {
            fmt.Fprintf(cmd.OutOrStdout(), "Cancelled\n")
            return nil
        }
    }

    // Delete from repository
    if err := repository.Delete(ctx, project.ID); err != nil {
        return fmt.Errorf("failed to remove project: %w", err)
    }

    fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed: %s\n", displayName)
    return nil
}
```

### Confirmation Prompt

**NOTE:** Prompt format must match AC1 exactly: `"Remove 'project-name' from tracking? [y/n]"` (no colon after brackets).

```go
// promptRemovalConfirmation prompts user to confirm project removal.
// Uses cmd.InOrStdin() for testability - tests can inject mock stdin.
// Returns true for confirmation (y/yes), false for cancellation (n/no).
func promptRemovalConfirmation(cmd *cobra.Command, projectName string) (bool, error) {
    scanner := bufio.NewScanner(cmd.InOrStdin())

    for {
        fmt.Fprintf(cmd.OutOrStdout(), "Remove '%s' from tracking? [y/n] ", projectName)

        if !scanner.Scan() {
            if err := scanner.Err(); err != nil {
                return false, err
            }
            return false, io.EOF // User pressed Ctrl+D
        }

        input := strings.TrimSpace(strings.ToLower(scanner.Text()))
        switch input {
        case "y", "yes":
            return true, nil
        case "n", "no":
            return false, nil
        default:
            fmt.Fprintf(cmd.OutOrStdout(), "Please enter 'y' or 'n'.\n")
            // Loop continues to re-prompt
        }
    }
}
```

### Exit Code Mapping

The remove command uses existing exit code infrastructure from `exitcodes.go`:

| Scenario | Domain Error | Exit Code |
|----------|--------------|-----------|
| Success | None | 0 |
| Cancelled | None | 0 |
| Project not found | `domain.ErrProjectNotFound` | 2 |
| Repository error | General | 1 |

### Test Patterns

Follow existing test patterns from `add_test.go` and `list_test.go`. Use external test package `cli_test`.

```go
package cli_test

import (
    "bytes"
    "strings"
    "testing"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// executeRemoveCommand helper - matches executeAddCommand pattern
func executeRemoveCommand(t *testing.T, args []string, stdin string) (string, error) {
    t.Helper()
    cli.ResetRemoveFlags()
    cmd := cli.NewRootCmd()
    cli.RegisterRemoveCommand(cmd)

    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetErr(&buf)
    if stdin != "" {
        cmd.SetIn(strings.NewReader(stdin))
    }

    fullArgs := append([]string{"remove"}, args...)
    cmd.SetArgs(fullArgs)

    err := cmd.Execute()
    return buf.String(), err
}

func TestRemove_ConfirmY_ProjectRemoved(t *testing.T) {
    mock := NewMockRepository()
    p, _ := domain.NewProject("/path/to/client-alpha", "")
    mock.projects[p.Path] = p
    cli.SetRepository(mock)

    output, err := executeRemoveCommand(t, []string{"client-alpha"}, "y\n")
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    if !strings.Contains(output, "✓ Removed: client-alpha") {
        t.Errorf("expected success message, got: %s", output)
    }

    // Verify project was deleted
    if len(mock.projects) != 0 {
        t.Error("expected project to be deleted from repository")
    }
}

func TestRemove_ConfirmN_ProjectKept(t *testing.T) {
    mock := NewMockRepository()
    p, _ := domain.NewProject("/path/to/client-alpha", "")
    mock.projects[p.Path] = p
    cli.SetRepository(mock)

    output, err := executeRemoveCommand(t, []string{"client-alpha"}, "n\n")
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    if !strings.Contains(output, "Cancelled") {
        t.Errorf("expected cancelled message, got: %s", output)
    }

    // Verify project was NOT deleted
    if len(mock.projects) != 1 {
        t.Error("expected project to remain in repository")
    }
}

func TestRemove_ForceFlag_NoPrompt(t *testing.T) {
    mock := NewMockRepository()
    p, _ := domain.NewProject("/path/to/client-alpha", "")
    mock.projects[p.Path] = p
    cli.SetRepository(mock)

    output, err := executeRemoveCommand(t, []string{"client-alpha", "--force"}, "")
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    if !strings.Contains(output, "✓ Removed: client-alpha") {
        t.Errorf("expected success message, got: %s", output)
    }

    // Verify no confirmation prompt was shown
    if strings.Contains(output, "[y/n]") {
        t.Error("expected no confirmation prompt with --force flag")
    }

    // Verify project was deleted
    if len(mock.projects) != 0 {
        t.Error("expected project to be deleted from repository")
    }
}

func TestRemove_ProjectNotFound_ExitCode2(t *testing.T) {
    mock := NewMockRepository()
    cli.SetRepository(mock)

    output, err := executeRemoveCommand(t, []string{"nonexistent"}, "")
    if err == nil {
        t.Fatal("expected error for nonexistent project")
    }

    if !strings.Contains(err.Error(), "not found") {
        t.Errorf("expected 'not found' error, got: %v", err)
    }

    // Verify error message format
    if !strings.Contains(output, "nonexistent") || !strings.Contains(err.Error(), "nonexistent") {
        t.Errorf("expected project name in error, got: %s / %v", output, err)
    }
}

func TestRemove_ByDisplayName(t *testing.T) {
    mock := NewMockRepository()
    p, _ := domain.NewProject("/path/to/client-alpha", "")
    p.DisplayName = "My Alpha Project"
    mock.projects[p.Path] = p
    cli.SetRepository(mock)

    output, err := executeRemoveCommand(t, []string{"My Alpha Project", "--force"}, "")
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    if !strings.Contains(output, "✓ Removed: My Alpha Project") {
        t.Errorf("expected success with display name, got: %s", output)
    }
}

func TestRemove_CaseInsensitiveConfirmation(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected bool // true = removed, false = cancelled
    }{
        {"lowercase y", "y\n", true},
        {"uppercase Y", "Y\n", true},
        {"yes", "yes\n", true},
        {"YES", "YES\n", true},
        {"lowercase n", "n\n", false},
        {"uppercase N", "N\n", false},
        {"no", "no\n", false},
        {"NO", "NO\n", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := NewMockRepository()
            p, _ := domain.NewProject("/path/to/test", "")
            mock.projects[p.Path] = p
            cli.SetRepository(mock)

            output, err := executeRemoveCommand(t, []string{"test"}, tt.input)
            if err != nil {
                t.Fatalf("expected no error, got: %v", err)
            }

            if tt.expected {
                if !strings.Contains(output, "✓ Removed") {
                    t.Errorf("expected removal, got: %s", output)
                }
                if len(mock.projects) != 0 {
                    t.Error("expected project to be deleted")
                }
            } else {
                if !strings.Contains(output, "Cancelled") {
                    t.Errorf("expected cancellation, got: %s", output)
                }
                if len(mock.projects) != 1 {
                    t.Error("expected project to remain")
                }
            }
        })
    }
}

func TestRemove_InvalidInput_Reprompts(t *testing.T) {
    mock := NewMockRepository()
    p, _ := domain.NewProject("/path/to/test", "")
    mock.projects[p.Path] = p
    cli.SetRepository(mock)

    // First invalid, then valid
    output, err := executeRemoveCommand(t, []string{"test"}, "maybe\ny\n")
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    if !strings.Contains(output, "Please enter") {
        t.Errorf("expected re-prompt message, got: %s", output)
    }
    if !strings.Contains(output, "✓ Removed") {
        t.Errorf("expected eventual removal, got: %s", output)
    }
}

func TestRemove_MissingArgument(t *testing.T) {
    mock := NewMockRepository()
    cli.SetRepository(mock)

    _, err := executeRemoveCommand(t, []string{}, "")
    if err == nil {
        t.Fatal("expected error for missing argument")
    }

    // Cobra should complain about missing args
    if !strings.Contains(err.Error(), "requires") && !strings.Contains(err.Error(), "argument") {
        t.Errorf("expected argument error, got: %v", err)
    }
}

func TestRemove_EOF_GracefulCancellation(t *testing.T) {
    mock := NewMockRepository()
    p, _ := domain.NewProject("/path/to/test", "")
    mock.projects[p.Path] = p
    cli.SetRepository(mock)

    // Empty stdin simulates EOF (Ctrl+D)
    output, err := executeRemoveCommand(t, []string{"test"}, "")
    if err != nil {
        t.Fatalf("expected no error for EOF cancellation, got: %v", err)
    }

    if !strings.Contains(output, "Cancelled") {
        t.Errorf("expected 'Cancelled' message on EOF, got: %s", output)
    }

    // Verify project was NOT deleted
    if len(mock.projects) != 1 {
        t.Error("expected project to remain after EOF cancellation")
    }
}
```

### MockRepository Extension

**NOTE:** The `MockRepository` in `add_test.go` already has `Delete()` method (lines 94-102). No need to add it - just reuse the existing mock.

### Architecture Compliance Checklist

- [x] CLI command in `internal/adapters/cli/`
- [x] Uses repository interface from `ports.ProjectRepository`
- [x] Uses domain types (`domain.Project`, `domain.ErrProjectNotFound`)
- [x] Context propagation (uses `cmd.Context()`)
- [x] Uses `cmd.InOrStdin()` for testable input
- [x] Uses `cmd.OutOrStdout()` for testable output
- [x] Exit code 0 for success (including cancellation)
- [x] Exit code 2 for project not found (per Architecture spec)
- [x] External test package (`package cli_test`) matches existing pattern
- [x] `ResetRemoveFlags()` for test isolation
- [x] `RegisterRemoveCommand()` for testable command registration
- [x] Follows existing CLI output patterns (✓ for success)

### Previous Story Patterns (Story 2.6, 2.7)

Apply these patterns from previous stories:
1. **Package-level flags** - Use `var removeForce bool` pattern
2. **ResetFlags function** - Add `ResetRemoveFlags()` for test isolation
3. **RegisterCommand pattern** - Add `RegisterRemoveCommand(parent *cobra.Command)` for testing
4. **cmd.InOrStdin()** - All input through this for testability
5. **cmd.OutOrStdout()** - All output through this for testability
6. **FindAll() + filter** - Use in-memory filtering to find by name (Story 2.6)
7. **Table-driven tests** - Use `tests []struct{...}` pattern
8. **External test package** - Use `package cli_test` to match existing pattern
9. **executeCommand helper** - Create `executeRemoveCommand()` matching existing patterns

### File Paths

| File | Purpose |
|------|---------|
| `internal/adapters/cli/remove.go` | Remove command implementation |
| `internal/adapters/cli/remove_test.go` | Remove command tests |

### References

- [Source: docs/epics.md#story-2.8] Story requirements (lines 891-932)
- [Source: docs/architecture.md#error-to-exit-code-mapping] Exit code conventions
- [Source: docs/project-context.md] Go patterns, error handling
- [Source: internal/adapters/cli/add.go] CLI command patterns, confirmation prompt patterns
- [Source: internal/adapters/cli/add_test.go] Test patterns, MockRepository
- [Source: internal/adapters/cli/list.go] effectiveName() pattern
- [Source: internal/core/ports/repository.go] ProjectRepository.Delete() interface
- [Source: docs/sprint-artifacts/2-6-project-name-collision-handling.md] Prompt patterns
- [Source: docs/sprint-artifacts/2-7-list-projects-command.md] Previous story patterns

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 2.8 requirements)
- docs/architecture.md (CLI patterns, exit codes, error handling)
- docs/project-context.md (Go patterns, hexagonal rules)
- internal/adapters/cli/add.go (CLI command pattern, confirmation prompt)
- internal/adapters/cli/add_test.go (Test patterns, MockRepository)
- internal/adapters/cli/list.go (effectiveName pattern)
- internal/core/ports/repository.go (ProjectRepository.Delete interface)
- docs/sprint-artifacts/2-6-project-name-collision-handling.md (Prompt patterns)
- docs/sprint-artifacts/2-7-list-projects-command.md (Previous story learnings)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story context creation phase.

### Completion Notes List

- All 5 tasks completed following red-green-refactor cycle
- 14 tests written covering all ACs and edge cases
- Tests pass: `make build`, `make lint`, `make test` all successful
- Implementation follows established patterns from Story 2.6/2.7

### File List

| File | Action | Description |
|------|--------|-------------|
| `internal/adapters/cli/remove.go` | Created | Remove command implementation (157 lines) |
| `internal/adapters/cli/remove_test.go` | Created | Remove command tests (16 tests, 24 cases) |
| `internal/adapters/cli/add_test.go` | Modified | Added deleteErr field to MockRepository |

## Change Log

| Date | Change |
|------|--------|
| 2025-12-13 | Story created with ready-for-dev status by SM Agent (Bob) |
| 2025-12-14 | **Validation improvements applied by SM Agent (Bob):** (1) Added Test 5.9 for EOF/Ctrl+D graceful cancellation. (2) Added note about findProjectByName similarity to checkNameCollision in add.go. (3) Fixed prompt format to match AC1 exactly (no colon). (4) Clarified MockRepository.Delete() already exists in add_test.go. (5) Added effectiveName() pattern reference from list.go. (6) Added TestRemove_EOF_GracefulCancellation test case. |
| 2025-12-14 | **Implementation completed by Dev Agent (Amelia):** All 5 tasks completed. Created remove.go (147 lines) and remove_test.go (14 tests). All ACs verified. `make build`, `make lint`, `make test` pass. Status changed to dev-complete. |
| 2025-12-14 | **Code Review fixes applied by Dev Agent (Amelia):** (H1) Fixed AC6 error message format - added ✗ prefix and silenced Cobra's duplicate output. (H2) Added TestRemove_DeleteFailure with deleteErr field in MockRepository. (M1) Added cross-reference comment to findProjectByName noting similarity to checkNameCollision. (M3) Added TestRemove_MultipleInvalidInputs_Reprompts for consecutive invalid input handling. Tests: 16 functions, 24 total cases. All pass. |
