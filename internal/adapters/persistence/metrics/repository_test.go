package metrics

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewMetricsRepository(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")

	repo := NewMetricsRepository(dbPath)

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
	if repo.dbPath != dbPath {
		t.Errorf("expected dbPath %q, got %q", dbPath, repo.dbPath)
	}
	// schemaOnce is zero value (not yet called) - verified by successful transition recording below
}

func TestRecordTransition_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)

	err := repo.RecordTransition(context.Background(), "proj-123", "plan", "tasks")

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("metrics.db was not created")
	}
}

func TestRecordTransition_SchemaCreatedOnFirstAccess(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)

	// Record a transition (triggers schema creation via sync.Once)
	_ = repo.RecordTransition(context.Background(), "proj-123", "plan", "tasks")

	// Verify tables exist by opening DB and querying
	db, err := repo.openDB(context.Background())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Check schema_version table exists
	var version int
	err = db.Get(&version, "SELECT version FROM schema_version LIMIT 1")
	if err != nil {
		t.Errorf("schema_version table not found: %v", err)
	}
	if version != SchemaVersion {
		t.Errorf("expected schema version %d, got %d", SchemaVersion, version)
	}

	// Check stage_transitions table exists by counting rows
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM stage_transitions")
	if err != nil {
		t.Errorf("stage_transitions table not found: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 transition, got %d", count)
	}
}

func TestRecordTransition_GracefulDegradation(t *testing.T) {
	// Use invalid path to simulate failure
	repo := NewMetricsRepository("/nonexistent/dir/metrics.db")

	err := repo.RecordTransition(context.Background(), "proj-123", "plan", "tasks")

	// Should return nil (graceful degradation), not error
	if err != nil {
		t.Errorf("expected nil (graceful degradation), got %v", err)
	}
}

func TestRecordTransition_GracefulDegradation_ReadOnlyDir(t *testing.T) {
	// Create a read-only directory
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}

	repo := NewMetricsRepository(filepath.Join(readOnlyDir, "metrics.db"))

	err := repo.RecordTransition(context.Background(), "proj-123", "plan", "tasks")

	// Should return nil (graceful degradation)
	if err != nil {
		t.Errorf("expected nil (graceful degradation), got %v", err)
	}
}

func TestRecordTransition_MultipleTransitions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record multiple transitions
	transitions := []struct {
		projectID string
		fromStage string
		toStage   string
	}{
		{"proj-1", "", "plan"},
		{"proj-1", "plan", "tasks"},
		{"proj-1", "tasks", "code"},
		{"proj-2", "", "plan"},
		{"proj-2", "plan", "code"},
	}

	for _, tr := range transitions {
		if err := repo.RecordTransition(ctx, tr.projectID, tr.fromStage, tr.toStage); err != nil {
			t.Errorf("RecordTransition failed: %v", err)
		}
	}

	// Verify all were recorded
	db, err := repo.openDB(ctx)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM stage_transitions")
	if err != nil {
		t.Fatalf("failed to count transitions: %v", err)
	}
	if count != len(transitions) {
		t.Errorf("expected %d transitions, got %d", len(transitions), count)
	}
}

func TestWALMode(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Trigger schema creation
	_ = repo.RecordTransition(ctx, "proj-1", "plan", "tasks")

	// Check WAL mode
	db, err := repo.openDB(ctx)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	var journalMode string
	if err := db.Get(&journalMode, "PRAGMA journal_mode;"); err != nil {
		t.Fatalf("failed to get journal_mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("expected journal_mode 'wal', got %q", journalMode)
	}
}

func TestGenerateUUID(t *testing.T) {
	// Generate multiple UUIDs and verify format
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		uuid := generateUUID()

		// Check format: 8-4-4-4-12 hex characters
		if len(uuid) != 36 {
			t.Errorf("UUID length should be 36, got %d: %s", len(uuid), uuid)
		}

		// Check uniqueness
		if seen[uuid] {
			t.Errorf("duplicate UUID generated: %s", uuid)
		}
		seen[uuid] = true

		// Check version bit (should be 4)
		if uuid[14] != '4' {
			t.Errorf("UUID version should be 4, got %c in %s", uuid[14], uuid)
		}

		// Check variant bits (should be 8, 9, a, or b)
		variant := uuid[19]
		if variant != '8' && variant != '9' && variant != 'a' && variant != 'b' {
			t.Errorf("UUID variant should be 8/9/a/b, got %c in %s", variant, uuid)
		}
	}
}

