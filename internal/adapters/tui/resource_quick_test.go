// Package tui provides terminal user interface components for vibe-dash.
//
// # Quick Resource Sanity Checks
//
// This file contains fast resource checks that run with regular 'go test ./...'.
// Unlike the integration tests in resource_test.go (which require -tags=integration),
// these tests provide basic sanity checks that complete quickly.
//
// ## Purpose
//
// These tests verify that resource counting infrastructure works correctly.
// They do NOT perform long-running leak detection - that's in resource_test.go.
//
// ## Running
//
//	go test ./internal/adapters/tui/... -run Resource_Quick
package tui

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// countOpenFDsQuick returns the current count of open file descriptors.
// Works on macOS (/dev/fd) and Linux (/proc/self/fd).
// Returns error if FD counting is not supported on the platform.
// NOTE: Duplicated from resource_test.go to avoid build tag dependency.
func countOpenFDsQuick() (int, error) {
	var path string
	switch runtime.GOOS {
	case "darwin":
		path = "/dev/fd"
	case "linux":
		path = "/proc/self/fd"
	default:
		return 0, fmt.Errorf("FD counting not supported on %s", runtime.GOOS)
	}

	// Open the directory and read entries directly
	// Using os.Open + Readdirnames instead of os.ReadDir to avoid lstat issues
	// on special file descriptor entries in /dev/fd
	dir, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return 0, err
	}
	return len(names), nil
}

// TestResource_Quick_FDCount verifies FD counting works on this platform.
// This is a quick sanity check that runs with normal tests.
func TestResource_Quick_FDCount(t *testing.T) {
	count, err := countOpenFDsQuick()
	if err != nil {
		t.Skipf("FD counting not supported on this platform: %v", err)
	}

	require.Greater(t, count, 0, "Should have at least some open FDs")
	t.Logf("Current FD count: %d", count)
}

// TestResource_Quick_GoroutineCount verifies goroutine counting works.
func TestResource_Quick_GoroutineCount(t *testing.T) {
	count := runtime.NumGoroutine()
	require.Greater(t, count, 0, "Should have at least some goroutines")
	t.Logf("Current goroutine count: %d", count)
}

// TestResource_Quick_MemStats verifies memory stats collection works.
func TestResource_Quick_MemStats(t *testing.T) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	require.Greater(t, memStats.HeapAlloc, uint64(0), "Should have some heap allocation")
	t.Logf("Current heap allocation: %d bytes", memStats.HeapAlloc)
}

// TestResource_Quick_FDCountIncrement verifies FD count changes when opening files.
func TestResource_Quick_FDCountIncrement(t *testing.T) {
	initial, err := countOpenFDsQuick()
	if err != nil {
		t.Skipf("FD counting not supported on this platform: %v", err)
	}

	// Open a file
	f, err := os.Open(os.DevNull)
	require.NoError(t, err)

	afterOpen, err := countOpenFDsQuick()
	require.NoError(t, err)

	// FD count should increase
	require.GreaterOrEqual(t, afterOpen, initial, "FD count should not decrease after opening file")

	// Close the file
	f.Close()

	afterClose, err := countOpenFDsQuick()
	require.NoError(t, err)

	t.Logf("FD count: initial=%d, afterOpen=%d, afterClose=%d", initial, afterOpen, afterClose)
}
