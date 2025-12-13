package cli_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// MockDetector implements ports.Detector for testing
type MockDetector struct {
	detectResult   *domain.DetectionResult
	detectErr      error
	multipleResult []*domain.DetectionResult
	multipleErr    error
}

func (m *MockDetector) Detect(_ context.Context, _ string) (*domain.DetectionResult, error) {
	return m.detectResult, m.detectErr
}

func (m *MockDetector) DetectMultiple(_ context.Context, _ string) ([]*domain.DetectionResult, error) {
	return m.multipleResult, m.multipleErr
}

// MockRepository implements ports.ProjectRepository for testing
type MockRepository struct {
	projects map[string]*domain.Project
	saveErr  error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{projects: make(map[string]*domain.Project)}
}

func (m *MockRepository) Save(_ context.Context, project *domain.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.projects[project.Path] = project
	return nil
}

func (m *MockRepository) FindByID(_ context.Context, id string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *MockRepository) FindByPath(_ context.Context, path string) (*domain.Project, error) {
	if p, ok := m.projects[path]; ok {
		return p, nil
	}
	return nil, domain.ErrProjectNotFound
}

func (m *MockRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0, len(m.projects))
	for _, p := range m.projects {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	result := make([]*domain.Project, 0)
	for _, p := range m.projects {
		if p.State == domain.StateHibernated {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockRepository) Delete(_ context.Context, id string) error {
	for path, p := range m.projects {
		if p.ID == id {
			delete(m.projects, path)
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

func (m *MockRepository) UpdateState(_ context.Context, id string, state domain.ProjectState) error {
	for _, p := range m.projects {
		if p.ID == id {
			p.State = state
			return nil
		}
	}
	return domain.ErrProjectNotFound
}

// executeAddCommand runs the add command with given args and returns output/error
func executeAddCommand(args []string) (string, error) {
	// Reset flags and root command for clean test
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"add"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

func TestAdd_CurrentDirectory(t *testing.T) {
	// Setup mock
	mock := NewMockRepository()
	cli.SetRepository(mock)

	// Create temp directory
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	// Execute add command with "."
	_, err := executeAddCommand([]string{"."})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify project was saved
	if len(mock.projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(mock.projects))
	}

	// Verify correct path was stored
	for path := range mock.projects {
		// Path should be canonical (might differ due to symlinks)
		if !filepath.IsAbs(path) {
			t.Errorf("expected absolute path, got %s", path)
		}
	}
}

func TestAdd_AbsolutePath(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	tmpDir := t.TempDir()

	_, err := executeAddCommand([]string{tmpDir})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(mock.projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(mock.projects))
	}
}

func TestAdd_WithNameFlag(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	tmpDir := t.TempDir()

	_, err := executeAddCommand([]string{tmpDir, "--name", "Custom Name"})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify display name was set
	for _, p := range mock.projects {
		if p.DisplayName != "Custom Name" {
			t.Errorf("expected DisplayName 'Custom Name', got '%s'", p.DisplayName)
		}
	}
}

func TestAdd_NonExistentPath(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	_, err := executeAddCommand([]string{"/this/path/does/not/exist"})

	if err == nil {
		t.Fatal("expected error for non-existent path")
	}

	// Verify exit code mapping
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitGeneralError {
		t.Errorf("expected exit code %d, got %d", cli.ExitGeneralError, exitCode)
	}
}

func TestAdd_AlreadyTracked(t *testing.T) {
	tmpDir := t.TempDir()

	// Get canonical path for the temp directory
	canonicalPath, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to get canonical path: %v", err)
	}

	// Setup mock with existing project at tmpDir path
	mock := NewMockRepository()
	mock.projects[canonicalPath] = &domain.Project{
		ID:   "abc123",
		Name: "existing-project",
		Path: canonicalPath,
	}
	cli.SetRepository(mock)

	// Execute add with same path
	_, err = executeAddCommand([]string{tmpDir})

	// Should return ErrProjectAlreadyExists (mapped to exit code 1)
	if err == nil {
		t.Fatal("expected error for already tracked project")
	}
	if cli.MapErrorToExitCode(err) != cli.ExitGeneralError {
		t.Errorf("expected exit code %d, got %d", cli.ExitGeneralError, cli.MapErrorToExitCode(err))
	}
}

func TestAdd_SymlinkCollision(t *testing.T) {
	// Create temp directory structure
	baseDir := t.TempDir()
	projectDir := filepath.Join(baseDir, "project")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	symlinkPath := filepath.Join(baseDir, "link-to-project")
	if err := os.Symlink(projectDir, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Get canonical path
	canonicalPath, err := filepath.EvalSymlinks(projectDir)
	if err != nil {
		t.Fatalf("failed to get canonical path: %v", err)
	}

	// Setup mock with project at canonical path
	mock := NewMockRepository()
	mock.projects[canonicalPath] = &domain.Project{
		ID:   "abc123",
		Name: "project",
		Path: canonicalPath,
	}
	cli.SetRepository(mock)

	// Try to add via symlink - should detect collision
	_, err = executeAddCommand([]string{symlinkPath})

	if err == nil {
		t.Fatal("expected error for symlink to already tracked project")
	}
	if cli.MapErrorToExitCode(err) != cli.ExitGeneralError {
		t.Errorf("expected exit code %d, got %d", cli.ExitGeneralError, cli.MapErrorToExitCode(err))
	}
}

func TestAdd_HomeDirectory(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	// Get home directory for comparison
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home directory: %v", err)
	}

	_, err = executeAddCommand([]string{"~"})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify home directory was added
	if len(mock.projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(mock.projects))
	}

	// Get canonical home path for comparison
	canonicalHome, err := filepath.EvalSymlinks(homeDir)
	if err != nil {
		t.Skipf("cannot get canonical home path: %v", err)
	}

	found := false
	for path := range mock.projects {
		if path == canonicalHome {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected project at home directory %s", canonicalHome)
	}
}

func TestAdd_NoArgs_DefaultsToCurrentDirectory(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	// Execute add command with no args
	_, err := executeAddCommand([]string{})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(mock.projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(mock.projects))
	}
}

func TestAdd_VerifyProjectFields(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	tmpDir := t.TempDir()
	dirName := filepath.Base(tmpDir)

	_, err := executeAddCommand([]string{tmpDir})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(mock.projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(mock.projects))
	}

	var project *domain.Project
	for _, p := range mock.projects {
		project = p
	}

	// Verify fields
	if project.ID == "" {
		t.Error("expected non-empty ID")
	}
	if project.Name != dirName {
		t.Errorf("expected Name '%s', got '%s'", dirName, project.Name)
	}
	if project.State != domain.StateActive {
		t.Errorf("expected State Active, got %v", project.State)
	}
	if !project.CreatedAt.IsZero() {
		// Good - timestamp should be set
	} else {
		t.Error("expected CreatedAt to be set")
	}
}

func TestAdd_SaveFailure(t *testing.T) {
	mock := NewMockRepository()
	mock.saveErr = errors.New("database connection failed")
	cli.SetRepository(mock)

	tmpDir := t.TempDir()

	_, err := executeAddCommand([]string{tmpDir})

	if err == nil {
		t.Fatal("expected error when save fails")
	}

	// Verify error message contains save context
	if !strings.Contains(err.Error(), "failed to save project") {
		t.Errorf("expected error to mention 'failed to save project', got: %v", err)
	}

	// Verify no project was saved
	if len(mock.projects) != 0 {
		t.Errorf("expected 0 projects saved on error, got %d", len(mock.projects))
	}
}

func TestAdd_WithDetectionService(t *testing.T) {
	repoMock := NewMockRepository()
	cli.SetRepository(repoMock)

	// Create mock detector with speckit result
	detectionResult := domain.NewDetectionResult(
		"speckit",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"plan.md found",
	)
	detectorMock := &MockDetector{
		detectResult: &detectionResult,
	}
	cli.SetDetectionService(detectorMock)
	defer cli.SetDetectionService(nil) // Reset after test

	tmpDir := t.TempDir()

	output, err := executeAddCommand([]string{tmpDir})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify detection result in output (Stage.String() capitalizes: "Plan", "Tasks", etc.)
	if !strings.Contains(output, "speckit") {
		t.Errorf("expected output to contain methodology 'speckit', got: %s", output)
	}
	if !strings.Contains(output, "Plan") {
		t.Errorf("expected output to contain stage 'Plan', got: %s", output)
	}

	// Verify project has detection fields populated
	for _, p := range repoMock.projects {
		if p.DetectedMethod != "speckit" {
			t.Errorf("expected DetectedMethod 'speckit', got '%s'", p.DetectedMethod)
		}
		if p.CurrentStage != domain.StagePlan {
			t.Errorf("expected CurrentStage StagePlan, got %v", p.CurrentStage)
		}
	}
}

func TestAdd_DetectionFailureIsNonFatal(t *testing.T) {
	repoMock := NewMockRepository()
	cli.SetRepository(repoMock)

	// Create mock detector that returns an error
	detectorMock := &MockDetector{
		detectErr: errors.New("detection failed"),
	}
	cli.SetDetectionService(detectorMock)
	defer cli.SetDetectionService(nil) // Reset after test

	tmpDir := t.TempDir()

	_, err := executeAddCommand([]string{tmpDir})

	// Detection failure should NOT cause add to fail
	if err != nil {
		t.Fatalf("expected no error (detection failure should be non-fatal), got: %v", err)
	}

	// Verify project was still saved
	if len(repoMock.projects) != 1 {
		t.Errorf("expected 1 project saved, got %d", len(repoMock.projects))
	}

	// Verify project has default detection values
	for _, p := range repoMock.projects {
		if p.DetectedMethod != "" {
			t.Errorf("expected empty DetectedMethod on detection failure, got '%s'", p.DetectedMethod)
		}
	}
}

func TestAdd_WithoutDetectionService(t *testing.T) {
	repoMock := NewMockRepository()
	cli.SetRepository(repoMock)
	cli.SetDetectionService(nil) // Explicitly nil

	tmpDir := t.TempDir()

	_, err := executeAddCommand([]string{tmpDir})

	// Should succeed without detection service
	if err != nil {
		t.Fatalf("expected no error without detection service, got: %v", err)
	}

	if len(repoMock.projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(repoMock.projects))
	}
}
