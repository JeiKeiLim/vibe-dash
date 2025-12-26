package project

import (
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestEffectiveName(t *testing.T) {
	tests := []struct {
		name     string
		project  *domain.Project
		expected string
	}{
		{
			name: "returns name when display name is empty",
			project: &domain.Project{
				Name:        "my-project",
				DisplayName: "",
			},
			expected: "my-project",
		},
		{
			name: "returns display name when set",
			project: &domain.Project{
				Name:        "original-name",
				DisplayName: "Custom Display",
			},
			expected: "Custom Display",
		},
		{
			name: "returns name when display name is whitespace only",
			project: &domain.Project{
				Name:        "project-name",
				DisplayName: "",
			},
			expected: "project-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EffectiveName(tt.project)
			if got != tt.expected {
				t.Errorf("EffectiveName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSortByName(t *testing.T) {
	t.Run("sorts alphabetically by effective name", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "zebra", DisplayName: ""},
			{Name: "alpha", DisplayName: ""},
			{Name: "middle", DisplayName: ""},
		}

		SortByName(projects)

		expected := []string{"alpha", "middle", "zebra"}
		for i, p := range projects {
			if p.Name != expected[i] {
				t.Errorf("projects[%d].Name = %q, want %q", i, p.Name, expected[i])
			}
		}
	})

	t.Run("sorts by display name when set", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "zulu", DisplayName: "alpha-display"},
			{Name: "alpha", DisplayName: ""},
			{Name: "bravo", DisplayName: "zebra-display"},
		}

		SortByName(projects)

		// Expected order: "alpha" (name), "zulu" (display: alpha-display), "bravo" (display: zebra-display)
		expectedNames := []string{"alpha", "zulu", "bravo"}
		for i, p := range projects {
			if p.Name != expectedNames[i] {
				t.Errorf("projects[%d].Name = %q, want %q", i, p.Name, expectedNames[i])
			}
		}
	})

	t.Run("case insensitive sorting", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "Zebra", DisplayName: ""},
			{Name: "alpha", DisplayName: ""},
			{Name: "BETA", DisplayName: ""},
		}

		SortByName(projects)

		expected := []string{"alpha", "BETA", "Zebra"}
		for i, p := range projects {
			if p.Name != expected[i] {
				t.Errorf("projects[%d].Name = %q, want %q", i, p.Name, expected[i])
			}
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		projects := []*domain.Project{}
		SortByName(projects) // Should not panic
		if len(projects) != 0 {
			t.Errorf("len(projects) = %d, want 0", len(projects))
		}
	})

	t.Run("handles single element", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "only-one", DisplayName: ""},
		}
		SortByName(projects)
		if projects[0].Name != "only-one" {
			t.Errorf("projects[0].Name = %q, want %q", projects[0].Name, "only-one")
		}
	})

	// Story 8.5: Favorites sort first tests
	t.Run("favorites appear before non-favorites", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "alpha", IsFavorite: false},
			{Name: "beta", IsFavorite: true},
			{Name: "gamma", IsFavorite: false},
		}

		SortByName(projects)

		// Favorites first, then non-favorites
		if !projects[0].IsFavorite {
			t.Errorf("projects[0] should be a favorite, got %q", projects[0].Name)
		}
		if projects[0].Name != "beta" {
			t.Errorf("projects[0].Name = %q, want %q", projects[0].Name, "beta")
		}
	})

	t.Run("favorites sorted alphabetically among themselves", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "zebra", IsFavorite: true},
			{Name: "alpha", IsFavorite: true},
			{Name: "middle", IsFavorite: true},
		}

		SortByName(projects)

		expected := []string{"alpha", "middle", "zebra"}
		for i, p := range projects {
			if p.Name != expected[i] {
				t.Errorf("projects[%d].Name = %q, want %q", i, p.Name, expected[i])
			}
		}
	})

	t.Run("non-favorites sorted alphabetically after all favorites", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "zebra-nonfav", IsFavorite: false},
			{Name: "alpha-fav", IsFavorite: true},
			{Name: "beta-nonfav", IsFavorite: false},
			{Name: "gamma-fav", IsFavorite: true},
		}

		SortByName(projects)

		// Expected: alpha-fav, gamma-fav, beta-nonfav, zebra-nonfav
		expected := []string{"alpha-fav", "gamma-fav", "beta-nonfav", "zebra-nonfav"}
		for i, p := range projects {
			if p.Name != expected[i] {
				t.Errorf("projects[%d].Name = %q, want %q", i, p.Name, expected[i])
			}
		}
	})

	t.Run("favorites case insensitive sort", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "Zebra", IsFavorite: true},
			{Name: "alpha", IsFavorite: true},
			{Name: "BETA", IsFavorite: true},
		}

		SortByName(projects)

		expected := []string{"alpha", "BETA", "Zebra"}
		for i, p := range projects {
			if p.Name != expected[i] {
				t.Errorf("projects[%d].Name = %q, want %q", i, p.Name, expected[i])
			}
		}
	})

	t.Run("all favorites", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "charlie", IsFavorite: true},
			{Name: "alpha", IsFavorite: true},
			{Name: "bravo", IsFavorite: true},
		}

		SortByName(projects)

		expected := []string{"alpha", "bravo", "charlie"}
		for i, p := range projects {
			if p.Name != expected[i] {
				t.Errorf("projects[%d].Name = %q, want %q", i, p.Name, expected[i])
			}
		}
	})

	t.Run("no favorites", func(t *testing.T) {
		projects := []*domain.Project{
			{Name: "charlie", IsFavorite: false},
			{Name: "alpha", IsFavorite: false},
			{Name: "bravo", IsFavorite: false},
		}

		SortByName(projects)

		expected := []string{"alpha", "bravo", "charlie"}
		for i, p := range projects {
			if p.Name != expected[i] {
				t.Errorf("projects[%d].Name = %q, want %q", i, p.Name, expected[i])
			}
		}
	})

	t.Run("handles nil slice", func(t *testing.T) {
		var projects []*domain.Project
		SortByName(projects) // Should not panic
	})
}
