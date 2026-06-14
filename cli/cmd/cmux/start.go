package cmux

import (
	"fmt"
	"os"
	"path/filepath"

	icmux "github.com/dlstadther/bootstrap/cli/internal/cmux"
	"github.com/spf13/cobra"
)

var (
	startNoRestore bool
	startOverride  bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start cmux workspaces from JSON workspace configs",
	Long: `Start cmux workspaces defined as JSON files in ~/.config/cmux/workspaces/.

Config format:

  {
    "name": "myproject",
    "cwd": "~/code/myproject",
    "panes": [
      {"command": "cac", "no_enter": true},
      {"split": "right", "command": "ls -al && bd ready"},
      {"split": "down", "command": "lazygit"}
    ]
  }

Workspace files in ~/.config/cmux/workspaces.local/ are loaded after the
main directory, allowing machine-specific overrides.

By default, cmux restore-session runs before creating workspaces.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}

		return icmux.Start(icmux.StartOptions{
			NoRestore:          startNoRestore,
			Override:           startOverride,
			WorkspacesDir:      filepath.Join(home, ".config", "cmux", "workspaces"),
			LocalWorkspacesDir: filepath.Join(home, ".config", "cmux", "workspaces.local"),
		}, &realExecutor{})
	},
}

func init() {
	startCmd.Flags().BoolVar(&startNoRestore, "no-restore", false, "Skip cmux restore-session step")
	startCmd.Flags().BoolVar(&startOverride, "override", false, "Replace workspaces named in configs (close existing before recreating)")
}
