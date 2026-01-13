package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// removeForce holds the --force flag value
var removeForce bool

// ResetRemoveFlags resets remove command flags for testing.
// Call this before each test to ensure clean state.
func ResetRemoveFlags() {
	removeForce = false
}

// newRemoveCmd creates the remove command.
func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <project-name>",
		Short: "Remove a project from tracking",
		Long: `Remove a project from the vibe-dash dashboard.

The project is identified by its name or display name.
By default, confirmation is required before removal.
Use --force to skip confirmation.

Examples:
  vdash remove client-alpha          # Remove with confirmation
  vdash remove client-alpha --force  # Remove immediately
  vdash remove "My Project"          # Remove by display name`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: projectCompletionFunc,
		RunE:              runRemove,
	}

	cmd.Flags().BoolVar(&removeForce, "force", false, "Remove without confirmation")

	return cmd
}

// RegisterRemoveCommand registers the remove command with the given parent command.
// Used for testing to create fresh command trees.
func RegisterRemoveCommand(parent *cobra.Command) {
	parent.AddCommand(newRemoveCmd())
}

func init() {
	RootCmd.AddCommand(newRemoveCmd())
}

// findProjectByName finds a project by name or display_name.
// Searches both fields to support AC5 (remove by display name).
//
// NOTE: Similar to checkNameCollision() in add.go (lines 209-223) which also
// uses FindAll() + in-memory filtering. They serve different purposes:
// - findProjectByName: retrieves a project for removal
// - checkNameCollision: detects if a name is already taken
// Consider extracting a shared helper if this pattern appears again.
func findProjectByName(ctx context.Context, repo ports.ProjectRepository, name string) (*domain.Project, error) {
	projects, err := repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	for _, p := range projects {
		if p.Name == name || p.DisplayName == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", domain.ErrProjectNotFound, name)
}

// runRemove implements the remove command logic.
func runRemove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if repository == nil {
		return fmt.Errorf("repository not initialized")
	}

	projectName := args[0]

	// Find project by name or display_name
	project, err := findProjectByName(ctx, repository, projectName)
	if err != nil {
		// AC6: Format error message with ✗ prefix and silence Cobra's default error output
		if errors.Is(err, domain.ErrProjectNotFound) {
			fmt.Fprintf(cmd.OutOrStdout(), "✗ Project not found: %s\n", projectName)
			cmd.SilenceErrors = true
		}
		return err
	}

	// Display name for output - same pattern as list.go effectiveName()
	displayName := project.Name
	if project.DisplayName != "" {
		displayName = project.DisplayName
	}

	// Confirmation unless --force
	if !removeForce {
		confirmed, err := promptRemovalConfirmation(cmd, displayName)
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintf(cmd.OutOrStdout(), "Cancelled\n")
				return nil
			}
			return err
		}
		if !confirmed {
			fmt.Fprintf(cmd.OutOrStdout(), "Cancelled\n")
			return nil
		}
	}

	// Store path before delete (needed for directory cleanup)
	projectPath := project.Path

	// Delete from repository (removes from DB and config)
	if err := repository.Delete(ctx, project.ID); err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}

	// Delete project directory using DirectoryManager (non-fatal if fails)
	if directoryManager != nil {
		if err := directoryManager.DeleteProjectDir(ctx, projectPath); err != nil {
			slog.Warn("failed to delete project directory", "path", projectPath, "error", err)
			// Continue - project removed from tracking, directory left behind
		}
	}

	if !IsQuiet() {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed: %s\n", displayName)
	}
	return nil
}

// promptRemovalConfirmation prompts user to confirm project removal.
// Uses cmd.InOrStdin() for testability - tests can inject mock stdin.
// Returns true for confirmation (y/yes), false for cancellation (n/no).
func promptRemovalConfirmation(cmd *cobra.Command, projectName string) (bool, error) {
	scanner := bufio.NewScanner(cmd.InOrStdin())

	for {
		fmt.Fprintf(cmd.OutOrStdout(), "Remove '%s' from tracking? [y/n] ", projectName)

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return false, err
			}
			return false, io.EOF // User pressed Ctrl+D
		}

		input := strings.TrimSpace(strings.ToLower(scanner.Text()))
		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintf(cmd.OutOrStdout(), "Please enter 'y' or 'n'.\n")
			// Loop continues to re-prompt
		}
	}
}
