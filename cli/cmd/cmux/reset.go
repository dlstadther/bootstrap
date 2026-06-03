package cmux

import (
	"fmt"
	"os"
	"path/filepath"

	icmux "github.com/dlstadther/bootstrap/cli/internal/cmux"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Close all cmux workspaces and rebuild from JSON configs",
	Long: `Closes every open cmux workspace, then rebuilds from JSON configs.

All workspaces are destroyed — including ad-hoc ones not defined in any config
file. restore-session is intentionally skipped; this is a clean-slate rebuild.

Must be run from outside cmux. Running from inside a cmux workspace will close
the calling terminal before the rebuild can run.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("CMUX_WORKSPACE_ID") != "" {
			return fmt.Errorf("must be run from outside cmux — closing workspaces would terminate this process")
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}

		return icmux.Reset(icmux.ResetOptions{
			WorkspacesDir:      filepath.Join(home, ".config", "cmux", "workspaces"),
			LocalWorkspacesDir: filepath.Join(home, ".config", "cmux", "workspaces.local"),
		}, &realExecutor{})
	},
}
