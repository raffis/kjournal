package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionPowerShellCmd = &cobra.Command{
	Use:   "powershell",
	Short: "Generates powershell completion scripts",
	Example: `To load completion run
. <(kjournal completion powershell)
To configure your powershell shell to load completions for each session add to your powershell profile
Windows:
cd "$env:USERPROFILE\Documents\WindowsPowerShell\Modules"
kjournal completion >> kjournal-completion.ps1
Linux:
cd "${XDG_CONFIG_HOME:-"$HOME/.config/"}/powershell/modules"
kjournal completion >> kjournal-completions.ps1`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenPowerShellCompletion(os.Stdout)
	},
}

func init() {
	completionCmd.AddCommand(completionPowerShellCmd)
}
