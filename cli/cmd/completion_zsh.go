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
. <(kjournal completion zsh)
To configure your zsh shell to load completions for each session add to your zshrc
# ~/.zshrc or ~/.profile
command -v kjournal >/dev/null && . <(kjournal completion zsh)
or write a cached file in one of the completion directories in your ${fpath}:
echo "${fpath// /\n}" | grep -i completion
kjournal completion zsh > _kjournal
mv _kjournal ~/.oh-my-zsh/completions  # oh-my-zsh
mv _kjournal ~/.zprezto/modules/completion/external/src/  # zprezto`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := rootCmd.GenZshCompletion(os.Stdout)
		// Cobra doesn't source zsh completion file, explicitly doing it here
		fmt.Println("compdef _kjournal kjournal")

		return err
	},
}

func init() {
	completionCmd.AddCommand(completionZshCmd)
}
