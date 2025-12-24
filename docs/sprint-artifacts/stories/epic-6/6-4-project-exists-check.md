# Story 6.4: Project Exists Check

Status: done

## Story

As a **scripter**,
I want **to check if a project is tracked with a silent exit code**,
So that **I can conditionally add projects or skip actions in automation scripts**.

## Acceptance Criteria

1. **AC1: Silent success when project exists**
   - Given project "client-alpha" is tracked
   - When I run `vibe exists client-alpha`
   - Then exit code is 0
   - And NO output is produced (silent success)

2. **AC2: Silent failure when project not found**
   - Given project "nonexistent" is not tracked
   - When I run `vibe exists nonexistent`
   - Then exit code is 2 (ErrProjectNotFound mapped)
   - And NO output is produced (silent failure)

3. **AC3: Lookup by name, display name, or path**
   - Given project "client-alpha" with display name "My Client"
   - When I run `vibe exists client-alpha` (by name)
   - Then exit code is 0
   - When I run `vibe exists "My Client"` (by display name)
   - Then exit code is 0
   - When I run `vibe exists /path/to/client-alpha` (by path)
   - Then exit code is 0

4. **AC4: Exit code 2 uses domain error**
   - Given project lookup fails
   - When command returns error
   - Then error is wrapped with `domain.ErrProjectNotFound`
   - And `MapErrorToExitCode()` returns 2

5. **AC5: Usage in scripts**
   - Given the command works correctly
   - When used in a bash script:
     ```bash
     if vibe exists my-project; then
       echo "Already tracked"
     else
       vibe add .
     fi
     ```
   - Then the script correctly branches based on exit code

6. **AC6: Requires project identifier argument**
   - Given no arguments provided
   - When I run `vibe exists`
   - Then error shows usage (Cobra default behavior)
   - And exit code is 1

## Tasks / Subtasks

- [x] Task 1: Create exists command file (AC: 1, 2, 3, 4, 6)
  - [x] 1.1: Create `internal/adapters/cli/exists.go`
  - [x] 1.2: Define `newExistsCmd()` with Cobra command structure
  - [x] 1.3: Implement `runExists()` function:
    - Require exactly 1 argument (project identifier)
    - Use `findProjectByIdentifier()` from status.go (reuse existing function)
    - On success: return nil (exit 0, no output)
    - On failure: wrap with `domain.ErrProjectNotFound`, return error (exit 2)
  - [x] 1.4: Register command in `init()` with `RootCmd.AddCommand()`
  - [x] 1.5: Add `RegisterExistsCommand()` for test registration

- [x] Task 2: Ensure silent output (AC: 1, 2)
  - [x] 2.1: Verify no `fmt.Fprintf()` or output on success path
  - [x] 2.2: Set `cmd.SilenceErrors = true` and `cmd.SilenceUsage = true` before returning error
  - [x] 2.3: Verify no output on failure path

- [x] Task 3: Unit tests (AC: 1, 2, 3, 4, 6)
  - [x] 3.1: Create `internal/adapters/cli/exists_test.go`
  - [x] 3.2: Test project exists by name → exit 0, no output
  - [x] 3.3: Test project exists by display name → exit 0, no output
  - [x] 3.4: Test project exists by path → exit 0, no output
  - [x] 3.5: Test project not found → exit 2, no output
  - [x] 3.6: Test no arguments → exit 1 (Cobra error)
  - [x] 3.7: Verify `MapErrorToExitCode()` returns 2 for returned error

- [x] Task 4: Integration verification (AC: 5)
  - [x] 4.1: Manual testing with real binary per User Testing Guide

## Dev Notes

### CRITICAL: Silent Command - Zero Output

This command communicates **ONLY through exit codes**:
- Exit 0 = project exists
- Exit 2 = project not found

**No fmt.Print/Fprintf calls allowed.** Unlike status.go which prints "✗ Project not found:", exists.go must be completely silent.

### CRITICAL: Error Handling - Do NOT Double-Wrap

