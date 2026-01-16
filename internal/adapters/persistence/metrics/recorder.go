package metrics

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// DefaultDebounceWindow is the default time to wait before committing a transition.
// Prevents excessive database writes during rapid stage changes.
const DefaultDebounceWindow = 10 * time.Second

// transitionRecorder is the interface for recording transitions (allows testing).
type transitionRecorder interface {
	RecordTransition(ctx context.Context, projectID, fromStage, toStage string) error
}

// pendingTransition represents a stage change waiting for debounce window to expire.
type pendingTransition struct {
	fromStage domain.Stage
	toStage   domain.Stage
	timer     *time.Timer
}

// MetricsRecorder handles stage transition detection and recording with debouncing.
// It tracks previous stages per project to detect changes and debounces rapid
// transitions to prevent database spam.
type MetricsRecorder struct {
	repo               transitionRecorder
	previousStages     map[string]domain.Stage
	pendingTransitions map[string]*pendingTransition
	debounceWindow     time.Duration
	mu                 sync.Mutex
}

// NewMetricsRecorder creates a new recorder. Accepts nil repository (graceful no-op).
func NewMetricsRecorder(repo transitionRecorder) *MetricsRecorder {
	return &MetricsRecorder{
		repo:               repo,
		previousStages:     make(map[string]domain.Stage),
		pendingTransitions: make(map[string]*pendingTransition),
		debounceWindow:     DefaultDebounceWindow,
	}
}

// OnDetection should be called when a project's stage is detected.
// It compares against the previous stage and schedules a transition if changed.
// Thread-safe, handles nil receiver gracefully.
func (r *MetricsRecorder) OnDetection(ctx context.Context, projectID string, newStage domain.Stage) {
	if r == nil || r.repo == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	prevStage, exists := r.previousStages[projectID]
	if exists && prevStage == newStage {
		// Stage unchanged - no transition to record (AC#5)
		return
	}

	// Determine from_stage: empty string for first detection (AC#4)
	fromStage := domain.Stage(0) // StageUnknown
	if exists {
		fromStage = prevStage
	}

	r.scheduleTransition(projectID, fromStage, newStage)
	r.previousStages[projectID] = newStage
}

// scheduleTransition schedules a transition for recording after debounce window.
// If a transition is already pending, updates the toStage (keeps original fromStage).
// Must be called with mu held.
func (r *MetricsRecorder) scheduleTransition(projectID string, fromStage, toStage domain.Stage) {
	if pending, exists := r.pendingTransitions[projectID]; exists {
		// Cancel existing timer, update toStage (keep original fromStage for debouncing)
		pending.timer.Stop()
		pending.toStage = toStage
	} else {
		r.pendingTransitions[projectID] = &pendingTransition{
			fromStage: fromStage,
			toStage:   toStage,
		}
	}

	pending := r.pendingTransitions[projectID]
	pending.timer = time.AfterFunc(r.debounceWindow, func() {
		r.commitTransition(projectID)
	})
}

// commitTransition writes the pending transition to the database.
// Uses context.Background() since the original context may be cancelled.
func (r *MetricsRecorder) commitTransition(projectID string) {
	r.mu.Lock()
	pending, exists := r.pendingTransitions[projectID]
	if !exists {
		r.mu.Unlock()
		return
	}

	// Convert stages to strings for database
	fromStageStr := ""
	if pending.fromStage != domain.StageUnknown {
		fromStageStr = pending.fromStage.String()
	}
	toStageStr := pending.toStage.String()

	delete(r.pendingTransitions, projectID)
	r.mu.Unlock()

	// Use background context - original context may be cancelled
	if err := r.repo.RecordTransition(context.Background(), projectID, fromStageStr, toStageStr); err != nil {
		slog.Warn("failed to record stage transition",
			"error", err,
			"project_id", projectID,
			"from_stage", fromStageStr,
			"to_stage", toStageStr)
	}
}

// Flush commits all pending transitions immediately.
// Should be called during shutdown to ensure no transitions are lost.
// Handles nil receiver gracefully.
func (r *MetricsRecorder) Flush(ctx context.Context) {
	if r == nil {
		return
	}

	r.mu.Lock()
	projectIDs := make([]string, 0, len(r.pendingTransitions))
	for id, pending := range r.pendingTransitions {
		pending.timer.Stop()
		projectIDs = append(projectIDs, id)
	}
	r.mu.Unlock()

	// Commit all pending transitions
	for _, id := range projectIDs {
		r.commitTransition(id)
	}
}
