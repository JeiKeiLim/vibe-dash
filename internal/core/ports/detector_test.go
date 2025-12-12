package ports_test

import (
	"context"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// mockDetector verifies interface compliance at compile time
type mockDetector struct{}

func (m *mockDetector) Name() string { return "mock" }

func (m *mockDetector) CanDetect(ctx context.Context, path string) bool {
	return path != ""
}

func (m *mockDetector) Detect(ctx context.Context, path string) (*domain.DetectionResult, error) {
	result := domain.NewDetectionResult("mock", domain.StageUnknown, domain.ConfidenceUncertain, "mock detection")
	return &result, nil
}

// Compile-time interface compliance check
var _ ports.MethodDetector = (*mockDetector)(nil)

func TestMethodDetector_InterfaceCompliance(t *testing.T) {
	var d ports.MethodDetector = &mockDetector{}

	t.Run("Name returns detector identifier", func(t *testing.T) {
		name := d.Name()
		if name != "mock" {
			t.Errorf("Name() = %q, want %q", name, "mock")
		}
	})

	t.Run("CanDetect accepts context and path", func(t *testing.T) {
		ctx := context.Background()
		result := d.CanDetect(ctx, "/some/path")
		if !result {
			t.Error("CanDetect() = false, want true for non-empty path")
		}
	})

	t.Run("Detect returns DetectionResult pointer", func(t *testing.T) {
		ctx := context.Background()
		result, err := d.Detect(ctx, "/some/path")
		if err != nil {
			t.Fatalf("Detect() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("Detect() returned nil result")
		}
		if result.Method != "mock" {
			t.Errorf("Detect().Method = %q, want %q", result.Method, "mock")
		}
	})

	t.Run("CanDetect accepts cancelled context without panic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// NOTE: This test only verifies the method signature accepts context.
		// Actual cancellation behavior (returning early, checking ctx.Done())
		// is tested at the adapter implementation level, not interface level.
		// Interface tests verify the contract shape; adapter tests verify behavior.
		_ = d.CanDetect(ctx, "/some/path")
	})

	t.Run("Detect accepts cancelled context without panic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// NOTE: This test only verifies the method signature accepts context.
		// Actual cancellation behavior (returning ctx.Err(), stopping work)
		// is tested at the adapter implementation level, not interface level.
		// Interface tests verify the contract shape; adapter tests verify behavior.
		_, _ = d.Detect(ctx, "/some/path")
	})
}
