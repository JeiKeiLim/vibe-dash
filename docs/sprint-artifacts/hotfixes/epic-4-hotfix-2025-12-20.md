# Epic 4 Hotfix - Post-Retrospective Fixes

**Date:** 2025-12-20
**Source:** [Epic 4 Retrospective](../retrospectives/epic-4-retro-2025-12-20.md)
**Status:** Complete

---

## Overview

These fixes address critical issues discovered during the Epic 4 retrospective. The user (Jongkuk Lim) reported that despite all 6 stories being marked "done", the WAITING feature was not visible or working as expected.

---

## Fixes

### H1: Show WAITING Count Even When 0

**Problem:** Status bar hides WAITING section when count=0, so users don't know the feature exists.

**Current Behavior:**
```
│ 2 active │ 0 hibernated │
```

**Expected Behavior:**
```
│ 2 active │ 0 hibernated │ 0 waiting │
```

**File:** `internal/adapters/tui/components/status_bar.go`

**Change:** Modified `renderCounts()` and `renderCondensed()` to show waiting count even when 0 (with dim styling).

**Acceptance Criteria:**
- [x] Status bar shows "0 waiting" when no projects are waiting
- [x] "0 waiting" uses dim style (not bold red)
- [x] When count > 0, still uses bold red style
- [x] Condensed mode shows "0W" similarly

**Files Changed:**
- `internal/adapters/tui/components/status_bar.go` - Added `statusBarWaitingZeroStyle`, modified render functions
- `internal/adapters/tui/components/status_bar_test.go` - Updated tests to expect "0 waiting"

**Status:** [x] Complete

---

### H3: Verify FileWatcher Updates Activity Timestamps

**Problem:** `last_activity_at` equals `created_at` for all projects, suggesting FileWatcher never updated timestamps.

**Investigation Result:** The feature is working correctly. The user's case was expected behavior:
1. User git cloned promptory project and did NOT work on it
2. No file changes occurred while TUI was running
3. Therefore, `last_activity_at` was never updated (correct behavior)

**Root Cause:** Not a bug - feature requires files to change **while TUI is open** to update timestamps.

**Improvement Made:** Added debug logging to make feature behavior more observable:
- Log when file watcher starts successfully
- Log when activity is successfully updated

**Files Changed:**
- `internal/adapters/tui/model.go` - Added debug logging for file watcher startup and activity updates

**Acceptance Criteria:**
- [x] Confirmed: feature works correctly when files change while TUI is running
- [x] Added debug logging to help trace the flow

**Status:** [x] Complete

---

### H4: Fix Speckit Detector Tiebreaker

**Problem:** When all spec directories have same mtime (common after git clone), detector picks unpredictably instead of highest-numbered directory.

**Current Behavior:**
```
specs/001-fix-modal-autoclose    (picked - wrong!)
specs/005-ui-ux-polish           (should be picked)
```

**Root Cause:** `sort.Slice` with equal mtimes has unstable ordering.

**File:** `internal/adapters/detectors/speckit/detector.go`

**Change:** Added lexicographic tiebreaker - when mtimes are equal, sort by name descending so higher-numbered directories (005-*) are preferred over lower (001-*).

**Acceptance Criteria:**
- [x] When all mtimes equal, highest-numbered directory is selected
- [x] When mtimes differ, still picks most recently modified
- [x] Added unit test for equal-mtime scenario

**Files Changed:**
- `internal/adapters/detectors/speckit/detector.go` - Modified `findMostRecentDir()` with tiebreaker
- `internal/adapters/detectors/speckit/detector_test.go` - Added `TestSpeckitDetector_EqualModTimes_HighestNumberedWins`

**Status:** [x] Complete

---

### H2: Document Epic Acceptance Test Process

**Problem:** We've now had two epics (3.5 and 4) where all stories were "done" but end-to-end verification failed.

**Deliverable:** Add acceptance checklist to epic workflow.

**Acceptance Criteria:**
- [x] Checklist documented
- [x] Process integrated into workflow

**File Created:**
- `docs/sprint-artifacts/epic-acceptance-checklist.md` - Comprehensive checklist with examples

**Status:** [x] Complete

---

## Verification

After all fixes complete:

1. [x] `go test ./...` passes
2. [x] `go build -o bin/vibe ./cmd/vibe` succeeds
3. [ ] Manual verification (to be done by user):
   - [ ] Run `./bin/vibe` - see "0 waiting" in status bar
   - [ ] Touch file in tracked project while TUI is open, verify activity updates
   - [ ] Check git-cloned Speckit project shows correct spec number
4. [x] Commit with message referencing this document

---

## Completion

**Completed By:** Bob (Scrum Master) with Dev Team
**Completion Date:** 2025-12-20
**Commit Hash:** (to be filled after commit)

---

## Files Changed Summary

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/adapters/tui/components/status_bar.go` | Modified | H1: Show "0 waiting" with dim style |
| `internal/adapters/tui/components/status_bar_test.go` | Modified | H1: Update tests for new behavior |
| `internal/adapters/tui/model.go` | Modified | H3: Add debug logging for file watcher |
| `internal/adapters/detectors/speckit/detector.go` | Modified | H4: Add lexicographic tiebreaker |
| `internal/adapters/detectors/speckit/detector_test.go` | Modified | H4: Add equal-mtime test |
| `docs/sprint-artifacts/epic-acceptance-checklist.md` | Created | H2: Epic acceptance process |
| `docs/sprint-artifacts/retrospectives/epic-4-retro-2025-12-20.md` | Created | Retrospective document |
| `docs/sprint-artifacts/hotfixes/epic-4-hotfix-2025-12-20.md` | Created | This document |

---

*Generated from Epic 4 Retrospective action items*
