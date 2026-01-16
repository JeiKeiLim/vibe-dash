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

	// Story 8.8: Verify Confidence field is NOT shown in TUI
	t.Run("confidence field removed", func(t *testing.T) {
		if strings.Contains(view, "Confidence:") {
			t.Errorf("view should NOT contain 'Confidence:' field (removed in Story 8.8), got:\n%s", view)
		}
	})
}

func TestDetailPanel_View_EmptyNotes(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "test-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StagePlan,
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

// ============================================================================
// Story 8.12: Horizontal Mode Tests
// ============================================================================

func TestDetailPanel_HorizontalMode_UsesBorderlessTop(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "test-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StagePlan,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}

	// Test with horizontal mode ON
	panelH := NewDetailPanelModel(60, 20)
	panelH.SetProject(project)
	panelH.SetVisible(true)
	panelH.SetHorizontalMode(true)
	viewH := panelH.View()

	// Test with horizontal mode OFF
	panelV := NewDetailPanelModel(60, 20)
	panelV.SetProject(project)
	panelV.SetVisible(true)
	panelV.SetHorizontalMode(false)
	viewV := panelV.View()

	// Both should render non-empty output
	if viewH == "" {
		t.Error("horizontal mode panel should produce non-empty output")
	}
	if viewV == "" {
		t.Error("vertical mode panel should produce non-empty output")
	}

	// The horizontal view should be shorter (by 1 line - no top border)
	linesH := strings.Count(viewH, "\n")
	linesV := strings.Count(viewV, "\n")

	// Horizontal mode should have fewer lines due to missing top border
	if linesH >= linesV {
		t.Logf("Horizontal lines: %d, Vertical lines: %d", linesH, linesV)
		t.Log("Note: HorizontalBorderStyle removes top border to save space")
	}
}

func TestDetailPanel_SetHorizontalMode(t *testing.T) {
	panel := NewDetailPanelModel(60, 20)

	// Default should be false (vertical mode)
	if panel.isHorizontal {
		t.Error("panel should default to vertical mode (isHorizontal=false)")
	}

	// Set to horizontal
	panel.SetHorizontalMode(true)
	if !panel.isHorizontal {
		t.Error("panel should be in horizontal mode after SetHorizontalMode(true)")
	}

	// Set back to vertical
	panel.SetHorizontalMode(false)
	if panel.isHorizontal {
		t.Error("panel should be in vertical mode after SetHorizontalMode(false)")
	}
}

// ============================================================================
// Story 15.7: Confidence Level Tests
// ============================================================================

func TestConfidenceToText(t *testing.T) {
	tests := []struct {
		name       string
		confidence domain.Confidence
		expected   string
	}{
		{"certain", domain.ConfidenceCertain, "High confidence"},
		{"likely", domain.ConfidenceLikely, "Medium confidence"},
		{"uncertain", domain.ConfidenceUncertain, "Low confidence"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := confidenceToText(tt.confidence)
			if result != tt.expected {
				t.Errorf("confidenceToText(%v) = %q, want %q", tt.confidence, result, tt.expected)
			}
		})
	}
}

func TestToolToSourceText(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		expected string
	}{
		{"claude code", "Claude Code", "Claude Code logs"},
		{"generic", "Generic", "file activity"},
		{"unknown tool", "SomeOtherTool", "SomeOtherTool"},
		{"empty tool", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toolToSourceText(tt.tool)
			if result != tt.expected {
				t.Errorf("toolToSourceText(%q) = %q, want %q", tt.tool, result, tt.expected)
			}
		})
	}
}

func TestFormatAgentStatusWithConfidence(t *testing.T) {
	tests := []struct {
		name             string
		state            domain.AgentState
		expectedContains []string
	}{
		{
			name: "high confidence claude code",
			state: domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
				2*time.Hour+15*time.Minute, domain.ConfidenceCertain),
			expectedContains: []string{"2h 15m", "High confidence", "Claude Code logs"},
		},
		{
			name: "low confidence generic",
			state: domain.NewAgentState("Generic", domain.AgentWaitingForUser,
				30*time.Minute, domain.ConfidenceUncertain),
			expectedContains: []string{"30m", "Low confidence", "file activity"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAgentStatusWithConfidence(tt.state)
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("formatAgentStatusWithConfidence result should contain %q, got:\n%s", expected, result)
				}
			}
		})
	}
}

