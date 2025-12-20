package tui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/project"
)

// Model represents the main TUI application state.
type Model struct {
	width            int  // Terminal width (from WindowSizeMsg)
	height           int  // Terminal height (from WindowSizeMsg)
	ready            bool // True after first WindowSizeMsg received
	showHelp         bool // Toggle help overlay
	hasPendingResize bool // True when resize is pending (debounce)
	pendingWidth     int  // Buffered width for debounced resize
	pendingHeight    int  // Buffered height for debounced resize

	// Path validation state
	viewMode          viewMode
	invalidProjects   []InvalidProject
	currentInvalidIdx int
	validationError   string // Error message to display in validation dialog

	// Project list state (Story 3.1)
	projects    []*domain.Project
	projectList components.ProjectListModel

	// Detail panel state (Story 3.3)
	showDetailPanel bool
	detailPanel     components.DetailPanelModel

	// Status bar state (Story 3.4)
	statusBar components.StatusBarModel

	// Refresh state (Story 3.6)
	isRefreshing    bool
	refreshTotal    int
	refreshProgress int
	refreshError    string

	// Note editing state (Story 3.7)
	isEditingNote  bool
	noteInput      textinput.Model // From charmbracelet/bubbles
	originalNote   string          // For cancel restoration
	noteEditTarget *domain.Project // Project being edited
	noteFeedback   string          // "✓ Note saved" or error message

	// Remove confirmation state (Story 3.9)
	isConfirmingRemove bool
	confirmTarget      *domain.Project // Project pending removal

	// Dependencies (injected)
	repository       ports.ProjectRepository
	detectionService ports.Detector        // Optional - may be nil if not wired
	waitingDetector  ports.WaitingDetector // Story 4.5: Optional - for WAITING indicator

	// Story 4.6: File watcher for real-time dashboard updates
	fileWatcher          ports.FileWatcher
	eventCh              <-chan ports.FileEvent
	watchCtx             context.Context
	watchCancel          context.CancelFunc
	fileWatcherAvailable bool // false if watcher failed to start
}

// resizeTickMsg is used for resize debouncing.
type resizeTickMsg struct{}

// Message types for async validation operations
type validationCompleteMsg struct {
	invalidProjects []InvalidProject
}

type validationErrorMsg struct {
	err error
}

type deleteProjectMsg struct {
	projectID string
	err       error
}

type moveProjectMsg struct {
	projectID string
	newPath   string
	err       error
}

type keepProjectMsg struct {
	projectID string
	err       error
}

// ProjectsLoadedMsg is sent when projects are loaded from the repository.
type ProjectsLoadedMsg struct {
	projects []*domain.Project
	err      error
}

// refreshCompleteMsg signals refresh is complete (Story 3.6).
type refreshCompleteMsg struct {
	refreshedCount int
	failedCount    int
	err            error // Only set if ALL projects failed
}

// clearRefreshMsgMsg signals to clear the refresh completion message (Story 3.6).
// Sent after 3-second timer expires.
type clearRefreshMsgMsg struct{}

// noteSavedMsg signals note was saved successfully (Story 3.7).
type noteSavedMsg struct {
	projectID string
	newNote   string
}

// noteSaveErrorMsg signals note save failed (Story 3.7).
type noteSaveErrorMsg struct {
	err error
}

// clearNoteFeedbackMsg signals to clear note feedback message (Story 3.7).
type clearNoteFeedbackMsg struct{}

// favoriteSavedMsg signals favorite was toggled successfully (Story 3.8).
type favoriteSavedMsg struct {
	projectID  string
	isFavorite bool
}

// favoriteSaveErrorMsg signals favorite save failed (Story 3.8).
type favoriteSaveErrorMsg struct {
	err error
}

// clearFavoriteFeedbackMsg signals to clear favorite feedback message (Story 3.8).
type clearFavoriteFeedbackMsg struct{}

// removeConfirmedMsg signals project removal was confirmed (Story 3.9).
type removeConfirmedMsg struct {
	projectID   string
	projectName string
	err         error
}

// removeConfirmTimeoutMsg signals confirmation dialog timed out (Story 3.9).
type removeConfirmTimeoutMsg struct{}

// clearRemoveFeedbackMsg signals to clear remove feedback message (Story 3.9).
type clearRemoveFeedbackMsg struct{}

// tickMsg is sent every 60 seconds to trigger timestamp recalculation (Story 4.2, AC4).
type tickMsg time.Time

