package bmad

import (
	"context"
	"os"
	"path/filepath"
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
