// Package filesystem provides OS abstraction for file system operations.
package filesystem

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

const (
	// DefaultDebounce is the default debounce window for file events (200ms from architecture doc).
	DefaultDebounce = 200 * time.Millisecond

	// eventBufferSize is the channel buffer size to prevent blocking on slow consumers.
	// 100 is sufficient to buffer typical editor save bursts without blocking.
	eventBufferSize = 100

	// defaultMaxWatchDepth is the maximum directory depth to recurse into (Story 8.1).
	// Prevents runaway recursion in deeply nested structures.
	defaultMaxWatchDepth = 10

	// defaultMaxDirsPerProject is the maximum number of directories to watch per project (Story 8.1).
	// Prevents fsnotify overload on large projects.
	defaultMaxDirsPerProject = 500
)

// skipPatterns defines directories to exclude from recursive watching.
// These are typically large vendor/build directories or VCS internals.
// NOTE: Intentionally NOT skipping "bin", "build", "dist" as these are common source
// directory names in some projects (per Dev Notes in Story 8.1).
var skipPatterns = map[string]bool{
	".git":         true,
	".svn":         true,
	".hg":          true,
	"node_modules": true,
	"vendor":       true, // Go vendor, PHP composer
	"__pycache__":  true,
	".vscode":      true,
	".idea":        true,
	"target":       true, // Rust/Maven/Cargo build output
	".cache":       true, // Various tools
}

// FsnotifyWatcher implements ports.FileWatcher using fsnotify.
// It provides debounced file system event watching with graceful shutdown support.
type FsnotifyWatcher struct {
	debounce time.Duration
	watcher  *fsnotify.Watcher
	mu       sync.Mutex
	closed   bool

	// Debounce state
	timer   *time.Timer
	pending map[string]ports.FileEvent

	// Story 7.1: Failed path tracking for graceful degradation
	failedPaths []string
}

// NewFsnotifyWatcher creates a new FsnotifyWatcher with the specified debounce duration.
// If debounce is 0, DefaultDebounce (200ms) is used.
func NewFsnotifyWatcher(debounce time.Duration) *FsnotifyWatcher {
	if debounce == 0 {
		debounce = DefaultDebounce
	}
	return &FsnotifyWatcher{
		debounce: debounce,
		pending:  make(map[string]ports.FileEvent),
	}
}

// Watch starts monitoring the specified paths for file system changes.
// Returns a channel that emits FileEvent for each detected change.
// The channel is closed when the watcher is closed or context is cancelled.
func (w *FsnotifyWatcher) Watch(ctx context.Context, paths []string) (<-chan ports.FileEvent, error) {
	// Check context cancellation at start (pattern from directory.go:53-59)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate paths
	if len(paths) == 0 {
		return nil, fmt.Errorf("%w: no paths provided to watch", domain.ErrPathNotAccessible)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil, fmt.Errorf("%w: watcher is closed", domain.ErrPathNotAccessible)
	}

	// Create fsnotify watcher
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create fsnotify watcher: %v", domain.ErrPathNotAccessible, err)
	}

	// Story 7.1: Reset failed paths for new watch session
	w.failedPaths = nil

	// Story 8.1: Collect all paths including subdirectories for recursive watching
	var allPaths []string
	for _, path := range paths {
		// Resolve to canonical path
		canonical, err := CanonicalPath(path)
		if err != nil {
			slog.Error("failed to resolve path", "path", path, "error", err)
			w.failedPaths = append(w.failedPaths, path)
			continue
		}

		// Story 8.1: Enumerate all subdirectories for this project path
		subdirs, err := getAllSubdirectories(canonical, defaultMaxWatchDepth, defaultMaxDirsPerProject)
		if err != nil {
			slog.Error("failed to enumerate subdirectories", "path", canonical, "error", err)
			// Still try to watch the root path
			allPaths = append(allPaths, canonical)
		} else {
			allPaths = append(allPaths, subdirs...)
		}
	}

	// Add all paths (including subdirs) to watcher
	validPaths := 0
	for _, watchPath := range allPaths {
		if err := fsWatcher.Add(watchPath); err != nil {
			slog.Debug("failed to watch path", "path", watchPath, "error", err)
			w.failedPaths = append(w.failedPaths, watchPath)
			continue
		}
		validPaths++
	}

	if validPaths == 0 {
		fsWatcher.Close()
		return nil, fmt.Errorf("%w: no valid paths to watch", domain.ErrPathNotAccessible)
	}

	// Story 8.1: Log total paths watched at debug level
	slog.Debug("watching paths", "count", validPaths, "total_enumerated", len(allPaths))

	// Warn if approaching OS limits
	if validPaths > 5000 {
		slog.Warn("high watch count may hit OS limits",
			"count", validPaths,
			"hint", "check /proc/sys/fs/inotify/max_user_watches on Linux")
	}

	// Log summary of partial failures if any paths failed
	if validPaths < len(allPaths) {
		slog.Warn("partial watch setup", "valid_paths", validPaths, "total_paths", len(allPaths))
	}

	w.watcher = fsWatcher
	w.pending = make(map[string]ports.FileEvent)

	// Create output channel with buffer to prevent blocking
	out := make(chan ports.FileEvent, eventBufferSize)

	// Start event processing goroutine
	go w.eventLoop(ctx, out)

	return out, nil
}

