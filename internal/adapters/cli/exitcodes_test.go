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

func TestExitCodeDescription(t *testing.T) {
	tests := []struct {
		code        int
		description string
	}{
		{ExitSuccess, "Success"},
		{ExitGeneralError, "General error (unhandled, user decision needed)"},
		{ExitProjectNotFound, "Project not found"},
		{ExitConfigInvalid, "Configuration invalid"},
		{ExitDetectionFailed, "Detection failed"},
		{99, "Unknown exit code"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got := ExitCodeDescription(tt.code)
			if got != tt.description {
				t.Errorf("ExitCodeDescription(%d) = %q, want %q", tt.code, got, tt.description)
			}
		})
	}
}

func TestSilentError(t *testing.T) {
	t.Run("wraps and unwraps error", func(t *testing.T) {
		inner := domain.ErrProjectNotFound
		silent := &SilentError{Err: inner}

		// Error() should delegate to inner
		if silent.Error() != inner.Error() {
			t.Errorf("Error() = %q, want %q", silent.Error(), inner.Error())
		}

		// Unwrap() should return inner
		if silent.Unwrap() != inner {
			t.Errorf("Unwrap() = %v, want %v", silent.Unwrap(), inner)
		}
	})

	t.Run("errors.Is works through SilentError", func(t *testing.T) {
		silent := &SilentError{Err: fmt.Errorf("context: %w", domain.ErrProjectNotFound)}

		if !errors.Is(silent, domain.ErrProjectNotFound) {
			t.Error("errors.Is should find ErrProjectNotFound through SilentError")
		}
	})

	t.Run("MapErrorToExitCode works through SilentError", func(t *testing.T) {
		silent := &SilentError{Err: domain.ErrProjectNotFound}

		got := MapErrorToExitCode(silent)
		if got != ExitProjectNotFound {
			t.Errorf("MapErrorToExitCode(SilentError) = %d, want %d", got, ExitProjectNotFound)
		}
	})
}

func TestIsSilentError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"regular error", errors.New("regular"), false},
		{"domain error", domain.ErrProjectNotFound, false},
		{"silent error", &SilentError{Err: errors.New("silent")}, true},
		{"wrapped silent error", fmt.Errorf("wrap: %w", &SilentError{Err: errors.New("inner")}), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSilentError(tt.err)
			if got != tt.expected {
				t.Errorf("IsSilentError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}
