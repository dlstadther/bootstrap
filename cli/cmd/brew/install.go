package brew

import (
	ibrew "github.com/dlstadther/bootstrap/cli/internal/brew"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install packages from the repo Brewfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ibrew.Install(&realExecutor{})
	},
}