// GetFailedPaths returns the list of paths that failed to watch (Story 7.1).
// Returns nil if all paths were successfully watched.
// Thread-safe.
func (w *FsnotifyWatcher) GetFailedPaths() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.failedPaths) == 0 {
		return nil
	}
	// Return a copy to prevent external modification
	result := make([]string, len(w.failedPaths))
	copy(result, w.failedPaths)
	return result
}

// Close stops watching and releases all resources.
// Close is idempotent - calling it multiple times is safe.
func (w *FsnotifyWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true

	// Stop debounce timer if running.
	// Note: timer.Stop() returns false if the timer already fired, meaning the callback
	// may be queued or executing. The flushPending callback checks w.closed to handle this.
	if w.timer != nil {
		w.timer.Stop()
		w.timer = nil
	}

	// Close fsnotify watcher if it exists
	if w.watcher != nil {
		if err := w.watcher.Close(); err != nil {
			return fmt.Errorf("failed to close fsnotify watcher: %w", err)
		}
		w.watcher = nil
	}

	slog.Debug("file watcher closed")
	return nil
}

// AddPath adds a new path to watch dynamically.
// Returns error if path is invalid or watcher is closed.
func (w *FsnotifyWatcher) AddPath(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed || w.watcher == nil {
		return fmt.Errorf("%w: watcher is not running", domain.ErrPathNotAccessible)
	}

	canonical, err := CanonicalPath(path)
	if err != nil {
		return err
	}

	if err := w.watcher.Add(canonical); err != nil {
		return fmt.Errorf("%w: failed to add path %s: %v", domain.ErrPathNotAccessible, canonical, err)
	}

	return nil
}

// RemovePath removes a path from watching.
// Returns error if path is not being watched or watcher is closed.
func (w *FsnotifyWatcher) RemovePath(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed || w.watcher == nil {
		return fmt.Errorf("%w: watcher is not running", domain.ErrPathNotAccessible)
	}

	canonical, err := CanonicalPath(path)
	if err != nil {
		return err
	}

	if err := w.watcher.Remove(canonical); err != nil {
		return fmt.Errorf("%w: failed to remove path %s: %v", domain.ErrPathNotAccessible, canonical, err)
	}

	return nil
}

// eventLoop processes fsnotify events and applies debouncing.
// It runs until context is cancelled or watcher is closed.
func (w *FsnotifyWatcher) eventLoop(ctx context.Context, out chan<- ports.FileEvent) {
	defer close(out)

	// Capture watcher reference - it may be set to nil during Close()
	w.mu.Lock()
	fsWatcher := w.watcher
	w.mu.Unlock()

	if fsWatcher == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			// Context cancelled - cleanup and exit
			w.flushPending(out)
			return

		case event, ok := <-fsWatcher.Events:
			if !ok {
				// Watcher closed
				w.flushPending(out)
				return
			}
			w.handleEvent(event, out)

		case err, ok := <-fsWatcher.Errors:
			if !ok {
				// Error channel closed
				w.flushPending(out)
				return
			}
			slog.Error("fsnotify error", "error", err)
		}
	}
}

