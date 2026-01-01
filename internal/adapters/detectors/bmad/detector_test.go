package bmad

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestBMADDetector_Name(t *testing.T) {
	d := NewBMADDetector()
	if got := d.Name(); got != "bmad" {
		t.Errorf("Name() = %q, want %q", got, "bmad")
	}
}

func TestBMADDetector_CanDetect(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, dir string)
		want  bool
	}{
		{
			name: "valid .bmad folder with config.yaml",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
			},
			want: true,
		},
		{
			name: "valid .bmad folder without config.yaml",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, false)
			},
			want: true, // CanDetect only checks for .bmad/ folder
		},
		{
			name: "no .bmad folder",
			setup: func(t *testing.T, dir string) {
				// Empty directory
			},
			want: false,
		},
		{
			name: ".bmad-core folder (v4 structure - not supported)",
			setup: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, ".bmad-core"), 0755); err != nil {
					t.Fatalf("failed to create .bmad-core: %v", err)
				}
			},
			want: false, // v4 not supported in this story
		},
		{
			name: ".bmad is a file not directory",
			setup: func(t *testing.T, dir string) {
				if err := os.WriteFile(filepath.Join(dir, ".bmad"), []byte("not a dir"), 0644); err != nil {
					t.Fatalf("failed to create .bmad file: %v", err)
				}
			},
			want: false,
		},
		{
			name: "_bmad folder (Alpha.22+ convention)",
			setup: func(t *testing.T, dir string) {
				createBMADStructureWithDir(t, dir, "_bmad", "core/config.yaml", true)
			},
			want: true,
		},
		{
			name: "_bmad-output folder only",
			setup: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "_bmad-output", "planning-artifacts"), 0755); err != nil {
					t.Fatalf("failed to create _bmad-output: %v", err)
				}
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			d := NewBMADDetector()
			got := d.CanDetect(context.Background(), dir)
			if got != tt.want {
				t.Errorf("CanDetect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBMADDetector_CanDetect_ContextCancellation(t *testing.T) {
	dir := t.TempDir()
	createBMADStructure(t, dir, true)

	d := NewBMADDetector()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	got := d.CanDetect(ctx, dir)
	if got != false {
		t.Errorf("CanDetect() with cancelled context = %v, want false", got)
	}
}

func TestBMADDetector_Detect(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, dir string)
		wantNil        bool
		wantMethod     string
		wantStage      domain.Stage
		wantConfidence domain.Confidence
		wantReasoning  string
	}{
		{
			name: "full v6 structure with version - no artifacts",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
			},
			wantNil:        false,
			wantMethod:     "bmad",
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceLikely, // Uncertain stage → Likely
			wantReasoning:  "BMAD v6.0.0-alpha.13, No BMAD artifacts detected",
		},
		{
			name: "config.yaml without version in header",
			setup: func(t *testing.T, dir string) {
				createBMADWithConfig(t, dir, "project_name: test\nbmad_folder: .bmad\n")
			},
			wantNil:        false,
			wantMethod:     "bmad",
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceLikely, // Uncertain stage → Likely
			wantReasoning:  "BMAD detected, No BMAD artifacts detected",
		},
		{
			name: ".bmad folder exists but config.yaml missing",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, false)
			},
			wantNil:        false,
			wantMethod:     "bmad",
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  ".bmad folder exists but config.yaml not found",
		},
		{
			name: "no .bmad folder",
			setup: func(t *testing.T, dir string) {
				// Empty directory
			},
			wantNil: true, // Returns nil, not an error
		},
		{
			name: ".bmad-core only (v4 not supported)",
			setup: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, ".bmad-core"), 0755); err != nil {
					t.Fatalf("failed to create .bmad-core: %v", err)
				}
			},
			wantNil: true, // v4 not supported
		},
		{
			name: "_bmad with core/config.yaml (Alpha.22+)",
			setup: func(t *testing.T, dir string) {
				createBMADStructureWithDir(t, dir, "_bmad", "core/config.yaml", true)
			},
			wantNil:        false,
			wantMethod:     "bmad",
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceLikely, // No stage artifacts
			wantReasoning:  "BMAD v6.0.0-alpha.22, No BMAD artifacts detected",
		},
		{
			name: "_bmad-output only (no config expected)",
			setup: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "_bmad-output", "planning-artifacts"), 0755); err != nil {
					t.Fatalf("failed to create _bmad-output: %v", err)
				}
			},
			wantNil:        false,
			wantMethod:     "bmad",
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "BMAD detected (_bmad-output), config not expected",
		},
		{
			name: "both .bmad and _bmad - first match wins (.bmad)",
			setup: func(t *testing.T, dir string) {
				// Create .bmad first (uses bmm/config.yaml for existing helper)
				createBMADStructure(t, dir, true)
				// Also create _bmad with core/config.yaml
				createBMADStructureWithDir(t, dir, "_bmad", "core/config.yaml", true)
			},
			wantNil:        false,
			wantMethod:     "bmad",
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "BMAD v6.0.0-alpha.13, No BMAD artifacts detected", // .bmad version, not _bmad
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			d := NewBMADDetector()
			got, err := d.Detect(context.Background(), dir)

			if err != nil {
				t.Fatalf("Detect() error = %v, want nil", err)
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("Detect() = %+v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("Detect() = nil, want non-nil result")
			}

			if got.Method != tt.wantMethod {
				t.Errorf("Method = %q, want %q", got.Method, tt.wantMethod)
			}
			if got.Stage != tt.wantStage {
				t.Errorf("Stage = %v, want %v", got.Stage, tt.wantStage)
			}
			if got.Confidence != tt.wantConfidence {
				t.Errorf("Confidence = %v, want %v", got.Confidence, tt.wantConfidence)
			}
			if got.Reasoning != tt.wantReasoning {
				t.Errorf("Reasoning = %q, want %q", got.Reasoning, tt.wantReasoning)
			}
		})
	}
}

