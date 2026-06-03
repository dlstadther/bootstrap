package cmux

import (
	"encoding/json"
	"fmt"
	"os"
	osExec "os/exec"
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
workspaces are closed and rebuilt from JSON configs.

Must be run from inside cmux. Running from outside (e.g. a plain terminal)
will fail because the cmux CLI requires workspace context to reach the daemon.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}

		return icmux.Reset(icmux.ResetOptions{
			WorkspacesDir:      filepath.Join(home, ".config", "cmux", "workspaces"),
			LocalWorkspacesDir: filepath.Join(home, ".config", "cmux", "workspaces.local"),
			SkipWorkspaceID:    callerWorkspaceRef(),
		}, &realExecutor{})
	},
}

// callerWorkspaceRef returns the workspace ref (e.g. "workspace:2") for the
// terminal that invoked this command, by calling cmux identify with the current
// process env (which still has CMUX_WORKSPACE_ID set). Returns "" if not inside
// cmux or if the ref cannot be determined.
func callerWorkspaceRef() string {
	if os.Getenv("CMUX_WORKSPACE_ID") == "" {
		return ""
	}
	out, err := osExec.Command("cmux", "identify").Output()
	if err != nil {
		return ""
	}
	var result struct {
		Caller struct {
			WorkspaceRef string `json:"workspace_ref"`
		} `json:"caller"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return ""
	}
	return result.Caller.WorkspaceRef
}
