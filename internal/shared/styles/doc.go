// Package styles provides centralized Lipgloss style definitions for the vibe-dash TUI.
//
// This package resolves import cycles by providing a shared location for style
// definitions that can be imported by both the main tui package and its components
// subpackage. Previously, each component file duplicated styles from tui/styles.go
// with comments like "mirrored from tui/styles.go to avoid import cycle".
//
// Architecture:
//
//	internal/shared/styles/styles.go (defines styles) <- Single source of truth
//	    ↑                    ↑
//	tui/styles.go        tui/components/*.go
//	(re-exports +        (imports directly)
//	 helper funcs)
//
// Usage:
//
//	import "github.com/JeiKeiLim/vibe-dash/internal/shared/styles"
//
//	text := styles.SelectedStyle.Render("highlighted row")
//	warning := styles.WarningStyle.Render("⚠️ Warning")
//
// Color Palette:
//
// All styles use the 16-color ANSI palette for maximum terminal compatibility.
// See styles.go for the complete color reference table.
package styles
