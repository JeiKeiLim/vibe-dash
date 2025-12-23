package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/project"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)

// Package-level flags (same pattern as list.go)
var statusJSON bool
var statusAPIVersion string
var statusAll bool

// ResetStatusFlags resets status command flags for testing.
// Call this before each test to ensure clean state.
func ResetStatusFlags() {
	statusJSON = false
	statusAPIVersion = "v1"
	statusAll = false
}

// StatusResponse represents the JSON output structure for single project.
// Different from ListResponse which uses "projects" array.
type StatusResponse struct {
	APIVersion string         `json:"api_version"` // Schema version (currently "v1")
	Project    ProjectSummary `json:"project"`     // Single project object (NOT array)
}

// newStatusCmd creates the status command.
func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [project-name]",
		Short: "Show status of a tracked project",
		Long: `Show detailed status of a specific project or all projects.

Use a project name, display name, or path to identify the project.
Use --all to show all projects (same as 'vibe list').

Examples:
  vibe status client-alpha          # By name
  vibe status "My Cool App"         # By display name
  vibe status /home/user/project    # By path
  vibe status client-alpha --json   # JSON output
  vibe status --all                 # All projects`,
		Args: cobra.MaximumNArgs(1),
		RunE: runStatus,
	}

	cmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&statusAPIVersion, "api-version", "v1", "API version for JSON output")
	cmd.Flags().BoolVar(&statusAll, "all", false, "Show all projects")

	return cmd
}

// RegisterStatusCommand registers the status command with the given parent command.
// Used for testing to create fresh command trees.
func RegisterStatusCommand(parent *cobra.Command) {
	parent.AddCommand(newStatusCmd())
}

func init() {
	RootCmd.AddCommand(newStatusCmd())
}

// findProjectByIdentifier finds a project by name, display_name, or path.
// Extends findProjectByName pattern from remove.go.
// Lookup order: Name → DisplayName → Path (canonicalized)
func findProjectByIdentifier(ctx context.Context, identifier string) (*domain.Project, error) {
	if repository == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	projects, err := repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// 1. Try exact name match (highest priority)
	for _, p := range projects {
		if p.Name == identifier {
			return p, nil
		}
	}

	// 2. Try display name match
	for _, p := range projects {
		if p.DisplayName == identifier {
			return p, nil
		}
	}

	// 3. Try path match (canonicalize input, ignore errors for non-paths)
	canonicalInput, err := filesystem.CanonicalPath(identifier)
	if err == nil { // Only try path match if input resolves to a valid path
		for _, p := range projects {
			if p.Path == canonicalInput {
				return p, nil
			}
		}
	}

	return nil, fmt.Errorf("%w: %s", domain.ErrProjectNotFound, identifier)
}

// runStatus implements the status command logic.
func runStatus(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// 1. Repository nil check (same as list.go:67-69)
	if repository == nil {
		return fmt.Errorf("repository not initialized")
	}

	// 2. API version validation (same as list.go:72-74)
	if statusAPIVersion != "v1" {
		return fmt.Errorf("unsupported API version: %s", statusAPIVersion)
	}

	// 3. Handle --all mode (delegate to list functions)
	if statusAll {
		projects, err := repository.FindAll(ctx)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}
		project.SortByName(projects)

		if statusJSON {
			return formatJSON(ctx, cmd, projects) // From list.go
		}

		// Plain text output
		if len(projects) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No projects tracked. Run 'vibe add .' to add one.\n")
			return nil
		}

		formatPlainText(cmd, projects) // From list.go
		return nil
	}

	// 4. Require project identifier if not --all (AC8)
	if len(args) == 0 {
		return fmt.Errorf("requires a project name or --all flag")
	}

	// 5. Find project by identifier
	identifier := args[0]
	proj, err := findProjectByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			// SilenceErrors/SilenceUsage pattern for clean error output
			fmt.Fprintf(cmd.OutOrStdout(), "✗ Project not found: %s\n", identifier)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
		}
		return err
	}

	// 6. Output single project
	if statusJSON {
		return formatStatusJSON(ctx, cmd, proj)
	}
	formatStatusPlainText(cmd, proj)
	return nil
}

// formatStatusPlainText formats a single project as indented key-value pairs.
func formatStatusPlainText(cmd *cobra.Command, p *domain.Project) {
	// First line: effective name (DisplayName if set, else Name)
	name := p.Name
	if p.DisplayName != "" {
		name = p.DisplayName
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", name)

	// Indented details
	fmt.Fprintf(cmd.OutOrStdout(), "  Path:        %s\n", p.Path)
	// Title case the method (e.g., "bmad" → "Bmad")
	method := p.DetectedMethod
	if method != "" {
		method = strings.ToUpper(method[:1]) + method[1:]
	}
	fmt.Fprintf(cmd.OutOrStdout(), "  Method:      %s\n", method)
	fmt.Fprintf(cmd.OutOrStdout(), "  Stage:       %s\n", p.CurrentStage.String())
	fmt.Fprintf(cmd.OutOrStdout(), "  Confidence:  %s\n", p.Confidence.String())
	fmt.Fprintf(cmd.OutOrStdout(), "  State:       %s\n", p.State.String())

	favorite := "No"
	if p.IsFavorite {
		favorite = "Yes"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "  Favorite:    %s\n", favorite)

	if p.Notes != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Notes:       %s\n", p.Notes)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "  Last Active: %s\n", timeformat.FormatRelativeTime(p.LastActivityAt))
}

// formatStatusJSON formats a single project as JSON output.
func formatStatusJSON(ctx context.Context, cmd *cobra.Command, p *domain.Project) error {
	// Build ProjectSummary using same nullable patterns as list.go:147-189
	var displayName *string
	if p.DisplayName != "" {
		displayName = &p.DisplayName
	}

	var notes *string
	if p.Notes != "" {
		notes = &p.Notes
	}

	var detectionReasoning *string
	if p.DetectionReasoning != "" {
		detectionReasoning = &p.DetectionReasoning
	}

	isWaiting := false
	var waitingMinutes *int
	if waitingDetector != nil {
		isWaiting = waitingDetector.IsWaiting(ctx, p)
		if isWaiting {
			mins := int(waitingDetector.WaitingDuration(ctx, p).Minutes())
			waitingMinutes = &mins
		}
	}

	response := StatusResponse{
		APIVersion: statusAPIVersion,
		Project: ProjectSummary{
			Name:                   p.Name,
			DisplayName:            displayName,
			Path:                   p.Path,
			Method:                 p.DetectedMethod,
			Stage:                  strings.ToLower(p.CurrentStage.String()),
			Confidence:             strings.ToLower(p.Confidence.String()),
			State:                  strings.ToLower(p.State.String()),
			IsFavorite:             p.IsFavorite,
			IsWaiting:              isWaiting,
			WaitingDurationMinutes: waitingMinutes,
			Notes:                  notes,
			DetectionReasoning:     detectionReasoning,
			LastActivityAt:         p.LastActivityAt.UTC().Format(time.RFC3339),
		},
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

