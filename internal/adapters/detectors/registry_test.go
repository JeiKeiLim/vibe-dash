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
