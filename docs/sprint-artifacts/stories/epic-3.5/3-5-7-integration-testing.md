# Story 3.5.7: Integration Testing

Status: done

## Story

As a developer,
I want comprehensive integration tests for the new storage structure,
So that we can be confident the system works correctly before moving to Epic 4.

## Acceptance Criteria

1. **AC1: Fresh installation lifecycle test** - Given fresh installation, when adding projects through full lifecycle (add → list → remove → verify), then all operations work correctly with proper directory structure creation and cleanup.

2. **AC2: Name collision resolution test** - Given 3 projects with 2 name collisions (e.g., `/workspace-a/api`, `/workspace-b/api`, `/workspace-c/api`), when all are added, then each has unique directory with correct disambiguation (`api/`, `workspace-b-api/`, `workspace-c-api/`).

3. **AC3: Config cascade integration test** - Given project with custom config AND master config with global settings, when loading project settings, then cascade (project → master → default) works correctly for all config keys.

4. **AC4: Lazy loading performance test** - Given 20 projects tracked, when dashboard loads, then startup completes within acceptable time (demonstrates lazy loading works, no hardcoded timeout assertion).

5. **AC5: All existing tests pass** - Given refactored storage structure, when running `make test-all`, then all existing tests pass with any necessary updates.

6. **AC6: Coordinator integration mock fixed** - Given the `DeleteProjectDir` method added in Story 3.5.6, when running integration tests, then coordinator_integration_test.go compiles and passes.

## Tasks / Subtasks

- [x] Task 1: Fix existing coordinator integration test mock (AC: 6) ✅ COMPLETED DURING VALIDATION
  - [x] Subtask 1.1: Add `DeleteProjectDir` method to `integrationMockDirectoryManager` in `coordinator_integration_test.go`
  - [x] Subtask 1.2: Verify existing coordinator integration tests pass

- [x] Task 2: Verify existing lifecycle tests cover AC1 (AC: 1)
  - **EXISTING COVERAGE:** `coordinator_integration_test.go` already contains `TestIntegration_FullLifecycle` and `TestIntegration_SaveFindDeleteCycle`
  - [x] Subtask 2.1: Run `go test -tags=integration ./internal/adapters/persistence/...` and verify these tests pass
  - [x] Subtask 2.2: If additional full-stack test needed (with real ViperLoader + DirectoryManager), add `TestIntegration_FullStackLifecycle` to existing `coordinator_integration_test.go` (not a new file) - NOT NEEDED, existing coverage sufficient
  - [x] Subtask 2.3: Ensure test uses `//go:build integration` build tag
  - [x] Subtask 2.4: Use `t.TempDir()` for automatic cleanup

- [x] Task 3: Verify existing collision resolution tests cover AC2 (AC: 2)
  - **EXISTING COVERAGE:** `directory_integration_test.go` already contains `TestIntegration_EnsureProjectDir_CollisionResolution`
  - [x] Subtask 3.1: Run `go test -tags=integration ./internal/adapters/filesystem/...` and verify collision test passes
  - [x] Subtask 3.2: Added `TestIntegration_CollisionResolution_ThreeProjects` to existing `directory_integration_test.go` for 3-project collision test
  - [x] Subtask 3.3: Follow existing test pattern from `TestIntegration_EnsureProjectDir_CollisionResolution`

- [x] Task 4: Verify existing config cascade tests cover AC3 (AC: 3)
  - **EXISTING COVERAGE:** `project_config_integration_test.go` already contains `TestIntegration_ConfigResolver_WithRealConfigs` which tests hibernation override + waiting threshold fallback
  - [x] Subtask 4.1: Run `go test -tags=integration ./internal/config/...` and verify cascade test passes
  - [x] Subtask 4.2: Existing coverage sufficient with `TestConfigResolver_CascadeOrder` testing full cascade

