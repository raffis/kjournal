package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generates zsh completion scripts",
	Example: `To load completion run
. <(kjournal shell-completion zsh)
To configure your zsh shell to load completions for each session add to your zshrc
# ~/.zshrc or ~/.profile
command -v kjournal >/dev/null && . <(kjournal shell-completion zsh)
or write a cached file in one of the completion directories in your ${fpath}:
echo "${fpath// /\n}" | grep -i completion
kjournal shell-completion zsh > _kjournal
mv _kjournal ~/.oh-my-zsh/completions  # oh-my-zsh
mv _kjournal ~/.zprezto/modules/completion/external/src/  # zprezto`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenZshCompletion(os.Stdout)
		// Cobra doesn't source zsh completion file, explicitly doing it here
		fmt.Println("compdef _kjournal kjournal")
	},
}

func init() {
	completionCmd.AddCommand(completionZshCmd)
}
