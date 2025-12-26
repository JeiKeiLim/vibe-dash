package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// executeRemoveCommand helper - matches executeAddCommand pattern
func executeRemoveCommand(t *testing.T, args []string, stdin string) (string, error) {
	t.Helper()
	cli.ResetRemoveFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterRemoveCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if stdin != "" {
		cmd.SetIn(strings.NewReader(stdin))
	} else {
		cmd.SetIn(strings.NewReader("")) // Empty stdin for EOF simulation
	}

	fullArgs := append([]string{"remove"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// ============================================================================
// Task 1: Basic remove command tests (AC: 1, 5, 6)
// ============================================================================

func TestRemove_ProjectNotFound_ExitCode2(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	output, err := executeRemoveCommand(t, []string{"nonexistent"}, "")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}

	// Verify error message format (AC6): "✗ Project not found: nonexistent"
	if !strings.Contains(output, "✗ Project not found: nonexistent") {
		t.Errorf("expected formatted error message with ✗ prefix, got: %s", output)
	}

	// Verify underlying error for exit code mapping
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}

	// Verify exit code 2 for project not found
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitProjectNotFound {
		t.Errorf("expected exit code %d, got %d", cli.ExitProjectNotFound, exitCode)
	}
}

func TestRemove_ByDisplayName(t *testing.T) {
	// AC5: Project found by display_name
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/client-alpha", "")
	p.DisplayName = "My Alpha Project"
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	output, err := executeRemoveCommand(t, []string{"My Alpha Project", "--force"}, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(output, "✓ Removed: My Alpha Project") {
		t.Errorf("expected success with display name, got: %s", output)
	}

	// Verify project was deleted
	if len(mock.Projects) != 0 {
		t.Error("expected project to be deleted from repository")
	}
}

func TestRemove_ByName(t *testing.T) {
	// Basic removal by Name field
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/client-alpha", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	output, err := executeRemoveCommand(t, []string{"client-alpha", "--force"}, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(output, "✓ Removed: client-alpha") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify project was deleted
	if len(mock.Projects) != 0 {
		t.Error("expected project to be deleted from repository")
	}
}

// ============================================================================
// Task 2: Confirmation prompt tests (AC: 1, 2, 3)
// ============================================================================

func TestRemove_ConfirmY_ProjectRemoved(t *testing.T) {
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/client-alpha", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	output, err := executeRemoveCommand(t, []string{"client-alpha"}, "y\n")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// AC1: Verify prompt format
	if !strings.Contains(output, "Remove 'client-alpha' from tracking? [y/n]") {
		t.Errorf("expected confirmation prompt, got: %s", output)
	}

	// AC2: Verify success message
	if !strings.Contains(output, "✓ Removed: client-alpha") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify project was deleted
	if len(mock.Projects) != 0 {
		t.Error("expected project to be deleted from repository")
	}
}

func TestRemove_ConfirmN_ProjectKept(t *testing.T) {
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/client-alpha", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	output, err := executeRemoveCommand(t, []string{"client-alpha"}, "n\n")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// AC3: Verify cancelled message
	if !strings.Contains(output, "Cancelled") {
		t.Errorf("expected cancelled message, got: %s", output)
	}

	// Verify project was NOT deleted
	if len(mock.Projects) != 1 {
		t.Error("expected project to remain in repository")
	}
}

func TestRemove_CaseInsensitiveConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool // true = removed, false = cancelled
	}{
		{"lowercase y", "y\n", true},
		{"uppercase Y", "Y\n", true},
		{"yes", "yes\n", true},
		{"YES", "YES\n", true},
		{"lowercase n", "n\n", false},
		{"uppercase N", "N\n", false},
		{"no", "no\n", false},
		{"NO", "NO\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRepository()
			p, _ := domain.NewProject("/path/to/test", "")
			mock.Projects[p.Path] = p
			cli.SetRepository(mock)

			output, err := executeRemoveCommand(t, []string{"test"}, tt.input)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if tt.expected {
				if !strings.Contains(output, "✓ Removed") {
					t.Errorf("expected removal, got: %s", output)
				}
				if len(mock.Projects) != 0 {
					t.Error("expected project to be deleted")
				}
			} else {
				if !strings.Contains(output, "Cancelled") {
					t.Errorf("expected cancellation, got: %s", output)
				}
				if len(mock.Projects) != 1 {
					t.Error("expected project to remain")
				}
			}
		})
	}
}

