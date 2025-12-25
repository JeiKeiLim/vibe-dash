package cli_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// mockResetRepository implements ports.ProjectRepository with Reset methods for testing
type mockResetRepository struct {
	resetProjectCalls []string
	resetAllCount     int
	resetProjectErr   error
	resetAllErr       error
	resetAllResult    int
}

func newMockResetRepository() *mockResetRepository {
	return &mockResetRepository{
		resetProjectCalls: make([]string, 0),
		resetAllResult:    3, // Default: simulate 3 projects reset
	}
}

func (m *mockResetRepository) Save(_ context.Context, _ *domain.Project) error { return nil }
func (m *mockResetRepository) FindByID(_ context.Context, _ string) (*domain.Project, error) {
	return nil, domain.ErrProjectNotFound
}
func (m *mockResetRepository) FindByPath(_ context.Context, _ string) (*domain.Project, error) {
	return nil, domain.ErrProjectNotFound
}
func (m *mockResetRepository) FindAll(_ context.Context) ([]*domain.Project, error) {
	return []*domain.Project{}, nil
}
func (m *mockResetRepository) FindActive(_ context.Context) ([]*domain.Project, error) {
	return []*domain.Project{}, nil
}
func (m *mockResetRepository) FindHibernated(_ context.Context) ([]*domain.Project, error) {
	return []*domain.Project{}, nil
}
func (m *mockResetRepository) Delete(_ context.Context, _ string) error { return nil }
func (m *mockResetRepository) UpdateState(_ context.Context, _ string, _ domain.ProjectState) error {
	return nil
}
func (m *mockResetRepository) UpdateLastActivity(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (m *mockResetRepository) ResetProject(_ context.Context, projectID string) error {
	m.resetProjectCalls = append(m.resetProjectCalls, projectID)
	return m.resetProjectErr
}
func (m *mockResetRepository) ResetAll(_ context.Context) (int, error) {
	m.resetAllCount++
	return m.resetAllResult, m.resetAllErr
}

func TestResetCommand_RequiresConfirm(t *testing.T) {
	// Setup
	cli.ResetResetFlags()
	mockRepo := newMockResetRepository()
	cli.SetRepository(mockRepo)
	defer cli.SetRepository(nil)

	// Create fresh command
	cmd := cli.NewRootCmd()
	cli.RegisterResetCommand(cmd)

	// Run without --confirm
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"reset", "my-project"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify warning message shown
	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("This deletes and recreates")) {
		t.Errorf("expected warning message, got: %s", output)
	}

	// Verify repo was NOT called
	if len(mockRepo.resetProjectCalls) > 0 {
		t.Error("expected ResetProject to NOT be called without --confirm")
	}
}

func TestResetCommand_SingleProject(t *testing.T) {
	// Setup
	cli.ResetResetFlags()
	mockRepo := newMockResetRepository()
	cli.SetRepository(mockRepo)
	defer cli.SetRepository(nil)

	// Create fresh command
	cmd := cli.NewRootCmd()
	cli.RegisterResetCommand(cmd)

	// Run with --confirm
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"reset", "my-project", "--confirm"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify repo was called
	if len(mockRepo.resetProjectCalls) != 1 {
		t.Fatalf("expected 1 ResetProject call, got %d", len(mockRepo.resetProjectCalls))
	}
	if mockRepo.resetProjectCalls[0] != "my-project" {
		t.Errorf("expected 'my-project', got %s", mockRepo.resetProjectCalls[0])
	}

	// Verify success message
	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("✓ Reset: my-project")) {
		t.Errorf("expected success message, got: %s", output)
	}
}

func TestResetCommand_AllProjects(t *testing.T) {
	// Setup
	cli.ResetResetFlags()
	mockRepo := newMockResetRepository()
	mockRepo.resetAllResult = 5
	cli.SetRepository(mockRepo)
	defer cli.SetRepository(nil)

	// Create fresh command
	cmd := cli.NewRootCmd()
	cli.RegisterResetCommand(cmd)

	// Run with --all --confirm
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"reset", "--all", "--confirm"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify ResetAll was called
	if mockRepo.resetAllCount != 1 {
		t.Fatalf("expected 1 ResetAll call, got %d", mockRepo.resetAllCount)
	}

	// Verify success message shows count
	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("✓ Reset 5 projects")) {
		t.Errorf("expected '✓ Reset 5 projects', got: %s", output)
	}
}

func TestResetCommand_NoProjectNoAll(t *testing.T) {
	// Setup
	cli.ResetResetFlags()
	mockRepo := newMockResetRepository()
	cli.SetRepository(mockRepo)
	defer cli.SetRepository(nil)

	// Create fresh command
	cmd := cli.NewRootCmd()
	cli.RegisterResetCommand(cmd)

	// Run with --confirm but no project and no --all
	cmd.SetArgs([]string{"reset", "--confirm"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing project")
	}

	// Verify error message
	if !bytes.Contains([]byte(err.Error()), []byte("specify project name/path or use --all")) {
		t.Errorf("expected specific error, got: %v", err)
	}
}

func TestResetCommand_ProjectNotFound(t *testing.T) {
	// Setup
	cli.ResetResetFlags()
	mockRepo := newMockResetRepository()
	mockRepo.resetProjectErr = domain.ErrProjectNotFound
	cli.SetRepository(mockRepo)
	defer cli.SetRepository(nil)

	// Create fresh command
	cmd := cli.NewRootCmd()
	cli.RegisterResetCommand(cmd)

	// Run reset on non-existent project
	cmd.SetArgs([]string{"reset", "nonexistent", "--confirm"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent project")
	}

	// Verify error contains context
	if !bytes.Contains([]byte(err.Error()), []byte("Reset failed")) {
		t.Errorf("expected 'Reset failed' error, got: %v", err)
	}
}

// RegisterResetCommand helper for testing
func init() {
	// This is defined in reset.go via RootCmd.AddCommand,
	// but for isolated testing we need this helper
}
