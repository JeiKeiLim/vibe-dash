// Package config provides configuration loading and management using Viper.
// It implements the ports.ConfigLoader interface for YAML-based configuration.
package config

import (
	"log/slog"
	"os"
	"path/filepath"
)

const (
	// DefaultConfigDirName is the name of the configuration directory.
	DefaultConfigDirName = ".vibe-dash"

	// DefaultConfigFileName is the name of the configuration file.
	DefaultConfigFileName = "config.yaml"
)

// GetDefaultConfigPath returns the default config file path.
// Uses os.UserHomeDir() for cross-platform compatibility.
func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home can't be determined
		slog.Warn("could not determine home directory", "error", err)
		return filepath.Join(".", DefaultConfigDirName, DefaultConfigFileName)
	}
	return filepath.Join(home, DefaultConfigDirName, DefaultConfigFileName)
}

// GetConfigDir returns the config directory path.
func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("could not determine home directory", "error", err)
		return filepath.Join(".", DefaultConfigDirName)
	}
	return filepath.Join(home, DefaultConfigDirName)
}
