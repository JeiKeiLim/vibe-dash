package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	resetAll     bool
	resetConfirm bool
)

func newResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset [project]",
		Short: "Reset project database to clean state",
		Long: `Delete and recreate a project's state.db. Config preserved.

The reset command deletes the project's SQLite database and allows it to be
recreated fresh on next access. This is useful for recovering from database
corruption or clearing cached detection state.

Examples:
  vibe reset my-project --confirm     # Reset single project
  vibe reset --all --confirm          # Reset all project databases`,
		Args: cobra.MaximumNArgs(1),
		RunE: runReset,
	}
	cmd.Flags().BoolVar(&resetAll, "all", false, "Reset all project databases")
	cmd.Flags().BoolVar(&resetConfirm, "confirm", false, "Confirm reset operation")
	return cmd
}

func init() {
	RootCmd.AddCommand(newResetCmd())
}

// RegisterResetCommand registers the reset command with the given parent command.
// Used for testing to create fresh command trees.
func RegisterResetCommand(parent *cobra.Command) {
	parent.AddCommand(newResetCmd())
}

// ResetResetFlags resets the reset command flags for testing.
func ResetResetFlags() {
	resetAll = false
	resetConfirm = false
}

func runReset(cmd *cobra.Command, args []string) error {
	if !resetConfirm {
		fmt.Fprintln(cmd.OutOrStdout(), "⚠ This deletes and recreates project database(s).")
		fmt.Fprintln(cmd.OutOrStdout(), "  Config.yaml is preserved. Use --confirm to proceed.")
		return nil
	}
	ctx := cmd.Context()

	// Check repository is initialized
	if repository == nil {
		return fmt.Errorf("repository not initialized")
	}

	if resetAll {
		count, err := repository.ResetAll(ctx)
		if err != nil {
			return fmt.Errorf("reset failed: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Reset %d projects\n", count)
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("specify project name/path or use --all")
	}

	projectID := args[0] // Could be name or path - resolve in coordinator
	if err := repository.ResetProject(ctx, projectID); err != nil {
		return fmt.Errorf("✗ Reset failed for %s: %w", projectID, err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "✓ Reset: %s\n", projectID)
	return nil
}
