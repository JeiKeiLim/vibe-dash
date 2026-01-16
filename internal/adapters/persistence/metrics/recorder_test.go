package metrics

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// mockMetricsRepository is a test double for MetricsRepository
type mockMetricsRepository struct {
	transitions []recordedTransition
	mu          sync.Mutex
	recordError error
}

type recordedTransition struct {
	projectID string
	fromStage string
	toStage   string
}

func (m *mockMetricsRepository) RecordTransition(_ context.Context, projectID, fromStage, toStage string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.recordError != nil {
		return m.recordError
	}
	m.transitions = append(m.transitions, recordedTransition{projectID, fromStage, toStage})
	return nil
}

func (m *mockMetricsRepository) getTransitions() []recordedTransition {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]recordedTransition, len(m.transitions))
	copy(result, m.transitions)
	return result
}

// TestOnDetection_FirstDetection tests AC#4: first detection has empty from_stage
func TestOnDetection_FirstDetection(t *testing.T) {
	mock := &mockMetricsRepository{}
	recorder := NewMetricsRecorder(mock)
	recorder.debounceWindow = 10 * time.Millisecond // Speed up tests

	ctx := context.Background()
	recorder.OnDetection(ctx, "/path/to/project", domain.StagePlan)

	// Wait for debounce
	time.Sleep(20 * time.Millisecond)

	transitions := mock.getTransitions()
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}

	if transitions[0].fromStage != "" {
		t.Errorf("expected empty from_stage for first detection, got %q", transitions[0].fromStage)
	}
	if transitions[0].toStage != "Plan" {
		t.Errorf("expected to_stage='Plan', got %q", transitions[0].toStage)
	}
	if transitions[0].projectID != "/path/to/project" {
		t.Errorf("expected project_id='/path/to/project', got %q", transitions[0].projectID)
	}
}

// TestOnDetection_StageChange tests AC#1: transition recorded on stage change
func TestOnDetection_StageChange(t *testing.T) {
	mock := &mockMetricsRepository{}
	recorder := NewMetricsRecorder(mock)
	recorder.debounceWindow = 10 * time.Millisecond

	ctx := context.Background()
	projectID := "/my/project"

	// First detection
	recorder.OnDetection(ctx, projectID, domain.StagePlan)
	time.Sleep(20 * time.Millisecond)

	// Stage change
	recorder.OnDetection(ctx, projectID, domain.StageTasks)
	time.Sleep(20 * time.Millisecond)

	transitions := mock.getTransitions()
	if len(transitions) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(transitions))
	}

	// Second transition should have correct from/to
	if transitions[1].fromStage != "Plan" {
		t.Errorf("expected from_stage='Plan', got %q", transitions[1].fromStage)
	}
	if transitions[1].toStage != "Tasks" {
		t.Errorf("expected to_stage='Tasks', got %q", transitions[1].toStage)
	}
}

// TestOnDetection_NoChange tests AC#5: no transition recorded when stage unchanged
func TestOnDetection_NoChange(t *testing.T) {
	mock := &mockMetricsRepository{}
	recorder := NewMetricsRecorder(mock)
	recorder.debounceWindow = 10 * time.Millisecond

	ctx := context.Background()
	projectID := "/my/project"

	// First detection
	recorder.OnDetection(ctx, projectID, domain.StagePlan)
	time.Sleep(20 * time.Millisecond)

	// Same stage again - should NOT record
	recorder.OnDetection(ctx, projectID, domain.StagePlan)
	time.Sleep(20 * time.Millisecond)

	transitions := mock.getTransitions()
	if len(transitions) != 1 {
		t.Errorf("expected only 1 transition (no duplicate), got %d", len(transitions))
	}
}

// TestOnDetection_Debouncing tests AC#6: rapid transitions result in single record
func TestOnDetection_Debouncing(t *testing.T) {
	mock := &mockMetricsRepository{}
	recorder := NewMetricsRecorder(mock)
	recorder.debounceWindow = 50 * time.Millisecond

	ctx := context.Background()
	projectID := "/my/project"

	// Rapid stage changes within debounce window
	recorder.OnDetection(ctx, projectID, domain.StagePlan)
	time.Sleep(10 * time.Millisecond)
	recorder.OnDetection(ctx, projectID, domain.StageTasks)
	time.Sleep(10 * time.Millisecond)
	recorder.OnDetection(ctx, projectID, domain.StageImplement)

	// Wait for debounce to complete
	time.Sleep(100 * time.Millisecond)

	transitions := mock.getTransitions()
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition after debounce, got %d", len(transitions))
	}

	// Should record original fromStage (empty) and final toStage (Implement)
	if transitions[0].fromStage != "" {
		t.Errorf("expected from_stage='' (first detection), got %q", transitions[0].fromStage)
	}
	if transitions[0].toStage != "Implement" {
		t.Errorf("expected to_stage='Implement' (final stage), got %q", transitions[0].toStage)
	}
}

