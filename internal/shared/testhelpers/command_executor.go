package testhelpers

import (
	"bytes"
	"strings"

	"github.com/spf13/cobra"
)

// RootCmdFactory is a function that creates a new root command.
type RootCmdFactory func() *cobra.Command

// RegisterCmdFunc is a function that registers a subcommand to a parent command.
type RegisterCmdFunc func(parent *cobra.Command)

// ExecuteCommand runs a cobra command with args and captures output.
// IMPORTANT: Caller must reset command flags before calling (e.g., cli.ResetAddFlags()).
//
// Usage:
//
//	cli.ResetAddFlags()  // Caller responsibility
//	output, err := testhelpers.ExecuteCommand(
//	    cli.NewRootCmd,
//	    cli.RegisterAddCommand,
//	    "add",
//	    []string{"."},
//	)
func ExecuteCommand(
	newRootCmd RootCmdFactory,
	registerCmd RegisterCmdFunc,
	cmdName string,
	args []string,
) (string, error) {
	return ExecuteCommandWithInput(newRootCmd, registerCmd, cmdName, args, "")
}

// ExecuteCommandWithInput runs a command with stdin input (for commands requiring user confirmation).
// Used by remove_test.go which needs: executeRemoveCommand(t, args, "y\n")
//
// Usage:
//
//	cli.ResetRemoveFlags()
//	output, err := testhelpers.ExecuteCommandWithInput(
//	    cli.NewRootCmd,
//	    cli.RegisterRemoveCommand,
//	    "remove",
//	    []string{"project-name"},
//	    "y\n",
//	)
func ExecuteCommandWithInput(
	newRootCmd RootCmdFactory,
	registerCmd RegisterCmdFunc,
	cmdName string,
	args []string,
	stdin string,
) (string, error) {
	cmd := newRootCmd()
	registerCmd(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if stdin != "" {
		cmd.SetIn(strings.NewReader(stdin))
	} else {
		cmd.SetIn(strings.NewReader("")) // Empty stdin for EOF simulation
	}

	fullArgs := append([]string{cmdName}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return buf.String(), err
}
