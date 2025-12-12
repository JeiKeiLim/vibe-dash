package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
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

		// User-facing output goes to stdout (not slog)
		// slog output goes to stderr for diagnostics
		fmt.Fprintln(cmd.OutOrStdout(), "TUI dashboard coming soon. Press Ctrl+C to exit.")

		// Wait for context cancellation (respects signal handling in main.go)
		<-cmd.Context().Done()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// It accepts a context for graceful shutdown support.
func Execute(ctx context.Context) error {
	return RootCmd.ExecuteContext(ctx)
}
