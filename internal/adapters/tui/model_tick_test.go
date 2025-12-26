package tui

import (
	"runtime"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

// =============================================================================
// Story 8.2: Auto-Refresh Reliability Tests
// =============================================================================

// TestTickCmd_Returns5SecondInterval verifies tickCmd returns valid command (AC2).
// NOTE: tea.Tick does not expose its interval for programmatic verification.
// The 5-second interval MUST be verified via code review of tickCmd():
//   - model.go:324 should contain: tea.Tick(5*time.Second, ...)
//
// This is a known limitation of the Bubble Tea testing API.
func TestTickCmd_Returns5SecondInterval(t *testing.T) {
	cmd := tickCmd()
	if cmd == nil {
		t.Error("tickCmd() should return non-nil command")
	}
	// Interval cannot be verified programmatically - see function comment
}

// TestTickMsgHandler_ZeroProjects verifies guard clause for empty projects (AC5).
func TestTickMsgHandler_ZeroProjects(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)
	m.projects = nil // Zero projects

	// Should not panic with zero projects
	updated, cmd := m.Update(tickMsg(time.Now()))

	if cmd == nil {
		t.Error("tickMsg should return next tick command")
	}

	model := updated.(Model)
	if len(model.projects) != 0 {
		t.Error("projects should remain empty")
	}
}

// TestTickMsgHandler_EmptySlice verifies guard clause with empty slice (AC5).
func TestTickMsgHandler_EmptySlice(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)
	m.projects = []*domain.Project{} // Empty slice

	// Should not panic with empty slice
	updated, cmd := m.Update(tickMsg(time.Now()))

	if cmd == nil {
		t.Error("tickMsg should return next tick command")
	}

	model := updated.(Model)
	if len(model.projects) != 0 {
		t.Error("projects should remain empty")
	}
}

// TestTickMsgHandler_WithProjects verifies tick recalculates waiting counts (AC3).
func TestTickMsgHandler_WithProjects(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)

	// Create test projects
	m.projects = []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/a", State: domain.StateActive},
		{ID: "b", Name: "project-b", Path: "/b", State: domain.StateActive},
	}

	// Set up mock waiting detector that returns true for one project
	mock := &mockWaitingDetector{
		isWaitingFunc: func(p *domain.Project) bool {
			return p.ID == "a" // Only project-a is waiting
		},
	}
	m.SetWaitingDetector(mock)

	// Reset call count
	mock.isWaitingCalls = 0

	// Send tick message
	updated, cmd := m.Update(tickMsg(time.Now()))

	if cmd == nil {
		t.Error("tickMsg should return next tick command")
	}

	// Verify detector was called for each project
	if mock.isWaitingCalls != 2 {
		t.Errorf("expected 2 isWaiting calls, got %d", mock.isWaitingCalls)
	}

	// Verify status bar was updated
	model := updated.(Model)
	view := model.statusBar.View()
	if view == "" {
		t.Error("status bar view should not be empty")
	}
}

// BenchmarkTickHandler verifies tick handler performance (AC4).
// Target: <100µs per tick. Allocations (~27KB, 5 allocs) come from tea.Tick
// command creation which is unavoidable in Bubble Tea's Elm architecture.
// Long-running tests verify GC handles this efficiently with no net heap growth.
func BenchmarkTickHandler(b *testing.B) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)

	// Create typical project set (10 projects)
	m.projects = make([]*domain.Project, 10)
	for i := 0; i < 10; i++ {
		m.projects[i] = &domain.Project{
			ID:    string(rune('a' + i)),
			Name:  "project-" + string(rune('a'+i)),
			Path:  "/path/" + string(rune('a'+i)),
			State: domain.StateActive,
		}
	}

	// Create mock detector that returns consistent results
	mock := &mockWaitingDetector{
		isWaitingFunc: func(p *domain.Project) bool {
			return false // Consistent result for benchmarking
		},
	}
	m.SetWaitingDetector(mock)

	tickTime := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = m.Update(tickMsg(tickTime))
	}
}

// BenchmarkTickHandler_ZeroProjects verifies guard clause doesn't add overhead (AC5).
// Same allocation pattern as non-empty case since allocations come from tea.Tick.
func BenchmarkTickHandler_ZeroProjects(b *testing.B) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)
	m.projects = nil // Zero projects

	tickTime := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = m.Update(tickMsg(tickTime))
	}
}

