package stageformat

import (
	"strings"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// FormatStageInfo returns condensed stage info for project list display.
// For BMAD: Parses DetectionReasoning -> "E8 S8.3 review"
// For Speckit: Returns CurrentStage.String() -> "Plan", "Specify", etc.
// For Unknown: Returns "-"
func FormatStageInfo(p *domain.Project) string {
	// Handle nil pointer
	if p == nil {
		return "-"
	}

	// Handle unknown/empty method first
	if p.DetectedMethod == "" || p.DetectedMethod == "unknown" {
		return "-"
	}

	// Handle unknown stage
	if p.CurrentStage == domain.StageUnknown {
		return "-"
	}

	// BMAD: Parse rich reasoning
	if p.DetectedMethod == "bmad" {
		if result := parseBMADReasoning(p.DetectionReasoning); result != "" {
			return result
		}
		// Fallback to CurrentStage for unknown patterns
		return p.CurrentStage.String()
	}

	// Speckit and others: Use CurrentStage directly
	return p.CurrentStage.String()
}

// FormatStageInfoWithWidth returns stage info truncated to maxWidth.
// Adds "..." if truncated.
func FormatStageInfoWithWidth(p *domain.Project, maxWidth int) string {
	info := FormatStageInfo(p)
	if len(info) <= maxWidth {
		return info
	}
	if maxWidth <= 3 {
		return info[:maxWidth]
	}
	return info[:maxWidth-3] + "..."
}

// parseBMADReasoning extracts display info from DetectionReasoning.
func parseBMADReasoning(reasoning string) string {
	if reasoning == "" {
		return ""
	}

	// Strip "BMAD vX.X.X, " prefix if present (detector.go:130 adds this)
	reasoning = stripBMADVersionPrefix(reasoning)

	switch {
	case strings.HasPrefix(reasoning, "Story "):
		return parseStoryReasoning(reasoning)
	case strings.HasPrefix(reasoning, "Epic "):
		return parseEpicReasoning(reasoning)
	case strings.Contains(reasoning, "Retrospective for Epic"):
		return parseRetroReasoning(reasoning)
	case strings.Contains(reasoning, "All epics complete"):
		return "Done"
	case strings.Contains(reasoning, "planning phase"):
		return "Planning"
	default:
		return ""
	}
}

// stripBMADVersionPrefix removes "BMAD vX.X.X, " prefix from reasoning.
// The BMAD detector (detector.go:130) prepends version info to stage reasoning.
func stripBMADVersionPrefix(reasoning string) string {
	// Pattern: "BMAD vX.X.X-suffix, actual reasoning"
	const prefix = "BMAD v"
	if !strings.HasPrefix(reasoning, prefix) {
		return reasoning
	}
	// Find the ", " separator after version
	idx := strings.Index(reasoning, ", ")
	if idx == -1 {
		return reasoning
	}
	return reasoning[idx+2:]
}

// parseStoryReasoning handles "Story X.Y.Z status" patterns.
// "Story 8.3 in code review" -> "E8 S8.3 review"
// "Story 4.5.2 being implemented" -> "E4 S4.5.2 impl"
func parseStoryReasoning(reasoning string) string {
	// Remove "Story " prefix
	rest := strings.TrimPrefix(reasoning, "Story ")
	if rest == "" {
		return ""
	}

	// Find the first space after story number
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) == 0 {
		return ""
	}

	storyNum := parts[0] // e.g., "8.3" or "4.5.2"

	// Extract epic number (first number before first dot)
	epicNum := extractEpicFromStory(storyNum)
	if epicNum == "" {
		return ""
	}

	// Determine status abbreviation from the rest
	statusAbbr := ""
	if len(parts) > 1 {
		statusAbbr = abbreviateStatus(parts[1])
	}

	// Format: "E{epic} S{story} {status}"
	if statusAbbr != "" {
		return "E" + epicNum + " S" + storyNum + " " + statusAbbr
	}
	return "E" + epicNum + " S" + storyNum
}

// parseEpicReasoning handles "Epic X.Y status" patterns.
// "Epic 4.5 started, preparing stories" -> "E4.5 prep"
// "Epic 4.5 stories complete, update epic status" -> "E4.5 done"
func parseEpicReasoning(reasoning string) string {
	// Remove "Epic " prefix
	rest := strings.TrimPrefix(reasoning, "Epic ")
	if rest == "" {
		return ""
	}

	// Find the first space after epic number
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) == 0 {
		return ""
	}

	epicNum := parts[0] // e.g., "4.5" or "8"

	// Determine status abbreviation
	statusAbbr := ""
	if len(parts) > 1 {
		statusAbbr = abbreviateEpicStatus(parts[1])
	}

	if statusAbbr != "" {
		return "E" + epicNum + " " + statusAbbr
	}
	return "E" + epicNum
}

// parseRetroReasoning handles retrospective patterns.
// "Retrospective for Epic 7 in progress" -> "E7 retro"
func parseRetroReasoning(reasoning string) string {
	// Find "Epic " after "Retrospective for "
	idx := strings.Index(reasoning, "Retrospective for Epic ")
	if idx == -1 {
		return ""
	}

	rest := reasoning[idx+len("Retrospective for Epic "):]
	// Extract epic number (until space or end)
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) == 0 {
		return ""
	}

	epicNum := parts[0]
	return "E" + epicNum + " retro"
}

// extractEpicFromStory extracts the epic number from a story number.
// "8.3" -> "8"
// "4.5.2" -> "4"
// "" -> ""
func extractEpicFromStory(storyNum string) string {
	if storyNum == "" {
		return ""
	}
	parts := strings.Split(storyNum, ".")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return parts[0]
}

// abbreviateStatus converts status description to abbreviation.
func abbreviateStatus(status string) string {
	status = strings.ToLower(status)

	switch {
	case strings.Contains(status, "code review") || strings.Contains(status, "in review"):
		return "review"
	case strings.Contains(status, "being implemented") || strings.Contains(status, "in-progress") || strings.Contains(status, "in progress"):
		return "impl"
	case strings.Contains(status, "ready for development") || strings.Contains(status, "ready-for-dev"):
		return "ready"
	case strings.Contains(status, "drafted"):
		return "draft"
	case strings.Contains(status, "backlog"):
		return "backlog"
	case strings.HasPrefix(status, "done") || strings.Contains(status, "completed"):
		return "done"
	default:
		return ""
	}
}

// abbreviateEpicStatus converts epic status description to abbreviation.
func abbreviateEpicStatus(status string) string {
	status = strings.ToLower(status)

	switch {
	case strings.Contains(status, "stories complete"):
		return "done"
	case strings.Contains(status, "started") || strings.Contains(status, "preparing"):
		return "prep"
	default:
		return ""
	}
}
