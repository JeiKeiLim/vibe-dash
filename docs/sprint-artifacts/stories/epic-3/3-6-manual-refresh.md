# Story 3.6: Manual Refresh

**Status:** done

## Quick Reference

| Category | Items |
|----------|-------|
| **Entry Points** | `internal/adapters/tui/model.go` (KeyRefresh handling), `internal/adapters/cli/refresh.go` (new file) |
| **Key Dependencies** | DetectionService (`core/services/detection_service.go`), ProjectRepository (`core/ports/repository.go`), KeyRefresh constant (`keys.go`) |
| **Files to Modify** | `model.go` (add refresh handling), `status_bar.go` (add refresh state), `root.go` (wire detection service), `tui/app.go` (accept detection service) |
| **Files to Create** | `cli/refresh.go`, `cli/refresh_test.go`, `model_refresh_test.go` |
| **Location** | `internal/adapters/tui/`, `internal/adapters/cli/` |
| **Interfaces Used** | `ports.ProjectRepository`, `ports.Detector` |

### Quick Task Summary (5 Tasks)

| # | Task | Key Deliverable |
|---|------|-----------------|
| 1 | Add refresh state to TUI model | isRefreshing bool, refreshedCount int, refreshError string |
| 2 | Implement TUI refresh handler | Handle 'r' key, trigger detection rescan for all projects |
| 3 | Add refresh status to status bar | Show "Refreshing... (N/M)" spinner during refresh |
| 4 | Create CLI refresh command | `vibe refresh` for non-interactive refresh |
| 5 | Add tests | TUI refresh behavior + CLI refresh command tests |

### Key Technical Decisions

| Decision | Value | Why |
|----------|-------|-----|
| Refresh mechanism | Async via tea.Cmd | Non-blocking - navigation still works during refresh (per AC) |
| Status display | Status bar spinner | Consistent with UX spec, visible feedback |
| Detection injection | Optional dependency | DetectionService may not be wired initially |
| CLI command | Separate `vibe refresh` | Per FR58 - trigger refresh via CLI |
| Error handling | Partial success | Continue on individual failures, report errors (per AC) |

## Story

**As a** user,
**I want** to manually refresh project detection,
**So that** I can force re-scan of artifacts.

## Acceptance Criteria

```gherkin
AC1: Given dashboard is displayed
     When I press 'r'
     Then all projects are re-scanned
     And spinner shows in status bar: "Refreshing... (N/M)"
     And detection runs for each project
     And stages update if artifacts changed
     And status bar shows: "Refreshed N projects"

AC2: Given refresh completes
     When status bar updates
     Then timestamp "last refreshed Xs ago" could update (future enhancement)

AC3: Given refresh encounters errors
     When some projects fail to scan
     Then partial success is reported
     And errors are logged (not shown in UI unless all fail)

AC4: Given I run `vibe refresh` from CLI
     When the command executes
     Then same refresh occurs non-interactively
     And exit code 0 on success

AC5: Given refresh is in progress
     When I navigate with j/k keys
     Then navigation still works (non-blocking refresh)
```

## Tasks / Subtasks

- [x] **Task 1: Add refresh state to TUI model** (AC: 1, 5)
  - [x] 1.1 Add refresh-related fields to Model struct in `model.go`:
    ```go
    // Refresh state (Story 3.6)
    isRefreshing    bool
    refreshTotal    int
    refreshProgress int
    refreshError    string

    // Dependencies (add to existing)
    detectionService ports.Detector  // Optional - may be nil if not wired
    ```
  - [x] 1.2 Add refresh message types:
    ```go
    // refreshStartMsg signals refresh has started
    type refreshStartMsg struct {
        total int // Total number of projects to refresh
    }

    // refreshProgressMsg updates refresh progress
    type refreshProgressMsg struct {
        current int
        total   int
    }

    // refreshCompleteMsg signals refresh is complete
    type refreshCompleteMsg struct {
        refreshedCount int
        failedCount    int
        err            error // Only set if ALL projects failed
    }

    // clearRefreshMsgMsg signals to clear the refresh completion message
    // Sent after 3-second timer expires (Task 3.3)
    type clearRefreshMsgMsg struct{}
    ```
  - [x] 1.3 Add SetDetectionService method to Model (preserves existing signature):
    ```go
    // SetDetectionService sets the detection service for refresh operations.
    // This is optional - if not set, refresh will show "Detection service not available".
    func (m *Model) SetDetectionService(svc ports.Detector) {
        m.detectionService = svc
    }
    ```
    **IMPORTANT:** Do NOT change NewModel signature - this breaks existing call sites.
    Use setter injection pattern like other optional dependencies.

  - [x] 1.4 Update tui.Run and root.go to wire DetectionService:

    **1.4.1 Update `internal/adapters/tui/app.go` - Add detector parameter to Run:**
    ```go
    // Run starts the TUI with the given repository and optional detection service.
    // The detector parameter can be nil - refresh will be disabled.
    func Run(ctx context.Context, repo ports.ProjectRepository, detector ports.Detector) error {
        m := NewModel(repo)
        if detector != nil {
            m.SetDetectionService(detector)
        }
        // ... rest of Run
    }
    ```

    **1.4.2 Update `internal/adapters/cli/root.go` - Wire detection service:**
    ```go
    Run: func(cmd *cobra.Command, args []string) {
        // ... existing repo init ...

        // Initialize detection service (uses package-level detectionService from add.go)
        // This is already initialized in main.go via SetDetectionService()
        if err := tui.Run(cmd.Context(), repo, detectionService); err != nil {
            slog.Error("TUI error", "error", err)
        }
    }
    ```
    **Note:** Uses existing `detectionService` package variable from `add.go:26`

