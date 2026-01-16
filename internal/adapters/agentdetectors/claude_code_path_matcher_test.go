package agentdetectors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPathToClaudeDir(t *testing.T) {
	m := NewClaudeCodePathMatcher()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name        string
		projectPath string
		wantSuffix  string // relative to home dir, empty if expect empty result
	}{
		{
			name:        "normal path",
			projectPath: "/Users/foo/bar",
			wantSuffix:  ".claude/projects/-Users-foo-bar",
		},
		{
			name:        "root path",
			projectPath: "/",
			wantSuffix:  ".claude/projects/-",
		},
		{
			name:        "path with spaces",
			projectPath: "/Users/foo/my project",
			wantSuffix:  ".claude/projects/-Users-foo-my project",
		},
		{
			name:        "empty path",
			projectPath: "",
			wantSuffix:  "", // Should return empty
		},
		{
			name:        "deeper nested path",
			projectPath: "/Users/limjk/GitHub/JeiKeiLim/vibe-dash",
			wantSuffix:  ".claude/projects/-Users-limjk-GitHub-JeiKeiLim-vibe-dash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.pathToClaudeDir(tt.projectPath)
			if tt.wantSuffix == "" {
				if got != "" {
					t.Errorf("pathToClaudeDir(%q) = %q, want empty", tt.projectPath, got)
				}
				return
			}
			want := filepath.Join(homeDir, tt.wantSuffix)
			if got != want {
				t.Errorf("pathToClaudeDir(%q) = %q, want %q", tt.projectPath, got, want)
			}
		})
	}
}

func TestPathToClaudeDir_RelativePath(t *testing.T) {
	m := NewClaudeCodePathMatcher()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Create a temp dir and test relative path conversion
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Test relative path "./subdir"
	subdir := "subdir"
	err = os.Mkdir(subdir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	got := m.pathToClaudeDir("./" + subdir)

	// The result should be an absolute path based on tmpDir
	if got == "" {
		t.Errorf("pathToClaudeDir(relative) returned empty, expected non-empty")
	}
	// Verify it contains the escaped absolute path using strings.HasPrefix
	wantPrefix := filepath.Join(homeDir, ".claude/projects/")
	if !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("pathToClaudeDir(relative) = %q, should start with %q", got, wantPrefix)
	}
}

func TestMatch_DirectoryNotExists(t *testing.T) {
	m := NewClaudeCodePathMatcher()
	ctx := context.Background()

	// Use a path that definitely won't exist in ~/.claude/projects/
	result, err := m.Match(ctx, "/nonexistent/project/path/12345")
	if err != nil {
		t.Errorf("Match() returned error: %v", err)
	}
	if result != "" {
		t.Errorf("Match() = %q, want empty string for nonexistent directory", result)
	}
}

func TestMatch_ClaudeNotInstalled(t *testing.T) {
	// This test verifies graceful handling when ~/.claude doesn't exist
	// We can't easily mock os.UserHomeDir, so we test the behavior indirectly
	m := NewClaudeCodePathMatcher()
	ctx := context.Background()

	// Any path should return empty when the directory doesn't exist
	result, err := m.Match(ctx, "/some/random/path")
	if err != nil {
		t.Errorf("Match() should not return error: %v", err)
	}
	// Result should be empty (graceful failure)
	// This is a valid test since /some/random/path won't have a matching Claude dir
	if result != "" {
		t.Logf("Match() = %q (may be valid if you have Claude Code logs for this path)", result)
	}
}

