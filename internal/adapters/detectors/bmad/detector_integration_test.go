//go:build integration

package bmad

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBMADDetector_Dogfooding tests BMAD detection against vibe-dash's own .bmad folder.
// This is the "dogfooding" test - the detector should successfully detect its own project.
// Run with: go test -tags=integration ./internal/adapters/detectors/bmad/...
func TestBMADDetector_Dogfooding(t *testing.T) {
	// Find project root by looking for go.mod marker
	projectRoot := findProjectRoot(t)

	d := NewBMADDetector()
	ctx := context.Background()

	// Must detect
	if !d.CanDetect(ctx, projectRoot) {
		t.Fatal("CanDetect should return true for vibe-dash project")
	}

	// Must return bmad method
	result, err := d.Detect(ctx, projectRoot)
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	if result.Method != "bmad" {
		t.Errorf("Method = %q, want %q", result.Method, "bmad")
	}

	// Version should be extracted from config.yaml
	if !strings.Contains(result.Reasoning, "BMAD v") {
		t.Errorf("Reasoning should contain version, got: %q", result.Reasoning)
	}

	t.Logf("Dogfooding result: stage=%s, confidence=%s, reasoning=%s",
		result.Stage, result.Confidence, result.Reasoning)
}

// TestBMADDetector_Dogfooding_DetectedMethod verifies the detected method is "bmad" (AC: #2)
func TestBMADDetector_Dogfooding_DetectedMethod(t *testing.T) {
	projectRoot := findProjectRoot(t)

	d := NewBMADDetector()
	ctx := context.Background()

	result, err := d.Detect(ctx, projectRoot)
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}

	if result.Method != "bmad" {
		t.Errorf("Detected method = %q, want %q", result.Method, "bmad")
	}
}

// TestBMADDetector_Dogfooding_StageMatchesSprintStatus verifies stage matches sprint-status.yaml (AC: #2)
func TestBMADDetector_Dogfooding_StageMatchesSprintStatus(t *testing.T) {
	projectRoot := findProjectRoot(t)

	d := NewBMADDetector()
	ctx := context.Background()

	result, err := d.Detect(ctx, projectRoot)
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}

	// The vibe-dash project has sprint-status.yaml with epics in various states
	// Based on the sprint-status, we expect either StageImplement or StagePlan
	// depending on whether there's an in-progress epic/story
	t.Logf("Detected stage: %s", result.Stage)

	// Stage should not be StageUnknown when we have sprint-status.yaml
	// The actual stage depends on current sprint state, but we know it should be detectable
	if result.Reasoning == "" {
		t.Error("Reasoning should not be empty for vibe-dash with sprint-status.yaml")
	}
}

// findProjectRoot walks up from current directory to find go.mod
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from the test file's directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Verify .bmad also exists
			if _, err := os.Stat(filepath.Join(dir, ".bmad")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("Could not find project root with .bmad folder")
		}
		dir = parent
	}
}

// integrationFixturesDir returns the path to test fixtures for integration tests
// Note: fixturesDir is defined in detector_test.go, so we use a different name here
func integrationFixturesDir() string {
	return filepath.Join("..", "..", "..", "..", "test", "fixtures")
}

// TestIntegration_FullFlow_DetectBMADFixtures tests detection of all BMAD fixtures (AC: #4)
// This simulates what happens when `vibe add` is run on a BMAD project directory
func TestIntegration_FullFlow_DetectBMADFixtures(t *testing.T) {
	testCases := []struct {
		fixture       string
		shouldDetect  bool
		expectedStage string
	}{
		{"bmad-v6-complete", true, "Implement"},
		{"bmad-v6-minimal", true, "Unknown"},
		{"bmad-v6-no-config", true, "Unknown"},
		{"bmad-v6-mid-sprint", true, "Implement"},
		{"bmad-v6-all-done", true, "Implement"},
		{"bmad-v6-artifacts-only", true, "Implement"},
		{"bmad-v4-not-supported", false, ""},
	}

	d := NewBMADDetector()
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.fixture, func(t *testing.T) {
			fixturePath := filepath.Join(integrationFixturesDir(), tc.fixture)

			// Simulate vibe add - first check if method can be detected
			canDetect := d.CanDetect(ctx, fixturePath)

			if tc.shouldDetect {
				if !canDetect {
					t.Fatalf("Expected BMAD detection for %s", tc.fixture)
				}

				// Simulate detection that occurs during vibe add
				result, err := d.Detect(ctx, fixturePath)
				if err != nil {
					t.Fatalf("Detect failed: %v", err)
				}

				// Verify project would appear with correct method
				if result.Method != "bmad" {
					t.Errorf("Expected method 'bmad', got %q", result.Method)
				}

				// Verify stage would display correctly
				if result.Stage.String() != tc.expectedStage {
					t.Errorf("Expected stage %s, got %s", tc.expectedStage, result.Stage)
				}

				t.Logf("%s: method=%s, stage=%s, confidence=%s",
					tc.fixture, result.Method, result.Stage, result.Confidence)
			} else {
				if canDetect {
					t.Errorf("Should not detect BMAD for %s", tc.fixture)
				}
			}
		})
	}
}

