package git

import (
	"os"
	"strings"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"

	"github.com/dlstadther/bootstrap/cli/internal/version"
)

// RepoPath resolves the bootstrap repo path: $BOOTSTRAP_REPO env var first,
// compiled-in RepoPath as fallback.
func RepoPath() string {
	if v := os.Getenv("BOOTSTRAP_REPO"); v != "" {
		return v
	}
	return version.RepoPath
}

// CurrentHash returns the HEAD commit hash of the repo at repoPath.
func CurrentHash(repoPath string, exec iexec.Executor) (string, error) {
	out, err := exec.Run("git", "-C", repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// IsDirty reports whether the repo at repoPath has uncommitted changes.
func IsDirty(repoPath string, exec iexec.Executor) (bool, error) {
	out, err := exec.Run("git", "-C", repoPath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}
