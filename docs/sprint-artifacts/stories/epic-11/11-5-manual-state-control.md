# Story 11.5: Manual State Control

Status: review

## Story

- **As a** user
- **I want** to manually hibernate or activate projects via CLI
- **So that** I can override automatic hibernation behavior

## User-Visible Changes

- **New:** `vibe hibernate <project>` command hibernates a project (moves to dormant)
- **New:** `vibe activate <project>` command activates a hibernated project (moves to active)
- **New:** Success messages: "✓ Hibernated: project-name" / "✓ Activated: project-name"
- **New:** Error when hibernating favorite: "Cannot hibernate favorite project: project-name" with hint
- **Changed:** `vibe status <project>` already shows State field - confirms hibernation worked

## Context & Background

### Previous Stories
- **Story 11.1** (Project State Model): Created `StateService` with `Hibernate()` and `Activate()` methods in `internal/core/services/state_service.go`
- **Story 11.2** (Auto-Hibernation): `HibernationService` automatically hibernates after N days of inactivity
- **Story 11.3** (Auto-Activation): File watcher auto-activates hibernated projects on activity
- **Story 11.4** (Hibernated View): TUI 'h' key shows hibernated projects, Enter reactivates

### Current State
- `StateService` exists at `internal/core/services/state_service.go` with:
  - `Hibernate(ctx, projectID)` - returns `ErrInvalidStateTransition` if already hibernated, `ErrFavoriteCannotHibernate` if favorite
  - `Activate(ctx, projectID)` - returns `ErrInvalidStateTransition` if already active
- `stateService` already injected in CLI package via `deps.go:23` (`var stateService ports.StateActivator`)
- `SetStateService()` wiring exists at `deps.go:76-78`
- `findProjectByIdentifier()` exists in `status.go:81-116` - REUSE for project lookup

### Functional Requirements
- **FR57**: Manually hibernate or activate projects via CLI

## Acceptance Criteria

### AC1: Hibernate Active Project
- **Given** project "client-alpha" is active
- **When** I run `vibe hibernate client-alpha`
- **Then** project state changes to Hibernated
- **And** message shows: "✓ Hibernated: client-alpha"
- **And** exit code is 0

### AC2: Activate Hibernated Project
- **Given** project "old-project" is hibernated
- **When** I run `vibe activate old-project`
- **Then** project state changes to Active
- **And** message shows: "✓ Activated: old-project"
- **And** exit code is 0

### AC3: Hibernate Already-Hibernated Project
- **Given** project "dormant" is already hibernated
- **When** I run `vibe hibernate dormant`
- **Then** message shows: "Project is already hibernated: dormant"
- **And** exit code is 0 (idempotent)

### AC4: Activate Already-Active Project
- **Given** project "working" is already active
- **When** I run `vibe activate working`
- **Then** message shows: "Project is already active: working"
- **And** exit code is 0 (idempotent)

### AC5: Hibernate Favorite Project (Warning)
- **Given** project "pinned" is favorited
- **When** I run `vibe hibernate pinned`
- **Then** StateService returns `ErrFavoriteCannotHibernate`
- **And** message shows: "Cannot hibernate favorite project: pinned"
- **And** hint shows: "Remove favorite status first with: vibe favorite pinned --off"
- **And** exit code is 1

### AC6: Project Not Found
- **Given** project "nonexistent" doesn't exist
- **When** I run `vibe hibernate nonexistent`
- **Then** message shows: "✗ Project not found: nonexistent"
- **And** exit code is 2 (ExitNotFound)

### AC7: Quiet Mode (--quiet / -q)
- **Given** --quiet flag is set
- **When** running hibernate or activate
- **Then** no output is shown on success
- **And** exit code is 0

### AC8: Project Lookup by Identifier
- **Given** project can be identified by name, display name, or path
- **When** running `vibe hibernate /path/to/project`
- **Then** project is found and hibernated
- (Same as status.go pattern - uses findProjectByIdentifier)

### AC9: Shell Completion
- **Given** user is typing `vibe hibernate <TAB>`
- **Then** active projects are suggested (already registered via projectCompletionFunc)

## Tasks / Subtasks