- [x] Task 5: Verify existing lazy loading tests cover AC4 (AC: 4)
  - **EXISTING COVERAGE:** `coordinator_integration_test.go` already contains `TestIntegration_LazyLoading` which verifies cache behavior
  - [x] Subtask 5.1: Run existing test and verify it passes
  - [x] Subtask 5.2: Extended `TestIntegration_LazyLoading` to use 20 projects (per AC4 requirement)
  - [x] Subtask 5.3: Verify lazy loading by checking cache state (NOT timing assertions)

- [x] Task 6: Run full test suite and fix any failures (AC: 5)
  - [x] Subtask 6.1: Run `make test` - all unit tests pass
  - [x] Subtask 6.2: Run `make test-all` - all integration tests pass
  - [x] Subtask 6.3: Run `make lint` - fixed pre-existing errcheck issues in coordinator_test.go and directory_test.go
  - [x] Subtask 6.4: Error scenarios already covered by existing tests

- [x] Task 7: Final verification
  - [x] Subtask 7.1: Run `make test` - all unit tests pass
  - [x] Subtask 7.2: Run `make test-all` - all integration tests pass
  - [x] Subtask 7.3: Run `make lint` - no linting errors
  - [x] Subtask 7.4: Manual verification:
    1. Add project: `./bin/vibe add .` ✅
    2. List: `./bin/vibe list` ✅
    3. Remove: `./bin/vibe remove <name>` ✅

## Dev Notes

### ⚠️ IMPORTANT: Before Creating New Tests

**CHECK EXISTING TESTS FIRST!** Many acceptance criteria are already covered:
- AC1 (Lifecycle): `TestIntegration_FullLifecycle`, `TestIntegration_SaveFindDeleteCycle` in `coordinator_integration_test.go`
- AC2 (Collision): `TestIntegration_EnsureProjectDir_CollisionResolution` in `directory_integration_test.go`
- AC3 (Config cascade): `TestIntegration_ConfigResolver_WithRealConfigs` in `project_config_integration_test.go`
- AC4 (Lazy loading): `TestIntegration_LazyLoading` in `coordinator_integration_test.go`
- AC6 (Mock fix): ✅ **COMPLETED** - `integrationMockDirectoryManager` now implements `DeleteProjectDir`

### Architecture Overview

This story validates the entire Epic 3.5 storage structure by ensuring existing integration tests pass and adding any missing coverage. The tests exercise all components working together:

```
                    ┌─────────────────────────────────────────┐
                    │         Integration Test                │
                    │    (orchestrates full lifecycle)        │
                    └─────────────────┬───────────────────────┘
                                      │
        ┌─────────────────────────────┼─────────────────────────────┐
        │                             │                             │
        ▼                             ▼                             ▼
┌───────────────────┐   ┌─────────────────────────┐   ┌───────────────────────┐
│  ViperLoader      │   │  DirectoryManager       │   │ RepositoryCoordinator │
│  (master config)  │   │  (project directories)  │   │ (aggregates DBs)      │
└─────────┬─────────┘   └───────────┬─────────────┘   └───────────┬───────────┘
          │                         │                             │
          ▼                         ▼                             ▼
    config.yaml               ~/.vibe-dash/                  per-project
    (projects map)             <project>/                    state.db
                                .project-path
```

### Test File Organization

| File | Purpose | Status |
|------|---------|--------|
| `internal/adapters/persistence/coordinator_integration_test.go` | Lifecycle, lazy loading tests | ✅ EXISTS - Mock FIXED |
| `internal/adapters/filesystem/directory_integration_test.go` | Collision resolution tests | ✅ EXISTS |
| `internal/config/project_config_integration_test.go` | Config cascade tests | ✅ EXISTS |

**Note:** Do NOT create `storage_integration_test.go` unless existing tests are insufficient. Add new tests to existing files.

### Key Implementation Patterns

**Integration Test Build Tag (REQUIRED):**
```go
//go:build integration

package persistence
```

**Helper for setting up test infrastructure:**
```go
func setupIntegrationTest(t *testing.T) (basePath string, loader *config.ViperLoader, dirMgr ports.DirectoryManager, coord *RepositoryCoordinator) {
    t.Helper()
    basePath = t.TempDir()

    // Create master config file
    configPath := filepath.Join(basePath, "config.yaml")
    initialConfig := `storage_version: 2
