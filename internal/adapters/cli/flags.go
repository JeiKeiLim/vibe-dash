package cli

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose    bool
	debug      bool
	configFile string
)

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging with file/line info")
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path (default: ~/.vibe-dash/config.yaml)")

	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		initLogging()
	}
}

// initLogging configures the slog logger based on command-line flags.
// --debug takes precedence over --verbose when both are specified.
// Logs go to stderr, not stdout (stdout is for user-facing output).
func initLogging() {
	var level slog.Level
	var addSource bool

	switch {
	case debug: // --debug takes precedence over --verbose
		level = slog.LevelDebug
		addSource = true
	case verbose:
		level = slog.LevelInfo
		addSource = false
	default:
		level = slog.LevelError
		addSource = false
	}

	opts := &slog.HandlerOptions{Level: level, AddSource: addSource}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}

// GetConfigFile returns the config file path specified by --config flag.
// Returns empty string if not specified.
func GetConfigFile() string {
	return configFile
}
