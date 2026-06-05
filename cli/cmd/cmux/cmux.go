package cmux

import (
	"os"
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
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(resetCmd)
	Cmd.AddCommand(clearCmd)
}

// realExecutor shells out to real commands.
type realExecutor struct{}

func (r *realExecutor) Run(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	// Strip auto-set cmux context vars so programmatic calls aren't scoped to
	// the calling terminal's workspace/surface when run from inside cmux.
	c.Env = stripCmuxContext(os.Environ())
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func stripCmuxContext(env []string) []string {
	skip := map[string]bool{
		"CMUX_WORKSPACE_ID": true,
		"CMUX_TAB_ID":       true,
		"CMUX_SURFACE_ID":   true,
	}
	result := make([]string, 0, len(env))
	for _, e := range env {
		key, _, _ := strings.Cut(e, "=")
		if !skip[key] {
			result = append(result, e)
		}
	}
	return result
}
