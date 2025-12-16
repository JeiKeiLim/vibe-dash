package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// Run starts the TUI application with the given context.
// The context is used for graceful shutdown on Ctrl+C.
// The repository parameter is used for project persistence operations.
// The detector parameter is optional - if nil, refresh will be disabled.
func Run(ctx context.Context, repo ports.ProjectRepository, detector ports.Detector) error {
	m := NewModel(repo)
	if detector != nil {
		m.SetDetectionService(detector)
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),  // Use alternate screen buffer
		tea.WithContext(ctx), // Respect context cancellation
	)

	_, err := p.Run()
	return err
}
