package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// Test 4.2: ResolvePath with empty string returns ErrPathNotAccessible
func TestResolvePath_EmptyString(t *testing.T) {
	_, err := ResolvePath("")
	if err == nil {
		t.Error("ResolvePath(\"\") expected error, got nil")
	}
	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("ResolvePath(\"\") error = %v, want ErrPathNotAccessible", err)
	}
}

// Test 4.3: ResolvePath with absolute path (existing)
func TestResolvePath_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()

	got, err := ResolvePath(tmpDir)
	if err != nil {
		t.Fatalf("ResolvePath(%q) error = %v", tmpDir, err)
	}
	if got != tmpDir {
		t.Errorf("ResolvePath(%q) = %q, want %q", tmpDir, got, tmpDir)
	}
}

// Test 4.4: ResolvePath with "." returns current directory
func TestResolvePath_CurrentDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}

	got, err := ResolvePath(".")
	if err != nil {
		t.Fatalf("ResolvePath(\".\") error = %v", err)
	}
	if got != cwd {
		t.Errorf("ResolvePath(\".\") = %q, want %q", got, cwd)
	}
}

// Test 4.5: ResolvePath with relative path
func TestResolvePath_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Change to tmpDir and resolve relative path
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("os.Chdir(%q) error = %v", tmpDir, err)
	}

	got, err := ResolvePath("subdir")
	if err != nil {
		t.Fatalf("ResolvePath(\"subdir\") error = %v", err)
	}
	// Use EvalSymlinks to get canonical path for comparison (handles macOS /var -> /private/var)
	wantCanonical, _ := filepath.EvalSymlinks(subDir)
	gotCanonical, _ := filepath.EvalSymlinks(got)
	if gotCanonical != wantCanonical {
		t.Errorf("ResolvePath(\"subdir\") = %q, want %q", got, subDir)
	}
}

// Test 4.6: ResolvePath with non-existent path returns ErrPathNotAccessible
func TestResolvePath_NonExistent(t *testing.T) {
	_, err := ResolvePath("/this/path/definitely/does/not/exist/xyz123")
	if err == nil {
		t.Error("ResolvePath(non-existent) expected error, got nil")
	}
	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("ResolvePath(non-existent) error = %v, want ErrPathNotAccessible", err)
	}
}

// Test 4.7: CanonicalPath resolves symlinks
func TestCanonicalPath_Symlink(t *testing.T) {
	tmpDir := t.TempDir()
	actualDir := filepath.Join(tmpDir, "actual")
	if err := os.Mkdir(actualDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	linkPath := filepath.Join(tmpDir, "link")
	if err := os.Symlink(actualDir, linkPath); err != nil {
		t.Skipf("Symlinks not supported: %v", err)
	}

	canonical, err := CanonicalPath(linkPath)
	if err != nil {
		t.Fatalf("CanonicalPath() error = %v", err)
	}
	// Get expected canonical path (handles macOS /var -> /private/var)
	wantCanonical, _ := filepath.EvalSymlinks(actualDir)
	if canonical != wantCanonical {
		t.Errorf("CanonicalPath(%q) = %q, want %q", linkPath, canonical, wantCanonical)
	}
}

// Test 4.8: CanonicalPath with regular path (no symlink)
func TestCanonicalPath_RegularPath(t *testing.T) {
	tmpDir := t.TempDir()

	got, err := CanonicalPath(tmpDir)
	if err != nil {
		t.Fatalf("CanonicalPath(%q) error = %v", tmpDir, err)
	}
	// Get expected canonical path (handles macOS /var -> /private/var)
	wantCanonical, _ := filepath.EvalSymlinks(tmpDir)
	if got != wantCanonical {
		t.Errorf("CanonicalPath(%q) = %q, want %q", tmpDir, got, wantCanonical)
	}
}

// Test 4.9: CanonicalPath with non-existent path returns error
func TestCanonicalPath_NonExistent(t *testing.T) {
	_, err := CanonicalPath("/this/path/definitely/does/not/exist/xyz123")
	if err == nil {
		t.Error("CanonicalPath(non-existent) expected error, got nil")
	}
	if !errors.Is(err, domain.ErrPathNotAccessible) {
		t.Errorf("CanonicalPath(non-existent) error = %v, want ErrPathNotAccessible", err)
	}
}

// Test 4.10: ExpandHome with "~" only
func TestExpandHome_TildeOnly(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}

	got, err := ExpandHome("~")
	if err != nil {
		t.Fatalf("ExpandHome(\"~\") error = %v", err)
	}
	if got != home {
		t.Errorf("ExpandHome(\"~\") = %q, want %q", got, home)
	}
}

