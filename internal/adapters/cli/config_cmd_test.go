package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

func TestConfigSet_WaitingThreshold_Success(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write initial project config
	configPath := filepath.Join(projectDir, "config.yaml")
	initialConfig := `notes: "test project"
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Create fresh command tree
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(createTestConfigCommand())

	// Override vibeHome for testing
	originalVibeHome := vibeHome
	vibeHome = tmpDir
	defer func() { vibeHome = originalVibeHome }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "test-project", "waiting-threshold", "15"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("expected success message in output")
	}

	// Verify file was updated
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if !bytes.Contains(content, []byte("agent_waiting_threshold_minutes: 15")) {
		t.Errorf("config file not updated correctly, got:\n%s", content)
	}
}

func TestConfigSet_WaitingThreshold_InvalidValue(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(createTestConfigCommand())

	originalVibeHome := vibeHome
	vibeHome = tmpDir
	defer func() { vibeHome = originalVibeHome }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "test-project", "waiting-threshold", "-5"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for negative threshold")
	}
}

func TestConfigSet_WaitingThreshold_NonNumeric(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(createTestConfigCommand())

	originalVibeHome := vibeHome
	vibeHome = tmpDir
	defer func() { vibeHome = originalVibeHome }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "test-project", "waiting-threshold", "abc"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for non-numeric threshold")
	}
}

func TestConfigSet_UnknownKey(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(createTestConfigCommand())

	originalVibeHome := vibeHome
	vibeHome = tmpDir
	defer func() { vibeHome = originalVibeHome }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "test-project", "unknown-key", "value"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for unknown config key")
	}
}

func TestConfigSet_ProjectNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create project directory - this tests the explicit directory check

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(createTestConfigCommand())

	originalVibeHome := vibeHome
	vibeHome = tmpDir
	defer func() { vibeHome = originalVibeHome }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "nonexistent-project", "waiting-threshold", "5"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent project")
	}

	// Verify error message mentions the project directory not being found
	errStr := err.Error()
	if !bytes.Contains([]byte(errStr), []byte("project directory not found")) {
		t.Errorf("error message should mention 'project directory not found', got: %s", errStr)
	}
}

func TestConfigSet_MissingArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{"config", "set"}},
		{"only project", []string{"config", "set", "project"}},
		{"only project and key", []string{"config", "set", "project", "waiting-threshold"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := NewRootCmd()
			root.AddCommand(createTestConfigCommand())
			buf := new(bytes.Buffer)
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs(tt.args)

			err := root.Execute()
			if err == nil {
				t.Error("expected error for missing args")
			}
		})
	}
}

// createTestConfigCommand creates a fresh config command for testing.
func createTestConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage vibe-dash configuration",
	}

	setCmd := &cobra.Command{
		Use:   "set <project> <key> <value>",
		Short: "Set a project configuration value",
		Args:  cobra.ExactArgs(3),
		RunE:  runConfigSet,
	}

	cmd.AddCommand(setCmd)
	return cmd
}

// ============================================================================
// Exit code tests for config command (Story 6.3: AC4, AC7)
// ============================================================================

func TestConfigSet_InvalidValue_ExitCode3(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(createTestConfigCommand())

	originalVibeHome := vibeHome
	vibeHome = tmpDir
	defer func() { vibeHome = originalVibeHome }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "test-project", "waiting-threshold", "abc"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-numeric threshold")
	}

	// Verify error is wrapped with ErrConfigInvalid
	if !errors.Is(err, domain.ErrConfigInvalid) {
		t.Errorf("expected error to wrap ErrConfigInvalid, got: %v", err)
	}

	// Verify exit code maps to 3
	exitCode := MapErrorToExitCode(err)
	if exitCode != ExitConfigInvalid {
		t.Errorf("expected exit code %d (ExitConfigInvalid), got %d", ExitConfigInvalid, exitCode)
	}
}

func TestConfigSet_NegativeValue_ExitCode3(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(createTestConfigCommand())

	originalVibeHome := vibeHome
	vibeHome = tmpDir
	defer func() { vibeHome = originalVibeHome }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	// Use -- to mark end of flags, so -5 is treated as a positional arg
	rootCmd.SetArgs([]string{"config", "set", "test-project", "waiting-threshold", "--", "-5"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for negative threshold")
	}

	// Verify error is wrapped with ErrConfigInvalid
	if !errors.Is(err, domain.ErrConfigInvalid) {
		t.Errorf("expected error to wrap ErrConfigInvalid, got: %v", err)
	}

	// Verify exit code maps to 3
	exitCode := MapErrorToExitCode(err)
	if exitCode != ExitConfigInvalid {
		t.Errorf("expected exit code %d (ExitConfigInvalid), got %d", ExitConfigInvalid, exitCode)
	}
}

func TestConfigSet_UnknownKey_ExitCode3(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(createTestConfigCommand())

	originalVibeHome := vibeHome
	vibeHome = tmpDir
	defer func() { vibeHome = originalVibeHome }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "test-project", "unknown-key", "value"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown config key")
	}

	// Verify error is wrapped with ErrConfigInvalid
	if !errors.Is(err, domain.ErrConfigInvalid) {
		t.Errorf("expected error to wrap ErrConfigInvalid, got: %v", err)
	}

	// Verify exit code maps to 3
	exitCode := MapErrorToExitCode(err)
	if exitCode != ExitConfigInvalid {
		t.Errorf("expected exit code %d (ExitConfigInvalid), got %d", ExitConfigInvalid, exitCode)
	}
}

func TestConfigSet_Success_ExitCode0(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write initial project config
	configPath := filepath.Join(projectDir, "config.yaml")
	initialConfig := `notes: "test project"
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(createTestConfigCommand())

	originalVibeHome := vibeHome
	vibeHome = tmpDir
	defer func() { vibeHome = originalVibeHome }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "test-project", "waiting-threshold", "15"})

	err := rootCmd.Execute()

	// Verify exit code maps to 0
	exitCode := MapErrorToExitCode(err)
	if exitCode != ExitSuccess {
		t.Errorf("expected exit code %d (ExitSuccess), got %d", ExitSuccess, exitCode)
	}
}