- [x] Task 1: Create hibernate.go CLI command (AC: #1, #3, #5, #6, #7, #8, #9)
  - [x] 1.1: Create `internal/adapters/cli/hibernate.go`
  - [x] 1.2: Implement `newHibernateCmd()` following favorite.go pattern
  - [x] 1.3: Add to `RootCmd` in init() function
  - [x] 1.4: Use `findProjectByIdentifier()` for project lookup (same as status.go)
  - [x] 1.5: Call `stateService.Hibernate()` (already injected via deps.go)
  - [x] 1.6: Handle `ErrInvalidStateTransition` (already hibernated - AC3)
  - [x] 1.7: Handle `ErrFavoriteCannotHibernate` with helpful hint (AC5)
  - [x] 1.8: Handle `ErrProjectNotFound` with exit code 2 (AC6)
  - [x] 1.9: Respect `IsQuiet()` for output suppression (AC7)
  - [x] 1.10: Register `projectCompletionFunc` for shell completion (AC9)

- [x] Task 2: Create activate.go CLI command (AC: #2, #4, #6, #7, #8, #9)
  - [x] 2.1: Create `internal/adapters/cli/activate.go`
  - [x] 2.2: Implement `newActivateCmd()` following favorite.go pattern
  - [x] 2.3: Add to `RootCmd` in init() function
  - [x] 2.4: Use `findProjectByIdentifier()` for project lookup
  - [x] 2.5: Call `stateService.Activate()` (already injected via deps.go)
  - [x] 2.6: Handle `ErrInvalidStateTransition` (already active - AC4)
  - [x] 2.7: Handle `ErrProjectNotFound` with exit code 2 (AC6)
  - [x] 2.8: Respect `IsQuiet()` for output suppression (AC7)
  - [x] 2.9: Register `projectCompletionFunc` for shell completion (AC9)

- [x] Task 3: Write comprehensive tests (AC: #1-9)
  - [x] 3.1: Create `internal/adapters/cli/hibernate_test.go`
  - [x] 3.2: Create `internal/adapters/cli/activate_test.go`
  - [x] 3.3: Test: Hibernate active project succeeds
  - [x] 3.4: Test: Activate hibernated project succeeds
  - [x] 3.5: Test: Hibernate already-hibernated is idempotent (no error)
  - [x] 3.6: Test: Activate already-active is idempotent (no error)
  - [x] 3.7: Test: Hibernate favorite returns error with hint
  - [x] 3.8: Test: Project not found returns ExitNotFound
  - [x] 3.9: Test: Quiet mode suppresses output
  - [x] 3.10: Test: Lookup by display name works
  - [x] 3.11: Test: Lookup by path works
  - [x] 3.12: Test: StateService nil returns error

## Technical Implementation Guide

### Overview
Add two new CLI commands (`vibe hibernate` and `vibe activate`) that wrap the existing `StateService.Hibernate()` and `StateService.Activate()` methods. These commands follow the same patterns as `favorite.go` and use existing infrastructure.

### Architecture Compliance

```
internal/adapters/cli/hibernate.go  ←  NEW: CLI command
internal/adapters/cli/activate.go   ←  NEW: CLI command
         ↓ uses
internal/adapters/cli/status.go     ←  EXISTING: findProjectByIdentifier()
internal/adapters/cli/deps.go       ←  EXISTING: stateService (already wired)
         ↓ calls
internal/core/ports/state.go        ←  EXISTING: StateActivator interface
internal/core/services/state_service.go ←  EXISTING: Hibernate(), Activate()
```

### File Changes

#### 1. `internal/adapters/cli/hibernate.go` (CREATE)

```go
package cli

import (
    "errors"
    "fmt"

    "github.com/spf13/cobra"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// newHibernateCmd creates the hibernate command.
func newHibernateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "hibernate <project-name>",
        Short: "Hibernate a project (mark as dormant)",
        Long: `Hibernate a project, moving it from active to dormant state.

Hibernated projects:
  - Don't appear in the main dashboard (press [h] to view)
  - Don't trigger agent waiting detection
  - Auto-activate when file changes are detected

Projects can be identified by name, display name, or path.
Favorite projects cannot be hibernated (remove favorite first).

Examples:
  vibe hibernate my-project             # By name
  vibe hibernate /home/user/my-project  # By path
  vibe hibernate "My Cool App"          # By display name`,
        Args:              cobra.ExactArgs(1),
        ValidArgsFunction: projectCompletionFunc,
        RunE:              runHibernate,
    }

    return cmd
}

// RegisterHibernateCommand registers the hibernate command with the given parent.
// Used for testing to create fresh command trees.
func RegisterHibernateCommand(parent *cobra.Command) {
    parent.AddCommand(newHibernateCmd())
}

func init() {
    RootCmd.AddCommand(newHibernateCmd())
}

func runHibernate(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()

    // Validate stateService is available
    if stateService == nil {
        return fmt.Errorf("state service not initialized")
    }

    identifier := args[0]

    // Find project using existing helper (from status.go)
    project, err := findProjectByIdentifier(ctx, identifier)
    if err != nil {
        if errors.Is(err, domain.ErrProjectNotFound) {
            cmd.SilenceErrors = true
            cmd.SilenceUsage = true
            fmt.Fprintf(cmd.OutOrStdout(), "✗ Project not found: %s\n", identifier)
        }
        return err
    }

    // Attempt hibernation via StateService
    err = stateService.Hibernate(ctx, project.ID)
    if err != nil {
        // Handle specific errors
        if errors.Is(err, domain.ErrInvalidStateTransition) {
            // Already hibernated - idempotent success (AC3)
            if !IsQuiet() {
                fmt.Fprintf(cmd.OutOrStdout(), "Project is already hibernated: %s\n", identifier)
            }
            return nil
        }
        if errors.Is(err, domain.ErrFavoriteCannotHibernate) {
            // Favorite cannot hibernate (AC5)
            cmd.SilenceErrors = true
            cmd.SilenceUsage = true
            fmt.Fprintf(cmd.OutOrStdout(), "Cannot hibernate favorite project: %s\n", identifier)
            fmt.Fprintf(cmd.OutOrStdout(), "Remove favorite status first with: vibe favorite %s --off\n", identifier)
            return err
        }
        return fmt.Errorf("failed to hibernate project: %w", err)
    }

    // Success output
    if !IsQuiet() {
        fmt.Fprintf(cmd.OutOrStdout(), "✓ Hibernated: %s\n", identifier)
    }

    return nil
}
```

#### 2. `internal/adapters/cli/activate.go` (CREATE)

```go
package cli

import (
    "errors"
    "fmt"

    "github.com/spf13/cobra"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// newActivateCmd creates the activate command.
func newActivateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "activate <project-name>",
        Short: "Activate a hibernated project",
        Long: `Activate a hibernated project, moving it back to active state.

Activated projects:
  - Appear in the main dashboard
  - Resume agent waiting detection
  - May auto-hibernate again after hibernation threshold days

Projects can be identified by name, display name, or path.

Examples:
  vibe activate my-project             # By name
  vibe activate /home/user/my-project  # By path
  vibe activate "My Cool App"          # By display name`,
        Args:              cobra.ExactArgs(1),
        ValidArgsFunction: projectCompletionFunc,
        RunE:              runActivate,
    }

    return cmd
}

