package config

import (
	"os"
	"path/filepath"
)

// GetDefaultBasePath returns the default vibe-dash storage directory.
// Returns ~/.vibe-dash on success, empty string on home dir lookup failure.
func GetDefaultBasePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".vibe-dash")
}