// TestOnDetection_NilRepository tests AC#2: graceful no-op when repository is nil
func TestOnDetection_NilRepository(t *testing.T) {
	recorder := NewMetricsRecorder(nil)

	ctx := context.Background()
	// Should not panic
	recorder.OnDetection(ctx, "/path/to/project", domain.StagePlan)

	// Also test nil receiver
	var nilRecorder *MetricsRecorder
	nilRecorder.OnDetection(ctx, "/path/to/project", domain.StagePlan)
}

// TestFlush_CommitsPending tests that Flush commits all pending transitions
func TestFlush_CommitsPending(t *testing.T) {
	mock := &mockMetricsRepository{}
	recorder := NewMetricsRecorder(mock)
	recorder.debounceWindow = 1 * time.Hour // Long window - won't fire naturally

	ctx := context.Background()

	// Create pending transitions (won't commit due to long debounce)
	recorder.OnDetection(ctx, "/project1", domain.StagePlan)
	recorder.OnDetection(ctx, "/project2", domain.StageTasks)

	// Should have no transitions yet
	if len(mock.getTransitions()) != 0 {
		t.Fatal("expected no transitions before flush")
	}

	// Flush should commit all pending
	recorder.Flush(ctx)

	transitions := mock.getTransitions()
	if len(transitions) != 2 {
		t.Fatalf("expected 2 transitions after flush, got %d", len(transitions))
	}
}

// TestFlush_CancelsTimers tests that Flush cancels timers (no goroutine leaks)
func TestFlush_CancelsTimers(t *testing.T) {
	mock := &mockMetricsRepository{}
	recorder := NewMetricsRecorder(mock)
	recorder.debounceWindow = 1 * time.Hour

	ctx := context.Background()
	recorder.OnDetection(ctx, "/project", domain.StagePlan)

	// Flush and verify pending is empty
	recorder.Flush(ctx)

	recorder.mu.Lock()
	pendingCount := len(recorder.pendingTransitions)
	recorder.mu.Unlock()

	if pendingCount != 0 {
		t.Errorf("expected 0 pending transitions after flush, got %d", pendingCount)
	}

	// Verify only 1 transition recorded (timer was cancelled)
	time.Sleep(50 * time.Millisecond) // Give time for any leaked goroutines
	if len(mock.getTransitions()) != 1 {
		t.Errorf("expected exactly 1 transition, got %d (possible goroutine leak)", len(mock.getTransitions()))
	}
}

// TestFlush_NilReceiver tests graceful handling of nil receiver
func TestFlush_NilReceiver(t *testing.T) {
	var recorder *MetricsRecorder
	recorder.Flush(context.Background()) // Should not panic
}

// TestOnDetection_MultipleProjects tests independent tracking per project
func TestOnDetection_MultipleProjects(t *testing.T) {
	mock := &mockMetricsRepository{}
	recorder := NewMetricsRecorder(mock)
	recorder.debounceWindow = 10 * time.Millisecond

	ctx := context.Background()

	// Multiple projects with different stages
	recorder.OnDetection(ctx, "/project-a", domain.StagePlan)
	recorder.OnDetection(ctx, "/project-b", domain.StageTasks)
	recorder.OnDetection(ctx, "/project-c", domain.StageImplement)

	time.Sleep(30 * time.Millisecond)

	transitions := mock.getTransitions()
	if len(transitions) != 3 {
		t.Fatalf("expected 3 transitions, got %d", len(transitions))
	}

	// Verify each project recorded independently
	projectMap := make(map[string]string)
	for _, tr := range transitions {
		projectMap[tr.projectID] = tr.toStage
	}

	if projectMap["/project-a"] != "Plan" {
		t.Error("project-a should be Plan")
	}
	if projectMap["/project-b"] != "Tasks" {
		t.Error("project-b should be Tasks")
	}
	if projectMap["/project-c"] != "Implement" {
		t.Error("project-c should be Implement")
	}
}
