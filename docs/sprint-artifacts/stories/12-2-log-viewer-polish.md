# Story 12.2: Log Viewer Polish

Status: Done

## Story

As a **developer using the log viewer**,
I want **improved keyboard navigation and visual feedback**,
So that **I can efficiently navigate, search, and read logs with familiar vim-style controls**.

## User-Visible Changes

- **New:** 'L' key opens session selector from project list view (alternative to Shift+Enter)
- **New:** `Ctrl+U` / `Ctrl+D` for half-page up/down navigation
- **New:** `gg` (double 'g') for jump to top (vim-standard, replaces single 'g')
- **New:** Search mode with `/` key, `n`/`N` for next/previous match with highlighting
- **New:** Match counter display (e.g., "3/15")
- **Changed:** ANSI color codes from cclv are preserved and displayed correctly

## Acceptance Criteria

### AC1: 'L' Key Opens Session Selector (Bug Fix)

```gherkin
Scenario: 'L' key triggers session selector from project list
  Given I am viewing the project list in normal view
  And a project is selected with Claude Code logs
  When I press 'l' (case-insensitive)
  Then the session picker overlay appears
  And shows available sessions sorted by recency

Scenario: 'L' key with no logs shows flash message
  Given I am viewing the project list in normal view
  And a project is selected without Claude Code logs
  When I press 'l'
  Then a flash message "No Claude Code logs for this project" appears
  And disappears after 2 seconds

Scenario: Help overlay shows 'L' shortcut
  Given I am viewing the dashboard
  When I press '?'
  Then the help overlay shows 'L' as session picker shortcut
  And 'Shift+Enter' is also shown for legacy compatibility
```

**Technical Notes:**
- Add `KeyLogOpenView = "l"` constant to keys.go (line ~34)
- Add `LogOpenView string` field to `KeyBindings` struct (line ~67)
- Add binding in `DefaultKeyBindings()` (line ~100)
- Add case in `handleKeyMsg()` switch statement (before line 1839, after KeyStateToggle case)
- Reuse `handleShiftEnterForSessionPicker()` for the session picker logic
- Show flash message when `CanRead()` returns false (follow Story 12.1 flash pattern)
- Update help overlay at views.go to show 'L' shortcut in Actions section
- Files: keys.go, model.go, views.go

### AC2: Enhanced Vim Navigation

```gherkin
Scenario: Ctrl+D scrolls half page down
  Given I am in log viewer mode
  When I press Ctrl+D
  Then the view scrolls down by half the visible height
  And auto-scroll is paused

Scenario: Ctrl+U scrolls half page up
  Given I am in log viewer mode
  When I press Ctrl+U
  Then the view scrolls up by half the visible height
  And auto-scroll is paused

Scenario: Double 'g' jumps to top
  Given I am in log viewer mode
  When I press 'g' twice within 500ms
  Then the view jumps to the first line
  And auto-scroll is paused

Scenario: Single 'g' does nothing (changed behavior)
  Given I am in log viewer mode
  When I press 'g' once and wait more than 500ms
  Then the view does not change
  And no navigation occurs
```

**Technical Notes:**
- Add `ctrl+d` and `ctrl+u` cases in `handleTextViewKeyMsg()` switch (model.go:2666-2738)
- Half-page calculation: `contentHeight / 2`
- Double-key detection state (add to Model struct):
  ```go
  lastKeyPress string    // Last key pressed in text view
  lastKeyTime  time.Time // Time of last key press
  ```
- Define `ggTimeoutMs = 500` as constant for easy adjustment
- Remove single 'g' jump-to-top behavior (currently at model.go:2707-2710)
- File: model.go

### AC3: Search with Highlight (New Feature)

```gherkin
Scenario: Enter search mode
  Given I am in log viewer mode
  When I press '/'
  Then a search input replaces the keybindings in footer
  And I can type a search query
  And scroll percentage remains visible

Scenario: Execute search
  Given I am in search input mode
  And I have typed "error"
  When I press Enter
  Then the view jumps to the first match
  And the match counter shows "1/N" where N is total matches
  And the current match line is highlighted with reverse video

Scenario: Navigate to next match
  Given search results exist
  When I press 'n'
  Then the view jumps to the next match
  And the match counter updates (e.g., "2/15")
  And highlight moves to new match

Scenario: Navigate to previous match
  Given search results exist
  When I press 'N' (Shift+N)
  Then the view jumps to the previous match
  And the match counter updates
  And highlight moves to new match

Scenario: Exit search mode
  Given I am in search mode
  When I press Escape or Ctrl+C
  Then search mode exits
  And the search input disappears
  And normal navigation resumes
  And match highlights are cleared

Scenario: No matches found
  Given I search for "xyznonexistent"
  When the search executes
  Then the match counter shows "0/0"
  And footer briefly shows "Pattern not found"
```

**Technical Notes:**
- Add search state to Model struct:
  ```go
  searchMode      bool
  searchQuery     string    // Current search query
  searchInput     string    // Text being typed (before Enter)
  searchIndex     int       // Current match index (0-based)
  searchMatches   []int     // Line numbers with matches
  ```