// Test 4.11: ExpandHome with "~/" prefix
func TestExpandHome_TildeSlash(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}

	got, err := ExpandHome("~/project")
	if err != nil {
		t.Fatalf("ExpandHome(\"~/project\") error = %v", err)
	}
	want := filepath.Join(home, "project")
	if got != want {
		t.Errorf("ExpandHome(\"~/project\") = %q, want %q", got, want)
	}
}

// Test 4.12: ExpandHome with "~foo" (no slash) - documented as ~/foo
func TestExpandHome_TildeNoSlash(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}

	got, err := ExpandHome("~foo")
	if err != nil {
		t.Fatalf("ExpandHome(\"~foo\") error = %v", err)
	}
	// Documented behavior: ~foo is treated as ~/foo
	want := filepath.Join(home, "foo")
	if got != want {
		t.Errorf("ExpandHome(\"~foo\") = %q, want %q", got, want)
	}
}

// Test 4.13: ExpandHome with no ~ returns original
func TestExpandHome_NoTilde(t *testing.T) {
	input := "/absolute/path"
	got, err := ExpandHome(input)
	if err != nil {
		t.Fatalf("ExpandHome(%q) error = %v", input, err)
	}
	if got != input {
		t.Errorf("ExpandHome(%q) = %q, want %q", input, got, input)
	}
}

// Test 4.14: ExpandHome with ~ in middle (not prefix) returns unchanged
func TestExpandHome_TildeInMiddle(t *testing.T) {
	input := "/path/~/here"
	got, err := ExpandHome(input)
	if err != nil {
		t.Fatalf("ExpandHome(%q) error = %v", input, err)
	}
	if got != input {
		t.Errorf("ExpandHome(%q) = %q, want %q", input, got, input)
	}
}

// Test 4.15: Two symlinks to same location produce same canonical path
func TestCanonicalPath_TwoSymlinksToSameLocation(t *testing.T) {
	tmpDir := t.TempDir()
	actualDir := filepath.Join(tmpDir, "actual")
	if err := os.Mkdir(actualDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	link1 := filepath.Join(tmpDir, "link1")
	link2 := filepath.Join(tmpDir, "link2")

	if err := os.Symlink(actualDir, link1); err != nil {
		t.Skipf("Symlinks not supported: %v", err)
	}
	if err := os.Symlink(actualDir, link2); err != nil {
		t.Fatalf("failed to create second symlink: %v", err)
	}

	canonical1, err := CanonicalPath(link1)
	if err != nil {
		t.Fatalf("CanonicalPath(%q) error = %v", link1, err)
	}

	canonical2, err := CanonicalPath(link2)
	if err != nil {
		t.Fatalf("CanonicalPath(%q) error = %v", link2, err)
	}

	if canonical1 != canonical2 {
		t.Errorf("Two symlinks to same location should produce same canonical path: %q != %q", canonical1, canonical2)
	}
	// Get expected canonical path (handles macOS /var -> /private/var)
	wantCanonical, _ := filepath.EvalSymlinks(actualDir)
	if canonical1 != wantCanonical {
		t.Errorf("Canonical path should be %q, got %q", wantCanonical, canonical1)
	}
}

// Additional test: ExpandHome with empty string - should return empty (no error from ExpandHome itself)
func TestExpandHome_EmptyString(t *testing.T) {
	got, err := ExpandHome("")
	if err != nil {
		t.Fatalf("ExpandHome(\"\") error = %v", err)
	}
	if got != "" {
		t.Errorf("ExpandHome(\"\") = %q, want \"\"", got)
	}
}
