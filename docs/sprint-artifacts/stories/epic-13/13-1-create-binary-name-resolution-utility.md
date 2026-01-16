# Story 13.1: Create Binary Name Resolution Utility

Status: done

## Story

As a developer,
I want a utility function that resolves the actual binary name at runtime,
so that all CLI output consistently reflects how the user invoked the command.

## User-Visible Changes

None - this is an internal infrastructure change. The utility enables future stories (13-2, 13-3) that will produce user-visible changes.

## Acceptance Criteria

1. **Given** the binary is invoked as `./vdash`, **when** the utility is called, **then** it returns "vdash"
2. **Given** the binary is invoked as `./my-custom-name`, **when** the utility is called, **then** it returns "my-custom-name"
3. **Given** the binary is invoked via symlink `~/bin/v`, **when** the utility is called, **then** it returns "v" (the symlink name)
4. **Given** `os.Args[0]` is empty or invalid, **when** the utility is called, **then** it returns "vdash" as fallback
5. **Given** `os.Args[0]` contains a full path `/usr/local/bin/vdash`, **when** the utility is called, **then** it returns "vdash" (base name only)

## DO NOT MODIFY

The following files should NOT be changed in this story:
- `internal/adapters/cli/root.go:13` - Cobra `Use` field (canonical for help text)
- `internal/adapters/cli/completion.go` - Help examples (canonical for docs)
- Any `Long:` or `Example:` strings - Keep "vdash" for documentation consistency

## Tasks / Subtasks

- [x] Task 1: Create binaryname.go file (AC: 1-5)
  - [x] Create `internal/adapters/cli/binaryname.go`
  - [x] Add package doc comment
  - [x] Implement `BinaryName() string` function
  - [x] Export `DefaultBinaryName` constant for reuse in Stories 13-2, 13-3
  - [x] Use `filepath.Base(os.Args[0])` for resolution
  - [x] Handle edge case: empty `os.Args[0]` → fallback to "vdash"
  - [x] Handle edge case: Args slice empty → fallback to "vdash"
  - [x] Handle edge case: `filepath.Base("/")` returns "/" → fallback to "vdash"
- [x] Task 2: Create unit tests (AC: 1-5)
  - [x] Create `internal/adapters/cli/binaryname_test.go`
  - [x] Test cases: normal invocation, full path, symlink scenario, empty args, root path
  - [x] Use table-driven tests following project convention
- [x] Task 3: Verify and finalize (AC: all)
  - [x] Confirm no conflicts with existing CLI code
  - [x] Verify function can be called from version.go, exitcodes.go, and other CLI files
  - [x] Run `make fmt && make lint` before marking complete
  - [x] Manual verification: `make build && ./bin/vdash --version` shows "vdash"

## Dev Notes

### Technical Requirements

**Location:** `internal/adapters/cli/binaryname.go`

**Function signature:**
```go
// DefaultBinaryName is the fallback name when os.Args[0] is unavailable or invalid.
const DefaultBinaryName = "vdash"

// BinaryName returns the name of the binary as invoked by the user.
// Uses the basename of os.Args[0], falling back to DefaultBinaryName if unavailable.
func BinaryName() string
```

**Key implementation decisions:**
- Use `filepath.Base(os.Args[0])` for cross-platform resolution
- Fallback to `DefaultBinaryName` for: empty string, ".", "/" values
- Extract `binaryNameFrom(arg0 string)` for unit testability (os.Args is global state)

### Architecture Compliance

This utility lives in `internal/adapters/cli/` (adapter layer) because it directly uses `os.Args` (external dependency). It does NOT belong in `internal/core/`.

### Existing Code Context

**Files that will use BinaryName() in future stories:**
- `version.go:7` - Version template (Story 13-2)
- `list.go:91` - Suggestion message (Story 13-3)

**Current hardcoded "vdash" locations (reference only):**
- `version.go:7`, `root.go:13`, `completion.go`, `list.go:91`

### Testing Guidance

**Test file:** `internal/adapters/cli/binaryname_test.go`

**Testability approach:** Extract `binaryNameFrom(arg0 string)` to test the logic without mocking global `os.Args`.

**Test table:**
```go
tests := []struct {
    name     string
    arg0     string
    expected string
}{
    {"normal invocation", "vdash", "vdash"},
    {"full path unix", "/usr/local/bin/vdash", "vdash"},
    {"full path windows", "C:\\Program Files\\vdash.exe", "vdash.exe"},
    {"full path with spaces", "/path/to/my tool", "my tool"},
    {"renamed binary", "my-custom-name", "my-custom-name"},
    {"symlink name", "v", "v"},
    {"empty string", "", "vdash"},
    {"dot only", ".", "vdash"},
    {"root path unix", "/", "vdash"},
}
```

**Manual verification:**
```bash
make build && ./bin/vdash version
# Expected: "vdash version X.Y.Z ..."
```

### References

- Epic source: `docs/epics-phase2.md#Story 1.1`
- Architecture: `docs/architecture.md#Project Structure` → CLI adapter location
- Patterns: `docs/project-context.md#Go Patterns` → Constructor/function patterns
- FR Coverage: FR-P2-19, FR-P2-20 (enabler for these stories)

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - No debug issues encountered.

### Completion Notes List

- Implemented `BinaryName()` and `binaryNameFrom()` functions in `internal/adapters/cli/binaryname.go`
- Used TDD approach: wrote failing tests first, then implementation
- Exported `DefaultBinaryName` constant for reuse in Stories 13-2, 13-3
- Handled edge cases: empty Args, empty string, ".", "/" all return fallback "vdash"
- Table-driven tests with platform-aware Windows test (skipped on non-Windows)
- All 1301 tests pass, `make fmt && make lint` successful
- Manual verification: `./bin/vdash --version` shows "vdash version v0.1.1..."

### Code Review Fixes Applied

- **M1 Fixed:** Enhanced `TestBinaryName()` to verify result matches `filepath.Base(os.Args[0])`
- **M2 Fixed:** Added `TestBinaryName_EmptyArgsDocumentation()` documenting why empty Args branch is untestable
- **M3 Fixed:** Moved package doc comment, kept only file-specific comment in binaryname.go
- All 1302 tests pass after fixes

### File List

- `internal/adapters/cli/binaryname.go` (created)
- `internal/adapters/cli/binaryname_test.go` (created)
