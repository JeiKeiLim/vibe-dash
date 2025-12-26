package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/project"
)

// ProjectListModel wraps a Bubbles list for displaying projects.
type ProjectListModel struct {
	list     list.Model
	projects []*domain.Project
	width    int
	height   int
	delegate ProjectItemDelegate
}

// NewProjectListModel creates a new ProjectListModel with the given projects and dimensions.
func NewProjectListModel(projects []*domain.Project, width, height int) ProjectListModel {
	// Sort projects alphabetically by effective name
	sortedProjects := make([]*domain.Project, len(projects))
	copy(sortedProjects, projects)
	project.SortByName(sortedProjects)

	// Convert to ProjectItem slice
	items := make([]list.Item, len(sortedProjects))
	for i, p := range sortedProjects {
		items[i] = ProjectItem{Project: p}
	}

	// Create custom delegate
	delegate := NewProjectItemDelegate(width)

	// Initialize Bubbles list
	l := list.New(items, delegate, width, height)
	l.SetShowHelp(false)         // We have our own help (Story 3.5)
	l.SetShowTitle(false)        // Title in our own header
	l.SetShowStatusBar(true)     // Shows "1/5" pagination
	l.SetFilteringEnabled(false) // For now, enable in post-MVP

	// Configure keymap
	l.KeyMap = list.DefaultKeyMap()
	l.KeyMap.ForceQuit.Unbind() // We handle quit ourselves

	// Select first item if available
	if len(items) > 0 {
		l.Select(0)
	}

	return ProjectListModel{
		list:     l,
		projects: sortedProjects,
		width:    width,
		height:   height,
		delegate: delegate,
	}
}

// SetProjects updates the list with new projects.
func (m *ProjectListModel) SetProjects(projects []*domain.Project) {
	// Sort projects alphabetically by effective name
	sortedProjects := make([]*domain.Project, len(projects))
	copy(sortedProjects, projects)
	project.SortByName(sortedProjects)

	// Convert to ProjectItem slice
	items := make([]list.Item, len(sortedProjects))
	for i, p := range sortedProjects {
		items[i] = ProjectItem{Project: p}
	}

	m.projects = sortedProjects
	m.list.SetItems(items)

	// Handle selection after list update (Story 3.9 AC6 fix)
	// Case 1: Index is out of bounds (e.g., removed last item) - select last valid item
	if len(items) > 0 && m.list.Index() >= len(items) {
		m.list.Select(len(items) - 1)
	}
	// Case 2: Nothing selected (Index < 0) - select first item
	if len(items) > 0 && m.list.Index() < 0 {
		m.list.Select(0)
	}
}

// SetSize updates the list dimensions for responsive layout.
func (m *ProjectListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
	m.delegate.SetWidth(width)
	// Update the delegate in the list to reflect new width
	m.list.SetDelegate(m.delegate)
}

// View renders the list to a string.
func (m ProjectListModel) View() string {
	return m.list.View()
}

// Update handles messages and returns updated model with commands.
func (m ProjectListModel) Update(msg tea.Msg) (ProjectListModel, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// SelectedProject returns the currently selected project, or nil if none.
func (m ProjectListModel) SelectedProject() *domain.Project {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	projectItem, ok := item.(ProjectItem)
	if !ok {
		return nil
	}
	return projectItem.Project
}

// HasProjects returns true if the list contains any projects.
func (m ProjectListModel) HasProjects() bool {
	return len(m.projects) > 0
}

// Index returns the current selection index.
func (m ProjectListModel) Index() int {
	return m.list.Index()
}

// Len returns the number of items in the list.
func (m ProjectListModel) Len() int {
	return len(m.projects)
}

// SetDelegateWaitingCallbacks sets the waiting detection callbacks on the delegate.
// Story 4.5: Enables WAITING indicator display in project list rows.
func (m *ProjectListModel) SetDelegateWaitingCallbacks(checker WaitingChecker, getter WaitingDurationGetter) {
	m.delegate.SetWaitingCallbacks(checker, getter)
	m.list.SetDelegate(m.delegate)
}

// Width returns the current width of the project list.
// Story 8.4: Used to detect zero-value component (uninitialized).
func (m ProjectListModel) Width() int {
	return m.width
}
