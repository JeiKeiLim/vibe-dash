//go:build integration

package metrics

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// TestRecorder_EndToEnd_IntegrationTest tests the full flow:
// OnDetection → Flush → query database → verify records
func TestRecorder_EndToEnd_IntegrationTest(t *testing.T) {
	// Setup: Create temp database
	tmpDir, err := os.MkdirTemp("", "metrics-integration-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	recorder := NewMetricsRecorder(repo)
	recorder.debounceWindow = 10 * time.Millisecond // Speed up test

	ctx := context.Background()
	projectID := "/test/project"

	// Act: Record stage transitions
	recorder.OnDetection(ctx, projectID, domain.StagePlan)
	time.Sleep(20 * time.Millisecond) // Wait for debounce

	recorder.OnDetection(ctx, projectID, domain.StageTasks)
	recorder.Flush(ctx) // Force commit

	// Verify: Query database directly
	db, err := repo.openDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM stage_transitions WHERE project_id = ?", projectID)
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Errorf("expected 2 transitions, got %d", count)
	}

	// Verify first transition has empty from_stage
	var firstFromStage string
	err = db.GetContext(ctx, &firstFromStage,
		"SELECT from_stage FROM stage_transitions WHERE project_id = ? ORDER BY transitioned_at LIMIT 1",
		projectID)
	if err != nil {
		t.Fatal(err)
	}

	if firstFromStage != "" {
		t.Errorf("first transition from_stage should be empty, got %q", firstFromStage)
	}

	// Verify second transition has correct from/to
	type transition struct {
		FromStage string `db:"from_stage"`
		ToStage   string `db:"to_stage"`
	}
	var second transition
	err = db.GetContext(ctx, &second,
		"SELECT from_stage, to_stage FROM stage_transitions WHERE project_id = ? ORDER BY transitioned_at DESC LIMIT 1",
		projectID)
	if err != nil {
		t.Fatal(err)
	}

	if second.FromStage != "Plan" {
		t.Errorf("second transition from_stage should be 'Plan', got %q", second.FromStage)
	}
	if second.ToStage != "Tasks" {
		t.Errorf("second transition to_stage should be 'Tasks', got %q", second.ToStage)
	}
}

// TestRecorder_ConcurrentProjects_IntegrationTest tests concurrent transitions from multiple projects
func TestRecorder_ConcurrentProjects_IntegrationTest(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "metrics-concurrent-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	recorder := NewMetricsRecorder(repo)
	recorder.debounceWindow = 10 * time.Millisecond

	ctx := context.Background()
	projectCount := 10
	stages := []domain.Stage{domain.StagePlan, domain.StageTasks, domain.StageImplement}

	// Act: Concurrent detection from multiple projects
	var wg sync.WaitGroup
	for i := 0; i < projectCount; i++ {
		projectID := "/project/" + string(rune('A'+i))
		for _, stage := range stages {
			wg.Add(1)
			go func(pid string, s domain.Stage) {
				defer wg.Done()
				recorder.OnDetection(ctx, pid, s)
			}(projectID, stage)
		}
	}
	wg.Wait()

	// Wait for debounce and flush
	time.Sleep(50 * time.Millisecond)
	recorder.Flush(ctx)

	// Verify: Each project should have exactly 1 transition (debouncing coalesces rapid changes)
	db, err := repo.openDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var totalCount int
	err = db.GetContext(ctx, &totalCount, "SELECT COUNT(*) FROM stage_transitions")
	if err != nil {
		t.Fatal(err)
	}

	// Each project should have exactly 1 transition (first detection → final stage)
	if totalCount != projectCount {
		t.Errorf("expected %d transitions (1 per project), got %d", projectCount, totalCount)
	}
}

// TestRecorder_RFC3339Nano_Timestamp_IntegrationTest verifies timestamps are in correct format
func TestRecorder_RFC3339Nano_Timestamp_IntegrationTest(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "metrics-timestamp-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	recorder := NewMetricsRecorder(repo)
	recorder.debounceWindow = 10 * time.Millisecond

	ctx := context.Background()

	// Act
	beforeTime := time.Now().UTC()
	recorder.OnDetection(ctx, "/test", domain.StagePlan)
	recorder.Flush(ctx)
	afterTime := time.Now().UTC()

	// Verify: Timestamp should be valid RFC3339Nano
	db, err := repo.openDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var timestampStr string
	err = db.GetContext(ctx, &timestampStr, "SELECT transitioned_at FROM stage_transitions LIMIT 1")
	if err != nil {
		t.Fatal(err)
	}

	timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		t.Errorf("timestamp %q is not valid RFC3339Nano: %v", timestampStr, err)
	}

	if timestamp.Before(beforeTime) || timestamp.After(afterTime) {
		t.Errorf("timestamp %v is outside expected range [%v, %v]", timestamp, beforeTime, afterTime)
	}
}

// TestRecorder_GracefulDegradation_IntegrationTest tests that metrics failure doesn't crash
func TestRecorder_GracefulDegradation_IntegrationTest(t *testing.T) {
	// Setup: Use non-existent directory to trigger failure
	recorder := NewMetricsRecorder(NewMetricsRepository("/nonexistent/path/metrics.db"))
	recorder.debounceWindow = 10 * time.Millisecond

	ctx := context.Background()

	// Act: Should not panic
	recorder.OnDetection(ctx, "/test", domain.StagePlan)
	time.Sleep(20 * time.Millisecond)
	recorder.Flush(ctx)

	// If we get here without panicking, the test passes
}
