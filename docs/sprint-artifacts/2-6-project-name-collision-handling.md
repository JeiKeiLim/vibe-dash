# Story 2.6: Project Name Collision Handling

**Status:** done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Point** | Modify `internal/adapters/cli/add.go` |
| **Key Dependencies** | `ports.ProjectRepository`, `domain.Project`, existing add command infrastructure |
| **Files to Modify** | `add.go`, `add_test.go` |
| **Files to Create** | None (extends existing add command) |
| **Location** | `internal/adapters/cli/` |
| **Interfaces Used** | `ports.ProjectRepository.FindAll()` for name lookup |

### Quick Task Summary (5 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Implement name collision detection | Use FindAll() + in-memory filter for name uniqueness |
| 2 | Create disambiguation prompt UI | Interactive prompt with cmd.InOrStdin() for testability |
| 3 | Implement --force auto-disambiguation | Automatic parent-directory prefix algorithm |
| 4 | Handle edge cases | Custom name re-collision, empty input, interrupt |
| 5 | Tests + integration validation | Table-driven tests with collision scenarios |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Name lookup method | `FindAll()` + filter | MVP simplicity, avoids interface changes |
| Detection timing | After project creation, before save | Need canonical path and project name |
| Disambiguation algorithm | Parent directory prefix | PRD specifies this algorithm |
| User input | `cmd.InOrStdin()` + bufio.Scanner | Enables testing with mock stdin |
| Force flag behavior | Auto-apply parent prefix | --force means no prompts, automatic resolution |
| Collision scope | Check Name AND DisplayName | Avoid dashboard confusion |

## Story

**As a** user,
**I want** project name collisions handled automatically,
**So that** I can track multiple projects with the same directory name.

## Acceptance Criteria

```gherkin
AC1: Given project "api-service" exists at /client-a/api-service
     When I add /client-b/api-service
     Then collision is detected
     And user is prompted with options

AC2: Given collision prompt is displayed
     When user selects suggested name (option 1)
     Then project added with auto-generated name
     And success message shows the name used

AC3: Given collision prompt is displayed
     When user enters custom name (option 2)
     Then project added with that custom name

AC4: Given collision scenario
     When --force flag is used
     Then automatic disambiguation applied without prompt
     And project is added silently with generated name

AC5: Given project is displayed
     Then display_name is shown if set, otherwise name

AC6: Given symlink points to same physical location as existing project
     When adding via symlink
     Then path collision is detected (not just name)
     And user sees: "Project already tracked: <name>"

AC7: Given project name collision
     When user enters custom name that also collides
     Then collision is reported again
     And user prompted to try different name
```

## Tasks / Subtasks

- [x] **Task 1: Implement name collision detection** (AC: 1, 6)
  - [x] 1.1 Create `checkNameCollision(ctx, repo, name string) (*domain.Project, error)` function
  - [x] 1.2 Use `repo.FindAll(ctx)` and filter by Name OR DisplayName match
  - [x] 1.3 Return existing project if collision found (for prompt display)
  - [x] 1.4 Integrate into add command flow after `domain.NewProject()`, before `repo.Save()`

- [x] **Task 2: Create disambiguation prompt** (AC: 1, 2, 3, 7)
  - [x] 2.1 Add `promptCollisionResolution(cmd *cobra.Command, existing, suggested string) (string, error)`
  - [x] 2.2 Use `bufio.NewScanner(cmd.InOrStdin())` for testable input
  - [x] 2.3 Display: suggested name (option 1), custom name (option 2)
  - [x] 2.4 Loop on custom name collision until unique or user quits
  - [x] 2.5 Handle Ctrl+C/Ctrl+D gracefully (check `scanner.Err()` and `io.EOF`)

- [x] **Task 3: Implement --force auto-disambiguation** (AC: 4, 5)
  - [x] 3.1 Add `--force` flag: `cmd.Flags().BoolVar(&addForce, "force", false, "Auto-resolve collisions")`
  - [x] 3.2 Create `generateUniqueName(ctx, repo, baseName, path string) (string, error)`
  - [x] 3.3 Algorithm: prepend parent dir, check collision, repeat with grandparent if needed
  - [x] 3.4 Edge case: path at root - use timestamp suffix as fallback
  - [x] 3.5 When --force is set, skip prompt and use generated name

- [x] **Task 4: Handle edge cases** (AC: 7)
  - [x] 4.1 Custom name collision: re-prompt with error message
  - [x] 4.2 Empty input: re-prompt, don't allow
  - [x] 4.3 Whitespace-only input: trim and re-prompt
  - [x] 4.4 Very long generated name: truncate at 50 chars

