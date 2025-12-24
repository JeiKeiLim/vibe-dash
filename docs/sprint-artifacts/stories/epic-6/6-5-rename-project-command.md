# Story 6.5: Rename Project Command

Status: done

## Story

As a **user**,
I want **to rename projects via CLI by setting or clearing display names**,
So that **I can manage project identities without using the TUI**.

## Acceptance Criteria

1. **AC1: Set display name**
   - Given project "api-service" exists
   - When I run `vibe rename api-service "Client A API"`
   - Then display_name is set to "Client A API"
   - And message shows: `✓ Renamed: api-service → Client A API`

2. **AC2: Clear display name with --clear**
   - Given project "api-service" has display_name "Client A API"
   - When I run `vibe rename api-service --clear`
   - Then display_name is cleared (empty string)
   - And message shows: `Cleared display name: api-service`

3. **AC3: Clear display name with empty string**
   - Given project "api-service" has display_name "Client A API"
   - When I run `vibe rename api-service ""`
   - Then display_name is cleared (empty string)
   - And message shows: `Cleared display name: api-service`

4. **AC4: Project not found error**
   - Given project "nonexistent" is not tracked
   - When I run `vibe rename nonexistent "New Name"`
   - Then error is returned with `domain.ErrProjectNotFound`
   - And exit code is 2 (via `MapErrorToExitCode`)

5. **AC5: Requires project name argument**
   - Given no arguments provided
   - When I run `vibe rename`
   - Then error shows usage (Cobra default behavior)
   - And exit code is 1

6. **AC6: Requires new name or --clear flag**
   - Given only project name provided
   - When I run `vibe rename api-service` (no name, no --clear)
   - Then error shows: `requires a new name or --clear flag`
   - And exit code is 1

7. **AC7: Lookup by name, display name, or path**
   - Given project "client-alpha" with display_name "My Client"
   - When I run `vibe rename "My Client" "New Display Name"`
   - Then lookup by display_name succeeds
   - And display_name is updated to "New Display Name"

8. **AC8: Idempotent clear on already-cleared project**
   - Given project "api-service" has no display_name (already empty)
   - When I run `vibe rename api-service --clear`
   - Then message shows: `☆ api-service has no display name`
   - And exit code is 0 (success, not error)

## Tasks / Subtasks

- [x] Task 1: Create rename command file (AC: 1, 2, 3, 4, 5, 6, 7)
  - [x] 1.1: Create `internal/adapters/cli/rename.go`
  - [x] 1.2: Define `newRenameCmd()` with Cobra command structure
    - Use: `rename <project-name> [new-display-name]`
    - Args: `cobra.RangeArgs(1, 2)` (project required, name optional if --clear)
    - Add `--clear` flag: `BoolVar(&renameClear, "clear", false, "Clear display name")`
  - [x] 1.3: Implement `runRename()` function:
    - Validate args: require new-name OR --clear flag (AC6)
    - Use `findProjectByIdentifier()` from status.go (reuse existing)
    - On not found: wrap with `domain.ErrProjectNotFound`, silence errors, return (AC4)
    - Set display_name, update UpdatedAt timestamp
    - Call `repository.Save()`
    - Output success message
  - [x] 1.4: Register command in `init()` with `RootCmd.AddCommand()`
  - [x] 1.5: Add `RegisterRenameCommand(parent *cobra.Command)` for test registration
  - [x] 1.6: Add `ResetRenameFlags()` for test cleanup

- [x] Task 2: Handle clear scenarios (AC: 2, 3, 8)
  - [x] 2.1: --clear flag clears display_name
  - [x] 2.2: Empty string "" clears display_name (same behavior as --clear)
  - [x] 2.3: Idempotent: if already empty, print info message and return success

- [x] Task 3: Unit tests (AC: 1-8)
  - [x] 3.1: Create `internal/adapters/cli/rename_test.go`
  - [x] 3.2: Test set display name -> `✓ Renamed:` message
  - [x] 3.3: Test clear with --clear flag -> `✓ Cleared display name:` message
  - [x] 3.4: Test clear with empty string "" -> same as --clear
  - [x] 3.5: Test project not found -> exit code 2, domain error
  - [x] 3.6: Test no arguments -> exit code 1 (Cobra error)
  - [x] 3.7: Test only project name (no new name, no --clear) -> error "requires..."
  - [x] 3.8: Test lookup by display_name succeeds
  - [x] 3.9: Test lookup by path succeeds
  - [x] 3.10: Test idempotent clear on already-empty -> `☆ ... has no display name`, exit 0
  - [x] 3.11: Verify UpdatedAt timestamp is updated after rename

