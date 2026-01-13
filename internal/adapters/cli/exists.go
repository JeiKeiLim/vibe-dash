package cli

import (
	"github.com/spf13/cobra"
)

// newExistsCmd creates the exists command.
// This command is completely silent - communicates ONLY through exit codes.
func newExistsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exists <project-name>",
		Short: "Check if a project is tracked (silent, exit code only)",
		Long: `Check if a project is tracked by name, display name, or path.

This command produces NO output - it communicates only through exit codes:
  Exit 0 = project exists
  Exit 2 = project not found

Examples:
  vdash exists client-alpha              # By name
  vdash exists "My Cool App"             # By display name
  vdash exists /home/user/project        # By path

Scripting usage:
  if vdash exists my-project; then
    echo "Already tracked"
  else
    vdash add .
  fi`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: projectCompletionFunc,
		RunE:              runExists,
	}

	return cmd
}

// RegisterExistsCommand registers the exists command with the given parent command.
// Used for testing to create fresh command trees.
func RegisterExistsCommand(parent *cobra.Command) {
	parent.AddCommand(newExistsCmd())
}

func init() {
	RootCmd.AddCommand(newExistsCmd())
}

// runExists implements the exists command logic.
// Silent command - NO output, only exit codes.
func runExists(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	identifier := args[0]

	// findProjectByIdentifier already wraps with domain.ErrProjectNotFound
	// DO NOT wrap again - just return the error directly
	_, err := findProjectByIdentifier(ctx, identifier)
	if err != nil {
		// Silence all output - this command communicates only via exit codes
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		// Wrap with SilentError to prevent main.go from logging
		return &SilentError{Err: err}
	}

	// Success: return nil (exit 0, no output)
	return nil
}
