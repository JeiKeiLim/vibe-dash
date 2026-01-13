package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// newActivateCmd creates the activate command.
func newActivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <project-name>",
		Short: "Activate a hibernated project",
		Long: `Activate a hibernated project, moving it back to active state.

Activated projects:
  - Appear in the main dashboard
  - Resume agent waiting detection
  - May auto-hibernate again after hibernation threshold days

Projects can be identified by name, display name, or path.

Examples:
  vdash activate my-project             # By name
  vdash activate /home/user/my-project  # By path
  vdash activate "My Cool App"          # By display name`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: projectCompletionFunc,
		RunE:              runActivate,
	}

	return cmd
}

// RegisterActivateCommand registers the activate command with the given parent.
// Used for testing to create fresh command trees.
func RegisterActivateCommand(parent *cobra.Command) {
	parent.AddCommand(newActivateCmd())
}

func init() {
	RootCmd.AddCommand(newActivateCmd())
}

func runActivate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Validate stateService is available
	if stateService == nil {
		return fmt.Errorf("state service not initialized")
	}

	identifier := args[0]

	// Find project using existing helper (from status.go)
	project, err := findProjectByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			fmt.Fprintf(cmd.OutOrStdout(), "✗ Project not found: %s\n", identifier)
		}
		return err
	}

	// Attempt activation via StateService
	err = stateService.Activate(ctx, project.ID)
	if err != nil {
		// Handle specific errors
		if errors.Is(err, domain.ErrInvalidStateTransition) {
			// Already active - idempotent success (AC4)
			if !IsQuiet() {
				fmt.Fprintf(cmd.OutOrStdout(), "Project is already active: %s\n", identifier)
			}
			return nil
		}
		return fmt.Errorf("failed to activate project: %w", err)
	}

	// Success output
	if !IsQuiet() {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Activated: %s\n", identifier)
	}

	return nil
}
