package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionFishCmd = &cobra.Command{
	Use:   "fish",
	Short: "Generates fish completion scripts",
	Example: `To configure your fish shell to load completions for each session write this script to your completions dir:
kjournal shell-completion fish > ~/.config/fish/completions/kjournal.fish
See http://fishshell.com/docs/current/index.html#completion-own for more details`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenFishCompletion(os.Stdout, true)
	},
}

func init() {
	completionCmd.AddCommand(completionFishCmd)
}
