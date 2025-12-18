# Epic 3.5: Storage Structure Alignment

**Status:** In Progress
**Priority:** CRITICAL - Must complete before Epic 4
**Created:** 2025-12-18
**Origin:** Epic 3 Retrospective - Storage structure deviation from PRD discovered

---

## Epic Overview

### Problem Statement

During Epic 3 retrospective, a significant deviation was discovered between the PRD specification and the actual implementation:

**PRD Specifies (lines 597-665):**
```
~/.vibe-dash/
  ├── config.yaml                 # Master index (single source of truth)
  ├── api-service/
  │   ├── config.yaml             # Project-specific settings
  │   └── state.db                # Per-project SQLite database
  └── client-b-api-service/
      ├── config.yaml
      └── state.db
```

**Current Implementation:**
```
~/.vibe-dash/
  ├── config.yaml                 # Master config ✅
  └── projects.db                 # Single centralized DB ❌
```

### Why This Matters

1. **Data Isolation:** All projects currently share one DB - no isolation
2. **Project-specific Settings:** Cannot set per-project config files as specified
3. **Scalability:** Single DB may bottleneck with many projects
4. **Backup/Restore:** Cannot backup/restore individual project state
5. **Epic 4 Dependency:** Agent Waiting Detection will write more state data - building on wrong structure makes later migration harder

### Decision

**Full PRD Compliance (Option A)** - Clean implementation since project is pre-release.

- No migration needed - nuke existing config/DB and start fresh
- Per-project subdirectories with `state.db` each
- Per-project `config.yaml` files
- Master config as pure path index + global settings
- Collision handling for directory names per PRD spec

---

## Stories

### Story 3.5.0: Cleanup Existing Storage

**Status:** Backlog

**As a** developer,
**I want** the existing storage structure removed,
**So that** we can implement the correct PRD-compliant structure without migration complexity.

**Acceptance Criteria:**

```gherkin
AC1: Given existing ~/.vibe-dash/projects.db exists
     When cleanup is performed
     Then projects.db is deleted

AC2: Given existing ~/.vibe-dash/config.yaml exists
     When cleanup is performed
     Then config.yaml is deleted

AC3: Given cleanup is complete
     When running `vibe` or `vibe add`
     Then new structure is created per PRD spec
```

**Tasks:**
- [ ] Delete ~/.vibe-dash/projects.db
- [ ] Delete ~/.vibe-dash/config.yaml
- [ ] Update any test fixtures using old structure
- [ ] Document the breaking change

**Manual Testing:**
1. Check if ~/.vibe-dash/ exists and contains old files
2. Delete the directory: `rm -rf ~/.vibe-dash/`
3. Run `vibe` - should create new structure (after Story 3.5.6)

---

### Story 3.5.1: Directory Manager with Collision Handling

**Status:** Backlog

**As a** user,
**I want** each project to have its own subdirectory with collision handling,
**So that** projects with the same name from different locations are stored separately.

**Acceptance Criteria:**

```gherkin
AC1: Given project at /home/user/api-service
     When adding the project
     Then directory ~/.vibe-dash/api-service/ is created

AC2: Given project api-service already tracked
     And new project at /home/user/client-b/api-service
     When adding the new project
     Then directory ~/.vibe-dash/client-b-api-service/ is created

AC3: Given collision still exists after parent disambiguation
     When adding another project
     Then grandparent directory is added (work-client-b-api-service)

AC4: Given project path with symlinks
     When calculating directory name
     Then canonical path (via filepath.EvalSymlinks) is used

AC5: Given same project path added twice
     When calculating directory name
     Then same directory name is returned (deterministic)

AC6: Given directory creation fails (permission denied, disk full)
     When EnsureProjectDir is called
     Then descriptive error is returned with path and cause
```

**Technical Context:**
- PRD Reference: lines 647-659 (Resolution Algorithm)
- Location: `internal/adapters/filesystem/directory.go`
- Interface: `DirectoryManager` in `internal/core/ports/`

