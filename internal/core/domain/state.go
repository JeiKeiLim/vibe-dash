package domain

import (
	"errors"
	"strings"
)

// ProjectState represents the activity state of a project
type ProjectState int

const (
	StateActive ProjectState = iota // Zero value = Active (safe default)
	StateHibernated
)

// String returns human-readable name. Default returns "Unknown" for safety.
func (s ProjectState) String() string {
	switch s {
	case StateActive:
		return "Active"
	case StateHibernated:
		return "Hibernated"
	default:
		return "Unknown"
	}
}

// ErrInvalidProjectState is returned when parsing an invalid project state string
var ErrInvalidProjectState = errors.New("invalid project state")

// ParseProjectState converts string to ProjectState. Case-insensitive.
func ParseProjectState(s string) (ProjectState, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "active":
		return StateActive, nil
	case "hibernated":
		return StateHibernated, nil
	case "":
		return StateActive, nil
	default:
		return StateActive, ErrInvalidProjectState
	}
}