// fileEventMsg wraps file system events for Bubble Tea message passing (Story 4.6).
type fileEventMsg struct {
	Path      string
	Operation ports.FileOperation
	Timestamp time.Time
}

// fileWatcherErrorMsg signals a file watcher error (Story 4.6).
type fileWatcherErrorMsg struct {
	err error
}

// NewModel creates a new Model with default values.
// The repository parameter is used for project persistence operations.
func NewModel(repo ports.ProjectRepository) Model {
	return Model{
		ready:           false,
		showHelp:        false,
		showDetailPanel: false, // Default closed, set based on height in resizeTickMsg
		viewMode:        viewModeNormal,
		repository:      repo,
		statusBar:       components.NewStatusBarModel(0), // Width set in resizeTickMsg
	}
}

// SetDetectionService sets the detection service for refresh operations.
// This is optional - if not set, refresh will show "Detection service not available".
func (m *Model) SetDetectionService(svc ports.Detector) {
	m.detectionService = svc
}

// SetWaitingDetector sets the waiting detector for WAITING indicators (Story 4.5).
// This is optional - if not set, waiting indicators will not be shown.
func (m *Model) SetWaitingDetector(detector ports.WaitingDetector) {
	m.waitingDetector = detector
}

// SetFileWatcher sets the file watcher for real-time updates (Story 4.6).
// This is optional - if not set, file watching is disabled.
func (m *Model) SetFileWatcher(watcher ports.FileWatcher) {
	m.fileWatcher = watcher
	m.fileWatcherAvailable = true // Assume available until proven otherwise
}

// isProjectWaiting wraps WaitingDetector.IsWaiting for component callbacks.
// Uses context.Background() since Bubble Tea Render() doesn't provide ctx.
// Story 4.5: Returns false if detector is nil.
func (m Model) isProjectWaiting(p *domain.Project) bool {
	if m.waitingDetector == nil {
		return false
	}
	return m.waitingDetector.IsWaiting(context.Background(), p)
}

// getWaitingDuration wraps WaitingDetector.WaitingDuration for component callbacks.
// Uses context.Background() since Bubble Tea Render() doesn't provide ctx.
// Story 4.5: Returns 0 if detector is nil.
func (m Model) getWaitingDuration(p *domain.Project) time.Duration {
	if m.waitingDetector == nil {
		return 0
	}
	return m.waitingDetector.WaitingDuration(context.Background(), p)
}

// shouldShowDetailPanelByDefault returns true if detail panel should be open by default
// based on terminal height. Per AC7: >= HeightThresholdTall (35) rows = open, otherwise closed.
func shouldShowDetailPanelByDefault(height int) bool {
	return height >= HeightThresholdTall
}

// isNarrowWidth returns true if terminal width is in narrow range (60-79).
// Used for Story 3.10 AC2 narrow warning display.
func isNarrowWidth(width int) bool {
	return width >= MinWidth && width < 80
}

// isWideWidth returns true if terminal width exceeds MaxContentWidth (>120).
// Used for Story 3.10 AC4 content capping and centering.
func isWideWidth(width int) bool {
	return width > MaxContentWidth
}

// statusBarHeight returns the status bar height based on terminal height.
// Returns 1 for condensed mode (height < 20), 2 for normal mode.
func statusBarHeight(height int) int {
	if height < MinHeight {
		return 1
	}
	return 2
}

// Init implements tea.Model. Returns commands to validate project paths and start periodic tick.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.validatePathsCmd(),
		tickCmd(), // Start periodic timestamp refresh (Story 4.2, AC4)
	)
}

// validatePathsCmd creates a command that validates all project paths.
func (m Model) validatePathsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		invalid, err := ValidateProjectPaths(ctx, m.repository)
		if err != nil {
			return validationErrorMsg{err}
		}
		return validationCompleteMsg{invalid}
	}
}

// loadProjectsCmd creates a command that loads all projects from the repository.
func (m Model) loadProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		projects, err := m.repository.FindAll(ctx)
		return ProjectsLoadedMsg{projects: projects, err: err}
	}
}

// tickCmd returns a command that ticks every 60 seconds for timestamp refresh (Story 4.2, AC4).
func tickCmd() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// waitForNextFileEventCmd waits for the next file event from the stored channel (Story 4.6).
// Returns nil if channel is not set or context is cancelled.
func (m Model) waitForNextFileEventCmd() tea.Cmd {
	if m.eventCh == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case <-m.watchCtx.Done():
			return nil // Context cancelled, stop listening
		case event, ok := <-m.eventCh:
			if !ok {
				return fileWatcherErrorMsg{err: fmt.Errorf("watcher channel closed")}
			}
			return fileEventMsg{
				Path:      event.Path,
				Operation: event.Operation,
				Timestamp: event.Timestamp,
			}
		}
	}
}

