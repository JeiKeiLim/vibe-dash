package tui

import "github.com/JeiKeiLim/vibe-dash/internal/core/domain"

// Test helper exports - only available in test files

// ValidationCompleteMsgForTest creates a validationCompleteMsg for testing
func ValidationCompleteMsgForTest(invalid []InvalidProject) validationCompleteMsg {
	return validationCompleteMsg{invalidProjects: invalid}
}

// IsValidationMode returns true if the model is in validation view mode
func IsValidationMode(m Model) bool {
	return m.viewMode == viewModeValidation
}

// RenderValidationDialogForTest exposes renderValidationDialog for testing
func RenderValidationDialogForTest(project *domain.Project, width, height int) string {
	return renderValidationDialog(project, width, height, "")
}

// RenderValidationDialogWithErrorForTest exposes renderValidationDialog with error for testing
func RenderValidationDialogWithErrorForTest(project *domain.Project, width, height int, errorMsg string) string {
	return renderValidationDialog(project, width, height, errorMsg)
}

// GetValidationError returns the current validation error from the model
func GetValidationError(m Model) string {
	return m.validationError
}

// EffectiveNameForTest exposes effectiveName for testing
func EffectiveNameForTest(p *domain.Project) string {
	return effectiveName(p)
}

// Story 16.3: Stats view test helpers

// IsStatsMode returns true if the model is in stats view mode
func IsStatsMode(m Model) bool {
	return m.viewMode == viewModeStats
}

// GetStatsActiveProjectIdx returns the saved project index for restoration
func GetStatsActiveProjectIdx(m Model) int {
	return m.statsActiveProjectIdx
}

// SetViewModeStatsForTest sets the model to stats view mode for testing
func SetViewModeStatsForTest(m *Model) {
	m.viewMode = viewModeStats
}

// EnterStatsViewForTest exposes enterStatsView for testing
func EnterStatsViewForTest(m *Model) {
	m.enterStatsView()
}

// ExitStatsViewForTest exposes exitStatsView for testing
func ExitStatsViewForTest(m *Model) {
	m.exitStatsView()
}

// GetProjectListIndexForTest returns the current project list index
func GetProjectListIndexForTest(m Model) int {
	return m.projectList.Index()
}
