package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// newHibernateCmd creates the hibernate command.
func newHibernateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hibernate <project-name>",
		Short: "Hibernate a project (mark as dormant)",
		Long: `Hibernate a project, moving it from active to dormant state.

Hibernated projects:
  - Don't appear in the main dashboard (press [h] to view)
  - Don't trigger agent waiting detection
  - Auto-activate when file changes are detected

Projects can be identified by name, display name, or path.
Favorite projects cannot be hibernated (remove favorite first).

Examples:
  vibe hibernate my-project             # By name
  vibe hibernate /home/user/my-project  # By path
  vibe hibernate "My Cool App"          # By display name`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: projectCompletionFunc,
		RunE:              runHibernate,
	}

	return cmd
}

// RegisterHibernateCommand registers the hibernate command with the given parent.
// Used for testing to create fresh command trees.
func RegisterHibernateCommand(parent *cobra.Command) {
	parent.AddCommand(newHibernateCmd())
}

func init() {
	RootCmd.AddCommand(newHibernateCmd())
}

func runHibernate(cmd *cobra.Command, args []string) error {
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

	// Attempt hibernation via StateService
	err = stateService.Hibernate(ctx, project.ID)
	if err != nil {
		// Handle specific errors
		if errors.Is(err, domain.ErrInvalidStateTransition) {
			// Already hibernated - idempotent success (AC3)
			if !IsQuiet() {
				fmt.Fprintf(cmd.OutOrStdout(), "Project is already hibernated: %s\n", identifier)
			}
			return nil
		}
		if errors.Is(err, domain.ErrFavoriteCannotHibernate) {
			// Favorite cannot hibernate (AC5)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			fmt.Fprintf(cmd.OutOrStdout(), "Cannot hibernate favorite project: %s\n", identifier)
			fmt.Fprintf(cmd.OutOrStdout(), "Remove favorite status first with: vibe favorite %s --off\n", identifier)
			return err
		}
		return fmt.Errorf("failed to hibernate project: %w", err)
	}

	// Success output
	if !IsQuiet() {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Hibernated: %s\n", identifier)
	}

	return nil
}
