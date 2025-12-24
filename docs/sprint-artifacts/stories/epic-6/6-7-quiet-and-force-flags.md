# Story 6.7: Quiet and Force Flags

Status: done

## Story

As a **scripter**,
I want **quiet mode and force flags for CLI commands**,
So that **I can automate vibe-dash operations without prompts or noisy output in scripts and CI/CD pipelines**.

## Acceptance Criteria

1. **AC1: Quiet mode suppresses success output**
   - Given I need silent operation
   - When I run `vibe add . --quiet`
   - Then no output on success (stdout is silent)
   - And exit code indicates result (0 for success)
   - And errors still go to stderr

2. **AC2: Quiet mode works with add command**
   - Given a project path exists
   - When I run `vibe add /path/to/project --quiet`
   - Then project is added silently
   - And exit code is 0
   - And no "Added:" message is printed

3. **AC3: Quiet mode works with remove command**
   - Given project "client-alpha" is tracked
   - When I run `vibe remove client-alpha --quiet --force`
   - Then project is removed silently
   - And exit code is 0
   - And no "Removed:" message is printed

4. **AC4: Force flag on add resolves collisions automatically**
   - Given project "api-service" already exists
   - When I run `vibe add /other/api-service --force`
   - Then collision is resolved automatically (parent directory prefix)
   - And no prompt is shown
   - And project is added with auto-generated unique name

5. **AC5: Force flag on remove skips confirmation**
   - Given project "client-alpha" is tracked
   - When I run `vibe remove client-alpha --force`
   - Then no confirmation prompt is shown
   - And project is removed immediately
   - And "Removed:" message is printed (unless --quiet also used)

6. **AC6: Combined flags work together**
   - Given project "client-alpha" is tracked
   - When I run `vibe remove client-alpha --force --quiet`
   - Then no confirmation prompt
   - And no output (silent removal)
   - And exit code is 0

7. **AC7: Quiet mode still reports errors to stderr**
   - Given invalid project path
   - When I run `vibe add /nonexistent --quiet`
   - Then error message goes to stderr
   - And exit code is non-zero (1 for general error)

8. **AC8: Global quiet flag (--quiet or -q)**
   - Given I want to suppress all output
   - When I run `vibe -q add .`
   - Then same behavior as `vibe add . --quiet`
   - And short form `-q` is supported

9. **AC9: Force flag behavior consistency**
   - Given --force is already implemented on add and remove
   - When I verify existing behavior
   - Then add --force auto-resolves name collisions (already implemented)
   - And remove --force skips confirmation (already implemented)
   - And behavior remains unchanged

## Tasks / Subtasks

- [x] Task 1: Add global --quiet flag (AC: 1, 7, 8)
  - [x] 1.1: Add `quiet bool` to package-level variables in `flags.go:10-15`
  - [x] 1.2: Add `RootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, ...)` in `flags.go:init()`
  - [x] 1.3: Add `IsQuiet() bool` getter function
  - [x] 1.4: Add `ResetQuietFlag()` for test isolation (follows ResetAddFlags pattern)
  - [x] 1.5: Unit test: `TestIsQuiet` and `TestGlobalQuietFlagPosition` (vibe -q add .)

- [x] Task 2: Update add command for quiet mode (AC: 2, 4)
  - [x] 2.1: Wrap lines 211-215 success output in `if !IsQuiet() { ... }`
  - [x] 2.2: Verify --force collision resolution unchanged (lines 175-177)
  - [x] 2.3: Unit test: quiet mode produces empty stdout

- [x] Task 3: Update remove command for quiet mode (AC: 3, 5, 6)
  - [x] 3.1: Wrap line 143 success output in `if !IsQuiet() { ... }`
  - [x] 3.2: Verify --force confirmation skip unchanged (lines 112-125)
  - [x] 3.3: Unit test: quiet+force produces empty stdout

- [x] Task 4: Update note/rename/favorite commands (AC: 1)
  - [x] 4.1: note.go lines 96-98: Wrap "Note saved/cleared" in IsQuiet()
  - [x] 4.2: rename.go lines 106, 118: Wrap "Cleared/Renamed" in IsQuiet()
  - [x] 4.3: favorite.go lines 116-118: Wrap "Favorited/Unfavorited" in IsQuiet()
  - [x] 4.4: Unit tests for each command

- [x] Task 5: Verify error handling (AC: 7)
  - [x] 5.1: Errors use `cmd.ErrOrStderr()` not stdout
  - [x] 5.2: Exit codes unaffected by --quiet
  - [x] 5.3: Unit test: errors appear on stderr with --quiet

- [x] Task 6: Integration verification
  - [x] 6.1: Manual testing per User Testing Guide

## Dev Notes

### Critical Rules