// startRefresh initiates async refresh of all projects (Story 3.6).
func (m Model) startRefresh() (tea.Model, tea.Cmd) {
	m.isRefreshing = true
	m.refreshTotal = len(m.projects)
	m.refreshProgress = 0
	m.refreshError = ""
	m.statusBar.SetRefreshing(true, 0, m.refreshTotal)

	return m, m.refreshProjectsCmd()
}

// refreshProjectsCmd creates a command that rescans all projects (Story 3.6).
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

// Update implements tea.Model. Handles messages and returns updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Route to note editing handler when in note editing mode (Story 3.7)
		if m.isEditingNote {
			return m.handleNoteEditingKeyMsg(msg)
		}
		// Route to remove confirmation handler when in confirmation mode (Story 3.9)
		if m.isConfirmingRemove {
			return m.handleRemoveConfirmationKeyMsg(msg)
		}
		// Route to validation handler when in validation mode
		if m.viewMode == viewModeValidation {
			return m.handleValidationKeyMsg(msg)
		}
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		// Debounce resize events (50ms per UX spec)
		m.hasPendingResize = true
		m.pendingWidth = msg.Width
		m.pendingHeight = msg.Height
		return m, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
			return resizeTickMsg{}
		})

	case resizeTickMsg:
		if m.hasPendingResize {
			wasReady := m.ready
			m.width = m.pendingWidth
			m.height = m.pendingHeight
			m.ready = true
			m.hasPendingResize = false

			// Set initial detail panel visibility only on first ready (AC4, AC5, AC6)
			if !wasReady {
				m.showDetailPanel = shouldShowDetailPanelByDefault(m.height)
				m.detailPanel.SetVisible(m.showDetailPanel)
			}

			// FIRST: Set condensed mode (affects status bar height) - Story 3.10 AC5
			m.statusBar.SetCondensed(m.height < MinHeight)

			// Calculate effective width for components (cap at MaxContentWidth - Story 3.10 AC4)
			effectiveWidth := m.width
			if isWideWidth(m.width) {
				effectiveWidth = MaxContentWidth
			}

			// Update status bar width with effective width (Story 3.4, 3.10)
			m.statusBar.SetWidth(effectiveWidth)

			// Calculate content height using helper (Story 3.10 AC5)
			contentHeight := m.height - statusBarHeight(m.height)

			// Update component dimensions with effective width (Story 3.10)
			if len(m.projects) > 0 {
				m.projectList.SetSize(effectiveWidth, contentHeight)
				m.detailPanel.SetSize(effectiveWidth, contentHeight)
			}
		}
		return m, nil

	case validationCompleteMsg:
		// Handle validation result from Init()
		if len(msg.invalidProjects) > 0 {
			m.viewMode = viewModeValidation
			m.invalidProjects = msg.invalidProjects
			m.currentInvalidIdx = 0
			return m, nil
		}
		// No invalid projects, load projects
		return m, m.loadProjectsCmd()

	case validationErrorMsg:
		// Log error and continue to normal view (non-fatal)
		slog.Error("Path validation failed", "error", msg.err)
		return m, nil

	case deleteProjectMsg:
		// Handle async delete completion
		if msg.err != nil {
			m.validationError = "Failed to delete: " + msg.err.Error()
			return m, nil
		}
		m.validationError = "" // Clear error on success
		m, cmd := m.advanceToNextInvalid()
		return m, cmd

	case moveProjectMsg:
		// Handle async move completion
		if msg.err != nil {
			m.validationError = "Failed to move: " + msg.err.Error()
			return m, nil
		}
		m.validationError = "" // Clear error on success
		m, cmd := m.advanceToNextInvalid()
		return m, cmd

	case keepProjectMsg:
		// Handle async keep completion
		if msg.err != nil {
			m.validationError = "Failed to keep: " + msg.err.Error()
			return m, nil
		}
		m.validationError = "" // Clear error on success
		m, cmd := m.advanceToNextInvalid()
		return m, cmd

	case ProjectsLoadedMsg:
		// Handle project loading result
		if msg.err != nil {
			slog.Error("Failed to load projects", "error", msg.err)
			m.projects = nil
			return m, nil
		}
		m.projects = msg.projects
		if len(m.projects) > 0 {
			contentHeight := m.height - 2 // Reserve 2 lines for status bar
			m.projectList = components.NewProjectListModel(m.projects, m.width, contentHeight)

			// Story 4.5: Wire waiting callbacks to project list delegate
			m.projectList.SetDelegateWaitingCallbacks(m.isProjectWaiting, m.getWaitingDuration)

			// Initialize detail panel with selected project
			m.detailPanel = components.NewDetailPanelModel(m.width, contentHeight)
			m.detailPanel.SetProject(m.projectList.SelectedProject())
			m.detailPanel.SetVisible(m.showDetailPanel)

			// Story 4.5: Wire waiting callbacks to detail panel
			m.detailPanel.SetWaitingCallbacks(m.isProjectWaiting, m.getWaitingDuration)

			// Update status bar counts (Story 3.4, 4.5)
			active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
			m.statusBar.SetCounts(active, hibernated, waiting)

			// Story 4.6: Start file watcher for real-time updates
			if m.fileWatcher != nil && m.fileWatcherAvailable {
				// Collect project paths
				paths := make([]string, len(m.projects))
				for i, p := range m.projects {
					paths[i] = p.Path
				}

				// Create watch context for cancellation
				m.watchCtx, m.watchCancel = context.WithCancel(context.Background())

				// Start watching
				eventCh, err := m.fileWatcher.Watch(m.watchCtx, paths)
				if err != nil {
					slog.Warn("failed to start file watcher", "error", err)
					m.fileWatcherAvailable = false
					m.statusBar.SetWatcherWarning("⚠️ File watching unavailable")
				} else {
					m.eventCh = eventCh
					// Epic 4 Hotfix H3: Log when file watcher starts successfully
					slog.Debug("file watcher started", "project_count", len(paths))
					return m, m.waitForNextFileEventCmd()
				}
			}
		}
		return m, nil

	case refreshCompleteMsg:
		// Handle refresh completion (Story 3.6)
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
		// Clear refresh completion message after 3 seconds (Story 3.6)
		m.statusBar.SetRefreshComplete("")
		return m, nil

	case noteSavedMsg:
		// Update local project state (Story 3.7)
		for _, p := range m.projects {
			if p.ID == msg.projectID {
				p.Notes = msg.newNote
				break
			}
		}
		// Update detail panel
		m.detailPanel.SetProject(m.projectList.SelectedProject())
		// Set feedback message and display in status bar (AC2: feedback shows "✓ Note saved")
		m.noteFeedback = "✓ Note saved"
		m.statusBar.SetRefreshComplete(m.noteFeedback)
		// Clear after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNoteFeedbackMsg{}
		})

	case noteSaveErrorMsg:
		// Handle note save failure (Story 3.7)
		m.noteFeedback = "✗ Failed to save note"
		m.statusBar.SetRefreshComplete(m.noteFeedback)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNoteFeedbackMsg{}
		})

	case clearNoteFeedbackMsg:
		// Clear note feedback message (Story 3.7)
		m.noteFeedback = ""
		m.statusBar.SetRefreshComplete("")
		return m, nil

	case favoriteSavedMsg:
		// Update local project state (Story 3.8)
		for _, p := range m.projects {
			if p.ID == msg.projectID {
				p.IsFavorite = msg.isFavorite
				break
			}
		}
		// Update detail panel
		m.detailPanel.SetProject(m.projectList.SelectedProject())
		// Set feedback message
		var feedback string
		if msg.isFavorite {
			feedback = "⭐ Favorited"
		} else {
			feedback = "☆ Unfavorited"
		}
		m.statusBar.SetRefreshComplete(feedback)
		// Clear after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearFavoriteFeedbackMsg{}
		})

	case favoriteSaveErrorMsg:
		m.statusBar.SetRefreshComplete("✗ Failed to toggle favorite")
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearFavoriteFeedbackMsg{}
		})

	case clearFavoriteFeedbackMsg:
		m.statusBar.SetRefreshComplete("")
		return m, nil

	case removeConfirmedMsg:
		// Handle async delete completion (Story 3.9)
		if msg.err != nil {
			m.statusBar.SetRefreshComplete("✗ Failed to remove: " + msg.err.Error())
			return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return clearRemoveFeedbackMsg{}
			})
		}
		// Remove from local projects slice
		var newProjects []*domain.Project
		for _, p := range m.projects {
			if p.ID != msg.projectID {
				newProjects = append(newProjects, p)
			}
		}
		m.projects = newProjects
		// Update project list component
		m.projectList.SetProjects(newProjects)
		// Update detail panel with new selection
		m.detailPanel.SetProject(m.projectList.SelectedProject())
		// Update status bar counts (Story 4.5)
		active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
		m.statusBar.SetCounts(active, hibernated, waiting)
		// Show success feedback
		m.statusBar.SetRefreshComplete("✓ Removed: " + msg.projectName)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearRemoveFeedbackMsg{}
		})

	case removeConfirmTimeoutMsg:
		// Auto-cancel confirmation after 30 seconds (Story 3.9, AC4)
		if m.isConfirmingRemove {
			m.isConfirmingRemove = false
			m.confirmTarget = nil
		}
		return m, nil

	case clearRemoveFeedbackMsg:
		m.statusBar.SetRefreshComplete("")
		return m, nil

	case tickMsg:
		// Periodic tick to trigger timestamp recalculation (Story 4.2, AC4).
		// The View() method calls FormatRelativeTime which recalculates on each render.
		// We just need to schedule the next tick.
		return m, tickCmd()

	case fileEventMsg:
		// Story 4.6: Handle file system event
		m.handleFileEvent(msg)
		// Re-subscribe to wait for next event
		return m, m.waitForNextFileEventCmd()

	case fileWatcherErrorMsg:
		// Story 4.6: Handle file watcher error (AC3)
		slog.Warn("file watcher error", "error", msg.err)
		m.fileWatcherAvailable = false
		m.statusBar.SetWatcherWarning("⚠️ File watching unavailable")
		return m, nil
	}

	return m, nil
}

