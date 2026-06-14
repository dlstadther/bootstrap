package tmux

import (
	"os"
	"path/filepath"
	"time"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
	"github.com/dlstadther/bootstrap/cli/internal/paths"

	"github.com/spf13/cobra"

	itmux "github.com/dlstadther/bootstrap/cli/internal/tmux"
)

var (
	startNoRestore bool
	startOverride  bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tmux sessions from YAML session configs",
	Long: `Start tmux sessions defined as YAML files in ~/.config/tmux/sessions/.

The config format is compatible with TMUXinator (tmuxinator.github.io):

  name: mysession
  root: ~/myproject
  windows:
    - main:
        layout: main-vertical
        panes:
          - vim
          - git log
          - command: top
            no_enter: true   # stage without running

By default, tmux-resurrect restore runs before creating sessions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := paths.Load()
		if err != nil {
			return err
		}

		return itmux.Start(itmux.StartOptions{
			NoRestore:        startNoRestore,
			Override:         startOverride,
			SessionsDir:      p.TmuxSessionsDir,
			LocalSessionsDir: p.TmuxLocalSessionsDir,
			ResurrectPath:    findResurrectScript(p.Home),
			AfterRestoreWait: 2 * time.Second,
		}, &iexec.Real{})
	},
}

func init() {
	startCmd.Flags().BoolVar(&startNoRestore, "no-restore", false, "Skip tmux-resurrect restore step")
	startCmd.Flags().BoolVar(&startOverride, "override", false, "Replace sessions named in configs (close existing before recreating)")
}

func findResurrectScript(home string) string {
	candidates := []string{
		filepath.Join(home, ".tmux", "plugins", "tmux-resurrect", "scripts", "restore.sh"),
		filepath.Join(home, ".config", "tmux", "plugins", "tmux-resurrect", "scripts", "restore.sh"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return candidates[0]
}
