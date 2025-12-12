package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
}

// resizeTickMsg is used for resize debouncing.
type resizeTickMsg struct{}

// NewModel creates a new Model with default values.
func NewModel() Model {
	return Model{
		ready:    false,
		showHelp: false,
	}
}

// Init implements tea.Model. Returns nil (no initial command needed).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles messages and returns updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		}
		return m, nil
	}

	return m, nil
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

	if m.showHelp {
		return renderHelpOverlay(m.width, m.height)
	}

	return renderEmptyView(m.width, m.height)
}
