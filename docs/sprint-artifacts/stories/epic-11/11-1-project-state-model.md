# Story 11.1: Project State Model

Status: Done

## Story

As a **developer**,
I want **a well-defined project state machine implemented**,
so that **state transitions between Active and Hibernated are explicit, validated, and tracked**.

## User-Visible Changes

None - this is an internal domain model enhancement. State machine logic will be used by subsequent stories (11.2-11.6) to expose user-facing hibernation features.

## Acceptance Criteria

1. **AC1: Verify existing state enum (no code changes)**
   - Confirm `ProjectState` enum has `StateActive`, `StateHibernated` in `internal/core/domain/state.go`
   - Confirm zero value is `StateActive`
   - Verification only - existing implementation is correct

2. **AC2: State transitions are well-defined**
   - Transition: `Active → Hibernated` (via auto-hibernate or manual)
   - Transition: `Hibernated → Active` (via auto-promote or manual)
   - Invalid transitions rejected with `ErrInvalidStateTransition`

3. **AC3: IsFavorite is independent of State**
   - `Favorite + Active`: Always visible, operates normally
   - `Favorite + Hibernated`: Still hibernated but marked as favorite
   - Favorites never auto-hibernate (FR30)

4. **AC4: State change triggers**
   - Database update via repository (`UpdateState` + `HibernatedAt`)
   - `UpdatedAt` timestamp updated
   - Wire-up to TUI deferred to Story 11.5

5. **AC5: StateService implements state transition logic**
   - `Hibernate(ctx, projectID)` - transitions Active → Hibernated, sets HibernatedAt
   - `Activate(ctx, projectID)` - transitions Hibernated → Active, clears HibernatedAt
   - Both reject invalid transitions with `ErrInvalidStateTransition`

6. **AC6: HibernatedAt timestamp tracked**
   - `Project` entity has `HibernatedAt *time.Time` field
   - Set to `time.Now()` when transitioning to Hibernated
   - Set to `nil` when transitioning to Active
   - Persisted via migration v3

7. **AC7: 95%+ test coverage for state logic**
   - Unit tests for valid transitions (Active→Hibernated, Hibernated→Active)
   - Unit tests for invalid transition rejection (Active→Active, Hibernated→Hibernated)
   - Unit tests for favorite+state independence (favorite blocks auto-hibernate)

## Tasks / Subtasks

