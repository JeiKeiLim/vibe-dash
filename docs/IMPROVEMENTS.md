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