- Search UI in footer: `"/{searchInput}_  [n/N] Next/Prev  {index}/{total}  {scrollPercent}%"`
- Match highlighting: Use Lipgloss reverse style on current match line
  ```go
  highlightStyle := lipgloss.NewStyle().Reverse(true)
  ```
- Handle '/' to enter search mode (only when not already in search mode)
- Handle 'n', 'N', Escape, Ctrl+C, Enter in search mode
- `findMatches(query string) []int` - case-insensitive substring search
- Reset search state when exiting text view (add to exit handler at model.go:2667-2679)
- Estimated: ~100-150 lines of new code
- File: model.go

### AC4: ANSI Color Support

```gherkin
Scenario: ANSI colors from cclv are displayed
  Given cclv produces output with ANSI color codes
  When I view the log in the TUI
  Then colors are rendered correctly
  And text is not garbled by color codes

Scenario: Long lines with colors truncate correctly
  Given a log line has ANSI codes and exceeds screen width
  When the line is displayed
  Then truncation respects visual width (not byte count)
  And ANSI codes are not split mid-sequence
  And "..." is appended at the visual end
```

**Technical Notes:**
- Current bug: model.go:2770-2771 uses `len(line)` which counts bytes not visual width
- Dependencies already available (indirect via Bubble Tea):
  - `github.com/mattn/go-runewidth` (go.mod:38) - visual width calculation
  - `github.com/muesli/ansi` (go.mod:39) - ANSI sequence handling
- Implement in model.go (or extract to `ansi_utils.go` if cleaner):
  ```go
  // stripANSI removes ANSI escape sequences for width calculation
  func stripANSI(s string) string

  // visibleWidth returns display width excluding ANSI codes
  func visibleWidth(s string) int

  // truncateToWidth truncates to visual width, preserving ANSI integrity
  func truncateToWidth(s string, width int) string
  ```
- Replace line 2770-2771:
  ```go
  // Before: if len(line) > effectiveWidth { line = line[:effectiveWidth-3] + "..." }
  // After:  if visibleWidth(line) > effectiveWidth { line = truncateToWidth(line, effectiveWidth-3) + "..." }
  ```
- File: model.go (or model.go + ansi_utils.go)

## Tasks

- [x] Task 1: 'L' Key Binding (AC1)
  - Add `KeyLogOpenView` constant and struct field to keys.go
  - Add case in `handleKeyMsg()` calling `handleShiftEnterForSessionPicker()`
  - Show flash message if no logs (reuse existing flash pattern)
  - Update help overlay in views.go

- [x] Task 2: Vim Navigation (AC2)
  - Add `ctrl+d` / `ctrl+u` half-page scroll in `handleTextViewKeyMsg()`
  - Add `lastKeyPress` / `lastKeyTime` fields to Model
  - Implement 'gg' detection with `ggTimeoutMs` constant (500ms)
  - Remove single 'g' behavior
  - Update help overlay with new shortcuts

- [x] Task 3: Search Functionality (AC3)
  - Add search state fields to Model struct
  - Implement '/' to enter search mode
  - Implement search input handling (typing, Enter, Escape, Ctrl+C)
  - Implement `findMatches()` function
  - Implement 'n'/'N' navigation with wrap-around
  - Render search footer (replaces keybindings, keeps scroll %)
  - Render match highlight (reverse video on current match line)
  - Reset search state in text view exit handler

- [x] Task 4: ANSI Color Support (AC4)
  - Implement `stripANSI()` using muesli/ansi or regex
  - Implement `visibleWidth()` using go-runewidth
  - Implement `truncateToWidth()` preserving ANSI sequences
  - Update `renderTextView()` line truncation logic
  - Test with cclv colored output

## Dev Notes

### Story 12.1 Patterns to Follow

This story builds on Story 12.1 infrastructure. Key patterns:
- `viewModeTextView` enum value (not viewModeLogs)
- Flash message pattern: `flashMsg{}` → `clearFlashMsg{}` with 2s timeout
- Session picker: `handleShiftEnterForSessionPicker()` already handles the picker logic
- Text view key handling in `handleTextViewKeyMsg()` at model.go:2658-2741

### Key Code Locations

| Purpose | Location |
|---------|----------|
| Key constants | keys.go:5-34 |
| KeyBindings struct | keys.go:36-67 |
| Main key handler switch | model.go:1693-1839 |
| Text view key handler | model.go:2658-2741 |
| Text view rendering | model.go:2743-2813 |
| Help overlay | views.go:83-160 |
| Flash message handling | model.go:1550-1560 |

### Search State Reset

When exiting text view (model.go:2667-2679), reset ALL state including new search fields:
```go
m.searchMode = false
m.searchQuery = ""
m.searchInput = ""
m.searchIndex = 0
m.searchMatches = nil
```

### Footer Layout Consideration

Current footer is ~72 chars. With search mode showing `"/{query}  [n/N] 3/15  75%"`, ensure it fits narrow terminals. Consider truncating query if needed:
```go
maxQueryDisplay := effectiveWidth - 30 // Reserve space for controls
if len(searchInput) > maxQueryDisplay {
    searchInput = searchInput[:maxQueryDisplay-3] + "..."
}
```

