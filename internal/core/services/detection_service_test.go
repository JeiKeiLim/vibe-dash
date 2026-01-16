// Package services_test provides tests for core business services.
package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
	"github.com/JeiKeiLim/vibe-dash/internal/core/services"
)

// mockRegistry implements ports.DetectorRegistry for testing
type mockRegistry struct {
	detectAllResult                 *domain.DetectionResult
	detectAllError                  error
	detectWithCoexistenceResults    []*domain.DetectionResult
	detectWithCoexistenceResultsSet bool // explicitly track if results were set
	detectors                       []ports.MethodDetector
}

func (m *mockRegistry) DetectAll(ctx context.Context, path string) (*domain.DetectionResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return m.detectAllResult, m.detectAllError
}

func (m *mockRegistry) Detectors() []ports.MethodDetector {
	return m.detectors
}

func (m *mockRegistry) DetectWithCoexistence(ctx context.Context, path string) ([]*domain.DetectionResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	// Return configured results if set
	if m.detectWithCoexistenceResultsSet {
		return m.detectWithCoexistenceResults, nil
	}
	// Fallback: return results based on detectors that can detect
	var results []*domain.DetectionResult
	for _, d := range m.detectors {
		if d.CanDetect(ctx, path) {
			result, err := d.Detect(ctx, path)
			if err == nil && result != nil {
				results = append(results, result)
			}
		}
	}
	return results, nil
}

// mockDetector implements ports.MethodDetector for testing DetectMultiple
type mockDetector struct {
	name         string
	canDetect    bool
	detectResult *domain.DetectionResult
	detectErr    error
}

func (m *mockDetector) Name() string { return m.name }
func (m *mockDetector) CanDetect(ctx context.Context, path string) bool {
	return m.canDetect
}
func (m *mockDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
	return m.detectResult, m.detectErr
}

// Task 1 Tests: DetectionService structure and constructor
func TestNewDetectionService_CreatesValidService(t *testing.T) {
	mock := &mockRegistry{}
	svc := services.NewDetectionService(mock)

	if svc == nil {
		t.Error("NewDetectionService returned nil")
	}
}

func TestNewDetectionService_PanicsWithNilRegistry(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry, got none")
		}
	}()
	services.NewDetectionService(nil)
}

// Task 2 Tests: Detect method
func TestDetectionService_Detect_DelegatesToRegistry(t *testing.T) {
	expectedResult := domain.NewDetectionResult(
		"speckit",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"plan.md found",
	)
	mock := &mockRegistry{
		detectAllResult: &expectedResult,
	}
	svc := services.NewDetectionService(mock)

	result, err := svc.Detect(context.Background(), "/some/path")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "speckit" {
		t.Errorf("Method = %q, want %q", result.Method, "speckit")
	}
}

func TestDetectionService_Detect_ReturnsUnknownWhenNoMatch(t *testing.T) {
	unknownResult := domain.NewDetectionResult(
		"unknown",
		domain.StageUnknown,
		domain.ConfidenceUncertain,
		"no methodology markers found",
	)
	mock := &mockRegistry{
		detectAllResult: &unknownResult,
	}
	svc := services.NewDetectionService(mock)

	result, err := svc.Detect(context.Background(), "/some/path")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "unknown" {
		t.Errorf("Method = %q, want %q", result.Method, "unknown")
	}
	if result.Stage != domain.StageUnknown {
		t.Errorf("Stage = %v, want %v", result.Stage, domain.StageUnknown)
	}
}

func TestDetectionService_Detect_WrapsRegistryError(t *testing.T) {
	mock := &mockRegistry{
		detectAllError: errors.New("internal registry error"),
	}
	svc := services.NewDetectionService(mock)

	_, err := svc.Detect(context.Background(), "/some/path")

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrDetectionFailed) {
		t.Errorf("expected ErrDetectionFailed, got %v", err)
	}
}

