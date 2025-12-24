package cli_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 6.5: CLI Rename Command Tests
// ============================================================================

// renameMockRepository implements ports.ProjectRepository for rename tests.
type renameMockRepository struct {
	projects   map[string]*domain.Project
	saveErr    error
	findAllErr error
}

func newRenameMockRepository() *renameMockRepository {
	return &renameMockRepository{projects: make(map[string]*domain.Project)}
}

func (m *renameMockRepository) withProjects(projects []*domain.Project) *renameMockRepository {
	for _, p := range projects {
		m.projects[p.Path] = p
	}
	return m
}

func (m *renameMockRepository) Save(_ context.Context, project *domain.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.projects[project.Path] = project
	return nil
}

func (m *renameMockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *renameMockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
	if p, ok := m.projects[path]; ok {
		return p, nil
	}
	return nil, domain.ErrProjectNotFound
}

func (m *renameMockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	result := make([]*domain.Project, 0, len(m.projects))
	for _, p := range m.projects {
		result = append(result, p)
	}
	return result, nil
}

func (m *renameMockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *renameMockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *renameMockRepository) Delete(_ context.Context, id string) error {
	for path, p := range m.projects {
		if p.ID == id {
			delete(m.projects, path)
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *renameMockRepository) UpdateState(_ context.Context, id string, state domain.ProjectState) error {
	for _, p := range m.projects {
		if p.ID == id {
			p.State = state
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *renameMockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}

// executeRenameCommand runs the rename command with given args and returns output/error
func executeRenameCommand(args []string) (string, error) {
	cmd := cli.NewRootCmd()
	cli.RegisterRenameCommand(cmd)
	cli.ResetRenameFlags()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"rename"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// TestRenameCmd_SetDisplayName verifies setting display name (AC1).
func TestRenameCmd_SetDisplayName(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "api-service", DisplayName: ""},
	}
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute
	output, err := executeRenameCommand([]string{"api-service", "Client A API"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Renamed: api-service → Client A API") {
		t.Errorf("expected '✓ Renamed: api-service → Client A API', got: %s", output)
	}
	if projects[0].DisplayName != "Client A API" {
		t.Errorf("expected DisplayName to be 'Client A API', got: %s", projects[0].DisplayName)
	}
}

// TestRenameCmd_ClearWithFlag verifies --clear flag clears display name (AC2).
func TestRenameCmd_ClearWithFlag(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "api-service", DisplayName: "Client A API"},
	}
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute
	output, err := executeRenameCommand([]string{"api-service", "--clear"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Cleared display name: api-service") {
		t.Errorf("expected '✓ Cleared display name: api-service', got: %s", output)
	}
	if projects[0].DisplayName != "" {
		t.Errorf("expected DisplayName to be empty, got: %s", projects[0].DisplayName)
	}
}

// TestRenameCmd_ClearWithEmptyString verifies empty string clears display name (AC3).
func TestRenameCmd_ClearWithEmptyString(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "api-service", DisplayName: "Client A API"},
	}
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute
	output, err := executeRenameCommand([]string{"api-service", ""})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Cleared display name: api-service") {
		t.Errorf("expected '✓ Cleared display name: api-service', got: %s", output)
	}
	if projects[0].DisplayName != "" {
		t.Errorf("expected DisplayName to be empty, got: %s", projects[0].DisplayName)
	}
}

// TestRenameCmd_ProjectNotFound verifies error for non-existent project (AC4).
func TestRenameCmd_ProjectNotFound(t *testing.T) {
	// Setup
	cli.SetRepository(newRenameMockRepository()) // Empty repository

	// Execute
	_, err := executeRenameCommand([]string{"nonexistent", "New Name"})

	// Assert
	if err == nil {
		t.Fatal("expected error for non-existent project")
	}
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("expected domain.ErrProjectNotFound, got: %v", err)
	}
}

