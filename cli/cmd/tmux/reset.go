package tmux

import (
	"fmt"
	"os"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
	"github.com/dlstadther/bootstrap/cli/internal/paths"

	"github.com/spf13/cobra"

	itmux "github.com/dlstadther/bootstrap/cli/internal/tmux"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Kill all tmux sessions and rebuild from YAML configs",
	Long: `Kills the tmux server entirely, then rebuilds sessions from YAML configs.

All sessions are destroyed — including ad-hoc ones not defined in any config file.
tmux-resurrect is intentionally skipped; this is a clean-slate rebuild.

Must be run from outside tmux. Running from inside a tmux session will cause
kill-server to terminate this process before the rebuild can run.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("TMUX") != "" {
			return fmt.Errorf("must be run from outside tmux — kill-server would terminate this process")
		}

		p, err := paths.Load()
		if err != nil {
			return err
		}

		return itmux.Reset(itmux.ResetOptions{
			SessionsDir:      p.TmuxSessionsDir,
			LocalSessionsDir: p.TmuxLocalSessionsDir,
		}, &iexec.Real{})
	},
}