settings:
  hibernation_days: 14
projects: {}
`
    os.WriteFile(configPath, []byte(initialConfig), 0644)

    loader = config.NewViperLoader(configPath)
    dirMgr = filesystem.NewDirectoryManager(basePath, &configPathAdapter{loader})
    coord = NewRepositoryCoordinator(loader, dirMgr, basePath)

    return basePath, loader, dirMgr, coord
}
```

**Mock Fix (COMPLETED):**

The `integrationMockDirectoryManager` in `coordinator_integration_test.go:425-448` has been updated to implement `DeleteProjectDir`. Pattern reference for similar mocks:

```go
// Pattern: Add deleteFunc field and method to mocks implementing DirectoryManager
deleteFunc func(ctx context.Context, projectPath string) error

func (m *MockDirectoryManager) DeleteProjectDir(ctx context.Context, projectPath string) error {
    if m.deleteFunc != nil { return m.deleteFunc(ctx, projectPath) }
    return nil
}
```

### Test Scenarios Detailed

**Scenario 1: Full Lifecycle Test**
```
1. Setup: t.TempDir() as base path
2. Create: ViperLoader, DirectoryManager, RepositoryCoordinator
3. Add: Create project domain object, call coordinator.Save()
4. Verify:
   - Directory exists at basePath/<project-name>/
   - state.db exists in directory
   - .project-path marker exists with correct path
   - Master config has project entry
5. List: Call coordinator.FindAll(), verify project in list
6. Remove: Call coordinator.Delete()
7. Verify:
   - FindByID returns ErrProjectNotFound
   - Master config no longer has project entry
   - Note: Physical directory cleanup is CLI's responsibility
```

**Scenario 2: Collision Resolution Test**
```
1. Create 3 fake project directories:
   - /tmp/workspace-a/api
   - /tmp/workspace-b/api
   - /tmp/workspace-c/api
2. Call EnsureProjectDir for each
3. Expected directories:
   - basePath/api/              (first wins)
   - basePath/workspace-b-api/  (parent disambiguation)
   - basePath/workspace-c-api/  (parent disambiguation)
4. Verify .project-path markers contain correct canonical paths
```

**Scenario 3: Config Cascade Test**
```
Master Config:
  settings:
    hibernation_days: 21
    agent_waiting_threshold_minutes: 12

Project Config:
  custom_hibernation_days: 7
  # no waiting threshold

Expected Results:
  GetEffectiveHibernationDays() -> 7 (project wins)
  GetEffectiveWaitingThreshold() -> 12 (master fallback)
```

**Scenario 4: Lazy Loading Test**
```
1. Setup: 20 project directories
2. Verify: Cache empty after coordinator creation
3. FindByID for one project
4. Verify: Cache has ~1 entry (only accessed repo)
5. FindAll
6. Verify: All repos eventually accessed for aggregation
7. Close
8. Verify: Cache cleared
9. FindAll again
10. Verify: Still works (lazy reload)
```

### Files Modified/Created

| File | Action | Purpose |
|------|--------|---------|
| `internal/adapters/persistence/coordinator_integration_test.go` | ✅ MODIFIED | Fixed mock to implement `DeleteProjectDir` |

### Existing Test Files (Already Adequate)

These files already have integration tests. Only modify if tests fail:
- `internal/adapters/filesystem/directory_integration_test.go` - Collision tests exist
- `internal/config/project_config_integration_test.go` - Config cascade tests exist
- `internal/adapters/persistence/coordinator_integration_test.go` - Lifecycle and lazy loading tests exist

### Dependencies

**Depends on (COMPLETED):**
- Story 3.5.1: `DirectoryManager` - Directory creation and collision handling
- Story 3.5.2: `sqlite.ProjectRepository` - Per-project database
- Story 3.5.3: Per-Project Config Files - Project config cascade
- Story 3.5.4: Master Config as Path Index - Project path mappings
- Story 3.5.5: `RepositoryCoordinator` - Multi-DB aggregation
- Story 3.5.6: CLI Commands Updated - Full wiring complete

