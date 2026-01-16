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
// isDeferred Tests (G24)
// =============================================================================

func TestIsDeferred(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		// Deferred variations - should return true
		{"deferred", true},
		{"deferred-post-mvp", true},
		{"deferred-to-v2", true},
		{"post-mvp", true},
		// Active statuses - should return false
		{"in-progress", false},
		{"done", false},
		{"backlog", false},
		{"review", false},
		{"ready-for-dev", false},
		{"drafted", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := isDeferred(tt.status)
			if got != tt.want {
				t.Errorf("isDeferred(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// TestIsDeferredWithNormalization verifies the combined flow: normalizeStatus() → isDeferred()
// M2 fix: Ensures uppercase/space/underscore variations of deferred statuses work correctly
func TestIsDeferredWithNormalization(t *testing.T) {
	tests := []struct {
		rawStatus string
		want      bool
	}{
		// Uppercase variations
		{"DEFERRED", true},
		{"DEFERRED-POST-MVP", true},
		{"POST-MVP", true},
		// Space/underscore variations
		{"deferred post mvp", true},
		{"deferred_post_mvp", true},
		{"Deferred To V2", true},
		{"post mvp", true},
		// Active statuses should still be false after normalization
		{"IN PROGRESS", false},
		{"DONE", false},
	}

	for _, tt := range tests {
		t.Run(tt.rawStatus, func(t *testing.T) {
			normalized := normalizeStatus(tt.rawStatus)
			got := isDeferred(normalized)
			if got != tt.want {
				t.Errorf("isDeferred(normalizeStatus(%q)) = %v, want %v (normalized=%q)",
					tt.rawStatus, got, tt.want, normalized)
			}
		})
	}
}

// =============================================================================
// isActiveStatus Tests (G23)
// =============================================================================

func TestIsActiveStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		// Active statuses - should return true
		{"in-progress", true},
		{"started", true},
		// Inactive statuses - should return false
		{"done", false},
		{"completed", false}, // Note: normalizeStatus would map this to "done" first
		{"optional", false},
		{"backlog", false},
		{"review", false},
		{"ready-for-dev", false},
		{"drafted", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := isActiveStatus(tt.status)
			if got != tt.want {
				t.Errorf("isActiveStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// TestIsActiveStatusWithNormalization verifies the combined flow: normalizeStatus() → isActiveStatus()
// G23: Ensures WIP and other variations correctly detect as active
func TestIsActiveStatusWithNormalization(t *testing.T) {
	tests := []struct {
		rawStatus string
		want      bool
	}{
		// Active variations - should be true after normalization
		{"WIP", true},
		{"wip", true},
		{"In Progress", true},
		{"in_progress", true},
		{"IN-PROGRESS", true},
		{"started", true},
		{"STARTED", true},
		// Inactive variations - should be false after normalization
		{"DONE", false},
		{"completed", false},
		{"Completed", false},
		{"COMPLETED", false},
		{"optional", false},
		{"OPTIONAL", false},
		{"backlog", false},
	}

	for _, tt := range tests {
		t.Run(tt.rawStatus, func(t *testing.T) {
			normalized := normalizeStatus(tt.rawStatus)
			got := isActiveStatus(normalized)
			if got != tt.want {
				t.Errorf("isActiveStatus(normalizeStatus(%q)) = %v, want %v (normalized=%q)",
					tt.rawStatus, got, tt.want, normalized)
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
		// G24: Deferred epic handling
		{
			name: "G24: single deferred epic with active epics",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "deferred-post-mvp",
					"1-1-feature": "deferred-post-mvp",
					"epic-2":      "in-progress",
					"2-1-feature": "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 2.1 being implemented",
		},
		{
			name: "G24: all epics deferred",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1": "deferred-post-mvp",
					"epic-2": "deferred",
				},
			},
			wantStage:      domain.StageUnknown,
			wantConfidence: domain.ConfidenceUncertain,
			wantReasoning:  "All epics deferred - no active development",
		},
		{
			name: "G24: deferred-to-v2 variation",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "deferred-to-v2",
					"epic-2":      "in-progress",
					"2-1-feature": "ready-for-dev",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 2.1 ready for development",
		},
		{
			name: "G24: post-mvp variation",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1": "post-mvp",
					"epic-2": "backlog",
				},
			},
			wantStage:      domain.StageSpecify,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "No epics in progress - planning phase",
		},
		{
			name: "G24: deferred epic stories not flagged as orphans",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "deferred-post-mvp",
					"1-1-feature": "backlog", // Story for deferred epic - should be silently ignored
					"1-2-feature": "drafted", // Another story for deferred epic - should be silently ignored
					"epic-2":      "in-progress",
					"2-1-feature": "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 2.1 being implemented", // Verifies no "[Warning: orphan story 1.x]" appears
		},
		// M1 fix: Story for deferred epic with different status (edge case)
		{
			name: "G24: story for deferred epic with in-progress status is ignored",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "deferred-post-mvp",
					"1-1-feature": "in-progress", // Story marked in-progress but epic is deferred
					"epic-2":      "in-progress",
					"2-1-feature": "backlog",
				},
			},
			wantStage:      domain.StagePlan,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 2.1 in backlog, needs drafting", // 1-1-feature is ignored, no orphan warning
		},
		// G23: Retrospective stage detection
		{
			name: "G23: all epics done with retro in-progress",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-6":               "done",
					"6-1-feature":          "done",
					"epic-7":               "done",
					"7-1-feature":          "done",
					"epic-7-retrospective": "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Retrospective for Epic 7 in progress",
		},
		{
			name: "G23: all epics done with retro completed",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-6":               "done",
					"6-1-feature":          "done",
					"epic-7":               "done",
					"7-1-feature":          "done",
					"epic-7-retrospective": "completed",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "All epics complete - project done",
		},
		{
			name: "G23: all epics done with retro optional",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-6":               "done",
					"6-1-feature":          "done",
					"epic-7":               "done",
					"7-1-feature":          "done",
					"epic-7-retrospective": "optional",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "All epics complete - project done",
		},
		{
			name: "G23: epic in-progress - retro ignored",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-6":               "done",
					"epic-7":               "in-progress",
					"7-1-feature":          "in-progress",
					"epic-7-retrospective": "in-progress", // Should be ignored when not all-done
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Story 7.1 being implemented",
		},
		{
			name: "G23: multiple retros - uses first by epic number",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-6":               "done",
					"epic-7":               "done",
					"epic-6-retrospective": "in-progress",
					"epic-7-retrospective": "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Retrospective for Epic 6 in progress", // 6 < 7 by string sort
		},
		{
			name: "G23: retro status normalization - WIP",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-7":               "done",
					"7-1-feature":          "done",
					"epic-7-retrospective": "WIP",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Retrospective for Epic 7 in progress",
		},
		{
			name: "G23: sub-epic retrospective format",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-4-5":               "done",
					"4-5-1-feature":          "done",
					"epic-4-5-retrospective": "in-progress",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Retrospective for Epic 4-5 in progress",
		},
		{
			name: "G23: retro status started",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-7":               "done",
					"7-1-feature":          "done",
					"epic-7-retrospective": "started",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "Retrospective for Epic 7 in progress",
		},
		{
			name: "G23: deferred epic retrospective ignored",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":               "deferred-post-mvp",
					"epic-1-retrospective": "in-progress", // Retro for deferred epic - ignored
					"epic-2":               "done",
					"2-1-feature":          "done",
				},
			},
			wantStage:      domain.StageImplement,
			wantConfidence: domain.ConfidenceCertain,
			wantReasoning:  "All epics complete - project done", // Only non-deferred epics count
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
// naturalCompare Tests (Story 7.13: Numeric Story Sorting)
// =============================================================================

func TestNaturalCompare(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool // true if a < b
	}{
		// AC1: Stories sort numerically within epic
		{"story 7-2 < 7-10", "7-2", "7-10", true},
		{"story 7-10 > 7-2", "7-10", "7-2", false},
		{"story 7-1 < 7-2", "7-1", "7-2", true},
		{"story 7-9 < 7-10", "7-9", "7-10", true},
		{"story 7-10 < 7-11", "7-10", "7-11", true},

		// AC2: Epics sort numerically
		{"epic-2 < epic-10", "epic-2", "epic-10", true},
		{"epic-10 > epic-2", "epic-10", "epic-2", false},
		{"epic-1 < epic-2", "epic-1", "epic-2", true},
		{"epic-9 < epic-10", "epic-9", "epic-10", true},

		// AC3: Sub-epics sort correctly
		{"epic-4 < epic-4-5", "epic-4", "epic-4-5", true},
		{"epic-4-5 < epic-4-6", "epic-4-5", "epic-4-6", true},
		{"epic-4-6 < epic-5", "epic-4-6", "epic-5", true},
		{"epic-4-5 < epic-5", "epic-4-5", "epic-5", true},
		{"epic-10 > epic-4-5", "epic-10", "epic-4-5", false},

		// AC7: Retrospectives sort numerically
		{"epic-6-retrospective < epic-7-retrospective", "6", "7", true},
		{"epic-7-retrospective < epic-10-retrospective", "7", "10", true},
		{"epic-10-retrospective > epic-6-retrospective", "10", "6", false},

		// Full story keys with descriptors
		{"story key 7-2-feature < 7-10-feature", "7-2-feature", "7-10-feature", true},
		{"story key 7-10-feature > 7-2-feature", "7-10-feature", "7-2-feature", false},
		{"story key 1-1-feature < 1-2-feature", "1-1-feature", "1-2-feature", true},
		{"story key 1-9-feature < 1-10-feature", "1-9-feature", "1-10-feature", true},

		// Sub-epic story keys
		{"sub-epic story 4-5-2-xxx < 4-5-10-xxx", "4-5-2-xxx", "4-5-10-xxx", true},
		{"sub-epic story 4-5-10-xxx > 4-5-2-xxx", "4-5-10-xxx", "4-5-2-xxx", false},

		// Edge cases
		{"empty strings equal", "", "", false},
		{"empty < non-empty", "", "a", true},
		{"non-empty > empty", "a", "", false},
		{"equal strings", "abc", "abc", false},
		{"pure numeric 2 < 10", "2", "10", true},
		{"pure numeric 10 > 2", "10", "2", false},

		// M3 fix: Leading zeros handling
		{"leading zero 7-01 < 7-2", "7-01", "7-2", true},   // 01 = 1 < 2
		{"leading zero 7-1 == 7-01", "7-1", "7-01", false}, // 1 == 1, then lengths equal
		{"leading zero 7-01 == 7-1", "7-01", "7-1", false}, // 1 == 1, then lengths equal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := naturalCompare(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("naturalCompare(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}

			// M1 fix: Verify antisymmetry property for non-equal pairs
			// If a < b, then b > a (i.e., !naturalCompare(b, a))
			if tt.a != tt.b {
				reverse := naturalCompare(tt.b, tt.a)
				if tt.want && reverse {
					t.Errorf("antisymmetry violation: naturalCompare(%q, %q)=true but naturalCompare(%q, %q)=true also", tt.a, tt.b, tt.b, tt.a)
				}
				// Note: Both can be false for numerically equal values like "7-01" vs "7-1"
			}
		})
	}
}

// TestNaturalCompare_Antisymmetry explicitly verifies the antisymmetry property (M1 fix)
func TestNaturalCompare_Antisymmetry(t *testing.T) {
	// For any a != b: if naturalCompare(a, b) == true, then naturalCompare(b, a) must be false
	pairs := []struct{ a, b string }{
		{"7-2", "7-10"},
		{"epic-2", "epic-10"},
		{"epic-4", "epic-4-5"},
		{"1-1-feature", "1-2-feature"},
		{"a", "b"},
		{"1", "2"},
	}

	for _, p := range pairs {
		t.Run(p.a+"_vs_"+p.b, func(t *testing.T) {
			ab := naturalCompare(p.a, p.b)
			ba := naturalCompare(p.b, p.a)

			// Both can't be true (antisymmetry)
			if ab && ba {
				t.Errorf("antisymmetry violated: both naturalCompare(%q, %q) and naturalCompare(%q, %q) are true", p.a, p.b, p.b, p.a)
			}
			// Note: Both can be false for numerically equal values like "7-01" vs "7-1"
		})
	}
}

// =============================================================================
// Numeric Ordering Integration Tests (Story 7.13)
// =============================================================================

// TestDetermineStageFromStatus_DoubleDigitStories verifies natural ordering (AC1, AC4)
func TestDetermineStageFromStatus_DoubleDigitStories(t *testing.T) {
	tests := []struct {
		name          string
		status        *SprintStatus
		wantStage     domain.Stage
		wantReasoning string
	}{
		{
			name: "AC1/AC4: stories 7-1 done, 7-2 and 7-10 backlog - should pick 7-2",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-7":       "in-progress",
					"7-1-feature":  "done",
					"7-2-feature":  "backlog",
					"7-10-feature": "backlog",
					"7-11-feature": "backlog",
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Story 7.2 in backlog, needs drafting", // NOT 7.10
		},
		{
			name: "stories 7-1 through 7-12, 7-2 in-progress",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-7":       "in-progress",
					"7-1-feature":  "done",
					"7-2-feature":  "in-progress",
					"7-3-feature":  "backlog",
					"7-9-feature":  "backlog",
					"7-10-feature": "backlog",
					"7-11-feature": "backlog",
					"7-12-feature": "backlog",
				},
			},
			wantStage:     domain.StageImplement,
			wantReasoning: "Story 7.2 being implemented",
		},
		{
			name: "all stories backlog, should pick 7-1 not 7-10",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-7":       "in-progress",
					"7-1-feature":  "backlog",
					"7-10-feature": "backlog",
					"7-11-feature": "backlog",
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Story 7.1 in backlog, needs drafting", // NOT 7.10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, _, reasoning := determineStageFromStatus(tt.status)
			if stage != tt.wantStage {
				t.Errorf("stage = %v, want %v", stage, tt.wantStage)
			}
			if reasoning != tt.wantReasoning {
				t.Errorf("reasoning = %q, want %q", reasoning, tt.wantReasoning)
			}
		})
	}
}

