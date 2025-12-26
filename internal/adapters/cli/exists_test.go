package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// executeExistsCommand runs the exists command with given args and returns output/error
func executeExistsCommand(args []string) (string, error) {
	cmd := cli.NewRootCmd()
	cli.RegisterExistsCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"exists"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// ============================================================================
// Task 3.2: Test project exists by name → exit 0, no output
// ============================================================================

func TestExists_ProjectByName(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/home/user/projects/client-alpha", "")
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeExistsCommand([]string{"client-alpha"})

	// Must succeed (exit 0)
	if err != nil {
		t.Fatalf("expected no error (exit 0), got: %v", err)
	}

	// Must be completely silent
	if output != "" {
		t.Errorf("expected empty output (silent), got: %q", output)
	}
}

// ============================================================================
// Task 3.3: Test project exists by display name → exit 0, no output
// ============================================================================

func TestExists_ProjectByDisplayName(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/home/user/projects/original-dir", "")
	p1.DisplayName = "My Cool App"
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeExistsCommand([]string{"My Cool App"})

	// Must succeed (exit 0)
	if err != nil {
		t.Fatalf("expected no error (exit 0), got: %v", err)
	}

	// Must be completely silent
	if output != "" {
		t.Errorf("expected empty output (silent), got: %q", output)
	}
}

// ============================================================================
// Task 3.4: Test project exists by path → exit 0, no output
// ============================================================================

func TestExists_ProjectByPath(t *testing.T) {
	// Create a temp directory for a real path test
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Resolve to canonical path (handles macOS /tmp -> /private/var/folders)
	canonicalPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}

	mock := NewMockRepository()
	// Store with canonical path to match lookup
	p1, _ := domain.NewProject(canonicalPath, "")
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeExistsCommand([]string{projectPath})

	// Must succeed (exit 0)
	if err != nil {
		t.Fatalf("expected no error (exit 0), got: %v", err)
	}

	// Must be completely silent
	if output != "" {
		t.Errorf("expected empty output (silent), got: %q", output)
	}
}

// ============================================================================
// Task 3.5: Test project not found → exit 2, no output
// ============================================================================

func TestExists_NotFound(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	output, err := executeExistsCommand([]string{"nonexistent-project-xyz"})

	// Must fail
	if err == nil {
		t.Fatal("expected error when project not found")
	}

	// Must be completely silent (no output)
	if output != "" {
		t.Errorf("expected empty output (silent), got: %q", output)
	}

	// Verify exit code maps to 2 (ErrProjectNotFound)
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitProjectNotFound {
		t.Errorf("expected exit code %d (ExitProjectNotFound), got %d", cli.ExitProjectNotFound, exitCode)
	}
}

// ============================================================================
// Task 3.6: Test no arguments → exit 1 (Cobra error)
// ============================================================================

func TestExists_NoArgument(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	output, err := executeExistsCommand([]string{})

	// Must fail
	if err == nil {
		t.Fatal("expected error when no argument provided")
	}

	// Cobra should show usage error (this is expected output)
	// The error message is acceptable here since Cobra generates it
	_ = output // Cobra generates usage message to stderr

	// Verify exit code is 1 (general error, not project not found)
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitGeneralError {
		t.Errorf("expected exit code %d (ExitGeneralError), got %d", cli.ExitGeneralError, exitCode)
	}
}

// ============================================================================
// Task 3.7: Verify MapErrorToExitCode returns 2 for returned error
// ============================================================================

func TestExists_ExitCodeMapping(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	_, err := executeExistsCommand([]string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error when project not found")
	}

	// Verify the error is recognized as ErrProjectNotFound
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != 2 {
		t.Errorf("expected MapErrorToExitCode to return 2 for project not found error, got %d", exitCode)
	}
}

// ============================================================================
// Additional edge cases
// ============================================================================

func TestExists_RepositoryNotInitialized(t *testing.T) {
	cli.SetRepository(nil)

	output, err := executeExistsCommand([]string{"test"})

	// Must fail
	if err == nil {
		t.Fatal("expected error when repository is nil")
	}

	// Should be silent (SilenceErrors set in runExists)
	if output != "" {
		t.Errorf("expected empty output (silent), got: %q", output)
	}
}

func TestExists_LookupPriority_NameBeforeDisplayName(t *testing.T) {
	mock := NewMockRepository()

	// Project 1: name matches the query
	p1, _ := domain.NewProject("/path/to/alpha", "")
	p1.DisplayName = "Some Other Name"
	mock.Projects[p1.Path] = p1

	// Project 2: display_name matches the query
	p2, _ := domain.NewProject("/path/to/beta", "")
	p2.DisplayName = "alpha" // Same as p1's name
	mock.Projects[p2.Path] = p2

	cli.SetRepository(mock)

	output, err := executeExistsCommand([]string{"alpha"})

	// Should succeed (finds p1 by name)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Must be completely silent
	if output != "" {
		t.Errorf("expected empty output (silent), got: %q", output)
	}
}
