package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors/speckit"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence"
	"github.com/JeiKeiLim/vibe-dash/internal/config"
	"github.com/JeiKeiLim/vibe-dash/internal/core/services"
)

// configPathAdapter implements ports.ProjectPathLookup for DirectoryManager.
type configPathAdapter struct {
	loader *config.ViperLoader
}

func (a *configPathAdapter) GetDirForPath(path string) string {
	cfg, err := a.loader.Load(context.Background())
	if err != nil {
		return ""
	}
	dirName, _ := cfg.GetDirectoryName(path)
	return dirName
}

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

	// Get base path with safety check (Story 3.5.6)
	basePath := config.GetDefaultBasePath()
	if basePath == "" {
		return fmt.Errorf("failed to determine base path: cannot access home directory")
	}

	// Create config adapter for DirectoryManager
	configAdapter := &configPathAdapter{loader: loader}

	// Create DirectoryManager with nil check
	dirMgr := filesystem.NewDirectoryManager(basePath, configAdapter)
	if dirMgr == nil {
		return fmt.Errorf("failed to initialize directory manager: cannot determine base path")
	}

	// Create RepositoryCoordinator (replaces single-DB sqlite.NewSQLiteRepository)
	coordinator := persistence.NewRepositoryCoordinator(loader, dirMgr, basePath)

	// Set repository (coordinator implements ports.ProjectRepository)
	cli.SetRepository(coordinator)

	// Set DirectoryManager for remove command
	cli.SetDirectoryManager(dirMgr)

	// Initialize detection service with registry (Story 2.5)
	registry := detectors.NewRegistry()
	registry.Register(speckit.NewSpeckitDetector())
	detectionSvc := services.NewDetectionService(registry)
	cli.SetDetectionService(detectionSvc)

	// Initialize WaitingThresholdResolver with cascade support (Story 4.4)
	// Priority: CLI flag > per-project config file > global config > default (10)
	thresholdResolver := config.NewWaitingThresholdResolver(
		cfg,
		basePath, // ~/.vibe-dash
		cli.GetWaitingThreshold(),
	)

	// Create WaitingDetector with resolver (Story 4.3/4.4)
	waitingDetector := services.NewWaitingDetector(thresholdResolver)

	// Log debug info about waiting detector initialization
	slog.Debug("waiting detector initialized",
		"cli_override", cli.GetWaitingThreshold(),
	)

	// Story 4.5: Pass waitingDetector to TUI for WAITING indicator display
	cli.SetWaitingDetector(waitingDetector)

	return cli.Execute(ctx)
}
