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
		Long: `Generate shell completion scripts for vibe.

To load completions:

Bash:
  $ source <(vibe completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ vibe completion bash > /etc/bash_completion.d/vibe
  # macOS:
  $ vibe completion bash > $(brew --prefix)/etc/bash_completion.d/vibe

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ vibe completion zsh > "${fpath[1]}/_vibe"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ vibe completion fish | source

  # To load completions for each session, execute once:
  $ vibe completion fish > ~/.config/fish/completions/vibe.fish

PowerShell:
  PS> vibe completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> vibe completion powershell > vibe.ps1
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