// RegisterActivateCommand registers the activate command with the given parent.
// Used for testing to create fresh command trees.
func RegisterActivateCommand(parent *cobra.Command) {
    parent.AddCommand(newActivateCmd())
}

func init() {
    RootCmd.AddCommand(newActivateCmd())
}

func runActivate(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()

    // Validate stateService is available
    if stateService == nil {
        return fmt.Errorf("state service not initialized")
    }

    identifier := args[0]

    // Find project using existing helper (from status.go)
    project, err := findProjectByIdentifier(ctx, identifier)
    if err != nil {
        if errors.Is(err, domain.ErrProjectNotFound) {
            cmd.SilenceErrors = true
            cmd.SilenceUsage = true
            fmt.Fprintf(cmd.OutOrStdout(), "✗ Project not found: %s\n", identifier)
        }
        return err
    }

    // Attempt activation via StateService
    err = stateService.Activate(ctx, project.ID)
    if err != nil {
        // Handle specific errors
        if errors.Is(err, domain.ErrInvalidStateTransition) {
            // Already active - idempotent success (AC4)
            if !IsQuiet() {
                fmt.Fprintf(cmd.OutOrStdout(), "Project is already active: %s\n", identifier)
            }
            return nil
        }
        return fmt.Errorf("failed to activate project: %w", err)
    }

    // Success output
    if !IsQuiet() {
        fmt.Fprintf(cmd.OutOrStdout(), "✓ Activated: %s\n", identifier)
    }

    return nil
}
```

#### 3. `internal/adapters/cli/hibernate_test.go` (CREATE)

**CRITICAL:** Follow `favorite_test.go` pattern - use `package cli_test` with external imports.

```go
package cli_test

