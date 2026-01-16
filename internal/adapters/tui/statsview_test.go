package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
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

// TestRenderStatsView_PlaceholderContent verifies Stats View shows placeholder (AC4).
func TestRenderStatsView_PlaceholderContent(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 100
	m.height = 30
	m.statusBar = components.NewStatusBarModel(100)

	m.enterStatsView()
	output := m.renderStatsView()

	// Verify placeholder content
	if !strings.Contains(output, "Project metrics will appear here") {
		t.Error("Stats View should display placeholder content")
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
