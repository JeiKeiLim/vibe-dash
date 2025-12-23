# Story 6.3: Exit Codes

Status: done

## Story

As a **scripter**,
I want **standard exit codes returned from all CLI commands**,
So that **I can handle errors programmatically in automation scripts and CI/CD pipelines**.

## Acceptance Criteria

1. **AC1: Success exit code**
   - Given any CLI command succeeds
   - When the command completes
   - Then exit code is 0

2. **AC2: General error exit code**
   - Given any unhandled or general error occurs
   - When the command fails
   - Then exit code is 1

3. **AC3: Project not found exit code**
   - Given project "nonexistent" does not exist
   - When I run `vibe status nonexistent` or `vibe remove nonexistent`
   - Then exit code is 2
   - And error message shows which project was not found

4. **AC4: Configuration invalid exit code**
   - Given config.yaml has invalid syntax or values
   - When I run `vibe config set waiting-threshold invalid`
   - Then exit code is 3
   - And error message describes the config issue

5. **AC5: Detection failed exit code**
   - Given methodology detection encounters an error
   - When detection fails for a project
   - Then exit code is 4
   - And error message describes what failed

6. **AC6: Exit codes documented in help**
   - Given I need to reference exit codes
   - When I run `vibe --help` or check documentation
   - Then exit codes are documented with meanings

7. **AC7: Consistent error-to-exit-code mapping**
   - Given domain errors are returned from services
   - When CLI commands propagate errors to main()
   - Then MapErrorToExitCode() consistently maps:
     - `ErrProjectNotFound` → 2
     - `ErrConfigInvalid` → 3
     - `ErrDetectionFailed` → 4
     - All other errors → 1

8. **AC8: Script usage example**
   - Given exit codes work correctly
   - When used in a bash script:
     ```bash
     if vibe status my-project > /dev/null 2>&1; then
       echo "Project exists"
     else
       case $? in
         2) echo "Project not found" ;;
         3) echo "Config issue" ;;
         *) echo "Error occurred" ;;
       esac
     fi
     ```
   - Then the script correctly branches based on exit code

## Tasks / Subtasks

- [x] Task 1: Verify exit code constants match documentation (AC: 1, 2, 3, 4, 5)
  - [x] 1.1: Review `internal/adapters/cli/exitcodes.go` constants
  - [x] 1.2: Verify constants match project-context.md table (0, 1, 2, 3, 4)
  - [x] 1.3: Ensure all domain error types have explicit mapping

- [x] Task 2: Audit CLI commands for exit code compliance (AC: 7)
  - [x] 2.1: Audit `add.go` - verify ErrProjectAlreadyExists returns exit 1
  - [x] 2.2: Audit `remove.go` - verify ErrProjectNotFound returns exit 2
  - [x] 2.3: Audit `status.go` - verify ErrProjectNotFound returns exit 2
  - [x] 2.4: Audit `list.go` - verify errors return appropriate codes
  - [x] 2.5: Audit `favorite.go` - verify ErrProjectNotFound returns exit 2
  - [x] 2.6: Audit `note.go` - verify ErrProjectNotFound returns exit 2
  - [x] 2.7: Audit `refresh.go` - verify errors return appropriate codes
  - [x] 2.8: Audit `config_cmd.go` - verify config errors return exit 3

- [x] Task 3: Add ErrConfigInvalid for config command errors (AC: 4)
  - [x] 3.1: Update `config_cmd.go` to wrap ALL validation errors with ErrConfigInvalid:
    - Invalid value (non-integer): `fmt.Errorf("%w: invalid value for waiting-threshold: %s", domain.ErrConfigInvalid, value)`
    - Negative value: `fmt.Errorf("%w: waiting-threshold must be >= 0, got %d", domain.ErrConfigInvalid, intVal)`
    - Unknown key: `fmt.Errorf("%w: unknown config key: %s", domain.ErrConfigInvalid, key)`
  - [x] 3.2: Add SilenceErrors/SilenceUsage pattern for clean error output (match status.go:164-165)
  - [x] 3.3: Test that invalid config values return exit code 3
  - [x] 3.4: Test that unknown config keys return exit code 3

- [x] Task 4: Add exit code documentation (AC: 6)
  - [x] 4.1: Update root command Long description with exit code table
  - [x] 4.2: Add `ExitCodeDescription()` function for programmatic access
  - [x] 4.3: Update README or docs with exit code reference (optional)

