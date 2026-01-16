package agentdetectors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const claudeProjectsDir = ".claude/projects"

// ClaudeCodePathMatcher matches project paths to Claude Code log directories.
// This is a HELPER struct used by ClaudeCodeDetector (Story 15.3), NOT an
// implementation of AgentActivityDetector interface.
type ClaudeCodePathMatcher struct {
	cache   map[string]string // projectPath → claudeDir (empty string = not found)
	cacheMu sync.RWMutex      // Thread-safe cache access
}

// NewClaudeCodePathMatcher creates a new path matcher with empty cache.
func NewClaudeCodePathMatcher() *ClaudeCodePathMatcher {
	return &ClaudeCodePathMatcher{
		cache: make(map[string]string),
	}
}

// pathToClaudeDir converts a project path to the Claude logs directory.
// Example: /Users/limjk/GitHub/JeiKeiLim/vibe-dash
//
//	→ ~/.claude/projects/-Users-limjk-GitHub-JeiKeiLim-vibe-dash/
func (m *ClaudeCodePathMatcher) pathToClaudeDir(projectPath string) string {
	// Handle empty path
	if projectPath == "" {
		return ""
	}

	// Convert relative to absolute
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return ""
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "" // Return empty instead of invalid "~" path
	}

	// Replace / with -
	escapedPath := strings.ReplaceAll(absPath, "/", "-")
	return filepath.Join(homeDir, claudeProjectsDir, escapedPath)
}

// normalizePath returns a canonical absolute path for cache key consistency.
// It resolves relative paths and symlinks to ensure equivalent paths
// (e.g., ./project and /full/path/project) use the same cache key.
func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	// Resolve symlinks for consistent cache keys (e.g., /var vs /private/var on macOS)
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		return resolved
	}
	return absPath
}

// Match finds the Claude Code logs directory for a project path.
// Returns empty string (not error) if Claude Code not installed or no logs found.
func (m *ClaudeCodePathMatcher) Match(ctx context.Context, projectPath string) (string, error) {
	// Respect context cancellation
	select {
	case <-ctx.Done():
		return "", nil // Graceful: return empty, not error
	default:
	}

	// Normalize to canonical path for consistent cache keys
	cacheKey := normalizePath(projectPath)

	// Check cache first (read lock)
	m.cacheMu.RLock()
	if cached, ok := m.cache[cacheKey]; ok {
		m.cacheMu.RUnlock()
		return cached, nil
	}
	m.cacheMu.RUnlock()

	// Compute Claude directory path
	claudeDir := m.pathToClaudeDir(projectPath)
	if claudeDir == "" {
		m.cacheResult(cacheKey, "")
		return "", nil
	}

	// Check if directory exists
	select {
	case <-ctx.Done():
		return "", nil
	default:
	}

	info, err := os.Stat(claudeDir)
	if err != nil || !info.IsDir() {
		m.cacheResult(cacheKey, "")
		return "", nil
	}

	// Verify directory has JSONL files (exclude agent-*.jsonl)
	hasLogs, err := m.hasJSONLFiles(claudeDir)
	if err != nil || !hasLogs {
		m.cacheResult(cacheKey, "")
		return "", nil
	}

	m.cacheResult(cacheKey, claudeDir)
	return claudeDir, nil
}

func (m *ClaudeCodePathMatcher) cacheResult(projectPath, result string) {
	m.cacheMu.Lock()
	m.cache[projectPath] = result
	m.cacheMu.Unlock()
}

func (m *ClaudeCodePathMatcher) hasJSONLFiles(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match LogReader behavior: skip agent-*.jsonl sub-sessions
		if strings.HasSuffix(name, ".jsonl") && !strings.HasPrefix(name, "agent-") {
			return true, nil
		}
	}
	return false, nil
}

// ClearCache clears the cached path lookups. Used for testing.
func (m *ClaudeCodePathMatcher) ClearCache() {
	m.cacheMu.Lock()
	m.cache = make(map[string]string)
	m.cacheMu.Unlock()
}