// TestRenameCmd_ProjectNotFoundExitCode2 verifies exit code 2 for not found (AC4).
func TestRenameCmd_ProjectNotFoundExitCode2(t *testing.T) {
	// Setup
	cli.SetRepository(newRenameMockRepository()) // Empty repository

	// Execute - we need to check error type to determine exit code
	_, err := executeRenameCommand([]string{"nonexistent", "New Name"})

	// Assert: Error should be ErrProjectNotFound which maps to exit code 2
	if err == nil {
		t.Fatal("expected error for non-existent project")
	}
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != 2 {
		t.Errorf("expected exit code 2, got: %d", exitCode)
	}
}

// TestRenameCmd_NoArguments verifies error when no arguments (AC5).
func TestRenameCmd_NoArguments(t *testing.T) {
	// Setup
	cli.SetRepository(newRenameMockRepository())

	// Execute with no args
	_, err := executeRenameCommand([]string{})

	// Assert - cobra should return error for missing required arg
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}

// TestRenameCmd_RequiresNewNameOrClear verifies error when only project name (AC6).
func TestRenameCmd_RequiresNewNameOrClear(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "api-service"},
	}
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute with only project name
	_, err := executeRenameCommand([]string{"api-service"})

	// Assert
	if err == nil {
		t.Fatal("expected error for missing new name or --clear flag")
	}
	if !strings.Contains(err.Error(), "requires a new name or --clear flag") {
		t.Errorf("expected 'requires a new name or --clear flag' error, got: %v", err)
	}
}

// TestRenameCmd_LookupByDisplayName verifies lookup by display name (AC7).
func TestRenameCmd_LookupByDisplayName(t *testing.T) {
	// Setup: Project with display name
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "client-alpha", DisplayName: "My Client"},
	}
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute using display name
	output, err := executeRenameCommand([]string{"My Client", "New Display Name"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Renamed") {
		t.Errorf("expected '✓ Renamed', got: %s", output)
	}
	if projects[0].DisplayName != "New Display Name" {
		t.Errorf("expected DisplayName to be 'New Display Name', got: %s", projects[0].DisplayName)
	}
}

// TestRenameCmd_LookupByPath verifies lookup by path (AC7).
func TestRenameCmd_LookupByPath(t *testing.T) {
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

	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: canonicalPath, Name: "my-project"},
	}
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute using path
	output, err := executeRenameCommand([]string{projectPath, "New Name"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Renamed") {
		t.Errorf("expected '✓ Renamed', got: %s", output)
	}
	if projects[0].DisplayName != "New Name" {
		t.Errorf("expected DisplayName to be 'New Name', got: %s", projects[0].DisplayName)
	}
}

// TestRenameCmd_IdempotentClear verifies idempotent clear on already-cleared (AC8).
func TestRenameCmd_IdempotentClear(t *testing.T) {
	// Setup: Project with no display name
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "api-service", DisplayName: ""},
	}
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute
	output, err := executeRenameCommand([]string{"api-service", "--clear"})

	// Assert: No error, idempotent message
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "☆ api-service has no display name") {
		t.Errorf("expected '☆ api-service has no display name', got: %s", output)
	}
}

// TestRenameCmd_IdempotentClearExitCode0 verifies exit code 0 for idempotent clear (AC8).
func TestRenameCmd_IdempotentClearExitCode0(t *testing.T) {
	// Setup: Project with no display name
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "api-service", DisplayName: ""},
	}
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute
	_, err := executeRenameCommand([]string{"api-service", "--clear"})

	// Assert: No error means exit code 0
	if err != nil {
		t.Errorf("expected no error (exit code 0), got: %v", err)
	}
}

