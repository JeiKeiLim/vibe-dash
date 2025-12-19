//go:build integration

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// TestFsnotifyWatcher_Integration_RealWorld tests the watcher with realistic directory structure.
func TestFsnotifyWatcher_Integration_RealWorld(t *testing.T) {
	// Create realistic project structure
	tmpDir, err := os.MkdirTemp("", "watcher-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectories like a real project
	subdirs := []string{"src", "specs", ".bmad"}
	for _, subdir := range subdirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, subdir), 0755); err != nil {
			t.Fatalf("failed to create %s: %v", subdir, err)
		}
	}

	w := NewFsnotifyWatcher(100 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Watch root directory
	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(200 * time.Millisecond)

	// Test various file operations
	tests := []struct {
		name      string
		operation func() error
		expected  ports.FileOperation
	}{
		{
			name: "create file",
			operation: func() error {
				return os.WriteFile(filepath.Join(tmpDir, "new-file.txt"), []byte("content"), 0644)
			},
			expected: ports.FileOpCreate,
		},
		{
			name: "modify file",
			operation: func() error {
				return os.WriteFile(filepath.Join(tmpDir, "new-file.txt"), []byte("modified"), 0644)
			},
			expected: ports.FileOpModify,
		},
		{
			name: "delete file",
			operation: func() error {
				return os.Remove(filepath.Join(tmpDir, "new-file.txt"))
			},
			expected: ports.FileOpDelete,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.operation(); err != nil {
				t.Fatalf("operation failed: %v", err)
			}

			select {
			case event, ok := <-events:
				if !ok {
					t.Fatal("channel closed unexpectedly")
				}
				// Accept create or modify for file operations since some systems
				// report writes differently
				if tc.expected == ports.FileOpModify &&
					(event.Operation == ports.FileOpCreate || event.Operation == ports.FileOpModify) {
					// OK
				} else if event.Operation != tc.expected {
					t.Errorf("expected operation %v, got %v", tc.expected, event.Operation)
				}
			case <-time.After(2 * time.Second):
				t.Errorf("timeout waiting for %s event", tc.name)
			}
		})
	}
}

// TestFsnotifyWatcher_Integration_GracefulShutdown tests proper cleanup on shutdown.
func TestFsnotifyWatcher_Integration_GracefulShutdown(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-shutdown-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewFsnotifyWatcher(50 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context and close - simulate graceful shutdown
	cancel()
	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify channel is closed
	select {
	case _, ok := <-events:
		if ok {
			// Drain remaining events
			for range events {
			}
		}
		// Channel closed as expected
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for channel close - cleanup may not be complete")
	}
}

// TestFsnotifyWatcher_Integration_SubdirectoryWatch tests watching key subdirectories.
func TestFsnotifyWatcher_Integration_SubdirectoryWatch(t *testing.T) {
	// Create temp directory with subdirs
	tmpDir, err := os.MkdirTemp("", "watcher-subdir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Watch both root and src subdirectory
	events, err := w.Watch(ctx, []string{tmpDir, srcDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create file in subdirectory
	testFile := filepath.Join(srcDir, "main.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should receive event
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		if event.Path == "" {
			t.Error("expected non-empty path")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for subdirectory event")
	}
}

// TestFsnotifyWatcher_Integration_RaceCondition tests concurrent operations.
func TestFsnotifyWatcher_Integration_RaceCondition(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-race-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Concurrent file operations
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(n int) {
			filename := filepath.Join(tmpDir, "file"+string(rune('0'+n))+".txt")
			for j := 0; j < 10; j++ {
				os.WriteFile(filename, []byte("content"), 0644)
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Wait for goroutines to finish
	for i := 0; i < 5; i++ {
		<-done
	}

	// Drain events and ensure no panic/crash
	timeout := time.After(500 * time.Millisecond)
drainLoop:
	for {
		select {
		case _, ok := <-events:
			if !ok {
				break drainLoop
			}
		case <-timeout:
			break drainLoop
		}
	}
}
