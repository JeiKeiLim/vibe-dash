# Story 13.3: Update Error Messages to Use Dynamic Name

Status: done

## Story

As a user,
I want error messages to use the actual binary name I invoked,
so that suggested commands in errors are copy-pasteable.

## User-Visible Changes

- **Changed:** Error message "No projects tracked. Run 'vdash add .' to add one." now uses actual binary name (e.g., "Run 'v add .' to add one." when invoked as `v`)
- **Changed:** Suggestion "Remove favorite status first with: vdash favorite ..." now uses actual binary name

## Acceptance Criteria

1. **Given** I run `./v add /nonexistent` (binary renamed to "v"), **when** path validation fails, **then** error message uses "v" (not hardcoded "vdash")
2. **Given** error messages suggest usage examples, **when** displayed to user, **then** examples use actual binary name (e.g., "try: v add .")
3. **Given** I run `./v list` with no projects tracked, **when** suggestion is displayed, **then** shows "Run 'v add .' to add one." (not "vdash add")
4. **Given** I run `./v status --all` with no projects tracked, **when** suggestion is displayed, **then** shows "Run 'v add .' to add one." (not "vdash add")
5. **Given** I run `./v hibernate my-favorite-project` where project is favorited, **when** error is displayed, **then** shows "Remove favorite status first with: v favorite ... --off" (not "vdash favorite")

## DO NOT MODIFY

The following should NOT be changed in this story (keep canonical "vdash" for documentation):
- `internal/adapters/cli/root.go` - Cobra `Use` field and `Long` description
- `internal/adapters/cli/completion.go` - Help/installation examples
- Any `Long:` or `Example:` strings in command definitions - Keep "vdash" for docs
- Help text examples (visible via `--help`) - Keep canonical for documentation
- `internal/adapters/cli/binaryname.go` - DefaultBinaryName constant (this IS the fallback)

## Tasks / Subtasks

- [x] Task 1: Update list.go suggestion message (AC: 3)
  - [x] Replace hardcoded "vdash" with `BinaryName()` call on line 91
  - [x] Change: `"No projects tracked. Run 'vdash add .' to add one.\n"` → `"No projects tracked. Run '%s add .' to add one.\n", BinaryName()`
  - [x] Verify with `make fmt && make lint`
- [x] Task 2: Update status.go suggestion message (AC: 4)
  - [x] Replace hardcoded "vdash" on line 146 using same pattern as Task 1
  - [x] Verify with `make fmt && make lint`
- [x] Task 3: Update hibernate.go suggestion message (AC: 5)
  - [x] Replace hardcoded "vdash" with `BinaryName()` call on line 86
  - [x] Note: Argument order changes from `(identifier)` to `(BinaryName(), identifier)`
  - [x] Change: `"...vdash favorite %s --off\n", identifier` → `"...%s favorite %s --off\n", BinaryName(), identifier`
  - [x] Verify with `make fmt && make lint`
- [x] Task 4: Update test assertions (AC: 3-5)
  - [x] Update `list_test.go:244` - change check from `"vdash add"` to pattern-based check
  - [x] Update `hibernate_test.go:249` - change check from `"vdash favorite test-project --off"` to pattern-based check
  - [x] See Testing Strategy section for recommended approach
- [x] Task 5: Manual verification (AC: 1-5)
  - [x] Build with `make build`
  - [x] Run `./bin/vdash list` (with no projects) - verify "vdash add ." suggestion
  - [x] Create symlink: `ln -sf ./bin/vdash ./bin/v`
  - [x] Run `./bin/v list` - verify "v add ." suggestion (not "vdash add")
  - [x] Run `./bin/v hibernate my-favorite-project` - verify "v favorite ..." in error (requires favorite project setup)
  - [x] Run `make test` to verify all tests pass
  - [x] Cleanup: `rm ./bin/v`

## Dev Notes

### Implementation Pattern (Single Example)

All three files use the same pattern. Here's the canonical example:

```go
// list.go line 91 - Before:
fmt.Fprintf(cmd.OutOrStdout(), "No projects tracked. Run 'vdash add .' to add one.\n")

// After:
fmt.Fprintf(cmd.OutOrStdout(), "No projects tracked. Run '%s add .' to add one.\n", BinaryName())
```

**Apply to:** `status.go:146` (identical change), `hibernate.go:86` (note: format string has two `%s` args after change)

**No import needed:** `BinaryName()` is in the same `cli` package.

### Files and Locations Summary

| File | Line | Change Required |
|------|------|-----------------|
| `internal/adapters/cli/list.go` | 91 | Replace `'vdash add .'` with `'%s add .', BinaryName()` |
| `internal/adapters/cli/status.go` | 146 | Same pattern as list.go |
| `internal/adapters/cli/hibernate.go` | 86 | `"...vdash favorite %s --off", identifier` → `"...%s favorite %s --off", BinaryName(), identifier` |
| `internal/adapters/cli/list_test.go` | 244 | Update assertion (see Testing Strategy) |
| `internal/adapters/cli/hibernate_test.go` | 249 | Update assertion (see Testing Strategy) |

### Testing Strategy

**Problem:** During `go test`, `os.Args[0]` is the test binary name (e.g., `list.test`), NOT "vdash". Tests that check for exact string matches will fail.

**Solution:** Use pattern-based assertions that match the dynamic portion without the binary name prefix:

```go
// list_test.go:244 - Before:
if !strings.Contains(output, "vdash add") {

// After - Option A (flexible pattern):
if !strings.Contains(output, " add .'") {
    t.Errorf("expected 'add .' suggestion, got: %s", output)
}

// After - Option B (use BinaryName() directly):
expected := fmt.Sprintf("Run '%s add .' to add one.", BinaryName())
if !strings.Contains(output, expected) {
    t.Errorf("expected suggestion with binary name, got: %s", output)
}
```

```go
// hibernate_test.go:249 - Before:
if !strings.Contains(output, "vdash favorite test-project --off") {

// After:
if !strings.Contains(output, "favorite test-project --off") {
    t.Errorf("expected favorite hint, got: %s", output)
}
```

### AC1 Clarification

AC1 mentions "path validation fails" showing "v: path not found" - this scenario is covered by Cobra's built-in error handling which already uses the command name dynamically. No additional code changes needed for AC1 beyond the three identified files.

### Existing Code Context

**Prerequisite implementations:**
- Story 13-1 (commit d7b65ea): `binaryname.go` with `BinaryName()` function
- Story 13-2 (commit db7d9f9): `version.go` uses same pattern

**BinaryName() behavior:**
- Returns `filepath.Base(os.Args[0])` (e.g., "v" for symlink `./bin/v`)
- Falls back to "vdash" if `os.Args[0]` is empty or invalid

### References

- Epic: `docs/epics-phase2.md` Story 1.3
- FR Coverage: FR-P2-19 (Use actual binary name in error messages)
- Prerequisite: Story 13-1 (commit d7b65ea) - BinaryName() utility
- Pattern reference: Story 13-2 (commit db7d9f9) - version.go implementation

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

### Completion Notes List

- Task 1-3: Updated `list.go:91`, `status.go:146`, `hibernate.go:86` to use `BinaryName()` instead of hardcoded "vdash"
- Task 4: Updated test assertions in `list_test.go:244` and `hibernate_test.go:249` to use pattern-based checks that work with dynamic binary names
- All 1302 tests pass
- Code formatted and linted successfully

### File List

- `internal/adapters/cli/list.go` - Updated suggestion message to use BinaryName()
- `internal/adapters/cli/status.go` - Updated suggestion message to use BinaryName()
- `internal/adapters/cli/hibernate.go` - Updated favorite removal hint to use BinaryName()
- `internal/adapters/cli/list_test.go` - Updated assertion to pattern-based check
- `internal/adapters/cli/hibernate_test.go` - Updated assertion to pattern-based check
- `internal/adapters/cli/status_test.go` - Added AC4 assertion for dynamic binary name in empty list hint