// handleEvent processes a single fsnotify event with debouncing.
func (w *FsnotifyWatcher) handleEvent(event fsnotify.Event, out chan<- ports.FileEvent) {
	// Translate fsnotify operation to ports.FileOperation
	op := translateOperation(event.Op)
	if op == -1 {
		// Unknown operation, skip
		return
	}

	// Get canonical path for the event
	canonical, err := CanonicalPath(event.Name)
	if err != nil {
		// Path may have been deleted, use original name
		canonical = event.Name
	}

	fileEvent := ports.FileEvent{
		Path:      canonical,
		Operation: op,
		Timestamp: time.Now(),
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Track event by path (overwrites previous event for same path)
	w.pending[canonical] = fileEvent

	// Reset debounce timer
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(w.debounce, func() {
		w.flushPending(out)
	})
}

// flushPending emits all pending events and clears the pending map.
// This method is called from eventLoop (on shutdown) and from timer callbacks.
// It acquires the mutex internally, so callers must NOT hold the lock.
func (w *FsnotifyWatcher) flushPending(out chan<- ports.FileEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if watcher is closed to prevent emitting events after shutdown
	if w.closed {
		return
	}

	for _, evt := range w.pending {
		select {
		case out <- evt:
		default:
			// Channel full, log warning but don't block
			slog.Warn("event channel full, dropping event", "path", evt.Path)
		}
	}
	w.pending = make(map[string]ports.FileEvent)
}

// translateOperation converts fsnotify.Op to ports.FileOperation.
// Returns -1 for unknown or ignored operations (chmod).
// The caller should skip events with -1 return value.
func translateOperation(op fsnotify.Op) ports.FileOperation {
	switch {
	case op&fsnotify.Create != 0:
		return ports.FileOpCreate
	case op&fsnotify.Write != 0:
		return ports.FileOpModify
	case op&fsnotify.Remove != 0:
		return ports.FileOpDelete
	case op&fsnotify.Rename != 0:
		// Rename is treated as delete (file moved away)
		return ports.FileOpDelete
	case op&fsnotify.Chmod != 0:
		// Ignore chmod events
		return -1
	default:
		return -1
	}
}

// shouldSkipDirectory returns true if the directory should be excluded from watching.
// CRITICAL: .bmad is the ONE exception to hidden directory rule - it contains
// methodology artifacts that must trigger waiting detection.
func shouldSkipDirectory(name string) bool {
	// EXCEPTION: .bmad is ALWAYS watched (methodology artifacts)
	if name == ".bmad" {
		return false
	}
	// Skip hidden directories (starting with .)
	if strings.HasPrefix(name, ".") {
		return true
	}
	// Skip known vendor/build directories
	return skipPatterns[name]
}

// getAllSubdirectories enumerates all subdirectories of rootPath up to maxDepth levels deep,
// returning at most maxDirs directories. Directories matching skipPatterns or hidden directories
// (except .bmad) are excluded.
func getAllSubdirectories(rootPath string, maxDepth, maxDirs int) ([]string, error) {
	start := time.Now()
	var dirs []string
	rootDepth := strings.Count(rootPath, string(filepath.Separator))
	limitReached := false

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip inaccessible directories, continue with others
			return nil
		}
		if !d.IsDir() {
			// Skip files
			return nil
		}

		// Check depth limit
		currentDepth := strings.Count(path, string(filepath.Separator)) - rootDepth
		if currentDepth > maxDepth {
			return filepath.SkipDir
		}

		// Check directory count limit
		if len(dirs) >= maxDirs {
			if !limitReached {
				slog.Warn("directory limit reached", "limit", maxDirs, "root", rootPath)
				limitReached = true
			}
			return filepath.SkipAll
		}

		// Skip directories by pattern (but not root itself)
		if path != rootPath {
			name := d.Name()
			if shouldSkipDirectory(name) {
				return filepath.SkipDir
			}
		}

		dirs = append(dirs, path)
		return nil
	})

	slog.Debug("directory enumeration complete",
		"root", rootPath,
		"count", len(dirs),
		"duration_ms", time.Since(start).Milliseconds())

	return dirs, err
}
