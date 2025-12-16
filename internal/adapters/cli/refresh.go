package cli

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
)

// newRefreshCmd creates the refresh command.
func newRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Refresh detection for all tracked projects",
		Long: `Re-scan all tracked projects to update their methodology stage.

This command runs the detection service against all projects and updates
their stage based on current artifacts.`,
		RunE: runRefresh,
	}
}

// RegisterRefreshCommand registers the refresh command with the given parent.
// Used for testing to create fresh command trees.
func RegisterRefreshCommand(parent *cobra.Command) {
	parent.AddCommand(newRefreshCmd())
}

func init() {
	RootCmd.AddCommand(newRefreshCmd())
}

func runRefresh(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Use package-level dependencies (injected via SetRepository/SetDetectionService in main.go)
	// These are defined in add.go:22-26
	if repository == nil {
		return fmt.Errorf("repository not initialized")
	}

	if detectionService == nil {
		return fmt.Errorf("detection service not initialized")
	}

	// Get all projects
	projects, err := repository.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No projects to refresh.")
		return nil
	}

	var refreshedCount, failedCount int
	for _, project := range projects {
		result, err := detectionService.Detect(ctx, project.Path)
		if err != nil {
			slog.Debug("detection failed", "project", project.Name, "error", err)
			failedCount++
			continue
		}

		project.DetectedMethod = result.Method
		project.CurrentStage = result.Stage
		project.Confidence = result.Confidence
		project.DetectionReasoning = result.Reasoning
		project.UpdatedAt = time.Now()

		if err := repository.Save(ctx, project); err != nil {
			slog.Debug("save failed", "project", project.Name, "error", err)
			failedCount++
			continue
		}

		refreshedCount++
	}

	// AC3: Only return error if ALL projects fail
	if refreshedCount == 0 && failedCount > 0 {
		return fmt.Errorf("all %d projects failed to refresh", failedCount)
	}

	// Success output
	fmt.Fprintf(cmd.OutOrStdout(), "Refreshed %d projects", refreshedCount)
	if failedCount > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), " (%d failed)", failedCount)
	}
	fmt.Fprintln(cmd.OutOrStdout())

	return nil
}