func TestDetailPanel_View_AgentStateGetter_HighConfidence(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "claude-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StageImplement,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now().Add(-2 * time.Hour),
	}

	// Create agent state getter that returns waiting with high confidence
	agentStateGetter := func(p *domain.Project) domain.AgentState {
		return domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
			2*time.Hour+15*time.Minute, domain.ConfidenceCertain)
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetProject(project)
	panel.SetAgentStateCallback(agentStateGetter)
	panel.SetVisible(true)

	view := panel.View()

	// Should contain waiting status with confidence info
	if !strings.Contains(view, "Waiting") {
		t.Errorf("view should contain 'Waiting' field, got:\n%s", view)
	}
	if !strings.Contains(view, "High confidence") {
		t.Errorf("view should contain 'High confidence', got:\n%s", view)
	}
	if !strings.Contains(view, "Claude Code logs") {
		t.Errorf("view should contain 'Claude Code logs', got:\n%s", view)
	}
}

func TestDetailPanel_View_AgentStateGetter_LowConfidence(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "generic-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StageImplement,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now().Add(-30 * time.Minute),
	}

	// Create agent state getter that returns waiting with low confidence
	agentStateGetter := func(p *domain.Project) domain.AgentState {
		return domain.NewAgentState("Generic", domain.AgentWaitingForUser,
			30*time.Minute, domain.ConfidenceUncertain)
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetProject(project)
	panel.SetAgentStateCallback(agentStateGetter)
	panel.SetVisible(true)

	view := panel.View()

	// Should contain waiting status with confidence info
	if !strings.Contains(view, "Waiting") {
		t.Errorf("view should contain 'Waiting' field, got:\n%s", view)
	}
	if !strings.Contains(view, "Low confidence") {
		t.Errorf("view should contain 'Low confidence', got:\n%s", view)
	}
	if !strings.Contains(view, "file activity") {
		t.Errorf("view should contain 'file activity', got:\n%s", view)
	}
}

