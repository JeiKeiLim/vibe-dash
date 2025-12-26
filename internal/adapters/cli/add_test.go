package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/testhelpers"
)

// executeAddCommand runs the add command with given args and returns output/error
func executeAddCommand(args []string) (string, error) {
	// Reset flags and root command for clean test
	cli.ResetAddFlags()
	return testhelpers.ExecuteCommand(cli.NewRootCmd, cli.RegisterAddCommand, "add", args)
}

func TestAdd_CurrentDirectory(t *testing.T) {
	// Setup mock
	mock := testhelpers.NewMockRepository()
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
	if len(mock.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(mock.Projects))
	}

	// Verify correct path was stored
	for path := range mock.Projects {
		// Path should be canonical (might differ due to symlinks)
		if !filepath.IsAbs(path) {
			t.Errorf("expected absolute path, got %s", path)
		}
	}
}

func TestAdd_AbsolutePath(t *testing.T) {
	mock := testhelpers.NewMockRepository()
	cli.SetRepository(mock)

	tmpDir := t.TempDir()

	_, err := executeAddCommand([]string{tmpDir})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(mock.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(mock.Projects))
	}
}

func TestAdd_WithNameFlag(t *testing.T) {
	mock := testhelpers.NewMockRepository()
	cli.SetRepository(mock)

	tmpDir := t.TempDir()

	_, err := executeAddCommand([]string{tmpDir, "--name", "Custom Name"})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify display name was set
	for _, p := range mock.Projects {
		if p.DisplayName != "Custom Name" {
			t.Errorf("expected DisplayName 'Custom Name', got '%s'", p.DisplayName)
		}
	}
}

func TestAdd_NonExistentPath(t *testing.T) {
	mock := testhelpers.NewMockRepository()
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
	mock := testhelpers.NewMockRepository()
	mock.Projects[canonicalPath] = &domain.Project{
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
	mock := testhelpers.NewMockRepository()
	mock.Projects[canonicalPath] = &domain.Project{
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
	mock := testhelpers.NewMockRepository()
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
	if len(mock.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(mock.Projects))
	}

	// Get canonical home path for comparison
	canonicalHome, err := filepath.EvalSymlinks(homeDir)
	if err != nil {
		t.Skipf("cannot get canonical home path: %v", err)
	}

	found := false
	for path := range mock.Projects {
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
	mock := testhelpers.NewMockRepository()
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

	if len(mock.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(mock.Projects))
	}
}

func TestAdd_VerifyProjectFields(t *testing.T) {
	mock := testhelpers.NewMockRepository()
	cli.SetRepository(mock)

	tmpDir := t.TempDir()
	dirName := filepath.Base(tmpDir)

	_, err := executeAddCommand([]string{tmpDir})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(mock.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(mock.Projects))
	}

	var project *domain.Project
	for _, p := range mock.Projects {
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
	mock := testhelpers.NewMockRepository()
	mock.SetSaveError(errors.New("database connection failed"))
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
	if len(mock.Projects) != 0 {
		t.Errorf("expected 0 projects saved on error, got %d", len(mock.Projects))
	}
}

func TestAdd_WithDetectionService(t *testing.T) {
	repoMock := testhelpers.NewMockRepository()
	cli.SetRepository(repoMock)

	// Create mock detector with speckit result
	detectionResult := domain.NewDetectionResult(
		"speckit",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"plan.md found",
	)
	detectorMock := testhelpers.NewMockDetector()
	detectorMock.SetResult(&detectionResult)
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
	for _, p := range repoMock.Projects {
		if p.DetectedMethod != "speckit" {
			t.Errorf("expected DetectedMethod 'speckit', got '%s'", p.DetectedMethod)
		}
		if p.CurrentStage != domain.StagePlan {
			t.Errorf("expected CurrentStage StagePlan, got %v", p.CurrentStage)
		}
	}
}

func TestAdd_DetectionFailureIsNonFatal(t *testing.T) {
	repoMock := testhelpers.NewMockRepository()
	cli.SetRepository(repoMock)

	// Create mock detector that returns an error
	detectorMock := testhelpers.NewMockDetector()
	detectorMock.SetError(errors.New("detection failed"))
	cli.SetDetectionService(detectorMock)
	defer cli.SetDetectionService(nil) // Reset after test

	tmpDir := t.TempDir()

	_, err := executeAddCommand([]string{tmpDir})

	// Detection failure should NOT cause add to fail
	if err != nil {
		t.Fatalf("expected no error (detection failure should be non-fatal), got: %v", err)
	}

	// Verify project was still saved
	if len(repoMock.Projects) != 1 {
		t.Errorf("expected 1 project saved, got %d", len(repoMock.Projects))
	}

	// Verify project has default detection values
	for _, p := range repoMock.Projects {
		if p.DetectedMethod != "" {
			t.Errorf("expected empty DetectedMethod on detection failure, got '%s'", p.DetectedMethod)
		}
	}
}

func TestAdd_WithoutDetectionService(t *testing.T) {
	repoMock := testhelpers.NewMockRepository()
	cli.SetRepository(repoMock)
	cli.SetDetectionService(nil) // Explicitly nil

	tmpDir := t.TempDir()

	_, err := executeAddCommand([]string{tmpDir})

	// Should succeed without detection service
	if err != nil {
		t.Fatalf("expected no error without detection service, got: %v", err)
	}

	if len(repoMock.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(repoMock.Projects))
	}
}

