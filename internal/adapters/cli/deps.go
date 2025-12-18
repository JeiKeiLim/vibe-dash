package cli

import "github.com/JeiKeiLim/vibe-dash/internal/core/ports"

// directoryManager handles project directory operations (delete).
var directoryManager ports.DirectoryManager

// SetDirectoryManager sets the directory manager for CLI commands.
func SetDirectoryManager(dm ports.DirectoryManager) {
	directoryManager = dm
}