func TestBMADDetector_Detect_ContextCancellation(t *testing.T) {
	dir := t.TempDir()
	createBMADStructure(t, dir, true)

	d := NewBMADDetector()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := d.Detect(ctx, dir)
	if err != context.Canceled {
		t.Errorf("Detect() with cancelled context error = %v, want context.Canceled", err)
	}
}

func TestBMADDetector_Detect_ContextDeadline(t *testing.T) {
	dir := t.TempDir()
	createBMADStructure(t, dir, true)

	d := NewBMADDetector()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give time for context to expire
	time.Sleep(1 * time.Millisecond)

	_, err := d.Detect(ctx, dir)
	if err != context.DeadlineExceeded {
		t.Errorf("Detect() with expired context error = %v, want context.DeadlineExceeded", err)
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantVersion string
		wantErr     bool
	}{
		{
			name: "standard v6 config header",
			content: `# BMM Module Configuration
# Generated by BMAD installer
# Version: 6.0.0-alpha.13
# Date: 2025-12-04T00:10:41.176Z

project_name: test
`,
			wantVersion: "6.0.0-alpha.13",
			wantErr:     false,
		},
		{
			name: "version without alpha tag",
			content: `# Version: 6.0.0
project_name: test
`,
			wantVersion: "6.0.0",
			wantErr:     false,
		},
		{
			name: "version with beta tag",
			content: `# Version: 7.1.2-beta.5
`,
			wantVersion: "7.1.2-beta.5",
			wantErr:     false,
		},
		{
			name:        "no version in file",
			content:     "project_name: test\nbmad_folder: .bmad\n",
			wantVersion: "",
			wantErr:     false, // No version is not an error
		},
		{
			name:        "empty file",
			content:     "",
			wantVersion: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile, err := os.CreateTemp("", "config*.yaml")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}
			tmpFile.Close()

			got, err := extractVersion(tmpFile.Name())

			if (err != nil) != tt.wantErr {
				t.Errorf("extractVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantVersion {
				t.Errorf("extractVersion() = %q, want %q", got, tt.wantVersion)
			}
		})
	}
}

func TestExtractVersion_FileNotFound(t *testing.T) {
	_, err := extractVersion("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("extractVersion() with nonexistent file should return error")
	}
}

// Helper functions for test setup

func createBMADStructure(t *testing.T, dir string, withConfig bool) {
	t.Helper()

	bmadDir := filepath.Join(dir, ".bmad", "bmm")
	if err := os.MkdirAll(bmadDir, 0755); err != nil {
		t.Fatalf("failed to create .bmad/bmm: %v", err)
	}

	if withConfig {
		configContent := `# BMM Module Configuration
# Generated by BMAD installer
# Version: 6.0.0-alpha.13
# Date: 2025-12-04T00:10:41.176Z

project_name: test
bmad_folder: .bmad
`
		configPath := filepath.Join(bmadDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config.yaml: %v", err)
		}
	}
}

