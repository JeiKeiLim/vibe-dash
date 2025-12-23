package cli

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui"
)

// RootCmd is the root command for the vibe CLI.
var RootCmd = &cobra.Command{
	Use:   "vibe",
	Short: "CLI dashboard for vibe coding projects",
	Long: `vibe-dash is a CLI dashboard for vibe coding projects.

Track AI-assisted coding project stages, detect when agents are waiting
for input, and manage your workflow across multiple projects.

Run 'vibe' with no arguments to launch the interactive dashboard.

Exit Codes:
  0  Success
  1  General error (unhandled, user decision needed)
  2  Project not found
  3  Configuration invalid
  4  Detection failed`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("vibe-dash starting")

		// Use package-level repository (injected via SetRepository in main.go)
		if repository == nil {
			slog.Error("Repository not initialized")
			return
		}

		// Pass detection service, waiting detector, and file watcher to TUI (Story 3.6, 4.5, 4.6)
		// Uses existing package variables from add.go
		if err := tui.Run(cmd.Context(), repository, detectionService, waitingDetector, fileWatcher); err != nil {
			slog.Error("TUI error", "error", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// It accepts a context for graceful shutdown support.
func Execute(ctx context.Context) error {
	return RootCmd.ExecuteContext(ctx)
}
