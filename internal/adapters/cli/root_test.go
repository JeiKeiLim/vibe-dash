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

func TestRootCmd_TUIPlaceholder(t *testing.T) {
	// Test that the Run function exists and contains the expected behavior
	// by calling it directly rather than through Execute() which has state issues
	if RootCmd.Run == nil {
		t.Fatal("RootCmd.Run should not be nil")
	}

	// Create a context that we'll cancel to simulate Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())

	// Create a command with the context attached
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)

	// Capture stdout
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Run in goroutine
	done := make(chan struct{})
	go func() {
		RootCmd.Run(cmd, []string{})
		close(done)
	}()

	// Wait for output to appear (poll with timeout instead of fixed sleep)
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		if buf.Len() > 0 {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	// Now cancel the context
	cancel()

	// Wait for function to finish
	select {
	case <-done:
		// Success - Run function exited after context cancellation
	case <-time.After(1 * time.Second):
		t.Fatal("Run function did not exit after context cancellation")
	}

	output := buf.String()
	// AC3: Should display placeholder message
	if !strings.Contains(output, "TUI dashboard") {
		t.Errorf("Output should mention TUI dashboard, got: %s", output)
	}
	if !strings.Contains(output, "Ctrl+C") {
		t.Errorf("Output should mention Ctrl+C to exit, got: %s", output)
	}
}
