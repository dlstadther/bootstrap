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

All workspaces except the calling one are destroyed — including ad-hoc ones not
defined in any config file. restore-session is intentionally skipped; this is a
clean-slate rebuild.

Safe to run from inside cmux: the active workspace is preserved. All other
workspaces are closed and rebuilt from JSON configs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}

		return icmux.Reset(icmux.ResetOptions{
			WorkspacesDir:      filepath.Join(home, ".config", "cmux", "workspaces"),
			LocalWorkspacesDir: filepath.Join(home, ".config", "cmux", "workspaces.local"),
			SkipWorkspaceID:    os.Getenv("CMUX_WORKSPACE_ID"),
		}, &realExecutor{})
	},
}
