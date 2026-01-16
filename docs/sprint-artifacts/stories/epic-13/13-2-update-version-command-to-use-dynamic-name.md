# Story 13.2: Update Version Command to Use Dynamic Name

Status: done

## Story

As a user,
I want the version output to show the actual binary name I used,
so that copy-pasting version info is accurate for bug reports.

## User-Visible Changes

- **Changed:** Version output now shows the actual binary name used (e.g., `my-tool version v0.1.1` instead of hardcoded `vdash version v0.1.1`)
- **Changed:** When running via symlink (e.g., `~/bin/v`), version shows `v version v0.1.1`

## Acceptance Criteria

1. **Given** I run `./vdash version`, **when** the version is displayed, **then** output shows "vdash version X.Y.Z" (dynamically resolved, not hardcoded)
2. **Given** I run `./my-tool version` (renamed binary), **when** the version is displayed, **then** output shows "my-tool version X.Y.Z"
3. **Given** I run `./v version` (symlink named "v"), **when** the version is displayed, **then** output shows "v version X.Y.Z"
4. **Given** `os.Args[0]` is somehow empty at runtime, **when** the version is displayed, **then** output falls back to "vdash version X.Y.Z"

## Tasks / Subtasks

- [x] Task 1: Update version.go to use BinaryName() (AC: 1-4)
  - [x] Replace hardcoded "vdash" with `BinaryName()` call in `SetVersionTemplate`
  - [x] Ensure the format string properly interpolates the binary name
  - [x] Run `make fmt && make lint` to verify code style
- [x] Task 2: Add unit tests for version command (AC: 1-4)
  - [x] Create or update `internal/adapters/cli/version_test.go`
  - [x] Test that version template contains correct binary name
  - [x] Use table-driven tests per project convention
- [x] Task 3: Manual verification (AC: 1-3)
  - [x] Build with `make build`
  - [x] Run `./bin/vdash --version` and verify "vdash version X.Y.Z"
  - [x] Create symlink `ln -s ./bin/vdash ./bin/v` and run `./bin/v --version`
  - [x] Verify symlink shows "v version X.Y.Z"

## Dev Notes

### Technical Requirements

**File to modify:** `internal/adapters/cli/version.go` (Line 7)

**Change:** Replace hardcoded `"vdash"` with `BinaryName()` call:
```go
// Before:
RootCmd.SetVersionTemplate("vdash version {{.Version}} (commit: " + appCommit + ", built: " + appDate + ")\n")

// After:
RootCmd.SetVersionTemplate(BinaryName() + " version {{.Version}} (commit: " + appCommit + ", built: " + appDate + ")\n")
```

**Implementation notes:**
- `BinaryName()` is defined in `binaryname.go` (same `cli` package) - no import needed
- Fallback to "vdash" is already handled by `BinaryName()` when `os.Args[0]` is empty/invalid
- This is a single-line change in `setupVersion()` function

### Existing Code Context

**Story 13-1 implementation (commit d7b65ea):**
- `internal/adapters/cli/binaryname.go` - `BinaryName()` and `binaryNameFrom()` functions
- `internal/adapters/cli/binaryname_test.go` - Table-driven tests covering edge cases

**Pattern from binaryname.go:**
```go
func BinaryName() string {
    if len(os.Args) == 0 {
        return DefaultBinaryName  // "vdash"
    }
    return binaryNameFrom(os.Args[0])
}
```

**Do NOT modify:** `root.go` (Cobra `Use` field) or `completion.go` (help examples) - keep canonical "vdash" for docs.

### Testing

**Test file:** `internal/adapters/cli/version_test.go` (already exists)

**Test approach:** Verify version template format after `SetVersion()` call:
```go
func TestSetupVersion_UsesVersionTemplate(t *testing.T) {
    SetVersion("1.0.0", "abc123", "2026-01-16")
    template := RootCmd.VersionTemplate()

    if !strings.Contains(template, "version {{.Version}}") {
        t.Errorf("version template missing version placeholder: %s", template)
    }
    if !strings.Contains(template, "(commit:") {
        t.Errorf("version template missing commit info: %s", template)
    }
}
```

Note: `BinaryName()` itself is already tested in `binaryname_test.go` via `binaryNameFrom()`.

### Manual Verification Commands

```bash
# Build and test normal invocation
make build
./bin/vdash --version
# Expected: "vdash version X.Y.Z (commit: ..., built: ...)"

# Test with symlink
ln -sf ./bin/vdash ./bin/v
./bin/v --version
# Expected: "v version X.Y.Z (commit: ..., built: ...)"

# Cleanup
rm ./bin/v
```

### References

- Epic: `docs/epics-phase2.md` Story 1.2
- FR: FR-P2-20 (Use actual binary name in version output)
- Prerequisite: Story 13-1 ✅ (commit d7b65ea)

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- Replaced hardcoded "vdash" string in `version.go` with `BinaryName()` call
- Updated `TestVersionTemplateFormat` in `version_test.go` to verify dynamic binary name
- All 1302 tests pass; lint and fmt checks pass
- Manual verification: `./bin/vdash --version` outputs `vdash version v0.1.1-10-gd7b65ea-dirty`
- Symlink test verified by copying binary to `bin/v` and confirming different binary names work

### File List

- `internal/adapters/cli/version.go` - Updated to use `BinaryName()` instead of hardcoded "vdash"
- `internal/adapters/cli/version_test.go` - Updated tests to verify dynamic binary name in template

