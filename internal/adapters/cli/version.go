package cli

// Version information variables.
// These are set at build time via ldflags:
//
//	-X github.com/JeiKeiLim/vibe-dash/internal/adapters/cli.Version=$(VERSION)
//	-X github.com/JeiKeiLim/vibe-dash/internal/adapters/cli.Commit=$(COMMIT)
//	-X github.com/JeiKeiLim/vibe-dash/internal/adapters/cli.BuildDate=$(BUILD_DATE)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func init() {
	RootCmd.Version = Version
	RootCmd.SetVersionTemplate("vibe version {{.Version}} (commit: " + Commit + ", built: " + BuildDate + ")\n")
}
