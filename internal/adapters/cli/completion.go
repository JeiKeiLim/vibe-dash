package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newCompletionCmd creates the completion command for shell autocompletion scripts.
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for vdash.

To load completions:

Bash:
  $ source <(vdash completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ vdash completion bash > /etc/bash_completion.d/vdash
  # macOS:
  $ vdash completion bash > $(brew --prefix)/etc/bash_completion.d/vdash

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ vdash completion zsh > "${fpath[1]}/_vdash"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ vdash completion fish | source

  # To load completions for each session, execute once:
  $ vdash completion fish > ~/.config/fish/completions/vdash.fish

PowerShell:
  PS> vdash completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> vdash completion powershell > vdash.ps1
  # and source this file from your PowerShell profile.
`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE:      runCompletion,
	}
	return cmd
}

// runCompletion generates shell completion scripts for the specified shell.
func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]
	rootCmd := cmd.Root() // CRITICAL: Get root command, not current cmd
	out := cmd.OutOrStdout()

	switch shell {
	case "bash":
		return rootCmd.GenBashCompletionV2(out, true)
	case "zsh":
		return rootCmd.GenZshCompletion(out)
	case "fish":
		return rootCmd.GenFishCompletion(out, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(out)
	default:
		// ValidArgs should prevent this, but handle gracefully
		return fmt.Errorf("unknown shell: %s (valid: bash, zsh, fish, powershell)", shell)
	}
}

// RegisterCompletionCommand registers the completion command with the given parent.
// Used for testing to create fresh command trees.
func RegisterCompletionCommand(parent *cobra.Command) {
	parent.AddCommand(newCompletionCmd())
}

func init() {
	RootCmd.AddCommand(newCompletionCmd())
}

// projectCompletionFunc provides completion for project names.
// Used by: status, remove, favorite, note, rename, exists commands.
// Searches both Name and DisplayName fields for matches.
func projectCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Don't complete if already have a project name
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Repository is package-level, injected via SetRepository() in main.go
	// See deps.go:14-21 for definition
	if repository == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	ctx := cmd.Context()
	projects, err := repository.FindAll(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	for _, p := range projects {
		// Add name
		completions = append(completions, p.Name)
		// Add display_name if different from name
		if p.DisplayName != "" && p.DisplayName != p.Name {
			completions = append(completions, p.DisplayName)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
