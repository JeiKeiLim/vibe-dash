package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ============================================================================
// Story 3.6: CLI Refresh Command Tests (AC4)
// ============================================================================

// executeRefreshCommand runs the refresh command and returns output/error.
func executeRefreshCommand() (string, error) {
	cmd := cli.NewRootCmd()
	cli.RegisterRefreshCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"refresh"})

	err := cmd.Execute()
	return buf.String(), err
}

func TestRefreshCmd_NoProjects(t *testing.T) {
	// Setup empty repository
	mock := NewMockRepository()
	cli.SetRepository(mock)
	cli.SetDetectionService(&MockDetector{
		detectResult: &domain.DetectionResult{Method: "test", Stage: domain.StagePlan},
	})

	output, err := executeRefreshCommand()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "No projects to refresh") {
		t.Errorf("expected 'No projects to refresh' message, got: %s", output)
	}
}

func TestRefreshCmd_Success(t *testing.T) {
	// Setup with projects
	mock := NewMockRepository()
	project1 := &domain.Project{ID: "1", Path: "/test1", Name: "test1"}
	project2 := &domain.Project{ID: "2", Path: "/test2", Name: "test2"}
	_ = mock.Save(context.Background(), project1)
	_ = mock.Save(context.Background(), project2)

	cli.SetRepository(mock)
	cli.SetDetectionService(&MockDetector{
		detectResult: &domain.DetectionResult{Method: "bmad", Stage: domain.StagePlan},
	})

	output, err := executeRefreshCommand()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Refreshed 2 projects") {
		t.Errorf("expected 'Refreshed 2 projects' message, got: %s", output)
	}
}

func TestRefreshCmd_SingleProject(t *testing.T) {
	mock := NewMockRepository()
	project := &domain.Project{ID: "1", Path: "/test1", Name: "test1"}
	_ = mock.Save(context.Background(), project)

	cli.SetRepository(mock)
	cli.SetDetectionService(&MockDetector{
		detectResult: &domain.DetectionResult{Method: "speckit", Stage: domain.StageSpecify},
	})

	output, err := executeRefreshCommand()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Refreshed 1 projects") {
		t.Errorf("expected 'Refreshed 1 projects' message, got: %s", output)
	}
}

func TestRefreshCmd_PartialFailure(t *testing.T) {
	// Setup: Two projects, one will fail detection
	mock := NewMockRepository()
	project1 := &domain.Project{ID: "1", Path: "/test1", Name: "test1"}
	project2 := &domain.Project{ID: "2", Path: "/fail", Name: "fail-project"}
	_ = mock.Save(context.Background(), project1)
	_ = mock.Save(context.Background(), project2)

	callCount := 0
	cli.SetRepository(mock)
	cli.SetDetectionService(&MockDetector{
		// This is a simplified mock - in reality we'd need path-based logic
		detectResult: &domain.DetectionResult{Method: "test", Stage: domain.StagePlan},
	})

	// For partial failure test, we need a more sophisticated mock
	// Since our mock is simple, let's use a mock that tracks calls
	failingDetector := &CallCountDetector{
		results: map[int]*domain.DetectionResult{
			0: {Method: "test", Stage: domain.StagePlan},
		},
		errors: map[int]error{
			1: errors.New("detection failed"),
		},
		callCount: &callCount,
	}
	cli.SetDetectionService(failingDetector)

	output, err := executeRefreshCommand()

	// Partial success should not return error (AC3)
	if err != nil {
		t.Fatalf("partial failure should not return error, got: %v", err)
	}
	// Should report partial success
	if !strings.Contains(output, "1 failed") {
		t.Errorf("expected '(1 failed)' in output, got: %s", output)
	}
}

func TestRefreshCmd_AllFail(t *testing.T) {
	// Setup: All projects fail detection
	mock := NewMockRepository()
	project := &domain.Project{ID: "1", Path: "/test1", Name: "test1"}
	_ = mock.Save(context.Background(), project)

	cli.SetRepository(mock)
	cli.SetDetectionService(&MockDetector{
		detectErr: errors.New("detection failed"),
	})

	_, err := executeRefreshCommand()

	// AC3: Only return error if ALL projects fail
	if err == nil {
		t.Error("expected error when all projects fail")
	}
	if !strings.Contains(err.Error(), "failed to refresh") {
		t.Errorf("expected 'failed to refresh' in error, got: %s", err.Error())
	}
}

func TestRefreshCmd_NoRepository(t *testing.T) {
	cli.SetRepository(nil)
	cli.SetDetectionService(&MockDetector{})

	_, err := executeRefreshCommand()

	if err == nil {
		t.Error("expected error when repository is nil")
	}
	if !strings.Contains(err.Error(), "repository not initialized") {
		t.Errorf("expected 'repository not initialized' error, got: %s", err.Error())
	}
}

func TestRefreshCmd_NoDetectionService(t *testing.T) {
	mock := NewMockRepository()
	cli.SetRepository(mock)
	cli.SetDetectionService(nil)

	_, err := executeRefreshCommand()

	if err == nil {
		t.Error("expected error when detection service is nil")
	}
	if !strings.Contains(err.Error(), "detection service not initialized") {
		t.Errorf("expected 'detection service not initialized' error, got: %s", err.Error())
	}
}

func TestRefreshCmd_UpdatesProjectFields(t *testing.T) {
	mock := NewMockRepository()
	project := &domain.Project{
		ID:             "1",
		Path:           "/test1",
		Name:           "test1",
		DetectedMethod: "unknown",
		CurrentStage:   domain.StageUnknown,
	}
	_ = mock.Save(context.Background(), project)

	cli.SetRepository(mock)
	cli.SetDetectionService(&MockDetector{
		detectResult: &domain.DetectionResult{
			Method:     "bmad",
			Stage:      domain.StagePlan,
			Confidence: domain.ConfidenceCertain,
			Reasoning:  "Found BMAD artifacts",
		},
	})

	_, err := executeRefreshCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify project was updated
	updated, _ := mock.FindByPath(context.Background(), "/test1")
	if updated.DetectedMethod != "bmad" {
		t.Errorf("expected DetectedMethod to be 'bmad', got: %s", updated.DetectedMethod)
	}
	if updated.CurrentStage != domain.StagePlan {
		t.Errorf("expected CurrentStage to be StagePlan, got: %v", updated.CurrentStage)
	}
	if updated.Confidence != domain.ConfidenceCertain {
		t.Errorf("expected Confidence to be Certain, got: %v", updated.Confidence)
	}
}

// CallCountDetector is a detector that returns different results based on call count.
type CallCountDetector struct {
	results   map[int]*domain.DetectionResult
	errors    map[int]error
	callCount *int
}

func (d *CallCountDetector) Detect(_ context.Context, _ string) (*domain.DetectionResult, error) {
	idx := *d.callCount
	*d.callCount++

	if err, ok := d.errors[idx]; ok {
		return nil, err
	}
	if result, ok := d.results[idx]; ok {
		return result, nil
	}
	return &domain.DetectionResult{Method: "unknown", Stage: domain.StageUnknown}, nil
}

func (d *CallCountDetector) DetectMultiple(_ context.Context, _ string) ([]*domain.DetectionResult, error) {
	return nil, nil
}
