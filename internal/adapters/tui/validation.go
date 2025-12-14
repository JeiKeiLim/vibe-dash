package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// viewMode determines which view to render
type viewMode int

const (
	viewModeNormal viewMode = iota
	viewModeValidation
)

// InvalidProject represents a project with an inaccessible path
type InvalidProject struct {
	Project *domain.Project
	Error   error
}

// ValidateProjectPaths checks all projects for inaccessible paths.
// Returns a slice of InvalidProject for projects where filesystem.ResolvePath returns error.
// Returns empty slice if all paths are valid (AC5).
func ValidateProjectPaths(ctx context.Context, repo ports.ProjectRepository) ([]InvalidProject, error) {
	projects, err := repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// CRITICAL: Return empty slice, not nil
	invalid := make([]InvalidProject, 0)

	for _, p := range projects {
		_, err := filesystem.ResolvePath(p.Path)
		if err != nil {
			invalid = append(invalid, InvalidProject{
				Project: p,
				Error:   err,
			})
		}
	}

	return invalid, nil
}

// renderValidationDialog renders the path validation dialog for a single project.
// Layout per AC1: project name, path, three options [D/M/K]
// If errorMsg is non-empty, displays error feedback to user.
func renderValidationDialog(project *domain.Project, width, height int, errorMsg string) string {
	title := WarningStyle.Render("Warning: Project path not found: " + effectiveName(project))

	lines := []string{
		"",
		title,
		"",
		DimStyle.Render(project.Path),
		"",
	}

	// Show error message if present (H1 fix: user feedback on failures)
	if errorMsg != "" {
		lines = append(lines, WaitingStyle.Render("Error: "+errorMsg))
		lines = append(lines, "")
	}

	lines = append(lines,
		"[D] Delete - Remove from dashboard",
		"[M] Move - Update to current directory",
		"[K] Keep - Maybe network mount, keep tracking",
		"",
	)

	content := strings.Join(lines, "\n")

	box := boxStyle.Width(60).Render(content)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// effectiveName returns DisplayName if set, otherwise Name
func effectiveName(p *domain.Project) string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.Name
}
