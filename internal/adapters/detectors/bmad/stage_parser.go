// Package bmad implements detection for the BMAD v6 workflow methodology.
// This file contains sprint-status.yaml parsing and stage determination logic.
package bmad

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// SprintStatus represents the parsed sprint-status.yaml file.
type SprintStatus struct {
	DevelopmentStatus map[string]string `yaml:"development_status"`
}

// epicKeyRegex matches epic keys like "epic-1", "epic-4-5".
var epicKeyRegex = regexp.MustCompile(`^epic-\d+(-\d+)?$`)

// normalizeStatus converts common LLM variations to canonical status values.
// Apply BEFORE switch statement comparison.
func normalizeStatus(status string) string {
	// 1. Lowercase and trim
	s := strings.ToLower(strings.TrimSpace(status))

	// 2. Normalize separators: spaces and underscores → hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// 3. Map synonyms (G17)
	synonyms := map[string]string{
		"complete":    "done",
		"completed":   "done",
		"finished":    "done",
		"wip":         "in-progress",
		"inprogress":  "in-progress",
		"reviewing":   "review",
		"in-review":   "review",
		"code-review": "review",
	}

	if canonical, ok := synonyms[s]; ok {
		return canonical
	}
	return s
}

// storyKeyRegex matches story keys like "1-1-project-scaffolding", "4-5-2-bmad-v6-...".
var storyKeyRegex = regexp.MustCompile(`^\d+-\d+-`)