- [x] **Task 2: Implement TUI refresh handler** (AC: 1, 3, 5)
  - [x] 2.1 Add KeyRefresh case to handleKeyMsg in `model.go`:
    ```go
    case KeyRefresh:
        if m.isRefreshing {
            return m, nil // Ignore if already refreshing
        }
        if m.detectionService == nil {
            // No detection service - show message and return
            m.refreshError = "Detection service not available"
            return m, nil
        }
        return m.startRefresh()
    ```
  - [x] 2.2 Implement startRefresh method:
    ```go
    // startRefresh initiates async refresh of all projects
    func (m Model) startRefresh() (tea.Model, tea.Cmd) {
        m.isRefreshing = true
        m.refreshTotal = len(m.projects)
        m.refreshProgress = 0
        m.refreshError = ""

        return m, m.refreshProjectsCmd()
    }
    ```
  - [x] 2.3 Implement refreshProjectsCmd that iterates projects:
    ```go
    // refreshProjectsCmd creates a command that rescans all projects
    func (m Model) refreshProjectsCmd() tea.Cmd {
        return func() tea.Msg {
            ctx := context.Background()
            var refreshedCount, failedCount int

            for _, project := range m.projects {
                select {
                case <-ctx.Done():
                    return refreshCompleteMsg{refreshedCount, failedCount, ctx.Err()}
                default:
                }

                // Run detection
                result, err := m.detectionService.Detect(ctx, project.Path)
                if err != nil {
                    slog.Debug("refresh detection failed", "project", project.Name, "error", err)
                    failedCount++
                    continue
                }

                // Update project with new detection result
                project.DetectedMethod = result.Method
                project.CurrentStage = result.Stage
                project.Confidence = result.Confidence
                project.DetectionReasoning = result.Reasoning
                project.UpdatedAt = time.Now()

                if err := m.repository.Save(ctx, project); err != nil {
                    slog.Debug("refresh save failed", "project", project.Name, "error", err)
                    failedCount++
                    continue
                }

                refreshedCount++
            }

            var resultErr error
            if refreshedCount == 0 && failedCount > 0 {
                resultErr = fmt.Errorf("all projects failed to refresh")
            }

            return refreshCompleteMsg{refreshedCount, failedCount, resultErr}
        }
    }
    ```
  - [x] 2.4 Handle refresh messages in Update:
    ```go
    case refreshCompleteMsg:
        m.isRefreshing = false
        m.statusBar.SetRefreshing(false, 0, 0)
        if msg.err != nil {
            m.refreshError = msg.err.Error()
            m.statusBar.SetRefreshComplete("Refresh failed")
            return m, nil
        }
        m.refreshError = ""
        m.statusBar.SetRefreshComplete(fmt.Sprintf("Refreshed %d projects", msg.refreshedCount))
        // Reload projects and start timer to clear message
        return m, tea.Batch(
            m.loadProjectsCmd(),
            tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
                return clearRefreshMsgMsg{}
            }),
        )

    case clearRefreshMsgMsg:
        m.statusBar.SetRefreshComplete("") // Clear the message
        return m, nil
    ```