- [x] **Task 5: Write tests** (AC: all)
  - [x] 5.1 Test: Name collision detected for same directory name
  - [x] 5.2 Test: Path collision (symlink to same location) - existing behavior
  - [x] 5.3 Test: Auto-disambiguation generates parent-prefixed name
  - [x] 5.4 Test: Custom name collision prompts again
  - [x] 5.5 Test: --force flag auto-resolves without prompt
  - [x] 5.6 Test: Multiple collision levels (grandparent prefix needed)
  - [x] 5.7 Test: Edge case - project at filesystem root (via timestamp fallback)
  - [x] 5.8 Extend `MockRepository` in add_test.go if needed for collision scenarios
  - [x] 5.9 Run `make build`, `make lint`, `make test`

## Dev Notes

### Collision Detection Strategy

**RECOMMENDED APPROACH:** Use `FindAll()` with in-memory filtering for MVP simplicity. This avoids interface changes to `ports.ProjectRepository`.

```go
// checkNameCollision checks if a project with the given name already exists.
// Checks both Name and DisplayName fields to prevent dashboard confusion.
func checkNameCollision(ctx context.Context, repo ports.ProjectRepository, name string) (*domain.Project, error) {
    projects, err := repo.FindAll(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to check name collision: %w", err)
    }

    for _, p := range projects {
        if p.Name == name || p.DisplayName == name {
            return p, nil // Collision found
        }
    }
    return nil, nil // No collision
}
```

### User Input with Testable Stdin

```go
// promptCollisionResolution prompts user to resolve a name collision.
// Uses cmd.InOrStdin() for testability - tests can inject mock stdin.
func promptCollisionResolution(cmd *cobra.Command, existingName, suggestedName string) (string, error) {
    fmt.Fprintf(cmd.OutOrStdout(), "Project name '%s' already exists.\n", existingName)
    fmt.Fprintf(cmd.OutOrStdout(), "Choose an option:\n")
    fmt.Fprintf(cmd.OutOrStdout(), "  1. Use suggested name: %s\n", suggestedName)
    fmt.Fprintf(cmd.OutOrStdout(), "  2. Enter a custom name\n")
    fmt.Fprintf(cmd.OutOrStdout(), "Enter choice (1/2): ")

    scanner := bufio.NewScanner(cmd.InOrStdin())
    if !scanner.Scan() {
        if err := scanner.Err(); err != nil {
            return "", err
        }
        return "", io.EOF // User pressed Ctrl+D
    }

    choice := strings.TrimSpace(scanner.Text())
    switch choice {
    case "1":
        return suggestedName, nil
    case "2":
        fmt.Fprintf(cmd.OutOrStdout(), "Enter custom name: ")
        if !scanner.Scan() {
            if err := scanner.Err(); err != nil {
                return "", err
            }
            return "", io.EOF
        }
        customName := strings.TrimSpace(scanner.Text())
        if customName == "" {
            return "", fmt.Errorf("name cannot be empty")
        }
        return customName, nil
    default:
        return "", fmt.Errorf("invalid choice: %s", choice)
    }
}
```

### Auto-Disambiguation Algorithm

```go
// generateUniqueName creates a unique name by prepending parent directories.
// Example: /home/user/clients/client-b/api-service
//   Base: api-service
//   Try 1: client-b-api-service
//   Try 2: clients-client-b-api-service (if still collides)
func generateUniqueName(ctx context.Context, repo ports.ProjectRepository, baseName, fullPath string) (string, error) {
    parts := strings.Split(filepath.Dir(fullPath), string(filepath.Separator))

    // Filter empty parts
    var validParts []string
    for _, p := range parts {
        if p != "" && p != "." {
            validParts = append(validParts, p)
        }
    }

    candidate := baseName
    prefixIdx := len(validParts) - 1

    for {
        existing, err := checkNameCollision(ctx, repo, candidate)
        if err != nil {
            return "", err
        }
        if existing == nil {
            return candidate, nil // Unique name found
        }

        if prefixIdx < 0 {
            // Ran out of path components - use timestamp suffix
            return fmt.Sprintf("%s-%d", candidate, time.Now().Unix()), nil
        }

        candidate = fmt.Sprintf("%s-%s", validParts[prefixIdx], candidate)
        prefixIdx--

        // Truncate if too long
        if len(candidate) > 50 {
            candidate = candidate[:50]
        }
    }
}
```

### --force Flag Registration

Add to `newAddCmd()` in add.go:

```go
var addForce bool

func newAddCmd() *cobra.Command {
    cmd := &cobra.Command{
        // ... existing setup ...
    }

    cmd.Flags().StringVar(&addName, "name", "", "Custom display name for the project")
    cmd.Flags().BoolVar(&addForce, "force", false, "Auto-resolve name collisions without prompting")

    return cmd
}
```

### Integration Point in add.go

Insert collision handling in `runAdd()` after project creation, before save:

```go
func runAdd(cmd *cobra.Command, args []string) error {
    // ... existing: repository check, path resolution, path collision check ...

    // Create new project
    project, err := domain.NewProject(canonicalPath, "")
    // ... existing: set display name, perform detection ...

    // === NEW: Name collision handling ===
    existing, err := checkNameCollision(cmd.Context(), repository, project.Name)
    if err != nil {
        return err
    }
    if existing != nil {
        if addForce {
            // Auto-resolve without prompt
            newName, err := generateUniqueName(cmd.Context(), repository, project.Name, canonicalPath)
            if err != nil {
                return fmt.Errorf("failed to generate unique name: %w", err)
            }
            project.DisplayName = newName
        } else {
            // Interactive prompt
            suggestedName := generateSuggestedName(project.Name, canonicalPath)
            resolvedName, err := promptCollisionResolution(cmd, project.Name, suggestedName)
            if err != nil {
                return err
            }
            // Check if custom name also collides (loop if needed)
            for {
                collision, _ := checkNameCollision(cmd.Context(), repository, resolvedName)
                if collision == nil {
                    break
                }
                fmt.Fprintf(cmd.OutOrStdout(), "Name '%s' also exists. Try another: ", resolvedName)
                // ... re-prompt logic ...
            }
            project.DisplayName = resolvedName
        }
    }
    // === END collision handling ===

    // Save to repository (existing code)
    if err := repository.Save(ctx, project); err != nil {
        return fmt.Errorf("failed to save project: %w", err)
    }
    // ... existing: success output ...
}
```

### Edge Cases

| Scenario | Expected Behavior | Domain Error |
|----------|-------------------|--------------|
| Path collision (same canonical path) | Existing behavior: "Project already tracked" | `ErrProjectAlreadyExists` |
| Name collision, user accepts suggested | Project added with suggested name | None |
| Custom name also collides | Re-prompt for different name | None (interactive) |
| Path at filesystem root `/` | Use timestamp suffix as fallback | None |
| Very long generated name | Truncate at 50 characters | None |
| Empty custom name | Re-prompt, don't allow | None (interactive) |
| User presses Ctrl+D | Return `io.EOF`, exit gracefully | None |
| --force exhausts all prefixes | Use timestamp suffix | None |

### Test Pattern for Mock Stdin

```go
func TestAdd_NameCollision_UserSelectsSuggested(t *testing.T) {
    mock := NewMockRepository()
    // Pre-populate with existing project
    existingProject, _ := domain.NewProject("/client-a/api-service", "")
    mock.projects[existingProject.Path] = existingProject
    cli.SetRepository(mock)

    // Create temp directory with same name
    tmpDir := t.TempDir()
    projectDir := filepath.Join(tmpDir, "api-service")
    os.Mkdir(projectDir, 0755)

    // Simulate user selecting option 1
    cmd := cli.NewRootCmd()
    cli.RegisterAddCommand(cmd)
    cmd.SetIn(strings.NewReader("1\n")) // User types "1" and hits enter

    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetArgs([]string{"add", projectDir})

    err := cmd.Execute()
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    // Verify project saved with disambiguated name
    // ...
}
```

### Architecture Compliance Checklist

- [ ] CLI command modification in `internal/adapters/cli/`
- [ ] NO interface changes to `ports.ProjectRepository` (uses FindAll)
- [ ] Uses domain types (`domain.Project`, `domain.ErrProjectAlreadyExists`)
- [ ] Context propagation (uses `ctx context.Context`)
- [ ] Uses `cmd.InOrStdin()` for testable input
- [ ] Follows existing CLI output patterns (✓, indentation)
- [ ] Error messages follow project conventions

### Previous Story Patterns (Story 2.5)

Apply these patterns from Story 2.5 code review:
1. **Interface abstractions** - Use ports interfaces for testability
2. **Error logging** - Use `slog.Debug` for non-critical failures
3. **Compile-time checks** - Add `var _ InterfaceType = (*ConcreteType)(nil)` if new types created
4. **Test coverage** - Include integration tests for CLI commands

### File Paths

| File | Purpose |
|------|---------|
| `internal/adapters/cli/add.go` | Add collision handling, --force flag |
| `internal/adapters/cli/add_test.go` | Add collision scenario tests |