// ============================================================================
// Story 2.6: Project Name Collision Handling Tests
// ============================================================================

func TestAdd_NameCollision_Detected(t *testing.T) {
	// Create two temp directories with same base name
	baseDir := t.TempDir()
	clientADir := filepath.Join(baseDir, "client-a", "api-service")
	clientBDir := filepath.Join(baseDir, "client-b", "api-service")

	if err := os.MkdirAll(clientADir, 0755); err != nil {
		t.Fatalf("failed to create client-a dir: %v", err)
	}
	if err := os.MkdirAll(clientBDir, 0755); err != nil {
		t.Fatalf("failed to create client-b dir: %v", err)
	}

	// Get canonical paths
	canonicalA, _ := filepath.EvalSymlinks(clientADir)

	// Setup mock with existing project at client-a/api-service
	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonicalA, "")
	mock.Projects[canonicalA] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// Try to add client-b/api-service (same name "api-service")
	// Without --force and without stdin input, should prompt and error
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("")) // Empty input triggers EOF

	cmd.SetArgs([]string{"add", clientBDir})
	_ = cmd.Execute()

	// Should detect collision and prompt - with empty stdin should fail gracefully
	output := buf.String()
	if !strings.Contains(output, "already exists") {
		t.Errorf("expected collision message mentioning 'already exists', got: %s", output)
	}
}

func TestAdd_NameCollision_UserSelectsSuggested(t *testing.T) {
	// Create two temp directories with same base name
	baseDir := t.TempDir()
	clientADir := filepath.Join(baseDir, "client-a", "api-service")
	clientBDir := filepath.Join(baseDir, "client-b", "api-service")

	if err := os.MkdirAll(clientADir, 0755); err != nil {
		t.Fatalf("failed to create client-a dir: %v", err)
	}
	if err := os.MkdirAll(clientBDir, 0755); err != nil {
		t.Fatalf("failed to create client-b dir: %v", err)
	}

	// Get canonical paths
	canonicalA, _ := filepath.EvalSymlinks(clientADir)
	canonicalB, _ := filepath.EvalSymlinks(clientBDir)

	// Setup mock with existing project at client-a/api-service
	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonicalA, "")
	mock.Projects[canonicalA] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// User selects option 1 (suggested name)
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("1\n")) // User types "1" and hits enter

	cmd.SetArgs([]string{"add", clientBDir})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	// Should show success with disambiguated name
	if !strings.Contains(output, "Added") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify project saved with disambiguated name (parent-prefixed)
	if saved, ok := mock.Projects[canonicalB]; ok {
		if saved.DisplayName == "" || saved.DisplayName == "api-service" {
			t.Errorf("expected disambiguated DisplayName, got: %s", saved.DisplayName)
		}
		// Should contain parent directory
		if !strings.Contains(saved.DisplayName, "client-b") {
			t.Errorf("expected DisplayName to contain 'client-b', got: %s", saved.DisplayName)
		}
	} else {
		t.Error("expected project to be saved")
	}
}

