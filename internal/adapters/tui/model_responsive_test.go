package tui

import (
	"strings"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 3.10: Responsive Layout Tests
// ============================================================================

// TestModel_NarrowWidth_ShowsWarning tests AC2
func TestModel_NarrowWidth_ShowsWarning(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 70 // In narrow range (60-79)
	m.height = 30
	m.ready = true // REQUIRED for View() to render dashboard
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 70, 28)
	m.statusBar = components.NewStatusBarModel(70)

	view := m.View()

	if !strings.Contains(view, NarrowWarning()) {
		t.Error("expected narrow warning to be shown for width 70")
	}
}

func TestModel_NarrowWidth_NotShownAt80(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 80 // Not in narrow range
	m.height = 30
	m.ready = true
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 28)
	m.statusBar = components.NewStatusBarModel(80)

	view := m.View()

	if strings.Contains(view, NarrowWarning()) {
		t.Error("expected no narrow warning at width 80")
	}
}

func TestModel_WideTerminal_ContentCapped(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 150 // Wide terminal
	m.height = 30
	m.ready = true
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 120, 28)
	m.statusBar = components.NewStatusBarModel(120)

	// Trigger resize to update component widths
	m.hasPendingResize = true
	m.pendingWidth = 150
	m.pendingHeight = 30
	newModel, _ := m.Update(resizeTickMsg{})
	m = newModel.(Model)

	// View should render without panic
	view := m.View()
	if len(view) == 0 {
		t.Error("expected non-empty view")
	}
}

func TestModel_ShortTerminal_CondensedStatusBar(t *testing.T) {
	repo := &favoriteMockRepository{}
	m := NewModel(repo)
	m.width = 80
	m.height = 18 // Short terminal (< 20)
	m.ready = true

	// Trigger resize to set condensed mode
	m.hasPendingResize = true
	m.pendingHeight = 18
	m.pendingWidth = 80
	newModel, _ := m.Update(resizeTickMsg{})
	m = newModel.(Model)

	view := m.statusBar.View()

	// Condensed view should be single line (no newline)
	if strings.Count(view, "\n") > 0 {
		t.Error("expected single-line condensed status bar")
	}
}

func TestModel_MediumHeight_ShowsDetailHint(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 80
	m.height = 25 // Medium height (20-34)
	m.ready = true
	m.showDetailPanel = false
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 23)
	m.statusBar = components.NewStatusBarModel(80)

	view := m.View()

	if !strings.Contains(view, "Press [d] for details") {
		t.Error("expected detail hint for height 25 with panel closed")
	}
}

func TestModel_MediumHeight_NoHintWhenPanelOpen(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 80
	m.height = 25
	m.ready = true
	m.showDetailPanel = true
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 23)
	m.statusBar = components.NewStatusBarModel(80)
	m.detailPanel = components.NewDetailPanelModel(80, 23)
	m.detailPanel.SetProject(m.projectList.SelectedProject())
	m.detailPanel.SetVisible(true)

	view := m.View()

	if strings.Contains(view, "Press [d] for details") {
		t.Error("expected no hint when detail panel is open")
	}
}

func TestModel_TallTerminal_NoDetailHint(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 80
	m.height = 40 // Tall terminal (>= 35)
	m.ready = true
	m.showDetailPanel = false // Even if manually closed
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 38)
	m.statusBar = components.NewStatusBarModel(80)

	view := m.View()

	// For tall terminals, hint should NOT be shown even if panel is closed
	// because user has enough space and chose to close it
	if strings.Contains(view, "Press [d] for details") {
		t.Error("expected no hint for tall terminal (height >= 35)")
	}
}

func TestIsNarrowWidth(t *testing.T) {
	tests := []struct {
		width    int
		expected bool
	}{
		{59, false},  // Below minimum
		{60, true},   // Start of narrow range
		{70, true},   // Middle of narrow range
		{79, true},   // End of narrow range
		{80, false},  // Standard width
		{120, false}, // Wide
	}

	for _, tt := range tests {
		result := isNarrowWidth(tt.width)
		if result != tt.expected {
			t.Errorf("isNarrowWidth(%d) = %v, want %v", tt.width, result, tt.expected)
		}
	}
}

// ============================================================================
// Additional boundary tests
// ============================================================================

func TestModel_NarrowWidth_BoundaryAt60(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 60 // Exactly at narrow start
	m.height = 30
	m.ready = true
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 60, 28)
	m.statusBar = components.NewStatusBarModel(60)

	view := m.View()

	if !strings.Contains(view, NarrowWarning()) {
		t.Error("expected narrow warning at width 60 (boundary)")
	}
}

func TestModel_NarrowWidth_BoundaryAt79(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 79 // Last narrow width
	m.height = 30
	m.ready = true
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 79, 28)
	m.statusBar = components.NewStatusBarModel(79)

	view := m.View()

	if !strings.Contains(view, NarrowWarning()) {
		t.Error("expected narrow warning at width 79 (boundary)")
	}
}

func TestModel_MediumHeight_BoundaryAt20(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 80
	m.height = 20 // Exactly at medium height start
	m.ready = true
	m.showDetailPanel = false
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 18)
	m.statusBar = components.NewStatusBarModel(80)

	view := m.View()

	if !strings.Contains(view, "Press [d] for details") {
		t.Error("expected detail hint at height 20 (boundary)")
	}
}

