package components

import (
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestDetailPanel_View_BasicFields(t *testing.T) {
	// Create a project with all fields populated
	createdAt := time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)
	lastActivity := time.Now().Add(-2 * time.Hour)

	project := &domain.Project{
		ID:                 "abc123",
		Name:               "test-project",
		Path:               "/home/user/projects/test-project",
		DisplayName:        "My Test Project",
		DetectedMethod:     "speckit",
		CurrentStage:       domain.StagePlan,
		Confidence:         domain.ConfidenceCertain,
		DetectionReasoning: "plan.md exists, no tasks.md",
		Notes:              "Some notes here",
		CreatedAt:          createdAt,
		LastActivityAt:     lastActivity,
	}

	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(project)
	panel.SetVisible(true)

	view := panel.View()

	// Verify all required fields are present
	tests := []struct {
		name     string
		contains string
	}{
		{"panel title", "DETAILS: My Test Project"},
		{"path field", "/home/user/projects/test-project"},
		{"method field", "speckit"},
		{"stage field", "Plan"},
		{"confidence field", "Certain"},
		{"detection field", "plan.md exists, no tasks.md"},
		{"notes field", "Some notes here"},
		{"added date", "2025-12-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(view, tt.contains) {
				t.Errorf("view should contain %q for %s, got:\n%s", tt.contains, tt.name, view)
			}
		})
	}
}

func TestDetailPanel_View_EmptyNotes(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "test-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StagePlan,
		Confidence:     domain.ConfidenceCertain,
		Notes:          "", // Empty notes
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}

	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(project)
	panel.SetVisible(true)

	view := panel.View()

	if !strings.Contains(view, "(none)") {
		t.Errorf("view should show '(none)' for empty notes, got:\n%s", view)
	}
}

func TestDetailPanel_View_UncertainConfidence(t *testing.T) {
	project := &domain.Project{
		ID:                 "abc123",
		Name:               "test-project",
		Path:               "/home/user/test",
		DetectedMethod:     "speckit",
		CurrentStage:       domain.StagePlan,
		Confidence:         domain.ConfidenceUncertain,
		DetectionReasoning: "Unable to determine stage",
		CreatedAt:          time.Now(),
		LastActivityAt:     time.Now(),
	}

	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(project)
	panel.SetVisible(true)

	view := panel.View()

	// Should contain "Uncertain" (styled or not)
	if !strings.Contains(view, "Uncertain") {
		t.Errorf("view should show 'Uncertain' for uncertain confidence, got:\n%s", view)
	}
}

func TestDetailPanel_View_NilProject(t *testing.T) {
	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(nil)
	panel.SetVisible(true)

	view := panel.View()

	// Should not panic and should show placeholder
	if view == "" {
		t.Error("view should not be empty for nil project")
	}
	if !strings.Contains(view, "No project selected") {
		t.Errorf("view should show 'No project selected' for nil project, got:\n%s", view)
	}
}

func TestDetailPanel_SetSize(t *testing.T) {
	panel := NewDetailPanelModel(60, 20)

	if panel.width != 60 || panel.height != 20 {
		t.Errorf("initial size should be 60x20, got %dx%d", panel.width, panel.height)
	}

	panel.SetSize(80, 30)

	if panel.width != 80 || panel.height != 30 {
		t.Errorf("size should be 80x30 after SetSize, got %dx%d", panel.width, panel.height)
	}
}

func TestDetailPanel_Visibility(t *testing.T) {
	panel := NewDetailPanelModel(60, 20)

	// Default should be not visible
	if panel.IsVisible() {
		t.Error("panel should not be visible by default")
	}

	panel.SetVisible(true)
	if !panel.IsVisible() {
		t.Error("panel should be visible after SetVisible(true)")
	}

	panel.SetVisible(false)
	if panel.IsVisible() {
		t.Error("panel should not be visible after SetVisible(false)")
	}
}

func TestDetailPanel_View_EmptyDetectionReasoning(t *testing.T) {
	project := &domain.Project{
		ID:                 "abc123",
		Name:               "test-project",
		Path:               "/home/user/test",
		DetectedMethod:     "speckit",
		CurrentStage:       domain.StagePlan,
		Confidence:         domain.ConfidenceCertain,
		DetectionReasoning: "", // Empty reasoning
		CreatedAt:          time.Now(),
		LastActivityAt:     time.Now(),
	}

	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(project)
	panel.SetVisible(true)

	view := panel.View()

	// Should show placeholder for empty detection reasoning
	if !strings.Contains(view, "No detection reasoning available") {
		t.Errorf("view should show placeholder for empty detection reasoning, got:\n%s", view)
	}
}