import (
    "bytes"
    "context"
    "strings"
    "testing"
    "time"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 11.5: CLI Hibernate Command Tests
// ============================================================================

// mockStateService implements ports.StateActivator for hibernate/activate tests.
type mockStateService struct {
    hibernateErr    error
    activateErr     error
    hibernateCalled bool
    activateCalled  bool
    lastProjectID   string
}

func (m *mockStateService) Hibernate(_ context.Context, projectID string) error {
    m.hibernateCalled = true
    m.lastProjectID = projectID
    return m.hibernateErr
}

func (m *mockStateService) Activate(_ context.Context, projectID string) error {
    m.activateCalled = true
    m.lastProjectID = projectID
    return m.activateErr
}

// hibernateMockRepository implements ports.ProjectRepository for hibernate tests.
// Follows favorite_test.go pattern - local mock per test file.
type hibernateMockRepository struct {
    projects   map[string]*domain.Project
    saveErr    error
    findAllErr error
}

func newHibernateMockRepository() *hibernateMockRepository {
    return &hibernateMockRepository{projects: make(map[string]*domain.Project)}
}

func (m *hibernateMockRepository) withProjects(projects []*domain.Project) *hibernateMockRepository {
    for _, p := range projects {
        m.projects[p.Path] = p
    }
    return m
}

func (m *hibernateMockRepository) Save(_ context.Context, project *domain.Project) error {
    if m.saveErr != nil {
        return m.saveErr
    }
    m.projects[project.Path] = project
    return nil
}

func (m *hibernateMockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
    for _, p := range m.projects {
        if p.ID == id {
            return p, nil
        }
    }
    return nil, domain.ErrProjectNotFound
}

func (m *hibernateMockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
    if p, ok := m.projects[path]; ok {
        return p, nil
    }
    return nil, domain.ErrProjectNotFound
}

func (m *hibernateMockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
    if m.findAllErr != nil {
        return nil, m.findAllErr
    }
    result := make([]*domain.Project, 0, len(m.projects))
    for _, p := range m.projects {
        result = append(result, p)
    }
    return result, nil
}

func (m *hibernateMockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
    result := make([]*domain.Project, 0)
    for _, p := range m.projects {
        if p.State == domain.StateActive {
            result = append(result, p)
        }
    }
    return result, nil
}

func (m *hibernateMockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
    result := make([]*domain.Project, 0)
    for _, p := range m.projects {
        if p.State == domain.StateHibernated {
            result = append(result, p)
        }
    }
    return result, nil
}

func (m *hibernateMockRepository) Delete(_ context.Context, id string) error {
    for path, p := range m.projects {
        if p.ID == id {
            delete(m.projects, path)
            return nil
        }
    }
    return domain.ErrProjectNotFound
}

func (m *hibernateMockRepository) UpdateState(_ context.Context, id string, state domain.ProjectState) error {
    for _, p := range m.projects {
        if p.ID == id {
            p.State = state
            return nil
        }
    }
    return domain.ErrProjectNotFound
}

func (m *hibernateMockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
    return nil
}

func (m *hibernateMockRepository) ResetProject(_ context.Context, _ string) error {
    return nil
}

func (m *hibernateMockRepository) ResetAll(_ context.Context) (int, error) {
    return 0, nil
}

// executeHibernateCommand runs the hibernate command with given args and returns output/error.
// Follows favorite_test.go:126-140 pattern.
func executeHibernateCommand(args []string) (string, error) {
    cmd := cli.NewRootCmd()
    cli.RegisterHibernateCommand(cmd)

    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetErr(&buf)

    fullArgs := append([]string{"hibernate"}, args...)
    cmd.SetArgs(fullArgs)

    err := cmd.Execute()
    return buf.String(), err
}

// TestHibernateCmd_ActiveProject_Succeeds verifies hibernating active project (AC1).
func TestHibernateCmd_ActiveProject_Succeeds(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", State: domain.StateActive},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute
    output, err := executeHibernateCommand([]string{"test-project"})

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !mockState.hibernateCalled {
        t.Error("expected Hibernate to be called")
    }
    if !strings.Contains(output, "✓ Hibernated") {
        t.Errorf("expected '✓ Hibernated', got: %s", output)
    }
}

// TestHibernateCmd_AlreadyHibernated_Idempotent verifies idempotent behavior (AC3).
func TestHibernateCmd_AlreadyHibernated_Idempotent(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", State: domain.StateHibernated},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{hibernateErr: domain.ErrInvalidStateTransition}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute
    output, err := executeHibernateCommand([]string{"test-project"})

    // Assert - should succeed (idempotent)
    if err != nil {
        t.Fatalf("expected no error (idempotent), got: %v", err)
    }
    if !strings.Contains(output, "already hibernated") {
        t.Errorf("expected 'already hibernated' message, got: %s", output)
    }
}

// TestHibernateCmd_FavoriteProject_ReturnsError verifies favorite rejection (AC5).
func TestHibernateCmd_FavoriteProject_ReturnsError(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", IsFavorite: true, State: domain.StateActive},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{hibernateErr: domain.ErrFavoriteCannotHibernate}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute
    output, err := executeHibernateCommand([]string{"test-project"})

    // Assert
    if err == nil {
        t.Fatal("expected error for favorite project")
    }
    if !strings.Contains(output, "Cannot hibernate favorite") {
        t.Errorf("expected favorite error message, got: %s", output)
    }
    if !strings.Contains(output, "vibe favorite test-project --off") {
        t.Errorf("expected hint about removing favorite, got: %s", output)
    }
}

// TestHibernateCmd_ProjectNotFound_ReturnsError verifies not found handling (AC6).
func TestHibernateCmd_ProjectNotFound_ReturnsError(t *testing.T) {
    // Setup - empty repository
    cli.SetRepository(newHibernateMockRepository())
    cli.SetStateService(&mockStateService{})
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute
    output, err := executeHibernateCommand([]string{"nonexistent"})

    // Assert
    if err == nil {
        t.Fatal("expected error for nonexistent project")
    }
    if !strings.Contains(output, "Project not found") {
        t.Errorf("expected 'Project not found' message, got: %s", output)
    }
}

// TestHibernateCmd_QuietMode_SuppressesOutput verifies quiet mode (AC7).
func TestHibernateCmd_QuietMode_SuppressesOutput(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", State: domain.StateActive},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    cli.ResetQuietFlag()
    cli.SetQuietForTest(true)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
        cli.ResetQuietFlag()
    }()

    // Execute
    cmd := cli.NewRootCmd()
    cli.RegisterHibernateCommand(cmd)

    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetErr(&buf)
    cmd.SetArgs([]string{"hibernate", "test-project"})

    err := cmd.Execute()

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if buf.String() != "" {
        t.Errorf("expected empty output with --quiet, got: %s", buf.String())
    }
}