1. **--force ALREADY EXISTS** on add (`add.go:64,92`) and remove (`remove.go:19,47`). DO NOT re-add.
2. **Use PersistentFlags** for --quiet (must work as `vibe -q add .` per AC8).
3. **Quiet suppresses stdout only** - errors MUST still go to stderr via `cmd.ErrOrStderr()`.
4. **Exit codes unchanged** - quiet mode is output-only, no behavior change.
5. **Add ResetQuietFlag()** - required for test isolation per project conventions.
6. **Flag conflict: --verbose + --quiet** - If both set, --quiet wins (silent operation is explicit intent).

### Commands NOT Updated

| Command | Reason |
|---------|--------|
| `list` | Primary output IS the list - quiet mode would be useless |
| `status` | Primary output IS the status - quiet mode would be useless |
| `completion` | Outputs completion script - quiet mode would break functionality |
| `exists` | Uses exit code only (Story 6.4) - already silent |

### Implementation Reference

**flags.go additions:**
```go
// Add to package-level vars (line ~11)
quiet bool

// Add to init() (line ~18)
RootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error output")

// Add getter + reset functions
func IsQuiet() bool { return quiet }
func ResetQuietFlag() { quiet = false }
```

**Command update pattern:**
```go
// Wrap success output:
if !IsQuiet() {
    fmt.Fprintf(cmd.OutOrStdout(), "✓ Added: %s\n", displayName)
}
```

### Exact Lines to Modify

| File | Line(s) | Current Code | Change |
|------|---------|--------------|--------|
| `flags.go` | 10-15 | var block | Add `quiet bool` |
| `flags.go` | 17-27 | init() | Add PersistentFlags for quiet |
| `flags.go` | NEW | - | Add IsQuiet(), ResetQuietFlag() |
| `add.go` | 211-215 | fmt.Fprintf success | Wrap in if !IsQuiet() |
| `remove.go` | 143 | fmt.Fprintf success | Wrap in if !IsQuiet() |
| `note.go` | 96-98 | fmt.Fprintln success | Wrap in if !IsQuiet() |
| `rename.go` | 96-97, 106, 118 | fmt.Fprintf success | Wrap in if !IsQuiet() |
| `favorite.go` | 97, 116-118 | fmt.Fprintf success | Wrap in if !IsQuiet() |

### Test Requirements

Use table-driven tests per project conventions:

```go
func TestQuietMode(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantOut  string // empty for quiet
    }{
        {"add with --quiet", []string{"add", tmpDir, "--quiet"}, ""},
        {"add with -q global", []string{"-q", "add", tmpDir}, ""},
        {"remove with --quiet --force", []string{"remove", "proj", "--quiet", "--force"}, ""},
    }
    // ... table-driven implementation
}
```

**CRITICAL TEST: AC8 Global Flag Position**
```go
// Must test: vibe -q add . (global flag BEFORE subcommand)
output, err := executeCommand(rootCmd, "-q", "add", tmpDir)
assert.Empty(t, output) // Verifies global flag works
```

### File Locations

| File | Action |
|------|--------|
| `internal/adapters/cli/flags.go` | MODIFY |
| `internal/adapters/cli/flags_test.go` | MODIFY |
| `internal/adapters/cli/add.go` | MODIFY |
| `internal/adapters/cli/add_test.go` | MODIFY |
| `internal/adapters/cli/remove.go` | MODIFY |
| `internal/adapters/cli/remove_test.go` | MODIFY |
| `internal/adapters/cli/note.go` | MODIFY |
| `internal/adapters/cli/note_test.go` | MODIFY |
| `internal/adapters/cli/rename.go` | MODIFY |
| `internal/adapters/cli/rename_test.go` | MODIFY |
| `internal/adapters/cli/favorite.go` | MODIFY |
| `internal/adapters/cli/favorite_test.go` | MODIFY |

### Project Context

- **Exit Codes:** 0=Success, 1=General error, 2=Project not found, 3=Config invalid, 4=Detection failed
- **Testing:** Co-locate tests, table-driven preferred, use `cmd.OutOrStdout()` for testability
- **Pattern from 6.6:** Package-level flag vars + getter functions + Register*Command() for tests

### References