// advanceToNextInvalid moves to the next invalid project or switches to normal view.
// Returns the model and a command to load projects if validation is complete.
func (m Model) advanceToNextInvalid() (Model, tea.Cmd) {
	m.currentInvalidIdx++
	if m.currentInvalidIdx >= len(m.invalidProjects) {
		m.viewMode = viewModeNormal
		// Load projects after validation is complete
		return m, m.loadProjectsCmd()
	}
	return m, nil
}

// handleValidationKeyMsg processes keyboard input in validation mode.
func (m Model) handleValidationKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Allow quit even in validation mode
	switch msg.String() {
	case KeyQuit, KeyForceQuit:
		// Story 4.6: Clean up file watcher on quit (AC10)
		if m.watchCancel != nil {
			m.watchCancel()
		}
		if m.fileWatcher != nil {
			m.fileWatcher.Close()
		}
		return m, tea.Quit
	}

	if m.currentInvalidIdx >= len(m.invalidProjects) {
		// All handled, switch to normal view
		m.viewMode = viewModeNormal
		return m, nil
	}

	currentProject := m.invalidProjects[m.currentInvalidIdx].Project

	switch strings.ToLower(msg.String()) {
	case "d":
		return m, m.deleteProjectCmd(currentProject.ID)
	case "m":
		return m, m.moveProjectCmd(currentProject)
	case "k":
		return m, m.keepProjectCmd(currentProject)
	}

	return m, nil
}

