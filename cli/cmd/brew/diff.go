package brew

import (
	"os"

	"github.com/spf13/cobra"

	ibrew "github.com/dlstadther/bootstrap/cli/internal/brew"
	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
	igit "github.com/dlstadther/bootstrap/cli/internal/git"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show drift between live brew state and the repo Brewfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := igit.RepoPath(os.Getenv)
		brewfile, err := ibrew.BrewfilePath(repoPath)
		if err != nil {
			return err
		}
		return ibrew.Sync(brewfile, &iexec.Real{})
	},
}
