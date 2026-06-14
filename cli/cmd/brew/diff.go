package brew

import (
	"path/filepath"

	ibrew "github.com/dlstadther/bootstrap/cli/internal/brew"
	igit "github.com/dlstadther/bootstrap/cli/internal/git"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show drift between live brew state and the repo Brewfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := igit.RepoPath()
		brewfile := filepath.Join(repoPath, "dotfiles", ".Brewfile")
		return ibrew.Sync(brewfile, &realExecutor{})
	},
}
