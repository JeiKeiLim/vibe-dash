# Story 8.9: Visual Polish Bundle

Status: done

## Story

As a **user**,
I want **small visual improvements to the dashboard**,
So that **the interface feels polished and consistent**.

## Problem Statement

During systematic UX elicitation (C->B->A methodology), several minor visual polish items were identified. These are individually small but collectively improve the dashboard's professional feel.

**Source:** Epic 8 UX Polish - items P12 through P17 from elicitation session

## Design Decisions

### P12: Focus Indicator '>' Review

**Decision:** KEEP `> ` - it's a common TUI pattern (vim, fzf, etc.)
**Action:** Document in help overlay as intentional.

### P13: Project Count Redundancy

**Decision:** REMOVE pagination from Bubbles list status bar
**Implementation:** Set `l.SetShowStatusBar(false)` in `project_list.go:40`

### P14: Status Bar Decoration

**Decision:** KEEP current `|` separators - no changes needed.

### P15: Time Display in Status Bar

**Decision:** DEFER to post-MVP

### P16: Custom Status Bar Patterns

**Decision:** DEFER to post-MVP

### P17: Terminal Emoji Compatibility

**Decision:** ADD fallback characters for terminals without emoji support

**Emoji Fallback Table:**

| Feature | Emoji | Fallback |
|---------|-------|----------|
| Favorite | `⭐` | `*` |
| Waiting | `⏸️` | `[W]` |
| Today | `✨` | `+` |
| This week | `⚡` | `~` |
| Warning | `⚠️` | `!` |

**Detection Logic:**
- If `TERM` contains "linux", "vt100", "vt220", "ansi", or equals "dumb" -> use fallback
- Add config option `use_emoji` (nil=auto-detect, true=force emoji, false=force fallback)

## Acceptance Criteria

1. **AC1: Bubbles List Pagination Hidden**
   - Given the project list is displayed
   - When I view the bottom of the list
   - Then the "1/5" style pagination indicator is NOT shown
   - And the status bar still shows "X active | X hibernated | X waiting"

2. **AC2: Focus Indicator Documented**
   - Given the help overlay (?)
   - When I view key bindings
   - Then there is a note explaining `>` is the selection indicator

3. **AC3: Emoji Fallback - Auto Detection**
   - Given TERM=linux or TERM=dumb
   - When I run vibe
   - Then fallback characters are used instead of emoji

4. **AC4: Emoji Config Override**
   - Given config `use_emoji: false`
   - When I run vibe in any terminal
   - Then fallback characters are always used

5. **AC5: Emoji Config Default**
   - Given config `use_emoji: true` (or not set)
   - When I run vibe in a capable terminal
   - Then emoji are used as normal

6. **AC6: Tests Pass**
   - Given all changes are made
   - When `make test && make lint` runs
   - Then all tests pass and no lint errors

## Tasks / Subtasks

- [x] Task 1: Hide Bubbles List Pagination (AC: 1)
  - [x] 1.1: In `internal/adapters/tui/components/project_list.go:40`, change `l.SetShowStatusBar(true)` to `l.SetShowStatusBar(false)`
  - [x] 1.2: Verify status bar component still shows its counts
  - [x] 1.3: Search `project_list_test.go` for pagination assertions and update if found

- [x] Task 2: Document Focus Indicator (AC: 2)
  - [x] 2.1: In `internal/adapters/tui/views.go`, add note to `renderHelpOverlay()` explaining `>` is the selection indicator
  - [x] 2.2: Add under "Navigation" section: `">        Selection indicator"`

- [x] Task 3: Add Config Option for Emoji (AC: 4, 5)
  - [x] 3.1: In `internal/core/ports/config.go`, add `UseEmoji *bool` field to `Config` struct (nil=auto, true=force, false=disable)
  - [x] 3.2: In `internal/config/loader.go`, add Viper binding for `settings.use_emoji`
  - [x] 3.3: Update `NewConfig()` to NOT set UseEmoji (nil = auto-detect is default)

- [x] Task 4: Implement Emoji Fallback System (AC: 3, 4, 5)
  - [x] 4.1: Create `internal/shared/emoji/emoji.go` (follows `timeformat` package pattern)
  - [x] 4.2: Implement `InitEmoji(configValue *bool)` - caches decision at startup
  - [x] 4.3: Implement `detectEmojiSupport()` - check TERM for "linux", "vt100", "vt220", "ansi", "dumb"
  - [x] 4.4: Implement accessor functions: `Star()`, `Waiting()`, `Today()`, `ThisWeek()`, `Warning()`
  - [x] 4.5: Create `internal/shared/emoji/emoji_test.go` with table-driven tests

- [x] Task 5: Wire Emoji Init at App Startup (AC: 3, 4, 5)
  - [x] 5.1: In `internal/adapters/tui/app.go`, call `emoji.InitEmoji(config.UseEmoji)` AFTER config loads, BEFORE TUI renders

