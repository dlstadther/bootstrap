package cmux

import (
	"fmt"
	"strings"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"

	"github.com/spf13/cobra"

	icmux "github.com/dlstadther/bootstrap/cli/internal/cmux"
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
		}, &iexec.CMux{})
	},
}

func init() {
	workspaceCmd.Flags().StringVar(&addCWD, "cwd", "", "Working directory for all panes (required)")
	workspaceCmd.Flags().StringVar(&addName, "name", "", "Workspace name override (default: basename of --cwd)")
	workspaceCmd.Flags().StringVar(&addAgent, "agent", "claude", fmt.Sprintf("Agent to stage in left pane (%s)", strings.Join(icmux.AllowedAgents, "|")))
	_ = workspaceCmd.MarkFlagRequired("cwd")
}
