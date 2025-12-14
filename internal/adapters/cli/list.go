package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/project"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)

// listJSON holds the --json flag value
var listJSON bool

// ResetListFlags resets list command flags for testing.
// Call this before each test to ensure clean state.
func ResetListFlags() {
	listJSON = false
}

// newListCmd creates the list command.
func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tracked projects",
		Long: `List all projects tracked by vibe-dash.

Shows project name, workflow stage, and time since last activity.
Use --json for machine-readable output.

Examples:
  vibe list           # Plain text output
  vibe list --json    # JSON output for scripting`,
		Args: cobra.NoArgs,
		RunE: runList,
	}

	cmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")

	return cmd
}

// RegisterListCommand registers the list command with the given parent command.
// Used for testing to create fresh command trees.
func RegisterListCommand(parent *cobra.Command) {
	parent.AddCommand(newListCmd())
}

func init() {
	RootCmd.AddCommand(newListCmd())
}

// runList implements the list command logic.
func runList(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	if repository == nil {
		return fmt.Errorf("repository not initialized")
	}

	projects, err := repository.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	// Sort alphabetically by effective name
	project.SortByName(projects)

	if listJSON {
		return formatJSON(cmd, projects)
	}

	// Plain text output
	if len(projects) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No projects tracked. Run 'vibe add .' to add one.\n")
		return nil
	}

	formatPlainText(cmd, projects)
	return nil
}

// formatPlainText formats projects as a plain text table.
func formatPlainText(cmd *cobra.Command, projects []*domain.Project) {
	// Header
	fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10s %12s\n", "PROJECT", "STAGE", "LAST ACTIVE")

	for _, p := range projects {
		name := project.EffectiveName(p)
		if len(name) > 40 {
			name = name[:37] + "..."
		}

		stage := p.CurrentStage.String()
		lastActive := timeformat.FormatRelativeTime(p.LastActivityAt)

		fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10s %12s\n", name, stage, lastActive)
	}
}

// ListResponse represents the JSON output structure
type ListResponse struct {
	APIVersion string           `json:"api_version"`
	Projects   []ProjectSummary `json:"projects"`
}

// ProjectSummary represents a single project in JSON output
type ProjectSummary struct {
	Name           string  `json:"name"`
	DisplayName    *string `json:"display_name"` // null if not set
	Path           string  `json:"path"`
	Method         string  `json:"method"`
	Stage          string  `json:"stage"`      // lowercase per Architecture spec
	Confidence     string  `json:"confidence"` // Default "uncertain" until DetectionResult stored
	State          string  `json:"state"`      // lowercase: "active" or "hibernated"
	IsFavorite     bool    `json:"is_favorite"`
	LastActivityAt string  `json:"last_activity_at"` // ISO 8601 UTC (RFC3339)
}

// formatJSON formats projects as JSON output.
func formatJSON(cmd *cobra.Command, projects []*domain.Project) error {
	response := ListResponse{
		APIVersion: "v1",
		Projects:   make([]ProjectSummary, 0, len(projects)),
	}

	for _, p := range projects {
		var displayName *string
		if p.DisplayName != "" {
			displayName = &p.DisplayName
		}

		response.Projects = append(response.Projects, ProjectSummary{
			Name:           p.Name,
			DisplayName:    displayName,
			Path:           p.Path,
			Method:         p.DetectedMethod,
			Stage:          strings.ToLower(p.CurrentStage.String()),
			Confidence:     "uncertain", // Default until DetectionResult is stored
			State:          strings.ToLower(p.State.String()),
			IsFavorite:     p.IsFavorite,
			LastActivityAt: p.LastActivityAt.UTC().Format(time.RFC3339),
		})
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}