// TestRenameCmd_UpdatesUpdatedAt verifies UpdatedAt is updated after rename.
func TestRenameCmd_UpdatesUpdatedAt(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", DisplayName: ""},
	}
	originalUpdatedAt := projects[0].UpdatedAt
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute
	_, err := executeRenameCommand([]string{"test-project", "New Name"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !projects[0].UpdatedAt.After(originalUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

// TestRenameCmd_FindAllError verifies error handling when FindAll fails.
func TestRenameCmd_FindAllError(t *testing.T) {
	// Setup: Repository that fails on FindAll
	mockRepo := newRenameMockRepository()
	mockRepo.findAllErr = errors.New("database connection failed")
	cli.SetRepository(mockRepo)

	// Execute
	_, err := executeRenameCommand([]string{"test-project", "New Name"})

	// Assert
	if err == nil {
		t.Fatal("expected error when FindAll fails")
	}
}

// TestRenameCmd_SaveError verifies error handling when Save fails.
func TestRenameCmd_SaveError(t *testing.T) {
	// Setup: Repository that fails on Save
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", DisplayName: ""},
	}
	mockRepo := newRenameMockRepository().withProjects(projects)
	mockRepo.saveErr = errors.New("disk full")
	cli.SetRepository(mockRepo)

	// Execute
	_, err := executeRenameCommand([]string{"test-project", "New Name"})

	// Assert
	if err == nil {
		t.Fatal("expected error when Save fails")
	}
	if !strings.Contains(err.Error(), "failed to save") {
		t.Errorf("expected 'failed to save' error, got: %v", err)
	}
}

// TestRenameCmd_EmptyStringWithClearFlag verifies empty string + --clear flag edge case.
// When both are provided, --clear takes precedence (checked first in code).
func TestRenameCmd_EmptyStringWithClearFlag(t *testing.T) {
	// Setup: Project with display name
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "api-service", DisplayName: "Current Name"},
	}
	cli.SetRepository(newRenameMockRepository().withProjects(projects))

	// Execute with both empty string AND --clear flag
	output, err := executeRenameCommand([]string{"api-service", "", "--clear"})

	// Assert: Should succeed (--clear takes precedence)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Cleared display name: api-service") {
		t.Errorf("expected clear message, got: %s", output)
	}
	if projects[0].DisplayName != "" {
		t.Errorf("expected DisplayName to be empty, got: %s", projects[0].DisplayName)
	}
}

// TestRenameCmd_HelpText verifies help text output is correct.
func TestRenameCmd_HelpText(t *testing.T) {
	cmd := cli.NewRootCmd()
	cli.RegisterRenameCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"rename", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify key elements in help text
	if !strings.Contains(output, "rename <project-name> [new-display-name]") {
		t.Error("help text missing usage pattern")
	}
	if !strings.Contains(output, "--clear") {
		t.Error("help text missing --clear flag")
	}
	if !strings.Contains(output, "Set or clear a project's display name") {
		t.Error("help text missing short description")
	}
}

// ============================================================================
// Story 6.7: Quiet Mode Tests
// ============================================================================

func TestRenameCmd_QuietMode_SetDisplayName(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", DisplayName: ""},
	}
	mockRepo := newRenameMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	cli.ResetRenameFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterRenameCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"rename", "test-project", "New Name"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet, got: %s", output)
	}

	// Verify display name was set
	if projects[0].DisplayName != "New Name" {
		t.Errorf("expected display name 'New Name', got: %s", projects[0].DisplayName)
	}
}

func TestRenameCmd_QuietMode_ClearDisplayName(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", DisplayName: "Existing Name"},
	}
	mockRepo := newRenameMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	cli.ResetRenameFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterRenameCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"rename", "test-project", "--clear"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet, got: %s", output)
	}

	// Verify display name was cleared
	if projects[0].DisplayName != "" {
		t.Errorf("expected display name cleared, got: %s", projects[0].DisplayName)
	}
}

func TestRenameCmd_QuietMode_IdempotentClear(t *testing.T) {
	// Idempotent clear should also be quiet
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", DisplayName: ""},
	}
	mockRepo := newRenameMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	cli.ResetRenameFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterRenameCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"rename", "test-project", "--clear"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet (idempotent clear), got: %s", output)
	}
}