// TestDetermineStageFromStatus_DoubleDigitEpics verifies epic ordering (AC2, AC3)
func TestDetermineStageFromStatus_DoubleDigitEpics(t *testing.T) {
	tests := []struct {
		name          string
		status        *SprintStatus
		wantStage     domain.Stage
		wantReasoning string
	}{
		{
			name: "AC2: epic-2 in-progress, epic-10 backlog",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":       "done",
					"epic-2":       "in-progress",
					"2-1-feature":  "in-progress",
					"epic-10":      "backlog",
					"10-1-feature": "backlog",
				},
			},
			wantStage:     domain.StageImplement,
			wantReasoning: "Story 2.1 being implemented", // epic-2 found first, not epic-10
		},
		{
			name: "AC2: epic-10 in-progress, epic-2 done",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":       "done",
					"epic-2":       "done",
					"2-1-feature":  "done",
					"epic-10":      "in-progress",
					"10-1-feature": "backlog",
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Story 10.1 in backlog, needs drafting",
		},
		{
			name: "AC3: sub-epic ordering epic-4, epic-4-5, epic-4-6, epic-5",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-4":        "done",
					"epic-4-5":      "in-progress",
					"4-5-1-feature": "backlog",
					"epic-4-6":      "backlog",
					"epic-5":        "backlog",
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Story 4.5.1 in backlog, needs drafting", // epic-4-5 is in-progress
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, _, reasoning := determineStageFromStatus(tt.status)
			if stage != tt.wantStage {
				t.Errorf("stage = %v, want %v", stage, tt.wantStage)
			}
			if reasoning != tt.wantReasoning {
				t.Errorf("reasoning = %q, want %q", reasoning, tt.wantReasoning)
			}
		})
	}
}