- [x] **Task 3: Add refresh status to status bar** (AC: 1, 2)
  - [x] 3.1 Add refresh state to StatusBarModel in `components/status_bar.go`:
    ```go
    type StatusBarModel struct {
        // ... existing fields
        isRefreshing    bool
        refreshProgress int
        refreshTotal    int
        lastRefreshMsg  string // "Refreshed N projects" or error
    }

    // SetRefreshing updates the refresh state
    func (s *StatusBarModel) SetRefreshing(isRefreshing bool, progress, total int) {
        s.isRefreshing = isRefreshing
        s.refreshProgress = progress
        s.refreshTotal = total
    }

    // SetRefreshComplete sets the completion message
    func (s *StatusBarModel) SetRefreshComplete(msg string) {
        s.lastRefreshMsg = msg
    }
    ```
  - [x] 3.2 Update renderCounts to show refresh status:
    ```go
    func (s StatusBarModel) renderCounts() string {
        // Show refresh spinner when refreshing
        if s.isRefreshing {
            spinnerText := fmt.Sprintf("Refreshing... (%d/%d)", s.refreshProgress, s.refreshTotal)
            return "| " + spinnerText + " |"
        }

        // Show refresh result for 3 seconds after completion
        if s.lastRefreshMsg != "" {
            parts := []string{
                fmt.Sprintf("%d active", s.activeCount),
                fmt.Sprintf("%d hibernated", s.hibernatedCount),
                s.lastRefreshMsg,
            }
            // ... rest of rendering
        }

        // Normal counts display
        // ... existing code
    }
    ```
  - [x] 3.3 Update model.go to set refresh state when starting:
    ```go
    // In startRefresh method, after setting m.isRefreshing = true:
    m.statusBar.SetRefreshing(true, 0, m.refreshTotal)

    // Note: Completion handling is in Task 2.4 (refreshCompleteMsg case)
    // which calls SetRefreshing(false, 0, 0) and SetRefreshComplete()
    ```

  - [x] 3.4 Add ClearRefreshComplete method to StatusBarModel (implemented via SetRefreshComplete("")):
    ```go
    // ClearRefreshComplete clears the completion message
    func (s *StatusBarModel) ClearRefreshComplete() {
        s.lastRefreshMsg = ""
    }
    ```

- [x] **Task 4: Create CLI refresh command** (AC: 4)

  **CRITICAL:** Use dependency injection pattern like `add.go`, NOT direct initialization.
  The `repository` and `detectionService` package variables are already defined in `add.go`.

  - [x] 4.1 Create `internal/adapters/cli/refresh.go`:
    ```go
    package cli

    import (
        "fmt"
        "log/slog"
        "time"

        "github.com/spf13/cobra"
    )

    // newRefreshCmd creates the refresh command.
    func newRefreshCmd() *cobra.Command {
        return &cobra.Command{
            Use:   "refresh",
            Short: "Refresh detection for all tracked projects",
            Long: `Re-scan all tracked projects to update their methodology stage.

This command runs the detection service against all projects and updates
their stage based on current artifacts.`,
            RunE: runRefresh,
        }
    }

    // RegisterRefreshCommand registers the refresh command with the given parent.
    // Used for testing to create fresh command trees.
    func RegisterRefreshCommand(parent *cobra.Command) {
        parent.AddCommand(newRefreshCmd())
    }

    func init() {
        RootCmd.AddCommand(newRefreshCmd())
    }

    func runRefresh(cmd *cobra.Command, args []string) error {
        ctx := cmd.Context()

        // Use package-level dependencies (injected via SetRepository/SetDetectionService in main.go)
        // These are defined in add.go:22-26
        if repository == nil {
            return fmt.Errorf("repository not initialized")
        }

        if detectionService == nil {
            return fmt.Errorf("detection service not initialized")
        }

        // Get all projects
        projects, err := repository.FindAll(ctx)
        if err != nil {
            return fmt.Errorf("failed to load projects: %w", err)
        }

        if len(projects) == 0 {
            fmt.Fprintln(cmd.OutOrStdout(), "No projects to refresh.")
            return nil
        }

        var refreshedCount, failedCount int
        for _, project := range projects {
            result, err := detectionService.Detect(ctx, project.Path)
            if err != nil {
                slog.Debug("detection failed", "project", project.Name, "error", err)
                failedCount++
                continue
            }

            project.DetectedMethod = result.Method
            project.CurrentStage = result.Stage
            project.Confidence = result.Confidence
            project.DetectionReasoning = result.Reasoning
            project.UpdatedAt = time.Now()

            if err := repository.Save(ctx, project); err != nil {
                slog.Debug("save failed", "project", project.Name, "error", err)
                failedCount++
                continue
            }

            refreshedCount++
        }

        // AC3: Only return error if ALL projects fail
        if refreshedCount == 0 && failedCount > 0 {
            return fmt.Errorf("all %d projects failed to refresh", failedCount)
        }

        // Success output
        fmt.Fprintf(cmd.OutOrStdout(), "Refreshed %d projects", refreshedCount)
        if failedCount > 0 {
            fmt.Fprintf(cmd.OutOrStdout(), " (%d failed)", failedCount)
        }
        fmt.Fprintln(cmd.OutOrStdout())

        return nil
    }
    ```

  - [x] 4.2 Verify command registration:
    - `init()` registers with `RootCmd` (production)
    - `RegisterRefreshCommand()` for test isolation
    - Uses `cmd.OutOrStdout()` for testable output

