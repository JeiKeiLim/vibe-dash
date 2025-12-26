# Story 8.10: Full-Width Layout & Column Rebalancing

Status: backlog

## Story

As a **user viewing the project list**,
I want **column widths optimized for actual content**,
So that **stage info and waiting status are readable without wasted space on short project names**.

## Problem Statement

1. `MaxContentWidth = 120` cap wastes space on wide monitors
2. Current column proportions are unbalanced:
   - Project name: Too wide (names are typically short)
   - Stage info: Too narrow (truncates important info)
   - Waiting status: Too narrow ("WAITING 2h 30m" barely fits)

## Acceptance Criteria

### Column Rebalancing (All Widths)

```gherkin
AC1: Given normal terminal width (80-120 cols)
     When viewing project list
     Then column proportions prioritize stage and status readability:
     - Project name: ~25% (enough for typical names)
     - Method: fixed ~6 chars
     - Stage info: ~40% (epic/story info needs space)
     - Status/Waiting: ~20% (fit "WAITING 2h 30m")
     - Activity: remaining

AC2: Given a long project name
     When name exceeds allocated width
     Then name truncates with ellipsis (not stage info)
```

### Full-Width Option (Wide Monitors)

```gherkin
AC3: Given config `max_content_width: 0` (disabled)
     When terminal is 200+ columns
     Then content uses full width with proportional columns

AC4: Given config `max_content_width: N` (custom cap)
     When terminal exceeds N columns
     Then content is capped and centered

AC5: Given no config set (default)
     When running
     Then backward-compatible 120-column cap
```

### Column Max Widths (Prevent Absurd Stretching)

```gherkin
AC6: Given ultra-wide terminal (>180 cols)
     When proportional widths calculated
     Then columns have sensible max widths:
     - Project name: max 40 chars
     - Stage info: max 80 chars
     - Status: max 25 chars
```

## Tasks

- [ ] Audit current column width calculations in project_item_delegate.go
- [ ] Redefine column proportions (name smaller, stage larger)
- [ ] Add max-width caps per column for ultra-wide
- [ ] Add `max_content_width` config option
- [ ] Update model.go effectiveWidth logic
- [ ] Test across 80, 100, 120, 160, 200 column widths
- [ ] Update any affected tests

## Technical Notes

Files to modify:
- `internal/adapters/tui/components/project_item_delegate.go` - column widths
- `internal/adapters/tui/model.go` - effectiveWidth logic
- `internal/adapters/tui/views.go` - MaxContentWidth constant
- `internal/config/` - new config option

## Dev Notes

_To be filled during implementation_

## Change Log

- 2025-12-26: Story created via correct-course workflow (user feedback during 8.4 review)
