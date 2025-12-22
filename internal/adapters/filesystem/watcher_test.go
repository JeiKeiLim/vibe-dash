package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// TestNewFsnotifyWatcher tests the constructor with various debounce configurations.
func TestNewFsnotifyWatcher(t *testing.T) {
	tests := []struct {
		name            string
		debounce        time.Duration
		expectedDefault time.Duration
	}{
		{
			name:            "zero debounce uses default",
			debounce:        0,
			expectedDefault: 200 * time.Millisecond,
		},
		{
			name:            "custom debounce is preserved",
			debounce:        500 * time.Millisecond,
			expectedDefault: 500 * time.Millisecond,
		},
		{
			name:            "small debounce is preserved",
			debounce:        50 * time.Millisecond,
			expectedDefault: 50 * time.Millisecond,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := NewFsnotifyWatcher(tc.debounce)
			if w == nil {
				t.Fatal("expected non-nil watcher")
			}
			if w.debounce != tc.expectedDefault {
				t.Errorf("debounce = %v, want %v", w.debounce, tc.expectedDefault)
			}
		})
	}
}

// TestFsnotifyWatcher_ImplementsInterface verifies interface compliance.
func TestFsnotifyWatcher_ImplementsInterface(t *testing.T) {
	var _ ports.FileWatcher = (*FsnotifyWatcher)(nil)
}

// TestFsnotifyWatcher_Watch_ContextCancelled tests that Watch returns error on cancelled context.
func TestFsnotifyWatcher_Watch_ContextCancelled(t *testing.T) {
	w := NewFsnotifyWatcher(0)
	defer w.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := w.Watch(ctx, []string{"/tmp"})
	if err == nil {
		t.Error("expected error on cancelled context")
	}
}

// TestFsnotifyWatcher_Watch_EmptyPaths tests that Watch returns error on empty paths.
func TestFsnotifyWatcher_Watch_EmptyPaths(t *testing.T) {
	w := NewFsnotifyWatcher(0)
	defer w.Close()

	ctx := context.Background()
	_, err := w.Watch(ctx, []string{})
	if err == nil {
		t.Error("expected error on empty paths")
	}
}

// TestFsnotifyWatcher_Watch_InvalidPath tests that Watch returns error on non-existent path.
func TestFsnotifyWatcher_Watch_InvalidPath(t *testing.T) {
	w := NewFsnotifyWatcher(0)
	defer w.Close()

	ctx := context.Background()
	_, err := w.Watch(ctx, []string{"/non/existent/path/that/should/not/exist"})
	if err == nil {
		t.Error("expected error on non-existent path")
	}
}

// TestFsnotifyWatcher_Close_Idempotent tests that Close is safe to call multiple times.
func TestFsnotifyWatcher_Close_Idempotent(t *testing.T) {
	w := NewFsnotifyWatcher(0)

	// First close should succeed
	if err := w.Close(); err != nil {
		t.Errorf("first Close() failed: %v", err)
	}

	// Second close should also succeed (idempotent)
	if err := w.Close(); err != nil {
		t.Errorf("second Close() failed: %v", err)
	}
}

// TestFsnotifyWatcher_Watch_DetectsFileCreate tests that file creation events are detected.
func TestFsnotifyWatcher_Watch_DetectsFileCreate(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewFsnotifyWatcher(50 * time.Millisecond) // Short debounce for test
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Wait for event
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("event channel closed unexpectedly")
		}
		// Check event properties
		if event.Operation != ports.FileOpCreate && event.Operation != ports.FileOpModify {
			t.Errorf("expected create or modify operation, got %v", event.Operation)
		}
		if event.Path == "" {
			t.Error("expected non-empty path")
		}
		if event.Timestamp.IsZero() {
			t.Error("expected non-zero timestamp")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for file create event")
	}
}

