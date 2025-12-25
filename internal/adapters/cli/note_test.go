package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 3.7: CLI Note Command Tests
// ============================================================================

// noteMockRepository implements ports.ProjectRepository for note tests.
type noteMockRepository struct {
	projects   map[string]*domain.Project
	saveErr    error
	findAllErr error
}

func newNoteMockRepository() *noteMockRepository {
	return &noteMockRepository{projects: make(map[string]*domain.Project)}
}

func (m *noteMockRepository) withProjects(projects []*domain.Project) *noteMockRepository {
	for _, p := range projects {
		m.projects[p.Path] = p
	}
	return m
}

func (m *noteMockRepository) Save(_ context.Context, project *domain.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.projects[project.Path] = project
	return nil
}

func (m *noteMockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *noteMockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
	if p, ok := m.projects[path]; ok {
		return p, nil
	}
	return nil, domain.ErrProjectNotFound
}

func (m *noteMockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	result := make([]*domain.Project, 0, len(m.projects))
	for _, p := range m.projects {
		result = append(result, p)
	}
	return result, nil
}

func (m *noteMockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *noteMockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *noteMockRepository) Delete(_ context.Context, id string) error {
	for path, p := range m.projects {
		if p.ID == id {
			delete(m.projects, path)
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *noteMockRepository) UpdateState(_ context.Context, id string, state domain.ProjectState) error {
	for _, p := range m.projects {
		if p.ID == id {
			p.State = state
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *noteMockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (m *noteMockRepository) ResetProject(_ context.Context, _ string) error {
	return nil
}

func (m *noteMockRepository) ResetAll(_ context.Context) (int, error) {
	return 0, nil
}

// executeNoteCommand runs the note command with given args and returns output/error
func executeNoteCommand(args []string) (string, error) {
	cmd := cli.NewRootCmd()
	cli.RegisterNoteCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"note"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// TestNoteCmd_ViewNote verifies viewing existing note (AC6).
func TestNoteCmd_ViewNote(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: "existing note content"},
	}
	cli.SetRepository(newNoteMockRepository().withProjects(projects))

	// Execute
	output, err := executeNoteCommand([]string{"test-project"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "existing note content") {
		t.Errorf("expected note content, got: %s", output)
	}
}

// TestNoteCmd_ViewNoNote verifies viewing project without note (AC6).
func TestNoteCmd_ViewNoNote(t *testing.T) {
	// Setup: Project with no note
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: ""},
	}
	cli.SetRepository(newNoteMockRepository().withProjects(projects))

	// Execute
	output, err := executeNoteCommand([]string{"test-project"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "(no note set)") {
		t.Errorf("expected '(no note set)', got: %s", output)
	}
}

// TestNoteCmd_SetNote verifies setting a note (AC5).
func TestNoteCmd_SetNote(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: ""},
	}
	mockRepo := newNoteMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	// Execute
	output, err := executeNoteCommand([]string{"test-project", "new note content"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Note saved") {
		t.Errorf("expected success message, got: %s", output)
	}
	// Verify note was updated in mock
	if projects[0].Notes != "new note content" {
		t.Errorf("expected note to be updated, got: %s", projects[0].Notes)
	}
}

// TestNoteCmd_ClearNote verifies clearing a note (AC5 with empty string).
func TestNoteCmd_ClearNote(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: "existing"},
	}
	mockRepo := newNoteMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	// Execute with empty note
	output, err := executeNoteCommand([]string{"test-project", ""})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Note cleared") {
		t.Errorf("expected clear message, got: %s", output)
	}
	if projects[0].Notes != "" {
		t.Errorf("expected note to be cleared, got: %s", projects[0].Notes)
	}
}

