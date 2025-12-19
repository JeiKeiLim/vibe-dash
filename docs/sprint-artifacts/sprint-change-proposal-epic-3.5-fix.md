# Sprint Change Proposal: Epic 3.5 Storage Fix

**Date:** 2025-12-19
**Status:** APPROVED
**Priority:** CRITICAL
**Origin:** Post-implementation discovery - TUI not wired to new storage architecture

---

## Issue Summary

Epic 3.5 (Storage Structure Alignment) was marked as "done" but implementation is incomplete. The TUI dashboard still uses the OLD centralized database architecture while CLI commands correctly use the new per-project storage.

### Symptoms Reported
1. `~/.vibe-dash/projects.db` recreated after deletion
2. `vibe add .` works but TUI dashboard shows empty
3. `vibe list` shows projects correctly (CLI uses new architecture)

### Root Cause
`internal/adapters/cli/root.go:27` creates OLD `sqlite.NewSQLiteRepository("")` instead of using the injected `RepositoryCoordinator`.

```go
// BUG: Creates old centralized DB
repo, err := sqlite.NewSQLiteRepository("")
```

---

## Epic Impact

| Epic | Impact | Action |
|------|--------|--------|
| Epic 3.5 | **INCOMPLETE** | Reopen, add fix story |
| Epic 4 | BLOCKED | Cannot start until 3.5 complete |
| Epic 5+ | No impact | Proceed after Epic 4 |

---

## Artifact Adjustments

### Code Changes Required

| # | File | Action | Complexity |
|---|------|--------|------------|
| 1 | `internal/adapters/cli/root.go` | Fix: Use injected `repository` variable | LOW |
| 2 | `internal/adapters/persistence/sqlite/repository.go` | DELETE (legacy) | LOW |
| 3 | `internal/adapters/persistence/sqlite/repository_test.go` | DELETE (legacy) | LOW |
| 4 | `internal/adapters/tui/validation_test.go` | Fix: Use `project_repository` | LOW |

### Document Updates

| Document | Change |
|----------|--------|
| `sprint-status.yaml` | Revert Epic 3.5 status to `in-progress` |
| Epic 3.5 storage structure | Add Story 3.5.8 for this fix |

---

## Recommended Path Forward

**Option 1: Direct Adjustment** (SELECTED)

- Simple code fix in `root.go`
- Delete legacy `repository.go` and tests
- Update one TUI test file
- Total effort: ~1-2 hours including testing

---

## PRD MVP Impact

**None** - This is a bug fix to complete already-planned functionality. MVP scope unchanged.

---

## Action Plan

### Story 3.5.8: Fix TUI Repository Wiring

**As a** user,
**I want** the TUI dashboard to show my projects,
**So that** I can use vibe-dash as intended.

**Acceptance Criteria:**
```gherkin
AC1: Given projects added via `vibe add`
     When launching TUI with `vibe`
     Then dashboard shows all added projects

AC2: Given fresh installation (no ~/.vibe-dash/)
     When running `vibe add .` then `vibe`
     Then project appears in dashboard

AC3: Given ~/.vibe-dash/ exists
     When checking directory contents
     Then projects.db does NOT exist (only per-project state.db files)
```

**Tasks:**
- [ ] Update `root.go` to use injected `repository`
- [ ] Delete `repository.go` (legacy centralized DB)
- [ ] Delete `repository_test.go`
- [ ] Update `validation_test.go` to use `project_repository`
- [ ] Run full test suite
- [ ] Manual verification: `rm -rf ~/.vibe-dash && vibe add . && vibe`

---

## Agent Handoff

| Role | Responsibility |
|------|----------------|
| **SM (Bob)** | Update sprint-status.yaml, create story file |
| **Dev (Amelia)** | Implement code changes |
| **SM (Bob)** | Validate and close story |

---

## Approval

- [x] Change analysis complete
- [x] User approved change proposals
- [ ] Implementation pending

**Approved by:** Jongkuk Lim
**Date:** 2025-12-19