func TestDetectionService_Detect_HandlesContextCancellation(t *testing.T) {
	expectedResult := domain.NewDetectionResult(
		"speckit",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"",
	)
	mock := &mockRegistry{
		detectAllResult: &expectedResult,
	}
	svc := services.NewDetectionService(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := svc.Detect(ctx, "/some/path")

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDetectionService_Detect_CancellationTiming(t *testing.T) {
	// AC6: cancellation must respond within 100ms
	expectedResult := domain.NewDetectionResult(
		"speckit",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"",
	)
	mock := &mockRegistry{
		detectAllResult: &expectedResult,
	}
	svc := services.NewDetectionService(mock)

	ctx, cancel := context.WithCancel(context.Background())

	// Start detection in goroutine
	done := make(chan error, 1)
	go func() {
		_, err := svc.Detect(ctx, "/some/path")
		done <- err
	}()

	// Cancel after brief delay
	time.Sleep(10 * time.Millisecond)
	cancelStart := time.Now()
	cancel()

	// Wait for completion
	select {
	case <-done:
		elapsed := time.Since(cancelStart)
		if elapsed > 100*time.Millisecond {
			t.Errorf("Cancellation took %v, expected < 100ms (AC6)", elapsed)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Detection did not respond to cancellation within 200ms")
	}
}

func TestDetectionService_Detect_ReturnsErrorForEmptyPath(t *testing.T) {
	mock := &mockRegistry{}
	svc := services.NewDetectionService(mock)

	_, err := svc.Detect(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty path")
	}
	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("expected ErrPathNotAccessible, got %v", err)
	}
}

// Task 3 Tests: DetectMultiple method
func TestDetectionService_DetectMultiple_CollectsAllMatchingResults(t *testing.T) {
	speckitResult := domain.NewDetectionResult(
		"speckit",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"plan.md found",
	)
	bmadResult := domain.NewDetectionResult(
		"bmad",
		domain.StageTasks,
		domain.ConfidenceLikely,
		".bmad folder found",
	)

	mock := &mockRegistry{
		detectors: []ports.MethodDetector{
			&mockDetector{
				name:         "speckit",
				canDetect:    true,
				detectResult: &speckitResult,
			},
			&mockDetector{
				name:         "bmad",
				canDetect:    true,
				detectResult: &bmadResult,
			},
		},
	}
	svc := services.NewDetectionService(mock)

	results, err := svc.DetectMultiple(context.Background(), "/some/path")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Method != "speckit" {
		t.Errorf("results[0].Method = %q, want %q", results[0].Method, "speckit")
	}
	if results[1].Method != "bmad" {
		t.Errorf("results[1].Method = %q, want %q", results[1].Method, "bmad")
	}
}

func TestDetectionService_DetectMultiple_ReturnsEmptySliceWhenNoMatches(t *testing.T) {
	mock := &mockRegistry{
		detectors: []ports.MethodDetector{
			&mockDetector{
				name:      "speckit",
				canDetect: false, // Does not match
			},
		},
	}
	svc := services.NewDetectionService(mock)

	results, err := svc.DetectMultiple(context.Background(), "/some/path")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty slice, got %d results", len(results))
	}
}

func TestDetectionService_DetectMultiple_ReturnsEmptySliceWhenAllDetectorsFail(t *testing.T) {
	mock := &mockRegistry{
		detectors: []ports.MethodDetector{
			&mockDetector{
				name:      "failing1",
				canDetect: true,
				detectErr: errors.New("detector failed"),
			},
			&mockDetector{
				name:      "failing2",
				canDetect: true,
				detectErr: errors.New("another failure"),
			},
		},
	}
	svc := services.NewDetectionService(mock)

	results, err := svc.DetectMultiple(context.Background(), "/some/path")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty slice when all detectors fail, got %d results", len(results))
	}
}

func TestDetectionService_DetectMultiple_ReturnsErrorForEmptyPath(t *testing.T) {
	mock := &mockRegistry{}
	svc := services.NewDetectionService(mock)

	_, err := svc.DetectMultiple(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty path")
	}
	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("expected ErrPathNotAccessible, got %v", err)
	}
}

