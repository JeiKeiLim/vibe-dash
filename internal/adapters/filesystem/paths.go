// Package filesystem provides OS abstraction for file system operations.
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// ResolvePath converts any path to absolute, verifying existence.
// Returns domain.ErrPathNotAccessible if path is empty, doesn't exist, or can't be accessed.
func ResolvePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: empty path", domain.ErrPathNotAccessible)
	}

	expanded, err := ExpandHome(path)
	if err != nil {
		return "", err // Already wrapped with domain error
	}

	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("%w: failed to resolve %s: %v", domain.ErrPathNotAccessible, path, err)
	}

	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("%w: path does not exist: %s", domain.ErrPathNotAccessible, path)
	}

	return absPath, nil
}

// CanonicalPath resolves symlinks to get the "true" physical path.
// Used for collision detection (same physical location via different paths).
// Returns domain.ErrPathNotAccessible if path doesn't exist or can't be resolved.
func CanonicalPath(path string) (string, error) {
	resolved, err := ResolvePath(path)
	if err != nil {
		return "", err
	}

	canonical, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return "", fmt.Errorf("%w: symlink resolution failed for %s", domain.ErrPathNotAccessible, path)
	}

	return canonical, nil
}

// ExpandHome expands ~ prefix to user's home directory.
// NOTE: ~user syntax (e.g., ~bob) is NOT supported - treated as ~/user.
// Returns original path unchanged if no ~ prefix.
// Returns domain.ErrPathNotAccessible if home directory cannot be determined.
func ExpandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w: cannot determine home directory", domain.ErrPathNotAccessible)
	}

	if path == "~" {
		return home, nil
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}

	// ~foo (no slash) -> treated as ~/foo (documented limitation)
	return filepath.Join(home, path[1:]), nil
}