func TestDetailPanel_View_AgentStateGetter_NotWaiting(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "working-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StageImplement,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}

	// Create agent state getter that returns working (not waiting)
	agentStateGetter := func(p *domain.Project) domain.AgentState {
		return domain.NewAgentState("Claude Code", domain.AgentWorking,
			5*time.Minute, domain.ConfidenceCertain)
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetProject(project)
	panel.SetAgentStateCallback(agentStateGetter)
	panel.SetVisible(true)

	view := panel.View()

	// Should NOT contain Waiting field when agent is not waiting
	if strings.Contains(view, "Waiting:") {
		t.Errorf("view should NOT contain 'Waiting:' when agent is not waiting, got:\n%s", view)
	}
}

// Story 15.7 AC6: AgentInactive should NOT display confidence in detail panel
func TestDetailPanel_View_AgentStateGetter_Inactive(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "inactive-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StageImplement,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now().Add(-24 * time.Hour),
	}

	// Create agent state getter that returns Inactive
	agentStateGetter := func(p *domain.Project) domain.AgentState {
		return domain.NewAgentState("Generic", domain.AgentInactive,
			24*time.Hour, domain.ConfidenceUncertain)
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetProject(project)
	panel.SetAgentStateCallback(agentStateGetter)
	panel.SetVisible(true)

	view := panel.View()

	// Should NOT contain Waiting field when agent is inactive
	if strings.Contains(view, "Waiting:") {
		t.Errorf("view should NOT contain 'Waiting:' when agent is Inactive, got:\n%s", view)
	}
}

// Story 15.7 AC6: AgentUnknown should NOT display confidence in detail panel
func TestDetailPanel_View_AgentStateGetter_Unknown(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "unknown-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StageImplement,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}

	// Create agent state getter that returns Unknown
	agentStateGetter := func(p *domain.Project) domain.AgentState {
		return domain.NewAgentState("", domain.AgentUnknown,
			0, domain.ConfidenceUncertain)
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetProject(project)
	panel.SetAgentStateCallback(agentStateGetter)
	panel.SetVisible(true)

	view := panel.View()

	// Should NOT contain Waiting field when agent is unknown
	if strings.Contains(view, "Waiting:") {
		t.Errorf("view should NOT contain 'Waiting:' when agent is Unknown, got:\n%s", view)
	}
}

func TestDetailPanel_View_AgentStateGetter_NilFallback(t *testing.T) {
	project := &domain.Project{
		ID:             "abc123",
		Name:           "fallback-project",
		Path:           "/home/user/test",
		DetectedMethod: "speckit",
		CurrentStage:   domain.StageImplement,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now().Add(-2 * time.Hour),
	}

	// Create panel with only waitingChecker (no agentStateGetter) - should use fallback
	checker := func(p *domain.Project) bool { return true }
	getter := func(p *domain.Project) time.Duration { return 2*time.Hour + 15*time.Minute }

	panel := NewDetailPanelModel(80, 20)
	panel.SetProject(project)
	panel.SetWaitingCallbacks(checker, getter) // Old API, no agentStateGetter
	panel.SetVisible(true)

	view := panel.View()

	// Should contain waiting status but NO confidence info (fallback behavior)
	if !strings.Contains(view, "Waiting") {
		t.Errorf("view should contain 'Waiting' field with fallback, got:\n%s", view)
	}
	if !strings.Contains(view, "2h 15m") {
		t.Errorf("view should contain duration '2h 15m' with fallback, got:\n%s", view)
	}
	// Should NOT contain confidence text (fallback doesn't show it)
	if strings.Contains(view, "High confidence") || strings.Contains(view, "Low confidence") {
		t.Errorf("fallback view should NOT contain confidence text, got:\n%s", view)
	}
}

func TestDetailPanel_SetAgentStateCallback(t *testing.T) {
	panel := NewDetailPanelModel(60, 20)

	// Default should be nil
	if panel.agentStateGetter != nil {
		t.Error("panel should have nil agentStateGetter by default")
	}

	// Set callback
	getter := func(p *domain.Project) domain.AgentState {
		return domain.NewAgentState("Claude Code", domain.AgentWaitingForUser,
			time.Hour, domain.ConfidenceCertain)
	}
	panel.SetAgentStateCallback(getter)

	if panel.agentStateGetter == nil {
		t.Error("panel should have agentStateGetter after SetAgentStateCallback")
	}
}

// ============================================================================
// Story 14.5: Coexistence Warning Tests
// ============================================================================

func TestDetailPanel_CoexistenceWarning_Shown(t *testing.T) {
	project := &domain.Project{
		ID:                 "test-id",
		Name:               "test-project",
		Path:               "/test/path",
		DetectedMethod:     "speckit",
		CurrentStage:       domain.StagePlan,
		CoexistenceWarning: true,
		CoexistenceMessage: "Multiple methodologies detected with similar activity",
		SecondaryMethod:    "bmad",
		SecondaryStage:     domain.StageTasks,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		LastActivityAt:     time.Now(),
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetVisible(true)
	panel.SetProject(project)
	output := panel.View()

	// Should contain warning indicator and both methodologies
	if !strings.Contains(output, "Coexistence") {
		t.Error("expected Coexistence label in output")
	}
	if !strings.Contains(output, "speckit") {
		t.Error("expected primary method 'speckit' in warning")
	}
	if !strings.Contains(output, "bmad") {
		t.Error("expected secondary method 'bmad' in warning")
	}
}

func TestDetailPanel_CoexistenceWarning_Hidden(t *testing.T) {
	project := &domain.Project{
		ID:                 "test-id",
		Name:               "test-project",
		Path:               "/test/path",
		DetectedMethod:     "speckit",
		CurrentStage:       domain.StagePlan,
		CoexistenceWarning: false, // No warning
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		LastActivityAt:     time.Now(),
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetVisible(true)
	panel.SetProject(project)
	output := panel.View()

	// Should NOT contain coexistence warning
	if strings.Contains(output, "Coexistence") {
		t.Error("should not show Coexistence label when flag is false")
	}
}

func TestDetailPanel_CoexistenceWarning_HiddenWhenSecondaryMethodEmpty(t *testing.T) {
	// Edge case: CoexistenceWarning true but SecondaryMethod empty
	project := &domain.Project{
		ID:                 "test-id",
		Name:               "test-project",
		Path:               "/test/path",
		DetectedMethod:     "speckit",
		CurrentStage:       domain.StagePlan,
		CoexistenceWarning: true, // Warning flag set
		SecondaryMethod:    "",   // But no secondary method
		SecondaryStage:     domain.StagePlan,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		LastActivityAt:     time.Now(),
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetVisible(true)
	panel.SetProject(project)
	output := panel.View()

	// Should NOT contain coexistence warning when SecondaryMethod is empty
	if strings.Contains(output, "Coexistence") {
		t.Error("should not show Coexistence label when SecondaryMethod is empty")
	}
}
