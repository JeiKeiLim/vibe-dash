package detectors_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors/speckit"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// mockDetector is a test double for ports.MethodDetector
type mockDetector struct {
	name        string
	canDetect   bool
	result      *domain.DetectionResult
	detectErr   error
	detectCalls int
}

func (m *mockDetector) Name() string {
	return m.name
}

func (m *mockDetector) CanDetect(ctx context.Context, path string) bool {
	return m.canDetect
}

func (m *mockDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
	m.detectCalls++
	return m.result, m.detectErr
}

func TestNewRegistry(t *testing.T) {
	r := detectors.NewRegistry()
	if r == nil {
		t.Error("NewRegistry() returned nil")
	}

	// Should have no detectors initially
	if got := len(r.Detectors()); got != 0 {
		t.Errorf("NewRegistry().Detectors() length = %d, want 0", got)
	}
}

func TestRegistry_Register(t *testing.T) {
	r := detectors.NewRegistry()
	detector := speckit.NewSpeckitDetector()

	r.Register(detector)

	if got := len(r.Detectors()); got != 1 {
		t.Errorf("Detectors() length after Register = %d, want 1", got)
	}

	// Verify the detector is the one we registered
	if r.Detectors()[0].Name() != "speckit" {
		t.Errorf("Registered detector Name() = %q, want %q", r.Detectors()[0].Name(), "speckit")
	}
}

func TestRegistry_RegisterMultiple(t *testing.T) {
	r := detectors.NewRegistry()

	r.Register(&mockDetector{name: "detector1"})
	r.Register(&mockDetector{name: "detector2"})
	r.Register(&mockDetector{name: "detector3"})

	if got := len(r.Detectors()); got != 3 {
		t.Errorf("Detectors() length = %d, want 3", got)
	}

	// Verify order is preserved
	names := []string{"detector1", "detector2", "detector3"}
	for i, d := range r.Detectors() {
		if d.Name() != names[i] {
			t.Errorf("Detector[%d].Name() = %q, want %q", i, d.Name(), names[i])
		}
	}
}

