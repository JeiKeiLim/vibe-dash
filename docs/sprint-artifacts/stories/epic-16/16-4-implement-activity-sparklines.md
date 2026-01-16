# Story 16.4: Implement Activity Sparklines

Status: done

## Story

As a user,
I want to see activity sparklines per project,
So that I can quickly visualize work patterns.

## User-Visible Changes

- **New:** Stats View displays activity sparklines for each project (e.g., `▁▂▃▅▂▁▇▅▂`)
- **New:** Sparklines represent stage transition activity density per time bucket
- **New:** Projects with no recent activity show flat or empty sparklines

## Acceptance Criteria

1. **Given** project has stage transitions over past 30 days
   **When** Stats View renders
   **Then** shows sparkline character graph (e.g., `▁▂▃▅▂▁▇▅▂`)

2. **And** sparkline represents activity density per day/week
   - Higher bars = more stage transitions in that time bucket
   - Uses 8 Unicode block characters: `▁▂▃▄▅▆▇█`

3. **And** projects with no recent activity show flat sparkline (`▁▁▁▁▁▁▁`)

4. **And** sparkline width adapts to available terminal width (7-14 buckets)

5. **And** Stats View continues to work when metrics.db is unavailable (graceful degradation)

## Tasks / Subtasks

- [x] Task 1: Add repository query methods for transitions (AC: #1)
  - [x] In `internal/adapters/persistence/metrics/queries.go`, add time-filtered query:
    ```go
    // selectByProjectWithTimeSQL retrieves transitions for a project since a given time
    const selectByProjectWithTimeSQL = `
    SELECT ` + transitionColumns + ` FROM stage_transitions
    WHERE project_id = ? AND transitioned_at >= ?
    ORDER BY transitioned_at ASC`
    ```
  - [x] In `internal/adapters/persistence/metrics/helpers.go`, add public domain type (distinct from internal `stageTransitionRow`):
    ```go
    // StageTransition is the public domain type for stage transition data.
    // Distinct from internal stageTransitionRow - this is the exported API type.
    type StageTransition struct {
        ID             string
        ProjectID      string
        FromStage      string
        ToStage        string
        TransitionedAt time.Time
    }

    // rowToTransition converts internal row to public domain type.
    func rowToTransition(row *stageTransitionRow) StageTransition {
        t, _ := time.Parse(time.RFC3339Nano, row.TransitionedAt)
        return StageTransition{
            ID:             row.ID,
            ProjectID:      row.ProjectID,
            FromStage:      row.FromStage,
            ToStage:        row.ToStage,
            TransitionedAt: t,
        }
    }
    ```
  - [x] In `internal/adapters/persistence/metrics/repository.go`, add:
    ```go
    // GetTransitionsByProject retrieves transitions for a project since the given time.
    // Returns empty slice on any error (graceful degradation).
    func (r *MetricsRepository) GetTransitionsByProject(ctx context.Context, projectID string, since time.Time) []StageTransition

    // GetTransitionsByTimeRange retrieves transitions within a time range.
    // Returns empty slice on any error (graceful degradation).
    func (r *MetricsRepository) GetTransitionsByTimeRange(ctx context.Context, from, to time.Time) []StageTransition
    ```
  - [x] Use `selectByProjectWithTimeSQL` for time-filtered project queries

- [x] Task 2: Create sparkline rendering module (AC: #2)
  - [x] Create `internal/adapters/tui/statsview/sparkline.go`:
    ```go
    package statsview

    // SparklineChars are Unicode block elements for sparkline visualization (ascending height)
    var SparklineChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

    // RenderSparkline generates a sparkline string from activity counts.
    // Input: counts slice (one value per time bucket)
    // Output: sparkline string of len(counts) characters
    func RenderSparkline(counts []int) string
    ```
  - [x] Normalize counts to 0-7 range (8 levels for 8 characters)
  - [x] Handle empty input: return empty string
  - [x] Handle all-zero input: return flat sparkline (`▁▁▁...`)
  - [x] Handle single-value: map to middle char

- [x] Task 3: Create activity bucketing module (AC: #1, #2)
  - [x] Create `internal/adapters/tui/statsview/activity.go`:
    ```go
    package statsview

    import "time"

    // BucketActivityCounts calculates activity per time bucket.
    // timestamps: transition times (oldest first expected, but handles any order)
    // buckets: number of time buckets (e.g., 7 for weekly)
    // timeRange: total time period to cover
    // now: reference time for bucket calculation (enables testing)
    // Returns: slice of counts, one per bucket (oldest first)
    func BucketActivityCounts(timestamps []time.Time, buckets int, timeRange time.Duration, now time.Time) []int
    ```
  - [x] Default: 7 buckets over 30 days (4.3 days per bucket)
  - [x] Handle empty timestamps: return all zeros
  - [x] Bucket index calculation: `idx = int((now - timestamp) / bucketDuration)` clamped to [0, buckets-1]

- [x] Task 4: Add metrics repository interface to TUI (AC: #1, #5)
  - [x] Create interface in `internal/adapters/tui/statsview/interfaces.go`:
    ```go
    package statsview

    import (
        "context"
        "time"
    )

    // Transition is a simplified type for sparkline calculations.
    // Contains only the timestamp needed for activity bucketing.
    // Avoids importing metrics package types into TUI layer.
    type Transition struct {
        TransitionedAt time.Time
    }

    // MetricsReader is the interface for reading metrics data.
    // TUI depends on interface, not concrete repository.
    // Returns []Transition to avoid cross-adapter coupling.
    type MetricsReader interface {
        // GetTransitionTimestamps retrieves transition times for a project since given time.
        // Returns empty slice on any error (graceful degradation).
        GetTransitionTimestamps(ctx context.Context, projectID string, since time.Time) []Transition
    }
    ```
  - [x] In metrics repository, add adapter method that returns `[]statsview.Transition`

- [x] Task 5: Implement Stats View data fetching (AC: #1, #5)
  - [x] In `internal/adapters/tui/model.go`, add:
    ```go
    // Story 16.4: Metrics reader for stats view (optional, graceful nil)
    metricsReader statsview.MetricsReader
    ```
  - [x] Add setter: `func (m *Model) SetMetricsReader(reader statsview.MetricsReader)`
  - [x] Add method to fetch project activity:
    ```go
    func (m Model) getProjectActivity(projectID string, buckets int) []int {
        if m.metricsReader == nil {
            return nil // Graceful degradation
        }
        // Get transitions for last 30 days
        ctx := context.Background()
        now := time.Now()
        since := now.Add(-30 * 24 * time.Hour)
        transitions := m.metricsReader.GetTransitionTimestamps(ctx, projectID, since)
        if len(transitions) == 0 {
            return nil
        }
        // Extract timestamps for bucketing
        timestamps := make([]time.Time, len(transitions))
        for i, t := range transitions {
            timestamps[i] = t.TransitionedAt
        }
        return statsview.BucketActivityCounts(timestamps, buckets, 30*24*time.Hour, now)
    }
    ```

- [x] Task 6: Update Stats View rendering (AC: #1, #3, #4)
  - [x] Update `internal/adapters/tui/statsview.go`:
    - Fetch transitions for all active projects
    - Generate sparkline per project
    - Layout: `[Project Name] [Sparkline] [Current Stage]`
  - [x] Handle narrow terminals: reduce bucket count (min 7)
  - [x] Handle wide terminals: max 14 buckets

- [x] Task 7: Wire MetricsRepository to TUI (AC: #1)
  - [x] In `cmd/vdash/main.go`, add after DirectoryManager initialization (~line 115):
    ```go
    // Story 16.4: Wire metrics repository to TUI for sparklines
    metricsDBPath := filepath.Join(dm.BaseDir(), "metrics.db")
    metricsRepo := metrics.NewMetricsRepository(metricsDBPath)
    ```
  - [x] In TUI model creation, set the reader:
    ```go
    // After model initialization
    model.SetMetricsReader(metricsRepo)
    ```
  - [x] Note: MetricsRepository already handles graceful degradation internally - no need to check if metrics.db exists

- [x] Task 8: Write comprehensive unit tests
  - [x] `internal/adapters/tui/statsview/sparkline_test.go`:
    ```go
    func TestRenderSparkline_EmptyInput(t *testing.T) { ... }
    func TestRenderSparkline_AllZeros(t *testing.T) { ... }
    func TestRenderSparkline_SingleValue(t *testing.T) { ... }
    func TestRenderSparkline_IncreasingValues(t *testing.T) { ... }
    func TestRenderSparkline_MaxValues(t *testing.T) { ... }
    ```
  - [x] `internal/adapters/tui/statsview/activity_test.go`:
    ```go
    func TestBucketActivityCounts_NoTimestamps(t *testing.T) { ... }
    func TestBucketActivityCounts_EvenDistribution(t *testing.T) { ... }
    func TestBucketActivityCounts_AllInOneBucket(t *testing.T) { ... }
    func TestBucketActivityCounts_TimestampsOutOfRange(t *testing.T) { ... }
    ```
  - [x] `internal/adapters/persistence/metrics/repository_test.go`:
    - Add tests for `GetTransitionsByProject` with time filter
    - Add tests for `GetTransitionsByTimeRange`
    - Add tests for `GetTransitionTimestamps` adapter method

- [x] Task 9: Update statsview_test.go for sparkline integration
  - [x] Test Stats View renders sparklines correctly
  - [x] Test graceful degradation when metricsReader is nil

## Dev Notes

### Architecture Alignment

Implements **FR-P2-16** from Phase 2 PRD: "System can display activity sparklines per project"

**Isolation Principle:** Stats View and sparkline rendering remain independent:
- Separate `statsview/` package for new modules
- Interface-based dependency (`MetricsReader`) enables testing
- Nil-safe: Stats View works without metrics data

**Import Direction (CRITICAL - Prevents Architecture Violation):**
```
cmd/vdash/main.go
    └── imports → internal/adapters/persistence/metrics (creates MetricsRepository)
    └── imports → internal/adapters/tui (creates Model)
    └── imports → internal/adapters/tui/statsview (for interface type)

internal/adapters/persistence/metrics
    └── imports → internal/adapters/tui/statsview (for Transition type in adapter method)

internal/adapters/tui
    └── imports → internal/adapters/tui/statsview (for interface and types)

internal/adapters/tui/statsview
    └── NO imports from persistence layer (clean separation)
```

The `statsview.Transition` type is defined in the TUI layer, and the metrics repository's `GetTransitionTimestamps` method adapts internal types to this TUI-layer type. This prevents cross-adapter coupling.

### File Modifications

| File | Changes |
|------|---------|
| `internal/adapters/persistence/metrics/queries.go` | Add `selectByProjectWithTimeSQL` query constant |
| `internal/adapters/persistence/metrics/repository.go` | Add `GetTransitionsByProject`, `GetTransitionsByTimeRange`, `GetTransitionTimestamps` methods |
| `internal/adapters/persistence/metrics/helpers.go` | Add `StageTransition` struct, `rowToTransition` helper |
| `internal/adapters/tui/statsview/sparkline.go` | **NEW:** Sparkline rendering |
| `internal/adapters/tui/statsview/activity.go` | **NEW:** Activity bucketing |
| `internal/adapters/tui/statsview/interfaces.go` | **NEW:** `Transition` type, `MetricsReader` interface |
| `internal/adapters/tui/statsview.go` | Update rendering with sparklines |
| `internal/adapters/tui/model.go` | Add `metricsReader` field and setter |
| `cmd/vdash/main.go` | Wire MetricsRepository to Model (~line 115) |

### Sparkline Implementation Details

**Unicode Block Characters (ascending height):**
```
▁ U+2581 LOWER ONE EIGHTH BLOCK
▂ U+2582 LOWER ONE QUARTER BLOCK
▃ U+2583 LOWER THREE EIGHTHS BLOCK
▄ U+2584 LOWER HALF BLOCK
▅ U+2585 LOWER FIVE EIGHTHS BLOCK
▆ U+2586 LOWER THREE QUARTERS BLOCK
▇ U+2587 LOWER SEVEN EIGHTHS BLOCK
█ U+2588 FULL BLOCK
```

**Normalization Algorithm:**
```go
func normalize(counts []int, levels int) []int {
    if len(counts) == 0 {
        return nil
    }

    max := 0
    for _, c := range counts {
        if c > max {
            max = c
        }
    }

    if max == 0 {
        // All zeros - return lowest level
        result := make([]int, len(counts))
        return result // All zeros = level 0 = '▁'
    }

    result := make([]int, len(counts))
    for i, c := range counts {
        // Scale to 0-(levels-1) range
        result[i] = (c * (levels - 1)) / max
    }
    return result
}
```

### Time Bucketing Strategy

**Default Configuration:**
- Time range: 30 days (configurable in Story 16.6)
- Default buckets: 7 (one per ~4.3 days)
- Wide terminal (>100 cols): 14 buckets
- Narrow terminal (<60 cols): 7 buckets

**Bucket Calculation:**
```go
// bucketIndex calculates which bucket a timestamp belongs to.
// now: reference time (for testability, use time.Now() in production)
// startTime: beginning of time range (now - timeRange)
// bucketDuration: duration of each bucket (timeRange / totalBuckets)
func bucketIndex(timestamp, startTime time.Time, bucketDuration time.Duration, totalBuckets int) int {
    elapsed := timestamp.Sub(startTime)
    idx := int(elapsed / bucketDuration)
    if idx >= totalBuckets {
        idx = totalBuckets - 1
    }
    if idx < 0 {
        idx = 0
    }
    return idx
}
```

**Activity Bucketing Implementation:**
```go
func BucketActivityCounts(timestamps []time.Time, buckets int, timeRange time.Duration, now time.Time) []int {
    if buckets <= 0 {
        return nil
    }
    counts := make([]int, buckets)
    if len(timestamps) == 0 {
        return counts // All zeros
    }
    startTime := now.Add(-timeRange)
    bucketDuration := timeRange / time.Duration(buckets)
    for _, ts := range timestamps {
        idx := bucketIndex(ts, startTime, bucketDuration, buckets)
        counts[idx]++
    }
    return counts
}
```

### Repository Method Implementation

```go
// GetTransitionsByProject retrieves transitions for a project since the given time.
// Returns empty slice on any error (graceful degradation).
func (r *MetricsRepository) GetTransitionsByProject(ctx context.Context, projectID string, since time.Time) []StageTransition {
    db, err := r.openDB(ctx)
    if err != nil {
        slog.Warn("metrics database unavailable", "error", err)
        return nil
    }
    defer db.Close()

    if err := r.ensureSchema(ctx, db); err != nil {
        slog.Warn("metrics schema error", "error", err)
        return nil
    }

    sinceStr := since.UTC().Format(time.RFC3339Nano)
    var rows []stageTransitionRow
    if err := db.SelectContext(ctx, &rows, selectByProjectWithTimeSQL, projectID, sinceStr); err != nil {
        slog.Warn("failed to query transitions", "error", err, "project_id", projectID)
        return nil
    }

    result := make([]StageTransition, len(rows))
    for i, row := range rows {
        result[i] = rowToTransition(&row)
    }
    return result
}

// GetTransitionTimestamps returns timestamps for TUI sparkline rendering.
// Avoids exposing internal types to TUI layer.
func (r *MetricsRepository) GetTransitionTimestamps(ctx context.Context, projectID string, since time.Time) []statsview.Transition {
    transitions := r.GetTransitionsByProject(ctx, projectID, since)
    result := make([]statsview.Transition, len(transitions))
    for i, t := range transitions {
        result[i] = statsview.Transition{TransitionedAt: t.TransitionedAt}
    }
    return result
}
```

### NFR Compliance

| Requirement | Target | This Story |
|-------------|--------|------------|
| NFR-P2-5 | Stats View render < 500ms | Sparkline rendering O(n) where n = transitions |
| NFR-P2-7 | Metrics failure doesn't crash | nil-safe metricsReader, empty slice returns |

### Previous Story Patterns

**From Story 16.1 (Repository):**
- Graceful degradation: return nil/empty on errors
- Use `slog.Warn()` for logging failures
- `sync.Once` for thread-safe schema init

**From Story 16.3 (Stats View):**
- Width calculation: `m.isWideWidth()` and `m.maxContentWidth`
- Content height: `m.height - statusBarHeight(m.height) - 2`
- Exit handling: `Esc` and `q` keys

### Testing Strategy

**Unit Tests Focus:**
1. Sparkline rendering with edge cases (empty, zeros, max values)
2. Activity bucketing correctness
3. Repository query methods with mock data
4. Stats View integration with nil metricsReader

**Manual Verification:**
- Run vdash with existing metrics.db → sparklines appear
- Run vdash without metrics.db → graceful empty view
- Resize terminal → sparklines adapt

## Dev Agent Record

### Context Reference

- Story 16.1 patterns for repository methods
- Story 16.3 patterns for Stats View rendering
- Phase 2 PRD FR-P2-16 requirement
- project-context.md for testing and architecture patterns

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - implementation straightforward, no debugging required.

### Completion Notes List

- All 9 tasks completed with comprehensive test coverage
- Repository methods implemented with graceful degradation pattern
- Sparkline rendering uses Unicode block characters (▁▂▃▄▅▆▇█)
- Activity bucketing supports 7-14 buckets based on terminal width
- Full TUI integration with metrics reader wiring from main.go through CLI to TUI
- Tests verify all acceptance criteria including nil-safety

### File List

| File | Action | Description |
|------|--------|-------------|
| `internal/adapters/persistence/metrics/queries.go` | Modified | Added `selectByProjectWithTimeSQL` query constant |
| `internal/adapters/persistence/metrics/helpers.go` | Modified | Added `StageTransition` public type and `rowToTransition` helper |
| `internal/adapters/persistence/metrics/repository.go` | Modified | Added `GetTransitionsByProject`, `GetTransitionsByTimeRange`, `GetTransitionTimestamps` methods |
| `internal/adapters/persistence/metrics/repository_test.go` | Modified | Added tests for new repository methods |
| `internal/adapters/tui/statsview/sparkline.go` | Created | Sparkline rendering module with `RenderSparkline` function |
| `internal/adapters/tui/statsview/activity.go` | Created | Activity bucketing module with `BucketActivityCounts` function |
| `internal/adapters/tui/statsview/interfaces.go` | Created | `Transition` type and `MetricsReader` interface |
| `internal/adapters/tui/statsview/sparkline_test.go` | Created | Comprehensive tests for sparkline rendering |
| `internal/adapters/tui/statsview/activity_test.go` | Created | Comprehensive tests for activity bucketing |
| `internal/adapters/tui/statsview.go` | Modified | Stats View rendering with sparkline integration |
| `internal/adapters/tui/model.go` | Modified | Added `metricsReader` field, setter, and `getProjectActivity` method |
| `internal/adapters/tui/statsview_test.go` | Modified | Added tests for sparkline integration in Stats View |
| `internal/adapters/tui/app.go` | Modified | Wire metricsReader to Model |
| `internal/adapters/cli/deps.go` | Modified | Added `metricsReader` variable and getter/setter |
| `internal/adapters/cli/root.go` | Modified | Pass metricsReader to TUI Run function |
| `cmd/vdash/main.go` | Modified | Create MetricsRepository and wire to CLI via SetMetricsReader |