- [x] Task 4: Integration verification
  - [x] 4.1: Manual testing with real binary per User Testing Guide

## Dev Notes

### Implementation Pattern - Follow favorite.go

The rename command follows the same pattern as `favorite.go` (primary reference):
- Package-level flag variable `renameClear` for `--clear` flag
- `ResetRenameFlags()` for test cleanup (see `favorite.go:17-19`)
- `RegisterRenameCommand()` for test registration (see `favorite.go:47-51`)
- Use `findProjectByIdentifier()` from status.go (lines 79-114)
- Idempotent behavior pattern from `favorite.go:94-98`

Also reference `note.go` for the set/clear output logic (lines 75-98).

### CRITICAL: Reuse findProjectByIdentifier

**DO NOT duplicate project lookup logic.** The function `findProjectByIdentifier()` in `status.go:79-114` already handles:
- Lookup by exact name
- Lookup by display_name
- Lookup by canonical path
- Returns wrapped `domain.ErrProjectNotFound`

```go
// In runRename(), just call:
proj, err := findProjectByIdentifier(ctx, args[0])
if err != nil {
    if errors.Is(err, domain.ErrProjectNotFound) {
        cmd.SilenceErrors = true
        cmd.SilenceUsage = true
    }
    return err
}
```

### Argument Validation Logic

```go
// Pseudo-code for argument validation (AC5, AC6)
func runRename(cmd *cobra.Command, args []string) error {
    if len(args) < 1 {
        // Cobra handles this with Args: cobra.RangeArgs(1, 2)
    }

    projectIdentifier := args[0]

    // Determine operation mode
    var newDisplayName string
    clearMode := false

    if renameClear {
        clearMode = true
    } else if len(args) == 2 {
        newDisplayName = args[1]
        if newDisplayName == "" {
            clearMode = true // Empty string = clear
        }
    } else {
        // No new name and no --clear flag
        return fmt.Errorf("requires a new name or --clear flag")
    }

    // ... rest of implementation
}
```

### Output Messages (Match Epic 6 Style)

Follow the established patterns from `favorite.go` and Epic 6:

- **Set name:** `✓ Renamed: {project} → {new_name}`
- **Clear name:** `✓ Cleared display name: {project}`
- **Already empty:** `☆ {project} has no display name`
- **Not found:** (no output - error returned with SilenceErrors)

**Important:** Use `→` (Unicode arrow), not `->`. This command does NOT use `SilentError` wrapper (unlike `exists.go`) because it produces output.

### Critical Rules

1. **Reuse `findProjectByIdentifier()`** - Same package, just call it. Do NOT duplicate lookup logic.
2. **Error wrapping** - `findProjectByIdentifier` already wraps with `domain.ErrProjectNotFound`. Do NOT re-wrap.
3. **Silence on not found** - Set `cmd.SilenceErrors = true` AND `cmd.SilenceUsage = true` before returning error.
4. **UpdatedAt timestamp** - Always set `targetProject.UpdatedAt = time.Now()` before save.
5. **No SilentError** - Unlike `exists.go`, this command produces output. Do NOT use `SilentError` wrapper.

### File Locations

| File | Action | Notes |
|------|--------|-------|
| `internal/adapters/cli/rename.go` | CREATE | ~80 lines |
| `internal/adapters/cli/rename_test.go` | CREATE | ~150 lines |
| `internal/adapters/cli/status.go:79-114` | REFERENCE | findProjectByIdentifier to reuse |
| `internal/adapters/cli/exitcodes.go` | REFERENCE | MapErrorToExitCode |

### Project Structure Notes

- Follows hexagonal architecture: CLI adapter calls repository port
- No direct domain modification - goes through repository.Save()
- DisplayName is a field on `domain.Project` struct (project.go:16)

### References

