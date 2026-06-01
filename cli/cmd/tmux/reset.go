package tmux

import (
	"fmt"
	"os"
	"path/filepath"

	itmux "github.com/dlstadther/bootstrap/cli/internal/tmux"
	"github.com/spf13/cobra"
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

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}

		sessionsDir := filepath.Join(home, ".config", "tmux", "sessions")
		localSessionsDir := filepath.Join(home, ".config", "tmux", "sessions.local")

		return itmux.Reset(itmux.ResetOptions{
			SessionsDir:      sessionsDir,
			LocalSessionsDir: localSessionsDir,
		}, &realExecutor{})
	},
}
