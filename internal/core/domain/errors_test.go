package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestDomainErrors_Exist(t *testing.T) {
	// Verify all domain errors are defined and not nil
	domainErrors := []error{
		ErrProjectNotFound,
		ErrProjectAlreadyExists,
		ErrDetectionFailed,
		ErrConfigInvalid,
		ErrPathNotAccessible,
		ErrInvalidStage,
		ErrInvalidConfidence,
	}

	for _, err := range domainErrors {
		if err == nil {
			t.Error("Domain error should not be nil")
		}
	}
}

func TestDomainErrors_Messages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrProjectNotFound", ErrProjectNotFound, "project not found"},
		{"ErrProjectAlreadyExists", ErrProjectAlreadyExists, "project already exists"},
		{"ErrDetectionFailed", ErrDetectionFailed, "detection failed"},
		{"ErrConfigInvalid", ErrConfigInvalid, "configuration invalid"},
		{"ErrPathNotAccessible", ErrPathNotAccessible, "path not accessible"},
		{"ErrInvalidStage", ErrInvalidStage, "invalid stage"},
		{"ErrInvalidConfidence", ErrInvalidConfidence, "invalid confidence level"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDomainErrors_Wrapping(t *testing.T) {
	// Test that errors can be wrapped and unwrapped correctly
	wrapped := fmt.Errorf("failed to access project: %w", ErrProjectNotFound)

	if !errors.Is(wrapped, ErrProjectNotFound) {
		t.Error("wrapped error should match ErrProjectNotFound")
	}

	// Test wrapping with context
	wrappedPath := fmt.Errorf("%w: path must be absolute", ErrPathNotAccessible)

	if !errors.Is(wrappedPath, ErrPathNotAccessible) {
		t.Error("wrapped error should match ErrPathNotAccessible")
	}
}

func TestDomainErrors_Distinct(t *testing.T) {
	// Verify all domain errors are distinct
	errs := map[string]error{
		"ErrProjectNotFound":      ErrProjectNotFound,
		"ErrProjectAlreadyExists": ErrProjectAlreadyExists,
		"ErrDetectionFailed":      ErrDetectionFailed,
		"ErrConfigInvalid":        ErrConfigInvalid,
		"ErrPathNotAccessible":    ErrPathNotAccessible,
		"ErrInvalidStage":         ErrInvalidStage,
		"ErrInvalidConfidence":    ErrInvalidConfidence,
	}

	// Compare each error with every other error
	for name1, err1 := range errs {
		for name2, err2 := range errs {
			if name1 != name2 && errors.Is(err1, err2) {
				t.Errorf("%s should not be equal to %s", name1, name2)
			}
		}
	}
}
