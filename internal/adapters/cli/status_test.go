package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// executeStatusCommand runs the status command with given args and returns output/error
func executeStatusCommand(args []string) (string, error) {
	cli.ResetStatusFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterStatusCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"status"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// ============================================================================
// Task 5.1: Test single project plain text output format
// ============================================================================

func TestStatus_PlainText_SingleProject(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/home/user/projects/client-alpha", "")
	p1.DetectedMethod = "speckit"
	p1.CurrentStage = domain.StagePlan
	p1.Confidence = domain.ConfidenceCertain
	p1.State = domain.StateActive
	p1.IsFavorite = false
	p1.Notes = "Waiting on API specs"
	p1.LastActivityAt = time.Now().Add(-2 * time.Hour)
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeStatusCommand([]string{"client-alpha"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// First line should be project name
	lines := strings.Split(output, "\n")
	if len(lines) == 0 || lines[0] != "client-alpha" {
		t.Errorf("expected first line 'client-alpha', got: %s", output)
	}

	// Check for indented key-value pairs
	if !strings.Contains(output, "  Path:        /home/user/projects/client-alpha") {
		t.Errorf("expected Path in output, got: %s", output)
	}
	if !strings.Contains(output, "  Method:") {
		t.Errorf("expected Method in output, got: %s", output)
	}
	if !strings.Contains(output, "  Stage:") {
		t.Errorf("expected Stage in output, got: %s", output)
	}
	if !strings.Contains(output, "  Confidence:") {
		t.Errorf("expected Confidence in output, got: %s", output)
	}
	if !strings.Contains(output, "  State:") {
		t.Errorf("expected State in output, got: %s", output)
	}
	if !strings.Contains(output, "  Favorite:    No") {
		t.Errorf("expected Favorite: No in output, got: %s", output)
	}
	if !strings.Contains(output, "  Notes:       Waiting on API specs") {
		t.Errorf("expected Notes in output, got: %s", output)
	}
	if !strings.Contains(output, "  Last Active:") {
		t.Errorf("expected Last Active in output, got: %s", output)
	}
}

func TestStatus_PlainText_WithDisplayName(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/home/user/projects/my-dir", "")
	p1.DisplayName = "My Cool App"
	p1.IsFavorite = true
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeStatusCommand([]string{"my-dir"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// First line should be display name
	lines := strings.Split(output, "\n")
	if len(lines) == 0 || lines[0] != "My Cool App" {
		t.Errorf("expected first line 'My Cool App', got: %s", lines[0])
	}

	// Favorite should be Yes
	if !strings.Contains(output, "  Favorite:    Yes") {
		t.Errorf("expected Favorite: Yes in output, got: %s", output)
	}
}

func TestStatus_PlainText_NoNotes(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/home/user/projects/test", "")
	p1.Notes = "" // No notes
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeStatusCommand([]string{"test"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Notes line should NOT appear when empty
	if strings.Contains(output, "Notes:") {
		t.Errorf("expected no Notes line when empty, got: %s", output)
	}
}

// ============================================================================
// Task 5.2: Test single project JSON output
// ============================================================================

func TestStatus_JSON_SingleProject(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/home/user/client-alpha", "")
	p1.DisplayName = ""
	p1.DetectedMethod = "speckit"
	p1.CurrentStage = domain.StagePlan
	p1.Confidence = domain.ConfidenceCertain
	p1.State = domain.StateActive
	p1.IsFavorite = false
	p1.Notes = "Waiting on API specs"
	p1.DetectionReasoning = "plan.md exists"
	p1.LastActivityAt = time.Now().Add(-45 * time.Minute)
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	// Set up mock waiting detector
	mockWD := &MockWaitingDetector{
		isWaiting:       true,
		waitingDuration: 45 * time.Minute,
	}
	cli.SetWaitingDetector(mockWD)
	defer cli.SetWaitingDetector(nil)

	output, err := executeStatusCommand([]string{"client-alpha", "--json"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Parse JSON - verify it's "project" (singular), not "projects" (array)
	var response struct {
		APIVersion string `json:"api_version"`
		Project    struct {
			Name                   string  `json:"name"`
			DisplayName            *string `json:"display_name"`
			Path                   string  `json:"path"`
			Method                 string  `json:"method"`
			Stage                  string  `json:"stage"`
			Confidence             string  `json:"confidence"`
			State                  string  `json:"state"`
			IsFavorite             bool    `json:"is_favorite"`
			IsWaiting              bool    `json:"is_waiting"`
			WaitingDurationMinutes *int    `json:"waiting_duration_minutes"`
			Notes                  *string `json:"notes"`
			DetectionReasoning     *string `json:"detection_reasoning"`
			LastActivityAt         string  `json:"last_activity_at"`
		} `json:"project"`
		Projects interface{} `json:"projects"` // Should be null/missing
	}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}

	// Verify api_version
	if response.APIVersion != "v1" {
		t.Errorf("expected api_version 'v1', got %s", response.APIVersion)
	}

	// Verify "projects" is NOT present (null or missing)
	if response.Projects != nil {
		t.Errorf("expected 'projects' to be null/missing for single project, got: %v", response.Projects)
	}

	// Verify project object fields
	proj := response.Project
	if proj.Name != "client-alpha" {
		t.Errorf("expected name 'client-alpha', got %s", proj.Name)
	}
	if proj.Path != "/home/user/client-alpha" {
		t.Errorf("expected path '/home/user/client-alpha', got %s", proj.Path)
	}
	if proj.Method != "speckit" {
		t.Errorf("expected method 'speckit', got %s", proj.Method)
	}
	if proj.Stage != "plan" {
		t.Errorf("expected lowercase stage 'plan', got %s", proj.Stage)
	}
	if proj.Confidence != "certain" {
		t.Errorf("expected lowercase confidence 'certain', got %s", proj.Confidence)
	}
	if proj.State != "active" {
		t.Errorf("expected lowercase state 'active', got %s", proj.State)
	}
	if proj.IsFavorite {
		t.Error("expected is_favorite to be false")
	}
	if !proj.IsWaiting {
		t.Error("expected is_waiting to be true")
	}
	if proj.WaitingDurationMinutes == nil || *proj.WaitingDurationMinutes != 45 {
		t.Errorf("expected waiting_duration_minutes 45, got %v", proj.WaitingDurationMinutes)
	}
	if proj.Notes == nil || *proj.Notes != "Waiting on API specs" {
		t.Errorf("expected notes 'Waiting on API specs', got %v", proj.Notes)
	}
	if proj.DetectionReasoning == nil || *proj.DetectionReasoning != "plan.md exists" {
		t.Errorf("expected detection_reasoning 'plan.md exists', got %v", proj.DetectionReasoning)
	}
	if proj.DisplayName != nil {
		t.Errorf("expected display_name to be null, got %v", proj.DisplayName)
	}

	// Verify ISO 8601 timestamp format
	_, err = time.Parse(time.RFC3339, proj.LastActivityAt)
	if err != nil {
		t.Errorf("expected ISO 8601 timestamp, got %s (parse error: %v)", proj.LastActivityAt, err)
	}
}

// ============================================================================
// Task 5.3: Test project not found returns exit code 2
// ============================================================================

func TestStatus_ProjectNotFound(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	output, err := executeStatusCommand([]string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error when project not found")
	}

	// Verify error message format
	if !strings.Contains(output, "✗ Project not found: nonexistent") {
		t.Errorf("expected '✗ Project not found: nonexistent', got: %s", output)
	}

	// Verify exit code mapping
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitProjectNotFound {
		t.Errorf("expected exit code %d (ExitProjectNotFound), got %d", cli.ExitProjectNotFound, exitCode)
	}
}

// ============================================================================
// Task 5.4: Test lookup by display_name
// ============================================================================

func TestStatus_LookupByDisplayName(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/home/user/projects/original-dir", "")
	p1.DisplayName = "My Cool App"
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeStatusCommand([]string{"My Cool App"})
	if err != nil {
		t.Fatalf("expected project found by display name, got error: %v", err)
	}

	// First line should be the display name
	lines := strings.Split(output, "\n")
	if len(lines) == 0 || lines[0] != "My Cool App" {
		t.Errorf("expected first line 'My Cool App', got: %s", lines[0])
	}
}

// ============================================================================
// Task 5.5: Test lookup by path
// ============================================================================

func TestStatus_LookupByPath(t *testing.T) {
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

	output, err := executeStatusCommand([]string{projectPath})
	if err != nil {
		t.Fatalf("expected project found by path, got error: %v", err)
	}

	// Should show project details
	if !strings.Contains(output, "myproject") {
		t.Errorf("expected project name in output, got: %s", output)
	}
}

// ============================================================================
// Task 5.6: Test --all flag produces same output as vibe list
// ============================================================================

func TestStatus_AllFlag_PlainText(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/path/to/bravo", "")
	p1.CurrentStage = domain.StageTasks
	mock.Projects[p1.Path] = p1

	p2, _ := domain.NewProject("/path/to/alpha", "")
	p2.CurrentStage = domain.StagePlan
	mock.Projects[p2.Path] = p2

	cli.SetRepository(mock)

	// Execute status --all
	statusOutput, statusErr := executeStatusCommand([]string{"--all"})
	if statusErr != nil {
		t.Fatalf("expected no error, got: %v", statusErr)
	}

	// Execute list (for comparison)
	cli.ResetListFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterListCommand(cmd)
	var listBuf bytes.Buffer
	cmd.SetOut(&listBuf)
	cmd.SetErr(&listBuf)
	cmd.SetArgs([]string{"list"})
	listErr := cmd.Execute()
	if listErr != nil {
		t.Fatalf("list command failed: %v", listErr)
	}
	listOutput := listBuf.String()

	// Should be same output
	if statusOutput != listOutput {
		t.Errorf("status --all output differs from list output\nstatus: %s\nlist: %s", statusOutput, listOutput)
	}
}

func TestStatus_AllFlag_JSON(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/path/to/test", "")
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeStatusCommand([]string{"--all", "--json"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should have "projects" array (like list), not "project" object
	var response struct {
		APIVersion string `json:"api_version"`
		Projects   []struct {
			Name string `json:"name"`
		} `json:"projects"`
	}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}

	if response.APIVersion != "v1" {
		t.Errorf("expected api_version 'v1', got %s", response.APIVersion)
	}
	if len(response.Projects) != 1 {
		t.Errorf("expected 1 project in array, got %d", len(response.Projects))
	}
	if response.Projects[0].Name != "test" {
		t.Errorf("expected project name 'test', got %s", response.Projects[0].Name)
	}
}

// ============================================================================
// Task 5.7: Test missing project name shows usage error
// ============================================================================

func TestStatus_MissingProjectName(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	_, err := executeStatusCommand([]string{})
	if err == nil {
		t.Fatal("expected error when no project name and no --all")
	}

	if !strings.Contains(err.Error(), "requires a project name or --all flag") {
		t.Errorf("expected 'requires a project name or --all flag' error, got: %v", err)
	}

	// Verify exit code is 1 (general error)
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitGeneralError {
		t.Errorf("expected exit code %d (ExitGeneralError), got %d", cli.ExitGeneralError, exitCode)
	}
}

// ============================================================================
// Task 5.8: Test invalid API version rejected
// ============================================================================

func TestStatus_InvalidAPIVersion(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/path/to/test", "")
	mock.Projects[p1.Path] = p1
	cli.SetRepository(mock)

	// v1 should be accepted
	_, err := executeStatusCommand([]string{"test", "--json", "--api-version=v1"})
	if err != nil {
		t.Fatalf("expected v1 to be accepted, got error: %v", err)
	}

	// v99 should be rejected with --json
	_, err = executeStatusCommand([]string{"test", "--json", "--api-version=v99"})
	if err == nil {
		t.Fatal("expected error for unsupported API version v99")
	}
	if !strings.Contains(err.Error(), "unsupported API version: v99") {
		t.Errorf("expected 'unsupported API version: v99' error, got: %v", err)
	}

	// v99 should also be rejected WITHOUT --json (api-version applies to all modes)
	_, err = executeStatusCommand([]string{"test", "--api-version=v99"})
	if err == nil {
		t.Fatal("expected error for unsupported API version v99 even without --json")
	}
	if !strings.Contains(err.Error(), "unsupported API version: v99") {
		t.Errorf("expected 'unsupported API version: v99' error, got: %v", err)
	}
}

// ============================================================================
// Additional edge case tests
// ============================================================================

func TestStatus_RepositoryNotInitialized(t *testing.T) {
	cli.SetRepository(nil)

	_, err := executeStatusCommand([]string{"test"})
	if err == nil {
		t.Fatal("expected error when repository is nil")
	}

	if !strings.Contains(err.Error(), "repository not initialized") {
		t.Errorf("expected 'repository not initialized' error, got: %v", err)
	}
}

func TestStatus_AllFlag_Empty(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	output, err := executeStatusCommand([]string{"--all"})
	if err != nil {
		t.Fatalf("expected no error for empty list, got: %v", err)
	}

	// Should show helpful message
	if !strings.Contains(output, "No projects tracked") {
		t.Errorf("expected empty message, got: %s", output)
	}
	// Story 13.3 AC4: Suggestion should use dynamic binary name
	if !strings.Contains(output, " add .'") {
		t.Errorf("expected 'add .' suggestion, got: %s", output)
	}
}

func TestStatus_AllFlag_JSON_Empty(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	output, err := executeStatusCommand([]string{"--all", "--json"})
	if err != nil {
		t.Fatalf("expected no error for empty JSON list, got: %v", err)
	}

	var response struct {
		APIVersion string        `json:"api_version"`
		Projects   []interface{} `json:"projects"`
	}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(response.Projects) != 0 {
		t.Errorf("expected empty projects array, got %d items", len(response.Projects))
	}
}

func TestStatus_LookupPriority_NameBeforeDisplayName(t *testing.T) {
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

	output, err := executeStatusCommand([]string{"alpha"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should find p1 (name match) before p2 (display_name match)
	// First line should be p1's display name "Some Other Name"
	lines := strings.Split(output, "\n")
	if len(lines) == 0 || lines[0] != "Some Other Name" {
		t.Errorf("expected first line 'Some Other Name' (name match priority), got: %s", lines[0])
	}
}

// Note: MockWaitingDetector is reused from list_test.go (same package)
// No need to duplicate it here.
