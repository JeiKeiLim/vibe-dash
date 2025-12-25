package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// mockProjectRepo is a minimal mock for testing ActivityTracker
type mockProjectRepo struct {
	mu             sync.Mutex
	lastActivityAt map[string]time.Time // id -> timestamp
	err            error
}

func newMockProjectRepo() *mockProjectRepo {
	return &mockProjectRepo{
		lastActivityAt: make(map[string]time.Time),
	}
}

func (m *mockProjectRepo) UpdateLastActivity(ctx context.Context, id string, timestamp time.Time) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastActivityAt[id] = timestamp
	return nil
}

func (m *mockProjectRepo) getLastActivity(id string) (time.Time, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.lastActivityAt[id]
	return t, ok
}

// Unused interface methods - minimal implementation
func (m *mockProjectRepo) Save(context.Context, *domain.Project) error               { return nil }
func (m *mockProjectRepo) FindByID(context.Context, string) (*domain.Project, error) { return nil, nil }
func (m *mockProjectRepo) FindByPath(context.Context, string) (*domain.Project, error) {
	return nil, nil
}
func (m *mockProjectRepo) FindAll(context.Context) ([]*domain.Project, error)             { return nil, nil }
func (m *mockProjectRepo) FindActive(context.Context) ([]*domain.Project, error)          { return nil, nil }
func (m *mockProjectRepo) FindHibernated(context.Context) ([]*domain.Project, error)      { return nil, nil }
func (m *mockProjectRepo) Delete(context.Context, string) error                           { return nil }
func (m *mockProjectRepo) UpdateState(context.Context, string, domain.ProjectState) error { return nil }
func (m *mockProjectRepo) ResetProject(context.Context, string) error                     { return nil }
func (m *mockProjectRepo) ResetAll(context.Context) (int, error)                          { return 0, nil }

func TestNewActivityTracker(t *testing.T) {
	repo := newMockProjectRepo()
	tracker := NewActivityTracker(repo)

	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
	if tracker.repo != repo {
		t.Error("expected tracker to store repo")
	}
}

func TestActivityTracker_SetProjects(t *testing.T) {
	tracker := NewActivityTracker(newMockProjectRepo())

	projects := []*domain.Project{
		{ID: "proj1", Path: "/path/to/project1"},
		{ID: "proj2", Path: "/path/to/project2/"},
	}

	tracker.SetProjects(projects)

	// Verify path normalization (trailing slash removed)
	if _, ok := tracker.projects["/path/to/project1"]; !ok {
		t.Error("expected project1 to be in cache")
	}
	if _, ok := tracker.projects["/path/to/project2"]; !ok {
		t.Error("expected project2 to be in cache (normalized without trailing slash)")
	}
	if _, ok := tracker.projects["/path/to/project2/"]; ok {
		t.Error("expected trailing slash to be removed")
	}
}

func TestActivityTracker_ProcessEvents_BasicEvent(t *testing.T) {
	repo := newMockProjectRepo()
	tracker := NewActivityTracker(repo)

	project := &domain.Project{ID: "proj1", Path: "/path/to/project1"}
	tracker.SetProjects([]*domain.Project{project})

	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan ports.FileEvent, 1)

	eventTime := time.Now()
	events <- ports.FileEvent{
		Path:      "/path/to/project1/src/main.go",
		Operation: ports.FileOpModify,
		Timestamp: eventTime,
	}
	close(events)

	// ProcessEvents should consume the event and exit when channel closes
	tracker.ProcessEvents(ctx, events)
	cancel() // cleanup

	// Verify LastActivityAt was updated
	if got, ok := repo.getLastActivity("proj1"); !ok {
		t.Error("expected last activity to be updated")
	} else if !got.Equal(eventTime) {
		t.Errorf("expected timestamp %v, got %v", eventTime, got)
	}
}

func TestActivityTracker_ProcessEvents_ContextCancellation(t *testing.T) {
	tracker := NewActivityTracker(newMockProjectRepo())
	tracker.SetProjects([]*domain.Project{{ID: "proj1", Path: "/path/to/project1"}})

	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan ports.FileEvent)

	done := make(chan struct{})
	go func() {
		tracker.ProcessEvents(ctx, events)
		close(done)
	}()

	// Cancel context
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Error("ProcessEvents did not stop on context cancellation")
	}
}

func TestActivityTracker_ProcessEvents_ChannelClose(t *testing.T) {
	tracker := NewActivityTracker(newMockProjectRepo())

	ctx := context.Background()
	events := make(chan ports.FileEvent)

	done := make(chan struct{})
	go func() {
		tracker.ProcessEvents(ctx, events)
		close(done)
	}()

	// Close channel
	close(events)

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Error("ProcessEvents did not stop on channel close")
	}
}

