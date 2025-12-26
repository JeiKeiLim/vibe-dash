# Story 8.11: Periodic Stage Re-Detection

Status: backlog

## Story

As a **user monitoring project progress**,
I want **stage detection to refresh automatically**,
So that **I see updated epic/story status without pressing [r] manually**.

## Problem Statement

Current auto-refresh behavior:
- Waiting status: Updates via file watcher events (event-driven, no polling)
- Stage detection: Only runs on manual [r] refresh

Stage info (Epic X, Story Y) requires reading/parsing files, so it runs on a separate timer.

## Design Decision

**Option B: Configurable periodic timer (default 30s)**

Rationale:
- Balances freshness vs I/O efficiency
- User can adjust based on preference
- 30s default is unnoticeable delay in practice
- Method-agnostic (works for BMAD, Speckit, future methods)
- No hardcoded file patterns per detection method

## Acceptance Criteria

```gherkin
AC1: Given default config (no setting)
     When 30 seconds pass since last detection
     Then stage detection runs automatically for all projects

AC2: Given config `stage_refresh_interval: 60s`
     When 60 seconds pass
     Then stage detection runs automatically

AC3: Given config `stage_refresh_interval: 0`
     When any time passes
     Then stage detection only runs on manual [r] refresh

AC4: Given periodic detection completes
     When new stage info differs from current
     Then project list and detail panel update immediately

AC5: Given user presses [r] manually
     When refresh completes
     Then periodic timer resets (avoids redundant detection)

AC6: Given stage detection is running
     When multiple projects exist
     Then all projects are re-detected in single batch
```

## Tasks

- [ ] Add `stage_refresh_interval` config option (duration, default 30s, 0=disabled)
- [ ] Create `stageRefreshTickMsg` in model.go
- [ ] Add periodic timer using `tea.Tick` pattern
- [ ] Handle `stageRefreshTickMsg` to call detection service
- [ ] Reset timer after manual [r] refresh
- [ ] Test with 10s, 30s, 60s, 0 (disabled) intervals

## Technical Notes

- Reuse existing `startRefresh()` logic or extract detection-only path
- Timer independent of UI tick (1s) - separate tea.Tick for stage refresh
- Config validation: minimum 10s if not 0 (prevent excessive I/O)

Files to modify:
- `internal/adapters/tui/model.go` - new tick message and handler
- `internal/config/` - new config option with validation

## Dev Notes

_To be filled during implementation_

## Change Log

- 2025-12-26: Story created via correct-course workflow (user feedback during 8.4 review)
