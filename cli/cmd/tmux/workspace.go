package tmux

import (
	"fmt"
	"strings"

	itmux "github.com/dlstadther/bootstrap/cli/internal/tmux"
	"github.com/spf13/cobra"
)

var (
	addName  string
	addCWD   string
	addAgent string
)

var workspaceCmd = &cobra.Command{
	Use:   "add",
	Short: "Open an agent workspace in tmux",
	Long: `Open a tmux workspace window for agentic development.

Creates a new session (or a new window in an existing session) with:
  - Left pane: agent command staged (no Enter)
  - Top-right pane: ls -al + bd ready
  - Bottom-right pane: lazygit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return itmux.Add(itmux.AddOptions{
			Name:  addName,
			CWD:   addCWD,
			Agent: addAgent,
		}, &realExecutor{})
	},
}

func init() {
	workspaceCmd.Flags().StringVar(&addCWD, "cwd", "", "Working directory for all panes (required)")
	workspaceCmd.Flags().StringVar(&addName, "name", "", "Window name override (default: basename of --cwd)")
	workspaceCmd.Flags().StringVar(&addAgent, "agent", "claude", fmt.Sprintf("Agent to stage in left pane (%s)", strings.Join(itmux.AllowedAgents, "|")))
	_ = workspaceCmd.MarkFlagRequired("cwd")
}
