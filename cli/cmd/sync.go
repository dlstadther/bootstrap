package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	isync "github.com/dlstadther/bootstrap/cli/internal/sync"
	"github.com/spf13/cobra"
)

var syncForce bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync runtime state: mise tools, Homebrew packages, Claude plugins",
	Long: `sync runs three idempotent steps in order:

  1. mise install       — install/update tool versions
  2. brew bundle check  — skip install if already satisfied (use --force to override)
  3. claude plugin install — install each enabled plugin from ~/.claude/settings.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := &realExecutor{}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}
		settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

		var runErr error
		if err := isync.SyncMise(exec); err != nil {
			fmt.Fprintf(os.Stderr, "mise error: %v\n", err)
			runErr = err
		}
		if err := isync.SyncBrew(exec, syncForce); err != nil {
			fmt.Fprintf(os.Stderr, "brew error: %v\n", err)
			runErr = err
		}
		if err := isync.SyncPlugins(settingsPath, exec); err != nil {
			fmt.Fprintf(os.Stderr, "plugins error: %v\n", err)
			runErr = err
		}
		return runErr
	},
}

func init() {
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "skip brew bundle check and run install unconditionally")
}
