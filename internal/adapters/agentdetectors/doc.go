// Package agentdetectors provides implementations of the AgentActivityDetector
// interface for detecting AI coding agent activity in projects.
//
// This package is part of the adapters layer in the hexagonal architecture.
// It contains infrastructure code that interacts with the filesystem to detect
// various AI agent states (working, waiting, idle).
//
// Implementations:
//   - ClaudeCodePathMatcher (Story 15.2): Matches project paths to Claude Code log directories
//   - ClaudeCodeLogParser (Story 15.3): Parses Claude Code JSONL logs with tail optimization
//   - ClaudeCodeDetector (Story 15.4): Main implementation of AgentActivityDetector for Claude Code.
//     Composes PathMatcher and LogParser to detect agent state from JSONL logs with high confidence.
//   - GenericDetector (Story 15.5): File activity fallback for any project. Scans filesystem for most
//     recent file modification time and returns Working/WaitingForUser with ConfidenceUncertain.
//     Used as fallback when tool-specific detectors (like ClaudeCodeDetector) don't match.
package agentdetectors
