package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestRootCmd_HelpOutput(t *testing.T) {
	resetTestState()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()

	// AC1: Help text should contain these elements
	expectedPhrases := []string{
		"vibe coding projects", // Description mentions vibe coding projects
		"AI",                   // Mentions AI-assisted
		"dashboard",            // Mentions dashboard
		"--verbose",            // Global flag
		"--debug",              // Global flag
		"--config",             // Global flag
		"Track",                // Describes tracking functionality
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Help output missing phrase %q\nGot: %s", phrase, output)
		}
	}
}

func TestRootCmd_ShortDescription(t *testing.T) {
	if RootCmd.Short == "" {
		t.Error("RootCmd.Short should not be empty")
	}
	if !strings.Contains(strings.ToLower(RootCmd.Short), "dashboard") {
		t.Errorf("Short description should mention 'dashboard', got: %s", RootCmd.Short)
	}
}

func TestRootCmd_LongDescription(t *testing.T) {
	if RootCmd.Long == "" {
		t.Error("RootCmd.Long should not be empty")
	}

	// Long description should have comprehensive content
	expectedContent := []string{
		"vibe coding projects",
		"Track",
		"AI",
	}

	for _, content := range expectedContent {
		if !strings.Contains(RootCmd.Long, content) {
			t.Errorf("Long description should contain %q, got: %s", content, RootCmd.Long)
		}
	}
}

func TestRootCmd_UsageLinePresent(t *testing.T) {
	// The Use field should be set
	if RootCmd.Use != "vibe" {
		t.Errorf("RootCmd.Use should be 'vibe', got: %s", RootCmd.Use)
	}
}

func TestRootCmd_HasRunFunction(t *testing.T) {
	// Root command should have a Run function (required for flags to show in help)
	if RootCmd.Run == nil {
		t.Error("RootCmd should have a Run function")
	}
}

func TestRootCmd_TUIIntegration(t *testing.T) {
	// Test that the Run function exists and is configured for TUI
	// Note: Actual TUI testing requires a terminal, so we only verify
	// the function exists and the command structure is correct.
	// TUI functionality is tested in internal/adapters/tui/model_test.go
	if RootCmd.Run == nil {
		t.Fatal("RootCmd.Run should not be nil")
	}

	// Verify the TUI starts (will fail immediately without TTY, which is expected)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := &cobra.Command{}
	cmd.SetContext(ctx)

	// Run in goroutine - will exit quickly due to no TTY in test environment
	done := make(chan struct{})
	go func() {
		RootCmd.Run(cmd, []string{})
		close(done)
	}()

	// TUI should exit quickly in test environment (no TTY available)
	select {
	case <-done:
		// Success - Run function completed (TUI failed to start due to no TTY, as expected)
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("Run function should exit quickly when no TTY is available")
	}
}