// TestHibernateCmd_StateServiceNil_ReturnsError verifies nil stateService handling.
func TestHibernateCmd_StateServiceNil_ReturnsError(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", State: domain.StateActive},
    }
    cli.SetRepository(newHibernateMockRepository().withProjects(projects))
    cli.SetStateService(nil) // Explicitly nil
    defer func() {
        cli.SetRepository(nil)
    }()

    // Execute
    _, err := executeHibernateCommand([]string{"test-project"})

    // Assert
    if err == nil {
        t.Fatal("expected error when stateService is nil")
    }
}

// TestHibernateCmd_FindByPath verifies path-based lookup (AC8).
func TestHibernateCmd_FindByPath(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test/my-project", Name: "my-project", State: domain.StateActive},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute using path
    output, err := executeHibernateCommand([]string{"/test/my-project"})

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !mockState.hibernateCalled {
        t.Error("expected Hibernate to be called via path lookup")
    }
    if !strings.Contains(output, "✓ Hibernated") {
        t.Errorf("expected success message, got: %s", output)
    }
}
```

#### 4. `internal/adapters/cli/activate_test.go` (CREATE)

**CRITICAL:** Follow `favorite_test.go` pattern - use `package cli_test` with external imports.
Reuses `mockStateService` and `hibernateMockRepository` types from hibernate_test.go (same package).

```go
package cli_test

