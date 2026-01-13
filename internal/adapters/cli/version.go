package cli

// setupVersion configures the version template for the CLI.
// Called from SetVersion after version info is injected from main.go.
func setupVersion() {
	RootCmd.Version = appVersion
	RootCmd.SetVersionTemplate("vdash version {{.Version}} (commit: " + appCommit + ", built: " + appDate + ")\n")
}
