package cli

import (
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// ============================================================================
// Story 8.7: Config Accessor Tests
// ============================================================================

func TestGetConfig_ReturnsDefaults_WhenNil(t *testing.T) {
	// Reset state and ensure cleanup on test completion
	appConfig = nil
	t.Cleanup(func() { appConfig = nil })

	cfg := GetConfig()

	if cfg == nil {
		t.Fatal("GetConfig() returned nil, expected non-nil defaults")
	}
	if cfg.AgentWaitingThresholdMinutes != 10 {
		t.Errorf("GetConfig().AgentWaitingThresholdMinutes = %d, want 10", cfg.AgentWaitingThresholdMinutes)
	}
	if cfg.RefreshIntervalSeconds != 10 {
		t.Errorf("GetConfig().RefreshIntervalSeconds = %d, want 10", cfg.RefreshIntervalSeconds)
	}
	if cfg.RefreshDebounceMs != 200 {
		t.Errorf("GetConfig().RefreshDebounceMs = %d, want 200", cfg.RefreshDebounceMs)
	}
	if cfg.DetailLayout != "horizontal" {
		t.Errorf("GetConfig().DetailLayout = %q, want %q", cfg.DetailLayout, "horizontal")
	}
	if cfg.HibernationDays != 14 {
		t.Errorf("GetConfig().HibernationDays = %d, want 14", cfg.HibernationDays)
	}
}

func TestSetConfig_StoresConfig(t *testing.T) {
	// Reset state and ensure cleanup on test completion
	appConfig = nil
	t.Cleanup(func() { appConfig = nil })

	cfg := &ports.Config{
		AgentWaitingThresholdMinutes: 5,
		RefreshIntervalSeconds:       30,
		RefreshDebounceMs:            500,
		DetailLayout:                 "vertical",
		HibernationDays:              7,
	}
	SetConfig(cfg)

	got := GetConfig()
	if got.AgentWaitingThresholdMinutes != 5 {
		t.Errorf("GetConfig().AgentWaitingThresholdMinutes = %d, want 5", got.AgentWaitingThresholdMinutes)
	}
	if got.RefreshIntervalSeconds != 30 {
		t.Errorf("GetConfig().RefreshIntervalSeconds = %d, want 30", got.RefreshIntervalSeconds)
	}
	if got.RefreshDebounceMs != 500 {
		t.Errorf("GetConfig().RefreshDebounceMs = %d, want 500", got.RefreshDebounceMs)
	}
	if got.DetailLayout != "vertical" {
		t.Errorf("GetConfig().DetailLayout = %q, want %q", got.DetailLayout, "vertical")
	}
	if got.HibernationDays != 7 {
		t.Errorf("GetConfig().HibernationDays = %d, want 7", got.HibernationDays)
	}
}

func TestSetConfig_OverwritesPrevious(t *testing.T) {
	// Reset state and ensure cleanup on test completion
	appConfig = nil
	t.Cleanup(func() { appConfig = nil })

	// Set first config
	cfg1 := &ports.Config{AgentWaitingThresholdMinutes: 5}
	SetConfig(cfg1)

	// Set second config
	cfg2 := &ports.Config{AgentWaitingThresholdMinutes: 15}
	SetConfig(cfg2)

	got := GetConfig()
	if got.AgentWaitingThresholdMinutes != 15 {
		t.Errorf("GetConfig().AgentWaitingThresholdMinutes = %d, want 15 (second config)", got.AgentWaitingThresholdMinutes)
	}
}

func TestGetConfig_ReturnsSameInstance(t *testing.T) {
	// Reset state and ensure cleanup on test completion
	appConfig = nil
	t.Cleanup(func() { appConfig = nil })

	cfg := &ports.Config{AgentWaitingThresholdMinutes: 5}
	SetConfig(cfg)

	got1 := GetConfig()
	got2 := GetConfig()

	if got1 != got2 {
		t.Error("GetConfig() should return the same instance")
	}
}
