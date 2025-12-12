package cli

import (
	"strings"
	"testing"
)

func TestVersionVariables(t *testing.T) {
	// Test that version variables have expected defaults (non-empty)
	tests := []struct {
		name     string
		variable string
	}{
		{"Version", Version},
		{"Commit", Commit},
		{"BuildDate", BuildDate},
	}

	for _, tt := range tests {
		t.Run(tt.name+" is set", func(t *testing.T) {
			if tt.variable == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}

func TestVersionTemplateFormat(t *testing.T) {
	// Test that version template contains expected format components
	template := RootCmd.VersionTemplate()

	expectedParts := []string{"vibe version", "commit:", "built:"}
	for _, part := range expectedParts {
		if !strings.Contains(template, part) {
			t.Errorf("Version template missing %q, got: %s", part, template)
		}
	}
}

func TestVersionFlagRegistered(t *testing.T) {
	// Verify --version flag is available (Cobra adds it when Version is set)
	if RootCmd.Version == "" {
		t.Error("RootCmd.Version should be set")
	}

	// Check that version flag exists in flags
	flag := RootCmd.Flags().Lookup("version")
	if flag == nil {
		t.Error("--version flag should be registered")
	}
}
