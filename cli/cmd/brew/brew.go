package brew

import (
	"github.com/spf13/cobra"
)

// Cmd is the top-level 'bs brew' group command.
var Cmd = &cobra.Command{
	Use:   "brew",
	Short: "Manage Homebrew packages via the repo Brewfile",
	// Prints help when invoked bare.
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(diffCmd)
	Cmd.AddCommand(dumpCmd)
	Cmd.AddCommand(installCmd)
}
