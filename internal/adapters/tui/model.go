package tui

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
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

	// Dependencies (injected)
	repository ports.ProjectRepository
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

// NewModel creates a new Model with default values.
// The repository parameter is used for project persistence operations.
func NewModel(repo ports.ProjectRepository) Model {
	return Model{
		ready:      false,
		showHelp:   false,
		viewMode:   viewModeNormal,
		repository: repo,
	}
}

// Init implements tea.Model. Returns a command to validate project paths.
func (m Model) Init() tea.Cmd {
	return m.validatePathsCmd()
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

// Update implements tea.Model. Handles messages and returns updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
			m.width = m.pendingWidth
			m.height = m.pendingHeight
			m.ready = true
			m.hasPendingResize = false
			// Update project list dimensions
			if len(m.projects) > 0 {
				m.projectList.SetSize(m.width, m.height)
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
			m.projectList = components.NewProjectListModel(m.projects, m.width, m.height)
		}
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
		return m, tea.Quit
	case KeyForceQuit:
		return m, tea.Quit
	case KeyHelp:
		m.showHelp = !m.showHelp
		return m, nil
	}

	// Forward key messages to project list when in normal mode
	if len(m.projects) > 0 {
		var cmd tea.Cmd
		m.projectList, cmd = m.projectList.Update(msg)
		return m, cmd
	}

	return m, nil
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

	// Render empty view if no projects (AC6)
	if len(m.projects) == 0 {
		return renderEmptyView(m.width, m.height)
	}

	// Render project list (Story 3.1)
	return m.renderDashboard()
}

// renderDashboard renders the main dashboard with project list.
func (m Model) renderDashboard() string {
	// For now, just the project list
	// Story 3.3 adds detail panel
	// Story 3.4 adds status bar
	return m.projectList.View()
}