- [x] Task 5: Unit tests for exit code compliance (AC: 1-5, 7)
  - [x] 5.1: Test add command - already exists returns exit 1
  - [x] 5.2: Test remove command - not found returns exit 2
  - [x] 5.3: Test status command - not found returns exit 2
  - [x] 5.4: Test favorite command - not found returns exit 2
  - [x] 5.5: Test note command - not found returns exit 2
  - [x] 5.6: Test config command - invalid value returns exit 3
  - [x] 5.7: Test config command - unknown key returns exit 3
  - [x] 5.8: Test all commands - success returns exit 0
  - [x] 5.9: Verify wrapped errors still map correctly (especially multi-layer wrapping)
  - [x] 5.10: Test config command shows clean error message (no Cobra duplicate)

- [x] Task 6: Integration test for script usage (AC: 8)
  - [x] 6.1: Create shell script test that verifies exit codes
  - [x] 6.2: Manual verification with real binary per User Testing Guide

## Dev Notes

### CRITICAL: Exit Code Infrastructure Already Exists

**The exit code mechanism is FULLY IMPLEMENTED.** This story is about:
1. Auditing existing commands for compliance
2. Adding proper error wrapping where missing
3. Documenting the exit codes
4. Ensuring all tests verify exit code behavior

### Existing Exit Code Implementation

**`internal/adapters/cli/exitcodes.go`:**
```go
const (
    ExitSuccess         = 0  // Command completed successfully
    ExitGeneralError    = 1  // Unhandled error, user decision needed
    ExitProjectNotFound = 2  // Project doesn't exist
    ExitConfigInvalid   = 3  // Configuration syntax/value error
    ExitDetectionFailed = 4  // Methodology detection failed
)

func MapErrorToExitCode(err error) int {
    // Uses errors.Is() for wrapped error support
    // Maps domain errors to exit codes
}
```

**`cmd/vibe/main.go` line 51-53:**
```go
if err := run(ctx); err != nil {
    slog.Error("application error", "error", err)
    os.Exit(cli.MapErrorToExitCode(err))
}
```

### Domain Errors to Exit Code Mapping

| Domain Error | Exit Code | When Used |
|--------------|-----------|-----------|
| (nil) | 0 | Success |
| `ErrProjectNotFound` | 2 | Project lookup fails |
| `ErrConfigInvalid` | 3 | Config parse/validation error |
| `ErrDetectionFailed` | 4 | Detection service error |
| `ErrProjectAlreadyExists` | 1 | Add duplicate project |
| `ErrPathNotAccessible` | 1 | Path doesn't exist or inaccessible |
| Any other error | 1 | General catch-all |

### Error Wrapping Pattern

Commands MUST wrap errors with domain errors for proper exit code mapping:

```go
// CORRECT - will map to exit code 2
return fmt.Errorf("%w: %s", domain.ErrProjectNotFound, projectName)

// INCORRECT - will map to exit code 1
return fmt.Errorf("project not found: %s", projectName)
```

### config_cmd.go Fix Required

Currently `config_cmd.go` returns generic errors at three locations. All must be wrapped with `ErrConfigInvalid`:

**Location 1: Line 57 - Invalid value (non-integer)**
```go
// BEFORE (maps to exit 1)
return fmt.Errorf("invalid value for waiting-threshold: %s", value)

// AFTER (maps to exit 3)
return fmt.Errorf("%w: invalid value for waiting-threshold: %s", domain.ErrConfigInvalid, value)
```

**Location 2: Line 60 - Negative value**
```go
// BEFORE (maps to exit 1)
return fmt.Errorf("waiting-threshold must be >= 0, got %d", intVal)

// AFTER (maps to exit 3)
return fmt.Errorf("%w: waiting-threshold must be >= 0, got %d", domain.ErrConfigInvalid, intVal)
```

**Location 3: Line 64 - Unknown config key**
```go
// BEFORE (maps to exit 1)
return fmt.Errorf("unknown config key: %s. Valid keys: waiting-threshold", key)

// AFTER (maps to exit 3)
return fmt.Errorf("%w: unknown config key: %s", domain.ErrConfigInvalid, key)
```

**CRITICAL: Add SilenceErrors Pattern**