func TestMatch_CacheHit(t *testing.T) {
	m := NewClaudeCodePathMatcher()
	ctx := context.Background()

	// Create a temp directory structure that mimics Claude Code
	tmpHome := t.TempDir()
	projectPath := "/test/project"
	escapedPath := "-test-project"
	claudeDir := filepath.Join(tmpHome, claudeProjectsDir, escapedPath)

	err := os.MkdirAll(claudeDir, 0755)
	if err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}

	// Create a JSONL file
	jsonlFile := filepath.Join(claudeDir, "session.jsonl")
	err = os.WriteFile(jsonlFile, []byte(`{"type":"test"}`), 0644)
	if err != nil {
		t.Fatalf("failed to create jsonl file: %v", err)
	}

	// Manually prime the cache with a known value
	m.cacheMu.Lock()
	m.cache[projectPath] = claudeDir
	m.cacheMu.Unlock()

	// Call Match - should return cached value without filesystem check
	result, err := m.Match(ctx, projectPath)
	if err != nil {
		t.Errorf("Match() error: %v", err)
	}
	if result != claudeDir {
		t.Errorf("Match() = %q, want cached value %q", result, claudeDir)
	}

	// Verify cache still contains the value
	m.cacheMu.RLock()
	cached, ok := m.cache[projectPath]
	m.cacheMu.RUnlock()
	if !ok || cached != claudeDir {
		t.Errorf("cache should contain %q = %q", projectPath, claudeDir)
	}
}

func TestMatch_CacheMiss(t *testing.T) {
	m := NewClaudeCodePathMatcher()
	ctx := context.Background()

	// Ensure cache is empty
	m.cacheMu.RLock()
	_, ok := m.cache["/uncached/path"]
	m.cacheMu.RUnlock()
	if ok {
		t.Fatal("cache should not contain /uncached/path initially")
	}

	// Call Match
	_, err := m.Match(ctx, "/uncached/path")
	if err != nil {
		t.Errorf("Match() error: %v", err)
	}

	// After Match, the result should be cached (even if empty)
	m.cacheMu.RLock()
	_, ok = m.cache["/uncached/path"]
	m.cacheMu.RUnlock()
	if !ok {
		t.Error("cache should contain /uncached/path after Match()")
	}
}

func TestMatch_ConcurrentAccess(t *testing.T) {
	m := NewClaudeCodePathMatcher()
	ctx := context.Background()

	// Spawn multiple goroutines calling Match simultaneously
	var wg sync.WaitGroup
	paths := []string{
		"/path/one",
		"/path/two",
		"/path/three",
		"/path/four",
		"/path/five",
	}

	// Run 10 iterations to increase chance of detecting race conditions
	for i := 0; i < 10; i++ {
		for _, p := range paths {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				_, err := m.Match(ctx, path)
				if err != nil {
					t.Errorf("Match(%q) error: %v", path, err)
				}
			}(p)
		}
	}

	wg.Wait()
	// If we get here without panic or race detector errors, the test passes
}

func TestMatch_ContextCancellation(t *testing.T) {
	m := NewClaudeCodePathMatcher()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	start := time.Now()
	result, err := m.Match(ctx, "/some/path")
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Match() should not return error on cancellation: %v", err)
	}
	if result != "" {
		t.Errorf("Match() = %q, want empty string on cancellation", result)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("Match() took %v, should return within 100ms on cancellation", elapsed)
	}
}

func TestMatch_NoJSONLFiles(t *testing.T) {
	m := NewClaudeCodePathMatcher()

	// We can't easily override home dir, so test hasJSONLFiles directly
	emptyDir := t.TempDir()
	hasFiles, err := m.hasJSONLFiles(emptyDir)
	if err != nil {
		t.Errorf("hasJSONLFiles() error: %v", err)
	}
	if hasFiles {
		t.Error("hasJSONLFiles() should return false for empty directory")
	}

	// Test with only agent-*.jsonl files (should be skipped)
	agentFile := filepath.Join(emptyDir, "agent-123.jsonl")
	err = os.WriteFile(agentFile, []byte(`{}`), 0644)
	if err != nil {
		t.Fatalf("failed to create agent file: %v", err)
	}

	hasFiles, err = m.hasJSONLFiles(emptyDir)
	if err != nil {
		t.Errorf("hasJSONLFiles() error: %v", err)
	}
	if hasFiles {
		t.Error("hasJSONLFiles() should return false when only agent-*.jsonl exists")
	}

	// Add a valid JSONL file
	validFile := filepath.Join(emptyDir, "session.jsonl")
	err = os.WriteFile(validFile, []byte(`{}`), 0644)
	if err != nil {
		t.Fatalf("failed to create valid file: %v", err)
	}

	hasFiles, err = m.hasJSONLFiles(emptyDir)
	if err != nil {
		t.Errorf("hasJSONLFiles() error: %v", err)
	}
	if !hasFiles {
		t.Error("hasJSONLFiles() should return true when valid JSONL exists")
	}
}

