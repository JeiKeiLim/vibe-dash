package filesystem

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// mockPathLookup implements ports.ProjectPathLookup for testing.
type mockPathLookup struct {
	paths map[string]string // canonical path → directory name
}

func (m *mockPathLookup) GetDirForPath(canonicalPath string) string {
	if m.paths == nil {
		return ""
	}
	return m.paths[canonicalPath]
}

// Subtask 4.1: Test basic directory name derivation
func TestGetProjectDirName_BasicDerivation(t *testing.T) {
	tempDir := t.TempDir()

	// Create test project directory
	projectPath := filepath.Join(tempDir, "api-service")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(filepath.Join(tempDir, "vibe-dash"), &mockPathLookup{})

	got, err := dm.GetProjectDirName(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "api-service"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

// Subtask 4.2: Test first collision (parent disambiguation)
func TestGetProjectDirName_FirstCollision(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create first project
	project1 := filepath.Join(tempDir, "user", "api-service")
	if err := os.MkdirAll(project1, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	// Create second project with same name in different location
	project2 := filepath.Join(tempDir, "client-b", "api-service")
	if err := os.MkdirAll(project2, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	// Add first project
	dir1, err := dm.EnsureProjectDir(context.Background(), project1)
	if err != nil {
		t.Fatalf("failed to create first project dir: %v", err)
	}
	if filepath.Base(dir1) != "api-service" {
		t.Errorf("first project: got %q, want %q", filepath.Base(dir1), "api-service")
	}

	// Add second project - should get parent disambiguation
	dir2, err := dm.GetProjectDirName(context.Background(), project2)
	if err != nil {
		t.Fatalf("failed to get second project dir name: %v", err)
	}
	expected := "client-b-api-service"
	if dir2 != expected {
		t.Errorf("second project: got %q, want %q", dir2, expected)
	}
}

// Subtask 4.3: Test second collision (grandparent)
func TestGetProjectDirName_SecondCollision(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create three projects with same name
	project1 := filepath.Join(tempDir, "user", "api-service")
	project2 := filepath.Join(tempDir, "work", "api-service")
	project3 := filepath.Join(tempDir, "other", "work", "api-service")

	for _, p := range []string{project1, project2, project3} {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatalf("failed to create test dir: %v", err)
		}
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// First project: api-service
	_, err := dm.EnsureProjectDir(ctx, project1)
	if err != nil {
		t.Fatalf("failed to create first project dir: %v", err)
	}

	// Second project: work-api-service
	_, err = dm.EnsureProjectDir(ctx, project2)
	if err != nil {
		t.Fatalf("failed to create second project dir: %v", err)
	}

	// Third project should get grandparent: other-work-api-service
	dir3, err := dm.GetProjectDirName(ctx, project3)
	if err != nil {
		t.Fatalf("failed to get third project dir name: %v", err)
	}
	expected := "other-work-api-service"
	if dir3 != expected {
		t.Errorf("third project: got %q, want %q", dir3, expected)
	}
}

// Subtask 4.4: Test deeply nested paths (5+ levels)
func TestGetProjectDirName_DeeplyNested(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create deeply nested project
	projectPath := filepath.Join(tempDir, "a", "b", "c", "d", "e", "project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	got, err := dm.GetProjectDirName(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "project"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

// Subtask 4.5: Test special characters in path
func TestGetProjectDirName_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		dirName  string
		expected string
	}{
		{
			name:     "spaces",
			dirName:  "my project",
			expected: "my-project",
		},
		{
			name:     "uppercase",
			dirName:  "MyProject",
			expected: "myproject",
		},
		{
			name:     "colons",
			dirName:  "api:service",
			expected: "api-service",
		},
		{
			name:     "multiple special chars",
			dirName:  "my--cool__project",
			expected: "my-cool-project",
		},
		{
			name:     "leading trailing special",
			dirName:  "-project-",
			expected: "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			basePath := filepath.Join(tempDir, "vibe-dash")

			projectPath := filepath.Join(tempDir, tt.dirName)
			if err := os.MkdirAll(projectPath, 0755); err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			dm := NewDirectoryManager(basePath, &mockPathLookup{})

			got, err := dm.GetProjectDirName(context.Background(), projectPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

// Subtask 4.6: Test determinism (same path = same result)
func TestGetProjectDirName_Determinism(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	projectPath := filepath.Join(tempDir, "api-service")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// First call
	dir1, err := dm.EnsureProjectDir(ctx, projectPath)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call with same path - should return same directory
	dir2, err := dm.GetProjectDirName(ctx, projectPath)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if filepath.Base(dir1) != dir2 {
		t.Errorf("not deterministic: first=%q, second=%q", filepath.Base(dir1), dir2)
	}
}

// Subtask 4.6 (continued): Test determinism via lookup
func TestGetProjectDirName_DeterminismViaLookup(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	projectPath := filepath.Join(tempDir, "api-service")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	// Pre-populate lookup with existing mapping
	canonicalPath, _ := CanonicalPath(projectPath)
	lookup := &mockPathLookup{
		paths: map[string]string{
			canonicalPath: "existing-dir-name",
		},
	}

	dm := NewDirectoryManager(basePath, lookup)

	got, err := dm.GetProjectDirName(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "existing-dir-name"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

// Subtask 4.7: Test error handling (invalid path)
func TestGetProjectDirName_InvalidPath(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	_, err := dm.GetProjectDirName(context.Background(), "/nonexistent/path/project")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}

	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("expected ErrPathNotAccessible, got: %v", err)
	}
}

// Subtask 4.7 (continued): Test error handling (empty path)
func TestGetProjectDirName_EmptyPath(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	_, err := dm.GetProjectDirName(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}

	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("expected ErrPathNotAccessible, got: %v", err)
	}
}

// Subtask 4.8: Test symlink resolution
func TestGetProjectDirName_SymlinkResolution(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create real directory
	realPath := filepath.Join(tempDir, "real-project")
	if err := os.MkdirAll(realPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	// Create symlink
	symlinkPath := filepath.Join(tempDir, "symlink-project")
	if err := os.Symlink(realPath, symlinkPath); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// Both paths should resolve to same canonical path
	dir1, err := dm.GetProjectDirName(ctx, realPath)
	if err != nil {
		t.Fatalf("real path failed: %v", err)
	}

	dir2, err := dm.GetProjectDirName(ctx, symlinkPath)
	if err != nil {
		t.Fatalf("symlink path failed: %v", err)
	}

	if dir1 != dir2 {
		t.Errorf("symlink not resolved: real=%q, symlink=%q", dir1, dir2)
	}
}

// Subtask 4.9: Test max recursion depth exceeded
func TestGetProjectDirName_MaxDepthExceeded(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create 11 projects that will all collide at every depth level
	// Structure: /same/same/same/same/same/same/same/same/same/same/project (11 same levels)
	// Each project at /tempDir/pX/same/same/.../project
	// After normalization they all become same-same-...-project
	projects := make([]string, 11)
	for i := 0; i < 11; i++ {
		// Create paths that all share the same parent chain after the unique root
		// /tempDir/unique0/same/same/same/same/same/same/same/same/same/same/project
		path := filepath.Join(tempDir, "unique"+string(rune('0'+i)))
		for j := 0; j < 10; j++ {
			path = filepath.Join(path, "same")
		}
		projects[i] = filepath.Join(path, "project")
		if err := os.MkdirAll(projects[i], 0755); err != nil {
			t.Fatalf("failed to create test dir: %v", err)
		}
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// Add first 10 projects - should succeed
	// These will get names: project, same-project, same-same-project, etc.
	for i := 0; i < 10; i++ {
		_, err := dm.EnsureProjectDir(ctx, projects[i])
		if err != nil {
			t.Fatalf("project %d failed: %v", i, err)
		}
	}

	// 11th project should fail with collision unresolvable
	// because all 10 depth levels are exhausted (same-same-same-...-project)
	_, err := dm.GetProjectDirName(ctx, projects[10])
	if err == nil {
		t.Fatal("expected error for max depth exceeded")
	}

	if !errors.Is(err, domain.ErrCollisionUnresolvable) {
		t.Errorf("expected ErrCollisionUnresolvable, got: %v", err)
	}
}

// Subtask 4.10: Test case sensitivity
func TestGetProjectDirName_CaseSensitivity(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create projects with different cases
	project1 := filepath.Join(tempDir, "Api-Service")
	project2 := filepath.Join(tempDir, "other", "api-service")

	if err := os.MkdirAll(project1, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := os.MkdirAll(project2, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// First project
	dir1, err := dm.EnsureProjectDir(ctx, project1)
	if err != nil {
		t.Fatalf("first project failed: %v", err)
	}

	// Second project - different original case but same normalized name
	// Should get parent disambiguation
	dir2, err := dm.GetProjectDirName(ctx, project2)
	if err != nil {
		t.Fatalf("second project failed: %v", err)
	}

	// Both should be lowercase after normalization
	if filepath.Base(dir1) != "api-service" {
		t.Errorf("first project not normalized: %q", filepath.Base(dir1))
	}

	// Second should have parent prefix due to collision
	if dir2 != "other-api-service" {
		t.Errorf("second project: got %q, want %q", dir2, "other-api-service")
	}
}

// Subtask 4.11: Test trailing slash handling
func TestGetProjectDirName_TrailingSlash(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	projectPath := filepath.Join(tempDir, "api-service")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// With trailing slash
	pathWithSlash := projectPath + string(filepath.Separator)
	dir1, err := dm.GetProjectDirName(ctx, pathWithSlash)
	if err != nil {
		t.Fatalf("path with trailing slash failed: %v", err)
	}

	// Without trailing slash
	dir2, err := dm.GetProjectDirName(ctx, projectPath)
	if err != nil {
		t.Fatalf("path without trailing slash failed: %v", err)
	}

	if dir1 != dir2 {
		t.Errorf("trailing slash affects result: with=%q, without=%q", dir1, dir2)
	}
}

// Test normalizeName function directly
func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"api-service", "api-service"},
		{"API-Service", "api-service"},
		{"My Project", "my-project"},
		{"my--project", "my-project"},
		{"-project-", "project"},
		{"api:service", "api-service"},
		{"api_service", "api-service"},
		{"api.service", "api-service"},
		{"  api  service  ", "api-service"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeName(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// Test buildPathSegments function
func TestBuildPathSegments(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"/home/user/project", []string{"project", "user", "home"}},
		{"/a/b/c", []string{"c", "b", "a"}},
		{"/project", []string{"project"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := buildPathSegments(tt.path)
			if len(got) != len(tt.expected) {
				t.Errorf("buildPathSegments(%q) = %v (len %d), want %v (len %d)",
					tt.path, got, len(got), tt.expected, len(tt.expected))
				return
			}
			for i, seg := range got {
				if seg != tt.expected[i] {
					t.Errorf("buildPathSegments(%q)[%d] = %q, want %q",
						tt.path, i, seg, tt.expected[i])
				}
			}
		})
	}
}

// Test buildDirName function
func TestBuildDirName(t *testing.T) {
	segments := []string{"project", "client", "work", "user"}

	tests := []struct {
		depth    int
		expected string
	}{
		{0, "project"},
		{1, "client-project"},
		{2, "work-client-project"},
		{3, "user-work-client-project"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := buildDirName(segments, tt.depth)
			if got != tt.expected {
				t.Errorf("buildDirName(segments, %d) = %q, want %q", tt.depth, got, tt.expected)
			}
		})
	}
}

// Test EnsureProjectDir creates directory
func TestEnsureProjectDir_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	projectPath := filepath.Join(tempDir, "my-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	fullPath, err := dm.EnsureProjectDir(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("EnsureProjectDir failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("created path is not a directory")
	}

	// Verify marker file exists
	markerPath := filepath.Join(fullPath, ".project-path")
	if _, err := os.Stat(markerPath); err != nil {
		t.Errorf("marker file not created: %v", err)
	}
}

// Test NewDirectoryManager with empty basePath uses default
func TestNewDirectoryManager_DefaultBasePath(t *testing.T) {
	dm := NewDirectoryManager("", &mockPathLookup{})

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".vibe-dash")

	if dm == nil {
		t.Fatal("NewDirectoryManager returned nil")
	}
	if dm.basePath != expected {
		t.Errorf("default basePath = %q, want %q", dm.basePath, expected)
	}
}

// Test NewDirectoryManager with nil configLookup (HIGH-3 fix)
func TestNewDirectoryManager_NilConfigLookup(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// nil configLookup should be handled gracefully
	dm := NewDirectoryManager(basePath, nil)
	if dm == nil {
		t.Fatal("NewDirectoryManager returned nil with valid basePath and nil lookup")
	}

	// Create a test project and ensure it works without lookup
	projectPath := filepath.Join(tempDir, "test-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	got, err := dm.GetProjectDirName(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("unexpected error with nil lookup: %v", err)
	}

	if got != "test-project" {
		t.Errorf("got %q, want %q", got, "test-project")
	}
}

// Test context cancellation (HIGH-1 fix)
func TestGetProjectDirName_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	projectPath := filepath.Join(tempDir, "api-service")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := dm.GetProjectDirName(ctx, projectPath)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}

	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("expected ErrPathNotAccessible, got: %v", err)
	}
}

// Test context cancellation for EnsureProjectDir
func TestEnsureProjectDir_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	projectPath := filepath.Join(tempDir, "api-service")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := dm.EnsureProjectDir(ctx, projectPath)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}

	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("expected ErrPathNotAccessible, got: %v", err)
	}
}

// Test Unicode path handling (HIGH-4 fix)
// Note: Pure Unicode directory names normalize to empty string, but the collision
// resolution algorithm falls back to parent directory names. Since temp directories
// have valid ASCII parent names, this provides a fallback.
func TestGetProjectDirName_UnicodePath(t *testing.T) {
	tests := []struct {
		name     string
		dirName  string
		expected string
	}{
		{
			name:     "mixed ascii and unicode",
			dirName:  "project-日本語",
			expected: "project", // Unicode stripped, trailing hyphen trimmed
		},
		{
			name:     "emoji in name",
			dirName:  "my-🚀-project",
			expected: "my-project", // Emoji normalized to hyphen, collapsed
		},
		{
			name:     "accented characters",
			dirName:  "café-project",
			expected: "caf-project", // é becomes hyphen
		},
		{
			name:     "numbers with unicode",
			dirName:  "api-日本-123",
			expected: "api-123", // Unicode stripped, numbers preserved
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			basePath := filepath.Join(tempDir, "vibe-dash")

			projectPath := filepath.Join(tempDir, tt.dirName)
			if err := os.MkdirAll(projectPath, 0755); err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			dm := NewDirectoryManager(basePath, &mockPathLookup{})

			got, err := dm.GetProjectDirName(context.Background(), projectPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

// Test that pure Unicode paths fall back to parent directory names
func TestGetProjectDirName_PureUnicodeFallback(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create ASCII parent directory, then Unicode child
	parentDir := filepath.Join(tempDir, "ascii-parent")
	projectPath := filepath.Join(parentDir, "日本語プロジェクト")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	got, err := dm.GetProjectDirName(context.Background(), projectPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fall back to parent directory name since Unicode normalizes to empty
	expected := "ascii-parent"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

// Test digits in path names (MEDIUM-4 fix)
func TestGetProjectDirName_WithDigits(t *testing.T) {
	tests := []struct {
		name     string
		dirName  string
		expected string
	}{
		{
			name:     "leading digits",
			dirName:  "123-api",
			expected: "123-api",
		},
		{
			name:     "trailing digits",
			dirName:  "api-123",
			expected: "api-123",
		},
		{
			name:     "only digits",
			dirName:  "12345",
			expected: "12345",
		},
		{
			name:     "mixed alphanumeric",
			dirName:  "api2service3",
			expected: "api2service3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			basePath := filepath.Join(tempDir, "vibe-dash")

			projectPath := filepath.Join(tempDir, tt.dirName)
			if err := os.MkdirAll(projectPath, 0755); err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			dm := NewDirectoryManager(basePath, &mockPathLookup{})

			got, err := dm.GetProjectDirName(context.Background(), projectPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

// Test relative path handling (MEDIUM-2 fix)
func TestGetProjectDirName_RelativePath(t *testing.T) {
	// Save current working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origWd)

	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create project directory
	projectPath := filepath.Join(tempDir, "my-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	// Use relative path
	got, err := dm.GetProjectDirName(context.Background(), "./my-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "my-project"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}
