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
	KeyHibernated = "h"
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
	Hibernated string
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
		Hibernated: KeyHibernated,
	}
}