- [x] **Task 5: Add tests** (AC: all)

  - [x] 5.1 Create `internal/adapters/tui/model_refresh_test.go`:
    ```go
    package tui

    import (
        "context"
        "testing"

        tea "github.com/charmbracelet/bubbletea"

        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
        "github.com/JeiKeiLim/vibe-dash/internal/core/ports"
    )

    // mockDetector implements ports.Detector for testing
    type mockDetector struct {
        detectFunc func(ctx context.Context, path string) (*domain.DetectionResult, error)
    }

    func (m *mockDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
        return m.detectFunc(ctx, path)
    }

    func (m *mockDetector) DetectMultiple(ctx context.Context, path string) ([]*domain.DetectionResult, error) {
        return nil, nil
    }

    func TestModel_RefreshKey_StartsRefresh(t *testing.T) {
        // Setup: Model with mock detector and projects
        repo := newMockRepository() // From existing test helpers
        repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test"}}

        m := NewModel(repo)
        m.projects = repo.projects
        m.SetDetectionService(&mockDetector{
            detectFunc: func(ctx context.Context, path string) (*domain.DetectionResult, error) {
                return &domain.DetectionResult{Method: "test", Stage: "planning"}, nil
            },
        })

        // Action: Send 'r' key
        newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
        updated := newModel.(Model)

        // Assert
        if !updated.isRefreshing {
            t.Error("expected isRefreshing to be true after pressing 'r'")
        }
        if cmd == nil {
            t.Error("expected command to be returned")
        }
    }

    func TestModel_RefreshKey_DisabledWithoutDetectionService(t *testing.T) {
        // Setup: Model WITHOUT detection service
        repo := newMockRepository()
        m := NewModel(repo)
        m.projects = []*domain.Project{{ID: "1"}}
        // detectionService is nil

        // Action
        newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
        updated := newModel.(Model)

        // Assert: Should NOT start refresh, should set error
        if updated.isRefreshing {
            t.Error("expected isRefreshing to be false when no detection service")
        }
        if updated.refreshError == "" {
            t.Error("expected refreshError to be set")
        }
    }

    func TestModel_RefreshKey_IgnoredWhenRefreshing(t *testing.T) {
        // Setup: Model already refreshing
        repo := newMockRepository()
        m := NewModel(repo)
        m.isRefreshing = true
        m.SetDetectionService(&mockDetector{})

        // Action
        newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
        updated := newModel.(Model)

        // Assert: No new command, still refreshing
        if cmd != nil {
            t.Error("expected nil command when already refreshing")
        }
        if !updated.isRefreshing {
            t.Error("expected isRefreshing to remain true")
        }
    }

    func TestModel_NavigationDuringRefresh(t *testing.T) {
        // Setup: Model with projects, currently refreshing
        repo := newMockRepository()
        m := NewModel(repo)
        m.projects = []*domain.Project{{ID: "1"}, {ID: "2"}}
        m.projectList = components.NewProjectListModel(m.projects, 80, 24)
        m.isRefreshing = true

        // Action: Press 'j' to navigate down
        newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
        updated := newModel.(Model)

        // Assert: Navigation should work (selection changed)
        if updated.projectList.SelectedIndex() == 0 {
            t.Error("expected navigation to work during refresh")
        }
    }
    ```

  - [x] 5.2 Create `internal/adapters/cli/refresh_test.go`:
    ```go
    package cli_test

    import (
        "bytes"
        "context"
        "testing"

        "github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
        "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
    )

    func TestRefreshCmd_NoProjects(t *testing.T) {
        // Setup
        root := cli.NewRootCmd()
        cli.RegisterRefreshCommand(root)
        cli.SetRepository(&mockRepo{projects: nil})
        cli.SetDetectionService(&mockDetector{})

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"refresh"})

        // Execute
        err := root.Execute()

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "No projects to refresh") {
            t.Errorf("expected 'No projects' message, got: %s", out.String())
        }
    }

    func TestRefreshCmd_Success(t *testing.T) {
        // Setup with mock that returns success
        projects := []*domain.Project{
            {ID: "1", Path: "/test1", Name: "test1"},
            {ID: "2", Path: "/test2", Name: "test2"},
        }
        root := cli.NewRootCmd()
        cli.RegisterRefreshCommand(root)
        cli.SetRepository(&mockRepo{projects: projects})
        cli.SetDetectionService(&mockDetector{
            detectResult: &domain.DetectionResult{Method: "bmad", Stage: "planning"},
        })

        var out bytes.Buffer
        root.SetOut(&out)
        root.SetArgs([]string{"refresh"})

        // Execute
        err := root.Execute()

        // Assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !strings.Contains(out.String(), "Refreshed 2 projects") {
            t.Errorf("expected success message, got: %s", out.String())
        }
    }

    func TestRefreshCmd_PartialFailure(t *testing.T) {
        // Setup: First project succeeds, second fails detection
        // ... similar setup with mock returning error for second project
        // Assert: Output shows "Refreshed 1 projects (1 failed)"
    }

    func TestRefreshCmd_AllFail(t *testing.T) {
        // Setup: Detection fails for all projects
        // Assert: Returns error, exit code via MapErrorToExitCode
    }
    ```

  - [x] 5.3 Add status bar refresh tests to `components/status_bar_test.go`:
    ```go
    func TestStatusBar_RefreshingState(t *testing.T) {
        s := NewStatusBarModel(100)
        s.SetRefreshing(true, 2, 5)

        view := s.View()

        if !strings.Contains(view, "Refreshing...") {
            t.Error("expected 'Refreshing...' in output")
        }
        if !strings.Contains(view, "2/5") {
            t.Error("expected progress '2/5' in output")
        }
    }

    func TestStatusBar_RefreshComplete(t *testing.T) {
        s := NewStatusBarModel(100)
        s.SetRefreshComplete("Refreshed 3 projects")

        view := s.View()

        if !strings.Contains(view, "Refreshed 3 projects") {
            t.Error("expected completion message in output")
        }
    }

    func TestStatusBar_RefreshComplete_Cleared(t *testing.T) {
        s := NewStatusBarModel(100)
        s.SetRefreshComplete("Refreshed 3 projects")
        s.SetRefreshComplete("") // Clear

        view := s.View()

        if strings.Contains(view, "Refreshed") {
            t.Error("expected completion message to be cleared")
        }
    }
    ```

  - [x] 5.4 Run verification:
    ```bash
    make test   # All tests pass
    make lint   # No lint errors
    make build  # Successful build
    ```