- [x] Task 1: Enhance Project entity with HibernatedAt (AC: #6)
  - [x] 1.1: Add `HibernatedAt *time.Time` field to `domain.Project` (project.go:27)
  - [x] 1.2: Update `NewProject()` to leave `HibernatedAt` as nil (already default)
  - [x] 1.3: Add `DaysSinceHibernated() int` helper method (returns 0 if nil)

- [x] Task 2: Create StateService (AC: #2, #5)
  - [x] 2.1: Create `internal/core/services/state_service.go`
  - [x] 2.2: Implement `NewStateService(repo ports.ProjectRepository)` constructor
  - [x] 2.3: Implement `Hibernate(ctx, projectID string) error` method
  - [x] 2.4: Implement `Activate(ctx, projectID string) error` method
  - [x] 2.5: Add `ErrInvalidStateTransition` to `domain/errors.go`

- [x] Task 3: Add state transition validation (AC: #2, #3)
  - [x] 3.1: `Hibernate()` returns `ErrInvalidStateTransition` if already hibernated
  - [x] 3.2: `Activate()` returns `ErrInvalidStateTransition` if already active
  - [x] 3.3: `Hibernate()` returns error if `project.IsFavorite == true` (FR30)
  - [x] 3.4: State transitions update `UpdatedAt` via repository

- [x] Task 4: Database schema update (AC: #6)
  - [x] 4.1: Add migration v3 to `migrations.go` with `ALTER TABLE projects ADD COLUMN hibernated_at TEXT;`
  - [x] 4.2: Update `schema.go` SchemaVersion to 3
  - [x] 4.3: Add `HibernatedAt sql.NullString` field to `projectRow` struct in `helpers.go`
  - [x] 4.4: Update `rowToProject()` in `helpers.go` to parse HibernatedAt
  - [x] 4.5: Update `projectColumns` in `queries.go` to include `hibernated_at`
  - [x] 4.6: Update `insertOrReplaceProjectSQL` parameter count (15 → 16)
  - [x] 4.7: Update `Save()` in `project_repository.go` to persist HibernatedAt

- [x] Task 5: Write comprehensive tests (AC: #7)
  - [x] 5.1: Unit tests for `Hibernate()` valid transition (Active→Hibernated)
  - [x] 5.2: Unit tests for `Activate()` valid transition (Hibernated→Active)
  - [x] 5.3: Unit tests for `Hibernate()` rejection when already hibernated
  - [x] 5.4: Unit tests for `Activate()` rejection when already active
  - [x] 5.5: Unit tests for `Hibernate()` rejection when favorite
  - [x] 5.6: Integration tests for HibernatedAt persistence round-trip

## Dev Notes

### What Needs Creation (This Story's Scope)

| Item | Action | Location |
|------|--------|----------|
| `StateService` | CREATE | `internal/core/services/state_service.go` |
| `ErrInvalidStateTransition` | ADD | `internal/core/domain/errors.go` |
| `HibernatedAt` field | ADD | `internal/core/domain/project.go:27` |
| `DaysSinceHibernated()` | ADD | `internal/core/domain/project.go` |
| Migration v3 | ADD | `internal/adapters/persistence/sqlite/migrations.go` |

### Key Design Decisions

1. **StateService vs extending ProjectService**: Separate StateService for single responsibility
2. **HibernatedAt as pointer**: `*time.Time` represents nullable (nil = active)
3. **Transition errors**: Return `ErrInvalidStateTransition` with context, not silent no-op
4. **Favorite guard**: Error type `ErrFavoriteCannotHibernate` or wrapped `ErrInvalidStateTransition`

### Architecture Compliance

```
internal/core/services/state_service.go  ←  NEW: state transition logic
         ↓ calls
internal/core/ports/repository.go        ←  EXISTING: ProjectRepository.Save()
         ↓ persists
internal/adapters/persistence/sqlite/    ←  MODIFY: migration v3, helpers, queries
```

StateService does NOT call `UpdateState()` directly - it:
1. Loads project via `repo.FindByID()`
2. Validates transition (state + favorite check)
3. Modifies `project.State` and `project.HibernatedAt`
4. Saves via `repo.Save()` (upsert handles all fields)

### Migration v3 Implementation

Add to `migrations.go` var `migrations` slice:

```go
{
    Version:     3,
    Description: "Add hibernated_at column to projects",
    SQL:         "ALTER TABLE projects ADD COLUMN hibernated_at TEXT;",
},
```

Update `schema.go`:
```go
const SchemaVersion = 3
```

### Database Field Mapping

Update `helpers.go` projectRow struct (after line 31):
```go
HibernatedAt   sql.NullString `db:"hibernated_at"`
```

Update `rowToProject()` to parse:
```go
var hibernatedAt *time.Time
if row.HibernatedAt.Valid {
    t, err := time.Parse(time.RFC3339Nano, row.HibernatedAt.String)
    if err == nil {
        hibernatedAt = &t
    }
}
// Then assign: HibernatedAt: hibernatedAt,
```

### Test Patterns

Use existing mock pattern from `activity_tracker_test.go`:

```go
type mockRepository struct {
    projects map[string]*domain.Project
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
    if p, ok := m.projects[id]; ok {
        return p, nil
    }
    return nil, domain.ErrProjectNotFound
}
```

### Wire-up Note

StateService wire-up to TUI/CLI is **out of scope** - deferred to Story 11.5 (Manual State Control).

## File List

**CREATE:**
- `internal/core/services/state_service.go` ✅
- `internal/core/services/state_service_test.go` ✅

**MODIFY:**
- `internal/core/domain/project.go` - add `HibernatedAt *time.Time`, `DaysSinceHibernated()` ✅
- `internal/core/domain/project_test.go` - add `TestProject_DaysSinceHibernated` ✅
- `internal/core/domain/errors.go` - add `ErrInvalidStateTransition`, `ErrFavoriteCannotHibernate` ✅
- `internal/adapters/persistence/sqlite/migrations.go` - add migration v3 ✅
- `internal/adapters/persistence/sqlite/schema.go` - bump SchemaVersion to 3 ✅
- `internal/adapters/persistence/sqlite/helpers.go` - add HibernatedAt to projectRow, rowToProject, nullTimeString() ✅
- `internal/adapters/persistence/sqlite/queries.go` - add hibernated_at to projectColumns ✅
- `internal/adapters/persistence/sqlite/project_repository.go` - update Save() for HibernatedAt ✅
- `internal/adapters/persistence/sqlite/project_repository_integration_test.go` - add HibernatedAt round-trip test ✅

**DO NOT MODIFY:**
- `internal/core/domain/state.go` - enum already correct ✅
- `internal/core/ports/repository.go` - interface unchanged (uses Save) ✅

## Verification Checklist

Before marking complete, verify:

- [x] `go build ./...` succeeds
- [x] `go test ./internal/core/services/...` passes with 95%+ coverage for state_service.go (100% achieved)
- [x] `go test ./internal/adapters/persistence/sqlite/...` passes
- [x] Migration v3 applies cleanly on fresh DB
- [x] Migration v3 applies cleanly on existing v2 DB (tested via repo unit tests)
- [x] `golangci-lint run` passes

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None - implementation proceeded without issues.

### Completion Notes List

- **AC1 Verified**: `ProjectState` enum confirmed in `state.go` with `StateActive` (zero value) and `StateHibernated`
- **AC2 Implemented**: State transitions Active→Hibernated and Hibernated→Active with `ErrInvalidStateTransition` for invalid transitions
- **AC3 Implemented**: `ErrFavoriteCannotHibernate` prevents hibernating favorite projects (FR30)
- **AC4 Implemented**: State changes persist via `repo.Save()`, `UpdatedAt` updated on transitions
- **AC5 Implemented**: `StateService` with `Hibernate()` and `Activate()` methods
- **AC6 Implemented**: `HibernatedAt *time.Time` field added, migration v3 created, persistence complete
- **AC7 Achieved**: 100% test coverage for `state_service.go` (11 unit tests covering all paths including Save() error paths)

### Code Review Fixes Applied (2026-01-02)

| Issue | Severity | Fix Applied |
|-------|----------|-------------|
| M1: Missing test for Save() error path | MEDIUM | Added `TestStateService_Hibernate_SaveError_ReturnsError` and `TestStateService_Activate_SaveError_ReturnsError` |
| M2: Inconsistent time.Now() pattern in Activate() | MEDIUM | Changed to use `now := time.Now()` consistent with Hibernate() |
| M3: DaysSinceHibernated() truncation undocumented | MEDIUM | Added doc comment clarifying truncating integer division |

### Change Log

| Date | Change | Files |
|------|--------|-------|
| 2026-01-02 | Added `HibernatedAt *time.Time` field to Project entity | `internal/core/domain/project.go` |
| 2026-01-02 | Added `DaysSinceHibernated()` helper method | `internal/core/domain/project.go` |
| 2026-01-02 | Added `ErrInvalidStateTransition`, `ErrFavoriteCannotHibernate` | `internal/core/domain/errors.go` |
| 2026-01-02 | Created StateService with Hibernate/Activate methods | `internal/core/services/state_service.go` |
| 2026-01-02 | Added migration v3 for hibernated_at column | `internal/adapters/persistence/sqlite/migrations.go` |
| 2026-01-02 | Updated schema version to 3 | `internal/adapters/persistence/sqlite/schema.go` |
| 2026-01-02 | Added HibernatedAt to projectRow struct | `internal/adapters/persistence/sqlite/helpers.go` |
| 2026-01-02 | Added nullTimeString() helper function | `internal/adapters/persistence/sqlite/helpers.go` |
| 2026-01-02 | Updated projectColumns to include hibernated_at | `internal/adapters/persistence/sqlite/queries.go` |
| 2026-01-02 | Updated Save() to persist HibernatedAt | `internal/adapters/persistence/sqlite/project_repository.go` |
| 2026-01-02 | Added unit tests for StateService | `internal/core/services/state_service_test.go` |
| 2026-01-02 | Added DaysSinceHibernated tests | `internal/core/domain/project_test.go` |
| 2026-01-02 | Added HibernatedAt integration tests | `internal/adapters/persistence/sqlite/project_repository_integration_test.go` |
| 2026-01-02 | Code review fixes: time consistency, doc comments, Save error tests | `state_service.go`, `state_service_test.go`, `project.go` |