**Tasks:**
- [ ] Create `DirectoryManager` interface in `internal/core/ports/directory.go`
- [ ] Implement `GetProjectDirName(projectPath string) string` - deterministic directory name
- [ ] Implement `EnsureProjectDir(projectPath string) (string, error)` - creates directory
- [ ] Implement collision resolution algorithm per PRD
- [ ] Use `filepath.EvalSymlinks()` for canonical paths
- [ ] Write comprehensive unit tests with collision scenarios
- [ ] Test edge cases: deeply nested paths, special characters

**Manual Testing:**
1. Add project: `vibe add /path/to/api-service`
2. Verify: `ls ~/.vibe-dash/` shows `api-service/`
3. Add collision: `vibe add /other/path/api-service`
4. Verify: `ls ~/.vibe-dash/` shows both directories with disambiguation

---

### Story 3.5.2: Per-Project SQLite Repository

**Status:** Backlog

**As a** developer,
**I want** each project to have its own SQLite database,
**So that** project data is isolated and can be backed up independently.

**Acceptance Criteria:**

```gherkin
AC1: Given project directory exists
     When repository is created for project
     Then state.db is created at ~/.vibe-dash/<project>/state.db

AC2: Given project repository
     When saving project state
     Then data is written to project-specific state.db

AC3: Given project state.db
     When enabling WAL mode
     Then concurrent reads are supported

AC4: Given corrupted state.db
     When corruption is detected
     Then recovery suggestion includes project-specific path
```

**Technical Context:**
- Schema: Simplified for single-project storage (no project_id needed in most tables)
- Location: `internal/adapters/persistence/sqlite/project_repository.go`
- Existing: Modify/replace `repository.go`

**Tasks:**
- [ ] Create `NewProjectRepository(projectDir string) (*ProjectRepository, error)`
- [ ] Update schema for single-project storage (simplify if possible)
- [ ] Maintain WAL mode and busy timeout configuration
- [ ] Update corruption handling with project-specific paths
- [ ] Write unit tests for project-specific DB operations
- [ ] Ensure schema migrations work per-project

**Manual Testing:**
1. Add project: `vibe add .`
2. Verify: `ls ~/.vibe-dash/<project>/` shows `state.db`
3. Check DB: `sqlite3 ~/.vibe-dash/<project>/state.db ".tables"`

---

### Story 3.5.3: Per-Project Config Files

**Status:** Backlog

**As a** user,
**I want** project-specific settings in separate config files,
**So that** I can configure each project independently.

**Acceptance Criteria:**

```gherkin
AC1: Given project is added
     When project directory is created
     Then config.yaml is created at ~/.vibe-dash/<project>/config.yaml

AC2: Given project config exists
     When reading project settings
     Then values from project config are used

AC3: Given project config and master config both have hibernation_days
     When resolving effective setting
     Then project config value takes precedence

AC4: Given only master config has setting
     When resolving effective setting
     Then master config value is used as fallback

AC5: Given neither config has setting
     When resolving effective setting
     Then built-in default is used
```

**Technical Context:**
- PRD Reference: lines 624-636 (Configuration Priority)
- Priority: CLI flags → project config → master config → defaults
- Location: `internal/config/project_config.go`

**Per-Project Config Format:**
```yaml
# ~/.vibe-dash/<project>/config.yaml
detected_method: "speckit"
last_scanned: "2025-12-08T07:03:00Z"
custom_hibernation_days: 7          # Optional override
agent_waiting_threshold_minutes: 5  # Optional override
notes: "Main API service"           # Project notes
```

**Tasks:**
- [ ] Create `ProjectConfigLoader` for per-project config files
- [ ] Define per-project config schema (detected_method, overrides, notes)
- [ ] Implement config cascade: project → master → defaults
- [ ] Auto-create project config on `vibe add`
- [ ] Write unit tests for cascade behavior
- [ ] Test edge cases: missing file, invalid YAML, partial config

**Manual Testing:**
1. Add project: `vibe add .`
2. Verify: `cat ~/.vibe-dash/<project>/config.yaml` shows project config
3. Edit: Add `custom_hibernation_days: 7`
4. Run `vibe` and verify project uses custom setting

---

### Story 3.5.4: Master Config as Path Index

**Status:** Backlog

