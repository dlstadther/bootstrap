package brew

import (
	"github.com/spf13/cobra"

	ibrew "github.com/dlstadther/bootstrap/cli/internal/brew"
	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install packages from the repo Brewfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ibrew.Install(&iexec.Real{})
	},
}