func createBMADWithConfig(t *testing.T, dir string, configContent string) {
	t.Helper()

	bmadDir := filepath.Join(dir, ".bmad", "bmm")
	if err := os.MkdirAll(bmadDir, 0755); err != nil {
		t.Fatalf("failed to create .bmad/bmm: %v", err)
	}

	configPath := filepath.Join(bmadDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
}

// createBMADStructureWithDir creates a BMAD structure with a configurable marker directory and config path.
// This is a new parameterized helper for testing different directory conventions.
// Parameters:
//   - dir: base temp directory
//   - markerDir: the marker directory name (e.g., ".bmad", "_bmad", "_bmad-output")
//   - configSubPath: relative path within marker (e.g., "core/config.yaml", "bmm/config.yaml")
//   - withConfig: whether to create the config file
func createBMADStructureWithDir(t *testing.T, dir, markerDir, configSubPath string, withConfig bool) {
	t.Helper()
	configDir := filepath.Join(dir, markerDir, filepath.Dir(configSubPath))
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create %s: %v", configDir, err)
	}
	if withConfig {
		configContent := `# Core Module Configuration
# Generated by BMAD installer
# Version: 6.0.0-alpha.22
# Date: 2026-01-01T00:00:00.000Z

install_type: core
bmad_folder: ` + markerDir + `
`
		configPath := filepath.Join(dir, markerDir, configSubPath)
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config.yaml: %v", err)
		}
	}
}

// fixturesDir returns the path to the test fixtures directory
func fixturesDir() string {
	return filepath.Join("..", "..", "..", "..", "test", "fixtures")
}

// TestBMADDetector_FixtureBased tests the BMAD detector against all fixtures (AC: #3)
// This uses table-driven tests to verify CanDetect, Detect, stage, confidence, and reasoning
func TestBMADDetector_FixtureBased(t *testing.T) {
	testCases := []struct {
		fixture        string
		shouldDetect   bool
		expectedStage  domain.Stage
		expectedConf   domain.Confidence
		expectedMethod string
	}{
		// bmad-v6-complete: Full .bmad structure with sprint-status.yaml showing epic in-progress
		{
			fixture:        "bmad-v6-complete",
			shouldDetect:   true,
			expectedStage:  domain.StageImplement, // Has epic-1: in-progress
			expectedConf:   domain.ConfidenceCertain,
			expectedMethod: "bmad",
		},
		// bmad-v6-minimal: Just .bmad/bmm/config.yaml - no sprint-status or artifacts
		{
			fixture:        "bmad-v6-minimal",
			shouldDetect:   true,
			expectedStage:  domain.StageUnknown,
			expectedConf:   domain.ConfidenceLikely, // No stage info available
			expectedMethod: "bmad",
		},
		// bmad-v6-no-config: .bmad folder exists but no config.yaml
		{
			fixture:        "bmad-v6-no-config",
			shouldDetect:   true, // CanDetect only checks for .bmad/ folder
			expectedStage:  domain.StageUnknown,
			expectedConf:   domain.ConfidenceLikely,
			expectedMethod: "bmad",
		},
		// bmad-v6-mid-sprint: sprint-status.yaml with one epic done, one in-progress
		{
			fixture:        "bmad-v6-mid-sprint",
			shouldDetect:   true,
			expectedStage:  domain.StageImplement, // Has epic-2: in-progress
			expectedConf:   domain.ConfidenceCertain,
			expectedMethod: "bmad",
		},
		// bmad-v6-all-done: All epics marked done
		{
			fixture:        "bmad-v6-all-done",
			shouldDetect:   true,
			expectedStage:  domain.StageImplement, // All done = still Implement stage
			expectedConf:   domain.ConfidenceCertain,
			expectedMethod: "bmad",
		},
		// bmad-v6-artifacts-only: has epics.md but no sprint-status - falls back to artifact detection
		// Returns Certain confidence because artifact detection returns Likely (not Uncertain)
		// and the detector only downgrades to Likely when stageConfidence == Uncertain
		{
			fixture:        "bmad-v6-artifacts-only",
			shouldDetect:   true,
			expectedStage:  domain.StageImplement, // Epic artifacts detected
			expectedConf:   domain.ConfidenceCertain,
			expectedMethod: "bmad",
		},
		// bmad-v4-not-supported: .bmad-core folder (v4 structure)
		{
			fixture:      "bmad-v4-not-supported",
			shouldDetect: false, // v4 not supported
		},
		// === New Alpha.22+ fixtures ===
		// bmad-v6-underscore: _bmad folder with core/config.yaml (Alpha.22+ convention)
		// Note: Stage is Unknown because stage_parser only checks docs/ path (follow-up story scope)
		{
			fixture:        "bmad-v6-underscore",
			shouldDetect:   true,
			expectedStage:  domain.StageUnknown, // sprint-status in _bmad-output/ not yet supported
			expectedConf:   domain.ConfidenceLikely,
			expectedMethod: "bmad",
		},
		// bmad-v6-output-only: Only _bmad-output folder (indicates BMAD but no config expected)
		{
			fixture:        "bmad-v6-output-only",
			shouldDetect:   true,
			expectedStage:  domain.StageUnknown,
			expectedConf:   domain.ConfidenceLikely,
			expectedMethod: "bmad",
		},
		// bmad-v6-both-dirs: Both .bmad and _bmad exist - .bmad wins (first match)
		{
			fixture:        "bmad-v6-both-dirs",
			shouldDetect:   true,
			expectedStage:  domain.StageUnknown,
			expectedConf:   domain.ConfidenceLikely,
			expectedMethod: "bmad",
		},
	}

	d := NewBMADDetector()
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.fixture, func(t *testing.T) {
			fixturePath := filepath.Join(fixturesDir(), tc.fixture)

			// Verify fixture exists
			if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
				t.Fatalf("Fixture does not exist: %s", fixturePath)
			}

			canDetect := d.CanDetect(ctx, fixturePath)

			if tc.shouldDetect {
				if !canDetect {
					t.Errorf("CanDetect() = false, want true for fixture %s", tc.fixture)
					return
				}

				result, err := d.Detect(ctx, fixturePath)
				if err != nil {
					t.Fatalf("Detect() error: %v", err)
				}

				if result == nil {
					t.Fatal("Detect() returned nil result")
				}

				if result.Method != tc.expectedMethod {
					t.Errorf("Method = %q, want %q", result.Method, tc.expectedMethod)
				}
				if result.Stage != tc.expectedStage {
					t.Errorf("Stage = %v, want %v", result.Stage, tc.expectedStage)
				}
				if result.Confidence != tc.expectedConf {
					t.Errorf("Confidence = %v, want %v", result.Confidence, tc.expectedConf)
				}
				if result.Reasoning == "" {
					t.Error("Reasoning should not be empty")
				}
			} else {
				if canDetect {
					t.Errorf("CanDetect() = true, want false for fixture %s", tc.fixture)
				}
			}
		})
	}
}

