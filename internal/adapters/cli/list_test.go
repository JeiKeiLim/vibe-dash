package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// executeListCommand runs the list command with given args and returns output/error
func executeListCommand(args []string) (string, error) {
	cli.ResetListFlags()
	cmd := cli.NewRootCmd()
	cli.RegisterListCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	fullArgs := append([]string{"list"}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}

// ============================================================================
// Task 1: Basic list command tests (AC1)
// ============================================================================

func TestList_PlainText_MultipleProjects(t *testing.T) {
	mock := NewMockRepository()

	// Create projects with different stages and times
	p1, _ := domain.NewProject("/path/to/bravo", "")
	p1.CurrentStage = domain.StageTasks
	p1.LastActivityAt = time.Now().Add(-2 * time.Hour)
	mock.projects[p1.Path] = p1

	p2, _ := domain.NewProject("/path/to/alpha", "")
	p2.CurrentStage = domain.StagePlan
	p2.LastActivityAt = time.Now().Add(-5 * time.Minute)
	mock.projects[p2.Path] = p2

	cli.SetRepository(mock)

	output, err := executeListCommand([]string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should have header
	if !strings.Contains(output, "PROJECT") {
		t.Errorf("expected header with PROJECT column, got: %s", output)
	}
	if !strings.Contains(output, "STAGE") {
		t.Errorf("expected header with STAGE column, got: %s", output)
	}
	if !strings.Contains(output, "LAST ACTIVE") {
		t.Errorf("expected header with LAST ACTIVE column, got: %s", output)
	}

	// Should contain project names
	if !strings.Contains(output, "alpha") {
		t.Errorf("expected output to contain 'alpha', got: %s", output)
	}
	if !strings.Contains(output, "bravo") {
		t.Errorf("expected output to contain 'bravo', got: %s", output)
	}

	// Should contain stages
	if !strings.Contains(output, "Plan") {
		t.Errorf("expected output to contain 'Plan', got: %s", output)
	}
	if !strings.Contains(output, "Tasks") {
		t.Errorf("expected output to contain 'Tasks', got: %s", output)
	}
}

func TestList_PlainText_RelativeTime(t *testing.T) {
	mock := NewMockRepository()

	// Test different time ranges
	p1, _ := domain.NewProject("/path/to/project1", "")
	p1.LastActivityAt = time.Now().Add(-30 * time.Second) // just now
	mock.projects[p1.Path] = p1

	p2, _ := domain.NewProject("/path/to/project2", "")
	p2.LastActivityAt = time.Now().Add(-5 * time.Minute) // 5m ago
	mock.projects[p2.Path] = p2

	p3, _ := domain.NewProject("/path/to/project3", "")
	p3.LastActivityAt = time.Now().Add(-3 * time.Hour) // 3h ago
	mock.projects[p3.Path] = p3

	p4, _ := domain.NewProject("/path/to/project4", "")
	p4.LastActivityAt = time.Now().Add(-48 * time.Hour) // 2d ago
	mock.projects[p4.Path] = p4

	cli.SetRepository(mock)

	output, err := executeListCommand([]string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check for relative time patterns
	if !strings.Contains(output, "just now") {
		t.Errorf("expected 'just now' for recent activity, got: %s", output)
	}
	if !strings.Contains(output, "5m ago") {
		t.Errorf("expected '5m ago' for minute activity, got: %s", output)
	}
	if !strings.Contains(output, "3h ago") {
		t.Errorf("expected '3h ago' for hour activity, got: %s", output)
	}
	if !strings.Contains(output, "2d ago") {
		t.Errorf("expected '2d ago' for day activity, got: %s", output)
	}
}

// ============================================================================
// Task 2: JSON output tests (AC2)
// ============================================================================

func TestList_JSON_Structure(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/path/to/test", "")
	p1.CurrentStage = domain.StagePlan
	p1.State = domain.StateActive
	p1.DetectedMethod = "speckit"
	mock.projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeListCommand([]string{"--json"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var response struct {
		APIVersion string `json:"api_version"`
		Projects   []struct {
			Name           string  `json:"name"`
			DisplayName    *string `json:"display_name"`
			Path           string  `json:"path"`
			Method         string  `json:"method"`
			Stage          string  `json:"stage"`
			Confidence     string  `json:"confidence"`
			State          string  `json:"state"`
			IsFavorite     bool    `json:"is_favorite"`
			LastActivityAt string  `json:"last_activity_at"`
		} `json:"projects"`
	}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if response.APIVersion != "v1" {
		t.Errorf("expected api_version 'v1', got %s", response.APIVersion)
	}
	if len(response.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(response.Projects))
	}

	proj := response.Projects[0]

	// Verify lowercase stage per Architecture spec
	if proj.Stage != "plan" {
		t.Errorf("expected lowercase stage 'plan', got %s", proj.Stage)
	}

	// Verify lowercase state per Architecture spec
	if proj.State != "active" {
		t.Errorf("expected lowercase state 'active', got %s", proj.State)
	}

	// Verify display_name is null when not set
	if proj.DisplayName != nil {
		t.Errorf("expected display_name to be null, got %v", proj.DisplayName)
	}

	// Verify ISO 8601 timestamp format
	_, err = time.Parse(time.RFC3339, proj.LastActivityAt)
	if err != nil {
		t.Errorf("expected ISO 8601 timestamp, got %s (parse error: %v)", proj.LastActivityAt, err)
	}

	// Verify confidence default
	if proj.Confidence != "uncertain" {
		t.Errorf("expected confidence 'uncertain', got %s", proj.Confidence)
	}
}

func TestList_JSON_WithDisplayName(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/path/to/test", "")
	p1.DisplayName = "Custom Display"
	mock.projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeListCommand([]string{"--json"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var response struct {
		Projects []struct {
			DisplayName *string `json:"display_name"`
		} `json:"projects"`
	}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if response.Projects[0].DisplayName == nil {
		t.Error("expected display_name to be set")
	} else if *response.Projects[0].DisplayName != "Custom Display" {
		t.Errorf("expected display_name 'Custom Display', got %s", *response.Projects[0].DisplayName)
	}
}

// ============================================================================
// Task 3: Empty list handling tests (AC3, AC4)
// ============================================================================

func TestList_EmptyList_PlainText(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	output, err := executeListCommand([]string{})
	if err != nil {
		t.Fatalf("expected no error for empty list, got: %v", err)
	}

	// AC3: Plain text shows helpful message
	if !strings.Contains(output, "No projects tracked") {
		t.Errorf("expected empty message, got: %s", output)
	}
	if !strings.Contains(output, "vibe add") {
		t.Errorf("expected add command hint, got: %s", output)
	}
}

func TestList_EmptyList_JSON(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)

	output, err := executeListCommand([]string{"--json"})
	if err != nil {
		t.Fatalf("expected no error for empty JSON list, got: %v", err)
	}

	// AC4: JSON shows empty array
	var response struct {
		APIVersion string        `json:"api_version"`
		Projects   []interface{} `json:"projects"`
	}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if response.APIVersion != "v1" {
		t.Errorf("expected api_version 'v1', got %s", response.APIVersion)
	}
	if len(response.Projects) != 0 {
		t.Errorf("expected empty projects array, got %d items", len(response.Projects))
	}
}

// ============================================================================
// Task 4: Sorting tests (AC5)
// ============================================================================

func TestList_SortedAlphabetically(t *testing.T) {
	mock := NewMockRepository()

	// Create projects in non-alphabetical order
	p1, _ := domain.NewProject("/path/to/zebra", "")
	mock.projects[p1.Path] = p1

	p2, _ := domain.NewProject("/path/to/alpha", "")
	mock.projects[p2.Path] = p2

	p3, _ := domain.NewProject("/path/to/middle", "")
	mock.projects[p3.Path] = p3

	cli.SetRepository(mock)

	output, err := executeListCommand([]string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check order: alpha should appear before middle, middle before zebra
	alphaIdx := strings.Index(output, "alpha")
	middleIdx := strings.Index(output, "middle")
	zebraIdx := strings.Index(output, "zebra")

	if alphaIdx > middleIdx {
		t.Errorf("expected 'alpha' before 'middle' in output")
	}
	if middleIdx > zebraIdx {
		t.Errorf("expected 'middle' before 'zebra' in output")
	}
}

func TestList_SortedByEffectiveName(t *testing.T) {
	mock := NewMockRepository()

	// p1: Name=zebra, DisplayName="aaa" -> effective name = "aaa"
	p1, _ := domain.NewProject("/path/to/zebra", "")
	p1.DisplayName = "aaa"
	mock.projects[p1.Path] = p1

	// p2: Name=alpha, no DisplayName -> effective name = "alpha"
	p2, _ := domain.NewProject("/path/to/alpha", "")
	mock.projects[p2.Path] = p2

	cli.SetRepository(mock)

	output, err := executeListCommand([]string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// "aaa" should appear before "alpha" (sorted by effective name)
	aaaIdx := strings.Index(output, "aaa")
	alphaIdx := strings.Index(output, "alpha")

	if aaaIdx > alphaIdx {
		t.Errorf("expected 'aaa' (display name) before 'alpha' in sorted output")
	}
}

func TestList_SortedCaseInsensitive(t *testing.T) {
	mock := NewMockRepository()

	p1, _ := domain.NewProject("/path/to/Zebra", "")
	mock.projects[p1.Path] = p1

	p2, _ := domain.NewProject("/path/to/alpha", "")
	mock.projects[p2.Path] = p2

	cli.SetRepository(mock)

	output, err := executeListCommand([]string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// "alpha" should appear before "Zebra" (case-insensitive)
	alphaIdx := strings.Index(output, "alpha")
	zebraIdx := strings.Index(output, "Zebra")

	if alphaIdx > zebraIdx {
		t.Errorf("expected 'alpha' before 'Zebra' (case-insensitive sort)")
	}
}

func TestList_JSON_SortedAlphabetically(t *testing.T) {
	mock := NewMockRepository()

	// Create projects in non-alphabetical order
	p1, _ := domain.NewProject("/path/to/zebra", "")
	mock.projects[p1.Path] = p1

	p2, _ := domain.NewProject("/path/to/alpha", "")
	mock.projects[p2.Path] = p2

	p3, _ := domain.NewProject("/path/to/middle", "")
	mock.projects[p3.Path] = p3

	cli.SetRepository(mock)

	output, err := executeListCommand([]string{"--json"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var response struct {
		Projects []struct {
			Name string `json:"name"`
		} `json:"projects"`
	}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(response.Projects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(response.Projects))
	}

	// Verify sort order: alpha, middle, zebra
	expectedOrder := []string{"alpha", "middle", "zebra"}
	for i, expected := range expectedOrder {
		if response.Projects[i].Name != expected {
			t.Errorf("expected project[%d] to be '%s', got '%s'", i, expected, response.Projects[i].Name)
		}
	}
}

func TestList_JSON_SortedByEffectiveName(t *testing.T) {
	mock := NewMockRepository()

	// p1: Name=zebra, DisplayName="aaa" -> effective name = "aaa"
	p1, _ := domain.NewProject("/path/to/zebra", "")
	p1.DisplayName = "aaa"
	mock.projects[p1.Path] = p1

	// p2: Name=alpha, no DisplayName -> effective name = "alpha"
	p2, _ := domain.NewProject("/path/to/alpha", "")
	mock.projects[p2.Path] = p2

	// p3: Name=beta, DisplayName="zzz" -> effective name = "zzz"
	p3, _ := domain.NewProject("/path/to/beta", "")
	p3.DisplayName = "zzz"
	mock.projects[p3.Path] = p3

	cli.SetRepository(mock)

	output, err := executeListCommand([]string{"--json"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var response struct {
		Projects []struct {
			Name        string  `json:"name"`
			DisplayName *string `json:"display_name"`
		} `json:"projects"`
	}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(response.Projects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(response.Projects))
	}

	// Verify sort by effective name: aaa (zebra), alpha, zzz (beta)
	// First project: zebra with DisplayName "aaa"
	if response.Projects[0].Name != "zebra" || response.Projects[0].DisplayName == nil || *response.Projects[0].DisplayName != "aaa" {
		t.Errorf("expected first project to be zebra with display_name 'aaa', got: name=%s, display_name=%v",
			response.Projects[0].Name, response.Projects[0].DisplayName)
	}

	// Second project: alpha (no DisplayName)
	if response.Projects[1].Name != "alpha" {
		t.Errorf("expected second project to be 'alpha', got: %s", response.Projects[1].Name)
	}

	// Third project: beta with DisplayName "zzz"
	if response.Projects[2].Name != "beta" || response.Projects[2].DisplayName == nil || *response.Projects[2].DisplayName != "zzz" {
		t.Errorf("expected third project to be beta with display_name 'zzz', got: name=%s, display_name=%v",
			response.Projects[2].Name, response.Projects[2].DisplayName)
	}
}

// ============================================================================
// Task 5: Additional tests
// ============================================================================

func TestList_DisplayNameShownWhenSet(t *testing.T) {
	mock := NewMockRepository()
	p1, _ := domain.NewProject("/path/to/original-dir-name", "")
	p1.DisplayName = "Custom Name"
	mock.projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeListCommand([]string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// AC5: display_name shown when set
	if !strings.Contains(output, "Custom Name") {
		t.Errorf("expected DisplayName 'Custom Name' in output, got: %s", output)
	}
	// Original dir name should NOT appear in PROJECT column
	// (it might appear in path column if we show it, but not as project name)
}

func TestList_BothActiveAndHibernated(t *testing.T) {
	mock := NewMockRepository()

	p1, _ := domain.NewProject("/path/to/active-project", "")
	p1.State = domain.StateActive
	mock.projects[p1.Path] = p1

	p2, _ := domain.NewProject("/path/to/hibernated-project", "")
	p2.State = domain.StateHibernated
	mock.projects[p2.Path] = p2

	cli.SetRepository(mock)

	output, err := executeListCommand([]string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// AC5: ALL projects shown regardless of state
	if !strings.Contains(output, "active-project") {
		t.Errorf("expected active project in output, got: %s", output)
	}
	if !strings.Contains(output, "hibernated-project") {
		t.Errorf("expected hibernated project in output, got: %s", output)
	}
}

func TestList_RepositoryError(t *testing.T) {
	mock := &MockRepositoryWithFindAllError{
		findAllErr: errors.New("database connection failed"),
	}
	cli.SetRepository(mock)

	_, err := executeListCommand([]string{})
	if err == nil {
		t.Fatal("expected error when repository fails")
	}

	// Verify exit code mapping
	exitCode := cli.MapErrorToExitCode(err)
	if exitCode != cli.ExitGeneralError {
		t.Errorf("expected exit code %d, got %d", cli.ExitGeneralError, exitCode)
	}
}

func TestList_RepositoryNotInitialized(t *testing.T) {
	cli.SetRepository(nil)

	_, err := executeListCommand([]string{})
	if err == nil {
		t.Fatal("expected error when repository is nil")
	}

	if !strings.Contains(err.Error(), "repository not initialized") {
		t.Errorf("expected 'repository not initialized' error, got: %v", err)
	}
}

func TestList_LongProjectNameTruncated(t *testing.T) {
	mock := NewMockRepository()

	// Create project with very long name (>40 chars)
	longName := "this-is-a-very-long-project-name-that-exceeds-forty-characters-limit"
	p1, _ := domain.NewProject("/path/to/"+longName, "")
	mock.projects[p1.Path] = p1
	cli.SetRepository(mock)

	output, err := executeListCommand([]string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should be truncated with "..."
	if !strings.Contains(output, "...") {
		t.Errorf("expected truncated name with '...', got: %s", output)
	}
}

// MockRepositoryWithFindAllError implements ports.ProjectRepository with FindAll error
type MockRepositoryWithFindAllError struct {
	findAllErr error
}

func (m *MockRepositoryWithFindAllError) Save(_ context.Context, _ *domain.Project) error {
	return nil
}

func (m *MockRepositoryWithFindAllError) FindByID(_ context.Context, _ string) (*domain.Project, error) {
	return nil, domain.ErrProjectNotFound
}

func (m *MockRepositoryWithFindAllError) FindByPath(_ context.Context, _ string) (*domain.Project, error) {
	return nil, domain.ErrProjectNotFound
}

func (m *MockRepositoryWithFindAllError) FindAll(_ context.Context) ([]*domain.Project, error) {
	return nil, m.findAllErr
}

func (m *MockRepositoryWithFindAllError) FindActive(_ context.Context) ([]*domain.Project, error) {
	return nil, m.findAllErr
}

func (m *MockRepositoryWithFindAllError) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	return nil, m.findAllErr
}

func (m *MockRepositoryWithFindAllError) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *MockRepositoryWithFindAllError) UpdateState(_ context.Context, _ string, _ domain.ProjectState) error {
	return nil
}

func (m *MockRepositoryWithFindAllError) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}