// TestDetermineStageFromStatus_DoubleDigitRetrospectives verifies retrospective ordering (AC7)
func TestDetermineStageFromStatus_DoubleDigitRetrospectives(t *testing.T) {
	tests := []struct {
		name          string
		status        *SprintStatus
		wantStage     domain.Stage
		wantReasoning string
	}{
		{
			name: "AC7: epic-6, epic-7, epic-10 retrospectives - should pick epic-6 first",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-6":                "done",
					"6-1-feature":           "done",
					"epic-7":                "done",
					"7-1-feature":           "done",
					"epic-10":               "done",
					"10-1-feature":          "done",
					"epic-6-retrospective":  "in-progress",
					"epic-7-retrospective":  "backlog",
					"epic-10-retrospective": "backlog",
				},
			},
			wantStage:     domain.StageImplement,
			wantReasoning: "Retrospective for Epic 6 in progress", // NOT epic-10
		},
		{
			name: "AC7: only epic-10 retrospective in-progress",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-6":                "done",
					"epic-7":                "done",
					"epic-10":               "done",
					"epic-6-retrospective":  "done",
					"epic-7-retrospective":  "done",
					"epic-10-retrospective": "in-progress",
				},
			},
			wantStage:     domain.StageImplement,
			wantReasoning: "Retrospective for Epic 10 in progress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, _, reasoning := determineStageFromStatus(tt.status)
			if stage != tt.wantStage {
				t.Errorf("stage = %v, want %v", stage, tt.wantStage)
			}
			if reasoning != tt.wantReasoning {
				t.Errorf("reasoning = %q, want %q", reasoning, tt.wantReasoning)
			}
		})
	}
}