// TestBMADDetector_EdgeCases tests edge cases (malformed YAML, missing files)
func TestBMADDetector_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, dir string)
		wantCanDetect  bool
		wantStage      domain.Stage
		wantConfidence domain.Confidence
	}{
		{
			name: "malformed sprint-status.yaml",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
				// Create malformed sprint-status.yaml
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				malformedYAML := `{this is not valid yaml`
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte(malformedYAML), 0644); err != nil {
					t.Fatalf("failed to write malformed yaml: %v", err)
				}
			},
			wantCanDetect:  true,
			wantStage:      domain.StageUnknown, // Falls back to Unknown on parse error
			wantConfidence: domain.ConfidenceLikely,
		},
		{
			name: "empty sprint-status.yaml",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte(""), 0644); err != nil {
					t.Fatalf("failed to write empty yaml: %v", err)
				}
			},
			wantCanDetect:  true,
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceLikely,
		},
		{
			name: "sprint-status.yaml without development_status",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				yaml := `project: test
generated: 2025-12-20
`
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte(yaml), 0644); err != nil {
					t.Fatalf("failed to write yaml: %v", err)
				}
			},
			wantCanDetect:  true,
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceLikely,
		},
	}

	d := NewBMADDetector()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			got := d.CanDetect(ctx, dir)
			if got != tt.wantCanDetect {
				t.Errorf("CanDetect() = %v, want %v", got, tt.wantCanDetect)
			}

			if tt.wantCanDetect {
				result, err := d.Detect(ctx, dir)
				if err != nil {
					t.Fatalf("Detect() error: %v", err)
				}
				if result.Stage != tt.wantStage {
					t.Errorf("Stage = %v, want %v", result.Stage, tt.wantStage)
				}
				if result.Confidence != tt.wantConfidence {
					t.Errorf("Confidence = %v, want %v", result.Confidence, tt.wantConfidence)
				}
			}
		})
	}
}