// ============================================================================
// Task 3: --force flag tests (AC: 4)
// ============================================================================

func TestRemove_ForceFlag_NoPrompt(t *testing.T) {
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/client-alpha", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	output, err := executeRemoveCommand(t, []string{"client-alpha", "--force"}, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// AC4: Verify success message
	if !strings.Contains(output, "✓ Removed: client-alpha") {
		t.Errorf("expected success message, got: %s", output)
	}

	// AC4: Verify no confirmation prompt was shown
	if strings.Contains(output, "[y/n]") {
		t.Error("expected no confirmation prompt with --force flag")
	}

	// Verify project was deleted
	if len(mock.Projects) != 0 {
		t.Error("expected project to be deleted from repository")
	}
}

// ============================================================================
// Task 4: Edge cases tests
// ============================================================================

func TestRemove_InvalidInput_Reprompts(t *testing.T) {
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	// First invalid, then valid
	output, err := executeRemoveCommand(t, []string{"test"}, "maybe\ny\n")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(output, "Please enter") {
		t.Errorf("expected re-prompt message, got: %s", output)
	}
	if !strings.Contains(output, "✓ Removed") {
		t.Errorf("expected eventual removal, got: %s", output)
	}
}

func TestRemove_MissingArgument(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	_, err := executeRemoveCommand(t, []string{}, "")
	if err == nil {
		t.Fatal("expected error for missing argument")
	}

	// Cobra should complain about missing args
	if !strings.Contains(err.Error(), "requires") && !strings.Contains(err.Error(), "argument") && !strings.Contains(err.Error(), "accepts") {
		t.Errorf("expected argument error, got: %v", err)
	}
}

func TestRemove_EOF_GracefulCancellation(t *testing.T) {
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	// Empty stdin simulates EOF (Ctrl+D)
	output, err := executeRemoveCommand(t, []string{"test"}, "")
	if err != nil {
		t.Fatalf("expected no error for EOF cancellation, got: %v", err)
	}

	if !strings.Contains(output, "Cancelled") {
		t.Errorf("expected 'Cancelled' message on EOF, got: %s", output)
	}

	// Verify project was NOT deleted
	if len(mock.Projects) != 1 {
		t.Error("expected project to remain after EOF cancellation")
	}
}

// ============================================================================
// Additional edge case tests
// ============================================================================

func TestRemove_ShowsDisplayNameInPrompt(t *testing.T) {
	// When project has DisplayName, prompt should show it
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/project", "")
	p.DisplayName = "My Cool Project"
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	output, err := executeRemoveCommand(t, []string{"My Cool Project"}, "n\n")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Prompt should show display name, not directory name
	if !strings.Contains(output, "Remove 'My Cool Project' from tracking?") {
		t.Errorf("expected prompt with DisplayName, got: %s", output)
	}
}

func TestRemove_ExitCodeZeroOnCancellation(t *testing.T) {
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	_, err := executeRemoveCommand(t, []string{"test"}, "n\n")

	// AC3: Exit code should be 0 on cancellation
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitSuccess {
		t.Errorf("expected exit code 0 on cancellation, got %d", exitCode)
	}
}

func TestRemove_ExitCodeZeroOnSuccess(t *testing.T) {
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	_, err := executeRemoveCommand(t, []string{"test", "--force"}, "")

	// AC4: Exit code should be 0 on success
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitSuccess {
		t.Errorf("expected exit code 0 on success, got %d", exitCode)
	}
}

// ============================================================================
// Code Review Fixes - Additional edge case tests
// ============================================================================

