package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 6.6: Shell Completion Tests
// ============================================================================

// completionMockRepository implements ports.ProjectRepository for completion tests.
type completionMockRepository struct {
	projects   []*domain.Project
	findAllErr error
}

func (m *completionMockRepository) Save(_ context.Context, _ *domain.Project) error {
	return nil
}

func (m *completionMockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *completionMockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.Path == path {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *completionMockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	return m.projects, nil
}

func (m *completionMockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *completionMockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *completionMockRepository) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *completionMockRepository) UpdateState(_ context.Context, _ string, _ domain.ProjectState) error {
	return nil
}

func (m *completionMockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (m *completionMockRepository) ResetProject(_ context.Context, _ string) error {
	return nil
}

func (m *completionMockRepository) ResetAll(_ context.Context) (int, error) {
	return 0, nil
}

// executeCompletionCommand runs the completion command with given args and returns output/error
func executeCompletionCommand(args []string) (string, error) {
	cmd := cli.NewRootCmd()
	cli.RegisterCompletionCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"completion"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// TestCompletionCmd_Bash verifies bash completion script generation (AC1).
func TestCompletionCmd_Bash(t *testing.T) {
	output, err := executeCompletionCommand([]string{"bash"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify bash-specific patterns (Story 6.6 Dev Notes: contains `_vdash` function pattern)
	if !strings.Contains(output, "_vdash") {
		t.Error("bash completion script should contain '_vdash' function")
	}
	if !strings.Contains(output, "COMPREPLY") {
		t.Error("bash completion script should contain 'COMPREPLY'")
	}
}

// TestCompletionCmd_Zsh verifies zsh completion script generation (AC2).
func TestCompletionCmd_Zsh(t *testing.T) {
	output, err := executeCompletionCommand([]string{"zsh"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify zsh-specific patterns (Story 6.6 Dev Notes: contains `#compdef` pattern)
	if !strings.Contains(output, "#compdef") {
		t.Error("zsh completion script should contain '#compdef'")
	}
}

// TestCompletionCmd_Fish verifies fish completion script generation (AC3).
func TestCompletionCmd_Fish(t *testing.T) {
	output, err := executeCompletionCommand([]string{"fish"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify fish-specific patterns (Story 6.6 Dev Notes: contains `complete -c vdash` pattern)
	if !strings.Contains(output, "complete -c vdash") {
		t.Error("fish completion script should contain 'complete -c vdash'")
	}
}

// TestCompletionCmd_PowerShell verifies PowerShell completion script generation (AC4).
func TestCompletionCmd_PowerShell(t *testing.T) {
	output, err := executeCompletionCommand([]string{"powershell"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify PowerShell-specific patterns (Story 6.6 Dev Notes: contains `Register-ArgumentCompleter`)
	if !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Error("powershell completion script should contain 'Register-ArgumentCompleter'")
	}
}

// TestCompletionCmd_InvalidShell verifies error for unknown shell (AC9).
func TestCompletionCmd_InvalidShell(t *testing.T) {
	_, err := executeCompletionCommand([]string{"invalid-shell"})

	// Cobra's ValidArgs should reject invalid shell
	if err == nil {
		t.Fatal("expected error for invalid shell")
	}
}

// TestCompletionCmd_NoArgs verifies error when no shell argument provided.
func TestCompletionCmd_NoArgs(t *testing.T) {
	_, err := executeCompletionCommand([]string{})

	// Cobra's ExactArgs(1) should require one argument
	if err == nil {
		t.Fatal("expected error for missing shell argument")
	}
}

// TestCompletionCmd_HelpText verifies help text contains installation instructions (AC8).
func TestCompletionCmd_HelpText(t *testing.T) {
	cmd := cli.NewRootCmd()
	cli.RegisterCompletionCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"completion", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify installation instructions for all shells
	if !strings.Contains(output, "bash") {
		t.Error("help text should mention bash")
	}
	if !strings.Contains(output, "zsh") {
		t.Error("help text should mention zsh")
	}
	if !strings.Contains(output, "fish") {
		t.Error("help text should mention fish")
	}
	if !strings.Contains(output, "powershell") {
		t.Error("help text should mention powershell")
	}
	if !strings.Contains(output, "source <(vdash completion bash)") {
		t.Error("help text should contain bash installation instructions")
	}
}

// TestProjectCompletionFunc_ReturnsProjectNames verifies completion returns project names (AC6).
func TestProjectCompletionFunc_ReturnsProjectNames(t *testing.T) {
	// Setup mock repository with projects
	mockRepo := &completionMockRepository{
		projects: []*domain.Project{
			{ID: "1", Name: "project-alpha", DisplayName: ""},
			{ID: "2", Name: "project-beta", DisplayName: "My Beta Project"},
			{ID: "3", Name: "project-gamma", DisplayName: "project-gamma"}, // Same as name
		},
	}
	cli.SetRepository(mockRepo)
	defer cli.SetRepository(nil)

	// Call projectCompletionFunc via the ValidArgsFunction on status command
	statusCmd := cli.NewRootCmd()
	cli.RegisterStatusCommand(statusCmd)

	// Find the status command
	var foundCmd *cobra.Command
	for _, c := range statusCmd.Commands() {
		if c.Name() == "status" {
			foundCmd = c
			break
		}
	}

	if foundCmd == nil {
		t.Fatal("status command not found")
	}

	// Test the ValidArgsFunction
	if foundCmd.ValidArgsFunction == nil {
		t.Fatal("status command should have ValidArgsFunction")
	}

	completions, directive := foundCmd.ValidArgsFunction(foundCmd, []string{}, "")

	// Verify completions contain project names
	if !contains(completions, "project-alpha") {
		t.Error("completions should contain 'project-alpha'")
	}
	if !contains(completions, "project-beta") {
		t.Error("completions should contain 'project-beta'")
	}
	if !contains(completions, "My Beta Project") {
		t.Error("completions should contain 'My Beta Project' (display name)")
	}
	if !contains(completions, "project-gamma") {
		t.Error("completions should contain 'project-gamma'")
	}

	// project-gamma display name equals name, should NOT be duplicated
	count := 0
	for _, c := range completions {
		if c == "project-gamma" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("'project-gamma' should appear exactly once, found %d times", count)
	}

	// Verify directive is NoFileComp
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

// TestProjectCompletionFunc_AlreadyHasArg verifies no completion when arg already provided.
func TestProjectCompletionFunc_AlreadyHasArg(t *testing.T) {
	// Setup mock repository
	mockRepo := &completionMockRepository{
		projects: []*domain.Project{
			{ID: "1", Name: "project-alpha"},
		},
	}
	cli.SetRepository(mockRepo)
	defer cli.SetRepository(nil)

	// Create status command
	statusCmd := cli.NewRootCmd()
	cli.RegisterStatusCommand(statusCmd)

	var foundCmd *cobra.Command
	for _, c := range statusCmd.Commands() {
		if c.Name() == "status" {
			foundCmd = c
			break
		}
	}

	if foundCmd == nil {
		t.Fatal("status command not found")
	}

	// Call with existing arg - should return no completions
	completions, directive := foundCmd.ValidArgsFunction(foundCmd, []string{"existing-arg"}, "")

	if completions != nil {
		t.Error("completions should be nil when arg already provided")
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

// TestProjectCompletionFunc_NilRepository verifies error directive when repository nil.
func TestProjectCompletionFunc_NilRepository(t *testing.T) {
	cli.SetRepository(nil)

	// Create status command
	statusCmd := cli.NewRootCmd()
	cli.RegisterStatusCommand(statusCmd)

	var foundCmd *cobra.Command
	for _, c := range statusCmd.Commands() {
		if c.Name() == "status" {
			foundCmd = c
			break
		}
	}

	if foundCmd == nil {
		t.Fatal("status command not found")
	}

	completions, directive := foundCmd.ValidArgsFunction(foundCmd, []string{}, "")

	if completions != nil {
		t.Error("completions should be nil when repository is nil")
	}
	if directive != cobra.ShellCompDirectiveError {
		t.Errorf("expected ShellCompDirectiveError, got %v", directive)
	}
}

// TestProjectCompletionFunc_FindAllError verifies error directive when FindAll fails.
func TestProjectCompletionFunc_FindAllError(t *testing.T) {
	mockRepo := &completionMockRepository{
		findAllErr: errors.New("database error"),
	}
	cli.SetRepository(mockRepo)
	defer cli.SetRepository(nil)

	// Create status command
	statusCmd := cli.NewRootCmd()
	cli.RegisterStatusCommand(statusCmd)

	var foundCmd *cobra.Command
	for _, c := range statusCmd.Commands() {
		if c.Name() == "status" {
			foundCmd = c
			break
		}
	}

	if foundCmd == nil {
		t.Fatal("status command not found")
	}

	completions, directive := foundCmd.ValidArgsFunction(foundCmd, []string{}, "")

	if completions != nil {
		t.Error("completions should be nil when FindAll errors")
	}
	if directive != cobra.ShellCompDirectiveError {
		t.Errorf("expected ShellCompDirectiveError, got %v", directive)
	}
}

// TestCommandsHaveValidArgsFunction verifies all project-accepting commands have completion.
func TestCommandsHaveValidArgsFunction(t *testing.T) {
	commands := []struct {
		name     string
		register func(parent *cobra.Command)
	}{
		{"status", cli.RegisterStatusCommand},
		{"remove", cli.RegisterRemoveCommand},
		{"favorite", cli.RegisterFavoriteCommand},
		{"note", cli.RegisterNoteCommand},
		{"rename", cli.RegisterRenameCommand},
		{"exists", cli.RegisterExistsCommand},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			rootCmd := cli.NewRootCmd()
			tc.register(rootCmd)

			var foundCmd *cobra.Command
			for _, c := range rootCmd.Commands() {
				if c.Name() == tc.name {
					foundCmd = c
					break
				}
			}

			if foundCmd == nil {
				t.Fatalf("%s command not found", tc.name)
			}

			if foundCmd.ValidArgsFunction == nil {
				t.Errorf("%s command should have ValidArgsFunction", tc.name)
			}
		})
	}
}

// contains checks if a slice contains a specific string.
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
