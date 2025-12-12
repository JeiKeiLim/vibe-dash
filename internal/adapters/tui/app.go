package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application with the given context.
// The context is used for graceful shutdown on Ctrl+C.
func Run(ctx context.Context) error {
	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),  // Use alternate screen buffer
		tea.WithContext(ctx), // Respect context cancellation
	)

	_, err := p.Run()
	return err
}