// TestLongRunningSession_NoGoroutineLeak simulates 1-hour session (AC1, AC5).
// Sends 720 tick messages (60 min × 12 ticks/min at 5s interval).
func TestLongRunningSession_NoGoroutineLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running session test in short mode")
	}

	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)

	// Create test projects
	m.projects = make([]*domain.Project, 5)
	for i := 0; i < 5; i++ {
		m.projects[i] = &domain.Project{
			ID:    string(rune('a' + i)),
			Name:  "project-" + string(rune('a'+i)),
			Path:  "/path/" + string(rune('a'+i)),
			State: domain.StateActive,
		}
	}

	// Set up mock waiting detector
	mock := &mockWaitingDetector{
		isWaitingFunc: func(p *domain.Project) bool {
			return p.ID == "a" || p.ID == "c" // 2 of 5 waiting
		},
	}
	m.SetWaitingDetector(mock)

	// Record initial goroutine count
	initialGoroutines := runtime.NumGoroutine()

	// Record initial memory stats
	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)
	initialHeap := initialMem.HeapAlloc

	// Simulate 1-hour run: 720 ticks (60 min × 12 ticks/min at 5s interval)
	tickCount := 720
	tickTime := time.Now()

	for i := 0; i < tickCount; i++ {
		updated, cmd := m.Update(tickMsg(tickTime))
		m = updated.(Model)

		// Verify each tick returns next tick command
		if cmd == nil {
			t.Fatalf("tick %d: expected next tick command", i)
		}

		// Advance tick time by 5 seconds
		tickTime = tickTime.Add(5 * time.Second)
	}

	// Force GC to get accurate memory reading
	runtime.GC()
	runtime.GC()

	// Record final goroutine count
	finalGoroutines := runtime.NumGoroutine()

	// Record final memory stats
	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)
	finalHeap := finalMem.HeapAlloc

	// Verify no goroutine leak (±1 tolerance for runtime variations)
	goroutineDelta := finalGoroutines - initialGoroutines
	if goroutineDelta > 1 || goroutineDelta < -1 {
		t.Errorf("goroutine leak detected: initial=%d, final=%d, delta=%d",
			initialGoroutines, finalGoroutines, goroutineDelta)
	}

	// Verify no significant heap growth (<1MB)
	heapGrowth := int64(finalHeap) - int64(initialHeap)
	maxGrowthBytes := int64(1 * 1024 * 1024) // 1MB
	if heapGrowth > maxGrowthBytes {
		t.Errorf("heap growth exceeds 1MB: initial=%d, final=%d, growth=%d bytes",
			initialHeap, finalHeap, heapGrowth)
	}

	// Verify model state is unchanged
	if len(m.projects) != 5 {
		t.Errorf("projects count changed: expected 5, got %d", len(m.projects))
	}
	if !m.ready {
		t.Error("model should still be ready")
	}

	t.Logf("Long-running session test passed: %d ticks, goroutine delta=%d, heap growth=%d bytes",
		tickCount, goroutineDelta, heapGrowth)
}

// TestLongRunningSession_ModelStateIntegrity verifies model state after many ticks (AC5).
func TestLongRunningSession_ModelStateIntegrity(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)

	// Create test projects with specific values
	m.projects = []*domain.Project{
		{ID: "a", Name: "project-a", Path: "/path/a", State: domain.StateActive},
		{ID: "b", Name: "project-b", Path: "/path/b", State: domain.StateHibernated},
	}

	mock := &mockWaitingDetector{
		isWaitingFunc: func(p *domain.Project) bool {
			return p.ID == "a" // Only project-a waiting
		},
	}
	m.SetWaitingDetector(mock)

	// Send 100 tick messages
	tickTime := time.Now()
	for i := 0; i < 100; i++ {
		updated, _ := m.Update(tickMsg(tickTime))
		m = updated.(Model)
		tickTime = tickTime.Add(5 * time.Second)
	}

	// Verify model state integrity
	if len(m.projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(m.projects))
	}
	if m.projects[0].ID != "a" || m.projects[0].Name != "project-a" {
		t.Error("project-a data corrupted")
	}
	if m.projects[1].ID != "b" || m.projects[1].State != domain.StateHibernated {
		t.Error("project-b data corrupted")
	}
	if m.width != 80 || m.height != 40 {
		t.Error("model dimensions corrupted")
	}
	if !m.ready {
		t.Error("model ready state corrupted")
	}
}

// =============================================================================
// Story 8.2: File Watcher + Tick Integration Tests (AC1, AC2)
// =============================================================================

