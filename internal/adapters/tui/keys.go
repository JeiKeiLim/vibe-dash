package tui

// Key binding constants for the TUI.
// These define the keyboard shortcuts used throughout the application.
const (
	// General
	KeyQuit      = "q"
	KeyForceQuit = "ctrl+c"
	KeyHelp      = "?"
	KeyEscape    = "esc"
	KeyDetail    = "d"

	// Navigation
	KeyDown      = "j"
	KeyDownArrow = "down"
	KeyUp        = "k"
	KeyUpArrow   = "up"

	// Actions
	KeyFavorite = "f"
	KeyNotes    = "n"
	KeyRemove   = "x"
	KeyAdd      = "a"
	KeyRefresh  = "r"

	// Views
	KeyHibernated  = "h"
	KeyStateToggle = "H" // Story 11.7: Manual state toggle (uppercase H)

	// Log Session (Story 12.1)
	KeyLogSession  = "S"           // Session picker in log view (AC6)
	KeyLogJumpEnd  = "G"           // Jump to end, resume auto-scroll (AC3)
	KeyShiftEnter  = "shift+enter" // Open session picker from project list
	KeyLogOpenView = "l"           // Story 12.2 AC1: Open session selector from project list

	// Stats View (Story 16.3)
	KeyStats = "s" // Open Stats View from Dashboard
)

// KeyBindings holds the current key bindings for the TUI.
// This allows for future customization of key bindings.
type KeyBindings struct {
	// General
	Quit      string
	ForceQuit string
	Help      string
	Escape    string
	Detail    string

	// Navigation
	Down      string
	DownArrow string
	Up        string
	UpArrow   string

	// Actions
	Favorite string
	Notes    string
	Remove   string
	Add      string
	Refresh  string

	// Views
	Hibernated  string
	StateToggle string // Story 11.7: Manual state toggle

	// Log Session (Story 12.1)
	LogSession  string
	LogJumpEnd  string
	ShiftEnter  string
	LogOpenView string // Story 12.2 AC1: Open session selector from project list

	// Stats View (Story 16.3)
	Stats string
}

// DefaultKeyBindings returns the default key bindings.
func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		// General
		Quit:      KeyQuit,
		ForceQuit: KeyForceQuit,
		Help:      KeyHelp,
		Escape:    KeyEscape,
		Detail:    KeyDetail,

		// Navigation
		Down:      KeyDown,
		DownArrow: KeyDownArrow,
		Up:        KeyUp,
		UpArrow:   KeyUpArrow,

		// Actions
		Favorite: KeyFavorite,
		Notes:    KeyNotes,
		Remove:   KeyRemove,
		Add:      KeyAdd,
		Refresh:  KeyRefresh,

		// Views
		Hibernated:  KeyHibernated,
		StateToggle: KeyStateToggle, // Story 11.7

		// Log Session (Story 12.1)
		LogSession:  KeyLogSession,
		LogJumpEnd:  KeyLogJumpEnd,
		ShiftEnter:  KeyShiftEnter,
		LogOpenView: KeyLogOpenView, // Story 12.2 AC1

		// Stats View (Story 16.3)
		Stats: KeyStats,
	}
}