func TestAdd_NameCollision_UserEntersCustomName(t *testing.T) {
	// Create two temp directories with same base name
	baseDir := t.TempDir()
	clientADir := filepath.Join(baseDir, "client-a", "api-service")
	clientBDir := filepath.Join(baseDir, "client-b", "api-service")

	if err := os.MkdirAll(clientADir, 0755); err != nil {
		t.Fatalf("failed to create client-a dir: %v", err)
	}
	if err := os.MkdirAll(clientBDir, 0755); err != nil {
		t.Fatalf("failed to create client-b dir: %v", err)
	}

	// Get canonical paths
	canonicalA, _ := filepath.EvalSymlinks(clientADir)
	canonicalB, _ := filepath.EvalSymlinks(clientBDir)

	// Setup mock with existing project at client-a/api-service
	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonicalA, "")
	mock.Projects[canonicalA] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// User selects option 2 (custom name) and enters "my-custom-api"
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("2\nmy-custom-api\n")) // Select 2, then enter custom name

	cmd.SetArgs([]string{"add", clientBDir})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify project saved with custom name
	if saved, ok := mock.Projects[canonicalB]; ok {
		if saved.DisplayName != "my-custom-api" {
			t.Errorf("expected DisplayName 'my-custom-api', got: %s", saved.DisplayName)
		}
	} else {
		t.Error("expected project to be saved")
	}
}

func TestAdd_NameCollision_ForceFlag(t *testing.T) {
	// Create two temp directories with same base name
	baseDir := t.TempDir()
	clientADir := filepath.Join(baseDir, "client-a", "api-service")
	clientBDir := filepath.Join(baseDir, "client-b", "api-service")

	if err := os.MkdirAll(clientADir, 0755); err != nil {
		t.Fatalf("failed to create client-a dir: %v", err)
	}
	if err := os.MkdirAll(clientBDir, 0755); err != nil {
		t.Fatalf("failed to create client-b dir: %v", err)
	}

	// Get canonical paths
	canonicalA, _ := filepath.EvalSymlinks(clientADir)
	canonicalB, _ := filepath.EvalSymlinks(clientBDir)

	// Setup mock with existing project at client-a/api-service
	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonicalA, "")
	mock.Projects[canonicalA] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// Use --force flag
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	// No stdin - --force should auto-resolve without prompting

	cmd.SetArgs([]string{"add", "--force", clientBDir})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error with --force, got: %v", err)
	}

	output := buf.String()
	// Should show success without prompting
	if !strings.Contains(output, "Added") {
		t.Errorf("expected success message, got: %s", output)
	}
	// Should NOT contain prompt text
	if strings.Contains(output, "Choose an option") {
		t.Errorf("--force should not show prompt, got: %s", output)
	}

	// Verify project saved with auto-disambiguated name
	if saved, ok := mock.Projects[canonicalB]; ok {
		if saved.DisplayName == "" || saved.DisplayName == "api-service" {
			t.Errorf("expected auto-disambiguated DisplayName, got: %s", saved.DisplayName)
		}
		// Should contain parent directory
		if !strings.Contains(saved.DisplayName, "client-b") {
			t.Errorf("expected DisplayName to contain 'client-b', got: %s", saved.DisplayName)
		}
	} else {
		t.Error("expected project to be saved")
	}
}

