# Story 3.5.0: Cleanup Existing Storage

Status: done

## Story

As a developer,
I want the existing storage structure removed,
So that we can implement the correct PRD-compliant structure without migration complexity.

## Acceptance Criteria

1. **AC1: Delete projects.db** - Given existing `~/.vibe-dash/projects.db` exists, when cleanup is performed, then `projects.db` is deleted.

2. **AC2: Delete config.yaml** - Given existing `~/.vibe-dash/config.yaml` exists, when cleanup is performed, then `config.yaml` is deleted.

3. **AC3: New structure on restart** - Given cleanup is complete, when running `vibe` or `vibe add`, then new structure is created per PRD spec (after Story 3.5.6).

## Tasks / Subtasks

- [x] Task 1: Delete existing storage files (AC: 1, 2)
  - [x] Subtask 1.1: Manually delete `~/.vibe-dash/projects.db` if it exists
  - [x] Subtask 1.2: Manually delete `~/.vibe-dash/config.yaml` if it exists
  - [x] Subtask 1.3: Optionally delete entire `~/.vibe-dash/` directory for clean start

- [x] Task 2: Update test fixtures that use old structure (AC: 3)
  - [x] Subtask 2.1: Identify test files that create/use `projects.db` directly
  - [x] Subtask 2.2: Mark tests that will need updates after new structure is implemented
  - [x] Subtask 2.3: Add TODO comments in affected test files using this format:
    ```go
    // TODO(story:3.5.7): Update test fixture to use per-project storage structure at ~/.vibe-dash/<project>/state.db
    ```

- [x] Task 3: Document the breaking change (AC: 3)
  - [x] Subtask 3.1: Add note to epic file about manual cleanup required
  - [x] Subtask 3.2: Update sprint-status.yaml notes if applicable

## Dev Notes

### Purpose of This Story

This is a **preparation story** for Epic 3.5. The current implementation uses:
- Single centralized database: `~/.vibe-dash/projects.db`
- Single config file: `~/.vibe-dash/config.yaml`

The PRD specifies per-project storage:
```
~/.vibe-dash/
  ├── config.yaml                 # Master index only
  ├── api-service/
  │   ├── config.yaml             # Project-specific settings
  │   └── state.db                # Per-project SQLite database
  └── client-b-api-service/
      ├── config.yaml
      └── state.db
```

### Why Cleanup Over Migration

Since vibe-dash is **pre-release** with no production users:
- Migration logic is unnecessary complexity
- Clean start is simpler and less error-prone
- No data loss concerns (test/development data only)

### Current Implementation Analysis

**Files to Delete:**
1. `~/.vibe-dash/projects.db` - Single SQLite database containing all projects
2. `~/.vibe-dash/config.yaml` - Master config with embedded project settings

**Repository Implementation:** [Source: internal/adapters/persistence/sqlite/repository.go:47-53]
```go
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
    if dbPath == "" {
        home, err := os.UserHomeDir()
        // ...
        dbPath = filepath.Join(home, ".vibe-dash", "projects.db")
    }
    // ...
}
```

**Config Implementation:** [Source: internal/config/loader.go:23-36]
```go
func NewViperLoader(configPath string) *ViperLoader {
    if configPath == "" {
        configPath = GetDefaultConfigPath()  // Returns ~/.vibe-dash/config.yaml
    }
    // ...
}
```

### Project Structure Notes

- This is a **manual cleanup story** - no code changes required
- Subsequent stories (3.5.1 - 3.5.7) will implement the new structure
- Tests using old paths will temporarily fail until Story 3.5.6 completes

### Test Files Affected

Tests that directly reference old storage structure:
- `internal/adapters/persistence/sqlite/repository_test.go` - Uses temporary DB but references structure
- `internal/config/loader_test.go` - Tests config loading/saving
- `internal/config/defaults_test.go` - Tests default config values

These tests will be updated as part of Story 3.5.7 (Integration Testing).

### References

- [Source: docs/prd.md#Configuration-Schema, lines 597-665] - PRD storage specification
- [Source: docs/sprint-artifacts/stories/epic-3.5/epic-3.5-storage-structure.md] - Epic 3.5 overview
- [Source: docs/architecture.md:1033-1035] - Current configuration file locations
- [Source: docs/sprint-artifacts/retrospectives/epic-3-retrospective.md] - Discovery of storage deviation

### Manual Testing

1. Check if `~/.vibe-dash/` exists:
   ```bash
   ls -la ~/.vibe-dash/
   ```

2. Delete the directory entirely:
   ```bash
   rm -rf ~/.vibe-dash/
   ```

3. Verify cleanup:
   ```bash
   ls ~/.vibe-dash/  # Should return "No such file or directory"
   ```

4. **IMPORTANT:** After Story 3.5.6 is complete, run `vibe` to verify new structure is created correctly.

### Verification Checklist

Use this checklist to verify story completion:

- [x] `~/.vibe-dash/projects.db` deleted (or confirmed not present)
- [x] `~/.vibe-dash/config.yaml` deleted (or confirmed not present)
- [x] (Optional) `~/.vibe-dash/` directory fully removed for clean start
- [x] TODO comments added to `internal/adapters/persistence/sqlite/repository_test.go`
- [x] TODO comments added to `internal/config/loader_test.go`
- [x] TODO comments added to `internal/config/defaults_test.go` (if applicable)
- [x] Epic file notes updated about manual cleanup
- [x] Sprint-status.yaml notes updated (if applicable)

## Dev Agent Record

### Context Reference

- Epic 3.5: Storage Structure Alignment
- Discovered during: Epic 3 Retrospective
- Blocking: Epic 4 (Agent Waiting Detection)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

- Deleted entire `~/.vibe-dash/` directory (including `projects.db` and `config.yaml`)
- Added TODO(story:3.5.7) comments to 3 test files for future update
- Documented breaking change in epic file change log
- Updated sprint-status.yaml to mark story as done
- All existing tests still pass after adding TODO comments

### File List

Files deleted (manual action via Dev Agent):
- `~/.vibe-dash/projects.db`
- `~/.vibe-dash/config.yaml`
- `~/.vibe-dash/` (entire directory)

Files modified:
- `internal/adapters/persistence/sqlite/repository_test.go` - Added TODO comment
- `internal/config/loader_test.go` - Added TODO comment
- `internal/config/defaults_test.go` - Added TODO comment
- `docs/sprint-artifacts/stories/epic-3.5/epic-3.5-storage-structure.md` - Added change log entry
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status to done