func TestRegistry_DetectAll_FirstMatchWins(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	result1 := domain.NewDetectionResult("method1", domain.StageSpecify, domain.ConfidenceCertain, "reason1")
	result2 := domain.NewDetectionResult("method2", domain.StagePlan, domain.ConfidenceCertain, "reason2")

	detector1 := &mockDetector{name: "first", canDetect: true, result: &result1}
	detector2 := &mockDetector{name: "second", canDetect: true, result: &result2}

	r.Register(detector1)
	r.Register(detector2)

	result, err := r.DetectAll(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// First detector should win
	if result.Method != "method1" {
		t.Errorf("DetectAll().Method = %q, want %q", result.Method, "method1")
	}

	// Second detector should not have been called
	if detector2.detectCalls > 0 {
		t.Error("Second detector should not have been called")
	}
}

func TestRegistry_DetectAll_SkipsNonMatching(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	result2 := domain.NewDetectionResult("method2", domain.StagePlan, domain.ConfidenceCertain, "reason2")

	detector1 := &mockDetector{name: "first", canDetect: false}
	detector2 := &mockDetector{name: "second", canDetect: true, result: &result2}

	r.Register(detector1)
	r.Register(detector2)

	result, err := r.DetectAll(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// Second detector should win since first doesn't match
	if result.Method != "method2" {
		t.Errorf("DetectAll().Method = %q, want %q", result.Method, "method2")
	}
}

func TestRegistry_DetectAll_NoMatchReturnsUnknown(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	detector1 := &mockDetector{name: "first", canDetect: false}
	detector2 := &mockDetector{name: "second", canDetect: false}

	r.Register(detector1)
	r.Register(detector2)

	result, err := r.DetectAll(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// Should return unknown
	if result.Method != "unknown" {
		t.Errorf("DetectAll().Method = %q, want %q", result.Method, "unknown")
	}
	if result.Stage != domain.StageUnknown {
		t.Errorf("DetectAll().Stage = %v, want %v", result.Stage, domain.StageUnknown)
	}
	if result.Confidence != domain.ConfidenceUncertain {
		t.Errorf("DetectAll().Confidence = %v, want %v", result.Confidence, domain.ConfidenceUncertain)
	}
}

func TestRegistry_DetectAll_EmptyRegistry(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	result, err := r.DetectAll(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// Should return unknown
	if result.Method != "unknown" {
		t.Errorf("DetectAll().Method = %q, want %q", result.Method, "unknown")
	}
}

func TestRegistry_DetectAll_ContinuesOnDetectError(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	result2 := domain.NewDetectionResult("method2", domain.StagePlan, domain.ConfidenceCertain, "reason2")

	// First detector matches but returns error
	detector1 := &mockDetector{name: "first", canDetect: true, detectErr: errors.New("detection failed")}
	// Second detector should be used
	detector2 := &mockDetector{name: "second", canDetect: true, result: &result2}

	r.Register(detector1)
	r.Register(detector2)

	result, err := r.DetectAll(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// Second detector should be used since first errored
	if result.Method != "method2" {
		t.Errorf("DetectAll().Method = %q, want %q", result.Method, "method2")
	}
}

func TestRegistry_DetectAll_CollectsErrorsInReasoning(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	// Both detectors match but return errors - no fallback
	detector1 := &mockDetector{name: "speckit", canDetect: true, detectErr: errors.New("file read error")}
	detector2 := &mockDetector{name: "bmad", canDetect: true, detectErr: errors.New("parse error")}

	r.Register(detector1)
	r.Register(detector2)

	result, err := r.DetectAll(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// Should return unknown with error details in reasoning
	if result.Method != "unknown" {
		t.Errorf("DetectAll().Method = %q, want %q", result.Method, "unknown")
	}
	// Reasoning should contain the collected errors
	if result.Reasoning == "no methodology markers found" {
		t.Error("Reasoning should contain error details, not default message")
	}
	if !containsSubstring(result.Reasoning, "speckit") {
		t.Errorf("Reasoning should mention speckit detector, got: %q", result.Reasoning)
	}
	if !containsSubstring(result.Reasoning, "file read error") {
		t.Errorf("Reasoning should contain error message, got: %q", result.Reasoning)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRegistry_DetectAll_ContextCancellation(t *testing.T) {
	r := detectors.NewRegistry()

	detector := &mockDetector{name: "first", canDetect: true}
	r.Register(detector)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := r.DetectAll(ctx, "/some/path")
	if err != context.Canceled {
		t.Errorf("DetectAll() with cancelled context should return context.Canceled, got %v", err)
	}
}

// Verify Registry works with actual MethodDetector interface
func TestRegistry_WithRealDetector(t *testing.T) {
	r := detectors.NewRegistry()

	// Register the real speckit detector
	r.Register(speckit.NewSpeckitDetector())

	// Verify interface compliance
	var _ ports.MethodDetector = r.Detectors()[0]
}

// ============================================================================
// DetectWithCoexistence Tests
// ============================================================================

func TestRegistry_DetectWithCoexistence_MultipleMethods(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	result1 := domain.NewDetectionResult("speckit", domain.StageSpecify, domain.ConfidenceCertain, "spec.md exists")
	result2 := domain.NewDetectionResult("bmad", domain.StageTasks, domain.ConfidenceCertain, "sprint-status.yaml exists")

	detector1 := &mockDetector{name: "speckit", canDetect: true, result: &result1}
	detector2 := &mockDetector{name: "bmad", canDetect: true, result: &result2}

	r.Register(detector1)
	r.Register(detector2)

	results, err := r.DetectWithCoexistence(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectWithCoexistence() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("DetectWithCoexistence() returned %d results, want 2", len(results))
	}
	// Verify results are returned in registration order
	if results[0].Method != "speckit" {
		t.Errorf("DetectWithCoexistence()[0].Method = %q, want %q", results[0].Method, "speckit")
	}
	if results[1].Method != "bmad" {
		t.Errorf("DetectWithCoexistence()[1].Method = %q, want %q", results[1].Method, "bmad")
	}
	// Both detectors should have been called
	if detector1.detectCalls != 1 {
		t.Errorf("First detector called %d times, want 1", detector1.detectCalls)
	}
	if detector2.detectCalls != 1 {
		t.Errorf("Second detector called %d times, want 1", detector2.detectCalls)
	}
}

func TestRegistry_DetectWithCoexistence_SingleMatch(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	result1 := domain.NewDetectionResult("speckit", domain.StageSpecify, domain.ConfidenceCertain, "spec.md exists")

	detector1 := &mockDetector{name: "speckit", canDetect: true, result: &result1}
	detector2 := &mockDetector{name: "bmad", canDetect: false}

	r.Register(detector1)
	r.Register(detector2)

	results, err := r.DetectWithCoexistence(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectWithCoexistence() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("DetectWithCoexistence() returned %d results, want 1", len(results))
	}
	if results[0].Method != "speckit" {
		t.Errorf("DetectWithCoexistence()[0].Method = %q, want %q", results[0].Method, "speckit")
	}
}

func TestRegistry_DetectWithCoexistence_NoMatchReturnsEmptySlice(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	detector1 := &mockDetector{name: "speckit", canDetect: false}
	detector2 := &mockDetector{name: "bmad", canDetect: false}

	r.Register(detector1)
	r.Register(detector2)

	results, err := r.DetectWithCoexistence(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectWithCoexistence() error = %v", err)
	}

	if results == nil {
		t.Error("DetectWithCoexistence() returned nil, want empty slice")
	}
	if len(results) != 0 {
		t.Errorf("DetectWithCoexistence() returned %d results, want 0", len(results))
	}
}

func TestRegistry_DetectWithCoexistence_ContextCancellation(t *testing.T) {
	r := detectors.NewRegistry()

	detector := &mockDetector{name: "first", canDetect: true}
	r.Register(detector)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := r.DetectWithCoexistence(ctx, "/some/path")
	if err != context.Canceled {
		t.Errorf("DetectWithCoexistence() with cancelled context should return context.Canceled, got %v", err)
	}
}

func TestRegistry_DetectWithCoexistence_ErrorHandling(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	result2 := domain.NewDetectionResult("bmad", domain.StageTasks, domain.ConfidenceCertain, "sprint-status.yaml exists")

	// First detector returns error
	detector1 := &mockDetector{name: "speckit", canDetect: true, detectErr: errors.New("file read error")}
	// Second detector succeeds
	detector2 := &mockDetector{name: "bmad", canDetect: true, result: &result2}

	r.Register(detector1)
	r.Register(detector2)

	results, err := r.DetectWithCoexistence(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectWithCoexistence() should not return error, got %v", err)
	}

	// Should continue and return the successful result
	if len(results) != 1 {
		t.Errorf("DetectWithCoexistence() returned %d results, want 1", len(results))
	}
	if results[0].Method != "bmad" {
		t.Errorf("DetectWithCoexistence()[0].Method = %q, want %q", results[0].Method, "bmad")
	}
	// Both detectors should have been called
	if detector1.detectCalls != 1 {
		t.Errorf("First detector called %d times, want 1", detector1.detectCalls)
	}
	if detector2.detectCalls != 1 {
		t.Errorf("Second detector called %d times, want 1", detector2.detectCalls)
	}
}

func TestRegistry_DetectWithCoexistence_NilResultHandling(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	result2 := domain.NewDetectionResult("bmad", domain.StageTasks, domain.ConfidenceCertain, "sprint-status.yaml exists")

	// First detector returns nil result (no error)
	detector1 := &mockDetector{name: "speckit", canDetect: true, result: nil}
	// Second detector succeeds
	detector2 := &mockDetector{name: "bmad", canDetect: true, result: &result2}

	r.Register(detector1)
	r.Register(detector2)

	results, err := r.DetectWithCoexistence(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectWithCoexistence() error = %v", err)
	}

	// Should skip nil result and return only the successful one
	if len(results) != 1 {
		t.Errorf("DetectWithCoexistence() returned %d results, want 1", len(results))
	}
	if results[0].Method != "bmad" {
		t.Errorf("DetectWithCoexistence()[0].Method = %q, want %q", results[0].Method, "bmad")
	}
}

func TestRegistry_DetectWithCoexistence_EmptyRegistry(t *testing.T) {
	r := detectors.NewRegistry()
	ctx := context.Background()

	results, err := r.DetectWithCoexistence(ctx, "/some/path")
	if err != nil {
		t.Fatalf("DetectWithCoexistence() error = %v", err)
	}

	if results == nil {
		t.Error("DetectWithCoexistence() returned nil, want empty slice")
	}
	if len(results) != 0 {
		t.Errorf("DetectWithCoexistence() returned %d results, want 0", len(results))
	}
}