func TestClearCache(t *testing.T) {
	m := NewClaudeCodePathMatcher()

	// Add some entries to cache
	m.cacheMu.Lock()
	m.cache["/path/one"] = "/claude/one"
	m.cache["/path/two"] = "/claude/two"
	m.cacheMu.Unlock()

	// Verify cache has entries
	m.cacheMu.RLock()
	count := len(m.cache)
	m.cacheMu.RUnlock()
	if count != 2 {
		t.Fatalf("cache should have 2 entries, got %d", count)
	}

	// Clear cache
	m.ClearCache()

	// Verify cache is empty
	m.cacheMu.RLock()
	count = len(m.cache)
	m.cacheMu.RUnlock()
	if count != 0 {
		t.Errorf("cache should be empty after ClearCache(), got %d entries", count)
	}
}

func TestMatch_EmptyPath(t *testing.T) {
	m := NewClaudeCodePathMatcher()
	ctx := context.Background()

	result, err := m.Match(ctx, "")
	if err != nil {
		t.Errorf("Match(\"\") should not return error: %v", err)
	}
	if result != "" {
		t.Errorf("Match(\"\") = %q, want empty string", result)
	}
}

func TestNewClaudeCodePathMatcher(t *testing.T) {
	m := NewClaudeCodePathMatcher()

	if m == nil {
		t.Fatal("NewClaudeCodePathMatcher() returned nil")
	}
	if m.cache == nil {
		t.Error("cache should be initialized")
	}
	if len(m.cache) != 0 {
		t.Error("cache should be empty initially")
	}
}

func TestMatch_CacheUsesNormalizedPaths(t *testing.T) {
	m := NewClaudeCodePathMatcher()
	ctx := context.Background()

	// Create a temp directory to test relative vs absolute path caching
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Create a subdirectory
	subdir := "testproject"
	err = os.Mkdir(subdir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	relativePath := "./" + subdir
	absolutePath := filepath.Join(tmpDir, subdir)

	// Normalize absolutePath to handle symlinks (e.g., /var -> /private/var on macOS)
	normalizedAbsPath := normalizePath(absolutePath)

	// Call Match with relative path first
	_, err = m.Match(ctx, relativePath)
	if err != nil {
		t.Errorf("Match(relative) error: %v", err)
	}

	// Check cache has normalized path (symlinks resolved)
	m.cacheMu.RLock()
	_, hasNormalized := m.cache[normalizedAbsPath]
	_, hasRelative := m.cache[relativePath]
	cacheLen := len(m.cache)
	m.cacheMu.RUnlock()

	if !hasNormalized {
		t.Errorf("cache should have normalized path as key, expected %q", normalizedAbsPath)
	}
	if hasRelative {
		t.Error("cache should NOT have relative path as key")
	}
	if cacheLen != 1 {
		t.Errorf("cache should have exactly 1 entry, got %d", cacheLen)
	}

	// Call Match with absolute path - should hit cache, not create new entry
	_, err = m.Match(ctx, absolutePath)
	if err != nil {
		t.Errorf("Match(absolute) error: %v", err)
	}

	m.cacheMu.RLock()
	cacheLen = len(m.cache)
	m.cacheMu.RUnlock()

	if cacheLen != 1 {
		t.Errorf("cache should still have exactly 1 entry after absolute path call, got %d", cacheLen)
	}
}
