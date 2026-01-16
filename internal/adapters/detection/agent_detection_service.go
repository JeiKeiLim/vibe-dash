package detection

import (
	"context"
	"log/slog"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/agentdetectors"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

const serviceName = "Agent Detection Service"

// detectionTimeout is the maximum time for a single detection.
// Per NFR-P2-1: Agent detection latency < 1 second.
const detectionTimeout = 1 * time.Second

// AgentDetectionService orchestrates multiple agent detectors with fallback.
// It tries Claude Code detection first (high confidence), then falls back
// to generic file-activity detection (low confidence).
//
// Located in adapters layer (not services) because it directly composes
// adapter-layer detectors. Per hexagonal architecture, core/services
// MUST NOT import adapters.
type AgentDetectionService struct {
	claudeDetector  ports.AgentActivityDetector
	genericDetector ports.AgentActivityDetector
}

// ServiceOption configures AgentDetectionService.
type ServiceOption func(*AgentDetectionService)

// WithClaudeDetector sets a custom Claude detector (for testing).
func WithClaudeDetector(d ports.AgentActivityDetector) ServiceOption {
	return func(s *AgentDetectionService) {
		s.claudeDetector = d
	}
}

// WithGenericDetector sets a custom generic detector (for testing).
func WithGenericDetector(d ports.AgentActivityDetector) ServiceOption {
	return func(s *AgentDetectionService) {
		s.genericDetector = d
	}
}

// NewAgentDetectionService creates a new service with optional configuration.
func NewAgentDetectionService(opts ...ServiceOption) *AgentDetectionService {
	s := &AgentDetectionService{}
	for _, opt := range opts {
		opt(s)
	}
	if s.claudeDetector == nil {
		s.claudeDetector = agentdetectors.NewClaudeCodeDetector()
	}
	if s.genericDetector == nil {
		s.genericDetector = agentdetectors.NewGenericDetector()
	}
	return s
}

// Compile-time interface compliance check.
var _ ports.AgentActivityDetector = (*AgentDetectionService)(nil)

// Name returns the service identifier.
func (s *AgentDetectionService) Name() string {
	return serviceName
}

// Detect determines the agent activity state for a project.
// Uses Claude Code detection with fallback to generic file-activity detection.
func (s *AgentDetectionService) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
	// Apply timeout per NFR-P2-1 (< 1 second)
	ctx, cancel := context.WithTimeout(ctx, detectionTimeout)
	defer cancel()

	// Respect context cancellation at entry
	select {
	case <-ctx.Done():
		return domain.NewAgentState(serviceName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
	default:
	}

	// Step 1: Try Claude Code detection (high confidence)
	claudeState, err := s.claudeDetector.Detect(ctx, projectPath)
	if err != nil {
		slog.Debug("Claude Code detection error, falling back to generic",
			"path", projectPath, "error", err)
		// Fall through to generic detection
	} else if !claudeState.IsUnknown() {
		// Claude detection succeeded with known state
		slog.Debug("Agent detection via Claude Code",
			"path", projectPath, "status", claudeState.Status.String())
		return claudeState, nil
	}

	// Check context between detector calls (Story 15.4 learning)
	select {
	case <-ctx.Done():
		return domain.NewAgentState(serviceName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
	default:
	}

	// Step 2: Fall back to generic file-activity detection (low confidence)
	genericState, err := s.genericDetector.Detect(ctx, projectPath)
	if err != nil {
		slog.Debug("Generic detection error",
			"path", projectPath, "error", err)
		return domain.NewAgentState(serviceName, domain.AgentUnknown, 0, domain.ConfidenceUncertain), err
	}

	slog.Debug("Agent detection via generic file activity",
		"path", projectPath, "status", genericState.Status.String())
	return genericState, nil
}
