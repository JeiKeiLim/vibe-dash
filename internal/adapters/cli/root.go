package cli

import (
	"context"

	"github.com/spf13/cobra"
)

// RootCmd is the root command for the vibe CLI.
var RootCmd = &cobra.Command{
	Use:   "vibe",
	Short: "vibe-dash - AI coding project dashboard",
	Long:  `vibe-dash is a terminal dashboard for tracking AI-assisted coding projects.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// It accepts a context for graceful shutdown support.
func Execute(ctx context.Context) error {
	return RootCmd.ExecuteContext(ctx)
}
