// Package components provides TUI components for the dashboard.
package components

import (
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/project"
)

// ProjectItem wraps a domain.Project to implement the list.Item interface.
type ProjectItem struct {
	Project *domain.Project
}

// FilterValue returns the value used for filtering.
// Returns the effective name (display name if set, otherwise name).
func (i ProjectItem) FilterValue() string {
	return project.EffectiveName(i.Project)
}

// Title returns the title for the list item.
// Returns the effective name (display name if set, otherwise name).
func (i ProjectItem) Title() string {
	return project.EffectiveName(i.Project)
}

// Description returns the description for the list item.
// Returns the current stage as a string.
func (i ProjectItem) Description() string {
	return i.Project.CurrentStage.String()
}

// EffectiveName returns the display name if set, otherwise the name.
// Convenience method that wraps the shared project.EffectiveName function.
func (i ProjectItem) EffectiveName() string {
	return project.EffectiveName(i.Project)
}