// TestDetermineStageFromStatus_MixedDoubleDigitScenario verifies combined scenario (AC5)
func TestDetermineStageFromStatus_MixedDoubleDigitScenario(t *testing.T) {
	// This test verifies that existing G-tests still pass with new sorting
	// and exercises a complex scenario with sub-epics and double-digit stories
	status := &SprintStatus{
		DevelopmentStatus: map[string]string{
			// Epic 4 (sub-epic structure)
			"epic-4":         "done",
			"4-1-feature":    "done",
			"epic-4-5":       "done",
			"4-5-1-feature":  "done",
			"4-5-2-feature":  "done",
			"4-5-10-feature": "done",
			// Epic 7 (double-digit stories)
			"epic-7":       "in-progress",
			"7-1-feature":  "done",
			"7-2-feature":  "done",
			"7-9-feature":  "done",
			"7-10-feature": "in-progress", // This is correctly picked
			"7-11-feature": "backlog",
			// Epic 10
			"epic-10":      "backlog",
			"10-1-feature": "backlog",
		},
	}

	stage, confidence, reasoning := determineStageFromStatus(status)

	if stage != domain.StageImplement {
		t.Errorf("stage = %v, want StageImplement", stage)
	}
	if confidence != domain.ConfidenceCertain {
		t.Errorf("confidence = %v, want ConfidenceCertain", confidence)
	}
	if reasoning != "Story 7.10 being implemented" {
		t.Errorf("reasoning = %q, want %q", reasoning, "Story 7.10 being implemented")
	}
}