import (
    "bytes"
    "strings"
    "testing"

    "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 11.5: CLI Activate Command Tests
// ============================================================================

// NOTE: mockStateService and hibernateMockRepository are defined in hibernate_test.go
// (same cli_test package, so they're accessible here)

// executeActivateCommand runs the activate command with given args and returns output/error.
// Follows favorite_test.go pattern.
func executeActivateCommand(args []string) (string, error) {
    cmd := cli.NewRootCmd()
    cli.RegisterActivateCommand(cmd)

    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetErr(&buf)

    fullArgs := append([]string{"activate"}, args...)
    cmd.SetArgs(fullArgs)

    err := cmd.Execute()
    return buf.String(), err
}

// TestActivateCmd_HibernatedProject_Succeeds verifies activating hibernated project (AC2).
func TestActivateCmd_HibernatedProject_Succeeds(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", State: domain.StateHibernated},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute
    output, err := executeActivateCommand([]string{"test-project"})

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !mockState.activateCalled {
        t.Error("expected Activate to be called")
    }
    if !strings.Contains(output, "✓ Activated") {
        t.Errorf("expected '✓ Activated', got: %s", output)
    }
}

// TestActivateCmd_AlreadyActive_Idempotent verifies idempotent behavior (AC4).
func TestActivateCmd_AlreadyActive_Idempotent(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", State: domain.StateActive},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{activateErr: domain.ErrInvalidStateTransition}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute
    output, err := executeActivateCommand([]string{"test-project"})

    // Assert - should succeed (idempotent)
    if err != nil {
        t.Fatalf("expected no error (idempotent), got: %v", err)
    }
    if !strings.Contains(output, "already active") {
        t.Errorf("expected 'already active' message, got: %s", output)
    }
}

// TestActivateCmd_ProjectNotFound_ReturnsError verifies not found handling (AC6).
func TestActivateCmd_ProjectNotFound_ReturnsError(t *testing.T) {
    // Setup - empty repository
    cli.SetRepository(newHibernateMockRepository())
    cli.SetStateService(&mockStateService{})
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute
    output, err := executeActivateCommand([]string{"nonexistent"})

    // Assert
    if err == nil {
        t.Fatal("expected error for nonexistent project")
    }
    if !strings.Contains(output, "Project not found") {
        t.Errorf("expected 'Project not found' message, got: %s", output)
    }
}

// TestActivateCmd_QuietMode_SuppressesOutput verifies quiet mode (AC7).
func TestActivateCmd_QuietMode_SuppressesOutput(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", State: domain.StateHibernated},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    cli.ResetQuietFlag()
    cli.SetQuietForTest(true)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
        cli.ResetQuietFlag()
    }()

    // Execute
    cmd := cli.NewRootCmd()
    cli.RegisterActivateCommand(cmd)

    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetErr(&buf)
    cmd.SetArgs([]string{"activate", "test-project"})

    err := cmd.Execute()

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if buf.String() != "" {
        t.Errorf("expected empty output with --quiet, got: %s", buf.String())
    }
}

// TestActivateCmd_ByDisplayName_Succeeds verifies display name lookup (AC8).
func TestActivateCmd_ByDisplayName_Succeeds(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", DisplayName: "My Cool App", State: domain.StateHibernated},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute using display name
    output, err := executeActivateCommand([]string{"My Cool App"})

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !mockState.activateCalled {
        t.Error("expected Activate to be called via display name lookup")
    }
    if !strings.Contains(output, "✓ Activated") {
        t.Errorf("expected success message, got: %s", output)
    }
}

// TestActivateCmd_StateServiceNil_ReturnsError verifies nil stateService handling.
func TestActivateCmd_StateServiceNil_ReturnsError(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test", Name: "test-project", State: domain.StateHibernated},
    }
    cli.SetRepository(newHibernateMockRepository().withProjects(projects))
    cli.SetStateService(nil) // Explicitly nil
    defer func() {
        cli.SetRepository(nil)
    }()

    // Execute
    _, err := executeActivateCommand([]string{"test-project"})

    // Assert
    if err == nil {
        t.Fatal("expected error when stateService is nil")
    }
}