func TestAdd_NameCollision_CustomNameAlsoCollides(t *testing.T) {
	// Create three temp directories - two existing and one new
	baseDir := t.TempDir()
	existingDir1 := filepath.Join(baseDir, "client-a", "api-service")
	existingDir2 := filepath.Join(baseDir, "client-b", "my-api")
	newDir := filepath.Join(baseDir, "client-c", "api-service")

	if err := os.MkdirAll(existingDir1, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.MkdirAll(existingDir2, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	canonical1, _ := filepath.EvalSymlinks(existingDir1)
	canonical2, _ := filepath.EvalSymlinks(existingDir2)
	canonicalNew, _ := filepath.EvalSymlinks(newDir)

	// Setup mock with two existing projects
	mock := testhelpers.NewMockRepository()
	project1, _ := domain.NewProject(canonical1, "")
	mock.Projects[canonical1] = project1

	project2, _ := domain.NewProject(canonical2, "")
	project2.DisplayName = "my-api" // Also reserve "my-api"
	mock.Projects[canonical2] = project2

	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// User tries "my-api" first (collides), then "unique-name"
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("2\nmy-api\nunique-name\n")) // Select 2, try "my-api" (fails), then "unique-name"

	cmd.SetArgs([]string{"add", newDir})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	// Should show collision warning for "my-api"
	if !strings.Contains(output, "also exists") {
		t.Errorf("expected collision warning for custom name, got: %s", output)
	}

	// Verify project saved with unique-name
	if saved, ok := mock.Projects[canonicalNew]; ok {
		if saved.DisplayName != "unique-name" {
			t.Errorf("expected DisplayName 'unique-name', got: %s", saved.DisplayName)
		}
	} else {
		t.Error("expected project to be saved")
	}
}

func TestAdd_NameCollision_EmptyCustomName(t *testing.T) {
	// Create two temp directories with same base name
	baseDir := t.TempDir()
	clientADir := filepath.Join(baseDir, "client-a", "api-service")
	clientBDir := filepath.Join(baseDir, "client-b", "api-service")

	if err := os.MkdirAll(clientADir, 0755); err != nil {
		t.Fatalf("failed to create client-a dir: %v", err)
	}
	if err := os.MkdirAll(clientBDir, 0755); err != nil {
		t.Fatalf("failed to create client-b dir: %v", err)
	}

	canonicalA, _ := filepath.EvalSymlinks(clientADir)
	canonicalB, _ := filepath.EvalSymlinks(clientBDir)

	// Setup mock
	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonicalA, "")
	mock.Projects[canonicalA] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// User selects option 2, enters empty string, then valid name
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("2\n\nvalid-name\n")) // Select 2, empty (fails), then "valid-name"

	cmd.SetArgs([]string{"add", clientBDir})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	// Should show empty name warning
	if !strings.Contains(output, "cannot be empty") {
		t.Errorf("expected empty name warning, got: %s", output)
	}

	// Verify project saved with valid-name
	if saved, ok := mock.Projects[canonicalB]; ok {
		if saved.DisplayName != "valid-name" {
			t.Errorf("expected DisplayName 'valid-name', got: %s", saved.DisplayName)
		}
	} else {
		t.Error("expected project to be saved")
	}
}

func TestAdd_NameCollision_MultipleCollisionLevels(t *testing.T) {
	// Test grandparent prefix when parent collision also exists
	baseDir := t.TempDir()
	dir1 := filepath.Join(baseDir, "org-x", "client-a", "api-service")
	dir2 := filepath.Join(baseDir, "org-x", "client-b", "api-service")
	dir3 := filepath.Join(baseDir, "org-y", "client-a", "api-service") // New - name AND parent-name exist

	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.MkdirAll(dir3, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	canonical1, _ := filepath.EvalSymlinks(dir1)
	canonical2, _ := filepath.EvalSymlinks(dir2)
	canonical3, _ := filepath.EvalSymlinks(dir3)

	// Setup mock with existing projects
	mock := testhelpers.NewMockRepository()
	project1, _ := domain.NewProject(canonical1, "")
	mock.Projects[canonical1] = project1 // "api-service"

	project2, _ := domain.NewProject(canonical2, "")
	project2.DisplayName = "client-a-api-service" // Reserve parent-prefixed name
	mock.Projects[canonical2] = project2

	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// Use --force to test auto-disambiguation with grandparent
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.SetArgs([]string{"add", "--force", dir3})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify project saved - should have unique name (grandparent or timestamp)
	if saved, ok := mock.Projects[canonical3]; ok {
		if saved.DisplayName == "" || saved.DisplayName == "api-service" || saved.DisplayName == "client-a-api-service" {
			t.Errorf("expected unique disambiguated DisplayName, got: %s", saved.DisplayName)
		}
	} else {
		t.Error("expected project to be saved")
	}
}

func TestAdd_PathCollision_SymlinkToSameLocation(t *testing.T) {
	// Test AC6: symlink to same physical location should detect PATH collision (not name)
	baseDir := t.TempDir()
	projectDir := filepath.Join(baseDir, "project")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	symlinkPath := filepath.Join(baseDir, "link-to-project")
	if err := os.Symlink(projectDir, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	canonicalPath, _ := filepath.EvalSymlinks(projectDir)

	// Setup mock with project at canonical path
	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonicalPath, "")
	mock.Projects[canonicalPath] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// Try to add via symlink
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.SetArgs([]string{"add", symlinkPath})
	err := cmd.Execute()

	// Should return PATH collision error (not name collision prompt)
	if err == nil {
		t.Fatal("expected error for path collision via symlink")
	}

	// Verify error mentions "already" (exists/tracked)
	if !strings.Contains(err.Error(), "already") {
		t.Errorf("expected path collision error mentioning 'already', got: %v", err)
	}
}

func TestAdd_NoCollision_DifferentNames(t *testing.T) {
	// Test that projects with different names don't trigger collision
	baseDir := t.TempDir()
	dir1 := filepath.Join(baseDir, "project-a")
	dir2 := filepath.Join(baseDir, "project-b") // Different name

	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	canonical1, _ := filepath.EvalSymlinks(dir1)
	canonical2, _ := filepath.EvalSymlinks(dir2)

	// Setup mock with existing project
	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonical1, "")
	mock.Projects[canonical1] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// Add project-b (different name, no collision)
	output, err := executeAddCommand([]string{dir2})

	if err != nil {
		t.Fatalf("expected no error for different names, got: %v", err)
	}

	// Should succeed without prompting
	if strings.Contains(output, "already exists") {
		t.Errorf("should not trigger collision for different names, got: %s", output)
	}

	// Verify both projects exist
	if len(mock.Projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(mock.Projects))
	}

	// Verify new project has no DisplayName (since no collision)
	if saved, ok := mock.Projects[canonical2]; ok {
		if saved.DisplayName != "" {
			t.Errorf("expected empty DisplayName for no collision, got: %s", saved.DisplayName)
		}
	} else {
		t.Error("expected project to be saved")
	}
}

