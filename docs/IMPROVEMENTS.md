# Technical Improvements Backlog

Future improvements and refactoring ideas that are not urgent but worth tracking.

---

## CLI: Use Dynamic Binary Name Instead of Hardcoded "vdash"

**Added**: 2026-01-13
**Priority**: Low
**Category**: Code Quality / Maintainability

### Current State
The binary name "vdash" is hardcoded in multiple places:
- `internal/adapters/cli/version.go` - version template
- `internal/adapters/cli/*.go` - help text examples (e.g., `vdash add .`)
- `internal/adapters/cli/completion.go` - completion script instructions

### Proposed Change
Use `RootCmd.Use` or `os.Args[0]` dynamically instead of hardcoding:

```go
// version.go - use command name dynamically
func setupVersion() {
    RootCmd.Version = appVersion
    RootCmd.SetVersionTemplate(RootCmd.Use + " version {{.Version}} (commit: " + appCommit + ", built: " + appDate + ")\n")
}
```

### Considerations
- **Version output**: Should definitely use dynamic name
- **Help text examples**: Debatable - hardcoded canonical name may be preferred for documentation consistency
- **Completion scripts**: Already uses `RootCmd.Use` via Cobra

### Impact
If we rename the binary in the future, only `RootCmd.Use` needs to change instead of searching/replacing across many files.

---

## Historical Note: Binary Renamed from `vibe` to `vdash`

**Date**: 2026-01-13 (v0.1.0 release)

The binary was renamed from `vibe` to `vdash` to avoid confusion and provide a more distinctive name.

**What was updated:**
- All user-visible CLI help text, examples, and error messages
- Build scripts (Makefile, GitHub Actions)
- Integration tests
- README and CLAUDE.md

**What was NOT updated (intentionally):**
- Internal design docs (`docs/prd.md`, `docs/project-context.md`) - left as historical artifacts
- Internal variable names (`vibeHome`) - no user impact
- Comments referencing the original design
- Temp directory names in tests

If you see `vibe` in older docs, it refers to this project before the rename.

---