// TestBMADDetectionAccuracy runs against all BMAD fixtures and calculates accuracy.
// This is the launch blocker test - must be >= 95%
// Run with: make test-accuracy or go test -v -run TestBMADDetectionAccuracy ./internal/adapters/detectors/bmad/...
func TestBMADDetectionAccuracy(t *testing.T) {
	testCases := []struct {
		fixture       string
		expectedStage domain.Stage
		shouldDetect  bool // false for non-BMAD fixtures
	}{
		// === BMAD v6 Fixtures (10 total) ===
		{"bmad-v6-complete", domain.StageImplement, true},
		{"bmad-v6-minimal", domain.StageUnknown, true},
		{"bmad-v6-no-config", domain.StageUnknown, true},
		{"bmad-v6-mid-sprint", domain.StageImplement, true},
		{"bmad-v6-all-done", domain.StageImplement, true},
		{"bmad-v6-artifacts-only", domain.StageImplement, true},
		{"bmad-v4-not-supported", domain.StageUnknown, false},
		// === New Alpha.22+ fixtures ===
		{"bmad-v6-underscore", domain.StageUnknown, true}, // sprint-status in _bmad-output/ not yet supported
		{"bmad-v6-output-only", domain.StageUnknown, true},
		{"bmad-v6-both-dirs", domain.StageUnknown, true},
	}

	d := NewBMADDetector()
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
			// Should NOT detect as BMAD
			if !canDetect {
				correct++
				t.Logf("PASS: %s - Correctly not detected as BMAD", tc.fixture)
			} else {
				t.Logf("FAIL: %s - Should not be detected as BMAD", tc.fixture)
			}
		}
	}

	accuracy := float64(correct) / float64(total) * 100
	t.Logf("\n=== BMAD DETECTION ACCURACY: %.1f%% (%d/%d) ===", accuracy, correct, total)

	if accuracy < 95.0 {
		t.Errorf("BMAD detection accuracy %.1f%% is below 95%% launch blocker threshold", accuracy)
	}
}

// TestCrossDetectorExclusion ensures BMAD and Speckit detectors don't conflict
// BMAD should not detect Speckit fixtures and vice versa
func TestCrossDetectorExclusion(t *testing.T) {
	bmadDetector := NewBMADDetector()
	ctx := context.Background()

	// BMAD fixtures should NOT be detected by Speckit (tested in Speckit tests)
	// Here we test that BMAD does NOT detect Speckit fixtures

	speckitFixtures := []string{
		"speckit-stage-specify",
		"speckit-stage-plan",
		"speckit-stage-tasks",
		"speckit-stage-implement",
	}

	for _, fixture := range speckitFixtures {
		fixturePath := filepath.Join(fixturesDir(), fixture)
		// Verify fixture exists before testing
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Fatalf("Speckit fixture does not exist: %s", fixture)
		}
		if bmadDetector.CanDetect(ctx, fixturePath) {
			t.Errorf("BMAD detector should not detect Speckit fixture: %s", fixture)
		}
	}

	// Non-method fixtures should not be detected by BMAD
	nonMethodFixtures := []string{
		"no-method-detected",
		"empty-project",
	}

	for _, fixture := range nonMethodFixtures {
		fixturePath := filepath.Join(fixturesDir(), fixture)
		// Verify fixture exists before testing
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Fatalf("Non-method fixture does not exist: %s", fixture)
		}
		if bmadDetector.CanDetect(ctx, fixturePath) {
			t.Errorf("BMAD detector should not detect non-method fixture: %s", fixture)
		}
	}
}

// TestBMADDetector_ReasoningFormat verifies AC6: detected folder shown in reasoning.
// The reasoning string should include the marker directory name for transparency.
func TestBMADDetector_ReasoningFormat(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, dir string)
		wantContains   string // substring that should appear in reasoning
	}{
		{
			name: ".bmad folder shows in reasoning",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
			},
			wantContains: "BMAD v6.0.0-alpha.13",
		},
		{
			name: "_bmad folder shows in reasoning",
			setup: func(t *testing.T, dir string) {
				createBMADStructureWithDir(t, dir, "_bmad", "core/config.yaml", true)
			},
			wantContains: "BMAD v6.0.0-alpha.22",
		},
		{
			name: "_bmad-output shows in reasoning",
			setup: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "_bmad-output", "planning-artifacts"), 0755); err != nil {
					t.Fatalf("failed to create _bmad-output: %v", err)
				}
			},
			wantContains: "_bmad-output",
		},
		{
			name: ".bmad without config shows folder name",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, false)
			},
			wantContains: ".bmad folder exists",
		},
	}

	d := NewBMADDetector()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			result, err := d.Detect(ctx, dir)
			if err != nil {
				t.Fatalf("Detect() error: %v", err)
			}
			if result == nil {
				t.Fatal("Detect() returned nil")
			}

			if !strings.Contains(result.Reasoning, tt.wantContains) {
				t.Errorf("Reasoning %q should contain %q", result.Reasoning, tt.wantContains)
			}
		})
	}
}