## Dev Notes

### Current State Analysis

**KeyRefresh constant:** Already exists in `keys.go` as `KeyRefresh = "r"`. No changes needed.

**DetectionService:** Exists in `core/services/detection_service.go`. Implements `ports.Detector` interface. Currently NOT wired into TUI but IS wired into CLI via `add.go:26`.

**Status Bar:** `components/status_bar.go` has counts display. Needs refresh state additions.

**Model:** Has `loadProjectsCmd()` pattern that can be reused for refresh completion reload.

### Detection Service Integration - CRITICAL

**The detection service is already set up for CLI commands** via `add.go`:
```go
// add.go:22-26
var repository ports.ProjectRepository
var detectionService ports.Detector

func SetRepository(repo ports.ProjectRepository) { ... }
func SetDetectionService(svc ports.Detector) { ... }
```

These are initialized in `main.go` before `cli.Execute()` is called.

**For this story:**
1. **CLI refresh:** Reuse existing `repository` and `detectionService` package variables
2. **TUI refresh:** Add optional setter `SetDetectionService()` to Model, wire through `tui.Run`

**DO NOT create new repository/detection instances in refresh.go** - this creates an empty registry with no detectors, causing detection to always return "unknown".

### Progress Updates Limitation

**Note:** Real-time progress updates (N/M) during refresh are architecturally complex with Bubble Tea's message-passing model. The current implementation:
- Sets initial progress to 0/total when starting
- Shows final result on completion

