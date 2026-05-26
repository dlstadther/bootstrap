package cmd

import (
	"github.com/dlstadther/bootstrap/cli/internal/audit"
	igit "github.com/dlstadther/bootstrap/cli/internal/git"
	"github.com/spf13/cobra"
)

var auditAll bool

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit dotfile symlinks and brew package drift",
	RunE: func(cmd *cobra.Command, args []string) error {
		return audit.Run(audit.Options{
			All:      auditAll,
			RepoPath: igit.RepoPath(),
		}, &realExecutor{})
	},
}

func init() {
	auditCmd.Flags().BoolVar(&auditAll, "all", false, "show OK symlinks in addition to problems")
}
