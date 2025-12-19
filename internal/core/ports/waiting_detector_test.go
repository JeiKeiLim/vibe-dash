package ports_test

import (
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
	"github.com/JeiKeiLim/vibe-dash/internal/core/services"
)

// TestWaitingDetectorInterface verifies that *services.WaitingDetector satisfies ports.WaitingDetector.
func TestWaitingDetectorInterface(t *testing.T) {
	// Compile-time check that *services.WaitingDetector implements ports.WaitingDetector
	var _ ports.WaitingDetector = (*services.WaitingDetector)(nil)
}