`findProjectByIdentifier()` in status.go:113 **already wraps with `domain.ErrProjectNotFound`**:
```go
return nil, fmt.Errorf("%w: %s", domain.ErrProjectNotFound, identifier)
```

In `runExists()`, **just return the error directly**:
```go
_, err := findProjectByIdentifier(ctx, identifier)
if err != nil {
    cmd.SilenceErrors = true
    cmd.SilenceUsage = true
    return err  // DO NOT wrap again!
}
```

### Implementation Checklist

1. **Create `exists.go`** with:
   - `newExistsCmd()` - Use: `exists <project-name>`, Args: `cobra.ExactArgs(1)`
   - `RegisterExistsCommand(parent *cobra.Command)` - For test registration (same pattern as status.go:66-70)
   - `init()` - Register with RootCmd
   - `runExists()` - Call findProjectByIdentifier, silence errors, return

2. **No package-level flags** - Unlike status.go, no flags needed. No `ResetExistsFlags()` required.

3. **Create `exists_test.go`** with imports:
   ```go
   import (
       "bytes"
       "testing"
       "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
       "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
   )
   ```

### Required Test Cases

| Test Name | Setup | Expected Exit | Expected Output |
|-----------|-------|---------------|-----------------|
| `TestExists_ProjectByName` | Mock with project | 0 | "" (empty) |
| `TestExists_ProjectByDisplayName` | Mock with display_name | 0 | "" (empty) |
| `TestExists_ProjectByPath` | Mock with real temp dir | 0 | "" (empty) |
| `TestExists_NotFound` | Empty mock | 2 | "" (empty) |
| `TestExists_NoArgument` | N/A | 1 | Cobra usage (via stderr) |
| `TestExists_ExitCodeMapping` | Empty mock | Verify `MapErrorToExitCode(err) == 2` | N/A |

### Anti-Patterns

| DON'T | DO |
|-------|-----|
| `fmt.Fprintf(cmd.OutOrStdout(), ...)` | Return nil (success) or error (failure) |
| `return fmt.Errorf("%w: %s", domain.ErrProjectNotFound, id)` | `return err` from findProjectByIdentifier |

### File Locations

| File | Action | Notes |
|------|--------|-------|
| `internal/adapters/cli/exists.go` | CREATE | ~30 lines |
| `internal/adapters/cli/exists_test.go` | CREATE | ~80 lines |
| `internal/adapters/cli/status.go:76-114` | REFERENCE | findProjectByIdentifier to reuse |
| `internal/adapters/cli/exitcodes.go` | REFERENCE | MapErrorToExitCode, ExitProjectNotFound (2) |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-6/6-4-project-exists-check.md`
- Previous story: `docs/sprint-artifacts/stories/epic-6/6-3-exit-codes.md`
- Project context: `docs/project-context.md`

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None required - straightforward implementation.

### Completion Notes List

1. Created `exists.go` (~60 lines) following the exact pattern from Dev Notes
2. Reuses `findProjectByIdentifier()` from status.go - no code duplication
3. Silent output achieved: no fmt.Printf/Fprintf calls, SilenceErrors/SilenceUsage set before error return
4. Error NOT double-wrapped - returns err directly from findProjectByIdentifier
5. All 8 unit tests pass (including edge cases)
6. Integration testing verified all ACs:
   - AC1: `./bin/vibe exists vibe-dash` → exit 0, no stdout output
   - AC2: `./bin/vibe exists nonexistent` → exit 2, no stdout output (slog.Error to stderr is from main.go, not the command)
   - AC3: Lookup by name, display name, and path all work (tested in unit tests)
   - AC4: MapErrorToExitCode returns 2 for ErrProjectNotFound
   - AC5: Script usage works: `if vibe exists project; then ... else ... fi`
   - AC6: Missing argument → exit 1 with Cobra usage message

### Code Review Fixes Applied

**Issue H1 (HIGH): Silent output violation - main.go slog.Error leaked to stderr**
- **Root cause:** main.go:52 logged all errors including silent command errors
- **Fix:** Added `SilentError` type in exitcodes.go that wraps errors to signal "don't log"
- **Changes:**
  - `exitcodes.go`: Added `SilentError` struct with `Unwrap()`, `IsSilentError()` helper
  - `exists.go`: Wrap error with `&SilentError{Err: err}` before returning
  - `main.go`: Check `IsSilentError(err)` before logging
  - `exitcodes_test.go`: Added tests for SilentError type
- **Result:** All 5 integration tests now pass (including silent on failure)

### File List

| File | Action |
|------|--------|
| `internal/adapters/cli/exists.go` | CREATED |
| `internal/adapters/cli/exists_test.go` | CREATED |
| `internal/adapters/cli/exitcodes.go` | MODIFIED (SilentError type) |
| `internal/adapters/cli/exitcodes_test.go` | MODIFIED (SilentError tests) |
| `cmd/vibe/main.go` | MODIFIED (IsSilentError check) |

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Build Binary

```bash
cd ~/GitHub/JeiKeiLim/vibe-dash
make build
```

### Step 2: Test Silent Success (AC1)

```bash
# First, ensure a project is tracked (e.g., vibe-dash)
./bin/vibe list  # Verify vibe-dash is tracked

