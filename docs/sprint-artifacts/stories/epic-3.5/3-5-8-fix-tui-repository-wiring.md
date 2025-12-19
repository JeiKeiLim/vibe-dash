# Story 3.5.8: Fix TUI Repository Wiring

**Status:** Done
**Priority:** CRITICAL
**Created:** 2025-12-19
**Origin:** Sprint Change Proposal - Post-implementation discovery

---

## User Story

**As a** user,
**I want** the TUI dashboard to show my projects,
**So that** I can use vibe-dash as intended.

---

## Problem Statement

The TUI dashboard (`vibe` command with no arguments) still uses the OLD centralized `projects.db` instead of the new per-project storage architecture implemented in Epic 3.5.

**Root Cause:** `internal/adapters/cli/root.go:27` creates `sqlite.NewSQLiteRepository("")` instead of using the injected `RepositoryCoordinator`.

**Impact:**
- `projects.db` recreated on every TUI launch
- TUI shows empty dashboard (reads from wrong DB)
- `vibe list` works correctly (uses coordinator)

---

## Acceptance Criteria

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

AC4: Given all tests run
     When executing `go test ./...`
     Then all tests pass

AC5: Given legacy repository code
     When checking codebase
     Then repository.go and repository_test.go are deleted
```

---

## Technical Tasks

### Task 1: Fix root.go TUI Wiring

**File:** `internal/adapters/cli/root.go`

**Current (BROKEN):**
```go
package cli

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/sqlite"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui"
)

var RootCmd = &cobra.Command{
	Use:   "vibe",
	Short: "CLI dashboard for vibe coding projects",
	// ... Long description ...
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("vibe-dash starting")

		// BUG: Creates OLD centralized DB
		repo, err := sqlite.NewSQLiteRepository("")
		if err != nil {
			slog.Error("Failed to initialize repository", "error", err)
			return
		}

		if err := tui.Run(cmd.Context(), repo, detectionService); err != nil {
			slog.Error("TUI error", "error", err)
		}
	},
}
```

**Fixed:**
```go
package cli

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui"
)

var RootCmd = &cobra.Command{
	Use:   "vibe",
	Short: "CLI dashboard for vibe coding projects",
	// ... Long description ...
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("vibe-dash starting")

		// Use package-level repository (injected via SetRepository in main.go)
		if repository == nil {
			slog.Error("Repository not initialized")
			return
		}

		if err := tui.Run(cmd.Context(), repository, detectionService); err != nil {
			slog.Error("TUI error", "error", err)
		}
	},
}
```

**Changes:**
- Remove import `"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/sqlite"`
- Replace `sqlite.NewSQLiteRepository("")` with `repository` (package variable from add.go)
- Add nil check for safety

---

### Task 2: Delete Legacy Repository

**Files to DELETE:**
- `internal/adapters/persistence/sqlite/repository.go`
- `internal/adapters/persistence/sqlite/repository_test.go`

**Rationale:**
- No production code uses these after Task 1 fix
- `project_repository.go` provides equivalent functionality
- No backward compatibility concerns (zero users)

---

### Task 3: Fix TUI Validation Test

**File:** `internal/adapters/tui/validation_test.go`

**Current (line 22):**
```go
repo, err := sqlite.NewSQLiteRepository(dbPath)
```

**Fixed:**
```go
// Create temp directory with .project-path marker
markerPath := filepath.Join(filepath.Dir(dbPath), ".project-path")
os.WriteFile(markerPath, []byte("/test/path"), 0644)

repo, err := sqlite.NewProjectRepository(filepath.Dir(dbPath))
```

**Note:** The test needs to create the `.project-path` marker file since `NewProjectRepository` requires it.

---

## Testing Checklist

- [x] `go test ./...` - All unit tests pass
- [x] `go build ./cmd/vibe` - Binary builds successfully
- [x] Manual test: `rm -rf ~/.vibe-dash && ./bin/vibe add . && ./bin/vibe`
- [x] Verify: `ls ~/.vibe-dash/` shows NO `projects.db`
- [x] Verify: `vibe list` shows the added project

---

## Definition of Done

- [x] `root.go` uses injected `repository` instead of creating new one
- [x] `repository.go` deleted
- [x] `repository_test.go` deleted
- [x] `validation_test.go` updated to use `NewProjectRepository`
- [x] All tests pass
- [x] Manual verification complete
- [x] `projects.db` no longer created

---

## Dev Agent Record

### Implementation Plan
1. Fix `root.go` to use injected repository instead of creating new OLD repository
2. Delete legacy `repository.go` and `repository_test.go`
3. Update `validation_test.go` to use `NewProjectRepository` with `.project-path` marker
4. Create `helpers.go` to extract shared functions from deleted `repository.go`

### Completion Notes
- Fixed TUI not showing projects by using injected `repository` variable in `root.go`
- Deleted legacy `repository.go` and `repository_test.go` that created `projects.db`
- Updated `validation_test.go` to use `NewProjectRepository` with required `.project-path` marker
- Created `helpers.go` with shared functions (`ErrDatabaseCorrupted`, `projectRow`, `rowToProject`, `nullString`, `boolToInt`, `stateToString`, `wrapDBError`)
- All tests pass, manual verification confirms no `projects.db` created and `vibe list` works

---

## File List

**Modified:**
- `internal/adapters/cli/root.go` - Use injected repository instead of creating new one

**Deleted:**
- `internal/adapters/persistence/sqlite/repository.go` - Legacy centralized DB code
- `internal/adapters/persistence/sqlite/repository_test.go` - Legacy tests

**Created:**
- `internal/adapters/persistence/sqlite/helpers.go` - Shared helper functions extracted from deleted repository.go

**Updated:**
- `internal/adapters/tui/validation_test.go` - Use NewProjectRepository with .project-path marker

---

## Change Log

| Date | Change |
|------|--------|
| 2025-12-19 | Story created from Sprint Change Proposal |
| 2025-12-19 | Implementation complete - All tasks done, tests pass, manual verification complete |
| 2025-12-19 | Code Review: TUI working correctly, identified .project-path redundancy - created Story 3.5.9 |

---

## Dependencies

- Story 3.5.5: Repository Coordinator (provides the coordinator)
- Story 3.5.6: Update CLI Commands (wired coordinator to CLI)

---

## References

- Sprint Change Proposal: `docs/sprint-artifacts/sprint-change-proposal-epic-3.5-fix.md`
- Architecture: `docs/architecture.md` (Storage section)
- PRD: lines 597-665 (Storage structure specification)