func TestDetectionService_DetectMultiple_HandlesContextCancellation(t *testing.T) {
	speckitResult := domain.NewDetectionResult(
		"speckit",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"",
	)
	mock := &mockRegistry{
		detectors: []ports.MethodDetector{
			&mockDetector{
				name:         "speckit",
				canDetect:    true,
				detectResult: &speckitResult,
			},
		},
	}
	svc := services.NewDetectionService(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := svc.DetectMultiple(ctx, "/some/path")

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDetectionService_DetectMultiple_PartialResultsOnMixedSuccess(t *testing.T) {
	// Test that successful detections are returned even if some fail
	speckitResult := domain.NewDetectionResult(
		"speckit",
		domain.StagePlan,
		domain.ConfidenceCertain,
		"plan.md found",
	)

	mock := &mockRegistry{
		detectors: []ports.MethodDetector{
			&mockDetector{
				name:         "speckit",
				canDetect:    true,
				detectResult: &speckitResult,
			},
			&mockDetector{
				name:      "failing",
				canDetect: true,
				detectErr: errors.New("detector failed"),
			},
		},
	}
	svc := services.NewDetectionService(mock)

	results, err := svc.DetectMultiple(context.Background(), "/some/path")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (partial success), got %d", len(results))
	}
	if results[0].Method != "speckit" {
		t.Errorf("results[0].Method = %q, want %q", results[0].Method, "speckit")
	}
}

// createResultWithTimestamp is a test helper for DetectWithCoexistenceSelection tests
func createTestResultWithTimestamp(method string, timestamp time.Time) *domain.DetectionResult {
	result := domain.NewDetectionResult(method, domain.StagePlan, domain.ConfidenceCertain, "test").WithTimestamp(timestamp)
	return &result
}

// Task 5 Tests: DetectWithCoexistenceSelection method
func TestDetectionService_DetectWithCoexistenceSelection_ClearWinner(t *testing.T) {
	now := time.Now()
	mock := &mockRegistry{
		detectWithCoexistenceResults: []*domain.DetectionResult{
			createTestResultWithTimestamp("speckit", now.Add(-7*24*time.Hour)), // 1 week ago
			createTestResultWithTimestamp("bmad", now),                         // now
		},
		detectWithCoexistenceResultsSet: true,
	}
	svc := services.NewDetectionService(mock)

	winner, all, err := svc.DetectWithCoexistenceSelection(context.Background(), "/test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if winner == nil {
		t.Fatal("expected winner, got nil")
	}
	if winner.Method != "bmad" {
		t.Errorf("winner.Method = %q, want %q", winner.Method, "bmad")
	}
	if len(all) != 2 {
		t.Errorf("len(all) = %d, want 2", len(all))
	}
}

func TestDetectionService_DetectWithCoexistenceSelection_Tie(t *testing.T) {
	now := time.Now()
	mock := &mockRegistry{
		detectWithCoexistenceResults: []*domain.DetectionResult{
			createTestResultWithTimestamp("speckit", now),
			createTestResultWithTimestamp("bmad", now.Add(-30*time.Minute)), // 30 min ago (within threshold)
		},
		detectWithCoexistenceResultsSet: true,
	}
	svc := services.NewDetectionService(mock)

	winner, all, err := svc.DetectWithCoexistenceSelection(context.Background(), "/test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if winner != nil {
		t.Errorf("expected nil winner for tie, got %v", winner)
	}
	if len(all) != 2 {
		t.Errorf("len(all) = %d, want 2", len(all))
	}
}

func TestDetectionService_DetectWithCoexistenceSelection_SingleResult(t *testing.T) {
	now := time.Now()
	mock := &mockRegistry{
		detectWithCoexistenceResults: []*domain.DetectionResult{
			createTestResultWithTimestamp("speckit", now),
		},
		detectWithCoexistenceResultsSet: true,
	}
	svc := services.NewDetectionService(mock)

	winner, all, err := svc.DetectWithCoexistenceSelection(context.Background(), "/test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if winner == nil {
		t.Fatal("expected winner, got nil")
	}
	if winner.Method != "speckit" {
		t.Errorf("winner.Method = %q, want %q", winner.Method, "speckit")
	}
	if len(all) != 1 {
		t.Errorf("len(all) = %d, want 1", len(all))
	}
}

func TestDetectionService_DetectWithCoexistenceSelection_NoResults(t *testing.T) {
	mock := &mockRegistry{
		detectWithCoexistenceResults:    []*domain.DetectionResult{},
		detectWithCoexistenceResultsSet: true,
	}
	svc := services.NewDetectionService(mock)

	winner, all, err := svc.DetectWithCoexistenceSelection(context.Background(), "/test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if winner == nil {
		t.Fatal("expected unknown result, got nil")
	}
	if winner.Method != "unknown" {
		t.Errorf("winner.Method = %q, want %q", winner.Method, "unknown")
	}
	if all != nil {
		t.Errorf("expected nil all results for unknown, got %v", all)
	}
}

func TestDetectionService_DetectWithCoexistenceSelection_ContextCancellation(t *testing.T) {
	mock := &mockRegistry{
		detectWithCoexistenceResults:    []*domain.DetectionResult{},
		detectWithCoexistenceResultsSet: true,
	}
	svc := services.NewDetectionService(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, err := svc.DetectWithCoexistenceSelection(ctx, "/test")

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDetectionService_DetectWithCoexistenceSelection_EmptyPath(t *testing.T) {
	mock := &mockRegistry{}
	svc := services.NewDetectionService(mock)

	_, _, err := svc.DetectWithCoexistenceSelection(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty path")
	}
	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("expected ErrPathNotAccessible, got %v", err)
	}
}

// mockRegistryWithCoexistenceError is a specialized mock that returns an error from DetectWithCoexistence
type mockRegistryWithCoexistenceError struct {
	mockRegistry
	coexistenceErr error
}

func (m *mockRegistryWithCoexistenceError) DetectWithCoexistence(ctx context.Context, path string) ([]*domain.DetectionResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return nil, m.coexistenceErr
}

func TestDetectionService_DetectWithCoexistenceSelection_RegistryError(t *testing.T) {
	mock := &mockRegistryWithCoexistenceError{
		coexistenceErr: errors.New("registry failed"),
	}
	svc := services.NewDetectionService(mock)

	_, _, err := svc.DetectWithCoexistenceSelection(context.Background(), "/test")

	if err == nil {
		t.Error("expected error when registry fails")
	}
	if !errors.Is(err, domain.ErrDetectionFailed) {
		t.Errorf("expected ErrDetectionFailed wrapper, got %v", err)
	}
}
