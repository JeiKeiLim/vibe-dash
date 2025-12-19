package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// repository is the project repository injected at startup.
// Package-level variable for MVP simplicity (Option A from Dev Notes).
var repository ports.ProjectRepository

// detectionService is the detection service injected at startup.
// Used by add command to detect project methodology.
var detectionService ports.Detector

// waitingDetector is the waiting detector injected at startup.
// Story 4.5: Used by TUI for WAITING indicator display.
var waitingDetector ports.WaitingDetector

// SetRepository sets the project repository for the add command.
// Used by main.go for production and tests for mocking.
func SetRepository(repo ports.ProjectRepository) {
	repository = repo
}

// SetDetectionService sets the detection service for the add command.
// Used by main.go for production and tests for mocking.
func SetDetectionService(svc ports.Detector) {
	detectionService = svc
}

// SetWaitingDetector sets the waiting detector for the TUI.
// Story 4.5: Used by main.go to inject the WaitingDetector for WAITING indicators.
func SetWaitingDetector(detector ports.WaitingDetector) {
	waitingDetector = detector
}

// addName holds the --name flag value
var addName string

// addForce holds the --force flag value
var addForce bool

// ResetAddFlags resets add command flags for testing.
// Call this before each test to ensure clean state.
func ResetAddFlags() {
	addName = ""
	addForce = false
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
	cmd.Flags().BoolVar(&addForce, "force", false, "Auto-resolve name collisions without prompting")

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

	// Name collision handling (only if no custom name provided)
	nameToCheck := project.Name
	if project.DisplayName != "" {
		nameToCheck = project.DisplayName
	}

	nameCollision, err := checkNameCollision(ctx, repository, nameToCheck)
	if err != nil {
		return err
	}

	if nameCollision != nil {
		// Generate suggested name
		suggestedName, err := generateUniqueName(ctx, repository, project.Name, canonicalPath)
		if err != nil {
			return fmt.Errorf("failed to generate unique name: %w", err)
		}

		if addForce {
			// Auto-resolve without prompt
			project.DisplayName = suggestedName
		} else {
			// Interactive prompt
			resolvedName, err := promptCollisionResolution(ctx, cmd, nameToCheck, suggestedName, repository)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return fmt.Errorf("operation cancelled")
				}
				return err
			}
			project.DisplayName = resolvedName
		}
	}

	// Perform detection if service is available
	if detectionService != nil {
		result, err := detectionService.Detect(ctx, canonicalPath)
		if err == nil && result != nil {
			project.DetectedMethod = result.Method
			project.CurrentStage = result.Stage
		}
		// Detection failure is non-fatal - project defaults to unknown
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
	if project.DetectedMethod != "" && project.DetectedMethod != "unknown" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Method: %s (%s)\n", project.DetectedMethod, project.CurrentStage)
	}

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

// checkNameCollision checks if a project with the given name already exists.
// Checks both Name and DisplayName fields to prevent dashboard confusion.
func checkNameCollision(ctx context.Context, repo ports.ProjectRepository, name string) (*domain.Project, error) {
	projects, err := repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check name collision: %w", err)
	}

	for _, p := range projects {
		if p.Name == name || p.DisplayName == name {
			return p, nil // Collision found
		}
	}
	return nil, nil // No collision
}

// generateUniqueName creates a unique name by prepending parent directories.
// Example: /home/user/clients/client-b/api-service
//
//	Base: api-service
//	Try 1: client-b-api-service
//	Try 2: clients-client-b-api-service (if still collides)
func generateUniqueName(ctx context.Context, repo ports.ProjectRepository, baseName, fullPath string) (string, error) {
	parts := strings.Split(filepath.Dir(fullPath), string(filepath.Separator))

	// Filter empty parts
	var validParts []string
	for _, p := range parts {
		if p != "" && p != "." {
			validParts = append(validParts, p)
		}
	}

	candidate := baseName
	prefixIdx := len(validParts) - 1
	truncatedCandidate := "" // Track if we've truncated to detect repeats

	for {
		existing, err := checkNameCollision(ctx, repo, candidate)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return candidate, nil // Unique name found
		}

		if prefixIdx < 0 {
			// Ran out of path components - use timestamp suffix
			return fmt.Sprintf("%s-%d", candidate, time.Now().Unix()), nil
		}

		candidate = fmt.Sprintf("%s-%s", validParts[prefixIdx], candidate)
		prefixIdx--

		// Truncate if too long
		if len(candidate) > 50 {
			newTruncated := candidate[:50]
			// If truncation produces same result as before, add timestamp to break the cycle
			if newTruncated == truncatedCandidate {
				return fmt.Sprintf("%s-%d", newTruncated[:40], time.Now().Unix()), nil
			}
			truncatedCandidate = newTruncated
			candidate = newTruncated
		}
	}
}

// promptCollisionResolution prompts user to resolve a name collision.
// Uses cmd.InOrStdin() for testability - tests can inject mock stdin.
func promptCollisionResolution(ctx context.Context, cmd *cobra.Command, existingName, suggestedName string, repo ports.ProjectRepository) (string, error) {
	fmt.Fprintf(cmd.OutOrStdout(), "Project name '%s' already exists.\n", existingName)
	fmt.Fprintf(cmd.OutOrStdout(), "Choose an option:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  1. Use suggested name: %s\n", suggestedName)
	fmt.Fprintf(cmd.OutOrStdout(), "  2. Enter a custom name\n")

	scanner := bufio.NewScanner(cmd.InOrStdin())

	for {
		fmt.Fprintf(cmd.OutOrStdout(), "Enter choice (1/2): ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", io.EOF // User pressed Ctrl+D
		}

		choice := strings.TrimSpace(scanner.Text())
		switch choice {
		case "1":
			return suggestedName, nil
		case "2":
			return promptCustomName(ctx, cmd, scanner, repo)
		default:
			fmt.Fprintf(cmd.OutOrStdout(), "Invalid choice '%s'. Please enter 1 or 2.\n", choice)
			// Loop continues to re-prompt
		}
	}
}

// promptCustomName prompts for a custom name and validates it doesn't collide.
func promptCustomName(ctx context.Context, cmd *cobra.Command, scanner *bufio.Scanner, repo ports.ProjectRepository) (string, error) {
	for {
		fmt.Fprintf(cmd.OutOrStdout(), "Enter custom name: ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
		customName := strings.TrimSpace(scanner.Text())

		// Validate empty input
		if customName == "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Name cannot be empty.\n")
			continue
		}

		// Check for collision with custom name
		collision, err := checkNameCollision(ctx, repo, customName)
		if err != nil {
			return "", err
		}
		if collision != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Name '%s' also exists. Try another.\n", customName)
			continue
		}

		return customName, nil
	}
}
