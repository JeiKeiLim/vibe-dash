package cli

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose          bool
	debug            bool
	configFile       string
	waitingThreshold int // -1 = use config, 0 = disabled, >0 = threshold in minutes
)

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging with file/line info")
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path (default: ~/.vibe-dash/config.yaml)")
	RootCmd.PersistentFlags().IntVar(&waitingThreshold, "waiting-threshold", -1,
		"Override agent waiting threshold in minutes (0 to disable, -1 to use config)")

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

// GetWaitingThreshold returns the CLI-specified waiting threshold.
// Returns -1 if not specified (use config), 0 if disabled, positive for threshold.
func GetWaitingThreshold() int {
	// Validate: values < -1 are invalid, treat as -1
	if waitingThreshold < -1 {
		slog.Warn("invalid --waiting-threshold value, using config",
			"value", waitingThreshold)
		return -1
	}
	return waitingThreshold
}