### Testing

| Test Type | Location | Coverage |
|-----------|----------|----------|
| Unit: ANSI functions | `model_test.go` or `ansi_utils_test.go` | stripANSI, visibleWidth, truncateToWidth |
| Unit: Search | `model_test.go` | findMatches with various inputs |
| Integration: Keys | Manual or teatest | 'L', Ctrl+U/D, 'gg', '/', 'n'/'N' |
| Manual: ANSI | Run with cclv | Color preservation, truncation |

### References

- [Source: model.go:2658-2741] - Existing text view key handling patterns
- [Source: model.go:1746-1751] - Shift+Enter session picker pattern to reuse
- [Source: keys.go:30-34] - Story 12.1 key constant pattern
- [Source: Story 12.1] - Flash message, session picker, text view infrastructure

## Dev Agent Record

### Context Reference

Investigation conducted via BMad Master agent exploration. Validated against Story 12.1 implementation patterns.

### Agent Model Used

Claude Opus 4.5

### Debug Log References

N/A - New story

### Completion Notes List

- Task 1: Added `KeyLogOpenView = "l"` constant and struct field to keys.go. Added case in `handleKeyMsg()` at model.go:1840-1845 that reuses `handleShiftEnterForSessionPicker()`. Updated help overlay in views.go to show 'L' shortcut.

- Task 2: Added `ggTimeoutMs = 500` constant, `lastKeyPress` and `lastKeyTime` fields to Model. Implemented double-key 'gg' detection in `handleTextViewKeyMsg()` at model.go:2757-2775. Added `ctrl+d` and `ctrl+u` half-page scroll at model.go:2739-2755. Updated help overlay with new shortcuts.

- Task 3: Added search state fields (`searchMode`, `searchQuery`, `searchInput`, `searchIndex`, `searchMatches`) to Model. Implemented '/' to enter search mode, Enter to execute search, 'n'/'N' for navigation, Escape/Ctrl+C to exit. Added `findMatches()` function at model.go:2892-2905. Updated footer rendering to show search input and match counter. Added match line highlighting with reverse video.

- Task 4: Implemented ANSI-aware string utilities using `github.com/charmbracelet/x/ansi` and `github.com/mattn/go-runewidth`. Added `stripANSI()`, `visibleWidth()`, and `truncateToWidth()` functions at model.go:2908-2975. Updated line truncation in `renderTextView()` to use visual width instead of byte count.

### Code Review Fixes (2026-01-14)

- **M1 Fixed:** Added '/' search shortcut and 'n/N' navigation to help overlay Log View section (views.go:121-122)
- **M2 Fixed:** Added "Pattern not found" flash message when search returns no matches (model.go:2869-2873)
- **M3 Fixed:** Made 'L' key case-insensitive by adding uppercase "L" case to handleKeyMsg (model.go:1856)
- **L1 Fixed:** Reset `lastKeyPress` in all navigation key handlers to prevent false 'gg' detection after other keys (model.go:2743, 2752, 2760, 2770, 2795, 2799, 2807, 2816)
- **cclv color:** Updated cclv command to use `--color=always` flag for proper ANSI color output in pipeline (model.go:2517)
- **Search typing bug:** Fixed shortcut keys (n, N, g, G, S, b, j, k, space) being intercepted instead of typed in search mode - all single-char shortcuts now check `searchMode` first and add to input if active
- **Search navigation bug:** Fixed Enter not exiting input mode - now `searchMode = false` after Enter so n/N can navigate through matches
- **Dashboard '/' freeze bug:** Unbound '/' filter key in bubbles/list component - pressing '/' in dashboard no longer freezes j/k navigation (project_list.go:46)
- **Project name in title:** Log viewer header now shows "ProjectName - session.jsonl" instead of just the session filename (model.go:2501-2509)
- **Tests Added:** 16 new unit tests for Ctrl+D/U, gg detection, search mode, 'L' key case-insensitivity, help overlay shortcuts, search mode typing, and Enter exiting input mode (model_test.go:4016-4440)

### File List

- `internal/adapters/tui/keys.go` - Added `KeyLogOpenView` constant and `LogOpenView` struct field
- `internal/adapters/tui/model.go` - Added search state fields, `ggTimeoutMs` constant, `lastKeyPress`/`lastKeyTime` fields, ANSI utility functions, search handling logic, vim navigation, cclv --color=always flag, "Pattern not found" flash message, 'L' case-insensitivity, lastKeyPress reset in navigation handlers, search mode typing for shortcut keys
- `internal/adapters/tui/views.go` - Updated help overlay with 'L', 'gg', 'Ctrl+D', 'Ctrl+U', '/', 'n/N' shortcuts
- `internal/adapters/tui/model_test.go` - Added unit tests for `stripANSI`, `visibleWidth`, `truncateToWidth`, `findMatches`, Ctrl+D/U scroll, gg detection, search mode, 'L' key case-insensitivity, search mode typing
- `internal/adapters/tui/components/project_list.go` - Unbound '/' filter key to prevent dashboard freeze
