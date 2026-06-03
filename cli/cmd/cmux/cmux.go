package cmux

import (
	"os/exec"
	"strings"

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
}

// realExecutor shells out to real commands.
type realExecutor struct{}

func (r *realExecutor) Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