**As a** user,
**I want** the master config to serve as a project path index with global settings,
**So that** project mappings and defaults are centrally managed.

**Acceptance Criteria:**

```gherkin
AC1: Given projects are tracked
     When viewing master config
     Then each project has entry with path and directory mapping

AC2: Given project is added
     When master config is updated
     Then project entry includes path and favorite status

AC3: Given global settings in master config
     When no project-specific override exists
     Then global setting is used

AC4: Given storage_version field
     When reading config
     Then version 2 structure is expected
```

**Technical Context:**
- PRD Reference: lines 607-621 (Master Config format)
- Location: Modify `internal/config/loader.go`

**Master Config Format:**
```yaml
storage_version: 2

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects:
  api-service:                    # directory name
    path: "/home/user/api-service"
    favorite: false
  client-b-api-service:           # disambiguated directory name
    path: "/home/user/client-b/api-service"
    favorite: true
```

**Tasks:**
- [ ] Update master config schema to storage_version: 2
- [ ] Store project path → directory name mappings
- [ ] Keep favorite status in master config (cross-project concern)
- [ ] Remove per-project overrides from master (moved to project config)
- [ ] Backward compatibility: gracefully handle old format
- [ ] Write tests for new format

**Manual Testing:**
1. Add projects: `vibe add /path/one && vibe add /path/two`
2. View: `cat ~/.vibe-dash/config.yaml`
3. Verify: Projects section shows path mappings

---

### Story 3.5.5: Repository Coordinator

**Status:** Backlog

**As a** developer,
**I want** a unified interface over multiple per-project repositories,
**So that** the service layer doesn't need to change.

**Acceptance Criteria:**

```gherkin
AC1: Given multiple projects exist
     When calling FindAll()
     Then all projects from all DBs are aggregated

AC2: Given project to save
     When calling Save(project)
     Then data is written to correct project's DB

AC3: Given project to delete
     When calling Delete(id)
     Then project is removed from its DB

AC4: Given 20 projects tracked
     When dashboard loads
     Then connections are opened lazily (not all at once)

AC5: Given repository coordinator
     When service layer uses ProjectRepository interface
     Then no code changes needed in services
```

**Technical Context:**
- Implements existing `ports.ProjectRepository` interface
- Location: `internal/adapters/persistence/coordinator.go`
- Lazy-loading to prevent file handle exhaustion

**Tasks:**
- [ ] Create `RepositoryCoordinator` implementing `ports.ProjectRepository`
- [ ] Lazy-load per-project repositories on demand
- [ ] Implement `FindAll()` aggregating from all project DBs
- [ ] Implement `Save()` routing to correct project DB
- [ ] Implement `Delete()` removing from correct DB
- [ ] Manage connection lifecycle (open/close per operation)
- [ ] Implement max concurrent DB connections limit (e.g., 10) to prevent file handle exhaustion at scale
- [ ] Write integration tests with multiple projects
- [ ] Verify no service layer changes needed

**Manual Testing:**
1. Add 3 projects: `vibe add /p1 && vibe add /p2 && vibe add /p3`
2. Run `vibe list` - all projects shown
3. Run `vibe remove p1` - only p1 removed
4. Run `vibe` - dashboard shows p2 and p3

---

### Story 3.5.6: Update CLI Commands

**Status:** Backlog

**As a** user,
**I want** CLI commands to work with the new storage structure,
**So that** I can add, list, and remove projects as before.

**Acceptance Criteria:**

```gherkin
AC1: Given new storage structure
     When running `vibe add /path/to/project`
     Then project directory, state.db, and config.yaml are created

AC2: Given project exists
     When running `vibe remove <name>`
     Then entire project directory is deleted

AC3: Given multiple projects
     When running `vibe list`
     Then all projects are shown from new structure

AC4: Given project with collision
     When running `vibe add`
     Then collision is resolved automatically
```

**Technical Context:**
- Modify: `internal/adapters/cli/add.go`
- Modify: `internal/adapters/cli/remove.go`
- Modify: `cmd/vibe/main.go` (wiring)

