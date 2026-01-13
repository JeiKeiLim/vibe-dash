package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 11.5: CLI Hibernate Command Tests
// ============================================================================

// mockStateService implements ports.StateActivator for hibernate/activate tests.
type mockStateService struct {
	hibernateErr    error
	activateErr     error
	hibernateCalled bool
	activateCalled  bool
	lastProjectID   string
}

func (m *mockStateService) Hibernate(_ context.Context, projectID string) error {
	m.hibernateCalled = true
	m.lastProjectID = projectID
	return m.hibernateErr
}

func (m *mockStateService) Activate(_ context.Context, projectID string) error {
	m.activateCalled = true
	m.lastProjectID = projectID
	return m.activateErr
}

// hibernateMockRepository implements ports.ProjectRepository for hibernate tests.
// Follows favorite_test.go pattern - local mock per test file.
type hibernateMockRepository struct {
	projects map[string]*domain.Project
	// For future error injection tests (currently unused):
	saveErr    error
	findAllErr error
}

func newHibernateMockRepository() *hibernateMockRepository {
	return &hibernateMockRepository{projects: make(map[string]*domain.Project)}
}

func (m *hibernateMockRepository) withProjects(projects []*domain.Project) *hibernateMockRepository {
	for _, p := range projects {
		m.projects[p.Path] = p
	}
	return m
}

func (m *hibernateMockRepository) Save(_ context.Context, project *domain.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.projects[project.Path] = project
	return nil
}

func (m *hibernateMockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *hibernateMockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
	if p, ok := m.projects[path]; ok {
		return p, nil
	}
	return nil, domain.ErrProjectNotFound
}

func (m *hibernateMockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	result := make([]*domain.Project, 0, len(m.projects))
	for _, p := range m.projects {
		result = append(result, p)
	}
	return result, nil
}

func (m *hibernateMockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *hibernateMockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *hibernateMockRepository) Delete(_ context.Context, id string) error {
	for path, p := range m.projects {
		if p.ID == id {
			delete(m.projects, path)
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *hibernateMockRepository) UpdateState(_ context.Context, id string, state domain.ProjectState) error {
	for _, p := range m.projects {
		if p.ID == id {
			p.State = state
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *hibernateMockRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (m *hibernateMockRepository) ResetProject(_ context.Context, _ string) error {
	return nil
}

func (m *hibernateMockRepository) ResetAll(_ context.Context) (int, error) {
	return 0, nil
}

// executeHibernateCommand runs the hibernate command with given args and returns output/error.
// Follows favorite_test.go:126-140 pattern.
func executeHibernateCommand(args []string) (string, error) {
	cmd := cli.NewRootCmd()
	cli.RegisterHibernateCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"hibernate"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// TestHibernateCmd_ActiveProject_Succeeds verifies hibernating active project (AC1).
func TestHibernateCmd_ActiveProject_Succeeds(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateActive},
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
	output, err := executeHibernateCommand([]string{"test-project"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mockState.hibernateCalled {
		t.Error("expected Hibernate to be called")
	}
	if !strings.Contains(output, "✓ Hibernated") {
		t.Errorf("expected '✓ Hibernated', got: %s", output)
	}
}

// TestHibernateCmd_AlreadyHibernated_Idempotent verifies idempotent behavior (AC3).
func TestHibernateCmd_AlreadyHibernated_Idempotent(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateHibernated},
	}
	mockRepo := newHibernateMockRepository().withProjects(projects)
	mockState := &mockStateService{hibernateErr: domain.ErrInvalidStateTransition}

	cli.SetRepository(mockRepo)
	cli.SetStateService(mockState)
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute
	output, err := executeHibernateCommand([]string{"test-project"})

	// Assert - should succeed (idempotent)
	if err != nil {
		t.Fatalf("expected no error (idempotent), got: %v", err)
	}
	if !strings.Contains(output, "already hibernated") {
		t.Errorf("expected 'already hibernated' message, got: %s", output)
	}
}

// TestHibernateCmd_FavoriteProject_ReturnsError verifies favorite rejection (AC5).
func TestHibernateCmd_FavoriteProject_ReturnsError(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", IsFavorite: true, State: domain.StateActive},
	}
	mockRepo := newHibernateMockRepository().withProjects(projects)
	mockState := &mockStateService{hibernateErr: domain.ErrFavoriteCannotHibernate}

	cli.SetRepository(mockRepo)
	cli.SetStateService(mockState)
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute
	output, err := executeHibernateCommand([]string{"test-project"})

	// Assert
	if err == nil {
		t.Fatal("expected error for favorite project")
	}
	if !strings.Contains(output, "Cannot hibernate favorite") {
		t.Errorf("expected favorite error message, got: %s", output)
	}
	if !strings.Contains(output, "vdash favorite test-project --off") {
		t.Errorf("expected hint about removing favorite, got: %s", output)
	}
}

// TestHibernateCmd_ProjectNotFound_ReturnsError verifies not found handling (AC6).
func TestHibernateCmd_ProjectNotFound_ReturnsError(t *testing.T) {
	// Setup - empty repository
	cli.SetRepository(newHibernateMockRepository())
	cli.SetStateService(&mockStateService{})
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute
	output, err := executeHibernateCommand([]string{"nonexistent"})

	// Assert
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
	if !strings.Contains(output, "Project not found") {
		t.Errorf("expected 'Project not found' message, got: %s", output)
	}
}

// TestHibernateCmd_QuietMode_SuppressesOutput verifies quiet mode (AC7).
func TestHibernateCmd_QuietMode_SuppressesOutput(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateActive},
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
	cli.RegisterHibernateCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hibernate", "test-project"})

	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output with --quiet, got: %s", buf.String())
	}
}

// TestHibernateCmd_StateServiceNil_ReturnsError verifies nil stateService handling.
func TestHibernateCmd_StateServiceNil_ReturnsError(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateActive},
	}
	cli.SetRepository(newHibernateMockRepository().withProjects(projects))
	cli.SetStateService(nil) // Explicitly nil
	defer func() {
		cli.SetRepository(nil)
	}()

	// Execute
	_, err := executeHibernateCommand([]string{"test-project"})

	// Assert
	if err == nil {
		t.Fatal("expected error when stateService is nil")
	}
}