To implement real-time updates, would need:
- Channel-based communication from tea.Cmd
- Multiple sequential commands with `tea.Batch`
- This is a nice-to-have enhancement, not MVP-critical

### Async Pattern

Follow existing pattern from `loadProjectsCmd` and `validatePathsCmd`:

```go
// Command returns a Msg
func (m Model) refreshProjectsCmd() tea.Cmd {
    return func() tea.Msg {
        // ... do work
        return refreshCompleteMsg{...}
    }
}

// Update handles the Msg
case refreshCompleteMsg:
    m.isRefreshing = false
    // ... update state
```

### Navigation During Refresh (AC5)

Navigation works during refresh because:
- `handleKeyMsg` handles j/k before checking refresh state
- `refreshProjectsCmd` is async via tea.Cmd
- Model updates are atomic per Update call

### Error Handling Strategy

Per AC3:
- Log individual project failures (slog.Debug)
- Continue to next project on failure
- Report "Refreshed N projects (M failed)" on partial success
- Only return error if ALL projects fail

### Project Context Compliance

Per `docs/project-context.md`:
- Context first: All service methods use `ctx context.Context`
- Error wrapping: Use `fmt.Errorf("...: %w", err)`
- Log at handling site: Log errors where handled
- Co-locate tests: `*_test.go` next to source

## Dev Agent Record

### Context Reference

Story context analyzed from:
- docs/epics.md (Story 3.6 requirements - lines 1316-1350)
- docs/architecture.md (DetectionService, ports/services patterns)
- docs/project-context.md (Go conventions, testing rules)
- internal/adapters/tui/model.go (Current Model structure, Update pattern)
- internal/adapters/tui/keys.go (KeyRefresh already defined)
- internal/adapters/tui/components/status_bar.go (Current status bar implementation)
- internal/core/services/detection_service.go (Detection interface)
- internal/core/ports/detector.go (Detector interface)
- internal/core/ports/repository.go (ProjectRepository interface)
- internal/adapters/cli/root.go (CLI command registration pattern)
- docs/sprint-artifacts/stories/epic-3/3-5-help-overlay.md (Previous story patterns)
- Git history: Stories 3.1-3.5 implementation patterns

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Story drafting phase.

### Completion Notes List

- All ACs implemented and tested
- TUI refresh via 'r' key triggers async detection rescan
- CLI refresh via `vibe refresh` command
- Status bar shows "Refreshing... (N/M)" during refresh
- Message clears after 3 seconds via tea.Tick timer
- Navigation works during refresh (non-blocking)
- Partial success handling: continues on individual failures
- Detection service injected via setter pattern (SetDetectionService)

### File List

**Modified:**
- `internal/adapters/tui/model.go` - Add refresh state, messages, handlers, SetDetectionService()
- `internal/adapters/tui/app.go` - Update Run() to accept optional detector parameter
- `internal/adapters/tui/components/status_bar.go` - Add refresh state display methods
- `internal/adapters/tui/components/status_bar_test.go` - Add refresh state tests
- `internal/adapters/cli/root.go` - Wire detection service to tui.Run()

**Created:**
- `internal/adapters/cli/refresh.go` - CLI refresh command (uses existing DI pattern)
- `internal/adapters/cli/refresh_test.go` - CLI refresh tests
- `internal/adapters/tui/model_refresh_test.go` - TUI refresh behavior tests

**Existing (Reference Only):**
- `internal/adapters/cli/add.go:22-37` - Package-level `repository` and `detectionService` variables + setters
- `internal/adapters/tui/keys.go:24` - `KeyRefresh = "r"` already defined

## Change Log

| Date | Change |
|------|--------|
| 2025-12-17 | Story created with ready-for-dev status by SM Agent (Bob) in YOLO mode. |
| 2025-12-17 | Story validated and improved: Fixed CLI DI pattern (C1), added clearRefreshMsgMsg type (C4), updated signature approach to use setter (C3), made tests more concrete (E3), added progress limitation note (E2). |
| 2025-12-17 | Implementation complete: All tasks and tests implemented. TUI and CLI refresh working. |
| 2025-12-17 | Code review fixes: Marked all subtasks complete, fixed tui.go→app.go filename, updated File List with status_bar_test.go, added completion notes. |
| 2025-12-17 | Code review fixes (code): Added [r] refresh to status bar shortcuts, fixed no-op test assertion in TestModel_ClearRefreshMsgMsg. Status → done. |
