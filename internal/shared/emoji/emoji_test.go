package emoji

import (
	"testing"
)

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
		{"vt220 terminal", "vt220", false},
		{"ansi terminal", "ansi", false},
		{"empty TERM defaults to emoji", "", true},
		{"screen supports emoji", "screen-256color", true},
		{"tmux supports emoji", "tmux-256color", true},
		{"contains linux", "linux-test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM", tt.term)
			result := detectEmojiSupport()
			if result != tt.expected {
				t.Errorf("detectEmojiSupport() with TERM=%q = %v, want %v", tt.term, result, tt.expected)
			}
		})
	}
}

func TestConfigOverride(t *testing.T) {
	tests := []struct {
		name        string
		configValue *bool
		term        string
		wantEmoji   bool
	}{
		{"nil config auto-detects xterm", nil, "xterm-256color", true},
		{"nil config auto-detects linux", nil, "linux", false},
		{"true forces emoji on linux", boolPtr(true), "linux", true},
		{"false disables emoji on xterm", boolPtr(false), "xterm-256color", false},
		{"true forces emoji on dumb", boolPtr(true), "dumb", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM", tt.term)
			InitEmoji(tt.configValue)

			// Test all accessor functions return consistent values
			if tt.wantEmoji {
				if Star() != "⭐" {
					t.Errorf("Star() = %q, want emoji", Star())
				}
				if Waiting() != "⏸️" {
					t.Errorf("Waiting() = %q, want emoji", Waiting())
				}
				if Today() != "✨" {
					t.Errorf("Today() = %q, want emoji", Today())
				}
				if ThisWeek() != "⚡" {
					t.Errorf("ThisWeek() = %q, want emoji", ThisWeek())
				}
				if Warning() != "⚠️" {
					t.Errorf("Warning() = %q, want emoji", Warning())
				}
			} else {
				if Star() != "*" {
					t.Errorf("Star() = %q, want fallback", Star())
				}
				if Waiting() != "[W]" {
					t.Errorf("Waiting() = %q, want fallback", Waiting())
				}
				if Today() != "+" {
					t.Errorf("Today() = %q, want fallback", Today())
				}
				if ThisWeek() != "~" {
					t.Errorf("ThisWeek() = %q, want fallback", ThisWeek())
				}
				if Warning() != "!" {
					t.Errorf("Warning() = %q, want fallback", Warning())
				}
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// Story 8.9 code review M3: Test EmptyStar accessor
func TestEmptyStar(t *testing.T) {
	tests := []struct {
		name     string
		useEmoji bool
		expected string
	}{
		{"emoji enabled", true, "☆"},
		{"emoji disabled", false, "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitEmoji(&tt.useEmoji)
			if got := EmptyStar(); got != tt.expected {
				t.Errorf("EmptyStar() = %q, want %q", got, tt.expected)
			}
		})
	}
}
