package cmux

import (
	icmux "github.com/dlstadther/bootstrap/cli/internal/cmux"
	"github.com/spf13/cobra"
)

var (
	addName  string
	addCWD   string
	addAgent string
)

var workspaceCmd = &cobra.Command{
	Use:   "add",
	Short: "Open an agent workspace in cmux",
	Long: `Open a cmux workspace for agentic development.

Creates a new workspace with:
  - Left pane: agent command staged (no Enter)
  - Top-right pane: ls -al + bd ready
  - Bottom-right pane: lazygit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return icmux.Add(icmux.AddOptions{
			Name:  addName,
			CWD:   addCWD,
			Agent: addAgent,
		}, &realExecutor{})
	},
}

func init() {
	workspaceCmd.Flags().StringVar(&addCWD, "cwd", "", "Working directory for all panes (required)")
	workspaceCmd.Flags().StringVar(&addName, "name", "", "Workspace name override (default: basename of --cwd)")
	workspaceCmd.Flags().StringVar(&addAgent, "agent", "claude", "Agent to stage in left pane (claude|codex|gemini|opencode|pi)")
	_ = workspaceCmd.MarkFlagRequired("cwd")
}
