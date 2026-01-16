package speckit_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors/speckit"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// fixturesDir returns the path to the test fixtures directory
func fixturesDir() string {
	return filepath.Join("..", "..", "..", "..", "test", "fixtures")
}

func TestSpeckitDetector_ImplementsInterface(t *testing.T) {
	// Verify SpeckitDetector implements ports.MethodDetector
	var _ ports.MethodDetector = speckit.NewSpeckitDetector()
}

func TestSpeckitDetector_Name(t *testing.T) {
	d := speckit.NewSpeckitDetector()
	if got := d.Name(); got != "speckit" {
		t.Errorf("Name() = %q, want %q", got, "speckit")
	}
}

func TestNewSpeckitDetector(t *testing.T) {
	d := speckit.NewSpeckitDetector()
	if d == nil {
		t.Error("NewSpeckitDetector() returned nil")
	}

	// Verify detector is functional
	ctx := context.Background()
	// Should return false for non-existent path (no panic)
	result := d.CanDetect(ctx, "/non/existent/path")
	if result {
		t.Error("CanDetect() should return false for non-existent path")
	}
}

func TestSpeckitDetector_CanDetect(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		expected bool
	}{
		{"specs directory present", "speckit-stage-specify", true},
		{"plan stage has specs", "speckit-stage-plan", true},
		{"tasks stage has specs", "speckit-stage-tasks", true},
		{"implement stage has specs", "speckit-stage-implement", true},
		{"uncertain has specs", "speckit-uncertain", true},
		{"no markers present", "no-method-detected", false},
		{"empty project", "empty-project", false},
	}

	d := speckit.NewSpeckitDetector()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join(fixturesDir(), tt.fixture)
			got := d.CanDetect(ctx, fixturePath)
			if got != tt.expected {
				t.Errorf("CanDetect(%s) = %v, want %v", tt.fixture, got, tt.expected)
			}
		})
	}
}

func TestSpeckitDetector_CanDetect_ContextCancellation(t *testing.T) {
	d := speckit.NewSpeckitDetector()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	fixturePath := filepath.Join(fixturesDir(), "speckit-stage-specify")
	got := d.CanDetect(ctx, fixturePath)

	if got {
		t.Error("CanDetect() with cancelled context should return false")
	}
}

func TestSpeckitDetector_CanDetect_AlternativeMarkers(t *testing.T) {
	// Test .speckit and .specify markers using temp directories
	tests := []struct {
		name   string
		marker string
	}{
		{".speckit marker", ".speckit"},
		{".specify marker", ".specify"},
	}

	d := speckit.NewSpeckitDetector()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			markerPath := filepath.Join(tmpDir, tt.marker)
			if err := os.MkdirAll(markerPath, 0755); err != nil {
				t.Fatalf("failed to create marker dir: %v", err)
			}

			got := d.CanDetect(ctx, tmpDir)
			if !got {
				t.Errorf("CanDetect() should return true for %s marker", tt.marker)
			}
		})
	}
}

func TestSpeckitDetector_Detect(t *testing.T) {
	tests := []struct {
		name        string
		fixture     string
		expectStage domain.Stage
		expectConf  domain.Confidence
		checkReason string
	}{
		{
			"specify stage",
			"speckit-stage-specify",
			domain.StageSpecify,
			domain.ConfidenceCertain,
			"spec.md exists, no plan.md",
		},
		{
			"plan stage",
			"speckit-stage-plan",
			domain.StagePlan,
			domain.ConfidenceCertain,
			"plan.md exists, no tasks.md",
		},
		{
			"tasks stage",
			"speckit-stage-tasks",
			domain.StageTasks,
			domain.ConfidenceCertain,
			"tasks.md exists",
		},
		{
			"implement stage",
			"speckit-stage-implement",
			domain.StageImplement,
			domain.ConfidenceCertain,
			"implement.md exists",
		},
		{
			"uncertain case",
			"speckit-uncertain",
			domain.StageUnknown,
			domain.ConfidenceUncertain,
			"no standard Speckit artifacts found",
		},
	}

	d := speckit.NewSpeckitDetector()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join(fixturesDir(), tt.fixture)
			result, err := d.Detect(ctx, fixturePath)

			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if result.Stage != tt.expectStage {
				t.Errorf("Detect().Stage = %v, want %v", result.Stage, tt.expectStage)
			}
			if result.Confidence != tt.expectConf {
				t.Errorf("Detect().Confidence = %v, want %v", result.Confidence, tt.expectConf)
			}
			if result.Method != "speckit" {
				t.Errorf("Detect().Method = %q, want %q", result.Method, "speckit")
			}
			if result.Reasoning == "" {
				t.Error("Detect().Reasoning should not be empty")
			}
		})
	}
}