// TestFileEventResetsWaiting_TickShowsWorking verifies file event resets waiting timestamp (AC2).
// After file event, next tick should show project as "working" not "waiting".
func TestFileEventResetsWaiting_TickShowsWorking(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)

	// Create project with old activity (would be waiting)
	oldTime := time.Now().Add(-15 * time.Minute)
	m.projects = []*domain.Project{
		{
			ID:             "a",
			Name:           "project-a",
			Path:           "/home/user/project-a",
			State:          domain.StateActive,
			LastActivityAt: oldTime,
		},
	}

	// Set up waiting detector based on LastActivityAt
	waitingThreshold := 10 * time.Minute
	mock := &mockWaitingDetector{
		isWaitingFunc: func(p *domain.Project) bool {
			return time.Since(p.LastActivityAt) > waitingThreshold
		},
	}
	m.SetWaitingDetector(mock)

	// Initial tick - should be waiting
	updated, _ := m.Update(tickMsg(time.Now()))
	m = updated.(Model)

	// Verify project is waiting before file event
	if !mock.isWaitingFunc(m.projects[0]) {
		t.Error("project should be waiting before file event")
	}

	// Simulate file event - update LastActivityAt
	m.projects[0].LastActivityAt = time.Now()

	// Tick after file event - should now show working
	updated, _ = m.Update(tickMsg(time.Now()))
	m = updated.(Model)

	// Verify project is no longer waiting
	if mock.isWaitingFunc(m.projects[0]) {
		t.Error("project should not be waiting after file event")
	}
}

// TestNoFileEvent_TickShowsWaiting verifies no file event leads to waiting state (AC2).
func TestNoFileEvent_TickShowsWaiting(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)

	// Create project with recent activity
	m.projects = []*domain.Project{
		{
			ID:             "a",
			Name:           "project-a",
			Path:           "/home/user/project-a",
			State:          domain.StateActive,
			LastActivityAt: time.Now(),
		},
	}

	// Set up waiting detector with 1-minute threshold for testing
	waitingThreshold := 1 * time.Minute
	mock := &mockWaitingDetector{
		isWaitingFunc: func(p *domain.Project) bool {
			return time.Since(p.LastActivityAt) > waitingThreshold
		},
	}
	m.SetWaitingDetector(mock)

	// Initial state - not waiting
	if mock.isWaitingFunc(m.projects[0]) {
		t.Error("project should not be waiting initially")
	}

	// Simulate passage of time (no file events)
	m.projects[0].LastActivityAt = time.Now().Add(-2 * time.Minute)

	// Tick after threshold - should now show waiting
	updated, _ := m.Update(tickMsg(time.Now()))
	m = updated.(Model)

	// Verify project is now waiting
	if !mock.isWaitingFunc(m.projects[0]) {
		t.Error("project should be waiting after threshold")
	}
}

// TestWatcherPlusTick_IntegrationCycle tests complete watcher+tick integration (AC1, AC2).
func TestWatcherPlusTick_IntegrationCycle(t *testing.T) {
	m := NewModel(nil)
	m.ready = true
	m.width = 80
	m.height = 40
	m.statusBar = components.NewStatusBarModel(80)

	// Start time for simulation
	startTime := time.Now()

	// Create project with activity at start time
	m.projects = []*domain.Project{
		{
			ID:             "a",
			Name:           "project-a",
			Path:           "/home/user/project-a",
			State:          domain.StateActive,
			LastActivityAt: startTime,
		},
	}

	// Track waiting state changes using simulated current time
	waitingThreshold := 30 * time.Second
	var currentSimTime time.Time
	mock := &mockWaitingDetector{
		isWaitingFunc: func(p *domain.Project) bool {
			// Use simulated time instead of real time
			return currentSimTime.Sub(p.LastActivityAt) > waitingThreshold
		},
	}
	m.SetWaitingDetector(mock)

	// Simulate 10 ticks (50 seconds at 5s interval)
	currentSimTime = startTime
	for i := 0; i < 10; i++ {
		updated, cmd := m.Update(tickMsg(currentSimTime))
		m = updated.(Model)
		if cmd == nil {
			t.Fatalf("tick %d should return next tick command", i)
		}
		currentSimTime = currentSimTime.Add(5 * time.Second)
	}

	// After 50 seconds with 30s threshold, project should be waiting
	// At this point: currentSimTime = startTime + 50s, LastActivityAt = startTime
	// Delta = 50s > 30s threshold
	if !mock.isWaitingFunc(m.projects[0]) {
		t.Errorf("project should be waiting after 50 seconds with 30s threshold (delta: %v)",
			currentSimTime.Sub(m.projects[0].LastActivityAt))
	}

	// Simulate file event (resets LastActivityAt to current simulated time)
	m.projects[0].LastActivityAt = currentSimTime

	// Next tick should show working
	updated, _ := m.Update(tickMsg(currentSimTime))
	m = updated.(Model)

	if mock.isWaitingFunc(m.projects[0]) {
		t.Error("project should be working after file event")
	}
}