func TestActivityTracker_PathToProjectMatching(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		eventPath   string
		shouldMatch bool
	}{
		{
			name:        "exact match",
			projectPath: "/path/to/project",
			eventPath:   "/path/to/project",
			shouldMatch: true,
		},
		{
			name:        "file in project root",
			projectPath: "/path/to/project",
			eventPath:   "/path/to/project/main.go",
			shouldMatch: true,
		},
		{
			name:        "file in nested subdirectory",
			projectPath: "/path/to/project",
			eventPath:   "/path/to/project/src/pkg/utils/helper.go",
			shouldMatch: true,
		},
		{
			name:        "different project with similar prefix",
			projectPath: "/path/to/project",
			eventPath:   "/path/to/project2/main.go",
			shouldMatch: false,
		},
		{
			name:        "parent directory event",
			projectPath: "/path/to/project",
			eventPath:   "/path/to/main.go",
			shouldMatch: false,
		},
		{
			name:        "completely unrelated path",
			projectPath: "/path/to/project",
			eventPath:   "/other/path/main.go",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewActivityTracker(newMockProjectRepo())
			tracker.SetProjects([]*domain.Project{{ID: "proj1", Path: tt.projectPath}})

			result := tracker.findProjectForPath(tt.eventPath)

			if tt.shouldMatch && result == nil {
				t.Error("expected path to match project")
			}
			if !tt.shouldMatch && result != nil {
				t.Error("expected path NOT to match project")
			}
		})
	}
}

func TestActivityTracker_UnknownPath_GracefulDegradation(t *testing.T) {
	repo := newMockProjectRepo()
	tracker := NewActivityTracker(repo)
	tracker.SetProjects([]*domain.Project{{ID: "proj1", Path: "/known/project"}})

	ctx := context.Background()
	events := make(chan ports.FileEvent, 1)

	// Event for unknown path
	events <- ports.FileEvent{
		Path:      "/unknown/path/file.go",
		Operation: ports.FileOpModify,
		Timestamp: time.Now(),
	}
	close(events)

	// Should not panic, should gracefully skip
	tracker.ProcessEvents(ctx, events)

	// No activity should be recorded
	if _, ok := repo.getLastActivity("proj1"); ok {
		t.Error("expected no activity update for unmatched path")
	}
}

func TestActivityTracker_MultipleProjects(t *testing.T) {
	repo := newMockProjectRepo()
	tracker := NewActivityTracker(repo)

	projects := []*domain.Project{
		{ID: "proj1", Path: "/projects/frontend"},
		{ID: "proj2", Path: "/projects/backend"},
		{ID: "proj3", Path: "/other/location"},
	}
	tracker.SetProjects(projects)

	ctx := context.Background()
	events := make(chan ports.FileEvent, 3)

	event1Time := time.Now()
	event2Time := event1Time.Add(time.Second)
	event3Time := event2Time.Add(time.Second)

	events <- ports.FileEvent{Path: "/projects/frontend/src/app.tsx", Timestamp: event1Time}
	events <- ports.FileEvent{Path: "/projects/backend/cmd/main.go", Timestamp: event2Time}
	events <- ports.FileEvent{Path: "/other/location/README.md", Timestamp: event3Time}
	close(events)

	tracker.ProcessEvents(ctx, events)

	// Verify each project got the right timestamp
	if got, ok := repo.getLastActivity("proj1"); !ok || !got.Equal(event1Time) {
		t.Errorf("proj1: expected %v, got %v (ok=%v)", event1Time, got, ok)
	}
	if got, ok := repo.getLastActivity("proj2"); !ok || !got.Equal(event2Time) {
		t.Errorf("proj2: expected %v, got %v (ok=%v)", event2Time, got, ok)
	}
	if got, ok := repo.getLastActivity("proj3"); !ok || !got.Equal(event3Time) {
		t.Errorf("proj3: expected %v, got %v (ok=%v)", event3Time, got, ok)
	}
}

func TestActivityTracker_EmptyProjectCache(t *testing.T) {
	repo := newMockProjectRepo()
	tracker := NewActivityTracker(repo)
	// Don't set any projects

	ctx := context.Background()
	events := make(chan ports.FileEvent, 1)

	events <- ports.FileEvent{
		Path:      "/some/path/file.go",
		Operation: ports.FileOpModify,
		Timestamp: time.Now(),
	}
	close(events)

	// Should not panic with empty cache
	tracker.ProcessEvents(ctx, events)
}

func TestActivityTracker_ConcurrentSetProjects(t *testing.T) {
	tracker := NewActivityTracker(newMockProjectRepo())

	projects1 := []*domain.Project{{ID: "proj1", Path: "/path1"}}
	projects2 := []*domain.Project{{ID: "proj2", Path: "/path2"}}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			tracker.SetProjects(projects1)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			tracker.SetProjects(projects2)
		}
	}()

	wg.Wait() // Should not race
}

func TestActivityTracker_RepositoryError_GracefulHandling(t *testing.T) {
	// Test that repository errors are logged but don't crash the tracker (M2 fix)
	repo := newMockProjectRepo()
	repo.err = domain.ErrProjectNotFound // Simulate repo error
	tracker := NewActivityTracker(repo)

	project := &domain.Project{ID: "proj1", Path: "/path/to/project"}
	tracker.SetProjects([]*domain.Project{project})

	ctx := context.Background()
	events := make(chan ports.FileEvent, 1)

	events <- ports.FileEvent{
		Path:      "/path/to/project/src/main.go",
		Operation: ports.FileOpModify,
		Timestamp: time.Now(),
	}
	close(events)

	// Should not panic, should gracefully handle error
	tracker.ProcessEvents(ctx, events)

	// Activity should not be recorded due to error
	if _, ok := repo.getLastActivity("proj1"); ok {
		t.Error("expected no activity update when repo returns error")
	}
}
