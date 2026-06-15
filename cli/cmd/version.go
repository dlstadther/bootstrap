package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"
	igit "github.com/dlstadther/bootstrap/cli/internal/git"
	"github.com/dlstadther/bootstrap/cli/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show compiled and repo commit hashes",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := igit.RepoPath(os.Getenv)
		return runVersion(os.Stdout, os.Stderr, repoPath, &iexec.Real{})
	},
}

// runVersion writes the compiled and repo commit hashes. Diagnostics about an
// unreadable repo go to errOut (see bs-0i1); the report goes to out. Pure apart
// from its injected writers, repoPath, and executor, so it is unit-testable
// without a real git repo.
func runVersion(out, errOut io.Writer, repoPath string, ex iexec.Executor) error {
	fmt.Fprintf(out, "compiled: %s  (%s)\n", version.CommitHash, version.BuildTime)

	if repoPath == "" {
		fmt.Fprintln(out, "repo:     unknown (RepoPath not set; run 'make install')")
		return nil
	}

	hash, err := igit.CurrentHash(repoPath, ex)
	if err != nil {
		fmt.Fprintf(errOut, "repo:     error reading repo hash: %v\n", err)
		fmt.Fprintln(errOut, "          (is git installed and is BOOTSTRAP_REPO a valid repo?)")
		return nil
	}
	dirty, err := igit.IsDirty(repoPath, ex)
	if err != nil {
		fmt.Fprintf(out, "repo:     %s  (error checking dirty state: %v)\n", hash, err)
		return nil
	}

	suffix := ""
	if dirty {
		suffix = "  (dirty)"
	} else if hash != version.CommitHash {
		suffix = "  (drift — run 'make install' to update)"
	}
	fmt.Fprintf(out, "repo:     %s%s\n", hash, suffix)
	return nil
}