- [x] Task 6: Replace Hardcoded Emoji (AC: 3)
  - [x] 6.1: In `internal/adapters/tui/components/delegate.go`, replace:
    - Line 167: `"⭐"` -> `emoji.Star()`
    - Line 188: `"✨"` -> `emoji.Today()`
    - Line 190: `"⚡"` -> `emoji.ThisWeek()`
    - Line 234: `"⏸️"` -> `emoji.Waiting()`
  - [x] 6.2: In `internal/adapters/tui/components/status_bar.go`, replace emoji with `emoji.XXX()` calls
  - [x] 6.3: Scan `detail_panel.go` for any emoji usage and update if found

- [x] Task 7: Run Tests and Lint (AC: 6)
  - [x] 7.1: `make test` - all tests pass
  - [x] 7.2: `make lint` - no warnings

## Dev Notes

### Key Code Locations

| File | Action | Line |
|------|--------|------|
| `internal/adapters/tui/components/project_list.go` | Change `SetShowStatusBar(true)` to `false` | 40 |
| `internal/adapters/tui/views.go` | Add `>` documentation to `renderHelpOverlay()` | ~84 |
| `internal/core/ports/config.go` | Add `UseEmoji *bool` field | ~35 |
| `internal/config/loader.go` | Add Viper binding for `use_emoji` | - |
| `internal/shared/emoji/emoji.go` | **NEW FILE** - Emoji fallback system | - |
| `internal/shared/emoji/emoji_test.go` | **NEW FILE** - Tests | - |
| `internal/adapters/tui/app.go` | Wire `emoji.InitEmoji()` at startup | - |
| `internal/adapters/tui/components/delegate.go` | Replace hardcoded emoji | 167, 188, 190, 234 |
| `internal/adapters/tui/components/status_bar.go` | Replace hardcoded emoji | - |

### Emoji Package Implementation

Create `internal/shared/emoji/emoji.go`:

```go
package emoji

import (
    "os"
    "strings"
)

var (
    useEmoji    bool
    initialized bool
)

// InitEmoji must be called once at startup, AFTER config loads.
// Pass config.UseEmoji (nil=auto, true=force, false=disable)
func InitEmoji(configValue *bool) {
    if configValue != nil {
        useEmoji = *configValue
    } else {
        useEmoji = detectEmojiSupport()
    }
    initialized = true
}

func detectEmojiSupport() bool {
    term := strings.ToLower(os.Getenv("TERM"))
    limitedTerminals := []string{"linux", "vt100", "vt220", "ansi", "dumb"}
    for _, lt := range limitedTerminals {
        if strings.Contains(term, lt) || term == lt {
            return false
        }
    }
    return true
}

func Star() string     { if useEmoji { return "⭐" }; return "*" }
func Waiting() string  { if useEmoji { return "⏸️" }; return "[W]" }
func Today() string    { if useEmoji { return "✨" }; return "+" }
func ThisWeek() string { if useEmoji { return "⚡" }; return "~" }
func Warning() string  { if useEmoji { return "⚠️" }; return "!" }
```

### Config Addition

In `internal/core/ports/config.go`, add to `Config` struct:

```go
// UseEmoji controls emoji display (Story 8.9)
// nil = auto-detect from TERM, true = force emoji, false = force fallback
UseEmoji *bool
```

### Previous Story Learnings

**From Story 8.8:** Detail panel modifications are straightforward. Keep changes focused.
**From Story 8.7:** Help overlay is in `views.go:renderHelpOverlay()`, NOT a separate `help.go` file.
**From Story 8.6:** Config additions need both `ports/config.go` AND `config/loader.go` updates.
**From Story 8.5:** `project_list.go` changes are simple. Tests co-located with source.

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| Create `internal/shared/styles/emoji.go` | Use `internal/shared/emoji/emoji.go` (follows `timeformat` pattern) |
| Look for `help.go` file | Help overlay is in `internal/adapters/tui/views.go` |
| Hardcode emoji checks in each file | Centralize in emoji package |
| Remove status bar counts | Only hide Bubbles pagination |
| Change `>` to something non-standard | Document as intentional |
| Tie emoji to NO_COLOR | Only check TERM for emoji detection |
| Call `detectEmojiSupport()` on every accessor | Cache result in `InitEmoji()` |
| Add `use_emoji` to wrong config file | Add to `ports/config.go`, bind in `config/loader.go` |

### Testing Strategy

**Unit Tests (`internal/shared/emoji/emoji_test.go`):**

```go
func TestDetectEmojiSupport(t *testing.T) {
    tests := []struct {
        name     string
        term     string
        expected bool
    }{
        {"xterm supports emoji", "xterm-256color", true},
        {"linux console", "linux", false},
        {"dumb terminal", "dumb", false},
        {"vt100 terminal", "vt100", false},
        {"empty TERM defaults to emoji", "", true},
    }
    // ... test implementation with t.Setenv()
}

func TestConfigOverride(t *testing.T) {
    // Test true/false/nil override behavior
}
```

