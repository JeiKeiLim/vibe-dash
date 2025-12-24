package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// renameClear holds the --clear flag value
var renameClear bool

// ResetRenameFlags resets rename command flags for testing.
func ResetRenameFlags() {
	renameClear = false
}

// newRenameCmd creates the rename command.
func newRenameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <project-name> [new-display-name]",
		Short: "Set or clear a project's display name",
		Long: `Set or clear a project's display name.

By default, sets the display name to the provided value.
Use --clear to remove the display name.
Passing an empty string "" also clears the display name.

Examples:
  vibe rename api-service "Client A API"  # Set display name
  vibe rename api-service --clear          # Clear display name
  vibe rename api-service ""               # Clear display name (alternative)`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: projectCompletionFunc,
		RunE:              runRename,
	}

	cmd.Flags().BoolVar(&renameClear, "clear", false, "Clear display name")

	return cmd
}

// RegisterRenameCommand registers the rename command with the given parent.
// Used for testing to create fresh command trees.
func RegisterRenameCommand(parent *cobra.Command) {
	parent.AddCommand(newRenameCmd())
}

func init() {
	RootCmd.AddCommand(newRenameCmd())
}

func runRename(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Repository nil check
	if repository == nil {
		return fmt.Errorf("repository not initialized")
	}

	projectIdentifier := args[0]

	// Determine operation mode
	var newDisplayName string
	clearMode := false

	if renameClear {
		clearMode = true
	} else if len(args) == 2 {
		newDisplayName = args[1]
		if newDisplayName == "" {
			clearMode = true // Empty string = clear
		}
	} else {
		// No new name and no --clear flag
		return fmt.Errorf("requires a new name or --clear flag")
	}

	// Find project by identifier (reuses findProjectByIdentifier from status.go)
	proj, err := findProjectByIdentifier(ctx, projectIdentifier)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
		}
		return err
	}

	// Handle clear mode
	if clearMode {
		if proj.DisplayName == "" {
			// Already empty - idempotent success (AC8)
			if !IsQuiet() {
				fmt.Fprintf(cmd.OutOrStdout(), "☆ %s has no display name\n", proj.Name)
			}
			return nil
		}
		proj.DisplayName = ""
		proj.UpdatedAt = time.Now()

		if err := repository.Save(ctx, proj); err != nil {
			return fmt.Errorf("failed to save display name: %w", err)
		}

		if !IsQuiet() {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Cleared display name: %s\n", proj.Name)
		}
		return nil
	}

	// Set new display name
	proj.DisplayName = newDisplayName
	proj.UpdatedAt = time.Now()

	if err := repository.Save(ctx, proj); err != nil {
		return fmt.Errorf("failed to save display name: %w", err)
	}

	if !IsQuiet() {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Renamed: %s → %s\n", proj.Name, newDisplayName)
	}
	return nil
}