// deleteProjectCmd creates a command to delete a project (AC2).
func (m Model) deleteProjectCmd(projectID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.repository.Delete(ctx, projectID)
		return deleteProjectMsg{projectID: projectID, err: err}
	}
}

// moveProjectCmd creates a command to move a project to current directory (AC3).
func (m Model) moveProjectCmd(project *domain.Project) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return moveProjectMsg{projectID: project.ID, err: err}
		}

		// Validate new path exists
		canonicalPath, err := filesystem.CanonicalPath(cwd)
		if err != nil {
			return moveProjectMsg{projectID: project.ID, err: err}
		}

		// Delete old project entry (ID is path-based, so it will change)
		// Ignore delete errors - old entry may not exist
		_ = m.repository.Delete(ctx, project.ID)

		// Update project with new path
		project.Path = canonicalPath
		project.ID = domain.GenerateID(canonicalPath) // ID is path-based
		project.PathMissing = false
		project.UpdatedAt = time.Now()

		// NOTE: detectionService is optional until Epic 2 detection is wired.
		// The Move action works without re-detection.

		// Save to repository
		if err := m.repository.Save(ctx, project); err != nil {
			return moveProjectMsg{projectID: project.ID, err: err}
		}

		return moveProjectMsg{projectID: project.ID, newPath: canonicalPath}
	}
}