func TestModel_MediumHeight_BoundaryAt34(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 80
	m.height = 34 // Last medium height
	m.ready = true
	m.showDetailPanel = false
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 32)
	m.statusBar = components.NewStatusBarModel(80)

	view := m.View()

	if !strings.Contains(view, "Press [d] for details") {
		t.Error("expected detail hint at height 34 (boundary)")
	}
}

func TestModel_TallTerminal_BoundaryAt35(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 80
	m.height = 35 // First tall height
	m.ready = true
	m.showDetailPanel = false
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 33)
	m.statusBar = components.NewStatusBarModel(80)

	view := m.View()

	if strings.Contains(view, "Press [d] for details") {
		t.Error("expected no hint at height 35 (tall boundary)")
	}
}

func TestModel_ShortTerminal_BoundaryAt19(t *testing.T) {
	repo := &favoriteMockRepository{}
	m := NewModel(repo)
	m.width = 80
	m.height = 19 // Last short height
	m.ready = true

	// Trigger resize
	m.hasPendingResize = true
	m.pendingHeight = 19
	m.pendingWidth = 80
	newModel, _ := m.Update(resizeTickMsg{})
	m = newModel.(Model)

	view := m.statusBar.View()

	// Should still be condensed at 19
	if strings.Count(view, "\n") > 0 {
		t.Error("expected condensed status bar at height 19")
	}
}

func TestModel_NormalHeight_StatusBarNotCondensed(t *testing.T) {
	repo := &favoriteMockRepository{}
	m := NewModel(repo)
	m.width = 80
	m.height = 20 // First normal height
	m.ready = true

	// Trigger resize
	m.hasPendingResize = true
	m.pendingHeight = 20
	m.pendingWidth = 80
	newModel, _ := m.Update(resizeTickMsg{})
	m = newModel.(Model)

	view := m.statusBar.View()

	// Should NOT be condensed at 20
	if strings.Count(view, "\n") != 1 {
		t.Errorf("expected normal (2-line) status bar at height 20, got lines: %d", strings.Count(view, "\n")+1)
	}
}

// Code review fix: Test hint suppression at height < 20 (short terminal)
func TestModel_ShortTerminal_NoDetailHint(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.width = 80
	m.height = 19 // Short terminal (< 20) - hint should NOT show
	m.ready = true
	m.showDetailPanel = false
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 80, 18)
	m.statusBar = components.NewStatusBarModel(80)
	m.statusBar.SetCondensed(true) // Would be set by resizeTickMsg

	view := m.View()

	// For short terminals (< 20), hint should NOT be shown because
	// the screen is too small to display hints - space is at a premium
	if strings.Contains(view, "Press [d] for details") {
		t.Error("expected no detail hint for short terminal (height < 20)")
	}
}

// Code review fix: Test helper functions
// Story 8.10: isWideWidth is now a method on Model using config-based maxContentWidth
func TestIsWideWidth(t *testing.T) {
	tests := []struct {
		width           int
		maxContentWidth int
		expected        bool
	}{
		{100, 120, false}, // Standard width, default max
		{120, 120, false}, // At max boundary
		{121, 120, true},  // Just over max
		{150, 120, true},  // Wide
		{200, 120, true},  // Very wide
		{200, 0, false},   // Unlimited (0) - always returns false
		{100, 80, true},   // Custom max (80), width > max
		{80, 80, false},   // Custom max (80), width = max
	}

	for _, tt := range tests {
		m := NewModel(nil)
		m.width = tt.width
		m.maxContentWidth = tt.maxContentWidth
		result := m.isWideWidth()
		if result != tt.expected {
			t.Errorf("isWideWidth() with width=%d, maxContentWidth=%d = %v, want %v",
				tt.width, tt.maxContentWidth, result, tt.expected)
		}
	}
}

func TestStatusBarHeight(t *testing.T) {
	tests := []struct {
		height   int
		expected int
	}{
		{15, 1}, // Short - condensed
		{19, 1}, // Short boundary
		{20, 2}, // Normal boundary
		{30, 2}, // Normal
		{50, 2}, // Tall
	}

	for _, tt := range tests {
		result := statusBarHeight(tt.height)
		if result != tt.expected {
			t.Errorf("statusBarHeight(%d) = %v, want %v", tt.height, result, tt.expected)
		}
	}
}

// Story 8.10: Test max_content_width=0 (unlimited/full-width mode)
func TestModel_FullWidthMode_MaxContentWidthZero(t *testing.T) {
	repo := &favoriteMockRepository{}
	repo.projects = []*domain.Project{{ID: "1", Path: "/test", Name: "test-project"}}

	m := NewModel(repo)
	m.maxContentWidth = 0 // Unlimited mode
	m.width = 200         // Very wide terminal
	m.height = 40
	m.ready = true
	m.projects = repo.projects
	m.projectList = components.NewProjectListModel(repo.projects, 200, 38)
	m.statusBar = components.NewStatusBarModel(200)

	// isWideWidth should return false for unlimited mode
	if m.isWideWidth() {
		t.Error("isWideWidth() should return false when maxContentWidth=0 (unlimited)")
	}

	// Render should use full width (no centering)
	view := m.View()

	// View should not be centered (no lipgloss.Place padding)
	// This is a basic check - the content should start at the left edge
	lines := strings.Split(view, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		// In unlimited mode, first character should not be spaces from centering
		// (unless it's part of the content itself)
		if len(firstLine) > 0 && strings.HasPrefix(firstLine, "                    ") {
			t.Error("in unlimited mode, content should not be centered with padding")
		}
	}
}
