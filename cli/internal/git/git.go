package git

import (
	"os"
	"strings"

	"github.com/dlstadther/bootstrap/cli/internal/version"
)

// Executor runs a command and returns combined output.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
}

// RepoPath resolves the bootstrap repo path: $BOOTSTRAP_REPO env var first,
// compiled-in RepoPath as fallback.
func RepoPath() string {
	if v := os.Getenv("BOOTSTRAP_REPO"); v != "" {
		return v
	}
	return version.RepoPath
}

// CurrentHash returns the HEAD commit hash of the repo at repoPath.
func CurrentHash(repoPath string, exec Executor) (string, error) {
	out, err := exec.Run("git", "-C", repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// IsDirty reports whether the repo at repoPath has uncommitted changes.
func IsDirty(repoPath string, exec Executor) (bool, error) {
	out, err := exec.Run("git", "-C", repoPath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}