// keepProjectCmd creates a command to keep a project with PathMissing flag (AC4).
func (m Model) keepProjectCmd(project *domain.Project) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Set PathMissing flag
		project.PathMissing = true
		project.UpdatedAt = time.Now()

		// Save to repository
		if err := m.repository.Save(ctx, project); err != nil {
			return keepProjectMsg{projectID: project.ID, err: err}
		}

		return keepProjectMsg{projectID: project.ID}
	}
}

// handleKeyMsg processes keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If help is showing, any key closes it (except '?' which toggles)
	if m.showHelp {
		if msg.String() == KeyHelp {
			m.showHelp = false
			return m, nil
		}
		// Any other key closes help
		m.showHelp = false
		return m, nil
	}

	switch msg.String() {
	case KeyQuit:
		// Story 4.6: Clean up file watcher on quit (AC10)
		if m.watchCancel != nil {
			m.watchCancel()
		}
		if m.fileWatcher != nil {
			m.fileWatcher.Close()
		}
		return m, tea.Quit
	case KeyForceQuit:
		// Story 4.6: Clean up file watcher on force quit (AC10)
		if m.watchCancel != nil {
			m.watchCancel()
		}
		if m.fileWatcher != nil {
			m.fileWatcher.Close()
		}
		return m, tea.Quit
	case KeyHelp:
		m.showHelp = !m.showHelp
		return m, nil
	case KeyEscape:
		// No-op in normal mode - future stories (3.7, 3.9) will add prompt cancellation
		return m, nil
	case KeyDetail:
		m.showDetailPanel = !m.showDetailPanel
		m.detailPanel.SetVisible(m.showDetailPanel)
		return m, nil
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
	case KeyNotes:
		// Story 3.7: Open note editor for selected project
		if m.isEditingNote {
			return m, nil // Ignore if already editing
		}
		if len(m.projects) == 0 {
			return m, nil // No project to edit
		}
		return m.startNoteEditing()
	case KeyFavorite:
		// Story 3.8: Toggle favorite status for selected project
		if len(m.projects) == 0 {
			return m, nil // No project to favorite
		}
		return m.toggleFavorite()
	case KeyRemove:
		// Story 3.9: Start remove confirmation for selected project
		if m.isConfirmingRemove {
			return m, nil // Already confirming
		}
		if len(m.projects) == 0 {
			return m, nil // No project to remove
		}
		return m.startRemoveConfirmation()
	}

	// Forward key messages to project list when in normal mode
	if len(m.projects) > 0 {
		var cmd tea.Cmd
		m.projectList, cmd = m.projectList.Update(msg)
		// Update detail panel with current selection (may have changed)
		m.detailPanel.SetProject(m.projectList.SelectedProject())
		return m, cmd
	}

	return m, nil
}

// startNoteEditing opens the note editor for the selected project (Story 3.7).
func (m Model) startNoteEditing() (tea.Model, tea.Cmd) {
	selected := m.projectList.SelectedProject()
	if selected == nil {
		return m, nil
	}

	m.isEditingNote = true
	m.noteEditTarget = selected
	m.originalNote = selected.Notes

	// Initialize text input with current note
	ti := textinput.New()
	ti.Placeholder = "Enter note..."
	ti.Focus()
	ti.CharLimit = 500 // Reasonable limit for notes
	ti.Width = m.width - 10
	ti.SetValue(selected.Notes)
	m.noteInput = ti

	return m, textinput.Blink
}

// handleNoteEditingKeyMsg processes keyboard input during note editing (Story 3.7).
func (m Model) handleNoteEditingKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Save note
		return m.saveNote()
	case tea.KeyEsc:
		// Cancel editing, restore original
		m.isEditingNote = false
		m.noteInput = textinput.Model{} // Clear
		return m, nil
	}

	// Forward to text input
	var cmd tea.Cmd
	m.noteInput, cmd = m.noteInput.Update(msg)
	return m, cmd
}

// saveNote saves the note to the repository (Story 3.7).
func (m Model) saveNote() (tea.Model, tea.Cmd) {
	newNote := strings.TrimSpace(m.noteInput.Value())
	m.isEditingNote = false

	return m, m.saveNoteCmd(m.noteEditTarget.ID, newNote)
}

// saveNoteCmd creates a command that saves the note to repository (Story 3.7).
func (m Model) saveNoteCmd(projectID, note string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Find project
		project, err := m.repository.FindByID(ctx, projectID)
		if err != nil {
			return noteSaveErrorMsg{err: err}
		}

		// Update note
		project.Notes = note
		project.UpdatedAt = time.Now()

		// Save
		if err := m.repository.Save(ctx, project); err != nil {
			return noteSaveErrorMsg{err: err}
		}

		return noteSavedMsg{projectID: projectID, newNote: note}
	}
}

