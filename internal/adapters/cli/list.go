package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/project"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/stageformat"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/timeformat"
)

// listJSON holds the --json flag value
var listJSON bool

// listAPIVersion holds the --api-version flag value
var listAPIVersion string

// ResetListFlags resets list command flags for testing.
// Call this before each test to ensure clean state.
func ResetListFlags() {
	listJSON = false
	listAPIVersion = "v1"
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
	cmd.Flags().StringVar(&listAPIVersion, "api-version", "v1", "API version for JSON output (currently only v1)")

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

	// Validate API version (AC3: only v1 supported currently)
	if listAPIVersion != "v1" {
		return fmt.Errorf("unsupported API version: %s", listAPIVersion)
	}

	projects, err := repository.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	// Sort alphabetically by effective name
	project.SortByName(projects)

	if listJSON {
		return formatJSON(ctx, cmd, projects)
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

		stage := stageformat.FormatStageInfo(p)
		lastActive := timeformat.FormatRelativeTime(p.LastActivityAt)

		fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10s %12s\n", name, stage, lastActive)
	}
}

// ListResponse represents the JSON output structure
type ListResponse struct {
	APIVersion    string           `json:"api_version"`
	Projects      []ProjectSummary `json:"projects"`
	ConfigWarning *string          `json:"config_warning,omitempty"` // Story 7.2: Config error message (null if no error)
}

// ProjectSummary represents a single project in JSON output
type ProjectSummary struct {
	Name                   string  `json:"name"`
	DisplayName            *string `json:"display_name"` // null if not set
	Path                   string  `json:"path"`
	Method                 string  `json:"method"`
	Stage                  string  `json:"stage"`      // lowercase per Architecture spec
	Confidence             string  `json:"confidence"` // lowercase: "certain", "likely", "uncertain"
	State                  string  `json:"state"`      // lowercase: "active" or "hibernated"
	IsFavorite             bool    `json:"is_favorite"`
	IsWaiting              bool    `json:"is_waiting"`               // Agent waiting detection status
	WaitingDurationMinutes *int    `json:"waiting_duration_minutes"` // Minutes waiting, null if not waiting
	Notes                  *string `json:"notes"`                    // User notes, null if not set
	DetectionReasoning     *string `json:"detection_reasoning"`      // Detection explanation, null if empty
	LastActivityAt         string  `json:"last_activity_at"`         // ISO 8601 UTC (RFC3339)
}

// formatJSON formats projects as JSON output.
func formatJSON(ctx context.Context, cmd *cobra.Command, projects []*domain.Project) error {
	// Story 7.2: Include config warning if present (AC7)
	var cfgWarning *string
	if configWarning != "" {
		cfgWarning = &configWarning
	}

	response := ListResponse{
		APIVersion:    listAPIVersion,
		Projects:      make([]ProjectSummary, 0, len(projects)),
		ConfigWarning: cfgWarning,
	}

	for _, p := range projects {
		// Nullable display_name (null if not set)
		var displayName *string
		if p.DisplayName != "" {
			displayName = &p.DisplayName
		}

		// Nullable notes (null if not set)
		var notes *string
		if p.Notes != "" {
			notes = &p.Notes
		}

		// Nullable detection_reasoning (null if not set)
		var detectionReasoning *string
		if p.DetectionReasoning != "" {
			detectionReasoning = &p.DetectionReasoning
		}

		// Waiting detection (AC4: is_waiting and waiting_duration_minutes)
		isWaiting := false
		var waitingMinutes *int
		if waitingDetector != nil {
			isWaiting = waitingDetector.IsWaiting(ctx, p)
			if isWaiting {
				mins := int(waitingDetector.WaitingDuration(ctx, p).Minutes())
				waitingMinutes = &mins
			}
		}

		response.Projects = append(response.Projects, ProjectSummary{
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
		})
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}
