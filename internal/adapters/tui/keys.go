package tui

// Key binding constants for the TUI.
// These define the keyboard shortcuts used throughout the application.
const (
	KeyQuit      = "q"
	KeyForceQuit = "ctrl+c"
	KeyHelp      = "?"
)

// KeyBindings holds the current key bindings for the TUI.
// This allows for future customization of key bindings.
type KeyBindings struct {
	Quit      string
	ForceQuit string
	Help      string
}

// DefaultKeyBindings returns the default key bindings.
func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		Quit:      KeyQuit,
		ForceQuit: KeyForceQuit,
		Help:      KeyHelp,
	}
}
