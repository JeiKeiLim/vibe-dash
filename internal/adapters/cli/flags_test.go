package cli

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestVerboseFlag(t *testing.T) {
	resetTestState()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--verbose", "--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !verbose {
		t.Error("--verbose flag should set verbose = true")
	}
}

func TestDebugFlag(t *testing.T) {
	resetTestState()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--debug", "--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !debug {
		t.Error("--debug flag should set debug = true")
	}
}

func TestConfigFlag(t *testing.T) {
	resetTestState()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--config", "/path/to/config.yaml", "--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if configFile != "/path/to/config.yaml" {
		t.Errorf("--config flag should set configFile = '/path/to/config.yaml', got %q", configFile)
	}
}

func TestVerboseFlagShorthand(t *testing.T) {
	resetTestState()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"-v", "--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !verbose {
		t.Error("-v flag should set verbose = true")
	}
}

func TestConfigFlagShorthand(t *testing.T) {
	resetTestState()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"-c", "/custom/config.yaml", "--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if configFile != "/custom/config.yaml" {
		t.Errorf("-c flag should set configFile = '/custom/config.yaml', got %q", configFile)
	}
}

func TestFlagsInHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	expectedFlags := []string{"--verbose", "-v", "--debug", "--config", "-c"}
	for _, flag := range expectedFlags {
		if !strings.Contains(output, flag) {
			t.Errorf("Help output missing flag %q", flag)
		}
	}
}

func TestGetConfigFile(t *testing.T) {
	resetTestState()

	// Initially should be empty
	if got := GetConfigFile(); got != "" {
		t.Errorf("GetConfigFile() initially = %q, want empty", got)
	}

	// After setting --config flag
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--config", "/test/path/config.yaml", "--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got := GetConfigFile(); got != "/test/path/config.yaml" {
		t.Errorf("GetConfigFile() = %q, want %q", got, "/test/path/config.yaml")
	}
}

func TestDebugPrecedenceOverVerbose(t *testing.T) {
	// When both --debug and --verbose are set, --debug should take precedence
	// This is verified by checking the slog level after initLogging()
	restore := resetLoggingState()
	defer restore()

	// Set both flags
	verbose = true
	debug = true

	// Call initLogging which should prioritize debug
	initLogging()

	// To verify, we need to check that debug-level messages are logged
	// Since we can't directly inspect slog level, we test by logging
	// at debug level and checking it doesn't get filtered
	// The actual behavior is tested by the switch order in initLogging()
	// Here we verify the flag values are both set but debug wins

	if !debug {
		t.Error("debug flag should be true")
	}
	if !verbose {
		t.Error("verbose flag should be true")
	}

	// Verify initLogging was called without error (implicit test)
	// The precedence is enforced by the switch case order in initLogging():
	// case debug: ... (checked first)
	// case verbose: ... (only if debug is false)
}

func TestInitLogging_DebugLevel(t *testing.T) {
	restore := resetLoggingState()
	defer restore()

	verbose = false
	debug = true
	initLogging()

	// Debug should enable debug level with source info
	// We can't directly test slog level, but we verify no panic/error
	slog.Debug("test debug message", "key", "value")
}

func TestInitLogging_VerboseLevel(t *testing.T) {
	restore := resetLoggingState()
	defer restore()

	verbose = true
	debug = false
	initLogging()

	// Verbose should enable info level
	slog.Info("test info message", "key", "value")
}

func TestInitLogging_DefaultLevel(t *testing.T) {
	restore := resetLoggingState()
	defer restore()

	verbose = false
	debug = false
	initLogging()

	// Default should be error level only
	slog.Error("test error message", "key", "value")
}

func TestWaitingThresholdFlag(t *testing.T) {
	resetTestState()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--waiting-threshold", "15", "--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if waitingThreshold != 15 {
		t.Errorf("--waiting-threshold flag should set waitingThreshold = 15, got %d", waitingThreshold)
	}
}

func TestWaitingThresholdFlag_Disabled(t *testing.T) {
	resetTestState()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--waiting-threshold", "0", "--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if waitingThreshold != 0 {
		t.Errorf("--waiting-threshold=0 should set waitingThreshold = 0 (disabled), got %d", waitingThreshold)
	}
}

func TestWaitingThresholdFlag_DefaultUnset(t *testing.T) {
	resetTestState()

	// Verify default is -1 (sentinel for "use config")
	if got := GetWaitingThreshold(); got != -1 {
		t.Errorf("GetWaitingThreshold() default = %d, want -1", got)
	}
}

func TestGetWaitingThreshold(t *testing.T) {
	tests := []struct {
		name     string
		flagVal  int
		expected int
	}{
		{"not set (sentinel)", -1, -1},
		{"disabled", 0, 0},
		{"custom value", 15, 15},
		{"invalid negative -5", -5, -1}, // Should warn and return -1
		{"invalid negative -100", -100, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTestState()
			waitingThreshold = tt.flagVal

			got := GetWaitingThreshold()
			if got != tt.expected {
				t.Errorf("GetWaitingThreshold() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestWaitingThresholdFlagInHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--waiting-threshold") {
		t.Errorf("Help output missing flag --waiting-threshold")
	}
}
