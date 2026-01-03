package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/config"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// vibeHome is the base path for vibe-dash storage (~/.vibe-dash).
// Package-level variable for testability.
var vibeHome = config.GetDefaultBasePath()

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage vibe-dash configuration",
	Long: `Manage vibe-dash configuration settings.

Use 'config set' to modify per-project configuration values.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <project> <key> <value>",
	Short: "Set a project configuration value",
	Long: `Set a configuration value for a specific project.

Supported keys:
  hibernation-days     Days of inactivity before auto-hibernation (0 to disable)
  waiting-threshold    Agent waiting threshold in minutes (0 to disable)

Examples:
  vibe config set my-project hibernation-days 30
  vibe config set my-project waiting-threshold 5
  vibe config set api-service waiting-threshold 0    # Disable detection`,
	Args: cobra.ExactArgs(3),
	RunE: runConfigSet,
}

func init() {
	configCmd.AddCommand(configSetCmd)
	RootCmd.AddCommand(configCmd)
}

// runConfigSet implements the 'config set' command logic.
func runConfigSet(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	key := args[1]
	value := args[2]

	var err error
	switch key {
	case "hibernation-days":
		var intVal int
		intVal, err = strconv.Atoi(value)
		if err != nil {
			err = fmt.Errorf("%w: invalid value for hibernation-days: %s", domain.ErrConfigInvalid, value)
		} else if intVal < 0 {
			err = fmt.Errorf("%w: hibernation-days must be >= 0, got %d", domain.ErrConfigInvalid, intVal)
		} else {
			return setProjectHibernationDays(cmd.Context(), cmd, projectID, intVal)
		}
	case "waiting-threshold":
		var intVal int
		intVal, err = strconv.Atoi(value)
		if err != nil {
			err = fmt.Errorf("%w: invalid value for waiting-threshold: %s", domain.ErrConfigInvalid, value)
		} else if intVal < 0 {
			err = fmt.Errorf("%w: waiting-threshold must be >= 0, got %d", domain.ErrConfigInvalid, intVal)
		} else {
			return setProjectWaitingThreshold(cmd.Context(), cmd, projectID, intVal)
		}
	default:
		err = fmt.Errorf("%w: unknown config key: %s", domain.ErrConfigInvalid, key)
	}

	// SilenceErrors/SilenceUsage pattern for clean error output (matches status.go:164-165)
	if err != nil && errors.Is(err, domain.ErrConfigInvalid) {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}
	return err
}

// setProjectWaitingThreshold updates the waiting threshold for a project.
func setProjectWaitingThreshold(ctx context.Context, cmd *cobra.Command, projectID string, threshold int) error {
	// Get vibe home directory
	projectDir := filepath.Join(vibeHome, projectID)

	// Check if project directory exists before proceeding
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("project directory not found: %s (expected at %s)", projectID, projectDir)
	}

	// Load existing project config (creates default if doesn't exist)
	loader, err := config.NewProjectConfigLoader(projectDir)
	if err != nil {
		return fmt.Errorf("failed to access project config: %w", err)
	}

	data, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	// Update threshold
	data.AgentWaitingThresholdMinutes = &threshold

	// Save back
	if err := loader.Save(ctx, data); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Set waiting-threshold=%d for project %s\n", threshold, projectID)
	return nil
}

// setProjectHibernationDays updates the hibernation days for a project.
func setProjectHibernationDays(ctx context.Context, cmd *cobra.Command, projectID string, days int) error {
	// Get vibe home directory
	projectDir := filepath.Join(vibeHome, projectID)

	// Check if project directory exists before proceeding
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("project directory not found: %s (expected at %s)", projectID, projectDir)
	}

	// Load existing project config (creates default if doesn't exist)
	loader, err := config.NewProjectConfigLoader(projectDir)
	if err != nil {
		return fmt.Errorf("failed to access project config: %w", err)
	}

	data, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	// Update hibernation days
	data.CustomHibernationDays = &days

	// Save back
	if err := loader.Save(ctx, data); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Set hibernation-days=%d for project %s\n", days, projectID)
	return nil
}
