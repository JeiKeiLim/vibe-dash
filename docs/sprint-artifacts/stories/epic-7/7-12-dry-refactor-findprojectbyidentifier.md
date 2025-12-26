# Story 7.12: DRY Refactor - findProjectByIdentifier Reuse

Status: done

## Story

As a **developer maintaining vibe-dash CLI commands**,
I want **project lookup logic centralized by reusing findProjectByIdentifier()**,
So that **I don't maintain duplicate lookup code across multiple commands and can ensure consistent project resolution behavior**.

## Problem Statement

Technical debt item H5 from Epic 6 Retrospective. The `favorite.go` and `note.go` files duplicate project lookup logic that already exists in `status.go` as `findProjectByIdentifier()`.

**Current State (Verified from source):**

| File | Function Location | Duplicated Logic |
|------|-------------------|------------------|
| `status.go:81-116` | `findProjectByIdentifier()` | Complete implementation: Name → DisplayName → Path lookup order |
| `favorite.go:75-80` | Inside `runFavorite()` | Manual loop: Name/DisplayName match only, NO path resolution |
| `note.go:60-65` | Inside `runNote()` | Manual loop: Name/DisplayName match only, NO path resolution |

**Impact:**
- **Inconsistent behavior:** `vibe favorite /path/to/project` fails, but `vibe status /path/to/project` works
- **Maintenance burden:** Any change to lookup logic requires 3-file updates
- **~26 lines of duplicate code** that can be eliminated
- **DRY violation** noted in Epic 6 retrospective action item H5

## Acceptance Criteria

1. **AC1: favorite.go Uses findProjectByIdentifier**
   - `runFavorite()` calls `findProjectByIdentifier(ctx, identifier)` instead of manual loop
   - Removes duplicate lookup code (lines ~68-89)

2. **AC2: note.go Uses findProjectByIdentifier**
   - `runNote()` calls `findProjectByIdentifier(ctx, identifier)` instead of manual loop
   - Removes duplicate lookup code (lines ~53-74)

3. **AC3: Path-Based Lookup Now Works for favorite/note**
   - `vibe favorite /home/user/my-project` succeeds (previously failed)
   - `vibe note /home/user/my-project "note text"` succeeds (previously failed)

4. **AC4: Error Handling Preserved**
   - When project not found: `cmd.SilenceErrors = true` and `cmd.SilenceUsage = true` are set
   - `ErrProjectNotFound` is returned with identifier in message

5. **AC5: All Existing Tests Pass**
   - `go test ./internal/adapters/cli/...` passes
   - No regressions in `favorite_test.go` and `note_test.go`

6. **AC6: New Path-Based Tests Added**
   - At least one test per command verifies path-based project lookup

## Tasks / Subtasks

- [x] Task 1: Update favorite.go to use findProjectByIdentifier (AC: 1, 3, 4)
  - [x] 1.1: Replace lines 68-89 with `findProjectByIdentifier` call
  - [x] 1.2: Rename `projectName` to `identifier` for semantic clarity
  - [x] 1.3: Preserve error handling pattern
  - [x] 1.4: Run `go test ./internal/adapters/cli/... -run TestFavorite -v`

- [x] Task 2: Update note.go to use findProjectByIdentifier (AC: 2, 3, 4)
  - [x] 2.1: Replace lines 53-74 with `findProjectByIdentifier` call
  - [x] 2.2: Rename `projectName` to `identifier` for semantic clarity
  - [x] 2.3: Preserve error handling pattern
  - [x] 2.4: Run `go test ./internal/adapters/cli/... -run TestNote -v`

- [x] Task 3: Add path-based lookup tests (AC: 5, 6)
  - [x] 3.1: Add `TestFavoriteCmd_FindByPath` in `favorite_test.go`
  - [x] 3.2: Add `TestNoteCmd_FindByPath` and `TestNoteCmd_ViewByPath` in `note_test.go`
  - [x] 3.3: Run full test suite `go test ./...`

- [x] Task 4: Final verification (AC: 5)
  - [x] 4.1: Run `golangci-lint run`
  - [x] 4.2: Run `make build && ./bin/vibe` - verify TUI works
  - [x] 4.3: Manual test: `./bin/vibe favorite <path>` and `./bin/vibe note <path> "test"`

