package tool

import (
	"github.com/spf13/cobra"
)

// Cmd is the top-level 'bs tool' group command.
var Cmd = &cobra.Command{
	Use:   "tool",
	Short: "Manage top-level CLI tools not handled by brew or mise",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(upgradeCmd)
}
