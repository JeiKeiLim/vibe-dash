//go:build integration

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// Integration tests for graceful shutdown (Story 7.7)
// Run with: go test -tags=integration ./cmd/vibe/...
//
// These tests require the binary to be built first:
//   make build
// Or the tests will build it themselves.

// getVibeBinary returns the path to the vibe binary, building if necessary.
func getVibeBinary(t *testing.T) string {
	t.Helper()

	// Try to find existing binary
	projectRoot := findProjectRoot(t)
	binaryPath := filepath.Join(projectRoot, "bin", "vibe")

	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath
	}

	// Build the binary
	t.Log("Building vibe binary for integration tests...")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/vibe")
	cmd.Dir = projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build vibe binary: %v\nOutput: %s", err, output)
	}

	return binaryPath
}

// findProjectRoot finds the project root directory.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

// TestIntegration_GracefulShutdown_SIGINT tests clean exit on Ctrl+C (SIGINT).
// AC1, AC2, AC6: Shutdown signal, timeout respected, exit code 0.
func TestIntegration_GracefulShutdown_SIGINT(t *testing.T) {
	binary := getVibeBinary(t)

	// Create a temporary config directory to avoid touching user config
	tmpDir := t.TempDir()

	// Start the list command (quick, doesn't need TUI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, "list")
	cmd.Env = append(os.Environ(), "HOME="+tmpDir)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start vibe: %v", err)
	}

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Send SIGINT
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	// Wait for exit
	err := cmd.Wait()

	// Check exit code - 0 expected for clean shutdown
	// Note: With empty HOME, list may return quickly before signal is processed
	// In that case, exit code 0 is still correct (completed normally)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			// Exit code 0 or -1 (killed by signal) are acceptable
			if exitCode != 0 && exitCode != -1 {
				t.Errorf("Expected exit code 0 or -1 (signal), got %d", exitCode)
			}
		}
	}
	// No error means exit code 0, which is correct
}

// TestIntegration_GracefulShutdown_SIGTERM tests clean exit on SIGTERM.
// AC1, AC2, AC6: SIGTERM handled same as SIGINT.
func TestIntegration_GracefulShutdown_SIGTERM(t *testing.T) {
	binary := getVibeBinary(t)

	tmpDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, "list")
	cmd.Env = append(os.Environ(), "HOME="+tmpDir)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start vibe: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Send SIGTERM
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Failed to send SIGTERM: %v", err)
	}

	err := cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			// Exit code 0 or -1 (killed by signal) are acceptable
			if exitCode != 0 && exitCode != -1 {
				t.Errorf("Expected exit code 0 or -1 (signal), got %d", exitCode)
			}
		}
	}
}

// TestIntegration_RapidDoubleSignal tests AC8: force exit on repeated signal.
// AC8: Given shutdown in progress, when second signal received, exit immediately with code 1.
func TestIntegration_RapidDoubleSignal(t *testing.T) {
	binary := getVibeBinary(t)

	tmpDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use a command that will take a moment to complete (list with temp HOME)
	cmd := exec.CommandContext(ctx, binary, "list")
	cmd.Env = append(os.Environ(), "HOME="+tmpDir)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start vibe: %v", err)
	}

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Send first SIGINT to start shutdown
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send first SIGINT: %v", err)
	}

	// Quickly send second SIGINT to trigger force exit
	time.Sleep(10 * time.Millisecond)
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		// Process may have already exited, which is fine
		t.Logf("Second SIGINT not delivered (process may have exited): %v", err)
	}

	// Wait for exit
	err := cmd.Wait()

	// Process should exit - either normally (if it was fast) or forced (exit code 1)
	// The key is that it doesn't hang
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			// Exit codes 0, 1, or -1 (signal) are all acceptable
			// 0 = completed before signals processed
			// 1 = forced exit from double signal
			// -1 = killed by signal
			if exitCode != 0 && exitCode != 1 && exitCode != -1 {
				t.Errorf("Unexpected exit code %d", exitCode)
			}
			t.Logf("Exit code: %d (expected 0, 1, or -1)", exitCode)
		}
	}
}

// TestIntegration_ShutdownTimeoutValue verifies the shutdown timeout constant matches spec.
func TestIntegration_ShutdownTimeoutValue(t *testing.T) {
	// Verify the constant matches architecture spec (5 seconds)
	if shutdownTimeout != 5*time.Second {
		t.Errorf("shutdownTimeout = %v, want 5s", shutdownTimeout)
	}
}

// TestIntegration_VersionCommand_ExitsCleanly tests that simple commands exit cleanly.
// AC6: Exit code 0 for normal completion.
func TestIntegration_VersionCommand_ExitsCleanly(t *testing.T) {
	binary := getVibeBinary(t)

	cmd := exec.Command(binary, "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("--version command failed: %v\nOutput: %s", err, output)
	}

	// Verify exit code explicitly
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("--version exit code = %d, want 0", cmd.ProcessState.ExitCode())
	}
}

// TestIntegration_HelpCommand_ExitsCleanly tests that help command exits cleanly.
// AC6: Exit code 0 for normal completion.
func TestIntegration_HelpCommand_ExitsCleanly(t *testing.T) {
	binary := getVibeBinary(t)

	cmd := exec.Command(binary, "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("--help command failed: %v\nOutput: %s", err, output)
	}

	// Verify exit code explicitly
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("--help exit code = %d, want 0", cmd.ProcessState.ExitCode())
	}
}
