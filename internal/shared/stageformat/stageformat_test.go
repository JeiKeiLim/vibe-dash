package stageformat

import (
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestFormatStageInfo(t *testing.T) {
	tests := []struct {
		name     string
		project  domain.Project
		expected string
	}{
		// BMAD reasoning patterns
		{
			name: "bmad story review",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 8.3 in code review",
				CurrentStage:       domain.StageTasks,
			},
			expected: "E8 S8.3 review",
		},
		{
			name: "bmad story impl",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 4.5.2 being implemented",
				CurrentStage:       domain.StageImplement,
			},
			expected: "E4 S4.5.2 impl",
		},
		{
			name: "bmad story ready",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 1.2 ready for development",
				CurrentStage:       domain.StagePlan,
			},
			expected: "E1 S1.2 ready",
		},
		{
			name: "bmad story drafted",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 1.2 drafted, needs review",
				CurrentStage:       domain.StagePlan,
			},
			expected: "E1 S1.2 draft",
		},
		{
			name: "bmad story backlog",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 1.2 in backlog, needs drafting",
				CurrentStage:       domain.StagePlan,
			},
			expected: "E1 S1.2 backlog",
		},
		{
			name: "bmad epic prep",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Epic 4.5 started, preparing stories",
				CurrentStage:       domain.StagePlan,
			},
			expected: "E4.5 prep",
		},
		{
			name: "bmad epic done",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Epic 4.5 stories complete, update epic status",
				CurrentStage:       domain.StageImplement,
			},
			expected: "E4.5 done",
		},
		{
			name: "bmad retro",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Retrospective for Epic 7 in progress",
				CurrentStage:       domain.StageImplement,
			},
			expected: "E7 retro",
		},
		{
			name: "bmad all done",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "All epics complete - project done",
				CurrentStage:       domain.StageImplement,
			},
			expected: "Done",
		},
		{
			name: "bmad planning",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "No epics in progress - planning phase",
				CurrentStage:       domain.StageSpecify,
			},
			expected: "Planning",
		},
		{
			name: "bmad empty reasoning",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "",
				CurrentStage:       domain.StagePlan,
			},
			expected: "Plan",
		},
		{
			name: "bmad unknown pattern",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Something unexpected",
				CurrentStage:       domain.StageImplement,
			},
			expected: "Implement",
		},

		// Speckit - uses CurrentStage.String() directly
		{
			name: "speckit specify",
			project: domain.Project{
				DetectedMethod: "speckit",
				CurrentStage:   domain.StageSpecify,
			},
			expected: "Specify",
		},
		{
			name: "speckit plan",
			project: domain.Project{
				DetectedMethod: "speckit",
				CurrentStage:   domain.StagePlan,
			},
			expected: "Plan",
		},
		{
			name: "speckit tasks",
			project: domain.Project{
				DetectedMethod: "speckit",
				CurrentStage:   domain.StageTasks,
			},
			expected: "Tasks",
		},
		{
			name: "speckit implement",
			project: domain.Project{
				DetectedMethod: "speckit",
				CurrentStage:   domain.StageImplement,
			},
			expected: "Implement",
		},

		// Unknown method - return "-"
		{
			name: "unknown method",
			project: domain.Project{
				DetectedMethod: "unknown",
			},
			expected: "-",
		},
		{
			name: "empty method",
			project: domain.Project{
				DetectedMethod: "",
			},
			expected: "-",
		},
		{
			name: "unknown stage",
			project: domain.Project{
				DetectedMethod: "speckit",
				CurrentStage:   domain.StageUnknown,
			},
			expected: "-",
		},

		// Edge cases
		{
			name: "bmad long story number",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 10.10.10 being implemented",
				CurrentStage:       domain.StageImplement,
			},
			expected: "E10 S10.10.10 impl",
		},
		{
			name: "bmad sub-epic retro",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Retrospective for Epic 4.5 in progress",
				CurrentStage:       domain.StageImplement,
			},
			expected: "E4.5 retro",
		},
		{
			name: "bmad story done",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 1.2 done and completed",
				CurrentStage:       domain.StageImplement,
			},
			expected: "E1 S1.2 done",
		},
		{
			name: "bmad unknown status pattern",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 1.2 has unknown status 'custom'",
				CurrentStage:       domain.StagePlan,
			},
			expected: "E1 S1.2", // No status abbreviation for unknown patterns
		},
		// BMAD version prefix handling (detector.go:130 adds version prefix)
		{
			name: "bmad with version prefix story",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "BMAD v6.0.0-alpha.13, Story 8.4 in backlog, needs drafting",
				CurrentStage:       domain.StagePlan,
			},
			expected: "E8 S8.4 backlog",
		},
		{
			name: "bmad with version prefix impl",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "BMAD v6.0.0-alpha.13, Story 1.1 being implemented",
				CurrentStage:       domain.StageImplement,
			},
			expected: "E1 S1.1 impl",
		},
		{
			name: "bmad with version prefix all done",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "BMAD v6.0.0-alpha.13, All epics complete - project done",
				CurrentStage:       domain.StageImplement,
			},
			expected: "Done",
		},
		{
			name: "bmad with version prefix review",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "BMAD v6.0.0-alpha.13, Story 1.1 in code review",
				CurrentStage:       domain.StageTasks,
			},
			expected: "E1 S1.1 review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatStageInfo(&tt.project)
			if result != tt.expected {
				t.Errorf("FormatStageInfo() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatStageInfoWithWidth(t *testing.T) {
	tests := []struct {
		name     string
		project  domain.Project
		maxWidth int
		expected string
	}{
		{
			name: "fits within width",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 8.3 in code review",
				CurrentStage:       domain.StageTasks,
			},
			maxWidth: 20,
			expected: "E8 S8.3 review",
		},
		{
			name: "truncate with ellipsis",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 8.3 in code review",
				CurrentStage:       domain.StageTasks,
			},
			maxWidth: 10,
			expected: "E8 S8.3...",
		},
		{
			name: "very narrow truncate",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 8.3 in code review",
				CurrentStage:       domain.StageTasks,
			},
			maxWidth: 5,
			expected: "E8...",
		},
		{
			name: "extremely narrow no ellipsis",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 8.3 in code review",
				CurrentStage:       domain.StageTasks,
			},
			maxWidth: 3,
			expected: "E8 ",
		},
		{
			name: "width 2 no ellipsis",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 8.3 in code review",
				CurrentStage:       domain.StageTasks,
			},
			maxWidth: 2,
			expected: "E8",
		},
		{
			name: "width 4 with ellipsis",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 8.3 in code review",
				CurrentStage:       domain.StageTasks,
			},
			maxWidth: 4,
			expected: "E...",
		},
		{
			name: "exact width match",
			project: domain.Project{
				DetectedMethod:     "bmad",
				DetectionReasoning: "Story 8.3 in code review",
				CurrentStage:       domain.StageTasks,
			},
			maxWidth: 14, // "E8 S8.3 review" is 14 chars
			expected: "E8 S8.3 review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatStageInfoWithWidth(&tt.project, tt.maxWidth)
			if result != tt.expected {
				t.Errorf("FormatStageInfoWithWidth() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseBMADReasoning(t *testing.T) {
	tests := []struct {
		name      string
		reasoning string
		expected  string
	}{
		{"empty", "", ""},
		{"story review", "Story 8.3 in code review", "E8 S8.3 review"},
		{"story impl", "Story 4.5.2 being implemented", "E4 S4.5.2 impl"},
		{"story ready", "Story 1.2 ready for development", "E1 S1.2 ready"},
		{"story drafted", "Story 1.2 drafted, needs review", "E1 S1.2 draft"},
		{"story backlog", "Story 1.2 in backlog, needs drafting", "E1 S1.2 backlog"},
		{"epic prep", "Epic 4.5 started, preparing stories", "E4.5 prep"},
		{"epic done", "Epic 4.5 stories complete, update epic status", "E4.5 done"},
		{"retro", "Retrospective for Epic 7 in progress", "E7 retro"},
		{"all done", "All epics complete - project done", "Done"},
		{"planning", "No epics in progress - planning phase", "Planning"},
		{"unknown pattern", "Something unexpected", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBMADReasoning(tt.reasoning)
			if result != tt.expected {
				t.Errorf("parseBMADReasoning(%q) = %q, want %q", tt.reasoning, result, tt.expected)
			}
		})
	}
}

func TestExtractEpicFromStory(t *testing.T) {
	tests := []struct {
		storyNum string
		expected string
	}{
		{"8.3", "8"},
		{"4.5.2", "4"},
		{"1.2", "1"},
		{"10.10.10", "10"},
		{"1", "1"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.storyNum, func(t *testing.T) {
			result := extractEpicFromStory(tt.storyNum)
			if result != tt.expected {
				t.Errorf("extractEpicFromStory(%q) = %q, want %q", tt.storyNum, result, tt.expected)
			}
		})
	}
}

func TestFormatStageInfo_NilProject(t *testing.T) {
	result := FormatStageInfo(nil)
	if result != "-" {
		t.Errorf("FormatStageInfo(nil) = %q, want %q", result, "-")
	}
}

func TestStripBMADVersionPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"BMAD v6.0.0-alpha.13, Story 8.3 in code review", "Story 8.3 in code review"},
		{"BMAD v6.0.0, All epics complete - project done", "All epics complete - project done"},
		{"BMAD v1.0.0, Epic 4.5 started, preparing stories", "Epic 4.5 started, preparing stories"},
		{"Story 8.3 in code review", "Story 8.3 in code review"},                             // No prefix
		{"BMAD v6.0.0-alpha.13 without separator", "BMAD v6.0.0-alpha.13 without separator"}, // No ", " separator
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := stripBMADVersionPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("stripBMADVersionPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatStageInfoWithWidth_NilProject(t *testing.T) {
	result := FormatStageInfoWithWidth(nil, 10)
	if result != "-" {
		t.Errorf("FormatStageInfoWithWidth(nil, 10) = %q, want %q", result, "-")
	}
}

func TestAbbreviateStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"in code review", "review"},
		{"in review", "review"},
		{"being implemented", "impl"},
		{"in-progress", "impl"},
		{"in progress", "impl"},
		{"ready for development", "ready"},
		{"ready-for-dev", "ready"},
		{"drafted", "draft"},
		{"drafted, needs review", "draft"},
		{"in backlog", "backlog"},
		{"in backlog, needs drafting", "backlog"},
		{"done", "done"},
		{"done and completed", "done"},
		{"completed successfully", "done"},
		{"something unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := abbreviateStatus(tt.status)
			if result != tt.expected {
				t.Errorf("abbreviateStatus(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

func TestAbbreviateEpicStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"stories complete", "done"},
		{"stories complete, update epic status", "done"},
		{"started, preparing stories", "prep"},
		{"preparing stories", "prep"},
		{"started", "prep"},
		{"something unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := abbreviateEpicStatus(tt.status)
			if result != tt.expected {
				t.Errorf("abbreviateEpicStatus(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}
