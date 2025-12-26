// Package project provides shared project utilities.
package project

import (
	"sort"
	"strings"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// EffectiveName returns DisplayName if set, otherwise Name.
// Used for display and sorting.
func EffectiveName(p *domain.Project) string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.Name
}

// SortByName sorts projects with favorites first, then alphabetically by effective name (case-insensitive).
// Story 8.5: Primary sort by favorite status (true before false), secondary sort alphabetically.
func SortByName(projects []*domain.Project) {
	sort.Slice(projects, func(i, j int) bool {
		// Primary sort: favorites first
		if projects[i].IsFavorite != projects[j].IsFavorite {
			return projects[i].IsFavorite // true before false
		}
		// Secondary sort: alphabetical by effective name (case-insensitive)
		nameI := EffectiveName(projects[i])
		nameJ := EffectiveName(projects[j])
		return strings.ToLower(nameI) < strings.ToLower(nameJ)
	})
}