### References

- [Source: docs/epics.md] Story 2.6 requirements (lines 793-834)
- [Source: docs/architecture.md] CLI adapter patterns
- [Source: docs/project-context.md] Go patterns, error handling
- [Source: internal/adapters/cli/add.go] Current add command implementation
- [Source: internal/adapters/cli/add_test.go] Existing test patterns, MockRepository
- [Source: docs/sprint-artifacts/2-5-detection-service.md] Previous story learnings

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 2.6 requirements)
- docs/architecture.md (Repository patterns, CLI conventions)
- docs/project-context.md (Go patterns, hexagonal rules)
- internal/adapters/cli/add.go (Current implementation)
- internal/adapters/cli/add_test.go (Test patterns, MockRepository)
- internal/core/ports/repository.go (ProjectRepository interface)
- docs/sprint-artifacts/2-5-detection-service.md (Previous story learnings)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes

**Completed:** 2025-12-13

**Implementation Summary:**
- Added `checkNameCollision()` function that uses `FindAll()` with in-memory filtering
- Added `generateUniqueName()` function with parent-directory prefix algorithm
- Added `promptCollisionResolution()` and `promptCustomName()` functions using `cmd.InOrStdin()` for testable input
- Added `--force` flag for auto-disambiguation without prompting
- Integrated collision handling into `runAdd()` flow after project creation, before save
- All ACs covered by comprehensive tests

**Tests Added:**
- `TestAdd_NameCollision_Detected` - Detects collision for same directory name
- `TestAdd_NameCollision_UserSelectsSuggested` - User picks suggested name
- `TestAdd_NameCollision_UserEntersCustomName` - User enters custom name
- `TestAdd_NameCollision_ForceFlag` - --force auto-resolves without prompting
- `TestAdd_NameCollision_CustomNameAlsoCollides` - Re-prompts when custom name collides
- `TestAdd_NameCollision_EmptyCustomName` - Rejects empty input
- `TestAdd_NameCollision_MultipleCollisionLevels` - Uses grandparent prefix when needed
- `TestAdd_PathCollision_SymlinkToSameLocation` - Path collision takes precedence
- `TestAdd_NoCollision_DifferentNames` - No collision for different names
- `TestAdd_DisplayNameCollision` - Checks both Name and DisplayName fields
- `TestAdd_AC5_DisplayNameShownInOutput` - Verifies AC5 display logic (Code Review)
- `TestAdd_InvalidChoiceReprompts` - Invalid choice re-prompts user (Code Review)

### File List

| File | Operation |
|------|-----------|
| `internal/adapters/cli/add.go` | Modified (add collision handling, --force flag, code review fixes) |
| `internal/adapters/cli/add_test.go` | Modified (add 12 collision tests including code review additions) |
| `docs/sprint-artifacts/2-6-project-name-collision-handling.md` | Modified (status update) |
| `docs/sprint-artifacts/sprint-status.yaml` | Modified (story status tracking) |

## Change Log

| Date | Change |
|------|--------|
| 2025-12-13 | Story created with ready-for-dev status by SM Agent (Bob) |
| 2025-12-13 | **Validation improvements applied:** (1) Clarified FindAll() approach over FindByName interface change. (2) Added explicit --force flag registration code. (3) Added complete promptCollisionResolution() with cmd.InOrStdin() pattern. (4) Added checkNameCollision() checking both Name and DisplayName. (5) Added graceful interrupt handling with io.EOF. (6) Added Domain Error column to edge cases table. (7) Replaced fragile line numbers with function flow references. (8) Condensed verbose sections for token efficiency. (9) Added test pattern for mock stdin. (10) Added MockRepository extension note. |
| 2025-12-13 | **Code Review Fixes Applied (Dev Agent - Amelia):** (1) **H1 FIXED:** generateUniqueName algorithm bug - changed `baseName` to `candidate` in progressive concatenation for correct grandparent prefixing. (2) **H2 FIXED:** Truncation duplicate cycle detection - added tracking of truncated candidate to detect and break infinite loops with timestamp fallback. (3) **M1 FIXED:** Added `TestAdd_AC5_DisplayNameShownInOutput` test verifying AC5 display logic. (4) **M2 FIXED:** Invalid choice now re-prompts instead of returning error - added loop with user feedback. (5) **L1 FIXED:** Reordered `ctx context.Context` to first parameter in `promptCollisionResolution` and `promptCustomName` per project conventions. (6) Added `TestAdd_InvalidChoiceReprompts` test. (7) Updated File List to include sprint-status.yaml. |
