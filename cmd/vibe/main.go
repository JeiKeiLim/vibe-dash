package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		slog.Info("shutdown signal received")
		cancel()
	}()

	// Run application with cancellable context
	if err := run(ctx); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(cli.MapErrorToExitCode(err))
	}
}

func run(ctx context.Context) error {
	// Load config (always succeeds, may log warnings for graceful degradation)
	loader := config.NewViperLoader("")
	cfg, _ := loader.Load(ctx) // Intentionally ignore error - graceful degradation

	// Store config for later use (MVP: logged only)
	// Future stories will wire this to services via dependency injection
	// Config consumed by: Story 4.1 (RefreshDebounceMs), 4.4 (AgentWaitingThresholdMinutes), 5.6 (HibernationDays)
	slog.Debug("config loaded",
		"hibernation_days", cfg.HibernationDays,
		"refresh_interval_seconds", cfg.RefreshIntervalSeconds,
		"refresh_debounce_ms", cfg.RefreshDebounceMs,
		"agent_waiting_threshold_minutes", cfg.AgentWaitingThresholdMinutes,
	)

	return cli.Execute(ctx)
}