// TestActivateCmd_FindByPath verifies path-based lookup (AC8).
func TestActivateCmd_FindByPath(t *testing.T) {
    // Setup
    projects := []*domain.Project{
        {ID: "1", Path: "/test/my-project", Name: "my-project", State: domain.StateHibernated},
    }
    mockRepo := newHibernateMockRepository().withProjects(projects)
    mockState := &mockStateService{}

    cli.SetRepository(mockRepo)
    cli.SetStateService(mockState)
    defer func() {
        cli.SetRepository(nil)
        cli.SetStateService(nil)
    }()

    // Execute using path
    output, err := executeActivateCommand([]string{"/test/my-project"})

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !mockState.activateCalled {
        t.Error("expected Activate to be called via path lookup")
    }
    if !strings.Contains(output, "✓ Activated") {
        t.Errorf("expected success message, got: %s", output)
    }
}
```

### Testing Strategy

#### Unit Tests

| Test Case | File | Coverage |
|-----------|------|----------|
| Hibernate active project succeeds | hibernate_test.go | AC1 |
| Hibernate already-hibernated is idempotent | hibernate_test.go | AC3 |
| Hibernate favorite returns error with hint | hibernate_test.go | AC5 |
| Hibernate project not found | hibernate_test.go | AC6 |
| Hibernate quiet mode | hibernate_test.go | AC7 |
| Hibernate stateService nil | hibernate_test.go | Edge case |
| Activate hibernated project succeeds | activate_test.go | AC2 |
| Activate already-active is idempotent | activate_test.go | AC4 |
| Activate project not found | activate_test.go | AC6 |
| Activate quiet mode | activate_test.go | AC7 |
| Activate by display name | activate_test.go | AC8 |
| Activate stateService nil | activate_test.go | Edge case |

### Edge Cases

1. **StateService not initialized**: Return clear error message
2. **Repository not initialized**: Return clear error message (handled by findProjectByIdentifier)
3. **Concurrent state change**: StateService handles this atomically via Save()
4. **Path with spaces**: Works via findProjectByIdentifier's canonicalization

## Dev Notes

| Decision | Rationale |
|----------|-----------|
| Idempotent behavior for already-in-state | User expectation - running `hibernate` on hibernated project should succeed quietly |
| Helpful hint for favorite rejection | FR57 implies user intent; guide them to solution |
| Reuse findProjectByIdentifier | Same lookup pattern as status.go, favorite.go, note.go |
| Use existing stateService injection | deps.go already has wiring from Story 11.3 |
| No new dependencies | Pure CLI wrapper around existing StateService |

### Reuse vs Create Quick Reference

| Item | Action | Source |
|------|--------|--------|
| `StateService.Hibernate()` | REUSE | `internal/core/services/state_service.go:34` |
| `StateService.Activate()` | REUSE | `internal/core/services/state_service.go:62` |
| `stateService` var | REUSE | `internal/adapters/cli/deps.go:23` |
| `SetStateService()` | REUSE | `internal/adapters/cli/deps.go:76` |
| `findProjectByIdentifier()` | REUSE | `internal/adapters/cli/status.go:81` |
| `projectCompletionFunc` | REUSE | `internal/adapters/cli/completion.go` |
| `IsQuiet()` | REUSE | `internal/adapters/cli/flags.go` |
| `domain.ErrInvalidStateTransition` | REUSE | `internal/core/domain/errors.go` |
| `domain.ErrFavoriteCannotHibernate` | REUSE | `internal/core/domain/errors.go` |
| `domain.ErrProjectNotFound` | REUSE | `internal/core/domain/errors.go` |
| `hibernate.go` | CREATE | NEW CLI command file |
| `activate.go` | CREATE | NEW CLI command file |
| `hibernate_test.go` | CREATE | NEW test file |
| `activate_test.go` | CREATE | NEW test file |

## Dependencies

- **Story 11.1**: StateService with Hibernate/Activate methods (DONE)
- **Story 11.3**: stateService injection in CLI deps.go (DONE)

## File List

**CREATE:**
- `internal/adapters/cli/hibernate.go` - New CLI command
- `internal/adapters/cli/activate.go` - New CLI command
- `internal/adapters/cli/hibernate_test.go` - Tests for hibernate
- `internal/adapters/cli/activate_test.go` - Tests for activate

**MODIFY:**
- `docs/sprint-artifacts/sprint-status.yaml` - Story status update

**DO NOT MODIFY:**
- `internal/core/services/state_service.go` - Already has Hibernate/Activate
- `internal/adapters/cli/deps.go` - Already has stateService wiring
- `internal/adapters/cli/status.go` - Already has findProjectByIdentifier

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Setup Test Projects

```bash
make build

# Add test project
./bin/vibe add /tmp/test-manual-state
mkdir -p /tmp/test-manual-state

# Verify project is active
./bin/vibe status test-manual-state
# Should show "State: Active"
```

### Step 2: Test Hibernate Command

| Check | Expected | Status |
|-------|----------|--------|
| Run `./bin/vibe hibernate test-manual-state` | Shows "✓ Hibernated: test-manual-state" | |
| Run `./bin/vibe status test-manual-state` | Shows "State: Hibernated" | |
| Run `./bin/vibe hibernate test-manual-state` again | Shows "Project is already hibernated" (idempotent) | |

### Step 3: Test Activate Command

| Check | Expected | Status |
|-------|----------|--------|
| Run `./bin/vibe activate test-manual-state` | Shows "✓ Activated: test-manual-state" | |
| Run `./bin/vibe status test-manual-state` | Shows "State: Active" | |
| Run `./bin/vibe activate test-manual-state` again | Shows "Project is already active" (idempotent) | |

### Step 4: Test Favorite Rejection

```bash
# Make project a favorite
./bin/vibe favorite test-manual-state

