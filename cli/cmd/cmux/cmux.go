package cmux

import (
	"github.com/spf13/cobra"
)

// Cmd is the top-level 'bs cmux' group command.
var Cmd = &cobra.Command{
	Use:   "cmux",
	Short: "Manage cmux workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(workspaceCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(resetCmd)
	Cmd.AddCommand(clearCmd)
}
