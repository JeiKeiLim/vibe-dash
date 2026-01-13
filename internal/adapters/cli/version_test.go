package cli

import (
	"strings"
	"testing"
)

func TestSetVersion(t *testing.T) {
	// Test that SetVersion properly sets version info
	SetVersion("1.2.3", "abc1234", "2024-01-01")

	if appVersion != "1.2.3" {
		t.Errorf("appVersion = %q, want %q", appVersion, "1.2.3")
	}
	if appCommit != "abc1234" {
		t.Errorf("appCommit = %q, want %q", appCommit, "abc1234")
	}
	if appDate != "2024-01-01" {
		t.Errorf("appDate = %q, want %q", appDate, "2024-01-01")
	}

	// Verify RootCmd.Version was updated
	if RootCmd.Version != "1.2.3" {
		t.Errorf("RootCmd.Version = %q, want %q", RootCmd.Version, "1.2.3")
	}
}

func TestVersionTemplateFormat(t *testing.T) {
	// Ensure version is set first
	SetVersion("test", "abc123", "2024-01-01")

	// Test that version template contains expected format components
	template := RootCmd.VersionTemplate()

	expectedParts := []string{"vdash version", "commit:", "built:"}
	for _, part := range expectedParts {
		if !strings.Contains(template, part) {
			t.Errorf("Version template missing %q, got: %s", part, template)
		}
	}
}

func TestVersionFlagRegistered(t *testing.T) {
	// Ensure version is set first
	SetVersion("test", "abc123", "2024-01-01")

	// Verify --version flag is available (Cobra adds it when Version is set)
	if RootCmd.Version == "" {
		t.Error("RootCmd.Version should be set")
	}

	// Initialize default flags to ensure version flag is registered
	// Cobra only adds the version flag after InitDefaultVersionFlag is called
	RootCmd.InitDefaultVersionFlag()

	// Check that version flag exists in flags
	flag := RootCmd.Flags().Lookup("version")
	if flag == nil {
		t.Error("--version flag should be registered")
	}
}
