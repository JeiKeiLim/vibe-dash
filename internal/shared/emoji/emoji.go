// Package emoji provides fallback characters for terminals without emoji support.
// Story 8.9: Centralizes emoji handling with TERM-based auto-detection.
package emoji

import (
	"os"
	"strings"
)

var useEmoji bool

// InitEmoji must be called once at startup, AFTER config loads.
// Pass config.UseEmoji (nil=auto, true=force, false=disable).
// If not called, accessors default to fallback mode (no emoji).
func InitEmoji(configValue *bool) {
	if configValue != nil {
		useEmoji = *configValue
	} else {
		useEmoji = detectEmojiSupport()
	}
}

// detectEmojiSupport checks TERM environment for limited terminals.
// Returns false for linux, vt100, vt220, ansi, dumb terminals.
func detectEmojiSupport() bool {
	term := strings.ToLower(os.Getenv("TERM"))
	limitedTerminals := []string{"linux", "vt100", "vt220", "ansi", "dumb"}
	for _, lt := range limitedTerminals {
		if term == lt || strings.Contains(term, lt) {
			return false
		}
	}
	return true
}

// Star returns the favorite indicator.
func Star() string {
	if useEmoji {
		return "⭐"
	}
	return "*"
}

// Waiting returns the waiting/paused indicator.
func Waiting() string {
	if useEmoji {
		return "⏸️"
	}
	return "[W]"
}

// Today returns the "modified today" indicator.
func Today() string {
	if useEmoji {
		return "✨"
	}
	return "+"
}

// ThisWeek returns the "modified this week" indicator.
func ThisWeek() string {
	if useEmoji {
		return "⚡"
	}
	return "~"
}

// Warning returns the warning indicator.
func Warning() string {
	if useEmoji {
		return "⚠️"
	}
	return "!"
}

// EmptyStar returns the unfavorited indicator (Story 8.9 code review M3).
func EmptyStar() string {
	if useEmoji {
		return "☆"
	}
	return "-"
}
