package domain

import "errors"

// Domain-level sentinel errors
var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrDetectionFailed      = errors.New("detection failed")
	ErrConfigInvalid        = errors.New("configuration invalid")
	ErrPathNotAccessible    = errors.New("path not accessible")
	ErrInvalidStage         = errors.New("invalid stage")
	ErrInvalidConfidence    = errors.New("invalid confidence level")
)
