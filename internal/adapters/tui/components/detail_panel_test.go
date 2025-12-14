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
