# Story 16.6: Add Date Range Selector

Status: done

## Story

As a user,
I want to select a date range for metrics,
So that I can focus on specific time periods.

## User-Visible Changes

- **New:** Stats View header shows current date range (e.g., "Last 30 days")
- **New:** Keyboard shortcuts `[` / `]` cycle through date range presets
- **New:** Presets available: 7 days, 30 days, 90 days, 1 year, All time
- **Changed:** Sparklines and breakdown update dynamically when range changes
- **Changed:** Breakdown view displays selected period in header

## Acceptance Criteria

1. **Given** Stats View is open
   **When** I press `]` key
   **Then** date range cycles to next preset (7d → 30d → 90d → 1y → All)

2. **And** given I press `[` key
   **When** currently on 90 days
   **Then** date range cycles back to 30 days

3. **And** sparklines and breakdown update immediately to reflect selected range

4. **And** default range is "Last 30 days" (preserves current behavior)

5. **And** Stats View header displays current range (e.g., "Activity (7d)", "Activity (90d)")

6. **And** breakdown view displays "Period: Last X days" or "Period: All time"

7. **And** date range selection persists within Stats View session (resets on re-entry)

## Tasks / Subtasks

- [x] Task 1: Create DateRange type and presets (AC: #1, #2, #4)
  - [x] Create `internal/adapters/tui/statsview/daterange.go`:
    ```go
    package statsview

    import "time"

    // DateRangePreset represents a predefined time period for metrics filtering.
    type DateRangePreset int

    const (
        DateRange7Days DateRangePreset = iota
        DateRange30Days   // Default
        DateRange90Days
        DateRange1Year
        DateRangeAllTime
    )

    // DateRange holds the selected preset and provides helper methods.
    type DateRange struct {
        Preset DateRangePreset
    }

    // DefaultDateRange returns the default 30-day range.
    func DefaultDateRange() DateRange

    // Next returns the next preset (wraps around).
    func (d DateRange) Next() DateRange

    // Prev returns the previous preset (wraps around).
    func (d DateRange) Prev() DateRange

    // Since returns the start time for this range (time.Now() - duration).
    // For AllTime, returns time.Time{} (zero value = no filter).
    func (d DateRange) Since() time.Time

    // Label returns human-readable label (e.g., "7d", "30d", "All").
    func (d DateRange) Label() string

    // HeaderLabel returns label for column header (e.g., "Activity (7d)").
    func (d DateRange) HeaderLabel() string

    // BreakdownLabel returns label for breakdown view (e.g., "Last 7 days").
    func (d DateRange) BreakdownLabel() string
    ```
  - [x] Implement preset cycling (Next/Prev wrap around)
  - [x] Implement Since() calculation:
    - 7d: now - 7*24h
    - 30d: now - 30*24h (default)
    - 90d: now - 90*24h
    - 1y: now - 365*24h
    - All: time.Time{} (zero value = effectively no filter)

- [x] Task 2: Add DateRange state to Model (AC: #4, #7)
  - [x] In `internal/adapters/tui/model.go` (~line 190, after statsBreakdownDurations), add state field:
    ```go
    // Stats View date range state (Story 16.6)
    statsDateRange statsview.DateRange  // Current date range preset
    ```
  - [x] Initialize to `statsview.DefaultDateRange()` in `enterStatsView()` (~line 3370)
  - [x] Date range resets on each Stats View entry (AC #7)

- [x] Task 3: Update key bindings for date range cycling (AC: #1, #2)
  - [x] In `internal/adapters/tui/model.go`, Update method (~line 2830), within Stats View key handling:
    ```go
    case "[":
        if m.statsBreakdownProject == nil {
            m.statsDateRange = m.statsDateRange.Prev()
        }
        return m, nil
    case "]":
        if m.statsBreakdownProject == nil {
            m.statsDateRange = m.statsDateRange.Next()
        }
        return m, nil
    ```
  - [x] Only cycle when NOT in breakdown view (project list view only)
  - [x] In breakdown view, `[` and `]` can be reserved for future use (or ignore)

- [x] Task 4: Update sparkline data fetching to use date range (AC: #3)
  - [x] Modify `getProjectActivity()` in model.go (~line 3390) to use `m.statsDateRange.Since()`:
    ```go
    func (m Model) getProjectActivity(projectID string, buckets int) []int {
        if m.metricsReader == nil {
            return nil
        }
        ctx := context.Background()
        since := m.statsDateRange.Since()  // Use selected range instead of hardcoded 30 days
        transitions := m.metricsReader.GetTransitionTimestamps(ctx, projectID, since)
        return statsview.CalculateActivityBuckets(transitions, buckets, since, time.Now())
    }
    ```
  - [x] Update `CalculateActivityBuckets` signature if needed to accept date range

- [x] Task 5: Update breakdown data fetching to use date range (AC: #3)
  - [x] Modify `getStageBreakdown()` in model.go (~line 3405) to use selected date range:
    ```go
    func (m Model) getStageBreakdown(projectID string) []statsview.StageDuration {
        if m.metricsReader == nil {
            return nil
        }
        ctx := context.Background()
        since := m.statsDateRange.Since()  // Use selected range
        transitions := m.metricsReader.GetFullTransitions(ctx, projectID, since)
        return statsview.CalculateFromFullTransitions(transitions, time.Now())
    }
    ```

- [x] Task 6: Update Stats View header to show date range (AC: #5)
  - [x] In `renderStatsProjectList()` (statsview.go:96), update column header:
    ```go
    headerLine := fmt.Sprintf("%-*s  %-*s  %s",
        nameWidth, "Project",
        sparklineWidth, m.statsDateRange.HeaderLabel(),  // e.g., "Activity (7d)"
        "Stage")
    ```
  - [x] Pass `statsDateRange` to render method or access via model

- [x] Task 7: Update breakdown view to show date range (AC: #6)
  - [x] In `renderStatsBreakdownView()` (statsview.go:255), update period line:
    ```go
    contentLines = append(contentLines, dimStyle.Render("Period: "+m.statsDateRange.BreakdownLabel()))
    ```

- [x] Task 8: Add key hint to Stats View header (AC: #1)
  - [x] Update Stats View header in `renderStatsProjectListView()` (statsview.go:25) to show `[/] Range` hint
  - [x] Combine with existing `[ESC] Back to Dashboard` hint
  - [x] Example: `"[/] Range  [ESC] Back"`

- [x] Task 9: Update CalculateActivityBuckets for variable date ranges
  - [x] In `internal/adapters/tui/statsview/activity.go`, added `CalculateTimeRangeFromTimestamps` function:
    ```go
    // CalculateActivityBuckets distributes transitions into time buckets.
    // since: start of date range (zero time means calculate from earliest transition)
    // now: end of date range (usually time.Now())
    func CalculateActivityBuckets(transitions []Transition, buckets int, since, now time.Time) []int {
        if len(transitions) == 0 || buckets <= 0 {
            return nil
        }

        // Handle All Time: find earliest transition
        startTime := since
        if since.IsZero() && len(transitions) > 0 {
            startTime = transitions[0].TransitionedAt
            for _, t := range transitions {
                if t.TransitionedAt.Before(startTime) {
                    startTime = t.TransitionedAt
                }
            }
        }

        timeRange := now.Sub(startTime)
        if timeRange <= 0 {
            return make([]int, buckets)
        }

        counts := make([]int, buckets)
        bucketDuration := timeRange / time.Duration(buckets)
        if bucketDuration <= 0 {
            return counts
        }

        for _, t := range transitions {
            elapsed := t.TransitionedAt.Sub(startTime)
            idx := int(elapsed / bucketDuration)
            if idx >= buckets {
                idx = buckets - 1
            }
            if idx < 0 {
                idx = 0
            }
            counts[idx]++
        }
        return counts
    }
    ```
  - [x] Handle edge case: `since.IsZero()` for All Time range
  - [x] Note: Existing `BucketActivityCounts` in activity.go used with new helper `CalculateTimeRangeFromTimestamps`

- [x] Task 10: Write unit tests for DateRange
  - [x] `internal/adapters/tui/statsview/daterange_test.go`:
    ```go
    func TestDateRange_Next_Cycles(t *testing.T) { ... }
    func TestDateRange_Prev_Cycles(t *testing.T) { ... }
    func TestDateRange_Since_Calculations(t *testing.T) { ... }
    func TestDateRange_Labels(t *testing.T) { ... }
    func TestDefaultDateRange(t *testing.T) { ... }
    ```

- [x] Task 11: Update existing Stats View tests
  - [x] `internal/adapters/tui/statsview_test.go`:
    - Test date range key handling (`[` and `]`)
    - Test date range persists within session
    - Test date range resets on re-entry
  - [x] Verify sparkline/breakdown use selected range

## Dev Notes

### Architecture Alignment

Implements **FR-P2-18** from Phase 2 PRD: "Users can select date range for metrics display"

**Isolation Principle:** Date range feature builds on existing Stats View infrastructure:
- Adds new type in `statsview/` package (clean separation)
- State managed in `model.go` (follows existing pattern)
- Reuses existing metrics query methods (already accept `since` parameter)

**Import Direction (CRITICAL - Prevents Architecture Violation):**
```
internal/adapters/tui
    └── imports → internal/adapters/tui/statsview (for DateRange type)

internal/adapters/tui/statsview
    └── NO imports from tui layer (clean separation)
    └── NO imports from model.go
```

### File Modifications

| File | Changes |
|------|---------|
| `internal/adapters/tui/statsview/daterange.go` | **NEW:** DateRange type, presets, cycling logic |
| `internal/adapters/tui/statsview/daterange_test.go` | **NEW:** Unit tests for DateRange |
| `internal/adapters/tui/statsview/sparkline.go` | Add/update CalculateActivityBuckets for variable ranges |
| `internal/adapters/tui/model.go` | Add `statsDateRange` field (~line 194), update `enterStatsView()` (~line 3370), update key handling (~line 2830) |
| `internal/adapters/tui/statsview.go` | Update header labels (line 122) and breakdown period display (line 255) |
| `internal/adapters/tui/statsview_test.go` | Add tests for date range key handling |

### Existing Code References

**model.go key locations:**
- `statsBreakdownProject` field: line 192
- `statsBreakdownDurations` field: line 193
- `metricsReaderInterface` interface: lines 40-43 (already accepts `since time.Time`)
- `getProjectActivity()` method: line 3390 (needs modification)
- `getStageBreakdown()` method: line 3405 (needs modification)
- `enterStatsView()` method: line 3370 (needs initialization)
- Stats View key handling: line 2830 (add `[` and `]` cases)

**statsview.go key locations:**
- `renderStatsProjectList()`: line 96 (header line at 122)
- `renderStatsBreakdownView()`: line 207 (period line at 255)

**Repository methods (already support `since` parameter - NO CHANGES NEEDED):**
- `GetTransitionTimestamps(ctx, projectID, since)` - already accepts time.Time
- `GetFullTransitions(ctx, projectID, since)` - already accepts time.Time

### Date Range Presets

| Preset | Label | Header Label | Breakdown Label | Since Calculation |
|--------|-------|--------------|-----------------|-------------------|
| 7 days | "7d" | "Activity (7d)" | "Last 7 days" | now - 7*24h |
| 30 days | "30d" | "Activity (30d)" | "Last 30 days" | now - 30*24h |
| 90 days | "90d" | "Activity (90d)" | "Last 90 days" | now - 90*24h |
| 1 year | "1y" | "Activity (1y)" | "Last year" | now - 365*24h |
| All time | "All" | "Activity (All)" | "All time" | time.Time{} |

### Key Binding Rationale

Using `[` and `]` because:
- They visually suggest "left/right" or "previous/next"
- Not used elsewhere in Stats View
- Easy to reach on standard keyboard
- Similar to vim-style bracket navigation

**Alternative considered:** Number keys (1-5) for direct preset selection
- Rejected: conflicts with potential future navigation
- May revisit if users request direct selection

### Sparkline Calculation for Variable Ranges

**Current behavior (30 days hardcoded in getProjectActivity):**
- Uses `now.Add(-30 * 24 * time.Hour)` as `since`
- Divides 30 days into N buckets

**New behavior (variable range via statsDateRange):**
- For fixed ranges (7d, 30d, 90d, 1y): divide range by bucket count
- For "All Time" (`since.IsZero()`):
  - Find earliest transition timestamp
  - Calculate span from earliest to now
  - Divide span by bucket count
  - Handle edge case: no transitions → return nil/flat sparkline

**Edge Cases:**
1. **No transitions:** Return nil (rendered as flat `▁▁▁▁▁▁▁`)
2. **All time with no data:** Same as above
3. **Short range with many transitions:** Buckets may overflow visually (cap at max bar)
4. **Date range shorter than transitions:** Some transitions excluded (expected)
5. **Zero time range:** Return empty counts

### NFR Compliance

| Requirement | Target | This Story |
|-------------|--------|------------|
| NFR-P2-5 | Stats View render < 500ms | DateRange calculation is O(1), no impact |
| NFR-P2-7 | Metrics failure doesn't crash | Existing graceful degradation unchanged |

### Previous Story Patterns

**From Story 16.5 (Breakdown) - docs/sprint-artifacts/stories/epic-16/16-5-implement-time-per-stage-breakdown.md:**
- Sub-state pattern (`statsBreakdownProject != nil`)
- Key handling in model.go Update method (line 2836)
- `getStageBreakdown()` method for data fetching
- Breakdown view rendering

**From Story 16.4 (Sparklines) - docs/sprint-artifacts/stories/epic-16/16-4-implement-activity-sparklines.md:**
- `metricsReader` interface pattern
- `getProjectActivity()` method for sparkline data
- Column header rendering in `renderStatsProjectList()`
- Graceful nil handling for metricsReader

**From Story 16.3 (Stats View):**
- `enterStatsView()` / `exitStatsView()` methods
- `statsViewScroll` state management
- `viewModeStats` constant

### Critical Implementation Notes

**DO NOT:**
- Allow date range changes in breakdown view (only in project list)
- Persist date range across Stats View sessions
- Add date range state to breakdown view (reuses project list's range)
- Modify repository methods (they already accept `since` parameter)

**MUST:**
- Initialize `statsDateRange` in `enterStatsView()` (not in model init)
- Use `m.statsDateRange.Since()` in both sparkline and breakdown data fetching
- Update column header dynamically based on selected range
- Handle `time.Time{}` (zero value) for All Time range
- Preserve sparkline bucket count calculation (independent of date range)

### Project Structure Notes

**New Files:**
- `internal/adapters/tui/statsview/daterange.go` - DateRange type and methods
- `internal/adapters/tui/statsview/daterange_test.go` - Unit tests

**Modified Files:**
- `internal/adapters/tui/model.go` - Add state field, enterStatsView init, key handling
- `internal/adapters/tui/statsview.go` - Update headers and period display
- `internal/adapters/tui/statsview/sparkline.go` - Add CalculateActivityBuckets function
- `internal/adapters/tui/statsview_test.go` - Add integration tests

### References

- [Source: docs/epics-phase2.md#Story 4.6 - Date Range Selector]
- [Source: docs/project-context.md#Phase 2 Additions - Progress Metrics]
- [Source: docs/sprint-artifacts/stories/epic-16/16-5-implement-time-per-stage-breakdown.md]
- [Source: docs/sprint-artifacts/stories/epic-16/16-4-implement-activity-sparklines.md]

## Dev Agent Record

### Context Reference

- Story 16.5 for breakdown patterns, `getStageBreakdown()`, renderStatsBreakdownView() at statsview.go:207
- Story 16.4 for sparkline patterns, `getProjectActivity()` at model.go:3390, CalculateActivityBuckets()
- Story 16.3 for Stats View structure, enterStatsView() at model.go:3370, key handling
- Phase 2 PRD FR-P2-18 requirement
- Existing `metricsReader.GetTransitionTimestamps()` already accepts `since` parameter
- Existing `metricsReader.GetFullTransitions()` already accepts `since` parameter
- Existing `metricsReaderInterface` at model.go:40-43 (no changes needed)
- Existing `statsBreakdownProject` at model.go:192
- Existing `statsBreakdownDurations` at model.go:193

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- All 1453 tests pass
- Lint passes after `make fmt`
- Build successful

### Completion Notes List

1. Created `daterange.go` with DateRange type and presets (7d, 30d, 90d, 1y, All)
2. Added `statsDateRange` field to Model (line 196)
3. Updated `enterStatsView()` to initialize date range to default on entry
4. Added `[` and `]` key handlers for cycling date ranges in project list view
5. Updated `getProjectActivity()` to use `m.statsDateRange.Since()` instead of hardcoded 30d
6. Updated `getStageBreakdown()` to use `m.statsDateRange.Since()` instead of hardcoded 30d
7. Added `CalculateTimeRangeFromTimestamps()` helper for "All time" range calculation
8. Updated column header to show dynamic label (e.g., "Activity (7d)")
9. Updated breakdown view to show dynamic period label (e.g., "Period: Last 7 days")
10. Updated Stats View header hint to `[ ] Range  [ESC] Back`
11. Added comprehensive tests for DateRange and integration tests

### Code Review Fixes (2026-01-16)

| Issue | Severity | Fix Applied |
|-------|----------|-------------|
| Key hint mismatch: `[/]` vs actual `[` and `]` keys | HIGH | Changed hint to `[ ] Range` for clarity |
| Duration() method not tested | HIGH | Added `TestDateRange_Duration` test |
| Missing test for invalid preset defaults | MEDIUM | Added `TestDateRange_InvalidPreset_Defaults` test |
| Test date year inconsistency (2025 vs 2026) | LOW | Updated daterange_test.go to use 2026 |

### File List

| File | Status | Changes |
|------|--------|---------|
| `internal/adapters/tui/statsview/daterange.go` | **NEW** | DateRange type, presets, cycling logic, labels |
| `internal/adapters/tui/statsview/daterange_test.go` | **NEW** | Unit tests for DateRange |
| `internal/adapters/tui/statsview/activity.go` | MODIFIED | Added `CalculateTimeRangeFromTimestamps()` helper |
| `internal/adapters/tui/statsview/activity_test.go` | MODIFIED | Added tests for `CalculateTimeRangeFromTimestamps()` |
| `internal/adapters/tui/model.go` | MODIFIED | Added `statsDateRange` field, updated `enterStatsView()`, `getProjectActivity()`, `getStageBreakdown()`, and key handling |
| `internal/adapters/tui/statsview.go` | MODIFIED | Updated header label and breakdown period display |
| `internal/adapters/tui/statsview_test.go` | MODIFIED | Added tests for date range functionality |

## User Testing Guide

**Time needed:** 5 minutes

### Step 1: Build and Run
```bash
make build
./bin/vdash
```

### Step 2: Enter Stats View
1. Press `s` to open Stats View
2. **Expected:** See project list with "STATS" header
3. **Expected:** Column header shows "Activity (30d)" (default range)

### Step 3: Test Date Range Cycling
1. Press `]` key
2. **Expected:** Header changes to "Activity (90d)"
3. **Expected:** Sparklines update (may show more/less activity)
4. Press `]` again
5. **Expected:** Header changes to "Activity (1y)"
6. Press `]` again
7. **Expected:** Header changes to "Activity (All)"
8. Press `]` again
9. **Expected:** Header wraps to "Activity (7d)"
10. Press `[` key
11. **Expected:** Header cycles back to "Activity (All)"

### Step 4: Test Breakdown View Period
1. Press `Enter` on a project
2. **Expected:** Breakdown view shows "Period: Last 7 days" (or current range)
3. Press `Esc` to return to list
4. Press `]` to change range to 30d
5. Press `Enter` on project again
6. **Expected:** Breakdown view shows "Period: Last 30 days"

### Step 5: Test Session Persistence
1. While in Stats View, cycle to "Activity (90d)"
2. Press `Esc` to return to Dashboard
3. Press `s` to re-enter Stats View
4. **Expected:** Range resets to "Activity (30d)" (default)

### Step 6: Test Key Disabled in Breakdown
1. Enter Stats View and press `Enter` on a project (breakdown view)
2. Press `]` key
3. **Expected:** Nothing happens (date range cycling disabled in breakdown)
4. Press `Esc` to return to list
5. Press `]` key
6. **Expected:** Date range cycles (works in list view)

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Date range not cycling | Do NOT approve, check key handling at model.go:2830 |
| Header not updating | Do NOT approve, check renderStatsProjectList |
| Sparklines not changing | Do NOT approve, check getProjectActivity |
| Breakdown period wrong | Do NOT approve, check renderStatsBreakdownView |
| Range not resetting on re-entry | Do NOT approve, check enterStatsView |
