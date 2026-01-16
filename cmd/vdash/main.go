package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/cli"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/detection"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors/bmad"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/detectors/speckit"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/logreaders"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/persistence/metrics"
	"github.com/JeiKeiLim/vibe-dash/internal/config"
	"github.com/JeiKeiLim/vibe-dash/internal/core/services"
)

// Version info - set by goreleaser via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
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

// shutdownTimeout is the maximum time to wait for graceful shutdown
const shutdownTimeout = 5 * time.Second

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		slog.Info("shutdown signal received")
		cancel()

		// Start timeout countdown
		select {
		case <-time.After(shutdownTimeout):
			slog.Warn("shutdown timeout exceeded, forcing exit")
			os.Exit(1)
		case <-done:
			// Clean exit - run() completed
		case <-sigCh:
			// Second signal - force exit immediately
			slog.Warn("force exit on repeated signal")
			os.Exit(1)
		}
	}()

	// Run application with cancellable context
	exitCode := 0
	if err := run(ctx); err != nil {
		// Only log if not a silent error (e.g., "exists" command uses exit codes only)
		if !cli.IsSilentError(err) {
			slog.Error("application error", "error", err)
		}
		exitCode = cli.MapErrorToExitCode(err)
	}
	// Signal clean completion to signal handler before exiting
	close(done)
	os.Exit(exitCode)
}

func run(ctx context.Context) error {
	// Set version info for CLI
	cli.SetVersion(version, commit, date)

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
		"detail_layout", cfg.DetailLayout,
	)

	// Story 8.6: Set detail panel layout mode for TUI
	cli.SetDetailLayout(cfg.DetailLayout)

	// Story 8.7: Store config for TUI help overlay display
	cli.SetConfig(cfg)

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

	// Cleanup coordinator with FRESH context (not cancelled ctx) for clean shutdown
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cleanupCancel()
		if err := coordinator.Close(cleanupCtx); err != nil {
			slog.Error("coordinator cleanup failed", "error", err)
		}
		slog.Debug("coordinator closed")
	}()

	// Set repository (coordinator implements ports.ProjectRepository)
	cli.SetRepository(coordinator)

	// Set DirectoryManager for remove command
	cli.SetDirectoryManager(dirMgr)

	// Initialize detection service with registry (Story 2.5)
	registry := detectors.NewRegistry()
	registry.Register(speckit.NewSpeckitDetector())
	registry.Register(bmad.NewBMADDetector())
	detectionSvc := services.NewDetectionService(registry)
	cli.SetDetectionService(detectionSvc)

	// Story 15.6: Create AgentDetectionService with Claude Code + Generic fallback detection.
	// This REPLACES the old threshold-based WaitingDetector (Story 4.3/4.4).
	// Benefits: Log-based detection (high confidence) with file-activity fallback (low confidence).
	agentService := detection.NewAgentDetectionService()
	waitingDetector := detection.NewAgentWaitingAdapter(agentService)

	slog.Debug("agent detection service initialized",
		"claude_detector", "ClaudeCodeDetector",
		"generic_detector", "GenericDetector",
	)

	// Story 4.5: Pass waitingDetector to TUI for WAITING indicator display
	cli.SetWaitingDetector(waitingDetector)

	// Story 4.6: Create FileWatcher for real-time dashboard updates
	debounce := time.Duration(cfg.RefreshDebounceMs) * time.Millisecond
	if debounce == 0 {
		debounce = filesystem.DefaultDebounce // 200ms
	}
	fileWatcher := filesystem.NewFsnotifyWatcher(debounce)
	defer fileWatcher.Close()

	slog.Debug("file watcher initialized", "debounce_ms", cfg.RefreshDebounceMs)

	// Pass to CLI for TUI integration
	cli.SetFileWatcher(fileWatcher)

	// Story 11.2: Create StateService and HibernationService for auto-hibernation
	stateService := services.NewStateService(coordinator)
	hibernationSvc := services.NewHibernationService(coordinator, stateService, cfg, basePath)
	cli.SetHibernationService(hibernationSvc)

	// Story 11.3: Wire StateService for auto-activation on file events
	cli.SetStateService(stateService)

	slog.Debug("hibernation service initialized",
		"global_hibernation_days", cfg.HibernationDays,
	)

	// Story 12.1: Initialize log reader registry for Claude Code log viewing
	logReaderReg := logreaders.NewRegistry()
	logReaderReg.Register(logreaders.NewClaudeCodeReader())
	cli.SetLogReaderRegistry(logReaderReg)

	slog.Debug("log reader registry initialized", "readers", len(logReaderReg.Readers()))

	// Story 16.2: Create MetricsRecorder for stage transition tracking
	metricsDBPath := filepath.Join(basePath, "metrics.db")
	metricsRepo := metrics.NewMetricsRepository(metricsDBPath)
	metricsRecorder := metrics.NewMetricsRecorder(metricsRepo)
	cli.SetMetricsRecorder(metricsRecorder)

	// Story 16.4: Wire metrics reader to TUI for stats view sparklines
	cli.SetMetricsReader(metricsRepo)

	// Flush pending metrics on shutdown (before coordinator.Close)
	defer func() {
		metricsRecorder.Flush(context.Background())
		slog.Debug("metrics recorder flushed")
	}()

	slog.Debug("metrics recorder initialized", "db_path", metricsDBPath)

	return cli.Execute(ctx)
}
