package tool

import (
	"os"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"

	"github.com/spf13/cobra"

	"github.com/dlstadther/bootstrap/cli/internal/toolupgrade"
)

var checkOnly bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check and optionally upgrade top-level tools (brew, claude, opencode)",
	Long: `upgrade checks each top-level tool's installed version against the latest
available version, prompts yes/no for every out-of-date tool up front, then applies
only the approved upgrades.

Use --check to print the status table without prompting or upgrading.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		executor := &iexec.Real{}
		return toolupgrade.Run(
			toolupgrade.Options{Check: checkOnly, Out: os.Stdout},
			executor,
			toolupgrade.Registry(),
			toolupgrade.StdinDecider(os.Stdin, os.Stdout),
		)
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&checkOnly, "check", false, "print status and exit without prompting or upgrading")
}
