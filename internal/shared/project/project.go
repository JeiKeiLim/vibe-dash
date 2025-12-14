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

// SortByName sorts projects alphabetically by effective name (case-insensitive).
func SortByName(projects []*domain.Project) {
	sort.Slice(projects, func(i, j int) bool {
		nameI := EffectiveName(projects[i])
		nameJ := EffectiveName(projects[j])
		return strings.ToLower(nameI) < strings.ToLower(nameJ)
	})
}