// =============================================================================
// G25-G27: Backlog Epic Stage Detection Tests (Story 9.5.7)
// =============================================================================

// TestDetermineStageFromStatus_G25_BacklogEpicWithStories tests that when all done epics
// are complete and the next epic is in backlog (not in-progress), we still analyze its stories.
func TestDetermineStageFromStatus_G25_BacklogEpicWithStories(t *testing.T) {
	tests := []struct {
		name          string
		status        *SprintStatus
		wantStage     domain.Stage
		wantReasoning string
	}{
		{
			name: "G25: done epics + backlog epic with backlog stories",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-9":                 "done",
					"9-1-feature":            "done",
					"epic-9-5":               "backlog",
					"9-5-1-feature":          "backlog",
					"9-5-2-feature":          "backlog",
					"epic-9-5-retrospective": "optional",
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Story 9.5.1 in backlog, needs drafting",
		},
		{
			name: "G25: done epics + backlog epic with ready-for-dev story",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-9":        "done",
					"9-1-feature":   "done",
					"epic-9-5":      "backlog",
					"9-5-1-feature": "ready-for-dev",
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Story 9.5.1 ready for development",
		},
		{
			name: "G25: done epics + backlog epic with drafted story",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-9":        "done",
					"9-1-feature":   "done",
					"epic-9-5":      "backlog",
					"9-5-1-feature": "drafted",
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Story 9.5.1 drafted, needs review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, _, reasoning := determineStageFromStatus(tt.status)
			if stage != tt.wantStage {
				t.Errorf("stage = %v, want %v", stage, tt.wantStage)
			}
			if reasoning != tt.wantReasoning {
				t.Errorf("reasoning = %q, want %q", reasoning, tt.wantReasoning)
			}
		})
	}
}