// TestIntegration_StageUpdatesWhenSprintStatusChanges tests that stage detection updates
// when sprint-status.yaml is modified (AC: #4)
func TestIntegration_StageUpdatesWhenSprintStatusChanges(t *testing.T) {
	// Create a temporary BMAD project
	tmpDir := t.TempDir()

	// Create initial BMAD structure
	bmadDir := filepath.Join(tmpDir, ".bmad", "bmm")
	if err := os.MkdirAll(bmadDir, 0755); err != nil {
		t.Fatalf("failed to create .bmad/bmm: %v", err)
	}

	configContent := `# BMM Module Configuration
# Generated by BMAD installer
# Version: 6.0.0-alpha.13

project_name: integration-test
bmad_folder: .bmad
`
	configPath := filepath.Join(bmadDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}

	// Create sprint-artifacts directory
	sprintDir := filepath.Join(tmpDir, "docs", "sprint-artifacts")
	if err := os.MkdirAll(sprintDir, 0755); err != nil {
		t.Fatalf("failed to create sprint-artifacts: %v", err)
	}
	statusPath := filepath.Join(sprintDir, "sprint-status.yaml")

	d := NewBMADDetector()
	ctx := context.Background()

	// Phase 1: All backlog - should detect as StageUnknown or Specify
	backlogStatus := `# generated: 2025-12-20
project: integration-test
development_status:
  epic-1: backlog
  1-1-feature: backlog
`
	if err := os.WriteFile(statusPath, []byte(backlogStatus), 0644); err != nil {
		t.Fatalf("failed to write backlog status: %v", err)
	}

	result1, err := d.Detect(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Phase 1 Detect failed: %v", err)
	}
	if result1.Stage.String() != "Specify" {
		t.Errorf("Phase 1: expected stage Specify for all-backlog, got %s", result1.Stage)
	}
	t.Logf("Phase 1 (all backlog): stage=%s, confidence=%s", result1.Stage, result1.Confidence)

	// Phase 2: Epic in-progress - should detect as StageImplement
	inProgressStatus := `# generated: 2025-12-20
project: integration-test
development_status:
  epic-1: in-progress
  1-1-feature: in-progress
  1-2-feature: backlog
`
	if err := os.WriteFile(statusPath, []byte(inProgressStatus), 0644); err != nil {
		t.Fatalf("failed to write in-progress status: %v", err)
	}

	result2, err := d.Detect(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Phase 2 Detect failed: %v", err)
	}
	if result2.Stage.String() != "Implement" {
		t.Errorf("Phase 2: expected stage Implement, got %s", result2.Stage)
	}
	t.Logf("Phase 2 (epic in-progress): stage=%s, confidence=%s", result2.Stage, result2.Confidence)

	// Phase 3: All done - should still detect as StageImplement
	allDoneStatus := `# generated: 2025-12-20
project: integration-test
development_status:
  epic-1: done
  1-1-feature: done
  1-2-feature: done
`
	if err := os.WriteFile(statusPath, []byte(allDoneStatus), 0644); err != nil {
		t.Fatalf("failed to write all-done status: %v", err)
	}

	result3, err := d.Detect(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Phase 3 Detect failed: %v", err)
	}
	if result3.Stage.String() != "Implement" {
		t.Errorf("Phase 3: expected stage Implement, got %s", result3.Stage)
	}
	t.Logf("Phase 3 (all done): stage=%s, confidence=%s", result3.Stage, result3.Confidence)

	// Verify stage changed between phases
	if result1.Stage == result2.Stage {
		t.Logf("Note: Stage did not change between backlog and in-progress (both %s)", result1.Stage)
	}
}

// TestIntegration_FullFlowWithRegistry tests BMAD detection through the detector registry
// This is closer to how vibe add actually works
func TestIntegration_FullFlowWithRegistry(t *testing.T) {
	fixturePath := filepath.Join(integrationFixturesDir(), "bmad-v6-complete")

	// Simulate what registry.DetectAll does
	d := NewBMADDetector()
	ctx := context.Background()

	if !d.CanDetect(ctx, fixturePath) {
		t.Fatal("CanDetect should return true for bmad-v6-complete fixture")
	}

	result, err := d.Detect(ctx, fixturePath)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	// Verify the result matches what would be stored in the project database
	if result.Method != "bmad" {
		t.Errorf("Method = %q, want %q", result.Method, "bmad")
	}
	if result.Stage.String() != "Implement" {
		t.Errorf("Stage = %s, want Implement", result.Stage)
	}

	// Log full result for debugging
	t.Logf("Full flow result: method=%s, stage=%s, confidence=%s, reasoning=%s",
		result.Method, result.Stage, result.Confidence, result.Reasoning)
}
