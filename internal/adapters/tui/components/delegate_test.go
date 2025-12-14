package components

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestProjectItemDelegate_Height(t *testing.T) {
	delegate := NewProjectItemDelegate(80)
	if got := delegate.Height(); got != 1 {
		t.Errorf("Height() = %d, want 1", got)
	}
}

func TestProjectItemDelegate_Spacing(t *testing.T) {
	delegate := NewProjectItemDelegate(80)
	if got := delegate.Spacing(); got != 0 {
		t.Errorf("Spacing() = %d, want 0", got)
	}
}

func TestProjectItemDelegate_Update(t *testing.T) {
	delegate := NewProjectItemDelegate(80)
	cmd := delegate.Update(nil, nil)
	if cmd != nil {
		t.Error("Update() should return nil")
	}
}

func TestProjectItemDelegate_Render_Selected(t *testing.T) {
	now := time.Now()
	delegate := NewProjectItemDelegate(80)

	project := &domain.Project{
		Name:           "test-project",
		CurrentStage:   domain.StageImplement,
		LastActivityAt: now.Add(-1 * time.Hour),
	}
	item := ProjectItem{Project: project}

	// Create a list with single item - it will be selected by default
	items := []list.Item{item}
	l := list.New(items, delegate, 80, 10)

	var buf bytes.Buffer
	delegate.Render(&buf, l, 0, item)
	output := buf.String()

	// Index 0 should be selected (the only item)
	if !strings.HasPrefix(output, "> ") {
		t.Errorf("Render() selected item prefix = %q, want prefix %q", output[:2], "> ")
	}
}

func TestProjectItemDelegate_Render_Unselected(t *testing.T) {
	now := time.Now()
	delegate := NewProjectItemDelegate(80)

	project1 := &domain.Project{
		Name:           "first-project",
		CurrentStage:   domain.StageImplement,
		LastActivityAt: now.Add(-1 * time.Hour),
	}
	project2 := &domain.Project{
		Name:           "second-project",
		CurrentStage:   domain.StagePlan,
		LastActivityAt: now.Add(-2 * 24 * time.Hour),
	}
	item1 := ProjectItem{Project: project1}
	item2 := ProjectItem{Project: project2}

	// Create a list with two items, select the first one
	items := []list.Item{item1, item2}
	l := list.New(items, delegate, 80, 10)
	l.Select(0) // Select first item

	// Render the second item (index 1), which is unselected
	var buf bytes.Buffer
	delegate.Render(&buf, l, 1, item2)
	output := buf.String()

	// Index 1 should NOT be selected
	if !strings.HasPrefix(output, "  ") {
		t.Errorf("Render() unselected item prefix = %q, want prefix %q", output[:2], "  ")
	}
}

func TestProjectItemDelegate_Render_WithDisplayName(t *testing.T) {
	now := time.Now()
	delegate := NewProjectItemDelegate(80)

	project := &domain.Project{
		Name:           "original-name",
		DisplayName:    "Custom Display",
		CurrentStage:   domain.StageSpecify,
		LastActivityAt: now.Add(-30 * time.Minute),
	}
	item := ProjectItem{Project: project}

	items := []list.Item{item}
	l := list.New(items, delegate, 80, 10)

	var buf bytes.Buffer
	delegate.Render(&buf, l, 0, item)
	output := buf.String()

	// Should contain display name, not original name
	if !strings.Contains(output, "Custom Display") {
		t.Errorf("Render() should contain display name 'Custom Display', got: %q", output)
	}
}

func TestProjectItemDelegate_RenderContainsProjectName(t *testing.T) {
	project := &domain.Project{
		Name:           "my-test-project",
		CurrentStage:   domain.StageImplement,
		LastActivityAt: time.Now(),
	}

	delegate := NewProjectItemDelegate(80)
	item := ProjectItem{Project: project}

	items := []list.Item{item}
	l := list.New(items, delegate, 80, 10)

	var buf bytes.Buffer
	delegate.Render(&buf, l, 0, item)
	output := buf.String()

	if !strings.Contains(output, "my-test-project") {
		t.Errorf("Render() output should contain project name, got: %q", output)
	}
}

func TestProjectItemDelegate_RenderContainsStageName(t *testing.T) {
	project := &domain.Project{
		Name:           "project",
		CurrentStage:   domain.StagePlan,
		LastActivityAt: time.Now(),
	}

	delegate := NewProjectItemDelegate(80)
	item := ProjectItem{Project: project}

	items := []list.Item{item}
	l := list.New(items, delegate, 80, 10)

	var buf bytes.Buffer
	delegate.Render(&buf, l, 0, item)
	output := buf.String()

	if !strings.Contains(output, "Plan") {
		t.Errorf("Render() output should contain stage name 'Plan', got: %q", output)
	}
}