# Test exists with known project
./bin/vibe exists vibe-dash
echo "Exit code: $?"
```

**Expected:**
- NO output (completely silent)
- Exit code: 0

### Step 3: Test Silent Failure (AC2)

```bash
# Test with nonexistent project
./bin/vibe exists nonexistent-project-xyz-12345
echo "Exit code: $?"
```

**Expected:**
- NO output (completely silent)
- Exit code: 2 (NOT 1)

### Step 4: Test Script Usage (AC5)

```bash
# Create and run test script
cat > /tmp/test-exists.sh << 'EOF'
#!/bin/bash

# Test with nonexistent project
if ./bin/vibe exists nonexistent-test-xyz 2>/dev/null; then
    echo "FAIL: Should not find nonexistent project"
else
    if [ $? -eq 2 ]; then
        echo "PASS: Exit code 2 for project not found"
    else
        echo "FAIL: Expected exit code 2, got $?"
    fi
fi

# Test with existing project (assumes vibe-dash is tracked)
if ./bin/vibe exists vibe-dash 2>/dev/null; then
    echo "PASS: Exit code 0 for existing project"
else
    echo "FAIL: vibe-dash should exist (exit code: $?)"
fi
EOF
chmod +x /tmp/test-exists.sh
/tmp/test-exists.sh
```

**Expected:**
```
PASS: Exit code 2 for project not found
PASS: Exit code 0 for existing project
```

### Step 5: Verify No Output

```bash
# Capture all output (should be empty)
output=$(./bin/vibe exists vibe-dash 2>&1)
if [ -z "$output" ]; then
    echo "PASS: No output on success"
else
    echo "FAIL: Unexpected output: $output"
fi

output=$(./bin/vibe exists nonexistent-xyz-12345 2>&1)
if [ -z "$output" ]; then
    echo "PASS: No output on failure"
else
    echo "FAIL: Unexpected output: $output"
fi
```

**Expected:**
```
PASS: No output on success
PASS: No output on failure
```

### Step 6: Test Missing Argument (AC6)

```bash
./bin/vibe exists
echo "Exit code: $?"
```

**Expected:**
- Error message about missing argument (Cobra default)
- Exit code: 1 (general error)

### Decision Guide

| Situation | Action |
|-----------|--------|
| All tests pass, no output on success/failure | Mark story `done` |
| Any output produced on success | Do NOT approve, remove fmt calls |
| Any output produced on failure | Do NOT approve, add SilenceErrors |
| Exit code 1 instead of 2 for not found | Do NOT approve, check error wrapping |
| Missing argument doesn't error | Do NOT approve, check Args validation |