After wrapping errors, add the SilenceErrors/SilenceUsage pattern to prevent Cobra from printing duplicate errors. Update `runConfigSet` to use a helper:

```go
func runConfigSet(cmd *cobra.Command, args []string) error {
    err := runConfigSetInternal(cmd, args)
    if err != nil && errors.Is(err, domain.ErrConfigInvalid) {
        cmd.SilenceErrors = true
        cmd.SilenceUsage = true
    }
    return err
}
```

Or inline in each error case before returning.

### Required Import for config_cmd.go

Add this import to `config_cmd.go`:
```go
import (
    // ... existing imports ...
    "errors"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)
```

### Existing Test Coverage

`exitcodes_test.go` already tests:
- All exit code constant values
- Direct domain error mapping
- Wrapped error mapping
- Deeply wrapped error mapping

### What Needs to be Added

1. **Audit each command** - Verify domain errors used correctly
2. **Fix config_cmd.go** - Wrap with ErrConfigInvalid
3. **Add root command documentation** - Exit code table in help text
4. **Add integration tests** - Per-command exit code verification

### Command-by-Command Audit Checklist

| Command | Not Found | Already Exists | Config Error | SilenceErrors | Status |
|---------|-----------|----------------|--------------|---------------|--------|
| `add` | N/A | ✅ Exit 1 | N/A | N/A | OK |
| `remove` | ✅ Exit 2 | N/A | N/A | ✅ | OK |
| `status` | ✅ Exit 2 | N/A | N/A | ✅ | OK |
| `list` | N/A | N/A | N/A | N/A | OK |
| `favorite` | ✅ Exit 2 | N/A | N/A | ❓ Verify | AUDIT |
| `note` | ✅ Exit 2 | N/A | N/A | ❓ Verify | AUDIT |
| `refresh` | N/A | N/A | N/A | N/A | OK |
| `config` | N/A | N/A | ❌ Generic | ❌ Missing | NEEDS FIX |

### Exit Code Flow (Reference)

```
config_cmd.go:runConfigSet()
    ↓ returns error wrapped with domain.ErrConfigInvalid
cmd/vibe/main.go:run() (line 51-53)
    ↓ calls cli.MapErrorToExitCode(err)
os.Exit(exitCode)
```

### File Locations

| File | Purpose |
|------|---------|
| `internal/adapters/cli/exitcodes.go` | Exit code constants and mapping (EXISTS) |
| `internal/adapters/cli/exitcodes_test.go` | Exit code tests (EXISTS) |
| `internal/adapters/cli/config_cmd.go` | NEEDS FIX - wrap with ErrConfigInvalid + SilenceErrors |
| `internal/adapters/cli/config_cmd_test.go` | ADD tests for exit code 3 |
| `internal/adapters/cli/favorite.go` | AUDIT - verify SilenceErrors pattern |
| `internal/adapters/cli/note.go` | AUDIT - verify SilenceErrors pattern |
| `internal/adapters/cli/root.go` | ADD exit code documentation to help |
| `internal/core/domain/errors.go` | Domain errors (EXISTS - ErrConfigInvalid available) |
| `cmd/vibe/main.go` | Exit code applied here (line 51-53) |

### Anti-Patterns to AVOID

| DON'T | DO |
|-------|-----|
| Return raw string errors for expected failures | Wrap with domain errors |
| Call `os.Exit()` in command code | Return error, let main() handle |
| Ignore SilenceErrors pattern | Use for custom error messages |
| Test exit codes by calling os.Exit | Test via MapErrorToExitCode() |
| Return generic "not found" strings | Wrap with `domain.ErrProjectNotFound` |
| Return generic config validation errors | Wrap with `domain.ErrConfigInvalid` |
| Forget SilenceUsage with SilenceErrors | Set both to prevent Cobra output |

## Dev Agent Record

### Context Reference

- Story file: `docs/sprint-artifacts/stories/epic-6/6-3-exit-codes.md`
- Previous story: `docs/sprint-artifacts/stories/epic-6/6-2-project-status-command.md`
- Project context: `docs/project-context.md` (Exit Codes table)
- Architecture: `docs/architecture.md` (Error-to-Exit-Code Mapping)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - No issues during implementation

### Completion Notes List