// parseSprintStatus reads and parses the sprint-status.yaml file.
func parseSprintStatus(ctx context.Context, path string) (*SprintStatus, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var status SprintStatus
	if err := yaml.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// determineStageFromStatus analyzes the sprint status and returns the current stage.
// Returns (stage, confidence, reasoning).
func determineStageFromStatus(status *SprintStatus) (domain.Stage, domain.Confidence, string) {
	if status == nil || status.DevelopmentStatus == nil || len(status.DevelopmentStatus) == 0 {
		return domain.StageUnknown, domain.ConfidenceUncertain, "sprint-status.yaml is empty"
	}

	// Collect epics and stories
	type epicInfo struct {
		key     string
		status  string
		stories []struct {
			key    string
			status string
		}
	}

	epics := make(map[string]*epicInfo)
	var epicOrder []string // Preserve order for finding first in-progress

	// First pass: collect epics
	for key, value := range status.DevelopmentStatus {
		// Skip retrospectives
		if strings.HasSuffix(key, "-retrospective") {
			continue
		}

		if epicKeyRegex.MatchString(key) {
			epics[key] = &epicInfo{
				key:    key,
				status: normalizeStatus(value),
			}
			epicOrder = append(epicOrder, key)
		}
	}

	// Sort epic order for deterministic behavior (map iteration is random)
	sort.Strings(epicOrder)

	// G14/G22: Track data quality warnings
	var warnings []string

	// Second pass: associate stories with epics
	for key, value := range status.DevelopmentStatus {
		// Skip retrospectives
		if strings.HasSuffix(key, "-retrospective") {
			continue
		}

		// G22: Check for empty status values
		if strings.TrimSpace(value) == "" {
			warnings = append(warnings, "empty status for "+key)
			continue
		}

		if storyKeyRegex.MatchString(key) {
			// Find parent epic by matching prefix
			// Story "4-5-2-xxx" belongs to epic "epic-4-5"
			storyPrefix := extractStoryPrefix(key)
			epicKey := "epic-" + storyPrefix

			if epic, ok := epics[epicKey]; ok {
				epic.stories = append(epic.stories, struct {
					key    string
					status string
				}{
					key:    key,
					status: normalizeStatus(value),
				})
			} else {
				// G14: Orphan story (no matching epic)
				warnings = append(warnings, "orphan story "+formatStoryKey(key))
			}
		}
	}

	// Helper to append warnings to reasoning
	appendWarnings := func(reasoning string) string {
		if len(warnings) > 0 {
			sort.Strings(warnings) // Deterministic warning order
			return reasoning + " [Warning: " + strings.Join(warnings, "; ") + "]"
		}
		return reasoning
	}

	// Count epics by status
	backlogCount := 0
	doneCount := 0
	var firstInProgressEpic *epicInfo

	for _, epicKey := range epicOrder {
		epic := epics[epicKey]
		switch epic.status {
		case "backlog":
			backlogCount++
		case "in-progress", "contexted":
			if firstInProgressEpic == nil {
				firstInProgressEpic = epic
			}
		case "done":
			doneCount++
		}
	}

	// G7: Check for done epics with active stories (check before all-done shortcut)
	for _, epicKey := range epicOrder {
		epic := epics[epicKey]
		if epic.status == "done" {
			for _, story := range epic.stories {
				if story.status == "review" {
					return domain.StageTasks, domain.ConfidenceLikely,
						appendWarnings("Epic done but Story " + formatStoryKey(story.key) + " in review")
				}
				if story.status == "in-progress" {
					return domain.StageImplement, domain.ConfidenceLikely,
						appendWarnings("Epic done but Story " + formatStoryKey(story.key) + " in-progress")
				}
			}
		}
	}

	// All epics done
	if len(epics) > 0 && doneCount == len(epics) {
		return domain.StageImplement, domain.ConfidenceCertain, appendWarnings("All epics complete - project done")
	}

	// G8: Check for backlog epics with active stories
	for _, epicKey := range epicOrder {
		epic := epics[epicKey]
		if epic.status == "backlog" {
			for _, story := range epic.stories {
				if story.status == "in-progress" || story.status == "done" || story.status == "review" {
					return domain.StageSpecify, domain.ConfidenceLikely,
						appendWarnings("Epic backlog but Story " + formatStoryKey(story.key) + " active")
				}
			}
		}
	}

	// All epics backlog
	if len(epics) > 0 && backlogCount == len(epics) {
		return domain.StageSpecify, domain.ConfidenceCertain, appendWarnings("No epics in progress - planning phase")
	}

	// Has in-progress epic - analyze its stories
	if firstInProgressEpic != nil {
		// G19: Sort stories for deterministic ordering
		sortedStories := make([]struct {
			key    string
			status string
		}, len(firstInProgressEpic.stories))
		copy(sortedStories, firstInProgressEpic.stories)
		sort.Slice(sortedStories, func(i, j int) bool {
			return sortedStories[i].key < sortedStories[j].key
		})

		// G2/G3/G19: Priority-based story selection
		// Priority: review > in-progress > ready-for-dev > drafted > backlog > done
		storyPriority := map[string]int{
			"review":        1,
			"in-progress":   2,
			"ready-for-dev": 3,
			"drafted":       4,
			"backlog":       5,
			"done":          6,
		}

		var selectedStory string
		var selectedStatus string
		const unsetPriority = 999
		selectedPriority := unsetPriority

		for _, story := range sortedStories {
			if p, ok := storyPriority[story.status]; ok && p < selectedPriority {
				selectedStory = story.key
				selectedStatus = story.status
				selectedPriority = p
			}
		}

		// Return based on selected story status
		switch selectedStatus {
		case "review":
			return domain.StageTasks, domain.ConfidenceCertain,
				appendWarnings("Story " + formatStoryKey(selectedStory) + " in code review")
		case "in-progress":
			return domain.StageImplement, domain.ConfidenceCertain,
				appendWarnings("Story " + formatStoryKey(selectedStory) + " being implemented")
		case "ready-for-dev":
			return domain.StagePlan, domain.ConfidenceCertain,
				appendWarnings("Story " + formatStoryKey(selectedStory) + " ready for development")
		case "drafted":
			return domain.StagePlan, domain.ConfidenceCertain,
				appendWarnings("Story " + formatStoryKey(selectedStory) + " drafted, needs review")
		}

		// G1: Check if ALL stories in this epic are done
		allDone := true
		hasStories := len(firstInProgressEpic.stories) > 0
		for _, story := range firstInProgressEpic.stories {
			if story.status != "done" {
				allDone = false
				break
			}
		}
		if hasStories && allDone {
			return domain.StageImplement, domain.ConfidenceCertain,
				appendWarnings(formatEpicKey(firstInProgressEpic.key) + " stories complete, update epic status")
		}

		// Epic in-progress but no stories started
		return domain.StagePlan, domain.ConfidenceCertain,
			appendWarnings(formatEpicKey(firstInProgressEpic.key) + " started, preparing stories")
	}

	// Fallback for unexpected states
	return domain.StageUnknown, domain.ConfidenceUncertain, appendWarnings("Unable to determine stage from sprint status")
}

// extractStoryPrefix extracts the epic prefix from a story key.
// "4-5-2-bmad-v6-xxx" -> "4-5"
// "1-1-project-scaffolding" -> "1"
func extractStoryPrefix(storyKey string) string {
	// Match the numeric prefix pattern (e.g., "4-5-2" or "1-1")
	parts := strings.Split(storyKey, "-")
	if len(parts) < 2 {
		return ""
	}

	// Find where non-numeric parts start
	var numericParts []string
	for i, part := range parts {
		if isNumeric(part) {
			numericParts = append(numericParts, part)
		} else {
			// Non-numeric part found, take all but the last numeric part
			if i > 1 {
				return strings.Join(numericParts[:len(numericParts)-1], "-")
			}
			// Only one numeric part before non-numeric
			return numericParts[0]
		}
	}

	// All parts are numeric (unlikely but handle it)
	if len(numericParts) > 1 {
		return strings.Join(numericParts[:len(numericParts)-1], "-")
	}
	return numericParts[0]
}

// isNumeric checks if a string contains only digits.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// formatStoryKey formats a story key for display.
// "4-5-2-bmad-v6-xxx" -> "4.5.2"
func formatStoryKey(key string) string {
	parts := strings.Split(key, "-")
	var numericParts []string
	for _, part := range parts {
		if isNumeric(part) {
			numericParts = append(numericParts, part)
		} else {
			break
		}
	}
	return strings.Join(numericParts, ".")
}

// formatEpicKey formats an epic key for display.
// "epic-4-5" -> "Epic 4.5"
func formatEpicKey(key string) string {
	// Remove "epic-" prefix
	numPart := strings.TrimPrefix(key, "epic-")
	// Replace dashes with dots
	return "Epic " + strings.ReplaceAll(numPart, "-", ".")
}

// detectStageFromArtifacts checks for BMAD artifact files as a fallback.
// This is used when sprint-status.yaml is not found or cannot be parsed.
func detectStageFromArtifacts(ctx context.Context, projectPath string) (domain.Stage, domain.Confidence, string, error) {
	docsPath := filepath.Join(projectPath, "docs")

	// Check context before scanning
	select {
	case <-ctx.Done():
		return domain.StageUnknown, domain.ConfidenceUncertain, "", ctx.Err()
	default:
	}

	// Check for epics file (highest priority - furthest along)
	epicPatterns := []string{"*epic*.md", "*Epic*.md"}
	for _, pattern := range epicPatterns {
		matches, _ := filepath.Glob(filepath.Join(docsPath, pattern))
		if len(matches) > 0 {
			return domain.StageImplement, domain.ConfidenceLikely, "Epics defined but no sprint status", nil
		}
	}

	select {
	case <-ctx.Done():
		return domain.StageUnknown, domain.ConfidenceUncertain, "", ctx.Err()
	default:
	}

	// Check for architecture file
	archPatterns := []string{"*architecture*.md", "*Architecture*.md"}
	for _, pattern := range archPatterns {
		matches, _ := filepath.Glob(filepath.Join(docsPath, pattern))
		if len(matches) > 0 {
			return domain.StagePlan, domain.ConfidenceLikely, "Architecture designed, no epics yet", nil
		}
	}

	select {
	case <-ctx.Done():
		return domain.StageUnknown, domain.ConfidenceUncertain, "", ctx.Err()
	default:
	}

	// Check for PRD file
	prdPatterns := []string{"*prd*.md", "*PRD*.md"}
	for _, pattern := range prdPatterns {
		matches, _ := filepath.Glob(filepath.Join(docsPath, pattern))
		if len(matches) > 0 {
			return domain.StageSpecify, domain.ConfidenceLikely, "PRD created, architecture pending", nil
		}
	}

	return domain.StageUnknown, domain.ConfidenceUncertain, "No BMAD artifacts detected", nil
}

// findSprintStatusPath searches for sprint-status.yaml in standard locations.
func findSprintStatusPath(projectPath string) string {
	// Primary location: docs/sprint-artifacts/sprint-status.yaml
	primary := filepath.Join(projectPath, "docs", "sprint-artifacts", "sprint-status.yaml")
	if _, err := os.Stat(primary); err == nil {
		return primary
	}

	// Alternative location: docs/sprint-status.yaml
	alt := filepath.Join(projectPath, "docs", "sprint-status.yaml")
	if _, err := os.Stat(alt); err == nil {
		return alt
	}

	return ""
}

// detectStage performs stage detection for a BMAD v6 project.
// It first tries to parse sprint-status.yaml, then falls back to artifact detection.
func (d *BMADDetector) detectStage(ctx context.Context, path string) (domain.Stage, domain.Confidence, string) {
	// Check context first
	select {
	case <-ctx.Done():
		return domain.StageUnknown, domain.ConfidenceUncertain, ""
	default:
	}

	// Try to find and parse sprint-status.yaml
	statusPath := findSprintStatusPath(path)
	if statusPath != "" {
		status, err := parseSprintStatus(ctx, statusPath)
		if err != nil {
			// Parse error - return unknown with reason
			return domain.StageUnknown, domain.ConfidenceUncertain, "sprint-status.yaml parse error"
		}

		return determineStageFromStatus(status)
	}

	// Fallback to artifact detection
	stage, confidence, reasoning, err := detectStageFromArtifacts(ctx, path)
	if err != nil {
		// Context cancellation
		return domain.StageUnknown, domain.ConfidenceUncertain, ""
	}

	return stage, confidence, reasoning
}
