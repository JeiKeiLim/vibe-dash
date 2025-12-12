package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestMapErrorToExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		// Nil error returns success
		{"nil error", nil, ExitSuccess},

		// Direct domain errors
		{"project not found", domain.ErrProjectNotFound, ExitProjectNotFound},
		{"config invalid", domain.ErrConfigInvalid, ExitConfigInvalid},
		{"detection failed", domain.ErrDetectionFailed, ExitDetectionFailed},

		// Wrapped errors - errors.Is() must traverse the chain
		{"wrapped project not found", fmt.Errorf("failed: %w", domain.ErrProjectNotFound), ExitProjectNotFound},
		{"wrapped config invalid", fmt.Errorf("load config: %w", domain.ErrConfigInvalid), ExitConfigInvalid},
		{"wrapped detection failed", fmt.Errorf("detect: %w", domain.ErrDetectionFailed), ExitDetectionFailed},

		// Deeply wrapped errors
		{"deeply wrapped project not found", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", domain.ErrProjectNotFound)), ExitProjectNotFound},

		// Other domain errors map to general error
		{"project already exists", domain.ErrProjectAlreadyExists, ExitGeneralError},
		{"path not accessible", domain.ErrPathNotAccessible, ExitGeneralError},
		{"invalid stage", domain.ErrInvalidStage, ExitGeneralError},
		{"invalid confidence", domain.ErrInvalidConfidence, ExitGeneralError},

		// Unknown errors map to general error
		{"unknown error", errors.New("unknown"), ExitGeneralError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapErrorToExitCode(tt.err)
			if got != tt.expected {
				t.Errorf("MapErrorToExitCode(%v) = %d, want %d", tt.err, got, tt.expected)
			}
		})
	}
}

func TestExitCodeConstants(t *testing.T) {
	// Verify exit code values match documented behavior
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitGeneralError", ExitGeneralError, 1},
		{"ExitProjectNotFound", ExitProjectNotFound, 2},
		{"ExitConfigInvalid", ExitConfigInvalid, 3},
		{"ExitDetectionFailed", ExitDetectionFailed, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.constant, tt.expected)
			}
		})
	}
}
