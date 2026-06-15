package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/dlstadther/bootstrap/cli/internal/audit"
	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
	igit "github.com/dlstadther/bootstrap/cli/internal/git"
)

var auditAll bool

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit dotfile symlinks and brew package drift",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAudit(os.Getenv, &iexec.Real{}, auditAll)
	},
}

// runAudit resolves the repo path via getenv, then audits symlinks and brew
// drift. Extracted from RunE so the repo-path wiring and error propagation are
// unit-testable with a fake env and executor.
func runAudit(getenv igit.EnvGetter, ex iexec.Executor, all bool) error {
	return audit.Run(audit.Options{
		All:      all,
		RepoPath: igit.RepoPath(getenv),
	}, ex)
}

func init() {
	auditCmd.Flags().BoolVar(&auditAll, "all", false, "show OK symlinks in addition to problems")
}
