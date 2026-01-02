package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 11.5: CLI Activate Command Tests
// ============================================================================

// NOTE: mockStateService and hibernateMockRepository are defined in hibernate_test.go
// (same cli_test package, so they're accessible here)

// executeActivateCommand runs the activate command with given args and returns output/error.
// Follows favorite_test.go pattern.
func executeActivateCommand(args []string) (string, error) {
	cmd := cli.NewRootCmd()
	cli.RegisterActivateCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"activate"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// TestActivateCmd_HibernatedProject_Succeeds verifies activating hibernated project (AC2).
func TestActivateCmd_HibernatedProject_Succeeds(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateHibernated},
	}
	mockRepo := newHibernateMockRepository().withProjects(projects)
	mockState := &mockStateService{}

	cli.SetRepository(mockRepo)
	cli.SetStateService(mockState)
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute
	output, err := executeActivateCommand([]string{"test-project"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mockState.activateCalled {
		t.Error("expected Activate to be called")
	}
	if !strings.Contains(output, "✓ Activated") {
		t.Errorf("expected '✓ Activated', got: %s", output)
	}
}

// TestActivateCmd_AlreadyActive_Idempotent verifies idempotent behavior (AC4).
func TestActivateCmd_AlreadyActive_Idempotent(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateActive},
	}
	mockRepo := newHibernateMockRepository().withProjects(projects)
	mockState := &mockStateService{activateErr: domain.ErrInvalidStateTransition}

	cli.SetRepository(mockRepo)
	cli.SetStateService(mockState)
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute
	output, err := executeActivateCommand([]string{"test-project"})

	// Assert - should succeed (idempotent)
	if err != nil {
		t.Fatalf("expected no error (idempotent), got: %v", err)
	}
	if !strings.Contains(output, "already active") {
		t.Errorf("expected 'already active' message, got: %s", output)
	}
}

// TestActivateCmd_ProjectNotFound_ReturnsError verifies not found handling (AC6).
func TestActivateCmd_ProjectNotFound_ReturnsError(t *testing.T) {
	// Setup - empty repository
	cli.SetRepository(newHibernateMockRepository())
	cli.SetStateService(&mockStateService{})
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute
	output, err := executeActivateCommand([]string{"nonexistent"})

	// Assert
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
	if !strings.Contains(output, "Project not found") {
		t.Errorf("expected 'Project not found' message, got: %s", output)
	}
}

// TestActivateCmd_QuietMode_SuppressesOutput verifies quiet mode (AC7).
func TestActivateCmd_QuietMode_SuppressesOutput(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateHibernated},
	}
	mockRepo := newHibernateMockRepository().withProjects(projects)
	mockState := &mockStateService{}

	cli.SetRepository(mockRepo)
	cli.SetStateService(mockState)
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
		cli.ResetQuietFlag()
	}()

	// Execute
	cmd := cli.NewRootCmd()
	cli.RegisterActivateCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"activate", "test-project"})

	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output with --quiet, got: %s", buf.String())
	}
}

// TestActivateCmd_ByDisplayName_Succeeds verifies display name lookup (AC8).
func TestActivateCmd_ByDisplayName_Succeeds(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", DisplayName: "My Cool App", State: domain.StateHibernated},
	}
	mockRepo := newHibernateMockRepository().withProjects(projects)
	mockState := &mockStateService{}

	cli.SetRepository(mockRepo)
	cli.SetStateService(mockState)
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute using display name
	output, err := executeActivateCommand([]string{"My Cool App"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mockState.activateCalled {
		t.Error("expected Activate to be called via display name lookup")
	}
	if !strings.Contains(output, "✓ Activated") {
		t.Errorf("expected success message, got: %s", output)
	}
}

// TestActivateCmd_StateServiceNil_ReturnsError verifies nil stateService handling.
func TestActivateCmd_StateServiceNil_ReturnsError(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateHibernated},
	}
	cli.SetRepository(newHibernateMockRepository().withProjects(projects))
	cli.SetStateService(nil) // Explicitly nil
	defer func() {
		cli.SetRepository(nil)
	}()

	// Execute
	_, err := executeActivateCommand([]string{"test-project"})

	// Assert
	if err == nil {
		t.Fatal("expected error when stateService is nil")
	}
}

// TestActivateCmd_FindByPath verifies path-based lookup (AC8).
func TestActivateCmd_FindByPath(t *testing.T) {
	// Setup: Create temp directory and get its canonical path
	tempDir := t.TempDir()
	canonicalDir, err := filesystem.CanonicalPath(tempDir)
	if err != nil {
		t.Fatalf("failed to get canonical path: %v", err)
	}

	projects := []*domain.Project{
		{ID: "1", Path: canonicalDir, Name: "my-project", State: domain.StateHibernated},
	}
	mockRepo := newHibernateMockRepository().withProjects(projects)
	mockState := &mockStateService{}

	cli.SetRepository(mockRepo)
	cli.SetStateService(mockState)
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute using path
	output, err := executeActivateCommand([]string{tempDir})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mockState.activateCalled {
		t.Error("expected Activate to be called via path lookup")
	}
	if !strings.Contains(output, "✓ Activated") {
		t.Errorf("expected success message, got: %s", output)
	}
}

// TestActivateCmd_NoArgs_ReturnsError verifies error when no arguments provided.
func TestActivateCmd_NoArgs_ReturnsError(t *testing.T) {
	// Setup
	cli.SetRepository(newHibernateMockRepository())
	cli.SetStateService(&mockStateService{})
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute with no args
	_, err := executeActivateCommand([]string{})

	// Assert - cobra should return error for missing required arg
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}

// TestActivateCmd_QuietMode_AlreadyActive_SuppressesOutput verifies quiet mode with idempotent case.
func TestActivateCmd_QuietMode_AlreadyActive_SuppressesOutput(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateActive},
	}
	mockRepo := newHibernateMockRepository().withProjects(projects)
	mockState := &mockStateService{activateErr: domain.ErrInvalidStateTransition}

	cli.SetRepository(mockRepo)
	cli.SetStateService(mockState)
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
		cli.ResetQuietFlag()
	}()

	// Execute
	cmd := cli.NewRootCmd()
	cli.RegisterActivateCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"activate", "test-project"})

	err := cmd.Execute()

	// Assert - should succeed (idempotent) with no output
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output with --quiet (idempotent), got: %s", buf.String())
	}
}
