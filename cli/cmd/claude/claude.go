package claude

import (
	"github.com/spf13/cobra"
)

// Cmd is the top-level 'bs claude' group command.
var Cmd = &cobra.Command{
	Use:   "claude",
	Short: "Manage Claude Code configuration and plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(pluginCmd)
}