## Dev Notes

### Key Facts

1. **Same Package Access:** `findProjectByIdentifier` is in `status.go` within the `cli` package. Since `favorite.go` and `note.go` are in the same package, they can call it directly - no imports needed.

2. **Path Resolution Included:** The function uses `filesystem.CanonicalPath()` internally, enabling path-based lookups. This is the new capability being added to favorite/note commands.

3. **Lookup Priority:** Name → DisplayName → Path (canonicalized)

### Refactoring Pattern

**Before (favorite.go:68-89 / note.go:53-74):**
```go
projectName := args[0]

// Find project by name or display name
projects, err := repository.FindAll(ctx)
if err != nil {
    return fmt.Errorf("failed to load projects: %w", err)
}

var targetProject *domain.Project
for _, p := range projects {
    if p.Name == projectName || p.DisplayName == projectName {
        targetProject = p
        break
    }
}

if targetProject == nil {
    err := fmt.Errorf("%w: %s", domain.ErrProjectNotFound, projectName)
    if errors.Is(err, domain.ErrProjectNotFound) {
        cmd.SilenceErrors = true
        cmd.SilenceUsage = true
    }
    return err
}
```

**After:**
```go
identifier := args[0]

targetProject, err := findProjectByIdentifier(ctx, identifier)
if err != nil {
    if errors.Is(err, domain.ErrProjectNotFound) {
        cmd.SilenceErrors = true
        cmd.SilenceUsage = true
    }
    return err
}
```

### Files to Modify

| File | Action | Estimate |
|------|--------|----------|
| `internal/adapters/cli/favorite.go` | Replace lines 68-89 with refactored pattern | -15 lines |
| `internal/adapters/cli/note.go` | Replace lines 53-74 with refactored pattern | -15 lines |
| `internal/adapters/cli/favorite_test.go` | Add `TestFavorite_ByPath` | +20 lines |
| `internal/adapters/cli/note_test.go` | Add `TestNote_ByPath` | +20 lines |

### Test Pattern

Use existing mock repository pattern from test files:

```go
func TestFavorite_ByPath(t *testing.T) {
    projectPath := "/home/user/my-project"
    projects := []*domain.Project{
        {ID: "1", Path: projectPath, Name: "my-project", IsFavorite: false},
    }
    cli.SetRepository(newFavoriteMockRepository().withProjects(projects))

    output, err := executeFavoriteCommand([]string{projectPath})

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !strings.Contains(output, "⭐ Favorited") {
        t.Errorf("expected '⭐ Favorited', got: %s", output)
    }
}

func TestNote_ByPath(t *testing.T) {
    projectPath := "/home/user/my-project"
    projects := []*domain.Project{
        {ID: "1", Path: projectPath, Name: "my-project", Notes: ""},
    }
    cli.SetRepository(newNoteMockRepository().withProjects(projects))

    output, err := executeNoteCommand([]string{projectPath, "path-based note"})

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !strings.Contains(output, "✓ Note saved") {
        t.Errorf("expected success message, got: %s", output)
    }
    if projects[0].Notes != "path-based note" {
        t.Errorf("expected note to be set")
    }
}
```

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Create new function in favorite.go/note.go | Use existing `findProjectByIdentifier` from status.go |
| Change `findProjectByIdentifier` signature | Use it exactly as-is |
| Remove error handling logic | Preserve SilenceErrors/SilenceUsage pattern |
| Forget path tests | Add at least one path-based test per command |
| Modify existing tests | Only add new path-based tests |

### Testing Commands

```bash
# Run all CLI tests
go test ./internal/adapters/cli/... -v

# Run specific tests
go test ./internal/adapters/cli/... -run TestFavorite -v
go test ./internal/adapters/cli/... -run TestNote -v

# Lint check
golangci-lint run

# Build and manual test
make build
./bin/vibe add /tmp/test-project  # Add test project if needed
./bin/vibe favorite /tmp/test-project  # Should work now
./bin/vibe note /tmp/test-project "test note"  # Should work now
```

### Completion Checklist

Before marking story complete:

- [x] `favorite.go` calls `findProjectByIdentifier` (no manual loop)
- [x] `note.go` calls `findProjectByIdentifier` (no manual loop)
- [x] `vibe favorite /path/to/project` works
- [x] `vibe note /path/to/project "note"` works
- [x] All existing tests pass
- [x] New path-based tests exist and pass
- [x] `golangci-lint run` passes

