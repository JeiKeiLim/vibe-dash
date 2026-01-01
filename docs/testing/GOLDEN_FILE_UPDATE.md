# Golden File Update Workflow

This document describes how to update golden files for TUI behavioral tests in vibe-dash.

## When to Update Golden Files

Update golden files when:
1. **Intentional layout changes** - UI redesigns, column width adjustments, etc.
2. **New features** - Adding new elements that appear in test output
3. **Bug fixes** - When the fix changes expected output

**DO NOT** update golden files when:
1. Tests are failing unexpectedly (investigate the cause first)
2. Changes are unintentional (the test caught a regression)
3. Running in CI (only developers should update golden files locally)

## Update Workflow

### Step 1: Verify the Change is Intentional

Before updating, ensure you understand why the golden file test is failing:

```bash
# Run the specific failing test
go test -tags=integration -v ./internal/adapters/tui/... -run 'TestLayout_'
```

Review the diff output carefully. Ask yourself:
- Is this change expected given my code changes?
- Does the new output look correct?

### Step 2: Update the Golden File

Use the `-update` flag to regenerate the golden file:

```bash
# Update specific test's golden file
go test -tags=integration -run 'TestSpecificName' ./internal/adapters/tui/... -update

# Update all anchor point tests
go test -tags=integration -run 'TestAnchor_' ./internal/adapters/tui/... -update

# Update all layout tests
go test -tags=integration -run 'TestLayout_' ./internal/adapters/tui/... -update
```

**CAUTION:** Avoid updating all golden files at once (`-update` on full suite) - this may hide unintended changes.

### Step 3: Review the Updated File

```bash
# See what changed
git diff internal/adapters/tui/testdata/
```

Verify the diff matches your expectations. Each golden file change should be traceable to a specific code change.

### Step 4: Commit with Descriptive Message

```bash
git add internal/adapters/tui/testdata/*.golden
git commit -m "test: Update golden files for [description]

[Explain why the golden files changed]"
```

## Golden File Locations

| Test Category | Location |
|--------------|----------|
| Anchor stability | `internal/adapters/tui/testdata/anchor/*.golden` |
| Layout consistency | `internal/adapters/tui/testdata/layout/*.golden` |

## Troubleshooting

### "Golden file mismatch" but output looks correct

Check that your local environment matches CI:
```bash
# Ensure deterministic output
export NO_COLOR=1
export FORCE_COLOR=0
export TERM=dumb
```

### Golden files differ between macOS and Linux

Ensure `lipgloss.SetColorProfile(termenv.Ascii)` is called in test setup. See `teatest_helpers_test.go` line 250.

### Update doesn't take effect

1. Delete the old golden file first: `rm internal/adapters/tui/testdata/layout/TestName.golden`
2. Run test with `-update`
3. Verify new file was created

## CI Behavior

- CI **never** updates golden files
- CI fails if golden file content doesn't match test output
- The error message shows the diff between expected and actual output
- Contributors must update golden files locally before pushing

## Git Configuration

The `.gitattributes` file configures golden file handling:

```
*.golden -text               # Preserve exact bytes (no CRLF conversion)
*.golden linguist-generated=true  # Hide diffs in GitHub PR by default
```

This means:
1. Golden files are stored as-is (no line ending normalization)
2. In GitHub PRs, golden file changes are collapsed by default (click to expand)

## Related Documentation

- [TUI Testing Research](./tui-testing-research.md) - Overall testing strategy
- [Story 9.3](../../sprint-artifacts/stories/epic-9/9-3-anchor-point-stability-tests.md) - Anchor tests
- [Story 9.4](../../sprint-artifacts/stories/epic-9/9-4-layout-consistency-tests.md) - Layout tests
