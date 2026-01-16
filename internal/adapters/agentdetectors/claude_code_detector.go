package agentdetectors

import (
	"context"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

const detectorName = "Claude Code"

// ClaudeCodeDetector detects agent activity state by parsing Claude Code JSONL logs.
// Implements ports.AgentActivityDetector interface.
type ClaudeCodeDetector struct {
	pathMatcher *ClaudeCodePathMatcher
	logParser   *ClaudeCodeLogParser
}

// DetectorOption is a functional option for configuring ClaudeCodeDetector.
type DetectorOption func(*ClaudeCodeDetector)

// WithPathMatcher sets a custom path matcher (for testing).
func WithPathMatcher(pm *ClaudeCodePathMatcher) DetectorOption {
	return func(d *ClaudeCodeDetector) {
		d.pathMatcher = pm
	}
}

// WithLogParser sets a custom log parser (for testing).
func WithLogParser(lp *ClaudeCodeLogParser) DetectorOption {
	return func(d *ClaudeCodeDetector) {
		d.logParser = lp
	}
}

// NewClaudeCodeDetector creates a new detector with optional configuration.
func NewClaudeCodeDetector(opts ...DetectorOption) *ClaudeCodeDetector {
	d := &ClaudeCodeDetector{}
	for _, opt := range opts {
		opt(d)
	}
	if d.pathMatcher == nil {
		d.pathMatcher = NewClaudeCodePathMatcher()
	}
	if d.logParser == nil {
		d.logParser = NewClaudeCodeLogParser()
	}
	return d
}

// Compile-time interface compliance check.
var _ ports.AgentActivityDetector = (*ClaudeCodeDetector)(nil)

// Name returns the detector identifier.
func (d *ClaudeCodeDetector) Name() string {
	return detectorName
}

// Detect determines the current agent activity state for a project.
func (d *ClaudeCodeDetector) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
	// Respect context cancellation at entry
	select {
	case <-ctx.Done():
		return domain.NewAgentState(detectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
	default:
	}

	// Step 1: Find Claude logs directory
	claudeDir, err := d.pathMatcher.Match(ctx, projectPath)
	if err != nil {
		// Unexpected error (permissions, etc.) - propagate
		return domain.NewAgentState(detectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), err
	}
	if claudeDir == "" {
		// Claude Code not installed or no logs for this project - graceful
		return domain.NewAgentState(detectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
	}

	// Check context between steps (AC7: respect cancellation)
	select {
	case <-ctx.Done():
		return domain.NewAgentState(detectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
	default:
	}

	// Step 2: Find most recent session
	sessionPath, err := d.logParser.FindMostRecentSession(ctx, claudeDir)
	if err != nil {
		return domain.NewAgentState(detectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), err
	}
	if sessionPath == "" {
		// Logs directory exists but no session files - inactive
		return domain.NewAgentState(detectorName, domain.AgentInactive, 0, domain.ConfidenceCertain), nil
	}

	// Check context between steps (AC7: respect cancellation)
	select {
	case <-ctx.Done():
		return domain.NewAgentState(detectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
	default:
	}

	// Step 3: Parse last assistant entry
	entry, err := d.logParser.ParseLastAssistantEntry(ctx, sessionPath)
	if err != nil {
		return domain.NewAgentState(detectorName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), err
	}
	if entry == nil {
		// Session exists but no assistant entries - inactive
		return domain.NewAgentState(detectorName, domain.AgentInactive, 0, domain.ConfidenceCertain), nil
	}

	// Step 4: Determine state from entry
	return d.determineState(entry), nil
}

// determineState interprets the stop_reason from a ClaudeLogEntry.
func (d *ClaudeCodeDetector) determineState(entry *ClaudeLogEntry) domain.AgentState {
	duration := time.Since(entry.Timestamp)

	// Handle zero timestamp (parsing failed) or future timestamp (clock skew)
	if entry.Timestamp.IsZero() || duration < 0 {
		duration = 0
	}

	switch {
	case entry.IsEndTurn():
		return domain.NewAgentState(detectorName, domain.AgentWaitingForUser, duration, domain.ConfidenceCertain)
	case entry.IsToolUse():
		return domain.NewAgentState(detectorName, domain.AgentWorking, duration, domain.ConfidenceCertain)
	default:
		return domain.NewAgentState(detectorName, domain.AgentUnknown, duration, domain.ConfidenceUncertain)
	}
}
