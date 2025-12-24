package cli

import (
	"errors"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// Exit codes for the CLI per Architecture "Error-to-Exit-Code Mapping"
const (
	ExitSuccess         = 0
	ExitGeneralError    = 1
	ExitProjectNotFound = 2
	ExitConfigInvalid   = 3
	ExitDetectionFailed = 4
)

// SilentError wraps an error to signal that it should not be logged.
// Used by commands like "exists" that communicate only via exit codes.
type SilentError struct {
	Err error
}

func (e *SilentError) Error() string {
	return e.Err.Error()
}

func (e *SilentError) Unwrap() error {
	return e.Err
}

// IsSilentError checks if an error should be silently handled (no logging).
func IsSilentError(err error) bool {
	var silent *SilentError
	return errors.As(err, &silent)
}

// ExitCodeDescription returns a human-readable description for an exit code.
// Used for programmatic access to exit code meanings.
func ExitCodeDescription(code int) string {
	switch code {
	case ExitSuccess:
		return "Success"
	case ExitGeneralError:
		return "General error (unhandled, user decision needed)"
	case ExitProjectNotFound:
		return "Project not found"
	case ExitConfigInvalid:
		return "Configuration invalid"
	case ExitDetectionFailed:
		return "Detection failed"
	default:
		return "Unknown exit code"
	}
}

// MapErrorToExitCode maps domain errors to CLI exit codes.
//
// Exit code mapping:
//   - ErrProjectNotFound     → 2 (specific, recoverable)
//   - ErrConfigInvalid       → 3 (specific, user can fix config)
//   - ErrDetectionFailed     → 4 (specific, retry may help)
//   - ErrProjectAlreadyExists → 1 (general - user decision needed)
//   - ErrPathNotAccessible   → 1 (general - filesystem issue)
//   - ErrInvalidStage        → 1 (general - internal error)
//   - ErrInvalidConfidence   → 1 (general - internal error)
//   - Any other error        → 1 (general catch-all)
func MapErrorToExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		return ExitProjectNotFound
	case errors.Is(err, domain.ErrConfigInvalid):
		return ExitConfigInvalid
	case errors.Is(err, domain.ErrDetectionFailed):
		return ExitDetectionFailed
	default:
		return ExitGeneralError
	}
}