**Required by:**
- Epic 4: Agent Waiting Detection - Must have stable storage foundation

### Previous Story Learnings (3.5.6)

From Story 3.5.6:
1. **DeleteProjectDir** was added to `ports.DirectoryManager` interface - mocks must implement it
2. **Context cancellation** pattern at start of every method
3. **Safety checks** for file operations (only delete within base path)
4. **Graceful degradation** - log warnings, continue on non-fatal errors

### Project Structure Notes

After this story, Epic 3.5 is complete:
```
~/.vibe-dash/
├── config.yaml                    # Master config with project index
├── api-service/
│   ├── .project-path              # Marker with canonical path
│   ├── config.yaml                # Project-specific settings
│   └── state.db                   # Per-project SQLite database
└── client-b-api-service/          # Disambiguated directory name
    ├── .project-path
    ├── config.yaml
    └── state.db
```

### Testing Strategy

**Unit Tests:** Already exist throughout codebase
**Integration Tests:** This story creates/fixes them
**Manual Testing:** Final verification checklist in Task 7.4

### Error Scenarios to Test

| Scenario | Expected Behavior |
|----------|------------------|
| Add project with invalid path | Return ErrPathNotAccessible |
| Add duplicate project | Return ErrProjectAlreadyExists |
| Remove non-existent project | Return ErrProjectNotFound |
| Load config with invalid YAML | Return ErrConfigInvalid with details |
| Permission denied on directory | Return ErrPathNotAccessible |

### References

- [Source: docs/sprint-artifacts/stories/epic-3.5/epic-3.5-storage-structure.md#Story 3.5.7]
- [Source: docs/architecture.md#Test Organization Patterns]
- [Source: docs/project-context.md#Testing Rules]
- [Source: internal/adapters/persistence/coordinator_integration_test.go] - Existing tests
- [Source: internal/adapters/filesystem/directory_integration_test.go] - Collision tests
- [Source: internal/config/project_config_integration_test.go] - Config cascade tests

## Dev Agent Record

### Context Reference

- Epic 3.5: Storage Structure Alignment
- Story Dependencies: Depends on 3.5.1-3.5.6 (all completed)
- PRD Reference: Lines 597-665 (Storage structure)
- Architecture Reference: Test Organization section
- Project Context: Testing rules, Go patterns

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- All 6 acceptance criteria verified and passing
- AC1: Lifecycle tests pass via `TestIntegration_FullLifecycle` and `TestIntegration_SaveFindDeleteCycle`
- AC2: Collision resolution test added `TestIntegration_CollisionResolution_ThreeProjects` for 3-project scenario
- AC3: Config cascade tests pass via `TestIntegration_ConfigResolver_WithRealConfigs` and `TestConfigResolver_CascadeOrder`
- AC4: Lazy loading test extended to 20 projects in `TestIntegration_LazyLoading`
- AC5: All tests pass (`make test`, `make test-all`, `make lint`)
- AC6: Mock fix was already completed during validation
- Fixed pre-existing lint errors in `coordinator_test.go` (errcheck) and `directory_test.go` (errcheck)
- Code review applied: Fixed error handling in lazy loading test setup loop, renamed test for consistency, added diagnostic logging

### File List

| File | Action | Purpose |
|------|--------|---------|
| `internal/adapters/filesystem/directory_integration_test.go` | MODIFIED | Added `TestIntegration_EnsureProjectDir_CollisionResolution_ThreeProjects` for AC2 |
| `internal/adapters/persistence/coordinator_integration_test.go` | MODIFIED | Extended `TestIntegration_LazyLoading` to 20 projects for AC4, fixed error handling |
| `internal/adapters/persistence/coordinator_test.go` | MODIFIED | Fixed errcheck lint errors |
| `internal/adapters/filesystem/directory_test.go` | MODIFIED | Fixed errcheck lint error in defer |
| `docs/sprint-artifacts/sprint-status.yaml` | MODIFIED | Updated story status |
