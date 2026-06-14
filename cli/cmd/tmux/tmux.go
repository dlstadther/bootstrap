package tmux

import (
	"github.com/spf13/cobra"
)

// Cmd is the top-level 'bs tmux' group command.
var Cmd = &cobra.Command{
	Use:   "tmux",
	Short: "Manage tmux sessions and workspaces",
	// Prints help when invoked bare.
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(workspaceCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(resetCmd)
}
