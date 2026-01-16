package detection

import (
	"context"
	"sync"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// cacheTTL is the time-to-live for cached detection results.
// Prevents repeated filesystem scans during rapid TUI updates.
const cacheTTL = 5 * time.Second

// cacheEntry holds a cached detection result.
type cacheEntry struct {
	state     domain.AgentState
	timestamp time.Time
}

// AgentWaitingAdapter adapts AgentDetectionService to the existing WaitingDetector interface.
// This enables backward compatibility with TUI components that expect WaitingDetector.
type AgentWaitingAdapter struct {
	service *AgentDetectionService
	cache   map[string]cacheEntry
	mu      sync.RWMutex
	now     func() time.Time
}

// NewAgentWaitingAdapter creates a new adapter wrapping the detection service.
func NewAgentWaitingAdapter(service *AgentDetectionService) *AgentWaitingAdapter {
	return &AgentWaitingAdapter{
		service: service,
		cache:   make(map[string]cacheEntry),
		now:     time.Now,
	}
}

// Compile-time interface compliance check.
var _ ports.WaitingDetector = (*AgentWaitingAdapter)(nil)

// IsWaiting returns true if the project's agent is waiting for user input.
func (a *AgentWaitingAdapter) IsWaiting(ctx context.Context, project *domain.Project) bool {
	if project == nil || project.State == domain.StateHibernated {
		return false
	}

	state := a.detectWithCache(ctx, project.Path)
	return state.IsWaiting()
}

// WaitingDuration returns how long the project has been waiting.
func (a *AgentWaitingAdapter) WaitingDuration(ctx context.Context, project *domain.Project) time.Duration {
	if project == nil || project.State == domain.StateHibernated {
		return 0
	}

	state := a.detectWithCache(ctx, project.Path)
	if !state.IsWaiting() {
		return 0
	}
	return state.Duration
}

// detectWithCache returns cached result if fresh, otherwise performs detection.
func (a *AgentWaitingAdapter) detectWithCache(ctx context.Context, projectPath string) domain.AgentState {
	// Check cache first
	a.mu.RLock()
	if entry, ok := a.cache[projectPath]; ok {
		if a.now().Sub(entry.timestamp) < cacheTTL {
			a.mu.RUnlock()
			return entry.state
		}
	}
	a.mu.RUnlock()

	// Cache miss or stale - perform detection
	state, _ := a.service.Detect(ctx, projectPath)

	// Update cache
	a.mu.Lock()
	a.cache[projectPath] = cacheEntry{
		state:     state,
		timestamp: a.now(),
	}
	a.mu.Unlock()

	return state
}

// ClearCache clears all cached entries (for testing).
func (a *AgentWaitingAdapter) ClearCache() {
	a.mu.Lock()
	a.cache = make(map[string]cacheEntry)
	a.mu.Unlock()
}