func TestRecordTransition_Timestamps(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record a transition
	_ = repo.RecordTransition(ctx, "proj-1", "plan", "tasks")

	// Query the recorded transition
	db, err := repo.openDB(ctx)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	var row stageTransitionRow
	err = db.Get(&row, "SELECT * FROM stage_transitions LIMIT 1")
	if err != nil {
		t.Fatalf("failed to query transition: %v", err)
	}

	// Verify fields
	if row.ProjectID != "proj-1" {
		t.Errorf("expected project_id 'proj-1', got %q", row.ProjectID)
	}
	if row.FromStage != "plan" {
		t.Errorf("expected from_stage 'plan', got %q", row.FromStage)
	}
	if row.ToStage != "tasks" {
		t.Errorf("expected to_stage 'tasks', got %q", row.ToStage)
	}
	if len(row.ID) != 36 {
		t.Errorf("expected UUID (36 chars), got %q", row.ID)
	}
	if len(row.TransitionedAt) == 0 {
		t.Error("transitioned_at should not be empty")
	}
}

func TestRecordTransition_EmptyFromStage(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record transition with empty from_stage (first detection)
	err := repo.RecordTransition(ctx, "proj-1", "", "plan")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// Verify it was recorded correctly
	db, err := repo.openDB(ctx)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	var row stageTransitionRow
	err = db.Get(&row, "SELECT * FROM stage_transitions WHERE project_id = ?", "proj-1")
	if err != nil {
		t.Fatalf("failed to query transition: %v", err)
	}

	if row.FromStage != "" {
		t.Errorf("expected empty from_stage, got %q", row.FromStage)
	}
}

func TestIndexes(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Trigger schema creation
	_ = repo.RecordTransition(ctx, "proj-1", "plan", "tasks")

	db, err := repo.openDB(ctx)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Query indexes
	var indexes []struct {
		Name string `db:"name"`
	}
	err = db.Select(&indexes, "SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='stage_transitions'")
	if err != nil {
		t.Fatalf("failed to query indexes: %v", err)
	}

	// Check expected indexes exist
	indexNames := make(map[string]bool)
	for _, idx := range indexes {
		indexNames[idx.Name] = true
	}

	if !indexNames["idx_stage_transitions_project"] {
		t.Error("expected idx_stage_transitions_project index to exist")
	}
	if !indexNames["idx_stage_transitions_time"] {
		t.Error("expected idx_stage_transitions_time index to exist")
	}
}

// Story 16.4: Tests for GetTransitionsByProject with time filter

func TestGetTransitionsByProject_EmptyResult(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record a transition to initialize the DB
	_ = repo.RecordTransition(ctx, "other-proj", "plan", "tasks")

	// Query a different project
	since := time.Now().Add(-24 * time.Hour)
	result := repo.GetTransitionsByProject(ctx, "nonexistent-proj", since)

	if result == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 transitions, got %d", len(result))
	}
}

func TestGetTransitionsByProject_WithTransitions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record multiple transitions for a project
	_ = repo.RecordTransition(ctx, "proj-1", "", "plan")
	_ = repo.RecordTransition(ctx, "proj-1", "plan", "tasks")
	_ = repo.RecordTransition(ctx, "proj-1", "tasks", "code")
	_ = repo.RecordTransition(ctx, "proj-2", "", "plan") // Different project

	// Query proj-1 transitions
	since := time.Now().Add(-24 * time.Hour)
	result := repo.GetTransitionsByProject(ctx, "proj-1", since)

	if len(result) != 3 {
		t.Fatalf("expected 3 transitions for proj-1, got %d", len(result))
	}

	// Verify results are in chronological order (ASC)
	for i, tr := range result {
		if tr.ProjectID != "proj-1" {
			t.Errorf("transition %d: expected project_id 'proj-1', got %q", i, tr.ProjectID)
		}
	}

	// First transition should be ""->plan (earliest)
	if result[0].FromStage != "" || result[0].ToStage != "plan" {
		t.Errorf("first transition should be '' -> 'plan', got %q -> %q",
			result[0].FromStage, result[0].ToStage)
	}
}

func TestGetTransitionsByProject_TimeFilter(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record a transition
	_ = repo.RecordTransition(ctx, "proj-1", "plan", "tasks")

	// Query with a future time - should find nothing
	futureTime := time.Now().Add(1 * time.Hour)
	result := repo.GetTransitionsByProject(ctx, "proj-1", futureTime)

	if len(result) != 0 {
		t.Errorf("expected 0 transitions with future since time, got %d", len(result))
	}

	// Query with past time - should find the transition
	pastTime := time.Now().Add(-1 * time.Hour)
	result = repo.GetTransitionsByProject(ctx, "proj-1", pastTime)

	if len(result) != 1 {
		t.Errorf("expected 1 transition with past since time, got %d", len(result))
	}
}

