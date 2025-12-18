package config

// TODO(story:3.5.7): Update test fixture to use per-project storage structure at ~/.vibe-dash/<project>/state.db
// Current defaults reference single config.yaml. After Epic 3.5 completes,
// defaults should reflect both master index and per-project config paths.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigDirName(t *testing.T) {
	if DefaultConfigDirName != ".vibe-dash" {
		t.Errorf("DefaultConfigDirName = %s, want .vibe-dash", DefaultConfigDirName)
	}
}

func TestDefaultConfigFileName(t *testing.T) {
	if DefaultConfigFileName != "config.yaml" {
		t.Errorf("DefaultConfigFileName = %s, want config.yaml", DefaultConfigFileName)
	}
}

func TestGetDefaultConfigPath_Format(t *testing.T) {
	path := GetDefaultConfigPath()

	// Verify it's an absolute path or starts with .
	if !filepath.IsAbs(path) && path[0] != '.' {
		t.Errorf("GetDefaultConfigPath() = %s, should be absolute or start with '.'", path)
	}

	// Verify it ends with the expected file
	if filepath.Base(path) != DefaultConfigFileName {
		t.Errorf("GetDefaultConfigPath() base = %s, want %s", filepath.Base(path), DefaultConfigFileName)
	}

	// Verify parent directory name
	parentDir := filepath.Base(filepath.Dir(path))
	if parentDir != DefaultConfigDirName {
		t.Errorf("GetDefaultConfigPath() parent dir = %s, want %s", parentDir, DefaultConfigDirName)
	}
}

func TestGetConfigDir_Format(t *testing.T) {
	dir := GetConfigDir()

	// Verify it ends with the expected directory name
	if filepath.Base(dir) != DefaultConfigDirName {
		t.Errorf("GetConfigDir() base = %s, want %s", filepath.Base(dir), DefaultConfigDirName)
	}
}

func TestGetDefaultConfigPath_UsesHomeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("could not determine home directory")
	}

	path := GetDefaultConfigPath()
	expected := filepath.Join(home, DefaultConfigDirName, DefaultConfigFileName)

	if path != expected {
		t.Errorf("GetDefaultConfigPath() = %s, want %s", path, expected)
	}
}

func TestGetConfigDir_UsesHomeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("could not determine home directory")
	}

	dir := GetConfigDir()
	expected := filepath.Join(home, DefaultConfigDirName)

	if dir != expected {
		t.Errorf("GetConfigDir() = %s, want %s", dir, expected)
	}
}