// TestDetermineStageFromStatus_G26_BacklogEpicNoStories tests that when a backlog epic
// has no stories defined, we return an appropriate planning message.
func TestDetermineStageFromStatus_G26_BacklogEpicNoStories(t *testing.T) {
	tests := []struct {
		name          string
		status        *SprintStatus
		wantStage     domain.Stage
		wantReasoning string
	}{
		{
			name: "G26: done epics + backlog epic with no stories",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-9":    "done",
					"9-1-story": "done",
					"epic-10":   "backlog",
					// No stories for epic-10
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Epic 10 in backlog, needs story planning",
		},
		{
			name: "G26: sub-epic (9.5) backlog with no stories",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-9":   "done",
					"epic-9-5": "backlog",
					// No stories for epic-9-5
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Epic 9.5 in backlog, needs story planning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, _, reasoning := determineStageFromStatus(tt.status)
			if stage != tt.wantStage {
				t.Errorf("stage = %v, want %v", stage, tt.wantStage)
			}
			if reasoning != tt.wantReasoning {
				t.Errorf("reasoning = %q, want %q", reasoning, tt.wantReasoning)
			}
		})
	}
}

// TestDetermineStageFromStatus_G27_MixedDoneBacklog tests that when some epics are done
// and some are backlog (none in-progress), we select the first backlog epic by natural order.
func TestDetermineStageFromStatus_G27_MixedDoneBacklog(t *testing.T) {
	tests := []struct {
		name          string
		status        *SprintStatus
		wantStage     domain.Stage
		wantReasoning string
	}{
		{
			name: "G27: done + multiple backlog epics, selects first backlog",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-8":        "done",
					"8-1-story":     "done",
					"epic-9":        "done",
					"9-1-story":     "done",
					"epic-10":       "backlog",
					"10-1-story":    "backlog",
					"epic-11":       "backlog",
					"11-1-story":    "backlog",
					"epic-9-5":      "backlog", // Out of order - should still pick epic-9-5 first
					"9-5-1-feature": "backlog",
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Story 9.5.1 in backlog, needs drafting", // epic-9-5 < epic-10 < epic-11
		},
		{
			name: "G27: all done except one backlog epic",
			status: &SprintStatus{
				DevelopmentStatus: map[string]string{
					"epic-1":      "done",
					"1-1-feature": "done",
					"epic-2":      "done",
					"2-1-feature": "done",
					"epic-3":      "backlog",
					"3-1-feature": "drafted",
				},
			},
			wantStage:     domain.StagePlan,
			wantReasoning: "Story 3.1 drafted, needs review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, _, reasoning := determineStageFromStatus(tt.status)
			if stage != tt.wantStage {
				t.Errorf("stage = %v, want %v", stage, tt.wantStage)
			}
			if reasoning != tt.wantReasoning {
				t.Errorf("reasoning = %q, want %q", reasoning, tt.wantReasoning)
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
			name: "_bmad-output location: _bmad-output/implementation-artifacts/sprint-status.yaml",
			setup: func(t *testing.T, dir string) {
				implDir := filepath.Join(dir, "_bmad-output", "implementation-artifacts")
				if err := os.MkdirAll(implDir, 0755); err != nil {
					t.Fatalf("failed to create _bmad-output/implementation-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(implDir, "sprint-status.yaml"), []byte("test"), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			wantFound: true,
			wantPath:  "_bmad-output/implementation-artifacts/sprint-status.yaml",
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
		{
			name: "docs takes precedence over _bmad-output",
			setup: func(t *testing.T, dir string) {
				// Create docs location
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte("docs"), 0644); err != nil {
					t.Fatalf("failed to write docs sprint-status.yaml: %v", err)
				}

				// Create _bmad-output location
				implDir := filepath.Join(dir, "_bmad-output", "implementation-artifacts")
				if err := os.MkdirAll(implDir, 0755); err != nil {
					t.Fatalf("failed to create _bmad-output/implementation-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(implDir, "sprint-status.yaml"), []byte("bmad-output"), 0644); err != nil {
					t.Fatalf("failed to write _bmad-output sprint-status.yaml: %v", err)
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

			// Test with nil config (hardcoded fallback paths)
			got, gotMtime := findSprintStatusPath(dir, nil)

			if tt.wantFound {
				wantFullPath := filepath.Join(dir, tt.wantPath)
				if got != wantFullPath {
					t.Errorf("findSprintStatusPath() = %q, want %q", got, wantFullPath)
				}
				// When found, should have non-zero mtime
				if gotMtime.IsZero() {
					t.Errorf("findSprintStatusPath() returned zero mtime for found path")
				}
			} else {
				if got != "" {
					t.Errorf("findSprintStatusPath() = %q, want empty string", got)
				}
				// When not found, mtime should be zero
				if !gotMtime.IsZero() {
					t.Errorf("findSprintStatusPath() returned non-zero mtime for not found: %v", gotMtime)
				}
			}
		})
	}
}

// TestFindSprintStatusPath_WithConfig tests config-based path resolution
func TestFindSprintStatusPath_WithConfig(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		config    *BMADConfig
		wantFound bool
		wantPath  string // relative to dir
	}{
		{
			name: "config with sprint_artifacts path",
			setup: func(t *testing.T, dir string) {
				sprintDir := filepath.Join(dir, "custom", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create custom sprint-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte("test"), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			config: &BMADConfig{
				SprintArtifacts: "{project-root}/custom/sprint-artifacts",
			},
			wantFound: true,
			wantPath:  "custom/sprint-artifacts/sprint-status.yaml",
		},
		{
			name: "config with implementation_artifacts path",
			setup: func(t *testing.T, dir string) {
				implDir := filepath.Join(dir, "_output", "impl-artifacts")
				if err := os.MkdirAll(implDir, 0755); err != nil {
					t.Fatalf("failed to create impl-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(implDir, "sprint-status.yaml"), []byte("test"), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			config: &BMADConfig{
				ImplementationArtifacts: "{project-root}/_output/impl-artifacts",
			},
			wantFound: true,
			wantPath:  "_output/impl-artifacts/sprint-status.yaml",
		},
		{
			name: "config with output_folder - finds in implementation-artifacts subdir",
			setup: func(t *testing.T, dir string) {
				implDir := filepath.Join(dir, "_bmad-output", "implementation-artifacts")
				if err := os.MkdirAll(implDir, 0755); err != nil {
					t.Fatalf("failed to create implementation-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(implDir, "sprint-status.yaml"), []byte("test"), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			config: &BMADConfig{
				OutputFolder: "{project-root}/_bmad-output",
			},
			wantFound: true,
			wantPath:  "_bmad-output/implementation-artifacts/sprint-status.yaml",
		},
		{
			name: "sprint_artifacts takes priority over implementation_artifacts",
			setup: func(t *testing.T, dir string) {
				// Create both locations
				sprintDir := filepath.Join(dir, "sprint-loc")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-loc: %v", err)
				}
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte("sprint"), 0644); err != nil {
					t.Fatalf("failed to write sprint sprint-status.yaml: %v", err)
				}

				implDir := filepath.Join(dir, "impl-loc")
				if err := os.MkdirAll(implDir, 0755); err != nil {
					t.Fatalf("failed to create impl-loc: %v", err)
				}
				if err := os.WriteFile(filepath.Join(implDir, "sprint-status.yaml"), []byte("impl"), 0644); err != nil {
					t.Fatalf("failed to write impl sprint-status.yaml: %v", err)
				}
			},
			config: &BMADConfig{
				SprintArtifacts:         "{project-root}/sprint-loc",
				ImplementationArtifacts: "{project-root}/impl-loc",
			},
			wantFound: true,
			wantPath:  "sprint-loc/sprint-status.yaml",
		},
		{
			name: "config path not found, falls back to hardcoded",
			setup: func(t *testing.T, dir string) {
				// Only create the hardcoded fallback location
				sprintDir := filepath.Join(dir, "docs", "sprint-artifacts")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					t.Fatalf("failed to create sprint-artifacts: %v", err)
				}
				if err := os.WriteFile(filepath.Join(sprintDir, "sprint-status.yaml"), []byte("fallback"), 0644); err != nil {
					t.Fatalf("failed to write sprint-status.yaml: %v", err)
				}
			},
			config: &BMADConfig{
				SprintArtifacts: "{project-root}/nonexistent/path",
			},
			wantFound: true,
			wantPath:  "docs/sprint-artifacts/sprint-status.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			got, gotMtime := findSprintStatusPath(dir, tt.config)

			if tt.wantFound {
				wantFullPath := filepath.Join(dir, tt.wantPath)
				if got != wantFullPath {
					t.Errorf("findSprintStatusPath() = %q, want %q", got, wantFullPath)
				}
				// When found, should have non-zero mtime
				if gotMtime.IsZero() {
					t.Errorf("findSprintStatusPath() returned zero mtime for found path")
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
			// Pass bmadDir for config-based path resolution
			bmadDir := filepath.Join(dir, ".bmad")
			stage, confidence, reasoning, _ := d.detectStage(context.Background(), dir, bmadDir)

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

	bmadDir := filepath.Join(dir, ".bmad")
	stage, confidence, reasoning, artifactMtime := d.detectStage(ctx, dir, bmadDir)

	if stage != domain.StageUnknown {
		t.Errorf("detectStage() with cancelled context stage = %v, want StageUnknown", stage)
	}
	if confidence != domain.ConfidenceUncertain {
		t.Errorf("detectStage() with cancelled context confidence = %v, want ConfidenceUncertain", confidence)
	}
	if reasoning != "" {
		t.Errorf("detectStage() with cancelled context reasoning = %q, want empty", reasoning)
	}
	if !artifactMtime.IsZero() {
		t.Errorf("detectStage() with cancelled context artifactMtime = %v, want zero", artifactMtime)
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