func TestRemove_DeleteFailure(t *testing.T) {
	// H2: Test repository.Delete() failure path
	mock := NewMockRepository()
	mock.SetDeleteError(errors.New("database error"))
	p, _ := domain.NewProject("/path/to/test", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	_, err := executeRemoveCommand(t, []string{"test", "--force"}, "")
	if err == nil {
		t.Fatal("expected error when delete fails")
	}

	if !strings.Contains(err.Error(), "failed to remove") {
		t.Errorf("expected wrapped error message, got: %v", err)
	}

	// Project should still exist since delete failed
	if len(mock.Projects) != 1 {
		t.Error("expected project to remain after delete failure")
	}
}

func TestRemove_MultipleInvalidInputs_Reprompts(t *testing.T) {
	// M3: Test multiple consecutive invalid inputs before valid
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	// Three invalid inputs, then valid
	output, err := executeRemoveCommand(t, []string{"test"}, "maybe\nwhat\nhuh\ny\n")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should have multiple re-prompt messages
	repromptCount := strings.Count(output, "Please enter")
	if repromptCount < 3 {
		t.Errorf("expected at least 3 re-prompts, got %d. Output: %s", repromptCount, output)
	}

	if !strings.Contains(output, "✓ Removed") {
		t.Errorf("expected eventual removal, got: %s", output)
	}

	// Verify project was eventually deleted
	if len(mock.Projects) != 0 {
		t.Error("expected project to be deleted after valid input")
	}
}

// ============================================================================
// Task 10: New storage structure tests (AC: 1, 2, 4 from Story 3.5.6)
// ============================================================================

func TestRemove_CallsDeleteProjectDir(t *testing.T) {
	// AC2: Verify DeleteProjectDir is called after repository.Delete()
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test-project", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	// Set up mock directory manager
	mockDM := NewMockDirectoryManager()
	SetMockDirectoryManager(mockDM)
	defer ClearMockDirectoryManager()

	output, err := executeRemoveCommand(t, []string{"test-project", "--force"}, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(output, "✓ Removed: test-project") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify DeleteProjectDir was called with the project path
	if len(mockDM.DeleteCalls()) != 1 {
		t.Errorf("expected 1 DeleteProjectDir call, got %d", len(mockDM.DeleteCalls()))
	} else if mockDM.DeleteCalls()[0] != "/path/to/test-project" {
		t.Errorf("DeleteProjectDir called with wrong path: got %s, want /path/to/test-project", mockDM.DeleteCalls()[0])
	}
}

func TestRemove_DirectoryDeletionErrorIsNonFatal(t *testing.T) {
	// AC2: Verify directory deletion failure doesn't fail the remove command
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test-project", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	// Set up mock directory manager with error
	mockDM := NewMockDirectoryManager()
	mockDM.SetDeleteError(errors.New("permission denied"))
	SetMockDirectoryManager(mockDM)
	defer ClearMockDirectoryManager()

	output, err := executeRemoveCommand(t, []string{"test-project", "--force"}, "")

	// Remove should succeed even when directory deletion fails
	if err != nil {
		t.Fatalf("expected no error (directory deletion is non-fatal), got: %v", err)
	}

	if !strings.Contains(output, "✓ Removed: test-project") {
		t.Errorf("expected success message despite directory deletion error, got: %s", output)
	}

	// Verify project was removed from repository
	if len(mock.Projects) != 0 {
		t.Error("expected project to be removed from repository")
	}

	// Verify DeleteProjectDir was still called
	if len(mockDM.DeleteCalls()) != 1 {
		t.Errorf("expected DeleteProjectDir to be called, got %d calls", len(mockDM.DeleteCalls()))
	}
}

func TestRemove_NoDirectoryManagerIsGraceful(t *testing.T) {
	// Verify remove works when directoryManager is nil (backward compatibility)
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test-project", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	// Explicitly clear directory manager
	ClearMockDirectoryManager()

	output, err := executeRemoveCommand(t, []string{"test-project", "--force"}, "")
	if err != nil {
		t.Fatalf("expected no error without directory manager, got: %v", err)
	}

	if !strings.Contains(output, "✓ Removed: test-project") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify project was removed from repository
	if len(mock.Projects) != 0 {
		t.Error("expected project to be removed from repository")
	}
}

// ============================================================================
// Story 6.7: Quiet Mode Tests
// ============================================================================

func TestRemove_QuietMode_SuppressesOutput(t *testing.T) {
	// AC3: remove with --quiet --force produces no output
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test-project", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	cli.ResetRemoveFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterRemoveCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"remove", "test-project", "--force"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify output is empty (quiet mode)
	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet --force, got: %s", output)
	}

	// Verify project was still removed
	if len(mock.Projects) != 0 {
		t.Error("expected project to be deleted from repository")
	}
}

func TestRemove_QuietMode_ForceRequired(t *testing.T) {
	// AC6: Combined --force --quiet works together
	mock := NewMockRepository()
	p, _ := domain.NewProject("/path/to/test-project", "")
	mock.Projects[p.Path] = p
	cli.SetRepository(mock)

	cli.ResetRemoveFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterRemoveCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"remove", "test-project", "--force"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	// Should be completely silent
	if output != "" {
		t.Errorf("AC6: expected empty output with combined --force --quiet, got: %s", output)
	}

	// Exit code should be 0
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitSuccess {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}
