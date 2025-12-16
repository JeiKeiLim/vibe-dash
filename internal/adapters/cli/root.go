package cli

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/sqlite"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui"
)

// RootCmd is the root command for the vibe CLI.
var RootCmd = &cobra.Command{
	Use:   "vibe",
	Short: "CLI dashboard for vibe coding projects",
	Long: `vibe-dash is a CLI dashboard for vibe coding projects.

Track AI-assisted coding project stages, detect when agents are waiting
for input, and manage your workflow across multiple projects.

Run 'vibe' with no arguments to launch the interactive dashboard.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("vibe-dash starting")

		// Initialize repository for path validation at launch
		repo, err := sqlite.NewSQLiteRepository("")
		if err != nil {
			slog.Error("Failed to initialize repository", "error", err)
			return
		}

		// Pass detection service to TUI for refresh functionality (Story 3.6)
		// Uses existing detectionService package variable from add.go
		if err := tui.Run(cmd.Context(), repo, detectionService); err != nil {
			slog.Error("TUI error", "error", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// It accepts a context for graceful shutdown support.
func Execute(ctx context.Context) error {
	return RootCmd.ExecuteContext(ctx)
}
