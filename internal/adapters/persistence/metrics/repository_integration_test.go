//go:build integration

package metrics_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/metrics"
)

func TestRecordTransition_Integration(t *testing.T) {
	// Use actual ~/.vibe-dash location for integration test
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	testDir := filepath.Join(homeDir, ".vibe-dash-test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	dbPath := filepath.Join(testDir, "metrics.db")
	repo := metrics.NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record transitions
	err = repo.RecordTransition(ctx, "test-project-1", "", "plan")
	if err != nil {
		t.Errorf("first transition failed: %v", err)
	}

	err = repo.RecordTransition(ctx, "test-project-1", "plan", "tasks")
	if err != nil {
		t.Errorf("second transition failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("metrics.db was not created")
	}

	// Verify WAL files exist
	walPath := dbPath + "-wal"
	shmPath := dbPath + "-shm"
	time.Sleep(100 * time.Millisecond) // Give SQLite time to create WAL files

	if _, err := os.Stat(walPath); os.IsNotExist(err) {
		// WAL file might not exist if no writes are pending, that's OK
		t.Logf("WAL file not found (may be checkpointed): %s", walPath)
	}
	if _, err := os.Stat(shmPath); os.IsNotExist(err) {
		t.Logf("SHM file not found (may be checkpointed): %s", shmPath)
	}
}

func TestRecordTransition_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := metrics.NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Run concurrent writes
	var wg sync.WaitGroup
	errors := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			projectID := "proj-concurrent"
			fromStage := "stage-" + string(rune('a'+i%5))
			toStage := "stage-" + string(rune('a'+(i+1)%5))
			if err := repo.RecordTransition(ctx, projectID, fromStage, toStage); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("concurrent write failed: %v", err)
	}
}

func TestRecordTransition_LargeVolume(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large volume test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "metrics.db")
	repo := metrics.NewMetricsRepository(dbPath)
	ctx := context.Background()

	// Record 1000 transitions
	for i := 0; i < 1000; i++ {
		if err := repo.RecordTransition(ctx, "proj-volume", "from", "to"); err != nil {
			t.Fatalf("transition %d failed: %v", i, err)
		}
	}

	// Verify file size is reasonable (should be well under 1MB for 1000 records)
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("failed to stat db: %v", err)
	}

	// 1000 records * ~100 bytes = ~100KB, add overhead = should be < 500KB
	if info.Size() > 500*1024 {
		t.Errorf("database size %d exceeds expected maximum", info.Size())
	}
}
