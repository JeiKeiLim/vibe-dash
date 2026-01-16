package agentdetectors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// TestNewGenericDetector_DefaultThreshold tests the default threshold is 10 minutes.
func TestNewGenericDetector_DefaultThreshold(t *testing.T) {
	g := NewGenericDetector()
	if g.threshold != 10*time.Minute {
		t.Errorf("default threshold = %v, want %v", g.threshold, 10*time.Minute)
	}
}

// TestNewGenericDetector_WithThreshold tests the custom threshold option.
func TestNewGenericDetector_WithThreshold(t *testing.T) {
	g := NewGenericDetector(WithThreshold(5 * time.Minute))
	if g.threshold != 5*time.Minute {
		t.Errorf("custom threshold = %v, want %v", g.threshold, 5*time.Minute)
	}
}

// TestNewGenericDetector_WithThreshold_Invalid tests invalid threshold values are ignored.
func TestNewGenericDetector_WithThreshold_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		threshold time.Duration
		want      time.Duration
	}{
		{"zero", 0, 10 * time.Minute},
		{"negative", -5 * time.Minute, 10 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenericDetector(WithThreshold(tt.threshold))
			if g.threshold != tt.want {
				t.Errorf("threshold = %v, want %v", g.threshold, tt.want)
			}
		})
	}
}

// TestNewGenericDetector_WithNow tests custom time function option.
func TestNewGenericDetector_WithNow(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	mockNow := func() time.Time { return fixedTime }

	g := NewGenericDetector(WithNow(mockNow))
	if g.now() != fixedTime {
		t.Errorf("now() = %v, want %v", g.now(), fixedTime)
	}
}

// TestNewGenericDetector_WithNow_Nil tests nil time function is ignored.
func TestNewGenericDetector_WithNow_Nil(t *testing.T) {
	g := NewGenericDetector(WithNow(nil))
	// Should still have a valid now function (default)
	if g.now == nil {
		t.Error("now function should not be nil")
	}
	// Check it returns reasonable time (close to actual now)
	diff := time.Since(g.now())
	if diff > time.Second || diff < -time.Second {
		t.Errorf("now() should return current time, diff = %v", diff)
	}
}

// Interface compliance test (compile-time)
var _ ports.AgentActivityDetector = (*GenericDetector)(nil)

// TestGenericDetector_Name tests Name() returns "Generic".
func TestGenericDetector_Name(t *testing.T) {
	g := NewGenericDetector()
	if got := g.Name(); got != "Generic" {
		t.Errorf("Name() = %q, want %q", got, "Generic")
	}
}

// TestDetect_RecentActivity_Working tests that recent file activity returns Working state (AC3).
func TestDetect_RecentActivity_Working(t *testing.T) {
	tmpDir := t.TempDir()
	// Create file with recent mtime (now)
	if err := os.WriteFile(filepath.Join(tmpDir, "recent.go"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	g := NewGenericDetector(WithThreshold(10 * time.Minute))
	state, err := g.Detect(context.Background(), tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Status != domain.AgentWorking {
		t.Errorf("Status = %v, want AgentWorking", state.Status)
	}
	if state.Confidence != domain.ConfidenceUncertain {
		t.Errorf("Confidence = %v, want ConfidenceUncertain", state.Confidence)
	}
	if state.Tool != "Generic" {
		t.Errorf("Tool = %q, want %q", state.Tool, "Generic")
	}
}

// TestDetect_OldActivity_WaitingForUser tests that old file activity returns WaitingForUser state (AC2).
func TestDetect_OldActivity_WaitingForUser(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "old.go")
	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set file mtime to 15 minutes ago
	oldTime := time.Now().Add(-15 * time.Minute)
	if err := os.Chtimes(filePath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set file time: %v", err)
	}

	g := NewGenericDetector(WithThreshold(10 * time.Minute))
	state, err := g.Detect(context.Background(), tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Status != domain.AgentWaitingForUser {
		t.Errorf("Status = %v, want AgentWaitingForUser", state.Status)
	}
	if state.Confidence != domain.ConfidenceUncertain {
		t.Errorf("Confidence = %v, want ConfidenceUncertain", state.Confidence)
	}
}

// TestDetect_NonexistentPath_Unknown tests non-existent path returns Unknown (AC7).
func TestDetect_NonexistentPath_Unknown(t *testing.T) {
	g := NewGenericDetector()
	state, err := g.Detect(context.Background(), "/nonexistent/path/12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v, want AgentUnknown", state.Status)
	}
	if state.Confidence != domain.ConfidenceUncertain {
		t.Errorf("Confidence = %v, want ConfidenceUncertain", state.Confidence)
	}
}

// TestGenericDetector_ContextCancelled_ReturnsPromptly tests context cancellation returns within 100ms (AC6).
func TestGenericDetector_ContextCancelled_ReturnsPromptly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	g := NewGenericDetector()

	start := time.Now()
	state, err := g.Detect(ctx, "/some/path")
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Errorf("Detect took %v, want < 100ms", elapsed)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v, want AgentUnknown", state.Status)
	}
}

