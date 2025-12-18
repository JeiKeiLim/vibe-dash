//go:build integration

package filesystem

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// Subtask 5.1: Create temporary directories for collision testing
// Subtask 5.2: Verify directory actually gets created
func TestIntegration_EnsureProjectDir_CreatesRealDirectory(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create a "real" project directory
	projectPath := filepath.Join(tempDir, "projects", "my-app")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	fullPath, err := dm.EnsureProjectDir(ctx, projectPath)
	if err != nil {
		t.Fatalf("EnsureProjectDir failed: %v", err)
	}

	// Verify the directory actually exists
	info, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("created directory doesn't exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("created path is not a directory")
	}

	// Verify the expected path structure
	expectedPath := filepath.Join(basePath, "my-app")
	if fullPath != expectedPath {
		t.Errorf("path mismatch: got %q, want %q", fullPath, expectedPath)
	}

	// Verify marker file exists and has correct content
	markerPath := filepath.Join(fullPath, ".project-path")
	data, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("marker file not readable: %v", err)
	}

	canonicalPath, _ := CanonicalPath(projectPath)
	if string(data) != canonicalPath {
		t.Errorf("marker content mismatch: got %q, want %q", string(data), canonicalPath)
	}
}

// Test collision with real directories
func TestIntegration_EnsureProjectDir_CollisionResolution(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create two project directories with same name
	project1 := filepath.Join(tempDir, "workspace-a", "api")
	project2 := filepath.Join(tempDir, "workspace-b", "api")

	if err := os.MkdirAll(project1, 0755); err != nil {
		t.Fatalf("failed to create project1: %v", err)
	}
	if err := os.MkdirAll(project2, 0755); err != nil {
		t.Fatalf("failed to create project2: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// Add first project
	dir1, err := dm.EnsureProjectDir(ctx, project1)
	if err != nil {
		t.Fatalf("first project failed: %v", err)
	}

	// Add second project - should get disambiguation
	dir2, err := dm.EnsureProjectDir(ctx, project2)
	if err != nil {
		t.Fatalf("second project failed: %v", err)
	}

	// Both directories should exist
	if _, err := os.Stat(dir1); err != nil {
		t.Errorf("dir1 doesn't exist: %v", err)
	}
	if _, err := os.Stat(dir2); err != nil {
		t.Errorf("dir2 doesn't exist: %v", err)
	}

	// They should be different directories
	if dir1 == dir2 {
		t.Errorf("collision not resolved: both are %q", dir1)
	}

	// Verify naming: first should be "api", second should be "workspace-b-api"
	if filepath.Base(dir1) != "api" {
		t.Errorf("first dir name: got %q, want %q", filepath.Base(dir1), "api")
	}
	if filepath.Base(dir2) != "workspace-b-api" {
		t.Errorf("second dir name: got %q, want %q", filepath.Base(dir2), "workspace-b-api")
	}
}

// TestIntegration_EnsureProjectDir_CollisionResolution_ThreeProjects tests AC2: 3 projects with 2 name collisions
// Verifies that given 3 projects with same name (api) in different workspaces,
// each gets a unique directory with correct disambiguation.
func TestIntegration_EnsureProjectDir_CollisionResolution_ThreeProjects(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create three project directories with same name
	project1 := filepath.Join(tempDir, "workspace-a", "api")
	project2 := filepath.Join(tempDir, "workspace-b", "api")
	project3 := filepath.Join(tempDir, "workspace-c", "api")

	if err := os.MkdirAll(project1, 0755); err != nil {
		t.Fatalf("failed to create project1: %v", err)
	}
	if err := os.MkdirAll(project2, 0755); err != nil {
		t.Fatalf("failed to create project2: %v", err)
	}
	if err := os.MkdirAll(project3, 0755); err != nil {
		t.Fatalf("failed to create project3: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// Add first project
	dir1, err := dm.EnsureProjectDir(ctx, project1)
	if err != nil {
		t.Fatalf("first project failed: %v", err)
	}

	// Add second project - should get disambiguation
	dir2, err := dm.EnsureProjectDir(ctx, project2)
	if err != nil {
		t.Fatalf("second project failed: %v", err)
	}

	// Add third project - should get disambiguation
	dir3, err := dm.EnsureProjectDir(ctx, project3)
	if err != nil {
		t.Fatalf("third project failed: %v", err)
	}

	// All directories should exist
	if _, err := os.Stat(dir1); err != nil {
		t.Errorf("dir1 doesn't exist: %v", err)
	}
	if _, err := os.Stat(dir2); err != nil {
		t.Errorf("dir2 doesn't exist: %v", err)
	}
	if _, err := os.Stat(dir3); err != nil {
		t.Errorf("dir3 doesn't exist: %v", err)
	}

	// All should be different directories
	if dir1 == dir2 || dir1 == dir3 || dir2 == dir3 {
		t.Errorf("collision not resolved: dir1=%q, dir2=%q, dir3=%q", dir1, dir2, dir3)
	}

	// Log directories for debugging
	t.Logf("Created directories: dir1=%q, dir2=%q, dir3=%q", filepath.Base(dir1), filepath.Base(dir2), filepath.Base(dir3))

	// Verify naming: first should be "api", second should be "workspace-b-api", third should be "workspace-c-api"
	if filepath.Base(dir1) != "api" {
		t.Errorf("first dir name: got %q, want %q", filepath.Base(dir1), "api")
	}
	if filepath.Base(dir2) != "workspace-b-api" {
		t.Errorf("second dir name: got %q, want %q", filepath.Base(dir2), "workspace-b-api")
	}
	if filepath.Base(dir3) != "workspace-c-api" {
		t.Errorf("third dir name: got %q, want %q", filepath.Base(dir3), "workspace-c-api")
	}

	// Verify marker files have correct content
	for i, test := range []struct {
		dir         string
		projectPath string
	}{
		{dir1, project1},
		{dir2, project2},
		{dir3, project3},
	} {
		markerPath := filepath.Join(test.dir, ".project-path")
		data, err := os.ReadFile(markerPath)
		if err != nil {
			t.Errorf("project %d: marker file not readable: %v", i+1, err)
			continue
		}

		canonicalPath, _ := CanonicalPath(test.projectPath)
		if string(data) != canonicalPath {
			t.Errorf("project %d: marker content mismatch: got %q, want %q", i+1, string(data), canonicalPath)
		}
	}
}

// Subtask 5.3: Test permission error handling with read-only directory
func TestIntegration_EnsureProjectDir_PermissionError(t *testing.T) {
	// Skip if running as root (permissions won't apply)
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	tempDir := t.TempDir()

	// Create read-only base directory
	basePath := filepath.Join(tempDir, "vibe-dash")
	if err := os.MkdirAll(basePath, 0755); err != nil {
		t.Fatalf("failed to create base dir: %v", err)
	}

	// Make it read-only
	if err := os.Chmod(basePath, 0444); err != nil {
		t.Fatalf("failed to chmod base dir: %v", err)
	}
	defer os.Chmod(basePath, 0755) // Restore for cleanup

	// Create project directory
	projectPath := filepath.Join(tempDir, "my-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})

	_, err := dm.EnsureProjectDir(context.Background(), projectPath)
	if err == nil {
		t.Fatal("expected permission error")
	}

	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("expected ErrPathNotAccessible, got: %v", err)
	}
}

// Test that same project path returns same directory name (determinism with marker file)
func TestIntegration_EnsureProjectDir_Determinism(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	projectPath := filepath.Join(tempDir, "my-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// First call creates directory
	dir1, err := dm.EnsureProjectDir(ctx, projectPath)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call should return same path
	dir2, err := dm.EnsureProjectDir(ctx, projectPath)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if dir1 != dir2 {
		t.Errorf("not deterministic: first=%q, second=%q", dir1, dir2)
	}

	// Verify only one directory exists in basePath (not two)
	entries, err := os.ReadDir(basePath)
	if err != nil {
		t.Fatalf("failed to read base dir: %v", err)
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 directory, got %d", count)
	}
}

// Test symlink handling in real filesystem
func TestIntegration_EnsureProjectDir_SymlinkHandling(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "vibe-dash")

	// Create real project directory
	realProject := filepath.Join(tempDir, "real-project")
	if err := os.MkdirAll(realProject, 0755); err != nil {
		t.Fatalf("failed to create real project: %v", err)
	}

	// Create symlink to it
	symlinkProject := filepath.Join(tempDir, "symlink-project")
	if err := os.Symlink(realProject, symlinkProject); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	dm := NewDirectoryManager(basePath, &mockPathLookup{})
	ctx := context.Background()

	// Add via real path
	dir1, err := dm.EnsureProjectDir(ctx, realProject)
	if err != nil {
		t.Fatalf("real path failed: %v", err)
	}

	// Add via symlink path - should recognize it's the same project
	dir2, err := dm.EnsureProjectDir(ctx, symlinkProject)
	if err != nil {
		t.Fatalf("symlink path failed: %v", err)
	}

	// Both should resolve to same directory
	if dir1 != dir2 {
		t.Errorf("symlink not resolved: real=%q, symlink=%q", dir1, dir2)
	}
}
