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

// retroKeyRegex matches retrospective keys like "epic-7-retrospective", "epic-4-5-retrospective".
// G23: Used to detect retrospectives when all epics are done.
var retroKeyRegex = regexp.MustCompile(`^epic-(\d+(?:-\d+)?)-retrospective$`)

// normalizeStatus converts common LLM variations to canonical status values.
// Apply BEFORE switch statement comparison.
func normalizeStatus(status string) string {
	// 1. Lowercase and trim
	s := strings.ToLower(strings.TrimSpace(status))

	// 2. Normalize separators: spaces and underscores → hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// 3. Collapse multiple hyphens into one (handles "ready__for__dev" → "ready-for-dev")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	// 4. Map synonyms (G17)
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

// isDeferred checks if a normalized status indicates a deferred epic.
// G24: Deferred epics should be completely skipped in stage detection.
func isDeferred(normalizedStatus string) bool {
	return strings.Contains(normalizedStatus, "deferred") ||
		strings.HasPrefix(normalizedStatus, "post-mvp")
}

// activeStatuses defines normalized statuses that indicate active/in-progress state.
// G23: Used for retrospective detection when all epics are done.
var activeStatuses = map[string]bool{
	"in-progress": true, // Most common - normalizeStatus maps WIP/wip/in_progress here
	"started":     true, // Direct usage
}

// isActiveStatus checks if a normalized status indicates an active/in-progress state.
func isActiveStatus(normalizedStatus string) bool {
	return activeStatuses[normalizedStatus]
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

	// G23: Track retrospectives for all-done case
	type retroInfo struct {
		key     string // e.g., "epic-7-retrospective"
		epicNum string // e.g., "7" or "4-5" for sub-epics
		status  string // normalized status
	}

	epics := make(map[string]*epicInfo)
	deferredEpics := make(map[string]bool) // G24: Track deferred epics to skip their stories
	var epicOrder []string                 // Preserve order for finding first in-progress
	var retrospectives []retroInfo         // G23: Collect retrospectives for all-done check

	// First pass: collect epics (identify deferred first - G24)
	for key, value := range status.DevelopmentStatus {
		// Skip retrospectives in first pass - collect after we know all deferred epics
		if strings.HasSuffix(key, "-retrospective") {
			continue
		}

		if epicKeyRegex.MatchString(key) {
			normalized := normalizeStatus(value)

			// G24: Skip deferred epics from active counting, track to prevent orphan warnings for their stories
			if isDeferred(normalized) {
				deferredEpics[key] = true
				continue
			}

			epics[key] = &epicInfo{
				key:    key,
				status: normalized,
			}
			epicOrder = append(epicOrder, key)
		}
	}

	// G23: Collect retrospectives (after we know all deferred epics)
	for key, value := range status.DevelopmentStatus {
		if matches := retroKeyRegex.FindStringSubmatch(key); matches != nil {
			epicNum := matches[1] // e.g., "7" or "4-5"
			epicKey := "epic-" + epicNum

			// G24+G23: Skip retrospectives for deferred epics (AC9)
			if deferredEpics[epicKey] {
				continue
			}

			retrospectives = append(retrospectives, retroInfo{
				key:     key,
				epicNum: epicNum,
				status:  normalizeStatus(value),
			})
		}
	}

	// Sort epic order for deterministic behavior (map iteration is random)
	sort.Strings(epicOrder)

	// G24: Check if all epics are deferred (no active epics found)
	if len(epics) == 0 {
		return domain.StageUnknown, domain.ConfidenceUncertain,
			"All epics deferred - no active development"
	}

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

			// G24: Skip stories for deferred epics (no orphan warning)
			if deferredEpics[epicKey] {
				continue
			}

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

	// G23: Sort retrospectives by epicNum for deterministic behavior (AC5)
	sort.Slice(retrospectives, func(i, j int) bool {
		return retrospectives[i].epicNum < retrospectives[j].epicNum
	})

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

	// All epics done - G23: Check for in-progress retrospective
	if len(epics) > 0 && doneCount == len(epics) {
		for _, retro := range retrospectives {
			if isActiveStatus(retro.status) {
				return domain.StageImplement, domain.ConfidenceCertain,
					appendWarnings("Retrospective for Epic " + retro.epicNum + " in progress")
			}
		}
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

		// Track stories with unknown status for fallback display
		var unknownStatusStories []struct {
			key    string
			status string
		}

		for _, story := range sortedStories {
			if p, ok := storyPriority[story.status]; ok && p < selectedPriority {
				selectedStory = story.key
				selectedStatus = story.status
				selectedPriority = p
			} else if _, known := storyPriority[story.status]; !known && story.status != "" {
				// Track unknown statuses (not empty, not in priority map)
				unknownStatusStories = append(unknownStatusStories, struct {
					key    string
					status string
				}{key: story.key, status: story.status})
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
		case "backlog":
			return domain.StagePlan, domain.ConfidenceCertain,
				appendWarnings("Story " + formatStoryKey(selectedStory) + " in backlog, needs drafting")
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

		// Epic in-progress but no known-status stories started
		// If there are stories with unknown status, show the first one
		if len(unknownStatusStories) > 0 {
			first := unknownStatusStories[0]
			warnings = append(warnings, "unknown status '"+first.status+"' for "+first.key)
			return domain.StagePlan, domain.ConfidenceLikely,
				appendWarnings("Story " + formatStoryKey(first.key) + " has unknown status '" + first.status + "'")
		}

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
