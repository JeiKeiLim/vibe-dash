package cli

import (
	"log/slog"
	"os"
)

// resetTestState resets RootCmd state for clean test execution.
// Call this at the start of each test that modifies global state.
func resetTestState() {
	verbose = false
	debug = false
	quiet = false
	configFile = ""
	waitingThreshold = -1
	// Reset persistent flags to their defaults (ignore errors as flags always exist)
	_ = RootCmd.PersistentFlags().Set("verbose", "false")
	_ = RootCmd.PersistentFlags().Set("debug", "false")
	_ = RootCmd.PersistentFlags().Set("quiet", "false")
	_ = RootCmd.PersistentFlags().Set("config", "")
	_ = RootCmd.PersistentFlags().Set("waiting-threshold", "-1")
	// Clear any previous args
	RootCmd.SetArgs(nil)
}

// saveSlogDefault saves the current default slog logger and returns a restore function.
// Use this to isolate tests that modify global slog state.
//
// Usage:
//
//	restore := saveSlogDefault()
//	defer restore()
//	// ... test code that modifies slog ...
func saveSlogDefault() func() {
	// slog doesn't expose a way to get the current default logger,
	// so we create a new one with the same default settings to restore.
	// This ensures tests start with a known state.
	originalHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})
	return func() {
		slog.SetDefault(slog.New(originalHandler))
	}
}

// resetLoggingState resets both test state and logging to defaults.
// Combines resetTestState() with slog restoration for complete isolation.
func resetLoggingState() func() {
	resetTestState()
	return saveSlogDefault()
}
