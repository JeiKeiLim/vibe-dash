// Package filesystem provides OS abstraction for file system operations.
package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

const maxCollisionDepth = 10

// normalizeRegex matches any character that is not alphanumeric or hyphen.
var normalizeRegex = regexp.MustCompile(`[^a-z0-9-]+`)

// multiHyphenRegex matches multiple consecutive hyphens.
var multiHyphenRegex = regexp.MustCompile(`-{2,}`)

// FilesystemDirectoryManager implements ports.DirectoryManager using the local filesystem.
type FilesystemDirectoryManager struct {
	basePath     string
	configLookup ports.ProjectPathLookup
}

// NewDirectoryManager creates DirectoryManager with configurable base path.
// basePath defaults to ~/.vibe-dash if empty string provided.
// configLookup provides existing project → directory mappings for determinism (can be nil).
// Returns nil if basePath is empty and home directory cannot be determined.
func NewDirectoryManager(basePath string, configLookup ports.ProjectPathLookup) *FilesystemDirectoryManager {
	if basePath == "" {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			// Cannot determine home directory - return nil to signal error
			// Caller should handle this gracefully
			return nil
		}
		basePath = filepath.Join(home, ".vibe-dash")
	}
	return &FilesystemDirectoryManager{
		basePath:     basePath,
		configLookup: configLookup,
	}
}

// GetProjectDirName returns deterministic directory name for project.
// Uses collision resolution if name already exists for different project.
// The context parameter is accepted for interface compliance and future cancellation support.
func (dm *FilesystemDirectoryManager) GetProjectDirName(ctx context.Context, projectPath string) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("%w: operation cancelled: %v", domain.ErrPathNotAccessible, ctx.Err())
	default:
	}

	// Resolve canonical path (handles symlinks)
	canonicalPath, err := CanonicalPath(projectPath)
	if err != nil {
		return "", err
	}

	// Check if this exact path already has a directory assigned (determinism)
	if dm.configLookup != nil {
		if existingDir := dm.configLookup.GetDirForPath(canonicalPath); existingDir != "" {
			return existingDir, nil
		}
	}

	// Build path segments for collision resolution
	segments := buildPathSegments(canonicalPath)
	if len(segments) == 0 {
		return "", fmt.Errorf("%w: cannot derive directory name from path: %s", domain.ErrPathNotAccessible, projectPath)
	}

	// Try to find unique name using increasing parent depth
	for depth := 0; depth < maxCollisionDepth && depth < len(segments); depth++ {
		name := buildDirName(segments, depth)

		// If normalized name is empty (e.g., pure Unicode path), skip this depth
		if name == "" {
			continue
		}

		// Check if this name already exists in the base directory
		dirPath := filepath.Join(dm.basePath, name)
		exists, err := dm.directoryExists(dirPath)
		if err != nil {
			return "", err
		}

		if !exists {
			// No collision - this name is available
			return name, nil
		}

		// Directory exists - check if it's the same project (determinism check via marker file)
		if dm.isSameProject(dirPath, canonicalPath) {
			return name, nil
		}

		// Collision with different project - try next depth
	}

	return "", fmt.Errorf("%w: path %s collides after %d levels", domain.ErrCollisionUnresolvable, projectPath, maxCollisionDepth)
}

// EnsureProjectDir creates project directory if not exists.
// Returns full path to created/existing directory.
// The context parameter is used for cancellation support.
func (dm *FilesystemDirectoryManager) EnsureProjectDir(ctx context.Context, projectPath string) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("%w: operation cancelled: %v", domain.ErrPathNotAccessible, ctx.Err())
	default:
	}

	dirName, err := dm.GetProjectDirName(ctx, projectPath)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(dm.basePath, dirName)

	// Create directory atomically (MkdirAll handles existing directories)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", fmt.Errorf("%w: failed to create directory %s: %v", domain.ErrPathNotAccessible, fullPath, err)
	}

	// Write marker file to track which canonical path this directory belongs to
	canonicalPath, _ := CanonicalPath(projectPath)
	if err := dm.writeProjectMarker(fullPath, canonicalPath); err != nil {
		return "", err
	}

	return fullPath, nil
}

// buildPathSegments extracts directory names from path in reverse order (child to root).
// Example: "/home/user/work/client/api-service" → ["api-service", "client", "work", "user", "home"]
func buildPathSegments(path string) []string {
	// Clean and normalize the path
	path = filepath.Clean(path)
	path = strings.TrimSuffix(path, string(filepath.Separator))

	var segments []string
	current := path

	for {
		dir := filepath.Base(current)
		if dir == "." || dir == string(filepath.Separator) || dir == "" {
			break
		}

		// Handle filesystem root
		parent := filepath.Dir(current)
		if parent == current {
			// At root - use "root" as final segment
			if len(segments) == 0 {
				segments = append(segments, "root")
			}
			break
		}

		segments = append(segments, dir)
		current = parent
	}

	return segments
}

// buildDirName constructs normalized directory name from segments at given depth.
// depth 0: just base name
// depth 1: parent-basename
// depth 2: grandparent-parent-basename
func buildDirName(segments []string, depth int) string {
	if depth >= len(segments) {
		depth = len(segments) - 1
	}

	var parts []string
	for i := depth; i >= 0; i-- {
		parts = append(parts, segments[i])
	}

	combined := strings.Join(parts, "-")
	return normalizeName(combined)
}

// normalizeName applies directory name normalization rules:
// 1. Convert to lowercase
// 2. Replace non-alphanumeric (except hyphen) with hyphens
// 3. Collapse multiple consecutive hyphens
// 4. Trim leading/trailing hyphens
func normalizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace non-alphanumeric (except hyphen) with hyphens
	name = normalizeRegex.ReplaceAllString(name, "-")

	// Collapse multiple hyphens
	name = multiHyphenRegex.ReplaceAllString(name, "-")

	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	return name
}

// directoryExists checks if a directory exists at the given path.
func (dm *FilesystemDirectoryManager) directoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("%w: failed to stat directory %s: %v", domain.ErrPathNotAccessible, path, err)
	}
	return info.IsDir(), nil
}

// isSameProject checks if the directory at dirPath belongs to the given canonical project path.
// Uses a marker file to track the original project path.
func (dm *FilesystemDirectoryManager) isSameProject(dirPath, canonicalPath string) bool {
	markerPath := filepath.Join(dirPath, ".project-path")
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == canonicalPath
}

// writeProjectMarker writes the canonical project path to a marker file.
func (dm *FilesystemDirectoryManager) writeProjectMarker(dirPath, canonicalPath string) error {
	markerPath := filepath.Join(dirPath, ".project-path")
	if err := os.WriteFile(markerPath, []byte(canonicalPath), 0644); err != nil {
		return fmt.Errorf("%w: failed to write project marker %s: %v", domain.ErrPathNotAccessible, markerPath, err)
	}
	return nil
}
