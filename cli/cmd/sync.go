package cmd

import (
	"fmt"
	"io"
	"os"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
	"github.com/dlstadther/bootstrap/cli/internal/paths"

	"github.com/spf13/cobra"

	isync "github.com/dlstadther/bootstrap/cli/internal/sync"
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
		p, err := paths.Load()
		if err != nil {
			return err
		}
		return runSync(os.Stderr, p.ClaudeSettings, &iexec.Streaming{}, syncForce)
	},
}

// runSync runs the three sync steps in order, reporting each step's error to
// errOut while continuing, and returns the last error so the command exits
// non-zero if any step failed. Extracted from RunE so the orchestration and
// error-aggregation logic is unit-testable with a fake executor.
func runSync(errOut io.Writer, settingsPath string, ex iexec.Executor, force bool) error {
	var runErr error
	if err := isync.SyncMise(ex); err != nil {
		fmt.Fprintf(errOut, "mise error: %v\n", err)
		runErr = err
	}
	if err := isync.SyncBrew(ex, force); err != nil {
		fmt.Fprintf(errOut, "brew error: %v\n", err)
		runErr = err
	}
	if err := isync.SyncPlugins(settingsPath, ex); err != nil {
		fmt.Fprintf(errOut, "plugins error: %v\n", err)
		runErr = err
	}
	return runErr
}

func init() {
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "skip brew bundle check and run install unconditionally")
}
