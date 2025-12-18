package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultBasePath(t *testing.T) {
	t.Run("returns ~/.vibe-dash when home directory is accessible", func(t *testing.T) {
		result := GetDefaultBasePath()

		// Should not be empty
		if result == "" {
			t.Fatal("GetDefaultBasePath() returned empty string")
		}

		// Should end with .vibe-dash
		if filepath.Base(result) != ".vibe-dash" {
			t.Errorf("GetDefaultBasePath() = %q, want path ending with '.vibe-dash'", result)
		}

		// Should be an absolute path
		if !filepath.IsAbs(result) {
			t.Errorf("GetDefaultBasePath() = %q, want absolute path", result)
		}

		// Should match home directory
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("Cannot determine home directory: %v", err)
		}
		expected := filepath.Join(home, ".vibe-dash")
		if result != expected {
			t.Errorf("GetDefaultBasePath() = %q, want %q", result, expected)
		}
	})
}
