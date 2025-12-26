//go:build integration

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

// =============================================================================
// Story 8.1: Recursive File Watching Integration Tests
// =============================================================================

// TestFsnotifyWatcher_Integration_RecursiveSubdirectory tests that files in subdirectories trigger events.
func TestFsnotifyWatcher_Integration_RecursiveSubdirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-recursive-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested subdirectory
	nestedDir := filepath.Join(tmpDir, "src", "pkg", "core")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Watch ONLY the root - subdirs should be watched automatically
	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(200 * time.Millisecond)

	// Create file in NESTED subdirectory (not root)
	testFile := filepath.Join(nestedDir, "main.go")
	if err := os.WriteFile(testFile, []byte("package core"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should receive event from nested subdirectory
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		if !strings.Contains(event.Path, "core") && !strings.Contains(event.Path, "main.go") {
			t.Errorf("expected event from nested dir, got path: %s", event.Path)
		}
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for event from nested subdirectory - recursive watching not working")
	}
}

// TestFsnotifyWatcher_Integration_DeeplyNested tests deeply nested file detection.
func TestFsnotifyWatcher_Integration_DeeplyNested(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-deep-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create deeply nested structure (a/b/c/d/e)
	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("failed to create deep dir: %v", err)
	}

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Create file in deeply nested directory
	testFile := filepath.Join(deepDir, "deep.txt")
	if err := os.WriteFile(testFile, []byte("deep content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		if !strings.Contains(event.Path, "deep.txt") {
			t.Errorf("expected event from deep file, got path: %s", event.Path)
		}
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for deeply nested file event")
	}
}

// TestFsnotifyWatcher_Integration_GitExcluded tests that .git directory is excluded.
func TestFsnotifyWatcher_Integration_GitExcluded(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-git-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git", "objects")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create src directory for positive control
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Touch file in .git - should NOT trigger event
	gitFile := filepath.Join(gitDir, "test")
	if err := os.WriteFile(gitFile, []byte("git data"), 0644); err != nil {
		t.Fatalf("failed to create .git file: %v", err)
	}

	// Wait briefly and check no .git events
	select {
	case event := <-events:
		if strings.Contains(event.Path, ".git") {
			t.Errorf(".git directory should be excluded, got event: %s", event.Path)
		}
	case <-time.After(500 * time.Millisecond):
		// Expected - no events from .git
	}

	// Positive control: src should still work
	srcFile := filepath.Join(srcDir, "test.go")
	if err := os.WriteFile(srcFile, []byte("package src"), 0644); err != nil {
		t.Fatalf("failed to create src file: %v", err)
	}

	select {
	case _, ok := <-events:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		// Good - received event from src
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for src event - positive control failed")
	}
}

// TestFsnotifyWatcher_Integration_BmadIncluded tests that .bmad directory IS watched.
func TestFsnotifyWatcher_Integration_BmadIncluded(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-bmad-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .bmad directory
	bmadDir := filepath.Join(tmpDir, ".bmad", "agents")
	if err := os.MkdirAll(bmadDir, 0755); err != nil {
		t.Fatalf("failed to create .bmad dir: %v", err)
	}

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Touch file in .bmad - SHOULD trigger event (exception to hidden rule)
	bmadFile := filepath.Join(bmadDir, "dev.md")
	if err := os.WriteFile(bmadFile, []byte("# Agent"), 0644); err != nil {
		t.Fatalf("failed to create .bmad file: %v", err)
	}

	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		if !strings.Contains(event.Path, ".bmad") && !strings.Contains(event.Path, "dev.md") {
			t.Errorf("expected .bmad event, got path: %s", event.Path)
		}
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for .bmad event - .bmad exception not working")
	}
}

// TestFsnotifyWatcher_Integration_ManyDirectories tests performance with many directories.
func TestFsnotifyWatcher_Integration_ManyDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-many-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create 100 directories
	for i := 0; i < 100; i++ {
		dir := filepath.Join(tmpDir, "pkg"+string(rune('a'+i/26))+string(rune('a'+i%26)))
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
	}

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}
	elapsed := time.Since(start)

	// Should complete in reasonable time (<2 seconds for 100 dirs)
	if elapsed > 2*time.Second {
		t.Errorf("Watch took too long: %v", elapsed)
	}

	time.Sleep(200 * time.Millisecond)

	// Create file in one of the directories
	testFile := filepath.Join(tmpDir, "pkgaa", "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	select {
	case _, ok := <-events:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for event in many-directory setup")
	}
}