## User Testing Guide

**Time needed:** 3 minutes

### Step 1: Build and Run

```bash
make build && ./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Bottom of list | No "1/5" pagination | Pagination visible |
| Status bar | Shows "X active \| X hibernated" | Missing counts |

### Step 2: Help Overlay

```bash
# Press ? in vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Navigation section | `>` documented as selection indicator | Not mentioned |

### Step 3: Emoji Fallback

```bash
TERM=linux ./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| Favorite indicator | `*` not `⭐` | Emoji still shows |
| Waiting indicator | `[W]` not `⏸️` | Emoji still shows |

### Step 4: Config Override

```bash
# Add to ~/.vibe-dash/config.yaml:
# settings:
#   use_emoji: false
./bin/vibe
```

| Check | Expected | Red Flag |
|-------|----------|----------|
| All indicators | Fallback characters | Any emoji visible |

### Decision Guide

| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Pagination still visible | Check `SetShowStatusBar(false)` in project_list.go:40 |
| Emoji fallback not working | Check TERM detection + InitEmoji wiring in app.go |
| Help missing `>` docs | Check `views.go:renderHelpOverlay()` changes |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None required - straightforward implementation.

### Completion Notes List

1. Task 1: Hid Bubbles list pagination by setting `SetShowStatusBar(false)` in project_list.go:40. Status bar component counts are separate and unaffected.

2. Task 2: Added `>` documentation to help overlay under Navigation section in views.go.

3. Task 3: Added `UseEmoji *bool` field to Config struct and Viper binding for `settings.use_emoji`. Default config file includes commented use_emoji option.

4. Task 4: Created `internal/shared/emoji/emoji.go` package with:
   - `InitEmoji(configValue *bool)` - must be called at startup
   - `detectEmojiSupport()` - checks TERM for limited terminals
   - Accessor functions: `Star()`, `Waiting()`, `Today()`, `ThisWeek()`, `Warning()`
   - 15 table-driven tests for TERM detection and config override

5. Task 5: Wired `emoji.InitEmoji()` in app.go before TUI renders.

6. Task 6: Replaced ALL hardcoded emoji across TUI:
   - `delegate.go`: Star, Today, ThisWeek, Waiting
   - `status_bar.go`: Waiting, Warning (in condensed mode)
   - `detail_panel.go`: Star, Waiting
   - `model.go`: Star, Warning (in watcher/corruption messages)
   - `views.go`: Converted `NarrowWarning` const to function for emoji fallback

7. Task 7: All tests pass, lint clean.

8. Updated test files to initialize emoji package with `useEmoji=true` in init() functions.

### Code Review Fixes Applied

**H1: emoji.go documentation** - Added documentation that uninitialized state defaults to fallback mode.

**M1: detectEmojiSupport condition** - Reordered condition to check exact match first for readability.

**M2: status_bar.go hardcoded warning** - Replaced hardcoded `⚠` with `emoji.Warning()` in condensed config warning.

**M3: model.go unfavorite indicator** - Added `emoji.EmptyStar()` function and used it for unfavorite feedback.

**L1: Help overlay `>` documentation** - Updated text to `Selection indicator (focused)` for clarity.

**Added test:** `TestEmptyStar` in emoji_test.go for new accessor function.

### File List

**New Files:**
- `internal/shared/emoji/emoji.go` - Emoji fallback system
- `internal/shared/emoji/emoji_test.go` - Unit tests for emoji package

**Modified Files:**
- `internal/adapters/tui/components/project_list.go:40` - SetShowStatusBar(false)
- `internal/adapters/tui/views.go` - Added `>` docs, NarrowWarning() function, emoji import
- `internal/core/ports/config.go` - Added UseEmoji *bool field
- `internal/config/loader.go` - Added use_emoji Viper binding + Save + default config
- `internal/adapters/tui/app.go` - Wire emoji.InitEmoji()
- `internal/adapters/tui/components/delegate.go` - Replace emoji with emoji.XXX() calls
- `internal/adapters/tui/components/status_bar.go` - Replace emoji with emoji.XXX() calls
- `internal/adapters/tui/components/detail_panel.go` - Replace emoji with emoji.XXX() calls
- `internal/adapters/tui/model.go` - Replace emoji with emoji.XXX() calls
- `internal/adapters/tui/components/delegate_test.go` - Add emoji.InitEmoji in init()
- `internal/adapters/tui/components/status_bar_test.go` - Add emoji.InitEmoji in init()
- `internal/adapters/tui/model_test.go` - Add emoji.InitEmoji in init()
- `internal/adapters/tui/model_favorite_test.go` - Add emoji.InitEmoji in init()
- `internal/adapters/tui/views_test.go` - Update NarrowWarning() calls
- `internal/adapters/tui/model_responsive_test.go` - Update NarrowWarning() calls