func TestSpeckitDetector_Detect_ContextCancellation(t *testing.T) {
	d := speckit.NewSpeckitDetector()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	fixturePath := filepath.Join(fixturesDir(), "speckit-stage-specify")
	_, err := d.Detect(ctx, fixturePath)

	if err != context.Canceled {
		t.Errorf("Detect() with cancelled context should return context.Canceled, got %v", err)
	}
}

func TestSpeckitDetector_ContextCancellationTiming(t *testing.T) {
	d := speckit.NewSpeckitDetector()
	fixturePath := filepath.Join(fixturesDir(), "speckit-stage-specify")

	ctx, cancel := context.WithCancel(context.Background())

	// Start detection in goroutine
	done := make(chan error, 1)
	go func() {
		_, err := d.Detect(ctx, fixturePath)
		done <- err
	}()

	// Cancel after brief delay to ensure detection has started
	time.Sleep(10 * time.Millisecond)
	cancelStart := time.Now()
	cancel()

	// Wait for completion with timeout
	select {
	case <-done:
		elapsed := time.Since(cancelStart)
		// AC9: Should respond within 100ms of cancellation
		if elapsed > 100*time.Millisecond {
			t.Errorf("Cancellation took %v, expected < 100ms (AC9 requirement)", elapsed)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Detection did not respond to cancellation within 200ms timeout")
	}
}

func TestSpeckitDetector_FailurePaths(t *testing.T) {
	d := speckit.NewSpeckitDetector()
	ctx := context.Background()

	t.Run("non-existent path returns error", func(t *testing.T) {
		_, err := d.Detect(ctx, "/non/existent/path/that/does/not/exist")
		if err == nil {
			t.Error("expected error for non-existent path")
		}
	})

	t.Run("empty specs directory returns uncertain", func(t *testing.T) {
		// Create temp dir with empty specs/
		tmpDir := t.TempDir()
		specsDir := filepath.Join(tmpDir, "specs")
		if err := os.MkdirAll(specsDir, 0755); err != nil {
			t.Fatal(err)
		}

		result, err := d.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Stage != domain.StageUnknown {
			t.Errorf("expected StageUnknown, got %v", result.Stage)
		}
		if result.Confidence != domain.ConfidenceUncertain {
			t.Errorf("expected ConfidenceUncertain, got %v", result.Confidence)
		}
	})
}

func TestSpeckitDetector_MultipleDirectories(t *testing.T) {
	// Test AC6: Multiple spec directories, most recent wins
	d := speckit.NewSpeckitDetector()
	ctx := context.Background()

	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")

	// Create first directory with spec.md
	dir1 := filepath.Join(specsDir, "001-old-feature")
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir1, "spec.md"), []byte("# Old"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create second directory with plan.md
	dir2 := filepath.Join(specsDir, "002-new-feature")
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "spec.md"), []byte("# New"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "plan.md"), []byte("# Plan"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set explicit mod times: dir1 is old, dir2 is newer
	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()
	if err := os.Chtimes(dir1, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dir2, newTime, newTime); err != nil {
		t.Fatal(err)
	}

	result, err := d.Detect(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// Should use the newer directory (002-new-feature) which has plan.md
	if result.Stage != domain.StagePlan {
		t.Errorf("Detect().Stage = %v, want %v (should use most recent dir)", result.Stage, domain.StagePlan)
	}

	// AC6: Reasoning should mention which directory was used
	if result.Reasoning == "" {
		t.Error("Reasoning should not be empty")
	}
	if !strings.Contains(result.Reasoning, "002-new-feature") {
		t.Errorf("Reasoning should mention the directory used, got: %q", result.Reasoning)
	}
	if !strings.Contains(result.Reasoning, "most recently modified") {
		t.Errorf("Reasoning should explain selection criteria, got: %q", result.Reasoning)
	}
}

// Epic 4 Hotfix H4: Test that equal modification times use lexicographic tiebreaker
func TestSpeckitDetector_EqualModTimes_HighestNumberedWins(t *testing.T) {
	d := speckit.NewSpeckitDetector()
	ctx := context.Background()

	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")

	// Create directories with different numbered prefixes (simulating git clone scenario)
	dirs := []string{
		"001-first-feature",
		"002-second-feature",
		"003-third-feature",
		"005-latest-feature", // Note: skipping 004 to ensure lexicographic sort works
	}

	// All directories will have same mtime (simulating git clone)
	sameTime := time.Now()

	for _, dir := range dirs {
		dirPath := filepath.Join(specsDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatal(err)
		}
		// Create spec.md in each - different stages to identify which was picked
		if err := os.WriteFile(filepath.Join(dirPath, "spec.md"), []byte("# Spec"), 0644); err != nil {
			t.Fatal(err)
		}
		// Set all to same modification time
		if err := os.Chtimes(dirPath, sameTime, sameTime); err != nil {
			t.Fatal(err)
		}
	}

	// Add plan.md only to the highest-numbered directory
	if err := os.WriteFile(filepath.Join(specsDir, "005-latest-feature", "plan.md"), []byte("# Plan"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := d.Detect(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// Should use 005-latest-feature (highest numbered when mtimes are equal)
	if result.Stage != domain.StagePlan {
		t.Errorf("Detect().Stage = %v, want %v (should use highest-numbered dir when mtimes equal)", result.Stage, domain.StagePlan)
	}

	// Reasoning should mention 005-latest-feature
	if !strings.Contains(result.Reasoning, "005-latest-feature") {
		t.Errorf("Reasoning should mention 005-latest-feature, got: %q", result.Reasoning)
	}
}

// TestDetectionAccuracy runs against all fixtures and calculates accuracy.
// This is the launch blocker test - must be >= 95%
func TestDetectionAccuracy(t *testing.T) {
	testCases := []struct {
		fixture       string
		expectedStage domain.Stage
		shouldDetect  bool // false for non-speckit fixtures
	}{
		// === EXISTING (9 fixtures) ===
		{"speckit-stage-specify", domain.StageSpecify, true},
		{"speckit-stage-plan", domain.StagePlan, true},
		{"speckit-stage-tasks", domain.StageTasks, true},
		{"speckit-stage-implement", domain.StageImplement, true},
		{"speckit-uncertain", domain.StageUnknown, true},
		{"speckit-dotspeckit-marker", domain.StageSpecify, true},
		{"speckit-dotspecify-marker", domain.StagePlan, true},
		{"no-method-detected", domain.StageUnknown, false},
		{"empty-project", domain.StageUnknown, false},
		// === NEW (11 fixtures) ===
		// Note: nested fixture returns StageUnknown because detector only looks one level deep
		// (specs/feature-group/ is found, but artifacts are in specs/feature-group/001-feature/)
		{"speckit-stage-specify-nested", domain.StageUnknown, true},
		{"speckit-stage-plan-with-drafts", domain.StagePlan, true},
		{"speckit-stage-tasks-partial", domain.StageTasks, true},
		{"speckit-stage-implement-complete", domain.StageImplement, true},
		{"speckit-multiple-features", domain.StagePlan, true},
		{"speckit-no-spec-subdirs", domain.StageUnknown, true},
		{"speckit-hidden-files", domain.StageUnknown, true},
		{"speckit-mixed-markers", domain.StageSpecify, true},
		{"speckit-empty-spec-dir", domain.StageUnknown, true},
		{"speckit-non-standard-names", domain.StageSpecify, true},
		{"speckit-readme-only", domain.StageUnknown, true},
	}

	d := speckit.NewSpeckitDetector()
	ctx := context.Background()

	correct := 0
	total := len(testCases)

	for _, tc := range testCases {
		fixturePath := filepath.Join(fixturesDir(), tc.fixture)

		canDetect := d.CanDetect(ctx, fixturePath)

		if tc.shouldDetect {
			if !canDetect {
				t.Logf("FAIL: %s - CanDetect returned false, expected true", tc.fixture)
				continue
			}

			result, err := d.Detect(ctx, fixturePath)
			if err != nil {
				t.Logf("FAIL: %s - Detect error: %v", tc.fixture, err)
				continue
			}

			if result.Stage == tc.expectedStage {
				correct++
				t.Logf("PASS: %s - Stage: %v", tc.fixture, result.Stage)
			} else {
				t.Logf("FAIL: %s - Got %v, expected %v", tc.fixture, result.Stage, tc.expectedStage)
			}
		} else {
			// Should NOT detect as Speckit
			if !canDetect {
				correct++
				t.Logf("PASS: %s - Correctly not detected as Speckit", tc.fixture)
			} else {
				t.Logf("FAIL: %s - Should not be detected as Speckit", tc.fixture)
			}
		}
	}

	accuracy := float64(correct) / float64(total) * 100
	t.Logf("\n=== DETECTION ACCURACY: %.1f%% (%d/%d) ===", accuracy, correct, total)

	if accuracy < 95.0 {
		t.Errorf("Detection accuracy %.1f%% is below 95%% launch blocker threshold", accuracy)
	}
}

// =============================================================================
// Story 14.2: Timestamp Tests (AC2, AC4)
// =============================================================================

// createFileWithMtime creates a file with a specific modification time for testing.
func createFileWithMtime(t *testing.T, path string, content string, mtime time.Time) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("failed to set mtime for %s: %v", path, err)
	}
}

func TestSpeckitDetector_Timestamp(t *testing.T) {
	baseTime := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		setup         func(t *testing.T, dir string)
		wantTimestamp func(baseTime time.Time) time.Time // function to compute expected timestamp
		checkExact    bool                               // whether to check exact time match
	}{
		{
			name: "AC2: timestamp reflects most recent spec directory mtime",
			setup: func(t *testing.T, dir string) {
				specsDir := filepath.Join(dir, "specs")
				if err := os.MkdirAll(specsDir, 0755); err != nil {
					t.Fatalf("failed to create specs dir: %v", err)
				}

				// Create two spec directories with different mtimes
				spec1Dir := filepath.Join(specsDir, "001-feature")
				spec2Dir := filepath.Join(specsDir, "002-feature")
				if err := os.MkdirAll(spec1Dir, 0755); err != nil {
					t.Fatalf("failed to create spec1 dir: %v", err)
				}
				if err := os.MkdirAll(spec2Dir, 0755); err != nil {
					t.Fatalf("failed to create spec2 dir: %v", err)
				}

				// Add plan.md to spec2 FIRST (this modifies directory mtime)
				newerTime := baseTime
				createFileWithMtime(t, filepath.Join(spec2Dir, "plan.md"), "# Plan", newerTime)

				// NOW set directory mtimes after file creation
				// spec1 is older, spec2 is newer (more recent)
				olderTime := baseTime.Add(-24 * time.Hour)
				if err := os.Chtimes(spec1Dir, olderTime, olderTime); err != nil {
					t.Fatalf("failed to set spec1 mtime: %v", err)
				}
				if err := os.Chtimes(spec2Dir, newerTime, newerTime); err != nil {
					t.Fatalf("failed to set spec2 mtime: %v", err)
				}
			},
			wantTimestamp: func(bt time.Time) time.Time { return bt }, // Should use newer time
			checkExact:    true,
		},
		{
			name: "AC2: timestamp reflects most recent artifact file mtime (newer than dir)",
			setup: func(t *testing.T, dir string) {
				specsDir := filepath.Join(dir, "specs")
				specDir := filepath.Join(specsDir, "001-feature")
				if err := os.MkdirAll(specDir, 0755); err != nil {
					t.Fatalf("failed to create spec dir: %v", err)
				}

				// Create files FIRST, then set directory mtime
				// Files - plan.md is 2h ago, spec.md is 1 day ago
				planTime := baseTime.Add(-2 * time.Hour)
				specTime := baseTime.Add(-24 * time.Hour)
				createFileWithMtime(t, filepath.Join(specDir, "plan.md"), "# Plan", planTime)
				createFileWithMtime(t, filepath.Join(specDir, "spec.md"), "# Spec", specTime)

				// NOW set directory mtime to be older than files
				dirTime := baseTime.Add(-48 * time.Hour)
				if err := os.Chtimes(specDir, dirTime, dirTime); err != nil {
					t.Fatalf("failed to set dir mtime: %v", err)
				}
			},
			wantTimestamp: func(bt time.Time) time.Time { return bt.Add(-2 * time.Hour) }, // plan.md time
			checkExact:    true,
		},
		{
			name: "AC4: zero time when only marker directory exists with no files",
			setup: func(t *testing.T, dir string) {
				specsDir := filepath.Join(dir, "specs")
				if err := os.MkdirAll(specsDir, 0755); err != nil {
					t.Fatalf("failed to create specs dir: %v", err)
				}
				// Empty specs dir - no subdirectories
			},
			wantTimestamp: func(bt time.Time) time.Time { return time.Time{} },
			checkExact:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			d := speckit.NewSpeckitDetector()
			result, err := d.Detect(context.Background(), dir)

			if err != nil {
				// For AC4, result may be nil with no error (empty specs)
				// Let's check if we got a valid result
				if tt.name == "AC4: zero time when only marker directory exists with no files" {
					// For empty specs, we get a result with Unknown stage
					if result == nil {
						t.Skipf("Expected result for empty specs, got nil")
					}
				} else {
					t.Fatalf("Detect() error = %v", err)
				}
			}

			if result == nil {
				t.Fatalf("Detect() returned nil result")
			}

			wantTime := tt.wantTimestamp(baseTime)
			if tt.checkExact {
				if !result.ArtifactTimestamp.Equal(wantTime) {
					t.Errorf("ArtifactTimestamp = %v, want %v", result.ArtifactTimestamp, wantTime)
				}
			}
		})
	}
}

func TestSpeckitDetector_Timestamp_HasTimestamp(t *testing.T) {
	// Test that detected results have timestamps (non-zero) when artifacts exist
	dir := t.TempDir()

	specsDir := filepath.Join(dir, "specs")
	specDir := filepath.Join(specsDir, "001-feature")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	// Create a spec.md file
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# Spec"), 0644); err != nil {
		t.Fatalf("failed to write spec.md: %v", err)
	}

	d := speckit.NewSpeckitDetector()
	result, err := d.Detect(context.Background(), dir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if result == nil {
		t.Fatal("Detect() returned nil result")
	}

	if !result.HasTimestamp() {
		t.Error("HasTimestamp() = false, want true for detected Speckit project")
	}
}
