package components

import (
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestProjectItem_FilterValue(t *testing.T) {
	tests := []struct {
		name     string
		project  *domain.Project
		expected string
	}{
		{
			name: "returns name when no display name",
			project: &domain.Project{
				Name:        "my-project",
				DisplayName: "",
			},
			expected: "my-project",
		},
		{
			name: "returns display name when set",
			project: &domain.Project{
				Name:        "original",
				DisplayName: "Custom Name",
			},
			expected: "Custom Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := ProjectItem{Project: tt.project}
			got := item.FilterValue()
			if got != tt.expected {
				t.Errorf("FilterValue() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestProjectItem_Title(t *testing.T) {
	tests := []struct {
		name     string
		project  *domain.Project
		expected string
	}{
		{
			name: "returns name when no display name",
			project: &domain.Project{
				Name:        "project-name",
				DisplayName: "",
			},
			expected: "project-name",
		},
		{
			name: "returns display name when set",
			project: &domain.Project{
				Name:        "internal-name",
				DisplayName: "Display Title",
			},
			expected: "Display Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := ProjectItem{Project: tt.project}
			got := item.Title()
			if got != tt.expected {
				t.Errorf("Title() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestProjectItem_Description(t *testing.T) {
	tests := []struct {
		name     string
		stage    domain.Stage
		expected string
	}{
		{"specify stage", domain.StageSpecify, "Specify"},
		{"plan stage", domain.StagePlan, "Plan"},
		{"tasks stage", domain.StageTasks, "Tasks"},
		{"implement stage", domain.StageImplement, "Implement"},
		{"unknown stage", domain.StageUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &domain.Project{
				Name:         "test",
				CurrentStage: tt.stage,
			}
			item := ProjectItem{Project: project}
			got := item.Description()
			if got != tt.expected {
				t.Errorf("Description() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestProjectItem_EffectiveName(t *testing.T) {
	tests := []struct {
		name     string
		project  *domain.Project
		expected string
	}{
		{
			name: "returns name when no display name",
			project: &domain.Project{
				Name:        "base-name",
				DisplayName: "",
			},
			expected: "base-name",
		},
		{
			name: "returns display name when set",
			project: &domain.Project{
				Name:        "base-name",
				DisplayName: "Friendly Name",
			},
			expected: "Friendly Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := ProjectItem{Project: tt.project}
			got := item.EffectiveName()
			if got != tt.expected {
				t.Errorf("EffectiveName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestProjectItem_ImplementsListItem(t *testing.T) {
	// Verify that ProjectItem implements the list.Item interface
	// by ensuring all required methods exist and return correct types
	project := &domain.Project{
		Name:         "test-project",
		DisplayName:  "Test Project",
		CurrentStage: domain.StageImplement,
	}
	item := ProjectItem{Project: project}

	// These calls verify the interface is implemented
	_ = item.FilterValue()
	_ = item.Title()
	_ = item.Description()

	// Also verify consistency: FilterValue and Title should match
	if item.FilterValue() != item.Title() {
		t.Errorf("FilterValue() = %q, Title() = %q, but they should match", item.FilterValue(), item.Title())
	}
}