// TestDetect_CustomThreshold tests custom threshold via WithThreshold option (AC5).
func TestDetect_CustomThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set file mtime to 3 minutes ago
	oldTime := time.Now().Add(-3 * time.Minute)
	if err := os.Chtimes(filePath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set file time: %v", err)
	}

	// Default threshold (10 min) → Working
	g1 := NewGenericDetector()
	state1, _ := g1.Detect(context.Background(), tmpDir)
	if state1.Status != domain.AgentWorking {
		t.Errorf("Default threshold: Status = %v, want AgentWorking", state1.Status)
	}

	// Custom threshold (2 min) → WaitingForUser
	g2 := NewGenericDetector(WithThreshold(2 * time.Minute))
	state2, _ := g2.Detect(context.Background(), tmpDir)
	if state2.Status != domain.AgentWaitingForUser {
		t.Errorf("Custom threshold: Status = %v, want AgentWaitingForUser", state2.Status)
	}
}

// TestDetect_HiddenFilesSkipped tests that hidden files are skipped.
func TestDetect_HiddenFilesSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only hidden files
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create hidden file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".git", "HEAD"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create .git/HEAD: %v", err)
	}

	g := NewGenericDetector()
	state, err := g.Detect(context.Background(), tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be Unknown since all files are hidden
	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v, want AgentUnknown (hidden files skipped)", state.Status)
	}
}

// TestDetect_WithNow tests duration calculation using WithNow option for time injection.
func TestDetect_WithNow(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Mock time to be 15 minutes in the future
	mockNow := func() time.Time {
		return time.Now().Add(15 * time.Minute)
	}

	g := NewGenericDetector(WithNow(mockNow), WithThreshold(10*time.Minute))
	state, _ := g.Detect(context.Background(), tmpDir)

	if state.Status != domain.AgentWaitingForUser {
		t.Errorf("Status = %v, want AgentWaitingForUser", state.Status)
	}
}

// TestDetect_DurationCalculation tests that duration is approximately correct (AC8).
func TestDetect_DurationCalculation(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set file mtime to exactly 5 minutes ago
	targetDuration := 5 * time.Minute
	oldTime := time.Now().Add(-targetDuration)
	if err := os.Chtimes(filePath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set file time: %v", err)
	}

	g := NewGenericDetector(WithThreshold(10 * time.Minute))
	state, err := g.Detect(context.Background(), tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Duration should be approximately 5 minutes (within 1 second tolerance)
	if state.Duration < targetDuration-time.Second || state.Duration > targetDuration+time.Second {
		t.Errorf("Duration = %v, want approximately %v", state.Duration, targetDuration)
	}
}

// TestDetect_ConfidenceAlwaysUncertain tests all states return Uncertain confidence (AC1-8).
func TestDetect_ConfidenceAlwaysUncertain(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() string // returns path to test
		status domain.AgentStatus
	}{
		{
			name: "working",
			setup: func() string {
				tmpDir := t.TempDir()
				_ = os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte(""), 0644)
				return tmpDir
			},
			status: domain.AgentWorking,
		},
		{
			name: "waiting",
			setup: func() string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "file.go")
				_ = os.WriteFile(filePath, []byte(""), 0644)
				oldTime := time.Now().Add(-15 * time.Minute)
				_ = os.Chtimes(filePath, oldTime, oldTime)
				return tmpDir
			},
			status: domain.AgentWaitingForUser,
		},
		{
			name: "unknown",
			setup: func() string {
				return "/nonexistent/path"
			},
			status: domain.AgentUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			g := NewGenericDetector(WithThreshold(10 * time.Minute))
			state, _ := g.Detect(context.Background(), path)

			if state.Confidence != domain.ConfidenceUncertain {
				t.Errorf("Confidence = %v, want ConfidenceUncertain", state.Confidence)
			}
			if state.Status != tt.status {
				t.Errorf("Status = %v, want %v", state.Status, tt.status)
			}
		})
	}
}