**Tasks:**
- [ ] Update `vibe add` to create project directory structure
- [ ] Update `vibe remove` to delete entire project directory
- [ ] Update `main.go` to wire RepositoryCoordinator
- [ ] Update `main.go` to wire DirectoryManager
- [ ] Ensure all existing CLI tests pass
- [ ] Add new tests for directory creation/deletion

**Manual Testing:**
1. Fresh start: `rm -rf ~/.vibe-dash/`
2. Add: `vibe add .` - creates structure
3. List: `vibe list` - shows project
4. Remove: `vibe remove <name>` - cleans up completely
5. Verify: `ls ~/.vibe-dash/` - project directory gone

---

### Story 3.5.7: Integration Testing

**Status:** Backlog

**As a** developer,
**I want** comprehensive integration tests for the new storage structure,
**So that** we can be confident the system works correctly.

**Acceptance Criteria:**

```gherkin
AC1: Given fresh installation
     When adding projects through full lifecycle
     Then all operations work correctly

AC2: Given 3 projects with 2 name collisions
     When all are added
     Then each has unique directory with correct disambiguation

AC3: Given project with custom config
     When loading project settings
     Then cascade (project → master → default) works correctly

AC4: Given 20 projects
     When dashboard loads
     Then startup time is < 1 second (lazy loading works)
```

**Technical Context:**
- Location: `internal/adapters/persistence/integration_test.go`
- Build tag: `//go:build integration`

**Tasks:**
- [ ] Create integration test for full add → list → remove cycle
- [ ] Create integration test for collision handling (3 projects, 2 same name)
- [ ] Create integration test for config cascade
- [ ] Create performance test for 20 projects (verify lazy loading)
- [ ] Update existing TUI tests to work with new structure
- [ ] Run full test suite: `make test-all`

**Manual Testing:**
1. Run integration tests: `go test -tags=integration ./...`
2. Create 20 test projects and verify dashboard performance
3. Test collision scenario manually
4. Test config override manually

---

## Story Dependency Graph

```
3.5.0 Cleanup
    │
    ▼
3.5.1 Directory Manager
    │
    ├──► 3.5.2 Per-Project Repository
    │         │
    │         └──► 3.5.5 Repository Coordinator
    │                      │
    └──► 3.5.3 Per-Project Config ──────┤
              │                          │
              └──► 3.5.4 Master Config ──┼──► 3.5.6 CLI Updates
                                         │         │
                                         │         ▼
                                         └──► 3.5.7 Integration Tests
```

**Parallel Tracks:**
- Storage track: 3.5.1 → 3.5.2 → 3.5.5
- Config track: 3.5.1 → 3.5.3 → 3.5.4
- Both merge at 3.5.6

---

## Architecture References

| Component | PRD Reference | Architecture Reference |
|-----------|---------------|------------------------|
| Directory structure | lines 597-605 | Project Structure section |
| Collision handling | lines 647-659 | Not explicitly covered (gap) |
| Config cascade | lines 632-636 | Configuration System section |
| Per-project DB | lines 841, 1003 | Persistence section |
| Canonical paths | line 664 | Path Resolution section |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Test fixture updates needed | High | Low | Update fixtures in Story 3.5.7 |
| Service layer changes | Low | Medium | Coordinator implements same interface |
| Performance regression | Low | Medium | Lazy loading, integration tests |
| Config migration complexity | N/A | N/A | Clean start - no migration needed |

---

## Definition of Done

- [ ] All 8 stories completed and reviewed
- [ ] All existing tests pass (updated as needed)
- [ ] New integration tests for storage structure
- [ ] Manual testing of full lifecycle complete
- [ ] PRD storage specification fully implemented
- [ ] Architecture document updated with collision handling algorithm reference
- [ ] Ready for Epic 4 (Agent Waiting Detection)

---

## Change Log

| Date | Change |
|------|--------|
| 2025-12-18 | Epic created by Architect (Winston) based on Epic 3 retrospective findings |
| 2025-12-18 | **Story 3.5.0 COMPLETE:** Manual cleanup performed - `~/.vibe-dash/` deleted. Users with existing installations must run `rm -rf ~/.vibe-dash/` before upgrading. |