// toggleFavorite toggles the favorite status of the selected project (Story 3.8).
func (m Model) toggleFavorite() (tea.Model, tea.Cmd) {
	selected := m.projectList.SelectedProject()
	if selected == nil {
		return m, nil
	}

	newFavorite := !selected.IsFavorite
	return m, m.saveFavoriteCmd(selected.ID, newFavorite)
}

// saveFavoriteCmd creates a command that saves the favorite status to repository (Story 3.8).
func (m Model) saveFavoriteCmd(projectID string, isFavorite bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Find project
		project, err := m.repository.FindByID(ctx, projectID)
		if err != nil {
			return favoriteSaveErrorMsg{err: err}
		}

		// Update favorite status
		project.IsFavorite = isFavorite
		project.UpdatedAt = time.Now()

		// Save
		if err := m.repository.Save(ctx, project); err != nil {
			return favoriteSaveErrorMsg{err: err}
		}

		return favoriteSavedMsg{projectID: projectID, isFavorite: isFavorite}
	}
}

// startRemoveConfirmation opens the remove confirmation dialog (Story 3.9).
func (m Model) startRemoveConfirmation() (tea.Model, tea.Cmd) {
	selected := m.projectList.SelectedProject()
	if selected == nil {
		return m, nil
	}

	m.isConfirmingRemove = true
	m.confirmTarget = selected

	// Start 30-second timeout timer (AC4)
	return m, tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return removeConfirmTimeoutMsg{}
	})
}

// handleRemoveConfirmationKeyMsg processes keyboard input during remove confirmation (Story 3.9).
func (m Model) handleRemoveConfirmationKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle Escape key by type (consistent with handleNoteEditingKeyMsg pattern)
	if msg.Type == tea.KeyEsc {
		m.isConfirmingRemove = false
		m.confirmTarget = nil
		return m, nil
	}

	// Handle character keys by string
	switch msg.String() {
	case "y", "Y":
		// Confirm removal
		m.isConfirmingRemove = false
		target := m.confirmTarget
		m.confirmTarget = nil

		// Use shared EffectiveName for display (Story 3.9 code review fix)
		projectName := project.EffectiveName(target)

		return m, m.removeProjectCmd(target.ID, projectName)

	case "n", "N":
		// Cancel removal
		m.isConfirmingRemove = false
		m.confirmTarget = nil
		return m, nil
	}

	// AC5: Ignore all other keys during confirmation
	return m, nil
}

// removeProjectCmd creates a command that deletes a project from repository (Story 3.9).
// Note: Reuses delete pattern from deleteProjectCmd but returns removeConfirmedMsg.
func (m Model) removeProjectCmd(projectID, projectName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.repository.Delete(ctx, projectID)
		return removeConfirmedMsg{projectID: projectID, projectName: projectName, err: err}
	}
}

// View implements tea.Model. Renders the UI to a string.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Check minimum terminal size
	if m.width < MinWidth || m.height < MinHeight {
		return renderTooSmallView(m.width, m.height)
	}

	// Render validation dialog when in validation mode (AC1, AC6)
	if m.viewMode == viewModeValidation && m.currentInvalidIdx < len(m.invalidProjects) {
		return renderValidationDialog(m.invalidProjects[m.currentInvalidIdx].Project, m.width, m.height, m.validationError)
	}

	// Render help overlay (overlays everything)
	if m.showHelp {
		return renderHelpOverlay(m.width, m.height)
	}

	// Render note editor dialog (overlays everything) (Story 3.7)
	if m.isEditingNote && m.noteEditTarget != nil {
		projectName := project.EffectiveName(m.noteEditTarget)
		return renderNoteEditor(projectName, m.noteInput, m.width, m.height)
	}

	// Render remove confirmation dialog (overlays everything) (Story 3.9)
	if m.isConfirmingRemove && m.confirmTarget != nil {
		projectName := project.EffectiveName(m.confirmTarget)
		return renderConfirmRemoveDialog(projectName, m.width, m.height)
	}

	// Render empty view if no projects (AC6)
	// Status bar is always visible per AC2, even in empty view
	if len(m.projects) == 0 {
		contentHeight := m.height - 2 // Reserve 2 lines for status bar
		emptyContent := renderEmptyView(m.width, contentHeight)
		statusBar := m.statusBar.View()
		return lipgloss.JoinVertical(lipgloss.Left, emptyContent, statusBar)
	}

	// Render project list (Story 3.1)
	return m.renderDashboard()
}

