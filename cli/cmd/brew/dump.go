package brew

import (
	ibrew "github.com/dlstadther/bootstrap/cli/internal/brew"
	igit "github.com/dlstadther/bootstrap/cli/internal/git"
	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Write live brew state back to the repo Brewfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := igit.RepoPath()
		brewfile, err := ibrew.BrewfilePath(repoPath)
		if err != nil {
			return err
		}
		return ibrew.Dump(brewfile, &realExecutor{})
	},
}