# Try to hibernate
./bin/vibe hibernate test-manual-state
# Should show error with hint to remove favorite
```

| Check | Expected | Status |
|-------|----------|--------|
| Error message | "Cannot hibernate favorite project" | |
| Hint message | "Remove favorite status first with: vibe favorite..." | |

### Step 5: Cleanup

```bash
./bin/vibe remove test-manual-state -y
```

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Any check fails | Do NOT approve, document issue |

## Verification Checklist

Before marking complete, verify:

- [x] `go build ./...` succeeds
- [x] `go test ./internal/adapters/cli/...` passes
- [x] `golangci-lint run` passes
- [x] User Testing Guide Step 2: Hibernate works
- [x] User Testing Guide Step 3: Activate works
- [x] User Testing Guide Step 4: Favorite rejection works

## Story Wrap Up (Agent Populates After Completion)

### Completion Checklist
- [x] All ACs verified
- [x] Tests pass
- [x] Code review findings addressed
- [x] Documentation updated if needed

### Dev Agent Record

**Implementation Date:** 2026-01-02

**Files Created:**
- `internal/adapters/cli/hibernate.go` - Hibernate command implementation
- `internal/adapters/cli/activate.go` - Activate command implementation
- `internal/adapters/cli/hibernate_test.go` - 12 unit tests for hibernate command
- `internal/adapters/cli/activate_test.go` - 9 unit tests for activate command

**Files Modified:**
- `internal/core/ports/state.go` - Added `Hibernate()` method to `StateActivator` interface
- `internal/core/ports/state_test.go` - Updated mock to include `Hibernate()` method
- `internal/adapters/tui/model_test.go` - Updated mock to include `Hibernate()` method

**Tests Added:**
- `TestHibernateCmd_ActiveProject_Succeeds` (AC1)
- `TestHibernateCmd_AlreadyHibernated_Idempotent` (AC3)
- `TestHibernateCmd_FavoriteProject_ReturnsError` (AC5)
- `TestHibernateCmd_ProjectNotFound_ReturnsError` (AC6)
- `TestHibernateCmd_QuietMode_SuppressesOutput` (AC7)
- `TestHibernateCmd_StateServiceNil_ReturnsError` (Edge case)
- `TestHibernateCmd_FindByPath` (AC8)
- `TestHibernateCmd_FindByDisplayName` (AC8)
- `TestHibernateCmd_NoArgs_ReturnsError` (Edge case)
- `TestHibernateCmd_QuietMode_AlreadyHibernated_SuppressesOutput` (AC7)
- `TestActivateCmd_HibernatedProject_Succeeds` (AC2)
- `TestActivateCmd_AlreadyActive_Idempotent` (AC4)
- `TestActivateCmd_ProjectNotFound_ReturnsError` (AC6)
- `TestActivateCmd_QuietMode_SuppressesOutput` (AC7)
- `TestActivateCmd_ByDisplayName_Succeeds` (AC8)
- `TestActivateCmd_StateServiceNil_ReturnsError` (Edge case)
- `TestActivateCmd_FindByPath` (AC8)
- `TestActivateCmd_NoArgs_ReturnsError` (Edge case)
- `TestActivateCmd_QuietMode_AlreadyActive_SuppressesOutput` (AC7)

**Implementation Notes:**
- Extended `StateActivator` interface to include `Hibernate()` method for CLI access
- Both commands follow existing patterns from `favorite.go` and `status.go`
- Reused `findProjectByIdentifier()` for project lookup (name, display name, or path)
- Reused `stateService` injection from Story 11.3
- All error handling follows existing CLI patterns (exit codes, silence flags)
- Shell completion via `projectCompletionFunc` was already available - just needed registration
- Interface change required updating 2 existing test mocks (TUI and ports)

**Interface Extension (L3 code review):**
- Added `Hibernate()` method to `StateActivator` interface (state.go:12)
- Reason: CLI needs Hibernate access but original interface only exposed Activate()
- TUI uses Activate() for auto-activation on file events (Story 11.3)
- CLI needs both for `vibe hibernate` and `vibe activate` commands (Story 11.5)

**Code Review Fixes Applied:**
- M1: Added Hibernate() test to TestStateActivator_Interface in state_test.go
- M2: Marked "Code review findings addressed" checkbox
- L1: Added comment explaining unused saveErr/findAllErr fields in mock
- L2: Improved StateActivator doc comment to clarify story provenance
- L3: Added interface extension rationale to Dev Agent Record
