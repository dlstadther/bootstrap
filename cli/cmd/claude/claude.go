package claude

import (
	"os/exec"
	"strings"

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

type realExecutor struct{}

func (r *realExecutor) Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *realExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