- [add.go:64,92] - Existing addForce flag definition
- [add.go:175-177] - Existing --force collision auto-resolution
- [remove.go:19,47] - Existing removeForce flag definition
- [remove.go:112-125] - Existing --force confirmation skip
- [flags.go:10-27] - Existing flag patterns (verbose, debug)
- [docs/epics.md:2509-2543] - Original story definition

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-6/6-7-quiet-and-force-flags.md`
- Previous story: `docs/sprint-artifacts/stories/epic-6/6-6-shell-completion.md`
- Project context: `docs/project-context.md`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None required - implementation straightforward.

### Completion Notes List

- Added `--quiet/-q` global flag to suppress success output
- All commands (add, remove, note, rename, favorite) respect quiet mode
- Errors still return properly with non-zero exit codes
- Tests added for all quiet mode functionality
- All existing tests continue to pass

### File List

| File | Action |
|------|--------|
| `internal/adapters/cli/flags.go` | MODIFIED - Added quiet flag, IsQuiet(), ResetQuietFlag(), SetQuietForTest() |
| `internal/adapters/cli/flags_test.go` | MODIFIED - Added quiet flag tests |
| `internal/adapters/cli/test_helpers_test.go` | MODIFIED - Added quiet to resetTestState() |
| `internal/adapters/cli/add.go` | MODIFIED - Wrapped success output in IsQuiet() check |
| `internal/adapters/cli/add_test.go` | MODIFIED - Added quiet mode tests |
| `internal/adapters/cli/remove.go` | MODIFIED - Wrapped success output in IsQuiet() check |
| `internal/adapters/cli/remove_test.go` | MODIFIED - Added quiet mode tests |
| `internal/adapters/cli/note.go` | MODIFIED - Wrapped success output in IsQuiet() check |
| `internal/adapters/cli/note_test.go` | MODIFIED - Added quiet mode tests |
| `internal/adapters/cli/rename.go` | MODIFIED - Wrapped success output in IsQuiet() check |
| `internal/adapters/cli/rename_test.go` | MODIFIED - Added quiet mode tests |
| `internal/adapters/cli/favorite.go` | MODIFIED - Wrapped success output in IsQuiet() check |
| `internal/adapters/cli/favorite_test.go` | MODIFIED - Added quiet mode tests |

### Change Log

- 2025-12-24: Story validated by SM agent. Applied improvements: Added ResetQuietFlag requirement, explicit line numbers for all commands, flag conflict rule, excluded commands table, AC8 global flag position test requirement, table-driven test pattern, condensed verbose sections for LLM optimization.
- 2025-12-24: Implementation completed by Dev agent. All tasks done, tests pass, linting clean.
- 2025-12-24: Code review passed by Dev agent. Zero issues found. All 9 ACs verified. Story marked done.

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Build Binary

```bash
cd ~/GitHub/JeiKeiLim/vibe-dash
make build
```

### Step 2: Test Quiet Mode on Add (AC1, AC2)

```bash
# Setup
./bin/vibe remove test-quiet-project --force 2>/dev/null || true
mkdir -p /tmp/test-quiet-project

# Test quiet add (should have NO output)
./bin/vibe add /tmp/test-quiet-project --quiet --name test-quiet-project
echo "Exit code: $?"

# Verify project was added
./bin/vibe list | grep test-quiet-project
```

**Expected:** No stdout, exit code 0, project in list.

### Step 3: Test Quiet Mode on Remove (AC3, AC6)

```bash
./bin/vibe remove test-quiet-project --quiet --force
echo "Exit code: $?"
./bin/vibe list | grep test-quiet-project || echo "Removed OK"
```

**Expected:** No stdout, exit code 0, project not in list.

### Step 4: Test Force Flag (AC4, AC5 - verify unchanged)

```bash
mkdir -p /tmp/clientA/api-service /tmp/clientB/api-service
./bin/vibe add /tmp/clientA/api-service --name "api-service"
./bin/vibe add /tmp/clientB/api-service --force  # Should auto-resolve
echo "Exit code: $?"
./bin/vibe list
```

**Expected:** No prompt, exit code 0, both projects tracked with unique names.

### Step 5: Test Errors to Stderr (AC7)

```bash
./bin/vibe add /nonexistent/path --quiet 2>&1
echo "Exit code: $?"
```

**Expected:** Error message visible, exit code non-zero.

### Step 6: Test Global Flag -q (AC8 - CRITICAL)

```bash
mkdir -p /tmp/test-short-quiet
./bin/vibe -q add /tmp/test-short-quiet --name test-short-quiet
echo "Exit code: $?"
./bin/vibe list | grep test-short-quiet
./bin/vibe remove test-short-quiet --force --quiet
```

**Expected:** `-q` before subcommand works, exit code 0.

### Cleanup

```bash
./bin/vibe remove api-service --force 2>/dev/null || true
./bin/vibe remove clientB-api-service --force 2>/dev/null || true
rm -rf /tmp/test-quiet-project /tmp/clientA /tmp/clientB /tmp/test-short-quiet
```

### Decision Guide

| Situation | Action |
|-----------|--------|
| All quiet mode tests pass | Mark story `done` |
| --quiet still produces stdout | Do NOT approve, check IsQuiet() wrapping |
| Errors don't appear with --quiet | Do NOT approve, verify stderr handling |
| -q before subcommand fails | Do NOT approve, check PersistentFlags |
| --force behavior changed | Do NOT approve, verify no regression |
| Exit codes wrong with --quiet | Do NOT approve, no behavior change allowed |
