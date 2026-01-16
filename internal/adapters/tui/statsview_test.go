package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/statsview"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// Story 16.3: Tests for Stats View functionality

// TestEnterStatsView_CapturesSelection verifies that entering stats view saves
// the current project list selection index for restoration (AC2).
func TestEnterStatsView_CapturesSelection(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30

	// Create and set projects
	projects := []*domain.Project{
		{ID: "p1", Name: "project-a", Path: "/path/a", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "p2", Name: "project-b", Path: "/path/b", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "p3", Name: "project-c", Path: "/path/c", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(nil, 100, 20)
	m.projectList.SetProjects(projects)

	// Select the second project (index 1)
	m.projectList.SelectByIndex(1)
	originalIndex := m.projectList.Index()

	// Enter stats view
	m.enterStatsView()

	// Verify captured selection
	if m.statsActiveProjectIdx != originalIndex {
		t.Errorf("Expected statsActiveProjectIdx to be %d, got %d", originalIndex, m.statsActiveProjectIdx)
	}

	// Verify view mode changed
	if m.viewMode != viewModeStats {
		t.Error("Expected viewMode to be viewModeStats after enterStatsView()")
	}

	// Verify scroll position reset
	if m.statsViewScroll != 0 {
		t.Errorf("Expected statsViewScroll to be 0, got %d", m.statsViewScroll)
	}
}

// TestExitStatsView_RestoresSelection verifies that exiting stats view restores
// the original project list selection (AC2).
func TestExitStatsView_RestoresSelection(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30

	// Create and set projects
	projects := []*domain.Project{
		{ID: "p1", Name: "project-a", Path: "/path/a", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "p2", Name: "project-b", Path: "/path/b", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "p3", Name: "project-c", Path: "/path/c", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(nil, 100, 20)
	m.projectList.SetProjects(projects)

	// Select the third project (index 2) and enter stats view
	m.projectList.SelectByIndex(2)
	m.enterStatsView()

	// Exit stats view
	m.exitStatsView()

	// Verify view mode restored
	if m.viewMode != viewModeNormal {
		t.Error("Expected viewMode to be viewModeNormal after exitStatsView()")
	}

	// Verify selection restored
	if m.projectList.Index() != 2 {
		t.Errorf("Expected project list index to be restored to 2, got %d", m.projectList.Index())
	}
}

// TestExitStatsView_BoundsCheck verifies that exiting stats view handles
// the case where the saved index is no longer valid (e.g., projects were removed).
func TestExitStatsView_BoundsCheck(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30

	// Create and set projects
	projects := []*domain.Project{
		{ID: "p1", Name: "project-a", Path: "/path/a", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "p2", Name: "project-b", Path: "/path/b", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "p3", Name: "project-c", Path: "/path/c", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(nil, 100, 20)
	m.projectList.SetProjects(projects)

	// Select the third project and enter stats view
	m.projectList.SelectByIndex(2)
	m.enterStatsView()

	// Simulate projects being reduced while in stats view
	m.projects = projects[:1] // Only one project remains
	m.projectList.SetProjects(m.projects)

	// Exit stats view - should not panic, and should not restore invalid index
	m.exitStatsView()

	// Verify view mode restored
	if m.viewMode != viewModeNormal {
		t.Error("Expected viewMode to be viewModeNormal after exitStatsView()")
	}

	// Selection should remain at a valid index (bounds check should prevent restore)
	idx := m.projectList.Index()
	if idx < 0 || idx >= len(m.projects) {
		t.Errorf("Project list index %d is out of bounds for %d projects", idx, len(m.projects))
	}
}

// TestExitStatsView_NegativeIndex verifies bounds check handles negative index.
func TestExitStatsView_NegativeIndex(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30

	projects := []*domain.Project{
		{ID: "p1", Name: "project-a", Path: "/path/a", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(nil, 100, 20)
	m.projectList.SetProjects(projects)

	// Manually set invalid negative index
	m.statsActiveProjectIdx = -1
	m.viewMode = viewModeStats

	// Exit should not panic
	m.exitStatsView()

	if m.viewMode != viewModeNormal {
		t.Error("Expected viewMode to be viewModeNormal after exitStatsView()")
	}
}

// TestHandleStatsViewKeyMsg_EscExits verifies Esc key exits Stats View (AC2).
func TestHandleStatsViewKeyMsg_EscExits(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30

	projects := []*domain.Project{
		{ID: "p1", Name: "project-a", Path: "/path/a", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(nil, 100, 20)
	m.projectList.SetProjects(projects)

	// Enter stats view
	m.enterStatsView()

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	result, _ := m.handleStatsViewKeyMsg(msg)

	if result.viewMode != viewModeNormal {
		t.Error("Esc key should exit Stats View and return to viewModeNormal")
	}
}

// TestHandleStatsViewKeyMsg_QExits verifies 'q' key exits Stats View (AC2).
func TestHandleStatsViewKeyMsg_QExits(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30

	projects := []*domain.Project{
		{ID: "p1", Name: "project-a", Path: "/path/a", State: domain.StateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	m.projects = projects
	m.projectList = components.NewProjectListModel(nil, 100, 20)
	m.projectList.SetProjects(projects)

	// Enter stats view
	m.enterStatsView()

	// Press 'q'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	result, _ := m.handleStatsViewKeyMsg(msg)

	if result.viewMode != viewModeNormal {
		t.Error("'q' key should exit Stats View and return to viewModeNormal")
	}
}

// TestStatsViewKeyBinding verifies KeyStats constant is set correctly.
func TestStatsViewKeyBinding(t *testing.T) {
	if KeyStats != "s" {
		t.Errorf("Expected KeyStats to be 's', got %q", KeyStats)
	}

	bindings := DefaultKeyBindings()
	if bindings.Stats != "s" {
		t.Errorf("Expected DefaultKeyBindings().Stats to be 's', got %q", bindings.Stats)
	}
}

// TestRenderStatsView_Header verifies Stats View renders with correct header (AC3).
func TestRenderStatsView_Header(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30
	m.statusBar = components.NewStatusBarModel(100)

	m.enterStatsView()
	output := m.renderStatsView()

	// Verify "STATS" title in header
	if !strings.Contains(output, "STATS") {
		t.Error("Stats View header should contain 'STATS' title")
	}

	// Verify back hint in header
	if !strings.Contains(output, "[ESC] Back to Dashboard") {
		t.Error("Stats View header should contain '[ESC] Back to Dashboard' hint")
	}
}

// TestRenderStatsView_NoProjectsMessage verifies Stats View shows "No projects to display"
// when there are no projects (Story 16.4 graceful handling).
func TestRenderStatsView_NoProjectsMessage(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30
	m.statusBar = components.NewStatusBarModel(100)
	m.projects = nil // No projects

	m.enterStatsView()
	output := m.renderStatsView()

	// Verify no projects message
	if !strings.Contains(output, "No projects to display") {
		t.Error("Stats View should display 'No projects to display' when no projects")
	}
}

// TestStatsKey_BlockedDuringHelp verifies 's' key doesn't open Stats View
// when help overlay is shown.
func TestStatsKey_BlockedDuringHelp(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.showHelp = true
	m.viewMode = viewModeNormal

	// Press 's'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	result, _ := m.Update(msg)
	updated := result.(Model)

	// Should not enter stats view while help is shown
	if updated.viewMode == viewModeStats {
		t.Error("Stats View should not open while help overlay is shown")
	}
}

// TestStatsKey_BlockedDuringNoteEdit verifies 's' key doesn't open Stats View
// during note editing.
func TestStatsKey_BlockedDuringNoteEdit(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.isEditingNote = true
	m.viewMode = viewModeNormal

	// Press 's'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	result, _ := m.Update(msg)
	updated := result.(Model)

	// Should not enter stats view while editing note
	if updated.viewMode == viewModeStats {
		t.Error("Stats View should not open while editing note")
	}
}

// TestStatsKey_BlockedDuringRemoveConfirm verifies 's' key doesn't open Stats View
// during remove confirmation.
func TestStatsKey_BlockedDuringRemoveConfirm(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.isConfirmingRemove = true
	m.viewMode = viewModeNormal

	// Press 's'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	result, _ := m.Update(msg)
	updated := result.(Model)

	// Should not enter stats view while confirming remove
	if updated.viewMode == viewModeStats {
		t.Error("Stats View should not open while confirming remove")
	}
}

// TestStatsKey_OpensStatsView verifies 's' key opens Stats View from normal mode.
func TestStatsKey_OpensStatsView(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.viewMode = viewModeNormal
	m.projectList = components.NewProjectListModel(nil, 100, 20)

	// Press 's'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	result, _ := m.Update(msg)
	updated := result.(Model)

	// Should enter stats view
	if updated.viewMode != viewModeStats {
		t.Error("'s' key should open Stats View")
	}
}

// TestUpdateRoutesStatsViewKeyMsg verifies Update routes key messages
// to handleStatsViewKeyMsg when in Stats View mode.
func TestUpdateRoutesStatsViewKeyMsg(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.viewMode = viewModeStats
	m.projectList = components.NewProjectListModel(nil, 100, 20)

	// Press Esc - should be routed to handleStatsViewKeyMsg
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	result, _ := m.Update(msg)
	updated := result.(Model)

	// Should exit stats view
	if updated.viewMode != viewModeNormal {
		t.Error("Esc in Stats View should exit to viewModeNormal")
	}
}

// TestHandleStatsViewKeyMsg_ScrollKeys verifies scroll keys don't change view mode
// and don't panic (stub for future implementation).
func TestHandleStatsViewKeyMsg_ScrollKeys(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30
	m.viewMode = viewModeStats

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"j key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}},
		{"k key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}},
		{"down arrow", tea.KeyMsg{Type: tea.KeyDown}},
		{"up arrow", tea.KeyMsg{Type: tea.KeyUp}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, cmd := m.handleStatsViewKeyMsg(tt.key)

			// Scroll keys should not exit stats view
			if result.viewMode != viewModeStats {
				t.Errorf("%s should not exit Stats View", tt.name)
			}

			// Scroll keys return nil cmd (future story will implement actual scrolling)
			if cmd != nil {
				t.Errorf("%s should return nil cmd, got %v", tt.name, cmd)
			}
		})
	}
}

// Story 16.4: Tests for Sparkline Integration

// TestRenderStatsView_ShowsProjectList verifies Stats View renders project list
// with sparklines when projects exist.
func TestRenderStatsView_ShowsProjectList(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30
	m.statusBar = components.NewStatusBarModel(100)
	m.projects = []*domain.Project{
		{ID: "p1", Name: "project-alpha", Path: "/path/a", State: domain.StateActive, CurrentStage: domain.StageTasks, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "p2", Name: "project-beta", Path: "/path/b", State: domain.StateActive, CurrentStage: domain.StageImplement, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	m.enterStatsView()
	output := m.renderStatsView()

	// Should show project names
	if !strings.Contains(output, "project-alpha") {
		t.Error("Stats View should display project-alpha")
	}
	if !strings.Contains(output, "project-beta") {
		t.Error("Stats View should display project-beta")
	}

	// Should show column headers
	if !strings.Contains(output, "Project") {
		t.Error("Stats View should have 'Project' column header")
	}
	if !strings.Contains(output, "Activity (30d)") {
		t.Error("Stats View should have 'Activity (30d)' column header")
	}
	if !strings.Contains(output, "Stage") {
		t.Error("Stats View should have 'Stage' column header")
	}
}

// TestRenderStatsView_ShowsFlatSparklineWithoutMetrics verifies Stats View shows
// flat sparkline (▁▁▁▁▁▁▁) when metricsReader is nil (AC3, AC5: graceful degradation).
func TestRenderStatsView_ShowsFlatSparklineWithoutMetrics(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30
	m.statusBar = components.NewStatusBarModel(100)
	m.metricsReader = nil // No metrics reader (graceful degradation)
	m.projects = []*domain.Project{
		{ID: "p1", Name: "project-test", Path: "/path/test", State: domain.StateActive, CurrentStage: domain.StagePlan, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	m.enterStatsView()
	output := m.renderStatsView()

	// Should show flat sparkline characters (▁)
	if !strings.Contains(output, "▁▁▁") {
		t.Error("Stats View should show flat sparkline when metricsReader is nil")
	}
}

// TestGetSparklineBuckets_NarrowWidth verifies bucket count for narrow terminals (AC4).
func TestGetSparklineBuckets_NarrowWidth(t *testing.T) {
	m := NewModel(nil)

	// Very narrow - minimum 7 buckets
	if buckets := m.getSparklineBuckets(50); buckets != 7 {
		t.Errorf("Expected 7 buckets for width 50, got %d", buckets)
	}

	// Width 60 - should be 7 buckets
	if buckets := m.getSparklineBuckets(60); buckets != 7 {
		t.Errorf("Expected 7 buckets for width 60, got %d", buckets)
	}
}

// TestGetSparklineBuckets_WideWidth verifies bucket count for wide terminals (AC4).
func TestGetSparklineBuckets_WideWidth(t *testing.T) {
	m := NewModel(nil)

	// Width > 100 - max 14 buckets
	if buckets := m.getSparklineBuckets(120); buckets != 14 {
		t.Errorf("Expected 14 buckets for width 120, got %d", buckets)
	}

	// Width exactly 100 - max 14 buckets
	if buckets := m.getSparklineBuckets(100); buckets < 13 {
		t.Errorf("Expected at least 13 buckets for width 100, got %d", buckets)
	}
}

// TestGetSparklineBuckets_MidWidth verifies bucket count scales with width (AC4).
func TestGetSparklineBuckets_MidWidth(t *testing.T) {
	m := NewModel(nil)

	// Width 80 - should be between 7 and 14
	buckets := m.getSparklineBuckets(80)
	if buckets < 7 || buckets > 14 {
		t.Errorf("Expected buckets between 7-14 for width 80, got %d", buckets)
	}
}

// TestGetProjectActivity_NilReader verifies graceful degradation when metricsReader is nil.
func TestGetProjectActivity_NilReader(t *testing.T) {
	m := NewModel(nil)
	m.metricsReader = nil

	result := m.getProjectActivity("test-project", 7)

	if result != nil {
		t.Errorf("Expected nil when metricsReader is nil, got %v", result)
	}
}

// mockMetricsReader implements metricsReaderInterface for testing.
type mockMetricsReader struct {
	transitions []statsview.Transition
}

func (m *mockMetricsReader) GetTransitionTimestamps(_ context.Context, _ string, _ time.Time) []statsview.Transition {
	return m.transitions
}

// TestGetProjectActivity_WithMockReader verifies full integration path with mock reader.
func TestGetProjectActivity_WithMockReader(t *testing.T) {
	m := NewModel(nil)
	now := time.Now()

	// Create mock reader with transitions spread over 7 days
	mockReader := &mockMetricsReader{
		transitions: []statsview.Transition{
			{TransitionedAt: now.Add(-1 * 24 * time.Hour)},  // 1 day ago
			{TransitionedAt: now.Add(-2 * 24 * time.Hour)},  // 2 days ago
			{TransitionedAt: now.Add(-3 * 24 * time.Hour)},  // 3 days ago
			{TransitionedAt: now.Add(-10 * 24 * time.Hour)}, // 10 days ago
			{TransitionedAt: now.Add(-20 * 24 * time.Hour)}, // 20 days ago
		},
	}
	m.metricsReader = mockReader

	result := m.getProjectActivity("test-project", 7)

	// Should return counts for 7 buckets
	if len(result) != 7 {
		t.Fatalf("Expected 7 buckets, got %d", len(result))
	}

	// Total count should equal number of transitions
	total := 0
	for _, c := range result {
		total += c
	}
	if total != 5 {
		t.Errorf("Expected total of 5 transitions, got %d", total)
	}
}

// TestGetProjectActivity_EmptyTransitions verifies behavior when reader returns empty slice.
func TestGetProjectActivity_EmptyTransitions(t *testing.T) {
	m := NewModel(nil)

	// Create mock reader with no transitions
	mockReader := &mockMetricsReader{
		transitions: []statsview.Transition{},
	}
	m.metricsReader = mockReader

	result := m.getProjectActivity("test-project", 7)

	// Should return nil when no transitions (graceful degradation)
	if result != nil {
		t.Errorf("Expected nil when no transitions, got %v", result)
	}
}
