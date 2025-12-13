package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// repository is the project repository injected at startup.
// Package-level variable for MVP simplicity (Option A from Dev Notes).
var repository ports.ProjectRepository

// SetRepository sets the project repository for the add command.
// Used by main.go for production and tests for mocking.
func SetRepository(repo ports.ProjectRepository) {
	repository = repo
}

// addName holds the --name flag value
var addName string

// ResetAddFlags resets add command flags for testing.
// Call this before each test to ensure clean state.
func ResetAddFlags() {
	addName = ""
}

// newAddCmd creates the add command.
func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [path]",
		Short: "Add a project to tracking",
		Long: `Add a project to the vibe-dash dashboard.

If no path is provided, the current directory is used.
The path is resolved to its canonical form (following symlinks).

Examples:
  vibe add .                  # Add current directory
  vibe add /path/to/project   # Add specific path
  vibe add . --name "My App"  # Add with custom display name`,
		Args: cobra.MaximumNArgs(1),
		RunE: runAdd,
	}

	cmd.Flags().StringVar(&addName, "name", "", "Custom display name for the project")

	return cmd
}

// RegisterAddCommand registers the add command with the given parent command.
// Used for testing to create fresh command trees.
func RegisterAddCommand(parent *cobra.Command) {
	parent.AddCommand(newAddCmd())
}

func init() {
	RootCmd.AddCommand(newAddCmd())
}

// runAdd implements the add command logic.
func runAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Check repository is initialized
	if repository == nil {
		return fmt.Errorf("repository not initialized")
	}

	// Get path (default to ".")
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Resolve to canonical path (handles ~, symlinks, existence check)
	canonicalPath, err := filesystem.CanonicalPath(path)
	if err != nil {
		// ErrPathNotAccessible will be mapped to exit code 1 by MapErrorToExitCode
		// Error already contains path info, no need to wrap again
		return err
	}

	// Check if already tracked (collision detection)
	existing, err := repository.FindByPath(ctx, canonicalPath)
	if err == nil {
		// Project exists - return domain error for proper exit code mapping
		displayName := existing.Name
		if existing.DisplayName != "" {
			displayName = existing.DisplayName
		}
		return fmt.Errorf("%w: %s", domain.ErrProjectAlreadyExists, displayName)
	}
	if !errors.Is(err, domain.ErrProjectNotFound) {
		// Unexpected error
		return fmt.Errorf("failed to check existing project: %w", err)
	}
	// domain.ErrProjectNotFound is expected - continue

	// Create new project
	project, err := domain.NewProject(canonicalPath, "")
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Set custom display name if provided
	if addName != "" {
		project.DisplayName = addName
	}

	// Save to repository
	if err := repository.Save(ctx, project); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	// Success output
	displayName := project.Name
	if project.DisplayName != "" {
		displayName = project.DisplayName
	}
	fmt.Fprintf(cmd.OutOrStdout(), "✓ Added: %s\n", displayName)
	fmt.Fprintf(cmd.OutOrStdout(), "  Path: %s\n", canonicalPath)

	return nil
}

// NewRootCmd creates a fresh root command for testing.
// This avoids state contamination between tests.
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vibe",
		Short: "CLI dashboard for vibe coding projects",
	}
}
