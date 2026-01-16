# Story 15.2: Implement Claude Code Project Path Matcher

Status: done

## Story

As a developer,
I want to match a project path to its Claude Code log directory,
So that I can find the relevant JSONL logs for agent state detection.

## User-Visible Changes

None - this is internal infrastructure for the Claude Code agent detector. User-visible changes will come in Story 15.6 when agent detection is integrated into the TUI dashboard.

## Acceptance Criteria

1. **AC1:** Given project path `/Users/jongkuk/projects/vibe-dash`, matcher finds corresponding directory in `~/.claude/projects/` by path escaping pattern
2. **AC2:** Given Claude Code uses `-` character to escape `/` in directory names, matcher correctly converts project paths (e.g., `/Users/foo/bar` → `-Users-foo-bar`)
3. **AC3:** Given no matching Claude Code project directory exists, matcher returns empty result (no error)
4. **AC4:** Given Claude Code is not installed (`~/.claude/projects` doesn't exist), matcher returns empty result gracefully
5. **AC5:** PathMatcher has `Match(ctx context.Context, projectPath string) (string, error)` method returning matched Claude logs directory path
6. **AC6:** PathMatcher implements caching to avoid repeated filesystem scans on consecutive calls for same project

## Tasks / Subtasks

- [x] Task 0: Create agentdetectors package structure (prerequisite)
  - [x] 0.1: Create directory `internal/adapters/agentdetectors/`
  - [x] 0.2: Create `internal/adapters/agentdetectors/doc.go` with package comment

- [x] Task 1: Create ClaudeCodePathMatcher struct (AC: 1, 2, 5, 6)
  - [x] 1.1: Create `internal/adapters/agentdetectors/claude_code_path_matcher.go`
  - [x] 1.2: Define struct with cache and mutex fields (see Struct Definition below)
  - [x] 1.3: Implement `NewClaudeCodePathMatcher()` constructor
  - [x] 1.4: Implement `pathToClaudeDir(projectPath string) string` private method

- [x] Task 2: Implement path escaping and conversion (AC: 2)
  - [x] 2.1: Convert relative paths to absolute using `filepath.Abs()` before processing
  - [x] 2.2: Replace `/` with `-` in project path using `strings.ReplaceAll`
  - [x] 2.3: Join escaped path with `~/.claude/projects/` base directory
  - [x] 2.4: Handle edge cases: empty path returns empty string, `os.UserHomeDir()` error returns empty string

- [x] Task 3: Implement `Match` method with directory validation (AC: 3, 4, 5)
  - [x] 3.1: Check `ctx.Done()` before filesystem operations (respect cancellation within 100ms)
  - [x] 3.2: Check cache first (read lock) and return cached value if present
  - [x] 3.3: Check if `~/.claude/projects` base directory exists (return empty string if not)
  - [x] 3.4: Check if escaped project directory exists
  - [x] 3.5: Verify directory contains at least one `*.jsonl` file (excluding `agent-*.jsonl` for consistency with LogReader)
  - [x] 3.6: Return empty string if any check fails (graceful handling, no error)

- [x] Task 4: Implement caching mechanism (AC: 6)
  - [x] 4.1: Check cache with read lock before filesystem lookup
  - [x] 4.2: Store result with write lock after lookup (including empty string for "not found")
  - [x] 4.3: Implement `ClearCache()` method for testing and invalidation

- [x] Task 5: Write comprehensive unit tests (AC: 1-6)
  - [x] 5.1: Create `internal/adapters/agentdetectors/claude_code_path_matcher_test.go`
  - [x] 5.2: Test `pathToClaudeDir` for various paths (normal, root, spaces, relative)
  - [x] 5.3: Test cache hit/miss behavior
  - [x] 5.4: Test concurrent cache access (race condition prevention)
  - [x] 5.5: Test graceful handling when Claude Code not installed
  - [x] 5.6: Test empty directory (no JSONL files) returns empty string
  - [x] 5.7: Test context cancellation returns empty result promptly

- [x] Task 6: Verify integration
  - [x] 6.1: Ensure path escaping matches `logreaders/claude_code.go` behavior exactly
  - [x] 6.2: Run `make lint && make test` - all must pass

## Dev Notes

### Struct Definition (Complete)

```go
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
```

### Path Escaping Logic (Copy from LogReaders)

The path escaping logic exists in `internal/adapters/logreaders/claude_code.go:200-212`. **Copy this logic exactly** - do NOT import from logreaders to avoid circular dependencies:

```go
const claudeProjectsDir = ".claude/projects"

// pathToClaudeDir converts a project path to the Claude logs directory.
// Example: /Users/limjk/GitHub/JeiKeiLim/vibe-dash
//          → ~/.claude/projects/-Users-limjk-GitHub-JeiKeiLim-vibe-dash/
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
```

### Match Method Implementation

```go
// Match finds the Claude Code logs directory for a project path.
// Returns empty string (not error) if Claude Code not installed or no logs found.
func (m *ClaudeCodePathMatcher) Match(ctx context.Context, projectPath string) (string, error) {
    // Respect context cancellation
    select {
    case <-ctx.Done():
        return "", nil // Graceful: return empty, not error
    default:
    }

    // Check cache first (read lock)
    m.cacheMu.RLock()
    if cached, ok := m.cache[projectPath]; ok {
        m.cacheMu.RUnlock()
        return cached, nil
    }
    m.cacheMu.RUnlock()

    // Compute Claude directory path
    claudeDir := m.pathToClaudeDir(projectPath)
    if claudeDir == "" {
        m.cacheResult(projectPath, "")
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
        m.cacheResult(projectPath, "")
        return "", nil
    }

    // Verify directory has JSONL files (exclude agent-*.jsonl)
    hasLogs, err := m.hasJSONLFiles(claudeDir)
    if err != nil || !hasLogs {
        m.cacheResult(projectPath, "")
        return "", nil
    }

    m.cacheResult(projectPath, claudeDir)
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
```

### Edge Cases

| Case | Input | Behavior |
|------|-------|----------|
| Normal absolute path | `/Users/foo/bar` | Escape and check `~/.claude/projects/-Users-foo-bar` |
| Relative path | `./project` | Convert to absolute first via `filepath.Abs()` |
| Empty path | `` | Return empty string immediately |
| Path with spaces | `/Users/foo/my project` | Escape to `-Users-foo-my project` (spaces preserved) |
| Claude not installed | Any | Return empty string (no error) |
| Dir exists but no logs | Valid dir, no *.jsonl | Return empty string |
| Home dir error | `os.UserHomeDir()` fails | Return empty string |
| Context cancelled | Any | Return empty string promptly |

### Testing Strategy

```go
func TestPathToClaudeDir(t *testing.T) {
    tests := []struct {
        name        string
        projectPath string
        wantSuffix  string // relative to home dir, empty if expect empty result
    }{
        {"normal path", "/Users/foo/bar", ".claude/projects/-Users-foo-bar"},
        {"root path", "/", ".claude/projects/-"},
        {"path with spaces", "/Users/foo/my project", ".claude/projects/-Users-foo-my project"},
        {"empty path", "", ""}, // Should return empty
    }
    // Table-driven test implementation
}

func TestMatch_CacheHit(t *testing.T) {
    // Create temp dir with JSONL file
    // First call: filesystem lookup
    // Second call: verify no filesystem access (mock or track calls)
}

func TestMatch_ConcurrentAccess(t *testing.T) {
    // Spawn multiple goroutines calling Match simultaneously
    // Use -race flag to detect race conditions
    // All should complete without panic
}

func TestMatch_ContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately
    result, err := matcher.Match(ctx, "/some/path")
    // Should return empty string, nil error, within 100ms
}

func TestMatch_ClaudeNotInstalled(t *testing.T) {
    // Use t.TempDir() as fake home, no .claude directory
    // Expect empty result, no error
}

func TestMatch_NoJSONLFiles(t *testing.T) {
    // Create .claude/projects/-path- directory but no *.jsonl files
    // Expect empty result
}
```

### Hexagonal Architecture

```
internal/adapters/agentdetectors/
├── doc.go                           # NEW: Package documentation
├── claude_code_path_matcher.go      # NEW: Path matching logic
└── claude_code_path_matcher_test.go # NEW: Tests

This is an ADAPTER (not core/ports), so:
- Can use os, filepath, sync, strings packages
- Can access filesystem directly
- Does NOT implement AgentActivityDetector interface (that's Story 15.3)
- Used as helper BY ClaudeCodeDetector
```

### Integration with Future Stories

Story 15.3/15.4 will create `ClaudeCodeDetector` that uses this PathMatcher:

```go
// In story 15.3:
type ClaudeCodeDetector struct {
    pathMatcher *ClaudeCodePathMatcher
}

func (d *ClaudeCodeDetector) Detect(ctx context.Context, projectPath string) (domain.AgentState, error) {
    // Use PathMatcher to find logs directory
    claudeDir, err := d.pathMatcher.Match(ctx, projectPath)
    if err != nil || claudeDir == "" {
        return domain.NewAgentState("Claude Code", domain.AgentUnknown, 0, domain.ConfidenceUncertain), nil
    }
    // Parse JSONL logs in claudeDir to determine agent state...
}
```

### References

- [Source: docs/prd-phase2.md] - FR-P2-6 requirement
- [Source: docs/epics-phase2.md#Story-3.2] - Story specification
- [Source: internal/adapters/logreaders/claude_code.go:200-212] - Existing path escaping logic
- [Source: internal/adapters/logreaders/claude_code.go:84-86] - Agent file skip pattern
- [Source: docs/sprint-artifacts/stories/epic-15/15-1-define-agentactivitydetector-interface-and-types.md] - Previous story (interface patterns)

## Dev Agent Record

### Context Reference

- Phase 2 Epic 15: Sub-1-Minute Agent Detection
- FR Coverage: FR-P2-6 (Match project path to Claude Code logs)
- Prerequisite: Story 15.1 (AgentActivityDetector interface) - DONE

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

1. Created `internal/adapters/agentdetectors/` package with doc.go
2. Implemented `ClaudeCodePathMatcher` struct with thread-safe cache using `sync.RWMutex`
3. Path escaping matches `logreaders/claude_code.go` exactly: `/` → `-` transformation
4. Enhanced with `filepath.Abs()` to handle relative paths (improvement over logreaders version)
5. All 12 unit tests pass with race detection enabled
6. Full test suite: 1206 tests pass, lint passes

### Code Review Fixes Applied

1. **M1 - Unrelated change reverted**: Removed unrelated `jq -C` color flag change from `internal/adapters/tui/model.go`
2. **M3 - Cache normalization**: Added `normalizePath()` function that uses `filepath.EvalSymlinks` to ensure relative paths (`./project`) and absolute paths (`/full/path/project`) share the same cache entry. This prevents duplicate cache entries on macOS where `/var` is a symlink to `/private/var`.
3. **L1 - doc.go clarity**: Updated "Current implementations:" to "Implementations:" and added "(planned)" suffix to future stories
4. **L2 - Test coverage**: Added `TestMatch_CacheUsesNormalizedPaths` test to verify cache key normalization

### File List

- `internal/adapters/agentdetectors/doc.go` - NEW: Package documentation
- `internal/adapters/agentdetectors/claude_code_path_matcher.go` - NEW: ClaudeCodePathMatcher implementation (with normalizePath helper)
- `internal/adapters/agentdetectors/claude_code_path_matcher_test.go` - NEW: Comprehensive tests (13 test functions)
