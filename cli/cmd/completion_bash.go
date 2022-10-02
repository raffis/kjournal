package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionBashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generates bash completion scripts",
	Example: `To load completion run
. <(kjournal shell-completion bash)
To configure your bash shell to load completions for each session add to your bashrc
# ~/.bashrc or ~/.profile
command -v kjournal >/dev/null && . <(kjournal shell-completion bash)`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenBashCompletion(os.Stdout)
	},
}

func init() {
	completionCmd.AddCommand(completionBashCmd)
}