// TestDetect_EmptyDirectory_Unknown tests empty directory returns Unknown.
func TestDetect_EmptyDirectory_Unknown(t *testing.T) {
	tmpDir := t.TempDir()
	// Empty directory - no files

	g := NewGenericDetector()
	state, err := g.Detect(context.Background(), tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Status != domain.AgentUnknown {
		t.Errorf("Status = %v, want AgentUnknown", state.Status)
	}
}

// TestDetect_SingleFile tests detecting on a single file path.
func TestDetect_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "single.go")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	g := NewGenericDetector(WithThreshold(10 * time.Minute))
	state, err := g.Detect(context.Background(), filePath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Status != domain.AgentWorking {
		t.Errorf("Status = %v, want AgentWorking", state.Status)
	}
}

// TestDetect_MostRecentFileWins tests that the most recent file determines the state (AC8).
func TestDetect_MostRecentFileWins(t *testing.T) {
	tmpDir := t.TempDir()

	// Create old file (15 minutes ago)
	oldFile := filepath.Join(tmpDir, "old.go")
	if err := os.WriteFile(oldFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create old file: %v", err)
	}
	oldTime := time.Now().Add(-15 * time.Minute)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set old file time: %v", err)
	}

	// Create recent file (just now)
	if err := os.WriteFile(filepath.Join(tmpDir, "recent.go"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create recent file: %v", err)
	}

	g := NewGenericDetector(WithThreshold(10 * time.Minute))
	state, err := g.Detect(context.Background(), tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Most recent file is recent, so should be Working
	if state.Status != domain.AgentWorking {
		t.Errorf("Status = %v, want AgentWorking (most recent file wins)", state.Status)
	}
}

// TestDetect_FutureTimestamp_ClampsToZero tests future timestamps are handled (clock skew).
func TestDetect_FutureTimestamp_ClampsToZero(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "future.go")
	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set file mtime to 10 minutes in the future
	futureTime := time.Now().Add(10 * time.Minute)
	if err := os.Chtimes(filePath, futureTime, futureTime); err != nil {
		t.Fatalf("failed to set file time: %v", err)
	}

	g := NewGenericDetector(WithThreshold(10 * time.Minute))
	state, err := g.Detect(context.Background(), tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Future timestamp should be clamped to 0 duration, so status is Working
	if state.Status != domain.AgentWorking {
		t.Errorf("Status = %v, want AgentWorking (future timestamp clamped)", state.Status)
	}
	// Duration should be 0 or very small, not negative
	if state.Duration < 0 {
		t.Errorf("Duration = %v, want >= 0 (clamped)", state.Duration)
	}
	if state.Duration > time.Second {
		t.Errorf("Duration = %v, want approximately 0 (clamped)", state.Duration)
	}
}

// TestDetect_ContextCancelledDuringWalk tests cancellation during filesystem walk (AC6).
func TestDetect_ContextCancelledDuringWalk(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many files to ensure walk takes some time
	for i := 0; i < 150; i++ {
		if err := os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i)), []byte(""), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a small delay (allow walk to start)
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	g := NewGenericDetector()
	start := time.Now()
	state, err := g.Detect(ctx, tmpDir)
	elapsed := time.Since(start)

	// Should complete promptly (much less than full walk time)
	if elapsed > 200*time.Millisecond {
		t.Errorf("Detect took %v during cancel, want < 200ms", elapsed)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// State could be Unknown (cancelled early) or Working/Waiting (if completed before cancel)
	// We just verify it returned without error and reasonably fast
	_ = state
}

// TestDetect_NestedFiles tests detection works with nested directory structure.
func TestDetect_NestedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "src", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}

	// Create old file at root
	rootFile := filepath.Join(tmpDir, "root.go")
	if err := os.WriteFile(rootFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create root file: %v", err)
	}
	oldTime := time.Now().Add(-15 * time.Minute)
	if err := os.Chtimes(rootFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set root file time: %v", err)
	}

	// Create recent file in nested dir
	if err := os.WriteFile(filepath.Join(subDir, "nested.go"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	g := NewGenericDetector(WithThreshold(10 * time.Minute))
	state, err := g.Detect(context.Background(), tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Most recent file is in nested dir, so should be Working
	if state.Status != domain.AgentWorking {
		t.Errorf("Status = %v, want AgentWorking (nested file detected)", state.Status)
	}
}