- [Source: docs/epics.md#Story-6.5] - Original story definition
- [Source: internal/adapters/cli/status.go:79-114] - findProjectByIdentifier function
- [Source: internal/adapters/cli/note.go] - Similar command pattern
- [Source: internal/adapters/cli/favorite.go] - Flag reset pattern
- [Source: internal/core/domain/project.go:16] - DisplayName field
- [Source: docs/project-context.md#Exit-Codes] - Exit code mapping (2 = ProjectNotFound)

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-6/6-5-rename-project-command.md`
- Previous story: `docs/sprint-artifacts/stories/epic-6/6-4-project-exists-check.md`
- Project context: `docs/project-context.md`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None required - straightforward implementation.

### Completion Notes List

- Implemented `rename` command following `favorite.go` pattern
- Reused `findProjectByIdentifier()` from status.go for lookup by name, display_name, or path
- Clear mode supports both `--clear` flag and empty string `""`
- Idempotent clear returns success with info message when already empty
- All 16 unit tests pass covering AC1-AC8 plus edge cases
- Exit code 2 returned for ErrProjectNotFound via MapErrorToExitCode
- Integration verified: all ACs pass with real binary
- ~110 lines implementation, ~480 lines tests

### Code Review (2025-12-24)

**Reviewer:** Claude Opus 4.5

**Issues Found:** 0 High, 4 Medium, 2 Low

**Fixes Applied:**
- M3: Added `TestRenameCmd_EmptyStringWithClearFlag` - edge case test for `"" --clear` combination
- M4: Added `TestRenameCmd_HelpText` - validates help output

**Deferred (out of scope):**
- M1/M2: `favorite.go` and `note.go` don't reuse `findProjectByIdentifier()` - refactor opportunity for future
- L1/L2: Minor doc and test mock consolidation - acceptable as-is

### File List

| File | Action |
|------|--------|
| `internal/adapters/cli/rename.go` | CREATE |
| `internal/adapters/cli/rename_test.go` | CREATE |

### Change Log

- 2025-12-24: Story implementation complete - all acceptance criteria verified
- 2025-12-24: Code review complete - 2 edge case tests added, status → done

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Build Binary

```bash
cd ~/GitHub/JeiKeiLim/vibe-dash
make build
```

### Step 2: Test Set Display Name (AC1)

```bash
# Ensure a project is tracked
./bin/vibe list

# Set display name (use an existing project name)
./bin/vibe rename vibe-dash "Vibe Dashboard Tool"
echo "Exit code: $?"
```

**Expected:**
- Output: `✓ Renamed: vibe-dash → Vibe Dashboard Tool`
- Exit code: 0

### Step 3: Verify Display Name Set

```bash
./bin/vibe status vibe-dash
```

**Expected:**
- First line shows "Vibe Dashboard Tool" (the display name)
- OR the name shows in the details

### Step 4: Test Clear with --clear (AC2)

```bash
./bin/vibe rename vibe-dash --clear
echo "Exit code: $?"
```

**Expected:**
- Output: `✓ Cleared display name: vibe-dash`
- Exit code: 0

### Step 5: Test Clear with Empty String (AC3)

```bash
# First set a name
./bin/vibe rename vibe-dash "Temp Name"

# Then clear with empty string
./bin/vibe rename vibe-dash ""
echo "Exit code: $?"
```

**Expected:**
- Output: `✓ Cleared display name: vibe-dash`
- Exit code: 0

### Step 6: Test Not Found (AC4)

```bash
./bin/vibe rename nonexistent-xyz-12345 "New Name"
echo "Exit code: $?"
```

**Expected:**
- Exit code: 2 (NOT 1)
- Error output about project not found

### Step 7: Test Missing Arguments (AC5, AC6)

```bash
# No arguments
./bin/vibe rename
echo "Exit code: $?"

# Only project name, no new name, no --clear
./bin/vibe rename vibe-dash
echo "Exit code: $?"
```

**Expected:**
- Both: Exit code 1
- First: Cobra usage error
- Second: Error about "requires a new name or --clear flag"

### Step 8: Test Idempotent Clear (AC8)

```bash
# Ensure display name is already cleared
./bin/vibe rename vibe-dash --clear

# Try to clear again
./bin/vibe rename vibe-dash --clear
echo "Exit code: $?"
```

**Expected:**
- Output: `☆ vibe-dash has no display name`
- Exit code: 0 (success, not error)

### Decision Guide

| Situation | Action |
|-----------|--------|
| All tests pass | Mark story `done` |
| Exit code 1 instead of 2 for not found | Do NOT approve, check error wrapping |
| Missing --clear flag handling | Do NOT approve, fix argument validation |
| Display name not actually saved | Do NOT approve, check repository.Save call |