func TestDetailPanel_View_NotVisible(t *testing.T) {
	project := &domain.Project{
		ID:   "abc123",
		Name: "test-project",
		Path: "/home/user/test",
	}

	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(project)
	panel.SetVisible(false)

	view := panel.View()

	// Should return empty when not visible
	if view != "" {
		t.Errorf("view should be empty when not visible, got:\n%s", view)
	}
}

func TestDetailPanel_View_LikelyConfidence(t *testing.T) {
	project := &domain.Project{
		ID:                 "abc123",
		Name:               "test-project",
		Path:               "/home/user/test",
		DetectedMethod:     "speckit",
		CurrentStage:       domain.StagePlan,
		Confidence:         domain.ConfidenceLikely,
		DetectionReasoning: "Some indicators present",
		CreatedAt:          time.Now(),
		LastActivityAt:     time.Now(),
	}

	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(project)
	panel.SetVisible(true)

	view := panel.View()

	if !strings.Contains(view, "Likely") {
		t.Errorf("view should show 'Likely' for likely confidence, got:\n%s", view)
	}
}

// ============================================================================
// Story 4.5: Waiting Field Tests
// ============================================================================

func TestDetailPanel_View_WaitingField_Shown(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "waiting-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StageImplement,
		Confidence:     domain.ConfidenceCertain,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now().Add(-2*time.Hour - 15*time.Minute),
	}

	// Create panel with waiting callbacks
	checker := func(p *domain.Project) bool { return true }
	getter := func(p *domain.Project) time.Duration { return 2*time.Hour + 15*time.Minute }

	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(project)
	panel.SetWaitingCallbacks(checker, getter)
	panel.SetVisible(true)

	view := panel.View()

	// Should contain Waiting field with detailed duration
	if !strings.Contains(view, "Waiting") {
		t.Errorf("view should contain 'Waiting' field when project is waiting, got:\n%s", view)
	}
	// Detailed format: "2h 15m"
	if !strings.Contains(view, "2h 15m") {
		t.Errorf("view should contain detailed duration '2h 15m', got:\n%s", view)
	}
}

func TestDetailPanel_View_WaitingField_Hidden(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "active-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StageImplement,
		Confidence:     domain.ConfidenceCertain,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}

	// Create panel with waiting callbacks - not waiting
	checker := func(p *domain.Project) bool { return false }
	getter := func(p *domain.Project) time.Duration { return 0 }

	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(project)
	panel.SetWaitingCallbacks(checker, getter)
	panel.SetVisible(true)

	view := panel.View()

	// Should NOT contain Waiting field when not waiting
	if strings.Contains(view, "Waiting:") {
		t.Errorf("view should NOT contain 'Waiting:' field when project is not waiting, got:\n%s", view)
	}
}

func TestDetailPanel_View_WaitingField_NilCallbacks(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "test-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StageImplement,
		Confidence:     domain.ConfidenceCertain,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now().Add(-2 * time.Hour),
	}

	// No callbacks set - should not show waiting
	panel := NewDetailPanelModel(60, 20)
	panel.SetProject(project)
	panel.SetVisible(true)

	view := panel.View()

	// Should NOT contain Waiting field when no callbacks
	if strings.Contains(view, "Waiting:") {
		t.Errorf("view should NOT contain 'Waiting:' when no callbacks set, got:\n%s", view)
	}
}

func TestDetailPanel_View_WaitingField_DurationFormats(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"15 minutes", 15 * time.Minute, "15m"},
		{"1 hour", 1 * time.Hour, "1h 0m"},
		{"2h 15m", 2*time.Hour + 15*time.Minute, "2h 15m"},
		{"1 day 5h", 24*time.Hour + 5*time.Hour, "1d 5h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &domain.Project{
				ID:             "abc123",
				Name:           "test-project",
				Path:           "/home/user/test",
				CurrentStage:   domain.StageImplement,
				CreatedAt:      time.Now(),
				LastActivityAt: time.Now().Add(-tt.duration),
			}

			checker := func(p *domain.Project) bool { return true }
			getter := func(p *domain.Project) time.Duration { return tt.duration }

			panel := NewDetailPanelModel(60, 20)
			panel.SetProject(project)
			panel.SetWaitingCallbacks(checker, getter)
			panel.SetVisible(true)

			view := panel.View()

			if !strings.Contains(view, tt.expected) {
				t.Errorf("view should contain %q for %v duration, got:\n%s", tt.expected, tt.duration, view)
			}
		})
	}
}