### References

- [Source: docs/sprint-artifacts/retrospectives/epic-6-retro-2025-12-25.md#H5] - DRY refactoring action item
- [Source: internal/adapters/cli/status.go:81-116] - `findProjectByIdentifier` implementation
- [Source: internal/adapters/cli/favorite.go:75-80] - Loop to remove
- [Source: internal/adapters/cli/note.go:60-65] - Loop to remove

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Verify Tests Pass

```bash
go test ./internal/adapters/cli/... -v
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| All tests pass | `ok` for package | Any `FAIL` |
| Path-based tests exist | Tests with "ByPath" in name | No path tests |

### Step 2: Verify Path-Based Commands Work

```bash
# Add a test project (if not exists)
mkdir -p /tmp/test-vibe-project
./bin/vibe add /tmp/test-vibe-project

# Test path-based favorite (NEW capability)
./bin/vibe favorite /tmp/test-vibe-project
./bin/vibe status /tmp/test-vibe-project  # Should show Favorite: Yes

# Test path-based note (NEW capability)
./bin/vibe note /tmp/test-vibe-project "test note"
./bin/vibe status /tmp/test-vibe-project  # Should show Notes: test note
```

### Step 3: Verify Duplicate Code Removed

```bash
# Check for remaining duplicate patterns
grep -n "for _, p := range projects" internal/adapters/cli/favorite.go
grep -n "for _, p := range projects" internal/adapters/cli/note.go
```

Expected: No output (loops removed)

### Decision Guide

| Situation | Action |
|-----------|--------|
| All tests pass, path commands work, no duplicates | Mark `done` |
| Tests fail | Do NOT approve, investigate failures |
| Path commands fail | Do NOT approve, check error handling |
| Duplicate loops still exist | Do NOT approve, incomplete refactoring |

## Dependencies

- Story 6.2 (Project Status Command) - Provides `findProjectByIdentifier()` function (COMPLETE)

## Review Follow-ups (Code Review)

- [ ] [AI-Review][MEDIUM] Duplicate mock repository implementations in `favorite_test.go:21-123` and `note_test.go:21-123` - extract to shared test helper (out of scope for this story, tracked for future)

## Dev Agent Record

### Context Reference

N/A - Story fully specified with implementation guidance

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Clean implementation, no debugging required

### Completion Notes List

- **AC1 Complete:** `favorite.go:68-75` now calls `findProjectByIdentifier` instead of manual loop (removed ~15 lines)
- **AC2 Complete:** `note.go:53-60` now calls `findProjectByIdentifier` instead of manual loop (removed ~15 lines)
- **AC3 Complete:** Verified `vibe favorite /path` and `vibe note /path` work correctly via manual testing
- **AC4 Complete:** Error handling pattern preserved with `cmd.SilenceErrors` and `cmd.SilenceUsage`
- **AC5 Complete:** All 190 CLI tests pass including existing favorite/note tests
- **AC6 Complete:** Added 3 new path-based tests: `TestFavoriteCmd_FindByPath`, `TestNoteCmd_FindByPath`, `TestNoteCmd_ViewByPath`
- **Test updates:** Updated `TestFavoriteCmd_FindAllError` and `TestNoteCmd_FindAllError` to expect new error message format from `findProjectByIdentifier` ("failed to find project" instead of "failed to load projects")
- **Net code reduction:** ~30 lines removed, ~50 test lines added for path coverage

### File List

- `internal/adapters/cli/favorite.go` - Modified: Replaced duplicate lookup with `findProjectByIdentifier` call
- `internal/adapters/cli/note.go` - Modified: Replaced duplicate lookup with `findProjectByIdentifier` call
- `internal/adapters/cli/favorite_test.go` - Modified: Added `TestFavoriteCmd_FindByPath`, `TestFavoriteCmd_FindByPath_NotFound`, updated error message expectation
- `internal/adapters/cli/note_test.go` - Modified: Added `TestNoteCmd_FindByPath`, `TestNoteCmd_ViewByPath`, `TestNoteCmd_FindByPath_NotFound`, updated error message expectation