// TestHibernateCmd_FindByPath verifies path-based lookup (AC8).
func TestHibernateCmd_FindByPath(t *testing.T) {
	// Setup: Create temp directory and get its canonical path
	tempDir := t.TempDir()
	canonicalDir, err := filesystem.CanonicalPath(tempDir)
	if err != nil {
		t.Fatalf("failed to get canonical path: %v", err)
	}

	projects := []*domain.Project{
		{ID: "1", Path: canonicalDir, Name: "my-project", State: domain.StateActive},
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
	output, err := executeHibernateCommand([]string{tempDir})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mockState.hibernateCalled {
		t.Error("expected Hibernate to be called via path lookup")
	}
	if !strings.Contains(output, "✓ Hibernated") {
		t.Errorf("expected success message, got: %s", output)
	}
}

// TestHibernateCmd_FindByDisplayName verifies display name lookup (AC8).
func TestHibernateCmd_FindByDisplayName(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "my-project", DisplayName: "My Cool App", State: domain.StateActive},
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
	output, err := executeHibernateCommand([]string{"My Cool App"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mockState.hibernateCalled {
		t.Error("expected Hibernate to be called via display name lookup")
	}
	if !strings.Contains(output, "✓ Hibernated") {
		t.Errorf("expected success message, got: %s", output)
	}
}

// TestHibernateCmd_NoArgs_ReturnsError verifies error when no arguments provided.
func TestHibernateCmd_NoArgs_ReturnsError(t *testing.T) {
	// Setup
	cli.SetRepository(newHibernateMockRepository())
	cli.SetStateService(&mockStateService{})
	defer func() {
		cli.SetRepository(nil)
		cli.SetStateService(nil)
	}()

	// Execute with no args
	_, err := executeHibernateCommand([]string{})

	// Assert - cobra should return error for missing required arg
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}

// TestHibernateCmd_QuietMode_AlreadyHibernated_SuppressesOutput verifies quiet mode with idempotent case.
func TestHibernateCmd_QuietMode_AlreadyHibernated_SuppressesOutput(t *testing.T) {
	// Setup
	projects := []*domain.Project{
		{ID: "1", Path: "/test", Name: "test-project", State: domain.StateHibernated},
	}
	mockRepo := newHibernateMockRepository().withProjects(projects)
	mockState := &mockStateService{hibernateErr: domain.ErrInvalidStateTransition}

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
	cli.RegisterHibernateCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hibernate", "test-project"})

	err := cmd.Execute()

	// Assert - should succeed (idempotent) with no output
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output with --quiet (idempotent), got: %s", buf.String())
	}
}
