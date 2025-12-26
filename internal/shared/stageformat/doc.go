// Package stageformat provides formatting functions for stage information display.
// It converts rich detection data (especially BMAD DetectionReasoning) into
// condensed display strings suitable for the project list view.
//
// This package is part of the shared layer and only imports from core/domain.
// It must NOT import from adapters or TUI packages.
//
// Example usage:
//
//	info := stageformat.FormatStageInfo(&project) // "E8 S8.3 review"
//	truncated := stageformat.FormatStageInfoWithWidth(&project, 10) // "E8 S8.3..."
package stageformat
