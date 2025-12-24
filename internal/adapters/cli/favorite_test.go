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
// Story 3.8: CLI Favorite Command Tests
// ============================================================================

// favoriteMockRepository implements ports.ProjectRepository for favorite tests.
type favoriteMockRepository struct {
	projects   map[string]*domain.Project
	saveErr    error
	findAllErr error
}

func newFavoriteMockRepository() *favoriteMockRepository {
	return &favoriteMockRepository{projects: make(map[string]*domain.Project)}
}

func (m *favoriteMockRepository) withProjects(projects []*domain.Project) *favoriteMockRepository {
	for _, p := range projects {
		m.projects[p.Path] = p
	}
	return m
}

func (m *favoriteMockRepository) Save(_ context.Context, project *domain.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.projects[project.Path] = project
	return nil
}

func (m *favoriteMockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *favoriteMockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
	if p, ok := m.projects[path]; ok {
		return p, nil
	}
	return nil, domain.ErrProjectNotFound
}

func (m *favoriteMockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	result := make([]*domain.Project, 0, len(m.projects))
	for _, p := range m.projects {
		result = append(result, p)
	}
	return result, nil
}

func (m *favoriteMockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *favoriteMockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *favoriteMockRepository) Delete(_ context.Context, id string) error {
	for path, p := range m.projects {
		if p.ID == id {
			delete(m.projects, path)
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *favoriteMockRepository) UpdateState(_ context.Context, id string, state domain.ProjectState) error {
	for _, p := range m.projects {
		if p.ID == id {
			p.State = state
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *favoriteMockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}

// executeFavoriteCommand runs the favorite command with given args and returns output/error
func executeFavoriteCommand(args []string) (string, error) {
	cmd := cli.NewRootCmd()
	cli.RegisterFavoriteCommand(cmd)
	cli.ResetFavoriteFlags()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"favorite"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// TestFavoriteCmd_ToggleOn verifies toggling favorite on (AC3).
func TestFavoriteCmd_ToggleOn(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
	}
	cli.SetRepository(newFavoriteMockRepository().withProjects(projects))

	// Execute
	output, err := executeFavoriteCommand([]string{"test-project"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "⭐ Favorited") {
		t.Errorf("expected '⭐ Favorited', got: %s", output)
	}
	if !projects[0].IsFavorite {
		t.Error("expected project IsFavorite to be true")
	}
}

// TestFavoriteCmd_ToggleOff verifies toggling favorite off (AC3).
func TestFavoriteCmd_ToggleOff(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: true},
	}
	cli.SetRepository(newFavoriteMockRepository().withProjects(projects))

	// Execute
	output, err := executeFavoriteCommand([]string{"test-project"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "☆ Unfavorited") {
		t.Errorf("expected '☆ Unfavorited', got: %s", output)
	}
	if projects[0].IsFavorite {
		t.Error("expected project IsFavorite to be false")
	}
}

// TestFavoriteCmd_ExplicitOff verifies --off flag removes favorite (AC4).
func TestFavoriteCmd_ExplicitOff(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: true},
	}
	cli.SetRepository(newFavoriteMockRepository().withProjects(projects))

	// Execute
	output, err := executeFavoriteCommand([]string{"test-project", "--off"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "☆ Unfavorited") {
		t.Errorf("expected '☆ Unfavorited', got: %s", output)
	}
	if projects[0].IsFavorite {
		t.Error("expected project IsFavorite to be false")
	}
}

// TestFavoriteCmd_OffIdempotent verifies --off is idempotent (AC5).
func TestFavoriteCmd_OffIdempotent(t *testing.T) {
	// Setup: Project already not favorited
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
	}
	cli.SetRepository(newFavoriteMockRepository().withProjects(projects))

	// Execute
	output, err := executeFavoriteCommand([]string{"test-project", "--off"})

	// Assert: No error, idempotent message
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "is not favorited") {
		t.Errorf("expected idempotent message, got: %s", output)
	}
}

// TestFavoriteCmd_ProjectNotFound verifies error for non-existent project.
func TestFavoriteCmd_ProjectNotFound(t *testing.T) {
	// Setup
	cli.SetRepository(newFavoriteMockRepository()) // Empty repository

	// Execute
	_, err := executeFavoriteCommand([]string{"nonexistent"})

	// Assert
	if err == nil {
		t.Fatal("expected error for non-existent project")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestFavoriteCmd_FindByDisplayName verifies finding project by display name.
func TestFavoriteCmd_FindByDisplayName(t *testing.T) {
	// Setup: Project with display name different from directory name
	projects := []*domain.Project{
		{ID: "1", Path: "/some/path", Name: "path", DisplayName: "my-custom-name", IsFavorite: false},
	}
	cli.SetRepository(newFavoriteMockRepository().withProjects(projects))

	// Execute using display name
	output, err := executeFavoriteCommand([]string{"my-custom-name"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "⭐ Favorited") {
		t.Errorf("expected to favorite project by display name, got: %s", output)
	}
	if !projects[0].IsFavorite {
		t.Error("expected project IsFavorite to be true")
	}
}

// TestFavoriteCmd_NoArgs verifies error when no arguments provided.
func TestFavoriteCmd_NoArgs(t *testing.T) {
	// Setup
	cli.SetRepository(newFavoriteMockRepository())

	// Execute with no args
	_, err := executeFavoriteCommand([]string{})

	// Assert - cobra should return error for missing required arg
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}

// TestFavoriteCmd_ExitCodeSuccess verifies exit code 0 on success (AC3).
func TestFavoriteCmd_ExitCodeSuccess(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
	}
	cli.SetRepository(newFavoriteMockRepository().withProjects(projects))

	// Execute
	_, err := executeFavoriteCommand([]string{"test-project"})

	// Assert: No error means exit code 0
	if err != nil {
		t.Errorf("expected no error (exit code 0), got: %v", err)
	}
}

// TestFavoriteCmd_UpdatesUpdatedAt verifies UpdatedAt is updated when favorite changes.
func TestFavoriteCmd_UpdatesUpdatedAt(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
	}
	originalUpdatedAt := projects[0].UpdatedAt
	cli.SetRepository(newFavoriteMockRepository().withProjects(projects))

	// Execute
	_, err := executeFavoriteCommand([]string{"test-project"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !projects[0].UpdatedAt.After(originalUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

// TestFavoriteCmd_FindAllError verifies error handling when FindAll fails.
func TestFavoriteCmd_FindAllError(t *testing.T) {
	// Setup: Repository that fails on FindAll
	mockRepo := newFavoriteMockRepository()
	mockRepo.findAllErr = errors.New("database connection failed")
	cli.SetRepository(mockRepo)

	// Execute
	_, err := executeFavoriteCommand([]string{"test-project"})

	// Assert
	if err == nil {
		t.Fatal("expected error when FindAll fails")
	}
	if !strings.Contains(err.Error(), "failed to load projects") {
		t.Errorf("expected 'failed to load projects' error, got: %v", err)
	}
}

// TestFavoriteCmd_SaveError verifies error handling when Save fails.
func TestFavoriteCmd_SaveError(t *testing.T) {
	// Setup: Repository that fails on Save
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
	}
	mockRepo := newFavoriteMockRepository().withProjects(projects)
	mockRepo.saveErr = errors.New("disk full")
	cli.SetRepository(mockRepo)

	// Execute
	_, err := executeFavoriteCommand([]string{"test-project"})

	// Assert
	if err == nil {
		t.Fatal("expected error when Save fails")
	}
	if !strings.Contains(err.Error(), "failed to save") {
		t.Errorf("expected 'failed to save' error, got: %v", err)
	}
}

// ============================================================================
// Story 6.7: Quiet Mode Tests
// ============================================================================

func TestFavoriteCmd_QuietMode_ToggleOn(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
	}
	mockRepo := newFavoriteMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	cli.ResetFavoriteFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterFavoriteCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"favorite", "test-project"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet, got: %s", output)
	}

	// Verify favorite was set
	if !projects[0].IsFavorite {
		t.Error("expected project to be favorited")
	}
}

func TestFavoriteCmd_QuietMode_ToggleOff(t *testing.T) {
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: true},
	}
	mockRepo := newFavoriteMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	cli.ResetFavoriteFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterFavoriteCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"favorite", "test-project"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet, got: %s", output)
	}

	// Verify favorite was toggled off
	if projects[0].IsFavorite {
		t.Error("expected project to be unfavorited")
	}
}

func TestFavoriteCmd_QuietMode_IdempotentOff(t *testing.T) {
	// Idempotent off should also be quiet
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: false},
	}
	mockRepo := newFavoriteMockRepository().withProjects(projects)
	cli.SetRepository(mockRepo)

	cli.ResetFavoriteFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterFavoriteCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"favorite", "test-project", "--off"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet (idempotent off), got: %s", output)
	}
}
