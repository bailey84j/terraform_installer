package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	completionLong = `Output shell completion code for the specified shell.
The shell code must be evaluated to provide interactive completions
of terraform-install commands.

For examples of loading/evaluating the completions see:
  terraform-install completion bash --help`

	completionExampleBash = `  # Installing bash completion on macOS using homebrew
  ## If running Bash 3.2 included with macOS
      brew install bash-completion
  ## or, if running Bash 4.1+
      brew install bash-completion@2
  ## If you've installed via other means, you may need add the completion to your completion directory
      terraform-install completion bash > $(brew --prefix)/etc/bash_completion.d/terraform-install

  # Installing bash completion on Linux
  ## Load the terraform-install completion code for bash into the current shell
      source <(terraform-install completion bash)
  ## Write bash completion code to a file and source it from .bash_profile
      terraform-install completion bash > ~/.terraform-install/completion.bash.inc
      printf "
        # Kubectl shell completion
        source '$HOME/.terraform-install/completion.bash.inc'
        " >> $HOME/.bash_profile
      source $HOME/.bash_profile`

	completionExampleZsh = `# Load the terraform-install completion code for zsh[1] into the current shell
      source <(terraform-install completion zsh)
  # Set the terraform-install completion code for zsh[1] to autoload on startup
      terraform-install completion zsh > "${fpath[1]}/_terraform-install"`
)

func newCompletionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Outputs shell completions for the terraform-install command",
		Long:  completionLong,
	}

	bashCompletionCmd := &cobra.Command{
		Use:     "bash",
		Short:   "Outputs the bash shell completions",
		Example: completionExampleBash,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Root().GenBashCompletion(os.Stdout)
		},
	}
	completionCmd.AddCommand(bashCompletionCmd)

	// The zsh completions didn't work for crawford, so commenting them out
	//zshCompletionCmd := &cobra.Command{
	//Use:     "zsh",
	//Short:   "Outputs the zsh shell completions",
	//Example: completionExampleZsh,
	//RunE: func(cmd *cobra.Command, _ []string) error {
	//return cmd.Root().GenZshCompletion(os.Stdout)
	//},
	//}
	//completionCmd.AddCommand(zshCompletionCmd)

	return completionCmd
}
