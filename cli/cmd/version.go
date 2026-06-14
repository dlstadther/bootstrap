package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
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

// realExecutor captures combined output; use for commands where output is parsed.
type realExecutor struct{}

func (r *realExecutor) Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// streamingExecutor streams stdout/stderr to the terminal while also capturing
// combined output so callers can include it in error messages.
type streamingExecutor struct{}

func (s *streamingExecutor) Run(cmd string, args ...string) (string, error) {
	var buf bytes.Buffer
	c := exec.Command(cmd, args...)
	c.Stdout = io.MultiWriter(os.Stdout, &buf)
	c.Stderr = io.MultiWriter(os.Stderr, &buf)
	err := c.Run()
	return strings.TrimSpace(buf.String()), err
}
