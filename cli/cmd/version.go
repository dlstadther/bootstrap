package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	igit "github.com/dlstadther/bootstrap/cli/internal/git"
	"github.com/dlstadther/bootstrap/cli/internal/version"
	"github.com/spf13/cobra"
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

		ex := &realExecutor{}
		hash, err := igit.CurrentHash(repoPath, ex)
		if err != nil {
			fmt.Printf("repo:     error reading repo hash: %v\n", err)
			return nil
		}
		dirty, _ := igit.IsDirty(repoPath, ex)

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

// realExecutor shells out to real commands.
type realExecutor struct{}

func (r *realExecutor) Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