func TestProjectItemDelegate_SetWidth(t *testing.T) {
	delegate := NewProjectItemDelegate(80)
	delegate.SetWidth(120)

	// Verify width changed (indirectly through calculateNameWidth)
	nameWidth := delegate.calculateNameWidth()
	if nameWidth <= 0 {
		t.Error("SetWidth should update width for name calculation")
	}
}

func TestProjectItemDelegate_CalculateNameWidth(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		wantAtMin bool // Should result in minimum width
	}{
		{"normal width", 80, false},
		{"large width", 120, false},
		{"small width", 40, true},
		{"very small width", 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delegate := NewProjectItemDelegate(tt.width)
			nameWidth := delegate.calculateNameWidth()

			if tt.wantAtMin && nameWidth != colNameMin {
				t.Errorf("calculateNameWidth() = %d, want minimum %d", nameWidth, colNameMin)
			}
			if !tt.wantAtMin && nameWidth <= colNameMin {
				t.Errorf("calculateNameWidth() = %d, should be greater than minimum %d", nameWidth, colNameMin)
			}
		})
	}
}

func TestProjectItemDelegate_Render_RecencyIndicator_Today(t *testing.T) {
	// Activity within 24 hours should show ✨
	now := time.Now()
	delegate := NewProjectItemDelegate(80)

	project := &domain.Project{
		Name:           "recent-project",
		CurrentStage:   domain.StageImplement,
		LastActivityAt: now.Add(-1 * time.Hour), // 1 hour ago
	}
	item := ProjectItem{Project: project}

	items := []list.Item{item}
	l := list.New(items, delegate, 80, 10)

	var buf bytes.Buffer
	delegate.Render(&buf, l, 0, item)
	output := buf.String()

	if !strings.Contains(output, "✨") {
		t.Errorf("Render() should contain ✨ for activity within 24 hours, got: %q", output)
	}
}

func TestProjectItemDelegate_Render_RecencyIndicator_ThisWeek(t *testing.T) {
	// Activity within 7 days but > 24 hours should show ⚡
	now := time.Now()
	delegate := NewProjectItemDelegate(80)

	project := &domain.Project{
		Name:           "week-project",
		CurrentStage:   domain.StagePlan,
		LastActivityAt: now.Add(-3 * 24 * time.Hour), // 3 days ago
	}
	item := ProjectItem{Project: project}

	items := []list.Item{item}
	l := list.New(items, delegate, 80, 10)

	var buf bytes.Buffer
	delegate.Render(&buf, l, 0, item)
	output := buf.String()

	if !strings.Contains(output, "⚡") {
		t.Errorf("Render() should contain ⚡ for activity within 7 days, got: %q", output)
	}
}

func TestProjectItemDelegate_Render_RecencyIndicator_Old(t *testing.T) {
	// Activity older than 7 days should NOT show ✨ or ⚡
	now := time.Now()
	delegate := NewProjectItemDelegate(80)

	project := &domain.Project{
		Name:           "old-project",
		CurrentStage:   domain.StageSpecify,
		LastActivityAt: now.Add(-30 * 24 * time.Hour), // 30 days ago
	}
	item := ProjectItem{Project: project}

	items := []list.Item{item}
	l := list.New(items, delegate, 80, 10)

	var buf bytes.Buffer
	delegate.Render(&buf, l, 0, item)
	output := buf.String()

	if strings.Contains(output, "✨") {
		t.Errorf("Render() should NOT contain ✨ for old activity, got: %q", output)
	}
	if strings.Contains(output, "⚡") {
		t.Errorf("Render() should NOT contain ⚡ for old activity, got: %q", output)
	}
}

func TestProjectItemDelegate_Render_RecencyIndicator_ZeroTime(t *testing.T) {
	// Zero time should NOT show any indicator
	delegate := NewProjectItemDelegate(80)

	project := &domain.Project{
		Name:           "zero-time-project",
		CurrentStage:   domain.StageImplement,
		LastActivityAt: time.Time{}, // Zero time
	}
	item := ProjectItem{Project: project}

	items := []list.Item{item}
	l := list.New(items, delegate, 80, 10)

	var buf bytes.Buffer
	delegate.Render(&buf, l, 0, item)
	output := buf.String()

	if strings.Contains(output, "✨") {
		t.Errorf("Render() should NOT contain ✨ for zero time, got: %q", output)
	}
	if strings.Contains(output, "⚡") {
		t.Errorf("Render() should NOT contain ⚡ for zero time, got: %q", output)
	}
}