func TestAdd_AC5_DisplayNameShownInOutput(t *testing.T) {
	// Test AC5: "Given project is displayed, Then display_name is shown if set, otherwise name"
	baseDir := t.TempDir()
	clientADir := filepath.Join(baseDir, "client-a", "api-service")
	clientBDir := filepath.Join(baseDir, "client-b", "api-service")

	if err := os.MkdirAll(clientADir, 0755); err != nil {
		t.Fatalf("failed to create client-a dir: %v", err)
	}
	if err := os.MkdirAll(clientBDir, 0755); err != nil {
		t.Fatalf("failed to create client-b dir: %v", err)
	}

	canonicalA, _ := filepath.EvalSymlinks(clientADir)
	canonicalB, _ := filepath.EvalSymlinks(clientBDir)

	// Setup mock with existing project
	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonicalA, "")
	mock.Projects[canonicalA] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// Test 1: With DisplayName set (via collision resolution)
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("1\n")) // Select suggested name

	cmd.SetArgs([]string{"add", clientBDir})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	// Verify DisplayName is shown in output (should contain parent prefix)
	if !strings.Contains(output, "client-b") {
		t.Errorf("AC5 violation: output should show DisplayName with 'client-b', got: %s", output)
	}

	// Verify project was saved at correct path
	if _, ok := mock.Projects[canonicalB]; !ok {
		t.Error("expected project to be saved at canonicalB path")
	}

	// Test 2: Without DisplayName (no collision)
	noCollisionDir := filepath.Join(baseDir, "unique-project")
	if err := os.Mkdir(noCollisionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	canonicalUnique, _ := filepath.EvalSymlinks(noCollisionDir)

	cli.ResetAddFlags()
	cmd2 := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd2)

	var buf2 bytes.Buffer
	cmd2.SetOut(&buf2)
	cmd2.SetErr(&buf2)

	cmd2.SetArgs([]string{"add", noCollisionDir})
	err = cmd2.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output2 := buf2.String()
	// Verify Name is shown when no DisplayName
	if !strings.Contains(output2, "unique-project") {
		t.Errorf("AC5 violation: output should show Name 'unique-project', got: %s", output2)
	}

	// Verify the saved project has no DisplayName
	if saved, ok := mock.Projects[canonicalUnique]; ok {
		if saved.DisplayName != "" {
			t.Errorf("expected empty DisplayName for non-collision project, got: %s", saved.DisplayName)
		}
	}
}