// TestNoteCmd_ProjectNotFound verifies error for non-existent project (AC5, AC6).
func TestNoteCmd_ProjectNotFound(t *testing.T) {
	// Setup
	cli.SetRepository(newNoteMockRepository()) // Empty repository

	// Execute
	_, err := executeNoteCommand([]string{"nonexistent"})

	// Assert
	if err == nil {
		t.Fatal("expected error for non-existent project")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestNoteCmd_FindByDisplayName verifies finding project by display name.
func TestNoteCmd_FindByDisplayName(t *testing.T) {
	// Setup: Project with display name different from directory name
	projects := []*domain.Project{
		{ID: "1", Path: "/some/path", Name: "path", DisplayName: "my-custom-name", Notes: "note here"},
	}
	cli.SetRepository(newNoteMockRepository().withProjects(projects))

	// Execute using display name
	output, err := executeNoteCommand([]string{"my-custom-name"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "note here") {
		t.Errorf("expected to find project by display name, got: %s", output)
	}
}

// TestNoteCmd_SetNoteByDisplayName verifies setting note using display name.
func TestNoteCmd_SetNoteByDisplayName(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/some/path", Name: "path", DisplayName: "my-custom-name", Notes: ""},
	}
	cli.SetRepository(newNoteMockRepository().withProjects(projects))

	// Execute
	output, err := executeNoteCommand([]string{"my-custom-name", "new note via display name"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Note saved") {
		t.Errorf("expected success message, got: %s", output)
	}
	if projects[0].Notes != "new note via display name" {
		t.Errorf("expected note to be updated, got: %s", projects[0].Notes)
	}
}

// TestNoteCmd_NoArgs verifies error when no arguments provided.
func TestNoteCmd_NoArgs(t *testing.T) {
	// Setup
	cli.SetRepository(newNoteMockRepository())

	// Execute with no args
	_, err := executeNoteCommand([]string{})

	// Assert - cobra should return error for missing required arg
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}

// TestNoteCmd_ExitCodeSuccess verifies exit code 0 on success (AC5).
func TestNoteCmd_ExitCodeSuccess(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: ""},
	}
	cli.SetRepository(newNoteMockRepository().withProjects(projects))

	// Execute
	_, err := executeNoteCommand([]string{"test-project", "new note"})

	// Assert: No error means exit code 0
	if err != nil {
		t.Errorf("expected no error (exit code 0), got: %v", err)
	}
}

// TestNoteCmd_UpdatesUpdatedAt verifies UpdatedAt is updated when note changes.
func TestNoteCmd_UpdatesUpdatedAt(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: "old"},
	}
	originalUpdatedAt := projects[0].UpdatedAt
	cli.SetRepository(newNoteMockRepository().withProjects(projects))

	// Execute
	_, err := executeNoteCommand([]string{"test-project", "new note"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !projects[0].UpdatedAt.After(originalUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

// TestNoteCmd_FindAllError verifies error handling when FindAll fails.
func TestNoteCmd_FindAllError(t *testing.T) {
	// Setup: Repository that fails on FindAll
	mockRepo := newNoteMockRepository()
	mockRepo.findAllErr = errors.New("database connection failed")
	cli.SetRepository(mockRepo)

	// Execute
	_, err := executeNoteCommand([]string{"test-project"})

	// Assert
	if err == nil {
		t.Fatal("expected error when FindAll fails")
	}
	if !strings.Contains(err.Error(), "failed to load projects") {
		t.Errorf("expected 'failed to load projects' error, got: %v", err)
	}
}

// TestNoteCmd_SaveError verifies error handling when Save fails.
func TestNoteCmd_SaveError(t *testing.T) {
	// Setup: Repository that fails on Save
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: ""},
	}
	mockRepo := newNoteMockRepository().withProjects(projects)
	mockRepo.saveErr = errors.New("disk full")
	cli.SetRepository(mockRepo)

	// Execute
	_, err := executeNoteCommand([]string{"test-project", "new note"})

	// Assert
	if err == nil {
		t.Fatal("expected error when Save fails")
	}
	if !strings.Contains(err.Error(), "failed to save note") {
		t.Errorf("expected 'failed to save note' error, got: %v", err)
	}
}

// TestNoteCmd_SpecialCharacters verifies notes with special characters work.
func TestNoteCmd_SpecialCharacters(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: ""},
	}
	cli.SetRepository(newNoteMockRepository().withProjects(projects))

	// Execute with special characters
	specialNote := "待っています 🚀 <script>alert('xss')</script>"
	output, err := executeNoteCommand([]string{"test-project", specialNote})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "✓ Note saved") {
		t.Errorf("expected success message, got: %s", output)
	}
	// Verify note was saved with special characters
	if projects[0].Notes != specialNote {
		t.Errorf("expected note with special chars, got: %s", projects[0].Notes)
	}
}

// ============================================================================
// Story 6.7: Quiet Mode Tests
// ============================================================================

func TestNoteCmd_QuietMode_SetNote(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: ""},
	}
	mockRepo := newNoteMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterNoteCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"note", "test-project", "new note"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet, got: %s", output)
	}

	// Verify note was still saved
	if projects[0].Notes != "new note" {
		t.Errorf("expected note to be saved, got: %s", projects[0].Notes)
	}
}

func TestNoteCmd_QuietMode_ClearNote(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", Notes: "existing note"},
	}
	mockRepo := newNoteMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterNoteCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"note", "test-project", ""})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet, got: %s", output)
	}

	// Verify note was cleared
	if projects[0].Notes != "" {
		t.Errorf("expected note to be cleared, got: %s", projects[0].Notes)
	}
}
