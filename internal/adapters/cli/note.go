package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// newNoteCmd creates the note command.
func newNoteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "note <project-name> [note-text]",
		Short: "View or set notes for a project",
		Long: `View or set notes for a tracked project.

If note-text is provided, sets the project note.
If no note-text is provided, displays the current note.

Examples:
  vibe note my-project "Waiting on API specs"   # Set note
  vibe note my-project ""                        # Clear note
  vibe note my-project                           # View current note`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: projectCompletionFunc,
		RunE:              runNote,
	}
}

// RegisterNoteCommand registers the note command with the given parent.
// Used for testing to create fresh command trees.
func RegisterNoteCommand(parent *cobra.Command) {
	parent.AddCommand(newNoteCmd())
}

func init() {
	RootCmd.AddCommand(newNoteCmd())
}

func runNote(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Use package-level repository (injected via SetRepository in main.go)
	if repository == nil {
		return fmt.Errorf("repository not initialized")
	}

	projectName := args[0]

	// Find project by name or display name
	projects, err := repository.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	var targetProject *domain.Project
	for _, p := range projects {
		if p.Name == projectName || p.DisplayName == projectName {
			targetProject = p
			break
		}
	}

	if targetProject == nil {
		err := fmt.Errorf("%w: %s", domain.ErrProjectNotFound, projectName)
		if errors.Is(err, domain.ErrProjectNotFound) {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
		}
		return err
	}

	// View mode (no note argument)
	if len(args) == 1 {
		if targetProject.Notes == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "(no note set)")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), targetProject.Notes)
		}
		return nil
	}

	// Set mode (note argument provided)
	newNote := args[1]
	targetProject.Notes = newNote
	targetProject.UpdatedAt = time.Now()

	if err := repository.Save(ctx, targetProject); err != nil {
		return fmt.Errorf("failed to save note: %w", err)
	}

	if !IsQuiet() {
		if newNote == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "✓ Note cleared")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "✓ Note saved")
		}
	}

	return nil
}