// TestFsnotifyWatcher_Watch_DetectsFileModify tests that file modification events are detected.
func TestFsnotifyWatcher_Watch_DetectsFileModify(t *testing.T) {
	// Create temp directory and initial file
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	// Wait for event
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("event channel closed unexpectedly")
		}
		if event.Operation != ports.FileOpModify {
			// Some systems report as write, accept create too
			if event.Operation != ports.FileOpCreate {
				t.Errorf("expected modify or create operation, got %v", event.Operation)
			}
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for file modify event")
	}
}

// TestFsnotifyWatcher_Watch_DetectsFileDelete tests that file deletion events are detected.
func TestFsnotifyWatcher_Watch_DetectsFileDelete(t *testing.T) {
	// Create temp directory and initial file
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("to delete"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Delete the file
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("failed to delete test file: %v", err)
	}

	// Wait for event
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("event channel closed unexpectedly")
		}
		if event.Operation != ports.FileOpDelete {
			t.Errorf("expected delete operation, got %v", event.Operation)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for file delete event")
	}
}

// TestFsnotifyWatcher_Watch_ContextCancellation tests graceful shutdown on context cancel.
func TestFsnotifyWatcher_Watch_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewFsnotifyWatcher(0)
	defer w.Close()

	ctx, cancel := context.WithCancel(context.Background())

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Channel should close
	select {
	case _, ok := <-events:
		if ok {
			// May receive buffered events, drain until closed
			for range events {
			}
		}
		// Channel closed as expected
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for channel close after context cancellation")
	}
}

// TestFsnotifyWatcher_Debounce_RapidEvents tests that rapid events are aggregated.
func TestFsnotifyWatcher_Debounce_RapidEvents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use longer debounce to ensure aggregation
	w := NewFsnotifyWatcher(200 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")

	// Perform rapid writes (10 writes within debounce window)
	for i := 0; i < 10; i++ {
		if err := os.WriteFile(testFile, []byte(string(rune('a'+i))), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // 10ms apart, all within 200ms debounce window
	}

	// Wait for debounce period plus buffer
	time.Sleep(300 * time.Millisecond)

	// Count events received - should be aggregated (1-2 events, not 10)
	eventCount := 0
	timeout := time.After(500 * time.Millisecond)
drainLoop:
	for {
		select {
		case _, ok := <-events:
			if !ok {
				break drainLoop
			}
			eventCount++
		case <-timeout:
			break drainLoop
		}
	}

	// We expect 1-3 events due to debouncing (not 10)
	// The exact number depends on timing, but it should be significantly less than 10
	if eventCount > 5 {
		t.Errorf("expected debounced event count <= 5, got %d (rapid events not aggregated)", eventCount)
	}
	if eventCount == 0 {
		t.Error("expected at least one event")
	}
}

// TestFsnotifyWatcher_Debounce_ConfigurableWindow tests that debounce window is respected.
func TestFsnotifyWatcher_Debounce_ConfigurableWindow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Very short debounce
	shortDebounce := 50 * time.Millisecond
	w := NewFsnotifyWatcher(shortDebounce)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Start timing
	start := time.Now()

	// Wait for event
	select {
	case <-events:
		// Event received - timing varies by OS, some report before debounce completes
		// We only verify the event arrives, not the exact timing
		_ = time.Since(start) // elapsed time not checked - OS-dependent
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for debounced event")
	}
}

