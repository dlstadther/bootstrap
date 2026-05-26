package tmux

import (
	"os/exec"
	"strings"

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
}

// realExecutor shells out to real commands.
type realExecutor struct{}

func (r *realExecutor) Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