func TestGetTransitionsByProject_GracefulDegradation(t *testing.T) {
	// Use invalid path to simulate failure
	repo := NewMetricsRepository("/nonexistent/dir/metrics.db")
	ctx := context.Background()

	result := repo.GetTransitionsByProject(ctx, "proj-1", time.Now().Add(-24*time.Hour))

	// Should return nil (graceful degradation)
	if result != nil {
		t.Errorf("expected nil (graceful degradation), got %v", result)
	}
}

func TestGetTransitionsByTimeRange_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record transitions
	_ = repo.RecordTransition(ctx, "proj-1", "", "plan")
	_ = repo.RecordTransition(ctx, "proj-2", "", "plan")
	_ = repo.RecordTransition(ctx, "proj-3", "", "plan")

	// Query all recent transitions
	from := time.Now().Add(-1 * time.Hour)
	to := time.Now().Add(1 * time.Hour)
	result := repo.GetTransitionsByTimeRange(ctx, from, to)

	if len(result) != 3 {
		t.Errorf("expected 3 transitions, got %d", len(result))
	}
}

func TestGetTransitionsByTimeRange_EmptyRange(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record a transition
	_ = repo.RecordTransition(ctx, "proj-1", "", "plan")

	// Query with future range - should find nothing
	from := time.Now().Add(1 * time.Hour)
	to := time.Now().Add(2 * time.Hour)
	result := repo.GetTransitionsByTimeRange(ctx, from, to)

	if len(result) != 0 {
		t.Errorf("expected 0 transitions in future range, got %d", len(result))
	}
}

func TestGetTransitionsByTimeRange_GracefulDegradation(t *testing.T) {
	repo := NewMetricsRepository("/nonexistent/dir/metrics.db")
	ctx := context.Background()

	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()
	result := repo.GetTransitionsByTimeRange(ctx, from, to)

	// Should return nil (graceful degradation)
	if result != nil {
		t.Errorf("expected nil (graceful degradation), got %v", result)
	}
}

func TestRowToTransition(t *testing.T) {
	row := &stageTransitionRow{
		ID:             "test-uuid-123",
		ProjectID:      "proj-1",
		FromStage:      "plan",
		ToStage:        "tasks",
		TransitionedAt: "2026-01-16T10:30:00.123456789Z",
	}

	result := rowToTransition(row)

	if result.ID != "test-uuid-123" {
		t.Errorf("expected ID 'test-uuid-123', got %q", result.ID)
	}
	if result.ProjectID != "proj-1" {
		t.Errorf("expected ProjectID 'proj-1', got %q", result.ProjectID)
	}
	if result.FromStage != "plan" {
		t.Errorf("expected FromStage 'plan', got %q", result.FromStage)
	}
	if result.ToStage != "tasks" {
		t.Errorf("expected ToStage 'tasks', got %q", result.ToStage)
	}
	if result.TransitionedAt.IsZero() {
		t.Error("expected non-zero TransitionedAt")
	}
	if result.TransitionedAt.Year() != 2026 {
		t.Errorf("expected year 2026, got %d", result.TransitionedAt.Year())
	}
}

// Story 16.4: Tests for GetTransitionTimestamps (statsview.MetricsReader interface)

func TestGetTransitionTimestamps_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record multiple transitions
	_ = repo.RecordTransition(ctx, "proj-1", "", "plan")
	_ = repo.RecordTransition(ctx, "proj-1", "plan", "tasks")
	_ = repo.RecordTransition(ctx, "proj-1", "tasks", "code")

	// Query via the interface method
	since := time.Now().Add(-1 * time.Hour)
	result := repo.GetTransitionTimestamps(ctx, "proj-1", since)

	if len(result) != 3 {
		t.Fatalf("expected 3 transitions, got %d", len(result))
	}

	// Verify each result has a valid timestamp
	for i, tr := range result {
		if tr.TransitionedAt.IsZero() {
			t.Errorf("transition %d: expected non-zero TransitionedAt", i)
		}
	}
}

func TestGetTransitionTimestamps_GracefulDegradation(t *testing.T) {
	repo := NewMetricsRepository("/nonexistent/dir/metrics.db")
	ctx := context.Background()

	result := repo.GetTransitionTimestamps(ctx, "proj-1", time.Now().Add(-24*time.Hour))

	// Should return empty slice (graceful degradation) - not nil
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestGetTransitionTimestamps_EmptyResult(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Initialize DB but don't record anything for this project
	_ = repo.RecordTransition(ctx, "other-proj", "", "plan")

	since := time.Now().Add(-1 * time.Hour)
	result := repo.GetTransitionTimestamps(ctx, "nonexistent", since)

	if len(result) != 0 {
		t.Errorf("expected 0 transitions for nonexistent project, got %d", len(result))
	}
}
