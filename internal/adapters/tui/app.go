package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/emoji"
)

// Run starts the TUI application with the given context.
// The context is used for graceful shutdown on Ctrl+C.
// The repository parameter is used for project persistence operations.
// The detector parameter is optional - if nil, refresh will be disabled.
// The waitingDetector parameter is optional - if nil, waiting indicators are disabled (Story 4.5).
// The fileWatcher parameter is optional - if nil, real-time updates are disabled (Story 4.6).
// The detailLayout parameter controls detail panel position (Story 8.6):
// "vertical" (default) = side-by-side, "horizontal" = stacked (top/bottom).
// The config parameter is used for help overlay display (Story 8.7) - nil-safe.
// Note: Config passed as parameter to avoid cli→tui→cli import cycle.
func Run(ctx context.Context, repo ports.ProjectRepository, detector ports.Detector, waitingDetector ports.WaitingDetector, fileWatcher ports.FileWatcher, detailLayout string, config *ports.Config) error {
	// Story 8.9: Initialize emoji fallback system BEFORE TUI renders
	var useEmoji *bool
	if config != nil {
		useEmoji = config.UseEmoji
	}
	emoji.InitEmoji(useEmoji)

	m := NewModel(repo)
	if detector != nil {
		m.SetDetectionService(detector)
	}
	// Story 4.5: Wire waiting detector for WAITING indicator display
	if waitingDetector != nil {
		m.SetWaitingDetector(waitingDetector)
	}
	// Story 4.6: Wire file watcher for real-time dashboard updates
	if fileWatcher != nil {
		m.SetFileWatcher(fileWatcher)
	}
	// Story 8.6: Set detail panel layout mode from config
	m.SetDetailLayout(detailLayout)
	// Story 8.7: Set config for help overlay display
	m.SetConfig(config)

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),  // Use alternate screen buffer
		tea.WithContext(ctx), // Respect context cancellation
	)

	_, err := p.Run()
	return err
}