func TestAdd_InvalidChoiceReprompts(t *testing.T) {
	// Test M2 fix: Invalid choice should re-prompt instead of error
	baseDir := t.TempDir()
	clientADir := filepath.Join(baseDir, "client-a", "api-service")
	clientBDir := filepath.Join(baseDir, "client-b", "api-service")

	if err := os.MkdirAll(clientADir, 0755); err != nil {
		t.Fatalf("failed to create client-a dir: %v", err)
	}
	if err := os.MkdirAll(clientBDir, 0755); err != nil {
		t.Fatalf("failed to create client-b dir: %v", err)
	}

	canonicalA, _ := filepath.EvalSymlinks(clientADir)
	canonicalB, _ := filepath.EvalSymlinks(clientBDir)

	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonicalA, "")
	mock.Projects[canonicalA] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// User enters invalid "3", then "x", then valid "1"
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("3\nx\n1\n")) // Invalid, invalid, then valid

	cmd.SetArgs([]string{"add", clientBDir})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error after re-prompting, got: %v", err)
	}

	output := buf.String()
	// Should show invalid choice message
	if !strings.Contains(output, "Invalid choice") {
		t.Errorf("expected invalid choice warning, got: %s", output)
	}

	// Should ultimately succeed
	if !strings.Contains(output, "Added") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify project saved
	if _, ok := mock.Projects[canonicalB]; !ok {
		t.Error("expected project to be saved after re-prompting")
	}
}

func TestAdd_DisplayNameCollision(t *testing.T) {
	// Test that collision also checks DisplayName field
	baseDir := t.TempDir()
	dir1 := filepath.Join(baseDir, "real-project")
	dir2 := filepath.Join(baseDir, "other-project")

	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	canonical1, _ := filepath.EvalSymlinks(dir1)
	canonical2, _ := filepath.EvalSymlinks(dir2)

	// Setup mock with existing project that has a DisplayName
	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonical1, "")
	existingProject.DisplayName = "my-display-name" // Set DisplayName different from directory name
	mock.Projects[canonical1] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	// Try to add other-project with --name matching existing DisplayName
	cli.ResetAddFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("1\n")) // Select suggested name

	cmd.SetArgs([]string{"add", "--name", "my-display-name", dir2})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	// Should detect collision with DisplayName
	if !strings.Contains(output, "already exists") {
		t.Errorf("expected collision with DisplayName, got: %s", output)
	}

	// Verify project saved with different name
	if saved, ok := mock.Projects[canonical2]; ok {
		if saved.DisplayName == "my-display-name" {
			t.Errorf("DisplayName should have been disambiguated, got: %s", saved.DisplayName)
		}
	} else {
		t.Error("expected project to be saved")
	}
}

// ============================================================================
// Story 6.7: Quiet Mode Tests
// ============================================================================

func TestAdd_QuietMode_SuppressesOutput(t *testing.T) {
	mock := testhelpers.NewMockRepository()
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	tmpDir := t.TempDir()

	// Execute add command with --quiet
	cli.ResetAddFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true) // Set quiet mode
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"add", tmpDir})

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify output is empty (quiet mode)
	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output with --quiet, got: %s", output)
	}

	// Verify project was still saved
	if len(mock.Projects) != 1 {
		t.Errorf("expected 1 project saved, got %d", len(mock.Projects))
	}
}

func TestAdd_QuietMode_GlobalFlagPosition(t *testing.T) {
	// Test AC8: vibe -q add . (global flag BEFORE subcommand)
	mock := testhelpers.NewMockRepository()
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	tmpDir := t.TempDir()

	cli.ResetAddFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true) // Simulates -q flag before subcommand
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"add", tmpDir})

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("AC8: global -q before subcommand should suppress output, got: %s", output)
	}
}

func TestAdd_QuietMode_WithForce(t *testing.T) {
	// Test quiet mode combined with force flag on name collision
	baseDir := t.TempDir()
	clientADir := filepath.Join(baseDir, "client-a", "api-service")
	clientBDir := filepath.Join(baseDir, "client-b", "api-service")

	if err := os.MkdirAll(clientADir, 0755); err != nil {
		t.Fatalf("failed to create client-a dir: %v", err)
	}
	if err := os.MkdirAll(clientBDir, 0755); err != nil {
		t.Fatalf("failed to create client-b dir: %v", err)
	}

	canonicalA, _ := filepath.EvalSymlinks(clientADir)
	canonicalB, _ := filepath.EvalSymlinks(clientBDir)

	mock := testhelpers.NewMockRepository()
	existingProject, _ := domain.NewProject(canonicalA, "")
	mock.Projects[canonicalA] = existingProject
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	cli.ResetAddFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true) // Set quiet mode
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.SetArgs([]string{"add", "--force", clientBDir})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error with --quiet --force, got: %v", err)
	}

	output := buf.String()
	// Should have no output at all
	if output != "" {
		t.Errorf("expected empty output with --quiet --force, got: %s", output)
	}

	// Verify project was saved with disambiguated name
	if saved, ok := mock.Projects[canonicalB]; ok {
		if saved.DisplayName == "" || saved.DisplayName == "api-service" {
			t.Errorf("expected auto-disambiguated DisplayName, got: %s", saved.DisplayName)
		}
	} else {
		t.Error("expected project to be saved")
	}
}