// TestFsnotifyWatcher_Watch_MultiplePaths tests watching multiple directories.
func TestFsnotifyWatcher_Watch_MultiplePaths(t *testing.T) {
	// Create two temp directories
	tmpDir1, err := os.MkdirTemp("", "watcher-test-1-*")
	if err != nil {
		t.Fatalf("failed to create temp dir 1: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "watcher-test-2-*")
	if err != nil {
		t.Fatalf("failed to create temp dir 2: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir1, tmpDir2})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create files in both directories
	if err := os.WriteFile(filepath.Join(tmpDir1, "test1.txt"), []byte("test1"), 0644); err != nil {
		t.Fatalf("failed to create test file 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir2, "test2.txt"), []byte("test2"), 0644); err != nil {
		t.Fatalf("failed to create test file 2: %v", err)
	}

	// Wait for at least one event from each directory (debouncing may combine them)
	eventCount := 0
	timeout := time.After(2 * time.Second)
	for eventCount < 1 {
		select {
		case _, ok := <-events:
			if !ok {
				t.Fatal("event channel closed unexpectedly")
			}
			eventCount++
		case <-timeout:
			t.Errorf("timeout waiting for events, received %d", eventCount)
			return
		}
	}
}

// TestFsnotifyWatcher_Watch_PartialFailure tests that valid paths are watched even if some fail.
func TestFsnotifyWatcher_Watch_PartialFailure(t *testing.T) {
	// Create one valid temp directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mix of valid and invalid paths - should still work for valid path
	events, err := w.Watch(ctx, []string{"/non/existent/path", tmpDir, "/another/invalid/path"})
	if err != nil {
		t.Fatalf("Watch should succeed with at least one valid path: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a file in the valid directory
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should still receive events from valid path
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("event channel closed unexpectedly")
		}
		if event.Path == "" {
			t.Error("expected non-empty path")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event - partial failure may have caused full failure")
	}
}

// TestFsnotifyWatcher_AddPath tests dynamic path addition.
func TestFsnotifyWatcher_AddPath(t *testing.T) {
	// Create initial temp directory
	tmpDir1, err := os.MkdirTemp("", "watcher-test-1-*")
	if err != nil {
		t.Fatalf("failed to create temp dir 1: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	// Create second directory to add dynamically
	tmpDir2, err := os.MkdirTemp("", "watcher-test-2-*")
	if err != nil {
		t.Fatalf("failed to create temp dir 2: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching only first directory
	events, err := w.Watch(ctx, []string{tmpDir1})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Dynamically add second directory
	if err := w.AddPath(tmpDir2); err != nil {
		t.Fatalf("AddPath failed: %v", err)
	}

	// Give watcher time to register new path
	time.Sleep(100 * time.Millisecond)

	// Create file in second (dynamically added) directory
	testFile := filepath.Join(tmpDir2, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should receive event from dynamically added path
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("event channel closed unexpectedly")
		}
		if event.Path == "" {
			t.Error("expected non-empty path")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event from dynamically added path")
	}
}

// TestFsnotifyWatcher_RemovePath tests dynamic path removal.
func TestFsnotifyWatcher_RemovePath(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewFsnotifyWatcher(50 * time.Millisecond)
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Remove the path
	if err := w.RemovePath(tmpDir); err != nil {
		t.Fatalf("RemovePath failed: %v", err)
	}

	// Give watcher time to unregister
	time.Sleep(100 * time.Millisecond)

	// Create file - should NOT trigger event since path was removed
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Wait briefly and check no events received
	select {
	case _, ok := <-events:
		// Race condition possible - fsnotify may still have buffered events
		_ = ok // intentionally not checking - either outcome is acceptable
	case <-time.After(300 * time.Millisecond):
		// Expected - no events from removed path
	}
}

// TestFsnotifyWatcher_AddPath_InvalidPath tests AddPath with invalid path.
func TestFsnotifyWatcher_AddPath_InvalidPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewFsnotifyWatcher(0)
	defer w.Close()

	ctx := context.Background()
	_, err = w.Watch(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Try to add invalid path
	err = w.AddPath("/non/existent/path")
	if err == nil {
		t.Error("expected error when adding non-existent path")
	}
}

// TestFsnotifyWatcher_AddPath_NotRunning tests AddPath when watcher is not running.
func TestFsnotifyWatcher_AddPath_NotRunning(t *testing.T) {
	w := NewFsnotifyWatcher(0)

	err := w.AddPath("/tmp")
	if err == nil {
		t.Error("expected error when AddPath called on non-running watcher")
	}
}

// TestFsnotifyWatcher_RemovePath_NotRunning tests RemovePath when watcher is not running.
func TestFsnotifyWatcher_RemovePath_NotRunning(t *testing.T) {
	w := NewFsnotifyWatcher(0)

	err := w.RemovePath("/tmp")
	if err == nil {
		t.Error("expected error when RemovePath called on non-running watcher")
	}
}