// renderDashboard renders the main dashboard with project list, optional detail panel, and status bar.
func (m Model) renderDashboard() string {
	// Calculate effective width (cap at MaxContentWidth for wide terminals - Story 3.10 AC4)
	// Note: isNarrowWidth (60-79) and isWideWidth (>120) are mutually exclusive
	effectiveWidth := m.width
	if isWideWidth(m.width) {
		effectiveWidth = MaxContentWidth
	}

	// Reserve lines for status bar using helper (Story 3.10 AC5)
	contentHeight := m.height - statusBarHeight(m.height)

	// Adjust content height if narrow warning is shown (Story 3.10 AC2)
	if isNarrowWidth(m.width) {
		contentHeight-- // Reserve 1 more line for warning
	}

	// Create a temporary model copy with effective width for rendering
	renderModel := m
	renderModel.width = effectiveWidth

	mainContent := renderModel.renderMainContent(contentHeight)

	// Build output parts
	var parts []string
	parts = append(parts, mainContent)

	// Add narrow warning if applicable (Story 3.10 AC2)
	if isNarrowWidth(m.width) {
		parts = append(parts, renderNarrowWarning(effectiveWidth))
	}

	// Use m.statusBar directly - width already set in resizeTickMsg
	parts = append(parts, m.statusBar.View())

	// Join content
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Center content if terminal is wider than MaxContentWidth (Story 3.10 AC4)
	if isWideWidth(m.width) {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)
	}

	return content
}

// renderMainContent renders the main content area (project list and optional detail panel).
func (m Model) renderMainContent(height int) string {
	// Show hint when terminal height MinHeight-HeightThresholdTall and detail panel closed (Story 3.10 AC6)
	// IMPORTANT: Use m.height (terminal height) not height parameter (contentHeight)
	// because AC6 defines behavior based on user-visible terminal size
	if m.height >= MinHeight && m.height < HeightThresholdTall && !m.showDetailPanel {
		hint := DimStyle.Render("Press [d] for details")
		return m.projectList.View() + "\n" + hint
	}

	// Full-width project list when detail panel is hidden
	if !m.showDetailPanel {
		return m.projectList.View()
	}

	// Split layout: list (60%) | detail (40%)
	listWidth := int(float64(m.width) * 0.6)
	detailWidth := m.width - listWidth - 1 // -1 for separator space

	// Create copies with updated sizes for this render
	projectList := m.projectList
	projectList.SetSize(listWidth, height)

	detailPanel := m.detailPanel
	detailPanel.SetSize(detailWidth, height)

	// Render side by side
	listView := projectList.View()
	detailView := detailPanel.View()

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}

// handleFileEvent processes a file system event and updates project state (Story 4.6).
func (m *Model) handleFileEvent(msg fileEventMsg) {
	// Find project by path prefix
	project := m.findProjectByPath(msg.Path)
	if project == nil {
		slog.Debug("event path not matched to project", "path", msg.Path)
		return
	}

	// Update repository (skip if nil - e.g., in tests without mocked repo)
	ctx := context.Background()
	if m.repository == nil {
		slog.Debug("repository is nil, skipping activity update", "project_id", project.ID)
		return
	}
	if err := m.repository.UpdateLastActivity(ctx, project.ID, msg.Timestamp); err != nil {
		slog.Warn("failed to update activity", "project_id", project.ID, "error", err)
		return
	}

	// Epic 4 Hotfix H3: Log successful activity update for debugging
	slog.Debug("activity updated", "project", project.Name, "path", msg.Path)

	// Update local state
	project.LastActivityAt = msg.Timestamp

	// Update detail panel if this is selected project
	if m.detailPanel.Project() != nil && m.detailPanel.Project().ID == project.ID {
		m.detailPanel.SetProject(project)
	}

	// Recalculate status bar (waiting may have cleared)
	active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
	m.statusBar.SetCounts(active, hibernated, waiting)
}

// findProjectByPath finds the project that owns the given file path (Story 4.6).
// Uses path prefix matching - returns the project whose Path is a prefix of eventPath.
func (m Model) findProjectByPath(eventPath string) *domain.Project {
	eventPath = strings.TrimSuffix(eventPath, "/")
	for _, p := range m.projects {
		projectPath := strings.TrimSuffix(p.Path, "/")
		if eventPath == projectPath || strings.HasPrefix(eventPath, projectPath+"/") {
			return p
		}
	}
	return nil
}