func TestAdd_QuietMode_ErrorsStillReturned(t *testing.T) {
	// AC7: Errors are still returned even in quiet mode
	mock := testhelpers.NewMockRepository()
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	cli.ResetAddFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"add", "/nonexistent/path"})

	err := cmd.Execute()

	// Error should still be returned
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}

	// Exit code should be non-zero
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode == cli.ExitSuccess {
		t.Errorf("expected non-zero exit code, got %d", exitCode)
	}
}

// ============================================================================
// Story 7.4: Progress Indicators Tests
// ============================================================================

func TestAdd_ShowsDetectionProgress(t *testing.T) {
	// AC4: "Detecting methodology..." is shown before detection
	repoMock := testhelpers.NewMockRepository()
	cli.SetRepository(repoMock)

	// Create mock detector with result
	detectionResult := domain.NewDetectionResult(
		"bmad",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"bmad config found",
	)
	detectorMock := testhelpers.NewMockDetector()
	detectorMock.SetResult(&detectionResult)
	cli.SetDetectionService(detectorMock)
	defer cli.SetDetectionService(nil)

	tmpDir := t.TempDir()

	// Ensure not in quiet mode
	cli.ResetQuietFlag()
	cli.SetQuietForTest(false)
	defer cli.ResetQuietFlag()

	output, err := executeAddCommand([]string{tmpDir})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should show detection progress message
	if !strings.Contains(output, "Detecting methodology...") {
		t.Errorf("expected 'Detecting methodology...' in output, got: %s", output)
	}

	// Should also show success message
	if !strings.Contains(output, "Added") {
		t.Errorf("expected 'Added' in output, got: %s", output)
	}
}

func TestAdd_NoDetectionProgress_QuietMode(t *testing.T) {
	// AC4: Detection progress is suppressed in quiet mode
	repoMock := testhelpers.NewMockRepository()
	cli.SetRepository(repoMock)

	// Create mock detector with result
	detectionResult := domain.NewDetectionResult(
		"bmad",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"bmad config found",
	)
	detectorMock := testhelpers.NewMockDetector()
	detectorMock.SetResult(&detectionResult)
	cli.SetDetectionService(detectorMock)
	defer cli.SetDetectionService(nil)

	tmpDir := t.TempDir()

	// Set quiet mode
	cli.ResetAddFlags()
	cli.ResetQuietFlag()
	cli.SetQuietForTest(true)
	defer cli.ResetQuietFlag()

	cmd := cli.NewRootCmd()
	cli.RegisterAddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"add", tmpDir})

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	// Should NOT show detection progress in quiet mode
	if strings.Contains(output, "Detecting methodology...") {
		t.Errorf("expected no 'Detecting methodology...' in quiet mode, got: %s", output)
	}

	// Should have no output at all in quiet mode
	if output != "" {
		t.Errorf("expected empty output in quiet mode, got: %s", output)
	}

	// Verify project was still saved with detection result
	if len(repoMock.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(repoMock.Projects))
	}
	for _, p := range repoMock.Projects {
		if p.DetectedMethod != "bmad" {
			t.Errorf("expected DetectedMethod 'bmad', got '%s'", p.DetectedMethod)
		}
	}
}

func TestAdd_NoDetectionProgress_WithoutDetector(t *testing.T) {
	// When no detection service is set, no detection progress should be shown
	repoMock := testhelpers.NewMockRepository()
	cli.SetRepository(repoMock)
	cli.SetDetectionService(nil) // No detection service

	tmpDir := t.TempDir()

	// Ensure not in quiet mode
	cli.ResetQuietFlag()
	cli.SetQuietForTest(false)
	defer cli.ResetQuietFlag()

	output, err := executeAddCommand([]string{tmpDir})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should NOT show detection progress when detector is nil
	if strings.Contains(output, "Detecting methodology...") {
		t.Errorf("expected no 'Detecting methodology...' without detector, got: %s", output)
	}

	// Should still show success message
	if !strings.Contains(output, "Added") {
		t.Errorf("expected 'Added' in output, got: %s", output)
	}
}
