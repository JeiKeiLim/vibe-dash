package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// favoriteOff holds the --off flag value
var favoriteOff bool

// ResetFavoriteFlags resets favorite command flags for testing.
func ResetFavoriteFlags() {
	favoriteOff = false
}

// newFavoriteCmd creates the favorite command.
func newFavoriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "favorite <project-name>",
		Short: "Toggle or remove favorite status for a project",
		Long: `Toggle or remove favorite status for a tracked project.

By default, toggles the favorite status (on→off or off→on).
Use --off to explicitly remove favorite status.

Favorited projects:
  - Display with ⭐ prefix in dashboard
  - Never auto-hibernate (always visible)

Examples:
  vibe favorite my-project       # Toggle favorite status
  vibe favorite my-project --off # Remove favorite status`,
		Args: cobra.ExactArgs(1),
		RunE: runFavorite,
	}

	cmd.Flags().BoolVar(&favoriteOff, "off", false, "Remove favorite status (instead of toggle)")

	return cmd
}

// RegisterFavoriteCommand registers the favorite command with the given parent.
// Used for testing to create fresh command trees.
func RegisterFavoriteCommand(parent *cobra.Command) {
	parent.AddCommand(newFavoriteCmd())
}

func init() {
	RootCmd.AddCommand(newFavoriteCmd())
}

func runFavorite(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("%w: %s", domain.ErrProjectNotFound, projectName)
	}

	// Determine new favorite status
	var newFavorite bool
	if favoriteOff {
		// Explicit --off: remove favorite
		if !targetProject.IsFavorite {
			// Already not favorited - idempotent success (AC5)
			fmt.Fprintf(cmd.OutOrStdout(), "☆ %s is not favorited\n", projectName)
			return nil
		}
		newFavorite = false
	} else {
		// Toggle mode
		newFavorite = !targetProject.IsFavorite
	}

	// Update and save
	targetProject.IsFavorite = newFavorite
	targetProject.UpdatedAt = time.Now()

	if err := repository.Save(ctx, targetProject); err != nil {
		return fmt.Errorf("failed to save favorite status: %w", err)
	}

	// Success output
	if newFavorite {
		fmt.Fprintf(cmd.OutOrStdout(), "⭐ Favorited: %s\n", projectName)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "☆ Unfavorited: %s\n", projectName)
	}

	return nil
}