1. **Exit code constants verified** - All 5 exit codes (0-4) match project-context.md documentation
2. **CLI commands audited** - All commands properly use domain errors for exit code mapping
3. **config_cmd.go fixed** - All validation errors now wrap with `ErrConfigInvalid` for exit code 3
4. **SilenceErrors pattern added** - favorite.go, note.go, and config_cmd.go now use SilenceErrors/SilenceUsage
5. **Exit code documentation added** - `vibe --help` now shows exit codes table
6. **ExitCodeDescription() added** - Programmatic access to exit code meanings
7. **Unit tests added** - 4 new config exit code tests, 1 ExitCodeDescription test
8. **Manual integration verified** - All exit codes tested with real binary

### File List

- `internal/adapters/cli/config_cmd.go` - Added ErrConfigInvalid wrapping + SilenceErrors
- `internal/adapters/cli/config_cmd_test.go` - Added 4 exit code tests
- `internal/adapters/cli/favorite.go` - Added SilenceErrors pattern for ErrProjectNotFound
- `internal/adapters/cli/note.go` - Added SilenceErrors pattern for ErrProjectNotFound
- `internal/adapters/cli/exitcodes.go` - Added ExitCodeDescription() function
- `internal/adapters/cli/exitcodes_test.go` - Added TestExitCodeDescription
- `internal/adapters/cli/root.go` - Added exit code documentation to Long description
- `internal/adapters/cli/status.go` - Verified SilenceErrors pattern for ErrProjectNotFound (from Story 6.2)

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Build Binary

```bash
cd ~/GitHub/JeiKeiLim/vibe-dash
make build
```

### Step 2: Test Success Exit Code (AC1)

```bash
./bin/vibe list
echo "Exit code: $?"
```

**Expected:** Exit code: 0

### Step 3: Test Project Not Found Exit Code (AC3)

```bash
./bin/vibe status nonexistent-project-xyz-12345
echo "Exit code: $?"
```

**Expected:**
- Error message: "Project not found: nonexistent-project-xyz-12345"
- Exit code: 2 (NOT 1)

```bash
./bin/vibe remove nonexistent-project-xyz-12345 --force
echo "Exit code: $?"
```

**Expected:** Exit code: 2

### Step 4: Test Config Invalid Exit Code (AC4)

```bash
# Test invalid (non-integer) value
./bin/vibe config set vibe-dash waiting-threshold abc
echo "Exit code: $?"
```

**Expected:**
- Error message about invalid value
- Exit code: 3 (NOT 1)
- NO duplicate error from Cobra (single line output)

```bash
# Test negative value
./bin/vibe config set vibe-dash waiting-threshold -5
echo "Exit code: $?"
```

**Expected:** Exit code: 3

```bash
# Test unknown key
./bin/vibe config set vibe-dash unknown-key 123
echo "Exit code: $?"
```

**Expected:** Exit code: 3 (NOT 1)

### Step 5: Test Script Usage (AC8)

```bash
# Create test script
cat > /tmp/test-exit-codes.sh << 'EOF'
#!/bin/bash
./bin/vibe status nonexistent-test-project-xyz > /dev/null 2>&1
case $? in
  0) echo "FAIL: Expected non-zero exit" ;;
  2) echo "PASS: Exit code 2 for project not found" ;;
  *) echo "FAIL: Expected exit code 2, got $?" ;;
esac
EOF
chmod +x /tmp/test-exit-codes.sh
/tmp/test-exit-codes.sh
```

**Expected:** "PASS: Exit code 2 for project not found"

### Step 6: Test Existing Project Operations

```bash
# If you have a tracked project (e.g., vibe-dash)
./bin/vibe status vibe-dash
echo "Exit code: $?"
```

**Expected:** Exit code: 0

### Step 7: Verify Documentation (AC6)

```bash
./bin/vibe --help 2>&1 | grep -i "exit"
```

**Expected:** Exit codes should be mentioned in help output (after implementation)

### Decision Guide

| Situation | Action |
|-----------|--------|
| All exit codes match expected | Mark story `done` |
| Project not found returns 1 instead of 2 | Do NOT approve, domain error not wrapped |
| Config error returns 1 instead of 3 | Do NOT approve, config_cmd.go needs fix |
| Unknown config key returns 1 instead of 3 | Do NOT approve, ErrConfigInvalid not used |
| Duplicate error message from Cobra | Do NOT approve, SilenceErrors missing |
| Exit codes not in help output | Check if Task 4 was completed |
| Script test fails | Check MapErrorToExitCode mapping |
