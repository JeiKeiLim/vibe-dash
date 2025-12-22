package bmad

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// =============================================================================
// normalizeStatus Tests
// =============================================================================

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// G15: Space/underscore variations
		{"in progress", "in-progress"},
		{"in_progress", "in-progress"},
		{"IN-PROGRESS", "in-progress"},
		// G17: Synonyms
		{"wip", "in-progress"},
		{"complete", "done"},
		{"completed", "done"},
		{"finished", "done"},
		{"reviewing", "review"},
		{"in-review", "review"},
		{"code-review", "review"},
		// Pass-through
		{"done", "done"},
		{"backlog", "backlog"},
		{"ready-for-dev", "ready-for-dev"},
		{"drafted", "drafted"},
		// Whitespace handling
		{"  in-progress  ", "in-progress"},
		{"In Progress", "in-progress"},
		// Multiple separator handling (Edge 2)
		{"ready__for__dev", "ready-for-dev"},
		{"in___progress", "in-progress"},
		{"ready_for__dev", "ready-for-dev"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeStatus(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeStatus(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// parseSprintStatus Tests
// =============================================================================

func TestParseSprintStatus(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		wantEpics   int
		wantStories int
	}{
		{
			name: "valid sprint-status.yaml",
			content: `development_status:
  epic-1: done
  1-1-project-scaffolding: done
  epic-2: in-progress
  2-1-feature: in-progress
`,
			wantErr:     false,
			wantEpics:   2,
			wantStories: 2,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false,
		},
		{
			name:    "invalid yaml syntax",
			content: "development_status:\n  - epic-1: done\n  invalid yaml here",
			wantErr: true,
		},
		{
			name: "no development_status key",
			content: `project: test
version: 1.0
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			dir := t.TempDir()
			path := filepath.Join(dir, "sprint-status.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			status, err := parseSprintStatus(context.Background(), path)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseSprintStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if status == nil {
				t.Fatalf("parseSprintStatus() returned nil status")
			}
		})
	}
}

func TestParseSprintStatus_FileNotFound(t *testing.T) {
	_, err := parseSprintStatus(context.Background(), "/nonexistent/path/sprint-status.yaml")
	if err == nil {
		t.Error("parseSprintStatus() with nonexistent file should return error")
	}
}

func TestParseSprintStatus_ContextCancelled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sprint-status.yaml")
	if err := os.WriteFile(path, []byte("development_status:\n  epic-1: done\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := parseSprintStatus(ctx, path)
	if err != context.Canceled {
		t.Errorf("parseSprintStatus() with cancelled context error = %v, want context.Canceled", err)
	}
}

func TestParseSprintStatus_ContextTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sprint-status.yaml")
	if err := os.WriteFile(path, []byte("development_status:\n  epic-1: done\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give time for context to expire
	time.Sleep(1 * time.Millisecond)

	_, err := parseSprintStatus(ctx, path)
	if err != context.DeadlineExceeded {
		t.Errorf("parseSprintStatus() with expired context error = %v, want context.DeadlineExceeded", err)
	}
}

// =============================================================================
// determineStageFromStatus Tests
// =============================================================================

func TestDetermineStageFromStatus(t *testing.T) {
	tests := []struct {
		name           string
		status         *SprintStatus
		wantStage      domain.Stage
		wantConfidence domain.Confidence
		wantReasoning  string
	}{
		{
			name:           "nil status",
			status:         nil,
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceUncertain,
			wantReasoning:  "sprint-status.yaml is empty",
		},
		{
			name: "empty development_status",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{},
			},
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceUncertain,
			wantReasoning:  "sprint-status.yaml is empty",
		},
		{
			name: "all epics backlog",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1": "backlog",
					"epic-2": "backlog",
				},
			},
			wantStage:      domain.StageSpecify,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "No epics in progress - planning phase",
		},
		{
			name: "epic in-progress, stories backlog",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":                  "in-progress",
					"1-1-project-scaffolding": "backlog",
					"1-2-domain-entities":     "backlog",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.1 in backlog, needs drafting",
		},
		{
			name: "story in-progress",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":                  "in-progress",
					"1-1-project-scaffolding": "done",
					"1-2-domain-entities":     "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.2 being implemented",
		},
		{
			name: "story in review",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":                  "in-progress",
					"1-1-project-scaffolding": "done",
					"1-2-domain-entities":     "review",
				},
			},
			wantStage:      domain.StageTasks,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.2 in code review",
		},
		{
			name: "all epics done",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":                  "done",
					"1-1-project-scaffolding": "done",
					"epic-2":                  "done",
					"2-1-feature":             "done",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "All epics complete - project done",
		},
		{
			name: "mixed: some done, one in-progress with story",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":                  "done",
					"1-1-project-scaffolding": "done",
					"epic-2":                  "in-progress",
					"2-1-feature":             "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 2.1 being implemented",
		},
		{
			name: "epic 4-5 format (sub-epic)",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-4-5":                      "in-progress",
					"4-5-1-bmad-v6-detector":        "done",
					"4-5-2-bmad-v6-stage-detection": "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 4.5.2 being implemented",
		},
		{
			name: "retrospective entries should be ignored",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":               "done",
					"1-1-feature":          "done",
					"epic-1-retrospective": "completed",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "All epics complete - project done",
		},
		{
			name: "contexted status treated as in-progress",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "contexted",
					"1-1-feature": "backlog",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.1 in backlog, needs drafting",
		},
		// G1: All stories done detection
		{
			name: "G1: epic in-progress, all stories done",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-1-feature": "done",
					"1-2-feature": "done",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Epic 1 stories complete, update epic status",
		},
		{
			name: "G1: multiple epics, one with all stories done",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-1-feature": "done",
					"1-2-feature": "done",
					"epic-2":      "backlog",
					"2-1-feature": "backlog",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Epic 1 stories complete, update epic status",
		},
		// G7: Epic done with active stories
		{
			name: "G7: epic done, story in-progress",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "done",
					"1-1-feature": "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "Epic done but Story 1.1 in-progress",
		},
		{
			name: "G7: epic done, story in review",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "done",
					"1-1-feature": "review",
				},
			},
			wantStage:      domain.StageTasks,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "Epic done but Story 1.1 in review",
		},
		// G8: Epic backlog with active stories
		{
			name: "G8: epic backlog, story in-progress",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "backlog",
					"1-1-feature": "in-progress",
				},
			},
			wantStage:      domain.StageSpecify,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "Epic backlog but Story 1.1 active",
		},
		{
			name: "G8: epic backlog, story done",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "backlog",
					"1-1-feature": "done",
				},
			},
			wantStage:      domain.StageSpecify,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "Epic backlog but Story 1.1 active",
		},
		// G2/G3: Drafted and Ready-for-Dev display appropriately
		{
			name: "G2: story with ready-for-dev status",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-1-feature": "ready-for-dev",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.1 ready for development",
		},
		{
			name: "G3: story with drafted status",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-1-feature": "drafted",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.1 drafted, needs review",
		},
		{
			name: "G2/G3: drafted-only scenario",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-1-feature": "drafted",
					"1-2-feature": "backlog",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.1 drafted, needs review",
		},
		{
			name: "G2/G3: ready-for-dev takes precedence over drafted",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-1-feature": "drafted",
					"1-2-feature": "ready-for-dev",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.2 ready for development",
		},
		// G19: Deterministic story order (first by sorted key when same priority)
		{
			name: "G19: multiple stories same priority, first by sorted key wins",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-3-feature": "in-progress",
					"1-1-feature": "in-progress",
					"1-2-feature": "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.1 being implemented",
		},
		// G14: Orphan stories warning
		{
			name: "G14: orphan story without matching epic",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-1-feature": "in-progress",
					"2-1-orphan":  "backlog", // No epic-2 defined
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 1.1 being implemented [Warning: orphan story 2.1]",
		},
		// G22: Empty status warning
		{
			name: "G22: empty status value",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-1-feature": "",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Epic 1 started, preparing stories [Warning: empty status for 1-1-feature]",
		},
		// Edge 15: Unknown status value shows story and status
		{
			name: "Edge15: unknown status shows story name and status value",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-1-feature": "some-random-status",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "Story 1.1 has unknown status 'some-random-status' [Warning: unknown status 'some-random-status' for 1-1-feature]",
		},
		// Edge 15 variant: multiple unknown statuses, first by sorted key wins
		{
			name: "Edge15: multiple unknown statuses, first by sorted key",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "in-progress",
					"1-2-feature": "weird-status",
					"1-1-feature": "another-weird",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "Story 1.1 has unknown status 'another-weird' [Warning: unknown status 'another-weird' for 1-1-feature]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, confidence, reasoning := determineStageFromStatus(tt.status)

			if stage != tt.wantStage {
				t.Errorf("determineStageFromStatus() stage = %v, want %v", stage, tt.wantStage)
			}
			if confidence != tt.wantConfidence {
				t.Errorf("determineStageFromStatus() confidence = %v, want %v", confidence, tt.wantConfidence)
			}
			if reasoning != tt.wantReasoning {
				t.Errorf("determineStageFromStatus() reasoning = %q, want %q", reasoning, tt.wantReasoning)
			}
		})
	}
}

// =============================================================================
// detectStageFromArtifacts Tests
// =============================================================================

func TestDetectStageFromArtifacts(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, dir string)
		wantStage      domain.Stage
		wantConfidence domain.Confidence
		wantReasoning  string
	}{
		{
			name: "has epic file",
			setup: func(t *testing.T, dir string) {
				docsDir := filepath.Join(dir, "docs")
				if err := os.MkdirAll(docsDir, 0755); err != nil {
					t.Fatalf("failed to create docs: %v", err)
				}
				if err := os.WriteFile(filepath.Join(docsDir, "epic-1.md"), []byte("# Epic 1"), 0644); err != nil {
					t.Fatalf("failed to write epic file: %v", err)
				}
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "Epics defined but no sprint status",
		},
		{
			name: "has architecture file",
			setup: func(t *testing.T, dir string) {
				docsDir := filepath.Join(dir, "docs")
				if err := os.MkdirAll(docsDir, 0755); err != nil {
					t.Fatalf("failed to create docs: %v", err)
				}
				if err := os.WriteFile(filepath.Join(docsDir, "architecture.md"), []byte("# Architecture"), 0644); err != nil {
					t.Fatalf("failed to write architecture file: %v", err)
				}
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "Architecture designed, no epics yet",
		},
		{
			name: "has PRD file",
			setup: func(t *testing.T, dir string) {
				docsDir := filepath.Join(dir, "docs")
				if err := os.MkdirAll(docsDir, 0755); err != nil {
					t.Fatalf("failed to create docs: %v", err)
				}
				if err := os.WriteFile(filepath.Join(docsDir, "prd.md"), []byte("# PRD"), 0644); err != nil {
					t.Fatalf("failed to write PRD file: %v", err)
				}
			},
			wantStage:      domain.StageSpecify,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "PRD created, architecture pending",
		},
		{
			name: "no artifacts",
			setup: func(t *testing.T, dir string) {
				// Empty directory or no docs folder
			},
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceUncertain,
			wantReasoning:  "No BMAD artifacts detected",
		},
		{
			name: "epic takes priority over architecture",
			setup: func(t *testing.T, dir string) {
				docsDir := filepath.Join(dir, "docs")
				if err := os.MkdirAll(docsDir, 0755); err != nil {
					t.Fatalf("failed to create docs: %v", err)
				}
				if err := os.WriteFile(filepath.Join(docsDir, "epic-1.md"), []byte("# Epic"), 0644); err != nil {
					t.Fatalf("failed to write epic file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(docsDir, "architecture.md"), []byte("# Arch"), 0644); err != nil {
					t.Fatalf("failed to write architecture file: %v", err)
				}
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceLikely,
			wantReasoning:  "Epics defined but no sprint status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			stage, confidence, reasoning, err := detectStageFromArtifacts(context.Background(), dir)

			if err != nil {
				t.Fatalf("detectStageFromArtifacts() error = %v", err)
			}

			if stage != tt.wantStage {
				t.Errorf("detectStageFromArtifacts() stage = %v, want %v", stage, tt.wantStage)
			}
			if confidence != tt.wantConfidence {
				t.Errorf("detectStageFromArtifacts() confidence = %v, want %v", confidence, tt.wantConfidence)
			}
			if reasoning != tt.wantReasoning {
				t.Errorf("detectStageFromArtifacts() reasoning = %q, want %q", reasoning, tt.wantReasoning)
			}
		})
	}
}

func TestDetectStageFromArtifacts_ContextCancellation(t *testing.T) {
	dir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, _, err := detectStageFromArtifacts(ctx, dir)
	if err != context.Canceled {
		t.Errorf("detectStageFromArtifacts() with cancelled context error = %v, want context.Canceled", err)
	}
}

// =============================================================================
// Helper Functions Tests
// =============================================================================

func TestExtractStoryPrefix(t *testing.T) {
	tests := []struct {
		storyKey string
		want     string
	}{
		{"1-1-project-scaffolding", "1"},
		{"1-2-domain-entities", "1"},
		{"4-5-1-bmad-v6-detector", "4-5"},
		{"4-5-2-bmad-v6-stage-detection", "4-5"},
		{"10-1-feature", "10"},
		{"1-10-feature", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.storyKey, func(t *testing.T) {
			got := extractStoryPrefix(tt.storyKey)
			if got != tt.want {
				t.Errorf("extractStoryPrefix(%q) = %q, want %q", tt.storyKey, got, tt.want)
			}
		})
	}
}

func TestFormatStoryKey(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"1-1-project-scaffolding", "1.1"},
		{"4-5-2-bmad-v6-stage-detection", "4.5.2"},
		{"10-1-feature", "10.1"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := formatStoryKey(tt.key)
			if got != tt.want {
				t.Errorf("formatStoryKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestFormatEpicKey(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"epic-1", "Epic 1"},
		{"epic-4-5", "Epic 4.5"},
		{"epic-10", "Epic 10"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := formatEpicKey(tt.key)
			if got != tt.want {
				t.Errorf("formatEpicKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"123", true},
		{"0", true},
		{"abc", false},
		{"1a2", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := isNumeric(tt.s)
			if got != tt.want {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// =============================================================================
// findSprintStatusPath Tests
// =============================================================================

func TestFindSprintStatusPath(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		wantFound bool
		wantPath  string // relative to dir
	}{
		{
			name: "primary location: docs/sprint-artifacts/sprint-status.yaml",
			setup: func(t *testing.T, dir string) {
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte("test"), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			wantFound: true,
			wantPath:  "docs/sprint-artifacts/sprint-status.yaml",
		},
		{
			name: "alternative location: docs/sprint-status.yaml",
			setup: func(t *testing.T, dir string) {
				docsDir := filepath.Join(dir, "docs")
				if err := os.MkdirAll(docsDir, 0755); err != nil {
					t.Fatalf("failed to create docs: %v", err)
				}
				if err := os.WriteFile(filepath.Join(docsDir, "sprint-status.yaml"), []byte("test"), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			wantFound: true,
			wantPath:  "docs/sprint-status.yaml",
		},
		{
			name: "not found",
			setup: func(t *testing.T, dir string) {
				// Empty directory
			},
			wantFound: false,
		},
		{
			name: "primary takes precedence over alternative",
			setup: func(t *testing.T, dir string) {
				// Create both locations
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte("primary"), 0644); err != nil {
					t.Fatalf("failed to write primary sprint-status.yaml: %v", err)
				}

				docsDir := filepath.Join(dir, "docs")
				if err := os.WriteFile(filepath.Join(docsDir, "sprint-status.yaml"), []byte("alt"), 0644); err != nil {
					t.Fatalf("failed to write alt sprint-status.yaml: %v", err)
				}
			},
			wantFound: true,
			wantPath:  "docs/sprint-artifacts/sprint-status.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			got := findSprintStatusPath(dir)

			if tt.wantFound {
				wantFullPath := filepath.Join(dir, tt.wantPath)
				if got != wantFullPath {
					t.Errorf("findSprintStatusPath() = %q, want %q", got, wantFullPath)
				}
			} else {
				if got != "" {
					t.Errorf("findSprintStatusPath() = %q, want empty string", got)
				}
			}
		})
	}
}

// =============================================================================
// Integration Tests: detectStage
// =============================================================================

func TestDetectStage_Integration(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, dir string)
		wantStage      domain.Stage
		wantConfidence domain.Confidence
		checkReasoning func(reasoning string) bool
	}{
		{
			name: "with valid sprint-status.yaml",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				content := `development_status:
  epic-1: in-progress
  1-1-feature: in-progress
`
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			checkReasoning: func(r string) bool { return r == "Story 1.1 being implemented" },
		},
		{
			name: "with malformed sprint-status.yaml falls back to artifacts",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				// Malformed YAML
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte("invalid: yaml: here"), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceUncertain,
			checkReasoning: func(r string) bool { return r == "sprint-status.yaml parse error" },
		},
		{
			name: "no sprint-status.yaml, has epics",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
				docsDir := filepath.Join(dir, "docs")
				if err := os.MkdirAll(docsDir, 0755); err != nil {
					t.Fatalf("failed to create docs: %v", err)
				}
				if err := os.WriteFile(filepath.Join(docsDir, "epic-1.md"), []byte("# Epic"), 0644); err != nil {
					t.Fatalf("failed to write epic: %v", err)
				}
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceLikely,
			checkReasoning: func(r string) bool { return r == "Epics defined but no sprint status" },
		},
		{
			name: "no sprint-status.yaml, no artifacts",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
			},
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceUncertain,
			checkReasoning: func(r string) bool { return r == "No BMAD artifacts detected" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			d := NewBMADDetector()
			stage, confidence, reasoning := d.detectStage(context.Background(), dir)

			if stage != tt.wantStage {
				t.Errorf("detectStage() stage = %v, want %v", stage, tt.wantStage)
			}
			if confidence != tt.wantConfidence {
				t.Errorf("detectStage() confidence = %v, want %v", confidence, tt.wantConfidence)
			}
			if !tt.checkReasoning(reasoning) {
				t.Errorf("detectStage() reasoning = %q, check failed", reasoning)
			}
		})
	}
}

func TestDetectStage_ContextCancellation(t *testing.T) {
	dir := t.TempDir()
	createBMADStructure(t, dir, true)

	d := NewBMADDetector()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	stage, confidence, reasoning := d.detectStage(ctx, dir)

	if stage != domain.StageUnknown {
		t.Errorf("detectStage() with cancelled context stage = %v, want StageUnknown", stage)
	}
	if confidence != domain.ConfidenceUncertain {
		t.Errorf("detectStage() with cancelled context confidence = %v, want ConfidenceUncertain", confidence)
	}
	if reasoning != "" {
		t.Errorf("detectStage() with cancelled context reasoning = %q, want empty", reasoning)
	}
}

// =============================================================================
// Full Detect Integration Tests with Stage
// =============================================================================

func TestDetect_WithStageDetection(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, dir string)
		wantStage      domain.Stage
		wantConfidence domain.Confidence
		checkReasoning func(reasoning string) bool
	}{
		{
			name: "BMAD project with sprint-status showing in-progress story",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				content := `development_status:
  epic-1: in-progress
  1-1-feature: in-progress
`
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			checkReasoning: func(r string) bool {
				return r == "BMAD v6.0.0-alpha.13, Story 1.1 being implemented"
			},
		},
		{
			name: "BMAD project with all epics done",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				content := `development_status:
  epic-1: done
  1-1-feature: done
  epic-2: done
  2-1-feature: done
`
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			checkReasoning: func(r string) bool {
				return r == "BMAD v6.0.0-alpha.13, All epics complete - project done"
			},
		},
		{
			name: "BMAD project with story in review",
			setup: func(t *testing.T, dir string) {
				createBMADStructure(t, dir, true)
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				content := `development_status:
  epic-1: in-progress
  1-1-feature: review
`
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			wantStage:      domain.StageTasks,
			wantConfidence: domain.ConfidenceCertain,
			checkReasoning: func(r string) bool {
				return r == "BMAD v6.0.0-alpha.13, Story 1.1 in code review"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			d := NewBMADDetector()
			result, err := d.Detect(context.Background(), dir)

			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if result == nil {
				t.Fatal("Detect() = nil, want non-nil result")
			}

			if result.Stage != tt.wantStage {
				t.Errorf("Detect() stage = %v, want %v", result.Stage, tt.wantStage)
			}
			if result.Confidence != tt.wantConfidence {
				t.Errorf("Detect() confidence = %v, want %v", result.Confidence, tt.wantConfidence)
			}
			if !tt.checkReasoning(result.Reasoning) {
				t.Errorf("Detect() reasoning = %q, check failed", result.Reasoning)
			}
		})
	}
}
