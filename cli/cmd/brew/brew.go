package brew

import (
	"os/exec"
	"strings"

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
	Cmd.AddCommand(syncCmd)
	Cmd.AddCommand(dumpCmd)
	Cmd.AddCommand(installCmd)
}

// realExecutor shells out to real commands.
type realExecutor struct{}

func (r *realExecutor) Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
