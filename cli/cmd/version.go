package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
	igit "github.com/dlstadther/bootstrap/cli/internal/git"
	"github.com/dlstadther/bootstrap/cli/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show compiled and repo commit hashes",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("compiled: %s  (%s)\n", version.CommitHash, version.BuildTime)

		repoPath := igit.RepoPath()
		if repoPath == "" {
			fmt.Println("repo:     unknown (RepoPath not set; run 'make install')")
			return nil
		}

		ex := &iexec.Real{}
		hash, err := igit.CurrentHash(repoPath, ex)
		if err != nil {
			fmt.Printf("repo:     error reading repo hash: %v\n", err)
			fmt.Println("          (is git installed and is BOOTSTRAP_REPO a valid repo?)")
			return nil
		}
		dirty, err := igit.IsDirty(repoPath, ex)
		if err != nil {
			fmt.Printf("repo:     %s  (error checking dirty state: %v)\n", hash, err)
			return nil
		}

		suffix := ""
		if dirty {
			suffix = "  (dirty)"
		} else if hash != version.CommitHash {
			suffix = "  (drift — run 'make install' to update)"
		}
		fmt.Printf("repo:     %s%s\n", hash, suffix)
		return nil
	},
}
